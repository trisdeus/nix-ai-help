# AI Provider Configuration Health Check Command

## Overview

The `nixai doctor` command includes comprehensive AI provider configuration validation to prevent and diagnose the "provider 'X' is not configured" error.

## Health Check Features

### AI Provider Validation

When running `nixai doctor`, the system performs the following AI provider checks:

```bash
nixai doctor
```

**Output Example:**

```text
🩺 nixai System Health Check
============================

✅ Configuration System
   ├─ User config file: ~/.config/nixai/config.yaml
   ├─ System config file: /etc/nixai/config.yaml  
   ├─ Embedded config: Available
   └─ Config validation: PASSED

✅ AI Provider Configuration
   ├─ Providers defined: 7 (ollama, openai, gemini, claude, groq, llamacpp, custom)
   ├─ Default provider: ollama
   ├─ Provider availability: ollama ✅, openai ❌ (no API key), gemini ❌ (no API key)
   └─ Fallback system: CONFIGURED

⚠️  AI Provider Connectivity  
   ├─ Ollama: CONNECTED (llama3 model available)
   ├─ OpenAI: NOT CONFIGURED (set OPENAI_API_KEY)
   ├─ Gemini: NOT CONFIGURED (set GEMINI_API_KEY)
   └─ Claude: NOT CONFIGURED (set CLAUDE_API_KEY)

✅ MCP Server
   ├─ Server status: RUNNING (port 39847)
   ├─ Documentation sources: 5 configured
   └─ VS Code integration: AVAILABLE

✅ Context System
   ├─ Context detection: FUNCTIONAL
   ├─ Cache status: VALID (last updated 2 minutes ago)
   └─ System detection: NixOS 24.05, Flakes enabled, Home Manager standalone

🎯 Recommendations:
   • Ollama is working - no action required for local AI
   • To use cloud providers, set API keys: OPENAI_API_KEY, GEMINI_API_KEY, CLAUDE_API_KEY
   • All core functionality is operational
```

### Configuration Diagnostic Commands

For more detailed AI provider diagnostics:

```bash
# Check specific provider configuration
nixai config get ai_models.providers.ollama

# Test AI provider connectivity
nixai doctor --provider ollama

# Validate configuration structure
nixai config validate

# Reset corrupted configuration
nixai config reset
```

### Automated Fix Suggestions

The doctor command provides automated fix suggestions:

```bash
# If providers are empty, doctor suggests:
❌ AI Provider Configuration
   └─ Empty providers configuration detected
   
🔧 Suggested Fix:
   nixai config reset
   
# This will restore complete provider definitions from embedded configuration
```

### Integration with IDE Health Checks

The health check system integrates with:

- **VS Code**: Through MCP server health endpoint
- **Neovim**: Via Lua health check functions  
- **CLI**: Direct command-line validation

## Related Commands

- `nixai config reset` - Reset configuration to defaults
- `nixai config validate` - Validate configuration structure
- `nixai context status` - Check context system health
- `nixai mcp-server status` - Check MCP server health

## Troubleshooting References

- [AI Provider Configuration Troubleshooting](TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md)
- [Developer Guidelines](DEVELOPER_GUIDELINES.md) - For developers working on configuration
- [Configuration Guide](config.md) - Complete configuration reference

---

**This health check system is part of nixai's comprehensive approach to preventing and diagnosing configuration issues before they affect users.**
