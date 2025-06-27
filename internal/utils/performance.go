package utils

import (
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// HTTPCache provides simple in-memory caching for HTTP responses
type HTTPCache struct {
	cache map[string]*CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// CacheEntry represents a cached HTTP response
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewHTTPCache creates a new HTTP cache with specified TTL
func NewHTTPCache(ttl time.Duration) *HTTPCache {
	cache := &HTTPCache{
		cache: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a value from cache
func (c *HTTPCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	entry, exists := c.cache[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores a value in cache
func (c *HTTPCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.cache[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// cleanup removes expired entries
func (c *HTTPCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, entry := range c.cache {
			if now.After(entry.ExpiresAt) {
				delete(c.cache, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Clear removes all entries from cache
func (c *HTTPCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]*CacheEntry)
}

// PathExtractor efficiently extracts path segments
type PathExtractor struct{}

// ExtractResourceName extracts resource name from paths like "/deployments/nginx"
func (p *PathExtractor) ExtractResourceName(path, prefix string) (string, bool) {
	if len(path) <= len(prefix) {
		return "", false
	}
	
	if path[:len(prefix)] != prefix {
		return "", false
	}
	
	resourceName := path[len(prefix):]
	if resourceName == "" || resourceName == "/" {
		return "", false
	}
	
	// Remove leading slash if present
	if resourceName[0] == '/' {
		resourceName = resourceName[1:]
	}
	
	return resourceName, true
}

// RequestMatcher provides efficient request matching
type RequestMatcher struct{}

// IsMethodAllowed checks if the HTTP method is allowed for the endpoint
func (r *RequestMatcher) IsMethodAllowed(ctx *fasthttp.RequestCtx, allowedMethods ...string) bool {
	method := string(ctx.Method())
	for _, allowed := range allowedMethods {
		if method == allowed {
			return true
		}
	}
	return false
}

// Global instances
var (
	Cache     = NewHTTPCache(5 * time.Minute)
	PathExt   = &PathExtractor{}
	ReqMatch  = &RequestMatcher{}
)
