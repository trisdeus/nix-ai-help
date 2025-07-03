package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TUI represents the main terminal user interface
type TUI struct {
	commands    []Command
	selected    int
	searchQuery string
	showHelp    bool
	width       int
	height      int
}

// Command represents a nixai command with metadata
type Command struct {
	Name        string
	Description string
	Category    string
	Usage       string
	Examples    []string
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
		{
			Name:        "ask",
			Description: "Ask AI questions about NixOS",
			Category:    "AI",
			Usage:       "nixai ask \"your question\"",
			Examples:    []string{"nixai ask \"how to configure nginx?\"", "nixai ask \"fix boot issues\""},
		},
		{
			Name:        "build",
			Description: "Build and analyze NixOS configurations",
			Category:    "Build",
			Usage:       "nixai build [options]",
			Examples:    []string{"nixai build", "nixai build --dry-run", "nixai build analyze"},
		},
		{
			Name:        "configure",
			Description: "Interactive NixOS configuration assistant",
			Category:    "Configuration",
			Usage:       "nixai configure [service]",
			Examples:    []string{"nixai configure", "nixai configure nginx", "nixai configure desktop"},
		},
		{
			Name:        "diagnose",
			Description: "Diagnose system issues and problems",
			Category:    "Diagnostics",
			Usage:       "nixai diagnose [component]",
			Examples:    []string{"nixai diagnose", "nixai diagnose boot", "nixai diagnose services"},
		},
		{
			Name:        "flake",
			Description: "Manage NixOS flakes and configurations",
			Category:    "Flakes",
			Usage:       "nixai flake [action]",
			Examples:    []string{"nixai flake create", "nixai flake validate", "nixai flake migrate"},
		},
		{
			Name:        "learn",
			Description: "Interactive NixOS learning modules",
			Category:    "Education",
			Usage:       "nixai learn [module]",
			Examples:    []string{"nixai learn list", "nixai learn basics", "nixai learn progress"},
		},
		{
			Name:        "workflow",
			Description: "Manage automated workflows",
			Category:    "Automation",
			Usage:       "nixai workflow [action]",
			Examples:    []string{"nixai workflow list", "nixai workflow create", "nixai workflow execute"},
		},
		{
			Name:        "plugin",
			Description: "Manage nixai plugins",
			Category:    "Extensibility",
			Usage:       "nixai plugin [action]",
			Examples:    []string{"nixai plugin list", "nixai plugin install", "nixai plugin create"},
		},
		{
			Name:        "version-control",
			Description: "Git-like configuration version control",
			Category:    "Version Control",
			Usage:       "nixai version-control [action]",
			Examples:    []string{"nixai version-control init", "nixai version-control commit", "nixai version-control branch"},
		},
		{
			Name:        "fleet",
			Description: "Manage machine fleet deployments",
			Category:    "Fleet Management",
			Usage:       "nixai fleet [action]",
			Examples:    []string{"nixai fleet list", "nixai fleet deploy", "nixai fleet status"},
		},
		{
			Name:        "team",
			Description: "Team collaboration management",
			Category:    "Collaboration",
			Usage:       "nixai team [action]",
			Examples:    []string{"nixai team create", "nixai team members", "nixai team permissions"},
		},
		{
			Name:        "web",
			Description: "Start the web interface",
			Category:    "Web Interface",
			Usage:       "nixai web start [options]",
			Examples:    []string{"nixai web start", "nixai web start --port 8080", "nixai web start --repo /path/to/repo"},
		},
		{
			Name:        "performance",
			Description: "Performance monitoring and optimization",
			Category:    "Performance",
			Usage:       "nixai performance [action]",
			Examples:    []string{"nixai performance stats", "nixai performance cache", "nixai performance report"},
		},
		{
			Name:        "hardware",
			Description: "Hardware detection and optimization",
			Category:    "Hardware",
			Usage:       "nixai hardware [action]",
			Examples:    []string{"nixai hardware detect", "nixai hardware optimize", "nixai hardware drivers"},
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
		
		// Handle number selection
		if num := parseNumber(input); num >= 0 && num < len(t.commands) {
			t.executeCommand(t.commands[num])
			continue
		}
		
		// Handle command name
		for i, cmd := range t.commands {
			if strings.EqualFold(cmd.Name, input) {
				t.executeCommand(t.commands[i])
				break
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
	instructions := helpStyle.Render("Enter command number/name, 'h' for help, 'q' to quit")
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
	help += "• Enter number (0-13) to select command\n"
	help += "• Enter command name directly\n"
	help += "• 'h' or 'help' to toggle this help\n"
	help += "• 'q', 'quit', or 'exit' to leave\n\n"
	
	help += categoryStyle.Render("Quick Examples:") + "\n"
	help += "• Type '0' or 'ask' to ask AI questions\n"
	help += "• Type '5' or 'learn' for learning modules\n"
	help += "• Type '11' or 'web' to start web interface\n\n"
	
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
	nixaiPath := "./nixai" // Assume nixai binary is in current directory
	cmd := exec.Command(nixaiPath, cmdName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	err := cmd.Run()
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
	nixaiPath := "./nixai"
	execCmd := exec.Command(nixaiPath, args...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin
	
	err := execCmd.Run()
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
	nixaiPath := "./nixai"
	execCmd := exec.Command(nixaiPath, cmd.Name, "--help")
	output, err := execCmd.CombinedOutput()
	if err == nil {
		fmt.Println(categoryStyle.Render("Detailed Command Help:"))
		fmt.Println(string(output))
	}
	
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
}

// parseNumber converts string input to number
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
	default: return -1
	}
}