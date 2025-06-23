package kubernetes

import (
	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubernetes"
	"github.com/spf13/cobra"
)

var deleteNamespace string

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Kubernetes resources",
}

var deletePodCmd = &cobra.Command{
	Use:   "pod [name]",
	Short: "Delete a pod",
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		podName := args[0]
		if err := kubernetes.DeletePod(cmd.Clientset, deleteNamespace, podName); err != nil {
			cmd.HandleError(err, "Failed to delete pod")
		}
	},
}

func init() {
	deleteCmd.PersistentFlags().StringVarP(&deleteNamespace, "namespace", "n", "default", "Namespace")
	deleteCmd.AddCommand(deletePodCmd)

	cmd.RootCmd.AddCommand(deleteCmd)
}
