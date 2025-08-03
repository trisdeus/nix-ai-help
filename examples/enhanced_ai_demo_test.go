package main

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/advanced"
	"nix-ai-help/pkg/logger"
)

// testProvider is a simple AI provider for testing
type testProvider struct {
	response string
}

// Query implements the AIProvider interface
func (tp *testProvider) Query(prompt string) (string, error) {
	return tp.response, nil
}

// GenerateResponse implements the Provider interface
func (tp *testProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return tp.response, nil
}

// StreamResponse implements the Provider interface
func (tp *testProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- ai.StreamResponse{
			Content: tp.response,
			Done:    true,
		}
	}()
	return ch, nil
}

// GetPartialResponse implements the Provider interface
func (tp *testProvider) GetPartialResponse() string {
	return ""
}

func TestEnhancedAIIntegration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a test provider
	provider := &testProvider{
		response: `{
  "score": 0.85,
  "explanation": "High confidence in technical accuracy"
}`,
	}

	// Create advanced AI coordinator config
	config := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	// Create advanced AI coordinator
	coordinator := advanced.NewAdvancedAICoordinator(provider, log, config)

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

func TestEnhancedAIIntegrationEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a test provider with empty response
	provider := &testProvider{
		response: "",
	}

	// Create advanced AI coordinator config
	config := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	// Create advanced AI coordinator
	coordinator := advanced.NewAdvancedAICoordinator(provider, log, config)

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

func TestEnhancedAIConfiguration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a test provider
	provider := &testProvider{
		response: `{
  "score": 0.85,
  "explanation": "High confidence in technical accuracy"
}`,
	}

	// Test coordinator with reasoning disabled
	config1 := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  false,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	coordinator1 := advanced.NewAdvancedAICoordinator(provider, log, config1)
	if coordinator1.EnableReasoning() {
		t.Error("Expected reasoning to be disabled")
	}

	// Test coordinator with correction disabled
	config2 := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: false,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	coordinator2 := advanced.NewAdvancedAICoordinator(provider, log, config2)
	if coordinator2.EnableCorrection() {
		t.Error("Expected correction to be disabled")
	}

	// Test coordinator with planning disabled
	config3 := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   false,
		EnableScoring:    true,
	}

	coordinator3 := advanced.NewAdvancedAICoordinator(provider, log, config3)
	if coordinator3.EnablePlanning() {
		t.Error("Expected planning to be disabled")
	}

	// Test coordinator with scoring disabled
	config4 := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    false,
	}

	coordinator4 := advanced.NewAdvancedAICoordinator(provider, log, config4)
	if coordinator4.EnableScoring() {
		t.Error("Expected scoring to be disabled")
	}

	// Test coordinator with all features disabled
	config5 := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  false,
		EnableCorrection: false,
		EnablePlanning:   false,
		EnableScoring:    false,
	}

	coordinator5 := advanced.NewAdvancedAICoordinator(provider, log, config5)
	if coordinator5.EnableReasoning() || coordinator5.EnableCorrection() || 
		coordinator5.EnablePlanning() || coordinator5.EnableScoring() {
		t.Error("Expected all features to be disabled")
	}

	// Test coordinator with all features enabled
	config6 := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}

	coordinator6 := advanced.NewAdvancedAICoordinator(provider, log, config6)
	if !coordinator6.EnableReasoning() || !coordinator6.EnableCorrection() || 
		!coordinator6.EnablePlanning() || !coordinator6.EnableScoring() {
		t.Error("Expected all features to be enabled")
	}
}