package plugins

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"nix-ai-help/pkg/logger"
)

// Registry implements the PluginRegistry interface
type Registry struct {
	plugins      map[string]PluginInterface
	metadata     map[string]*PluginMetadata
	capabilities map[string][]string // capability -> plugin names
	mutex        sync.RWMutex
	logger       *logger.Logger
}

// NewRegistry creates a new plugin registry
func NewRegistry(log *logger.Logger) PluginRegistry {
	return &Registry{
		plugins:      make(map[string]PluginInterface),
		metadata:     make(map[string]*PluginMetadata),
		capabilities: make(map[string][]string),
		logger:       log,
	}
}

// Register registers a plugin in the registry
func (r *Registry) Register(plugin PluginInterface) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	// Create metadata
	metadata := &PluginMetadata{
		Name:         plugin.Name(),
		Version:      plugin.Version(),
		Description:  plugin.Description(),
		Author:       plugin.Author(),
		Repository:   plugin.Repository(),
		License:      plugin.License(),
		Dependencies: plugin.Dependencies(),
		Capabilities: plugin.Capabilities(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Register plugin
	r.plugins[name] = plugin
	r.metadata[name] = metadata

	// Index capabilities
	for _, capability := range plugin.Capabilities() {
		if r.capabilities[capability] == nil {
			r.capabilities[capability] = make([]string, 0)
		}
		r.capabilities[capability] = append(r.capabilities[capability], name)
	}

	r.logger.Info(fmt.Sprintf("Plugin '%s' registered successfully", name))
	return nil
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// Remove from capabilities index
	for _, capability := range plugin.Capabilities() {
		if pluginList, exists := r.capabilities[capability]; exists {
			for i, pluginName := range pluginList {
				if pluginName == name {
					r.capabilities[capability] = append(pluginList[:i], pluginList[i+1:]...)
					break
				}
			}
			// Remove capability if no plugins left
			if len(r.capabilities[capability]) == 0 {
				delete(r.capabilities, capability)
			}
		}
	}

	// Remove plugin and metadata
	delete(r.plugins, name)
	delete(r.metadata, name)

	r.logger.Info(fmt.Sprintf("Plugin '%s' unregistered successfully", name))
	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (PluginInterface, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugin, exists := r.plugins[name]
	return plugin, exists
}

// List returns all registered plugins
func (r *Registry) List() []PluginInterface {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	plugins := make([]PluginInterface, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}

	// Sort by name for consistent ordering
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name() < plugins[j].Name()
	})

	return plugins
}

// ListByCapability returns plugins that have a specific capability
func (r *Registry) ListByCapability(capability string) []PluginInterface {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	pluginNames, exists := r.capabilities[capability]
	if !exists {
		return []PluginInterface{}
	}

	plugins := make([]PluginInterface, 0, len(pluginNames))
	for _, name := range pluginNames {
		if plugin, exists := r.plugins[name]; exists {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// Search searches for plugins by query
func (r *Registry) Search(query string) []PluginInterface {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return r.List()
	}

	var matches []PluginInterface

	for _, plugin := range r.plugins {
		// Search in name, description, capabilities, and author
		searchText := strings.ToLower(fmt.Sprintf("%s %s %s %s",
			plugin.Name(),
			plugin.Description(),
			strings.Join(plugin.Capabilities(), " "),
			plugin.Author(),
		))

		if strings.Contains(searchText, query) {
			matches = append(matches, plugin)
		}
	}

	// Sort by relevance (name matches first, then description, etc.)
	sort.Slice(matches, func(i, j int) bool {
		nameI := strings.ToLower(matches[i].Name())
		nameJ := strings.ToLower(matches[j].Name())

		// Exact name match comes first
		if strings.Contains(nameI, query) && !strings.Contains(nameJ, query) {
			return true
		}
		if !strings.Contains(nameI, query) && strings.Contains(nameJ, query) {
			return false
		}

		// Otherwise sort alphabetically
		return nameI < nameJ
	})

	return matches
}

// GetMetadata retrieves metadata for a plugin
func (r *Registry) GetMetadata(name string) (*PluginMetadata, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	metadata, exists := r.metadata[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	// Return a copy to prevent modification
	metadataCopy := *metadata
	return &metadataCopy, nil
}

// GetCapabilities returns all available capabilities
func (r *Registry) GetCapabilities() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	capabilities := make([]string, 0, len(r.capabilities))
	for capability := range r.capabilities {
		capabilities = append(capabilities, capability)
	}

	sort.Strings(capabilities)
	return capabilities
}

// GetPluginsByTag returns plugins that have a specific tag
func (r *Registry) GetPluginsByTag(tag string) []PluginInterface {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var matches []PluginInterface

	for _, plugin := range r.plugins {
		operations := plugin.GetOperations()
		for _, op := range operations {
			for _, opTag := range op.Tags {
				if strings.EqualFold(opTag, tag) {
					matches = append(matches, plugin)
					goto nextPlugin
				}
			}
		}
	nextPlugin:
	}

	return matches
}

// GetPluginsByAuthor returns plugins by a specific author
func (r *Registry) GetPluginsByAuthor(author string) []PluginInterface {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var matches []PluginInterface

	for _, plugin := range r.plugins {
		if strings.EqualFold(plugin.Author(), author) {
			matches = append(matches, plugin)
		}
	}

	return matches
}

// GetPluginStatistics returns statistics about registered plugins
func (r *Registry) GetPluginStatistics() RegistryStatistics {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := RegistryStatistics{
		TotalPlugins:      len(r.plugins),
		TotalCapabilities: len(r.capabilities),
		Authors:           make(map[string]int),
		Licenses:          make(map[string]int),
		Capabilities:      make(map[string]int),
	}

	for _, plugin := range r.plugins {
		// Count by author
		author := plugin.Author()
		if author != "" {
			stats.Authors[author]++
		}

		// Count by license
		license := plugin.License()
		if license != "" {
			stats.Licenses[license]++
		}

		// Count by capabilities
		for _, capability := range plugin.Capabilities() {
			stats.Capabilities[capability]++
		}
	}

	return stats
}

// RegistryStatistics contains statistics about the plugin registry
type RegistryStatistics struct {
	TotalPlugins      int            `json:"total_plugins"`
	TotalCapabilities int            `json:"total_capabilities"`
	Authors           map[string]int `json:"authors"`
	Licenses          map[string]int `json:"licenses"`
	Capabilities      map[string]int `json:"capabilities"`
}

// ValidatePlugin validates a plugin meets registry requirements
func (r *Registry) ValidatePlugin(plugin PluginInterface) error {
	// Check required fields
	if strings.TrimSpace(plugin.Name()) == "" {
		return fmt.Errorf("plugin name is required")
	}

	if strings.TrimSpace(plugin.Version()) == "" {
		return fmt.Errorf("plugin version is required")
	}

	if strings.TrimSpace(plugin.Description()) == "" {
		return fmt.Errorf("plugin description is required")
	}

	// Check operations
	operations := plugin.GetOperations()
	if len(operations) == 0 {
		return fmt.Errorf("plugin must define at least one operation")
	}

	for _, op := range operations {
		if strings.TrimSpace(op.Name) == "" {
			return fmt.Errorf("operation name is required")
		}
		if strings.TrimSpace(op.Description) == "" {
			return fmt.Errorf("operation description is required for '%s'", op.Name)
		}
	}

	return nil
}

// UpdateMetadata updates metadata for a plugin
func (r *Registry) UpdateMetadata(name string, metadata *PluginMetadata) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	metadata.UpdatedAt = time.Now()
	r.metadata[name] = metadata

	return nil
}

// ExportRegistry exports the registry to a serializable format
func (r *Registry) ExportRegistry() RegistryExport {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	export := RegistryExport{
		ExportedAt: time.Now(),
		Plugins:    make([]PluginMetadata, 0, len(r.metadata)),
	}

	for _, metadata := range r.metadata {
		export.Plugins = append(export.Plugins, *metadata)
	}

	// Sort by name
	sort.Slice(export.Plugins, func(i, j int) bool {
		return export.Plugins[i].Name < export.Plugins[j].Name
	})

	return export
}

// RegistryExport represents an exported registry
type RegistryExport struct {
	ExportedAt time.Time        `json:"exported_at"`
	Plugins    []PluginMetadata `json:"plugins"`
}

// Clear removes all plugins from the registry
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.plugins = make(map[string]PluginInterface)
	r.metadata = make(map[string]*PluginMetadata)
	r.capabilities = make(map[string][]string)

	r.logger.Info("Plugin registry cleared")
}
