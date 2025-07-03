# nixai version-control - Configuration Version Control

The `nixai version-control` command provides Git-like version control capabilities specifically designed for NixOS configurations, enabling team collaboration, change tracking, and configuration management workflows.

## Overview

Version control for NixOS configurations brings software development best practices to system administration, allowing teams to collaborate on infrastructure configurations with branching, merging, history tracking, and rollback capabilities.

## Commands

### Repository Management

```bash
# Initialize configuration repository
nixai version-control init [--path /etc/nixos] [--remote-url git@github.com:company/nixos-configs]

# Clone existing configuration repository
nixai version-control clone <repository-url> [--path /etc/nixos]

# Show repository status
nixai version-control status [--detailed]

# Show configuration changes
nixai version-control diff [--staged] [--file configuration.nix]
```

### Commit Management

```bash
# Create configuration commit
nixai version-control commit [--message "commit message"] [--all]

# Show commit history
nixai version-control history [--limit 20] [--author username] [--since "1 week ago"]

# Show specific commit details
nixai version-control show <commit-id> [--detailed]

# Revert to previous commit
nixai version-control revert <commit-id> [--confirm]
```

### Branch Management

```bash
# List branches
nixai version-control branch list [--all] [--remote]

# Create new branch
nixai version-control branch create <branch-name> [--from main]

# Switch to branch
nixai version-control branch switch <branch-name> [--create-if-missing]

# Merge branches
nixai version-control branch merge <source-branch> [--strategy auto|manual]

# Delete branch
nixai version-control branch delete <branch-name> [--force]
```

### Team Collaboration

```bash
# Manage configuration teams
nixai version-control team create <team-name> [--description "Team description"]

# Add team members
nixai version-control team add-member <team-name> <username> [--role admin|editor|viewer]

# List team members
nixai version-control team list [team-name] [--detailed]

# Team permissions management
nixai version-control team permissions <team-name> [--branch branch-name] [--action grant|revoke]
```

## Getting Started

### Initialize Configuration Repository

Set up version control for your NixOS configuration:

```bash
# Initialize in default location (/etc/nixos)
nixai version-control init

# Initialize with remote repository
nixai version-control init \
  --remote-url git@github.com:company/nixos-configs \
  --branch main \
  --setup-hooks
```

The initialization process:
1. Creates Git repository in configuration directory
2. Sets up nixai-specific Git hooks
3. Creates initial commit with current configuration
4. Configures remote repository (if specified)
5. Sets up branching strategy and workflows

### Basic Workflow

```bash
# Check current status
nixai version-control status

# Make configuration changes
# Edit /etc/nixos/configuration.nix

# Review changes
nixai version-control diff

# Commit changes
nixai version-control commit --message "Enable SSH service and configure firewall"

# View history
nixai version-control history --limit 5
```

## Advanced Features

### Configuration Validation

Automatic validation before commits:

```bash
# Enable pre-commit validation
nixai version-control config set validation.pre-commit true

# Configure validation rules
nixai version-control config set validation.rules \
  "syntax-check,security-audit,performance-check"

# Manual validation
nixai version-control validate [--fix-issues]
```

Example validation output:
```text
Configuration Validation Results:
✓ Syntax check passed
✓ Security audit passed
⚠ Performance check: 2 warnings found
  - Consider enabling zram for better memory management
  - Large log retention period may impact disk usage
✗ Custom rules: 1 error found
  - SSH root login is enabled (security policy violation)

Validation: Failed (1 error, 2 warnings)
```

### Branching Strategies

#### Feature Branch Workflow

```bash
# Create feature branch
nixai version-control branch create feature/add-web-server

# Work on feature
nixai version-control commit --message "Add nginx configuration"
nixai version-control commit --message "Configure SSL certificates"

# Switch back to main and merge
nixai version-control branch switch main
nixai version-control branch merge feature/add-web-server

# Clean up
nixai version-control branch delete feature/add-web-server
```

#### Environment Branching

```bash
# Create environment-specific branches
nixai version-control branch create development --from main
nixai version-control branch create staging --from main
nixai version-control branch create production --from main

# Deploy to environments
nixai version-control deploy development --target dev-servers
nixai version-control deploy staging --target staging-servers
nixai version-control deploy production --target prod-servers
```

### Collaborative Workflows

#### Pull Request Workflow

```bash
# Create feature branch and push
nixai version-control branch create feature/security-hardening
# Make changes...
nixai version-control commit --message "Implement security hardening"
nixai version-control push origin feature/security-hardening

# Create pull request (integrates with GitHub/GitLab)
nixai version-control pr create \
  --title "Security hardening improvements" \
  --description "Implements firewall rules, SSH hardening, and audit logging" \
  --reviewers @security-team
```

#### Code Review Integration

```bash
# Review configuration changes
nixai version-control review <commit-id> [--interactive]

# Add review comments
nixai version-control review comment \
  --file configuration.nix \
  --line 42 \
  --message "Consider using more restrictive firewall rules"

# Approve changes
nixai version-control review approve <commit-id>
```

### Team Management

#### Create and Manage Teams

```bash
# Create development team
nixai version-control team create dev-team \
  --description "Development team for infrastructure configurations"

# Add team members with roles
nixai version-control team add-member dev-team alice --role admin
nixai version-control team add-member dev-team bob --role editor
nixai version-control team add-member dev-team charlie --role viewer

# Set branch permissions
nixai version-control team permissions dev-team \
  --branch main \
  --action grant \
  --permissions read,review

nixai version-control team permissions dev-team \
  --branch "feature/*" \
  --action grant \
  --permissions read,write,create,delete
```

#### Role-Based Access Control

| Role | Permissions |
|------|-------------|
| **Admin** | Full access to all branches, team management, settings |
| **Editor** | Read/write access to assigned branches, create feature branches |
| **Viewer** | Read-only access to repositories and history |
| **Reviewer** | Editor permissions plus approve/reject pull requests |

## Configuration Management

### Repository Configuration

Configure version control behavior:

```yaml
version_control:
  # Repository settings
  repository:
    path: "/etc/nixos"
    remote_url: "git@github.com:company/nixos-configs"
    default_branch: "main"
    auto_push: false
  
  # Commit settings
  commits:
    require_message: true
    message_template: "[{type}] {description}\n\n{details}"
    sign_commits: true
    gpg_key_id: "your-gpg-key-id"
  
  # Validation settings
  validation:
    pre_commit: true
    pre_push: true
    rules:
      - syntax-check
      - security-audit
      - performance-check
      - custom-policies
  
  # Branching strategy
  branching:
    strategy: "git-flow"  # git-flow, github-flow, gitlab-flow
    protect_main: true
    require_reviews: true
    auto_delete_merged: true
  
  # Team settings
  teams:
    default_permissions:
      main: "read"
      develop: "read,write"
      feature: "read,write,create"
    review_requirements:
      main: 2
      develop: 1
      feature: 0
```

### Hooks and Automation

#### Pre-commit Hooks

```bash
# Enable standard hooks
nixai version-control hooks enable pre-commit \
  --hooks syntax-check,security-scan,format-check

# Custom hook configuration
nixai version-control hooks configure pre-commit \
  --hook custom-policy-check \
  --script "/usr/local/bin/nixos-policy-check"
```

#### CI/CD Integration

```yaml
# .nixai/workflows/ci.yml
name: NixOS Configuration CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: nixai/setup-action@v1
      - run: nixai version-control validate --strict
      - run: nixai build test --configuration test-environment.nix
  
  deploy:
    if: github.ref == 'refs/heads/main'
    needs: validate
    runs-on: ubuntu-latest
    steps:
      - run: nixai fleet deploy --config-source git --git-ref $GITHUB_SHA
```

## Integration with Fleet Management

Combine version control with fleet management for comprehensive infrastructure automation:

```bash
# Deploy specific commit to fleet
nixai fleet deploy --config-source version-control \
  --commit-id abc123 \
  --target production

# Deploy branch to specific machines
nixai fleet deploy --config-source version-control \
  --branch feature/new-service \
  --machines web-01,web-02

# Rollback fleet deployment
nixai fleet rollback --to-commit def456
```

## Backup and Recovery

### Configuration Snapshots

```bash
# Create configuration snapshot
nixai version-control snapshot create \
  --name "pre-upgrade-$(date +%Y%m%d)" \
  --description "Snapshot before system upgrade"

# List snapshots
nixai version-control snapshot list [--sort-by date]

# Restore from snapshot
nixai version-control snapshot restore pre-upgrade-20240703
```

### Repository Backup

```bash
# Backup repository
nixai version-control backup create \
  --output /backup/nixos-configs-$(date +%Y%m%d).tar.gz \
  --include-history \
  --compress

# Restore repository
nixai version-control backup restore \
  --source /backup/nixos-configs-20240703.tar.gz \
  --target /etc/nixos
```

## Security Features

### Commit Signing

```bash
# Setup GPG signing
nixai version-control security setup-gpg \
  --key-id your-gpg-key-id \
  --auto-sign

# Verify commit signatures
nixai version-control security verify-signatures [--all]

# Show signature status
nixai version-control history --show-signatures
```

### Access Auditing

```bash
# View access logs
nixai version-control audit access [--user username] [--since "24h ago"]

# Configuration change audit
nixai version-control audit changes \
  --sensitive-files "hardware-configuration.nix,secrets.nix"

# Security audit report
nixai version-control audit security-report \
  --output security-audit-$(date +%Y%m%d).json
```

## Troubleshooting

### Common Issues

**Merge conflicts:**
```bash
# Show conflict details
nixai version-control status --conflicts

# Resolve conflicts interactively
nixai version-control merge resolve --interactive

# Abort merge and reset
nixai version-control merge abort
```

**Repository corruption:**
```bash
# Check repository integrity
nixai version-control fsck [--repair]

# Rebuild from backup
nixai version-control backup restore --force

# Reset to last known good state
nixai version-control reset --hard origin/main
```

**Permission issues:**
```bash
# Fix repository permissions
nixai version-control permissions fix

# Reset team permissions
nixai version-control team permissions reset --team dev-team

# Regenerate access tokens
nixai version-control auth regenerate-tokens
```

### Debug Mode

Enable comprehensive debugging:

```bash
nixai version-control --debug commit --verbose
```

Debug output includes:
- Git operation details
- Hook execution logs
- Validation process steps
- Team permission checks

## Best Practices

1. **Commit Hygiene**: Use clear, descriptive commit messages
2. **Branch Strategy**: Choose appropriate branching strategy for team size
3. **Code Review**: Require reviews for production changes
4. **Testing**: Validate configurations before committing
5. **Security**: Use commit signing and access controls
6. **Backup**: Regular repository backups and snapshots
7. **Documentation**: Maintain clear change documentation
8. **Automation**: Integrate with CI/CD for automatic validation

For more detailed examples and workflow patterns, see the [Version Control Guide](../examples/version-control/) and [Team Collaboration Patterns](../guides/team-collaboration.md).