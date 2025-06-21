package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/informer"
	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	watchNamespace   string
	resyncPeriod     time.Duration
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch Kubernetes resources",
	Long: `Watch Kubernetes resources and log events in real-time.
	
Examples:
  k6s watch deployments                    # Watch deployments in all namespaces
  k6s watch deployments -n kube-system    # Watch deployments in kube-system namespace
  k6s watch deployments --resync 30s      # Watch with custom resync period`,
}

// watchDeploymentsCmd represents the watch deployments command
var watchDeploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Watch deployment events",
	Long: `Watch Kubernetes deployment events and log them in real-time.
	
The informer will watch for ADD, UPDATE, and DELETE events on deployments
and log them with structured logging. This is useful for monitoring deployment
changes and debugging deployment issues.

Examples:
  k6s watch deployments                    # Watch deployments in all namespaces
  k6s watch deployments -n kube-system    # Watch deployments in kube-system namespace  
  k6s watch deployments --resync 30s      # Watch with 30 second resync period`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		logger.Init()
		
		// Create Kubernetes client
		client, err := kubernetes.NewClient(kubeconfig)
		if err != nil {
			logger.Error("Failed to create Kubernetes client", err, nil)
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		// Create deployment informer
		deploymentInformer := informer.NewDeploymentInformer(client, watchNamespace, resyncPeriod)

		// Setup context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Setup signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		
		// Start informer in goroutine
		errCh := make(chan error, 1)
		go func() {
			err := deploymentInformer.Start(ctx)
			if err != nil && err != context.Canceled {
				errCh <- err
			}
		}()

		logger.Info("Deployment informer started, watching for events...", map[string]interface{}{
			"namespace":      watchNamespace,
			"resync_period":  resyncPeriod.String(),
		})

		// Wait for signal or error
		select {
		case sig := <-sigCh:
			logger.Info("Received signal, shutting down", map[string]interface{}{
				"signal": sig.String(),
			})
			cancel()
		case err := <-errCh:
			logger.Error("Informer error", err, nil)
			return err
		}

		// Stop informer
		deploymentInformer.Stop()
		
		logger.Info("Deployment informer stopped", nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.AddCommand(watchDeploymentsCmd)

	// Flags for watch command
	watchCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional)")
	watchCmd.PersistentFlags().StringVarP(&watchNamespace, "namespace", "n", "", "Kubernetes namespace to watch (empty = all namespaces)")
	watchCmd.PersistentFlags().DurationVar(&resyncPeriod, "resync", 30*time.Second, "Resync period for the informer")
}
