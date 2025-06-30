// pkg/testing/helpers.go
package testing

import (
	"context"
	"testing"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// TestEnvironment provides a test environment for integration tests
type TestEnvironment struct {
	TestEnv   *envtest.Environment
	Config    *config.Config
	Manager   manager.Manager
	K8sClient client.Client
	Scheme    *runtime.Scheme
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{},
		ErrorIfCRDPathMissing: false,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("Failed to start test environment: %v", err)
	}

	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("Failed to add scheme: %v", err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return &TestEnvironment{
		TestEnv:   testEnv,
		Config:    createTestConfig(),
		K8sClient: k8sClient,
		Scheme:    scheme,
	}
}

// Stop stops the test environment
func (te *TestEnvironment) Stop() error {
	return te.TestEnv.Stop()
}

// createTestConfig creates a test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		LogLevel: "debug",
		Controller: config.ControllerConfig{
			Mode: "single",
			Single: config.SingleClusterConfig{
				Namespace:   "default",
				MetricsPort: 8080,
				HealthPort:  8081,
				LeaderElection: config.LeaderElectionConfig{
					Enabled:   false,
					ID:        "test-controller",
					Namespace: "default",
				},
			},
			ResyncPeriod: 30 * time.Second,
		},
	}
}

// CreateTestDeployment creates a test deployment
func CreateTestDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
}

// WaitForDeployment waits for a deployment to be ready
func WaitForDeployment(ctx context.Context, client client.Client, name, namespace string, timeout time.Duration) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return timeoutCtx.Err()
		default:
			deployment := &appsv1.Deployment{}
			if err := client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment); err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				return nil
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

// MockKubernetesClient creates a fake Kubernetes client for testing
func MockKubernetesClient() *fake.Clientset {
	return fake.NewSimpleClientset()
}

// AssertEventually asserts that a condition becomes true within a timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	start := time.Now()
	for time.Since(start) < timeout {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Condition not met within timeout: %s", message)
}

// AssertNever asserts that a condition never becomes true within a timeout
func AssertNever(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	start := time.Now()
	for time.Since(start) < timeout {
		if condition() {
			t.Fatalf("Condition became true when it should not have: %s", message)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
