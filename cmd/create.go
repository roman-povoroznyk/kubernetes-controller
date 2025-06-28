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
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	deploymentName string
	replicas       int32
	image          string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Kubernetes deployment",
	Long:  `Create a Kubernetes deployment with specified parameters.`,
	Run: func(cmd *cobra.Command, args []string) {
		createDeployment()
	},
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Kubernetes deployment",
	Long:  `Delete a Kubernetes deployment by name.`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteDeployment()
	},
}

func createDeployment() {
	log.Info().
		Str("name", deploymentName).
		Str("namespace", namespace).
		Int32("replicas", replicas).
		Str("image", image).
		Msg("Creating deployment")

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
		return
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        deploymentName,
				"created-by": "k8s-controller",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentName,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create deployment")
		return
	}

	log.Info().
		Str("name", result.GetObjectMeta().GetName()).
		Str("namespace", result.GetObjectMeta().GetNamespace()).
		Str("uid", string(result.GetObjectMeta().GetUID())).
		Msg("Deployment created successfully")
}

func deleteDeployment() {
	log.Info().
		Str("name", deploymentName).
		Str("namespace", namespace).
		Msg("Deleting deployment")

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
		return
	}

	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err = deploymentsClient.Delete(context.TODO(), deploymentName, deleteOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete deployment")
		return
	}

	log.Info().
		Str("name", deploymentName).
		Str("namespace", namespace).
		Msg("Deployment deleted successfully")
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	var kubeconfigPath string

	if kubeconfig != "" {
		kubeconfigPath = kubeconfig
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)

	// Create command flags
	createCmd.Flags().StringVar(&deploymentName, "name", "", "Name of the deployment (required)")
	createCmd.Flags().Int32Var(&replicas, "replicas", 1, "Number of replicas")
	createCmd.Flags().StringVar(&image, "image", "nginx:latest", "Container image")
	_ = createCmd.MarkFlagRequired("name")

	// Delete command flags
	deleteCmd.Flags().StringVar(&deploymentName, "name", "", "Name of the deployment to delete (required)")
	_ = deleteCmd.MarkFlagRequired("name")
}
