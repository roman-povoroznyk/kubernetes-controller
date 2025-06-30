package cluster

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
)

// ClusterRegistry defines the interface for managing cluster configurations
type ClusterRegistry interface {
	GetEnabledClusters() map[string]ClusterClient
	GetCluster(name string) (ClusterClient, bool)
	AddCluster(name string, config *ClusterConfig) error
	RemoveCluster(name string) error
	ListClusters() []string
}

// ClusterClient represents a client for a specific cluster
type ClusterClient interface {
	GetName() string
	GetRestConfig() (*rest.Config, error)
	GetKubernetesClient() (kubernetes.Interface, error)
	IsEnabled() bool
	TestConnection(ctx context.Context) error
}

// ClusterConfig represents the configuration for a single cluster
type ClusterConfig struct {
	Name       string `yaml:"name" json:"name"`
	KubeConfig string `yaml:"kubeconfig" json:"kubeconfig"`
	Context    string `yaml:"context" json:"context"`
	Namespace  string `yaml:"namespace" json:"namespace"`
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	Primary    bool   `yaml:"primary" json:"primary"`
	
	// Internal fields
	restConfig *rest.Config
	kubeClient kubernetes.Interface
}

// NewClusterConfig creates a new cluster configuration
func NewClusterConfig(name string) *ClusterConfig {
	return &ClusterConfig{
		Name:    name,
		Enabled: true,
	}
}

// GetName returns the cluster name
func (c *ClusterConfig) GetName() string {
	return c.Name
}

// GetRestConfig returns the REST configuration for this cluster
func (c *ClusterConfig) GetRestConfig() (*rest.Config, error) {
	if c.restConfig != nil {
		return c.restConfig, nil
	}
	
	if c.KubeConfig != "" {
		// Use specific kubeconfig file
		config, err := clientcmd.BuildConfigFromFlags("", c.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig %s: %w", c.KubeConfig, err)
		}
		
		if c.Context != "" {
			// Load the config file to get context
			configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: c.KubeConfig},
				&clientcmd.ConfigOverrides{CurrentContext: c.Context},
			)
			config, err = configLoader.ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to load context %s from kubeconfig %s: %w", c.Context, c.KubeConfig, err)
			}
		}
		
		c.restConfig = config
	} else {
		// Use default kubeconfig
		config, err := rest.InClusterConfig()
		if err != nil {
			// Try loading from default kubeconfig location
			config, err = clientcmd.BuildConfigFromFlags("", "")
			if err != nil {
				return nil, fmt.Errorf("failed to build config from default kubeconfig: %w", err)
			}
		}
		c.restConfig = config
	}
	
	return c.restConfig, nil
}

// GetKubernetesClient returns a Kubernetes client for this cluster
func (c *ClusterConfig) GetKubernetesClient() (kubernetes.Interface, error) {
	if c.kubeClient != nil {
		return c.kubeClient, nil
	}
	
	config, err := c.GetRestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}
	
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}
	
	c.kubeClient = client
	return client, nil
}

// IsEnabled returns whether the cluster is enabled
func (c *ClusterConfig) IsEnabled() bool {
	return c.Enabled
}

// TestConnection tests connectivity to the cluster
func (c *ClusterConfig) TestConnection(ctx context.Context) error {
	client, err := c.GetKubernetesClient()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client: %w", err)
	}
	
	// Try to get server version as a connectivity test
	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to cluster %s: %w", c.Name, err)
	}
	
	return nil
}

// InMemoryClusterRegistry is a simple in-memory implementation of ClusterRegistry
type InMemoryClusterRegistry struct {
	clusters map[string]*ClusterConfig
}

// NewInMemoryClusterRegistry creates a new in-memory cluster registry
func NewInMemoryClusterRegistry() *InMemoryClusterRegistry {
	return &InMemoryClusterRegistry{
		clusters: make(map[string]*ClusterConfig),
	}
}

// GetEnabledClusters returns all enabled clusters
func (r *InMemoryClusterRegistry) GetEnabledClusters() map[string]ClusterClient {
	enabled := make(map[string]ClusterClient)
	for name, config := range r.clusters {
		if config.Enabled {
			enabled[name] = config
		}
	}
	return enabled
}

// GetCluster returns a specific cluster by name
func (r *InMemoryClusterRegistry) GetCluster(name string) (ClusterClient, bool) {
	config, exists := r.clusters[name]
	return config, exists
}

// AddCluster adds a new cluster to the registry
func (r *InMemoryClusterRegistry) AddCluster(name string, config *ClusterConfig) error {
	if config == nil {
		return fmt.Errorf("cluster config cannot be nil")
	}
	
	if config.Name == "" {
		config.Name = name
	}
	
	r.clusters[name] = config
	return nil
}

// RemoveCluster removes a cluster from the registry
func (r *InMemoryClusterRegistry) RemoveCluster(name string) error {
	delete(r.clusters, name)
	return nil
}

// ListClusters returns a list of all cluster names
func (r *InMemoryClusterRegistry) ListClusters() []string {
	names := make([]string, 0, len(r.clusters))
	for name := range r.clusters {
		names = append(names, name)
	}
	return names
}
