package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/controller"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// controllerCmd represents the controller command group
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Kubernetes controller management commands",
	Long: `Commands for managing the Kubernetes deployment controller.

The controller supports both single-cluster and multi-cluster modes:
- Single-cluster: Uses controller-runtime for watching a single cluster
- Multi-cluster: Uses custom informers for watching multiple clusters

Features:
• Real-time deployment monitoring with structured logging
• Leader election for high availability
• Prometheus metrics endpoint
• Health and readiness probes
• Graceful shutdown handling`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runController(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// startCmd represents the start controller command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the deployment controller",
	Long: `Start the Kubernetes controller to watch and log deployment events.

The controller watches for Deployment resources and logs all events including:
- Deployment creation, updates, and deletions
- Replica count changes
- Image updates
- Label and annotation changes

Examples:
  # Start single-cluster controller
  k6s controller start

  # Start with specific namespace
  k6s controller start --namespace production

  # Start multi-cluster controller
  k6s controller start --mode multi --config ~/.k6s/clusters.yaml

  # Start with debug logging and custom metrics port
  k6s controller start --log-level debug --metrics-port 9090`,
	RunE: runController,
}

var (
	// Controller mode
	controllerMode string

	// Single cluster configuration
	namespace               string
	metricsPort            int
	healthPort             int
	enableLeaderElection   bool
	leaderElectionID       string
	leaderElectionNamespace string

	// Multi-cluster configuration
	configFile   string
	resyncPeriod time.Duration

	// Common options
	kubeconfig string
	inCluster  bool
)

func init() {
	// Add controller command to root
	rootCmd.AddCommand(controllerCmd)
	
	// Add subcommands
	controllerCmd.AddCommand(startCmd)

	// Controller mode flags
	startCmd.Flags().StringVar(&controllerMode, "mode", "single", "controller mode (single, multi)")
	_ = viper.BindPFlag("controller.mode", startCmd.Flags().Lookup("mode"))

	// Single cluster flags
	startCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace to watch (empty = all namespaces)")
	startCmd.Flags().IntVar(&metricsPort, "metrics-port", 8080, "port for metrics endpoint")
	startCmd.Flags().IntVar(&healthPort, "health-port", 8081, "port for health probes")
	startCmd.Flags().BoolVar(&enableLeaderElection, "enable-leader-election", true, "enable leader election for controller manager")
	startCmd.Flags().StringVar(&leaderElectionID, "leader-election-id", "k6s-controller", "leader election ID")
	startCmd.Flags().StringVar(&leaderElectionNamespace, "leader-election-namespace", "default", "namespace for leader election")

	// Multi-cluster flags
	startCmd.Flags().StringVar(&configFile, "config-file", "", "path to multi-cluster configuration file")
	startCmd.Flags().DurationVar(&resyncPeriod, "resync-period", 30*time.Second, "resync period for informers")

	// Common flags
	startCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file (default: auto-detect)")
	startCmd.Flags().BoolVar(&inCluster, "in-cluster", false, "use in-cluster configuration")

	// Bind flags to viper
	_ = viper.BindPFlag("controller.single.namespace", startCmd.Flags().Lookup("namespace"))
	_ = viper.BindPFlag("controller.single.metrics_port", startCmd.Flags().Lookup("metrics-port"))
	_ = viper.BindPFlag("controller.single.health_port", startCmd.Flags().Lookup("health-port"))
	_ = viper.BindPFlag("controller.single.leader_election.enabled", startCmd.Flags().Lookup("enable-leader-election"))
	_ = viper.BindPFlag("controller.single.leader_election.id", startCmd.Flags().Lookup("leader-election-id"))
	_ = viper.BindPFlag("controller.single.leader_election.namespace", startCmd.Flags().Lookup("leader-election-namespace"))
	_ = viper.BindPFlag("controller.config_file", startCmd.Flags().Lookup("config-file"))
	_ = viper.BindPFlag("controller.resync_period", startCmd.Flags().Lookup("resync-period"))
	_ = viper.BindPFlag("kubeconfig", startCmd.Flags().Lookup("kubeconfig"))
	_ = viper.BindPFlag("in-cluster", startCmd.Flags().Lookup("in-cluster"))
}

func runController(cmd *cobra.Command, args []string) error {
	// Initialize logger with config from viper
	loggerConfig := logger.Config{
		Level:       viper.GetString("log.level"),
		Format:      viper.GetString("log.format"),
		Development: viper.GetBool("log.development"),
		Caller:      viper.GetBool("log.caller"),
	}
	logger.Init(loggerConfig)

	log := logger.WithComponent("controller-cmd")

	// Load configuration
	configPath := cfgFile
	if configPath == "" {
		configPath = viper.GetString("config")
	}
	log.Info("Loading configuration", map[string]interface{}{
		"config_path": configPath,
	})
	
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	validator := config.NewConfigValidator(cfg)
	if err := validator.ValidateAll(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Override with command-line flags
	if cmd.Flags().Changed("namespace") {
		cfg.Controller.Single.Namespace = viper.GetString("controller.single.namespace")
	}
	if cmd.Flags().Changed("metrics-port") {
		cfg.Controller.Single.MetricsPort = viper.GetInt("controller.single.metrics_port")
	}
	if cmd.Flags().Changed("health-port") {
		cfg.Controller.Single.HealthPort = viper.GetInt("controller.single.health_port")
	}
	if cmd.Flags().Changed("enable-leader-election") {
		cfg.Controller.Single.LeaderElection.Enabled = viper.GetBool("controller.single.leader_election.enabled")
	}

	// Determine mode
	mode := viper.GetString("controller.mode")
	if mode == "" {
		mode = controllerMode
	}

	log.Info("Starting k6s controller", map[string]interface{}{
		"mode":       mode,
		"version":    Version,
		"namespace":  cfg.Controller.Single.Namespace,
		"metrics":    fmt.Sprintf(":%d", cfg.Controller.Single.MetricsPort),
		"health":     fmt.Sprintf(":%d", cfg.Controller.Single.HealthPort),
		"clusters":   len(cfg.MultiCluster.Clusters),
	})

	// Create controller manager
	mgr, err := controller.NewManager(cfg, mode)
	if err != nil {
		return fmt.Errorf("failed to create controller manager: %w", err)
	}

	// Setup signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Setup graceful shutdown
	go func() {
		<-ctx.Done()
		log.Info("Received shutdown signal, starting graceful shutdown...", nil)
		// Add graceful shutdown logic here
	}()

	// Start the manager
	log.Info("Starting controller manager", nil)
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("controller manager failed: %w", err)
	}

	log.Info("Controller manager stopped", nil)
	return nil
}


