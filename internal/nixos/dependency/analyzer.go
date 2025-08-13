// Package dependency provides NixOS configuration dependency analysis
package dependency

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/hardware"
	"nix-ai-help/pkg/logger"
)

// DependencyAnalyzer analyzes NixOS configuration dependencies
type DependencyAnalyzer struct {
	logger           *logger.Logger
	hardwareDetector *hardware.EnhancedHardwareDetector
	ruleEngine       *RuleEngine
	cache            *AnalysisCache
}

// AnalysisCache stores dependency analysis results
type AnalysisCache struct {
	results   map[string]*AnalysisResult
	timestamp time.Time
	ttl       time.Duration
}

// DependencyAnalysis represents a complete dependency analysis
type DependencyAnalysis struct {
	ConfigurationOptions []*ConfigOption       `json:"configuration_options"`
	Dependencies         []*Dependency         `json:"dependencies"`
	Conflicts            []*Conflict           `json:"conflicts"`
	Recommendations      []*Recommendation     `json:"recommendations"`
	DependencyGraph      *DependencyGraph      `json:"dependency_graph"`
	HardwareDependencies []*HardwareDependency `json:"hardware_dependencies"`
	ValidationResults    *ValidationResults    `json:"validation_results"`
	AnalysisMetadata     *AnalysisMetadata     `json:"analysis_metadata"`
}

// ConfigOption represents a NixOS configuration option
type ConfigOption struct {
	Name          string                 `json:"name"`
	Value         interface{}            `json:"value"`
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Module        string                 `json:"module"`
	Required      bool                   `json:"required"`
	DefaultValue  interface{}            `json:"default_value,omitempty"`
	ValidValues   []interface{}          `json:"valid_values,omitempty"`
	Dependencies  []string               `json:"dependencies,omitempty"`
	Conflicts     []string               `json:"conflicts,omitempty"`
	HardwareReqs  []string               `json:"hardware_requirements,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// Dependency represents a relationship between configuration options
type Dependency struct {
	From          string         `json:"from"`
	To            string         `json:"to"`
	Type          DependencyType `json:"type"`
	Condition     string         `json:"condition,omitempty"`
	Strength      float64        `json:"strength"` // 0.0 to 1.0
	Description   string         `json:"description"`
	AutoResolve   bool           `json:"auto_resolve"`
	Resolution    string         `json:"resolution,omitempty"`
}

// DependencyType represents the type of dependency relationship
type DependencyType string

const (
	DependencyRequired    DependencyType = "required"
	DependencyRecommended DependencyType = "recommended"
	DependencyOptional    DependencyType = "optional"
	DependencyImplies     DependencyType = "implies"
	DependencyMutex       DependencyType = "mutex"
	DependencyHardware    DependencyType = "hardware"
)

// Conflict represents a configuration conflict
type Conflict struct {
	Options     []string      `json:"options"`
	Type        ConflictType  `json:"type"`
	Severity    string        `json:"severity"` // critical, warning, info
	Description string        `json:"description"`
	Resolution  []string      `json:"resolution"`
	AutoFix     bool          `json:"auto_fix"`
	FixActions  []string      `json:"fix_actions,omitempty"`
}

// ConflictType represents the type of configuration conflict
type ConflictType string

const (
	ConflictMutualExclusion ConflictType = "mutual_exclusion"
	ConflictVersionMismatch ConflictType = "version_mismatch"
	ConflictResourceLimit   ConflictType = "resource_limit"
	ConflictHardwareLimit   ConflictType = "hardware_limit"
	ConflictServiceConflict ConflictType = "service_conflict"
	ConflictModuleConflict  ConflictType = "module_conflict"
)

// Recommendation represents a configuration recommendation
type Recommendation struct {
	Type        RecommendationType `json:"type"`
	Priority    int                `json:"priority"` // 1-10, higher is more important
	Option      string             `json:"option"`
	Action      string             `json:"action"` // add, remove, modify
	Value       interface{}        `json:"value,omitempty"`
	Reason      string             `json:"reason"`
	Benefits    []string           `json:"benefits"`
	Risks       []string           `json:"risks,omitempty"`
	HardwareBased bool             `json:"hardware_based"`
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationPerformance   RecommendationType = "performance"
	RecommendationSecurity      RecommendationType = "security"
	RecommendationCompatibility RecommendationType = "compatibility"
	RecommendationOptimization  RecommendationType = "optimization"
	RecommendationHardware      RecommendationType = "hardware"
	RecommendationMaintenance   RecommendationType = "maintenance"
)

// DependencyGraph represents the dependency relationship graph
type DependencyGraph struct {
	Nodes []*GraphNode `json:"nodes"`
	Edges []*GraphEdge `json:"edges"`
	Cycles [][]string  `json:"cycles,omitempty"`
}

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Level       int                    `json:"level"` // Dependency depth level
	Category    string                 `json:"category"`
	Required    bool                   `json:"required"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

// GraphEdge represents an edge in the dependency graph
type GraphEdge struct {
	From     string         `json:"from"`
	To       string         `json:"to"`
	Type     DependencyType `json:"type"`
	Weight   float64        `json:"weight"`
	Label    string         `json:"label,omitempty"`
}

// HardwareDependency represents hardware-specific dependencies
type HardwareDependency struct {
	Option           string   `json:"option"`
	HardwareType     string   `json:"hardware_type"`
	HardwareVendor   string   `json:"hardware_vendor,omitempty"`
	HardwareModel    string   `json:"hardware_model,omitempty"`
	RequiredDrivers  []string `json:"required_drivers"`
	RequiredFirmware []string `json:"required_firmware"`
	RequiredModules  []string `json:"required_modules"`
	OptionalPackages []string `json:"optional_packages,omitempty"`
	ConfigSnippet    string   `json:"config_snippet,omitempty"`
	Detected         bool     `json:"detected"`
	Compatible       bool     `json:"compatible"`
	Notes            string   `json:"notes,omitempty"`
}

// ValidationResults represents configuration validation results
type ValidationResults struct {
	Valid           bool                `json:"valid"`
	Errors          []*ValidationError  `json:"errors,omitempty"`
	Warnings        []*ValidationError  `json:"warnings,omitempty"`
	Suggestions     []string            `json:"suggestions,omitempty"`
	Score           float64             `json:"score"` // 0.0 to 1.0
	OptimizationTips []string           `json:"optimization_tips,omitempty"`
}

// ValidationError represents a validation error or warning
type ValidationError struct {
	Type        string `json:"type"`
	Option      string `json:"option"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
	Resolution  string `json:"resolution,omitempty"`
	Line        int    `json:"line,omitempty"`
	Column      int    `json:"column,omitempty"`
}

// AnalysisMetadata contains metadata about the analysis process
type AnalysisMetadata struct {
	AnalysisTime      time.Time     `json:"analysis_time"`
	AnalysisDuration  time.Duration `json:"analysis_duration"`
	OptionsAnalyzed   int           `json:"options_analyzed"`
	DependenciesFound int           `json:"dependencies_found"`
	ConflictsFound    int           `json:"conflicts_found"`
	HardwareDetected  bool          `json:"hardware_detected"`
	CacheHit          bool          `json:"cache_hit"`
	AnalyzerVersion   string        `json:"analyzer_version"`
}

// AnalysisResult represents a cached analysis result
type AnalysisResult struct {
	Analysis  *DependencyAnalysis `json:"analysis"`
	Timestamp time.Time           `json:"timestamp"`
	Hash      string              `json:"hash"`
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(logger *logger.Logger) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		logger:           logger,
		hardwareDetector: hardware.NewEnhancedHardwareDetector(logger),
		ruleEngine:       NewRuleEngine(),
		cache:            NewAnalysisCache(10 * time.Minute),
	}
}

// NewAnalysisCache creates a new analysis cache
func NewAnalysisCache(ttl time.Duration) *AnalysisCache {
	return &AnalysisCache{
		results: make(map[string]*AnalysisResult),
		ttl:     ttl,
	}
}

// AnalyzeConfiguration performs comprehensive dependency analysis
func (da *DependencyAnalyzer) AnalyzeConfiguration(ctx context.Context, configContent string, options *AnalysisOptions) (*DependencyAnalysis, error) {
	startTime := time.Now()
	da.logger.Info("Starting configuration dependency analysis")

	// Check cache first
	configHash := da.calculateConfigHash(configContent)
	if cachedResult := da.getCachedResult(configHash); cachedResult != nil {
		da.logger.Info("Using cached analysis result")
		cachedResult.Analysis.AnalysisMetadata.CacheHit = true
		return cachedResult.Analysis, nil
	}

	// Parse configuration options
	configOptions, err := da.parseConfigurationOptions(configContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %v", err)
	}

	// Detect hardware if enabled
	var hardwareInfo *hardware.EnhancedHardwareInfo
	if options != nil && options.IncludeHardwareAnalysis {
		hardwareInfo, err = da.hardwareDetector.DetectEnhancedHardware(ctx)
		if err != nil {
			da.logger.Warn(fmt.Sprintf("Hardware detection failed: %v", err))
		}
	}

	// Analyze dependencies
	dependencies := da.analyzeDependencies(configOptions)

	// Detect conflicts
	conflicts := da.detectConflicts(configOptions, dependencies)

	// Generate recommendations
	recommendations := da.generateRecommendations(configOptions, dependencies, conflicts, hardwareInfo)

	// Build dependency graph
	dependencyGraph := da.buildDependencyGraph(configOptions, dependencies)

	// Analyze hardware dependencies
	hardwareDependencies := da.analyzeHardwareDependencies(configOptions, hardwareInfo)

	// Validate configuration
	validationResults := da.validateConfiguration(configOptions, dependencies, conflicts)

	// Create analysis result
	analysis := &DependencyAnalysis{
		ConfigurationOptions: configOptions,
		Dependencies:         dependencies,
		Conflicts:            conflicts,
		Recommendations:      recommendations,
		DependencyGraph:      dependencyGraph,
		HardwareDependencies: hardwareDependencies,
		ValidationResults:    validationResults,
		AnalysisMetadata: &AnalysisMetadata{
			AnalysisTime:      startTime,
			AnalysisDuration:  time.Since(startTime),
			OptionsAnalyzed:   len(configOptions),
			DependenciesFound: len(dependencies),
			ConflictsFound:    len(conflicts),
			HardwareDetected:  hardwareInfo != nil,
			CacheHit:          false,
			AnalyzerVersion:   "1.0.0",
		},
	}

	// Cache result
	da.cacheResult(configHash, analysis)

	da.logger.Info(fmt.Sprintf("Configuration analysis completed: %d options, %d dependencies, %d conflicts", 
		len(configOptions), len(dependencies), len(conflicts)))

	return analysis, nil
}

// AnalysisOptions configures the analysis behavior
type AnalysisOptions struct {
	IncludeHardwareAnalysis bool     `json:"include_hardware_analysis"`
	AnalysisLevel          string   `json:"analysis_level"` // basic, standard, comprehensive
	FocusAreas             []string `json:"focus_areas,omitempty"`
	IgnoreWarnings         bool     `json:"ignore_warnings"`
	GenerateOptimizations  bool     `json:"generate_optimizations"`
}

// Helper methods for dependency analysis will be implemented in separate files
func (da *DependencyAnalyzer) calculateConfigHash(content string) string {
	// Simple hash calculation - could be enhanced with proper hashing
	return fmt.Sprintf("%d", len(content)+strings.Count(content, "\n"))
}

func (da *DependencyAnalyzer) getCachedResult(hash string) *AnalysisResult {
	if result, exists := da.cache.results[hash]; exists {
		if time.Since(result.Timestamp) < da.cache.ttl {
			return result
		}
		// Remove expired cache entry
		delete(da.cache.results, hash)
	}
	return nil
}

func (da *DependencyAnalyzer) cacheResult(hash string, analysis *DependencyAnalysis) {
	da.cache.results[hash] = &AnalysisResult{
		Analysis:  analysis,
		Timestamp: time.Now(),
		Hash:      hash,
	}
}