package tui

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/ai/function"
	"nix-ai-help/pkg/logger"
)

// IntelligentDiscovery provides AI-powered command discovery and suggestions
type IntelligentDiscovery struct {
	commands         []Command
	usageAnalytics   *UsageAnalytics
	contextAnalyzer  *CommandContextAnalyzer
	fuzzyMatcher     *FuzzyMatcher
	aiSuggester      *AISuggester
	logger           *logger.Logger
	lastActivity     time.Time
	sessionCommands  []string
}

// UsageAnalytics tracks command usage patterns
type UsageAnalytics struct {
	commandFrequency map[string]int
	commandTiming    map[string][]time.Time
	commandSequences map[string][]string // Commands that often follow this command
	errorPatterns    map[string]int
	successPatterns  map[string]int
	contextPatterns  map[string]map[string]int // context -> command -> frequency
	userPreferences  *UserPreferences
	lastUpdated      time.Time
}

// UserPreferences stores learned user behavior patterns
type UserPreferences struct {
	PreferredProviders   []string          `json:"preferred_providers"`
	CommonFlags          map[string]string `json:"common_flags"`
	FrequentPatterns     []string          `json:"frequent_patterns"`
	PreferredCategories  []string          `json:"preferred_categories"`
	SkippedSuggestions   map[string]int    `json:"skipped_suggestions"`
	AcceptedSuggestions  map[string]int    `json:"accepted_suggestions"`
	TimeOfDayPatterns    map[int][]string  `json:"time_patterns"` // hour -> commands
	CommandAliases       map[string]string `json:"aliases"`
}

// CommandContextAnalyzer analyzes context for better suggestions
type CommandContextAnalyzer struct {
	systemContext    *SystemContext
	projectContext   *ProjectContext
	temporalContext  *TemporalContext
	errorContext     *ErrorContext
}

// SystemContext contains system state information
type SystemContext struct {
	CurrentDirectory   string            `json:"current_directory"`
	RunningServices    []string          `json:"running_services"`
	RecentErrors       []string          `json:"recent_errors"`
	SystemLoad         float64           `json:"system_load"`
	MemoryUsage        float64           `json:"memory_usage"`
	DiskUsage          float64           `json:"disk_usage"`
	ActiveConnections  int               `json:"active_connections"`
	InstalledPackages  []string          `json:"installed_packages"`
	ConfigurationState map[string]string `json:"config_state"`
	LastUpdated        time.Time         `json:"last_updated"`
}

// ProjectContext contains project-specific information
type ProjectContext struct {
	HasFlakeNix        bool              `json:"has_flake_nix"`
	HasConfigurationNix bool             `json:"has_configuration_nix"`
	HasDevShell        bool              `json:"has_dev_shell"`
	ProjectType        string            `json:"project_type"`
	Dependencies       []string          `json:"dependencies"`
	RecentFiles        []string          `json:"recent_files"`
	GitBranch          string            `json:"git_branch"`
	GitStatus          string            `json:"git_status"`
	BuildSystem        string            `json:"build_system"`
}

// TemporalContext contains time-based patterns
type TemporalContext struct {
	CurrentHour       int               `json:"current_hour"`
	DayOfWeek         int               `json:"day_of_week"`
	RecentCommands    []TimedCommand    `json:"recent_commands"`
	SessionDuration   time.Duration     `json:"session_duration"`
	CommandFrequency  map[string]int    `json:"command_frequency"`
	TimePatterns      map[int][]string  `json:"time_patterns"`
}

// TimedCommand represents a command with timestamp
type TimedCommand struct {
	Command   string    `json:"command"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Duration  time.Duration `json:"duration"`
}

// ErrorContext contains error and troubleshooting context
type ErrorContext struct {
	RecentErrors      []RecentError     `json:"recent_errors"`
	ErrorPatterns     map[string]int    `json:"error_patterns"`
	CommonSolutions   map[string]string `json:"common_solutions"`
	TroubleshootingMode bool            `json:"troubleshooting_mode"`
}

// RecentError represents a recent error with context
type RecentError struct {
	Command     string    `json:"command"`
	Error       string    `json:"error"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
	Solution    string    `json:"solution"`
}

// FuzzyMatcher provides fuzzy string matching for commands
type FuzzyMatcher struct {
	threshold    float64
	weightName   float64
	weightDesc   float64
	weightUsage  float64
	weightKeywords float64
}

// AISuggester provides AI-powered command suggestions
type AISuggester struct {
	functionRegistry *function.FunctionManager
	logger           *logger.Logger
	enabled          bool
}

// SuggestionScore represents a scored command suggestion
type SuggestionScore struct {
	Command     Command
	Score       float64
	Reason      string
	Confidence  float64
	Category    string
	Urgency     string // "low", "medium", "high", "critical"
	Context     map[string]interface{}
	Keywords    []string
	Examples    []string
}

// NewIntelligentDiscovery creates a new intelligent discovery system
func NewIntelligentDiscovery(logger *logger.Logger) *IntelligentDiscovery {
	return &IntelligentDiscovery{
		commands:        getAvailableCommands(),
		usageAnalytics:  NewUsageAnalytics(),
		contextAnalyzer: NewCommandContextAnalyzer(),
		fuzzyMatcher:    NewFuzzyMatcher(),
		aiSuggester:     NewAISuggester(logger),
		logger:          logger,
		lastActivity:    time.Now(),
		sessionCommands: make([]string, 0),
	}
}

// NewUsageAnalytics creates a new usage analytics instance
func NewUsageAnalytics() *UsageAnalytics {
	return &UsageAnalytics{
		commandFrequency: make(map[string]int),
		commandTiming:    make(map[string][]time.Time),
		commandSequences: make(map[string][]string),
		errorPatterns:    make(map[string]int),
		successPatterns:  make(map[string]int),
		contextPatterns:  make(map[string]map[string]int),
		userPreferences:  NewUserPreferences(),
		lastUpdated:      time.Now(),
	}
}

// NewUserPreferences creates default user preferences
func NewUserPreferences() *UserPreferences {
	return &UserPreferences{
		PreferredProviders:  []string{"ollama", "claude", "openai"},
		CommonFlags:         make(map[string]string),
		FrequentPatterns:    make([]string, 0),
		PreferredCategories: []string{"AI", "Diagnostics", "Configuration"},
		SkippedSuggestions:  make(map[string]int),
		AcceptedSuggestions: make(map[string]int),
		TimeOfDayPatterns:   make(map[int][]string),
		CommandAliases:      make(map[string]string),
	}
}

// NewCommandContextAnalyzer creates a new context analyzer
func NewCommandContextAnalyzer() *CommandContextAnalyzer {
	return &CommandContextAnalyzer{
		systemContext:   &SystemContext{LastUpdated: time.Now()},
		projectContext:  &ProjectContext{},
		temporalContext: &TemporalContext{CurrentHour: time.Now().Hour(), DayOfWeek: int(time.Now().Weekday())},
		errorContext:    &ErrorContext{RecentErrors: make([]RecentError, 0)},
	}
}

// NewFuzzyMatcher creates a new fuzzy matcher with default weights
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{
		threshold:      0.3, // Minimum match score
		weightName:     0.4,
		weightDesc:     0.3,
		weightUsage:    0.2,
		weightKeywords: 0.1,
	}
}

// NewAISuggester creates a new AI-powered suggester
func NewAISuggester(logger *logger.Logger) *AISuggester {
	return &AISuggester{
		functionRegistry: function.GetGlobalRegistry(),
		logger:           logger,
		enabled:          true,
	}
}

// GetIntelligentSuggestions returns AI-powered command suggestions
func (id *IntelligentDiscovery) GetIntelligentSuggestions(ctx context.Context, input string, maxSuggestions int) ([]SuggestionScore, error) {
	id.lastActivity = time.Now()
	
	// Update context
	id.updateContext()
	
	// Get base suggestions using fuzzy matching
	fuzzySuggestions := id.getFuzzySuggestions(input, maxSuggestions*2)
	
	// Enhance with usage analytics
	analyticsSuggestions := id.enhanceWithAnalytics(fuzzySuggestions, input)
	
	// Add context-aware suggestions
	contextSuggestions := id.getContextAwareSuggestions(input, maxSuggestions)
	
	// Combine and deduplicate
	allSuggestions := id.combineSuggestions(analyticsSuggestions, contextSuggestions)
	
	// Apply AI enhancement if enabled
	if id.aiSuggester.enabled {
		aiEnhanced, err := id.aiSuggester.enhanceSuggestions(ctx, allSuggestions, input, id.contextAnalyzer)
		if err != nil {
			id.logger.Debug(fmt.Sprintf("AI enhancement failed: %v", err))
		} else {
			allSuggestions = aiEnhanced
		}
	}
	
	// Sort by score and limit results
	sort.Slice(allSuggestions, func(i, j int) bool {
		return allSuggestions[i].Score > allSuggestions[j].Score
	})
	
	if len(allSuggestions) > maxSuggestions {
		allSuggestions = allSuggestions[:maxSuggestions]
	}
	
	return allSuggestions, nil
}

// getFuzzySuggestions performs fuzzy matching on commands
func (id *IntelligentDiscovery) getFuzzySuggestions(input string, maxResults int) []SuggestionScore {
	if input == "" {
		return id.getPopularCommands(maxResults)
	}
	
	suggestions := make([]SuggestionScore, 0)
	
	for _, cmd := range id.commands {
		score := id.fuzzyMatcher.calculateScore(input, cmd)
		if score > id.fuzzyMatcher.threshold {
			suggestion := SuggestionScore{
				Command:    cmd,
				Score:      score,
				Reason:     id.fuzzyMatcher.generateReason(input, cmd, score),
				Confidence: score,
				Category:   cmd.Category,
				Context:    make(map[string]interface{}),
				Keywords:   id.extractKeywords(cmd),
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions
}

// calculateScore calculates fuzzy match score
func (fm *FuzzyMatcher) calculateScore(input string, cmd Command) float64 {
	input = strings.ToLower(input)
	
	// Score different fields
	nameScore := fm.stringScore(input, strings.ToLower(cmd.Name))
	descScore := fm.stringScore(input, strings.ToLower(cmd.Description))
	usageScore := fm.stringScore(input, strings.ToLower(cmd.Usage))
	
	// Calculate weighted score
	totalScore := nameScore*fm.weightName + 
		descScore*fm.weightDesc + 
		usageScore*fm.weightUsage
	
	// Bonus for exact prefix match
	if strings.HasPrefix(strings.ToLower(cmd.Name), input) {
		totalScore += 0.3
	}
	
	// Bonus for word boundary matches
	if strings.Contains(strings.ToLower(cmd.Name), input) {
		totalScore += 0.2
	}
	
	return math.Min(totalScore, 1.0)
}

// stringScore calculates similarity between two strings
func (fm *FuzzyMatcher) stringScore(input, target string) float64 {
	if input == "" {
		return 0.0
	}
	
	if input == target {
		return 1.0
	}
	
	if strings.HasPrefix(target, input) {
		return 0.8 + 0.2*(float64(len(input))/float64(len(target)))
	}
	
	if strings.Contains(target, input) {
		return 0.6 + 0.2*(float64(len(input))/float64(len(target)))
	}
	
	// Calculate character-based similarity
	return fm.levenshteinSimilarity(input, target)
}

// levenshteinSimilarity calculates similarity using Levenshtein distance
func (fm *FuzzyMatcher) levenshteinSimilarity(s1, s2 string) float64 {
	if len(s1) == 0 {
		return float64(len(s2))
	}
	if len(s2) == 0 {
		return float64(len(s1))
	}
	
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}
	
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1),      // deletion, insertion
				matrix[i-1][j-1]+cost,                        // substitution
			)
		}
	}
	
	distance := matrix[len(s1)][len(s2)]
	maxLen := max(len(s1), len(s2))
	
	return 1.0 - float64(distance)/float64(maxLen)
}

// generateReason generates explanation for why command was suggested
func (fm *FuzzyMatcher) generateReason(input string, cmd Command, score float64) string {
	if strings.HasPrefix(strings.ToLower(cmd.Name), strings.ToLower(input)) {
		return fmt.Sprintf("Starts with '%s'", input)
	}
	if strings.Contains(strings.ToLower(cmd.Name), strings.ToLower(input)) {
		return fmt.Sprintf("Contains '%s'", input)
	}
	if strings.Contains(strings.ToLower(cmd.Description), strings.ToLower(input)) {
		return fmt.Sprintf("Description matches '%s'", input)
	}
	if score > 0.7 {
		return "High similarity match"
	}
	return "Fuzzy match"
}

// enhanceWithAnalytics enhances suggestions using usage analytics
func (id *IntelligentDiscovery) enhanceWithAnalytics(suggestions []SuggestionScore, input string) []SuggestionScore {
	enhanced := make([]SuggestionScore, len(suggestions))
	copy(enhanced, suggestions)
	
	for i := range enhanced {
		cmdName := enhanced[i].Command.Name
		
		// Boost score based on usage frequency
		if freq, exists := id.usageAnalytics.commandFrequency[cmdName]; exists {
			boost := math.Log10(float64(freq+1)) * 0.1
			enhanced[i].Score += boost
			enhanced[i].Reason += fmt.Sprintf(" (used %d times)", freq)
		}
		
		// Boost based on recent usage
		if timings, exists := id.usageAnalytics.commandTiming[cmdName]; exists && len(timings) > 0 {
			lastUsed := timings[len(timings)-1]
			if time.Since(lastUsed) < 24*time.Hour {
				enhanced[i].Score += 0.1
				enhanced[i].Reason += " (recently used)"
			}
		}
		
		// Boost based on context patterns
		currentContext := id.getCurrentContext()
		if contextMap, exists := id.usageAnalytics.contextPatterns[currentContext]; exists {
			if freq, exists := contextMap[cmdName]; exists {
				boost := float64(freq) * 0.05
				enhanced[i].Score += boost
				enhanced[i].Reason += fmt.Sprintf(" (common in %s)", currentContext)
			}
		}
		
		// Boost based on command sequences
		if len(id.sessionCommands) > 0 {
			lastCommand := id.sessionCommands[len(id.sessionCommands)-1]
			if sequences, exists := id.usageAnalytics.commandSequences[lastCommand]; exists {
				for _, nextCmd := range sequences {
					if nextCmd == cmdName {
						enhanced[i].Score += 0.2
						enhanced[i].Reason += " (follows " + lastCommand + ")"
						break
					}
				}
			}
		}
	}
	
	return enhanced
}

// getContextAwareSuggestions generates context-specific suggestions
func (id *IntelligentDiscovery) getContextAwareSuggestions(input string, maxResults int) []SuggestionScore {
	suggestions := make([]SuggestionScore, 0)
	
	// Add error-specific suggestions
	if id.contextAnalyzer.errorContext.TroubleshootingMode {
		suggestions = append(suggestions, id.getTroubleshootingSuggestions()...)
	}
	
	// Add project-specific suggestions
	if id.contextAnalyzer.projectContext.HasFlakeNix {
		suggestions = append(suggestions, id.getFlakeSuggestions()...)
	}
	
	// Add time-based suggestions
	suggestions = append(suggestions, id.getTimeBasedSuggestions()...)
	
	// Add system state suggestions
	suggestions = append(suggestions, id.getSystemStateSuggestions()...)
	
	return suggestions
}

// getTroubleshootingSuggestions returns troubleshooting-focused suggestions
func (id *IntelligentDiscovery) getTroubleshootingSuggestions() []SuggestionScore {
	suggestions := []SuggestionScore{
		{
			Command:    Command{Name: "diagnose", Description: "Diagnose system and configuration issues", Category: "Diagnostics", Usage: "nixai diagnose [category]", Examples: []string{"nixai diagnose boot"}},
			Score:      0.9,
			Reason:     "Troubleshooting mode active",
			Confidence: 0.9,
			Urgency:    "high",
			Context:    make(map[string]interface{}),
		},
		{
			Command:    Command{Name: "doctor", Description: "Comprehensive system health check", Category: "Diagnostics", Usage: "nixai doctor [options]", Examples: []string{"nixai doctor --full"}},
			Score:      0.8,
			Reason:     "Health check recommended",
			Confidence: 0.8,
			Urgency:    "medium",
			Context:    make(map[string]interface{}),
		},
		{
			Command:    Command{Name: "logs", Description: "Analyze system and service logs", Category: "Diagnostics", Usage: "nixai logs [service]", Examples: []string{"nixai logs nginx"}},
			Score:      0.7,
			Reason:     "Check logs for errors",
			Confidence: 0.7,
			Urgency:    "medium",
			Context:    make(map[string]interface{}),
		},
	}
	
	return suggestions
}

// getFlakeSuggestions returns flake-specific suggestions
func (id *IntelligentDiscovery) getFlakeSuggestions() []SuggestionScore {
	return []SuggestionScore{
		{
			Command:    Command{Name: "flake", Description: "Manage Nix flakes", Category: "Flakes", Usage: "nixai flake [action]", Examples: []string{"nixai flake create"}},
			Score:      0.8,
			Reason:     "Flake project detected",
			Confidence: 0.8,
			Context:    make(map[string]interface{}),
		},
		{
			Command:    Command{Name: "build", Description: "Build and manage NixOS configurations", Category: "Build", Usage: "nixai build [action]", Examples: []string{"nixai build analyze"}},
			Score:      0.7,
			Reason:     "Build flake project",
			Confidence: 0.7,
			Context:    make(map[string]interface{}),
		},
	}
}

// getTimeBasedSuggestions returns time-appropriate suggestions
func (id *IntelligentDiscovery) getTimeBasedSuggestions() []SuggestionScore {
	suggestions := make([]SuggestionScore, 0)
	currentHour := time.Now().Hour()
	
	// Morning suggestions (6-12)
	if currentHour >= 6 && currentHour < 12 {
		suggestions = append(suggestions, SuggestionScore{
			Command:    Command{Name: "health", Description: "System health monitoring and prediction", Category: "Monitoring", Usage: "nixai health [action]", Examples: []string{"nixai health status"}},
			Score:      0.6,
			Reason:     "Morning health check",
			Confidence: 0.6,
			Context:    make(map[string]interface{}),
		})
	}
	
	// Evening suggestions (18-23)
	if currentHour >= 18 && currentHour < 23 {
		suggestions = append(suggestions, SuggestionScore{
			Command:    Command{Name: "gc", Description: "Garbage collection and cleanup", Category: "Maintenance", Usage: "nixai gc [action]", Examples: []string{"nixai gc run"}},
			Score:      0.6,
			Reason:     "Evening maintenance",
			Confidence: 0.6,
			Context:    make(map[string]interface{}),
		})
	}
	
	return suggestions
}

// getSystemStateSuggestions returns suggestions based on system state
func (id *IntelligentDiscovery) getSystemStateSuggestions() []SuggestionScore {
	suggestions := make([]SuggestionScore, 0)
	
	// High memory usage suggestions
	if id.contextAnalyzer.systemContext.MemoryUsage > 0.8 {
		suggestions = append(suggestions, SuggestionScore{
			Command:    Command{Name: "performance", Description: "Performance analysis and optimization", Category: "Performance", Usage: "nixai performance [action]", Examples: []string{"nixai performance stats"}},
			Score:      0.8,
			Reason:     "High memory usage detected",
			Confidence: 0.8,
			Urgency:    "high",
			Context:    make(map[string]interface{}),
		})
	}
	
	// High disk usage suggestions
	if id.contextAnalyzer.systemContext.DiskUsage > 0.9 {
		suggestions = append(suggestions, SuggestionScore{
			Command:    Command{Name: "gc", Description: "Garbage collection and cleanup", Category: "Maintenance", Usage: "nixai gc [action]", Examples: []string{"nixai gc run"}},
			Score:      0.9,
			Reason:     "Low disk space",
			Confidence: 0.9,
			Context:    make(map[string]interface{}),
			Urgency:    "critical",
		})
	}
	
	return suggestions
}

// enhanceSuggestions uses AI to enhance suggestions
func (as *AISuggester) enhanceSuggestions(ctx context.Context, suggestions []SuggestionScore, input string, contextAnalyzer *CommandContextAnalyzer) ([]SuggestionScore, error) {
	if !as.enabled || len(suggestions) == 0 {
		return suggestions, nil
	}
	
	// For now, return enhanced suggestions with AI-generated reasons
	enhanced := make([]SuggestionScore, len(suggestions))
	copy(enhanced, suggestions)
	
	for i := range enhanced {
		// Enhance with AI-generated context
		enhanced[i] = as.enhanceSingleSuggestion(enhanced[i], input, contextAnalyzer)
	}
	
	return enhanced, nil
}

// enhanceSingleSuggestion enhances a single suggestion with AI context
func (as *AISuggester) enhanceSingleSuggestion(suggestion SuggestionScore, input string, contextAnalyzer *CommandContextAnalyzer) SuggestionScore {
	// Initialize Context map if nil
	if suggestion.Context == nil {
		suggestion.Context = make(map[string]interface{})
	}
	
	// Add AI-enhanced context
	suggestion.Context["ai_enhanced"] = true
	suggestion.Context["input_analysis"] = as.analyzeInput(input)
	suggestion.Context["system_context"] = contextAnalyzer.systemContext
	
	// Note: Reason enhancement removed to keep suggestions clean
	// The original reason is kept as-is without emoji additions
	
	return suggestion
}

// analyzeInput provides basic input analysis
func (as *AISuggester) analyzeInput(input string) map[string]interface{} {
	analysis := make(map[string]interface{})
	
	// Basic pattern detection
	if matched, _ := regexp.MatchString(`\b(error|fail|broken|issue|problem)\b`, input); matched {
		analysis["intent"] = "troubleshooting"
		analysis["confidence"] = 0.8
	} else if matched, _ := regexp.MatchString(`\b(configure|setup|install)\b`, input); matched {
		analysis["intent"] = "configuration"
		analysis["confidence"] = 0.7
	} else if matched, _ := regexp.MatchString(`\b(build|compile|make)\b`, input); matched {
		analysis["intent"] = "building"
		analysis["confidence"] = 0.7
	}
	
	analysis["length"] = len(input)
	analysis["words"] = len(strings.Fields(input))
	
	return analysis
}

// Helper functions
func (id *IntelligentDiscovery) updateContext() {
	// Update temporal context
	id.contextAnalyzer.temporalContext.CurrentHour = time.Now().Hour()
	id.contextAnalyzer.temporalContext.DayOfWeek = int(time.Now().Weekday())
	
	// Update session duration
	// This would be implemented based on actual session tracking
}

func (id *IntelligentDiscovery) getCurrentContext() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (id *IntelligentDiscovery) getPopularCommands(maxResults int) []SuggestionScore {
	// Return most popular commands based on frequency
	popularCommands := []string{"ask", "help", "configure", "build", "diagnose", "doctor", "search", "health"}
	
	suggestions := make([]SuggestionScore, 0, len(popularCommands))
	for i, cmdName := range popularCommands {
		if i >= maxResults {
			break
		}
		
		for _, cmd := range id.commands {
			if cmd.Name == cmdName {
				suggestions = append(suggestions, SuggestionScore{
					Command:    cmd,
					Score:      0.5 + float64(len(popularCommands)-i)*0.05,
					Reason:     "Popular command",
					Confidence: 0.6,
					Category:   cmd.Category,
					Context:    make(map[string]interface{}),
				})
				break
			}
		}
	}
	
	return suggestions
}

func (id *IntelligentDiscovery) combineSuggestions(suggestions1, suggestions2 []SuggestionScore) []SuggestionScore {
	combined := make(map[string]SuggestionScore)
	
	// Add first set
	for _, s := range suggestions1 {
		combined[s.Command.Name] = s
	}
	
	// Add second set, keeping highest scores
	for _, s := range suggestions2 {
		if existing, exists := combined[s.Command.Name]; exists {
			if s.Score > existing.Score {
				combined[s.Command.Name] = s
			}
		} else {
			combined[s.Command.Name] = s
		}
	}
	
	// Convert back to slice
	result := make([]SuggestionScore, 0, len(combined))
	for _, s := range combined {
		result = append(result, s)
	}
	
	return result
}

func (id *IntelligentDiscovery) extractKeywords(cmd Command) []string {
	keywords := make([]string, 0)
	
	// Extract from name
	keywords = append(keywords, cmd.Name)
	
	// Extract from description
	words := strings.Fields(strings.ToLower(cmd.Description))
	for _, word := range words {
		if len(word) > 3 && !isCommonWord(word) {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "that": true,
		"this": true, "from": true, "they": true, "have": true, "your": true,
	}
	return commonWords[word]
}

// RecordCommand records a command execution for analytics
func (id *IntelligentDiscovery) RecordCommand(command string, success bool, duration time.Duration) {
	id.sessionCommands = append(id.sessionCommands, command)
	
	// Update frequency
	id.usageAnalytics.commandFrequency[command]++
	
	// Update timing
	id.usageAnalytics.commandTiming[command] = append(id.usageAnalytics.commandTiming[command], time.Now())
	
	// Update success/error patterns
	if success {
		id.usageAnalytics.successPatterns[command]++
	} else {
		id.usageAnalytics.errorPatterns[command]++
	}
	
	// Update context patterns
	context := id.getCurrentContext()
	if id.usageAnalytics.contextPatterns[context] == nil {
		id.usageAnalytics.contextPatterns[context] = make(map[string]int)
	}
	id.usageAnalytics.contextPatterns[context][command]++
	
	// Update command sequences
	if len(id.sessionCommands) >= 2 {
		prevCommand := id.sessionCommands[len(id.sessionCommands)-2]
		id.usageAnalytics.commandSequences[prevCommand] = append(id.usageAnalytics.commandSequences[prevCommand], command)
	}
	
	id.usageAnalytics.lastUpdated = time.Now()
}

// RecordSuggestionInteraction records user interaction with suggestions
func (id *IntelligentDiscovery) RecordSuggestionInteraction(suggestion string, accepted bool) {
	if accepted {
		id.usageAnalytics.userPreferences.AcceptedSuggestions[suggestion]++
	} else {
		id.usageAnalytics.userPreferences.SkippedSuggestions[suggestion]++
	}
}

// GetAnalytics returns usage analytics data
func (id *IntelligentDiscovery) GetAnalytics() *UsageAnalytics {
	return id.usageAnalytics
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getAvailableCommands returns the list of available nixai commands
func getAvailableCommands() []Command {
	return []Command{
		// Core AI commands
		{Name: "ask", Description: "Ask AI about NixOS or system issues", Category: "AI", Usage: "nixai ask [question]", Examples: []string{"nixai ask \"how to configure nginx?\""}},
		{Name: "ai-config", Description: "Configure AI provider settings", Category: "AI", Usage: "nixai ai-config [provider]", Examples: []string{"nixai ai-config ollama"}},
		{Name: "configure", Description: "Generate NixOS configurations", Category: "Configuration", Usage: "nixai configure [service]", Examples: []string{"nixai configure nginx"}},
		{Name: "explain-option", Description: "Explain NixOS configuration options", Category: "Documentation", Usage: "nixai explain-option [option]", Examples: []string{"nixai explain-option services.nginx"}},
		{Name: "explain-home-option", Description: "Explain Home Manager options", Category: "Documentation", Usage: "nixai explain-home-option [option]", Examples: []string{"nixai explain-home-option programs.git"}},
		
		// Build and development
		{Name: "build", Description: "Build and manage NixOS configurations", Category: "Build", Usage: "nixai build [action]", Examples: []string{"nixai build analyze"}},
		{Name: "devenv", Description: "Manage development environments", Category: "Development", Usage: "nixai devenv [action]", Examples: []string{"nixai devenv create"}},
		{Name: "import", Description: "Import configurations and packages", Category: "Import", Usage: "nixai import [source]", Examples: []string{"nixai import github:user/repo"}},
		{Name: "templates", Description: "Manage configuration templates", Category: "Templates", Usage: "nixai templates [action]", Examples: []string{"nixai templates list"}},
		
		// Diagnostics and troubleshooting
		{Name: "diagnose", Description: "Diagnose system and configuration issues", Category: "Diagnostics", Usage: "nixai diagnose [category]", Examples: []string{"nixai diagnose boot"}},
		{Name: "doctor", Description: "Comprehensive system health check", Category: "Diagnostics", Usage: "nixai doctor [options]", Examples: []string{"nixai doctor --full"}},
		{Name: "error", Description: "Analyze and explain error messages", Category: "Troubleshooting", Usage: "nixai error [message]", Examples: []string{"nixai error \"build failed\""}},
		{Name: "logs", Description: "Analyze system and service logs", Category: "Diagnostics", Usage: "nixai logs [service]", Examples: []string{"nixai logs nginx"}},
		{Name: "performance", Description: "Performance analysis and optimization", Category: "Performance", Usage: "nixai performance [action]", Examples: []string{"nixai performance stats"}},
		{Name: "health", Description: "System health monitoring and prediction", Category: "Monitoring", Usage: "nixai health [action]", Examples: []string{"nixai health status"}},
		
		// Package and dependency management
		{Name: "deps", Description: "Analyze package dependencies", Category: "Dependencies", Usage: "nixai deps [package]", Examples: []string{"nixai deps firefox"}},
		{Name: "package-repo", Description: "Package repository analysis", Category: "Packages", Usage: "nixai package-repo [action]", Examples: []string{"nixai package-repo analyze"}},
		{Name: "search", Description: "Search for packages and options", Category: "Search", Usage: "nixai search [term]", Examples: []string{"nixai search nginx"}},
		{Name: "store", Description: "Nix store analysis and management", Category: "Store", Usage: "nixai store [action]", Examples: []string{"nixai store analyze"}},
		{Name: "gc", Description: "Garbage collection and cleanup", Category: "Maintenance", Usage: "nixai gc [action]", Examples: []string{"nixai gc run"}},
		
		// Flake management
		{Name: "flake", Description: "Manage Nix flakes", Category: "Flakes", Usage: "nixai flake [action]", Examples: []string{"nixai flake create"}},
		{Name: "migrate", Description: "Migrate to flakes or newer configurations", Category: "Migration", Usage: "nixai migrate [target]", Examples: []string{"nixai migrate flakes"}},
		
		// System and hardware
		{Name: "hardware", Description: "Hardware detection and optimization", Category: "Hardware", Usage: "nixai hardware [action]", Examples: []string{"nixai hardware detect"}},
		{Name: "context", Description: "System context analysis", Category: "Analysis", Usage: "nixai context [type]", Examples: []string{"nixai context system"}},
		{Name: "intelligence", Description: "AI system intelligence features", Category: "AI", Usage: "nixai intelligence [action]", Examples: []string{"nixai intelligence status"}},
		
		// Learning and help
		{Name: "learn", Description: "Interactive learning modules", Category: "Education", Usage: "nixai learn [module]", Examples: []string{"nixai learn basics"}},
		{Name: "help", Description: "Get help with commands", Category: "Help", Usage: "nixai help [command]", Examples: []string{"nixai help build"}},
		{Name: "community", Description: "Access community resources", Category: "Support", Usage: "nixai community [topic]", Examples: []string{"nixai community"}},
		{Name: "snippets", Description: "Code snippets and examples", Category: "Examples", Usage: "nixai snippets [category]", Examples: []string{"nixai snippets list"}},
		
		// Team and collaboration
		{Name: "team", Description: "Team collaboration features", Category: "Collaboration", Usage: "nixai team [action]", Examples: []string{"nixai team create"}},
		{Name: "fleet", Description: "Fleet management", Category: "Fleet Management", Usage: "nixai fleet [action]", Examples: []string{"nixai fleet list"}},
		{Name: "machines", Description: "Multi-machine management", Category: "Infrastructure", Usage: "nixai machines [action]", Examples: []string{"nixai machines list"}},
		
		// Workflow and automation
		{Name: "workflow", Description: "Workflow automation", Category: "Automation", Usage: "nixai workflow [action]", Examples: []string{"nixai workflow list"}},
		{Name: "plugin", Description: "Plugin management", Category: "Plugins", Usage: "nixai plugin [action]", Examples: []string{"nixai plugin list"}},
		{Name: "version-control", Description: "Version control integration", Category: "VCS", Usage: "nixai version-control [action]", Examples: []string{"nixai version-control status"}},
		{Name: "execute", Description: "Execute system commands safely", Category: "Execution", Usage: "nixai execute [command]", Examples: []string{"nixai execute status"}},
		
		// Web and integrations
		{Name: "web", Description: "Web interface for nixai", Category: "Web", Usage: "nixai web [action]", Examples: []string{"nixai web start"}},
		{Name: "mcp-server", Description: "Model Context Protocol server", Category: "Integration", Usage: "nixai mcp-server [action]", Examples: []string{"nixai mcp-server start"}},
		{Name: "neovim-setup", Description: "Neovim integration setup", Category: "Integration", Usage: "nixai neovim-setup [action]", Examples: []string{"nixai neovim-setup install"}},
		
		// Configuration management
		{Name: "config", Description: "Configuration management", Category: "Configuration", Usage: "nixai config [action]", Examples: []string{"nixai config show"}},
		{Name: "completion", Description: "Shell completion setup", Category: "Setup", Usage: "nixai completion [shell]", Examples: []string{"nixai completion bash"}},
		{Name: "tui", Description: "Terminal user interface", Category: "Interface", Usage: "nixai tui", Examples: []string{"nixai tui"}},
		
		// Integrated plugins
		{Name: "system-info", Description: "System information plugin", Category: "System", Usage: "nixai system-info [action]", Examples: []string{"nixai system-info health"}},
		{Name: "package-monitor", Description: "Package monitoring plugin", Category: "Monitoring", Usage: "nixai package-monitor [action]", Examples: []string{"nixai package-monitor list"}},
	}
}