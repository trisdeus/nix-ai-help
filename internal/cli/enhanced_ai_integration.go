package cli

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/advanced"
	"nix-ai-help/pkg/logger"
)

// EnhancedAIIntegration provides integration with advanced AI features
type EnhancedAIIntegration struct {
	provider    ai.Provider
	logger      *logger.Logger
	coordinator  *advanced.AdvancedAICoordinator
	enableAdvanced bool
}

// NewEnhancedAIIntegration creates a new enhanced AI integration
func NewEnhancedAIIntegration(provider ai.Provider, log *logger.Logger, enableAdvanced bool) *EnhancedAIIntegration {
	// Create advanced AI coordinator with all features enabled
	config := advanced.AdvancedAICoordinatorConfig{
		EnableReasoning:  true,
		EnableCorrection: true,
		EnablePlanning:   true,
		EnableScoring:    true,
	}
	
	coordinator := advanced.NewAdvancedAICoordinator(provider, log, config)
	
	return &EnhancedAIIntegration{
		provider:      provider,
		logger:        log,
		coordinator:   coordinator,
		enableAdvanced: enableAdvanced,
	}
}

// ProcessQueryWithAdvancedAI processes a query with enhanced AI features
func (eai *EnhancedAIIntegration) ProcessQueryWithAdvancedAI(ctx context.Context, query string) (string, error) {
	if !eai.enableAdvanced {
		// Use regular AI processing
		return eai.processQueryWithRegularAI(ctx, query)
	}
	
	// Use enhanced AI processing
	response, err := eai.coordinator.ProcessQuery(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to process query with advanced AI: %w", err)
	}
	
	// Format the response
	formatted := eai.coordinator.FormatResponse(response)
	return formatted, nil
}

// processQueryWithRegularAI processes a query with regular AI (fallback)
func (eai *EnhancedAIIntegration) processQueryWithRegularAI(ctx context.Context, query string) (string, error) {
	// Build a standard prompt
	prompt := fmt.Sprintf(`You are a NixOS expert helping users with configuration questions.

User question: %s

Please provide a concise, accurate response focusing on NixOS-specific solutions.
Never suggest nix-env commands - always use nix profile or flakes instead.
Always verify package names and configuration options with official documentation.
Include code examples when appropriate.
End with a reminder to rebuild with 'sudo nixos-rebuild switch'.`, query)
	
	// Query the AI provider
	if p, ok := eai.provider.(interface {
		GenerateResponse(context.Context, string) (string, error)
	}); ok {
		response, err := p.GenerateResponse(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to generate response: %w", err)
		}
		return response, nil
	}
	
	if p, ok := eai.provider.(interface{ Query(string) (string, error) }); ok {
		response, err := p.Query(prompt)
		if err != nil {
			return "", fmt.Errorf("failed to query provider: %w", err)
		}
		return response, nil
	}
	
	return "", fmt.Errorf("provider does not implement GenerateResponse or Query")
}

// FormatEnhancedAIResponse formats an enhanced AI response for display
func (eai *EnhancedAIIntegration) FormatEnhancedAIResponse(response *advanced.AdvancedAIResponse) string {
	if response == nil {
		return "No response to format"
	}
	
	return eai.coordinator.FormatResponse(response)
}

// GetEnhancedAIStatus returns the status of the enhanced AI system
func (eai *EnhancedAIIntegration) GetEnhancedAIStatus() string {
	if !eai.enableAdvanced {
		return "Enhanced AI features: Disabled"
	}
	
	var features []string
	
	if eai.coordinator.EnableReasoning() {
		features = append(features, "Chain-of-Thought Reasoning")
	}
	
	if eai.coordinator.EnableCorrection() {
		features = append(features, "Self-Correction")
	}
	
	if eai.coordinator.EnablePlanning() {
		features = append(features, "Task Planning")
	}
	
	if eai.coordinator.EnableScoring() {
		features = append(features, "Confidence Scoring")
	}
	
	if len(features) == 0 {
		return "Enhanced AI features: None enabled"
	}
	
	return fmt.Sprintf("Enhanced AI features: %s", strings.Join(features, ", "))
}