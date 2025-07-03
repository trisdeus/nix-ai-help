package execution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/security"
	"nix-ai-help/internal/types"
	"nix-ai-help/pkg/logger"
)

// SafeExecutor provides secure command execution with validation and auditing
type SafeExecutor struct {
	permissionManager *security.PermissionManager
	auditLogger       *security.AuditLogger
	sudoManager       *security.SudoManager
	commandResolver   *CommandResolver
	config           *config.ExecutionConfig
	logger           *logger.Logger
	dryRunMode       bool
}

// NewSafeExecutor creates a new SafeExecutor instance
func NewSafeExecutor(
	permissionManager *security.PermissionManager,
	auditLogger *security.AuditLogger,
	sudoManager *security.SudoManager,
	config *config.ExecutionConfig,
	logger *logger.Logger,
) *SafeExecutor {
	commandResolver := NewCommandResolver(logger)
	
	return &SafeExecutor{
		permissionManager: permissionManager,
		auditLogger:       auditLogger,
		sudoManager:       sudoManager,
		commandResolver:   commandResolver,
		config:           config,
		logger:           logger,
		dryRunMode:       config.DryRunDefault,
	}
}

// ExecuteCommand executes a command with full security validation
func (se *SafeExecutor) ExecuteCommand(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// Log the execution attempt
	se.auditLogger.LogCommandAttempt(req)
	
	// Resolve command availability and get execution strategy
	resolution, resolveErr := se.commandResolver.ResolveCommand(ctx, req.Command)
	if resolveErr != nil {
		se.logger.Warn(fmt.Sprintf("Command resolution failed for %s: %v", req.Command, resolveErr))
	}
	
	// If command is not available but can be run with nix run, update the request
	if resolution != nil && resolution.Availability == CommandNixRunnable {
		se.logger.Info(fmt.Sprintf("Command %s not available locally, using nix run: %s", req.Command, resolution.NixRunCommand))
		
		// Parse the nix run command
		parts := strings.Fields(resolution.NixRunCommand)
		if len(parts) > 0 {
			req.Command = parts[0] // "nix"
			// Combine nix run arguments with original arguments
			nixArgs := parts[1:] // ["run", "nixpkgs#package"]
			if len(req.Args) > 0 {
				nixArgs = append(nixArgs, "--")
				nixArgs = append(nixArgs, req.Args...)
			}
			req.Args = nixArgs
		}
		
		// Add informative description about nix run usage
		if req.Description == "" {
			req.Description = fmt.Sprintf("Running %s via nix run (package: %s)", req.Command, resolution.NixPackage)
		} else {
			req.Description = fmt.Sprintf("%s (via nix run - package: %s)", req.Description, resolution.NixPackage)
		}
	}
	
	// Validate the command (after potential nix run modification)
	if err := se.validateCommand(req); err != nil {
		se.auditLogger.LogCommandRejected(req, err.Error())
		return nil, fmt.Errorf("command validation failed: %w", err)
	}
	
	// Check if dry run is requested or enabled by default
	if req.DryRun || se.dryRunMode {
		return se.dryRun(ctx, req)
	}
	
	// Check if confirmation is required
	if se.permissionManager.RequiresConfirmation(req) {
		confirmed, err := se.requestConfirmation(req)
		if err != nil {
			return nil, fmt.Errorf("confirmation request failed: %w", err)
		}
		if !confirmed {
			se.auditLogger.LogCommandDenied(req, "user denied confirmation")
			return &ExecutionResult{
				Success:   false,
				Error:     "execution denied by user",
				Command:   req.Command,
				Timestamp: startTime,
			}, nil
		}
	}
	
	// Execute with or without sudo based on requirements
	var result *ExecutionResult
	var err error
	
	if req.RequiresSudo {
		result, err = se.executeWithSudo(ctx, req)
	} else {
		result, err = se.executeRegular(ctx, req)
	}
	
	// Log the execution result
	if err != nil {
		se.auditLogger.LogCommandFailed(req, err.Error())
	} else {
		se.auditLogger.LogCommandSuccess(req, result)
	}
	
	return result, err
}

// ExecuteWithSudo executes a command with elevated privileges
func (se *SafeExecutor) ExecuteWithSudo(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
	req.RequiresSudo = true
	return se.ExecuteCommand(ctx, req)
}

// DryRun simulates command execution without actually running it
func (se *SafeExecutor) DryRun(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
	req.DryRun = true
	return se.dryRun(ctx, req)
}

// ExecuteCompound executes multiple commands as a single operation
func (se *SafeExecutor) ExecuteCompound(ctx context.Context, operation CompoundOperation) (*CompoundResult, error) {
	result := &CompoundResult{
		OperationID:    operation.ID,
		StartTime:      time.Now(),
		FailedAt:       -1,
		CommandResults: make([]*ExecutionResult, 0, len(operation.Commands)),
	}
	
	se.logger.Info("Starting compound operation")
	
	// Execute commands in sequence
	for i, cmd := range operation.Commands {
		se.logger.Info("Executing command in sequence")
		
		cmdResult, err := se.ExecuteCommand(ctx, cmd)
		result.CommandResults = append(result.CommandResults, cmdResult)
		
		if err != nil || !cmdResult.Success {
			result.FailedAt = i
			se.logger.Error("Command failed in sequence")
			
			// Execute rollback if configured
			if len(operation.RollbackCmds) > 0 {
				se.logger.Info("Executing rollback commands...")
				result.RollbackExecuted = true
				se.executeRollback(ctx, operation.RollbackCmds, i)
			}
			
			if !operation.AllowPartial {
				result.Success = false
				result.EndTime = time.Now()
				return result, fmt.Errorf("compound operation failed at step %d: %w", i+1, err)
			}
		}
	}
	
	result.Success = true
	result.EndTime = time.Now()
	se.logger.Info("Compound operation completed successfully")
	
	return result, nil
}

// SetDryRunMode enables or disables dry run mode
func (se *SafeExecutor) SetDryRunMode(enabled bool) {
	se.dryRunMode = enabled
	se.logger.Info("Dry run mode updated")
}

// GetCommandSuggestion provides suggestions for command execution
func (se *SafeExecutor) GetCommandSuggestion(ctx context.Context, command string) (string, error) {
	resolution, err := se.commandResolver.ResolveCommand(ctx, command)
	if err != nil {
		return "", err
	}
	
	if resolution == nil {
		return fmt.Sprintf("Command '%s' status unknown", command), nil
	}
	
	return se.commandResolver.GetExecutionSuggestion(resolution), nil
}

// ResolveCommand exposes command resolution functionality
func (se *SafeExecutor) ResolveCommand(ctx context.Context, command string) (*CommandResolution, error) {
	return se.commandResolver.ResolveCommand(ctx, command)
}

// GetCommandResolver returns the command resolver for advanced usage
func (se *SafeExecutor) GetCommandResolver() *CommandResolver {
	return se.commandResolver
}

// validateCommand performs comprehensive command validation
func (se *SafeExecutor) validateCommand(req CommandRequest) error {
	// Check if command is allowed
	if !se.permissionManager.IsCommandAllowed(req.Command, req.Args) {
		return &SecurityError{
			Command: req.Command,
			Reason:  "command not in allowed list",
			Code:    "COMMAND_NOT_ALLOWED",
		}
	}
	
	// Validate command category
	if !se.isValidCategory(req.Category) {
		return &ValidationError{
			Command: req.Command,
			Reason:  fmt.Sprintf("invalid category: %s", req.Category),
			Code:    "INVALID_CATEGORY",
		}
	}
	
	// Check working directory restrictions
	if req.WorkingDir != "" && !se.permissionManager.IsDirectoryAllowed(req.WorkingDir) {
		return &SecurityError{
			Command: req.Command,
			Reason:  fmt.Sprintf("working directory not allowed: %s", req.WorkingDir),
			Code:    "DIRECTORY_NOT_ALLOWED",
		}
	}
	
	// Validate environment variables
	for key := range req.Environment {
		if !se.permissionManager.IsEnvironmentVariableAllowed(key) {
			return &SecurityError{
				Command: req.Command,
				Reason:  fmt.Sprintf("environment variable not allowed: %s", key),
				Code:    "ENV_VAR_NOT_ALLOWED",
			}
		}
	}
	
	// Check for dangerous argument patterns
	if err := se.validateArguments(req.Args); err != nil {
		return err
	}
	
	return nil
}

// executeRegular executes a command without sudo
func (se *SafeExecutor) executeRegular(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// Set up the command
	cmd := exec.CommandContext(ctx, req.Command, req.Args...)
	
	// Set working directory if specified
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	} else if se.config.DefaultWorkingDir != "" {
		cmd.Dir = se.config.DefaultWorkingDir
	}
	
	// Set environment variables
	cmd.Env = se.buildEnvironment(req.Environment)
	
	// Set timeout if specified
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	} else if se.config.MaxExecutionTime > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, se.config.MaxExecutionTime)
		defer cancel()
	}
	
	// Execute the command
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	result := &ExecutionResult{
		Command:   fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")),
		Output:    string(output),
		Duration:  duration,
		Timestamp: startTime,
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		
		// Extract exit code if available
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}
	
	return result, err
}

// executeWithSudo executes a command with elevated privileges
func (se *SafeExecutor) executeWithSudo(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
	// Convert to shared types for sudo manager
	typesReq := types.CommandRequest(req)
	result, err := se.sudoManager.ExecuteWithSudo(ctx, typesReq)
	if err != nil {
		return nil, err
	}
	// Convert back to execution types
	return (*ExecutionResult)(result), nil
}

// dryRun simulates command execution
func (se *SafeExecutor) dryRun(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
	se.logger.Info("DRY RUN: Command simulation")
	
	result := &ExecutionResult{
		Success:   true,
		ExitCode:  0,
		Output:    fmt.Sprintf("DRY RUN: Would execute: %s %s", req.Command, strings.Join(req.Args, " ")),
		Command:   fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")),
		Timestamp: time.Now(),
		DryRun:    true,
	}
	
	return result, nil
}

// requestConfirmation asks the user for confirmation
func (se *SafeExecutor) requestConfirmation(req CommandRequest) (bool, error) {
	fmt.Printf("⚠️  Command requires confirmation:\n")
	fmt.Printf("   Command: %s %s\n", req.Command, strings.Join(req.Args, " "))
	fmt.Printf("   Description: %s\n", req.Description)
	fmt.Printf("   Category: %s\n", req.Category)
	if req.RequiresSudo {
		fmt.Printf("   Requires sudo: yes\n")
	}
	fmt.Printf("\nContinue? (y/N): ")
	
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// executeRollback executes rollback commands
func (se *SafeExecutor) executeRollback(ctx context.Context, rollbackCmds []CommandRequest, failedAt int) {
	// Execute rollback commands in reverse order up to the failed command
	for i := min(failedAt, len(rollbackCmds)-1); i >= 0; i-- {
		cmd := rollbackCmds[i]
		cmd.Description = fmt.Sprintf("ROLLBACK: %s", cmd.Description)
		
		se.logger.Info("Executing rollback command")
		result, err := se.executeRegular(ctx, cmd)
		
		if err != nil || !result.Success {
			se.logger.Error("Rollback command failed")
		} else {
			se.logger.Info("Rollback command succeeded")
		}
	}
}

// validateArguments checks command arguments for dangerous patterns
func (se *SafeExecutor) validateArguments(args []string) error {
	for _, arg := range args {
		// Check for shell injection patterns
		if strings.Contains(arg, ";") || strings.Contains(arg, "&&") || strings.Contains(arg, "||") {
			return &SecurityError{
				Reason: "argument contains shell command separators",
				Code:   "SHELL_INJECTION_RISK",
			}
		}
		
		// Check for pipe operations
		if strings.Contains(arg, "|") {
			return &SecurityError{
				Reason: "argument contains pipe operations",
				Code:   "PIPE_NOT_ALLOWED",
			}
		}
		
		// Check for redirection
		if strings.Contains(arg, ">") || strings.Contains(arg, "<") {
			return &SecurityError{
				Reason: "argument contains redirection operators",
				Code:   "REDIRECTION_NOT_ALLOWED",
			}
		}
	}
	
	return nil
}

// buildEnvironment creates the environment for command execution
func (se *SafeExecutor) buildEnvironment(customEnv map[string]string) []string {
	// Start with filtered system environment
	env := []string{}
	
	// Add allowed system environment variables
	for _, allowedVar := range se.config.AllowedEnvironmentVariables {
		if value := os.Getenv(allowedVar); value != "" {
			env = append(env, fmt.Sprintf("%s=%s", allowedVar, value))
		}
	}
	
	// Add custom environment variables
	for key, value := range customEnv {
		if se.permissionManager.IsEnvironmentVariableAllowed(key) {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	return env
}

// isValidCategory checks if the command category is valid
func (se *SafeExecutor) isValidCategory(category string) bool {
	validCategories := []string{
		string(CategoryPackage),
		string(CategorySystem),
		string(CategoryConfiguration),
		string(CategoryDevelopment),
		string(CategoryUtility),
	}
	
	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	
	return false
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}