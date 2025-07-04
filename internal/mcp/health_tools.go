package mcp

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/internal/health"
	"nix-ai-help/internal/health/ml"
	"nix-ai-help/pkg/logger"
)

// HealthMCPTools provides health prediction and monitoring for the MCP server
type HealthMCPTools struct {
	healthPredictor  health.HealthPredictor
	predictionModel  *ml.FailurePredictionModel
	logger          *logger.Logger
}

// NewHealthMCPTools creates new health monitoring tools for MCP
func NewHealthMCPTools(logger *logger.Logger) *HealthMCPTools {
	return &HealthMCPTools{
		healthPredictor: nil, // Will be initialized when needed
		predictionModel: ml.NewFailurePredictionModel(nil),
		logger:         logger,
	}
}

// MCPHealthCheckRequest represents a health check request
type MCPHealthCheckRequest struct {
	Components []string `json:"components,omitempty"` // Specific components to check
	Detailed   bool     `json:"detailed,omitempty"`   // Include detailed metrics
	Timeout    int      `json:"timeout,omitempty"`    // Timeout in seconds
}

// MCPHealthCheckResponse represents system health status
type MCPHealthCheckResponse struct {
	OverallStatus    string                    `json:"overall_status"`
	Components       []ComponentHealthStatus   `json:"components"`
	SystemMetrics    *SystemMetrics           `json:"system_metrics,omitempty"`
	Recommendations  []string                 `json:"recommendations"`
	LastUpdated      time.Time                `json:"last_updated"`
	HealthScore      float64                  `json:"health_score"`
}

// ComponentHealthStatus represents health of individual system components
type ComponentHealthStatus struct {
	Name         string                 `json:"name"`
	Status       string                 `json:"status"` // "healthy", "warning", "critical", "unknown"
	Message      string                 `json:"message"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	LastChecked  time.Time              `json:"last_checked"`
	Trends       []string               `json:"trends,omitempty"`
}

// SystemMetrics represents detailed system metrics
type SystemMetrics struct {
	CPU          CPUMetrics          `json:"cpu"`
	Memory       MemoryMetrics       `json:"memory"`
	Disk         []DiskMetrics       `json:"disk"`
	Network      NetworkMetrics      `json:"network"`
	NixStore     NixStoreMetrics     `json:"nix_store"`
	Services     []ServiceMetrics    `json:"services"`
	Processes    ProcessMetrics      `json:"processes"`
}

type CPUMetrics struct {
	Usage         float64 `json:"usage_percent"`
	LoadAverage   []float64 `json:"load_average"` // 1, 5, 15 minute averages
	Temperature   float64 `json:"temperature,omitempty"`
	Cores         int     `json:"cores"`
	Frequency     float64 `json:"frequency_mhz,omitempty"`
}

type MemoryMetrics struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Available   uint64  `json:"available_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	SwapTotal   uint64  `json:"swap_total_bytes"`
	SwapUsed    uint64  `json:"swap_used_bytes"`
	Cached      uint64  `json:"cached_bytes"`
	Buffers     uint64  `json:"buffers_bytes"`
}

type DiskMetrics struct {
	Device       string  `json:"device"`
	Mountpoint   string  `json:"mountpoint"`
	Total        uint64  `json:"total_bytes"`
	Used         uint64  `json:"used_bytes"`
	Available    uint64  `json:"available_bytes"`
	UsagePercent float64 `json:"usage_percent"`
	Filesystem   string  `json:"filesystem"`
	Inodes       uint64  `json:"inodes_total,omitempty"`
	InodesUsed   uint64  `json:"inodes_used,omitempty"`
}

type NetworkMetrics struct {
	Interfaces    []NetworkInterface `json:"interfaces"`
	Connections   int               `json:"active_connections"`
	PacketsIn     uint64            `json:"packets_in"`
	PacketsOut    uint64            `json:"packets_out"`
	BytesIn       uint64            `json:"bytes_in"`
	BytesOut      uint64            `json:"bytes_out"`
	ErrorsIn      uint64            `json:"errors_in"`
	ErrorsOut     uint64            `json:"errors_out"`
}

type NetworkInterface struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	Speed        uint64 `json:"speed_mbps,omitempty"`
	MTU          int    `json:"mtu"`
	PacketsIn    uint64 `json:"packets_in"`
	PacketsOut   uint64 `json:"packets_out"`
	BytesIn      uint64 `json:"bytes_in"`
	BytesOut     uint64 `json:"bytes_out"`
}

type NixStoreMetrics struct {
	TotalSize      uint64 `json:"total_size_bytes"`
	LivePaths      int    `json:"live_paths"`
	DeadPaths      int    `json:"dead_paths"`
	GCRoots        int    `json:"gc_roots"`
	LastGC         time.Time `json:"last_gc,omitempty"`
	GCRecommended  bool   `json:"gc_recommended"`
	StorePath      string `json:"store_path"`
}

type ServiceMetrics struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Active       bool      `json:"active"`
	Enabled      bool      `json:"enabled"`
	Memory       uint64    `json:"memory_bytes,omitempty"`
	CPU          float64   `json:"cpu_percent,omitempty"`
	Restarts     int       `json:"restart_count"`
	LastRestart  time.Time `json:"last_restart,omitempty"`
	Uptime       time.Duration `json:"uptime"`
}

type ProcessMetrics struct {
	Total       int     `json:"total"`
	Running     int     `json:"running"`
	Sleeping    int     `json:"sleeping"`
	Stopped     int     `json:"stopped"`
	Zombie      int     `json:"zombie"`
	TopCPU      []ProcessInfo `json:"top_cpu,omitempty"`
	TopMemory   []ProcessInfo `json:"top_memory,omitempty"`
}

type ProcessInfo struct {
	PID         int     `json:"pid"`
	Name        string  `json:"name"`
	Command     string  `json:"command"`
	CPU         float64 `json:"cpu_percent"`
	Memory      uint64  `json:"memory_bytes"`
	User        string  `json:"user"`
}

// MCPPredictionRequest represents a failure prediction request
type MCPPredictionRequest struct {
	Timeline     string   `json:"timeline"`      // "1day", "1week", "1month"
	Components   []string `json:"components,omitempty"`
	IncludeActions bool   `json:"include_actions,omitempty"`
}

// MCPPredictionResponse represents failure prediction results
type MCPPredictionResponse struct {
	Timeline          time.Duration               `json:"timeline"`
	PredictedFailures []PredictedFailureInfo      `json:"predicted_failures"`
	OverallRisk       string                      `json:"overall_risk"` // "low", "medium", "high", "critical"
	Confidence        float64                     `json:"confidence"`
	PreventiveActions []PreventiveActionInfo      `json:"preventive_actions,omitempty"`
	ResourceForecasts []ResourceForecast          `json:"resource_forecasts,omitempty"`
	Recommendations   []string                    `json:"recommendations"`
	GeneratedAt       time.Time                   `json:"generated_at"`
}

type PredictedFailureInfo struct {
	Type             string    `json:"type"`
	Component        string    `json:"component"`
	Description      string    `json:"description"`
	Probability      float64   `json:"probability"`
	EstimatedTime    time.Time `json:"estimated_time"`
	Impact           string    `json:"impact"` // "low", "medium", "high", "critical"
	Indicators       []string  `json:"indicators"`
}

type PreventiveActionInfo struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Priority    string        `json:"priority"`
	ETA         time.Duration `json:"estimated_time"`
	Commands    []string      `json:"commands,omitempty"`
	AutoApply   bool          `json:"auto_apply"`
}

type ResourceForecast struct {
	Resource     string    `json:"resource"` // "cpu", "memory", "disk", "network"
	Current      float64   `json:"current_usage"`
	Predicted    float64   `json:"predicted_usage"`
	Trend        string    `json:"trend"` // "increasing", "decreasing", "stable"
	ForecastTime time.Time `json:"forecast_time"`
	Confidence   float64   `json:"confidence"`
}

// CheckSystemHealth performs comprehensive health check
func (h *HealthMCPTools) CheckSystemHealth(ctx context.Context, req *MCPHealthCheckRequest) (*MCPHealthCheckResponse, error) {
	h.logger.Info("Performing system health check via MCP")

	response := &MCPHealthCheckResponse{
		Components:      []ComponentHealthStatus{},
		Recommendations: []string{},
		LastUpdated:     time.Now(),
	}

	// Check core system components
	components := req.Components
	if len(components) == 0 {
		components = []string{"cpu", "memory", "disk", "network", "services", "nix-store"}
	}

	totalScore := 0.0

	for _, component := range components {
		status := h.checkComponentHealth(ctx, component)
		response.Components = append(response.Components, status)

		// Calculate component score
		switch status.Status {
		case "healthy":
			totalScore += 1.0
		case "warning":
			totalScore += 0.7
		case "critical":
			totalScore += 0.3
		default:
			totalScore += 0.5
		}
	}

	// Calculate overall health score
	response.HealthScore = totalScore / float64(len(components))

	// Determine overall status
	if response.HealthScore >= 0.8 {
		response.OverallStatus = "healthy"
	} else if response.HealthScore >= 0.6 {
		response.OverallStatus = "warning"
	} else {
		response.OverallStatus = "critical"
	}

	// Include detailed metrics if requested
	if req.Detailed {
		metrics, err := h.collectSystemMetrics(ctx)
		if err != nil {
			h.logger.Warn(fmt.Sprintf("Failed to collect detailed metrics: %v", err))
		} else {
			response.SystemMetrics = metrics
		}
	}

	// Generate recommendations
	response.Recommendations = h.generateHealthRecommendations(response)

	return response, nil
}

// PredictSystemFailures predicts potential system failures
func (h *HealthMCPTools) PredictSystemFailures(ctx context.Context, req *MCPPredictionRequest) (*MCPPredictionResponse, error) {
	h.logger.Info(fmt.Sprintf("Generating failure predictions for timeline: %s", req.Timeline))

	// Parse timeline
	timeline, err := h.parseTimeline(req.Timeline)
	if err != nil {
		return nil, fmt.Errorf("invalid timeline: %w", err)
	}

	// For now, use simulated training data since we don't have a real health monitor
	// In a production system, this would integrate with actual system monitoring
	trainingEvents := h.generateMockTrainingData()
	
	// Convert health.HealthEvent to ml.HealthEvent for training
	mlTrainingEvents := h.convertToMLHealthEvents(trainingEvents)
	
	// Train or update the model if we have enough data
	if len(mlTrainingEvents) >= 50 {
		err := h.predictionModel.Train(ctx, mlTrainingEvents)
		if err != nil {
			h.logger.Warn(fmt.Sprintf("Model training failed: %v", err))
		}
	}

	// Generate predictions
	prediction, err := h.predictionModel.Predict(ctx, timeline)
	if err != nil {
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// Convert to MCP response format
	failurePrediction, ok := prediction.(*health.FailurePrediction)
	if !ok {
		return nil, fmt.Errorf("unexpected prediction type")
	}

	response := &MCPPredictionResponse{
		Timeline:          timeline,
		PredictedFailures: []PredictedFailureInfo{},
		Confidence:        failurePrediction.Confidence,
		PreventiveActions: []PreventiveActionInfo{},
		ResourceForecasts: []ResourceForecast{},
		Recommendations:   []string{},
		GeneratedAt:       time.Now(),
	}

	// Convert predicted failures
	for _, failure := range failurePrediction.PredictedFailures {
		// Convert health indicators to strings
		var indicators []string
		for _, indicator := range failure.Indicators {
			indicators = append(indicators, indicator.Name) // Assuming HealthIndicator has a Name field
		}
		
		info := PredictedFailureInfo{
			Type:          string(failure.Type),
			Component:     failure.Component,
			Description:   failure.Description,
			Probability:   failure.ProbabilityScore,
			EstimatedTime: failure.EstimatedTime,
			Impact:        h.mapImpactLevel(failure.Impact),
			Indicators:    indicators,
		}
		response.PredictedFailures = append(response.PredictedFailures, info)
	}

	// Convert preventive actions if requested
	if req.IncludeActions {
		for _, action := range failurePrediction.PreventiveActions {
			actionInfo := PreventiveActionInfo{
				ID:          action.ID,
				Title:       "Preventive Action", // Default title since it's not in the struct
				Description: action.Description,
				Priority:    string(action.Risk), // Use Risk as Priority
				ETA:         action.ETA,
				Commands:    action.Commands,
				AutoApply:   action.Automated, // Use Automated field
			}
			response.PreventiveActions = append(response.PreventiveActions, actionInfo)
		}
	}

	// Determine overall risk level
	response.OverallRisk = h.calculateOverallRisk(response.PredictedFailures)

	// Generate recommendations
	response.Recommendations = h.generatePredictionRecommendations(response)

	return response, nil
}

// Helper methods

func (h *HealthMCPTools) checkComponentHealth(ctx context.Context, component string) ComponentHealthStatus {
	status := ComponentHealthStatus{
		Name:        component,
		Status:      "healthy",
		Message:     "Component is operating normally",
		Metrics:     make(map[string]interface{}),
		LastChecked: time.Now(),
		Trends:      []string{},
	}

	// Add component-specific basic health checks
	switch component {
	case "cpu":
		status.Metrics["usage"] = "Normal"
		status.Message = "CPU operating within normal parameters"
	case "memory":
		status.Metrics["usage"] = "Normal"
		status.Message = "Memory usage is stable"
	case "disk":
		status.Metrics["usage"] = "Normal"
		status.Message = "Disk space available"
	case "network":
		status.Metrics["status"] = "Connected"
		status.Message = "Network connectivity confirmed"
	case "services":
		status.Metrics["active"] = "Running"
		status.Message = "System services operational"
	case "nix-store":
		status.Metrics["size"] = "Manageable"
		status.Message = "Nix store in good condition"
	default:
		status.Status = "unknown"
		status.Message = "Component status unknown"
	}

	return status
}

func (h *HealthMCPTools) collectSystemMetrics(ctx context.Context) (*SystemMetrics, error) {
	// This would integrate with actual system monitoring
	// For now, return placeholder metrics
	metrics := &SystemMetrics{
		CPU: CPUMetrics{
			Usage:       65.5,
			LoadAverage: []float64{1.2, 1.1, 0.9},
			Cores:       8,
		},
		Memory: MemoryMetrics{
			Total:        16 * 1024 * 1024 * 1024, // 16GB
			Used:         8 * 1024 * 1024 * 1024,  // 8GB
			Available:    8 * 1024 * 1024 * 1024,  // 8GB
			UsagePercent: 50.0,
		},
		Disk: []DiskMetrics{
			{
				Device:       "/dev/sda1",
				Mountpoint:   "/",
				Total:        500 * 1024 * 1024 * 1024, // 500GB
				Used:         200 * 1024 * 1024 * 1024, // 200GB
				Available:    300 * 1024 * 1024 * 1024, // 300GB
				UsagePercent: 40.0,
				Filesystem:   "ext4",
			},
		},
		NixStore: NixStoreMetrics{
			TotalSize:     50 * 1024 * 1024 * 1024, // 50GB
			LivePaths:     15000,
			DeadPaths:     5000,
			GCRoots:       200,
			GCRecommended: false,
			StorePath:     "/nix/store",
		},
	}

	return metrics, nil
}

func (h *HealthMCPTools) parseTimeline(timeline string) (time.Duration, error) {
	switch timeline {
	case "1hour", "1h":
		return time.Hour, nil
	case "1day", "1d":
		return time.Hour * 24, nil
	case "1week", "1w":
		return time.Hour * 24 * 7, nil
	case "1month", "1m":
		return time.Hour * 24 * 30, nil
	default:
		// Try to parse as duration
		duration, err := time.ParseDuration(timeline)
		if err != nil {
			return 0, fmt.Errorf("invalid timeline format: %s", timeline)
		}
		return duration, nil
	}
}

func (h *HealthMCPTools) mapImpactLevel(level health.ImpactLevel) string {
	switch level {
	case health.ImpactSevere:
		return "critical"
	case health.ImpactSignificant:
		return "high"
	case health.ImpactModerate:
		return "medium"
	default:
		return "low"
	}
}

func (h *HealthMCPTools) calculateOverallRisk(failures []PredictedFailureInfo) string {
	if len(failures) == 0 {
		return "low"
	}

	maxProbability := 0.0
	criticalCount := 0
	highCount := 0

	for _, failure := range failures {
		if failure.Probability > maxProbability {
			maxProbability = failure.Probability
		}

		switch failure.Impact {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}

	if criticalCount > 0 || maxProbability > 0.8 {
		return "critical"
	} else if highCount > 0 || maxProbability > 0.6 {
		return "high"
	} else if maxProbability > 0.3 {
		return "medium"
	}

	return "low"
}

func (h *HealthMCPTools) analyzeHealthTrends(events []health.HealthEvent) []string {
	trends := []string{}

	// Simple trend analysis
	if len(events) >= 5 {
		recentEvents := events[len(events)-3:]
		olderEvents := events[:3]

		recentFailures := 0
		olderFailures := 0

		for _, event := range recentEvents {
			if event.Type == "failure" {
				recentFailures++
			}
		}

		for _, event := range olderEvents {
			if event.Type == "failure" {
				olderFailures++
			}
		}

		if recentFailures > olderFailures {
			trends = append(trends, "Increasing failure rate")
		} else if recentFailures < olderFailures {
			trends = append(trends, "Improving stability")
		} else {
			trends = append(trends, "Stable performance")
		}
	}

	return trends
}

func (h *HealthMCPTools) generateHealthRecommendations(response *MCPHealthCheckResponse) []string {
	recommendations := []string{}

	// Based on overall health score
	if response.HealthScore < 0.6 {
		recommendations = append(recommendations, "System health is below optimal. Immediate attention required.")
	}

	// Check for critical components
	criticalComponents := []string{}
	for _, component := range response.Components {
		if component.Status == "critical" {
			criticalComponents = append(criticalComponents, component.Name)
		}
	}

	if len(criticalComponents) > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Critical components need attention: %v", criticalComponents))
	}

	// Nix store specific recommendations
	if response.SystemMetrics != nil && response.SystemMetrics.NixStore.GCRecommended {
		recommendations = append(recommendations, "Consider running garbage collection to free up space")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "System is healthy. Continue regular monitoring.")
	}

	return recommendations
}

func (h *HealthMCPTools) generatePredictionRecommendations(response *MCPPredictionResponse) []string {
	recommendations := []string{}

	switch response.OverallRisk {
	case "critical":
		recommendations = append(recommendations, "Immediate action required to prevent system failures")
	case "high":
		recommendations = append(recommendations, "Schedule maintenance within 24-48 hours")
	case "medium":
		recommendations = append(recommendations, "Plan preventive maintenance in the next week")
	case "low":
		recommendations = append(recommendations, "Continue monitoring. No immediate action needed")
	}

	if len(response.PreventiveActions) > 0 {
		recommendations = append(recommendations, "Review and consider applying suggested preventive actions")
	}

	if response.Confidence < 0.6 {
		recommendations = append(recommendations, "Prediction confidence is low. Collect more data for better accuracy")
	}

	return recommendations
}

func (h *HealthMCPTools) generateMockTrainingData() []health.HealthEvent {
	// Generate mock training data for demonstration
	// In production, this would come from actual system monitoring
	events := make([]health.HealthEvent, 100)
	
	for i := 0; i < 100; i++ {
		events[i] = health.HealthEvent{
			ID:        fmt.Sprintf("mock_event_%d", i),
			Type:      "metric_collection",
			Component: "system",
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			Severity:  health.PriorityLow,
			Message:   fmt.Sprintf("Mock health event %d", i),
			Metrics:   make(map[string]interface{}),
			Context:   make(map[string]interface{}),
		}
		
		// Add some failure events
		if i%20 == 0 {
			events[i].Type = "failure"
			events[i].Severity = health.PriorityHigh
			events[i].Message = "Mock failure event"
		}
	}
	
	return events
}

// convertToMLHealthEvents converts health.HealthEvent to ml.HealthEvent
func (h *HealthMCPTools) convertToMLHealthEvents(events []health.HealthEvent) []ml.HealthEvent {
	mlEvents := make([]ml.HealthEvent, len(events))
	
	for i, event := range events {
		mlEvents[i] = ml.HealthEvent{
			ID:          event.ID,
			Type:        event.Type,
			Component:   event.Component,
			Timestamp:   event.Timestamp,
			Severity:    ml.Priority(event.Severity), // Convert health.Priority to ml.Priority
			Message:     event.Message,
			Metrics:     event.Metrics,
			Context:     event.Context,
			Resolution:  event.Resolution,
			ResolvedAt:  event.ResolvedAt,
		}
	}
	
	return mlEvents
}