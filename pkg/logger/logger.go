package logger

import (
	"os"
	"strings"

	"github.com/roman-povoroznyk/k6s/pkg/config"
	"github.com/roman-povoroznyk/k6s/pkg/version"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog with additional context
type Logger struct {
	logger zerolog.Logger
}

// New creates a new logger instance
func New() *Logger {
	// Load configuration
	cfg, _ := config.Load()
	
	// Get environment from config
	env := strings.ToLower(cfg.App.Env)
	if env == "" {
		env = "development"
	}

	// Get log level from config
	level := strings.ToLower(cfg.Log.Level)
	var zeroLevel zerolog.Level
	switch level {
	case "trace":
		zeroLevel = zerolog.TraceLevel
	case "debug":
		zeroLevel = zerolog.DebugLevel
	case "info":
		zeroLevel = zerolog.InfoLevel
	case "warn", "warning":
		zeroLevel = zerolog.WarnLevel
	case "error":
		zeroLevel = zerolog.ErrorLevel
	case "fatal":
		zeroLevel = zerolog.FatalLevel
	default:
		zeroLevel = zerolog.InfoLevel
	}

	var logger zerolog.Logger

	if env == "prod" || env == "production" {
		// Production: JSON format
		zerolog.SetGlobalLevel(zeroLevel)
		logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", cfg.App.Name).
			Str("environment", env).
			Str("version", version.GetVersion()).
			Logger()
	} else {
		// Development: Pretty format
		zerolog.SetGlobalLevel(zeroLevel)
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		logger = zerolog.New(output).With().
			Timestamp().
			Str("service", cfg.App.Name).
			Str("environment", env).
			Str("version", version.GetVersion()).
			Logger()
	}

	return &Logger{logger: logger}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	l.addFields(event, fields)
	event.Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields)
	event.Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	l.addFields(event, fields)
	event.Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields)
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields)
	event.Msg(msg)
}

// Trace logs a trace message
func (l *Logger) Trace(msg string, fields map[string]interface{}) {
	event := l.logger.Trace()
	l.addFields(event, fields)
	event.Msg(msg)
}

// addFields adds custom fields to the log event
func (l *Logger) addFields(event *zerolog.Event, fields map[string]interface{}) {
	if fields == nil {
		return
	}
	for key, value := range fields {
		event.Interface(key, value)
	}
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger
func Init() {
	globalLogger = New()
	log.Logger = globalLogger.logger
}

// Debug logs using global logger
func Debug(msg string, fields map[string]interface{}) {
	if globalLogger == nil {
		Init()
	}
	globalLogger.Debug(msg, fields)
}

// Info logs using global logger
func Info(msg string, fields map[string]interface{}) {
	if globalLogger == nil {
		Init()
	}
	globalLogger.Info(msg, fields)
}

// Warn logs using global logger
func Warn(msg string, fields map[string]interface{}) {
	if globalLogger == nil {
		Init()
	}
	globalLogger.Warn(msg, fields)
}

// Error logs using global logger
func Error(msg string, err error, fields map[string]interface{}) {
	if globalLogger == nil {
		Init()
	}
	globalLogger.Error(msg, err, fields)
}

// Fatal logs using global logger
func Fatal(msg string, err error, fields map[string]interface{}) {
	if globalLogger == nil {
		Init()
	}
	globalLogger.Fatal(msg, err, fields)
}

// Trace logs using global logger
func Trace(msg string, fields map[string]interface{}) {
	if globalLogger == nil {
		Init()
	}
	globalLogger.Trace(msg, fields)
}
