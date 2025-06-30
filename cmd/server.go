package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	serverPort         int
	enableInformer     bool
	informerNamespace  string
	informerResyncTime string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the HTTP server with optional deployment informer",
	Long: `Start the HTTP server for handling API requests.
	
The server provides REST API endpoints for health checks, version info,
and deployment resources. When informer is enabled, it provides real-time
deployment data via /api/v1/deployments endpoints.

Examples:
  k6s server                                    # start server on default port 8080
  k6s server --port 9090                       # start server on port 9090
  k6s server --enable-informer                 # start server with deployment informer
  k6s server --enable-informer --namespace=prod # start server with informer for specific namespace
  K6S_SERVER_PORT=8081 k6s server              # start server using env var`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get port from viper (supports env vars, config files, and flags)
		port := viper.GetInt("server.port")
		if port == 0 {
			port = serverPort // fallback to flag value
		}
		
		logger.Info("Starting k6s server", map[string]interface{}{
			"component":      "server",
			"port":           port,
			"version":        Version,
			"enable_informer": enableInformer,
		})

		// Create server
		srv := server.New(port)
		
		// Setup informer if enabled
		if enableInformer {
			if err := setupDeploymentInformer(srv); err != nil {
				logger.Fatal("Failed to setup deployment informer", err, nil)
			}
		}

		// Setup graceful shutdown
		// Start server in goroutine
		serverError := make(chan error, 1)
		go func() {
			serverError <- srv.Start()
		}()

		// Wait for interrupt signal
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-serverError:
			if err != nil {
				logger.Fatal("Server failed to start", err, map[string]interface{}{
					"port": port,
				})
			}
		case <-interrupt:
			logger.Info("Received interrupt signal, shutting down server", nil)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	
	// Add server-specific flags
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "server port")
	serverCmd.Flags().BoolVar(&enableInformer, "enable-informer", false, "enable deployment informer for API endpoints")
	serverCmd.Flags().StringVar(&informerNamespace, "namespace", "", "kubernetes namespace to watch (empty = all namespaces)")
	serverCmd.Flags().StringVar(&informerResyncTime, "resync-period", "", "informer cache resync period (e.g., 5m, 30s)")
	
	// Bind flags to viper for environment variable support
	if err := viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port")); err != nil {
		logger.Error("Failed to bind port flag", err, nil)
	}
	
	// Allow environment variables
	if err := viper.BindEnv("server.port", "K6S_SERVER_PORT"); err != nil {
		logger.Error("Failed to bind server port env", err, nil)
	}
}

// setupDeploymentInformer creates and starts deployment informer for server
func setupDeploymentInformer(srv *server.Server) error {
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using defaults", map[string]interface{}{
			"config_file": configFile,
			"error":       err.Error(),
		})
		cfg = config.DefaultConfig()
	}

	// Override with command line flags
	if informerNamespace != "" {
		cfg.Controller.Single.Namespace = informerNamespace
	}
	if informerResyncTime != "" {
		if duration, parseErr := time.ParseDuration(informerResyncTime); parseErr == nil {
			cfg.Controller.ResyncPeriod = duration
		} else {
			logger.Warn("Invalid resync period, using default", map[string]interface{}{
				"provided": informerResyncTime,
				"default":  cfg.Controller.ResyncPeriod,
				"error":    parseErr.Error(),
			})
		}
	}

	// Create Kubernetes client
	client, err := kubernetes.NewClient("")
	if err != nil {
		return err
	}

	// Create informer with config
	informer := kubernetes.NewDeploymentInformerWithConfig(client.Clientset(), cfg)

	// Set informer in server
	srv.SetDeploymentInformer(informer)

	// Start informer
	logger.Info("Starting deployment informer", map[string]interface{}{
		"namespace":     cfg.Controller.Single.Namespace,
		"resync_period": cfg.Controller.ResyncPeriod,
	})

	return informer.Start()
}
