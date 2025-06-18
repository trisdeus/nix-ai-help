# Nix Files Reorganization Plan

## Overview

User feedback suggests reorganizing Nix-related packaging files into a dedicated `nix/` folder to improve project organization and separate packaging concerns from the core Go application.

## Current Structure (Scattered)

```
root/
├── flake.nix                    # Main flake (stays at root)
├── package.nix                  # → nix/package.nix
├── modules/                     # → nix/modules/
│   ├── nixos.nix
│   ├── home-manager.nix
│   ├── flake.example.nix
│   ├── nixvim-nixai-example.nix
│   ├── nixai-nvim.lua
│   └── README.md
├── example-home-manager.nix     # → nix/examples/
├── example-nixai-usage.nix      # → nix/examples/
├── example-user-flake.nix       # → nix/examples/
├── standalone-install.nix       # → nix/
├── test-flake-free-hm.nix       # → nix/examples/
└── [Go source code remains at root]
```

## Proposed New Structure

```
root/
├── flake.nix                    # Stays at root (standard)
├── nix/                         # NEW: All Nix packaging
│   ├── package.nix              # Main package definition
│   ├── standalone-install.nix   # Standalone installation
│   ├── modules/                 # NixOS/Home Manager modules
│   │   ├── nixos.nix
│   │   ├── home-manager.nix
│   │   ├── nixai-nvim.lua
│   │   └── README.md
│   └── examples/                # Configuration examples
│       ├── flake.example.nix
│       ├── nixvim-nixai-example.nix
│       ├── home-manager.nix
│       ├── nixai-usage.nix
│       ├── user-flake.nix
│       └── test-flake-free-hm.nix
└── [Go source code at root]
```

## Benefits

1. **Clear Separation**: Nix packaging vs Go application code
2. **Better Organization**: All Nix files in one logical location
3. **Easier Navigation**: Developers can quickly find Nix-related files
4. **Standard Practice**: Many projects organize packaging this way
5. **Cleaner Root**: Less clutter in the main directory

## Required Changes

### 1. File Moves
- Move `package.nix` → `nix/package.nix`
- Move `modules/` → `nix/modules/`
- Move examples → `nix/examples/`
- Move `standalone-install.nix` → `nix/`

### 2. Update References in `flake.nix`
```nix
# Before
packages.nixai = pkgs.callPackage ./package.nix { ... };
nixosModules.default = import ./modules/nixos.nix;
homeManagerModules.default = import ./modules/home-manager.nix;

# After  
packages.nixai = pkgs.callPackage ./nix/package.nix { ... };
nixosModules.default = import ./nix/modules/nixos.nix;
homeManagerModules.default = import ./nix/modules/home-manager.nix;
```

### 3. Update Module Cross-References
- Update paths in modules that reference `../package.nix`
- Update example imports and paths
- Update documentation references

### 4. Update Documentation
- Update all installation guides
- Update README.md with new paths
- Update example configurations

## Implementation Steps

1. **Phase 1**: Create `nix/` structure and move files
2. **Phase 2**: Update `flake.nix` and module references  
3. **Phase 3**: Update documentation and examples
4. **Phase 4**: Test all installation methods
5. **Phase 5**: Update CI/CD if needed

## Compatibility

- This change is **breaking** for manual imports (non-flake usage)
- Flake-based usage remains the same (paths are internal)
- Need to update documentation for manual installation methods

## User Impact

- **Flake users**: No impact (recommended approach)
- **Manual users**: Need to update import paths
- **Documentation**: All examples need path updates

---

This reorganization addresses the user feedback and aligns with best practices for polyglot projects with multiple packaging systems.
