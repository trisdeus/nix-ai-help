package cli

import (
	"github.com/spf13/cobra"
)

// errorCmd is the main error command that contains all error management subcommands
var errorCmd = &cobra.Command{
	Use:   "error",
	Short: "Error handling and analytics management",
	Long: `Manage error handling, analytics, and debug settings for nixai.

The error management system provides:
- Analytics and pattern detection for system errors
- Error reporting and export capabilities
- Debug mode management
- Error history and trend analysis
- Smart error categorization and suggestions

Examples:
  nixai error status                    # Show error handling status
  nixai error analytics                 # View error patterns and analytics
  nixai error report                    # Generate error report
  nixai error clear                     # Clear error history
  nixai error debug on                  # Enable debug mode
  nixai error debug off                 # Disable debug mode

Use 'nixai error [command] --help' for more information about a specific command.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand provided, show help
		_ = cmd.Help()
	},
}

// NewErrorCommand creates a new error command with all subcommands
func NewErrorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   errorCmd.Use,
		Short: errorCmd.Short,
		Long:  errorCmd.Long,
		Run:   errorCmd.Run,
	}

	// Add all error management subcommands
	cmd.AddCommand(errorAnalyticsCmd)
	cmd.AddCommand(errorReportCmd)
	cmd.AddCommand(errorClearCmd)
	cmd.AddCommand(errorStatusCmd)
	cmd.AddCommand(errorDebugCmd)

	return cmd
}

// Initialize the error command system
func init() {
	// The error commands are initialized in error_commands.go
	// This init function ensures the main error command is available
}
