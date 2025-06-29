package cmd

import (
	"fmt"
	"os"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/controller"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsPort int
	namespace   string
)

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Run the deployment controller (logs all deployment events)",
	Long: `Run a Kubernetes controller that watches Deployment resources and logs
all events (add, update, delete) with detailed information about changes.

The controller uses controller-runtime and provides structured logging
for deployment lifecycle events including replicas, images, and labels.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Starting controller-runtime manager", map[string]interface{}{
			"namespace":    namespace,
			"metricsPort": metricsPort,
		})
		
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
			Metrics: server.Options{
				BindAddress: fmt.Sprintf(":%d", metricsPort),
			},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create manager: %v\n", err)
			os.Exit(1)
		}
		
		if err := controller.AddDeploymentControllerToManager(mgr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add deployment controller: %v\n", err)
			os.Exit(1)
		}
		
		logger.Info("Starting manager loop", nil)
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			fmt.Fprintf(os.Stderr, "Manager exited with error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(controllerCmd)
	controllerCmd.Flags().IntVar(&metricsPort, "metrics-port", 8080, "Port for metrics endpoint")
	controllerCmd.Flags().StringVar(&namespace, "namespace", "", "Namespace to watch (empty = all namespaces)")
}
