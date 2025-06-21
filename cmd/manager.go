package cmd

import (
	"github.com/roman-povoroznyk/k6s/pkg/controller"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
)

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start the controller manager",
	Long: `Start the controller manager with controller-runtime.
This command starts a Kubernetes controller that watches deployment resources
and reconciles their state using the controller-runtime framework.

The manager provides:
- Deployment controller with reconciliation logic
- Leader election support for high availability
- Metrics and health endpoints
- Graceful shutdown handling

Examples:
  k6s manager                                    # Start with default settings
  k6s manager --metrics-addr :8081              # Custom metrics port
  k6s manager --enable-leader-election          # Enable leader election
  k6s manager --health-probe-addr :8082         # Custom health probe port`,
	RunE: runManager,
}

var (
	metricsAddr          string
	enableLeaderElection bool
	leaderElectionID     string
	probeAddr            string
	managerNamespace     string
)

func init() {
	rootCmd.AddCommand(managerCmd)
	
	managerCmd.Flags().StringVar(&metricsAddr, "metrics-addr", ":8081", 
		"The address the metric endpoint binds to")
	managerCmd.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager")
	managerCmd.Flags().StringVar(&leaderElectionID, "leader-election-id", "k6s-controller-leader",
		"The name of the leader election ID")
	managerCmd.Flags().StringVar(&probeAddr, "health-probe-addr", ":8082",
		"The address the probe endpoint binds to")
	managerCmd.Flags().StringVar(&managerNamespace, "watch-namespace", "",
		"Namespace to watch for resources. If empty, watches all namespaces")
}

func runManager(cmd *cobra.Command, args []string) error {
	logger.Info("Starting k6s controller manager", map[string]interface{}{
		"metrics_addr":       metricsAddr,
		"leader_election":    enableLeaderElection,
		"leader_election_id": leaderElectionID,
		"probe_addr":         probeAddr,
		"watch_namespace":    managerNamespace,
	})

	// Create manager configuration
	cfg := controller.ManagerConfig{
		MetricsAddr:          metricsAddr,
		EnableLeaderElection: enableLeaderElection,
		LeaderElectionID:     leaderElectionID,
		ProbeAddr:            probeAddr,
		Namespace:            managerNamespace,
	}

	// Create manager
	mgr, err := controller.NewManager(cfg)
	if err != nil {
		logger.Error("Failed to create manager", err, map[string]interface{}{})
		return err
	}

	// Setup signal handler for graceful shutdown
	ctx := controller.SetupSignalHandler()

	logger.Info("Controller manager configured successfully", map[string]interface{}{
		"metrics_endpoint": "http://localhost" + metricsAddr + "/metrics",
		"health_endpoint":  "http://localhost" + probeAddr + "/healthz",
		"ready_endpoint":   "http://localhost" + probeAddr + "/readyz",
	})

	// Start manager
	if err := mgr.Start(ctx); err != nil {
		logger.Error("Failed to start manager", err, map[string]interface{}{})
		return err
	}

	return nil
}
