// Package learning - Adaptive Learning Engine
// Advanced Learning System Phase 2.2 Implementation
package learning

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// AdaptiveLearningEngine provides personalized learning experiences based on user behavior and performance
type AdaptiveLearningEngine struct {
	userProfile     *UserProfile
	aiProvider      ai.Provider
	logger          *logger.Logger
	config          *config.UserConfig
	learningContext *LearningContext
}

// UserProfile tracks comprehensive user learning data and preferences
type UserProfile struct {
	UserID            string                     `json:"user_id"`
	Name              string                     `json:"name,omitempty"`
	SkillLevel        SkillLevel                 `json:"skill_level"`
	LearningStyle     LearningStyle              `json:"learning_style"`
	Preferences       LearningPreferences        `json:"preferences"`
	CompetencyMap     map[string]CompetencyLevel `json:"competency_map"`
	InteractionData   []UserInteraction          `json:"interaction_data"`
	LearningGoals     []LearningGoal             `json:"learning_goals"`
	WeakAreas         []string                   `json:"weak_areas"`
	StrengthAreas     []string                   `json:"strength_areas"`
	PersonalityType   LearnerPersonalityType     `json:"personality_type"`
	LastActive        time.Time                  `json:"last_active"`
	TotalLearningTime time.Duration              `json:"total_learning_time"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

// LearningContext provides contextual information for adaptive learning
type LearningContext struct {
	CurrentSession    *LearningSession      `json:"current_session"`
	RecentPerformance []PerformanceMetric   `json:"recent_performance"`
	EnvironmentInfo   SystemEnvironmentInfo `json:"environment_info"`
	TimeOfDay         string                `json:"time_of_day"`
	SessionDuration   time.Duration         `json:"session_duration"`
	DistractionLevel  DistractionLevel      `json:"distraction_level"`
}

// UserInteraction represents a single user learning interaction
type UserInteraction struct {
	ID               string                 `json:"id"`
	Timestamp        time.Time              `json:"timestamp"`
	InteractionType  InteractionType        `json:"interaction_type"`
	Content          string                 `json:"content"`
	Context          map[string]interface{} `json:"context"`
	Duration         time.Duration          `json:"duration"`
	Success          bool                   `json:"success"`
	Difficulty       DifficultyLevel        `json:"difficulty"`
	Engagement       EngagementLevel        `json:"engagement"`
	Outcome          InteractionOutcome     `json:"outcome"`
	MetricsCollected map[string]float64     `json:"metrics_collected"`
}

// LearningGoal represents a specific learning objective
type LearningGoal struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	TargetLevel CompetencyLevel `json:"target_level"`
	Deadline    *time.Time      `json:"deadline,omitempty"`
	Priority    Priority        `json:"priority"`
	Status      GoalStatus      `json:"status"`
	Progress    float64         `json:"progress"` // 0.0 to 1.0
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// LearningSession tracks a single learning session
type LearningSession struct {
	ID               string          `json:"id"`
	StartTime        time.Time       `json:"start_time"`
	EndTime          *time.Time      `json:"end_time,omitempty"`
	Topic            string          `json:"topic"`
	ModulesCompleted []string        `json:"modules_completed"`
	TimeSpent        time.Duration   `json:"time_spent"`
	PerformanceScore float64         `json:"performance_score"`
	EngagementLevel  EngagementLevel `json:"engagement_level"`
	Notes            string          `json:"notes,omitempty"`
	Achievements     []Achievement   `json:"achievements"`
}

// PerformanceMetric tracks user performance over time
type PerformanceMetric struct {
	Timestamp      time.Time     `json:"timestamp"`
	Topic          string        `json:"topic"`
	Score          float64       `json:"score"`
	CompletionTime time.Duration `json:"completion_time"`
	AttemptsNeeded int           `json:"attempts_needed"`
	HintsUsed      int           `json:"hints_used"`
	ErrorPatterns  []string      `json:"error_patterns"`
	SkillLevel     SkillLevel    `json:"skill_level"`
}

// SystemEnvironmentInfo captures system context for learning adaptation
type SystemEnvironmentInfo struct {
	NixOSVersion      string        `json:"nixos_version"`
	ConfigType        string        `json:"config_type"` // flake, channels, etc.
	InstalledPackages []string      `json:"installed_packages"`
	ActiveServices    []string      `json:"active_services"`
	SystemResources   ResourceInfo  `json:"system_resources"`
	ErrorHistory      []SystemError `json:"error_history"`
}

// Achievement represents a learning milestone
type Achievement struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url,omitempty"`
	EarnedAt    time.Time `json:"earned_at"`
	Points      int       `json:"points"`
	Rarity      Rarity    `json:"rarity"`
}

// Enums and types for classification
type SkillLevel string

const (
	SkillBeginner     SkillLevel = "beginner"
	SkillIntermediate SkillLevel = "intermediate"
	SkillAdvanced     SkillLevel = "advanced"
	SkillExpert       SkillLevel = "expert"
)

type LearningStyle string

const (
	StyleVisual      LearningStyle = "visual"
	StyleAuditory    LearningStyle = "auditory"
	StyleKinesthetic LearningStyle = "kinesthetic"
	StyleReading     LearningStyle = "reading"
	StyleMultimodal  LearningStyle = "multimodal"
)

type CompetencyLevel string

const (
	CompetencyNone         CompetencyLevel = "none"
	CompetencyBeginner     CompetencyLevel = "beginner"
	CompetencyBasic        CompetencyLevel = "basic"
	CompetencyIntermediate CompetencyLevel = "intermediate"
	CompetencyAdvanced     CompetencyLevel = "advanced"
	CompetencyExpert       CompetencyLevel = "expert"
	CompetencyMastery      CompetencyLevel = "mastery"
)

type InteractionType string

const (
	InteractionQuiz        InteractionType = "quiz"
	InteractionTutorial    InteractionType = "tutorial"
	InteractionPractice    InteractionType = "practice"
	InteractionQuestion    InteractionType = "question"
	InteractionExploration InteractionType = "exploration"
	InteractionAssessment  InteractionType = "assessment"
	InteractionReview      InteractionType = "review"
)

type DifficultyLevel string

const (
	DifficultyVeryEasy DifficultyLevel = "very_easy"
	DifficultyEasy     DifficultyLevel = "easy"
	DifficultyMedium   DifficultyLevel = "medium"
	DifficultyHard     DifficultyLevel = "hard"
	DifficultyVeryHard DifficultyLevel = "very_hard"
)

type EngagementLevel string

const (
	EngagementLow    EngagementLevel = "low"
	EngagementMedium EngagementLevel = "medium"
	EngagementHigh   EngagementLevel = "high"
)

type DistractionLevel string

const (
	DistractionLow    DistractionLevel = "low"
	DistractionMedium DistractionLevel = "medium"
	DistractionHigh   DistractionLevel = "high"
)

type InteractionOutcome string

const (
	OutcomeSuccess   InteractionOutcome = "success"
	OutcomePartial   InteractionOutcome = "partial"
	OutcomeFailure   InteractionOutcome = "failure"
	OutcomeAbandoned InteractionOutcome = "abandoned"
	OutcomeSkipped   InteractionOutcome = "skipped"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

type GoalStatus string

const (
	GoalActive    GoalStatus = "active"
	GoalCompleted GoalStatus = "completed"
	GoalPaused    GoalStatus = "paused"
	GoalCancelled GoalStatus = "cancelled"
)

type Rarity string

const (
	RarityCommon    Rarity = "common"
	RarityUncommon  Rarity = "uncommon"
	RarityRare      Rarity = "rare"
	RarityEpic      Rarity = "epic"
	RarityLegendary Rarity = "legendary"
)

type LearnerPersonalityType string

const (
	PersonalityExplorer   LearnerPersonalityType = "explorer"
	PersonalityAchiever   LearnerPersonalityType = "achiever"
	PersonalitySocializer LearnerPersonalityType = "socializer"
	PersonalityKiller     LearnerPersonalityType = "killer"
	PersonalityPragmatist LearnerPersonalityType = "pragmatist"
)

// Supporting structures
type LearningPreferences struct {
	PreferredDifficulty  DifficultyLevel `json:"preferred_difficulty"`
	SessionLength        time.Duration   `json:"session_length"`
	NotificationsEnabled bool            `json:"notifications_enabled"`
	ProgressSharing      bool            `json:"progress_sharing"`
	AdaptiveDifficulty   bool            `json:"adaptive_difficulty"`
	GameificationEnabled bool            `json:"gamification_enabled"`
	PreferredTopics      []string        `json:"preferred_topics"`
	AvoidedTopics        []string        `json:"avoided_topics"`
	LearningTimes        []TimeWindow    `json:"learning_times"`
}

type TimeWindow struct {
	DayOfWeek string        `json:"day_of_week"`
	StartTime string        `json:"start_time"`
	EndTime   string        `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

type ResourceInfo struct {
	CPUCores     int     `json:"cpu_cores"`
	MemoryGB     float64 `json:"memory_gb"`
	DiskSpaceGB  float64 `json:"disk_space_gb"`
	NetworkSpeed string  `json:"network_speed"`
}

type SystemError struct {
	Timestamp time.Time `json:"timestamp"`
	ErrorType string    `json:"error_type"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
	Context   string    `json:"context"`
}

// NewAdaptiveLearningEngine creates a new adaptive learning engine
func NewAdaptiveLearningEngine(aiProvider ai.Provider, cfg *config.UserConfig, log *logger.Logger) (*AdaptiveLearningEngine, error) {
	if aiProvider == nil {
		return nil, fmt.Errorf("AI provider cannot be nil")
	}

	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	engine := &AdaptiveLearningEngine{
		aiProvider: aiProvider,
		logger:     log,
		config:     cfg,
		learningContext: &LearningContext{
			TimeOfDay:        getCurrentTimeOfDay(),
			DistractionLevel: DistractionLow,
		},
	}

	// Load or initialize user profile
	if err := engine.loadUserProfile(); err != nil {
		engine.logger.Warn(fmt.Sprintf("Failed to load user profile, creating new one: %v", err))
		engine.createNewUserProfile()
	}

	return engine, nil
}

// LoadUserProfile loads user profile from disk
func (ale *AdaptiveLearningEngine) loadUserProfile() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	profilePath := filepath.Join(configDir, "nixai", "learning", "user_profile.json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to read user profile: %w", err)
	}

	var profile UserProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("failed to unmarshal user profile: %w", err)
	}

	ale.userProfile = &profile
	return nil
}

// SaveUserProfile saves user profile to disk
func (ale *AdaptiveLearningEngine) SaveUserProfile() error {
	if ale.userProfile == nil {
		return fmt.Errorf("no user profile to save")
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	profileDir := filepath.Join(configDir, "nixai", "learning")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	ale.userProfile.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(ale.userProfile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %w", err)
	}

	profilePath := filepath.Join(profileDir, "user_profile.json")
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write user profile: %w", err)
	}

	ale.logger.Debug("User profile saved successfully")
	return nil
}

// CreateNewUserProfile initializes a new user profile
func (ale *AdaptiveLearningEngine) createNewUserProfile() {
	ale.userProfile = &UserProfile{
		UserID:            utils.HashString(fmt.Sprintf("nixai_user_%d", time.Now().Unix())),
		SkillLevel:        SkillBeginner,
		LearningStyle:     StyleMultimodal,
		CompetencyMap:     make(map[string]CompetencyLevel),
		InteractionData:   []UserInteraction{},
		LearningGoals:     []LearningGoal{},
		WeakAreas:         []string{},
		StrengthAreas:     []string{},
		PersonalityType:   PersonalityExplorer,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		LastActive:        time.Now(),
		TotalLearningTime: 0,
		Preferences: LearningPreferences{
			PreferredDifficulty:  DifficultyMedium,
			SessionLength:        30 * time.Minute,
			NotificationsEnabled: true,
			ProgressSharing:      false,
			AdaptiveDifficulty:   true,
			GameificationEnabled: true,
			PreferredTopics:      []string{"basics", "configuration"},
			AvoidedTopics:        []string{},
			LearningTimes:        []TimeWindow{},
		},
	}

	// Initialize basic competency map
	ale.initializeCompetencyMap()
}

// InitializeCompetencyMap sets up initial competency tracking
func (ale *AdaptiveLearningEngine) initializeCompetencyMap() {
	basicTopics := []string{
		"nix-basics", "nixos-basics", "configuration-management",
		"package-management", "service-management", "flakes",
		"modules", "overlays", "derivations", "channels",
		"home-manager", "networking", "security", "troubleshooting",
		"development-environments", "containerization", "virtualization",
	}

	for _, topic := range basicTopics {
		ale.userProfile.CompetencyMap[topic] = CompetencyNone
	}
}

// RecordInteraction records a user learning interaction
func (ale *AdaptiveLearningEngine) RecordInteraction(interactionType InteractionType, content string, context map[string]interface{}, duration time.Duration, success bool) {
	interaction := UserInteraction{
		ID:               utils.HashString(fmt.Sprintf("%s_%d", string(interactionType), time.Now().UnixNano())),
		Timestamp:        time.Now(),
		InteractionType:  interactionType,
		Content:          content,
		Context:          context,
		Duration:         duration,
		Success:          success,
		Difficulty:       ale.getDynamicDifficulty(),
		Engagement:       ale.calculateEngagement(duration, success),
		Outcome:          ale.determineOutcome(success, duration),
		MetricsCollected: ale.collectMetrics(duration, success),
	}

	ale.userProfile.InteractionData = append(ale.userProfile.InteractionData, interaction)
	ale.userProfile.TotalLearningTime += duration
	ale.userProfile.LastActive = time.Now()

	// Update competency based on interaction
	ale.updateCompetencyFromInteraction(interaction)

	// Limit interaction history to last 1000 entries
	if len(ale.userProfile.InteractionData) > 1000 {
		ale.userProfile.InteractionData = ale.userProfile.InteractionData[len(ale.userProfile.InteractionData)-1000:]
	}

	ale.logger.Debug(fmt.Sprintf("Recorded interaction: %s", interaction.ID))
}

// GetPersonalizedContent generates personalized learning content
func (ale *AdaptiveLearningEngine) GetPersonalizedContent(ctx context.Context, topic string, targetDifficulty DifficultyLevel) (*PersonalizedContent, error) {
	if ale.userProfile == nil {
		return nil, fmt.Errorf("user profile not initialized")
	}

	// Analyze user's current competency in the topic
	currentCompetency := ale.getTopicCompetency(topic)

	// Determine optimal difficulty based on user profile and target
	optimalDifficulty := ale.calculateOptimalDifficulty(topic, targetDifficulty)

	// Build context for AI generation
	learningContext := ale.buildLearningContext(topic, currentCompetency, optimalDifficulty)

	// Generate personalized content using AI
	prompt := ale.buildPersonalizationPrompt(topic, learningContext)

	response, err := ale.aiProvider.Query(prompt)
	if err != nil {
		ale.logger.Error(fmt.Sprintf("Failed to generate personalized content: %v", err))
		return ale.getFallbackContent(topic, optimalDifficulty), nil
	}

	content := &PersonalizedContent{
		Topic:               topic,
		Difficulty:          optimalDifficulty,
		LearningStyle:       ale.userProfile.LearningStyle,
		Content:             response,
		EstimatedDuration:   ale.estimateContentDuration(response),
		Prerequisites:       ale.getTopicPrerequisites(topic),
		LearningObjectives:  ale.generateLearningObjectives(topic, currentCompetency),
		NextSteps:           ale.suggestNextSteps(topic, currentCompetency),
		PersonalizationNote: ale.generatePersonalizationNote(learningContext),
		CreatedAt:           time.Now(),
	}

	return content, nil
}

// PersonalizedContent represents AI-generated personalized learning content
type PersonalizedContent struct {
	Topic               string          `json:"topic"`
	Difficulty          DifficultyLevel `json:"difficulty"`
	LearningStyle       LearningStyle   `json:"learning_style"`
	Content             string          `json:"content"`
	EstimatedDuration   time.Duration   `json:"estimated_duration"`
	Prerequisites       []string        `json:"prerequisites"`
	LearningObjectives  []string        `json:"learning_objectives"`
	NextSteps           []string        `json:"next_steps"`
	PersonalizationNote string          `json:"personalization_note"`
	CreatedAt           time.Time       `json:"created_at"`
}

// Helper methods
func getCurrentTimeOfDay() string {
	hour := time.Now().Hour()
	switch {
	case hour < 6:
		return "early_morning"
	case hour < 12:
		return "morning"
	case hour < 18:
		return "afternoon"
	case hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (ale *AdaptiveLearningEngine) getDynamicDifficulty() DifficultyLevel {
	if ale.userProfile.Preferences.AdaptiveDifficulty {
		// Calculate based on recent performance
		recentInteractions := ale.getRecentInteractions(10)
		if len(recentInteractions) == 0 {
			return ale.userProfile.Preferences.PreferredDifficulty
		}

		successRate := ale.calculateSuccessRate(recentInteractions)
		switch {
		case successRate > 0.8:
			return ale.increaseDifficulty(ale.userProfile.Preferences.PreferredDifficulty)
		case successRate < 0.4:
			return ale.decreaseDifficulty(ale.userProfile.Preferences.PreferredDifficulty)
		default:
			return ale.userProfile.Preferences.PreferredDifficulty
		}
	}
	return ale.userProfile.Preferences.PreferredDifficulty
}

func (ale *AdaptiveLearningEngine) calculateEngagement(duration time.Duration, success bool) EngagementLevel {
	expectedDuration := ale.userProfile.Preferences.SessionLength
	engagementScore := float64(duration) / float64(expectedDuration)

	if success {
		engagementScore *= 1.2 // Boost for success
	}

	switch {
	case engagementScore < 0.5:
		return EngagementLow
	case engagementScore < 1.5:
		return EngagementMedium
	default:
		return EngagementHigh
	}
}

func (ale *AdaptiveLearningEngine) determineOutcome(success bool, duration time.Duration) InteractionOutcome {
	if duration < 10*time.Second {
		return OutcomeSkipped
	}
	if success {
		return OutcomeSuccess
	}
	if duration > ale.userProfile.Preferences.SessionLength*2 {
		return OutcomeAbandoned
	}
	return OutcomeFailure
}

func (ale *AdaptiveLearningEngine) collectMetrics(duration time.Duration, success bool) map[string]float64 {
	metrics := make(map[string]float64)
	metrics["duration_minutes"] = duration.Minutes()
	metrics["success_rate"] = func() float64 {
		if success {
			return 1.0
		}
		return 0.0
	}()
	metrics["engagement_score"] = ale.calculateEngagementScore(duration, success)
	return metrics
}

func (ale *AdaptiveLearningEngine) calculateEngagementScore(duration time.Duration, success bool) float64 {
	baseScore := math.Min(duration.Minutes()/30.0, 2.0) // Cap at 2 hours
	if success {
		baseScore *= 1.5
	}
	return math.Min(baseScore, 10.0) // Cap at 10
}

func (ale *AdaptiveLearningEngine) updateCompetencyFromInteraction(interaction UserInteraction) {
	// Extract topic from interaction context
	topic := ale.extractTopicFromContext(interaction)
	if topic == "" {
		return
	}

	currentLevel := ale.userProfile.CompetencyMap[topic]

	// Calculate competency change based on success and difficulty
	change := ale.calculateCompetencyChange(interaction)

	newLevel := ale.adjustCompetencyLevel(currentLevel, change)
	ale.userProfile.CompetencyMap[topic] = newLevel

	// Update strength/weakness areas
	ale.updateStrengthWeaknessAreas(topic, newLevel, interaction.Success)
}

func (ale *AdaptiveLearningEngine) extractTopicFromContext(interaction UserInteraction) string {
	if topic, exists := interaction.Context["topic"]; exists {
		if topicStr, ok := topic.(string); ok {
			return topicStr
		}
	}
	// Try to infer from content
	return ale.inferTopicFromContent(interaction.Content)
}

func (ale *AdaptiveLearningEngine) inferTopicFromContent(content string) string {
	// Simple keyword-based topic inference
	topicKeywords := map[string][]string{
		"nix-basics":               {"nix", "basics", "introduction"},
		"nixos-basics":             {"nixos", "operating system", "os"},
		"configuration-management": {"configuration", "config", "settings"},
		"package-management":       {"package", "packages", "install", "search"},
		"service-management":       {"service", "services", "systemd", "enable"},
		"flakes":                   {"flake", "flakes", "flake.nix"},
		"modules":                  {"module", "modules", "nixos module"},
		"home-manager":             {"home-manager", "home manager", "hm"},
	}

	content = strings.ToLower(content)
	for topic, keywords := range topicKeywords {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return topic
			}
		}
	}
	return ""
}

func (ale *AdaptiveLearningEngine) calculateCompetencyChange(interaction UserInteraction) float64 {
	baseChange := 0.1

	// Adjust for success/failure
	if interaction.Success {
		baseChange *= 1.0
	} else {
		baseChange *= -0.3
	}

	// Adjust for difficulty
	difficultyMultiplier := map[DifficultyLevel]float64{
		DifficultyVeryEasy: 0.5,
		DifficultyEasy:     0.7,
		DifficultyMedium:   1.0,
		DifficultyHard:     1.3,
		DifficultyVeryHard: 1.5,
	}

	if multiplier, exists := difficultyMultiplier[interaction.Difficulty]; exists {
		baseChange *= multiplier
	}

	// Adjust for engagement
	engagementMultiplier := map[EngagementLevel]float64{
		EngagementLow:    0.8,
		EngagementMedium: 1.0,
		EngagementHigh:   1.2,
	}

	if multiplier, exists := engagementMultiplier[interaction.Engagement]; exists {
		baseChange *= multiplier
	}

	return baseChange
}

func (ale *AdaptiveLearningEngine) adjustCompetencyLevel(current CompetencyLevel, change float64) CompetencyLevel {
	levels := []CompetencyLevel{
		CompetencyNone,
		CompetencyBeginner,
		CompetencyBasic,
		CompetencyIntermediate,
		CompetencyAdvanced,
		CompetencyExpert,
		CompetencyMastery,
	}

	currentIndex := 0
	for i, level := range levels {
		if level == current {
			currentIndex = i
			break
		}
	}

	// Convert change to level adjustment
	levelChange := int(change * 10) // Rough conversion
	newIndex := currentIndex + levelChange

	// Clamp to valid range
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(levels) {
		newIndex = len(levels) - 1
	}

	return levels[newIndex]
}

func (ale *AdaptiveLearningEngine) updateStrengthWeaknessAreas(topic string, level CompetencyLevel, success bool) {
	// Remove from current lists
	ale.removeFromSlice(&ale.userProfile.StrengthAreas, topic)
	ale.removeFromSlice(&ale.userProfile.WeakAreas, topic)

	// Add to appropriate list based on competency and recent performance
	if level >= CompetencyAdvanced && success {
		ale.userProfile.StrengthAreas = append(ale.userProfile.StrengthAreas, topic)
	} else if level <= CompetencyBasic || !success {
		ale.userProfile.WeakAreas = append(ale.userProfile.WeakAreas, topic)
	}

	// Keep lists manageable
	if len(ale.userProfile.StrengthAreas) > 10 {
		ale.userProfile.StrengthAreas = ale.userProfile.StrengthAreas[len(ale.userProfile.StrengthAreas)-10:]
	}
	if len(ale.userProfile.WeakAreas) > 10 {
		ale.userProfile.WeakAreas = ale.userProfile.WeakAreas[len(ale.userProfile.WeakAreas)-10:]
	}
}

func (ale *AdaptiveLearningEngine) removeFromSlice(slice *[]string, item string) {
	for i, v := range *slice {
		if v == item {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			break
		}
	}
}

func (ale *AdaptiveLearningEngine) getRecentInteractions(count int) []UserInteraction {
	if len(ale.userProfile.InteractionData) <= count {
		return ale.userProfile.InteractionData
	}
	return ale.userProfile.InteractionData[len(ale.userProfile.InteractionData)-count:]
}

func (ale *AdaptiveLearningEngine) calculateSuccessRate(interactions []UserInteraction) float64 {
	if len(interactions) == 0 {
		return 0.5 // Default middle value
	}

	successCount := 0
	for _, interaction := range interactions {
		if interaction.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(interactions))
}

func (ale *AdaptiveLearningEngine) increaseDifficulty(current DifficultyLevel) DifficultyLevel {
	switch current {
	case DifficultyVeryEasy:
		return DifficultyEasy
	case DifficultyEasy:
		return DifficultyMedium
	case DifficultyMedium:
		return DifficultyHard
	case DifficultyHard:
		return DifficultyVeryHard
	default:
		return current
	}
}

func (ale *AdaptiveLearningEngine) decreaseDifficulty(current DifficultyLevel) DifficultyLevel {
	switch current {
	case DifficultyVeryHard:
		return DifficultyHard
	case DifficultyHard:
		return DifficultyMedium
	case DifficultyMedium:
		return DifficultyEasy
	case DifficultyEasy:
		return DifficultyVeryEasy
	default:
		return current
	}
}

// Additional helper methods for content generation
func (ale *AdaptiveLearningEngine) getTopicCompetency(topic string) CompetencyLevel {
	if level, exists := ale.userProfile.CompetencyMap[topic]; exists {
		return level
	}
	return CompetencyNone
}

func (ale *AdaptiveLearningEngine) calculateOptimalDifficulty(topic string, target DifficultyLevel) DifficultyLevel {
	currentCompetency := ale.getTopicCompetency(topic)

	// Map competency to difficulty
	competencyDifficultyMap := map[CompetencyLevel]DifficultyLevel{
		CompetencyNone:         DifficultyVeryEasy,
		CompetencyBeginner:     DifficultyEasy,
		CompetencyBasic:        DifficultyMedium,
		CompetencyIntermediate: DifficultyMedium,
		CompetencyAdvanced:     DifficultyHard,
		CompetencyExpert:       DifficultyVeryHard,
		CompetencyMastery:      DifficultyVeryHard,
	}

	optimal := competencyDifficultyMap[currentCompetency]

	// Consider target difficulty
	if target != "" {
		// Weighted average of optimal and target
		return ale.blendDifficulties(optimal, target, 0.7) // 70% weight to optimal
	}

	return optimal
}

func (ale *AdaptiveLearningEngine) blendDifficulties(d1, d2 DifficultyLevel, weight float64) DifficultyLevel {
	difficulties := []DifficultyLevel{
		DifficultyVeryEasy, DifficultyEasy, DifficultyMedium, DifficultyHard, DifficultyVeryHard,
	}

	index1, index2 := 0, 0
	for i, d := range difficulties {
		if d == d1 {
			index1 = i
		}
		if d == d2 {
			index2 = i
		}
	}

	blendedIndex := int(float64(index1)*weight + float64(index2)*(1-weight))
	if blendedIndex >= len(difficulties) {
		blendedIndex = len(difficulties) - 1
	}

	return difficulties[blendedIndex]
}

func (ale *AdaptiveLearningEngine) buildLearningContext(topic string, competency CompetencyLevel, difficulty DifficultyLevel) map[string]interface{} {
	context := map[string]interface{}{
		"topic":               topic,
		"user_competency":     competency,
		"target_difficulty":   difficulty,
		"user_skill_level":    ale.userProfile.SkillLevel,
		"learning_style":      ale.userProfile.LearningStyle,
		"personality_type":    ale.userProfile.PersonalityType,
		"weak_areas":          ale.userProfile.WeakAreas,
		"strength_areas":      ale.userProfile.StrengthAreas,
		"recent_performance":  ale.analyzeRecentPerformance(),
		"session_preferences": ale.userProfile.Preferences,
		"time_of_day":         ale.learningContext.TimeOfDay,
		"session_duration":    ale.learningContext.SessionDuration,
	}

	return context
}

func (ale *AdaptiveLearningEngine) analyzeRecentPerformance() map[string]interface{} {
	recent := ale.getRecentInteractions(20)
	if len(recent) == 0 {
		return map[string]interface{}{
			"success_rate": 0.5,
			"avg_duration": 0,
			"trend":        "stable",
		}
	}

	successRate := ale.calculateSuccessRate(recent)
	avgDuration := ale.calculateAverageDuration(recent)
	trend := ale.calculatePerformanceTrend(recent)

	return map[string]interface{}{
		"success_rate":       successRate,
		"avg_duration":       avgDuration.Minutes(),
		"trend":              trend,
		"total_interactions": len(recent),
	}
}

func (ale *AdaptiveLearningEngine) calculateAverageDuration(interactions []UserInteraction) time.Duration {
	if len(interactions) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, interaction := range interactions {
		total += interaction.Duration
	}

	return total / time.Duration(len(interactions))
}

func (ale *AdaptiveLearningEngine) calculatePerformanceTrend(interactions []UserInteraction) string {
	if len(interactions) < 5 {
		return "insufficient_data"
	}

	// Calculate success rate for first half vs second half
	midPoint := len(interactions) / 2
	firstHalf := interactions[:midPoint]
	secondHalf := interactions[midPoint:]

	firstSuccess := ale.calculateSuccessRate(firstHalf)
	secondSuccess := ale.calculateSuccessRate(secondHalf)

	diff := secondSuccess - firstSuccess
	switch {
	case diff > 0.1:
		return "improving"
	case diff < -0.1:
		return "declining"
	default:
		return "stable"
	}
}

func (ale *AdaptiveLearningEngine) buildPersonalizationPrompt(topic string, context map[string]interface{}) string {
	prompt := fmt.Sprintf(`Create personalized NixOS learning content for the topic: %s

User Profile:
- Skill Level: %v
- Learning Style: %v
- Competency in %s: %v
- Personality Type: %v
- Weak Areas: %v
- Strength Areas: %v

Performance Context:
- Recent Performance: %v
- Preferred Difficulty: %v
- Session Duration: %v minutes
- Time of Day: %v

Requirements:
1. Create content appropriate for the user's competency level
2. Adapt to their learning style (visual/auditory/kinesthetic/reading/multimodal)
3. Address their weak areas while building on strengths
4. Include practical examples and exercises
5. Provide clear learning objectives
6. Suggest next steps for continued learning

Format the response as structured learning content with:
- Introduction and learning objectives
- Main content with examples
- Practice exercises
- Summary and next steps

Make it engaging and personalized to their profile.`,
		topic,
		context["user_skill_level"],
		context["learning_style"],
		topic,
		context["user_competency"],
		context["personality_type"],
		context["weak_areas"],
		context["strength_areas"],
		context["recent_performance"],
		context["target_difficulty"],
		context["session_duration"],
		context["time_of_day"],
	)

	return prompt
}

func (ale *AdaptiveLearningEngine) getFallbackContent(topic string, difficulty DifficultyLevel) *PersonalizedContent {
	// Provide basic fallback content when AI generation fails
	content := &PersonalizedContent{
		Topic:               topic,
		Difficulty:          difficulty,
		LearningStyle:       ale.userProfile.LearningStyle,
		Content:             ale.generateBasicContent(topic, difficulty),
		EstimatedDuration:   30 * time.Minute,
		Prerequisites:       ale.getTopicPrerequisites(topic),
		LearningObjectives:  ale.generateLearningObjectives(topic, ale.getTopicCompetency(topic)),
		NextSteps:           ale.suggestNextSteps(topic, ale.getTopicCompetency(topic)),
		PersonalizationNote: "Basic content - AI personalization unavailable",
		CreatedAt:           time.Now(),
	}

	return content
}

func (ale *AdaptiveLearningEngine) generateBasicContent(topic string, difficulty DifficultyLevel) string {
	// Basic content templates for different topics
	templates := map[string]string{
		"nix-basics":   "Learn the fundamentals of Nix package manager: declarative package management, Nix expressions, and basic commands.",
		"nixos-basics": "Introduction to NixOS operating system: configuration.nix, system configuration, and rebuilding your system.",
		"flakes":       "Understanding Nix flakes: flake.nix structure, inputs and outputs, and modern Nix project management.",
		"modules":      "Working with NixOS modules: creating custom modules, module system, and configuration options.",
	}

	if template, exists := templates[topic]; exists {
		return fmt.Sprintf("# %s\n\n%s\n\nDifficulty: %s", topic, template, difficulty)
	}

	return fmt.Sprintf("# %s\n\nLearning content for %s at %s difficulty level.", topic, topic, difficulty)
}

func (ale *AdaptiveLearningEngine) estimateContentDuration(content string) time.Duration {
	// Rough estimation based on content length
	words := len(strings.Fields(content))
	readingSpeed := 200 // words per minute
	minutes := words / readingSpeed
	if minutes < 5 {
		minutes = 5 // Minimum 5 minutes
	}
	if minutes > 60 {
		minutes = 60 // Maximum 1 hour
	}
	return time.Duration(minutes) * time.Minute
}

func (ale *AdaptiveLearningEngine) getTopicPrerequisites(topic string) []string {
	prerequisites := map[string][]string{
		"nix-basics":               {},
		"nixos-basics":             {"nix-basics"},
		"configuration-management": {"nixos-basics"},
		"package-management":       {"nix-basics"},
		"service-management":       {"nixos-basics", "configuration-management"},
		"flakes":                   {"nix-basics", "nixos-basics"},
		"modules":                  {"nixos-basics", "configuration-management"},
		"overlays":                 {"nix-basics", "package-management"},
		"home-manager":             {"nixos-basics", "configuration-management"},
	}

	if prereqs, exists := prerequisites[topic]; exists {
		return prereqs
	}
	return []string{}
}

func (ale *AdaptiveLearningEngine) generateLearningObjectives(topic string, competency CompetencyLevel) []string {
	objectives := map[string]map[CompetencyLevel][]string{
		"nix-basics": {
			CompetencyNone: {
				"Understand what Nix is and its benefits",
				"Learn basic Nix commands",
				"Install your first package with Nix",
			},
			CompetencyBeginner: {
				"Use Nix expressions effectively",
				"Manage package environments",
				"Understand derivations",
			},
			CompetencyBasic: {
				"Write custom Nix expressions",
				"Use overlays for package modifications",
				"Manage multiple package versions",
			},
		},
		"nixos-basics": {
			CompetencyNone: {
				"Understand NixOS philosophy",
				"Navigate configuration.nix",
				"Perform basic system rebuilds",
			},
			CompetencyBeginner: {
				"Configure system services",
				"Manage users and groups",
				"Handle system updates safely",
			},
		},
	}

	if topicObjectives, exists := objectives[topic]; exists {
		if levelObjectives, exists := topicObjectives[competency]; exists {
			return levelObjectives
		}
	}

	// Default objectives
	return []string{
		fmt.Sprintf("Learn fundamental concepts of %s", topic),
		fmt.Sprintf("Apply %s knowledge in practical scenarios", topic),
		fmt.Sprintf("Build confidence with %s tools and techniques", topic),
	}
}

func (ale *AdaptiveLearningEngine) suggestNextSteps(topic string, competency CompetencyLevel) []string {
	nextSteps := map[string]map[CompetencyLevel][]string{
		"nix-basics": {
			CompetencyNone: {
				"Practice with nix-shell environments",
				"Explore the Nix packages collection",
				"Learn about NixOS fundamentals",
			},
			CompetencyBeginner: {
				"Study NixOS configuration management",
				"Experiment with Nix flakes",
				"Learn about Home Manager",
			},
		},
		"nixos-basics": {
			CompetencyNone: {
				"Set up a test NixOS environment",
				"Practice system configuration changes",
				"Learn about NixOS modules",
			},
			CompetencyBeginner: {
				"Explore advanced service configuration",
				"Learn about networking in NixOS",
				"Study security hardening",
			},
		},
	}

	if topicSteps, exists := nextSteps[topic]; exists {
		if levelSteps, exists := topicSteps[competency]; exists {
			return levelSteps
		}
	}

	// Default next steps
	return []string{
		fmt.Sprintf("Practice %s concepts with hands-on exercises", topic),
		fmt.Sprintf("Explore advanced %s topics", topic),
		"Join the NixOS community for support and discussions",
	}
}

func (ale *AdaptiveLearningEngine) generatePersonalizationNote(context map[string]interface{}) string {
	notes := []string{}

	if learningStyle, ok := context["learning_style"].(LearningStyle); ok {
		switch learningStyle {
		case StyleVisual:
			notes = append(notes, "Content optimized for visual learning with diagrams and examples")
		case StyleAuditory:
			notes = append(notes, "Content structured for auditory learning with clear explanations")
		case StyleKinesthetic:
			notes = append(notes, "Hands-on exercises included for kinesthetic learning")
		case StyleReading:
			notes = append(notes, "Detailed reading materials provided")
		case StyleMultimodal:
			notes = append(notes, "Mixed learning approaches for comprehensive understanding")
		}
	}

	if recentPerf, ok := context["recent_performance"].(map[string]interface{}); ok {
		if trend, ok := recentPerf["trend"].(string); ok {
			switch trend {
			case "improving":
				notes = append(notes, "Content slightly increased in difficulty due to improving performance")
			case "declining":
				notes = append(notes, "Content adjusted to provide more support and reinforcement")
			case "stable":
				notes = append(notes, "Content difficulty maintained based on consistent performance")
			}
		}
	}

	if len(notes) == 0 {
		return "Content personalized based on your learning profile"
	}

	return strings.Join(notes, ". ")
}

// GetUserProfile returns the current user profile
func (ale *AdaptiveLearningEngine) GetUserProfile() *UserProfile {
	return ale.userProfile
}

// UpdateSkillLevel updates the user's overall skill level
func (ale *AdaptiveLearningEngine) UpdateSkillLevel(level SkillLevel) {
	ale.userProfile.SkillLevel = level
	ale.userProfile.UpdatedAt = time.Now()
}

// SetLearningGoal adds a new learning goal
func (ale *AdaptiveLearningEngine) SetLearningGoal(title, description string, targetLevel CompetencyLevel, deadline *time.Time, priority Priority) {
	goal := LearningGoal{
		ID:          utils.HashString(fmt.Sprintf("%s_%d", title, time.Now().Unix())),
		Title:       title,
		Description: description,
		TargetLevel: targetLevel,
		Deadline:    deadline,
		Priority:    priority,
		Status:      GoalActive,
		Progress:    0.0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ale.userProfile.LearningGoals = append(ale.userProfile.LearningGoals, goal)
	ale.userProfile.UpdatedAt = time.Now()
}

// GetActiveGoals returns all active learning goals
func (ale *AdaptiveLearningEngine) GetActiveGoals() []LearningGoal {
	var activeGoals []LearningGoal
	for _, goal := range ale.userProfile.LearningGoals {
		if goal.Status == GoalActive {
			activeGoals = append(activeGoals, goal)
		}
	}

	// Sort by priority and deadline
	sort.Slice(activeGoals, func(i, j int) bool {
		// Priority comparison
		priorityOrder := map[Priority]int{
			PriorityCritical: 4,
			PriorityHigh:     3,
			PriorityMedium:   2,
			PriorityLow:      1,
		}

		if priorityOrder[activeGoals[i].Priority] != priorityOrder[activeGoals[j].Priority] {
			return priorityOrder[activeGoals[i].Priority] > priorityOrder[activeGoals[j].Priority]
		}

		// If same priority, sort by deadline
		if activeGoals[i].Deadline != nil && activeGoals[j].Deadline != nil {
			return activeGoals[i].Deadline.Before(*activeGoals[j].Deadline)
		}

		return false
	})

	return activeGoals
}

// UpdateGoalProgress updates progress for a specific goal
func (ale *AdaptiveLearningEngine) UpdateGoalProgress(goalID string, progress float64) {
	for i, goal := range ale.userProfile.LearningGoals {
		if goal.ID == goalID {
			ale.userProfile.LearningGoals[i].Progress = math.Max(0.0, math.Min(1.0, progress))
			ale.userProfile.LearningGoals[i].UpdatedAt = time.Now()

			// Mark as completed if progress reaches 100%
			if progress >= 1.0 {
				ale.userProfile.LearningGoals[i].Status = GoalCompleted
			}

			ale.userProfile.UpdatedAt = time.Now()
			break
		}
	}
}

// GetLearningRecommendations provides personalized learning recommendations
func (ale *AdaptiveLearningEngine) GetLearningRecommendations() []LearningRecommendation {
	recommendations := []LearningRecommendation{}

	// Recommend based on weak areas
	for _, weakArea := range ale.userProfile.WeakAreas {
		recommendations = append(recommendations, LearningRecommendation{
			Type:          RecommendationWeakArea,
			Topic:         weakArea,
			Reason:        fmt.Sprintf("Focus on %s to improve weak areas", weakArea),
			Priority:      PriorityHigh,
			Difficulty:    ale.getRecommendedDifficulty(weakArea),
			EstimatedTime: 30 * time.Minute,
		})
	}

	// Recommend based on learning goals
	activeGoals := ale.GetActiveGoals()
	for _, goal := range activeGoals {
		if goal.Progress < 1.0 {
			recommendations = append(recommendations, LearningRecommendation{
				Type:          RecommendationGoal,
				Topic:         goal.Title,
				Reason:        fmt.Sprintf("Continue working towards goal: %s", goal.Description),
				Priority:      goal.Priority,
				Difficulty:    ale.difficultyFromCompetency(goal.TargetLevel),
				EstimatedTime: 45 * time.Minute,
			})
		}
	}

	// Recommend next logical topics
	nextTopics := ale.getNextLogicalTopics()
	for _, topic := range nextTopics {
		recommendations = append(recommendations, LearningRecommendation{
			Type:          RecommendationNext,
			Topic:         topic,
			Reason:        fmt.Sprintf("Ready to explore %s based on current competencies", topic),
			Priority:      PriorityMedium,
			Difficulty:    ale.getRecommendedDifficulty(topic),
			EstimatedTime: 40 * time.Minute,
		})
	}

	// Sort recommendations by priority
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[Priority]int{
			PriorityCritical: 4,
			PriorityHigh:     3,
			PriorityMedium:   2,
			PriorityLow:      1,
		}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	// Limit to top 10 recommendations
	if len(recommendations) > 10 {
		recommendations = recommendations[:10]
	}

	return recommendations
}

// LearningRecommendation represents a personalized learning suggestion
type LearningRecommendation struct {
	Type          RecommendationType `json:"type"`
	Topic         string             `json:"topic"`
	Reason        string             `json:"reason"`
	Priority      Priority           `json:"priority"`
	Difficulty    DifficultyLevel    `json:"difficulty"`
	EstimatedTime time.Duration      `json:"estimated_time"`
}

type RecommendationType string

const (
	RecommendationWeakArea RecommendationType = "weak_area"
	RecommendationGoal     RecommendationType = "goal"
	RecommendationNext     RecommendationType = "next_topic"
	RecommendationReview   RecommendationType = "review"
)

func (ale *AdaptiveLearningEngine) getRecommendedDifficulty(topic string) DifficultyLevel {
	competency := ale.getTopicCompetency(topic)
	return ale.difficultyFromCompetency(competency)
}

func (ale *AdaptiveLearningEngine) difficultyFromCompetency(competency CompetencyLevel) DifficultyLevel {
	mapping := map[CompetencyLevel]DifficultyLevel{
		CompetencyNone:         DifficultyVeryEasy,
		CompetencyBeginner:     DifficultyEasy,
		CompetencyBasic:        DifficultyMedium,
		CompetencyIntermediate: DifficultyMedium,
		CompetencyAdvanced:     DifficultyHard,
		CompetencyExpert:       DifficultyVeryHard,
		CompetencyMastery:      DifficultyVeryHard,
	}

	if difficulty, exists := mapping[competency]; exists {
		return difficulty
	}
	return DifficultyMedium
}

func (ale *AdaptiveLearningEngine) getNextLogicalTopics() []string {
	// Analyze competency map to suggest logical next topics
	readyTopics := []string{}

	topicDependencies := map[string][]string{
		"nixos-basics":             {"nix-basics"},
		"configuration-management": {"nixos-basics"},
		"service-management":       {"configuration-management"},
		"flakes":                   {"nix-basics", "nixos-basics"},
		"modules":                  {"configuration-management"},
		"overlays":                 {"nix-basics", "package-management"},
		"home-manager":             {"nixos-basics", "configuration-management"},
		"networking":               {"nixos-basics", "service-management"},
		"security":                 {"nixos-basics", "configuration-management"},
		"development-environments": {"nix-basics", "package-management"},
	}

	for topic, dependencies := range topicDependencies {
		// Check if user has not learned this topic yet
		if ale.getTopicCompetency(topic) <= CompetencyBeginner {
			// Check if all dependencies are met
			allDepsMet := true
			for _, dep := range dependencies {
				if ale.getTopicCompetency(dep) < CompetencyBasic {
					allDepsMet = false
					break
				}
			}

			if allDepsMet {
				readyTopics = append(readyTopics, topic)
			}
		}
	}

	return readyTopics
}
