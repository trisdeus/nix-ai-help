# NixAI Issues Fix Summary

## Issues Resolved

### 1. ❌ `unknown flag: --socket-path` Error
**Problem**: The home-manager module was trying to use `--socket-path` flag with `nixai mcp-server start`, but this flag wasn't implemented in the CLI.

**Solution**: ✅ Added support for `--socket-path` flag to the `mcp-server` command.

**Changes Made**:
- Added `socketPath` variable and flag to the CLI commands
- Updated `handleMCPServerStart` function to accept and use custom socket path
- Updated `handleMCPServerRestart` function to pass socket path parameter
- Fixed daemon mode to properly pass the socket path to the background process

**Test Results**:
```bash
# Command now works correctly:
nixai mcp-server start --socket-path=$HOME/.local/share/nixai/mcp.sock --daemon
# ✅ MCP server started in daemon mode
# Process ID: 3745689
# HTTP Server: http://localhost:8081
# Unix Socket: /home/olafkfreund/.local/share/nixai/mcp.sock
```

### 2. ❌ Fixed-output derivation error with fetchFromGitHub
**Problem**: The home-manager module had a circular dependency in the package definition that caused `fixed-output derivations must not reference store paths` error.

**Solution**: ✅ Updated the home-manager module to provide clear error message with proper usage examples instead of trying to automatically resolve the package.

**Changes Made**:
- Replaced the problematic automatic package resolution with an informative error message
- Provided three clear approaches for users to provide the nixai package explicitly
- Created example configuration files showing proper usage

## How to Use nixai Now

### Option 1: Home Manager with fetchFromGitHub (Recommended for non-flake users)

```nix
{ config, pkgs, ... }: let
  nixai-src = pkgs.fetchFromGitHub {
    owner = "olafkfreund";
    repo = "nix-ai-help";
    rev = "main"; # or specific commit
    sha256 = ""; # Leave empty for first build
  };
  nixai-package = pkgs.callPackage (nixai-src + "/package.nix") {};
in {
  services.nixai = {
    enable = true;
    mcp = {
      enable = true;
      package = nixai-package;
      socketPath = "${config.home.homeDirectory}/.local/share/nixai/mcp.sock";
    };
  };
}
```

### Option 2: Flake-based approach (Recommended for flake users)

In your `flake.nix`:
```nix
inputs.nixai.url = "github:olafkfreund/nix-ai-help";
```

In your home-manager configuration:
```nix
services.nixai = {
  enable = true;
  mcp = {
    enable = true;
    package = inputs.nixai.packages.${pkgs.system}.nixai;
  };
};
```

### Option 3: Direct system installation

```nix
{ config, pkgs, ... }: let
  nixai = pkgs.callPackage (pkgs.fetchFromGitHub {
    owner = "olafkfreund";
    repo = "nix-ai-help";
    rev = "main";
    sha256 = ""; # Leave empty for first build
  } + "/package.nix") {};
in {
  environment.systemPackages = [ nixai ];
}
```

## New Features Added

### ✨ Custom Socket Path Support
The `mcp-server` command now supports specifying custom socket paths:

```bash
# Start with custom socket path
nixai mcp-server start --socket-path=/custom/path/mcp.sock

# Start in daemon mode with custom socket
nixai mcp-server start -d --socket-path=$HOME/.local/share/nixai/mcp.sock

# Restart with custom socket (maintains the path)
nixai mcp-server restart --socket-path=/custom/path/mcp.sock
```

### 📋 Available Commands
```bash
nixai mcp-server start                               # Start with default socket
nixai mcp-server start -d                            # Start in daemon mode  
nixai mcp-server start --socket-path=/custom/path    # Start with custom socket
nixai mcp-server stop                                # Stop the server
nixai mcp-server status                              # Check server status
nixai mcp-server restart                             # Restart the server
nixai mcp-server query "services.nginx.enable"      # Query documentation
```

## What's Fixed

1. ✅ **CLI Flag Support**: `--socket-path` flag now works correctly
2. ✅ **Home Manager Module**: No more circular dependency errors
3. ✅ **Daemon Mode**: Properly passes socket path to background process
4. ✅ **Socket Path Persistence**: Restart command maintains custom socket paths
5. ✅ **Clear Error Messages**: Informative guidance when package isn't provided
6. ✅ **Multiple Installation Methods**: Support for flake and non-flake approaches

## Next Steps

1. **Update your configuration** using one of the provided examples
2. **Run `nixos-rebuild switch`** or `home-manager switch` to apply changes  
3. **Test the MCP server** with `nixai mcp-server start --socket-path=$HOME/.local/share/nixai/mcp.sock`
4. **Check VS Code integration** works with the configured socket path

## Files Changed

- `internal/cli/commands.go`: Added `--socket-path` flag support
- `modules/home-manager.nix`: Fixed package dependency issue
- Created example configuration files for different use cases

The nixai tool should now work correctly with both manual builds and home-manager integration! 🎉
