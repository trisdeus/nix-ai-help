package function

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/execution"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

// ExecuteCommandFunction provides AI-controlled command execution
type ExecuteCommandFunction struct {
	*functionbase.BaseFunction
	executor      *execution.SafeExecutor
	logger        *logger.Logger
	interactive   bool
	maxRetries    int
	retryDelay    time.Duration
}

// ExecuteCommandParams represents parameters for command execution
type ExecuteCommandParams struct {
	Command      string            `json:"command"`
	Args         []string          `json:"args,omitempty"`
	Description  string            `json:"description"`
	Category     string            `json:"category"`
	RequiresSudo bool              `json:"requiresSudo,omitempty"`
	WorkingDir   string            `json:"workingDir,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	DryRun       bool              `json:"dryRun,omitempty"`
	Timeout      string            `json:"timeout,omitempty"`
}

// ExecuteCommandResult represents the result of command execution
type ExecuteCommandResult struct {
	Success     bool              `json:"success"`
	ExitCode    int               `json:"exitCode"`
	Output      string            `json:"output"`
	Error       string            `json:"error,omitempty"`
	Duration    string            `json:"duration"`
	Command     string            `json:"command"`
	DryRun      bool              `json:"dryRun"`
	Timestamp   string            `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewExecuteCommandFunction creates a new execute command function
func NewExecuteCommandFunction(
	permissionManager *security.PermissionManager,
	auditLogger *security.AuditLogger,
	sudoManager *security.SudoManager,
	config *config.ExecutionConfig,
	logger *logger.Logger,
) *ExecuteCommandFunction {
	executor := execution.NewSafeExecutor(
		permissionManager,
		auditLogger,
		sudoManager,
		config,
		logger,
	)
	
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("command", "The command to execute (e.g., 'nix', 'nixos-rebuild', 'systemctl')", true),
		functionbase.ArrayParam("args", "Command arguments as an array of strings", false),
		functionbase.StringParam("description", "Human-readable description of what this command does", true),
		functionbase.StringParamWithEnum("category", "Command category for permission validation", true, []string{"package", "system", "configuration", "development", "utility"}),
		functionbase.BoolParam("requiresSudo", "Whether this command requires sudo privileges", false, false),
		functionbase.StringParam("workingDir", "Working directory for command execution", false),
		functionbase.ObjectParam("environment", "Environment variables for command execution", false),
		functionbase.BoolParam("dryRun", "If true, show what would be executed without running", false, false),
		functionbase.StringParam("timeout", "Execution timeout (e.g., '5m', '30s')", false),
	}
	
	baseFunc := functionbase.NewBaseFunction(
		"execute_command",
		"Execute system commands safely with permission validation and security checks",
		parameters,
	)
	
	return &ExecuteCommandFunction{
		BaseFunction: baseFunc,
		executor:     executor,
		logger:       logger,
		interactive:  true,
		maxRetries:   3,
		retryDelay:   time.Second * 2,
	}
}

// Execute implements the FunctionInterface
func (ecf *ExecuteCommandFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	startTime := time.Now()
	
	// Validate parameters
	if err := ecf.ValidateParameters(params); err != nil {
		return functionbase.ErrorResult(err, time.Since(startTime)), nil
	}
	
	// Call the original implementation
	result, err := ecf.Call(params)
	if err != nil {
		return functionbase.ErrorResult(err, time.Since(startTime)), nil
	}
	
	return functionbase.SuccessResult(result, time.Since(startTime)), nil
}

// Call executes the command with the given parameters (internal implementation)
func (ecf *ExecuteCommandFunction) Call(params map[string]interface{}) (interface{}, error) {
	// Parse parameters
	cmdParams, err := ecf.parseParameters(params)
	if err != nil {
		return ecf.createErrorResult("Parameter parsing failed", err), nil
	}
	
	// Validate command parameters
	if err := ecf.validateParameters(cmdParams); err != nil {
		return ecf.createErrorResult("Parameter validation failed", err), nil
	}
	
	// Create execution request
	req, err := ecf.createExecutionRequest(cmdParams)
	if err != nil {
		return ecf.createErrorResult("Failed to create execution request", err), nil
	}
	
	// Execute command with context and timeout
	ctx := context.Background()
	if cmdParams.Timeout != "" {
		if timeout, err := time.ParseDuration(cmdParams.Timeout); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}
	
	// Execute with retries for transient failures
	var lastErr error
	for attempt := 1; attempt <= ecf.maxRetries; attempt++ {
		result, err := ecf.executor.ExecuteCommand(ctx, *req)
		
		if err == nil {
			// Success - return result
			return ecf.createSuccessResult(result, cmdParams), nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !ecf.isRetryableError(err) || attempt == ecf.maxRetries {
			break
		}
		
		ecf.logger.Debug("Command execution failed, retrying")
		time.Sleep(ecf.retryDelay)
	}
	
	return ecf.createErrorResult("Command execution failed", lastErr), nil
}

// parseParameters parses the raw parameters into structured format
func (ecf *ExecuteCommandFunction) parseParameters(params map[string]interface{}) (*ExecuteCommandParams, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}
	
	var cmdParams ExecuteCommandParams
	if err := json.Unmarshal(data, &cmdParams); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}
	
	return &cmdParams, nil
}

// validateParameters validates the command parameters
func (ecf *ExecuteCommandFunction) validateParameters(params *ExecuteCommandParams) error {
	if params.Command == "" {
		return fmt.Errorf("command is required")
	}
	
	if params.Description == "" {
		return fmt.Errorf("description is required")
	}
	
	if params.Category == "" {
		return fmt.Errorf("category is required")
	}
	
	// Validate category
	validCategories := []string{"package", "system", "configuration", "development", "utility"}
	valid := false
	for _, cat := range validCategories {
		if params.Category == cat {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid category: %s", params.Category)
	}
	
	return nil
}

// createExecutionRequest creates an execution request from parameters
func (ecf *ExecuteCommandFunction) createExecutionRequest(params *ExecuteCommandParams) (*execution.CommandRequest, error) {
	req := &execution.CommandRequest{
		Command:      params.Command,
		Args:         params.Args,
		RequiresSudo: params.RequiresSudo,
		WorkingDir:   params.WorkingDir,
		Environment:  params.Environment,
		Description:  params.Description,
		Category:     params.Category,
		DryRun:       params.DryRun,
	}
	
	// Parse timeout if provided
	if params.Timeout != "" {
		timeout, err := time.ParseDuration(params.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %s", params.Timeout)
		}
		req.Timeout = timeout
	}
	
	return req, nil
}

// createSuccessResult creates a success result from execution result
func (ecf *ExecuteCommandFunction) createSuccessResult(result *execution.ExecutionResult, params *ExecuteCommandParams) *ExecuteCommandResult {
	return &ExecuteCommandResult{
		Success:   result.Success,
		ExitCode:  result.ExitCode,
		Output:    result.Output,
		Error:     result.Error,
		Duration:  result.Duration.String(),
		Command:   result.Command,
		DryRun:    result.DryRun,
		Timestamp: result.Timestamp.Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"category":    params.Category,
			"description": params.Description,
			"retryCount":  0,
		},
	}
}

// createErrorResult creates an error result
func (ecf *ExecuteCommandFunction) createErrorResult(message string, err error) *ExecuteCommandResult {
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}
	
	return &ExecuteCommandResult{
		Success:   false,
		ExitCode:  -1,
		Error:     errorMsg,
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata: map[string]interface{}{
			"errorType": "function_error",
		},
	}
}

// isRetryableError determines if an error is retryable
func (ecf *ExecuteCommandFunction) isRetryableError(err error) bool {
	// Check for specific error types that are retryable
	errorStr := err.Error()
	
	// Network-related errors
	if contains(errorStr, "connection refused") ||
		contains(errorStr, "timeout") ||
		contains(errorStr, "temporary failure") {
		return true
	}
	
	// Resource-related errors
	if contains(errorStr, "resource temporarily unavailable") ||
		contains(errorStr, "too many open files") {
		return true
	}
	
	// Security errors are generally not retryable
	if contains(errorStr, "permission denied") ||
		contains(errorStr, "not allowed") ||
		contains(errorStr, "forbidden") {
		return false
	}
	
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 findSubstring(s, substr)))
}

// findSubstring performs a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SetInteractive sets whether the function should use interactive mode
func (ecf *ExecuteCommandFunction) SetInteractive(interactive bool) {
	ecf.interactive = interactive
}

// SetMaxRetries sets the maximum number of retries for failed commands
func (ecf *ExecuteCommandFunction) SetMaxRetries(maxRetries int) {
	if maxRetries >= 0 {
		ecf.maxRetries = maxRetries
	}
}

// SetRetryDelay sets the delay between retries
func (ecf *ExecuteCommandFunction) SetRetryDelay(delay time.Duration) {
	if delay > 0 {
		ecf.retryDelay = delay
	}
}

// GetExecutor returns the underlying safe executor (for testing)
func (ecf *ExecuteCommandFunction) GetExecutor() *execution.SafeExecutor {
	return ecf.executor
}