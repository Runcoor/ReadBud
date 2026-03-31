package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/spf13/viper"

	"readbud/internal/adapter"
	"readbud/internal/integration"
	crawlerInteg "readbud/internal/integration/crawler"
	imageInteg "readbud/internal/integration/image"
	"readbud/internal/integration/llm"
	searchInteg "readbud/internal/integration/search"
	"readbud/internal/pkg/database"
	"readbud/internal/pkg/logger"
	"readbud/internal/pkg/sse"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"
	"readbud/internal/worker"
)

func main() {
	initConfig()
	logger.Init(viper.GetString("log.level"))
	defer logger.L.Sync()

	log.Println("ReadBud Worker starting...")

	// Database
	dbCfg := database.Config{
		Host:            viper.GetString("database.host"),
		Port:            viper.GetInt("database.port"),
		User:            viper.GetString("database.user"),
		Password:        viper.GetString("database.password"),
		DBName:          viper.GetString("database.dbname"),
		SSLMode:         viper.GetString("database.sslmode"),
		MaxOpenConns:    viper.GetInt("database.max_open_conns"),
		MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: viper.GetInt("database.conn_max_lifetime"),
	}
	db, err := database.New(context.Background(), dbCfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Redis config
	redisAddr := viper.GetString("redis.addr")
	redisPassword := viper.GetString("redis.password")
	redisDB := viper.GetInt("redis.db")

	// Repositories
	taskRepo := postgres.NewTaskRepository(db)
	draftRepo := postgres.NewArticleDraftRepository(db)
	blockRepo := postgres.NewArticleBlockRepository(db)
	sourceRepo := postgres.NewSourceDocumentRepository(db)
	providerRepo := postgres.NewProviderConfigRepository(db)

	// Asynq client (for enqueuing next stages)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Redis publisher for SSE events (cross-process via Redis pub/sub)
	sseRedis := sse.NewRedisClient(redisAddr, redisPassword, redisDB)
	ssePublisher := sse.NewRedisPublisher(sseRedis)

	// Services
	encSecret := viper.GetString("jwt.secret")
	providerSvc := service.NewProviderConfigService(providerRepo, encSecret)
	taskSvc := service.NewTaskService(taskRepo, draftRepo, ssePublisher, asynqClient)

	// LLM provider — dynamic from DB config with stub fallback
	stubLLM := llm.NewStubLLMProvider(logger.L)
	providerFactory := integration.NewProviderFactory(providerSvc, logger.L)
	lazyLLM := integration.NewLazyLLMProvider(providerFactory, stubLLM)

	// Image search provider (Pexels) — resolve from DB config with stub fallback
	var imageSearchProvider adapter.ImageSearchProvider
	imgSearchSecret, err := providerSvc.GetDecryptedSecret(context.Background(), "image_search")
	if err == nil && imgSearchSecret != "" {
		apiKey := imgSearchSecret
		var sec struct {
			APIKey string `json:"api_key"`
		}
		if json.Unmarshal([]byte(imgSearchSecret), &sec) == nil && sec.APIKey != "" {
			apiKey = sec.APIKey
		}
		imageSearchProvider = imageInteg.NewPexelsProvider(apiKey, logger.L)
		logger.L.Info("image search provider: Pexels")
	} else {
		imageSearchProvider = imageInteg.NewStubImageSearchProvider(logger.L)
		logger.L.Info("image search provider: stub (no config)")
	}

	// Image gen provider — lazy from DB config with stub fallback
	stubImageGen := imageInteg.NewStubImageGenProvider(logger.L)
	imageGenProvider := integration.NewLazyImageGenProvider(providerFactory, stubImageGen)

	// Google Search provider — resolve from DB config with stub fallback
	var searchProvider adapter.SearchProvider
	searchCfg, searchCfgErr := providerSvc.GetActiveByType(context.Background(), "search")
	if searchCfgErr == nil && searchCfg != nil {
		searchSecret, _ := providerSvc.DecryptSecret(context.Background(), searchCfg)
		apiKey := integration.ParseAPIKey(searchSecret)
		var cfg struct {
			SearchEngineID string `json:"search_engine_id"`
		}
		_ = json.Unmarshal(searchCfg.ConfigJSON, &cfg)
		if apiKey != "" && cfg.SearchEngineID != "" {
			searchProvider = searchInteg.NewGoogleSearchProvider(apiKey, cfg.SearchEngineID, logger.L)
			logger.L.Info("search provider: Google Custom Search")
		}
	}
	if searchProvider == nil {
		searchProvider = searchInteg.NewStubSearchProvider(logger.L)
		logger.L.Info("search provider: stub (no config)")
	}

	// Jina Reader crawler provider — resolve from DB config with stub fallback
	var crawlerProvider adapter.CrawlerProvider
	crawlerSecret, crawlerErr := providerSvc.GetDecryptedSecret(context.Background(), "crawler")
	if crawlerErr == nil && crawlerSecret != "" {
		apiKey := crawlerSecret
		var sec struct {
			APIKey string `json:"api_key"`
		}
		if json.Unmarshal([]byte(crawlerSecret), &sec) == nil && sec.APIKey != "" {
			apiKey = sec.APIKey
		}
		crawlerProvider = crawlerInteg.NewJinaReaderProvider(apiKey, logger.L)
		logger.L.Info("crawler provider: Jina Reader")
	} else {
		crawlerProvider = searchInteg.NewStubCrawlerProvider(logger.L)
		logger.L.Info("crawler provider: stub (no config)")
	}

	// Worker server
	workerCfg := worker.ServerConfig{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
		Concurrency:   5,
	}

	srv := worker.NewServer(workerCfg, taskSvc, draftRepo, blockRepo, sourceRepo, lazyLLM, searchProvider, crawlerProvider, imageSearchProvider, imageGenProvider, logger.L)

	// Start
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start worker: %v", err)
	}
	fmt.Println("ReadBud Worker is running. Press Ctrl+C to stop.")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	srv.Shutdown()
	log.Println("Worker stopped.")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("/app/configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
}
