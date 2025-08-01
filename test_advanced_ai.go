package main

import (
	"context"
	"fmt"
	"os"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/advanced"
	"nix-ai-help/pkg/logger"
)

func main() {
	// Create a simple provider for testing
	provider := &simpleProvider{
		response: `To enable nginx in NixOS, add the following to your configuration.nix:
		
		services.nginx.enable = true;
		services.nginx.virtualHosts."localhost" = {
		  root = "/var/www";
		};

		Then rebuild your system with:
		
		sudo nixos-rebuild switch`,
	}
	
	log := logger.NewLogger()
	
	// Create advanced AI coordinator
	config := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}
	
	coordinator := advanced.NewAdvancedAICoordinator(provider, log, config)
	
	// Process a sample query
	ctx := context.Background()
	response, err := coordinator.ProcessQuery(ctx, "How to enable nginx in NixOS?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Format and print the response
	fmt.Println(coordinator.FormatResponse(response))
}

// simpleProvider is a simple provider for testing
type simpleProvider struct {
	response string
}

func (sp *simpleProvider) Query(prompt string) (string, error) {
	return sp.response, nil
}

func (sp *simpleProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return sp.response, nil
}

func (sp *simpleProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- ai.StreamResponse{
			Content: sp.response,
			Done:    true,
		}
	}()
	return ch, nil
}

func (sp *simpleProvider) GetPartialResponse() string {
	return ""
}