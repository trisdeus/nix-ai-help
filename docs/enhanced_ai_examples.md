# Enhanced AI Examples

This document provides examples of how to use nixai's enhanced AI capabilities.

## Chain-of-Thought Reasoning

The AI now shows its step-by-step reasoning process:

```bash
nixai ask "How to configure nginx in NixOS?" --verbose
```

This will show the AI's reasoning chain, including:
- Problem identification
- Information gathering
- Solution formulation
- Validation steps

## Self-Correction

The AI automatically reviews and corrects its own responses:

```bash
nixai ask "What is the difference between flakes and channels?" --verbose
```

The AI will:
- Generate an initial response
- Review its own response for accuracy
- Correct any issues it identifies
- Present the final, corrected response

## Multi-Step Task Planning

For complex tasks, the AI breaks them down into actionable steps:

```bash
nixai ask "How to set up a development environment for Python and Django?" --verbose
```

The AI will:
- Analyze the task complexity
- Create a step-by-step plan
- Show prerequisites and dependencies
- Provide estimated completion times

## Confidence Scoring

Each response includes a confidence score:

```bash
nixai ask "How to enable SSH in NixOS?" --verbose
```

The response will include:
- A numerical confidence score (0.0-1.0)
- Detailed explanation of the score
- Quality indicators and warnings
- Recommendations for improvement

## Real Plugin Integration

nixai now supports dynamically loadable plugins:

```bash
# List available plugins
nixai plugin list

# Install a plugin
nixai plugin install /path/to/plugin.so

# Enable a plugin
nixai plugin enable my-plugin

# Use plugin commands
nixai my-plugin command --option value

# Disable a plugin
nixai plugin disable my-plugin

# Uninstall a plugin
nixai plugin uninstall my-plugin
```

## Using All Features Together

For the best experience, use all enhanced AI features together:

```bash
# Enable verbose mode to see all enhanced features
nixai ask "How to configure a web server with SSL in NixOS?" --verbose

# Use the TUI for interactive experience
nixai tui
```

In the TUI:
1. Type your query in natural language
2. Use Tab for intelligent command completion
3. Navigate with arrow keys
4. Press Enter to execute commands
5. View enhanced AI features in action

## Advanced Usage Examples

### Complex Configuration Task

```bash
nixai ask "How to set up a Kubernetes cluster with NixOS nodes?" --verbose
```

This query will trigger:
- Chain-of-Thought reasoning for cluster setup
- Self-correction to ensure accuracy
- Multi-step task planning with dependencies
- Confidence scoring for the response
- Plugin integration for Kubernetes-specific commands

### Development Environment Setup

```bash
nixai ask "Create a development environment for Rust with debugging tools" --verbose
```

This will show:
- Step-by-step setup process
- Tool installation instructions
- Configuration recommendations
- Best practices for debugging

### System Migration

```bash
nixai ask "How to migrate from Ubuntu to NixOS smoothly?" --verbose
```

This will provide:
- Migration planning with phases
- Risk assessment and mitigation
- Step-by-step migration process
- Rollback strategies

## Customization

You can customize the enhanced AI features:

```bash
# Enable/disable specific features
nixai config set ai.enable_reasoning true
nixai config set ai.enable_correction false
nixai config set ai.enable_planning true
nixai config set ai.enable_scoring true

# Set confidence thresholds
nixai config set ai.confidence_threshold 0.8

# Configure plugin directories
nixai config set plugin.directories "[\"/usr/share/nixai/plugins\", \"/home/user/.local/share/nixai/plugins\"]"
```

## Plugin Development

Create your own plugins to extend nixai's capabilities:

```bash
# Create a new plugin from template
nixai plugin create --template basic my-custom-plugin

# Build the plugin
cd my-custom-plugin
go build -buildmode=plugin -o my-custom-plugin.so .

# Install the plugin
nixai plugin install ./my-custom-plugin.so

# Use the plugin
nixai my-custom-plugin command
```

The plugin system supports:
- Dynamic loading/unloading
- Secure sandboxing
- Version management
- Marketplace integration
- Community sharing

These enhanced AI capabilities make nixai a more powerful and trustworthy tool for NixOS users of all skill levels.