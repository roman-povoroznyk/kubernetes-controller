/*
Copyright © 2025 Roman Povoroznyk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/roman-povoroznyk/k8s/pkg/informer"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var (
	serverPort         int
	deploymentInformer *informer.DeploymentInformer
	informerMutex      sync.RWMutex
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server with deployment informer",
	Long:  `Start a high-performance HTTP server with deployment informer integration.`,
	Run: func(cmd *cobra.Command, args []string) {
		startHTTPServer()
	},
}

func startHTTPServer() {
	log.Info().
		Int("port", serverPort).
		Str("namespace", namespace).
		Bool("all_namespaces", allNamespaces).
		Msg("Starting HTTP server with deployment informer")

	// Create Kubernetes client
	clientset, err := createKubernetesClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Determine target namespace for informer
	targetNamespace := namespace
	if allNamespaces {
		targetNamespace = ""
	}

	// Create deployment informer
	informerMutex.Lock()
	deploymentInformer = informer.NewDeploymentInformer(clientset, targetNamespace)
	informerMutex.Unlock()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start informer in background
	go func() {
		log.Info().Msg("Starting deployment informer")
		if err := deploymentInformer.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Deployment informer failed")
		}
	}()

	// Wait a bit for informer to initialize
	time.Sleep(2 * time.Second)

	// Start HTTP server in background
	go func() {
		handler := createHTTPHandler()
		addr := fmt.Sprintf(":%d", serverPort)
		
		log.Info().
			Int("port", serverPort).
			Str("addr", addr).
			Msg("Starting FastHTTP server")

		if err := fasthttp.ListenAndServe(addr, handler); err != nil {
			log.Fatal().Err(err).Msg("Failed to start FastHTTP server")
		}
	}()

	// Wait for interrupt signal
	<-sigCh
	log.Info().Msg("Received interrupt signal, shutting down...")
	cancel()
}

func createHTTPHandler() fasthttp.RequestHandler {
	return loggingMiddleware(func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/health":
			handleHealth(ctx)
		case "/deployments":
			handleDeployments(ctx)
		case "/":
			handleRoot(ctx)
		default:
			handleNotFound(ctx)
		}
	})
}

// loggingMiddleware adds HTTP request/response logging
func loggingMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		
		// Log incoming request
		log.Info().
			Str("method", string(ctx.Method())).
			Str("path", string(ctx.Path())).
			Str("remote_addr", ctx.RemoteAddr().String()).
			Str("user_agent", string(ctx.UserAgent())).
			Msg("HTTP request started")

		// Execute handler
		next(ctx)

		// Log response
		duration := time.Since(start)
		log.Info().
			Str("method", string(ctx.Method())).
			Str("path", string(ctx.Path())).
			Int("status_code", ctx.Response.StatusCode()).
			Dur("duration", duration).
			Int("response_size", len(ctx.Response.Body())).
			Msg("HTTP request completed")
	}
}

func handleHealth(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"status":"healthy","version":"%s","timestamp":"%s"}`, 
		Version, time.Now().Format(time.RFC3339))
}

func handleDeployments(ctx *fasthttp.RequestCtx) {
	informerMutex.RLock()
	defer informerMutex.RUnlock()

	if deploymentInformer == nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		ctx.SetContentType("application/json")
		fmt.Fprintf(ctx, `{"error":"informer not ready"}`)
		return
	}

	deployments := deploymentInformer.GetDeployments()
	
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	
	fmt.Fprintf(ctx, `{"deployments":[`)
	for i, deployment := range deployments {
		if i > 0 {
			fmt.Fprintf(ctx, `,`)
		}
		fmt.Fprintf(ctx, `{"name":"%s","namespace":"%s","replicas":%d,"ready_replicas":%d}`,
			deployment.Name,
			deployment.Namespace,
			*deployment.Spec.Replicas,
			deployment.Status.ReadyReplicas,
		)
	}
	fmt.Fprintf(ctx, `],"count":%d}`, len(deployments))
}

func handleRoot(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("text/plain")
	fmt.Fprintf(ctx, "k8s-controller FastHTTP server v%s\nEndpoints:\n- /health\n- /deployments\n", Version)
}

func handleNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"error":"not found","path":"%s"}`, string(ctx.Path()))
}

func init() {
	rootCmd.AddCommand(serverCmd)
	
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to run the HTTP server on")
}
