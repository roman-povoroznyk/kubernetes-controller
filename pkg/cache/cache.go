// pkg/cache/cache.go
package cache

import (
	"sync"
	"time"
)

// Cache represents a generic cache interface
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
	Keys() []string
}

// MemoryCache implements an in-memory cache with TTL support
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	value      interface{}
	expiration int64
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*cacheItem),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	// Check if item has expired
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil, false
	}
	
	return item.value, true
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}
	
	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*cacheItem)
}

// Size returns the number of items in the cache
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}

// Keys returns all keys in the cache
func (c *MemoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	
	return keys
}

// cleanup removes expired items from the cache
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.removeExpired()
	}
}

// removeExpired removes expired items from the cache
func (c *MemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now().UnixNano()
	for key, item := range c.items {
		if item.expiration > 0 && now > item.expiration {
			delete(c.items, key)
		}
	}
}

// LRUCache implements a Least Recently Used cache
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*lruItem
	head     *lruItem
	tail     *lruItem
}

type lruItem struct {
	key   string
	value interface{}
	prev  *lruItem
	next  *lruItem
}

// NewLRUCache creates a new LRU cache with specified capacity
func NewLRUCache(capacity int) *LRUCache {
	cache := &LRUCache{
		capacity: capacity,
		items:    make(map[string]*lruItem),
	}
	
	// Initialize dummy head and tail
	cache.head = &lruItem{}
	cache.tail = &lruItem{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head
	
	return cache
}

// Get retrieves a value from the LRU cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	// Move to front
	c.moveToFront(item)
	
	return item.value, true
}

// Set stores a value in the LRU cache
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	item, exists := c.items[key]
	if exists {
		// Update existing item
		item.value = value
		c.moveToFront(item)
		return
	}
	
	// Create new item
	newItem := &lruItem{
		key:   key,
		value: value,
	}
	
	c.items[key] = newItem
	c.addToFront(newItem)
	
	// Check capacity
	if len(c.items) > c.capacity {
		c.removeLast()
	}
}

// Delete removes a value from the LRU cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	item, exists := c.items[key]
	if !exists {
		return
	}
	
	c.removeItem(item)
	delete(c.items, key)
}

// Clear removes all items from the LRU cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]*lruItem)
	c.head.next = c.tail
	c.tail.prev = c.head
}

// Size returns the number of items in the LRU cache
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return len(c.items)
}

// Keys returns all keys in the LRU cache
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keys := make([]string, 0, len(c.items))
	for k := range c.items {
		keys = append(keys, k)
	}
	
	return keys
}

// moveToFront moves an item to the front of the list
func (c *LRUCache) moveToFront(item *lruItem) {
	c.removeItem(item)
	c.addToFront(item)
}

// addToFront adds an item to the front of the list
func (c *LRUCache) addToFront(item *lruItem) {
	item.prev = c.head
	item.next = c.head.next
	c.head.next.prev = item
	c.head.next = item
}

// removeItem removes an item from the list
func (c *LRUCache) removeItem(item *lruItem) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

// removeLast removes the last item from the cache
func (c *LRUCache) removeLast() {
	lastItem := c.tail.prev
	c.removeItem(lastItem)
	delete(c.items, lastItem.key)
}
