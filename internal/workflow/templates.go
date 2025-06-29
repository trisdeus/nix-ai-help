package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateManager manages workflow templates
type TemplateManager struct {
	templates map[string]*WorkflowTemplate
	logger    Logger
}

// WorkflowTemplate represents a workflow template
type WorkflowTemplate struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Category    string                 `yaml:"category"`
	Tags        []string               `yaml:"tags"`
	Parameters  []TemplateParameter    `yaml:"parameters"`
	Workflow    *WorkflowDefinition    `yaml:"workflow"`
	Variables   map[string]interface{} `yaml:"variables"`
}

// TemplateParameter represents a template parameter
type TemplateParameter struct {
	Name         string      `yaml:"name"`
	Description  string      `yaml:"description"`
	Type         string      `yaml:"type"` // string, int, bool, choice
	Required     bool        `yaml:"required"`
	DefaultValue interface{} `yaml:"default"`
	Choices      []string    `yaml:"choices,omitempty"`
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(logger Logger) *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*WorkflowTemplate),
		logger:    logger,
	}

	// Initialize built-in templates
	tm.initializeBuiltinTemplates()

	return tm
}

// GetTemplate retrieves a template by ID
func (tm *TemplateManager) GetTemplate(id string) (*WorkflowTemplate, error) {
	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// ListTemplates returns all available templates
func (tm *TemplateManager) ListTemplates() []*WorkflowTemplate {
	templates := make([]*WorkflowTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// CreateWorkflowFromTemplate creates a workflow from a template with parameters
func (tm *TemplateManager) CreateWorkflowFromTemplate(templateID string, parameters map[string]interface{}) (*WorkflowDefinition, error) {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	// Validate parameters
	if err := tm.validateParameters(template, parameters); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Clone template workflow
	workflow := tm.cloneWorkflowDefinition(template.Workflow)

	// Apply parameters
	if err := tm.applyParameters(workflow, template, parameters); err != nil {
		return nil, fmt.Errorf("failed to apply parameters: %w", err)
	}

	return workflow, nil
}

// validateParameters validates template parameters
func (tm *TemplateManager) validateParameters(template *WorkflowTemplate, parameters map[string]interface{}) error {
	for _, param := range template.Parameters {
		value, exists := parameters[param.Name]

		// Check required parameters
		if param.Required && !exists {
			return fmt.Errorf("required parameter missing: %s", param.Name)
		}

		// Use default value if not provided
		if !exists && param.DefaultValue != nil {
			parameters[param.Name] = param.DefaultValue
			continue
		}

		if !exists {
			continue
		}

		// Validate parameter type
		if err := tm.validateParameterType(param, value); err != nil {
			return fmt.Errorf("parameter %s: %w", param.Name, err)
		}

		// Validate choices
		if len(param.Choices) > 0 {
			if err := tm.validateParameterChoice(param, value); err != nil {
				return fmt.Errorf("parameter %s: %w", param.Name, err)
			}
		}
	}

	return nil
}

// validateParameterType validates parameter type
func (tm *TemplateManager) validateParameterType(param TemplateParameter, value interface{}) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "int":
		switch value.(type) {
		case int, int32, int64, float64, float32:
			// OK
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "choice":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string choice, got %T", value)
		}
	}
	return nil
}

// validateParameterChoice validates parameter choice
func (tm *TemplateManager) validateParameterChoice(param TemplateParameter, value interface{}) error {
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("choice parameter must be string")
	}

	for _, choice := range param.Choices {
		if choice == strValue {
			return nil
		}
	}

	return fmt.Errorf("invalid choice '%s', valid options: %v", strValue, param.Choices)
}

// applyParameters applies parameters to workflow definition
func (tm *TemplateManager) applyParameters(workflow *WorkflowDefinition, template *WorkflowTemplate, parameters map[string]interface{}) error {
	// Merge template variables with parameters
	if workflow.Variables == nil {
		workflow.Variables = make(map[string]interface{})
	}

	// Add template variables
	for key, value := range template.Variables {
		workflow.Variables[key] = value
	}

	// Add parameters as variables
	for key, value := range parameters {
		workflow.Variables[key] = value
	}

	// Replace parameter placeholders in workflow definition
	if err := tm.replacePlaceholders(workflow, parameters); err != nil {
		return err
	}

	return nil
}

// replacePlaceholders replaces parameter placeholders in workflow
func (tm *TemplateManager) replacePlaceholders(workflow *WorkflowDefinition, parameters map[string]interface{}) error {
	// Replace in workflow name and description
	workflow.Name = tm.replacePlaceholdersInString(workflow.Name, parameters)
	workflow.Description = tm.replacePlaceholdersInString(workflow.Description, parameters)

	// Replace in tasks
	for i := range workflow.Tasks {
		task := &workflow.Tasks[i]
		task.Name = tm.replacePlaceholdersInString(task.Name, parameters)
		task.Description = tm.replacePlaceholdersInString(task.Description, parameters)

		// Replace in action parameters
		if cmdParam, exists := task.Action.Parameters["command"]; exists {
			if cmdStr, ok := cmdParam.(string); ok {
				task.Action.Parameters["command"] = tm.replacePlaceholdersInString(cmdStr, parameters)
			}
		}

		// Replace in action parameters map
		for key, value := range task.Action.Parameters {
			if strValue, ok := value.(string); ok {
				task.Action.Parameters[key] = tm.replacePlaceholdersInString(strValue, parameters)
			}
		}
	}

	return nil
}

// replacePlaceholdersInString replaces parameter placeholders in a string
func (tm *TemplateManager) replacePlaceholdersInString(text string, parameters map[string]interface{}) string {
	result := text
	for key, value := range parameters {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// cloneWorkflowDefinition creates a deep copy of a workflow definition
func (tm *TemplateManager) cloneWorkflowDefinition(workflow *WorkflowDefinition) *WorkflowDefinition {
	if workflow == nil {
		return &WorkflowDefinition{
			Variables: make(map[string]interface{}),
		}
	}

	clone := &WorkflowDefinition{
		Name:        workflow.Name,
		Description: workflow.Description,
		Version:     workflow.Version,
		Author:      workflow.Author,
		Tags:        make([]string, len(workflow.Tags)),
		Metadata:    make(map[string]string),
		Config:      workflow.Config,
		Tasks:       make([]TaskDefinition, len(workflow.Tasks)),
		Triggers:    make([]TriggerDefinition, len(workflow.Triggers)),
		Variables:   make(map[string]interface{}),
	}

	// Copy tags
	copy(clone.Tags, workflow.Tags)

	// Copy metadata
	for key, value := range workflow.Metadata {
		clone.Metadata[key] = value
	}

	// Copy tasks
	copy(clone.Tasks, workflow.Tasks)

	// Copy triggers
	copy(clone.Triggers, workflow.Triggers)

	// Copy variables
	for key, value := range workflow.Variables {
		clone.Variables[key] = value
	}

	return clone
}

// RegisterTemplate registers a new template
func (tm *TemplateManager) RegisterTemplate(template *WorkflowTemplate) error {
	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	if _, exists := tm.templates[template.ID]; exists {
		return fmt.Errorf("template with ID %s already exists", template.ID)
	}

	tm.templates[template.ID] = template
	tm.logger.Info("Registered workflow template: %s", template.ID)

	return nil
}

// LoadTemplatesFromDirectory loads templates from a directory
func (tm *TemplateManager) LoadTemplatesFromDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		tm.logger.Debug("Template directory does not exist: %s", dir)
		return nil
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		template, err := tm.loadTemplateFromFile(path)
		if err != nil {
			tm.logger.Warn("Failed to load template from %s: %v", path, err)
			return nil
		}

		if err := tm.RegisterTemplate(template); err != nil {
			tm.logger.Warn("Failed to register template from %s: %v", path, err)
		}

		return nil
	})
}

// loadTemplateFromFile loads a template from a YAML file
func (tm *TemplateManager) loadTemplateFromFile(path string) (*WorkflowTemplate, error) {
	// This would parse the YAML file and create a WorkflowTemplate
	// For now, return a simple implementation
	return &WorkflowTemplate{
		ID:          filepath.Base(path),
		Name:        filepath.Base(path),
		Description: "Template loaded from " + path,
	}, nil
}

// initializeBuiltinTemplates initializes built-in workflow templates
func (tm *TemplateManager) initializeBuiltinTemplates() {
	// System Update Template
	tm.templates["system-update"] = &WorkflowTemplate{
		ID:          "system-update",
		Name:        "System Update",
		Description: "Automated NixOS system update with backup and verification",
		Category:    "maintenance",
		Tags:        []string{"update", "maintenance", "system"},
		Parameters: []TemplateParameter{
			{
				Name:         "backup_enabled",
				Description:  "Enable backup before update",
				Type:         "bool",
				Required:     false,
				DefaultValue: true,
			},
			{
				Name:         "reboot_required",
				Description:  "Reboot after update if required",
				Type:         "bool",
				Required:     false,
				DefaultValue: false,
			},
		},
		Variables: map[string]interface{}{
			"update_channel": "nixos-unstable",
		},
	}

	// Package Cleanup Template
	tm.templates["package-cleanup"] = &WorkflowTemplate{
		ID:          "package-cleanup",
		Name:        "Package Cleanup",
		Description: "Automated package cleanup and garbage collection",
		Category:    "maintenance",
		Tags:        []string{"cleanup", "maintenance", "storage"},
		Parameters: []TemplateParameter{
			{
				Name:         "max_age_days",
				Description:  "Maximum age of packages to keep (days)",
				Type:         "int",
				Required:     false,
				DefaultValue: 30,
			},
			{
				Name:         "dry_run",
				Description:  "Perform dry run without actually removing packages",
				Type:         "bool",
				Required:     false,
				DefaultValue: false,
			},
		},
	}

	tm.logger.Info("Initialized %d built-in workflow templates", len(tm.templates))
}
