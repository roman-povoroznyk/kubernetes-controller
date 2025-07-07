package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger represents a structured logger with context
type Logger struct {
	logger zerolog.Logger
}

// Global logger instance
var global *Logger

// Config holds logger configuration
type Config struct {
	Level       string `yaml:"level" env:"LOG_LEVEL" default:"info"`
	Format      string `yaml:"format" env:"LOG_FORMAT" default:"console"`
	Development bool   `yaml:"development" env:"LOG_DEVELOPMENT" default:"false"`
	Caller      bool   `yaml:"caller" env:"LOG_CALLER" default:"false"`
}

// Init initializes the global logger with environment-aware configuration
func Init(config Config) {
	configureLogger(config)
	global = &Logger{
		logger: log.Logger,
	}
}

// InitLegacy initializes the global logger with a simple level string (backward compatibility)
func InitLegacy(level string) {
	Init(Config{Level: level})
}

// New creates a new logger instance
func New() *Logger {
	if global == nil {
		Init(Config{})
	}
	return &Logger{
		logger: log.Logger,
	}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger.With().Ctx(ctx).Logger(),
	}
}

// WithNamespace returns a logger with namespace context
func (l *Logger) WithNamespace(namespace string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("namespace", namespace).Logger(),
	}
}

// WithDeployment returns a logger with deployment context
func (l *Logger) WithDeployment(deployment string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("deployment", deployment).Logger(),
	}
}

// WithCluster returns a logger with cluster context
func (l *Logger) WithCluster(cluster string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("cluster", cluster).Logger(),
	}
}

// WithComponent returns a logger with component context
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		logger: l.logger.With().Str("component", component).Logger(),
	}
}

// WithField returns a logger with a single field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

// WithFields returns a logger with multiple fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	event := l.logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return &Logger{
		logger: event.Logger(),
	}
}

// GetLogr returns a basic logr.Logger interface (placeholder for future integration)
func (l *Logger) GetLogr() logr.Logger {
	// For now, return a no-op logger since we don't have zerologr
	noop := &noopLogger{}
	return logr.New(noop)
}

// Debug logs a debug message with optional fields
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs an info message with optional fields
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Warn logs a warning message with optional fields
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	event := l.logger.Warn()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Error logs an error message with optional fields
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Fatal logs a fatal message with optional fields and exits
func (l *Logger) Fatal(msg string, err error, fields map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Trace logs a trace message with optional fields
func (l *Logger) Trace(msg string, fields map[string]interface{}) {
	event := l.logger.Trace()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Global convenience functions

// WithContext returns a logger with context values
func WithContext(ctx context.Context) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithContext(ctx)
}

// WithNamespace returns a logger with namespace context
func WithNamespace(namespace string) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithNamespace(namespace)
}

// WithDeployment returns a logger with deployment context
func WithDeployment(deployment string) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithDeployment(deployment)
}

// WithCluster returns a logger with cluster context
func WithCluster(cluster string) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithCluster(cluster)
}

// WithComponent returns a logger with component context
func WithComponent(component string) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithComponent(component)
}

// WithField returns a logger with a single field
func WithField(key string, value interface{}) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithField(key, value)
}

// WithFields returns a logger with multiple fields
func WithFields(fields map[string]interface{}) *Logger {
	if global == nil {
		Init(Config{})
	}
	return global.WithFields(fields)
}

// Debug logs a debug message with optional fields
func Debug(msg string, fields map[string]interface{}) {
	if global == nil {
		Init(Config{})
	}
	global.Debug(msg, fields)
}

// Info logs an info message with optional fields
func Info(msg string, fields map[string]interface{}) {
	if global == nil {
		Init(Config{})
	}
	global.Info(msg, fields)
}

// Warn logs a warning message with optional fields
func Warn(msg string, fields map[string]interface{}) {
	if global == nil {
		Init(Config{})
	}
	global.Warn(msg, fields)
}

// Error logs an error message with optional fields
func Error(msg string, err error, fields map[string]interface{}) {
	if global == nil {
		Init(Config{})
	}
	global.Error(msg, err, fields)
}

// Fatal logs a fatal message with optional fields and exits
func Fatal(msg string, err error, fields map[string]interface{}) {
	if global == nil {
		Init(Config{})
	}
	global.Fatal(msg, err, fields)
}

// Trace logs a trace message with optional fields
func Trace(msg string, fields map[string]interface{}) {
	if global == nil {
		Init(Config{})
	}
	global.Trace(msg, fields)
}

// parseLogLevel converts string to zerolog.Level
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

// configureLogger configures the global logger with environment-specific settings
func configureLogger(config Config) {
	// Set global log level
	level := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(level)
	
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Determine environment mode
	env := os.Getenv("ENV")
	isProd := env == "prod" || env == "production"
	
	// Override with explicit config
	if config.Development {
		isProd = false
	}

	// Configure format
	var logger zerolog.Logger
	
	if config.Format == "json" || (isProd && config.Format != "console") {
		// JSON format for production or explicit request
		logger = zerolog.New(os.Stderr).With().
			Timestamp().
			Str("service", "k6s-controller").
			Str("version", getVersion()).
			Logger()
		
		if isProd {
			logger = logger.With().Str("environment", "production").Logger()
		} else {
			logger = logger.With().Str("environment", "development").Logger()
		}
	} else {
		// Console format for development
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
			NoColor:    os.Getenv("NO_COLOR") != "",
		}
		
		logger = zerolog.New(consoleWriter).With().
			Timestamp().
			Str("service", "k6s-controller").
			Logger()
	}
	
	// Add caller information if requested or for debug/trace levels
	if config.Caller || level <= zerolog.DebugLevel {
		logger = logger.With().Caller().Logger()
	}
	
	log.Logger = logger
}

// getVersion returns the application version from environment or default
func getVersion() string {
	if version := os.Getenv("K6S_VERSION"); version != "" {
		return version
	}
	return "dev"
}

// noopLogger is a no-op implementation of logr.LogSink
type noopLogger struct{}

func (l *noopLogger) Init(info logr.RuntimeInfo) {}
func (l *noopLogger) Enabled(level int) bool { return false }
func (l *noopLogger) Info(level int, msg string, keysAndValues ...interface{}) {}
func (l *noopLogger) Error(err error, msg string, keysAndValues ...interface{}) {}
func (l *noopLogger) WithValues(keysAndValues ...interface{}) logr.LogSink { return l }
func (l *noopLogger) WithName(name string) logr.LogSink { return l }
