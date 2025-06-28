/*
Copyright © 2025 Roman Povoroznyk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package cmd implements the command-line interface for the k8s application.
// It uses the Cobra library to provide a structured CLI with subcommands and flags.
package cmd

import (
	"os"
	

	"github.com/roman-povoroznyk/k8s/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	// Version holds the current version of the application.
	// This value can be overridden at build time using ldflags:
	// go build -ldflags "-X github.com/roman-povoroznyk/k8s/cmd.Version=v1.0.0"
	Version = "dev"

	// Global flags using pflag (POSIX/GNU style)
	logLevel      string
	kubeconfig    string
	namespace     string
	allNamespaces bool
)

// rootCmd represents the base command when called without any subcommands.
// It serves as the entry point for the CLI application.
var rootCmd = &cobra.Command{
	Use:   "k8s",
	Short: "A Kubernetes controller written in Go",
	Long:  `k8s is a Kubernetes controller built with Cobra CLI library.`,
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		// Initialize logger for all commands except version
		if cmd.Use != "version" {
			logger.Init(logLevel)
			log.Info().Str("version", Version).Msg("Starting k8s")
		}
	},
	Run: func(cmd *cobra.Command, _ []string) {
		// If no subcommand is specified, show help
		_ = cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags using pflag (POSIX/GNU style)
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Set the logging level (trace, debug, info, warn, error, fatal, panic)")
	
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "",
		"Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default",
		"Kubernetes namespace to operate in")
	
	rootCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false,
		"If present, list resources across all namespaces")

	// Configure version output
	rootCmd.SetVersionTemplate("k8s version {{.Version}}\n")
	
	// Improve error handling
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}
