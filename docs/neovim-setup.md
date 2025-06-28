# nixa## 🆕 Enhanced Features

The `nixai neovim-setup` command features **comprehensive CLI integration** with advanced Neovim management:

### ✨ **CLI Features**
- **🎯 Parameter Input**: Neovim configuration options and plugin selection through CLI interface
- **📊 Real-Time Setup Progress**: Live Neovim installation and configuration progress
- **⌨️ Command Discovery**: Enhanced options with support for 2 subcommands and 5 flags
- **🔧 Setup Wizard**: Step-by-step Neovim configuration with AI-guided recommendations
- **📋 Context-Aware Setup**: Automatic detection of existing Neovim configuration for seamless integrationetup

Complete Neovim setup and configuration for NixOS development with MCP integration, plugin management, and AI-powered assistance.

---

## 🆕 TUI Integration & Enhanced Features

The `nixai neovim-setup` command now features **comprehensive TUI integration** with advanced Neovim management:

### ✨ **Interactive TUI Features**
- **🎯 Interactive Parameter Input**: Neovim configuration options and plugin selection through modern TUI interface
- **📊 Real-Time Setup Progress**: Live Neovim installation and configuration progress within the TUI
- **⌨️ Command Discovery**: Enhanced command browser with `[INPUT]` indicators for 5 subcommands and 2 flags
- **🔧 Interactive Setup Wizard**: Step-by-step Neovim configuration with AI-guided plugin recommendations
- **📋 Context-Aware Setup**: Integration with existing NixOS configuration for seamless Neovim integration

### 📝 **Advanced Neovim Integration Features**
- **🧠 AI-Powered Configuration**: Intelligent Neovim configuration generation based on development needs
- **🔌 MCP Server Integration**: Direct integration with Model Context Protocol for enhanced AI assistance
- **📦 Smart Plugin Management**: AI-recommended plugin selection with dependency resolution
- **🎨 Theme and UI Optimization**: Automated theme selection and UI configuration for optimal productivity
- **⚡ Performance Optimization**: Lazy loading configuration and performance tuning for large codebases
- **🔍 Language Server Integration**: Automatic LSP configuration for Nix, Go, Python, TypeScript, and more
- **🛠️ Development Workflow Integration**: Integration with Git, terminal, and project management tools

## Command Structure

```sh
nixai neovim-setup [subcommand] [flags]
```

### Available Subcommands (5 Total)

| Subcommand | Description | CLI Support |
|------------|-------------|-------------|
| `install` | Install and configure Neovim with AI-recommended setup | ✅ Full Support |
| `plugins` | Manage Neovim plugins with intelligent recommendations | ✅ Full Support |
| `config` | Generate optimized Neovim configuration files | ✅ Full Support |
| `mcp` | Setup MCP integration for AI-powered assistance | ✅ Full Support |
| `update` | Update Neovim configuration and plugins | ✅ Full Support |

### Enhanced Flags (2 Total)

| Flag | Description | CLI Support |
|------|-------------|-------------|
| `--minimal` | Use minimal Neovim configuration for lightweight setup | ✅ Full Support |
| `--full` | Use full-featured Neovim setup with all integrations | ✅ Full Support |

## Command Help Output

```sh
./nixai neovim-setup --help
Set up and configure Neovim for NixOS development.

Usage:
  nixai neovim-setup [flags]

Flags:
  -h, --help   help for neovim-setup
  --minimal    Use a minimal Neovim configuration
  --full       Use a full-featured Neovim setup

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai neovim-setup
  nixai neovim-setup --minimal
```

---

## Real Life Examples

- **Set up a minimal Neovim config for NixOS:**
  ```sh
  nixai neovim-setup --minimal
  # Installs a basic Neovim config for development
  ```
- **Set up a full-featured Neovim config:**
  ```sh
  nixai neovim-setup --full
  # Installs a full-featured Neovim config with plugins
  ```
