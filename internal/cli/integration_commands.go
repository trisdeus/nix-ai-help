package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"nix-ai-help/internal/integration"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// IntegrationCommands provides integration CLI commands
type IntegrationCommands struct {
	service *integration.Service
	logger  *logger.Logger
}

// NewIntegrationCommands creates integration commands
func NewIntegrationCommands(service *integration.Service, logger *logger.Logger) *IntegrationCommands {
	return &IntegrationCommands{
		service: service,
		logger:  logger,
	}
}

// CreateCommand creates the main integration command
func (ic *IntegrationCommands) CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai-config",
		Short: "AI-powered configuration generation and management",
		Long: `Generate, deploy, and manage NixOS configurations using AI.
This command integrates all nixai systems to provide comprehensive configuration management.`,
		Example: `  # Generate configuration with AI
  nixai ai-config generate --type desktop --description "GNOME desktop with development tools"

  # Deploy AI-generated configuration to fleet
  nixai ai-config deploy --config abc123 --targets server01,server02

  # Create collaborative editing session
  nixai ai-config collaborate --team myteam --config abc123

  # Execute plugin workflow
  nixai ai-config plugin-workflow --plugin ai-generator --workflow generate-server`,
	}

	// Add subcommands
	cmd.AddCommand(ic.createGenerateCommand())
	cmd.AddCommand(ic.createDeployCommand())
	cmd.AddCommand(ic.createCollaborateCommand())
	cmd.AddCommand(ic.createPluginWorkflowCommand())

	return cmd
}

// createGenerateCommand creates the AI configuration generation command
func (ic *IntegrationCommands) createGenerateCommand() *cobra.Command {
	var (
		configType  string
		description string
		services    []string
		packages    []string
		environment string
		output      string
		format      string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate NixOS configuration using AI",
		Long: `Use artificial intelligence to generate a complete NixOS configuration
based on your requirements. The generated configuration is automatically
versioned and can be deployed to your fleet.`,
		Example: `  # Generate a desktop configuration
  nixai ai-config generate --type desktop --description "GNOME desktop for development"

  # Generate a server configuration with specific services
  nixai ai-config generate --type server --services nginx,postgresql --packages git,docker

  # Generate configuration for specific environment
  nixai ai-config generate --type development --environment staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			request := integration.AIConfigRequest{
				Type:        configType,
				Description: description,
				Services:    services,
				Packages:    packages,
				Environment: environment,
			}

			response, err := ic.service.GenerateConfigurationWithAI(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to generate configuration: %w", err)
			}

			return ic.displayConfigGeneration(response, output, format)
		},
	}

	cmd.Flags().StringVar(&configType, "type", "", "Configuration type (desktop, server, development, minimal)")
	cmd.Flags().StringVar(&description, "description", "", "Description of desired configuration")
	cmd.Flags().StringSliceVar(&services, "services", nil, "Required services (comma-separated)")
	cmd.Flags().StringSliceVar(&packages, "packages", nil, "Required packages (comma-separated)")
	cmd.Flags().StringVar(&environment, "environment", "production", "Target environment")
	cmd.Flags().StringVar(&output, "output", "", "Save configuration to file")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("description")

	return cmd
}

// createDeployCommand creates the fleet deployment command
func (ic *IntegrationCommands) createDeployCommand() *cobra.Command {
	var (
		configHash      string
		targets         []string
		name            string
		strategy        string
		batchSize       int
		rollbackEnabled bool
		autoStart       bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy configuration to fleet machines",
		Long: `Deploy a versioned configuration from the repository to selected
fleet machines. Supports multiple deployment strategies and automatic rollback.`,
		Example: `  # Deploy to specific machines
  nixai ai-config deploy --config abc123 --targets server01,server02

  # Rolling deployment with custom batch size
  nixai ai-config deploy --config abc123 --targets server01,server02,server03 --strategy rolling --batch-size 2

  # Auto-start deployment
  nixai ai-config deploy --config abc123 --targets server01 --start`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			request := integration.FleetDeployRequest{
				Name:            name,
				ConfigHash:      configHash,
				Targets:         targets,
				CreatedBy:       "cli-user", // TODO: Get from user context
				RollbackEnabled: rollbackEnabled,
				AutoStart:       autoStart,
			}

			// Set deployment strategy
			if strategy != "" {
				// TODO: Parse strategy parameters
			}

			deployment, err := ic.service.DeployConfigurationToFleet(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to deploy configuration: %w", err)
			}

			utils.FormatHeader("Configuration Deployed")
			utils.FormatKeyValue("Deployment ID", deployment.ID)
			utils.FormatKeyValue("Configuration", deployment.ConfigHash)
			utils.FormatKeyValue("Targets", fmt.Sprintf("%d machines", len(deployment.Targets)))
			utils.FormatKeyValue("Status", string(deployment.Status))

			if autoStart {
				fmt.Println("\n✓ Deployment started automatically")
			} else {
				fmt.Printf("\nTo monitor deployment progress:\n  nixai fleet deployment status %s\n", deployment.ID)
			}

			utils.FormatDivider()
			return nil
		},
	}

	cmd.Flags().StringVar(&configHash, "config", "", "Configuration hash to deploy (required)")
	cmd.Flags().StringSliceVar(&targets, "targets", nil, "Target machine IDs (comma-separated)")
	cmd.Flags().StringVar(&name, "name", "", "Deployment name")
	cmd.Flags().StringVar(&strategy, "strategy", "rolling", "Deployment strategy (rolling, parallel)")
	cmd.Flags().IntVar(&batchSize, "batch-size", 1, "Batch size for rolling deployments")
	cmd.Flags().BoolVar(&rollbackEnabled, "rollback", true, "Enable automatic rollback")
	cmd.Flags().BoolVar(&autoStart, "start", false, "Start deployment immediately")

	cmd.MarkFlagRequired("config")
	cmd.MarkFlagRequired("targets")

	return cmd
}

// createCollaborateCommand creates the collaboration command
func (ic *IntegrationCommands) createCollaborateCommand() *cobra.Command {
	var (
		teamID     string
		configHash string
		userID     string
	)

	cmd := &cobra.Command{
		Use:   "collaborate",
		Short: "Start collaborative configuration editing",
		Long: `Create a collaborative editing session for configuration development.
Team members can work together in real-time on configuration files.`,
		Example: `  # Start collaboration session
  nixai ai-config collaborate --team myteam --config abc123 --user alice

  # Join existing session
  nixai ai-config collaborate --team myteam --config abc123 --user bob`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			request := integration.CollabSessionRequest{
				ConfigHash: configHash,
				TeamID:     teamID,
				UserID:     userID,
			}

			session, err := ic.service.CreateCollaborativeSession(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to create collaboration session: %w", err)
			}

			utils.FormatHeader("Collaboration Session Created")
			utils.FormatKeyValue("Session ID", session.ID)
			utils.FormatKeyValue("Configuration", session.ConfigHash)
			utils.FormatKeyValue("Team", session.TeamID)
			utils.FormatKeyValue("Status", session.Status)
			utils.FormatKeyValue("Participants", fmt.Sprintf("%d", len(session.Participants)))

			fmt.Printf("\nWeb interface: http://localhost:34567/collaborate/%s\n", session.ID)
			utils.FormatDivider()

			return nil
		},
	}

	cmd.Flags().StringVar(&teamID, "team", "", "Team ID (required)")
	cmd.Flags().StringVar(&configHash, "config", "", "Configuration hash (required)")
	cmd.Flags().StringVar(&userID, "user", "", "User ID (required)")

	cmd.MarkFlagRequired("team")
	cmd.MarkFlagRequired("config")
	cmd.MarkFlagRequired("user")

	return cmd
}

// createPluginWorkflowCommand creates the plugin workflow command
func (ic *IntegrationCommands) createPluginWorkflowCommand() *cobra.Command {
	var (
		pluginID     string
		workflowName string
		parameters   []string
	)

	cmd := &cobra.Command{
		Use:   "plugin-workflow",
		Short: "Execute plugin-based workflow",
		Long: `Execute advanced workflows using plugins. Plugins can provide
specialized functionality for configuration generation, validation, and deployment.`,
		Example: `  # Execute AI generation workflow
  nixai ai-config plugin-workflow --plugin ai-generator --workflow generate-config

  # Execute validation workflow with parameters
  nixai ai-config plugin-workflow --plugin validator --workflow validate-security --params config=abc123,level=strict`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Parse parameters
			params := make(map[string]interface{})
			for _, param := range parameters {
				parts := strings.SplitN(param, "=", 2)
				if len(parts) == 2 {
					params[parts[0]] = parts[1]
				}
			}

			request := integration.PluginWorkflowRequest{
				PluginID:     pluginID,
				WorkflowName: workflowName,
				Parameters:   params,
			}

			response, err := ic.service.ExecutePluginWorkflow(ctx, request)
			if err != nil {
				return fmt.Errorf("failed to execute plugin workflow: %w", err)
			}

			utils.FormatHeader("Plugin Workflow Executed")
			utils.FormatKeyValue("Plugin", response.PluginID)
			utils.FormatKeyValue("Workflow", response.WorkflowName)
			utils.FormatKeyValue("Executed At", response.ExecutedAt.Format("2006-01-02 15:04:05"))

			fmt.Println("\nResult:")
			if result, ok := response.Result.(string); ok {
				fmt.Println(result)
			} else {
				json.NewEncoder(os.Stdout).Encode(response.Result)
			}

			utils.FormatDivider()
			return nil
		},
	}

	cmd.Flags().StringVar(&pluginID, "plugin", "", "Plugin ID (required)")
	cmd.Flags().StringVar(&workflowName, "workflow", "", "Workflow name (required)")
	cmd.Flags().StringSliceVar(&parameters, "params", nil, "Parameters (key=value format)")

	cmd.MarkFlagRequired("plugin")
	cmd.MarkFlagRequired("workflow")

	return cmd
}

// displayConfigGeneration displays configuration generation results
func (ic *IntegrationCommands) displayConfigGeneration(response *integration.AIConfigResponse, output, format string) error {
	if format == "json" {
		return json.NewEncoder(os.Stdout).Encode(response)
	}

	utils.FormatHeader("AI Configuration Generated")
	utils.FormatKeyValue("Branch", response.Branch)
	utils.FormatKeyValue("Commit Hash", response.CommitHash)

	if len(response.Suggestions) > 0 {
		fmt.Printf("\n💡 Suggestions:\n")
		for _, suggestion := range response.Suggestions {
			fmt.Printf("  • %s\n", suggestion)
		}
	}

	if len(response.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range response.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	// Save to file if requested
	if output != "" {
		err := os.WriteFile(output, []byte(response.Configuration), 0644)
		if err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		utils.FormatKeyValue("Saved to", output)
	}

	fmt.Printf("\nConfiguration:\n")
	fmt.Println("```nix")
	fmt.Println(response.Configuration)
	fmt.Println("```")

	fmt.Printf("\nTo deploy this configuration:\n")
	fmt.Printf("  nixai ai-config deploy --config %s --targets <machine-ids>\n", response.CommitHash)

	utils.FormatDivider()
	return nil
}
