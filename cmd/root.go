package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultLogLevel   = "info"
	defaultTimeout    = 10 * time.Second
	defaultTimeFormat = "15:04:05"
)

var (
	Clientset      *kubernetes.Clientset
	logLevel       string
	kubeconfigPath string
)

func init() {
	// Configure human-friendly console output for logs
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: defaultTimeFormat,
		NoColor:    false,
	}
	log.Logger = log.Output(output)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Add pflags
	pflag.StringVarP(&logLevel, "log-level", "l", defaultLogLevel,
		"Log level (trace, debug, info, warn, error)")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "",
		"Path to kubeconfig file (defaults to KUBECONFIG env var or ~/.kube/config)")

	// Bind pflag to cobra
	rootCmd.PersistentFlags().AddFlagSet(pflag.CommandLine)

	// Initialize Viper
	initViper()
}

// initViper initializes Viper for environment variables
func initViper() {
	// Use environment variables with prefix "K8S_CTRL_"
	viper.SetEnvPrefix("K8S_CTRL")

	// Replace dashes with underscores in env vars
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Automatically read from environment variables
	viper.AutomaticEnv()

	// Bind flags to viper
	viper.BindPFlag("log-level", pflag.Lookup("log-level"))
	viper.BindPFlag("kubeconfig", pflag.Lookup("kubeconfig"))

	// Set defaults if needed
	viper.SetDefault("log-level", defaultLogLevel)
}

// parseLogLevel converts string to zerolog.Level
func parseLogLevel(lvl string) zerolog.Level {
	switch strings.ToLower(lvl) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// configureLogger sets up logger with specified level
func configureLogger(level zerolog.Level) {
	zerolog.SetGlobalLevel(level)
	log.Debug().Str("level", level.String()).Msg("Log level set")
}

// handleError logs error and exits
func handleError(err error, action string) {
	if err != nil {
		log.Error().Err(err).Msg(action)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "k8s-ctrl",
	Short: "Kubernetes controller CLI",
	Long:  `k8s-ctrl - lightweight CLI for managing Kubernetes resources`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Get log level from viper (environment or flag)
		logLevelStr := viper.GetString("log-level")
		level := parseLogLevel(logLevelStr)
		configureLogger(level)

		// Enhanced logging with more context
		log.Info().
			Str("command", cmd.Name()).
			Strs("args", args).
			Str("log-level", logLevelStr).
			Msg("Command started")

		if Clientset != nil {
			return nil
		}

		// Get kubeconfig path with priorities:
		// 1. Command line flag / environment variable (via viper)
		// 2. KUBECONFIG env var (legacy support)
		// 3. Default ~/.kube/config
		kubeconfig := viper.GetString("kubeconfig")
		if kubeconfig == "" {
			// Legacy support for KUBECONFIG env var
			kubeconfig = os.Getenv("KUBECONFIG")
			if kubeconfig == "" {
				kubeconfig = clientcmd.RecommendedHomeFile
			}
		}

		log.Debug().Str("kubeconfig", kubeconfig).Msg("Using kubeconfig")

		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Error().Err(err).Str("path", kubeconfig).Msg("Failed to load kubeconfig")
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
