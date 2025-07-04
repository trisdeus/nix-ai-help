// Package cache provides response pre-generation and caching capabilities
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// ResponsePreGenerator handles intelligent response pre-generation
type ResponsePreGenerator struct {
	behaviorAnalyzer  *BehaviorAnalyzer
	cache            *IntelligentCache
	pregenQueue      chan PregenRequest
	workerPool       []PregenWorker
	config           *config.UserConfig
	logger           *logger.Logger
	mu               sync.RWMutex
	stats            PregenStats
	running          bool
	responseGenerator ResponseGenerator // For testing and integration
}

// PregenRequest represents a request to pre-generate a response
type PregenRequest struct {
	ID          string                 `json:"id"`
	Query       string                 `json:"query"`
	Context     QueryContext           `json:"context"`
	Priority    PregenPriority         `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpectedUse time.Time              `json:"expected_use"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PregenPriority represents different priority levels for pre-generation
type PregenPriority int

const (
	PriorityLow PregenPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// PregenWorker handles actual response generation
type PregenWorker struct {
	ID       int
	requests chan PregenRequest
	results  chan PregenResult
	quit     chan bool
	active   bool
}

// PregenResult represents the result of pre-generation
type PregenResult struct {
	RequestID   string                 `json:"request_id"`
	Query       string                 `json:"query"`
	Response    string                 `json:"response"`
	CacheKey    string                 `json:"cache_key"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// PregenStats tracks pre-generation statistics
type PregenStats struct {
	TotalRequests     int           `json:"total_requests"`
	SuccessfulPregens int           `json:"successful_pregens"`
	FailedPregens     int           `json:"failed_pregens"`
	CacheHits         int           `json:"cache_hits"`
	CacheMisses       int           `json:"cache_misses"`
	AverageGenTime    time.Duration `json:"average_gen_time"`
	QueueLength       int           `json:"queue_length"`
	WorkerUtilization float64       `json:"worker_utilization"`
	LastUpdate        time.Time     `json:"last_update"`
}

// IntelligentCache provides smart caching with invalidation
type IntelligentCache struct {
	entries         map[string]*CacheEntry
	accessTimes     map[string]time.Time
	dependencies    map[string][]string // Key to dependent keys
	tags            map[string][]string // Tag to keys
	config          *CacheConfig
	logger          *logger.Logger
	mu              sync.RWMutex
	maxSize         int
	ttl             time.Duration
	cleanupInterval time.Duration
	cleanupStop     chan bool
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	CreatedAt   time.Time              `json:"created_at"`
	LastAccessed time.Time             `json:"last_accessed"`
	AccessCount  int                    `json:"access_count"`
	TTL         time.Duration          `json:"ttl"`
	Tags        []string               `json:"tags"`
	Dependencies []string              `json:"dependencies"`
	Metadata    map[string]interface{} `json:"metadata"`
	Size        int64                  `json:"size"`
}

// CacheConfig configures cache behavior
type CacheConfig struct {
	MaxSize         int           `json:"max_size"`         // Max number of entries
	DefaultTTL      time.Duration `json:"default_ttl"`      // Default time to live
	MaxMemoryMB     int           `json:"max_memory_mb"`    // Max memory usage in MB
	CleanupInterval time.Duration `json:"cleanup_interval"` // Cleanup frequency
	PregenEnabled   bool          `json:"pregen_enabled"`   // Enable pre-generation
	PregenWorkers   int           `json:"pregen_workers"`   // Number of pre-gen workers
	HitRatioTarget  float64       `json:"hit_ratio_target"` // Target cache hit ratio
}

// ResponseGenerator defines the interface for generating responses
type ResponseGenerator interface {
	GenerateResponse(ctx context.Context, query string, context QueryContext) (string, error)
}

// NewResponsePreGenerator creates a new response pre-generator
func NewResponsePreGenerator(analyzer *BehaviorAnalyzer, cfg *config.UserConfig) *ResponsePreGenerator {
	cacheConfig := &CacheConfig{
		MaxSize:         10000,
		DefaultTTL:      30 * time.Minute,
		MaxMemoryMB:     256,
		CleanupInterval: 5 * time.Minute,
		PregenEnabled:   true,
		PregenWorkers:   3,
		HitRatioTarget:  0.85,
	}

	cache := NewIntelligentCache(cacheConfig)
	
	rpg := &ResponsePreGenerator{
		behaviorAnalyzer: analyzer,
		cache:           cache,
		pregenQueue:     make(chan PregenRequest, 1000),
		workerPool:      make([]PregenWorker, cacheConfig.PregenWorkers),
		config:          cfg,
		logger:          logger.NewLogger(),
		stats:           PregenStats{LastUpdate: time.Now()},
	}

	return rpg
}

// NewIntelligentCache creates a new intelligent cache
func NewIntelligentCache(config *CacheConfig) *IntelligentCache {
	cache := &IntelligentCache{
		entries:         make(map[string]*CacheEntry),
		accessTimes:     make(map[string]time.Time),
		dependencies:    make(map[string][]string),
		tags:            make(map[string][]string),
		config:          config,
		logger:          logger.NewLogger(),
		maxSize:         config.MaxSize,
		ttl:             config.DefaultTTL,
		cleanupInterval: config.CleanupInterval,
		cleanupStop:     make(chan bool),
	}

	// Start cleanup routine
	go cache.cleanupRoutine()

	return cache
}

// Start initializes and starts the pre-generation system
func (rpg *ResponsePreGenerator) Start(ctx context.Context) error {
	rpg.mu.Lock()
	defer rpg.mu.Unlock()

	if rpg.running {
		return fmt.Errorf("pre-generator already running")
	}

	rpg.logger.Info("Starting response pre-generation system")

	// Start worker pool
	for i := range rpg.workerPool {
		worker := PregenWorker{
			ID:       i,
			requests: make(chan PregenRequest, 100),
			results:  make(chan PregenResult, 100),
			quit:     make(chan bool),
			active:   true,
		}
		rpg.workerPool[i] = worker
		go rpg.runWorker(ctx, &worker)
	}

	// Start request dispatcher
	go rpg.requestDispatcher(ctx)

	// Start result processor
	go rpg.resultProcessor(ctx)

	// Start predictive pre-generation
	go rpg.predictivePregeneration(ctx)

	rpg.running = true
	rpg.logger.Info(fmt.Sprintf("Started %d pre-generation workers", len(rpg.workerPool)))

	return nil
}

// Stop gracefully stops the pre-generation system
func (rpg *ResponsePreGenerator) Stop() error {
	rpg.mu.Lock()
	defer rpg.mu.Unlock()

	if !rpg.running {
		return nil
	}

	rpg.logger.Info("Stopping response pre-generation system")

	// Stop workers
	for i := range rpg.workerPool {
		rpg.workerPool[i].quit <- true
		rpg.workerPool[i].active = false
	}

	// Stop cache cleanup
	rpg.cache.cleanupStop <- true

	rpg.running = false
	rpg.logger.Info("Pre-generation system stopped")

	return nil
}

// QueuePregeneration queues a response for pre-generation
func (rpg *ResponsePreGenerator) QueuePregeneration(query string, context QueryContext, priority PregenPriority) error {
	request := PregenRequest{
		ID:          rpg.generateRequestID(query, context),
		Query:       query,
		Context:     context,
		Priority:    priority,
		CreatedAt:   time.Now(),
		ExpectedUse: time.Now().Add(5 * time.Minute), // Default: expect use in 5 minutes
		Metadata:    make(map[string]interface{}),
	}

	select {
	case rpg.pregenQueue <- request:
		rpg.stats.TotalRequests++
		rpg.stats.QueueLength = len(rpg.pregenQueue)
		rpg.logger.Info(fmt.Sprintf("Queued pre-generation for query: %s", query))
		return nil
	default:
		return fmt.Errorf("pre-generation queue is full")
	}
}

// GetCachedResponse retrieves a cached response if available
func (rpg *ResponsePreGenerator) GetCachedResponse(query string, context QueryContext) (string, bool) {
	cacheKey := rpg.generateCacheKey(query, context)
	
	if entry := rpg.cache.Get(cacheKey); entry != nil {
		if response, ok := entry.Value.(string); ok {
			rpg.stats.CacheHits++
			rpg.logger.Info(fmt.Sprintf("Cache hit for query: %s", query))
			return response, true
		}
	}

	rpg.stats.CacheMisses++
	return "", false
}

// PredictAndPregenerate analyzes behavior patterns and pre-generates likely queries
func (rpg *ResponsePreGenerator) PredictAndPregenerate(ctx context.Context, currentContext QueryContext) error {
	prediction, err := rpg.behaviorAnalyzer.PredictNextQueries(ctx, currentContext)
	if err != nil {
		return fmt.Errorf("failed to predict queries: %w", err)
	}

	rpg.logger.Info(fmt.Sprintf("Predicted %d likely queries for pre-generation", len(prediction.NextLikelyQueries)))

	for _, predicted := range prediction.NextLikelyQueries {
		if predicted.PreGenerate && predicted.Probability > 0.6 {
			priority := rpg.calculatePriority(predicted.Probability, prediction.Timeframe)
			
			err := rpg.QueuePregeneration(predicted.Query, currentContext, priority)
			if err != nil {
				rpg.logger.Error(fmt.Sprintf("Failed to queue pre-generation for %s: %v", predicted.Query, err))
			}
		}
	}

	return nil
}

// Worker methods

func (rpg *ResponsePreGenerator) runWorker(ctx context.Context, worker *PregenWorker) {
	rpg.logger.Info(fmt.Sprintf("Starting pre-generation worker %d", worker.ID))

	for {
		select {
		case <-worker.quit:
			rpg.logger.Info(fmt.Sprintf("Stopping worker %d", worker.ID))
			return
		case request := <-worker.requests:
			result := rpg.processPregenRequest(ctx, request)
			worker.results <- result
		case <-ctx.Done():
			return
		}
	}
}

func (rpg *ResponsePreGenerator) requestDispatcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case request := <-rpg.pregenQueue:
			// Find least busy worker
			selectedWorker := rpg.selectWorker()
			if selectedWorker != nil {
				selectedWorker.requests <- request
			} else {
				rpg.logger.Warn("No available workers for pre-generation request")
			}
		}
	}
}

func (rpg *ResponsePreGenerator) resultProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Process results from all workers
			for i := range rpg.workerPool {
				select {
				case result := <-rpg.workerPool[i].results:
					rpg.processPregenResult(result)
				default:
					// No result available, continue
				}
			}
			time.Sleep(100 * time.Millisecond) // Small delay to prevent busy waiting
		}
	}
}

func (rpg *ResponsePreGenerator) predictivePregeneration(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute) // Check for predictions every 2 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get current context (simplified)
			currentContext := QueryContext{
				TimeOfDay: time.Now().Hour(),
				DayOfWeek: int(time.Now().Weekday()),
			}

			err := rpg.PredictAndPregenerate(ctx, currentContext)
			if err != nil {
				rpg.logger.Error(fmt.Sprintf("Predictive pre-generation failed: %v", err))
			}
		}
	}
}

func (rpg *ResponsePreGenerator) processPregenRequest(ctx context.Context, request PregenRequest) PregenResult {
	startTime := time.Now()
	
	// Check if already cached
	cacheKey := rpg.generateCacheKey(request.Query, request.Context)
	if entry := rpg.cache.Get(cacheKey); entry != nil {
		return PregenResult{
			RequestID:   request.ID,
			Query:       request.Query,
			CacheKey:    cacheKey,
			Duration:    time.Since(startTime),
			Success:     true,
			GeneratedAt: time.Now(),
			Metadata:    map[string]interface{}{"source": "cache"},
		}
	}

	// Generate response using the configured response generator or mock
	var response string
	var err error
	
	if rpg.responseGenerator != nil {
		response, err = rpg.responseGenerator.GenerateResponse(ctx, request.Query, request.Context)
		if err != nil {
			return PregenResult{
				RequestID:   request.ID,
				Query:       request.Query,
				CacheKey:    cacheKey,
				Duration:    time.Since(startTime),
				Success:     false,
				Error:       err.Error(),
				GeneratedAt: time.Now(),
				Metadata:    map[string]interface{}{"source": "error"},
			}
		}
	} else {
		// Fallback to mock response generation
		response = rpg.generateMockResponse(request.Query, request.Context)
	}
	
	// Cache the result
	rpg.cache.Set(cacheKey, response, rpg.cache.ttl, []string{"prediction"}, nil)

	return PregenResult{
		RequestID:   request.ID,
		Query:       request.Query,
		Response:    response,
		CacheKey:    cacheKey,
		Duration:    time.Since(startTime),
		Success:     true,
		GeneratedAt: time.Now(),
		Metadata:    map[string]interface{}{"source": "generated"},
	}
}

func (rpg *ResponsePreGenerator) processPregenResult(result PregenResult) {
	if result.Success {
		rpg.stats.SuccessfulPregens++
		rpg.logger.Info(fmt.Sprintf("Pre-generated response for: %s (took %v)", result.Query, result.Duration))
	} else {
		rpg.stats.FailedPregens++
		rpg.logger.Error(fmt.Sprintf("Failed to pre-generate response for: %s - %s", result.Query, result.Error))
	}

	// Update average generation time
	if rpg.stats.SuccessfulPregens > 0 {
		totalTime := rpg.stats.AverageGenTime * time.Duration(rpg.stats.SuccessfulPregens-1)
		rpg.stats.AverageGenTime = (totalTime + result.Duration) / time.Duration(rpg.stats.SuccessfulPregens)
	}

	rpg.stats.LastUpdate = time.Now()
}

func (rpg *ResponsePreGenerator) selectWorker() *PregenWorker {
	// Simple round-robin selection (could be more sophisticated)
	for i := range rpg.workerPool {
		if rpg.workerPool[i].active && len(rpg.workerPool[i].requests) < cap(rpg.workerPool[i].requests)-1 {
			return &rpg.workerPool[i]
		}
	}
	return nil
}

func (rpg *ResponsePreGenerator) calculatePriority(probability float64, timeframe time.Duration) PregenPriority {
	if probability > 0.9 {
		return PriorityCritical
	} else if probability > 0.8 {
		return PriorityHigh
	} else if probability > 0.7 {
		return PriorityNormal
	}
	return PriorityLow
}

func (rpg *ResponsePreGenerator) generateRequestID(query string, context QueryContext) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", query, context.WorkingDirectory, time.Now().Unix())))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
}

func (rpg *ResponsePreGenerator) generateCacheKey(query string, context QueryContext) string {
	// Create a deterministic cache key based on query and relevant context
	contextStr := fmt.Sprintf("%s:%s:%s", context.WorkingDirectory, context.ProjectType, query)
	hash := sha256.Sum256([]byte(contextStr))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes
}

func (rpg *ResponsePreGenerator) generateMockResponse(query string, context QueryContext) string {
	// Mock response generation (in real implementation, this would call AI provider)
	return fmt.Sprintf("Mock response for query: %s in context: %s", query, context.WorkingDirectory)
}

// Cache methods

// Get retrieves an entry from the cache
func (ic *IntelligentCache) Get(key string) *CacheEntry {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	entry, exists := ic.entries[key]
	if !exists {
		return nil
	}

	// Check if expired
	if time.Since(entry.CreatedAt) > entry.TTL {
		// Remove expired entry (defer to cleanup routine for thread safety)
		go func() {
			ic.mu.Lock()
			defer ic.mu.Unlock()
			delete(ic.entries, key)
			delete(ic.accessTimes, key)
		}()
		return nil
	}

	// Update access metadata
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	ic.accessTimes[key] = time.Now()

	return entry
}

// Set stores an entry in the cache
func (ic *IntelligentCache) Set(key string, value interface{}, ttl time.Duration, tags []string, dependencies []string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	// Check if we need to evict entries
	if len(ic.entries) >= ic.maxSize {
		ic.evictLRU()
	}

	entry := &CacheEntry{
		Key:          key,
		Value:        value,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  0,
		TTL:         ttl,
		Tags:        tags,
		Dependencies: dependencies,
		Metadata:    make(map[string]interface{}),
		Size:        ic.estimateSize(value),
	}

	ic.entries[key] = entry
	ic.accessTimes[key] = time.Now()

	// Update tag index
	for _, tag := range tags {
		ic.tags[tag] = append(ic.tags[tag], key)
	}

	// Update dependency index
	if len(dependencies) > 0 {
		ic.dependencies[key] = dependencies
	}

	ic.logger.Info(fmt.Sprintf("Cached entry with key: %s, tags: %v", key, tags))
}

// InvalidateByTag invalidates all entries with a specific tag
func (ic *IntelligentCache) InvalidateByTag(tag string) int {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	keys, exists := ic.tags[tag]
	if !exists {
		return 0
	}

	invalidated := 0
	for _, key := range keys {
		if _, exists := ic.entries[key]; exists {
			delete(ic.entries, key)
			delete(ic.accessTimes, key)
			invalidated++
		}
	}

	delete(ic.tags, tag)
	ic.logger.Info(fmt.Sprintf("Invalidated %d entries with tag: %s", invalidated, tag))

	return invalidated
}

// InvalidateByDependency invalidates entries that depend on a specific key
func (ic *IntelligentCache) InvalidateByDependency(dependency string) int {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	invalidated := 0
	for key, deps := range ic.dependencies {
		for _, dep := range deps {
			if dep == dependency {
				if _, exists := ic.entries[key]; exists {
					delete(ic.entries, key)
					delete(ic.accessTimes, key)
					delete(ic.dependencies, key)
					invalidated++
				}
				break
			}
		}
	}

	ic.logger.Info(fmt.Sprintf("Invalidated %d entries depending on: %s", invalidated, dependency))
	return invalidated
}

// GetStats returns cache statistics
func (ic *IntelligentCache) GetStats() map[string]interface{} {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	totalSize := int64(0)
	for _, entry := range ic.entries {
		totalSize += entry.Size
	}

	return map[string]interface{}{
		"entries":     len(ic.entries),
		"max_size":    ic.maxSize,
		"total_size":  totalSize,
		"tags":        len(ic.tags),
		"dependencies": len(ic.dependencies),
		"utilization": float64(len(ic.entries)) / float64(ic.maxSize),
	}
}

func (ic *IntelligentCache) evictLRU() {
	if len(ic.entries) == 0 {
		return
	}

	// Find least recently used entry
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, accessTime := range ic.accessTimes {
		if accessTime.Before(oldestTime) {
			oldestTime = accessTime
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(ic.entries, oldestKey)
		delete(ic.accessTimes, oldestKey)
		delete(ic.dependencies, oldestKey)
		
		// Remove from tag index
		for tag, keys := range ic.tags {
			for i, key := range keys {
				if key == oldestKey {
					ic.tags[tag] = append(keys[:i], keys[i+1:]...)
					if len(ic.tags[tag]) == 0 {
						delete(ic.tags, tag)
					}
					break
				}
			}
		}
		
		ic.logger.Info(fmt.Sprintf("Evicted LRU entry: %s", oldestKey))
	}
}

func (ic *IntelligentCache) cleanupRoutine() {
	ticker := time.NewTicker(ic.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ic.cleanupStop:
			return
		case <-ticker.C:
			ic.cleanup()
		}
	}
}

func (ic *IntelligentCache) cleanup() {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	now := time.Now()
	expired := 0

	for key, entry := range ic.entries {
		if now.Sub(entry.CreatedAt) > entry.TTL {
			delete(ic.entries, key)
			delete(ic.accessTimes, key)
			delete(ic.dependencies, key)
			expired++
		}
	}

	if expired > 0 {
		ic.logger.Info(fmt.Sprintf("Cleaned up %d expired cache entries", expired))
	}
}

func (ic *IntelligentCache) estimateSize(value interface{}) int64 {
	// Simple size estimation (could be more sophisticated)
	switch v := value.(type) {
	case string:
		return int64(len(v))
	case []byte:
		return int64(len(v))
	default:
		return 1024 // Default size estimate
	}
}

// GetStats returns pre-generation statistics
func (rpg *ResponsePreGenerator) GetStats() PregenStats {
	rpg.mu.RLock()
	defer rpg.mu.RUnlock()

	stats := rpg.stats
	stats.QueueLength = len(rpg.pregenQueue)
	
	// Calculate worker utilization
	activeWorkers := 0
	for _, worker := range rpg.workerPool {
		if worker.active {
			activeWorkers++
		}
	}
	
	if len(rpg.workerPool) > 0 {
		stats.WorkerUtilization = float64(activeWorkers) / float64(len(rpg.workerPool))
	}

	return stats
}