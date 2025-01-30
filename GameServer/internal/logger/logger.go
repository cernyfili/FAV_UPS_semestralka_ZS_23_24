package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var Log *logrus.Logger

// LoggerConfig defines the configuration for the logger
type LoggerConfig struct {
	LogToFile       bool   // Enable logging to a file
	FilePath        string // Path to the log file
	UseJSONFormat   bool   // Use JSON format for logs
	LogLevel        string // Log level (debug, info, warn, error, fatal, panic)
	EnableCaller    bool   // Include file and line number in logs
	TimestampFormat string // Custom timestamp format
}

// InitLogger initializes the logger based on the provided configuration
func InitLogger(config LoggerConfig) error {
	Log = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return err
	}
	Log.SetLevel(level)

	// Set log format
	if config.UseJSONFormat {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: config.TimestampFormat,
		})
	} else {
		// Customize the TextFormatter directly
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:          true, // Show full timestamp
			TimestampFormat:        config.TimestampFormat,
			DisableColors:          false, // Disable colors in output
			DisableQuote:           true,  // Disable quote around message
			DisableSorting:         false, // Disable sorting of log fields
			DisableLevelTruncation: true,  // Disable truncating log level name
			ForceColors:            true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				// Format path and line number for clickable links in JetBrains
				absPath, _ := filepath.Abs(f.File)
				return "", fmt.Sprintf("%s:%d", absPath, f.Line)
			},
		})
	}

	// Enable caller information if requested
	Log.SetReportCaller(config.EnableCaller)

	// Set output
	if config.LogToFile && config.FilePath != "" {
		// Create the log file if it doesn't exist
		err := os.MkdirAll(filepath.Dir(config.FilePath), 0755)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(config.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		// Multi-writer for logging to console and file
		multiWriter := io.MultiWriter(os.Stdout, file)
		Log.SetOutput(multiWriter)
	} else {
		Log.SetOutput(os.Stdout)
	}

	return nil
}
