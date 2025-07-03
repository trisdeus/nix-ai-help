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

Available TUI modes:
• Default: Claude Code-style interface with command box at bottom
• Classic: Original menu-based interface

The Claude Code-style interface provides:
• Command box at the bottom with completion
• Command execution output above
• Command history navigation with ↑/↓
• Tab completion for commands
• Real-time suggestions`,
		Example: `  # Start the Claude Code-style TUI (default)
  nixai tui

  # Start the classic menu-based TUI
  nixai tui --classic`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classic, _ := cmd.Flags().GetBool("classic")
			
			if classic {
				fmt.Println("🚀 Starting NixAI Classic Terminal Interface...")
				fmt.Println()
				
				// Create and start the classic TUI
				tuiInstance := tui.NewTUI()
				return tuiInstance.Start()
			} else {
				fmt.Println("🚀 Starting NixAI Terminal Interface (Claude Code Style)...")
				fmt.Println()
				
				// Start the Claude Code-style TUI
				return tui.StartClaudeTUI()
			}
		},
	}

	// Add flags
	tuiCmd.Flags().Bool("classic", false, "Use the classic menu-based TUI interface")

	rootCmd.AddCommand(tuiCmd)
}