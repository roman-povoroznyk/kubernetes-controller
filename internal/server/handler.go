package server

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

// HandleRequests is the main HTTP request handler
func HandleRequests(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())

	switch path {
	case "/health":
		if method != fasthttp.MethodGet {
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			fmt.Fprintf(ctx, "Method Not Allowed")
			return
		}
		handleHealth(ctx)
	case "/":
		handleDefault(ctx)
	default:
		handleNotFound(ctx)
	}
}

// handleHealth handles health check requests
func handleHealth(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "OK")
}

// handleDefault handles the root route
func handleDefault(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	fmt.Fprintf(ctx, "Hello from k8s-ctrl FastHTTP server!")
}

// handleNotFound handles requests to non-existent routes
func handleNotFound(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusNotFound)
	fmt.Fprintf(ctx, "404 Not Found: %s", ctx.Path())
}
