package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// EnhancedContext represents enhanced context information for AI interactions
type EnhancedContext struct {
	*config.NixOSContext
	HistoricalInteractions []Interaction `json:"historical_interactions"`
	UserPreferences        Preferences   `json:"user_preferences"`
	SessionHistory         []string      `json:"session_history"`
	Timestamp              time.Time     `json:"timestamp"`
}

// Interaction represents a user-AI interaction
type Interaction struct {
	Timestamp   time.Time `json:"timestamp"`
	UserQuery   string    `json:"user_query"`
	AIResponse  string    `json:"ai_response"`
	ContextUsed string    `json:"context_used"`
}

// Preferences represents user preferences learned over time
type Preferences struct {
	PreferredProvider     string   `json:"preferred_provider"`
	PreferredModel        string   `json:"preferred_model"`
	FavoriteCommands      []string `json:"favorite_commands"`
	AvoidedTopics         []string `json:"avoided_topics"`
	ResponseDetailLevel   string   `json:"response_detail_level"` // brief, normal, detailed
	CodeExamplePreference  string   `json:"code_example_preference"` // none, minimal, comprehensive
}

// EnhancedContextManager manages enhanced context for AI interactions
type EnhancedContextManager struct {
	contextFile string
	mutex       sync.RWMutex
	logger      *logger.Logger
}

// NewEnhancedContextManager creates a new enhanced context manager
func NewEnhancedContextManager(logger *logger.Logger) *EnhancedContextManager {
	homeDir, _ := os.UserHomeDir()
	contextFile := filepath.Join(homeDir, ".config", "nixai", "enhanced_context.json")
	
	// Ensure directory exists
	dir := filepath.Dir(contextFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		logger.Warn(fmt.Sprintf("Failed to create context directory: %v", err))
	}
	
	return &EnhancedContextManager{
		contextFile: contextFile,
		logger:      logger,
	}
}

// LoadContext loads enhanced context from file
func (ecm *EnhancedContextManager) LoadContext() (*EnhancedContext, error) {
	ecm.mutex.Lock()
	defer ecm.mutex.Unlock()
	
	// Check if file exists
	if _, err := os.Stat(ecm.contextFile); os.IsNotExist(err) {
		return &EnhancedContext{
			NixOSContext:           &config.NixOSContext{},
			HistoricalInteractions: []Interaction{},
			UserPreferences:        Preferences{},
			SessionHistory:         []string{},
			Timestamp:              time.Now(),
		}, nil
	}
	
	// Read file
	data, err := os.ReadFile(ecm.contextFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read context file: %w", err)
	}
	
	// Parse JSON
	var context EnhancedContext
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, fmt.Errorf("failed to parse context file: %w", err)
	}
	
	return &context, nil
}

// SaveContext saves enhanced context to file
func (ecm *EnhancedContextManager) SaveContext(context *EnhancedContext) error {
	ecm.mutex.Lock()
	defer ecm.mutex.Unlock()
	
	// Update timestamp
	context.Timestamp = time.Now()
	
	// Convert to JSON
	data, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(ecm.contextFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write context file: %w", err)
	}
	
	return nil
}

// AddInteraction adds a new interaction to the context
func (ecm *EnhancedContextManager) AddInteraction(context *EnhancedContext, userQuery, aiResponse, contextUsed string) {
	ecm.mutex.Lock()
	defer ecm.mutex.Unlock()
	
	interaction := Interaction{
		Timestamp:   time.Now(),
		UserQuery:   userQuery,
		AIResponse:  aiResponse,
		ContextUsed: contextUsed,
	}
	
	// Add to interactions, keeping only the last 50
	context.HistoricalInteractions = append(context.HistoricalInteractions, interaction)
	if len(context.HistoricalInteractions) > 50 {
		context.HistoricalInteractions = context.HistoricalInteractions[len(context.HistoricalInteractions)-50:]
	}
	
	// Update session history
	context.SessionHistory = append(context.SessionHistory, userQuery)
	if len(context.SessionHistory) > 20 {
		context.SessionHistory = context.SessionHistory[len(context.SessionHistory)-20:]
	}
}

// UpdatePreferences updates user preferences based on interactions
func (ecm *EnhancedContextManager) UpdatePreferences(context *EnhancedContext, userQuery string) {
	ecm.mutex.Lock()
	defer ecm.mutex.Unlock()
	
	// Simple preference learning algorithm
	userQuery = strings.ToLower(userQuery)
	
	// Detect command preferences
	commandIndicators := map[string][]string{
		"ask":        {"ask", "question", "help"},
		"configure":  {"configure", "setup", "config"},
		"diagnose":   {"diagnose", "fix", "troubleshoot", "error"},
		"build":      {"build", "compile", "nixos-rebuild"},
		"flake":      {"flake", "flakes"},
		"search":     {"search", "find", "look for"},
		"package-repo": {"package", "repository", "repo"},
	}
	
	for command, indicators := range commandIndicators {
		for _, indicator := range indicators {
			if strings.Contains(userQuery, indicator) {
				// Add to favorites if not already there
				found := false
				for _, fav := range context.UserPreferences.FavoriteCommands {
					if fav == command {
						found = true
						break
					}
				}
				if !found {
					context.UserPreferences.FavoriteCommands = append(context.UserPreferences.FavoriteCommands, command)
				}
				break
			}
		}
	}
	
	// Detect detail level preferences
	detailIndicators := map[string][]string{
		"detailed": {"detailed", "comprehensive", "thorough", "extensive"},
		"brief":    {"brief", "short", "concise", "quick"},
	}
	
	for level, indicators := range detailIndicators {
		for _, indicator := range indicators {
			if strings.Contains(userQuery, indicator) {
				context.UserPreferences.ResponseDetailLevel = level
				break
			}
		}
	}
	
	// Detect code example preferences
	codeExampleIndicators := map[string][]string{
		"comprehensive": {"example", "code", "sample", "show me"},
		"minimal":       {"just", "simple", "basic"},
		"none":          {"no code", "without code", "text only"},
	}
	
	for level, indicators := range codeExampleIndicators {
		for _, indicator := range indicators {
			if strings.Contains(userQuery, indicator) {
				context.UserPreferences.CodeExamplePreference = level
				break
			}
		}
	}
}

// BuildEnhancedPrompt creates an enhanced prompt with historical context and preferences
func (ecm *EnhancedContextManager) BuildEnhancedPrompt(basePrompt string, nixosContext *config.NixOSContext, enhancedContext *EnhancedContext) string {
	builder := NewNixOSContextBuilder()
	
	// Start with base NixOS context
	contextualPrompt := builder.BuildContextualPrompt(basePrompt, nixosContext)
	
	// Add preferences guidance
	preferencesGuidance := ecm.buildPreferencesGuidance(enhancedContext.UserPreferences)
	if preferencesGuidance != "" {
		contextualPrompt += "\n\n" + preferencesGuidance
	}
	
	// Add historical context if available
	if len(enhancedContext.HistoricalInteractions) > 0 {
		historicalContext := ecm.buildHistoricalContext(enhancedContext.HistoricalInteractions)
		if historicalContext != "" {
			contextualPrompt += "\n\n" + historicalContext
		}
	}
	
	// Add session history if available
	if len(enhancedContext.SessionHistory) > 0 {
		sessionHistory := ecm.buildSessionHistory(enhancedContext.SessionHistory)
		if sessionHistory != "" {
			contextualPrompt += "\n\n" + sessionHistory
		}
	}
	
	return contextualPrompt
}

// buildPreferencesGuidance creates guidance based on user preferences
func (ecm *EnhancedContextManager) buildPreferencesGuidance(preferences Preferences) string {
	var guidance strings.Builder
	
	guidance.WriteString("=== USER PREFERENCES ===\n")
	
	if preferences.ResponseDetailLevel != "" {
		switch preferences.ResponseDetailLevel {
		case "detailed":
			guidance.WriteString("User prefers detailed explanations\n")
		case "brief":
			guidance.WriteString("User prefers brief explanations\n")
		default:
			guidance.WriteString("User prefers normal detail level\n")
		}
	}
	
	if preferences.CodeExamplePreference != "" {
		switch preferences.CodeExamplePreference {
		case "comprehensive":
			guidance.WriteString("User prefers comprehensive code examples\n")
		case "minimal":
			guidance.WriteString("User prefers minimal code examples\n")
		case "none":
			guidance.WriteString("User prefers no code examples\n")
		}
	}
	
	if len(preferences.FavoriteCommands) > 0 {
		guidance.WriteString(fmt.Sprintf("User frequently uses: %s\n", 
			strings.Join(preferences.FavoriteCommands, ", ")))
	}
	
	if len(preferences.AvoidedTopics) > 0 {
		guidance.WriteString(fmt.Sprintf("User avoids: %s\n", 
			strings.Join(preferences.AvoidedTopics, ", ")))
	}
	
	guidance.WriteString("=== END PREFERENCES ===")
	
	return guidance.String()
}

// buildHistoricalContext creates context from historical interactions
func (ecm *EnhancedContextManager) buildHistoricalContext(interactions []Interaction) string {
	if len(interactions) == 0 {
		return ""
	}
	
	var context strings.Builder
	context.WriteString("=== RECENT INTERACTIONS ===\n")
	
	// Include last 3 interactions for context
	startIndex := len(interactions) - 3
	if startIndex < 0 {
		startIndex = 0
	}
	
	for i := startIndex; i < len(interactions); i++ {
		interaction := interactions[i]
		context.WriteString(fmt.Sprintf("User: %s\n", interaction.UserQuery))
		context.WriteString(fmt.Sprintf("AI: %s\n", interaction.AIResponse))
		if i < len(interactions)-1 {
			context.WriteString("\n")
		}
	}
	
	context.WriteString("=== END INTERACTIONS ===")
	
	return context.String()
}

// buildSessionHistory creates context from session history
func (ecm *EnhancedContextManager) buildSessionHistory(sessionHistory []string) string {
	if len(sessionHistory) == 0 {
		return ""
	}
	
	var context strings.Builder
	context.WriteString("=== SESSION HISTORY ===\n")
	context.WriteString("Recent user queries in this session:\n")
	
	// Include last 5 queries
	startIndex := len(sessionHistory) - 5
	if startIndex < 0 {
		startIndex = 0
	}
	
	for i := startIndex; i < len(sessionHistory); i++ {
		context.WriteString(fmt.Sprintf("- %s\n", sessionHistory[i]))
	}
	
	context.WriteString("=== END SESSION HISTORY ===")
	
	return context.String()
}

// GetContextSummary returns a summary of the enhanced context
func (ecm *EnhancedContextManager) GetContextSummary(enhancedContext *EnhancedContext) string {
	if enhancedContext == nil {
		return "Enhanced context: None"
	}
	
	var parts []string
	
	// Add basic NixOS context summary
	builder := NewNixOSContextBuilder()
	nixosSummary := builder.GetContextSummary(enhancedContext.NixOSContext)
	if nixosSummary != "" {
		parts = append(parts, nixosSummary)
	}
	
	// Add interaction count
	if len(enhancedContext.HistoricalInteractions) > 0 {
		parts = append(parts, fmt.Sprintf("Interactions: %d", len(enhancedContext.HistoricalInteractions)))
	}
	
	// Add preference summary
	if len(enhancedContext.UserPreferences.FavoriteCommands) > 0 {
		parts = append(parts, fmt.Sprintf("Favorites: %s", 
			strings.Join(enhancedContext.UserPreferences.FavoriteCommands, ", ")))
	}
	
	return "Enhanced Context: " + strings.Join(parts, " | ")
}