package controller

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// InformerManager integrates custom informers with controller-runtime manager
type InformerManager struct {
	client        kubernetes.Interface
	namespace     string
	resyncPeriod  time.Duration
}

// NewInformerManager creates a new informer manager
func NewInformerManager(client kubernetes.Interface, namespace string, resyncPeriod time.Duration) *InformerManager {
	return &InformerManager{
		client:       client,
		namespace:    namespace,
		resyncPeriod: resyncPeriod,
	}
}

// AddToManager adds informers to the controller-runtime manager as runnables
func (im *InformerManager) AddToManager(mgr manager.Manager, enableDeploymentInformer, enablePodInformer bool) error {
	if enableDeploymentInformer {
		deploymentInformer := &DeploymentInformerRunnable{
			client:       im.client,
			namespace:    im.namespace,
			resyncPeriod: im.resyncPeriod,
		}
		if err := mgr.Add(deploymentInformer); err != nil {
			return err
		}
		log.Info().Msg("Added deployment informer to controller-runtime manager")
	}

	if enablePodInformer {
		podInformer := &PodInformerRunnable{
			client:       im.client,
			namespace:    im.namespace,
			resyncPeriod: im.resyncPeriod,
		}
		if err := mgr.Add(podInformer); err != nil {
			return err
		}
		log.Info().Msg("Added pod informer to controller-runtime manager")
	}

	return nil
}

// DeploymentInformerRunnable implements manager.Runnable for deployment informer
type DeploymentInformerRunnable struct {
	client       kubernetes.Interface
	namespace    string
	resyncPeriod time.Duration
}

// Start implements manager.Runnable
func (d *DeploymentInformerRunnable) Start(ctx context.Context) error {
	log.Info().
		Str("namespace", d.namespace).
		Dur("resync-period", d.resyncPeriod).
		Msg("Starting deployment informer via controller-runtime manager")

	watchlist := cache.NewListWatchFromClient(
		d.client.AppsV1().RESTClient(),
		"deployments",
		d.namespace,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		nil, // No specific object type needed for logging
		d.resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Info().
					Str("component", "deployment-informer-manager").
					Msg("Deployment added via manager")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Info().
					Str("component", "deployment-informer-manager").
					Msg("Deployment updated via manager")
			},
			DeleteFunc: func(obj interface{}) {
				log.Info().
					Str("component", "deployment-informer-manager").
					Msg("Deployment deleted via manager")
			},
		},
	)

	controller.Run(ctx.Done())
	return nil
}

// PodInformerRunnable implements manager.Runnable for pod informer
type PodInformerRunnable struct {
	client       kubernetes.Interface
	namespace    string
	resyncPeriod time.Duration
}

// Start implements manager.Runnable
func (p *PodInformerRunnable) Start(ctx context.Context) error {
	log.Info().
		Str("namespace", p.namespace).
		Dur("resync-period", p.resyncPeriod).
		Msg("Starting pod informer via controller-runtime manager")

	watchlist := cache.NewListWatchFromClient(
		p.client.CoreV1().RESTClient(),
		"pods",
		p.namespace,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		nil, // No specific object type needed for logging
		p.resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Info().
					Str("component", "pod-informer-manager").
					Msg("Pod added via manager")
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Info().
					Str("component", "pod-informer-manager").
					Msg("Pod updated via manager")
			},
			DeleteFunc: func(obj interface{}) {
				log.Info().
					Str("component", "pod-informer-manager").
					Msg("Pod deleted via manager")
			},
		},
	)

	controller.Run(ctx.Done())
	return nil
}

// NeedsLeaderElection implements manager.LeaderElectionRunnable (optional)
func (d *DeploymentInformerRunnable) NeedsLeaderElection() bool {
	return false // Informers don't need leader election
}

func (p *PodInformerRunnable) NeedsLeaderElection() bool {
	return false // Informers don't need leader election
}
