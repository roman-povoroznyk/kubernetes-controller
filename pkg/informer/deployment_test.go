package informer

import (
	"context"
	"testing"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
)

func TestNewDeploymentInformer(t *testing.T) {
	// This test requires a running Kubernetes cluster
	t.Skip("Skipping deployment informer tests - requires running cluster")
	
	client, err := kubernetes.NewClient("")
	if err != nil {
		t.Skipf("Cannot create Kubernetes client: %v", err)
	}

	tests := []struct {
		name         string
		namespace    string
		resyncPeriod time.Duration
	}{
		{
			name:         "all namespaces",
			namespace:    "",
			resyncPeriod: 30 * time.Second,
		},
		{
			name:         "specific namespace",
			namespace:    "default",
			resyncPeriod: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			informer := NewDeploymentInformer(client, tt.namespace, tt.resyncPeriod)
			
			if informer == nil {
				t.Errorf("NewDeploymentInformer() returned nil")
				return
			}
			
			if informer.namespace != tt.namespace {
				t.Errorf("NewDeploymentInformer() namespace = %v, want %v", informer.namespace, tt.namespace)
			}
			
			if informer.resyncPeriod != tt.resyncPeriod {
				t.Errorf("NewDeploymentInformer() resyncPeriod = %v, want %v", informer.resyncPeriod, tt.resyncPeriod)
			}
		})
	}
}

func TestDeploymentInformer_Methods(t *testing.T) {
	// This test requires a running Kubernetes cluster
	t.Skip("Skipping deployment informer methods tests - requires running cluster")
	
	client, err := kubernetes.NewClient("")
	if err != nil {
		t.Skipf("Cannot create Kubernetes client: %v", err)
	}

	informer := NewDeploymentInformer(client, "default", 30*time.Second)
	
	t.Run("HasSynced before start", func(t *testing.T) {
		if informer.HasSynced() {
			t.Errorf("HasSynced() should return false before starting")
		}
	})

	t.Run("Start and stop", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			err := informer.Start(ctx)
			if err != nil && err != context.DeadlineExceeded {
				errCh <- err
			}
		}()

		// Wait a bit for informer to start
		time.Sleep(2 * time.Second)
		
		// Stop informer
		informer.Stop()
		
		// Check for errors
		select {
		case err := <-errCh:
			t.Errorf("Start() returned error: %v", err)
		default:
			// No error, which is expected
		}
	})
}
