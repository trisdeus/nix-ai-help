# nixai health - System Health Monitoring and Prediction

The `nixai health` command provides comprehensive system health monitoring with real-time metrics collection, predictive analysis, and intelligent recommendations.

## Overview

The health monitoring system continuously tracks system performance, detects anomalies, and provides predictive insights to prevent system failures before they occur. All metrics are collected from real system sources with CPU-aware thresholds and accurate calculations.

## Usage

```bash
# Basic health status check
nixai health status

# Continuous monitoring mode
nixai health monitor

# Predictive health analysis
nixai health predict

# Detailed health report
nixai health report --detailed

# JSON output for automation
nixai health status --format json
```

## Key Features

### 🔍 Real-Time System Monitoring

- **CPU Usage**: Real-time CPU utilization from `/proc/stat`
- **Memory Usage**: Current memory consumption from `/proc/meminfo`
- **Disk Usage**: File system usage and I/O metrics
- **Network Usage**: Intelligent network utilization calculation (not cumulative bytes)
- **Process Count**: Actual running processes from `/proc/` PID directories
- **Load Average**: CPU-aware load thresholds (cores × 1.5 for critical)

### 🧠 Intelligent Thresholds

The health system uses **CPU-aware thresholds** that adapt to your system's capabilities:

- **Load Average**: Critical threshold = Number of CPU cores × 1.5
  - 8-core system: Critical at 12.0 load average
  - 128-core system: Critical at 192.0 load average
- **Network Usage**: Heuristic calculation based on traffic patterns, not cumulative bytes
- **Process Count**: Monitors actual running processes, not total processes since boot

### 📊 Health Assessment Levels

- **🟢 Good**: All metrics within normal ranges
- **🟡 Warning**: Some metrics approaching critical thresholds
- **🔴 Critical**: Immediate attention required
- **⚠️ Unknown**: Unable to determine status (system access issues)

## Commands

### Status Check

```bash
# Quick status overview
nixai health status

# Detailed status with metrics
nixai health status --verbose

# JSON format for scripts
nixai health status --format json
```

**Example Output:**
```text
🏥 System Health Assessment

📊 Overall Status: Good
🖥️  CPU Usage: 15.2% (Good)
🧠 Memory Usage: 45.8% (Good)
💾 Disk Usage: 62.1% (Warning)
🌐 Network Usage: 12.3% (Good)
⚡ Load Average: 2.1 (Good - Critical: 12.0 for 8 cores)
🔢 Process Count: 1,234 (Good)

💡 Recommendations:
- Consider cleaning up disk space (62.1% usage)
- System is performing well overall
```

### Continuous Monitoring

```bash
# Start monitoring mode
nixai health monitor

# Monitor with custom interval
nixai health monitor --interval 30s

# Monitor specific metrics
nixai health monitor --metrics cpu,memory,disk
```

**Monitor Mode Features:**
- Real-time metric updates
- Threshold breach alerts
- Historical trend display
- Automatic anomaly detection

### Predictive Analysis

```bash
# Generate health predictions
nixai health predict

# Predict specific timeframe
nixai health predict --horizon 24h

# Include failure probability
nixai health predict --include-probability
```

**Prediction Capabilities:**
- Resource exhaustion forecasting
- Performance degradation trends
- Potential failure point identification
- Maintenance window recommendations

### Detailed Reporting

```bash
# Comprehensive health report
nixai health report

# Export report to file
nixai health report --output health-report.json

# Include historical data
nixai health report --history 7d
```

## Configuration

### Health Monitoring Settings

Configure health monitoring in `configs/default.yaml`:

```yaml
health:
  monitoring:
    enabled: true
    interval: "30s"
    metrics_retention: "7d"
    
  thresholds:
    cpu_warning: 70.0
    cpu_critical: 85.0
    memory_warning: 80.0
    memory_critical: 95.0
    disk_warning: 80.0
    disk_critical: 95.0
    load_multiplier: 1.5  # CPU cores × this = critical load
    
  network:
    calculation_method: "heuristic"  # or "bandwidth_based"
    low_usage_threshold: 5.0
    moderate_usage_threshold: 15.0
    high_usage_threshold: 50.0
    
  predictions:
    enabled: true
    ml_model: "isolation_forest"
    confidence_threshold: 0.7
    horizon: "24h"
```

### Alert Configuration

```yaml
health:
  alerts:
    enabled: true
    channels: ["log", "notification"]
    
    rules:
      - name: "high_cpu"
        condition: "cpu > 85"
        severity: "critical"
        cooldown: "5m"
        
      - name: "memory_pressure"
        condition: "memory > 90"
        severity: "warning"
        cooldown: "10m"
        
      - name: "load_spike"
        condition: "load > cores * 2.0"
        severity: "critical"
        cooldown: "2m"
```

## Integration Examples

### With System Monitoring

```bash
# Integration with system-info
nixai health status && nixai system-info --brief

# Combined diagnostics
nixai doctor && nixai health predict
```

### With Fleet Management

```bash
# Fleet-wide health monitoring
nixai fleet health --all-machines

# Deploy based on health status
nixai fleet deploy --healthy-only
```

### With Performance Monitoring

```bash
# Health + performance analysis
nixai health status && nixai performance stats

# Continuous monitoring
nixai health monitor &
nixai performance monitor
```

## Automation and Scripting

### Health Status Checks

```bash
#!/bin/bash
# Health check script

health_status=$(nixai health status --format json)
overall_status=$(echo "$health_status" | jq -r '.overall_status')

if [ "$overall_status" = "critical" ]; then
    echo "CRITICAL: System health issues detected"
    nixai health report --output /tmp/health-critical.json
    # Send alert
elif [ "$overall_status" = "warning" ]; then
    echo "WARNING: System performance degradation"
    nixai health predict --horizon 6h
fi
```

### Maintenance Scheduling

```bash
#!/bin/bash
# Predictive maintenance

predictions=$(nixai health predict --format json)
failure_risk=$(echo "$predictions" | jq -r '.failure_probability')

if (( $(echo "$failure_risk > 0.7" | bc -l) )); then
    echo "Scheduling maintenance: High failure risk ($failure_risk)"
    # Schedule maintenance window
fi
```

## Troubleshooting

### Common Issues

**High Load Average on Multi-Core Systems:**
- The system now correctly calculates critical thresholds
- 128-core system: Critical at 192.0 (128 × 1.5)
- Check if the load is proportional to your CPU count

**Network Always Showing High Usage:**
- Fixed: No longer reads cumulative bytes since boot
- Now uses heuristic calculation based on traffic patterns
- Returns realistic percentages (5-50% range)

**Incorrect Process Count:**
- Fixed: Now counts actual running processes from `/proc/` PID directories
- No longer reads total processes created since boot from `/proc/stat`

### Debug Mode

```bash
# Enable debug logging
nixai health status --debug

# Verbose metric collection
nixai health monitor --debug --verbose

# Test metric collection
nixai health test-metrics
```

### Performance Issues

```bash
# Reduce monitoring frequency
nixai health monitor --interval 60s

# Disable predictions
nixai health status --no-predictions

# Monitor specific metrics only
nixai health monitor --metrics cpu,memory
```

## Best Practices

1. **Regular Monitoring**: Set up continuous monitoring for production systems
2. **Threshold Tuning**: Adjust thresholds based on your system's normal operating patterns
3. **Predictive Maintenance**: Use predictions to schedule maintenance during low-usage periods
4. **Integration**: Combine with other nixai commands for comprehensive system management
5. **Automation**: Script health checks for automated monitoring and alerting

## Related Commands

- [`nixai doctor`](doctor.md) - Comprehensive system diagnostics
- [`nixai system-info`](system-info.md) - System information display
- [`nixai performance`](performance.md) - Performance monitoring and optimization
- [`nixai diagnose`](diagnose.md) - System issue diagnosis

For more advanced health monitoring configurations and enterprise features, see the [Health Monitoring Guide](../guides/health-monitoring.md).