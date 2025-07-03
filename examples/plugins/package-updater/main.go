package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/plugins"
)

// PackageUpdaterPlugin manages package updates with AI assistance
type PackageUpdaterPlugin struct {
	config         plugins.PluginConfig
	running        bool
	mutex          sync.RWMutex
	availableUpdates []PackageUpdate
	updateHistory    []UpdateRecord
	lastCheck        time.Time
	updatePolicy     UpdatePolicy
}

type PackageUpdate struct {
	Name            string    `json:"name"`
	CurrentVersion  string    `json:"current_version"`
	AvailableVersion string   `json:"available_version"`
	Description     string    `json:"description"`
	Size            int64     `json:"size"`
	Priority        string    `json:"priority"` // critical, high, medium, low
	Category        string    `json:"category"` // security, feature, bugfix
	Dependencies    []string  `json:"dependencies"`
	BreakingChanges bool      `json:"breaking_changes"`
	ReleaseNotes    string    `json:"release_notes"`
	LastUpdated     time.Time `json:"last_updated"`
}

type UpdateRecord struct {
	PackageName     string    `json:"package_name"`
	FromVersion     string    `json:"from_version"`
	ToVersion       string    `json:"to_version"`
	Status          string    `json:"status"` // success, failed, skipped
	Timestamp       time.Time `json:"timestamp"`
	Duration        time.Duration `json:"duration"`
	ErrorMessage    string    `json:"error_message,omitempty"`
}

type UpdatePolicy struct {
	AutoUpdate       bool     `json:"auto_update"`
	SecurityOnly     bool     `json:"security_only"`
	AllowedCategories []string `json:"allowed_categories"`
	ExcludedPackages []string `json:"excluded_packages"`
	MaxUpdatesPerRun int      `json:"max_updates_per_run"`
	RequireApproval  bool     `json:"require_approval"`
	BackupBeforeUpdate bool   `json:"backup_before_update"`
}

type UpdatePlan struct {
	TotalUpdates     int              `json:"total_updates"`
	SecurityUpdates  int              `json:"security_updates"`
	BreakingUpdates  int              `json:"breaking_updates"`
	TotalSize        int64            `json:"total_size"`
	EstimatedTime    time.Duration    `json:"estimated_time"`
	Updates          []PackageUpdate  `json:"updates"`
	Warnings         []string         `json:"warnings"`
	Recommendations  []string         `json:"recommendations"`
}

// Metadata methods
func (p *PackageUpdaterPlugin) Name() string        { return "package-updater" }
func (p *PackageUpdaterPlugin) Version() string     { return "1.0.0" }
func (p *PackageUpdaterPlugin) Description() string { return "AI-powered package update management with smart scheduling" }
func (p *PackageUpdaterPlugin) Author() string      { return "NixAI Team" }
func (p *PackageUpdaterPlugin) Repository() string  { return "https://github.com/nixai/plugins/package-updater" }
func (p *PackageUpdaterPlugin) License() string     { return "MIT" }

func (p *PackageUpdaterPlugin) Dependencies() []string {
	return []string{"nix", "nixos-rebuild"}
}

func (p *PackageUpdaterPlugin) Capabilities() []string {
	return []string{
		"package-management",
		"update-scheduling",
		"security-analysis",
		"dependency-resolution",
		"rollback-support",
	}
}

// Lifecycle methods
func (p *PackageUpdaterPlugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.config = config
	p.availableUpdates = []PackageUpdate{}
	p.updateHistory = []UpdateRecord{}
	
	// Set default update policy
	p.updatePolicy = UpdatePolicy{
		AutoUpdate:         false,
		SecurityOnly:       false,
		AllowedCategories:  []string{"security", "bugfix", "feature"},
		ExcludedPackages:   []string{},
		MaxUpdatesPerRun:   10,
		RequireApproval:    true,
		BackupBeforeUpdate: true,
	}
	
	// Override with config values if provided
	if config.Configuration != nil {
		if policy, ok := config.Configuration["update_policy"]; ok {
			if policyMap, ok := policy.(map[string]interface{}); ok {
				p.loadUpdatePolicy(policyMap)
			}
		}
	}
	
	return nil
}

func (p *PackageUpdaterPlugin) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.running {
		return fmt.Errorf("plugin is already running")
	}
	
	// Start update checking goroutine
	go p.updateCheckLoop(ctx)
	
	p.running = true
	return nil
}

func (p *PackageUpdaterPlugin) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if !p.running {
		return fmt.Errorf("plugin is not running")
	}
	
	p.running = false
	return nil
}

func (p *PackageUpdaterPlugin) Cleanup(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.availableUpdates = nil
	p.updateHistory = nil
	
	return nil
}

func (p *PackageUpdaterPlugin) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

// Execution methods
func (p *PackageUpdaterPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "check-updates":
		return p.checkUpdates(ctx, params)
	case "list-updates":
		return p.listUpdates(ctx, params)
	case "create-update-plan":
		return p.createUpdatePlan(ctx, params)
	case "apply-updates":
		return p.applyUpdates(ctx, params)
	case "get-update-history":
		return p.getUpdateHistory(ctx, params)
	case "set-update-policy":
		return p.setUpdatePolicy(ctx, params)
	case "get-update-policy":
		return p.getUpdatePolicy(ctx, params)
	case "rollback-update":
		return p.rollbackUpdate(ctx, params)
	case "analyze-package":
		return p.analyzePackage(ctx, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *PackageUpdaterPlugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "check-updates",
			Description: "Check for available package updates",
			Parameters:  []plugins.PluginParameter{},
			ReturnType:  "UpdateSummary",
			Tags:        []string{"updates", "check"},
		},
		{
			Name:        "list-updates",
			Description: "List available package updates with details",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "category",
					Type:        "string",
					Description: "Filter by update category",
					Required:    false,
				},
				{
					Name:        "priority",
					Type:        "string",
					Description: "Filter by priority level",
					Required:    false,
				},
			},
			ReturnType: "[]PackageUpdate",
			Tags:       []string{"updates", "list"},
		},
		{
			Name:        "create-update-plan",
			Description: "Create an intelligent update plan",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "include_breaking",
					Type:        "boolean",
					Description: "Include updates with breaking changes",
					Required:    false,
					Default:     false,
				},
			},
			ReturnType: "UpdatePlan",
			Tags:       []string{"updates", "planning"},
		},
		{
			Name:        "apply-updates",
			Description: "Apply selected package updates",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "packages",
					Type:        "array",
					Description: "List of package names to update",
					Required:    false,
				},
				{
					Name:        "dry_run",
					Type:        "boolean",
					Description: "Simulate updates without applying",
					Required:    false,
					Default:     false,
				},
			},
			ReturnType: "UpdateResult",
			Tags:       []string{"updates", "apply"},
		},
		{
			Name:        "get-update-history",
			Description: "Get package update history",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "limit",
					Type:        "integer",
					Description: "Maximum number of records to return",
					Required:    false,
					Default:     50,
				},
			},
			ReturnType: "[]UpdateRecord",
			Tags:       []string{"history", "audit"},
		},
		{
			Name:        "analyze-package",
			Description: "Analyze a specific package for update details",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "package_name",
					Type:        "string",
					Description: "Name of the package to analyze",
					Required:    true,
				},
			},
			ReturnType: "PackageAnalysis",
			Tags:       []string{"analysis", "package"},
		},
	}
}

func (p *PackageUpdaterPlugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	schemas := map[string]*plugins.PluginSchema{
		"check-updates": {
			Type:       "object",
			Properties: map[string]plugins.PluginSchemaProperty{},
		},
		"list-updates": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"category": {
					Type:        "string",
					Description: "Filter by category",
					Enum:        []string{"security", "bugfix", "feature"},
				},
				"priority": {
					Type:        "string",
					Description: "Filter by priority",
					Enum:        []string{"critical", "high", "medium", "low"},
				},
			},
		},
		"apply-updates": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"packages": {
					Type:        "array",
					Description: "Package names to update",
				},
				"dry_run": {
					Type:        "boolean",
					Description: "Simulate without applying",
				},
			},
		},
	}
	
	if schema, exists := schemas[operation]; exists {
		return schema, nil
	}
	
	return nil, fmt.Errorf("unknown operation: %s", operation)
}

// Health and Status methods
func (p *PackageUpdaterPlugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	status := plugins.HealthHealthy
	message := "Package updater running normally"
	var issues []plugins.HealthIssue
	
	if !p.running {
		status = plugins.HealthUnhealthy
		message = "Package updater is not running"
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityError,
			Component: "updater",
			Message:   "Updater service stopped",
			Timestamp: time.Now(),
		})
	}
	
	// Check for pending security updates
	securityUpdates := 0
	for _, update := range p.availableUpdates {
		if update.Category == "security" {
			securityUpdates++
		}
	}
	
	if securityUpdates > 0 {
		status = plugins.HealthDegraded
		message = fmt.Sprintf("%d security updates available", securityUpdates)
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityWarning,
			Component: "security",
			Message:   fmt.Sprintf("%d security updates pending", securityUpdates),
			Timestamp: time.Now(),
		})
	}
	
	return plugins.PluginHealth{
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Issues:    issues,
	}
}

func (p *PackageUpdaterPlugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{
		ExecutionCount:       0,
		TotalExecutionTime:   0,
		AverageExecutionTime: 0,
		LastExecutionTime:    p.lastCheck,
		ErrorCount:           0,
		SuccessRate:          100.0,
		StartTime:            time.Now(),
		CustomMetrics: map[string]interface{}{
			"available_updates": len(p.availableUpdates),
			"last_check":        p.lastCheck,
			"security_updates":  p.countSecurityUpdates(),
			"update_history":    len(p.updateHistory),
		},
	}
}

func (p *PackageUpdaterPlugin) GetStatus() plugins.PluginStatus {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	state := plugins.StateRunning
	if !p.running {
		state = plugins.StateStopped
	}
	
	return plugins.PluginStatus{
		State:       state,
		Message:     "Package updater active",
		LastUpdated: time.Now(),
		Version:     p.Version(),
	}
}

// Operation implementations
func (p *PackageUpdaterPlugin) checkUpdates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Run nix-env -u --dry-run to check for updates
	cmd := exec.Command("nix-env", "-u", "--dry-run")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to check updates: %v", err)
	}
	
	// Parse output to extract available updates
	p.availableUpdates = p.parseNixUpdates(string(output))
	p.lastCheck = time.Now()
	
	summary := map[string]interface{}{
		"total_updates":    len(p.availableUpdates),
		"security_updates": p.countSecurityUpdates(),
		"breaking_updates": p.countBreakingUpdates(),
		"last_check":       p.lastCheck,
		"updates":          p.availableUpdates,
	}
	
	return summary, nil
}

func (p *PackageUpdaterPlugin) listUpdates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	updates := p.availableUpdates
	
	// Apply filters
	if category, ok := params["category"].(string); ok {
		var filtered []PackageUpdate
		for _, update := range updates {
			if update.Category == category {
				filtered = append(filtered, update)
			}
		}
		updates = filtered
	}
	
	if priority, ok := params["priority"].(string); ok {
		var filtered []PackageUpdate
		for _, update := range updates {
			if update.Priority == priority {
				filtered = append(filtered, update)
			}
		}
		updates = filtered
	}
	
	return updates, nil
}

func (p *PackageUpdaterPlugin) createUpdatePlan(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	includeBreaking := false
	if val, ok := params["include_breaking"].(bool); ok {
		includeBreaking = val
	}
	
	var planUpdates []PackageUpdate
	var warnings []string
	var recommendations []string
	
	// Filter updates based on policy and parameters
	for _, update := range p.availableUpdates {
		// Skip breaking changes if not explicitly included
		if update.BreakingChanges && !includeBreaking {
			warnings = append(warnings, fmt.Sprintf("Skipping %s due to breaking changes", update.Name))
			continue
		}
		
		// Check if package is excluded
		if p.isPackageExcluded(update.Name) {
			continue
		}
		
		// Check category allowlist
		if !p.isCategoryAllowed(update.Category) {
			continue
		}
		
		planUpdates = append(planUpdates, update)
	}
	
	// Sort by priority
	sort.Slice(planUpdates, func(i, j int) bool {
		return p.getPriorityWeight(planUpdates[i].Priority) > p.getPriorityWeight(planUpdates[j].Priority)
	})
	
	// Limit updates if policy specifies
	if len(planUpdates) > p.updatePolicy.MaxUpdatesPerRun {
		planUpdates = planUpdates[:p.updatePolicy.MaxUpdatesPerRun]
		warnings = append(warnings, fmt.Sprintf("Limited to %d updates per policy", p.updatePolicy.MaxUpdatesPerRun))
	}
	
	// Generate recommendations
	securityCount := 0
	for _, update := range planUpdates {
		if update.Category == "security" {
			securityCount++
		}
	}
	
	if securityCount > 0 {
		recommendations = append(recommendations, "Prioritize security updates for immediate installation")
	}
	
	if len(planUpdates) > 5 {
		recommendations = append(recommendations, "Consider applying updates in smaller batches")
	}
	
	totalSize := int64(0)
	for _, update := range planUpdates {
		totalSize += update.Size
	}
	
	plan := UpdatePlan{
		TotalUpdates:     len(planUpdates),
		SecurityUpdates:  securityCount,
		BreakingUpdates:  p.countBreakingInList(planUpdates),
		TotalSize:        totalSize,
		EstimatedTime:    time.Duration(len(planUpdates)*30) * time.Second, // Rough estimate
		Updates:          planUpdates,
		Warnings:         warnings,
		Recommendations:  recommendations,
	}
	
	return plan, nil
}

func (p *PackageUpdaterPlugin) applyUpdates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dryRun := false
	if val, ok := params["dry_run"].(bool); ok {
		dryRun = val
	}
	
	var packagesToUpdate []string
	if packages, ok := params["packages"].([]interface{}); ok {
		for _, pkg := range packages {
			if name, ok := pkg.(string); ok {
				packagesToUpdate = append(packagesToUpdate, name)
			}
		}
	} else {
		// Use all available updates if no specific packages specified
		for _, update := range p.availableUpdates {
			packagesToUpdate = append(packagesToUpdate, update.Name)
		}
	}
	
	results := make(map[string]interface{})
	results["dry_run"] = dryRun
	results["packages"] = packagesToUpdate
	results["status"] = "success"
	
	if dryRun {
		results["message"] = "Dry run completed - no packages were actually updated"
		return results, nil
	}
	
	// Apply updates
	startTime := time.Now()
	var successCount, failCount int
	var errors []string
	
	for _, packageName := range packagesToUpdate {
		err := p.updatePackage(packageName)
		if err != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("%s: %v", packageName, err))
			
			// Record failed update
			p.updateHistory = append(p.updateHistory, UpdateRecord{
				PackageName: packageName,
				Status:      "failed",
				Timestamp:   time.Now(),
				Duration:    time.Since(startTime),
				ErrorMessage: err.Error(),
			})
		} else {
			successCount++
			
			// Record successful update
			p.updateHistory = append(p.updateHistory, UpdateRecord{
				PackageName: packageName,
				Status:      "success",
				Timestamp:   time.Now(),
				Duration:    time.Since(startTime),
			})
		}
	}
	
	results["success_count"] = successCount
	results["fail_count"] = failCount
	results["errors"] = errors
	results["duration"] = time.Since(startTime).String()
	
	if failCount > 0 {
		results["status"] = "partial"
	}
	
	return results, nil
}

func (p *PackageUpdaterPlugin) getUpdateHistory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	limit := 50
	if val, ok := params["limit"].(float64); ok {
		limit = int(val)
	}
	
	history := p.updateHistory
	if len(history) > limit {
		history = history[len(history)-limit:]
	}
	
	return history, nil
}

func (p *PackageUpdaterPlugin) setUpdatePolicy(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if policy, ok := params["policy"].(map[string]interface{}); ok {
		p.loadUpdatePolicy(policy)
		return true, nil
	}
	
	return false, fmt.Errorf("invalid policy parameter")
}

func (p *PackageUpdaterPlugin) getUpdatePolicy(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return p.updatePolicy, nil
}

func (p *PackageUpdaterPlugin) rollbackUpdate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	packageName, ok := params["package_name"].(string)
	if !ok {
		return nil, fmt.Errorf("package_name parameter required")
	}
	
	// Implement rollback using nix-env --rollback
	cmd := exec.Command("nix-env", "--rollback")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("rollback failed: %v", err)
	}
	
	return map[string]interface{}{
		"status":       "success",
		"package_name": packageName,
		"output":       string(output),
	}, nil
}

func (p *PackageUpdaterPlugin) analyzePackage(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	packageName, ok := params["package_name"].(string)
	if !ok {
		return nil, fmt.Errorf("package_name parameter required")
	}
	
	// Find package in available updates
	for _, update := range p.availableUpdates {
		if update.Name == packageName {
			analysis := map[string]interface{}{
				"package":          update,
				"risk_level":       p.assessUpdateRisk(update),
				"dependencies":     update.Dependencies,
				"breaking_changes": update.BreakingChanges,
				"recommendations":  p.generatePackageRecommendations(update),
			}
			return analysis, nil
		}
	}
	
	return nil, fmt.Errorf("package %s not found in available updates", packageName)
}

// Helper methods
func (p *PackageUpdaterPlugin) parseNixUpdates(output string) []PackageUpdate {
	var updates []PackageUpdate
	
	// Simple parsing - in real implementation would be more sophisticated
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "->") {
			// Extract package info from nix output
			re := regexp.MustCompile(`(\S+)-(\S+)\s*->\s*(\S+)-(\S+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 5 {
				update := PackageUpdate{
					Name:             matches[1],
					CurrentVersion:   matches[2],
					AvailableVersion: matches[4],
					Priority:         "medium",
					Category:         "feature",
					LastUpdated:      time.Now(),
				}
				
				// Classify update type
				if strings.Contains(strings.ToLower(line), "security") {
					update.Category = "security"
					update.Priority = "high"
				}
				
				updates = append(updates, update)
			}
		}
	}
	
	return updates
}

func (p *PackageUpdaterPlugin) countSecurityUpdates() int {
	count := 0
	for _, update := range p.availableUpdates {
		if update.Category == "security" {
			count++
		}
	}
	return count
}

func (p *PackageUpdaterPlugin) countBreakingUpdates() int {
	count := 0
	for _, update := range p.availableUpdates {
		if update.BreakingChanges {
			count++
		}
	}
	return count
}

func (p *PackageUpdaterPlugin) countBreakingInList(updates []PackageUpdate) int {
	count := 0
	for _, update := range updates {
		if update.BreakingChanges {
			count++
		}
	}
	return count
}

func (p *PackageUpdaterPlugin) isPackageExcluded(packageName string) bool {
	for _, excluded := range p.updatePolicy.ExcludedPackages {
		if excluded == packageName {
			return true
		}
	}
	return false
}

func (p *PackageUpdaterPlugin) isCategoryAllowed(category string) bool {
	for _, allowed := range p.updatePolicy.AllowedCategories {
		if allowed == category {
			return true
		}
	}
	return false
}

func (p *PackageUpdaterPlugin) getPriorityWeight(priority string) int {
	switch priority {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func (p *PackageUpdaterPlugin) updatePackage(packageName string) error {
	cmd := exec.Command("nix-env", "-iA", "nixpkgs."+packageName)
	_, err := cmd.CombinedOutput()
	return err
}

func (p *PackageUpdaterPlugin) assessUpdateRisk(update PackageUpdate) string {
	if update.BreakingChanges {
		return "high"
	}
	if update.Category == "security" {
		return "low"
	}
	if update.Priority == "critical" {
		return "medium"
	}
	return "low"
}

func (p *PackageUpdaterPlugin) generatePackageRecommendations(update PackageUpdate) []string {
	var recommendations []string
	
	if update.Category == "security" {
		recommendations = append(recommendations, "Apply this security update immediately")
	}
	
	if update.BreakingChanges {
		recommendations = append(recommendations, "Test in development environment before applying")
		recommendations = append(recommendations, "Review breaking changes documentation")
	}
	
	if len(update.Dependencies) > 0 {
		recommendations = append(recommendations, "Check dependency compatibility")
	}
	
	return recommendations
}

func (p *PackageUpdaterPlugin) loadUpdatePolicy(policyMap map[string]interface{}) {
	if autoUpdate, ok := policyMap["auto_update"].(bool); ok {
		p.updatePolicy.AutoUpdate = autoUpdate
	}
	
	if securityOnly, ok := policyMap["security_only"].(bool); ok {
		p.updatePolicy.SecurityOnly = securityOnly
	}
	
	if maxUpdates, ok := policyMap["max_updates_per_run"].(float64); ok {
		p.updatePolicy.MaxUpdatesPerRun = int(maxUpdates)
	}
}

func (p *PackageUpdaterPlugin) updateCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(6 * time.Hour) // Check every 6 hours
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.running {
				return
			}
			p.checkUpdates(ctx, nil)
		}
	}
}

// Plugin entry point
var Plugin PackageUpdaterPlugin

func init() {
	Plugin = PackageUpdaterPlugin{}
}