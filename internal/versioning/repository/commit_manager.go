package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// CommitManager handles configuration commit operations
type CommitManager struct {
	repo   *ConfigRepository
	logger *logger.Logger
}

// NewCommitManager creates a new commit manager
func NewCommitManager(repo *ConfigRepository, logger *logger.Logger) *CommitManager {
	return &CommitManager{
		repo:   repo,
		logger: logger,
	}
}

// CommitOptions contains options for creating commits
type CommitOptions struct {
	Message     string            `json:"message"`
	Author      string            `json:"author"`
	Email       string            `json:"email"`
	Files       map[string]string `json:"files"`
	Metadata    map[string]string `json:"metadata"`
	Tags        []string          `json:"tags"`
	Environment string            `json:"environment"`
	ChangeType  ChangeType        `json:"change_type"`
}

// ChangeType represents the type of configuration change
type ChangeType string

const (
	ChangeTypeFeature     ChangeType = "feature"
	ChangeTypeBugfix      ChangeType = "bugfix"
	ChangeTypeUpdate      ChangeType = "update"
	ChangeTypeRollback    ChangeType = "rollback"
	ChangeTypeMigration   ChangeType = "migration"
	ChangeTypeHotfix      ChangeType = "hotfix"
	ChangeTypeMaintenance ChangeType = "maintenance"
)

// CommitWithOptions creates a commit with detailed options
func (cm *CommitManager) CommitWithOptions(ctx context.Context, options *CommitOptions) (*ConfigSnapshot, error) {
	// Validate options
	if options.Message == "" {
		return nil, fmt.Errorf("commit message is required")
	}

	if len(options.Files) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}

	// Add commit metadata
	metadata := make(map[string]string)
	for k, v := range options.Metadata {
		metadata[k] = v
	}

	metadata["change_type"] = string(options.ChangeType)
	metadata["environment"] = options.Environment
	metadata["commit_time"] = time.Now().Format(time.RFC3339)

	// Add author information if provided
	if options.Author != "" {
		metadata["author"] = options.Author
	}
	if options.Email != "" {
		metadata["email"] = options.Email
	}

	// Create the commit
	snapshot, err := cm.repo.Commit(ctx, options.Message, options.Files, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	// Add tags if specified
	for _, tag := range options.Tags {
		if err := cm.repo.TagSnapshot(ctx, snapshot.ID, tag); err != nil {
			cm.logger.Warn(fmt.Sprintf("Failed to add tag %s: %v", tag, err))
		}
	}

	cm.logger.Info(fmt.Sprintf("Created commit %s (%s): %s",
		snapshot.ID[:8], options.ChangeType, options.Message))

	return snapshot, nil
}

// CreateFeatureCommit creates a commit for a new feature
func (cm *CommitManager) CreateFeatureCommit(ctx context.Context, message string, files map[string]string, featureDetails map[string]string) (*ConfigSnapshot, error) {
	options := &CommitOptions{
		Message:    fmt.Sprintf("feat: %s", message),
		Files:      files,
		Metadata:   featureDetails,
		ChangeType: ChangeTypeFeature,
	}

	return cm.CommitWithOptions(ctx, options)
}

// CreateBugfixCommit creates a commit for a bug fix
func (cm *CommitManager) CreateBugfixCommit(ctx context.Context, message string, files map[string]string, bugDetails map[string]string) (*ConfigSnapshot, error) {
	options := &CommitOptions{
		Message:    fmt.Sprintf("fix: %s", message),
		Files:      files,
		Metadata:   bugDetails,
		ChangeType: ChangeTypeBugfix,
	}

	return cm.CommitWithOptions(ctx, options)
}

// CreateHotfixCommit creates a commit for a critical hotfix
func (cm *CommitManager) CreateHotfixCommit(ctx context.Context, message string, files map[string]string, urgencyLevel string) (*ConfigSnapshot, error) {
	metadata := map[string]string{
		"urgency_level": urgencyLevel,
		"hotfix_time":   time.Now().Format(time.RFC3339),
	}

	options := &CommitOptions{
		Message:    fmt.Sprintf("hotfix: %s", message),
		Files:      files,
		Metadata:   metadata,
		ChangeType: ChangeTypeHotfix,
		Tags:       []string{fmt.Sprintf("hotfix-%d", time.Now().Unix())},
	}

	return cm.CommitWithOptions(ctx, options)
}

// CommitDiff represents the difference between two commits
type CommitDiff struct {
	FromCommit   *ConfigSnapshot      `json:"from_commit"`
	ToCommit     *ConfigSnapshot      `json:"to_commit"`
	FilesAdded   []string             `json:"files_added"`
	FilesRemoved []string             `json:"files_removed"`
	FilesChanged []string             `json:"files_changed"`
	Changes      map[string]*FileDiff `json:"changes"`
	Summary      *DiffSummary         `json:"summary"`
}

// FileDiff represents changes to a single file
type FileDiff struct {
	Filename     string  `json:"filename"`
	ChangeType   string  `json:"change_type"` // added, removed, modified
	LinesAdded   int     `json:"lines_added"`
	LinesRemoved int     `json:"lines_removed"`
	OldContent   string  `json:"old_content,omitempty"`
	NewContent   string  `json:"new_content,omitempty"`
	Hunks        []*Hunk `json:"hunks,omitempty"`
}

// Hunk represents a contiguous block of changes
type Hunk struct {
	OldStart int      `json:"old_start"`
	OldLines int      `json:"old_lines"`
	NewStart int      `json:"new_start"`
	NewLines int      `json:"new_lines"`
	Lines    []string `json:"lines"`
}

// DiffSummary provides a summary of changes between commits
type DiffSummary struct {
	TotalFiles    int `json:"total_files"`
	FilesAdded    int `json:"files_added"`
	FilesRemoved  int `json:"files_removed"`
	FilesModified int `json:"files_modified"`
	TotalLines    int `json:"total_lines"`
	LinesAdded    int `json:"lines_added"`
	LinesRemoved  int `json:"lines_removed"`
}

// GetCommitDiff returns the difference between two commits
func (cm *CommitManager) GetCommitDiff(ctx context.Context, fromCommitID, toCommitID string) (*CommitDiff, error) {
	fromCommit, err := cm.repo.GetSnapshot(ctx, fromCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from commit: %w", err)
	}

	toCommit, err := cm.repo.GetSnapshot(ctx, toCommitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to commit: %w", err)
	}

	diff := &CommitDiff{
		FromCommit: fromCommit,
		ToCommit:   toCommit,
		Changes:    make(map[string]*FileDiff),
	}

	// Find files that exist in both commits, only in from, or only in to
	allFiles := make(map[string]bool)
	for filename := range fromCommit.Files {
		allFiles[filename] = true
	}
	for filename := range toCommit.Files {
		allFiles[filename] = true
	}

	for filename := range allFiles {
		oldContent, existsInFrom := fromCommit.Files[filename]
		newContent, existsInTo := toCommit.Files[filename]

		fileDiff := &FileDiff{
			Filename: filename,
		}

		if !existsInFrom && existsInTo {
			// File was added
			fileDiff.ChangeType = "added"
			fileDiff.NewContent = newContent
			fileDiff.LinesAdded = countLines(newContent)
			diff.FilesAdded = append(diff.FilesAdded, filename)
		} else if existsInFrom && !existsInTo {
			// File was removed
			fileDiff.ChangeType = "removed"
			fileDiff.OldContent = oldContent
			fileDiff.LinesRemoved = countLines(oldContent)
			diff.FilesRemoved = append(diff.FilesRemoved, filename)
		} else if oldContent != newContent {
			// File was modified
			fileDiff.ChangeType = "modified"
			fileDiff.OldContent = oldContent
			fileDiff.NewContent = newContent
			fileDiff.LinesAdded, fileDiff.LinesRemoved = calculateLineDiff(oldContent, newContent)
			diff.FilesChanged = append(diff.FilesChanged, filename)
		} else {
			// File unchanged, skip
			continue
		}

		diff.Changes[filename] = fileDiff
	}

	// Calculate summary
	diff.Summary = &DiffSummary{
		TotalFiles:    len(diff.Changes),
		FilesAdded:    len(diff.FilesAdded),
		FilesRemoved:  len(diff.FilesRemoved),
		FilesModified: len(diff.FilesChanged),
	}

	for _, fileDiff := range diff.Changes {
		diff.Summary.LinesAdded += fileDiff.LinesAdded
		diff.Summary.LinesRemoved += fileDiff.LinesRemoved
	}

	diff.Summary.TotalLines = diff.Summary.LinesAdded + diff.Summary.LinesRemoved

	return diff, nil
}

// GetCommitHistory returns paginated commit history
func (cm *CommitManager) GetCommitHistory(ctx context.Context, limit, offset int, filter *HistoryFilter) (*CommitHistory, error) {
	snapshots, err := cm.repo.ListSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	// Apply filters
	filtered := cm.applyHistoryFilter(snapshots, filter)

	// Sort by timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply pagination
	total := len(filtered)
	start := offset
	end := offset + limit

	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}

	var commits []*ConfigSnapshot
	if start < end {
		commits = filtered[start:end]
	}

	return &CommitHistory{
		Commits:       commits,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
		HasMore:       end < total,
		FilterApplied: filter != nil,
	}, nil
}

// HistoryFilter contains filter options for commit history
type HistoryFilter struct {
	Author      string     `json:"author,omitempty"`
	ChangeType  ChangeType `json:"change_type,omitempty"`
	Environment string     `json:"environment,omitempty"`
	FromDate    *time.Time `json:"from_date,omitempty"`
	ToDate      *time.Time `json:"to_date,omitempty"`
	Message     string     `json:"message,omitempty"` // Substring search in commit message
}

// CommitHistory represents paginated commit history
type CommitHistory struct {
	Commits       []*ConfigSnapshot `json:"commits"`
	Total         int               `json:"total"`
	Limit         int               `json:"limit"`
	Offset        int               `json:"offset"`
	HasMore       bool              `json:"has_more"`
	FilterApplied bool              `json:"filter_applied"`
}

// applyHistoryFilter applies filters to the commit list
func (cm *CommitManager) applyHistoryFilter(snapshots []*ConfigSnapshot, filter *HistoryFilter) []*ConfigSnapshot {
	if filter == nil {
		return snapshots
	}

	var filtered []*ConfigSnapshot
	for _, snapshot := range snapshots {
		if cm.matchesFilter(snapshot, filter) {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered
}

// matchesFilter checks if a snapshot matches the given filter
func (cm *CommitManager) matchesFilter(snapshot *ConfigSnapshot, filter *HistoryFilter) bool {
	if filter.Author != "" && snapshot.Author != filter.Author {
		return false
	}

	if filter.ChangeType != "" {
		changeType, exists := snapshot.Metadata["change_type"]
		if !exists || ChangeType(changeType) != filter.ChangeType {
			return false
		}
	}

	if filter.Environment != "" {
		environment, exists := snapshot.Metadata["environment"]
		if !exists || environment != filter.Environment {
			return false
		}
	}

	if filter.FromDate != nil && snapshot.Timestamp.Before(*filter.FromDate) {
		return false
	}

	if filter.ToDate != nil && snapshot.Timestamp.After(*filter.ToDate) {
		return false
	}

	if filter.Message != "" {
		if !containsIgnoreCase(snapshot.Message, filter.Message) {
			return false
		}
	}

	return true
}

// GetCommitStats returns statistics about commits
func (cm *CommitManager) GetCommitStats(ctx context.Context, timeRange *TimeRange) (*CommitStats, error) {
	snapshots, err := cm.repo.ListSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	stats := &CommitStats{
		Total:         len(snapshots),
		ByChangeType:  make(map[ChangeType]int),
		ByEnvironment: make(map[string]int),
		ByAuthor:      make(map[string]int),
		ByDay:         make(map[string]int),
		TimeRange:     timeRange,
	}

	for _, snapshot := range snapshots {
		// Apply time range filter if specified
		if timeRange != nil {
			if timeRange.From != nil && snapshot.Timestamp.Before(*timeRange.From) {
				continue
			}
			if timeRange.To != nil && snapshot.Timestamp.After(*timeRange.To) {
				continue
			}
		}

		// Count by change type
		if changeType, exists := snapshot.Metadata["change_type"]; exists {
			stats.ByChangeType[ChangeType(changeType)]++
		}

		// Count by environment
		if environment, exists := snapshot.Metadata["environment"]; exists {
			stats.ByEnvironment[environment]++
		}

		// Count by author
		stats.ByAuthor[snapshot.Author]++

		// Count by day
		day := snapshot.Timestamp.Format("2006-01-02")
		stats.ByDay[day]++
	}

	return stats, nil
}

// TimeRange represents a time range for filtering
type TimeRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// CommitStats contains statistics about commits
type CommitStats struct {
	Total         int                `json:"total"`
	ByChangeType  map[ChangeType]int `json:"by_change_type"`
	ByEnvironment map[string]int     `json:"by_environment"`
	ByAuthor      map[string]int     `json:"by_author"`
	ByDay         map[string]int     `json:"by_day"`
	TimeRange     *TimeRange         `json:"time_range,omitempty"`
}

// Helper functions

func countLines(content string) int {
	if content == "" {
		return 0
	}
	lines := 1
	for _, char := range content {
		if char == '\n' {
			lines++
		}
	}
	return lines
}

func calculateLineDiff(oldContent, newContent string) (added, removed int) {
	// Simple line counting diff - in a real implementation, you'd use a proper diff algorithm
	oldLines := countLines(oldContent)
	newLines := countLines(newContent)

	if newLines > oldLines {
		added = newLines - oldLines
	} else if oldLines > newLines {
		removed = oldLines - newLines
	}

	return added, removed
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
