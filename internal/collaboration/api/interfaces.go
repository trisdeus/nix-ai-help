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

// ExternalLearningAPI handles learning from external sources like GitHub
type ExternalLearningAPI interface {
	// GitHub code search integration
	SearchGitHubConfigurations(ctx context.Context, query *GitHubSearchQuery) (*GitHubSearchResults, error)
	AnalyzeGitHubRepository(ctx context.Context, repo *GitHubRepository) (*RepositoryAnalysis, error)
	ValidateExternalContent(ctx context.Context, content *ExternalContent) (*ContentValidation, error)
	
	// Quality control
	FilterHighQualityResults(ctx context.Context, results *GitHubSearchResults) (*GitHubSearchResults, error)
	ExtractConfigurationPatterns(ctx context.Context, content *ExternalContent) (*ConfigurationPatterns, error)
	
	// Privacy and anonymization
	AnonymizeGitHubData(ctx context.Context, data *GitHubData) (*AnonymizedData, error)
	ApplyPrivacyFilters(ctx context.Context, content *ExternalContent) (*ExternalContent, error)
}

// GitHub integration types
type GitHubSearchQuery struct {
	Keywords    []string          `json:"keywords"`
	Language    string            `json:"language,omitempty"`
	FileType    string            `json:"file_type,omitempty"`
	Stars       *IntRange         `json:"stars,omitempty"`
	Forks       *IntRange         `json:"forks,omitempty"`
	Updated     *TimeRange        `json:"updated,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	MaxResults  int               `json:"max_results,omitempty"`
	SortBy      string            `json:"sort_by,omitempty"`
}

type IntRange struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

type TimeRange struct {
	After  *time.Time `json:"after,omitempty"`
	Before *time.Time `json:"before,omitempty"`
}

type GitHubSearchResults struct {
	Query       *GitHubSearchQuery  `json:"query"`
	Repositories []GitHubRepository `json:"repositories"`
	TotalCount  int                `json:"total_count"`
	SearchTime  time.Duration      `json:"search_time"`
	QualityScore float64           `json:"quality_score"`
	GeneratedAt time.Time         `json:"generated_at"`
}

type GitHubRepository struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	FullName     string            `json:"full_name"`
	Description  string            `json:"description"`
	URL          string            `json:"url"`
	CloneURL     string            `json:"clone_url"`
	Language     string            `json:"language"`
	Stars        int               `json:"stars"`
	Forks        int               `json:"forks"`
	Topics       []string          `json:"topics"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Owner        GitHubUser        `json:"owner"`
	License      *GitHubLicense    `json:"license,omitempty"`
	Files        []GitHubFile      `json:"files,omitempty"`
	QualityScore float64           `json:"quality_score"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type GitHubUser struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	Type      string `json:"type"`
}

type GitHubLicense struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
}

type GitHubFile struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	Content     string    `json:"content,omitempty"`
	SHA         string    `json:"sha"`
	DownloadURL string    `json:"download_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RepositoryAnalysis struct {
	Repository      *GitHubRepository     `json:"repository"`
	ConfigFiles     []ConfigFileAnalysis  `json:"config_files"`
	Patterns        []ConfigurationPattern `json:"patterns"`
	SecurityIssues  []SecurityIssue       `json:"security_issues"`
	BestPractices   []BestPractice        `json:"best_practices"`
	QualityMetrics  QualityMetrics        `json:"quality_metrics"`
	Recommendations []string              `json:"recommendations"`
	AnalyzedAt      time.Time             `json:"analyzed_at"`
}

type ConfigFileAnalysis struct {
	File            *GitHubFile           `json:"file"`
	FileType        string                `json:"file_type"`
	Complexity      ComplexityMetrics     `json:"complexity"`
	Dependencies    []string              `json:"dependencies"`
	Services        []string              `json:"services"`
	SecurityConfig  []SecurityConfig      `json:"security_config"`
	PerformanceHints []PerformanceHint    `json:"performance_hints"`
	Documentation   DocumentationLevel    `json:"documentation"`
	Maintainability float64               `json:"maintainability"`
}

type ComplexityMetrics struct {
	Lines         int     `json:"lines"`
	Functions     int     `json:"functions"`
	CyclomaticComplexity int `json:"cyclomatic_complexity"`
	NestingDepth  int     `json:"nesting_depth"`
	Score         float64 `json:"score"`
}

type SecurityConfig struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Compliant   bool   `json:"compliant"`
}

type PerformanceHint struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Suggestion  string `json:"suggestion"`
}

type DocumentationLevel struct {
	Score    float64 `json:"score"`
	Comments int     `json:"comments"`
	README   bool    `json:"readme"`
	Examples bool    `json:"examples"`
}

type SecurityIssue struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	File        string    `json:"file"`
	Line        int       `json:"line,omitempty"`
	Suggestion  string    `json:"suggestion"`
	References  []string  `json:"references"`
	DetectedAt  time.Time `json:"detected_at"`
}

type BestPractice struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Applied     bool     `json:"applied"`
	Benefit     string   `json:"benefit"`
	References  []string `json:"references"`
}

type QualityMetrics struct {
	OverallScore    float64 `json:"overall_score"`
	Maintainability float64 `json:"maintainability"`
	Security        float64 `json:"security"`
	Performance     float64 `json:"performance"`
	Documentation   float64 `json:"documentation"`
	TestCoverage    float64 `json:"test_coverage"`
	CodeComplexity  float64 `json:"code_complexity"`
}

type ExternalContent struct {
	Source      string                 `json:"source"`
	Type        string                 `json:"type"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Retrieved   time.Time              `json:"retrieved"`
	Fingerprint string                 `json:"fingerprint"`
}

type ContentValidation struct {
	Valid         bool              `json:"valid"`
	QualityScore  float64           `json:"quality_score"`
	Issues        []ValidationIssue `json:"issues"`
	Suggestions   []string          `json:"suggestions"`
	SafetyLevel   string            `json:"safety_level"`
	ValidatedAt   time.Time         `json:"validated_at"`
}

type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
}

type ConfigurationPatterns struct {
	Patterns    []ConfigurationPattern `json:"patterns"`
	Categories  []string               `json:"categories"`
	Confidence  float64                `json:"confidence"`
	ExtractedAt time.Time              `json:"extracted_at"`
}

type GitHubData struct {
	Repositories []GitHubRepository `json:"repositories"`
	Users        []GitHubUser       `json:"users"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type AnonymizedData struct {
	Data        interface{} `json:"data"`
	Anonymized  []string    `json:"anonymized_fields"`
	Method      string      `json:"anonymization_method"`
	ProcessedAt time.Time   `json:"processed_at"`
}