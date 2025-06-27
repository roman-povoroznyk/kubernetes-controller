//go:build integration

package informer

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestPodInformer(t *testing.T) {
	// Set up envtest
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "testdata", "crds")},
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	defer func() {
		err := testEnv.Stop()
		assert.NoError(t, err)
	}()

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// Create test namespace
	testNamespace := "test-pod-informer"
	_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Set up informer config
	config := PodInformerConfig{
		Namespace:    testNamespace,
		ResyncPeriod: 1 * time.Second,
	}

	// Start informer in background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	informerDone := make(chan error)
	go func() {
		informerDone <- StartPodInformer(ctx, clientset, config)
	}()

	// Wait a bit for informer to start
	time.Sleep(2 * time.Second)

	// Create a test pod
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: testNamespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "nginx:latest",
				},
			},
		},
	}

	// Create pod
	_, err = clientset.CoreV1().Pods(testNamespace).Create(context.TODO(), testPod, metav1.CreateOptions{})
	require.NoError(t, err)

	// Wait for informer to process the event
	time.Sleep(1 * time.Second)

	// Update pod (add a label)
	testPod.Labels = map[string]string{"test": "updated"}
	_, err = clientset.CoreV1().Pods(testNamespace).Update(context.TODO(), testPod, metav1.UpdateOptions{})
	require.NoError(t, err)

	// Wait for update event
	time.Sleep(1 * time.Second)

	// Delete pod
	err = clientset.CoreV1().Pods(testNamespace).Delete(context.TODO(), testPod.Name, metav1.DeleteOptions{})
	require.NoError(t, err)

	// Wait for delete event
	time.Sleep(1 * time.Second)

	// Cancel context to stop informer
	cancel()

	// Wait for informer to finish
	select {
	case err := <-informerDone:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Informer did not stop within timeout")
	}
}

func TestPodInformerHelperFunctions(t *testing.T) {
	// Test pod with container
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: "nginx:1.20"},
			},
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{RestartCount: 3},
				{RestartCount: 2},
			},
		},
	}

	assert.Equal(t, "nginx:1.20", getMainPodImage(pod))
	assert.True(t, getPodReadyCondition(pod))
	assert.Equal(t, 5, getPodRestartCount(pod))

	// Test empty pod
	emptyPod := &corev1.Pod{}
	assert.Equal(t, "unknown", getMainPodImage(emptyPod))
	assert.False(t, getPodReadyCondition(emptyPod))
	assert.Equal(t, 0, getPodRestartCount(emptyPod))
}
