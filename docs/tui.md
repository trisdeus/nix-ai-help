# nixai tui - Enhanced Terminal User Interface

The `nixai tui` command provides two distinct terminal user interface experiences: a modern Claude Code-style interface and a classic menu-based interface.

## Overview

The TUI (Terminal User Interface) offers an interactive, keyboard-driven experience for accessing all nixai functionality without the need for a web browser. Choose between modern and classic interfaces based on your preferences and terminal capabilities.

## Usage

```bash
# Modern Claude Code-style interface (default)
nixai tui

# Classic menu-based interface
nixai tui --classic

# Alternative command (same as default)
nixai interactive
```

## Interface Options

### Modern Interface (Default)

The modern interface provides a Claude Code-style experience with enhanced navigation and real-time interaction:

```text
┌─ Commands (40+ total) ─────────────┬─ Execution Panel ─────────────────┐
│                                    │                                   │
│ ask [INPUT]                        │ Welcome to nixai TUI!             │
│   Ask any NixOS question           │ Select a command from the left    │
│ web start                          │ panel to get started.             │
│   Launch modern web interface      │                                   │
│ plugin list                        │ Latest Updates:                   │
│   Show installed plugins           │ • Modern web interface            │
│ fleet health                       │ • Enhanced plugin system          │
│   Check fleet status               │ • Fleet management                │
│ hardware detect                    │ • Version control integration     │
│   Comprehensive hardware analysis  │ • Real-time collaboration         │
│                                    │                                   │
│ (Showing 1-10 of 40+)             │ [INPUT] = Interactive Parameters  │
└────────────────────────────────────┴───────────────────────────────────┘
Commands | ?:Changelog | Tab:Switch | ↑↓:Navigate | Enter:Select | nixai v2.0.0
```

### Classic Interface

The classic interface provides a traditional menu-driven experience:

```text
═══════════════════════════════════════════════════════════════════════════════
                           nixai Interactive Menu v2.0.0
═══════════════════════════════════════════════════════════════════════════════

Main Categories:
  1. AI & Questions         - Ask questions and get AI assistance
  2. System Management      - Hardware, diagnostics, and health checks
  3. Package & Search       - Search, explain, and manage packages
  4. Configuration          - Build, configure, and manage NixOS settings
  5. Development           - Flakes, development environments, and tools
  6. Web Interface         - Launch modern web dashboard
  7. Plugin System         - Manage and install plugins
  8. Fleet Management      - Multi-machine deployment and monitoring
  9. Advanced Tools        - Advanced system utilities and analysis

Enter number (1-9), 'help' for command help, or 'exit' to quit:
```

## Features

### 🎯 Modern Interface Features

#### Dual-Panel Layout
- **Left Panel**: Command browser with search and navigation
- **Right Panel**: Execution area with real-time output
- **Status Bar**: Version info, help shortcuts, and navigation hints

#### Enhanced Navigation
- **Arrow Keys**: Navigate through command list
- **Tab Key**: Switch between panels
- **Page Up/Down**: Scroll through long lists
- **Enter**: Execute selected command
- **ESC**: Cancel operation or go back

#### Interactive Elements
- **Command Search**: Type `/` to search commands
- **Parameter Input**: `[INPUT]` commands provide guided parameter entry
- **Real-time Output**: Live command execution with progress indicators
- **Help System**: Press `?` for changelog and feature updates

#### Visual Enhancements
- **Syntax Highlighting**: Color-coded output for better readability
- **Progress Indicators**: Real-time feedback for long-running operations
- **Status Messages**: Clear indication of command status and results
- **Responsive Layout**: Adapts to terminal size changes

### 🔧 Classic Interface Features

#### Menu Navigation
- **Numbered Options**: Simple numeric selection for all functions
- **Category Organization**: Commands grouped by functionality
- **Breadcrumb Navigation**: Clear indication of current location
- **Back/Exit Options**: Easy navigation between menus

#### Compatibility
- **Wide Terminal Support**: Works on any terminal with basic capabilities
- **Screen Reader Friendly**: Compatible with accessibility tools
- **Low Resource Usage**: Minimal memory and CPU requirements
- **Fallback Mode**: Automatic fallback for unsupported terminals

## Interactive Commands

Both interfaces support interactive parameter input for complex commands:

### AI Questions
```text
Enter your question: How do I enable SSH in NixOS?
Select AI provider [1] Ollama [2] OpenAI [3] Gemini: 2
Select role [1] Default [2] System Admin [3] Security Expert: 2
Processing... ✓ Response ready
```

### Configuration Building
```text
Configuration type [1] Desktop [2] Server [3] Minimal: 1
Desktop environment [1] GNOME [2] KDE [3] XFCE: 1
Enable additional services? [y/N]: y
Select services: [1] SSH [2] Firewall [3] Docker: 1,2
Generating configuration... ✓ Complete
```

### Hardware Analysis
```text
Analysis depth [1] Quick [2] Comprehensive [3] Full: 2
Include optimization suggestions? [Y/n]: Y
Scanning hardware... ✓ Analysis complete
Display optimization recommendations? [Y/n]: Y
```

## Keyboard Shortcuts

### Modern Interface
- `↑↓` - Navigate command list
- `Tab` - Switch between panels
- `Enter` - Execute selected command
- `/` - Search commands
- `?` - Show changelog and help
- `Ctrl+C` - Exit interface
- `Page Up/Down` - Scroll through long lists
- `Home/End` - Jump to start/end of list

### Classic Interface
- `1-9` - Select menu option
- `Enter` - Confirm selection
- `q` or `exit` - Quit interface
- `back` - Return to previous menu
- `help` - Show command help
- `clear` - Clear screen

## Configuration

### TUI Settings

Configure the TUI experience through the configuration file:

```yaml
tui:
  default_interface: "modern"  # "modern" or "classic"
  theme:
    primary_color: "blue"
    accent_color: "green"
    text_color: "white"
    background_color: "black"
  
  features:
    syntax_highlighting: true
    real_time_output: true
    progress_indicators: true
    auto_complete: true
  
  layout:
    command_panel_width: 40
    min_terminal_width: 80
    min_terminal_height: 24
    status_bar_enabled: true
  
  behavior:
    auto_save_session: true
    confirm_destructive_actions: true
    enable_shortcuts: true
    scroll_buffer_size: 1000
```

### Environment Variables

Control TUI behavior with environment variables:

```bash
# Force classic interface
export NIXAI_TUI_CLASSIC=1

# Disable color output
export NIXAI_NO_COLOR=1

# Set terminal size detection
export NIXAI_TUI_AUTO_SIZE=1

# Enable debug mode
export NIXAI_TUI_DEBUG=1
```

## Advanced Usage

### Batch Command Execution

Execute multiple commands in sequence:

```bash
# Use the TUI for guided multi-command workflows
nixai tui --batch-mode

# Example batch operations:
# 1. Run system diagnostics
# 2. Update package list
# 3. Check for configuration issues
# 4. Generate optimization report
```

### Session Management

The TUI supports session persistence:

```text
Session Features:
- Command history preservation
- Recent selections memory
- Customized panel layouts
- Bookmark frequently used commands
- Resume interrupted operations
```

### Integration with External Tools

Use the TUI as part of larger workflows:

```bash
# Pipe output to external tools
nixai tui --output-format json | jq '.results'

# Integration with scripts
nixai tui --non-interactive --command "hardware detect"

# Export session data
nixai tui --export-session session.json
```

## Accessibility Features

### Modern Interface
- **High Contrast Mode**: Enhanced visibility for low-vision users
- **Screen Reader Support**: Compatible with accessibility tools
- **Keyboard Navigation**: Complete functionality without mouse
- **Adjustable Text Size**: Respects terminal font settings
- **Color Blind Support**: Distinguishable without color dependency

### Classic Interface
- **Simple Layout**: Clean, distraction-free interface
- **Large Text Options**: Configurable for better visibility
- **Voice Navigation**: Compatible with voice control software
- **Minimal Visual Elements**: Reduces cognitive load
- **Consistent Navigation**: Predictable menu structure

## Performance Optimization

### Resource Usage
- **Efficient Rendering**: Optimized for low CPU usage
- **Memory Management**: Minimal memory footprint
- **Network Efficiency**: Reduced API calls for better responsiveness
- **Cache Management**: Smart caching for frequently accessed data

### Large Configuration Handling
- **Lazy Loading**: Load content as needed
- **Pagination**: Handle large command lists efficiently
- **Search Optimization**: Fast command and option searching
- **Background Processing**: Non-blocking operations where possible

## Troubleshooting

### Common Issues

**Terminal size too small:**
```bash
# Check minimum requirements
echo "Current size: $(tput cols)x$(tput lines)"
# Minimum: 80x24 for modern interface
```

**Interface rendering issues:**
```bash
# Force classic interface for compatibility
nixai tui --classic

# Or disable advanced features
export NIXAI_TUI_SIMPLE=1
nixai tui
```

**Color display problems:**
```bash
# Disable color output
nixai tui --no-color

# Or check terminal color support
echo $COLORTERM
```

**Keyboard shortcuts not working:**
```bash
# Check terminal key mapping
nixai tui --debug-keys

# Reset to defaults
nixai config reset-tui
```

### Debug Mode

Enable comprehensive debugging:

```bash
nixai tui --debug --verbose
```

Debug output includes:
- Terminal capability detection
- Key binding information
- Rendering performance metrics
- Command execution details

### Performance Issues

For slow performance on older systems:

```bash
# Use lightweight mode
nixai tui --lightweight

# Disable animations
nixai tui --no-animations

# Reduce refresh rate
nixai tui --refresh-rate 1
```

## Best Practices

1. **Terminal Setup**: Use a modern terminal emulator for best experience
2. **Font Selection**: Choose a monospace font with good Unicode support
3. **Size Optimization**: Maintain at least 80x24 terminal size
4. **Color Scheme**: Use high contrast themes for better visibility
5. **Regular Updates**: Keep nixai updated for latest TUI improvements

For more detailed examples and advanced configurations, see the [TUI Guide](../examples/tui-interface/) and [Configuration Reference](config.md).