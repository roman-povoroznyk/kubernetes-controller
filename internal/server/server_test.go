package server

import (
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestHandleHealth(t *testing.T) {
	// Create a test context
	ctx := &fasthttp.RequestCtx{}

	// Call the handler
	handleHealth(ctx)

	// Check status code
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	// Check response body
	body := string(ctx.Response.Body())
	if body != "OK" {
		t.Errorf("Expected response 'OK', got '%s'", body)
	}
}

func TestHandleDefault(t *testing.T) {
	// Create a test context
	ctx := &fasthttp.RequestCtx{}

	// Call the handler
	handleDefault(ctx)

	// Check status code
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusOK, ctx.Response.StatusCode())
	}

	// Check response body
	body := string(ctx.Response.Body())
	expected := "Hello from k8s-ctrl FastHTTP server!"
	if body != expected {
		t.Errorf("Expected '%s', got '%s'", expected, body)
	}
}

func TestHandleNotFound(t *testing.T) {
	// Create a test context
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/not-found")

	// Call the handler
	handleNotFound(ctx)

	// Check status code
	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", fasthttp.StatusNotFound, ctx.Response.StatusCode())
	}

	// Check response body contains the path
	body := strings.TrimSpace(string(ctx.Response.Body()))
	expected := `{"error":"Not Found","message":"404 Not Found: /not-found not found"}`
	if body != expected {
		t.Errorf("Expected '%s', got '%s'", expected, body)
	}
}

func TestHandleRequests(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		expected string
		status   int
	}{
		{"Health endpoint GET", "/health", "GET", "OK", 200},
		{"Health endpoint POST - not allowed", "/health", "POST", `{"error":"Method Not Allowed","message":"Method POST is not allowed for this endpoint"}`, 405},
		{"Root endpoint", "/", "GET", "Hello from k8s-ctrl FastHTTP server!", 200},
		{"Unknown endpoint", "/unknown", "GET", `{"error":"Not Found","message":"404 Not Found: /unknown not found"}`, 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod(tt.method)

			HandleRequests(ctx)

			// Check status code
			if ctx.Response.StatusCode() != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, ctx.Response.StatusCode())
			}

			// Check response body
			body := strings.TrimSpace(string(ctx.Response.Body()))
			if body != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, body)
			}
		})
	}
}
