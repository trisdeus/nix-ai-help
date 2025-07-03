package execution

import (
	"context"
	"os"
	"strings"
	"testing"

	"nix-ai-help/pkg/logger"
)

func TestSystemContextDetection(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	// Test that system context is detected
	if resolver.systemContext == "" {
		t.Error("Expected system context to be detected")
	}
	
	t.Logf("Detected system context: %s", resolver.systemContext)
}

func TestInstallationOptionsGeneration(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Test with a command that should have nixpkgs mapping
	resolution, err := resolver.ResolveCommand(ctx, "htop")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if resolution.Availability == CommandNixRunnable {
		t.Logf("Testing installation options for htop")
		
		// Check that installation options are generated
		if len(resolution.InstallationOptions) == 0 {
			t.Error("Expected installation options to be generated")
		}
		
		// Check that at least one option is marked as recommended
		hasRecommended := false
		for _, option := range resolution.InstallationOptions {
			if option.Recommended {
				hasRecommended = true
				break
			}
		}
		
		if !hasRecommended {
			t.Error("Expected at least one installation option to be recommended")
		}
		
		// Log all installation options for review
		for i, option := range resolution.InstallationOptions {
			t.Logf("Option %d: %s", i+1, option.Method)
			t.Logf("  Command: %s", option.Command)
			t.Logf("  Config File: %s", option.ConfigFile)
			t.Logf("  Recommended: %v", option.Recommended)
			t.Logf("  Description: %s", option.Description)
			if option.ConfigSnippet != "" {
				t.Logf("  Config Snippet: %s", option.ConfigSnippet)
			}
			t.Logf("")
		}
	} else {
		t.Logf("htop is already available, system context: %s", resolution.SystemContext)
	}
}

func TestEnhancedExecutionSuggestion(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Test with a command that would use nix run
	resolution, err := resolver.ResolveCommand(ctx, "nonexistent-test-command-12345")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if resolution.Availability == CommandUnavailable {
		suggestion := resolver.GetExecutionSuggestion(resolution)
		t.Logf("Suggestion for unavailable command:\n%s", suggestion)
		
		if !strings.Contains(suggestion, "not available") {
			t.Error("Expected suggestion to indicate command is not available")
		}
	}
	
	// Test with firefox which should have installation options
	firefoxRes, err := resolver.ResolveCommand(ctx, "firefox")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	firefoxSuggestion := resolver.GetExecutionSuggestion(firefoxRes)
	t.Logf("Firefox suggestion:\n%s", firefoxSuggestion)
	
	// The suggestion should either show it's available or provide installation options
	if firefoxRes.Availability == CommandNixRunnable {
		if !strings.Contains(firefoxSuggestion, "Installation Options") {
			t.Error("Expected firefox suggestion to contain installation options")
		}
	}
}

func TestSystemContextSpecificOptions(t *testing.T) {
	log := logger.NewLogger()
	
	testCases := []struct {
		name    string
		context SystemContext
		pkg     string
	}{
		{"NixOS", ContextNixOS, "git"},
		{"Home Manager", ContextHomeManager, "git"},
		{"Flakes", ContextFlakes, "git"},
		{"Development", ContextDevelopment, "git"},
		{"Profile", ContextProfile, "git"},
		{"Generic", ContextGeneric, "git"},
	}
	
	for _, tc := range testCases {
		t.Run(string(tc.context), func(t *testing.T) {
			resolver := NewCommandResolver(log)
			
			// Override the system context for testing
			resolver.systemContext = tc.context
			
			// Generate installation options for the package
			options := resolver.generateInstallationOptions(tc.pkg)
			
			if len(options) == 0 {
				t.Errorf("Expected installation options for context %s", tc.context)
				return
			}
			
			t.Logf("Installation options for %s context:", tc.context)
			for i, option := range options {
				t.Logf("  %d. %s: %s", i+1, option.Method, option.Command)
				t.Logf("     File: %s", option.ConfigFile)
				t.Logf("     Recommended: %v", option.Recommended)
				
				// Validate that options make sense for the context
				switch tc.context {
				case ContextNixOS:
					if i == 0 && !strings.Contains(option.ConfigSnippet, "environment.systemPackages") {
						t.Errorf("Expected NixOS option to mention environment.systemPackages")
					}
				case ContextHomeManager:
					if i == 0 && !strings.Contains(option.ConfigSnippet, "home.packages") {
						t.Errorf("Expected Home Manager option to mention home.packages")
					}
				case ContextFlakes:
					if i == 0 && !strings.Contains(option.ConfigFile, "flake.nix") {
						t.Errorf("Expected Flakes option to reference flake.nix")
					}
				case ContextProfile:
					if i == 0 && !strings.Contains(option.Command, "nix profile install") {
						t.Errorf("Expected Profile option to use nix profile install")
					}
				}
			}
		})
	}
}

func TestCacheStatsIncludeSystemContext(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()
	
	// Resolve a command to populate cache
	_, err := resolver.ResolveCommand(ctx, "git")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Check cache stats include system context
	stats := resolver.GetCacheStats()
	
	systemContext, exists := stats["system_context"]
	if !exists {
		t.Error("Expected cache stats to include system_context")
	}
	
	if systemContext == "" {
		t.Error("Expected system_context to be non-empty")
	}
	
	t.Logf("Cache stats: %+v", stats)
}

func TestInstallationOptionsValidation(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	testPackages := []string{"git", "firefox", "htop", "curl"}
	
	for _, pkg := range testPackages {
		t.Run(pkg, func(t *testing.T) {
			options := resolver.generateInstallationOptions(pkg)
			
			if len(options) == 0 {
				t.Errorf("Expected installation options for package %s", pkg)
				return
			}
			
			for i, option := range options {
				// Validate required fields
				if option.Method == "" {
					t.Errorf("Option %d for %s: Method is empty", i, pkg)
				}
				if option.Command == "" {
					t.Errorf("Option %d for %s: Command is empty", i, pkg)
				}
				if option.ConfigFile == "" {
					t.Errorf("Option %d for %s: ConfigFile is empty", i, pkg)
				}
				if option.Description == "" {
					t.Errorf("Option %d for %s: Description is empty", i, pkg)
				}
				
				// Validate that the package name appears in command or snippet
				if !strings.Contains(option.Command, pkg) && !strings.Contains(option.ConfigSnippet, pkg) {
					t.Errorf("Option %d for %s: Package name not found in command or snippet", i, pkg)
				}
			}
		})
	}
}

func TestFileSystemDetection(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	
	// Test file existence check (should work for existing files)
	if !resolver.fileExists("/") {
		t.Error("Expected root directory to exist")
	}
	
	// Test file existence check (should fail for non-existent files)
	if resolver.fileExists("/this/path/should/not/exist/12345") {
		t.Error("Expected non-existent path to return false")
	}
	
	// Test flake.nix detection in current working directory
	hasFlake := resolver.findFlakeNix()
	t.Logf("Flake.nix detected in current directory tree: %v", hasFlake)
	
	// This should work regardless of whether we actually have a flake
	if cwd, err := os.Getwd(); err == nil {
		t.Logf("Current working directory: %s", cwd)
	}
}