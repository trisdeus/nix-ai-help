package performance

import (
	"fmt"
	"sync"
	"time"

	"nix-ai-help/internal/cache"
	"nix-ai-help/pkg/logger"
)

// MetricType represents different types of performance metrics
type MetricType string

const (
	MetricAIQuery            MetricType = "ai_query"
	MetricCacheHit           MetricType = "cache_hit"
	MetricCacheMiss          MetricType = "cache_miss"
	MetricDocumentationQuery MetricType = "documentation_query"
	MetricMCPQuery           MetricType = "mcp_query"
	MetricSystemDiagnostic   MetricType = "system_diagnostic"
	MetricParallelOperation  MetricType = "parallel_operation"
)

// Metric represents a single performance measurement
type Metric struct {
	Type      MetricType
	Name      string
	Duration  time.Duration
	Timestamp time.Time
	Tags      map[string]string
	Success   bool
	Error     string
}

// MetricsSummary provides aggregated performance data
type MetricsSummary struct {
	TotalOperations  int64
	SuccessfulOps    int64
	FailedOps        int64
	AverageDuration  time.Duration
	MinDuration      time.Duration
	MaxDuration      time.Duration
	CacheHitRate     float64
	OperationsByType map[MetricType]int64
	RecentOperations []Metric
	PerformanceGains map[string]float64
}

// Monitor tracks performance metrics across nixai operations
type Monitor struct {
	metrics []Metric
	mutex   sync.RWMutex
	logger  *logger.Logger

	// Performance baselines for comparison
	baselines map[string]time.Duration
}

// NewMonitor creates a new performance monitor
func NewMonitor(log *logger.Logger) *Monitor {
	if log == nil {
		log = logger.NewLogger()
	}

	return &Monitor{
		metrics:   make([]Metric, 0, 1000), // Pre-allocate for 1000 metrics
		logger:    log,
		baselines: make(map[string]time.Duration),
	}
}

// RecordMetric adds a new performance metric
func (m *Monitor) RecordMetric(metric Metric) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric.Timestamp = time.Now()
	m.metrics = append(m.metrics, metric)

	// Keep only last 1000 metrics to prevent memory growth
	if len(m.metrics) > 1000 {
		m.metrics = m.metrics[len(m.metrics)-1000:]
	}

	// Log significant performance events
	if metric.Duration > 5*time.Second {
		m.logger.Warn(fmt.Sprintf("Slow operation detected: %s took %v",
			metric.Name, metric.Duration))
	}

	// Log performance improvements
	if baseline, exists := m.baselines[metric.Name]; exists {
		improvement := float64(baseline-metric.Duration) / float64(baseline) * 100
		if improvement > 50 { // 50% improvement
			m.logger.Info(fmt.Sprintf("Performance improvement: %s is %.1f%% faster",
				metric.Name, improvement))
		}
	} else {
		// Set baseline for first measurement
		m.baselines[metric.Name] = metric.Duration
	}
}

// StartTimer returns a function to record timing for an operation
func (m *Monitor) StartTimer(metricType MetricType, name string, tags map[string]string) func(success bool, err error) {
	start := time.Now()

	return func(success bool, err error) {
		duration := time.Since(start)

		metric := Metric{
			Type:     metricType,
			Name:     name,
			Duration: duration,
			Tags:     tags,
			Success:  success,
		}

		if err != nil {
			metric.Error = err.Error()
		}

		m.RecordMetric(metric)
	}
}

// GetSummary returns aggregated performance metrics
func (m *Monitor) GetSummary() MetricsSummary {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.metrics) == 0 {
		return MetricsSummary{
			OperationsByType: make(map[MetricType]int64),
			PerformanceGains: make(map[string]float64),
			RecentOperations: []Metric{},
		}
	}

	summary := MetricsSummary{
		OperationsByType: make(map[MetricType]int64),
		PerformanceGains: make(map[string]float64),
	}

	var totalDuration time.Duration
	var cacheHits, cacheMisses int64
	minDuration := m.metrics[0].Duration
	maxDuration := m.metrics[0].Duration

	// Analyze all metrics
	for _, metric := range m.metrics {
		summary.TotalOperations++
		if metric.Success {
			summary.SuccessfulOps++
		} else {
			summary.FailedOps++
		}

		totalDuration += metric.Duration
		summary.OperationsByType[metric.Type]++

		if metric.Duration < minDuration {
			minDuration = metric.Duration
		}
		if metric.Duration > maxDuration {
			maxDuration = metric.Duration
		}

		// Track cache performance
		if metric.Type == MetricCacheHit {
			cacheHits++
		} else if metric.Type == MetricCacheMiss {
			cacheMisses++
		}
	}

	// Calculate derived metrics
	if summary.TotalOperations > 0 {
		summary.AverageDuration = totalDuration / time.Duration(summary.TotalOperations)
		summary.MinDuration = minDuration
		summary.MaxDuration = maxDuration
	}

	if cacheHits+cacheMisses > 0 {
		summary.CacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses) * 100
	}

	// Calculate performance gains against baselines
	for name, baseline := range m.baselines {
		// Find recent average for this operation
		var recentDuration time.Duration
		var recentCount int
		for i := len(m.metrics) - 1; i >= 0 && recentCount < 10; i-- {
			if m.metrics[i].Name == name {
				recentDuration += m.metrics[i].Duration
				recentCount++
			}
		}

		if recentCount > 0 {
			avgRecent := recentDuration / time.Duration(recentCount)
			improvement := float64(baseline-avgRecent) / float64(baseline) * 100
			summary.PerformanceGains[name] = improvement
		}
	}

	// Get recent operations (last 10)
	start := len(m.metrics) - 10
	if start < 0 {
		start = 0
	}
	summary.RecentOperations = make([]Metric, len(m.metrics)-start)
	copy(summary.RecentOperations, m.metrics[start:])

	return summary
}

// GetCachePerformance returns cache-specific performance metrics
func (m *Monitor) GetCachePerformance(cacheStats *cache.CombinedCacheStats) CachePerformance {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	perf := CachePerformance{
		HitRate:         0,
		MissRate:        0,
		AverageHitTime:  0,
		AverageMissTime: 0,
	}

	if cacheStats == nil {
		return perf
	}

	// Calculate hit rate from cache stats
	totalRequests := cacheStats.Memory.Hits + cacheStats.Memory.Misses
	totalRequests += cacheStats.Disk.Hits + cacheStats.Disk.Misses

	if totalRequests > 0 {
		totalHits := cacheStats.Memory.Hits + cacheStats.Disk.Hits
		perf.HitRate = float64(totalHits) / float64(totalRequests) * 100
		perf.MissRate = 100 - perf.HitRate
	}

	// Calculate average response times from our metrics
	var hitDuration, missDuration time.Duration
	var hitCount, missCount int

	for _, metric := range m.metrics {
		if metric.Type == MetricCacheHit {
			hitDuration += metric.Duration
			hitCount++
		} else if metric.Type == MetricCacheMiss {
			missDuration += metric.Duration
			missCount++
		}
	}

	if hitCount > 0 {
		perf.AverageHitTime = hitDuration / time.Duration(hitCount)
	}
	if missCount > 0 {
		perf.AverageMissTime = missDuration / time.Duration(missCount)
	}

	return perf
}

// CachePerformance represents cache-specific performance metrics
type CachePerformance struct {
	HitRate         float64
	MissRate        float64
	AverageHitTime  time.Duration
	AverageMissTime time.Duration
}

// FormatSummary returns a human-readable performance summary
func (m *Monitor) FormatSummary() string {
	summary := m.GetSummary()

	result := fmt.Sprintf("📊 Performance Summary\n")
	result += fmt.Sprintf("=====================\n")
	result += fmt.Sprintf("Total Operations: %d\n", summary.TotalOperations)
	result += fmt.Sprintf("Success Rate: %.1f%%\n",
		float64(summary.SuccessfulOps)/float64(summary.TotalOperations)*100)
	result += fmt.Sprintf("Average Duration: %v\n", summary.AverageDuration)
	result += fmt.Sprintf("Cache Hit Rate: %.1f%%\n", summary.CacheHitRate)

	if len(summary.PerformanceGains) > 0 {
		result += fmt.Sprintf("\n🚀 Performance Improvements:\n")
		for name, gain := range summary.PerformanceGains {
			if gain > 0 {
				result += fmt.Sprintf("  %s: %.1f%% faster\n", name, gain)
			}
		}
	}

	result += fmt.Sprintf("\n📈 Operations by Type:\n")
	for metricType, count := range summary.OperationsByType {
		result += fmt.Sprintf("  %s: %d\n", metricType, count)
	}

	return result
}

// Reset clears all metrics (useful for testing)
func (m *Monitor) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics = m.metrics[:0]
	m.baselines = make(map[string]time.Duration)
}
