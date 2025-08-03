package cli

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/advanced"
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

func TestEnhancedAIIntegration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider2{
		response: "Mock AI response to test query",
	}

	// Create enhanced AI integration with advanced features disabled
	eaiDisabled := NewEnhancedAIIntegration(provider, log, false)

	// Test that integration was created successfully
	if eaiDisabled == nil {
		t.Fatal("Expected enhanced AI integration with advanced features disabled, got nil")
	}

	// Create enhanced AI integration with advanced features enabled
	eaiEnabled := NewEnhancedAIIntegration(provider, log, true)

	// Test that integration was created successfully
	if eaiEnabled == nil {
		t.Fatal("Expected enhanced AI integration with advanced features enabled, got nil")
	}

	// Test processing query with advanced features disabled (fallback to regular)
	ctx := context.Background()
	responseDisabled, err := eaiDisabled.ProcessQueryWithAdvancedAI(ctx, "Test query")
	if err != nil {
		t.Fatalf("ProcessQueryWithAdvancedAI failed with advanced features disabled: %v", err)
	}

	if responseDisabled == "" {
		t.Error("Expected response with advanced features disabled, got empty string")
	}

	// Test processing query with advanced features enabled
	responseEnabled, err := eaiEnabled.ProcessQueryWithAdvancedAI(ctx, "Test query")
	if err != nil {
		t.Fatalf("ProcessQueryWithAdvancedAI failed with advanced features enabled: %v", err)
	}

	if responseEnabled == "" {
		t.Error("Expected response with advanced features enabled, got empty string")
	}

	// Test getting enhanced AI status with features disabled
	statusDisabled := eaiDisabled.GetEnhancedAIStatus()
	if statusDisabled == "" {
		t.Error("Expected status with advanced features disabled, got empty string")
	}

	// Test getting enhanced AI status with features enabled
	statusEnabled := eaiEnabled.GetEnhancedAIStatus()
	if statusEnabled == "" {
		t.Error("Expected status with advanced features enabled, got empty string")
	}

	// Test formatting enhanced AI response with nil input
	formattedNil := eaiEnabled.FormatEnhancedAIResponse(nil)
	if formattedNil == "" {
		t.Error("Expected formatted response for nil input, got empty string")
	}

	// Test formatting enhanced AI response with valid input
	response := &advanced.AdvancedAIResponse{
		Task:      "Test query",
		Timestamp: "2025-08-01 15:30:45",
	}
	
	formatted := eaiEnabled.FormatEnhancedAIResponse(response)
	if formatted == "" {
		t.Error("Expected formatted response, got empty string")
	}
}

func TestEnhancedAIIntegrationEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider2{
		response: "",
	}

	// Create enhanced AI integration with advanced features enabled
	eai := NewEnhancedAIIntegration(provider, log, true)

	// Test processing query with empty input
	ctx := context.Background()
	response, err := eai.ProcessQueryWithAdvancedAI(ctx, "")
	if err != nil {
		t.Fatalf("ProcessQueryWithAdvancedAI failed with empty input: %v", err)
	}

	if response == "" {
		t.Log("Response is empty for empty query (expected in test environment)")
	}

	// Test processing query with very long input
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	response2, err := eai.ProcessQueryWithAdvancedAI(ctx, longQuery)
	if err != nil {
		t.Fatalf("ProcessQueryWithAdvancedAI failed with long input: %v", err)
	}

	if response2 == "" {
		t.Log("Response is empty for long query (expected in test environment)")
	}

	// Test processing query with nil context
	response3, err := eai.ProcessQueryWithAdvancedAI(nil, "Test query")
	if err != nil {
		t.Fatalf("ProcessQueryWithAdvancedAI failed with nil context: %v", err)
	}

	if response3 == "" {
		t.Log("Response is empty for nil context (expected in test environment)")
	}

	// Test getting enhanced AI status
	status := eai.GetEnhancedAIStatus()
	if status == "" {
		t.Error("Expected status, got empty string")
	}

	// Test formatting enhanced AI response with invalid input
	formatted := eai.FormatEnhancedAIResponse(nil)
	if formatted == "" {
		t.Error("Expected formatted response for nil input, got empty string")
	}
}

func TestEnhancedAIIntegrationConfiguration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create enhanced AI integrations with different configurations
	cfg1 := &mockProvider2{
		response: "Mock response 1",
	}
	eai1 := NewEnhancedAIIntegration(cfg1, log, false)
	if eai1 == nil {
		t.Error("Expected enhanced AI integration with config 1, got nil")
	}

	cfg2 := &mockProvider2{
		response: "Mock response 2",
	}
	eai2 := NewEnhancedAIIntegration(cfg2, log, true)
	if eai2 == nil {
		t.Error("Expected enhanced AI integration with config 2, got nil")
	}

	cfg3 := &mockProvider2{
		response: "Mock response 3",
	}
	eai3 := NewEnhancedAIIntegration(cfg3, log, false)
	if eai3 == nil {
		t.Error("Expected enhanced AI integration with config 3, got nil")
	}

	// Test processing queries with different configurations
	ctx := context.Background()
	
	_, err1 := eai1.ProcessQueryWithAdvancedAI(ctx, "Test query 1")
	if err1 != nil {
		t.Logf("ProcessQueryWithAdvancedAI with config 1 failed: %v", err1)
	}

	_, err2 := eai2.ProcessQueryWithAdvancedAI(ctx, "Test query 2")
	if err2 != nil {
		t.Logf("ProcessQueryWithAdvancedAI with config 2 failed: %v", err2)
	}

	_, err3 := eai3.ProcessQueryWithAdvancedAI(ctx, "Test query 3")
	if err3 != nil {
		t.Logf("ProcessQueryWithAdvancedAI with config 3 failed: %v", err3)
	}

	// Test getting enhanced AI status with different configurations
	status1 := eai1.GetEnhancedAIStatus()
	if status1 == "" {
		t.Error("Expected status with config 1, got empty string")
	}

	status2 := eai2.GetEnhancedAIStatus()
	if status2 == "" {
		t.Error("Expected status with config 2, got empty string")
	}

	status3 := eai3.GetEnhancedAIStatus()
	if status3 == "" {
		t.Error("Expected status with config 3, got empty string")
	}

	// Test formatting enhanced AI response with different configurations
	response := &advanced.AdvancedAIResponse{
		Task:      "Test query",
		Timestamp: "2025-08-01 15:30:45",
	}
	
	formatted1 := eai1.FormatEnhancedAIResponse(response)
	if formatted1 == "" {
		t.Error("Expected formatted response with config 1, got empty string")
	}

	formatted2 := eai2.FormatEnhancedAIResponse(response)
	if formatted2 == "" {
		t.Error("Expected formatted response with config 2, got empty string")
	}

	formatted3 := eai3.FormatEnhancedAIResponse(response)
	if formatted3 == "" {
		t.Error("Expected formatted response with config 3, got empty string")
	}
}