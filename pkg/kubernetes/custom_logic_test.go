package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentChangeAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "default", time.Second*30)
	analyzer := NewDeploymentChangeAnalyzer(informer)

	t.Run("AnalyzeUpdate - Replica Changes", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 1)
		newDeploy := createTestDeployment("test-app", "nginx:1.0", 3)

		changes := analyzer.AnalyzeUpdate(oldDeploy, newDeploy)

		require.Len(t, changes, 1)
		assert.Equal(t, "spec", changes[0].Type)
		assert.Equal(t, "replicas", changes[0].Field)
		assert.Equal(t, int32(1), changes[0].OldValue)
		assert.Equal(t, int32(3), changes[0].NewValue)
		assert.Contains(t, changes[0].Description, "Replicas changed from 1 to 3")
	})

	t.Run("AnalyzeUpdate - Image Changes", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 2)
		newDeploy := createTestDeployment("test-app", "nginx:1.1", 2)

		changes := analyzer.AnalyzeUpdate(oldDeploy, newDeploy)

		require.Len(t, changes, 1)
		assert.Equal(t, "spec", changes[0].Type)
		assert.Equal(t, "containers[0].image", changes[0].Field)
		assert.Equal(t, "nginx:1.0", changes[0].OldValue)
		assert.Equal(t, "nginx:1.1", changes[0].NewValue)
		assert.Contains(t, changes[0].Description, "image changed from nginx:1.0 to nginx:1.1")
	})

	t.Run("AnalyzeUpdate - Multiple Changes", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 1)
		newDeploy := createTestDeployment("test-app", "nginx:2.0", 5)

		changes := analyzer.AnalyzeUpdate(oldDeploy, newDeploy)

		assert.Len(t, changes, 2)
		
		// Should have both replica and image changes
		hasReplicaChange := false
		hasImageChange := false
		
		for _, change := range changes {
			if change.Field == "replicas" {
				hasReplicaChange = true
				assert.Equal(t, int32(1), change.OldValue)
				assert.Equal(t, int32(5), change.NewValue)
			}
			if change.Field == "containers[0].image" {
				hasImageChange = true
				assert.Equal(t, "nginx:1.0", change.OldValue)
				assert.Equal(t, "nginx:2.0", change.NewValue)
			}
		}
		
		assert.True(t, hasReplicaChange, "Should detect replica change")
		assert.True(t, hasImageChange, "Should detect image change")
	})

	t.Run("AnalyzeUpdate - Labels Changes", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 2)
		oldDeploy.Labels = map[string]string{"app": "test", "version": "v1"}
		
		newDeploy := createTestDeployment("test-app", "nginx:1.0", 2)
		newDeploy.Labels = map[string]string{"app": "test", "version": "v2", "env": "prod"}

		changes := analyzer.AnalyzeUpdate(oldDeploy, newDeploy)

		require.Len(t, changes, 1)
		assert.Equal(t, "metadata", changes[0].Type)
		assert.Equal(t, "labels", changes[0].Field)
		assert.Equal(t, "Labels changed", changes[0].Description)
	})

	t.Run("AnalyzeUpdate - Resource Changes", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 2)
		oldDeploy.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
		}
		
		newDeploy := createTestDeployment("test-app", "nginx:1.0", 2)
		newDeploy.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		}

		changes := analyzer.AnalyzeUpdate(oldDeploy, newDeploy)

		require.Len(t, changes, 1)
		assert.Equal(t, "spec", changes[0].Type)
		assert.Equal(t, "containers[0].resources", changes[0].Field)
		assert.Contains(t, changes[0].Description, "resources changed")
	})

	t.Run("AnalyzeUpdate - No Changes", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 2)
		newDeploy := createTestDeployment("test-app", "nginx:1.0", 2)

		changes := analyzer.AnalyzeUpdate(oldDeploy, newDeploy)

		assert.Len(t, changes, 0)
	})
}

func TestCustomLogicEventHandler(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "default", time.Second*30)
	handler := NewCustomLogicEventHandler(informer)

	t.Run("OnAdd", func(t *testing.T) {
		deploy := createTestDeployment("test-app", "nginx:1.0", 2)
		
		// Should not panic
		assert.NotPanics(t, func() {
			handler.OnAdd(deploy)
		})
	})

	t.Run("OnUpdate", func(t *testing.T) {
		oldDeploy := createTestDeployment("test-app", "nginx:1.0", 1)
		newDeploy := createTestDeployment("test-app", "nginx:2.0", 3)
		
		// Should not panic
		assert.NotPanics(t, func() {
			handler.OnUpdate(oldDeploy, newDeploy)
		})
	})

	t.Run("OnDelete", func(t *testing.T) {
		deploy := createTestDeployment("test-app", "nginx:1.0", 2)
		
		// Should not panic
		assert.NotPanics(t, func() {
			handler.OnDelete(deploy)
		})
	})
}

func TestAnalyzeDelete(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "default", time.Second*30)
	analyzer := NewDeploymentChangeAnalyzer(informer)

	t.Run("AnalyzeDelete - Basic Analysis", func(t *testing.T) {
		deploy := createTestDeployment("test-app", "nginx:1.0", 3)
		deploy.Labels = map[string]string{"app": "test-app", "env": "prod"}
		
		analysis := analyzer.AnalyzeDelete(deploy)

		assert.Equal(t, "not_found", analysis["cache_status"])
		assert.Equal(t, true, analysis["had_replicas"])
		assert.Equal(t, "default", analysis["namespace"])
		assert.Equal(t, int64(1), analysis["generation"])
		assert.NotNil(t, analysis["labels"])
		assert.NotNil(t, analysis["creation_timestamp"])
	})

	t.Run("AnalyzeDelete - Zero Replicas", func(t *testing.T) {
		deploy := createTestDeployment("test-app", "nginx:1.0", 0)
		
		analysis := analyzer.AnalyzeDelete(deploy)

		assert.Equal(t, false, analysis["had_replicas"])
	})
}

// Helper function to create test deployments
func createTestDeployment(name, image string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
			Generation:        1,
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "25%",
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ObservedGeneration: 1,
			Replicas:           replicas,
			UpdatedReplicas:    replicas,
			ReadyReplicas:      replicas,
			AvailableReplicas:  replicas,
		},
	}
}
