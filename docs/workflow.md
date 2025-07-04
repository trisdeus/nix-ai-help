# nixai workflow - Workflow Management and Automation

The `nixai workflow` command provides workflow automation, task orchestration, and process management for NixOS administration tasks with real action execution and authentic system operations.

## Usage

```bash
nixai workflow [command] [options]
```

## Commands

```bash
# List available workflows
nixai workflow list [--category system|deployment|maintenance] [--status active|completed]

# Create new workflow
nixai workflow create <name> [--template basic|advanced] [--description "description"]

# Execute workflow
nixai workflow run <workflow-name> [--parameters key=value] [--async]

# Show workflow status
nixai workflow status [workflow-name] [--detailed] [--follow]

# Workflow history and logs
nixai workflow history [--limit 20] [--filter succeeded|failed|running]

# Schedule workflow execution
nixai workflow schedule <workflow-name> [--cron "0 2 * * *"] [--timezone UTC]
```

## Features

### Workflow Automation
- **Real Action Execution**: Authentic file operations, package management, and service control
- **Pre-built workflow templates** for common NixOS tasks
- **Custom workflow creation** with actual system command execution
- **Task orchestration and sequencing** with real progress tracking
- **Conditional execution logic** based on actual system state

### Process Management
- **Real Command Execution**: Using `nix-env`, `systemctl`, and file operations
- **Workflow execution monitoring** with authentic progress feedback
- **Progress tracking and reporting** from actual system operations
- **Error handling and recovery** with real exit codes and system responses
- **Parallel task execution** with proper process management

### Scheduling and Triggers
- Cron-based scheduling
- Event-driven triggers
- Manual execution control
- Dependency-based execution

## Examples

```bash
# List all available workflows
nixai workflow list --category maintenance

# Create a system update workflow
nixai workflow create weekly-update --template maintenance

# Run system health check workflow
nixai workflow run health-check --async

# Schedule nightly backup workflow
nixai workflow schedule backup-system --cron "0 2 * * *"
```

For workflow automation examples and custom workflow creation, see the main documentation.