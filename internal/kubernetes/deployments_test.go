package kubernetes

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateDeployment(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a deployment
	err := CreateDeployment(clientset, "default", "test-deployment")
	if err != nil {
		t.Fatalf("Error creating deployment: %v", err)
	}

	// Verify the deployment was created
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "test-deployment", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Error getting created deployment: %v", err)
	}

	if deployment.Name != "test-deployment" {
		t.Errorf("Expected deployment name 'test-deployment', got '%s'", deployment.Name)
	}

	if *deployment.Spec.Replicas != 1 {
		t.Errorf("Expected 1 replica, got %d", *deployment.Spec.Replicas)
	}

	// Test creating with existing name (should fail)
	err = CreateDeployment(clientset, "default", "test-deployment")
	if err == nil {
		t.Error("Expected error when creating deployment with duplicate name, got nil")
	}
}

func TestDeleteDeployment(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a test deployment first
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
	}

	// Add the deployment to the fake clientset
	_, err := clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating test deployment: %v", err)
	}

	// Delete the deployment
	err = DeleteDeployment(clientset, "default", "test-deployment")
	if err != nil {
		t.Errorf("Error deleting deployment: %v", err)
	}

	// Verify the deployment was deleted
	_, err = clientset.AppsV1().Deployments("default").Get(context.Background(), "test-deployment", metav1.GetOptions{})
	if err == nil {
		t.Error("Expected error getting deleted deployment, got nil")
	}

	// Test deleting non-existent deployment
	err = DeleteDeployment(clientset, "default", "nonexistent")
	if err == nil {
		t.Error("Expected error deleting non-existent deployment, got nil")
	}
}

func TestListDeployments(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a test deployment
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas:     2,
			UpdatedReplicas:   3,
			AvailableReplicas: 2,
		},
	}

	// Add the deployment to the fake clientset
	_, err := clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Error creating test deployment: %v", err)
	}

	// Test listing deployments
	err = ListDeployments(clientset, "default")
	if err != nil {
		t.Errorf("Error listing deployments: %v", err)
	}

	// Test listing deployments in empty namespace
	err = ListDeployments(clientset, "non-existent")
	if err != nil {
		t.Errorf("Error listing deployments in empty namespace: %v", err)
	}
}
