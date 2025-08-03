package reasoning

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// mockProvider3 is a mock AI provider for testing self-improvement
type mockProvider3 struct {
	response string
}

func (mp *mockProvider3) Query(prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider3) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider3) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
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

func (mp *mockProvider3) GetPartialResponse() string {
	return ""
}

func TestSelfImprover(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider3{
		response: `{
			"steps": [
				{
					"id": "step-1",
					"type": "accuracy",
					"description": "Fix technical inaccuracy in command syntax",
					"action": "Replace 'nix-env -i' with 'nix profile install'",
					"justification": "nix-env is deprecated, newer Nix versions prefer nix profile",
					"confidence": 0.95
				}
			]
		}`,
	}

	// Create self-improver
	improver := NewSelfImprover(provider, log)

	// Test self-improvement
	ctx := context.Background()
	improvement, err := improver.Improve(ctx,
		"How to install packages in NixOS?",
		"Use nix-env -i to install packages",
		"The response uses deprecated commands. Please use nix profile install instead.")

	if err != nil {
		t.Fatalf("Self-improvement failed: %v", err)
	}

	if improvement == nil {
		t.Fatal("Expected self-improvement, got nil")
	}

	if improvement.OriginalPrompt != "How to install packages in NixOS?" {
		t.Errorf("Expected original prompt 'How to install packages in NixOS?', got '%s'", improvement.OriginalPrompt)
	}

	if improvement.OriginalAnswer != "Use nix-env -i to install packages" {
		t.Errorf("Expected original answer 'Use nix-env -i to install packages', got '%s'", improvement.OriginalAnswer)
	}

	if improvement.Feedback != "The response uses deprecated commands. Please use nix profile install instead." {
		t.Errorf("Expected feedback 'The response uses deprecated commands. Please use nix profile install instead.', got '%s'", improvement.Feedback)
	}

	if len(improvement.ImprovementPlan) == 0 {
		t.Error("Expected improvement plan, got none")
	}

	if len(improvement.AppliedSteps) == 0 {
		t.Error("Expected applied steps, got none")
	}

	if improvement.FinalAnswer == "" {
		t.Error("Expected improved answer, got empty string")
	}

	if improvement.Confidence <= 0 {
		t.Errorf("Expected positive confidence, got %f", improvement.Confidence)
	}

	// Test formatting
	formatted := improver.FormatSelfImprovement(improvement)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}

	// Test with different query
	improvement2, err := improver.Improve(ctx,
		"What is the difference between flakes and channels?",
		"Flakes are modern, channels are legacy",
		"Please provide more detailed comparison with examples")

	if err != nil {
		t.Fatalf("Self-improvement failed for second query: %v", err)
	}

	if improvement2 == nil {
		t.Fatal("Expected self-improvement for second query, got nil")
	}

	if improvement2.OriginalPrompt != "What is the difference between flakes and channels?" {
		t.Errorf("Expected prompt 'What is the difference between flakes and channels?', got '%s'", improvement2.OriginalPrompt)
	}

	if improvement2.Feedback != "Please provide more detailed comparison with examples" {
		t.Errorf("Expected feedback 'Please provide more detailed comparison with examples', got '%s'", improvement2.Feedback)
	}
}

func TestSelfImproverEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider3{
		response: "",
	}

	// Create self-improver
	improver := NewSelfImprover(provider, log)

	// Test with empty inputs
	ctx := context.Background()
	improvement, err := improver.Improve(ctx, "", "", "")

	if err != nil {
		t.Fatalf("Self-improvement failed with empty inputs: %v", err)
	}

	if improvement == nil {
		t.Fatal("Expected self-improvement with empty inputs, got nil")
	}

	// Test with very long inputs
	longPrompt := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	longAnswer := "This is a correspondingly long answer that provides detailed information about configuring NixOS for complex scenarios involving multiple subsystems, services, and hardware components. The answer addresses package management strategies, service configuration patterns, hardware-specific settings, network configurations, security hardening measures, and performance optimization techniques."
	longFeedback := "This is a very detailed feedback that points out multiple issues with the response, including lack of specific examples, missing technical details, and insufficient context for beginners."

	improvement2, err := improver.Improve(ctx, longPrompt, longAnswer, longFeedback)

	if err != nil {
		t.Fatalf("Self-improvement failed with long inputs: %v", err)
	}

	if improvement2 == nil {
		t.Fatal("Expected self-improvement with long inputs, got nil")
	}

	if improvement2.OriginalPrompt != longPrompt {
		t.Errorf("Expected long prompt to be preserved, got '%s'", improvement2.OriginalPrompt)
	}

	if improvement2.OriginalAnswer != longAnswer {
		t.Errorf("Expected long answer to be preserved, got '%s'", improvement2.OriginalAnswer)
	}

	if improvement2.Feedback != longFeedback {
		t.Errorf("Expected long feedback to be preserved, got '%s'", improvement2.Feedback)
	}
}