package kubernetes

import (
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentChangeAnalyzer provides custom logic for analyzing deployment changes
type DeploymentChangeAnalyzer struct {
	informer *DeploymentInformer
}

// NewDeploymentChangeAnalyzer creates a new deployment change analyzer
func NewDeploymentChangeAnalyzer(informer *DeploymentInformer) *DeploymentChangeAnalyzer {
	return &DeploymentChangeAnalyzer{
		informer: informer,
	}
}

// DeploymentChange represents a specific change in a deployment
type DeploymentChange struct {
	Type        string      `json:"type"`
	Field       string      `json:"field"`
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Description string      `json:"description"`
}

// AnalyzeUpdate performs detailed analysis of deployment update using cache search
func (dca *DeploymentChangeAnalyzer) AnalyzeUpdate(oldObj, newObj *appsv1.Deployment) []DeploymentChange {
	var changes []DeploymentChange

	// Verify object exists in cache
	cachedObj, err := dca.informer.GetDeployment(newObj.Namespace, newObj.Name)
	if err != nil {
		log.Warn().
			Err(err).
			Str("namespace", newObj.Namespace).
			Str("name", newObj.Name).
			Msg("Failed to get deployment from cache")
	} else {
		log.Debug().
			Str("namespace", newObj.Namespace).
			Str("name", newObj.Name).
			Int64("cache_generation", cachedObj.Generation).
			Int64("new_generation", newObj.Generation).
			Msg("Cache lookup successful")
	}

	// Analyze replica changes
	if oldObj.Spec.Replicas != nil && newObj.Spec.Replicas != nil {
		if *oldObj.Spec.Replicas != *newObj.Spec.Replicas {
			changes = append(changes, DeploymentChange{
				Type:        "spec",
				Field:       "replicas",
				OldValue:    *oldObj.Spec.Replicas,
				NewValue:    *newObj.Spec.Replicas,
				Description: fmt.Sprintf("Replicas changed from %d to %d", *oldObj.Spec.Replicas, *newObj.Spec.Replicas),
			})
		}
	}

	// Analyze image changes
	changes = append(changes, dca.analyzeImageChanges(oldObj, newObj)...)

	// Analyze labels and annotations
	changes = append(changes, dca.analyzeLabelChanges(oldObj, newObj)...)
	changes = append(changes, dca.analyzeAnnotationChanges(oldObj, newObj)...)

	// Analyze rolling update strategy
	changes = append(changes, dca.analyzeStrategyChanges(oldObj, newObj)...)

	// Analyze resource limits and requests
	changes = append(changes, dca.analyzeResourceChanges(oldObj, newObj)...)

	return changes
}

// AnalyzeDelete performs analysis of deployment deletion with cache verification
func (dca *DeploymentChangeAnalyzer) AnalyzeDelete(obj *appsv1.Deployment) map[string]interface{} {
	analysis := make(map[string]interface{})

	// Check if still exists in cache (might be a delayed event)
	cachedObj, err := dca.informer.GetDeployment(obj.Namespace, obj.Name)
	if err == nil && cachedObj != nil {
		analysis["cache_status"] = "still_exists"
		analysis["cache_generation"] = cachedObj.Generation
		log.Warn().
			Str("namespace", obj.Namespace).
			Str("name", obj.Name).
			Msg("Delete event received but deployment still exists in cache")
	} else {
		analysis["cache_status"] = "not_found"
		log.Debug().
			Str("namespace", obj.Namespace).
			Str("name", obj.Name).
			Msg("Delete event confirmed - deployment not in cache")
	}

	// Check for related deployments (e.g., same app label)
	relatedDeployments, err := dca.findRelatedDeployments(obj)
	if err == nil {
		analysis["related_deployments"] = len(relatedDeployments)
		if len(relatedDeployments) > 0 {
			var relatedNames []string
			for _, dep := range relatedDeployments {
				relatedNames = append(relatedNames, dep.Name)
			}
			analysis["related_names"] = relatedNames
		}
	}

	// Analyze deletion impact
	analysis["had_replicas"] = obj.Spec.Replicas != nil && *obj.Spec.Replicas > 0
	analysis["namespace"] = obj.Namespace
	analysis["labels"] = obj.Labels
	analysis["creation_timestamp"] = obj.CreationTimestamp
	analysis["generation"] = obj.Generation

	return analysis
}

// analyzeImageChanges compares container images between old and new deployments
func (dca *DeploymentChangeAnalyzer) analyzeImageChanges(oldObj, newObj *appsv1.Deployment) []DeploymentChange {
	var changes []DeploymentChange

	oldContainers := oldObj.Spec.Template.Spec.Containers
	newContainers := newObj.Spec.Template.Spec.Containers

	// Compare container images
	for i, newContainer := range newContainers {
		if i < len(oldContainers) {
			oldImage := oldContainers[i].Image
			newImage := newContainer.Image
			if oldImage != newImage {
				changes = append(changes, DeploymentChange{
					Type:        "spec",
					Field:       fmt.Sprintf("containers[%d].image", i),
					OldValue:    oldImage,
					NewValue:    newImage,
					Description: fmt.Sprintf("Container %s image changed from %s to %s", newContainer.Name, oldImage, newImage),
				})
			}
		}
	}

	return changes
}

// analyzeLabelChanges compares labels between old and new deployments
func (dca *DeploymentChangeAnalyzer) analyzeLabelChanges(oldObj, newObj *appsv1.Deployment) []DeploymentChange {
	var changes []DeploymentChange

	if !reflect.DeepEqual(oldObj.Labels, newObj.Labels) {
		changes = append(changes, DeploymentChange{
			Type:        "metadata",
			Field:       "labels",
			OldValue:    oldObj.Labels,
			NewValue:    newObj.Labels,
			Description: "Labels changed",
		})
	}

	return changes
}

// analyzeAnnotationChanges compares annotations between old and new deployments
func (dca *DeploymentChangeAnalyzer) analyzeAnnotationChanges(oldObj, newObj *appsv1.Deployment) []DeploymentChange {
	var changes []DeploymentChange

	if !reflect.DeepEqual(oldObj.Annotations, newObj.Annotations) {
		changes = append(changes, DeploymentChange{
			Type:        "metadata",
			Field:       "annotations",
			OldValue:    oldObj.Annotations,
			NewValue:    newObj.Annotations,
			Description: "Annotations changed",
		})
	}

	return changes
}

// analyzeStrategyChanges compares deployment strategy between old and new deployments
func (dca *DeploymentChangeAnalyzer) analyzeStrategyChanges(oldObj, newObj *appsv1.Deployment) []DeploymentChange {
	var changes []DeploymentChange

	if !reflect.DeepEqual(oldObj.Spec.Strategy, newObj.Spec.Strategy) {
		changes = append(changes, DeploymentChange{
			Type:        "spec",
			Field:       "strategy",
			OldValue:    oldObj.Spec.Strategy,
			NewValue:    newObj.Spec.Strategy,
			Description: "Deployment strategy changed",
		})
	}

	return changes
}

// analyzeResourceChanges compares container resources between old and new deployments
func (dca *DeploymentChangeAnalyzer) analyzeResourceChanges(oldObj, newObj *appsv1.Deployment) []DeploymentChange {
	var changes []DeploymentChange

	oldContainers := oldObj.Spec.Template.Spec.Containers
	newContainers := newObj.Spec.Template.Spec.Containers

	for i, newContainer := range newContainers {
		if i < len(oldContainers) {
			oldResources := oldContainers[i].Resources
			newResources := newContainer.Resources

			if !reflect.DeepEqual(oldResources, newResources) {
				changes = append(changes, DeploymentChange{
					Type:        "spec",
					Field:       fmt.Sprintf("containers[%d].resources", i),
					OldValue:    oldResources,
					NewValue:    newResources,
					Description: fmt.Sprintf("Container %s resources changed", newContainer.Name),
				})
			}
		}
	}

	return changes
}

// findRelatedDeployments searches cache for deployments with similar labels
func (dca *DeploymentChangeAnalyzer) findRelatedDeployments(obj *appsv1.Deployment) ([]*appsv1.Deployment, error) {
	allDeployments, err := dca.informer.ListDeployments()
	if err != nil {
		return nil, err
	}

	var related []*appsv1.Deployment
	
	// Look for deployments with same app label
	if appLabel, exists := obj.Labels["app"]; exists {
		for _, dep := range allDeployments {
			if dep.Namespace == obj.Namespace && dep.Name != obj.Name {
				if depAppLabel, depExists := dep.Labels["app"]; depExists && depAppLabel == appLabel {
					related = append(related, dep)
				}
			}
		}
	}

	return related, nil
}

// CustomLogicEventHandler implements DeploymentEventHandler with custom logic
type CustomLogicEventHandler struct {
	analyzer *DeploymentChangeAnalyzer
}

// NewCustomLogicEventHandler creates a new custom logic event handler
func NewCustomLogicEventHandler(informer *DeploymentInformer) *CustomLogicEventHandler {
	return &CustomLogicEventHandler{
		analyzer: NewDeploymentChangeAnalyzer(informer),
	}
}

func (h *CustomLogicEventHandler) OnAdd(obj *appsv1.Deployment) {
	log.Info().
		Str("namespace", obj.Namespace).
		Str("name", obj.Name).
		Int32("replicas", *obj.Spec.Replicas).
		Str("handler", "custom_logic").
		Msg("Deployment added with custom analysis")

	// Could add custom logic for ADD events here
	// For example, validate deployment configuration, check for conflicts, etc.
}

func (h *CustomLogicEventHandler) OnUpdate(oldObj, newObj *appsv1.Deployment) {
	changes := h.analyzer.AnalyzeUpdate(oldObj, newObj)
	
	logEvent := log.Info().
		Str("namespace", newObj.Namespace).
		Str("name", newObj.Name).
		Str("handler", "custom_logic").
		Int("changes_count", len(changes))

	// Add details about specific changes
	var changeTypes []string
	var changeFields []string
	for _, change := range changes {
		changeTypes = append(changeTypes, change.Type)
		changeFields = append(changeFields, change.Field)
		
		// Log detailed change information
		log.Debug().
			Str("namespace", newObj.Namespace).
			Str("name", newObj.Name).
			Str("change_type", change.Type).
			Str("field", change.Field).
			Interface("old_value", change.OldValue).
			Interface("new_value", change.NewValue).
			Str("description", change.Description).
			Msg("Deployment change detail")
	}

	if len(changeTypes) > 0 {
		logEvent = logEvent.
			Strs("change_types", changeTypes).
			Strs("change_fields", changeFields)
	}

	logEvent.Msg("Deployment updated with custom analysis")
}

func (h *CustomLogicEventHandler) OnDelete(obj *appsv1.Deployment) {
	analysis := h.analyzer.AnalyzeDelete(obj)
	
	logEvent := log.Info().
		Str("namespace", obj.Namespace).
		Str("name", obj.Name).
		Str("handler", "custom_logic")

	// Add analysis details to log
	for key, value := range analysis {
		switch v := value.(type) {
		case string:
			logEvent = logEvent.Str(key, v)
		case int:
			logEvent = logEvent.Int(key, v)
		case int64:
			logEvent = logEvent.Int64(key, v)
		case bool:
			logEvent = logEvent.Bool(key, v)
		case []string:
			logEvent = logEvent.Strs(key, v)
		default:
			logEvent = logEvent.Interface(key, v)
		}
	}

	logEvent.Msg("Deployment deleted with custom analysis")
}
