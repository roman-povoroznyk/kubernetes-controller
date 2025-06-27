package informer

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"github.com/roman-povoroznyk/k8s/pkg/business"
)

// DeploymentInformer wraps Kubernetes deployment informer
type DeploymentInformer struct {
	clientset   kubernetes.Interface
	informer    cache.SharedIndexInformer
	config      *InformerConfig
	ruleEngine  *business.RuleEngine
}

// NewDeploymentInformer creates a new deployment informer with config
func NewDeploymentInformer(clientset kubernetes.Interface, config *InformerConfig) *DeploymentInformer {
	if !config.IsResourceEnabled("deployments") {
		log.Info().Msg("Deployments informer is disabled by configuration")
		return nil
	}

	factory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		config.ResyncPeriod,
		informers.WithNamespace(metav1.NamespaceAll),
	)

	informer := factory.Apps().V1().Deployments().Informer()
	
	di := &DeploymentInformer{
		clientset:  clientset,
		informer:   informer,
		config:     config,
		ruleEngine: business.NewRuleEngine(),
	}

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    di.onAdd,
		UpdateFunc: di.onUpdate,
		DeleteFunc: di.onDelete,
	})

	return di
}

// Start starts the informer
func (di *DeploymentInformer) Start(ctx context.Context) error {
	if di == nil {
		return nil
	}

	log.Info().Msg("Starting deployment informer...")
	
	go di.informer.Run(ctx.Done())

	// Wait for cache to sync
	if !cache.WaitForCacheSync(ctx.Done(), di.informer.HasSynced) {
		return fmt.Errorf("failed to sync deployment informer cache")
	}

	log.Info().Msg("Deployment informer cache synced")
	return nil
}

func (di *DeploymentInformer) onAdd(obj interface{}) {
	deployment := obj.(*appsv1.Deployment)
	
	if !di.config.IsNamespaceWatched(deployment.Namespace) {
		return
	}

	log.Info().
		Str("name", deployment.Name).
		Str("namespace", deployment.Namespace).
		Msg("Deployment added")

	// Validate business rules
	if err := di.ruleEngine.ValidateDeployment(deployment); err != nil {
		log.Error().Err(err).Msg("Business rule validation failed for new deployment")
	}
}

func (di *DeploymentInformer) onUpdate(oldObj, newObj interface{}) {
	deployment := newObj.(*appsv1.Deployment)
	
	if !di.config.IsNamespaceWatched(deployment.Namespace) {
		return
	}

	log.Info().
		Str("name", deployment.Name).
		Str("namespace", deployment.Namespace).
		Msg("Deployment updated")

	// Validate business rules
	if err := di.ruleEngine.ValidateDeployment(deployment); err != nil {
		log.Error().Err(err).Msg("Business rule validation failed for updated deployment")
	}
}

func (di *DeploymentInformer) onDelete(obj interface{}) {
	deployment := obj.(*appsv1.Deployment)
	
	if !di.config.IsNamespaceWatched(deployment.Namespace) {
		return
	}

	log.Info().
		Str("name", deployment.Name).
		Str("namespace", deployment.Namespace).
		Msg("Deployment deleted")
}
