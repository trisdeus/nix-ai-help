# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**nixai** is a NixOS AI assistant written in Go 1.23+ that provides intelligent system administration through natural language commands. It uses clean architecture with modular AI provider management and specializes in NixOS configuration, troubleshooting, and hardware optimization.

## Development Commands

### Build Commands
```bash
# Recommended: Nix flake build
nix build

# Development build
just build

# Production build with optimizations  
just build-prod

# Multi-architecture builds
just build-all
```

### Test Commands
```bash
# Quick tests (CI equivalent)
just test

# Full comprehensive testing
just test-full

# Specific test categories
just test-mcp         # MCP integration tests
just test-vscode      # VS Code integration tests
just test-providers   # AI provider tests
just test-functions   # AI function tests
```

### Run Commands
```bash
# Direct execution
./nixai --help

# Development modes
just run
just run-interactive  # Modern TUI mode
just run-mcp         # MCP server mode

# Nix execution
nix run . -- --help
```

## Architecture Overview

### Clean Architecture Layers
- **CLI Layer** (`/internal/cli/`) - Command-line interface and command implementations
- **AI Layer** (`/internal/ai/`) - 26+ specialized agents, 29+ functions, multi-provider management
- **Core Layer** (`/internal/nixos/`, `/internal/packaging/`) - NixOS-specific logic and package analysis
- **Infrastructure** (`/pkg/`) - Logging, utilities, error handling

### Key Components
- **Agent System** (`/internal/ai/agent/`) - Role-based AI agents for different commands with specialized prompts
- **Function Calling** (`/internal/ai/function/`) - Structured AI functions with comprehensive testing
- **Multi-Provider Management** - Unified interface for 7 AI providers (Ollama, OpenAI, Gemini, Claude, Groq, LlamaCpp, GitHub Copilot)
- **Modern TUI** (`/internal/tui/`) - Bubble Tea-based terminal interface with two-panel layout
- **MCP Server** (`/internal/mcp/`) - Model Context Protocol integration for VS Code

### Configuration System
- **Main Config**: `configs/default.yaml` - Comprehensive configuration with AI provider settings
- **Privacy-First**: Defaults to local Ollama provider
- **Multi-Machine Support**: Flake-based deployment system

## Development Guidelines

### Adding New AI Functions
1. Create function in `/internal/ai/function/` with structured parameters
2. Add comprehensive tests in the same directory
3. Register in the function registry
4. Update agent prompts if needed

### Adding New Commands
1. Implement in `/internal/cli/` following existing patterns
2. Create specialized agent in `/internal/ai/agent/` if needed
3. Add command to main CLI registration
4. Include help documentation

### AI Provider Integration
- All providers implement common interface in `/internal/ai/`
- Configuration-driven provider selection
- Automatic fallback and health checking
- Provider-specific implementations in separate files

### Testing Requirements
- Unit tests for all new functions and components
- Integration tests for AI provider interactions
- MCP server tests for VS Code integration
- Use `just test` for quick feedback, `just test-full` for comprehensive testing

## Code Patterns

### Error Handling
- Centralized error handling in `/pkg/errors/`
- Structured error responses with context
- Analytics integration for error tracking

### Logging
- Centralized logging system in `/pkg/logger/`
- Structured logging with different levels
- Integration with AI operations for debugging

### Configuration Management
- YAML-based configuration system
- Environment variable overrides
- Validation and default value handling

## Important Notes

### AI Function Development
- All AI functions must have comprehensive test coverage
- Functions should be stateless and deterministic where possible
- Use structured parameters for reliable AI integration

### NixOS Integration
- Leverage existing NixOS parsing utilities in `/internal/nixos/`
- Follow Nix expression generation patterns
- Validate generated configurations before execution

### TUI Development
- Built on Bubble Tea framework
- Maintain accessibility without Unicode dependencies
- Follow two-panel layout pattern for consistency