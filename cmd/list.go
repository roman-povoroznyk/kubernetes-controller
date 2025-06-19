package cmd

import (
    "fmt"
    "os"
    "kubernetes-controller/internal/kubeops"
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
            fmt.Println("Error listing pods:", err)
            os.Exit(1)
        }
    },
}

func init() {
    listCmd.PersistentFlags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace")
    listCmd.AddCommand(listPodCmd)
    rootCmd.AddCommand(listCmd)
}
