package logger

import (
	"bytes"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestInit(t *testing.T) {
	// Test that Init doesn't panic
	Init()
	
	// Verify default level is set
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("Expected default level to be InfoLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestSetLevel(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected zerolog.Level
	}{
		{TraceLevel, zerolog.TraceLevel},
		{DebugLevel, zerolog.DebugLevel},
		{InfoLevel, zerolog.InfoLevel},
		{WarnLevel, zerolog.WarnLevel},
		{ErrorLevel, zerolog.ErrorLevel},
		{LogLevel("invalid"), zerolog.InfoLevel}, // Should default to info
	}

	for _, test := range tests {
		SetLevel(test.input)
		if zerolog.GlobalLevel() != test.expected {
			t.Errorf("SetLevel(%s): expected %v, got %v", test.input, test.expected, zerolog.GlobalLevel())
		}
	}
}

func TestLogFunctions(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.Logger = log.Output(&buf)
	
	// Set to debug level to capture all logs
	SetLevel(DebugLevel)
	
	// Test Info
	Info("test info message", map[string]interface{}{
		"key": "value",
	})
	
	if !bytes.Contains(buf.Bytes(), []byte("test info message")) {
		t.Error("Info log message not found in output")
	}
	
	if !bytes.Contains(buf.Bytes(), []byte("value")) {
		t.Error("Info log field not found in output")
	}
	
	// Clear buffer
	buf.Reset()
	
	// Test Debug
	Debug("test debug message", nil)
	
	if !bytes.Contains(buf.Bytes(), []byte("test debug message")) {
		t.Error("Debug log message not found in output")
	}
	
	// Clear buffer
	buf.Reset()
	
	// Test Warn
	Warn("test warn message", nil)
	
	if !bytes.Contains(buf.Bytes(), []byte("test warn message")) {
		t.Error("Warn log message not found in output")
	}
	
	// Clear buffer
	buf.Reset()
	
	// Test Error
	Error("test error message", nil, nil)
	
	if !bytes.Contains(buf.Bytes(), []byte("test error message")) {
		t.Error("Error log message not found in output")
	}
	
	// Restore original logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
