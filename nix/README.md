# nixai Nix Packaging

This directory contains all Nix-related packaging files for nixai, organized for clarity and maintainability.

## Directory Structure

```
nix/
├── package.nix              # Main package definition
├── standalone-install.nix   # Non-flake installation helper
├── modules/                 # NixOS and Home Manager modules
│   ├── nixos.nix           # NixOS system module
│   ├── home-manager.nix    # Home Manager user module
│   ├── nixai-nvim.lua      # Neovim integration script
│   └── README.md           # Module documentation
└── examples/               # Configuration examples and tests
    ├── flake.example.nix          # Example flake configuration
    ├── nixvim-nixai-example.nix   # Nixvim integration example
    ├── example-home-manager.nix   # Home Manager example
    ├── example-nixai-usage.nix    # General usage examples
    ├── example-user-flake.nix     # User flake example
    ├── test-flake-free-hm.nix     # Flake-free home-manager test
    ├── test-import.nix            # Import test
    └── test-nixai-import.nix      # nixai import test
```

## Quick Installation

### Flake-based (Recommended)

Add to your `flake.nix`:
```nix
inputs.nixai.url = "github:olafkfreund/nix-ai-help";
```

Then use the modules:
```nix
# For NixOS
nixai.nixosModules.default

# For Home Manager  
nixai.homeManagerModules.default
```

### Non-flake Installation

From this directory:
```bash
nix-build standalone-install.nix
```

## Package Definition

The main package is defined in `package.nix` with the following features:

- **Multi-system support**: x86_64-linux, aarch64-linux, x86_64-darwin, aarch64-darwin
- **Go module build**: Uses `buildGoModule` with proper vendor hash
- **Shell completions**: Automatically generated for bash, fish, and zsh
- **Version injection**: Build-time version, git commit, and build date
- **Flexible source**: Supports both local and remote sources

## Modules

### NixOS Module (`modules/nixos.nix`)
- System-wide nixai installation
- systemd service for MCP server
- Configurable AI providers and models
- Security hardening options

### Home Manager Module (`modules/home-manager.nix`)
- Per-user nixai installation
- User-level systemd service
- VS Code and Neovim integrations
- Context-aware AI assistance

## Examples

The `examples/` directory contains:

- **Working configurations** for different use cases
- **Test configurations** for validation
- **Integration examples** with popular tools
- **Migration helpers** for different installation methods

## Development

To modify the packaging:

1. **Edit `package.nix`** for build changes
2. **Update modules/** for NixOS/Home Manager integration
3. **Test with examples/** configurations
4. **Update documentation** as needed

## Path Updates

After the reorganization, all references have been updated:

- **Flake references**: `./nix/package.nix`, `./nix/modules/nixos.nix`
- **Module cross-references**: Use relative paths within `nix/`
- **Example imports**: Point to `nix/modules/` and `nix/package.nix`

## Migration Notes

This structure change from the old organization:

### Before (Scattered)
```
root/
├── package.nix
├── modules/
├── example-*.nix
└── standalone-install.nix
```

### After (Organized)  
```
root/
├── flake.nix
├── nix/
│   ├── package.nix
│   ├── modules/
│   └── examples/
└── [Go source code]
```

### Breaking Changes

- **Manual imports**: Update paths from `./modules/` to `./nix/modules/`
- **Non-flake usage**: Update references to `./package.nix` → `./nix/package.nix`
- **Example usage**: All examples moved to `nix/examples/`

### Migration Help

For existing configurations, update import paths:

```nix
# Old
imports = [ ./path/to/nixai/modules/nixos.nix ];
callPackage ./path/to/nixai/package.nix {};

# New  
imports = [ ./path/to/nixai/nix/modules/nixos.nix ];
callPackage ./path/to/nixai/nix/package.nix {};
```

## Testing

All packaging components can be tested:

```bash
# Test flake
nix flake check

# Test build
nix build

# Test modules (requires NixOS/Home Manager)
nixos-rebuild dry-build --flake .
home-manager build --flake .

# Test standalone
cd nix && nix-build standalone-install.nix
```

---

For more information, see the main [FLAKE_INTEGRATION_GUIDE.md](../docs/FLAKE_INTEGRATION_GUIDE.md) and [modules/README.md](modules/README.md).
