package kubernetes

import (
	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubernetes"
	"github.com/spf13/cobra"
)

var createNamespace string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Kubernetes resources",
}

var createPodCmd = &cobra.Command{
	Use:   "pod [name]",
	Short: "Create a pod",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		podName := args[0]
		if err := kubernetes.CreatePod(cmd.Clientset, createNamespace, podName); err != nil {
			cmd.HandleError(err, "Failed to create pod")
		}
	},
}

var createDeploymentCmd = &cobra.Command{
	Use:   "deployment [name]",
	Short: "Create a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		name := args[0]
		if err := kubernetes.CreateDeployment(cmd.Clientset, createNamespace, name); err != nil {
			cmd.HandleError(err, "Failed to create deployment")
		}
	},
}

func init() {
	createCmd.PersistentFlags().StringVarP(&createNamespace, "namespace", "n", "default", "Namespace")
	createCmd.AddCommand(createPodCmd)
	createCmd.AddCommand(createDeploymentCmd)

	cmd.RootCmd.AddCommand(createCmd)
}
