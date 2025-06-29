package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nix-ai-help/pkg/logger"
)

// ConfigRepository manages simple configuration versioning
type ConfigRepository struct {
	path   string
	logger *logger.Logger
}

// ConfigSnapshot represents a snapshot of configuration files
type ConfigSnapshot struct {
	ID         string            `json:"id"`
	Message    string            `json:"message"`
	Author     string            `json:"author"`
	Timestamp  time.Time         `json:"timestamp"`
	Files      map[string]string `json:"files"`
	Tags       []string          `json:"tags"`
	Metadata   map[string]string `json:"metadata"`
	ParentHash string            `json:"parent_hash"`
}

// NewConfigRepository creates a new configuration repository
func NewConfigRepository(path string, logger *logger.Logger) (*ConfigRepository, error) {
	repo := &ConfigRepository{
		path:   path,
		logger: logger,
	}

	if err := repo.init(); err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	return repo, nil
}

// init initializes the repository directory
func (cr *ConfigRepository) init() error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(cr.path, 0755); err != nil {
		return fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Create snapshots directory
	snapshotsDir := filepath.Join(cr.path, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	cr.logger.Info("Repository initialized at " + cr.path)
	return nil
}

// CreateSnapshot creates a new snapshot of the specified files
func (cr *ConfigRepository) CreateSnapshot(ctx context.Context, message, author string, filePaths []string) (*ConfigSnapshot, error) {
	snapshot := &ConfigSnapshot{
		ID:        cr.generateSnapshotID(),
		Message:   message,
		Author:    author,
		Timestamp: time.Now(),
		Files:     make(map[string]string),
		Tags:      []string{},
		Metadata:  make(map[string]string),
	}

	// Read file contents
	for _, filePath := range filePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		snapshot.Files[filePath] = string(content)
	}

	// Save snapshot
	if err := cr.saveSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	cr.logger.Info(fmt.Sprintf("Created snapshot %s: %s", snapshot.ID[:8], message))
	return snapshot, nil
}

// GetSnapshot retrieves a snapshot by ID
func (cr *ConfigRepository) GetSnapshot(ctx context.Context, snapshotID string) (*ConfigSnapshot, error) {
	snapshotPath := filepath.Join(cr.path, "snapshots", snapshotID+".json")

	content, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot %s: %w", snapshotID, err)
	}

	var snapshot ConfigSnapshot
	if err := json.Unmarshal(content, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot %s: %w", snapshotID, err)
	}

	return &snapshot, nil
}

// ListSnapshots returns all snapshots
func (cr *ConfigRepository) ListSnapshots(ctx context.Context) ([]*ConfigSnapshot, error) {
	snapshotsDir := filepath.Join(cr.path, "snapshots")

	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	var snapshots []*ConfigSnapshot
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			snapshotID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
			snapshot, err := cr.GetSnapshot(ctx, snapshotID)
			if err != nil {
				cr.logger.Info(fmt.Sprintf("Failed to load snapshot %s: %v", snapshotID, err))
				continue
			}
			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots, nil
}

// CreateBranch creates a new branch (simplified implementation)
func (cr *ConfigRepository) CreateBranch(ctx context.Context, name string, fromCommit string) error {
	branchesDir := filepath.Join(cr.path, "branches")
	if err := os.MkdirAll(branchesDir, 0755); err != nil {
		return fmt.Errorf("failed to create branches directory: %w", err)
	}

	branchPath := filepath.Join(branchesDir, name)
	if err := os.WriteFile(branchPath, []byte(fromCommit), 0644); err != nil {
		return fmt.Errorf("failed to create branch %s: %w", name, err)
	}

	cr.logger.Info(fmt.Sprintf("Created branch: %s", name))
	return nil
}

// SwitchBranch switches to a different branch
func (cr *ConfigRepository) SwitchBranch(ctx context.Context, name string) error {
	branchPath := filepath.Join(cr.path, "branches", name)
	if _, err := os.Stat(branchPath); os.IsNotExist(err) {
		return fmt.Errorf("branch %s does not exist", name)
	}

	currentBranchPath := filepath.Join(cr.path, "current_branch")
	if err := os.WriteFile(currentBranchPath, []byte(name), 0644); err != nil {
		return fmt.Errorf("failed to switch to branch %s: %w", name, err)
	}

	cr.logger.Info(fmt.Sprintf("Switched to branch: %s", name))
	return nil
}

// DeleteBranch deletes a branch
func (cr *ConfigRepository) DeleteBranch(ctx context.Context, name string) error {
	branchPath := filepath.Join(cr.path, "branches", name)
	if err := os.Remove(branchPath); err != nil {
		return fmt.Errorf("failed to delete branch %s: %w", name, err)
	}

	cr.logger.Info(fmt.Sprintf("Deleted branch: %s", name))
	return nil
}

// ListBranches returns all branches
func (cr *ConfigRepository) ListBranches(ctx context.Context) ([]string, error) {
	branchesDir := filepath.Join(cr.path, "branches")

	entries, err := os.ReadDir(branchesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read branches directory: %w", err)
	}

	var branches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			branches = append(branches, entry.Name())
		}
	}

	return branches, nil
}

// TagSnapshot adds a tag to a snapshot
func (cr *ConfigRepository) TagSnapshot(ctx context.Context, snapshotID, tagName string) error {
	snapshot, err := cr.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("failed to get snapshot %s: %w", snapshotID, err)
	}

	snapshot.Tags = append(snapshot.Tags, tagName)

	if err := cr.saveSnapshot(snapshot); err != nil {
		return fmt.Errorf("failed to save tagged snapshot: %w", err)
	}

	cr.logger.Info(fmt.Sprintf("Tagged snapshot %s with tag: %s", snapshotID[:8], tagName))
	return nil
}

// ListTags returns all tags
func (cr *ConfigRepository) ListTags(ctx context.Context) (map[string]string, error) {
	snapshots, err := cr.ListSnapshots(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	tags := make(map[string]string)
	for _, snapshot := range snapshots {
		for _, tag := range snapshot.Tags {
			tags[tag] = snapshot.ID
		}
	}

	return tags, nil
}

// GetCurrentBranch gets the current active branch
func (cr *ConfigRepository) GetCurrentBranch(ctx context.Context) (string, error) {
	currentBranchPath := filepath.Join(cr.path, "current_branch")

	content, err := os.ReadFile(currentBranchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "main", nil // Default branch
		}
		return "", fmt.Errorf("failed to read current branch: %w", err)
	}

	return string(content), nil
}

// Commit creates a commit with files and metadata
func (cr *ConfigRepository) Commit(ctx context.Context, message string, files map[string]string, metadata map[string]string) (*ConfigSnapshot, error) {
	snapshot := &ConfigSnapshot{
		ID:        cr.generateSnapshotID(),
		Message:   message,
		Author:    "nixai", // Default author
		Timestamp: time.Now(),
		Files:     files,
		Tags:      []string{},
		Metadata:  metadata,
	}

	// Save snapshot
	if err := cr.saveSnapshot(snapshot); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	cr.logger.Info(fmt.Sprintf("Created commit %s: %s", snapshot.ID[:8], message))
	return snapshot, nil
}

// generateSnapshotID generates a unique snapshot ID
func (cr *ConfigRepository) generateSnapshotID() string {
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", timestamp)))
	return fmt.Sprintf("%x", hash)[:16]
}

// saveSnapshot saves a snapshot to disk
func (cr *ConfigRepository) saveSnapshot(snapshot *ConfigSnapshot) error {
	snapshotPath := filepath.Join(cr.path, "snapshots", snapshot.ID+".json")

	content, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if err := os.WriteFile(snapshotPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}
