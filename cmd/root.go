package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version  = "v0.10.0"
	cfgFile  string
	logLevel string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k6s",
	Short: "A production-grade Kubernetes controller for managing deployments",
	Long: `k6s is a powerful Kubernetes controller that provides both single-cluster 
and multi-cluster deployment management capabilities.

Features:
• Single and multi-cluster deployment monitoring
• Real-time deployment change tracking with informers
• Leader election for high availability
• Prometheus metrics and health endpoints
• Structured logging with multiple levels
• HTTP API server with FastHTTP
• Configuration via YAML files or environment variables

Examples:
  # Start single-cluster controller
  k6s controller start

  # Start multi-cluster controller
  k6s controller start --mode multi --config ~/.k6s/clusters.yaml

  # Add a cluster to multi-cluster config
  k6s cluster add production --kubeconfig ~/.kube/prod-config

  # List configured clusters
  k6s cluster list

  # Check cluster connectivity
  k6s cluster check-connectivity

  # Start HTTP server for API access
  k6s server --port 8080`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip logging setup for certain commands that need clean output
		if cmd.Use == "version" || cmd.Use == "completion" {
			return
		}

		// Initialize logger with the specified log level
		logger.Init(logger.Config{
			Level:       logLevel,
			Format:      "text",
			Development: false,
			Caller:      false,
		})
		
		logger.Info("k6s controller starting", map[string]interface{}{
			"version": Version,
			"command": cmd.Use,
		})
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is specified, show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k6s/k6s.yaml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", 
		fmt.Sprintf("log level (%s)", getValidLogLevels()))

	// Bind flags to viper
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))

	// Version flags - using SetVersionTemplate for proper Cobra integration
	rootCmd.SetVersionTemplate("k6s version {{.Version}}\n")

	// Silence automatic help/usage output on errors since we already log them
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name "k6s" (without extension).
		viper.AddConfigPath(home + "/.k6s")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("k6s")
	}

	// Environment variables
	viper.SetEnvPrefix("K6S")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// We'll log this after logger is initialized
	}
}

// getValidLogLevels returns a string listing all valid log levels
func getValidLogLevels() string {
	return "trace, debug, info, warn, error, fatal, panic"
}
