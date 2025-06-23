package cmd

import (
	"fmt"
	"os"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version = "v0.5.2"
var logLevel string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k6s",
	Short: "Kubernetes controller and deployment management tool",
	Long: `k6s is a CLI tool for managing Kubernetes deployments and provides
a foundation for building custom controllers.

This tool demonstrates modern practices for Kubernetes controller development
using client-go and controller-runtime.`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Get log level from viper (checks env vars, config files, and flags)
		level := viper.GetString("log.level")
		if level == "" {
			level = logLevel // fallback to flag value
		}
		logger.SetLevel(logger.LogLevel(level))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Initialize logger
	logger.Init()
	
	// Add global flags using pflag
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", 
		"Set log level (trace, debug, info, warn, error)")
	
	// Bind flag to viper for environment variable support
	if err := viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		logger.Error("Failed to bind log level flag", err, nil)
	}
	
	// Set environment variable prefix
	viper.SetEnvPrefix("K6S")
	viper.AutomaticEnv()
	
	// Allow environment variables like K6S_LOG_LEVEL
	if err := viper.BindEnv("log.level", "K6S_LOG_LEVEL"); err != nil {
		logger.Error("Failed to bind log level env", err, nil)
	}
	
	// Add flag completion
	if err := rootCmd.RegisterFlagCompletionFunc("log-level", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"trace", "debug", "info", "warn", "error"}, cobra.ShellCompDirectiveDefault
	}); err != nil {
		logger.Error("Failed to register flag completion", err, nil)
	}
}
