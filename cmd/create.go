package cmd

import (
    "fmt"
    "os"
    "kubernetes-controller/internal/kubeops"
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
    Run: func(cmd *cobra.Command, args []string) {
        if err := kubeops.CreatePod(Clientset, createNamespace, args[0]); err != nil {
            fmt.Println("Error creating pod:", err)
            os.Exit(1)
        }
        fmt.Println("Pod created:", args[0])
    },
}

func init() {
    createCmd.PersistentFlags().StringVarP(&createNamespace, "namespace", "n", "default", "Namespace")
    createCmd.AddCommand(createPodCmd)
    rootCmd.AddCommand(createCmd)
}
