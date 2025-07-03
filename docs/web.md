# nixai web - Modern Web Interface

The `nixai web` command provides a modern, responsive web interface for managing NixOS configurations through an intuitive dashboard.

## Overview

The web interface offers a comprehensive graphical alternative to the command-line interface, featuring real-time monitoring, configuration building, team collaboration, and system management capabilities.

## Usage

```bash
nixai web start [options]
```

## Options

- `--port <port>` - Server port (default: 8080)
- `--host <host>` - Server host (default: localhost)
- `--no-browser` - Don't automatically open browser
- `--dev` - Enable development mode with live reload
- `--auth` - Enable authentication (default: disabled for local access)
- `--config <file>` - Custom configuration file
- `--verbose` - Enable verbose logging

## Features

### 🏠 Dashboard Interface

The web interface provides a comprehensive dashboard with multiple sections:

#### System Overview
- Real-time system status and health indicators
- NixOS version and configuration type (flakes/channels)
- Active services and running processes
- Resource usage (CPU, memory, disk, network)
- Recent system changes and updates

#### Configuration Builder
- Visual configuration editor with syntax highlighting
- Template library for common configurations
- Real-time validation and error checking
- Preview mode for testing configurations
- Integration with existing system configurations

#### Package Management
- Visual package browser with search and filtering
- Package dependency visualization
- Installation/removal with impact analysis
- Update management with rollback capabilities
- Custom package repository integration

### 🔧 Management Features

#### Service Management
- Start/stop/restart system services
- Service status monitoring and logs
- Service dependency visualization
- Configuration editing for systemd services

#### Hardware Information
- Comprehensive hardware detection and display
- Driver status and recommendations
- Performance metrics and optimization suggestions
- Hardware compatibility checking

#### Build System
- Build status monitoring and progress tracking
- Build log viewing with syntax highlighting
- Build cache management and optimization
- CI/CD integration for automated builds

### 👥 Collaboration Features

#### Team Management
- Multi-user configuration management
- Role-based access control
- Configuration approval workflows
- Change tracking and history

#### Configuration Sharing
- Export/import configuration profiles
- Community configuration marketplace
- Version control integration
- Configuration templates and snippets

### 📊 Monitoring & Analytics

#### System Metrics
- Real-time performance monitoring
- Historical data visualization
- Custom metric dashboards
- Alert configuration and notifications

#### Configuration Analytics
- Configuration complexity analysis
- Security assessment and recommendations
- Performance impact analysis
- Best practices compliance checking

## Getting Started

### Basic Usage

1. **Start the web server:**
```bash
nixai web start
```

2. **Access the interface:**
   - Opens automatically in default browser
   - Navigate to `http://localhost:8080`
   - Use `--port` to specify different port

3. **Explore the dashboard:**
   - System overview shows current status
   - Use navigation menu to access different sections
   - Configuration builder for editing system config

### Development Mode

Enable development features for testing and customization:

```bash
nixai web start --dev --verbose
```

Development mode includes:
- Live reload for interface updates
- Debug information and logging
- Development API endpoints
- Hot module replacement

### Custom Configuration

Use a custom configuration file for advanced settings:

```bash
nixai web start --config /path/to/web-config.yaml
```

Example web configuration:
```yaml
web:
  server:
    port: 8080
    host: "0.0.0.0"
    cors_enabled: true
    max_upload_size: "10MB"
  
  features:
    authentication: false
    team_features: true
    real_time_monitoring: true
    configuration_builder: true
  
  security:
    allowed_origins: ["http://localhost:3000"]
    api_rate_limit: 100
    session_timeout: "24h"
  
  integrations:
    git_enabled: true
    docker_enabled: false
    ci_cd_webhooks: true
```

## API Integration

The web interface provides REST and WebSocket APIs for integration:

### REST API Endpoints

- `GET /api/system/status` - System health and information
- `GET /api/packages` - Installed packages list
- `POST /api/packages/install` - Install packages
- `GET /api/services` - System services status
- `POST /api/config/validate` - Validate configurations
- `GET /api/builds` - Build history and status

### WebSocket Events

- Real-time system metrics updates
- Build progress notifications
- Service status changes
- Configuration change events

### Example API Usage

```bash
# Get system status
curl http://localhost:8080/api/system/status

# Install package via API
curl -X POST http://localhost:8080/api/packages/install \
  -H "Content-Type: application/json" \
  -d '{"packages": ["firefox", "git"]}'

# Validate configuration
curl -X POST http://localhost:8080/api/config/validate \
  -H "Content-Type: application/json" \
  -d '{"config": "services.nginx.enable = true;"}'
```

## Security Considerations

### Local Development
- Default configuration allows local access only
- No authentication required for localhost
- CORS protection enabled by default

### Production Deployment
- Enable authentication for remote access
- Configure HTTPS with SSL certificates
- Set up proper firewall rules
- Use reverse proxy for additional security

### Access Control
- Role-based permissions system
- API key authentication for external access
- Session management and timeout controls
- Audit logging for all configuration changes

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
nixai web start --port 8081
```

**Browser doesn't open automatically:**
```bash
nixai web start --no-browser
# Then manually open http://localhost:8080
```

**Permission errors:**
- Ensure user has access to NixOS configuration files
- Run with appropriate permissions for system changes
- Check firewall settings for network access

**API connection issues:**
- Verify server is running with `ps aux | grep nixai`
- Check logs with `--verbose` flag
- Test connectivity with `curl http://localhost:8080/api/system/status`

### Debug Mode

Enable comprehensive logging for troubleshooting:

```bash
nixai web start --dev --verbose --log-level debug
```

### Performance Optimization

For better performance with large configurations:

```bash
nixai web start --port 8080 --config high-performance.yaml
```

Example high-performance configuration:
```yaml
web:
  performance:
    cache_enabled: true
    compression: true
    static_file_caching: true
    database_pool_size: 20
    worker_threads: 4
```

## Integration Examples

### CI/CD Integration

Integrate with GitLab CI/CD for automated deployments:

```yaml
# .gitlab-ci.yml
deploy_nixos:
  stage: deploy
  script:
    - curl -X POST $NIXAI_WEB_URL/api/deploy \
        -H "Authorization: Bearer $API_TOKEN" \
        -d '{"branch": "$CI_COMMIT_REF_NAME"}'
```

### Monitoring Integration

Connect with Prometheus for metrics collection:

```bash
# Start web interface with metrics endpoint
nixai web start --enable-metrics --metrics-port 9090
```

### External Tool Integration

Use the web API for custom tooling:

```python
import requests

# Get system information
response = requests.get('http://localhost:8080/api/system/status')
system_info = response.json()

# Install packages programmatically
packages = ['neovim', 'firefox', 'git']
requests.post('http://localhost:8080/api/packages/install', 
              json={'packages': packages})
```

## Best Practices

1. **Regular Backups**: Use the web interface to create configuration snapshots
2. **Testing Changes**: Always test configurations in preview mode first
3. **Access Control**: Enable authentication for any remote access
4. **Monitoring**: Set up alerts for system health and configuration changes
5. **Documentation**: Use the built-in documentation features for team configurations

For more advanced usage and integration examples, see the [Web Interface Guide](../examples/web-interface/) and [API Documentation](api.md).