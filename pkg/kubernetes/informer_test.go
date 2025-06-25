package kubernetes

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestEventHandler implements DeploymentEventHandler for testing
type TestEventHandler struct {
	OnAddCalled    bool
	OnUpdateCalled bool
	OnDeleteCalled bool
	LastDeployment *appsv1.Deployment
}

func (h *TestEventHandler) OnAdd(obj *appsv1.Deployment) {
	h.OnAddCalled = true
	h.LastDeployment = obj
}

func (h *TestEventHandler) OnUpdate(oldObj, newObj *appsv1.Deployment) {
	h.OnUpdateCalled = true
	h.LastDeployment = newObj
}

func (h *TestEventHandler) OnDelete(obj *appsv1.Deployment) {
	h.OnDeleteCalled = true
	h.LastDeployment = obj
}

func TestNewDeploymentInformer(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	tests := []struct {
		name         string
		namespace    string
		resyncPeriod time.Duration
		expected     struct {
			namespace    string
			resyncPeriod time.Duration
		}
	}{
		{
			name:         "default namespace and resync period",
			namespace:    "",
			resyncPeriod: 0,
			expected: struct {
				namespace    string
				resyncPeriod time.Duration
			}{
				namespace:    metav1.NamespaceAll,
				resyncPeriod: 30 * time.Second,
			},
		},
		{
			name:         "specific namespace and resync period",
			namespace:    "test-namespace",
			resyncPeriod: 1 * time.Minute,
			expected: struct {
				namespace    string
				resyncPeriod time.Duration
			}{
				namespace:    "test-namespace",
				resyncPeriod: 1 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			informer := NewDeploymentInformer(clientset, tt.namespace, tt.resyncPeriod)

			if informer == nil {
				t.Fatal("expected informer to be created, got nil")
			}

			if informer.namespace != tt.expected.namespace {
				t.Errorf("expected namespace %s, got %s", tt.expected.namespace, informer.namespace)
			}

			if informer.resyncPeriod != tt.expected.resyncPeriod {
				t.Errorf("expected resync period %v, got %v", tt.expected.resyncPeriod, informer.resyncPeriod)
			}

			if informer.started {
				t.Error("expected informer to not be started initially")
			}

			if len(informer.eventHandlers) != 1 {
				t.Errorf("expected 1 default event handler, got %d", len(informer.eventHandlers))
			}
		})
	}
}

func TestDeploymentInformer_AddEventHandler(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "test", 30*time.Second)

	testHandler := &TestEventHandler{}
	informer.AddEventHandler(testHandler)

	if len(informer.eventHandlers) != 2 { // 1 default + 1 test handler
		t.Errorf("expected 2 event handlers, got %d", len(informer.eventHandlers))
	}
}

func TestDeploymentInformer_IsStarted(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "test", 30*time.Second)

	if informer.IsStarted() {
		t.Error("expected informer to not be started initially")
	}

	// Mock starting the informer by setting the flag directly
	informer.started = true

	if !informer.IsStarted() {
		t.Error("expected informer to be started after setting flag")
	}
}

func TestDeploymentInformer_Stop(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "test", 30*time.Second)

	// Mock starting the informer
	informer.started = true
	
	// Test stopping when started
	informer.Stop()

	if informer.IsStarted() {
		t.Error("expected informer to be stopped after calling Stop()")
	}

	// Test stopping when already stopped (should not panic)
	informer.Stop()
}

func TestDefaultDeploymentEventHandler(t *testing.T) {
	handler := &DefaultDeploymentEventHandler{}
	
	// Create test deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
		},
	}

	// Test that methods don't panic (they just log)
	handler.OnAdd(deployment)
	handler.OnUpdate(deployment, deployment)
	handler.OnDelete(deployment)
}

func TestDeploymentInformer_GetDeployment_NotStarted(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "test", 30*time.Second)

	_, err := informer.GetDeployment("test", "deployment")
	if err == nil {
		t.Error("expected error when getting deployment from non-started informer")
	}

	expectedErr := "informer is not started"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestDeploymentInformer_ListDeployments_NotStarted(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "test", 30*time.Second)

	_, err := informer.ListDeployments()
	if err == nil {
		t.Error("expected error when listing deployments from non-started informer")
	}

	expectedErr := "informer is not started"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestDeploymentInformer_HasSynced(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	informer := NewDeploymentInformer(clientset, "test", 30*time.Second)

	// Before starting, HasSynced should return false
	if informer.HasSynced() {
		t.Error("expected HasSynced to return false for non-started informer")
	}
}

// Helper function to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}
