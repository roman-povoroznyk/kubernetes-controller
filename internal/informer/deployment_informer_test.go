// +build integration

package informer

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestDeploymentInformer(t *testing.T) {
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
	testNamespace := "test-informer"
	_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	// Set up informer config
	config := DeploymentInformerConfig{
		Namespace:    testNamespace,
		ResyncPeriod: 1 * time.Second,
	}

	// Start informer in background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	informerDone := make(chan error)
	go func() {
		informerDone <- StartDeploymentInformer(ctx, clientset, config)
	}()

	// Wait a bit for informer to start
	time.Sleep(2 * time.Second)

	// Create a test deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: testNamespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
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
							Name:  "test",
							Image: "nginx:alpine",
						},
					},
				},
			},
		},
	}

	_, err = clientset.AppsV1().Deployments(testNamespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	require.NoError(t, err)

	// Update the deployment
	deployment.Spec.Replicas = int32Ptr(2)
	_, err = clientset.AppsV1().Deployments(testNamespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	require.NoError(t, err)

	// Delete the deployment
	err = clientset.AppsV1().Deployments(testNamespace).Delete(context.TODO(), "test-deployment", metav1.DeleteOptions{})
	require.NoError(t, err)

	// Wait a bit for events to be processed
	time.Sleep(2 * time.Second)

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

func int32Ptr(i int32) *int32 {
	return &i
}
