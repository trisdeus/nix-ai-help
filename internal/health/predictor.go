// Package health provides the main health prediction system
package health

import (
	"context"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/health/ml"
	"nix-ai-help/pkg/logger"
)

// SystemHealthPredictor implements the main health prediction system
type SystemHealthPredictor struct {
	mu                    sync.RWMutex
	logger                *logger.Logger
	config                *config.UserConfig
	healthConfig          *HealthConfig
	failurePredictionModel *ml.FailurePredictionModel
	anomalyDetector       *ml.AnomalyDetector
	resourceForecaster    *ml.ResourceForecaster
	systemMonitor         *SystemMonitor
	remediationEngine     *RemediationEngine
	eventHistory          []HealthEvent
	running               bool
	lastHealthCheck       time.Time
	currentHealthStatus   *HealthAssessment
	predictionCache       map[string]*FailurePrediction
	cacheExpiry           time.Duration
}

// SystemMonitor collects real-time system metrics
type SystemMonitor struct {
	mu               sync.RWMutex
	logger           *logger.Logger
	metrics          map[string]float64
	lastUpdate       time.Time
	updateInterval   time.Duration
	running          bool
	collectors       map[string]MetricCollector
}

// MetricCollector defines interface for collecting specific metrics
type MetricCollector interface {
	CollectMetrics() (map[string]float64, error)
	GetName() string
	IsHealthy() bool
}

// RemediationEngine generates and executes remediation suggestions
type RemediationEngine struct {
	mu                sync.RWMutex
	logger            *logger.Logger
	suggestionRules   map[FailureType][]RemediationRule
	executionHistory  []RemediationExecution
	autoExecutionEnabled bool
	riskThreshold     RiskLevel
}

// RemediationRule defines how to generate remediation suggestions
type RemediationRule struct {
	ID               string                 `json:"id"`
	FailureType      FailureType            `json:"failure_type"`
	Conditions       []RuleCondition        `json:"conditions"`
	Actions          []RemediationAction    `json:"actions"`
	Priority         Priority               `json:"priority"`
	AutoExecutable   bool                   `json:"auto_executable"`
	RiskLevel        RiskLevel              `json:"risk_level"`
	Description      string                 `json:"description"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RuleCondition defines when a remediation rule applies
type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "gt", "lt", "contains", "matches"
	Value    interface{} `json:"value"`
	Weight   float64     `json:"weight"`
}

// RemediationExecution tracks remediation action execution
type RemediationExecution struct {
	ID            string                 `json:"id"`
	SuggestionID  string                 `json:"suggestion_id"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Status        string                 `json:"status"` // "running", "completed", "failed", "rollback"
	ExecutedActions []ActionExecution    `json:"executed_actions"`
	Result        string                 `json:"result"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ActionExecution tracks individual action execution
type ActionExecution struct {
	Step        int           `json:"step"`
	Action      string        `json:"action"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Status      string        `json:"status"`
	Output      string        `json:"output,omitempty"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration"`
}

// NewSystemHealthPredictor creates a new system health predictor
func NewSystemHealthPredictor(cfg *config.UserConfig) *SystemHealthPredictor {
	healthConfig := &HealthConfig{
		MonitoringInterval:      5 * time.Minute,
		PredictionTimeline:      7 * 24 * time.Hour, // 7 days
		AnomalyThreshold:        0.8,
		EnableAutoRemediation:   false, // Start with manual mode
		SecurityScanInterval:    24 * time.Hour,
		ModelUpdateInterval:     6 * time.Hour,
		DataRetentionPeriod:     30 * 24 * time.Hour, // 30 days
		AlertingEnabled:         true,
		LogLevel:                "info",
		ResourceThresholds: map[string]ResourceThreshold{
			"cpu": {
				Warning:       75.0,
				Critical:      90.0,
				Maximum:       100.0,
				Unit:          "%",
				AlertsEnabled: true,
			},
			"memory": {
				Warning:       80.0,
				Critical:      95.0,
				Maximum:       100.0,
				Unit:          "%",
				AlertsEnabled: true,
			},
			"disk": {
				Warning:       80.0,
				Critical:      90.0,
				Maximum:       100.0,
				Unit:          "%",
				AlertsEnabled: true,
			},
		},
	}

	// Create ML components
	mlConfig := &ml.MLConfig{
		TrainingWindow:        30 * 24 * time.Hour,
		PredictionHorizon:     7 * 24 * time.Hour,
		MinTrainingEvents:     100,
		ModelUpdateInterval:   24 * time.Hour,
		FeatureImportanceThreshold: 0.1,
		AnomalyThreshold:      0.8,
		ConfidenceThreshold:   0.7,
		MaxPatterns:           1000,
	}

	failureModel := ml.NewFailurePredictionModel(mlConfig)
	anomalyDetector := ml.NewAnomalyDetector()
	resourceForecaster := ml.NewResourceForecaster()

	systemMonitor := NewSystemMonitor()
	remediationEngine := NewRemediationEngine()

	return &SystemHealthPredictor{
		logger:                logger.NewLogger(),
		config:                cfg,
		healthConfig:          healthConfig,
		failurePredictionModel: failureModel,
		anomalyDetector:       anomalyDetector,
		resourceForecaster:    resourceForecaster,
		systemMonitor:         systemMonitor,
		remediationEngine:     remediationEngine,
		eventHistory:          make([]HealthEvent, 0),
		predictionCache:       make(map[string]*FailurePrediction),
		cacheExpiry:           10 * time.Minute,
		currentHealthStatus: &HealthAssessment{
			OverallHealth:   HealthUnknown,
			ComponentHealth: make(map[string]HealthStatus),
			ActiveIssues:    make([]HealthIssue, 0),
			LastUpdate:      time.Now(),
		},
	}
}

// Start initializes and starts the health prediction system
func (shp *SystemHealthPredictor) Start(ctx context.Context) error {
	shp.mu.Lock()
	defer shp.mu.Unlock()

	if shp.running {
		return fmt.Errorf("health predictor already running")
	}

	shp.logger.Info("Starting system health predictor")

	// Start system monitor
	if err := shp.systemMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start system monitor: %w", err)
	}

	// Load historical data and train models
	if err := shp.loadHistoricalData(ctx); err != nil {
		shp.logger.Warn(fmt.Sprintf("Failed to load historical data: %v", err))
	}

	// Start background monitoring
	go shp.backgroundMonitoring(ctx)

	// Start periodic model updates
	go shp.periodicModelUpdates(ctx)

	// Start cache cleanup
	go shp.cacheCleanup(ctx)

	shp.running = true
	shp.logger.Info("System health predictor started successfully")

	return nil
}

// Stop gracefully stops the health prediction system
func (shp *SystemHealthPredictor) Stop() error {
	shp.mu.Lock()
	defer shp.mu.Unlock()

	if !shp.running {
		return nil
	}

	shp.logger.Info("Stopping system health predictor")

	// Stop system monitor
	if err := shp.systemMonitor.Stop(); err != nil {
		shp.logger.Error(fmt.Sprintf("Error stopping system monitor: %v", err))
	}

	shp.running = false
	shp.logger.Info("System health predictor stopped")

	return nil
}

// PredictFailures implements HealthPredictor interface
func (shp *SystemHealthPredictor) PredictFailures(ctx context.Context, timeline time.Duration) (*FailurePrediction, error) {
	shp.mu.RLock()
	defer shp.mu.RUnlock()

	// Check cache first
	cacheKey := fmt.Sprintf("failures_%v", timeline)
	if cached, exists := shp.predictionCache[cacheKey]; exists {
		if time.Since(cached.GeneratedAt) < shp.cacheExpiry {
			shp.logger.Info("Returning cached failure prediction")
			return cached, nil
		}
	}

	shp.logger.Info(fmt.Sprintf("Predicting failures for timeline: %v", timeline))

	// Generate prediction using ML model
	prediction, err := shp.failurePredictionModel.Predict(ctx, timeline)
	if err != nil {
		return nil, fmt.Errorf("failed to predict failures: %w", err)
	}

	mlPrediction, ok := prediction.(*ml.FailurePrediction)
	if !ok {
		return nil, fmt.Errorf("unexpected prediction type")
	}

	// Convert ML prediction to health prediction
	failurePrediction := convertFailurePredictionFromML(mlPrediction)

	// Enhance prediction with additional context
	shp.enhancePrediction(failurePrediction)

	// Cache the prediction
	shp.predictionCache[cacheKey] = failurePrediction

	shp.logger.Info(fmt.Sprintf("Generated failure prediction with %d predicted failures", 
		len(failurePrediction.PredictedFailures)))

	return failurePrediction, nil
}

// AnalyzeSystemHealth implements HealthPredictor interface
func (shp *SystemHealthPredictor) AnalyzeSystemHealth(ctx context.Context) (*HealthAssessment, error) {
	shp.mu.Lock()
	defer shp.mu.Unlock()

	shp.logger.Info("Analyzing current system health")

	// Force immediate metrics collection
	shp.systemMonitor.updateMetrics()
	
	// Get current metrics from system monitor
	currentMetrics := shp.systemMonitor.GetCurrentMetrics()

	// Detect anomalies
	mlAnomalyReport, err := shp.anomalyDetector.DetectAnomalies(currentMetrics)
	if err != nil {
		shp.logger.Error(fmt.Sprintf("Failed to detect anomalies: %v", err))
		mlAnomalyReport = &ml.AnomalyReport{DetectedAnomalies: []ml.Anomaly{}}
	}

	// Convert ML anomaly report to health anomaly report
	anomalyReport := convertAnomalyReportFromML(mlAnomalyReport)

	// Calculate component health status
	componentHealth := shp.calculateComponentHealth(currentMetrics)

	// Calculate overall health
	overallHealth := shp.calculateOverallHealth(componentHealth, anomalyReport)

	// Generate active issues from anomalies
	activeIssues := shp.generateHealthIssues(anomalyReport.DetectedAnomalies)

	// Calculate resource utilization
	resourceUtil := shp.calculateResourceUtilization(currentMetrics)

	// Get security status
	securityStatus := shp.getSecurityStatus()

	// Calculate trend analysis
	trendAnalysis := shp.calculateTrendAnalysis()

	// Generate recommendations
	recommendations := shp.generateRecommendations(activeIssues, currentMetrics)

	assessment := &HealthAssessment{
		OverallHealth:       overallHealth,
		ComponentHealth:     componentHealth,
		ActiveIssues:        activeIssues,
		PerformanceMetrics:  currentMetrics,
		ResourceUtilization: resourceUtil,
		SecurityStatus:      securityStatus,
		TrendAnalysis:       trendAnalysis,
		LastUpdate:          time.Now(),
		Recommendations:     recommendations,
	}

	// Update cached health status
	shp.currentHealthStatus = assessment
	shp.lastHealthCheck = time.Now()

	shp.logger.Info(fmt.Sprintf("System health analysis completed: %s overall health, %d active issues", 
		overallHealth, len(activeIssues)))

	return assessment, nil
}

// ForecastResources implements HealthPredictor interface
func (shp *SystemHealthPredictor) ForecastResources(ctx context.Context, timeline time.Duration) (*ResourceForecast, error) {
	shp.mu.RLock()
	defer shp.mu.RUnlock()

	shp.logger.Info(fmt.Sprintf("Forecasting resource usage for timeline: %v", timeline))

	// Generate forecast using resource forecaster
	mlForecast, err := shp.resourceForecaster.Forecast(timeline)
	if err != nil {
		return nil, fmt.Errorf("failed to forecast resources: %w", err)
	}

	// Convert ML forecast to health forecast
	forecast := convertResourceForecastFromML(mlForecast)

	// Enhance forecast with current context
	shp.enhanceForecast(forecast)

	shp.logger.Info(fmt.Sprintf("Resource forecast generated for %d metrics", len(forecast.Predictions)))

	return forecast, nil
}

// DetectAnomalies implements HealthPredictor interface
func (shp *SystemHealthPredictor) DetectAnomalies(ctx context.Context) (*AnomalyReport, error) {
	shp.mu.RLock()
	defer shp.mu.RUnlock()

	shp.logger.Info("Detecting system anomalies")

	// Get current metrics
	currentMetrics := shp.systemMonitor.GetCurrentMetrics()

	// Detect anomalies
	mlReport, err := shp.anomalyDetector.DetectAnomalies(currentMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to detect anomalies: %w", err)
	}

	// Convert ML report to health report
	report := convertAnomalyReportFromML(mlReport)

	shp.logger.Info(fmt.Sprintf("Anomaly detection completed: %d anomalies detected", 
		len(report.DetectedAnomalies)))

	return report, nil
}

// GetRemediationSuggestions implements HealthPredictor interface
func (shp *SystemHealthPredictor) GetRemediationSuggestions(ctx context.Context, issues []HealthIssue) (*RemediationPlan, error) {
	shp.mu.RLock()
	defer shp.mu.RUnlock()

	shp.logger.Info(fmt.Sprintf("Generating remediation suggestions for %d issues", len(issues)))

	// Generate remediation plan
	plan, err := shp.remediationEngine.GenerateRemediationPlan(issues)
	if err != nil {
		return nil, fmt.Errorf("failed to generate remediation plan: %w", err)
	}

	shp.logger.Info(fmt.Sprintf("Remediation plan generated with %d suggestions", len(plan.Suggestions)))

	return plan, nil
}

// GetHealthStatus returns the current cached health status
func (shp *SystemHealthPredictor) GetHealthStatus() *HealthAssessment {
	shp.mu.RLock()
	defer shp.mu.RUnlock()

	return shp.currentHealthStatus
}

// GetModelInfo returns information about the ML models
func (shp *SystemHealthPredictor) GetModelInfo() map[string]ModelInfo {
	shp.mu.RLock()
	defer shp.mu.RUnlock()

	info := make(map[string]ModelInfo)
	mlInfo := shp.failurePredictionModel.GetInfo()
	info["failure_prediction"] = ModelInfo{
		Name:         mlInfo.Name,
		Version:      mlInfo.Version,
		Type:         mlInfo.Type,
		Accuracy:     mlInfo.Accuracy,
		LastTrained:  mlInfo.LastTrained,
		DataPoints:   mlInfo.DataPoints,
		Features:     mlInfo.Features,
		Metadata:     mlInfo.Metadata,
	}

	return info
}

// Type conversion functions to bridge between health and ml types

func convertHealthEventToML(event HealthEvent) ml.HealthEvent {
	return ml.HealthEvent{
		ID:          event.ID,
		Type:        event.Type,
		Component:   event.Component,
		Timestamp:   event.Timestamp,
		Severity:    ml.Priority(event.Severity),
		Message:     event.Message,
		Metrics:     event.Metrics,
		Context:     event.Context,
		Resolution:  event.Resolution,
		ResolvedAt:  event.ResolvedAt,
	}
}

func convertHealthEventsToML(events []HealthEvent) []ml.HealthEvent {
	mlEvents := make([]ml.HealthEvent, len(events))
	for i, event := range events {
		mlEvents[i] = convertHealthEventToML(event)
	}
	return mlEvents
}

func convertFailurePredictionFromML(mlPred *ml.FailurePrediction) *FailurePrediction {
	if mlPred == nil {
		return nil
	}

	failures := make([]PredictedFailure, len(mlPred.PredictedFailures))
	for i, f := range mlPred.PredictedFailures {
		failures[i] = PredictedFailure{
			ID:               f.ID,
			Type:             FailureType(f.Type),
			Component:        f.Component,
			Description:      f.Description,
			ProbabilityScore: f.ProbabilityScore,
			EstimatedTime:    f.EstimatedTime,
			Impact:           ImpactLevel(f.Impact),
			Indicators:       convertHealthIndicatorsFromML(f.Indicators),
			HistoricalData:   convertHistoricalEventsFromML(f.HistoricalData),
			Metadata:         f.Metadata,
		}
	}

	actions := make([]PreventiveAction, len(mlPred.PreventiveActions))
	for i, a := range mlPred.PreventiveActions {
		actions[i] = PreventiveAction{
			ID:           a.ID,
			Description:  a.Description,
			Commands:     a.Commands,
			Automated:    a.Automated,
			Risk:         RiskLevel(a.Risk),
			ETA:          a.ETA,
			Dependencies: a.Dependencies,
			Metadata:     a.Metadata,
		}
	}

	return &FailurePrediction{
		Timeline:          mlPred.Timeline,
		PredictedFailures: failures,
		Confidence:        mlPred.Confidence,
		RiskLevel:         RiskLevel(mlPred.RiskLevel),
		PreventiveActions: actions,
		GeneratedAt:       mlPred.GeneratedAt,
		ModelVersion:      mlPred.ModelVersion,
		Metadata:          mlPred.Metadata,
	}
}

func convertHealthIndicatorsFromML(mlIndicators []ml.HealthIndicator) []HealthIndicator {
	indicators := make([]HealthIndicator, len(mlIndicators))
	for i, ind := range mlIndicators {
		indicators[i] = HealthIndicator{
			Name:      ind.Name,
			Value:     ind.Value,
			Threshold: ind.Threshold,
			Unit:      ind.Unit,
			Trend:     ind.Trend,
			Severity:  Priority(ind.Severity),
			Source:    ind.Source,
		}
	}
	return indicators
}

func convertHistoricalEventsFromML(mlEvents []ml.HistoricalEvent) []HistoricalEvent {
	events := make([]HistoricalEvent, len(mlEvents))
	for i, evt := range mlEvents {
		events[i] = HistoricalEvent{
			Timestamp:   evt.Timestamp,
			EventType:   evt.EventType,
			Description: evt.Description,
			Severity:    Priority(evt.Severity),
			Resolution:  evt.Resolution,
			Metadata:    evt.Metadata,
		}
	}
	return events
}

func convertAnomalyReportFromML(mlReport *ml.AnomalyReport) *AnomalyReport {
	if mlReport == nil {
		return nil
	}

	anomalies := make([]Anomaly, len(mlReport.DetectedAnomalies))
	for i, a := range mlReport.DetectedAnomalies {
		evidence := make([]AnomalyEvidence, len(a.Evidence))
		for j, e := range a.Evidence {
			evidence[j] = AnomalyEvidence{
				Metric:        e.Metric,
				ExpectedValue: e.ExpectedValue,
				ActualValue:   e.ActualValue,
				Deviation:     e.Deviation,
				Timestamp:     e.Timestamp,
				Confidence:    e.Confidence,
			}
		}

		anomalies[i] = Anomaly{
			ID:          a.ID,
			Type:        a.Type,
			Component:   a.Component,
			Description: a.Description,
			Score:       a.Score,
			Severity:    Priority(a.Severity),
			DetectedAt:  a.DetectedAt,
			Evidence:    evidence,
			Context:     a.Context,
			Status:      a.Status,
		}
	}

	return &AnomalyReport{
		DetectedAnomalies: anomalies,
		AnomalyScore:      mlReport.AnomalyScore,
		BaselineDeviation: mlReport.BaselineDeviation,
		DetectionModel:    mlReport.DetectionModel,
		TimeWindow:        mlReport.TimeWindow,
		GeneratedAt:       mlReport.GeneratedAt,
		Metadata:          mlReport.Metadata,
	}
}

func convertResourceForecastFromML(mlForecast *ml.ResourceForecast) *ResourceForecast {
	if mlForecast == nil {
		return nil
	}

	cpuForecast := make([]ResourceDataPoint, len(mlForecast.CPUForecast))
	for i, dp := range mlForecast.CPUForecast {
		cpuForecast[i] = ResourceDataPoint{
			Timestamp:  dp.Timestamp,
			Value:      dp.Value,
			Predicted:  dp.Predicted,
			Confidence: dp.Confidence,
		}
	}

	memoryForecast := make([]ResourceDataPoint, len(mlForecast.MemoryForecast))
	for i, dp := range mlForecast.MemoryForecast {
		memoryForecast[i] = ResourceDataPoint{
			Timestamp:  dp.Timestamp,
			Value:      dp.Value,
			Predicted:  dp.Predicted,
			Confidence: dp.Confidence,
		}
	}

	diskForecast := make([]ResourceDataPoint, len(mlForecast.DiskForecast))
	for i, dp := range mlForecast.DiskForecast {
		diskForecast[i] = ResourceDataPoint{
			Timestamp:  dp.Timestamp,
			Value:      dp.Value,
			Predicted:  dp.Predicted,
			Confidence: dp.Confidence,
		}
	}

	networkForecast := make([]ResourceDataPoint, len(mlForecast.NetworkForecast))
	for i, dp := range mlForecast.NetworkForecast {
		networkForecast[i] = ResourceDataPoint{
			Timestamp:  dp.Timestamp,
			Value:      dp.Value,
			Predicted:  dp.Predicted,
			Confidence: dp.Confidence,
		}
	}

	predictions := make(map[string]ResourcePrediction)
	for k, p := range mlForecast.Predictions {
		predictions[k] = ResourcePrediction{
			Resource:         p.Resource,
			CurrentValue:     p.CurrentValue,
			PredictedValue:   p.PredictedValue,
			ChangeRate:       p.ChangeRate,
			TimeToThreshold:  p.TimeToThreshold,
			Confidence:       p.Confidence,
			Model:            p.Model,
			Metadata:         p.Metadata,
		}
	}

	thresholds := make(map[string]ResourceThreshold)
	for k, t := range mlForecast.Thresholds {
		thresholds[k] = ResourceThreshold{
			Warning:       t.Warning,
			Critical:      t.Critical,
			Maximum:       t.Maximum,
			Unit:          t.Unit,
			AlertsEnabled: t.AlertsEnabled,
		}
	}

	alerts := make([]ResourceAlert, len(mlForecast.Alerts))
	for i, a := range mlForecast.Alerts {
		alerts[i] = ResourceAlert{
			ID:            a.ID,
			Resource:      a.Resource,
			Type:          a.Type,
			Message:       a.Message,
			Threshold:     a.Threshold,
			CurrentValue:  a.CurrentValue,
			PredictedTime: a.PredictedTime,
			CreatedAt:     a.CreatedAt,
			Acknowledged:  a.Acknowledged,
		}
	}

	return &ResourceForecast{
		Timeline:        mlForecast.Timeline,
		CPUForecast:     cpuForecast,
		MemoryForecast:  memoryForecast,
		DiskForecast:    diskForecast,
		NetworkForecast: networkForecast,
		Predictions:     predictions,
		Thresholds:      thresholds,
		Alerts:          alerts,
		ModelAccuracy:   mlForecast.ModelAccuracy,
		Confidence:      mlForecast.Confidence,
		GeneratedAt:     mlForecast.GeneratedAt,
		Metadata:        make(map[string]interface{}),
	}
}

// UpdateModels forces an update of the ML models
func (shp *SystemHealthPredictor) UpdateModels(ctx context.Context) error {
	shp.mu.Lock()
	defer shp.mu.Unlock()

	shp.logger.Info("Updating ML models with recent data")

	// Get recent events for training
	recentEvents := shp.getRecentEvents(shp.healthConfig.ModelUpdateInterval)

	if len(recentEvents) < 10 {
		shp.logger.Warn("Insufficient recent events for model update")
		return nil
	}

	// Convert events to ML format and update failure prediction model
	mlEvents := convertHealthEventsToML(recentEvents)
	if err := shp.failurePredictionModel.Update(ctx, mlEvents); err != nil {
		return fmt.Errorf("failed to update failure prediction model: %w", err)
	}

	shp.logger.Info("ML models updated successfully")
	return nil
}

// Private methods

func (shp *SystemHealthPredictor) loadHistoricalData(ctx context.Context) error {
	// In a real implementation, this would load from persistent storage
	shp.logger.Info("Loading historical health data")

	// Generate some mock historical data for training
	mockEvents := shp.generateMockHistoricalData()
	shp.eventHistory = append(shp.eventHistory, mockEvents...)

	// Train models with historical data
	if len(shp.eventHistory) >= 100 {
		mlEvents := convertHealthEventsToML(shp.eventHistory)
		if err := shp.failurePredictionModel.Train(ctx, mlEvents); err != nil {
			return fmt.Errorf("failed to train failure prediction model: %w", err)
		}
	}

	shp.logger.Info(fmt.Sprintf("Loaded %d historical events", len(shp.eventHistory)))
	return nil
}

func (shp *SystemHealthPredictor) backgroundMonitoring(ctx context.Context) {
	ticker := time.NewTicker(shp.healthConfig.MonitoringInterval)
	defer ticker.Stop()

	shp.logger.Info("Starting background health monitoring")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := shp.performHealthCheck(ctx); err != nil {
				shp.logger.Error(fmt.Sprintf("Health check failed: %v", err))
			}
		}
	}
}

func (shp *SystemHealthPredictor) periodicModelUpdates(ctx context.Context) {
	ticker := time.NewTicker(shp.healthConfig.ModelUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := shp.UpdateModels(ctx); err != nil {
				shp.logger.Error(fmt.Sprintf("Model update failed: %v", err))
			}
		}
	}
}

func (shp *SystemHealthPredictor) cacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			shp.cleanupExpiredCache()
		}
	}
}

func (shp *SystemHealthPredictor) performHealthCheck(ctx context.Context) error {
	// Record current system state as a health event
	currentMetrics := shp.systemMonitor.GetCurrentMetrics()

	event := HealthEvent{
		ID:        fmt.Sprintf("health_check_%d", time.Now().Unix()),
		Type:      "health_check",
		Component: "system",
		Timestamp: time.Now(),
		Severity:  PriorityLow,
		Message:   "Periodic health check",
		Metrics:   make(map[string]interface{}),
		Context:   make(map[string]interface{}),
	}

	// Convert metrics to interface{} map
	for k, v := range currentMetrics {
		event.Metrics[k] = v
	}

	// Add to event history
	shp.mu.Lock()
	shp.eventHistory = append(shp.eventHistory, event)
	
	// Trim history if too large
	if len(shp.eventHistory) > 10000 {
		shp.eventHistory = shp.eventHistory[len(shp.eventHistory)-10000:]
	}
	shp.mu.Unlock()

	// Update current health status
	_, err := shp.AnalyzeSystemHealth(ctx)
	return err
}

func (shp *SystemHealthPredictor) enhancePrediction(prediction *FailurePrediction) {
	// Add additional context to prediction
	prediction.Metadata["system_info"] = map[string]interface{}{
		"hostname":     "nixos-system",
		"kernel":       "linux",
		"uptime":       "72h",
		"architecture": "x86_64",
	}

	// Enhance preventive actions with NixOS-specific commands
	for i := range prediction.PreventiveActions {
		action := &prediction.PreventiveActions[i]
		action.Metadata["platform"] = "nixos"
		
		// Add NixOS-specific commands
		switch action.Description {
		case "Clean up disk space to prevent storage failure":
			action.Commands = append(action.Commands, "nix-store --gc", "nix-store --optimize")
		case "Update security packages and scan for vulnerabilities":
			action.Commands = append(action.Commands, "nixos-rebuild switch --upgrade-all")
		}
	}
}

func (shp *SystemHealthPredictor) enhanceForecast(forecast *ResourceForecast) {
	// Add system-specific thresholds and metadata
	forecast.Metadata = map[string]interface{}{
		"system_type":    "nixos",
		"prediction_method": "arima",
		"monitoring_interval": shp.healthConfig.MonitoringInterval.String(),
	}
}

func (shp *SystemHealthPredictor) calculateComponentHealth(metrics map[string]float64) map[string]HealthStatus {
	componentHealth := make(map[string]HealthStatus)

	// Define component mappings
	components := map[string][]string{
		"cpu":     {"cpu_usage", "load_average"},
		"memory":  {"memory_usage"},
		"disk":    {"disk_usage"},
		"network": {"network_usage"},
		"system":  {"process_count"},
	}

	for component, metricList := range components {
		status := shp.calculateSingleComponentHealth(metrics, metricList)
		componentHealth[component] = status
	}

	return componentHealth
}

func (shp *SystemHealthPredictor) calculateSingleComponentHealth(metrics map[string]float64, metricList []string) HealthStatus {
	criticalCount := 0
	warningCount := 0
	totalCount := 0

	for _, metric := range metricList {
		if value, exists := metrics[metric]; exists {
			threshold := shp.getMetricThreshold(metric)
			totalCount++

			if value > threshold.Critical {
				criticalCount++
			} else if value > threshold.Warning {
				warningCount++
			}
		}
	}

	if totalCount == 0 {
		return HealthUnknown
	}

	criticalRatio := float64(criticalCount) / float64(totalCount)
	warningRatio := float64(warningCount) / float64(totalCount)

	if criticalRatio > 0.5 {
		return HealthCritical
	} else if criticalRatio > 0 || warningRatio > 0.7 {
		return HealthPoor
	} else if warningRatio > 0.3 {
		return HealthFair
	} else if warningRatio > 0 {
		return HealthGood
	}

	return HealthExcellent
}

func (shp *SystemHealthPredictor) calculateOverallHealth(componentHealth map[string]HealthStatus, anomalyReport *AnomalyReport) HealthStatus {
	// Count health levels
	statusCounts := make(map[HealthStatus]int)
	for _, status := range componentHealth {
		statusCounts[status]++
	}

	// Factor in anomalies
	criticalAnomalies := 0
	for _, anomaly := range anomalyReport.DetectedAnomalies {
		if anomaly.Severity == PriorityCritical {
			criticalAnomalies++
		}
	}

	// Determine overall health
	if statusCounts[HealthCritical] > 0 || criticalAnomalies > 0 {
		return HealthCritical
	} else if statusCounts[HealthPoor] > 1 {
		return HealthPoor
	} else if statusCounts[HealthPoor] > 0 || statusCounts[HealthFair] > 2 {
		return HealthFair
	} else if statusCounts[HealthFair] > 0 || len(anomalyReport.DetectedAnomalies) > 0 {
		return HealthGood
	}

	return HealthExcellent
}

func (shp *SystemHealthPredictor) generateHealthIssues(anomalies []Anomaly) []HealthIssue {
	var issues []HealthIssue

	for _, anomaly := range anomalies {
		issue := HealthIssue{
			ID:          fmt.Sprintf("issue_%s", anomaly.ID),
			Type:        "anomaly",
			Component:   anomaly.Component,
			Description: anomaly.Description,
			Severity:    anomaly.Severity,
			Status:      "active",
			DetectedAt:  anomaly.DetectedAt,
			Indicators: []HealthIndicator{
				{
					Name:      anomaly.Component,
					Value:     anomaly.Score,
					Threshold: 1.0,
					Unit:      "score",
					Trend:     "increasing",
					Severity:  anomaly.Severity,
					Source:    "anomaly_detector",
				},
			},
			Suggestions: []string{
				fmt.Sprintf("Investigate %s anomaly", anomaly.Component),
				"Check system logs for related errors",
				"Monitor component for continued anomalous behavior",
			},
			Metadata: map[string]interface{}{
				"anomaly_id":    anomaly.ID,
				"anomaly_score": anomaly.Score,
				"detection_time": anomaly.DetectedAt,
			},
		}

		issues = append(issues, issue)
	}

	return issues
}

func (shp *SystemHealthPredictor) calculateResourceUtilization(metrics map[string]float64) ResourceUtilization {
	return ResourceUtilization{
		CPU: ResourceMetric{
			Current:     metrics["cpu_usage"],
			Average:     metrics["cpu_usage"], // Simplified
			Peak:        metrics["cpu_usage"] * 1.2,
			Unit:        "%",
			Threshold:   75.0,
			Status:      shp.getMetricStatus(metrics["cpu_usage"], 75.0),
			LastUpdated: time.Now(),
		},
		Memory: ResourceMetric{
			Current:     metrics["memory_usage"],
			Average:     metrics["memory_usage"],
			Peak:        metrics["memory_usage"] * 1.1,
			Unit:        "%",
			Threshold:   80.0,
			Status:      shp.getMetricStatus(metrics["memory_usage"], 80.0),
			LastUpdated: time.Now(),
		},
		Disk: ResourceMetric{
			Current:     metrics["disk_usage"],
			Average:     metrics["disk_usage"],
			Peak:        metrics["disk_usage"],
			Unit:        "%",
			Threshold:   85.0,
			Status:      shp.getMetricStatus(metrics["disk_usage"], 85.0),
			LastUpdated: time.Now(),
		},
		Network: ResourceMetric{
			Current:     metrics["network_usage"],
			Average:     metrics["network_usage"],
			Peak:        metrics["network_usage"] * 1.5,
			Unit:        "%",
			Threshold:   70.0,
			Status:      shp.getMetricStatus(metrics["network_usage"], 70.0),
			LastUpdated: time.Now(),
		},
		LoadAvg:   []float64{metrics["load_average"], metrics["load_average"], metrics["load_average"]},
		Processes: int(metrics["process_count"]),
		UpdatedAt: time.Now(),
	}
}

func (shp *SystemHealthPredictor) getSecurityStatus() SecurityStatus {
	// Simplified security status
	return SecurityStatus{
		ThreatLevel:        "low",
		VulnerabilityCount: 0,
		LastSecurityScan:   time.Now().Add(-12 * time.Hour),
		ActiveThreats:      []SecurityThreat{},
		SecurityScore:      95.0,
		Recommendations: []SecurityAction{
			{
				Action:      "update_packages",
				Description: "Update system packages to latest versions",
				Priority:    PriorityMedium,
				Automated:   true,
				Commands:    []string{"nixos-rebuild switch --upgrade"},
			},
		},
	}
}

func (shp *SystemHealthPredictor) calculateTrendAnalysis() TrendAnalysis {
	// Simplified trend analysis
	return TrendAnalysis{
		CPUTrend: TrendData{
			Direction:  "stable",
			Rate:       0.1,
			Confidence: 0.8,
			Prediction: 75.0,
			Volatility: 0.2,
		},
		MemoryTrend: TrendData{
			Direction:  "increasing",
			Rate:       0.5,
			Confidence: 0.7,
			Prediction: 82.0,
			Volatility: 0.3,
		},
		DiskTrend: TrendData{
			Direction:  "increasing",
			Rate:       0.2,
			Confidence: 0.9,
			Prediction: 67.0,
			Volatility: 0.1,
		},
		NetworkTrend: TrendData{
			Direction:  "stable",
			Rate:       0.05,
			Confidence: 0.6,
			Prediction: 30.0,
			Volatility: 0.4,
		},
		PerformanceTrend: TrendData{
			Direction:  "stable",
			Rate:       0.0,
			Confidence: 0.8,
			Prediction: 85.0,
			Volatility: 0.15,
		},
		AnalysisPeriod: 24 * time.Hour,
		GeneratedAt:    time.Now(),
	}
}

func (shp *SystemHealthPredictor) generateRecommendations(issues []HealthIssue, metrics map[string]float64) []Recommendation {
	var recommendations []Recommendation

	// Generate recommendations based on issues
	for _, issue := range issues {
		rec := Recommendation{
			ID:          fmt.Sprintf("rec_%s", issue.ID),
			Type:        "issue_resolution",
			Title:       fmt.Sprintf("Resolve %s issue", issue.Type),
			Description: fmt.Sprintf("Address the %s anomaly in %s", issue.Type, issue.Component),
			Priority:    issue.Severity,
			Category:    "anomaly_resolution",
			Actions:     issue.Suggestions,
			Benefits:    []string{"Improve system stability", "Prevent potential failures"},
			Risks:       []string{"Minimal risk"},
			Effort:      "low",
		}
		recommendations = append(recommendations, rec)
	}

	// Generate proactive recommendations based on metrics
	if metrics["disk_usage"] > 75.0 {
		rec := Recommendation{
			ID:          "rec_disk_cleanup",
			Type:        "proactive_maintenance",
			Title:       "Disk Space Cleanup",
			Description: "Disk usage is approaching warning threshold",
			Priority:    PriorityMedium,
			Category:    "maintenance",
			Actions:     []string{"Run nix-collect-garbage", "Clean temporary files", "Archive old logs"},
			Benefits:    []string{"Free up disk space", "Prevent disk full errors"},
			Risks:       []string{"Low risk of removing needed data"},
			Effort:      "low",
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

func (shp *SystemHealthPredictor) getMetricThreshold(metric string) ResourceThreshold {
	if threshold, exists := shp.healthConfig.ResourceThresholds[metric]; exists {
		return threshold
	}

	// Default thresholds
	defaultThresholds := map[string]ResourceThreshold{
		"cpu_usage":     {Warning: 75.0, Critical: 90.0, Maximum: 100.0, Unit: "%"},
		"memory_usage":  {Warning: 80.0, Critical: 95.0, Maximum: 100.0, Unit: "%"},
		"disk_usage":    {Warning: 80.0, Critical: 90.0, Maximum: 100.0, Unit: "%"},
		"network_usage": {Warning: 70.0, Critical: 85.0, Maximum: 100.0, Unit: "%"},
		"process_count": {Warning: 1200.0, Critical: 2000.0, Maximum: 4000.0, Unit: "count"},
	}

	// Handle load_average specially - make it CPU-aware
	if metric == "load_average" {
		numCPU := float64(runtime.NumCPU())
		return ResourceThreshold{
			Warning:  numCPU * 0.75,    // 75% of CPU cores
			Critical: numCPU * 1.5,     // 150% of CPU cores (some oversubscription is OK)
			Maximum:  numCPU * 3.0,     // 300% of CPU cores
			Unit:     "",
		}
	}

	if threshold, exists := defaultThresholds[metric]; exists {
		return threshold
	}

	return ResourceThreshold{Warning: 75.0, Critical: 90.0, Maximum: 100.0, Unit: ""}
}

func (shp *SystemHealthPredictor) getMetricStatus(value, threshold float64) string {
	if value > threshold*1.2 {
		return "critical"
	} else if value > threshold {
		return "warning"
	} else if value > threshold*0.8 {
		return "normal"
	}
	return "good"
}

func (shp *SystemHealthPredictor) getRecentEvents(duration time.Duration) []HealthEvent {
	cutoff := time.Now().Add(-duration)
	var recentEvents []HealthEvent

	for _, event := range shp.eventHistory {
		if event.Timestamp.After(cutoff) {
			recentEvents = append(recentEvents, event)
		}
	}

	return recentEvents
}

func (shp *SystemHealthPredictor) cleanupExpiredCache() {
	shp.mu.Lock()
	defer shp.mu.Unlock()

	for key, prediction := range shp.predictionCache {
		if time.Since(prediction.GeneratedAt) > shp.cacheExpiry {
			delete(shp.predictionCache, key)
		}
	}
}

func (shp *SystemHealthPredictor) generateMockHistoricalData() []HealthEvent {
	var events []HealthEvent

	// Generate mock events for training
	for i := 0; i < 200; i++ {
		timestamp := time.Now().Add(-time.Duration(i) * time.Hour)
		
		event := HealthEvent{
			ID:        fmt.Sprintf("mock_event_%d", i),
			Type:      "metric_collection",
			Component: "system",
			Timestamp: timestamp,
			Severity:  PriorityLow,
			Message:   "Mock historical data point",
			Metrics: map[string]interface{}{
				"cpu_usage":     50.0 + float64(i%30),
				"memory_usage":  60.0 + float64(i%25),
				"disk_usage":    40.0 + float64(i%20),
				"network_usage": 20.0 + float64(i%15),
				"load_average":  1.0 + float64(i%4),
				"process_count": 100.0 + float64(i%50),
			},
			Context: make(map[string]interface{}),
		}

		// Occasionally add failure events
		if i%20 == 0 {
			event.Type = "failure"
			event.Severity = PriorityHigh
			event.Message = "Mock failure event"
		}

		events = append(events, event)
	}

	return events
}

// NewSystemMonitor creates a new system monitor
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		logger:         logger.NewLogger(),
		metrics:        make(map[string]float64),
		updateInterval: 30 * time.Second,
		collectors:     make(map[string]MetricCollector),
	}
}

// Start starts the system monitor
func (sm *SystemMonitor) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("system monitor already running")
	}

	sm.logger.Info("Starting system monitor")

	// Start metric collection
	go sm.collectMetrics(ctx)

	sm.running = true
	return nil
}

// Stop stops the system monitor
func (sm *SystemMonitor) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.running = false
	return nil
}

// GetCurrentMetrics returns current system metrics
func (sm *SystemMonitor) GetCurrentMetrics() map[string]float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]float64)
	for k, v := range sm.metrics {
		result[k] = v
	}
	return result
}

func (sm *SystemMonitor) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(sm.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sm.updateMetrics()
		}
	}
}

func (sm *SystemMonitor) updateMetrics() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Collect real system metrics
	sm.metrics = sm.collectRealSystemMetrics()
	sm.lastUpdate = time.Now()
}

// collectRealSystemMetrics gathers actual system metrics from /proc and other sources
func (sm *SystemMonitor) collectRealSystemMetrics() map[string]float64 {
	metrics := make(map[string]float64)
	
	// CPU Usage
	cpuUsage := sm.getCPUUsage()
	metrics["cpu_usage"] = cpuUsage
	sm.logger.Info(fmt.Sprintf("Collected CPU usage: %.2f%%", cpuUsage))
	
	// Memory Usage
	memUsage := sm.getMemoryUsage()
	metrics["memory_usage"] = memUsage
	sm.logger.Info(fmt.Sprintf("Collected memory usage: %.2f%%", memUsage))
	
	// Disk Usage
	diskUsage := sm.getDiskUsage()
	metrics["disk_usage"] = diskUsage
	sm.logger.Info(fmt.Sprintf("Collected disk usage: %.2f%%", diskUsage))
	
	// Load Average
	loadAvg := sm.getLoadAverage()
	metrics["load_average"] = loadAvg
	sm.logger.Info(fmt.Sprintf("Collected load average: %.2f", loadAvg))
	
	// Process Count
	processCount := sm.getProcessCount()
	metrics["process_count"] = processCount
	sm.logger.Info(fmt.Sprintf("Collected process count: %.0f", processCount))
	
	// Network Usage (simplified)
	netUsage := sm.getNetworkUsage()
	metrics["network_usage"] = netUsage
	sm.logger.Info(fmt.Sprintf("Collected network usage: %.2f", netUsage))
	
	// Error rate and response time (mock for now as these need specific monitoring)
	metrics["error_rate"] = 0.01
	metrics["response_time"] = 50.0
	
	return metrics
}

// getCPUUsage calculates current CPU usage percentage
func (sm *SystemMonitor) getCPUUsage() float64 {
	// Read /proc/stat for CPU times
	data, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		sm.logger.Warn(fmt.Sprintf("Failed to read /proc/stat: %v", err))
		return 0.0
	}
	
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0.0
	}
	
	// Parse first line which contains overall CPU stats
	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0.0
	}
	
	// Calculate CPU usage (simplified)
	var idle, total float64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			continue
		}
		total += val
		if i == 4 { // idle time is the 4th field
			idle = val
		}
	}
	
	if total == 0 {
		return 0.0
	}
	
	usage := (total - idle) / total * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	
	return usage
}

// getMemoryUsage calculates current memory usage percentage
func (sm *SystemMonitor) getMemoryUsage() float64 {
	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		sm.logger.Warn(fmt.Sprintf("Failed to read /proc/meminfo: %v", err))
		return 0.0
	}
	
	lines := strings.Split(string(data), "\n")
	var memTotal, memFree, memAvailable float64
	
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				memTotal = val
			}
		case "MemFree:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				memFree = val
			}
		case "MemAvailable:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				memAvailable = val
			}
		}
	}
	
	if memTotal == 0 {
		return 0.0
	}
	
	// Use MemAvailable if available, otherwise use MemFree
	available := memAvailable
	if available == 0 {
		available = memFree
	}
	
	usage := (memTotal - available) / memTotal * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	
	return usage
}

// getDiskUsage calculates disk usage for root partition
func (sm *SystemMonitor) getDiskUsage() float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		sm.logger.Warn(fmt.Sprintf("Failed to get disk stats: %v", err))
		return 0.0
	}
	
	// Calculate disk usage percentage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	
	if total == 0 {
		return 0.0
	}
	
	usage := float64(used) / float64(total) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	
	return usage
}

// getLoadAverage gets the 1-minute load average
func (sm *SystemMonitor) getLoadAverage() float64 {
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		sm.logger.Warn(fmt.Sprintf("Failed to read /proc/loadavg: %v", err))
		return 0.0
	}
	
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0.0
	}
	
	loadAvg, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0.0
	}
	
	return loadAvg
}

// getProcessCount counts the number of running processes
func (sm *SystemMonitor) getProcessCount() float64 {
	// Count actual running processes by counting PID directories in /proc
	entries, err := ioutil.ReadDir("/proc")
	if err != nil {
		return 0.0
	}
	
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if directory name is numeric (PID)
			if _, err := strconv.Atoi(entry.Name()); err == nil {
				count++
			}
		}
	}
	
	return float64(count)
}

// getNetworkUsage calculates approximate network usage
func (sm *SystemMonitor) getNetworkUsage() float64 {
	// Read network interface statistics
	data, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return 0.0
	}
	
	lines := strings.Split(string(data), "\n")
	var totalRxBytes, totalTxBytes float64
	activeInterfaces := 0
	
	for _, line := range lines {
		// Skip header lines and loopback interface
		if strings.Contains(line, ":") && !strings.Contains(line, "lo:") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				// RX bytes (field 1) + TX bytes (field 9)
				if rxBytes, err := strconv.ParseFloat(fields[1], 64); err == nil {
					totalRxBytes += rxBytes
				}
				if txBytes, err := strconv.ParseFloat(fields[9], 64); err == nil {
					totalTxBytes += txBytes
				}
				activeInterfaces++
			}
		}
	}
	
	// Since we can't easily measure current throughput without storing previous values,
	// we'll use a heuristic based on interface activity
	// For now, return a low utilization unless there's significant accumulated traffic
	totalBytes := totalRxBytes + totalTxBytes
	
	// If interfaces exist but have very little traffic, assume low usage
	if activeInterfaces == 0 {
		return 0.0
	}
	
	// Heuristic: Very rough estimate based on accumulated traffic
	// This is not ideal but better than always returning 100%
	// In a real implementation, this would track rates over time
	if totalBytes < 1024*1024*100 { // Less than 100MB total
		return 5.0  // Low usage
	} else if totalBytes < 1024*1024*1024 { // Less than 1GB total
		return 15.0 // Moderate usage
	} else if totalBytes < 1024*1024*1024*10 { // Less than 10GB total
		return 35.0 // Higher usage
	} else {
		return 50.0 // High usage (but not critical)
	}
}

// NewRemediationEngine creates a new remediation engine
func NewRemediationEngine() *RemediationEngine {
	engine := &RemediationEngine{
		logger:               logger.NewLogger(),
		suggestionRules:      make(map[FailureType][]RemediationRule),
		executionHistory:     make([]RemediationExecution, 0),
		autoExecutionEnabled: false,
		riskThreshold:        RiskMedium,
	}

	engine.initializeDefaultRules()
	return engine
}

// GenerateRemediationPlan generates a remediation plan for given issues
func (re *RemediationEngine) GenerateRemediationPlan(issues []HealthIssue) (*RemediationPlan, error) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.logger.Info(fmt.Sprintf("Generating remediation plan for %d issues", len(issues)))

	var suggestions []RemediationSuggestion
	overallRisk := RiskLow
	estimatedTime := time.Duration(0)

	for _, issue := range issues {
		// Generate suggestions for this issue
		issueSuggestions := re.generateSuggestionsForIssue(issue)
		suggestions = append(suggestions, issueSuggestions...)

		// Update overall risk
		for _, suggestion := range issueSuggestions {
			if suggestion.Risk > overallRisk {
				overallRisk = suggestion.Risk
			}
			estimatedTime += 5 * time.Minute // Estimated time per suggestion
		}
	}

	// Calculate automation level
	automationLevel := re.calculateAutomationLevel(suggestions)

	// Generate risk assessment
	riskAssessment := re.generateRiskAssessment(suggestions, overallRisk)

	// Generate rollback plan
	rollbackSteps := re.generateRollbackPlan(suggestions)

	plan := &RemediationPlan{
		ID:              fmt.Sprintf("plan_%d", time.Now().Unix()),
		Issues:          issues,
		Suggestions:     suggestions,
		AutomationLevel: automationLevel,
		EstimatedTime:   estimatedTime,
		RiskAssessment:  riskAssessment,
		Dependencies:    []string{},
		Rollback:        rollbackSteps,
		CreatedAt:       time.Now(),
		Priority:        re.calculateOverallPriority(issues),
	}

	re.logger.Info(fmt.Sprintf("Generated remediation plan with %d suggestions", len(suggestions)))
	return plan, nil
}

func (re *RemediationEngine) initializeDefaultRules() {
	// Initialize default remediation rules
	re.suggestionRules[FailureDiskSpace] = []RemediationRule{
		{
			ID:          "disk_cleanup_nix",
			FailureType: FailureDiskSpace,
			Conditions: []RuleCondition{
				{Field: "disk_usage", Operator: "gt", Value: 80.0, Weight: 1.0},
			},
			Actions: []RemediationAction{
				{Step: 1, Description: "Run Nix garbage collection", Commands: []string{"nix-collect-garbage -d"}, Automated: true},
				{Step: 2, Description: "Optimize Nix store", Commands: []string{"nix-store --optimize"}, Automated: true},
				{Step: 3, Description: "Clean journal logs", Commands: []string{"journalctl --vacuum-time=7d"}, Automated: true},
			},
			Priority:       PriorityHigh,
			AutoExecutable: true,
			RiskLevel:      RiskLow,
			Description:    "Clean up disk space using NixOS-specific tools",
		},
	}

	re.suggestionRules[FailureMemoryLeak] = []RemediationRule{
		{
			ID:          "memory_service_restart",
			FailureType: FailureMemoryLeak,
			Actions: []RemediationAction{
				{Step: 1, Description: "Restart affected service", Commands: []string{"systemctl restart ${service}"}, Automated: false},
				{Step: 2, Description: "Monitor memory usage", Commands: []string{"systemctl status ${service}"}, Automated: true},
			},
			Priority:       PriorityHigh,
			AutoExecutable: false,
			RiskLevel:      RiskMedium,
			Description:    "Restart service to resolve memory leak",
		},
	}
}

func (re *RemediationEngine) generateSuggestionsForIssue(issue HealthIssue) []RemediationSuggestion {
	var suggestions []RemediationSuggestion

	// Determine failure type from issue
	failureType := re.mapIssueToFailureType(issue)

	// Get rules for this failure type
	rules, exists := re.suggestionRules[failureType]
	if !exists {
		// Generate generic suggestion
		suggestions = append(suggestions, re.generateGenericSuggestion(issue))
		return suggestions
	}

	// Generate suggestions based on rules
	for _, rule := range rules {
		if re.ruleApplies(rule, issue) {
			suggestion := RemediationSuggestion{
				ID:          fmt.Sprintf("suggestion_%s_%d", rule.ID, time.Now().Unix()),
				Title:       rule.Description,
				Description: fmt.Sprintf("Apply %s to resolve %s", rule.Description, issue.Description),
				Type:        re.getActionType(rule),
				Actions:     rule.Actions,
				Priority:    rule.Priority,
				Risk:        rule.RiskLevel,
				Confidence:  0.8, // Default confidence
				Effort:      re.calculateEffort(rule.Actions),
				Benefits:    []string{"Resolve system issue", "Improve stability"},
				Prerequisites: []string{},
				Metadata: map[string]interface{}{
					"rule_id":    rule.ID,
					"issue_id":   issue.ID,
					"component":  issue.Component,
				},
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

func (re *RemediationEngine) mapIssueToFailureType(issue HealthIssue) FailureType {
	// Map issue types to failure types
	switch {
	case strings.Contains(strings.ToLower(issue.Component), "disk"):
		return FailureDiskSpace
	case strings.Contains(strings.ToLower(issue.Component), "memory"):
		return FailureMemoryLeak
	case strings.Contains(strings.ToLower(issue.Component), "cpu"):
		return FailurePerformance
	case strings.Contains(strings.ToLower(issue.Component), "network"):
		return FailureNetworkIssue
	default:
		return FailureServiceCrash
	}
}

func (re *RemediationEngine) ruleApplies(rule RemediationRule, issue HealthIssue) bool {
	// Check if rule conditions are met
	for _, condition := range rule.Conditions {
		if !re.evaluateCondition(condition, issue) {
			return false
		}
	}
	return true
}

func (re *RemediationEngine) evaluateCondition(condition RuleCondition, issue HealthIssue) bool {
	// Simplified condition evaluation
	return true // In real implementation, would evaluate actual conditions
}

func (re *RemediationEngine) generateGenericSuggestion(issue HealthIssue) RemediationSuggestion {
	return RemediationSuggestion{
		ID:          fmt.Sprintf("generic_%s", issue.ID),
		Title:       "Investigate Issue",
		Description: fmt.Sprintf("Manually investigate %s issue", issue.Component),
		Type:        "manual",
		Priority:    issue.Severity,
		Risk:        RiskLow,
		Confidence:  0.5,
		Effort:      "medium",
		Actions: []RemediationAction{
			{
				Step:        1,
				Description: "Check system logs",
				Commands:    []string{fmt.Sprintf("journalctl -u %s --since '1 hour ago'", issue.Component)},
				Automated:   true,
			},
		},
	}
}

func (re *RemediationEngine) getActionType(rule RemediationRule) string {
	if rule.AutoExecutable {
		return "automated"
	}
	return "manual"
}

func (re *RemediationEngine) calculateEffort(actions []RemediationAction) string {
	if len(actions) <= 1 {
		return "low"
	} else if len(actions) <= 3 {
		return "medium"
	}
	return "high"
}

func (re *RemediationEngine) calculateAutomationLevel(suggestions []RemediationSuggestion) AutomationLevel {
	automated := 0
	total := len(suggestions)

	for _, suggestion := range suggestions {
		if suggestion.Type == "automated" {
			automated++
		}
	}

	if total == 0 {
		return AutomationNone
	}

	ratio := float64(automated) / float64(total)
	if ratio >= 0.8 {
		return AutomationHigh
	} else if ratio >= 0.5 {
		return AutomationMedium
	} else if ratio > 0 {
		return AutomationLow
	}
	return AutomationNone
}

func (re *RemediationEngine) generateRiskAssessment(suggestions []RemediationSuggestion, overallRisk RiskLevel) RiskAssessment {
	return RiskAssessment{
		OverallRisk:        overallRisk,
		RiskFactors:        []RiskFactor{},
		MitigationPlan:     []string{"Review changes before execution", "Have rollback plan ready"},
		SuccessProbability: 0.85,
		RollbackRisk:       RiskLow,
		Metadata:           make(map[string]interface{}),
	}
}

func (re *RemediationEngine) generateRollbackPlan(suggestions []RemediationSuggestion) []RollbackStep {
	var steps []RollbackStep

	for i, suggestion := range suggestions {
		step := RollbackStep{
			Step:        i + 1,
			Description: fmt.Sprintf("Rollback changes from: %s", suggestion.Title),
			Commands:    []string{"# Rollback commands would be generated based on original actions"},
			Validation:  "Check system status",
			Timeout:     5 * time.Minute,
			Critical:    suggestion.Priority == PriorityCritical,
		}
		steps = append(steps, step)
	}

	return steps
}

func (re *RemediationEngine) calculateOverallPriority(issues []HealthIssue) Priority {
	maxPriority := PriorityLow

	for _, issue := range issues {
		if issue.Severity > maxPriority {
			maxPriority = issue.Severity
		}
	}

	return maxPriority
}