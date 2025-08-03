# GitHub Issue #46 Resolution: "ollama returned status 404"

## Problem Summary

User @mawkler reported getting "ollama returned status 404" error when trying to use nix-ai-help with Ollama. The user had Ollama running with the `deepseek-r1` model downloaded, but nixai was trying to use the hardcoded default model `llama3`.

## Root Cause Analysis

The issue was caused by:

1. **Hardcoded Default Model**: The system was hardcoded to use `llama3` as the default Ollama model
2. **No Model Validation**: There was no validation to check if the configured model actually exists on the user's Ollama installation
3. **Poor Error Messages**: The 404 error didn't provide helpful information about what models were available
4. **No Auto-Detection**: The system couldn't automatically detect and use available models

## Solution Implemented

### 1. Enhanced Model Configuration Support

- **Added DeepSeek Models**: Added `deepseek-r1` and `deepseek-r1:8b` to the supported models in `configs/default.yaml`
- **Flexible Model Matching**: Updated model validation to match base names (e.g., `llama3` matches `llama3:latest`)
- **Auto-Detection**: Implemented automatic model detection and fallback to available models

### 2. Improved Error Handling

- **Better 404 Messages**: Enhanced error messages to show available models when a 404 occurs
- **Model Validation**: Added `ValidateModel()` function to check if models exist before use
- **Helpful Suggestions**: Error messages now include commands to fix the issue

### 3. New Provider Management Commands

Added comprehensive provider management with:

```bash
# List all available providers and their status
nixai provider list

# Show available models for Ollama
nixai provider models --provider ollama

# Test a specific provider and model
nixai provider test --provider ollama --model deepseek-r1

# Show configuration help
nixai provider config
```

### 4. Auto-Detection and Fallback

- **Health Checks**: Added health checking for Ollama server
- **Model Discovery**: Automatically detect available models from Ollama
- **Smart Fallback**: If configured model doesn't exist, auto-switch to first available model
- **Informative Logging**: Clear logs about model validation and auto-switching

## Usage for the Reported Issue

For users with `deepseek-r1` model:

### Option 1: Use Provider Commands
```bash
# Check what models you have
nixai provider models --provider ollama

# Test with your specific model
nixai provider test --provider ollama --model deepseek-r1

# Use in regular commands
nixai ask "help me configure nginx" --provider ollama --model deepseek-r1
```

### Option 2: Update Configuration
Edit your nixai configuration to set:
```yaml
ai_provider: ollama
ai_model: deepseek-r1  # or deepseek-r1:latest
```

### Option 3: Use Environment Variables
```bash
export NIXAI_PROVIDER=ollama
export NIXAI_MODEL=deepseek-r1
nixai ask "help me with system configuration"
```

## Verification

The solution has been tested and verified to:

1. ✅ **Detect Available Models**: Successfully lists all Ollama models
2. ✅ **Auto-Match Model Names**: Matches `llama3` to `llama3:latest` automatically  
3. ✅ **Provide Better Errors**: Shows available models in 404 error messages
4. ✅ **Support DeepSeek**: Added DeepSeek models to configuration
5. ✅ **Auto-Fallback**: Automatically switches to available models when configured model is missing
6. ✅ **Test Successfully**: Provider test command works with all model types

## Example Output

```bash
$ nixai provider models --provider ollama
Available models for ollama:
========================
1. mistral-small3.1:latest
2. llama3:latest
3. deepseek-r1:latest
4. qwen2.5-coder:latest

$ nixai provider test --provider ollama --model deepseek-r1
Testing provider: ollama
Health check... ✅ Passed
Sending test query... ✅ Success
Response: AI provider test successful
```

## Technical Changes

### Files Modified:
- `internal/ai/ollama.go`: Enhanced model validation and error handling
- `internal/ai/manager.go`: Added auto-detection and fallback logic
- `configs/default.yaml`: Added DeepSeek and other model definitions
- `internal/cli/provider_commands.go`: New provider management commands
- `internal/cli/commands.go`: Integrated provider commands

### New Features:
- Model auto-detection from Ollama API
- Intelligent model name matching (base name to full name with tags)
- Comprehensive provider testing and diagnostics
- Better error messages with actionable suggestions
- Health checking for Ollama connectivity

## Impact

This solution resolves the GitHub issue and provides:

- **Better User Experience**: Clear error messages and auto-detection
- **Easier Configuration**: Multiple ways to specify models
- **Better Debugging**: Comprehensive provider testing commands
- **Future-Proof**: Support for any Ollama model, not just hardcoded ones
- **Backward Compatible**: Existing configurations continue to work

The user should now be able to use their `deepseek-r1` model without any issues!