package dependency

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/nixos/dependency"
	"nix-ai-help/pkg/logger"
)

// DependencyFunction handles NixOS configuration dependency analysis
type DependencyFunction struct {
	*functionbase.BaseFunction
	analyzer *dependency.DependencyAnalyzer
	logger   *logger.Logger
}

// DependencyRequest represents the input parameters for dependency analysis
type DependencyRequest struct {
	ConfigContent           string                         `json:"config_content"`
	AnalysisLevel          string                         `json:"analysis_level,omitempty"`
	IncludeHardwareAnalysis bool                          `json:"include_hardware_analysis,omitempty"`
	FocusAreas             []string                       `json:"focus_areas,omitempty"`
	GenerateOptimizations  bool                          `json:"generate_optimizations,omitempty"`
	OutputFormat           string                         `json:"output_format,omitempty"`
}

// DependencyResponse represents the output of dependency analysis
type DependencyResponse struct {
	Analysis          *dependency.DependencyAnalysis `json:"analysis"`
	Summary           *AnalysisSummary              `json:"summary"`
	ActionableInsights []*ActionableInsight          `json:"actionable_insights"`
	ConfigSuggestions string                         `json:"config_suggestions,omitempty"`
	ExecutionTime     time.Duration                  `json:"execution_time"`
}

// AnalysisSummary provides a high-level summary of the analysis
type AnalysisSummary struct {
	OverallScore      float64 `json:"overall_score"`
	TotalOptions      int     `json:"total_options"`
	Dependencies      int     `json:"dependencies"`
	RequiredDeps      int     `json:"required_dependencies"`
	Conflicts         int     `json:"conflicts"`
	CriticalConflicts int     `json:"critical_conflicts"`
	Recommendations   int     `json:"recommendations"`
	HardwareOptimized bool    `json:"hardware_optimized"`
	Status            string  `json:"status"`
	KeyIssues         []string `json:"key_issues,omitempty"`
}

// ActionableInsight represents a specific actionable recommendation
type ActionableInsight struct {
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Action      string   `json:"action"`
	ConfigLine  string   `json:"config_line,omitempty"`
	Benefits    []string `json:"benefits"`
	Category    string   `json:"category"`
}

// NewDependencyFunction creates a new dependency analysis function
func NewDependencyFunction() *DependencyFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("config_content", "NixOS configuration content to analyze", true),
		functionbase.StringParamWithOptions("analysis_level", "Level of analysis detail", false,
			[]string{"basic", "standard", "comprehensive"}, nil, nil),
		functionbase.BoolParam("include_hardware", "Include hardware-specific analysis", false),
		functionbase.ArrayParam("focus_areas", "Specific areas to focus analysis on", false),
		functionbase.BoolParam("optimizations", "Generate optimization suggestions", false),
	}

	baseFunc := functionbase.NewBaseFunction(
		"dependency-analysis",
		"Analyze NixOS configuration dependencies, conflicts, and optimization opportunities",
		parameters,
	)

	return &DependencyFunction{
		BaseFunction: baseFunc,
		analyzer:     dependency.NewDependencyAnalyzer(logger.NewLogger()),
		logger:       logger.NewLogger(),
	}
}

// Name returns the function name
func (f *DependencyFunction) Name() string {
	return "dependency-analysis"
}

// Description returns the function description
func (f *DependencyFunction) Description() string {
	return "Analyze NixOS configuration dependencies, detect conflicts, and provide optimization recommendations"
}

// Version returns the function version
func (f *DependencyFunction) Version() string {
	return "1.0.0"
}

// Parameters returns the function parameter schema
func (f *DependencyFunction) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"config_content": map[string]interface{}{
				"type":        "string",
				"description": "NixOS configuration content to analyze",
			},
			"analysis_level": map[string]interface{}{
				"type":        "string",
				"description": "Level of analysis detail",
				"enum":        []string{"basic", "standard", "comprehensive"},
				"default":     "standard",
			},
			"include_hardware_analysis": map[string]interface{}{
				"type":        "boolean",
				"description": "Include hardware-specific dependency analysis",
				"default":     true,
			},
			"focus_areas": map[string]interface{}{
				"type":        "array",
				"description": "Specific configuration areas to focus on",
				"items":       map[string]interface{}{"type": "string"},
			},
			"generate_optimizations": map[string]interface{}{
				"type":        "boolean",
				"description": "Generate optimization suggestions",
				"default":     true,
			},
			"output_format": map[string]interface{}{
				"type":        "string",
				"description": "Output format for results",
				"enum":        []string{"structured", "summary", "actionable"},
				"default":     "actionable",
			},
		},
		"required": []string{"config_content"},
	}
}

// Execute runs the dependency analysis function
func (f *DependencyFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()

	// Parse request parameters
	req := &DependencyRequest{
		AnalysisLevel:          "standard",
		IncludeHardwareAnalysis: true,
		GenerateOptimizations:  true,
		OutputFormat:           "actionable",
	}

	if configContent, ok := params["config_content"].(string); ok {
		req.ConfigContent = configContent
	} else {
		return functionbase.ErrorResult(fmt.Errorf("config_content is required"), time.Since(startTime)), nil
	}

	if level, ok := params["analysis_level"].(string); ok {
		req.AnalysisLevel = level
	}
	if includeHW, ok := params["include_hardware_analysis"].(bool); ok {
		req.IncludeHardwareAnalysis = includeHW
	}
	if genOpt, ok := params["generate_optimizations"].(bool); ok {
		req.GenerateOptimizations = genOpt
	}
	if format, ok := params["output_format"].(string); ok {
		req.OutputFormat = format
	}
	if focusAreas, ok := params["focus_areas"].([]interface{}); ok {
		for _, area := range focusAreas {
			if areaStr, ok := area.(string); ok {
				req.FocusAreas = append(req.FocusAreas, areaStr)
			}
		}
	}

	f.logger.Info(fmt.Sprintf("Starting dependency analysis: level=%s, hardware=%t", 
		req.AnalysisLevel, req.IncludeHardwareAnalysis))

	// Perform dependency analysis
	analysisOptions := &dependency.AnalysisOptions{
		IncludeHardwareAnalysis: req.IncludeHardwareAnalysis,
		AnalysisLevel:          req.AnalysisLevel,
		FocusAreas:             req.FocusAreas,
		GenerateOptimizations:  req.GenerateOptimizations,
	}

	analysis, err := f.analyzer.AnalyzeConfiguration(ctx, req.ConfigContent, analysisOptions)
	if err != nil {
		return functionbase.ErrorResult(fmt.Errorf("dependency analysis failed: %v", err), time.Since(startTime)), nil
	}

	// Generate response based on output format
	response := &DependencyResponse{
		Analysis:      analysis,
		ExecutionTime: time.Since(startTime),
	}

	response.Summary = f.generateAnalysisSummary(analysis)
	response.ActionableInsights = f.generateActionableInsights(analysis)

	if req.OutputFormat == "actionable" || req.GenerateOptimizations {
		response.ConfigSuggestions = f.generateConfigSuggestions(analysis)
	}

	f.logger.Info(fmt.Sprintf("Dependency analysis completed: %d options, %d dependencies, %d conflicts", 
		len(analysis.ConfigurationOptions), len(analysis.Dependencies), len(analysis.Conflicts)))

	return functionbase.SuccessResult(response, time.Since(startTime)), nil
}

// generateAnalysisSummary creates a high-level analysis summary
func (f *DependencyFunction) generateAnalysisSummary(analysis *dependency.DependencyAnalysis) *AnalysisSummary {
	requiredDeps := 0
	for _, dep := range analysis.Dependencies {
		if dep.Type == dependency.DependencyRequired {
			requiredDeps++
		}
	}

	criticalConflicts := 0
	for _, conflict := range analysis.Conflicts {
		if conflict.Severity == "critical" {
			criticalConflicts++
		}
	}

	var keyIssues []string
	if criticalConflicts > 0 {
		keyIssues = append(keyIssues, fmt.Sprintf("%d critical conflicts require immediate attention", criticalConflicts))
	}
	if requiredDeps > 0 {
		keyIssues = append(keyIssues, fmt.Sprintf("%d required dependencies may be missing", requiredDeps))
	}

	status := "healthy"
	overallScore := analysis.ValidationResults.Score
	if criticalConflicts > 0 {
		status = "critical"
	} else if len(analysis.Conflicts) > 0 || requiredDeps > 5 {
		status = "warning"
	}

	hardwareOptimized := len(analysis.HardwareDependencies) > 0
	for _, dep := range analysis.HardwareDependencies {
		if dep.Compatible {
			hardwareOptimized = true
			break
		}
	}

	return &AnalysisSummary{
		OverallScore:      overallScore,
		TotalOptions:      len(analysis.ConfigurationOptions),
		Dependencies:      len(analysis.Dependencies),
		RequiredDeps:      requiredDeps,
		Conflicts:         len(analysis.Conflicts),
		CriticalConflicts: criticalConflicts,
		Recommendations:   len(analysis.Recommendations),
		HardwareOptimized: hardwareOptimized,
		Status:            status,
		KeyIssues:         keyIssues,
	}
}

// generateActionableInsights creates actionable recommendations
func (f *DependencyFunction) generateActionableInsights(analysis *dependency.DependencyAnalysis) []*ActionableInsight {
	var insights []*ActionableInsight

	// Critical conflicts first
	for _, conflict := range analysis.Conflicts {
		if conflict.Severity == "critical" {
			insight := &ActionableInsight{
				Type:        "conflict",
				Priority:    "critical",
				Title:       "Configuration Conflict",
				Description: conflict.Description,
				Action:      "resolve",
				Benefits:    []string{"Prevents system errors", "Ensures stable configuration"},
				Category:    "compatibility",
			}

			if len(conflict.Resolution) > 0 {
				insight.Action = conflict.Resolution[0]
			}

			insights = append(insights, insight)
		}
	}

	// Required dependencies
	for _, dep := range analysis.Dependencies {
		if dep.Type == dependency.DependencyRequired {
			insight := &ActionableInsight{
				Type:        "dependency",
				Priority:    "high",
				Title:       fmt.Sprintf("Missing Required Option: %s", dep.To),
				Description: fmt.Sprintf("Required by %s: %s", dep.From, dep.Description),
				Action:      "add_option",
				ConfigLine:  fmt.Sprintf("%s = true;", dep.To),
				Benefits:    []string{"Ensures proper functionality", "Resolves dependency requirements"},
				Category:    "dependencies",
			}
			insights = append(insights, insight)
		}
	}

	// Hardware-based recommendations
	for _, rec := range analysis.Recommendations {
		if rec.HardwareBased && rec.Priority >= 7 {
			priority := "medium"
			if rec.Priority >= 8 {
				priority = "high"
			}

			configLine := fmt.Sprintf("%s = %v;", rec.Option, rec.Value)
			if rec.Action == "remove" {
				configLine = fmt.Sprintf("# %s = %v; # Disabled", rec.Option, rec.Value)
			}

			insight := &ActionableInsight{
				Type:        "hardware",
				Priority:    priority,
				Title:       fmt.Sprintf("Hardware Optimization: %s", rec.Option),
				Description: rec.Reason,
				Action:      rec.Action,
				ConfigLine:  configLine,
				Benefits:    rec.Benefits,
				Category:    "hardware",
			}
			insights = append(insights, insight)
		}
	}

	// Performance optimizations
	for _, rec := range analysis.Recommendations {
		if rec.Type == dependency.RecommendationPerformance && rec.Priority >= 5 {
			insight := &ActionableInsight{
				Type:        "performance",
				Priority:    "medium",
				Title:       fmt.Sprintf("Performance: %s", rec.Option),
				Description: rec.Reason,
				Action:      rec.Action,
				ConfigLine:  fmt.Sprintf("%s = %v;", rec.Option, rec.Value),
				Benefits:    rec.Benefits,
				Category:    "performance",
			}
			insights = append(insights, insight)
		}
	}

	return insights
}

// generateConfigSuggestions generates configuration suggestions
func (f *DependencyFunction) generateConfigSuggestions(analysis *dependency.DependencyAnalysis) string {
	var suggestions []string

	suggestions = append(suggestions, "# NixOS Configuration Suggestions")
	suggestions = append(suggestions, "# Generated by nixai dependency analysis")
	suggestions = append(suggestions, "")
	suggestions = append(suggestions, "{")

	// Add missing required dependencies
	addedOptions := make(map[string]bool)
	for _, dep := range analysis.Dependencies {
		if dep.Type == dependency.DependencyRequired && !addedOptions[dep.To] {
			comment := fmt.Sprintf("  # Required by %s", dep.From)
			option := fmt.Sprintf("  %s = true;", dep.To)
			suggestions = append(suggestions, comment, option, "")
			addedOptions[dep.To] = true
		}
	}

	// Add high-priority hardware recommendations
	for _, rec := range analysis.Recommendations {
		if rec.HardwareBased && rec.Priority >= 7 && !addedOptions[rec.Option] {
			comment := fmt.Sprintf("  # %s", rec.Reason)
			var option string
			if rec.Action == "add" {
				switch v := rec.Value.(type) {
				case []string:
					option = fmt.Sprintf("  %s = [ %s ];", rec.Option, 
						strings.Join(f.quoteStrings(v), " "))
				case string:
					option = fmt.Sprintf("  %s = \"%s\";", rec.Option, v)
				default:
					option = fmt.Sprintf("  %s = %v;", rec.Option, v)
				}
			}
			
			if option != "" {
				suggestions = append(suggestions, comment, option, "")
				addedOptions[rec.Option] = true
			}
		}
	}

	// Add performance optimizations
	for _, rec := range analysis.Recommendations {
		if rec.Type == dependency.RecommendationPerformance && !addedOptions[rec.Option] {
			comment := fmt.Sprintf("  # Performance: %s", rec.Reason)
			option := fmt.Sprintf("  %s = %v;", rec.Option, rec.Value)
			suggestions = append(suggestions, comment, option, "")
			addedOptions[rec.Option] = true
		}
	}

	suggestions = append(suggestions, "}")

	return strings.Join(suggestions, "\n")
}

// Helper method to quote strings for Nix arrays
func (f *DependencyFunction) quoteStrings(strings []string) []string {
	quoted := make([]string, len(strings))
	for i, s := range strings {
		quoted[i] = fmt.Sprintf("\"%s\"", s)
	}
	return quoted
}