package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestNew(t *testing.T) {
	port := 8080
	server := New(port)
	
	if server == nil {
		t.Fatal("New() returned nil")
		return // This will never be reached, but helps static analysis
	}
	
	if server.port != port {
		t.Errorf("Expected port %d, got %d", port, server.port)
	}
}

func TestServerEndpoints(t *testing.T) {
	// Find available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	
	server := New(port)
	
	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for server to start
	time.Sleep(200 * time.Millisecond)
	
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	
	// Test health endpoint
	statusCode, body, err := fasthttp.Get(nil, baseURL+"/health")
	if err != nil {
		t.Fatalf("Failed to get health endpoint: %v", err)
	}
	
	if statusCode != fasthttp.StatusOK {
		t.Errorf("Health endpoint: expected status %d, got %d", fasthttp.StatusOK, statusCode)
	}
	
	expectedHealth := `{"status":"ok"}`
	if string(body) != expectedHealth {
		t.Errorf("Health endpoint: expected body %s, got %s", expectedHealth, string(body))
	}
	
	// Test version endpoint
	statusCode, body, err = fasthttp.Get(nil, baseURL+"/version")
	if err != nil {
		t.Fatalf("Failed to get version endpoint: %v", err)
	}
	
	if statusCode != fasthttp.StatusOK {
		t.Errorf("Version endpoint: expected status %d, got %d", fasthttp.StatusOK, statusCode)
	}
	
	expectedVersion := `{"version":"v0.9.1"}`
	if string(body) != expectedVersion {
		t.Errorf("Version endpoint: expected body %s, got %s", expectedVersion, string(body))
	}
	
	// Test 404 endpoint
	statusCode, body, err = fasthttp.Get(nil, baseURL+"/nonexistent")
	if err != nil {
		t.Fatalf("Failed to get nonexistent endpoint: %v", err)
	}
	
	if statusCode != fasthttp.StatusNotFound {
		t.Errorf("Nonexistent endpoint: expected status %d, got %d", fasthttp.StatusNotFound, statusCode)
	}
	
	expectedNotFound := `{"error":"not found"}`
	if string(body) != expectedNotFound {
		t.Errorf("Nonexistent endpoint: expected body %s, got %s", expectedNotFound, string(body))
	}
}
