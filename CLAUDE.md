# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**nixai** is a revolutionary NixOS AI assistant written in Go 1.23+ that provides intelligent system administration through natural language commands. It uses clean architecture with modular AI provider management and specializes in NixOS configuration, troubleshooting, and hardware optimization. **Current version: 2.2.0**

## Development Environment

### Host System
- **Operating System**: NixOS (flakes-enabled)
- **Build System**: Nix + Go 1.23+
- **Development Tools**: just, golangci-lint, git
- **Package Manager**: Nix flakes (preferred) + Go modules

### Architecture
- **Runtime**: NixOS with flakes
- **Build Tools**: Nix build system + justfile for development
- **Dependencies**: Managed via Nix flakes and Go modules
- **Deployment**: Nix flake-based with `nix run` support

## Development Commands

### Build Commands
```bash
# Recommended: Nix flake build (clean version 2.2.0)
nix build
./result/bin/nixai --version  # Shows: nixai version 2.2.0

# Development build (with git info)
just build
./nixai --version             # Shows: nixai version v2.2.0-dirty

# Production build with optimizations  
just build-prod

# Multi-architecture builds
just build-all

# Instant run without installation
nix run . -- --help
nix run . -- tui            # Launch intelligent TUI
nix run . -- health status  # Real system monitoring
```

### Test Commands
```bash
# Quick tests (CI equivalent) - Core packages only
just test

# Full comprehensive testing - All packages including CLI/TUI
just test-full

# Specific test categories
just test-mcp         # MCP integration tests
just test-vscode      # VS Code integration tests
just test-providers   # AI provider tests
just test-functions   # AI function tests
```

### Run Commands
```bash
# Direct execution (development)
./nixai --help
./nixai tui                  # Launch intelligent TUI with AI search
./nixai health status        # Real system health monitoring

# Development modes
just run
just run-interactive         # Modern TUI mode with intelligent search
just run-mcp                # MCP server mode

# Nix execution (recommended)
nix run . -- --help
nix run . -- ask "help me with health status"  # Natural language queries
nix run . -- tui           # Intelligent TUI interface
```

## Latest Major Features (v2.2.0)

### 🤖 Intelligent TUI with AI-Powered Command Search
- **Location**: `/internal/tui/tui.go` + `/internal/tui/intelligent_discovery.go`
- **Feature**: Natural language query understanding with AI-powered command discovery
- **Usage**: Type "help me with health status" in TUI for intelligent suggestions with fuzzy matching
- **Implementation**: `intelligentCommandSearch()` with relevance scoring, usage analytics, and context-aware recommendations

### 🔍 Real System Monitoring (No Mock Data)
- **Health System**: `/internal/health/predictor.go` - Real `/proc/*` data collection
- **Performance**: `/internal/cli/performance_commands.go` - Actual cache statistics
- **Fleet Management**: `/internal/fleet/manager.go` - Real SSH connectivity and health checks
- **Workflow System**: `/internal/workflow/executor.go` - Actual system command execution

### 🔧 Enhanced Hardware Detection and Profiling
- **Location**: `/internal/hardware/enhanced_detector.go` + detection/profiling methods
- **Feature**: Comprehensive hardware analysis with detailed CPU, GPU, memory, storage, network profiling
- **Implementation**: Hardware caching with TTL, vendor-specific capability detection, security feature analysis
- **Integration**: Enhanced hardware function with fallback to basic detection for reliability

### 📊 Configuration Dependency Analysis  
- **Location**: `/internal/nixos/dependency/` - Complete dependency analysis system
- **Feature**: Rule-based dependency detection, conflict resolution, hardware-aware recommendations
- **Implementation**: 25+ predefined rules, dependency graph generation, circular dependency detection
- **Function**: AI-powered `dependency-analysis` function with actionable insights and config suggestions

### 📈 CPU-Aware System Thresholds
- **Load Average**: Critical threshold = CPU cores × 1.5 (e.g., 192.0 for 128-core system)
- **Network Usage**: Heuristic calculation, not cumulative bytes since boot
- **Process Count**: Actual running processes from `/proc/` PID directories

## Architecture Overview

### Clean Architecture Layers
- **CLI Layer** (`/internal/cli/`) - Command-line interface and command implementations
- **AI Layer** (`/internal/ai/`) - 26+ specialized agents, 30+ functions, multi-provider management
- **Core Layer** (`/internal/nixos/`, `/internal/packaging/`) - NixOS-specific logic and package analysis
- **Infrastructure** (`/pkg/`) - Logging, utilities, error handling

### Key Components
- **Agent System** (`/internal/ai/agent/`) - Role-based AI agents for different commands with specialized prompts
- **Function Calling** (`/internal/ai/function/`) - Structured AI functions with comprehensive testing
- **Multi-Provider Management** - Unified interface for 7 AI providers (Ollama, OpenAI, Gemini, Claude, Groq, LlamaCpp, GitHub Copilot)
- **Intelligent TUI** (`/internal/tui/`) - Bubble Tea-based terminal interface with AI-powered command discovery and fuzzy search
- **Enhanced Hardware Detection** (`/internal/hardware/`) - Comprehensive hardware profiling with caching and vendor-specific analysis  
- **Dependency Analysis System** (`/internal/nixos/dependency/`) - Rule-based configuration analysis with conflict detection
- **MCP Server** (`/internal/mcp/`) - Model Context Protocol integration for VS Code/Neovim
- **Real Monitoring** (`/internal/health/`, `/internal/fleet/`) - Authentic system metrics collection

### Configuration System
- **Main Config**: `configs/default.yaml` - Comprehensive configuration with AI provider settings
- **Privacy-First**: Defaults to local Ollama provider
- **Multi-Machine Support**: Flake-based deployment system
- **Version Management**: Git tags + Nix flakes for clean versioning

## Development Guidelines

### Latest Code Changes (v2.2.0)
1. **All mock data eliminated** - Health, performance, fleet, and workflow systems now use real data
2. **Intelligent TUI implemented** - Natural language command search with AI suggestions
3. **CPU-aware thresholds** - System monitoring adapts to hardware capabilities
4. **Real action execution** - Workflow system performs actual file operations and system commands

### Adding New AI Functions
1. Create function in `/internal/ai/function/` with structured parameters
2. Add comprehensive tests in the same directory
3. Register in the function registry
4. Update agent prompts if needed
5. **NEW**: Ensure function uses real system data, not mock/simulated data

### Adding New Commands
1. Implement in `/internal/cli/` following existing patterns
2. Create specialized agent in `/internal/ai/agent/` if needed
3. Add command to main CLI registration
4. Include help documentation
5. **NEW**: Add command to TUI's intelligent search keyword mappings in `/internal/tui/tui.go`

### AI Provider Integration
- All providers implement common interface in `/internal/ai/`
- Configuration-driven provider selection
- Automatic fallback and health checking
- Provider-specific implementations in separate files

### Testing Requirements
- Unit tests for all new functions and components
- Integration tests for AI provider interactions
- MCP server tests for VS Code integration
- **Real system testing**: Verify actual data collection, not mock responses
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

### Real Data Collection (NEW)
- **Health Monitoring**: Read from `/proc/stat`, `/proc/meminfo`, `/proc/loadavg`
- **Performance Metrics**: Use actual cache hit/miss rates, not placeholders
- **Network Usage**: Calculate actual utilization, not cumulative bytes
- **Process Count**: Count real running processes, not total since boot

## Important Notes

### System Requirements
- **Host OS**: NixOS with flakes enabled
- **Nix Version**: 2.4+ with flake support
- **Go Version**: 1.23+ (managed via Nix)
- **Dependencies**: All managed through Nix flakes

### Build System Strategy
- **Development**: `just build` uses git tags with dirty status (v2.2.0-dirty)
- **Distribution**: `nix build` uses clean version numbers (2.2.0)
- **CI/Package**: Nix flakes provide reproducible builds

### AI Function Development
- All AI functions must have comprehensive test coverage
- Functions should be stateless and deterministic where possible
- Use structured parameters for reliable AI integration
- **NEW**: Functions must use real system data - no mock or simulated responses

### NixOS Integration
- Leverage existing NixOS parsing utilities in `/internal/nixos/`
- Follow Nix expression generation patterns
- Validate generated configurations before execution
- **NEW**: Test with actual NixOS configurations, not synthetic examples

### TUI Development
- Built on Bubble Tea framework
- Maintain accessibility without Unicode dependencies
- Follow two-panel layout pattern for consistency
- **NEW**: Implement intelligent search features using keyword mappings in `intelligentCommandSearch()`

### System Monitoring Best Practices
- **Real Data Only**: Never use mock, simulated, or placeholder data
- **CPU-Aware**: Scale thresholds based on actual hardware capabilities
- **Accurate Calculations**: Use proper formulas for network usage, load averages, etc.
- **Process Monitoring**: Count actual running processes, not historical totals

## Latest Changes Tracking

### Completed in v2.2.0
- ✅ Intelligent TUI with AI-powered command search
- ✅ Real system monitoring (eliminated all mock data)
- ✅ CPU-aware threshold calculations
- ✅ Accurate network utilization measurement
- ✅ Real workflow action execution
- ✅ Enhanced hardware detection with comprehensive profiling system
- ✅ Configuration dependency analysis with conflict detection  
- ✅ Version update to 2.2.0 across all build systems
- ✅ Comprehensive documentation updates

### Current Priorities
1. **Maintain Real Data Standards**: Ensure all new features use authentic system data
2. **Enhance Intelligent Search**: Expand keyword mappings and improve relevance scoring
3. **Performance Optimization**: Optimize real-time monitoring with minimal system impact
4. **User Experience**: Continue improving natural language interface capabilities

### Development Workflow
```bash
# Standard development cycle
nix develop              # Enter dev environment
just build              # Development build
just test               # Quick tests
./nixai tui             # Test intelligent TUI
./nixai health status   # Test real monitoring
nix build              # Clean production build
```

### Version Management
- **Git Tags**: Used by justfile for development builds (shows dirty status)
- **Nix Flakes**: Provide clean version numbers for distribution
- **Current Version**: 2.2.0 (updated across all systems)

## Best Practices for Contributors

1. **Real Data First**: Always implement actual system monitoring, never use mock data
2. **Test Thoroughly**: Verify functionality with real NixOS systems
3. **Intelligent Integration**: Add new commands to TUI keyword mappings
4. **Documentation**: Update both technical docs and user guides
5. **Version Consistency**: Maintain version synchronization across build systems
6. **NixOS Native**: Leverage Nix flakes and NixOS-specific features

## Future Development

### Planned Enhancements
- Enhanced intelligent search with machine learning
- Real-time collaborative features
- Advanced system prediction capabilities
- Expanded hardware optimization

### Architecture Evolution
- Continue clean architecture principles
- Expand real-time monitoring capabilities
- Enhance AI provider ecosystem
- Improve natural language understanding

This project represents the cutting edge of NixOS system management with AI-powered intelligence and authentic system monitoring. All development should maintain these high standards for real data collection and intelligent user interaction.