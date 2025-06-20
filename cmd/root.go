package cmd

import (
	"fmt"
	"os"

	"github.com/roman-povoroznyk/k6s/pkg/config"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/roman-povoroznyk/k6s/pkg/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k6s",
	Short: "k6s - Kubernetes Controller CLI",
	Long: `k6s is a CLI tool and Kubernetes controller for managing deployments.
	
It provides commands for listing, creating, and managing Kubernetes resources,
as well as running a controller with informers and webhooks.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			buildInfo := version.GetBuildInfo()
			fmt.Printf("k6s version: %s\n", buildInfo["version"])
			if buildInfo["git_commit"] != "unknown" {
				fmt.Printf("Git commit: %s\n", buildInfo["git_commit"])
			}
			if buildInfo["build_date"] != "unknown" {
				fmt.Printf("Build date: %s\n", buildInfo["build_date"])
			}
			return
		}
		
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}
		
		// Initialize logger
		logger.Init()
		logger.Info("Welcome to k6s - Kubernetes Controller CLI!", map[string]interface{}{
			"version": version.GetVersion(),
			"env":     cfg.App.Env,
		})
		fmt.Println("Welcome to k6s - Kubernetes Controller CLI!")
		fmt.Println("Use --help for available commands.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Version flag (local to root command)
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
	
	// Global flags for configuration
	rootCmd.PersistentFlags().String("config", "", "config file (default is ./configs/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Set log level (trace, debug, info, warn, error)")
	rootCmd.PersistentFlags().String("env", "development", "Environment (development, production)")
}
