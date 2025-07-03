package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/execution"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

// ExecutionFunction provides AI-controlled command execution
type ExecutionFunction struct {
	*functionbase.BaseFunction
	executor    *execution.SafeExecutor
	logger      *logger.Logger
	interactive bool
	maxRetries  int
	retryDelay  time.Duration
}

// ExecutionRequest represents the input parameters for command execution
type ExecutionRequest struct {
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

// ExecutionResponse represents the result of command execution
type ExecutionResponse struct {
	Success     bool                   `json:"success"`
	ExitCode    int                    `json:"exitCode"`
	Output      string                 `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    string                 `json:"duration"`
	Command     string                 `json:"command"`
	DryRun      bool                   `json:"dryRun"`
	Timestamp   string                 `json:"timestamp"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewExecutionFunction creates a new execution function without dependencies
func NewExecutionFunction() *ExecutionFunction {
	// Define function parameters
	parameters := []functionbase.FunctionParameter{
		functionbase.StringParam("command", "The command to execute (e.g., 'nix', 'nixos-rebuild', 'systemctl')", true),
		functionbase.ArrayParam("args", "Command arguments as an array of strings", false),
		functionbase.StringParam("description", "Human-readable description of what this command does", true),
		functionbase.StringParamWithEnum("category", "Command category for permission validation", true, 
			[]string{"package", "system", "configuration", "development", "utility"}),
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

	// Add examples to the schema
	schema := baseFunc.Schema()
	schema.Examples = []functionbase.FunctionExample{
		{
			Description: "Install a package with nix-env",
			Parameters: map[string]interface{}{
				"command":     "nix-env",
				"args":        []string{"-iA", "nixpkgs.firefox"},
				"description": "Install Firefox browser using nix-env",
				"category":    "package",
			},
			Expected: "Package installation output and success status",
		},
		{
			Description: "Rebuild NixOS configuration",
			Parameters: map[string]interface{}{
				"command":      "nixos-rebuild",
				"args":         []string{"switch"},
				"description":  "Apply NixOS configuration changes",
				"category":     "system",
				"requiresSudo": true,
			},
			Expected: "System rebuild output and status",
		},
		{
			Description: "Check system status (dry run)",
			Parameters: map[string]interface{}{
				"command":     "systemctl",
				"args":        []string{"status", "sshd"},
				"description": "Check SSH daemon status",
				"category":    "utility",
				"dryRun":      true,
			},
			Expected: "Dry run output showing what would be executed",
		},
	}
	baseFunc.SetSchema(schema)

	return &ExecutionFunction{
		BaseFunction: baseFunc,
		logger:       logger.NewLogger(),
		interactive:  true,
		maxRetries:   3,
		retryDelay:   time.Second * 2,
	}
}

// Initialize sets up the execution function with required dependencies
func (ef *ExecutionFunction) Initialize(
	permissionManager *security.PermissionManager,
	auditLogger *security.AuditLogger,
	sudoManager *security.SudoManager,
	config *config.ExecutionConfig,
) error {
	if permissionManager == nil || auditLogger == nil || sudoManager == nil || config == nil {
		return fmt.Errorf("all dependencies are required for execution function")
	}

	ef.executor = execution.NewSafeExecutor(
		permissionManager,
		auditLogger,
		sudoManager,
		config,
		ef.logger,
	)

	ef.logger.Info("Execution function initialized with security components")
	return nil
}

// Execute implements the FunctionInterface
func (ef *ExecutionFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
	ef.logger.Debug("Starting execution function")
	startTime := time.Now()

	// Check if the function is properly initialized
	if ef.executor == nil {
		return functionbase.CreateErrorResult(
			fmt.Errorf("execution function not initialized"),
			"Function requires initialization with security components",
		), nil
	}

	// Report progress if callback is available
	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    1,
			Total:      5,
			Percentage: 20,
			Message:    "Parsing command parameters",
			Stage:      "preparation",
		})
	}

	// Parse parameters into structured request
	request, err := ef.parseRequest(params)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to parse request parameters"), nil
	}

	// Validate required parameters
	if err := ef.validateRequest(request); err != nil {
		return functionbase.CreateErrorResult(err, "Parameter validation failed"), nil
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    2,
			Total:      5,
			Percentage: 40,
			Message:    "Creating execution request",
			Stage:      "validation",
		})
	}

	// Create execution request
	execReq, err := ef.createExecutionRequest(request)
	if err != nil {
		return functionbase.CreateErrorResult(err, "Failed to create execution request"), nil
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    3,
			Total:      5,
			Percentage: 60,
			Message:    fmt.Sprintf("Executing: %s %s", request.Command, strings.Join(request.Args, " ")),
			Stage:      "execution",
		})
	}

	// Execute command with context and timeout
	execCtx := ctx
	if request.Timeout != "" {
		if timeout, err := time.ParseDuration(request.Timeout); err == nil {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Execute with retries for transient failures
	var lastErr error
	var result *execution.ExecutionResult
	for attempt := 1; attempt <= ef.maxRetries; attempt++ {
		result, err = ef.executor.ExecuteCommand(execCtx, *execReq)

		if err == nil {
			// Success
			break
		}

		lastErr = err

		// Check if error is retryable
		if !ef.isRetryableError(err) || attempt == ef.maxRetries {
			break
		}

		ef.logger.Debug(fmt.Sprintf("Command execution failed (attempt %d/%d), retrying", attempt, ef.maxRetries))
		time.Sleep(ef.retryDelay)
	}

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    4,
			Total:      5,
			Percentage: 80,
			Message:    "Processing execution result",
			Stage:      "processing",
		})
	}

	// Handle execution result
	if lastErr != nil {
		response := ef.createErrorResponse(request, lastErr)
		if options != nil && options.ProgressCallback != nil {
			options.ProgressCallback(functionbase.Progress{
				Current:    5,
				Total:      5,
				Percentage: 100,
				Message:    "Execution failed",
				Stage:      "complete",
			})
		}
		return functionbase.CreateSuccessResult(response, "Command execution completed with errors"), nil
	}

	// Create success response
	response := ef.createSuccessResponse(result, request)

	if options != nil && options.ProgressCallback != nil {
		options.ProgressCallback(functionbase.Progress{
			Current:    5,
			Total:      5,
			Percentage: 100,
			Message:    "Execution completed successfully",
			Stage:      "complete",
		})
	}

	ef.logger.Debug("Execution function completed successfully")
	
	return &functionbase.FunctionResult{
		Success:   true,
		Data:      response,
		Duration:  time.Since(startTime),
		Timestamp: time.Now(),
	}, nil
}

// parseRequest converts raw parameters to structured ExecutionRequest
func (ef *ExecutionFunction) parseRequest(params map[string]interface{}) (*ExecutionRequest, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var request ExecutionRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	return &request, nil
}

// validateRequest validates the execution request
func (ef *ExecutionFunction) validateRequest(request *ExecutionRequest) error {
	if request.Command == "" {
		return fmt.Errorf("command is required")
	}

	if request.Description == "" {
		return fmt.Errorf("description is required")
	}

	if request.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate category
	validCategories := []string{"package", "system", "configuration", "development", "utility"}
	valid := false
	for _, cat := range validCategories {
		if request.Category == cat {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid category: %s", request.Category)
	}

	return nil
}

// createExecutionRequest creates an execution request from parameters
func (ef *ExecutionFunction) createExecutionRequest(request *ExecutionRequest) (*execution.CommandRequest, error) {
	execReq := &execution.CommandRequest{
		Command:      request.Command,
		Args:         request.Args,
		RequiresSudo: request.RequiresSudo,
		WorkingDir:   request.WorkingDir,
		Environment:  request.Environment,
		Description:  request.Description,
		Category:     request.Category,
		DryRun:       request.DryRun,
	}

	// Parse timeout if provided
	if request.Timeout != "" {
		timeout, err := time.ParseDuration(request.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %s", request.Timeout)
		}
		execReq.Timeout = timeout
	}

	return execReq, nil
}

// createSuccessResponse creates a success response from execution result
func (ef *ExecutionFunction) createSuccessResponse(result *execution.ExecutionResult, request *ExecutionRequest) *ExecutionResponse {
	return &ExecutionResponse{
		Success:     result.Success,
		ExitCode:    result.ExitCode,
		Output:      result.Output,
		Error:       result.Error,
		Duration:    result.Duration.String(),
		Command:     result.Command,
		DryRun:      result.DryRun,
		Timestamp:   result.Timestamp.Format(time.RFC3339),
		Category:    request.Category,
		Description: request.Description,
		Metadata: map[string]interface{}{
			"retryCount":      0,
			"executionTime":   result.Duration.Milliseconds(),
			"commandCategory": request.Category,
		},
	}
}

// createErrorResponse creates an error response
func (ef *ExecutionFunction) createErrorResponse(request *ExecutionRequest, err error) *ExecutionResponse {
	return &ExecutionResponse{
		Success:     false,
		ExitCode:    -1,
		Error:       err.Error(),
		Command:     fmt.Sprintf("%s %s", request.Command, strings.Join(request.Args, " ")),
		Timestamp:   time.Now().Format(time.RFC3339),
		Category:    request.Category,
		Description: request.Description,
		Metadata: map[string]interface{}{
			"errorType": "execution_error",
		},
	}
}

// isRetryableError determines if an error is retryable
func (ef *ExecutionFunction) isRetryableError(err error) bool {
	errorStr := strings.ToLower(err.Error())

	// Network-related errors
	if strings.Contains(errorStr, "connection refused") ||
		strings.Contains(errorStr, "timeout") ||
		strings.Contains(errorStr, "temporary failure") {
		return true
	}

	// Resource-related errors
	if strings.Contains(errorStr, "resource temporarily unavailable") ||
		strings.Contains(errorStr, "too many open files") {
		return true
	}

	// Security errors are generally not retryable
	if strings.Contains(errorStr, "permission denied") ||
		strings.Contains(errorStr, "not allowed") ||
		strings.Contains(errorStr, "forbidden") {
		return false
	}

	return false
}

// SetInteractive sets whether the function should use interactive mode
func (ef *ExecutionFunction) SetInteractive(interactive bool) {
	ef.interactive = interactive
}

// SetMaxRetries sets the maximum number of retries for failed commands
func (ef *ExecutionFunction) SetMaxRetries(maxRetries int) {
	if maxRetries >= 0 {
		ef.maxRetries = maxRetries
	}
}

// SetRetryDelay sets the delay between retries
func (ef *ExecutionFunction) SetRetryDelay(delay time.Duration) {
	if delay > 0 {
		ef.retryDelay = delay
	}
}

// GetExecutor returns the underlying safe executor (for testing)
func (ef *ExecutionFunction) GetExecutor() *execution.SafeExecutor {
	return ef.executor
}

// IsInitialized returns true if the function has been initialized with dependencies
func (ef *ExecutionFunction) IsInitialized() bool {
	return ef.executor != nil
}