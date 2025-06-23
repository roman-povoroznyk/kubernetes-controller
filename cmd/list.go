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
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes deployments",
	Long:  `List Kubernetes deployments in specified namespace using client-go.`,
	Run: func(cmd *cobra.Command, args []string) {
		listDeployments()
	},
}

func listDeployments() {
	log.Info().
		Str("namespace", namespace).
		Bool("all_namespaces", allNamespaces).
		Msg("Listing deployments")

	// Create Kubernetes client
	clientset, err := createKubernetesClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Determine target namespace
	targetNamespace := namespace
	if allNamespaces {
		targetNamespace = ""
	}

	// List deployments
	deployments, err := clientset.AppsV1().Deployments(targetNamespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list deployments")
	}

	// Output results
	if allNamespaces {
		fmt.Printf("Found %d deployments across all namespaces:\n", len(deployments.Items))
	} else {
		fmt.Printf("Found %d deployments in namespace '%s':\n", len(deployments.Items), namespace)
	}

	for _, deployment := range deployments.Items {
		if allNamespaces {
			fmt.Printf("  %s/%s (Ready: %d/%d)\n",
				deployment.Namespace,
				deployment.Name,
				deployment.Status.ReadyReplicas,
				deployment.Status.Replicas,
			)
		} else {
			fmt.Printf("  %s (Ready: %d/%d)\n",
				deployment.Name,
				deployment.Status.ReadyReplicas,
				deployment.Status.Replicas,
			)
		}
	}

	log.Info().Int("count", len(deployments.Items)).Msg("Deployments listed successfully")
}

func createKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try to use kubeconfig from flag, environment, or default location
	kubeconfigPath := kubeconfig
	if kubeconfigPath == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeconfigPath = env
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Try to load kubeconfig
	if _, err := os.Stat(kubeconfigPath); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
		}
		log.Debug().Str("kubeconfig", kubeconfigPath).Msg("Using kubeconfig file")
	} else {
		// Fall back to in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
		log.Debug().Msg("Using in-cluster config")
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return clientset, nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
