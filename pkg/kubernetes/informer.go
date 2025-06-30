package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// DeploymentInformer manages deployment informers with list/watch functionality
type DeploymentInformer struct {
	clientset       kubernetes.Interface
	informer        cache.SharedIndexInformer
	stopper         chan struct{}
	namespace       string
	resyncPeriod    time.Duration
	started         bool
	mu              sync.RWMutex
	eventHandlers   []DeploymentEventHandler
}

// DeploymentEventHandler defines the interface for handling deployment events
type DeploymentEventHandler interface {
	OnAdd(obj *appsv1.Deployment)
	OnUpdate(oldObj, newObj *appsv1.Deployment)
	OnDelete(obj *appsv1.Deployment)
}

// DefaultDeploymentEventHandler provides a default implementation with logging
type DefaultDeploymentEventHandler struct{}

func (h *DefaultDeploymentEventHandler) OnAdd(obj *appsv1.Deployment) {
	log.Info().
		Str("namespace", obj.Namespace).
		Str("name", obj.Name).
		Int32("replicas", *obj.Spec.Replicas).
		Msg("Deployment added")
}

func (h *DefaultDeploymentEventHandler) OnUpdate(oldObj, newObj *appsv1.Deployment) {
	logEvent := log.Info().
		Str("namespace", newObj.Namespace).
		Str("name", newObj.Name)

	// Check for replica changes
	if oldObj.Spec.Replicas != nil && newObj.Spec.Replicas != nil && *oldObj.Spec.Replicas != *newObj.Spec.Replicas {
		logEvent = logEvent.
			Int32("old_replicas", *oldObj.Spec.Replicas).
			Int32("new_replicas", *newObj.Spec.Replicas)
	}

	// Check for image changes
	if len(oldObj.Spec.Template.Spec.Containers) > 0 && len(newObj.Spec.Template.Spec.Containers) > 0 {
		oldImage := oldObj.Spec.Template.Spec.Containers[0].Image
		newImage := newObj.Spec.Template.Spec.Containers[0].Image
		if oldImage != newImage {
			logEvent = logEvent.
				Str("old_image", oldImage).
				Str("new_image", newImage)
		}
	}

	// Check for generation changes (indicates spec changes)
	if oldObj.Generation != newObj.Generation {
		logEvent = logEvent.
			Int64("generation", newObj.Generation).
			Str("change_type", "spec")
	} else if oldObj.Status.ObservedGeneration != newObj.Status.ObservedGeneration {
		logEvent = logEvent.
			Int64("observed_generation", newObj.Status.ObservedGeneration).
			Str("change_type", "status")
	}

	logEvent.Msg("Deployment updated")
}

func (h *DefaultDeploymentEventHandler) OnDelete(obj *appsv1.Deployment) {
	log.Info().
		Str("namespace", obj.Namespace).
		Str("name", obj.Name).
		Msg("Deployment deleted")
}

// KubectlStyleEventHandler provides kubectl-like output without logs
type KubectlStyleEventHandler struct{}

func (h *KubectlStyleEventHandler) OnAdd(obj *appsv1.Deployment) {
	fmt.Printf("ADDED     %s/%s\n", obj.Namespace, obj.Name)
}

func (h *KubectlStyleEventHandler) OnUpdate(oldObj, newObj *appsv1.Deployment) {
	fmt.Printf("MODIFIED  %s/%s\n", newObj.Namespace, newObj.Name)
}

func (h *KubectlStyleEventHandler) OnDelete(obj *appsv1.Deployment) {
	fmt.Printf("DELETED   %s/%s\n", obj.Namespace, obj.Name)
}

// NewDeploymentInformer creates a new deployment informer
func NewDeploymentInformer(clientset kubernetes.Interface, namespace string, resyncPeriod time.Duration) *DeploymentInformer {
	if resyncPeriod == 0 {
		resyncPeriod = 30 * time.Second
	}

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.AppsV1().Deployments(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), options)
		},
	}

	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&appsv1.Deployment{},
		resyncPeriod,
		cache.Indexers{},
	)

	di := &DeploymentInformer{
		clientset:    clientset,
		informer:     informer,
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
		stopper:      make(chan struct{}),
		started:      false,
	}

	// Add default event handler
	di.AddEventHandler(&DefaultDeploymentEventHandler{})

	return di
}

// NewDeploymentInformerWithKubectlStyle creates a new deployment informer with kubectl-style output
func NewDeploymentInformerWithKubectlStyle(clientset kubernetes.Interface, namespace string, resyncPeriod time.Duration) *DeploymentInformer {
	if resyncPeriod == 0 {
		resyncPeriod = 30 * time.Second
	}

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.AppsV1().Deployments(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), options)
		},
	}

	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&appsv1.Deployment{},
		resyncPeriod,
		cache.Indexers{},
	)

	di := &DeploymentInformer{
		clientset:    clientset,
		informer:     informer,
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
		stopper:      make(chan struct{}),
		started:      false,
	}

	// Add kubectl-style event handler instead of default
	di.AddEventHandler(&KubectlStyleEventHandler{})

	return di
}

// NewDeploymentInformerWithCustomLogic creates a new deployment informer with custom logic handler
func NewDeploymentInformerWithCustomLogic(clientset kubernetes.Interface, namespace string, resyncPeriod time.Duration) *DeploymentInformer {
	if resyncPeriod == 0 {
		resyncPeriod = 30 * time.Second
	}

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.AppsV1().Deployments(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), options)
		},
	}

	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&appsv1.Deployment{},
		resyncPeriod,
		cache.Indexers{},
	)

	di := &DeploymentInformer{
		clientset:    clientset,
		informer:     informer,
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
		stopper:      make(chan struct{}),
		started:      false,
	}

	// Add custom logic event handler instead of default
	di.AddEventHandler(NewCustomLogicEventHandler(di))

	return di
}

// NewDeploymentInformerWithConfig creates a new deployment informer using configuration
func NewDeploymentInformerWithConfig(clientset kubernetes.Interface, cfg *config.Config) *DeploymentInformer {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	namespace := cfg.Controller.Single.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	resyncPeriod := cfg.Controller.ResyncPeriod
	if resyncPeriod == 0 {
		resyncPeriod = 30 * time.Second
	}

	// Create list/watch functions
	listWatcher := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.AppsV1().Deployments(namespace).List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.AppsV1().Deployments(namespace).Watch(context.TODO(), options)
		},
	}

	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&appsv1.Deployment{},
		resyncPeriod,
		cache.Indexers{},
	)

	di := &DeploymentInformer{
		clientset:    clientset,
		informer:     informer,
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
		stopper:      make(chan struct{}),
		started:      false,
	}

	// Add default event handler
	di.AddEventHandler(&DefaultDeploymentEventHandler{})

	log.Debug().
		Str("namespace", namespace).
		Dur("resync_period", resyncPeriod).
		Msg("Created deployment informer with configuration")

	return di
}

// AddEventHandler adds an event handler to the informer
func (di *DeploymentInformer) AddEventHandler(handler DeploymentEventHandler) {
	di.mu.Lock()
	defer di.mu.Unlock()

	di.eventHandlers = append(di.eventHandlers, handler)

	if di.started {
		log.Warn().Msg("Adding event handler to already started informer")
	}
}

// Start starts the informer
func (di *DeploymentInformer) Start() error {
	di.mu.Lock()
	defer di.mu.Unlock()

	if di.started {
		return fmt.Errorf("informer is already started")
	}

	// Add event handlers to the informer
	_, err := di.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				for _, handler := range di.eventHandlers {
					handler.OnAdd(deployment)
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if oldDeployment, ok := oldObj.(*appsv1.Deployment); ok {
				if newDeployment, ok := newObj.(*appsv1.Deployment); ok {
					for _, handler := range di.eventHandlers {
						handler.OnUpdate(oldDeployment, newDeployment)
					}
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if deployment, ok := obj.(*appsv1.Deployment); ok {
				for _, handler := range di.eventHandlers {
					handler.OnDelete(deployment)
				}
			}
		},
	})

	if err != nil {
		return fmt.Errorf("failed to add event handler: %w", err)
	}

	// Start the informer
	go di.informer.Run(di.stopper)

	// Wait for cache to sync
	if !cache.WaitForCacheSync(di.stopper, di.informer.HasSynced) {
		close(di.stopper)
		return fmt.Errorf("failed to sync cache")
	}

	di.started = true

	return nil
}

// Stop stops the informer
func (di *DeploymentInformer) Stop() {
	di.mu.Lock()
	defer di.mu.Unlock()

	if !di.started {
		return
	}

	close(di.stopper)
	di.started = false
}

// IsStarted returns whether the informer is started
func (di *DeploymentInformer) IsStarted() bool {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.started
}

// GetDeployment retrieves a deployment from the cache
func (di *DeploymentInformer) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	if !di.IsStarted() {
		return nil, fmt.Errorf("informer is not started")
	}

	key := name
	if namespace != "" {
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
		return nil, fmt.Errorf("unexpected object type in cache")
	}

	return deployment, nil
}

// ListDeployments returns all deployments from the cache
func (di *DeploymentInformer) ListDeployments() ([]*appsv1.Deployment, error) {
	if !di.IsStarted() {
		return nil, fmt.Errorf("informer is not started")
	}

	objects := di.informer.GetIndexer().List()
	deployments := make([]*appsv1.Deployment, 0, len(objects))

	for _, obj := range objects {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			deployments = append(deployments, deployment)
		}
	}

	return deployments, nil
}

// HasSynced returns true if the informer's cache has synced
func (di *DeploymentInformer) HasSynced() bool {
	return di.informer.HasSynced()
}
