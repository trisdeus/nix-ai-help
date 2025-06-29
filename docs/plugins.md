# nixai Plugin System

The nixai Plugin System provides a comprehensive framework for extending nixai functionality through dynamically loadable plugins. This system supports secure execution, package management, community marketplace, and developer tools.

## Features

### 🔌 Dynamic Loading
- Hot-loading of plugins without restart
- Runtime plugin discovery and loading
- Plugin lifecycle management (load, start, stop, unload)
- Configuration-driven plugin management

### 🛡️ Security Sandbox
- Isolated execution environment for plugins
- Configurable security policies
- Resource limits (memory, CPU, network)
- Permission-based access control
- Capability-based security model

### 📦 Package Manager
- Plugin distribution and update system
- Multiple repository support (official, community, local)
- Automatic dependency resolution
- Signature verification and integrity checks
- Version management and updates

### 🤝 Community Marketplace
- Community-contributed plugin ecosystem
- Plugin discovery and search
- Ratings and reviews system
- Featured and popular plugins
- Plugin submission and management

### ⚡ Native Performance
- High-performance plugin execution
- Efficient resource utilization
- Optimized loading and communication
- Event-driven architecture

## Architecture

### Core Components

1. **Plugin Manager** (`manager.go`)
   - Central coordinator for plugin lifecycle
   - Plugin loading, starting, stopping, and unloading
   - Configuration management
   - Health monitoring and metrics

2. **Plugin Registry** (`registry.go`)
   - Plugin discovery and registration
   - Metadata management
   - Capability indexing
   - Search and filtering

3. **Plugin Loader** (`loader.go`)
   - Dynamic plugin loading from various sources
   - Plugin validation and verification
   - Go plugin support and extension points
   - Dependency resolution

4. **Security Sandbox** (`sandbox.go`)
   - Secure execution environment
   - Resource monitoring and limits
   - Permission checking and enforcement
   - Security policy management

5. **Event System** (`events.go`)
   - Event-driven plugin communication
   - Event bus with filtering and routing
   - Plugin event handling and publishing
   - Metrics and monitoring events

6. **Package Manager** (`package_manager.go`)
   - Plugin distribution and updates
   - Repository management
   - Download and installation
   - Version control and rollback

7. **Marketplace** (`marketplace.go`)
   - Community plugin ecosystem
   - Search and discovery
   - Ratings and reviews
   - Plugin statistics and analytics

8. **Template System** (`templates.go`)
   - Plugin scaffolding and creation
   - Built-in templates for different plugin types
   - Code generation and project setup
   - Developer tools and utilities

## Configuration

### Plugin System Configuration

```yaml
plugins:
  enabled: true
  directory: ~/.config/nixai/plugins
  cache_directory: ~/.cache/nixai/plugins
  config_directory: ~/.config/nixai/plugins/config
  auto_discover: true
  auto_update: false
  sandbox_enabled: true
  max_concurrent: 5
  timeout: 30
  
  repositories:
    - name: "official"
      url: "https://plugins.nixai.dev"
      type: "official"
      enabled: true
      priority: 1
      verified: true
    - name: "community"
      url: "https://community.nixai.dev/plugins"
      type: "community"
      enabled: true
      priority: 2
      verified: false
  
  marketplace:
    enabled: true
    base_url: "https://marketplace.nixai.dev"
    cache_duration: 3600
    featured_plugins: 10
    search_timeout: 15
  
  security:
    sandbox_enabled: true
    allow_network: false
    allow_filesystem_write: false
    allow_system_calls: false
    max_memory_mb: 512
    max_cpu_percent: 50
    allowed_domains: []
    blocked_capabilities: ["CAP_SYS_ADMIN", "CAP_NET_ADMIN"]
  
  package_manager:
    verify_signatures: true
    allow_unsigned: false
    update_check_interval: 86400
    download_timeout: 300
    max_download_size_mb: 100
```

## CLI Commands

### Plugin Management

```bash
# List installed plugins
nixai plugin list

# Search for plugins
nixai plugin search <query>

# Install a plugin
nixai plugin install <plugin-name>

# Uninstall a plugin
nixai plugin uninstall <plugin-name>

# Enable/disable plugins
nixai plugin enable <plugin-name>
nixai plugin disable <plugin-name>

# Get plugin information
nixai plugin info <plugin-name>

# Check plugin status
nixai plugin status <plugin-name>

# Execute plugin operations
nixai plugin execute <plugin-name> <operation> [args...]

# Discover available plugins
nixai plugin discover

# Validate plugin
nixai plugin validate <plugin-path>

# View plugin metrics
nixai plugin metrics [plugin-name]

# View plugin events
nixai plugin events [--filter=<pattern>]

# Create new plugin from template
nixai plugin create <template-name> <output-directory>
```

### Repository Management

```bash
# List repositories
nixai plugin repo list

# Add repository
nixai plugin repo add <name> <url> [--type=<type>]

# Remove repository
nixai plugin repo remove <name>

# Update repositories
nixai plugin repo update

# Search in repositories
nixai plugin repo search <query>
```

## Plugin Development

### Plugin Interface

All plugins must implement the `PluginInterface`:

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

### Creating a Plugin

1. **Use Template System**:
   ```bash
   nixai plugin create basic-go my-plugin
   cd my-plugin
   ```

2. **Implement Plugin Interface**:
   ```go
   package main

   import (
       "context"
       "nix-ai-help/internal/plugins"
   )

   type MyPlugin struct {
       config plugins.PluginConfig
       running bool
   }

   func (p *MyPlugin) Name() string { return "my-plugin" }
   func (p *MyPlugin) Version() string { return "1.0.0" }
   // ... implement other methods
   ```

3. **Build Plugin**:
   ```bash
   go build -buildmode=plugin -o my-plugin.so main.go
   ```

4. **Install Plugin**:
   ```bash
   nixai plugin install ./my-plugin.so
   ```

### Plugin Templates

Available templates:
- `basic-go`: Basic Go plugin template
- `advanced-go`: Advanced Go plugin with full feature set
- `nixos-integration`: NixOS-specific plugin template
- `ai-provider`: AI provider integration plugin
- `tool-integration`: External tool integration plugin

## Security Model

### Sandbox Execution

Plugins run in a secure sandbox environment with:
- **Process Isolation**: Separate process space
- **Resource Limits**: CPU, memory, and time constraints
- **Network Restrictions**: Configurable network access
- **File System Controls**: Limited file system access
- **Capability Restrictions**: Linux capability filtering

### Permission System

Plugins request specific permissions:
- `network`: Network access
- `filesystem.read`: File system read access
- `filesystem.write`: File system write access
- `system.exec`: Execute system commands
- `ai.provider`: Access to AI providers
- `nixos.config`: Access to NixOS configuration

### Security Policies

Configure security policies per plugin:
```yaml
plugins:
  my-plugin:
    security_policy:
      allow_network: true
      allowed_domains: ["api.github.com", "nixos.org"]
      max_memory_mb: 256
      max_cpu_percent: 25
      permissions: ["network", "filesystem.read"]
```

## Event System

### Event Types

- `plugin.loaded`: Plugin loaded successfully
- `plugin.started`: Plugin started
- `plugin.stopped`: Plugin stopped
- `plugin.error`: Plugin error occurred
- `plugin.operation`: Plugin operation executed
- `plugin.health`: Plugin health status change

### Event Handling

```go
// Subscribe to events
eventBus.Subscribe(func(event plugins.PluginEvent) {
    if event.Type == "plugin.error" {
        log.Printf("Plugin error: %s", event.Data["error"])
    }
})

// Publish events
eventBus.Publish(plugins.PluginEvent{
    Type: "plugin.operation",
    PluginName: "my-plugin",
    Data: map[string]interface{}{
        "operation": "backup",
        "status": "completed",
    },
})
```

## Monitoring and Metrics

### Health Monitoring

- Plugin health checks
- Resource usage monitoring
- Performance metrics
- Error tracking and reporting

### Metrics Collection

- Operation execution times
- Resource consumption
- Success/failure rates
- Event statistics

### Monitoring Integration

```bash
# View plugin metrics
nixai plugin metrics my-plugin

# Export metrics for monitoring systems
nixai plugin metrics --export=prometheus

# Real-time monitoring
nixai plugin monitor --watch
```

## Best Practices

### Plugin Development

1. **Error Handling**: Implement robust error handling
2. **Resource Management**: Clean up resources properly
3. **Configuration**: Support flexible configuration
4. **Testing**: Write comprehensive tests
5. **Documentation**: Provide clear documentation
6. **Security**: Follow security best practices

### Performance

1. **Lazy Loading**: Load resources only when needed
2. **Caching**: Cache expensive operations
3. **Async Operations**: Use async for long-running tasks
4. **Resource Limits**: Respect resource constraints
5. **Clean Shutdown**: Implement proper cleanup

### Security

1. **Minimal Permissions**: Request only necessary permissions
2. **Input Validation**: Validate all inputs
3. **Output Sanitization**: Sanitize outputs
4. **Secure Communication**: Use secure protocols
5. **Audit Logging**: Log security-relevant events

## Troubleshooting

### Common Issues

1. **Plugin Load Failures**:
   - Check plugin compatibility
   - Verify dependencies
   - Check file permissions

2. **Security Violations**:
   - Review security policies
   - Check permission requirements
   - Verify allowed domains/capabilities

3. **Performance Issues**:
   - Monitor resource usage
   - Check for memory leaks
   - Profile plugin operations

4. **Communication Errors**:
   - Verify event bus connectivity
   - Check network configurations
   - Monitor event delivery

### Debug Mode

Enable debug logging:
```bash
nixai --log-level debug plugin status my-plugin
```

### Plugin Diagnostics

```bash
# Check plugin health
nixai plugin health my-plugin

# Validate plugin configuration
nixai plugin validate-config my-plugin

# Test plugin operations
nixai plugin test my-plugin operation-name
```

## Migration Guide

### From Manual Integration

1. **Extract Functionality**: Identify reusable components
2. **Create Plugin**: Use template system
3. **Implement Interface**: Follow plugin interface
4. **Test Integration**: Validate functionality
5. **Deploy**: Install and configure plugin

### Plugin Updates

1. **Version Compatibility**: Check version requirements
2. **Configuration Migration**: Update configuration files
3. **Data Migration**: Migrate plugin data if needed
4. **Testing**: Validate updated functionality
5. **Rollback Plan**: Prepare rollback strategy

## Contributing

### Plugin Contributions

1. **Development**: Follow development guidelines
2. **Testing**: Ensure comprehensive test coverage
3. **Documentation**: Provide complete documentation
4. **Review**: Submit for community review
5. **Publication**: Publish to marketplace

### Core System Contributions

1. **Issues**: Report bugs and feature requests
2. **Pull Requests**: Submit improvements
3. **Testing**: Contribute test cases
4. **Documentation**: Improve documentation
5. **Community**: Help other developers

## Roadmap

### Phase 1 (Current)
- ✅ Core plugin system implementation
- ✅ Security sandbox
- ✅ Package manager
- ✅ CLI integration
- ✅ Template system

### Phase 2 (Next)
- 🔄 Web UI integration
- 🔄 Advanced security features
- 🔄 Plugin signing and verification
- 🔄 Performance optimizations
- 🔄 Marketplace launch

### Phase 3 (Future)
- 📋 Plugin composition and workflows
- 📋 Advanced monitoring and analytics
- 📋 Cross-platform support
- 📋 Enterprise features
- 📋 Plugin ecosystem expansion

## Support

- **Documentation**: [docs/plugins/](./plugins/)
- **Examples**: [examples/plugins/](../examples/plugins/)
- **Community**: [GitHub Discussions](https://github.com/nixai/discussions)
- **Issues**: [GitHub Issues](https://github.com/nixai/issues)
- **Discord**: [nixai Community](https://discord.gg/nixai)
