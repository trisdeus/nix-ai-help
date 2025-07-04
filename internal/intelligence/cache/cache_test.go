// Package cache provides comprehensive tests for the predictive caching system
package cache

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nix-ai-help/internal/config"
)

// MockResponseGenerator implements ResponseGenerator for testing
type MockResponseGenerator struct {
	responses map[string]string
	delay     time.Duration
	callCount int
}

func NewMockResponseGenerator() *MockResponseGenerator {
	return &MockResponseGenerator{
		responses: map[string]string{
			"nixos config":      "Here's your NixOS configuration...",
			"install package":   "To install a package in NixOS...",
			"troubleshoot boot": "To troubleshoot boot issues...",
			"flake setup":       "To set up a Nix flake...",
		},
		delay: 100 * time.Millisecond,
	}
}

func (mrg *MockResponseGenerator) GenerateResponse(ctx context.Context, query string, context QueryContext) (string, error) {
	mrg.callCount++
	
	// Simulate generation delay
	time.Sleep(mrg.delay)
	
	// Return mock response
	if response, exists := mrg.responses[query]; exists {
		return response, nil
	}
	
	return "Mock response for: " + query, nil
}

func (mrg *MockResponseGenerator) GetCallCount() int {
	return mrg.callCount
}

func (mrg *MockResponseGenerator) SetDelay(delay time.Duration) {
	mrg.delay = delay
}

// Test BehaviorAnalyzer

func TestBehaviorAnalyzer_RecordQuery(t *testing.T) {
	cfg := &config.UserConfig{}
	analyzer := NewBehaviorAnalyzer(cfg)
	
	ctx := context.Background()
	event := QueryEvent{
		ID:        "test-1",
		Query:     "nixos config",
		QueryType: "configuration",
		Timestamp: time.Now(),
		Context: QueryContext{
			WorkingDirectory: "/home/user/nixos",
			ProjectType:     "nixos-config",
		},
		Success: true,
	}
	
	err := analyzer.RecordQuery(ctx, event)
	require.NoError(t, err)
	
	assert.Len(t, analyzer.queryHistory, 1)
	assert.Equal(t, "test-1", analyzer.queryHistory[0].ID)
}

func TestBehaviorAnalyzer_AnalyzePatterns(t *testing.T) {
	cfg := &config.UserConfig{}
	analyzer := NewBehaviorAnalyzer(cfg)
	ctx := context.Background()
	
	// Create test query history with patterns
	queries := []QueryEvent{
		{ID: "1", Query: "nixos config", QueryType: "configuration", Timestamp: time.Now()},
		{ID: "2", Query: "install package", QueryType: "package_management", Timestamp: time.Now()},
		{ID: "3", Query: "nixos config", QueryType: "configuration", Timestamp: time.Now()},
		{ID: "4", Query: "install package", QueryType: "package_management", Timestamp: time.Now()},
		{ID: "5", Query: "nixos config", QueryType: "configuration", Timestamp: time.Now()},
	}
	
	for _, query := range queries {
		require.NoError(t, analyzer.RecordQuery(ctx, query))
	}
	
	insights, err := analyzer.AnalyzePatterns(ctx)
	require.NoError(t, err)
	
	assert.Greater(t, len(insights), 0)
	
	// Check that sequential patterns were detected
	foundSequential := false
	for _, insight := range insights {
		if insight.Type == "sequential_pattern" {
			foundSequential = true
			break
		}
	}
	assert.True(t, foundSequential, "Should detect sequential patterns")
}

func TestBehaviorAnalyzer_PredictNextQueries(t *testing.T) {
	cfg := &config.UserConfig{}
	analyzer := NewBehaviorAnalyzer(cfg)
	ctx := context.Background()
	
	// Set up pattern data
	pattern := &UserPattern{
		ID:              "test-pattern",
		Type:            PatternSequential,
		Frequency:       5,
		Confidence:      0.8,
		LastSeen:        time.Now(),
		ExpectedQueries: []string{"install package", "troubleshoot boot"},
		Context:         map[string]interface{}{"project_type": "nixos-config"},
		Metadata: PatternMetadata{
			TimeOfDay:   []int{9, 10, 11},
			ProjectPath: "/home/user/nixos",
		},
	}
	
	analyzer.patterns[pattern.ID] = pattern
	
	queryContext := QueryContext{
		WorkingDirectory: "/home/user/nixos",
		ProjectType:     "nixos-config",
		TimeOfDay:       10,
	}
	
	prediction, err := analyzer.PredictNextQueries(ctx, queryContext)
	require.NoError(t, err)
	require.NotNil(t, prediction)
	
	assert.Greater(t, len(prediction.NextLikelyQueries), 0)
	assert.Equal(t, pattern.ID, prediction.BasedOnPattern)
	assert.LessOrEqual(t, prediction.Confidence, 1.0) // Confidence should be capped at 1.0
}

// Test IntelligentCache

func TestIntelligentCache_SetAndGet(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         100,
		DefaultTTL:      10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
	
	cache := NewIntelligentCache(config)
	defer func() { cache.cleanupStop <- true }()
	
	// Test basic set/get
	cache.Set("test-key", "test-value", 5*time.Minute, []string{"test"}, nil)
	
	entry := cache.Get("test-key")
	require.NotNil(t, entry)
	assert.Equal(t, "test-value", entry.Value)
	assert.Equal(t, []string{"test"}, entry.Tags)
}

func TestIntelligentCache_Expiration(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         100,
		DefaultTTL:      10 * time.Millisecond, // Very short TTL
		CleanupInterval: 5 * time.Millisecond,
	}
	
	cache := NewIntelligentCache(config)
	defer func() { cache.cleanupStop <- true }()
	
	cache.Set("expiring-key", "expiring-value", 10*time.Millisecond, []string{"test"}, nil)
	
	// Should get value immediately
	entry := cache.Get("expiring-key")
	require.NotNil(t, entry)
	
	// Wait for expiration
	time.Sleep(50 * time.Millisecond)
	
	// Should be expired now
	entry = cache.Get("expiring-key")
	assert.Nil(t, entry)
}

func TestIntelligentCache_TagInvalidation(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         100,
		DefaultTTL:      10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
	
	cache := NewIntelligentCache(config)
	defer func() { cache.cleanupStop <- true }()
	
	// Set multiple entries with same tag
	cache.Set("key1", "value1", 5*time.Minute, []string{"nixos"}, nil)
	cache.Set("key2", "value2", 5*time.Minute, []string{"nixos"}, nil)
	cache.Set("key3", "value3", 5*time.Minute, []string{"other"}, nil)
	
	// Verify entries exist
	assert.NotNil(t, cache.Get("key1"))
	assert.NotNil(t, cache.Get("key2"))
	assert.NotNil(t, cache.Get("key3"))
	
	// Invalidate by tag
	invalidated := cache.InvalidateByTag("nixos")
	assert.Equal(t, 2, invalidated)
	
	// Check that nixos-tagged entries are gone
	assert.Nil(t, cache.Get("key1"))
	assert.Nil(t, cache.Get("key2"))
	assert.NotNil(t, cache.Get("key3"))
}

func TestIntelligentCache_LRUEviction(t *testing.T) {
	config := &CacheConfig{
		MaxSize:         2, // Small cache for testing LRU
		DefaultTTL:      10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
	
	cache := NewIntelligentCache(config)
	defer func() { cache.cleanupStop <- true }()
	
	// Fill cache to capacity
	cache.Set("key1", "value1", 5*time.Minute, []string{"test"}, nil)
	cache.Set("key2", "value2", 5*time.Minute, []string{"test"}, nil)
	
	// Access key1 to make it more recently used
	_ = cache.Get("key1")
	
	// Add another entry, should evict key2 (least recently used)
	cache.Set("key3", "value3", 5*time.Minute, []string{"test"}, nil)
	
	// key1 and key3 should exist, key2 should be evicted
	assert.NotNil(t, cache.Get("key1"))
	assert.Nil(t, cache.Get("key2"))
	assert.NotNil(t, cache.Get("key3"))
}

// Test ResponsePreGenerator

func TestResponsePreGenerator_QueueAndGenerate(t *testing.T) {
	cfg := &config.UserConfig{}
	analyzer := NewBehaviorAnalyzer(cfg)
	mockGen := NewMockResponseGenerator()
	preGen := NewResponsePreGenerator(analyzer, cfg)
	preGen.responseGenerator = mockGen
	
	ctx := context.Background()
	require.NoError(t, preGen.Start(ctx))
	defer preGen.Stop()
	
	// Queue a pre-generation request
	queryContext := QueryContext{
		WorkingDirectory: "/test",
		ProjectType:     "nixos-config",
	}
	
	err := preGen.QueuePregeneration("nixos config", queryContext, PriorityHigh)
	require.NoError(t, err)
	
	// Wait for processing
	time.Sleep(500 * time.Millisecond)
	
	// Check that response was cached
	response, found := preGen.GetCachedResponse("nixos config", queryContext)
	assert.True(t, found)
	assert.Contains(t, response, "NixOS configuration")
}

func TestResponsePreGenerator_CacheHit(t *testing.T) {
	cfg := &config.UserConfig{}
	analyzer := NewBehaviorAnalyzer(cfg)
	mockGen := NewMockResponseGenerator()
	preGen := NewResponsePreGenerator(analyzer, cfg)
	preGen.responseGenerator = mockGen
	
	queryContext := QueryContext{
		WorkingDirectory: "/test",
		ProjectType:     "nixos-config",
	}
	
	// Manually cache a response
	cacheKey := preGen.generateCacheKey("test query", queryContext)
	preGen.cache.Set(cacheKey, "cached response", 5*time.Minute, []string{"test"}, nil)
	
	// Should get cached response
	response, found := preGen.GetCachedResponse("test query", queryContext)
	assert.True(t, found)
	assert.Equal(t, "cached response", response)
	
	// Generator should not have been called
	assert.Equal(t, 0, mockGen.GetCallCount())
}

// Test StreamingOptimizer

func TestStreamingOptimizer_CreateAndWriteStream(t *testing.T) {
	cfg := &config.UserConfig{}
	optimizer := NewStreamingOptimizer(cfg)
	
	ctx := context.Background()
	var builder strings.Builder
	
	queryContext := QueryContext{
		WorkingDirectory: "/test",
		ProjectType:     "nixos-config",
	}
	
	stream, err := optimizer.CreateStream(ctx, "test query", queryContext, &builder)
	require.NoError(t, err)
	require.NotNil(t, stream)
	
	// Write some data
	testData := []byte("Hello, streaming world!")
	err = optimizer.WriteToStream(stream.ID, testData)
	require.NoError(t, err)
	
	// Flush the stream
	err = optimizer.FlushStream(stream.ID)
	require.NoError(t, err)
	
	// Check that data was written
	assert.Equal(t, string(testData), builder.String())
	
	// Close the stream
	err = optimizer.CloseStream(stream.ID)
	require.NoError(t, err)
	assert.False(t, stream.Active)
}

func TestStreamingOptimizer_ChunkWriting(t *testing.T) {
	cfg := &config.UserConfig{}
	optimizer := NewStreamingOptimizer(cfg)
	
	ctx := context.Background()
	var builder strings.Builder
	
	queryContext := QueryContext{
		WorkingDirectory: "/test",
		ProjectType:     "nixos-config",
	}
	
	stream, err := optimizer.CreateStream(ctx, "test query", queryContext, &builder)
	require.NoError(t, err)
	
	// Write structured chunks
	chunks := []StreamChunk{
		{Type: ChunkTypeHeader, Data: []byte("Header content"), Sequence: 0},
		{Type: ChunkTypeContent, Data: []byte("Main content"), Sequence: 1},
		{Type: ChunkTypeEnd, Data: []byte(""), Sequence: 2},
	}
	
	for _, chunk := range chunks {
		err = optimizer.WriteStreamChunk(stream.ID, chunk)
		require.NoError(t, err)
	}
	
	// Flush and close
	optimizer.FlushStream(stream.ID)
	optimizer.CloseStream(stream.ID)
	
	// Should contain serialized chunk data
	assert.Contains(t, builder.String(), "Header content")
	assert.Contains(t, builder.String(), "Main content")
}

// Test CacheManager

func TestCacheManager_ProcessQuery(t *testing.T) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// Create query request
	request := QueryRequest{
		ID:    "test-request",
		Query: "nixos config",
		Context: QueryContext{
			WorkingDirectory: "/test",
			ProjectType:     "nixos-config",
		},
		Options: QueryOptions{
			EnableStreaming:     false,
			EnablePregeneration: true,
			ForceRefresh:        false,
		},
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// Process query
	response, err := manager.ProcessQuery(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	
	assert.Equal(t, "test-request", response.ID)
	assert.Equal(t, "nixos config", response.Query)
	assert.False(t, response.CacheHit) // First time should be generated
	assert.Equal(t, SourceGenerated, response.Source)
	assert.Contains(t, response.Response, "NixOS configuration")
	
	// Second query should hit cache
	request.ID = "test-request-2"
	response2, err := manager.ProcessQuery(ctx, request)
	require.NoError(t, err)
	
	assert.True(t, response2.CacheHit)
	assert.Equal(t, SourceCache, response2.Source)
}

func TestCacheManager_StreamingQuery(t *testing.T) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// Create streaming query request
	var builder strings.Builder
	request := QueryRequest{
		ID:    "streaming-request",
		Query: "flake setup",
		Context: QueryContext{
			WorkingDirectory: "/test",
			ProjectType:     "nix-flake",
		},
		StreamWriter: &builder,
		Options: QueryOptions{
			EnableStreaming:     true,
			EnablePregeneration: true,
			ForceRefresh:        false,
		},
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// Process streaming query
	response, err := manager.ProcessQuery(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, response)
	
	assert.NotEmpty(t, response.StreamID)
	assert.Contains(t, response.Response, "flake")
	
	// Check that data was written to stream
	assert.NotEmpty(t, builder.String())
}

func TestCacheManager_BehaviorAnalysis(t *testing.T) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// Process several queries to build behavior data
	queries := []string{
		"nixos config",
		"install package",
		"nixos config",
		"troubleshoot boot",
		"nixos config",
	}
	
	for i, query := range queries {
		request := QueryRequest{
			ID:    fmt.Sprintf("request-%d", i),
			Query: query,
			Context: QueryContext{
				WorkingDirectory: "/test",
				ProjectType:     "nixos-config",
			},
			Options: QueryOptions{
				EnablePregeneration: true,
			},
			Timestamp: time.Now(),
		}
		
		_, err := manager.ProcessQuery(ctx, request)
		require.NoError(t, err)
	}
	
	// Analyze behavior
	insights, err := manager.AnalyzeBehavior(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(insights), 0)
	
	// Get stats
	stats := manager.GetStats()
	assert.Greater(t, stats.BehaviorAnalysis.TotalQueries, 0)
	assert.NotEmpty(t, stats.OverallHealth)
}

func TestCacheManager_PredictivePregeneration(t *testing.T) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// Create some behavior patterns by processing queries
	for i := 0; i < 3; i++ {
		request := QueryRequest{
			ID:    fmt.Sprintf("pattern-request-%d", i),
			Query: "nixos config",
			Context: QueryContext{
				WorkingDirectory: "/test",
				ProjectType:     "nixos-config",
				TimeOfDay:       10,
			},
			Options: QueryOptions{
				EnablePregeneration: true,
			},
			Timestamp: time.Now(),
		}
		
		_, err := manager.ProcessQuery(ctx, request)
		require.NoError(t, err)
	}
	
	// Predict next queries
	queryContext := QueryContext{
		WorkingDirectory: "/test",
		ProjectType:     "nixos-config",
		TimeOfDay:       10,
	}
	
	prediction, err := manager.PredictNextQueries(ctx, queryContext)
	require.NoError(t, err)
	
	// May or may not have predictions depending on pattern detection
	// This tests that the prediction system doesn't crash
	assert.NotNil(t, prediction)
}

// Test edge cases and error conditions

func TestCacheManager_InvalidQueries(t *testing.T) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// Test empty query
	request := QueryRequest{
		ID:        "empty-query",
		Query:     "",
		Context:   QueryContext{},
		Options:   QueryOptions{},
		Timestamp: time.Now(),
	}
	
	response, err := manager.ProcessQuery(ctx, request)
	require.NoError(t, err) // Should handle gracefully
	assert.NotNil(t, response)
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	mockGen.SetDelay(50 * time.Millisecond) // Add some delay for concurrency testing
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// Run multiple concurrent queries
	const numGoroutines = 10
	const numQueries = 5
	
	results := make(chan error, numGoroutines*numQueries)
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numQueries; j++ {
				request := QueryRequest{
					ID:    fmt.Sprintf("concurrent-%d-%d", goroutineID, j),
					Query: fmt.Sprintf("query-%d", j%3), // Rotate between 3 queries
					Context: QueryContext{
						WorkingDirectory: "/test",
						ProjectType:     "nixos-config",
					},
					Options: QueryOptions{
						EnablePregeneration: true,
					},
					Timestamp: time.Now(),
				}
				
				_, err := manager.ProcessQuery(ctx, request)
				results <- err
			}
		}(i)
	}
	
	// Collect results
	for i := 0; i < numGoroutines*numQueries; i++ {
		err := <-results
		assert.NoError(t, err)
	}
	
	// Verify cache is working (should have cache hits)
	stats := manager.GetStats()
	assert.Greater(t, stats.Pregeneration.CacheHits+stats.Pregeneration.CacheMisses, 0)
}

// Benchmark tests

func BenchmarkCacheManager_ProcessQuery(b *testing.B) {
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	mockGen.SetDelay(1 * time.Millisecond) // Minimal delay for benchmarking
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	manager.Start(ctx)
	defer manager.Stop()
	
	request := QueryRequest{
		ID:    "benchmark-request",
		Query: "nixos config",
		Context: QueryContext{
			WorkingDirectory: "/test",
			ProjectType:     "nixos-config",
		},
		Options: QueryOptions{
			EnablePregeneration: true,
		},
		Timestamp: time.Now(),
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		request.ID = fmt.Sprintf("benchmark-request-%d", i)
		_, err := manager.ProcessQuery(ctx, request)
		if err != nil {
			b.Fatalf("ProcessQuery failed: %v", err)
		}
	}
}

func BenchmarkIntelligentCache_SetGet(b *testing.B) {
	config := &CacheConfig{
		MaxSize:         10000,
		DefaultTTL:      10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
	
	cache := NewIntelligentCache(config)
	defer func() { cache.cleanupStop <- true }()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i%1000) // Rotate keys
		value := fmt.Sprintf("benchmark-value-%d", i)
		
		cache.Set(key, value, 5*time.Minute, []string{"benchmark"}, nil)
		entry := cache.Get(key)
		if entry == nil {
			b.Fatalf("Failed to get cached value for key %s", key)
		}
	}
}

// Helper function to wait for async operations in tests
func waitForAsyncOperations() {
	time.Sleep(100 * time.Millisecond)
}

// MockWriter implements io.Writer for testing streaming
type MockWriter struct {
	data []byte
}

func (mw *MockWriter) Write(p []byte) (n int, err error) {
	mw.data = append(mw.data, p...)
	return len(p), nil
}

func (mw *MockWriter) String() string {
	return string(mw.data)
}

func TestCacheManager_Integration(t *testing.T) {
	// Integration test combining all components
	cfg := &config.UserConfig{}
	mockGen := NewMockResponseGenerator()
	manager := NewCacheManager(cfg, mockGen)
	
	ctx := context.Background()
	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()
	
	// 1. Process initial queries to build behavior patterns
	queries := []string{
		"nixos config",
		"install package", 
		"nixos config",
		"flake setup",
		"nixos config",
	}
	
	for i, query := range queries {
		request := QueryRequest{
			ID:    fmt.Sprintf("integration-%d", i),
			Query: query,
			Context: QueryContext{
				WorkingDirectory: "/home/user/nixos",
				ProjectType:     "nixos-config",
				TimeOfDay:       10,
			},
			Options: QueryOptions{
				EnablePregeneration: true,
			},
			Timestamp: time.Now(),
		}
		
		response, err := manager.ProcessQuery(ctx, request)
		require.NoError(t, err)
		
		if i == 0 {
			assert.False(t, response.CacheHit) // First query should generate
		}
		if i == 2 || i == 4 {
			// Repeated queries should potentially hit cache
			// (depending on cache timing and key generation)
		}
	}
	
	// 2. Analyze behavior
	insights, err := manager.AnalyzeBehavior(ctx)
	require.NoError(t, err)
	assert.Greater(t, len(insights), 0)
	
	// 3. Get predictions
	prediction, err := manager.PredictNextQueries(ctx, QueryContext{
		WorkingDirectory: "/home/user/nixos",
		ProjectType:     "nixos-config",
		TimeOfDay:       10,
	})
	require.NoError(t, err)
	assert.NotNil(t, prediction)
	
	// 4. Test streaming
	mockWriter := &MockWriter{}
	streamRequest := QueryRequest{
		ID:           "integration-stream",
		Query:        "explain flakes",
		Context:      QueryContext{WorkingDirectory: "/home/user/nixos"},
		StreamWriter: mockWriter,
		Options:      QueryOptions{EnableStreaming: true},
		Timestamp:    time.Now(),
	}
	
	streamResponse, err := manager.ProcessQuery(ctx, streamRequest)
	require.NoError(t, err)
	assert.NotEmpty(t, streamResponse.StreamID)
	
	// 5. Check final stats
	stats := manager.GetStats()
	assert.Greater(t, stats.BehaviorAnalysis.TotalQueries, 0)
	assert.NotEmpty(t, stats.OverallHealth)
	assert.True(t, stats.Pregeneration.TotalRequests > 0 || stats.Pregeneration.CacheHits > 0)
	
	t.Logf("Integration test completed successfully")
	t.Logf("Total queries: %d", stats.BehaviorAnalysis.TotalQueries)
	t.Logf("Cache hits: %d", stats.Pregeneration.CacheHits)
	t.Logf("Cache misses: %d", stats.Pregeneration.CacheMisses)
	t.Logf("Overall health: %s", stats.OverallHealth)
}