package server

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/roman-povoroznyk/k6s/pkg/api"
	"github.com/roman-povoroznyk/k6s/pkg/informer"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/roman-povoroznyk/k6s/pkg/version"
	"github.com/valyala/fasthttp"
)

// Server wraps fasthttp server
type Server struct {
	port           int
	deploymentAPI  *api.DeploymentAPI
}

// New creates a new server instance
func New(port int) *Server {
	return &Server{
		port: port,
	}
}

// SetDeploymentInformer sets the deployment informer for API endpoints
func (s *Server) SetDeploymentInformer(deploymentInformer *informer.DeploymentInformer) {
	s.deploymentAPI = api.NewDeploymentAPI(deploymentInformer)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	handler := s.createHandler()
	
	address := ":" + strconv.Itoa(s.port)
	logger.Info("Server starting", map[string]interface{}{
		"address": address,
	})
	
	return fasthttp.ListenAndServe(address, handler)
}

// createHandler creates the main request handler
func (s *Server) createHandler() fasthttp.RequestHandler {
	return s.loggingMiddleware(func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		method := string(ctx.Method())
		
		switch {
		case path == "/":
			s.handleRoot(ctx)
		case path == "/health":
			s.handleHealth(ctx)
		case path == "/api/v1/info":
			s.handleInfo(ctx)
		case path == "/api/v1/deployments" && method == "GET":
			s.handleAPIDeploymentsList(ctx)
		case path == "/api/v1/deployment" && method == "GET":
			s.handleAPIDeploymentGet(ctx)
		case path == "/api/v1/health" && method == "GET":
			s.handleAPIHealth(ctx)
		default:
			s.handleNotFound(ctx)
		}
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		method := string(ctx.Method())
		path := string(ctx.Path())
		remoteAddr := ctx.RemoteAddr().String()
		
		logger.Info("HTTP request", map[string]interface{}{
			"method":      method,
			"path":        path,
			"remote_addr": remoteAddr,
		})
		
		next(ctx)
		
		logger.Info("HTTP response", map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": ctx.Response.StatusCode(),
			"remote_addr": remoteAddr,
		})
	}
}

// handleRoot handles root endpoint
func (s *Server) handleRoot(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString(`{"message":"k6s API server","status":"running"}`)
}

// handleHealth handles health check endpoint
func (s *Server) handleHealth(ctx *fasthttp.RequestCtx) {
	logger.Debug("Health check requested", map[string]interface{}{
		"remote_addr": ctx.RemoteAddr().String(),
	})
	
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString(`{"status":"healthy"}`)
}

// handleInfo handles info endpoint
func (s *Server) handleInfo(ctx *fasthttp.RequestCtx) {
	logger.Debug("Info requested", map[string]interface{}{
		"remote_addr": ctx.RemoteAddr().String(),
	})
	
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	
	// Use version package for dynamic version info
	buildInfo := version.GetBuildInfo()
	response := `{"name":"k6s","version":"` + buildInfo["version"] + `","description":"Kubernetes Controller CLI"`
	if buildInfo["git_commit"] != "unknown" {
		response += `,"git_commit":"` + buildInfo["git_commit"] + `"`
	}
	response += `}`
	
	ctx.WriteString(response)
}

// handleNotFound handles 404 errors
func (s *Server) handleNotFound(ctx *fasthttp.RequestCtx) {
	logger.Warn("Not found", map[string]interface{}{
		"path":        string(ctx.Path()),
		"remote_addr": ctx.RemoteAddr().String(),
	})
	
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.WriteString("Not found")
}

// handleAPIDeploymentsList handles API deployment list requests
func (s *Server) handleAPIDeploymentsList(ctx *fasthttp.RequestCtx) {
	if s.deploymentAPI == nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetContentType("application/json")
		ctx.WriteString(`{"error": "Deployment API not available", "message": "Informer not configured"}`)
		return
	}
	
	ctx.SetContentType("application/json")
	
	// Check if informer cache is synced
	if !s.deploymentAPI.Informer.HasSynced() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.WriteString(`{"error": "Informer cache not synced", "message": "Cache is still synchronizing"}`)
		return
	}

	// Get deployments from cache
	deployments, err := s.deploymentAPI.Informer.ListDeployments()
	if err != nil {
		logger.Error("Failed to list deployments from cache", err, nil)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString(`{"error": "Failed to list deployments", "message": "` + err.Error() + `"}`)
		return
	}

	// Convert to API response format
	response := api.DeploymentListResponse{
		Items: make([]api.DeploymentResponse, 0, len(deployments)),
		Total: len(deployments),
	}

	for _, deployment := range deployments {
		item := s.deploymentAPI.ConvertDeploymentToResponse(deployment)
		response.Items = append(response.Items, item)
	}

	// Write JSON response
	data, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response", err, nil)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString(`{"error": "Failed to encode response"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(data)

	logger.Debug("Listed deployments from cache", map[string]interface{}{
		"count": len(deployments),
	})
}

// handleAPIDeploymentGet handles API deployment get requests
func (s *Server) handleAPIDeploymentGet(ctx *fasthttp.RequestCtx) {
	if s.deploymentAPI == nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetContentType("application/json")
		ctx.WriteString(`{"error": "Deployment API not available", "message": "Informer not configured"}`)
		return
	}
	
	ctx.SetContentType("application/json")
	
	// Extract namespace and name from query parameters
	namespace := string(ctx.QueryArgs().Peek("namespace"))
	name := string(ctx.QueryArgs().Peek("name"))

	if name == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.WriteString(`{"error": "Missing deployment name", "message": "name parameter is required"}`)
		return
	}

	if namespace == "" {
		namespace = "default"
	}

	// Check if informer cache is synced
	if !s.deploymentAPI.Informer.HasSynced() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.WriteString(`{"error": "Informer cache not synced", "message": "Cache is still synchronizing"}`)
		return
	}

	// Get deployment from cache
	deployment, err := s.deploymentAPI.Informer.GetDeployment(namespace, name)
	if err != nil {
		logger.Error("Failed to get deployment from cache", err, map[string]interface{}{
			"namespace": namespace,
			"name":      name,
		})
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.WriteString(`{"error": "Deployment not found", "message": "` + err.Error() + `"}`)
		return
	}

	// Convert to API response format
	response := s.deploymentAPI.ConvertDeploymentToResponse(deployment)

	// Write JSON response
	data, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response", err, nil)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString(`{"error": "Failed to encode response"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(data)

	logger.Debug("Got deployment from cache", map[string]interface{}{
		"namespace": namespace,
		"name":      name,
	})
}

// handleAPIHealth handles API health check requests
func (s *Server) handleAPIHealth(ctx *fasthttp.RequestCtx) {
	if s.deploymentAPI == nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetContentType("application/json")
		ctx.WriteString(`{"error": "Deployment API not available", "message": "Informer not configured"}`)
		return
	}
	
	ctx.SetContentType("application/json")

	health := map[string]interface{}{
		"status":      "ok",
		"cache_synced": s.deploymentAPI.Informer.HasSynced(),
		"timestamp":   time.Now(),
	}

	// Write JSON response
	data, err := json.Marshal(health)
	if err != nil {
		logger.Error("Failed to marshal health response", err, nil)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.WriteString(`{"error": "Failed to encode response"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(data)
}
