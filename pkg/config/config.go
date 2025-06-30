package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	// General configuration
	LogLevel string `yaml:"log_level" json:"log_level"`
	
	// Controller configuration
	Controller ControllerConfig `yaml:"controller" json:"controller"`
	
	// Multi-cluster configuration
	MultiCluster MultiClusterConfig `yaml:"multi_cluster" json:"multi_cluster"`
	
	// Legacy fields for backward compatibility
	Informer *LegacyInformerConfig `yaml:"informer,omitempty" json:"informer,omitempty"`
	Watch    *LegacyWatchConfig    `yaml:"watch,omitempty" json:"watch,omitempty"`
	
	// Direct multi-cluster fields (for compatibility with existing files)
	DefaultNamespace           string          `yaml:"default_namespace,omitempty" json:"default_namespace,omitempty"`
	ConnectionTimeout          *time.Duration  `yaml:"connection_timeout,omitempty" json:"connection_timeout,omitempty"`
	MaxConcurrentConnections   *int            `yaml:"max_concurrent_connections,omitempty" json:"max_concurrent_connections,omitempty"`
	Clusters                   []ClusterConfig `yaml:"clusters,omitempty" json:"clusters,omitempty"`
}

// LegacyInformerConfig represents legacy informer configuration for backward compatibility
type LegacyInformerConfig struct {
	Namespace             string        `yaml:"namespace" json:"namespace"`
	ResyncPeriod          string        `yaml:"resync_period" json:"resync_period"`
	EnableCustomLogic     bool          `yaml:"enable_custom_logic" json:"enable_custom_logic"`
	LabelSelector         string        `yaml:"label_selector" json:"label_selector"`
	WorkerPoolSize        int           `yaml:"worker_pool_size" json:"worker_pool_size"`
	QueueSize             int           `yaml:"queue_size" json:"queue_size"`
}

// LegacyWatchConfig represents legacy watch configuration for backward compatibility
type LegacyWatchConfig struct {
	PollInterval  string `yaml:"poll_interval" json:"poll_interval"`
	Timeout       string `yaml:"timeout" json:"timeout"`
	MaxRetries    int    `yaml:"max_retries" json:"max_retries"`
	RetryBackoff  string `yaml:"retry_backoff" json:"retry_backoff"`
}

// ControllerConfig represents controller-specific configuration
type ControllerConfig struct {
	// Mode can be "single" or "multi"
	Mode string `yaml:"mode" json:"mode"`
	
	// Single cluster configuration
	Single SingleClusterConfig `yaml:"single" json:"single"`
	
	// Multi-cluster configuration file path
	ConfigFile string `yaml:"config_file" json:"config_file"`
	
	// Resync period for informers
	ResyncPeriod time.Duration `yaml:"resync_period" json:"resync_period"`
}

// SingleClusterConfig represents single cluster mode configuration
type SingleClusterConfig struct {
	// Namespace to watch (empty = all namespaces)
	Namespace string `yaml:"namespace" json:"namespace"`
	
	// Metrics configuration
	MetricsPort int `yaml:"metrics_port" json:"metrics_port"`
	
	// Health check configuration
	HealthPort int `yaml:"health_port" json:"health_port"`
	
	// Leader election configuration
	LeaderElection LeaderElectionConfig `yaml:"leader_election" json:"leader_election"`
}

// LeaderElectionConfig represents leader election configuration
type LeaderElectionConfig struct {
	// Enable leader election
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// Leader election ID
	ID string `yaml:"id" json:"id"`
	
	// Leader election namespace
	Namespace string `yaml:"namespace" json:"namespace"`
}

// MultiClusterConfig represents multi-cluster configuration
type MultiClusterConfig struct {
	// Test connectivity when listing clusters
	TestConnectivity bool `yaml:"test_connectivity" json:"test_connectivity"`
	
	// Multi-cluster settings
	DefaultNamespace       string        `yaml:"default_namespace" json:"default_namespace"`
	ConnectionTimeout      time.Duration `yaml:"connection_timeout" json:"connection_timeout"`
	MaxConcurrentConns     int           `yaml:"max_concurrent_connections" json:"max_concurrent_connections"`
	
	// Clusters configuration
	Clusters []ClusterConfig `yaml:"clusters" json:"clusters"`
}

// ClusterConfig represents a single cluster configuration
type ClusterConfig struct {
	Name       string `yaml:"name" json:"name"`
	KubeConfig string `yaml:"kubeconfig" json:"kubeconfig"`
	Context    string `yaml:"context" json:"context"`
	Namespace  string `yaml:"namespace" json:"namespace"`
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Primary    bool   `yaml:"primary" json:"primary"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		LogLevel: "info",
		Controller: ControllerConfig{
			Mode: "single",
			Single: SingleClusterConfig{
				Namespace:   "",
				MetricsPort: 8080,
				HealthPort:  8081,
				LeaderElection: LeaderElectionConfig{
					Enabled:   true,
					ID:        "k6s-controller",
					Namespace: "default",
				},
			},
			ConfigFile:   "",
			ResyncPeriod: 30 * time.Second,
		},
		MultiCluster: MultiClusterConfig{
			TestConnectivity:       false,
			DefaultNamespace:       "default",
			ConnectionTimeout:      30 * time.Second,
			MaxConcurrentConns:     10,
			Clusters:               []ClusterConfig{},
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configFile string) (*Config, error) {
	// Start with default config
	config := DefaultConfig()
	
	// If no config file specified, try default location
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return config, nil // Return default config if can't get home dir
		}
		configFile = filepath.Join(homeDir, ".k6s", "k6s.yaml")
	}
	
	// Check if file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return config, nil // Return default config if file doesn't exist
	}
	
	// Read file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", configFile, err)
	}
	
	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %v", configFile, err)
	}
	
	// Migrate legacy configuration if needed
	if err := migrateLegacyConfig(config); err != nil {
		return nil, fmt.Errorf("failed to migrate legacy config: %v", err)
	}
	
	return config, nil
}

// migrateLegacyConfig migrates old configuration format to new format
func migrateLegacyConfig(config *Config) error {
	// Migrate direct cluster fields to MultiCluster
	if len(config.Clusters) > 0 {
		config.MultiCluster.Clusters = config.Clusters
		config.Clusters = nil // Clear legacy field
	}
	
	if config.DefaultNamespace != "" {
		config.MultiCluster.DefaultNamespace = config.DefaultNamespace
		config.DefaultNamespace = "" // Clear legacy field
	}
	
	if config.ConnectionTimeout != nil {
		config.MultiCluster.ConnectionTimeout = *config.ConnectionTimeout
		config.ConnectionTimeout = nil // Clear legacy field
	}
	
	if config.MaxConcurrentConnections != nil {
		config.MultiCluster.MaxConcurrentConns = *config.MaxConcurrentConnections
		config.MaxConcurrentConnections = nil // Clear legacy field
	}
	
	// Migrate legacy informer config
	if config.Informer != nil {
		if config.Informer.Namespace != "" {
			config.Controller.Single.Namespace = config.Informer.Namespace
		}
		
		if config.Informer.ResyncPeriod != "" {
			if duration, err := time.ParseDuration(config.Informer.ResyncPeriod); err == nil {
				config.Controller.ResyncPeriod = duration
			}
		}
		
		// Clear legacy field after migration
		config.Informer = nil
	}
	
	// Migrate legacy watch config (store in controller config for future use)
	if config.Watch != nil {
		// For now, we'll just clear it since we don't have equivalent fields
		// This could be extended in the future
		config.Watch = nil
	}
	
	return nil
}

// ResolveConfig resolves configuration from multiple sources in priority order:
// 1. Environment variables (highest priority)
// 2. CLI flags
// 3. Config file
// 4. Default values (lowest priority)
func ResolveConfig(config *Config, flagValues *FlagValues) *Config {
	resolved := &Config{}
	*resolved = *config // Copy config values
	
	// Override with flag values if provided
	if flagValues != nil {
		if flagValues.LogLevel != "" {
			resolved.LogLevel = flagValues.LogLevel
		}
		if flagValues.Mode != "" {
			resolved.Controller.Mode = flagValues.Mode
		}
		if flagValues.ConfigFile != "" {
			resolved.Controller.ConfigFile = flagValues.ConfigFile
		}
		if flagValues.ResyncPeriod != 0 {
			resolved.Controller.ResyncPeriod = flagValues.ResyncPeriod
		}
		if flagValues.Namespace != "" {
			resolved.Controller.Single.Namespace = flagValues.Namespace
		}
		if flagValues.MetricsPort != 0 {
			resolved.Controller.Single.MetricsPort = flagValues.MetricsPort
		}
		if flagValues.HealthPort != 0 {
			resolved.Controller.Single.HealthPort = flagValues.HealthPort
		}
		if flagValues.LeaderElectionEnabled != nil {
			resolved.Controller.Single.LeaderElection.Enabled = *flagValues.LeaderElectionEnabled
		}
		if flagValues.LeaderElectionID != "" {
			resolved.Controller.Single.LeaderElection.ID = flagValues.LeaderElectionID
		}
		if flagValues.LeaderElectionNamespace != "" {
			resolved.Controller.Single.LeaderElection.Namespace = flagValues.LeaderElectionNamespace
		}
		if flagValues.TestConnectivity != nil {
			resolved.MultiCluster.TestConnectivity = *flagValues.TestConnectivity
		}
	}
	
	// Override with environment variables (highest priority)
	if envValue := os.Getenv("K6S_LOG_LEVEL"); envValue != "" {
		resolved.LogLevel = envValue
	}
	if envValue := os.Getenv("K6S_CONTROLLER_MODE"); envValue != "" {
		resolved.Controller.Mode = envValue
	}
	if envValue := os.Getenv("K6S_CONTROLLER_CONFIG_FILE"); envValue != "" {
		resolved.Controller.ConfigFile = envValue
	}
	if envValue := os.Getenv("K6S_CONTROLLER_RESYNC_PERIOD"); envValue != "" {
		if duration, err := time.ParseDuration(envValue); err == nil {
			resolved.Controller.ResyncPeriod = duration
		}
	}
	if envValue := os.Getenv("K6S_CONTROLLER_NAMESPACE"); envValue != "" {
		resolved.Controller.Single.Namespace = envValue
	}
	if envValue := os.Getenv("K6S_CONTROLLER_METRICS_PORT"); envValue != "" {
		if port, err := strconv.Atoi(envValue); err == nil {
			resolved.Controller.Single.MetricsPort = port
		}
	}
	if envValue := os.Getenv("K6S_CONTROLLER_HEALTH_PORT"); envValue != "" {
		if port, err := strconv.Atoi(envValue); err == nil {
			resolved.Controller.Single.HealthPort = port
		}
	}
	if envValue := os.Getenv("K6S_CONTROLLER_LEADER_ELECTION_ENABLED"); envValue != "" {
		if enabled, err := strconv.ParseBool(envValue); err == nil {
			resolved.Controller.Single.LeaderElection.Enabled = enabled
		}
	}
	if envValue := os.Getenv("K6S_CONTROLLER_LEADER_ELECTION_ID"); envValue != "" {
		resolved.Controller.Single.LeaderElection.ID = envValue
	}
	if envValue := os.Getenv("K6S_CONTROLLER_LEADER_ELECTION_NAMESPACE"); envValue != "" {
		resolved.Controller.Single.LeaderElection.Namespace = envValue
	}
	if envValue := os.Getenv("K6S_MULTI_CLUSTER_TEST_CONNECTIVITY"); envValue != "" {
		if testConn, err := strconv.ParseBool(envValue); err == nil {
			resolved.MultiCluster.TestConnectivity = testConn
		}
	}
	
	return resolved
}

// FlagValues holds flag values passed from CLI
type FlagValues struct {
	LogLevel                   string
	Mode                       string
	ConfigFile                 string
	ResyncPeriod               time.Duration
	Namespace                  string
	MetricsPort                int
	HealthPort                 int
	LeaderElectionEnabled      *bool
	LeaderElectionID           string
	LeaderElectionNamespace    string
	TestConnectivity           *bool
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate log level
	validLogLevels := []string{"trace", "debug", "info", "warn", "error"}
	logLevel := strings.ToLower(c.LogLevel)
	valid := false
	for _, level := range validLogLevels {
		if logLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)", c.LogLevel, strings.Join(validLogLevels, ", "))
	}
	
	// Validate controller mode
	if c.Controller.Mode != "single" && c.Controller.Mode != "multi" {
		return fmt.Errorf("invalid controller mode: %s (must be 'single' or 'multi')", c.Controller.Mode)
	}
	
	// Validate ports
	if c.Controller.Single.MetricsPort < 1 || c.Controller.Single.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port: %d (must be between 1 and 65535)", c.Controller.Single.MetricsPort)
	}
	if c.Controller.Single.HealthPort < 1 || c.Controller.Single.HealthPort > 65535 {
		return fmt.Errorf("invalid health port: %d (must be between 1 and 65535)", c.Controller.Single.HealthPort)
	}
	
	// Validate resync period
	if c.Controller.ResyncPeriod < time.Second {
		return fmt.Errorf("invalid resync period: %v (must be at least 1 second)", c.Controller.ResyncPeriod)
	}
	
	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".k6s", "k6s.yaml")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir(configPath string) error {
	dir := filepath.Dir(configPath)
	return os.MkdirAll(dir, 0755)
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configFile string) error {
	// If no config file specified, use default
	if configFile == "" {
		configFile = GetDefaultConfigPath()
	}
	
	// Ensure config directory exists
	if err := EnsureConfigDir(configFile); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	
	// Write to file
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	
	return nil
}
