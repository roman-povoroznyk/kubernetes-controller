package server

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/kubernetes"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/valyala/fasthttp"
)

// Server represents the HTTP server
type Server struct {
	port              int
	deploymentHandler *DeploymentHandler
}

// New creates a new server instance
func New(port int) *Server {
	return &Server{
		port: port,
	}
}

// SetDeploymentInformer sets the deployment informer for API endpoints
func (s *Server) SetDeploymentInformer(informer *kubernetes.DeploymentInformer) {
	s.deploymentHandler = NewDeploymentHandler(informer)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger.Info("Starting HTTP server", map[string]interface{}{
		"port": s.port,
	})

	// Create request handler with logging middleware
	requestHandler := s.loggingMiddleware(func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		
		switch {
		case path == "/health":
			s.handleHealth(ctx)
		case path == "/version":
			s.handleVersion(ctx)
		case strings.HasPrefix(path, "/api/v1/deployments"):
			if s.deploymentHandler != nil {
				s.deploymentHandler.HandleDeployments(ctx)
			} else {
				s.handleServiceUnavailable(ctx, "Deployment informer not configured")
			}
		default:
			s.handleNotFound(ctx)
		}
	})

	// Start server
	addr := ":" + strconv.Itoa(s.port)
	logger.Info("Server listening", map[string]interface{}{
		"address": addr,
	})
	
	return fasthttp.ListenAndServe(addr, requestHandler)
}

// handleHealth handles health check endpoint
func (s *Server) handleHealth(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"status":"ok"}`)
}

// handleVersion handles version endpoint
func (s *Server) handleVersion(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, `{"version":"v0.9.1"}`)
}

// handleNotFound handles 404 responses
func (s *Server) handleNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"error":"not found"}`)
}

// handleServiceUnavailable handles 503 responses
func (s *Server) handleServiceUnavailable(ctx *fasthttp.RequestCtx, message string) {
	ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"error":"service unavailable","message":"%s"}`, message)
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		
		// Call the next handler
		next(ctx)
		
		// Log the request
		duration := time.Since(start)
		logger.Info("HTTP request", map[string]interface{}{
			"method":     string(ctx.Method()),
			"path":       string(ctx.Path()),
			"status":     ctx.Response.StatusCode(),
			"duration":   duration.String(),
			"user_agent": string(ctx.UserAgent()),
			"remote_ip":  ctx.RemoteIP().String(),
		})
	}
}
