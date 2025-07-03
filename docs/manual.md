# nixai manual - Built-in Manual System

The `nixai manual` command provides a comprehensive built-in manual system with detailed documentation for all nixai commands and concepts.

## Overview

The built-in manual system offers extensive documentation directly within nixai, eliminating the need to consult external documentation for most use cases. It includes detailed command references, examples, troubleshooting guides, and conceptual explanations.

## Usage

```bash
# Show manual index
nixai manual

# Get help for a specific command
nixai manual <command-name>

# Browse by category
nixai manual --category <category>

# Search manual content
nixai manual --search <search-term>

# Interactive manual navigation
nixai manual --interactive

# List all topics
nixai manual --list
```

## Features

### 📚 Comprehensive Documentation

The manual system includes documentation for:
- All 40+ nixai commands with detailed explanations
- Usage examples and common patterns
- Configuration options and best practices
- Troubleshooting guides and solutions
- Conceptual explanations (agents, providers, workflows)

### 🔍 Powerful Search

- **Full-text search** through all manual content
- **Category browsing** for organized exploration
- **Related topic suggestions** for discovery
- **Similar topic matching** for typos and partial matches

### 🎯 Interactive Navigation

- **Interactive mode** with command-line interface
- **Real-time help** during manual browsing
- **Topic cross-references** with "See Also" sections
- **Contextual examples** for immediate application

### 📖 Rich Content Format

- **Formatted output** with syntax highlighting
- **Structured sections** (description, usage, examples)
- **Cross-references** between related commands
- **Progressive detail levels** from basic to advanced

## Manual Categories

### Core Commands
- **ask** - AI question interface and assistance
- Essential commands for basic nixai usage

### Interface
- **web** - Modern web dashboard interface
- **tui** - Enhanced terminal user interface
- **interactive** - Interactive command-line interface

### Development
- **build** - Build system management and troubleshooting
- **devenv** - Development environment creation
- **flake** - Nix flake management and operations
- **package-repo** - Repository analysis and packaging

### Configuration
- **config** - Configuration management
- **configure** - Interactive configuration generation
- **templates** - Configuration templates
- **snippets** - Configuration snippets

### Management
- **fleet** - Fleet management for multiple machines
- **version-control** - Git-like configuration versioning
- **machines** - Multi-machine deployment
- **migrate** - Migration assistance
- **store** - Nix store management
- **gc** - Garbage collection

### Monitoring
- **doctor** - System health diagnostics
- **diagnose** - System diagnostics and troubleshooting
- **performance** - Performance monitoring
- **logs** - Log analysis
- **package-monitor** - Package monitoring

### System
- **hardware** - Hardware detection and optimization
- **system-info** - System information display

### Extension
- **plugin** - Plugin system management

### Concepts
- **agents** - AI agent system explanation
- **providers** - AI provider configuration

## Examples

### Basic Usage

```bash
# View the manual index
nixai manual

# Get help for the ask command
nixai manual ask

# Learn about the web interface
nixai manual web

# Understand AI agents
nixai manual agents
```

### Category Browsing

```bash
# Browse development commands
nixai manual --category Development

# View all interface options
nixai manual --category Interface

# Explore configuration tools
nixai manual --category Configuration
```

### Search Functionality

```bash
# Search for flake-related content
nixai manual --search flakes

# Find plugin information
nixai manual --search plugin

# Look for security topics
nixai manual --search security
```

### Interactive Mode

```bash
# Start interactive manual
nixai manual --interactive

# Interactive commands:
manual> help           # Show interactive help
manual> list           # List all topics
manual> search flakes  # Search for term
manual> ask            # View ask command help
manual> quit           # Exit interactive mode
```

## Manual Content Structure

Each manual entry includes:

### 📋 **Header Information**
- Command title and category
- Brief description
- Command classification

### 📖 **Detailed Content**
- Comprehensive feature explanation
- Usage patterns and workflows
- Configuration options
- Integration capabilities

### 💡 **Practical Examples**
- Real-world usage scenarios
- Command-line examples with options
- Common workflow demonstrations
- Best practice implementations

### 🔗 **Cross-References**
- Related commands and topics
- Workflow integration suggestions
- Conceptual background links
- Advanced usage pointers

## Advanced Features

### Integration with Help System

The manual system integrates seamlessly with nixai's existing help:

```bash
# Standard help
nixai --help                # General help
nixai ask --help           # Command-specific help

# Enhanced manual
nixai manual ask           # Comprehensive manual entry
nixai manual --search ask  # Find all ask-related content
```

### Context-Aware Assistance

The manual system provides context-aware help:

```bash
# Get specific help for your use case
nixai manual fleet         # Fleet management details
nixai manual web           # Web interface comprehensive guide
nixai manual getting-started  # New user orientation
```

### Offline Documentation

All manual content is built into nixai:
- **No internet required** for documentation access
- **Consistent versioning** with nixai releases
- **Always up-to-date** with installed features
- **Fast access** without external dependencies

## Best Practices

### 🎯 **Getting Started**
1. Start with `nixai manual` to see available topics
2. Use `nixai manual getting-started` for new user guidance
3. Explore categories with `nixai manual --category <name>`
4. Use search when looking for specific functionality

### 🔍 **Finding Information**
1. **Search first**: Use `--search` for specific terms
2. **Browse categories**: Use `--category` for organized exploration
3. **Follow cross-references**: Check "See Also" sections
4. **Use interactive mode**: For guided exploration

### 📚 **Learning Workflows**
1. **Start with concepts**: Understand agents and providers
2. **Learn core commands**: Master ask, build, diagnose
3. **Explore interfaces**: Try web and tui commands
4. **Advanced features**: Fleet management and version control

## Integration Examples

### With Development Workflow

```bash
# Learn about development commands
nixai manual --category Development

# Get specific flake help
nixai manual flake

# Understand build troubleshooting
nixai manual build

# Set up development environment
nixai manual devenv
```

### With System Administration

```bash
# System health and monitoring
nixai manual doctor
nixai manual diagnose
nixai manual performance

# Configuration management
nixai manual configure
nixai manual templates
nixai manual version-control
```

### With Multi-machine Management

```bash
# Fleet management concepts
nixai manual fleet

# Machine deployment
nixai manual machines

# Version control for configurations
nixai manual version-control
```

## Troubleshooting

### Manual Not Working

```bash
# Verify manual is available
nixai manual --help

# Check for updates
nixai --version

# Rebuild if needed
nix build
```

### Missing Topics

```bash
# List all available topics
nixai manual --list

# Search for partial matches
nixai manual --search <partial-term>

# Check category organization
nixai manual --category <category>
```

### Interactive Mode Issues

```bash
# Exit and restart
manual> quit
nixai manual --interactive

# Use non-interactive mode
nixai manual <topic>
```

The built-in manual system provides comprehensive, accessible documentation for all nixai functionality, making it easy to discover features, learn workflows, and troubleshoot issues without leaving the command line.

For external documentation and community resources, see [community resources](community.md) and the [main documentation](../README.md).