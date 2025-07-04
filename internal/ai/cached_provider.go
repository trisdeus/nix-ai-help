// Package ai provides cache-aware AI provider implementations
package ai

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/intelligence/cache"
	"nix-ai-help/pkg/logger"
)

// CachedProvider wraps any AI provider with intelligent caching capabilities
type CachedProvider struct {
	provider     Provider
	cacheManager *cache.CacheManager
	config       *config.UserConfig
	logger       *logger.Logger
	enabled      bool
}

// CachedProviderConfig configures cache behavior for AI providers
type CachedProviderConfig struct {
	EnableCaching       bool          `json:"enable_caching"`
	EnableBehaviorAnalysis bool       `json:"enable_behavior_analysis"`
	EnablePregeneration bool          `json:"enable_pregeneration"`
	EnableStreaming     bool          `json:"enable_streaming"`
	CacheTimeout        time.Duration `json:"cache_timeout"`
	PregenThreshold     float64       `json:"pregen_threshold"`
}

// NewCachedProvider creates a new cache-aware AI provider
func NewCachedProvider(provider Provider, cfg *config.UserConfig) (*CachedProvider, error) {
	// Create cache manager with the provider as the response generator
	responseGen := &ProviderResponseGenerator{provider: provider}
	cacheManager := cache.NewCacheManager(cfg, responseGen)

	cachedProvider := &CachedProvider{
		provider:     provider,
		cacheManager: cacheManager,
		config:       cfg,
		logger:       logger.NewLogger(),
		enabled:      true, // Enable by default
	}

	return cachedProvider, nil
}

// Start initializes the cached provider and starts cache subsystems
func (cp *CachedProvider) Start(ctx context.Context) error {
	if !cp.enabled {
		return nil
	}

	cp.logger.Info("Starting cached AI provider")
	return cp.cacheManager.Start(ctx)
}

// Stop gracefully stops the cached provider
func (cp *CachedProvider) Stop() error {
	if !cp.enabled {
		return nil
	}

	cp.logger.Info("Stopping cached AI provider")
	return cp.cacheManager.Stop()
}

// Query implements the Provider interface with caching
func (cp *CachedProvider) Query(prompt string) (string, error) {
	ctx := context.Background()
	return cp.GenerateResponse(ctx, prompt)
}

// GenerateResponse implements the Provider interface with intelligent caching
func (cp *CachedProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if !cp.enabled {
		return cp.provider.GenerateResponse(ctx, prompt)
	}

	// Create query context
	queryContext := cp.buildQueryContext(ctx, prompt)
	
	// Create query request
	request := cache.QueryRequest{
		ID:        cp.generateRequestID(prompt),
		Query:     prompt,
		Context:   queryContext,
		Options: cache.QueryOptions{
			EnableStreaming:     false, // Standard non-streaming query
			EnablePregeneration: true,
			ForceRefresh:        false,
			Priority:           "normal",
			Timeout:            30 * time.Second,
		},
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Process through cache manager
	response, err := cp.cacheManager.ProcessQuery(ctx, request)
	if err != nil {
		cp.logger.Error(fmt.Sprintf("Cache manager error: %v", err))
		// Fallback to direct provider call
		return cp.provider.GenerateResponse(ctx, prompt)
	}

	cp.logger.Info(fmt.Sprintf("Query processed via cache (source: %s, cache_hit: %t)", 
		response.Source, response.CacheHit))

	return response.Response, nil
}

// StreamResponse implements streaming with cache integration
func (cp *CachedProvider) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	if !cp.enabled {
		return cp.provider.StreamResponse(ctx, prompt)
	}

	// Check cache first for streaming responses
	queryContext := cp.buildQueryContext(ctx, prompt)
	if cachedResponse, found := cp.cacheManager.GetCachedResponse(prompt, queryContext); found {
		// Return cached response as a stream
		return cp.streamCachedResponse(cachedResponse), nil
	}

	// Create streaming response channel
	responseChan := make(chan StreamResponse, 10)

	go func() {
		defer close(responseChan)

		// Create string builder to capture streaming response
		var responseBuilder strings.Builder
		streamWriter := &StreamWriter{builder: &responseBuilder, responseChan: responseChan}

		// Create query request with streaming enabled
		request := cache.QueryRequest{
			ID:           cp.generateRequestID(prompt),
			Query:        prompt,
			Context:      queryContext,
			StreamWriter: streamWriter,
			Options: cache.QueryOptions{
				EnableStreaming:     true,
				EnablePregeneration: true,
				ForceRefresh:        false,
				Priority:           "normal",
				Timeout:            60 * time.Second,
			},
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		// Process with streaming through cache manager
		response, err := cp.cacheManager.ProcessQuery(ctx, request)
		if err != nil {
			cp.logger.Error(fmt.Sprintf("Streaming cache error: %v", err))
			
			// Fallback to direct provider streaming
			providerChan, providerErr := cp.provider.StreamResponse(ctx, prompt)
			if providerErr != nil {
				responseChan <- StreamResponse{
					Content: "",
					Error:   fmt.Errorf("both cache and provider failed: cache=%v, provider=%v", err, providerErr),
					Done:    true,
				}
				return
			}

			// Forward provider responses
			for chunk := range providerChan {
				responseChan <- chunk
			}
			return
		}

		// Send final response if not already streamed
		if response.Source != cache.SourceStreaming {
			responseChan <- StreamResponse{
				Content: response.Response,
				Done:    true,
			}
		}
	}()

	return responseChan, nil
}

// GetPartialResponse delegates to the underlying provider
func (cp *CachedProvider) GetPartialResponse() string {
	return cp.provider.GetPartialResponse()
}

// GetCacheStats returns comprehensive cache statistics
func (cp *CachedProvider) GetCacheStats() cache.CacheStats {
	if !cp.enabled {
		return cache.CacheStats{}
	}
	return cp.cacheManager.GetStats()
}

// AnalyzeBehavior performs behavior analysis
func (cp *CachedProvider) AnalyzeBehavior(ctx context.Context) ([]cache.BehaviorInsight, error) {
	if !cp.enabled {
		return nil, fmt.Errorf("caching not enabled")
	}
	return cp.cacheManager.AnalyzeBehavior(ctx)
}

// PredictNextQueries predicts likely next queries
func (cp *CachedProvider) PredictNextQueries(ctx context.Context) (*cache.PredictionResult, error) {
	if !cp.enabled {
		return nil, fmt.Errorf("caching not enabled")
	}
	
	queryContext := cp.buildQueryContext(ctx, "")
	return cp.cacheManager.PredictNextQueries(ctx, queryContext)
}

// InvalidateCache invalidates cache entries by tag
func (cp *CachedProvider) InvalidateCache(tag string) int {
	if !cp.enabled {
		return 0
	}
	return cp.cacheManager.InvalidateCache(tag)
}

// SaveBehaviorPatterns saves behavior patterns to file
func (cp *CachedProvider) SaveBehaviorPatterns(filePath string) error {
	if !cp.enabled {
		return fmt.Errorf("caching not enabled")
	}
	return cp.cacheManager.SaveBehaviorPatterns(filePath)
}

// LoadBehaviorPatterns loads behavior patterns from file
func (cp *CachedProvider) LoadBehaviorPatterns(filePath string) error {
	if !cp.enabled {
		return fmt.Errorf("caching not enabled")
	}
	return cp.cacheManager.LoadBehaviorPatterns(filePath)
}

// SetCacheEnabled enables or disables caching
func (cp *CachedProvider) SetCacheEnabled(enabled bool) {
	cp.enabled = enabled
	if enabled {
		cp.logger.Info("Cache enabled for AI provider")
	} else {
		cp.logger.Info("Cache disabled for AI provider")
	}
}

// IsCacheEnabled returns whether caching is enabled
func (cp *CachedProvider) IsCacheEnabled() bool {
	return cp.enabled
}

// Private helper methods

func (cp *CachedProvider) buildQueryContext(ctx context.Context, prompt string) cache.QueryContext {
	// Extract context information
	workingDir := "/"
	if wd, ok := ctx.Value("working_directory").(string); ok && wd != "" {
		workingDir = wd
	}

	projectType := "unknown"
	if pt, ok := ctx.Value("project_type").(string); ok && pt != "" {
		projectType = pt
	}

	// Build previous queries from context if available
	var previousQueries []string
	if pq, ok := ctx.Value("previous_queries").([]string); ok {
		previousQueries = pq
	}

	return cache.QueryContext{
		WorkingDirectory: workingDir,
		ProjectType:      projectType,
		PreviousQueries:  previousQueries,
		TimeOfDay:        time.Now().Hour(),
		DayOfWeek:        int(time.Now().Weekday()),
		SessionLength:    time.Since(time.Now()), // Simplified
		Environment:      make(map[string]string),
	}
}

func (cp *CachedProvider) generateRequestID(prompt string) string {
	return fmt.Sprintf("req_%d_%d", time.Now().Unix(), len(prompt))
}

func (cp *CachedProvider) streamCachedResponse(response string) <-chan StreamResponse {
	responseChan := make(chan StreamResponse, 1)
	
	go func() {
		defer close(responseChan)
		
		// Stream cached response in chunks to simulate real-time delivery
		chunkSize := 100
		for i := 0; i < len(response); i += chunkSize {
			end := i + chunkSize
			if end > len(response) {
				end = len(response)
			}

			chunk := response[i:end]
			responseChan <- StreamResponse{
				Content: chunk,
				Done:    end >= len(response),
			}

			// Small delay to simulate streaming
			if end < len(response) {
				time.Sleep(20 * time.Millisecond)
			}
		}
	}()

	return responseChan
}

// ProviderResponseGenerator adapts a Provider to implement cache.ResponseGenerator
type ProviderResponseGenerator struct {
	provider Provider
}

// GenerateResponse implements cache.ResponseGenerator
func (prg *ProviderResponseGenerator) GenerateResponse(ctx context.Context, query string, context cache.QueryContext) (string, error) {
	// Add context information to the context
	enhancedCtx := ctx
	enhancedCtx = contextWithValue(enhancedCtx, "working_directory", context.WorkingDirectory)
	enhancedCtx = contextWithValue(enhancedCtx, "project_type", context.ProjectType)
	enhancedCtx = contextWithValue(enhancedCtx, "previous_queries", context.PreviousQueries)

	return prg.provider.GenerateResponse(enhancedCtx, query)
}

// StreamWriter implements io.Writer to capture streaming responses
type StreamWriter struct {
	builder      *strings.Builder
	responseChan chan<- StreamResponse
}

// Compile-time check that StreamWriter implements io.Writer
var _ io.Writer = (*StreamWriter)(nil)

// Write implements io.Writer
func (sw *StreamWriter) Write(p []byte) (n int, err error) {
	content := string(p)
	
	// Send to response channel
	sw.responseChan <- StreamResponse{
		Content: content,
		Done:    false,
	}

	// Also capture in builder
	return sw.builder.Write(p)
}

// contextWithValue is a helper to add values to context (simplified implementation)
func contextWithValue(ctx context.Context, key, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// CachedProviderFactory extends ProviderFactory with caching capabilities
type CachedProviderFactory struct {
	*ProviderFactory
	cachedProviders map[string]*CachedProvider
	config          *config.UserConfig
}

// NewCachedProviderFactory creates a factory that provides cached AI providers
func NewCachedProviderFactory(cfg *config.UserConfig) *CachedProviderFactory {
	return &CachedProviderFactory{
		ProviderFactory: NewProviderFactory(),
		cachedProviders: make(map[string]*CachedProvider),
		config:          cfg,
	}
}

// GetCachedProvider retrieves or creates a cached version of a provider
func (cpf *CachedProviderFactory) GetCachedProvider(name string) (*CachedProvider, error) {
	// Check if we already have a cached version
	if cachedProvider, exists := cpf.cachedProviders[name]; exists {
		return cachedProvider, nil
	}

	// Get the base provider
	provider, exists := cpf.GetProvider(name)
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	// Create cached version
	cachedProvider, err := NewCachedProvider(provider, cpf.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cached provider: %w", err)
	}

	// Store for future use
	cpf.cachedProviders[name] = cachedProvider

	return cachedProvider, nil
}

// StartAllCaches starts caching for all cached providers
func (cpf *CachedProviderFactory) StartAllCaches(ctx context.Context) error {
	for name, cachedProvider := range cpf.cachedProviders {
		if err := cachedProvider.Start(ctx); err != nil {
			return fmt.Errorf("failed to start cache for provider %s: %w", name, err)
		}
	}
	return nil
}

// StopAllCaches stops caching for all cached providers
func (cpf *CachedProviderFactory) StopAllCaches() error {
	for name, cachedProvider := range cpf.cachedProviders {
		if err := cachedProvider.Stop(); err != nil {
			return fmt.Errorf("failed to stop cache for provider %s: %w", name, err)
		}
	}
	return nil
}