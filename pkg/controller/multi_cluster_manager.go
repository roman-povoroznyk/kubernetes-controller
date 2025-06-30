package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/cluster"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// MultiClusterManager manages controllers across multiple clusters
type MultiClusterManager struct {
	registry    cluster.ClusterRegistry
	managers    map[string]manager.Manager
	reconcilers map[string]*DeploymentReconciler
	log         logr.Logger
	
	// Configuration
	namespace   string
	concurrency int
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mutex  sync.RWMutex
}

// NewMultiClusterManager creates a new multi-cluster manager
func NewMultiClusterManager(registry cluster.ClusterRegistry, namespace string, concurrency int) *MultiClusterManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &MultiClusterManager{
		registry:    registry,
		managers:    make(map[string]manager.Manager),
		reconcilers: make(map[string]*DeploymentReconciler),
		log:         logger.WithComponent("multi-cluster-manager").GetLogr(),
		namespace:   namespace,
		concurrency: concurrency,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the multi-cluster manager
func (m *MultiClusterManager) Start(ctx context.Context) error {
	m.log.Info("Starting multi-cluster manager", "namespace", m.namespace, "concurrency", m.concurrency)
	
	// Get enabled clusters
	clusters := m.registry.GetEnabledClusters()
	if len(clusters) == 0 {
		return fmt.Errorf("no enabled clusters found")
	}
	
	// Start managers for each cluster
	for clusterName, clusterConfig := range clusters {
		if err := m.startClusterManager(clusterName, clusterConfig); err != nil {
			m.log.Error(err, "Failed to start cluster manager", "cluster", clusterName)
			return fmt.Errorf("failed to start cluster manager %s: %w", clusterName, err)
		}
	}
	
	m.log.Info("Multi-cluster manager started", "clusters", len(clusters))
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Stop all managers
	m.log.Info("Stopping multi-cluster manager")
	m.cancel()
	m.wg.Wait()
	
	return nil
}

// startClusterManager starts a manager for a specific cluster
func (m *MultiClusterManager) startClusterManager(clusterName string, clusterConfig cluster.ClusterClient) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if manager already exists
	if _, exists := m.managers[clusterName]; exists {
		return fmt.Errorf("manager for cluster %s already exists", clusterName)
	}
	
	// Create REST config
	restConfig, err := clusterConfig.GetRestConfig()
	if err != nil {
		return fmt.Errorf("failed to get REST config for cluster %s: %w", clusterName, err)
	}
	
	// Create manager options
	opts := ctrl.Options{
		Scheme: runtime.NewScheme(),
		Metrics: server.Options{
			BindAddress: "0", // Disable metrics for individual cluster managers
		},
		HealthProbeBindAddress: "0", // Disable health probes for individual cluster managers
		LeaderElection:         false, // Leader election is handled at the multi-cluster level
		LeaderElectionID:       "",
		Logger:                 logger.WithCluster(clusterName).GetLogr(),
	}
	
	// Add namespace filter if specified
	if m.namespace != "" {
		opts.Cache.DefaultNamespaces = map[string]cache.Config{
			m.namespace: {},
		}
	}
	
	// Add schemes
	if err := appsv1.AddToScheme(opts.Scheme); err != nil {
		return fmt.Errorf("failed to add apps/v1 scheme: %w", err)
	}
	
	// Create manager
	mgr, err := ctrl.NewManager(restConfig, opts)
	if err != nil {
		return fmt.Errorf("failed to create manager for cluster %s: %w", clusterName, err)
	}
	
	// Create and add deployment reconciler
	reconciler := NewDeploymentReconciler(mgr, clusterName, m.namespace, m.concurrency)
	if err := reconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup deployment reconciler for cluster %s: %w", clusterName, err)
	}
	
	// Store manager and reconciler
	m.managers[clusterName] = mgr
	m.reconcilers[clusterName] = reconciler
	
	// Start manager in a goroutine
	m.wg.Add(1)
	go func(clusterName string, mgr manager.Manager) {
		defer m.wg.Done()
		
		m.log.Info("Starting cluster manager", "cluster", clusterName)
		if err := mgr.Start(m.ctx); err != nil {
			m.log.Error(err, "Cluster manager failed", "cluster", clusterName)
		}
		m.log.Info("Cluster manager stopped", "cluster", clusterName)
	}(clusterName, mgr)
	
	return nil
}

// AddCluster adds a new cluster to the multi-cluster manager
func (m *MultiClusterManager) AddCluster(clusterName string, clusterConfig cluster.ClusterClient) error {
	m.log.Info("Adding cluster", "cluster", clusterName)
	
	if err := m.startClusterManager(clusterName, clusterConfig); err != nil {
		return fmt.Errorf("failed to add cluster %s: %w", clusterName, err)
	}
	
	m.log.Info("Cluster added successfully", "cluster", clusterName)
	return nil
}

// RemoveCluster removes a cluster from the multi-cluster manager
func (m *MultiClusterManager) RemoveCluster(clusterName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.log.Info("Removing cluster", "cluster", clusterName)
	
	// Remove from maps
	delete(m.managers, clusterName)
	delete(m.reconcilers, clusterName)
	
	// Note: We don't actively stop the manager here as it's complex to do safely
	// The manager will stop when the main context is cancelled
	
	m.log.Info("Cluster removed", "cluster", clusterName)
	return nil
}

// GetClusterStatus returns the status of all clusters
func (m *MultiClusterManager) GetClusterStatus() map[string]ClusterStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	status := make(map[string]ClusterStatus)
	for clusterName, mgr := range m.managers {
		status[clusterName] = ClusterStatus{
			Name:      clusterName,
			Ready:     m.isManagerReady(mgr),
			StartTime: time.Now(), // TODO: Track actual start time
		}
	}
	
	return status
}

// isManagerReady checks if a manager is ready
func (m *MultiClusterManager) isManagerReady(mgr manager.Manager) bool {
	// Try to get client to check if manager is ready
	if mgr.GetClient() == nil {
		return false
	}
	
	// Additional readiness checks can be added here
	return true
}

// ClusterStatus represents the status of a cluster
type ClusterStatus struct {
	Name      string    `json:"name"`
	Ready     bool      `json:"ready"`
	StartTime time.Time `json:"start_time"`
}

// EnhancedMultiClusterReconciler is a single reconciler that handles multiple clusters
type EnhancedMultiClusterReconciler struct {
	registry cluster.ClusterRegistry
	log      logr.Logger
	
	// Configuration
	namespace   string
	concurrency int
}

// NewEnhancedMultiClusterReconciler creates a new enhanced multi-cluster reconciler
func NewEnhancedMultiClusterReconciler(registry cluster.ClusterRegistry, namespace string, concurrency int) *EnhancedMultiClusterReconciler {
	return &EnhancedMultiClusterReconciler{
		registry:    registry,
		log:         logger.WithComponent("enhanced-multi-cluster-reconciler").GetLogr(),
		namespace:   namespace,
		concurrency: concurrency,
	}
}

// SetupWithManager sets up the enhanced multi-cluster reconciler with a manager
func (r *EnhancedMultiClusterReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(r.createEventFilter()).
		Complete(r)
}

// createEventFilter creates event filters for the enhanced reconciler
func (r *EnhancedMultiClusterReconciler) createEventFilter() predicate.Predicate {
	return predicate.And(
		predicate.GenerationChangedPredicate{},
		predicate.ResourceVersionChangedPredicate{},
		r.createNamespaceFilter(),
	)
}

// createNamespaceFilter creates a namespace filter if specified
func (r *EnhancedMultiClusterReconciler) createNamespaceFilter() predicate.Predicate {
	if r.namespace == "" {
		return predicate.NewPredicateFuncs(func(object client.Object) bool {
			return true
		})
	}
	
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetNamespace() == r.namespace
	})
}

// Reconcile reconciles deployments across multiple clusters
func (r *EnhancedMultiClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.log.WithValues("deployment", req.NamespacedName)
	
	// For now, we'll log the event and return
	log.Info("Multi-cluster deployment event", 
		"namespace", req.Namespace, 
		"name", req.Name,
		"clusters", len(r.registry.GetEnabledClusters()))
	
	return ctrl.Result{}, nil
}

// GetMetrics returns metrics for the multi-cluster manager
func (m *MultiClusterManager) GetMetrics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	metrics := map[string]interface{}{
		"total_clusters":   len(m.managers),
		"active_clusters":  0,
		"cluster_status":   m.GetClusterStatus(),
		"namespace_filter": m.namespace,
		"concurrency":      m.concurrency,
	}
	
	// Count active clusters
	for _, mgr := range m.managers {
		if m.isManagerReady(mgr) {
			metrics["active_clusters"] = metrics["active_clusters"].(int) + 1
		}
	}
	
	return metrics
}
