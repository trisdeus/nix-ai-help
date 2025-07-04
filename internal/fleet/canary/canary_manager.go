package canary

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"nix-ai-help/internal/fleet"
	"nix-ai-help/pkg/logger"
)

// CanaryManager handles canary deployments with intelligent rollback
type CanaryManager struct {
	logger       *logger.Logger
	fleetManager *fleet.FleetManager
	deployments  map[string]*CanaryDeployment
	mu           sync.RWMutex
}

// NewCanaryManager creates a new canary deployment manager
func NewCanaryManager(logger *logger.Logger, fleetManager *fleet.FleetManager) *CanaryManager {
	return &CanaryManager{
		logger:       logger,
		fleetManager: fleetManager,
		deployments:  make(map[string]*CanaryDeployment),
	}
}

// CanaryDeployment represents a canary deployment with intelligent monitoring
type CanaryDeployment struct {
	ID                string                    `json:"id"`
	Name              string                    `json:"name"`
	Status            CanaryStatus              `json:"status"`
	Config            CanaryConfig              `json:"config"`
	CanaryInstances   []string                  `json:"canary_instances"`
	ProductionInstances []string                `json:"production_instances"`
	StartTime         time.Time                 `json:"start_time"`
	EndTime           *time.Time                `json:"end_time,omitempty"`
	Metrics           *CanaryMetrics            `json:"metrics"`
	RollbackTriggers  []RollbackTrigger         `json:"rollback_triggers"`
	DecisionHistory   []CanaryDecision          `json:"decision_history"`
	AutoRollback      bool                      `json:"auto_rollback"`
	ProgressStages    []CanaryStage             `json:"progress_stages"`
	CurrentStage      int                       `json:"current_stage"`
}

// CanaryStatus represents the status of a canary deployment
type CanaryStatus string

const (
	CanaryStatusPending    CanaryStatus = "pending"
	CanaryStatusRunning    CanaryStatus = "running"
	CanaryStatusSuccessful CanaryStatus = "successful"
	CanaryStatusFailed     CanaryStatus = "failed"
	CanaryStatusRolledBack CanaryStatus = "rolled_back"
	CanaryStatusPaused     CanaryStatus = "paused"
)

// CanaryConfig defines the configuration for a canary deployment
type CanaryConfig struct {
	TrafficPercentage    float64           `json:"traffic_percentage"`    // Percentage of traffic to canary
	Duration             time.Duration     `json:"duration"`             // Total canary duration
	ProgressiveRollout   bool              `json:"progressive_rollout"`   // Gradually increase traffic
	HealthCheckInterval  time.Duration     `json:"health_check_interval"`
	MetricsCollection    MetricsConfig     `json:"metrics_collection"`
	RollbackTriggers     []TriggerConfig   `json:"rollback_triggers"`
	SuccessThresholds    SuccessThresholds `json:"success_thresholds"`
	NotificationConfig   NotificationConfig `json:"notification_config"`
}

// MetricsConfig defines what metrics to collect during canary
type MetricsConfig struct {
	ErrorRate       bool          `json:"error_rate"`
	ResponseTime    bool          `json:"response_time"`
	Throughput      bool          `json:"throughput"`
	CPUUsage        bool          `json:"cpu_usage"`
	MemoryUsage     bool          `json:"memory_usage"`
	CustomMetrics   []string      `json:"custom_metrics"`
	CollectionWindow time.Duration `json:"collection_window"`
}

// TriggerConfig defines conditions that trigger rollback
type TriggerConfig struct {
	Type      string  `json:"type"`       // error_rate, response_time, throughput, cpu, memory
	Threshold float64 `json:"threshold"`  // Threshold value
	Duration  time.Duration `json:"duration"` // How long threshold must be exceeded
	Operator  string  `json:"operator"`   // gt, lt, gte, lte
}

// SuccessThresholds define when canary is considered successful
type SuccessThresholds struct {
	MaxErrorRate     float64 `json:"max_error_rate"`      // Maximum acceptable error rate
	MaxResponseTime  float64 `json:"max_response_time"`   // Maximum acceptable response time (ms)
	MinThroughput    float64 `json:"min_throughput"`      // Minimum acceptable throughput
	MaxCPUUsage      float64 `json:"max_cpu_usage"`       // Maximum acceptable CPU usage
	MaxMemoryUsage   float64 `json:"max_memory_usage"`    // Maximum acceptable memory usage
	RequiredDuration time.Duration `json:"required_duration"` // How long thresholds must be met
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	Enabled   bool     `json:"enabled"`
	Channels  []string `json:"channels"`  // slack, email, webhook
	Webhooks  []string `json:"webhooks"`
	SlackChannel string `json:"slack_channel"`
	EmailList []string `json:"email_list"`
}

// CanaryMetrics holds the metrics collected during canary deployment
type CanaryMetrics struct {
	CanaryMetrics     MetricSnapshot `json:"canary_metrics"`
	ProductionMetrics MetricSnapshot `json:"production_metrics"`
	Comparison        MetricComparison `json:"comparison"`
	LastUpdated       time.Time      `json:"last_updated"`
}

// MetricSnapshot represents metrics at a point in time
type MetricSnapshot struct {
	ErrorRate     float64 `json:"error_rate"`      // Percentage
	ResponseTime  float64 `json:"response_time"`   // Milliseconds
	Throughput    float64 `json:"throughput"`      // Requests per second
	CPUUsage      float64 `json:"cpu_usage"`       // Percentage
	MemoryUsage   float64 `json:"memory_usage"`    // Percentage
	RequestCount  int64   `json:"request_count"`
	ErrorCount    int64   `json:"error_count"`
	Timestamp     time.Time `json:"timestamp"`
}

// MetricComparison compares canary vs production metrics
type MetricComparison struct {
	ErrorRateDiff     float64 `json:"error_rate_diff"`     // Positive = canary worse
	ResponseTimeDiff  float64 `json:"response_time_diff"`  // Positive = canary slower
	ThroughputDiff    float64 `json:"throughput_diff"`     // Positive = canary better
	CPUUsageDiff      float64 `json:"cpu_usage_diff"`      // Positive = canary uses more
	MemoryUsageDiff   float64 `json:"memory_usage_diff"`   // Positive = canary uses more
	StatisticalSignificance bool `json:"statistical_significance"`
	ConfidenceLevel     float64 `json:"confidence_level"`
}

// RollbackTrigger represents a condition that triggered rollback
type RollbackTrigger struct {
	Type        string    `json:"type"`
	Threshold   float64   `json:"threshold"`
	ActualValue float64   `json:"actual_value"`
	Duration    time.Duration `json:"duration"`
	TriggeredAt time.Time `json:"triggered_at"`
	Reason      string    `json:"reason"`
}

// CanaryDecision represents a decision made during canary deployment
type CanaryDecision struct {
	Timestamp   time.Time `json:"timestamp"`
	Decision    string    `json:"decision"`    // continue, pause, rollback, promote
	Reason      string    `json:"reason"`
	Confidence  float64   `json:"confidence"`  // 0-1 confidence in decision
	Metrics     MetricSnapshot `json:"metrics"`
	AutoDecision bool      `json:"auto_decision"` // Was this an automated decision?
}

// CanaryStage represents a stage in progressive rollout
type CanaryStage struct {
	Stage           int       `json:"stage"`
	TrafficPercent  float64   `json:"traffic_percent"`
	Duration        time.Duration `json:"duration"`
	StartTime       *time.Time `json:"start_time,omitempty"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	Status          string    `json:"status"` // pending, running, completed, failed
	Metrics         *MetricSnapshot `json:"metrics,omitempty"`
	HealthCheck     bool      `json:"health_check"`
}

// CreateCanaryDeployment creates a new canary deployment
func (cm *CanaryManager) CreateCanaryDeployment(ctx context.Context, deployment *CanaryDeployment) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate deployment configuration
	if err := cm.validateCanaryConfig(deployment); err != nil {
		return fmt.Errorf("invalid canary configuration: %w", err)
	}

	// Check that canary and production instances exist
	if err := cm.validateInstances(ctx, deployment); err != nil {
		return fmt.Errorf("invalid instances: %w", err)
	}

	// Initialize deployment
	deployment.Status = CanaryStatusPending
	deployment.StartTime = time.Now()
	deployment.DecisionHistory = []CanaryDecision{}
	deployment.RollbackTriggers = []RollbackTrigger{}

	// Create progressive rollout stages if enabled
	if deployment.Config.ProgressiveRollout {
		deployment.ProgressStages = cm.createProgressiveStages(deployment.Config)
		deployment.CurrentStage = 0
	}

	// Initialize metrics collection
	deployment.Metrics = &CanaryMetrics{
		LastUpdated: time.Now(),
	}

	cm.deployments[deployment.ID] = deployment
	cm.logger.Info(fmt.Sprintf("Created canary deployment: %s", deployment.ID))

	return nil
}

// StartCanaryDeployment starts a canary deployment
func (cm *CanaryManager) StartCanaryDeployment(ctx context.Context, deploymentID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	deployment, exists := cm.deployments[deploymentID]
	if !exists {
		return fmt.Errorf("canary deployment %s not found", deploymentID)
	}

	if deployment.Status != CanaryStatusPending {
		return fmt.Errorf("canary deployment %s is not in pending status", deploymentID)
	}

	deployment.Status = CanaryStatusRunning
	deployment.StartTime = time.Now()

	cm.logger.Info(fmt.Sprintf("Started canary deployment: %s", deploymentID))

	// Start metrics collection and monitoring
	go cm.monitorCanaryDeployment(ctx, deployment)

	return nil
}

// monitorCanaryDeployment monitors a canary deployment and makes decisions
func (cm *CanaryManager) monitorCanaryDeployment(ctx context.Context, deployment *CanaryDeployment) {
	ticker := time.NewTicker(deployment.Config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if deployment.Status != CanaryStatusRunning {
				return
			}

			// Collect metrics
			if err := cm.collectMetrics(ctx, deployment); err != nil {
				cm.logger.Error(fmt.Sprintf("Error collecting metrics for canary %s: %v", deployment.ID, err))
				continue
			}

			// Evaluate rollback triggers
			if triggered, trigger := cm.evaluateRollbackTriggers(deployment); triggered {
				cm.logger.Warn(fmt.Sprintf("Rollback trigger activated for canary %s: %s", deployment.ID, trigger.Reason))
				deployment.RollbackTriggers = append(deployment.RollbackTriggers, trigger)
				
				if deployment.AutoRollback {
					if err := cm.rollbackCanary(ctx, deployment); err != nil {
						cm.logger.Error(fmt.Sprintf("Auto-rollback failed for canary %s: %v", deployment.ID, err))
					}
					return
				}
			}

			// Make deployment decision
			decision := cm.makeCanaryDecision(deployment)
			deployment.DecisionHistory = append(deployment.DecisionHistory, decision)

			switch decision.Decision {
			case "continue":
				// Continue monitoring
				continue
			case "promote":
				if err := cm.promoteCanary(ctx, deployment); err != nil {
					cm.logger.Error(fmt.Sprintf("Canary promotion failed for %s: %v", deployment.ID, err))
				}
				return
			case "rollback":
				if err := cm.rollbackCanary(ctx, deployment); err != nil {
					cm.logger.Error(fmt.Sprintf("Canary rollback failed for %s: %v", deployment.ID, err))
				}
				return
			case "pause":
				deployment.Status = CanaryStatusPaused
				cm.logger.Info(fmt.Sprintf("Canary deployment paused: %s", deployment.ID))
				return
			}
		}
	}
}

// collectMetrics collects metrics from canary and production instances
func (cm *CanaryManager) collectMetrics(ctx context.Context, deployment *CanaryDeployment) error {
	// Collect canary metrics
	canaryMetrics, err := cm.collectInstanceMetrics(ctx, deployment.CanaryInstances)
	if err != nil {
		return fmt.Errorf("error collecting canary metrics: %w", err)
	}

	// Collect production metrics
	productionMetrics, err := cm.collectInstanceMetrics(ctx, deployment.ProductionInstances)
	if err != nil {
		return fmt.Errorf("error collecting production metrics: %w", err)
	}

	// Update deployment metrics
	deployment.Metrics.CanaryMetrics = canaryMetrics
	deployment.Metrics.ProductionMetrics = productionMetrics
	deployment.Metrics.Comparison = cm.compareMetrics(canaryMetrics, productionMetrics)
	deployment.Metrics.LastUpdated = time.Now()

	return nil
}

// collectInstanceMetrics collects metrics from a set of instances
func (cm *CanaryManager) collectInstanceMetrics(ctx context.Context, instances []string) (MetricSnapshot, error) {
	var totalErrorRate, totalResponseTime, totalThroughput, totalCPU, totalMemory float64
	var totalRequests, totalErrors int64
	validInstances := 0

	for _, instanceID := range instances {
		machine, err := cm.fleetManager.GetMachine(ctx, instanceID)
		if err != nil {
			cm.logger.Warn(fmt.Sprintf("Cannot get machine %s: %v", instanceID, err))
			continue
		}

		// Collect metrics from machine (this would integrate with actual monitoring)
		// For now, we'll simulate realistic metrics
		metrics := cm.simulateInstanceMetrics(machine)
		
		totalErrorRate += metrics.ErrorRate
		totalResponseTime += metrics.ResponseTime
		totalThroughput += metrics.Throughput
		totalCPU += metrics.CPUUsage
		totalMemory += metrics.MemoryUsage
		totalRequests += metrics.RequestCount
		totalErrors += metrics.ErrorCount
		validInstances++
	}

	if validInstances == 0 {
		return MetricSnapshot{}, fmt.Errorf("no valid instances found")
	}

	return MetricSnapshot{
		ErrorRate:     totalErrorRate / float64(validInstances),
		ResponseTime:  totalResponseTime / float64(validInstances),
		Throughput:    totalThroughput / float64(validInstances),
		CPUUsage:      totalCPU / float64(validInstances),
		MemoryUsage:   totalMemory / float64(validInstances),
		RequestCount:  totalRequests,
		ErrorCount:    totalErrors,
		Timestamp:     time.Now(),
	}, nil
}

// simulateInstanceMetrics simulates realistic metrics for demonstration
func (cm *CanaryManager) simulateInstanceMetrics(machine *fleet.Machine) MetricSnapshot {
	// In real implementation, this would collect actual metrics
	// For demonstration, we'll create realistic simulated metrics
	baseErrorRate := 0.5 // 0.5% base error rate
	baseResponseTime := 150.0 // 150ms base response time
	baseThroughput := 100.0 // 100 RPS base throughput
	
	// Add some variation based on machine health
	healthMultiplier := 1.0
	if machine.Health.Overall == "warning" {
		healthMultiplier = 1.2
	} else if machine.Health.Overall == "critical" {
		healthMultiplier = 1.5
	}

	return MetricSnapshot{
		ErrorRate:     baseErrorRate * healthMultiplier,
		ResponseTime:  baseResponseTime * healthMultiplier,
		Throughput:    baseThroughput / healthMultiplier,
		CPUUsage:      machine.Health.CPU.Usage,
		MemoryUsage:   machine.Health.Memory.Usage,
		RequestCount:  int64(baseThroughput * 60), // 1 minute worth
		ErrorCount:    int64(baseThroughput * 60 * (baseErrorRate * healthMultiplier / 100)),
		Timestamp:     time.Now(),
	}
}

// compareMetrics compares canary and production metrics
func (cm *CanaryManager) compareMetrics(canary, production MetricSnapshot) MetricComparison {
	comparison := MetricComparison{
		ErrorRateDiff:    canary.ErrorRate - production.ErrorRate,
		ResponseTimeDiff: canary.ResponseTime - production.ResponseTime,
		ThroughputDiff:   canary.Throughput - production.Throughput,
		CPUUsageDiff:     canary.CPUUsage - production.CPUUsage,
		MemoryUsageDiff:  canary.MemoryUsage - production.MemoryUsage,
	}

	// Calculate statistical significance (simplified)
	totalCanaryRequests := canary.RequestCount
	totalProductionRequests := production.RequestCount
	
	if totalCanaryRequests > 100 && totalProductionRequests > 100 {
		// Simplified statistical significance check
		errorRateDiffSignificance := math.Abs(comparison.ErrorRateDiff)
		responseTimeDiffSignificance := math.Abs(comparison.ResponseTimeDiff)
		
		if errorRateDiffSignificance > 0.5 || responseTimeDiffSignificance > 20 {
			comparison.StatisticalSignificance = true
			comparison.ConfidenceLevel = 0.95
		}
	}

	return comparison
}

// evaluateRollbackTriggers checks if any rollback triggers are activated
func (cm *CanaryManager) evaluateRollbackTriggers(deployment *CanaryDeployment) (bool, RollbackTrigger) {
	metrics := deployment.Metrics.CanaryMetrics
	comparison := deployment.Metrics.Comparison

	for _, trigger := range deployment.Config.RollbackTriggers {
		var currentValue float64
		var triggered bool

		switch trigger.Type {
		case "error_rate":
			currentValue = metrics.ErrorRate
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		case "response_time":
			currentValue = metrics.ResponseTime
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		case "throughput":
			currentValue = metrics.Throughput
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		case "cpu":
			currentValue = metrics.CPUUsage
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		case "memory":
			currentValue = metrics.MemoryUsage
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		case "error_rate_diff":
			currentValue = comparison.ErrorRateDiff
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		case "response_time_diff":
			currentValue = comparison.ResponseTimeDiff
			triggered = cm.evaluateThreshold(currentValue, trigger.Threshold, trigger.Operator)
		}

		if triggered {
			return true, RollbackTrigger{
				Type:        trigger.Type,
				Threshold:   trigger.Threshold,
				ActualValue: currentValue,
				Duration:    trigger.Duration,
				TriggeredAt: time.Now(),
				Reason:      fmt.Sprintf("%s exceeded threshold: %.2f > %.2f", trigger.Type, currentValue, trigger.Threshold),
			}
		}
	}

	return false, RollbackTrigger{}
}

// evaluateThreshold evaluates if a value meets a threshold condition
func (cm *CanaryManager) evaluateThreshold(value, threshold float64, operator string) bool {
	switch operator {
	case "gt":
		return value > threshold
	case "lt":
		return value < threshold
	case "gte":
		return value >= threshold
	case "lte":
		return value <= threshold
	default:
		return false
	}
}

// makeCanaryDecision makes an intelligent decision about canary deployment
func (cm *CanaryManager) makeCanaryDecision(deployment *CanaryDeployment) CanaryDecision {
	metrics := deployment.Metrics.CanaryMetrics
	comparison := deployment.Metrics.Comparison
	thresholds := deployment.Config.SuccessThresholds

	// Check if canary has been running long enough
	runningTime := time.Since(deployment.StartTime)
	if runningTime < thresholds.RequiredDuration {
		return CanaryDecision{
			Timestamp:    time.Now(),
			Decision:     "continue",
			Reason:       fmt.Sprintf("Canary needs to run for %v more", thresholds.RequiredDuration-runningTime),
			Confidence:   0.8,
			Metrics:      metrics,
			AutoDecision: true,
		}
	}

	// Calculate confidence score based on metrics
	confidence := cm.calculateDecisionConfidence(metrics, comparison, thresholds)

	// Check success criteria
	if cm.meetsSuccessThresholds(metrics, thresholds) && comparison.StatisticalSignificance {
		if comparison.ErrorRateDiff <= 0.1 && comparison.ResponseTimeDiff <= 10 {
			return CanaryDecision{
				Timestamp:    time.Now(),
				Decision:     "promote",
				Reason:       "Canary meets all success criteria and shows no performance degradation",
				Confidence:   confidence,
				Metrics:      metrics,
				AutoDecision: true,
			}
		}
	}

	// Check failure criteria
	if metrics.ErrorRate > thresholds.MaxErrorRate || 
	   metrics.ResponseTime > thresholds.MaxResponseTime ||
	   comparison.ErrorRateDiff > 1.0 || comparison.ResponseTimeDiff > 50 {
		return CanaryDecision{
			Timestamp:    time.Now(),
			Decision:     "rollback",
			Reason:       "Canary performance is significantly worse than production",
			Confidence:   confidence,
			Metrics:      metrics,
			AutoDecision: true,
		}
	}

	// Continue monitoring
	return CanaryDecision{
		Timestamp:    time.Now(),
		Decision:     "continue",
		Reason:       "Canary is performing within acceptable parameters",
		Confidence:   confidence,
		Metrics:      metrics,
		AutoDecision: true,
	}
}

// calculateDecisionConfidence calculates confidence in the decision
func (cm *CanaryManager) calculateDecisionConfidence(metrics MetricSnapshot, comparison MetricComparison, thresholds SuccessThresholds) float64 {
	var confidence float64 = 0.5

	// Increase confidence with more data
	if metrics.RequestCount > 1000 {
		confidence += 0.2
	}
	if metrics.RequestCount > 10000 {
		confidence += 0.1
	}

	// Increase confidence with statistical significance
	if comparison.StatisticalSignificance {
		confidence += 0.2
	}

	// Decrease confidence if metrics are borderline
	if math.Abs(metrics.ErrorRate-thresholds.MaxErrorRate) < 0.1 {
		confidence -= 0.1
	}
	if math.Abs(metrics.ResponseTime-thresholds.MaxResponseTime) < 10 {
		confidence -= 0.1
	}

	// Ensure confidence is between 0 and 1
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// meetsSuccessThresholds checks if metrics meet success criteria
func (cm *CanaryManager) meetsSuccessThresholds(metrics MetricSnapshot, thresholds SuccessThresholds) bool {
	return metrics.ErrorRate <= thresholds.MaxErrorRate &&
		   metrics.ResponseTime <= thresholds.MaxResponseTime &&
		   metrics.Throughput >= thresholds.MinThroughput &&
		   metrics.CPUUsage <= thresholds.MaxCPUUsage &&
		   metrics.MemoryUsage <= thresholds.MaxMemoryUsage
}

// promoteCanary promotes a canary deployment to production
func (cm *CanaryManager) promoteCanary(ctx context.Context, deployment *CanaryDeployment) error {
	cm.logger.Info(fmt.Sprintf("Promoting canary deployment: %s", deployment.ID))

	// In real implementation, this would:
	// 1. Route 100% traffic to canary instances
	// 2. Update load balancer configuration
	// 3. Update DNS records if needed
	// 4. Terminate old production instances
	// 5. Promote canary instances to production

	deployment.Status = CanaryStatusSuccessful
	now := time.Now()
	deployment.EndTime = &now

	cm.logger.Info(fmt.Sprintf("Successfully promoted canary deployment: %s", deployment.ID))
	return nil
}

// rollbackCanary rolls back a canary deployment
func (cm *CanaryManager) rollbackCanary(ctx context.Context, deployment *CanaryDeployment) error {
	cm.logger.Info(fmt.Sprintf("Rolling back canary deployment: %s", deployment.ID))

	// In real implementation, this would:
	// 1. Route 100% traffic back to production instances
	// 2. Terminate canary instances
	// 3. Update load balancer configuration
	// 4. Clean up any temporary resources

	deployment.Status = CanaryStatusRolledBack
	now := time.Now()
	deployment.EndTime = &now

	cm.logger.Info(fmt.Sprintf("Successfully rolled back canary deployment: %s", deployment.ID))
	return nil
}

// validateCanaryConfig validates canary deployment configuration
func (cm *CanaryManager) validateCanaryConfig(deployment *CanaryDeployment) error {
	config := deployment.Config

	if config.TrafficPercentage <= 0 || config.TrafficPercentage > 100 {
		return fmt.Errorf("traffic percentage must be between 0 and 100")
	}

	if config.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	if config.HealthCheckInterval <= 0 {
		return fmt.Errorf("health check interval must be positive")
	}

	if len(deployment.CanaryInstances) == 0 {
		return fmt.Errorf("at least one canary instance is required")
	}

	if len(deployment.ProductionInstances) == 0 {
		return fmt.Errorf("at least one production instance is required")
	}

	return nil
}

// validateInstances validates that all instances exist in the fleet
func (cm *CanaryManager) validateInstances(ctx context.Context, deployment *CanaryDeployment) error {
	allInstances := append(deployment.CanaryInstances, deployment.ProductionInstances...)
	
	for _, instanceID := range allInstances {
		_, err := cm.fleetManager.GetMachine(ctx, instanceID)
		if err != nil {
			return fmt.Errorf("instance %s not found in fleet: %w", instanceID, err)
		}
	}

	return nil
}

// createProgressiveStages creates stages for progressive rollout
func (cm *CanaryManager) createProgressiveStages(config CanaryConfig) []CanaryStage {
	stages := []CanaryStage{}
	
	// Create progressive stages: 5%, 10%, 25%, 50%, 100%
	percentages := []float64{5, 10, 25, 50, 100}
	stageDuration := config.Duration / time.Duration(len(percentages))

	for i, percentage := range percentages {
		if percentage > config.TrafficPercentage {
			percentage = config.TrafficPercentage
		}

		stages = append(stages, CanaryStage{
			Stage:          i + 1,
			TrafficPercent: percentage,
			Duration:       stageDuration,
			Status:         "pending",
			HealthCheck:    true,
		})

		if percentage >= config.TrafficPercentage {
			break
		}
	}

	return stages
}

// GetCanaryDeployment retrieves a canary deployment by ID
func (cm *CanaryManager) GetCanaryDeployment(ctx context.Context, deploymentID string) (*CanaryDeployment, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	deployment, exists := cm.deployments[deploymentID]
	if !exists {
		return nil, fmt.Errorf("canary deployment %s not found", deploymentID)
	}

	return deployment, nil
}

// ListCanaryDeployments lists all canary deployments
func (cm *CanaryManager) ListCanaryDeployments(ctx context.Context) ([]*CanaryDeployment, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	deployments := make([]*CanaryDeployment, 0, len(cm.deployments))
	for _, deployment := range cm.deployments {
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// PromoteCanary promotes a canary deployment to production (public wrapper)
func (cm *CanaryManager) PromoteCanary(ctx context.Context, deployment *CanaryDeployment) error {
	return cm.promoteCanary(ctx, deployment)
}

// RollbackCanary rolls back a canary deployment (public wrapper)
func (cm *CanaryManager) RollbackCanary(ctx context.Context, deployment *CanaryDeployment) error {
	return cm.rollbackCanary(ctx, deployment)
}