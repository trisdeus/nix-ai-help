// Package intelligence provides smart recommendations engine for NixOS systems
package intelligence

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// RecommendationsEngine provides intelligent recommendations based on system analysis
type RecommendationsEngine struct {
	logger             *logger.Logger
	aiProvider         ai.Provider
	systemAnalyzer     *SystemAnalyzer
	predictor          *Predictor
	conflictDetector   *ConflictDetector
	dependencyAnalyzer *DependencyAnalyzer
	cache              map[string]*RecommendationSet
	mu                 sync.RWMutex
}

// RecommendationSet contains comprehensive recommendations for the system
type RecommendationSet struct {
	// Core Recommendations
	SystemOptimizations          []SystemRecommendation        `json:"system_optimizations"`
	SecurityRecommendations      []SecurityRecommendation      `json:"security_recommendations"`
	PerformanceRecommendations   []PerformanceRecommendation   `json:"performance_recommendations"`
	MaintenanceRecommendations   []MaintenanceRecommendation   `json:"maintenance_recommendations"`
	ConfigurationRecommendations []ConfigurationRecommendation `json:"configuration_recommendations"`

	// Contextual Recommendations
	UserSpecificRecommendations []UserRecommendation        `json:"user_specific_recommendations"`
	WorkflowRecommendations     []WorkflowRecommendation    `json:"workflow_recommendations"`
	EnvironmentRecommendations  []EnvironmentRecommendation `json:"environment_recommendations"`

	// Meta Information
	TotalRecommendations int            `json:"total_recommendations"`
	PriorityBreakdown    map[string]int `json:"priority_breakdown"`
	CategoryBreakdown    map[string]int `json:"category_breakdown"`
	EstimatedImpact      string         `json:"estimated_impact"`
	TimeToImplement      time.Duration  `json:"time_to_implement"`

	// Analysis Context
	BasedOnAnalysis  []string          `json:"based_on_analysis"`
	SystemSnapshot   string            `json:"system_snapshot"`
	GeneratedAt      time.Time         `json:"generated_at"`
	Confidence       float64           `json:"confidence"`
	AIReasoningChain []AIReasoningStep `json:"ai_reasoning_chain"`
}

// SystemRecommendation represents a system-level optimization recommendation
type SystemRecommendation struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"` // optimization, cleanup, update, migration
	Priority    string  `json:"priority"` // critical, high, medium, low
	Confidence  float64 `json:"confidence"`

	// Implementation Details
	Actions       []ActionStep `json:"actions"`
	Prerequisites []string     `json:"prerequisites"`
	Verification  []string     `json:"verification"`
	Rollback      []string     `json:"rollback"`

	// Impact Analysis
	Benefits           []string `json:"benefits"`
	Risks              []string `json:"risks"`
	SideEffects        []string `json:"side_effects"`
	AffectedComponents []string `json:"affected_components"`

	// Resource Requirements
	EstimatedTime   time.Duration `json:"estimated_time"`
	RequiredSkill   string        `json:"required_skill"` // beginner, intermediate, advanced
	SystemDowntime  time.Duration `json:"system_downtime"`
	DiskSpaceImpact int64         `json:"disk_space_impact"`

	// Context
	TriggeredBy   []string `json:"triggered_by"`
	RelatedIssues []string `json:"related_issues"`
	Documentation []string `json:"documentation"`
	Tags          []string `json:"tags"`
}

// ActionStep represents a specific action to take
type ActionStep struct {
	StepNumber      int    `json:"step_number"`
	Description     string `json:"description"`
	Command         string `json:"command,omitempty"`
	ExpectedOutput  string `json:"expected_output,omitempty"`
	FailureHandling string `json:"failure_handling,omitempty"`
	Critical        bool   `json:"critical"`
	Automated       bool   `json:"automated"`
}

// SecurityRecommendation represents security-focused recommendations
type SecurityRecommendation struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Severity       string   `json:"severity"`        // critical, high, medium, low
	SecurityDomain string   `json:"security_domain"` // network, filesystem, access, crypto
	ThreatModel    []string `json:"threat_model"`

	Actions             []ActionStep     `json:"actions"`
	ComplianceStandards []string         `json:"compliance_standards"`
	SecurityMetrics     []SecurityMetric `json:"security_metrics"`

	Confidence    float64       `json:"confidence"`
	EstimatedTime time.Duration `json:"estimated_time"`
	RequiredSkill string        `json:"required_skill"`
}

// SecurityMetric represents a measurable security improvement
type SecurityMetric struct {
	Name            string `json:"name"`
	CurrentValue    string `json:"current_value"`
	TargetValue     string `json:"target_value"`
	ImprovementType string `json:"improvement_type"` // increase, decrease, enable, disable
}

// PerformanceRecommendation represents performance optimization recommendations
type PerformanceRecommendation struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	PerformanceArea string `json:"performance_area"` // cpu, memory, disk, network, boot

	CurrentMetrics  []PerformanceMetric `json:"current_metrics"`
	ExpectedGains   []PerformanceGain   `json:"expected_gains"`
	Actions         []ActionStep        `json:"actions"`
	MonitoringSetup []string            `json:"monitoring_setup"`

	Priority      string        `json:"priority"`
	Confidence    float64       `json:"confidence"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Reversible    bool          `json:"reversible"`
}

// PerformanceMetric represents a performance measurement
type PerformanceMetric struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Baseline  float64 `json:"baseline"`
	Threshold float64 `json:"threshold"`
}

// PerformanceGain represents expected performance improvement
type PerformanceGain struct {
	Metric         string  `json:"metric"`
	ExpectedChange string  `json:"expected_change"`
	Confidence     float64 `json:"confidence"`
	TimeFrame      string  `json:"time_frame"`
}

// MaintenanceRecommendation represents system maintenance recommendations
type MaintenanceRecommendation struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	MaintenanceType string `json:"maintenance_type"` // preventive, corrective, adaptive, perfective

	Schedule   MaintenanceSchedule `json:"schedule"`
	Actions    []ActionStep        `json:"actions"`
	Automation AutomationInfo      `json:"automation"`

	Priority      string        `json:"priority"`
	EstimatedTime time.Duration `json:"estimated_time"`
	RequiredSkill string        `json:"required_skill"`
}

// MaintenanceSchedule represents when maintenance should be performed
type MaintenanceSchedule struct {
	Frequency     string    `json:"frequency"` // daily, weekly, monthly, quarterly, yearly
	PreferredTime string    `json:"preferred_time"`
	Dependencies  []string  `json:"dependencies"`
	Blackouts     []string  `json:"blackouts"` // Times to avoid
	NextDue       time.Time `json:"next_due"`
}

// AutomationInfo represents automation possibilities
type AutomationInfo struct {
	CanAutomate     bool         `json:"can_automate"`
	AutomationLevel string       `json:"automation_level"` // full, partial, none
	Requirements    []string     `json:"requirements"`
	SetupSteps      []ActionStep `json:"setup_steps"`
	MonitoringNeeds []string     `json:"monitoring_needs"`
}

// ConfigurationRecommendation represents configuration improvement recommendations
type ConfigurationRecommendation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ConfigScope string `json:"config_scope"` // system, user, service, package

	CurrentConfig     ConfigChange `json:"current_config"`
	RecommendedConfig ConfigChange `json:"recommended_config"`
	ConfigDiff        string       `json:"config_diff"`

	Actions         []ActionStep `json:"actions"`
	ValidationSteps []string     `json:"validation_steps"`

	Priority       string  `json:"priority"`
	Confidence     float64 `json:"confidence"`
	RequiresReboot bool    `json:"requires_reboot"`
}

// ConfigChange represents a configuration change
type ConfigChange struct {
	FilePath string      `json:"file_path"`
	Section  string      `json:"section"`
	Option   string      `json:"option"`
	Value    interface{} `json:"value"`
	Format   string      `json:"format"` // nix, json, yaml, ini
}

// UserRecommendation represents user-specific recommendations
type UserRecommendation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	UserContext string `json:"user_context"` // developer, administrator, desktop_user

	LearningPath    []LearningStep   `json:"learning_path"`
	ToolSuggestions []ToolSuggestion `json:"tool_suggestions"`
	WorkflowTips    []string         `json:"workflow_tips"`

	Priority   string  `json:"priority"`
	Confidence float64 `json:"confidence"`
	SkillLevel string  `json:"skill_level"` // beginner, intermediate, advanced
}

// LearningStep represents a learning opportunity
type LearningStep struct {
	Topic          string        `json:"topic"`
	Description    string        `json:"description"`
	Resources      []string      `json:"resources"`
	PracticalTasks []string      `json:"practical_tasks"`
	EstimatedTime  time.Duration `json:"estimated_time"`
}

// ToolSuggestion represents a suggested tool or package
type ToolSuggestion struct {
	ToolName       string   `json:"tool_name"`
	Purpose        string   `json:"purpose"`
	InstallCommand string   `json:"install_command"`
	ConfigTips     []string `json:"config_tips"`
	Alternatives   []string `json:"alternatives"`
}

// WorkflowRecommendation represents workflow optimization recommendations
type WorkflowRecommendation struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	WorkflowType string `json:"workflow_type"` // development, administration, backup, deployment

	CurrentWorkflow   []WorkflowStep `json:"current_workflow"`
	OptimizedWorkflow []WorkflowStep `json:"optimized_workflow"`
	ImprovementAreas  []string       `json:"improvement_areas"`

	Priority        string  `json:"priority"`
	Confidence      float64 `json:"confidence"`
	ExpectedBenefit string  `json:"expected_benefit"`
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	StepName    string        `json:"step_name"`
	Description string        `json:"description"`
	Tools       []string      `json:"tools"`
	Duration    time.Duration `json:"duration"`
	Automatable bool          `json:"automatable"`
}

// EnvironmentRecommendation represents environment-specific recommendations
type EnvironmentRecommendation struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	EnvironmentType string `json:"environment_type"` // development, staging, production, desktop

	EnvironmentOptimizations []EnvironmentOptimization `json:"environment_optimizations"`
	BestPractices            []string                  `json:"best_practices"`
	ComplianceRequirements   []string                  `json:"compliance_requirements"`

	Priority            string   `json:"priority"`
	Confidence          float64  `json:"confidence"`
	ApplicableScenarios []string `json:"applicable_scenarios"`
}

// EnvironmentOptimization represents an environment-specific optimization
type EnvironmentOptimization struct {
	Area             string       `json:"area"`
	CurrentState     string       `json:"current_state"`
	RecommendedState string       `json:"recommended_state"`
	Actions          []ActionStep `json:"actions"`
	Benefits         []string     `json:"benefits"`
}

// AIReasoningStep represents AI reasoning in recommendation generation
type AIReasoningStep struct {
	Step       int      `json:"step"`
	Reasoning  string   `json:"reasoning"`
	Evidence   []string `json:"evidence"`
	Confidence float64  `json:"confidence"`
	Sources    []string `json:"sources"`
}

// NewRecommendationsEngine creates a new intelligent recommendations engine
func NewRecommendationsEngine(
	log *logger.Logger,
	aiProvider ai.Provider,
	systemAnalyzer *SystemAnalyzer,
	predictor *Predictor,
	conflictDetector *ConflictDetector,
	dependencyAnalyzer *DependencyAnalyzer,
) *RecommendationsEngine {
	return &RecommendationsEngine{
		logger:             log,
		aiProvider:         aiProvider,
		systemAnalyzer:     systemAnalyzer,
		predictor:          predictor,
		conflictDetector:   conflictDetector,
		dependencyAnalyzer: dependencyAnalyzer,
		cache:              make(map[string]*RecommendationSet),
	}
}

// GenerateRecommendations creates comprehensive recommendations based on all available intelligence
func (re *RecommendationsEngine) GenerateRecommendations(ctx context.Context, userConfig *config.UserConfig) (*RecommendationSet, error) {
	re.logger.Info("Generating comprehensive recommendations")
	startTime := time.Now()

	// Check cache first
	cacheKey := re.generateCacheKey(userConfig)
	if cached := re.getCachedRecommendations(cacheKey); cached != nil {
		re.logger.Info("Returning cached recommendations")
		return cached, nil
	}

	// Gather all intelligence data in parallel
	var wg sync.WaitGroup
	var (
		systemAnalysis     *SystemAnalysis
		predictions        *PredictionResult
		conflictAnalysis   *ConflictAnalysis
		dependencyAnalysis *DependencyAnalysis
		errors             = make(chan error, 4)
	)

	// System analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		analysis, err := re.systemAnalyzer.AnalyzeSystem(ctx, userConfig)
		if err != nil {
			errors <- fmt.Errorf("system analysis: %w", err)
			return
		}
		systemAnalysis = analysis
	}()

	// Predictive analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		pred, err := re.predictor.GeneratePredictions(ctx, userConfig)
		if err != nil {
			errors <- fmt.Errorf("predictions: %w", err)
			return
		}
		predictions = pred
	}()

	// Conflict detection
	wg.Add(1)
	go func() {
		defer wg.Done()
		conflicts, err := re.conflictDetector.DetectConflicts(ctx, userConfig)
		if err != nil {
			errors <- fmt.Errorf("conflict detection: %w", err)
			return
		}
		conflictAnalysis = conflicts
	}()

	// Dependency analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		deps, err := re.dependencyAnalyzer.AnalyzeDependencies(ctx, userConfig)
		if err != nil {
			errors <- fmt.Errorf("dependency analysis: %w", err)
			return
		}
		dependencyAnalysis = deps
	}()

	// Wait for all analyses
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		re.logger.Warn(fmt.Sprintf("Intelligence gathering warning: %v", err))
	}

	// Create intelligence context for AI-powered recommendations
	intelligenceContext := &IntelligenceContext{
		SystemAnalysis:     systemAnalysis,
		Predictions:        predictions,
		ConflictAnalysis:   conflictAnalysis,
		DependencyAnalysis: dependencyAnalysis,
		UserConfig:         userConfig,
	}

	// Generate recommendations
	recommendations := &RecommendationSet{
		GeneratedAt:       startTime,
		SystemSnapshot:    fmt.Sprintf("%s_%s", systemAnalysis.SystemType, systemAnalysis.Hostname),
		PriorityBreakdown: make(map[string]int),
		CategoryBreakdown: make(map[string]int),
	}

	// Generate different types of recommendations
	re.generateSystemOptimizations(intelligenceContext, recommendations)
	re.generateSecurityRecommendations(intelligenceContext, recommendations)
	re.generatePerformanceRecommendations(intelligenceContext, recommendations)
	re.generateMaintenanceRecommendations(intelligenceContext, recommendations)
	re.generateConfigurationRecommendations(intelligenceContext, recommendations)
	re.generateUserSpecificRecommendations(intelligenceContext, recommendations)
	re.generateWorkflowRecommendations(intelligenceContext, recommendations)
	re.generateEnvironmentRecommendations(intelligenceContext, recommendations)

	// Use AI to enhance recommendations
	re.enhanceWithAIRecommendations(ctx, intelligenceContext, recommendations)

	// Calculate final metrics and metadata
	re.calculateRecommendationMetrics(recommendations)

	// Cache the results
	re.cacheRecommendations(cacheKey, recommendations)

	re.logger.Info(fmt.Sprintf("Generated %d recommendations in %v (confidence: %.1f%%)",
		recommendations.TotalRecommendations, time.Since(startTime), recommendations.Confidence*100))

	return recommendations, nil
}

// IntelligenceContext contains all intelligence data for recommendation generation
type IntelligenceContext struct {
	SystemAnalysis     *SystemAnalysis
	Predictions        *PredictionResult
	ConflictAnalysis   *ConflictAnalysis
	DependencyAnalysis *DependencyAnalysis
	UserConfig         *config.UserConfig
}

// generateSystemOptimizations creates system-level optimization recommendations
func (re *RecommendationsEngine) generateSystemOptimizations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var optimizations []SystemRecommendation

	// Disk space optimization from predictions
	if ctx.Predictions != nil {
		for _, maintenance := range ctx.Predictions.MaintenanceSuggestions {
			if strings.Contains(strings.ToLower(maintenance.Task), "cleanup") ||
				strings.Contains(strings.ToLower(maintenance.Task), "garbage") {
				optimizations = append(optimizations, SystemRecommendation{
					ID:          fmt.Sprintf("sys_opt_%d", len(optimizations)+1),
					Title:       "System Cleanup Optimization",
					Description: fmt.Sprintf("Optimize system storage: %s", maintenance.Description),
					Category:    "cleanup",
					Priority:    maintenance.Priority,
					Confidence:  maintenance.Confidence,
					Actions: []ActionStep{
						{
							StepNumber:  1,
							Description: "Run Nix garbage collection",
							Command:     "nix-collect-garbage -d",
							Critical:    false,
							Automated:   true,
						},
						{
							StepNumber:  2,
							Description: "Optimize Nix store",
							Command:     "nix-store --optimize",
							Critical:    false,
							Automated:   true,
						},
					},
					Benefits:      []string{"Free disk space", "Improve performance", "Remove old generations"},
					Risks:         []string{"Cannot rollback removed generations"},
					EstimatedTime: maintenance.EstimatedTime,
					RequiredSkill: "beginner",
					TriggeredBy:   []string{"predictive_analysis"},
					Tags:          []string{"cleanup", "maintenance", "disk_space"},
				})
			}
		}
	}

	// Dependency optimization
	if ctx.DependencyAnalysis != nil && len(ctx.DependencyAnalysis.Graph.OrphanedNodes) > 0 {
		optimizations = append(optimizations, SystemRecommendation{
			ID:          fmt.Sprintf("sys_opt_%d", len(optimizations)+1),
			Title:       "Remove Orphaned Packages",
			Description: fmt.Sprintf("Remove %d orphaned packages to simplify system", len(ctx.DependencyAnalysis.Graph.OrphanedNodes)),
			Category:    "optimization",
			Priority:    "medium",
			Confidence:  0.9,
			Actions: []ActionStep{
				{
					StepNumber:  1,
					Description: "List orphaned packages",
					Command:     "nix-env --query --available",
					Critical:    false,
					Automated:   true,
				},
				{
					StepNumber:  2,
					Description: "Remove orphaned packages",
					Command:     "nix-env --uninstall [package-names]",
					Critical:    false,
					Automated:   false,
				},
			},
			Benefits:      []string{"Reduced complexity", "Less disk usage", "Fewer conflicts"},
			Risks:         []string{"Minimal - packages have no dependents"},
			EstimatedTime: 15 * time.Minute,
			RequiredSkill: "beginner",
			TriggeredBy:   []string{"dependency_analysis"},
			Tags:          []string{"optimization", "cleanup", "orphaned"},
		})
	}

	recommendations.SystemOptimizations = optimizations
}

// generateSecurityRecommendations creates security-focused recommendations
func (re *RecommendationsEngine) generateSecurityRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var securityRecs []SecurityRecommendation

	// Security recommendations from predictions
	if ctx.Predictions != nil {
		for _, security := range ctx.Predictions.SecuritySuggestions {
			securityRec := SecurityRecommendation{
				ID:             fmt.Sprintf("sec_%d", len(securityRecs)+1),
				Title:          security.Title,
				Description:    security.Description,
				Severity:       security.Severity,
				SecurityDomain: security.Category,
				ThreatModel:    []string{security.Category},
				Confidence:     security.Confidence,
				EstimatedTime:  30 * time.Minute, // Default estimate
				RequiredSkill:  "intermediate",
			}

			// Convert implementation steps to actions
			for i, impl := range security.Implementation {
				securityRec.Actions = append(securityRec.Actions, ActionStep{
					StepNumber:  i + 1,
					Description: impl,
					Critical:    security.Severity == "critical" || security.Severity == "high",
					Automated:   false,
				})
			}

			securityRecs = append(securityRecs, securityRec)
		}
	}

	// Security recommendations from dependency analysis
	if ctx.DependencyAnalysis != nil && ctx.DependencyAnalysis.SecurityAnalysis != nil {
		secAnalysis := ctx.DependencyAnalysis.SecurityAnalysis

		if len(secAnalysis.VulnerablePackages) > 0 {
			securityRecs = append(securityRecs, SecurityRecommendation{
				ID:             fmt.Sprintf("sec_%d", len(securityRecs)+1),
				Title:          "Update Vulnerable Packages",
				Description:    fmt.Sprintf("Found %d packages with known vulnerabilities", len(secAnalysis.VulnerablePackages)),
				Severity:       "high",
				SecurityDomain: "packages",
				ThreatModel:    []string{"known_vulnerabilities", "package_security"},
				Actions: []ActionStep{
					{
						StepNumber:  1,
						Description: "Update vulnerable packages",
						Command:     "nixos-rebuild switch --upgrade",
						Critical:    true,
						Automated:   true,
					},
					{
						StepNumber:  2,
						Description: "Verify security updates",
						Command:     "nix-env --query --installed",
						Critical:    false,
						Automated:   true,
					},
				},
				Confidence:    0.95,
				EstimatedTime: 20 * time.Minute,
				RequiredSkill: "beginner",
			})
		}
	}

	recommendations.SecurityRecommendations = securityRecs
}

// generatePerformanceRecommendations creates performance optimization recommendations
func (re *RecommendationsEngine) generatePerformanceRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var perfRecs []PerformanceRecommendation

	// Performance recommendations from predictions
	if ctx.Predictions != nil {
		for _, perf := range ctx.Predictions.PerformanceSuggestions {
			perfRec := PerformanceRecommendation{
				ID:              fmt.Sprintf("perf_%d", len(perfRecs)+1),
				Title:           fmt.Sprintf("Optimize %s Performance", strings.Title(perf.Component)),
				Description:     perf.Issue,
				PerformanceArea: perf.Component,
				Priority:        perf.Priority,
				Confidence:      perf.Confidence,
				EstimatedTime:   20 * time.Minute,
				Reversible:      perf.Reversible,
			}

			// Add expected gains
			perfRec.ExpectedGains = []PerformanceGain{
				{
					Metric:         perf.Component,
					ExpectedChange: perf.ExpectedGain,
					Confidence:     perf.Confidence,
					TimeFrame:      "immediate",
				},
			}

			// Convert implementation steps to actions
			for i, impl := range perf.Implementation {
				perfRec.Actions = append(perfRec.Actions, ActionStep{
					StepNumber:  i + 1,
					Description: impl,
					Critical:    perf.Priority == "high" || perf.Priority == "critical",
					Automated:   strings.Contains(strings.ToLower(impl), "enable") || strings.Contains(strings.ToLower(impl), "set"),
				})
			}

			perfRecs = append(perfRecs, perfRec)
		}
	}

	recommendations.PerformanceRecommendations = perfRecs
}

// generateMaintenanceRecommendations creates maintenance recommendations
func (re *RecommendationsEngine) generateMaintenanceRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var maintRecs []MaintenanceRecommendation

	// Maintenance recommendations from predictions
	if ctx.Predictions != nil {
		for _, maint := range ctx.Predictions.MaintenanceSuggestions {
			maintRec := MaintenanceRecommendation{
				ID:              fmt.Sprintf("maint_%d", len(maintRecs)+1),
				Title:           maint.Task,
				Description:     maint.Description,
				MaintenanceType: "preventive",
				Priority:        maint.Priority,
				EstimatedTime:   maint.EstimatedTime,
				RequiredSkill:   "beginner",
				Schedule: MaintenanceSchedule{
					Frequency: maint.Frequency,
					NextDue:   maint.NextDue,
				},
				Automation: AutomationInfo{
					CanAutomate:     strings.Contains(maint.Automation, "automat"),
					AutomationLevel: "partial",
					Requirements:    []string{"systemd timers", "cron jobs"},
				},
			}

			// Convert commands to actions
			for i, cmd := range maint.Commands {
				maintRec.Actions = append(maintRec.Actions, ActionStep{
					StepNumber:  i + 1,
					Description: fmt.Sprintf("Execute maintenance command: %s", cmd),
					Command:     cmd,
					Critical:    false,
					Automated:   true,
				})
			}

			maintRecs = append(maintRecs, maintRec)
		}
	}

	recommendations.MaintenanceRecommendations = maintRecs
}

// generateConfigurationRecommendations creates configuration improvement recommendations
func (re *RecommendationsEngine) generateConfigurationRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var configRecs []ConfigurationRecommendation

	// Configuration recommendations from predictions
	if ctx.Predictions != nil {
		for _, config := range ctx.Predictions.ConfigSuggestions {
			configRec := ConfigurationRecommendation{
				ID:          fmt.Sprintf("config_%d", len(configRecs)+1),
				Title:       fmt.Sprintf("Configure %s", config.ModuleName),
				Description: config.Reason,
				ConfigScope: "system",
				Priority:    config.Priority,
				Confidence:  config.Confidence,
				RequiresReboot: strings.Contains(strings.ToLower(config.ModuleName), "boot") ||
					strings.Contains(strings.ToLower(config.ModuleName), "kernel"),
				CurrentConfig: ConfigChange{
					FilePath: config.ConfigPath,
					Section:  config.ModuleName,
					Option:   config.Option,
					Value:    config.CurrentValue,
					Format:   "nix",
				},
				RecommendedConfig: ConfigChange{
					FilePath: config.ConfigPath,
					Section:  config.ModuleName,
					Option:   config.Option,
					Value:    config.SuggestedValue,
					Format:   "nix",
				},
				ConfigDiff: config.Example,
				Actions: []ActionStep{
					{
						StepNumber:  1,
						Description: fmt.Sprintf("Update %s configuration", config.ModuleName),
						Command:     fmt.Sprintf("# Add to configuration.nix: %s", config.Example),
						Critical:    config.Priority == "high",
						Automated:   false,
					},
					{
						StepNumber:  2,
						Description: "Rebuild system configuration",
						Command:     "nixos-rebuild switch",
						Critical:    true,
						Automated:   true,
					},
				},
			}

			configRecs = append(configRecs, configRec)
		}
	}

	recommendations.ConfigurationRecommendations = configRecs
}

// generateUserSpecificRecommendations creates user-tailored recommendations
func (re *RecommendationsEngine) generateUserSpecificRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var userRecs []UserRecommendation

	// Determine user context from system analysis
	userContext := re.determineUserContext(ctx.SystemAnalysis)

	// Package suggestions based on user context
	if ctx.Predictions != nil {
		for _, pkg := range ctx.Predictions.PackageSuggestions {
			userRec := UserRecommendation{
				ID:          fmt.Sprintf("user_%d", len(userRecs)+1),
				Title:       fmt.Sprintf("Install %s for %s", pkg.PackageName, pkg.Category),
				Description: pkg.Reason,
				UserContext: userContext,
				Priority:    pkg.Priority,
				Confidence:  pkg.Confidence,
				SkillLevel:  "beginner",
				ToolSuggestions: []ToolSuggestion{
					{
						ToolName:       pkg.PackageName,
						Purpose:        pkg.Reason,
						InstallCommand: pkg.InstallCommand,
						Alternatives:   pkg.AlternativePackages,
					},
				},
			}

			if len(pkg.Benefits) > 0 {
				userRec.WorkflowTips = pkg.Benefits
			}

			userRecs = append(userRecs, userRec)
		}
	}

	recommendations.UserSpecificRecommendations = userRecs
}

// generateWorkflowRecommendations creates workflow optimization recommendations
func (re *RecommendationsEngine) generateWorkflowRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var workflowRecs []WorkflowRecommendation

	// Analyze common workflows based on installed packages and services
	if ctx.SystemAnalysis != nil {
		workflows := re.identifyWorkflows(ctx.SystemAnalysis)

		for _, workflow := range workflows {
			workflowRec := WorkflowRecommendation{
				ID:              fmt.Sprintf("workflow_%d", len(workflowRecs)+1),
				Title:           fmt.Sprintf("Optimize %s Workflow", workflow),
				Description:     fmt.Sprintf("Suggestions to improve your %s workflow", workflow),
				WorkflowType:    workflow,
				Priority:        "medium",
				Confidence:      0.7,
				ExpectedBenefit: "Improved productivity and efficiency",
			}

			workflowRecs = append(workflowRecs, workflowRec)
		}
	}

	recommendations.WorkflowRecommendations = workflowRecs
}

// generateEnvironmentRecommendations creates environment-specific recommendations
func (re *RecommendationsEngine) generateEnvironmentRecommendations(ctx *IntelligenceContext, recommendations *RecommendationSet) {
	var envRecs []EnvironmentRecommendation

	// Determine environment type
	envType := re.determineEnvironmentType(ctx.SystemAnalysis)

	envRec := EnvironmentRecommendation{
		ID:              fmt.Sprintf("env_%d", len(envRecs)+1),
		Title:           fmt.Sprintf("Optimize %s Environment", strings.Title(envType)),
		Description:     fmt.Sprintf("Recommendations specific to %s environment setup", envType),
		EnvironmentType: envType,
		Priority:        "medium",
		Confidence:      0.8,
		BestPractices: []string{
			"Regular system updates",
			"Backup configuration files",
			"Monitor system resources",
			"Use version control for configurations",
		},
	}

	envRecs = append(envRecs, envRec)
	recommendations.EnvironmentRecommendations = envRecs
}

// enhanceWithAIRecommendations uses AI to generate additional insights
func (re *RecommendationsEngine) enhanceWithAIRecommendations(ctx context.Context, intelligenceCtx *IntelligenceContext, recommendations *RecommendationSet) {
	if re.aiProvider == nil {
		return
	}

	// Create AI prompt with intelligence context
	prompt := re.buildAIRecommendationPrompt(intelligenceCtx)

	response, err := re.aiProvider.GenerateResponse(ctx, prompt)
	if err != nil {
		re.logger.Warn(fmt.Sprintf("AI enhancement failed: %v", err))
		return
	}

	// Parse AI response and add reasoning chain
	aiReasoning := re.parseAIRecommendationResponse(response)
	recommendations.AIReasoningChain = aiReasoning

	re.logger.Info("Enhanced recommendations with AI insights")
}

// Helper methods

func (re *RecommendationsEngine) determineUserContext(systemAnalysis *SystemAnalysis) string {
	if systemAnalysis == nil {
		return "general"
	}

	// Simple heuristic based on installed packages
	packageNames := make([]string, len(systemAnalysis.InstalledPackages))
	for i, pkg := range systemAnalysis.InstalledPackages {
		packageNames[i] = strings.ToLower(pkg.Name)
	}

	devTools := []string{"git", "gcc", "python", "nodejs", "vim", "code", "make", "cargo"}
	adminTools := []string{"systemctl", "docker", "kubernetes", "nginx", "apache"}
	desktopTools := []string{"firefox", "chrome", "libreoffice", "gimp", "vlc"}

	devCount := 0
	adminCount := 0
	desktopCount := 0

	for _, pkg := range packageNames {
		for _, tool := range devTools {
			if strings.Contains(pkg, tool) {
				devCount++
				break
			}
		}
		for _, tool := range adminTools {
			if strings.Contains(pkg, tool) {
				adminCount++
				break
			}
		}
		for _, tool := range desktopTools {
			if strings.Contains(pkg, tool) {
				desktopCount++
				break
			}
		}
	}

	if devCount > adminCount && devCount > desktopCount {
		return "developer"
	}
	if adminCount > devCount && adminCount > desktopCount {
		return "administrator"
	}
	if desktopCount > 0 {
		return "desktop_user"
	}

	return "general"
}

func (re *RecommendationsEngine) identifyWorkflows(systemAnalysis *SystemAnalysis) []string {
	var workflows []string

	if systemAnalysis == nil {
		return workflows
	}

	// Check for development workflow
	devPackages := []string{"git", "gcc", "python", "nodejs", "make"}
	hasDevPackages := false
	for _, pkg := range systemAnalysis.InstalledPackages {
		for _, devPkg := range devPackages {
			if strings.Contains(strings.ToLower(pkg.Name), devPkg) {
				hasDevPackages = true
				break
			}
		}
		if hasDevPackages {
			break
		}
	}
	if hasDevPackages {
		workflows = append(workflows, "development")
	}

	// Check for system administration workflow
	adminServices := []string{"nginx", "apache", "docker", "postgresql", "mysql"}
	hasAdminServices := false
	for _, service := range systemAnalysis.EnabledServices {
		for _, adminSvc := range adminServices {
			if strings.Contains(strings.ToLower(service.Name), adminSvc) {
				hasAdminServices = true
				break
			}
		}
		if hasAdminServices {
			break
		}
	}
	if hasAdminServices {
		workflows = append(workflows, "administration")
	}

	return workflows
}

func (re *RecommendationsEngine) determineEnvironmentType(systemAnalysis *SystemAnalysis) string {
	if systemAnalysis == nil {
		return "desktop"
	}

	// Check for server indicators
	serverServices := []string{"sshd", "nginx", "apache", "docker", "kubernetes"}
	for _, service := range systemAnalysis.EnabledServices {
		for _, serverSvc := range serverServices {
			if strings.Contains(strings.ToLower(service.Name), serverSvc) {
				return "production"
			}
		}
	}

	// Check for desktop indicators
	desktopServices := []string{"display-manager", "xserver", "wayland"}
	for _, service := range systemAnalysis.EnabledServices {
		for _, desktopSvc := range desktopServices {
			if strings.Contains(strings.ToLower(service.Name), desktopSvc) {
				return "desktop"
			}
		}
	}

	return "development"
}

func (re *RecommendationsEngine) buildAIRecommendationPrompt(ctx *IntelligenceContext) string {
	return fmt.Sprintf(`
Analyze the following NixOS system intelligence data and provide additional recommendations:

System Analysis:
- System Type: %s
- Installed Packages: %d
- Enabled Services: %d
- Security Score: %.1f

Detected Issues:
- Conflicts: %d
- Circular Dependencies: %d
- Vulnerable Packages: %d

Current Recommendations: %d

Please provide:
1. Additional optimization opportunities
2. Potential risks or concerns
3. Long-term maintenance strategies
4. Integration improvements

Focus on practical, actionable advice for NixOS users.
`,
		ctx.SystemAnalysis.SystemType,
		len(ctx.SystemAnalysis.InstalledPackages),
		len(ctx.SystemAnalysis.EnabledServices),
		ctx.SystemAnalysis.SecuritySettings.SecurityScore,
		ctx.ConflictAnalysis.TotalConflicts,
		len(ctx.DependencyAnalysis.Graph.CircularDeps),
		len(ctx.DependencyAnalysis.SecurityAnalysis.VulnerablePackages),
		len(ctx.Predictions.PackageSuggestions)+len(ctx.Predictions.ConfigSuggestions))
}

func (re *RecommendationsEngine) parseAIRecommendationResponse(response string) []AIReasoningStep {
	// Simple parsing - in reality, this would be more sophisticated
	return []AIReasoningStep{
		{
			Step:       1,
			Reasoning:  "AI analyzed system state and provided additional insights",
			Evidence:   []string{"System analysis", "Pattern recognition", "Best practices"},
			Confidence: 0.8,
			Sources:    []string{"AI model", "NixOS best practices"},
		},
	}
}

func (re *RecommendationsEngine) calculateRecommendationMetrics(recommendations *RecommendationSet) {
	// Count total recommendations
	recommendations.TotalRecommendations = len(recommendations.SystemOptimizations) +
		len(recommendations.SecurityRecommendations) +
		len(recommendations.PerformanceRecommendations) +
		len(recommendations.MaintenanceRecommendations) +
		len(recommendations.ConfigurationRecommendations) +
		len(recommendations.UserSpecificRecommendations) +
		len(recommendations.WorkflowRecommendations) +
		len(recommendations.EnvironmentRecommendations)

	// Calculate priority breakdown
	priorities := []string{"critical", "high", "medium", "low"}
	for _, priority := range priorities {
		recommendations.PriorityBreakdown[priority] = 0
	}

	// Count priorities from all recommendation types
	for _, rec := range recommendations.SystemOptimizations {
		recommendations.PriorityBreakdown[rec.Priority]++
	}
	for _, rec := range recommendations.SecurityRecommendations {
		if rec.Severity != "" {
			recommendations.PriorityBreakdown[rec.Severity]++
		}
	}
	for _, rec := range recommendations.PerformanceRecommendations {
		recommendations.PriorityBreakdown[rec.Priority]++
	}

	// Calculate category breakdown
	categories := []string{"optimization", "security", "performance", "maintenance", "configuration", "user", "workflow", "environment"}
	for _, category := range categories {
		recommendations.CategoryBreakdown[category] = 0
	}

	recommendations.CategoryBreakdown["optimization"] = len(recommendations.SystemOptimizations)
	recommendations.CategoryBreakdown["security"] = len(recommendations.SecurityRecommendations)
	recommendations.CategoryBreakdown["performance"] = len(recommendations.PerformanceRecommendations)
	recommendations.CategoryBreakdown["maintenance"] = len(recommendations.MaintenanceRecommendations)
	recommendations.CategoryBreakdown["configuration"] = len(recommendations.ConfigurationRecommendations)
	recommendations.CategoryBreakdown["user"] = len(recommendations.UserSpecificRecommendations)
	recommendations.CategoryBreakdown["workflow"] = len(recommendations.WorkflowRecommendations)
	recommendations.CategoryBreakdown["environment"] = len(recommendations.EnvironmentRecommendations)

	// Calculate overall confidence
	totalConfidence := 0.0
	confidenceCount := 0

	for _, rec := range recommendations.SystemOptimizations {
		totalConfidence += rec.Confidence
		confidenceCount++
	}
	for _, rec := range recommendations.SecurityRecommendations {
		totalConfidence += rec.Confidence
		confidenceCount++
	}

	if confidenceCount > 0 {
		recommendations.Confidence = totalConfidence / float64(confidenceCount)
	} else {
		recommendations.Confidence = 0.5
	}

	// Estimate total implementation time
	totalTime := time.Duration(0)
	for _, rec := range recommendations.SystemOptimizations {
		totalTime += rec.EstimatedTime
	}
	for _, rec := range recommendations.SecurityRecommendations {
		totalTime += rec.EstimatedTime
	}
	for _, rec := range recommendations.PerformanceRecommendations {
		totalTime += rec.EstimatedTime
	}
	recommendations.TimeToImplement = totalTime

	// Set estimated impact
	criticalCount := recommendations.PriorityBreakdown["critical"]
	highCount := recommendations.PriorityBreakdown["high"]

	switch {
	case criticalCount > 0:
		recommendations.EstimatedImpact = "High - Critical issues found"
	case highCount > 3:
		recommendations.EstimatedImpact = "Medium-High - Multiple important improvements"
	case recommendations.TotalRecommendations > 10:
		recommendations.EstimatedImpact = "Medium - Various optimizations available"
	default:
		recommendations.EstimatedImpact = "Low-Medium - System is well-configured"
	}

	// Set analysis sources
	recommendations.BasedOnAnalysis = []string{
		"System Analysis", "Predictive Intelligence", "Conflict Detection",
		"Dependency Analysis", "AI Enhancement",
	}
}

func (re *RecommendationsEngine) generateCacheKey(userConfig *config.UserConfig) string {
	return fmt.Sprintf("recommendations_%s_%d", userConfig.AIProvider, time.Now().Unix()/3600) // Cache for 1 hour
}

func (re *RecommendationsEngine) getCachedRecommendations(key string) *RecommendationSet {
	re.mu.RLock()
	defer re.mu.RUnlock()

	return re.cache[key]
}

func (re *RecommendationsEngine) cacheRecommendations(key string, recommendations *RecommendationSet) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.cache[key] = recommendations

	// Limit cache size
	if len(re.cache) > 10 {
		oldestKey := ""
		oldestTime := time.Now()
		for k, v := range re.cache {
			if v.GeneratedAt.Before(oldestTime) {
				oldestTime = v.GeneratedAt
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(re.cache, oldestKey)
		}
	}
}

// ClearCache clears the recommendations cache
func (re *RecommendationsEngine) ClearCache() {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.cache = make(map[string]*RecommendationSet)
	re.logger.Info("Recommendations cache cleared")
}
