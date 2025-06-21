package api

import (
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentProvider provides an interface for accessing deployment data
// This abstraction allows the API to work with different sources:
// - Original informer-based implementation
// - New controller-runtime based implementation
type DeploymentProvider interface {
	GetDeployment(namespace, name string) (*appsv1.Deployment, error)
	ListDeployments() ([]*appsv1.Deployment, error)
	HasSynced() bool
}

// ControllerAdapter adapts the controller-runtime controller to the DeploymentProvider interface
type ControllerAdapter struct {
	Controller interface {
		GetDeployment(namespace, name string) (*appsv1.Deployment, error)
		ListDeployments() ([]*appsv1.Deployment, error)
	}
}

// NewControllerAdapter creates a new adapter for controller-runtime controller
func NewControllerAdapter(controller interface {
	GetDeployment(namespace, name string) (*appsv1.Deployment, error)
	ListDeployments() ([]*appsv1.Deployment, error)
}) *ControllerAdapter {
	return &ControllerAdapter{
		Controller: controller,
	}
}

// GetDeployment implements DeploymentProvider
func (c *ControllerAdapter) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	return c.Controller.GetDeployment(namespace, name)
}

// ListDeployments implements DeploymentProvider
func (c *ControllerAdapter) ListDeployments() ([]*appsv1.Deployment, error) {
	return c.Controller.ListDeployments()
}

// HasSynced implements DeploymentProvider
func (c *ControllerAdapter) HasSynced() bool {
	// Controller-runtime controllers are always considered synced once reconcile starts
	return true
}
