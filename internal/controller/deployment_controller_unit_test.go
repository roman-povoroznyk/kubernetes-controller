package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDeploymentReconciler_Reconcile(t *testing.T) {
	// Create scheme
	scheme := runtime.NewScheme()
	err := clientgoscheme.AddToScheme(scheme)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		deployment     *appsv1.Deployment
		expectError    bool
		expectedResult ctrl.Result
	}{
		{
			name: "reconcile existing deployment",
			deployment: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "default",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: &[]int32{1}[0],
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
			},
			expectError:    false,
			expectedResult: ctrl.Result{},
		},
		{
			name:           "reconcile non-existent deployment",
			deployment:     nil, // This will test the deletion case
			expectError:    false,
			expectedResult: ctrl.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			var objs []runtime.Object
			if tt.deployment != nil {
				objs = append(objs, tt.deployment)
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objs...).
				Build()

			// Create reconciler
			r := &DeploymentReconciler{
				Client: fakeClient,
				Scheme: scheme,
			}

			// Create reconcile request
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-deployment",
					Namespace: "default",
				},
			}

			// Call reconcile
			result, err := r.Reconcile(context.TODO(), req)

			// Check results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetReplicaCount(t *testing.T) {
	tests := []struct {
		name     string
		replicas *int32
		expected int32
	}{
		{
			name:     "nil replicas",
			replicas: nil,
			expected: 0,
		},
		{
			name:     "zero replicas",
			replicas: &[]int32{0}[0],
			expected: 0,
		},
		{
			name:     "positive replicas",
			replicas: &[]int32{3}[0],
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReplicaCount(tt.replicas)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEventType(t *testing.T) {
	tests := []struct {
		name       string
		deployment *appsv1.Deployment
		expected   string
	}{
		{
			name: "create event - zero observed generation",
			deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 0,
				},
			},
			expected: "CREATE",
		},
		{
			name: "update event - non-zero observed generation",
			deployment: &appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					ObservedGeneration: 1,
				},
			},
			expected: "UPDATE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEventType(tt.deployment)
			assert.Equal(t, tt.expected, result)
		})
	}
}
