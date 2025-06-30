package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	
	// Configuration
	cluster     string
	namespace   string
	concurrency int
}

// NewDeploymentReconciler creates a new DeploymentReconciler
func NewDeploymentReconciler(mgr manager.Manager, cluster, namespace string, concurrency int) *DeploymentReconciler {
	return &DeploymentReconciler{
		Client:      mgr.GetClient(),
		Log:         logger.WithComponent("deployment-controller").WithCluster(cluster).GetLogr(),
		Scheme:      mgr.GetScheme(),
		cluster:     cluster,
		namespace:   namespace,
		concurrency: concurrency,
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *DeploymentReconciler) SetupWithManager(mgr manager.Manager) error {
	// Build the controller with predicates
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(r.createEventFilter()).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.concurrency,
		}).
		Complete(r)
}

// createEventFilter creates event filters for the controller
func (r *DeploymentReconciler) createEventFilter() predicate.Predicate {
	return predicate.And(
		predicate.GenerationChangedPredicate{},
		predicate.ResourceVersionChangedPredicate{},
		r.createNamespaceFilter(),
	)
}

// createNamespaceFilter creates a namespace filter if specified
func (r *DeploymentReconciler) createNamespaceFilter() predicate.Predicate {
	if r.namespace == "" {
		return predicate.NewPredicateFuncs(func(object client.Object) bool {
			return true
		})
	}
	
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetNamespace() == r.namespace
	})
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("deployment", req.NamespacedName)
	
	// Start timing
	start := time.Now()
	defer func() {
		log.V(1).Info("Reconciliation completed", "duration", time.Since(start))
	}()

	// Fetch the Deployment instance
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Object not found, log deletion event
			r.logDeploymentEvent(log, "delete", req.NamespacedName, nil)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get deployment")
		return ctrl.Result{}, err
	}

	// Determine event type and log accordingly
	eventType := r.determineEventType(deployment)
	r.logDeploymentEvent(log, eventType, req.NamespacedName, deployment)

	// Additional reconciliation logic can be added here
	// For now, this controller focuses on logging events
	
	return ctrl.Result{}, nil
}

// determineEventType determines the event type based on deployment metadata
func (r *DeploymentReconciler) determineEventType(deployment *appsv1.Deployment) string {
	age := time.Since(deployment.CreationTimestamp.Time)
	
	// If the deployment is very new (less than 5 seconds), consider it an add event
	if age < 5*time.Second && deployment.Generation == 1 {
		return "add"
	}
	
	// If generation > 1, it's an update to the spec
	if deployment.Generation > 1 {
		return "update"
	}
	
	// If observed generation is less than generation, it's a pending update
	if deployment.Status.ObservedGeneration < deployment.Generation {
		return "pending"
	}
	
	// Otherwise it's a status sync
	return "sync"
}

// logDeploymentEvent logs deployment events with structured logging
func (r *DeploymentReconciler) logDeploymentEvent(log logr.Logger, eventType string, namespacedName types.NamespacedName, deployment *appsv1.Deployment) {
	baseFields := map[string]interface{}{
		"cluster":   r.cluster,
		"event":     eventType,
		"namespace": namespacedName.Namespace,
		"name":      namespacedName.Name,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if deployment == nil {
		// Deletion event
		log.Info("Deployment deleted", baseFields)
		return
	}

	// Add deployment-specific fields
	fields := baseFields
	fields["generation"] = deployment.Generation
	fields["observed_generation"] = deployment.Status.ObservedGeneration
	fields["resource_version"] = deployment.ResourceVersion
	fields["uid"] = deployment.UID
	fields["created"] = deployment.CreationTimestamp.Format(time.RFC3339)
	
	// Add spec fields
	fields["replicas"] = r.getReplicasValue(deployment.Spec.Replicas)
	fields["selector"] = deployment.Spec.Selector.MatchLabels
	fields["strategy"] = deployment.Spec.Strategy.Type
	
	// Add container information
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		container := deployment.Spec.Template.Spec.Containers[0]
		fields["image"] = container.Image
		fields["container_name"] = container.Name
		
		// Add resource requests/limits if present
		if container.Resources.Requests != nil {
			fields["cpu_request"] = container.Resources.Requests.Cpu().String()
			fields["memory_request"] = container.Resources.Requests.Memory().String()
		}
		if container.Resources.Limits != nil {
			fields["cpu_limit"] = container.Resources.Limits.Cpu().String()
			fields["memory_limit"] = container.Resources.Limits.Memory().String()
		}
	}

	// Add status fields
	fields["ready_replicas"] = deployment.Status.ReadyReplicas
	fields["available_replicas"] = deployment.Status.AvailableReplicas
	fields["updated_replicas"] = deployment.Status.UpdatedReplicas
	fields["unavailable_replicas"] = deployment.Status.UnavailableReplicas
	
	// Add labels and annotations (limited to avoid log spam)
	if len(deployment.Labels) > 0 {
		fields["labels"] = r.limitMapSize(deployment.Labels, 5)
	}
	if len(deployment.Annotations) > 0 {
		fields["annotations"] = r.limitMapSize(deployment.Annotations, 3)
	}

	// Add conditions
	if len(deployment.Status.Conditions) > 0 {
		var conditions []map[string]interface{}
		for _, condition := range deployment.Status.Conditions {
			conditions = append(conditions, map[string]interface{}{
				"type":   condition.Type,
				"status": condition.Status,
				"reason": condition.Reason,
			})
		}
		fields["conditions"] = conditions
	}

	// Log with appropriate level based on event type
	switch eventType {
	case "add":
		log.Info("Deployment created", fields)
	case "update":
		log.Info("Deployment updated", fields)
	case "delete":
		log.Info("Deployment deleted", fields)
	case "pending":
		log.Info("Deployment update pending", fields)
	default:
		log.V(1).Info("Deployment status sync", fields)
	}
}

// getReplicasValue safely gets the replicas value
func (r *DeploymentReconciler) getReplicasValue(replicas *int32) int32 {
	if replicas == nil {
		return 1 // Default replica count
	}
	return *replicas
}

// limitMapSize limits the size of a map to avoid log spam
func (r *DeploymentReconciler) limitMapSize(m map[string]string, maxSize int) map[string]string {
	if len(m) <= maxSize {
		return m
	}
	
	result := make(map[string]string, maxSize)
	count := 0
	for k, v := range m {
		if count >= maxSize {
			result["..."] = fmt.Sprintf("and %d more", len(m)-maxSize)
			break
		}
		result[k] = v
		count++
	}
	return result
}

// AddToManager adds the deployment controller to the manager
func AddToManager(mgr manager.Manager, cluster, namespace string, concurrency int) error {
	if concurrency <= 0 {
		concurrency = 1
	}
	
	reconciler := NewDeploymentReconciler(mgr, cluster, namespace, concurrency)
	return reconciler.SetupWithManager(mgr)
}
