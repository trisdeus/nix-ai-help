package performance

import (
	"testing"
	"time"

	"nix-ai-help/internal/cache"
	"nix-ai-help/pkg/logger"
)

func TestPerformanceMonitor(t *testing.T) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	// Test recording metrics
	metric := Metric{
		Type:     MetricAIQuery,
		Name:     "test_query",
		Duration: 100 * time.Millisecond,
		Tags:     map[string]string{"provider": "test"},
		Success:  true,
	}

	monitor.RecordMetric(metric)

	// Test getting summary
	summary := monitor.GetSummary()
	if summary.TotalOperations != 1 {
		t.Errorf("Expected 1 operation, got %d", summary.TotalOperations)
	}

	if summary.SuccessfulOps != 1 {
		t.Errorf("Expected 1 successful operation, got %d", summary.SuccessfulOps)
	}

	// Test timer functionality
	finishTimer := monitor.StartTimer(MetricCacheHit, "test_cache", nil)
	time.Sleep(10 * time.Millisecond)
	finishTimer(true, nil)

	summary = monitor.GetSummary()
	if summary.TotalOperations != 2 {
		t.Errorf("Expected 2 operations, got %d", summary.TotalOperations)
	}
}

func TestCachePerformance(t *testing.T) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	// Simulate cache hits and misses
	hitMetric := Metric{
		Type:     MetricCacheHit,
		Name:     "cache_test",
		Duration: 5 * time.Millisecond,
		Success:  true,
	}

	missMetric := Metric{
		Type:     MetricCacheMiss,
		Name:     "cache_test",
		Duration: 50 * time.Millisecond,
		Success:  true,
	}

	// Record 7 hits and 3 misses
	for i := 0; i < 7; i++ {
		monitor.RecordMetric(hitMetric)
	}
	for i := 0; i < 3; i++ {
		monitor.RecordMetric(missMetric)
	}

	// Create mock cache stats
	cacheStats := &cache.CombinedCacheStats{
		Memory: cache.CacheStats{
			Hits:   7,
			Misses: 3,
		},
	}

	cachePerf := monitor.GetCachePerformance(cacheStats)
	expectedHitRate := 70.0 // 7 hits out of 10 total

	if cachePerf.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %.1f%%, got %.1f%%", expectedHitRate, cachePerf.HitRate)
	}

	if cachePerf.MissRate != 30.0 {
		t.Errorf("Expected miss rate 30.0%%, got %.1f%%", cachePerf.MissRate)
	}
}

func TestPerformanceGains(t *testing.T) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	operationName := "test_operation"

	// Record initial baseline (slow)
	baselineMetric := Metric{
		Type:     MetricAIQuery,
		Name:     operationName,
		Duration: 1000 * time.Millisecond,
		Success:  true,
	}
	monitor.RecordMetric(baselineMetric)

	// Record improved performance
	improvedMetric := Metric{
		Type:     MetricAIQuery,
		Name:     operationName,
		Duration: 200 * time.Millisecond,
		Success:  true,
	}
	monitor.RecordMetric(improvedMetric)

	summary := monitor.GetSummary()

	// Check if performance improvement is tracked
	if gain, exists := summary.PerformanceGains[operationName]; exists {
		if gain <= 0 {
			t.Errorf("Expected positive performance gain, got %.1f%%", gain)
		}
		t.Logf("Performance improvement: %.1f%%", gain)
	}
}

func TestFormatSummary(t *testing.T) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	// Add some test metrics
	metrics := []Metric{
		{Type: MetricAIQuery, Name: "ai_test", Duration: 100 * time.Millisecond, Success: true},
		{Type: MetricCacheHit, Name: "cache_test", Duration: 5 * time.Millisecond, Success: true},
		{Type: MetricDocumentationQuery, Name: "doc_test", Duration: 200 * time.Millisecond, Success: true},
	}

	for _, metric := range metrics {
		monitor.RecordMetric(metric)
	}

	formatted := monitor.FormatSummary()

	// Check that formatted output contains expected sections
	if formatted == "" {
		t.Error("Formatted summary should not be empty")
	}

	// Should contain performance summary header
	if !contains(formatted, "Performance Summary") {
		t.Error("Formatted summary should contain 'Performance Summary'")
	}

	// Should contain operations by type
	if !contains(formatted, "Operations by Type") {
		t.Error("Formatted summary should contain 'Operations by Type'")
	}

	t.Logf("Formatted summary:\n%s", formatted)
}

func TestReset(t *testing.T) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	// Add some metrics
	metric := Metric{
		Type:     MetricAIQuery,
		Name:     "test",
		Duration: 100 * time.Millisecond,
		Success:  true,
	}
	monitor.RecordMetric(metric)

	summary := monitor.GetSummary()
	if summary.TotalOperations == 0 {
		t.Error("Should have operations before reset")
	}

	// Reset and verify
	monitor.Reset()
	summary = monitor.GetSummary()

	if summary.TotalOperations != 0 {
		t.Errorf("Expected 0 operations after reset, got %d", summary.TotalOperations)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		found := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkRecordMetric(b *testing.B) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	metric := Metric{
		Type:     MetricAIQuery,
		Name:     "benchmark_test",
		Duration: 100 * time.Millisecond,
		Success:  true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordMetric(metric)
	}
}

func BenchmarkGetSummary(b *testing.B) {
	logger := logger.NewLogger()
	monitor := NewMonitor(logger)

	// Pre-populate with metrics
	for i := 0; i < 100; i++ {
		metric := Metric{
			Type:     MetricAIQuery,
			Name:     "benchmark_test",
			Duration: time.Duration(i) * time.Millisecond,
			Success:  true,
		}
		monitor.RecordMetric(metric)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.GetSummary()
	}
}
