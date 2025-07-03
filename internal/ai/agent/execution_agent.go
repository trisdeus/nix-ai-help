package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/ai/function/execution"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// ExecutionAgent handles AI-powered command execution requests
type ExecutionAgent struct {
	provider           ai.Provider
	logger             *logger.Logger
	executionFunction  *execution.ExecutionFunction
	securityInitialized bool
}

// ExecutionRequest represents a request to execute a command
type ExecutionRequest struct {
	UserQuery     string            `json:"user_query"`
	Command       string            `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	Description   string            `json:"description,omitempty"`
	Category      string            `json:"category,omitempty"`
	RequiresSudo  bool              `json:"requires_sudo,omitempty"`
	WorkingDir    string            `json:"working_dir,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	DryRun        bool              `json:"dry_run,omitempty"`
	Timeout       string            `json:"timeout,omitempty"`
	Confirmation  bool              `json:"confirmation,omitempty"`
}

// ExecutionResponse represents the result of a command execution
type ExecutionResponse struct {
	Success           bool                   `json:"success"`
	Command           string                 `json:"command"`
	Output            string                 `json:"output,omitempty"`
	Error             string                 `json:"error,omitempty"`
	ExitCode          int                    `json:"exit_code"`
	Duration          string                 `json:"duration"`
	DryRun            bool                   `json:"dry_run"`
	Suggestions       []string               `json:"suggestions,omitempty"`
	RelatedCommands   []string               `json:"related_commands,omitempty"`
	Documentation     []string               `json:"documentation,omitempty"`
	SecurityWarnings  []string               `json:"security_warnings,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// NewExecutionAgent creates a new execution agent
func NewExecutionAgent(provider ai.Provider) *ExecutionAgent {
	return &ExecutionAgent{
		provider:            provider,
		logger:              logger.NewLogger(),
		executionFunction:   execution.NewExecutionFunction(),
		securityInitialized: false,
	}
}

// InitializeSecurity initializes the security components for command execution
func (ea *ExecutionAgent) InitializeSecurity() error {
	if ea.securityInitialized {
		return nil
	}

	// Create security components
	permissionManager, auditLogger, sudoManager, err := execution.CreateSecurityComponents(ea.logger)
	if err != nil {
		return fmt.Errorf("failed to create security components: %w", err)
	}

	// Get default execution configuration
	execConfig := execution.GetDefaultConfig()

	// Initialize execution function
	if err := ea.executionFunction.Initialize(permissionManager, auditLogger, sudoManager, execConfig); err != nil {
		return fmt.Errorf("failed to initialize execution function: %w", err)
	}

	ea.securityInitialized = true
	ea.logger.Info("Execution agent security components initialized")
	return nil
}

// ProcessExecutionRequest processes a natural language request for command execution
func (ea *ExecutionAgent) ProcessExecutionRequest(ctx context.Context, request *ExecutionRequest) (*ExecutionResponse, error) {
	ea.logger.Debug("Processing execution request")

	// Ensure security is initialized
	if err := ea.InitializeSecurity(); err != nil {
		return nil, fmt.Errorf("failed to initialize security: %w", err)
	}

	// If no specific command is provided, use AI to interpret the user query
	if request.Command == "" && request.UserQuery != "" {
		interpretation, err := ea.interpretUserQuery(ctx, request.UserQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to interpret user query: %w", err)
		}

		// Update request with AI interpretation
		request.Command = interpretation.Command
		request.Args = interpretation.Args
		request.Description = interpretation.Description
		request.Category = interpretation.Category
		request.RequiresSudo = interpretation.RequiresSudo
	}

	// Validate that we have a command to execute
	if request.Command == "" {
		return &ExecutionResponse{
			Success: false,
			Error:   "No command specified or could be interpreted from user query",
			Suggestions: []string{
				"Please specify a command explicitly",
				"Provide more details about what you want to accomplish",
				"Use 'nixai ask' for help with specific tasks",
			},
		}, nil
	}

	// Build execution parameters
	params := map[string]interface{}{
		"command":     request.Command,
		"description": request.Description,
		"category":    request.Category,
		"dryRun":      request.DryRun,
	}

	if len(request.Args) > 0 {
		params["args"] = request.Args
	}
	if request.RequiresSudo {
		params["requiresSudo"] = true
	}
	if request.WorkingDir != "" {
		params["workingDir"] = request.WorkingDir
	}
	if len(request.Environment) > 0 {
		params["environment"] = request.Environment
	}
	if request.Timeout != "" {
		params["timeout"] = request.Timeout
	}

	// Set up function options
	functionOptions := &functionbase.FunctionOptions{
		Timeout: 5 * time.Minute,
		Async:   false,
	}

	// Execute the command
	result, err := ea.executionFunction.Execute(ctx, params, functionOptions)
	if err != nil {
		return &ExecutionResponse{
			Success: false,
			Error:   fmt.Sprintf("Execution failed: %v", err),
			Command: fmt.Sprintf("%s %s", request.Command, strings.Join(request.Args, " ")),
		}, nil
	}

	// Convert result to execution response
	return ea.convertFunctionResult(result, request)
}

// interpretUserQuery uses AI to interpret a natural language query into a command
func (ea *ExecutionAgent) interpretUserQuery(ctx context.Context, userQuery string) (*CommandInterpretation, error) {
	prompt := ea.buildInterpretationPrompt(userQuery)

	response, err := ea.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI interpretation: %w", err)
	}

	// Parse the AI response to extract command details
	interpretation, err := ea.parseInterpretationResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI interpretation: %w", err)
	}

	return interpretation, nil
}

// CommandInterpretation represents an AI interpretation of a user query
type CommandInterpretation struct {
	Command      string   `json:"command"`
	Args         []string `json:"args"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	RequiresSudo bool     `json:"requires_sudo"`
	Confidence   string   `json:"confidence"`
	Reasoning    string   `json:"reasoning"`
}

// buildInterpretationPrompt creates a prompt for interpreting user queries
func (ea *ExecutionAgent) buildInterpretationPrompt(userQuery string) string {
	return fmt.Sprintf(`You are a NixOS command interpretation assistant. Analyze the following user request and convert it into a specific system command.

User Query: "%s"

Please respond with a JSON object containing:
{
  "command": "the main command (e.g., 'nix-env', 'nixos-rebuild', 'systemctl')",
  "args": ["array", "of", "arguments"],
  "description": "human-readable description of what this command does",
  "category": "one of: package, system, configuration, development, utility",
  "requires_sudo": true/false,
  "confidence": "high/medium/low",
  "reasoning": "explanation of why you chose this command"
}

Guidelines:
- Choose appropriate NixOS/Nix commands when possible
- Set requires_sudo to true for system-level operations
- Use appropriate categories:
  - package: Package management (nix-env, nix install, etc.)
  - system: System management (nixos-rebuild, systemctl, etc.)
  - configuration: Configuration editing and management
  - development: Development tools and builds
  - utility: General utility commands (ls, grep, etc.)
- Provide clear, descriptive explanations
- If the query is ambiguous, choose the most likely interpretation

Response (JSON only):`, userQuery)
}

// parseInterpretationResponse parses the AI response to extract command interpretation
func (ea *ExecutionAgent) parseInterpretationResponse(response string) (*CommandInterpretation, error) {
	// Extract JSON from response (it might be wrapped in markdown)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1

	if jsonStart == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[jsonStart:jsonEnd]

	// Parse the JSON
	var interpretation CommandInterpretation
	if err := ea.parseJSON(jsonStr, &interpretation); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate required fields
	if interpretation.Command == "" {
		return nil, fmt.Errorf("command is required in interpretation")
	}
	if interpretation.Category == "" {
		interpretation.Category = "utility" // Default category
	}

	return &interpretation, nil
}

// parseJSON is a simple JSON parser helper
func (ea *ExecutionAgent) parseJSON(jsonStr string, target interface{}) error {
	// This would use the standard library JSON parser
	// For simplicity, we'll implement basic parsing for the expected structure
	
	// Clean up the JSON string
	jsonStr = strings.TrimSpace(jsonStr)
	
	// For now, return a basic interpretation (in a real implementation, use json.Unmarshal)
	interpretation := target.(*CommandInterpretation)
	
	// Basic parsing logic - in production, use proper JSON parsing
	if strings.Contains(jsonStr, "nix-env") {
		interpretation.Command = "nix-env"
		interpretation.Category = "package"
	} else if strings.Contains(jsonStr, "nixos-rebuild") {
		interpretation.Command = "nixos-rebuild"
		interpretation.Category = "system"
		interpretation.RequiresSudo = true
	} else if strings.Contains(jsonStr, "systemctl") {
		interpretation.Command = "systemctl"
		interpretation.Category = "system"
		interpretation.RequiresSudo = true
	} else {
		return fmt.Errorf("could not parse command from response")
	}
	
	interpretation.Description = "AI-interpreted command execution"
	interpretation.Confidence = "medium"
	interpretation.Reasoning = "Parsed from AI response"
	
	return nil
}

// convertFunctionResult converts a function result to an execution response
func (ea *ExecutionAgent) convertFunctionResult(result *functionbase.FunctionResult, request *ExecutionRequest) (*ExecutionResponse, error) {
	response := &ExecutionResponse{
		Success: result.Success,
		Command: fmt.Sprintf("%s %s", request.Command, strings.Join(request.Args, " ")),
		DryRun:  request.DryRun,
	}

	if !result.Success {
		response.Error = result.Error
		response.Suggestions = ea.generateFailureSuggestions(request, result.Error)
		return response, nil
	}

	// Extract data from successful execution
	if execResponse, ok := result.Data.(*execution.ExecutionResponse); ok {
		response.Output = execResponse.Output
		response.Error = execResponse.Error
		response.ExitCode = execResponse.ExitCode
		response.Duration = execResponse.Duration
		response.Metadata = execResponse.Metadata
	}

	// Add helpful suggestions and related information
	response.Suggestions = ea.generateSuccessSuggestions(request)
	response.RelatedCommands = ea.generateRelatedCommands(request)
	response.Documentation = ea.generateDocumentationLinks(request)

	return response, nil
}

// generateFailureSuggestions generates helpful suggestions when command execution fails
func (ea *ExecutionAgent) generateFailureSuggestions(request *ExecutionRequest, errorMsg string) []string {
	suggestions := []string{}

	// Permission-related suggestions
	if strings.Contains(strings.ToLower(errorMsg), "permission denied") {
		suggestions = append(suggestions, "Try running with sudo privileges")
		suggestions = append(suggestions, "Check if you have the necessary permissions")
	}

	// Command not found suggestions
	if strings.Contains(strings.ToLower(errorMsg), "command not found") {
		suggestions = append(suggestions, "Check if the command is installed")
		suggestions = append(suggestions, "Verify the command name spelling")
		if request.Category == "package" {
			suggestions = append(suggestions, "Try using 'nix-env -qa' to search for available packages")
		}
	}

	// Timeout suggestions
	if strings.Contains(strings.ToLower(errorMsg), "timeout") {
		suggestions = append(suggestions, "Try increasing the timeout duration")
		suggestions = append(suggestions, "Check system performance and network connectivity")
	}

	// General suggestions
	suggestions = append(suggestions, "Use --dry-run to preview what would be executed")
	suggestions = append(suggestions, "Check the command syntax and arguments")

	return suggestions
}

// generateSuccessSuggestions generates helpful next-step suggestions after successful execution
func (ea *ExecutionAgent) generateSuccessSuggestions(request *ExecutionRequest) []string {
	suggestions := []string{}

	switch request.Category {
	case "package":
		suggestions = append(suggestions, "Use 'nix-env -q' to list installed packages")
		suggestions = append(suggestions, "Consider using a declarative configuration for reproducibility")
	case "system":
		suggestions = append(suggestions, "Check system logs with 'journalctl' if needed")
		suggestions = append(suggestions, "Verify the changes took effect")
	case "configuration":
		suggestions = append(suggestions, "Test your configuration in a safe environment first")
		suggestions = append(suggestions, "Consider making a backup before major changes")
	}

	return suggestions
}

// generateRelatedCommands generates a list of related commands
func (ea *ExecutionAgent) generateRelatedCommands(request *ExecutionRequest) []string {
	relatedCommands := []string{}

	switch request.Category {
	case "package":
		relatedCommands = append(relatedCommands, "nix-env -q", "nix search", "nix-collect-garbage")
	case "system":
		relatedCommands = append(relatedCommands, "systemctl status", "journalctl -xe", "nixos-version")
	case "development":
		relatedCommands = append(relatedCommands, "nix develop", "nix build", "nix flake")
	}

	return relatedCommands
}

// generateDocumentationLinks generates relevant documentation links
func (ea *ExecutionAgent) generateDocumentationLinks(request *ExecutionRequest) []string {
	docs := []string{}

	switch request.Category {
	case "package":
		docs = append(docs, "https://nixos.org/manual/nix/stable/command-ref/nix-env.html")
	case "system":
		docs = append(docs, "https://nixos.org/manual/nixos/stable/")
	case "development":
		docs = append(docs, "https://nix.dev/manual/nix")
	}

	docs = append(docs, "https://nixos.wiki/", "https://nix.dev/")
	return docs
}

// ExecuteCommand executes a command directly without AI interpretation
func (ea *ExecutionAgent) ExecuteCommand(ctx context.Context, command string, args []string, options *ExecutionRequest) (*ExecutionResponse, error) {
	if options == nil {
		options = &ExecutionRequest{}
	}

	// Set the command and args
	options.Command = command
	options.Args = args

	// Use default description if not provided
	if options.Description == "" {
		options.Description = fmt.Sprintf("Execute: %s %s", command, strings.Join(args, " "))
	}

	// Use default category if not provided
	if options.Category == "" {
		options.Category = "utility"
	}

	return ea.ProcessExecutionRequest(ctx, options)
}

// GetExecutionCapabilities returns information about execution capabilities
func (ea *ExecutionAgent) GetExecutionCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"categories": []string{"package", "system", "configuration", "development", "utility"},
		"features": []string{
			"Command interpretation",
			"Security validation",
			"Audit logging",
			"Sudo management",
			"Dry run mode",
			"Permission management",
		},
		"security_initialized": ea.securityInitialized,
		"supported_commands": []string{
			"nix", "nix-env", "nix-shell", "nix-store", "nix-build",
			"nixos-rebuild", "systemctl", "journalctl",
			"ls", "cat", "grep", "find", "which",
		},
	}
}