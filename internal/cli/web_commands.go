package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/internal/web"
	"nix-ai-help/pkg/logger"
)

// AddWebCommands adds web-related commands to the CLI
func AddWebCommands(rootCmd *cobra.Command, logger *logger.Logger) {
	// Web interface command group
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Web interface management",
		Long:  "Manage the nixai web interface with dashboard, configuration builder, fleet management, and team collaboration",
	}

	// Web start subcommand
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the web interface server",
		Long: `Start the enhanced web interface for nixai with modern dashboard and collaboration features.

The web interface provides:
• Modern responsive dashboard with light/dark theme
• Visual configuration builder with drag-and-drop
• Real-time team collaboration
• Git-like version control for configurations
• Fleet management and deployment
• AI-powered configuration generation
• System monitoring and health dashboards`,
		Example: `  # Start web interface on default port
  nixai web start

  # Start with custom port
  nixai web start --port 8080

  # Start with specific repository
  nixai web start --repo /path/to/nixos/config

  # Start with both custom port and repository
  nixai web start --port 8080 --repo /etc/nixos`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get flags
			port, _ := cmd.Flags().GetInt("port")
			repoPath, _ := cmd.Flags().GetString("repo")

			// Resolve repository path
			if repoPath == "" {
				repoPath = "."
			}
			absRepoPath, err := filepath.Abs(repoPath)
			if err != nil {
				return fmt.Errorf("failed to resolve repository path: %w", err)
			}

			// Initialize components
			repo, err := repository.NewConfigRepository(absRepoPath, logger)
			if err != nil {
				logger.Warn(fmt.Sprintf("Could not initialize repository at %s: %v", absRepoPath, err))
				logger.Info("Web interface will start with limited functionality")
			}

			teamManager := team.NewTeamManager(logger)

			// Print startup information
			fmt.Printf("🌐 Starting NixAI Web Interface\n")
			fmt.Printf("📂 Repository: %s\n", absRepoPath)
			fmt.Printf("🌍 Server: http://localhost:%d\n", port)
			fmt.Printf("📊 Dashboard: http://localhost:%d/dashboard\n", port)
			fmt.Printf("🎨 Builder: http://localhost:%d/builder\n", port)
			fmt.Printf("🚀 Fleet: http://localhost:%d/fleet\n", port)
			fmt.Printf("👥 Teams: http://localhost:%d/teams\n", port)
			fmt.Printf("📝 Versions: http://localhost:%d/versions\n", port)
			fmt.Printf("\n🎯 Features:\n")
			fmt.Printf("  • Modern responsive UI with light/dark theme\n")
			fmt.Printf("  • Visual configuration builder\n")
			fmt.Printf("  • Real-time collaboration\n")
			fmt.Printf("  • Version control & fleet management\n")
			fmt.Printf("  • AI-powered assistance\n")
			fmt.Printf("\n💡 Tip: Use Ctrl+C to stop the server\n\n")

			// Start the enhanced web server
			return startEnhancedWebServer(port, repo, teamManager, logger)
		},
	}

	// Add flags to start command
	startCmd.Flags().IntP("port", "p", 34567, "Port to run the web interface on (default: 34567)")
	startCmd.Flags().StringP("repo", "r", "", "Path to the NixOS configuration repository (default: current directory)")

	// Add subcommands
	webCmd.AddCommand(startCmd)
	rootCmd.AddCommand(webCmd)
}

// startEnhancedWebServer starts the enhanced web server with all features
func startEnhancedWebServer(port int, repo *repository.ConfigRepository, teamManager *team.TeamManager, logger *logger.Logger) error {
	// Create and start the enhanced web server
	server := web.NewEnhancedServer(repo, teamManager, logger)
	return server.Start(port)
}