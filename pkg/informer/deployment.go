// Package informer provides Kubernetes resource informers for watching
// and caching Kubernetes resources using client-go.
package informer

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// DeploymentInformer wraps a deployment informer with additional functionality
type DeploymentInformer struct {
	informer cache.SharedIndexInformer
	stopCh   chan struct{}
}

// NewDeploymentInformer creates a new deployment informer for the specified namespace
func NewDeploymentInformer(clientset *kubernetes.Clientset, namespace string) *DeploymentInformer {
	// Create factory with options
	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		30*time.Second, // Resync period
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fields.Everything().String()
		}),
	)

	// Get deployment informer
	informer := factory.Apps().V1().Deployments().Informer()

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			log.Info().
				Str("name", deployment.Name).
				Str("namespace", deployment.Namespace).
				Int32("replicas", *deployment.Spec.Replicas).
				Msg("Deployment added")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeployment := oldObj.(*appsv1.Deployment)
			newDeployment := newObj.(*appsv1.Deployment)
			
			log.Info().
				Str("name", newDeployment.Name).
				Str("namespace", newDeployment.Namespace).
				Int32("old_replicas", *oldDeployment.Spec.Replicas).
				Int32("new_replicas", *newDeployment.Spec.Replicas).
				Msg("Deployment updated")
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			log.Info().
				Str("name", deployment.Name).
				Str("namespace", deployment.Namespace).
				Msg("Deployment deleted")
		},
	})

	return &DeploymentInformer{
		informer: informer,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the informer and waits for cache sync
func (di *DeploymentInformer) Start(ctx context.Context) error {
	log.Info().Msg("Starting deployment informer")

	// Start the informer
	go di.informer.Run(di.stopCh)

	// Wait for cache to sync
	if !cache.WaitForCacheSync(ctx.Done(), di.informer.HasSynced) {
		close(di.stopCh)
		return fmt.Errorf("failed to wait for cache sync")
	}

	log.Info().Msg("Deployment informer cache synced successfully")

	// Wait for context cancellation
	<-ctx.Done()
	close(di.stopCh)
	log.Info().Msg("Deployment informer stopped")

	return nil
}

// GetDeployments returns all deployments from the informer cache
func (di *DeploymentInformer) GetDeployments() []*appsv1.Deployment {
	var deployments []*appsv1.Deployment
	
	for _, obj := range di.informer.GetStore().List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			deployments = append(deployments, deployment)
		}
	}
	
	return deployments
}

// GetDeploymentNames returns names of all deployments from the informer cache
func (di *DeploymentInformer) GetDeploymentNames() []string {
	var names []string
	
	for _, obj := range di.informer.GetStore().List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			names = append(names, deployment.Name)
		}
	}
	
	return names
}
