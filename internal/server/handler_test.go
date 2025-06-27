package server

import (
	"testing"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/internal/utils"
	"github.com/valyala/fasthttp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestFormatAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		timeAgo  time.Duration
		expected string
	}{
		{"30 seconds", 30 * time.Second, "30s"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"2 hours", 2 * time.Hour, "2h"},
		{"3 days", 72 * time.Hour, "3d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := now.Add(-tt.timeAgo)
			result := formatAge(testTime)

			// For seconds test, we accept 29s-31s range due to timing
			if tt.name == "30 seconds" {
				if result != "29s" && result != "30s" && result != "31s" {
					t.Errorf("Expected around 30s, got %s", result)
				}
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetReplicaCountHandler(t *testing.T) {
	tests := []struct {
		name     string
		replicas *int32
		expected int32
	}{
		{"nil replicas", nil, 0},
		{"zero replicas", int32Ptr(0), 0},
		{"positive replicas", int32Ptr(3), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.K8sHelpers.GetReplicaCount(tt.replicas)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetMainContainerImageHandler(t *testing.T) {
	t.Run("with containers", func(t *testing.T) {
		deployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Image: "nginx:latest"},
						},
					},
				},
			},
		}

		result := utils.K8sHelpers.GetMainContainerImage(deployment)
		if result != "nginx:latest" {
			t.Errorf("Expected 'nginx:latest', got '%s'", result)
		}
	})

	t.Run("without containers", func(t *testing.T) {
		deployment := &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
		}

		result := utils.K8sHelpers.GetMainContainerImage(deployment)
		if result != "unknown" {
			t.Errorf("Expected 'unknown', got '%s'", result)
		}
	})
}

func TestGetMainPodImageHandler(t *testing.T) {
	t.Run("with containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "redis:alpine"},
				},
			},
		}

		result := utils.K8sHelpers.GetMainPodImage(pod)
		if result != "redis:alpine" {
			t.Errorf("Expected 'redis:alpine', got '%s'", result)
		}
	})

	t.Run("without containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{},
			},
		}

		result := utils.K8sHelpers.GetMainPodImage(pod)
		if result != "unknown" {
			t.Errorf("Expected 'unknown', got '%s'", result)
		}
	})
}

func TestGetPodReadyStatus(t *testing.T) {
	t.Run("ready containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Ready: true},
					{Ready: true},
				},
			},
		}

		result := getPodReadyStatus(pod)
		if result != "2/2" {
			t.Errorf("Expected '2/2', got '%s'", result)
		}
	})

	t.Run("mixed ready containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{Ready: true},
					{Ready: false},
				},
			},
		}

		result := getPodReadyStatus(pod)
		if result != "1/2" {
			t.Errorf("Expected '1/2', got '%s'", result)
		}
	})

	t.Run("no containers", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{},
			},
		}

		result := getPodReadyStatus(pod)
		if result != "0/0" {
			t.Errorf("Expected '0/0', got '%s'", result)
		}
	})
}

func TestGetPodRestartCountHandler(t *testing.T) {
	t.Run("with restarts", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{RestartCount: 2},
					{RestartCount: 3},
				},
			},
		}

		result := int(utils.K8sHelpers.GetPodRestartCount(pod))
		if result != 5 {
			t.Errorf("Expected 5, got %d", result)
		}
	})

	t.Run("no restarts", func(t *testing.T) {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{RestartCount: 0},
					{RestartCount: 0},
				},
			},
		}

		result := int(utils.K8sHelpers.GetPodRestartCount(pod))
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})
}

func TestAPIEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{"Deployments names endpoint", "/deployments/names", "GET", 200},
		{"Deployments endpoint", "/deployments", "GET", 200},
		{"Pods names endpoint", "/pods/names", "GET", 200},
		{"Pods endpoint", "/pods", "GET", 200},
		{"POST to deployments (not allowed)", "/deployments", "POST", 405},
		{"PUT to pods (not allowed)", "/pods", "PUT", 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod(tt.method)

			HandleRequests(ctx)

			if ctx.Response.StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, ctx.Response.StatusCode())
			}
		})
	}
}

func TestIndividualResourceEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
		setupMock      func()
	}{
		{
			name:           "Get deployment by name - not found (no informer)",
			path:           "/deployments/test-deployment",
			method:         "GET",
			expectedStatus: 404, // Expected since no informer is running in unit tests
			setupMock: func() {
				// In unit tests, informer is not running so this will return 404
			},
		},
		{
			name:           "Get deployment by name - not found",
			path:           "/deployments/nonexistent",
			method:         "GET",
			expectedStatus: 404,
			setupMock: func() {
				// Mock would be set up in integration tests
			},
		},
		{
			name:           "Get pod by name - not found (no informer)",
			path:           "/pods/test-pod",
			method:         "GET",
			expectedStatus: 404, // Expected since no informer is running in unit tests
			setupMock: func() {
				// In unit tests, informer is not running so this will return 404
			},
		},
		{
			name:           "Get pod by name - not found",
			path:           "/pods/nonexistent",
			method:         "GET",
			expectedStatus: 404,
			setupMock: func() {
				// Mock would be set up in integration tests
			},
		},
		{
			name:           "POST to individual deployment - not allowed",
			path:           "/deployments/test-deployment",
			method:         "POST",
			expectedStatus: 405,
			setupMock:      func() {},
		},
		{
			name:           "PUT to individual pod - not allowed",
			path:           "/pods/test-pod",
			method:         "PUT",
			expectedStatus: 405,
			setupMock:      func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.SetMethod(tt.method)
			ctx.Request.SetRequestURI(tt.path)

			HandleRequests(ctx)

			if ctx.Response.StatusCode() != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, ctx.Response.StatusCode())
			}
		})
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
