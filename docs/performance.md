# nixai performance - Performance Monitoring and Optimization

The `nixai performance` command provides real-time system performance monitoring, cache statistics analysis, and optimization recommendations for NixOS systems. All metrics are collected from actual system sources with accurate cache hit/miss tracking.

## Usage

```bash
nixai performance [command] [options]
```

## Commands

```bash
# System performance and cache statistics
nixai performance stats

# Real-time cache monitoring  
nixai performance cache

# Performance analysis and recommendations
nixai performance analyze [--output report.json] [--include-history]

# Generate performance report
nixai performance report [--detailed] [--format json|table]

# Monitor system metrics
nixai performance monitor [--interval 5s] [--metrics cpu,memory,cache]

# Cache optimization suggestions
nixai performance optimize-cache

# Show performance history
nixai performance history [--period 24h|7d|30d] [--format chart|table]
```

## Features

### Real-time Monitoring
- **Cache Statistics**: Real cache hit/miss rates, not placeholder data
- **System Performance**: CPU utilization and load averages
- **Memory Usage**: Current memory and swap utilization  
- **Cache Efficiency**: Hit rates, miss rates, and cache optimization opportunities
- **Performance Metrics**: Real-time system performance indicators

### Performance Analysis
- Bottleneck identification
- Resource utilization patterns
- Performance regression detection
- Comparative analysis with baselines

### Optimization Recommendations
- Boot time optimization
- Memory management tuning
- Disk performance improvements
- Network configuration optimization
- Service and daemon optimization

## Examples

```bash
# Real-time performance monitoring
nixai performance monitor --real-time --metrics cpu,memory

# Comprehensive performance analysis
nixai performance analyze --include-history --output perf-report.json

# System benchmark with comparison
nixai performance benchmark --suite full --compare-baseline

# Get optimization recommendations
nixai performance optimize --category boot,memory
```

For detailed usage examples and performance tuning guides, see the main documentation.