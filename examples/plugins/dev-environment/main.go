package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/plugins"
)

// DevEnvironmentPlugin manages development environments with AI assistance
type DevEnvironmentPlugin struct {
	config       plugins.PluginConfig
	running      bool
	mutex        sync.RWMutex
	environments map[string]*DevEnvironment
	templates    map[string]*EnvironmentTemplate
	lastScan     time.Time
}

type DevEnvironment struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`         // go, python, rust, nodejs, etc.
	Path         string                 `json:"path"`
	Status       string                 `json:"status"`       // active, inactive, broken
	Description  string                 `json:"description"`
	Languages    []Language             `json:"languages"`
	Tools        []Tool                 `json:"tools"`
	Services     []Service              `json:"services"`
	Dependencies []Dependency           `json:"dependencies"`
	Config       EnvironmentConfig      `json:"config"`
	Health       EnvironmentHealth      `json:"health"`
	Metrics      EnvironmentMetrics     `json:"metrics"`
	CreatedAt    time.Time              `json:"created_at"`
	LastUsed     time.Time              `json:"last_used"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type Language struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Compiler string `json:"compiler,omitempty"`
	Runtime  string `json:"runtime,omitempty"`
}

type Tool struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Purpose     string            `json:"purpose"`
	Config      map[string]string `json:"config,omitempty"`
	Essential   bool              `json:"essential"`
}

type Service struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`        // database, cache, queue, etc.
	Version     string            `json:"version"`
	Port        int               `json:"port,omitempty"`
	Config      map[string]string `json:"config"`
	Status      string            `json:"status"`      // running, stopped, error
	HealthCheck string            `json:"health_check,omitempty"`
}

type Dependency struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Type     string `json:"type"`     // system, language, library
	Optional bool   `json:"optional"`
	Purpose  string `json:"purpose"`
}

type EnvironmentConfig struct {
	ShellHook       string            `json:"shell_hook"`
	EnvVars         map[string]string `json:"env_vars"`
	Aliases         map[string]string `json:"aliases"`
	InitScripts     []string          `json:"init_scripts"`
	ConfigFiles     []ConfigFile      `json:"config_files"`
	ProjectSettings ProjectSettings   `json:"project_settings"`
}

type ConfigFile struct {
	Path     string `json:"path"`
	Template string `json:"template"`
	Content  string `json:"content"`
}

type ProjectSettings struct {
	BuildCommand   string   `json:"build_command"`
	TestCommand    string   `json:"test_command"`
	StartCommand   string   `json:"start_command"`
	DevScript      string   `json:"dev_script"`
	WatchPatterns  []string `json:"watch_patterns"`
	IgnorePatterns []string `json:"ignore_patterns"`
}

type EnvironmentHealth struct {
	Status       string           `json:"status"`       // healthy, degraded, unhealthy
	LastCheck    time.Time        `json:"last_check"`
	Issues       []HealthIssue    `json:"issues"`
	Score        int              `json:"score"`        // 0-100
	Checks       []HealthCheck    `json:"checks"`
	Suggestions  []string         `json:"suggestions"`
}

type HealthIssue struct {
	Type        string    `json:"type"`         // missing_dependency, outdated_tool, config_error
	Severity    string    `json:"severity"`     // low, medium, high, critical
	Component   string    `json:"component"`
	Message     string    `json:"message"`
	Resolution  string    `json:"resolution"`
	Timestamp   time.Time `json:"timestamp"`
}

type HealthCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`  // pass, fail, warning
	Message string `json:"message"`
}

type EnvironmentMetrics struct {
	Usage           UsageStats       `json:"usage"`
	Performance     PerformanceStats `json:"performance"`
	Dependencies    DependencyStats  `json:"dependencies"`
	ProjectActivity ActivityStats    `json:"project_activity"`
}

type UsageStats struct {
	TotalSessions    int           `json:"total_sessions"`
	LastSession      time.Time     `json:"last_session"`
	TotalHours       float64       `json:"total_hours"`
	AverageSession   time.Duration `json:"average_session"`
	WeeklyUsage      float64       `json:"weekly_usage"`
}

type PerformanceStats struct {
	ActivationTime   time.Duration `json:"activation_time"`
	BuildTime        time.Duration `json:"build_time"`
	TestTime         time.Duration `json:"test_time"`
	MemoryUsage      int64         `json:"memory_usage"`
	DiskUsage        int64         `json:"disk_usage"`
}

type DependencyStats struct {
	Total           int `json:"total"`
	Outdated        int `json:"outdated"`
	Vulnerable      int `json:"vulnerable"`
	Missing         int `json:"missing"`
}

type ActivityStats struct {
	LastCommit      time.Time `json:"last_commit"`
	FilesChanged    int       `json:"files_changed"`
	LinesAdded      int       `json:"lines_added"`
	LinesRemoved    int       `json:"lines_removed"`
	TestsRun        int       `json:"tests_run"`
	BuildsTriggered int       `json:"builds_triggered"`
}

type EnvironmentTemplate struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`     // web, mobile, data, ml, etc.
	Languages    []Language             `json:"languages"`
	Tools        []Tool                 `json:"tools"`
	Services     []Service              `json:"services"`
	Dependencies []Dependency           `json:"dependencies"`
	ConfigFiles  []ConfigFile           `json:"config_files"`
	InitCommands []string               `json:"init_commands"`
	Examples     []ProjectExample       `json:"examples"`
	Metadata     map[string]interface{} `json:"metadata"`
	Popular      bool                   `json:"popular"`
	Verified     bool                   `json:"verified"`
}

type ProjectExample struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	GitRepo     string `json:"git_repo,omitempty"`
	Commands    []string `json:"commands"`
}

// Metadata methods
func (p *DevEnvironmentPlugin) Name() string        { return "dev-environment" }
func (p *DevEnvironmentPlugin) Version() string     { return "1.0.0" }
func (p *DevEnvironmentPlugin) Description() string { return "AI-powered development environment management and setup" }
func (p *DevEnvironmentPlugin) Author() string      { return "NixAI Team" }
func (p *DevEnvironmentPlugin) Repository() string  { return "https://github.com/nixai/plugins/dev-environment" }
func (p *DevEnvironmentPlugin) License() string     { return "MIT" }

func (p *DevEnvironmentPlugin) Dependencies() []string {
	return []string{"nix", "direnv", "git"}
}

func (p *DevEnvironmentPlugin) Capabilities() []string {
	return []string{
		"environment-management",
		"template-generation",
		"dependency-resolution",
		"project-analysis",
		"automated-setup",
		"health-monitoring",
	}
}

// Lifecycle methods
func (p *DevEnvironmentPlugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.config = config
	p.environments = make(map[string]*DevEnvironment)
	p.templates = make(map[string]*EnvironmentTemplate)
	
	// Load built-in templates
	p.loadBuiltinTemplates()
	
	return nil
}

func (p *DevEnvironmentPlugin) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.running {
		return fmt.Errorf("plugin is already running")
	}
	
	// Start monitoring goroutine
	go p.monitoringLoop(ctx)
	
	p.running = true
	return nil
}

func (p *DevEnvironmentPlugin) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if !p.running {
		return fmt.Errorf("plugin is not running")
	}
	
	p.running = false
	return nil
}

func (p *DevEnvironmentPlugin) Cleanup(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.environments = nil
	p.templates = nil
	
	return nil
}

func (p *DevEnvironmentPlugin) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

// Execution methods
func (p *DevEnvironmentPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "list-environments":
		return p.listEnvironments(ctx, params)
	case "create-environment":
		return p.createEnvironment(ctx, params)
	case "delete-environment":
		return p.deleteEnvironment(ctx, params)
	case "activate-environment":
		return p.activateEnvironment(ctx, params)
	case "deactivate-environment":
		return p.deactivateEnvironment(ctx, params)
	case "list-templates":
		return p.listTemplates(ctx, params)
	case "create-from-template":
		return p.createFromTemplate(ctx, params)
	case "analyze-project":
		return p.analyzeProject(ctx, params)
	case "suggest-environment":
		return p.suggestEnvironment(ctx, params)
	case "health-check":
		return p.healthCheck(ctx, params)
	case "update-dependencies":
		return p.updateDependencies(ctx, params)
	case "backup-environment":
		return p.backupEnvironment(ctx, params)
	case "restore-environment":
		return p.restoreEnvironment(ctx, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *DevEnvironmentPlugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "list-environments",
			Description: "List all development environments",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "type",
					Type:        "string",
					Description: "Filter by environment type",
					Required:    false,
				},
				{
					Name:        "status",
					Type:        "string",
					Description: "Filter by environment status",
					Required:    false,
				},
			},
			ReturnType: "[]DevEnvironment",
			Tags:       []string{"environments", "list"},
		},
		{
			Name:        "create-environment",
			Description: "Create a new development environment",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "name",
					Type:        "string",
					Description: "Environment name",
					Required:    true,
				},
				{
					Name:        "type",
					Type:        "string",
					Description: "Environment type (go, python, nodejs, etc.)",
					Required:    true,
				},
				{
					Name:        "path",
					Type:        "string",
					Description: "Project path",
					Required:    true,
				},
				{
					Name:        "template",
					Type:        "string",
					Description: "Template to use (optional)",
					Required:    false,
				},
			},
			ReturnType: "DevEnvironment",
			Tags:       []string{"environments", "create"},
		},
		{
			Name:        "list-templates",
			Description: "List available environment templates",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "category",
					Type:        "string",
					Description: "Filter by template category",
					Required:    false,
				},
			},
			ReturnType: "[]EnvironmentTemplate",
			Tags:       []string{"templates", "list"},
		},
		{
			Name:        "create-from-template",
			Description: "Create environment from template",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "template_id",
					Type:        "string",
					Description: "Template ID to use",
					Required:    true,
				},
				{
					Name:        "name",
					Type:        "string",
					Description: "Environment name",
					Required:    true,
				},
				{
					Name:        "path",
					Type:        "string",
					Description: "Project path",
					Required:    true,
				},
			},
			ReturnType: "DevEnvironment",
			Tags:       []string{"environments", "templates"},
		},
		{
			Name:        "analyze-project",
			Description: "Analyze existing project and suggest environment",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "path",
					Type:        "string",
					Description: "Project path to analyze",
					Required:    true,
				},
			},
			ReturnType: "ProjectAnalysis",
			Tags:       []string{"analysis", "ai"},
		},
		{
			Name:        "health-check",
			Description: "Check environment health and dependencies",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "environment_id",
					Type:        "string",
					Description: "Environment ID to check",
					Required:    false,
				},
			},
			ReturnType: "HealthReport",
			Tags:       []string{"health", "diagnostics"},
		},
	}
}

func (p *DevEnvironmentPlugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	schemas := map[string]*plugins.PluginSchema{
		"create-environment": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"name": {
					Type:        "string",
					Description: "Environment name",
				},
				"type": {
					Type:        "string",
					Description: "Environment type",
					Enum:        []string{"go", "python", "nodejs", "rust", "java", "cpp"},
				},
				"path": {
					Type:        "string",
					Description: "Project path",
				},
			},
			Required: []string{"name", "type", "path"},
		},
	}
	
	if schema, exists := schemas[operation]; exists {
		return schema, nil
	}
	
	return nil, fmt.Errorf("unknown operation: %s", operation)
}

// Health and Status methods
func (p *DevEnvironmentPlugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	status := plugins.HealthHealthy
	message := "Development environment manager running normally"
	var issues []plugins.HealthIssue
	
	if !p.running {
		status = plugins.HealthUnhealthy
		message = "Development environment manager is not running"
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityError,
			Component: "manager",
			Message:   "Manager service stopped",
			Timestamp: time.Now(),
		})
	}
	
	// Check for broken environments
	brokenCount := 0
	for _, env := range p.environments {
		if env.Status == "broken" {
			brokenCount++
		}
	}
	
	if brokenCount > 0 {
		status = plugins.HealthDegraded
		message = fmt.Sprintf("%d environments need attention", brokenCount)
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityWarning,
			Component: "environments",
			Message:   fmt.Sprintf("%d broken environments", brokenCount),
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

func (p *DevEnvironmentPlugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{
		ExecutionCount:       0,
		TotalExecutionTime:   0,
		AverageExecutionTime: 0,
		LastExecutionTime:    p.lastScan,
		ErrorCount:           0,
		SuccessRate:          100.0,
		StartTime:            time.Now(),
		CustomMetrics: map[string]interface{}{
			"environments_count": len(p.environments),
			"templates_count":    len(p.templates),
			"last_scan":          p.lastScan,
			"active_environments": p.countActiveEnvironments(),
		},
	}
}

func (p *DevEnvironmentPlugin) GetStatus() plugins.PluginStatus {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	state := plugins.StateRunning
	if !p.running {
		state = plugins.StateStopped
	}
	
	return plugins.PluginStatus{
		State:       state,
		Message:     "Development environment manager active",
		LastUpdated: time.Now(),
		Version:     p.Version(),
	}
}

// Operation implementations
func (p *DevEnvironmentPlugin) listEnvironments(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	var environments []DevEnvironment
	
	// Apply filters
	envType, hasType := params["type"].(string)
	status, hasStatus := params["status"].(string)
	
	for _, env := range p.environments {
		// Type filter
		if hasType && env.Type != envType {
			continue
		}
		
		// Status filter
		if hasStatus && env.Status != status {
			continue
		}
		
		environments = append(environments, *env)
	}
	
	return environments, nil
}

func (p *DevEnvironmentPlugin) createEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter required")
	}
	
	envType, ok := params["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type parameter required")
	}
	
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter required")
	}
	
	// Check if environment already exists
	p.mutex.RLock()
	for _, env := range p.environments {
		if env.Name == name {
			p.mutex.RUnlock()
			return nil, fmt.Errorf("environment with name '%s' already exists", name)
		}
	}
	p.mutex.RUnlock()
	
	// Create environment
	env := &DevEnvironment{
		ID:          generateID(),
		Name:        name,
		Type:        envType,
		Path:        path,
		Status:      "inactive",
		Description: fmt.Sprintf("%s development environment", envType),
		CreatedAt:   time.Now(),
		Config: EnvironmentConfig{
			EnvVars: make(map[string]string),
			Aliases: make(map[string]string),
		},
		Metadata: make(map[string]interface{}),
	}
	
	// Set up environment based on type
	err := p.setupEnvironmentByType(env, envType)
	if err != nil {
		return nil, fmt.Errorf("failed to setup environment: %v", err)
	}
	
	// Create necessary files and configuration
	err = p.createEnvironmentFiles(env)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment files: %v", err)
	}
	
	p.mutex.Lock()
	p.environments[env.ID] = env
	p.mutex.Unlock()
	
	return env, nil
}

func (p *DevEnvironmentPlugin) deleteEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	envID, ok := params["environment_id"].(string)
	if !ok {
		return nil, fmt.Errorf("environment_id parameter required")
	}
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	env, exists := p.environments[envID]
	if !exists {
		return nil, fmt.Errorf("environment not found")
	}
	
	// Clean up environment files
	err := p.cleanupEnvironmentFiles(env)
	if err != nil {
		return false, fmt.Errorf("failed to cleanup environment: %v", err)
	}
	
	delete(p.environments, envID)
	
	return true, nil
}

func (p *DevEnvironmentPlugin) activateEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	envID, ok := params["environment_id"].(string)
	if !ok {
		return nil, fmt.Errorf("environment_id parameter required")
	}
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	env, exists := p.environments[envID]
	if !exists {
		return nil, fmt.Errorf("environment not found")
	}
	
	// Generate activation script
	script := p.generateActivationScript(env)
	
	env.Status = "active"
	env.LastUsed = time.Now()
	
	return map[string]interface{}{
		"environment_id": envID,
		"status":         "activated",
		"script":         script,
		"instructions":   p.getActivationInstructions(env),
	}, nil
}

func (p *DevEnvironmentPlugin) deactivateEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	envID, ok := params["environment_id"].(string)
	if !ok {
		return nil, fmt.Errorf("environment_id parameter required")
	}
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	env, exists := p.environments[envID]
	if !exists {
		return nil, fmt.Errorf("environment not found")
	}
	
	env.Status = "inactive"
	
	return map[string]interface{}{
		"environment_id": envID,
		"status":         "deactivated",
	}, nil
}

func (p *DevEnvironmentPlugin) listTemplates(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	var templates []EnvironmentTemplate
	
	category, hasCategory := params["category"].(string)
	
	for _, template := range p.templates {
		if hasCategory && template.Category != category {
			continue
		}
		
		templates = append(templates, *template)
	}
	
	return templates, nil
}

func (p *DevEnvironmentPlugin) createFromTemplate(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	templateID, ok := params["template_id"].(string)
	if !ok {
		return nil, fmt.Errorf("template_id parameter required")
	}
	
	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name parameter required")
	}
	
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter required")
	}
	
	p.mutex.RLock()
	template, exists := p.templates[templateID]
	p.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("template not found")
	}
	
	// Create environment from template
	env := &DevEnvironment{
		ID:           generateID(),
		Name:         name,
		Type:         template.Category,
		Path:         path,
		Status:       "inactive",
		Description:  template.Description,
		Languages:    template.Languages,
		Tools:        template.Tools,
		Services:     template.Services,
		Dependencies: template.Dependencies,
		CreatedAt:    time.Now(),
		Config: EnvironmentConfig{
			EnvVars:     make(map[string]string),
			Aliases:     make(map[string]string),
			ConfigFiles: template.ConfigFiles,
		},
		Metadata: make(map[string]interface{}),
	}
	
	// Execute template initialization
	err := p.executeTemplateInit(env, template)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize from template: %v", err)
	}
	
	p.mutex.Lock()
	p.environments[env.ID] = env
	p.mutex.Unlock()
	
	return env, nil
}

func (p *DevEnvironmentPlugin) analyzeProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	projectPath, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter required")
	}
	
	analysis := p.performProjectAnalysis(projectPath)
	
	return analysis, nil
}

func (p *DevEnvironmentPlugin) suggestEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	projectPath, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter required")
	}
	
	analysis := p.performProjectAnalysis(projectPath)
	suggestions := p.generateEnvironmentSuggestions(analysis)
	
	return map[string]interface{}{
		"analysis":    analysis,
		"suggestions": suggestions,
	}, nil
}

func (p *DevEnvironmentPlugin) healthCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if envID, ok := params["environment_id"].(string); ok {
		return p.checkSingleEnvironment(envID)
	}
	
	// Check all environments
	return p.checkAllEnvironments()
}

func (p *DevEnvironmentPlugin) updateDependencies(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	envID, ok := params["environment_id"].(string)
	if !ok {
		return nil, fmt.Errorf("environment_id parameter required")
	}
	
	p.mutex.RLock()
	env, exists := p.environments[envID]
	p.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("environment not found")
	}
	
	// Update dependencies based on environment type
	results := p.updateEnvironmentDependencies(env)
	
	return results, nil
}

func (p *DevEnvironmentPlugin) backupEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	envID, ok := params["environment_id"].(string)
	if !ok {
		return nil, fmt.Errorf("environment_id parameter required")
	}
	
	p.mutex.RLock()
	env, exists := p.environments[envID]
	p.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("environment not found")
	}
	
	backupPath := p.createEnvironmentBackup(env)
	
	return map[string]interface{}{
		"environment_id": envID,
		"backup_path":    backupPath,
		"timestamp":      time.Now(),
	}, nil
}

func (p *DevEnvironmentPlugin) restoreEnvironment(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	backupPath, ok := params["backup_path"].(string)
	if !ok {
		return nil, fmt.Errorf("backup_path parameter required")
	}
	
	env, err := p.restoreFromBackup(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to restore environment: %v", err)
	}
	
	p.mutex.Lock()
	p.environments[env.ID] = env
	p.mutex.Unlock()
	
	return env, nil
}

// Helper methods
func (p *DevEnvironmentPlugin) loadBuiltinTemplates() {
	// Go template
	p.templates["go-basic"] = &EnvironmentTemplate{
		ID:          "go-basic",
		Name:        "Go Basic",
		Description: "Basic Go development environment",
		Category:    "go",
		Languages: []Language{
			{Name: "go", Version: "latest"},
		},
		Tools: []Tool{
			{Name: "gopls", Purpose: "Language server", Essential: true},
			{Name: "gofmt", Purpose: "Code formatter", Essential: true},
			{Name: "golint", Purpose: "Linter", Essential: false},
		},
		Dependencies: []Dependency{
			{Name: "git", Type: "system", Purpose: "Version control"},
		},
		Popular:  true,
		Verified: true,
	}
	
	// Python template
	p.templates["python-basic"] = &EnvironmentTemplate{
		ID:          "python-basic",
		Name:        "Python Basic",
		Description: "Basic Python development environment",
		Category:    "python",
		Languages: []Language{
			{Name: "python", Version: "3.11"},
		},
		Tools: []Tool{
			{Name: "pip", Purpose: "Package manager", Essential: true},
			{Name: "pylsp", Purpose: "Language server", Essential: true},
			{Name: "black", Purpose: "Code formatter", Essential: false},
		},
		Dependencies: []Dependency{
			{Name: "python311", Type: "system", Purpose: "Python interpreter"},
		},
		Popular:  true,
		Verified: true,
	}
	
	// Node.js template
	p.templates["nodejs-basic"] = &EnvironmentTemplate{
		ID:          "nodejs-basic",
		Name:        "Node.js Basic",
		Description: "Basic Node.js development environment",
		Category:    "nodejs",
		Languages: []Language{
			{Name: "nodejs", Version: "18"},
		},
		Tools: []Tool{
			{Name: "npm", Purpose: "Package manager", Essential: true},
			{Name: "typescript", Purpose: "TypeScript support", Essential: false},
		},
		Dependencies: []Dependency{
			{Name: "nodejs_18", Type: "system", Purpose: "Node.js runtime"},
		},
		Popular:  true,
		Verified: true,
	}
}

func (p *DevEnvironmentPlugin) setupEnvironmentByType(env *DevEnvironment, envType string) error {
	switch envType {
	case "go":
		env.Languages = []Language{{Name: "go", Version: "latest"}}
		env.Tools = []Tool{
			{Name: "gopls", Purpose: "Language server", Essential: true},
			{Name: "gofmt", Purpose: "Code formatter", Essential: true},
		}
		env.Config.EnvVars["GOPATH"] = filepath.Join(env.Path, ".go")
		env.Config.ProjectSettings.BuildCommand = "go build"
		env.Config.ProjectSettings.TestCommand = "go test ./..."
		
	case "python":
		env.Languages = []Language{{Name: "python", Version: "3.11"}}
		env.Tools = []Tool{
			{Name: "pip", Purpose: "Package manager", Essential: true},
			{Name: "pylsp", Purpose: "Language server", Essential: true},
		}
		env.Config.EnvVars["PYTHONPATH"] = env.Path
		env.Config.ProjectSettings.TestCommand = "pytest"
		
	case "nodejs":
		env.Languages = []Language{{Name: "nodejs", Version: "18"}}
		env.Tools = []Tool{
			{Name: "npm", Purpose: "Package manager", Essential: true},
		}
		env.Config.ProjectSettings.BuildCommand = "npm run build"
		env.Config.ProjectSettings.TestCommand = "npm test"
		env.Config.ProjectSettings.StartCommand = "npm start"
	}
	
	return nil
}

func (p *DevEnvironmentPlugin) createEnvironmentFiles(env *DevEnvironment) error {
	// Create .envrc file for direnv
	envrcPath := filepath.Join(env.Path, ".envrc")
	envrcContent := p.generateEnvrcContent(env)
	
	err := os.WriteFile(envrcPath, []byte(envrcContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create .envrc: %v", err)
	}
	
	// Create shell.nix file
	shellNixPath := filepath.Join(env.Path, "shell.nix")
	shellNixContent := p.generateShellNixContent(env)
	
	err = os.WriteFile(shellNixPath, []byte(shellNixContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create shell.nix: %v", err)
	}
	
	return nil
}

func (p *DevEnvironmentPlugin) generateEnvrcContent(env *DevEnvironment) string {
	content := "use nix\n\n"
	
	// Add environment variables
	for key, value := range env.Config.EnvVars {
		content += fmt.Sprintf("export %s=\"%s\"\n", key, value)
	}
	
	// Add aliases
	for alias, command := range env.Config.Aliases {
		content += fmt.Sprintf("alias %s=\"%s\"\n", alias, command)
	}
	
	return content
}

func (p *DevEnvironmentPlugin) generateShellNixContent(env *DevEnvironment) string {
	content := "{ pkgs ? import <nixpkgs> {} }:\n\n"
	content += "pkgs.mkShell {\n"
	content += "  buildInputs = with pkgs; [\n"
	
	// Add language packages
	for _, lang := range env.Languages {
		switch lang.Name {
		case "go":
			content += "    go\n"
		case "python":
			content += "    python311\n"
		case "nodejs":
			content += "    nodejs_18\n"
		}
	}
	
	// Add tools
	for _, tool := range env.Tools {
		if tool.Essential {
			content += fmt.Sprintf("    %s\n", tool.Name)
		}
	}
	
	content += "  ];\n"
	
	// Add shell hook
	if env.Config.ShellHook != "" {
		content += fmt.Sprintf("  shellHook = ''%s'';\n", env.Config.ShellHook)
	}
	
	content += "}\n"
	
	return content
}

func (p *DevEnvironmentPlugin) cleanupEnvironmentFiles(env *DevEnvironment) error {
	files := []string{
		filepath.Join(env.Path, ".envrc"),
		filepath.Join(env.Path, "shell.nix"),
	}
	
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	
	return nil
}

func (p *DevEnvironmentPlugin) generateActivationScript(env *DevEnvironment) string {
	script := "#!/bin/bash\n\n"
	script += "# Activate development environment\n"
	script += fmt.Sprintf("cd \"%s\"\n", env.Path)
	script += "direnv allow\n"
	script += "nix-shell\n"
	
	return script
}

func (p *DevEnvironmentPlugin) getActivationInstructions(env *DevEnvironment) []string {
	return []string{
		fmt.Sprintf("Navigate to project directory: cd %s", env.Path),
		"Allow direnv: direnv allow",
		"Enter nix shell: nix-shell",
		"Your development environment is now active!",
	}
}

func (p *DevEnvironmentPlugin) executeTemplateInit(env *DevEnvironment, template *EnvironmentTemplate) error {
	// Create config files from template
	for _, configFile := range template.ConfigFiles {
		filePath := filepath.Join(env.Path, configFile.Path)
		
		// Create directory if needed
		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
		
		// Write file content
		err = os.WriteFile(filePath, []byte(configFile.Content), 0644)
		if err != nil {
			return err
		}
	}
	
	// Execute init commands
	for _, command := range template.InitCommands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Dir = env.Path
		_, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("init command failed: %s: %v", command, err)
		}
	}
	
	return nil
}

func (p *DevEnvironmentPlugin) performProjectAnalysis(projectPath string) map[string]interface{} {
	analysis := map[string]interface{}{
		"path":              projectPath,
		"detected_languages": []string{},
		"detected_tools":     []string{},
		"config_files":       []string{},
		"package_files":      []string{},
		"recommendations":    []string{},
	}
	
	// Check for common files and patterns
	files, err := os.ReadDir(projectPath)
	if err != nil {
		return analysis
	}
	
	languages := []string{}
	tools := []string{}
	configFiles := []string{}
	packageFiles := []string{}
	
	for _, file := range files {
		name := file.Name()
		
		// Language detection
		switch {
		case strings.HasSuffix(name, ".go") || name == "go.mod":
			if !contains(languages, "go") {
				languages = append(languages, "go")
			}
		case strings.HasSuffix(name, ".py") || name == "requirements.txt" || name == "setup.py":
			if !contains(languages, "python") {
				languages = append(languages, "python")
			}
		case strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".ts") || name == "package.json":
			if !contains(languages, "nodejs") {
				languages = append(languages, "nodejs")
			}
		case strings.HasSuffix(name, ".rs") || name == "Cargo.toml":
			if !contains(languages, "rust") {
				languages = append(languages, "rust")
			}
		}
		
		// Package files
		switch name {
		case "package.json", "package-lock.json", "yarn.lock":
			packageFiles = append(packageFiles, name)
		case "requirements.txt", "Pipfile", "pyproject.toml":
			packageFiles = append(packageFiles, name)
		case "go.mod", "go.sum":
			packageFiles = append(packageFiles, name)
		case "Cargo.toml", "Cargo.lock":
			packageFiles = append(packageFiles, name)
		}
		
		// Config files
		switch name {
		case ".envrc", "shell.nix", "flake.nix":
			configFiles = append(configFiles, name)
		case "Dockerfile", "docker-compose.yml":
			configFiles = append(configFiles, name)
		}
		
		// Tools detection
		switch name {
		case "Makefile":
			tools = append(tools, "make")
		case "Dockerfile":
			tools = append(tools, "docker")
		}
	}
	
	analysis["detected_languages"] = languages
	analysis["detected_tools"] = tools
	analysis["config_files"] = configFiles
	analysis["package_files"] = packageFiles
	
	return analysis
}

func (p *DevEnvironmentPlugin) generateEnvironmentSuggestions(analysis map[string]interface{}) []map[string]interface{} {
	var suggestions []map[string]interface{}
	
	languages, _ := analysis["detected_languages"].([]string)
	
	for _, lang := range languages {
		if template, exists := p.templates[lang+"-basic"]; exists {
			suggestion := map[string]interface{}{
				"template_id":   template.ID,
				"template_name": template.Name,
				"description":   template.Description,
				"confidence":    90,
				"reasons": []string{
					fmt.Sprintf("Detected %s files in project", lang),
					"Template matches project structure",
				},
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions
}

func (p *DevEnvironmentPlugin) checkSingleEnvironment(envID string) (interface{}, error) {
	p.mutex.RLock()
	env, exists := p.environments[envID]
	p.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("environment not found")
	}
	
	health := p.assessEnvironmentHealth(env)
	
	return health, nil
}

func (p *DevEnvironmentPlugin) checkAllEnvironments() (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	totalEnvs := len(p.environments)
	healthyEnvs := 0
	brokenEnvs := 0
	
	var issues []string
	var recommendations []string
	
	for _, env := range p.environments {
		health := p.assessEnvironmentHealth(env)
		
		if health.Status == "healthy" {
			healthyEnvs++
		} else if health.Status == "unhealthy" {
			brokenEnvs++
			issues = append(issues, fmt.Sprintf("Environment %s is unhealthy", env.Name))
		}
	}
	
	overallStatus := "healthy"
	if brokenEnvs > 0 {
		overallStatus = "degraded"
	}
	if brokenEnvs > totalEnvs/2 {
		overallStatus = "unhealthy"
	}
	
	return map[string]interface{}{
		"overall_status":      overallStatus,
		"total_environments":  totalEnvs,
		"healthy_environments": healthyEnvs,
		"broken_environments": brokenEnvs,
		"issues":              issues,
		"recommendations":     recommendations,
		"last_check":          time.Now(),
	}, nil
}

func (p *DevEnvironmentPlugin) assessEnvironmentHealth(env *DevEnvironment) EnvironmentHealth {
	score := 100
	var issues []HealthIssue
	var checks []HealthCheck
	var suggestions []string
	
	// Check if path exists
	if _, err := os.Stat(env.Path); os.IsNotExist(err) {
		score -= 50
		issues = append(issues, HealthIssue{
			Type:       "missing_path",
			Severity:   "critical",
			Component:  "filesystem",
			Message:    "Project path does not exist",
			Resolution: "Recreate project directory or update path",
			Timestamp:  time.Now(),
		})
		checks = append(checks, HealthCheck{
			Name:    "Path Exists",
			Status:  "fail",
			Message: "Project path not found",
		})
	} else {
		checks = append(checks, HealthCheck{
			Name:    "Path Exists",
			Status:  "pass",
			Message: "Project path is accessible",
		})
	}
	
	// Check for required files
	requiredFiles := []string{".envrc", "shell.nix"}
	for _, file := range requiredFiles {
		filePath := filepath.Join(env.Path, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			score -= 10
			checks = append(checks, HealthCheck{
				Name:    fmt.Sprintf("File %s", file),
				Status:  "fail",
				Message: fmt.Sprintf("Required file %s is missing", file),
			})
			suggestions = append(suggestions, fmt.Sprintf("Create missing file: %s", file))
		} else {
			checks = append(checks, HealthCheck{
				Name:    fmt.Sprintf("File %s", file),
				Status:  "pass",
				Message: fmt.Sprintf("File %s exists", file),
			})
		}
	}
	
	// Check tool availability
	for _, tool := range env.Tools {
		if tool.Essential {
			cmd := exec.Command("which", tool.Name)
			err := cmd.Run()
			if err != nil {
				score -= 15
				issues = append(issues, HealthIssue{
					Type:       "missing_dependency",
					Severity:   "high",
					Component:  "tools",
					Message:    fmt.Sprintf("Essential tool %s not found", tool.Name),
					Resolution: fmt.Sprintf("Install %s", tool.Name),
					Timestamp:  time.Now(),
				})
				checks = append(checks, HealthCheck{
					Name:    fmt.Sprintf("Tool %s", tool.Name),
					Status:  "fail",
					Message: fmt.Sprintf("Tool %s not available", tool.Name),
				})
			} else {
				checks = append(checks, HealthCheck{
					Name:    fmt.Sprintf("Tool %s", tool.Name),
					Status:  "pass",
					Message: fmt.Sprintf("Tool %s is available", tool.Name),
				})
			}
		}
	}
	
	status := "healthy"
	if score < 70 {
		status = "degraded"
	}
	if score < 40 {
		status = "unhealthy"
	}
	
	return EnvironmentHealth{
		Status:      status,
		LastCheck:   time.Now(),
		Issues:      issues,
		Score:       score,
		Checks:      checks,
		Suggestions: suggestions,
	}
}

func (p *DevEnvironmentPlugin) updateEnvironmentDependencies(env *DevEnvironment) map[string]interface{} {
	results := map[string]interface{}{
		"environment_id": env.ID,
		"updates":        []string{},
		"errors":         []string{},
		"status":         "success",
	}
	
	// Update based on environment type
	switch env.Type {
	case "nodejs":
		cmd := exec.Command("npm", "update")
		cmd.Dir = env.Path
		output, err := cmd.CombinedOutput()
		if err != nil {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("npm update failed: %v", err))
			results["status"] = "partial"
		} else {
			results["updates"] = append(results["updates"].([]string), "Updated npm packages")
		}
		results["output"] = string(output)
		
	case "python":
		cmd := exec.Command("pip", "list", "--outdated")
		cmd.Dir = env.Path
		output, err := cmd.CombinedOutput()
		if err == nil {
			results["outdated_packages"] = string(output)
		}
		
	case "go":
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = env.Path
		output, err := cmd.CombinedOutput()
		if err != nil {
			results["errors"] = append(results["errors"].([]string), fmt.Sprintf("go mod tidy failed: %v", err))
			results["status"] = "partial"
		} else {
			results["updates"] = append(results["updates"].([]string), "Tidied go modules")
		}
		results["output"] = string(output)
	}
	
	return results
}

func (p *DevEnvironmentPlugin) createEnvironmentBackup(env *DevEnvironment) string {
	backupDir := filepath.Join(os.TempDir(), "nixai-env-backups")
	os.MkdirAll(backupDir, 0755)
	
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s-%d.json", env.Name, time.Now().Unix()))
	
	data, _ := json.MarshalIndent(env, "", "  ")
	os.WriteFile(backupPath, data, 0644)
	
	return backupPath
}

func (p *DevEnvironmentPlugin) restoreFromBackup(backupPath string) (*DevEnvironment, error) {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, err
	}
	
	var env DevEnvironment
	err = json.Unmarshal(data, &env)
	if err != nil {
		return nil, err
	}
	
	// Generate new ID for restored environment
	env.ID = generateID()
	
	return &env, nil
}

func (p *DevEnvironmentPlugin) countActiveEnvironments() int {
	count := 0
	for _, env := range p.environments {
		if env.Status == "active" {
			count++
		}
	}
	return count
}

func (p *DevEnvironmentPlugin) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Monitor every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.running {
				return
			}
			p.lastScan = time.Now()
		}
	}
}

// Utility functions
func generateID() string {
	return fmt.Sprintf("env-%d", time.Now().Unix())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Plugin entry point
var Plugin DevEnvironmentPlugin

func init() {
	Plugin = DevEnvironmentPlugin{}
}