// Package cache provides intelligent caching for AI responses, documentation, and system data
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os/user"
	"path/filepath"
	"time"
)

// Cache defines the interface for all cache implementations
type Cache interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key string) ([]byte, bool)

	// Set stores a value with key and TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Clear removes all entries from cache
	Clear(ctx context.Context) error

	// Stats returns cache statistics
	Stats() CacheStats

	// Close gracefully shuts down the cache
	Close() error
}

// CacheStats provides cache performance metrics
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	HitRate     float64   `json:"hit_rate"`
	Size        int64     `json:"size"`       // Number of entries
	SizeBytes   int64     `json:"size_bytes"` // Memory usage in bytes
	Evictions   int64     `json:"evictions"`  // Number of evicted entries
	LastCleanup time.Time `json:"last_cleanup"`
}

// CacheConfig defines cache configuration options
type CacheConfig struct {
	// Memory cache settings
	MemoryMaxSize int           `yaml:"memory_max_size" json:"memory_max_size"` // Max entries in memory
	MemoryTTL     time.Duration `yaml:"memory_ttl" json:"memory_ttl"`           // Default TTL for memory cache

	// Disk cache settings
	DiskEnabled bool          `yaml:"disk_enabled" json:"disk_enabled"`   // Enable persistent disk cache
	DiskPath    string        `yaml:"disk_path" json:"disk_path"`         // Path for disk cache
	DiskMaxSize int64         `yaml:"disk_max_size" json:"disk_max_size"` // Max disk cache size in bytes
	DiskTTL     time.Duration `yaml:"disk_ttl" json:"disk_ttl"`           // Default TTL for disk cache

	// Cleanup settings
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"` // How often to run cleanup
	CompactInterval time.Duration `yaml:"compact_interval" json:"compact_interval"` // How often to compact disk cache
}

// DefaultCacheConfig returns sensible defaults for cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MemoryMaxSize:   1000,              // 1000 entries in memory
		MemoryTTL:       30 * time.Minute,  // 30 minutes for memory cache
		DiskEnabled:     true,              // Enable disk persistence
		DiskPath:        "",                // Will be set to user cache dir
		DiskMaxSize:     100 * 1024 * 1024, // 100MB disk cache limit
		DiskTTL:         24 * time.Hour,    // 24 hours for disk cache
		CleanupInterval: 5 * time.Minute,   // Cleanup every 5 minutes
		CompactInterval: 1 * time.Hour,     // Compact every hour
	}
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Key         string    `json:"key"`
	Value       []byte    `json:"value"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
	Size        int64     `json:"size"`
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// UpdateAccess updates access tracking for the entry
func (e *CacheEntry) UpdateAccess() {
	e.AccessCount++
	e.LastAccess = time.Now()
}

// GenerateKey creates a consistent cache key from input data
func GenerateKey(prefix string, data ...string) string {
	h := sha256.New()
	h.Write([]byte(prefix))
	for _, d := range data {
		h.Write([]byte(d))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// CacheKeyPrefixes define consistent prefixes for different cache types
const (
	PrefixAIResponse    = "ai_response"
	PrefixDocumentation = "docs"
	PrefixSystemInfo    = "system"
	PrefixNixOSOptions  = "nixos_options"
	PrefixPackages      = "packages"
	PrefixConfiguration = "config"
)

// ConfigCacheConfig represents cache configuration from the config package
// This is used to avoid import cycles while converting config settings
type ConfigCacheConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	MemoryMaxSize   int    `yaml:"memory_max_size" json:"memory_max_size"`
	MemoryTTL       int    `yaml:"memory_ttl" json:"memory_ttl"`
	DiskEnabled     bool   `yaml:"disk_enabled" json:"disk_enabled"`
	DiskPath        string `yaml:"disk_path" json:"disk_path"`
	DiskMaxSize     int64  `yaml:"disk_max_size" json:"disk_max_size"`
	DiskTTL         int    `yaml:"disk_ttl" json:"disk_ttl"`
	CleanupInterval int    `yaml:"cleanup_interval" json:"cleanup_interval"`
	CompactInterval int    `yaml:"compact_interval" json:"compact_interval"`
}

// FromConfigCacheConfig converts a config.CacheConfig to cache.CacheConfig
func FromConfigCacheConfig(configCache ConfigCacheConfig) *CacheConfig {
	// Set default disk path if not specified
	diskPath := configCache.DiskPath
	if diskPath == "" {
		if usr, err := user.Current(); err == nil {
			diskPath = filepath.Join(usr.HomeDir, ".cache", "nixai")
		} else {
			diskPath = "/tmp/nixai-cache"
		}
	}

	return &CacheConfig{
		MemoryMaxSize:   configCache.MemoryMaxSize,
		MemoryTTL:       time.Duration(configCache.MemoryTTL) * time.Minute,
		DiskEnabled:     configCache.DiskEnabled,
		DiskPath:        diskPath,
		DiskMaxSize:     configCache.DiskMaxSize * 1024 * 1024, // Convert MB to bytes
		DiskTTL:         time.Duration(configCache.DiskTTL) * time.Hour,
		CleanupInterval: time.Duration(configCache.CleanupInterval) * time.Minute,
		CompactInterval: time.Duration(configCache.CompactInterval) * time.Minute,
	}
}
