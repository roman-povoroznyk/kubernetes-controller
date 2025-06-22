package kubeops

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func init() {
	// Disable logging during tests
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func TestCreatePod(t *testing.T) {
	client := fake.NewSimpleClientset()
	ns := "default"
	name := "test-pod"

	err := CreatePod(client, ns, name)
	if err != nil {
		t.Fatalf("CreatePod failed: %v", err)
	}

	pod, err := client.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Pod not found after creation: %v", err)
	}
	if pod.Name != name {
		t.Errorf("expected pod name %q, got %q", name, pod.Name)
	}
}

func TestListPods(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"}},
	)
	ns := "default"

	err := ListPods(client, ns)
	if err != nil {
		t.Fatalf("ListPods failed: %v", err)
	}
}

func TestDeletePod(t *testing.T) {
	client := fake.NewSimpleClientset(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-to-delete", Namespace: "default"}},
	)
	ns := "default"
	name := "pod-to-delete"

	err := DeletePod(client, ns, name)
	if err != nil {
		t.Fatalf("DeletePod failed: %v", err)
	}

	_, err = client.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err == nil {
		t.Errorf("Pod was not deleted")
	}
}
