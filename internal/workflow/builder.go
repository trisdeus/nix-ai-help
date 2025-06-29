package workflow

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/pkg/utils"
)

// WorkflowBuilder provides interactive workflow creation
type WorkflowBuilder struct {
	logger Logger
	reader *bufio.Reader
}

// NewWorkflowBuilder creates a new workflow builder
func NewWorkflowBuilder(logger Logger) *WorkflowBuilder {
	return &WorkflowBuilder{
		logger: logger,
		reader: bufio.NewReader(os.Stdin),
	}
}

// BuildWorkflowInteractive creates a workflow through interactive prompts
func (b *WorkflowBuilder) BuildWorkflowInteractive() (*WorkflowDefinition, error) {
	fmt.Println(utils.FormatHeader("🔨 Interactive Workflow Builder"))
	fmt.Println()
	fmt.Println("Let's create a new workflow step by step...")
	fmt.Println()

	workflow := &WorkflowDefinition{
		Tasks:     make([]TaskDefinition, 0),
		Triggers:  make([]TriggerDefinition, 0),
		Variables: make(map[string]interface{}),
		Metadata:  make(map[string]string),
	}

	// Collect basic information
	if err := b.collectBasicInfo(workflow); err != nil {
		return nil, fmt.Errorf("failed to collect basic info: %w", err)
	}

	// Collect configuration
	if err := b.collectConfiguration(workflow); err != nil {
		return nil, fmt.Errorf("failed to collect configuration: %w", err)
	}

	// Collect tasks
	if err := b.collectTasks(workflow); err != nil {
		return nil, fmt.Errorf("failed to collect tasks: %w", err)
	}

	// Collect triggers (optional)
	if err := b.collectTriggers(workflow); err != nil {
		return nil, fmt.Errorf("failed to collect triggers: %w", err)
	}

	// Collect variables (optional)
	if err := b.collectVariables(workflow); err != nil {
		return nil, fmt.Errorf("failed to collect variables: %w", err)
	}

	fmt.Println()
	fmt.Println(utils.FormatSuccess("✅ Workflow created successfully!"))
	fmt.Println()

	return workflow, nil
}

// collectBasicInfo collects basic workflow information
func (b *WorkflowBuilder) collectBasicInfo(workflow *WorkflowDefinition) error {
	fmt.Println(utils.FormatSubheader("📝 Basic Information"))
	fmt.Println()

	// Name
	name, err := b.promptRequired("Workflow name")
	if err != nil {
		return err
	}
	workflow.Name = name

	// Description
	description, err := b.promptRequired("Description")
	if err != nil {
		return err
	}
	workflow.Description = description

	// Version (optional)
	version, err := b.promptOptional("Version", "1.0.0")
	if err != nil {
		return err
	}
	workflow.Version = version

	// Author (optional)
	author, err := b.promptOptional("Author", "")
	if err != nil {
		return err
	}
	workflow.Author = author

	// Tags (optional)
	tagsStr, err := b.promptOptional("Tags (comma-separated)", "")
	if err != nil {
		return err
	}
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		workflow.Tags = tags
	}

	return nil
}

// collectConfiguration collects workflow configuration
func (b *WorkflowBuilder) collectConfiguration(workflow *WorkflowDefinition) error {
	fmt.Println()
	fmt.Println(utils.FormatSubheader("⚙️  Configuration"))
	fmt.Println()

	// Max concurrent tasks
	maxConcurrentStr, err := b.promptOptional("Max concurrent tasks", "5")
	if err != nil {
		return err
	}
	if maxConcurrentStr != "" {
		maxConcurrent, err := strconv.Atoi(maxConcurrentStr)
		if err != nil {
			return fmt.Errorf("invalid max concurrent tasks: %w", err)
		}
		workflow.Config.MaxConcurrentTasks = maxConcurrent
	}

	// Timeout
	timeoutStr, err := b.promptOptional("Timeout (e.g., 30m, 1h)", "30m")
	if err != nil {
		return err
	}
	if timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		workflow.Config.Timeout = timeout
	}

	// Retry count
	retryCountStr, err := b.promptOptional("Retry count", "3")
	if err != nil {
		return err
	}
	if retryCountStr != "" {
		retryCount, err := strconv.Atoi(retryCountStr)
		if err != nil {
			return fmt.Errorf("invalid retry count: %w", err)
		}
		workflow.Config.RetryCount = retryCount
	}

	// On failure behavior
	onFailure, err := b.promptChoice("On failure behavior", []string{"stop", "continue", "rollback"}, "stop")
	if err != nil {
		return err
	}
	workflow.Config.OnFailure = onFailure

	return nil
}

// collectTasks collects workflow tasks
func (b *WorkflowBuilder) collectTasks(workflow *WorkflowDefinition) error {
	fmt.Println()
	fmt.Println(utils.FormatSubheader("📋 Tasks"))
	fmt.Println()

	for {
		task, err := b.collectTask(len(workflow.Tasks) + 1)
		if err != nil {
			return err
		}

		workflow.Tasks = append(workflow.Tasks, *task)

		if !b.promptYesNo("Add another task?", false) {
			break
		}
		fmt.Println()
	}

	return nil
}

// collectTask collects a single task
func (b *WorkflowBuilder) collectTask(taskNum int) (*TaskDefinition, error) {
	fmt.Printf("Task %d:\n", taskNum)

	task := &TaskDefinition{
		Variables: make(map[string]interface{}),
	}

	// ID
	defaultID := fmt.Sprintf("task-%d", taskNum)
	id, err := b.promptOptional("Task ID", defaultID)
	if err != nil {
		return nil, err
	}
	task.ID = id

	// Name
	name, err := b.promptRequired("Task name")
	if err != nil {
		return nil, err
	}
	task.Name = name

	// Description
	description, err := b.promptOptional("Description", "")
	if err != nil {
		return nil, err
	}
	task.Description = description

	// Type
	taskType, err := b.promptChoice("Task type", []string{"action", "conditional", "parallel", "sequential"}, "action")
	if err != nil {
		return nil, err
	}
	task.Type = taskType

	// Action (if task type is action)
	if taskType == "action" {
		action, err := b.collectAction()
		if err != nil {
			return nil, err
		}
		task.Action = *action
	}

	// Dependencies
	deps, err := b.promptOptional("Dependencies (comma-separated task IDs)", "")
	if err != nil {
		return nil, err
	}
	if deps != "" {
		depList := strings.Split(deps, ",")
		for i, dep := range depList {
			depList[i] = strings.TrimSpace(dep)
		}
		task.DependsOn = depList
	}

	return task, nil
}

// collectAction collects action information
func (b *WorkflowBuilder) collectAction() (*ActionDefinition, error) {
	action := &ActionDefinition{
		Parameters: make(map[string]interface{}),
	}

	// Action type
	actionTypes := []string{
		"command",
		"nixos-rebuild",
		"file-edit",
		"file-copy",
		"service-control",
		"package-install",
		"config-update",
		"script-run",
	}

	actionType, err := b.promptChoice("Action type", actionTypes, "command")
	if err != nil {
		return nil, err
	}
	action.Type = actionType

	// Collect parameters based on action type
	switch actionType {
	case "command":
		cmd, err := b.promptRequired("Command to execute")
		if err != nil {
			return nil, err
		}
		action.Parameters["command"] = cmd

		args, err := b.promptOptional("Arguments (space-separated)", "")
		if err != nil {
			return nil, err
		}
		if args != "" {
			action.Parameters["args"] = strings.Fields(args)
		}

		workDir, err := b.promptOptional("Working directory", "")
		if err != nil {
			return nil, err
		}
		if workDir != "" {
			action.Parameters["workDir"] = workDir
		}

	case "nixos-rebuild":
		operation, err := b.promptChoice("Operation", []string{"switch", "boot", "test", "build"}, "switch")
		if err != nil {
			return nil, err
		}
		action.Parameters["operation"] = operation

		flakePath, err := b.promptOptional("Flake path", "")
		if err != nil {
			return nil, err
		}
		if flakePath != "" {
			action.Parameters["flakePath"] = flakePath
		}

	case "file-edit":
		filePath, err := b.promptRequired("File path")
		if err != nil {
			return nil, err
		}
		action.Parameters["filePath"] = filePath

		content, err := b.promptOptional("Content", "")
		if err != nil {
			return nil, err
		}
		if content != "" {
			action.Parameters["content"] = content
		}

	case "service-control":
		serviceName, err := b.promptRequired("Service name")
		if err != nil {
			return nil, err
		}
		action.Parameters["serviceName"] = serviceName

		operation, err := b.promptChoice("Operation", []string{"start", "stop", "restart", "reload", "enable", "disable"}, "start")
		if err != nil {
			return nil, err
		}
		action.Parameters["operation"] = operation
	}

	return action, nil
}

// collectTriggers collects workflow triggers
func (b *WorkflowBuilder) collectTriggers(workflow *WorkflowDefinition) error {
	fmt.Println()
	fmt.Println(utils.FormatSubheader("🔔 Triggers (Optional)"))
	fmt.Println()

	if !b.promptYesNo("Add triggers to this workflow?", false) {
		return nil
	}

	for {
		trigger, err := b.collectTrigger()
		if err != nil {
			return err
		}

		workflow.Triggers = append(workflow.Triggers, *trigger)

		if !b.promptYesNo("Add another trigger?", false) {
			break
		}
		fmt.Println()
	}

	return nil
}

// collectTrigger collects a single trigger
func (b *WorkflowBuilder) collectTrigger() (*TriggerDefinition, error) {
	trigger := &TriggerDefinition{
		Parameters: make(map[string]interface{}),
	}

	// Trigger type
	triggerTypes := []string{
		"manual",
		"schedule",
		"file-change",
		"event",
	}

	triggerType, err := b.promptChoice("Trigger type", triggerTypes, "manual")
	if err != nil {
		return nil, err
	}
	trigger.Type = triggerType

	// Collect parameters based on trigger type
	switch triggerType {
	case "schedule":
		schedule, err := b.promptRequired("Cron schedule (e.g., '0 2 * * *')")
		if err != nil {
			return nil, err
		}
		trigger.Parameters["schedule"] = schedule

	case "file-change":
		filePath, err := b.promptRequired("File/directory path to watch")
		if err != nil {
			return nil, err
		}
		trigger.Parameters["path"] = filePath

		events, err := b.promptOptional("Events to watch (create,modify,delete)", "modify")
		if err != nil {
			return nil, err
		}
		trigger.Parameters["events"] = strings.Split(events, ",")
	}

	return trigger, nil
}

// collectVariables collects workflow variables
func (b *WorkflowBuilder) collectVariables(workflow *WorkflowDefinition) error {
	fmt.Println()
	fmt.Println(utils.FormatSubheader("🔧 Variables (Optional)"))
	fmt.Println()

	if !b.promptYesNo("Add variables to this workflow?", false) {
		return nil
	}

	for {
		name, err := b.promptRequired("Variable name")
		if err != nil {
			return err
		}

		value, err := b.promptRequired("Variable value")
		if err != nil {
			return err
		}

		workflow.Variables[name] = value

		if !b.promptYesNo("Add another variable?", false) {
			break
		}
	}

	return nil
}

// Helper prompt functions

func (b *WorkflowBuilder) promptRequired(label string) (string, error) {
	for {
		fmt.Printf("%s: ", label)
		input, err := b.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)
		if input != "" {
			return input, nil
		}

		fmt.Println(utils.FormatError("This field is required. Please enter a value."))
	}
}

func (b *WorkflowBuilder) promptOptional(label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s (optional): ", label)
	}

	input, err := b.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}

	return input, nil
}

func (b *WorkflowBuilder) promptChoice(label string, choices []string, defaultChoice string) (string, error) {
	fmt.Printf("%s [%s] (", label, defaultChoice)
	for i, choice := range choices {
		if i > 0 {
			fmt.Print("/")
		}
		fmt.Print(choice)
	}
	fmt.Print("): ")

	input, err := b.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultChoice, nil
	}

	// Validate choice
	for _, choice := range choices {
		if strings.EqualFold(input, choice) {
			return choice, nil
		}
	}

	return "", fmt.Errorf("invalid choice: %s", input)
}

func (b *WorkflowBuilder) promptYesNo(label string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	fmt.Printf("%s [%s] (y/n): ", label, defaultStr)

	input, err := b.reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}
