package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/informer"
	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/roman-povoroznyk/k6s/pkg/server"
	"github.com/spf13/cobra"
)

var (
	servePort        int
	serveNamespace   string
	serveResyncPeriod time.Duration
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server with deployment informer",
	Long: `Start HTTP server with deployment informer for real-time deployment monitoring and API access.
	
This command combines the HTTP server functionality with deployment informer,
providing both RESTful API endpoints and real-time event monitoring.

Examples:
  k6s serve --port 8080                    # Serve on port 8080, watch all namespaces
  k6s serve --port 8080 -n default         # Serve and watch only default namespace
  k6s serve --port 8080 --resync 60s       # Serve with custom resync period`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		logger.Init()
		
		// Create Kubernetes client
		client, err := kubernetes.NewClient(kubeconfig)
		if err != nil {
			logger.Error("Failed to create Kubernetes client", err, nil)
			return err
		}

		// Create deployment informer
		deploymentInformer := informer.NewDeploymentInformer(client, serveNamespace, serveResyncPeriod)

		// Create server and set informer
		srv := server.New(servePort)
		srv.SetDeploymentInformer(deploymentInformer)

		// Setup context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		
		// Start informer in goroutine
		go func() {
			logger.Info("Starting deployment informer for API cache", map[string]interface{}{
				"namespace":      serveNamespace,
				"resync_period":  serveResyncPeriod.String(),
			})
			
			err := deploymentInformer.Start(ctx)
			if err != nil && err != context.Canceled {
				logger.Error("Informer error", err, nil)
			}
		}()

		// Start server in goroutine
		go func() {
			logger.Info("Starting HTTP server with API endpoints", map[string]interface{}{
				"port": servePort,
			})
			
			err := srv.Start()
			if err != nil {
				logger.Error("Server error", err, nil)
			}
		}()

		// Wait for signal
		sig := <-sigCh
		logger.Info("Received signal, shutting down", map[string]interface{}{
			"signal": sig.String(),
		})

		// Stop informer
		cancel()
		deploymentInformer.Stop()
		
		logger.Info("Server and informer stopped", nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Flags for serve command
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional)")
	serveCmd.Flags().StringVarP(&serveNamespace, "namespace", "n", "", "Kubernetes namespace to watch (empty = all namespaces)")
	serveCmd.Flags().DurationVar(&serveResyncPeriod, "resync", 30*time.Second, "Resync period for the informer")
}
