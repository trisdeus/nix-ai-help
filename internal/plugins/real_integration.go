package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// RealPluginIntegration implements the actual plugin integration with the system
type RealPluginIntegration struct {
	manager     PluginManager
	registry    PluginRegistry
	loader      PluginLoader
	discovery   *PluginDiscovery
	config      *config.UserConfig
	logger      *logger.Logger
	pluginPath  string
	initialized bool
	mutex       sync.RWMutex
}

// NewRealPluginIntegration creates a new real plugin integration handler
func NewRealPluginIntegration(cfg *config.UserConfig, log *logger.Logger) *RealPluginIntegration {
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
		config:     cfg,
		logger:     log,
		pluginPath: pluginPath,
	}
}

// Initialize sets up the real plugin system
func (rpi *RealPluginIntegration) Initialize(ctx context.Context) error {
	rpi.mutex.Lock()
	defer rpi.mutex.Unlock()

	if rpi.initialized {
		return nil
	}

	rpi.logger.Info("Initializing real plugin system")

	// Initialize plugin components
	manager := NewManager(rpi.config, rpi.logger)
	registry := NewRegistry(rpi.logger)
	loader := NewLoader(rpi.logger)
	discovery := NewPluginDiscovery(rpi.logger)

	rpi.manager = manager
	rpi.registry = registry
	rpi.loader = loader
	rpi.discovery = discovery

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

	rpi.initialized = true
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
			MaxExecutionTime: 30 * time.Second,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
			AllowedPaths:     []string{"/nix/store", "/tmp"},
			NetworkAccess:    true,
		},
		SecurityPolicy: SecurityPolicy{
			AllowFileSystem:  true,
			AllowNetwork:     true,
			AllowSystemCalls: false,
			SandboxLevel:     SandboxBasic,
		},
		UpdatePolicy: UpdatePolicy{
			AutoUpdate:         false,
			UpdateChannel:      "stable",
			CheckInterval:      24 * time.Hour,
			RequireApproval:    true,
			BackupBeforeUpdate: true,
		},
	}

	// Load the plugin
	pluginInst, err := rpi.loader.Load(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Initialize plugin context
	ctx := context.Background()
	
	// Initialize plugin
	if err := pluginInst.Initialize(ctx, config); err != nil {
		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// Register plugin
	if err := rpi.registry.Register(pluginInst); err != nil {
		return fmt.Errorf("plugin registration failed: %w", err)
	}

	// Start plugin context
	ctx = context.Background()
	
	// Start plugin
	if err := pluginInst.Start(ctx); err != nil {
		rpi.logger.Warn(fmt.Sprintf("Failed to start plugin %s: %v", pluginName, err))
	} else {
		rpi.logger.Info(fmt.Sprintf("Plugin '%s' started successfully", pluginName))
	}

	rpi.logger.Info(fmt.Sprintf("Plugin '%s' loaded successfully", pluginName))
	return nil
}

// ListPlugins returns a list of all loaded plugins
func (rpi *RealPluginIntegration) ListPlugins() []PluginInterface {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()

	if !rpi.initialized || rpi.manager == nil {
		return []PluginInterface{}
	}

	return rpi.manager.ListPlugins()
}

// GetPlugin returns a specific plugin by name
func (rpi *RealPluginIntegration) GetPlugin(name string) (PluginInterface, bool) {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()

	if !rpi.initialized || rpi.manager == nil {
		return nil, false
	}

	return rpi.manager.GetPlugin(name)
}

// InstallPlugin installs a plugin from a path
func (rpi *RealPluginIntegration) InstallPlugin(pluginPath string) error {
	rpi.mutex.Lock()
	defer rpi.mutex.Unlock()

	if !rpi.initialized || rpi.manager == nil {
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

	// Create plugin config
	pluginName := strings.TrimSuffix(filepath.Base(destPath), filepath.Ext(destPath))
	config := PluginConfig{
		Name:          pluginName,
		Enabled:       true,
		Version:       "unknown",
		Configuration: make(map[string]interface{}),
		Environment:   make(map[string]string),
		Resources: ResourceLimits{
			MaxMemoryMB:      100,
			MaxCPUPercent:    50,
			MaxExecutionTime: 30 * time.Second,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
			AllowedPaths:     []string{"/nix/store", "/tmp"},
			NetworkAccess:    true,
		},
		SecurityPolicy: SecurityPolicy{
			AllowFileSystem:  true,
			AllowNetwork:     true,
			AllowSystemCalls: false,
			SandboxLevel:     SandboxBasic,
		},
		UpdatePolicy: UpdatePolicy{
			AutoUpdate:         false,
			UpdateChannel:      "stable",
			CheckInterval:      24 * time.Hour,
			RequireApproval:    true,
			BackupBeforeUpdate: true,
		},
	}

	// Load and register the installed plugin
	return rpi.manager.LoadPlugin(destPath, config)
}

// UninstallPlugin uninstalls a plugin by name
func (rpi *RealPluginIntegration) UninstallPlugin(name string) error {
	rpi.mutex.Lock()
	defer rpi.mutex.Unlock()

	if !rpi.initialized || rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	// Unload the plugin first
	if err := rpi.manager.UnloadPlugin(name); err != nil {
		rpi.logger.Warn(fmt.Sprintf("Failed to unload plugin %s: %v", name, err))
	}

	// Remove plugin file
	pluginPath := filepath.Join(rpi.pluginPath, name+".so")
	if err := os.Remove(pluginPath); err != nil {
		rpi.logger.Warn(fmt.Sprintf("Failed to remove plugin file %s: %v", pluginPath, err))
		return err
	}

	rpi.logger.Info(fmt.Sprintf("Plugin '%s' uninstalled successfully", name))
	return nil
}

// EnablePlugin enables a plugin
func (rpi *RealPluginIntegration) EnablePlugin(name string) error {
	rpi.mutex.Lock()
	defer rpi.mutex.Unlock()

	if !rpi.initialized || rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	// In our implementation, we'll start the plugin to enable it
	return rpi.manager.StartPlugin(name)
}

// DisablePlugin disables a plugin
func (rpi *RealPluginIntegration) DisablePlugin(name string) error {
	rpi.mutex.Lock()
	defer rpi.mutex.Unlock()

	if !rpi.initialized || rpi.manager == nil {
		return fmt.Errorf("plugin manager not initialized")
	}

	// In our implementation, we'll stop the plugin to disable it
	return rpi.manager.StopPlugin(name)
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
	if rpi.initialized && rpi.manager != nil {
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

	// Sort commands alphabetically
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

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

// IsInitialized returns whether the plugin system is initialized
func (rpi *RealPluginIntegration) IsInitialized() bool {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()
	return rpi.initialized
}

// GetPluginStatus returns the status of a specific plugin
func (rpi *RealPluginIntegration) GetPluginStatus(name string) (PluginStatus, error) {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()

	if !rpi.initialized || rpi.manager == nil {
		return PluginStatus{}, fmt.Errorf("plugin manager not initialized")
	}

	pluginInst, exists := rpi.manager.GetPlugin(name)
	if !exists {
		return PluginStatus{}, fmt.Errorf("plugin '%s' not found", name)
	}

	return pluginInst.GetStatus(), nil
}

// GetPluginHealth returns the health information of a specific plugin
func (rpi *RealPluginIntegration) GetPluginHealth(name string) (PluginHealth, error) {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()

	if !rpi.initialized || rpi.manager == nil {
		return PluginHealth{}, fmt.Errorf("plugin manager not initialized")
	}

	pluginInst, exists := rpi.manager.GetPlugin(name)
	if !exists {
		return PluginHealth{}, fmt.Errorf("plugin '%s' not found", name)
	}

	return pluginInst.HealthCheck(context.Background()), nil
}

// GetPluginMetrics returns the metrics of a specific plugin
func (rpi *RealPluginIntegration) GetPluginMetrics(name string) (PluginMetrics, error) {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()

	if !rpi.initialized || rpi.manager == nil {
		return PluginMetrics{}, fmt.Errorf("plugin manager not initialized")
	}

	pluginInst, exists := rpi.manager.GetPlugin(name)
	if !exists {
		return PluginMetrics{}, fmt.Errorf("plugin '%s' not found", name)
	}

	return pluginInst.GetMetrics(), nil
}

// ExecutePluginOperation executes an operation on a plugin
func (rpi *RealPluginIntegration) ExecutePluginOperation(name, operation string, params map[string]interface{}) (interface{}, error) {
	rpi.mutex.RLock()
	defer rpi.mutex.RUnlock()

	if !rpi.initialized || rpi.manager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}

	pluginInst, exists := rpi.manager.GetPlugin(name)
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	ctx := context.Background()
	return pluginInst.Execute(ctx, operation, params)
}