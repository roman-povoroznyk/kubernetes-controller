package informer

import (
	"context"
	"time"

	"k8s.io/client-go/kubernetes"
)

// InformerConfig holds common configuration for all informers
type InformerConfig struct {
	Namespace    string
	ResyncPeriod time.Duration
}

// ResourceInformer defines the interface for resource informers
type ResourceInformer interface {
	Start(ctx context.Context, clientset *kubernetes.Clientset, config InformerConfig) error
	Stop()
}

// InformerManager manages multiple informers
type InformerManager struct {
	informers []ResourceInformer
}

// NewInformerManager creates a new informer manager
func NewInformerManager() *InformerManager {
	return &InformerManager{
		informers: make([]ResourceInformer, 0),
	}
}

// AddInformer adds an informer to the manager
func (im *InformerManager) AddInformer(informer ResourceInformer) {
	im.informers = append(im.informers, informer)
}

// StartAll starts all registered informers
func (im *InformerManager) StartAll(ctx context.Context, clientset *kubernetes.Clientset, config InformerConfig) error {
	for _, informer := range im.informers {
		go func(inf ResourceInformer) {
			if err := inf.Start(ctx, clientset, config); err != nil {
				// Log error but don't stop other informers
				return
			}
		}(informer)
	}
	return nil
}

// StopAll stops all registered informers
func (im *InformerManager) StopAll() {
	for _, informer := range im.informers {
		informer.Stop()
	}
}
