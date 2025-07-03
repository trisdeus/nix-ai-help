package config

// PluginAutoDiscoveryConfig extends the existing PluginConfig with auto-discovery features
type PluginAutoDiscoveryConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	WatchDirectories bool     `yaml:"watch_directories" json:"watch_directories"`
	LoadOnStartup    bool     `yaml:"load_on_startup" json:"load_on_startup"`
	AutoEnable       bool     `yaml:"auto_enable" json:"auto_enable"`
	ExcludedPlugins  []string `yaml:"excluded_plugins" json:"excluded_plugins"`
	AutoLoadPlugins  []string `yaml:"auto_load_plugins" json:"auto_load_plugins"`
	PluginDirs       []string `yaml:"plugin_dirs" json:"plugin_dirs"`
}

// GetDefaultPluginAutoDiscoveryConfig returns the default auto-discovery configuration
func GetDefaultPluginAutoDiscoveryConfig() *PluginAutoDiscoveryConfig {
	return &PluginAutoDiscoveryConfig{
		Enabled:         true,
		WatchDirectories: false,
		LoadOnStartup:   true,
		AutoEnable:      false, // Manual enable for security
		ExcludedPlugins: []string{},
		AutoLoadPlugins: []string{
			// Auto-load our integrated commands
			"system-info",
			"package-monitor",
		},
		PluginDirs: []string{
			"/usr/share/nixai/plugins",
			"/usr/local/share/nixai/plugins",
			"~/.local/share/nixai/plugins",
			"~/.config/nixai/plugins",
			"./plugins",
		},
	}
}

// GetPluginAutoDiscoveryConfig gets auto-discovery config from the existing plugin config
func (c *UserConfig) GetPluginAutoDiscoveryConfig() PluginAutoDiscoveryConfig {
	defaults := GetDefaultPluginAutoDiscoveryConfig()
	
	// If no plugin config exists, return defaults
	if c.Plugin.Directory == "" {
		return *defaults
	}
	
	// Map existing config to auto-discovery config
	return PluginAutoDiscoveryConfig{
		Enabled:          c.Plugin.Enabled && c.Plugin.AutoDiscover,
		WatchDirectories: false, // Not implemented yet
		LoadOnStartup:    c.Plugin.Enabled,
		AutoEnable:       false, // Always manual for security
		ExcludedPlugins:  []string{},
		AutoLoadPlugins: []string{
			"system-info",
			"package-monitor",
		},
		PluginDirs: []string{
			c.Plugin.Directory,
			c.Plugin.CacheDirectory,
			c.Plugin.ConfigDirectory,
		},
	}
}