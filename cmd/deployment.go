package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/kubernetes"
	"github.com/spf13/cobra"
)

var (
	deployAllNamespaces   bool
	deployKubeconfig      string
	deployCreateImage     string
	deployCreateReplicas  int32
	deployCreateNamespace string
	deployDeleteNamespace string
	deployWatch           bool
	deployWatchResync     time.Duration
	deployNamespace       string
)

// deploymentCmd represents the deployment command group
var deploymentCmd = &cobra.Command{
	Use:     "deployment",
	Aliases: []string{"deploy", "deployments"},
	Short:   "Manage Kubernetes deployments",
	Long:    `Manage Kubernetes deployments with list, create, and delete operations.`,
}

// deploymentListCmd represents the deployment list command
var deploymentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes deployments",
	Long:  `List Kubernetes deployments in the specified namespace or all namespaces. Use --watch to monitor for changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := kubernetes.NewClient(deployKubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating kubernetes client: %v\n", err)
			os.Exit(1)
		}

		// Determine namespace
		namespace := deployNamespace
		if deployAllNamespaces {
			namespace = ""
		}

		if deployWatch {
			// Watch mode using informer
			informer := kubernetes.NewDeploymentInformer(client.Clientset(), namespace, deployWatchResync)

			err = informer.Start()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error starting informer: %v\n", err)
				os.Exit(1)
			}

			// Set up signal handling
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			// Wait for interrupt signal
			<-ctx.Done()

			informer.Stop()
		} else {
			// Regular list mode
			deployments, err := client.DeploymentList(namespace)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error listing deployments: %v\n", err)
				os.Exit(1)
			}

			kubernetes.DeploymentPrint(deployments.Items, deployAllNamespaces)
		}
	},
}

// deploymentCreateCmd represents the deployment create command
var deploymentCreateCmd = &cobra.Command{
	Use:   "create [NAME]",
	Short: "Create a new deployment",
	Long:  `Create a new Kubernetes deployment with specified image and replica count.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if deployCreateImage == "" {
			fmt.Fprintf(os.Stderr, "error: --image flag is required\n")
			os.Exit(1)
		}

		client, err := kubernetes.NewClient(deployKubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating kubernetes client: %v\n", err)
			os.Exit(1)
		}

		if deployCreateNamespace == "" {
			deployCreateNamespace = "default"
		}

		err = client.DeploymentCreate(deployCreateNamespace, name, deployCreateImage, deployCreateReplicas)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating deployment: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("deployment.apps/%s created\n", name)
	},
}

// deploymentDeleteCmd represents the deployment delete command
var deploymentDeleteCmd = &cobra.Command{
	Use:   "delete [NAME]",
	Short: "Delete a deployment",
	Long:  `Delete a Kubernetes deployment by name.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		client, err := kubernetes.NewClient(deployKubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating kubernetes client: %v\n", err)
			os.Exit(1)
		}

		if deployDeleteNamespace == "" {
			deployDeleteNamespace = "default"
		}

		err = client.DeploymentDelete(deployDeleteNamespace, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error deleting deployment: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("deployment.apps \"%s\" deleted\n", name)
	},
}

func init() {
	rootCmd.AddCommand(deploymentCmd)

	// Add subcommands
	deploymentCmd.AddCommand(deploymentListCmd)
	deploymentCmd.AddCommand(deploymentCreateCmd)
	deploymentCmd.AddCommand(deploymentDeleteCmd)

	// List command flags
	deploymentListCmd.Flags().BoolVarP(&deployAllNamespaces, "all-namespaces", "A", false, "List deployments across all namespaces")
	deploymentListCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "default", "Kubernetes namespace")
	deploymentListCmd.Flags().BoolVarP(&deployWatch, "watch", "w", false, "Watch for changes")
	deploymentListCmd.Flags().DurationVar(&deployWatchResync, "resync-period", 30*time.Second, "Resync period for the informer (only used with --watch)")
	deploymentListCmd.Flags().StringVar(&deployKubeconfig, "kubeconfig", "", "Path to kubeconfig file")

	// Create command flags
	deploymentCreateCmd.Flags().StringVar(&deployCreateImage, "image", "", "Container image (required)")
	deploymentCreateCmd.Flags().Int32Var(&deployCreateReplicas, "replicas", 1, "Number of replicas")
	deploymentCreateCmd.Flags().StringVarP(&deployCreateNamespace, "namespace", "n", "default", "Kubernetes namespace")
	deploymentCreateCmd.Flags().StringVar(&deployKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	if err := deploymentCreateCmd.MarkFlagRequired("image"); err != nil {
		panic(fmt.Sprintf("Failed to mark image flag as required: %v", err))
	}

	// Delete command flags
	deploymentDeleteCmd.Flags().StringVarP(&deployDeleteNamespace, "namespace", "n", "default", "Kubernetes namespace")
	deploymentDeleteCmd.Flags().StringVar(&deployKubeconfig, "kubeconfig", "", "Path to kubeconfig file")
}
