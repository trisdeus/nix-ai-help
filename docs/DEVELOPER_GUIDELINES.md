# Developer Guidelines for nixai

## Table of Contents

- [AI Provider Configuration](#ai-provider-configuration)
- [Testing Requirements](#testing-requirements)
- [Code Review Checklist](#code-review-checklist)
- [Common Patterns](#common-patterns)
- [Error Handling](#error-handling)

---

## AI Provider Configuration

### ⚠️ CRITICAL: Configuration Loading Pattern

**Always use the `EnsureConfigHasProviders` pattern when working with AI providers.** This prevents the ["provider not configured" error](TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md) that can occur when user configurations have empty provider definitions.

#### ✅ Correct Pattern

```go
import (
    "nix-ai-help/internal/ai"
    "nix-ai-help/internal/config"
    "nix-ai-help/pkg/logger"
)

func handleAICommand() error {
    // 1. Load user configuration
    cfg, err := config.LoadUserConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // 2. CRITICAL: Ensure config has provider definitions
    cfg, err = EnsureConfigHasProviders(cfg)
    if err != nil {
        return fmt.Errorf("failed to ensure config providers: %w", err)
    }

    // 3. Create AI provider manager
    manager := ai.NewProviderManager(cfg, logger.NewLogger())
    
    // 4. Get provider and proceed
    provider, err := manager.GetProvider("ollama")
    if err != nil {
        return fmt.Errorf("failed to initialize AI provider: %w", err)
    }
    
    // ... use provider
}
```

#### ❌ Incorrect Pattern

```go
// DON'T DO THIS - Can fail with "provider not configured" error
func handleAICommand() error {
    cfg, err := config.LoadUserConfig()
    if err != nil {
        return err
    }
    
    // Missing EnsureConfigHasProviders!
    manager := ai.NewProviderManager(cfg, logger.NewLogger())
    provider, err := manager.GetProvider("ollama") // Can fail!
    // ...
}
```

### Provider Initialization Guidelines

1. **Use ProviderManager**: Always use the unified `ai.ProviderManager` for provider creation
2. **Legacy Compatibility**: Use `GetLegacyAIProvider()` when interfacing with old code
3. **Error Handling**: Provide specific error messages for configuration issues
4. **Fallback Strategy**: Implement graceful degradation to Ollama when possible

#### Example: Legacy Interface Compatibility

```go
func getLegacyProvider(cfg *config.UserConfig) (ai.AIProvider, error) {
    // Ensure config has providers
    cfg, err := EnsureConfigHasProviders(cfg)
    if err != nil {
        return nil, err
    }
    
    // Use helper function for legacy compatibility
    return GetLegacyAIProvider(cfg, logger.NewLogger())
}
```

---

## Testing Requirements

### Configuration Testing

All new features that load configurations must include tests for:

1. **Empty Provider Configuration**: Test behavior when `providers: {}` is present
2. **Missing Configuration File**: Test behavior when config file doesn't exist  
3. **Corrupted Configuration**: Test behavior with malformed YAML
4. **Provider Fallback**: Test fallback mechanisms work correctly

#### Test Template

```go
func TestAIProviderConfigurationHandling(t *testing.T) {
    t.Run("EmptyProvidersConfig", func(t *testing.T) {
        // Create config with empty providers
        cfg := &config.UserConfig{
            AIModels: config.AIModelsConfig{
                Providers: map[string]config.AIProviderConfig{},
            },
        }
        
        // Test that EnsureConfigHasProviders fixes it
        result, err := EnsureConfigHasProviders(cfg)
        require.NoError(t, err)
        assert.NotEmpty(t, result.AIModels.Providers)
        assert.Contains(t, result.AIModels.Providers, "ollama")
    })
    
    t.Run("MissingConfigFile", func(t *testing.T) {
        // Test behavior when config file doesn't exist
        // Should fall back to embedded configuration
    })
}
```

### Integration Testing

1. **Provider Initialization**: Test each provider can be initialized correctly
2. **Configuration Loading**: Test configuration loading from different sources
3. **Error Scenarios**: Test error handling for various failure modes
4. **Fallback Mechanisms**: Test provider fallback logic

---

## Code Review Checklist

### For AI Provider Related Changes

- [ ] **Configuration Loading**: Does the code use `EnsureConfigHasProviders`?
- [ ] **Error Handling**: Are configuration errors handled gracefully?
- [ ] **Provider Creation**: Is `ai.NewProviderManager` used correctly?
- [ ] **Legacy Compatibility**: If interfacing with old code, is `GetLegacyAIProvider` used?
- [ ] **Testing**: Are configuration edge cases tested?
- [ ] **Documentation**: Are configuration requirements documented?

### For Configuration Related Changes

- [ ] **Complete Configs**: Does config creation use `config.DefaultUserConfig()`?
- [ ] **Provider Preservation**: Are provider definitions preserved during updates?
- [ ] **Validation**: Is configuration validated after creation/modification?
- [ ] **Migration**: Are version compatibility issues considered?

### For CLI Command Changes

- [ ] **Provider Initialization**: Is the standard pattern followed?
- [ ] **Progress Indicators**: Are long operations (API calls) indicated to user?
- [ ] **Error Messages**: Are error messages actionable and specific?
- [ ] **Help Text**: Is help text updated with examples?

---

## Common Patterns

### CLI Command Structure

```go
func handleCommand(args []string) error {
    // 1. Parse arguments and validate input
    if len(args) < 1 {
        return fmt.Errorf("usage: nixai command <required-arg>")
    }
    
    // 2. Load and ensure configuration
    cfg, err := config.LoadUserConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    cfg, err = EnsureConfigHasProviders(cfg)
    if err != nil {
        return fmt.Errorf("failed to ensure config: %w", err)
    }
    
    // 3. Initialize AI provider
    manager := ai.NewProviderManager(cfg, logger.NewLogger())
    provider, err := manager.GetProvider(cfg.AIModels.SelectionPreferences.DefaultProvider)
    if err != nil {
        // Try fallback to ollama
        provider, err = manager.GetProvider("ollama")
        if err != nil {
            return fmt.Errorf("failed to initialize AI provider: %w", err)
        }
    }
    
    // 4. Show progress for long operations
    fmt.Print(utils.FormatInfo("Processing... "))
    
    // 5. Execute operation
    result, err := provider.Query(prompt)
    if err != nil {
        fmt.Println(utils.FormatError("failed"))
        return fmt.Errorf("AI query failed: %w", err)
    }
    
    fmt.Println(utils.FormatSuccess("done"))
    
    // 6. Format and display results
    fmt.Println(utils.FormatMarkdown(result))
    
    return nil
}
```

### Configuration Management

```go
// ✅ Good: Complete configuration creation
func createDefaultConfig() *config.UserConfig {
    return config.DefaultUserConfig() // Has all provider definitions
}

// ✅ Good: Configuration updates that preserve providers
func updateConfig(cfg *config.UserConfig, newSetting string) error {
    // Ensure providers exist before updating
    cfg, err := EnsureConfigHasProviders(cfg)
    if err != nil {
        return err
    }
    
    // Update specific setting
    cfg.SomeSetting = newSetting
    
    return config.SaveUserConfig(cfg)
}

// ❌ Bad: Manual configuration creation
func createBadConfig() *config.UserConfig {
    return &config.UserConfig{
        AIProvider: "ollama",
        AIModels: config.AIModelsConfig{
            Providers: make(map[string]config.AIProviderConfig), // Empty!
        },
    }
}
```

---

## Error Handling

### Configuration Errors

Provide actionable error messages that guide users to solutions:

```go
// ✅ Good: Actionable error message
if len(cfg.AIModels.Providers) == 0 {
    return fmt.Errorf("AI providers not configured. Run 'nixai config reset' to fix this issue. See docs/TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md for details")
}

// ❌ Bad: Vague error message  
if len(cfg.AIModels.Providers) == 0 {
    return fmt.Errorf("configuration error")
}
```

### Provider Errors

Implement graceful degradation and clear fallback messaging:

```go
provider, err := manager.GetProvider("claude")
if err != nil {
    log.Warn("Failed to initialize Claude provider: %v", err)
    log.Info("Falling back to Ollama provider...")
    
    provider, err = manager.GetProvider("ollama")
    if err != nil {
        return fmt.Errorf("failed to initialize any AI provider. Ensure Ollama is installed and running, or configure API keys for cloud providers. See docs/ai-providers.md for setup instructions")
    }
    
    fmt.Println(utils.FormatWarning("Using Ollama fallback provider"))
}
```

---

## Documentation Requirements

When adding new features:

1. **Update Help Text**: Ensure all commands have current help text with examples
2. **Update README**: Add new features to the feature list and examples
3. **Update Manual**: Update `docs/MANUAL.md` with comprehensive documentation
4. **Provider Documentation**: Update `docs/ai-providers.md` for provider-related changes
5. **Configuration Documentation**: Update `docs/config.md` for configuration changes

### Documentation Templates

#### CLI Help Text
```go
Use: nixai command [options] <args>

Description:
  Brief description of what the command does and its purpose.

Examples:
  nixai command example-arg          # Basic usage
  nixai command --option value arg   # With options
  
Options:
  -o, --option string   Description of option (default "value")
  -h, --help           Show this help message

For more information: docs/command-name.md
```

---

## Related Documentation

- [AI Provider Configuration Troubleshooting](TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md)
- [AI Providers Guide](ai-providers.md)
- [Configuration Guide](config.md)
- [Testing Guidelines](../tests/README.md)

---

**Last Updated**: June 29, 2025  
**Version**: 1.0.7+

> These guidelines are mandatory for all contributors to ensure consistency, reliability, and maintainability of the nixai project.
