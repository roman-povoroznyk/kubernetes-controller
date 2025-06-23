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

func init() {
	listCmd.PersistentFlags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace")
	listCmd.AddCommand(listPodCmd)

	cmd.RootCmd.AddCommand(listCmd)
}
