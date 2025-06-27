package business

import (
	"fmt"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentRule represents a business rule for deployments
type DeploymentRule struct {
	Name string
	Func func(*appsv1.Deployment) error
}

// RuleEngine manages business rules
type RuleEngine struct {
	rules []DeploymentRule
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules: []DeploymentRule{
			{Name: "min-replicas", Func: minReplicasRule},
			{Name: "required-labels", Func: requiredLabelsRule},
			{Name: "resource-limits", Func: resourceLimitsRule},
		},
	}
}

// ValidateDeployment validates deployment against all rules
func (re *RuleEngine) ValidateDeployment(deployment *appsv1.Deployment) error {
	for _, rule := range re.rules {
		if err := rule.Func(deployment); err != nil {
			log.Error().Err(err).Str("rule", rule.Name).Msg("Business rule validation failed")
			return fmt.Errorf("rule %s failed: %w", rule.Name, err)
		}
	}
	log.Info().Str("deployment", deployment.Name).Msg("All business rules passed")
	return nil
}

// minReplicasRule ensures minimum replicas
func minReplicasRule(deployment *appsv1.Deployment) error {
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas < 2 {
		return fmt.Errorf("minimum replicas should be 2, got %d", *deployment.Spec.Replicas)
	}
	return nil
}

// requiredLabelsRule ensures required labels exist
func requiredLabelsRule(deployment *appsv1.Deployment) error {
	requiredLabels := []string{"app", "version", "environment"}
	for _, label := range requiredLabels {
		if _, exists := deployment.Labels[label]; !exists {
			return fmt.Errorf("required label %s is missing", label)
		}
	}
	return nil
}

// resourceLimitsRule ensures resource limits are set
func resourceLimitsRule(deployment *appsv1.Deployment) error {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Resources.Limits == nil {
			return fmt.Errorf("container %s must have resource limits", container.Name)
		}
	}
	return nil
}

import "kubernetes-controller/pkg/metrics"

// ValidateDeployment validates deployment against all rules with metrics
func (re *RuleEngine) ValidateDeploymentWithMetrics(deployment *appsv1.Deployment) error {
	for _, rule := range re.rules {
		if err := rule.Func(deployment); err != nil {
			log.Error().Err(err).Str("rule", rule.Name).Msg("Business rule validation failed")
			metrics.RecordBusinessRuleValidation(rule.Name, "failed", deployment.Namespace)
			return fmt.Errorf("rule %s failed: %w", rule.Name, err)
		}
		metrics.RecordBusinessRuleValidation(rule.Name, "passed", deployment.Namespace)
	}
	log.Info().Str("deployment", deployment.Name).Msg("All business rules passed")
	return nil
}
