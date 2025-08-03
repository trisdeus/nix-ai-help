package advanced

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// PluginAIAgent represents an AI agent implemented as a plugin
type PluginAIAgent struct {
	provider     ai.Provider
	logger       *logger.Logger
	name         string
	description  string
	capabilities []string
	role         string
}

// NewPluginAIAgent creates a new plugin-based AI agent
func NewPluginAIAgent(provider ai.Provider, log *logger.Logger, name, description, role string) *PluginAIAgent {
	return &PluginAIAgent{
		provider:     provider,
		logger:       log,
		name:         name,
		description:  description,
		capabilities: []string{"ai-agent", "plugin-based"},
		role:         role,
	}
}

// Name returns the name of the agent
func (paa *PluginAIAgent) Name() string {
	return paa.name
}

// Description returns the description of the agent
func (paa *PluginAIAgent) Description() string {
	return paa.description
}

// Role returns the role of the agent
func (paa *PluginAIAgent) Role() string {
	return paa.role
}

// Capabilities returns the capabilities of the agent
func (paa *PluginAIAgent) Capabilities() []string {
	return paa.capabilities
}

// GenerateResponse generates a response using the plugin-based AI agent
func (paa *PluginAIAgent) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	paa.logger.Info(fmt.Sprintf("Plugin AI Agent '%s' generating response", paa.name))
	
	// Create agent-specific prompt
	agentPrompt := fmt.Sprintf(`You are a %s AI agent. Your role is to assist with %s-related tasks.

%s

Generate a comprehensive, accurate, and helpful response to this query:
"%s"

Follow these guidelines:
1. Use your specialized knowledge in %s
2. Provide specific, actionable advice
3. Include relevant examples when appropriate
4. Reference official documentation
5. Warn about potential risks or pitfalls
6. Suggest verification steps

Return your response as plain text.`, 
		paa.role, 
		strings.ToLower(paa.role),
		paa.description,
		prompt,
		strings.ToLower(paa.role))

	response, err := paa.provider.GenerateResponse(ctx, agentPrompt)
	if err != nil {
		return "", fmt.Errorf("plugin AI agent failed to generate response: %w", err)
	}

	return response, nil
}

// SetRole sets the role for the agent
func (paa *PluginAIAgent) SetRole(role string) error {
	paa.role = role
	return nil
}

// SetContext sets the context for the agent
func (paa *PluginAIAgent) SetContext(context interface{}) error {
	// For plugin agents, context is handled through the prompt
	return nil
}

// GetPartialResponse returns a partial response if available
func (paa *PluginAIAgent) GetPartialResponse() string {
	return ""
}

// StreamResponse streams a response using the plugin-based AI agent
func (paa *PluginAIAgent) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	paa.logger.Info(fmt.Sprintf("Plugin AI Agent '%s' streaming response", paa.name))
	
	// Create agent-specific prompt
	agentPrompt := fmt.Sprintf(`You are a %s AI agent. Your role is to assist with %s-related tasks.

%s

Generate a comprehensive, accurate, and helpful response to this query:
"%s"

Follow these guidelines:
1. Use your specialized knowledge in %s
2. Provide specific, actionable advice
3. Include relevant examples when appropriate
4. Reference official documentation
5. Warn about potential risks or pitfalls
6. Suggest verification steps

Return your response as a stream of plain text chunks.`, 
		paa.role, 
		strings.ToLower(paa.role),
		paa.description,
		prompt,
		strings.ToLower(paa.role))

	return paa.provider.StreamResponse(ctx, agentPrompt)
}

// ValidateResponse validates a response from the plugin-based AI agent
func (paa *PluginAIAgent) ValidateResponse(ctx context.Context, response, prompt string) (bool, error) {
	paa.logger.Info(fmt.Sprintf("Plugin AI Agent '%s' validating response", paa.name))
	
	// Create validation prompt
	validationPrompt := fmt.Sprintf(`You are validating the response from an AI agent.

Original query: "%s"
AI response: "%s"

Evaluate this response and provide a simple yes/no answer:
Is this response technically accurate, complete, and helpful for the user's query?

Respond with only "yes" or "no".`, prompt, response)

	validatedResponse, err := paa.provider.GenerateResponse(ctx, validationPrompt)
	if err != nil {
		return false, fmt.Errorf("plugin AI agent failed to validate response: %w", err)
	}

	validatedResponse = strings.ToLower(strings.TrimSpace(validatedResponse))
	return validatedResponse == "yes", nil
}

// ImproveResponse improves a response from the plugin-based AI agent
func (paa *PluginAIAgent) ImproveResponse(ctx context.Context, response, prompt string) (string, error) {
	paa.logger.Info(fmt.Sprintf("Plugin AI Agent '%s' improving response", paa.name))
	
	// Create improvement prompt
	improvementPrompt := fmt.Sprintf(`You are improving a response from an AI agent.

Original query: "%s"
Current AI response: "%s"

Identify any issues with this response and provide an improved version that:
1. Fixes technical inaccuracies
2. Adds missing information
3. Clarifies confusing sections
4. Improves structure and flow
5. Ensures completeness

Return only the improved response as plain text.`, prompt, response)

	improvedResponse, err := paa.provider.GenerateResponse(ctx, improvementPrompt)
	if err != nil {
		return "", fmt.Errorf("plugin AI agent failed to improve response: %w", err)
	}

	return improvedResponse, nil
}

// GetMetrics returns metrics for the plugin-based AI agent
func (paa *PluginAIAgent) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"agent_name":    paa.name,
		"agent_role":    paa.role,
		"queries_handled": 0,
		"avg_response_time": "0s",
		"success_rate":  1.0,
		"last_query":    time.Now().Format(time.RFC3339),
	}
}