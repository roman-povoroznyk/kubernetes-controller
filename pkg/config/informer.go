package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// InformerConfig contains configuration for deployment informer
type InformerConfig struct {
	// ResyncPeriod is the period for full resync of the informer cache
	ResyncPeriod time.Duration `mapstructure:"resync_period" yaml:"resync_period" json:"resync_period"`
	
	// Namespace is the namespace to watch (empty means all namespaces)
	Namespace string `mapstructure:"namespace" yaml:"namespace" json:"namespace"`
	
	// EnableCustomLogic enables custom logic for deployment change analysis
	EnableCustomLogic bool `mapstructure:"enable_custom_logic" yaml:"enable_custom_logic" json:"enable_custom_logic"`
	
	// KubectlStyle enables kubectl-style output formatting
	KubectlStyle bool `mapstructure:"kubectl_style" yaml:"kubectl_style" json:"kubectl_style"`
	
	// LabelSelector is a label selector to filter deployments
	LabelSelector string `mapstructure:"label_selector" yaml:"label_selector" json:"label_selector"`
	
	// FieldSelector is a field selector to filter deployments
	FieldSelector string `mapstructure:"field_selector" yaml:"field_selector" json:"field_selector"`
	
	// WorkerPoolSize is the number of workers for processing events
	WorkerPoolSize int `mapstructure:"worker_pool_size" yaml:"worker_pool_size" json:"worker_pool_size"`
	
	// QueueSize is the size of the event queue
	QueueSize int `mapstructure:"queue_size" yaml:"queue_size" json:"queue_size"`
}

// WatchConfig contains configuration for watch mode
type WatchConfig struct {
	// PollInterval is the interval for polling changes in watch mode
	PollInterval time.Duration `mapstructure:"poll_interval" yaml:"poll_interval" json:"poll_interval"`
	
	// Timeout is the timeout for watch operations
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	
	// MaxRetries is the maximum number of retries for failed operations
	MaxRetries int `mapstructure:"max_retries" yaml:"max_retries" json:"max_retries"`
	
	// RetryBackoff is the backoff duration between retries
	RetryBackoff time.Duration `mapstructure:"retry_backoff" yaml:"retry_backoff" json:"retry_backoff"`
}

// Config is the main configuration structure
type Config struct {
	// Informer configuration
	Informer InformerConfig `mapstructure:"informer" yaml:"informer" json:"informer"`
	
	// Watch configuration
	Watch WatchConfig `mapstructure:"watch" yaml:"watch" json:"watch"`
	
	// LogLevel is the logging level
	LogLevel string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Informer: InformerConfig{
			ResyncPeriod:      10 * time.Minute,
			Namespace:         "",
			EnableCustomLogic: false,
			KubectlStyle:      false,
			LabelSelector:     "",
			FieldSelector:     "",
			WorkerPoolSize:    5,
			QueueSize:         100,
		},
		Watch: WatchConfig{
			PollInterval: 1 * time.Second,
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			RetryBackoff: 2 * time.Second,
		},
		LogLevel: "info",
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	cfg := DefaultConfig()
	
	// Create a new viper instance to avoid conflicts
	v := viper.New()
	
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Look for config file in various locations
		v.SetConfigName("k6s")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.k6s")
		v.AddConfigPath("/etc/k6s")
	}
	
	// Environment variable configuration
	v.SetEnvPrefix("K6S")
	v.AutomaticEnv()
	
	// Enable automatic replacement of . with _ in env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Explicitly bind environment variables
	var bindErrors []string
	
	if err := v.BindEnv("log_level", "K6S_LOG_LEVEL"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("log_level: %v", err))
	}
	if err := v.BindEnv("informer.resync_period", "K6S_INFORMER_RESYNC_PERIOD"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.resync_period: %v", err))
	}
	if err := v.BindEnv("informer.namespace", "K6S_INFORMER_NAMESPACE"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.namespace: %v", err))
	}
	if err := v.BindEnv("informer.enable_custom_logic", "K6S_INFORMER_ENABLE_CUSTOM_LOGIC"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.enable_custom_logic: %v", err))
	}
	if err := v.BindEnv("informer.kubectl_style", "K6S_INFORMER_KUBECTL_STYLE"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.kubectl_style: %v", err))
	}
	if err := v.BindEnv("informer.label_selector", "K6S_INFORMER_LABEL_SELECTOR"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.label_selector: %v", err))
	}
	if err := v.BindEnv("informer.field_selector", "K6S_INFORMER_FIELD_SELECTOR"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.field_selector: %v", err))
	}
	if err := v.BindEnv("informer.worker_pool_size", "K6S_INFORMER_WORKER_POOL_SIZE"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.worker_pool_size: %v", err))
	}
	if err := v.BindEnv("informer.queue_size", "K6S_INFORMER_QUEUE_SIZE"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("informer.queue_size: %v", err))
	}
	if err := v.BindEnv("watch.poll_interval", "K6S_WATCH_POLL_INTERVAL"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("watch.poll_interval: %v", err))
	}
	if err := v.BindEnv("watch.timeout", "K6S_WATCH_TIMEOUT"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("watch.timeout: %v", err))
	}
	if err := v.BindEnv("watch.max_retries", "K6S_WATCH_MAX_RETRIES"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("watch.max_retries: %v", err))
	}
	if err := v.BindEnv("watch.retry_backoff", "K6S_WATCH_RETRY_BACKOFF"); err != nil {
		bindErrors = append(bindErrors, fmt.Sprintf("watch.retry_backoff: %v", err))
	}
	
	// If there were any binding errors, return them
	if len(bindErrors) > 0 {
		return nil, fmt.Errorf("failed to bind environment variables: %s", strings.Join(bindErrors, ", "))
	}
	
	// Set defaults
	v.SetDefault("informer.resync_period", "10m")
	v.SetDefault("informer.namespace", "")
	v.SetDefault("informer.enable_custom_logic", false)
	v.SetDefault("informer.kubectl_style", false)
	v.SetDefault("informer.label_selector", "")
	v.SetDefault("informer.field_selector", "")
	v.SetDefault("informer.worker_pool_size", 5)
	v.SetDefault("informer.queue_size", 100)
	v.SetDefault("watch.poll_interval", "1s")
	v.SetDefault("watch.timeout", "30s")
	v.SetDefault("watch.max_retries", 3)
	v.SetDefault("watch.retry_backoff", "2s")
	v.SetDefault("log_level", "info")
	
	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}
	
	// Unmarshal configuration
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Informer.ResyncPeriod < 0 {
		return fmt.Errorf("informer.resync_period cannot be negative")
	}
	
	if c.Informer.WorkerPoolSize <= 0 {
		return fmt.Errorf("informer.worker_pool_size must be positive")
	}
	
	if c.Informer.QueueSize <= 0 {
		return fmt.Errorf("informer.queue_size must be positive")
	}
	
	if c.Watch.PollInterval <= 0 {
		return fmt.Errorf("watch.poll_interval must be positive")
	}
	
	if c.Watch.Timeout <= 0 {
		return fmt.Errorf("watch.timeout must be positive")
	}
	
	if c.Watch.MaxRetries < 0 {
		return fmt.Errorf("watch.max_retries cannot be negative")
	}
	
	if c.Watch.RetryBackoff < 0 {
		return fmt.Errorf("watch.retry_backoff cannot be negative")
	}
	
	return nil
}
