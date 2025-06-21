package server

import (
	"bytes"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
)

func TestServer_handleRoot(t *testing.T) {
	s := New(8080)
	
	var ctx fasthttp.RequestCtx
	ctx.Request.SetRequestURI("/")
	ctx.Request.Header.SetMethod("GET")
	
	s.handleRoot(&ctx)
	
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status 200, got %d", ctx.Response.StatusCode())
	}
	
	body := string(ctx.Response.Body())
	if !bytes.Contains([]byte(body), []byte("k6s API server")) {
		t.Error("Response should contain 'k6s API server'")
	}
}

func TestServer_handleHealth(t *testing.T) {
	s := New(8080)
	
	var ctx fasthttp.RequestCtx
	ctx.Request.SetRequestURI("/health")
	ctx.Request.Header.SetMethod("GET")
	
	s.handleHealth(&ctx)
	
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status 200, got %d", ctx.Response.StatusCode())
	}
	
	body := string(ctx.Response.Body())
	if !bytes.Contains([]byte(body), []byte("healthy")) {
		t.Error("Response should contain 'healthy'")
	}
}

func TestServer_handleInfo(t *testing.T) {
	s := New(8080)
	
	var ctx fasthttp.RequestCtx
	ctx.Request.SetRequestURI("/api/v1/info")
	ctx.Request.Header.SetMethod("GET")
	
	s.handleInfo(&ctx)
	
	if ctx.Response.StatusCode() != fasthttp.StatusOK {
		t.Errorf("Expected status 200, got %d", ctx.Response.StatusCode())
	}
	
	body := string(ctx.Response.Body())
	if !bytes.Contains([]byte(body), []byte("k6s")) {
		t.Error("Response should contain 'k6s'")
	}
}

func TestServer_handleNotFound(t *testing.T) {
	s := New(8080)
	
	var ctx fasthttp.RequestCtx
	ctx.Request.SetRequestURI("/nonexistent")
	ctx.Request.Header.SetMethod("GET")
	
	s.handleNotFound(&ctx)
	
	if ctx.Response.StatusCode() != fasthttp.StatusNotFound {
		t.Errorf("Expected status 404, got %d", ctx.Response.StatusCode())
	}
	
	body := string(ctx.Response.Body())
	if !bytes.Contains([]byte(body), []byte("Not found")) {
		t.Errorf("Response should contain 'Not found', got: %s", body)
	}
}

func TestServer_Integration(t *testing.T) {
	s := New(0) // Use random port
	handler := s.createHandler()
	
	// Test server with in-memory connection
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()
	
	go func() {
		if err := fasthttp.Serve(ln, handler); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	client := &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}
	
	// Test root endpoint
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	
	req.SetRequestURI("http://test/")
	
	err := client.Do(req, resp)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode())
	}
}
