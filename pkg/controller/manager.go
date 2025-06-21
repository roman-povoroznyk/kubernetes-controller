package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// ManagerConfig holds configuration for the controller manager
type ManagerConfig struct {
	MetricsAddr          string
	EnableLeaderElection bool
	LeaderElectionID     string
	ProbeAddr            string
	Namespace            string
	SyncPeriod           time.Duration
}

// DefaultManagerConfig returns default configuration
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		MetricsAddr:          ":8081",
		EnableLeaderElection: false,
		LeaderElectionID:     "k6s-controller-leader",
		ProbeAddr:            ":8082",
		Namespace:            "",
		SyncPeriod:           time.Minute * 10,
	}
}

// Manager wraps controller-runtime manager
type Manager struct {
	ctrl.Manager
	config             ManagerConfig
	deploymentController *DeploymentController
}

// NewManager creates a new controller manager
func NewManager(cfg ManagerConfig) (*Manager, error) {
	// Setup logger for controller-runtime
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add client-go scheme: %w", err)
	}
	
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add apps/v1 scheme: %w", err)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: cfg.MetricsAddr,
		},
		HealthProbeBindAddress: cfg.ProbeAddr,
		LeaderElection:         cfg.EnableLeaderElection,
		LeaderElectionID:       cfg.LeaderElectionID,
		Controller: config.Controller{
			GroupKindConcurrency: map[string]int{
				"Deployment.apps": 2,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to add health check: %w", err)
	}
	
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to add ready check: %w", err)
	}

	// Create deployment controller
	deploymentController := NewDeploymentController(mgr)
	if err := deploymentController.SetupWithManager(mgr); err != nil {
		return nil, fmt.Errorf("failed to setup deployment controller: %w", err)
	}

	return &Manager{
		Manager:              mgr,
		config:               cfg,
		deploymentController: deploymentController,
	}, nil
}

// Start starts the manager
func (m *Manager) Start(ctx context.Context) error {
	logger.Info("Starting controller manager", map[string]interface{}{
		"metrics_addr":           m.config.MetricsAddr,
		"health_probe_addr":      m.config.ProbeAddr,
		"leader_election":        m.config.EnableLeaderElection,
		"leader_election_id":     m.config.LeaderElectionID,
		"namespace":              m.config.Namespace,
		"sync_period":            m.config.SyncPeriod.String(),
	})

	return m.Manager.Start(ctx)
}

// GetDeploymentController returns the deployment controller
func (m *Manager) GetDeploymentController() *DeploymentController {
	return m.deploymentController
}

// SetupSignalHandler sets up signal handling for graceful shutdown
func SetupSignalHandler() context.Context {
	return ctrl.SetupSignalHandler()
}

// GetManagerFromEnv creates manager configuration from environment variables
func GetManagerFromEnv() ManagerConfig {
	cfg := DefaultManagerConfig()
	
	if addr := os.Getenv("METRICS_ADDR"); addr != "" {
		cfg.MetricsAddr = addr
	}
	
	if addr := os.Getenv("HEALTH_PROBE_ADDR"); addr != "" {
		cfg.ProbeAddr = addr
	}
	
	if os.Getenv("ENABLE_LEADER_ELECTION") == "true" {
		cfg.EnableLeaderElection = true
	}
	
	if id := os.Getenv("LEADER_ELECTION_ID"); id != "" {
		cfg.LeaderElectionID = id
	}
	
	if ns := os.Getenv("WATCH_NAMESPACE"); ns != "" {
		cfg.Namespace = ns
	}
	
	if period := os.Getenv("SYNC_PERIOD"); period != "" {
		if d, err := time.ParseDuration(period); err == nil {
			cfg.SyncPeriod = d
		}
	}
	
	return cfg
}
