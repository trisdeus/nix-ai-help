package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/plugins"
)

// SimpleCustomAgent demonstrates how to create a basic custom agent as a plugin
type SimpleCustomAgent struct {
	name        string
	version     string
	description string
	started     bool
}

// Plugin metadata methods
func (p *SimpleCustomAgent) Name() string        { return p.name }
func (p *SimpleCustomAgent) Version() string     { return p.version }
func (p *SimpleCustomAgent) Description() string { return p.description }
func (p *SimpleCustomAgent) Author() string      { return "Example User" }
func (p *SimpleCustomAgent) Repository() string  { return "https://github.com/example/simple-agent" }
func (p *SimpleCustomAgent) License() string     { return "MIT" }
func (p *SimpleCustomAgent) Dependencies() []string { return []string{} }
func (p *SimpleCustomAgent) Capabilities() []string {
	return []string{"analysis", "advice", "explanation"}
}

// Lifecycle methods
func (p *SimpleCustomAgent) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	p.name = "simple-custom-agent"
	p.version = "1.0.0"
	p.description = "A simple example of a custom NixOS assistant agent"
	return nil
}

func (p *SimpleCustomAgent) Start(ctx context.Context) error {
	p.started = true
	return nil
}

func (p *SimpleCustomAgent) Stop(ctx context.Context) error {
	p.started = false
	return nil
}

func (p *SimpleCustomAgent) Cleanup(ctx context.Context) error {
	return nil
}

// Operation definitions
func (p *SimpleCustomAgent) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "explain-concept",
			Description: "Explain NixOS concepts in simple terms",
			Parameters: map[string]plugins.PluginParameter{
				"concept": {
					Type:        "string",
					Description: "The NixOS concept to explain",
					Required:    true,
				},
				"level": {
					Type:        "string",
					Description: "Explanation level (beginner/intermediate/advanced)",
					Required:    false,
					Default:     "beginner",
				},
			},
			ReturnType: "string",
			Examples: []plugins.PluginExample{
				{
					Description: "Explain flakes to a beginner",
					Parameters: map[string]interface{}{
						"concept": "flakes",
						"level":   "beginner",
					},
					Expected: "Simple explanation of NixOS flakes",
				},
			},
		},
		{
			Name:        "analyze-config",
			Description: "Analyze a NixOS configuration snippet",
			Parameters: map[string]plugins.PluginParameter{
				"config": {
					Type:        "string",
					Description: "NixOS configuration to analyze",
					Required:    true,
				},
				"focus": {
					Type:        "string",
					Description: "What to focus on (security/performance/best-practices)",
					Required:    false,
					Default:     "best-practices",
				},
			},
			ReturnType: "object",
		},
		{
			Name:        "suggest-improvement",
			Description: "Suggest improvements for a given scenario",
			Parameters: map[string]plugins.PluginParameter{
				"scenario": {
					Type:        "string",
					Description: "Description of the current setup or problem",
					Required:    true,
				},
				"goal": {
					Type:        "string",
					Description: "What you want to achieve",
					Required:    false,
					Default:     "optimization",
				},
			},
			ReturnType: "object",
		},
	}
}

// Schema method
func (p *SimpleCustomAgent) GetSchema(operation string) (*plugins.PluginSchema, error) {
	operations := p.GetOperations()
	for _, op := range operations {
		if op.Name == operation {
			return &plugins.PluginSchema{
				Name:        op.Name,
				Description: op.Description,
				Parameters:  op.Parameters,
				ReturnType:  op.ReturnType,
			}, nil
		}
	}
	return nil, fmt.Errorf("operation %s not found", operation)
}

// Main execution method
func (p *SimpleCustomAgent) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	if !p.started {
		return nil, fmt.Errorf("plugin not started")
	}

	switch operation {
	case "explain-concept":
		return p.explainConcept(params)
	case "analyze-config":
		return p.analyzeConfig(params)
	case "suggest-improvement":
		return p.suggestImprovement(params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// Operation implementations
func (p *SimpleCustomAgent) explainConcept(params map[string]interface{}) (interface{}, error) {
	concept := params["concept"].(string)
	level := "beginner"
	if l, ok := params["level"].(string); ok {
		level = l
	}

	// Simple concept explanations (in a real implementation, this would use AI)
	explanations := map[string]map[string]string{
		"flakes": {
			"beginner": `NixOS Flakes are like "recipes" for your system that include all the ingredients (dependencies) needed. 

Think of it this way:
• Traditional Nix: "Use whatever versions are available"
• Flakes: "Use exactly these specific versions"

Benefits:
• Reproducible builds across different machines
• Easy sharing of configurations
• Better dependency management
• Input locking for consistency

Example flake.nix:
{
  description = "My NixOS config";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [ ./configuration.nix ];
    };
  };
}`,
			"intermediate": `Flakes provide hermetic, reproducible Nix expressions with explicit input dependencies.

Key concepts:
• flake.nix: The entry point defining inputs/outputs
• flake.lock: Locks input revisions for reproducibility  
• Inputs: External dependencies (nixpkgs, other flakes)
• Outputs: What your flake produces (packages, systems, etc.)

Advantages over channels:
• Explicit dependency declaration
• Reproducible evaluation across machines
• Composable and shareable
• Built-in support for Git repositories

Commands:
• nix flake init: Create new flake
• nix flake update: Update all inputs
• nix flake lock: Generate lock file
• nixos-rebuild switch --flake .#hostname`,
			"advanced": `Flakes implement a pure evaluation model with cryptographic input pinning.

Architecture:
• Pure evaluation: No access to global state or channels
• Content-addressed storage: Inputs identified by content hash
• Lazy evaluation: Only evaluate required outputs
• Schema validation: Strict input/output schema enforcement

Implementation details:
• Uses libgit2 for Git operations
• Implements tree hashing for reproducibility  
• Supports multiple input types (github, gitlab, path, etc.)
• Output schema supports packages, apps, devShells, etc.

Advanced patterns:
• Flake modules for reusable configurations
• Input follows for sharing dependencies
• Output specialization for multi-system support
• Custom schemas for domain-specific outputs`,
		},
		"nixpkgs": {
			"beginner": `Nixpkgs is the massive collection of software packages available for NixOS.

Think of it as:
• App Store for NixOS with 80,000+ packages
• All packages are built from source
• Multiple versions can coexist
• Packages are isolated from each other

How it works:
• Each package has a "derivation" (build recipe)
• Built packages go into /nix/store/
• Symlinks create the appearance of traditional paths
• Rolling updates keep packages current

Search packages: https://search.nixos.org/packages`,
			"intermediate": `Nixpkgs is a collection of package derivations organized as a Git repository.

Structure:
• pkgs/: Package definitions organized by category
• lib/: Utility functions for Nix expressions
• nixos/: NixOS modules and system configurations
• doc/: Documentation and contribution guidelines

Key concepts:
• Derivations: Build instructions for packages
• stdenv: Standard build environment
• Overlays: Modify or extend package sets
• Cross-compilation: Build for different architectures
• Hydra: Continuous integration system`,
			"advanced": `Nixpkgs implements a functional package management system with advanced features.

Advanced features:
• Multiple output splitting (dev, doc, out)
• Cross-compilation infrastructure
• Bootstrap process for self-hosting
• Package staging (development workflow)
• Automated vulnerability scanning

Internals:
• Fixed-output derivations for source fetching
• Build-time vs. runtime dependencies
• Dependency graph optimization
• Binary cache distribution via NAR files
• Content-addressed derivations (experimental)`,
		},
	}

	if conceptExpl, exists := explanations[strings.ToLower(concept)]; exists {
		if explanation, levelExists := conceptExpl[level]; levelExists {
			return explanation, nil
		}
	}

	// Fallback for unknown concepts
	return fmt.Sprintf(`I don't have a specific explanation for "%s" yet, but here's some general guidance:

For NixOS concepts, I recommend:
1. Check the NixOS manual: https://nixos.org/manual/nixos/stable/
2. Visit the Nix pills tutorial: https://nixos.org/guides/nix-pills/
3. Search the community wiki: https://nixos.wiki/
4. Ask on the NixOS Discourse: https://discourse.nixos.org/

If this is a specific package or option, try:
• nixai search %s
• nixai explain-option %s
• man configuration.nix`, concept, concept), nil
}

func (p *SimpleCustomAgent) analyzeConfig(params map[string]interface{}) (interface{}, error) {
	config := params["config"].(string)
	focus := "best-practices"
	if f, ok := params["focus"].(string); ok {
		focus = f
	}

	analysis := map[string]interface{}{
		"config":    config,
		"focus":     focus,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Simple analysis logic
	issues := []string{}
	suggestions := []string{}

	// Basic checks
	if strings.Contains(config, "users.users.root") {
		issues = append(issues, "Direct root user configuration detected")
		suggestions = append(suggestions, "Consider using users.users.<name> with wheel group instead")
	}

	if !strings.Contains(config, "networking.firewall") {
		issues = append(issues, "No firewall configuration found")
		suggestions = append(suggestions, "Add networking.firewall.enable = true; for security")
	}

	if strings.Contains(config, "services.openssh.enable = true") && !strings.Contains(config, "passwordAuthentication = false") {
		issues = append(issues, "SSH password authentication may be enabled")
		suggestions = append(suggestions, "Consider services.openssh.settings.PasswordAuthentication = false;")
	}

	if strings.Count(config, "\n") > 100 {
		suggestions = append(suggestions, "Large configuration file - consider splitting into modules")
	}

	analysis["issues"] = issues
	analysis["suggestions"] = suggestions
	analysis["severity"] = p.calculateSeverity(issues)

	return analysis, nil
}

func (p *SimpleCustomAgent) suggestImprovement(params map[string]interface{}) (interface{}, error) {
	scenario := params["scenario"].(string)
	goal := "optimization"
	if g, ok := params["goal"].(string); ok {
		goal = g
	}

	suggestions := map[string]interface{}{
		"scenario":  scenario,
		"goal":      goal,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Generate suggestions based on keywords
	improvements := []string{}

	scenario_lower := strings.ToLower(scenario)

	if strings.Contains(scenario_lower, "slow") || strings.Contains(scenario_lower, "performance") {
		improvements = append(improvements, 
			"Enable SSD optimizations: boot.tmp.cleanOnBoot = true;",
			"Consider using zram: zramSwap.enable = true;",
			"Enable parallel builds: nix.settings.max-jobs = \"auto\";",
		)
	}

	if strings.Contains(scenario_lower, "security") {
		improvements = append(improvements,
			"Enable fail2ban: services.fail2ban.enable = true;",
			"Configure firewall: networking.firewall.enable = true;",
			"Disable root login: users.users.root.hashedPassword = \"!\";",
		)
	}

	if strings.Contains(scenario_lower, "development") || strings.Contains(scenario_lower, "coding") {
		improvements = append(improvements,
			"Use development shells: nix develop",
			"Consider direnv integration: programs.direnv.enable = true;",
			"Set up development tools: environment.systemPackages = with pkgs; [ git vim ];",
		)
	}

	if strings.Contains(scenario_lower, "desktop") {
		improvements = append(improvements,
			"Enable desktop environment: services.xserver.desktopManager.gnome.enable = true;",
			"Configure audio: sound.enable = true; hardware.pulseaudio.enable = true;",
			"Add desktop packages: environment.systemPackages = with pkgs; [ firefox libreoffice ];",
		)
	}

	if len(improvements) == 0 {
		improvements = append(improvements,
			"Keep system updated: nixos-rebuild switch --upgrade",
			"Use garbage collection: nix-collect-garbage -d",
			"Consider using flakes for reproducibility",
		)
	}

	suggestions["improvements"] = improvements
	suggestions["priority"] = p.prioritizeImprovements(improvements, goal)

	return suggestions, nil
}

// Helper methods
func (p *SimpleCustomAgent) calculateSeverity(issues []string) string {
	if len(issues) == 0 {
		return "none"
	}
	if len(issues) >= 3 {
		return "high"
	}
	if len(issues) >= 2 {
		return "medium"
	}
	return "low"
}

func (p *SimpleCustomAgent) prioritizeImprovements(improvements []string, goal string) map[string][]string {
	priority := map[string][]string{
		"high":   []string{},
		"medium": []string{},
		"low":    []string{},
	}

	for _, improvement := range improvements {
		if strings.Contains(improvement, "security") || strings.Contains(improvement, "firewall") {
			priority["high"] = append(priority["high"], improvement)
		} else if strings.Contains(improvement, "performance") || strings.Contains(improvement, "optimization") {
			priority["medium"] = append(priority["medium"], improvement)
		} else {
			priority["low"] = append(priority["low"], improvement)
		}
	}

	return priority
}

// Health and metrics
func (p *SimpleCustomAgent) HealthCheck(ctx context.Context) plugins.PluginHealth {
	status := "healthy"
	if !p.started {
		status = "stopped"
	}

	return plugins.PluginHealth{
		Status:    status,
		LastCheck: time.Now(),
		Message:   "Simple custom agent is operational",
	}
}

func (p *SimpleCustomAgent) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{
		ExecutionCount:   0,
		TotalExecutions:  0,
		AverageExecution: 0,
		LastExecution:    time.Time{},
		ErrorCount:       0,
		SuccessRate:      1.0,
	}
}

// Required plugin export
var Plugin SimpleCustomAgent

func NewPlugin() plugins.PluginInterface {
	return &Plugin
}