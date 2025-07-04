package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"nix-ai-help/pkg/utils"
)

// TUI represents the main terminal user interface
type TUI struct {
	commands       []Command
	selected       int
	searchQuery    string
	showHelp       bool
	showSuggestions bool
	suggestions    []CommandSuggestion
	width          int
	height         int
}

// Command represents a nixai command with metadata
type Command struct {
	Name        string
	Description string
	Category    string
	Usage       string
	Examples    []string
}

// CommandSuggestion represents an intelligent command recommendation
type CommandSuggestion struct {
	Command     Command
	Relevance   float64
	Reason      string
	Keywords    []string
	UsageHint   string
}

// NewTUI creates a new TUI instance
func NewTUI() *TUI {
	return &TUI{
		commands: getAvailableCommands(),
		selected: 0,
		width:    80,
		height:   24,
	}
}

// getAvailableCommands returns all available nixai commands
func getAvailableCommands() []Command {
	return []Command{
		// AI & Configuration
		{
			Name:        "ai-config",
			Description: "AI-powered configuration generation and management",
			Category:    "AI",
			Usage:       "nixai ai-config [action]",
			Examples:    []string{"nixai ai-config generate", "nixai ai-config validate", "nixai ai-config optimize"},
		},
		{
			Name:        "ask",
			Description: "Ask AI questions about NixOS",
			Category:    "AI",
			Usage:       "nixai ask \"your question\"",
			Examples:    []string{"nixai ask \"how to configure nginx?\"", "nixai ask \"fix boot issues\""},
		},
		{
			Name:        "configure",
			Description: "Interactive NixOS configuration assistant",
			Category:    "Configuration",
			Usage:       "nixai configure [service]",
			Examples:    []string{"nixai configure", "nixai configure nginx", "nixai configure desktop"},
		},
		{
			Name:        "config",
			Description: "Manage nixai configuration settings",
			Category:    "Configuration",
			Usage:       "nixai config [action]",
			Examples:    []string{"nixai config show", "nixai config set cache.enabled true", "nixai config reset"},
		},
		
		// Build & Development
		{
			Name:        "build",
			Description: "Build and analyze NixOS configurations",
			Category:    "Build",
			Usage:       "nixai build [options]",
			Examples:    []string{"nixai build", "nixai build --dry-run", "nixai build analyze"},
		},
		{
			Name:        "devenv",
			Description: "Create and manage development environments with devenv",
			Category:    "Development",
			Usage:       "nixai devenv [action]",
			Examples:    []string{"nixai devenv create", "nixai devenv shell", "nixai devenv status"},
		},
		{
			Name:        "dev",
			Description: "Developer Experience Revolution - intelligent development environment management",
			Category:    "Development",
			Usage:       "nixai dev [action]",
			Examples:    []string{"nixai dev setup my-app --language go --editor vscode", "nixai dev template list", "nixai dev env create web-app --template react-typescript"},
		},
		{
			Name:        "flake",
			Description: "Manage NixOS flakes and configurations",
			Category:    "Flakes",
			Usage:       "nixai flake [action]",
			Examples:    []string{"nixai flake create", "nixai flake validate", "nixai flake migrate"},
		},
		
		// Diagnostics & Health
		{
			Name:        "diagnose",
			Description: "Diagnose system issues and problems",
			Category:    "Diagnostics",
			Usage:       "nixai diagnose [component]",
			Examples:    []string{"nixai diagnose", "nixai diagnose boot", "nixai diagnose services"},
		},
		{
			Name:        "doctor",
			Description: "Run comprehensive NixOS health checks and diagnostics",
			Category:    "Diagnostics",
			Usage:       "nixai doctor [options]",
			Examples:    []string{"nixai doctor", "nixai doctor --verbose", "nixai doctor --fix"},
		},
		{
			Name:        "health",
			Description: "System health monitoring and prediction",
			Category:    "Monitoring",
			Usage:       "nixai health [action]",
			Examples:    []string{"nixai health status", "nixai health predict", "nixai health monitor"},
		},
		{
			Name:        "logs",
			Description: "Analyze and diagnose NixOS system logs",
			Category:    "Diagnostics",
			Usage:       "nixai logs [options]",
			Examples:    []string{"nixai logs analyze", "nixai logs tail", "nixai logs search"},
		},
		
		// System Management
		{
			Name:        "execute",
			Description: "Execute commands with AI-powered safety validation",
			Category:    "Execution",
			Usage:       "nixai execute [command] [args...]",
			Examples:    []string{"nixai execute nix-env -iA nixpkgs.firefox", "nixai execute --dry-run nixos-rebuild switch"},
		},
		{
			Name:        "gc",
			Description: "AI-powered garbage collection analysis and cleanup",
			Category:    "Maintenance",
			Usage:       "nixai gc [action]",
			Examples:    []string{"nixai gc analyze", "nixai gc collect", "nixai gc optimize"},
		},
		{
			Name:        "hardware",
			Description: "Hardware detection and optimization",
			Category:    "Hardware",
			Usage:       "nixai hardware [action]",
			Examples:    []string{"nixai hardware detect", "nixai hardware optimize", "nixai hardware drivers"},
		},
		{
			Name:        "system-info",
			Description: "System information and health monitoring",
			Category:    "Monitoring",
			Usage:       "nixai system-info [options]",
			Examples:    []string{"nixai system-info", "nixai system-info --detailed", "nixai system-info --json"},
		},
		
		// Intelligence & Analysis
		{
			Name:        "intelligence",
			Description: "AI-powered system intelligence and recommendations",
			Category:    "Intelligence",
			Usage:       "nixai intelligence [action]",
			Examples:    []string{"nixai intelligence analyze", "nixai intelligence predict", "nixai intelligence conflicts"},
		},
		{
			Name:        "context",
			Description: "Manage NixOS system context detection and caching",
			Category:    "Intelligence",
			Usage:       "nixai context [action]",
			Examples:    []string{"nixai context detect", "nixai context cache", "nixai context clear"},
		},
		{
			Name:        "deps",
			Description: "Analyze NixOS configuration dependencies and imports",
			Category:    "Analysis",
			Usage:       "nixai deps [options]",
			Examples:    []string{"nixai deps analyze", "nixai deps graph", "nixai deps validate"},
		},
		
		// Package Management
		{
			Name:        "search",
			Description: "Search for NixOS packages/services and get config/AI tips",
			Category:    "Packages",
			Usage:       "nixai search [query]",
			Examples:    []string{"nixai search firefox", "nixai search --config nginx", "nixai search --ai python"},
		},
		{
			Name:        "package-monitor",
			Description: "Package monitoring and update management",
			Category:    "Packages",
			Usage:       "nixai package-monitor [action]",
			Examples:    []string{"nixai package-monitor status", "nixai package-monitor updates", "nixai package-monitor security"},
		},
		{
			Name:        "package-repo",
			Description: "Analyze Git repositories and generate Nix derivations",
			Category:    "Packages",
			Usage:       "nixai package-repo [repo-url]",
			Examples:    []string{"nixai package-repo https://github.com/user/repo", "nixai package-repo analyze"},
		},
		{
			Name:        "store",
			Description: "Manage, backup, and analyze the Nix store",
			Category:    "Store",
			Usage:       "nixai store [action]",
			Examples:    []string{"nixai store analyze", "nixai store backup", "nixai store optimize"},
		},
		
		// Documentation & Help
		{
			Name:        "explain-option",
			Description: "Explain a NixOS option using AI and documentation",
			Category:    "Documentation",
			Usage:       "nixai explain-option [option]",
			Examples:    []string{"nixai explain-option services.nginx.enable", "nixai explain-option boot.loader.grub"},
		},
		{
			Name:        "explain-home-option",
			Description: "Explain a Home Manager option using AI and documentation",
			Category:    "Documentation",
			Usage:       "nixai explain-home-option [option]",
			Examples:    []string{"nixai explain-home-option programs.git.enable", "nixai explain-home-option home.stateVersion"},
		},
		{
			Name:        "manual",
			Description: "Built-in comprehensive manual system",
			Category:    "Documentation",
			Usage:       "nixai manual [topic]",
			Examples:    []string{"nixai manual", "nixai manual search flakes", "nixai manual configuration"},
		},
		{
			Name:        "learn",
			Description: "Interactive NixOS learning modules",
			Category:    "Education",
			Usage:       "nixai learn [module]",
			Examples:    []string{"nixai learn list", "nixai learn basics", "nixai learn progress"},
		},
		
		// Templates & Snippets
		{
			Name:        "templates",
			Description: "List and manage project templates for NixOS, Home Manager, and related setups",
			Category:    "Templates",
			Usage:       "nixai templates [action]",
			Examples:    []string{"nixai templates list", "nixai templates create", "nixai templates apply"},
		},
		{
			Name:        "snippets",
			Description: "Show, add, or manage code snippets for NixOS, Home Manager, and related workflows",
			Category:    "Templates",
			Usage:       "nixai snippets [action]",
			Examples:    []string{"nixai snippets list", "nixai snippets add", "nixai snippets search"},
		},
		{
			Name:        "import",
			Description: "Import configurations and templates",
			Category:    "Templates",
			Usage:       "nixai import [source]",
			Examples:    []string{"nixai import config.nix", "nixai import --from-url", "nixai import --migrate"},
		},
		
		// Fleet & Machine Management
		{
			Name:        "fleet",
			Description: "Manage machine fleet deployments",
			Category:    "Fleet Management",
			Usage:       "nixai fleet [action]",
			Examples:    []string{"nixai fleet list", "nixai fleet deploy", "nixai fleet status"},
		},
		{
			Name:        "machines",
			Description: "Manage and deploy NixOS configurations across multiple machines",
			Category:    "Fleet Management",
			Usage:       "nixai machines [action]",
			Examples:    []string{"nixai machines list", "nixai machines deploy", "nixai machines add"},
		},
		
		// Automation & Workflows
		{
			Name:        "workflow",
			Description: "Manage automated workflows",
			Category:    "Automation",
			Usage:       "nixai workflow [action]",
			Examples:    []string{"nixai workflow list", "nixai workflow create", "nixai workflow execute"},
		},
		{
			Name:        "migrate",
			Description: "AI-powered migration assistant for channels and flakes",
			Category:    "Migration",
			Usage:       "nixai migrate [action]",
			Examples:    []string{"nixai migrate to-flakes", "nixai migrate channels", "nixai migrate analyze"},
		},
		
		// Collaboration & Team
		{
			Name:        "team",
			Description: "Team collaboration management",
			Category:    "Collaboration",
			Usage:       "nixai team [action]",
			Examples:    []string{"nixai team create", "nixai team members", "nixai team permissions"},
		},
		
		// Version Control
		{
			Name:        "version-control",
			Description: "Git-like configuration version control",
			Category:    "Version Control",
			Usage:       "nixai version-control [action]",
			Examples:    []string{"nixai version-control init", "nixai version-control commit", "nixai version-control branch"},
		},
		
		// Extensions & Integration
		{
			Name:        "plugin",
			Description: "Manage nixai plugins",
			Category:    "Extensibility",
			Usage:       "nixai plugin [action]",
			Examples:    []string{"nixai plugin list", "nixai plugin install", "nixai plugin create"},
		},
		{
			Name:        "mcp-server",
			Description: "Manage the Model Context Protocol (MCP) server",
			Category:    "Integration",
			Usage:       "nixai mcp-server [action]",
			Examples:    []string{"nixai mcp-server start", "nixai mcp-server stop", "nixai mcp-server status"},
		},
		{
			Name:        "neovim-setup",
			Description: "Set up Neovim integration with nixai MCP server",
			Category:    "Integration",
			Usage:       "nixai neovim-setup [options]",
			Examples:    []string{"nixai neovim-setup", "nixai neovim-setup --config-path", "nixai neovim-setup --verify"},
		},
		
		// Web & Interfaces
		{
			Name:        "web",
			Description: "Start the web interface",
			Category:    "Web Interface",
			Usage:       "nixai web start [options]",
			Examples:    []string{"nixai web start", "nixai web start --port 8080", "nixai web start --repo /path/to/repo"},
		},
		
		// Performance & Monitoring
		{
			Name:        "performance",
			Description: "Performance monitoring and optimization",
			Category:    "Performance",
			Usage:       "nixai performance [action]",
			Examples:    []string{"nixai performance stats", "nixai performance cache", "nixai performance report"},
		},
		
		// Error Handling & Support
		{
			Name:        "error",
			Description: "Error handling and analytics management",
			Category:    "Support",
			Usage:       "nixai error [action]",
			Examples:    []string{"nixai error analyze", "nixai error report", "nixai error clear"},
		},
		{
			Name:        "community",
			Description: "Show NixOS community resources and support links",
			Category:    "Support",
			Usage:       "nixai community [topic]",
			Examples:    []string{"nixai community", "nixai community discourse", "nixai community github"},
		},
		
		// Utility
		{
			Name:        "completion",
			Description: "Generate autocompletion scripts for your shell",
			Category:    "Utility",
			Usage:       "nixai completion [shell]",
			Examples:    []string{"nixai completion bash", "nixai completion zsh", "nixai completion fish"},
		},
	}
}

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1A1AA")).
			Padding(0, 1)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#71717A")).
				Italic(true)

	categoryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	usageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Italic(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true)
)

// Start runs the TUI interface
func (t *TUI) Start() error {
	// Simple terminal-based interface for now
	fmt.Print("\033[2J\033[H") // Clear screen
	
	for {
		t.render()
		
		// Get user input
		var input string
		fmt.Print("\n> ")
		fmt.Scanln(&input)
		
		if input == "q" || input == "quit" || input == "exit" {
			break
		}
		
		if input == "h" || input == "help" {
			t.showHelp = !t.showHelp
			continue
		}
		
		// Handle suggestions mode toggle
		if input == "s" || input == "suggest" {
			t.showSuggestions = !t.showSuggestions
			continue
		}
		
		// Handle number selection
		if num := parseNumber(input); num >= 0 && num < len(t.commands) {
			t.executeCommand(t.commands[num])
			continue
		}
		
		// Check for exact command name match first
		exactMatch := false
		for i, cmd := range t.commands {
			if strings.EqualFold(cmd.Name, input) {
				t.executeCommand(t.commands[i])
				exactMatch = true
				break
			}
		}
		
		// If no exact match, check if it looks like a natural language query
		if !exactMatch && t.isNaturalLanguageQuery(input) {
			suggestions := t.intelligentCommandSearch(input)
			if len(suggestions) > 0 {
				t.suggestions = suggestions
				t.showIntelligentSuggestions(input)
				continue
			}
		}
	}
	
	return nil
}

// render displays the TUI interface
func (t *TUI) render() {
	fmt.Print("\033[2J\033[H") // Clear screen
	
	// Title
	title := titleStyle.Render("🚀 NixAI Terminal Interface")
	fmt.Println(title)
	fmt.Println()
	
	// Instructions
	instructions := helpStyle.Render("Enter command number/name, ask questions in natural language, 'h' for help, 'q' to quit")
	fmt.Println(instructions)
	fmt.Println()
	
	// Commands list
	fmt.Println(categoryStyle.Render("Available Commands:"))
	fmt.Println()
	
	currentCategory := ""
	for i, cmd := range t.commands {
		if cmd.Category != currentCategory {
			if currentCategory != "" {
				fmt.Println()
			}
			fmt.Println(categoryStyle.Render(fmt.Sprintf("▶ %s", cmd.Category)))
			currentCategory = cmd.Category
		}
		
		// Command entry
		numberStr := fmt.Sprintf("[%2d]", i)
		nameStr := cmd.Name
		descStr := cmd.Description
		
		if i == t.selected {
			fmt.Printf("  %s %s - %s\n", 
				selectedStyle.Render(numberStr),
				selectedStyle.Render(nameStr),
				selectedStyle.Render(descStr))
		} else {
			fmt.Printf("  %s %s - %s\n",
				normalStyle.Render(numberStr),
				normalStyle.Render(nameStr),
				descriptionStyle.Render(descStr))
		}
	}
	
	// Help section
	if t.showHelp {
		fmt.Println()
		fmt.Println(borderStyle.Render(t.renderHelp()))
	}
}

// renderHelp returns the help content
func (t *TUI) renderHelp() string {
	help := titleStyle.Render("📖 Help & Quick Start") + "\n\n"
	
	help += categoryStyle.Render("Navigation:") + "\n"
	help += "• Enter number (0-43) to select command\n"
	help += "• Enter command name directly\n"
	help += "• Ask questions in natural language (e.g., 'help me with health status')\n"
	help += "• 'h' or 'help' to toggle this help\n"
	help += "• 'q', 'quit', or 'exit' to leave\n\n"
	
	help += categoryStyle.Render("Quick Examples:") + "\n"
	help += "• Type '1' or 'ask' to ask AI questions\n"
	help += "• Type 'help me with health status' for intelligent suggestions\n"
	help += "• Type 'how do I monitor system performance' for AI guidance\n"
	help += "• Type 'web' to start web interface\n\n"
	
	help += categoryStyle.Render("Categories:") + "\n"
	categories := make(map[string][]string)
	for _, cmd := range t.commands {
		categories[cmd.Category] = append(categories[cmd.Category], cmd.Name)
	}
	
	for category, commands := range categories {
		help += fmt.Sprintf("• %s: %s\n", category, strings.Join(commands, ", "))
	}
	
	return help
}

// executeCommand runs the selected command
func (t *TUI) executeCommand(cmd Command) {
	fmt.Print("\033[2J\033[H") // Clear screen
	
	fmt.Println(titleStyle.Render(fmt.Sprintf("🚀 Executing: %s", cmd.Name)))
	fmt.Println()
	
	fmt.Println(categoryStyle.Render("Description:"))
	fmt.Printf("  %s\n\n", cmd.Description)
	
	fmt.Println(categoryStyle.Render("Usage:"))
	fmt.Printf("  %s\n\n", usageStyle.Render(cmd.Usage))
	
	if len(cmd.Examples) > 0 {
		fmt.Println(categoryStyle.Render("Examples:"))
		for _, example := range cmd.Examples {
			fmt.Printf("  %s\n", usageStyle.Render(example))
		}
		fmt.Println()
	}
	
	fmt.Println(helpStyle.Render("Choose an action:"))
	fmt.Println("  [1] Run basic command")
	fmt.Println("  [2] Run with options")
	fmt.Println("  [3] Show detailed help")
	fmt.Println("  [0] Back to main menu")
	fmt.Println()
	
	var choice string
	fmt.Print("> ")
	fmt.Scanln(&choice)
	
	switch choice {
	case "1":
		t.runCommand(cmd.Name)
	case "2":
		t.runCommandWithOptions(cmd)
	case "3":
		t.showDetailedHelp(cmd)
	default:
		return
	}
}

// runCommand executes a basic nixai command
func (t *TUI) runCommand(cmdName string) {
	fmt.Println(categoryStyle.Render(fmt.Sprintf("Running: nixai %s", cmdName)))
	fmt.Println()
	
	// Execute the actual nixai command
	nixaiPath, err := utils.GetExecutablePath()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}
	cmd := exec.Command(nixaiPath, cmdName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
	}
	
	fmt.Println()
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
}

// runCommandWithOptions executes a command with user-specified options
func (t *TUI) runCommandWithOptions(cmd Command) {
	fmt.Printf("Enter options for '%s' (or press Enter for none): ", cmd.Name)
	var options string
	fmt.Scanln(&options)
	
	args := []string{cmd.Name}
	if options != "" {
		args = append(args, strings.Fields(options)...)
	}
	
	fmt.Println(categoryStyle.Render(fmt.Sprintf("Running: nixai %s", strings.Join(args, " "))))
	fmt.Println()
	
	// Execute the actual nixai command
	nixaiPath, err := utils.GetExecutablePath()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}
	execCmd := exec.Command(nixaiPath, args...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin
	
	err = execCmd.Run()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
	}
	
	fmt.Println()
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
}

// showDetailedHelp shows detailed help for a command
func (t *TUI) showDetailedHelp(cmd Command) {
	fmt.Print("\033[2J\033[H") // Clear screen
	
	fmt.Println(titleStyle.Render(fmt.Sprintf("📖 Help: %s", cmd.Name)))
	fmt.Println()
	
	help := borderStyle.Render(fmt.Sprintf(
		"%s\n\n%s\n  %s\n\n%s\n%s",
		categoryStyle.Render("Description:"),
		cmd.Description,
		usageStyle.Render(cmd.Usage),
		categoryStyle.Render("Examples:"),
		strings.Join(cmd.Examples, "\n"),
	))
	
	fmt.Println(help)
	fmt.Println()
	
	// Try to get actual help from nixai command
	nixaiPath, err := utils.GetExecutablePath()
	if err != nil {
		// If we can't get the path, skip the help command
		fmt.Print("Press Enter to continue...")
		fmt.Scanln()
		return
	}
	execCmd := exec.Command(nixaiPath, cmd.Name, "--help")
	output, err := execCmd.CombinedOutput()
	if err == nil {
		fmt.Println(categoryStyle.Render("Detailed Command Help:"))
		fmt.Println(string(output))
	}
	
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
}


// IntelligentCommandSearch analyzes user query and suggests relevant commands (exported for testing)
func (t *TUI) IntelligentCommandSearch(query string) []CommandSuggestion {
	return t.intelligentCommandSearch(query)
}

// intelligentCommandSearch analyzes user query and suggests relevant commands
func (t *TUI) intelligentCommandSearch(query string) []CommandSuggestion {
	query = strings.ToLower(strings.TrimSpace(query))
	var suggestions []CommandSuggestion
	
	// Define keyword mappings for intelligent suggestions
	keywordMappings := map[string][]string{
		// Health & Monitoring
		"health":     {"health", "doctor", "diagnose", "system-info", "performance"},
		"status":     {"health", "system-info", "performance", "fleet"},
		"monitor":    {"health", "system-info", "performance", "package-monitor"},
		"diagnose":   {"diagnose", "doctor", "health", "logs"},
		"doctor":     {"doctor", "health", "diagnose"},
		"check":      {"doctor", "health", "diagnose", "system-info"},
		
		// Configuration
		"config":     {"configure", "config", "ai-config", "explain-option"},
		"configure":  {"configure", "ai-config", "config"},
		"setup":      {"dev", "configure", "neovim-setup", "ai-config"},
		"generate":   {"dev", "configure", "ai-config", "templates"},
		
		// Package Management
		"package":    {"search", "package-monitor", "package-repo", "gc"},
		"packages":   {"search", "package-monitor", "package-repo", "gc"},
		"install":    {"search", "package-repo", "execute"},
		"search":     {"search", "package-repo"},
		"update":     {"package-monitor", "migrate"},
		"upgrade":    {"package-monitor", "migrate"},
		
		// Build & Development
		"build":      {"build", "devenv", "flake"},
		"compile":    {"build", "devenv"},
		"develop":    {"dev", "devenv", "build", "flake"},
		"development": {"dev", "devenv", "build"},
		"environment": {"dev", "devenv", "configure"},
		"project":    {"dev", "devenv"},
		"scaffold":   {"dev", "templates"},
		"deps":       {"dev"},
		"ide":        {"dev"},
		"editor":     {"dev"},
		"vscode":     {"dev"},
		"neovim":     {"dev"},
		"vim":        {"dev"},
		"emacs":      {"dev"},
		
		// Flakes
		"flake":      {"flake", "migrate", "build"},
		"flakes":     {"flake", "migrate", "build"},
		
		// Documentation & Learning
		"help":       {"manual", "learn", "explain-option", "explain-home-option"},
		"learn":      {"learn", "manual", "explain-option"},
		"explain":    {"explain-option", "explain-home-option", "manual"},
		"document":   {"manual", "explain-option", "explain-home-option"},
		"tutorial":   {"learn", "manual"},
		
		// System Management
		"system":     {"system-info", "health", "doctor", "hardware"},
		"hardware":   {"hardware", "system-info", "health"},
		"network":    {"health", "system-info", "diagnose"},
		"performance": {"performance", "health", "system-info"},
		
		// Logs & Errors
		"log":        {"logs", "diagnose", "error"},
		"logs":       {"logs", "diagnose", "error"},
		"error":      {"error", "diagnose", "logs", "doctor"},
		"errors":     {"error", "diagnose", "logs", "doctor"},
		
		// Fleet & Machines
		"fleet":      {"fleet", "machines"},
		"deploy":     {"fleet", "machines", "workflow"},
		"machine":    {"machines", "fleet", "hardware"},
		"machines":   {"machines", "fleet"},
		
		// Storage & Cleanup
		"storage":    {"store", "gc", "performance"},
		"cleanup":    {"gc", "store", "performance"},
		"garbage":    {"gc", "store"},
		"disk":       {"gc", "store", "health", "system-info"},
		
		// Version Control
		"git":        {"version-control", "import"},
		"version":    {"version-control", "migrate"},
		"history":    {"version-control"},
		
		// Templates & Snippets
		"template":   {"dev", "templates", "snippets", "import"},
		"templates":  {"dev", "templates", "snippets", "import"},
		"snippet":    {"snippets", "templates"},
		"snippets":   {"snippets", "templates"},
		
		// Intelligence & Analysis
		"analyze":    {"intelligence", "deps", "diagnose", "performance"},
		"analysis":   {"intelligence", "deps", "diagnose", "performance"},
		"predict":    {"intelligence", "health"},
		"recommend":  {"intelligence", "doctor"},
		"conflict":   {"intelligence", "deps", "diagnose"},
		"conflicts":  {"intelligence", "deps", "diagnose"},
		"dependency": {"deps", "intelligence"},
		"dependencies": {"dev", "deps", "intelligence"},
		
		// Migration & Import
		"migrate":    {"migrate", "import", "flake"},
		"migration":  {"migrate", "import", "flake"},
		"import":     {"import", "migrate", "templates"},
		
		// Workflow & Automation
		"workflow":   {"workflow", "fleet", "execute"},
		"automate":   {"workflow", "execute"},
		"automation": {"workflow", "execute"},
		"execute":    {"execute", "workflow"},
		"run":        {"execute", "build", "workflow"},
		
		// Web Interface
		"web":        {"web"},
		"interface":  {"web", "tui"},
		"ui":         {"web", "tui"},
		
		// Community & Support
		"community":  {"community", "learn", "manual"},
		"support":    {"community", "error", "manual"},
		"forum":      {"community"},
	}
	
	// Score commands based on relevance
	for _, cmd := range t.commands {
		relevance := t.calculateRelevance(query, cmd, keywordMappings)
		if relevance > 0.1 { // Only include if somewhat relevant
			suggestion := CommandSuggestion{
				Command:   cmd,
				Relevance: relevance,
				Reason:    t.generateReason(query, cmd),
				Keywords:  t.extractMatchingKeywords(query, cmd),
				UsageHint: t.generateUsageHint(query, cmd),
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	// Sort by relevance (highest first)
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Relevance > suggestions[i].Relevance {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}
	
	// Return top 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	
	return suggestions
}

// calculateRelevance scores how relevant a command is to the user query
func (t *TUI) calculateRelevance(query string, cmd Command, keywordMappings map[string][]string) float64 {
	query = strings.ToLower(query)
	cmdName := strings.ToLower(cmd.Name)
	cmdDesc := strings.ToLower(cmd.Description)
	
	score := 0.0
	
	// Direct name match gets highest score
	if strings.Contains(query, cmdName) || strings.Contains(cmdName, query) {
		score += 1.0
	}
	
	// Description match gets good score
	queryWords := strings.Fields(query)
	for _, word := range queryWords {
		if strings.Contains(cmdDesc, word) {
			score += 0.3
		}
		if strings.Contains(cmdName, word) {
			score += 0.5
		}
	}
	
	// Keyword mapping match
	for keyword, commands := range keywordMappings {
		if strings.Contains(query, keyword) {
			for _, mappedCmd := range commands {
				if mappedCmd == cmdName {
					score += 0.7
				}
			}
		}
	}
	
	// Category relevance
	categoryName := strings.ToLower(cmd.Category)
	for _, word := range queryWords {
		if strings.Contains(categoryName, word) {
			score += 0.2
		}
	}
	
	return score
}

// generateReason explains why a command was suggested
func (t *TUI) generateReason(query string, cmd Command) string {
	cmdName := cmd.Name
	query = strings.ToLower(query)
	
	// Specific reason patterns
	if strings.Contains(query, "health") && cmdName == "health" {
		return "Perfect match for health monitoring and system status"
	}
	if strings.Contains(query, "status") && cmdName == "health" {
		return "Health command provides comprehensive system status information"
	}
	if strings.Contains(query, "monitor") && cmdName == "health" {
		return "Health command includes real-time monitoring capabilities"
	}
	if strings.Contains(query, "diagnose") && cmdName == "doctor" {
		return "Doctor command provides comprehensive system diagnostics"
	}
	if strings.Contains(query, "config") && cmdName == "configure" {
		return "Configure command helps with NixOS configuration setup"
	}
	if strings.Contains(query, "package") && cmdName == "search" {
		return "Search command helps find and install packages"
	}
	if strings.Contains(query, "build") && cmdName == "build" {
		return "Build command handles NixOS configuration building"
	}
	if strings.Contains(query, "flake") && cmdName == "flake" {
		return "Flake command manages Nix flakes and modern configurations"
	}
	
	// Generic patterns
	if strings.Contains(strings.ToLower(cmd.Description), strings.ToLower(query)) {
		return fmt.Sprintf("Command description matches your query about '%s'", query)
	}
	
	return fmt.Sprintf("Suggested based on relevance to '%s'", query)
}

// extractMatchingKeywords finds which keywords from the query match the command
func (t *TUI) extractMatchingKeywords(query string, cmd Command) []string {
	var keywords []string
	queryWords := strings.Fields(strings.ToLower(query))
	cmdText := strings.ToLower(fmt.Sprintf("%s %s %s", cmd.Name, cmd.Description, cmd.Category))
	
	for _, word := range queryWords {
		if strings.Contains(cmdText, word) && len(word) > 2 { // Skip short words
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// generateUsageHint provides a specific usage suggestion based on the query
func (t *TUI) generateUsageHint(query string, cmd Command) string {
	query = strings.ToLower(query)
	cmdName := cmd.Name
	
	// Specific usage hints
	if strings.Contains(query, "health status") && cmdName == "health" {
		return "Try: nixai health status"
	}
	if strings.Contains(query, "health") && cmdName == "health" {
		return "Try: nixai health status  or  nixai health monitor"
	}
	if strings.Contains(query, "diagnose") && cmdName == "doctor" {
		return "Try: nixai doctor  or  nixai doctor --verbose"
	}
	if strings.Contains(query, "system info") && cmdName == "system-info" {
		return "Try: nixai system-info  or  nixai system-info --detailed"
	}
	if strings.Contains(query, "performance") && cmdName == "performance" {
		return "Try: nixai performance stats  or  nixai performance cache"
	}
	if strings.Contains(query, "package search") && cmdName == "search" {
		return "Try: nixai search firefox  or  nixai search --config nginx"
	}
	
	// Generic hint using first example
	if len(cmd.Examples) > 0 {
		return fmt.Sprintf("Try: %s", cmd.Examples[0])
	}
	
	return fmt.Sprintf("Try: %s", cmd.Usage)
}

// isNaturalLanguageQuery detects if the input looks like a natural language question/request
func (t *TUI) isNaturalLanguageQuery(input string) bool {
	input = strings.ToLower(strings.TrimSpace(input))
	
	// Check for question words
	questionWords := []string{"how", "what", "why", "when", "where", "which", "can", "should", "help"}
	for _, word := range questionWords {
		if strings.HasPrefix(input, word+" ") {
			return true
		}
	}
	
	// Check for help phrases
	helpPhrases := []string{"help me", "i need", "i want", "show me", "find", "search for"}
	for _, phrase := range helpPhrases {
		if strings.Contains(input, phrase) {
			return true
		}
	}
	
	// Check for multiple words (likely a sentence)
	words := strings.Fields(input)
	if len(words) >= 3 {
		return true
	}
	
	// Check for action verbs
	actionVerbs := []string{"monitor", "check", "analyze", "fix", "configure", "install", "build", "deploy"}
	for _, verb := range actionVerbs {
		if strings.Contains(input, verb) {
			return true
		}
	}
	
	return false
}

// showIntelligentSuggestions displays AI-powered command suggestions
func (t *TUI) showIntelligentSuggestions(query string) {
	fmt.Print("\033[2J\033[H") // Clear screen
	
	fmt.Println(titleStyle.Render("🤖 AI-Powered Command Suggestions"))
	fmt.Println()
	fmt.Printf("Query: %s\n", usageStyle.Render(fmt.Sprintf("\"%s\"", query)))
	fmt.Println()
	
	if len(t.suggestions) == 0 {
		fmt.Println(helpStyle.Render("No relevant commands found for your query."))
		fmt.Println()
		fmt.Print("Press Enter to continue...")
		fmt.Scanln()
		return
	}
	
	fmt.Println(categoryStyle.Render("Suggested Commands:"))
	fmt.Println()
	
	for i, suggestion := range t.suggestions {
		// Command header
		relevanceStr := fmt.Sprintf("%.0f%% match", suggestion.Relevance*100)
		fmt.Printf("  %s %s %s\n",
			selectedStyle.Render(fmt.Sprintf("[%d]", i+1)),
			selectedStyle.Render(suggestion.Command.Name),
			normalStyle.Render(fmt.Sprintf("(%s)", relevanceStr)))
		
		// Description
		fmt.Printf("     %s\n", descriptionStyle.Render(suggestion.Command.Description))
		
		// Reason
		fmt.Printf("     💡 %s\n", helpStyle.Render(suggestion.Reason))
		
		// Usage hint
		if suggestion.UsageHint != "" {
			fmt.Printf("     🔧 %s\n", usageStyle.Render(suggestion.UsageHint))
		}
		
		// Keywords if any
		if len(suggestion.Keywords) > 0 {
			fmt.Printf("     🏷️  Keywords: %s\n", 
				normalStyle.Render(strings.Join(suggestion.Keywords, ", ")))
		}
		
		fmt.Println()
	}
	
	fmt.Println(helpStyle.Render("Enter suggestion number to execute, or press Enter to go back"))
	fmt.Print("> ")
	
	var choice string
	fmt.Scanln(&choice)
	
	// Handle suggestion selection
	if num := parseNumber(choice); num > 0 && num <= len(t.suggestions) {
		selectedSuggestion := t.suggestions[num-1]
		t.executeCommand(selectedSuggestion.Command)
	}
}

// updateParseNumber to handle more numbers for suggestions
func parseNumber(input string) int {
	switch input {
	case "0": return 0
	case "1": return 1
	case "2": return 2
	case "3": return 3
	case "4": return 4
	case "5": return 5
	case "6": return 6
	case "7": return 7
	case "8": return 8
	case "9": return 9
	case "10": return 10
	case "11": return 11
	case "12": return 12
	case "13": return 13
	case "14": return 14
	case "15": return 15
	case "16": return 16
	case "17": return 17
	case "18": return 18
	case "19": return 19
	case "20": return 20
	case "21": return 21
	case "22": return 22
	case "23": return 23
	case "24": return 24
	case "25": return 25
	case "26": return 26
	case "27": return 27
	case "28": return 28
	case "29": return 29
	case "30": return 30
	case "31": return 31
	case "32": return 32
	case "33": return 33
	case "34": return 34
	case "35": return 35
	case "36": return 36
	case "37": return 37
	case "38": return 38
	case "39": return 39
	case "40": return 40
	case "41": return 41
	case "42": return 42
	case "43": return 43
	default: return -1
	}
}