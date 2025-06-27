package server

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/kubernetes"
	"github.com/valyala/fasthttp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentHandler_HandleDeployments(t *testing.T) {
	// Create fake Kubernetes client
	fakeClient := fake.NewSimpleClientset()

	// Create test deployment
	testDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
			CreationTimestamp: metav1.NewTime(time.Now().Add(-10 * time.Minute)),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:1.20",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     2,
			UpdatedReplicas:   3,
			AvailableReplicas: 2,
		},
	}

	// Add deployment to fake client
	_, err := fakeClient.AppsV1().Deployments("default").Create(
		context.TODO(), testDeploy, metav1.CreateOptions{},
	)
	if err != nil {
		t.Fatalf("Failed to create test deployment: %v", err)
	}

	// Create informer (note: this won't actually start in tests)
	informer := kubernetes.NewDeploymentInformer(fakeClient, "", 10*time.Minute)

	// Create handler
	handler := NewDeploymentHandler(informer)

	t.Run("List deployments when informer not started", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/api/v1/deployments")
		ctx.Request.Header.SetMethod("GET")

		handler.HandleDeployments(ctx)

		if ctx.Response.StatusCode() != fasthttp.StatusServiceUnavailable {
			t.Errorf("Expected status %d, got %d", fasthttp.StatusServiceUnavailable, ctx.Response.StatusCode())
		}

		var response ErrorResponse
		if err := json.Unmarshal(ctx.Response.Body(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Error != "Service unavailable" {
			t.Errorf("Expected error 'Service unavailable', got '%s'", response.Error)
		}
	})

	t.Run("Get deployment with invalid path", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/api/v1/deployments/too/many/parts")
		ctx.Request.Header.SetMethod("GET")

		handler.HandleDeployments(ctx)

		if ctx.Response.StatusCode() != fasthttp.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", fasthttp.StatusBadRequest, ctx.Response.StatusCode())
		}
	})

	t.Run("Unsupported method", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/api/v1/deployments")
		ctx.Request.Header.SetMethod("POST")

		handler.HandleDeployments(ctx)

		if ctx.Response.StatusCode() != fasthttp.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", fasthttp.StatusMethodNotAllowed, ctx.Response.StatusCode())
		}
	})

	t.Run("Invalid endpoint", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetRequestURI("/api/v1/invalid")
		ctx.Request.Header.SetMethod("GET")

		handler.HandleDeployments(ctx)

		if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
			t.Errorf("Expected status %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
		}
	})
}

func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		time     time.Time
		expected string
	}{
		{now.Add(-30 * time.Second), "30s"},
		{now.Add(-2 * time.Minute), "2m"},
		{now.Add(-3 * time.Hour), "3h"},
		{now.Add(-2 * 24 * time.Hour), "2d"},
	}

	for _, test := range tests {
		result := formatAge(test.time)
		if result != test.expected {
			t.Errorf("formatAge(%v) = %s, expected %s", test.time, result, test.expected)
		}
	}
}

func TestConvertDeploymentToResponse(t *testing.T) {
	handler := &DeploymentHandler{}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "production",
			Labels: map[string]string{
				"app":  "web",
				"tier": "frontend",
			},
			CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Hour)),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(5),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "web-container",
							Image: "nginx:1.21",
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     4,
			UpdatedReplicas:   5,
			AvailableReplicas: 4,
		},
	}

	response := handler.convertDeploymentToResponse(deployment)

	if response.Name != "test-deployment" {
		t.Errorf("Expected name 'test-deployment', got '%s'", response.Name)
	}

	if response.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got '%s'", response.Namespace)
	}

	if response.Replicas != 5 {
		t.Errorf("Expected replicas 5, got %d", response.Replicas)
	}

	if response.Ready != 4 {
		t.Errorf("Expected ready 4, got %d", response.Ready)
	}

	if response.Image != "nginx:1.21" {
		t.Errorf("Expected image 'nginx:1.21', got '%s'", response.Image)
	}

	if response.Labels["app"] != "web" {
		t.Errorf("Expected label app='web', got '%s'", response.Labels["app"])
	}

	if response.Age != "1h" {
		t.Errorf("Expected age '1h', got '%s'", response.Age)
	}
}

// int32Ptr returns a pointer to an int32 value
func int32Ptr(i int32) *int32 {
	return &i
}
