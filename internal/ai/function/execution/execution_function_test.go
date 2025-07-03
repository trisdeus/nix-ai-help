package execution

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/execution"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

func TestNewExecutionFunction(t *testing.T) {
	execFunc := NewExecutionFunction()

	if execFunc == nil {
		t.Fatal("Expected execution function to be created")
	}

	if execFunc.Name() != "execute_command" {
		t.Errorf("Expected function name 'execute_command', got '%s'", execFunc.Name())
	}

	if execFunc.Description() == "" {
		t.Error("Expected function description to be set")
	}

	if !execFunc.IsInitialized() {
		t.Log("Function is not initialized (expected - requires dependency injection)")
	}
}

func TestExecutionFunctionSchema(t *testing.T) {
	execFunc := NewExecutionFunction()
	schema := execFunc.Schema()

	if schema.Name != "execute_command" {
		t.Errorf("Expected schema name 'execute_command', got '%s'", schema.Name)
	}

	// Check required parameters
	requiredParams := []string{"command", "description", "category"}
	for _, required := range requiredParams {
		found := false
		for _, param := range schema.Parameters {
			if param.Name == required && param.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required parameter '%s' not found in schema", required)
		}
	}

	// Check optional parameters
	optionalParams := []string{"args", "requiresSudo", "workingDir", "environment", "dryRun", "timeout"}
	for _, optional := range optionalParams {
		found := false
		for _, param := range schema.Parameters {
			if param.Name == optional && !param.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Optional parameter '%s' not found in schema", optional)
		}
	}

	// Check examples
	if len(schema.Examples) == 0 {
		t.Error("Expected function examples to be provided")
	}
}

func TestExecutionFunctionValidation(t *testing.T) {
	execFunc := NewExecutionFunction()

	tests := []struct {
		name      string
		params    map[string]interface{}
		expectErr bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"command":     "nix-env",
				"args":        []interface{}{"-iA", "nixpkgs.firefox"},
				"description": "Install Firefox browser",
				"category":    "package",
			},
			expectErr: false,
		},
		{
			name: "missing command",
			params: map[string]interface{}{
				"description": "Some description",
				"category":    "package",
			},
			expectErr: true,
		},
		{
			name: "missing description",
			params: map[string]interface{}{
				"command":  "nix-env",
				"category": "package",
			},
			expectErr: true,
		},
		{
			name: "missing category",
			params: map[string]interface{}{
				"command":     "nix-env",
				"description": "Some description",
			},
			expectErr: true,
		},
		{
			name: "invalid category",
			params: map[string]interface{}{
				"command":     "nix-env",
				"description": "Some description",
				"category":    "invalid_category",
			},
			expectErr: true,
		},
		{
			name: "valid with optional parameters",
			params: map[string]interface{}{
				"command":      "nixos-rebuild",
				"args":         []interface{}{"switch"},
				"description":  "Rebuild NixOS",
				"category":     "system",
				"requiresSudo": true,
				"workingDir":   "/etc/nixos",
				"environment":  map[string]interface{}{"NIX_PATH": "/custom/path"},
				"dryRun":       false,
				"timeout":      "5m",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := execFunc.ValidateParameters(tt.params)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateParameters() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestExecutionFunctionInitialization(t *testing.T) {
	execFunc := NewExecutionFunction()

	// Function should not be initialized initially
	if execFunc.IsInitialized() {
		t.Error("Function should not be initialized before calling Initialize()")
	}

	// Create mock security components
	log := logger.NewLogger()
	config := GetDefaultConfig()
	permissionManager := security.NewPermissionManager(config, log)
	
	auditLogger, err := security.NewAuditLogger("/tmp/test_audit.log", false, log) // Disabled for tests
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	sudoConfig := &security.SudoConfig{
		SessionTimeout:    time.Hour,
		PasswordTimeout:   time.Minute * 15,
		MaxAttempts:       3,
		RequirePassword:   true,
		AllowPasswordless: false,
		PreserveEnv:       []string{"PATH", "HOME"},
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, log)

	// Initialize the function
	err = execFunc.Initialize(permissionManager, auditLogger, sudoManager, config)
	if err != nil {
		t.Fatalf("Failed to initialize execution function: %v", err)
	}

	// Function should now be initialized
	if !execFunc.IsInitialized() {
		t.Error("Function should be initialized after calling Initialize()")
	}

	// Test initialization with nil dependencies
	execFunc2 := NewExecutionFunction()
	err = execFunc2.Initialize(nil, auditLogger, sudoManager, config)
	if err == nil {
		t.Error("Expected error when initializing with nil permission manager")
	}
}

func TestExecutionFunctionConfiguration(t *testing.T) {
	execFunc := NewExecutionFunction()

	// Test default configuration
	execFunc.SetInteractive(false)
	execFunc.SetMaxRetries(5)
	execFunc.SetRetryDelay(time.Second * 3)

	// These should not cause errors (methods exist and work)
	// We can't easily test the internal state without exposing getters
	// but we can verify the methods exist and don't panic
}

func TestExecutionFunctionWithoutInitialization(t *testing.T) {
	execFunc := NewExecutionFunction()
	ctx := context.Background()

	params := map[string]interface{}{
		"command":     "echo",
		"args":        []string{"hello"},
		"description": "Test command",
		"category":    "utility",
	}

	options := &functionbase.FunctionOptions{
		Timeout: time.Minute,
	}

	// Should fail because function is not initialized
	result, err := execFunc.Execute(ctx, params, options)
	if err != nil {
		t.Fatalf("Execute() should not return error, got: %v", err)
	}

	if result.Success {
		t.Error("Execute() should fail when function is not initialized")
	}

	if result.Error == "" {
		t.Error("Execute() should return error message when not initialized")
	}
}

func TestParseExecutionRequest(t *testing.T) {
	execFunc := NewExecutionFunction()

	tests := []struct {
		name      string
		params    map[string]interface{}
		expectErr bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"command":     "nix-env",
				"args":        []interface{}{"--install", "firefox"},
				"description": "Install Firefox",
				"category":    "package",
				"dryRun":      true,
			},
			expectErr: false,
		},
		{
			name: "invalid JSON",
			params: map[string]interface{}{
				"command": make(chan int), // This will cause JSON marshal to fail
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := execFunc.parseRequest(tt.params)
			if (err != nil) != tt.expectErr {
				t.Errorf("parseRequest() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestCreateExecutionRequest(t *testing.T) {
	execFunc := NewExecutionFunction()

	tests := []struct {
		name      string
		request   *ExecutionRequest
		expectErr bool
	}{
		{
			name: "valid request",
			request: &ExecutionRequest{
				Command:     "nix-env",
				Args:        []string{"-iA", "nixpkgs.firefox"},
				Description: "Install Firefox",
				Category:    "package",
				Timeout:     "5m",
			},
			expectErr: false,
		},
		{
			name: "invalid timeout",
			request: &ExecutionRequest{
				Command:     "nix-env",
				Description: "Install Firefox",
				Category:    "package",
				Timeout:     "invalid_duration",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := execFunc.createExecutionRequest(tt.request)
			if (err != nil) != tt.expectErr {
				t.Errorf("createExecutionRequest() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestRetryableErrorDetection(t *testing.T) {
	execFunc := NewExecutionFunction()

	tests := []struct {
		name        string
		errMsg      string
		expectRetry bool
	}{
		{
			name:        "connection refused",
			errMsg:      "connection refused",
			expectRetry: true,
		},
		{
			name:        "timeout error",
			errMsg:      "operation timeout",
			expectRetry: true,
		},
		{
			name:        "temporary failure",
			errMsg:      "temporary failure in name resolution",
			expectRetry: true,
		},
		{
			name:        "permission denied",
			errMsg:      "permission denied",
			expectRetry: false,
		},
		{
			name:        "not allowed",
			errMsg:      "command not allowed",
			expectRetry: false,
		},
		{
			name:        "forbidden",
			errMsg:      "forbidden operation",
			expectRetry: false,
		},
		{
			name:        "resource unavailable",
			errMsg:      "resource temporarily unavailable",
			expectRetry: true,
		},
		{
			name:        "unknown error",
			errMsg:      "some unknown error",
			expectRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &TestError{message: tt.errMsg}
			result := execFunc.isRetryableError(err)
			if result != tt.expectRetry {
				t.Errorf("isRetryableError() = %v, expectRetry %v", result, tt.expectRetry)
			}
		})
	}
}

// TestError is a simple error implementation for testing
type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

func TestCreateResponseMethods(t *testing.T) {
	execFunc := NewExecutionFunction()

	// Test success response creation
	mockResult := &execution.ExecutionResult{
		Success:   true,
		ExitCode:  0,
		Output:    "command completed successfully",
		Duration:  time.Second,
		Command:   "echo hello",
		Timestamp: time.Now(),
		DryRun:    false,
	}

	mockRequest := &ExecutionRequest{
		Command:     "echo",
		Args:        []string{"hello"},
		Description: "Test echo command",
		Category:    "utility",
	}

	successResponse := execFunc.createSuccessResponse(mockResult, mockRequest)
	if !successResponse.Success {
		t.Error("Expected success response to have Success=true")
	}
	if successResponse.Category != mockRequest.Category {
		t.Errorf("Expected category %s, got %s", mockRequest.Category, successResponse.Category)
	}

	// Test error response creation
	testErr := &TestError{message: "test error"}
	errorResponse := execFunc.createErrorResponse(mockRequest, testErr)
	if errorResponse.Success {
		t.Error("Expected error response to have Success=false")
	}
	if errorResponse.ExitCode != -1 {
		t.Errorf("Expected exit code -1, got %d", errorResponse.ExitCode)
	}
	if errorResponse.Error != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", errorResponse.Error)
	}
}

// Remove the mock type and use the actual execution.ExecutionResult

func BenchmarkExecutionFunctionValidation(b *testing.B) {
	execFunc := NewExecutionFunction()
	params := map[string]interface{}{
		"command":     "nix-env",
		"args":        []string{"-iA", "nixpkgs.firefox"},
		"description": "Install Firefox browser",
		"category":    "package",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = execFunc.ValidateParameters(params)
	}
}

func BenchmarkParseExecutionRequest(b *testing.B) {
	execFunc := NewExecutionFunction()
	params := map[string]interface{}{
		"command":     "nix-env",
		"args":        []interface{}{"--install", "firefox"},
		"description": "Install Firefox",
		"category":    "package",
		"dryRun":      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = execFunc.parseRequest(params)
	}
}