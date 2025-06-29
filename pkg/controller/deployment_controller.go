package controller

import (
	"context"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DeploymentReconciler logs every event for Deployment resources
// (add, update, delete) with detailed change analysis
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var deploy appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &deploy)
	if err != nil {
		// Object not found, means it was deleted
		logger.Info("Deployment event", map[string]interface{}{
			"event":     "delete",
			"namespace": req.Namespace,
			"name":      req.Name,
			"timestamp": deploy.DeletionTimestamp,
		})
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Determine event type based on generation and creation timestamp
	eventType := r.determineEventType(&deploy)
	
	switch eventType {
	case "add":
		r.logAddEvent(&deploy)
	case "update":
		r.logUpdateEvent(&deploy)
	default:
		r.logSyncEvent(&deploy)
	}
	
	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) determineEventType(deploy *appsv1.Deployment) string {
	// If generation is 1, it's a new deployment (add)
	if deploy.Generation == 1 {
		return "add"
	}
	// If generation > 1, it's an update to the spec
	if deploy.Generation > 1 {
		return "update"
	}
	// Otherwise it's a status sync
	return "sync"
}

func (r *DeploymentReconciler) logAddEvent(deploy *appsv1.Deployment) {
	logger.Info("Deployment event", map[string]interface{}{
		"event":      "add",
		"namespace":  deploy.Namespace,
		"name":       deploy.Name,
		"replicas":   r.getReplicasValue(deploy.Spec.Replicas),
		"image":      r.getFirstContainerImage(deploy),
		"labels":     deploy.Labels,
		"generation": deploy.Generation,
		"created":    deploy.CreationTimestamp,
	})
}

func (r *DeploymentReconciler) logUpdateEvent(deploy *appsv1.Deployment) {
	logger.Info("Deployment event", map[string]interface{}{
		"event":           "update",
		"namespace":       deploy.Namespace,
		"name":            deploy.Name,
		"generation":      deploy.Generation,
		"replicas":        r.getReplicasValue(deploy.Spec.Replicas),
		"image":           r.getFirstContainerImage(deploy),
		"labels":          deploy.Labels,
		"ready_replicas":  deploy.Status.ReadyReplicas,
		"updated_replicas": deploy.Status.UpdatedReplicas,
		"observed_generation": deploy.Status.ObservedGeneration,
	})
}

func (r *DeploymentReconciler) logSyncEvent(deploy *appsv1.Deployment) {
	logger.Debug("Deployment event", map[string]interface{}{
		"event":         "sync",
		"namespace":     deploy.Namespace,
		"name":          deploy.Name,
		"ready_replicas": deploy.Status.ReadyReplicas,
		"desired_replicas": r.getReplicasValue(deploy.Spec.Replicas),
	})
}

func (r *DeploymentReconciler) getReplicasValue(replicas *int32) int32 {
	if replicas == nil {
		return 1 // Default replica count
	}
	return *replicas
}

func (r *DeploymentReconciler) getFirstContainerImage(deploy *appsv1.Deployment) string {
	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		return deploy.Spec.Template.Spec.Containers[0].Image
	}
	return "unknown"
}

func AddDeploymentControllerToManager(mgr manager.Manager) error {
	r := &DeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
