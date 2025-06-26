package kubernetes

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/informer"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubernetes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	watchNamespace  string
	watchInCluster  bool
	watchKubeconfig string
	watchResyncTime time.Duration
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch Kubernetes resources for changes",
	Long:  "Watch Kubernetes resources and log events as they happen",
}

var watchDeploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Watch deployment events",
	Long:  "Watch for deployment creation, updates, and deletions",
	RunE: func(c *cobra.Command, args []string) error {
		// Create Kubernetes client with appropriate auth method
		clientset, err := kubernetes.NewKubernetesClient(watchInCluster, watchKubeconfig)
		if err != nil {
			cmd.HandleError(err, "Failed to create Kubernetes client")
			return err
		}

		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			log.Info().Msg("Received shutdown signal")
			cancel()
		}()

		// Configure and start the informer
		config := informer.DeploymentInformerConfig{
			Namespace:    watchNamespace,
			ResyncPeriod: watchResyncTime,
		}

		log.Info().
			Str("namespace", watchNamespace).
			Bool("in_cluster", watchInCluster).
			Str("kubeconfig", watchKubeconfig).
			Msg("Starting deployment watch...")

		return informer.StartDeploymentInformer(ctx, clientset, config)
	},
}

var watchPodCmd = &cobra.Command{
	Use:   "pod",
	Short: "Watch pod events",
	Long:  "Watch for pod creation, updates, and deletions",
	RunE: func(c *cobra.Command, args []string) error {
		// Create Kubernetes client with appropriate auth method
		clientset, err := kubernetes.NewKubernetesClient(watchInCluster, watchKubeconfig)
		if err != nil {
			cmd.HandleError(err, "Failed to create Kubernetes client")
			return err
		}

		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			log.Info().Msg("Received shutdown signal")
			cancel()
		}()

		// Configure and start the informer
		config := informer.PodInformerConfig{
			Namespace:    watchNamespace,
			ResyncPeriod: watchResyncTime,
		}

		log.Info().
			Str("namespace", watchNamespace).
			Bool("in_cluster", watchInCluster).
			Str("kubeconfig", watchKubeconfig).
			Msg("Starting pod watch...")

		return informer.StartPodInformer(ctx, clientset, config)
	},
}

func init() {
	watchCmd.PersistentFlags().StringVarP(&watchNamespace, "namespace", "n", "default", "Namespace to watch")
	watchCmd.PersistentFlags().BoolVar(&watchInCluster, "in-cluster", false, "Use in-cluster authentication")
	watchCmd.PersistentFlags().StringVar(&watchKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	watchCmd.PersistentFlags().DurationVar(&watchResyncTime, "resync-period", 30*time.Second, "Resync period for the informer")

	watchCmd.AddCommand(watchDeploymentCmd)
	watchCmd.AddCommand(watchPodCmd)

	cmd.RootCmd.AddCommand(watchCmd)
}
