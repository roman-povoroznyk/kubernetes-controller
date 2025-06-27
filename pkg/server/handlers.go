package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/valyala/fasthttp"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentHandler handles deployment-related API requests
type DeploymentHandler struct {
	informer *kubernetes.DeploymentInformer
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(informer *kubernetes.DeploymentInformer) *DeploymentHandler {
	return &DeploymentHandler{
		informer: informer,
	}
}

// DeploymentResponse represents a deployment in API response
type DeploymentResponse struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Replicas  int32             `json:"replicas"`
	Ready     int32             `json:"ready"`
	Updated   int32             `json:"updated"`
	Available int32             `json:"available"`
	Age       string            `json:"age"`
	Image     string            `json:"image,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// DeploymentListResponse represents the response for deployment list
type DeploymentListResponse struct {
	Items []DeploymentResponse `json:"items"`
	Count int                  `json:"count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HandleDeployments handles deployment-related requests
func (dh *DeploymentHandler) HandleDeployments(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())
	
	logger.Debug("Handling deployment request", map[string]interface{}{
		"method": method,
		"path":   path,
	})

	switch method {
	case "GET":
		if path == "/api/v1/deployments" {
			dh.handleListDeployments(ctx)
		} else if strings.HasPrefix(path, "/api/v1/deployments/") {
			dh.handleGetDeployment(ctx)
		} else {
			dh.sendError(ctx, fasthttp.StatusNotFound, "Not found", "Invalid deployment endpoint")
		}
	default:
		dh.sendError(ctx, fasthttp.StatusMethodNotAllowed, "Method not allowed", fmt.Sprintf("Method %s is not supported", method))
	}
}

// handleListDeployments handles GET /api/v1/deployments
func (dh *DeploymentHandler) handleListDeployments(ctx *fasthttp.RequestCtx) {
	// Check if informer is ready
	if !dh.informer.IsStarted() {
		dh.sendError(ctx, fasthttp.StatusServiceUnavailable, "Service unavailable", "Deployment informer is not started")
		return
	}

	if !dh.informer.HasSynced() {
		dh.sendError(ctx, fasthttp.StatusServiceUnavailable, "Service unavailable", "Deployment informer cache is not synced")
		return
	}

	// Get deployments from cache
	deployments, err := dh.informer.ListDeployments()
	if err != nil {
		logger.Error("Failed to list deployments from cache", err, map[string]interface{}{})
		dh.sendError(ctx, fasthttp.StatusInternalServerError, "Internal server error", "Failed to retrieve deployments")
		return
	}

	// Filter by namespace if specified
	namespace := string(ctx.QueryArgs().Peek("namespace"))
	if namespace != "" {
		filteredDeployments := make([]*appsv1.Deployment, 0)
		for _, dep := range deployments {
			if dep.Namespace == namespace {
				filteredDeployments = append(filteredDeployments, dep)
			}
		}
		deployments = filteredDeployments
	}

	// Convert to response format
	response := DeploymentListResponse{
		Items: make([]DeploymentResponse, 0, len(deployments)),
		Count: len(deployments),
	}

	for _, dep := range deployments {
		response.Items = append(response.Items, dh.convertDeploymentToResponse(dep))
	}

	logger.Info("Listed deployments", map[string]interface{}{
		"count":     response.Count,
		"namespace": namespace,
	})

	dh.sendJSON(ctx, fasthttp.StatusOK, response)
}

// handleGetDeployment handles GET /api/v1/deployments/{namespace}/{name}
func (dh *DeploymentHandler) handleGetDeployment(ctx *fasthttp.RequestCtx) {
	// Parse path to extract namespace and name
	path := string(ctx.Path())
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/deployments/"), "/")
	
	var namespace, name string
	if len(parts) == 1 {
		// /api/v1/deployments/{name} - assume default namespace
		name = parts[0]
		namespace = "default"
	} else if len(parts) == 2 {
		// /api/v1/deployments/{namespace}/{name}
		namespace = parts[0]
		name = parts[1]
	} else {
		dh.sendError(ctx, fasthttp.StatusBadRequest, "Bad request", "Invalid deployment path format")
		return
	}

	// Check if informer is ready
	if !dh.informer.IsStarted() {
		dh.sendError(ctx, fasthttp.StatusServiceUnavailable, "Service unavailable", "Deployment informer is not started")
		return
	}

	if !dh.informer.HasSynced() {
		dh.sendError(ctx, fasthttp.StatusServiceUnavailable, "Service unavailable", "Deployment informer cache is not synced")
		return
	}

	// Get deployment from cache
	deployment, err := dh.informer.GetDeployment(namespace, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			dh.sendError(ctx, fasthttp.StatusNotFound, "Not found", fmt.Sprintf("Deployment %s/%s not found", namespace, name))
		} else {
			logger.Error("Failed to get deployment from cache", err, map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			})
			dh.sendError(ctx, fasthttp.StatusInternalServerError, "Internal server error", "Failed to retrieve deployment")
		}
		return
	}

	response := dh.convertDeploymentToResponse(deployment)
	
	logger.Info("Retrieved deployment", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})

	dh.sendJSON(ctx, fasthttp.StatusOK, response)
}

// convertDeploymentToResponse converts a Kubernetes deployment to API response format
func (dh *DeploymentHandler) convertDeploymentToResponse(dep *appsv1.Deployment) DeploymentResponse {
	response := DeploymentResponse{
		Name:      dep.Name,
		Namespace: dep.Namespace,
		Labels:    dep.Labels,
	}

	// Set replica counts
	if dep.Spec.Replicas != nil {
		response.Replicas = *dep.Spec.Replicas
	}
	response.Ready = dep.Status.ReadyReplicas
	response.Updated = dep.Status.UpdatedReplicas
	response.Available = dep.Status.AvailableReplicas

	// Calculate age
	if !dep.CreationTimestamp.IsZero() {
		response.Age = formatAge(dep.CreationTimestamp.Time)
	}

	// Get first container image
	if len(dep.Spec.Template.Spec.Containers) > 0 {
		response.Image = dep.Spec.Template.Spec.Containers[0].Image
	}

	return response
}

// sendJSON sends a JSON response
func (dh *DeploymentHandler) sendJSON(ctx *fasthttp.RequestCtx, statusCode int, data interface{}) {
	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal JSON response", err, map[string]interface{}{})
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, `{"error":"internal server error","message":"failed to marshal response"}`)
		return
	}
	
	if _, err := ctx.Write(jsonData); err != nil {
		// If we can't write the response, there's not much we can do
		// Log it but don't return another error as this could cause a loop
		// The status code has already been set appropriately above
		_ = err // Suppress staticcheck warning about empty branch
	}
}

// sendError sends an error response
func (dh *DeploymentHandler) sendError(ctx *fasthttp.RequestCtx, statusCode int, errType, message string) {
	response := ErrorResponse{
		Error:   errType,
		Message: message,
	}
	dh.sendJSON(ctx, statusCode, response)
}

// formatAge formats a time duration into a human-readable age string
func formatAge(t time.Time) string {
	duration := time.Since(t)
	
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}
