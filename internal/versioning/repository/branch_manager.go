package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// BranchManager handles configuration branch management
type BranchManager struct {
	repo   *ConfigRepository
	logger *logger.Logger
}

// NewBranchManager creates a new branch manager
func NewBranchManager(repo *ConfigRepository, logger *logger.Logger) *BranchManager {
	return &BranchManager{
		repo:   repo,
		logger: logger,
	}
}

// BranchInfo contains information about a branch
type BranchInfo struct {
	Name        string    `json:"name"`
	Current     bool      `json:"current"`
	LastCommit  string    `json:"last_commit"`
	CommitHash  string    `json:"commit_hash"`
	Author      string    `json:"author"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	Environment string    `json:"environment"` // production, staging, development
	Protected   bool      `json:"protected"`
}

// CreateFeatureBranch creates a new feature branch with standard naming
func (bm *BranchManager) CreateFeatureBranch(ctx context.Context, featureName string, description string) (*BranchInfo, error) {
	branchName := fmt.Sprintf("feature/%s", sanitizeBranchName(featureName))

	// Check if branch already exists
	branches, err := bm.repo.ListBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	for _, branch := range branches {
		if branch == branchName {
			return nil, fmt.Errorf("branch %s already exists", branchName)
		}
	}

	// Create branch from current HEAD
	if err := bm.repo.CreateBranch(ctx, branchName, ""); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Switch to new branch
	if err := bm.repo.SwitchBranch(ctx, branchName); err != nil {
		return nil, fmt.Errorf("failed to switch to branch: %w", err)
	}

	bm.logger.Info(fmt.Sprintf("Created feature branch: %s", branchName))

	return &BranchInfo{
		Name:        branchName,
		Current:     true,
		Description: description,
		Environment: "development",
		Protected:   false,
		Timestamp:   time.Now(),
	}, nil
}

// CreateEnvironmentBranch creates a branch for a specific environment
func (bm *BranchManager) CreateEnvironmentBranch(ctx context.Context, environment string, fromBranch string) (*BranchInfo, error) {
	branchName := fmt.Sprintf("env/%s", sanitizeBranchName(environment))

	// Get commit hash from source branch
	var fromCommit string
	if fromBranch != "" {
		// Switch to source branch temporarily to get its HEAD
		currentBranch, err := bm.repo.GetCurrentBranch(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get current branch: %w", err)
		}

		if currentBranch != fromBranch {
			if err := bm.repo.SwitchBranch(ctx, fromBranch); err != nil {
				return nil, fmt.Errorf("failed to switch to source branch: %w", err)
			}
			defer func() {
				bm.repo.SwitchBranch(ctx, currentBranch)
			}()
		}
	}

	// Create environment branch
	if err := bm.repo.CreateBranch(ctx, branchName, fromCommit); err != nil {
		return nil, fmt.Errorf("failed to create environment branch: %w", err)
	}

	bm.logger.Info(fmt.Sprintf("Created environment branch: %s", branchName))

	return &BranchInfo{
		Name:        branchName,
		Current:     false,
		Description: fmt.Sprintf("Configuration for %s environment", environment),
		Environment: environment,
		Protected:   environment == "production", // Protect production branches
		Timestamp:   time.Now(),
	}, nil
}

// ListBranchesWithInfo returns detailed information about all branches
func (bm *BranchManager) ListBranchesWithInfo(ctx context.Context) ([]*BranchInfo, error) {
	branches, err := bm.repo.ListBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	currentBranch, err := bm.repo.GetCurrentBranch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	var branchInfos []*BranchInfo
	for _, branch := range branches {
		branchInfo := &BranchInfo{
			Name:    branch,
			Current: branch == currentBranch,
		}

		// Determine environment and protection status
		if strings.HasPrefix(branch, "env/production") || branch == "main" || branch == "master" {
			branchInfo.Environment = "production"
			branchInfo.Protected = true
		} else if strings.HasPrefix(branch, "env/staging") {
			branchInfo.Environment = "staging"
			branchInfo.Protected = false
		} else if strings.HasPrefix(branch, "feature/") {
			branchInfo.Environment = "development"
			branchInfo.Protected = false
		} else {
			branchInfo.Environment = "unknown"
			branchInfo.Protected = false
		}

		// Set description based on branch type
		if strings.HasPrefix(branch, "feature/") {
			branchInfo.Description = fmt.Sprintf("Feature branch: %s", strings.TrimPrefix(branch, "feature/"))
		} else if strings.HasPrefix(branch, "env/") {
			env := strings.TrimPrefix(branch, "env/")
			branchInfo.Description = fmt.Sprintf("Environment configuration: %s", env)
		} else {
			branchInfo.Description = "Main configuration branch"
		}

		branchInfos = append(branchInfos, branchInfo)
	}

	return branchInfos, nil
}

// MergeBranch merges one branch into another
func (bm *BranchManager) MergeBranch(ctx context.Context, sourceBranch, targetBranch string, message string) error {
	// This is a simplified merge - in a real implementation, you'd want to handle conflicts
	currentBranch, err := bm.repo.GetCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Switch to target branch
	if currentBranch != targetBranch {
		if err := bm.repo.SwitchBranch(ctx, targetBranch); err != nil {
			return fmt.Errorf("failed to switch to target branch: %w", err)
		}
		defer func() {
			if currentBranch != targetBranch {
				bm.repo.SwitchBranch(ctx, currentBranch)
			}
		}()
	}

	// Get latest snapshot from source branch
	// This is a simplified implementation - real merge logic would be more complex
	if err := bm.repo.SwitchBranch(ctx, sourceBranch); err != nil {
		return fmt.Errorf("failed to switch to source branch: %w", err)
	}

	snapshots, err := bm.repo.ListSnapshots(ctx)
	if err != nil {
		return fmt.Errorf("failed to get snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		return fmt.Errorf("no snapshots found in source branch")
	}

	latestSnapshot := snapshots[0] // First snapshot is the latest

	// Switch back to target branch
	if err := bm.repo.SwitchBranch(ctx, targetBranch); err != nil {
		return fmt.Errorf("failed to switch back to target branch: %w", err)
	}

	// Create merge commit
	mergeMessage := fmt.Sprintf("Merge branch '%s' into '%s'", sourceBranch, targetBranch)
	if message != "" {
		mergeMessage = fmt.Sprintf("%s\n\n%s", mergeMessage, message)
	}

	_, err = bm.repo.Commit(ctx, mergeMessage, latestSnapshot.Files, map[string]string{
		"merge_source": sourceBranch,
		"merge_target": targetBranch,
		"merge_type":   "branch_merge",
	})

	if err != nil {
		return fmt.Errorf("failed to create merge commit: %w", err)
	}

	bm.logger.Info(fmt.Sprintf("Merged branch %s into %s", sourceBranch, targetBranch))
	return nil
}

// DeleteBranchSafe safely deletes a branch with protection checks
func (bm *BranchManager) DeleteBranchSafe(ctx context.Context, branchName string, force bool) error {
	branchInfos, err := bm.ListBranchesWithInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get branch info: %w", err)
	}

	var branchInfo *BranchInfo
	for _, info := range branchInfos {
		if info.Name == branchName {
			branchInfo = info
			break
		}
	}

	if branchInfo == nil {
		return fmt.Errorf("branch %s not found", branchName)
	}

	// Check if branch is protected
	if branchInfo.Protected && !force {
		return fmt.Errorf("branch %s is protected, use force flag to delete", branchName)
	}

	// Check if branch is current
	if branchInfo.Current {
		return fmt.Errorf("cannot delete current branch %s", branchName)
	}

	// Delete branch
	if err := bm.repo.DeleteBranch(ctx, branchName); err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}

	bm.logger.Info(fmt.Sprintf("Deleted branch: %s", branchName))
	return nil
}

// GetBranchHistory returns the commit history for a specific branch
func (bm *BranchManager) GetBranchHistory(ctx context.Context, branchName string) ([]*ConfigSnapshot, error) {
	currentBranch, err := bm.repo.GetCurrentBranch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Switch to target branch if needed
	if currentBranch != branchName {
		if err := bm.repo.SwitchBranch(ctx, branchName); err != nil {
			return nil, fmt.Errorf("failed to switch to branch: %w", err)
		}
		defer func() {
			bm.repo.SwitchBranch(ctx, currentBranch)
		}()
	}

	snapshots, err := bm.repo.ListSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	return snapshots, nil
}

// sanitizeBranchName sanitizes a branch name to be Git-compatible
func sanitizeBranchName(name string) string {
	// Replace spaces and special characters with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ToLower(name)

	// Remove any remaining problematic characters
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
