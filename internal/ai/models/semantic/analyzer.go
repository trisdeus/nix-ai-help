// Package semantic provides semantic analysis capabilities for NixOS configurations
package semantic

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/pkg/logger"
)

// SemanticAnalyzer provides intelligent analysis of NixOS configurations
type SemanticAnalyzer struct {
	Logger *logger.Logger
}

// AnalysisResult contains the results of semantic analysis
type AnalysisResult struct {
	ConfigPath        string                 `json:"config_path"`
	Intent            ConfigIntent           `json:"intent"`
	Issues            []Issue                `json:"issues"`
	Suggestions       []Suggestion           `json:"suggestions"`
	Dependencies      []Dependency           `json:"dependencies"`
	SecurityAnalysis  SecurityAnalysis       `json:"security_analysis"`
	Performance       PerformanceAnalysis    `json:"performance_analysis"`
	Compatibility     CompatibilityAnalysis  `json:"compatibility_analysis"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ConfigIntent represents the inferred intent of a configuration
type ConfigIntent struct {
	Purpose       string            `json:"purpose"`        // "desktop", "server", "development", etc.
	Services      []string          `json:"services"`       // Detected services
	Environment   string            `json:"environment"`    // "development", "production", "testing"
	UserProfile   string            `json:"user_profile"`   // "beginner", "intermediate", "expert"
	Architecture  string            `json:"architecture"`   // "x86_64", "aarch64", etc.
	Confidence    float64           `json:"confidence"`     // 0.0 - 1.0
	Context       map[string]string `json:"context"`
}

// Issue represents a configuration issue
type Issue struct {
	Type        string            `json:"type"`         // "security", "performance", "compatibility", "syntax"
	Severity    string            `json:"severity"`     // "critical", "warning", "info"
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Location    Location          `json:"location"`
	Fix         *Fix              `json:"fix,omitempty"`
	References  []string          `json:"references"`
	Context     map[string]string `json:"context"`
}

// Suggestion represents an optimization or improvement suggestion
type Suggestion struct {
	Type        string            `json:"type"`         // "optimization", "security", "best_practice"
	Priority    string            `json:"priority"`     // "high", "medium", "low"
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Before      string            `json:"before"`
	After       string            `json:"after"`
	Rationale   string            `json:"rationale"`
	References  []string          `json:"references"`
	Context     map[string]string `json:"context"`
}

// Dependency represents a configuration dependency
type Dependency struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`         // "service", "package", "module"
	Required    bool     `json:"required"`
	Missing     bool     `json:"missing"`
	Version     string   `json:"version,omitempty"`
	Alternatives []string `json:"alternatives,omitempty"`
}

// SecurityAnalysis contains security-specific analysis
type SecurityAnalysis struct {
	Score           float64           `json:"score"`           // 0.0 - 1.0
	Vulnerabilities []Vulnerability   `json:"vulnerabilities"`
	Recommendations []string          `json:"recommendations"`
	Compliance      map[string]bool   `json:"compliance"`      // SOC2, HIPAA, etc.
	Risk            string            `json:"risk"`            // "low", "medium", "high", "critical"
}

// PerformanceAnalysis contains performance-specific analysis
type PerformanceAnalysis struct {
	Score           float64         `json:"score"`           // 0.0 - 1.0
	Bottlenecks     []Bottleneck    `json:"bottlenecks"`
	Optimizations   []Optimization  `json:"optimizations"`
	ResourceUsage   ResourceUsage   `json:"resource_usage"`
	Predictions     []Prediction    `json:"predictions"`
}

// CompatibilityAnalysis contains compatibility analysis
type CompatibilityAnalysis struct {
	NixOSVersion    string              `json:"nixos_version"`
	Compatibility   float64             `json:"compatibility"`   // 0.0 - 1.0
	DeprecatedFeatures []DeprecatedFeature `json:"deprecated_features"`
	MigrationPath   []MigrationStep     `json:"migration_path"`
	Warnings        []string            `json:"warnings"`
}

// Location represents a location in the configuration file
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
	Length int    `json:"length"`
}

// Fix represents a suggested fix for an issue
type Fix struct {
	Type        string `json:"type"`         // "replace", "add", "remove"
	Description string `json:"description"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value"`
	Automatic   bool   `json:"automatic"`    // Whether this can be auto-applied
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string   `json:"id"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Component   string   `json:"component"`
	Fix         string   `json:"fix"`
	References  []string `json:"references"`
}

// Bottleneck represents a performance bottleneck
type Bottleneck struct {
	Component   string  `json:"component"`
	Type        string  `json:"type"`        // "cpu", "memory", "disk", "network"
	Impact      string  `json:"impact"`      // "high", "medium", "low"
	Description string  `json:"description"`
	Solution    string  `json:"solution"`
}

// Optimization represents a performance optimization
type Optimization struct {
	Component   string  `json:"component"`
	Type        string  `json:"type"`
	Benefit     string  `json:"benefit"`
	Effort      string  `json:"effort"`      // "low", "medium", "high"
	Description string  `json:"description"`
	Implementation string `json:"implementation"`
}

// ResourceUsage represents predicted resource usage
type ResourceUsage struct {
	CPU        string `json:"cpu"`
	Memory     string `json:"memory"`
	Disk       string `json:"disk"`
	Network    string `json:"network"`
	Confidence float64 `json:"confidence"`
}

// Prediction represents a future prediction
type Prediction struct {
	Type        string    `json:"type"`         // "failure", "capacity", "maintenance"
	Timeframe   string    `json:"timeframe"`    // "1 day", "1 week", "1 month"
	Probability float64   `json:"probability"`
	Description string    `json:"description"`
	Action      string    `json:"action"`
}

// DeprecatedFeature represents a deprecated configuration feature
type DeprecatedFeature struct {
	Feature     string `json:"feature"`
	Version     string `json:"deprecated_in"`
	Replacement string `json:"replacement"`
	Urgency     string `json:"urgency"`       // "low", "medium", "high"
}

// MigrationStep represents a step in a migration path
type MigrationStep struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Automatic   bool   `json:"automatic"`
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		Logger: logger.NewLogger(),
	}
}

// AnalyzeConfiguration performs comprehensive semantic analysis of a NixOS configuration
func (sa *SemanticAnalyzer) AnalyzeConfiguration(ctx context.Context, configPath string, content string) (*AnalysisResult, error) {
	sa.Logger.Info(fmt.Sprintf("Starting semantic analysis for %s with content length %d", configPath, len(content)))

	result := &AnalysisResult{
		ConfigPath: configPath,
		Issues:     []Issue{},
		Suggestions: []Suggestion{},
		Dependencies: []Dependency{},
		Metadata:   make(map[string]interface{}),
	}

	// Parse configuration intent
	intent, err := sa.analyzeIntent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze intent: %w", err)
	}
	result.Intent = intent

	// Detect issues
	issues, err := sa.detectIssues(content)
	if err != nil {
		sa.Logger.Error(fmt.Sprintf("Failed to detect issues: %v", err))
	} else {
		result.Issues = issues
	}

	// Generate suggestions
	suggestions, err := sa.generateSuggestions(content, intent)
	if err != nil {
		sa.Logger.Error(fmt.Sprintf("Failed to generate suggestions: %v", err))
	} else {
		result.Suggestions = suggestions
	}

	// Analyze dependencies
	dependencies, err := sa.analyzeDependencies(content)
	if err != nil {
		sa.Logger.Error(fmt.Sprintf("Failed to analyze dependencies: %v", err))
	} else {
		result.Dependencies = dependencies
	}

	// Perform security analysis
	securityAnalysis, err := sa.performSecurityAnalysis(content)
	if err != nil {
		sa.Logger.Error(fmt.Sprintf("Failed to perform security analysis: %v", err))
	} else {
		result.SecurityAnalysis = securityAnalysis
	}

	// Perform performance analysis
	performanceAnalysis, err := sa.performPerformanceAnalysis(content, intent)
	if err != nil {
		sa.Logger.Error(fmt.Sprintf("Failed to perform performance analysis: %v", err))
	} else {
		result.Performance = performanceAnalysis
	}

	// Perform compatibility analysis
	compatibilityAnalysis, err := sa.performCompatibilityAnalysis(content)
	if err != nil {
		sa.Logger.Error(fmt.Sprintf("Failed to perform compatibility analysis: %v", err))
	} else {
		result.Compatibility = compatibilityAnalysis
	}

	sa.Logger.Info(fmt.Sprintf("Semantic analysis completed: %d issues, %d suggestions, %d dependencies, security score: %.2f, performance score: %.2f",
		len(result.Issues), len(result.Suggestions), len(result.Dependencies), result.SecurityAnalysis.Score, result.Performance.Score))

	return result, nil
}

// analyzeIntent infers the intent and purpose of the configuration
func (sa *SemanticAnalyzer) analyzeIntent(content string) (ConfigIntent, error) {
	intent := ConfigIntent{
		Context: make(map[string]string),
	}

	// Detect purpose based on enabled services and packages
	if sa.containsPatterns(content, []string{"services.xserver", "services.gnome", "services.kde"}) {
		intent.Purpose = "desktop"
		intent.Environment = "personal"
	} else if sa.containsPatterns(content, []string{"services.nginx", "services.apache", "services.postgresql"}) {
		intent.Purpose = "server"
		intent.Environment = "production"
	} else if sa.containsPatterns(content, []string{"development", "devenv", "nix-shell"}) {
		intent.Purpose = "development"
		intent.Environment = "development"
	} else {
		intent.Purpose = "minimal"
		intent.Environment = "unknown"
	}

	// Detect services
	intent.Services = sa.extractServices(content)

	// Detect user profile based on configuration complexity
	complexity := sa.calculateComplexity(content)
	if complexity < 0.3 {
		intent.UserProfile = "beginner"
	} else if complexity < 0.7 {
		intent.UserProfile = "intermediate"
	} else {
		intent.UserProfile = "expert"
	}

	// Detect architecture
	if strings.Contains(content, "x86_64") {
		intent.Architecture = "x86_64"
	} else if strings.Contains(content, "aarch64") {
		intent.Architecture = "aarch64"
	} else {
		intent.Architecture = "unknown"
	}

	// Calculate confidence based on detected patterns
	intent.Confidence = sa.calculateIntentConfidence(intent, content)

	return intent, nil
}

// containsPatterns checks if content contains any of the given patterns
func (sa *SemanticAnalyzer) containsPatterns(content string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// extractServices extracts service names from the configuration
func (sa *SemanticAnalyzer) extractServices(content string) []string {
	var services []string
	
	// Match services.servicename patterns
	serviceRegex := regexp.MustCompile(`services\.(\w+)\s*=`)
	matches := serviceRegex.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			services = append(services, match[1])
		}
	}
	
	return services
}

// calculateComplexity calculates the complexity score of a configuration
func (sa *SemanticAnalyzer) calculateComplexity(content string) float64 {
	factors := map[string]float64{
		"function":     0.1,
		"let":          0.05,
		"import":       0.05,
		"pkgs.":        0.02,
		"services.":    0.03,
		"systemd.":     0.04,
		"users.":       0.02,
		"networking.":  0.03,
		"hardware.":    0.03,
		"boot.":        0.02,
	}
	
	var score float64
	for pattern, weight := range factors {
		count := strings.Count(content, pattern)
		score += float64(count) * weight
	}
	
	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// calculateIntentConfidence calculates confidence in the intent analysis
func (sa *SemanticAnalyzer) calculateIntentConfidence(intent ConfigIntent, content string) float64 {
	confidence := 0.5 // Base confidence
	
	// Increase confidence based on detected patterns
	if len(intent.Services) > 0 {
		confidence += 0.2
	}
	
	if intent.Purpose != "unknown" {
		confidence += 0.2
	}
	
	if intent.Architecture != "unknown" {
		confidence += 0.1
	}
	
	// Normalize to 0-1 range
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// detectIssues identifies potential issues in the configuration
func (sa *SemanticAnalyzer) detectIssues(content string) ([]Issue, error) {
	var issues []Issue
	
	// Check for common security issues
	securityIssues := sa.detectSecurityIssues(content)
	issues = append(issues, securityIssues...)
	
	// Check for performance issues
	performanceIssues := sa.detectPerformanceIssues(content)
	issues = append(issues, performanceIssues...)
	
	// Check for syntax issues
	syntaxIssues := sa.detectSyntaxIssues(content)
	issues = append(issues, syntaxIssues...)
	
	// Check for compatibility issues
	compatibilityIssues := sa.detectCompatibilityIssues(content)
	issues = append(issues, compatibilityIssues...)
	
	return issues, nil
}

// detectSecurityIssues identifies security-related issues
func (sa *SemanticAnalyzer) detectSecurityIssues(content string) []Issue {
	var issues []Issue
	
	// Check for disabled firewall
	if strings.Contains(content, "networking.firewall.enable = false") {
		issues = append(issues, Issue{
			Type:        "security",
			Severity:    "warning",
			Title:       "Firewall Disabled",
			Description: "The firewall is explicitly disabled, which may expose your system to network attacks.",
			Fix: &Fix{
				Type:        "replace",
				Description: "Enable the firewall",
				OldValue:    "networking.firewall.enable = false",
				NewValue:    "networking.firewall.enable = true",
				Automatic:   true,
			},
			References: []string{
				"https://nixos.org/manual/nixos/stable/#sec-firewall",
			},
		})
	}
	
	// Check for root SSH access
	if strings.Contains(content, "PermitRootLogin = \"yes\"") {
		issues = append(issues, Issue{
			Type:        "security",
			Severity:    "critical",
			Title:       "Root SSH Login Enabled",
			Description: "Allowing root login via SSH is a significant security risk.",
			Fix: &Fix{
				Type:        "replace",
				Description: "Disable root SSH login",
				OldValue:    "PermitRootLogin = \"yes\"",
				NewValue:    "PermitRootLogin = \"no\"",
				Automatic:   true,
			},
			References: []string{
				"https://nixos.org/manual/nixos/stable/#sec-ssh",
			},
		})
	}
	
	return issues
}

// detectPerformanceIssues identifies performance-related issues
func (sa *SemanticAnalyzer) detectPerformanceIssues(content string) []Issue {
	var issues []Issue
	
	// Check for missing swap configuration on systems with limited RAM
	if !strings.Contains(content, "swapDevices") && !strings.Contains(content, "zramSwap") {
		issues = append(issues, Issue{
			Type:        "performance",
			Severity:    "info",
			Title:       "No Swap Configuration",
			Description: "No swap configuration detected. Consider adding swap for better memory management.",
			Fix: &Fix{
				Type:        "add",
				Description: "Add zram swap configuration",
				NewValue:    "zramSwap.enable = true;",
				Automatic:   false,
			},
		})
	}
	
	return issues
}

// detectSyntaxIssues identifies syntax-related issues
func (sa *SemanticAnalyzer) detectSyntaxIssues(content string) []Issue {
	var issues []Issue
	
	// Try to parse as Nix expression (simplified check)
	if strings.Count(content, "{") != strings.Count(content, "}") {
		issues = append(issues, Issue{
			Type:        "syntax",
			Severity:    "critical",
			Title:       "Unmatched Braces",
			Description: "The number of opening and closing braces don't match.",
		})
	}
	
	return issues
}

// detectCompatibilityIssues identifies compatibility-related issues
func (sa *SemanticAnalyzer) detectCompatibilityIssues(content string) []Issue {
	var issues []Issue
	
	// Check for deprecated options (simplified examples)
	deprecatedOptions := map[string]string{
		"services.xserver.displayManager.slim":     "services.xserver.displayManager.lightdm or services.xserver.displayManager.gdm",
		"services.mysql":                          "services.mysql.package = pkgs.mariadb",
		"sound.enable":                            "security.rtkit.enable and hardware.pulseaudio.enable or services.pipewire",
	}
	
	for deprecated, replacement := range deprecatedOptions {
		if strings.Contains(content, deprecated) {
			issues = append(issues, Issue{
				Type:        "compatibility",
				Severity:    "warning",
				Title:       "Deprecated Option",
				Description: fmt.Sprintf("Option '%s' is deprecated", deprecated),
				Fix: &Fix{
					Type:        "replace",
					Description: fmt.Sprintf("Replace with %s", replacement),
					OldValue:    deprecated,
					NewValue:    replacement,
					Automatic:   false,
				},
			})
		}
	}
	
	return issues
}

// generateSuggestions generates optimization and improvement suggestions
func (sa *SemanticAnalyzer) generateSuggestions(content string, intent ConfigIntent) ([]Suggestion, error) {
	var suggestions []Suggestion
	
	// Suggest security improvements
	if intent.Purpose == "server" && !strings.Contains(content, "fail2ban") {
		suggestions = append(suggestions, Suggestion{
			Type:        "security",
			Priority:    "high",
			Title:       "Enable Fail2ban",
			Description: "Consider enabling fail2ban for server protection against brute force attacks.",
			After:       "services.fail2ban.enable = true;",
			Rationale:   "Fail2ban helps protect against automated attacks by temporarily blocking IP addresses that show suspicious behavior.",
		})
	}
	
	// Suggest performance improvements
	if intent.Purpose == "desktop" && !strings.Contains(content, "zramSwap") {
		suggestions = append(suggestions, Suggestion{
			Type:        "optimization",
			Priority:    "medium",
			Title:       "Enable ZRAM Swap",
			Description: "Enable ZRAM swap for better memory management on desktop systems.",
			After:       "zramSwap.enable = true;",
			Rationale:   "ZRAM provides compressed swap in memory, improving performance without requiring disk space.",
		})
	}
	
	return suggestions, nil
}

// analyzeDependencies analyzes configuration dependencies
func (sa *SemanticAnalyzer) analyzeDependencies(content string) ([]Dependency, error) {
	var dependencies []Dependency
	
	// Extract package dependencies
	packageRegex := regexp.MustCompile(`pkgs\.(\w+)`)
	matches := packageRegex.FindAllStringSubmatch(content, -1)
	
	packageMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			packageName := match[1]
			if !packageMap[packageName] {
				packageMap[packageName] = true
				dependencies = append(dependencies, Dependency{
					Name:     packageName,
					Type:     "package",
					Required: true,
					Missing:  false, // Would need to check actual system
				})
			}
		}
	}
	
	return dependencies, nil
}

// performSecurityAnalysis performs comprehensive security analysis
func (sa *SemanticAnalyzer) performSecurityAnalysis(content string) (SecurityAnalysis, error) {
	analysis := SecurityAnalysis{
		Vulnerabilities: []Vulnerability{},
		Recommendations: []string{},
		Compliance:      make(map[string]bool),
	}
	
	// Calculate security score based on various factors
	score := 1.0
	
	// Firewall check
	if strings.Contains(content, "networking.firewall.enable = false") {
		score -= 0.3
		analysis.Recommendations = append(analysis.Recommendations, "Enable firewall protection")
	}
	
	// SSH security
	if strings.Contains(content, "PermitRootLogin = \"yes\"") {
		score -= 0.4
		analysis.Vulnerabilities = append(analysis.Vulnerabilities, Vulnerability{
			ID:          "SSH-001",
			Severity:    "high",
			Description: "Root SSH login is enabled",
			Component:   "openssh",
			Fix:         "Set PermitRootLogin to \"no\"",
		})
	}
	
	// User management
	if !strings.Contains(content, "users.users") {
		score -= 0.2
		analysis.Recommendations = append(analysis.Recommendations, "Configure user accounts properly")
	}
	
	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	analysis.Score = score
	
	// Determine risk level
	if score >= 0.8 {
		analysis.Risk = "low"
	} else if score >= 0.6 {
		analysis.Risk = "medium"
	} else if score >= 0.4 {
		analysis.Risk = "high"
	} else {
		analysis.Risk = "critical"
	}
	
	return analysis, nil
}

// performPerformanceAnalysis performs comprehensive performance analysis
func (sa *SemanticAnalyzer) performPerformanceAnalysis(content string, intent ConfigIntent) (PerformanceAnalysis, error) {
	analysis := PerformanceAnalysis{
		Bottlenecks:   []Bottleneck{},
		Optimizations: []Optimization{},
		Predictions:   []Prediction{},
	}
	
	// Calculate performance score
	score := 0.8 // Base score
	
	// Check for performance optimizations
	if strings.Contains(content, "zramSwap.enable = true") {
		score += 0.1
	}
	
	if strings.Contains(content, "services.fstrim.enable = true") {
		score += 0.05
	}
	
	// Check for potential bottlenecks based on intent
	if intent.Purpose == "desktop" {
		if !strings.Contains(content, "hardware.opengl") {
			analysis.Bottlenecks = append(analysis.Bottlenecks, Bottleneck{
				Component:   "graphics",
				Type:        "gpu",
				Impact:      "medium",
				Description: "Graphics acceleration may not be optimally configured",
				Solution:    "Enable hardware.opengl for better graphics performance",
			})
		}
	}
	
	// Suggest optimizations
	if !strings.Contains(content, "zramSwap") {
		analysis.Optimizations = append(analysis.Optimizations, Optimization{
			Component:      "memory",
			Type:          "swap",
			Benefit:       "Improved memory management",
			Effort:        "low",
			Description:   "Enable ZRAM swap for better memory utilization",
			Implementation: "zramSwap.enable = true;",
		})
	}
	
	// Resource usage estimation
	analysis.ResourceUsage = ResourceUsage{
		CPU:        "moderate",
		Memory:     "4-8GB",
		Disk:       "20-50GB",
		Network:    "low",
		Confidence: 0.7,
	}
	
	analysis.Score = score
	return analysis, nil
}

// performCompatibilityAnalysis performs compatibility analysis
func (sa *SemanticAnalyzer) performCompatibilityAnalysis(content string) (CompatibilityAnalysis, error) {
	analysis := CompatibilityAnalysis{
		NixOSVersion:       "24.05",
		Compatibility:      0.9,
		DeprecatedFeatures: []DeprecatedFeature{},
		MigrationPath:      []MigrationStep{},
		Warnings:          []string{},
	}
	
	// Check for deprecated features
	if strings.Contains(content, "sound.enable") {
		analysis.DeprecatedFeatures = append(analysis.DeprecatedFeatures, DeprecatedFeature{
			Feature:     "sound.enable",
			Version:     "22.05",
			Replacement: "hardware.pulseaudio.enable or services.pipewire",
			Urgency:     "medium",
		})
		analysis.Compatibility -= 0.1
	}
	
	return analysis, nil
}