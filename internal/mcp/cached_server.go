package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/internal/cache"
	"nix-ai-help/internal/performance"
	"nix-ai-help/pkg/logger"
)

// CachedMCPServer wraps the existing MCP server with enhanced caching capabilities
type CachedMCPServer struct {
	server  *Server
	cache   *cache.Manager
	monitor *performance.Monitor
	logger  *logger.Logger

	// Performance tracking
	docQueryCount int64
	cacheHits     int64
	cacheMisses   int64
	mutex         sync.RWMutex
}

// NewCachedMCPServer creates a new MCP server with enhanced caching
func NewCachedMCPServer(server *Server, cacheConfig *cache.CacheConfig, log *logger.Logger) (*CachedMCPServer, error) {
	if log == nil {
		log = logger.NewLogger()
	}

	// Initialize cache manager
	cacheManager, err := cache.NewManager(cacheConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// Initialize performance monitor
	monitor := performance.NewMonitor(log)

	return &CachedMCPServer{
		server:  server,
		cache:   cacheManager,
		monitor: monitor,
		logger:  log,
	}, nil
}

// CachedDocQuery performs a documentation query with enhanced caching
func (cms *CachedMCPServer) CachedDocQuery(ctx context.Context, query string, sources ...string) (string, error) {
	// Start performance monitoring
	finishTimer := cms.monitor.StartTimer(performance.MetricDocumentationQuery, "mcp_doc_query", map[string]string{
		"query":   query,
		"sources": fmt.Sprintf("%v", sources),
	})

	// Create cache key
	cacheKey := cms.createDocCacheKey(query, sources)

	// Try to get cached response first
	if cms.cache != nil {
		if cachedResponse, found := cms.cache.GetDocumentation(ctx, cacheKey); found {
			cms.logger.Debug(fmt.Sprintf("Documentation cache hit for query: %s", query))

			// Record cache hit
			cms.recordCacheHit()
			cms.monitor.RecordMetric(performance.Metric{
				Type: performance.MetricCacheHit,
				Name: "mcp_doc_query",
				Tags: map[string]string{
					"query": query,
					"type":  "documentation",
				},
				Success: true,
			})

			finishTimer(true, nil)
			return string(cachedResponse), nil
		}

		// Record cache miss
		cms.recordCacheMiss()
		cms.monitor.RecordMetric(performance.Metric{
			Type: performance.MetricCacheMiss,
			Name: "mcp_doc_query",
			Tags: map[string]string{
				"query": query,
				"type":  "documentation",
			},
			Success: true,
		})
	}

	// Cache miss - query the documentation sources
	result := cms.server.mcpServer.handleDocQuery(query, sources...)

	// Cache the successful response
	if cms.cache != nil && result != "" && !containsError(result) {
		if err := cms.cache.SetDocumentation(ctx, cacheKey, []byte(result)); err != nil {
			cms.logger.Debug(fmt.Sprintf("Failed to cache documentation response: %v", err))
		} else {
			cms.logger.Debug(fmt.Sprintf("Cached documentation response for query: %s", query))
		}
	}

	cms.recordDocQuery()
	finishTimer(true, nil)
	return result, nil
}

// ParallelDocQuery performs multiple documentation queries in parallel
func (cms *CachedMCPServer) ParallelDocQuery(ctx context.Context, queries []struct {
	Query   string
	Sources []string
}) []struct {
	Query    string
	Response string
	Error    error
	Duration time.Duration
} {
	results := make([]struct {
		Query    string
		Response string
		Error    error
		Duration time.Duration
	}, len(queries))

	var wg sync.WaitGroup

	// Limit concurrent operations
	semaphore := make(chan struct{}, 3) // Max 3 concurrent doc queries

	for i, query := range queries {
		wg.Add(1)
		go func(index int, q struct {
			Query   string
			Sources []string
		}) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			start := time.Now()
			response, err := cms.CachedDocQuery(ctx, q.Query, q.Sources...)
			duration := time.Since(start)

			results[index] = struct {
				Query    string
				Response string
				Error    error
				Duration time.Duration
			}{
				Query:    q.Query,
				Response: response,
				Error:    err,
				Duration: duration,
			}
		}(i, query)
	}

	wg.Wait()

	// Record parallel operation metric
	cms.monitor.RecordMetric(performance.Metric{
		Type: performance.MetricParallelOperation,
		Name: "parallel_doc_query",
		Tags: map[string]string{
			"count": fmt.Sprintf("%d", len(queries)),
		},
		Success: true,
	})

	return results
}

// BatchPrewarmDocumentation preloads common documentation queries
func (cms *CachedMCPServer) BatchPrewarmDocumentation(ctx context.Context, commonQueries []string) {
	if cms.cache == nil {
		cms.logger.Debug("Cache not available, skipping documentation prewarm")
		return
	}

	go func() {
		cms.logger.Info(fmt.Sprintf("Prewarming documentation cache with %d common queries", len(commonQueries)))

		for _, query := range commonQueries {
			// Perform the query to cache it
			_, err := cms.CachedDocQuery(ctx, query)
			if err != nil {
				cms.logger.Debug(fmt.Sprintf("Failed to prewarm documentation query '%s': %v", query, err))
			} else {
				cms.logger.Debug(fmt.Sprintf("Prewarmed documentation query: %s", query))
			}

			// Small delay between queries
			time.Sleep(50 * time.Millisecond)
		}

		cms.logger.Info("Documentation cache prewarming completed")
	}()
}

// GetPerformanceStats returns performance statistics for the MCP server
func (cms *CachedMCPServer) GetPerformanceStats() MCPPerformanceStats {
	cms.mutex.RLock()
	defer cms.mutex.RUnlock()

	summary := cms.monitor.GetSummary()

	var hitRate float64
	if cms.docQueryCount > 0 {
		hitRate = float64(cms.cacheHits) / float64(cms.docQueryCount) * 100
	}

	return MCPPerformanceStats{
		TotalQueries:    cms.docQueryCount,
		CacheHits:       cms.cacheHits,
		CacheMisses:     cms.cacheMisses,
		CacheHitRate:    hitRate,
		AverageResponse: summary.AverageDuration,
		Summary:         summary,
	}
}

// MCPPerformanceStats represents MCP server performance metrics
type MCPPerformanceStats struct {
	TotalQueries    int64
	CacheHits       int64
	CacheMisses     int64
	CacheHitRate    float64
	AverageResponse time.Duration
	Summary         performance.MetricsSummary
}

// FormatPerformanceReport returns a human-readable performance report
func (cms *CachedMCPServer) FormatPerformanceReport() string {
	stats := cms.GetPerformanceStats()

	result := fmt.Sprintf("📊 MCP Server Performance Report\n")
	result += fmt.Sprintf("=================================\n")
	result += fmt.Sprintf("Total Documentation Queries: %d\n", stats.TotalQueries)
	result += fmt.Sprintf("Cache Hit Rate: %.1f%%\n", stats.CacheHitRate)
	result += fmt.Sprintf("Average Response Time: %v\n", stats.AverageResponse)
	result += fmt.Sprintf("Cache Hits: %d\n", stats.CacheHits)
	result += fmt.Sprintf("Cache Misses: %d\n", stats.CacheMisses)

	result += fmt.Sprintf("\n%s", cms.monitor.FormatSummary())

	return result
}

// ClearCache clears all cached documentation
func (cms *CachedMCPServer) ClearCache(ctx context.Context) error {
	if cms.cache == nil {
		return fmt.Errorf("cache not available")
	}

	return cms.cache.Clear(ctx)
}

// Close gracefully shuts down the cached MCP server
func (cms *CachedMCPServer) Close() error {
	if cms.cache != nil {
		return cms.cache.Close()
	}
	return nil
}

// Helper methods

func (cms *CachedMCPServer) createDocCacheKey(query string, sources []string) string {
	if len(sources) == 0 {
		return fmt.Sprintf("doc:%s", query)
	}
	return fmt.Sprintf("doc:%s:sources:%v", query, sources)
}

func (cms *CachedMCPServer) recordDocQuery() {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()
	cms.docQueryCount++
}

func (cms *CachedMCPServer) recordCacheHit() {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()
	cms.cacheHits++
}

func (cms *CachedMCPServer) recordCacheMiss() {
	cms.mutex.Lock()
	defer cms.mutex.Unlock()
	cms.cacheMisses++
}

func containsError(response string) bool {
	errorIndicators := []string{
		"Error",
		"error",
		"Failed",
		"failed",
		"No documentation found",
		"not found",
		"404",
	}

	for _, indicator := range errorIndicators {
		if len(response) > 0 && len(response) < 10000 { // Avoid scanning very large responses
			if fmt.Sprintf("%s", response) != "" && len(response) > 0 {
				// Simple contains check for error indicators
				for i := 0; i < len(response)-len(indicator)+1; i++ {
					if i+len(indicator) <= len(response) {
						if response[i:i+len(indicator)] == indicator {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// GetCacheStats returns cache statistics
func (cms *CachedMCPServer) GetCacheStats() *cache.CombinedCacheStats {
	if cms.cache == nil {
		return nil
	}

	stats := cms.cache.Stats()
	return &stats
}
