package main

import (
	"gameserver/internal"
	"gameserver/internal/logger"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"log"
)

func initLogger() {
	// Initialize the logger with desired configuration
	config := logger.LoggerConfig{
		LogToFile:       true,
		FilePath:        constants.CLogFilePath,
		UseJSONFormat:   false,
		LogLevel:        "debug",
		EnableCaller:    true,
		TimestampFormat: "2006-01-02 15:04:05,000",
	}

	err := logger.InitLogger(config)
	if err != nil {
		errorHandeling.PrintError(err)
		log.Fatalf("Failed to initialize logger: %v", err)
	}
}

func main() {
	initLogger()

	logger.Log.Info("Starting server...")

	internal.StartServer()
}
