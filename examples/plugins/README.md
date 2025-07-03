# NixAI Example Plugins

This directory contains comprehensive example plugins that demonstrate the power and flexibility of the NixAI plugin system. Each plugin showcases different aspects of system management and development automation with AI-powered intelligence.

## 🚀 Available Example Plugins

### 1. **System Monitor Plugin** 📊
**Path**: `system-monitor/`

Real-time system monitoring and alerting plugin with AI-powered insights.

**Key Features:**
- CPU, memory, disk, and temperature monitoring
- Configurable alert thresholds and notifications
- Health scoring system (0-100)
- Performance recommendations
- Resource usage trend analysis

**Use Cases:**
- Server monitoring and maintenance
- Performance bottleneck identification
- Proactive system health management
- Resource optimization

```bash
# Quick start
nixai plugin load system-monitor
nixai plugin execute system-monitor get-metrics
nixai plugin execute system-monitor health-check
```

---

### 2. **Package Updater Plugin** 📦
**Path**: `package-updater/`

Intelligent package update management with security analysis and automated scheduling.

**Key Features:**
- Smart update prioritization (security-first)
- Risk assessment and breaking change detection
- Automated update plans with AI recommendations
- Rollback support and backup integration
- Policy-based update management

**Use Cases:**
- Automated security patching
- Controlled system updates
- Dependency conflict resolution
- Change management workflows

```bash
# Quick start
nixai plugin load package-updater
nixai plugin execute package-updater check-updates
nixai plugin execute package-updater create-update-plan
```

---

### 3. **Service Manager Plugin** 🔧
**Path**: `service-manager/`

AI-powered systemd service management with intelligent health analysis and automated troubleshooting.

**Key Features:**
- Complete service lifecycle management
- Real-time health monitoring and scoring
- Dependency analysis and impact assessment
- Log integration and AI-powered diagnostics
- Automated issue resolution suggestions

**Use Cases:**
- Service orchestration and monitoring
- Automated service recovery
- Performance optimization
- Troubleshooting and diagnostics

```bash
# Quick start
nixai plugin load service-manager
nixai plugin execute service-manager list-services
nixai plugin execute service-manager analyze-service '{"service_name": "nginx"}'
```

---

### 4. **Development Environment Plugin** 🛠️
**Path**: `dev-environment/`

Intelligent development environment management with AI-powered project analysis and template generation.

**Key Features:**
- Multi-language environment support (Go, Python, Node.js, Rust, etc.)
- AI-powered project analysis and suggestions
- Template-based environment creation
- Automated dependency resolution
- Environment health monitoring

**Use Cases:**
- Rapid development setup
- Project onboarding automation
- Consistent development environments
- Cross-platform development support

```bash
# Quick start
nixai plugin load dev-environment
nixai plugin execute dev-environment analyze-project '{"path": "/my/project"}'
nixai plugin execute dev-environment create-environment '{"name": "my-app", "type": "go", "path": "/my/project"}'
```

## 🎯 Plugin Architecture Overview

Each example plugin demonstrates key architectural patterns:

### **Standard Plugin Interface**
All plugins implement the core `PluginInterface` with:
- **Metadata Methods**: Name, version, description, capabilities
- **Lifecycle Management**: Initialize, start, stop, cleanup
- **Operation Execution**: Execute commands with parameters
- **Health Monitoring**: Health checks and metrics collection

### **AI Integration Patterns**
- **Smart Analysis**: AI-powered data analysis and insights
- **Recommendation Engine**: Context-aware suggestions
- **Predictive Capabilities**: Trend analysis and forecasting
- **Natural Language Processing**: Human-readable outputs

### **Configuration Management**
- **YAML Configuration**: Structured plugin configuration
- **Schema Validation**: Parameter validation and type checking
- **Environment Variables**: Runtime configuration support
- **Policy-Based Settings**: Configurable behavior rules

### **Security & Sandboxing**
- **Resource Limits**: CPU, memory, and execution time constraints
- **File System Permissions**: Controlled file access
- **Network Restrictions**: Limited network access
- **Capability-Based Security**: Fine-grained permission model

## 📁 Plugin Structure

Each plugin follows a consistent structure:

```
plugin-name/
├── main.go           # Core plugin implementation
├── README.md         # Comprehensive documentation
├── plugin.yaml       # Plugin metadata and configuration
├── go.mod           # Go module dependencies
└── Makefile         # Build and installation scripts
```

### **Core Files Explained**

#### `main.go`
- Complete plugin implementation
- Demonstrates all plugin interface methods
- Shows best practices for error handling
- Includes comprehensive operation examples

#### `plugin.yaml`
- Plugin metadata and configuration schema
- Operation definitions with parameters
- Security and resource constraints
- Update policies and dependencies

#### `README.md`
- Detailed usage examples and documentation
- API reference with parameter descriptions
- Integration examples and best practices
- Troubleshooting guides

## 🔧 Development Workflow

### **Building Plugins**
```bash
# Build single plugin
cd system-monitor
make build

# Install to user directory
make install

# Create distributable package
make package
```

### **Testing Plugins**
```bash
# Load and test plugin
nixai plugin load system-monitor
nixai plugin status system-monitor
nixai plugin test system-monitor

# Execute operations
nixai plugin execute system-monitor get-metrics
nixai plugin execute system-monitor health-check
```

### **Plugin Management**
```bash
# List all plugins
nixai plugin list

# Get plugin information
nixai plugin info system-monitor

# Update plugin
nixai plugin update system-monitor

# Uninstall plugin
nixai plugin uninstall system-monitor
```

## 🎨 Customization Examples

### **Custom Alert Rules**
```yaml
# system-monitor custom configuration
monitoring:
  thresholds:
    cpu_warning: 60.0
    memory_critical: 90.0
  alerts:
    enable_notifications: true
    notification_levels: ["critical"]
```

### **Custom Update Policies**
```yaml
# package-updater custom configuration
update_policy:
  auto_update: true
  security_only: true
  excluded_packages: ["kernel", "systemd"]
  max_updates_per_run: 5
```

### **Custom Environment Templates**
```yaml
# dev-environment custom template
languages:
  go:
    default_version: "1.21"
    enable_modules: true
  python:
    default_version: "3.11"
    package_manager: "poetry"
```

## 🔗 Integration Examples

### **System Monitoring Dashboard**
```bash
#!/bin/bash
# Combined monitoring script using multiple plugins

echo "=== System Health Dashboard ==="

# System metrics
echo "📊 System Metrics:"
nixai plugin execute system-monitor get-metrics | jq '{cpu_usage, memory_usage, disk_usage}'

# Service health
echo "🔧 Service Health:"
nixai plugin execute service-manager health-check | jq '{overall_health, failed_services}'

# Package updates
echo "📦 Available Updates:"
nixai plugin execute package-updater check-updates | jq '{total_updates, security_updates}'

# Development environments
echo "🛠️ Development Environments:"
nixai plugin execute dev-environment list-environments | jq '.[] | {name, type, status}'
```

### **Automated Maintenance Script**
```bash
#!/bin/bash
# Automated system maintenance using plugins

# 1. Check system health
HEALTH=$(nixai plugin execute system-monitor health-check)
if [[ $(echo "$HEALTH" | jq -r '.overall_status') != "healthy" ]]; then
  echo "⚠️ System health issues detected"
fi

# 2. Restart failed services
FAILED_SERVICES=$(nixai plugin execute service-manager list-services '{"status": "failed"}')
echo "$FAILED_SERVICES" | jq -r '.[].name' | while read service; do
  nixai plugin execute service-manager restart-service "{\"service_name\": \"$service\"}"
done

# 3. Apply security updates
nixai plugin execute package-updater create-update-plan '{"include_breaking": false}' | \
  jq -r '.updates[] | select(.category == "security") | .name' | \
  xargs -I {} nixai plugin execute package-updater apply-updates "{\"packages\": [\"{}\"]}"

echo "✅ Automated maintenance completed"
```

### **Development Workflow Automation**
```bash
#!/bin/bash
# Automated development environment setup

PROJECT_PATH="$1"
PROJECT_NAME=$(basename "$PROJECT_PATH")

# Analyze project
echo "🔍 Analyzing project structure..."
ANALYSIS=$(nixai plugin execute dev-environment analyze-project "{\"path\": \"$PROJECT_PATH\"}")

# Get AI suggestions
echo "🤖 Getting AI recommendations..."
SUGGESTIONS=$(nixai plugin execute dev-environment suggest-environment "{\"path\": \"$PROJECT_PATH\"}")

# Create environment from best suggestion
TEMPLATE_ID=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].template_id')
ENV_RESULT=$(nixai plugin execute dev-environment create-from-template "{
  \"template_id\": \"$TEMPLATE_ID\",
  \"name\": \"$PROJECT_NAME\",
  \"path\": \"$PROJECT_PATH\"
}")

ENV_ID=$(echo "$ENV_RESULT" | jq -r '.id')

# Activate environment
echo "🚀 Activating development environment..."
nixai plugin execute dev-environment activate-environment "{\"environment_id\": \"$ENV_ID\"}"

echo "✅ Development environment ready for $PROJECT_NAME"
```

## 📚 Learning Path

### **Beginner**: Start with Basic Usage
1. Load and explore the System Monitor plugin
2. Try basic operations like `get-metrics` and `health-check`
3. Experiment with different parameters and filters

### **Intermediate**: Combine Multiple Plugins
1. Use Package Updater with System Monitor for maintenance workflows
2. Integrate Service Manager with monitoring for service management
3. Create basic automation scripts

### **Advanced**: Custom Development
1. Study the plugin source code structure
2. Modify existing plugins for custom requirements
3. Develop new plugins using these examples as templates
4. Contribute to the plugin ecosystem

## 🤝 Contributing

### **Plugin Development Guidelines**
1. Follow the established plugin interface patterns
2. Include comprehensive documentation and examples
3. Implement proper error handling and validation
4. Add appropriate security constraints
5. Test thoroughly across different environments

### **Submitting New Examples**
1. Create plugin following the standard structure
2. Include detailed README with usage examples
3. Provide proper YAML configuration
4. Test plugin installation and operation
5. Submit PR with clear description

## 📖 Additional Resources

- **Plugin API Documentation**: `/docs/plugins/api.md`
- **Security Guidelines**: `/docs/plugins/security.md`
- **Best Practices**: `/docs/plugins/best-practices.md`
- **Testing Guide**: `/docs/plugins/testing.md`

## 🔍 Next Steps

1. **Explore the Examples**: Try each plugin to understand its capabilities
2. **Build Custom Solutions**: Modify examples for your specific needs
3. **Create New Plugins**: Use these as templates for new functionality
4. **Share with Community**: Contribute your plugins back to the ecosystem

---

**These example plugins demonstrate the full potential of the NixAI plugin system - from system monitoring to development automation, all powered by AI intelligence.** 🚀