// Package api provides interfaces for collaborative intelligence features
package api

import (
	"context"
	"time"
)

// CollaborativeAPI defines the main interface for collaborative intelligence
type CollaborativeAPI interface {
	// Knowledge Sharing
	ShareKnowledge(ctx context.Context, knowledge *KnowledgeItem) error
	SearchKnowledge(ctx context.Context, query *KnowledgeQuery) (*KnowledgeResults, error)
	GetKnowledgeItem(ctx context.Context, id string) (*KnowledgeItem, error)
	
	// Community Solutions
	SubmitSolution(ctx context.Context, solution *Solution) error
	FindSolutions(ctx context.Context, problem *ProblemDescription) (*SolutionResults, error)
	RateSolution(ctx context.Context, solutionID string, rating *Rating) error
	
	// Pattern Learning
	SubmitPattern(ctx context.Context, pattern *ConfigurationPattern) error
	DiscoverPatterns(ctx context.Context, context *PatternContext) (*PatternResults, error)
	LearnFromPattern(ctx context.Context, patternID string, feedback *PatternFeedback) error
	
	// Privacy & Anonymization
	AnonymizeData(ctx context.Context, data interface{}) (interface{}, error)
	GetPrivacyPolicy(ctx context.Context) (*PrivacyPolicy, error)
	SetPrivacyPreferences(ctx context.Context, prefs *PrivacyPreferences) error
	
	// Federated Learning
	ContributeToModel(ctx context.Context, contribution *ModelContribution) error
	GetModelUpdates(ctx context.Context, modelID string) (*ModelUpdate, error)
	
	// Community Stats
	GetCommunityStats(ctx context.Context) (*CommunityStats, error)
	GetUserStats(ctx context.Context, userID string) (*UserStats, error)
}

// KnowledgeManager handles knowledge sharing and management
type KnowledgeManager interface {
	// Knowledge lifecycle
	CreateKnowledge(ctx context.Context, knowledge *KnowledgeItem) error
	UpdateKnowledge(ctx context.Context, id string, updates *KnowledgeUpdate) error
	DeleteKnowledge(ctx context.Context, id string) error
	
	// Knowledge discovery
	SearchByTags(ctx context.Context, tags []string) (*KnowledgeResults, error)
	SearchByCategory(ctx context.Context, category string) (*KnowledgeResults, error)
	SearchBySimilarity(ctx context.Context, reference *KnowledgeItem) (*KnowledgeResults, error)
	
	// Knowledge validation
	ValidateKnowledge(ctx context.Context, knowledge *KnowledgeItem) (*ValidationResult, error)
	ModerateKnowledge(ctx context.Context, id string, moderation *ModerationAction) error
	
	// Knowledge analytics
	GetKnowledgeMetrics(ctx context.Context, id string) (*KnowledgeMetrics, error)
	GetTrendingKnowledge(ctx context.Context, timeframe string) (*KnowledgeResults, error)
}

// SolutionEngine handles community solution management
type SolutionEngine interface {
	// Solution lifecycle
	CreateSolution(ctx context.Context, solution *Solution) error
	UpdateSolution(ctx context.Context, id string, updates *SolutionUpdate) error
	ArchiveSolution(ctx context.Context, id string, reason string) error
	
	// Solution matching
	MatchSolutions(ctx context.Context, problem *ProblemDescription) (*SolutionMatches, error)
	RankSolutions(ctx context.Context, solutions []*Solution, criteria *RankingCriteria) ([]*RankedSolution, error)
	
	// Solution validation
	ValidateSolution(ctx context.Context, solution *Solution) (*SolutionValidation, error)
	TestSolution(ctx context.Context, solution *Solution, environment *TestEnvironment) (*TestResult, error)
	
	// Solution analytics
	GetSolutionEffectiveness(ctx context.Context, id string) (*EffectivenessMetrics, error)
	GetSolutionUsage(ctx context.Context, id string) (*UsageMetrics, error)
}

// PatternAnalyzer handles configuration pattern analysis and learning
type PatternAnalyzer interface {
	// Pattern discovery
	DiscoverPatterns(ctx context.Context, configurations []*Configuration) ([]*ConfigurationPattern, error)
	AnalyzePattern(ctx context.Context, pattern *ConfigurationPattern) (*PatternAnalysis, error)
	ComparePatterns(ctx context.Context, pattern1, pattern2 *ConfigurationPattern) (*PatternComparison, error)
	
	// Pattern learning
	LearnFromSuccess(ctx context.Context, successCase *SuccessCase) error
	LearnFromFailure(ctx context.Context, failureCase *FailureCase) error
	UpdatePatternKnowledge(ctx context.Context, knowledge *PatternKnowledge) error
	
	// Pattern recommendation
	RecommendPatterns(ctx context.Context, context *RecommendationContext) (*PatternRecommendations, error)
	GetPatternSuggestions(ctx context.Context, configuration *Configuration) (*PatternSuggestions, error)
	
	// Pattern evolution
	TrackPatternEvolution(ctx context.Context, patternID string) (*PatternEvolution, error)
	PredictPatternTrends(ctx context.Context, timeframe string) (*PatternTrends, error)
}

// FederatedLearningClient handles federated learning operations
type FederatedLearningClient interface {
	// Model participation
	JoinFederatedModel(ctx context.Context, modelID string, capabilities *ClientCapabilities) error
	LeaveFederatedModel(ctx context.Context, modelID string) error
	
	// Learning contributions
	SubmitGradients(ctx context.Context, modelID string, gradients *ModelGradients) error
	ReceiveModelUpdate(ctx context.Context, modelID string) (*ModelUpdate, error)
	
	// Privacy-preserving operations
	ComputePrivateGradients(ctx context.Context, data *TrainingData, privacy *PrivacyParams) (*PrivateGradients, error)
	ValidateModelUpdate(ctx context.Context, update *ModelUpdate) (*ValidationResult, error)
	
	// Client status
	GetLearningStatus(ctx context.Context) (*LearningStatus, error)
	GetContributionHistory(ctx context.Context) (*ContributionHistory, error)
}

// PrivacyManager handles privacy and anonymization
type PrivacyManager interface {
	// Data anonymization
	AnonymizeConfiguration(ctx context.Context, config *Configuration) (*AnonymizedConfiguration, error)
	AnonymizeError(ctx context.Context, errorData *ErrorData) (*AnonymizedError, error)
	AnonymizeUsageData(ctx context.Context, usage *UsageData) (*AnonymizedUsage, error)
	
	// Privacy assessment
	AssessPrivacyRisk(ctx context.Context, data interface{}) (*PrivacyRisk, error)
	CheckDataSensitivity(ctx context.Context, data interface{}) (*SensitivityReport, error)
	
	// Consent management
	GetConsentStatus(ctx context.Context, userID string) (*ConsentStatus, error)
	UpdateConsent(ctx context.Context, userID string, consent *ConsentUpdate) error
	
	// Privacy compliance
	GeneratePrivacyReport(ctx context.Context, userID string) (*PrivacyReport, error)
	HandleDataDeletionRequest(ctx context.Context, userID string) error
}

// Data Types for Collaborative Intelligence

// KnowledgeItem represents a piece of shared knowledge
type KnowledgeItem struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Tags          []string               `json:"tags"`
	Content       interface{}            `json:"content"`
	Author        string                 `json:"author"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Version       int                    `json:"version"`
	Rating        float64                `json:"rating"`
	UsageCount    int                    `json:"usage_count"`
	Verified      bool                   `json:"verified"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// KnowledgeQuery represents a search query for knowledge
type KnowledgeQuery struct {
	Query       string            `json:"query"`
	Category    string            `json:"category,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Author      string            `json:"author,omitempty"`
	Limit       int               `json:"limit"`
	Offset      int               `json:"offset"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	SortBy      string            `json:"sort_by"`
	SortOrder   string            `json:"sort_order"`
}

// KnowledgeResults represents search results
type KnowledgeResults struct {
	Items      []*KnowledgeItem `json:"items"`
	Total      int              `json:"total"`
	Query      *KnowledgeQuery  `json:"query"`
	Timestamp  time.Time        `json:"timestamp"`
	Suggestions []string        `json:"suggestions,omitempty"`
}

// Solution represents a community solution
type Solution struct {
	ID            string                 `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Problem       *ProblemDescription    `json:"problem"`
	Steps         []*SolutionStep        `json:"steps"`
	Code          string                 `json:"code,omitempty"`
	Author        string                 `json:"author"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Rating        float64                `json:"rating"`
	Difficulty    string                 `json:"difficulty"`
	Verified      bool                   `json:"verified"`
	SuccessRate   float64                `json:"success_rate"`
	Tags          []string               `json:"tags"`
	Prerequisites []string               `json:"prerequisites"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ProblemDescription describes a problem to be solved
type ProblemDescription struct {
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	Symptoms      []string               `json:"symptoms"`
	Environment   *EnvironmentContext    `json:"environment"`
	ErrorMessages []string               `json:"error_messages,omitempty"`
	Context       map[string]interface{} `json:"context"`
	Urgency       string                 `json:"urgency"`
}

// ConfigurationPattern represents a reusable configuration pattern
type ConfigurationPattern struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Pattern      interface{}            `json:"pattern"`
	Variables    []*PatternVariable     `json:"variables"`
	Conditions   []*PatternCondition    `json:"conditions"`
	Benefits     []string               `json:"benefits"`
	Drawbacks    []string               `json:"drawbacks,omitempty"`
	Popularity   float64                `json:"popularity"`
	Reliability  float64                `json:"reliability"`
	Complexity   string                 `json:"complexity"`
	Author       string                 `json:"author"`
	CreatedAt    time.Time              `json:"created_at"`
	UsageCount   int                    `json:"usage_count"`
	SuccessRate  float64                `json:"success_rate"`
	Tags         []string               `json:"tags"`
	Examples     []*PatternExample      `json:"examples"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ModelContribution represents a contribution to federated learning
type ModelContribution struct {
	ID           string                 `json:"id"`
	ModelID      string                 `json:"model_id"`
	ClientID     string                 `json:"client_id"`
	Gradients    interface{}            `json:"gradients"`
	DataSize     int                    `json:"data_size"`
	Privacy      *PrivacyParams         `json:"privacy"`
	Timestamp    time.Time              `json:"timestamp"`
	Metrics      map[string]float64     `json:"metrics"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// PrivacyPreferences represents user privacy preferences
type PrivacyPreferences struct {
	UserID                  string                 `json:"user_id"`
	AllowDataSharing        bool                   `json:"allow_data_sharing"`
	AllowAnalytics          bool                   `json:"allow_analytics"`
	AnonymizationLevel      string                 `json:"anonymization_level"`
	DataRetentionPeriod     int                    `json:"data_retention_period"`
	AllowFederatedLearning  bool                   `json:"allow_federated_learning"`
	ShareErrorReports       bool                   `json:"share_error_reports"`
	ShareUsagePatterns      bool                   `json:"share_usage_patterns"`
	ShareConfigurations     bool                   `json:"share_configurations"`
	CustomPreferences       map[string]interface{} `json:"custom_preferences"`
	UpdatedAt               time.Time              `json:"updated_at"`
}

// Supporting Types

type KnowledgeUpdate struct {
	Title       *string                `json:"title,omitempty"`
	Description *string                `json:"description,omitempty"`
	Content     interface{}            `json:"content,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Score    float64  `json:"score"`
	Issues   []string `json:"issues"`
	Warnings []string `json:"warnings"`
}

type ModerationAction struct {
	Action    string `json:"action"`
	Reason    string `json:"reason"`
	Moderator string `json:"moderator"`
}

type KnowledgeMetrics struct {
	Views       int     `json:"views"`
	Downloads   int     `json:"downloads"`
	Shares      int     `json:"shares"`
	Rating      float64 `json:"rating"`
	Feedback    int     `json:"feedback_count"`
}

type SolutionStep struct {
	Order       int    `json:"order"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	Expected    string `json:"expected,omitempty"`
}

type EnvironmentContext struct {
	NixOSVersion  string                 `json:"nixos_version"`
	Architecture  string                 `json:"architecture"`
	Hardware      map[string]interface{} `json:"hardware"`
	Services      []string               `json:"services"`
	Packages      []string               `json:"packages"`
}

type PatternVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Default      interface{} `json:"default,omitempty"`
	Required     bool        `json:"required"`
	Validation   string      `json:"validation,omitempty"`
}

type PatternCondition struct {
	Type        string      `json:"type"`
	Expression  string      `json:"expression"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
}

type PatternExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Input       interface{}            `json:"input"`
	Output      interface{}            `json:"output"`
	Context     map[string]interface{} `json:"context"`
}

type PrivacyParams struct {
	Method          string  `json:"method"`
	Epsilon         float64 `json:"epsilon,omitempty"`
	Delta           float64 `json:"delta,omitempty"`
	NoiseScale      float64 `json:"noise_scale,omitempty"`
	ClippingBound   float64 `json:"clipping_bound,omitempty"`
}

type CommunityStats struct {
	TotalUsers          int                    `json:"total_users"`
	ActiveUsers         int                    `json:"active_users"`
	TotalKnowledge      int                    `json:"total_knowledge"`
	TotalSolutions      int                    `json:"total_solutions"`
	TotalPatterns       int                    `json:"total_patterns"`
	RecentActivity      []*ActivitySummary     `json:"recent_activity"`
	TopContributors     []*ContributorSummary  `json:"top_contributors"`
	PopularCategories   []*CategorySummary     `json:"popular_categories"`
	Timestamp          time.Time              `json:"timestamp"`
}

type UserStats struct {
	UserID             string    `json:"user_id"`
	JoinedAt           time.Time `json:"joined_at"`
	ContributionsCount int       `json:"contributions_count"`
	KnowledgeShared    int       `json:"knowledge_shared"`
	SolutionsProvided  int       `json:"solutions_provided"`
	PatternsCreated    int       `json:"patterns_created"`
	Rating             float64   `json:"rating"`
	ReputationPoints   int       `json:"reputation_points"`
	Badges             []string  `json:"badges"`
	LastActive         time.Time `json:"last_active"`
}

type ActivitySummary struct {
	Type        string    `json:"type"`
	Count       int       `json:"count"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

type ContributorSummary struct {
	UserID       string  `json:"user_id"`
	Username     string  `json:"username"`
	Contributions int    `json:"contributions"`
	Rating       float64 `json:"rating"`
}

type CategorySummary struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Growth   float64 `json:"growth"`
}

// Event Types for Real-time Collaboration

type CollaborationEvent interface {
	GetType() string
	GetTimestamp() time.Time
	GetUserID() string
}

type KnowledgeSharedEvent struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	UserID      string    `json:"user_id"`
	KnowledgeID string    `json:"knowledge_id"`
	Category    string    `json:"category"`
}

type SolutionFoundEvent struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	UserID      string    `json:"user_id"`
	SolutionID  string    `json:"solution_id"`
	ProblemType string    `json:"problem_type"`
	Success     bool      `json:"success"`
}

type PatternDiscoveredEvent struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	UserID      string    `json:"user_id"`
	PatternID   string    `json:"pattern_id"`
	Confidence  float64   `json:"confidence"`
}

func (e *KnowledgeSharedEvent) GetType() string     { return e.Type }
func (e *KnowledgeSharedEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *KnowledgeSharedEvent) GetUserID() string  { return e.UserID }

func (e *SolutionFoundEvent) GetType() string      { return e.Type }
func (e *SolutionFoundEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *SolutionFoundEvent) GetUserID() string    { return e.UserID }

func (e *PatternDiscoveredEvent) GetType() string  { return e.Type }
func (e *PatternDiscoveredEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *PatternDiscoveredEvent) GetUserID() string { return e.UserID }