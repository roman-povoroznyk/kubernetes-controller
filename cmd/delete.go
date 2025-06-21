package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Kubernetes resources",
	Long: `Delete Kubernetes resources like deployments.
	
Examples:
  k6s delete deployment nginx
  k6s delete deployment web -n my-namespace`,
}

// deleteDeploymentCmd represents the delete deployment command
var deleteDeploymentCmd = &cobra.Command{
	Use:   "deployment [name]",
	Short: "Delete a deployment",
	Long: `Delete a Kubernetes deployment by name.
	
Examples:
  k6s delete deployment nginx
  k6s delete deployment web -n my-namespace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentName := args[0]
		
		// Initialize logger
		logger.Init()
		
		// Create Kubernetes client
		client, err := kubernetes.NewClient(kubeconfig)
		if err != nil {
			logger.Error("Failed to create Kubernetes client", err, nil)
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if namespace == "" {
			namespace = "default"
		}

		// Delete deployment
		err = client.DeleteDeployment(ctx, namespace, deploymentName)
		if err != nil {
			logger.Error("Failed to delete deployment", err, map[string]interface{}{
				"name":      deploymentName,
				"namespace": namespace,
			})
			return err
		}

		fmt.Printf("Deployment %s deleted successfully from namespace %s\n", deploymentName, namespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteDeploymentCmd)

	// Inherit global flags
	deleteCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional)")
	deleteCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (default: default)")
}
