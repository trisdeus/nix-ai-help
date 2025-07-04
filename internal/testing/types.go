package testing

import (
	"context"
	"time"
)

// TestEnvironment represents a virtual testing environment
type TestEnvironment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Status          EnvironmentStatus      `json:"status"`
	Configuration   string                 `json:"configuration"`
	BaseImage       string                 `json:"base_image"`
	Resources       ResourceAllocation     `json:"resources"`
	CreatedAt       time.Time              `json:"created_at"`
	LastModified    time.Time              `json:"last_modified"`
	Metrics         *EnvironmentMetrics    `json:"metrics,omitempty"`
	Snapshots       []Snapshot             `json:"snapshots"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// EnvironmentStatus represents the status of a test environment
type EnvironmentStatus string

const (
	StatusCreating   EnvironmentStatus = "creating"
	StatusRunning    EnvironmentStatus = "running"
	StatusStopping   EnvironmentStatus = "stopping"
	StatusStopped    EnvironmentStatus = "stopped"
	StatusFailed     EnvironmentStatus = "failed"
	StatusDestroyed  EnvironmentStatus = "destroyed"
)

// ResourceAllocation defines resource limits for test environments
type ResourceAllocation struct {
	CPUCores   int    `json:"cpu_cores"`
	MemoryMB   int    `json:"memory_mb"`
	DiskGB     int    `json:"disk_gb"`
	NetworkMBs int    `json:"network_mbs"`
}

// EnvironmentMetrics contains performance metrics for a test environment
type EnvironmentMetrics struct {
	CPUUsage       float64           `json:"cpu_usage"`
	MemoryUsage    float64           `json:"memory_usage"`
	DiskUsage      float64           `json:"disk_usage"`
	NetworkTraffic float64           `json:"network_traffic"`
	BootTime       time.Duration     `json:"boot_time"`
	ServiceHealth  map[string]string `json:"service_health"`
	LastUpdated    time.Time         `json:"last_updated"`
}

// Snapshot represents a snapshot of a test environment
type Snapshot struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
}

// TestConfiguration represents a configuration to be tested
type TestConfiguration struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Content          string                 `json:"content"`
	Type             ConfigurationType      `json:"type"`
	ValidationRules  []ValidationRule       `json:"validation_rules"`
	ExpectedOutcomes []ExpectedOutcome      `json:"expected_outcomes"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ConfigurationType represents the type of NixOS configuration
type ConfigurationType string

const (
	ConfigurationSystem   ConfigurationType = "system"
	ConfigurationService  ConfigurationType = "service"
	ConfigurationPackage  ConfigurationType = "package"
	ConfigurationModule   ConfigurationType = "module"
	ConfigurationComplete ConfigurationType = "complete"
)

// ValidationRule defines a rule for validating configuration
type ValidationRule struct {
	Name        string              `json:"name"`
	Type        ValidationRuleType  `json:"type"`
	Expression  string              `json:"expression"`
	Severity    ValidationSeverity  `json:"severity"`
	Description string              `json:"description"`
	Enabled     bool                `json:"enabled"`
}

// ValidationRuleType represents the type of validation rule
type ValidationRuleType string

const (
	RuleTypeSyntax      ValidationRuleType = "syntax"
	RuleTypeSemantic    ValidationRuleType = "semantic"
	RuleTypeSecurity    ValidationRuleType = "security"
	RuleTypePerformance ValidationRuleType = "performance"
	RuleTypeCustom      ValidationRuleType = "custom"
)

// ValidationSeverity represents the severity of a validation rule
type ValidationSeverity string

const (
	SeverityInfo     ValidationSeverity = "info"
	SeverityWarning  ValidationSeverity = "warning"
	SeverityError    ValidationSeverity = "error"
	SeverityCritical ValidationSeverity = "critical"
)

// ExpectedOutcome defines what we expect from a configuration test
type ExpectedOutcome struct {
	Type        OutcomeType `json:"type"`
	Metric      string      `json:"metric"`
	Operator    string      `json:"operator"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
}

// OutcomeType represents the type of expected outcome
type OutcomeType string

const (
	OutcomePerformance OutcomeType = "performance"
	OutcomeResource    OutcomeType = "resource"
	OutcomeSecurity    OutcomeType = "security"
	OutcomeService     OutcomeType = "service"
	OutcomeCustom      OutcomeType = "custom"
)

// TestResults represents the results of a configuration test
type TestResults struct {
	TestID           string              `json:"test_id"`
	ConfigurationID  string              `json:"configuration_id"`
	EnvironmentID    string              `json:"environment_id"`
	Status           TestStatus          `json:"status"`
	StartTime        time.Time           `json:"start_time"`
	EndTime          *time.Time          `json:"end_time,omitempty"`
	Duration         time.Duration       `json:"duration"`
	ValidationErrors []ValidationError   `json:"validation_errors"`
	PerformanceData  *PerformanceData    `json:"performance_data,omitempty"`
	SecurityAnalysis *SecurityAnalysis   `json:"security_analysis,omitempty"`
	RollbackPlan     *RollbackPlan       `json:"rollback_plan,omitempty"`
	Recommendations  []Recommendation    `json:"recommendations"`
	Artifacts        []TestArtifact      `json:"artifacts"`
}

// TestStatus represents the status of a test
type TestStatus string

const (
	TestStatusPending    TestStatus = "pending"
	TestStatusRunning    TestStatus = "running"
	TestStatusCompleted  TestStatus = "completed"
	TestStatusFailed     TestStatus = "failed"
	TestStatusCancelled  TestStatus = "cancelled"
	TestStatusTimeout    TestStatus = "timeout"
)

// ValidationError represents a validation error found during testing
type ValidationError struct {
	Rule        string             `json:"rule"`
	Type        ValidationRuleType `json:"type"`
	Severity    ValidationSeverity `json:"severity"`
	Message     string             `json:"message"`
	Line        int                `json:"line,omitempty"`
	Column      int                `json:"column,omitempty"`
	Suggestion  string             `json:"suggestion,omitempty"`
}

// PerformanceData contains performance metrics from testing
type PerformanceData struct {
	BootTime         time.Duration            `json:"boot_time"`
	ServiceStartTime map[string]time.Duration `json:"service_start_time"`
	ResourceUsage    ResourceUsage            `json:"resource_usage"`
	Benchmarks       map[string]float64       `json:"benchmarks"`
	LoadTests        []LoadTestResult         `json:"load_tests"`
}

// ResourceUsage represents resource utilization metrics
type ResourceUsage struct {
	CPU    ResourceMetric `json:"cpu"`
	Memory ResourceMetric `json:"memory"`
	Disk   ResourceMetric `json:"disk"`
	Network ResourceMetric `json:"network"`
}

// ResourceMetric represents metrics for a specific resource
type ResourceMetric struct {
	Average float64 `json:"average"`
	Peak    float64 `json:"peak"`
	Minimum float64 `json:"minimum"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
}

// LoadTestResult represents results from load testing
type LoadTestResult struct {
	Name           string        `json:"name"`
	Duration       time.Duration `json:"duration"`
	RequestsPerSec float64       `json:"requests_per_sec"`
	ErrorRate      float64       `json:"error_rate"`
	ResponseTime   ResourceMetric `json:"response_time"`
}

// SecurityAnalysis contains security analysis results
type SecurityAnalysis struct {
	SecurityScore    float64            `json:"security_score"`
	Vulnerabilities  []Vulnerability    `json:"vulnerabilities"`
	SecurityRules    []SecurityRuleResult `json:"security_rules"`
	ComplianceCheck  map[string]bool    `json:"compliance_check"`
	ThreatAnalysis   *ThreatAnalysis    `json:"threat_analysis,omitempty"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string             `json:"id"`
	Type        string             `json:"type"`
	Severity    ValidationSeverity `json:"severity"`
	Description string             `json:"description"`
	Component   string             `json:"component"`
	CVE         string             `json:"cve,omitempty"`
	Mitigation  string             `json:"mitigation"`
}

// SecurityRuleResult represents the result of a security rule check
type SecurityRuleResult struct {
	Rule    string `json:"rule"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
	Score   float64 `json:"score"`
}

// ThreatAnalysis represents threat modeling results
type ThreatAnalysis struct {
	AttackVectors []string           `json:"attack_vectors"`
	RiskLevel     string             `json:"risk_level"`
	Mitigations   []string           `json:"mitigations"`
	Recommendations []Recommendation `json:"recommendations"`
}

// RollbackPlan contains rollback strategy and procedures
type RollbackPlan struct {
	ID                 string               `json:"id"`
	EstimatedTime      time.Duration        `json:"estimated_time"`
	SuccessProbability float64              `json:"success_probability"`
	Steps              []RollbackStep       `json:"steps"`
	Prerequisites      []string             `json:"prerequisites"`
	VerificationSteps  []VerificationStep   `json:"verification_steps"`
	RiskAssessment     *RiskAssessment      `json:"risk_assessment"`
}

// RollbackStep represents a single step in the rollback process
type RollbackStep struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Command     string        `json:"command"`
	Timeout     time.Duration `json:"timeout"`
	Critical    bool          `json:"critical"`
	Rollbackable bool         `json:"rollbackable"`
}

// VerificationStep represents a verification step after rollback
type VerificationStep struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	ExpectedResult string `json:"expected_result"`
	Critical    bool   `json:"critical"`
}

// RiskAssessment represents risk analysis for rollback
type RiskAssessment struct {
	OverallRisk    string             `json:"overall_risk"`
	DataLossRisk   string             `json:"data_loss_risk"`
	DowntimeRisk   string             `json:"downtime_risk"`
	ServiceRisk    map[string]string  `json:"service_risk"`
	Mitigations    []string           `json:"mitigations"`
}

// Recommendation represents an optimization or improvement suggestion
type Recommendation struct {
	ID          string             `json:"id"`
	Type        RecommendationType `json:"type"`
	Priority    string             `json:"priority"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Impact      string             `json:"impact"`
	Effort      string             `json:"effort"`
	Actions     []string           `json:"actions"`
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationPerformance RecommendationType = "performance"
	RecommendationSecurity    RecommendationType = "security"
	RecommendationReliability RecommendationType = "reliability"
	RecommendationMaintenance RecommendationType = "maintenance"
	RecommendationCost        RecommendationType = "cost"
)

// TestArtifact represents an artifact generated during testing
type TestArtifact struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
}

// TestManager interface defines the main testing operations
type TestManager interface {
	// Environment Management
	CreateEnvironment(ctx context.Context, config *TestEnvironment) (*TestEnvironment, error)
	GetEnvironment(ctx context.Context, id string) (*TestEnvironment, error)
	ListEnvironments(ctx context.Context) ([]*TestEnvironment, error)
	UpdateEnvironment(ctx context.Context, env *TestEnvironment) error
	DeleteEnvironment(ctx context.Context, id string) error
	
	// Testing Operations
	RunTest(ctx context.Context, config *TestConfiguration, envID string) (*TestResults, error)
	GetTestResults(ctx context.Context, testID string) (*TestResults, error)
	ListTests(ctx context.Context, filter *TestFilter) ([]*TestResults, error)
	CancelTest(ctx context.Context, testID string) error
	
	// Snapshot Management
	CreateSnapshot(ctx context.Context, envID string, name string) (*Snapshot, error)
	RestoreSnapshot(ctx context.Context, envID string, snapshotID string) error
	DeleteSnapshot(ctx context.Context, snapshotID string) error
	
	// Rollback Operations
	GenerateRollbackPlan(ctx context.Context, config *TestConfiguration) (*RollbackPlan, error)
	ExecuteRollback(ctx context.Context, plan *RollbackPlan) error
	ValidateRollback(ctx context.Context, plan *RollbackPlan) error
}

// TestFilter represents filtering options for test queries
type TestFilter struct {
	Status          []TestStatus   `json:"status,omitempty"`
	ConfigurationID string         `json:"configuration_id,omitempty"`
	EnvironmentID   string         `json:"environment_id,omitempty"`
	StartTime       *time.Time     `json:"start_time,omitempty"`
	EndTime         *time.Time     `json:"end_time,omitempty"`
	Limit           int            `json:"limit,omitempty"`
	Offset          int            `json:"offset,omitempty"`
}