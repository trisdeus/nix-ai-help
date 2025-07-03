package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nix-ai-help/pkg/utils"
	"nix-ai-help/pkg/version"
)

// ClaudeTUI represents the Claude Code-style TUI
type ClaudeTUI struct {
	textInput       textinput.Model
	output          []string
	commandHistory  []string
	historyIndex    int
	width           int
	height          int
	suggestions     []string
	showSuggestions bool
	selectedSuggestion int
	currentTheme    string
	styles          ThemeStyles
	
	// Plugin integration
	pluginCommands  []string
	pluginSuggestions []string
}

// NewClaudeTUI creates a new Claude Code-style TUI
func NewClaudeTUI() *ClaudeTUI {
	ti := textinput.New()
	ti.Placeholder = "Enter nixai command..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Initialize with Gruvbox theme as default
	defaultTheme := "gruvbox"
	theme := themes[defaultTheme]
	styles := createThemeStyles(theme)

	// Initialize plugin integration
	tui := &ClaudeTUI{
		textInput:      ti,
		output:         []string{},
		commandHistory: []string{},
		historyIndex:   -1,
		suggestions:    getAllCommandSuggestions(),
		currentTheme:   defaultTheme,
		styles:         styles,
		pluginCommands: []string{},
		pluginSuggestions: []string{},
	}
	
	// Initialize plugin manager for dynamic commands
	tui.initializePluginIntegration()
	
	return tui
}

// initializePluginIntegration sets up plugin command discovery for TUI
func (m *ClaudeTUI) initializePluginIntegration() {
	// Initialize the integrated plugin commands
	m.updatePluginCommands()
}

// updatePluginCommands refreshes the list of available plugin commands
func (m *ClaudeTUI) updatePluginCommands() {
	// This would query the plugin manager for available commands
	// For now, add the built-in plugin commands we created
	m.pluginCommands = []string{
		"system-info", "system-info health", "system-info status", "system-info cpu",
		"system-info memory", "system-info disk", "system-info processes", "system-info monitor",
		"package-monitor", "package-monitor list", "package-monitor updates", "package-monitor security",
		"package-monitor analyze", "package-monitor orphans", "package-monitor stats",
	}
	
	// Add plugin suggestions to the main suggestions list
	m.pluginSuggestions = append(m.pluginCommands,
		// Add example usage
		"system-info health --json",
		"system-info monitor --interval 3",
		"package-monitor list --detailed",
		"package-monitor updates --security",
	)
	
	// Merge with main suggestions
	allSuggestions := append(m.suggestions, m.pluginSuggestions...)
	m.suggestions = removeDuplicates(allSuggestions)
}

// getAllCommandSuggestions returns all available nixai commands for completion
func getAllCommandSuggestions() []string {
	return []string{
		// Core AI commands
		"ai-config", "ask", "configure", "explain-option", "explain-home-option",
		
		// Build and development
		"build", "devenv", "import", "templates",
		
		// Diagnostics and troubleshooting
		"diagnose", "doctor", "error", "logs", "performance",
		
		// Package and dependency management
		"deps", "package-repo", "search", "store", "gc",
		
		// Flake management
		"flake", "migrate",
		
		// System and hardware
		"hardware", "context", "intelligence",
		
		// Integrated plugin commands
		"system-info", "package-monitor",
		
		// Learning and help
		"learn", "help", "community", "snippets",
		
		// Team and collaboration
		"team", "fleet", "machines",
		
		// Workflow and automation
		"workflow", "plugin", "version-control", "execute",
		
		// Web and integrations
		"web", "mcp-server", "neovim-setup",
		
		// Configuration management
		"config", "completion", "tui",
		
		// Common command examples with options
		"ask \"how to configure nginx?\"",
		"ask \"fix boot issues\"",
		"ask \"setup development environment\"",
		"build --dry-run",
		"build analyze",
		"configure nginx",
		"configure desktop",
		"configure development",
		"diagnose boot",
		"diagnose services",
		"diagnose network",
		"doctor --full",
		"doctor --quick",
		"flake create",
		"flake validate",
		"flake migrate",
		"learn list",
		"learn basics",
		"learn advanced",
		"web start",
		"web start --port 8080",
		"web start --repo /path/to/repo",
		"fleet list",
		"fleet deploy",
		"fleet status",
		"team create",
		"team members",
		"team permissions",
		"search nginx",
		"search postgresql",
		"hardware detect",
		"hardware optimize",
		"performance stats",
		"performance cache",
		"mcp-server start",
		"mcp-server status",
		"neovim-setup install",
		"package-repo analyze",
		"store analyze",
		"gc run",
		"templates list",
		"snippets list",
		
		// Integrated plugin examples
		"system-info health",
		"system-info status",
		"system-info cpu",
		"system-info memory",
		"system-info disk",
		"system-info processes",
		"system-info monitor",
		"system-info all",
		"package-monitor list",
		"package-monitor updates",
		"package-monitor security",
		"package-monitor analyze",
		"package-monitor stats",
		
		// Execution examples
		"execute status",
		"execute config",
		"execute history",
		"execute nix-env -iA nixpkgs.firefox",
		"execute --dry-run nixos-rebuild switch",
		"execute --category package nix-collect-garbage -d",
		"execute --description \"Update system\" nixos-rebuild switch",
		
		"help ask",
		"help build",
		"help configure",
		"help diagnose",
		"help flake",
		"help web",
		"help system-info",
		"help package-monitor",
	}
}

// Theme represents a color theme for the TUI
type Theme struct {
	Name       string
	Background string
	Foreground string
	Primary    string
	Secondary  string
	Success    string
	Warning    string
	Error      string
	Muted      string
	Accent     string
	Border     string
}

// Available themes
var themes = map[string]Theme{
	"gruvbox": {
		Name:       "Gruvbox Dark",
		Background: "#282828",
		Foreground: "#ebdbb2",
		Primary:    "#fabd2f",
		Secondary:  "#83a598",
		Success:    "#b8bb26",
		Warning:    "#fe8019",
		Error:      "#fb4934",
		Muted:      "#928374",
		Accent:     "#d3869b",
		Border:     "#504945",
	},
	"dracula": {
		Name:       "Dracula",
		Background: "#282a36",
		Foreground: "#f8f8f2",
		Primary:    "#bd93f9",
		Secondary:  "#8be9fd",
		Success:    "#50fa7b",
		Warning:    "#ffb86c",
		Error:      "#ff5555",
		Muted:      "#6272a4",
		Accent:     "#ff79c6",
		Border:     "#44475a",
	},
	"nord": {
		Name:       "Nord",
		Background: "#2e3440",
		Foreground: "#eceff4",
		Primary:    "#88c0d0",
		Secondary:  "#81a1c1",
		Success:    "#a3be8c",
		Warning:    "#ebcb8b",
		Error:      "#bf616a",
		Muted:      "#4c566a",
		Accent:     "#b48ead",
		Border:     "#3b4252",
	},
	"tokyo-night": {
		Name:       "Tokyo Night",
		Background: "#1a1b26",
		Foreground: "#c0caf5",
		Primary:    "#7aa2f7",
		Secondary:  "#bb9af7",
		Success:    "#9ece6a",
		Warning:    "#e0af68",
		Error:      "#f7768e",
		Muted:      "#565f89",
		Accent:     "#ff9e64",
		Border:     "#292e42",
	},
	"catppuccin": {
		Name:       "Catppuccin Mocha",
		Background: "#1e1e2e",
		Foreground: "#cdd6f4",
		Primary:    "#89b4fa",
		Secondary:  "#cba6f7",
		Success:    "#a6e3a1",
		Warning:    "#f9e2af",
		Error:      "#f38ba8",
		Muted:      "#6c7086",
		Accent:     "#f5c2e7",
		Border:     "#313244",
	},
}

// ThemeStyles contains all the styled components for the current theme
type ThemeStyles struct {
	header      lipgloss.Style
	commandBox  lipgloss.Style
	output      lipgloss.Style
	suggestion  lipgloss.Style
	selectedSug lipgloss.Style
	prompt      lipgloss.Style
	timestamp   lipgloss.Style
	error       lipgloss.Style
	success     lipgloss.Style
	warning     lipgloss.Style
	muted       lipgloss.Style
	accent      lipgloss.Style
	logo        lipgloss.Style
	versionInfo lipgloss.Style
}

// createThemeStyles creates styled components from a theme
func createThemeStyles(theme Theme) ThemeStyles {
	return ThemeStyles{
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color(theme.Border)),

		commandBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Primary)).
			Padding(0, 1).
			Margin(0, 1).
			Background(lipgloss.Color(theme.Background)),

		output: lipgloss.NewStyle().
			Padding(0, 2).
			MarginBottom(1).
			Foreground(lipgloss.Color(theme.Foreground)),

		suggestion: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)).
			Background(lipgloss.Color(theme.Background)).
			Padding(0, 1),

		selectedSug: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Background)).
			Background(lipgloss.Color(theme.Primary)).
			Padding(0, 1).
			Bold(true),

		prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true),

		timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)).
			Italic(true),

		error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Error)).
			Bold(true),

		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Success)).
			Bold(true),

		warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Warning)).
			Bold(true),

		muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)),

		accent: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Accent)).
			Bold(true),

		logo: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Primary)).
			Bold(true),

		versionInfo: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Muted)).
			Italic(true).
			Align(lipgloss.Right).
			Padding(0, 1),
	}
}

// Init initializes the TUI
func (m *ClaudeTUI) Init() tea.Cmd {
	m.showLogo()
	return textinput.Blink
}

// Update handles messages and updates the model
func (m *ClaudeTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyUp:
			if m.showSuggestions && len(m.suggestions) > 0 {
				m.selectedSuggestion--
				if m.selectedSuggestion < 0 {
					m.selectedSuggestion = len(m.getFilteredSuggestions()) - 1
				}
				return m, nil
			}
			// Navigate command history
			if len(m.commandHistory) > 0 {
				if m.historyIndex == -1 {
					m.historyIndex = len(m.commandHistory) - 1
				} else if m.historyIndex > 0 {
					m.historyIndex--
				}
				if m.historyIndex >= 0 && m.historyIndex < len(m.commandHistory) {
					m.textInput.SetValue(m.commandHistory[m.historyIndex])
				}
			}
			return m, nil

		case tea.KeyDown:
			if m.showSuggestions && len(m.suggestions) > 0 {
				m.selectedSuggestion++
				if m.selectedSuggestion >= len(m.getFilteredSuggestions()) {
					m.selectedSuggestion = 0
				}
				return m, nil
			}
			// Navigate command history
			if len(m.commandHistory) > 0 {
				if m.historyIndex < len(m.commandHistory)-1 {
					m.historyIndex++
					m.textInput.SetValue(m.commandHistory[m.historyIndex])
				} else {
					m.historyIndex = -1
					m.textInput.SetValue("")
				}
			}
			return m, nil

		case tea.KeyTab:
			if m.showSuggestions && len(m.getFilteredSuggestions()) > 0 {
				filtered := m.getFilteredSuggestions()
				if m.selectedSuggestion >= 0 && m.selectedSuggestion < len(filtered) {
					m.textInput.SetValue(filtered[m.selectedSuggestion])
					m.showSuggestions = false
				}
			}
			return m, nil

		case tea.KeyEnter:
			input := strings.TrimSpace(m.textInput.Value())
			if input != "" {
				m.executeCommand(input)
				m.commandHistory = append(m.commandHistory, input)
				m.historyIndex = -1
				m.textInput.SetValue("")
				m.showSuggestions = false
			}
			return m, nil

		default:
			// Update suggestions based on input
			m.textInput, cmd = m.textInput.Update(msg)
			m.updateSuggestions()
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 8 // Account for borders and padding
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// updateSuggestions updates the suggestion list based on current input
func (m *ClaudeTUI) updateSuggestions() {
	input := m.textInput.Value()
	if len(input) > 0 {
		m.showSuggestions = true
		m.selectedSuggestion = 0
	} else {
		m.showSuggestions = false
	}
}

// getFilteredSuggestions returns suggestions that match the current input
func (m *ClaudeTUI) getFilteredSuggestions() []string {
	input := strings.ToLower(m.textInput.Value())
	if input == "" {
		return m.suggestions[:10] // Show first 10 suggestions when no input
	}

	// Parse the current input to provide intelligent completion
	return m.getIntelligentSuggestions(input)
}

// getIntelligentSuggestions provides context-aware command completion
func (m *ClaudeTUI) getIntelligentSuggestions(input string) []string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return m.suggestions[:10]
	}

	// Get the current command being typed
	currentCommand := parts[0]
	
	// If we're still typing the first word (command name)
	if !strings.HasSuffix(input, " ") && len(parts) == 1 {
		return m.getCommandSuggestions(currentCommand)
	}
	
	// If we have a complete command and are looking for flags/options
	if len(parts) >= 1 {
		return m.getCommandOptionSuggestions(currentCommand, parts[1:], input)
	}

	return []string{}
}

// getCommandSuggestions returns command name suggestions
func (m *ClaudeTUI) getCommandSuggestions(partial string) []string {
	var filtered []string
	for _, suggestion := range m.suggestions {
		cmd := strings.Fields(suggestion)[0]
		if strings.HasPrefix(strings.ToLower(cmd), strings.ToLower(partial)) {
			filtered = append(filtered, suggestion)
		}
	}
	
	// Limit to 8 suggestions
	if len(filtered) > 8 {
		filtered = filtered[:8]
	}
	
	return filtered
}

// getCommandOptionSuggestions returns flag and option suggestions for commands
func (m *ClaudeTUI) getCommandOptionSuggestions(command string, args []string, fullInput string) []string {
	suggestions := []string{}
	
	// Get available flags for the command
	flags := m.getCommandFlags(command)
	
	// Get the last word being typed
	lastWord := ""
	if len(args) > 0 {
		lastWord = args[len(args)-1]
	}
	
	// If the last character is a space, show all available options
	if strings.HasSuffix(fullInput, " ") {
		suggestions = append(suggestions, flags...)
	} else {
		// Filter flags based on partial input
		for _, flag := range flags {
			if strings.HasPrefix(strings.ToLower(flag), strings.ToLower(lastWord)) {
				suggestions = append(suggestions, strings.TrimSpace(fullInput[:len(fullInput)-len(lastWord)]) + flag)
			}
		}
	}
	
	// Add contextual value suggestions for certain flags
	suggestions = append(suggestions, m.getValueSuggestions(command, args, fullInput)...)
	
	// Limit to 8 suggestions
	if len(suggestions) > 8 {
		suggestions = suggestions[:8]
	}
	
	return suggestions
}

// getCommandFlags returns available flags for a specific command
func (m *ClaudeTUI) getCommandFlags(command string) []string {
	flagMap := map[string][]string{
		"ask": {
			"--provider openai", "--provider claude", "--provider gemini", "--provider ollama",
			"--model gpt-4", "--model claude-3", "--model gemini-pro",
			"--context-file", "--role", "--agent",
			"--nixos-path", "--help",
		},
		"build": {
			"--dry-run", "--show-trace", "--verbose", "--quiet",
			"--nixos-path", "--target", "--help",
			"analyze", "monitor", "recovery",
		},
		"configure": {
			"nginx", "desktop", "development", "server", "security",
			"--interactive", "--template", "--help",
		},
		"diagnose": {
			"boot", "services", "network", "hardware", "performance",
			"--full", "--quick", "--output-file", "--help",
		},
		"doctor": {
			"--full", "--quick", "--fix", "--verbose", "--help",
		},
		"flake": {
			"create", "validate", "migrate", "update", "check",
			"--help", "--verbose",
		},
		"web": {
			"start", "stop", "status",
			"--port 8080", "--port 3000", "--port 8000",
			"--repo", "--host", "--help",
		},
		"plugin": {
			"list", "load", "unload", "status", "execute",
			"--help", "--verbose",
		},
		"service": {
			"start", "stop", "restart", "status", "enable", "disable",
			"--help", "--all",
		},
		"search": {
			"--limit 10", "--limit 20", "--limit 50",
			"--category", "--help",
		},
		"learn": {
			"list", "start", "progress", "complete",
			"basics", "advanced", "expert",
			"--help", "--interactive",
		},
		"help": {
			"ask", "build", "configure", "diagnose", "doctor",
			"flake", "web", "plugin", "search", "learn",
		},
	}
	
	if flags, exists := flagMap[command]; exists {
		return flags
	}
	
	// Default flags for unknown commands
	return []string{"--help", "--verbose", "--quiet"}
}

// getValueSuggestions returns contextual value suggestions
func (m *ClaudeTUI) getValueSuggestions(command string, args []string, fullInput string) []string {
	suggestions := []string{}
	
	// Check if we're completing a flag value
	if len(args) >= 2 {
		lastFlag := args[len(args)-2]
		
		switch lastFlag {
		case "--provider":
			suggestions = []string{
				fullInput + "openai", fullInput + "claude", 
				fullInput + "gemini", fullInput + "ollama",
			}
		case "--model":
			suggestions = []string{
				fullInput + "gpt-4", fullInput + "claude-3",
				fullInput + "gemini-pro", fullInput + "llama3",
			}
		case "--port":
			suggestions = []string{
				fullInput + "8080", fullInput + "3000", 
				fullInput + "8000", fullInput + "9000",
			}
		case "--role":
			suggestions = []string{
				fullInput + "diagnoser", fullInput + "explainer",
				fullInput + "builder", fullInput + "assistant",
			}
		case "--agent":
			suggestions = []string{
				fullInput + "ask", fullInput + "build",
				fullInput + "diagnose", fullInput + "configure",
			}
		}
	}
	
	// Command-specific contextual suggestions
	switch command {
	case "configure":
		if !strings.Contains(fullInput, "nginx") && !strings.Contains(fullInput, "desktop") {
			suggestions = append(suggestions, 
				fullInput + "nginx",
				fullInput + "desktop", 
				fullInput + "development",
				fullInput + "server",
			)
		}
	case "diagnose":
		if !strings.Contains(fullInput, "boot") && !strings.Contains(fullInput, "services") {
			suggestions = append(suggestions,
				fullInput + "boot",
				fullInput + "services",
				fullInput + "network",
				fullInput + "hardware",
			)
		}
	case "learn":
		if !strings.Contains(fullInput, "basics") && !strings.Contains(fullInput, "advanced") {
			suggestions = append(suggestions,
				fullInput + "basics",
				fullInput + "advanced", 
				fullInput + "expert",
			)
		}
	}
	
	return suggestions
}

// executeCommand executes the given nixai command
func (m *ClaudeTUI) executeCommand(input string) {
	timestamp := time.Now().Format("15:04:05")
	m.addOutput(fmt.Sprintf("[%s] %s", 
		m.styles.timestamp.Render(timestamp),
		m.styles.prompt.Render("$ nixai "+input)))

	// Handle built-in commands
	switch {
	case input == "help":
		m.showHelp()
		return
	case input == "clear":
		m.output = []string{}
		m.addOutput("Terminal cleared")
		return
	case input == "history":
		m.showHistory()
		return
	case strings.HasPrefix(input, "theme"):
		m.handleThemeCommand(input)
		return
	case input == "exit" || input == "quit":
		m.addOutput("Goodbye!")
		return
	}

	// Execute nixai command
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	// Try to execute the actual nixai command
	nixaiPath, pathErr := utils.GetExecutablePath()
	if pathErr != nil {
		m.addOutput(m.styles.error.Render(fmt.Sprintf("Error getting executable path: %v", pathErr)))
		m.addOutput("") // Add empty line for separation
		return
	}
	cmd := exec.Command(nixaiPath, args...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.addOutput(m.styles.error.Render(fmt.Sprintf("Error: %v", err)))
		if len(output) > 0 {
			m.addOutput(m.styles.error.Render(string(output)))
		}
	} else {
		if len(output) > 0 {
			// Split output into lines and add each one
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					m.addOutput(line)
				}
			}
		} else {
			m.addOutput(m.styles.success.Render("Command executed successfully"))
		}
	}
	m.addOutput("") // Add empty line for separation
}

// addOutput adds a line to the output
func (m *ClaudeTUI) addOutput(line string) {
	m.output = append(m.output, line)
	
	// Keep only the last 100 lines to prevent memory issues
	if len(m.output) > 100 {
		m.output = m.output[len(m.output)-100:]
	}
}

// showHelp displays help information
func (m *ClaudeTUI) showHelp() {
	m.addOutput(m.styles.success.Render("NixAI Commands:"))
	m.addOutput("")
	
	commands := getAvailableCommands()
	currentCategory := ""
	
	for _, cmd := range commands {
		if cmd.Category != currentCategory {
			m.addOutput(m.styles.prompt.Render("▶ " + cmd.Category))
			currentCategory = cmd.Category
		}
		m.addOutput(fmt.Sprintf("  %s - %s", cmd.Name, cmd.Description))
	}
	
	m.addOutput("")
	m.addOutput(m.styles.prompt.Render("Built-in Commands:"))
	m.addOutput("  help - Show this help")
	m.addOutput("  clear - Clear terminal")
	m.addOutput("  history - Show command history")
	m.addOutput("  exit/quit - Exit nixai")
	m.addOutput("  theme [name] - Change theme")
	m.addOutput("")
	m.addOutput(m.styles.prompt.Render("Navigation & Completion:"))
	m.addOutput("  ↑/↓ - Navigate history and suggestions")
	m.addOutput("  Tab - Complete suggestion")
	m.addOutput("  Space - Show available flags/options")
	m.addOutput("  Ctrl+C/Esc - Exit")
	m.addOutput("")
	m.addOutput(m.styles.prompt.Render("Smart Completion Features:"))
	m.addOutput("  • Command names: Type partial command names")
	m.addOutput("  • Flags & options: Automatic flag suggestions")
	m.addOutput("  • Values: Context-aware value completion")
	m.addOutput("  • Examples: 'ask --provider ' shows provider options")
	m.addOutput("  • Examples: 'web --port ' shows common port numbers")
	m.addOutput("  • Examples: 'configure ' shows configuration targets")
	m.addOutput("")
	m.addOutput(m.styles.accent.Render("Available Themes:"))
	for name, theme := range themes {
		if name == m.currentTheme {
			m.addOutput(fmt.Sprintf("  %s %s (current)", m.styles.success.Render("●"), theme.Name))
		} else {
			m.addOutput(fmt.Sprintf("  %s %s", m.styles.muted.Render("○"), theme.Name))
		}
	}
	m.addOutput("")
}

// showHistory displays command history
func (m *ClaudeTUI) showHistory() {
	if len(m.commandHistory) == 0 {
		m.addOutput("No command history")
		return
	}
	
	m.addOutput(m.styles.success.Render("Command History:"))
	for i, cmd := range m.commandHistory {
		m.addOutput(fmt.Sprintf("  %d: %s", i+1, cmd))
	}
	m.addOutput("")
}

// handleThemeCommand handles theme switching commands
func (m *ClaudeTUI) handleThemeCommand(input string) {
	parts := strings.Fields(input)
	
	if len(parts) == 1 {
		// Show current theme and available themes
		m.addOutput(m.styles.accent.Render("Current Theme:"))
		m.addOutput(fmt.Sprintf("  %s", themes[m.currentTheme].Name))
		m.addOutput("")
		m.addOutput(m.styles.accent.Render("Available Themes:"))
		for name, theme := range themes {
			if name == m.currentTheme {
				m.addOutput(fmt.Sprintf("  %s %s (current)", m.styles.success.Render("●"), theme.Name))
			} else {
				m.addOutput(fmt.Sprintf("  %s %s", m.styles.muted.Render("○"), theme.Name))
			}
		}
		m.addOutput("")
		m.addOutput(m.styles.muted.Render("Usage: theme [name]"))
		return
	}
	
	if len(parts) >= 2 {
		themeName := parts[1]
		
		// Check if theme exists
		if theme, exists := themes[themeName]; exists {
			// Switch to new theme
			m.currentTheme = themeName
			m.styles = createThemeStyles(theme)
			
			m.addOutput(m.styles.success.Render(fmt.Sprintf("Switched to %s theme", theme.Name)))
			m.addOutput("")
			
			// Show a preview of the new theme colors
			m.addOutput(m.styles.accent.Render("Theme Preview:"))
			m.addOutput(fmt.Sprintf("  %s Primary color", m.styles.prompt.Render("●")))
			m.addOutput(fmt.Sprintf("  %s Success color", m.styles.success.Render("●")))
			m.addOutput(fmt.Sprintf("  %s Warning color", m.styles.warning.Render("●")))
			m.addOutput(fmt.Sprintf("  %s Error color", m.styles.error.Render("●")))
			m.addOutput(fmt.Sprintf("  %s Accent color", m.styles.accent.Render("●")))
			m.addOutput(fmt.Sprintf("  %s Muted color", m.styles.muted.Render("●")))
		} else {
			m.addOutput(m.styles.error.Render(fmt.Sprintf("Theme '%s' not found", themeName)))
			m.addOutput("")
			m.addOutput(m.styles.muted.Render("Available themes:"))
			for _, theme := range themes {
				m.addOutput(fmt.Sprintf("  %s", theme.Name))
			}
		}
	}
	m.addOutput("")
}

// showLogo displays the NixAI ASCII logo on startup
func (m *ClaudeTUI) showLogo() {
	// Logo ASCII art
	logoLines := []string{
		"",
		m.styles.logo.Render("      ███╗   ██╗██╗██╗  ██╗ █████╗ ██╗"),
		m.styles.logo.Render("      ████╗  ██║██║╚██╗██╔╝██╔══██╗██║"),
		m.styles.logo.Render("      ██╔██╗ ██║██║ ╚███╔╝ ███████║██║"),
		m.styles.logo.Render("      ██║╚██╗██║██║ ██╔██╗ ██╔══██║██║"),
		m.styles.logo.Render("      ██║ ╚████║██║██╔╝ ██╗██║  ██║██║"),
		m.styles.logo.Render("      ╚═╝  ╚═══╝╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝"),
		"",
	}
	
	// Create feature box using lipgloss
	featureBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(themes[m.currentTheme].Success)).
		Padding(0, 1).
		Margin(0, 0, 0, 6). // Left margin to center with logo
		Render(
			"AI-Powered NixOS Assistant\n" +
			"Intelligent Configuration\n" +
			"Smart Troubleshooting\n" +
			"Declarative System Management")
	
	// Help text
	helpLines := []string{
		"",
		m.styles.muted.Render("      → Type 'help' to see all available commands"),
		m.styles.muted.Render("      → Start typing for intelligent suggestions"),
		m.styles.muted.Render("      → Use 'theme [name]' to change themes"),
		m.styles.accent.Render(fmt.Sprintf("      → Current theme: %s", themes[m.currentTheme].Name)),
		"",
	}
	
	// Add all logo lines
	for _, line := range logoLines {
		m.addOutput(line)
	}
	
	// Add feature box
	m.addOutput(featureBox)
	
	// Add help lines
	for _, line := range helpLines {
		m.addOutput(line)
	}
}

// View renders the TUI
func (m *ClaudeTUI) View() string {
	var sections []string

	// Header
	header := m.styles.header.Render(fmt.Sprintf("🚀 NixAI Terminal Interface - %s Theme", themes[m.currentTheme].Name))
	sections = append(sections, header)

	// Calculate available height for output
	headerHeight := 3 // Header + borders
	commandBoxHeight := 4 // Command box + borders
	suggestionsHeight := 0
	if m.showSuggestions {
		filtered := m.getFilteredSuggestions()
		suggestionsHeight = len(filtered) + 2 // Suggestions + borders
	}
	
	availableHeight := m.height - headerHeight - commandBoxHeight - suggestionsHeight - 2

	// Output section (command execution results)
	outputLines := m.output
	if len(outputLines) > availableHeight && availableHeight > 0 {
		outputLines = outputLines[len(outputLines)-availableHeight:]
	}

	outputContent := strings.Join(outputLines, "\n")
	if outputContent == "" {
		outputContent = " " // Ensure some content for proper rendering
	}
	
	outputSection := m.styles.output.Render(outputContent)
	sections = append(sections, outputSection)

	// Suggestions section (if visible)
	if m.showSuggestions {
		suggestions := m.renderSuggestions()
		sections = append(sections, suggestions)
	}

	// Command input box at the bottom
	commandContent := m.styles.prompt.Render("nixai > ") + m.textInput.View()
	commandBox := m.styles.commandBox.Width(m.width - 4).Render(commandContent)
	sections = append(sections, commandBox)

	// Version information under command box
	versionText := m.styles.versionInfo.Width(m.width - 2).Render(
		fmt.Sprintf("nixai v%s | Theme: %s | Press Ctrl+C to exit", 
			version.Get().Short(), 
			themes[m.currentTheme].Name))
	sections = append(sections, versionText)

	return strings.Join(sections, "\n")
}

// renderSuggestions renders the suggestions dropdown
func (m *ClaudeTUI) renderSuggestions() string {
	filtered := m.getFilteredSuggestions()
	if len(filtered) == 0 {
		return ""
	}

	var suggestionLines []string
	suggestionLines = append(suggestionLines, m.styles.accent.Render("Suggestions:"))
	
	for i, suggestion := range filtered {
		if i == m.selectedSuggestion {
			suggestionLines = append(suggestionLines, m.styles.selectedSug.Render("  "+suggestion))
		} else {
			suggestionLines = append(suggestionLines, m.styles.suggestion.Render("  "+suggestion))
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(themes[m.currentTheme].Border)).
		Padding(0, 1).
		Margin(0, 1).
		Render(strings.Join(suggestionLines, "\n"))
}

// removeDuplicates removes duplicate strings from a slice while preserving order
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// StartClaudeTUI starts the Claude Code-style TUI
func StartClaudeTUI() error {
	p := tea.NewProgram(NewClaudeTUI(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}