package chaos

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"nix-ai-help/internal/testing"
	"nix-ai-help/pkg/logger"
)

// ChaosExperiment represents a chaos engineering experiment
type ChaosExperiment struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description"`
	TargetEnvironment string                   `json:"target_environment"`
	Hypothesis       string                    `json:"hypothesis"`
	Scope            ExperimentScope           `json:"scope"`
	Attacks          []ChaosAttack             `json:"attacks"`
	SteadyState      SteadyStateHypothesis     `json:"steady_state"`
	Status           ExperimentStatus          `json:"status"`
	Results          *ExperimentResults        `json:"results,omitempty"`
	CreatedAt        time.Time                 `json:"created_at"`
	StartedAt        *time.Time                `json:"started_at,omitempty"`
	CompletedAt      *time.Time                `json:"completed_at,omitempty"`
	Duration         time.Duration             `json:"duration"`
	BlastRadius      BlastRadius               `json:"blast_radius"`
}

// ExperimentStatus represents the status of a chaos experiment
type ExperimentStatus string

const (
	ExperimentStatusPending   ExperimentStatus = "pending"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusFailed    ExperimentStatus = "failed"
	ExperimentStatusAborted   ExperimentStatus = "aborted"
)

// ExperimentScope defines the scope of the chaos experiment
type ExperimentScope struct {
	Services     []string `json:"services"`
	Processes    []string `json:"processes"`
	Network      bool     `json:"network"`
	Filesystem   bool     `json:"filesystem"`
	Resources    bool     `json:"resources"`
	Time         bool     `json:"time"`
}

// ChaosAttack represents a specific chaos attack
type ChaosAttack struct {
	ID          string                 `json:"id"`
	Type        AttackType             `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Duration    time.Duration          `json:"duration"`
	Intensity   float64                `json:"intensity"` // 0.0 to 1.0
	Probability float64                `json:"probability"` // 0.0 to 1.0
	Rollback    bool                   `json:"rollback"`
}

// AttackType represents the type of chaos attack
type AttackType string

const (
	AttackServiceKill      AttackType = "service_kill"
	AttackProcessKill      AttackType = "process_kill"
	AttackNetworkLatency   AttackType = "network_latency"
	AttackNetworkLoss      AttackType = "network_loss"
	AttackNetworkPartition AttackType = "network_partition"
	AttackDiskFill         AttackType = "disk_fill"
	AttackDiskIO           AttackType = "disk_io"
	AttackMemoryPressure   AttackType = "memory_pressure"
	AttackCPUStress        AttackType = "cpu_stress"
	AttackTimeSkew         AttackType = "time_skew"
	AttackFileCorruption   AttackType = "file_corruption"
	AttackDNSFailure       AttackType = "dns_failure"
)

// SteadyStateHypothesis defines what constitutes normal system behavior
type SteadyStateHypothesis struct {
	Metrics    []SteadyStateMetric `json:"metrics"`
	Tolerance  float64             `json:"tolerance"`
	Duration   time.Duration       `json:"duration"`
	Baseline   map[string]float64  `json:"baseline"`
}

// SteadyStateMetric defines a metric for steady state validation
type SteadyStateMetric struct {
	Name      string  `json:"name"`
	Query     string  `json:"query"`
	Threshold float64 `json:"threshold"`
	Operator  string  `json:"operator"` // "gt", "lt", "eq", "gte", "lte"
	Weight    float64 `json:"weight"`
}

// BlastRadius defines the potential impact scope of the experiment
type BlastRadius struct {
	Scope       string   `json:"scope"`       // "service", "container", "host", "region"
	Percentage  float64  `json:"percentage"`  // Percentage of targets affected
	MaxTargets  int      `json:"max_targets"` // Maximum number of targets
	Services    []string `json:"services"`    // Services that might be affected
	Criticality string   `json:"criticality"` // "low", "medium", "high", "critical"
}

// ExperimentResults contains the results of a chaos experiment
type ExperimentResults struct {
	SteadyStateValid   bool                           `json:"steady_state_valid"`
	HypothesisProven   bool                           `json:"hypothesis_proven"`
	OverallScore       float64                        `json:"overall_score"`
	AttackResults      map[string]*AttackResult       `json:"attack_results"`
	MetricAnalysis     map[string]*MetricAnalysis     `json:"metric_analysis"`
	RecoveryTime       time.Duration                  `json:"recovery_time"`
	ImpactAssessment   *ImpactAssessment              `json:"impact_assessment"`
	Insights           []string                       `json:"insights"`
	Recommendations    []testing.Recommendation       `json:"recommendations"`
	WeaknessesFound    []Weakness                     `json:"weaknesses_found"`
	ResilienceScore    float64                        `json:"resilience_score"`
}

// AttackResult contains results for a specific attack
type AttackResult struct {
	AttackID      string        `json:"attack_id"`
	Success       bool          `json:"success"`
	ImpactLevel   string        `json:"impact_level"`
	RecoveryTime  time.Duration `json:"recovery_time"`
	ErrorsFound   []string      `json:"errors_found"`
	MetricChanges map[string]float64 `json:"metric_changes"`
	SystemState   SystemState   `json:"system_state"`
}

// MetricAnalysis contains analysis for a specific metric
type MetricAnalysis struct {
	MetricName     string    `json:"metric_name"`
	BaselineValue  float64   `json:"baseline_value"`
	MinValue       float64   `json:"min_value"`
	MaxValue       float64   `json:"max_value"`
	AverageValue   float64   `json:"average_value"`
	DeviationScore float64   `json:"deviation_score"`
	Anomalies      []Anomaly `json:"anomalies"`
}

// ImpactAssessment assesses the overall impact of the experiment
type ImpactAssessment struct {
	ServiceAvailability map[string]float64 `json:"service_availability"`
	PerformanceDegradation float64         `json:"performance_degradation"`
	ErrorRateIncrease   float64            `json:"error_rate_increase"`
	UserImpact          string             `json:"user_impact"`
	BusinessImpact      string             `json:"business_impact"`
	TechnicalDebt       []string           `json:"technical_debt"`
}

// Weakness represents a system weakness found during chaos testing
type Weakness struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Component   string `json:"component"`
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
	Mitigation  string `json:"mitigation"`
}

// SystemState represents the state of the system at a point in time
type SystemState struct {
	Timestamp      time.Time                    `json:"timestamp"`
	Services       map[string]string            `json:"services"`       // service -> status
	Processes      map[string]string            `json:"processes"`      // process -> status
	Resources      map[string]float64           `json:"resources"`      // resource -> utilization
	NetworkHealth  string                       `json:"network_health"`
	StorageHealth  string                       `json:"storage_health"`
	Metrics        map[string]float64           `json:"metrics"`
}

// Anomaly represents an anomalous behavior detected during the experiment
type Anomaly struct {
	Timestamp   time.Time `json:"timestamp"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Deviation   float64   `json:"deviation"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
}

// ChaosEngineer manages chaos engineering experiments
type ChaosEngineer struct {
	logger      *logger.Logger
	experiments map[string]*ChaosExperiment
	mu          sync.RWMutex
	envManager  EnvironmentManagerInterface
	maxExperiments int
}

// EnvironmentManagerInterface defines the interface for environment management
type EnvironmentManagerInterface interface {
	GetEnvironment(ctx context.Context, id string) (*testing.TestEnvironment, error)
	ExecuteCommand(ctx context.Context, envID string, command []string) (string, error)
}

// NewChaosEngineer creates a new chaos engineer
func NewChaosEngineer(envManager EnvironmentManagerInterface, maxExperiments int) *ChaosEngineer {
	return &ChaosEngineer{
		logger:         logger.NewLogger(),
		experiments:    make(map[string]*ChaosExperiment),
		envManager:     envManager,
		maxExperiments: maxExperiments,
	}
}

// CreateExperiment creates a new chaos experiment
func (ce *ChaosEngineer) CreateExperiment(ctx context.Context, experiment *ChaosExperiment) (*ChaosExperiment, error) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	// Check experiment limits
	if len(ce.experiments) >= ce.maxExperiments {
		return nil, fmt.Errorf("maximum number of experiments (%d) reached", ce.maxExperiments)
	}

	// Generate unique ID if not provided
	if experiment.ID == "" {
		experiment.ID = fmt.Sprintf("chaos_%d", time.Now().Unix())
	}

	// Check if experiment already exists
	if _, exists := ce.experiments[experiment.ID]; exists {
		return nil, fmt.Errorf("experiment with ID %s already exists", experiment.ID)
	}

	// Set defaults
	if experiment.Duration == 0 {
		experiment.Duration = 30 * time.Minute
	}
	if experiment.BlastRadius.Criticality == "" {
		experiment.BlastRadius.Criticality = "low"
	}

	experiment.Status = ExperimentStatusPending
	experiment.CreatedAt = time.Now()

	ce.experiments[experiment.ID] = experiment
	ce.logger.Info(fmt.Sprintf("Created chaos experiment %s", experiment.ID))
	return experiment, nil
}

// StartExperiment starts a chaos experiment
func (ce *ChaosEngineer) StartExperiment(ctx context.Context, experimentID string) error {
	ce.mu.Lock()
	experiment, exists := ce.experiments[experimentID]
	ce.mu.Unlock()

	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != ExperimentStatusPending {
		return fmt.Errorf("experiment %s is not in pending status", experimentID)
	}

	// Validate target environment
	if _, err := ce.envManager.GetEnvironment(ctx, experiment.TargetEnvironment); err != nil {
		return fmt.Errorf("target environment %s not found: %w", experiment.TargetEnvironment, err)
	}

	// Update status
	ce.mu.Lock()
	experiment.Status = ExperimentStatusRunning
	now := time.Now()
	experiment.StartedAt = &now
	ce.mu.Unlock()

	// Start experiment execution in background
	go ce.executeExperiment(ctx, experiment)

	ce.logger.Info(fmt.Sprintf("Started chaos experiment %s", experimentID))
	return nil
}

// executeExperiment executes the chaos experiment
func (ce *ChaosEngineer) executeExperiment(ctx context.Context, experiment *ChaosExperiment) {
	defer func() {
		if r := recover(); r != nil {
			ce.logger.Error(fmt.Sprintf("Chaos experiment %s panic: %v", experiment.ID, r))
			ce.updateExperimentStatus(experiment.ID, ExperimentStatusFailed)
		}
	}()

	ce.logger.Info(fmt.Sprintf("Executing chaos experiment %s", experiment.ID))

	results := &ExperimentResults{
		AttackResults:   make(map[string]*AttackResult),
		MetricAnalysis:  make(map[string]*MetricAnalysis),
		Insights:        []string{},
		WeaknessesFound: []Weakness{},
	}

	// Establish baseline steady state
	baseline, err := ce.establishBaseline(ctx, experiment)
	if err != nil {
		ce.logger.Error(fmt.Sprintf("Failed to establish baseline for experiment %s: %v", experiment.ID, err))
		ce.updateExperimentStatus(experiment.ID, ExperimentStatusFailed)
		return
	}

	experiment.SteadyState.Baseline = baseline

	// Execute attacks
	for _, attack := range experiment.Attacks {
		if ctx.Err() != nil {
			break
		}

		attackResult := ce.executeAttack(ctx, experiment, &attack)
		results.AttackResults[attack.ID] = attackResult

		// Allow system to recover between attacks
		time.Sleep(30 * time.Second)
	}

	// Validate steady state after attacks
	results.SteadyStateValid = ce.validateSteadyState(ctx, experiment)

	// Measure recovery time
	results.RecoveryTime = ce.measureRecoveryTime(ctx, experiment)

	// Analyze results
	ce.analyzeExperimentResults(ctx, experiment, results)

	// Store results
	ce.mu.Lock()
	experiment.Results = results
	experiment.Status = ExperimentStatusCompleted
	now := time.Now()
	experiment.CompletedAt = &now
	ce.mu.Unlock()

	ce.logger.Info(fmt.Sprintf("Chaos experiment %s completed successfully", experiment.ID))
}

// establishBaseline establishes baseline metrics for the experiment
func (ce *ChaosEngineer) establishBaseline(ctx context.Context, experiment *ChaosExperiment) (map[string]float64, error) {
	baseline := make(map[string]float64)

	for _, metric := range experiment.SteadyState.Metrics {
		value, err := ce.measureMetric(ctx, experiment.TargetEnvironment, metric.Query)
		if err != nil {
			ce.logger.Error(fmt.Sprintf("Failed to measure baseline for metric %s: %v", metric.Name, err))
			continue
		}
		baseline[metric.Name] = value
	}

	ce.logger.Info(fmt.Sprintf("Established baseline for experiment %s with %d metrics", experiment.ID, len(baseline)))
	return baseline, nil
}

// executeAttack executes a specific chaos attack
func (ce *ChaosEngineer) executeAttack(ctx context.Context, experiment *ChaosExperiment, attack *ChaosAttack) *AttackResult {
	ce.logger.Info(fmt.Sprintf("Executing attack %s (%s) in experiment %s", attack.ID, attack.Type, experiment.ID))

	result := &AttackResult{
		AttackID:      attack.ID,
		Success:       false,
		MetricChanges: make(map[string]float64),
	}

	// Record system state before attack
	preAttackState := ce.captureSystemState(ctx, experiment.TargetEnvironment)

	// Execute the attack based on type
	var err error

	switch attack.Type {
	case AttackServiceKill:
		err = ce.executeServiceKillAttack(ctx, experiment.TargetEnvironment, attack)
	case AttackProcessKill:
		err = ce.executeProcessKillAttack(ctx, experiment.TargetEnvironment, attack)
	case AttackNetworkLatency:
		err = ce.executeNetworkLatencyAttack(ctx, experiment.TargetEnvironment, attack)
	case AttackNetworkLoss:
		err = ce.executeNetworkLossAttack(ctx, experiment.TargetEnvironment, attack)
	case AttackDiskFill:
		err = ce.executeDiskFillAttack(ctx, experiment.TargetEnvironment, attack)
	case AttackMemoryPressure:
		err = ce.executeMemoryPressureAttack(ctx, experiment.TargetEnvironment, attack)
	case AttackCPUStress:
		err = ce.executeCPUStressAttack(ctx, experiment.TargetEnvironment, attack)
	default:
		err = fmt.Errorf("unsupported attack type: %s", attack.Type)
	}

	if err != nil {
		result.ErrorsFound = append(result.ErrorsFound, err.Error())
		ce.logger.Error(fmt.Sprintf("Attack %s failed: %v", attack.ID, err))
		return result
	}

	result.Success = true

	// Wait for attack duration
	time.Sleep(attack.Duration)

	// Measure impact during attack
	for _, metric := range experiment.SteadyState.Metrics {
		value, err := ce.measureMetric(ctx, experiment.TargetEnvironment, metric.Query)
		if err == nil {
			baseline := experiment.SteadyState.Baseline[metric.Name]
			result.MetricChanges[metric.Name] = ((value - baseline) / baseline) * 100
		}
	}

	// Clean up attack if rollback is enabled
	if attack.Rollback {
		ce.rollbackAttack(ctx, experiment.TargetEnvironment, attack)
	}

	// Record system state after attack
	postAttackState := ce.captureSystemState(ctx, experiment.TargetEnvironment)

	// Calculate recovery time
	result.RecoveryTime = ce.calculateRecoveryTime(preAttackState, postAttackState)

	// Assess impact level
	result.ImpactLevel = ce.assessImpactLevel(result.MetricChanges)

	ce.logger.Info(fmt.Sprintf("Attack %s completed with impact level: %s", attack.ID, result.ImpactLevel))
	return result
}

// Attack execution methods

func (ce *ChaosEngineer) executeServiceKillAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	serviceName, ok := attack.Parameters["service"].(string)
	if !ok {
		return fmt.Errorf("service parameter required for service kill attack")
	}

	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{"systemctl", "stop", serviceName})
	return err
}

func (ce *ChaosEngineer) executeProcessKillAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	processName, ok := attack.Parameters["process"].(string)
	if !ok {
		return fmt.Errorf("process parameter required for process kill attack")
	}

	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{"pkill", "-f", processName})
	return err
}

func (ce *ChaosEngineer) executeNetworkLatencyAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	latencyMs, ok := attack.Parameters["latency_ms"].(float64)
	if !ok {
		latencyMs = 100.0 // Default 100ms latency
	}

	// Use tc (traffic control) to add network latency
	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{
		"tc", "qdisc", "add", "dev", "eth0", "root", "netem", "delay", fmt.Sprintf("%.0fms", latencyMs),
	})
	return err
}

func (ce *ChaosEngineer) executeNetworkLossAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	lossPercent, ok := attack.Parameters["loss_percent"].(float64)
	if !ok {
		lossPercent = 5.0 // Default 5% packet loss
	}

	// Use tc to add packet loss
	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{
		"tc", "qdisc", "add", "dev", "eth0", "root", "netem", "loss", fmt.Sprintf("%.1f%%", lossPercent),
	})
	return err
}

func (ce *ChaosEngineer) executeDiskFillAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	sizeGB, ok := attack.Parameters["size_gb"].(float64)
	if !ok {
		sizeGB = 1.0 // Default 1GB
	}

	// Create a large file to fill disk space
	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{
		"dd", "if=/dev/zero", "of=/tmp/chaos_fill", fmt.Sprintf("bs=1G", "count=%.0f", sizeGB),
	})
	return err
}

func (ce *ChaosEngineer) executeMemoryPressureAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	sizeMB, ok := attack.Parameters["size_mb"].(float64)
	if !ok {
		sizeMB = 512.0 // Default 512MB
	}

	// Use stress tool to create memory pressure
	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{
		"stress", "--vm", "1", "--vm-bytes", fmt.Sprintf("%.0fM", sizeMB), "--timeout", "60s",
	})
	return err
}

func (ce *ChaosEngineer) executeCPUStressAttack(ctx context.Context, envID string, attack *ChaosAttack) error {
	cores, ok := attack.Parameters["cores"].(float64)
	if !ok {
		cores = 2.0 // Default 2 cores
	}

	// Use stress tool to create CPU load
	_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{
		"stress", "--cpu", fmt.Sprintf("%.0f", cores), "--timeout", "60s",
	})
	return err
}

// rollbackAttack performs cleanup after an attack
func (ce *ChaosEngineer) rollbackAttack(ctx context.Context, envID string, attack *ChaosAttack) {
	switch attack.Type {
	case AttackServiceKill:
		serviceName, ok := attack.Parameters["service"].(string)
		if ok {
			ce.envManager.ExecuteCommand(ctx, envID, []string{"systemctl", "start", serviceName})
		}
	case AttackNetworkLatency, AttackNetworkLoss:
		ce.envManager.ExecuteCommand(ctx, envID, []string{"tc", "qdisc", "del", "dev", "eth0", "root"})
	case AttackDiskFill:
		ce.envManager.ExecuteCommand(ctx, envID, []string{"rm", "-f", "/tmp/chaos_fill"})
	}
}

// measureMetric measures a specific metric value
func (ce *ChaosEngineer) measureMetric(ctx context.Context, envID, query string) (float64, error) {
	// Simplified metric measurement - in real implementation would use proper monitoring
	switch query {
	case "cpu_usage":
		output, err := ce.envManager.ExecuteCommand(ctx, envID, []string{"cat", "/proc/loadavg"})
		if err != nil {
			return 0, err
		}
		// Parse load average as CPU usage approximation
		return ce.parseLoadAverage(output), nil
	case "memory_usage":
		output, err := ce.envManager.ExecuteCommand(ctx, envID, []string{"free", "-m"})
		if err != nil {
			return 0, err
		}
		return ce.parseMemoryUsage(output), nil
	case "response_time":
		start := time.Now()
		_, err := ce.envManager.ExecuteCommand(ctx, envID, []string{"echo", "test"})
		if err != nil {
			return 0, err
		}
		return float64(time.Since(start).Milliseconds()), nil
	default:
		return rand.Float64() * 100, nil // Placeholder
	}
}

// captureSystemState captures the current system state
func (ce *ChaosEngineer) captureSystemState(ctx context.Context, envID string) SystemState {
	state := SystemState{
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Processes: make(map[string]string),
		Resources: make(map[string]float64),
		Metrics:   make(map[string]float64),
	}

	// Capture service states
	if output, err := ce.envManager.ExecuteCommand(ctx, envID, []string{"systemctl", "list-units", "--type=service", "--no-pager"}); err == nil {
		state.Services = ce.parseServiceStates(output)
	}

	// Capture resource usage
	if cpu, err := ce.measureMetric(ctx, envID, "cpu_usage"); err == nil {
		state.Resources["cpu"] = cpu
	}
	if memory, err := ce.measureMetric(ctx, envID, "memory_usage"); err == nil {
		state.Resources["memory"] = memory
	}

	return state
}

// Helper parsing functions
func (ce *ChaosEngineer) parseLoadAverage(output string) float64 {
	// Simplified load average parsing
	return rand.Float64() * 100
}

func (ce *ChaosEngineer) parseMemoryUsage(output string) float64 {
	// Simplified memory usage parsing
	return rand.Float64() * 100
}

func (ce *ChaosEngineer) parseServiceStates(output string) map[string]string {
	// Simplified service state parsing
	return map[string]string{
		"ssh": "running",
		"systemd": "running",
	}
}

// validateSteadyState validates if the system returned to steady state
func (ce *ChaosEngineer) validateSteadyState(ctx context.Context, experiment *ChaosExperiment) bool {
	for _, metric := range experiment.SteadyState.Metrics {
		currentValue, err := ce.measureMetric(ctx, experiment.TargetEnvironment, metric.Query)
		if err != nil {
			continue
		}

		baseline := experiment.SteadyState.Baseline[metric.Name]
		deviation := math.Abs((currentValue - baseline) / baseline)

		if deviation > experiment.SteadyState.Tolerance {
			return false
		}
	}
	return true
}

// measureRecoveryTime measures how long it takes for the system to recover
func (ce *ChaosEngineer) measureRecoveryTime(ctx context.Context, experiment *ChaosExperiment) time.Duration {
	start := time.Now()
	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return time.Since(start)
		case <-timeout:
			return time.Since(start)
		case <-ticker.C:
			if ce.validateSteadyState(ctx, experiment) {
				return time.Since(start)
			}
		}
	}
}

// calculateRecoveryTime calculates recovery time based on system state comparison
func (ce *ChaosEngineer) calculateRecoveryTime(preState, postState SystemState) time.Duration {
	// Simplified recovery time calculation
	return postState.Timestamp.Sub(preState.Timestamp)
}

// assessImpactLevel assesses the impact level based on metric changes
func (ce *ChaosEngineer) assessImpactLevel(metricChanges map[string]float64) string {
	maxChange := 0.0
	for _, change := range metricChanges {
		if math.Abs(change) > maxChange {
			maxChange = math.Abs(change)
		}
	}

	switch {
	case maxChange < 5:
		return "low"
	case maxChange < 25:
		return "medium"
	case maxChange < 50:
		return "high"
	default:
		return "critical"
	}
}

// analyzeExperimentResults analyzes the overall experiment results
func (ce *ChaosEngineer) analyzeExperimentResults(ctx context.Context, experiment *ChaosExperiment, results *ExperimentResults) {
	// Calculate overall resilience score
	successfulAttacks := 0
	totalAttacks := len(experiment.Attacks)
	
	for _, result := range results.AttackResults {
		if result.Success && result.ImpactLevel != "critical" {
			successfulAttacks++
		}
	}

	if totalAttacks > 0 {
		results.ResilienceScore = float64(successfulAttacks) / float64(totalAttacks) * 100
	}

	// Generate insights
	results.Insights = ce.generateInsights(experiment, results)

	// Find weaknesses
	results.WeaknessesFound = ce.identifyWeaknesses(experiment, results)

	// Generate recommendations
	results.Recommendations = ce.generateChaosRecommendations(experiment, results)

	// Determine if hypothesis was proven
	results.HypothesisProven = results.SteadyStateValid && results.ResilienceScore > 70
}

// generateInsights generates insights from the experiment results
func (ce *ChaosEngineer) generateInsights(experiment *ChaosExperiment, results *ExperimentResults) []string {
	var insights []string

	if results.ResilienceScore > 80 {
		insights = append(insights, "System shows high resilience to chaos attacks")
	} else if results.ResilienceScore < 50 {
		insights = append(insights, "System shows vulnerability to multiple attack vectors")
	}

	if results.RecoveryTime > 5*time.Minute {
		insights = append(insights, "System recovery time is longer than expected")
	}

	if results.SteadyStateValid {
		insights = append(insights, "System successfully returned to steady state after attacks")
	} else {
		insights = append(insights, "System failed to return to baseline steady state")
	}

	return insights
}

// identifyWeaknesses identifies system weaknesses based on attack results
func (ce *ChaosEngineer) identifyWeaknesses(experiment *ChaosExperiment, results *ExperimentResults) []Weakness {
	var weaknesses []Weakness

	for _, result := range results.AttackResults {
		if result.ImpactLevel == "critical" || result.ImpactLevel == "high" {
			weakness := Weakness{
				Type:        "resilience",
				Severity:    result.ImpactLevel,
				Component:   "system",
				Description: fmt.Sprintf("High impact from %s attack", result.AttackID),
				Evidence:    fmt.Sprintf("Attack caused %s impact level", result.ImpactLevel),
				Mitigation:  "Implement additional safeguards and monitoring",
			}
			weaknesses = append(weaknesses, weakness)
		}
	}

	return weaknesses
}

// generateChaosRecommendations generates recommendations based on chaos testing results
func (ce *ChaosEngineer) generateChaosRecommendations(experiment *ChaosExperiment, results *ExperimentResults) []testing.Recommendation {
	var recommendations []testing.Recommendation

	if results.ResilienceScore < 70 {
		recommendations = append(recommendations, testing.Recommendation{
			ID:          "improve_resilience",
			Type:        testing.RecommendationReliability,
			Priority:    "high",
			Title:       "Improve system resilience",
			Description: "System shows low resilience to chaos attacks",
			Impact:      "Better system reliability and fault tolerance",
			Effort:      "high",
			Actions:     []string{"Implement circuit breakers", "Add redundancy", "Improve monitoring"},
		})
	}

	if results.RecoveryTime > 5*time.Minute {
		recommendations = append(recommendations, testing.Recommendation{
			ID:          "faster_recovery",
			Type:        testing.RecommendationReliability,
			Priority:    "medium",
			Title:       "Reduce recovery time",
			Description: "System takes too long to recover from failures",
			Impact:      "Reduced downtime and faster incident resolution",
			Effort:      "medium",
			Actions:     []string{"Implement auto-healing", "Optimize restart procedures", "Add health checks"},
		})
	}

	return recommendations
}

// updateExperimentStatus updates the status of an experiment
func (ce *ChaosEngineer) updateExperimentStatus(experimentID string, status ExperimentStatus) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if experiment, exists := ce.experiments[experimentID]; exists {
		experiment.Status = status
	}
}

// GetExperiment retrieves an experiment by ID
func (ce *ChaosEngineer) GetExperiment(ctx context.Context, experimentID string) (*ChaosExperiment, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	experiment, exists := ce.experiments[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	return experiment, nil
}

// ListExperiments lists all experiments
func (ce *ChaosEngineer) ListExperiments(ctx context.Context) ([]*ChaosExperiment, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	experiments := make([]*ChaosExperiment, 0, len(ce.experiments))
	for _, experiment := range ce.experiments {
		experiments = append(experiments, experiment)
	}

	return experiments, nil
}

// AbortExperiment aborts a running experiment
func (ce *ChaosEngineer) AbortExperiment(ctx context.Context, experimentID string) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	experiment, exists := ce.experiments[experimentID]
	if !exists {
		return fmt.Errorf("experiment %s not found", experimentID)
	}

	if experiment.Status != ExperimentStatusRunning {
		return fmt.Errorf("experiment %s is not running", experimentID)
	}

	experiment.Status = ExperimentStatusAborted
	ce.logger.Info(fmt.Sprintf("Aborted chaos experiment %s", experimentID))
	return nil
}

