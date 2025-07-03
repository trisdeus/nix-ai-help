# NixAI Claude Code-Style TUI

The NixAI Terminal User Interface (TUI) now features a Claude Code-style interface that provides a modern, interactive command-line experience similar to code editors and AI assistants.

## 🚀 Features

### **Command Box at Bottom**
- Modern command input at the bottom of the screen
- Real-time command suggestions as you type
- Command history navigation with ↑/↓ arrows
- Tab completion for commands and options

### **Command Execution Output Above**
- All command output displays in the main area above the command box
- Scrollable output with automatic line limiting
- Color-coded output (success, error, info)
- Timestamps for each command execution

### **Interactive Features**
- **Command Completion**: Type partial commands to see suggestions
- **History Navigation**: Use ↑/↓ to navigate through previous commands
- **Tab Completion**: Press Tab to complete the selected suggestion
- **Built-in Commands**: `help`, `clear`, `history`, `exit`

## 🎯 Getting Started

### Start the TUI
```bash
# Start Claude Code-style TUI (default)
nixai tui

# Start classic menu-based TUI
nixai tui --classic
```

### Basic Commands
```bash
# Get help
help

# Clear the terminal
clear

# Show command history
history

# Exit the TUI
exit
# or press Ctrl+C / Esc
```

## 📋 Command Examples

### NixOS Configuration Commands
```bash
# Ask AI questions
ask "how to configure nginx?"
ask "fix boot issues"

# Build and analyze configurations
build
build --dry-run
build analyze

# Interactive configuration
configure
configure nginx
configure desktop

# System diagnostics
diagnose
diagnose boot
diagnose services
```

### Flake Management
```bash
# Flake operations
flake create
flake validate
flake migrate

# Learning modules
learn list
learn basics
learn progress
```

### Fleet and Team Management
```bash
# Fleet operations
fleet list
fleet deploy
fleet status

# Team collaboration
team create
team members
team permissions
```

### Web Interface
```bash
# Start web interface
web start
web start --port 8080
web start --repo /path/to/repo
```

## ⌨️ Keyboard Navigation

| Key | Action |
|-----|--------|
| `↑` | Navigate command history / Move up in suggestions |
| `↓` | Navigate command history / Move down in suggestions |
| `Tab` | Complete selected suggestion |
| `Enter` | Execute command |
| `Ctrl+C` | Exit TUI |
| `Esc` | Exit TUI |

## 🎨 Interface Layout

```
┌─────────────────────────────────────────────────────────────┐
│ 🚀 NixAI Terminal Interface - Claude Code Style           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ [15:30:45] $ nixai ask "how to configure nginx?"           │
│ To configure nginx in NixOS, you can use the following...  │
│                                                             │
│ [15:31:02] $ nixai build                                   │
│ Building NixOS configuration...                            │
│ Build completed successfully                               │
│                                                             │
│ [15:31:15] $ nixai help                                    │
│ NixAI Commands:                                            │
│                                                             │
│ ▶ AI                                                        │
│   ask - Ask AI questions about NixOS                      │
│                                                             │
│ ▶ Build                                                     │
│   build - Build and analyze NixOS configurations          │
│                                                             │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│ Suggestions:                                                │
│   ask "how to configure nginx?"                           │
│   ask "fix boot issues"                                   │
│   build --dry-run                                         │
├─────────────────────────────────────────────────────────────┤
│ ╭─────────────────────────────────────────────────────────╮ │
│ │ nixai > configure ngi_                                 │ │
│ ╰─────────────────────────────────────────────────────────╯ │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Advanced Features

### Command Suggestions
The TUI provides intelligent command suggestions based on:
- Available nixai commands
- Common command patterns
- Previous command history
- Contextual completions

### Command History
- Persistent command history across sessions
- Navigate with ↑/↓ arrows
- View history with `history` command
- Search through previous commands

### Output Management
- Automatic scrolling for long outputs
- Color-coded messages (success, error, warning)
- Timestamped command execution
- Line limiting to prevent memory issues

## 🎯 Tips & Tricks

1. **Quick Commands**: Type the first few letters and press Tab to complete
2. **History Search**: Use ↑/↓ to quickly find previous commands
3. **Built-in Help**: Type `help` followed by any command for detailed information
4. **Clear Output**: Use `clear` to clean the terminal when it gets cluttered
5. **Exit Options**: Both `exit` command and Ctrl+C work to quit

## 🐛 Troubleshooting

### TUI Won't Start
```bash
# Check if dependencies are installed
go mod tidy

# Try classic TUI instead
nixai tui --classic
```

### Command Not Working
```bash
# Check command syntax
help [command-name]

# Verify nixai installation
nixai --version
```

### Display Issues
- Resize terminal window
- Check terminal compatibility
- Try different terminal emulator

## 🔄 Comparison: Claude Code Style vs Classic

| Feature | Claude Code Style | Classic TUI |
|---------|------------------|-------------|
| Command Input | Bottom command box | Inline prompts |
| Completion | Real-time suggestions | Number selection |
| History | ↑/↓ navigation | Manual entry |
| Output | Scrollable area above | Full screen clear |
| UI Style | Modern, editor-like | Menu-based |
| Learning Curve | Familiar to developers | Simple and direct |

## 📚 Related Documentation

- [NixAI CLI Commands](./MANUAL.md)
- [Configuration Guide](./config.md)
- [Web Interface Guide](./web-interface.md)
- [Troubleshooting](./TROUBLESHOOTING.md)

---

The Claude Code-style TUI provides a modern, efficient way to interact with NixAI that feels familiar to developers who use code editors and AI assistants. It combines the power of the command line with the convenience of modern UI patterns.