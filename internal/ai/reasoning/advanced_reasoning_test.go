package reasoning

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// mockProvider is a mock AI provider for testing
type mockProvider struct {
	response string
}

func (mp *mockProvider) Query(prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
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

func (mp *mockProvider) GetPartialResponse() string {
	return ""
}

func TestAdvancedReasoner(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider{
		response: `{"components":[{"id":"comp-1","title":"Main Component","description":"Test component"}]}`,
	}

	// Create advanced reasoner
	reasoner := NewAdvancedReasoner(provider, log)

	// Test generating a reasoning chain
	ctx := context.Background()
	chain, err := reasoner.GenerateReasoningChain(ctx, "How to configure nginx in NixOS?")
	if err != nil {
		t.Fatalf("GenerateReasoningChain failed: %v", err)
	}

	if chain == nil {
		t.Fatal("Expected reasoning chain, got nil")
	}

	if chain.Query != "How to configure nginx in NixOS?" {
		t.Errorf("Expected query 'How to configure nginx in NixOS?', got '%s'", chain.Query)
	}

	if len(chain.Steps) == 0 {
		t.Error("Expected reasoning steps, got none")
	}

	// Test formatting
	formatted := reasoner.FormatReasoningChain(chain)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}

	// Test with different query
	chain2, err := reasoner.GenerateReasoningChain(ctx, "What is the difference between flakes and channels?")
	if err != nil {
		t.Fatalf("GenerateReasoningChain failed for second query: %v", err)
	}

	if chain2 == nil {
		t.Fatal("Expected reasoning chain for second query, got nil")
	}

	if chain2.Query != "What is the difference between flakes and channels?" {
		t.Errorf("Expected query 'What is the difference between flakes and channels?', got '%s'", chain2.Query)
	}
}

func TestAdvancedReasonerEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider{
		response: "",
	}

	// Create advanced reasoner
	reasoner := NewAdvancedReasoner(provider, log)

	// Test with empty query
	ctx := context.Background()
	chain, err := reasoner.GenerateReasoningChain(ctx, "")
	if err != nil {
		t.Fatalf("GenerateReasoningChain failed with empty query: %v", err)
	}

	if chain == nil {
		t.Fatal("Expected reasoning chain with empty query, got nil")
	}

	// Test with very long query
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	chain2, err := reasoner.GenerateReasoningChain(ctx, longQuery)
	if err != nil {
		t.Fatalf("GenerateReasoningChain failed with long query: %v", err)
	}

	if chain2 == nil {
		t.Fatal("Expected reasoning chain with long query, got nil")
	}

	if chain2.Query != longQuery {
		t.Errorf("Expected long query to be preserved, got '%s'", chain2.Query)
	}
}