# Plugin Development Guide

This guide explains how to develop plugins for nixai, the AI-powered NixOS assistant.

## Overview

nixai supports a plugin system that allows extending its functionality with custom commands and capabilities. Plugins are implemented as Go shared libraries that conform to the PluginInterface.

## Plugin Interface

All plugins must implement the `PluginInterface` defined in `internal/plugins/api.go`:

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

## Plugin Structure

A basic plugin consists of:

1. A Go package with a `NewPlugin` function that returns a `PluginInterface`
2. Implementation of all required methods in the `PluginInterface`
3. Compilation as a Go plugin with `go build -buildmode=plugin`

## Creating a New Plugin

### 1. Use the Plugin Template

nixai provides a plugin template to help you get started:

```bash
# Create a new plugin from template
nixai plugin create --template basic my-new-plugin

# Navigate to the plugin directory
cd my-new-plugin

# Edit the plugin implementation
# Modify the plugin.go file to implement your functionality

# Build the plugin
go build -buildmode=plugin -o my-new-plugin.so .

# Install the plugin
nixai plugin install ./my-new-plugin.so
```

### 2. Manual Plugin Creation

You can also create a plugin manually:

1. Create a new directory for your plugin:
   ```bash
   mkdir my-plugin
   cd my-plugin
   ```

2. Create your main plugin file:
   ```go
   package main

   import (
       "context"
       "fmt"
       
       "nix-ai-help/internal/plugins"
   )

   // MyPlugin implements the PluginInterface
   type MyPlugin struct {
       name        string
       version     string
       description string
       author      string
       running     bool
   }

   // NewPlugin creates a new instance of the plugin
   func NewPlugin() plugins.PluginInterface {
       return &MyPlugin{
           name:        "my-plugin",
           version:     "1.0.0",
           description: "My custom plugin for nixai",
           author:      "Your Name",
           running:     false,
       }
   }

   // Implement all required methods...
   func (mp *MyPlugin) Name() string {
       return mp.name
   }

   func (mp *MyPlugin) Version() string {
       return mp.version
   }

   // ... (implement all other required methods)
   ```

3. Build your plugin:
   ```bash
   go build -buildmode=plugin -o my-plugin.so .
   ```

4. Install your plugin:
   ```bash
   nixai plugin install ./my-plugin.so
   ```

## Required Methods

### Metadata Methods

These methods provide information about the plugin:

- `Name()` - Returns the plugin name
- `Version()` - Returns the plugin version
- `Description()` - Returns a brief description
- `Author()` - Returns the plugin author
- `Repository()` - Returns the plugin repository URL
- `License()` - Returns the plugin license
- `Dependencies()` - Returns a list of dependencies
- `Capabilities()` - Returns a list of capabilities

### Lifecycle Methods

These methods manage the plugin lifecycle:

- `Initialize(ctx context.Context, config PluginConfig) error` - Initializes the plugin with configuration
- `Start(ctx context.Context) error` - Starts the plugin
- `Stop(ctx context.Context) error` - Stops the plugin
- `Cleanup(ctx context.Context) error` - Cleans up resources
- `IsRunning() bool` - Returns whether the plugin is currently running

### Execution Methods

These methods handle plugin execution:

- `Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)` - Executes an operation with parameters
- `GetOperations() []PluginOperation` - Returns available operations
- `GetSchema(operation string) (*PluginSchema, error)` - Returns schema for an operation

### Health and Status Methods

These methods provide health and status information:

- `HealthCheck(ctx context.Context) PluginHealth` - Performs a health check
- `GetMetrics() PluginMetrics` - Returns performance metrics
- `GetStatus() PluginStatus` - Returns current status

## Plugin Configuration

Plugins receive configuration through the `PluginConfig` struct:

```go
type PluginConfig struct {
    Name          string                 `json:"name" yaml:"name"`
    Enabled       bool                   `json:"enabled" yaml:"enabled"`
    Version       string                 `json:"version" yaml:"version"`
    Configuration map[string]interface{} `json:"configuration" yaml:"configuration"`
    Environment   map[string]string      `json:"environment" yaml:"environment"`
    Resources     ResourceLimits         `json:"resources" yaml:"resources"`
    SecurityPolicy SecurityPolicy        `json:"security_policy" yaml:"security_policy"`
    UpdatePolicy  UpdatePolicy           `json:"update_policy" yaml:"update_policy"`
}
```

## Plugin Operations

Plugins define operations through the `GetOperations()` method. Each operation includes:

- Name
- Description
- Parameters
- Return type
- Examples
- Tags

Example:
```go
func (mp *MyPlugin) GetOperations() []PluginOperation {
    return []PluginOperation{
        {
            Name:        "hello",
            Description: "Say hello to someone",
            Parameters: []PluginParameter{
                {
                    Name:        "name",
                    Description: "Name to say hello to",
                    Type:        "string",
                    Required:    false,
                },
            },
            ReturnType: "string",
            Examples: []OperationExample{
                {
                    Name:        "Say hello to Alice",
                    Description: "Example greeting",
                    Parameters: map[string]interface{}{
                        "name": "Alice",
                    },
                    Expected: "Hello, Alice!",
                },
            },
            Tags: []string{"greeting", "example"},
        },
    }
}
```

## Plugin Security

Plugins run in a sandboxed environment with restricted access:

1. Limited file system access
2. Controlled network access
3. Resource limits (memory, CPU, execution time)
4. Restricted system calls

Security policies can be configured through the `SecurityPolicy` struct:

```go
type SecurityPolicy struct {
    AllowFileSystem     bool         `json:"allow_file_system" yaml:"allow_file_system"`
    AllowNetwork        bool         `json:"allow_network" yaml:"allow_network"`
    AllowSystemCalls    bool         `json:"allow_system_calls" yaml:"allow_system_calls"`
    AllowedDomains      []string     `json:"allowed_domains" yaml:"allowed_domains"`
    RequiredPermissions []string     `json:"required_permissions" yaml:"required_permissions"`
    SandboxLevel        SandboxLevel `json:"sandbox_level" yaml:"sandbox_level"`
}
```

## Plugin Testing

Test your plugin with:

```bash
# Install the plugin
nixai plugin install ./my-plugin.so

# List plugins to verify installation
nixai plugin list

# Execute plugin operations
nixai plugin execute my-plugin hello --params '{"name": "world"}'

# Check plugin status
nixai plugin status my-plugin

# Uninstall when done
nixai plugin uninstall my-plugin
```

## Publishing Plugins

To share your plugin with the community:

1. Ensure your plugin follows best practices
2. Add comprehensive documentation and examples
3. Test thoroughly across different environments
4. Publish to the nixai plugin marketplace
5. Submit to the community repository

Example publishing command:
```bash
# Publish to the marketplace
nixai plugin publish ./my-plugin.so --name "my-plugin" --description "My awesome plugin"
```

## Best Practices

1. **Follow the Interface**: Implement all required methods correctly
2. **Error Handling**: Handle errors gracefully and provide meaningful error messages
3. **Context Awareness**: Respect context cancellation and timeouts
4. **Resource Management**: Clean up resources properly in Cleanup()
5. **Security**: Follow security best practices and respect sandboxing
6. **Documentation**: Provide clear documentation and examples
7. **Testing**: Include comprehensive tests for your plugin
8. **Versioning**: Use semantic versioning for your plugins
9. **Compatibility**: Ensure compatibility with different NixOS versions
10. **Performance**: Optimize for performance and memory usage

## Example Plugins

Refer to the following example plugins for guidance:

1. `examples/plugins/example_plugin.go` - A simple example plugin
2. `examples/plugins/system_info_plugin.go` - System information plugin
3. `examples/plugins/package_monitor_plugin.go` - Package monitoring plugin

These examples demonstrate various aspects of plugin development and can be used as templates for your own plugins.

## Troubleshooting

Common issues and solutions:

1. **Plugin Not Loading**: 
   - Check that the plugin exports a `NewPlugin` function
   - Ensure the plugin implements the `PluginInterface` correctly
   - Verify the plugin is compiled with `-buildmode=plugin`

2. **Method Not Found**:
   - Make sure all required methods are implemented
   - Check method signatures match the interface exactly

3. **Security Restrictions**:
   - Review the plugin's security policy
   - Adjust sandboxing level if necessary
   - Request additional permissions if needed

4. **Performance Issues**:
   - Profile plugin execution with metrics
   - Optimize resource usage and execution time
   - Implement proper cancellation handling

## Resources

- [Go Plugin Documentation](https://golang.org/pkg/plugin/)
- [nixai GitHub Repository](https://github.com/olafkfreund/nix-ai-help)
- [NixOS Manual](https://nixos.org/manual/nixos/stable/)
- [Nixpkgs Manual](https://nixos.org/manual/nixpkgs/stable/)

Happy plugin development!