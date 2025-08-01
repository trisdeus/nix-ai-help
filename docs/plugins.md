# NixAI Plugin System

The NixAI plugin system allows extending the functionality of nixai with custom commands and capabilities. This document explains how to use and develop plugins for nixai.

## Using Plugins

### Listing Available Plugins

To see all available plugins, use:

```bash
nixai plugin list
```

This will show both integrated plugins (built-in) and any external plugins that are installed.

### Installing Plugins

To install an external plugin:

```bash
nixai plugin install /path/to/plugin.so
```

### Managing Plugins

Enable or disable plugins:

```bash
nixai plugin enable plugin-name
nixai plugin disable plugin-name
```

Uninstall plugins:

```bash
nixai plugin uninstall plugin-name
```

## Integrated Plugins

NixAI comes with two integrated plugins:

### System Info (`system-info`)
Provides system information and health monitoring capabilities:
- `nixai system-info health`: Check system health
- `nixai system-info status`: Show system status
- `nixai system-info cpu`: CPU information
- `nixai system-info memory`: Memory information
- `nixai system-info disk`: Disk usage information
- `nixai system-info processes`: Running processes
- `nixai system-info monitor`: Monitor system in real-time
- `nixai system-info all`: All system information

### Package Monitor (`package-monitor`)
Monitors packages and manages updates:
- `nixai package-monitor list`: List installed packages
- `nixai package-monitor updates`: Check for package updates
- `nixai package-monitor security`: Security-related package information
- `nixai package-monitor analyze`: Analyze packages
- `nixai package-monitor stats`: Package statistics

## Developing Plugins

To develop a plugin for nixai, you need to:

1. Implement the `PluginInterface` defined in `internal/plugins/api.go`
2. Export a `NewPlugin()` function that returns an instance of your plugin
3. Compile your plugin as a shared library (`.so` file on Linux)

### Plugin Interface

Plugins must implement the following interface:

```go
type PluginInterface interface {
    // Metadata
    Name() string
    Version() string
    Description() string
    Author() string
    Repository() string
    License() string
    Dependencies() []string
    Capabilities() []string

    // Lifecycle
    Initialize(ctx context.Context, config PluginConfig) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Cleanup(ctx context.Context) error
    IsRunning() bool

    // Execution
    Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
    GetOperations() []PluginOperation
    GetSchema(operation string) (*PluginSchema, error)

    // Health and Status
    HealthCheck(ctx context.Context) PluginHealth
    GetMetrics() PluginMetrics
    GetStatus() PluginStatus
}
```

### Example Plugin

Here's a minimal example of a plugin:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "nix-ai-help/internal/plugins"
)

type ExamplePlugin struct {
    name        string
    version     string
    description string
    author      string
    running     bool
}

func NewPlugin() plugins.PluginInterface {
    return &ExamplePlugin{
        name:        "example-plugin",
        version:     "1.0.0",
        description: "An example plugin for nixai",
        author:      "Your Name",
    }
}

// Implement all required methods...

func (p *ExamplePlugin) Name() string {
    return p.name
}

func (p *ExamplePlugin) Version() string {
    return p.version
}

// ... (implement all other required methods)

func main() {
    // This function is required but not used for plugins
}
```

### Building Plugins

To build a plugin, use:

```bash
go build -buildmode=plugin -o example-plugin.so .
```

### Installing Plugins

Once built, install the plugin by copying it to the nixai plugins directory:

```bash
cp example-plugin.so ~/.nixai/plugins/
```

Then load it with:

```bash
nixai plugin install ~/.nixai/plugins/example-plugin.so
```

## Plugin Security

NixAI takes plugin security seriously:

1. All plugins must be validated before loading
2. Plugins run in a restricted environment by default
3. Network access can be controlled through security policies
4. File system access is limited to approved paths
5. Resource usage is monitored and limited

When installing third-party plugins, always:
- Verify the source is trustworthy
- Check the plugin code if possible
- Review the permissions the plugin requests
- Monitor the plugin's behavior after installation

## Plugin Directory Structure

NixAI looks for plugins in the following directories:

1. `~/.nixai/plugins/` (user plugins)
2. `/usr/local/share/nixai/plugins/` (system plugins)
3. `/usr/share/nixai/plugins/` (distribution plugins)

You can also specify custom plugin directories in the nixai configuration file.