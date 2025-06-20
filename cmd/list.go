package cmd

import (
    "kubernetes-controller/internal/kubeops"
    "github.com/spf13/cobra"
    "github.com/rs/zerolog/log"
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
        log.Info().Str("namespace", listNamespace).Msg("Listing pods")

        if err := kubeops.ListPods(Clientset, listNamespace); err != nil {
            handleError(err, "Failed to list pods")
        }

        log.Info().Msg("Pods listed successfully")
    },
}

func init() {
    listCmd.PersistentFlags().StringVarP(&listNamespace, "namespace", "n", "default", "Namespace")
    listCmd.AddCommand(listPodCmd)
    rootCmd.AddCommand(listCmd)
}
