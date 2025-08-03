// Package ml provides tests for ML components
package ml

import (
	"testing"
	"time"

	"nix-ai-help/pkg/logger"
)

func TestNewFeatureExtractor(t *testing.T) {
	extractor := NewFeatureExtractor()

	if extractor == nil {
		t.Fatal("Expected feature extractor to be created, got nil")
	}

	if extractor.featureNames == nil {
		t.Error("Expected feature names to be initialized")
	}

	if extractor.scalingFactors == nil {
		t.Error("Expected scaling factors to be initialized")
	}

	if extractor.featureWindows == nil {
		t.Error("Expected feature windows to be initialized")
	}

	if extractor.aggregations == nil {
		t.Error("Expected aggregations to be initialized")
	}

	// Verify default features are included
	expectedFeatures := []string{"cpu_usage", "memory_usage", "disk_usage", "network_usage", "load_average", "error_rate", "process_count"}
	for _, expected := range expectedFeatures {
		found := false
		for _, feature := range extractor.featureNames {
			if feature == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected feature %s to be in default features", expected)
		}
	}
}

func TestFeatureExtractorExtractFeatures(t *testing.T) {
	extractor := NewFeatureExtractor()

	// Create test metrics
	metrics := map[string]float64{
		"cpu_usage":     70.0,
		"memory_usage":  80.0,
		"disk_usage":    60.0,
		"network_usage": 40.0,
		"load_average":  2.5,
		"error_rate":    0.05,
		"process_count": 150.0,
	}

	features, err := extractor.ExtractFeatures(metrics)
	if err != nil {
		t.Fatalf("Failed to extract features: %v", err)
	}

	if features == nil {
		t.Fatal("Expected features to be returned, got nil")
	}

	// Verify all base metrics are included
	for metric, value := range metrics {
		if extractedValue, exists := features[metric]; !exists {
			t.Errorf("Expected metric %s to be in extracted features", metric)
		} else if extractedValue != value {
			t.Errorf("Expected metric %s value %f, got %f", metric, value, extractedValue)
		}
	}

	// Verify derived features are calculated
	expectedDerived := []string{"cpu_memory_ratio", "disk_error_correlation", "load_per_process", "resource_pressure_index"}
	for _, derived := range expectedDerived {
		if _, exists := features[derived]; !exists {
			t.Errorf("Expected derived feature %s to be calculated", derived)
		}
	}

	// Verify features are scaled
	for metric := range metrics {
		if scaledValue, exists := features[metric+"_scaled"]; !exists {
			t.Errorf("Expected scaled feature %s_scaled to be calculated", metric)
		} else {
			// Scaled values should be different from original (unless scaling factor is 1.0)
			if scaledValue < 0 || scaledValue > 1 {
				t.Errorf("Expected scaled feature %s to be normalized between 0 and 1, got %f", metric, scaledValue)
			}
		}
	}
}

func TestFeatureExtractorNormalizeFeatures(t *testing.T) {
	extractor := NewFeatureExtractor()

	features := map[string]float64{
		"cpu_usage":    70.0,
		"memory_usage": 80.0,
		"disk_usage":   60.0,
	}

	normalized := extractor.normalizeFeatures(features)

	// Verify all features are normalized
	for feature := range features {
		if _, exists := normalized[feature]; !exists {
			t.Errorf("Expected feature %s to be in normalized features", feature)
		}
	}

	// Test with extreme values
	extremeFeatures := map[string]float64{
		"cpu_usage":    120.0,  // Above normal range
		"memory_usage": -10.0,  // Below normal range
		"disk_usage":   50.0,   // Normal range
	}

	extremeNormalized := extractor.normalizeFeatures(extremeFeatures)

	for feature, value := range extremeNormalized {
		if value < 0 || value > 1 {
			t.Errorf("Expected normalized feature %s to be between 0 and 1, got %f", feature, value)
		}
	}
}

func TestFeatureExtractorCalculateDerivedFeatures(t *testing.T) {
	extractor := NewFeatureExtractor()

	features := map[string]float64{
		"cpu_usage":     70.0,
		"memory_usage":  80.0,
		"disk_usage":    60.0,
		"network_usage": 40.0,
		"load_average":  2.5,
		"error_rate":    0.05,
		"process_count": 150.0,
	}

	derived := extractor.calculateDerivedFeatures(features)

	if derived == nil {
		t.Fatal("Expected derived features to be returned, got nil")
	}

	// Test CPU/Memory ratio
	if ratio, exists := derived["cpu_memory_ratio"]; exists {
		expectedRatio := 70.0 / 80.0
		if ratio != expectedRatio {
			t.Errorf("Expected cpu_memory_ratio %f, got %f", expectedRatio, ratio)
		}
	} else {
		t.Error("Expected cpu_memory_ratio to be calculated")
	}

	// Test load per process
	if loadPerProcess, exists := derived["load_per_process"]; exists {
		expectedLoadPerProcess := 2.5 / 150.0
		if loadPerProcess != expectedLoadPerProcess {
			t.Errorf("Expected load_per_process %f, got %f", expectedLoadPerProcess, loadPerProcess)
		}
	} else {
		t.Error("Expected load_per_process to be calculated")
	}

	// Test resource pressure index
	if _, exists := derived["resource_pressure_index"]; !exists {
		t.Error("Expected resource_pressure_index to be calculated")
	}
}

func TestNewAnomalyDetector(t *testing.T) {
	detector := NewAnomalyDetector()

	if detector == nil {
		t.Fatal("Expected anomaly detector to be created, got nil")
	}

	if detector.models == nil {
		t.Error("Expected models map to be initialized")
	}

	if detector.thresholds == nil {
		t.Error("Expected thresholds map to be initialized")
	}

	if detector.baselineData == nil {
		t.Error("Expected baseline data to be initialized")
	}

	if detector.detectionWindow <= 0 {
		t.Error("Expected detection window to be positive")
	}
}

func TestAnomalyDetectorDetectAnomalies(t *testing.T) {
	detector := NewAnomalyDetector()

	// Create test metrics
	metrics := map[string]float64{
		"cpu_usage":     70.0,
		"memory_usage":  80.0,
		"disk_usage":    60.0,
		"network_usage": 40.0,
		"load_average":  2.5,
		"error_rate":    0.05,
		"process_count": 150.0,
	}

	report, err := detector.DetectAnomalies(metrics)
	if err != nil {
		t.Fatalf("Failed to detect anomalies: %v", err)
	}

	if report == nil {
		t.Fatal("Expected anomaly report to be returned, got nil")
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

	// Test with highly anomalous metrics
	anomalousMetrics := map[string]float64{
		"cpu_usage":     150.0,  // Impossible value
		"memory_usage":  200.0,  // Impossible value
		"disk_usage":    -10.0,  // Impossible value
		"error_rate":    1.0,    // 100% error rate
	}

	anomalousReport, err := detector.DetectAnomalies(anomalousMetrics)
	if err != nil {
		t.Fatalf("Failed to detect anomalies with anomalous data: %v", err)
	}

	if anomalousReport.AnomalyScore <= report.AnomalyScore {
		t.Errorf("Expected higher anomaly score for anomalous data: %f <= %f", 
			anomalousReport.AnomalyScore, report.AnomalyScore)
	}
}

func TestIsolationForestNewIsolationForest(t *testing.T) {
	forest := NewIsolationForest(100, 10, 8)

	if forest == nil {
		t.Fatal("Expected isolation forest to be created, got nil")
	}

	if forest.NumTrees != 100 {
		t.Errorf("Expected 100 trees, got %d", forest.NumTrees)
	}

	if forest.SampleSize != 10 {
		t.Errorf("Expected sample size 10, got %d", forest.SampleSize)
	}

	if forest.MaxDepth != 8 {
		t.Errorf("Expected max depth 8, got %d", forest.MaxDepth)
	}

	if len(forest.Trees) != 100 {
		t.Errorf("Expected 100 trees to be created, got %d", len(forest.Trees))
	}
}

func TestIsolationForestTrain(t *testing.T) {
	forest := NewIsolationForest(10, 5, 4) // Smaller forest for testing

	// Create training data
	trainingData := []map[string]float64{
		{"cpu": 50.0, "memory": 60.0},
		{"cpu": 55.0, "memory": 65.0},
		{"cpu": 45.0, "memory": 55.0},
		{"cpu": 52.0, "memory": 62.0},
		{"cpu": 48.0, "memory": 58.0},
		{"cpu": 53.0, "memory": 63.0},
		{"cpu": 47.0, "memory": 57.0},
		{"cpu": 51.0, "memory": 61.0},
		{"cpu": 49.0, "memory": 59.0},
		{"cpu": 54.0, "memory": 64.0},
	}

	err := forest.Train(trainingData)
	if err != nil {
		t.Fatalf("Failed to train isolation forest: %v", err)
	}

	if forest.TrainingPoints != len(trainingData) {
		t.Errorf("Expected training points %d, got %d", len(trainingData), forest.TrainingPoints)
	}

	// Verify trees were built
	for i, tree := range forest.Trees {
		if tree.Root == nil {
			t.Errorf("Tree %d missing root node", i)
		}
	}
}

func TestIsolationForestDetectAnomaly(t *testing.T) {
	forest := NewIsolationForest(10, 5, 4)

	// Train with normal data
	normalData := []map[string]float64{
		{"cpu": 50.0, "memory": 60.0},
		{"cpu": 55.0, "memory": 65.0},
		{"cpu": 45.0, "memory": 55.0},
		{"cpu": 52.0, "memory": 62.0},
		{"cpu": 48.0, "memory": 58.0},
		{"cpu": 53.0, "memory": 63.0},
		{"cpu": 47.0, "memory": 57.0},
		{"cpu": 51.0, "memory": 61.0},
		{"cpu": 49.0, "memory": 59.0},
		{"cpu": 54.0, "memory": 64.0},
	}

	forest.Train(normalData)

	// Test normal point
	normalPoint := map[string]float64{"cpu": 50.0, "memory": 60.0}
	normalScore := forest.DetectAnomaly(normalPoint)

	if normalScore < 0 || normalScore > 1 {
		t.Errorf("Expected anomaly score between 0 and 1, got %f", normalScore)
	}

	// Test anomalous point
	anomalousPoint := map[string]float64{"cpu": 200.0, "memory": 300.0}
	anomalousScore := forest.DetectAnomaly(anomalousPoint)

	if anomalousScore <= normalScore {
		t.Errorf("Expected higher anomaly score for anomalous point: %f <= %f", 
			anomalousScore, normalScore)
	}

	// Test edge case: empty point
	emptyPoint := map[string]float64{}
	emptyScore := forest.DetectAnomaly(emptyPoint)
	if emptyScore < 0 || emptyScore > 1 {
		t.Errorf("Expected valid anomaly score for empty point, got %f", emptyScore)
	}
}

func TestNewResourceForecaster(t *testing.T) {
	forecaster := NewResourceForecaster()

	if forecaster == nil {
		t.Fatal("Expected resource forecaster to be created, got nil")
	}

	if forecaster.models == nil {
		t.Error("Expected models map to be initialized")
	}

	if forecaster.historicalData == nil {
		t.Error("Expected historical data to be initialized")
	}

	if forecaster.forecastWindow <= 0 {
		t.Error("Expected forecast window to be positive")
	}
}

func TestResourceForecasterForecast(t *testing.T) {
	forecaster := NewResourceForecaster()

	// Add some historical data for training
	historical := []health.ResourceDataPoint{
		{Timestamp: time.Now().Add(-5 * time.Hour), Value: 50.0},
		{Timestamp: time.Now().Add(-4 * time.Hour), Value: 55.0},
		{Timestamp: time.Now().Add(-3 * time.Hour), Value: 60.0},
		{Timestamp: time.Now().Add(-2 * time.Hour), Value: 65.0},
		{Timestamp: time.Now().Add(-1 * time.Hour), Value: 70.0},
	}

	forecaster.mu.Lock()
	forecaster.historicalData["cpu"] = historical
	forecaster.historicalData["memory"] = historical
	forecaster.historicalData["disk"] = historical
	forecaster.historicalData["network"] = historical
	forecaster.mu.Unlock()

	timeline := 3 * time.Hour
	forecast, err := forecaster.Forecast(timeline)
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

	// Verify forecasts for expected resources
	expectedResources := []string{"cpu", "memory", "disk", "network"}
	for _, resource := range expectedResources {
		if _, exists := forecast.Predictions[resource]; !exists {
			t.Errorf("Expected forecast for resource %s", resource)
		}

		if len(forecast.CPUForecast) == 0 && resource == "cpu" {
			t.Error("Expected CPU forecast data points")
		}
	}

	// Verify forecast accuracy and confidence
	if forecast.ModelAccuracy < 0 || forecast.ModelAccuracy > 1 {
		t.Errorf("Expected model accuracy between 0 and 1, got %f", forecast.ModelAccuracy)
	}

	if forecast.Confidence < 0 || forecast.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", forecast.Confidence)
	}
}

func TestResourceForecasterAddDataPoint(t *testing.T) {
	forecaster := NewResourceForecaster()

	dataPoint := health.ResourceDataPoint{
		Timestamp: time.Now(),
		Value:     75.0,
		Predicted: false,
	}

	forecaster.AddDataPoint("cpu", dataPoint)

	// Verify data point was added
	forecaster.mu.RLock()
	cpuData, exists := forecaster.historicalData["cpu"]
	forecaster.mu.RUnlock()

	if !exists {
		t.Error("Expected CPU data to exist after adding data point")
	}

	if len(cpuData) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(cpuData))
	}

	if cpuData[0].Value != 75.0 {
		t.Errorf("Expected data point value 75.0, got %f", cpuData[0].Value)
	}

	// Add many data points to test trimming
	for i := 0; i < 2000; i++ {
		point := health.ResourceDataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Value:     float64(i),
		}
		forecaster.AddDataPoint("test_resource", point)
	}

	forecaster.mu.RLock()
	testData := forecaster.historicalData["test_resource"]
	forecaster.mu.RUnlock()

	if len(testData) > 1000 {
		t.Errorf("Expected data to be trimmed to 1000 points, got %d", len(testData))
	}
}

func TestARIMAModelNewARIMAModel(t *testing.T) {
	model := NewARIMAModel(1, 1, 1) // Simple ARIMA(1,1,1)

	if model == nil {
		t.Fatal("Expected ARIMA model to be created, got nil")
	}

	if model.P != 1 {
		t.Errorf("Expected P=1, got %d", model.P)
	}

	if model.D != 1 {
		t.Errorf("Expected D=1, got %d", model.D)
	}

	if model.Q != 1 {
		t.Errorf("Expected Q=1, got %d", model.Q)
	}
}

func TestARIMAModelFit(t *testing.T) {
	model := NewARIMAModel(1, 1, 1)

	// Create time series data with trend
	timeSeries := []float64{10, 12, 14, 16, 18, 20, 22, 24, 26, 28}

	err := model.Fit(timeSeries)
	if err != nil {
		t.Fatalf("Failed to fit ARIMA model: %v", err)
	}

	if !model.Fitted {
		t.Error("Expected model to be fitted after successful fit")
	}

	// Test with insufficient data
	shortSeries := []float64{1, 2}
	err = model.Fit(shortSeries)
	if err == nil {
		t.Error("Expected error when fitting with insufficient data")
	}
}

func TestARIMAModelForecast(t *testing.T) {
	model := NewARIMAModel(1, 1, 1)

	// Fit model first
	timeSeries := []float64{10, 12, 14, 16, 18, 20, 22, 24, 26, 28}
	err := model.Fit(timeSeries)
	if err != nil {
		t.Fatalf("Failed to fit model: %v", err)
	}

	// Test forecast
	forecast, err := model.Forecast(5) // Forecast 5 steps ahead
	if err != nil {
		t.Fatalf("Failed to forecast: %v", err)
	}

	if len(forecast) != 5 {
		t.Errorf("Expected 5 forecast points, got %d", len(forecast))
	}

	// Forecast values should be reasonable (continuing the trend)
	for i, value := range forecast {
		if value <= 0 {
			t.Errorf("Forecast point %d has unexpected value: %f", i, value)
		}
	}

	// Test forecast without fitting
	unfittedModel := NewARIMAModel(1, 1, 1)
	_, err = unfittedModel.Forecast(5)
	if err == nil {
		t.Error("Expected error when forecasting with unfitted model")
	}
}

// Benchmark tests

func BenchmarkFeatureExtractorExtractFeatures(b *testing.B) {
	extractor := NewFeatureExtractor()

	metrics := map[string]float64{
		"cpu_usage":     70.0,
		"memory_usage":  80.0,
		"disk_usage":    60.0,
		"network_usage": 40.0,
		"load_average":  2.5,
		"error_rate":    0.05,
		"process_count": 150.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractFeatures(metrics)
		if err != nil {
			b.Fatalf("Failed to extract features: %v", err)
		}
	}
}

func BenchmarkAnomalyDetectorDetectAnomalies(b *testing.B) {
	detector := NewAnomalyDetector()

	metrics := map[string]float64{
		"cpu_usage":     70.0,
		"memory_usage":  80.0,
		"disk_usage":    60.0,
		"network_usage": 40.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := detector.DetectAnomalies(metrics)
		if err != nil {
			b.Fatalf("Failed to detect anomalies: %v", err)
		}
	}
}

func BenchmarkIsolationForestDetectAnomaly(b *testing.B) {
	forest := NewIsolationForest(100, 10, 8)

	// Train forest
	trainingData := make([]map[string]float64, 100)
	for i := 0; i < 100; i++ {
		trainingData[i] = map[string]float64{
			"cpu":    float64(50 + i%20),
			"memory": float64(60 + i%15),
		}
	}
	forest.Train(trainingData)

	testPoint := map[string]float64{"cpu": 75.0, "memory": 85.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		forest.DetectAnomaly(testPoint)
	}
}

func BenchmarkARIMAModelForecast(b *testing.B) {
	model := NewARIMAModel(1, 1, 1)

	// Fit model
	timeSeries := make([]float64, 100)
	for i := 0; i < 100; i++ {
		timeSeries[i] = float64(i) + float64(i%10)*0.5 // Simple trend with noise
	}
	model.Fit(timeSeries)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Forecast(10)
		if err != nil {
			b.Fatalf("Failed to forecast: %v", err)
		}
	}
}