package cache

import (
	"context"
	"sync"
	"time"
)

// MemoryCache implements an in-memory LRU cache with TTL support
type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	order   *LRUList
	config  *CacheConfig
	stats   CacheStats
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// LRUNode represents a node in the LRU doubly-linked list
type LRUNode struct {
	key  string
	prev *LRUNode
	next *LRUNode
}

// LRUList implements a doubly-linked list for LRU tracking
type LRUList struct {
	head *LRUNode
	tail *LRUNode
	size int
}

// NewLRUList creates a new LRU list
func NewLRUList() *LRUList {
	head := &LRUNode{}
	tail := &LRUNode{}
	head.next = tail
	tail.prev = head

	return &LRUList{
		head: head,
		tail: tail,
		size: 0,
	}
}

// AddToFront adds a node to the front of the list
func (l *LRUList) AddToFront(node *LRUNode) {
	node.prev = l.head
	node.next = l.head.next
	l.head.next.prev = node
	l.head.next = node
	l.size++
}

// Remove removes a node from the list
func (l *LRUList) Remove(node *LRUNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
	l.size--
}

// RemoveTail removes and returns the tail node
func (l *LRUList) RemoveTail() *LRUNode {
	if l.size == 0 {
		return nil
	}

	last := l.tail.prev
	l.Remove(last)
	return last
}

// MoveToFront moves an existing node to the front
func (l *LRUList) MoveToFront(node *LRUNode) {
	l.Remove(node)
	l.AddToFront(node)
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config *CacheConfig) *MemoryCache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cache := &MemoryCache{
		entries: make(map[string]*CacheEntry),
		order:   NewLRUList(),
		config:  config,
		stats:   CacheStats{LastCleanup: time.Now()},
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}

	// Start background cleanup routine
	go cache.cleanupRoutine()

	return cache
}

// Get retrieves a value from the memory cache
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	entry, exists := mc.entries[key]
	if !exists {
		mc.stats.Misses++
		return nil, false
	}

	// Check if expired
	if entry.IsExpired() {
		mc.mu.RUnlock()
		mc.mu.Lock()
		delete(mc.entries, key)
		mc.mu.Unlock()
		mc.mu.RLock()
		mc.stats.Misses++
		return nil, false
	}

	// Update access tracking
	entry.UpdateAccess()

	// Move to front in LRU order (need to upgrade to write lock)
	mc.mu.RUnlock()
	mc.mu.Lock()
	if node := mc.findNode(key); node != nil {
		mc.order.MoveToFront(node)
	}
	mc.mu.Unlock()
	mc.mu.RLock()

	mc.stats.Hits++
	return entry.Value, true
}

// Set stores a value in the memory cache
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	if ttl == 0 {
		ttl = mc.config.MemoryTTL
	}

	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		AccessCount: 1,
		LastAccess:  now,
		Size:        int64(len(value)),
	}

	// Check if key already exists
	if existingEntry, exists := mc.entries[key]; exists {
		// Update existing entry
		mc.stats.SizeBytes -= existingEntry.Size
		mc.entries[key] = entry
		mc.stats.SizeBytes += entry.Size

		// Move to front
		if node := mc.findNode(key); node != nil {
			mc.order.MoveToFront(node)
		}
		return nil
	}

	// Add new entry
	mc.entries[key] = entry
	mc.stats.Size++
	mc.stats.SizeBytes += entry.Size

	// Add to LRU list
	node := &LRUNode{key: key}
	mc.order.AddToFront(node)

	// Check if we need to evict
	mc.evictIfNeeded()

	return nil
}

// Delete removes a key from the memory cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if entry, exists := mc.entries[key]; exists {
		delete(mc.entries, key)
		mc.stats.Size--
		mc.stats.SizeBytes -= entry.Size

		// Remove from LRU list
		if node := mc.findNode(key); node != nil {
			mc.order.Remove(node)
		}
	}

	return nil
}

// Clear removes all entries from the memory cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.entries = make(map[string]*CacheEntry)
	mc.order = NewLRUList()
	mc.stats.Size = 0
	mc.stats.SizeBytes = 0

	return nil
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats() CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := mc.stats
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}

	return stats
}

// Close gracefully shuts down the memory cache
func (mc *MemoryCache) Close() error {
	close(mc.stopCh)
	<-mc.doneCh
	return nil
}

// findNode finds the LRU node for a given key (assumes lock is held)
func (mc *MemoryCache) findNode(key string) *LRUNode {
	current := mc.order.head.next
	for current != mc.order.tail {
		if current.key == key {
			return current
		}
		current = current.next
	}
	return nil
}

// evictIfNeeded evicts LRU entries if cache is at capacity (assumes lock is held)
func (mc *MemoryCache) evictIfNeeded() {
	for int(mc.stats.Size) > mc.config.MemoryMaxSize {
		// Remove LRU entry
		if tail := mc.order.RemoveTail(); tail != nil {
			if entry, exists := mc.entries[tail.key]; exists {
				delete(mc.entries, tail.key)
				mc.stats.Size--
				mc.stats.SizeBytes -= entry.Size
				mc.stats.Evictions++
			}
		} else {
			break // No more entries to evict
		}
	}
}

// cleanupRoutine runs periodic cleanup of expired entries
func (mc *MemoryCache) cleanupRoutine() {
	defer close(mc.doneCh)

	ticker := time.NewTicker(mc.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-mc.stopCh:
			return
		case <-ticker.C:
			mc.cleanup()
		}
	}
}

// cleanup removes expired entries (called by background routine)
func (mc *MemoryCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	keysToDelete := make([]string, 0)

	// Find expired entries
	for key, entry := range mc.entries {
		if entry.IsExpired() {
			keysToDelete = append(keysToDelete, key)
		}
	}

	// Remove expired entries
	for _, key := range keysToDelete {
		if entry := mc.entries[key]; entry != nil {
			delete(mc.entries, key)
			mc.stats.Size--
			mc.stats.SizeBytes -= entry.Size

			// Remove from LRU list
			if node := mc.findNode(key); node != nil {
				mc.order.Remove(node)
			}
		}
	}

	mc.stats.LastCleanup = now
}
