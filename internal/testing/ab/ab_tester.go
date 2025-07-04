package ab

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"nix-ai-help/internal/testing"
	"nix-ai-help/pkg/logger"
)

// ABTest represents an A/B test configuration
type ABTest struct {
	ID                string                      `json:"id"`
	Name              string                      `json:"name"`
	Description       string                      `json:"description"`
	ConfigA           *testing.TestConfiguration `json:"config_a"`
	ConfigB           *testing.TestConfiguration `json:"config_b"`
	TestParameters    TestParameters              `json:"test_parameters"`
	Status            ABTestStatus                `json:"status"`
	Results           *ABTestResults              `json:"results,omitempty"`
	CreatedAt         time.Time                   `json:"created_at"`
	StartedAt         *time.Time                  `json:"started_at,omitempty"`
	CompletedAt       *time.Time                  `json:"completed_at,omitempty"`
	EnvironmentAID    string                      `json:"environment_a_id"`
	EnvironmentBID    string                      `json:"environment_b_id"`
}

// ABTestStatus represents the status of an A/B test
type ABTestStatus string

const (
	ABTestStatusPending   ABTestStatus = "pending"
	ABTestStatusRunning   ABTestStatus = "running"
	ABTestStatusCompleted ABTestStatus = "completed"
	ABTestStatusFailed    ABTestStatus = "failed"
	ABTestStatusCancelled ABTestStatus = "cancelled"
)

// TestParameters defines parameters for A/B testing
type TestParameters struct {
	Duration         time.Duration          `json:"duration"`
	SampleSize       int                    `json:"sample_size"`
	Metrics          []string               `json:"metrics"`
	SuccessCriteria  []SuccessCriterion     `json:"success_criteria"`
	LoadProfile      LoadProfile            `json:"load_profile"`
	ConfidenceLevel  float64                `json:"confidence_level"`
	MinimumEffect    float64                `json:"minimum_effect"`
}

// SuccessCriterion defines what constitutes success in the A/B test
type SuccessCriterion struct {
	Metric    string  `json:"metric"`
	Operator  string  `json:"operator"` // "gt", "lt", "eq", "gte", "lte"
	Value     float64 `json:"value"`
	Weight    float64 `json:"weight"`
}

// LoadProfile defines the load testing profile
type LoadProfile struct {
	Pattern           string        `json:"pattern"` // "constant", "ramp", "spike", "wave"
	InitialLoad       int           `json:"initial_load"`
	MaxLoad           int           `json:"max_load"`
	RampDuration      time.Duration `json:"ramp_duration"`
	SustainDuration   time.Duration `json:"sustain_duration"`
	RequestsPerSecond int           `json:"requests_per_second"`
}

// ABTestResults contains the results of an A/B test
type ABTestResults struct {
	OverallWinner    string                     `json:"overall_winner"` // "A", "B", or "tie"
	Confidence       float64                    `json:"confidence"`
	StatisticalPower float64                    `json:"statistical_power"`
	EffectSize       float64                    `json:"effect_size"`
	MetricResults    map[string]*MetricResult   `json:"metric_results"`
	Performance      *PerformanceComparison     `json:"performance"`
	ResourceUsage    *ResourceComparison        `json:"resource_usage"`
	Summary          string                     `json:"summary"`
	Recommendations  []testing.Recommendation   `json:"recommendations"`
}

// MetricResult contains results for a specific metric
type MetricResult struct {
	ConfigA         MetricData `json:"config_a"`
	ConfigB         MetricData `json:"config_b"`
	Winner          string     `json:"winner"`
	Significance    float64    `json:"significance"`
	EffectSize      float64    `json:"effect_size"`
	ConfidenceLevel float64    `json:"confidence_level"`
}

// MetricData contains statistical data for a metric
type MetricData struct {
	Mean               float64   `json:"mean"`
	Median             float64   `json:"median"`
	StandardDeviation  float64   `json:"std_dev"`
	Min                float64   `json:"min"`
	Max                float64   `json:"max"`
	P95                float64   `json:"p95"`
	P99                float64   `json:"p99"`
	SampleSize         int       `json:"sample_size"`
	Values             []float64 `json:"values"`
}

// PerformanceComparison compares performance metrics
type PerformanceComparison struct {
	BootTime         MetricComparison `json:"boot_time"`
	ResponseTime     MetricComparison `json:"response_time"`
	Throughput       MetricComparison `json:"throughput"`
	ErrorRate        MetricComparison `json:"error_rate"`
	ServiceStartTime MetricComparison `json:"service_start_time"`
}

// ResourceComparison compares resource usage
type ResourceComparison struct {
	CPU     MetricComparison `json:"cpu"`
	Memory  MetricComparison `json:"memory"`
	Disk    MetricComparison `json:"disk"`
	Network MetricComparison `json:"network"`
}

// MetricComparison represents a comparison between two configurations
type MetricComparison struct {
	ConfigA        float64 `json:"config_a"`
	ConfigB        float64 `json:"config_b"`
	Difference     float64 `json:"difference"`
	PercentChange  float64 `json:"percent_change"`
	Winner         string  `json:"winner"`
	Significant    bool    `json:"significant"`
}

// ABTester manages A/B testing operations
type ABTester struct {
	logger      *logger.Logger
	tests       map[string]*ABTest
	mu          sync.RWMutex
	envManager  EnvironmentManagerInterface
	maxTests    int
}

// EnvironmentManagerInterface defines the interface for environment management
type EnvironmentManagerInterface interface {
	CreateEnvironment(ctx context.Context, config *testing.TestEnvironment) (*testing.TestEnvironment, error)
	GetEnvironment(ctx context.Context, id string) (*testing.TestEnvironment, error)
	DeleteEnvironment(ctx context.Context, id string) error
	ExecuteCommand(ctx context.Context, envID string, command []string) (string, error)
}

// NewABTester creates a new A/B tester
func NewABTester(envManager EnvironmentManagerInterface, maxTests int) *ABTester {
	return &ABTester{
		logger:     logger.NewLogger(),
		tests:      make(map[string]*ABTest),
		envManager: envManager,
		maxTests:   maxTests,
	}
}

// CreateTest creates a new A/B test
func (ab *ABTester) CreateTest(ctx context.Context, test *ABTest) (*ABTest, error) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	// Check test limits
	if len(ab.tests) >= ab.maxTests {
		return nil, fmt.Errorf("maximum number of tests (%d) reached", ab.maxTests)
	}

	// Generate unique ID if not provided
	if test.ID == "" {
		test.ID = fmt.Sprintf("abtest_%d", time.Now().Unix())
	}

	// Check if test already exists
	if _, exists := ab.tests[test.ID]; exists {
		return nil, fmt.Errorf("test with ID %s already exists", test.ID)
	}

	// Set defaults
	if test.TestParameters.Duration == 0 {
		test.TestParameters.Duration = 30 * time.Minute
	}
	if test.TestParameters.SampleSize == 0 {
		test.TestParameters.SampleSize = 100
	}
	if test.TestParameters.ConfidenceLevel == 0 {
		test.TestParameters.ConfidenceLevel = 0.95
	}

	test.Status = ABTestStatusPending
	test.CreatedAt = time.Now()

	ab.tests[test.ID] = test
	ab.logger.Info(fmt.Sprintf("Created A/B test %s", test.ID))
	return test, nil
}

// StartTest starts an A/B test
func (ab *ABTester) StartTest(ctx context.Context, testID string) error {
	ab.mu.Lock()
	test, exists := ab.tests[testID]
	ab.mu.Unlock()

	if !exists {
		return fmt.Errorf("test %s not found", testID)
	}

	if test.Status != ABTestStatusPending {
		return fmt.Errorf("test %s is not in pending status", testID)
	}

	// Update status
	ab.mu.Lock()
	test.Status = ABTestStatusRunning
	now := time.Now()
	test.StartedAt = &now
	ab.mu.Unlock()

	// Start test execution in background
	go ab.executeTest(ctx, test)

	ab.logger.Info(fmt.Sprintf("Started A/B test %s", testID))
	return nil
}

// executeTest executes the A/B test
func (ab *ABTester) executeTest(ctx context.Context, test *ABTest) {
	defer func() {
		if r := recover(); r != nil {
			ab.logger.Error(fmt.Sprintf("A/B test %s panic: %v", test.ID, r))
			ab.updateTestStatus(test.ID, ABTestStatusFailed)
		}
	}()

	ab.logger.Info(fmt.Sprintf("Executing A/B test %s", test.ID))

	// Create test environments
	envA, envB, err := ab.createTestEnvironments(ctx, test)
	if err != nil {
		ab.logger.Error(fmt.Sprintf("Failed to create environments for test %s: %v", test.ID, err))
		ab.updateTestStatus(test.ID, ABTestStatusFailed)
		return
	}

	test.EnvironmentAID = envA.ID
	test.EnvironmentBID = envB.ID

	// Wait for environments to be ready
	if err := ab.waitForEnvironments(ctx, envA.ID, envB.ID); err != nil {
		ab.logger.Error(fmt.Sprintf("Environments not ready for test %s: %v", test.ID, err))
		ab.updateTestStatus(test.ID, ABTestStatusFailed)
		return
	}

	// Run the test
	results, err := ab.runComparison(ctx, test, envA.ID, envB.ID)
	if err != nil {
		ab.logger.Error(fmt.Sprintf("Failed to run comparison for test %s: %v", test.ID, err))
		ab.updateTestStatus(test.ID, ABTestStatusFailed)
		return
	}

	// Store results
	ab.mu.Lock()
	test.Results = results
	test.Status = ABTestStatusCompleted
	now := time.Now()
	test.CompletedAt = &now
	ab.mu.Unlock()

	// Clean up environments
	ab.cleanupEnvironments(ctx, envA.ID, envB.ID)

	ab.logger.Info(fmt.Sprintf("A/B test %s completed successfully", test.ID))
}

// createTestEnvironments creates test environments for A/B testing
func (ab *ABTester) createTestEnvironments(ctx context.Context, test *ABTest) (*testing.TestEnvironment, *testing.TestEnvironment, error) {
	// Create environment A
	envAConfig := &testing.TestEnvironment{
		ID:            fmt.Sprintf("%s_env_a", test.ID),
		Name:          fmt.Sprintf("%s Environment A", test.Name),
		Configuration: test.ConfigA.Content,
		Resources: testing.ResourceAllocation{
			CPUCores: 2,
			MemoryMB: 2048,
			DiskGB:   10,
		},
	}

	envA, err := ab.envManager.CreateEnvironment(ctx, envAConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create environment A: %w", err)
	}

	// Create environment B
	envBConfig := &testing.TestEnvironment{
		ID:            fmt.Sprintf("%s_env_b", test.ID),
		Name:          fmt.Sprintf("%s Environment B", test.Name),
		Configuration: test.ConfigB.Content,
		Resources: testing.ResourceAllocation{
			CPUCores: 2,
			MemoryMB: 2048,
			DiskGB:   10,
		},
	}

	envB, err := ab.envManager.CreateEnvironment(ctx, envBConfig)
	if err != nil {
		// Clean up environment A on failure
		ab.envManager.DeleteEnvironment(ctx, envA.ID)
		return nil, nil, fmt.Errorf("failed to create environment B: %w", err)
	}

	return envA, envB, nil
}

// waitForEnvironments waits for both environments to be ready
func (ab *ABTester) waitForEnvironments(ctx context.Context, envAID, envBID string) error {
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for environments to be ready")
		case <-ticker.C:
			envA, err := ab.envManager.GetEnvironment(ctx, envAID)
			if err != nil {
				continue
			}
			envB, err := ab.envManager.GetEnvironment(ctx, envBID)
			if err != nil {
				continue
			}

			if envA.Status == testing.StatusRunning && envB.Status == testing.StatusRunning {
				return nil
			}
		}
	}
}

// runComparison runs the actual comparison between configurations
func (ab *ABTester) runComparison(ctx context.Context, test *ABTest, envAID, envBID string) (*ABTestResults, error) {
	results := &ABTestResults{
		MetricResults: make(map[string]*MetricResult),
		Performance:   &PerformanceComparison{},
		ResourceUsage: &ResourceComparison{},
	}

	// Run performance benchmarks
	perfA, err := ab.runPerformanceBenchmark(ctx, envAID, test.TestParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to run performance benchmark for config A: %w", err)
	}

	perfB, err := ab.runPerformanceBenchmark(ctx, envBID, test.TestParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to run performance benchmark for config B: %w", err)
	}

	// Analyze results
	ab.analyzePerformanceResults(results, perfA, perfB)
	ab.analyzeResourceUsage(results, perfA, perfB)
	ab.determineOverallWinner(results, test.TestParameters.SuccessCriteria)
	ab.generateRecommendations(results, test)

	return results, nil
}

// runPerformanceBenchmark runs performance benchmarks on an environment
func (ab *ABTester) runPerformanceBenchmark(ctx context.Context, envID string, params TestParameters) (*BenchmarkResults, error) {
	results := &BenchmarkResults{
		BootTime:      ab.measureBootTime(ctx, envID),
		ResponseTimes: ab.measureResponseTimes(ctx, envID, params),
		ResourceUsage: ab.measureResourceUsage(ctx, envID, params.Duration),
		Throughput:    ab.measureThroughput(ctx, envID, params),
		ErrorRate:     ab.measureErrorRate(ctx, envID, params),
	}

	return results, nil
}

// BenchmarkResults contains benchmark results for an environment
type BenchmarkResults struct {
	BootTime      time.Duration
	ResponseTimes []float64
	ResourceUsage ResourceMetrics
	Throughput    float64
	ErrorRate     float64
}

// ResourceMetrics contains resource usage metrics
type ResourceMetrics struct {
	CPU     []float64
	Memory  []float64
	Disk    []float64
	Network []float64
}

// measureBootTime measures the boot time of the environment
func (ab *ABTester) measureBootTime(ctx context.Context, envID string) time.Duration {
	// Simplified boot time measurement
	start := time.Now()
	
	// Check if system is responsive
	_, err := ab.envManager.ExecuteCommand(ctx, envID, []string{"echo", "ready"})
	if err != nil {
		return 0
	}
	
	return time.Since(start)
}

// measureResponseTimes measures response times for various operations
func (ab *ABTester) measureResponseTimes(ctx context.Context, envID string, params TestParameters) []float64 {
	var responseTimes []float64
	
	for i := 0; i < params.SampleSize; i++ {
		start := time.Now()
		_, err := ab.envManager.ExecuteCommand(ctx, envID, []string{"systemctl", "status"})
		duration := time.Since(start)
		
		if err == nil {
			responseTimes = append(responseTimes, float64(duration.Milliseconds()))
		}
	}
	
	return responseTimes
}

// measureResourceUsage measures resource usage over time
func (ab *ABTester) measureResourceUsage(ctx context.Context, envID string, duration time.Duration) ResourceMetrics {
	metrics := ResourceMetrics{
		CPU:     []float64{},
		Memory:  []float64{},
		Disk:    []float64{},
		Network: []float64{},
	}

	// Sample resource usage every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	timeout := time.After(duration)

	for {
		select {
		case <-ctx.Done():
			return metrics
		case <-timeout:
			return metrics
		case <-ticker.C:
			// Get current resource usage
			env, err := ab.envManager.GetEnvironment(ctx, envID)
			if err == nil && env.Metrics != nil {
				metrics.CPU = append(metrics.CPU, env.Metrics.CPUUsage)
				metrics.Memory = append(metrics.Memory, env.Metrics.MemoryUsage)
				metrics.Disk = append(metrics.Disk, env.Metrics.DiskUsage)
				metrics.Network = append(metrics.Network, env.Metrics.NetworkTraffic)
			}
		}
	}
}

// measureThroughput measures system throughput
func (ab *ABTester) measureThroughput(ctx context.Context, envID string, params TestParameters) float64 {
	// Simplified throughput measurement
	start := time.Now()
	operations := 0
	
	for time.Since(start) < time.Minute {
		_, err := ab.envManager.ExecuteCommand(ctx, envID, []string{"echo", "test"})
		if err == nil {
			operations++
		}
	}
	
	return float64(operations) / time.Since(start).Minutes()
}

// measureErrorRate measures the error rate of operations
func (ab *ABTester) measureErrorRate(ctx context.Context, envID string, params TestParameters) float64 {
	total := 0
	errors := 0
	
	for i := 0; i < params.SampleSize; i++ {
		total++
		_, err := ab.envManager.ExecuteCommand(ctx, envID, []string{"true"})
		if err != nil {
			errors++
		}
	}
	
	if total == 0 {
		return 0
	}
	return float64(errors) / float64(total) * 100
}

// analyzePerformanceResults analyzes performance comparison results
func (ab *ABTester) analyzePerformanceResults(results *ABTestResults, perfA, perfB *BenchmarkResults) {
	results.Performance.BootTime = ab.compareMetric(
		float64(perfA.BootTime.Milliseconds()),
		float64(perfB.BootTime.Milliseconds()),
		"lower_better",
	)

	results.Performance.ResponseTime = ab.compareMetric(
		ab.calculateMean(perfA.ResponseTimes),
		ab.calculateMean(perfB.ResponseTimes),
		"lower_better",
	)

	results.Performance.Throughput = ab.compareMetric(
		perfA.Throughput,
		perfB.Throughput,
		"higher_better",
	)

	results.Performance.ErrorRate = ab.compareMetric(
		perfA.ErrorRate,
		perfB.ErrorRate,
		"lower_better",
	)
}

// analyzeResourceUsage analyzes resource usage comparison
func (ab *ABTester) analyzeResourceUsage(results *ABTestResults, perfA, perfB *BenchmarkResults) {
	results.ResourceUsage.CPU = ab.compareMetric(
		ab.calculateMean(perfA.ResourceUsage.CPU),
		ab.calculateMean(perfB.ResourceUsage.CPU),
		"lower_better",
	)

	results.ResourceUsage.Memory = ab.compareMetric(
		ab.calculateMean(perfA.ResourceUsage.Memory),
		ab.calculateMean(perfB.ResourceUsage.Memory),
		"lower_better",
	)

	results.ResourceUsage.Disk = ab.compareMetric(
		ab.calculateMean(perfA.ResourceUsage.Disk),
		ab.calculateMean(perfB.ResourceUsage.Disk),
		"lower_better",
	)

	results.ResourceUsage.Network = ab.compareMetric(
		ab.calculateMean(perfA.ResourceUsage.Network),
		ab.calculateMean(perfB.ResourceUsage.Network),
		"lower_better",
	)
}

// compareMetric compares two metric values and returns comparison result
func (ab *ABTester) compareMetric(valueA, valueB float64, comparison string) MetricComparison {
	diff := valueB - valueA
	percentChange := 0.0
	if valueA != 0 {
		percentChange = (diff / valueA) * 100
	}

	var winner string
	significant := math.Abs(percentChange) > 5.0 // 5% significance threshold

	switch comparison {
	case "higher_better":
		if valueA > valueB {
			winner = "A"
		} else if valueB > valueA {
			winner = "B"
		} else {
			winner = "tie"
		}
	case "lower_better":
		if valueA < valueB {
			winner = "A"
		} else if valueB < valueA {
			winner = "B"
		} else {
			winner = "tie"
		}
	}

	return MetricComparison{
		ConfigA:       valueA,
		ConfigB:       valueB,
		Difference:    diff,
		PercentChange: percentChange,
		Winner:        winner,
		Significant:   significant,
	}
}

// determineOverallWinner determines the overall winner based on success criteria
func (ab *ABTester) determineOverallWinner(results *ABTestResults, criteria []SuccessCriterion) {
	scoreA := 0.0
	scoreB := 0.0
	totalWeight := 0.0

	// Calculate weighted scores based on success criteria
	for _, criterion := range criteria {
		var comparison MetricComparison
		
		switch criterion.Metric {
		case "boot_time":
			comparison = results.Performance.BootTime
		case "response_time":
			comparison = results.Performance.ResponseTime
		case "throughput":
			comparison = results.Performance.Throughput
		case "cpu_usage":
			comparison = results.ResourceUsage.CPU
		case "memory_usage":
			comparison = results.ResourceUsage.Memory
		default:
			continue
		}

		if comparison.Winner == "A" {
			scoreA += criterion.Weight
		} else if comparison.Winner == "B" {
			scoreB += criterion.Weight
		}
		totalWeight += criterion.Weight
	}

	// Determine overall winner
	if scoreA > scoreB {
		results.OverallWinner = "A"
		results.Confidence = scoreA / totalWeight
	} else if scoreB > scoreA {
		results.OverallWinner = "B"
		results.Confidence = scoreB / totalWeight
	} else {
		results.OverallWinner = "tie"
		results.Confidence = 0.5
	}

	// Calculate statistical power and effect size
	results.StatisticalPower = ab.calculateStatisticalPower(results)
	results.EffectSize = ab.calculateEffectSize(results)
}

// generateRecommendations generates recommendations based on test results
func (ab *ABTester) generateRecommendations(results *ABTestResults, test *ABTest) {
	var recommendations []testing.Recommendation

	// Performance recommendations
	if results.Performance.ResponseTime.Winner == "A" && results.Performance.ResponseTime.Significant {
		recommendations = append(recommendations, testing.Recommendation{
			ID:          "perf_response_time",
			Type:        testing.RecommendationPerformance,
			Priority:    "high",
			Title:       "Configuration A shows better response times",
			Description: fmt.Sprintf("Configuration A has %.1f%% better response times", math.Abs(results.Performance.ResponseTime.PercentChange)),
			Impact:      "Improved user experience and system responsiveness",
			Effort:      "low",
			Actions:     []string{"Deploy Configuration A for better performance"},
		})
	}

	// Resource usage recommendations
	if results.ResourceUsage.Memory.Winner == "A" && results.ResourceUsage.Memory.Significant {
		recommendations = append(recommendations, testing.Recommendation{
			ID:          "resource_memory",
			Type:        testing.RecommendationCost,
			Priority:    "medium",
			Title:       "Configuration A uses less memory",
			Description: fmt.Sprintf("Configuration A uses %.1f%% less memory", math.Abs(results.ResourceUsage.Memory.PercentChange)),
			Impact:      "Lower resource costs and better scalability",
			Effort:      "low",
			Actions:     []string{"Consider Configuration A for cost optimization"},
		})
	}

	results.Recommendations = recommendations
}

// Helper functions for statistical calculations
func (ab *ABTester) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (ab *ABTester) calculateStatisticalPower(results *ABTestResults) float64 {
	// Simplified statistical power calculation
	return 0.8 // 80% power assumption
}

func (ab *ABTester) calculateEffectSize(results *ABTestResults) float64 {
	// Simplified effect size calculation
	return 0.5 // Medium effect size assumption
}

// updateTestStatus updates the status of a test
func (ab *ABTester) updateTestStatus(testID string, status ABTestStatus) {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	if test, exists := ab.tests[testID]; exists {
		test.Status = status
	}
}

// cleanupEnvironments cleans up test environments
func (ab *ABTester) cleanupEnvironments(ctx context.Context, envAID, envBID string) {
	if err := ab.envManager.DeleteEnvironment(ctx, envAID); err != nil {
		ab.logger.Error(fmt.Sprintf("Failed to delete environment A (%s): %v", envAID, err))
	}
	if err := ab.envManager.DeleteEnvironment(ctx, envBID); err != nil {
		ab.logger.Error(fmt.Sprintf("Failed to delete environment B (%s): %v", envBID, err))
	}
}

// GetTest retrieves a test by ID
func (ab *ABTester) GetTest(ctx context.Context, testID string) (*ABTest, error) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	test, exists := ab.tests[testID]
	if !exists {
		return nil, fmt.Errorf("test %s not found", testID)
	}

	return test, nil
}

// ListTests lists all tests
func (ab *ABTester) ListTests(ctx context.Context) ([]*ABTest, error) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	tests := make([]*ABTest, 0, len(ab.tests))
	for _, test := range ab.tests {
		tests = append(tests, test)
	}

	// Sort by creation time
	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt.After(tests[j].CreatedAt)
	})

	return tests, nil
}

// CancelTest cancels a running test
func (ab *ABTester) CancelTest(ctx context.Context, testID string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	test, exists := ab.tests[testID]
	if !exists {
		return fmt.Errorf("test %s not found", testID)
	}

	if test.Status != ABTestStatusRunning {
		return fmt.Errorf("test %s is not running", testID)
	}

	test.Status = ABTestStatusCancelled
	
	// Clean up environments if they exist
	if test.EnvironmentAID != "" {
		go ab.envManager.DeleteEnvironment(ctx, test.EnvironmentAID)
	}
	if test.EnvironmentBID != "" {
		go ab.envManager.DeleteEnvironment(ctx, test.EnvironmentBID)
	}

	ab.logger.Info(fmt.Sprintf("Cancelled A/B test %s", testID))
	return nil
}