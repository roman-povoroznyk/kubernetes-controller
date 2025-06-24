package kubernetes

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps kubernetes client with helper methods
type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client from kubeconfig
func NewClient(kubeconfig string) (*Client, error) {
	// Build kubeconfig path
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %w", err)
	}

	return &Client{clientset: clientset}, nil
}
