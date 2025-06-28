package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// Manager coordinates multiple cache layers (memory + disk)
type Manager struct {
	memory Cache
	disk   Cache
	config *CacheConfig
	logger *logger.Logger
}

// NewManager creates a new multi-tier cache manager
func NewManager(config *CacheConfig, log *logger.Logger) (*Manager, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	if log == nil {
		log = logger.NewLogger()
	}

	// Create memory cache
	memoryCache := NewMemoryCache(config)

	// Create disk cache if enabled
	var diskCache Cache
	if config.DiskEnabled {
		dc, err := NewDiskCache(config)
		if err != nil {
			log.Warn(fmt.Sprintf("Failed to create disk cache, using memory only: %v", err))
			diskCache = nil
		} else {
			diskCache = dc
		}
	}

	return &Manager{
		memory: memoryCache,
		disk:   diskCache,
		config: config,
		logger: log,
	}, nil
}

// Get retrieves a value from cache, checking memory first, then disk
func (m *Manager) Get(ctx context.Context, key string) ([]byte, bool) {
	// Try memory cache first
	if value, found := m.memory.Get(ctx, key); found {
		return value, true
	}

	// Try disk cache if available
	if m.disk != nil {
		if value, found := m.disk.Get(ctx, key); found {
			// Promote to memory cache for faster future access
			go func() {
				if err := m.memory.Set(context.Background(), key, value, m.config.MemoryTTL); err != nil {
					m.logger.Debug(fmt.Sprintf("Failed to promote cache entry to memory: %v", err))
				}
			}()
			return value, true
		}
	}

	return nil, false
}

// Set stores a value in both memory and disk caches
func (m *Manager) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Store in memory cache
	if err := m.memory.Set(ctx, key, value, ttl); err != nil {
		m.logger.Error(fmt.Sprintf("Failed to store in memory cache: %v", err))
	}

	// Store in disk cache if available
	if m.disk != nil {
		if err := m.disk.Set(ctx, key, value, ttl); err != nil {
			m.logger.Debug(fmt.Sprintf("Failed to store in disk cache: %v", err))
		}
	}

	return nil
}

// Delete removes a key from both caches
func (m *Manager) Delete(ctx context.Context, key string) error {
	// Remove from memory cache
	if err := m.memory.Delete(ctx, key); err != nil {
		m.logger.Debug(fmt.Sprintf("Failed to delete from memory cache: %v", err))
	}

	// Remove from disk cache if available
	if m.disk != nil {
		if err := m.disk.Delete(ctx, key); err != nil {
			m.logger.Debug(fmt.Sprintf("Failed to delete from disk cache: %v", err))
		}
	}

	return nil
}

// Clear removes all entries from both caches
func (m *Manager) Clear(ctx context.Context) error {
	// Clear memory cache
	if err := m.memory.Clear(ctx); err != nil {
		m.logger.Error(fmt.Sprintf("Failed to clear memory cache: %v", err))
	}

	// Clear disk cache if available
	if m.disk != nil {
		if err := m.disk.Clear(ctx); err != nil {
			m.logger.Error(fmt.Sprintf("Failed to clear disk cache: %v", err))
		}
	}

	return nil
}

// Stats returns combined cache statistics
func (m *Manager) Stats() CombinedCacheStats {
	memStats := m.memory.Stats()

	var diskStats CacheStats
	if m.disk != nil {
		diskStats = m.disk.Stats()
	}

	return CombinedCacheStats{
		Memory: memStats,
		Disk:   diskStats,
		Combined: CacheStats{
			Hits:        memStats.Hits + diskStats.Hits,
			Misses:      memStats.Misses + diskStats.Misses,
			Size:        memStats.Size + diskStats.Size,
			SizeBytes:   memStats.SizeBytes + diskStats.SizeBytes,
			Evictions:   memStats.Evictions + diskStats.Evictions,
			LastCleanup: memStats.LastCleanup,
		},
	}
}

// Close gracefully shuts down all caches
func (m *Manager) Close() error {
	if err := m.memory.Close(); err != nil {
		m.logger.Error(fmt.Sprintf("Failed to close memory cache: %v", err))
	}

	if m.disk != nil {
		if err := m.disk.Close(); err != nil {
			m.logger.Error(fmt.Sprintf("Failed to close disk cache: %v", err))
		}
	}

	return nil
}

// CombinedCacheStats provides statistics for multi-tier cache
type CombinedCacheStats struct {
	Memory   CacheStats `json:"memory"`
	Disk     CacheStats `json:"disk"`
	Combined CacheStats `json:"combined"`
}

// CalculateHitRate calculates the overall hit rate
func (s *CombinedCacheStats) CalculateHitRate() float64 {
	total := s.Combined.Hits + s.Combined.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Combined.Hits) / float64(total)
}

// AI Response Cache Helper Functions

// GetAIResponse retrieves a cached AI response
func (m *Manager) GetAIResponse(ctx context.Context, provider, model, prompt string) ([]byte, bool) {
	key := GenerateKey(PrefixAIResponse, provider, model, prompt)
	return m.Get(ctx, key)
}

// SetAIResponse caches an AI response with default TTL
func (m *Manager) SetAIResponse(ctx context.Context, provider, model, prompt string, response []byte) error {
	key := GenerateKey(PrefixAIResponse, provider, model, prompt)
	// AI responses cached for 30 days
	return m.Set(ctx, key, response, 30*24*time.Hour)
}

// Documentation Cache Helper Functions

// GetDocumentation retrieves documentation from cache
func (m *Manager) GetDocumentation(ctx context.Context, key string) ([]byte, bool) {
	return m.Get(ctx, "doc:"+key)
}

// SetDocumentation stores documentation in cache
func (m *Manager) SetDocumentation(ctx context.Context, key string, data []byte) error {
	return m.Set(ctx, "doc:"+key, data, m.config.MemoryTTL)
}

// GetDocumentationResponse retrieves a cached documentation response
func (m *Manager) GetDocumentationResponse(ctx context.Context, query string, sources []string) ([]byte, bool) {
	if m == nil || m.memory == nil {
		return nil, false
	}

	// Create cache key that includes both query and sources
	sourcesStr := strings.Join(sources, ",")
	key := fmt.Sprintf("doc:%s:%s", query, sourcesStr)

	// Try memory cache first
	if data, found := m.memory.Get(ctx, key); found {
		return data, true
	}

	// Try disk cache if enabled
	if m.disk != nil {
		if data, found := m.disk.Get(ctx, key); found {
			// Promote to memory cache
			_ = m.memory.Set(ctx, key, data, m.config.MemoryTTL)
			return data, true
		}
	}

	return nil, false
}

// SetDocumentationResponse stores a documentation response in cache
func (m *Manager) SetDocumentationResponse(ctx context.Context, query string, sources []string, response []byte) error {
	if m == nil || m.memory == nil {
		return nil
	}

	// Create cache key that includes both query and sources
	sourcesStr := strings.Join(sources, ",")
	key := fmt.Sprintf("doc:%s:%s", query, sourcesStr)

	// Store in memory cache
	if err := m.memory.Set(ctx, key, response, m.config.MemoryTTL); err != nil {
		return fmt.Errorf("failed to store in memory cache: %w", err)
	}

	// Store in disk cache if enabled
	if m.disk != nil {
		if err := m.disk.Set(ctx, key, response, m.config.DiskTTL); err != nil {
			m.logger.Debug(fmt.Sprintf("Failed to store in disk cache: %v", err))
			// Don't return error for disk cache failures
		}
	}

	return nil
}

// GetMCPQuery retrieves MCP query result from cache
func (m *Manager) GetMCPQuery(ctx context.Context, query string) ([]byte, bool) {
	return m.Get(ctx, "mcp:"+query)
}

// SetMCPQuery stores MCP query result in cache
func (m *Manager) SetMCPQuery(ctx context.Context, query string, data []byte) error {
	return m.Set(ctx, "mcp:"+query, data, m.config.MemoryTTL)
}

// GetSystemDiagnostic retrieves system diagnostic from cache
func (m *Manager) GetSystemDiagnostic(ctx context.Context, key string) ([]byte, bool) {
	return m.Get(ctx, "diag:"+key)
}

// SetSystemDiagnostic stores system diagnostic in cache
func (m *Manager) SetSystemDiagnostic(ctx context.Context, key string, data []byte) error {
	return m.Set(ctx, "diag:"+key, data, m.config.MemoryTTL)
}

// System Info Cache Helper Functions

// GetSystemInfo retrieves cached system information
func (m *Manager) GetSystemInfo(ctx context.Context, infoType string) ([]byte, bool) {
	key := GenerateKey(PrefixSystemInfo, infoType)
	return m.Get(ctx, key)
}

// SetSystemInfo caches system information with shorter TTL
func (m *Manager) SetSystemInfo(ctx context.Context, infoType string, info []byte) error {
	key := GenerateKey(PrefixSystemInfo, infoType)
	// System info cached for 1 hour
	return m.Set(ctx, key, info, 1*time.Hour)
}

// Package Cache Helper Functions

// GetPackageInfo retrieves cached package information
func (m *Manager) GetPackageInfo(ctx context.Context, packageName string) ([]byte, bool) {
	key := GenerateKey(PrefixPackages, packageName)
	return m.Get(ctx, key)
}

// SetPackageInfo caches package information
func (m *Manager) SetPackageInfo(ctx context.Context, packageName string, info []byte) error {
	key := GenerateKey(PrefixPackages, packageName)
	// Package info cached for 24 hours
	return m.Set(ctx, key, info, 24*time.Hour)
}

// Configuration Cache Helper Functions

// GetConfiguration retrieves cached configuration data
func (m *Manager) GetConfiguration(ctx context.Context, configPath string) ([]byte, bool) {
	key := GenerateKey(PrefixConfiguration, configPath)
	return m.Get(ctx, key)
}

// SetConfiguration caches configuration data
func (m *Manager) SetConfiguration(ctx context.Context, configPath string, config []byte) error {
	key := GenerateKey(PrefixConfiguration, configPath)
	// Configuration cached for 4 hours
	return m.Set(ctx, key, config, 4*time.Hour)
}

// InvalidateByPrefix removes all cache entries with a specific prefix
func (m *Manager) InvalidateByPrefix(ctx context.Context, prefix string) error {
	// This is a simplified implementation
	// In a production system, we might want to track keys by prefix
	m.logger.Debug(fmt.Sprintf("Invalidating cache entries with prefix: %s", prefix))

	// For now, we'll implement this as a future enhancement
	// when we add key tracking by prefix

	return nil
}
