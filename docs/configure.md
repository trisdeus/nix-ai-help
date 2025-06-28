# nixai configure

NixOS and Home Manager configuration with AI-guided setup wizard, preset templates, and intelligent recommendation system.

---

## 🆕 Enhanced Features

The `nixai configure` command features **comprehensive CLI integration** with advanced configuration capabilities:

### ✨ **CLI Features**
- **🎯 Parameter Input**: Complete configuration wizard through CLI interface
- **📊 Real-Time Configuration Preview**: Live configuration generation with validation
- **⌨️ Command Discovery**: Enhanced command options with all 8 configurable flags
- **🔧 Configuration Wizard**: Step-by-step setup with AI-guided recommendations
- **📋 Context-Aware Configuration**: Automatic detection of existing setup for seamless integration

### ⚙️ **Advanced Configuration Features**
- **🧠 AI-Guided Setup Wizard**: Intelligent configuration assistant with context-aware recommendations
- **📝 Preset Templates**: Desktop, server, minimal, development, and gaming configuration presets
- **🎯 Smart Hardware Detection**: Automatic optimization based on detected hardware and use cases
- **🔧 Modular Configuration**: Component-based configuration with dependency management
- **📊 Configuration Validation**: Real-time validation with error detection and fix suggestions
- **🔄 Version Management**: Configuration versioning with rollback support and change tracking
- **🎨 Desktop Environment Integration**: Automated setup for GNOME, KDE, XFCE, i3, and custom WMs
- **🛡️ Security Hardening**: Automated security configuration with compliance templates

## Command Structure

```sh
nixai configure [flags]
```

### Enhanced Flags (8 Total)

| Flag | Description | CLI Support |
|------|-------------|-------------|
| `--preset <type>` | Use configuration preset (desktop/server/minimal/dev/gaming) | ✅ Full Support |
| `--hardware` | Enable automatic hardware optimization | ✅ Full Support |
| `--desktop <env>` | Configure desktop environment (gnome/kde/xfce/i3/custom) | ✅ Full Support |
| `--services` | Service selection and configuration | ✅ Full Support |
| `--security` | Apply security hardening configurations | ✅ Full Support |
| `--file <path>` | Specify custom configuration file to use | ✅ Full Support |
| `--home` | Configure Home Manager instead of NixOS | ✅ Full Support |
| `--validate` | Validate configuration without applying changes | ✅ Full Support |

## Command Help Output

```sh
./nixai configure --help
Interactive or scripted configuration of NixOS or Home Manager.

Usage:
  nixai configure [flags]

Flags:
  -h, --help   help for configure
  --file      Specify a configuration file to use
  --home      Configure Home Manager instead of NixOS

Global Flags:
  -a, --ask string          Ask a question about NixOS configuration
  -n, --nixos-path string   Path to your NixOS configuration folder (containing flake.nix or configuration.nix)

Examples:
  nixai configure
  nixai configure --file myconfig.nix
  nixai configure --home
```

---

## Real Life Examples

- **Start interactive configuration for NixOS:**
  ```sh
  nixai configure
  # Walks you through configuration interactively
  ```
- **Configure Home Manager interactively:**
  ```sh
  nixai configure --home
  # Starts Home Manager configuration wizard
  ```
- **Use a specific configuration file:**
  ```sh
  nixai configure --file myconfig.nix
  # Loads and applies settings from myconfig.nix
  ```
