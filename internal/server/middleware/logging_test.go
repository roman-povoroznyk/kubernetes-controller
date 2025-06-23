package middleware

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestRequestLogger(t *testing.T) {
	// Create a mock handler that we'll wrap with our middleware
	mockHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("test response")
	}

	// Wrap the mock handler with our middleware
	handler := RequestLogger(mockHandler)

	// Create a request context for testing
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/test")
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.Header.Set("User-Agent", "go-test")

	// Call the handler
	handler(ctx)

	// Check that the X-Request-ID header was set
	requestID := string(ctx.Response.Header.Peek("X-Request-ID"))
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}

	// Check that the handler was actually called
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status 200, got %d", ctx.Response.StatusCode())
	}

	// Check the response body
	body := string(ctx.Response.Body())
	if body != "test response" {
		t.Errorf("Expected 'test response', got '%s'", body)
	}
}
