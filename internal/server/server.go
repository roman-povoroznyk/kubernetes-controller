package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/internal/server/middleware"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

// Config holds server configuration
type Config struct {
	Port            int
	ShutdownTimeout time.Duration
}

// Start starts the FastHTTP server with the given configuration
// and handles graceful shutdown
func Start(config Config) error {
	addr := fmt.Sprintf(":%d", config.Port)
	log.Info().Int("port", config.Port).Msg("Starting FastHTTP server")

	// Apply the logging middleware to all requests
	handler := middleware.RequestLogger(HandleRequests)

	// Create a server instance for graceful shutdown
	server := &fasthttp.Server{
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Name:         "k8s-ctrl",
	}

	// Channel to signal when the server has completely shut down
	serverClosed := make(chan struct{})

	// Run server in goroutine
	go func() {
		if err := server.ListenAndServe(addr); err != nil {
			log.Error().Err(err).Msg("Error in FastHTTP server")
		}
		close(serverClosed)
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Gracefully shutdown
	log.Info().Msg("Shutting down server...")
	shutdownTimeout := config.ShutdownTimeout
	if shutdownTimeout == 0 {
		shutdownTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("error shutting down server: %w", err)
	}

	// Wait for the server to close
	select {
	case <-serverClosed:
		log.Info().Msg("Server gracefully stopped")
	case <-ctx.Done():
		log.Warn().Msg("Server shutdown timed out")
		return fmt.Errorf("server shutdown timed out")
	}

	return nil
}
