package function

import (
	"os"
	"testing"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

func TestExecuteCommandFunction_GetDefinition(t *testing.T) {
	// Create test configuration
	execConfig := &config.ExecutionConfig{
		Enabled:              true,
		ConfirmationRequired: false, // Disable for testing
		DryRunDefault:       true,   // Use dry run for testing
		MaxExecutionTime:    time.Minute * 5,
		AllowedCommands:     []string{"echo", "ls", "cat"},
		ForbiddenCommands:   []string{"rm", "dd"},
	}
	
	// Create test components
	logger := logger.NewLogger()
	permissionManager := security.NewPermissionManager(execConfig, logger)
	auditLogger, _ := security.NewAuditLogger("/tmp/test-audit.log", false, logger)
	sudoConfig := &security.SudoConfig{
		SessionTimeout:  time.Minute * 30,
		PasswordTimeout: time.Minute * 15,
		MaxAttempts:     3,
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, logger)
	
	// Create function
	function := NewExecuteCommandFunction(permissionManager, auditLogger, sudoManager, execConfig, logger)
	
	// Test that function was created successfully
	if function == nil {
		t.Fatal("Expected function to be created")
	}
	
	// Verify function can validate parameters (smoke test)
	testParams := map[string]interface{}{
		"command":     "echo",
		"args":        []interface{}{"hello"},
		"description": "Test echo command",
		"category":    "utility",
	}
	
	err := function.ValidateParameters(testParams)
	if err != nil {
		t.Errorf("Expected parameter validation to pass, got error: %v", err)
	}
}

func TestExecuteCommandFunction_Call_DryRun(t *testing.T) {
	// Create test configuration
	execConfig := &config.ExecutionConfig{
		Enabled:              true,
		ConfirmationRequired: false,
		DryRunDefault:       true,
		MaxExecutionTime:    time.Minute * 5,
		AllowedCommands:     []string{"echo", "ls", "cat"},
	}
	
	// Create test components
	logger := logger.NewLogger()
	permissionManager := security.NewPermissionManager(execConfig, logger)
	auditLogger, _ := security.NewAuditLogger("/tmp/test-audit.log", false, logger)
	sudoConfig := &security.SudoConfig{
		SessionTimeout:  time.Minute * 30,
		PasswordTimeout: time.Minute * 15,
		MaxAttempts:     3,
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, logger)
	
	// Create function
	function := NewExecuteCommandFunction(permissionManager, auditLogger, sudoManager, execConfig, logger)
	
	// Test parameters
	params := map[string]interface{}{
		"command":     "echo",
		"args":        []string{"hello", "world"},
		"description": "Test echo command",
		"category":    "utility",
		"dryRun":      true,
	}
	
	// Call function
	result, err := function.Call(params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Verify result
	execResult, ok := result.(*ExecuteCommandResult)
	if !ok {
		t.Fatal("Expected ExecuteCommandResult")
	}
	
	if !execResult.Success {
		t.Errorf("Expected success=true, got success=%v, error=%s", execResult.Success, execResult.Error)
	}
	
	if !execResult.DryRun {
		t.Error("Expected dryRun=true")
	}
	
	if execResult.Output == "" {
		t.Error("Expected non-empty output")
	}
}

func TestExecuteCommandFunction_Call_InvalidParameters(t *testing.T) {
	// Create test configuration
	execConfig := &config.ExecutionConfig{
		Enabled:          true,
		AllowedCommands: []string{"echo"},
	}
	
	// Create test components
	logger := logger.NewLogger()
	permissionManager := security.NewPermissionManager(execConfig, logger)
	auditLogger, _ := security.NewAuditLogger("/tmp/test-audit.log", false, logger)
	sudoConfig := &security.SudoConfig{}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, logger)
	
	// Create function
	function := NewExecuteCommandFunction(permissionManager, auditLogger, sudoManager, execConfig, logger)
	
	testCases := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name: "missing command",
			params: map[string]interface{}{
				"description": "Test",
				"category":    "utility",
			},
		},
		{
			name: "missing description",
			params: map[string]interface{}{
				"command":  "echo",
				"category": "utility",
			},
		},
		{
			name: "missing category",
			params: map[string]interface{}{
				"command":     "echo",
				"description": "Test",
			},
		},
		{
			name: "invalid category",
			params: map[string]interface{}{
				"command":     "echo",
				"description": "Test",
				"category":    "invalid",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := function.Call(tc.params)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			execResult, ok := result.(*ExecuteCommandResult)
			if !ok {
				t.Fatal("Expected ExecuteCommandResult")
			}
			
			if execResult.Success {
				t.Error("Expected success=false for invalid parameters")
			}
			
			if execResult.Error == "" {
				t.Error("Expected non-empty error message")
			}
		})
	}
}

func TestExecuteCommandFunction_Call_ForbiddenCommand(t *testing.T) {
	// Create test configuration
	execConfig := &config.ExecutionConfig{
		Enabled:           true,
		ConfirmationRequired: false,
		AllowedCommands:   []string{"echo"},
		ForbiddenCommands: []string{"rm"},
	}
	
	// Create test components
	logger := logger.NewLogger()
	permissionManager := security.NewPermissionManager(execConfig, logger)
	auditLogger, _ := security.NewAuditLogger("/tmp/test-audit.log", false, logger)
	sudoConfig := &security.SudoConfig{}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, logger)
	
	// Create function
	function := NewExecuteCommandFunction(permissionManager, auditLogger, sudoManager, execConfig, logger)
	
	// Test forbidden command
	params := map[string]interface{}{
		"command":     "rm",
		"args":        []string{"-rf", "/"},
		"description": "Dangerous command",
		"category":    "system",
	}
	
	result, err := function.Call(params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	execResult, ok := result.(*ExecuteCommandResult)
	if !ok {
		t.Fatal("Expected ExecuteCommandResult")
	}
	
	if execResult.Success {
		t.Error("Expected success=false for forbidden command")
	}
	
	if execResult.Error == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestExecuteCommandFunction_SetConfiguration(t *testing.T) {
	// Create test configuration
	execConfig := &config.ExecutionConfig{
		Enabled: true,
	}
	
	// Create test components
	logger := logger.NewLogger()
	permissionManager := security.NewPermissionManager(execConfig, logger)
	auditLogger, _ := security.NewAuditLogger("/tmp/test-audit.log", false, logger)
	sudoConfig := &security.SudoConfig{}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, logger)
	
	// Create function
	function := NewExecuteCommandFunction(permissionManager, auditLogger, sudoManager, execConfig, logger)
	
	// Test configuration setters
	function.SetInteractive(false)
	function.SetMaxRetries(5)
	function.SetRetryDelay(time.Second * 3)
	
	// Verify by checking that these don't panic and the function still works
	if function.maxRetries != 5 {
		t.Errorf("Expected maxRetries=5, got %d", function.maxRetries)
	}
	
	if function.retryDelay != time.Second*3 {
		t.Errorf("Expected retryDelay=3s, got %v", function.retryDelay)
	}
}

func TestExecuteCommandFunction_parseParameters(t *testing.T) {
	// Create minimal function for testing
	function := &ExecuteCommandFunction{}
	
	testCases := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"command":     "echo",
				"args":        []string{"hello"},
				"description": "Test command",
				"category":    "utility",
			},
			expectError: false,
		},
		{
			name: "parameters with timeout",
			params: map[string]interface{}{
				"command":     "sleep",
				"args":        []string{"1"},
				"description": "Test sleep",
				"category":    "utility",
				"timeout":     "5s",
			},
			expectError: false,
		},
		{
			name: "invalid parameters structure",
			params: map[string]interface{}{
				"command": map[string]interface{}{"invalid": "structure"},
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := function.parseParameters(tc.params)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil result")
				}
			}
		})
	}
}

// Cleanup function for tests
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	
	// Cleanup test files
	os.Remove("/tmp/test-audit.log")
	
	os.Exit(code)
}