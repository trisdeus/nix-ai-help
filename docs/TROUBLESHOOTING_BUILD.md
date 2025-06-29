# Troubleshooting Guide: nixai Build Issues

## Problem Description
You're encountering build errors where the Nix build system cannot find `go.mod` file, and you're getting version 1.0.2 instead of the current 1.0.7. The error shows a directory structure with only `bin/` and `share/` directories instead of the full Go source code.

## Root Cause Analysis
The issue occurs because:
1. The flake input is pointing to an old commit/release
2. The source being fetched doesn't include Go source files (possibly a binary release)
3. There might be a caching issue with Nix store

## Solutions

### Solution 1: Update Your Flake Lock
In your consuming flake directory, run:
```bash
nix flake update nixai  # Update just nixai input
nix flake update        # Or update all inputs
```

### Solution 2: Pin to Latest Commit
In your flake.nix, specify the exact commit:
```nix
inputs = {
  nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  nixai = {
    url = "github:olafkfreund/nix-ai-help/main";  # Use main branch
    inputs.nixpkgs.follows = "nixpkgs";
  };
};
```

### Solution 3: Use Path Reference (if local)
If you have the nixai repository locally:
```nix
nixai = {
  url = "path:/path/to/nixai";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

### Solution 4: Force Rebuild
Clear Nix caches and rebuild:
```bash
nix store gc  # Clean up store
nix flake update nixai
nixos-rebuild switch  # or home-manager switch
```

### Solution 5: Verify Package Version
Test the package directly:
```bash
nix shell github:olafkfreund/nix-ai-help#nixai
nixai --version  # Should show 1.0.7
```

## Correct Configuration Examples

### For NixOS configuration.nix:
```nix
{ inputs, config, lib, pkgs, ... }:
{
  imports = [ inputs.nixai.nixosModules.default ];
  
  services.nixai = {
    enable = true;
    mcp.enable = true;
  };
}
```

### For Home Manager:
```nix
{ inputs, config, lib, pkgs, ... }:
{
  imports = [ inputs.nixai.homeManagerModules.default ];
  
  services.nixai = {
    enable = true;  # Set to true to enable
    mcp = {
      enable = true;
      aiProvider = "copilot";
      aiModel = "gpt-4";
    };
  };
}
```

### Main flake.nix structure:
```nix
{
  description = "My system with nixai";
  
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    
    nixai = {
      url = "github:olafkfreund/nix-ai-help";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };
  
  outputs = { self, nixpkgs, home-manager, nixai, ... }: {
    nixosConfigurations.hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      specialArgs = { inherit inputs; };
      modules = [
        ./configuration.nix
        nixai.nixosModules.default
      ];
    };
  };
}
```

## Debugging Commands

Check what version is being pulled:
```bash
nix eval .#nixai.packages.x86_64-linux.nixai.version
```

Check flake inputs:
```bash
nix flake metadata nixai
```

Build directly from GitHub:
```bash
nix build github:olafkfreund/nix-ai-help#nixai
```

## Current Working Version
The current nixai flake provides:
- Package: `nixai-1.0.7`
- NixOS Module: `nixosModules.default`
- Home Manager Module: `homeManagerModules.default`
- Apps: `apps.x86_64-linux.nixai`

---

## See Also

- [AI Provider Configuration Issues](TROUBLESHOOTING_AI_PROVIDER_CONFIGURATION.md) - For "provider not configured" errors
- [Installation Guide](INSTALLATION.md) - Complete installation instructions
- [Configuration Guide](config.md) - Configuration setup and management
