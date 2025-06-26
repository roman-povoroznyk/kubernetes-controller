package informer

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// PodInformerConfig holds configuration for the pod informer
type PodInformerConfig struct {
	Namespace    string
	ResyncPeriod time.Duration
}

// StartPodInformer starts a shared informer for Pods
func StartPodInformer(ctx context.Context, clientset *kubernetes.Clientset, config PodInformerConfig) error {
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

	informer := factory.Core().V1().Pods().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				log.Warn().Msg("Received non-pod object in AddFunc")
				return
			}
			log.Info().
				Str("event", "ADD").
				Str("pod", pod.Name).
				Str("namespace", pod.Namespace).
				Str("phase", string(pod.Status.Phase)).
				Str("node", pod.Spec.NodeName).
				Str("image", getMainPodImage(pod)).
				Msg("Pod added")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*corev1.Pod)
			if !ok {
				log.Warn().Msg("Received non-pod object in UpdateFunc")
				return
			}
			oldPod, ok := oldObj.(*corev1.Pod)
			if !ok {
				log.Warn().Msg("Received non-pod object in UpdateFunc for oldObj")
				return
			}

			// Only log meaningful updates (phase changes, readiness changes)
			if pod.Status.Phase != oldPod.Status.Phase ||
				getPodReadyCondition(pod) != getPodReadyCondition(oldPod) ||
				pod.Status.ContainerStatuses != nil && oldPod.Status.ContainerStatuses != nil &&
					len(pod.Status.ContainerStatuses) > 0 && len(oldPod.Status.ContainerStatuses) > 0 &&
					pod.Status.ContainerStatuses[0].Ready != oldPod.Status.ContainerStatuses[0].Ready {
				log.Info().
					Str("event", "UPDATE").
					Str("pod", pod.Name).
					Str("namespace", pod.Namespace).
					Str("phase", string(pod.Status.Phase)).
					Str("node", pod.Spec.NodeName).
					Bool("ready", getPodReadyCondition(pod)).
					Int("restart_count", getPodRestartCount(pod)).
					Msg("Pod updated")
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				// Handle DeletedFinalStateUnknown
				if deletedFinalStateUnknown, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					pod, ok = deletedFinalStateUnknown.Obj.(*corev1.Pod)
					if !ok {
						log.Warn().Msg("Received non-pod object in DeletedFinalStateUnknown")
						return
					}
				} else {
					log.Warn().Msg("Received non-pod object in DeleteFunc")
					return
				}
			}
			log.Info().
				Str("event", "DELETE").
				Str("pod", pod.Name).
				Str("namespace", pod.Namespace).
				Str("phase", string(pod.Status.Phase)).
				Msg("Pod deleted")
		},
	})

	log.Info().
		Str("namespace", config.Namespace).
		Dur("resync_period", config.ResyncPeriod).
		Msg("Starting pod informer...")

	factory.Start(ctx.Done())

	// Wait for cache sync
	for t, ok := range factory.WaitForCacheSync(ctx.Done()) {
		if !ok {
			err := fmt.Errorf("failed to sync informer for %v", t)
			log.Error().Err(err).Msg("Pod informer cache sync failed")
			return err
		}
	}

	log.Info().Msg("Pod informer cache synced. Watching for events...")
	<-ctx.Done() // Block until context is cancelled
	log.Info().Msg("Pod informer stopped")
	return nil
}

// Helper functions for pod logging

// getMainPodImage gets the image of the first container
func getMainPodImage(pod *corev1.Pod) string {
	if len(pod.Spec.Containers) > 0 {
		return pod.Spec.Containers[0].Image
	}
	return "unknown"
}

// getPodReadyCondition checks if pod is ready
func getPodReadyCondition(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// getPodRestartCount gets total restart count for all containers
func getPodRestartCount(pod *corev1.Pod) int {
	totalRestarts := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		totalRestarts += int(containerStatus.RestartCount)
	}
	return totalRestarts
}
