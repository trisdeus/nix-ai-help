package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// Loader implements the PluginLoader interface for dynamic loading of Go plugins
type Loader struct {
	loadedPlugins map[string]*plugin.Plugin
	logger        *logger.Logger
}

// NewLoader creates a new plugin loader
func NewLoader(log *logger.Logger) PluginLoader {
	return &Loader{
		loadedPlugins: make(map[string]*plugin.Plugin),
		logger:        log,
	}
}

// Load loads a plugin from the specified path
func (l *Loader) Load(path string) (PluginInterface, error) {
	l.logger.Info(fmt.Sprintf("Loading plugin from: %s", path))

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", path)
	}

	// Load the plugin
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
	if err := l.validatePluginInterface(pluginInstance); err != nil {
		return nil, fmt.Errorf("plugin interface validation failed: %w", err)
	}

	// Store the loaded plugin
	pluginName := pluginInstance.Name()
	l.loadedPlugins[pluginName] = p

	l.logger.Info(fmt.Sprintf("Plugin '%s' loaded successfully", pluginName))
	return pluginInstance, nil
}

// LoadFromSource loads a plugin from source code (for development/testing)
func (l *Loader) LoadFromSource(source []byte, name string) (PluginInterface, error) {
	// This would be used for loading plugins from source code
	// For now, we'll return an error indicating this is not implemented
	return nil, fmt.Errorf("loading from source is not yet implemented")
}

// Unload unloads a plugin
func (l *Loader) Unload(plugin PluginInterface) error {
	pluginName := plugin.Name()

	// Remove from loaded plugins map
	if _, exists := l.loadedPlugins[pluginName]; exists {
		delete(l.loadedPlugins, pluginName)
		l.logger.Info(fmt.Sprintf("Plugin '%s' unloaded", pluginName))
	}

	// Note: Go's plugin package doesn't support unloading plugins at runtime
	// The plugin will remain in memory until the process exits
	return nil
}

// Reload reloads a plugin
func (l *Loader) Reload(plugin PluginInterface) error {
	pluginName := plugin.Name()

	// Get the original path (we'd need to store this during load)
	// For now, return an error indicating reload is not supported
	return fmt.Errorf("plugin reload not supported for '%s'", pluginName)
}

// ValidatePlugin validates a plugin file before loading
func (l *Loader) ValidatePlugin(path string) error {
	l.logger.Debug(fmt.Sprintf("Validating plugin: %s", path))

	// Check file exists and is readable
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access plugin file: %w", err)
	}

	// Check it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("plugin path is not a regular file: %s", path)
	}

	// Check file extension (should be .so on Linux)
	ext := filepath.Ext(path)
	if ext != ".so" {
		return fmt.Errorf("invalid plugin file extension: %s (expected .so)", ext)
	}

	// Check file size (basic sanity check)
	if info.Size() == 0 {
		return fmt.Errorf("plugin file is empty: %s", path)
	}

	// Additional validation could include:
	// - ELF header validation
	// - Symbol table inspection
	// - Security checks

	return nil
}

// GetLoadedPlugins returns the names of all loaded plugins
func (l *Loader) GetLoadedPlugins() []string {
	names := make([]string, 0, len(l.loadedPlugins))
	for name := range l.loadedPlugins {
		names = append(names, name)
	}
	return names
}

// validatePluginInterface validates that a plugin correctly implements the PluginInterface
func (l *Loader) validatePluginInterface(plugin PluginInterface) error {
	// Check required methods exist and return appropriate types
	pluginType := reflect.TypeOf(plugin)
	if pluginType.Kind() != reflect.Ptr {
		return fmt.Errorf("plugin must be a pointer type")
	}

	// Test basic metadata methods
	name := plugin.Name()
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	version := plugin.Version()
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}

	description := plugin.Description()
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("plugin description cannot be empty")
	}

	// Test that operations are properly defined
	operations := plugin.GetOperations()
	if len(operations) == 0 {
		return fmt.Errorf("plugin must define at least one operation")
	}

	// Validate each operation
	for _, op := range operations {
		if strings.TrimSpace(op.Name) == "" {
			return fmt.Errorf("operation name cannot be empty")
		}
		if strings.TrimSpace(op.Description) == "" {
			return fmt.Errorf("operation description cannot be empty for '%s'", op.Name)
		}
	}

	// Test that health check doesn't panic
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := plugin.HealthCheck(ctx)
	if health.Status < HealthUnknown || health.Status > HealthCritical {
		return fmt.Errorf("plugin health check returned invalid status")
	}

	return nil
}

// PluginBuilder helps build plugins from templates
type PluginBuilder struct {
	logger *logger.Logger
}

// NewPluginBuilder creates a new plugin builder
func NewPluginBuilder(log *logger.Logger) *PluginBuilder {
	return &PluginBuilder{
		logger: log,
	}
}

// BuildPlugin builds a plugin from a template
func (pb *PluginBuilder) BuildPlugin(template PluginTemplate, outputPath string) error {
	// This would generate plugin code from a template
	// For now, we'll return an error indicating this is not implemented
	return fmt.Errorf("plugin building from templates is not yet implemented")
}

// PluginDiscovery helps discover plugins in directories
type PluginDiscovery struct {
	logger *logger.Logger
}

// NewPluginDiscovery creates a new plugin discovery service
func NewPluginDiscovery(log *logger.Logger) *PluginDiscovery {
	return &PluginDiscovery{
		logger: log,
	}
}

// DiscoverPlugins discovers plugins in the specified directories
func (pd *PluginDiscovery) DiscoverPlugins(directories []string) ([]string, error) {
	var plugins []string

	for _, dir := range directories {
		pd.logger.Debug(fmt.Sprintf("Discovering plugins in: %s", dir))

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			pd.logger.Warn(fmt.Sprintf("Plugin directory does not exist: %s", dir))
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Look for .so files
			if info.Mode().IsRegular() && filepath.Ext(path) == ".so" {
				plugins = append(plugins, path)
			}

			return nil
		})

		if err != nil {
			pd.logger.Warn(fmt.Sprintf("Error walking plugin directory '%s': %v", dir, err))
		}
	}

	pd.logger.Info(fmt.Sprintf("Discovered %d plugins", len(plugins)))
	return plugins, nil
}

// GetPluginDirectories returns standard plugin directories
func (pd *PluginDiscovery) GetPluginDirectories() []string {
	homeDir, _ := os.UserHomeDir()

	return []string{
		"/usr/share/nixai/plugins",
		"/usr/local/share/nixai/plugins",
		filepath.Join(homeDir, ".local", "share", "nixai", "plugins"),
		filepath.Join(homeDir, ".config", "nixai", "plugins"),
		"./plugins", // Current directory
	}
}

// GetSystemPluginDirectory returns the system-wide plugin directory
func (pd *PluginDiscovery) GetSystemPluginDirectory() string {
	return "/usr/share/nixai/plugins"
}

// GetUserPluginDirectory returns the user-specific plugin directory
func (pd *PluginDiscovery) GetUserPluginDirectory() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "share", "nixai", "plugins")
}

// EnsurePluginDirectories creates plugin directories if they don't exist
func (pd *PluginDiscovery) EnsurePluginDirectories() error {
	userDir := pd.GetUserPluginDirectory()

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user plugin directory: %w", err)
	}

	configDir := filepath.Join(filepath.Dir(userDir), "..", "..", ".config", "nixai", "plugins")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config plugin directory: %w", err)
	}

	return nil
}
