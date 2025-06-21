package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test loading config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Check default values
	if cfg.App.Name != "k6s" {
		t.Errorf("Expected app name 'k6s', got '%s'", cfg.App.Name)
	}
	
	if cfg.App.Env != "development" {
		t.Errorf("Expected env 'development', got '%s'", cfg.App.Env)
	}
	
	if cfg.Log.Level != "info" {
		t.Errorf("Expected log level 'info', got '%s'", cfg.Log.Level)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Set test environment variables
	os.Setenv("K6S_APP_ENV", "test")
	os.Setenv("K6S_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("K6S_APP_ENV")
		os.Unsetenv("K6S_LOG_LEVEL")
	}()
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	if cfg.App.Env != "test" {
		t.Errorf("Expected env 'test', got '%s'", cfg.App.Env)
	}
	
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Log.Level)
	}
}

func TestGetGlobalConfig(t *testing.T) {
	cfg := Get()
	if cfg == nil {
		t.Error("Global config should not be nil")
	}
	
	if cfg.App.Name == "" {
		t.Error("App name should not be empty")
	}
}
