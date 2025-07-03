package execution

import (
	"context"
	"strings"
	"testing"

	"nix-ai-help/pkg/logger"
)

// TestRealisticScenarioDemo demonstrates the enhanced installation suggestions
// in a way that shows the user-facing improvements
func TestRealisticScenarioDemo(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()

	// Simulate a user trying to install a development tool
	t.Log("=== Scenario: User wants to install 'htop' for system monitoring ===")
	
	resolution, err := resolver.ResolveCommand(ctx, "htop")
	if err != nil {
		t.Fatalf("Error resolving htop: %v", err)
	}

	t.Logf("System Context Detected: %s", resolver.systemContext)
	t.Logf("Command Availability: %s", resolution.Availability)

	// Show the user-friendly suggestion
	suggestion := resolver.GetExecutionSuggestion(resolution)
	t.Log("\n=== User-Facing Suggestion ===")
	t.Log(suggestion)

	// Also demonstrate a package that definitely requires nix run
	t.Log("\n=== Scenario: User wants a less common package (example: cowsay) ===")
	
	// Override the resolver to simulate cowsay being unavailable locally
	testResolver := NewCommandResolver(log)
	// Manually create a scenario where cowsay would need nix run
	cowsayRes := &CommandResolution{
		Command:       "cowsay",
		Availability:  CommandNixRunnable,
		NixPackage:    "cowsay",
		NixRunCommand: "nix run nixpkgs#cowsay",
		Description:   "A configurable talking cow",
		EstimatedSize: "850KB",
		SystemContext: testResolver.systemContext,
	}
	
	// Generate installation options for the detected context
	cowsayRes.InstallationOptions = testResolver.generateInstallationOptions("cowsay")
	
	cowsaySuggestion := testResolver.GetExecutionSuggestion(cowsayRes)
	t.Log("\n=== Cowsay Installation Suggestion ===")
	t.Log(cowsaySuggestion)
}

// TestInstallationChoicesByContext shows different suggestions for different system types
func TestInstallationChoicesByContext(t *testing.T) {
	log := logger.NewLogger()
	packageName := "neofetch"

	contexts := []struct {
		name    string
		context SystemContext
		desc    string
	}{
		{"Personal NixOS Desktop", ContextNixOS, "System-wide installation recommended"},
		{"Development Laptop", ContextHomeManager, "User-specific installation via Home Manager"},
		{"Flake-based Project", ContextFlakes, "Declarative flake configuration"},
		{"Temporary Dev Environment", ContextDevelopment, "Project-specific shell.nix"},
		{"Modern Nix User", ContextProfile, "User profile with nix profile commands"},
		{"Legacy Nix Setup", ContextGeneric, "Traditional nix-env installation"},
	}

	for _, scenario := range contexts {
		t.Run(scenario.name, func(t *testing.T) {
			resolver := NewCommandResolver(log)
			// Override context for testing
			resolver.systemContext = scenario.context
			
			options := resolver.generateInstallationOptions(packageName)
			
			t.Logf("=== %s ===", scenario.name)
			t.Logf("Context: %s (%s)", scenario.context, scenario.desc)
			t.Logf("Package: %s", packageName)
			t.Log("")
			
			for i, option := range options {
				indicator := "   "
				if option.Recommended {
					indicator = "⭐ "
				}
				
				t.Logf("%s%d. %s", indicator, i+1, option.Method)
				t.Logf("      Command: %s", option.Command)
				t.Logf("      File: %s", option.ConfigFile)
				t.Logf("      Config: %s", option.ConfigSnippet)
				t.Logf("      Info: %s", option.Description)
				
				if i < len(options)-1 {
					t.Log("")
				}
			}
			
			t.Log("\n" + strings.Repeat("-", 60))
		})
	}
}

// TestUserWorkflowSimulation simulates a complete user workflow
func TestUserWorkflowSimulation(t *testing.T) {
	log := logger.NewLogger()
	resolver := NewCommandResolver(log)
	ctx := context.Background()

	t.Log("=== Complete User Workflow Simulation ===")
	t.Log("User scenario: Developer setting up a new Rust project")
	t.Log("")

	devTools := []string{"rustc", "cargo", "git", "tree"}

	for i, tool := range devTools {
		t.Logf("Step %d: Checking if '%s' is available...", i+1, tool)
		
		resolution, err := resolver.ResolveCommand(ctx, tool)
		if err != nil {
			t.Logf("  Error checking %s: %v", tool, err)
			continue
		}

		switch resolution.Availability {
		case CommandAvailable:
			t.Logf("  ✅ %s is already installed and ready to use", tool)
			
		case CommandNixRunnable:
			t.Logf("  🚀 %s can be installed via nix. Options:", tool)
			
			if len(resolution.InstallationOptions) > 0 {
				for j, option := range resolution.InstallationOptions {
					prefix := "     "
					if option.Recommended {
						prefix = "  ⭐ "
					}
					t.Logf("%s%d. %s: %s", prefix, j+1, option.Method, option.Command)
				}
			}
			
		case CommandUnavailable:
			t.Logf("  ❌ %s is not available", tool)
			if len(resolution.Suggestions) > 0 {
				t.Logf("     Alternatives: %v", resolution.Suggestions)
			}
		}
		
		t.Log("")
	}

	// Show system context summary
	t.Logf("System Context: %s", resolver.systemContext)
	
	// Show cache statistics
	stats := resolver.GetCacheStats()
	t.Logf("Cache: %d total, %d available, %d via nix run", 
		stats["total_cached"], stats["available"], stats["nix_runnable"])
}

// TestContextualRecommendations shows how recommendations change based on context
func TestContextualRecommendations(t *testing.T) {
	t.Log("=== Contextual Recommendations Demo ===")
	t.Log("How installation suggestions adapt to different environments")
	t.Log("")

	scenarios := map[string]struct {
		context SystemContext
		useCase string
	}{
		"Personal Computer": {ContextNixOS, "Installing system-wide tools like htop, git"},
		"Work Laptop": {ContextHomeManager, "User-specific tools that don't need admin access"},
		"CI/CD Pipeline": {ContextFlakes, "Reproducible build environments"},
		"Temporary Project": {ContextDevelopment, "Project-specific dependencies"},
		"Container/Cloud": {ContextProfile, "Lightweight, user-profile installations"},
	}

	packageToTest := "jq" // JSON processor - common utility

	for scenario, config := range scenarios {
		t.Logf("--- %s ---", scenario)
		t.Logf("Use case: %s", config.useCase)
		
		log := logger.NewLogger()
		resolver := NewCommandResolver(log)
		resolver.systemContext = config.context
		
		options := resolver.generateInstallationOptions(packageToTest)
		
		if len(options) > 0 && options[0].Recommended {
			t.Logf("Recommended: %s", options[0].Method)
			t.Logf("Command: %s", options[0].Command)
			t.Logf("Rationale: %s", options[0].Description)
		}
		
		t.Log("")
	}
}