# nixai deps - Dependency Analysis

The `nixai deps` command provides comprehensive dependency analysis for NixOS configurations, helping identify issues, optimize dependencies, and understand configuration relationships.

## Usage

```bash
nixai deps [command] [options]
```

## Commands

```bash
# Analyze configuration dependencies
nixai deps analyze [--config-file path] [--output json|tree|graph]

# Show dependency tree
nixai deps tree [--depth 5] [--package package-name]

# Find dependency conflicts
nixai deps conflicts [--resolve] [--suggest-fixes]

# Show reverse dependencies
nixai deps reverse <package-name> [--include-system]

# Dependency graph visualization
nixai deps graph [--output deps.png] [--format svg|png|dot]

# Optimize dependencies
nixai deps optimize [--remove-unused] [--suggest-alternatives]
```

## Features

### Configuration Analysis
- Parse NixOS configuration files
- Identify direct and indirect dependencies
- Detect circular dependencies
- Show import chains and relationships

### Conflict Resolution
- Detect version conflicts
- Identify incompatible packages
- Suggest resolution strategies
- Generate conflict-free configurations

### Optimization
- Find unused dependencies
- Suggest lighter alternatives
- Identify redundant packages
- Optimize build dependencies

## Examples

```bash
# Basic dependency analysis
nixai deps analyze

# Show package dependency tree
nixai deps tree --package firefox --depth 3

# Find and resolve conflicts
nixai deps conflicts --resolve --suggest-fixes

# Generate dependency graph
nixai deps graph --output system-deps.svg --format svg
```

For detailed usage examples and troubleshooting, see the main documentation.