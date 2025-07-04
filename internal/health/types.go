// Package health provides predictive system health monitoring and failure prediction
package health

import (
	"context"
	"time"
)

// HealthPredictor defines the interface for system health prediction
type HealthPredictor interface {
	// PredictFailures analyzes system state and predicts potential failures
	PredictFailures(ctx context.Context, timeline time.Duration) (*FailurePrediction, error)
	
	// AnalyzeSystemHealth provides current health assessment
	AnalyzeSystemHealth(ctx context.Context) (*HealthAssessment, error)
	
	// ForecastResources predicts resource usage patterns
	ForecastResources(ctx context.Context, timeline time.Duration) (*ResourceForecast, error)
	
	// DetectAnomalies identifies unusual system behavior
	DetectAnomalies(ctx context.Context) (*AnomalyReport, error)
	
	// GetRemediationSuggestions provides automated fix suggestions
	GetRemediationSuggestions(ctx context.Context, issues []HealthIssue) (*RemediationPlan, error)
}

// FailurePrediction represents predicted system failures
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

// PredictedFailure represents a specific predicted failure
type PredictedFailure struct {
	ID               string                 `json:"id"`
	Type             FailureType            `json:"type"`
	Component        string                 `json:"component"`
	Description      string                 `json:"description"`
	ProbabilityScore float64                `json:"probability_score"` // 0.0 - 1.0
	EstimatedTime    time.Time              `json:"estimated_time"`
	Impact           ImpactLevel            `json:"impact"`
	Indicators       []HealthIndicator      `json:"indicators"`
	HistoricalData   []HistoricalEvent      `json:"historical_data"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// HealthAssessment provides current system health status
type HealthAssessment struct {
	OverallHealth     HealthStatus           `json:"overall_health"`
	ComponentHealth   map[string]HealthStatus `json:"component_health"`
	ActiveIssues      []HealthIssue          `json:"active_issues"`
	PerformanceMetrics map[string]float64    `json:"performance_metrics"`
	ResourceUtilization ResourceUtilization  `json:"resource_utilization"`
	SecurityStatus    SecurityStatus         `json:"security_status"`
	TrendAnalysis     TrendAnalysis          `json:"trend_analysis"`
	LastUpdate        time.Time              `json:"last_update"`
	Recommendations   []Recommendation       `json:"recommendations"`
}

// ResourceForecast predicts future resource usage
type ResourceForecast struct {
	Timeline         time.Duration                    `json:"timeline"`
	CPUForecast      []ResourceDataPoint              `json:"cpu_forecast"`
	MemoryForecast   []ResourceDataPoint              `json:"memory_forecast"`
	DiskForecast     []ResourceDataPoint              `json:"disk_forecast"`
	NetworkForecast  []ResourceDataPoint              `json:"network_forecast"`
	Predictions      map[string]ResourcePrediction    `json:"predictions"`
	Thresholds       map[string]ResourceThreshold     `json:"thresholds"`
	Alerts           []ResourceAlert                  `json:"alerts"`
	ModelAccuracy    float64                          `json:"model_accuracy"`
	Confidence       float64                          `json:"confidence"`
	GeneratedAt      time.Time                        `json:"generated_at"`
	Metadata         map[string]interface{}           `json:"metadata"`
}

// AnomalyReport identifies unusual system behavior
type AnomalyReport struct {
	DetectedAnomalies []Anomaly              `json:"detected_anomalies"`
	AnomalyScore      float64                `json:"anomaly_score"`
	BaselineDeviation float64                `json:"baseline_deviation"`
	DetectionModel    string                 `json:"detection_model"`
	TimeWindow        time.Duration          `json:"time_window"`
	GeneratedAt       time.Time              `json:"generated_at"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// RemediationPlan provides automated fix suggestions
type RemediationPlan struct {
	ID              string                 `json:"id"`
	Issues          []HealthIssue          `json:"issues"`
	Suggestions     []RemediationSuggestion `json:"suggestions"`
	AutomationLevel AutomationLevel        `json:"automation_level"`
	EstimatedTime   time.Duration          `json:"estimated_time"`
	RiskAssessment  RiskAssessment         `json:"risk_assessment"`
	Dependencies    []string               `json:"dependencies"`
	Rollback        []RollbackStep         `json:"rollback"`
	CreatedAt       time.Time              `json:"created_at"`
	Priority        Priority               `json:"priority"`
}

// Enums and constants

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

type HealthStatus string

const (
	HealthExcellent HealthStatus = "excellent"
	HealthGood      HealthStatus = "good"
	HealthFair      HealthStatus = "fair"
	HealthPoor      HealthStatus = "poor"
	HealthCritical  HealthStatus = "critical"
	HealthUnknown   HealthStatus = "unknown"
)

type AutomationLevel string

const (
	AutomationNone   AutomationLevel = "none"
	AutomationLow    AutomationLevel = "low"
	AutomationMedium AutomationLevel = "medium"
	AutomationHigh   AutomationLevel = "high"
	AutomationFull   AutomationLevel = "full"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// Supporting structures

type PreventiveAction struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Commands    []string               `json:"commands"`
	Automated   bool                   `json:"automated"`
	Risk        RiskLevel              `json:"risk"`
	ETA         time.Duration          `json:"eta"`
	Dependencies []string              `json:"dependencies"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type HealthIndicator struct {
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Unit      string    `json:"unit"`
	Trend     string    `json:"trend"` // "increasing", "decreasing", "stable"
	Severity  Priority  `json:"severity"`
	Source    string    `json:"source"`
}

type HistoricalEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Severity    Priority               `json:"severity"`
	Resolution  string                 `json:"resolution,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type HealthIssue struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Component   string                 `json:"component"`
	Description string                 `json:"description"`
	Severity    Priority               `json:"severity"`
	Status      string                 `json:"status"`
	DetectedAt  time.Time              `json:"detected_at"`
	Indicators  []HealthIndicator      `json:"indicators"`
	Suggestions []string               `json:"suggestions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ResourceUtilization struct {
	CPU      ResourceMetric `json:"cpu"`
	Memory   ResourceMetric `json:"memory"`
	Disk     ResourceMetric `json:"disk"`
	Network  ResourceMetric `json:"network"`
	GPU      ResourceMetric `json:"gpu,omitempty"`
	LoadAvg  []float64      `json:"load_avg"`
	Processes int           `json:"processes"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type ResourceMetric struct {
	Current     float64   `json:"current"`
	Average     float64   `json:"average"`
	Peak        float64   `json:"peak"`
	Unit        string    `json:"unit"`
	Threshold   float64   `json:"threshold"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"last_updated"`
}

type SecurityStatus struct {
	ThreatLevel       string               `json:"threat_level"`
	VulnerabilityCount int                 `json:"vulnerability_count"`
	LastSecurityScan   time.Time           `json:"last_security_scan"`
	ActiveThreats     []SecurityThreat     `json:"active_threats"`
	SecurityScore     float64              `json:"security_score"`
	Recommendations   []SecurityAction     `json:"recommendations"`
}

type SecurityThreat struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    Priority  `json:"severity"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detected_at"`
	Status      string    `json:"status"`
	Mitigation  string    `json:"mitigation"`
}

type SecurityAction struct {
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	Automated   bool      `json:"automated"`
	Commands    []string  `json:"commands"`
}

type TrendAnalysis struct {
	CPUTrend        TrendData `json:"cpu_trend"`
	MemoryTrend     TrendData `json:"memory_trend"`
	DiskTrend       TrendData `json:"disk_trend"`
	NetworkTrend    TrendData `json:"network_trend"`
	PerformanceTrend TrendData `json:"performance_trend"`
	AnalysisPeriod  time.Duration `json:"analysis_period"`
	GeneratedAt     time.Time  `json:"generated_at"`
}

type TrendData struct {
	Direction   string  `json:"direction"` // "increasing", "decreasing", "stable"
	Rate        float64 `json:"rate"`      // Rate of change per unit time
	Confidence  float64 `json:"confidence"`
	Prediction  float64 `json:"prediction"` // Predicted value at end of timeline
	Volatility  float64 `json:"volatility"`
}

type Recommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    Priority  `json:"priority"`
	Category    string    `json:"category"`
	Actions     []string  `json:"actions"`
	Benefits    []string  `json:"benefits"`
	Risks       []string  `json:"risks"`
	Effort      string    `json:"effort"` // "low", "medium", "high"
}

type ResourceDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Predicted bool      `json:"predicted"`
	Confidence float64  `json:"confidence"`
}

type ResourcePrediction struct {
	Resource      string                 `json:"resource"`
	CurrentValue  float64                `json:"current_value"`
	PredictedValue float64               `json:"predicted_value"`
	ChangeRate    float64                `json:"change_rate"`
	TimeToThreshold time.Duration        `json:"time_to_threshold"`
	Confidence    float64                `json:"confidence"`
	Model         string                 `json:"model"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type ResourceThreshold struct {
	Warning   float64 `json:"warning"`
	Critical  float64 `json:"critical"`
	Maximum   float64 `json:"maximum"`
	Unit      string  `json:"unit"`
	AlertsEnabled bool `json:"alerts_enabled"`
}

type ResourceAlert struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"`
	Type        string    `json:"type"` // "warning", "critical", "prediction"
	Message     string    `json:"message"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	PredictedTime time.Time `json:"predicted_time,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Acknowledged bool     `json:"acknowledged"`
}

type Anomaly struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Component   string                 `json:"component"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`        // Anomaly score (higher = more anomalous)
	Severity    Priority               `json:"severity"`
	DetectedAt  time.Time              `json:"detected_at"`
	Evidence    []AnomalyEvidence      `json:"evidence"`
	Context     map[string]interface{} `json:"context"`
	Status      string                 `json:"status"`
}

type AnomalyEvidence struct {
	Metric      string    `json:"metric"`
	ExpectedValue float64 `json:"expected_value"`
	ActualValue float64   `json:"actual_value"`
	Deviation   float64   `json:"deviation"`
	Timestamp   time.Time `json:"timestamp"`
	Confidence  float64   `json:"confidence"`
}

type RemediationSuggestion struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"` // "automated", "manual", "hybrid"
	Actions     []RemediationAction    `json:"actions"`
	Priority    Priority               `json:"priority"`
	Risk        RiskLevel              `json:"risk"`
	Confidence  float64                `json:"confidence"`
	Effort      string                 `json:"effort"`
	Benefits    []string               `json:"benefits"`
	Prerequisites []string             `json:"prerequisites"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type RemediationAction struct {
	Step        int      `json:"step"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Automated   bool     `json:"automated"`
	Validation  string   `json:"validation"`
	Rollback    string   `json:"rollback"`
	Timeout     time.Duration `json:"timeout"`
	CriticalPath bool    `json:"critical_path"`
}

type RiskAssessment struct {
	OverallRisk     RiskLevel              `json:"overall_risk"`
	RiskFactors     []RiskFactor           `json:"risk_factors"`
	MitigationPlan  []string               `json:"mitigation_plan"`
	SuccessProbability float64             `json:"success_probability"`
	RollbackRisk    RiskLevel              `json:"rollback_risk"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type RiskFactor struct {
	Factor      string    `json:"factor"`
	Impact      ImpactLevel `json:"impact"`
	Probability float64   `json:"probability"`
	Mitigation  string    `json:"mitigation"`
}

type RollbackStep struct {
	Step        int      `json:"step"`
	Description string   `json:"description"`
	Commands    []string `json:"commands"`
	Validation  string   `json:"validation"`
	Timeout     time.Duration `json:"timeout"`
	Critical    bool     `json:"critical"`
}

// Configuration types

type HealthConfig struct {
	MonitoringInterval    time.Duration `json:"monitoring_interval"`
	PredictionTimeline    time.Duration `json:"prediction_timeline"`
	AnomalyThreshold      float64       `json:"anomaly_threshold"`
	EnableAutoRemediation bool          `json:"enable_auto_remediation"`
	ResourceThresholds    map[string]ResourceThreshold `json:"resource_thresholds"`
	SecurityScanInterval  time.Duration `json:"security_scan_interval"`
	ModelUpdateInterval   time.Duration `json:"model_update_interval"`
	DataRetentionPeriod   time.Duration `json:"data_retention_period"`
	AlertingEnabled       bool          `json:"alerting_enabled"`
	LogLevel              string        `json:"log_level"`
}

// Event types for monitoring and ML training

type HealthEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Component   string                 `json:"component"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    Priority               `json:"severity"`
	Message     string                 `json:"message"`
	Metrics     map[string]interface{} `json:"metrics"`
	Context     map[string]interface{} `json:"context"`
	Resolution  string                 `json:"resolution,omitempty"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// ML Model interfaces

type MLModel interface {
	// Train the model with historical data
	Train(ctx context.Context, data []HealthEvent) error
	
	// Predict using the trained model
	Predict(ctx context.Context, input interface{}) (interface{}, error)
	
	// Evaluate model performance
	Evaluate(ctx context.Context, testData []HealthEvent) (*ModelMetrics, error)
	
	// Get model information
	GetInfo() ModelInfo
	
	// Update model with new data
	Update(ctx context.Context, data []HealthEvent) error
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