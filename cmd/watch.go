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

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/roman-povoroznyk/k8s/pkg/informer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch Kubernetes deployments using informers",
	Long:  `Watch Kubernetes deployments for changes using client-go informers.`,
	Run: func(cmd *cobra.Command, args []string) {
		watchDeployments()
	},
}

func watchDeployments() {
	log.Info().
		Str("namespace", namespace).
		Msg("Starting deployment watcher")

	// Create Kubernetes client
	clientset, err := createKubernetesClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Determine target namespace
	if allNamespaces {
		log.Info().Msg("Watching all namespaces")
	} else {
		log.Info().Str("namespace", namespace).Msg("Watching specific namespace")
	}

	// Load informer config
	config, err := informer.LoadInformerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load informer config")
	}

	// Create deployment informer
	deploymentInformer := informer.NewDeploymentInformer(clientset, config)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info().Msg("Received interrupt signal, shutting down...")
		cancel()
	}()

	// Start informer
	log.Info().Msg("Starting deployment informer - press Ctrl+C to stop")
	if err := deploymentInformer.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start deployment informer")
	}
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
