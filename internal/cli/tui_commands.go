package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/config"
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
• Default: Execution-aware interface with AI-powered command detection
• Classic: Original menu-based interface
• Simple: Basic Claude Code-style interface

The execution-aware interface provides:
• AI-powered natural language command detection
• Real-time execution status and monitoring
• Security validation and confirmation prompts
• Split-panel view with execution history
• Command box with intelligent completion
• Interactive execution management`,
		Example: `  # Start the execution-aware TUI (default)
  nixai tui

  # Start the classic menu-based TUI
  nixai tui --classic

  # Start simple TUI without execution features
  nixai tui --simple`,
		RunE: func(cmd *cobra.Command, args []string) error {
			classic, _ := cmd.Flags().GetBool("classic")
			simple, _ := cmd.Flags().GetBool("simple")
			
			if classic {
				fmt.Println("🚀 Starting NixAI Classic Terminal Interface...")
				fmt.Println()
				
				// Create and start the classic TUI
				tuiInstance := tui.NewTUI()
				return tuiInstance.Start()
			} else if simple {
				fmt.Println("🚀 Starting NixAI Terminal Interface (Simple Mode)...")
				fmt.Println()
				
				// Start the simple Claude Code-style TUI
				return tui.StartClaudeTUI()
			} else {
				fmt.Println("🚀 Starting NixAI Execution-Aware Terminal Interface...")
				fmt.Println()
				
				// Load configuration
				cfg, err := config.LoadUserConfig()
				if err != nil {
					logger.Warn(fmt.Sprintf("Failed to load configuration, using defaults: %v", err))
					cfg = config.DefaultUserConfig()
				}
				
				// Start the execution-aware TUI
				return StartExecutionAwareTUI(cfg, logger)
			}
		},
	}

	// Add flags
	tuiCmd.Flags().Bool("classic", false, "Use the classic menu-based TUI interface")
	tuiCmd.Flags().Bool("simple", false, "Use the simple TUI interface without execution features")

	rootCmd.AddCommand(tuiCmd)
}

// StartExecutionAwareTUI starts the execution-aware TUI
func StartExecutionAwareTUI(cfg *config.UserConfig, logger *logger.Logger) error {
	tui, err := tui.NewExecutionAwareTUI(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create execution-aware TUI: %w", err)
	}
	defer tui.Close()
	
	return tui.Start()
}