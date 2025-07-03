# System Monitor Plugin

A comprehensive system monitoring plugin for NixAI that provides real-time resource monitoring, alerting, and health checks.

## Features

### 🔍 **System Monitoring**
- **CPU Usage**: Real-time CPU utilization tracking
- **Memory Usage**: RAM consumption monitoring  
- **Disk Usage**: Storage space utilization
- **Load Average**: System load monitoring
- **Uptime**: System uptime tracking
- **Network Statistics**: Bytes sent/received tracking

### 🚨 **Intelligent Alerting**
- **Configurable Thresholds**: Set custom warning/critical levels
- **Multi-level Alerts**: Warning and critical alert levels
- **Real-time Notifications**: Immediate alerts when thresholds exceeded
- **Alert History**: Track alert patterns over time

### 🏥 **Health Checks**
- **Comprehensive Health Reports**: Overall system status assessment
- **Actionable Recommendations**: AI-powered suggestions for optimization
- **Performance Insights**: Identify bottlenecks and issues
- **Trend Analysis**: Monitor resource usage patterns

## Installation

1. **Copy Plugin**: Place the plugin in your nixai plugins directory:
   ```bash
   cp -r system-monitor ~/.nixai/plugins/
   ```

2. **Load Plugin**: Load the plugin using nixai:
   ```bash
   nixai plugin load system-monitor
   ```

3. **Start Monitoring**: Start the plugin:
   ```bash
   nixai plugin start system-monitor
   ```

## Configuration

### Default Thresholds
```json
{
  "cpu_warning": 70.0,
  "cpu_critical": 90.0,
  "memory_warning": 80.0,
  "memory_critical": 95.0,
  "disk_warning": 85.0,
  "disk_critical": 95.0,
  "temp_warning": 70.0,
  "temp_critical": 85.0
}
```

### Custom Configuration
```bash
# Set custom thresholds
nixai plugin execute system-monitor set-thresholds '{
  "thresholds": {
    "cpu_warning": 60.0,
    "memory_warning": 75.0,
    "disk_critical": 90.0
  }
}'
```

## Usage Examples

### 📊 **Get Current Metrics**
```bash
# Get all system metrics
nixai plugin execute system-monitor get-metrics

# Example output:
{
  "cpu_usage": 45.2,
  "memory_usage": 67.8,
  "disk_usage": 45.0,
  "load_average": 1.23,
  "uptime": "2h34m12s",
  "temperature": 58.5,
  "network_stats": {
    "bytes_received": 1048576,
    "bytes_sent": 524288
  }
}
```

### 🚨 **View Active Alerts**
```bash
# Get all alerts
nixai plugin execute system-monitor get-alerts

# Get only critical alerts
nixai plugin execute system-monitor get-alerts '{"level": "critical"}'

# Example output:
[
  {
    "type": "memory",
    "level": "warning",
    "message": "Memory usage elevated: 82.3%",
    "value": 82.3,
    "threshold": 80.0,
    "timestamp": "2024-01-15T10:30:00Z"
  }
]
```

### 🏥 **Health Check**
```bash
# Comprehensive health check
nixai plugin execute system-monitor health-check

# Example output:
{
  "overall_status": "warning",
  "metrics": { ... },
  "alerts": [ ... ],
  "recommendations": [
    "Memory usage is high - consider closing unused applications",
    "Consider upgrading to SSD for better disk performance"
  ]
}
```

### ⚙️ **Threshold Management**
```bash
# View current thresholds
nixai plugin execute system-monitor get-thresholds

# Update thresholds
nixai plugin execute system-monitor set-thresholds '{
  "thresholds": {
    "cpu_warning": 65.0,
    "cpu_critical": 85.0,
    "memory_warning": 75.0
  }
}'
```

## Available Operations

| Operation | Description | Parameters |
|-----------|-------------|------------|
| `get-metrics` | Get current system metrics | None |
| `get-alerts` | Get system alerts | `level` (optional) |
| `health-check` | Comprehensive health assessment | None |
| `set-thresholds` | Update alert thresholds | `thresholds` (object) |
| `get-thresholds` | Get current thresholds | None |

## Integration with NixAI

### 🤖 **AI-Powered Insights**
The plugin integrates with NixAI's AI capabilities to provide:
- **Smart Recommendations**: Context-aware optimization suggestions
- **Predictive Alerts**: Early warning based on usage patterns
- **Automated Fixes**: AI-generated solutions for common issues

### 📈 **Dashboard Integration**
- **Web Dashboard**: View metrics in nixai web interface
- **Real-time Updates**: Live monitoring dashboard
- **Historical Charts**: Track performance over time

### 🔔 **Notification Integration**
- **System Notifications**: Desktop alerts for critical issues
- **Email Alerts**: Configure email notifications
- **Webhook Support**: Integrate with external monitoring systems

## Monitoring Best Practices

### 🎯 **Threshold Configuration**
- **Environment-Specific**: Adjust thresholds based on your system
- **Gradual Tuning**: Start conservative, adjust based on patterns
- **Seasonal Adjustments**: Account for varying workloads

### 📊 **Regular Health Checks**
```bash
# Set up automated health checks
echo "0 */6 * * * nixai plugin execute system-monitor health-check" | crontab -
```

### 🔍 **Proactive Monitoring**
- **Trend Analysis**: Monitor long-term patterns
- **Capacity Planning**: Use metrics for hardware decisions
- **Performance Optimization**: Regular system tuning

## Troubleshooting

### Plugin Not Starting
```bash
# Check plugin status
nixai plugin status system-monitor

# View plugin logs
nixai plugin logs system-monitor

# Restart plugin
nixai plugin restart system-monitor
```

### High Resource Usage
- **Adjust Check Interval**: Increase monitoring interval
- **Selective Monitoring**: Disable unused metrics
- **Threshold Tuning**: Reduce false alerts

### Missing Metrics
- **Permission Issues**: Ensure proper file access permissions
- **System Compatibility**: Verify /proc filesystem access
- **Dependencies**: Check required system tools

## Advanced Features

### 🔌 **Plugin API**
```go
// Custom integration example
metrics, err := plugin.Execute(ctx, "get-metrics", nil)
if err != nil {
    log.Printf("Failed to get metrics: %v", err)
}

// Health check integration
health, err := plugin.Execute(ctx, "health-check", nil)
if err != nil {
    log.Printf("Health check failed: %v", err)
}
```

### 📡 **Event Integration**
- **Custom Events**: Emit events for external systems
- **Webhook Support**: HTTP callbacks for alerts
- **Message Queue**: Integration with message brokers

## Security

### 🔒 **Access Control**
- **File Permissions**: Read-only access to system files
- **Sandboxing**: Runs in restricted environment
- **Resource Limits**: CPU and memory constraints

### 🛡️ **Data Privacy**
- **Local Processing**: All data stays on your system
- **No External Calls**: No data transmitted outside
- **Configurable Logging**: Control log verbosity

## Contributing

This plugin is part of the NixAI ecosystem. Contributions welcome:

1. **Feature Requests**: Open issues for new features
2. **Bug Reports**: Report issues with detailed logs
3. **Pull Requests**: Submit improvements and fixes

## License

MIT License - see LICENSE file for details.

---

**Made with ❤️ for the NixOS community**