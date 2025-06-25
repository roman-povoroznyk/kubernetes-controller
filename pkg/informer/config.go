package informer

import (
	"time"
	"github.com/spf13/viper"
)

// InformerConfig holds configuration for informers
type InformerConfig struct {
	ResyncPeriod    time.Duration `mapstructure:"resync_period"`
	Workers         int           `mapstructure:"workers"`
	EnabledResources []string     `mapstructure:"enabled_resources"`
	Namespaces      []string      `mapstructure:"namespaces"`
}

// LoadInformerConfig loads informer configuration from viper
func LoadInformerConfig() (*InformerConfig, error) {
	viper.SetDefault("informer.resync_period", "30s")
	viper.SetDefault("informer.workers", 5)
	viper.SetDefault("informer.enabled_resources", []string{"deployments", "services", "pods"})
	viper.SetDefault("informer.namespaces", []string{"default"})

	var config InformerConfig
	if err := viper.UnmarshalKey("informer", &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// IsResourceEnabled checks if resource type is enabled
func (c *InformerConfig) IsResourceEnabled(resource string) bool {
	for _, r := range c.EnabledResources {
		if r == resource {
			return true
		}
	}
	return false
}

// IsNamespaceWatched checks if namespace is being watched
func (c *InformerConfig) IsNamespaceWatched(namespace string) bool {
	for _, ns := range c.Namespaces {
		if ns == namespace || ns == "*" {
			return true
		}
	}
	return false
}
