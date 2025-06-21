package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Log        LogConfig        `mapstructure:"log"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
	Informer   InformerConfig   `mapstructure:"informer"`
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string `mapstructure:"level"`
}

// KubernetesConfig holds Kubernetes-specific configuration
type KubernetesConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
	Namespace  string `mapstructure:"namespace"`
}

// InformerConfig holds informer-specific configuration
type InformerConfig struct {
	ResyncPeriod string `mapstructure:"resync_period"`
	Namespace    string `mapstructure:"namespace"`
}

// Global config instance
var globalConfig *Config

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.SetEnvPrefix("K6S")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Try to read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	globalConfig = &cfg
	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "k6s")
	viper.SetDefault("app.env", "development")

	// Log defaults
	viper.SetDefault("log.level", "info")
}

// Get returns the global configuration
func Get() *Config {
	if globalConfig == nil {
		cfg, err := Load()
		if err != nil {
			panic("Failed to load configuration: " + err.Error())
		}
		return cfg
	}
	return globalConfig
}

// GetString returns a string configuration value
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt returns an integer configuration value
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool returns a boolean configuration value
func GetBool(key string) bool {
	return viper.GetBool(key)
}
