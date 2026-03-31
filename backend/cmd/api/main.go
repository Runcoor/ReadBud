package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"readbud/internal/api"
	apiHTTP "readbud/internal/api/http"
	"readbud/internal/api/middleware"
	"readbud/internal/pkg/crypto"
	"readbud/internal/pkg/database"
	"readbud/internal/pkg/logger"
	"readbud/internal/pkg/sse"
	"readbud/internal/repository/postgres"
	"readbud/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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

		// Run auto-migration in development mode
		if viper.GetString("server.mode") == "debug" {
			if migrateErr := database.AutoMigrate(db); migrateErr != nil {
				log.Printf("WARNING: Auto-migration failed: %v", migrateErr)
			} else {
				log.Println("Database auto-migration completed successfully")
			}
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
	r.Use(corsMiddleware())

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

	// Wire up services and handlers
	if db != nil {
		// Repositories
		userRepo := postgres.NewUserRepository(db)
		providerRepo := postgres.NewProviderConfigRepository(db)
		wechatRepo := postgres.NewWechatAccountRepository(db)
		taskRepo := postgres.NewTaskRepository(db)

		// Services
		authSvc := service.NewAuthService(userRepo, jwtCfg)
		providerSvc := service.NewProviderConfigService(providerRepo, encSecret)
		wechatSvc := service.NewWechatAccountService(wechatRepo, encSecret)
		taskSvc := service.NewTaskService(taskRepo, sseHub)

		// Handlers
		authHandler := apiHTTP.NewAuthHandler(authSvc)
		providerHandler := apiHTTP.NewProviderHandler(providerSvc)
		wechatHandler := apiHTTP.NewWechatHandler(wechatSvc)
		taskHandler := apiHTTP.NewTaskHandler(taskSvc)

		// Public routes (no auth required)
		authHandler.RegisterRoutes(v1)

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtCfg))
		{
			protected.POST("/auth/refresh", authHandler.RefreshToken)
			providerHandler.RegisterRoutes(protected)
			wechatHandler.RegisterRoutes(protected)
			taskHandler.RegisterRoutes(protected)
			protected.GET("/tasks/:id/events", sseHub.ServeHTTP("id"))
		}
	} else {
		// Fallback when DB is unavailable
		v1.POST("/auth/login", func(c *gin.Context) {
			api.Error(c, 503, 503, "数据库不可用，请稍后重试")
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
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
