package controller

import (
	"context"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles deployment events and logs them with structured logging
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.With().
		Str("namespace", req.Namespace).
		Str("name", req.Name).
		Str("component", "deployment-controller").
		Logger()

	// Try to fetch the deployment to determine the event type
	var deployment appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Deployment was deleted
			logger.Info().Msg("Deployment DELETE event received")
			return ctrl.Result{}, nil
		}
		logger.Error().Err(err).Msg("Failed to get deployment")
		return ctrl.Result{}, err
	}

	// Deployment exists - log event with details
	logger.Info().
		Time("created", deployment.CreationTimestamp.Time).
		Int32("replicas", getReplicaCount(deployment.Spec.Replicas)).
		Int32("ready_replicas", deployment.Status.ReadyReplicas).
		Int32("available_replicas", deployment.Status.AvailableReplicas).
		Str("generation", deployment.ResourceVersion).
		Msgf("Deployment %s event received", getEventType(&deployment))

	return ctrl.Result{}, nil
}

// getEventType determines if this is a CREATE or UPDATE event
func getEventType(deployment *appsv1.Deployment) string {
	if deployment.Status.ObservedGeneration == 0 {
		return "CREATE"
	}
	return "UPDATE"
}

// getReplicaCount safely gets replica count from pointer
func getReplicaCount(replicas *int32) int32 {
	if replicas == nil {
		return 0
	}
	return *replicas
}

// AddDeploymentController adds the deployment controller to the manager
func AddDeploymentController(mgr manager.Manager) error {
	r := &DeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}
