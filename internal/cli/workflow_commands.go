package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"nix-ai-help/internal/workflow"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// Workflow command with subcommands
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Automated workflow management and execution",
	Long: `Manage and execute automated workflows for NixOS configuration and management tasks.

The workflow system allows you to define, manage, and execute automated sequences of actions
for common NixOS operations like system updates, package installations, service management,
and configuration deployments.

Available subcommands:
  list      - List all registered workflows
  show      - Display workflow details
  execute   - Execute a workflow
  status    - Check workflow execution status
  create    - Create a new workflow from template
  validate  - Validate workflow definition

Examples:
  nixai workflow list                           # List all workflows
  nixai workflow show system-update            # Show workflow details
  nixai workflow execute system-update         # Execute workflow`,
	Run: func(cmd *cobra.Command, args []string) {
		showWorkflowHelp()
	},
}

// Workflow list command
var workflowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered workflows",
	Long:  `List all workflows that are currently registered in the workflow engine.`,
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkflowList(cmd, args)
	},
}

// Workflow show command
var workflowShowCmd = &cobra.Command{
	Use:   "show <workflow-id>",
	Short: "Display detailed workflow information",
	Long:  `Display detailed information about a specific workflow.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkflowShow(cmd, args)
	},
}

// Workflow execute command
var workflowExecuteCmd = &cobra.Command{
	Use:   "execute <workflow-id>",
	Short: "Execute a workflow",
	Long:  `Execute a workflow either synchronously or asynchronously.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkflowExecute(cmd, args)
	},
}

// Workflow create command
var workflowCreateCmd = &cobra.Command{
	Use:   "create [workflow-name]",
	Short: "Create a new workflow interactively",
	Long:  `Create a new workflow using interactive prompts or from a template.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkflowCreate(cmd, args)
	},
}

// Workflow validate command
var workflowValidateCmd = &cobra.Command{
	Use:   "validate <workflow-file>",
	Short: "Validate a workflow definition file",
	Long:  `Validate a workflow definition YAML file for syntax and structure.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleWorkflowValidate(cmd, args)
	},
}

// CreateWorkflowCommand creates the workflow command with all subcommands
func CreateWorkflowCommand() *cobra.Command {
	// Add subcommands
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowShowCmd)
	workflowCmd.AddCommand(workflowExecuteCmd)
	workflowCmd.AddCommand(workflowCreateCmd)
	workflowCmd.AddCommand(workflowValidateCmd)

	return workflowCmd
}

// Helper functions

func showWorkflowHelp() {
	fmt.Println(utils.FormatHeader("🔄 Automated Workflow Management"))
	fmt.Println()
	fmt.Println("Manage and execute automated workflows for NixOS configuration and operations.")
	fmt.Println()
	fmt.Println(utils.FormatSubheader("Available Commands"))
	fmt.Println()
	fmt.Println("  " + utils.FormatKeyValue("list", "List all registered workflows"))
	fmt.Println("  " + utils.FormatKeyValue("show", "Display workflow details"))
	fmt.Println("  " + utils.FormatKeyValue("execute", "Execute a workflow"))
	fmt.Println()
	fmt.Println(utils.FormatSubheader("Quick Examples"))
	fmt.Println()
	fmt.Println("  nixai workflow list")
	fmt.Println("  nixai workflow execute system-update")
	fmt.Println()
	fmt.Println(utils.FormatTip("Use 'nixai workflow <command> --help' for detailed command information"))
}

// Command handlers - enhanced implementations
func handleWorkflowList(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("📋 Workflow List"))
	fmt.Println()

	// Initialize workflow storage
	config := workflow.WorkflowStorageConfig{
		WorkflowDirs: []string{
			"./workflows/templates",
			filepath.Join(os.Getenv("HOME"), ".config/nixai/workflows"),
			"/etc/nixai/workflows",
		},
		AutoReload: true,
	}

	// Create a simple logger adapter
	log := &simpleLogger{}
	storage := workflow.NewWorkflowStorage(config, log)

	// Load workflows
	if err := storage.LoadWorkflows(); err != nil {
		fmt.Println(utils.FormatError("Failed to load workflows: " + err.Error()))
		return
	}

	// Get workflow metadata
	workflows := storage.GetWorkflowMetadata()

	if len(workflows) == 0 {
		fmt.Println(utils.FormatInfo("No workflows found"))
		fmt.Println()
		fmt.Println("You can create workflows using:")
		fmt.Println("  nixai workflow create")
		fmt.Println("  nixai workflow create --template system-update")
		return
	}

	fmt.Printf("Found %d workflow(s):\n\n", len(workflows))

	for _, wf := range workflows {
		fmt.Printf("• %s\n", utils.FormatKeyValue(wf.ID, wf.Name))
		if wf.Description != "" {
			fmt.Printf("  %s\n", wf.Description)
		}
		fmt.Printf("  Tasks: %d", wf.TaskCount)
		if wf.TriggerCount > 0 {
			fmt.Printf(" | Triggers: %d", wf.TriggerCount)
		}
		if len(wf.Tags) > 0 {
			fmt.Printf(" | Tags: %v", wf.Tags)
		}
		fmt.Println()
		fmt.Println()
	}
}

// Simple logger adapter for workflow system
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, args ...interface{}) {
	// Only log in debug mode
}

func (l *simpleLogger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("[INFO] "+msg+"\n", args...)
	} else {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

func (l *simpleLogger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("[WARN] "+msg+"\n", args...)
	} else {
		fmt.Printf("[WARN] %s\n", msg)
	}
}

func (l *simpleLogger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("[ERROR] "+msg+"\n", args...)
	} else {
		fmt.Printf("[ERROR] %s\n", msg)
	}
}

func handleWorkflowShow(cmd *cobra.Command, args []string) {
	workflowID := args[0]
	fmt.Println(utils.FormatHeader("🔍 Workflow Details: " + workflowID))
	fmt.Println()

	// Initialize workflow storage
	config := workflow.WorkflowStorageConfig{
		WorkflowDirs: []string{
			"./workflows/templates",
			filepath.Join(os.Getenv("HOME"), ".config/nixai/workflows"),
			"/etc/nixai/workflows",
		},
		AutoReload: true,
	}

	log := &simpleLogger{}
	storage := workflow.NewWorkflowStorage(config, log)

	// Load workflows
	if err := storage.LoadWorkflows(); err != nil {
		fmt.Println(utils.FormatError("Failed to load workflows: " + err.Error()))
		return
	}

	// Get workflow definition
	workflowDef, err := storage.GetWorkflow(workflowID)
	if err != nil {
		fmt.Println(utils.FormatError("Workflow not found: " + workflowID))
		fmt.Println()
		fmt.Println("Available workflows:")
		for _, wf := range storage.GetWorkflowMetadata() {
			fmt.Printf("  • %s\n", wf.ID)
		}
		return
	}

	// Display workflow information
	fmt.Println(utils.FormatKeyValue("Name", workflowDef.Name))
	fmt.Println(utils.FormatKeyValue("Description", workflowDef.Description))
	fmt.Println(utils.FormatKeyValue("Version", workflowDef.Version))
	if workflowDef.Author != "" {
		fmt.Println(utils.FormatKeyValue("Author", workflowDef.Author))
	}
	if len(workflowDef.Tags) > 0 {
		fmt.Println(utils.FormatKeyValue("Tags", fmt.Sprintf("%v", workflowDef.Tags)))
	}

	// Configuration
	fmt.Println()
	fmt.Println(utils.FormatSubheader("⚙️  Configuration"))
	if workflowDef.Config.MaxConcurrentTasks > 0 {
		fmt.Println(utils.FormatKeyValue("Max Concurrent Tasks", fmt.Sprintf("%d", workflowDef.Config.MaxConcurrentTasks)))
	}
	if workflowDef.Config.Timeout > 0 {
		fmt.Println(utils.FormatKeyValue("Timeout", workflowDef.Config.Timeout.String()))
	}
	if workflowDef.Config.RetryCount > 0 {
		fmt.Println(utils.FormatKeyValue("Retry Count", fmt.Sprintf("%d", workflowDef.Config.RetryCount)))
	}
	if workflowDef.Config.OnFailure != "" {
		fmt.Println(utils.FormatKeyValue("On Failure", workflowDef.Config.OnFailure))
	}

	// Tasks
	fmt.Println()
	fmt.Println(utils.FormatSubheader("📋 Tasks"))
	for i, task := range workflowDef.Tasks {
		fmt.Printf("%d. %s\n", i+1, utils.FormatKeyValue(task.ID, task.Name))
		if task.Description != "" {
			fmt.Printf("   %s\n", task.Description)
		}
		fmt.Printf("   Type: %s", task.Type)
		if len(task.DependsOn) > 0 {
			fmt.Printf(" | Depends on: %v", task.DependsOn)
		}
		if task.Action.Type != "" {
			fmt.Printf(" | Action: %s", task.Action.Type)
		}
		fmt.Println()
	}

	// Triggers
	if len(workflowDef.Triggers) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubheader("🔔 Triggers"))
		for i, trigger := range workflowDef.Triggers {
			fmt.Printf("%d. Type: %s\n", i+1, trigger.Type)
			if trigger.Condition != "" {
				fmt.Printf("   Condition: %s\n", trigger.Condition)
			}
			for key, value := range trigger.Parameters {
				fmt.Printf("   %s: %v\n", key, value)
			}
		}
	}

	// Variables
	if len(workflowDef.Variables) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubheader("🔧 Variables"))
		for key, value := range workflowDef.Variables {
			fmt.Println(utils.FormatKeyValue(key, fmt.Sprintf("%v", value)))
		}
	}
}

func handleWorkflowExecute(cmd *cobra.Command, args []string) {
	workflowID := args[0]
	fmt.Println(utils.FormatHeader("▶️  Executing Workflow: " + workflowID))
	fmt.Println()

	// Initialize workflow storage and engine
	config := workflow.WorkflowStorageConfig{
		WorkflowDirs: []string{
			"./workflows/templates",
			filepath.Join(os.Getenv("HOME"), ".config/nixai/workflows"),
			"/etc/nixai/workflows",
		},
		AutoReload: true,
	}

	log := &simpleLogger{}
	storage := workflow.NewWorkflowStorage(config, log)

	// Load workflows
	if err := storage.LoadWorkflows(); err != nil {
		fmt.Println(utils.FormatError("Failed to load workflows: " + err.Error()))
		return
	}

	// Get workflow definition
	workflowDef, err := storage.GetWorkflow(workflowID)
	if err != nil {
		fmt.Println(utils.FormatError("Workflow not found: " + workflowID))
		return
	}

	// Convert to workflow and execute
	parser := workflow.NewWorkflowParser(log)
	wf, err := parser.ConvertToWorkflow(workflowDef)
	if err != nil {
		fmt.Println(utils.FormatError("Failed to convert workflow: " + err.Error()))
		return
	}

	fmt.Println("🔄 Starting workflow execution...")
	fmt.Printf("📋 Workflow: %s (%d tasks)\n", wf.Name, len(wf.Tasks))
	fmt.Println("✅ Workflow validation passed")
	fmt.Println()

	// Show what would be executed
	fmt.Println(utils.FormatSubheader("📋 Execution Plan"))
	for i, task := range wf.Tasks {
		fmt.Printf("%d. %s\n", i+1, task.Name)
		if len(task.DependsOn) > 0 {
			fmt.Printf("   Dependencies: %v\n", task.DependsOn)
		}
		if len(task.Actions) > 0 {
			for _, action := range task.Actions {
				fmt.Printf("   Action: %s", action.Type)
				if action.Command != "" {
					fmt.Printf(" (%s)", action.Command)
				}
				fmt.Println()
			}
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatInfo("Workflow execution will be fully implemented in Phase 2.3 Week 3"))
	fmt.Println(utils.FormatTip("For now, this shows the execution plan. Full execution coming soon!"))
}

func handleWorkflowCreate(cmd *cobra.Command, args []string) {
	fmt.Println(utils.FormatHeader("🔨 Create New Workflow"))
	fmt.Println()

	log := &simpleLogger{}
	builder := workflow.NewWorkflowBuilder(log)

	workflowDef, err := builder.BuildWorkflowInteractive()
	if err != nil {
		fmt.Println(utils.FormatError("Failed to create workflow: " + err.Error()))
		return
	}

	// Save the workflow
	config := workflow.WorkflowStorageConfig{
		WorkflowDirs: []string{
			filepath.Join(os.Getenv("HOME"), ".config/nixai/workflows"),
		},
		AutoReload: true,
	}

	storage := workflow.NewWorkflowStorage(config, log)
	workflowID := workflow.GenerateWorkflowID(workflowDef.Name)

	if err := storage.SaveWorkflow(workflowID, workflowDef); err != nil {
		fmt.Println(utils.FormatError("Failed to save workflow: " + err.Error()))
		return
	}

	fmt.Println()
	fmt.Println(utils.FormatSuccess("✅ Workflow created successfully!"))
	fmt.Printf("ID: %s\n", workflowID)
	fmt.Printf("Name: %s\n", workflowDef.Name)
	fmt.Println()
	fmt.Println("You can now:")
	fmt.Printf("  nixai workflow show %s\n", workflowID)
	fmt.Printf("  nixai workflow execute %s\n", workflowID)
}

func handleWorkflowValidate(cmd *cobra.Command, args []string) {
	filename := args[0]
	fmt.Println(utils.FormatHeader("✅ Validate Workflow: " + filename))
	fmt.Println()

	log := &simpleLogger{}
	parser := workflow.NewWorkflowParser(log)

	// Parse the workflow file
	workflowDef, err := parser.ParseWorkflowFile(filename)
	if err != nil {
		fmt.Println(utils.FormatError("Validation failed: " + err.Error()))
		return
	}

	// Try to convert it to internal format
	_, err = parser.ConvertToWorkflow(workflowDef)
	if err != nil {
		fmt.Println(utils.FormatError("Conversion failed: " + err.Error()))
		return
	}

	fmt.Println(utils.FormatSuccess("✅ Workflow validation passed!"))
	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Name", workflowDef.Name))
	fmt.Println(utils.FormatKeyValue("Tasks", fmt.Sprintf("%d", len(workflowDef.Tasks))))
	fmt.Println(utils.FormatKeyValue("Triggers", fmt.Sprintf("%d", len(workflowDef.Triggers))))
}
