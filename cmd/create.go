package cmd

import (
    "kubernetes-controller/internal/kubeops"
    "github.com/spf13/cobra"
    "github.com/rs/zerolog/log"
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
        podName := args[0]
        log.Info().Str("namespace", createNamespace).Str("name", podName).Msg("Creating pod")

        if err := kubeops.CreatePod(Clientset, createNamespace, podName); err != nil {
            handleError(err, "Failed to create pod")
        }

        log.Info().Str("name", podName).Msg("Pod created successfully")
    },
}

func init() {
    createCmd.PersistentFlags().StringVarP(&createNamespace, "namespace", "n", "default", "Namespace")
    createCmd.AddCommand(createPodCmd)
    rootCmd.AddCommand(createCmd)
}
