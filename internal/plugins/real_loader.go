package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"nix-ai-help/pkg/logger"
)

// RealPluginLoader implements the actual plugin loading functionality
type RealPluginLoader struct {
	logger *logger.Logger
}

// NewRealPluginLoader creates a new real plugin loader
func NewRealPluginLoader(log *logger.Logger) *RealPluginLoader {
	return &RealPluginLoader{
		logger: log,
	}
}

// Load loads a plugin from the specified path
func (rpl *RealPluginLoader) Load(path string) (PluginInterface, error) {
	rpl.logger.Info(fmt.Sprintf("Loading plugin from: %s", path))

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", path)
	}

	// Load the plugin using Go's plugin package
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for the NewPlugin function
	newPluginSym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin does not export NewPlugin function: %w", err)
	}

	// Verify the function signature
	newPluginFunc, ok := newPluginSym.(func() PluginInterface)
	if !ok {
		return nil, fmt.Errorf("NewPlugin function has incorrect signature")
	}

	// Create the plugin instance
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		return nil, fmt.Errorf("NewPlugin returned nil")
	}

	// Validate the plugin implements the interface correctly
	if err := rpl.validatePluginInterface(pluginInstance); err != nil {
		return nil, fmt.Errorf("plugin interface validation failed: %w", err)
	}

	// Store the loaded plugin
	pluginName := pluginInstance.Name()
	rpl.logger.Info(fmt.Sprintf("Plugin '%s' loaded successfully", pluginName))
	return pluginInstance, nil
}

// validatePluginInterface validates that a plugin implements the interface correctly
func (rpl *RealPluginLoader) validatePluginInterface(plugin PluginInterface) error {
	// Check required methods
	if plugin.Name() == "" {
		return fmt.Errorf("plugin.Name() returned empty string")
	}

	if plugin.Version() == "" {
		rpl.logger.Warn(fmt.Sprintf("Plugin '%s' has no version", plugin.Name()))
	}

	if plugin.Description() == "" {
		rpl.logger.Warn(fmt.Sprintf("Plugin '%s' has no description", plugin.Name()))
	}

	// Try to call some methods to make sure they don't panic
	ctx := context.Background()
	
	// Test GetOperations
	ops := plugin.GetOperations()
	rpl.logger.Debug(fmt.Sprintf("Plugin '%s' has %d operations", plugin.Name(), len(ops)))
	
	// Test HealthCheck
	health := plugin.HealthCheck(ctx)
	rpl.logger.Debug(fmt.Sprintf("Plugin '%s' health: %s", plugin.Name(), health.Status))

	// Test GetMetrics
	_ = plugin.GetMetrics()
	rpl.logger.Debug(fmt.Sprintf("Plugin '%s' metrics collected", plugin.Name()))

	// Test GetStatus
	status := plugin.GetStatus()
	rpl.logger.Debug(fmt.Sprintf("Plugin '%s' status: %s", plugin.Name(), status.State))

	return nil
}

// Unload unloads a plugin (stub - Go's plugin package doesn't support unloading)
func (rpl *RealPluginLoader) Unload(plugin PluginInterface) error {
	if plugin == nil {
		return nil // Nothing to unload
	}
	
	pluginName := plugin.Name()
	rpl.logger.Info(fmt.Sprintf("Plugin '%s' unloaded", pluginName))
	
	// Note: Go's plugin package doesn't support unloading plugins at runtime
	// The plugin will remain in memory until the process exits
	return nil
}

// ValidatePlugin validates a plugin file before loading
func (rpl *RealPluginLoader) ValidatePlugin(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", path)
	}

	// Check file extension
	if filepath.Ext(path) != ".so" {
		return fmt.Errorf("plugin file must have .so extension")
	}

	// Check file permissions
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat plugin file: %w", err)
	}

	// Check if it's executable
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("plugin file is not executable")
	}

	// Try to open the plugin to verify it's a valid Go plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("plugin file is not a valid Go plugin: %w", err)
	}

	// Look for required symbols
	requiredSymbols := []string{"NewPlugin"}
	for _, symbol := range requiredSymbols {
		if _, err := p.Lookup(symbol); err != nil {
			return fmt.Errorf("plugin missing required symbol '%s': %w", symbol, err)
		}
	}

	rpl.logger.Info(fmt.Sprintf("Plugin file '%s' validated successfully", path))
	return nil
}

// DiscoverPlugins discovers plugins in standard directories
func (rpl *RealPluginLoader) DiscoverPlugins(pluginDirs []string) ([]string, error) {
	var plugins []string
	
	for _, dir := range pluginDirs {
		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			rpl.logger.Debug(fmt.Sprintf("Plugin directory does not exist: %s", dir))
			continue
		}
		
		// Read directory contents
		entries, err := os.ReadDir(dir)
		if err != nil {
			rpl.logger.Warn(fmt.Sprintf("Failed to read plugin directory %s: %v", dir, err))
			continue
		}
		
		// Look for .so files
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			name := entry.Name()
			if filepath.Ext(name) == ".so" {
				pluginPath := filepath.Join(dir, name)
				plugins = append(plugins, pluginPath)
			}
		}
	}
	
	rpl.logger.Info(fmt.Sprintf("Discovered %d plugins", len(plugins)))
	return plugins, nil
}

// GetPluginDirectories returns standard plugin directories to search
func (rpl *RealPluginLoader) GetPluginDirectories() []string {
	homeDir := os.Getenv("HOME")
	
	pluginDirs := []string{
		"/usr/share/nixai/plugins",
		"/usr/local/share/nixai/plugins",
		filepath.Join(homeDir, ".local/share/nixai/plugins"),
		"./plugins",
	}
	
	return pluginDirs
}

// InstallPlugin installs a plugin from a source path to the plugin directory
func (rpl *RealPluginLoader) InstallPlugin(sourcePath, destDir string) error {
	// Validate source plugin
	if err := rpl.ValidatePlugin(sourcePath); err != nil {
		return fmt.Errorf("source plugin validation failed: %w", err)
	}
	
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}
	
	// Copy plugin file
	sourceFile := filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, sourceFile)
	
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source plugin: %w", err)
	}
	
	if err := os.WriteFile(destPath, sourceData, 0755); err != nil {
		return fmt.Errorf("failed to write plugin to destination: %w", err)
	}
	
	rpl.logger.Info(fmt.Sprintf("Plugin installed to: %s", destPath))
	return nil
}

// UninstallPlugin uninstalls a plugin by name
func (rpl *RealPluginLoader) UninstallPlugin(name, pluginDir string) error {
	pluginPath := filepath.Join(pluginDir, name+".so")
	
	// Check if plugin exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin '%s' not found in directory: %s", name, pluginDir)
	}
	
	// Remove plugin file
	if err := os.Remove(pluginPath); err != nil {
		return fmt.Errorf("failed to remove plugin: %w", err)
	}
	
	rpl.logger.Info(fmt.Sprintf("Plugin '%s' uninstalled from: %s", name, pluginPath))
	return nil
}

// ListPlugins lists all plugins in the specified directories
func (rpl *RealPluginLoader) ListPlugins(pluginDirs []string) ([]PluginInterface, error) {
	var plugins []PluginInterface
	
	// Discover plugin files
	pluginFiles, err := rpl.DiscoverPlugins(pluginDirs)
	if err != nil {
		return nil, fmt.Errorf("plugin discovery failed: %w", err)
	}
	
	// Load each plugin
	for _, pluginPath := range pluginFiles {
		plugin, err := rpl.Load(pluginPath)
		if err != nil {
			rpl.logger.Warn(fmt.Sprintf("Failed to load plugin from %s: %v", pluginPath, err))
			continue
		}
		
		plugins = append(plugins, plugin)
	}
	
	return plugins, nil
}

// GetPlugin gets a specific plugin by name
func (rpl *RealPluginLoader) GetPlugin(name string, pluginDirs []string) (PluginInterface, error) {
	// Discover plugin files
	pluginFiles, err := rpl.DiscoverPlugins(pluginDirs)
	if err != nil {
		return nil, fmt.Errorf("plugin discovery failed: %w", err)
	}
	
	// Look for the specific plugin
	for _, pluginPath := range pluginFiles {
		// Extract plugin name from path
		pluginName := strings.TrimSuffix(filepath.Base(pluginPath), filepath.Ext(pluginPath))
		if pluginName == name {
			// Load the plugin
			plugin, err := rpl.Load(pluginPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load plugin: %w", err)
			}
			
			return plugin, nil
		}
	}
	
	return nil, fmt.Errorf("plugin '%s' not found", name)
}