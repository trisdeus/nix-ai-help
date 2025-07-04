# nixai test - Safe Configuration Testing

**Phase 3.1 - Enterprise-Grade Configuration Testing Infrastructure**

The `nixai test` command provides comprehensive safe configuration testing with virtual environments, statistical analysis, chaos engineering, and automated rollback capabilities. Test your NixOS configurations with confidence before deployment to production systems.

---

## 🧪 Core Testing Components

### 🛡️ Virtual Environment Testing
- **Isolated NixOS Containers**: Test configurations in completely isolated environments using nixos-container
- **Real System Monitoring**: Authentic data collection from `/proc/*` filesystem - no mock responses
- **Resource Management**: CPU, memory, and disk quotas to prevent system overload
- **Automatic Cleanup**: Environments are automatically destroyed after testing

### ⚖️ A/B Testing Framework
- **Statistical Comparison**: Compare two configurations with confidence intervals and effect size analysis
- **Performance Benchmarking**: Boot time, response time, throughput, and resource usage comparison
- **Significance Testing**: P-value calculations and statistical power analysis
- **Automated Recommendations**: AI-powered suggestions based on test results

### 💥 Chaos Engineering
- **12+ Attack Types**: Service kills, network issues, resource stress, time skew, file corruption
- **Resilience Scoring**: Quantitative assessment of system stability under stress
- **Weakness Detection**: Identify vulnerabilities and single points of failure
- **Recovery Analysis**: Measure system recovery time and self-healing capabilities

### 🔄 Automated Rollback
- **Risk Assessment**: Calculate rollback success probability based on configuration complexity
- **Step-by-Step Execution**: Multi-step rollback with verification at each stage
- **Emergency Recovery**: Abort mechanisms and manual intervention capabilities
- **Prerequisites Validation**: Ensure rollback conditions are met before execution

### 📈 Performance Simulation
- **Workload Modeling**: Support for constant, ramp, spike, wave, and realistic patterns
- **Capacity Planning**: Project future resource needs and bottleneck analysis
- **Resource Projections**: CPU, memory, disk, and network utilization forecasts
- **Scalability Testing**: Determine breaking points and optimization opportunities

---

## 🚀 Command Reference

### Basic Usage

```bash
# Quick comprehensive test
nixai test quick /etc/nixos/configuration.nix

# Get help for specific subcommand
nixai test env --help
nixai test compare --help
nixai test chaos --help
```

### Global Flags

- `--duration` - Test duration (default: 30m)
- `--sample-size` - Sample size for statistical tests (default: 100)
- `--confidence` - Confidence level for statistical tests (default: 0.95)
- `--workload` - Workload pattern: constant, ramp, spike, wave (default: constant)
- `--verbose` - Enable verbose output

---

## 🔧 Subcommands

### Virtual Environment Management (`nixai test env`)

Manage isolated testing environments for safe configuration validation.

#### Create Environment
```bash
# Create test environment from configuration file
nixai test env create /etc/nixos/configuration.nix

# Create with specific resource allocation
nixai test env create myconfig.nix --cpu 4 --memory 4096 --disk 20
```

#### List Environments
```bash
# List all test environments with status
nixai test env list

# Show detailed environment information
nixai test env status env_1234567890
```

#### Environment Cleanup
```bash
# Delete specific environment
nixai test env delete env_1234567890

# Clean up all stopped environments
nixai test env cleanup
```

### A/B Testing (`nixai test compare`)

Statistical comparison between two configurations with comprehensive analysis.

#### Create A/B Test
```bash
# Compare two configuration files
nixai test compare create config-a.nix config-b.nix

# A/B test with custom parameters
nixai test compare create config-a.nix config-b.nix \
  --duration 1h \
  --sample-size 200 \
  --confidence 0.99
```

#### Run A/B Test
```bash
# Start A/B test execution
nixai test compare start abtest_1234567890

# Monitor test progress
nixai test compare status abtest_1234567890
```

#### View Results
```bash
# Get comprehensive A/B test results
nixai test compare results abtest_1234567890

# Export results to JSON for further analysis
nixai test compare results abtest_1234567890 --format json > results.json
```

**Sample A/B Test Results:**
```
📊 A/B Test Results: abtest_1234567890

Overall Winner: A
Confidence: 87.5%
Statistical Power: 80.0%
Effect Size: 0.32

🚀 Performance Comparison:
- Boot Time: A=12.3s, B=15.7s, Winner=A (21.7% faster) ⚠️
- Response Time: A=87.2ms, B=92.1ms, Winner=A (5.3% faster)
- Throughput: A=1247 ops/s, B=1189 ops/s, Winner=A (4.9% higher)
- Error Rate: A=0.1%, B=0.3%, Winner=A (66.7% lower) ⚠️

💾 Resource Usage Comparison:
- CPU: A=34.2%, B=41.7%, Winner=A (18.0% lower) ⚠️
- Memory: A=67.8%, B=72.1%, Winner=A (6.0% lower)
- Disk: A=23.4%, B=25.8%, Winner=A (9.3% lower)
- Network: A=12.3 MB/s, B=11.7 MB/s, Winner=A (5.1% higher)

💡 Recommendations:
- Configuration A shows better response times: Configuration A has 21.7% better response times
- Configuration A uses less memory: Configuration A uses 18.0% less memory
```

### Chaos Engineering (`nixai test chaos`)

Resilience testing through controlled failure injection and system stress.

#### Create Chaos Experiment
```bash
# Create chaos experiment for environment
nixai test chaos create env_1234567890

# Custom chaos experiment with specific attacks
nixai test chaos create env_1234567890 \
  --attacks service_kill,cpu_stress,network_latency \
  --duration 45m \
  --intensity 0.7
```

#### Execute Chaos Test
```bash
# Start chaos experiment
nixai test chaos start chaos_1234567890

# Monitor experiment progress
nixai test chaos status chaos_1234567890

# Abort running experiment if needed
nixai test chaos abort chaos_1234567890
```

#### Analyze Results
```bash
# Get chaos experiment results
nixai test chaos results chaos_1234567890

# Generate resilience report
nixai test chaos report chaos_1234567890 --format pdf
```

**Sample Chaos Engineering Results:**
```
💥 Chaos Experiment Results: chaos_1234567890

Status: completed
Target Environment: env_1234567890
Resilience Score: 78.3%
Steady State Valid: true
Hypothesis Proven: true
Recovery Time: 2m34s

⚔️ Attack Results:
- service_kill_1: Success=true, Impact=medium, Recovery=45s
- cpu_stress_1: Success=true, Impact=low, Recovery=12s
- network_latency_1: Success=true, Impact=high, Recovery=3m15s

🔍 Weaknesses Found:
- network (high): High impact from network_latency_1 attack
- service_restart (medium): Service recovery took longer than expected

💡 Insights:
- System shows good resilience to chaos attacks
- System successfully returned to steady state after attacks
- Network latency has significant impact on system performance
```

### Rollback Management (`nixai test rollback`)

Automated rollback planning and execution with risk assessment.

#### Generate Rollback Plan
```bash
# Create rollback plan for configuration
nixai test rollback plan /etc/nixos/configuration.nix

# Plan with custom risk tolerance
nixai test rollback plan myconfig.nix --risk-tolerance high
```

#### Execute Rollback
```bash
# Execute rollback plan in environment
nixai test rollback execute rollback_1234567890 env_1234567890

# Dry-run rollback (simulation only)
nixai test rollback execute rollback_1234567890 env_1234567890 --dry-run
```

#### Monitor Rollback
```bash
# Check rollback execution status
nixai test rollback status execution_1234567890

# View detailed rollback logs
nixai test rollback logs execution_1234567890
```

**Sample Rollback Results:**
```
🔄 Rollback Execution Status: execution_1234567890

Status: completed
Current Step: 5
Started: 2025-07-04 12:30:15
Completed: 2025-07-04 12:33:42
Duration: 3m27s
Success Rate: 100.0%

📋 Step Results:
✅ Step 1 (emergency_snapshot): completed
✅ Step 2 (stop_services): completed
✅ Step 3 (rollback_config): completed
✅ Step 4 (restart_services): completed
✅ Step 5 (verify_health): completed
```

### Performance Simulation (`nixai test simulate`)

Model system performance under various workload conditions.

#### Create Simulation
```bash
# Create performance simulation
nixai test simulate create /etc/nixos/configuration.nix

# Simulation with specific workload pattern
nixai test simulate create myconfig.nix \
  --workload spike \
  --duration 2h \
  --peak-load 1000
```

#### Run Simulation
```bash
# Start performance simulation
nixai test simulate start simulation_1234567890

# Monitor simulation progress
nixai test simulate progress simulation_1234567890
```

#### Analyze Results
```bash
# Get simulation results
nixai test simulate results simulation_1234567890

# Generate capacity planning report
nixai test simulate capacity-report simulation_1234567890
```

**Sample Performance Simulation Results:**
```
📈 Performance Simulation Results: simulation_1234567890

Status: completed
Progress: 100.0%
Overall Score: 82.1/100
Performance Grade: B+

💾 Resource Utilization:
Peak CPU: 87.3%
Peak Memory: 74.2%
Peak Disk: 45.8%
Peak Network: 156.7 MB/s

🚧 Bottlenecks Identified:
Primary: cpu (Score: 8.7)
- cpu: Context switching overhead under high load (Impact: 8.7)
- memory: Cache pressure during peak usage (Impact: 4.2)

💡 Recommendations:
- Optimize CPU usage: Consider CPU governor tuning for better performance
- Improve memory efficiency: Enable zram compression to reduce memory pressure
```

### Quick Testing (`nixai test quick`)

Comprehensive test suite that combines multiple testing approaches for rapid validation.

```bash
# Run full test suite on configuration
nixai test quick /etc/nixos/configuration.nix

# Quick test with verbose output
nixai test quick myconfig.nix --verbose

# Quick test with custom duration
nixai test quick myconfig.nix --duration 15m
```

**Quick Test Process:**
1. **Environment Creation**: Creates isolated test environment
2. **Rollback Planning**: Generates emergency rollback plan
3. **Performance Baseline**: Establishes performance metrics
4. **Basic Chaos Testing**: Runs essential resilience tests
5. **Summary Report**: Provides overall configuration assessment

---

## 🔬 Advanced Features

### Statistical Analysis
- **Confidence Intervals**: 95% confidence intervals for all performance metrics
- **Effect Size Calculation**: Cohen's d for meaningful difference assessment
- **Power Analysis**: Statistical power calculation for test reliability
- **Significance Testing**: P-value calculation for hypothesis validation

### Real System Integration
- **Authentic Data Collection**: Direct reads from `/proc/stat`, `/proc/meminfo`, `/proc/loadavg`
- **CPU-Aware Thresholds**: Scaling based on actual hardware capabilities
- **Process Monitoring**: Real running process counts, not historical totals
- **Network Analysis**: Actual utilization calculation, not cumulative bytes

### Resource Management
- **Environment Isolation**: Complete separation using nixos-container
- **Resource Quotas**: CPU cores, memory limits, disk space allocation
- **Automatic Cleanup**: Environments destroyed after testing completion
- **Resource Monitoring**: Real-time tracking of resource consumption

### Safety Features
- **Emergency Abort**: Stop any operation immediately if needed
- **Risk Assessment**: Automated risk calculation before dangerous operations
- **Verification Steps**: Multi-level validation before executing changes
- **Backup Creation**: Automatic snapshots before testing begins

---

## 📊 Testing Methodologies

### Virtual Environment Testing
Best for isolated testing of individual configurations:
- **Use Case**: Testing new packages, services, or system changes
- **Safety**: Complete isolation from host system
- **Performance**: Minimal overhead with container technology
- **Cleanup**: Automatic environment destruction

### A/B Testing
Best for comparing two configuration options:
- **Use Case**: Choosing between different optimization approaches
- **Methodology**: Statistical comparison with confidence intervals
- **Metrics**: Performance, resource usage, reliability
- **Decision**: Data-driven configuration selection

### Chaos Engineering
Best for resilience validation:
- **Use Case**: Testing system stability under stress
- **Methodology**: Controlled failure injection
- **Analysis**: Recovery time and weakness identification
- **Improvement**: Targeted resilience enhancements

### Performance Simulation
Best for capacity planning:
- **Use Case**: Predicting system behavior under load
- **Methodology**: Workload pattern modeling
- **Output**: Bottleneck analysis and capacity projections
- **Planning**: Resource allocation and scaling decisions

---

## 🛡️ Security Considerations

### Environment Isolation
- **Container Technology**: Uses nixos-container for complete isolation
- **Network Separation**: Isolated network namespaces for testing
- **Filesystem Isolation**: Separate filesystem trees prevent contamination
- **Process Isolation**: Testing processes cannot affect host system

### Resource Protection
- **CPU Limits**: Prevent testing from overwhelming host CPU
- **Memory Limits**: Protect host memory from test environment usage
- **Disk Quotas**: Limit disk space consumption for test environments
- **Network Throttling**: Prevent network saturation during testing

### Data Safety
- **No Production Access**: Test environments cannot access production data
- **Temporary Storage**: All test data stored in temporary locations
- **Automatic Cleanup**: Complete removal of test artifacts after completion
- **Backup Verification**: Ensure backups exist before destructive operations

---

## 🔧 Configuration Options

### Test Environment Configuration
```yaml
# ~/.config/nixai/test-config.yaml
test_environments:
  default_resources:
    cpu_cores: 2
    memory_mb: 2048
    disk_gb: 10
  container_runtime: "nixos-container"
  cleanup_after: "24h"
  max_environments: 50

# A/B Testing Configuration
ab_testing:
  default_duration: "30m"
  default_sample_size: 100
  confidence_level: 0.95
  significance_threshold: 0.05

# Chaos Engineering Configuration
chaos_engineering:
  default_attacks: ["service_kill", "cpu_stress", "memory_pressure"]
  max_intensity: 0.8
  safety_timeout: "10m"
  recovery_timeout: "5m"

# Performance Simulation Configuration
performance_simulation:
  default_workload: "constant"
  workload_patterns:
    spike_duration: "5m"
    ramp_duration: "10m"
    wave_period: "15m"
```

---

## 💡 Best Practices

### Environment Management
1. **Clean Up Regularly**: Remove unused test environments to save resources
2. **Monitor Resources**: Keep track of resource usage across all environments
3. **Use Descriptive Names**: Name environments clearly for easy identification
4. **Backup Before Testing**: Always create backups before destructive testing

### A/B Testing
1. **Define Success Metrics**: Clearly specify what constitutes success
2. **Use Adequate Sample Size**: Ensure statistical power for reliable results
3. **Control Variables**: Keep external factors constant between tests
4. **Validate Results**: Cross-verify significant findings with additional tests

### Chaos Engineering
1. **Start Small**: Begin with low-intensity attacks and gradually increase
2. **Monitor Closely**: Watch system behavior during chaos experiments
3. **Have Abort Plan**: Always be ready to stop experiments if needed
4. **Learn from Failures**: Document and address discovered weaknesses

### Performance Simulation
1. **Use Realistic Workloads**: Model actual usage patterns accurately
2. **Test Multiple Scenarios**: Simulate various load conditions
3. **Plan for Growth**: Model future capacity needs based on trends
4. **Validate Simulations**: Compare simulation results with real data

---

## 🐛 Troubleshooting

### Common Issues

#### Environment Creation Fails
```bash
# Check system resources
nixai test env status --resources

# Verify nixos-container is available
which nixos-container

# Check for permission issues
sudo nixos-container list
```

#### A/B Test Shows No Significant Difference
```bash
# Increase sample size for more statistical power
nixai test compare create config-a.nix config-b.nix --sample-size 500

# Extend test duration for more data points
nixai test compare create config-a.nix config-b.nix --duration 2h

# Lower confidence level for easier significance
nixai test compare create config-a.nix config-b.nix --confidence 0.90
```

#### Chaos Experiment Fails to Complete
```bash
# Check experiment status and logs
nixai test chaos status chaos_1234567890
nixai test chaos logs chaos_1234567890

# Reduce attack intensity
nixai test chaos create env_123 --intensity 0.3

# Use fewer concurrent attacks
nixai test chaos create env_123 --attacks service_kill
```

#### Performance Simulation Unrealistic Results
```bash
# Verify workload pattern matches real usage
nixai test simulate create config.nix --workload realistic

# Increase simulation duration for better accuracy
nixai test simulate create config.nix --duration 4h

# Check resource allocation matches target system
nixai test simulate create config.nix --cpu 8 --memory 16384
```

### Debug Mode
```bash
# Enable verbose logging for all test operations
export NIXAI_TEST_DEBUG=1
nixai test --verbose [command]

# Check test environment logs
nixai test env logs env_1234567890

# Validate test configuration
nixai test validate-config /etc/nixos/configuration.nix
```

---

## 🔗 Related Commands

- [`nixai health`](health.md) - System health monitoring and diagnostics
- [`nixai build`](build.md) - Build troubleshooting and optimization
- [`nixai diagnose`](diagnose.md) - System diagnostics and issue detection
- [`nixai fleet`](fleet.md) - Multi-machine configuration management
- [`nixai version-control`](version-control.md) - Configuration version control

---

## 🎯 Real-World Examples

### Development Workflow Testing
```bash
# Test development environment configuration
nixai test quick dev-env.nix

# Compare development vs production configs
nixai test compare create dev-config.nix prod-config.nix

# Test resilience of development setup
nixai test chaos create dev_env --attacks service_kill,network_latency
```

### Production Deployment Validation
```bash
# Comprehensive production readiness test
nixai test quick production.nix --duration 2h --verbose

# Test rollback procedures
nixai test rollback plan production.nix
nixai test rollback execute rollback_plan prod_test_env --dry-run

# Validate capacity planning
nixai test simulate create production.nix --workload realistic --duration 24h
```

### Configuration Optimization
```bash
# Compare different optimization approaches
nixai test compare create config-baseline.nix config-optimized.nix

# Test system resilience after optimization
nixai test chaos create optimized_env --duration 1h

# Validate performance improvements
nixai test simulate create config-optimized.nix --workload spike
```

---

The `nixai test` command provides enterprise-grade testing capabilities that ensure your NixOS configurations are reliable, performant, and resilient before deployment to production systems. Use these testing tools to build confidence in your configuration changes and maintain system stability.