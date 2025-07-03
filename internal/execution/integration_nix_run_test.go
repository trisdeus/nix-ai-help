package execution

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

func TestNixRunIntegration(t *testing.T) {
	// Create test configuration
	cfg := &config.ExecutionConfig{
		Enabled:                     true,
		DryRunDefault:              true,
		AllowedCommands:            []string{"nix", "echo", "firefox", "htop"},
		AllowedDirectories:         []string{"/tmp", "/home"},
		AllowedEnvironmentVariables: []string{"PATH", "HOME"},
		MaxExecutionTime:           0, // No timeout for tests
	}
	
	log := logger.NewLogger()
	
	// Create security components
	permissionManager := security.NewPermissionManager(cfg, log)
	auditLogger, err := security.NewAuditLogger("/tmp/test_audit.log", false, log)
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	defer auditLogger.Close()
	
	sudoConfig := &security.SudoConfig{
		SessionTimeout:  time.Hour,
		PasswordTimeout: time.Minute * 15,
		MaxAttempts:     3,
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, log)
	
	// Create SafeExecutor with command resolution
	executor := NewSafeExecutor(permissionManager, auditLogger, sudoManager, cfg, log)
	
	ctx := context.Background()
	
	t.Run("Resolve available command", func(t *testing.T) {
		resolution, err := executor.ResolveCommand(ctx, "echo")
		if err != nil {
			t.Fatalf("Expected no error resolving echo, got: %v", err)
		}
		
		if resolution.Availability != CommandAvailable {
			t.Errorf("Expected echo to be available, got: %s", resolution.Availability)
		}
	})
	
	t.Run("Resolve nix runnable command", func(t *testing.T) {
		resolution, err := executor.ResolveCommand(ctx, "firefox")
		if err != nil {
			t.Fatalf("Expected no error resolving firefox, got: %v", err)
		}
		
		// firefox should either be available or nix runnable
		if resolution.Availability != CommandAvailable && resolution.Availability != CommandNixRunnable {
			t.Errorf("Expected firefox to be available or nix runnable, got: %s", resolution.Availability)
		}
		
		if resolution.Availability == CommandNixRunnable {
			if resolution.NixRunCommand == "" {
				t.Error("Expected nix run command to be populated")
			}
			
			if resolution.NixPackage != "firefox" {
				t.Errorf("Expected nix package to be 'firefox', got: %s", resolution.NixPackage)
			}
		}
	})
	
	t.Run("Get execution suggestion", func(t *testing.T) {
		suggestion, err := executor.GetCommandSuggestion(ctx, "firefox")
		if err != nil {
			t.Fatalf("Expected no error getting suggestion, got: %v", err)
		}
		
		if suggestion == "" {
			t.Error("Expected non-empty suggestion")
		}
		
		t.Logf("Firefox suggestion: %s", suggestion)
	})
	
	t.Run("Execute with nix run transformation", func(t *testing.T) {
		// Create a command request for a package that should use nix run
		req := CommandRequest{
			Command:     "htop",
			Args:        []string{},
			Description: "System monitor",
			Category:    "utility",
			DryRun:      true, // Always dry run in tests
		}
		
		result, err := executor.ExecuteCommand(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error executing htop, got: %v", err)
		}
		
		if !result.Success {
			t.Error("Expected execution to succeed")
		}
		
		if !result.DryRun {
			t.Error("Expected dry run execution")
		}
		
		// The command might have been transformed to use nix run
		t.Logf("Final command: %s", result.Command)
		t.Logf("Output: %s", result.Output)
	})
}

func TestCommandResolutionCaching(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// First resolution should populate cache
	resolution1, err := resolver.ResolveCommand(ctx, "git")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Second resolution should use cache
	resolution2, err := resolver.ResolveCommand(ctx, "git")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Should be identical (from cache)
	if resolution1.LastChecked != resolution2.LastChecked {
		t.Error("Expected cache to be used (same LastChecked time)")
	}
	
	// Check cache statistics
	stats := resolver.GetCacheStats()
	totalCached := stats["total_cached"].(int)
	if totalCached < 1 {
		t.Error("Expected at least one cached entry")
	}
	
	available := stats["available"].(int)
	nixRunnable := stats["nix_runnable"].(int)
	unavailable := stats["unavailable"].(int)
	
	t.Logf("Cache stats - Total: %d, Available: %d, Nix Runnable: %d, Unavailable: %d", 
		totalCached, available, nixRunnable, unavailable)
}

func TestMultipleCommandResolution(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	testCommands := []string{
		"ls",       // Should be available
		"git",      // Should be available or nix runnable
		"firefox",  // Should be nix runnable (unless installed)
		"htop",     // Should be nix runnable (unless installed)
		"nonexistent-command-12345", // Should be unavailable
	}
	
	results := make(map[string]*CommandResolution)
	
	for _, cmd := range testCommands {
		resolution, err := resolver.ResolveCommand(ctx, cmd)
		if err != nil {
			t.Logf("Warning: Error resolving %s: %v", cmd, err)
			continue
		}
		
		results[cmd] = resolution
		
		if resolution != nil {
			suggestion := resolver.GetExecutionSuggestion(resolution)
			t.Logf("Command: %s, Availability: %s, Suggestion: %s", 
				cmd, resolution.Availability, suggestion)
		}
	}
	
	// Verify we got reasonable results
	if len(results) < 3 {
		t.Errorf("Expected to resolve at least 3 commands, got %d", len(results))
	}
	
	// ls should definitely be available
	if lsRes, exists := results["ls"]; exists && lsRes.Availability != CommandAvailable {
		t.Errorf("Expected 'ls' to be available, got: %s", lsRes.Availability)
	}
	
	// The non-existent command should be unavailable
	if nonExistentRes, exists := results["nonexistent-command-12345"]; exists {
		if nonExistentRes.Availability != CommandUnavailable && nonExistentRes.Availability != CommandUnknown {
			t.Errorf("Expected non-existent command to be unavailable or unknown, got: %s", 
				nonExistentRes.Availability)
		}
	}
}

func TestNixRunCommandGeneration(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	// Test nix run command generation for various scenarios
	testCases := []struct {
		name     string
		resolution *CommandResolution
		args     []string
		expected string
	}{
		{
			name: "Simple nix run",
			resolution: &CommandResolution{
				Command:        "firefox",
				Availability:   CommandNixRunnable,
				NixRunCommand: "nix run nixpkgs#firefox",
			},
			args:     []string{},
			expected: "nix run nixpkgs#firefox",
		},
		{
			name: "Nix run with arguments",
			resolution: &CommandResolution{
				Command:        "firefox",
				Availability:   CommandNixRunnable,
				NixRunCommand: "nix run nixpkgs#firefox",
			},
			args:     []string{"--help"},
			expected: "nix run nixpkgs#firefox -- --help",
		},
		{
			name: "Available command",
			resolution: &CommandResolution{
				Command:      "echo",
				Availability: CommandAvailable,
			},
			args:     []string{"hello", "world"},
			expected: "echo hello world",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := resolver.ResolveCommandString(tc.resolution, tc.args)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}