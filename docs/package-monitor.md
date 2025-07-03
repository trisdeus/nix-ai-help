# nixai package-monitor - Package Monitoring

The `nixai package-monitor` command provides comprehensive package monitoring, update tracking, and security analysis for installed packages.

## Usage

```bash
nixai package-monitor [command] [options]
```

## Commands

```bash
# Show package monitoring overview
nixai package-monitor overview [--format dashboard|table] [--include-stats]

# Monitor package updates
nixai package-monitor updates [--check-security] [--auto-update safe] [--notify]

# Package security analysis
nixai package-monitor security [--scan-vulnerabilities] [--severity critical|high|medium]

# Track package usage and dependencies
nixai package-monitor usage [--package package-name] [--period 30d] [--detailed]

# Package health check
nixai package-monitor health [--fix-issues] [--verify-integrity]

# Generate package reports
nixai package-monitor report [--type security|updates|usage] [--output report.json]
```

## Features

### Update Monitoring
- Real-time package update tracking
- Security update prioritization
- Automated safe update application
- Update notification system

### Security Analysis
- Vulnerability scanning and assessment
- CVE database integration
- Security advisory tracking
- Risk assessment and prioritization

### Usage Analytics
- Package usage statistics
- Dependency tracking and analysis
- Resource consumption monitoring
- Performance impact assessment

## Examples

```bash
# View package monitoring dashboard
nixai package-monitor overview --format dashboard

# Check for security updates
nixai package-monitor security --scan-vulnerabilities --severity critical

# Monitor package usage patterns
nixai package-monitor usage --period 30d --detailed

# Generate comprehensive package report
nixai package-monitor report --type security --output security-report.json
```

Example output:
```text
Package Monitoring Overview:
┌─────────────────────┬─────────┬──────────────┬─────────────┐
│ Category            │ Count   │ Updates Avail│ Security    │
├─────────────────────┼─────────┼──────────────┼─────────────┤
│ System Packages     │   1,247 │     23       │ 3 Critical  │
│ User Packages       │     156 │      8       │ 1 High      │
│ Development Tools   │      89 │     12       │ 0           │
│ Desktop Applications│      67 │      5       │ 2 Medium    │
└─────────────────────┴─────────┴──────────────┴─────────────┘

Recent Security Alerts:
• firefox: CVE-2024-1234 (Critical) - RCE vulnerability
• openssl: CVE-2024-5678 (High) - Certificate validation bypass
• curl: CVE-2024-9012 (Medium) - Information disclosure
```

For package security best practices and automated monitoring setup, see the main documentation.