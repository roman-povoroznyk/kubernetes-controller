package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// DeploymentController reconciles a Deployment object
type DeploymentController struct {
	client.Client
	Scheme *runtime.Scheme
	
	// Cache for deployments (optional, for API access)
	deploymentCache map[string]*appsv1.Deployment
}

// NewDeploymentController creates a new deployment controller
func NewDeploymentController(mgr ctrl.Manager) *DeploymentController {
	return &DeploymentController{
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		deploymentCache: make(map[string]*appsv1.Deployment),
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *DeploymentController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the Deployment instance
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Deployment was deleted
			logger.Info("Deployment deleted", map[string]interface{}{
				"name":      req.Name,
				"namespace": req.Namespace,
				"event":     "DELETE",
			})
			
			// Remove from cache
			key := fmt.Sprintf("%s/%s", req.Namespace, req.Name)
			delete(r.deploymentCache, key)
			
			return ctrl.Result{}, nil
		}
		logger.Error("Failed to get deployment", err, map[string]interface{}{
			"name":      req.Name,
			"namespace": req.Namespace,
		})
		return ctrl.Result{}, err
	}

	// Update cache
	key := fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)
	oldDeployment, exists := r.deploymentCache[key]
	r.deploymentCache[key] = &deployment

	if !exists {
		// New deployment
		logger.Info("Deployment added", map[string]interface{}{
			"name":      deployment.Name,
			"namespace": deployment.Namespace,
			"replicas":  *deployment.Spec.Replicas,
			"image":     getFirstContainerImage(&deployment),
			"event":     "ADD",
		})
	} else {
		// Updated deployment
		if oldDeployment.ResourceVersion != deployment.ResourceVersion {
			logger.Info("Deployment updated", map[string]interface{}{
				"name":                   deployment.Name,
				"namespace":              deployment.Namespace,
				"replicas":               *deployment.Spec.Replicas,
				"ready_replicas":         deployment.Status.ReadyReplicas,
				"old_resource_version":   oldDeployment.ResourceVersion,
				"new_resource_version":   deployment.ResourceVersion,
				"event":                  "UPDATE",
			})
		}
	}

	// Here you can add business logic for processing deployments
	// For now, we just log and return
	
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}

// GetDeployment gets a deployment from the controller's cache
func (r *DeploymentController) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	deployment, exists := r.deploymentCache[key]
	if !exists {
		return nil, fmt.Errorf("deployment %s not found in cache", key)
	}
	
	// Return a copy to avoid modifications
	deploymentCopy := deployment.DeepCopy()
	return deploymentCopy, nil
}

// ListDeployments lists all deployments from the controller's cache
func (r *DeploymentController) ListDeployments() ([]*appsv1.Deployment, error) {
	deployments := make([]*appsv1.Deployment, 0, len(r.deploymentCache))
	
	for _, deployment := range r.deploymentCache {
		// Return copies to avoid modifications
		deployments = append(deployments, deployment.DeepCopy())
	}
	
	return deployments, nil
}

// getFirstContainerImage safely gets the image of the first container
func getFirstContainerImage(deployment *appsv1.Deployment) string {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		return deployment.Spec.Template.Spec.Containers[0].Image
	}
	return "unknown"
}
