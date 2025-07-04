package collaboration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/collaboration/api"
	"nix-ai-help/internal/community"
	"nix-ai-help/pkg/logger"
)

// GitHubLearningService implements ExternalLearningAPI for GitHub integration
type GitHubLearningService struct {
	githubClient     *community.GitHubClient
	qualityThreshold float64
	logger           *logger.Logger
}

// NewGitHubLearningService creates a new GitHub learning service
func NewGitHubLearningService(apiToken string, qualityThreshold float64) *GitHubLearningService {
	return &GitHubLearningService{
		githubClient:     community.NewGitHubClient(apiToken),
		qualityThreshold: qualityThreshold,
		logger:           logger.NewLoggerWithLevel("info"),
	}
}

// SearchGitHubConfigurations searches for NixOS configurations on GitHub
func (g *GitHubLearningService) SearchGitHubConfigurations(ctx context.Context, query *api.GitHubSearchQuery) (*api.GitHubSearchResults, error) {
	startTime := time.Now()

	// Build search query string
	queryStr := g.buildSearchQuery(query)
	
	// Perform search
	repos, err := g.githubClient.SearchRepositories(queryStr, query.MaxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to search GitHub repositories: %w", err)
	}

	// Convert to API format and calculate quality scores
	results := &api.GitHubSearchResults{
		Query:       query,
		Repositories: make([]api.GitHubRepository, 0, len(repos)),
		TotalCount:  len(repos),
		SearchTime:  time.Since(startTime),
		GeneratedAt: time.Now(),
	}

	totalQualityScore := 0.0
	for _, repo := range repos {
		apiRepo := g.convertToAPIRepository(repo)
		apiRepo.QualityScore = g.calculateRepositoryQuality(apiRepo)
		results.Repositories = append(results.Repositories, apiRepo)
		totalQualityScore += apiRepo.QualityScore
	}

	if len(results.Repositories) > 0 {
		results.QualityScore = totalQualityScore / float64(len(results.Repositories))
	}

	g.logger.Info(fmt.Sprintf("GitHub search completed: found %d repositories with avg quality %.2f",
		len(results.Repositories), results.QualityScore))

	return results, nil
}

// AnalyzeGitHubRepository performs deep analysis of a repository
func (g *GitHubLearningService) AnalyzeGitHubRepository(ctx context.Context, repo *api.GitHubRepository) (*api.RepositoryAnalysis, error) {
	analysis := &api.RepositoryAnalysis{
		Repository:      repo,
		ConfigFiles:     []api.ConfigFileAnalysis{},
		Patterns:        []api.ConfigurationPattern{},
		SecurityIssues:  []api.SecurityIssue{},
		BestPractices:   []api.BestPractice{},
		QualityMetrics:  api.QualityMetrics{},
		Recommendations: []string{},
		AnalyzedAt:      time.Now(),
	}

	// Analyze common NixOS configuration files
	nixosFiles := []string{
		"configuration.nix",
		"hardware-configuration.nix", 
		"flake.nix",
		"home.nix",
		"default.nix",
	}

	for _, fileName := range nixosFiles {
		content, err := g.githubClient.GetFileContent(repo.Owner.Login, repo.Name, fileName)
		if err != nil {
			continue // File might not exist
		}

		fileAnalysis := g.analyzeConfigFile(fileName, content)
		analysis.ConfigFiles = append(analysis.ConfigFiles, fileAnalysis)
	}

	// Extract patterns from config files
	patterns := g.extractPatternsFromAnalysis(analysis.ConfigFiles)
	analysis.Patterns = patterns

	// Security analysis
	securityIssues := g.performSecurityAnalysis(analysis.ConfigFiles)
	analysis.SecurityIssues = securityIssues

	// Best practices analysis
	bestPractices := g.analyzeBestPractices(analysis.ConfigFiles)
	analysis.BestPractices = bestPractices

	// Calculate quality metrics
	analysis.QualityMetrics = g.calculateQualityMetrics(analysis)

	// Generate recommendations
	analysis.Recommendations = g.generateRecommendations(analysis)

	return analysis, nil
}

// ValidateExternalContent validates content from external sources
func (g *GitHubLearningService) ValidateExternalContent(ctx context.Context, content *api.ExternalContent) (*api.ContentValidation, error) {
	validation := &api.ContentValidation{
		Valid:       true,
		Issues:      []api.ValidationIssue{},
		Suggestions: []string{},
		SafetyLevel: "safe",
		ValidatedAt: time.Now(),
	}

	// Check for malicious patterns
	maliciousPatterns := []string{
		`rm\s+-rf\s+/`,
		`wget.*\|\s*sh`,
		`curl.*\|\s*bash`,
		`eval\s*\$\(`,
		`exec\s+/bin/sh`,
	}

	for _, pattern := range maliciousPatterns {
		if matched, _ := regexp.MatchString(pattern, content.Content); matched {
			validation.Valid = false
			validation.SafetyLevel = "dangerous"
			validation.Issues = append(validation.Issues, api.ValidationIssue{
				Type:        "security",
				Severity:    "critical",
				Description: fmt.Sprintf("Potentially malicious pattern detected: %s", pattern),
				Location:    "content",
			})
		}
	}

	// Check for quality indicators
	qualityChecks := g.performQualityChecks(content.Content)
	validation.QualityScore = qualityChecks.score
	validation.Issues = append(validation.Issues, qualityChecks.issues...)
	validation.Suggestions = append(validation.Suggestions, qualityChecks.suggestions...)

	return validation, nil
}

// FilterHighQualityResults filters results based on quality thresholds
func (g *GitHubLearningService) FilterHighQualityResults(ctx context.Context, results *api.GitHubSearchResults) (*api.GitHubSearchResults, error) {
	filtered := &api.GitHubSearchResults{
		Query:        results.Query,
		Repositories: []api.GitHubRepository{},
		SearchTime:   results.SearchTime,
		GeneratedAt:  time.Now(),
	}

	for _, repo := range results.Repositories {
		if repo.QualityScore >= g.qualityThreshold {
			filtered.Repositories = append(filtered.Repositories, repo)
		}
	}

	filtered.TotalCount = len(filtered.Repositories)

	// Recalculate average quality score
	totalQuality := 0.0
	for _, repo := range filtered.Repositories {
		totalQuality += repo.QualityScore
	}
	if len(filtered.Repositories) > 0 {
		filtered.QualityScore = totalQuality / float64(len(filtered.Repositories))
	}

	g.logger.Info(fmt.Sprintf("Filtered repositories: %d -> %d (threshold: %.2f)",
		len(results.Repositories), len(filtered.Repositories), g.qualityThreshold))

	return filtered, nil
}

// ExtractConfigurationPatterns extracts patterns from content
func (g *GitHubLearningService) ExtractConfigurationPatterns(ctx context.Context, content *api.ExternalContent) (*api.ConfigurationPatterns, error) {
	patterns := &api.ConfigurationPatterns{
		Patterns:    []api.ConfigurationPattern{},
		Categories:  []string{},
		Confidence:  0.0,
		ExtractedAt: time.Now(),
	}

	// Define pattern matchers for common NixOS patterns
	patternMatchers := map[string]*regexp.Regexp{
		"systemd_service": regexp.MustCompile(`systemd\.services\.(\w+)\s*=\s*{[^}]*}`),
		"package_install": regexp.MustCompile(`environment\.systemPackages\s*=.*?with\s+pkgs;\s*\[(.*?)\]`),
		"user_config":     regexp.MustCompile(`users\.users\.(\w+)\s*=\s*{[^}]*}`),
		"module_import":   regexp.MustCompile(`imports\s*=\s*\[(.*?)\]`),
		"option_set":      regexp.MustCompile(`(\w+\.\w+(?:\.\w+)*)\s*=\s*([^;]+);`),
	}

	categoryCounts := make(map[string]int)

	for category, matcher := range patternMatchers {
		matches := matcher.FindAllStringSubmatch(content.Content, -1)
		for i, match := range matches {
			if len(match) > 0 {
				pattern := api.ConfigurationPattern{
					ID:          fmt.Sprintf("%s_%d", category, i),
					Name:        fmt.Sprintf("%s pattern", strings.ReplaceAll(category, "_", " ")),
					Category:    category,
					Description: g.getPatternDescription(category),
					Pattern:     match[0],
					UseCase:     strings.Join(g.getPatternUseCases(category), ","),
					Frequency:   1,
					Success:     0.8,
					Created:     time.Now(),
					Metadata:    map[string]interface{}{"source": content.Source},
				}
				patterns.Patterns = append(patterns.Patterns, pattern)
				categoryCounts[category]++
			}
		}
	}

	// Calculate categories and confidence
	for category := range categoryCounts {
		patterns.Categories = append(patterns.Categories, category)
	}

	if len(patterns.Patterns) > 0 {
		patterns.Confidence = float64(len(patterns.Patterns)) / 10.0 // Simple confidence metric
		if patterns.Confidence > 1.0 {
			patterns.Confidence = 1.0
		}
	}

	return patterns, nil
}

// AnonymizeGitHubData anonymizes GitHub data for privacy
func (g *GitHubLearningService) AnonymizeGitHubData(ctx context.Context, data *api.GitHubData) (*api.AnonymizedData, error) {
	anonymized := &api.AnonymizedData{
		Anonymized:  []string{},
		Method:      "hash_substitution",
		ProcessedAt: time.Now(),
	}

	// Create anonymized copy
	anonymizedRepos := make([]api.GitHubRepository, len(data.Repositories))
	for i, repo := range data.Repositories {
		anonRepo := repo
		
		// Anonymize sensitive fields
		anonRepo.Owner.Login = g.hashString(repo.Owner.Login)
		anonRepo.FullName = g.hashString(repo.FullName)
		anonRepo.URL = "https://github.com/anonymous/repo"
		anonRepo.CloneURL = "https://github.com/anonymous/repo.git"
		
		anonymizedRepos[i] = anonRepo
		anonymized.Anonymized = append(anonymized.Anonymized, "owner.login", "full_name", "url", "clone_url")
	}

	anonymized.Data = map[string]interface{}{
		"repositories": anonymizedRepos,
		"users":        []api.GitHubUser{}, // Remove user data
		"metadata":     data.Metadata,
	}

	return anonymized, nil
}

// ApplyPrivacyFilters applies privacy filters to content
func (g *GitHubLearningService) ApplyPrivacyFilters(ctx context.Context, content *api.ExternalContent) (*api.ExternalContent, error) {
	filtered := *content
	
	// Remove sensitive patterns
	sensitivePatterns := []string{
		`password\s*=\s*"[^"]*"`,
		`token\s*=\s*"[^"]*"`,
		`key\s*=\s*"[^"]*"`,
		`secret\s*=\s*"[^"]*"`,
		`api[_-]?key\s*=\s*"[^"]*"`,
	}

	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		filtered.Content = re.ReplaceAllString(filtered.Content, `$1 = "<REDACTED>"`)
	}

	// Add privacy metadata
	if filtered.Metadata == nil {
		filtered.Metadata = make(map[string]interface{})
	}
	filtered.Metadata["privacy_filtered"] = true
	filtered.Metadata["filtered_at"] = time.Now()

	return &filtered, nil
}

// Helper methods

func (g *GitHubLearningService) buildSearchQuery(query *api.GitHubSearchQuery) string {
	var parts []string
	
	// Add keywords
	for _, keyword := range query.Keywords {
		parts = append(parts, keyword)
	}
	
	// Add language filter
	if query.Language != "" {
		parts = append(parts, fmt.Sprintf("language:%s", query.Language))
	}
	
	// Add file type filter
	if query.FileType != "" {
		parts = append(parts, fmt.Sprintf("extension:%s", query.FileType))
	}
	
	// Add stars filter
	if query.Stars != nil {
		if query.Stars.Min != nil && query.Stars.Max != nil {
			parts = append(parts, fmt.Sprintf("stars:%d..%d", *query.Stars.Min, *query.Stars.Max))
		} else if query.Stars.Min != nil {
			parts = append(parts, fmt.Sprintf("stars:>=%d", *query.Stars.Min))
		} else if query.Stars.Max != nil {
			parts = append(parts, fmt.Sprintf("stars:<=%d", *query.Stars.Max))
		}
	}
	
	return strings.Join(parts, " ")
}

func (g *GitHubLearningService) convertToAPIRepository(repo community.GitHubRepository) api.GitHubRepository {
	return api.GitHubRepository{
		ID:          fmt.Sprintf("%d", repo.ID),
		Name:        repo.Name,
		FullName:    repo.FullName,
		Description: repo.Description,
		URL:         repo.URL,
		CloneURL:    repo.CloneURL,
		Language:    repo.Language,
		Stars:       repo.Stars,
		Forks:       repo.Forks,
		Topics:      repo.Topics,
		CreatedAt:   repo.CreatedAt,
		UpdatedAt:   repo.UpdatedAt,
		Owner: api.GitHubUser{
			Login:     repo.Owner.Login,
			ID:        repo.Owner.ID,
			AvatarURL: repo.Owner.AvatarURL,
			URL:       repo.Owner.URL,
			Type:      "User",
		},
		Metadata: make(map[string]interface{}),
	}
}

func (g *GitHubLearningService) calculateRepositoryQuality(repo api.GitHubRepository) float64 {
	score := 0.0
	
	// Stars factor (0-0.3)
	starScore := float64(repo.Stars) / (float64(repo.Stars) + 100.0)
	score += starScore * 0.3
	
	// Activity factor (0-0.2)
	daysSinceUpdate := time.Since(repo.UpdatedAt).Hours() / 24
	activityScore := 1.0 / (1.0 + daysSinceUpdate/30.0) // Recent activity is better
	score += activityScore * 0.2
	
	// Description factor (0-0.1)
	if len(repo.Description) > 20 {
		score += 0.1
	}
	
	// Topic factor (0-0.2)
	nixosTopics := 0
	for _, topic := range repo.Topics {
		if strings.Contains(strings.ToLower(topic), "nix") {
			nixosTopics++
		}
	}
	topicScore := float64(nixosTopics) / 5.0
	if topicScore > 1.0 {
		topicScore = 1.0
	}
	score += topicScore * 0.2
	
	// Fork factor (0-0.2)
	forkScore := float64(repo.Forks) / (float64(repo.Forks) + 50.0)
	score += forkScore * 0.2
	
	return score
}

func (g *GitHubLearningService) analyzeConfigFile(fileName, content string) api.ConfigFileAnalysis {
	return api.ConfigFileAnalysis{
		File: &api.GitHubFile{
			Name:    fileName,
			Type:    "file",
			Content: content,
		},
		FileType:        g.detectFileType(fileName),
		Complexity:      g.calculateComplexity(content),
		Dependencies:    g.extractDependencies(content),
		Services:        g.extractServices(content),
		SecurityConfig:  g.analyzeSecurityConfig(content),
		PerformanceHints: g.analyzePerformance(content),
		Documentation:   g.analyzeDocumentation(content),
		Maintainability: g.calculateMaintainability(content),
	}
}

func (g *GitHubLearningService) detectFileType(fileName string) string {
	if strings.HasSuffix(fileName, ".nix") {
		return "nix"
	}
	return "unknown"
}

func (g *GitHubLearningService) calculateComplexity(content string) api.ComplexityMetrics {
	lines := strings.Split(content, "\n")
	nonEmptyLines := 0
	indentLevels := make(map[int]int)
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			nonEmptyLines++
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			indentLevels[indent]++
		}
	}
	
	maxIndent := 0
	for indent := range indentLevels {
		if indent > maxIndent {
			maxIndent = indent
		}
	}
	
	return api.ComplexityMetrics{
		Lines:         len(lines),
		Functions:     strings.Count(content, "="),
		CyclomaticComplexity: maxIndent + 1,
		NestingDepth:  maxIndent / 2,
		Score:         float64(nonEmptyLines + maxIndent),
	}
}

func (g *GitHubLearningService) extractDependencies(content string) []string {
	var deps []string
	
	// Extract package dependencies
	packageRegex := regexp.MustCompile(`pkgs\.(\w+)`)
	matches := packageRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			deps = append(deps, match[1])
		}
	}
	
	return g.uniqueStrings(deps)
}

func (g *GitHubLearningService) extractServices(content string) []string {
	var services []string
	
	// Extract systemd services
	serviceRegex := regexp.MustCompile(`services\.(\w+)`)
	matches := serviceRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			services = append(services, match[1])
		}
	}
	
	return g.uniqueStrings(services)
}

func (g *GitHubLearningService) analyzeSecurityConfig(content string) []api.SecurityConfig {
	var configs []api.SecurityConfig
	
	// Check for common security configurations
	securityChecks := map[string]string{
		"firewall.enable":          "Firewall configuration",
		"openssh.permitRootLogin":  "SSH root login setting",
		"services.fail2ban.enable": "Fail2ban intrusion detection",
		"security.sudo":            "Sudo configuration",
	}
	
	for pattern, description := range securityChecks {
		if strings.Contains(content, pattern) {
			configs = append(configs, api.SecurityConfig{
				Type:        "security_setting",
				Description: description,
				Severity:    "medium",
				Compliant:   true,
			})
		}
	}
	
	return configs
}

func (g *GitHubLearningService) analyzePerformance(content string) []api.PerformanceHint {
	var hints []api.PerformanceHint
	
	// Check for performance-related configurations
	if strings.Contains(content, "boot.loader.grub") {
		hints = append(hints, api.PerformanceHint{
			Type:        "bootloader",
			Description: "GRUB bootloader detected",
			Impact:      "low",
			Suggestion:  "Consider systemd-boot for faster boot times",
		})
	}
	
	return hints
}

func (g *GitHubLearningService) analyzeDocumentation(content string) api.DocumentationLevel {
	lines := strings.Split(content, "\n")
	commentCount := 0
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			commentCount++
		}
	}
	
	score := float64(commentCount) / float64(len(lines))
	if score > 1.0 {
		score = 1.0
	}
	
	return api.DocumentationLevel{
		Score:    score,
		Comments: commentCount,
		README:   false, // Would need to check for README file
		Examples: strings.Contains(content, "example"),
	}
}

func (g *GitHubLearningService) calculateMaintainability(content string) float64 {
	complexity := g.calculateComplexity(content)
	documentation := g.analyzeDocumentation(content)
	
	// Simple maintainability score based on complexity and documentation
	complexityScore := 1.0 - (complexity.Score / 1000.0) // Lower complexity is better
	if complexityScore < 0 {
		complexityScore = 0
	}
	
	return (complexityScore + documentation.Score) / 2.0
}

func (g *GitHubLearningService) extractPatternsFromAnalysis(files []api.ConfigFileAnalysis) []api.ConfigurationPattern {
	var patterns []api.ConfigurationPattern
	
	for _, file := range files {
		// Extract service patterns
		for _, service := range file.Services {
			patterns = append(patterns, api.ConfigurationPattern{
				ID:          fmt.Sprintf("service_%s", service),
				Name:        fmt.Sprintf("%s service configuration", service),
				Category:    "systemd_services",
				Description: fmt.Sprintf("Configuration pattern for %s service", service),
				Pattern:     fmt.Sprintf("services.%s", service),
				UseCase:     "service_management",
				Frequency:   1,
				Success:     0.9,
				Created:     time.Now(),
				Metadata:    map[string]interface{}{"service": service},
			})
		}
	}
	
	return patterns
}

func (g *GitHubLearningService) performSecurityAnalysis(files []api.ConfigFileAnalysis) []api.SecurityIssue {
	var issues []api.SecurityIssue
	
	for _, file := range files {
		if file.File != nil && strings.Contains(file.File.Content, "permitRootLogin = \"yes\"") {
			issues = append(issues, api.SecurityIssue{
				ID:          "ssh_root_login",
				Type:        "configuration",
				Severity:    "high",
				Description: "SSH root login is enabled",
				File:        file.File.Name,
				Suggestion:  "Disable root login and use sudo instead",
				References:  []string{"https://nixos.org/manual/nixos/stable/index.html#sec-ssh"},
				DetectedAt:  time.Now(),
			})
		}
	}
	
	return issues
}

func (g *GitHubLearningService) analyzeBestPractices(files []api.ConfigFileAnalysis) []api.BestPractice {
	var practices []api.BestPractice
	
	for _, file := range files {
		if file.File != nil && strings.Contains(file.File.Content, "imports =") {
			practices = append(practices, api.BestPractice{
				ID:          "modular_config",
				Category:    "organization",
				Description: "Uses modular configuration with imports",
				Applied:     true,
				Benefit:     "Improves maintainability and organization",
				References:  []string{"https://nixos.org/manual/nixos/stable/index.html#sec-modularity"},
			})
		}
	}
	
	return practices
}

func (g *GitHubLearningService) calculateQualityMetrics(analysis *api.RepositoryAnalysis) api.QualityMetrics {
	var totalMaintainability, totalDocumentation float64
	fileCount := float64(len(analysis.ConfigFiles))
	
	if fileCount == 0 {
		return api.QualityMetrics{}
	}
	
	for _, file := range analysis.ConfigFiles {
		totalMaintainability += file.Maintainability
		totalDocumentation += file.Documentation.Score
	}
	
	avgMaintainability := totalMaintainability / fileCount
	avgDocumentation := totalDocumentation / fileCount
	
	// Security score based on issues (fewer issues = higher score)
	securityScore := 1.0
	if len(analysis.SecurityIssues) > 0 {
		securityScore = 1.0 / (1.0 + float64(len(analysis.SecurityIssues))*0.2)
	}
	
	// Overall score is average of all metrics
	overallScore := (avgMaintainability + securityScore + avgDocumentation) / 3.0
	
	return api.QualityMetrics{
		OverallScore:    overallScore,
		Maintainability: avgMaintainability,
		Security:        securityScore,
		Documentation:   avgDocumentation,
		Performance:     0.8, // Default performance score
		TestCoverage:    0.0, // Would need to analyze test files
		CodeComplexity:  0.5, // Average complexity
	}
}

func (g *GitHubLearningService) generateRecommendations(analysis *api.RepositoryAnalysis) []string {
	var recommendations []string
	
	if analysis.QualityMetrics.Documentation < 0.3 {
		recommendations = append(recommendations, "Add more comments and documentation to improve maintainability")
	}
	
	if len(analysis.SecurityIssues) > 0 {
		recommendations = append(recommendations, "Address security issues to improve system security")
	}
	
	if len(analysis.BestPractices) < 3 {
		recommendations = append(recommendations, "Consider implementing more NixOS best practices")
	}
	
	return recommendations
}

func (g *GitHubLearningService) performQualityChecks(content string) struct {
	score       float64
	issues      []api.ValidationIssue
	suggestions []string
} {
	result := struct {
		score       float64
		issues      []api.ValidationIssue
		suggestions []string
	}{
		score:       0.8, // Default score
		issues:      []api.ValidationIssue{},
		suggestions: []string{},
	}
	
	// Check for basic quality indicators
	if len(content) < 100 {
		result.score -= 0.2
		result.issues = append(result.issues, api.ValidationIssue{
			Type:        "quality",
			Severity:    "low",
			Description: "Content is very short",
			Location:    "content",
		})
	}
	
	if !strings.Contains(content, "#") {
		result.score -= 0.1
		result.suggestions = append(result.suggestions, "Add comments to improve readability")
	}
	
	return result
}

func (g *GitHubLearningService) getPatternDescription(category string) string {
	descriptions := map[string]string{
		"systemd_service": "Systemd service configuration pattern",
		"package_install": "Package installation pattern",
		"user_config":     "User configuration pattern",
		"module_import":   "Module import pattern",
		"option_set":      "Option setting pattern",
	}
	
	if desc, exists := descriptions[category]; exists {
		return desc
	}
	return "Configuration pattern"
}

func (g *GitHubLearningService) getPatternUseCases(category string) []string {
	useCases := map[string][]string{
		"systemd_service": {"service_management", "system_configuration"},
		"package_install": {"package_management", "environment_setup"},
		"user_config":     {"user_management", "access_control"},
		"module_import":   {"modular_configuration", "code_organization"},
		"option_set":      {"system_configuration", "option_management"},
	}
	
	if cases, exists := useCases[category]; exists {
		return cases
	}
	return []string{"general_configuration"}
}

func (g *GitHubLearningService) hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:12] // Use first 12 characters
}

func (g *GitHubLearningService) uniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	sort.Strings(result)
	return result
}