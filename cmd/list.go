package cmd

import (
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubeops"
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
	Run: func(cmd *cobra.Command, args []string) {
		if err := kubeops.ListPods(Clientset, listNamespace); err != nil {
			handleError(err, "Failed to list pods")
		}
	},
}

func init() {
	listCmd.PersistentFlags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace")
	listCmd.AddCommand(listPodCmd)
	rootCmd.AddCommand(listCmd)
}
