package ai

import (
	"context"
	"testing"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	responses map[string]string
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		responses: make(map[string]string),
	}
}

func (mp *MockProvider) Query(prompt string) (string, error) {
	if response, exists := mp.responses[prompt]; exists {
		return response, nil
	}
	return "Mock response for: " + prompt, nil
}

func (mp *MockProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return mp.Query(prompt)
}

func (mp *MockProvider) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	responseChan := make(chan StreamResponse, 1)
	go func() {
		defer close(responseChan)
		response, err := mp.Query(prompt)
		responseChan <- StreamResponse{
			Content: response,
			Done:    true,
			Error:   err,
		}
	}()
	return responseChan, nil
}

func (mp *MockProvider) GetPartialResponse() string {
	return ""
}

func (mp *MockProvider) SetResponse(prompt, response string) {
	mp.responses[prompt] = response
}

func TestExecutionAwareProviderIntegration(t *testing.T) {
	log := logger.NewLogger()
	
	// Create mock base provider
	mockProvider := NewMockProvider()
	mockProvider.SetResponse("test prompt", "mock response")
	
	// Test with execution disabled
	config := &ExecutionWrapperConfig{
		Enabled:       false,
		AutoExecute:   false,
		DryRunDefault: true,
		Patterns:      []string{},
	}
	
	wrapper := NewExecutionAwareProvider(mockProvider, config, log)
	
	response, err := wrapper.Query("test prompt")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if response != "mock response" {
		t.Errorf("Expected 'mock response', got '%s'", response)
	}
	
	// Test execution capabilities
	if wrapper.IsExecutionEnabled() {
		t.Error("Expected execution to be disabled")
	}
	
	if wrapper.IsAutoExecuteEnabled() {
		t.Error("Expected auto-execute to be disabled")
	}
}

func TestExecutionAwareProviderDetection(t *testing.T) {
	log := logger.NewLogger()
	
	// Create mock base provider
	mockProvider := NewMockProvider()
	
	// Test with execution enabled
	config := &ExecutionWrapperConfig{
		Enabled:       true,
		AutoExecute:   false,
		DryRunDefault: true,
		Patterns: []string{
			`(?i)\b(install|add)\s+\w+`,
			`(?i)\bplease (run|execute)`,
		},
	}
	
	wrapper := NewExecutionAwareProvider(mockProvider, config, log)
	
	// Test detection for execution requests
	tests := []struct {
		name           string
		prompt         string
		expectExecution bool
	}{
		{
			name:           "install command detected",
			prompt:         "install firefox",
			expectExecution: true,
		},
		{
			name:           "please run command detected",
			prompt:         "please run this command",
			expectExecution: true,
		},
		{
			name:           "normal query",
			prompt:         "how does nix work?",
			expectExecution: false,
		},
		{
			name:           "add package detected",
			prompt:         "add vim to my configuration",
			expectExecution: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execReq := wrapper.detectExecutionRequest(tt.prompt)
			hasExecReq := execReq != nil
			
			if hasExecReq != tt.expectExecution {
				t.Errorf("For prompt '%s': expected execution detection %v, got %v", 
					tt.prompt, tt.expectExecution, hasExecReq)
			}
		})
	}
}

func TestProviderManagerExecutionIntegration(t *testing.T) {
	log := logger.NewLogger()
	
	// Create test configuration with execution enabled
	cfg := &config.UserConfig{
		AIProvider: "mock",
		LogLevel:   "info",
		Execution: config.ExecutionConfig{
			Enabled:       true,
			DryRunDefault: true,
		},
		AIModels: config.AIModelsConfig{
			Providers: map[string]config.AIProviderConfig{
				"mock": {
					Name:            "Mock Provider",
					Type:            "test",
					Available:       true,
					RequiresAPIKey:  false,
					SupportsStreaming: true,
					Models: map[string]config.AIModelConfig{
						"test": {
							Name:        "Test Model",
							Type:        "chat",
							Default:     true,
						},
					},
				},
			},
			SelectionPreferences: config.AISelectionPreferences{
				DefaultProvider: "mock",
				DefaultModels: map[string]string{
					"mock": "test",
				},
			},
		},
		Cache: config.CacheConfig{
			Enabled: false, // Disable cache for tests
		},
	}
	
	pm := NewProviderManager(cfg, log)
	
	// Test execution management methods
	if !pm.IsExecutionEnabled() {
		t.Error("Expected execution to be enabled")
	}
	
	if pm.IsAutoExecuteEnabled() {
		t.Error("Expected auto-execute to be disabled by default")
	}
	
	// Test enabling auto-execution
	pm.EnableAutoExecution()
	if !pm.IsAutoExecuteEnabled() {
		t.Error("Expected auto-execute to be enabled after calling EnableAutoExecution")
	}
	
	// Test disabling auto-execution
	pm.DisableAutoExecution()
	if pm.IsAutoExecuteEnabled() {
		t.Error("Expected auto-execute to be disabled after calling DisableAutoExecution")
	}
	
	// Test disabling execution detection
	pm.SetExecutionEnabled(false)
	if pm.IsExecutionEnabled() {
		t.Error("Expected execution to be disabled after calling SetExecutionEnabled(false)")
	}
	
	// Test execution capabilities
	capabilities := pm.GetExecutionCapabilities()
	if capabilities["enabled"] != false {
		t.Error("Expected execution enabled to be false in capabilities")
	}
	
	if capabilities["auto_execute"] != false {
		t.Error("Expected auto_execute to be false in capabilities")
	}
}

func TestExecutionAwareProviderMethodDelegation(t *testing.T) {
	log := logger.NewLogger()
	
	mockProvider := NewMockProvider()
	mockProvider.SetResponse("test", "expected response")
	
	config := &ExecutionWrapperConfig{
		Enabled:       false, // Disable execution to test delegation
		AutoExecute:   false,
		DryRunDefault: true,
		Patterns:      []string{},
	}
	
	wrapper := NewExecutionAwareProvider(mockProvider, config, log)
	
	// Test Query method delegation
	response, err := wrapper.Query("test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response != "expected response" {
		t.Errorf("Expected 'expected response', got '%s'", response)
	}
	
	// Test GenerateResponse method delegation
	ctx := context.Background()
	response, err = wrapper.GenerateResponse(ctx, "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response != "expected response" {
		t.Errorf("Expected 'expected response', got '%s'", response)
	}
	
	// Test StreamResponse method delegation
	responseChan, err := wrapper.StreamResponse(ctx, "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	streamResponse := <-responseChan
	if streamResponse.Error != nil {
		t.Errorf("Unexpected error in stream: %v", streamResponse.Error)
	}
	if streamResponse.Content != "expected response" {
		t.Errorf("Expected 'expected response' in stream, got '%s'", streamResponse.Content)
	}
	if !streamResponse.Done {
		t.Error("Expected stream to be done")
	}
	
	// Test GetPartialResponse delegation
	partial := wrapper.GetPartialResponse()
	expected := mockProvider.GetPartialResponse()
	if partial != expected {
		t.Errorf("Expected '%s', got '%s'", expected, partial)
	}
}