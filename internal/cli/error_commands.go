package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/errors"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// Error management command definitions

// errorAnalyticsCmd shows error analytics and patterns
var errorAnalyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "View error analytics and patterns",
	Long: `Display comprehensive error analytics including patterns, frequency, and recommendations.

Shows:
- Detected error patterns and their frequency
- Most common error types and codes
- AI-generated recommendations for error resolution
- Error trends and analytics data

Examples:
  nixai error analytics`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := handleErrorAnalytics(cfg, args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// errorReportCmd generates and exports error reports
var errorReportCmd = &cobra.Command{
	Use:   "report [filename]",
	Short: "Generate and optionally export error report",
	Long: `Generate a comprehensive error report and optionally export it to a file.

The report includes:
- Recent error history
- Error resolution status
- Diagnostic information
- Context and metadata

Examples:
  nixai error report                    # Display report
  nixai error report errors.json       # Export to file`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := handleErrorReport(cfg, args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// errorClearCmd clears error history and analytics
var errorClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear error history and analytics data",
	Long: `Clear all stored error history and analytics data.

This action:
- Removes all error tracking data
- Clears analytics patterns and metrics
- Resets error counters and history
- Cannot be undone

Examples:
  nixai error clear`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := handleErrorClear(cfg, args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// errorStatusCmd shows error handling status
var errorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current error handling status",
	Long: `Display the current status of error handling and analytics systems.

Shows:
- Error tracking system status
- Debug mode configuration
- Recent error statistics
- Analytics data location

Examples:
  nixai error status`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := handleErrorStatus(cfg, args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// errorDebugCmd manages debug mode for error tracking
var errorDebugCmd = &cobra.Command{
	Use:   "debug [on|off]",
	Short: "Enable/disable debug mode for error tracking",
	Long: `Enable or disable debug mode for enhanced error tracking and diagnostics.

Debug mode provides:
- Detailed error stack traces
- Enhanced logging and context
- Additional diagnostic information
- More comprehensive error analytics

Examples:
  nixai error debug                     # Show current status
  nixai error debug on                 # Enable debug mode
  nixai error debug off                # Disable debug mode`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		if err := handleErrorDebug(cfg, args); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// handleErrorCommand handles error management commands
func handleErrorCommand(args []string) error {
	if len(args) == 0 {
		showErrorHelp()
		return nil
	}

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	subcommand := args[0]
	switch subcommand {
	case "analytics":
		return handleErrorAnalytics(cfg, args[1:])
	case "report":
		return handleErrorReport(cfg, args[1:])
	case "clear":
		return handleErrorClear(cfg, args[1:])
	case "status":
		return handleErrorStatus(cfg, args[1:])
	case "debug":
		return handleErrorDebug(cfg, args[1:])
	default:
		return fmt.Errorf("unknown error subcommand: %s", subcommand)
	}
}

// showErrorHelp displays help for error management commands
func showErrorHelp() {
	fmt.Println(utils.FormatHeader("🚨 Error Management Commands"))
	fmt.Println()
	fmt.Println(utils.FormatSubsection("Available Commands", ""))
	fmt.Println("  analytics          - View error analytics and patterns")
	fmt.Println("  report [file]      - Generate and optionally export error report")
	fmt.Println("  clear              - Clear error history and analytics data")
	fmt.Println("  status             - Show current error handling status")
	fmt.Println("  debug [on|off]     - Enable/disable debug mode for error tracking")
	fmt.Println()
	fmt.Println(utils.FormatTip("Error management helps track and resolve recurring issues"))
}

// handleErrorAnalytics shows error analytics and patterns
func handleErrorAnalytics(cfg *config.UserConfig, args []string) error {
	fmt.Println(utils.FormatHeader("📊 Error Analytics"))
	fmt.Println()

	// Initialize error manager
	errorManagerConfig := &errors.ErrorManagerConfig{
		DebugMode:           cfg.LogLevel == "debug" || cfg.LogLevel == "trace",
		GracefulDegradation: true,
		AnalyticsEnabled:    true,
		AnalyticsDataDir:    filepath.Join(os.Getenv("HOME"), ".config", "nixai", "error_analytics"),
		RetryConfig:         errors.DefaultRetryConfig(),
		MaxLastErrors:       50,
	}

	errorManager := errors.NewErrorManager(errorManagerConfig)

	// Get analytics report
	report := errorManager.GetAnalyticsReport()
	if report == nil {
		fmt.Println(utils.FormatWarning("No analytics data available"))
		return nil
	}

	// Display error patterns
	fmt.Println(utils.FormatSubsection("Error Patterns", ""))
	if len(report.DetectedPatterns) == 0 {
		fmt.Println("  No error patterns detected")
	} else {
		for i, pattern := range report.DetectedPatterns {
			if i >= 10 { // Limit to top 10
				break
			}
			fmt.Printf("  %d. %s (Count: %d)\n", i+1, pattern.Pattern, pattern.Count)
		}
	}

	fmt.Println()

	// Display error frequency
	fmt.Println(utils.FormatSubsection("Error Frequency", ""))
	if len(report.TopErrors) == 0 {
		fmt.Println("  No error frequency data available")
	} else {
		for i, freq := range report.TopErrors {
			if i >= 10 { // Limit to top 10
				break
			}
			fmt.Printf("  %s: %d occurrences\n", freq.Code, freq.Count)
		}
	}

	fmt.Println()

	// Display recommendations
	fmt.Println(utils.FormatSubsection("Recommendations", ""))
	if len(report.RecommendedActions) == 0 {
		fmt.Println("  No recommendations available")
	} else {
		for i, rec := range report.RecommendedActions {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}

	return nil
}

// handleErrorReport generates and optionally exports an error report
func handleErrorReport(cfg *config.UserConfig, args []string) error {
	fmt.Println(utils.FormatHeader("📋 Error Report"))
	fmt.Println()

	// Initialize error manager
	errorManagerConfig := &errors.ErrorManagerConfig{
		DebugMode:           cfg.LogLevel == "debug" || cfg.LogLevel == "trace",
		GracefulDegradation: true,
		AnalyticsEnabled:    true,
		AnalyticsDataDir:    filepath.Join(os.Getenv("HOME"), ".config", "nixai", "error_analytics"),
		RetryConfig:         errors.DefaultRetryConfig(),
		MaxLastErrors:       50,
	}

	errorManager := errors.NewErrorManager(errorManagerConfig)

	// Get recent errors
	recentErrors := errorManager.GetLastErrors(20)

	fmt.Println(utils.FormatSubsection("Recent Errors", ""))
	if len(recentErrors) == 0 {
		fmt.Println("  No recent errors found")
	} else {
		for i, lastErr := range recentErrors {
			status := "❌"
			if lastErr.Resolved {
				status = "✅"
			}

			fmt.Printf("  %d. %s [%s] %s\n", i+1, status, lastErr.Error.Code, lastErr.Error.Message)
			fmt.Printf("     Time: %s\n", lastErr.Timestamp.Format("2006-01-02 15:04:05"))
			if lastErr.Context != "" {
				fmt.Printf("     Context: %s\n", lastErr.Context)
			}
			fmt.Println()
		}
	}

	// Export to file if specified
	if len(args) > 0 {
		filename := args[0]
		if !strings.HasSuffix(filename, ".json") {
			filename += ".json"
		}

		fmt.Print(utils.FormatInfo(fmt.Sprintf("Exporting report to %s... ", filename)))

		if err := errorManager.ExportAnalytics(filename); err != nil {
			fmt.Println(utils.FormatError("failed"))
			return fmt.Errorf("failed to export report: %v", err)
		}

		fmt.Println(utils.FormatSuccess("done"))
		fmt.Println(utils.FormatKeyValue("Report saved to", filename))
	}

	return nil
}

// handleErrorClear clears error history and analytics data
func handleErrorClear(cfg *config.UserConfig, args []string) error {
	fmt.Println(utils.FormatHeader("🧹 Clear Error Data"))
	fmt.Println()

	analyticsDir := filepath.Join(os.Getenv("HOME"), ".config", "nixai", "error_analytics")

	fmt.Print(utils.FormatInfo("Clearing error analytics data... "))

	// Remove analytics directory
	if err := os.RemoveAll(analyticsDir); err != nil {
		fmt.Println(utils.FormatError("failed"))
		return fmt.Errorf("failed to clear analytics data: %v", err)
	}

	// Recreate directory
	if err := os.MkdirAll(analyticsDir, 0755); err != nil {
		fmt.Println(utils.FormatWarning("partial"))
		fmt.Printf("Warning: Failed to recreate analytics directory: %v\n", err)
	} else {
		fmt.Println(utils.FormatSuccess("done"))
	}

	fmt.Println(utils.FormatKeyValue("Status", "Error analytics data cleared"))
	fmt.Println(utils.FormatTip("New error tracking will start fresh"))

	return nil
}

// handleErrorStatus shows current error handling status
func handleErrorStatus(cfg *config.UserConfig, args []string) error {
	fmt.Println(utils.FormatHeader("📈 Error Handling Status"))
	fmt.Println()

	// Initialize error manager
	errorManagerConfig := &errors.ErrorManagerConfig{
		DebugMode:           cfg.LogLevel == "debug" || cfg.LogLevel == "trace",
		GracefulDegradation: true,
		AnalyticsEnabled:    true,
		AnalyticsDataDir:    filepath.Join(os.Getenv("HOME"), ".config", "nixai", "error_analytics"),
		RetryConfig:         errors.DefaultRetryConfig(),
		MaxLastErrors:       50,
	}

	errorManager := errors.NewErrorManager(errorManagerConfig)

	// Check if analytics directory exists
	analyticsDir := filepath.Join(os.Getenv("HOME"), ".config", "nixai", "error_analytics")
	if _, err := os.Stat(analyticsDir); os.IsNotExist(err) {
		fmt.Println(utils.FormatKeyValue("Analytics Status", "❌ Not initialized"))
	} else {
		fmt.Println(utils.FormatKeyValue("Analytics Status", "✅ Active"))
	}

	// Debug mode status
	debugMode := cfg.LogLevel == "debug" || cfg.LogLevel == "trace"
	if debugMode {
		fmt.Println(utils.FormatKeyValue("Debug Mode", "✅ Enabled"))
	} else {
		fmt.Println(utils.FormatKeyValue("Debug Mode", "❌ Disabled"))
	}

	// Recent error count
	recentErrors := errorManager.GetLastErrors(50)
	fmt.Println(utils.FormatKeyValue("Recent Errors", fmt.Sprintf("%d", len(recentErrors))))

	// Resolved vs unresolved
	resolved := 0
	for _, err := range recentErrors {
		if err.Resolved {
			resolved++
		}
	}
	unresolved := len(recentErrors) - resolved

	fmt.Println(utils.FormatKeyValue("Resolved Errors", fmt.Sprintf("%d", resolved)))
	fmt.Println(utils.FormatKeyValue("Unresolved Errors", fmt.Sprintf("%d", unresolved)))

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Analytics Directory", analyticsDir))

	return nil
}

// handleErrorDebug enables/disables debug mode
func handleErrorDebug(cfg *config.UserConfig, args []string) error {
	fmt.Println(utils.FormatHeader("🐛 Error Debug Mode"))
	fmt.Println()

	if len(args) == 0 {
		// Show current status
		debugMode := cfg.LogLevel == "debug" || cfg.LogLevel == "trace"
		if debugMode {
			fmt.Println(utils.FormatKeyValue("Debug Mode", "✅ Enabled"))
		} else {
			fmt.Println(utils.FormatKeyValue("Debug Mode", "❌ Disabled"))
		}
		fmt.Println()
		fmt.Println(utils.FormatTip("Use 'nixai error debug on' or 'nixai error debug off' to change"))
		return nil
	}

	action := strings.ToLower(args[0])
	switch action {
	case "on", "enable", "true":
		// This would require updating the config file
		fmt.Println(utils.FormatWarning("Debug mode configuration requires manual config file editing"))
		fmt.Println(utils.FormatTip("Set 'log_level: debug' in your config file"))

	case "off", "disable", "false":
		fmt.Println(utils.FormatWarning("Debug mode configuration requires manual config file editing"))
		fmt.Println(utils.FormatTip("Set 'log_level: info' in your config file"))

	default:
		return fmt.Errorf("invalid debug action: %s (use 'on' or 'off')", action)
	}

	return nil
}
