package nixlang

import (
	"fmt"
	"regexp"
	"strings"
)

// NixAnalyzer provides comprehensive analysis of Nix expressions
type NixAnalyzer struct {
	parser           *NixParser
	patterns         map[string]*AnalysisPattern
	securityRules    []SecurityRule
	optimizationRules []OptimizationRule
}

// AnalysisResult contains the results of Nix expression analysis
type AnalysisResult struct {
	Expression       *NixExpression      `json:"expression"`
	Issues           []Issue             `json:"issues"`
	Suggestions      []Suggestion        `json:"suggestions"`
	SecurityFindings []SecurityFinding   `json:"security_findings"`
	Optimizations    []Optimization      `json:"optimizations"`
	Intent           IntentAnalysis      `json:"intent"`
	Complexity       ComplexityAnalysis  `json:"complexity"`
	Dependencies     DependencyAnalysis  `json:"dependencies"`
	Quality          QualityMetrics      `json:"quality"`
}

// Issue represents a problem found in the Nix expression
type Issue struct {
	Type        IssueType `json:"type"`
	Severity    Severity  `json:"severity"`
	Message     string    `json:"message"`
	Position    Position  `json:"position"`
	Suggestion  string    `json:"suggestion,omitempty"`
	Category    string    `json:"category"`
	Code        string    `json:"code"`
}

// IssueType represents different types of issues
type IssueType string

const (
	IssueSyntax      IssueType = "syntax"
	IssueSemantic    IssueType = "semantic"
	IssueSecurity    IssueType = "security"
	IssuePerformance IssueType = "performance"
	IssueStyle       IssueType = "style"
	IssueAntiPattern IssueType = "antipattern"
)

// Severity represents issue severity levels
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
	SeverityHint    Severity = "hint"
)

// Suggestion represents an improvement suggestion
type Suggestion struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Code        string    `json:"code,omitempty"`
	Position    Position  `json:"position"`
	Confidence  float64   `json:"confidence"`
	Impact      string    `json:"impact"`
	Category    string    `json:"category"`
}

// SecurityFinding represents security-related findings
type SecurityFinding struct {
	Type         string    `json:"type"`
	Severity     Severity  `json:"severity"`
	Description  string    `json:"description"`
	Position     Position  `json:"position"`
	Mitigation   string    `json:"mitigation"`
	CWE          string    `json:"cwe,omitempty"`
	CVSS         float64   `json:"cvss,omitempty"`
}

// Optimization represents a performance or efficiency optimization
type Optimization struct {
	Type         string    `json:"type"`
	Description  string    `json:"description"`
	Impact       string    `json:"impact"`
	Code         string    `json:"code,omitempty"`
	EstimatedGain string   `json:"estimated_gain"`
}

// IntentAnalysis analyzes the intent behind the configuration
type IntentAnalysis struct {
	PrimaryIntent    string            `json:"primary_intent"`
	SecondaryIntents []string          `json:"secondary_intents"`
	Confidence       float64           `json:"confidence"`
	Context          string            `json:"context"`
	Patterns         []string          `json:"patterns"`
	Annotations      map[string]string `json:"annotations"`
}

// ComplexityAnalysis provides complexity metrics
type ComplexityAnalysis struct {
	Total           int                    `json:"total"`
	Cyclomatic      int                    `json:"cyclomatic"`
	Cognitive       int                    `json:"cognitive"`
	Nesting         int                    `json:"nesting"`
	ByType          map[string]int         `json:"by_type"`
	Hotspots        []ComplexityHotspot    `json:"hotspots"`
}

// ComplexityHotspot identifies complex areas
type ComplexityHotspot struct {
	Position    Position `json:"position"`
	Complexity  int      `json:"complexity"`
	Type        string   `json:"type"`
	Suggestion  string   `json:"suggestion"`
}

// DependencyAnalysis analyzes dependencies
type DependencyAnalysis struct {
	Direct       []Dependency     `json:"direct"`
	Indirect     []Dependency     `json:"indirect"`
	Circular     []string         `json:"circular"`
	Unused       []string         `json:"unused"`
	Missing      []string         `json:"missing"`
	Graph        DependencyGraph  `json:"graph"`
}

// Dependency represents a dependency relationship
type Dependency struct {
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Version      string    `json:"version,omitempty"`
	Source       string    `json:"source"`
	Required     bool      `json:"required"`
	Position     Position  `json:"position"`
}

// DependencyGraph represents the dependency graph
type DependencyGraph struct {
	Nodes []DependencyNode `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Metadata map[string]string `json:"metadata"`
}

// DependencyEdge represents an edge in the dependency graph
type DependencyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

// QualityMetrics provides code quality metrics
type QualityMetrics struct {
	Overall         float64           `json:"overall"`
	Maintainability float64           `json:"maintainability"`
	Reliability     float64           `json:"reliability"`
	Security        float64           `json:"security"`
	Performance     float64           `json:"performance"`
	Readability     float64           `json:"readability"`
	Details         map[string]float64 `json:"details"`
}

// AnalysisPattern defines patterns for analyzing Nix expressions
type AnalysisPattern struct {
	Name        string         `json:"name"`
	Pattern     *regexp.Regexp `json:"-"`
	Intent      string         `json:"intent"`
	Category    string         `json:"category"`
	Confidence  float64        `json:"confidence"`
	Description string         `json:"description"`
}

// SecurityRule defines security analysis rules
type SecurityRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Pattern     *regexp.Regexp `json:"-"`
	Severity    Severity       `json:"severity"`
	Description string         `json:"description"`
	Mitigation  string         `json:"mitigation"`
	CWE         string         `json:"cwe"`
}

// OptimizationRule defines optimization rules
type OptimizationRule struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Pattern      *regexp.Regexp `json:"-"`
	Description  string         `json:"description"`
	Impact       string         `json:"impact"`
	Replacement  string         `json:"replacement"`
}

// NewNixAnalyzer creates a new Nix analyzer
func NewNixAnalyzer() *NixAnalyzer {
	analyzer := &NixAnalyzer{
		parser:   NewNixParser(),
		patterns: make(map[string]*AnalysisPattern),
	}
	
	analyzer.initializePatterns()
	analyzer.initializeSecurityRules()
	analyzer.initializeOptimizationRules()
	
	return analyzer
}

// AnalyzeExpression performs comprehensive analysis of a Nix expression
func (a *NixAnalyzer) AnalyzeExpression(source string) (*AnalysisResult, error) {
	// Parse the expression
	expr, err := a.parser.ParseExpression(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}
	
	result := &AnalysisResult{
		Expression: expr,
		Issues:     []Issue{},
		Suggestions: []Suggestion{},
		SecurityFindings: []SecurityFinding{},
		Optimizations: []Optimization{},
	}
	
	// Perform various analyses
	a.analyzeIntent(expr, result)
	a.analyzeComplexity(expr, result)
	a.analyzeDependencies(expr, result)
	a.analyzeSecurity(expr, result)
	a.analyzeOptimizations(expr, result)
	a.analyzeQuality(expr, result)
	a.detectAntiPatterns(expr, result)
	
	return result, nil
}

// initializePatterns sets up analysis patterns
func (a *NixAnalyzer) initializePatterns() {
	patterns := []struct {
		name        string
		pattern     string
		intent      string
		category    string
		confidence  float64
		description string
	}{
		{
			"nixos_service_config",
			`services\.\w+\.enable\s*=\s*true`,
			"service_configuration",
			"nixos",
			0.9,
			"NixOS service configuration pattern",
		},
		{
			"package_installation",
			`environment\.systemPackages\s*=`,
			"package_management",
			"nixos",
			0.85,
			"System package installation",
		},
		{
			"user_configuration",
			`users\.users\.\w+`,
			"user_management",
			"nixos",
			0.8,
			"User account configuration",
		},
		{
			"networking_config",
			`networking\.(hostName|domain|firewall)`,
			"network_configuration",
			"nixos",
			0.85,
			"Network configuration",
		},
		{
			"security_config",
			`security\.(sudo|pam|acme)`,
			"security_configuration",
			"nixos",
			0.9,
			"Security-related configuration",
		},
		{
			"hardware_config",
			`hardware\.(cpu|graphics|bluetooth)`,
			"hardware_configuration",
			"nixos",
			0.8,
			"Hardware configuration",
		},
		{
			"boot_config",
			`boot\.(loader|kernel|initrd)`,
			"boot_configuration",
			"nixos",
			0.85,
			"Boot and kernel configuration",
		},
		{
			"development_env",
			`(mkShell|buildInputs|shellHook)`,
			"development_environment",
			"development",
			0.8,
			"Development environment setup",
		},
		{
			"package_derivation",
			`(stdenv\.mkDerivation|buildGoModule|buildPythonPackage)`,
			"package_derivation",
			"packaging",
			0.9,
			"Package derivation definition",
		},
		{
			"flake_definition",
			`(inputs|outputs|description)\s*=`,
			"flake_configuration",
			"flakes",
			0.85,
			"Nix flake definition",
		},
	}
	
	for _, p := range patterns {
		compiled, err := regexp.Compile(p.pattern)
		if err == nil {
			a.patterns[p.name] = &AnalysisPattern{
				Name:        p.name,
				Pattern:     compiled,
				Intent:      p.intent,
				Category:    p.category,
				Confidence:  p.confidence,
				Description: p.description,
			}
		}
	}
}

// initializeSecurityRules sets up security analysis rules
func (a *NixAnalyzer) initializeSecurityRules() {
	rules := []struct {
		id          string
		name        string
		pattern     string
		severity    Severity
		description string
		mitigation  string
		cwe         string
	}{
		{
			"insecure_url",
			"Insecure HTTP URL",
			`http://[^\s"']+`,
			SeverityWarning,
			"Using insecure HTTP URLs instead of HTTPS",
			"Replace HTTP URLs with HTTPS equivalents",
			"CWE-319",
		},
		{
			"hardcoded_secret",
			"Hardcoded Secret",
			`(password|secret|key|token)\s*=\s*"[^"]+`,
			SeverityError,
			"Hardcoded secrets in configuration",
			"Use external secret management or environment variables",
			"CWE-798",
		},
		{
			"weak_permissions",
			"Weak File Permissions",
			`mode\s*=\s*"[0-7]*[2367][0-7]*"`,
			SeverityWarning,
			"File permissions allow world write access",
			"Remove world write permissions for security",
			"CWE-276",
		},
		{
			"root_execution",
			"Root Execution",
			`user\s*=\s*"root"`,
			SeverityWarning,
			"Service running as root user",
			"Use a dedicated service user with minimal privileges",
			"CWE-250",
		},
		{
			"disabled_firewall",
			"Disabled Firewall",
			`networking\.firewall\.enable\s*=\s*false`,
			SeverityWarning,
			"Firewall is disabled",
			"Enable firewall and configure appropriate rules",
			"CWE-284",
		},
		{
			"weak_ssh_config",
			"Weak SSH Configuration",
			`services\.openssh\.permitRootLogin\s*=\s*"yes"`,
			SeverityError,
			"SSH allows root login",
			"Disable root login and use sudo for administrative access",
			"CWE-250",
		},
		{
			"insecure_package",
			"Insecure Package",
			`pkgs\.(flash|java|adobe)`,
			SeverityInfo,
			"Using potentially insecure packages",
			"Consider alternatives or ensure packages are up to date",
			"CWE-1104",
		},
	}
	
	for _, rule := range rules {
		compiled, err := regexp.Compile(rule.pattern)
		if err == nil {
			a.securityRules = append(a.securityRules, SecurityRule{
				ID:          rule.id,
				Name:        rule.name,
				Pattern:     compiled,
				Severity:    rule.severity,
				Description: rule.description,
				Mitigation:  rule.mitigation,
				CWE:         rule.cwe,
			})
		}
	}
}

// initializeOptimizationRules sets up optimization rules
func (a *NixAnalyzer) initializeOptimizationRules() {
	rules := []struct {
		id          string
		name        string
		pattern     string
		description string
		impact      string
		replacement string
	}{
		{
			"unnecessary_with",
			"Unnecessary with statement",
			`with\s+pkgs;\s*\[\s*(\w+)\s*\]`,
			"Using 'with pkgs' for single package",
			"Reduces namespace pollution",
			"Use pkgs.package directly",
		},
		{
			"duplicate_packages",
			"Duplicate package references",
			`(pkgs\.\w+).*\1`,
			"Package referenced multiple times",
			"Reduces redundancy",
			"Extract to variable",
		},
		{
			"large_attribute_set",
			"Large attribute set",
			`{\s*([^}]{500,})\s*}`,
			"Very large attribute set",
			"Improves maintainability",
			"Split into multiple modules",
		},
		{
			"nested_imports",
			"Deeply nested imports",
			`import\s+.*import\s+.*import`,
			"Multiple nested imports",
			"Reduces complexity",
			"Flatten import structure",
		},
		{
			"unused_variable",
			"Unused variable",
			`(\w+)\s*=.*(?!\1)`,
			"Variable defined but not used",
			"Reduces clutter",
			"Remove unused variable",
		},
	}
	
	for _, rule := range rules {
		compiled, err := regexp.Compile(rule.pattern)
		if err == nil {
			a.optimizationRules = append(a.optimizationRules, OptimizationRule{
				ID:          rule.id,
				Name:        rule.name,
				Pattern:     compiled,
				Description: rule.description,
				Impact:      rule.impact,
				Replacement: rule.replacement,
			})
		}
	}
}

// analyzeIntent determines the intent behind the configuration
func (a *NixAnalyzer) analyzeIntent(expr *NixExpression, result *AnalysisResult) {
	intent := IntentAnalysis{
		Confidence:  0.0,
		Patterns:    []string{},
		Annotations: make(map[string]string),
	}
	
	// Convert expression to string for pattern matching
	source := a.expressionToString(expr)
	
	var bestMatch *AnalysisPattern
	var bestConfidence float64
	
	// Find the best matching pattern
	for _, pattern := range a.patterns {
		if pattern.Pattern.MatchString(source) {
			intent.Patterns = append(intent.Patterns, pattern.Name)
			if pattern.Confidence > bestConfidence {
				bestConfidence = pattern.Confidence
				bestMatch = pattern
			}
		}
	}
	
	if bestMatch != nil {
		intent.PrimaryIntent = bestMatch.Intent
		intent.Context = bestMatch.Category
		intent.Confidence = bestConfidence
		intent.Annotations["primary_pattern"] = bestMatch.Name
		intent.Annotations["description"] = bestMatch.Description
	}
	
	// Analyze secondary intents
	for _, pattern := range a.patterns {
		if pattern != bestMatch && pattern.Pattern.MatchString(source) {
			intent.SecondaryIntents = append(intent.SecondaryIntents, pattern.Intent)
		}
	}
	
	result.Intent = intent
}

// analyzeComplexity calculates complexity metrics
func (a *NixAnalyzer) analyzeComplexity(expr *NixExpression, result *AnalysisResult) {
	complexity := ComplexityAnalysis{
		ByType:   make(map[string]int),
		Hotspots: []ComplexityHotspot{},
	}
	
	// Calculate various complexity metrics
	complexity.Total = a.calculateTotalComplexity(expr)
	complexity.Cyclomatic = a.calculateCyclomaticComplexity(expr)
	complexity.Cognitive = a.calculateCognitiveComplexity(expr)
	complexity.Nesting = a.calculateNestingDepth(expr)
	
	// Count complexity by type
	a.countComplexityByType(expr, complexity.ByType)
	
	// Find complexity hotspots
	a.findComplexityHotspots(expr, &complexity.Hotspots)
	
	result.Complexity = complexity
}

// analyzeDependencies analyzes dependency relationships
func (a *NixAnalyzer) analyzeDependencies(expr *NixExpression, result *AnalysisResult) {
	deps := DependencyAnalysis{
		Direct:   []Dependency{},
		Indirect: []Dependency{},
		Circular: []string{},
		Unused:   []string{},
		Missing:  []string{},
		Graph: DependencyGraph{
			Nodes: []DependencyNode{},
			Edges: []DependencyEdge{},
		},
	}
	
	// Collect direct dependencies
	a.collectDirectDependencies(expr, &deps.Direct)
	
	// Build dependency graph
	a.buildDependencyGraph(expr, &deps.Graph)
	
	// Detect circular dependencies
	deps.Circular = a.detectCircularDependencies(&deps.Graph)
	
	result.Dependencies = deps
}

// analyzeSecurity performs security analysis
func (a *NixAnalyzer) analyzeSecurity(expr *NixExpression, result *AnalysisResult) {
	source := a.expressionToString(expr)
	
	for _, rule := range a.securityRules {
		if matches := rule.Pattern.FindAllStringIndex(source, -1); matches != nil {
			for _, match := range matches {
				finding := SecurityFinding{
					Type:        rule.ID,
					Severity:    rule.Severity,
					Description: rule.Description,
					Position:    a.indexToPosition(source, match[0]),
					Mitigation:  rule.Mitigation,
					CWE:         rule.CWE,
				}
				
				// Calculate CVSS score based on severity
				switch rule.Severity {
				case SeverityError:
					finding.CVSS = 7.5
				case SeverityWarning:
					finding.CVSS = 4.0
				case SeverityInfo:
					finding.CVSS = 2.0
				}
				
				result.SecurityFindings = append(result.SecurityFindings, finding)
			}
		}
	}
}

// analyzeOptimizations finds optimization opportunities
func (a *NixAnalyzer) analyzeOptimizations(expr *NixExpression, result *AnalysisResult) {
	source := a.expressionToString(expr)
	
	for _, rule := range a.optimizationRules {
		if rule.Pattern.MatchString(source) {
			optimization := Optimization{
				Type:          rule.ID,
				Description:   rule.Description,
				Impact:        rule.Impact,
				EstimatedGain: "Low", // Could be calculated more precisely
			}
			
			if rule.Replacement != "" {
				optimization.Code = rule.Replacement
			}
			
			result.Optimizations = append(result.Optimizations, optimization)
		}
	}
}

// analyzeQuality calculates quality metrics
func (a *NixAnalyzer) analyzeQuality(expr *NixExpression, result *AnalysisResult) {
	quality := QualityMetrics{
		Details: make(map[string]float64),
	}
	
	// Calculate individual metrics
	quality.Maintainability = a.calculateMaintainability(result)
	quality.Reliability = a.calculateReliability(result)
	quality.Security = a.calculateSecurity(result)
	quality.Performance = a.calculatePerformance(result)
	quality.Readability = a.calculateReadability(expr)
	
	// Calculate overall quality
	quality.Overall = (quality.Maintainability + quality.Reliability + 
					  quality.Security + quality.Performance + quality.Readability) / 5.0
	
	// Store detailed metrics
	quality.Details["complexity_score"] = float64(result.Complexity.Total) / 100.0
	quality.Details["security_issues"] = float64(len(result.SecurityFindings))
	quality.Details["optimization_opportunities"] = float64(len(result.Optimizations))
	
	result.Quality = quality
}

// detectAntiPatterns finds common anti-patterns
func (a *NixAnalyzer) detectAntiPatterns(expr *NixExpression, result *AnalysisResult) {
	source := a.expressionToString(expr)
	
	antiPatterns := []struct {
		pattern     *regexp.Regexp
		message     string
		suggestion  string
		severity    Severity
	}{
		{
			regexp.MustCompile(`with\s+pkgs;\s+with\s+`),
			"Nested 'with' statements create namespace conflicts",
			"Use specific package references instead",
			SeverityWarning,
		},
		{
			regexp.MustCompile(`(\w+)\s*=\s*\w*\1\w*`),
			"Potential self-referential assignment detected",
			"Check for infinite recursion",
			SeverityError,
		},
		{
			regexp.MustCompile(`import\s+<nixpkgs>\s+{\s*}`),
			"Empty nixpkgs import",
			"Remove unnecessary import or add config",
			SeverityInfo,
		},
		{
			regexp.MustCompile(`[\s\S]{5000,}`),
			"Very large configuration file",
			"Consider splitting into multiple modules",
			SeverityInfo,
		},
	}
	
	for _, ap := range antiPatterns {
		if matches := ap.pattern.FindAllStringIndex(source, -1); matches != nil {
			for _, match := range matches {
				issue := Issue{
					Type:       IssueAntiPattern,
					Severity:   ap.severity,
					Message:    ap.message,
					Position:   a.indexToPosition(source, match[0]),
					Suggestion: ap.suggestion,
					Category:   "anti-pattern",
					Code:       "AP001",
				}
				result.Issues = append(result.Issues, issue)
			}
		}
	}
}

// Helper methods for complexity calculations
func (a *NixAnalyzer) calculateTotalComplexity(expr *NixExpression) int {
	if expr == nil {
		return 0
	}
	return expr.GetComplexity()
}

func (a *NixAnalyzer) calculateCyclomaticComplexity(expr *NixExpression) int {
	// Simplified cyclomatic complexity calculation
	complexity := 1 // Base complexity
	
	if expr.Type == ExprIf {
		complexity += 1
	}
	
	for _, child := range expr.Children {
		complexity += a.calculateCyclomaticComplexity(&child)
	}
	
	for _, attr := range expr.Attrs {
		complexity += a.calculateCyclomaticComplexity(&attr)
	}
	
	return complexity
}

func (a *NixAnalyzer) calculateCognitiveComplexity(expr *NixExpression) int {
	// Cognitive complexity considers nesting and control flow
	return a.calculateNestingDepth(expr) * 2
}

func (a *NixAnalyzer) calculateNestingDepth(expr *NixExpression) int {
	if expr == nil {
		return 0
	}
	
	maxDepth := 0
	
	for _, child := range expr.Children {
		depth := a.calculateNestingDepth(&child)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	
	for _, attr := range expr.Attrs {
		depth := a.calculateNestingDepth(&attr)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	
	return maxDepth + 1
}

func (a *NixAnalyzer) countComplexityByType(expr *NixExpression, counts map[string]int) {
	if expr == nil {
		return
	}
	
	counts[string(expr.Type)]++
	
	for _, child := range expr.Children {
		a.countComplexityByType(&child, counts)
	}
	
	for _, attr := range expr.Attrs {
		a.countComplexityByType(&attr, counts)
	}
}

func (a *NixAnalyzer) findComplexityHotspots(expr *NixExpression, hotspots *[]ComplexityHotspot) {
	if expr == nil {
		return
	}
	
	complexity := expr.GetComplexity()
	if complexity > 10 { // Threshold for complexity hotspot
		hotspot := ComplexityHotspot{
			Position:   expr.Position,
			Complexity: complexity,
			Type:       string(expr.Type),
			Suggestion: "Consider breaking this into smaller parts",
		}
		*hotspots = append(*hotspots, hotspot)
	}
	
	for _, child := range expr.Children {
		a.findComplexityHotspots(&child, hotspots)
	}
	
	for _, attr := range expr.Attrs {
		a.findComplexityHotspots(&attr, hotspots)
	}
}

// Helper methods for dependency analysis
func (a *NixAnalyzer) collectDirectDependencies(expr *NixExpression, deps *[]Dependency) {
	if expr == nil {
		return
	}
	
	for _, depName := range expr.Metadata.Dependencies {
		dep := Dependency{
			Name:     depName,
			Type:     "variable",
			Source:   "local",
			Required: true,
			Position: expr.Position,
		}
		*deps = append(*deps, dep)
	}
	
	for _, child := range expr.Children {
		a.collectDirectDependencies(&child, deps)
	}
	
	for _, attr := range expr.Attrs {
		a.collectDirectDependencies(&attr, deps)
	}
}

func (a *NixAnalyzer) buildDependencyGraph(expr *NixExpression, graph *DependencyGraph) {
	// Simplified dependency graph building
	if expr == nil {
		return
	}
	
	// Add node for this expression
	node := DependencyNode{
		ID:   fmt.Sprintf("node_%p", expr),
		Name: string(expr.Type),
		Type: string(expr.Type),
		Metadata: make(map[string]string),
	}
	graph.Nodes = append(graph.Nodes, node)
	
	// Add edges for dependencies
	for _, dep := range expr.Metadata.Dependencies {
		edge := DependencyEdge{
			From: node.ID,
			To:   dep,
			Type: "depends_on",
		}
		graph.Edges = append(graph.Edges, edge)
	}
}

func (a *NixAnalyzer) detectCircularDependencies(graph *DependencyGraph) []string {
	// Simplified circular dependency detection
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var circular []string
	
	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true
		
		for _, edge := range graph.Edges {
			if edge.From == node {
				if inStack[edge.To] {
					circular = append(circular, fmt.Sprintf("%s -> %s", node, edge.To))
					return true
				} else if !visited[edge.To] && dfs(edge.To) {
					return true
				}
			}
		}
		
		inStack[node] = false
		return false
	}
	
	for _, node := range graph.Nodes {
		if !visited[node.ID] {
			dfs(node.ID)
		}
	}
	
	return circular
}

// Helper methods for quality calculations
func (a *NixAnalyzer) calculateMaintainability(result *AnalysisResult) float64 {
	base := 1.0
	
	// Reduce score based on complexity
	complexityPenalty := float64(result.Complexity.Total) / 100.0
	if complexityPenalty > 1.0 {
		complexityPenalty = 1.0
	}
	
	// Reduce score based on issues
	issuesPenalty := float64(len(result.Issues)) / 20.0
	if issuesPenalty > 1.0 {
		issuesPenalty = 1.0
	}
	
	return base - complexityPenalty - issuesPenalty
}

func (a *NixAnalyzer) calculateReliability(result *AnalysisResult) float64 {
	base := 1.0
	
	// Count error-level issues
	errors := 0
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError {
			errors++
		}
	}
	
	errorPenalty := float64(errors) / 10.0
	if errorPenalty > 1.0 {
		errorPenalty = 1.0
	}
	
	return base - errorPenalty
}

func (a *NixAnalyzer) calculateSecurity(result *AnalysisResult) float64 {
	base := 1.0
	
	// Penalize based on security findings
	securityPenalty := 0.0
	for _, finding := range result.SecurityFindings {
		switch finding.Severity {
		case SeverityError:
			securityPenalty += 0.3
		case SeverityWarning:
			securityPenalty += 0.2
		case SeverityInfo:
			securityPenalty += 0.1
		}
	}
	
	if securityPenalty > 1.0 {
		securityPenalty = 1.0
	}
	
	return base - securityPenalty
}

func (a *NixAnalyzer) calculatePerformance(result *AnalysisResult) float64 {
	base := 1.0
	
	// Penalize based on optimization opportunities
	optimizationPenalty := float64(len(result.Optimizations)) / 20.0
	if optimizationPenalty > 1.0 {
		optimizationPenalty = 1.0
	}
	
	return base - optimizationPenalty
}

func (a *NixAnalyzer) calculateReadability(expr *NixExpression) float64 {
	base := 1.0
	
	// Simple readability calculation based on complexity
	if expr == nil {
		return base
	}
	
	complexity := float64(expr.GetComplexity())
	readabilityPenalty := complexity / 200.0
	if readabilityPenalty > 1.0 {
		readabilityPenalty = 1.0
	}
	
	return base - readabilityPenalty
}

// Utility methods
func (a *NixAnalyzer) expressionToString(expr *NixExpression) string {
	// Simplified expression to string conversion
	// In a real implementation, this would reconstruct the Nix syntax
	if expr == nil {
		return ""
	}
	
	switch expr.Type {
	case ExprString:
		if str, ok := expr.Value.(string); ok {
			return fmt.Sprintf("\"%s\"", str)
		}
	case ExprNumber:
		return fmt.Sprintf("%v", expr.Value)
	case ExprBool:
		return fmt.Sprintf("%v", expr.Value)
	case ExprVariable:
		if str, ok := expr.Value.(string); ok {
			return str
		}
	case ExprAttrSet:
		var parts []string
		for key, value := range expr.Attrs {
			valueStr := a.expressionToString(&value)
			parts = append(parts, fmt.Sprintf("%s = %s;", key, valueStr))
		}
		return fmt.Sprintf("{ %s }", strings.Join(parts, " "))
	case ExprList:
		var parts []string
		for _, child := range expr.Children {
			parts = append(parts, a.expressionToString(&child))
		}
		return fmt.Sprintf("[ %s ]", strings.Join(parts, " "))
	}
	
	return fmt.Sprintf("<%s>", expr.Type)
}

func (a *NixAnalyzer) indexToPosition(source string, index int) Position {
	line := 1
	column := 1
	
	for i := 0; i < index && i < len(source); i++ {
		if source[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	
	return Position{
		Line:   line,
		Column: column,
		Offset: index,
	}
}