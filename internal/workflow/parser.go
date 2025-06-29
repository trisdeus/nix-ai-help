package workflow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// WorkflowDefinition represents a workflow definition in YAML format
type WorkflowDefinition struct {
	// Metadata
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Author      string            `yaml:"author,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`

	// Configuration
	Config WorkflowConfig `yaml:"config,omitempty"`

	// Tasks
	Tasks []TaskDefinition `yaml:"tasks"`

	// Triggers
	Triggers []TriggerDefinition `yaml:"triggers,omitempty"`

	// Variables
	Variables map[string]interface{} `yaml:"variables,omitempty"`
}

// WorkflowConfig defines workflow execution configuration
type WorkflowConfig struct {
	MaxConcurrentTasks int           `yaml:"maxConcurrentTasks,omitempty"`
	Timeout            time.Duration `yaml:"timeout,omitempty"`
	RetryCount         int           `yaml:"retryCount,omitempty"`
	OnFailure          string        `yaml:"onFailure,omitempty"` // continue, stop, rollback
	LogLevel           string        `yaml:"logLevel,omitempty"`
}

// TaskDefinition represents a task definition in YAML format
type TaskDefinition struct {
	ID          string                 `yaml:"id"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Type        string                 `yaml:"type"` // action, conditional, parallel, sequential
	DependsOn   []string               `yaml:"dependsOn,omitempty"`
	Condition   string                 `yaml:"condition,omitempty"`
	Timeout     time.Duration          `yaml:"timeout,omitempty"`
	RetryCount  int                    `yaml:"retryCount,omitempty"`
	OnFailure   string                 `yaml:"onFailure,omitempty"`
	Action      ActionDefinition       `yaml:"action,omitempty"`
	Variables   map[string]interface{} `yaml:"variables,omitempty"`
	Tags        []string               `yaml:"tags,omitempty"`
}

// ActionDefinition represents an action definition in YAML format
type ActionDefinition struct {
	Type       string                 `yaml:"type"` // command, nixos-rebuild, file-edit, etc.
	Parameters map[string]interface{} `yaml:"parameters"`
}

// TriggerDefinition represents a trigger definition in YAML format
type TriggerDefinition struct {
	Type       string                 `yaml:"type"` // schedule, file-change, manual, etc.
	Condition  string                 `yaml:"condition,omitempty"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

// WorkflowParser handles parsing workflow definitions from YAML
type WorkflowParser struct {
	logger Logger
}

// NewWorkflowParser creates a new workflow parser
func NewWorkflowParser(logger Logger) *WorkflowParser {
	return &WorkflowParser{
		logger: logger,
	}
}

// ParseWorkflow parses a workflow definition from a reader
func (p *WorkflowParser) ParseWorkflow(reader io.Reader) (*WorkflowDefinition, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow data: %w", err)
	}

	var workflow WorkflowDefinition
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	// Validate the workflow
	if err := p.validateWorkflow(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return &workflow, nil
}

// ParseWorkflowFile parses a workflow definition from a file
func (p *WorkflowParser) ParseWorkflowFile(filename string) (*WorkflowDefinition, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open workflow file %s: %w", filename, err)
	}
	defer file.Close()

	workflow, err := p.ParseWorkflow(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow file %s: %w", filename, err)
	}

	return workflow, nil
}

// ParseWorkflowDirectory parses all workflow files in a directory
func (p *WorkflowParser) ParseWorkflowDirectory(directory string) (map[string]*WorkflowDefinition, error) {
	workflows := make(map[string]*WorkflowDefinition)

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process YAML/YML files
		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml")) {
			workflow, parseErr := p.ParseWorkflowFile(path)
			if parseErr != nil {
				p.logger.Warn("Failed to parse workflow file", "file", path, "error", parseErr)
				return nil // Continue with other files
			}

			// Use filename without extension as key if no name is provided
			key := workflow.Name
			if key == "" {
				key = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}

			workflows[key] = workflow
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow directory %s: %w", directory, err)
	}

	return workflows, nil
}

// validateWorkflow validates a workflow definition
func (p *WorkflowParser) validateWorkflow(workflow *WorkflowDefinition) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Tasks) == 0 {
		return fmt.Errorf("workflow must have at least one task")
	}

	// Validate task IDs are unique
	taskIDs := make(map[string]bool)
	for _, task := range workflow.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID is required")
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true

		// Validate task dependencies
		for _, dep := range task.DependsOn {
			if !taskIDs[dep] && dep != task.ID {
				// Check if dependency exists in previously validated tasks
				found := false
				for _, prevTask := range workflow.Tasks {
					if prevTask.ID == dep {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("task %s depends on non-existent task: %s", task.ID, dep)
				}
			}
		}

		// Validate action
		if task.Type == "action" && task.Action.Type == "" {
			return fmt.Errorf("task %s has action type but no action specified", task.ID)
		}
	}

	return nil
}

// ConvertToWorkflow converts a WorkflowDefinition to a Workflow
func (p *WorkflowParser) ConvertToWorkflow(def *WorkflowDefinition) (*Workflow, error) {
	// Convert variables from interface{} to string
	variables := make(map[string]string)
	for k, v := range def.Variables {
		variables[k] = fmt.Sprintf("%v", v)
	}

	workflow := &Workflow{
		ID:          GenerateWorkflowID(def.Name),
		Name:        def.Name,
		Description: def.Description,
		Version:     def.Version,
		Author:      def.Author,
		Tags:        def.Tags,
		Status:      StatusPending,
		Metadata:    def.Metadata,
		Variables:   variables,
		Tasks:       make([]Task, 0, len(def.Tasks)),
		Triggers:    make([]Trigger, 0, len(def.Triggers)),
		MaxRetries:  def.Config.RetryCount,
		Timeout:     def.Config.Timeout,
	}

	// Convert tasks
	for _, taskDef := range def.Tasks {
		task, err := p.convertToTask(&taskDef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert task %s: %w", taskDef.ID, err)
		}
		workflow.Tasks = append(workflow.Tasks, *task)
	}

	// Convert triggers
	for _, triggerDef := range def.Triggers {
		trigger, err := p.convertToTrigger(&triggerDef)
		if err != nil {
			return nil, fmt.Errorf("failed to convert trigger: %w", err)
		}
		workflow.Triggers = append(workflow.Triggers, *trigger)
	}

	return workflow, nil
}

// convertToTask converts a TaskDefinition to a Task
func (p *WorkflowParser) convertToTask(def *TaskDefinition) (*Task, error) {
	// Convert variables from interface{} to string
	variables := make(map[string]string)
	for k, v := range def.Variables {
		variables[k] = fmt.Sprintf("%v", v)
	}

	task := &Task{
		ID:          def.ID,
		Name:        def.Name,
		Description: def.Description,
		Status:      TaskStatusPending,
		DependsOn:   def.DependsOn,
		Condition:   def.Condition,
		Variables:   variables,
		MaxRetries:  def.RetryCount,
		Timeout:     def.Timeout,
		Actions:     make([]Action, 0),
	}

	// Convert action if present
	if def.Action.Type != "" {
		action := Action{
			ID:     def.ID + "-action",
			Type:   ActionType(def.Action.Type),
			Config: def.Action.Parameters,
		}

		// Map common parameters to action fields
		if cmd, ok := def.Action.Parameters["command"]; ok {
			action.Command = fmt.Sprintf("%v", cmd)
		}
		if args, ok := def.Action.Parameters["args"]; ok {
			if argList, ok := args.([]interface{}); ok {
				action.Args = make([]string, len(argList))
				for i, arg := range argList {
					action.Args[i] = fmt.Sprintf("%v", arg)
				}
			}
		}
		if workDir, ok := def.Action.Parameters["workDir"]; ok {
			action.WorkingDir = fmt.Sprintf("%v", workDir)
		}

		task.Actions = append(task.Actions, action)
	}

	return task, nil
}

// convertToTrigger converts a TriggerDefinition to a Trigger
func (p *WorkflowParser) convertToTrigger(def *TriggerDefinition) (*Trigger, error) {
	trigger := &Trigger{
		ID:        generateTriggerID(def.Type),
		Type:      TriggerType(def.Type),
		Condition: def.Condition,
		Config:    def.Parameters,
		Enabled:   true,
	}

	// Map common parameters to trigger fields
	if schedule, ok := def.Parameters["schedule"]; ok {
		trigger.Schedule = fmt.Sprintf("%v", schedule)
	}
	if path, ok := def.Parameters["path"]; ok {
		trigger.Path = fmt.Sprintf("%v", path)
	}
	if event, ok := def.Parameters["event"]; ok {
		trigger.Event = fmt.Sprintf("%v", event)
	}

	return trigger, nil
}

// generateTriggerID generates a trigger ID
func generateTriggerID(triggerType string) string {
	return fmt.Sprintf("trigger-%s-%d", triggerType, time.Now().Unix())
}

// GenerateWorkflowID generates a workflow ID from the name
func GenerateWorkflowID(name string) string {
	// Convert to lowercase and replace spaces with hyphens
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")

	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
