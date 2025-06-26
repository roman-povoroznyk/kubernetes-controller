package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps Kubernetes client with additional functionality
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewClient creates a new Kubernetes client
// It tries in-cluster config first, then falls back to kubeconfig
func NewClient(kubeconfig string) (*Client, error) {
	var config *rest.Config
	var err error

	// If kubeconfig path is provided, use it
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig %s: %w", kubeconfig, err)
		}
	} else {
		// Try in-cluster config first
		config, err = rest.InClusterConfig()
		if err != nil {
			// Fall back to default kubeconfig location
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %w", err)
			}
			
			defaultKubeconfig := filepath.Join(home, ".kube", "config")
			if _, err := os.Stat(defaultKubeconfig); os.IsNotExist(err) {
				return nil, fmt.Errorf("no kubeconfig found and not running in cluster")
			}
			
			config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
			if err != nil {
				return nil, fmt.Errorf("failed to build config from default kubeconfig: %w", err)
			}
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// ListDeployments lists all deployments in the specified namespace
func (c *Client) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	if namespace == "" {
		namespace = "default"
	}

	logger.Debug("Listing deployments", map[string]interface{}{
		"namespace": namespace,
	})

	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace %s: %w", namespace, err)
	}

	logger.Info("Successfully listed deployments", map[string]interface{}{
		"namespace": namespace,
		"count":     len(deployments.Items),
	})

	return deployments, nil
}

// ListDeploymentsAllNamespaces lists all deployments across all namespaces
func (c *Client) ListDeploymentsAllNamespaces(ctx context.Context) (*appsv1.DeploymentList, error) {
	logger.Debug("Listing deployments in all namespaces", nil)

	deployments, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in all namespaces: %w", err)
	}

	logger.Info("Successfully listed deployments from all namespaces", map[string]interface{}{
		"count": len(deployments.Items),
	})

	return deployments, nil
}

// GetDeployment gets a specific deployment by name and namespace
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	logger.Debug("Getting deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})

	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s in namespace %s: %w", name, namespace, err)
	}

	logger.Info("Successfully got deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})

	return deployment, nil
}

// CreateDeployment creates a new deployment in the specified namespace
func (c *Client) CreateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	if namespace == "" {
		namespace = "default"
	}

	logger.Debug("Creating deployment", map[string]interface{}{
		"name":      deployment.Name,
		"namespace": namespace,
		"image":     deployment.Spec.Template.Spec.Containers[0].Image,
		"replicas":  *deployment.Spec.Replicas,
	})

	result, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment %s in namespace %s: %w", deployment.Name, namespace, err)
	}

	logger.Info("Successfully created deployment", map[string]interface{}{
		"name":      result.Name,
		"namespace": result.Namespace,
		"replicas":  *result.Spec.Replicas,
	})

	return result, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (c *Client) GetClientset() *kubernetes.Clientset {
	return c.clientset
}

// DeleteDeployment deletes a deployment in the specified namespace
func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = "default"
	}

	logger.Debug("Deleting deployment", map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	})

	err := c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s in namespace %s: %w", name, namespace, err)
	}

	logger.Info("Successfully deleted deployment", map[string]interface{}{
		"name":      name,
		"namespace": namespace,
	})

	return nil
}
# Kubernetes integration
