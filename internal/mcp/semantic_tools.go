package mcp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/models/semantic"
	"nix-ai-help/internal/collaboration"
	"nix-ai-help/internal/collaboration/api"
	"nix-ai-help/pkg/logger"
)

// BasicConfigIntent provides a simple implementation of ConfigIntent for MCP
type BasicConfigIntent struct {
	Primary    string   `json:"primary"`
	Secondary  []string `json:"secondary"`
	ConfigType string   `json:"config_type"`
	Confidence float64  `json:"confidence"`
}

func (b *BasicConfigIntent) GetPrimaryIntent() string {
	return b.Primary
}

func (b *BasicConfigIntent) GetSecondaryIntents() []string {
	return b.Secondary
}

func (b *BasicConfigIntent) GetConfigurationType() string {
	return b.ConfigType
}

func (b *BasicConfigIntent) GetConfidence() float64 {
	return b.Confidence
}

// SemanticMCPTools provides AI-powered semantic analysis for the MCP server
type SemanticMCPTools struct {
	semanticEngine   *semantic.NixOSSemanticEngine
	githubLearning   *collaboration.GitHubLearningService
	logger          *logger.Logger
}

// NewSemanticMCPTools creates new semantic analysis tools for MCP
func NewSemanticMCPTools(logger *logger.Logger) *SemanticMCPTools {
	return &SemanticMCPTools{
		semanticEngine: semantic.NewNixOSSemanticEngine(),
		githubLearning: collaboration.NewGitHubLearningService("", 0.7), // High quality threshold
		logger:        logger,
	}
}

// MCPSemanticAnalysisRequest represents a semantic analysis request from MCP
type MCPSemanticAnalysisRequest struct {
	ConfigurationText string `json:"configuration_text"`
	AnalysisType     string `json:"analysis_type"` // "intent", "security", "performance", "patterns"
	Context          string `json:"context,omitempty"`
}

// MCPSemanticAnalysisResponse represents the semantic analysis response
type MCPSemanticAnalysisResponse struct {
	Intent              *semantic.ConfigIntent        `json:"intent,omitempty"`
	SecurityIssues      []api.SecurityIssue          `json:"security_issues,omitempty"`
	PerformanceHints    []api.PerformanceHint        `json:"performance_hints,omitempty"`
	Patterns           []api.ConfigurationPattern    `json:"patterns,omitempty"`
	QualityScore       float64                       `json:"quality_score"`
	Recommendations    []string                      `json:"recommendations"`
	ArchitecturalStyle string                        `json:"architectural_style,omitempty"`
	ComplexityMetrics  *api.ComplexityMetrics        `json:"complexity_metrics,omitempty"`
}

// MCPGitHubSearchRequest represents a GitHub learning request
type MCPGitHubSearchRequest struct {
	Query          string   `json:"query"`
	Keywords       []string `json:"keywords,omitempty"`
	Language       string   `json:"language,omitempty"`
	MaxResults     int      `json:"max_results,omitempty"`
	QualityThreshold float64 `json:"quality_threshold,omitempty"`
}

// MCPGitHubSearchResponse represents GitHub search results
type MCPGitHubSearchResponse struct {
	Repositories    []api.GitHubRepository `json:"repositories"`
	TotalFound      int                   `json:"total_found"`
	AverageQuality  float64               `json:"average_quality"`
	SearchTime      string                `json:"search_time"`
	Recommendations []string              `json:"recommendations"`
}

// MCPPatternExtractionRequest represents pattern extraction request
type MCPPatternExtractionRequest struct {
	Source      string `json:"source"` // "config", "github", "community"
	Content     string `json:"content,omitempty"`
	Repository  string `json:"repository,omitempty"`
	Category    string `json:"category,omitempty"`
}

// MCPPatternExtractionResponse represents extracted patterns
type MCPPatternExtractionResponse struct {
	Patterns    []api.ConfigurationPattern `json:"patterns"`
	Categories  []string                   `json:"categories"`
	Confidence  float64                    `json:"confidence"`
	Insights    []string                   `json:"insights"`
	UsageTips   []string                   `json:"usage_tips"`
}

// AnalyzeConfigurationSemantics performs deep semantic analysis of NixOS configuration
func (s *SemanticMCPTools) AnalyzeConfigurationSemantics(ctx context.Context, req *MCPSemanticAnalysisRequest) (*MCPSemanticAnalysisResponse, error) {
	s.logger.Info(fmt.Sprintf("Performing semantic analysis: type=%s, length=%d", req.AnalysisType, len(req.ConfigurationText)))

	response := &MCPSemanticAnalysisResponse{
		Recommendations: []string{},
	}

	// For now, provide simplified semantic analysis
	// In production, this would use the full semantic engine capabilities
	
	// Perform intent analysis if requested
	if req.AnalysisType == "intent" || req.AnalysisType == "all" {
		intent := s.analyzeConfigurationIntent(req.ConfigurationText)
		response.Intent = &intent
		response.ArchitecturalStyle = s.detectArchitecturalPattern(req.ConfigurationText)
		response.ComplexityMetrics = s.calculateComplexityMetrics(req.ConfigurationText)
	}

	// Perform security analysis
	if req.AnalysisType == "security" || req.AnalysisType == "all" {
		securityIssues := s.analyzeSecurityIssues(req.ConfigurationText)
		response.SecurityIssues = securityIssues
		
		if len(securityIssues) > 0 {
			response.Recommendations = append(response.Recommendations, 
				fmt.Sprintf("Found %d security issues that should be addressed", len(securityIssues)))
		}
	}

	// Perform performance analysis
	if req.AnalysisType == "performance" || req.AnalysisType == "all" {
		perfHints := s.analyzePerformanceHints(req.ConfigurationText)
		response.PerformanceHints = perfHints
		
		if len(perfHints) > 0 {
			response.Recommendations = append(response.Recommendations,
				"Consider implementing the suggested performance optimizations")
		}
	}

	// Extract patterns
	if req.AnalysisType == "patterns" || req.AnalysisType == "all" {
		patterns, err := s.extractPatterns(ctx, req.ConfigurationText)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Pattern extraction failed: %v", err))
		} else {
			response.Patterns = patterns
		}
	}

	// Calculate overall quality score
	response.QualityScore = s.calculateOverallQuality(response)

	// Add general recommendations
	if response.QualityScore < 0.6 {
		response.Recommendations = append(response.Recommendations,
			"Configuration quality is below recommended threshold. Consider refactoring for better maintainability.")
	}

	if len(response.Recommendations) == 0 {
		response.Recommendations = append(response.Recommendations,
			"Configuration looks good! No major issues detected.")
	}

	return response, nil
}

// SearchGitHubConfigurations searches GitHub for relevant NixOS configurations
func (s *SemanticMCPTools) SearchGitHubConfigurations(ctx context.Context, req *MCPGitHubSearchRequest) (*MCPGitHubSearchResponse, error) {
	s.logger.Info(fmt.Sprintf("Searching GitHub for: %s (max: %d)", req.Query, req.MaxResults))

	// Build search query
	query := &api.GitHubSearchQuery{
		Keywords:   req.Keywords,
		Language:   req.Language,
		MaxResults: req.MaxResults,
	}

	// Add default keywords if none provided
	if len(query.Keywords) == 0 {
		query.Keywords = []string{"nixos", req.Query}
	}

	if query.MaxResults == 0 {
		query.MaxResults = 10
	}

	// Perform search
	results, err := s.githubLearning.SearchGitHubConfigurations(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GitHub search failed: %w", err)
	}

	// Filter by quality if threshold specified
	if req.QualityThreshold > 0 {
		filtered, err := s.githubLearning.FilterHighQualityResults(ctx, results)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Quality filtering failed: %v", err))
		} else {
			results = filtered
		}
	}

	response := &MCPGitHubSearchResponse{
		Repositories:   results.Repositories,
		TotalFound:     results.TotalCount,
		AverageQuality: results.QualityScore,
		SearchTime:     results.SearchTime.String(),
		Recommendations: []string{},
	}

	// Add recommendations based on results
	if len(results.Repositories) == 0 {
		response.Recommendations = append(response.Recommendations,
			"No repositories found. Try broader search terms or lower quality threshold.")
	} else if results.QualityScore < 0.5 {
		response.Recommendations = append(response.Recommendations,
			"Average quality is low. Consider using --quality-threshold flag to filter results.")
	} else {
		response.Recommendations = append(response.Recommendations,
			fmt.Sprintf("Found %d high-quality repositories. Click to explore configurations.", len(results.Repositories)))
	}

	return response, nil
}

// ExtractConfigurationPatterns extracts reusable patterns from various sources
func (s *SemanticMCPTools) ExtractConfigurationPatterns(ctx context.Context, req *MCPPatternExtractionRequest) (*MCPPatternExtractionResponse, error) {
	s.logger.Info(fmt.Sprintf("Extracting patterns from: %s", req.Source))

	var patterns []api.ConfigurationPattern
	var err error

	switch req.Source {
	case "config":
		if req.Content == "" {
			return nil, fmt.Errorf("content is required for config source")
		}
		patterns, err = s.extractPatterns(ctx, req.Content)
		
	case "github":
		if req.Repository == "" {
			return nil, fmt.Errorf("repository is required for github source")
		}
		patterns, err = s.extractPatternsFromGitHub(ctx, req.Repository)
		
	default:
		return nil, fmt.Errorf("unsupported source: %s", req.Source)
	}

	if err != nil {
		return nil, fmt.Errorf("pattern extraction failed: %w", err)
	}

	// Group patterns by category
	categories := make(map[string]bool)
	for _, pattern := range patterns {
		categories[pattern.Category] = true
	}

	categoryList := make([]string, 0, len(categories))
	for category := range categories {
		categoryList = append(categoryList, category)
	}

	// Calculate confidence based on pattern frequency and quality
	confidence := 0.0
	if len(patterns) > 0 {
		totalSuccess := 0.0
		for _, pattern := range patterns {
			totalSuccess += pattern.Success
		}
		confidence = totalSuccess / float64(len(patterns))
	}

	response := &MCPPatternExtractionResponse{
		Patterns:   patterns,
		Categories: categoryList,
		Confidence: confidence,
		Insights:   s.generatePatternInsights(patterns),
		UsageTips:  s.generateUsageTips(patterns),
	}

	return response, nil
}

// Helper methods

func (s *SemanticMCPTools) analyzeSecurityIssues(content string) []api.SecurityIssue {
	var issues []api.SecurityIssue

	// Check for common security issues
	securityChecks := map[string]struct {
		severity    string
		description string
		suggestion  string
	}{
		"permitRootLogin.*yes": {
			severity:    "high",
			description: "SSH root login is enabled",
			suggestion:  "Disable root login and use sudo instead",
		},
		"passwordAuthentication.*true": {
			severity:    "medium", 
			description: "Password authentication enabled for SSH",
			suggestion:  "Consider using key-based authentication only",
		},
		"services\\.openssh\\.enable.*=.*true.*services\\.openssh\\.passwordAuthentication.*=.*true": {
			severity:    "medium",
			description: "SSH enabled with password authentication",
			suggestion:  "Use key-based authentication for better security",
		},
	}

	for pattern, check := range securityChecks {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			issues = append(issues, api.SecurityIssue{
				ID:          fmt.Sprintf("sec_%d", len(issues)+1),
				Type:        "configuration",
				Severity:    check.severity,
				Description: check.description,
				Suggestion:  check.suggestion,
				References:  []string{"https://nixos.org/manual/nixos/stable/index.html#sec-ssh"},
			})
		}
	}

	return issues
}

func (s *SemanticMCPTools) analyzePerformanceHints(content string) []api.PerformanceHint {
	var hints []api.PerformanceHint

	// Check for performance optimizations
	if matched, _ := regexp.MatchString(`boot\.loader\.grub`, content); matched {
		hints = append(hints, api.PerformanceHint{
			Type:        "bootloader",
			Description: "Using GRUB bootloader",
			Impact:      "medium",
			Suggestion:  "Consider systemd-boot for faster boot times on UEFI systems",
		})
	}

	if matched, _ := regexp.MatchString(`services\.xserver\.enable.*=.*true`, content); matched {
		hints = append(hints, api.PerformanceHint{
			Type:        "display_server",
			Description: "X11 display server enabled",
			Impact:      "low",
			Suggestion:  "Consider Wayland for better performance on modern hardware",
		})
	}

	return hints
}

func (s *SemanticMCPTools) extractPatterns(ctx context.Context, content string) ([]api.ConfigurationPattern, error) {
	// Use the GitHub learning service pattern extraction
	externalContent := &api.ExternalContent{
		Source:  "mcp",
		Type:    "nix_config",
		Content: content,
	}

	patterns, err := s.githubLearning.ExtractConfigurationPatterns(ctx, externalContent)
	if err != nil {
		return nil, err
	}

	return patterns.Patterns, nil
}

func (s *SemanticMCPTools) extractPatternsFromGitHub(ctx context.Context, repository string) ([]api.ConfigurationPattern, error) {
	// This would integrate with GitHub API to analyze a specific repository
	// For now, return empty patterns as this requires more complex implementation
	s.logger.Info(fmt.Sprintf("GitHub repository pattern extraction not yet implemented for: %s", repository))
	return []api.ConfigurationPattern{}, nil
}

func (s *SemanticMCPTools) calculateOverallQuality(response *MCPSemanticAnalysisResponse) float64 {
	score := 0.8 // Base score

	// Penalize for security issues
	if len(response.SecurityIssues) > 0 {
		highSeverityCount := 0
		for _, issue := range response.SecurityIssues {
			if issue.Severity == "high" || issue.Severity == "critical" {
				highSeverityCount++
			}
		}
		score -= float64(highSeverityCount) * 0.2
		score -= float64(len(response.SecurityIssues)-highSeverityCount) * 0.1
	}

	// Bonus for patterns found
	if len(response.Patterns) > 0 {
		score += float64(len(response.Patterns)) * 0.05
	}

	// Bonus for performance hints (means we found optimizations)
	if len(response.PerformanceHints) > 0 {
		score += 0.1
	}

	// Ensure score is in valid range
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

func (s *SemanticMCPTools) generatePatternInsights(patterns []api.ConfigurationPattern) []string {
	if len(patterns) == 0 {
		return []string{"No patterns detected in the configuration"}
	}

	insights := []string{
		fmt.Sprintf("Detected %d configuration patterns", len(patterns)),
	}

	// Categorize patterns
	categoryCount := make(map[string]int)
	for _, pattern := range patterns {
		categoryCount[pattern.Category]++
	}

	for category, count := range categoryCount {
		insights = append(insights, fmt.Sprintf("Found %d %s patterns", count, category))
	}

	return insights
}

func (s *SemanticMCPTools) generateUsageTips(patterns []api.ConfigurationPattern) []string {
	if len(patterns) == 0 {
		return []string{"Add more configuration patterns to get usage tips"}
	}

	tips := []string{
		"These patterns can be reused in other NixOS configurations",
		"Consider extracting common patterns into separate modules",
	}

	// Add specific tips based on pattern categories
	categoryCount := make(map[string]int)
	for _, pattern := range patterns {
		categoryCount[pattern.Category]++
	}

	if categoryCount["systemd_services"] > 0 {
		tips = append(tips, "Consider grouping related services for better organization")
	}

	if categoryCount["package_install"] > 0 {
		tips = append(tips, "Large package lists can be moved to separate files for maintainability")
	}

	return tips
}

// Simplified semantic analysis methods

func (s *SemanticMCPTools) analyzeConfigurationIntent(content string) semantic.ConfigIntent {
	// Create a basic intent analysis
	intent := semantic.ConfigIntent{
		Purpose:      s.detectPrimaryIntent(content),
		Services:     []string{},
		Environment:  s.detectConfigurationType(content),
		UserProfile:  "intermediate",
		Architecture: "x86_64",
		Confidence:   0.8,
		Context:      make(map[string]string),
	}
	
	return intent
}

func (s *SemanticMCPTools) detectPrimaryIntent(content string) string {
	if strings.Contains(content, "services.") {
		return "service_configuration"
	} else if strings.Contains(content, "environment.systemPackages") {
		return "package_management"
	} else if strings.Contains(content, "users.") {
		return "user_management"
	} else if strings.Contains(content, "hardware.") {
		return "hardware_configuration"
	} else {
		return "system_configuration"
	}
}

func (s *SemanticMCPTools) detectConfigurationType(content string) string {
	if strings.Contains(content, "flake.nix") {
		return "flake"
	} else if strings.Contains(content, "home.nix") {
		return "home_manager"
	} else {
		return "nixos_configuration"
	}
}

func (s *SemanticMCPTools) detectArchitecturalPattern(content string) string {
	if strings.Count(content, "imports") > 0 {
		return "modular"
	} else if strings.Count(content, "services.") > 5 {
		return "service_oriented"
	} else if strings.Count(content, "environment.") > 3 {
		return "environment_focused"
	} else {
		return "monolithic"
	}
}

func (s *SemanticMCPTools) calculateComplexityMetrics(content string) *api.ComplexityMetrics {
	lines := strings.Split(content, "\n")
	nonEmptyLines := 0
	maxIndent := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			nonEmptyLines++
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			if indent > maxIndent {
				maxIndent = indent
			}
		}
	}
	
	return &api.ComplexityMetrics{
		Lines:         len(lines),
		Functions:     strings.Count(content, "="),
		CyclomaticComplexity: maxIndent / 2,
		NestingDepth:  maxIndent / 4,
		Score:         float64(nonEmptyLines + maxIndent/2),
	}
}