package errors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ErrorAnalytics tracks and analyzes error patterns
type ErrorAnalytics struct {
	mu           sync.RWMutex
	errorCounts  map[ErrorCode]int
	errorHistory []ErrorEvent
	patterns     map[string]int
	sessionStart time.Time
	dataDir      string
}

// ErrorEvent represents a recorded error event
type ErrorEvent struct {
	Code       ErrorCode              `json:"code"`
	Category   ErrorCategory          `json:"category"`
	Severity   ErrorSeverity          `json:"severity"`
	Message    string                 `json:"message"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	SessionID  string                 `json:"session_id"`
	UserAction string                 `json:"user_action,omitempty"`
	Resolution string                 `json:"resolution,omitempty"`
}

// ErrorPattern represents a detected error pattern
type ErrorPattern struct {
	Pattern     string        `json:"pattern"`
	Count       int           `json:"count"`
	FirstSeen   time.Time     `json:"first_seen"`
	LastSeen    time.Time     `json:"last_seen"`
	Category    ErrorCategory `json:"category"`
	Severity    ErrorSeverity `json:"severity"`
	Suggestions []string      `json:"suggestions"`
}

// AnalyticsReport provides insights into error patterns
type AnalyticsReport struct {
	SessionDuration    time.Duration         `json:"session_duration"`
	TotalErrors        int                   `json:"total_errors"`
	UniqueErrors       int                   `json:"unique_errors"`
	ErrorsByCategory   map[ErrorCategory]int `json:"errors_by_category"`
	ErrorsBySeverity   map[ErrorSeverity]int `json:"errors_by_severity"`
	TopErrors          []ErrorFrequency      `json:"top_errors"`
	DetectedPatterns   []ErrorPattern        `json:"detected_patterns"`
	RecommendedActions []string              `json:"recommended_actions"`
	Timestamp          time.Time             `json:"timestamp"`
}

// ErrorFrequency represents error frequency data
type ErrorFrequency struct {
	Code      ErrorCode `json:"code"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// NewErrorAnalytics creates a new error analytics instance
func NewErrorAnalytics(dataDir string) *ErrorAnalytics {
	analytics := &ErrorAnalytics{
		errorCounts:  make(map[ErrorCode]int),
		errorHistory: make([]ErrorEvent, 0),
		patterns:     make(map[string]int),
		sessionStart: time.Now(),
		dataDir:      dataDir,
	}

	// Load historical data if available
	analytics.loadHistoricalData()

	return analytics
}

// RecordError records an error for analysis
func (ea *ErrorAnalytics) RecordError(err error, userAction string) {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	var event ErrorEvent

	if nixaiErr, ok := err.(*NixAIError); ok {
		event = ErrorEvent{
			Code:       nixaiErr.Code,
			Category:   nixaiErr.Category,
			Severity:   nixaiErr.Severity,
			Message:    nixaiErr.Message,
			Context:    nixaiErr.Context,
			Timestamp:  time.Now(),
			SessionID:  ea.generateSessionID(),
			UserAction: userAction,
		}

		ea.errorCounts[nixaiErr.Code]++
	} else {
		// For generic errors, try to categorize
		code, category, severity := ea.categorizeGenericError(err)
		event = ErrorEvent{
			Code:       code,
			Category:   category,
			Severity:   severity,
			Message:    err.Error(),
			Timestamp:  time.Now(),
			SessionID:  ea.generateSessionID(),
			UserAction: userAction,
		}

		ea.errorCounts[code]++
	}

	ea.errorHistory = append(ea.errorHistory, event)

	// Detect patterns
	pattern := ea.extractPattern(event)
	ea.patterns[pattern]++

	// Limit history size to prevent memory issues
	if len(ea.errorHistory) > 1000 {
		ea.errorHistory = ea.errorHistory[100:] // Keep last 900 entries
	}

	// Persist data periodically
	if len(ea.errorHistory)%10 == 0 {
		go ea.persistData()
	}
}

// RecordResolution records how an error was resolved
func (ea *ErrorAnalytics) RecordResolution(errorCode ErrorCode, resolution string) {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	// Find the most recent error with this code and update its resolution
	for i := len(ea.errorHistory) - 1; i >= 0; i-- {
		if ea.errorHistory[i].Code == errorCode && ea.errorHistory[i].Resolution == "" {
			ea.errorHistory[i].Resolution = resolution
			break
		}
	}
}

// GenerateReport generates an analytics report
func (ea *ErrorAnalytics) GenerateReport() *AnalyticsReport {
	ea.mu.RLock()
	defer ea.mu.RUnlock()

	report := &AnalyticsReport{
		SessionDuration:    time.Since(ea.sessionStart),
		TotalErrors:        len(ea.errorHistory),
		UniqueErrors:       len(ea.errorCounts),
		ErrorsByCategory:   make(map[ErrorCategory]int),
		ErrorsBySeverity:   make(map[ErrorSeverity]int),
		TopErrors:          ea.getTopErrors(10),
		DetectedPatterns:   ea.getDetectedPatterns(),
		RecommendedActions: ea.generateRecommendations(),
		Timestamp:          time.Now(),
	}

	// Calculate category and severity distributions
	for _, event := range ea.errorHistory {
		report.ErrorsByCategory[event.Category]++
		report.ErrorsBySeverity[event.Severity]++
	}

	return report
}

// GetTopErrors returns the most frequent errors
func (ea *ErrorAnalytics) GetTopErrors(limit int) []ErrorFrequency {
	ea.mu.RLock()
	defer ea.mu.RUnlock()

	return ea.getTopErrors(limit)
}

// getTopErrors internal method to get top errors
func (ea *ErrorAnalytics) getTopErrors(limit int) []ErrorFrequency {
	frequencies := make([]ErrorFrequency, 0, len(ea.errorCounts))

	// Calculate first and last seen times for each error
	firstSeen := make(map[ErrorCode]time.Time)
	lastSeen := make(map[ErrorCode]time.Time)

	for _, event := range ea.errorHistory {
		if first, exists := firstSeen[event.Code]; !exists || event.Timestamp.Before(first) {
			firstSeen[event.Code] = event.Timestamp
		}
		if last, exists := lastSeen[event.Code]; !exists || event.Timestamp.After(last) {
			lastSeen[event.Code] = event.Timestamp
		}
	}

	for code, count := range ea.errorCounts {
		frequencies = append(frequencies, ErrorFrequency{
			Code:      code,
			Count:     count,
			FirstSeen: firstSeen[code],
			LastSeen:  lastSeen[code],
		})
	}

	// Sort by count (descending)
	sort.Slice(frequencies, func(i, j int) bool {
		return frequencies[i].Count > frequencies[j].Count
	})

	if limit > 0 && len(frequencies) > limit {
		frequencies = frequencies[:limit]
	}

	return frequencies
}

// getDetectedPatterns returns detected error patterns
func (ea *ErrorAnalytics) getDetectedPatterns() []ErrorPattern {
	patterns := make([]ErrorPattern, 0, len(ea.patterns))

	for pattern, count := range ea.patterns {
		if count >= 2 { // Only patterns that occurred multiple times
			// Find first and last occurrence
			var firstSeen, lastSeen time.Time
			var category ErrorCategory
			var severity ErrorSeverity

			for _, event := range ea.errorHistory {
				eventPattern := ea.extractPattern(event)
				if eventPattern == pattern {
					if firstSeen.IsZero() || event.Timestamp.Before(firstSeen) {
						firstSeen = event.Timestamp
						category = event.Category
						severity = event.Severity
					}
					if lastSeen.IsZero() || event.Timestamp.After(lastSeen) {
						lastSeen = event.Timestamp
					}
				}
			}

			patterns = append(patterns, ErrorPattern{
				Pattern:     pattern,
				Count:       count,
				FirstSeen:   firstSeen,
				LastSeen:    lastSeen,
				Category:    category,
				Severity:    severity,
				Suggestions: ea.getSuggestionsForPattern(pattern),
			})
		}
	}

	// Sort by count (descending)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}

// generateRecommendations generates recommendations based on error patterns
func (ea *ErrorAnalytics) generateRecommendations() []string {
	recommendations := make([]string, 0)

	// Analyze error categories
	categoryCount := make(map[ErrorCategory]int)
	for _, event := range ea.errorHistory {
		categoryCount[event.Category]++
	}

	// Generate category-specific recommendations
	if categoryCount[CategoryNetwork] > 3 {
		recommendations = append(recommendations,
			"Frequent network errors detected. Check your internet connection and firewall settings.")
	}

	if categoryCount[CategoryAI] > 3 {
		recommendations = append(recommendations,
			"Multiple AI provider errors. Consider switching to Ollama for local processing.")
	}

	if categoryCount[CategoryNixOS] > 3 {
		recommendations = append(recommendations,
			"Repeated NixOS errors. Consider running 'nix-collect-garbage' and updating channels.")
	}

	if categoryCount[CategoryMCP] > 2 {
		recommendations = append(recommendations,
			"MCP server issues detected. Try restarting the MCP server: 'nixai mcp-server restart'")
	}

	// Check for error frequency
	totalErrors := len(ea.errorHistory)
	if totalErrors > 10 {
		recommendations = append(recommendations,
			"High error rate detected. Consider reviewing configuration and system health.")
	}

	// Check for unresolved errors
	unresolvedCount := 0
	for _, event := range ea.errorHistory {
		if event.Resolution == "" {
			unresolvedCount++
		}
	}

	if unresolvedCount > 5 {
		recommendations = append(recommendations,
			"Many unresolved errors. Consider running 'nixai doctor' for system diagnostics.")
	}

	return recommendations
}

// extractPattern extracts a pattern from an error event
func (ea *ErrorAnalytics) extractPattern(event ErrorEvent) string {
	// Simple pattern based on code and category
	return fmt.Sprintf("%s:%s", event.Category, event.Code)
}

// getSuggestionsForPattern returns suggestions for a specific pattern
func (ea *ErrorAnalytics) getSuggestionsForPattern(pattern string) []string {
	// Basic pattern-based suggestions
	switch {
	case contains(string(pattern), string(CategoryNetwork)):
		return []string{
			"Check network connectivity",
			"Verify firewall settings",
			"Try again later",
		}
	case contains(string(pattern), string(CategoryAI)):
		return []string{
			"Check AI provider configuration",
			"Verify API keys",
			"Consider using Ollama for local processing",
		}
	case contains(string(pattern), string(CategoryNixOS)):
		return []string{
			"Run nix-collect-garbage",
			"Update channels",
			"Check configuration syntax",
		}
	default:
		return []string{
			"Check logs for more details",
			"Try the operation again",
		}
	}
}

// categorizeGenericError attempts to categorize a generic error
func (ea *ErrorAnalytics) categorizeGenericError(err error) (ErrorCode, ErrorCategory, ErrorSeverity) {
	errorMsg := strings.ToLower(err.Error())

	// Network errors
	if contains(errorMsg, "timeout") {
		return ErrorCodeNetworkTimeout, CategoryNetwork, SeverityMedium
	}
	if contains(errorMsg, "connection refused") {
		return ErrorCodeConnectionRefused, CategoryNetwork, SeverityMedium
	}

	// File system errors
	if contains(errorMsg, "permission denied") {
		return ErrorCodePermissionDenied, CategoryFileSystem, SeverityMedium
	}
	if contains(errorMsg, "no such file") || contains(errorMsg, "file not found") {
		return ErrorCodeFileNotFound, CategoryFileSystem, SeverityMedium
	}

	// NixOS errors
	if contains(errorMsg, "build failed") || contains(errorMsg, "nix") {
		return ErrorCodeNixBuildFailed, CategoryNixOS, SeverityHigh
	}

	// Default to unknown
	return ErrorCodeUnknown, CategoryInternal, SeverityLow
}

// generateSessionID generates a session identifier
func (ea *ErrorAnalytics) generateSessionID() string {
	return fmt.Sprintf("session_%d", ea.sessionStart.Unix())
}

// persistData saves analytics data to disk
func (ea *ErrorAnalytics) persistData() {
	if ea.dataDir == "" {
		return
	}

	// Ensure directory exists
	if err := os.MkdirAll(ea.dataDir, 0755); err != nil {
		return
	}

	// Save error history
	historyFile := filepath.Join(ea.dataDir, "error_history.json")
	ea.saveErrorHistory(historyFile)

	// Save analytics report
	reportFile := filepath.Join(ea.dataDir, "analytics_report.json")
	ea.saveAnalyticsReport(reportFile)
}

// saveErrorHistory saves error history to file
func (ea *ErrorAnalytics) saveErrorHistory(filename string) {
	data, err := json.MarshalIndent(ea.errorHistory, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(filename, data, 0644)
}

// saveAnalyticsReport saves analytics report to file
func (ea *ErrorAnalytics) saveAnalyticsReport(filename string) {
	report := ea.GenerateReport()
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(filename, data, 0644)
}

// loadHistoricalData loads historical analytics data
func (ea *ErrorAnalytics) loadHistoricalData() {
	if ea.dataDir == "" {
		return
	}

	historyFile := filepath.Join(ea.dataDir, "error_history.json")
	if data, err := os.ReadFile(historyFile); err == nil {
		var history []ErrorEvent
		if json.Unmarshal(data, &history) == nil {
			// Only load recent history (last 7 days)
			cutoff := time.Now().AddDate(0, 0, -7)
			for _, event := range history {
				if event.Timestamp.After(cutoff) {
					ea.errorHistory = append(ea.errorHistory, event)
					ea.errorCounts[event.Code]++
					pattern := ea.extractPattern(event)
					ea.patterns[pattern]++
				}
			}
		}
	}
}

// ExportReport exports analytics report to a file
func (ea *ErrorAnalytics) ExportReport(filename string) error {
	report := ea.GenerateReport()
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetErrorRate returns the error rate per hour
func (ea *ErrorAnalytics) GetErrorRate() float64 {
	ea.mu.RLock()
	defer ea.mu.RUnlock()

	if ea.sessionStart.IsZero() {
		return 0
	}

	hours := time.Since(ea.sessionStart).Hours()
	if hours == 0 {
		return 0
	}

	return float64(len(ea.errorHistory)) / hours
}

// GetMostProblematicCategory returns the category with the most errors
func (ea *ErrorAnalytics) GetMostProblematicCategory() ErrorCategory {
	ea.mu.RLock()
	defer ea.mu.RUnlock()

	categoryCount := make(map[ErrorCategory]int)
	for _, event := range ea.errorHistory {
		categoryCount[event.Category]++
	}

	var maxCategory ErrorCategory
	var maxCount int

	for category, count := range categoryCount {
		if count > maxCount {
			maxCount = count
			maxCategory = category
		}
	}

	return maxCategory
}

// ClearData clears all analytics data
func (ea *ErrorAnalytics) ClearData() error {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	// Clear all in-memory data
	ea.errorHistory = nil
	ea.errorCounts = make(map[ErrorCode]int)
	ea.patterns = make(map[string]int)
	ea.sessionStart = time.Now()

	// Clear persistent data files
	errorHistoryFile := filepath.Join(ea.dataDir, "error_history.json")
	analyticsFile := filepath.Join(ea.dataDir, "analytics_report.json")

	// Remove files if they exist
	if err := os.Remove(errorHistoryFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove error history file: %w", err)
	}

	if err := os.Remove(analyticsFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove analytics report file: %w", err)
	}

	return nil
}
