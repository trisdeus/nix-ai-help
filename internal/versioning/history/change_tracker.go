package history

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/pkg/logger"
)

// ChangeTracker tracks configuration changes over time
type ChangeTracker struct {
	repo   *repository.ConfigRepository
	logger *logger.Logger
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker(repo *repository.ConfigRepository, logger *logger.Logger) *ChangeTracker {
	return &ChangeTracker{
		repo:   repo,
		logger: logger,
	}
}

// ChangeEvent represents a configuration change event
type ChangeEvent struct {
	ID        string                 `json:"id"`
	Type      ChangeEventType        `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	User      string                 `json:"user"`
	Branch    string                 `json:"branch"`
	CommitID  string                 `json:"commit_id"`
	Files     []string               `json:"files"`
	Summary   string                 `json:"summary"`
	Details   *ChangeDetails         `json:"details"`
	Impact    *ChangeImpact          `json:"impact"`
	Metadata  map[string]interface{} `json:"metadata"`
	Tags      []string               `json:"tags"`
}

// ChangeEventType represents different types of change events
type ChangeEventType string

const (
	EventTypeCommit     ChangeEventType = "commit"
	EventTypeMerge      ChangeEventType = "merge"
	EventTypeBranch     ChangeEventType = "branch"
	EventTypeTag        ChangeEventType = "tag"
	EventTypeRollback   ChangeEventType = "rollback"
	EventTypeDeployment ChangeEventType = "deployment"
	EventTypeBackup     ChangeEventType = "backup"
	EventTypeRestore    ChangeEventType = "restore"
)

// ChangeDetails contains detailed information about a change
type ChangeDetails struct {
	FilesAdded      []string             `json:"files_added"`
	FilesModified   []string             `json:"files_modified"`
	FilesRemoved    []string             `json:"files_removed"`
	LinesAdded      int                  `json:"lines_added"`
	LinesRemoved    int                  `json:"lines_removed"`
	ConfigSections  []string             `json:"config_sections"`
	Services        []string             `json:"services"`
	Packages        []string             `json:"packages"`
	SecurityChanges bool                 `json:"security_changes"`
	BreakingChanges bool                 `json:"breaking_changes"`
	Diff            map[string]*FileDiff `json:"diff,omitempty"`
}

// ChangeImpact represents the potential impact of a change
type ChangeImpact struct {
	Severity        ImpactSeverity `json:"severity"`
	AffectedSystems []string       `json:"affected_systems"`
	RequiresReboot  bool           `json:"requires_reboot"`
	RequiresReload  []string       `json:"requires_reload"`
	RiskLevel       RiskLevel      `json:"risk_level"`
	Rollbackable    bool           `json:"rollbackable"`
	TestingNeeded   bool           `json:"testing_needed"`
}

// ImpactSeverity represents the severity of a change
type ImpactSeverity string

const (
	SeverityLow      ImpactSeverity = "low"
	SeverityMedium   ImpactSeverity = "medium"
	SeverityHigh     ImpactSeverity = "high"
	SeverityCritical ImpactSeverity = "critical"
)

// RiskLevel represents the risk level of a change
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// FileDiff represents changes to a single file
type FileDiff struct {
	Filename     string      `json:"filename"`
	ChangeType   string      `json:"change_type"`
	LinesAdded   int         `json:"lines_added"`
	LinesRemoved int         `json:"lines_removed"`
	Hunks        []*DiffHunk `json:"hunks"`
}

// DiffHunk represents a contiguous block of changes
type DiffHunk struct {
	OldStart int      `json:"old_start"`
	OldLines int      `json:"old_lines"`
	NewStart int      `json:"new_start"`
	NewLines int      `json:"new_lines"`
	Lines    []string `json:"lines"`
}

// TrackCommit tracks a new commit
func (ct *ChangeTracker) TrackCommit(ctx context.Context, snapshot *repository.ConfigSnapshot) (*ChangeEvent, error) {
	// Get current branch
	branch, err := ct.repo.GetCurrentBranch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Analyze the change
	details, err := ct.analyzeCommitChanges(ctx, snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze changes: %w", err)
	}

	// Calculate impact
	impact := ct.calculateChangeImpact(details, snapshot)

	event := &ChangeEvent{
		ID:        fmt.Sprintf("%s-%d", snapshot.ID[:8], time.Now().Unix()),
		Type:      EventTypeCommit,
		Timestamp: snapshot.Timestamp,
		User:      snapshot.Author,
		Branch:    branch,
		CommitID:  snapshot.ID,
		Files:     ct.getFileList(snapshot.Files),
		Summary:   snapshot.Message,
		Details:   details,
		Impact:    impact,
		Metadata:  ct.convertMetadata(snapshot.Metadata),
		Tags:      snapshot.Tags,
	}

	ct.logger.Info(fmt.Sprintf("Tracked commit %s: %s", snapshot.ID[:8], snapshot.Message))
	return event, nil
}

// TrackMerge tracks a merge operation
func (ct *ChangeTracker) TrackMerge(ctx context.Context, sourceBranch, targetBranch string, mergeCommit *repository.ConfigSnapshot) (*ChangeEvent, error) {
	details, err := ct.analyzeCommitChanges(ctx, mergeCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze merge changes: %w", err)
	}

	impact := ct.calculateChangeImpact(details, mergeCommit)

	event := &ChangeEvent{
		ID:        fmt.Sprintf("merge-%s-%d", mergeCommit.ID[:8], time.Now().Unix()),
		Type:      EventTypeMerge,
		Timestamp: mergeCommit.Timestamp,
		User:      mergeCommit.Author,
		Branch:    targetBranch,
		CommitID:  mergeCommit.ID,
		Files:     ct.getFileList(mergeCommit.Files),
		Summary:   fmt.Sprintf("Merged %s into %s", sourceBranch, targetBranch),
		Details:   details,
		Impact:    impact,
		Metadata: map[string]interface{}{
			"source_branch": sourceBranch,
			"target_branch": targetBranch,
			"merge_type":    "branch_merge",
		},
	}

	ct.logger.Info(fmt.Sprintf("Tracked merge from %s to %s", sourceBranch, targetBranch))
	return event, nil
}

// TrackDeployment tracks a deployment event
func (ct *ChangeTracker) TrackDeployment(ctx context.Context, commitID, environment, target string) (*ChangeEvent, error) {
	snapshot, err := ct.repo.GetSnapshot(ctx, commitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	event := &ChangeEvent{
		ID:        fmt.Sprintf("deploy-%s-%d", commitID[:8], time.Now().Unix()),
		Type:      EventTypeDeployment,
		Timestamp: time.Now(),
		User:      "system", // Deployments are usually automated
		CommitID:  commitID,
		Files:     ct.getFileList(snapshot.Files),
		Summary:   fmt.Sprintf("Deployed to %s environment on %s", environment, target),
		Metadata: map[string]interface{}{
			"environment": environment,
			"target":      target,
			"deployed_at": time.Now().Format(time.RFC3339),
		},
	}

	ct.logger.Info(fmt.Sprintf("Tracked deployment of %s to %s", commitID[:8], environment))
	return event, nil
}

// analyzeCommitChanges analyzes the changes in a commit
func (ct *ChangeTracker) analyzeCommitChanges(ctx context.Context, snapshot *repository.ConfigSnapshot) (*ChangeDetails, error) {
	details := &ChangeDetails{
		FilesAdded:    []string{},
		FilesModified: []string{},
		FilesRemoved:  []string{},
		Diff:          make(map[string]*FileDiff),
	}

	// Get parent commit for comparison
	if snapshot.ParentHash != "" {
		parentSnapshot, err := ct.repo.GetSnapshot(ctx, snapshot.ParentHash)
		if err != nil {
			ct.logger.Warn(fmt.Sprintf("Could not get parent snapshot: %v", err))
		} else {
			ct.compareSnapshots(parentSnapshot, snapshot, details)
		}
	} else {
		// First commit - all files are added
		for filename := range snapshot.Files {
			details.FilesAdded = append(details.FilesAdded, filename)
		}
	}

	// Analyze configuration content
	ct.analyzeConfigurationContent(snapshot, details)

	return details, nil
}

// compareSnapshots compares two snapshots to identify changes
func (ct *ChangeTracker) compareSnapshots(parent, current *repository.ConfigSnapshot, details *ChangeDetails) {
	// Find added, modified, and removed files
	for filename, content := range current.Files {
		if parentContent, exists := parent.Files[filename]; exists {
			if parentContent != content {
				details.FilesModified = append(details.FilesModified, filename)
				details.Diff[filename] = ct.createFileDiff(filename, parentContent, content)
			}
		} else {
			details.FilesAdded = append(details.FilesAdded, filename)
			details.Diff[filename] = ct.createFileDiff(filename, "", content)
		}
	}

	for filename := range parent.Files {
		if _, exists := current.Files[filename]; !exists {
			details.FilesRemoved = append(details.FilesRemoved, filename)
			details.Diff[filename] = ct.createFileDiff(filename, parent.Files[filename], "")
		}
	}

	// Calculate line counts
	for _, fileDiff := range details.Diff {
		details.LinesAdded += fileDiff.LinesAdded
		details.LinesRemoved += fileDiff.LinesRemoved
	}
}

// createFileDiff creates a file diff structure
func (ct *ChangeTracker) createFileDiff(filename, oldContent, newContent string) *FileDiff {
	diff := &FileDiff{
		Filename: filename,
	}

	if oldContent == "" && newContent != "" {
		diff.ChangeType = "added"
		diff.LinesAdded = ct.countLines(newContent)
	} else if oldContent != "" && newContent == "" {
		diff.ChangeType = "removed"
		diff.LinesRemoved = ct.countLines(oldContent)
	} else {
		diff.ChangeType = "modified"
		diff.LinesAdded, diff.LinesRemoved = ct.calculateLineDiff(oldContent, newContent)
	}

	return diff
}

// analyzeConfigurationContent analyzes the configuration content for specific changes
func (ct *ChangeTracker) analyzeConfigurationContent(snapshot *repository.ConfigSnapshot, details *ChangeDetails) {
	for filename, content := range snapshot.Files {
		// Analyze Nix files
		if ct.isNixFile(filename) {
			ct.analyzeNixContent(content, details)
		}

		// Check for security-related changes
		if ct.containsSecurityKeywords(content) {
			details.SecurityChanges = true
		}

		// Check for breaking changes
		if ct.containsBreakingKeywords(content) {
			details.BreakingChanges = true
		}
	}
}

// analyzeNixContent analyzes Nix configuration content
func (ct *ChangeTracker) analyzeNixContent(content string, details *ChangeDetails) {
	lines := ct.splitLines(content)

	for _, line := range lines {
		line = ct.trimSpace(line)

		// Detect configuration sections
		if ct.contains(line, "environment.systemPackages") {
			details.ConfigSections = ct.appendUnique(details.ConfigSections, "systemPackages")
		}
		if ct.contains(line, "services.") {
			details.ConfigSections = ct.appendUnique(details.ConfigSections, "services")
		}
		if ct.contains(line, "networking.") {
			details.ConfigSections = ct.appendUnique(details.ConfigSections, "networking")
		}
		if ct.contains(line, "users.") {
			details.ConfigSections = ct.appendUnique(details.ConfigSections, "users")
		}

		// Extract service names
		if ct.contains(line, "services.") && ct.contains(line, ".enable = true") {
			serviceName := ct.extractServiceName(line)
			if serviceName != "" {
				details.Services = ct.appendUnique(details.Services, serviceName)
			}
		}

		// Extract package names (simplified)
		if ct.contains(line, "pkgs.") {
			packageName := ct.extractPackageName(line)
			if packageName != "" {
				details.Packages = ct.appendUnique(details.Packages, packageName)
			}
		}
	}
}

// calculateChangeImpact calculates the potential impact of changes
func (ct *ChangeTracker) calculateChangeImpact(details *ChangeDetails, snapshot *repository.ConfigSnapshot) *ChangeImpact {
	impact := &ChangeImpact{
		Severity:        SeverityLow,
		AffectedSystems: []string{},
		RequiresReload:  []string{},
		RiskLevel:       RiskLow,
		Rollbackable:    true,
		TestingNeeded:   false,
	}

	// Increase severity based on change scope
	if len(details.FilesModified)+len(details.FilesAdded)+len(details.FilesRemoved) > 10 {
		impact.Severity = SeverityHigh
		impact.RiskLevel = RiskHigh
		impact.TestingNeeded = true
	} else if len(details.FilesModified)+len(details.FilesAdded)+len(details.FilesRemoved) > 5 {
		impact.Severity = SeverityMedium
		impact.RiskLevel = RiskMedium
	}

	// Check for critical changes
	if details.SecurityChanges {
		impact.Severity = SeverityHigh
		impact.RiskLevel = RiskHigh
		impact.TestingNeeded = true
	}

	if details.BreakingChanges {
		impact.Severity = SeverityCritical
		impact.RiskLevel = RiskCritical
		impact.TestingNeeded = true
		impact.Rollbackable = false // Breaking changes might not be easily rollbackable
	}

	// Determine reboot requirements
	for _, service := range details.Services {
		if ct.requiresReboot(service) {
			impact.RequiresReboot = true
		} else {
			impact.RequiresReload = append(impact.RequiresReload, service)
		}
	}

	// Determine affected systems
	for _, section := range details.ConfigSections {
		impact.AffectedSystems = append(impact.AffectedSystems, section)
	}

	return impact
}

// Helper functions

func (ct *ChangeTracker) getFileList(files map[string]string) []string {
	var fileList []string
	for filename := range files {
		fileList = append(fileList, filename)
	}
	return fileList
}

func (ct *ChangeTracker) convertMetadata(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

func (ct *ChangeTracker) isNixFile(filename string) bool {
	return ct.hasSuffix(filename, ".nix")
}

func (ct *ChangeTracker) containsSecurityKeywords(content string) bool {
	keywords := []string{"firewall", "security", "auth", "password", "key", "cert", "ssl", "tls"}
	contentLower := ct.toLower(content)
	for _, keyword := range keywords {
		if ct.contains(contentLower, keyword) {
			return true
		}
	}
	return false
}

func (ct *ChangeTracker) containsBreakingKeywords(content string) bool {
	keywords := []string{"remove", "delete", "disable", "deprecat"}
	contentLower := ct.toLower(content)
	for _, keyword := range keywords {
		if ct.contains(contentLower, keyword) {
			return true
		}
	}
	return false
}

func (ct *ChangeTracker) extractServiceName(line string) string {
	// Extract service name from lines like "services.nginx.enable = true"
	if !ct.contains(line, "services.") {
		return ""
	}

	start := ct.findIndex(line, "services.") + len("services.")
	end := ct.findIndex(line[start:], ".")
	if end == -1 {
		return ""
	}

	return line[start : start+end]
}

func (ct *ChangeTracker) extractPackageName(line string) string {
	// Extract package name from lines containing "pkgs."
	if !ct.contains(line, "pkgs.") {
		return ""
	}

	start := ct.findIndex(line, "pkgs.") + len("pkgs.")
	// Find the end of the package name (space, semicolon, etc.)
	end := len(line)
	for i, char := range line[start:] {
		if char == ' ' || char == ';' || char == ')' || char == ']' {
			end = start + i
			break
		}
	}

	if end > start {
		return line[start:end]
	}
	return ""
}

func (ct *ChangeTracker) requiresReboot(service string) bool {
	rebootServices := []string{"kernel", "systemd", "networking", "firewall"}
	for _, rebootService := range rebootServices {
		if ct.contains(service, rebootService) {
			return true
		}
	}
	return false
}

func (ct *ChangeTracker) appendUnique(slice []string, item string) []string {
	for _, existing := range slice {
		if existing == item {
			return slice
		}
	}
	return append(slice, item)
}

func (ct *ChangeTracker) countLines(content string) int {
	if content == "" {
		return 0
	}
	count := 1
	for _, char := range content {
		if char == '\n' {
			count++
		}
	}
	return count
}

func (ct *ChangeTracker) calculateLineDiff(oldContent, newContent string) (added, removed int) {
	oldLines := ct.countLines(oldContent)
	newLines := ct.countLines(newContent)

	if newLines > oldLines {
		added = newLines - oldLines
	} else if oldLines > newLines {
		removed = oldLines - newLines
	}

	return added, removed
}

// String utility functions
func (ct *ChangeTracker) splitLines(s string) []string {
	if s == "" {
		return []string{}
	}

	var lines []string
	var current string

	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func (ct *ChangeTracker) trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

func (ct *ChangeTracker) contains(s, substr string) bool {
	return ct.findIndex(s, substr) != -1
}

func (ct *ChangeTracker) findIndex(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (ct *ChangeTracker) hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func (ct *ChangeTracker) toLower(s string) string {
	var result string
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			result += string(char + 'a' - 'A')
		} else {
			result += string(char)
		}
	}
	return result
}
