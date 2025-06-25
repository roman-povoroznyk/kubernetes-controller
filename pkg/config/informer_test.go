package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	assert.Equal(t, 10*time.Minute, cfg.Informer.ResyncPeriod)
	assert.Equal(t, "", cfg.Informer.Namespace)
	assert.False(t, cfg.Informer.EnableCustomLogic)
	assert.False(t, cfg.Informer.KubectlStyle)
	assert.Equal(t, "", cfg.Informer.LabelSelector)
	assert.Equal(t, "", cfg.Informer.FieldSelector)
	assert.Equal(t, 5, cfg.Informer.WorkerPoolSize)
	assert.Equal(t, 100, cfg.Informer.QueueSize)
	
	assert.Equal(t, 1*time.Second, cfg.Watch.PollInterval)
	assert.Equal(t, 30*time.Second, cfg.Watch.Timeout)
	assert.Equal(t, 3, cfg.Watch.MaxRetries)
	assert.Equal(t, 2*time.Second, cfg.Watch.RetryBackoff)
	
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadConfigFile(t *testing.T) {
	// Create temporary config file
	configContent := `
informer:
  resync_period: "5m"
  namespace: "test"
  enable_custom_logic: true
  worker_pool_size: 3

watch:
  poll_interval: "500ms"
  timeout: "15s"

log_level: "debug"
`
	
	tmpFile, err := os.CreateTemp("", "k6s-test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()
	
	// Load config
	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)
	
	assert.Equal(t, 5*time.Minute, cfg.Informer.ResyncPeriod)
	assert.Equal(t, "test", cfg.Informer.Namespace)
	assert.True(t, cfg.Informer.EnableCustomLogic)
	assert.Equal(t, 3, cfg.Informer.WorkerPoolSize)
	
	assert.Equal(t, 500*time.Millisecond, cfg.Watch.PollInterval)
	assert.Equal(t, 15*time.Second, cfg.Watch.Timeout)
	
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadConfigNonExistentFile(t *testing.T) {
	// Should fall back to defaults when file doesn't exist
	cfg, err := LoadConfig("")
	require.NoError(t, err)
	
	// Should have default values
	assert.Equal(t, 10*time.Minute, cfg.Informer.ResyncPeriod)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		modifyFunc  func(*Config)
		expectError bool
	}{
		{
			name:        "valid config",
			modifyFunc:  func(c *Config) {},
			expectError: false,
		},
		{
			name: "negative resync period",
			modifyFunc: func(c *Config) {
				c.Informer.ResyncPeriod = -1 * time.Second
			},
			expectError: true,
		},
		{
			name: "zero worker pool size",
			modifyFunc: func(c *Config) {
				c.Informer.WorkerPoolSize = 0
			},
			expectError: true,
		},
		{
			name: "negative queue size",
			modifyFunc: func(c *Config) {
				c.Informer.QueueSize = -1
			},
			expectError: true,
		},
		{
			name: "zero poll interval",
			modifyFunc: func(c *Config) {
				c.Watch.PollInterval = 0
			},
			expectError: true,
		},
		{
			name: "negative timeout",
			modifyFunc: func(c *Config) {
				c.Watch.Timeout = -1 * time.Second
			},
			expectError: true,
		},
		{
			name: "negative max retries",
			modifyFunc: func(c *Config) {
				c.Watch.MaxRetries = -1
			},
			expectError: true,
		},
		{
			name: "negative retry backoff",
			modifyFunc: func(c *Config) {
				c.Watch.RetryBackoff = -1 * time.Second
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modifyFunc(cfg)
			
			err := cfg.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigEnvironmentVariables(t *testing.T) {
	// Set environment variables
	originalEnv := make(map[string]string)
	envVars := map[string]string{
		"K6S_LOG_LEVEL":                    "debug",
		"K6S_INFORMER_NAMESPACE":          "env-test",
		"K6S_INFORMER_ENABLE_CUSTOM_LOGIC": "true",
		"K6S_INFORMER_RESYNC_PERIOD":      "3m",
		"K6S_WATCH_POLL_INTERVAL":         "200ms",
	}
	
	// Save original environment and set test values
	for key, value := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}
	
	// Restore environment after test
	defer func() {
		for key, originalValue := range originalEnv {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}()
	
	// Load config (should pick up env vars)
	cfg, err := LoadConfig("")
	require.NoError(t, err)
	
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "env-test", cfg.Informer.Namespace)
	assert.True(t, cfg.Informer.EnableCustomLogic)
	assert.Equal(t, 3*time.Minute, cfg.Informer.ResyncPeriod)
	assert.Equal(t, 200*time.Millisecond, cfg.Watch.PollInterval)
}
