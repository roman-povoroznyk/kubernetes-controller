package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var serverPort int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server",
	Long:  `Start a high-performance HTTP server.`,
	Run: func(cmd *cobra.Command, args []string) {
		startHTTPServer()
	},
}

func startHTTPServer() {
	log.Info().Int("port", serverPort).Msg("Starting FastHTTP server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		handler := createBasicHTTPHandler()
		addr := fmt.Sprintf(":%d", serverPort)
		
		if err := fasthttp.ListenAndServe(addr, handler); err != nil {
			log.Fatal().Err(err).Msg("Failed to start FastHTTP server")
		}
	}()

	<-sigCh
	log.Info().Msg("Shutting down server...")
	cancel()
}

func createBasicHTTPHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/health":
			handleHealth(ctx)
		case "/":
			handleRoot(ctx)
		default:
			handleNotFound(ctx)
		}
	}
}

func handleHealth(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, `{"status":"healthy","version":"%s","timestamp":"%s"}`, 
		Version, time.Now().Format(time.RFC3339))
}

func handleRoot(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("text/plain")
	fmt.Fprintf(ctx, "k8s-controller FastHTTP server v%s\n", Version)
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
