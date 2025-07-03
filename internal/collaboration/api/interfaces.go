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
	
	// Model Updates (Federated Learning)
	SubmitModelUpdate(ctx context.Context, update *ModelUpdate) error
	GetModelUpdates(ctx context.Context, modelID string) ([]*ModelUpdate, error)
	ApplyModelUpdate(ctx context.Context, update *ModelUpdate) error
}

// FederatedLearningAPI handles distributed model training
type FederatedLearningAPI interface {
	// Model coordination
	JoinTraining(ctx context.Context, modelID string) error
	LeaveTraining(ctx context.Context, modelID string) error
	GetTrainingStatus(ctx context.Context, modelID string) (*TrainingStatus, error)
	
	// Gradient sharing
	ShareGradients(ctx context.Context, gradients *GradientUpdate) error
	AggregateGradients(ctx context.Context, modelID string) (*AggregatedGradients, error)
	
	// Privacy-preserving training
	AnonymizeGradients(ctx context.Context, gradients *GradientUpdate) (*GradientUpdate, error)
	ValidatePrivacy(ctx context.Context, data interface{}) (*PrivacyValidation, error)
}

// CommunityAPI handles community-driven features
type CommunityAPI interface {
	// Community solutions
	SubmitSolution(ctx context.Context, solution *Solution) error
	VoteSolution(ctx context.Context, solutionID string, vote *Vote) error
	CommentSolution(ctx context.Context, solutionID string, comment *Comment) error
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

// Data Types

// Knowledge sharing types
type KnowledgeItem struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Author      string                 `json:"author"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type KnowledgeQuery struct {
	Query       string            `json:"query"`
	Category    string            `json:"category,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Limit       int               `json:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
}

type KnowledgeResults struct {
	Items      []KnowledgeItem `json:"items"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	HasMore    bool            `json:"has_more"`
}

// Solution types
type Solution struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Code        string                 `json:"code"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Author      string                 `json:"author"`
	Created     time.Time              `json:"created"`
	Rating      float64                `json:"rating"`
	Votes       int                    `json:"votes"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ProblemDescription struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Context     map[string]string `json:"context"`
	Tags        []string          `json:"tags"`
}

type SolutionResults struct {
	Solutions []Solution `json:"solutions"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	HasMore   bool       `json:"has_more"`
}

type Rating struct {
	Value   int    `json:"value"`   // 1-5 stars
	Comment string `json:"comment,omitempty"`
	UserID  string `json:"user_id"`
}

// Pattern types
type ConfigurationPattern struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Pattern     string                 `json:"pattern"`
	Category    string                 `json:"category"`
	UseCase     string                 `json:"use_case"`
	Frequency   int                    `json:"frequency"`
	Success     float64                `json:"success_rate"`
	Created     time.Time              `json:"created"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type PatternContext struct {
	Category     string            `json:"category"`
	Environment  string            `json:"environment"`
	Requirements []string          `json:"requirements"`
	Constraints  map[string]string `json:"constraints"`
}

type PatternResults struct {
	Patterns []ConfigurationPattern `json:"patterns"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	HasMore  bool                   `json:"has_more"`
}

type PatternFeedback struct {
	PatternID   string                 `json:"pattern_id"`
	Successful  bool                   `json:"successful"`
	Comments    string                 `json:"comments"`
	Improvements []string              `json:"improvements"`
	Context     map[string]interface{} `json:"context"`
}

// Privacy types
type PrivacyPolicy struct {
	Version     string            `json:"version"`
	LastUpdated time.Time         `json:"last_updated"`
	Policies    map[string]string `json:"policies"`
}

type PrivacyPreferences struct {
	ShareData        bool     `json:"share_data"`
	ShareConfigs     bool     `json:"share_configs"`
	SharePatterns    bool     `json:"share_patterns"`
	AnonymizeLevel   string   `json:"anonymize_level"` // "none", "basic", "full"
	AllowedCategories []string `json:"allowed_categories"`
}

// Federated learning types
type ModelUpdate struct {
	ModelID     string                 `json:"model_id"`
	Version     string                 `json:"version"`
	UpdateType  string                 `json:"update_type"` // "weights", "gradients", "parameters"
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Created     time.Time              `json:"created"`
}

type SolutionUpdate struct {
	SolutionID  string                 `json:"solution_id"`
	UpdateType  string                 `json:"update_type"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Author      string                 `json:"author"`
	Created     time.Time              `json:"created"`
}

type SolutionMatches struct {
	Exact   []Solution `json:"exact"`
	Similar []Solution `json:"similar"`
	Related []Solution `json:"related"`
}

type RankingCriteria struct {
	SortBy    string  `json:"sort_by"`    // "rating", "date", "relevance"
	Order     string  `json:"order"`      // "asc", "desc"
	MinRating float64 `json:"min_rating"`
	MaxAge    int     `json:"max_age_days"`
}

// Additional missing types
type RankedSolution struct {
	Solution *Solution `json:"solution"`
	Score    float64   `json:"score"`
	Reason   string    `json:"reason"`
}

type SolutionValidation struct {
	Valid   bool     `json:"valid"`
	Issues  []string `json:"issues"`
	Warnings []string `json:"warnings"`
}

type TestEnvironment struct {
	OS           string            `json:"os"`
	Version      string            `json:"version"`
	Architecture string            `json:"architecture"`
	Resources    map[string]string `json:"resources"`
}

type TestResult struct {
	Success     bool                   `json:"success"`
	Output      string                 `json:"output"`
	Errors      []string               `json:"errors"`
	Duration    time.Duration          `json:"duration"`
	Environment *TestEnvironment       `json:"environment"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type EffectivenessMetrics struct {
	SuccessRate  float64 `json:"success_rate"`
	UsageCount   int     `json:"usage_count"`
	AverageRating float64 `json:"average_rating"`
	LastUsed     time.Time `json:"last_used"`
}

type UsageMetrics struct {
	TotalUses    int       `json:"total_uses"`
	UniqueUsers  int       `json:"unique_users"`
	LastUsed     time.Time `json:"last_used"`
	PopularityTrend string `json:"popularity_trend"`
}

type TrainingStatus struct {
	ModelID       string    `json:"model_id"`
	Status        string    `json:"status"`
	Participants  int       `json:"participants"`
	Progress      float64   `json:"progress"`
	LastUpdate    time.Time `json:"last_update"`
}

type GradientUpdate struct {
	ModelID     string                 `json:"model_id"`
	Gradients   []byte                 `json:"gradients"`
	Epoch       int                    `json:"epoch"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type AggregatedGradients struct {
	ModelID       string    `json:"model_id"`
	Gradients     []byte    `json:"gradients"`
	Participants  int       `json:"participants"`
	Epoch         int       `json:"epoch"`
	Timestamp     time.Time `json:"timestamp"`
}

type PrivacyValidation struct {
	Valid         bool     `json:"valid"`
	PrivacyLevel  string   `json:"privacy_level"`
	Issues        []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

type Vote struct {
	UserID    string    `json:"user_id"`
	Value     int       `json:"value"` // -1, 0, 1
	Timestamp time.Time `json:"timestamp"`
}

type Comment struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Votes     int       `json:"votes"`
}