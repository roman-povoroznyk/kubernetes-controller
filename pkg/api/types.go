package api

import (
	"time"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentResponse represents API response for deployment
type DeploymentResponse struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`
	Ready     int32             `json:"ready"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
	Status    string            `json:"status"`
}

// DeploymentListResponse represents list of deployments
type DeploymentListResponse struct {
	Items []DeploymentResponse `json:"items"`
	Total int                  `json:"total"`
}

// CreateDeploymentRequest represents request to create deployment
type CreateDeploymentRequest struct {
	Name      string            `json:"name" validate:"required"`
	Namespace string            `json:"namespace" validate:"required"`
	Image     string            `json:"image" validate:"required"`
	Replicas  int32             `json:"replicas" validate:"min=1"`
	Labels    map[string]string `json:"labels"`
	Port      int32             `json:"port,omitempty"`
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// FromK8sDeployment converts Kubernetes deployment to API response
func FromK8sDeployment(dep *appsv1.Deployment) DeploymentResponse {
	status := "Unknown"
	if dep.Status.ReadyReplicas == *dep.Spec.Replicas {
		status = "Ready"
	} else if dep.Status.ReadyReplicas > 0 {
		status = "Partially Ready"
	} else {
		status = "Not Ready"
	}

	return DeploymentResponse{
		Name:      dep.Name,
		Namespace: dep.Namespace,
		Replicas:  *dep.Spec.Replicas,
		Ready:     dep.Status.ReadyReplicas,
		Labels:    dep.Labels,
		CreatedAt: dep.CreationTimestamp.Time,
		Status:    status,
	}
}
