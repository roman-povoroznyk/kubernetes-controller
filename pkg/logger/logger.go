package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel represents supported log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	TraceLevel LogLevel = "trace"
)

// Init initializes the logger with console output and timestamp
func Init() {
	// Configure zerolog for human-readable console output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	
	// Set default level to info
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// SetLevel sets the global log level
func SetLevel(level LogLevel) {
	switch strings.ToLower(string(level)) {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// Trace logs a message at trace level
func Trace(msg string, fields map[string]interface{}) {
	event := log.Trace()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Debug logs a message at debug level
func Debug(msg string, fields map[string]interface{}) {
	event := log.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs a message at info level
func Info(msg string, fields map[string]interface{}) {
	event := log.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a message at warn level
func Warn(msg string, fields map[string]interface{}) {
	event := log.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs a message at error level
func Error(msg string, err error, fields map[string]interface{}) {
	event := log.Error()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal logs a message at fatal level and exits
func Fatal(msg string, err error, fields map[string]interface{}) {
	event := log.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}
