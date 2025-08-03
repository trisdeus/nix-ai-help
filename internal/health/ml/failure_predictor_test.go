// Package ml provides tests for the failure prediction ML model
package ml

import (
	"context"
	"testing"
	"time"
)

func TestNewFailurePredictionModel(t *testing.T) {
	// Test with nil config (should use defaults)
	model := NewFailurePredictionModel(nil)
	if model == nil {
		t.Fatal("Expected model to be created, got nil")
	}

	if model.config == nil {
		t.Error("Expected config to be initialized with defaults")
	}

	if model.patterns == nil {
		t.Error("Expected patterns map to be initialized")
	}

	if model.featureExtractor == nil {
		t.Error("Expected feature extractor to be initialized")
	}

	if model.anomalyDetector == nil {
		t.Error("Expected anomaly detector to be initialized")
	}

	if model.resourceForecaster == nil {
		t.Error("Expected resource forecaster to be initialized")
	}

	// Test with custom config
	customConfig := &MLConfig{
		TrainingWindow:                 15 * 24 * time.Hour,
		PredictionHorizon:             3 * 24 * time.Hour,
		MinTrainingEvents:             50,
		ModelUpdateInterval:           12 * time.Hour,
		FeatureImportanceThreshold:    0.2,
		AnomalyThreshold:              0.9,
		ConfidenceThreshold:           0.8,
		MaxPatterns:                   500,
	}

	customModel := NewFailurePredictionModel(customConfig)
	if customModel.config.TrainingWindow != customConfig.TrainingWindow {
		t.Errorf("Expected training window %v, got %v", 
			customConfig.TrainingWindow, customModel.config.TrainingWindow)
	}
}

func TestFailurePredictionModelTrain(t *testing.T) {
	model := NewFailurePredictionModel(nil)
	ctx := context.Background()

	// Create training data
	trainingData := []health.HealthEvent{
		{
			ID:        "event_1",
			Type:      "metric_collection",
			Component: "system",
			Timestamp: time.Now().Add(-24 * time.Hour),
			Severity:  health.PriorityLow,
			Metrics: map[string]interface{}{
				"cpu_usage":     70.0,
				"memory_usage":  60.0,
				"disk_usage":    50.0,
				"error_rate":    0.01,
			},
		},
		{
			ID:        "event_2",
			Type:      "failure",
			Component: "disk",
			Timestamp: time.Now().Add(-12 * time.Hour),
			Severity:  health.PriorityHigh,
			Metrics: map[string]interface{}{
				"cpu_usage":     75.0,
				"memory_usage":  80.0,
				"disk_usage":    95.0,
				"error_rate":    0.1,
			},
		},
		{
			ID:        "event_3",
			Type:      "metric_collection",
			Component: "system",
			Timestamp: time.Now().Add(-6 * time.Hour),
			Severity:  health.PriorityLow,
			Metrics: map[string]interface{}{
				"cpu_usage":     65.0,
				"memory_usage":  55.0,
				"disk_usage":    48.0,
				"error_rate":    0.005,
			},
		},
	}

	// Add more events to meet minimum training requirements
	for i := 0; i < 100; i++ {
		event := health.HealthEvent{
			ID:        fmt.Sprintf("training_event_%d", i),
			Type:      "metric_collection",
			Component: "system",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			Severity:  health.PriorityLow,
			Metrics: map[string]interface{}{
				"cpu_usage":     float64(50 + i%30),
				"memory_usage":  float64(40 + i%25),
				"disk_usage":    float64(30 + i%20),
				"error_rate":    float64(i%10) / 1000.0,
			},
		}
		
		// Add some failure events
		if i%20 == 0 {
			event.Type = "failure"
			event.Severity = health.PriorityHigh
		}
		
		trainingData = append(trainingData, event)
	}

	// Test training
	err := model.Train(ctx, trainingData)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Verify model was trained
	if len(model.historicalEvents) != len(trainingData) {
		t.Errorf("Expected %d historical events, got %d", 
			len(trainingData), len(model.historicalEvents))
	}

	if model.lastTrainingTime.IsZero() {
		t.Error("Expected lastTrainingTime to be set after training")
	}

	// Test training with insufficient data
	insufficientData := trainingData[:10] // Less than minimum required
	err = model.Train(ctx, insufficientData)
	if err == nil {
		t.Error("Expected error when training with insufficient data")
	}
}

func TestFailurePredictionModelPredict(t *testing.T) {
	model := NewFailurePredictionModel(nil)
	ctx := context.Background()

	// Train model first
	trainingData := generateMockTrainingData(150)
	err := model.Train(ctx, trainingData)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Test prediction
	timeline := 7 * 24 * time.Hour
	result, err := model.Predict(ctx, timeline)
	if err != nil {
		t.Fatalf("Failed to predict: %v", err)
	}

	prediction, ok := result.(*health.FailurePrediction)
	if !ok {
		t.Fatal("Expected FailurePrediction result")
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

	// Verify predicted failures structure
	for i, failure := range prediction.PredictedFailures {
		if failure.ID == "" {
			t.Errorf("Predicted failure %d missing ID", i)
		}
		if failure.ProbabilityScore < 0 || failure.ProbabilityScore > 1 {
			t.Errorf("Predicted failure %d has invalid probability: %f", 
				i, failure.ProbabilityScore)
		}
		if failure.EstimatedTime.IsZero() {
			t.Errorf("Predicted failure %d missing estimated time", i)
		}
	}

	// Verify preventive actions
	for i, action := range prediction.PreventiveActions {
		if action.ID == "" {
			t.Errorf("Preventive action %d missing ID", i)
		}
		if action.Description == "" {
			t.Errorf("Preventive action %d missing description", i)
		}
	}
}

func TestFailurePredictionModelEvaluate(t *testing.T) {
	model := NewFailurePredictionModel(nil)
	ctx := context.Background()

	// Train model
	trainingData := generateMockTrainingData(150)
	err := model.Train(ctx, trainingData)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Create test data
	testData := generateMockTrainingData(50)

	// Test evaluation
	metrics, err := model.Evaluate(ctx, testData)
	if err != nil {
		t.Fatalf("Failed to evaluate model: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}

	// Verify metrics are in valid ranges
	if metrics.Accuracy < 0 || metrics.Accuracy > 1 {
		t.Errorf("Expected accuracy between 0 and 1, got %f", metrics.Accuracy)
	}

	if metrics.Precision < 0 || metrics.Precision > 1 {
		t.Errorf("Expected precision between 0 and 1, got %f", metrics.Precision)
	}

	if metrics.Recall < 0 || metrics.Recall > 1 {
		t.Errorf("Expected recall between 0 and 1, got %f", metrics.Recall)
	}

	if metrics.F1Score < 0 || metrics.F1Score > 1 {
		t.Errorf("Expected F1 score between 0 and 1, got %f", metrics.F1Score)
	}

	if metrics.EvaluatedAt.IsZero() {
		t.Error("Expected EvaluatedAt to be set")
	}
}

func TestFailurePredictionModelGetInfo(t *testing.T) {
	model := NewFailurePredictionModel(nil)

	info := model.GetInfo()

	if info.Name == "" {
		t.Error("Expected model name to be set")
	}

	if info.Type == "" {
		t.Error("Expected model type to be set")
	}

	if info.Version == "" {
		t.Error("Expected model version to be set")
	}

	expectedFeatures := []string{"cpu_usage", "memory_usage", "disk_usage", "network_usage", "load_average", "error_rate"}
	if len(info.Features) == 0 {
		t.Error("Expected features to be listed")
	}

	for _, expectedFeature := range expectedFeatures {
		found := false
		for _, feature := range info.Features {
			if feature == expectedFeature {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected feature %s to be in model info", expectedFeature)
		}
	}
}

func TestFailurePredictionModelUpdate(t *testing.T) {
	model := NewFailurePredictionModel(nil)
	ctx := context.Background()

	// Train model initially
	initialData := generateMockTrainingData(150)
	err := model.Train(ctx, initialData)
	if err != nil {
		t.Fatalf("Failed to train model initially: %v", err)
	}

	initialTrainingTime := model.lastTrainingTime

	// Wait a moment to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Update with new data
	newData := generateMockTrainingData(50)
	err = model.Update(ctx, newData)
	if err != nil {
		t.Fatalf("Failed to update model: %v", err)
	}

	// Verify update occurred
	if !model.lastTrainingTime.After(initialTrainingTime) {
		t.Error("Expected lastTrainingTime to be updated after model update")
	}

	// Verify new data was incorporated
	expectedTotal := len(initialData) + len(newData)
	if len(model.historicalEvents) != expectedTotal {
		t.Errorf("Expected %d total events after update, got %d", 
			expectedTotal, len(model.historicalEvents))
	}
}

func TestCalculateFailureProbability(t *testing.T) {
	model := NewFailurePredictionModel(nil)

	// Create test pattern
	pattern := &FailurePattern{
		ID:        "test_pattern",
		Type:      health.FailureDiskSpace,
		Component: "disk",
		Frequency: 20, // 20% historical frequency
		SuccessRate: 0.8,
		LastSeen:  time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
		Features: map[string]float64{
			"disk_usage": 85.0,
			"error_rate": 0.05,
		},
		Precursors: []PatternPrecursor{
			{
				Event:      "disk_usage",
				Threshold:  80.0,
				Direction:  "above",
				Importance: 0.9,
				Frequency:  0.8,
			},
		},
	}

	// Test with matching current features
	currentFeatures := map[string]float64{
		"disk_usage": 82.0,
		"error_rate": 0.04,
	}

	timeline := 7 * 24 * time.Hour
	probability := model.calculateFailureProbability(pattern, currentFeatures, timeline)

	if probability < 0 || probability > 1 {
		t.Errorf("Expected probability between 0 and 1, got %f", probability)
	}

	// Test with non-matching features (should have lower probability)
	nonMatchingFeatures := map[string]float64{
		"disk_usage": 30.0, // Much lower than pattern
		"error_rate": 0.001,
	}

	lowProbability := model.calculateFailureProbability(pattern, nonMatchingFeatures, timeline)

	if lowProbability >= probability {
		t.Errorf("Expected lower probability for non-matching features: %f >= %f", 
			lowProbability, probability)
	}
}

func TestCalculateFeatureSimilarity(t *testing.T) {
	model := NewFailurePredictionModel(nil)

	patternFeatures := map[string]float64{
		"cpu_usage":    70.0,
		"memory_usage": 80.0,
		"disk_usage":   60.0,
	}

	// Test identical features
	identicalFeatures := map[string]float64{
		"cpu_usage":    70.0,
		"memory_usage": 80.0,
		"disk_usage":   60.0,
	}

	similarity := model.calculateFeatureSimilarity(patternFeatures, identicalFeatures)
	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for identical features, got %f", similarity)
	}

	// Test similar features
	similarFeatures := map[string]float64{
		"cpu_usage":    72.0,  // Close to 70.0
		"memory_usage": 78.0,  // Close to 80.0
		"disk_usage":   62.0,  // Close to 60.0
	}

	similarSimilarity := model.calculateFeatureSimilarity(patternFeatures, similarFeatures)
	if similarSimilarity <= 0.8 {
		t.Errorf("Expected high similarity for close features, got %f", similarSimilarity)
	}

	// Test very different features
	differentFeatures := map[string]float64{
		"cpu_usage":    10.0,  // Very different from 70.0
		"memory_usage": 20.0,  // Very different from 80.0
		"disk_usage":   30.0,  // Different from 60.0
	}

	differentSimilarity := model.calculateFeatureSimilarity(patternFeatures, differentFeatures)
	if differentSimilarity >= similarSimilarity {
		t.Errorf("Expected lower similarity for different features: %f >= %f", 
			differentSimilarity, similarSimilarity)
	}
}

func TestEvaluatePrecursors(t *testing.T) {
	model := NewFailurePredictionModel(nil)

	precursors := []PatternPrecursor{
		{
			Event:      "disk_usage",
			Threshold:  80.0,
			Direction:  "above",
			Importance: 0.9,
			Frequency:  0.8,
		},
		{
			Event:      "error_rate",
			Threshold:  0.01,
			Direction:  "above",
			Importance: 0.7,
			Frequency:  0.6,
		},
	}

	// Test with precursors met
	metFeatures := map[string]float64{
		"disk_usage": 85.0,  // Above 80.0 threshold
		"error_rate": 0.02,  // Above 0.01 threshold
	}

	metScore := model.evaluatePrecursors(precursors, metFeatures)
	if metScore <= 0.5 {
		t.Errorf("Expected high precursor score when conditions are met, got %f", metScore)
	}

	// Test with precursors not met
	unmetFeatures := map[string]float64{
		"disk_usage": 70.0,   // Below 80.0 threshold
		"error_rate": 0.005,  // Below 0.01 threshold
	}

	unmetScore := model.evaluatePrecursors(precursors, unmetFeatures)
	if unmetScore >= metScore {
		t.Errorf("Expected lower precursor score when conditions are not met: %f >= %f", 
			unmetScore, metScore)
	}
}

func TestGeneratePreventiveActions(t *testing.T) {
	model := NewFailurePredictionModel(nil)

	// Test disk space failure
	diskFailure := health.PredictedFailure{
		Type:      health.FailureDiskSpace,
		Component: "disk",
	}

	actions := model.generatePreventiveActions(diskFailure)
	if len(actions) == 0 {
		t.Error("Expected preventive actions for disk space failure")
	}

	// Verify action structure
	for i, action := range actions {
		if action.ID == "" {
			t.Errorf("Action %d missing ID", i)
		}
		if action.Description == "" {
			t.Errorf("Action %d missing description", i)
		}
		if action.ETA <= 0 {
			t.Errorf("Action %d has invalid ETA: %v", i, action.ETA)
		}
	}

	// Test memory leak failure
	memoryFailure := health.PredictedFailure{
		Type:      health.FailureMemoryLeak,
		Component: "memory",
	}

	memoryActions := model.generatePreventiveActions(memoryFailure)
	if len(memoryActions) == 0 {
		t.Error("Expected preventive actions for memory leak failure")
	}

	// Test service crash failure
	serviceFailure := health.PredictedFailure{
		Type:      health.FailureServiceCrash,
		Component: "nginx",
	}

	serviceActions := model.generatePreventiveActions(serviceFailure)
	if len(serviceActions) == 0 {
		t.Error("Expected preventive actions for service crash failure")
	}
}

func TestLearnPatternsFromEvents(t *testing.T) {
	model := NewFailurePredictionModel(nil)

	// Create events with patterns
	events := []health.HealthEvent{
		{
			ID:        "event_1",
			Type:      "metric_collection",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Metrics: map[string]interface{}{
				"disk_usage": 85.0,
				"error_rate": 0.05,
			},
		},
		{
			ID:        "event_2",
			Type:      "failure",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Severity:  health.PriorityHigh,
			Metrics: map[string]interface{}{
				"disk_usage": 95.0,
				"error_rate": 0.1,
			},
		},
	}

	model.learnPatternsFromEvents(events)

	if len(model.patterns) == 0 {
		t.Error("Expected patterns to be learned from events")
	}

	// Verify pattern structure
	for id, pattern := range model.patterns {
		if pattern.ID != id {
			t.Errorf("Pattern ID mismatch: %s != %s", pattern.ID, id)
		}
		if pattern.LastSeen.IsZero() {
			t.Errorf("Pattern %s missing LastSeen timestamp", id)
		}
		if pattern.Frequency <= 0 {
			t.Errorf("Pattern %s has invalid frequency: %d", id, pattern.Frequency)
		}
	}
}

// Helper functions for tests

func generateMockTrainingData(count int) []health.HealthEvent {
	var events []health.HealthEvent
	
	for i := 0; i < count; i++ {
		event := health.HealthEvent{
			ID:        fmt.Sprintf("mock_event_%d", i),
			Type:      "metric_collection",
			Component: "system",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			Severity:  health.PriorityLow,
			Metrics: map[string]interface{}{
				"cpu_usage":     float64(40 + i%40),
				"memory_usage":  float64(30 + i%50),
				"disk_usage":    float64(20 + i%60),
				"network_usage": float64(10 + i%30),
				"load_average":  float64(1 + i%5),
				"error_rate":    float64(i%20) / 1000.0,
				"process_count": float64(100 + i%100),
			},
			Context: make(map[string]interface{}),
		}
		
		// Add some failure events
		if i%25 == 0 {
			event.Type = "failure"
			event.Severity = health.PriorityHigh
			event.Message = "Mock failure event"
			
			// Simulate failure conditions
			event.Metrics["disk_usage"] = 95.0
			event.Metrics["error_rate"] = 0.1
		}
		
		events = append(events, event)
	}
	
	return events
}

// Benchmark tests

func BenchmarkFailurePredictionModelPredict(b *testing.B) {
	model := NewFailurePredictionModel(nil)
	ctx := context.Background()

	// Train model
	trainingData := generateMockTrainingData(200)
	model.Train(ctx, trainingData)

	timeline := 7 * 24 * time.Hour

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.Predict(ctx, timeline)
		if err != nil {
			b.Fatalf("Failed to predict: %v", err)
		}
	}
}

func BenchmarkCalculateFeatureSimilarity(b *testing.B) {
	model := NewFailurePredictionModel(nil)

	patternFeatures := map[string]float64{
		"cpu_usage":    70.0,
		"memory_usage": 80.0,
		"disk_usage":   60.0,
		"error_rate":   0.05,
	}

	currentFeatures := map[string]float64{
		"cpu_usage":    72.0,
		"memory_usage": 78.0,
		"disk_usage":   62.0,
		"error_rate":   0.04,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.calculateFeatureSimilarity(patternFeatures, currentFeatures)
	}
}

func BenchmarkLearnPatternsFromEvents(b *testing.B) {
	model := NewFailurePredictionModel(nil)
	events := generateMockTrainingData(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.learnPatternsFromEvents(events)
	}
}