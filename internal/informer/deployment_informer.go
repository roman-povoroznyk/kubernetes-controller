package informer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Global informer instances for API access
var (
	deploymentInformer cache.SharedIndexInformer
	deploymentMutex    sync.RWMutex
)

// DeploymentInformerConfig holds configuration for the deployment informer
type DeploymentInformerConfig struct {
	Namespace    string
	ResyncPeriod time.Duration
}

// StartDeploymentInformer starts a shared informer for Deployments
func StartDeploymentInformer(ctx context.Context, clientset *kubernetes.Clientset, config DeploymentInformerConfig) error {
	if config.ResyncPeriod == 0 {
		config.ResyncPeriod = 30 * time.Second
	}

	if config.Namespace == "" {
		config.Namespace = "default"
	}

	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		config.ResyncPeriod,
		informers.WithNamespace(config.Namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.FieldSelector = fields.Everything().String()
		}),
	)

	informer := factory.Apps().V1().Deployments().Informer()

	// Store informer globally for API access
	deploymentMutex.Lock()
	deploymentInformer = informer
	deploymentMutex.Unlock()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment, ok := obj.(*appsv1.Deployment)
			if !ok {
				log.Warn().Msg("Received non-deployment object in AddFunc")
				return
			}
			log.Info().
				Str("event", "ADD").
				Str("deployment", deployment.Name).
				Str("namespace", deployment.Namespace).
				Int32("replicas", getReplicaCount(deployment.Spec.Replicas)).
				Str("image", getMainContainerImage(deployment)).
				Msg("Deployment added")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deployment, ok := newObj.(*appsv1.Deployment)
			if !ok {
				log.Warn().Msg("Received non-deployment object in UpdateFunc")
				return
			}
			oldDeployment, ok := oldObj.(*appsv1.Deployment)
			if !ok {
				log.Warn().Msg("Received non-deployment object in UpdateFunc for oldObj")
				return
			}

			// Only log meaningful updates
			if deployment.Generation != oldDeployment.Generation ||
				deployment.Status.ReadyReplicas != oldDeployment.Status.ReadyReplicas ||
				deployment.Status.UpdatedReplicas != oldDeployment.Status.UpdatedReplicas {
				log.Info().
					Str("event", "UPDATE").
					Str("deployment", deployment.Name).
					Str("namespace", deployment.Namespace).
					Int32("ready", deployment.Status.ReadyReplicas).
					Int32("updated", deployment.Status.UpdatedReplicas).
					Int32("replicas", getReplicaCount(deployment.Spec.Replicas)).
					Int64("generation", deployment.Generation).
					Msg("Deployment updated")
			}
		},
		DeleteFunc: func(obj interface{}) {
			deployment, ok := obj.(*appsv1.Deployment)
			if !ok {
				// Handle DeletedFinalStateUnknown
				if deletedFinalStateUnknown, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					deployment, ok = deletedFinalStateUnknown.Obj.(*appsv1.Deployment)
					if !ok {
						log.Warn().Msg("Received non-deployment object in DeletedFinalStateUnknown")
						return
					}
				} else {
					log.Warn().Msg("Received non-deployment object in DeleteFunc")
					return
				}
			}
			log.Info().
				Str("event", "DELETE").
				Str("deployment", deployment.Name).
				Str("namespace", deployment.Namespace).
				Msg("Deployment deleted")
		},
	})

	log.Info().
		Str("namespace", config.Namespace).
		Dur("resync_period", config.ResyncPeriod).
		Msg("Starting deployment informer...")

	factory.Start(ctx.Done())

	// Wait for cache sync
	for t, ok := range factory.WaitForCacheSync(ctx.Done()) {
		if !ok {
			err := fmt.Errorf("failed to sync informer for %v", t)
			log.Error().Err(err).Msg("Informer cache sync failed")
			return err
		}
	}

	log.Info().Msg("Deployment informer cache synced. Watching for events...")
	<-ctx.Done() // Block until context is cancelled
	log.Info().Msg("Deployment informer stopped")
	return nil
}

// Helper functions for better logging

// getReplicaCount safely gets replica count from pointer
func getReplicaCount(replicas *int32) int32 {
	if replicas == nil {
		return 0
	}
	return *replicas
}

// getMainContainerImage gets the image of the first container
func getMainContainerImage(deployment *appsv1.Deployment) string {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		return deployment.Spec.Template.Spec.Containers[0].Image
	}
	return "unknown"
}

// API access functions

// GetDeploymentNames returns a slice of deployment names from the informer's cache
func GetDeploymentNames() []string {
	deploymentMutex.RLock()
	defer deploymentMutex.RUnlock()

	var names []string
	if deploymentInformer == nil {
		return names
	}

	for _, obj := range deploymentInformer.GetStore().List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			names = append(names, deployment.Name)
		}
	}
	return names
}

// GetDeployments returns a slice of deployments from the informer's cache
func GetDeployments() []*appsv1.Deployment {
	deploymentMutex.RLock()
	defer deploymentMutex.RUnlock()

	var deployments []*appsv1.Deployment
	if deploymentInformer == nil {
		return deployments
	}

	for _, obj := range deploymentInformer.GetStore().List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			deployments = append(deployments, deployment)
		}
	}
	return deployments
}

// GetDeploymentByName returns a specific deployment by name from the informer's cache
func GetDeploymentByName(name string) (*appsv1.Deployment, bool) {
	deploymentMutex.RLock()
	defer deploymentMutex.RUnlock()

	if deploymentInformer == nil {
		return nil, false
	}

	for _, obj := range deploymentInformer.GetStore().List() {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			if deployment.Name == name {
				return deployment, true
			}
		}
	}
	return nil, false
}
