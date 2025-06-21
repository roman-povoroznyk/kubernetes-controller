package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	deploymentName string
	deploymentImage string
	deploymentReplicas int32
	deploymentPort int32
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Kubernetes resources",
	Long: `Create Kubernetes resources like deployments.
	
Examples:
  k6s create deployment nginx --image=nginx:latest --replicas=3 --port=80
  k6s create deployment web --image=httpd:latest -n my-namespace`,
}

// createDeploymentCmd represents the create deployment command
var createDeploymentCmd = &cobra.Command{
	Use:   "deployment [name]",
	Short: "Create a deployment",
	Long: `Create a Kubernetes deployment with the specified configuration.
	
Examples:
  k6s create deployment nginx --image=nginx:latest --replicas=3 --port=80
  k6s create deployment web --image=httpd:latest --replicas=2 -n my-namespace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deploymentName = args[0]
		
		// Initialize logger
		logger.Init()
		
		// Create Kubernetes client
		client, err := kubernetes.NewClient(kubeconfig)
		if err != nil {
			logger.Error("Failed to create Kubernetes client", err, nil)
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if namespace == "" {
			namespace = "default"
		}

		// Create deployment object
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: namespace,
				Labels: map[string]string{
					"app": deploymentName,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &deploymentReplicas,
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
								Image: deploymentImage,
							},
						},
					},
				},
			},
		}

		// Add port configuration if specified
		if deploymentPort > 0 {
			deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: deploymentPort,
					Protocol:      corev1.ProtocolTCP,
				},
			}
			
			// Add readiness probe
			deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/",
						Port: intstr.FromInt32(deploymentPort),
					},
				},
				InitialDelaySeconds: 5,
				PeriodSeconds:       10,
			}
		}

		// Create deployment
		result, err := client.CreateDeployment(ctx, namespace, deployment)
		if err != nil {
			logger.Error("Failed to create deployment", err, map[string]interface{}{
				"name":      deploymentName,
				"namespace": namespace,
			})
			return err
		}

		logger.Info("Successfully created deployment", map[string]interface{}{
			"name":      result.Name,
			"namespace": result.Namespace,
			"replicas":  *result.Spec.Replicas,
		})

		fmt.Printf("Deployment %s created successfully in namespace %s\n", result.Name, result.Namespace)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createDeploymentCmd)

	// Flags for create deployment command
	createDeploymentCmd.Flags().StringVar(&deploymentImage, "image", "", "Container image (required)")
	createDeploymentCmd.Flags().Int32Var(&deploymentReplicas, "replicas", 1, "Number of replicas")
	createDeploymentCmd.Flags().Int32Var(&deploymentPort, "port", 0, "Container port (optional)")
	createDeploymentCmd.MarkFlagRequired("image")

	// Inherit global flags
	createCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional)")
	createCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (default: default)")
}
