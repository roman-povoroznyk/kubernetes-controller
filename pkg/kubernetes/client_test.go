package kubernetes

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		wantError  bool
	}{
		{
			name:       "non-existent kubeconfig",
			kubeconfig: "/non/existent/path",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.kubeconfig)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("NewClient() expected error but got none")
				}
				if client != nil {
					t.Errorf("NewClient() expected nil client but got %v", client)
				}
			} else {
				if err != nil {
					t.Errorf("NewClient() unexpected error: %v", err)
				}
				if client == nil {
					t.Errorf("NewClient() expected client but got nil")
				}
			}
		})
	}
	
	// Test with empty kubeconfig - this may succeed if kubeconfig exists
	t.Run("empty kubeconfig", func(t *testing.T) {
		client, err := NewClient("")
		// Don't assert specific behavior since it depends on environment
		// Just ensure it doesn't panic
		if err != nil {
			t.Logf("NewClient with empty kubeconfig failed as expected: %v", err)
		}
		if client != nil {
			t.Logf("NewClient with empty kubeconfig succeeded")
		}
	})
}

func TestClient_Methods(t *testing.T) {
	// These tests require a running Kubernetes cluster
	// In a real CI environment, you might use envtest or kind
	t.Skip("Skipping Kubernetes client tests - requires running cluster")
	
	client, err := NewClient("")
	if err != nil {
		t.Skipf("Cannot create Kubernetes client: %v", err)
	}

	ctx := context.Background()

	t.Run("ListDeployments", func(t *testing.T) {
		deployments, err := client.ListDeployments(ctx, "default")
		if err != nil {
			t.Errorf("ListDeployments() error = %v", err)
			return
		}
		if deployments == nil {
			t.Errorf("ListDeployments() returned nil")
		}
	})

	t.Run("ListDeploymentsAllNamespaces", func(t *testing.T) {
		deployments, err := client.ListDeploymentsAllNamespaces(ctx)
		if err != nil {
			t.Errorf("ListDeploymentsAllNamespaces() error = %v", err)
			return
		}
		if deployments == nil {
			t.Errorf("ListDeploymentsAllNamespaces() returned nil")
		}
	})
}
