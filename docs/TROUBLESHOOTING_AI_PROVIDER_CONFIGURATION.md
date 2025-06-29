# AI Provider Configuration Troubleshooting Guide

## Issue: "Failed to initialize AI provider: provider 'ollama' is not configured"

### Problem Description

When running `nixai ask` or other AI-powered commands, you may encounter the error:
```
❌ Failed to initialize AI provider: provider 'ollama' is not configured: provider 'ollama' not found
```

This error occurs when the user configuration file (`~/.config/nixai/config.yaml`) has an empty or incomplete AI providers configuration, specifically:

```yaml
ai_models:
    providers: {}  # ← Empty providers map causes the error
```

### Root Cause

The issue stems from a mismatch between two configuration systems:

1. **Embedded YAML Configuration** (`configs/default.yaml`) - Contains complete provider definitions
2. **User Configuration** (`~/.config/nixai/config.yaml`) - May be created with incomplete data

When the user config file is created with empty providers, the AI provider manager cannot find the requested provider (e.g., 'ollama'), causing initialization to fail.

### How This Happens

1. **Initial Setup**: User config file gets created with incomplete provider definitions
2. **Config Reset**: Old reset functions may create minimal configs instead of full defaults
3. **Manual Editing**: User accidentally removes or corrupts provider configurations
4. **Version Mismatch**: Upgrading from older versions that had different config structures

### Solution Implemented

A configuration validation and fallback system was implemented in `internal/cli/common_helpers.go`:

```go
// EnsureConfigHasProviders ensures the config has proper AI provider definitions
// If providers are empty, it loads from the embedded YAML configuration
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
```

This function is called in all AI command handlers before creating the provider manager:

```go
// Load configuration
cfg, err := config.LoadUserConfig()
if err != nil {
    return fmt.Errorf("Failed to load config: %w", err)
}

// Ensure config has proper provider definitions
cfg, err = EnsureConfigHasProviders(cfg)
if err != nil {
    return fmt.Errorf("Failed to ensure config providers: %w", err)
}

// Create AI provider manager
manager := ai.NewProviderManager(cfg, logger.NewLogger())
```

### Immediate Fix for Users

If you encounter this error, you can fix it immediately by:

#### Option 1: Reset Configuration
```bash
nixai config reset
```

#### Option 2: Remove and Recreate Config
```bash
rm ~/.config/nixai/config.yaml
nixai ask "test question"  # This will recreate the config
```

#### Option 3: Manual Verification
Check your config file:
```bash
cat ~/.config/nixai/config.yaml | grep -A 5 "providers:"
```

If it shows `providers: {}`, the config needs to be reset.

### Prevention Measures

#### For Developers

1. **Always use `EnsureConfigHasProviders`** before creating AI provider managers
2. **Test config reset functions** to ensure they create complete configurations
3. **Validate config loading** in integration tests
4. **Use embedded YAML as fallback** when user config is incomplete

#### For Configuration Functions

When implementing config-related functions:

```go
// ✅ Good: Use complete default config
func resetConfig() {
    cfg := config.DefaultUserConfig()  // This has complete providers
    err := config.SaveUserConfig(cfg)
    // ...
}

// ❌ Bad: Create minimal config manually
func resetConfig() {
    cfg := &config.UserConfig{
        AIProvider: "ollama",
        AIModels: config.AIModelsConfig{
            Providers: map[string]config.AIProviderConfig{}, // Empty!
        },
    }
    err := config.SaveUserConfig(cfg)
    // ...
}
```

#### For Code Review

When reviewing code that creates or modifies configurations:

1. Verify that `DefaultUserConfig()` is used for complete configs
2. Check that provider definitions are preserved during config operations
3. Ensure fallback mechanisms are in place for incomplete configs
4. Test configuration loading with empty/corrupted config files

### Testing

To test this issue and verify fixes:

```bash
# 1. Create empty providers config
echo "ai_models:
  providers: {}" > ~/.config/nixai/config.yaml

# 2. Try to run ask command (should work with fix)
nixai ask "test"

# 3. Verify providers are now populated
grep -A 10 "providers:" ~/.config/nixai/config.yaml
```

### Files Modified

The following files contain the fix:

- `internal/cli/common_helpers.go` - Added `EnsureConfigHasProviders()` function
- `internal/cli/direct_commands.go` - Applied fix to all AI command handlers
- `internal/cli/build_commands_enhanced.go` - Applied fix to build commands
- `internal/cli/commands.go` - Applied fix to legacy command handlers

### Related Issues

This fix also prevents similar issues with:

- Empty model configurations
- Missing selection preferences
- Incomplete provider definitions
- Configuration version mismatches

### Future Improvements

1. **Config Validation**: Add comprehensive config validation on load
2. **Migration System**: Implement config migration for version upgrades
3. **Health Checks**: Add `nixai doctor` command to diagnose config issues
4. **Better Error Messages**: Provide more specific error messages for config problems

### Conclusion

This issue was caused by a gap between the rich embedded configuration and the user configuration system. The implemented solution provides automatic fallback to ensure AI providers are always available, preventing this class of configuration errors from affecting users.

The fix is backward-compatible and transparent to users - if their config is incomplete, it's automatically fixed without requiring manual intervention.
