# AI-Controlled Command Execution Implementation Plan

## Overview

This document outlines the implementation plan for adding AI-controlled command execution capabilities to nixai. This feature will allow AI agents to safely execute system commands, manage packages, edit configurations, and perform system operations based on natural language user requests.

## Goals

- **User Efficiency**: Allow users to perform complex NixOS operations through natural language
- **Safety First**: Implement robust security and permission systems
- **AI Integration**: Seamlessly integrate command execution with existing AI agents
- **User Control**: Maintain user oversight and confirmation for critical operations

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Interface                          │
│                    (CLI, TUI, Web, MCP)                        │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                     AI Agents                                   │
│  ┌─────────────┐ ┌──────────────┐ ┌─────────────────────────┐   │
│  │   Package   │ │ Configuration│ │   System Management     │   │
│  │   Agent     │ │    Agent     │ │       Agent             │   │
│  └─────────────┘ └──────────────┘ └─────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                AI Function System                               │
│              execute_command Function                           │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                Safe Executor                                    │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐   │
│  │   Permission    │ │   Sudo Manager  │ │   Audit Logger  │   │
│  │   Manager       │ │                 │ │                 │   │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘   │
└─────────────────────┬───────────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────────┐
│                System Commands                                  │
│        (nix, nixos-rebuild, systemctl, etc.)                   │
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Foundation & Security Framework (Week 1)

#### 1.1 Core Execution Framework
**Files to create:**
- `internal/execution/safe_executor.go`
- `internal/execution/types.go`
- `internal/execution/validation.go`
- `internal/security/permission_manager.go`
- `internal/security/audit_logger.go`

**Key Components:**
```go
// Core execution types
type CommandRequest struct {
    Command      string            `json:"command"`
    Args         []string          `json:"args"`
    RequiresSudo bool             `json:"requiresSudo"`
    WorkingDir   string            `json:"workingDir,omitempty"`
    Environment  map[string]string `json:"environment,omitempty"`
    Description  string            `json:"description"`
    Category     string            `json:"category"`
    DryRun       bool             `json:"dryRun,omitempty"`
}

type ExecutionResult struct {
    Success    bool              `json:"success"`
    ExitCode   int               `json:"exitCode"`
    Output     string            `json:"output"`
    Error      string            `json:"error,omitempty"`
    Duration   time.Duration     `json:"duration"`
    Command    string            `json:"command"`
    Timestamp  time.Time         `json:"timestamp"`
}

type SafeExecutor struct {
    permissionManager *PermissionManager
    auditLogger       *AuditLogger
    config           *config.ExecutionConfig
    logger           *logger.Logger
}
```

**Permission System:**
```go
type PermissionManager struct {
    allowedCommands    map[string]CommandPermission
    commandCategories  map[string][]string
    userConfig         *config.UserConfig
    restrictionRules   []RestrictionRule
}

type CommandPermission struct {
    Command              string
    AllowedArgs          []string
    ForbiddenArgs        []string
    RequiresConfirmation bool
    RequiresSudo         bool
    Category             string
    MaxExecutionTime     time.Duration
    AllowedDirectories   []string
}
```

**Deliverables:**
- [ ] Safe command execution framework
- [ ] Permission validation system
- [ ] Audit logging for all command executions
- [ ] Configuration schema for execution settings
- [ ] Basic command whitelist (nix, git, ls, cat, etc.)
- [ ] Unit tests for core execution logic

#### 1.2 Configuration Integration
**Files to update:**
- `configs/default.yaml`
- `internal/config/config.go`

**Configuration Schema:**
```yaml
execution:
  enabled: true
  confirmationRequired: true
  dryRunDefault: false
  maxExecutionTime: "10m"
  
  categories:
    package:
      commands: ["nix-env", "nix", "nix-shell", "nix-store"]
      requiresConfirmation: false
      maxExecutionTime: "5m"
    
    system:
      commands: ["nixos-rebuild", "systemctl", "journalctl"]
      requiresConfirmation: true
      requiresSudo: true
      maxExecutionTime: "15m"
    
    configuration:
      commands: ["nvim", "vim", "nano", "cp", "mv"]
      requiresConfirmation: true
      allowedDirectories: ["/etc/nixos", "/home"]
    
    development:
      commands: ["nix develop", "nix flake", "nix build"]
      requiresConfirmation: false
      maxExecutionTime: "30m"

  security:
    auditLogging: true
    auditLogPath: "/var/log/nixai/commands.log"
    maxConcurrentCommands: 3
    allowEnvironmentVariables: ["PATH", "NIX_PATH", "HOME"]
    
  restrictions:
    forbiddenCommands: ["rm", "dd", "mkfs", "fdisk"]
    forbiddenPaths: ["/boot", "/sys", "/proc"]
    requireSudoFor: ["nixos-rebuild", "systemctl"]
```

### Phase 2: AI Function Integration (Week 2)

#### 2.1 Execute Command AI Function
**Files to create:**
- `internal/ai/function/execute_command.go`
- `internal/ai/function/execute_command_test.go`

**Function Implementation:**
```go
type ExecuteCommandFunction struct {
    executor     *execution.SafeExecutor
    logger       *logger.Logger
    interactive  bool
}

func (ecf *ExecuteCommandFunction) GetDefinition() FunctionDefinition {
    return FunctionDefinition{
        Name:        "execute_command",
        Description: "Execute system commands safely with permission validation",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "command": map[string]interface{}{
                    "type":        "string",
                    "description": "The command to execute (e.g., 'nix', 'nixos-rebuild')",
                },
                "args": map[string]interface{}{
                    "type":        "array",
                    "description": "Command arguments",
                    "items":       map[string]interface{}{"type": "string"},
                },
                "description": map[string]interface{}{
                    "type":        "string", 
                    "description": "Human-readable description of what this command does",
                },
                "category": map[string]interface{}{
                    "type":        "string",
                    "enum":        []string{"package", "system", "configuration", "development"},
                    "description": "Command category for permission validation",
                },
                "requiresSudo": map[string]interface{}{
                    "type":        "boolean",
                    "description": "Whether this command requires sudo privileges",
                },
                "dryRun": map[string]interface{}{
                    "type":        "boolean", 
                    "description": "If true, show what would be executed without running",
                },
            },
            "required": []string{"command", "description", "category"},
        },
    }
}
```

#### 2.2 Agent Enhancement
**Files to update:**
- `internal/ai/agent/ask_agent.go`
- `internal/ai/agent/package_repo_agent.go`
- `internal/ai/agent/config_agent.go`
- `internal/ai/agent/build_agent.go`

**Enhanced Agent Prompts:**
```go
const EnhancedPackageAgentPrompt = `
You are a NixOS package management assistant with command execution capabilities.

Available functions:
- search_packages: Search for packages
- execute_command: Execute package management commands

When users request package operations:
1. Explain what you'll do
2. Use execute_command function for actual operations
3. Provide helpful context and next steps

Examples:

User: "Install firefox"
Response: I'll install Firefox for you. I can either:
1. Add it to your system configuration (recommended)
2. Install it temporarily with nix-env

Let me add it to your system packages:

{
  "function": "execute_command",
  "parameters": {
    "command": "nix-env",
    "args": ["-iA", "nixpkgs.firefox"],
    "description": "Install Firefox browser",
    "category": "package"
  }
}

User: "Run htop"
Response: I'll run htop for you:

{
  "function": "execute_command", 
  "parameters": {
    "command": "nix",
    "args": ["run", "nixpkgs#htop"],
    "description": "Run htop system monitor",
    "category": "package"
  }
}
`
```

**Deliverables:**
- [ ] execute_command AI function with full validation
- [ ] Enhanced agent prompts for command execution
- [ ] Integration tests with existing AI system
- [ ] Function registration in AI manager
- [ ] Documentation for AI function usage

### Phase 3: Sudo Management & Security (Week 3)

#### 3.1 Sudo Management System
**Files to create:**
- `internal/security/sudo_manager.go`
- `internal/security/password_manager.go`
- `internal/security/elevation_strategy.go`

**Sudo Manager Implementation:**
```go
type SudoManager struct {
    strategy        ElevationStrategy
    passwordCache   *PasswordCache
    sessionManager  *SessionManager
    logger          *logger.Logger
}

type ElevationStrategy interface {
    Elevate(ctx context.Context, cmd string, args []string) (*ExecutionResult, error)
    RequiresPassword() bool
    ValidateAccess() error
}

// Different elevation strategies
type PasswordElevationStrategy struct {
    passwordManager *PasswordManager
}

type PolicyKitElevationStrategy struct {
    // Use PolicyKit for GUI elevation
}

type SudoersElevationStrategy struct {
    // Pre-configured sudoers rules
}
```

**Password Management:**
```go
type PasswordManager struct {
    cache       map[string]*CachedPassword
    cacheExpiry time.Duration
    maxAttempts int
    logger      *logger.Logger
}

type CachedPassword struct {
    hashedPassword string
    timestamp     time.Time
    attempts      int
}

func (pm *PasswordManager) RequestPassword(ctx context.Context) (string, error) {
    // Secure password input with context cancellation
    if pm.hasCachedPassword() {
        return pm.getCachedPassword()
    }
    
    return pm.promptPassword(ctx)
}

func (pm *PasswordManager) promptPassword(ctx context.Context) (string, error) {
    fmt.Print("🔐 Enter sudo password: ")
    
    // Use golang.org/x/term for secure input
    password, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return "", fmt.Errorf("failed to read password: %w", err)
    }
    
    return string(password), nil
}
```

#### 3.2 Advanced Security Features
**Files to create:**
- `internal/security/session_manager.go`
- `internal/security/privilege_escalation.go`
- `internal/security/command_validator.go`

**Session Management:**
```go
type SessionManager struct {
    activeSessions map[string]*ElevatedSession
    maxDuration    time.Duration
    logger         *logger.Logger
}

type ElevatedSession struct {
    SessionID   string
    StartTime   time.Time
    LastUsed    time.Time
    Commands    []string
    UserID      int
    MaxDuration time.Duration
}

func (sm *SessionManager) CreateSession(userID int) (*ElevatedSession, error) {
    session := &ElevatedSession{
        SessionID:   generateSessionID(),
        StartTime:   time.Now(),
        LastUsed:    time.Now(),
        UserID:      userID,
        MaxDuration: sm.maxDuration,
    }
    
    sm.activeSessions[session.SessionID] = session
    return session, nil
}
```

**Deliverables:**
- [ ] Complete sudo management system
- [ ] Multiple elevation strategies (password, PolicyKit, sudoers)
- [ ] Secure password caching with expiration
- [ ] Session management for elevated privileges
- [ ] Advanced security validation
- [ ] Security audit logging

### Phase 4: Configuration & Editor Integration (Week 4)

#### 4.1 Configuration Management
**Files to create:**
- `internal/execution/config_manager.go`
- `internal/execution/editor_integration.go`
- `internal/execution/backup_manager.go`

**Configuration Manager:**
```go
type ConfigManager struct {
    backupManager *BackupManager
    editor        EditorConfig
    validator     *ConfigValidator
    logger        *logger.Logger
}

type EditorConfig struct {
    DefaultEditor   string
    EditorPath      string
    EditorArgs      []string
    SupportsLSP     bool
    ConfigTemplate  string
}

func (cm *ConfigManager) EditConfiguration(ctx context.Context, filePath string) error {
    // 1. Create backup
    backup, err := cm.backupManager.CreateBackup(filePath)
    if err != nil {
        return fmt.Errorf("failed to create backup: %w", err)
    }
    
    // 2. Open in editor
    if err := cm.openEditor(ctx, filePath); err != nil {
        cm.backupManager.RestoreBackup(backup)
        return fmt.Errorf("editor failed: %w", err)
    }
    
    // 3. Validate configuration
    if err := cm.validator.Validate(filePath); err != nil {
        cm.logger.Warn("Configuration validation failed: %v", err)
        // Offer to restore backup or continue
    }
    
    return nil
}
```

**Backup Management:**
```go
type BackupManager struct {
    backupDir string
    maxBackups int
    logger    *logger.Logger
}

type Backup struct {
    OriginalPath string
    BackupPath   string
    Timestamp    time.Time
    Size         int64
    Checksum     string
}

func (bm *BackupManager) CreateBackup(filePath string) (*Backup, error) {
    timestamp := time.Now()
    backupName := fmt.Sprintf("%s.%s.backup", 
        filepath.Base(filePath), 
        timestamp.Format("20060102-150405"))
    
    backupPath := filepath.Join(bm.backupDir, backupName)
    
    // Copy file with verification
    checksum, err := bm.copyWithChecksum(filePath, backupPath)
    if err != nil {
        return nil, err
    }
    
    backup := &Backup{
        OriginalPath: filePath,
        BackupPath:   backupPath,
        Timestamp:    timestamp,
        Checksum:     checksum,
    }
    
    bm.cleanupOldBackups(filePath)
    return backup, nil
}
```

#### 4.2 Advanced Command Operations
**Files to create:**
- `internal/execution/compound_operations.go`
- `internal/execution/rollback_manager.go`
- `internal/execution/operation_queue.go`

**Compound Operations:**
```go
type CompoundOperation struct {
    ID          string
    Name        string
    Description string
    Commands    []CommandRequest
    RollbackCmd []CommandRequest
    executor    *SafeExecutor
}

func (co *CompoundOperation) Execute(ctx context.Context) (*CompoundResult, error) {
    result := &CompoundResult{
        OperationID: co.ID,
        StartTime:   time.Now(),
    }
    
    // Execute commands in sequence
    for i, cmd := range co.Commands {
        cmdResult, err := co.executor.ExecuteCommand(ctx, cmd)
        result.CommandResults = append(result.CommandResults, cmdResult)
        
        if err != nil {
            // Rollback on failure
            result.RollbackExecuted = true
            co.rollback(ctx, i)
            return result, err
        }
    }
    
    result.Success = true
    result.EndTime = time.Now()
    return result, nil
}
```

**Deliverables:**
- [ ] Configuration file management
- [ ] Editor integration (neovim, vim, nano)
- [ ] Automatic backup and restore
- [ ] Configuration validation
- [ ] Compound operation support
- [ ] Rollback mechanisms

### Phase 5: Advanced Features & Polish (Week 5)

#### 5.1 Advanced AI Integration
**Files to create:**
- `internal/ai/function/edit_config.go`
- `internal/ai/function/manage_packages.go`
- `internal/ai/function/system_rebuild.go`

**Specialized AI Functions:**
```go
// High-level package management function
type ManagePackagesFunction struct {
    executor      *execution.SafeExecutor
    configManager *execution.ConfigManager
}

func (mpf *ManagePackagesFunction) Call(params map[string]interface{}) (interface{}, error) {
    action := params["action"].(string) // "install", "remove", "search"
    packages := params["packages"].([]string)
    method := params["method"].(string) // "declarative", "imperative"
    
    switch action {
    case "install":
        if method == "declarative" {
            return mpf.addToConfiguration(packages)
        }
        return mpf.installImperatively(packages)
    case "remove":
        return mpf.removePackages(packages)
    case "search":
        return mpf.searchPackages(packages[0])
    }
    
    return nil, fmt.Errorf("unknown action: %s", action)
}
```

#### 5.2 User Experience Enhancements
**Files to create:**
- `internal/execution/progress_tracker.go`
- `internal/execution/command_history.go`
- `internal/execution/suggestions.go`

**Progress Tracking:**
```go
type ProgressTracker struct {
    operations map[string]*OperationProgress
    callbacks  []ProgressCallback
    logger     *logger.Logger
}

type OperationProgress struct {
    ID          string
    Description string
    StartTime   time.Time
    Progress    float64  // 0.0 to 1.0
    Stage       string
    Output      []string
    Completed   bool
}

func (pt *ProgressTracker) TrackOperation(id, description string, callback ProgressCallback) {
    progress := &OperationProgress{
        ID:          id,
        Description: description,
        StartTime:   time.Now(),
        Progress:    0.0,
    }
    
    pt.operations[id] = progress
    pt.callbacks = append(pt.callbacks, callback)
}
```

#### 5.3 Testing & Validation
**Files to create:**
- `internal/execution/testing/mock_executor.go`
- `internal/execution/testing/integration_tests.go`
- `tests/execution/e2e_tests.go`

**Comprehensive Testing:**
```go
type MockExecutor struct {
    commands     []CommandRequest
    responses    map[string]*ExecutionResult
    shouldFail   map[string]bool
    delays       map[string]time.Duration
}

func (me *MockExecutor) ExecuteCommand(ctx context.Context, req CommandRequest) (*ExecutionResult, error) {
    me.commands = append(me.commands, req)
    
    key := fmt.Sprintf("%s:%s", req.Command, strings.Join(req.Args, " "))
    
    if me.shouldFail[key] {
        return nil, fmt.Errorf("mock execution failed for: %s", key)
    }
    
    if delay, exists := me.delays[key]; exists {
        time.Sleep(delay)
    }
    
    if result, exists := me.responses[key]; exists {
        return result, nil
    }
    
    return &ExecutionResult{
        Success:   true,
        ExitCode:  0,
        Output:    fmt.Sprintf("Mock output for: %s", key),
        Timestamp: time.Now(),
    }, nil
}
```

**Deliverables:**
- [ ] Advanced AI functions for complex operations
- [ ] Real-time progress tracking
- [ ] Command history and suggestions
- [ ] Comprehensive test suite
- [ ] Performance benchmarks
- [ ] Security audit

## User Experience Scenarios

### Scenario 1: Package Installation
```
User: "I want to install and run btop to monitor my system"

nixai: I'll help you install and run btop, a modern system monitor.

🔍 Searching for btop package...
✅ Found: btop (system monitor)

I can:
1. Run it temporarily with `nix run nixpkgs#btop`
2. Add it to your system configuration permanently

Which would you prefer? (1/2): 1

⚡ Running btop temporarily...
🎯 Executing: nix run nixpkgs#btop
[btop interface opens]
```

### Scenario 2: Configuration Management
```
User: "Add docker to my system and configure it for my user"

nixai: I'll add Docker to your NixOS configuration and set it up for your user.

📋 This will:
1. Enable virtualisation.docker.enable = true
2. Add your user to the docker group
3. Open the configuration for review
4. Rebuild the system

⚠️  Requires sudo for system rebuild. Continue? (y/n): y

🔧 Modifying /etc/nixos/configuration.nix...
✅ Added Docker configuration
📝 Opening configuration in neovim for review...
[neovim opens with highlighted changes]

💻 Ready to rebuild system? (y/n): y
🔐 Enter sudo password: [secure input]
🚀 Rebuilding NixOS configuration...
✅ System rebuilt successfully!

🐳 Docker is now available. Try: docker run hello-world
```

### Scenario 3: Development Environment
```
User: "Set up a Python development environment with poetry"

nixai: I'll create a Python development environment with Poetry.

🐍 Setting up Python development environment...

📁 Creating project directory: ~/dev/python-project
🔧 Generating shell.nix with Python 3.11 and Poetry...
📝 Opening shell.nix in your editor for customization...

✅ Environment ready! 
🚀 To enter: cd ~/dev/python-project && nix-shell
🎯 Or use: nix develop (if using flakes)
```

## Security Considerations

### 1. Command Validation
- **Whitelist-based**: Only allow pre-approved commands
- **Argument validation**: Validate command arguments against known patterns
- **Path restrictions**: Limit operations to safe directories
- **Resource limits**: Timeout and memory constraints

### 2. Privilege Management
- **Least privilege**: Request minimal necessary permissions
- **Session management**: Time-limited elevated sessions
- **Audit logging**: Complete command execution audit trail
- **User confirmation**: Explicit approval for dangerous operations

### 3. Input Sanitization
- **Command injection prevention**: Sanitize all user inputs
- **Argument escaping**: Proper shell argument escaping
- **Environment isolation**: Controlled environment variables
- **Working directory restrictions**: Limit to safe directories

### 4. Emergency Procedures
- **Kill switches**: Ability to terminate running operations
- **Rollback mechanisms**: Automatic system rollback on failure
- **Safe mode**: Disable execution in case of security concerns
- **Emergency access**: Alternative access methods if system breaks

## Configuration Examples

### Basic Configuration
```yaml
# ~/.config/nixai/execution.yaml
execution:
  enabled: true
  confirmationRequired: true
  defaultEditor: "nvim"
  
  categories:
    package:
      confirmationRequired: false
      maxExecutionTime: "5m"
    system:
      confirmationRequired: true
      requiresSudo: true
```

### Advanced Configuration
```yaml
execution:
  enabled: true
  
  security:
    auditLogging: true
    sessionTimeout: "30m"
    maxConcurrentCommands: 2
    
  permissions:
    allowedCommands:
      - "nix*"
      - "nixos-rebuild"
      - "systemctl"
      - "journalctl"
    
    forbiddenCommands:
      - "rm -rf /"
      - "dd if=*"
      - "mkfs*"
      
    sudoCommands:
      - "nixos-rebuild"
      - "systemctl start"
      - "systemctl stop"
      - "systemctl restart"

  editor:
    default: "nvim"
    backup: true
    validateOnSave: true
```

## Risks & Mitigation

### High-Risk Scenarios
1. **Accidental System Damage**
   - *Mitigation*: Comprehensive backup system, rollback mechanisms
   
2. **Privilege Escalation**
   - *Mitigation*: Strict sudo management, session timeouts
   
3. **Command Injection**
   - *Mitigation*: Input sanitization, command whitelisting
   
4. **Resource Exhaustion**
   - *Mitigation*: Execution timeouts, resource limits

### Medium-Risk Scenarios
1. **Configuration Corruption**
   - *Mitigation*: Automatic backups, validation
   
2. **Unintended Command Execution**
   - *Mitigation*: User confirmation, dry-run mode
   
3. **Password Exposure**
   - *Mitigation*: Secure password handling, no logging

## Success Metrics

### Technical Metrics
- **Security**: Zero successful privilege escalations
- **Reliability**: 99.9% successful command execution rate
- **Performance**: <2s average command initiation time
- **Recovery**: <30s average rollback time

### User Experience Metrics
- **Efficiency**: 50% reduction in manual command typing
- **Accuracy**: 95% successful task completion rate
- **Learning**: 80% user retention after first week
- **Satisfaction**: >4.5/5 user satisfaction score

## Future Enhancements

### Phase 6: Advanced Features (Future)
- **AI Command Generation**: Generate complex command sequences
- **Machine Learning**: Learn user patterns and preferences
- **Remote Execution**: Execute commands on remote NixOS machines
- **Workflow Automation**: Create reusable command workflows
- **Integration APIs**: External tool integration
- **Voice Commands**: Voice-controlled command execution

### Phase 7: Enterprise Features (Future)
- **Multi-user Management**: Team-based permission systems
- **Compliance Logging**: Advanced audit and compliance features
- **Policy Management**: Organization-wide execution policies
- **Remote Monitoring**: Centralized command execution monitoring
- **Integration Ecosystem**: Enterprise tool integrations

## Conclusion

This implementation plan provides a comprehensive roadmap for adding safe, AI-controlled command execution to nixai. The phased approach ensures security and reliability while delivering immediate value to users. The system will transform nixai from a helpful assistant into a powerful automation platform that makes NixOS management accessible to users of all skill levels.

The key to success is maintaining the balance between powerful automation capabilities and robust security measures, ensuring users can accomplish complex tasks safely and efficiently through natural language interaction.