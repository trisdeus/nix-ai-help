package reasoning

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// mockProvider2 is a mock AI provider for testing
type mockProvider2 struct {
	response string
}

func (mp *mockProvider2) Query(prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider2) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider2) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- ai.StreamResponse{
			Content: mp.response,
			Done:    true,
		}
	}()
	return ch, nil
}

func (mp *mockProvider2) GetPartialResponse() string {
	return ""
}

func TestSelfReflector(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider2{
		response: `{
			"components": [
				{
					"id": "comp-1",
					"title": "Configuration Component",
					"description": "NixOS configuration setup"
				}
			]
		}`,
	}

	// Create self-reflector
	reflector := NewSelfReflector(provider, log)

	// Test self-reflection
	ctx := context.Background()
	reflection, err := reflector.Reflect(ctx, 
		"How to configure nginx in NixOS?", 
		"Enable nginx service in configuration.nix")

	if err != nil {
		t.Fatalf("Self-reflection failed: %v", err)
	}

	if reflection == nil {
		t.Fatal("Expected self-reflection, got nil")
	}

	if reflection.OriginalPrompt != "How to configure nginx in NixOS?" {
		t.Errorf("Expected original prompt 'How to configure nginx in NixOS?', got '%s'", reflection.OriginalPrompt)
	}

	if reflection.OriginalAnswer != "Enable nginx service in configuration.nix" {
		t.Errorf("Expected original answer 'Enable nginx service in configuration.nix', got '%s'", reflection.OriginalAnswer)
	}

	if len(reflection.Reflections) == 0 {
		t.Error("Expected reflections, got none")
	}

	if len(reflection.Improvements) == 0 {
		t.Error("Expected improvements, got none")
	}

	// Test formatting
	formatted := reflector.FormatSelfReflection(reflection)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}

	// Test with different query
	reflection2, err := reflector.Reflect(ctx,
		"What is the difference between flakes and channels?",
		"Flakes are modern, channels are legacy")

	if err != nil {
		t.Fatalf("Self-reflection failed for second query: %v", err)
	}

	if reflection2 == nil {
		t.Fatal("Expected self-reflection for second query, got nil")
	}

	if reflection2.OriginalPrompt != "What is the difference between flakes and channels?" {
		t.Errorf("Expected query 'What is the difference between flakes and channels?', got '%s'", reflection2.OriginalPrompt)
	}
}

func TestSelfReflectorEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider2{
		response: "",
	}

	// Create self-reflector
	reflector := NewSelfReflector(provider, log)

	// Test with empty query and answer
	ctx := context.Background()
	reflection, err := reflector.Reflect(ctx, "", "")

	if err != nil {
		t.Fatalf("Self-reflection failed with empty inputs: %v", err)
	}

	if reflection == nil {
		t.Fatal("Expected self-reflection with empty inputs, got nil")
	}

	// Test with very long query and answer
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	longAnswer := "This is a correspondingly long answer that provides detailed information about configuring NixOS for complex scenarios involving multiple subsystems, services, and hardware components. The answer addresses package management strategies, service configuration patterns, hardware-specific settings, network configurations, security hardening measures, and performance optimization techniques."

	reflection2, err := reflector.Reflect(ctx, longQuery, longAnswer)

	if err != nil {
		t.Fatalf("Self-reflection failed with long inputs: %v", err)
	}

	if reflection2 == nil {
		t.Fatal("Expected self-reflection with long inputs, got nil")
	}

	if reflection2.OriginalPrompt != longQuery {
		t.Errorf("Expected long query to be preserved, got '%s'", reflection2.OriginalPrompt)
	}

	if reflection2.OriginalAnswer != longAnswer {
		t.Errorf("Expected long answer to be preserved, got '%s'", reflection2.OriginalAnswer)
	}
}