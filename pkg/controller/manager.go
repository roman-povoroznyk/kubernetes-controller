package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/cluster"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Manager wraps the controller-runtime manager with additional functionality
type Manager struct {
	mgr         manager.Manager
	registry    cluster.ClusterRegistry
	multiMgr    *MultiClusterManager
	log         logr.Logger
	config      *config.Config
	mode        string // "single" or "multi"
}

// NewManager creates a new controller manager
func NewManager(cfg *config.Config, mode string) (*Manager, error) {
	log := logger.WithComponent("controller-manager")
	
	// Create cluster registry
	clusterRegistry := cluster.NewInMemoryClusterRegistry()
	
	// Add default cluster if none configured
	if len(cfg.MultiCluster.Clusters) == 0 {
		defaultCluster := cluster.NewClusterConfig("default")
		if err := clusterRegistry.AddCluster("default", defaultCluster); err != nil {
			return nil, fmt.Errorf("failed to add default cluster: %w", err)
		}
	} else {
		// Add configured clusters
		for _, clusterConfig := range cfg.MultiCluster.Clusters {
			clusterClient := &cluster.ClusterConfig{
				Name:       clusterConfig.Name,
				KubeConfig: clusterConfig.KubeConfig,
				Context:    clusterConfig.Context,
				Namespace:  clusterConfig.Namespace,
				Enabled:    clusterConfig.Enabled,
				Primary:    clusterConfig.Primary,
			}
			if err := clusterRegistry.AddCluster(clusterConfig.Name, clusterClient); err != nil {
				return nil, fmt.Errorf("failed to add cluster %s: %w", clusterConfig.Name, err)
			}
		}
	}
	
	// Determine mode
	if mode == "" {
		mode = cfg.Controller.Mode
		if mode == "" {
			mode = "single"
		}
	}
	
	log.Info("Controller manager mode", map[string]interface{}{
		"mode": mode,
	})
	
	var mgr manager.Manager
	var multiMgr *MultiClusterManager
	
	if mode == "multi" {
		// Multi-cluster mode - create multi-cluster manager
		multiMgr = NewMultiClusterManager(clusterRegistry, cfg.Controller.Single.Namespace, 1)
		log.Info("Multi-cluster manager created", nil)
	} else {
		// Single cluster mode - create standard manager
		var err error
		mgr, err = createSingleClusterManager(cfg, log)
		if err != nil {
			return nil, fmt.Errorf("failed to create single cluster manager: %w", err)
		}
		log.Info("Single cluster manager created", nil)
	}
	
	return &Manager{
		mgr:      mgr,
		registry: clusterRegistry,
		multiMgr: multiMgr,
		log:      log.GetLogr(),
		config:   cfg,
		mode:     mode,
	}, nil
}

// createSingleClusterManager creates a manager for single cluster mode
func createSingleClusterManager(cfg *config.Config, log *logger.Logger) (manager.Manager, error) {
	log.Info("Creating single cluster manager", nil)
	
	// Get REST config - for now use default
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Info("In-cluster config failed, trying out-of-cluster config", map[string]interface{}{"error": err.Error()})
		// Try out-of-cluster config
		restConfig, err = ctrl.GetConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
		}
	}
	
	log.Info("Kubernetes config obtained", map[string]interface{}{"host": restConfig.Host})
	
	// Create manager options
	opts := ctrl.Options{
		Scheme: runtime.NewScheme(),
		Metrics: server.Options{
			BindAddress: fmt.Sprintf(":%d", cfg.Controller.Single.MetricsPort),
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443, // Default webhook port
		}),
		HealthProbeBindAddress: fmt.Sprintf(":%d", cfg.Controller.Single.HealthPort),
		LeaderElection:         cfg.Controller.Single.LeaderElection.Enabled,
		LeaderElectionID:       cfg.Controller.Single.LeaderElection.ID,
		LeaderElectionNamespace: cfg.Controller.Single.LeaderElection.Namespace,
		Logger:                 log.GetLogr(),
	}
	
	log.Info("Manager options configured", map[string]interface{}{
		"metrics_port": cfg.Controller.Single.MetricsPort,
		"health_port": cfg.Controller.Single.HealthPort,
		"leader_election": cfg.Controller.Single.LeaderElection.Enabled,
		"leader_election_namespace": cfg.Controller.Single.LeaderElection.Namespace,
		"namespace": cfg.Controller.Single.Namespace,
	})
	
	// Add namespace filter if specified
	if cfg.Controller.Single.Namespace != "" {
		opts.Cache.DefaultNamespaces = map[string]cache.Config{
			cfg.Controller.Single.Namespace: {},
		}
		log.Info("Added namespace filter", map[string]interface{}{"namespace": cfg.Controller.Single.Namespace})
	}
	
	// Add schemes
	log.Info("Adding schemes to manager", nil)
	if err := appsv1.AddToScheme(opts.Scheme); err != nil {
		return nil, fmt.Errorf("failed to add apps/v1 scheme: %w", err)
	}
	
	// Create manager
	log.Info("Creating controller-runtime manager", nil)
	mgr, err := ctrl.NewManager(restConfig, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}
	log.Info("Controller-runtime manager created successfully", nil)
	
	// Add deployment reconciler
	log.Info("Adding deployment reconciler to manager", nil)
	if err := AddToManager(mgr, "default", cfg.Controller.Single.Namespace, 1); err != nil {
		return nil, fmt.Errorf("failed to add deployment controller: %w", err)
	}
	log.Info("Deployment reconciler added successfully", nil)
	
	// Add health checks
	log.Info("Adding health checks", nil)
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to add health check: %w", err)
	}
	
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to add ready check: %w", err)
	}
	log.Info("Health checks added successfully", nil)
	
	return mgr, nil
}

// Start starts the controller manager
func (m *Manager) Start(ctx context.Context) error {
	m.log.Info("Starting controller manager", "mode", m.mode)
	
	if m.mode == "multi" {
		// Multi-cluster mode
		return m.multiMgr.Start(ctx)
	} else {
		// Single cluster mode
		return m.mgr.Start(ctx)
	}
}

// Stop stops the controller manager
func (m *Manager) Stop() error {
	m.log.Info("Stopping controller manager")
	
	if m.mode == "multi" && m.multiMgr != nil {
		// Multi-cluster manager doesn't have a direct stop method
		// It will stop when the context is cancelled
		return nil
	}
	
	// For single cluster mode, the manager will stop when context is cancelled
	return nil
}

// GetMetrics returns metrics for the controller manager
func (m *Manager) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"mode":           m.mode,
		"namespace":      m.config.Controller.Single.Namespace,
		"concurrency":    1, // Default concurrency
		"leader_election": m.config.Controller.Single.LeaderElection.Enabled,
	}
	
	if m.mode == "multi" && m.multiMgr != nil {
		// Add multi-cluster metrics
		multiMetrics := m.multiMgr.GetMetrics()
		for k, v := range multiMetrics {
			metrics[k] = v
		}
	} else {
		// Add single cluster metrics
		clusters := m.registry.GetEnabledClusters()
		metrics["clusters"] = len(clusters)
		
		if len(clusters) > 0 {
			clusterNames := make([]string, 0, len(clusters))
			for name := range clusters {
				clusterNames = append(clusterNames, name)
			}
			metrics["cluster_names"] = clusterNames
		}
	}
	
	return metrics
}

// AddCluster adds a new cluster to the manager (multi-cluster mode only)
func (m *Manager) AddCluster(clusterName string, clusterConfig *cluster.ClusterConfig) error {
	if m.mode != "multi" {
		return fmt.Errorf("add cluster operation only supported in multi-cluster mode")
	}
	
	if m.multiMgr == nil {
		return fmt.Errorf("multi-cluster manager not initialized")
	}
	
	// Add to registry
	return m.registry.AddCluster(clusterName, clusterConfig)
}

// RemoveCluster removes a cluster from the manager (multi-cluster mode only)
func (m *Manager) RemoveCluster(clusterName string) error {
	if m.mode != "multi" {
		return fmt.Errorf("remove cluster operation only supported in multi-cluster mode")
	}
	
	if m.multiMgr == nil {
		return fmt.Errorf("multi-cluster manager not initialized")
	}
	
	// Remove from registry
	return m.registry.RemoveCluster(clusterName)
}

// GetClusterStatus returns the status of all clusters
func (m *Manager) GetClusterStatus() map[string]ClusterStatus {
	if m.mode == "multi" && m.multiMgr != nil {
		return m.multiMgr.GetClusterStatus()
	}
	
	// For single cluster mode, return basic status
	status := make(map[string]ClusterStatus)
	clusters := m.registry.GetEnabledClusters()
	
	for name := range clusters {
		status[name] = ClusterStatus{
			Name:      name,
			Ready:     true, // Assume ready in single cluster mode
			StartTime: time.Now(),
		}
	}
	
	return status
}

// GetRegistry returns the cluster registry
func (m *Manager) GetRegistry() cluster.ClusterRegistry {
	return m.registry
}

// GetMode returns the current mode
func (m *Manager) GetMode() string {
	return m.mode
}
