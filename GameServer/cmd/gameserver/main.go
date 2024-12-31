package main

import (
	"gameserver/internal/logger"
	"gameserver/internal/utils"
	"log"
)

func initLogger() {
	// Initialize the logger with desired configuration
	config := logger.LoggerConfig{
		LogToFile:       true,
		FilePath:        utils.LogFilePath,
		UseJSONFormat:   false,
		LogLevel:        "debug",
		EnableCaller:    true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	err := logger.InitLogger(config)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

func main() {
	initLogger()

	logger.Log.Info("Starting server...")

	//internal.StartServer()
}
