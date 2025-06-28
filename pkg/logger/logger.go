// Package logger provides structured logging functionality using zerolog.
// It offers production-ready logging with configurable levels and formats.
package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger with the specified log level.
// It configures zerolog for structured logging with human-readable timestamps.
func Init(level string) {
	logLevel := parseLogLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Configure time format for better readability
	zerolog.TimeFieldFormat = time.RFC3339

	// Use console writer for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05",
	}

	// Set global logger
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}

// parseLogLevel converts string log level to zerolog.Level.
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}
