package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/informer"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentAPI provides HTTP API handlers for deployment operations
type DeploymentAPI struct {
	Provider DeploymentProvider
}

// NewDeploymentAPI creates a new deployment API handler with informer
func NewDeploymentAPI(informer *informer.DeploymentInformer) *DeploymentAPI {
	return &DeploymentAPI{
		Provider: informer,
	}
}

// NewDeploymentAPIWithProvider creates a new deployment API handler with custom provider
func NewDeploymentAPIWithProvider(provider DeploymentProvider) *DeploymentAPI {
	return &DeploymentAPI{
		Provider: provider,
	}
}

// DeploymentResponse represents the API response for deployment data
type DeploymentResponse struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Labels            map[string]string `json:"labels,omitempty"`
	Annotations       map[string]string `json:"annotations,omitempty"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	UpdatedReplicas   int32             `json:"updated_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	Image             string            `json:"image"`
	CreationTimestamp time.Time         `json:"creation_timestamp"`
	Status            string            `json:"status"`
}

// DeploymentListResponse represents the API response for deployment list
type DeploymentListResponse struct {
	Items []DeploymentResponse `json:"items"`
	Total int                  `json:"total"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ListDeployments handles GET /api/v1/deployments
func (da *DeploymentAPI) ListDeployments(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	if da.Provider == nil {
		da.writeError(w, http.StatusServiceUnavailable, "Provider not available", "Provider not configured")
		return
	}

	if !da.Provider.HasSynced() {
		da.writeError(w, http.StatusServiceUnavailable, "Cache not synced", "Cache is still synchronizing")
		return
	}

	// Get deployments from cache
	deployments, err := da.Provider.ListDeployments()
	if err != nil {
		logger.Error("Failed to list deployments from cache", err, nil)
		da.writeError(w, http.StatusInternalServerError, "Failed to list deployments", err.Error())
		return
	}

	// Convert to API response format
	response := DeploymentListResponse{
		Items: make([]DeploymentResponse, 0, len(deployments)),
		Total: len(deployments),
	}

	for _, deployment := range deployments {
		item := da.ConvertDeploymentToResponse(deployment)
		response.Items = append(response.Items, item)
	}

	// Write response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", err, nil)
	}

	logger.Debug("Listed deployments from cache", map[string]interface{}{
		"count": len(deployments),
	})
}

// GetDeployment handles GET /api/v1/deployments/{namespace}/{name}
func (da *DeploymentAPI) GetDeployment(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	if da.Provider == nil {
		da.writeError(w, http.StatusServiceUnavailable, "Provider not available", "Provider not configured")
		return
	}

	if !da.Provider.HasSynced() {
		da.writeError(w, http.StatusServiceUnavailable, "Cache not synced", "Cache is still synchronizing")
		return
	}

	// Extract namespace and name from URL path
	// This is a simple implementation - in production you'd use a router like gorilla/mux
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if name == "" {
		da.writeError(w, http.StatusBadRequest, "Missing deployment name", "name parameter is required")
		return
	}

	if namespace == "" {
		namespace = "default"
	}

	// Get deployment from cache
	deployment, err := da.Provider.GetDeployment(namespace, name)
	if err != nil {
		logger.Error("Failed to get deployment from cache", err, map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		})
		da.writeError(w, http.StatusNotFound, "Deployment not found", err.Error())
		return
	}

	// Convert to API response format
	response := da.ConvertDeploymentToResponse(deployment)

	// Write response
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response", err, nil)
	}

	logger.Debug("Got deployment from cache", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
}

// HealthCheck handles GET /api/v1/health
func (da *DeploymentAPI) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"cache_synced": false,
		"status":      "not ready",
		"timestamp":   time.Now(),
	}
	if da.Provider != nil && da.Provider.HasSynced() {
		resp["cache_synced"] = true
		resp["status"] = "ok"
	}
	json.NewEncoder(w).Encode(resp)
}

// ConvertDeploymentToResponse converts a Kubernetes deployment to API response format
func (da *DeploymentAPI) ConvertDeploymentToResponse(deployment *appsv1.Deployment) DeploymentResponse {
	status := "Unknown"
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		status = "Ready"
	} else if deployment.Status.ReadyReplicas > 0 {
		status = "Progressing"
	} else {
		status = "NotReady"
	}

	image := ""
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		image = deployment.Spec.Template.Spec.Containers[0].Image
	}

	return DeploymentResponse{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Labels:            deployment.Labels,
		Annotations:       deployment.Annotations,
		Replicas:          *deployment.Spec.Replicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		Image:             image,
		CreationTimestamp: deployment.CreationTimestamp.Time,
		Status:            status,
	}
}

// writeError writes an error response
func (da *DeploymentAPI) writeError(w http.ResponseWriter, code int, error string, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   error,
		Message: message,
		Code:    code,
	})
}
# API handlers
