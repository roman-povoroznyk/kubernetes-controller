package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var Clientset *kubernetes.Clientset

var RootCmd = &cobra.Command{
	Use:   "k8s-ctrl",
	Short: "A lightweight CLI for managing Kubernetes resources",
	Long:  `k8s-ctrl is a command-line tool to interact with your Kubernetes cluster, allowing you to manage resources like Pods.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := setupLogging(); err != nil {
			return err
		}

		if Clientset != nil {
			return nil
		}

		log.Debug().Msg("Initializing Kubernetes client")
		kubeconfig := viper.GetString("kubeconfig")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}

		log.Debug().Str("path", kubeconfig).Msg("Using kubeconfig file")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to build config from flags: %w", err)
		}

		Clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes clientset: %w", err)
		}
		log.Info().Msg("Kubernetes client initialized successfully")
		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("CLI execution failed")
		os.Exit(1)
	}
}

func HandleError(err error, message string) {
	log.Error().Err(err).Msg(message)
	os.Exit(1)
}

func init() {
	cobra.OnInitialize(initViper)

	RootCmd.PersistentFlags().StringP("kubeconfig", "k", "", "path to the kubeconfig file")
	RootCmd.PersistentFlags().StringP("log-level", "l", "info", "log level (debug, info, warn, error)")

	viper.BindPFlag("kubeconfig", RootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("log-level", RootCmd.PersistentFlags().Lookup("log-level"))
}

func initViper() {
	viper.SetEnvPrefix("K8S_CTRL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func setupLogging() error {
	logLevelStr := viper.GetString("log-level")
	level, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", logLevelStr, err)
	}

	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	})
	return nil
}
