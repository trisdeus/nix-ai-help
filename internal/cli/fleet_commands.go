package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"nix-ai-help/internal/fleet"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// FleetCommands provides fleet management CLI commands
type FleetCommands struct {
	fleetManager *fleet.FleetManager
	logger       *logger.Logger
}

// NewFleetCommands creates fleet management commands
func NewFleetCommands(fleetManager *fleet.FleetManager, logger *logger.Logger) *FleetCommands {
	return &FleetCommands{
		fleetManager: fleetManager,
		logger:       logger,
	}
}

// CreateCommand creates the main fleet command
func (fc *FleetCommands) CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fleet",
		Short: "Manage NixOS fleet deployments and monitoring",
		Long: `The fleet command provides comprehensive fleet management capabilities for NixOS machines.
It supports multi-machine deployments, health monitoring, and centralized configuration management.`,
		Example: `  # List all machines in the fleet
  nixai fleet list

  # Add a new machine to the fleet
  nixai fleet add-machine --id server01 --name "Production Server 1" --address 192.168.1.10

  # Deploy configuration to fleet
  nixai fleet deploy --config abc123 --targets server01,server02

  # Monitor fleet health
  nixai fleet health

  # Create a rolling deployment
  nixai fleet deploy --config abc123 --strategy rolling --batch-size 2`,
	}

	// Add subcommands
	cmd.AddCommand(fc.createListCommand())
	cmd.AddCommand(fc.createAddMachineCommand())
	cmd.AddCommand(fc.createRemoveMachineCommand())
	cmd.AddCommand(fc.createHealthCommand())
	cmd.AddCommand(fc.createDeployCommand())
	cmd.AddCommand(fc.createDeploymentCommand())
	cmd.AddCommand(fc.createMonitorCommand())

	return cmd
}

// createListCommand creates the list machines command
func (fc *FleetCommands) createListCommand() *cobra.Command {
	var (
		environment string
		tags        []string
		format      string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all machines in the fleet",
		Long:  "Display a list of all machines registered in the fleet with their current status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var machines []*fleet.Machine
			var err error

			// Filter by environment if specified
			if environment != "" {
				machines, err = fc.fleetManager.GetMachinesByEnvironment(ctx, environment)
			} else if len(tags) > 0 {
				machines, err = fc.fleetManager.GetMachinesByTag(ctx, tags)
			} else {
				machines, err = fc.fleetManager.ListMachines(ctx)
			}

			if err != nil {
				return fmt.Errorf("failed to list machines: %w", err)
			}

			return fc.displayMachines(machines, format)
		},
	}

	cmd.Flags().StringVar(&environment, "environment", "", "Filter by environment (production, staging, development)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Filter by tags (comma-separated)")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// createAddMachineCommand creates the add machine command
func (fc *FleetCommands) createAddMachineCommand() *cobra.Command {
	var machine fleet.Machine

	cmd := &cobra.Command{
		Use:   "add-machine",
		Short: "Add a new machine to the fleet",
		Long:  "Register a new NixOS machine in the fleet for management and deployment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Validate required fields
			if machine.ID == "" {
				return fmt.Errorf("machine ID is required")
			}
			if machine.Name == "" {
				return fmt.Errorf("machine name is required")
			}
			if machine.Address == "" {
				return fmt.Errorf("machine address is required")
			}

			err := fc.fleetManager.AddMachine(ctx, &machine)
			if err != nil {
				return fmt.Errorf("failed to add machine: %w", err)
			}

			utils.FormatHeader("Machine Added Successfully")
			utils.FormatKeyValue("ID", machine.ID)
			utils.FormatKeyValue("Name", machine.Name)
			utils.FormatKeyValue("Address", machine.Address)
			utils.FormatKeyValue("Environment", machine.Environment)
			utils.FormatDivider()

			return nil
		},
	}

	cmd.Flags().StringVar(&machine.ID, "id", "", "Unique machine ID (required)")
	cmd.Flags().StringVar(&machine.Name, "name", "", "Human-readable machine name (required)")
	cmd.Flags().StringVar(&machine.Address, "address", "", "Machine IP address or hostname (required)")
	cmd.Flags().StringVar(&machine.SSHConfig.User, "ssh-user", "root", "SSH username")
	cmd.Flags().IntVar(&machine.SSHConfig.Port, "ssh-port", 22, "SSH port")
	cmd.Flags().StringVar(&machine.SSHConfig.KeyPath, "ssh-key", "", "Path to SSH private key")
	cmd.Flags().StringVar(&machine.Environment, "environment", "production", "Environment (production, staging, development)")
	cmd.Flags().StringSliceVar(&machine.Tags, "tags", nil, "Machine tags (comma-separated)")

	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("address")

	return cmd
}

// createRemoveMachineCommand creates the remove machine command
func (fc *FleetCommands) createRemoveMachineCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-machine <machine-id>",
		Short: "Remove a machine from the fleet",
		Long:  "Unregister a machine from fleet management. This will prevent future deployments to the machine.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			machineID := args[0]

			err := fc.fleetManager.RemoveMachine(ctx, machineID)
			if err != nil {
				return fmt.Errorf("failed to remove machine: %w", err)
			}

			utils.FormatHeader("Machine Removed Successfully")
			utils.FormatKeyValue("ID", machineID)
			utils.FormatDivider()

			return nil
		},
	}

	return cmd
}

// createHealthCommand creates the fleet health command
func (fc *FleetCommands) createHealthCommand() *cobra.Command {
	var (
		machineID string
		format    string
	)

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Display fleet health status",
		Long:  "Show comprehensive health information for the entire fleet or a specific machine.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if machineID != "" {
				return fc.displayMachineHealth(ctx, machineID, format)
			}

			return fc.displayFleetHealth(ctx, format)
		},
	}

	cmd.Flags().StringVar(&machineID, "machine", "", "Show health for specific machine")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// createDeployCommand creates the deploy command
func (fc *FleetCommands) createDeployCommand() *cobra.Command {
	var (
		configHash       string
		targets          []string
		name             string
		strategy         string
		batchSize        int
		batchDelay       int
		failureThreshold float64
		autoStart        bool
		rollbackEnabled  bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy configuration to fleet machines",
		Long:  "Create and optionally start a deployment of NixOS configuration to selected fleet machines.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if configHash == "" {
				return fmt.Errorf("configuration hash is required")
			}
			if len(targets) == 0 {
				return fmt.Errorf("at least one target machine is required")
			}
			if name == "" {
				name = fmt.Sprintf("deploy-%d", time.Now().Unix())
			}

			// Create deployment request
			req := fleet.DeploymentRequest{
				Name:            name,
				ConfigHash:      configHash,
				Targets:         targets,
				CreatedBy:       "cli-user", // TODO: Get from user context
				RollbackEnabled: rollbackEnabled,
				Strategy: fleet.DeploymentStrategy{
					Type:             strategy,
					BatchSize:        batchSize,
					BatchDelay:       batchDelay,
					FailureThreshold: failureThreshold,
				},
			}

			deployment, err := fc.fleetManager.CreateDeployment(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create deployment: %w", err)
			}

			utils.FormatHeader("Deployment Created")
			utils.FormatKeyValue("ID", deployment.ID)
			utils.FormatKeyValue("Name", deployment.Name)
			utils.FormatKeyValue("Configuration", deployment.ConfigHash)
			utils.FormatKeyValue("Targets", fmt.Sprintf("%d machines", len(deployment.Targets)))
			utils.FormatKeyValue("Strategy", deployment.Strategy.Type)

			if autoStart {
				err = fc.fleetManager.StartDeployment(ctx, deployment.ID)
				if err != nil {
					return fmt.Errorf("failed to start deployment: %w", err)
				}
				fmt.Println("\n✓ Deployment started")
			} else {
				fmt.Printf("\nTo start the deployment, run:\n  nixai fleet deployment start %s\n", deployment.ID)
			}

			utils.FormatDivider()
			return nil
		},
	}

	cmd.Flags().StringVar(&configHash, "config", "", "Configuration hash to deploy (required)")
	cmd.Flags().StringSliceVar(&targets, "targets", nil, "Target machine IDs (comma-separated, required)")
	cmd.Flags().StringVar(&name, "name", "", "Deployment name (auto-generated if not specified)")
	cmd.Flags().StringVar(&strategy, "strategy", "rolling", "Deployment strategy (rolling, parallel, blue_green, canary)")
	cmd.Flags().IntVar(&batchSize, "batch-size", 1, "Number of machines to deploy to simultaneously")
	cmd.Flags().IntVar(&batchDelay, "batch-delay", 10, "Delay between batches in seconds")
	cmd.Flags().Float64Var(&failureThreshold, "failure-threshold", 0.1, "Failure threshold (0.0-1.0)")
	cmd.Flags().BoolVar(&autoStart, "start", false, "Start deployment immediately")
	cmd.Flags().BoolVar(&rollbackEnabled, "rollback", true, "Enable automatic rollback on failure")

	cmd.MarkFlagRequired("config")
	cmd.MarkFlagRequired("targets")

	return cmd
}

// createDeploymentCommand creates the deployment management command
func (fc *FleetCommands) createDeploymentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: "Manage deployments",
		Long:  "View and manage fleet deployments including starting, stopping, and monitoring progress.",
	}

	cmd.AddCommand(fc.createDeploymentListCommand())
	cmd.AddCommand(fc.createDeploymentStatusCommand())
	cmd.AddCommand(fc.createDeploymentStartCommand())
	cmd.AddCommand(fc.createDeploymentCancelCommand())

	return cmd
}

// createDeploymentListCommand creates the deployment list command
func (fc *FleetCommands) createDeploymentListCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all deployments",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			deployments, err := fc.fleetManager.ListDeployments(ctx)
			if err != nil {
				return fmt.Errorf("failed to list deployments: %w", err)
			}

			return fc.displayDeployments(deployments, format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// createDeploymentStatusCommand creates the deployment status command
func (fc *FleetCommands) createDeploymentStatusCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "status <deployment-id>",
		Short: "Show deployment status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			deploymentID := args[0]

			deployment, err := fc.fleetManager.GetDeployment(ctx, deploymentID)
			if err != nil {
				return fmt.Errorf("failed to get deployment: %w", err)
			}

			return fc.displayDeploymentStatus(deployment, format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// createDeploymentStartCommand creates the deployment start command
func (fc *FleetCommands) createDeploymentStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start <deployment-id>",
		Short: "Start a deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			deploymentID := args[0]

			err := fc.fleetManager.StartDeployment(ctx, deploymentID)
			if err != nil {
				return fmt.Errorf("failed to start deployment: %w", err)
			}

			fmt.Printf("✓ Deployment %s started\n", deploymentID)
			return nil
		},
	}

	return cmd
}

// createDeploymentCancelCommand creates the deployment cancel command
func (fc *FleetCommands) createDeploymentCancelCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel <deployment-id>",
		Short: "Cancel a running deployment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			deploymentID := args[0]

			err := fc.fleetManager.CancelDeployment(ctx, deploymentID)
			if err != nil {
				return fmt.Errorf("failed to cancel deployment: %w", err)
			}

			fmt.Printf("✓ Deployment %s cancelled\n", deploymentID)
			return nil
		},
	}

	return cmd
}

// createMonitorCommand creates the fleet monitoring command
func (fc *FleetCommands) createMonitorCommand() *cobra.Command {
	var interval time.Duration

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start fleet monitoring",
		Long:  "Start continuous monitoring of fleet health and status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			monitor := fleet.NewMonitor(fc.fleetManager)

			fmt.Printf("Starting fleet monitoring (interval: %s)\n", interval)
			fmt.Println("Press Ctrl+C to stop monitoring")

			err := monitor.Start(ctx, interval)
			if err != nil {
				return fmt.Errorf("failed to start monitoring: %w", err)
			}

			// Wait for interrupt
			<-ctx.Done()

			return monitor.Stop()
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 30*time.Second, "Monitoring interval")

	return cmd
}

// Display functions

func (fc *FleetCommands) displayMachines(machines []*fleet.Machine, format string) error {
	if format == "json" {
		return json.NewEncoder(os.Stdout).Encode(machines)
	}

	utils.FormatHeader("Fleet Machines")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tADDRESS\tSTATUS\tENVIRONMENT\tHEALTH\tLAST SEEN")
	fmt.Fprintln(w, "─\t─\t─\t─\t─\t─\t─")

	for _, machine := range machines {
		lastSeen := "Never"
		if !machine.LastSeen.IsZero() {
			lastSeen = machine.LastSeen.Format("2006-01-02 15:04")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			machine.ID,
			machine.Name,
			machine.Address,
			machine.Status,
			machine.Environment,
			machine.Health.Overall,
			lastSeen,
		)
	}

	w.Flush()
	utils.FormatDivider()
	fmt.Printf("Total machines: %d\n", len(machines))

	return nil
}

func (fc *FleetCommands) displayFleetHealth(ctx context.Context, format string) error {
	monitor := fleet.NewMonitor(fc.fleetManager)
	health, err := monitor.GetFleetHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to get fleet health: %w", err)
	}

	if format == "json" {
		return json.NewEncoder(os.Stdout).Encode(health)
	}

	utils.FormatHeader("Fleet Health Summary")
	utils.FormatKeyValue("Overall Status", health.OverallStatus)
	utils.FormatKeyValue("Health Percentage", fmt.Sprintf("%.1f%%", health.HealthPercentage))
	utils.FormatKeyValue("Total Machines", strconv.Itoa(health.TotalMachines))
	utils.FormatKeyValue("Online", strconv.Itoa(health.OnlineMachines))
	utils.FormatKeyValue("Offline", strconv.Itoa(health.OfflineMachines))
	utils.FormatKeyValue("Degraded", strconv.Itoa(health.DegradedMachines))
	utils.FormatKeyValue("Maintenance", strconv.Itoa(health.MaintenanceMachines))

	if len(health.Alerts) > 0 {
		fmt.Printf("\n🚨 Active Alerts (%d):\n", len(health.Alerts))
		for _, alert := range health.Alerts {
			fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(alert.Level), alert.Source, alert.Message)
		}
	}

	utils.FormatDivider()
	return nil
}

func (fc *FleetCommands) displayMachineHealth(ctx context.Context, machineID, format string) error {
	machine, err := fc.fleetManager.GetMachine(ctx, machineID)
	if err != nil {
		return fmt.Errorf("failed to get machine: %w", err)
	}

	if format == "json" {
		return json.NewEncoder(os.Stdout).Encode(machine.Health)
	}

	utils.FormatHeader(fmt.Sprintf("Machine Health: %s", machine.Name))
	utils.FormatKeyValue("Overall Status", machine.Health.Overall)
	utils.FormatKeyValue("CPU Usage", fmt.Sprintf("%.1f%% (%s)", machine.Health.CPU.Usage, machine.Health.CPU.Status))
	utils.FormatKeyValue("Memory Usage", fmt.Sprintf("%.1f%% (%s)", machine.Health.Memory.Usage, machine.Health.Memory.Status))
	utils.FormatKeyValue("Disk Usage", fmt.Sprintf("%.1f%% (%s)", machine.Health.Disk.Usage, machine.Health.Disk.Status))
	utils.FormatKeyValue("Network Status", machine.Health.Network.Status)
	utils.FormatKeyValue("Last Check", machine.Health.LastCheck.Format("2006-01-02 15:04:05"))

	if len(machine.Health.Services) > 0 {
		fmt.Println("\nServices:")
		for _, service := range machine.Health.Services {
			fmt.Printf("  %s: %s (since %s)\n", service.Name, service.Status, service.Since.Format("2006-01-02 15:04"))
		}
	}

	utils.FormatDivider()
	return nil
}

func (fc *FleetCommands) displayDeployments(deployments []*fleet.Deployment, format string) error {
	if format == "json" {
		return json.NewEncoder(os.Stdout).Encode(deployments)
	}

	utils.FormatHeader("Fleet Deployments")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCONFIG\tSTATUS\tPROGRESS\tCREATED\tCREATED BY")
	fmt.Fprintln(w, "─\t─\t─\t─\t─\t─\t─")

	for _, deployment := range deployments {
		progress := fmt.Sprintf("%d/%d (%d%%)",
			deployment.Progress.Completed,
			deployment.Progress.Total,
			deployment.Progress.Percentage)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			deployment.ID[:8]+"...",
			deployment.Name,
			deployment.ConfigHash[:8]+"...",
			deployment.Status,
			progress,
			deployment.CreatedAt.Format("2006-01-02 15:04"),
			deployment.CreatedBy,
		)
	}

	w.Flush()
	utils.FormatDivider()
	fmt.Printf("Total deployments: %d\n", len(deployments))

	return nil
}

func (fc *FleetCommands) displayDeploymentStatus(deployment *fleet.Deployment, format string) error {
	if format == "json" {
		return json.NewEncoder(os.Stdout).Encode(deployment)
	}

	utils.FormatHeader(fmt.Sprintf("Deployment Status: %s", deployment.Name))
	utils.FormatKeyValue("ID", deployment.ID)
	utils.FormatKeyValue("Configuration", deployment.ConfigHash)
	utils.FormatKeyValue("Status", string(deployment.Status))
	utils.FormatKeyValue("Progress", fmt.Sprintf("%d/%d (%d%%)",
		deployment.Progress.Completed,
		deployment.Progress.Total,
		deployment.Progress.Percentage))
	utils.FormatKeyValue("Strategy", deployment.Strategy.Type)
	utils.FormatKeyValue("Created", deployment.CreatedAt.Format("2006-01-02 15:04:05"))
	utils.FormatKeyValue("Created By", deployment.CreatedBy)

	if deployment.StartedAt != nil {
		utils.FormatKeyValue("Started", deployment.StartedAt.Format("2006-01-02 15:04:05"))
	}

	if deployment.CompletedAt != nil {
		utils.FormatKeyValue("Completed", deployment.CompletedAt.Format("2006-01-02 15:04:05"))
	}

	if len(deployment.Results) > 0 {
		fmt.Println("\nMachine Results:")
		for machineID, result := range deployment.Results {
			status := result.Status
			if result.Error != "" {
				status += " - " + result.Error
			}
			fmt.Printf("  %s: %s\n", machineID, status)
		}
	}

	utils.FormatDivider()
	return nil
}
