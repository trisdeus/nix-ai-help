package main

import (
	"context"
	"fmt"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/advanced"
	"nix-ai-help/pkg/logger"
)

// ExampleAIProvider is a simple AI provider for demonstration
type ExampleAIProvider struct {
	name  string
	model string
}

// Query implements the AIProvider interface
func (eap *ExampleAIProvider) Query(prompt string) (string, error) {
	return fmt.Sprintf("Example response to: %s", prompt), nil
}

// GenerateResponse implements the Provider interface
func (eap *ExampleAIProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Simple response generation based on prompt content
	switch {
	case contains(prompt, "nginx"):
		return `{    "score": 0.85,    "explanation": "To enable nginx in NixOS, add the following to your configuration.nix:  \nservices.nginx = {    enable = true;    virtualHosts = {      \"localhost\" = {        root = \"/var/www\";      };    };  };\n\nThen rebuild your system with:\nsudo nixos-rebuild switch"}`, nil
	case contains(prompt, "package"):
		return `{    "score": 0.90,    "explanation": "To install a package in NixOS, add it to environment.systemPackages in your configuration.nix:  \nenvironment.systemPackages = with pkgs; [    firefox    git    vim  ];\n\nAlternatively, for temporary installation:\nnix-env -iA nixpkgs.firefox\n\nFor modern flakes-based systems:\nnix profile install nixpkgs#firefox"}`, nil
	default:
		return `{    "score": 0.75,    "explanation": "General NixOS configuration guidance. Please be more specific for detailed help."}`, nil
	}
}

// StreamResponse implements the Provider interface
func (eap *ExampleAIProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
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
func (eap *ExampleAIProvider) GetPartialResponse() string {
	return "Partial response"
}

// GetName returns the provider name
func (eap *ExampleAIProvider) GetName() string {
	return eap.name
}

// GetModel returns the model name
func (eap *ExampleAIProvider) GetModel() string {
	return eap.model
}

// SetTimeout sets the timeout for the provider
func (eap *ExampleAIProvider) SetTimeout(timeout int) {
	// No-op for example
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

	// Create an example AI provider
	provider := &ExampleAIProvider{
		name:  "example-ai",
		model: "example-model",
	}

	// Create advanced AI coordinator
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