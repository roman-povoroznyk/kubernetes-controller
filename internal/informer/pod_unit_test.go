package informer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestPodHelperFunctions(t *testing.T) {
	t.Run("getMainPodImage with containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "nginx:latest"},
					{Image: "redis:alpine"},
				},
			},
		}

		result := getMainPodImage(pod)
		if result != "nginx:latest" {
			t.Errorf("Expected 'nginx:latest', got '%s'", result)
		}
	})

	t.Run("getMainPodImage with no containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{},
			},
		}

		result := getMainPodImage(pod)
		if result != "unknown" {
			t.Errorf("Expected 'unknown', got '%s'", result)
		}
	})

	t.Run("getPodReadyCondition with ready pod", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}

		result := getPodReadyCondition(pod)
		if !result {
			t.Error("Expected pod to be ready")
		}
	})

	t.Run("getPodReadyCondition with not ready pod", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		}

		result := getPodReadyCondition(pod)
		if result {
			t.Error("Expected pod to not be ready")
		}
	})

	t.Run("getPodReadyCondition with no conditions", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{},
			},
		}

		result := getPodReadyCondition(pod)
		if result {
			t.Error("Expected pod to not be ready when no conditions")
		}
	})

	t.Run("getPodRestartCount with restarts", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{RestartCount: 2},
					{RestartCount: 3},
				},
			},
		}

		result := getPodRestartCount(pod)
		if result != 5 {
			t.Errorf("Expected 5 restarts, got %d", result)
		}
	})

	t.Run("getPodRestartCount with no containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{},
			},
		}

		result := getPodRestartCount(pod)
		if result != 0 {
			t.Errorf("Expected 0 restarts, got %d", result)
		}
	})
}

func TestGetPodNamesWithoutInformer(t *testing.T) {
	// Test when informer is nil
	names := GetPodNames()
	if len(names) != 0 {
		t.Errorf("Expected empty slice when informer is nil, got %v", names)
	}
}

func TestGetPodsWithoutInformer(t *testing.T) {
	// Test when informer is nil
	pods := GetPods()
	if len(pods) != 0 {
		t.Errorf("Expected empty slice when informer is nil, got %v", pods)
	}
}

func TestGetPodByNameWithoutInformer(t *testing.T) {
	// Test when informer is nil
	pod, found := GetPodByName("test")
	if found {
		t.Error("Expected not found when informer is nil")
	}
	if pod != nil {
		t.Errorf("Expected nil pod when informer is nil, got %v", pod)
	}
}
