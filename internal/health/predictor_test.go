// Package health provides tests for the health prediction system
package health

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/internal/config"
)

func TestNewSystemHealthPredictor(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	if predictor == nil {
		t.Fatal("Expected predictor to be created, got nil")
	}

	if predictor.config != cfg {
		t.Error("Expected config to be set correctly")
	}

	if predictor.healthConfig == nil {
		t.Error("Expected health config to be initialized")
	}

	if predictor.failurePredictionModel == nil {
		t.Error("Expected failure prediction model to be initialized")
	}

	if predictor.anomalyDetector == nil {
		t.Error("Expected anomaly detector to be initialized")
	}

	if predictor.resourceForecaster == nil {
		t.Error("Expected resource forecaster to be initialized")
	}

	if predictor.systemMonitor == nil {
		t.Error("Expected system monitor to be initialized")
	}

	if predictor.remediationEngine == nil {
		t.Error("Expected remediation engine to be initialized")
	}
}

func TestSystemHealthPredictorStartStop(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Test start
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}

	if !predictor.running {
		t.Error("Expected predictor to be running after start")
	}

	// Test double start
	if err := predictor.Start(ctx); err == nil {
		t.Error("Expected error when starting already running predictor")
	}

	// Test stop
	if err := predictor.Stop(); err != nil {
		t.Fatalf("Failed to stop predictor: %v", err)
	}

	if predictor.running {
		t.Error("Expected predictor to be stopped after stop")
	}

	// Test double stop
	if err := predictor.Stop(); err != nil {
		t.Error("Expected no error when stopping already stopped predictor")
	}
}

func TestPredictFailures(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Test prediction
	timeline := 7 * 24 * time.Hour
	prediction, err := predictor.PredictFailures(ctx, timeline)
	if err != nil {
		t.Fatalf("Failed to predict failures: %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected prediction to be returned, got nil")
	}

	if prediction.Timeline != timeline {
		t.Errorf("Expected timeline %v, got %v", timeline, prediction.Timeline)
	}

	if prediction.GeneratedAt.IsZero() {
		t.Error("Expected GeneratedAt to be set")
	}

	if prediction.Confidence < 0 || prediction.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", prediction.Confidence)
	}

	// Test caching
	prediction2, err := predictor.PredictFailures(ctx, timeline)
	if err != nil {
		t.Fatalf("Failed to get cached prediction: %v", err)
	}

	if prediction2.GeneratedAt != prediction.GeneratedAt {
		t.Error("Expected cached prediction to have same generation time")
	}
}

func TestAnalyzeSystemHealth(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Test health analysis
	assessment, err := predictor.AnalyzeSystemHealth(ctx)
	if err != nil {
		t.Fatalf("Failed to analyze system health: %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected assessment to be returned, got nil")
	}

	if assessment.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}

	if assessment.ComponentHealth == nil {
		t.Error("Expected ComponentHealth to be initialized")
	}

	if assessment.ActiveIssues == nil {
		t.Error("Expected ActiveIssues to be initialized")
	}

	if assessment.PerformanceMetrics == nil {
		t.Error("Expected PerformanceMetrics to be initialized")
	}

	// Verify component health includes expected components
	expectedComponents := []string{"cpu", "memory", "disk", "network", "system"}
	for _, component := range expectedComponents {
		if _, exists := assessment.ComponentHealth[component]; !exists {
			t.Errorf("Expected component %s to be in health assessment", component)
		}
	}
}

func TestForecastResources(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Test resource forecasting
	timeline := 3 * 24 * time.Hour
	forecast, err := predictor.ForecastResources(ctx, timeline)
	if err != nil {
		t.Fatalf("Failed to forecast resources: %v", err)
	}

	if forecast == nil {
		t.Fatal("Expected forecast to be returned, got nil")
	}

	if forecast.Timeline != timeline {
		t.Errorf("Expected timeline %v, got %v", timeline, forecast.Timeline)
	}

	if forecast.GeneratedAt.IsZero() {
		t.Error("Expected GeneratedAt to be set")
	}

	if forecast.ModelAccuracy < 0 || forecast.ModelAccuracy > 1 {
		t.Errorf("Expected model accuracy between 0 and 1, got %f", forecast.ModelAccuracy)
	}

	if forecast.Confidence < 0 || forecast.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", forecast.Confidence)
	}

	// Verify forecast includes expected resources
	expectedResources := []string{"cpu", "memory", "disk", "network"}
	for _, resource := range expectedResources {
		if _, exists := forecast.Predictions[resource]; !exists {
			t.Errorf("Expected resource %s to be in forecast", resource)
		}
	}
}

func TestDetectAnomalies(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Test anomaly detection
	report, err := predictor.DetectAnomalies(ctx)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	if report == nil {
		t.Fatal("Expected report to be returned, got nil")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Expected GeneratedAt to be set")
	}

	if report.DetectedAnomalies == nil {
		t.Error("Expected DetectedAnomalies to be initialized")
	}

	if report.AnomalyScore < 0 {
		t.Errorf("Expected non-negative anomaly score, got %f", report.AnomalyScore)
	}
}

func TestGetRemediationSuggestions(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Create mock health issues
	issues := []HealthIssue{
		{
			ID:          "test_issue_1",
			Type:        "anomaly",
			Component:   "disk",
			Description: "High disk usage detected",
			Severity:    PriorityHigh,
			Status:      "active",
			DetectedAt:  time.Now(),
		},
		{
			ID:          "test_issue_2",
			Type:        "anomaly",
			Component:   "memory",
			Description: "Memory leak suspected",
			Severity:    PriorityMedium,
			Status:      "active",
			DetectedAt:  time.Now(),
		},
	}

	// Test remediation suggestions
	plan, err := predictor.GetRemediationSuggestions(ctx, issues)
	if err != nil {
		t.Fatalf("Failed to get remediation suggestions: %v", err)
	}

	if plan == nil {
		t.Fatal("Expected plan to be returned, got nil")
	}

	if len(plan.Issues) != len(issues) {
		t.Errorf("Expected %d issues in plan, got %d", len(issues), len(plan.Issues))
	}

	if len(plan.Suggestions) == 0 {
		t.Error("Expected at least one suggestion in plan")
	}

	if plan.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	// Verify suggestions have required fields
	for i, suggestion := range plan.Suggestions {
		if suggestion.ID == "" {
			t.Errorf("Suggestion %d missing ID", i)
		}
		if suggestion.Title == "" {
			t.Errorf("Suggestion %d missing Title", i)
		}
		if suggestion.Description == "" {
			t.Errorf("Suggestion %d missing Description", i)
		}
	}
}

func TestGetHealthStatus(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Test initial health status
	status := predictor.GetHealthStatus()
	if status == nil {
		t.Fatal("Expected status to be returned, got nil")
	}

	if status.OverallHealth != HealthUnknown {
		t.Errorf("Expected initial health to be unknown, got %s", status.OverallHealth)
	}
}

func TestGetModelInfo(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Test model info
	info := predictor.GetModelInfo()
	if info == nil {
		t.Fatal("Expected info to be returned, got nil")
	}

	if _, exists := info["failure_prediction"]; !exists {
		t.Error("Expected failure_prediction model info")
	}

	failureInfo := info["failure_prediction"]
	if failureInfo.Name == "" {
		t.Error("Expected model name to be set")
	}
}

func TestUpdateModels(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor to initialize models
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Add some mock events to event history
	mockEvents := []HealthEvent{
		{
			ID:        "test_event_1",
			Type:      "metric_collection",
			Component: "system",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Severity:  PriorityLow,
			Message:   "Test event",
			Metrics:   map[string]interface{}{"cpu_usage": 70.0},
		},
		{
			ID:        "test_event_2",
			Type:      "failure",
			Component: "disk",
			Timestamp: time.Now().Add(-30 * time.Minute),
			Severity:  PriorityHigh,
			Message:   "Disk failure event",
			Metrics:   map[string]interface{}{"disk_usage": 95.0},
		},
	}

	predictor.mu.Lock()
	predictor.eventHistory = append(predictor.eventHistory, mockEvents...)
	predictor.mu.Unlock()

	// Test model update
	if err := predictor.UpdateModels(ctx); err != nil {
		t.Fatalf("Failed to update models: %v", err)
	}
}

func TestCalculateComponentHealth(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Test with normal metrics
	normalMetrics := map[string]float64{
		"cpu_usage":     50.0,
		"memory_usage":  60.0,
		"disk_usage":    40.0,
		"network_usage": 30.0,
		"load_average":  2.0,
		"process_count": 150.0,
	}

	componentHealth := predictor.calculateComponentHealth(normalMetrics)
	
	expectedComponents := []string{"cpu", "memory", "disk", "network", "system"}
	for _, component := range expectedComponents {
		if _, exists := componentHealth[component]; !exists {
			t.Errorf("Expected component %s in health results", component)
		}
		
		health := componentHealth[component]
		if health == HealthUnknown {
			t.Errorf("Expected valid health status for %s, got unknown", component)
		}
	}

	// Test with critical metrics
	criticalMetrics := map[string]float64{
		"cpu_usage":     95.0,
		"memory_usage":  98.0,
		"disk_usage":    92.0,
		"network_usage": 90.0,
		"load_average":  10.0,
		"process_count": 400.0,
	}

	criticalHealth := predictor.calculateComponentHealth(criticalMetrics)
	
	// At least some components should be in critical state
	hasCritical := false
	for _, health := range criticalHealth {
		if health == HealthCritical || health == HealthPoor {
			hasCritical = true
			break
		}
	}
	
	if !hasCritical {
		t.Error("Expected at least one component to be in critical/poor health with high metrics")
	}
}

func TestGenerateHealthIssues(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Create mock anomalies
	anomalies := []Anomaly{
		{
			ID:          "anomaly_1",
			Type:        "resource_spike",
			Component:   "cpu",
			Description: "CPU usage spike detected",
			Score:       0.9,
			Severity:    PriorityHigh,
			DetectedAt:  time.Now(),
			Status:      "active",
		},
		{
			ID:          "anomaly_2",
			Type:        "memory_leak",
			Component:   "memory",
			Description: "Potential memory leak",
			Score:       0.7,
			Severity:    PriorityMedium,
			DetectedAt:  time.Now(),
			Status:      "active",
		},
	}

	issues := predictor.generateHealthIssues(anomalies)

	if len(issues) != len(anomalies) {
		t.Errorf("Expected %d issues, got %d", len(anomalies), len(issues))
	}

	for i, issue := range issues {
		if issue.ID == "" {
			t.Errorf("Issue %d missing ID", i)
		}
		if issue.Component != anomalies[i].Component {
			t.Errorf("Issue %d component mismatch: expected %s, got %s", 
				i, anomalies[i].Component, issue.Component)
		}
		if issue.Severity != anomalies[i].Severity {
			t.Errorf("Issue %d severity mismatch: expected %s, got %s", 
				i, anomalies[i].Severity, issue.Severity)
		}
		if len(issue.Indicators) == 0 {
			t.Errorf("Issue %d missing indicators", i)
		}
		if len(issue.Suggestions) == 0 {
			t.Errorf("Issue %d missing suggestions", i)
		}
	}
}

func TestCalculateResourceUtilization(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	metrics := map[string]float64{
		"cpu_usage":     70.0,
		"memory_usage":  80.0,
		"disk_usage":    60.0,
		"network_usage": 40.0,
		"load_average":  3.0,
		"process_count": 200.0,
	}

	utilization := predictor.calculateResourceUtilization(metrics)

	// Test CPU
	if utilization.CPU.Current != 70.0 {
		t.Errorf("Expected CPU current to be 70.0, got %f", utilization.CPU.Current)
	}
	if utilization.CPU.Unit != "%" {
		t.Errorf("Expected CPU unit to be %%, got %s", utilization.CPU.Unit)
	}

	// Test Memory
	if utilization.Memory.Current != 80.0 {
		t.Errorf("Expected Memory current to be 80.0, got %f", utilization.Memory.Current)
	}

	// Test Disk
	if utilization.Disk.Current != 60.0 {
		t.Errorf("Expected Disk current to be 60.0, got %f", utilization.Disk.Current)
	}

	// Test Network
	if utilization.Network.Current != 40.0 {
		t.Errorf("Expected Network current to be 40.0, got %f", utilization.Network.Current)
	}

	// Test LoadAvg
	if len(utilization.LoadAvg) != 3 {
		t.Errorf("Expected LoadAvg to have 3 values, got %d", len(utilization.LoadAvg))
	}

	// Test Processes
	if utilization.Processes != 200 {
		t.Errorf("Expected Processes to be 200, got %d", utilization.Processes)
	}

	// Test timestamps
	if utilization.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestGenerateRecommendations(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Create mock issues
	issues := []HealthIssue{
		{
			ID:        "issue_1",
			Type:      "anomaly",
			Component: "cpu",
			Severity:  PriorityHigh,
		},
	}

	// Create metrics with high disk usage to trigger proactive recommendation
	metrics := map[string]float64{
		"cpu_usage":     50.0,
		"memory_usage":  60.0,
		"disk_usage":    80.0, // Above 75% threshold
		"network_usage": 30.0,
	}

	recommendations := predictor.generateRecommendations(issues, metrics)

	if len(recommendations) == 0 {
		t.Error("Expected at least one recommendation")
	}

	// Should have recommendation for the issue
	hasIssueRec := false
	for _, rec := range recommendations {
		if rec.Type == "issue_resolution" {
			hasIssueRec = true
			break
		}
	}
	if !hasIssueRec {
		t.Error("Expected issue resolution recommendation")
	}

	// Should have proactive disk cleanup recommendation
	hasDiskRec := false
	for _, rec := range recommendations {
		if rec.ID == "rec_disk_cleanup" {
			hasDiskRec = true
			break
		}
	}
	if !hasDiskRec {
		t.Error("Expected disk cleanup recommendation for high disk usage")
	}
}

func TestGetMetricThreshold(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Test configured threshold
	cpuThreshold := predictor.getMetricThreshold("cpu")
	if cpuThreshold.Warning != 75.0 {
		t.Errorf("Expected CPU warning threshold 75.0, got %f", cpuThreshold.Warning)
	}
	if cpuThreshold.Critical != 90.0 {
		t.Errorf("Expected CPU critical threshold 90.0, got %f", cpuThreshold.Critical)
	}

	// Test default threshold for unknown metric
	unknownThreshold := predictor.getMetricThreshold("unknown_metric")
	if unknownThreshold.Warning != 75.0 {
		t.Errorf("Expected default warning threshold 75.0, got %f", unknownThreshold.Warning)
	}
	if unknownThreshold.Critical != 90.0 {
		t.Errorf("Expected default critical threshold 90.0, got %f", unknownThreshold.Critical)
	}
}

func TestGetMetricStatus(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	threshold := 75.0

	// Test good status
	if status := predictor.getMetricStatus(50.0, threshold); status != "good" {
		t.Errorf("Expected 'good' status for value 50.0, got '%s'", status)
	}

	// Test normal status
	if status := predictor.getMetricStatus(65.0, threshold); status != "normal" {
		t.Errorf("Expected 'normal' status for value 65.0, got '%s'", status)
	}

	// Test warning status
	if status := predictor.getMetricStatus(80.0, threshold); status != "warning" {
		t.Errorf("Expected 'warning' status for value 80.0, got '%s'", status)
	}

	// Test critical status
	if status := predictor.getMetricStatus(95.0, threshold); status != "critical" {
		t.Errorf("Expected 'critical' status for value 95.0, got '%s'", status)
	}
}

func TestCleanupExpiredCache(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	// Add cached prediction
	prediction := &FailurePrediction{
		Timeline:    7 * 24 * time.Hour,
		GeneratedAt: time.Now().Add(-20 * time.Minute), // Older than default 10-minute expiry
	}
	predictor.predictionCache["test_key"] = prediction

	// Add recent prediction
	recentPrediction := &FailurePrediction{
		Timeline:    7 * 24 * time.Hour,
		GeneratedAt: time.Now(),
	}
	predictor.predictionCache["recent_key"] = recentPrediction

	// Run cleanup
	predictor.cleanupExpiredCache()

	// Expired cache should be removed
	if _, exists := predictor.predictionCache["test_key"]; exists {
		t.Error("Expected expired cache entry to be removed")
	}

	// Recent cache should remain
	if _, exists := predictor.predictionCache["recent_key"]; !exists {
		t.Error("Expected recent cache entry to remain")
	}
}

func TestGenerateMockHistoricalData(t *testing.T) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)

	events := predictor.generateMockHistoricalData()

	if len(events) == 0 {
		t.Error("Expected mock historical data to be generated")
	}

	// Verify event structure
	for i, event := range events {
		if event.ID == "" {
			t.Errorf("Event %d missing ID", i)
		}
		if event.Timestamp.IsZero() {
			t.Errorf("Event %d missing timestamp", i)
		}
		if event.Metrics == nil {
			t.Errorf("Event %d missing metrics", i)
		}

		// Check for expected metrics
		expectedMetrics := []string{"cpu_usage", "memory_usage", "disk_usage", "network_usage", "load_average", "process_count"}
		for _, metric := range expectedMetrics {
			if _, exists := event.Metrics[metric]; !exists {
				t.Errorf("Event %d missing metric %s", i, metric)
			}
		}
	}

	// Verify some events are failure events
	hasFailure := false
	for _, event := range events {
		if event.Type == "failure" {
			hasFailure = true
			break
		}
	}
	if !hasFailure {
		t.Error("Expected some mock events to be failure events")
	}
}

// Benchmark tests

func BenchmarkPredictFailures(b *testing.B) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	predictor.Start(ctx)
	defer predictor.Stop()

	timeline := 7 * 24 * time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := predictor.PredictFailures(ctx, timeline)
		if err != nil {
			b.Fatalf("Failed to predict failures: %v", err)
		}
	}
}

func BenchmarkAnalyzeSystemHealth(b *testing.B) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	predictor.Start(ctx)
	defer predictor.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := predictor.AnalyzeSystemHealth(ctx)
		if err != nil {
			b.Fatalf("Failed to analyze system health: %v", err)
		}
	}
}

func BenchmarkDetectAnomalies(b *testing.B) {
	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	predictor.Start(ctx)
	defer predictor.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := predictor.DetectAnomalies(ctx)
		if err != nil {
			b.Fatalf("Failed to detect anomalies: %v", err)
		}
	}
}

// Integration tests

func TestHealthPredictionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.UserConfig{}
	predictor := NewSystemHealthPredictor(cfg)
	ctx := context.Background()

	// Start predictor
	if err := predictor.Start(ctx); err != nil {
		t.Fatalf("Failed to start predictor: %v", err)
	}
	defer predictor.Stop()

	// Wait for system monitor to collect some data
	time.Sleep(2 * time.Second)

	// Test full workflow
	t.Run("HealthAnalysis", func(t *testing.T) {
		assessment, err := predictor.AnalyzeSystemHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to analyze health: %v", err)
		}
		if assessment == nil {
			t.Fatal("Expected health assessment")
		}
	})

	t.Run("FailurePrediction", func(t *testing.T) {
		prediction, err := predictor.PredictFailures(ctx, 7*24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to predict failures: %v", err)
		}
		if prediction == nil {
			t.Fatal("Expected failure prediction")
		}
	})

	t.Run("AnomalyDetection", func(t *testing.T) {
		report, err := predictor.DetectAnomalies(ctx)
		if err != nil {
			t.Fatalf("Failed to detect anomalies: %v", err)
		}
		if report == nil {
			t.Fatal("Expected anomaly report")
		}
	})

	t.Run("ResourceForecasting", func(t *testing.T) {
		forecast, err := predictor.ForecastResources(ctx, 3*24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to forecast resources: %v", err)
		}
		if forecast == nil {
			t.Fatal("Expected resource forecast")
		}
	})

	t.Run("RemediationSuggestions", func(t *testing.T) {
		// Get current issues
		assessment, err := predictor.AnalyzeSystemHealth(ctx)
		if err != nil {
			t.Fatalf("Failed to analyze health: %v", err)
		}

		// Generate remediation plan (even if no issues)
		plan, err := predictor.GetRemediationSuggestions(ctx, assessment.ActiveIssues)
		if err != nil {
			t.Fatalf("Failed to get remediation suggestions: %v", err)
		}
		if plan == nil {
			t.Fatal("Expected remediation plan")
		}
	})
}