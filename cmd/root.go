package cmd

import (
    "os"
    "fmt"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "github.com/spf13/cobra"
)

var Clientset *kubernetes.Clientset

var rootCmd = &cobra.Command{
    Use:   "controller",
    Short: "Kubernetes resource operations CLI",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        if Clientset != nil {
            return nil
        }
        kubeconfig := os.Getenv("KUBECONFIG")
        if kubeconfig == "" {
            kubeconfig = clientcmd.RecommendedHomeFile
        }
        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            return fmt.Errorf("failed to load kubeconfig: %w", err)
        }
        Clientset, err = kubernetes.NewForConfig(config)
        if err != nil {
            return fmt.Errorf("failed to create clientset: %w", err)
        }
        return nil
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
