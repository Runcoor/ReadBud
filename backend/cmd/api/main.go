// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"readbud/internal/adapter"
	"readbud/internal/api"
	apiHTTP "readbud/internal/api/http"
	"readbud/internal/api/middleware"
	"readbud/internal/integration"
	imageStub "readbud/internal/integration/image"
	"readbud/internal/integration/llm"
	"readbud/internal/integration/storage"
	"readbud/internal/integration/wechat"
	"readbud/internal/pkg/crypto"
	"readbud/internal/pkg/database"
	"readbud/internal/pkg/logger"
	"readbud/internal/pkg/sse"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	initConfig()

	// Initialize logger
	logger.Init(viper.GetString("log.level"))
	defer func() {
		_ = logger.L.Sync()
	}()

	// JWT config
	jwtCfg := crypto.JWTConfig{
		Secret: viper.GetString("jwt.secret"),
		Expiry: time.Duration(viper.GetInt("jwt.expiry")) * time.Hour,
	}

	// Initialize database
	db, err := database.New(context.Background(), database.Config{
		Host:            viper.GetString("database.host"),
		Port:            viper.GetInt("database.port"),
		User:            viper.GetString("database.user"),
		Password:        viper.GetString("database.password"),
		DBName:          viper.GetString("database.dbname"),
		SSLMode:         viper.GetString("database.sslmode"),
		MaxOpenConns:    viper.GetInt("database.max_open_conns"),
		MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: viper.GetInt("database.conn_max_lifetime"),
	})
	if err != nil {
		log.Printf("WARNING: Database connection failed: %v (API will start without DB)", err)
	} else {
		defer func() {
			_ = database.Close(db)
		}()

		// Run auto-migration
		if migrateErr := database.AutoMigrate(db); migrateErr != nil {
			log.Printf("WARNING: Auto-migration failed: %v", migrateErr)
		} else {
			log.Println("Database auto-migration completed successfully")
		}
	}

	// Set Gin mode
	mode := viper.GetString("server.mode")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(corsMiddleware())
	r.Use(middleware.RequestSizeLimit(middleware.DefaultMaxBodySize))
	r.Use(middleware.RequestLogger())

	// Audit service for business operation tracking
	auditSvc := service.NewAuditService(logger.L)
	_ = auditSvc // Available for handler injection when needed

	// Health check
	r.GET("/health", func(c *gin.Context) {
		api.OK(c, gin.H{
			"status":  "ok",
			"service": "readbud-api",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")

	// Encryption key for secrets (derived from JWT secret)
	encSecret := viper.GetString("jwt.secret")

	// SSE hub for real-time task progress
	sseHub := sse.NewHub()

	// Storage provider — local FS or stub fallback
	storageProvider, storageRoot := newStorageProvider()
	if storageRoot != "" {
		r.Static("/static/images", storageRoot)
		logger.S().Infof("serving static images from %s at /static/images", storageRoot)
	}

	// Redis subscriber: forwards worker SSE events to the local Hub
	redisAddr := viper.GetString("redis.addr")
	redisPassword := viper.GetString("redis.password")
	redisDB := viper.GetInt("redis.db")
	sseRedis := sse.NewRedisClient(redisAddr, redisPassword, redisDB)
	go sse.StartRedisSubscriber(context.Background(), sseRedis, sseHub)

	// Wire up services and handlers
	if db != nil {
		// Repositories
		userRepo := postgres.NewUserRepository(db)
		providerRepo := postgres.NewProviderConfigRepository(db)
		wechatRepo := postgres.NewWechatAccountRepository(db)
		taskRepo := postgres.NewTaskRepository(db)
		draftRepo := postgres.NewArticleDraftRepository(db)
		blockRepo := postgres.NewArticleBlockRepository(db)
		sourceRepo := postgres.NewSourceDocumentRepository(db)

		// Publish repositories
		publishJobRepo := postgres.NewPublishJobRepository(db)
		publishRecordRepo := postgres.NewPublishRecordRepository(db)
		assetRepo := postgres.NewAssetRepository(db)

		// Extension token repository (browser plugin auth)
		extensionTokenRepo := postgres.NewExtensionTokenRepository(db)

		// Metrics repository
		metricsRepo := postgres.NewMetricsSnapshotRepository(db)

		// Topic library repository
		topicLibraryRepo := postgres.NewTopicLibraryRepository(db)

		// Distribution package repository
		distributionRepo := postgres.NewDistributionPackageRepository(db)

		// Draft version and citation repositories
		draftVersionRepo := postgres.NewDraftVersionRepository(db)
		citationRepo := postgres.NewContentCitationRepository(db)

		// Review rule repository
		reviewRuleRepo := postgres.NewReviewRuleRepository(db)

		// Brand and style profile repositories
		brandRepo := postgres.NewBrandProfileRepository(db)
		styleRepo := postgres.NewStyleProfileRepository(db)

		// Stub adapters for development (used as fallbacks for non-WeChat providers).
		stubLLM := llm.NewStubLLMProvider(logger.L)
		stubImageGen := imageStub.NewStubImageGenProvider(logger.L)

		// WeChat: production-grade implementations of token, publisher, and metrics sync.
		tokenRedis := sse.NewRedisClient(redisAddr, redisPassword, redisDB)
		wechatEncKey := crypto.DeriveKey(encSecret)
		tokenProv, err := wechat.NewRealTokenProvider(wechat.RealTokenProviderConfig{
			Resolver:  wechat.NewDBAccountResolver(wechatRepo, wechatEncKey),
			Persister: wechat.NewRepoAccountPersister(wechatRepo),
			Cache:     wechat.NewRedisTokenCache(tokenRedis, ""),
			Logger:    logger.L,
		})
		if err != nil {
			log.Fatalf("init wechat token provider: %v", err)
		}
		wechatPublisher := wechat.NewRealWeChatPublisher(nil, logger.L)
		wechatMetricsSync := wechat.NewRealMetricsSyncProvider(nil, logger.L)

		// Asynq client for task queue
		asynqClient := asynq.NewClient(asynq.RedisClientOpt{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		})

		// Services
		authSvc := service.NewAuthService(userRepo, jwtCfg)
		providerSvc := service.NewProviderConfigService(providerRepo, encSecret)

		// Dynamic provider factory — resolves real providers from DB config
		providerFactory := integration.NewProviderFactory(providerSvc, logger.L)
		lazyLLM := integration.NewLazyLLMProvider(providerFactory, stubLLM)
		lazyImageGen := integration.NewLazyImageGenProvider(providerFactory, stubImageGen)
		wechatSvc := service.NewWechatAccountService(wechatRepo, encSecret)
		taskSvc := service.NewTaskService(taskRepo, draftRepo, sseHub, asynqClient, brandRepo)
		draftSvc := service.NewDraftService(draftRepo, blockRepo, sourceRepo, taskRepo, assetRepo, storageProvider)
		coverImageSvc := service.NewCoverImageService(draftRepo, assetRepo, lazyImageGen, storageProvider, logger.L)
		contentImageSvc := service.NewContentImageService(assetRepo, wechatPublisher, storageProvider, tokenProv, logger.L)
		wechatPackageSvc := service.NewWechatPackageService(draftRepo, assetRepo, storageProvider, logger.L)
		extensionTokenSvc := service.NewExtensionTokenService(extensionTokenRepo)
		publishSvc := service.NewPublishService(publishJobRepo, publishRecordRepo, draftRepo, wechatRepo, wechatPublisher, tokenProv, contentImageSvc, coverImageSvc, logger.L)
		metricsSvc := service.NewMetricsService(metricsRepo, publishRecordRepo, wechatMetricsSync, tokenProv, logger.L)
		topicLibrarySvc := service.NewTopicLibraryService(topicLibraryRepo, taskRepo, metricsRepo, logger.L)
		distributionSvc := service.NewDistributionService(distributionRepo, draftRepo, blockRepo, lazyLLM, logger.L)
		draftVersionSvc := service.NewDraftVersionService(draftVersionRepo, draftRepo, blockRepo, citationRepo)
		citationSvc := service.NewCitationService(citationRepo, draftRepo, blockRepo, sourceRepo)
		reviewRuleSvc := service.NewReviewRuleService(reviewRuleRepo)
		brandSvc := service.NewBrandProfileService(brandRepo)
		styleSvc := service.NewStyleProfileService(styleRepo)

		// Handlers
		authHandler := apiHTTP.NewAuthHandler(authSvc)
		providerHandler := apiHTTP.NewProviderHandler(providerSvc, providerFactory, logger.L)
		wechatHandler := apiHTTP.NewWechatHandler(wechatSvc)
		taskHandler := apiHTTP.NewTaskHandler(taskSvc)
		draftHandler := apiHTTP.NewDraftHandler(draftSvc, coverImageSvc, wechatPackageSvc)
		extensionTokenHandler := apiHTTP.NewExtensionTokenHandler(extensionTokenSvc)
		sourceHandler := apiHTTP.NewSourceHandler(draftSvc)
		publishHandler := apiHTTP.NewPublishHandler(publishSvc, draftRepo, wechatRepo)
		metricsHandler := apiHTTP.NewMetricsHandler(metricsSvc, wechatRepo)
		topicHandler := apiHTTP.NewTopicHandler(topicLibrarySvc)
		distributionHandler := apiHTTP.NewDistributionHandler(distributionSvc)
		draftVersionHandler := apiHTTP.NewDraftVersionHandler(draftVersionSvc, citationSvc)
		reviewRuleHandler := apiHTTP.NewReviewRuleHandler(reviewRuleSvc)
		brandHandler := apiHTTP.NewBrandHandler(brandSvc, styleSvc)

		// Public routes (no auth required)
		authHandler.RegisterRoutes(v1)

		// Protected routes — webapp JWT auth.
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtCfg))
		{
			protected.POST("/auth/refresh", authHandler.RefreshToken)
			providerHandler.RegisterRoutes(protected)
			wechatHandler.RegisterRoutes(protected)
			taskHandler.RegisterRoutes(protected)
			draftHandler.RegisterRoutes(protected)
			publishHandler.RegisterRoutes(protected)
			metricsHandler.RegisterRoutes(protected)
			topicHandler.RegisterRoutes(protected)
			distributionHandler.RegisterRoutes(protected)
			draftVersionHandler.RegisterRoutes(protected)
			reviewRuleHandler.RegisterRoutes(protected)
			brandHandler.RegisterRoutes(protected)
			extensionTokenHandler.RegisterRoutes(protected)
			protected.GET("/tasks/:id/events", sseHub.ServeHTTP("id"))
			protected.GET("/tasks/:id/sources", sourceHandler.GetTaskSources)
		}

		// Hybrid-auth routes — accept either JWT (webapp) or extension token (browser plugin).
		// Currently only /drafts/:id/wechat-package, which the extension fetches before
		// auto-filling the WeChat editor.
		hybrid := v1.Group("")
		hybrid.Use(middleware.CombinedAuth(extensionTokenSvc, middleware.JWTAuth(jwtCfg)))
		{
			draftHandler.RegisterPackageRoute(hybrid)
		}
	} else {
		// Fallback when DB is unavailable
		v1.POST("/auth/login", func(c *gin.Context) {
			api.ServiceUnavailable(c, "数据库不可用，请稍后重试")
		})
	}

	// Start server
	port := viper.GetInt("server.port")
	addr := fmt.Sprintf(":%d", port)
	logger.S().Infof("ReadBud API server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// newStorageProvider returns a StorageProvider based on the storage.provider config key.
// Returns the provider plus the on-disk root dir (empty string when provider is not local).
// Falls back to the stub when the configured provider is unknown so we never crash on boot.
func newStorageProvider() (adapter.StorageProvider, string) {
	provider := viper.GetString("storage.provider")
	rootDir := viper.GetString("storage.root_dir")
	publicBase := viper.GetString("storage.public_base")
	switch provider {
	case "local":
		if rootDir == "" {
			rootDir = "./data/images"
		}
		if publicBase == "" {
			publicBase = "/static/images"
		}
		return storage.NewLocalStorageProvider(rootDir, publicBase, logger.L), rootDir
	default:
		logger.L.Warn("storage.provider not 'local', falling back to stub",
			zap.String("provider", provider),
		)
		return storage.NewStubStorageProvider(logger.L), ""
	}
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
}

// corsMiddleware adds CORS headers for frontend dev server.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
