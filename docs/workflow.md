# nixai workflow - Workflow Management and Automation

The `nixai workflow` command provides workflow automation, task orchestration, and process management for NixOS administration tasks.

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
- Pre-built workflow templates
- Custom workflow creation
- Task orchestration and sequencing
- Conditional execution logic

### Process Management
- Workflow execution monitoring
- Progress tracking and reporting
- Error handling and recovery
- Parallel task execution

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