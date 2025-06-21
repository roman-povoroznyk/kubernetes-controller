package main

import (
	"errors"
	"os"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	// Demo different log levels
	logger.Trace("Starting application", map[string]interface{}{
		"component": "main", 
	})

	logger.Debug("Configuration loaded", map[string]interface{}{
		"env": os.Getenv("ENV"),
		"config_source": "environment",
	})

	logger.Info("Application started", map[string]interface{}{
		"version": "0.1.0",
		"port":    8080,
	})

	logger.Warn("High resource usage", map[string]interface{}{
		"cpu_usage":    85.5,
		"memory_usage": 78.2,
	})

	// Simulate an error
	err := errors.New("connection timeout")
	logger.Error("Database connection failed", err, map[string]interface{}{
		"database": "postgres",
		"retry":    3,
	})
}
