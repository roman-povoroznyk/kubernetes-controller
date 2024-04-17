package cmd

import (
	"fmt"
	"os"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
)

var Version = "v0.1.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k6s",
	Short: "Kubernetes controller and deployment management tool",
	Long: `k6s is a CLI tool for managing Kubernetes deployments and provides
a foundation for building custom controllers.

This tool demonstrates modern practices for Kubernetes controller development
using client-go and controller-runtime.`,
	Version: Version,
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
	
	// Global flags can be added here
}
