package errors

import (
	"fmt"
	"strings"
	"time"
)

// UserFriendlyMessageGenerator generates user-friendly error messages
type UserFriendlyMessageGenerator struct {
	templates map[ErrorCode]MessageTemplate
}

// MessageTemplate defines how to generate user-friendly messages for specific error codes
type MessageTemplate struct {
	UserMessage     string            `json:"user_message"`
	TechnicalDetail string            `json:"technical_detail"`
	Suggestions     []string          `json:"suggestions"`
	DocsLinks       []string          `json:"docs_links"`
	Severity        ErrorSeverity     `json:"severity"`
	Category        ErrorCategory     `json:"category"`
	Variables       map[string]string `json:"variables,omitempty"`
}

// NewUserFriendlyMessageGenerator creates a new message generator with default templates
func NewUserFriendlyMessageGenerator() *UserFriendlyMessageGenerator {
	generator := &UserFriendlyMessageGenerator{
		templates: make(map[ErrorCode]MessageTemplate),
	}
	generator.loadDefaultTemplates()
	return generator
}

// loadDefaultTemplates loads default message templates for common error codes
func (ufmg *UserFriendlyMessageGenerator) loadDefaultTemplates() {
	// Network errors
	ufmg.templates[ErrorCodeNetworkTimeout] = MessageTemplate{
		UserMessage:     "Network connection timed out. Please check your internet connection.",
		TechnicalDetail: "A network request exceeded the timeout threshold.",
		Suggestions: []string{
			"Check your internet connection",
			"Try again in a few moments",
			"Check if there are any network restrictions or firewalls blocking the connection",
		},
		DocsLinks: []string{
			"https://wiki.nixos.org/wiki/Networking",
		},
		Severity: SeverityMedium,
		Category: CategoryNetwork,
	}

	ufmg.templates[ErrorCodeConnectionRefused] = MessageTemplate{
		UserMessage:     "Connection was refused by the server. The service may be unavailable.",
		TechnicalDetail: "The target server actively refused the connection request.",
		Suggestions: []string{
			"Check if the service is running",
			"Verify the server address and port",
			"Try again later as the service may be temporarily unavailable",
		},
		Severity: SeverityMedium,
		Category: CategoryNetwork,
	}

	// AI provider errors
	ufmg.templates[ErrorCodeAIProviderUnavailable] = MessageTemplate{
		UserMessage:     "AI service is currently unavailable. nixai will use local processing when possible.",
		TechnicalDetail: "The configured AI provider could not be reached or returned an error.",
		Suggestions: []string{
			"Check your AI provider configuration in ~/.config/nixai/config.yaml",
			"Verify your API keys are valid and not expired",
			"Try switching to a different AI provider (e.g., Ollama for local processing)",
			"Check the AI provider's status page for known issues",
		},
		DocsLinks: []string{
			"https://docs.nixai.com/configuration/ai-providers",
		},
		Severity: SeverityHigh,
		Category: CategoryAI,
	}

	ufmg.templates[ErrorCodeAIRateLimited] = MessageTemplate{
		UserMessage:     "AI service rate limit exceeded. Please wait before making more requests.",
		TechnicalDetail: "The AI provider has throttled requests due to rate limiting.",
		Suggestions: []string{
			"Wait a few minutes before trying again",
			"Consider upgrading your AI provider plan for higher rate limits",
			"Switch to Ollama for unlimited local AI processing",
		},
		Severity: SeverityMedium,
		Category: CategoryAI,
	}

	ufmg.templates[ErrorCodeAIQuotaExceeded] = MessageTemplate{
		UserMessage:     "AI service quota exceeded. You've reached the usage limit for your plan.",
		TechnicalDetail: "The AI provider indicates that the usage quota has been exceeded.",
		Suggestions: []string{
			"Check your AI provider dashboard for usage details",
			"Upgrade your plan for higher quotas",
			"Switch to Ollama for unlimited local AI processing",
			"Wait until your quota resets (usually monthly)",
		},
		Severity: SeverityHigh,
		Category: CategoryAI,
	}

	// File system errors
	ufmg.templates[ErrorCodeFileNotFound] = MessageTemplate{
		UserMessage:     "Required file not found. Please check the file path.",
		TechnicalDetail: "A required file or directory could not be located.",
		Suggestions: []string{
			"Verify the file path is correct",
			"Check if the file exists and is accessible",
			"Ensure you have the necessary permissions to access the file",
		},
		Severity: SeverityMedium,
		Category: CategoryFileSystem,
	}

	ufmg.templates[ErrorCodePermissionDenied] = MessageTemplate{
		UserMessage:     "Permission denied. You don't have the necessary permissions for this operation.",
		TechnicalDetail: "The operation failed due to insufficient file system permissions.",
		Suggestions: []string{
			"Check file and directory permissions",
			"Run the command with appropriate privileges if necessary",
			"Ensure you're the owner of the files or directories involved",
		},
		Severity: SeverityMedium,
		Category: CategoryFileSystem,
	}

	ufmg.templates[ErrorCodeConfigInvalid] = MessageTemplate{
		UserMessage:     "Configuration file contains invalid settings. Please check your configuration.",
		TechnicalDetail: "The configuration file has syntax errors or invalid values.",
		Suggestions: []string{
			"Check ~/.config/nixai/config.yaml for syntax errors",
			"Validate YAML syntax using an online validator",
			"Reset to default configuration if needed: nixai config --reset",
			"Review the configuration documentation",
		},
		DocsLinks: []string{
			"https://docs.nixai.com/configuration",
		},
		Severity: SeverityHigh,
		Category: CategoryFileSystem,
	}

	// NixOS errors
	ufmg.templates[ErrorCodeNixBuildFailed] = MessageTemplate{
		UserMessage:     "Nix build failed. There may be dependency issues or configuration problems.",
		TechnicalDetail: "The Nix build process encountered an error and could not complete.",
		Suggestions: []string{
			"Check the build logs for specific error details",
			"Run 'nix-collect-garbage' to clean up corrupted store paths",
			"Update your channels: nix-channel --update",
			"Try building with '--show-trace' for more detailed error information",
			"Check if your configuration.nix has any syntax errors",
		},
		DocsLinks: []string{
			"https://wiki.nixos.org/wiki/FAQ",
			"https://wiki.nixos.org/wiki/Troubleshooting",
		},
		Severity: SeverityHigh,
		Category: CategoryNixOS,
	}

	ufmg.templates[ErrorCodeNixConfigInvalid] = MessageTemplate{
		UserMessage:     "NixOS configuration is invalid. Please check your configuration.nix file.",
		TechnicalDetail: "The NixOS configuration file contains syntax errors or invalid options.",
		Suggestions: []string{
			"Check /etc/nixos/configuration.nix for syntax errors",
			"Run 'nixos-rebuild dry-run' to validate configuration without applying",
			"Use 'nix-instantiate --parse' to check syntax",
			"Review the NixOS manual for correct option syntax",
		},
		DocsLinks: []string{
			"https://nixos.org/manual/nixos/stable/",
		},
		Severity: SeverityHigh,
		Category: CategoryNixOS,
	}

	ufmg.templates[ErrorCodeNixStoreCorrupted] = MessageTemplate{
		UserMessage:     "Nix store corruption detected. Your system may need repair.",
		TechnicalDetail: "The Nix store has corrupted files or missing dependencies.",
		Suggestions: []string{
			"Run 'nix-store --verify --check-contents' to identify corrupted files",
			"Use 'nix-store --repair-path' to repair specific paths",
			"Consider running 'nix-collect-garbage -d' to clean up",
			"If problems persist, you may need to reinstall affected packages",
		},
		DocsLinks: []string{
			"https://wiki.nixos.org/wiki/Nix_store",
		},
		Severity: SeverityCritical,
		Category: CategoryNixOS,
	}

	// MCP server errors
	ufmg.templates[ErrorCodeMCPServerUnavailable] = MessageTemplate{
		UserMessage:     "MCP server is not available. Documentation queries may be limited.",
		TechnicalDetail: "The Model Context Protocol server could not be reached or started.",
		Suggestions: []string{
			"Start the MCP server: nixai mcp-server start",
			"Check MCP server status: nixai mcp-server status",
			"Restart the MCP server: nixai mcp-server restart",
			"Check if port 3001 is available and not blocked by firewall",
		},
		Severity: SeverityMedium,
		Category: CategoryMCP,
	}

	ufmg.templates[ErrorCodeMCPSocketError] = MessageTemplate{
		UserMessage:     "MCP socket connection error. Falling back to HTTP connection.",
		TechnicalDetail: "Unix socket connection to MCP server failed.",
		Suggestions: []string{
			"Check if socket file exists and has correct permissions",
			"MCP server will automatically fallback to HTTP",
			"Try restarting the MCP server if problems persist",
		},
		Severity: SeverityLow,
		Category: CategoryMCP,
	}

	// Cache errors
	ufmg.templates[ErrorCodeCacheCorrupted] = MessageTemplate{
		UserMessage:     "Cache data is corrupted. Cache will be rebuilt automatically.",
		TechnicalDetail: "The cache files are corrupted or in an inconsistent state.",
		Suggestions: []string{
			"Cache will be automatically rebuilt",
			"You can manually clear cache: nixai performance cache clear",
			"No action required - nixai will handle this automatically",
		},
		Severity: SeverityLow,
		Category: CategoryCache,
	}

	ufmg.templates[ErrorCodeStorageFull] = MessageTemplate{
		UserMessage:     "Storage space is full. Please free up disk space.",
		TechnicalDetail: "The file system has insufficient space for the operation.",
		Suggestions: []string{
			"Free up disk space by removing unnecessary files",
			"Run 'nix-collect-garbage -d' to clean up old Nix generations",
			"Check disk usage: df -h",
			"Consider moving large files to external storage",
		},
		Severity: SeverityHigh,
		Category: CategoryFileSystem,
	}

	// Validation errors
	ufmg.templates[ErrorCodeInvalidInput] = MessageTemplate{
		UserMessage:     "Invalid input provided. Please check your command arguments.",
		TechnicalDetail: "The provided input does not meet the expected format or constraints.",
		Suggestions: []string{
			"Check the command help: nixai <command> --help",
			"Verify all required arguments are provided",
			"Ensure input format matches the expected pattern",
		},
		Severity: SeverityLow,
		Category: CategoryValidation,
	}

	ufmg.templates[ErrorCodeMissingParameter] = MessageTemplate{
		UserMessage:     "Required parameter is missing. Please provide all necessary arguments.",
		TechnicalDetail: "A required parameter was not provided for the operation.",
		Suggestions: []string{
			"Check the command help: nixai <command> --help",
			"Ensure all required parameters are provided",
			"Review the command examples in the documentation",
		},
		Severity: SeverityLow,
		Category: CategoryValidation,
	}

	// Internal errors
	ufmg.templates[ErrorCodeInternalError] = MessageTemplate{
		UserMessage:     "An internal error occurred. This is likely a bug that should be reported.",
		TechnicalDetail: "An unexpected internal error occurred in nixai.",
		Suggestions: []string{
			"Try the operation again",
			"Report this issue to the nixai developers",
			"Include the error details and steps to reproduce",
			"Check if there's a known issue or workaround",
		},
		DocsLinks: []string{
			"https://github.com/nixai-org/nixai/issues",
		},
		Severity: SeverityCritical,
		Category: CategoryInternal,
	}

	ufmg.templates[ErrorCodePanicRecovered] = MessageTemplate{
		UserMessage:     "A critical error was recovered. The operation was safely aborted.",
		TechnicalDetail: "A panic was recovered to prevent application crash.",
		Suggestions: []string{
			"Try the operation again",
			"Report this issue with the provided stack trace",
			"This is likely a bug that needs to be fixed",
		},
		Severity: SeverityCritical,
		Category: CategoryInternal,
	}
}

// GenerateUserFriendlyMessage generates a user-friendly message for an error
func (ufmg *UserFriendlyMessageGenerator) GenerateUserFriendlyMessage(err error) string {
	if nixaiErr, ok := err.(*NixAIError); ok {
		return ufmg.GenerateForNixAIError(nixaiErr)
	}

	// For non-NixAIError, try to match by message content
	return ufmg.GenerateForGenericError(err)
}

// GenerateForNixAIError generates a user-friendly message for a NixAIError
func (ufmg *UserFriendlyMessageGenerator) GenerateForNixAIError(err *NixAIError) string {
	if template, exists := ufmg.templates[err.Code]; exists {
		message := template.UserMessage

		// If the error already has a user message, prefer that
		if err.UserMessage != "" {
			message = err.UserMessage
		}

		// Add suggestions if available
		suggestions := err.Suggestions
		if len(suggestions) == 0 && len(template.Suggestions) > 0 {
			suggestions = template.Suggestions
		}

		if len(suggestions) > 0 {
			message += "\n\nSuggested actions:"
			for i, suggestion := range suggestions {
				message += fmt.Sprintf("\n  %d. %s", i+1, suggestion)
			}
		}

		// Add technical details if in debug mode or for critical errors
		if err.Severity == SeverityCritical || err.Details != "" {
			if err.Details != "" {
				message += fmt.Sprintf("\n\nTechnical details: %s", err.Details)
			} else if template.TechnicalDetail != "" {
				message += fmt.Sprintf("\n\nTechnical details: %s", template.TechnicalDetail)
			}
		}

		return message
	}

	// Fallback to default message if no template found
	message := err.Message
	if err.UserMessage != "" {
		message = err.UserMessage
	}

	if len(err.Suggestions) > 0 {
		message += "\n\nSuggested actions:"
		for i, suggestion := range err.Suggestions {
			message += fmt.Sprintf("\n  %d. %s", i+1, suggestion)
		}
	}

	return message
}

// GenerateForGenericError generates a user-friendly message for a generic error
func (ufmg *UserFriendlyMessageGenerator) GenerateForGenericError(err error) string {
	errorMsg := strings.ToLower(err.Error())

	// Try to match common error patterns
	if strings.Contains(errorMsg, "timeout") {
		if template, exists := ufmg.templates[ErrorCodeNetworkTimeout]; exists {
			return template.UserMessage
		}
	}

	if strings.Contains(errorMsg, "connection refused") {
		if template, exists := ufmg.templates[ErrorCodeConnectionRefused]; exists {
			return template.UserMessage
		}
	}

	if strings.Contains(errorMsg, "permission denied") {
		if template, exists := ufmg.templates[ErrorCodePermissionDenied]; exists {
			return template.UserMessage
		}
	}

	if strings.Contains(errorMsg, "no such file") || strings.Contains(errorMsg, "file not found") {
		if template, exists := ufmg.templates[ErrorCodeFileNotFound]; exists {
			return template.UserMessage
		}
	}

	if strings.Contains(errorMsg, "build failed") || strings.Contains(errorMsg, "nix build") {
		if template, exists := ufmg.templates[ErrorCodeNixBuildFailed]; exists {
			return template.UserMessage
		}
	}

	// Default fallback
	return fmt.Sprintf("An error occurred: %s\n\nIf this problem persists, please report it to the nixai developers.", err.Error())
}

// AddTemplate adds a custom message template
func (ufmg *UserFriendlyMessageGenerator) AddTemplate(code ErrorCode, template MessageTemplate) {
	ufmg.templates[code] = template
}

// GetTemplate returns the template for an error code
func (ufmg *UserFriendlyMessageGenerator) GetTemplate(code ErrorCode) (MessageTemplate, bool) {
	template, exists := ufmg.templates[code]
	return template, exists
}

// GetAllTemplates returns all available templates
func (ufmg *UserFriendlyMessageGenerator) GetAllTemplates() map[ErrorCode]MessageTemplate {
	return ufmg.templates
}

// FormatErrorForDisplay formats an error for display to the user
func FormatErrorForDisplay(err error, includeDebugInfo bool) string {
	generator := NewUserFriendlyMessageGenerator()
	message := generator.GenerateUserFriendlyMessage(err)

	if includeDebugInfo {
		if nixaiErr, ok := err.(*NixAIError); ok {
			message += fmt.Sprintf("\n\n[DEBUG] Error Code: %s", nixaiErr.Code)
			message += fmt.Sprintf("\n[DEBUG] Severity: %s", nixaiErr.Severity)
			message += fmt.Sprintf("\n[DEBUG] Category: %s", nixaiErr.Category)
			message += fmt.Sprintf("\n[DEBUG] Retryable: %v", nixaiErr.Retryable)
			message += fmt.Sprintf("\n[DEBUG] Timestamp: %s", nixaiErr.Timestamp.Format(time.RFC3339))

			if len(nixaiErr.Context) > 0 {
				message += "\n[DEBUG] Context:"
				for key, value := range nixaiErr.Context {
					message += fmt.Sprintf("\n  %s: %v", key, value)
				}
			}

			if nixaiErr.StackTrace != "" {
				message += fmt.Sprintf("\n\n[DEBUG] Stack Trace:\n%s", nixaiErr.StackTrace)
			}
		}
	}

	return message
}
