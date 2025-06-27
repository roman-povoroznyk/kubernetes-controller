package utils

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// KubernetesHelpers contains common Kubernetes utility functions
type KubernetesHelpers struct{}

// GetReplicaCount safely gets replica count from pointer
func (h *KubernetesHelpers) GetReplicaCount(replicas *int32) int32 {
	if replicas == nil {
		return 0
	}
	return *replicas
}

// GetMainContainerImage gets the image of the first container in deployment
func (h *KubernetesHelpers) GetMainContainerImage(deployment *appsv1.Deployment) string {
	if deployment == nil || len(deployment.Spec.Template.Spec.Containers) == 0 {
		return "unknown"
	}
	return deployment.Spec.Template.Spec.Containers[0].Image
}

// GetMainPodImage gets the image of the first container in pod
func (h *KubernetesHelpers) GetMainPodImage(pod *corev1.Pod) string {
	if pod == nil || len(pod.Spec.Containers) == 0 {
		return "unknown"
	}
	return pod.Spec.Containers[0].Image
}

// GetPodReadyCondition checks if pod is ready
func (h *KubernetesHelpers) GetPodReadyCondition(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// GetPodRestartCount gets total restart count for all containers in pod
func (h *KubernetesHelpers) GetPodRestartCount(pod *corev1.Pod) int32 {
	if pod == nil {
		return 0
	}
	var count int32
	for _, status := range pod.Status.ContainerStatuses {
		count += status.RestartCount
	}
	return count
}

// Global instance for easy access
var K8sHelpers = &KubernetesHelpers{}
