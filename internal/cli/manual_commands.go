package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"nix-ai-help/pkg/logger"
)

// ManualContent represents a manual entry
type ManualContent struct {
	Title       string
	Category    string
	Description string
	Content     string
	Examples    []string
	SeeAlso     []string
}

// ManualSystem manages the built-in manual system
type ManualSystem struct {
	entries map[string]*ManualContent
	logger  *logger.Logger
}

// NewManualSystem creates a new manual system
func NewManualSystem(log *logger.Logger) *ManualSystem {
	ms := &ManualSystem{
		entries: make(map[string]*ManualContent),
		logger:  log,
	}
	
	// Initialize all manual entries
	ms.initializeManualEntries()
	
	return ms
}

// CreateManualCommand creates the main manual command
func CreateManualCommand() *cobra.Command {
	manualCmd := &cobra.Command{
		Use:   "manual [topic]",
		Short: "Built-in comprehensive manual system",
		Long: `Built-in comprehensive manual system for nixai.

Access detailed documentation, examples, and help for all nixai commands and concepts.
Navigate through topics, search content, and get contextual help.`,
		Example: `  # Show manual index
  nixai manual

  # Get help for a specific command
  nixai manual ask

  # Browse by category
  nixai manual --category configuration

  # Search manual content
  nixai manual --search "flakes"

  # Interactive manual navigation
  nixai manual --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			ms := NewManualSystem(log)
			
			interactive, _ := cmd.Flags().GetBool("interactive")
			category, _ := cmd.Flags().GetString("category")
			search, _ := cmd.Flags().GetString("search")
			list, _ := cmd.Flags().GetBool("list")
			
			if interactive {
				return ms.runInteractiveManual()
			}
			
			if search != "" {
				return ms.searchManual(search)
			}
			
			if category != "" {
				return ms.showCategory(category)
			}
			
			if list {
				return ms.listAllTopics()
			}
			
			if len(args) == 0 {
				return ms.showManualIndex()
			}
			
			return ms.showManualTopic(args[0])
		},
	}
	
	manualCmd.Flags().BoolP("interactive", "i", false, "Interactive manual navigation")
	manualCmd.Flags().StringP("category", "c", "", "Show topics in specific category")
	manualCmd.Flags().StringP("search", "s", "", "Search manual content")
	manualCmd.Flags().BoolP("list", "l", false, "List all available topics")
	
	return manualCmd
}

// showManualIndex displays the main manual index
func (ms *ManualSystem) showManualIndex() error {
	fmt.Println("📚 nixai Built-in Manual v2.0.0")
	fmt.Println("===============================")
	fmt.Println()
	
	// Group entries by category
	categories := make(map[string][]*ManualContent)
	for _, entry := range ms.entries {
		categories[entry.Category] = append(categories[entry.Category], entry)
	}
	
	// Sort categories
	var categoryNames []string
	for category := range categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)
	
	for _, category := range categoryNames {
		fmt.Printf("## %s\n", category)
		
		// Sort entries within category
		entries := categories[category]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Title < entries[j].Title
		})
		
		for _, entry := range entries {
			fmt.Printf("  %-20s - %s\n", entry.Title, entry.Description)
		}
		fmt.Println()
	}
	
	fmt.Println("📖 Usage:")
	fmt.Println("  nixai manual <topic>          - View specific topic")
	fmt.Println("  nixai manual -i               - Interactive navigation")
	fmt.Println("  nixai manual -c <category>    - Browse by category")
	fmt.Println("  nixai manual -s <term>        - Search manual")
	fmt.Println("  nixai manual -l               - List all topics")
	
	return nil
}

// showManualTopic displays a specific manual topic
func (ms *ManualSystem) showManualTopic(topic string) error {
	entry, exists := ms.entries[topic]
	if !exists {
		// Try to find similar topics
		similar := ms.findSimilarTopics(topic)
		fmt.Printf("❌ Topic '%s' not found.\n\n", topic)
		
		if len(similar) > 0 {
			fmt.Println("🔍 Did you mean:")
			for _, s := range similar {
				fmt.Printf("  - %s\n", s)
			}
		}
		
		fmt.Println("\n💡 Use 'nixai manual -l' to see all available topics")
		return nil
	}
	
	// Display the manual entry
	fmt.Printf("📖 %s\n", entry.Title)
	fmt.Printf("Category: %s\n", entry.Category)
	fmt.Println(strings.Repeat("=", len(entry.Title)+4))
	fmt.Println()
	
	fmt.Println(entry.Content)
	
	if len(entry.Examples) > 0 {
		fmt.Println("\n💡 Examples:")
		fmt.Println("============")
		for i, example := range entry.Examples {
			fmt.Printf("\n%d. %s\n", i+1, example)
		}
	}
	
	if len(entry.SeeAlso) > 0 {
		fmt.Println("\n🔗 See Also:")
		fmt.Println("============")
		for _, seeAlso := range entry.SeeAlso {
			fmt.Printf("  - %s\n", seeAlso)
		}
	}
	
	return nil
}

// showCategory displays all topics in a category
func (ms *ManualSystem) showCategory(category string) error {
	var entries []*ManualContent
	for _, entry := range ms.entries {
		if strings.EqualFold(entry.Category, category) {
			entries = append(entries, entry)
		}
	}
	
	if len(entries) == 0 {
		fmt.Printf("❌ No topics found in category '%s'\n", category)
		fmt.Println("\n📂 Available categories:")
		
		categories := make(map[string]bool)
		for _, entry := range ms.entries {
			categories[entry.Category] = true
		}
		
		var categoryNames []string
		for cat := range categories {
			categoryNames = append(categoryNames, cat)
		}
		sort.Strings(categoryNames)
		
		for _, cat := range categoryNames {
			fmt.Printf("  - %s\n", cat)
		}
		return nil
	}
	
	fmt.Printf("📂 %s Commands\n", category)
	fmt.Println(strings.Repeat("=", len(category)+10))
	fmt.Println()
	
	// Sort entries
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Title < entries[j].Title
	})
	
	for _, entry := range entries {
		fmt.Printf("📖 %s\n", entry.Title)
		fmt.Printf("   %s\n", entry.Description)
		fmt.Printf("   Usage: nixai manual %s\n\n", strings.ToLower(entry.Title))
	}
	
	return nil
}

// searchManual searches through manual content
func (ms *ManualSystem) searchManual(searchTerm string) error {
	searchTerm = strings.ToLower(searchTerm)
	var results []*ManualContent
	
	for _, entry := range ms.entries {
		// Search in title, description, and content
		if strings.Contains(strings.ToLower(entry.Title), searchTerm) ||
		   strings.Contains(strings.ToLower(entry.Description), searchTerm) ||
		   strings.Contains(strings.ToLower(entry.Content), searchTerm) {
			results = append(results, entry)
		}
	}
	
	if len(results) == 0 {
		fmt.Printf("🔍 No results found for '%s'\n", searchTerm)
		return nil
	}
	
	fmt.Printf("🔍 Search Results for '%s' (%d found)\n", searchTerm, len(results))
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()
	
	for _, result := range results {
		fmt.Printf("📖 %s (%s)\n", result.Title, result.Category)
		fmt.Printf("   %s\n", result.Description)
		fmt.Printf("   Usage: nixai manual %s\n\n", strings.ToLower(result.Title))
	}
	
	return nil
}

// listAllTopics lists all available topics
func (ms *ManualSystem) listAllTopics() error {
	var topics []string
	for topic := range ms.entries {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	
	fmt.Printf("📚 All Manual Topics (%d total)\n", len(topics))
	fmt.Println("================================")
	fmt.Println()
	
	for i, topic := range topics {
		if i%3 == 0 && i > 0 {
			fmt.Println()
		}
		fmt.Printf("%-25s", topic)
		if (i+1)%3 == 0 {
			fmt.Println()
		}
	}
	
	if len(topics)%3 != 0 {
		fmt.Println()
	}
	
	fmt.Println("\n💡 Use 'nixai manual <topic>' to view detailed help")
	
	return nil
}

// findSimilarTopics finds topics similar to the search term
func (ms *ManualSystem) findSimilarTopics(searchTerm string) []string {
	var similar []string
	searchTerm = strings.ToLower(searchTerm)
	
	for topic := range ms.entries {
		topicLower := strings.ToLower(topic)
		
		// Check for partial matches
		if strings.Contains(topicLower, searchTerm) || strings.Contains(searchTerm, topicLower) {
			similar = append(similar, topic)
		}
	}
	
	// Limit to 5 suggestions
	if len(similar) > 5 {
		similar = similar[:5]
	}
	
	return similar
}

// runInteractiveManual starts the interactive manual navigation
func (ms *ManualSystem) runInteractiveManual() error {
	fmt.Println("🚀 Interactive Manual Mode")
	fmt.Println("===========================")
	fmt.Println("Commands: help, list, search <term>, <topic>, quit")
	fmt.Println()
	
	for {
		fmt.Print("manual> ")
		
		var input string
		fmt.Scanln(&input)
		
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		
		switch {
		case input == "quit" || input == "exit" || input == "q":
			fmt.Println("👋 Goodbye!")
			return nil
			
		case input == "help" || input == "h":
			fmt.Println("📚 Interactive Manual Commands:")
			fmt.Println("  help           - Show this help")
			fmt.Println("  list           - List all topics")
			fmt.Println("  search <term>  - Search manual")
			fmt.Println("  <topic>        - View specific topic")
			fmt.Println("  quit           - Exit manual")
			fmt.Println()
			
		case input == "list" || input == "l":
			ms.listAllTopics()
			
		case strings.HasPrefix(input, "search "):
			searchTerm := strings.TrimPrefix(input, "search ")
			ms.searchManual(searchTerm)
			
		default:
			if _, exists := ms.entries[input]; exists {
				ms.showManualTopic(input)
			} else {
				fmt.Printf("❌ Unknown topic or command: %s\n", input)
				fmt.Println("💡 Use 'help' for commands or 'list' for topics")
			}
		}
		
		fmt.Println()
	}
}

// initializeManualEntries initializes all manual entries
func (ms *ManualSystem) initializeManualEntries() {
	// Core Commands
	ms.addManualEntry("ask", &ManualContent{
		Title:       "ask - AI Question Interface",
		Category:    "Core Commands",
		Description: "Ask AI-powered questions about NixOS",
		Content: `The 'ask' command provides a direct interface to AI-powered assistance for NixOS questions.

Usage:
  nixai ask [question]
  nixai -a "[question]"

Features:
- Context-aware responses based on your system configuration
- Support for multiple AI providers (Ollama, OpenAI, Gemini, Claude, Groq)
- Role-based responses for specialized expertise
- Integration with system context detection

The AI analyzes your question along with your system context to provide
personalized, actionable answers for your specific NixOS setup.`,
		Examples: []string{
			`nixai ask "How do I enable SSH?"`,
			`nixai -a "Debug my failing build" --agent diagnose`,
			`nixai ask "Configure firewall for web server" --role security-expert`,
			`nixai ask "Optimize boot time" --provider gemini`,
		},
		SeeAlso: []string{"diagnose", "context", "agents", "providers"},
	})

	ms.addManualEntry("web", &ManualContent{
		Title:       "web - Modern Web Interface",
		Category:    "Interface",
		Description: "Launch the modern web dashboard interface",
		Content: `The 'web' command starts a comprehensive web interface for managing NixOS configurations.

Usage:
  nixai web start [options]

Features:
- Responsive dashboard with real-time monitoring
- Visual configuration builder with syntax highlighting
- Team collaboration and approval workflows
- Package management and service control
- Hardware information and optimization
- Build system monitoring and log viewing

The web interface provides a graphical alternative to command-line operations,
making nixai accessible to users who prefer visual interfaces.`,
		Examples: []string{
			`nixai web start`,
			`nixai web start --port 8080 --host 0.0.0.0`,
			`nixai web start --dev --verbose`,
			`nixai web start --auth --config web-config.yaml`,
		},
		SeeAlso: []string{"tui", "interactive", "config"},
	})

	ms.addManualEntry("tui", &ManualContent{
		Title:       "tui - Enhanced Terminal Interface",
		Category:    "Interface",
		Description: "Launch the enhanced terminal user interface",
		Content: `The 'tui' command provides an enhanced terminal user interface with two modes.

Usage:
  nixai tui              # Modern Claude Code-style interface
  nixai tui --classic    # Classic menu-based interface
  nixai interactive      # Alias for 'nixai tui'

Modern Interface Features:
- Dual-panel layout with command browser and execution area
- Real-time command execution with progress indicators
- Enhanced navigation with keyboard shortcuts
- Syntax highlighting and formatted output

Classic Interface Features:
- Traditional menu-driven navigation
- Wide terminal compatibility
- Screen reader friendly
- Minimal resource usage

Both interfaces provide full access to all nixai functionality through
an interactive, keyboard-driven experience.`,
		Examples: []string{
			`nixai tui`,
			`nixai tui --classic`,
			`nixai interactive`,
			`NIXAI_TUI_CLASSIC=1 nixai tui`,
		},
		SeeAlso: []string{"web", "interactive", "help"},
	})

	ms.addManualEntry("plugin", &ManualContent{
		Title:       "plugin - Plugin System Management",
		Category:    "Extension",
		Description: "Manage the comprehensive plugin system",
		Content: `The 'plugin' command manages nixai's comprehensive plugin system.

Usage:
  nixai plugin <subcommand> [options]

Available Subcommands:
- list                  List installed plugins
- search <query>        Search for available plugins
- install <plugin>      Install a plugin
- uninstall <plugin>    Remove a plugin
- enable/disable <plugin> Control plugin state
- info <plugin>         Show plugin information
- status <plugin>       Check plugin health
- execute <plugin> <op> Execute plugin operations
- create <template> <dir> Create new plugin
- validate <path>       Validate plugin files
- discover              Discover available plugins
- metrics [plugin]      View plugin performance
- events                Monitor plugin events

The plugin system provides secure, dynamically loadable extensions
with sandbox security, community marketplace, and development tools.`,
		Examples: []string{
			`nixai plugin list`,
			`nixai plugin search system-monitor`,
			`nixai plugin install system-info`,
			`nixai plugin create basic-go my-plugin`,
			`nixai plugin execute system-info get-cpu`,
		},
		SeeAlso: []string{"integration", "development"},
	})

	ms.addManualEntry("fleet", &ManualContent{
		Title:       "fleet - Fleet Management System",
		Category:    "Management",
		Description: "Manage multiple NixOS machines as a unified fleet",
		Content: `The 'fleet' command provides comprehensive multi-machine management.

Usage:
  nixai fleet <subcommand> [options]

Key Subcommands:
- list                     List fleet machines
- add-machine <hostname>   Add machine to fleet
- remove-machine <hostname> Remove machine from fleet
- health                   Check fleet health status
- monitor                  Real-time fleet monitoring
- deploy [target]          Deploy configurations
- deployment <action>      Manage deployments

Fleet Features:
- Centralized configuration management
- Rolling deployment strategies
- Health monitoring and alerting
- SSH key management and distribution
- Role-based machine grouping
- Deployment history and rollback

Perfect for managing development clusters, production environments,
or mixed heterogeneous infrastructures.`,
		Examples: []string{
			`nixai fleet list`,
			`nixai fleet add-machine web-01 --role web-server`,
			`nixai fleet deploy --strategy rolling --batch-size 2`,
			`nixai fleet health --detailed`,
			`nixai fleet monitor --follow`,
		},
		SeeAlso: []string{"version-control", "deploy", "machines"},
	})

	ms.addManualEntry("version-control", &ManualContent{
		Title:       "version-control - Configuration Version Control",
		Category:    "Management",
		Description: "Git-like version control for NixOS configurations",
		Content: `The 'version-control' command provides Git-like version control for configurations.

Usage:
  nixai version-control <subcommand> [options]

Key Features:
- Configuration repository initialization
- Commit tracking and history management
- Branch-based development workflows
- Team collaboration with role-based access
- Merge conflict resolution
- Configuration validation and hooks

Subcommands:
- init [--path] [--remote-url]  Initialize repository
- commit [--message]            Create commit
- history [--limit]             Show commit history
- branch <action> [name]        Branch management
- team <action> [name]          Team management

Brings software development best practices to system administration,
enabling collaborative infrastructure management with full history tracking.`,
		Examples: []string{
			`nixai version-control init --remote-url git@github.com:company/nixos`,
			`nixai version-control commit --message "Enable SSH service"`,
			`nixai version-control branch create feature/web-server`,
			`nixai version-control team create dev-team`,
		},
		SeeAlso: []string{"fleet", "config", "team-collaboration"},
	})

	// Add comprehensive entries for all commands
	ms.addCoreCommandEntries()
	ms.addSystemCommandEntries()
	ms.addDevelopmentCommandEntries()
	ms.addConfigurationCommandEntries()
	ms.addManagementCommandEntries()
	ms.addMonitoringCommandEntries()
	ms.addConceptEntries()
}

// addCoreCommandEntries adds manual entries for core commands
func (ms *ManualSystem) addCoreCommandEntries() {
	ms.addManualEntry("hardware", &ManualContent{
		Title:       "hardware - Hardware Management",
		Category:    "System",
		Description: "Comprehensive hardware detection and optimization",
		Content: `The 'hardware' command provides comprehensive hardware management capabilities.

Subcommands:
- detect      Comprehensive hardware analysis
- optimize    AI-powered optimization recommendations
- drivers     Driver and firmware management
- laptop      Laptop-specific optimizations
- compare     Compare current vs optimal settings
- function    Advanced hardware function calling

Features:
- Automatic hardware detection and identification
- Performance optimization recommendations
- Driver installation and configuration
- Power management for laptops
- Hardware compatibility checking
- Thermal management and monitoring`,
		Examples: []string{
			`nixai hardware detect`,
			`nixai hardware optimize --dry-run`,
			`nixai hardware drivers --auto-install`,
			`nixai hardware laptop --power-save`,
		},
		SeeAlso: []string{"diagnose", "performance", "system-info"},
	})

	ms.addManualEntry("diagnose", &ManualContent{
		Title:       "diagnose - System Diagnostics",
		Category:    "System",
		Description: "AI-powered system diagnostics and troubleshooting",
		Content: `The 'diagnose' command provides intelligent system diagnostics.

Features:
- Multi-format log analysis (systemd, kernel, applications)
- AI-powered issue detection and resolution suggestions
- Build failure analysis with pattern recognition
- Configuration validation and error detection
- Real-time system health monitoring

The diagnose system uses AI to analyze logs, identify patterns,
and provide actionable solutions for common NixOS issues.`,
		Examples: []string{
			`nixai diagnose`,
			`nixai diagnose /var/log/nixos-rebuild.log`,
			`journalctl -xe | nixai diagnose`,
			`nixai diagnose --type system --context "boot failure"`,
		},
		SeeAlso: []string{"doctor", "logs", "error", "build"},
	})

	// Add more core commands...
	ms.addManualEntry("build", &ManualContent{
		Title:       "build - Build System Management",
		Category:    "Development",
		Description: "Advanced build troubleshooting and optimization",
		Content: `The 'build' command provides comprehensive build system management.

Subcommands:
- debug          Deep build failure analysis
- retry          Intelligent retry with fixes
- cache-miss     Analyze cache miss reasons
- environment    Build environment analysis
- dependencies   Dependency conflict resolution
- performance    Build performance optimization
- cleanup        Build cache cleanup
- validate       Build configuration validation
- monitor        Real-time build monitoring
- compare        Build configuration comparison

Advanced Features:
- Pattern recognition for common build failures
- Automated fix suggestions and application
- Cache optimization and management
- Dependency conflict resolution
- Performance profiling and optimization`,
		Examples: []string{
			`nixai build debug`,
			`nixai build retry --smart-cache`,
			`nixai build performance --profile`,
			`nixai build cleanup --aggressive`,
		},
		SeeAlso: []string{"diagnose", "deps", "performance"},
	})
}

// addSystemCommandEntries adds entries for system monitoring commands
func (ms *ManualSystem) addSystemCommandEntries() {
	ms.addManualEntry("performance", &ManualContent{
		Title:       "performance - Performance Monitoring",
		Category:    "Monitoring",
		Description: "System performance monitoring and optimization",
		Content: `The 'performance' command provides comprehensive performance monitoring.

Features:
- Real-time system metrics monitoring
- Performance analysis and bottleneck identification
- Optimization recommendations
- Benchmark testing and comparison
- Historical performance tracking

Subcommands:
- overview       System performance overview
- monitor        Real-time monitoring
- analyze        Performance analysis
- benchmark      System benchmarking
- optimize       Optimization suggestions
- history        Performance history`,
		Examples: []string{
			`nixai performance overview --real-time`,
			`nixai performance monitor --metrics cpu,memory`,
			`nixai performance benchmark --suite full`,
			`nixai performance optimize --category boot`,
		},
		SeeAlso: []string{"hardware", "system-info", "monitor"},
	})

	ms.addManualEntry("system-info", &ManualContent{
		Title:       "system-info - System Information",
		Category:    "Information",
		Description: "Comprehensive system information display",
		Content: `The 'system-info' command displays detailed system information.

Features:
- Hardware specifications and capabilities
- Software environment details
- Network configuration information
- Security status and configuration
- Resource usage and availability

Information Categories:
- Hardware (CPU, memory, storage, graphics)
- Software (kernel, packages, services)
- Network (interfaces, configuration, connectivity)
- Security (firewall, users, permissions)`,
		Examples: []string{
			`nixai system-info`,
			`nixai system-info --detailed --include-hardware`,
			`nixai system-info --category network --verbose`,
			`nixai system-info --export --output system-report.json`,
		},
		SeeAlso: []string{"hardware", "performance", "diagnose"},
	})
}

// addConceptEntries adds entries for important concepts
func (ms *ManualSystem) addConceptEntries() {
	ms.addManualEntry("agents", &ManualContent{
		Title:       "AI Agents - Specialized AI Behavior",
		Category:    "Concepts",
		Description: "Understanding nixai's AI agent system",
		Content: `AI Agents provide specialized behavior and expertise domains for different types of assistance.

Available Agents:
- diagnose         System troubleshooting specialist
- security         Security hardening expert
- performance      Performance optimization specialist
- development      Development environment expert
- hardware         Hardware configuration specialist
- learning         Educational guidance provider

Agent Selection:
Use the --agent flag to specify which agent should handle your request.
Agents have specialized knowledge and provide focused assistance.

Role Integration:
Agents work with roles to provide even more specialized responses.
Combine agents with roles for maximum effectiveness.`,
		Examples: []string{
			`nixai ask "Debug build failure" --agent diagnose`,
			`nixai ask "Harden system security" --agent security`,
			`nixai ask "Optimize performance" --agent performance`,
		},
		SeeAlso: []string{"ask", "roles", "providers"},
	})

	ms.addManualEntry("providers", &ManualContent{
		Title:       "AI Providers - Multiple AI Backend Support",
		Category:    "Concepts",
		Description: "Understanding and configuring AI providers",
		Content: `nixai supports multiple AI providers for maximum flexibility and choice.

Supported Providers:
- Ollama (default)     Local inference, privacy-first
- OpenAI               Industry-leading performance
- Gemini               Advanced reasoning capabilities
- Claude               Constitutional AI approach
- Groq                 Ultra-fast inference
- LlamaCpp             CPU-optimized local inference
- GitHub Copilot       GitHub integration

Provider Configuration:
All providers are configured through the user config file (~/.config/nixai/config.yaml).
Set your preferred provider and model for different use cases.

Quality Recommendations:
- OpenAI/Claude: Best for complex NixOS tasks
- Gemini: Good accuracy with strong reasoning
- Groq: Fast iteration and development
- Ollama: Privacy-first local inference`,
		Examples: []string{
			`# Configure in ~/.config/nixai/config.yaml`,
			`ai_provider: gemini`,
			`ai_model: gemini-2.5-pro`,
			`export GEMINI_API_KEY="your-key"`,
		},
		SeeAlso: []string{"ask", "config", "agents"},
	})

	ms.addManualEntry("getting-started", &ManualContent{
		Title:       "Getting Started - nixai Basics",
		Category:    "Getting Started",
		Description: "Quick start guide for new users",
		Content: `Welcome to nixai! Here's how to get started with your AI-powered NixOS assistant.

Quick Start:
1. Ask a question: nixai ask "How do I enable SSH?"
2. Launch TUI: nixai tui
3. Get system info: nixai system-info
4. Run diagnostics: nixai doctor

Key Commands to Try:
- nixai ask "question"     Ask AI for help
- nixai tui               Interactive interface
- nixai web start         Web dashboard
- nixai hardware detect   Hardware analysis
- nixai diagnose          System diagnostics

Configuration:
- System config: /etc/nixai/config.yaml
- User config: ~/.config/nixai/config.yaml
- AI providers: Multiple options available

Getting Help:
- nixai manual            This built-in manual
- nixai help             Command-specific help
- nixai --help           General help
- nixai community        Community resources`,
		Examples: []string{
			`nixai ask "Enable Bluetooth"`,
			`nixai tui`,
			`nixai doctor`,
			`nixai manual ask`,
		},
		SeeAlso: []string{"ask", "tui", "doctor", "config"},
	})
}

// addDevelopmentCommandEntries adds entries for development commands
func (ms *ManualSystem) addDevelopmentCommandEntries() {
	ms.addManualEntry("flake", &ManualContent{
		Title:       "flake - Flake Management",
		Category:    "Development",
		Description: "Comprehensive Nix flake management and operations",
		Content: `The 'flake' command provides complete flake lifecycle management.

Features:
- Flake initialization and template creation
- Dependency management and updates
- Flake validation and checking
- Development shell creation
- Build and deployment operations

Subcommands:
- init         Initialize new flake project
- update       Update flake dependencies
- check        Validate flake configuration
- show         Display flake information
- develop      Enter development shell`,
		Examples: []string{
			`nixai flake init`,
			`nixai flake update --inputs nixpkgs`,
			`nixai flake check --verbose`,
			`nixai flake develop`,
		},
		SeeAlso: []string{"devenv", "build", "templates"},
	})

	ms.addManualEntry("devenv", &ManualContent{
		Title:       "devenv - Development Environments",
		Category:    "Development",
		Description: "Create and manage development environments",
		Content: `The 'devenv' command manages project-specific development environments.

Features:
- Language-specific environment templates
- Tool and dependency management
- Shell integration and activation
- Environment isolation and reproducibility
- Cross-platform compatibility

Supported Languages:
- Python, Node.js, Rust, Go, Java
- C/C++, Ruby, PHP, Haskell
- Custom environments with Nix expressions`,
		Examples: []string{
			`nixai devenv create python`,
			`nixai devenv create node --version 18`,
			`nixai devenv activate my-project`,
			`nixai devenv list`,
		},
		SeeAlso: []string{"flake", "templates", "build"},
	})

	ms.addManualEntry("package-repo", &ManualContent{
		Title:       "package-repo - Package Repository Analysis",
		Category:    "Development",
		Description: "AI-powered analysis and packaging of Git repositories",
		Content: `The 'package-repo' command analyzes repositories and generates Nix packages.

Features:
- Automatic language detection and analysis
- Dependency analysis and resolution
- Nix derivation generation
- Build system detection
- Security vulnerability scanning

Supported Sources:
- GitHub, GitLab, and other Git repositories
- Local project directories
- Archive files and tarballs
- Package registries (npm, PyPI, crates.io)`,
		Examples: []string{
			`nixai package-repo https://github.com/user/project`,
			`nixai package-repo --local ./my-project`,
			`nixai package-repo <url> --output derivation.nix`,
			`nixai package-repo <url> --analyze-only`,
		},
		SeeAlso: []string{"flake", "build", "devenv"},
	})
}

// addConfigurationCommandEntries adds entries for configuration commands
func (ms *ManualSystem) addConfigurationCommandEntries() {
	ms.addManualEntry("config", &ManualContent{
		Title:       "config - Configuration Management",
		Category:    "Configuration",
		Description: "Manage nixai configuration settings",
		Content: `The 'config' command manages nixai's configuration system.

Features:
- View and modify configuration settings
- Provider and model management
- Reset to default configurations
- Export and import configurations
- Validation and troubleshooting

Configuration Categories:
- AI providers and models
- System paths and directories
- Logging and debugging settings
- Feature flags and experimental options`,
		Examples: []string{
			`nixai config show`,
			`nixai config set ai.provider gemini`,
			`nixai config reset`,
			`nixai config validate`,
		},
		SeeAlso: []string{"providers", "configure", "setup"},
	})

	ms.addManualEntry("configure", &ManualContent{
		Title:       "configure - Interactive Configuration",
		Category:    "Configuration",
		Description: "Interactive NixOS configuration generation",
		Content: `The 'configure' command provides guided configuration creation.

Features:
- Natural language configuration generation
- Template-based configuration building
- Interactive parameter input
- Configuration validation and optimization
- Security and best practices enforcement

Generation Modes:
- AI-powered: Natural language to configuration
- Template-based: Common scenario templates
- Interactive: Guided question-answer workflow
- Hybrid: Combination of approaches`,
		Examples: []string{
			`nixai configure`,
			`nixai configure --search "web server nginx"`,
			`nixai configure --desktop --advanced`,
			`nixai configure --output my-config.nix`,
		},
		SeeAlso: []string{"templates", "config", "build"},
	})

	ms.addManualEntry("templates", &ManualContent{
		Title:       "templates - Configuration Templates",
		Category:    "Configuration",
		Description: "Manage and use configuration templates",
		Content: `The 'templates' command manages reusable configuration templates.

Features:
- Browse available templates
- Create custom templates
- Template parameterization
- Category-based organization
- Community template sharing

Template Categories:
- Desktop environments (GNOME, KDE, XFCE)
- Server configurations (web, database, mail)
- Development setups (languages, tools)
- Security hardening templates
- Gaming and multimedia setups`,
		Examples: []string{
			`nixai templates list`,
			`nixai templates show desktop-gnome`,
			`nixai templates apply web-server`,
			`nixai templates create my-template`,
		},
		SeeAlso: []string{"configure", "snippets", "config"},
	})

	ms.addManualEntry("snippets", &ManualContent{
		Title:       "snippets - Configuration Snippets",
		Category:    "Configuration",
		Description: "Manage reusable configuration snippets",
		Content: `The 'snippets' command manages small, reusable configuration pieces.

Features:
- Browse snippet library
- Search by functionality
- Custom snippet creation
- Copy and paste integration
- Category-based organization

Snippet Types:
- Service configurations
- Package installations
- Environment settings
- Hardware configurations
- Security settings`,
		Examples: []string{
			`nixai snippets search "graphics"`,
			`nixai snippets show ssh-config`,
			`nixai snippets add my-snippet`,
			`nixai snippets list --category services`,
		},
		SeeAlso: []string{"templates", "configure", "examples"},
	})
}

// addManagementCommandEntries adds entries for management commands
func (ms *ManualSystem) addManagementCommandEntries() {
	ms.addManualEntry("machines", &ManualContent{
		Title:       "machines - Multi-machine Management",
		Category:    "Management",
		Description: "Manage multiple NixOS machines with flake-based deployment",
		Content: `The 'machines' command provides multi-machine configuration management.

Features:
- Flake-based machine definitions
- Remote deployment and management
- Configuration synchronization
- Host-specific customizations
- Network topology management

Machine Types:
- Physical servers and workstations
- Virtual machines and containers
- Cloud instances (AWS, GCP, Azure)
- Development and testing environments
- IoT and embedded devices`,
		Examples: []string{
			`nixai machines list`,
			`nixai machines deploy my-machine`,
			`nixai machines show my-machine`,
			`nixai machines add-host server01`,
		},
		SeeAlso: []string{"fleet", "deploy", "flake"},
	})

	ms.addManualEntry("migrate", &ManualContent{
		Title:       "migrate - Migration Assistant",
		Category:    "Management",
		Description: "AI-powered migration assistance for NixOS",
		Content: `The 'migrate' command provides intelligent migration assistance.

Features:
- Channels to flakes migration
- Configuration modernization
- System upgrade assistance
- Backup and rollback support
- Compatibility checking

Migration Types:
- Channel-based to flake-based systems
- Legacy to modern NixOS configurations
- Package format migrations
- Service configuration updates
- Hardware platform migrations`,
		Examples: []string{
			`nixai migrate analyze`,
			`nixai migrate to-flakes`,
			`nixai migrate --backup-name "pre-migration"`,
			`nixai migrate to-flakes --dry-run`,
		},
		SeeAlso: []string{"flake", "backup", "version-control"},
	})

	ms.addManualEntry("store", &ManualContent{
		Title:       "store - Nix Store Management",
		Category:    "Management",
		Description: "Nix store analysis, optimization, and management",
		Content: `The 'store' command provides comprehensive Nix store management.

Features:
- Store usage analysis and reporting
- Cleanup and optimization tools
- Backup and restore operations
- Store verification and repair
- Performance optimization

Operations:
- Size analysis and breakdown
- Dependency tracking
- Orphaned package removal
- Store integrity checking
- Cache management`,
		Examples: []string{
			`nixai store analyze`,
			`nixai store cleanup --dry-run`,
			`nixai store backup --output store-backup.tar`,
			`nixai store verify`,
		},
		SeeAlso: []string{"gc", "performance", "backup"},
	})

	ms.addManualEntry("gc", &ManualContent{
		Title:       "gc - Garbage Collection",
		Category:    "Management",
		Description: "AI-guided garbage collection and cleanup",
		Content: `The 'gc' command provides intelligent garbage collection management.

Features:
- AI-powered cleanup analysis
- Safe generation management
- Disk usage optimization
- Custom retention policies
- Impact assessment before cleanup

Subcommands:
- analyze           Analyze cleanup opportunities
- safe-clean        AI-guided safe cleanup
- compare-generations Compare system generations
- disk-usage        Visualize store usage

Safety Features:
- Generation preservation logic
- Rollback capability protection
- Critical path analysis
- User confirmation prompts`,
		Examples: []string{
			`nixai gc analyze`,
			`nixai gc safe-clean --keep-generations 10`,
			`nixai gc compare-generations`,
			`nixai gc disk-usage --detailed`,
		},
		SeeAlso: []string{"store", "performance", "cleanup"},
	})
}

// addMonitoringCommandEntries adds entries for monitoring commands
func (ms *ManualSystem) addMonitoringCommandEntries() {
	ms.addManualEntry("doctor", &ManualContent{
		Title:       "doctor - Health Checks",
		Category:    "Monitoring",
		Description: "Comprehensive system health diagnostics",
		Content: `The 'doctor' command performs comprehensive system health checks.

Health Check Categories:
- System configuration validation
- Service status and health
- Resource usage and availability
- Security configuration review
- Performance analysis
- Network connectivity tests

Diagnostic Features:
- Automated issue detection
- Resolution recommendations
- System optimization suggestions
- Security vulnerability scanning
- Configuration best practices review`,
		Examples: []string{
			`nixai doctor`,
			`nixai doctor --detailed`,
			`nixai doctor --category security`,
			`nixai doctor --fix-issues`,
		},
		SeeAlso: []string{"diagnose", "performance", "security"},
	})

	ms.addManualEntry("logs", &ManualContent{
		Title:       "logs - Log Analysis",
		Category:    "Monitoring",
		Description: "AI-powered log analysis and troubleshooting",
		Content: `The 'logs' command provides intelligent log analysis capabilities.

Features:
- Multi-format log parsing (systemd, syslog, application)
- Pattern recognition and anomaly detection
- Error correlation and root cause analysis
- Real-time log monitoring
- Historical log analysis

Log Sources:
- System logs (systemd, kernel)
- Application logs
- Build logs (nixos-rebuild)
- Service-specific logs
- Custom log files`,
		Examples: []string{
			`nixai logs analyze`,
			`nixai logs --follow --service nginx`,
			`nixai logs --since "1h ago" --level error`,
			`journalctl -xe | nixai logs analyze`,
		},
		SeeAlso: []string{"diagnose", "doctor", "monitor"},
	})

	ms.addManualEntry("package-monitor", &ManualContent{
		Title:       "package-monitor - Package Monitoring",
		Category:    "Monitoring",
		Description: "Comprehensive package monitoring and security analysis",
		Content: `The 'package-monitor' command provides package monitoring capabilities.

Features:
- Real-time package update tracking
- Security vulnerability scanning
- Usage analytics and statistics
- Dependency monitoring
- Performance impact assessment

Monitoring Categories:
- Security updates and CVE tracking
- Version update notifications
- Dependency change tracking
- Resource usage monitoring
- License compliance checking`,
		Examples: []string{
			`nixai package-monitor overview`,
			`nixai package-monitor security --scan-vulnerabilities`,
			`nixai package-monitor usage --period 30d`,
			`nixai package-monitor report --type security`,
		},
		SeeAlso: []string{"security", "updates", "packages"},
	})
}

// addManualEntry adds a manual entry to the system
func (ms *ManualSystem) addManualEntry(key string, content *ManualContent) {
	ms.entries[key] = content
}