# Making Unix Sockets More Robust

## 🔍 Current Unix Socket Issues in nixai

### 1. **File System Dependencies**
- **Issue**: Socket files can have permission problems, directory access issues
- **Current Impact**: Users may get "permission denied" when connecting to `/tmp/nixai-mcp.sock`

### 2. **Stale Socket Files**
- **Issue**: If server crashes, socket file remains and prevents restart
- **Current Mitigation**: `os.Remove(socketPath)` before creating new socket
- **Problem**: Race conditions if multiple processes try to create socket

### 3. **Path Length Limitations**
- **Issue**: Unix socket paths are limited to ~100 characters on most systems
- **Current Risk**: Long paths like `/home/very/long/username/.local/share/nixai/mcp.sock` may fail

### 4. **Cross-Platform Compatibility**
- **Issue**: Windows has poor Unix socket support
- **Current Impact**: nixai MCP server may not work reliably on Windows

## 🛠️ Robustness Improvements for Unix Sockets

### 1. **Enhanced Socket File Management**

```go
// Enhanced socket cleanup with locking
func createSocketSafely(socketPath string) (net.Listener, error) {
    // Create parent directory if it doesn't exist
    dir := filepath.Dir(socketPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create socket directory: %v", err)
    }
    
    // Use file locking to prevent race conditions
    lockFile := socketPath + ".lock"
    lock, err := createLockFile(lockFile)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire socket lock: %v", err)
    }
    defer lock.Close()
    
    // Check if socket is already in use
    if isSocketInUse(socketPath) {
        return nil, fmt.Errorf("socket already in use: %s", socketPath)
    }
    
    // Remove stale socket file
    if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
        return nil, fmt.Errorf("failed to remove stale socket: %v", err)
    }
    
    // Create socket with proper permissions
    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        return nil, err
    }
    
    // Set socket file permissions (readable/writable by owner and group)
    if err := os.Chmod(socketPath, 0660); err != nil {
        listener.Close()
        return nil, fmt.Errorf("failed to set socket permissions: %v", err)
    }
    
    return listener, nil
}

func isSocketInUse(socketPath string) bool {
    conn, err := net.Dial("unix", socketPath)
    if err == nil {
        conn.Close()
        return true
    }
    return false
}
```

### 2. **Abstract Sockets (Linux-specific)**

```go
// Use abstract sockets on Linux to avoid filesystem issues
func createAbstractSocket(name string) (net.Listener, error) {
    // Abstract socket names start with null byte
    socketPath := "\x00nixai-mcp-" + name
    return net.Listen("unix", socketPath)
}
```

### 3. **Health Checking and Recovery**

```go
type RobustMCPServer struct {
    *MCPServer
    healthChecker *time.Ticker
    socketPath    string
}

func (r *RobustMCPServer) startHealthChecker() {
    r.healthChecker = time.NewTicker(30 * time.Second)
    go func() {
        defer r.healthChecker.Stop()
        for {
            select {
            case <-r.healthChecker.C:
                if !r.isHealthy() {
                    r.logger.Warn("Socket health check failed, attempting recovery")
                    r.recover()
                }
            case <-r.ctx.Done():
                return
            }
        }
    }()
}

func (r *RobustMCPServer) isHealthy() bool {
    // Check if socket file exists and is accessible
    info, err := os.Stat(r.socketPath)
    if err != nil {
        return false
    }
    
    // Check if it's actually a socket
    return info.Mode()&os.ModeSocket != 0
}

func (r *RobustMCPServer) recover() {
    r.logger.Info("Attempting socket recovery")
    r.Stop()
    time.Sleep(1 * time.Second)
    if err := r.Start(r.socketPath); err != nil {
        r.logger.Error(fmt.Sprintf("Socket recovery failed: %v", err))
    }
}
```

### 4. **Enhanced Bridge Script**

```bash
#!/bin/bash
# Robust MCP Bridge Script

set -euo pipefail

SOCKET_PATH="${NIXAI_MCP_SOCKET:-/tmp/nixai-mcp.sock}"
MAX_RETRIES=5
RETRY_DELAY=2
HEALTH_CHECK_TIMEOUT=5

# Enhanced server health check
check_server_health() {
    local socket_path="$1"
    
    # Check if socket file exists
    if [[ ! -S "$socket_path" ]]; then
        return 1
    fi
    
    # Check if socket is responsive (with timeout)
    if timeout "$HEALTH_CHECK_TIMEOUT" socat /dev/null "UNIX-CONNECT:$socket_path" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Cleanup stale sockets
cleanup_stale_socket() {
    local socket_path="$1"
    
    if [[ -e "$socket_path" && ! -S "$socket_path" ]]; then
        echo "Removing stale non-socket file: $socket_path" >&2
        rm -f "$socket_path"
    fi
}

# Wait for server to become available
wait_for_server() {
    local socket_path="$1"
    local max_wait=30
    local wait_time=0
    
    while [[ $wait_time -lt $max_wait ]]; do
        if check_server_health "$socket_path"; then
            return 0
        fi
        
        sleep 1
        ((wait_time++))
    done
    
    return 1
}

# Enhanced connection with better error handling
connect_with_retry() {
    local attempt=1
    
    # Cleanup any stale sockets first
    cleanup_stale_socket "$SOCKET_PATH"
    
    while [[ $attempt -le $MAX_RETRIES ]]; do
        echo "Connection attempt $attempt/$MAX_RETRIES..." >&2
        
        if check_server_health "$SOCKET_PATH"; then
            echo "Connecting to MCP server..." >&2
            exec socat STDIO "UNIX-CONNECT:$SOCKET_PATH"
            return 0
        fi
        
        if [[ $attempt -eq 1 ]]; then
            echo "MCP server not available, trying to start it..." >&2
            nixai mcp-server start --daemon 2>/dev/null || true
            
            if wait_for_server "$SOCKET_PATH"; then
                echo "MCP server started successfully" >&2
                exec socat STDIO "UNIX-CONNECT:$SOCKET_PATH"
                return 0
            fi
        fi
        
        echo "Attempt $attempt failed, retrying in ${RETRY_DELAY}s..." >&2
        sleep $RETRY_DELAY
        ((attempt++))
    done
    
    echo "Failed to connect after $MAX_RETRIES attempts" >&2
    echo "Troubleshooting:" >&2
    echo "  - Check if nixai is installed: which nixai" >&2
    echo "  - Check server status: nixai mcp-server status" >&2
    echo "  - Check logs: journalctl --user -u nixai-mcp -f" >&2
    exit 1
}

# Main execution
connect_with_retry
```

### 5. **Fallback Mechanism**

```go
// Fallback to TCP if Unix socket fails
func (m *MCPServer) StartWithFallback(socketPath string, tcpPort int) error {
    // Try Unix socket first
    if err := m.Start(socketPath); err != nil {
        m.logger.Warn(fmt.Sprintf("Unix socket failed (%v), falling back to TCP", err))
        return m.StartTCP(fmt.Sprintf("localhost:%d", tcpPort))
    }
    return nil
}
```

### 6. **Socket Permissions and Security**

```go
func setSocketPermissions(socketPath string, ownerUID, groupGID int) error {
    // Set ownership
    if err := os.Chown(socketPath, ownerUID, groupGID); err != nil {
        return fmt.Errorf("failed to set socket ownership: %v", err)
    }
    
    // Set permissions (owner: rw, group: rw, others: none)
    if err := os.Chmod(socketPath, 0660); err != nil {
        return fmt.Errorf("failed to set socket permissions: %v", err)
    }
    
    return nil
}

// Create socket in secure directory
func createSecureSocket(socketPath string) (net.Listener, error) {
    dir := filepath.Dir(socketPath)
    
    // Create directory with restricted permissions
    if err := os.MkdirAll(dir, 0750); err != nil {
        return nil, err
    }
    
    // Remove stale socket
    os.Remove(socketPath)
    
    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        return nil, err
    }
    
    // Set socket permissions
    if err := os.Chmod(socketPath, 0660); err != nil {
        listener.Close()
        return nil, err
    }
    
    return listener, nil
}
```

## 📊 Comparison: Unix Sockets vs TCP

| Aspect | Unix Sockets | TCP Sockets |
|--------|--------------|-------------|
| **Performance** | ✅ Faster (no network stack) | ⚠️ Slightly slower |
| **Security** | ✅ Filesystem permissions | ⚠️ Network accessible |
| **Reliability** | ❌ File system dependencies | ✅ More robust |
| **Cross-platform** | ❌ Windows issues | ✅ Universal support |
| **Debugging** | ❌ Hard to inspect | ✅ Network tools available |
| **Permission Issues** | ❌ Common | ✅ Rare |
| **Cleanup** | ❌ Manual required | ✅ Automatic |

## 🎯 Recommended Approach

### **Hybrid Strategy** (Best of Both Worlds)

1. **Default to TCP** (port 39847) for reliability
2. **Unix socket as option** for performance-critical local use
3. **Automatic fallback** from Unix socket to TCP on failure

```go
func (m *MCPServer) StartRobust(config *config.MCPServerConfig) error {
    // Try methods in order of preference
    methods := []func() error{
        func() error { return m.StartTCP(fmt.Sprintf("%s:%d", config.Host, config.MCPPort)) },
        func() error { return m.Start(config.SocketPath) },
    }
    
    var lastErr error
    for i, method := range methods {
        if err := method(); err != nil {
            lastErr = err
            m.logger.Warn(fmt.Sprintf("Method %d failed: %v", i+1, err))
            continue
        }
        return nil
    }
    
    return fmt.Errorf("all connection methods failed, last error: %v", lastErr)
}
```

## 🔧 Implementation Priority

1. **HIGH**: Implement robust socket cleanup and error handling
2. **MEDIUM**: Add health checking and recovery mechanisms  
3. **LOW**: Add abstract socket support for Linux
4. **COMPLETED**: TCP fallback (already implemented via mcpPort configuration)

The TCP implementation we just completed addresses most Unix socket reliability issues, making it the preferred approach for production use!
