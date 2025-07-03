# nixai system-info - System Information

The `nixai system-info` command provides comprehensive system information display and analysis for NixOS systems.

## Usage

```bash
nixai system-info [options]
```

## Options

```bash
# Basic system information
nixai system-info [--format table|json|yaml]

# Detailed system report
nixai system-info --detailed [--include-hardware] [--include-performance]

# Specific information categories
nixai system-info --category hardware|software|network|security [--verbose]

# Export system information
nixai system-info --export [--output system-info.json] [--format json|yaml|html]

# Compare with baseline
nixai system-info --compare [--baseline-file baseline.json]
```

## Features

### System Overview
- NixOS version and configuration details
- Hardware specifications and capabilities
- Installed packages and services
- System resource usage and availability

### Hardware Information
- CPU details (model, cores, frequency, features)
- Memory configuration (total, available, type)
- Storage devices (disks, partitions, filesystems)
- Network interfaces and configuration
- Graphics hardware and drivers

### Software Environment
- Kernel version and modules
- System services and their status
- Environment variables and paths
- User accounts and permissions
- Installed package versions

## Examples

```bash
# Basic system information
nixai system-info

# Detailed hardware report
nixai system-info --detailed --include-hardware

# Export comprehensive system report
nixai system-info --export --output system-report.json --format json

# Show only network information
nixai system-info --category network --verbose
```

Example output:
```text
System Information Summary:
┌────────────────────┬─────────────────────────────────────┐
│ NixOS Version      │ 24.05 (Uakari)                     │
│ Kernel             │ Linux 6.6.32                       │
│ Architecture       │ x86_64                              │
│ Hostname           │ nixos-workstation                   │
│ Uptime             │ 5 days, 14:32:18                   │
│ Load Average       │ 0.45, 0.52, 0.48                   │
└────────────────────┴─────────────────────────────────────┘

Hardware Overview:
┌────────────────────┬─────────────────────────────────────┐
│ CPU                │ Intel Core i7-12700K (12 cores)    │
│ Memory             │ 32 GB DDR4-3200                    │
│ Storage            │ 1TB NVMe SSD + 2TB SATA HDD        │
│ Graphics           │ NVIDIA RTX 4070 + Intel UHD 770    │
└────────────────────┴─────────────────────────────────────┘
```

For comprehensive system analysis and hardware optimization recommendations, see the main documentation.