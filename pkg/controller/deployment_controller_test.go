package controller

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewDeploymentController(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	controller := &DeploymentController{
		Scheme:          scheme,
		deploymentCache: make(map[string]*appsv1.Deployment),
	}
	
	if controller == nil {
		t.Fatal("Expected controller to be created")
	}
	
	if controller.Scheme != scheme {
		t.Fatal("Expected controller scheme to match")
	}
	
	if controller.deploymentCache == nil {
		t.Fatal("Expected deployment cache to be initialized")
	}
}

func TestDeploymentController_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)

	// Create a test deployment
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
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
	}

	// Create fake client
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(deployment).
		Build()

	// Create controller
	controller := &DeploymentController{
		Client:          fakeClient,
		Scheme:          scheme,
		deploymentCache: make(map[string]*appsv1.Deployment),
	}

	// Test reconcile
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-deployment",
			Namespace: "default",
		},
	}

	ctx := context.Background()
	result, err := controller.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	if result.RequeueAfter != time.Minute*5 {
		t.Fatalf("Expected requeue after 5 minutes, got %v", result.RequeueAfter)
	}

	// Check cache
	cachedDeployment, err := controller.GetDeployment("default", "test-deployment")
	if err != nil {
		t.Fatalf("Failed to get deployment from cache: %v", err)
	}

	if cachedDeployment.Name != "test-deployment" {
		t.Fatalf("Expected deployment name 'test-deployment', got %s", cachedDeployment.Name)
	}
}

func TestDeploymentController_ListDeployments(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	controller := &DeploymentController{
		Scheme:          scheme,
		deploymentCache: make(map[string]*appsv1.Deployment),
	}

	// Test empty cache
	deployments, err := controller.ListDeployments()
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}

	if len(deployments) != 0 {
		t.Fatalf("Expected 0 deployments, got %d", len(deployments))
	}

	// Add deployment to cache
	replicas := int32(2)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
	}

	controller.deploymentCache["default/test-deployment"] = deployment

	// Test non-empty cache
	deployments, err = controller.ListDeployments()
	if err != nil {
		t.Fatalf("ListDeployments failed: %v", err)
	}

	if len(deployments) != 1 {
		t.Fatalf("Expected 1 deployment, got %d", len(deployments))
	}

	if deployments[0].Name != "test-deployment" {
		t.Fatalf("Expected deployment name 'test-deployment', got %s", deployments[0].Name)
	}
}
