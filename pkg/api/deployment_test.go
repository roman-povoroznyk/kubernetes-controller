package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/informer"
	"github.com/roman-povoroznyk/k6s/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentAPI_ConvertDeploymentToResponse(t *testing.T) {
	// Create mock deployment
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test",
			},
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     3,
			UpdatedReplicas:   3,
			AvailableReplicas: 3,
		},
	}

	api := &DeploymentAPI{}
	response := api.ConvertDeploymentToResponse(deployment)

	// Verify response
	if response.Name != "test-deployment" {
		t.Errorf("Expected name 'test-deployment', got '%s'", response.Name)
	}
	
	if response.Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got '%s'", response.Namespace)
	}
	
	if response.Replicas != 3 {
		t.Errorf("Expected replicas 3, got %d", response.Replicas)
	}
	
	if response.Image != "nginx:latest" {
		t.Errorf("Expected image 'nginx:latest', got '%s'", response.Image)
	}
	
	if response.Status != "Ready" {
		t.Errorf("Expected status 'Ready', got '%s'", response.Status)
	}
}

func TestDeploymentAPI_ListDeployments_NoInformer(t *testing.T) {
	api := &DeploymentAPI{} // No informer set
	
	req := httptest.NewRequest("GET", "/api/v1/deployments", nil)
	w := httptest.NewRecorder()

	// This should handle the case where informer is nil gracefully
	api.ListDeployments(w, req)

	// Should return service unavailable since informer is nil
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d, body: %s", http.StatusServiceUnavailable, w.Code, w.Body.String())
	}
}

func TestDeploymentAPI_HealthCheck(t *testing.T) {
	// This test requires a running Kubernetes cluster for a real informer
	t.Skip("Skipping API tests - requires running cluster")
	
	client, err := kubernetes.NewClient("")
	if err != nil {
		t.Skipf("Cannot create Kubernetes client: %v", err)
	}

	deploymentInformer := informer.NewDeploymentInformer(client, "default", 30*time.Second)
	api := NewDeploymentAPI(deploymentInformer)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	api.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var health map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &health)
	if err != nil {
		t.Errorf("Failed to unmarshal health response: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", health["status"])
	}

	if _, exists := health["cache_synced"]; !exists {
		t.Errorf("Expected cache_synced field in health response")
	}
}
