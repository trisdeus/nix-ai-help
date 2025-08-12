package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/tui"
	"nix-ai-help/pkg/logger"
)

// AddTUICommands adds TUI-related commands to the CLI
func AddTUICommands(rootCmd *cobra.Command, logger *logger.Logger) {
	// Main TUI command
	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Start the interactive terminal interface",
		Long: `Start the nixai Terminal User Interface (TUI) for easy command discovery and execution.

The TUI provides:
• Claude Code-style interface with smart command completion
• Real-time command suggestions and help
• Multiple color themes (gruvbox, dracula, nord, tokyo-night, catppuccin)
• Command history and navigation
• Intelligent command categorization and filtering
• Plugin system integration`,
		Example: `  # Start the TUI
  nixai tui`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🚀 Starting NixAI Terminal Interface...")
			fmt.Println()
			
			// Start the TUI
			return tui.Start()
		},
	}

	rootCmd.AddCommand(tuiCmd)
}

