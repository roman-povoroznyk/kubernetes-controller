package informer

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestDeploymentHelperFunctions(t *testing.T) {
	t.Run("getReplicaCount with nil", func(t *testing.T) {
		result := getReplicaCount(nil)
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("getReplicaCount with value", func(t *testing.T) {
		replicas := int32(3)
		result := getReplicaCount(&replicas)
		if result != 3 {
			t.Errorf("Expected 3, got %d", result)
		}
	})

	t.Run("getMainContainerImage with containers", func(t *testing.T) {
		deployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Image: "nginx:latest"},
							{Image: "redis:alpine"},
						},
					},
				},
			},
		}

		result := getMainContainerImage(deployment)
		if result != "nginx:latest" {
			t.Errorf("Expected 'nginx:latest', got '%s'", result)
		}
	})

	t.Run("getMainContainerImage with no containers", func(t *testing.T) {
		deployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
		}

		result := getMainContainerImage(deployment)
		if result != "unknown" {
			t.Errorf("Expected 'unknown', got '%s'", result)
		}
	})
}

func TestGetDeploymentNamesWithoutInformer(t *testing.T) {
	// Test when informer is nil
	names := GetDeploymentNames()
	if len(names) != 0 {
		t.Errorf("Expected empty slice when informer is nil, got %v", names)
	}
}

func TestGetDeploymentsWithoutInformer(t *testing.T) {
	// Test when informer is nil
	deployments := GetDeployments()
	if len(deployments) != 0 {
		t.Errorf("Expected empty slice when informer is nil, got %v", deployments)
	}
}

func TestGetDeploymentByNameWithoutInformer(t *testing.T) {
	// Test when informer is nil
	deployment, found := GetDeploymentByName("test")
	if found {
		t.Error("Expected not found when informer is nil")
	}
	if deployment != nil {
		t.Errorf("Expected nil deployment when informer is nil, got %v", deployment)
	}
}
