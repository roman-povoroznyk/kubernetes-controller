package kubernetes

import (
	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubernetes"
	"github.com/spf13/cobra"
)

var listNamespace string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes resources",
}

var listPodCmd = &cobra.Command{
	Use:   "pod",
	Short: "List pods",
	Run: func(c *cobra.Command, args []string) {
		if err := kubernetes.ListPods(cmd.Clientset, listNamespace); err != nil {
			cmd.HandleError(err, "Failed to list pods")
		}
	},
}

var listDeploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "List deployments",
	Run: func(c *cobra.Command, args []string) {
		if err := kubernetes.ListDeployments(cmd.Clientset, listNamespace); err != nil {
			cmd.HandleError(err, "Failed to list deployments")
		}
	},
}

func init() {
	listCmd.PersistentFlags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace")
	listCmd.AddCommand(listPodCmd)
	listCmd.AddCommand(listDeploymentCmd)

	cmd.RootCmd.AddCommand(listCmd)
}
