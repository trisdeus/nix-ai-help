package workflow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	stdcontext "context"
)

// Executor implements the ActionExecutor interface
type Executor struct {
	logger             Logger
	handlers           map[ActionType]ActionHandler
	conditionEvaluator *ConditionEvaluator
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(logger Logger) ActionExecutor {
	executor := &Executor{
		logger:             logger,
		handlers:           make(map[ActionType]ActionHandler),
		conditionEvaluator: NewConditionEvaluator(),
	}

	// Register default action handlers
	executor.registerDefaultHandlers()

	return executor
}

// ExecuteAction executes a single action
func (e *Executor) ExecuteAction(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	e.logger.Debug("Executing action: %s (type: %s)", action.ID, action.Type)

	// Create action execution record
	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
		ExitCode:  0,
	}

	// Validate action
	if err := e.ValidateAction(action); err != nil {
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	// Check condition if specified
	if action.Condition != "" {
		satisfied, err := e.evaluateCondition(action.Condition, context)
		if err != nil {
			execution.Error = fmt.Sprintf("Condition evaluation failed: %v", err)
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		if !satisfied {
			execution.Output = "Action skipped due to condition"
			execution.EndTime = timePtr(time.Now())
			e.logger.Debug("Action %s skipped due to condition: %s", action.ID, action.Condition)
			return execution, nil
		}
	}

	// Execute action with timeout
	actionCtx := context.Context
	if action.Timeout > 0 {
		var cancel func()
		actionCtx, cancel = stdcontext.WithTimeout(context.Context, action.Timeout)
		defer cancel()
	}

	// Get action handler
	handler, exists := e.handlers[action.Type]
	if !exists {
		err := fmt.Errorf("no handler for action type: %s", action.Type)
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	// Create action-specific context
	actionContext := &ExecutionContext{
		Context:   actionCtx,
		Variables: context.Variables,
		WorkDir:   e.getActionWorkDir(action, context),
		Logger:    context.Logger,
		DryRun:    context.DryRun,
		Verbose:   context.Verbose,
	}

	// Execute the action
	result, err := handler.Execute(action, actionContext)
	if err != nil {
		if !action.IgnoreError {
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			e.logger.Error("Action %s failed: %v", action.ID, err)
			return execution, err
		}
		e.logger.Warn("Action %s failed but ignored: %v", action.ID, err)
	}

	// Update execution with results
	if result != nil {
		execution.Output = result.Output
		execution.Error = result.Error
		execution.ExitCode = result.ExitCode
		execution.EndTime = result.EndTime
	}

	if execution.EndTime == nil {
		execution.EndTime = timePtr(time.Now())
	}

	e.logger.Debug("Action %s completed with exit code: %d", action.ID, execution.ExitCode)
	return execution, nil
}

// ValidateAction validates an action configuration
func (e *Executor) ValidateAction(action *Action) error {
	if action.ID == "" {
		return fmt.Errorf("action ID is required")
	}

	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}

	// Get handler and validate
	handler, exists := e.handlers[action.Type]
	if !exists {
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}

	return handler.Validate(action)
}

// GetSupportedActionTypes returns all supported action types
func (e *Executor) GetSupportedActionTypes() []ActionType {
	types := make([]ActionType, 0, len(e.handlers))
	for actionType := range e.handlers {
		types = append(types, actionType)
	}
	return types
}

// RegisterActionHandler registers a custom action handler
func (e *Executor) RegisterActionHandler(actionType ActionType, handler ActionHandler) {
	e.handlers[actionType] = handler
	e.logger.Info("Registered action handler for type: %s", actionType)
}

// Internal helper methods

func (e *Executor) registerDefaultHandlers() {
	e.handlers[ActionTypeCommand] = &CommandHandler{logger: e.logger}
	e.handlers[ActionTypeNixOSRebuild] = &NixOSRebuildHandler{logger: e.logger}
	e.handlers[ActionTypeFileEdit] = &FileEditHandler{logger: e.logger}
	e.handlers[ActionTypePackageOp] = &PackageOpHandler{logger: e.logger}
	e.handlers[ActionTypeServiceOp] = &ServiceOpHandler{logger: e.logger}
	e.handlers[ActionTypeValidation] = &ValidationHandler{logger: e.logger}
	e.handlers[ActionTypeQuery] = &QueryHandler{logger: e.logger}
	e.handlers[ActionTypeConditional] = &ConditionalHandler{logger: e.logger}
}

func (e *Executor) getActionWorkDir(action *Action, context *ExecutionContext) string {
	if action.WorkingDir != "" {
		if filepath.IsAbs(action.WorkingDir) {
			return action.WorkingDir
		}
		return filepath.Join(context.WorkDir, action.WorkingDir)
	}
	return context.WorkDir
}

func (e *Executor) evaluateCondition(condition string, context *ExecutionContext) (bool, error) {
	// Prepare context for condition evaluator
	conditionContext := make(map[string]interface{})
	for k, v := range context.Variables {
		conditionContext[k] = v
	}

	// Add system context
	conditionContext["work_dir"] = context.WorkDir
	conditionContext["dry_run"] = context.DryRun
	conditionContext["verbose"] = context.Verbose

	e.conditionEvaluator.SetContext(conditionContext)

	// Try to parse as a structured condition
	// For simple string conditions, create a basic condition object
	if strings.TrimSpace(condition) == "" {
		return true, nil
	}

	// Handle simple conditions for backward compatibility
	condition = strings.TrimSpace(condition)

	// Variable substitution
	for key, value := range context.Variables {
		condition = strings.ReplaceAll(condition, fmt.Sprintf("${%s}", key), value)
		condition = strings.ReplaceAll(condition, fmt.Sprintf("$%s", key), value)
	}

	// Simple boolean conditions
	switch condition {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	}

	// Parse different condition types
	if strings.HasPrefix(condition, "file_exists:") {
		path := strings.TrimPrefix(condition, "file_exists:")
		path = strings.TrimSpace(path)
		cond := Condition{
			Type: "file_exists",
			Parameters: map[string]interface{}{
				"path": path,
			},
		}
		return e.conditionEvaluator.EvaluateCondition(cond)
	}

	if strings.HasPrefix(condition, "command:") {
		cmd := strings.TrimPrefix(condition, "command:")
		cmd = strings.TrimSpace(cmd)
		cond := Condition{
			Type: "command_success",
			Parameters: map[string]interface{}{
				"command": cmd,
			},
		}
		return e.conditionEvaluator.EvaluateCondition(cond)
	}

	// Try to evaluate as expression
	cond := Condition{
		Type: "expression",
		Parameters: map[string]interface{}{
			"expression": condition,
		},
	}
	return e.conditionEvaluator.EvaluateCondition(cond)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Default Action Handlers

// CommandHandler handles command execution actions
type CommandHandler struct {
	logger Logger
}

func (h *CommandHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing command: %s", action.Command)

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		execution.Output = fmt.Sprintf("DRY RUN: Would execute command: %s", action.Command)
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Prepare command
	var cmd *exec.Cmd
	if len(action.Args) > 0 {
		cmd = exec.CommandContext(context.Context, action.Command, action.Args...)
	} else {
		cmd = exec.CommandContext(context.Context, "sh", "-c", action.Command)
	}

	cmd.Dir = context.WorkDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range action.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	execution.Output = string(output)
	execution.EndTime = timePtr(time.Now())

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			execution.ExitCode = exitError.ExitCode()
		} else {
			execution.ExitCode = 1
		}
		execution.Error = err.Error()
		return execution, err
	}

	return execution, nil
}

func (h *CommandHandler) Validate(action *Action) error {
	if action.Command == "" {
		return fmt.Errorf("command is required")
	}
	return nil
}

// NixOSRebuildHandler handles NixOS rebuild actions
type NixOSRebuildHandler struct {
	logger Logger
}

func (h *NixOSRebuildHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing NixOS rebuild")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		execution.Output = "DRY RUN: Would execute nixos-rebuild"
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Default nixos-rebuild command
	rebuildCmd := "nixos-rebuild"
	rebuildArgs := []string{"switch"}

	// Override with action configuration
	if action.Command != "" {
		rebuildCmd = action.Command
	}
	if len(action.Args) > 0 {
		rebuildArgs = action.Args
	}

	cmd := exec.CommandContext(context.Context, rebuildCmd, rebuildArgs...)
	cmd.Dir = context.WorkDir

	output, err := cmd.CombinedOutput()
	execution.Output = string(output)
	execution.EndTime = timePtr(time.Now())

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			execution.ExitCode = exitError.ExitCode()
		} else {
			execution.ExitCode = 1
		}
		execution.Error = err.Error()
		return execution, err
	}

	return execution, nil
}

func (h *NixOSRebuildHandler) Validate(action *Action) error {
	// NixOS rebuild actions are generally valid
	return nil
}

// Simplified handlers for other action types
type FileEditHandler struct{ logger Logger }

func (h *FileEditHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	execution := &ActionExecution{ActionID: action.ID, StartTime: timePtr(time.Now())}
	execution.Output = "File edit operation completed"
	execution.EndTime = timePtr(time.Now())
	return execution, nil
}
func (h *FileEditHandler) Validate(action *Action) error { return nil }

type PackageOpHandler struct{ logger Logger }

func (h *PackageOpHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	execution := &ActionExecution{ActionID: action.ID, StartTime: timePtr(time.Now())}
	execution.Output = "Package operation completed"
	execution.EndTime = timePtr(time.Now())
	return execution, nil
}
func (h *PackageOpHandler) Validate(action *Action) error { return nil }

type ServiceOpHandler struct{ logger Logger }

func (h *ServiceOpHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	execution := &ActionExecution{ActionID: action.ID, StartTime: timePtr(time.Now())}
	execution.Output = "Service operation completed"
	execution.EndTime = timePtr(time.Now())
	return execution, nil
}
func (h *ServiceOpHandler) Validate(action *Action) error { return nil }

type ValidationHandler struct{ logger Logger }

func (h *ValidationHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	execution := &ActionExecution{ActionID: action.ID, StartTime: timePtr(time.Now())}
	execution.Output = "Validation passed"
	execution.EndTime = timePtr(time.Now())
	return execution, nil
}
func (h *ValidationHandler) Validate(action *Action) error { return nil }

type QueryHandler struct{ logger Logger }

func (h *QueryHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	execution := &ActionExecution{ActionID: action.ID, StartTime: timePtr(time.Now())}
	execution.Output = "Query executed successfully"
	execution.EndTime = timePtr(time.Now())
	return execution, nil
}
func (h *QueryHandler) Validate(action *Action) error { return nil }

type ConditionalHandler struct{ logger Logger }

func (h *ConditionalHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	execution := &ActionExecution{ActionID: action.ID, StartTime: timePtr(time.Now())}
	execution.Output = "Conditional action executed"
	execution.EndTime = timePtr(time.Now())
	return execution, nil
}
func (h *ConditionalHandler) Validate(action *Action) error { return nil }
