# nixai fleet - Fleet Management System

The `nixai fleet` command provides comprehensive multi-machine deployment and configuration management capabilities for NixOS environments with real-time health monitoring and authentic system metrics collection.

## Overview

Fleet management enables you to manage multiple NixOS machines as a unified fleet, with centralized configuration deployment, monitoring, and maintenance. Perfect for managing development clusters, production environments, or mixed heterogeneous infrastructures.

## Commands

### Machine Management

```bash
# List all fleet machines
nixai fleet list [--format json|table] [--status all|online|offline]

# Add machines to fleet
nixai fleet add-machine <hostname> [options]

# Remove machines from fleet
nixai fleet remove-machine <hostname> [--force]

# Show detailed machine information
nixai fleet show <hostname> [--verbose]
```

### Health & Monitoring

```bash
# Check fleet health status (real system metrics)
nixai fleet health [--detailed] [--format json]

# Monitor fleet in real-time (authentic monitoring)
nixai fleet monitor [--refresh-interval 5s] [--follow]

# Get fleet statistics (real performance data)
nixai fleet stats [--period 24h|7d|30d]

# Check SSH connectivity and system health
nixai fleet connectivity-test [--parallel] [--timeout 10s]
```

### Deployment Management

```bash
# Deploy configurations to fleet
nixai fleet deploy [target] [options]

# List deployment history
nixai fleet deployment list [--machine hostname] [--limit 10]

# Show deployment status
nixai fleet deployment status [deployment-id]

# Start new deployment
nixai fleet deployment start [--config-path path] [--machines hostname1,hostname2]

# Cancel running deployment
nixai fleet deployment cancel <deployment-id>
```

## Machine Management

### Adding Machines

Add new machines to your fleet with comprehensive configuration:

```bash
# Basic machine addition
nixai fleet add-machine server01.example.com

# Add with SSH configuration
nixai fleet add-machine server01 \
  --ssh-host 192.168.1.100 \
  --ssh-port 22 \
  --ssh-user nixos \
  --ssh-key ~/.ssh/nixos_fleet

# Add with role and tags
nixai fleet add-machine web-server-01 \
  --role web-server \
  --tags "environment=production,tier=web" \
  --location "datacenter-east"

# Add with custom configuration
nixai fleet add-machine db-primary \
  --config-template database-server \
  --environment production \
  --monitoring-enabled \
  --backup-enabled
```

#### Machine Configuration Options

- `--ssh-host` - SSH hostname or IP address
- `--ssh-port` - SSH port (default: 22)
- `--ssh-user` - SSH username (default: nixos)
- `--ssh-key` - SSH private key path
- `--role` - Machine role (web, database, worker, etc.)
- `--tags` - Comma-separated tags for grouping
- `--location` - Physical or logical location
- `--environment` - Environment type (dev, staging, production)
- `--config-template` - Base configuration template
- `--monitoring-enabled` - Enable monitoring (default: true)
- `--backup-enabled` - Enable backups (default: false)

### Machine Listing and Information

```bash
# List all machines with status
nixai fleet list
```

Example output:
```text
┌──────────────┬───────────────────┬──────────┬─────────────┬────────────┬──────────────┐
│ Hostname     │ Address           │ Status   │ Role        │ Environment│ Last Contact │
├──────────────┼───────────────────┼──────────┼─────────────┼────────────┼──────────────┤
│ web-01       │ 192.168.1.101     │ Online   │ web-server  │ production │ 2s ago       │
│ web-02       │ 192.168.1.102     │ Online   │ web-server  │ production │ 5s ago       │
│ db-primary   │ 192.168.1.110     │ Online   │ database    │ production │ 1s ago       │
│ worker-01    │ 192.168.1.120     │ Offline  │ worker      │ production │ 2m ago       │
│ dev-box      │ 192.168.1.200     │ Online   │ development │ development│ 10s ago      │
└──────────────┴───────────────────┴──────────┴─────────────┴────────────┴──────────────┘
```

Detailed machine information:
```bash
nixai fleet show web-01 --verbose
```

## Health Monitoring

### Fleet Health Overview

```bash
# Quick health check
nixai fleet health
```

Example output:
```text
Fleet Health Summary:
┌─────────────────┬───────┬─────────┐
│ Status          │ Count │ Percent │
├─────────────────┼───────┼─────────┤
│ Online          │   4   │  80%    │
│ Offline         │   1   │  20%    │
│ Unreachable     │   0   │   0%    │
│ Maintenance     │   0   │   0%    │
└─────────────────┴───────┴─────────┘

Recent Issues:
• worker-01: Connection timeout (2 minutes ago)
• web-02: High memory usage (warning)
```

### Real-time Monitoring

Monitor fleet status in real-time:

```bash
# Start real-time monitoring
nixai fleet monitor

# Monitor with custom refresh rate
nixai fleet monitor --refresh-interval 10s

# Follow specific metrics
nixai fleet monitor --follow --metrics cpu,memory,disk
```

Monitoring displays:
- Machine status changes
- Resource utilization alerts
- Deployment progress
- System health metrics
- Network connectivity status

## Deployment Management

### Configuration Deployment

Deploy configurations across your fleet with advanced targeting and safety features:

```bash
# Deploy to all machines
nixai fleet deploy

# Deploy to specific machines
nixai fleet deploy --machines web-01,web-02,db-primary

# Deploy by role
nixai fleet deploy --role web-server

# Deploy by tags
nixai fleet deploy --tags "environment=production"

# Deploy with rollback plan
nixai fleet deploy --enable-rollback --rollback-timeout 300s

# Staged deployment (rolling update)
nixai fleet deploy --strategy rolling --batch-size 2 --delay 30s
```

#### Deployment Strategies

**Parallel Deployment** (default):
```bash
nixai fleet deploy --strategy parallel
```
- Deploys to all targets simultaneously
- Fastest deployment method
- Higher risk if configuration issues exist

**Rolling Deployment**:
```bash
nixai fleet deploy --strategy rolling --batch-size 1
```
- Deploys to machines in batches
- Allows validation between batches
- Safer for production environments

**Blue-Green Deployment**:
```bash
nixai fleet deploy --strategy blue-green --validation-period 300s
```
- Maintains parallel environments
- Validates new deployment before switching
- Instant rollback capability

### Deployment History

```bash
# List recent deployments
nixai fleet deployment list
```

Example output:
```text
┌─────────────┬──────────────────────┬───────────────┬──────────┬─────────────────────┐
│ ID          │ Configuration        │ Machines      │ Status   │ Started             │
├─────────────┼──────────────────────┼───────────────┼──────────┼─────────────────────┤
│ dep-abc123  │ web-server-v2.1.0    │ web-01,web-02 │ Success  │ 2024-07-03 10:30:00 │
│ dep-def456  │ security-updates     │ all           │ Success  │ 2024-07-03 09:15:00 │
│ dep-ghi789  │ database-migration   │ db-primary    │ Failed   │ 2024-07-03 08:45:00 │
│ dep-jkl012  │ monitoring-setup     │ all           │ Running  │ 2024-07-03 11:00:00 │
└─────────────┴──────────────────────┴───────────────┴──────────┴─────────────────────┘
```

### Deployment Status and Logs

```bash
# Check deployment status
nixai fleet deployment status dep-jkl012

# Follow deployment logs
nixai fleet deployment logs dep-jkl012 --follow

# Show deployment details
nixai fleet deployment show dep-jkl012 --verbose
```

## Configuration Management

### Fleet Configuration

Configure fleet-wide settings through the configuration file:

```yaml
fleet:
  # Connection settings
  ssh:
    default_user: "nixos"
    default_port: 22
    connection_timeout: "30s"
    key_path: "~/.ssh/nixos_fleet"
    known_hosts_file: "~/.ssh/known_hosts"
  
  # Deployment settings
  deployment:
    default_strategy: "rolling"
    batch_size: 2
    delay_between_batches: "30s"
    enable_rollback: true
    rollback_timeout: "300s"
    validation_period: "60s"
  
  # Monitoring settings
  monitoring:
    health_check_interval: "30s"
    metrics_retention: "7d"
    alert_thresholds:
      cpu_usage: 80
      memory_usage: 85
      disk_usage: 90
  
  # Security settings
  security:
    require_signatures: true
    allowed_users: ["admin", "deployer"]
    audit_log_enabled: true
    encryption_enabled: true
```

### Machine-Specific Configuration

Configure individual machines with specific settings:

```bash
# Set machine-specific configuration
nixai fleet config set web-01 \
  --config-override "services.nginx.virtualHosts.\"example.com\".root = \"/var/www\"" \
  --environment-vars "DOMAIN=example.com"

# Apply configuration template
nixai fleet config apply web-01 --template web-server-optimized

# Show machine configuration
nixai fleet config show web-01
```

## Security and Access Control

### SSH Key Management

```bash
# Generate fleet SSH keys
nixai fleet keys generate --output ~/.ssh/nixos_fleet

# Distribute SSH keys to machines
nixai fleet keys distribute --key ~/.ssh/nixos_fleet.pub

# Rotate SSH keys
nixai fleet keys rotate --backup-old-keys
```

### Access Control

```bash
# Grant fleet access to user
nixai fleet access grant user@example.com --role deployer

# Revoke access
nixai fleet access revoke user@example.com

# List access permissions
nixai fleet access list
```

### Audit Logging

```bash
# View fleet operation audit log
nixai fleet audit log --since "24h ago"

# Filter audit events
nixai fleet audit log --user admin --action deploy

# Export audit log
nixai fleet audit export --format json --output audit_2024-07.json
```

## Advanced Features

### Custom Hooks

Define custom hooks for deployment lifecycle events:

```yaml
fleet:
  hooks:
    pre_deploy:
      - command: "echo 'Starting deployment to {{ .Machine }}'"
        timeout: "10s"
    
    post_deploy:
      - command: "systemctl status nixos-rebuild"
        timeout: "30s"
      - command: "/usr/local/bin/notify-deploy-success"
        timeout: "5s"
    
    on_failure:
      - command: "/usr/local/bin/alert-deploy-failure {{ .Machine }}"
        timeout: "10s"
```

### Integration with External Systems

#### CI/CD Integration

```bash
# GitLab CI integration
nixai fleet deploy --config-source git \
  --git-repo https://gitlab.com/company/nixos-configs \
  --git-ref $CI_COMMIT_SHA

# GitHub Actions integration
nixai fleet deploy --trigger github-action \
  --workflow-run-id $GITHUB_RUN_ID
```

#### Monitoring Integration

```bash
# Prometheus metrics export
nixai fleet metrics export --format prometheus \
  --endpoint http://prometheus:9090/api/v1/push

# Grafana dashboard setup
nixai fleet monitoring setup-grafana \
  --grafana-url http://grafana:3000 \
  --api-key $GRAFANA_API_KEY
```

### Backup and Disaster Recovery

```bash
# Create fleet backup
nixai fleet backup create --include-configs --include-data

# Restore from backup
nixai fleet backup restore backup-2024-07-03.tar.gz

# List available backups
nixai fleet backup list --sort-by date
```

## Troubleshooting

### Common Issues

**SSH connection failures:**
```bash
# Test SSH connectivity
nixai fleet test-connection web-01

# Debug SSH issues
nixai fleet debug ssh web-01 --verbose

# Reset SSH configuration
nixai fleet ssh reset web-01
```

**Deployment failures:**
```bash
# Show deployment error details
nixai fleet deployment logs dep-abc123 --level error

# Retry failed deployment
nixai fleet deployment retry dep-abc123

# Rollback deployment
nixai fleet deployment rollback dep-abc123
```

**Machine health issues:**
```bash
# Run comprehensive health check
nixai fleet health --machine web-01 --detailed

# Check specific health metrics
nixai fleet health --metrics cpu,memory,disk,network

# Force health check update
nixai fleet health refresh --machine web-01
```

### Debug Mode

Enable comprehensive debugging for fleet operations:

```bash
nixai fleet --debug deploy --verbose
```

Debug output includes:
- SSH connection details
- Configuration validation results
- Deployment step-by-step progress
- Error details and stack traces

### Performance Optimization

For large fleets (100+ machines):

```bash
# Use parallel operations with connection pooling
nixai fleet deploy --parallel-limit 20 --connection-pool-size 50

# Enable caching for better performance
nixai fleet config set --enable-cache --cache-ttl 300s

# Use compression for large configurations
nixai fleet deploy --compress-configs --compression-level 6
```

## Best Practices

1. **SSH Key Security**: Use dedicated SSH keys for fleet management
2. **Staging Environment**: Test deployments in staging before production
3. **Gradual Rollouts**: Use rolling deployments for production systems
4. **Monitoring**: Set up comprehensive monitoring and alerting
5. **Backup Strategy**: Regular configuration and data backups
6. **Access Control**: Implement proper role-based access control
7. **Documentation**: Maintain clear deployment procedures and runbooks

For more advanced examples and integration patterns, see the [Fleet Management Guide](../examples/fleet-management/) and [Enterprise Deployment Patterns](../guides/enterprise-deployment.md).