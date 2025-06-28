package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DiskCache implements a persistent disk-based cache
type DiskCache struct {
	mu       sync.RWMutex
	config   *CacheConfig
	basePath string
	stats    CacheStats
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewDiskCache creates a new disk-based cache
func NewDiskCache(config *CacheConfig) (*DiskCache, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// Set default cache path if not specified
	if config.DiskPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		config.DiskPath = filepath.Join(homeDir, ".cache", "nixai")
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(config.DiskPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cache := &DiskCache{
		config:   config,
		basePath: config.DiskPath,
		stats:    CacheStats{LastCleanup: time.Now()},
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}

	// Initialize cache size
	if err := cache.calculateDiskUsage(); err != nil {
		return nil, fmt.Errorf("failed to calculate disk usage: %w", err)
	}

	// Start background maintenance
	go cache.maintenanceRoutine()

	return cache, nil
}

// Get retrieves a value from the disk cache
func (dc *DiskCache) Get(ctx context.Context, key string) ([]byte, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	filePath := dc.getFilePath(key)

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			dc.stats.Misses++
			return nil, false
		}
		dc.stats.Misses++
		return nil, false
	}

	// Read cache entry
	data, err := os.ReadFile(filePath)
	if err != nil {
		dc.stats.Misses++
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		dc.stats.Misses++
		return nil, false
	}

	// Check if expired
	if entry.IsExpired() {
		// Remove expired file in background
		go func() {
			os.Remove(filePath)
		}()
		dc.stats.Misses++
		return nil, false
	}

	// Update access time in background to avoid blocking
	go func() {
		dc.mu.Lock()
		defer dc.mu.Unlock()

		entry.UpdateAccess()
		if updatedData, err := json.Marshal(entry); err == nil {
			os.WriteFile(filePath, updatedData, 0644)
		}
	}()

	// Update modification time for LRU tracking
	now := time.Now()
	os.Chtimes(filePath, now, now)

	dc.stats.Hits++
	return entry.Value, true
}

// Set stores a value in the disk cache
func (dc *DiskCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()
	if ttl == 0 {
		ttl = dc.config.DiskTTL
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

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	filePath := dc.getFilePath(key)

	// Check if file already exists to calculate size difference
	var oldSize int64
	if info, err := os.Stat(filePath); err == nil {
		oldSize = info.Size()
	}

	// Write to temporary file first, then rename (atomic operation)
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	// Update stats
	newSize := int64(len(data))
	if oldSize == 0 {
		dc.stats.Size++
	}
	dc.stats.SizeBytes = dc.stats.SizeBytes - oldSize + newSize

	// Check if we need to evict old entries
	go dc.evictIfNeeded()

	return nil
}

// Delete removes a key from the disk cache
func (dc *DiskCache) Delete(ctx context.Context, key string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	filePath := dc.getFilePath(key)

	// Get file size before deletion
	if info, err := os.Stat(filePath); err == nil {
		dc.stats.Size--
		dc.stats.SizeBytes -= info.Size()
	}

	return os.Remove(filePath)
}

// Clear removes all entries from the disk cache
func (dc *DiskCache) Clear(ctx context.Context) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Remove all files in cache directory
	entries, err := os.ReadDir(dc.basePath)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".cache" {
			filePath := filepath.Join(dc.basePath, entry.Name())
			os.Remove(filePath)
		}
	}

	dc.stats.Size = 0
	dc.stats.SizeBytes = 0

	return nil
}

// Stats returns cache statistics
func (dc *DiskCache) Stats() CacheStats {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	stats := dc.stats
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}

	return stats
}

// Close gracefully shuts down the disk cache
func (dc *DiskCache) Close() error {
	close(dc.stopCh)
	<-dc.doneCh
	return nil
}

// getFilePath returns the file path for a cache key
func (dc *DiskCache) getFilePath(key string) string {
	// Use first 2 characters for subdirectory to avoid too many files in one directory
	subdir := ""
	if len(key) >= 2 {
		subdir = key[:2]
	}

	dirPath := filepath.Join(dc.basePath, subdir)
	os.MkdirAll(dirPath, 0755)

	return filepath.Join(dirPath, key+".cache")
}

// calculateDiskUsage calculates current disk usage
func (dc *DiskCache) calculateDiskUsage() error {
	var totalSize int64
	var fileCount int64

	err := filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking, ignore errors
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".cache" {
			if info, err := d.Info(); err == nil {
				totalSize += info.Size()
				fileCount++
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	dc.stats.Size = fileCount
	dc.stats.SizeBytes = totalSize

	return nil
}

// evictIfNeeded removes old files if cache exceeds size limit
func (dc *DiskCache) evictIfNeeded() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if dc.stats.SizeBytes <= dc.config.DiskMaxSize {
		return
	}

	// Get list of cache files with their modification times
	type fileInfo struct {
		path    string
		modTime time.Time
		size    int64
	}

	var files []fileInfo

	filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".cache" {
			if info, err := d.Info(); err == nil {
				files = append(files, fileInfo{
					path:    path,
					modTime: info.ModTime(),
					size:    info.Size(),
				})
			}
		}

		return nil
	})

	// Sort by modification time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove oldest files until we're under the size limit
	for _, file := range files {
		if dc.stats.SizeBytes <= dc.config.DiskMaxSize {
			break
		}

		if err := os.Remove(file.path); err == nil {
			dc.stats.Size--
			dc.stats.SizeBytes -= file.size
			dc.stats.Evictions++
		}
	}
}

// maintenanceRoutine runs periodic maintenance tasks
func (dc *DiskCache) maintenanceRoutine() {
	defer close(dc.doneCh)

	cleanupTicker := time.NewTicker(dc.config.CleanupInterval)
	compactTicker := time.NewTicker(dc.config.CompactInterval)
	defer cleanupTicker.Stop()
	defer compactTicker.Stop()

	for {
		select {
		case <-dc.stopCh:
			return
		case <-cleanupTicker.C:
			dc.cleanup()
		case <-compactTicker.C:
			dc.compact()
		}
	}
}

// cleanup removes expired cache files
func (dc *DiskCache) cleanup() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	now := time.Now()
	var removedCount int64
	var removedSize int64

	filepath.WalkDir(dc.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".cache" {
			// Read cache entry to check expiration
			if data, err := os.ReadFile(path); err == nil {
				var entry CacheEntry
				if json.Unmarshal(data, &entry) == nil && entry.IsExpired() {
					if info, err := d.Info(); err == nil {
						if os.Remove(path) == nil {
							removedCount++
							removedSize += info.Size()
						}
					}
				}
			}
		}

		return nil
	})

	dc.stats.Size -= removedCount
	dc.stats.SizeBytes -= removedSize
	dc.stats.LastCleanup = now
}

// compact reorganizes cache files for better performance
func (dc *DiskCache) compact() {
	// For now, compact just runs cleanup and eviction
	// In the future, this could reorganize files or defragment
	dc.cleanup()
	dc.evictIfNeeded()
}
