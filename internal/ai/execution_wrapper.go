package ai

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai/types"
	"nix-ai-help/pkg/logger"
)

// ExecutionAwareProvider wraps an AI provider to automatically detect and execute commands
type ExecutionAwareProvider struct {
	baseProvider      Provider
	logger           *logger.Logger
	enabled          bool
	autoExecute      bool
	executionPatterns []*regexp.Regexp
}

// ExecutionWrapperConfig configures the execution-aware provider wrapper
type ExecutionWrapperConfig struct {
	Enabled       bool                     `json:"enabled"`        // Enable execution detection
	AutoExecute   bool                     `json:"auto_execute"`   // Automatically execute detected commands
	DryRunDefault bool                     `json:"dry_run_default"` // Use dry run by default
	Patterns      []string                 `json:"patterns"`       // Custom execution patterns
}

// NewExecutionAwareProvider creates a new execution-aware provider wrapper
func NewExecutionAwareProvider(baseProvider Provider, config *ExecutionWrapperConfig, log *logger.Logger) *ExecutionAwareProvider {
	if log == nil {
		log = logger.NewLogger()
	}

	if config == nil {
		config = getDefaultExecutionConfig()
	}

	// Compile execution detection patterns
	patterns := make([]*regexp.Regexp, 0, len(config.Patterns))
	for _, pattern := range config.Patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			patterns = append(patterns, regex)
		} else {
			log.Warn(fmt.Sprintf("Invalid execution pattern: %s", pattern))
		}
	}

	// Add default patterns if none provided
	if len(patterns) == 0 {
		patterns = getDefaultExecutionPatterns()
	}

	wrapper := &ExecutionAwareProvider{
		baseProvider:      baseProvider,
		logger:           log,
		enabled:          config.Enabled,
		autoExecute:      config.AutoExecute,
		executionPatterns: patterns,
	}

	return wrapper
}

// Query implements the Provider interface with execution awareness
func (eap *ExecutionAwareProvider) Query(prompt string) (string, error) {
	ctx := context.Background()
	return eap.GenerateResponse(ctx, prompt)
}

// GenerateResponse implements context-aware response generation with execution detection
func (eap *ExecutionAwareProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	// Check if execution detection is enabled
	if !eap.enabled {
		return eap.baseProvider.GenerateResponse(ctx, prompt)
	}

	// Detect if this is an execution request
	executionRequest := eap.detectExecutionRequest(prompt)
	if executionRequest == nil {
		// Not an execution request, use base provider
		return eap.baseProvider.GenerateResponse(ctx, prompt)
	}

	eap.logger.Info("Detected execution request in user query")

	// Handle execution request
	return eap.handleExecutionRequest(ctx, prompt, executionRequest)
}

// StreamResponse implements streaming with execution awareness
func (eap *ExecutionAwareProvider) StreamResponse(ctx context.Context, prompt string) (<-chan StreamResponse, error) {
	// For streaming, we need to handle execution requests differently
	// First check if it's an execution request
	if eap.enabled {
		executionRequest := eap.detectExecutionRequest(prompt)
		if executionRequest != nil {
			// Convert execution result to streaming response
			return eap.streamExecutionResponse(ctx, prompt, executionRequest)
		}
	}

	// Not an execution request, delegate to base provider
	return eap.baseProvider.StreamResponse(ctx, prompt)
}

// GetPartialResponse delegates to the base provider
func (eap *ExecutionAwareProvider) GetPartialResponse() string {
	return eap.baseProvider.GetPartialResponse()
}

// DetectExecutionRequest analyzes the prompt to detect execution requests (public)
func (eap *ExecutionAwareProvider) DetectExecutionRequest(prompt string) *types.ExecutionRequest {
	return eap.detectExecutionRequest(prompt)
}

// detectExecutionRequest analyzes the prompt to detect execution requests
func (eap *ExecutionAwareProvider) detectExecutionRequest(prompt string) *types.ExecutionRequest {
	promptLower := strings.ToLower(prompt)

	// Check against execution patterns
	for _, pattern := range eap.executionPatterns {
		if pattern.MatchString(promptLower) {
			return &types.ExecutionRequest{
				UserQuery:    prompt,
				Confirmation: !eap.autoExecute, // Require confirmation if not auto-executing
				DryRun:       !eap.autoExecute, // Use dry run if not auto-executing
			}
		}
	}

	return nil
}

// handleExecutionRequest processes an execution request
func (eap *ExecutionAwareProvider) handleExecutionRequest(ctx context.Context, originalPrompt string, execReq *types.ExecutionRequest) (string, error) {
	if eap.autoExecute {
		// For now, we suggest rather than execute automatically for safety
		// In a full implementation, this would integrate with the execution system
		return eap.generateExecutionSuggestion(ctx, originalPrompt, execReq)
	} else {
		// Provide execution suggestion without executing
		suggestion, err := eap.generateExecutionSuggestion(ctx, originalPrompt, execReq)
		if err != nil {
			// Fall back to base provider if suggestion generation fails
			return eap.baseProvider.GenerateResponse(ctx, originalPrompt)
		}
		return suggestion, nil
	}
}

// generateExecutionSuggestion creates a suggestion for command execution
func (eap *ExecutionAwareProvider) generateExecutionSuggestion(ctx context.Context, originalPrompt string, execReq *types.ExecutionRequest) (string, error) {
	// Generate execution suggestion using the base provider with enhanced prompt
	enhancedPrompt := fmt.Sprintf(`You detected an execution request: "%s"

Please provide:
1. The specific command that should be executed
2. A brief explanation of what it does
3. Any safety considerations
4. The nixai execute command syntax to run it safely

Format your response with clear sections and use the nixai execute command for safe execution.`, originalPrompt)
	
	response, err := eap.baseProvider.GenerateResponse(ctx, enhancedPrompt)
	if err != nil {
		return "", err
	}

	// Add execution detection notice
	suggestion := "🤖 **Execution Request Detected**\n\n" + response
	
	// Add safety reminder
	suggestion += "\n\n⚠️ **Safety Note:** Use `nixai execute` for safe command execution with built-in security validation."

	return suggestion, nil
}

// streamExecutionResponse converts execution results to streaming format
func (eap *ExecutionAwareProvider) streamExecutionResponse(ctx context.Context, prompt string, execReq *types.ExecutionRequest) (<-chan StreamResponse, error) {
	resultChan := make(chan StreamResponse, 5)

	go func() {
		defer close(resultChan)

		// Send initial response
		resultChan <- StreamResponse{
			Content: "🤖 Processing execution request...\n\n",
			Done:    false,
		}

		// Generate execution suggestion
		suggestion, err := eap.generateExecutionSuggestion(ctx, prompt, execReq)
		if err != nil {
			resultChan <- StreamResponse{
				Content: fmt.Sprintf("❌ Failed to generate execution suggestion: %v", err),
				Done:    true,
				Error:   err,
			}
			return
		}

		// Stream the suggestion
		resultChan <- StreamResponse{
			Content: suggestion,
			Done:    true,
		}
	}()

	return resultChan, nil
}

// formatExecutionResult formats an execution result for user display
func (eap *ExecutionAwareProvider) formatExecutionResult(result *types.ExecutionResponse) string {
	var response strings.Builder

	if result.Success {
		response.WriteString("✅ **Command Executed Successfully**\n\n")
		
		if result.Command != "" {
			response.WriteString(fmt.Sprintf("**Command:** `%s`\n", result.Command))
		}
		
		if result.Duration != "" {
			response.WriteString(fmt.Sprintf("**Duration:** %s\n", result.Duration))
		}
		
		if result.Output != "" {
			response.WriteString(fmt.Sprintf("\n**Output:**\n```\n%s\n```\n", result.Output))
		}
		
		if result.DryRun {
			response.WriteString("\n📝 **Note:** This was a dry run - no actual changes were made.\n")
		}
	} else {
		response.WriteString("❌ **Command Execution Failed**\n\n")
		
		if result.Error != "" {
			response.WriteString(fmt.Sprintf("**Error:** %s\n", result.Error))
		}
		
		if result.Command != "" {
			response.WriteString(fmt.Sprintf("**Attempted Command:** `%s`\n", result.Command))
		}
	}

	// Add suggestions
	if len(result.Suggestions) > 0 {
		response.WriteString("\n**Suggestions:**\n")
		for _, suggestion := range result.Suggestions {
			response.WriteString(fmt.Sprintf("- %s\n", suggestion))
		}
	}

	// Add related commands
	if len(result.RelatedCommands) > 0 {
		response.WriteString("\n**Related Commands:**\n")
		for _, cmd := range result.RelatedCommands {
			response.WriteString(fmt.Sprintf("- `%s`\n", cmd))
		}
	}

	// Add documentation
	if len(result.Documentation) > 0 {
		response.WriteString("\n**Documentation:**\n")
		for _, doc := range result.Documentation {
			response.WriteString(fmt.Sprintf("- %s\n", doc))
		}
	}

	return response.String()
}

// getDefaultExecutionConfig returns default execution wrapper configuration
func getDefaultExecutionConfig() *ExecutionWrapperConfig {
	return &ExecutionWrapperConfig{
		Enabled:       true,
		AutoExecute:   false, // Safe default - suggest rather than execute
		DryRunDefault: true,
		Patterns: []string{
			`(?i)\b(install|add|remove|delete|uninstall)\s+\w+`,
			`(?i)\b(rebuild|switch|build)\b`,
			`(?i)\b(run|execute|start|stop|restart)\s+\w+`,
			`(?i)\b(update|upgrade|download)\s+\w+`,
			`(?i)\b(enable|disable)\s+\w+`,
			`(?i)\b(check|status|list|show)\s+\w+`,
			`(?i)\bcan you (run|execute|install|build|start|stop)`,
			`(?i)\bplease (run|execute|install|build|start|stop)`,
			`(?i)\bhow do i (install|run|execute|build|start|stop)`,
		},
	}
}

// getDefaultExecutionPatterns returns compiled default execution patterns
func getDefaultExecutionPatterns() []*regexp.Regexp {
	patterns := []string{
		`(?i)\b(install|add|remove|delete|uninstall)\s+\w+`,
		`(?i)\b(rebuild|switch|build)\b`,
		`(?i)\b(run|execute|start|stop|restart)\s+\w+`,
		`(?i)\b(update|upgrade|download)\s+\w+`,
		`(?i)\b(enable|disable)\s+\w+`,
		`(?i)\b(check|status|list|show)\s+\w+`,
		`(?i)\bcan you (run|execute|install|build|start|stop)`,
		`(?i)\bplease (run|execute|install|build|start|stop)`,
		`(?i)\bhow do i (install|run|execute|build|start|stop)`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, regex)
		}
	}

	return compiled
}

// EnableAutoExecution enables automatic command execution (use with caution)
func (eap *ExecutionAwareProvider) EnableAutoExecution() {
	eap.autoExecute = true
	eap.logger.Warn("Auto-execution enabled - commands will be executed automatically")
}

// DisableAutoExecution disables automatic command execution (safer mode)
func (eap *ExecutionAwareProvider) DisableAutoExecution() {
	eap.autoExecute = false
	eap.logger.Info("Auto-execution disabled - commands will be suggested only")
}

// IsExecutionEnabled returns whether execution detection is enabled
func (eap *ExecutionAwareProvider) IsExecutionEnabled() bool {
	return eap.enabled
}

// IsAutoExecuteEnabled returns whether auto-execution is enabled
func (eap *ExecutionAwareProvider) IsAutoExecuteEnabled() bool {
	return eap.autoExecute
}

// GetExecutionCapabilities returns information about execution capabilities
func (eap *ExecutionAwareProvider) GetExecutionCapabilities() map[string]interface{} {
	if !eap.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	capabilities := map[string]interface{}{
		"enabled":       true,
		"auto_execute":  eap.autoExecute,
		"pattern_count": len(eap.executionPatterns),
		"detection_patterns": len(eap.executionPatterns),
		"suggestion_mode": !eap.autoExecute,
	}
	
	return capabilities
}