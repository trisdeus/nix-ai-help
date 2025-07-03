package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"nix-ai-help/internal/ai/function/execution"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/internal/config"
	execTypes "nix-ai-help/internal/execution"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

// TestEndToEndExecution tests the complete execution pipeline
func TestEndToEndExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := logger.NewLogger()
	
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	auditLogPath := tmpDir + "/audit.log"

	// Setup execution configuration
	config := &config.ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        true, // Use dry run for tests
		ConfirmationRequired: false, // Disable confirmation for tests
		MaxExecutionTime:     time.Minute,
		AllowedCommands:      []string{"echo", "ls", "cat", "true", "false"},
		ForbiddenCommands:    []string{"rm", "dd"},
		SudoCommands:         []string{},
		AllowedDirectories:   []string{tmpDir, "/tmp"},
		ForbiddenPaths:       []string{"/boot", "/dev"},
		AllowedEnvironmentVariables: []string{"PATH", "HOME", "TEST_VAR"},
	}

	// Create security components
	permissionManager := security.NewPermissionManager(config, log)
	
	auditLogger, err := security.NewAuditLogger(auditLogPath, true, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	sudoConfig := &security.SudoConfig{
		SessionTimeout:    time.Hour,
		PasswordTimeout:   time.Minute * 15,
		MaxAttempts:       3,
		RequirePassword:   false, // Disable for tests
		AllowPasswordless: true,
		PreserveEnv:       []string{"PATH", "HOME"},
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, log)

	// Create execution function
	execFunc := execution.NewExecutionFunction()
	err = execFunc.Initialize(permissionManager, auditLogger, sudoManager, config)
	if err != nil {
		t.Fatalf("Failed to initialize execution function: %v", err)
	}

	// Test cases for end-to-end execution
	tests := []struct {
		name      string
		params    map[string]interface{}
		expectErr bool
	}{
		{
			name: "simple echo command",
			params: map[string]interface{}{
				"command":     "echo",
				"args":        []string{"Hello, World!"},
				"description": "Test echo command",
				"category":    "utility",
				"dryRun":      true,
			},
			expectErr: false,
		},
		{
			name: "list directory",
			params: map[string]interface{}{
				"command":     "ls",
				"args":        []string{tmpDir},
				"description": "List test directory",
				"category":    "utility",
				"dryRun":      true,
			},
			expectErr: false,
		},
		{
			name: "forbidden command",
			params: map[string]interface{}{
				"command":     "rm",
				"args":        []string{"-rf", "/"},
				"description": "Dangerous command",
				"category":    "utility",
			},
			expectErr: true,
		},
		{
			name: "command with environment variables",
			params: map[string]interface{}{
				"command":     "echo",
				"args":        []string{"$TEST_VAR"},
				"description": "Echo environment variable",
				"category":    "utility",
				"environment": map[string]string{"TEST_VAR": "test_value"},
				"dryRun":      true,
			},
			expectErr: false,
		},
	}

	ctx := context.Background()
	options := &functionbase.FunctionOptions{
		Timeout: time.Minute,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := execFunc.Execute(ctx, tt.params, options)

			if tt.expectErr {
				if err != nil || (result != nil && !result.Success) {
					// Expected error or failure
					return
				}
				t.Error("Expected error but got success")
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			if !result.Success {
				t.Errorf("Expected success but got failure: %s", result.Error)
			}
		})
	}
}

// TestSecurityIntegration tests security component integration
func TestSecurityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := logger.NewLogger()
	tmpDir := t.TempDir()

	// Create a more restrictive configuration for security testing
	config := &config.ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        false,
		ConfirmationRequired: false,
		MaxExecutionTime:     time.Second * 10,
		AllowedCommands:      []string{"echo", "true"},
		ForbiddenCommands:    []string{"false", "rm*", "dd"},
		SudoCommands:         []string{},
		AllowedDirectories:   []string{tmpDir},
		ForbiddenPaths:       []string{"/etc", "/boot", "/dev", "/proc", "/sys"},
		AllowedEnvironmentVariables: []string{"PATH"},
	}

	permissionManager := security.NewPermissionManager(config, log)
	
	auditLogger, err := security.NewAuditLogger(tmpDir+"/security_audit.log", true, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	sudoConfig := &security.SudoConfig{
		SessionTimeout:    time.Minute * 5,
		PasswordTimeout:   time.Minute,
		MaxAttempts:       2,
		RequirePassword:   false,
		AllowPasswordless: true,
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, log)

	safeExecutor := execTypes.NewSafeExecutor(permissionManager, auditLogger, sudoManager, config, log)

	// Test security validations
	tests := []struct {
		name      string
		request   execTypes.CommandRequest
		expectErr bool
		errorType string
	}{
		{
			name: "allowed command in allowed directory",
			request: execTypes.CommandRequest{
				Command:     "echo",
				Args:        []string{"test"},
				Description: "Safe command",
				Category:    "utility",
				WorkingDir:  tmpDir,
			},
			expectErr: false,
		},
		{
			name: "forbidden command",
			request: execTypes.CommandRequest{
				Command:     "false",
				Description: "Forbidden command",
				Category:    "utility",
			},
			expectErr: true,
			errorType: "security",
		},
		{
			name: "command in forbidden directory",
			request: execTypes.CommandRequest{
				Command:     "echo",
				Args:        []string{"test"},
				Description: "Command in forbidden directory",
				Category:    "utility",
				WorkingDir:  "/etc",
			},
			expectErr: true,
			errorType: "security",
		},
		{
			name: "forbidden environment variable",
			request: execTypes.CommandRequest{
				Command:     "echo",
				Description: "Command with forbidden env var",
				Category:    "utility",
				Environment: map[string]string{"SECRET": "value"},
			},
			expectErr: true,
			errorType: "security",
		},
		{
			name: "shell injection attempt",
			request: execTypes.CommandRequest{
				Command:     "echo",
				Args:        []string{"test; rm -rf /"},
				Description: "Shell injection attempt",
				Category:    "utility",
			},
			expectErr: true,
			errorType: "security",
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := safeExecutor.ExecuteCommand(ctx, tt.request)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				// Could check error type if needed
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestAuditingIntegration tests audit logging integration
func TestAuditingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := logger.NewLogger()
	tmpDir := t.TempDir()
	auditLogPath := tmpDir + "/audit_test.log"

	config := &config.ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        true,
		ConfirmationRequired: false,
		AllowedCommands:      []string{"echo", "true", "false"},
		AllowedDirectories:   []string{tmpDir},
		AllowedEnvironmentVariables: []string{"PATH"},
	}

	permissionManager := security.NewPermissionManager(config, log)
	
	auditLogger, err := security.NewAuditLogger(auditLogPath, true, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	sudoConfig := &security.SudoConfig{}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, log)

	safeExecutor := execTypes.NewSafeExecutor(permissionManager, auditLogger, sudoManager, config, log)

	// Execute several commands to generate audit events
	commands := []execTypes.CommandRequest{
		{
			Command:     "echo",
			Args:        []string{"audit test 1"},
			Description: "First audit test",
			Category:    "utility",
			DryRun:      true,
		},
		{
			Command:     "true",
			Description: "Success command",
			Category:    "utility",
			DryRun:      true,
		},
		{
			Command:     "echo",
			Args:        []string{"audit test 2"},
			Description: "Second audit test",
			Category:    "utility",
			DryRun:      true,
		},
	}

	ctx := context.Background()

	for i, cmd := range commands {
		t.Run(fmt.Sprintf("audit_command_%d", i), func(t *testing.T) {
			_, err := safeExecutor.ExecuteCommand(ctx, cmd)
			if err != nil {
				t.Errorf("Unexpected error in audit test: %v", err)
			}
		})
	}

	// Verify audit log file was created and has content
	if _, err := os.Stat(auditLogPath); os.IsNotExist(err) {
		t.Error("Audit log file was not created")
	}

	// Read audit log to verify events were logged
	auditData, err := os.ReadFile(auditLogPath)
	if err != nil {
		t.Errorf("Failed to read audit log: %v", err)
	}

	if len(auditData) == 0 {
		t.Error("Audit log is empty")
	}

	auditContent := string(auditData)
	if !strings.Contains(auditContent, "command_attempt") {
		t.Error("Expected command_attempt events in audit log")
	}
}

// TestPerformanceIntegration tests performance under load
func TestPerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance integration test in short mode")
	}

	log := logger.NewLogger()
	tmpDir := t.TempDir()

	config := &config.ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        true, // Use dry run for performance tests
		ConfirmationRequired: false,
		AllowedCommands:      []string{"echo"},
		AllowedDirectories:   []string{tmpDir},
		AllowedEnvironmentVariables: []string{"PATH"},
	}

	permissionManager := security.NewPermissionManager(config, log)
	
	auditLogger, err := security.NewAuditLogger(tmpDir+"/perf_audit.log", true, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()

	sudoConfig := &security.SudoConfig{}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, log)

	safeExecutor := execTypes.NewSafeExecutor(permissionManager, auditLogger, sudoManager, config, log)

	// Performance test: execute many commands concurrently
	numCommands := 100
	if testing.Short() {
		numCommands = 10
	}

	ctx := context.Background()
	start := time.Now()

	// Use a channel to collect results
	results := make(chan error, numCommands)

	for i := 0; i < numCommands; i++ {
		go func(index int) {
			request := execTypes.CommandRequest{
				Command:     "echo",
				Args:        []string{fmt.Sprintf("performance test %d", index)},
				Description: fmt.Sprintf("Performance test command %d", index),
				Category:    "utility",
				DryRun:      true,
			}

			_, err := safeExecutor.ExecuteCommand(ctx, request)
			results <- err
		}(i)
	}

	// Collect all results
	var errors []error
	for i := 0; i < numCommands; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	duration := time.Since(start)

	if len(errors) > 0 {
		t.Errorf("Got %d errors out of %d commands: %v", len(errors), numCommands, errors[0])
	}

	avgDuration := duration / time.Duration(numCommands)
	t.Logf("Performance test: %d commands in %v (avg: %v per command)", numCommands, duration, avgDuration)

	// Performance assertion: each command should complete within reasonable time
	maxAvgDuration := time.Millisecond * 100
	if avgDuration > maxAvgDuration {
		t.Errorf("Performance degraded: average duration %v exceeds maximum %v", avgDuration, maxAvgDuration)
	}
}

// TestConfigurationIntegration tests configuration loading and validation
func TestConfigurationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping configuration integration test in short mode")
	}

	log := logger.NewLogger()
	tmpDir := t.TempDir()

	// Test configuration manager
	configPath := tmpDir + "/execution_config.yaml"
	configManager := config.NewExecutionConfigManager(configPath, log)

	// Test default configuration
	defaultConfig := configManager.GetDefaultExecutionConfig()
	if err := configManager.ValidateExecutionConfig(defaultConfig); err != nil {
		t.Errorf("Default configuration validation failed: %v", err)
	}

	// Test saving and loading configuration
	if err := configManager.SaveExecutionConfig(defaultConfig); err != nil {
		t.Errorf("Failed to save configuration: %v", err)
	}

	loadedConfig, err := configManager.LoadExecutionConfig()
	if err != nil {
		t.Errorf("Failed to load configuration: %v", err)
	}

	// Verify loaded configuration matches saved configuration
	if loadedConfig.Enabled != defaultConfig.Enabled {
		t.Errorf("Configuration mismatch: Enabled %v != %v", loadedConfig.Enabled, defaultConfig.Enabled)
	}

	if len(loadedConfig.AllowedCommands) != len(defaultConfig.AllowedCommands) {
		t.Errorf("Configuration mismatch: AllowedCommands count %d != %d", 
			len(loadedConfig.AllowedCommands), len(defaultConfig.AllowedCommands))
	}

	// Test configuration summary
	summary := configManager.GetConfigSummary(loadedConfig)
	if summary["enabled"] != loadedConfig.Enabled {
		t.Error("Configuration summary mismatch")
	}

	// Test invalid configuration validation
	invalidConfig := &config.ExecutionConfig{
		Enabled:           true,
		MaxExecutionTime:  -1, // Invalid negative time
		AllowedCommands:   []string{""},   // Empty command
		ForbiddenCommands: []string{"rm"}, // Valid
	}

	if err := configManager.ValidateExecutionConfig(invalidConfig); err == nil {
		t.Error("Expected validation error for invalid configuration")
	}
}

// Helper function to simulate fmt.Sprintf since we need it
func fmt_Sprintf(format string, args ...interface{}) string {
	// Simple implementation for our test needs
	if len(args) == 0 {
		return format
	}
	
	// For our specific test case with "performance test %d"
	if format == "performance test %d" && len(args) == 1 {
		if index, ok := args[0].(int); ok {
			return "performance test " + itoa(index)
		}
	}
	
	if format == "Performance test command %d" && len(args) == 1 {
		if index, ok := args[0].(int); ok {
			return "Performance test command " + itoa(index)
		}
	}
	
	if format == "audit_command_%d" && len(args) == 1 {
		if index, ok := args[0].(int); ok {
			return "audit_command_" + itoa(index)
		}
	}
	
	return format
}

// Simple integer to string conversion
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	digits := make([]byte, 0, 10)
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	
	return string(digits)
}

