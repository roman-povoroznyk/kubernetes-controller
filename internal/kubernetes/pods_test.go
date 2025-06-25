package kubernetes

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreatePod(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Test creating a pod
	err := CreatePod(clientset, "default", "test-pod")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the pod was created
	pod, err := clientset.CoreV1().Pods("default").Get(context.TODO(), "test-pod", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Expected pod to exist, got error: %v", err)
	}
	if pod.Name != "test-pod" {
		t.Errorf("Expected pod name 'test-pod', got '%s'", pod.Name)
	}
}

func TestDeletePod(t *testing.T) {
	// Create a fake clientset with a pod
	clientset := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	})

	// Test deleting the pod
	err := DeletePod(clientset, "default", "test-pod")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the pod was deleted
	_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "test-pod", metav1.GetOptions{})
	if err == nil {
		t.Error("Expected error (pod not found), got nil")
	}
}

func TestListPods(t *testing.T) {
	// Create a fake clientset with some pods
	clientset := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "default",
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "default",
			},
		},
	)

	// Test listing pods
	err := ListPods(clientset, "default")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}
