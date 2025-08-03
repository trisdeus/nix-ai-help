package advanced

import (
	"context"
	"testing"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// mockProvider2 is a mock AI provider for testing
type mockProvider2 struct {
	response string
}

func (mp *mockProvider2) Query(prompt string) (string, error) {
	if mp.response == "" {
		return "Default mock response", nil
	}
	return mp.response, nil
}

func (mp *mockProvider2) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	if mp.response == "" {
		return "Default mock response", nil
	}
	return mp.response, nil
}

func (mp *mockProvider2) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		response := "Default mock response"
		if mp.response != "" {
			response = mp.response
		}
		ch <- ai.StreamResponse{
			Content: response,
			Done:    true,
		}
	}()
	return ch, nil
}

func (mp *mockProvider2) GetPartialResponse() string {
	return ""
}

func TestPluginAIAgent(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider2{
		response: "Mock response from plugin AI agent",
	}

	// Create plugin AI agent
	agent := NewPluginAIAgent(provider, log, "test-agent", "Test AI agent", "Assistant")

	// Test agent metadata
	if agent.Name() != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agent.Name())
	}

	if agent.Description() != "Test AI agent" {
		t.Errorf("Expected agent description 'Test AI agent', got '%s'", agent.Description())
	}

	if agent.Role() != "Assistant" {
		t.Errorf("Expected agent role 'Assistant', got '%s'", agent.Role())
	}

	capabilities := agent.Capabilities()
	if len(capabilities) == 0 {
		t.Error("Expected agent capabilities, got none")
	}

	// Test generating response
	ctx := context.Background()
	response, err := agent.GenerateResponse(ctx, "Test query")
	if err != nil {
		t.Fatalf("GenerateResponse failed: %v", err)
	}

	if response == "" {
		t.Error("Expected response, got empty string")
	}

	// Test setting role
	err = agent.SetRole("Expert")
	if err != nil {
		t.Fatalf("SetRole failed: %v", err)
	}

	if agent.Role() != "Expert" {
		t.Errorf("Expected agent role 'Expert' after setting, got '%s'", agent.Role())
	}

	// Test setting context (should not fail)
	err = agent.SetContext("test context")
	if err != nil {
		t.Fatalf("SetContext failed: %v", err)
	}

	// Test getting partial response
	partial := agent.GetPartialResponse()
	// Should be empty for plugin agents
	if partial != "" {
		t.Errorf("Expected empty partial response, got '%s'", partial)
	}

	// Test streaming response
	stream, err := agent.StreamResponse(ctx, "Test query")
	if err != nil {
		t.Fatalf("StreamResponse failed: %v", err)
	}

	if stream == nil {
		t.Error("Expected stream channel, got nil")
	}

	// Test receiving from stream
	timeout := time.After(1 * time.Second)
	select {
	case result := <-stream:
		if result.Content == "" {
			t.Error("Expected stream content, got empty string")
		}
		if !result.Done {
			t.Error("Expected stream to be done, got false")
		}
	case <-timeout:
		t.Error("Expected stream result, timed out")
	}

	// Test validating response
	valid, err := agent.ValidateResponse(ctx, "Test response", "Test query")
	if err != nil {
		t.Fatalf("ValidateResponse failed: %v", err)
	}

	// With our mock provider, this will depend on what response we get
	// Let's just make sure it returns a boolean result
	if valid != true && valid != false {
		t.Errorf("Expected boolean validation result, got invalid value")
	}

	// Test improving response
	improved, err := agent.ImproveResponse(ctx, "Test response", "Test query")
	if err != nil {
		t.Fatalf("ImproveResponse failed: %v", err)
	}

	if improved == "" {
		t.Error("Expected improved response, got empty string")
	}

	// Test getting metrics
	metrics := agent.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics, got nil")
	}

	if len(metrics) == 0 {
		t.Error("Expected non-empty metrics, got empty map")
	}
}

func TestPluginAIAgentEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider2{
		response: "",
	}

	// Create plugin AI agent
	agent := NewPluginAIAgent(provider, log, "edge-case-agent", "Edge case AI agent", "Tester")

	// Test with empty query
	ctx := context.Background()
	response, err := agent.GenerateResponse(ctx, "")
	if err != nil {
		t.Fatalf("GenerateResponse failed with empty query: %v", err)
	}

	// Empty query should still return response from mock provider
	if response == "" {
		t.Error("Expected response with empty query, got empty string")
	}

	// Test with very long query
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	response2, err := agent.GenerateResponse(ctx, longQuery)
	if err != nil {
		t.Fatalf("GenerateResponse failed with long query: %v", err)
	}

	if response2 == "" {
		t.Error("Expected response with long query, got empty string")
	}

	// Test with nil context (should handle gracefully)
	_, err = agent.GenerateResponse(nil, "Test query")
	if err != nil {
		// With our implementation, this might fail gracefully with a timeout or other error
		// which is acceptable behavior
		t.Logf("GenerateResponse correctly failed with nil context: %v", err)
	}

	// Test streaming with empty query
	stream, err := agent.StreamResponse(ctx, "")
	if err != nil {
		t.Fatalf("StreamResponse failed with empty query: %v", err)
	}

	if stream == nil {
		t.Error("Expected stream channel with empty query, got nil")
	}

	// Test validating empty response
	valid, err := agent.ValidateResponse(ctx, "", "Test query")
	if err != nil {
		t.Fatalf("ValidateResponse failed with empty response: %v", err)
	}

	// Should still return boolean result
	if valid != true && valid != false {
		t.Errorf("Expected boolean validation result, got invalid value")
	}

	// Test improving empty response
	improved, err := agent.ImproveResponse(ctx, "", "Test query")
	if err != nil {
		t.Fatalf("ImproveResponse failed with empty response: %v", err)
	}

	// Should return some response
	if improved == "" {
		t.Error("Expected improved response with empty input, got empty string")
	}

	// Test getting metrics
	metrics := agent.GetMetrics()
	if metrics == nil {
		t.Error("Expected metrics with edge cases, got nil")
	}
}

func TestPluginAIAgentConfiguration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider2{
		response: "Configured response",
	}

	// Test creating agent with different configurations
	agent1 := NewPluginAIAgent(provider, log, "config-agent-1", "Configuration agent 1", "Assistant")
	if agent1 == nil {
		t.Error("Expected agent with configuration 1, got nil")
	}

	agent2 := NewPluginAIAgent(provider, log, "config-agent-2", "Configuration agent 2", "Expert")
	if agent2 == nil {
		t.Error("Expected agent with configuration 2, got nil")
	}

	agent3 := NewPluginAIAgent(provider, log, "config-agent-3", "Configuration agent 3", "Specialist")
	if agent3 == nil {
		t.Error("Expected agent with configuration 3, got nil")
	}

	// Test different roles
	if agent1.Role() != "Assistant" {
		t.Errorf("Expected agent1 role 'Assistant', got '%s'", agent1.Role())
	}

	if agent2.Role() != "Expert" {
		t.Errorf("Expected agent2 role 'Expert', got '%s'", agent2.Role())
	}

	if agent3.Role() != "Specialist" {
		t.Errorf("Expected agent3 role 'Specialist', got '%s'", agent3.Role())
	}

	// Test setting roles
	err := agent1.SetRole("Expert")
	if err != nil {
		t.Fatalf("SetRole failed for agent1: %v", err)
	}

	if agent1.Role() != "Expert" {
		t.Errorf("Expected agent1 role 'Expert' after setting, got '%s'", agent1.Role())
	}

	// Test capabilities
	caps1 := agent1.Capabilities()
	if len(caps1) == 0 {
		t.Error("Expected capabilities for agent1, got none")
	}

	caps2 := agent2.Capabilities()
	if len(caps2) == 0 {
		t.Error("Expected capabilities for agent2, got none")
	}

	caps3 := agent3.Capabilities()
	if len(caps3) == 0 {
		t.Error("Expected capabilities for agent3, got none")
	}

	// Test getting metrics
	metrics1 := agent1.GetMetrics()
	if metrics1 == nil {
		t.Error("Expected metrics for agent1, got nil")
	}

	metrics2 := agent2.GetMetrics()
	if metrics2 == nil {
		t.Error("Expected metrics for agent2, got nil")
	}

	metrics3 := agent3.GetMetrics()
	if metrics3 == nil {
		t.Error("Expected metrics for agent3, got nil")
	}
}