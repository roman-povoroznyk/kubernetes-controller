package cmd

import (
    "os"
    "fmt"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "github.com/spf13/cobra"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

var Clientset *kubernetes.Clientset

func init() {
    output := zerolog.ConsoleWriter{
        Out:        os.Stdout,
        TimeFormat: "15:04:05",
        NoColor:    false,
    }
    log.Logger = log.Output(output)
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func handleError(err error, action string) {
    if err != nil {
        log.Error().Err(err).Msg(action)
        os.Exit(1)
    }
}

var rootCmd = &cobra.Command{
    Use:   "controller",
    Short: "Kubernetes resource operations CLI",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        log.Info().Str("command", cmd.Name()).Strs("args", args).Msg("Command started")
        if Clientset != nil {
            return nil
        }

        kubeconfig := os.Getenv("KUBECONFIG")
        if kubeconfig == "" {
            kubeconfig = clientcmd.RecommendedHomeFile
        }

        config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            log.Error().Err(err).Msg("Failed to load kubeconfig")
            return fmt.Errorf("failed to load kubeconfig: %w", err)
        }

        Clientset, err = kubernetes.NewForConfig(config)
        if err != nil {
            log.Error().Err(err).Msg("Failed to create clientset")
            return fmt.Errorf("failed to create clientset: %w", err)
        }

        log.Info().Msg("Kubernetes client initialized successfully")
        return nil
    },
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        log.Info().Str("command", cmd.Name()).Msg("Command finished")
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Error().Err(err).Msg("CLI execution failed")
        os.Exit(1)
    }
}
