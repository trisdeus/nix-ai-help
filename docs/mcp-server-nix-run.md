# Using MCP Server with nix run

The nixai MCP (Model Context Protocol) server can be used with `nix run` using the ephemeral mode designed specifically for this purpose.

## Problem with Standard MCP Server

The MCP server is designed to run as a persistent daemon process, which doesn't work well with `nix run` because:

1. **Persistent Processes**: MCP server needs to run continuously in the background
2. **`nix run` is Ephemeral**: Processes exit when the command finishes
3. **State Management**: MCP server needs persistent socket files and process management
4. **System Integration**: VS Code/Neovim integration expects a persistent service

## Solution: Ephemeral Mode

nixai provides an `--ephemeral` mode specifically designed for `nix run` usage:

### Basic Usage

```bash
# Start MCP server in ephemeral mode with nix run
nix run github:olafkfreund/nix-ai-help -- mcp-server start --ephemeral

# Short form
nix run github:olafkfreund/nix-ai-help -- mcp-server start -e
```

### Features of Ephemeral Mode

1. **Automatic Cleanup**: Temporary socket files are automatically removed on exit
2. **Unique Socket Paths**: Uses PID-based socket paths to avoid conflicts
3. **Graceful Shutdown**: Handles Ctrl+C and termination signals properly
4. **No Daemon Mode**: Runs in foreground for better `nix run` compatibility

### Example Session

```bash
# Terminal 1: Start the MCP server
$ nix run github:olafkfreund/nix-ai-help -- mcp-server start -e

🚀 Starting MCP Server

Running in ephemeral mode (nix run compatible)
Initializing MCP server... Starting Unix socket server... ✅ Unix socket server started successfully

✅ Unix socket server started successfully

HTTP Server: http://localhost:3001
Unix Socket: /tmp/nixai-mcp-12345.sock

Ephemeral mode: Press Ctrl+C to stop the server
Socket will be automatically cleaned up on exit

# Server is now running...
# Press Ctrl+C to stop and cleanup
```

### Integration with Editors

For editor integration, you'll need to:

1. **Start the MCP server** in ephemeral mode in a separate terminal
2. **Configure your editor** to use the displayed socket path
3. **Use the server** while it's running
4. **Stop with Ctrl+C** when done

### Comparison with Daemon Mode

| Feature | Daemon Mode | Ephemeral Mode |
|---------|-------------|----------------|
| Background Process | ✅ | ❌ |
| Automatic Startup | ✅ | ❌ |
| `nix run` Compatible | ❌ | ✅ |
| Auto Cleanup | ❌ | ✅ |
| Editor Integration | ✅ (persistent) | ✅ (while running) |
| Process Management | Complex | Simple |

### When to Use Each Mode

**Use Ephemeral Mode (`-e`) when:**
- Running with `nix run`
- Testing or development
- Temporary usage
- Don't want persistent processes

**Use Daemon Mode (`-d`) when:**
- Permanent installation
- Production usage
- Editor integration
- Want automatic startup

### Advanced Usage

```bash
# Custom socket path in ephemeral mode
nix run github:olafkfreund/nix-ai-help -- mcp-server start -e --socket-path /tmp/my-nixai.sock

# Check if MCP server is running (from another terminal)
nix run github:olafkfreund/nix-ai-help -- mcp-server status

# Query the MCP server directly
nix run github:olafkfreund/nix-ai-help -- mcp-server query "how to configure nginx"
```

### Troubleshooting

**Socket Permission Issues:**
```bash
# Ensure /tmp directory is writable
ls -la /tmp/

# Use custom socket path if needed
nix run github:olafkfreund/nix-ai-help -- mcp-server start -e --socket-path ~/nixai.sock
```

**Port Conflicts:**
```bash
# Check if ports are in use
netstat -tlnp | grep 3001

# The ephemeral mode handles this automatically with unique socket paths
```

**Editor Connection Issues:**
1. Make sure the MCP server is running in ephemeral mode
2. Use the exact socket path shown in the server output
3. The socket only exists while the server is running

This ephemeral mode makes nixai's MCP server fully compatible with `nix run` while maintaining all the functionality needed for documentation queries and editor integration.