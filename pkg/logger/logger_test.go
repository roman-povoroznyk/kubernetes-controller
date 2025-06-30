package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestInit(t *testing.T) {
	// Test that Init doesn't panic
	Init(Config{Level: "info"})
	
	// Verify default level is set
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("Expected default level to be InfoLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestLoggerCreation(t *testing.T) {
	// Test basic logger creation
	logger := New()
	if logger == nil {
		t.Error("Expected logger to be created, got nil")
	}
}

func TestLoggerWithComponents(t *testing.T) {
	logger := New()
	
	// Test with component
	compLogger := logger.WithComponent("test")
	if compLogger == nil {
		t.Error("Expected component logger to be created, got nil")
	}
	
	// Test with cluster
	clusterLogger := logger.WithCluster("test-cluster")
	if clusterLogger == nil {
		t.Error("Expected cluster logger to be created, got nil")
	}
	
	// Test with namespace
	nsLogger := logger.WithNamespace("test-namespace")
	if nsLogger == nil {
		t.Error("Expected namespace logger to be created, got nil")
	}
}

func TestLogFunctions(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.Logger = log.Output(&buf)
	
	// Test basic logging functions
	logger := New()
	
	logger.Info("test info message", nil)
	logger.Debug("test debug message", nil)
	logger.Warn("test error message", nil)
	
	// Check that something was logged
	if buf.Len() == 0 {
		t.Error("Expected log output, but got none")
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Test global logger functions
	Info("test global info", nil)
	Debug("test global debug", nil)
	Warn("test global warn", nil)
	
	// These should not panic
}

func TestLoggerWithFields(t *testing.T) {
	logger := New()
	
	// Test with single field
	fieldLogger := logger.WithField("key", "value")
	if fieldLogger == nil {
		t.Error("Expected field logger to be created, got nil")
	}
	
	// Test with multiple fields
	fieldsLogger := logger.WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	if fieldsLogger == nil {
		t.Error("Expected fields logger to be created, got nil")
	}
}
