package server

import (
	"fmt"
	"strconv"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/valyala/fasthttp"
)

// Server represents the HTTP server
type Server struct {
	port int
}

// New creates a new server instance
func New(port int) *Server {
	return &Server{
		port: port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger.Info("Starting HTTP server", map[string]interface{}{
		"port": s.port,
	})

	// Create request handler
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/health":
			s.handleHealth(ctx)
		case "/version":
			s.handleVersion(ctx)
		default:
			s.handleNotFound(ctx)
		}
	}

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
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"version":"v0.4.0"}`)
}

// handleNotFound handles 404 responses
func (s *Server) handleNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"error":"not found"}`)
}
