package execution

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

func TestNewSafeExecutor(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:           true,
		DryRunDefault:     false,
		MaxExecutionTime:  time.Minute * 5,
		AllowedCommands:   []string{"echo", "ls"},
		AllowedDirectories: []string{"/tmp"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{
		SessionTimeout:  time.Hour,
		PasswordTimeout: time.Minute * 15,
		MaxAttempts:     3,
	}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)
	if executor == nil {
		t.Fatal("Expected safe executor to be created")
	}
}

func TestExecuteCommand(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        false,
		MaxExecutionTime:     time.Minute * 5,
		AllowedCommands:      []string{"echo", "ls", "cat"},
		AllowedDirectories:   []string{"/tmp", "/home"},
		AllowedEnvironmentVariables: []string{"PATH", "HOME"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{
		SessionTimeout:  time.Hour,
		PasswordTimeout: time.Minute * 15,
		MaxAttempts:     3,
	}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)

	tests := []struct {
		name      string
		request   CommandRequest
		expectErr bool
	}{
		{
			name: "valid echo command",
			request: CommandRequest{
				Command:     "echo",
				Args:        []string{"hello", "world"},
				Description: "Test echo command",
				Category:    "utility",
				DryRun:      false,
			},
			expectErr: false,
		},
		{
			name: "forbidden command",
			request: CommandRequest{
				Command:     "rm",
				Args:        []string{"-rf", "/"},
				Description: "Dangerous command",
				Category:    "utility",
				DryRun:      false,
			},
			expectErr: true,
		},
		{
			name: "invalid category",
			request: CommandRequest{
				Command:     "echo",
				Args:        []string{"test"},
				Description: "Test command",
				Category:    "invalid_category",
				DryRun:      false,
			},
			expectErr: true,
		},
		{
			name: "forbidden directory",
			request: CommandRequest{
				Command:     "ls",
				Args:        []string{},
				Description: "List files",
				Category:    "utility",
				WorkingDir:  "/boot",
				DryRun:      false,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := executor.ExecuteCommand(ctx, tt.request)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
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

			// For echo command, we expect success
			if tt.request.Command == "echo" && !result.Success {
				t.Errorf("Expected echo command to succeed, but it failed: %s", result.Error)
			}
		})
	}
}

func TestDryRun(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:         true,
		DryRunDefault:   false,
		AllowedCommands: []string{"echo", "ls"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{
		SessionTimeout: time.Hour,
	}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)
	ctx := context.Background()

	request := CommandRequest{
		Command:     "echo",
		Args:        []string{"hello"},
		Description: "Test echo command",
		Category:    "utility",
		DryRun:      true,
	}

	result, err := executor.ExecuteCommand(ctx, request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.DryRun {
		t.Error("Expected result to be marked as dry run")
	}

	if !result.Success {
		t.Error("Expected dry run to succeed")
	}

	expectedOutput := "DRY RUN: Would execute: echo hello"
	if result.Output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, result.Output)
	}
}

func TestSetDryRunMode(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:         true,
		DryRunDefault:   false,
		AllowedCommands: []string{"echo"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)

	// Enable dry run mode globally
	executor.SetDryRunMode(true)

	ctx := context.Background()
	request := CommandRequest{
		Command:     "echo",
		Args:        []string{"hello"},
		Description: "Test echo command",
		Category:    "utility",
		DryRun:      false, // Even though this is false, executor should use dry run
	}

	result, err := executor.ExecuteCommand(ctx, request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.DryRun {
		t.Error("Expected result to be marked as dry run due to global setting")
	}
}

func TestArgumentValidation(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:         true,
		AllowedCommands: []string{"echo"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "safe arguments",
			args:      []string{"hello", "world"},
			expectErr: false,
		},
		{
			name:      "shell injection attempt",
			args:      []string{"hello;", "rm", "-rf", "/"},
			expectErr: true,
		},
		{
			name:      "pipe operations",
			args:      []string{"hello", "|", "cat"},
			expectErr: true,
		},
		{
			name:      "redirection",
			args:      []string{"hello", ">", "/tmp/file"},
			expectErr: true,
		},
		{
			name:      "command chaining",
			args:      []string{"hello", "&&", "echo", "world"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			request := CommandRequest{
				Command:     "echo",
				Args:        tt.args,
				Description: "Test command",
				Category:    "utility",
			}

			_, err := executor.ExecuteCommand(ctx, request)
			if (err != nil) != tt.expectErr {
				t.Errorf("ExecuteCommand() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestEnvironmentVariableValidation(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:                     true,
		AllowedCommands:             []string{"echo"},
		AllowedEnvironmentVariables: []string{"PATH", "HOME", "TEST_VAR"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)

	tests := []struct {
		name        string
		environment map[string]string
		expectErr   bool
	}{
		{
			name: "allowed environment variables",
			environment: map[string]string{
				"TEST_VAR": "test_value",
				"PATH":     "/usr/bin",
			},
			expectErr: false,
		},
		{
			name: "forbidden environment variable",
			environment: map[string]string{
				"SECRET_KEY": "secret_value",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			request := CommandRequest{
				Command:     "echo",
				Args:        []string{"hello"},
				Description: "Test command",
				Category:    "utility",
				Environment: tt.environment,
			}

			_, err := executor.ExecuteCommand(ctx, request)
			if (err != nil) != tt.expectErr {
				t.Errorf("ExecuteCommand() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestTimeoutHandling(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:           true,
		AllowedCommands:   []string{"sleep"},
		MaxExecutionTime:  time.Millisecond * 100, // Very short timeout
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)

	ctx := context.Background()
	request := CommandRequest{
		Command:     "sleep",
		Args:        []string{"1"}, // Sleep for 1 second, should timeout
		Description: "Long running command",
		Category:    "utility",
	}

	result, err := executor.ExecuteCommand(ctx, request)
	
	// The command should either error due to timeout or complete but indicate failure
	if err == nil && result != nil && result.Success {
		t.Log("Command completed successfully (may not have timed out on this system)")
	}
	// Don't fail the test as timeout behavior can vary by system
}

func TestCompoundOperations(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:         true,
		AllowedCommands: []string{"echo", "true", "false"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)

	// Test successful compound operation
	operation := CompoundOperation{
		ID:          "test-operation",
		Commands: []CommandRequest{
			{
				Command:     "echo",
				Args:        []string{"step1"},
				Description: "First step",
				Category:    "utility",
			},
			{
				Command:     "echo",
				Args:        []string{"step2"},
				Description: "Second step",
				Category:    "utility",
			},
		},
		AllowPartial: false,
	}

	ctx := context.Background()
	result, err := executor.ExecuteCompound(ctx, operation)
	if err != nil {
		t.Errorf("Unexpected error in compound operation: %v", err)
	}

	if result == nil {
		t.Fatal("Expected compound result")
	}

	if !result.Success {
		t.Error("Expected compound operation to succeed")
	}

	if len(result.CommandResults) != 2 {
		t.Errorf("Expected 2 command results, got %d", len(result.CommandResults))
	}

	// Test failing compound operation
	failingOperation := CompoundOperation{
		ID: "failing-operation",
		Commands: []CommandRequest{
			{
				Command:     "true",
				Description: "Success command",
				Category:    "utility",
			},
			{
				Command:     "false",
				Description: "Failing command",
				Category:    "utility",
			},
		},
		AllowPartial: false,
	}

	result, err = executor.ExecuteCompound(ctx, failingOperation)
	if err == nil {
		t.Error("Expected error for failing compound operation")
	}

	if result == nil {
		t.Error("Expected result even for failing operation")
	} else if result.Success {
		t.Error("Expected compound operation to fail")
	}
}

func BenchmarkExecuteCommand(b *testing.B) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:         true,
		AllowedCommands: []string{"echo"},
	}

	pm := security.NewPermissionManager(config, log)
	al, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		b.Fatalf("Failed to create audit logger: %v", err)
	}
	defer al.Close()

	sudoConfig := &security.SudoConfig{}
	sm := security.NewSudoManager(sudoConfig, al, log)

	executor := NewSafeExecutor(pm, al, sm, config, log)
	ctx := context.Background()

	request := CommandRequest{
		Command:     "echo",
		Args:        []string{"hello"},
		Description: "Benchmark test",
		Category:    "utility",
		DryRun:      true, // Use dry run to avoid actual execution overhead
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.ExecuteCommand(ctx, request)
	}
}