package cmd

import (
    "fmt"
    "os"
    "kubernetes-controller/internal/kubeops"
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
        if err := kubeops.DeletePod(Clientset, deleteNamespace, args[0]); err != nil {
            fmt.Println("Error deleting pod:", err)
            os.Exit(1)
        }
        fmt.Println("Pod deleted:", args[0])
    },
}

func init() {
    deleteCmd.PersistentFlags().StringVarP(&deleteNamespace, "namespace", "n", "default", "Namespace")
    deleteCmd.AddCommand(deletePodCmd)
    rootCmd.AddCommand(deleteCmd)
}
