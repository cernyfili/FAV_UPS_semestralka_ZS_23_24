package main

import (
	"gameserver/internal"
	"gameserver/internal/logger"
	"gameserver/internal/utils/constants"
	"gameserver/internal/utils/errorHandeling"
	"gameserver/internal/utils/helpers"
	"log"
	"time"
)

func initLogger() {
	// Initialize the logger with desired configuration

	folderPath := constants.CLogsFolderPath

	currentTime := time.Now().Format("2006-01-02_15-04-05")
	filePath := folderPath + "/log_" + currentTime + ".txt"

	config := logger.LoggerConfig{
		LogToFile:       true,
		FilePath:        filePath,
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

func setServerConfig(filepath string) {
	ip, port, err := helpers.ReadConfigFile(filepath)
	if err != nil {
		errorHandeling.PrintError(err)
		log.Fatalf("Failed to read server config: %v", err)
	}

	// Set the server configuration using the read IP and port
	constants.CConIPadress = ip
	constants.CConnPort = port
}

func main() {

	setServerConfig(constants.CConfigFilePath)

	initLogger()

	logger.Log.Info("Starting server...")

	internal.StartServer()
}
