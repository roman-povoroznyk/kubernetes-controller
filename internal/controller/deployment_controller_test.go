//go:build integration

package controller

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
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestDeploymentController(t *testing.T) {
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

	// Create scheme
	scheme := runtime.NewScheme()
	err = clientgoscheme.AddToScheme(scheme)
	require.NoError(t, err)

	// Create manager
	mgr, err := ctrl.NewManager(cfg, manager.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)

	// Add deployment controller
	err = AddDeploymentController(mgr)
	require.NoError(t, err)

	// Start manager in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mgrDone := make(chan error)
	go func() {
		mgrDone <- mgr.Start(ctx)
	}()

	// Wait for manager to start
	time.Sleep(2 * time.Second)

	// Create client for direct access
	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	require.NoError(t, err)

	// Create test namespace
	testNamespace := "test-controller"
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
	}
	err = k8sClient.Create(ctx, namespace)
	require.NoError(t, err)

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

	// Test CREATE event
	err = k8sClient.Create(ctx, deployment)
	require.NoError(t, err)

	// Wait for controller to process
	time.Sleep(1 * time.Second)

	// Test UPDATE event
	deployment.Spec.Replicas = int32Ptr(2)
	err = k8sClient.Update(ctx, deployment)
	require.NoError(t, err)

	// Wait for controller to process
	time.Sleep(1 * time.Second)

	// Test DELETE event
	err = k8sClient.Delete(ctx, deployment)
	require.NoError(t, err)

	// Wait for controller to process
	time.Sleep(1 * time.Second)

	// Cancel context to stop manager
	cancel()

	// Wait for manager to finish
	select {
	case err := <-mgrDone:
		// Context cancellation is expected
		if err != nil && err != context.Canceled {
			t.Errorf("Manager exited with unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Manager did not stop within timeout")
	}
}

func int32Ptr(i int32) *int32 {
	return &i
}
