// Package cache provides a comprehensive cache management system for nixai
package cache

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// CacheManager coordinates all caching subsystems
type CacheManager struct {
	behaviorAnalyzer    *BehaviorAnalyzer
	responsePregen      *ResponsePreGenerator
	streamingOptimizer  *StreamingOptimizer
	analytics          *StreamingAnalytics
	config             *config.UserConfig
	logger             *logger.Logger
	mu                 sync.RWMutex
	running            bool
	responseGenerator  ResponseGenerator
}

// CacheManagerConfig configures the cache manager
type CacheManagerConfig struct {
	EnableBehaviorAnalysis bool          `json:"enable_behavior_analysis"`
	EnablePregeneration   bool          `json:"enable_pregeneration"`
	EnableStreaming       bool          `json:"enable_streaming"`
	EnableAnalytics       bool          `json:"enable_analytics"`
	MaxCacheSize          int           `json:"max_cache_size"`
	TTL                   time.Duration `json:"ttl"`
	WorkerCount           int           `json:"worker_count"`
	PregenThreshold       float64       `json:"pregen_threshold"`
}

// CacheStats provides comprehensive cache statistics
type CacheStats struct {
	BehaviorAnalysis BehaviorStats  `json:"behavior_analysis"`
	Pregeneration   PregenStats    `json:"pregeneration"`
	Streaming       StreamStats    `json:"streaming"`
	Cache           map[string]interface{} `json:"cache"`
	OverallHealth   string         `json:"overall_health"`
	Uptime          time.Duration  `json:"uptime"`
	LastUpdate      time.Time      `json:"last_update"`
}

// BehaviorStats contains behavior analysis statistics
type BehaviorStats struct {
	TotalQueries      int                    `json:"total_queries"`
	PatternsDetected  int                    `json:"patterns_detected"`
	ActiveSessions    int                    `json:"active_sessions"`
	PredictionAccuracy float64              `json:"prediction_accuracy"`
	Insights          []BehaviorInsight      `json:"insights"`
	TopPatterns       []string               `json:"top_patterns"`
}

// QueryRequest represents a request to the cache system
type QueryRequest struct {
	ID           string                 `json:"id"`
	Query        string                 `json:"query"`
	Context      QueryContext           `json:"context"`
	StreamWriter io.Writer              `json:"-"`
	Options      QueryOptions           `json:"options"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// QueryOptions configures query processing
type QueryOptions struct {
	EnableStreaming     bool          `json:"enable_streaming"`
	EnablePregeneration bool          `json:"enable_pregeneration"`
	ForceRefresh        bool          `json:"force_refresh"`
	Priority            string        `json:"priority"`
	Timeout             time.Duration `json:"timeout"`
}

// QueryResponse represents a response from the cache system
type QueryResponse struct {
	ID           string                 `json:"id"`
	Query        string                 `json:"query"`
	Response     string                 `json:"response"`
	Source       ResponseSource         `json:"source"`
	Duration     time.Duration          `json:"duration"`
	CacheHit     bool                   `json:"cache_hit"`
	StreamID     string                 `json:"stream_id,omitempty"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

// ResponseSource indicates where the response came from
type ResponseSource string

const (
	SourceCache        ResponseSource = "cache"
	SourcePregenerated ResponseSource = "pregenerated"
	SourceGenerated    ResponseSource = "generated"
	SourceStreaming    ResponseSource = "streaming"
)

// NewCacheManager creates a new cache manager
func NewCacheManager(cfg *config.UserConfig, responseGen ResponseGenerator) *CacheManager {
	behaviorAnalyzer := NewBehaviorAnalyzer(cfg)
	responsePregen := NewResponsePreGenerator(behaviorAnalyzer, cfg)
	streamingOptimizer := NewStreamingOptimizer(cfg)
	analytics := NewStreamingAnalytics()

	return &CacheManager{
		behaviorAnalyzer:   behaviorAnalyzer,
		responsePregen:     responsePregen,
		streamingOptimizer: streamingOptimizer,
		analytics:         analytics,
		config:            cfg,
		logger:            logger.NewLogger(),
		responseGenerator: responseGen,
	}
}

// Start initializes and starts all cache subsystems
func (cm *CacheManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.running {
		return fmt.Errorf("cache manager already running")
	}

	cm.logger.Info("Starting nixai cache manager")

	// Start response pre-generator
	if err := cm.responsePregen.Start(ctx); err != nil {
		return fmt.Errorf("failed to start response pre-generator: %w", err)
	}

	cm.running = true
	cm.logger.Info("Cache manager started successfully")

	// Start background tasks
	go cm.backgroundTasks(ctx)

	return nil
}

// Stop gracefully stops all cache subsystems
func (cm *CacheManager) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.running {
		return nil
	}

	cm.logger.Info("Stopping nixai cache manager")

	// Stop response pre-generator
	if err := cm.responsePregen.Stop(); err != nil {
		cm.logger.Error(fmt.Sprintf("Error stopping pre-generator: %v", err))
	}

	cm.running = false
	cm.logger.Info("Cache manager stopped")

	return nil
}

// ProcessQuery processes a query through the cache system
func (cm *CacheManager) ProcessQuery(ctx context.Context, request QueryRequest) (*QueryResponse, error) {
	startTime := time.Now()
	
	// Record query for behavior analysis
	queryEvent := QueryEvent{
		ID:        request.ID,
		Query:     request.Query,
		QueryType: cm.inferQueryType(request.Query),
		Timestamp: request.Timestamp,
		Context:   request.Context,
		Metadata:  request.Metadata,
	}

	if err := cm.behaviorAnalyzer.RecordQuery(ctx, queryEvent); err != nil {
		cm.logger.Error(fmt.Sprintf("Failed to record query for behavior analysis: %v", err))
	}

	// Try to get cached response first
	if !request.Options.ForceRefresh {
		if cachedResponse, found := cm.responsePregen.GetCachedResponse(request.Query, request.Context); found {
			response := &QueryResponse{
				ID:          request.ID,
				Query:       request.Query,
				Response:    cachedResponse,
				Source:      SourceCache,
				Duration:    time.Since(startTime),
				CacheHit:    true,
				Confidence:  1.0,
				GeneratedAt: time.Now(),
				Metadata:    make(map[string]interface{}),
			}
			
			cm.logger.Info(fmt.Sprintf("Cache hit for query: %s", request.Query))
			return response, nil
		}
	}

	// Generate new response
	var response string
	var err error
	var streamID string

	if request.Options.EnableStreaming && request.StreamWriter != nil {
		// Use streaming generation
		stream, err := cm.streamingOptimizer.CreateStream(ctx, request.Query, request.Context, request.StreamWriter)
		if err != nil {
			return nil, fmt.Errorf("failed to create stream: %w", err)
		}

		streamID = stream.ID
		cm.analytics.RecordStreamStart(streamID)

		// Optimize stream for query type
		cm.streamingOptimizer.OptimizeStreamForQuery(stream, request.Query)

		// Generate response with streaming
		response, err = cm.generateResponseWithStreaming(ctx, request, stream)
		if err != nil {
			cm.streamingOptimizer.CloseStream(streamID)
			return nil, fmt.Errorf("failed to generate streaming response: %w", err)
		}

		cm.analytics.RecordStreamEnd(streamID)
		cm.streamingOptimizer.CloseStream(streamID)
	} else {
		// Generate response normally
		response, err = cm.responseGenerator.GenerateResponse(ctx, request.Query, request.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to generate response: %w", err)
		}
	}

	// Cache the response for future use
	cacheKey := cm.responsePregen.generateCacheKey(request.Query, request.Context)
	cm.responsePregen.cache.Set(cacheKey, response, 30*time.Minute, []string{"generated"}, nil)

	// Trigger predictive pre-generation
	go func() {
		if err := cm.responsePregen.PredictAndPregenerate(ctx, request.Context); err != nil {
			cm.logger.Error(fmt.Sprintf("Failed to trigger predictive pre-generation: %v", err))
		}
	}()

	queryResponse := &QueryResponse{
		ID:          request.ID,
		Query:       request.Query,
		Response:    response,
		Source:      SourceGenerated,
		Duration:    time.Since(startTime),
		CacheHit:    false,
		StreamID:    streamID,
		Confidence:  0.9, // Default confidence for generated responses
		GeneratedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	if streamID != "" {
		queryResponse.Source = SourceStreaming
	}

	return queryResponse, nil
}

// GetStats returns comprehensive cache statistics
func (cm *CacheManager) GetStats() CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Collect behavior analysis stats
	patterns := cm.behaviorAnalyzer.GetPatterns()
	insights := cm.behaviorAnalyzer.GetUserInsights()
	
	behaviorStats := BehaviorStats{
		TotalQueries:     len(cm.behaviorAnalyzer.queryHistory),
		PatternsDetected: len(patterns),
		Insights:         insights,
		TopPatterns:      cm.getTopPatterns(patterns),
	}

	// Get pre-generation stats
	pregenStats := cm.responsePregen.GetStats()

	// Get streaming stats
	streamStats := cm.streamingOptimizer.GetStreamStats()

	// Get cache stats
	cacheStats := cm.responsePregen.cache.GetStats()

	// Calculate overall health
	health := cm.calculateOverallHealth(pregenStats, streamStats, cacheStats)

	return CacheStats{
		BehaviorAnalysis: behaviorStats,
		Pregeneration:   pregenStats,
		Streaming:       streamStats,
		Cache:           cacheStats,
		OverallHealth:   health,
		LastUpdate:      time.Now(),
	}
}

// AnalyzeBehavior performs behavior analysis on recent queries
func (cm *CacheManager) AnalyzeBehavior(ctx context.Context) ([]BehaviorInsight, error) {
	return cm.behaviorAnalyzer.AnalyzePatterns(ctx)
}

// PredictNextQueries predicts likely next queries based on current context
func (cm *CacheManager) PredictNextQueries(ctx context.Context, currentContext QueryContext) (*PredictionResult, error) {
	return cm.behaviorAnalyzer.PredictNextQueries(ctx, currentContext)
}

// InvalidateCache invalidates cache entries by tag or pattern
func (cm *CacheManager) InvalidateCache(tag string) int {
	return cm.responsePregen.cache.InvalidateByTag(tag)
}

// SaveBehaviorPatterns saves behavior patterns to persistent storage
func (cm *CacheManager) SaveBehaviorPatterns(filePath string) error {
	return cm.behaviorAnalyzer.SavePatterns(filePath)
}

// LoadBehaviorPatterns loads behavior patterns from persistent storage
func (cm *CacheManager) LoadBehaviorPatterns(filePath string) error {
	return cm.behaviorAnalyzer.LoadPatterns(filePath)
}

// GetCachedResponse retrieves a cached response if available
func (cm *CacheManager) GetCachedResponse(query string, context QueryContext) (string, bool) {
	return cm.responsePregen.GetCachedResponse(query, context)
}

// Private methods

func (cm *CacheManager) backgroundTasks(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.performMaintenance(ctx)
		}
	}
}

func (cm *CacheManager) performMaintenance(ctx context.Context) {
	cm.logger.Info("Performing cache maintenance")

	// Analyze behavior patterns
	if _, err := cm.behaviorAnalyzer.AnalyzePatterns(ctx); err != nil {
		cm.logger.Error(fmt.Sprintf("Failed to analyze behavior patterns: %v", err))
	}

	// Clean up inactive streams
	// This would be implemented with proper stream lifecycle management

	cm.logger.Info("Cache maintenance completed")
}

func (cm *CacheManager) generateResponseWithStreaming(ctx context.Context, request QueryRequest, stream *ResponseStream) (string, error) {
	// This is a simplified streaming implementation
	// In a real implementation, this would integrate with the actual AI provider
	// to stream the response in real-time
	
	response, err := cm.responseGenerator.GenerateResponse(ctx, request.Query, request.Context)
	if err != nil {
		return "", err
	}

	// Simulate streaming by writing in chunks
	chunkSize := 256
	for i := 0; i < len(response); i += chunkSize {
		end := i + chunkSize
		if end > len(response) {
			end = len(response)
		}

		chunk := StreamChunk{
			Type:     ChunkTypeContent,
			Sequence: i / chunkSize,
			Data:     []byte(response[i:end]),
		}

		if err := cm.streamingOptimizer.WriteStreamChunk(stream.ID, chunk); err != nil {
			return "", fmt.Errorf("failed to write chunk: %w", err)
		}

		cm.analytics.RecordChunk(stream.ID, len(chunk.Data))

		// Small delay to simulate real-time streaming
		time.Sleep(10 * time.Millisecond)
	}

	// Write end chunk
	endChunk := StreamChunk{
		Type:     ChunkTypeEnd,
		Sequence: -1,
		Data:     []byte(""),
	}
	cm.streamingOptimizer.WriteStreamChunk(stream.ID, endChunk)

	return response, nil
}

func (cm *CacheManager) inferQueryType(query string) string {
	// Reuse the logic from behavior analyzer
	return cm.behaviorAnalyzer.inferQueryType(query)
}

func (cm *CacheManager) getTopPatterns(patterns map[string]*UserPattern) []string {
	type patternFreq struct {
		ID   string
		Freq int
	}

	var sorted []patternFreq
	for id, pattern := range patterns {
		sorted = append(sorted, patternFreq{ID: id, Freq: pattern.Frequency})
	}

	// Sort by frequency (simple bubble sort for small datasets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Freq < sorted[j+1].Freq {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	var top []string
	for i := 0; i < len(sorted) && i < 5; i++ {
		top = append(top, sorted[i].ID)
	}

	return top
}

func (cm *CacheManager) calculateOverallHealth(pregenStats PregenStats, streamStats StreamStats, cacheStats map[string]interface{}) string {
	// Simple health calculation based on various metrics
	score := 0.0

	// Cache hit rate (target: >0.8)
	if pregenStats.CacheHits+pregenStats.CacheMisses > 0 {
		hitRate := float64(pregenStats.CacheHits) / float64(pregenStats.CacheHits+pregenStats.CacheMisses)
		if hitRate > 0.8 {
			score += 0.3
		} else if hitRate > 0.6 {
			score += 0.2
		} else if hitRate > 0.4 {
			score += 0.1
		}
	}

	// Pre-generation success rate (target: >0.9)
	if pregenStats.SuccessfulPregens+pregenStats.FailedPregens > 0 {
		successRate := float64(pregenStats.SuccessfulPregens) / float64(pregenStats.SuccessfulPregens+pregenStats.FailedPregens)
		if successRate > 0.9 {
			score += 0.3
		} else if successRate > 0.8 {
			score += 0.2
		} else if successRate > 0.7 {
			score += 0.1
		}
	}

	// Worker utilization (target: 0.6-0.8)
	if pregenStats.WorkerUtilization >= 0.6 && pregenStats.WorkerUtilization <= 0.8 {
		score += 0.2
	} else if pregenStats.WorkerUtilization >= 0.4 && pregenStats.WorkerUtilization <= 0.9 {
		score += 0.1
	}

	// Streaming performance
	if streamStats.ActiveStreams > 0 && streamStats.ErrorRate < 0.1 {
		score += 0.2
	}

	if score >= 0.8 {
		return "excellent"
	} else if score >= 0.6 {
		return "good"
	} else if score >= 0.4 {
		return "fair"
	} else {
		return "poor"
	}
}