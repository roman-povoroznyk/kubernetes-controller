package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	allNamespaces bool
	kubeconfig    string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes deployments",
	Long:  `List Kubernetes deployments in the specified namespace or all namespaces.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Build kubeconfig path
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}

		// Load kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading kubeconfig: %v\n", err)
			os.Exit(1)
		}

		// Create Kubernetes client
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating kubernetes client: %v\n", err)
			os.Exit(1)
		}

		// Determine namespace
		namespace := "default"
		if allNamespaces {
			namespace = ""
		}

		// List deployments
		deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error listing deployments: %v\n", err)
			os.Exit(1)
		}

		// Print results
		printDeployments(deployments.Items, allNamespaces)
	},
}

// printDeployments prints deployments in kubectl-like format
func printDeployments(deployments []appsv1.Deployment, showNamespace bool) {
	if len(deployments) == 0 {
		fmt.Println("no deployments found")
		return
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	if showNamespace {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	} else {
		fmt.Fprintln(w, "NAME\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")
	}

	// Print each deployment
	for _, deploy := range deployments {
		ready := fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas)
		upToDate := fmt.Sprintf("%d", deploy.Status.UpdatedReplicas)
		available := fmt.Sprintf("%d", deploy.Status.AvailableReplicas)
		age := FormatAge(deploy.CreationTimestamp.Time)

		if showNamespace {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				deploy.Namespace, deploy.Name, ready, upToDate, available, age)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				deploy.Name, ready, upToDate, available, age)
		}
	}
}

// FormatAge formats age like kubectl
func FormatAge(t time.Time) string {
	now := time.Now()
	age := now.Sub(t)

	days := int(age.Hours() / 24)
	hours := int(age.Hours()) % 24
	minutes := int(age.Minutes()) % 60
	seconds := int(age.Seconds()) % 60

	if days >= 7 {
		// 7d+ - show only days
		return fmt.Sprintf("%dd", days)
	} else if days >= 2 {
		// 2-7d - show days and hours
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	} else if days >= 1 {
		// 1-2d - show days and hours
		if hours > 0 {
			return fmt.Sprintf("%dd%dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	} else if age.Hours() >= 1 {
		// 1h+ - show hours and minutes
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	} else if age.Minutes() >= 1 {
		// 1m+ - show minutes and seconds
		if seconds > 0 {
			return fmt.Sprintf("%dm%ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	} else {
		// <1m - show seconds
		return fmt.Sprintf("%ds", seconds)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List deployments across all namespaces")
	listCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
}
