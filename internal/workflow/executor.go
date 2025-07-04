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

// FileEditHandler handles file editing actions
type FileEditHandler struct{ logger Logger }

func (h *FileEditHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing file edit action")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		execution.Output = fmt.Sprintf("DRY RUN: Would edit file: %s", action.Config["file"])
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Get file path from parameters
	filePath, exists := action.Config["file"].(string)
	if !exists || filePath == "" {
		err := fmt.Errorf("file parameter is required for file edit action")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	// Make file path absolute if relative
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(context.WorkDir, filePath)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Create file if it doesn't exist
		if createFlag, ok := action.Config["create"].(bool); ok && createFlag {
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				execution.Error = fmt.Sprintf("Failed to create directory: %v", err)
				execution.ExitCode = 1
				execution.EndTime = timePtr(time.Now())
				return execution, fmt.Errorf("failed to create directory: %w", err)
			}
		} else {
			err := fmt.Errorf("file does not exist: %s", filePath)
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
	}

	// Handle different edit operations
	operation, _ := action.Config["operation"].(string)
	switch operation {
	case "append":
		content, _ := action.Config["content"].(string)
		if content != "" {
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				execution.Error = fmt.Sprintf("Failed to open file: %v", err)
				execution.ExitCode = 1
				execution.EndTime = timePtr(time.Now())
				return execution, err
			}
			defer file.Close()

			if _, err := file.WriteString(content + "\n"); err != nil {
				execution.Error = fmt.Sprintf("Failed to write to file: %v", err)
				execution.ExitCode = 1
				execution.EndTime = timePtr(time.Now())
				return execution, err
			}
		}
		execution.Output = fmt.Sprintf("Appended content to file: %s", filePath)

	case "replace":
		content, _ := action.Config["content"].(string)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			execution.Error = fmt.Sprintf("Failed to write file: %v", err)
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		execution.Output = fmt.Sprintf("Replaced content in file: %s", filePath)

	default:
		// Default to creating or touching the file
		if _, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0644); err != nil {
			execution.Error = fmt.Sprintf("Failed to create/touch file: %v", err)
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		execution.Output = fmt.Sprintf("Created/touched file: %s", filePath)
	}

	execution.EndTime = timePtr(time.Now())
	return execution, nil
}

func (h *FileEditHandler) Validate(action *Action) error {
	if action.Config == nil {
		return fmt.Errorf("parameters are required for file edit action")
	}
	if _, exists := action.Config["file"]; !exists {
		return fmt.Errorf("file parameter is required")
	}
	return nil
}

// PackageOpHandler handles package operations
type PackageOpHandler struct{ logger Logger }

func (h *PackageOpHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing package operation")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		operation, _ := action.Config["operation"].(string)
		packageName, _ := action.Config["package"].(string)
		execution.Output = fmt.Sprintf("DRY RUN: Would %s package: %s", operation, packageName)
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Get operation and package name from parameters
	operation, exists := action.Config["operation"].(string)
	if !exists {
		err := fmt.Errorf("operation parameter is required for package operation")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	packageName, exists := action.Config["package"].(string)
	if !exists {
		err := fmt.Errorf("package parameter is required for package operation")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	var cmd *exec.Cmd
	switch operation {
	case "install":
		// For NixOS, we use nix-env or add to configuration
		cmd = exec.CommandContext(context.Context, "nix-env", "-iA", "nixpkgs."+packageName)
	case "remove", "uninstall":
		cmd = exec.CommandContext(context.Context, "nix-env", "--uninstall", packageName)
	case "upgrade":
		cmd = exec.CommandContext(context.Context, "nix-env", "--upgrade", packageName)
	case "search":
		cmd = exec.CommandContext(context.Context, "nix", "search", "nixpkgs", packageName)
	default:
		err := fmt.Errorf("unsupported package operation: %s", operation)
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	cmd.Dir = context.WorkDir

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

	h.logger.Info("Package operation completed: %s %s", operation, packageName)
	return execution, nil
}

func (h *PackageOpHandler) Validate(action *Action) error {
	if action.Config == nil {
		return fmt.Errorf("parameters are required for package operation")
	}
	if _, exists := action.Config["operation"]; !exists {
		return fmt.Errorf("operation parameter is required")
	}
	if _, exists := action.Config["package"]; !exists {
		return fmt.Errorf("package parameter is required")
	}
	
	// Validate operation type
	operation, _ := action.Config["operation"].(string)
	validOps := []string{"install", "remove", "uninstall", "upgrade", "search"}
	valid := false
	for _, op := range validOps {
		if operation == op {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid operation: %s. Must be one of: %v", operation, validOps)
	}
	
	return nil
}

// ServiceOpHandler handles systemd service operations
type ServiceOpHandler struct{ logger Logger }

func (h *ServiceOpHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing service operation")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		operation, _ := action.Config["operation"].(string)
		serviceName, _ := action.Config["service"].(string)
		execution.Output = fmt.Sprintf("DRY RUN: Would %s service: %s", operation, serviceName)
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Get operation and service name from parameters
	operation, exists := action.Config["operation"].(string)
	if !exists {
		err := fmt.Errorf("operation parameter is required for service operation")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	serviceName, exists := action.Config["service"].(string)
	if !exists {
		err := fmt.Errorf("service parameter is required for service operation")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	var cmd *exec.Cmd
	switch operation {
	case "start":
		cmd = exec.CommandContext(context.Context, "sudo", "systemctl", "start", serviceName)
	case "stop":
		cmd = exec.CommandContext(context.Context, "sudo", "systemctl", "stop", serviceName)
	case "restart":
		cmd = exec.CommandContext(context.Context, "sudo", "systemctl", "restart", serviceName)
	case "reload":
		cmd = exec.CommandContext(context.Context, "sudo", "systemctl", "reload", serviceName)
	case "enable":
		cmd = exec.CommandContext(context.Context, "sudo", "systemctl", "enable", serviceName)
	case "disable":
		cmd = exec.CommandContext(context.Context, "sudo", "systemctl", "disable", serviceName)
	case "status":
		cmd = exec.CommandContext(context.Context, "systemctl", "status", serviceName)
	case "is-active":
		cmd = exec.CommandContext(context.Context, "systemctl", "is-active", serviceName)
	case "is-enabled":
		cmd = exec.CommandContext(context.Context, "systemctl", "is-enabled", serviceName)
	default:
		err := fmt.Errorf("unsupported service operation: %s", operation)
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	cmd.Dir = context.WorkDir

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
		// Some systemctl commands return non-zero exit codes for informational purposes
		// Don't treat status checks as errors if they just indicate the service is not running
		if operation == "status" || operation == "is-active" || operation == "is-enabled" {
			execution.Error = "" // Clear error for status checks
			h.logger.Debug("Service %s %s returned exit code %d: %s", serviceName, operation, execution.ExitCode, string(output))
		} else {
			execution.Error = err.Error()
			return execution, err
		}
	}

	h.logger.Info("Service operation completed: %s %s", operation, serviceName)
	return execution, nil
}

func (h *ServiceOpHandler) Validate(action *Action) error {
	if action.Config == nil {
		return fmt.Errorf("parameters are required for service operation")
	}
	if _, exists := action.Config["operation"]; !exists {
		return fmt.Errorf("operation parameter is required")
	}
	if _, exists := action.Config["service"]; !exists {
		return fmt.Errorf("service parameter is required")
	}
	
	// Validate operation type
	operation, _ := action.Config["operation"].(string)
	validOps := []string{"start", "stop", "restart", "reload", "enable", "disable", "status", "is-active", "is-enabled"}
	valid := false
	for _, op := range validOps {
		if operation == op {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid operation: %s. Must be one of: %v", operation, validOps)
	}
	
	return nil
}

// ValidationHandler handles validation actions
type ValidationHandler struct{ logger Logger }

func (h *ValidationHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing validation action")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		validationType, _ := action.Config["type"].(string)
		execution.Output = fmt.Sprintf("DRY RUN: Would perform %s validation", validationType)
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Get validation type from parameters
	validationType, exists := action.Config["type"].(string)
	if !exists {
		err := fmt.Errorf("type parameter is required for validation action")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	switch validationType {
	case "file_exists":
		filePath, exists := action.Config["path"].(string)
		if !exists {
			err := fmt.Errorf("path parameter is required for file_exists validation")
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}

		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(context.WorkDir, filePath)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			err = fmt.Errorf("file does not exist: %s", filePath)
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		execution.Output = fmt.Sprintf("File exists: %s", filePath)

	case "command_success":
		command, exists := action.Config["command"].(string)
		if !exists {
			err := fmt.Errorf("command parameter is required for command_success validation")
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}

		cmd := exec.CommandContext(context.Context, "sh", "-c", command)
		cmd.Dir = context.WorkDir
		output, err := cmd.CombinedOutput()

		if err != nil {
			err = fmt.Errorf("validation command failed: %s", string(output))
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		execution.Output = fmt.Sprintf("Command validation passed: %s", command)

	case "nixos_config":
		// Validate NixOS configuration syntax
		configPath, _ := action.Config["config_path"].(string)
		if configPath == "" {
			configPath = "/etc/nixos/configuration.nix"
		}

		cmd := exec.CommandContext(context.Context, "nixos-rebuild", "dry-build")
		cmd.Dir = context.WorkDir
		output, err := cmd.CombinedOutput()

		if err != nil {
			err = fmt.Errorf("NixOS configuration validation failed: %s", string(output))
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		execution.Output = "NixOS configuration validation passed"

	case "service_running":
		serviceName, exists := action.Config["service"].(string)
		if !exists {
			err := fmt.Errorf("service parameter is required for service_running validation")
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}

		cmd := exec.CommandContext(context.Context, "systemctl", "is-active", serviceName)
		output, err := cmd.CombinedOutput()
		outputStr := strings.TrimSpace(string(output))

		if err != nil || outputStr != "active" {
			err = fmt.Errorf("service %s is not running: %s", serviceName, outputStr)
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		execution.Output = fmt.Sprintf("Service is running: %s", serviceName)

	default:
		err := fmt.Errorf("unsupported validation type: %s", validationType)
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	execution.EndTime = timePtr(time.Now())
	h.logger.Info("Validation completed: %s", validationType)
	return execution, nil
}

func (h *ValidationHandler) Validate(action *Action) error {
	if action.Config == nil {
		return fmt.Errorf("parameters are required for validation action")
	}
	if _, exists := action.Config["type"]; !exists {
		return fmt.Errorf("type parameter is required")
	}
	
	// Validate validation type
	validationType, _ := action.Config["type"].(string)
	validTypes := []string{"file_exists", "command_success", "nixos_config", "service_running"}
	valid := false
	for _, vType := range validTypes {
		if validationType == vType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid validation type: %s. Must be one of: %v", validationType, validTypes)
	}
	
	return nil
}

// QueryHandler handles query actions (information gathering)
type QueryHandler struct{ logger Logger }

func (h *QueryHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing query action")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		queryType, _ := action.Config["type"].(string)
		execution.Output = fmt.Sprintf("DRY RUN: Would execute %s query", queryType)
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Get query type from parameters
	queryType, exists := action.Config["type"].(string)
	if !exists {
		err := fmt.Errorf("type parameter is required for query action")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	var cmd *exec.Cmd
	switch queryType {
	case "system_info":
		cmd = exec.CommandContext(context.Context, "uname", "-a")
	case "disk_usage":
		path, _ := action.Config["path"].(string)
		if path == "" {
			path = "/"
		}
		cmd = exec.CommandContext(context.Context, "df", "-h", path)
	case "memory_usage":
		cmd = exec.CommandContext(context.Context, "free", "-h")
	case "cpu_info":
		cmd = exec.CommandContext(context.Context, "lscpu")
	case "nixos_version":
		cmd = exec.CommandContext(context.Context, "nixos-version")
	case "nix_packages":
		cmd = exec.CommandContext(context.Context, "nix-env", "--query", "--installed")
	case "services_status":
		serviceName, _ := action.Config["service"].(string)
		if serviceName != "" {
			cmd = exec.CommandContext(context.Context, "systemctl", "status", serviceName)
		} else {
			cmd = exec.CommandContext(context.Context, "systemctl", "list-units", "--type=service", "--state=running")
		}
	case "network_info":
		cmd = exec.CommandContext(context.Context, "ip", "addr", "show")
	case "file_content":
		filePath, exists := action.Config["file"].(string)
		if !exists {
			err := fmt.Errorf("file parameter is required for file_content query")
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(context.WorkDir, filePath)
		}
		cmd = exec.CommandContext(context.Context, "cat", filePath)
	case "directory_listing":
		dirPath, _ := action.Config["path"].(string)
		if dirPath == "" {
			dirPath = context.WorkDir
		}
		if !filepath.IsAbs(dirPath) {
			dirPath = filepath.Join(context.WorkDir, dirPath)
		}
		cmd = exec.CommandContext(context.Context, "ls", "-la", dirPath)
	case "command":
		command, exists := action.Config["command"].(string)
		if !exists {
			err := fmt.Errorf("command parameter is required for command query")
			execution.Error = err.Error()
			execution.ExitCode = 1
			execution.EndTime = timePtr(time.Now())
			return execution, err
		}
		cmd = exec.CommandContext(context.Context, "sh", "-c", command)
	default:
		err := fmt.Errorf("unsupported query type: %s", queryType)
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	cmd.Dir = context.WorkDir

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

	h.logger.Info("Query completed: %s", queryType)
	return execution, nil
}

func (h *QueryHandler) Validate(action *Action) error {
	if action.Config == nil {
		return fmt.Errorf("parameters are required for query action")
	}
	if _, exists := action.Config["type"]; !exists {
		return fmt.Errorf("type parameter is required")
	}
	
	// Validate query type
	queryType, _ := action.Config["type"].(string)
	validTypes := []string{
		"system_info", "disk_usage", "memory_usage", "cpu_info", "nixos_version",
		"nix_packages", "services_status", "network_info", "file_content",
		"directory_listing", "command",
	}
	valid := false
	for _, qType := range validTypes {
		if queryType == qType {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid query type: %s. Must be one of: %v", queryType, validTypes)
	}
	
	return nil
}

// ConditionalHandler handles conditional actions
type ConditionalHandler struct{ logger Logger }

func (h *ConditionalHandler) Execute(action *Action, context *ExecutionContext) (*ActionExecution, error) {
	h.logger.Debug("Executing conditional action")

	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: timePtr(time.Now()),
	}

	if context.DryRun {
		condition, _ := action.Config["condition"].(string)
		execution.Output = fmt.Sprintf("DRY RUN: Would evaluate condition: %s", condition)
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Get condition from parameters
	condition, exists := action.Config["condition"].(string)
	if !exists {
		err := fmt.Errorf("condition parameter is required for conditional action")
		execution.Error = err.Error()
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	// Create a condition evaluator instance
	evaluator := NewConditionEvaluator()
	
	// Set up context for evaluation
	conditionContext := make(map[string]interface{})
	for k, v := range context.Variables {
		conditionContext[k] = v
	}
	conditionContext["work_dir"] = context.WorkDir
	conditionContext["dry_run"] = context.DryRun
	conditionContext["verbose"] = context.Verbose
	
	evaluator.SetContext(conditionContext)

	// Handle different condition formats
	var conditionObj Condition
	
	// Try to parse as a simple boolean condition first
	switch strings.TrimSpace(condition) {
	case "true", "1", "yes":
		execution.Output = "Condition evaluated to true"
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	case "false", "0", "no":
		execution.Output = "Condition evaluated to false"
		execution.EndTime = timePtr(time.Now())
		return execution, nil
	}

	// Parse structured conditions
	if strings.HasPrefix(condition, "file_exists:") {
		path := strings.TrimSpace(strings.TrimPrefix(condition, "file_exists:"))
		conditionObj = Condition{
			Type: "file_exists",
			Parameters: map[string]interface{}{
				"path": path,
			},
		}
	} else if strings.HasPrefix(condition, "command:") {
		cmd := strings.TrimSpace(strings.TrimPrefix(condition, "command:"))
		conditionObj = Condition{
			Type: "command_success",
			Parameters: map[string]interface{}{
				"command": cmd,
			},
		}
	} else {
		// Treat as expression
		conditionObj = Condition{
			Type: "expression",
			Parameters: map[string]interface{}{
				"expression": condition,
			},
		}
	}

	// Evaluate the condition
	result, err := evaluator.EvaluateCondition(conditionObj)
	if err != nil {
		execution.Error = fmt.Sprintf("Condition evaluation failed: %v", err)
		execution.ExitCode = 1
		execution.EndTime = timePtr(time.Now())
		return execution, err
	}

	if result {
		execution.Output = "Condition evaluated to true"
		
		// Execute then action if specified
		if thenAction, exists := action.Config["then"]; exists {
			execution.Output += "; executing then action"
			// In a full implementation, we would execute the then action here
			// For now, just log what would happen
			h.logger.Debug("Would execute then action: %v", thenAction)
		}
	} else {
		execution.Output = "Condition evaluated to false"
		
		// Execute else action if specified
		if elseAction, exists := action.Config["else"]; exists {
			execution.Output += "; executing else action"
			// In a full implementation, we would execute the else action here
			// For now, just log what would happen
			h.logger.Debug("Would execute else action: %v", elseAction)
		}
	}

	execution.EndTime = timePtr(time.Now())
	h.logger.Info("Conditional action completed: %s -> %t", condition, result)
	return execution, nil
}

func (h *ConditionalHandler) Validate(action *Action) error {
	if action.Config == nil {
		return fmt.Errorf("parameters are required for conditional action")
	}
	if _, exists := action.Config["condition"]; !exists {
		return fmt.Errorf("condition parameter is required")
	}
	return nil
}
