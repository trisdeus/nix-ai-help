package errors

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// EnhancedErrorCommunicator provides AI-powered contextual error messages
type EnhancedErrorCommunicator struct {
	messageGenerator  *UserFriendlyMessageGenerator
	contextAnalyzer   *ErrorContextAnalyzer
	solutionSuggester *SolutionSuggester
	patternMatcher    *ErrorPatternMatcher
}

// ErrorContextAnalyzer analyzes error context to provide better messages
type ErrorContextAnalyzer struct {
	commandPatterns   map[string][]string
	systemInfo        *SystemInfo
	recentCommands    []string
	configState       *ConfigState
}

// SolutionSuggester provides intelligent solution suggestions
type SolutionSuggester struct {
	knowledgeBase    *ErrorKnowledgeBase
	adaptiveFilters  map[string][]string
	successPatterns  map[string]string
}

// ErrorPatternMatcher matches error patterns for better categorization
type ErrorPatternMatcher struct {
	nixPatterns     []*EnhancedErrorPattern
	providerPatterns []*EnhancedErrorPattern
	systemPatterns   []*EnhancedErrorPattern
	customPatterns   []*EnhancedErrorPattern
}

// EnhancedErrorPattern defines a pattern for matching specific error types (renamed to avoid conflict)
type EnhancedErrorPattern struct {
	Name        string            `json:"name"`
	Pattern     *regexp.Regexp    `json:"-"`
	PatternStr  string            `json:"pattern"`
	Category    ErrorCategory     `json:"category"`
	Severity    ErrorSeverity     `json:"severity"`
	Context     map[string]string `json:"context"`
	Solutions   []string          `json:"solutions"`
	DocsLinks   []string          `json:"docs_links"`
	Conditions  []string          `json:"conditions"`
}

// SystemInfo contains relevant system information for error context
type SystemInfo struct {
	NixOSVersion     string    `json:"nixos_version"`
	Architecture     string    `json:"architecture"`
	AvailableMemory  int64     `json:"available_memory"`
	DiskSpace        int64     `json:"disk_space"`
	ActiveProviders  []string  `json:"active_providers"`
	RunningServices  []string  `json:"running_services"`
	LastUpdated      time.Time `json:"last_updated"`
}

// ConfigState tracks configuration state for context
type ConfigState struct {
	ConfigPath       string            `json:"config_path"`
	LastModified     time.Time         `json:"last_modified"`
	ActiveProviders  map[string]bool   `json:"active_providers"`
	CacheEnabled     bool              `json:"cache_enabled"`
	DebugMode        bool              `json:"debug_mode"`
	CustomSettings   map[string]string `json:"custom_settings"`
}

// ErrorKnowledgeBase contains curated solutions and patterns
type ErrorKnowledgeBase struct {
	CommonSolutions   map[string][]string `json:"common_solutions"`
	ProviderSpecific  map[string][]string `json:"provider_specific"`
	PlatformSpecific  map[string][]string `json:"platform_specific"`
	VersionSpecific   map[string][]string `json:"version_specific"`
	FrequentCombos    map[string]string   `json:"frequent_combos"`
}

// NewEnhancedErrorCommunicator creates a new enhanced error communicator
func NewEnhancedErrorCommunicator() *EnhancedErrorCommunicator {
	return &EnhancedErrorCommunicator{
		messageGenerator:  NewUserFriendlyMessageGenerator(),
		contextAnalyzer:   NewErrorContextAnalyzer(),
		solutionSuggester: NewSolutionSuggester(),
		patternMatcher:    NewErrorPatternMatcher(),
	}
}

// NewErrorContextAnalyzer creates a new error context analyzer
func NewErrorContextAnalyzer() *ErrorContextAnalyzer {
	return &ErrorContextAnalyzer{
		commandPatterns: make(map[string][]string),
		systemInfo:      &SystemInfo{LastUpdated: time.Now()},
		recentCommands:  make([]string, 0, 10),
		configState:     &ConfigState{},
	}
}

// NewSolutionSuggester creates a new solution suggester
func NewSolutionSuggester() *SolutionSuggester {
	return &SolutionSuggester{
		knowledgeBase:   NewErrorKnowledgeBase(),
		adaptiveFilters: make(map[string][]string),
		successPatterns: make(map[string]string),
	}
}

// NewErrorPatternMatcher creates a new error pattern matcher
func NewErrorPatternMatcher() *ErrorPatternMatcher {
	matcher := &ErrorPatternMatcher{
		nixPatterns:      make([]*EnhancedErrorPattern, 0),
		providerPatterns: make([]*EnhancedErrorPattern, 0),
		systemPatterns:   make([]*EnhancedErrorPattern, 0),
		customPatterns:   make([]*EnhancedErrorPattern, 0),
	}
	matcher.loadDefaultPatterns()
	return matcher
}

// NewErrorKnowledgeBase creates a new error knowledge base
func NewErrorKnowledgeBase() *ErrorKnowledgeBase {
	kb := &ErrorKnowledgeBase{
		CommonSolutions:  make(map[string][]string),
		ProviderSpecific: make(map[string][]string),
		PlatformSpecific: make(map[string][]string),
		VersionSpecific:  make(map[string][]string),
		FrequentCombos:   make(map[string]string),
	}
	kb.loadKnowledgeBase()
	return kb
}

// GenerateEnhancedMessage generates an AI-powered enhanced error message
func (eec *EnhancedErrorCommunicator) GenerateEnhancedMessage(ctx context.Context, err error, context string) string {
	// Start with basic user-friendly message
	baseMessage := eec.messageGenerator.GenerateUserFriendlyMessage(err)
	
	// Analyze error context
	errorContext := eec.contextAnalyzer.AnalyzeContext(err, context)
	
	// Match error patterns for better categorization
	matchedPatterns := eec.patternMatcher.MatchPatterns(err)
	
	// Generate intelligent solutions
	solutions := eec.solutionSuggester.GenerateSolutions(err, errorContext, matchedPatterns)
	
	// Build enhanced message
	enhanced := &strings.Builder{}
	enhanced.WriteString(baseMessage)
	
	// Add contextual information if available
	if errorContext != nil && len(errorContext.RelevantInfo) > 0 {
		enhanced.WriteString("\n\n📋 Context:")
		for _, info := range errorContext.RelevantInfo {
			enhanced.WriteString(fmt.Sprintf("\n  • %s", info))
		}
	}
	
	// Add intelligent solutions
	if len(solutions.PrioritizedSolutions) > 0 {
		enhanced.WriteString("\n\n🔧 Recommended Solutions:")
		for i, solution := range solutions.PrioritizedSolutions[:min(len(solutions.PrioritizedSolutions), 3)] {
			enhanced.WriteString(fmt.Sprintf("\n  %d. %s", i+1, solution.Description))
			if solution.Command != "" {
				enhanced.WriteString(fmt.Sprintf("\n     Command: %s", solution.Command))
			}
			if solution.Confidence > 0.8 {
				enhanced.WriteString(" ⭐")
			}
		}
	}
	
	// Add relevant documentation links
	if len(solutions.RelevantDocs) > 0 {
		enhanced.WriteString("\n\n📚 Documentation:")
		for _, doc := range solutions.RelevantDocs[:min(len(solutions.RelevantDocs), 2)] {
			enhanced.WriteString(fmt.Sprintf("\n  • %s", doc))
		}
	}
	
	// Add preventive measures for critical errors
	if len(solutions.PreventiveMeasures) > 0 {
		enhanced.WriteString("\n\n🛡️ Prevention:")
		for _, measure := range solutions.PreventiveMeasures[:min(len(solutions.PreventiveMeasures), 2)] {
			enhanced.WriteString(fmt.Sprintf("\n  • %s", measure))
		}
	}
	
	return enhanced.String()
}

// ContextualInfo represents analyzed error context
type ContextualInfo struct {
	ErrorType       string            `json:"error_type"`
	RelevantInfo    []string          `json:"relevant_info"`
	SystemState     map[string]string `json:"system_state"`
	RecentActivity  []string          `json:"recent_activity"`
	ConfigIssues    []string          `json:"config_issues"`
	EnvironmentVars []string          `json:"environment_vars"`
}

// SolutionSet represents a set of solutions for an error
type SolutionSet struct {
	PrioritizedSolutions []Solution `json:"prioritized_solutions"`
	RelevantDocs        []string   `json:"relevant_docs"`
	PreventiveMeasures  []string   `json:"preventive_measures"`
	RelatedErrors       []string   `json:"related_errors"`
}

// Solution represents a potential solution to an error
type Solution struct {
	Description string  `json:"description"`
	Command     string  `json:"command,omitempty"`
	Confidence  float64 `json:"confidence"`
	Complexity  string  `json:"complexity"` // "simple", "intermediate", "advanced"
	Automated   bool    `json:"automated"`
	RiskLevel   string  `json:"risk_level"` // "low", "medium", "high"
}

// AnalyzeContext analyzes the context of an error
func (eca *ErrorContextAnalyzer) AnalyzeContext(err error, context string) *ContextualInfo {
	info := &ContextualInfo{
		RelevantInfo:    make([]string, 0),
		SystemState:     make(map[string]string),
		RecentActivity:  make([]string, 0),
		ConfigIssues:    make([]string, 0),
		EnvironmentVars: make([]string, 0),
	}
	
	// Analyze error type
	info.ErrorType = eca.categorizeError(err)
	
	// Add system context
	eca.addSystemContext(info)
	
	// Add configuration context
	eca.addConfigContext(info)
	
	// Add recent activity context
	eca.addRecentActivityContext(info, context)
	
	return info
}

// categorizeError categorizes the error type
func (eca *ErrorContextAnalyzer) categorizeError(err error) string {
	errorMsg := strings.ToLower(err.Error())
	
	switch {
	case strings.Contains(errorMsg, "ollama"):
		return "AI Provider (Ollama)"
	case strings.Contains(errorMsg, "openai"):
		return "AI Provider (OpenAI)"
	case strings.Contains(errorMsg, "gemini"):
		return "AI Provider (Gemini)"
	case strings.Contains(errorMsg, "claude"):
		return "AI Provider (Claude)"
	case strings.Contains(errorMsg, "nix"):
		return "NixOS System"
	case strings.Contains(errorMsg, "cache"):
		return "Cache System"
	case strings.Contains(errorMsg, "network"):
		return "Network"
	case strings.Contains(errorMsg, "permission"):
		return "Permissions"
	default:
		return "General"
	}
}

// addSystemContext adds system information to context
func (eca *ErrorContextAnalyzer) addSystemContext(info *ContextualInfo) {
	// Check system resources
	if eca.systemInfo.AvailableMemory < 1000000000 { // Less than 1GB
		info.RelevantInfo = append(info.RelevantInfo, "Low available memory detected")
	}
	
	if eca.systemInfo.DiskSpace < 5000000000 { // Less than 5GB
		info.RelevantInfo = append(info.RelevantInfo, "Low disk space detected")
	}
	
	// Add active providers
	if len(eca.systemInfo.ActiveProviders) > 0 {
		info.SystemState["active_providers"] = strings.Join(eca.systemInfo.ActiveProviders, ", ")
	}
}

// addConfigContext adds configuration information to context
func (eca *ErrorContextAnalyzer) addConfigContext(info *ContextualInfo) {
	if eca.configState.DebugMode {
		info.RelevantInfo = append(info.RelevantInfo, "Debug mode is enabled")
	}
	
	if !eca.configState.CacheEnabled {
		info.RelevantInfo = append(info.RelevantInfo, "Caching is disabled")
	}
}

// addRecentActivityContext adds recent activity information
func (eca *ErrorContextAnalyzer) addRecentActivityContext(info *ContextualInfo, context string) {
	if context != "" {
		info.RecentActivity = append(info.RecentActivity, context)
	}
	
	// Add recent commands
	info.RecentActivity = append(info.RecentActivity, eca.recentCommands...)
}

// GenerateSolutions generates intelligent solutions based on error analysis
func (ss *SolutionSuggester) GenerateSolutions(err error, context *ContextualInfo, patterns []*EnhancedErrorPattern) *SolutionSet {
	solutions := &SolutionSet{
		PrioritizedSolutions: make([]Solution, 0),
		RelevantDocs:        make([]string, 0),
		PreventiveMeasures:  make([]string, 0),
		RelatedErrors:       make([]string, 0),
	}
	
	// Generate solutions based on error type
	ss.addTypedSolutions(solutions, err, context)
	
	// Add pattern-based solutions
	ss.addPatternSolutions(solutions, patterns)
	
	// Add context-specific solutions
	ss.addContextualSolutions(solutions, context)
	
	// Sort solutions by confidence
	ss.sortSolutionsByConfidence(solutions)
	
	return solutions
}

// addTypedSolutions adds solutions based on error type
func (ss *SolutionSuggester) addTypedSolutions(solutions *SolutionSet, err error, context *ContextualInfo) {
	errorMsg := strings.ToLower(err.Error())
	
	// AI Provider specific solutions
	if strings.Contains(errorMsg, "ollama") {
		solutions.PrioritizedSolutions = append(solutions.PrioritizedSolutions, Solution{
			Description: "Check if Ollama service is running",
			Command:     "systemctl status ollama || ollama serve",
			Confidence:  0.9,
			Complexity:  "simple",
			Automated:   false,
			RiskLevel:   "low",
		})
		solutions.RelevantDocs = append(solutions.RelevantDocs, "https://ollama.ai/docs/installation")
	}
	
	if strings.Contains(errorMsg, "connection refused") || strings.Contains(errorMsg, "timeout") {
		solutions.PrioritizedSolutions = append(solutions.PrioritizedSolutions, Solution{
			Description: "Test network connectivity",
			Command:     "ping -c 3 google.com",
			Confidence:  0.8,
			Complexity:  "simple",
			Automated:   false,
			RiskLevel:   "low",
		})
	}
	
	if strings.Contains(errorMsg, "permission denied") {
		solutions.PrioritizedSolutions = append(solutions.PrioritizedSolutions, Solution{
			Description: "Check file permissions and ownership",
			Command:     "ls -la ~/.config/nixai/",
			Confidence:  0.85,
			Complexity:  "simple",
			Automated:   false,
			RiskLevel:   "low",
		})
	}
}

// addPatternSolutions adds solutions based on matched patterns
func (ss *SolutionSuggester) addPatternSolutions(solutions *SolutionSet, patterns []*EnhancedErrorPattern) {
	for _, pattern := range patterns {
		for _, solution := range pattern.Solutions {
			solutions.PrioritizedSolutions = append(solutions.PrioritizedSolutions, Solution{
				Description: solution,
				Confidence:  0.7,
				Complexity:  "intermediate",
				Automated:   false,
				RiskLevel:   "medium",
			})
		}
		solutions.RelevantDocs = append(solutions.RelevantDocs, pattern.DocsLinks...)
	}
}

// addContextualSolutions adds solutions based on system context
func (ss *SolutionSuggester) addContextualSolutions(solutions *SolutionSet, context *ContextualInfo) {
	if context == nil {
		return
	}
	
	// Add memory-related solutions
	for _, info := range context.RelevantInfo {
		if strings.Contains(strings.ToLower(info), "memory") {
			solutions.PrioritizedSolutions = append(solutions.PrioritizedSolutions, Solution{
				Description: "Free up system memory",
				Command:     "sudo sysctl vm.drop_caches=1",
				Confidence:  0.6,
				Complexity:  "intermediate",
				Automated:   false,
				RiskLevel:   "medium",
			})
			solutions.PreventiveMeasures = append(solutions.PreventiveMeasures, 
				"Monitor system memory usage regularly")
		}
		
		if strings.Contains(strings.ToLower(info), "disk") {
			solutions.PrioritizedSolutions = append(solutions.PrioritizedSolutions, Solution{
				Description: "Clean up disk space",
				Command:     "nix-collect-garbage -d",
				Confidence:  0.8,
				Complexity:  "simple",
				Automated:   false,
				RiskLevel:   "low",
			})
			solutions.PreventiveMeasures = append(solutions.PreventiveMeasures, 
				"Set up automatic garbage collection")
		}
	}
}

// sortSolutionsByConfidence sorts solutions by confidence score
func (ss *SolutionSuggester) sortSolutionsByConfidence(solutions *SolutionSet) {
	// Simple bubble sort by confidence (descending)
	for i := 0; i < len(solutions.PrioritizedSolutions); i++ {
		for j := i + 1; j < len(solutions.PrioritizedSolutions); j++ {
			if solutions.PrioritizedSolutions[i].Confidence < solutions.PrioritizedSolutions[j].Confidence {
				solutions.PrioritizedSolutions[i], solutions.PrioritizedSolutions[j] = 
					solutions.PrioritizedSolutions[j], solutions.PrioritizedSolutions[i]
			}
		}
	}
}

// MatchPatterns matches error against known patterns
func (epm *ErrorPatternMatcher) MatchPatterns(err error) []*EnhancedErrorPattern {
	matches := make([]*EnhancedErrorPattern, 0)
	errorMsg := err.Error()
	
	// Check all pattern categories
	allPatterns := append(epm.nixPatterns, epm.providerPatterns...)
	allPatterns = append(allPatterns, epm.systemPatterns...)
	allPatterns = append(allPatterns, epm.customPatterns...)
	
	for _, pattern := range allPatterns {
		if pattern.Pattern.MatchString(errorMsg) {
			matches = append(matches, pattern)
		}
	}
	
	return matches
}

// loadDefaultPatterns loads default error patterns
func (epm *ErrorPatternMatcher) loadDefaultPatterns() {
	// Provider patterns
	epm.providerPatterns = append(epm.providerPatterns, &EnhancedErrorPattern{
		Name:       "Ollama Connection Error",
		Pattern:    regexp.MustCompile(`ollama.*connection\s+refused|ollama.*timeout`),
		PatternStr: "ollama.*connection\\s+refused|ollama.*timeout",
		Category:   CategoryAI,
		Severity:   SeverityMedium,
		Solutions: []string{
			"Start Ollama service: systemctl start ollama",
			"Check Ollama status: ollama list",
			"Verify Ollama installation: which ollama",
		},
		DocsLinks: []string{"https://ollama.ai/docs/troubleshooting"},
	})
	
	// Nix patterns
	epm.nixPatterns = append(epm.nixPatterns, &EnhancedErrorPattern{
		Name:       "Nix Build Failure",
		Pattern:    regexp.MustCompile(`nix.*build.*failed|builder.*failed`),
		PatternStr: "nix.*build.*failed|builder.*failed",
		Category:   CategoryNixOS,
		Severity:   SeverityHigh,
		Solutions: []string{
			"Check build logs for specific errors",
			"Clean Nix store: nix-collect-garbage",
			"Update channels: nix-channel --update",
		},
		DocsLinks: []string{"https://wiki.nixos.org/wiki/Troubleshooting"},
	})
	
	// System patterns
	epm.systemPatterns = append(epm.systemPatterns, &EnhancedErrorPattern{
		Name:       "Permission Denied",
		Pattern:    regexp.MustCompile(`permission\s+denied|access\s+denied`),
		PatternStr: "permission\\s+denied|access\\s+denied",
		Category:   CategoryFileSystem,
		Severity:   SeverityMedium,
		Solutions: []string{
			"Check file permissions: ls -la",
			"Verify ownership: ls -la | grep username",
			"Fix permissions: chmod 755 or chown user:group",
		},
		DocsLinks: []string{"https://wiki.nixos.org/wiki/File_permissions"},
	})
}

// loadKnowledgeBase loads the error knowledge base
func (ekb *ErrorKnowledgeBase) loadKnowledgeBase() {
	// Common solutions
	ekb.CommonSolutions["connection_refused"] = []string{
		"Check if the service is running",
		"Verify network connectivity",
		"Check firewall settings",
		"Restart the service",
	}
	
	ekb.CommonSolutions["permission_denied"] = []string{
		"Check file ownership and permissions",
		"Run with appropriate privileges",
		"Verify access rights",
		"Check SELinux/AppArmor policies",
	}
	
	// Provider-specific solutions
	ekb.ProviderSpecific["ollama"] = []string{
		"Ensure Ollama is installed and running",
		"Check Ollama model availability",
		"Verify Ollama configuration",
		"Update Ollama to latest version",
	}
	
	ekb.ProviderSpecific["openai"] = []string{
		"Verify API key is valid and not expired",
		"Check OpenAI service status",
		"Verify account has sufficient credits",
		"Check rate limiting and quotas",
	}
	
	// Platform-specific solutions
	ekb.PlatformSpecific["nixos"] = []string{
		"Check NixOS configuration syntax",
		"Update system channels",
		"Rebuild NixOS configuration",
		"Check system logs for details",
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}