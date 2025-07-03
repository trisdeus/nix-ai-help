# nixai performance - Performance Monitoring and Optimization

The `nixai performance` command provides system performance monitoring, analysis, and optimization recommendations for NixOS systems.

## Usage

```bash
nixai performance [command] [options]
```

## Commands

```bash
# System performance overview
nixai performance overview [--real-time] [--duration 60s]

# Monitor specific metrics
nixai performance monitor [--metrics cpu,memory,disk,network] [--interval 5s]

# Performance analysis and recommendations
nixai performance analyze [--output report.json] [--include-history]

# Benchmark system performance
nixai performance benchmark [--suite basic|full|custom] [--compare-baseline]

# Generate optimization suggestions
nixai performance optimize [--category all|boot|memory|disk|network]

# Show performance history
nixai performance history [--period 24h|7d|30d] [--format chart|table]
```

## Features

### Real-time Monitoring
- CPU utilization and load averages
- Memory usage and swap utilization
- Disk I/O and storage performance
- Network throughput and latency
- System temperature and throttling

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