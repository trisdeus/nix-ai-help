# MCP Server TCP Port Configuration - Implementation Summary

## ✅ COMPLETED

### 1. Configuration Structure Updated
- Added `MCPPort` field to `MCPServerConfig` struct in `internal/config/config.go`
- Updated both YAML and JSON tags: `yaml:"mcp_port" json:"mcp_port"`

### 2. Default Configuration Values Updated
- **Embedded Configuration**: Updated `EmbeddedDefaultConfig` to include `mcp_port: 39847`
- **Default Configuration File**: Updated `configs/default.yaml` to include `mcp_port: 39847`
- **DefaultUserConfig Function**: Added `MCPPort: 39847` to the default configuration

### 3. CLI Integration Updated
- Updated `internal/cli/commands.go` to handle `mcp_port` configuration using `MCPServer.MCPPort`
- Updated `internal/cli/direct_commands.go` to handle `mcp_port` configuration using `MCPServer.MCPPort`
- Both get and set operations for `mcp_port` now work correctly

### 4. NixOS Module Integration Updated
- **NixOS Module**: Added `mcpPort` option to `nix/modules/nixos.nix` with default value 39847
- **Home Manager Module**: Added `mcpPort` option to `nix/modules/home-manager.nix` with default value 39847
- **Configuration Generation**: Both modules now include `mcp_port` field in generated config files

### 5. Documentation Updated
- **NixOS Module README**: Updated port references from 8082 to 39847
- **Flake Integration Guide**: Updated port references and added explanation of TCP vs Unix socket usage
- **Documentation Examples**: Updated all configuration examples to use port 39847

### 6. Configuration Fixed
- Fixed corrupted YAML in embedded configuration
- Fixed indentation issues in embedded configuration
- Verified embedded configuration loads correctly with `mcp_port: 39847`

### 7. Testing Verified
- ✅ Configuration loading works correctly
- ✅ CLI `config get mcp_port` returns correct value (39847)
- ✅ CLI `config set mcp_port <value>` works correctly
- ✅ Embedded configuration parsing works
- ✅ User configuration parsing works
- ✅ All configuration tests pass

## 📋 NEXT STEPS (Not Yet Implemented)

### 1. MCP Server TCP Implementation
- **Priority**: HIGH
- **Description**: Implement `StartTCP` method in `internal/mcp/server.go`
- **Details**: Create TCP server that listens on `cfg.MCPServer.MCPPort` instead of Unix socket

### 2. Server Startup Logic Update
- **Priority**: HIGH  
- **Description**: Modify main server startup to use TCP instead of Unix sockets
- **Files**: Update server initialization to call `StartTCP` instead of `Start`

### 3. Bridge Script Updates
- **Priority**: MEDIUM
- **Description**: Update `scripts/mcp-bridge.sh` to support TCP connections
- **Details**: Add TCP connection support alongside existing Unix socket support

### 4. CLI Status Commands Update
- **Priority**: MEDIUM
- **Description**: Update status checking commands to work with TCP ports
- **Details**: Check TCP port availability instead of socket file existence

### 5. VS Code Integration Update
- **Priority**: LOW
- **Description**: Update VS Code integration scripts if needed
- **Details**: Ensure VS Code MCP extension can connect to TCP port

### 6. Tests Addition
- **Priority**: LOW
- **Description**: Add tests specifically for TCP MCP server functionality
- **Details**: Test TCP connection, port binding, error handling

## 🔧 Configuration Summary

| Configuration Type | HTTP Port | MCP TCP Port | MCP Socket Path |
|-------------------|-----------|--------------|-----------------|
| NixOS System | 8080 | 39847 | `/run/nixai/mcp.sock` |
| Home Manager | 8081 | 39847 | `$HOME/.local/share/nixai/mcp.sock` |

## 🚀 Usage

After system rebuild/home-manager switch, users can:

```bash
# Check MCP port configuration
nixai config get mcp_port

# Change MCP port if needed  
nixai config set mcp_port 12345

# Use in NixOS configuration
services.nixai.mcp.mcpPort = 39847;

# Use in Home Manager configuration  
services.nixai.mcp.mcpPort = 39847;
```

## 📝 Notes

- **Port 39847** was chosen as a non-standard port to avoid conflicts
- **Backward Compatibility**: Unix socket support remains in place during transition
- **System vs User**: System config takes precedence when both exist
- **TCP Benefits**: Better cross-platform support, easier debugging, no file permission issues
