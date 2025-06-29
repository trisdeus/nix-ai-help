// Package cli provides the command-line interface for nixai
package cli

import (
	"context"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// CreateUserConfigFromYAML creates a UserConfig from the YAML configuration
func CreateUserConfigFromYAML() (*config.UserConfig, error) {
	// Try to load from file first, then fall back to embedded
	yamlConfig, err := config.LoadEmbeddedYAMLConfig()
	if err != nil {
		return nil, err
	}

	// Convert YAMLConfig to UserConfig
	userConfig := &config.UserConfig{
		AIProvider:  yamlConfig.AIProvider,
		LogLevel:    yamlConfig.LogLevel,
		AIModels:    yamlConfig.AIModels,
		MCPServer:   yamlConfig.MCPServer,
		Nixos:       yamlConfig.Nixos,
		Diagnostics: yamlConfig.Diagnostics,
		Commands:    yamlConfig.Commands,
		AITimeouts:  yamlConfig.AITimeouts,
		Devenv:      yamlConfig.Devenv,
		CustomAI:    yamlConfig.CustomAI,
		Discourse:   yamlConfig.Discourse,
		Cache:       yamlConfig.Cache,
		Plugin:      yamlConfig.Plugin,
	}

	return userConfig, nil
}

// EnsureConfigHasProviders ensures the config has proper AI provider definitions
// If providers are empty, it loads from the embedded YAML configuration
//
// CRITICAL FIX: This function prevents the "provider 'ollama' is not configured" error
// that occurs when user config files have empty providers: {} due to incomplete config
// creation, config reset issues, or version upgrades. See docs/TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md
//
// This function MUST be called before creating any AI provider manager to ensure
// proper fallback to embedded YAML configuration when user config is incomplete.
func EnsureConfigHasProviders(cfg *config.UserConfig) (*config.UserConfig, error) {
	// Check if providers are empty or missing
	if cfg.AIModels.Providers == nil || len(cfg.AIModels.Providers) == 0 {
		// Load from embedded YAML and merge providers
		yamlConfig, err := config.LoadEmbeddedYAMLConfig()
		if err != nil {
			return cfg, err // Return original config if we can't load YAML
		}

		// Update the config with providers from YAML
		cfg.AIModels = yamlConfig.AIModels

		// Ensure default provider is set if empty
		if cfg.AIModels.SelectionPreferences.DefaultProvider == "" {
			cfg.AIModels.SelectionPreferences.DefaultProvider = "ollama"
		}
	}

	return cfg, nil
}

// GetAIProviderManager creates and returns a provider manager using the configuration system
func GetAIProviderManager(cfg *config.UserConfig, log *logger.Logger) *ai.ProviderManager {
	return ai.NewProviderManager(cfg, log)
}

// GetLegacyAIProvider gets a legacy AIProvider using the new ProviderManager system
func GetLegacyAIProvider(cfg *config.UserConfig, log *logger.Logger) (ai.AIProvider, error) {
	// If no config provided, create one from YAML
	if cfg == nil {
		var err error
		cfg, err = CreateUserConfigFromYAML()
		if err != nil {
			return ai.NewOllamaLegacyProvider("llama3"), nil
		}
	}

	manager := ai.NewProviderManager(cfg, log)

	// Get the configured default provider or fall back to ollama
	defaultProvider := cfg.AIModels.SelectionPreferences.DefaultProvider
	if defaultProvider == "" {
		defaultProvider = "ollama"
	}

	provider, err := manager.GetProvider(defaultProvider)
	if err != nil {
		// Fall back to ollama legacy provider on error
		return ai.NewOllamaLegacyProvider("llama3"), nil
	}

	// Use NewProviderWrapper to convert Provider to AIProvider
	return &ProviderToLegacyAdapter{provider: provider}, nil
}

// ProviderToLegacyAdapter adapts a Provider to the legacy AIProvider interface
type ProviderToLegacyAdapter struct {
	provider ai.Provider
}

// Query implements the legacy AIProvider interface
func (p *ProviderToLegacyAdapter) Query(prompt string) (string, error) {
	if provider, ok := p.provider.(interface {
		QueryWithContext(context.Context, string) (string, error)
	}); ok {
		return provider.QueryWithContext(context.Background(), prompt)
	}
	if provider, ok := p.provider.(interface{ Query(string) (string, error) }); ok {
		return provider.Query(prompt)
	}
	return "", context.DeadlineExceeded // or another suitable error
}

// InitializeAIProvider creates the appropriate AI provider based on configuration
// Deprecated: Use GetLegacyAIProvider() for new code
func InitializeAIProvider(cfg *config.UserConfig) ai.AIProvider {
	provider, err := GetLegacyAIProvider(cfg, logger.NewLogger())
	if err != nil {
		// Fall back to ollama legacy provider on error
		return ai.NewOllamaLegacyProvider("llama3")
	}
	return provider
}

// SummarizeBuildOutput extracts error messages from build output
func SummarizeBuildOutput(output string) string {
	lines := strings.Split(output, "\n")
	var summary []string
	for _, line := range lines {
		if strings.Contains(line, "error:") ||
			strings.Contains(line, "failed") ||
			strings.Contains(line, "cannot") {
			summary = append(summary, line)
		}
	}
	return strings.Join(summary, "\n")
}
