package main

import (
	"log"

	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	initConfig()

	// TODO: Initialize Asynq worker server
	// TODO: Register job handlers
	// TODO: Start worker

	log.Println("ReadBud Worker starting...")
	log.Println("Worker not yet implemented — waiting for Asynq setup (HY-289)")

	select {} // Block forever
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
