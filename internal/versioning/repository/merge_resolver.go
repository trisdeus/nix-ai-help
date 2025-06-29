package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// MergeResolver handles merge conflict resolution
type MergeResolver struct {
	repo   *ConfigRepository
	logger *logger.Logger
}

// NewMergeResolver creates a new merge resolver
func NewMergeResolver(repo *ConfigRepository, logger *logger.Logger) *MergeResolver {
	return &MergeResolver{
		repo:   repo,
		logger: logger,
	}
}

// MergeConflict represents a merge conflict
type MergeConflict struct {
	Filename      string              `json:"filename"`
	ConflictType  ConflictType        `json:"conflict_type"`
	BaseContent   string              `json:"base_content"`
	LocalContent  string              `json:"local_content"`
	RemoteContent string              `json:"remote_content"`
	Resolution    *ConflictResolution `json:"resolution,omitempty"`
	Metadata      map[string]string   `json:"metadata"`
}

// ConflictType represents the type of conflict
type ConflictType string

const (
	ConflictTypeContent    ConflictType = "content"
	ConflictTypeAddAdd     ConflictType = "add_add"
	ConflictTypeDeleteEdit ConflictType = "delete_edit"
	ConflictTypeEditDelete ConflictType = "edit_delete"
	ConflictTypeRename     ConflictType = "rename"
)

// ConflictResolution represents how a conflict was resolved
type ConflictResolution struct {
	Strategy     ResolutionStrategy `json:"strategy"`
	ResolvedBy   string             `json:"resolved_by"`
	ResolvedAt   time.Time          `json:"resolved_at"`
	FinalContent string             `json:"final_content"`
	Comments     string             `json:"comments"`
	Manual       bool               `json:"manual"`
}

// ResolutionStrategy represents different resolution strategies
type ResolutionStrategy string

const (
	StrategyTakeLocal      ResolutionStrategy = "take_local"
	StrategyTakeRemote     ResolutionStrategy = "take_remote"
	StrategyTakeBase       ResolutionStrategy = "take_base"
	StrategyManualMerge    ResolutionStrategy = "manual_merge"
	StrategyAutoMerge      ResolutionStrategy = "auto_merge"
	StrategySmartMerge     ResolutionStrategy = "smart_merge"
	StrategyCustomStrategy ResolutionStrategy = "custom"
)

// MergeResult represents the result of a merge operation
type MergeResult struct {
	Success       bool              `json:"success"`
	ConflictCount int               `json:"conflict_count"`
	Conflicts     []*MergeConflict  `json:"conflicts,omitempty"`
	MergedFiles   map[string]string `json:"merged_files"`
	Message       string            `json:"message"`
	CommitID      string            `json:"commit_id,omitempty"`
}

// MergeOptions contains options for merge operations
type MergeOptions struct {
	SourceBranch  string                        `json:"source_branch"`
	TargetBranch  string                        `json:"target_branch"`
	Message       string                        `json:"message"`
	Author        string                        `json:"author"`
	Strategy      ResolutionStrategy            `json:"strategy"`
	AutoResolve   bool                          `json:"auto_resolve"`
	ConflictRules map[string]ResolutionStrategy `json:"conflict_rules"`
	IgnoreFiles   []string                      `json:"ignore_files"`
}

// AttemptMerge attempts to merge two branches with conflict detection
func (mr *MergeResolver) AttemptMerge(ctx context.Context, options *MergeOptions) (*MergeResult, error) {
	// Get snapshots from both branches
	sourceSnapshot, err := mr.getBranchHead(ctx, options.SourceBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get source branch head: %w", err)
	}

	targetSnapshot, err := mr.getBranchHead(ctx, options.TargetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get target branch head: %w", err)
	}

	// Detect conflicts
	conflicts, err := mr.detectConflicts(sourceSnapshot, targetSnapshot, options)
	if err != nil {
		return nil, fmt.Errorf("failed to detect conflicts: %w", err)
	}

	result := &MergeResult{
		ConflictCount: len(conflicts),
		Conflicts:     conflicts,
		MergedFiles:   make(map[string]string),
	}

	// If no conflicts, perform the merge
	if len(conflicts) == 0 {
		return mr.performCleanMerge(ctx, sourceSnapshot, targetSnapshot, options)
	}

	// If auto-resolve is enabled, try to resolve conflicts
	if options.AutoResolve {
		resolvedConflicts, mergedFiles, err := mr.autoResolveConflicts(conflicts, options)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-resolve conflicts: %w", err)
		}

		// Check if all conflicts were resolved
		unresolvedCount := 0
		for _, conflict := range resolvedConflicts {
			if conflict.Resolution == nil {
				unresolvedCount++
			} else {
				mergedFiles[conflict.Filename] = conflict.Resolution.FinalContent
			}
		}

		result.Conflicts = resolvedConflicts
		result.MergedFiles = mergedFiles

		if unresolvedCount == 0 {
			// All conflicts resolved, create merge commit
			commitID, err := mr.createMergeCommit(ctx, mergedFiles, options)
			if err != nil {
				return nil, fmt.Errorf("failed to create merge commit: %w", err)
			}
			result.Success = true
			result.CommitID = commitID
			result.Message = "Merge completed successfully with auto-resolution"
		} else {
			result.Success = false
			result.Message = fmt.Sprintf("%d conflicts require manual resolution", unresolvedCount)
		}
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("Merge blocked by %d conflicts", len(conflicts))
	}

	return result, nil
}

// detectConflicts identifies conflicts between two snapshots
func (mr *MergeResolver) detectConflicts(source, target *ConfigSnapshot, options *MergeOptions) ([]*MergeConflict, error) {
	var conflicts []*MergeConflict

	// Get all unique files from both snapshots
	allFiles := make(map[string]bool)
	for filename := range source.Files {
		allFiles[filename] = true
	}
	for filename := range target.Files {
		allFiles[filename] = true
	}

	for filename := range allFiles {
		// Skip ignored files
		if mr.shouldIgnoreFile(filename, options.IgnoreFiles) {
			continue
		}

		sourceContent, sourceExists := source.Files[filename]
		targetContent, targetExists := target.Files[filename]

		conflict := mr.analyzeFileConflict(filename, sourceContent, sourceExists, targetContent, targetExists)
		if conflict != nil {
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts, nil
}

// analyzeFileConflict analyzes potential conflicts for a single file
func (mr *MergeResolver) analyzeFileConflict(filename, sourceContent string, sourceExists bool, targetContent string, targetExists bool) *MergeConflict {
	if !sourceExists && !targetExists {
		return nil // File doesn't exist in either branch
	}

	if sourceExists && targetExists {
		if sourceContent == targetContent {
			return nil // Files are identical
		}

		// Content conflict
		return &MergeConflict{
			Filename:      filename,
			ConflictType:  ConflictTypeContent,
			LocalContent:  targetContent,
			RemoteContent: sourceContent,
			Metadata: map[string]string{
				"source_size": fmt.Sprintf("%d", len(sourceContent)),
				"target_size": fmt.Sprintf("%d", len(targetContent)),
			},
		}
	}

	if sourceExists && !targetExists {
		// File exists in source but not target (add/delete conflict)
		return &MergeConflict{
			Filename:      filename,
			ConflictType:  ConflictTypeDeleteEdit,
			RemoteContent: sourceContent,
			Metadata: map[string]string{
				"conflict_reason": "file_deleted_in_target",
			},
		}
	}

	if !sourceExists && targetExists {
		// File exists in target but not source (delete/add conflict)
		return &MergeConflict{
			Filename:     filename,
			ConflictType: ConflictTypeEditDelete,
			LocalContent: targetContent,
			Metadata: map[string]string{
				"conflict_reason": "file_deleted_in_source",
			},
		}
	}

	return nil
}

// autoResolveConflicts attempts to automatically resolve conflicts
func (mr *MergeResolver) autoResolveConflicts(conflicts []*MergeConflict, options *MergeOptions) ([]*MergeConflict, map[string]string, error) {
	resolvedFiles := make(map[string]string)

	for _, conflict := range conflicts {
		strategy := mr.determineResolutionStrategy(conflict, options)
		resolution, err := mr.applyResolutionStrategy(conflict, strategy)
		if err != nil {
			mr.logger.Warn(fmt.Sprintf("Failed to auto-resolve conflict in %s: %v", conflict.Filename, err))
			continue
		}

		if resolution != nil {
			conflict.Resolution = resolution
			resolvedFiles[conflict.Filename] = resolution.FinalContent
			mr.logger.Info(fmt.Sprintf("Auto-resolved conflict in %s using %s strategy",
				conflict.Filename, strategy))
		}
	}

	return conflicts, resolvedFiles, nil
}

// determineResolutionStrategy determines the best strategy for a conflict
func (mr *MergeResolver) determineResolutionStrategy(conflict *MergeConflict, options *MergeOptions) ResolutionStrategy {
	// Check for file-specific rules
	if strategy, exists := options.ConflictRules[conflict.Filename]; exists {
		return strategy
	}

	// Check for pattern-based rules
	for pattern, strategy := range options.ConflictRules {
		if mr.matchesPattern(conflict.Filename, pattern) {
			return strategy
		}
	}

	// Default strategy
	if options.Strategy != "" {
		return options.Strategy
	}

	// Smart defaults based on conflict type
	switch conflict.ConflictType {
	case ConflictTypeDeleteEdit:
		return StrategyTakeRemote // Prefer additions
	case ConflictTypeEditDelete:
		return StrategyTakeLocal // Prefer keeping existing
	case ConflictTypeContent:
		return mr.smartContentMerge(conflict)
	default:
		return StrategyManualMerge
	}
}

// smartContentMerge attempts intelligent content merging
func (mr *MergeResolver) smartContentMerge(conflict *MergeConflict) ResolutionStrategy {
	// Check if this is a NixOS configuration file
	if strings.HasSuffix(conflict.Filename, ".nix") {
		return mr.smartNixMerge(conflict)
	}

	// Check for YAML files
	if strings.HasSuffix(conflict.Filename, ".yaml") || strings.HasSuffix(conflict.Filename, ".yml") {
		return StrategySmartMerge
	}

	// Default to manual for complex content conflicts
	return StrategyManualMerge
}

// smartNixMerge provides intelligent merging for Nix files
func (mr *MergeResolver) smartNixMerge(conflict *MergeConflict) ResolutionStrategy {
	localLines := strings.Split(conflict.LocalContent, "\n")
	remoteLines := strings.Split(conflict.RemoteContent, "\n")

	// Simple heuristics for Nix files
	localHasPackages := mr.containsNixSection(localLines, "environment.systemPackages")
	remoteHasPackages := mr.containsNixSection(remoteLines, "environment.systemPackages")

	localHasServices := mr.containsNixSection(localLines, "services.")
	remoteHasServices := mr.containsNixSection(remoteLines, "services.")

	// If both have different sections, try smart merge
	if (localHasPackages && remoteHasServices) || (localHasServices && remoteHasPackages) {
		return StrategySmartMerge
	}

	return StrategyManualMerge
}

// applyResolutionStrategy applies a resolution strategy to a conflict
func (mr *MergeResolver) applyResolutionStrategy(conflict *MergeConflict, strategy ResolutionStrategy) (*ConflictResolution, error) {
	resolution := &ConflictResolution{
		Strategy:   strategy,
		ResolvedBy: "nixai-auto-resolver",
		ResolvedAt: time.Now(),
		Manual:     false,
	}

	switch strategy {
	case StrategyTakeLocal:
		resolution.FinalContent = conflict.LocalContent
		resolution.Comments = "Automatically resolved by taking local version"

	case StrategyTakeRemote:
		resolution.FinalContent = conflict.RemoteContent
		resolution.Comments = "Automatically resolved by taking remote version"

	case StrategyTakeBase:
		resolution.FinalContent = conflict.BaseContent
		resolution.Comments = "Automatically resolved by taking base version"

	case StrategySmartMerge:
		merged, err := mr.performSmartMerge(conflict)
		if err != nil {
			return nil, fmt.Errorf("smart merge failed: %w", err)
		}
		resolution.FinalContent = merged
		resolution.Comments = "Automatically resolved using smart merge"

	case StrategyAutoMerge:
		merged, err := mr.performAutoMerge(conflict)
		if err != nil {
			return nil, fmt.Errorf("auto merge failed: %w", err)
		}
		resolution.FinalContent = merged
		resolution.Comments = "Automatically resolved using auto merge"

	default:
		return nil, nil // Cannot auto-resolve
	}

	return resolution, nil
}

// performSmartMerge performs intelligent merging based on file content
func (mr *MergeResolver) performSmartMerge(conflict *MergeConflict) (string, error) {
	// For Nix files, try to merge different sections
	if strings.HasSuffix(conflict.Filename, ".nix") {
		return mr.mergeNixContent(conflict.LocalContent, conflict.RemoteContent)
	}

	// For other files, use simple line-based merging
	return mr.mergeLinesBased(conflict.LocalContent, conflict.RemoteContent)
}

// mergeNixContent intelligently merges Nix configuration content
func (mr *MergeResolver) mergeNixContent(local, remote string) (string, error) {
	// This is a simplified implementation
	// In a real system, you'd parse the Nix AST and merge semantically

	localLines := strings.Split(local, "\n")
	remoteLines := strings.Split(remote, "\n")

	// Simple approach: combine unique lines
	merged := make(map[string]bool)
	var result []string

	// Add lines from local first
	for _, line := range localLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !merged[trimmed] {
			result = append(result, line)
			merged[trimmed] = true
		}
	}

	// Add unique lines from remote
	for _, line := range remoteLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !merged[trimmed] {
			result = append(result, line)
			merged[trimmed] = true
		}
	}

	return strings.Join(result, "\n"), nil
}

// Helper functions

func (mr *MergeResolver) getBranchHead(ctx context.Context, branchName string) (*ConfigSnapshot, error) {
	// Switch to branch and get latest snapshot
	currentBranch, err := mr.repo.GetCurrentBranch(ctx)
	if err != nil {
		return nil, err
	}

	if currentBranch != branchName {
		if err := mr.repo.SwitchBranch(ctx, branchName); err != nil {
			return nil, err
		}
		defer func() {
			mr.repo.SwitchBranch(ctx, currentBranch)
		}()
	}

	snapshots, err := mr.repo.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found in branch %s", branchName)
	}

	return snapshots[0], nil // Latest snapshot
}

func (mr *MergeResolver) shouldIgnoreFile(filename string, ignoreFiles []string) bool {
	for _, pattern := range ignoreFiles {
		if mr.matchesPattern(filename, pattern) {
			return true
		}
	}
	return false
}

func (mr *MergeResolver) matchesPattern(filename, pattern string) bool {
	// Simple pattern matching - in a real implementation, use proper glob matching
	if pattern == filename {
		return true
	}

	if strings.Contains(pattern, "*") {
		prefix := strings.Split(pattern, "*")[0]
		return strings.HasPrefix(filename, prefix)
	}

	return false
}

func (mr *MergeResolver) containsNixSection(lines []string, section string) bool {
	for _, line := range lines {
		if strings.Contains(line, section) {
			return true
		}
	}
	return false
}

func (mr *MergeResolver) performAutoMerge(conflict *MergeConflict) (string, error) {
	// Simple auto-merge: prefer larger content
	if len(conflict.RemoteContent) > len(conflict.LocalContent) {
		return conflict.RemoteContent, nil
	}
	return conflict.LocalContent, nil
}

func (mr *MergeResolver) mergeLinesBased(local, remote string) (string, error) {
	localLines := strings.Split(local, "\n")
	remoteLines := strings.Split(remote, "\n")

	// Simple merge: combine unique lines
	seen := make(map[string]bool)
	var result []string

	for _, line := range localLines {
		if !seen[line] {
			result = append(result, line)
			seen[line] = true
		}
	}

	for _, line := range remoteLines {
		if !seen[line] {
			result = append(result, line)
			seen[line] = true
		}
	}

	return strings.Join(result, "\n"), nil
}

func (mr *MergeResolver) performCleanMerge(ctx context.Context, source, target *ConfigSnapshot, options *MergeOptions) (*MergeResult, error) {
	// Merge all files (no conflicts detected)
	mergedFiles := make(map[string]string)

	// Start with target files
	for filename, content := range target.Files {
		mergedFiles[filename] = content
	}

	// Add/overwrite with source files
	for filename, content := range source.Files {
		mergedFiles[filename] = content
	}

	// Create merge commit
	commitID, err := mr.createMergeCommit(ctx, mergedFiles, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge commit: %w", err)
	}

	return &MergeResult{
		Success:     true,
		MergedFiles: mergedFiles,
		CommitID:    commitID,
		Message:     "Clean merge completed successfully",
	}, nil
}

func (mr *MergeResolver) createMergeCommit(ctx context.Context, files map[string]string, options *MergeOptions) (string, error) {
	message := options.Message
	if message == "" {
		message = fmt.Sprintf("Merge branch '%s' into '%s'", options.SourceBranch, options.TargetBranch)
	}

	metadata := map[string]string{
		"merge_source": options.SourceBranch,
		"merge_target": options.TargetBranch,
		"merge_type":   "automatic",
	}

	snapshot, err := mr.repo.Commit(ctx, message, files, metadata)
	if err != nil {
		return "", err
	}

	return snapshot.ID, nil
}
