package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestDeploymentReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	reconciler := &DeploymentReconciler{
		Client: c,
		Scheme: scheme,
	}
	
	// Test case 1: Create a new deployment (add event)
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-deploy",
			Namespace:  "default",
			Generation: 1, // First generation = add event
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:1.14",
						},
					},
				},
			},
		},
	}
	if err := c.Create(context.TODO(), deploy); err != nil {
		t.Fatalf("failed to create deployment: %v", err)
	}
	
	// Reconcile for add event
	_, err := reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: client.ObjectKey{Namespace: "default", Name: "test-deploy"},
	})
	if err != nil {
		t.Errorf("unexpected error on add: %v", err)
	}

	// Test case 2: Update deployment (update event)
	deploy.Generation = 2 // Increment generation = update event
	deploy.Spec.Replicas = int32Ptr(3)
	deploy.Spec.Template.Spec.Containers[0].Image = "nginx:1.16"
	if err := c.Update(context.TODO(), deploy); err != nil {
		t.Fatalf("failed to update deployment: %v", err)
	}
	
	_, err = reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: client.ObjectKey{Namespace: "default", Name: "test-deploy"},
	})
	if err != nil {
		t.Errorf("unexpected error on update: %v", err)
	}
	
	// Test case 3: Delete deployment (delete event)
	if err := c.Delete(context.TODO(), deploy); err != nil {
		t.Fatalf("failed to delete deployment: %v", err)
	}
	
	_, err = reconciler.Reconcile(context.TODO(), reconcile.Request{
		NamespacedName: client.ObjectKey{Namespace: "default", Name: "test-deploy"},
	})
	if err != nil {
		t.Errorf("unexpected error on delete: %v", err)
	}
}

func TestDeploymentReconciler_EventTypes(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)

	reconciler := &DeploymentReconciler{
		Scheme: scheme,
	}
	
	// Test add event detection
	addDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Generation: 1},
	}
	if reconciler.determineEventType(addDeploy) != "add" {
		t.Errorf("expected add event, got %s", reconciler.determineEventType(addDeploy))
	}
	
	// Test update event detection
	updateDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Generation: 2},
	}
	if reconciler.determineEventType(updateDeploy) != "update" {
		t.Errorf("expected update event, got %s", reconciler.determineEventType(updateDeploy))
	}
	
	// Test sync event detection
	syncDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Generation: 0},
	}
	if reconciler.determineEventType(syncDeploy) != "sync" {
		t.Errorf("expected sync event, got %s", reconciler.determineEventType(syncDeploy))
	}
}

func int32Ptr(i int32) *int32 { return &i }
