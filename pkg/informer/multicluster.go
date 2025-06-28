package informer

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClusterConfig represents configuration for a cluster
type ClusterConfig struct {
	Name       string `mapstructure:"name"`
	Kubeconfig string `mapstructure:"kubeconfig"`
	Context    string `mapstructure:"context"`
	Enabled    bool   `mapstructure:"enabled"`
}

// MultiClusterManager manages informers across multiple clusters
type MultiClusterManager struct {
	clusters map[string]*ClusterInformers
	config   *InformerConfig
	mu       sync.RWMutex
}

// ClusterInformers holds informers for a specific cluster
type ClusterInformers struct {
	Name               string
	Clientset          kubernetes.Interface
	DeploymentInformer *DeploymentInformer
	ctx                context.Context
	cancel             context.CancelFunc
}

// NewMultiClusterManager creates a new multi-cluster manager
func NewMultiClusterManager(config *InformerConfig) *MultiClusterManager {
	return &MultiClusterManager{
		clusters: make(map[string]*ClusterInformers),
		config:   config,
	}
}

// AddCluster adds a cluster to be managed
func (mcm *MultiClusterManager) AddCluster(clusterConfig ClusterConfig) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	if !clusterConfig.Enabled {
		log.Info().Str("cluster", clusterConfig.Name).Msg("Cluster is disabled, skipping")
		return nil
	}

	// Build config from kubeconfig
	var config *rest.Config
	var err error

	if clusterConfig.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", clusterConfig.Kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return fmt.Errorf("failed to build config for cluster %s: %w", clusterConfig.Name, err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset for cluster %s: %w", clusterConfig.Name, err)
	}

	// Create context for this cluster
	ctx, cancel := context.WithCancel(context.Background())

	// Create informers
	deploymentInformer := NewDeploymentInformer(clientset, mcm.config)

	clusterInformers := &ClusterInformers{
		Name:               clusterConfig.Name,
		Clientset:          clientset,
		DeploymentInformer: deploymentInformer,
		ctx:                ctx,
		cancel:             cancel,
	}

	mcm.clusters[clusterConfig.Name] = clusterInformers

	log.Info().Str("cluster", clusterConfig.Name).Msg("Added cluster to multi-cluster manager")
	return nil
}

// Start starts all cluster informers
func (mcm *MultiClusterManager) Start() error {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	for name, cluster := range mcm.clusters {
		log.Info().Str("cluster", name).Msg("Starting cluster informers")

		if cluster.DeploymentInformer != nil {
			go func(ci *ClusterInformers) {
				if err := ci.DeploymentInformer.Start(ci.ctx); err != nil {
					log.Error().Err(err).Str("cluster", ci.Name).Msg("Failed to start deployment informer")
				}
			}(cluster)
		}
	}

	log.Info().Int("clusters", len(mcm.clusters)).Msg("Started multi-cluster informers")
	return nil
}

// Stop stops all cluster informers
func (mcm *MultiClusterManager) Stop() {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	for name, cluster := range mcm.clusters {
		log.Info().Str("cluster", name).Msg("Stopping cluster informers")
		cluster.cancel()
	}
}

// GetClusterNames returns list of managed cluster names
func (mcm *MultiClusterManager) GetClusterNames() []string {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	names := make([]string, 0, len(mcm.clusters))
	for name := range mcm.clusters {
		names = append(names, name)
	}
	return names
}

// GetClusterClientset returns clientset for specific cluster
func (mcm *MultiClusterManager) GetClusterClientset(clusterName string) (kubernetes.Interface, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	cluster, exists := mcm.clusters[clusterName]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", clusterName)
	}

	return cluster.Clientset, nil
}
