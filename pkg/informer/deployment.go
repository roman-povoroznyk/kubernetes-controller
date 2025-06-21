package informer

import (
	"context"
	"fmt"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// DeploymentInformer wraps Kubernetes deployment informer
type DeploymentInformer struct {
	client        *kubernetes.Client
	factory       informers.SharedInformerFactory
	informer      cache.SharedIndexInformer
	stopCh        chan struct{}
	namespace     string
	resyncPeriod  time.Duration
}

// NewDeploymentInformer creates a new deployment informer
func NewDeploymentInformer(client *kubernetes.Client, namespace string, resyncPeriod time.Duration) *DeploymentInformer {
	// Create informer factory
	var factory informers.SharedInformerFactory
	if namespace == "" {
		// Watch all namespaces
		factory = informers.NewSharedInformerFactory(client.GetClientset(), resyncPeriod)
	} else {
		// Watch specific namespace
		factory = informers.NewSharedInformerFactoryWithOptions(
			client.GetClientset(),
			resyncPeriod,
			informers.WithNamespace(namespace),
		)
	}

	// Get deployment informer
	deploymentInformer := factory.Apps().V1().Deployments().Informer()

	return &DeploymentInformer{
		client:       client,
		factory:      factory,
		informer:     deploymentInformer,
		stopCh:       make(chan struct{}),
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
	}
}

// Start starts the informer and begins watching for deployment events
func (di *DeploymentInformer) Start(ctx context.Context) error {
	// Add event handlers
	_, err := di.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			logger.Info("Deployment added", map[string]interface{}{
				"name":      deployment.Name,
				"namespace": deployment.Namespace,
				"replicas":  *deployment.Spec.Replicas,
				"image":     deployment.Spec.Template.Spec.Containers[0].Image,
				"event":     "ADD",
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeployment := oldObj.(*appsv1.Deployment)
			newDeployment := newObj.(*appsv1.Deployment)
			
			// Only log if something meaningful changed
			if oldDeployment.ResourceVersion != newDeployment.ResourceVersion {
				logger.Info("Deployment updated", map[string]interface{}{
					"name":            newDeployment.Name,
					"namespace":       newDeployment.Namespace,
					"replicas":        *newDeployment.Spec.Replicas,
					"ready_replicas":  newDeployment.Status.ReadyReplicas,
					"old_resource_version": oldDeployment.ResourceVersion,
					"new_resource_version": newDeployment.ResourceVersion,
					"event":           "UPDATE",
				})
			}
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			logger.Info("Deployment deleted", map[string]interface{}{
				"name":      deployment.Name,
				"namespace": deployment.Namespace,
				"event":     "DELETE",
			})
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add event handler: %w", err)
	}

	logger.Info("Starting deployment informer", map[string]interface{}{
		"namespace":      di.namespace,
		"resync_period":  di.resyncPeriod.String(),
	})

	// Start the informer
	di.factory.Start(di.stopCh)

	// Wait for cache to sync
	if !cache.WaitForCacheSync(ctx.Done(), di.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	logger.Info("Deployment informer cache synced successfully", map[string]interface{}{
		"namespace": di.namespace,
	})

	// Keep the informer running
	<-ctx.Done()
	return ctx.Err()
}

// Stop stops the informer
func (di *DeploymentInformer) Stop() {
	logger.Info("Stopping deployment informer", map[string]interface{}{
		"namespace": di.namespace,
	})
	close(di.stopCh)
}

// GetDeployment gets a deployment from the informer's cache
func (di *DeploymentInformer) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	key := name
	if namespace != "" && di.namespace == "" {
		// Multi-namespace informer needs namespace/name key
		key = namespace + "/" + name
	}

	obj, exists, err := di.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment from cache: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("deployment %s not found in cache", key)
	}

	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil, fmt.Errorf("object is not a deployment")
	}

	return deployment, nil
}

// ListDeployments lists all deployments from the informer's cache
func (di *DeploymentInformer) ListDeployments() ([]*appsv1.Deployment, error) {
	objs := di.informer.GetIndexer().List()
	deployments := make([]*appsv1.Deployment, 0, len(objs))

	for _, obj := range objs {
		deployment, ok := obj.(*appsv1.Deployment)
		if !ok {
			continue
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// HasSynced returns true if the informer's cache has synced
func (di *DeploymentInformer) HasSynced() bool {
	return di.informer.HasSynced()
}
