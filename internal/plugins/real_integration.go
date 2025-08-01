package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// RealPluginIntegration handles real plugin integration with the system
type RealPluginIntegration struct {
	manager    PluginManager
	registry   PluginRegistry
	loader     PluginLoader
	discovery  *PluginDiscovery
	config     *config.UserConfig
	logger     *logger.Logger
	pluginPath string
}

// NewRealPluginIntegration creates a new real plugin integration handler
func NewRealPluginIntegration(cfg *config.UserConfig, log *logger.Logger) *RealPluginIntegration {
	// Initialize the real plugin components
	manager := NewManager(cfg, log)
	registry := NewRegistry(log)
	loader := NewLoader(log)
	discovery := NewPluginDiscovery(log)
	
	// Determine plugin path
	pluginPath := filepath.Join(os.Getenv("HOME"), ".nixai", "plugins")
	if cfg.Plugin.Directory != "" {
		pluginPath = cfg.Plugin.Directory
	}
	
	// Ensure plugin directory exists
	if err := os.MkdirAll(pluginPath, 0755); err != nil {
		log.Warn(fmt.Sprintf("Failed to create plugin directory: %v", err))
	}
	
	return &RealPluginIntegration{
		manager:    manager,
		registry:   registry,
		loader:     loader,
		discovery:  discovery,
		config:     cfg,
		logger:     log,
		pluginPath: pluginPath,
	}
}

// Initialize sets up the real plugin system
func (rpi *RealPluginIntegration) Initialize(ctx context.Context) error {
	rpi.logger.Info("Initializing real plugin system")
	
	// Discover plugins in standard directories
	pluginDirs := rpi.discovery.GetPluginDirectories()
	
	rpi.logger.Info(fmt.Sprintf("Searching for plugins in directories: %v", pluginDirs))
	
	// Discover and load plugins
	plugins, err := rpi.discovery.DiscoverPlugins(pluginDirs)
	if err != nil {
		rpi.logger.Warn(fmt.Sprintf("Plugin discovery failed: %v", err))
		return err
	}
	
	rpi.logger.Info(fmt.Sprintf("Discovered %d plugins", len(plugins)))
	
	// Load discovered plugins
	for _, pluginPath := range plugins {
		if err := rpi.loadPlugin(pluginPath); err != nil {
			rpi.logger.Warn(fmt.Sprintf("Failed to load plugin %s: %v", pluginPath, err))
		}
	}
	
	rpi.logger.Info("Real plugin system initialized")
	return nil
}

// loadPlugin loads a plugin from a path
func (rpi *RealPluginIntegration) loadPlugin(pluginPath string) error {
	rpi.logger.Info(fmt.Sprintf("Loading plugin from: %s", pluginPath))
	
	// Validate plugin first
	if err := rpi.loader.ValidatePlugin(pluginPath); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}
	
	// Extract plugin name from path
	pluginName := strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))
	
	// Create plugin config
	config := PluginConfig{
		Name:          pluginName,
		Enabled:       true,
		Version:       "unknown",
		Configuration: make(map[string]interface{}),
		Environment:   make(map[string]string),
		Resources: ResourceLimits{
			MaxMemoryMB:      100,
			MaxCPUPercent:    50,
			MaxExecutionTime: 30 * 1000000000, // 30 seconds in nanoseconds
			NetworkAccess:    true,
		},
		SecurityPolicy: SecurityPolicy{
			AllowFileSystem:  true,
			AllowNetwork:     true,
			AllowSystemCalls: false,
			SandboxLevel:     SandboxBasic,
		},
	}
	
	// Load the plugin
	plugin, err := rpi.loader.Load(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}
	
	// Initialize plugin
	ctx := context.Background()
	if err := plugin.Initialize(ctx, config); err != nil {
		return fmt.Errorf("plugin initialization failed: %w", err)
	}
	
	// Register plugin
	if err := rpi.registry.Register(plugin); err != nil {
		return fmt.Errorf("plugin registration failed: %w", err)
	}
	
	rpi.logger.Info(fmt.Sprintf("Plugin '%s' loaded successfully", pluginName))
	return nil
}

// ListPlugins returns a list of all loaded plugins
func (rpi *RealPluginIntegration) ListPlugins() []PluginInterface {
	if rpi.manager == nil {
		return []PluginInterface{}
	}
	
	return rpi.manager.ListPlugins()
}

// GetPlugin returns a specific plugin by name
func (rpi *RealPluginIntegration) GetPlugin(name string) (PluginInterface, bool) {
	if rpi.manager == nil {
		return nil, false
	}
	
	return rpi.manager.GetPlugin(name)
}

// InstallPlugin installs a plugin from a path
func (rpi *RealPluginIntegration) InstallPlugin(pluginPath string) error {
	if rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	
	// Copy plugin to plugin directory
	destPath := filepath.Join(rpi.pluginPath, filepath.Base(pluginPath))
	
	// Read source file
	srcData, err := os.ReadFile(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to read source plugin: %w", err)
	}
	
	// Write to destination
	if err := os.WriteFile(destPath, srcData, 0755); err != nil {
		return fmt.Errorf("failed to copy plugin: %w", err)
	}
	
	// Load the installed plugin
	config := PluginConfig{
		Name:          strings.TrimSuffix(filepath.Base(destPath), filepath.Ext(destPath)),
		Enabled:       true,
		Version:       "unknown",
		Configuration: make(map[string]interface{}),
		Environment:   make(map[string]string),
		Resources: ResourceLimits{
			MaxMemoryMB:      100,
			MaxCPUPercent:    50,
			MaxExecutionTime: 30 * 1000000000,
			NetworkAccess:    true,
		},
		SecurityPolicy: SecurityPolicy{
			AllowFileSystem:  true,
			AllowNetwork:     true,
			AllowSystemCalls: false,
			SandboxLevel:     SandboxBasic,
		},
	}
	
	return rpi.manager.LoadPlugin(destPath, config)
}

// UninstallPlugin uninstalls a plugin by name
func (rpi *RealPluginIntegration) UninstallPlugin(name string) error {
	if rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	
	return rpi.manager.UnloadPlugin(name)
}

// EnablePlugin enables a plugin
func (rpi *RealPluginIntegration) EnablePlugin(name string) error {
	if rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	
	// In our implementation, all loaded plugins are enabled by default
	// Future enhancement could add explicit enable/disable functionality
	rpi.logger.Info(fmt.Sprintf("Plugin '%s' enabled", name))
	return nil
}

// DisablePlugin disables a plugin
func (rpi *RealPluginIntegration) DisablePlugin(name string) error {
	if rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}
	
	// In our implementation, we'll unload the plugin to disable it
	return rpi.manager.UnloadPlugin(name)
}

// GetPluginCommands returns all commands provided by plugins
func (rpi *RealPluginIntegration) GetPluginCommands() []PluginCommand {
	var commands []PluginCommand
	
	// Get integrated plugin commands (existing built-in commands)
	integratedCommands := rpi.GetIntegratedCommands()
	for _, cmd := range integratedCommands {
		commands = append(commands, PluginCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Type:        "integrated",
			Version:     cmd.Version,
			Author:      cmd.Author,
			Examples:    cmd.Examples,
		})
	}
	
	// Get external plugin commands
	if rpi.manager != nil {
		plugins := rpi.manager.ListPlugins()
		for _, plugin := range plugins {
			operations := plugin.GetOperations()
			for _, op := range operations {
				commands = append(commands, PluginCommand{
					Name:        fmt.Sprintf("%s %s", plugin.Name(), op.Name),
					Description: op.Description,
					Type:        "external",
					Version:     plugin.Version(),
					Author:      plugin.Author(),
					Examples: []string{
						fmt.Sprintf("nixai plugin execute %s %s", plugin.Name(), op.Name),
					},
				})
			}
		}
	}
	
	return commands
}

// PluginCommand represents a command provided by a plugin
type PluginCommand struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // integrated or external
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Examples    []string `json:"examples"`
}

// GetIntegratedCommands returns information about integrated plugin commands
func (rpi *RealPluginIntegration) GetIntegratedCommands() []IntegratedPluginInfo {
	return []IntegratedPluginInfo{
		{
			Name:        "system-info",
			Description: "System information and health monitoring",
			Version:     "1.0.0",
			Author:      "NixAI Team",
			Type:        "built-in",
			Commands: []string{
				"status", "health", "cpu", "memory", "disk", "processes", "monitor", "all",
			},
			Examples: []string{
				"nixai system-info health",
				"nixai system-info status --json",
				"nixai system-info monitor --interval 5",
			},
		},
		{
			Name:        "package-monitor",
			Description: "Package monitoring and update management",
			Version:     "1.0.0",
			Author:      "NixAI Team",
			Type:        "built-in",
			Commands: []string{
				"list", "updates", "security", "analyze", "orphans", "outdated", "stats",
			},
			Examples: []string{
				"nixai package-monitor list --detailed",
				"nixai package-monitor updates --security",
				"nixai package-monitor stats --json",
			},
		},
	}
}