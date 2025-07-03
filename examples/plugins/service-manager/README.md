# Service Manager Plugin

An intelligent systemd service management plugin for NixAI that provides AI-powered service monitoring, health analysis, and automated troubleshooting.

## Features

### 🔧 **Service Management**
- **Complete Service Control**: Start, stop, restart, enable, disable services
- **Real-time Status Monitoring**: Live service status and state tracking
- **Bulk Operations**: Manage multiple services simultaneously
- **Service Discovery**: Automatic detection and cataloging of services

### 🔍 **Intelligent Monitoring**
- **Health Scoring**: AI-powered health assessment (0-100 scale)
- **Performance Tracking**: CPU, memory, and resource usage monitoring
- **Dependency Analysis**: Service dependency mapping and impact analysis
- **Log Integration**: Automatic log collection and analysis

### 🚨 **Smart Alerting**
- **Configurable Alerts**: Custom alert rules for service failures
- **Threshold Monitoring**: Memory, CPU, and restart count thresholds
- **Proactive Notifications**: Early warning system for service issues
- **Alert History**: Track alert patterns and resolution times

### 🤖 **AI-Powered Analysis**
- **Health Recommendations**: AI-generated optimization suggestions
- **Issue Diagnosis**: Automatic problem identification and solutions
- **Security Assessment**: Service security posture analysis
- **Performance Optimization**: Resource usage optimization recommendations

## Installation

1. **Copy Plugin**: Place the plugin in your nixai plugins directory:
   ```bash
   cp -r service-manager ~/.nixai/plugins/
   ```

2. **Load Plugin**: Load the plugin using nixai:
   ```bash
   nixai plugin load service-manager
   ```

3. **Start Plugin**: Start the service manager:
   ```bash
   nixai plugin start service-manager
   ```

## Usage Examples

### 📋 **List Services**
```bash
# List all services
nixai plugin execute service-manager list-services

# Filter by status
nixai plugin execute service-manager list-services '{"status": "failed"}'

# Filter by name pattern
nixai plugin execute service-manager list-services '{"pattern": "nginx"}'

# Example output:
[
  {
    "name": "nginx",
    "status": "active",
    "sub_state": "running",
    "description": "A high performance web server",
    "main_pid": 1234,
    "memory_usage": 52428800,
    "start_time": "2024-01-15T10:00:00Z",
    "health": {
      "status": "healthy",
      "score": 95,
      "issues": [],
      "suggestions": []
    }
  }
]
```

### 🔍 **Get Service Details**
```bash
# Get detailed service information
nixai plugin execute service-manager get-service '{"service_name": "nginx"}'

# Example output:
{
  "name": "nginx",
  "status": "active",
  "sub_state": "running",
  "load_state": "loaded",
  "description": "A high performance web server",
  "main_pid": 1234,
  "memory_usage": 52428800,
  "cpu_usage": 2.5,
  "start_time": "2024-01-15T10:00:00Z",
  "restart_count": 0,
  "dependencies": ["network.target"],
  "health": {
    "status": "healthy",
    "score": 95,
    "last_check": "2024-01-15T11:30:00Z",
    "issues": [],
    "suggestions": []
  },
  "config": {
    "type": "forking",
    "exec_start": "/usr/sbin/nginx",
    "user": "nginx",
    "group": "nginx",
    "restart": "on-failure"
  },
  "metrics": {
    "uptime": "1h30m0s",
    "total_restarts": 0,
    "avg_memory_usage": 50331648,
    "peak_memory_usage": 52428800,
    "error_count": 0
  }
}
```

### ⚡ **Service Control Operations**
```bash
# Start a service
nixai plugin execute service-manager start-service '{"service_name": "nginx"}'

# Stop a service
nixai plugin execute service-manager stop-service '{"service_name": "nginx"}'

# Restart a service
nixai plugin execute service-manager restart-service '{"service_name": "nginx"}'

# Enable service (start on boot)
nixai plugin execute service-manager enable-service '{"service_name": "nginx"}'

# Disable service
nixai plugin execute service-manager disable-service '{"service_name": "nginx"}'

# Reload service configuration
nixai plugin execute service-manager reload-service '{"service_name": "nginx"}'

# Example output:
{
  "action": "restart",
  "service_name": "nginx",
  "success": true,
  "message": "Service restarted successfully",
  "timestamp": "2024-01-15T11:30:00Z",
  "duration": "2.5s"
}
```

### 🤖 **AI-Powered Service Analysis**
```bash
# Comprehensive service analysis
nixai plugin execute service-manager analyze-service '{"service_name": "nginx"}'

# Example output:
{
  "service": {...},
  "health_score": 95,
  "issues": [],
  "recommendations": [
    "Service is running optimally",
    "Consider enabling log rotation for better disk management"
  ],
  "performance": {
    "memory_usage": 52428800,
    "cpu_usage": 2.5,
    "restart_count": 0,
    "uptime": "1h30m0s",
    "performance_score": 92
  },
  "security": {
    "security_score": 85,
    "issues": [],
    "user": "nginx",
    "group": "nginx"
  }
}
```

### 📄 **Service Logs**
```bash
# Get recent service logs
nixai plugin execute service-manager get-service-logs '{
  "service_name": "nginx",
  "lines": 50
}'

# Example output:
[
  "Jan 15 11:30:00 server nginx[1234]: Starting nginx...",
  "Jan 15 11:30:01 server nginx[1234]: nginx started successfully",
  "Jan 15 11:30:02 server nginx[1234]: Server ready to accept connections"
]
```

### 🔗 **Dependency Analysis**
```bash
# Get service dependencies
nixai plugin execute service-manager get-service-dependencies '{
  "service_name": "nginx"
}'

# Example output:
{
  "service": "nginx",
  "dependencies": [
    "network.target",
    "multi-user.target",
    "system.slice"
  ]
}
```

### 🏥 **Health Monitoring**
```bash
# System-wide health check
nixai plugin execute service-manager health-check

# Example output:
{
  "overall_health": "healthy",
  "total_services": 45,
  "active_services": 42,
  "failed_services": 1,
  "critical_issues": [
    "Service postgresql has failed"
  ],
  "warnings": [
    "Service docker using high memory: 512 MB"
  ],
  "last_check": "2024-01-15T11:30:00Z"
}

# Single service health check
nixai plugin execute service-manager health-check '{"service_name": "nginx"}'

# Example output:
{
  "status": "healthy",
  "score": 95,
  "last_check": "2024-01-15T11:30:00Z",
  "issues": [],
  "suggestions": [
    "Service is performing well",
    "No optimization needed at this time"
  ]
}
```

## Available Operations

| Operation | Description | Parameters |
|-----------|-------------|------------|
| `list-services` | List all services with filtering | `status`, `pattern` |
| `get-service` | Get detailed service information | `service_name` |
| `start-service` | Start a systemd service | `service_name` |
| `stop-service` | Stop a systemd service | `service_name` |
| `restart-service` | Restart a systemd service | `service_name` |
| `enable-service` | Enable service for auto-start | `service_name` |
| `disable-service` | Disable service auto-start | `service_name` |
| `reload-service` | Reload service configuration | `service_name` |
| `analyze-service` | AI-powered service analysis | `service_name` |
| `get-service-logs` | Get recent service logs | `service_name`, `lines` |
| `get-service-dependencies` | Get service dependencies | `service_name` |
| `health-check` | Comprehensive health assessment | `service_name` (optional) |

## Advanced Features

### 🎯 **Health Scoring System**
Services are scored from 0-100 based on:
- **Service State** (50 points): Running vs failed/inactive
- **Restart Stability** (20 points): Low restart count indicates stability
- **Resource Usage** (20 points): Reasonable CPU/memory consumption
- **Error Rate** (10 points): Low error count in logs

### 📊 **Performance Monitoring**
```bash
# Monitor service performance metrics
{
  "memory_usage": 52428800,      # Current memory usage in bytes
  "cpu_usage": 2.5,              # CPU usage percentage
  "restart_count": 0,            # Number of restarts
  "uptime": "1h30m0s",          # Service uptime
  "performance_score": 92        # Overall performance score
}
```

### 🔒 **Security Assessment**
```bash
# Security posture analysis
{
  "security_score": 85,          # Security score (0-100)
  "issues": [],                  # Security concerns
  "user": "nginx",               # Service user
  "group": "nginx",              # Service group
  "recommendations": [           # Security improvements
    "Service properly isolated with dedicated user"
  ]
}
```

### 🚨 **Alert Configuration**
```bash
# Example alert rules
[
  {
    "id": "high_memory",
    "service_name": "*",           # All services
    "condition": "memory_high",
    "threshold": 500.0,            # 500MB threshold
    "enabled": true,
    "description": "Alert when service memory usage is high"
  },
  {
    "id": "service_failed",
    "service_name": "nginx",
    "condition": "failed",
    "threshold": 0,
    "enabled": true,
    "description": "Alert when nginx service fails"
  }
]
```

## Integration Examples

### 🔄 **Automated Service Management**
```bash
#!/bin/bash
# Auto-restart failed services script

FAILED_SERVICES=$(nixai plugin execute service-manager list-services '{"status": "failed"}' | jq -r '.[].name')

for service in $FAILED_SERVICES; do
  echo "Restarting failed service: $service"
  nixai plugin execute service-manager restart-service "{\"service_name\": \"$service\"}"
done
```

### 📈 **Health Monitoring Dashboard**
```bash
#!/bin/bash
# Generate service health report

echo "=== Service Health Report ==="
nixai plugin execute service-manager health-check

echo ""
echo "=== Critical Services ==="
nixai plugin execute service-manager list-services '{"pattern": "nginx|postgresql|redis"}' | jq '.[] | {name, status, health}'
```

### 🔔 **Alert Integration**
```bash
#!/bin/bash
# Check for service alerts and send notifications

HEALTH=$(nixai plugin execute service-manager health-check)
FAILED_COUNT=$(echo $HEALTH | jq '.failed_services')

if [ "$FAILED_COUNT" -gt 0 ]; then
  echo "ALERT: $FAILED_COUNT services have failed" | mail -s "Service Alert" admin@company.com
fi
```

## Best Practices

### 🎯 **Monitoring Strategy**
1. **Regular Health Checks**: Set up automated health monitoring
2. **Threshold Tuning**: Adjust alert thresholds based on service behavior
3. **Dependency Awareness**: Monitor critical service dependencies
4. **Performance Baselines**: Establish normal operation baselines

### 🔧 **Service Optimization**
```bash
# Identify services needing attention
nixai plugin execute service-manager list-services | jq '.[] | select(.health.score < 80)'

# Analyze high-resource services
nixai plugin execute service-manager list-services | jq '.[] | select(.memory_usage > 100000000)'
```

### 📊 **Performance Tuning**
```bash
# Generate performance report
for service in nginx postgresql redis; do
  echo "=== $service Performance ==="
  nixai plugin execute service-manager analyze-service "{\"service_name\": \"$service\"}" | jq '.performance'
done
```

## Troubleshooting

### Common Issues

#### Service Analysis Failures
```bash
# Check if systemctl is available
which systemctl

# Verify service exists
systemctl status service-name

# Check plugin status
nixai plugin status service-manager
```

#### Permission Issues
```bash
# Ensure proper systemd access
sudo systemctl --version

# Check user permissions
groups $USER
```

#### High Resource Usage
```bash
# Identify resource-heavy services
nixai plugin execute service-manager list-services | jq '.[] | select(.memory_usage > 1000000000)'

# Analyze specific service
nixai plugin execute service-manager analyze-service '{"service_name": "heavy-service"}'
```

## Security Considerations

### 🔒 **Access Control**
- **Service Permissions**: Validate systemd access permissions
- **User Isolation**: Ensure services run with appropriate users
- **Resource Limits**: Monitor and enforce resource constraints

### 🛡️ **Best Practices**
- **Principle of Least Privilege**: Services should run with minimal permissions
- **Regular Audits**: Periodic security assessment of service configurations
- **Log Monitoring**: Continuous monitoring of service logs for security events

## Performance Optimization

### ⚡ **Efficient Operations**
- **Batch Processing**: Group similar operations for efficiency
- **Caching**: Cache service information to reduce system calls
- **Selective Monitoring**: Monitor only critical services frequently

### 📊 **Resource Management**
- **Memory Optimization**: Monitor plugin memory usage
- **CPU Efficiency**: Optimize monitoring frequency
- **Network Minimal**: Minimize network calls during monitoring

## Contributing

Areas for enhancement:

1. **Enhanced AI Analysis**: More sophisticated failure prediction
2. **Custom Metrics**: User-defined performance metrics
3. **Integration Plugins**: Better integration with monitoring systems
4. **Visualization**: Graphical service dependency mapping

## License

MIT License - see LICENSE file for details.

---

**Professional systemd service management for NixOS** 🔧