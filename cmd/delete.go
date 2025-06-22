package cmd

import (
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubeops"
	"github.com/rs/zerolog/log"
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
	Run: func(cmd *cobra.Command, args []string) {
		podName := args[0]
		log.Info().Str("namespace", deleteNamespace).Str("name", podName).Msg("Deleting pod")

		if err := kubeops.DeletePod(Clientset, deleteNamespace, podName); err != nil {
			handleError(err, "Failed to delete pod")
		}

		log.Info().Str("name", podName).Msg("Pod deleted successfully")
	},
}

func init() {
	deleteCmd.PersistentFlags().StringVarP(&deleteNamespace, "namespace", "n", "default", "Namespace")
	deleteCmd.AddCommand(deletePodCmd)
	rootCmd.AddCommand(deleteCmd)
}
