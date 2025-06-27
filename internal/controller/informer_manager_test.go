package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes/fake"
)

func TestInformerManager(t *testing.T) {
	// Test creating informer manager
	fakeClient := fake.NewSimpleClientset()
	namespace := "test-namespace"
	resyncPeriod := 30 * time.Second

	informerMgr := NewInformerManager(fakeClient, namespace, resyncPeriod)

	assert.NotNil(t, informerMgr, "InformerManager should be created")
	assert.Equal(t, namespace, informerMgr.namespace, "Namespace should be set correctly")
	assert.Equal(t, resyncPeriod, informerMgr.resyncPeriod, "Resync period should be set correctly")
	assert.Equal(t, fakeClient, informerMgr.client, "Client should be set correctly")
}

func TestInformerRunnableLeaderElection(t *testing.T) {
	// Test that informer runnables don't need leader election
	fakeClient := fake.NewSimpleClientset()
	
	deploymentInformer := &DeploymentInformerRunnable{
		client:       fakeClient,
		namespace:    "default",
		resyncPeriod: 30 * time.Second,
	}
	
	podInformer := &PodInformerRunnable{
		client:       fakeClient,
		namespace:    "default", 
		resyncPeriod: 30 * time.Second,
	}

	assert.False(t, deploymentInformer.NeedsLeaderElection(), "Deployment informer should not need leader election")
	assert.False(t, podInformer.NeedsLeaderElection(), "Pod informer should not need leader election")
}
