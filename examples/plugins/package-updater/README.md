# Package Updater Plugin

An intelligent package update management plugin for NixAI that provides AI-powered update scheduling, security analysis, and automated dependency resolution.

## Features

### 🔄 **Smart Update Management**
- **Automated Update Checking**: Regular scanning for available package updates
- **Intelligent Scheduling**: AI-powered update prioritization and scheduling
- **Security-First Updates**: Prioritize security patches and critical fixes
- **Dependency Resolution**: Automatic dependency analysis and conflict detection

### 🛡️ **Security & Risk Analysis**
- **Security Update Identification**: Automatically identify security-related updates
- **Risk Assessment**: Evaluate update risk levels (low, medium, high)
- **Breaking Change Detection**: Identify updates with potential breaking changes
- **Rollback Support**: Easy rollback mechanism for problematic updates

### 📊 **Update Planning**
- **Smart Update Plans**: Create optimized update schedules
- **Batch Processing**: Group related updates for efficient processing
- **Impact Analysis**: Predict update impact on system stability
- **Resource Planning**: Estimate time and bandwidth requirements

### 🎯 **Policy-Based Management**
- **Update Policies**: Configurable rules for automatic updates
- **Package Exclusions**: Exclude specific packages from updates
- **Category Filtering**: Filter updates by type (security, bugfix, feature)
- **Approval Workflows**: Require manual approval for critical changes

## Installation

1. **Copy Plugin**: Place the plugin in your nixai plugins directory:
   ```bash
   cp -r package-updater ~/.nixai/plugins/
   ```

2. **Load Plugin**: Load the plugin using nixai:
   ```bash
   nixai plugin load package-updater
   ```

3. **Start Plugin**: Start the update manager:
   ```bash
   nixai plugin start package-updater
   ```

## Configuration

### Default Update Policy
```json
{
  "auto_update": false,
  "security_only": false,
  "allowed_categories": ["security", "bugfix", "feature"],
  "excluded_packages": [],
  "max_updates_per_run": 10,
  "require_approval": true,
  "backup_before_update": true
}
```

### Custom Policy Configuration
```bash
# Set custom update policy
nixai plugin execute package-updater set-update-policy '{
  "policy": {
    "auto_update": true,
    "security_only": true,
    "max_updates_per_run": 5,
    "excluded_packages": ["kernel", "systemd"]
  }
}'
```

## Usage Examples

### 🔍 **Check for Updates**
```bash
# Check for available updates
nixai plugin execute package-updater check-updates

# Example output:
{
  "total_updates": 15,
  "security_updates": 3,
  "breaking_updates": 1,
  "last_check": "2024-01-15T10:30:00Z",
  "updates": [...]
}
```

### 📋 **List Available Updates**
```bash
# List all updates
nixai plugin execute package-updater list-updates

# Filter by category
nixai plugin execute package-updater list-updates '{"category": "security"}'

# Filter by priority
nixai plugin execute package-updater list-updates '{"priority": "critical"}'

# Example output:
[
  {
    "name": "openssl",
    "current_version": "1.1.1",
    "available_version": "1.1.2",
    "category": "security",
    "priority": "critical",
    "breaking_changes": false,
    "description": "Security update for OpenSSL"
  }
]
```

### 📅 **Create Update Plan**
```bash
# Create smart update plan
nixai plugin execute package-updater create-update-plan

# Include breaking changes
nixai plugin execute package-updater create-update-plan '{"include_breaking": true}'

# Example output:
{
  "total_updates": 8,
  "security_updates": 3,
  "breaking_updates": 0,
  "total_size": 52428800,
  "estimated_time": "4m0s",
  "updates": [...],
  "warnings": [],
  "recommendations": [
    "Prioritize security updates for immediate installation"
  ]
}
```

### ⚡ **Apply Updates**
```bash
# Apply all planned updates
nixai plugin execute package-updater apply-updates

# Apply specific packages
nixai plugin execute package-updater apply-updates '{
  "packages": ["firefox", "git", "vim"]
}'

# Dry run (simulate without applying)
nixai plugin execute package-updater apply-updates '{"dry_run": true}'

# Example output:
{
  "status": "success",
  "success_count": 7,
  "fail_count": 1,
  "duration": "3m45s",
  "errors": ["vim: dependency conflict"]
}
```

### 📊 **View Update History**
```bash
# Get update history
nixai plugin execute package-updater get-update-history

# Limit results
nixai plugin execute package-updater get-update-history '{"limit": 20}'

# Example output:
[
  {
    "package_name": "firefox",
    "from_version": "120.0",
    "to_version": "121.0",
    "status": "success",
    "timestamp": "2024-01-15T10:45:00Z",
    "duration": "2m30s"
  }
]
```

### 🔍 **Analyze Specific Package**
```bash
# Analyze package update
nixai plugin execute package-updater analyze-package '{
  "package_name": "firefox"
}'

# Example output:
{
  "package": {...},
  "risk_level": "low",
  "dependencies": ["gtk3", "libx11"],
  "breaking_changes": false,
  "recommendations": [
    "Safe to update - no breaking changes detected"
  ]
}
```

### 🔄 **Rollback Updates**
```bash
# Rollback last update
nixai plugin execute package-updater rollback-update '{
  "package_name": "firefox"
}'
```

## Available Operations

| Operation | Description | Parameters |
|-----------|-------------|------------|
| `check-updates` | Check for available updates | None |
| `list-updates` | List updates with filtering | `category`, `priority` |
| `create-update-plan` | Create intelligent update plan | `include_breaking` |
| `apply-updates` | Apply package updates | `packages`, `dry_run` |
| `get-update-history` | View update history | `limit` |
| `analyze-package` | Analyze specific package | `package_name` |
| `set-update-policy` | Configure update policy | `policy` |
| `get-update-policy` | View current policy | None |
| `rollback-update` | Rollback package update | `package_name` |

## Advanced Features

### 🤖 **AI-Powered Recommendations**
- **Update Prioritization**: AI ranks updates by importance
- **Compatibility Analysis**: Predict potential conflicts
- **Optimization Suggestions**: Recommend update strategies
- **Performance Impact**: Assess update performance implications

### 📈 **Monitoring & Analytics**
- **Update Patterns**: Track update frequency and success rates
- **Performance Metrics**: Monitor update duration and resource usage
- **Trend Analysis**: Identify patterns in package updates
- **Health Monitoring**: Continuous system health assessment

### 🔧 **Integration Capabilities**
- **CI/CD Integration**: Automated testing pipelines
- **Notification Systems**: Email, Slack, webhook notifications
- **External Tools**: Integration with monitoring systems
- **Custom Hooks**: Pre/post-update script execution

## Update Categories

### 🛡️ **Security Updates**
- **Critical Vulnerabilities**: CVE patches and security fixes
- **Priority Handling**: Automatic prioritization of security updates
- **Fast-Track Processing**: Expedited security update deployment

### 🐛 **Bug Fixes**
- **Stability Improvements**: Bug fixes and stability patches
- **Performance Fixes**: Performance-related improvements
- **Compatibility Updates**: Compatibility and integration fixes

### ✨ **Feature Updates**
- **New Features**: New functionality and capabilities
- **Enhancement Updates**: Improvements to existing features
- **User Experience**: UI/UX improvements and updates

## Best Practices

### 🎯 **Update Strategy**
```bash
# 1. Regular update checks
echo "0 6 * * * nixai plugin execute package-updater check-updates" | crontab -

# 2. Security-first approach
nixai plugin execute package-updater set-update-policy '{
  "policy": {"security_only": true, "auto_update": true}
}'

# 3. Staged deployment
nixai plugin execute package-updater create-update-plan '{"include_breaking": false}'
```

### 📊 **Monitoring Setup**
```bash
# Regular health checks
nixai plugin execute package-updater get-update-history '{"limit": 10}'

# Policy review
nixai plugin execute package-updater get-update-policy
```

### 🔄 **Rollback Preparation**
- **System Snapshots**: Create snapshots before major updates
- **Dependency Mapping**: Understand package dependencies
- **Rollback Testing**: Test rollback procedures regularly

## Troubleshooting

### Common Issues

#### Update Failures
```bash
# Check update history for errors
nixai plugin execute package-updater get-update-history

# Analyze failed package
nixai plugin execute package-updater analyze-package '{"package_name": "failed-package"}'
```

#### Dependency Conflicts
```bash
# Create update plan to see conflicts
nixai plugin execute package-updater create-update-plan

# Use dry run to test updates
nixai plugin execute package-updater apply-updates '{"dry_run": true}'
```

#### Policy Issues
```bash
# Review current policy
nixai plugin execute package-updater get-update-policy

# Reset to default policy
nixai plugin execute package-updater set-update-policy '{
  "policy": {
    "auto_update": false,
    "require_approval": true
  }
}'
```

## Security Considerations

### 🔒 **Access Control**
- **Permission Validation**: Verify update permissions
- **Signature Verification**: Validate package signatures
- **Source Authentication**: Verify package sources

### 🛡️ **Risk Management**
- **Backup Requirements**: Mandatory backups before updates
- **Rollback Capabilities**: Quick rollback for failed updates
- **Testing Integration**: Automated testing before deployment

## Performance Optimization

### ⚡ **Efficient Updates**
- **Parallel Processing**: Concurrent update processing
- **Bandwidth Management**: Optimize download schedules
- **Resource Monitoring**: Monitor system resources during updates

### 📊 **Metrics & Monitoring**
- **Update Duration**: Track update completion times
- **Success Rates**: Monitor update success/failure rates
- **Resource Usage**: Track CPU, memory, and network usage

## Contributing

Contributions welcome! Areas for improvement:

1. **Enhanced AI Analysis**: More sophisticated risk assessment
2. **Additional Integrations**: Support for more package managers
3. **Performance Optimizations**: Faster update processing
4. **UI Improvements**: Better visualization and reporting

## License

MIT License - see LICENSE file for details.

---

**Intelligent package management for the modern NixOS ecosystem** 🚀