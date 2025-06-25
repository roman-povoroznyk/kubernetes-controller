package kubernetes

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

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
	_, err := clientset.AppsV1().Deployments("default").Create(nil, deployment, metav1.CreateOptions{})
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
