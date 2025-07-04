// Package cache provides intelligent caching and behavior analysis for nixai
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// BehaviorAnalyzer analyzes user interaction patterns to optimize caching
type BehaviorAnalyzer struct {
	patterns       map[string]*UserPattern
	queryHistory   []QueryEvent
	sessionData    map[string]*SessionData
	config         *config.UserConfig
	logger         *logger.Logger
	mu             sync.RWMutex
	maxHistory     int
	analysisWindow time.Duration
}

// UserPattern represents a detected user behavior pattern
type UserPattern struct {
	ID              string                 `json:"id"`
	Type            PatternType            `json:"type"`
	Frequency       int                    `json:"frequency"`
	Confidence      float64                `json:"confidence"`      // 0.0 - 1.0
	LastSeen        time.Time              `json:"last_seen"`
	Triggers        []string               `json:"triggers"`        // What triggers this pattern
	ExpectedQueries []string               `json:"expected_queries"` // Likely follow-up queries
	Context         map[string]interface{} `json:"context"`
	Metadata        PatternMetadata        `json:"metadata"`
}

// PatternType represents different types of behavior patterns
type PatternType string

const (
	PatternSequential    PatternType = "sequential"     // Sequential command patterns
	PatternTimeBasedRecurring PatternType = "time_recurring" // Time-based recurring patterns
	PatternContextual    PatternType = "contextual"     // Context-driven patterns
	PatternProjectSpecific PatternType = "project_specific" // Project-specific workflows
	PatternLearning      PatternType = "learning"       // Learning/exploration patterns
	PatternTroubleshooting PatternType = "troubleshooting" // Troubleshooting workflows
	PatternRoutine       PatternType = "routine"        // Daily/weekly routines
)

// PatternMetadata contains additional pattern information
type PatternMetadata struct {
	TimeOfDay     []int     `json:"time_of_day"`     // Hours when pattern is active (0-23)
	DayOfWeek     []int     `json:"day_of_week"`     // Days when pattern is active (0-6)
	Duration      time.Duration `json:"duration"`    // Typical duration of pattern
	Complexity    string    `json:"complexity"`      // "simple", "moderate", "complex"
	Domain        string    `json:"domain"`          // "nixos", "development", "system_admin"
	ProjectPath   string    `json:"project_path"`    // Associated project directory
	SuccessRate   float64   `json:"success_rate"`    // Pattern completion success rate
}

// QueryEvent represents a user query/interaction
type QueryEvent struct {
	ID          string                 `json:"id"`
	Query       string                 `json:"query"`
	QueryType   string                 `json:"query_type"`   // "command", "config", "troubleshoot", etc.
	Timestamp   time.Time              `json:"timestamp"`
	Response    string                 `json:"response"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	Context     QueryContext           `json:"context"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// QueryContext provides context about the query
type QueryContext struct {
	WorkingDirectory string            `json:"working_directory"`
	ProjectType     string            `json:"project_type"`     // "nix-flake", "nixos-config", "home-manager", etc.
	PreviousQueries []string          `json:"previous_queries"` // Last 5 queries
	TimeOfDay       int               `json:"time_of_day"`      // Hour 0-23
	DayOfWeek       int               `json:"day_of_week"`      // 0-6 (Sunday=0)
	SessionLength   time.Duration     `json:"session_length"`
	Environment     map[string]string `json:"environment"`      // Relevant env vars
}

// SessionData tracks data for a user session
type SessionData struct {
	ID          string        `json:"id"`
	StartTime   time.Time     `json:"start_time"`
	LastActivity time.Time    `json:"last_activity"`
	Queries     []QueryEvent  `json:"queries"`
	Patterns    []string      `json:"patterns"`    // Pattern IDs detected in this session
	Context     QueryContext  `json:"context"`
	Active      bool          `json:"active"`
}

// BehaviorInsight represents insights derived from behavior analysis
type BehaviorInsight struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Data        map[string]interface{} `json:"data"`
	Suggestions []string               `json:"suggestions"`
}

// PredictionResult represents a prediction about user behavior
type PredictionResult struct {
	NextLikelyQueries []PredictedQuery `json:"next_likely_queries"`
	Confidence        float64          `json:"confidence"`
	Timeframe        time.Duration    `json:"timeframe"`
	Context          string           `json:"context"`
	BasedOnPattern   string           `json:"based_on_pattern"`
}

// PredictedQuery represents a predicted user query
type PredictedQuery struct {
	Query      string    `json:"query"`
	Probability float64  `json:"probability"`
	Type       string    `json:"type"`
	Context    string    `json:"context"`
	PreGenerate bool     `json:"pre_generate"` // Should we pre-generate response?
}

// NewBehaviorAnalyzer creates a new behavior analyzer
func NewBehaviorAnalyzer(cfg *config.UserConfig) *BehaviorAnalyzer {
	return &BehaviorAnalyzer{
		patterns:       make(map[string]*UserPattern),
		queryHistory:   make([]QueryEvent, 0),
		sessionData:    make(map[string]*SessionData),
		config:         cfg,
		logger:         logger.NewLogger(),
		maxHistory:     10000, // Keep last 10k queries
		analysisWindow: 24 * time.Hour * 30, // Analyze last 30 days
	}
}

// RecordQuery records a user query for behavior analysis
func (ba *BehaviorAnalyzer) RecordQuery(ctx context.Context, event QueryEvent) error {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	// Add to history
	ba.queryHistory = append(ba.queryHistory, event)

	// Trim history if too large
	if len(ba.queryHistory) > ba.maxHistory {
		ba.queryHistory = ba.queryHistory[len(ba.queryHistory)-ba.maxHistory:]
	}

	// Update session data
	sessionID := ba.getOrCreateSession(event)
	if session, exists := ba.sessionData[sessionID]; exists {
		session.Queries = append(session.Queries, event)
		session.LastActivity = event.Timestamp
	}

	ba.logger.Info(fmt.Sprintf("Recorded query: %s (type: %s, duration: %v)", 
		event.Query, event.QueryType, event.Duration))

	return nil
}

// AnalyzePatterns analyzes recorded queries to identify behavior patterns
func (ba *BehaviorAnalyzer) AnalyzePatterns(ctx context.Context) ([]BehaviorInsight, error) {
	ba.mu.Lock()
	defer ba.mu.Unlock()

	ba.logger.Info("Starting behavior pattern analysis")

	var insights []BehaviorInsight

	// Analyze sequential patterns
	sequentialInsights := ba.analyzeSequentialPatterns()
	insights = append(insights, sequentialInsights...)

	// Analyze time-based patterns
	timeBasedInsights := ba.analyzeTimeBasedPatterns()
	insights = append(insights, timeBasedInsights...)

	// Analyze contextual patterns
	contextualInsights := ba.analyzeContextualPatterns()
	insights = append(insights, contextualInsights...)

	// Analyze project-specific patterns
	projectInsights := ba.analyzeProjectPatterns()
	insights = append(insights, projectInsights...)

	// Update pattern cache
	ba.updatePatternCache(insights)

	ba.logger.Info(fmt.Sprintf("Identified %d behavior insights", len(insights)))

	return insights, nil
}

// PredictNextQueries predicts what the user is likely to query next
func (ba *BehaviorAnalyzer) PredictNextQueries(ctx context.Context, currentContext QueryContext) (*PredictionResult, error) {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	prediction := &PredictionResult{
		NextLikelyQueries: make([]PredictedQuery, 0),
		Timeframe:        5 * time.Minute, // Default prediction window
	}

	// Find matching patterns
	matchingPatterns := ba.findMatchingPatterns(currentContext)
	if len(matchingPatterns) == 0 {
		return prediction, nil
	}

	// Use the highest confidence pattern
	bestPattern := matchingPatterns[0]
	prediction.BasedOnPattern = bestPattern.ID
	prediction.Confidence = bestPattern.Confidence

	// Generate predictions based on pattern
	for _, expectedQuery := range bestPattern.ExpectedQueries {
		probability := ba.calculateQueryProbability(expectedQuery, bestPattern, currentContext)
		
		predicted := PredictedQuery{
			Query:       expectedQuery,
			Probability: probability,
			Type:        ba.inferQueryType(expectedQuery),
			Context:     bestPattern.ID,
			PreGenerate: probability > 0.7, // Pre-generate high probability queries
		}
		
		prediction.NextLikelyQueries = append(prediction.NextLikelyQueries, predicted)
	}

	// Sort by probability
	sort.Slice(prediction.NextLikelyQueries, func(i, j int) bool {
		return prediction.NextLikelyQueries[i].Probability > prediction.NextLikelyQueries[j].Probability
	})

	ba.logger.Info(fmt.Sprintf("Predicted %d likely queries based on pattern %s", 
		len(prediction.NextLikelyQueries), bestPattern.ID))

	return prediction, nil
}

// GetUserInsights returns insights about user behavior
func (ba *BehaviorAnalyzer) GetUserInsights() []BehaviorInsight {
	ba.mu.RLock()
	defer ba.mu.RUnlock()

	var insights []BehaviorInsight

	// Usage frequency insights
	if len(ba.queryHistory) > 0 {
		insights = append(insights, BehaviorInsight{
			Type:        "usage_frequency",
			Description: fmt.Sprintf("Average of %.1f queries per day", ba.calculateDailyAverage()),
			Confidence:  0.9,
			Data: map[string]interface{}{
				"total_queries": len(ba.queryHistory),
				"active_days":   ba.calculateActiveDays(),
			},
		})
	}

	// Most common query types
	commonTypes := ba.calculateCommonQueryTypes()
	if len(commonTypes) > 0 {
		insights = append(insights, BehaviorInsight{
			Type:        "query_preferences",
			Description: fmt.Sprintf("Most common query type: %s", commonTypes[0]),
			Confidence:  0.8,
			Data: map[string]interface{}{
				"query_types": commonTypes,
			},
		})
	}

	// Time-based patterns
	if timePattern := ba.detectTimePattern(); timePattern != "" {
		insights = append(insights, BehaviorInsight{
			Type:        "time_pattern",
			Description: timePattern,
			Confidence:  0.75,
		})
	}

	return insights
}

// Helper methods

func (ba *BehaviorAnalyzer) getOrCreateSession(event QueryEvent) string {
	// Simple session logic: group queries within 30 minutes
	sessionID := fmt.Sprintf("session_%d", event.Timestamp.Unix()/(30*60))
	
	if _, exists := ba.sessionData[sessionID]; !exists {
		ba.sessionData[sessionID] = &SessionData{
			ID:          sessionID,
			StartTime:   event.Timestamp,
			LastActivity: event.Timestamp,
			Queries:     make([]QueryEvent, 0),
			Patterns:    make([]string, 0),
			Context:     event.Context,
			Active:      true,
		}
	}
	
	return sessionID
}

func (ba *BehaviorAnalyzer) analyzeSequentialPatterns() []BehaviorInsight {
	var insights []BehaviorInsight
	
	// Look for common command sequences (length 2-5)
	sequences := make(map[string]int)
	
	for i := 0; i < len(ba.queryHistory)-1; i++ {
		for length := 2; length <= 5 && i+length <= len(ba.queryHistory); length++ {
			var sequence []string
			for j := 0; j < length; j++ {
				sequence = append(sequence, ba.queryHistory[i+j].QueryType)
			}
			key := strings.Join(sequence, "->")
			sequences[key]++
		}
	}
	
	// Find sequences that occur at least 3 times
	for seq, count := range sequences {
		if count >= 3 {
			pattern := &UserPattern{
				ID:              fmt.Sprintf("seq_%s", strings.ReplaceAll(seq, "->", "_")),
				Type:            PatternSequential,
				Frequency:       count,
				Confidence:      float64(count) / float64(len(ba.queryHistory)) * 2, // Boost for sequences
				LastSeen:        time.Now(),
				Triggers:        strings.Split(seq, "->")[:1], // First item is trigger
				ExpectedQueries: strings.Split(seq, "->")[1:], // Rest are expected
				Context:         map[string]interface{}{"sequence": seq},
				Metadata: PatternMetadata{
					Complexity:  "moderate",
					Domain:      "general",
					SuccessRate: 0.8, // Default
				},
			}
			
			ba.patterns[pattern.ID] = pattern
			
			insights = append(insights, BehaviorInsight{
				Type:        "sequential_pattern",
				Description: fmt.Sprintf("Detected sequence pattern: %s (occurs %d times)", seq, count),
				Confidence:  pattern.Confidence,
				Data:        map[string]interface{}{"pattern_id": pattern.ID, "sequence": seq},
			})
		}
	}
	
	return insights
}

func (ba *BehaviorAnalyzer) analyzeTimeBasedPatterns() []BehaviorInsight {
	var insights []BehaviorInsight
	
	// Analyze hour-of-day patterns
	hourCounts := make(map[int]int)
	for _, event := range ba.queryHistory {
		hour := event.Timestamp.Hour()
		hourCounts[hour]++
	}
	
	// Find peak hours (more than average + 1 std dev)
	total := len(ba.queryHistory)
	average := float64(total) / 24.0
	
	var peakHours []int
	for hour, count := range hourCounts {
		if float64(count) > average*1.5 { // Simple threshold
			peakHours = append(peakHours, hour)
		}
	}
	
	if len(peakHours) > 0 {
		pattern := &UserPattern{
			ID:         "time_based_usage",
			Type:       PatternTimeBasedRecurring,
			Frequency:  len(peakHours),
			Confidence: 0.7,
			LastSeen:   time.Now(),
			Context:    map[string]interface{}{"peak_hours": peakHours},
			Metadata: PatternMetadata{
				TimeOfDay:   peakHours,
				Complexity:  "simple",
				Domain:      "general",
				SuccessRate: 0.9,
			},
		}
		
		ba.patterns[pattern.ID] = pattern
		
		insights = append(insights, BehaviorInsight{
			Type:        "time_based_pattern",
			Description: fmt.Sprintf("Most active during hours: %v", peakHours),
			Confidence:  0.7,
			Data:        map[string]interface{}{"peak_hours": peakHours},
		})
	}
	
	return insights
}

func (ba *BehaviorAnalyzer) analyzeContextualPatterns() []BehaviorInsight {
	var insights []BehaviorInsight
	
	// Group queries by working directory
	dirPatterns := make(map[string][]QueryEvent)
	for _, event := range ba.queryHistory {
		dir := event.Context.WorkingDirectory
		if dir != "" {
			dirPatterns[dir] = append(dirPatterns[dir], event)
		}
	}
	
	// Find directories with significant activity
	for dir, events := range dirPatterns {
		if len(events) >= 5 { // At least 5 queries in this directory
			queryTypes := make(map[string]int)
			for _, event := range events {
				queryTypes[event.QueryType]++
			}
			
			// Find dominant query type
			var dominantType string
			maxCount := 0
			for qType, count := range queryTypes {
				if count > maxCount {
					maxCount = count
					dominantType = qType
				}
			}
			
			pattern := &UserPattern{
				ID:         fmt.Sprintf("context_%s", strings.ReplaceAll(dir, "/", "_")),
				Type:       PatternContextual,
				Frequency:  len(events),
				Confidence: float64(maxCount) / float64(len(events)),
				LastSeen:   events[len(events)-1].Timestamp,
				Context:    map[string]interface{}{"directory": dir, "dominant_type": dominantType},
				Metadata: PatternMetadata{
					ProjectPath: dir,
					Complexity:  "moderate",
					Domain:      dominantType,
					SuccessRate: 0.85,
				},
			}
			
			ba.patterns[pattern.ID] = pattern
			
			insights = append(insights, BehaviorInsight{
				Type:        "contextual_pattern",
				Description: fmt.Sprintf("Directory %s associated with %s queries", dir, dominantType),
				Confidence:  pattern.Confidence,
				Data:        map[string]interface{}{"directory": dir, "query_type": dominantType},
			})
		}
	}
	
	return insights
}

func (ba *BehaviorAnalyzer) analyzeProjectPatterns() []BehaviorInsight {
	var insights []BehaviorInsight
	
	// Analyze project types
	projectTypes := make(map[string][]QueryEvent)
	for _, event := range ba.queryHistory {
		pType := event.Context.ProjectType
		if pType != "" {
			projectTypes[pType] = append(projectTypes[pType], event)
		}
	}
	
	for projectType, events := range projectTypes {
		if len(events) >= 3 {
			// Calculate success rate
			successCount := 0
			for _, event := range events {
				if event.Success {
					successCount++
				}
			}
			successRate := float64(successCount) / float64(len(events))
			
			pattern := &UserPattern{
				ID:         fmt.Sprintf("project_%s", projectType),
				Type:       PatternProjectSpecific,
				Frequency:  len(events),
				Confidence: 0.8,
				LastSeen:   time.Now(),
				Context:    map[string]interface{}{"project_type": projectType},
				Metadata: PatternMetadata{
					Complexity:  "moderate",
					Domain:      projectType,
					SuccessRate: successRate,
				},
			}
			
			ba.patterns[pattern.ID] = pattern
			
			insights = append(insights, BehaviorInsight{
				Type:        "project_pattern",
				Description: fmt.Sprintf("Working with %s projects (success rate: %.1f%%)", projectType, successRate*100),
				Confidence:  0.8,
				Data:        map[string]interface{}{"project_type": projectType, "success_rate": successRate},
			})
		}
	}
	
	return insights
}

func (ba *BehaviorAnalyzer) updatePatternCache(insights []BehaviorInsight) {
	// Update pattern metadata based on insights
	for _, insight := range insights {
		if patternID, ok := insight.Data["pattern_id"].(string); ok {
			if pattern, exists := ba.patterns[patternID]; exists {
				pattern.LastSeen = time.Now()
				pattern.Confidence = insight.Confidence
			}
		}
	}
}

func (ba *BehaviorAnalyzer) findMatchingPatterns(context QueryContext) []*UserPattern {
	var matching []*UserPattern
	
	for _, pattern := range ba.patterns {
		score := ba.calculatePatternMatchScore(pattern, context)
		if score > 0.5 { // Threshold for pattern matching
			patternCopy := *pattern
			patternCopy.Confidence = score
			matching = append(matching, &patternCopy)
		}
	}
	
	// Sort by confidence (descending)
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].Confidence > matching[j].Confidence
	})
	
	return matching
}

func (ba *BehaviorAnalyzer) calculatePatternMatchScore(pattern *UserPattern, context QueryContext) float64 {
	score := 0.0
	
	// Time-based matching
	if len(pattern.Metadata.TimeOfDay) > 0 {
		for _, hour := range pattern.Metadata.TimeOfDay {
			if hour == context.TimeOfDay {
				score += 0.3
				break
			}
		}
	}
	
	// Project-based matching
	if pattern.Metadata.ProjectPath != "" && pattern.Metadata.ProjectPath == context.WorkingDirectory {
		score += 0.4
	}
	
	// Context-based matching
	if projectType, ok := pattern.Context["project_type"].(string); ok {
		if projectType == context.ProjectType {
			score += 0.3
		}
	}
	
	// Base confidence
	score += pattern.Confidence * 0.3
	
	// Ensure score doesn't exceed 1.0
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

func (ba *BehaviorAnalyzer) calculateQueryProbability(query string, pattern *UserPattern, context QueryContext) float64 {
	// Base probability from pattern confidence
	probability := pattern.Confidence * 0.7
	
	// Adjust based on recency of pattern
	timeSinceLastSeen := time.Since(pattern.LastSeen)
	if timeSinceLastSeen < time.Hour {
		probability += 0.2
	} else if timeSinceLastSeen < 24*time.Hour {
		probability += 0.1
	}
	
	// Adjust based on pattern frequency
	if pattern.Frequency > 5 {
		probability += 0.1
	}
	
	// Ensure probability is between 0 and 1
	if probability > 1.0 {
		probability = 1.0
	}
	if probability < 0.0 {
		probability = 0.0
	}
	
	return probability
}

func (ba *BehaviorAnalyzer) inferQueryType(query string) string {
	query = strings.ToLower(query)
	
	if strings.Contains(query, "build") || strings.Contains(query, "compile") {
		return "build"
	}
	if strings.Contains(query, "install") || strings.Contains(query, "add") {
		return "package_management"
	}
	if strings.Contains(query, "config") || strings.Contains(query, "configure") {
		return "configuration"
	}
	if strings.Contains(query, "error") || strings.Contains(query, "fix") || strings.Contains(query, "debug") {
		return "troubleshooting"
	}
	if strings.Contains(query, "service") || strings.Contains(query, "systemd") {
		return "service_management"
	}
	
	return "general"
}

func (ba *BehaviorAnalyzer) calculateDailyAverage() float64 {
	if len(ba.queryHistory) == 0 {
		return 0
	}
	
	firstQuery := ba.queryHistory[0].Timestamp
	lastQuery := ba.queryHistory[len(ba.queryHistory)-1].Timestamp
	days := lastQuery.Sub(firstQuery).Hours() / 24
	
	if days < 1 {
		days = 1
	}
	
	return float64(len(ba.queryHistory)) / days
}

func (ba *BehaviorAnalyzer) calculateActiveDays() int {
	daySet := make(map[string]bool)
	
	for _, event := range ba.queryHistory {
		day := event.Timestamp.Format("2006-01-02")
		daySet[day] = true
	}
	
	return len(daySet)
}

func (ba *BehaviorAnalyzer) calculateCommonQueryTypes() []string {
	typeCounts := make(map[string]int)
	
	for _, event := range ba.queryHistory {
		typeCounts[event.QueryType]++
	}
	
	type typeCount struct {
		Type  string
		Count int
	}
	
	var sorted []typeCount
	for t, c := range typeCounts {
		sorted = append(sorted, typeCount{Type: t, Count: c})
	}
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})
	
	var result []string
	for _, tc := range sorted {
		result = append(result, tc.Type)
	}
	
	return result
}

func (ba *BehaviorAnalyzer) detectTimePattern() string {
	if len(ba.queryHistory) < 10 {
		return ""
	}
	
	// Simple heuristic for time patterns
	hourCounts := make(map[int]int)
	for _, event := range ba.queryHistory {
		hour := event.Timestamp.Hour()
		hourCounts[hour]++
	}
	
	// Find peak usage time
	maxCount := 0
	peakHour := -1
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}
	
	if maxCount > len(ba.queryHistory)/6 { // More than ~17% of queries in one hour
		if peakHour >= 9 && peakHour <= 17 {
			return "Most active during business hours"
		} else if peakHour >= 18 && peakHour <= 22 {
			return "Most active during evening hours"
		} else {
			return "Most active during off-hours"
		}
	}
	
	return ""
}

// GetPatterns returns all detected patterns
func (ba *BehaviorAnalyzer) GetPatterns() map[string]*UserPattern {
	ba.mu.RLock()
	defer ba.mu.RUnlock()
	
	// Return a copy to prevent external modification
	patterns := make(map[string]*UserPattern)
	for k, v := range ba.patterns {
		patternCopy := *v
		patterns[k] = &patternCopy
	}
	
	return patterns
}

// SavePatterns saves patterns to persistent storage
func (ba *BehaviorAnalyzer) SavePatterns(filePath string) error {
	ba.mu.RLock()
	defer ba.mu.RUnlock()
	
	data := map[string]interface{}{
		"patterns":      ba.patterns,
		"query_history": ba.queryHistory[max(0, len(ba.queryHistory)-1000):], // Save last 1000 queries
		"sessions":      ba.sessionData,
		"timestamp":     time.Now(),
	}
	
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patterns: %w", err)
	}
	
	// In a real implementation, this would write to file
	ba.logger.Info(fmt.Sprintf("Would save %d patterns to %s", len(ba.patterns), filePath))
	_ = jsonData // Suppress unused variable warning
	
	return nil
}

// LoadPatterns loads patterns from persistent storage
func (ba *BehaviorAnalyzer) LoadPatterns(filePath string) error {
	ba.mu.Lock()
	defer ba.mu.Unlock()
	
	// In a real implementation, this would read from file
	ba.logger.Info(fmt.Sprintf("Would load patterns from %s", filePath))
	
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}