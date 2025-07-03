package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/ai/function/execution"
	"nix-ai-help/internal/ai/functionbase"
	"nix-ai-help/pkg/logger"
)

// ExecuteOptions holds options for the execute command
type ExecuteOptions struct {
	Command      string
	Args         []string
	Description  string
	Category     string
	RequiresSudo bool
	WorkingDir   string
	Environment  map[string]string
	DryRun       bool
	Timeout      string
	Interactive  bool
	Force        bool
}

// NewExecuteCommand creates the execute command
func NewExecuteCommand() *cobra.Command {
	opts := &ExecuteOptions{
		Environment: make(map[string]string),
	}

	cmd := &cobra.Command{
		Use:   "execute [command] [args...]",
		Short: "Execute system commands safely with AI-powered validation",
		Long: `Execute system commands with comprehensive security validation, permission management,
audit logging, and optional sudo elevation. This command provides AI-controlled
command execution with safety features.

Examples:
  # Install a package
  nixai execute nix-env -iA nixpkgs.firefox --description "Install Firefox browser" --category package

  # Rebuild NixOS configuration with sudo
  nixai execute nixos-rebuild switch --description "Apply NixOS changes" --category system --sudo

  # Check system status (dry run)
  nixai execute systemctl status sshd --description "Check SSH status" --category utility --dry-run

  # Execute with custom environment
  nixai execute nix build --description "Build project" --category development --env "NIX_PATH=/custom/path"

Categories:
  - package: Package management (nix-env, nix, etc.)
  - system: System management (nixos-rebuild, systemctl, etc.)
  - configuration: Configuration editing and management
  - development: Development tools and builds
  - utility: General utility commands (ls, grep, etc.)`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Command = args[0]
			if len(args) > 1 {
				opts.Args = args[1:]
			}

			return runExecuteCommand(cmd.Context(), opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.Description, "description", "d", "", "Description of what the command does (required)")
	cmd.Flags().StringVarP(&opts.Category, "category", "c", "", "Command category: package, system, configuration, development, utility (required)")
	cmd.Flags().BoolVar(&opts.RequiresSudo, "sudo", false, "Execute with sudo privileges")
	cmd.Flags().StringVarP(&opts.WorkingDir, "working-dir", "w", "", "Working directory for command execution")
	cmd.Flags().StringToStringVarP(&opts.Environment, "env", "e", nil, "Environment variables (key=value)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show what would be executed without running")
	cmd.Flags().StringVarP(&opts.Timeout, "timeout", "t", "", "Execution timeout (e.g., '5m', '30s')")
	cmd.Flags().BoolVarP(&opts.Interactive, "interactive", "i", true, "Enable interactive mode with confirmations")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Skip confirmation prompts (use with caution)")

	// Mark required flags
	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("category")

	return cmd
}

// runExecuteCommand executes the command with the given options
func runExecuteCommand(ctx context.Context, opts *ExecuteOptions) error {
	log := logger.NewLogger()

	// Validate category
	validCategories := []string{"package", "system", "configuration", "development", "utility"}
	categoryValid := false
	for _, cat := range validCategories {
		if opts.Category == cat {
			categoryValid = true
			break
		}
	}
	if !categoryValid {
		return fmt.Errorf("invalid category: %s. Valid categories: %s", opts.Category, strings.Join(validCategories, ", "))
	}

	// Create execution function and initializer
	execFunc := execution.NewExecutionFunction()

	// Create security components and initializer
	permissionManager, auditLogger, sudoManager, err := execution.CreateSecurityComponents(log)
	if err != nil {
		return fmt.Errorf("failed to create security components: %w", err)
	}

	// Initialize execution function
	execConfig := execution.GetDefaultConfig()
	if err := execFunc.Initialize(permissionManager, auditLogger, sudoManager, execConfig); err != nil {
		return fmt.Errorf("failed to initialize execution function: %w", err)
	}

	// Set interactive mode
	execFunc.SetInteractive(opts.Interactive)

	// Build parameters for the function call
	params := map[string]interface{}{
		"command":     opts.Command,
		"args":        opts.Args,
		"description": opts.Description,
		"category":    opts.Category,
		"dryRun":      opts.DryRun,
	}

	if opts.RequiresSudo {
		params["requiresSudo"] = true
	}

	if opts.WorkingDir != "" {
		params["workingDir"] = opts.WorkingDir
	}

	if len(opts.Environment) > 0 {
		params["environment"] = opts.Environment
	}

	if opts.Timeout != "" {
		params["timeout"] = opts.Timeout
	}

	// Set up progress callback
	var progressCallback functionbase.ProgressCallback
	if opts.Interactive {
		progressCallback = func(progress functionbase.Progress) {
			fmt.Printf("\r[%s] %.0f%% - %s", progress.Stage, progress.Percentage, progress.Message)
			if progress.Percentage >= 100 {
				fmt.Println()
			}
		}
	}

	// Set up function options
	functionOptions := &functionbase.FunctionOptions{
		Timeout:          5 * time.Minute,
		ProgressCallback: progressCallback,
		Async:            false,
	}

	// Show confirmation if not in force mode and not dry run
	if !opts.Force && !opts.DryRun && opts.Interactive {
		fmt.Printf("Execute command: %s %s\n", opts.Command, strings.Join(opts.Args, " "))
		fmt.Printf("Description: %s\n", opts.Description)
		fmt.Printf("Category: %s\n", opts.Category)
		if opts.RequiresSudo {
			fmt.Printf("Sudo: required\n")
		}
		if opts.WorkingDir != "" {
			fmt.Printf("Working directory: %s\n", opts.WorkingDir)
		}
		if len(opts.Environment) > 0 {
			fmt.Printf("Environment variables: %v\n", opts.Environment)
		}
		fmt.Print("\nContinue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Execution cancelled")
			return nil
		}
	}

	fmt.Printf("Executing: %s %s\n", opts.Command, strings.Join(opts.Args, " "))

	// Execute the function
	result, err := execFunc.Execute(ctx, params, functionOptions)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Handle the result
	if !result.Success {
		fmt.Printf("❌ Execution failed: %s\n", result.Error)
		return fmt.Errorf("command execution failed")
	}

	// Extract execution response
	if execResponse, ok := result.Data.(*execution.ExecutionResponse); ok {
		fmt.Printf("✅ Execution completed successfully\n")
		fmt.Printf("Duration: %s\n", execResponse.Duration)
		fmt.Printf("Exit code: %d\n", execResponse.ExitCode)

		if execResponse.Output != "" {
			fmt.Printf("\nOutput:\n%s\n", execResponse.Output)
		}

		if execResponse.DryRun {
			fmt.Printf("📝 This was a dry run - no actual execution occurred\n")
		}

		if execResponse.Metadata != nil {
			if executionTime, ok := execResponse.Metadata["executionTime"].(float64); ok {
				fmt.Printf("Execution time: %.2fms\n", executionTime)
			}
		}
	}

	log.Info("Command execution completed successfully")
	return nil
}

// NewExecuteStatusCommand creates a command to show execution status and history
func NewExecuteStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show execution status and recent command history",
		Long:  `Display information about recent command executions, active sessions, and audit logs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExecuteStatusCommand(cmd.Context())
		},
	}

	return cmd
}

// runExecuteStatusCommand shows execution status
func runExecuteStatusCommand(ctx context.Context) error {
	log := logger.NewLogger()

	// Create security components
	_, auditLogger, sudoManager, err := execution.CreateSecurityComponents(log)
	if err != nil {
		return fmt.Errorf("failed to create security components: %w", err)
	}

	fmt.Println("📊 Execution Status")
	fmt.Println("==================")

	// Show active sudo sessions
	sessions := sudoManager.GetSessionInfo()
	if len(sessions) > 0 {
		fmt.Printf("\n🔐 Active Sudo Sessions: %d\n", len(sessions))
		for _, session := range sessions {
			fmt.Printf("  Session %s:\n", session.SessionID)
			fmt.Printf("    Started: %s\n", session.StartTime.Format(time.RFC3339))
			fmt.Printf("    Last used: %s\n", session.LastUsed.Format(time.RFC3339))
			fmt.Printf("    Commands: %d\n", session.CommandCount)
			fmt.Printf("    Valid: %t\n", session.Valid)
		}
	} else {
		fmt.Println("\n🔐 No active sudo sessions")
	}

	// Show audit statistics
	since := time.Now().Add(-24 * time.Hour) // Last 24 hours
	stats, err := auditLogger.GetAuditStatistics(since)
	if err == nil && len(stats) > 0 {
		fmt.Printf("\n📈 Audit Statistics (last 24h):\n")
		for key, value := range stats {
			fmt.Printf("  %s: %d\n", key, value)
		}
	}

	return nil
}

// NewExecuteConfigCommand creates a command to manage execution configuration
func NewExecuteConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage execution configuration settings",
		Long:  `View and modify configuration settings for command execution, including security policies and allowed commands.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := execution.GetDefaultConfig()
			fmt.Println("🔧 Execution Configuration")
			fmt.Println("=========================")
			fmt.Printf("Enabled: %t\n", config.Enabled)
			fmt.Printf("Dry run default: %t\n", config.DryRunDefault)
			fmt.Printf("Confirmation required: %t\n", config.ConfirmationRequired)
			fmt.Printf("Max execution time: %s\n", time.Duration(config.MaxExecutionTime))
			fmt.Printf("Allowed commands: %d\n", len(config.AllowedCommands))
			fmt.Printf("Forbidden commands: %d\n", len(config.ForbiddenCommands))
			fmt.Printf("Sudo commands: %d\n", len(config.SudoCommands))
			fmt.Printf("Allowed directories: %d\n", len(config.AllowedDirectories))
			fmt.Printf("Forbidden paths: %d\n", len(config.ForbiddenPaths))
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current execution configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := execution.GetDefaultConfig()
			fmt.Println("🔧 Execution Configuration")
			fmt.Println("=========================")
			fmt.Printf("Enabled: %t\n", config.Enabled)
			fmt.Printf("Dry run default: %t\n", config.DryRunDefault)
			fmt.Printf("Confirmation required: %t\n", config.ConfirmationRequired)
			fmt.Printf("Max execution time: %s\n", time.Duration(config.MaxExecutionTime))
			fmt.Printf("Allowed commands: %d\n", len(config.AllowedCommands))
			fmt.Printf("Forbidden commands: %d\n", len(config.ForbiddenCommands))
			fmt.Printf("Sudo commands: %d\n", len(config.SudoCommands))
			fmt.Printf("Allowed directories: %d\n", len(config.AllowedDirectories))
			fmt.Printf("Forbidden paths: %d\n", len(config.ForbiddenPaths))
			return nil
		},
	})

	return cmd
}