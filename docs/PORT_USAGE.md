# Port Usage Documentation

This document tracks port usage in the nixai project to avoid conflicts.

## Reserved Ports

### Web Interface
- **Port 34567** - Default web interface port for enhanced server
  - Used by: `nixai web` command
  - Services: Dashboard, Builder, Fleet Management, Teams, Version Control
  - WebSocket: Real-time updates and collaboration

### Prohibited Ports
- **Port 8080** - ALREADY IN USE by other services
  - DO NOT USE this port for any nixai services
  - Commonly used by development servers and other applications

### MCP Server Ports
- **Port Range 12000-12999** - MCP server instances
  - Default: Auto-assigned within this range
  - Configurable via nixai configuration

### AI Service Ports
- **Port 11434** - Default Ollama server port (external dependency)
- **Port Range 13000-13999** - Reserved for future AI service integrations

## Port Configuration

### Web Interface Port
```bash
# Default port
nixai web

# Custom port
nixai web --port 35000
```

### Environment Variables
```bash
# Override default web port
export NIXAI_WEB_PORT=35000

# MCP server port range
export NIXAI_MCP_PORT_START=12000
export NIXAI_MCP_PORT_END=12999
```

## Best Practices

1. **Always check port availability** before assigning new services
2. **Use port ranges** for similar services (e.g., MCP servers)
3. **Document new port assignments** in this file
4. **Avoid common ports** like 8080, 3000, 5000, etc.
5. **Use high port numbers** (>10000) to avoid system conflicts

## Testing Port Conflicts

```bash
# Check if port is in use
lsof -i :8080

# Find available ports in range
nmap -p 34000-35000 localhost
```

## Port Allocation Strategy

- **30000-39999**: Web interfaces and frontends
- **40000-49999**: API services
- **50000-59999**: Background services
- **60000-65535**: Testing and development

This ensures clear separation and reduces conflict potential.
