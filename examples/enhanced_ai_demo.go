package main

import (
	"context"
	"fmt"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/advanced"
	"nix-ai-help/pkg/logger"
)

// SimpleProvider is a minimal AI provider for demonstration
type SimpleProvider struct {
	name  string
	model string
}

// Query implements the AIProvider interface
func (sp *SimpleProvider) Query(prompt string) (string, error) {
	return fmt.Sprintf("Response to: %s", prompt), nil
}

// GenerateResponse implements the Provider interface
func (sp *SimpleProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Simple response generation based on prompt content
	switch {
	case contains(prompt, "nginx"):
		return `{"score": 0.85, "explanation": "To enable nginx in NixOS, add the following to your configuration.nix:\n\nservices.nginx = {\n  enable = true;\n  virtualHosts = {\n    \"localhost\" = {\n      root = \"/var/www\";\n    };\n  };\n};\n\nThen rebuild your system with:\nsudo nixos-rebuild switch"}`, nil
	case contains(prompt, "package"):
		return `{"score": 0.90, "explanation": "To install a package in NixOS, add it to environment.systemPackages in your configuration.nix:\n\nenvironment.systemPackages = with pkgs; [\n  firefox\n  git\n  vim\n];\n\nAlternatively, for temporary installation:\nnix-env -iA nixpkgs.firefox\n\nFor modern flakes-based systems:\nnix profile install nixpkgs#firefox"}`, nil
	default:
		return `{"score": 0.75, "explanation": "General NixOS configuration guidance. Please be more specific for detailed help."}`, nil
	}
}

// StreamResponse implements the Provider interface
func (sp *SimpleProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- ai.StreamResponse{
			Content: fmt.Sprintf("Streaming response to: %s", prompt),
			Done:    true,
			}
	}()
	return ch, nil
}

// GetPartialResponse implements the Provider interface
func (sp *SimpleProvider) GetPartialResponse() string {
	return ""
}

// contains checks if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	s = fmt.Sprintf("%s", s)
	substr = fmt.Sprintf("%s", substr)
	return len(s) >= len(substr) && 
		(len(s) == len(substr) && s == substr ||
		 len(s) > len(substr) && (s[:len(substr)] == substr || 
			s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func main() {
	// Create a logger
	log := logger.NewLogger()

	// Create a simple AI provider
	provider := &SimpleProvider{
		name:  "simple-ai",
		model: "simple-model",
	}

	// Create advanced AI coordinator with all features enabled
	config := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}
	
	coordinator := advanced.NewAdvancedAICoordinator(provider, log, config)

	// Example 1: Enable nginx in NixOS
	fmt.Println("=== Example 1: Enable nginx in NixOS ===")
	ctx := context.Background()
	response, err := coordinator.ProcessQuery(ctx, "How to enable nginx in NixOS?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(coordinator.FormatResponse(response))
	fmt.Println()

	// Example 2: Install a package
	fmt.Println("=== Example 2: Install a package ===")
	response, err = coordinator.ProcessQuery(ctx, "How to install firefox in NixOS?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(coordinator.FormatResponse(response))
	fmt.Println()

	// Example 3: Complex task with planning
	fmt.Println("=== Example 3: Complex task with planning ===")
	response, err = coordinator.ProcessQuery(ctx, "How to set up a development environment for Python and Django?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(coordinator.FormatResponse(response))
	fmt.Println()

	// Example 4: Edge case with self-correction
	fmt.Println("=== Example 4: Edge case with self-correction ===")
	response, err = coordinator.ProcessQuery(ctx, "")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(coordinator.FormatResponse(response))
	fmt.Println()

	fmt.Println("All examples completed successfully!")
}