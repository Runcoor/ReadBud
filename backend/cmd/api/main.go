package main

import (
	"fmt"
	"log"

	"readbud/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	initConfig()

	// Initialize logger
	logger.Init(viper.GetString("log.level"))

	// Set Gin mode
	mode := viper.GetString("server.mode")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()
	r.Use(gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "readbud-api",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Auth
		v1.POST("/auth/login", func(c *gin.Context) {
			c.JSON(501, gin.H{"message": "not implemented"})
		})
	}

	// Start server
	port := viper.GetInt("server.port")
	addr := fmt.Sprintf(":%d", port)
	log.Printf("ReadBud API server starting on %s", addr)
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
