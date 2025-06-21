package logger

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	logger := New()
	if logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestInit(t *testing.T) {
	// Test that Init doesn't panic
	Init()
	
	// Test that global logger is initialized
	if globalLogger == nil {
		t.Error("Global logger should be initialized after Init()")
	}
}

func TestLogLevels(t *testing.T) {
	Init()
	
	// Test that these don't panic
	Debug("test debug", nil)
	Info("test info", nil)
	Warn("test warn", nil)
	Error("test error", nil, nil)
	Trace("test trace", nil)
}

func TestLogWithFields(t *testing.T) {
	Init()
	
	fields := map[string]interface{}{
		"test_field": "test_value",
		"number":     42,
		"bool":       true,
	}
	
	// Test that these don't panic
	Debug("test with fields", fields)
	Info("test with fields", fields)
	Warn("test with fields", fields)
	Error("test with fields", nil, fields)
}

func TestEnvironmentConfig(t *testing.T) {
	// Test production environment
	os.Setenv("K6S_APP_ENV", "production")
	defer os.Unsetenv("K6S_APP_ENV")
	
	logger := New()
	if logger == nil {
		t.Error("Logger should not be nil in production mode")
	}
	
	// Test development environment
	os.Setenv("K6S_APP_ENV", "development")
	logger = New()
	if logger == nil {
		t.Error("Logger should not be nil in development mode")
	}
}
