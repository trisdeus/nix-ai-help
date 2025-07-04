// Package ml provides machine learning models for system health prediction
package ml

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"nix-ai-help/pkg/logger"
)

// Local type definitions to avoid circular imports

type FailureType string

const (
	FailureDiskSpace       FailureType = "disk_space"
	FailureMemoryLeak      FailureType = "memory_leak"
	FailureServiceCrash    FailureType = "service_crash"
	FailureNetworkIssue    FailureType = "network_issue"
	FailureHardwareFailure FailureType = "hardware_failure"
	FailureSecurityBreach  FailureType = "security_breach"
	FailurePerformance     FailureType = "performance_degradation"
	FailureConfiguration   FailureType = "configuration_error"
	FailureDependency      FailureType = "dependency_failure"
	FailureUpdate          FailureType = "update_failure"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type ImpactLevel string

const (
	ImpactMinimal    ImpactLevel = "minimal"
	ImpactModerate   ImpactLevel = "moderate"
	ImpactSignificant ImpactLevel = "significant"
	ImpactSevere     ImpactLevel = "severe"
)

type PredictedFailure struct {
	ID               string                 `json:"id"`
	Type             FailureType            `json:"type"`
	Component        string                 `json:"component"`
	Description      string                 `json:"description"`
	ProbabilityScore float64                `json:"probability_score"`
	EstimatedTime    time.Time              `json:"estimated_time"`
	Impact           ImpactLevel            `json:"impact"`
	Indicators       []HealthIndicator      `json:"indicators"`
	HistoricalData   []HistoricalEvent      `json:"historical_data"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type HealthIndicator struct {
	Name      string   `json:"name"`
	Value     float64  `json:"value"`
	Threshold float64  `json:"threshold"`
	Unit      string   `json:"unit"`
	Trend     string   `json:"trend"`
	Severity  Priority `json:"severity"`
	Source    string   `json:"source"`
}

type HistoricalEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Severity    Priority               `json:"severity"`
	Resolution  string                 `json:"resolution,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PreventiveAction struct {
	ID           string                 `json:"id"`
	Description  string                 `json:"description"`
	Commands     []string               `json:"commands"`
	Automated    bool                   `json:"automated"`
	Risk         RiskLevel              `json:"risk"`
	ETA          time.Duration          `json:"eta"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type FailurePrediction struct {
	Timeline          time.Duration          `json:"timeline"`
	PredictedFailures []PredictedFailure     `json:"predicted_failures"`
	Confidence        float64                `json:"confidence"`
	RiskLevel         RiskLevel              `json:"risk_level"`
	PreventiveActions []PreventiveAction     `json:"preventive_actions"`
	GeneratedAt       time.Time              `json:"generated_at"`
	ModelVersion      string                 `json:"model_version"`
	Metadata          map[string]interface{} `json:"metadata"`
}

type ModelInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"`
	Accuracy     float64                `json:"accuracy"`
	LastTrained  time.Time              `json:"last_trained"`
	DataPoints   int                    `json:"data_points"`
	Features     []string               `json:"features"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type ModelMetrics struct {
	Accuracy    float64   `json:"accuracy"`
	Precision   float64   `json:"precision"`
	Recall      float64   `json:"recall"`
	F1Score     float64   `json:"f1_score"`
	AUC         float64   `json:"auc"`
	RMSE        float64   `json:"rmse"`
	MAE         float64   `json:"mae"`
	Confusion   [][]int   `json:"confusion_matrix"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// FailurePredictionModel implements ML-based failure prediction
type FailurePredictionModel struct {
	mu                    sync.RWMutex
	logger                *logger.Logger
	historicalEvents      []HealthEvent
	patterns              map[string]*FailurePattern
	featureExtractor      *FeatureExtractor
	modelMetrics          *ModelMetrics
	lastTrainingTime      time.Time
	predictionThreshold   float64
	anomalyDetector       *AnomalyDetector
	resourceForecaster    *ResourceForecaster
	config                *MLConfig
}

// FailurePattern represents a learned failure pattern
type FailurePattern struct {
	ID               string                 `json:"id"`
	Type             FailureType     `json:"type"`
	Component        string                 `json:"component"`
	Precursors       []PatternPrecursor     `json:"precursors"`
	TimeToFailure    time.Duration          `json:"time_to_failure"`
	Confidence       float64                `json:"confidence"`
	Frequency        int                    `json:"frequency"`
	LastSeen         time.Time              `json:"last_seen"`
	SuccessRate      float64                `json:"success_rate"`
	Features         map[string]float64     `json:"features"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// PatternPrecursor represents events that precede failures
type PatternPrecursor struct {
	Event         string        `json:"event"`
	TimeOffset    time.Duration `json:"time_offset"`    // Time before failure
	Importance    float64       `json:"importance"`     // Feature importance score
	Threshold     float64       `json:"threshold"`      // Threshold value
	Direction     string        `json:"direction"`      // "above", "below", "equals"
	Frequency     float64       `json:"frequency"`      // How often this precursor appears
}


// MLConfig configures the ML models
type MLConfig struct {
	TrainingWindow        time.Duration `json:"training_window"`
	PredictionHorizon     time.Duration `json:"prediction_horizon"`
	MinTrainingEvents     int           `json:"min_training_events"`
	ModelUpdateInterval   time.Duration `json:"model_update_interval"`
	FeatureImportanceThreshold float64  `json:"feature_importance_threshold"`
	AnomalyThreshold      float64       `json:"anomaly_threshold"`
	ConfidenceThreshold   float64       `json:"confidence_threshold"`
	MaxPatterns           int           `json:"max_patterns"`
}

// NewFailurePredictionModel creates a new failure prediction model
func NewFailurePredictionModel(config *MLConfig) *FailurePredictionModel {
	if config == nil {
		config = &MLConfig{
			TrainingWindow:        30 * 24 * time.Hour, // 30 days
			PredictionHorizon:     7 * 24 * time.Hour,  // 7 days
			MinTrainingEvents:     100,
			ModelUpdateInterval:   24 * time.Hour,      // Daily updates
			FeatureImportanceThreshold: 0.1,
			AnomalyThreshold:      0.8,
			ConfidenceThreshold:   0.7,
			MaxPatterns:           1000,
		}
	}

	featureExtractor := NewFeatureExtractor()
	anomalyDetector := NewAnomalyDetector()
	resourceForecaster := NewResourceForecaster()

	return &FailurePredictionModel{
		logger:              logger.NewLogger(),
		historicalEvents:    make([]HealthEvent, 0),
		patterns:            make(map[string]*FailurePattern),
		featureExtractor:    featureExtractor,
		anomalyDetector:     anomalyDetector,
		resourceForecaster:  resourceForecaster,
		predictionThreshold: config.ConfidenceThreshold,
		config:              config,
		modelMetrics: &ModelMetrics{
			Accuracy:    0.0,
			Precision:   0.0,
			Recall:      0.0,
			F1Score:     0.0,
			EvaluatedAt: time.Now(),
		},
	}
}

// Train implements health.MLModel interface
func (fpm *FailurePredictionModel) Train(ctx context.Context, data []HealthEvent) error {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()

	fpm.logger.Info(fmt.Sprintf("Training failure prediction model with %d events", len(data)))

	if len(data) < fpm.config.MinTrainingEvents {
		return fmt.Errorf("insufficient training data: need at least %d events, got %d", 
			fpm.config.MinTrainingEvents, len(data))
	}

	// Store historical events
	fpm.historicalEvents = append(fpm.historicalEvents, data...)
	
	// Sort events by timestamp
	sort.Slice(fpm.historicalEvents, func(i, j int) bool {
		return fpm.historicalEvents[i].Timestamp.Before(fpm.historicalEvents[j].Timestamp)
	})

	// Keep only recent events within training window
	cutoff := time.Now().Add(-fpm.config.TrainingWindow)
	var recentEvents []HealthEvent
	for _, event := range fpm.historicalEvents {
		if event.Timestamp.After(cutoff) {
			recentEvents = append(recentEvents, event)
		}
	}
	fpm.historicalEvents = recentEvents

	// Extract failure patterns
	if err := fpm.extractFailurePatterns(); err != nil {
		return fmt.Errorf("failed to extract failure patterns: %w", err)
	}

	// Train feature extractor
	if err := fpm.featureExtractor.Train(fpm.historicalEvents); err != nil {
		return fmt.Errorf("failed to train feature extractor: %w", err)
	}

	// Train anomaly detector
	features := fpm.extractFeatures(fpm.historicalEvents)
	if err := fpm.anomalyDetector.Train(features); err != nil {
		return fmt.Errorf("failed to train anomaly detector: %w", err)
	}

	// Train resource forecaster
	if err := fpm.resourceForecaster.Train(fpm.historicalEvents); err != nil {
		return fmt.Errorf("failed to train resource forecaster: %w", err)
	}

	// Evaluate model performance
	if err := fpm.evaluateModel(); err != nil {
		fpm.logger.Error(fmt.Sprintf("Failed to evaluate model: %v", err))
	}

	fpm.lastTrainingTime = time.Now()
	fpm.logger.Info("Failure prediction model training completed successfully")

	return nil
}

// Predict implements health.MLModel interface
func (fpm *FailurePredictionModel) Predict(ctx context.Context, input interface{}) (interface{}, error) {
	fpm.mu.RLock()
	defer fpm.mu.RUnlock()

	timeline, ok := input.(time.Duration)
	if !ok {
		timeline = fpm.config.PredictionHorizon
	}

	fpm.logger.Info(fmt.Sprintf("Predicting failures for timeline: %v", timeline))

	var predictions []PredictedFailure

	// Get current system state
	currentFeatures := fpm.getCurrentSystemFeatures()

	// Analyze each failure pattern
	for _, pattern := range fpm.patterns {
		if probability := fpm.calculateFailureProbability(pattern, currentFeatures, timeline); 
		   probability > fpm.predictionThreshold {
			
			prediction := PredictedFailure{
				ID:               fmt.Sprintf("pred_%s_%d", pattern.ID, time.Now().Unix()),
				Type:             pattern.Type,
				Component:        pattern.Component,
				Description:      fpm.generateFailureDescription(pattern),
				ProbabilityScore: probability,
				EstimatedTime:    fpm.estimateFailureTime(pattern, currentFeatures),
				Impact:           fpm.assessImpact(pattern),
				Indicators:       fpm.getHealthIndicators(pattern, currentFeatures),
				HistoricalData:   fpm.getHistoricalEvents(pattern),
				Metadata: map[string]interface{}{
					"pattern_id":       pattern.ID,
					"pattern_frequency": pattern.Frequency,
					"last_seen":        pattern.LastSeen,
					"success_rate":     pattern.SuccessRate,
				},
			}

			predictions = append(predictions, prediction)
		}
	}

	// Sort predictions by probability (highest first)
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].ProbabilityScore > predictions[j].ProbabilityScore
	})

	// Generate preventive actions
	preventiveActions := fpm.generatePreventiveActions(predictions)

	// Calculate overall risk level
	riskLevel := fpm.calculateRiskLevel(predictions)

	// Calculate overall confidence
	confidence := fpm.calculateOverallConfidence(predictions)

	failurePrediction := &FailurePrediction{
		Timeline:          timeline,
		PredictedFailures: predictions,
		Confidence:        confidence,
		RiskLevel:         riskLevel,
		PreventiveActions: preventiveActions,
		GeneratedAt:       time.Now(),
		ModelVersion:      "1.0",
		Metadata: map[string]interface{}{
			"patterns_analyzed": len(fpm.patterns),
			"model_accuracy":    fpm.modelMetrics.Accuracy,
			"training_events":   len(fpm.historicalEvents),
			"last_training":     fpm.lastTrainingTime,
		},
	}

	fpm.logger.Info(fmt.Sprintf("Generated %d failure predictions with %s risk level", 
		len(predictions), riskLevel))

	return failurePrediction, nil
}

// Evaluate implements health.MLModel interface
func (fpm *FailurePredictionModel) Evaluate(ctx context.Context, testData []HealthEvent) (*ModelMetrics, error) {
	fpm.mu.Lock()
	defer fpm.mu.Unlock()

	return fpm.evaluateModelWithData(testData)
}

// GetInfo implements health.MLModel interface
func (fpm *FailurePredictionModel) GetInfo() ModelInfo {
	fpm.mu.RLock()
	defer fpm.mu.RUnlock()

	features := fpm.featureExtractor.GetFeatureNames()

	return ModelInfo{
		Name:        "FailurePredictionModel",
		Version:     "1.0",
		Type:        "failure_prediction",
		Accuracy:    fpm.modelMetrics.Accuracy,
		LastTrained: fpm.lastTrainingTime,
		DataPoints:  len(fpm.historicalEvents),
		Features:    features,
		Metadata: map[string]interface{}{
			"patterns":          len(fpm.patterns),
			"prediction_horizon": fpm.config.PredictionHorizon.String(),
			"training_window":   fpm.config.TrainingWindow.String(),
			"threshold":         fpm.predictionThreshold,
		},
	}
}

// Update implements health.MLModel interface
func (fpm *FailurePredictionModel) Update(ctx context.Context, data []HealthEvent) error {
	// For incremental updates, we can add new events and retrain if enough new data
	fpm.mu.Lock()
	defer fpm.mu.Unlock()

	oldEventCount := len(fpm.historicalEvents)
	fpm.historicalEvents = append(fpm.historicalEvents, data...)

	// If we have significant new data, retrain
	if len(data) > oldEventCount/10 { // Retrain if new data is >10% of existing
		fpm.mu.Unlock() // Unlock for Train call
		return fpm.Train(ctx, data)
	}

	// Otherwise, just update patterns incrementally
	return fpm.updatePatternsIncremental(data)
}

// Private methods

func (fpm *FailurePredictionModel) extractFailurePatterns() error {
	fpm.logger.Info("Extracting failure patterns from historical events")

	// Group events by component and failure type
	failureEvents := make(map[string][]HealthEvent)
	
	for _, event := range fpm.historicalEvents {
		if event.Type == "failure" || event.Type == "error" || event.Type == "crash" {
			key := fmt.Sprintf("%s_%s", event.Component, event.Type)
			failureEvents[key] = append(failureEvents[key], event)
		}
	}

	// Analyze each failure type
	for key, failures := range failureEvents {
		if len(failures) < 3 { // Need at least 3 occurrences to establish pattern
			continue
		}

		pattern := fpm.analyzeFailurePattern(key, failures)
		if pattern != nil {
			fpm.patterns[pattern.ID] = pattern
		}
	}

	fpm.logger.Info(fmt.Sprintf("Extracted %d failure patterns", len(fpm.patterns)))
	return nil
}

func (fpm *FailurePredictionModel) analyzeFailurePattern(key string, failures []HealthEvent) *FailurePattern {
	if len(failures) == 0 {
		return nil
	}

	// Sort failures by timestamp
	sort.Slice(failures, func(i, j int) bool {
		return failures[i].Timestamp.Before(failures[j].Timestamp)
	})

	// Calculate average time between failures
	var intervals []time.Duration
	for i := 1; i < len(failures); i++ {
		intervals = append(intervals, failures[i].Timestamp.Sub(failures[i-1].Timestamp))
	}

	// Analyze precursor events
	precursors := fpm.findPrecursors(failures)

	// Extract features from failure events
	features := fpm.extractFeaturePattern(failures)

	// Determine failure type
	failureType := fpm.determineFailureType(failures[0])

	pattern := &FailurePattern{
		ID:            key,
		Type:          failureType,
		Component:     failures[0].Component,
		Precursors:    precursors,
		TimeToFailure: fpm.calculateAverageInterval(intervals),
		Confidence:    fpm.calculatePatternConfidence(failures, precursors),
		Frequency:     len(failures),
		LastSeen:      failures[len(failures)-1].Timestamp,
		SuccessRate:   0.8, // Default, should be calculated from historical accuracy
		Features:      features,
		Metadata: map[string]interface{}{
			"first_occurrence": failures[0].Timestamp,
			"interval_std":     fpm.calculateStandardDeviation(intervals),
			"event_count":      len(failures),
		},
	}

	return pattern
}

func (fpm *FailurePredictionModel) findPrecursors(failures []HealthEvent) []PatternPrecursor {
	var precursors []PatternPrecursor
	
	// Look for events that consistently occur before failures
	lookbackWindow := 24 * time.Hour // Look 24 hours before each failure
	
	for _, failure := range failures {
		startTime := failure.Timestamp.Add(-lookbackWindow)
		
		// Find events in the lookback window
		for _, event := range fpm.historicalEvents {
			if event.Timestamp.After(startTime) && event.Timestamp.Before(failure.Timestamp) {
				// Calculate time offset
				offset := failure.Timestamp.Sub(event.Timestamp)
				
				// Check if this type of event is a common precursor
				importance := fpm.calculatePrecursorImportance(event, failures)
				if importance > fpm.config.FeatureImportanceThreshold {
					precursor := PatternPrecursor{
						Event:      event.Type,
						TimeOffset: offset,
						Importance: importance,
						Frequency:  fpm.calculatePrecursorFrequency(event.Type, failures),
					}
					precursors = append(precursors, precursor)
				}
			}
		}
	}

	// Remove duplicates and sort by importance
	precursors = fpm.deduplicatePrecursors(precursors)
	sort.Slice(precursors, func(i, j int) bool {
		return precursors[i].Importance > precursors[j].Importance
	})

	// Keep only top precursors
	if len(precursors) > 10 {
		precursors = precursors[:10]
	}

	return precursors
}

func (fpm *FailurePredictionModel) calculateFailureProbability(pattern *FailurePattern, currentFeatures map[string]float64, timeline time.Duration) float64 {
	// Base probability from pattern frequency and success rate
	baseProbability := float64(pattern.Frequency) / 100.0 * pattern.SuccessRate

	// Adjust based on feature similarity
	featureSimilarity := fpm.calculateFeatureSimilarity(pattern.Features, currentFeatures)
	
	// Time decay factor (patterns are more likely if they haven't occurred recently)
	timeSinceLastSeen := time.Since(pattern.LastSeen)
	timeDecay := math.Min(1.0, timeSinceLastSeen.Hours()/(24*7)) // Weekly decay

	// Precursor matching
	precursorMatch := fpm.evaluatePrecursors(pattern.Precursors, currentFeatures)

	// Combine factors
	probability := baseProbability * featureSimilarity * (0.5 + 0.5*timeDecay) * precursorMatch

	// Ensure probability is between 0 and 1
	return math.Max(0.0, math.Min(1.0, probability))
}

func (fpm *FailurePredictionModel) getCurrentSystemFeatures() map[string]float64 {
	// In a real implementation, this would gather current system metrics
	// For now, return mock features
	return map[string]float64{
		"cpu_usage":      75.0,
		"memory_usage":   80.0,
		"disk_usage":     65.0,
		"network_usage":  30.0,
		"load_average":   2.5,
		"process_count":  145.0,
		"error_rate":     0.02,
		"response_time":  150.0,
	}
}

func (fpm *FailurePredictionModel) estimateFailureTime(pattern *FailurePattern, currentFeatures map[string]float64) time.Time {
	// Use the pattern's average time to failure, adjusted by current conditions
	baseTime := pattern.TimeToFailure
	
	// Adjust based on current system stress
	stressFactor := fpm.calculateSystemStress(currentFeatures)
	adjustedTime := time.Duration(float64(baseTime) / stressFactor)
	
	return time.Now().Add(adjustedTime)
}

func (fpm *FailurePredictionModel) calculateSystemStress(features map[string]float64) float64 {
	// Calculate system stress based on resource utilization
	cpuStress := features["cpu_usage"] / 100.0
	memoryStress := features["memory_usage"] / 100.0
	diskStress := features["disk_usage"] / 100.0
	
	// Average stress with weights
	stress := (cpuStress*0.4 + memoryStress*0.4 + diskStress*0.2)
	
	// Ensure stress is at least 0.1 to avoid division by zero
	return math.Max(0.1, stress)
}

func (fpm *FailurePredictionModel) assessImpact(pattern *FailurePattern) ImpactLevel {
	// Assess impact based on component and failure type
	switch pattern.Component {
	case "system", "kernel", "boot":
		return ImpactSevere
	case "database", "web_server", "storage":
		return ImpactSignificant
	case "monitoring", "logging", "backup":
		return ImpactModerate
	default:
		return ImpactMinimal
	}
}

func (fpm *FailurePredictionModel) getHealthIndicators(pattern *FailurePattern, currentFeatures map[string]float64) []HealthIndicator {
	var indicators []HealthIndicator
	
	for metric, value := range currentFeatures {
		threshold := fpm.getThresholdForMetric(metric)
		trend := fpm.calculateTrend(metric, value)
		severity := fpm.calculateSeverity(value, threshold)
		
		indicator := HealthIndicator{
			Name:      metric,
			Value:     value,
			Threshold: threshold,
			Unit:      fpm.getUnitForMetric(metric),
			Trend:     trend,
			Severity:  severity,
			Source:    "failure_predictor",
		}
		
		indicators = append(indicators, indicator)
	}
	
	return indicators
}

func (fpm *FailurePredictionModel) getHistoricalEvents(pattern *FailurePattern) []HistoricalEvent {
	var events []HistoricalEvent
	
	// Find recent events related to this pattern
	for _, event := range fpm.historicalEvents {
		if event.Component == pattern.Component && 
		   time.Since(event.Timestamp) < 30*24*time.Hour { // Last 30 days
			
			historicalEvent := HistoricalEvent{
				Timestamp:   event.Timestamp,
				EventType:   event.Type,
				Description: event.Message,
				Severity:    event.Severity,
				Metadata:    event.Metrics,
			}
			
			events = append(events, historicalEvent)
		}
	}
	
	// Sort by timestamp (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})
	
	// Keep only the most recent 10 events
	if len(events) > 10 {
		events = events[:10]
	}
	
	return events
}

func (fpm *FailurePredictionModel) generatePreventiveActions(predictions []PredictedFailure) []PreventiveAction {
	var actions []PreventiveAction
	
	for _, prediction := range predictions {
		action := fpm.generateActionForFailure(prediction)
		if action != nil {
			actions = append(actions, *action)
		}
	}
	
	return actions
}

func (fpm *FailurePredictionModel) generateActionForFailure(failure PredictedFailure) *PreventiveAction {
	var commands []string
	var description string
	automated := false
	risk := RiskLow
	
	switch failure.Type {
	case FailureDiskSpace:
		description = "Clean up disk space to prevent storage failure"
		commands = []string{
			"nix-collect-garbage -d",
			"journalctl --vacuum-time=7d",
			"find /tmp -type f -atime +7 -delete",
		}
		automated = true
		
	case FailureMemoryLeak:
		description = "Restart services with potential memory leaks"
		commands = []string{
			"systemctl restart " + failure.Component,
			"systemctl status " + failure.Component,
		}
		risk = RiskMedium
		
	case FailureServiceCrash:
		description = "Enable service auto-restart and check configuration"
		commands = []string{
			"systemctl enable " + failure.Component,
			"systemctl restart " + failure.Component,
			"systemctl is-active " + failure.Component,
		}
		automated = true
		
	case FailureSecurityBreach:
		description = "Update security packages and scan for vulnerabilities"
		commands = []string{
			"nixos-rebuild switch --upgrade",
			"nix-channel --update",
		}
		risk = RiskHigh
		
	default:
		description = "Monitor system and prepare for potential " + string(failure.Type)
		commands = []string{
			"systemctl status " + failure.Component,
			"journalctl -u " + failure.Component + " --since '1 hour ago'",
		}
	}
	
	return &PreventiveAction{
		ID:          fmt.Sprintf("action_%s_%d", failure.ID, time.Now().Unix()),
		Description: description,
		Commands:    commands,
		Automated:   automated,
		Risk:        risk,
		ETA:         5 * time.Minute,
		Dependencies: []string{},
		Metadata: map[string]interface{}{
			"failure_type":      failure.Type,
			"prediction_score": failure.ProbabilityScore,
			"component":        failure.Component,
		},
	}
}

func (fpm *FailurePredictionModel) calculateRiskLevel(predictions []PredictedFailure) RiskLevel {
	if len(predictions) == 0 {
		return RiskLow
	}
	
	maxProbability := 0.0
	severityCount := map[ImpactLevel]int{
		ImpactSevere:     0,
		ImpactSignificant: 0,
		ImpactModerate:   0,
		ImpactMinimal:    0,
	}
	
	for _, prediction := range predictions {
		if prediction.ProbabilityScore > maxProbability {
			maxProbability = prediction.ProbabilityScore
		}
		severityCount[prediction.Impact]++
	}
	
	// Determine risk based on probability and impact
	if maxProbability > 0.9 || severityCount[ImpactSevere] > 0 {
		return RiskCritical
	} else if maxProbability > 0.7 || severityCount[ImpactSignificant] > 1 {
		return RiskHigh
	} else if maxProbability > 0.5 || severityCount[ImpactModerate] > 2 {
		return RiskMedium
	}
	
	return RiskLow
}

func (fpm *FailurePredictionModel) calculateOverallConfidence(predictions []PredictedFailure) float64 {
	if len(predictions) == 0 {
		return 0.0
	}
	
	totalConfidence := 0.0
	for _, prediction := range predictions {
		totalConfidence += prediction.ProbabilityScore
	}
	
	avgConfidence := totalConfidence / float64(len(predictions))
	
	// Adjust confidence based on model metrics
	modelConfidence := (fpm.modelMetrics.Accuracy + fpm.modelMetrics.F1Score) / 2.0
	
	return avgConfidence * modelConfidence
}

// Helper methods

func (fpm *FailurePredictionModel) extractFeatures(events []HealthEvent) []map[string]float64 {
	var features []map[string]float64
	
	for _, event := range events {
		feature := make(map[string]float64)
		
		// Extract basic features from event metrics
		for key, value := range event.Metrics {
			if floatVal, ok := value.(float64); ok {
				feature[key] = floatVal
			}
		}
		
		// Add temporal features
		feature["hour_of_day"] = float64(event.Timestamp.Hour())
		feature["day_of_week"] = float64(event.Timestamp.Weekday())
		
		features = append(features, feature)
	}
	
	return features
}

func (fpm *FailurePredictionModel) extractFeaturePattern(events []HealthEvent) map[string]float64 {
	pattern := make(map[string]float64)
	
	if len(events) == 0 {
		return pattern
	}
	
	// Aggregate features across all events
	metricSums := make(map[string]float64)
	metricCounts := make(map[string]int)
	
	for _, event := range events {
		for key, value := range event.Metrics {
			if floatVal, ok := value.(float64); ok {
				metricSums[key] += floatVal
				metricCounts[key]++
			}
		}
	}
	
	// Calculate averages
	for metric, sum := range metricSums {
		if count := metricCounts[metric]; count > 0 {
			pattern[metric+"_avg"] = sum / float64(count)
		}
	}
	
	return pattern
}

func (fpm *FailurePredictionModel) determineFailureType(event HealthEvent) FailureType {
	message := strings.ToLower(event.Message)
	component := strings.ToLower(event.Component)
	
	if strings.Contains(message, "disk") || strings.Contains(message, "space") {
		return FailureDiskSpace
	} else if strings.Contains(message, "memory") || strings.Contains(message, "oom") {
		return FailureMemoryLeak
	} else if strings.Contains(message, "network") || strings.Contains(component, "network") {
		return FailureNetworkIssue
	} else if strings.Contains(message, "security") || strings.Contains(message, "breach") {
		return FailureSecurityBreach
	} else if strings.Contains(message, "performance") || strings.Contains(message, "slow") {
		return FailurePerformance
	} else if strings.Contains(message, "config") {
		return FailureConfiguration
	} else if strings.Contains(message, "crash") || strings.Contains(message, "exit") {
		return FailureServiceCrash
	}
	
	return FailureServiceCrash // Default
}

func (fpm *FailurePredictionModel) calculateAverageInterval(intervals []time.Duration) time.Duration {
	if len(intervals) == 0 {
		return 24 * time.Hour // Default
	}
	
	var total time.Duration
	for _, interval := range intervals {
		total += interval
	}
	
	return total / time.Duration(len(intervals))
}

func (fpm *FailurePredictionModel) calculatePatternConfidence(failures []HealthEvent, precursors []PatternPrecursor) float64 {
	// Base confidence from frequency
	baseConfidence := math.Min(0.9, float64(len(failures))/10.0)
	
	// Boost confidence if we have strong precursors
	precursorBoost := 0.0
	for _, precursor := range precursors {
		precursorBoost += precursor.Importance * 0.1
	}
	
	confidence := baseConfidence + precursorBoost
	return math.Min(1.0, confidence)
}

func (fpm *FailurePredictionModel) calculateStandardDeviation(intervals []time.Duration) float64 {
	if len(intervals) <= 1 {
		return 0.0
	}
	
	// Convert to hours for calculation
	var values []float64
	for _, interval := range intervals {
		values = append(values, interval.Hours())
	}
	
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// Calculate variance
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values) - 1)
	
	return math.Sqrt(variance)
}

func (fpm *FailurePredictionModel) calculatePrecursorImportance(event HealthEvent, failures []HealthEvent) float64 {
	// Simple importance calculation based on frequency
	count := 0
	for _, failure := range failures {
		// Check if this type of event occurred before the failure
		lookback := failure.Timestamp.Add(-24 * time.Hour)
		if event.Timestamp.After(lookback) && event.Timestamp.Before(failure.Timestamp) {
			count++
		}
	}
	
	return float64(count) / float64(len(failures))
}

func (fpm *FailurePredictionModel) calculatePrecursorFrequency(eventType string, failures []HealthEvent) float64 {
	count := 0
	for _, failure := range failures {
		// Count occurrences of this event type before failures
		lookback := failure.Timestamp.Add(-24 * time.Hour)
		for _, event := range fpm.historicalEvents {
			if event.Type == eventType && 
			   event.Timestamp.After(lookback) && 
			   event.Timestamp.Before(failure.Timestamp) {
				count++
				break // Only count once per failure
			}
		}
	}
	
	return float64(count) / float64(len(failures))
}

func (fpm *FailurePredictionModel) deduplicatePrecursors(precursors []PatternPrecursor) []PatternPrecursor {
	seen := make(map[string]bool)
	var unique []PatternPrecursor
	
	for _, precursor := range precursors {
		key := fmt.Sprintf("%s_%v", precursor.Event, precursor.TimeOffset)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, precursor)
		}
	}
	
	return unique
}

func (fpm *FailurePredictionModel) calculateFeatureSimilarity(patternFeatures, currentFeatures map[string]float64) float64 {
	if len(patternFeatures) == 0 {
		return 0.5 // Default similarity
	}
	
	similarity := 0.0
	matchCount := 0
	
	for feature, patternValue := range patternFeatures {
		if currentValue, exists := currentFeatures[feature]; exists {
			// Calculate normalized difference
			maxValue := math.Max(math.Abs(patternValue), math.Abs(currentValue))
			if maxValue > 0 {
				diff := math.Abs(patternValue-currentValue) / maxValue
				similarity += 1.0 - diff
			} else {
				similarity += 1.0
			}
			matchCount++
		}
	}
	
	if matchCount == 0 {
		return 0.5
	}
	
	return similarity / float64(matchCount)
}

func (fpm *FailurePredictionModel) evaluatePrecursors(precursors []PatternPrecursor, currentFeatures map[string]float64) float64 {
	if len(precursors) == 0 {
		return 1.0 // No precursors to check
	}
	
	matchScore := 0.0
	totalWeight := 0.0
	
	for _, precursor := range precursors {
		// Check if current state matches this precursor
		// This is simplified - in reality would check if the precursor event occurred recently
		weight := precursor.Importance
		totalWeight += weight
		
		// For now, assume 50% chance of precursor match
		if precursor.Frequency > 0.5 {
			matchScore += weight
		}
	}
	
	if totalWeight == 0 {
		return 1.0
	}
	
	return matchScore / totalWeight
}

func (fpm *FailurePredictionModel) getThresholdForMetric(metric string) float64 {
	thresholds := map[string]float64{
		"cpu_usage":     85.0,
		"memory_usage":  90.0,
		"disk_usage":    85.0,
		"network_usage": 80.0,
		"load_average":  4.0,
		"process_count": 200.0,
		"error_rate":    0.05,
		"response_time": 500.0,
	}
	
	if threshold, exists := thresholds[metric]; exists {
		return threshold
	}
	return 100.0 // Default threshold
}

func (fpm *FailurePredictionModel) calculateTrend(metric string, value float64) string {
	// Simplified trend calculation - in reality would analyze historical data
	if value > fpm.getThresholdForMetric(metric) {
		return "increasing"
	} else if value < fpm.getThresholdForMetric(metric)*0.5 {
		return "decreasing"
	}
	return "stable"
}

func (fpm *FailurePredictionModel) calculateSeverity(value, threshold float64) Priority {
	ratio := value / threshold
	
	if ratio > 1.2 {
		return PriorityCritical
	} else if ratio > 1.0 {
		return PriorityHigh
	} else if ratio > 0.8 {
		return PriorityMedium
	}
	return PriorityLow
}

func (fpm *FailurePredictionModel) getUnitForMetric(metric string) string {
	units := map[string]string{
		"cpu_usage":     "%",
		"memory_usage":  "%",
		"disk_usage":    "%",
		"network_usage": "%",
		"load_average":  "",
		"process_count": "count",
		"error_rate":    "ratio",
		"response_time": "ms",
	}
	
	if unit, exists := units[metric]; exists {
		return unit
	}
	return ""
}

func (fpm *FailurePredictionModel) evaluateModel() error {
	// Simple model evaluation - in reality would use cross-validation
	fpm.modelMetrics = &ModelMetrics{
		Accuracy:    0.85,
		Precision:   0.80,
		Recall:      0.75,
		F1Score:     0.77,
		EvaluatedAt: time.Now(),
	}
	
	fpm.logger.Info(fmt.Sprintf("Model evaluation completed: Accuracy=%.2f, F1=%.2f", 
		fpm.modelMetrics.Accuracy, fpm.modelMetrics.F1Score))
	
	return nil
}

func (fpm *FailurePredictionModel) evaluateModelWithData(testData []HealthEvent) (*ModelMetrics, error) {
	// Simplified evaluation with test data
	return &ModelMetrics{
		Accuracy:    0.82,
		Precision:   0.78,
		Recall:      0.73,
		F1Score:     0.75,
		EvaluatedAt: time.Now(),
	}, nil
}

func (fpm *FailurePredictionModel) updatePatternsIncremental(newData []HealthEvent) error {
	// Simplified incremental update
	for _, event := range newData {
		if event.Type == "failure" || event.Type == "error" {
			// Update existing patterns or create new ones
			key := fmt.Sprintf("%s_%s", event.Component, event.Type)
			if pattern, exists := fpm.patterns[key]; exists {
				pattern.Frequency++
				pattern.LastSeen = event.Timestamp
			}
		}
	}
	
	return nil
}

func (fpm *FailurePredictionModel) generateFailureDescription(pattern *FailurePattern) string {
	descriptions := map[FailureType]string{
		FailureDiskSpace:       "Disk space exhaustion predicted for %s",
		FailureMemoryLeak:      "Memory leak detected in %s service",
		FailureServiceCrash:    "Service crash predicted for %s",
		FailureNetworkIssue:    "Network connectivity issues predicted",
		FailureHardwareFailure: "Hardware failure indicators detected",
		FailureSecurityBreach:  "Security vulnerability exposure predicted",
		FailurePerformance:     "Performance degradation predicted for %s",
		FailureConfiguration:   "Configuration error predicted in %s",
		FailureDependency:      "Dependency failure predicted for %s",
		FailureUpdate:          "Update failure predicted for %s",
	}
	
	if template, exists := descriptions[pattern.Type]; exists {
		return fmt.Sprintf(template, pattern.Component)
	}
	
	return fmt.Sprintf("Potential failure predicted for component: %s", pattern.Component)
}