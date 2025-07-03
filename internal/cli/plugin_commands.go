package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/plugins"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// PluginCLIManager manages plugin-related CLI commands
type PluginCLIManager struct {
	manager   plugins.PluginManager
	registry  plugins.PluginRegistry
	loader    plugins.PluginLoader
	discovery *plugins.PluginDiscovery
	config    *config.UserConfig
	logger    *logger.Logger
}

// NewPluginCLIManager creates a new plugin CLI manager
func NewPluginCLIManager(cfg *config.UserConfig, log *logger.Logger) *PluginCLIManager {
	manager := plugins.NewManager(cfg, log)
	registry := plugins.NewRegistry(log)
	loader := plugins.NewLoader(log)
	discovery := plugins.NewPluginDiscovery(log)

	return &PluginCLIManager{
		manager:   manager,
		registry:  registry,
		loader:    loader,
		discovery: discovery,
		config:    cfg,
		logger:    log,
	}
}

// CreatePluginCommands creates all plugin-related commands
func (pcm *PluginCLIManager) CreatePluginCommands() *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage nixai plugins",
		Long: `Manage plugins for nixai. Plugins extend nixai functionality with custom operations and capabilities.

Examples:
  nixai plugin list                    # List all installed plugins
  nixai plugin search web              # Search for plugins
  nixai plugin install my-plugin      # Install a plugin
  nixai plugin enable my-plugin       # Enable a plugin
  nixai plugin disable my-plugin      # Disable a plugin
  nixai plugin status my-plugin       # Show plugin status
  nixai plugin execute my-plugin op   # Execute plugin operation`,
	}

	// Add subcommands
	pluginCmd.AddCommand(pcm.createListCommand())
	pluginCmd.AddCommand(pcm.createSearchCommand())
	pluginCmd.AddCommand(pcm.createInstallCommand())
	pluginCmd.AddCommand(pcm.createUninstallCommand())
	pluginCmd.AddCommand(pcm.createEnableCommand())
	pluginCmd.AddCommand(pcm.createDisableCommand())
	pluginCmd.AddCommand(pcm.createStatusCommand())
	pluginCmd.AddCommand(pcm.createInfoCommand())
	pluginCmd.AddCommand(pcm.createExecuteCommand())
	pluginCmd.AddCommand(pcm.createDiscoverCommand())
	pluginCmd.AddCommand(pcm.createValidateCommand())
	pluginCmd.AddCommand(pcm.createMetricsCommand())
	pluginCmd.AddCommand(pcm.createEventsCommand())
	pluginCmd.AddCommand(pcm.createCreateCommand())
	pluginCmd.AddCommand(pcm.createIntegratedCommand())

	return pluginCmd
}

// createListCommand creates the list command
func (pcm *PluginCLIManager) createListCommand() *cobra.Command {
	var showAll bool
	var showCapabilities bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.listPlugins(cmd.OutOrStdout(), showAll, showCapabilities)
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show all plugins including disabled ones")
	cmd.Flags().BoolVar(&showCapabilities, "capabilities", false, "Show plugin capabilities")

	return cmd
}

// createSearchCommand creates the search command
func (pcm *PluginCLIManager) createSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for plugins",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			return pcm.searchPlugins(cmd.OutOrStdout(), query)
		},
	}

	return cmd
}

// createInstallCommand creates the install command
func (pcm *PluginCLIManager) createInstallCommand() *cobra.Command {
	var fromPath string
	var enable bool

	cmd := &cobra.Command{
		Use:   "install [plugin-name-or-path]",
		Short: "Install a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginPath := args[0]
			if fromPath != "" {
				pluginPath = fromPath
			}
			return pcm.installPlugin(cmd.OutOrStdout(), pluginPath, enable)
		},
	}

	cmd.Flags().StringVar(&fromPath, "from", "", "Install from specific path")
	cmd.Flags().BoolVar(&enable, "enable", true, "Enable plugin after installation")

	return cmd
}

// createUninstallCommand creates the uninstall command
func (pcm *PluginCLIManager) createUninstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [plugin-name]",
		Short: "Uninstall a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.uninstallPlugin(cmd.OutOrStdout(), args[0])
		},
	}

	return cmd
}

// createEnableCommand creates the enable command
func (pcm *PluginCLIManager) createEnableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable [plugin-name]",
		Short: "Enable a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.enablePlugin(cmd.OutOrStdout(), args[0])
		},
	}

	return cmd
}

// createDisableCommand creates the disable command
func (pcm *PluginCLIManager) createDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable [plugin-name]",
		Short: "Disable a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.disablePlugin(cmd.OutOrStdout(), args[0])
		},
	}

	return cmd
}

// createStatusCommand creates the status command
func (pcm *PluginCLIManager) createStatusCommand() *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:   "status [plugin-name]",
		Short: "Show plugin status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := ""
			if len(args) > 0 {
				pluginName = args[0]
			}
			return pcm.showStatus(cmd.OutOrStdout(), pluginName, detailed)
		},
	}

	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed status information")

	return cmd
}

// createInfoCommand creates the info command
func (pcm *PluginCLIManager) createInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [plugin-name]",
		Short: "Show detailed plugin information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.showInfo(cmd.OutOrStdout(), args[0])
		},
	}

	return cmd
}

// createExecuteCommand creates the execute command
func (pcm *PluginCLIManager) createExecuteCommand() *cobra.Command {
	var params string

	cmd := &cobra.Command{
		Use:   "execute [plugin-name] [operation]",
		Short: "Execute a plugin operation",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			operation := args[1]

			var paramMap map[string]interface{}
			if params != "" {
				if err := json.Unmarshal([]byte(params), &paramMap); err != nil {
					return fmt.Errorf("invalid parameters JSON: %w", err)
				}
			}

			return pcm.executeOperation(cmd.OutOrStdout(), pluginName, operation, paramMap)
		},
	}

	cmd.Flags().StringVar(&params, "params", "", "Operation parameters as JSON")

	return cmd
}

// createDiscoverCommand creates the discover command
func (pcm *PluginCLIManager) createDiscoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover plugins in standard directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.discoverPlugins(cmd.OutOrStdout())
		},
	}

	return cmd
}

// createValidateCommand creates the validate command
func (pcm *PluginCLIManager) createValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [plugin-path]",
		Short: "Validate a plugin file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.validatePlugin(cmd.OutOrStdout(), args[0])
		},
	}

	return cmd
}

// createMetricsCommand creates the metrics command
func (pcm *PluginCLIManager) createMetricsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics [plugin-name]",
		Short: "Show plugin metrics",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := ""
			if len(args) > 0 {
				pluginName = args[0]
			}
			return pcm.showMetrics(cmd.OutOrStdout(), pluginName)
		},
	}

	return cmd
}

// createEventsCommand creates the events command
func (pcm *PluginCLIManager) createEventsCommand() *cobra.Command {
	var limit int
	var eventType string
	var source string

	cmd := &cobra.Command{
		Use:   "events",
		Short: "Show plugin events",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.showEvents(cmd.OutOrStdout(), limit, eventType, source)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Number of events to show")
	cmd.Flags().StringVar(&eventType, "type", "", "Filter by event type")
	cmd.Flags().StringVar(&source, "source", "", "Filter by event source")

	return cmd
}

// createCreateCommand creates the create command
func (pcm *PluginCLIManager) createCreateCommand() *cobra.Command {
	var template string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "create [plugin-name]",
		Short: "Create a new plugin from template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pcm.createPlugin(cmd.OutOrStdout(), args[0], template, outputDir)
		},
	}

	cmd.Flags().StringVar(&template, "template", "basic", "Plugin template to use")
	cmd.Flags().StringVar(&outputDir, "output", ".", "Output directory")

	return cmd
}

// Implementation methods

func (pcm *PluginCLIManager) listPlugins(out io.Writer, showAll, showCapabilities bool) error {
	fmt.Fprintln(out, utils.FormatHeader("📦 Installed Plugins"))
	fmt.Fprintln(out)

	plugins := pcm.manager.ListPlugins()
	if len(plugins) == 0 {
		fmt.Fprintln(out, utils.FormatInfo("No plugins installed"))
		return nil
	}

	for _, plugin := range plugins {
		status, _ := pcm.manager.GetPluginStatus(plugin.Name())

		statusIcon := "❓"
		statusColor := utils.FormatInfo
		if status != nil {
			switch status.State {
			case 3: // StateRunning
				statusIcon = "✅"
				statusColor = utils.FormatSuccess
			case 5: // StateStopped
				statusIcon = "⏹️"
				statusColor = utils.FormatWarning
			case 6: // StateError
				statusIcon = "❌"
				statusColor = utils.FormatError
			case 7: // StateDisabled
				statusIcon = "🚫"
				statusColor = utils.FormatInfo
			}
		}

		fmt.Fprintln(out, statusColor(fmt.Sprintf("%s %s (%s) - %s",
			statusIcon, plugin.Name(), plugin.Version(), plugin.Description())))

		if showCapabilities {
			capabilities := plugin.Capabilities()
			if len(capabilities) > 0 {
				fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("   Capabilities: %s", strings.Join(capabilities, ", "))))
			}
		}
	}

	return nil
}

func (pcm *PluginCLIManager) searchPlugins(out io.Writer, query string) error {
	fmt.Fprintln(out, utils.FormatHeader("🔍 Plugin Search Results"))
	fmt.Fprintln(out)

	if query != "" {
		fmt.Fprintln(out, utils.FormatKeyValue("Query", query))
		fmt.Fprintln(out)
	}

	plugins := pcm.registry.Search(query)
	if len(plugins) == 0 {
		if query == "" {
			fmt.Fprintln(out, utils.FormatInfo("No plugins available"))
		} else {
			fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("No plugins found matching '%s'", query)))
		}
		return nil
	}

	for _, plugin := range plugins {
		fmt.Fprintln(out, utils.FormatSuccess(fmt.Sprintf("📦 %s (%s)", plugin.Name(), plugin.Version())))
		fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("   %s", plugin.Description())))

		capabilities := plugin.Capabilities()
		if len(capabilities) > 0 {
			fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("   Capabilities: %s", strings.Join(capabilities, ", "))))
		}
		fmt.Fprintln(out)
	}

	return nil
}

func (pcm *PluginCLIManager) installPlugin(out io.Writer, pluginPath string, enable bool) error {
	fmt.Fprintln(out, utils.FormatHeader("📦 Installing Plugin"))
	fmt.Fprintln(out)

	// Check if path exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}

	// Extract plugin name from path for configuration
	pluginName := strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))

	config := plugins.PluginConfig{
		Name:          pluginName,
		Enabled:       enable,
		Version:       "unknown",
		Configuration: make(map[string]interface{}),
		Environment:   make(map[string]string),
		Resources: plugins.ResourceLimits{
			MaxMemoryMB:      100,
			MaxCPUPercent:    50,
			MaxExecutionTime: 30 * time.Second,
			NetworkAccess:    true,
		},
		SecurityPolicy: plugins.SecurityPolicy{
			AllowFileSystem:  true,
			AllowNetwork:     true,
			AllowSystemCalls: false,
			SandboxLevel:     plugins.SandboxBasic,
		},
	}

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin Path", pluginPath))
	fmt.Fprintln(out, utils.FormatKeyValue("Plugin Name", pluginName))
	fmt.Fprintln(out)

	if err := pcm.manager.LoadPlugin(pluginPath, config); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	if enable {
		if err := pcm.manager.StartPlugin(pluginName); err != nil {
			pcm.logger.Warn(fmt.Sprintf("Plugin installed but failed to start: %v", err))
			fmt.Fprintln(out, utils.FormatWarning("Plugin installed but failed to start"))
		} else {
			fmt.Fprintln(out, utils.FormatSuccess("Plugin installed and started successfully"))
		}
	} else {
		fmt.Fprintln(out, utils.FormatSuccess("Plugin installed successfully"))
	}

	return nil
}

func (pcm *PluginCLIManager) uninstallPlugin(out io.Writer, pluginName string) error {
	fmt.Fprintln(out, utils.FormatHeader("🗑️ Uninstalling Plugin"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin", pluginName))
	fmt.Fprintln(out)

	if err := pcm.manager.UnloadPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	fmt.Fprintln(out, utils.FormatSuccess("Plugin uninstalled successfully"))
	return nil
}

func (pcm *PluginCLIManager) enablePlugin(out io.Writer, pluginName string) error {
	fmt.Fprintln(out, utils.FormatHeader("✅ Enabling Plugin"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin", pluginName))
	fmt.Fprintln(out)

	if err := pcm.manager.StartPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to enable plugin: %w", err)
	}

	fmt.Fprintln(out, utils.FormatSuccess("Plugin enabled successfully"))
	return nil
}

func (pcm *PluginCLIManager) disablePlugin(out io.Writer, pluginName string) error {
	fmt.Fprintln(out, utils.FormatHeader("⏹️ Disabling Plugin"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin", pluginName))
	fmt.Fprintln(out)

	if err := pcm.manager.StopPlugin(pluginName); err != nil {
		return fmt.Errorf("failed to disable plugin: %w", err)
	}

	fmt.Fprintln(out, utils.FormatSuccess("Plugin disabled successfully"))
	return nil
}

func (pcm *PluginCLIManager) showStatus(out io.Writer, pluginName string, detailed bool) error {
	if pluginName == "" {
		return pcm.showAllStatus(out, detailed)
	}

	fmt.Fprintln(out, utils.FormatHeader(fmt.Sprintf("📊 Plugin Status: %s", pluginName)))
	fmt.Fprintln(out)

	plugin, exists := pcm.manager.GetPlugin(pluginName)
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// Basic info
	fmt.Fprintln(out, utils.FormatKeyValue("Name", plugin.Name()))
	fmt.Fprintln(out, utils.FormatKeyValue("Version", plugin.Version()))
	fmt.Fprintln(out, utils.FormatKeyValue("Description", plugin.Description()))
	fmt.Fprintln(out, utils.FormatKeyValue("Author", plugin.Author()))

	// Status
	status, err := pcm.manager.GetPluginStatus(pluginName)
	if err != nil {
		fmt.Fprintln(out, utils.FormatError(fmt.Sprintf("Failed to get status: %v", err)))
	} else {
		fmt.Fprintln(out, utils.FormatKeyValue("State", status.State.String()))
		fmt.Fprintln(out, utils.FormatKeyValue("Message", status.Message))
		fmt.Fprintln(out, utils.FormatKeyValue("Last Updated", status.LastUpdated.Format(time.RFC3339)))
	}

	if detailed {
		// Health
		fmt.Fprintln(out)
		fmt.Fprintln(out, utils.FormatHeader("🏥 Health Status"))

		health, err := pcm.manager.GetPluginHealth(pluginName)
		if err != nil {
			fmt.Fprintln(out, utils.FormatError(fmt.Sprintf("Failed to get health: %v", err)))
		} else {
			fmt.Fprintln(out, utils.FormatKeyValue("Status", health.Status.String()))
			fmt.Fprintln(out, utils.FormatKeyValue("Message", health.Message))
			fmt.Fprintln(out, utils.FormatKeyValue("Last Check", health.LastCheck.Format(time.RFC3339)))
			fmt.Fprintln(out, utils.FormatKeyValue("Uptime", health.Uptime.String()))
		}

		// Operations
		fmt.Fprintln(out)
		fmt.Fprintln(out, utils.FormatHeader("⚙️ Available Operations"))

		operations := plugin.GetOperations()
		for _, op := range operations {
			fmt.Fprintln(out, utils.FormatSuccess(fmt.Sprintf("• %s - %s", op.Name, op.Description)))
		}
	}

	return nil
}

func (pcm *PluginCLIManager) showAllStatus(out io.Writer, detailed bool) error {
	fmt.Fprintln(out, utils.FormatHeader("📊 All Plugin Status"))
	fmt.Fprintln(out)

	plugins := pcm.manager.ListPlugins()
	if len(plugins) == 0 {
		fmt.Fprintln(out, utils.FormatInfo("No plugins installed"))
		return nil
	}

	for i, plugin := range plugins {
		if i > 0 {
			fmt.Fprintln(out, utils.FormatDivider())
		}

		fmt.Fprintln(out, utils.FormatHeader(fmt.Sprintf("Plugin: %s", plugin.Name())))

		status, _ := pcm.manager.GetPluginStatus(plugin.Name())
		if status != nil {
			fmt.Fprintln(out, utils.FormatKeyValue("State", status.State.String()))
			fmt.Fprintln(out, utils.FormatKeyValue("Message", status.Message))
		}
	}

	return nil
}

func (pcm *PluginCLIManager) showInfo(out io.Writer, pluginName string) error {
	fmt.Fprintln(out, utils.FormatHeader(fmt.Sprintf("ℹ️ Plugin Information: %s", pluginName)))
	fmt.Fprintln(out)

	plugin, exists := pcm.manager.GetPlugin(pluginName)
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// Basic information
	fmt.Fprintln(out, utils.FormatKeyValue("Name", plugin.Name()))
	fmt.Fprintln(out, utils.FormatKeyValue("Version", plugin.Version()))
	fmt.Fprintln(out, utils.FormatKeyValue("Description", plugin.Description()))
	fmt.Fprintln(out, utils.FormatKeyValue("Author", plugin.Author()))
	fmt.Fprintln(out, utils.FormatKeyValue("Repository", plugin.Repository()))
	fmt.Fprintln(out, utils.FormatKeyValue("License", plugin.License()))

	// Dependencies
	dependencies := plugin.Dependencies()
	if len(dependencies) > 0 {
		fmt.Fprintln(out, utils.FormatKeyValue("Dependencies", strings.Join(dependencies, ", ")))
	}

	// Capabilities
	capabilities := plugin.Capabilities()
	if len(capabilities) > 0 {
		fmt.Fprintln(out, utils.FormatKeyValue("Capabilities", strings.Join(capabilities, ", ")))
	}

	// Operations
	fmt.Fprintln(out)
	fmt.Fprintln(out, utils.FormatHeader("⚙️ Available Operations"))

	operations := plugin.GetOperations()
	for _, op := range operations {
		fmt.Fprintln(out, utils.FormatSuccess(fmt.Sprintf("• %s", op.Name)))
		fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("  %s", op.Description)))

		if len(op.Parameters) > 0 {
			fmt.Fprintln(out, utils.FormatInfo("  Parameters:"))
			for _, param := range op.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("    - %s (%s)%s: %s", param.Name, param.Type, required, param.Description)))
			}
		}
		fmt.Fprintln(out)
	}

	return nil
}

func (pcm *PluginCLIManager) executeOperation(out io.Writer, pluginName, operation string, params map[string]interface{}) error {
	fmt.Fprintln(out, utils.FormatHeader(fmt.Sprintf("⚡ Executing Plugin Operation")))
	fmt.Fprintln(out)

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin", pluginName))
	fmt.Fprintln(out, utils.FormatKeyValue("Operation", operation))

	if len(params) > 0 {
		paramsJSON, _ := json.MarshalIndent(params, "", "  ")
		fmt.Fprintln(out, utils.FormatKeyValue("Parameters", string(paramsJSON)))
	}

	fmt.Fprintln(out)

	result, err := pcm.manager.ExecutePluginOperation(pluginName, operation, params)
	if err != nil {
		return fmt.Errorf("operation execution failed: %w", err)
	}

	fmt.Fprintln(out, utils.FormatHeader("📤 Result"))

	if result != nil {
		resultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("%v", result)))
		} else {
			fmt.Fprintln(out, string(resultJSON))
		}
	} else {
		fmt.Fprintln(out, utils.FormatInfo("Operation completed successfully (no result)"))
	}

	return nil
}

func (pcm *PluginCLIManager) discoverPlugins(out io.Writer) error {
	fmt.Fprintln(out, utils.FormatHeader("🔍 Discovering Plugins"))
	fmt.Fprintln(out)

	directories := pcm.discovery.GetPluginDirectories()

	fmt.Fprintln(out, utils.FormatInfo("Searching in directories:"))
	for _, dir := range directories {
		fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("  • %s", dir)))
	}
	fmt.Fprintln(out)

	plugins, err := pcm.discovery.DiscoverPlugins(directories)
	if err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	if len(plugins) == 0 {
		fmt.Fprintln(out, utils.FormatInfo("No plugins found in search directories"))
		return nil
	}

	fmt.Fprintln(out, utils.FormatSuccess(fmt.Sprintf("Found %d plugin(s):", len(plugins))))
	for _, pluginPath := range plugins {
		fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("  • %s", pluginPath)))
	}

	return nil
}

func (pcm *PluginCLIManager) validatePlugin(out io.Writer, pluginPath string) error {
	fmt.Fprintln(out, utils.FormatHeader("✅ Validating Plugin"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin Path", pluginPath))
	fmt.Fprintln(out)

	if err := pcm.loader.ValidatePlugin(pluginPath); err != nil {
		fmt.Fprintln(out, utils.FormatError(fmt.Sprintf("Validation failed: %v", err)))
		return err
	}

	fmt.Fprintln(out, utils.FormatSuccess("Plugin validation passed"))
	return nil
}

func (pcm *PluginCLIManager) showMetrics(out io.Writer, pluginName string) error {
	if pluginName == "" {
		return pcm.showAllMetrics(out)
	}

	fmt.Fprintln(out, utils.FormatHeader(fmt.Sprintf("📊 Plugin Metrics: %s", pluginName)))
	fmt.Fprintln(out)

	metrics, err := pcm.manager.GetPluginMetrics(pluginName)
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	fmt.Fprintln(out, utils.FormatKeyValue("Execution Count", fmt.Sprintf("%d", metrics.ExecutionCount)))
	fmt.Fprintln(out, utils.FormatKeyValue("Error Count", fmt.Sprintf("%d", metrics.ErrorCount)))
	fmt.Fprintln(out, utils.FormatKeyValue("Success Rate", fmt.Sprintf("%.2f%%", metrics.SuccessRate*100)))
	fmt.Fprintln(out, utils.FormatKeyValue("Total Execution Time", metrics.TotalExecutionTime.String()))
	fmt.Fprintln(out, utils.FormatKeyValue("Average Execution Time", metrics.AverageExecutionTime.String()))

	if !metrics.LastExecutionTime.IsZero() {
		fmt.Fprintln(out, utils.FormatKeyValue("Last Execution", metrics.LastExecutionTime.Format(time.RFC3339)))
	}

	fmt.Fprintln(out, utils.FormatKeyValue("Memory Usage", fmt.Sprintf("%d bytes", metrics.MemoryUsage)))
	fmt.Fprintln(out, utils.FormatKeyValue("CPU Usage", fmt.Sprintf("%.2f%%", metrics.CPUUsage)))

	if len(metrics.CustomMetrics) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, utils.FormatHeader("📈 Custom Metrics"))
		for name, value := range metrics.CustomMetrics {
			fmt.Fprintln(out, utils.FormatKeyValue(name, fmt.Sprintf("%v", value)))
		}
	}

	return nil
}

func (pcm *PluginCLIManager) showAllMetrics(out io.Writer) error {
	fmt.Fprintln(out, utils.FormatHeader("📊 All Plugin Metrics"))
	fmt.Fprintln(out)

	plugins := pcm.manager.ListPlugins()
	if len(plugins) == 0 {
		fmt.Fprintln(out, utils.FormatInfo("No plugins installed"))
		return nil
	}

	for i, plugin := range plugins {
		if i > 0 {
			fmt.Fprintln(out, utils.FormatDivider())
		}

		fmt.Fprintln(out, utils.FormatHeader(fmt.Sprintf("Plugin: %s", plugin.Name())))

		metrics, err := pcm.manager.GetPluginMetrics(plugin.Name())
		if err != nil {
			fmt.Fprintln(out, utils.FormatError(fmt.Sprintf("Failed to get metrics: %v", err)))
			continue
		}

		fmt.Fprintln(out, utils.FormatKeyValue("Executions", fmt.Sprintf("%d", metrics.ExecutionCount)))
		fmt.Fprintln(out, utils.FormatKeyValue("Errors", fmt.Sprintf("%d", metrics.ErrorCount)))
		fmt.Fprintln(out, utils.FormatKeyValue("Success Rate", fmt.Sprintf("%.1f%%", metrics.SuccessRate*100)))
	}

	return nil
}

func (pcm *PluginCLIManager) showEvents(out io.Writer, limit int, eventType, source string) error {
	fmt.Fprintln(out, utils.FormatHeader("📋 Plugin Events"))
	fmt.Fprintln(out)

	// This would integrate with the event bus to show recent events
	// For now, we'll show a placeholder message
	fmt.Fprintln(out, utils.FormatInfo("Event history feature is not yet implemented"))
	fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("Would show %d events", limit)))
	if eventType != "" {
		fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("Filtered by type: %s", eventType)))
	}
	if source != "" {
		fmt.Fprintln(out, utils.FormatInfo(fmt.Sprintf("Filtered by source: %s", source)))
	}

	return nil
}

func (pcm *PluginCLIManager) createPlugin(out io.Writer, pluginName, template, outputDir string) error {
	fmt.Fprintln(out, utils.FormatHeader("🛠️ Creating Plugin"))
	fmt.Fprintln(out)

	fmt.Fprintln(out, utils.FormatKeyValue("Plugin Name", pluginName))
	fmt.Fprintln(out, utils.FormatKeyValue("Template", template))
	fmt.Fprintln(out, utils.FormatKeyValue("Output Directory", outputDir))
	fmt.Fprintln(out)

	// Create template manager
	templateManager := plugins.NewPluginTemplateManager(pcm.logger)

	// Check if template exists
	availableTemplates := templateManager.GetAvailableTemplates()
	if _, exists := availableTemplates[template]; !exists {
		fmt.Fprintln(out, utils.FormatError("Template '"+template+"' not found"))
		fmt.Fprintln(out, utils.FormatInfo("Available templates:"))
		for name, tmpl := range availableTemplates {
			fmt.Fprintln(out, utils.FormatInfo("  • "+name+": "+tmpl.Description))
		}
		return fmt.Errorf("template '%s' not found", template)
	}

	// Create scaffolding options
	options := plugins.PluginScaffoldOptions{
		PluginName:   pluginName,
		OutputDir:    outputDir,
		Template:     template,
		Variables:    make(map[string]interface{}),
		Interactive:  false,
		OverwriteAll: false,
		GitInit:      false,
		License:      "MIT",
	}

	// Scaffold the plugin
	if err := templateManager.ScaffoldPlugin(options); err != nil {
		fmt.Fprintln(out, utils.FormatError("Failed to create plugin: "+err.Error()))
		return err
	}

	fmt.Fprintln(out, utils.FormatSuccess("✅ Plugin '"+pluginName+"' created successfully!"))
	fmt.Fprintln(out)

	// Show next steps
	pluginDir := filepath.Join(outputDir, pluginName)
	fmt.Fprintln(out, utils.FormatInfo("Next steps:"))
	fmt.Fprintln(out, utils.FormatInfo("1. cd "+pluginDir))
	fmt.Fprintln(out, utils.FormatInfo("2. go mod tidy"))
	fmt.Fprintln(out, utils.FormatInfo("3. go build -buildmode=plugin -o "+pluginName+".so ."))
	fmt.Fprintln(out, utils.FormatInfo("4. nixai plugin install "+pluginName+".so"))

	return nil
}

// createIntegratedCommand creates the integrated command to show built-in plugin commands
func (pcm *PluginCLIManager) createIntegratedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "integrated",
		Short: "Show integrated plugin commands",
		Long: `Display information about integrated plugin commands that are built into nixai.
		
These commands provide plugin-like functionality but are implemented as native nixai commands
for better performance and reliability.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			integration := plugins.NewSimplePluginIntegration(nil, pcm.config, pcm.logger)
			integration.ListIntegratedPlugins()
			return nil
		},
	}

	return cmd
}
