package reasoning

import (
	"context"
	"fmt"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// mockProvider4 is a mock AI provider for testing confidence scoring
type mockProvider4 struct {
	responses map[string]string
	logger    *logger.Logger
}

func (mp *mockProvider4) Query(prompt string) (string, error) {
	// Return a default response if no specific one is found
	if response, exists := mp.responses[prompt]; exists {
		return response, nil
	}
	
	return `{
  "score": 0.75,
  "explanation": "Default mock response"
}`, nil
}

func (mp *mockProvider4) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Return a default response if no specific one is found
	if response, exists := mp.responses[prompt]; exists {
		// Ensure response is valid JSON
		mp.validateJSONResponse(response)
		return response, nil
	}
	
	return `{
  "score": 0.75,
  "explanation": "Default mock response"
}`, nil
}

func (mp *mockProvider4) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		
		// Return a default response if no specific one is found
		response := `{
  "score": 0.75,
  "explanation": "Default mock response"
}`
		if res, exists := mp.responses[prompt]; exists {
			// Ensure response is valid JSON
			mp.validateJSONResponse(res)
			response = res
		}
		
		ch <- ai.StreamResponse{
			Content: response,
			Done:    true,
		}
	}()
	return ch, nil
}

// validateJSONResponse validates that a response is valid JSON
func (mp *mockProvider4) validateJSONResponse(response string) {
	// In a real implementation, we would validate the JSON here
	// For now, we'll just log it for debugging
	mp.logger.Debug(fmt.Sprintf("Mock provider response: %s", response))
}

func (mp *mockProvider4) GetPartialResponse() string {
	return ""
}

func TestConfidenceScorer(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
		// Create a mock provider
	provider := &mockProvider4{
		responses: map[string]string{
			"Evaluate the technical accuracy of the following AI response to the query:\nQuery: \"How to configure nginx in NixOS?\"\nResponse: \"Enable nginx service in configuration.nix\"\n\nConsider these aspects:\n1. Are the commands and configuration options mentioned correct?\n2. Are there any factual errors in the response?\n3. Does the response align with current NixOS best practices?\n4. Are technical terms used correctly?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the accuracy evaluation\"\n}": `{\n  \"score\": 0.85,\n  \"explanation\": \"High confidence in technical accuracy\"\n}`,
			"Evaluate the clarity and readability of the following AI response to the query:\nQuery: \"How to configure nginx in NixOS?\"\nResponse: \"Enable nginx service in configuration.nix\"\n\nConsider these aspects:\n1. Is the language clear and understandable?\n2. Are complex concepts explained well?\n3. Is the structure logical and easy to follow?\n4. Are there any confusing or ambiguous sections?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the clarity evaluation\"\n}": `{\n  \"score\": 0.75,\n  \"explanation\": \"Good clarity but could use more detail\"\n}`,
			"Evaluate the completeness of the following AI response to the query:\nQuery: \"How to configure nginx in NixOS?\"\nResponse: \"Enable nginx service in configuration.nix\"\n\nConsider these aspects:\n1. Does the response fully address the question?\n2. Are all relevant aspects covered?\n3. Are there any important omissions?\n4. Is additional information needed for a complete understanding?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the completeness evaluation\"\n}": `{\n  \"score\": 0.65,\n  \"explanation\": \"Could be more complete with examples and context\"\n}`,
			"Evaluate the relevance of the following AI response to the query:\nQuery: \"How to configure nginx in NixOS?\"\nResponse: \"Enable nginx service in configuration.nix\"\n\nConsider these aspects:\n1. Does the response directly address the question asked?\n2. Are there any irrelevant sections or tangents?\n3. Is the focus maintained on the main topic?\n4. Are examples and analogies appropriately chosen?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the relevance evaluation\"\n}": `{\n  \"score\": 0.90,\n  \"explanation\": \"Highly relevant to the question\"\n}`,
			"Evaluate the logical correctness of the following AI response to the query:\nQuery: \"How to configure nginx in NixOS?\"\nResponse: \"Enable nginx service in configuration.nix\"\n\nConsider these aspects:\n1. Are the logical steps sound and valid?\n2. Are conclusions properly supported by evidence?\n3. Are there any logical fallacies or inconsistencies?\n4. Is the reasoning internally consistent?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the correctness evaluation\"\n}": `{\n  \"score\": 0.80,\n  \"explanation\": \"Logically sound but could be expanded\"\n}`,
			"Evaluate the helpfulness of the following AI response to the query:\nQuery: \"How to configure nginx in NixOS?\"\nResponse: \"Enable nginx service in configuration.nix\"\n\nConsider these aspects:\n1. Does the response actually help the user accomplish their goal?\n2. Is the information actionable and practical?\n3. Are examples and suggestions concrete and useful?\n4. Would a beginner find this response helpful?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the helpfulness evaluation\"\n}": `{\n  \"score\": 0.70,\n  \"explanation\": \"Helpful but could include more actionable steps\"\n}`,
		},
		logger: log,
	}

	// Create confidence scorer
	scorer := NewConfidenceScorer(provider, log)

	// Test evaluating a response
	ctx := context.Background()
	score, err := scorer.EvaluateResponse(ctx,
		"How to configure nginx in NixOS?",
		"Enable nginx service in configuration.nix")

	if err != nil {
		t.Fatalf("EvaluateResponse failed: %v", err)
	}

	if score == nil {
		t.Fatal("Expected confidence score, got nil")
	}

	if score.Query != "How to configure nginx in NixOS?" {
		t.Errorf("Expected query 'How to configure nginx in NixOS?', got '%s'", score.Query)
	}

	if score.Response != "Enable nginx service in configuration.nix" {
		t.Errorf("Expected response 'Enable nginx service in configuration.nix', got '%s'", score.Response)
	}

	if score.OverallScore <= 0 {
		t.Errorf("Expected positive overall score, got %f", score.OverallScore)
	}

	if len(score.Dimensions) == 0 {
		t.Error("Expected dimensions, got none")
	}

	// Test formatting
	formatted := scorer.FormatDetailedConfidenceScore(score)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}

	// Test with different query
	score2, err := scorer.EvaluateResponse(ctx,
		"What is the difference between flakes and channels?",
		"Flakes are modern, channels are legacy")

	if err != nil {
		t.Fatalf("EvaluateResponse failed for second query: %v", err)
	}

	if score2 == nil {
		t.Fatal("Expected confidence score for second query, got nil")
	}

	if score2.Query != "What is the difference between flakes and channels?" {
		t.Errorf("Expected query 'What is the difference between flakes and channels?', got '%s'", score2.Query)
	}
}

func TestConfidenceScorerEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with long query response
	provider := &mockProvider4{
		responses: map[string]string{
			"Evaluate the technical accuracy of the following AI response to the query:\nQuery: \"This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization.\"\nResponse: \"This is a correspondingly long response that provides detailed information about configuring NixOS for complex scenarios involving multiple subsystems, services, and hardware components. The response addresses package management strategies, service configuration patterns, hardware-specific settings, network configurations, security hardening measures, and performance optimization techniques.\"\n\nConsider these aspects:\n1. Are the commands and configuration options mentioned correct?\n2. Are there any factual errors in the response?\n3. Does the response align with current NixOS best practices?\n4. Are technical terms used correctly?\n\nReturn your evaluation in this JSON format:\n{\n  \"score\": 0.85,\n  \"explanation\": \"Detailed explanation of the accuracy evaluation\"\n}": `{\n  \"score\": 0.85,\n  \"explanation\": \"High confidence in technical accuracy for long query\"\n}`,
		},
		logger: log,
	}

	// Create confidence scorer
	scorer := NewConfidenceScorer(provider, log)

	// Test with empty query and response
	ctx := context.Background()
	score, err := scorer.EvaluateResponse(ctx, "", "")

	if err != nil {
		t.Fatalf("EvaluateResponse failed with empty inputs: %v", err)
	}

	if score == nil {
		t.Fatal("Expected confidence score with empty inputs, got nil")
	}

	// Test with very long query and response
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	longResponse := "This is a correspondingly long response that provides detailed information about configuring NixOS for complex scenarios involving multiple subsystems, services, and hardware components. The response addresses package management strategies, service configuration patterns, hardware-specific settings, network configurations, security hardening measures, and performance optimization techniques."

	score2, err := scorer.EvaluateResponse(ctx, longQuery, longResponse)

	if err != nil {
		t.Fatalf("EvaluateResponse failed with long inputs: %v", err)
	}

	if score2 == nil {
		t.Fatal("Expected confidence score with long inputs, got nil")
	}

	if score2.Query != longQuery {
		t.Errorf("Expected long query to be preserved, got '%s'", score2.Query)
	}

	if score2.Response != longResponse {
		t.Errorf("Expected long response to be preserved, got '%s'", score2.Response)
	}

	// Test calculateOverallScore with empty dimensions
	emptyScore := scorer.calculateOverallScore([]ConfidenceDimension{})
	if emptyScore != 0.0 {
		t.Errorf("Expected 0.0 for empty dimensions, got %f", emptyScore)
	}

	// Test calculateOverallScore with zero weights
	zeroWeightDimensions := []ConfidenceDimension{
		{
			Name:        "test",
			Description: "test dimension",
			Weight:      0.0,
			Score:       0.8,
		},
	}
	
	zeroWeightScore := scorer.calculateOverallScore(zeroWeightDimensions)
	if zeroWeightScore != 0.0 {
		t.Errorf("Expected 0.0 for zero weights, got %f", zeroWeightScore)
	}

	// Test generateRecommendations with low scores
	lowScoreDimensions := []ConfidenceDimension{
		{
			Name:        "accuracy",
			Description: "Technical accuracy",
			Weight:      0.25,
			Score:       0.3,
		},
		{
			Name:        "clarity",
			Description: "Clarity and readability",
			Weight:      0.15,
			Score:       0.4,
		},
	}
	
	recommendations := scorer.generateRecommendations(lowScoreDimensions)
	if len(recommendations) == 0 {
		t.Error("Expected recommendations for low-score dimensions, got none")
	}

	// Test generateWarnings with low scores
	warnings := scorer.generateWarnings(lowScoreDimensions)
	if len(warnings) == 0 {
		t.Error("Expected warnings for low-score dimensions, got none")
	}

	// Test removeDuplicates
	duplicates := []string{"item1", "item2", "item1", "item3", "item2"}
	unique := scorer.removeDuplicates(duplicates)
	if len(unique) != 3 {
		t.Errorf("Expected 3 unique items, got %d", len(unique))
	}
}