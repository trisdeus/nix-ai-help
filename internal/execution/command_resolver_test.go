package execution

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/pkg/logger"
)

func TestNewCommandResolver(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	if resolver == nil {
		t.Fatal("Expected command resolver instance, got nil")
	}
	
	if resolver.logger == nil {
		t.Error("Expected logger to be initialized")
	}
	
	if resolver.cache == nil {
		t.Error("Expected cache to be initialized")
	}
	
	if resolver.cacheTimeout == 0 {
		t.Error("Expected cache timeout to be set")
	}
}

func TestResolveCommandAvailable(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Test with a command that should be available (ls exists on most systems)
	resolution, err := resolver.ResolveCommand(ctx, "ls")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if resolution == nil {
		t.Fatal("Expected resolution, got nil")
	}
	
	if resolution.Command != "ls" {
		t.Errorf("Expected command 'ls', got '%s'", resolution.Command)
	}
	
	if resolution.Availability != CommandAvailable {
		t.Errorf("Expected availability %s, got %s", CommandAvailable, resolution.Availability)
	}
}

func TestResolveCommandDirectMapping(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Test with a command that has a known nixpkgs mapping
	resolution, err := resolver.ResolveCommand(ctx, "firefox")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if resolution == nil {
		t.Fatal("Expected resolution, got nil")
	}
	
	// firefox might be available or available via nix run
	if resolution.Availability != CommandAvailable && resolution.Availability != CommandNixRunnable {
		t.Errorf("Expected availability %s or %s, got %s", 
			CommandAvailable, CommandNixRunnable, resolution.Availability)
	}
	
	if resolution.Availability == CommandNixRunnable {
		if resolution.NixRunCommand == "" {
			t.Error("Expected nix run command to be set")
		}
		
		if resolution.NixPackage == "" {
			t.Error("Expected nix package to be set")
		}
	}
}

func TestResolveCommandString(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	// Test available command
	availableResolution := &CommandResolution{
		Command:      "ls",
		Availability: CommandAvailable,
	}
	
	result := resolver.ResolveCommandString(availableResolution, []string{"-la"})
	expected := "ls -la"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test nix runnable command
	nixResolution := &CommandResolution{
		Command:        "firefox",
		Availability:   CommandNixRunnable,
		NixRunCommand: "nix run nixpkgs#firefox",
	}
	
	result = resolver.ResolveCommandString(nixResolution, []string{"--help"})
	expected = "nix run nixpkgs#firefox -- --help"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestGetExecutionSuggestion(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	tests := []struct {
		name       string
		resolution *CommandResolution
		shouldContain string
	}{
		{
			name: "Available command",
			resolution: &CommandResolution{
				Command:      "ls",
				Availability: CommandAvailable,
			},
			shouldContain: "available and ready",
		},
		{
			name: "Nix runnable command",
			resolution: &CommandResolution{
				Command:        "firefox",
				Availability:   CommandNixRunnable,
				NixRunCommand: "nix run nixpkgs#firefox",
				NixPackage:    "firefox",
				Description:   "Mozilla Firefox web browser",
			},
			shouldContain: "temporarily using",
		},
		{
			name: "Unavailable command",
			resolution: &CommandResolution{
				Command:      "nonexistent",
				Availability: CommandUnavailable,
			},
			shouldContain: "not available",
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			suggestion := resolver.GetExecutionSuggestion(test.resolution)
			if suggestion == "" {
				t.Error("Expected non-empty suggestion")
			}
			
			if test.shouldContain != "" && !contains(suggestion, test.shouldContain) {
				t.Errorf("Expected suggestion to contain '%s', got: %s", test.shouldContain, suggestion)
			}
		})
	}
}

func TestCacheManagement(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Resolve a command to populate cache
	_, err := resolver.ResolveCommand(ctx, "ls")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Check cache stats
	stats := resolver.GetCacheStats()
	if stats == nil {
		t.Fatal("Expected cache stats, got nil")
	}
	
	totalCached, ok := stats["total_cached"].(int)
	if !ok || totalCached < 1 {
		t.Error("Expected at least one cached entry")
	}
	
	// Clear cache
	resolver.ClearCache()
	
	// Check cache is cleared
	stats = resolver.GetCacheStats()
	totalCached, ok = stats["total_cached"].(int)
	if !ok || totalCached != 0 {
		t.Error("Expected cache to be cleared")
	}
}

func TestDirectPackageMappings(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Test some known direct mappings
	testMappings := map[string]string{
		"git":     "git",
		"curl":    "curl",
		"firefox": "firefox",
		"docker":  "docker",
		"rg":      "ripgrep",
	}
	
	for command, expectedPackage := range testMappings {
		t.Run(command, func(t *testing.T) {
			result := resolver.searchNixpkgsDirectly(ctx, command)
			if result == nil {
				t.Errorf("Expected mapping for command %s", command)
				return
			}
			
			if result.Package != expectedPackage {
				t.Errorf("Expected package %s for command %s, got %s", 
					expectedPackage, command, result.Package)
			}
		})
	}
}

func TestFindSimilarCommands(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	suggestions := resolver.findSimilarCommands("ls")
	if len(suggestions) == 0 {
		t.Error("Expected suggestions for 'ls' command")
	}
	
	// Should suggest alternatives like exa, tree
	hasSuggestion := false
	for _, suggestion := range suggestions {
		if suggestion == "exa" || suggestion == "tree" {
			hasSuggestion = true
			break
		}
	}
	
	if !hasSuggestion {
		t.Error("Expected suggestions to include alternatives like 'exa' or 'tree'")
	}
}

func TestCommandAvailabilityDetection(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	// Test with a command that should definitely exist
	if !resolver.isCommandAvailable("ls") {
		t.Error("Expected 'ls' command to be available")
	}
	
	// Test with a command that should not exist
	if resolver.isCommandAvailable("this-command-definitely-does-not-exist-12345") {
		t.Error("Expected non-existent command to be unavailable")
	}
}

func TestCacheTimeout(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	// Set a very short cache timeout for testing
	resolver.cacheTimeout = time.Millisecond * 10
	
	ctx := context.Background()
	
	// Resolve a command
	resolution1, err := resolver.ResolveCommand(ctx, "ls")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Wait for cache to expire
	time.Sleep(time.Millisecond * 20)
	
	// Resolve again - should not use cache
	resolution2, err := resolver.ResolveCommand(ctx, "ls")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Times should be different (new resolution)
	if resolution1.LastChecked.Equal(resolution2.LastChecked) {
		t.Error("Expected different resolution times after cache expiry")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || (len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   func() bool {
			   for i := 1; i <= len(s)-len(substr); i++ {
				   if s[i:i+len(substr)] == substr {
					   return true
				   }
			   }
			   return false
		   }())))
}