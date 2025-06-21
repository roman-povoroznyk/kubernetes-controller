package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
)

var (
	kubeconfig    string
	namespace     string
	allNamespaces bool
	outputFormat  string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes resources",
	Long: `List Kubernetes resources like deployments.
	
Examples:
  k6s list deployments                    # List deployments in default namespace
  k6s list deployments -n kube-system    # List deployments in kube-system namespace
  k6s list deployments --all-namespaces  # List deployments in all namespaces
  k6s list deployments -o json           # Output in JSON format`,
}

// deploymentsCmd represents the deployments subcommand
var deploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "List deployments",
	Long: `List Kubernetes deployments in the specified namespace.
	
By default, lists deployments in the 'default' namespace.
Use -n/--namespace to specify a different namespace.
Use --all-namespaces to list deployments across all namespaces.`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if allNamespaces {
			// List deployments in all namespaces
			deployments, err := client.ListDeploymentsAllNamespaces(ctx)
			if err != nil {
				logger.Error("Failed to list deployments", err, nil)
				return err
			}

			if outputFormat == "json" {
				return printJSON(deployments)
			}
			return printDeploymentsTable(deployments.Items, true)
		} else {
			// List deployments in specific namespace
			if namespace == "" {
				namespace = "default"
			}
			
			deployments, err := client.ListDeployments(ctx, namespace)
			if err != nil {
				logger.Error("Failed to list deployments", err, map[string]interface{}{
					"namespace": namespace,
				})
				return err
			}

			if outputFormat == "json" {
				return printJSON(deployments)
			}
			return printDeploymentsTable(deployments.Items, false)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(deploymentsCmd)

	// Flags for list command
	listCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional)")
	listCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (default: default)")
	listCmd.PersistentFlags().BoolVar(&allNamespaces, "all-namespaces", false, "List resources across all namespaces")
	listCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")
}

// printDeploymentsTable prints deployments in a table format
func printDeploymentsTable(deployments []appsv1.Deployment, showNamespace bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// Print header
	if showNamespace {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	} else {
		fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	}

	// Print deployments
	for _, deployment := range deployments {
		ready := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, deployment.Status.Replicas)
		upToDate := fmt.Sprintf("%d", deployment.Status.UpdatedReplicas)
		available := fmt.Sprintf("%d", deployment.Status.AvailableReplicas)
		age := formatAge(deployment.CreationTimestamp.Time)

		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				deployment.Namespace,
				deployment.Name,
				ready,
				upToDate,
				available,
				age,
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				deployment.Name,
				ready,
				upToDate,
				available,
				age,
			)
		}
	}

	return nil
}

// printJSON prints the object in JSON format
func printJSON(obj interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(obj)
}

// formatAge formats the age of a resource
func formatAge(creationTime time.Time) string {
	age := time.Since(creationTime)
	
	if age < time.Minute {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	} else if age < time.Hour {
		return fmt.Sprintf("%dm", int(age.Minutes()))
	} else if age < 24*time.Hour {
		return fmt.Sprintf("%dh", int(age.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(age.Hours()/24))
	}
}
