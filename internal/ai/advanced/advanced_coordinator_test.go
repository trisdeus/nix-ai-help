package advanced

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// mockProvider5 is a mock AI provider for testing advanced AI coordination
type mockProvider5 struct {
	response string
}

func (mp *mockProvider5) Query(prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider5) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider5) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
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

func (mp *mockProvider5) GetPartialResponse() string {
	return ""
}

func TestAdvancedAICoordinator(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider5{
		response: `{
  "score": 0.85,
  "explanation": "High confidence in technical accuracy"
}`,
	}

	// Create advanced AI coordinator config
	config := AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	// Create advanced AI coordinator
	coordinator := NewAdvancedAICoordinator(provider, log, config)

	// Test processing a query
	ctx := context.Background()
	response, err := coordinator.ProcessQuery(ctx, "How to configure nginx in NixOS?")
	if err != nil {
		t.Fatalf("ProcessQuery failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Task != "How to configure nginx in NixOS?" {
		t.Errorf("Expected task 'How to configure nginx in NixOS?', got '%s'", response.Task)
	}

	if response.OriginalResponse == "" {
		t.Error("Expected original response, got empty string")
	}

	// Test formatting
	formatted := coordinator.FormatResponse(response)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}

	// Test with different query
	response2, err := coordinator.ProcessQuery(ctx, "What is the difference between flakes and channels?")
	if err != nil {
		t.Fatalf("ProcessQuery failed for second query: %v", err)
	}

	if response2 == nil {
		t.Fatal("Expected response for second query, got nil")
	}

	if response2.Task != "What is the difference between flakes and channels?" {
		t.Errorf("Expected task 'What is the difference between flakes and channels?', got '%s'", response2.Task)
	}
}

func TestAdvancedAICoordinatorEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider5{
		response: "",
	}

	// Create advanced AI coordinator config
	config := AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	// Create advanced AI coordinator
	coordinator := NewAdvancedAICoordinator(provider, log, config)

	// Test with empty query
	ctx := context.Background()
	response, err := coordinator.ProcessQuery(ctx, "")
	if err != nil {
		t.Fatalf("ProcessQuery failed with empty query: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response with empty query, got nil")
	}

	// Test with very long query
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	response2, err := coordinator.ProcessQuery(ctx, longQuery)
	if err != nil {
		t.Fatalf("ProcessQuery failed with long query: %v", err)
	}

	if response2 == nil {
		t.Fatal("Expected response with long query, got nil")
	}

	if response2.Task != longQuery {
		t.Errorf("Expected long query to be preserved, got '%s'", response2.Task)
	}

	// Test complex task detection
	isComplex := coordinator.IsComplexTask(longQuery)
	if !isComplex {
		t.Error("Expected long query to be detected as complex")
	}

	// Test simple task detection
	isSimpleComplex := coordinator.IsComplexTask("help")
	if isSimpleComplex {
		t.Error("Expected simple query not to be detected as complex")
	}
}

func TestAdvancedAICoordinatorConfiguration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider5{
		response: `{
  "score": 0.85,
  "explanation": "High confidence in technical accuracy"
}`,
	}

	// Test coordinator with reasoning disabled
	config1 := AdvancedAICoordinatorConfig{
		EnableReasoning:  false,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	coordinator1 := NewAdvancedAICoordinator(provider, log, config1)
	if coordinator1.enableReasoning {
		t.Error("Expected reasoning to be disabled")
	}

	// Test coordinator with correction disabled
	config2 := AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: false,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	coordinator2 := NewAdvancedAICoordinator(provider, log, config2)
	if coordinator2.enableCorrection {
		t.Error("Expected correction to be disabled")
	}

	// Test coordinator with planning disabled
	config3 := AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   false,
		EnableScoring:    true,
	}

	coordinator3 := NewAdvancedAICoordinator(provider, log, config3)
	if coordinator3.enablePlanning {
		t.Error("Expected planning to be disabled")
	}

	// Test coordinator with scoring disabled
	config4 := AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    false,
	}

	coordinator4 := NewAdvancedAICoordinator(provider, log, config4)
	if coordinator4.enableScoring {
		t.Error("Expected scoring to be disabled")
	}

	// Test coordinator with all features disabled
	config5 := AdvancedAICoordinatorConfig{
		EnableReasoning:  false,
		EnableCorrection: false,
		EnablePlanning:   false,
		EnableScoring:    false,
	}

	coordinator5 := NewAdvancedAICoordinator(provider, log, config5)
	if coordinator5.enableReasoning || coordinator5.enableCorrection || 
		coordinator5.enablePlanning || coordinator5.enableScoring {
		t.Error("Expected all features to be disabled")
	}

	// Test coordinator with all features enabled
	config6 := AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	coordinator6 := NewAdvancedAICoordinator(provider, log, config6)
	if !coordinator6.enableReasoning || !coordinator6.enableCorrection || 
		!coordinator6.enablePlanning || !coordinator6.enableScoring {
		t.Error("Expected all features to be enabled")
	}
}