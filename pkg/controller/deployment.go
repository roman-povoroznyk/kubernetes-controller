package controller

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/roman-povoroznyk/k8s/pkg/business"
)

// DeploymentReconciler reconciles Deployment objects
type DeploymentReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RuleEngine *business.RuleEngine
}

// Reconcile handles deployment reconciliation
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Info().
		Str("namespace", req.Namespace).
		Str("name", req.Name).
		Msg("Reconciling deployment")

	// Fetch the deployment
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if errors.IsNotFound(err) {
			log.Info().
				Str("namespace", req.Namespace).
				Str("name", req.Name).
				Msg("Deployment not found, assuming deleted")
			return ctrl.Result{}, nil
		}
		log.Error().Err(err).Msg("Failed to get deployment")
		return ctrl.Result{}, err
	}

	// Validate business rules
	if err := r.RuleEngine.ValidateDeployment(&deployment); err != nil {
		log.Error().Err(err).Msg("Business rule validation failed")
		// Could add status update here to reflect validation failure
		return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
	}

	// Custom logic for deployment management
	if err := r.handleDeploymentLogic(ctx, &deployment); err != nil {
		log.Error().Err(err).Msg("Failed to handle deployment logic")
		return ctrl.Result{}, err
	}

	log.Info().
		Str("namespace", req.Namespace).
		Str("name", req.Name).
		Msg("Successfully reconciled deployment")

	return ctrl.Result{}, nil
}

// handleDeploymentLogic implements custom deployment management logic
func (r *DeploymentReconciler) handleDeploymentLogic(ctx context.Context, deployment *appsv1.Deployment) error {
	// Example: Ensure deployment has proper annotations
	if deployment.Annotations == nil {
		deployment.Annotations = make(map[string]string)
	}

	// Add managed annotation
	if deployment.Annotations["managed-by"] != "kubernetes-controller" {
		deployment.Annotations["managed-by"] = "kubernetes-controller"
		deployment.Annotations["managed-at"] = time.Now().Format(time.RFC3339)
		
		if err := r.Update(ctx, deployment); err != nil {
			return err
		}
		
		log.Info().
			Str("namespace", deployment.Namespace).
			Str("name", deployment.Name).
			Msg("Updated deployment with management annotations")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
