package plugins

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// Manager implements the PluginManager interface
type Manager struct {
	plugins     map[string]*PluginWrapper
	registry    PluginRegistry
	loader      PluginLoader
	sandbox     *Sandbox
	config      *config.UserConfig
	logger      *logger.Logger
	eventBus    *EventBus
	metrics     *MetricsCollector
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	marketplace *Marketplace // Add marketplace integration
}

// PluginWrapper wraps a plugin with additional management metadata
type PluginWrapper struct {
	Plugin     PluginInterface
	Config     PluginConfig
	Status     PluginStatus
	Health     PluginHealth
	Metrics    PluginMetrics
	LastAccess time.Time
	LoadTime   time.Time
	mutex      sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager(cfg *config.UserConfig, log *logger.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	registry := NewRegistry(log)
	loader := NewLoader(log)
	sandbox := NewSandbox(cfg, log)
	eventBus := NewEventBus(log)
	metrics := NewMetricsCollector(log)
	marketplace := NewMarketplace(log) // Create marketplace instance

	return &Manager{
		plugins:     make(map[string]*PluginWrapper),
		registry:    registry,
		loader:      loader,
		sandbox:     sandbox,
		config:      cfg,
		logger:      log,
		eventBus:    eventBus,
		metrics:     metrics,
		ctx:         ctx,
		cancel:      cancel,
		marketplace: marketplace, // Add marketplace to manager
	}
}

// LoadPlugin loads a plugin from the specified path
func (m *Manager) LoadPlugin(path string, config PluginConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.logger.Info(fmt.Sprintf("Loading plugin from path: %s", path))

	// Check if plugin is already loaded
	if _, exists := m.plugins[config.Name]; exists {
		return fmt.Errorf("plugin '%s' is already loaded", config.Name)
	}

	// Validate plugin before loading
	if err := m.loader.ValidatePlugin(path); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Load the plugin
	plugin, err := m.loader.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Create wrapper
	wrapper := &PluginWrapper{
		Plugin:     plugin,
		Config:     config,
		LoadTime:   time.Now(),
		LastAccess: time.Now(),
		Status: PluginStatus{
			State:       StateLoading,
			Message:     "Plugin loaded successfully",
			LastUpdated: time.Now(),
			Version:     plugin.Version(),
		},
	}

	// Initialize plugin in sandbox
	initCtx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	if err := m.sandbox.Execute(initCtx, plugin, func() error {
		return plugin.Initialize(initCtx, config)
	}); err != nil {
		m.loader.Unload(plugin)
		return fmt.Errorf("plugin initialization failed: %w", err)
	}

	// Register plugin
	if err := m.registry.Register(plugin); err != nil {
		m.loader.Unload(plugin)
		return fmt.Errorf("plugin registration failed: %w", err)
	}

	// Store wrapper
	m.plugins[config.Name] = wrapper
	wrapper.Status.State = StateInitializing

	// Emit event
	m.eventBus.Emit(PluginEvent{
		Type:      "plugin.loaded",
		Source:    config.Name,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plugin_name": config.Name,
			"version":     plugin.Version(),
			"path":        path,
		},
	})

	m.logger.Info(fmt.Sprintf("Plugin '%s' loaded successfully", config.Name))
	return nil
}

// UnloadPlugin unloads a plugin
func (m *Manager) UnloadPlugin(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	wrapper, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	m.logger.Info(fmt.Sprintf("Unloading plugin: %s", name))

	// Stop plugin if running
	if wrapper.Status.State == StateRunning {
		if err := wrapper.Plugin.Stop(m.ctx); err != nil {
			m.logger.Warn(fmt.Sprintf("Error stopping plugin '%s': %v", name, err))
		}
	}

	// Cleanup plugin
	if err := wrapper.Plugin.Cleanup(m.ctx); err != nil {
		m.logger.Warn(fmt.Sprintf("Error cleaning up plugin '%s': %v", name, err))
	}

	// Unregister from registry
	if err := m.registry.Unregister(name); err != nil {
		m.logger.Warn(fmt.Sprintf("Error unregistering plugin '%s': %v", name, err))
	}

	// Unload from loader
	if err := m.loader.Unload(wrapper.Plugin); err != nil {
		m.logger.Warn(fmt.Sprintf("Error unloading plugin '%s': %v", name, err))
	}

	// Remove from plugins map
	delete(m.plugins, name)

	// Emit event
	m.eventBus.Emit(PluginEvent{
		Type:      "plugin.unloaded",
		Source:    name,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plugin_name": name,
		},
	})

	m.logger.Info(fmt.Sprintf("Plugin '%s' unloaded successfully", name))
	return nil
}

// StartPlugin starts a plugin
func (m *Manager) StartPlugin(name string) error {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	if wrapper.Status.State == StateRunning {
		return fmt.Errorf("plugin '%s' is already running", name)
	}

	m.logger.Info(fmt.Sprintf("Starting plugin: %s", name))

	// Start plugin in sandbox
	startCtx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	if err := m.sandbox.Execute(startCtx, wrapper.Plugin, func() error {
		return wrapper.Plugin.Start(startCtx)
	}); err != nil {
		wrapper.Status.State = StateError
		wrapper.Status.Message = fmt.Sprintf("Failed to start: %v", err)
		wrapper.Status.LastUpdated = time.Now()
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	wrapper.Status.State = StateRunning
	wrapper.Status.Message = "Plugin started successfully"
	wrapper.Status.LastUpdated = time.Now()

	// Emit event
	m.eventBus.Emit(PluginEvent{
		Type:      "plugin.started",
		Source:    name,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plugin_name": name,
		},
	})

	m.logger.Info(fmt.Sprintf("Plugin '%s' started successfully", name))
	return nil
}

// StopPlugin stops a plugin
func (m *Manager) StopPlugin(name string) error {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	if wrapper.Status.State != StateRunning {
		return fmt.Errorf("plugin '%s' is not running", name)
	}

	m.logger.Info(fmt.Sprintf("Stopping plugin: %s", name))

	wrapper.Status.State = StateStopping
	wrapper.Status.LastUpdated = time.Now()

	// Stop plugin in sandbox
	stopCtx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()

	if err := m.sandbox.Execute(stopCtx, wrapper.Plugin, func() error {
		return wrapper.Plugin.Stop(stopCtx)
	}); err != nil {
		wrapper.Status.State = StateError
		wrapper.Status.Message = fmt.Sprintf("Failed to stop: %v", err)
		wrapper.Status.LastUpdated = time.Now()
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	wrapper.Status.State = StateStopped
	wrapper.Status.Message = "Plugin stopped successfully"
	wrapper.Status.LastUpdated = time.Now()

	// Emit event
	m.eventBus.Emit(PluginEvent{
		Type:      "plugin.stopped",
		Source:    name,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"plugin_name": name,
		},
	})

	m.logger.Info(fmt.Sprintf("Plugin '%s' stopped successfully", name))
	return nil
}

// RestartPlugin restarts a plugin
func (m *Manager) RestartPlugin(name string) error {
	if err := m.StopPlugin(name); err != nil {
		return err
	}

	// Wait a moment before restarting
	time.Sleep(100 * time.Millisecond)

	return m.StartPlugin(name)
}

// GetPlugin retrieves a plugin by name
func (m *Manager) GetPlugin(name string) (PluginInterface, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	wrapper, exists := m.plugins[name]
	if !exists {
		return nil, false
	}

	wrapper.LastAccess = time.Now()
	return wrapper.Plugin, true
}

// ListPlugins returns all loaded plugins
func (m *Manager) ListPlugins() []PluginInterface {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	plugins := make([]PluginInterface, 0, len(m.plugins))
	for _, wrapper := range m.plugins {
		plugins = append(plugins, wrapper.Plugin)
	}

	return plugins
}

// GetPluginStatus returns the status of a plugin
func (m *Manager) GetPluginStatus(name string) (*PluginStatus, error) {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	wrapper.mutex.RLock()
	defer wrapper.mutex.RUnlock()

	status := wrapper.Status
	return &status, nil
}

// GetPluginHealth returns the health status of a plugin
func (m *Manager) GetPluginHealth(name string) (*PluginHealth, error) {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	// Perform health check
	healthCtx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	var health PluginHealth
	if err := m.sandbox.Execute(healthCtx, wrapper.Plugin, func() error {
		health = wrapper.Plugin.HealthCheck(healthCtx)
		return nil
	}); err != nil {
		health = PluginHealth{
			Status:    HealthCritical,
			Message:   fmt.Sprintf("Health check failed: %v", err),
			LastCheck: time.Now(),
		}
	}

	wrapper.mutex.Lock()
	wrapper.Health = health
	wrapper.mutex.Unlock()

	return &health, nil
}

// GetPluginMetrics returns the metrics of a plugin
func (m *Manager) GetPluginMetrics(name string) (*PluginMetrics, error) {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	// Get metrics from plugin
	metrics := wrapper.Plugin.GetMetrics()

	wrapper.mutex.Lock()
	wrapper.Metrics = metrics
	wrapper.mutex.Unlock()

	return &metrics, nil
}

// UpdatePlugin updates a plugin to the latest version
func (m *Manager) UpdatePlugin(name string) error {
	// This would integrate with the package manager to update plugins
	// For now, we'll implement a basic version
	return fmt.Errorf("plugin updates not yet implemented")
}

// ConfigurePlugin updates the configuration of a plugin
func (m *Manager) ConfigurePlugin(name string, config PluginConfig) error {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	wrapper.mutex.Lock()
	defer wrapper.mutex.Unlock()

	// Stop plugin if running
	wasRunning := wrapper.Status.State == StateRunning
	if wasRunning {
		if err := wrapper.Plugin.Stop(m.ctx); err != nil {
			return fmt.Errorf("failed to stop plugin for reconfiguration: %w", err)
		}
	}

	// Reinitialize with new config
	if err := wrapper.Plugin.Initialize(m.ctx, config); err != nil {
		return fmt.Errorf("failed to reconfigure plugin: %w", err)
	}

	wrapper.Config = config

	// Restart if it was running
	if wasRunning {
		if err := wrapper.Plugin.Start(m.ctx); err != nil {
			return fmt.Errorf("failed to restart plugin after reconfiguration: %w", err)
		}
	}

	m.logger.Info(fmt.Sprintf("Plugin '%s' reconfigured successfully", name))
	return nil
}

// ExecutePluginOperation executes an operation on a plugin
func (m *Manager) ExecutePluginOperation(name, operation string, params map[string]interface{}) (interface{}, error) {
	m.mutex.RLock()
	wrapper, exists := m.plugins[name]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	if wrapper.Status.State != StateRunning {
		return nil, fmt.Errorf("plugin '%s' is not running", name)
	}

	// Execute operation in sandbox
	execCtx, cancel := context.WithTimeout(m.ctx, wrapper.Config.Resources.MaxExecutionTime)
	defer cancel()

	var result interface{}
	var execErr error

	if err := m.sandbox.Execute(execCtx, wrapper.Plugin, func() error {
		result, execErr = wrapper.Plugin.Execute(execCtx, operation, params)
		return execErr
	}); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// Update metrics
	wrapper.mutex.Lock()
	wrapper.Metrics.ExecutionCount++
	wrapper.Metrics.LastExecutionTime = time.Now()
	if execErr != nil {
		wrapper.Metrics.ErrorCount++
	}
	wrapper.Metrics.SuccessRate = float64(wrapper.Metrics.ExecutionCount-wrapper.Metrics.ErrorCount) / float64(wrapper.Metrics.ExecutionCount)
	wrapper.LastAccess = time.Now()
	wrapper.mutex.Unlock()

	return result, execErr
}

// SubscribeToEvents subscribes to plugin events
func (m *Manager) SubscribeToEvents(handler EventHandler) error {
	return m.eventBus.Subscribe(handler)
}

// UnsubscribeFromEvents unsubscribes from plugin events
func (m *Manager) UnsubscribeFromEvents(handler EventHandler) error {
	return m.eventBus.Unsubscribe(handler)
}

// GetMarketplace returns the marketplace instance
func (m *Manager) GetMarketplace() *Marketplace {
	return m.marketplace
}

// SearchPluginsInMarketplace searches for plugins in the marketplace
func (m *Manager) SearchPluginsInMarketplace(ctx context.Context, query string, filters SearchFilters, sortBy SortOption) (*SearchResult, error) {
	return m.marketplace.Search(ctx, query, filters, sortBy, 0, 100)
}

// GetPluginFromMarketplace retrieves a plugin from the marketplace by ID
func (m *Manager) GetPluginFromMarketplace(ctx context.Context, pluginID string) (*MarketplacePlugin, error) {
	return m.marketplace.GetPlugin(ctx, pluginID)
}

// GetPopularPluginsFromMarketplace retrieves popular plugins from the marketplace
func (m *Manager) GetPopularPluginsFromMarketplace(ctx context.Context, category string, limit int) ([]MarketplacePlugin, error) {
	return m.marketplace.GetPopularPlugins(ctx, category, limit)
}

// GetFeaturedPluginsFromMarketplace retrieves featured plugins from the marketplace
func (m *Manager) GetFeaturedPluginsFromMarketplace(ctx context.Context) ([]MarketplacePlugin, error) {
	return m.marketplace.GetFeaturedPlugins(ctx)
}

// GetNewPluginsFromMarketplace retrieves new plugins from the marketplace
func (m *Manager) GetNewPluginsFromMarketplace(ctx context.Context, limit int) ([]MarketplacePlugin, error) {
	return m.marketplace.GetNewPlugins(ctx, limit)
}

// GetPluginReviewsFromMarketplace retrieves reviews for a plugin from the marketplace
func (m *Manager) GetPluginReviewsFromMarketplace(ctx context.Context, pluginID string) ([]PluginReview, error) {
	return m.marketplace.GetPluginReviews(ctx, pluginID, 0, 100)
}

// GetMarketplaceStats retrieves marketplace statistics
func (m *Manager) GetMarketplaceStats(ctx context.Context) (*MarketplaceStats, error) {
	return m.marketplace.GetMarketplaceStats(ctx)
}

// InstallPluginFromMarketplace installs a plugin from the marketplace
func (m *Manager) InstallPluginFromMarketplace(ctx context.Context, pluginID string) error {
	// Get plugin from marketplace
	plugin, err := m.marketplace.GetPlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin from marketplace: %w", err)
	}

	// Download plugin binary
	resp, err := http.Get(plugin.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	// Create plugin directory if it doesn't exist
	pluginDir := filepath.Join(os.Getenv("HOME"), ".local/share/nixai/plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Save plugin to file
	pluginPath := filepath.Join(pluginDir, fmt.Sprintf("%s.so", pluginID))
	out, err := os.Create(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to create plugin file: %w", err)
	}
	defer out.Close()

	// Copy downloaded content to file
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save plugin: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(pluginPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Verify checksum
	if plugin.Checksum != "" {
		// In a real implementation, we would verify the checksum here
		m.logger.Info(fmt.Sprintf("Plugin %s downloaded to %s", pluginID, pluginPath))
	}

	return nil
}

// UpdatePluginFromMarketplace updates a plugin from the marketplace
func (m *Manager) UpdatePluginFromMarketplace(ctx context.Context, pluginID string) error {
	// Get current plugin version
	currentPlugin, exists := m.GetPlugin(pluginID)
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginID)
	}
	
	currentVersion := currentPlugin.Version()
	
	// Get plugin from marketplace
	marketplacePlugin, err := m.marketplace.GetPlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin from marketplace: %w", err)
	}
	
	// Check if update is needed
	if marketplacePlugin.Version == currentVersion {
		m.logger.Info(fmt.Sprintf("Plugin '%s' is already up to date (version %s)", pluginID, currentVersion))
		return nil
	}
	
	// Stop current plugin
	if err := m.StopPlugin(pluginID); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to stop plugin '%s' before update: %v", pluginID, err))
	}
	
	// Download updated plugin binary
	resp, err := http.Get(marketplacePlugin.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download updated plugin: %w", err)
	}
	defer resp.Body.Close()
	
	// Create plugin directory if it doesn't exist
	pluginDir := filepath.Join(os.Getenv("HOME"), ".local/share/nixai/plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}
	
	// Save updated plugin to file
	pluginPath := filepath.Join(pluginDir, fmt.Sprintf("%s.so", pluginID))
	backupPath := pluginPath + ".bak"
	
	// Backup current plugin
	if _, err := os.Stat(pluginPath); err == nil {
		if err := os.Rename(pluginPath, backupPath); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to backup current plugin: %v", err))
		}
	}
	
	out, err := os.Create(pluginPath)
	if err != nil {
		// Restore backup if creating new file fails
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			_ = os.Rename(backupPath, pluginPath)
		}
		return fmt.Errorf("failed to create updated plugin file: %w", err)
	}
	defer out.Close()
	
	// Copy downloaded content to file
	if _, err := io.Copy(out, resp.Body); err != nil {
		// Restore backup if copying fails
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			_ = os.Rename(backupPath, pluginPath)
		}
		return fmt.Errorf("failed to save updated plugin: %w", err)
	}
	
	// Set executable permissions
	if err := os.Chmod(pluginPath, 0755); err != nil {
		// Restore backup if setting permissions fails
		if _, backupErr := os.Stat(backupPath); backupErr == nil {
			_ = os.Rename(backupPath, pluginPath)
		}
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}
	
	// Remove backup
	_ = os.Remove(backupPath)
	
	// Verify checksum
	if marketplacePlugin.Checksum != "" {
		// In a real implementation, we would verify the checksum here
		m.logger.Info(fmt.Sprintf("Plugin %s updated to version %s", pluginID, marketplacePlugin.Version))
	}
	
	// Reload plugin
	if err := m.reloadPlugin(pluginID); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to reload updated plugin '%s': %v", pluginID, err))
	}
	
	return nil
}

// reloadPlugin reloads a plugin
func (m *Manager) reloadPlugin(pluginID string) error {
	// First stop the plugin if running
	if err := m.StopPlugin(pluginID); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to stop plugin '%s' before reload: %v", pluginID, err))
	}
	
	// Then unload the plugin
	if err := m.UnloadPlugin(pluginID); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to unload plugin '%s' before reload: %v", pluginID, err))
	}
	
	// Finally load the plugin again
	config := PluginConfig{
		Name:          pluginID,
		Enabled:       true,
		Version:       "",
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
	
	if err := m.LoadPlugin(pluginID, config); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to load plugin '%s' after reload: %v", pluginID, err))
		return err
	}
	
	return nil
}

// RemovePluginFromMarketplace removes a plugin installed from the marketplace
func (m *Manager) RemovePluginFromMarketplace(ctx context.Context, pluginID string) error {
	// First stop the plugin if it's running
	if err := m.StopPlugin(pluginID); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to stop plugin '%s' before removal: %v", pluginID, err))
	}
	
	// Then unload the plugin
	if err := m.UnloadPlugin(pluginID); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to unload plugin '%s' before removal: %v", pluginID, err))
	}
	
	// Remove plugin file
	pluginDir := filepath.Join(os.Getenv("HOME"), ".local/share/nixai/plugins")
	pluginPath := filepath.Join(pluginDir, fmt.Sprintf("%s.so", pluginID))
	
	if err := os.Remove(pluginPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin '%s' not found at %s", pluginID, pluginPath)
		}
		return fmt.Errorf("failed to remove plugin file: %w", err)
	}
	
	m.logger.Info(fmt.Sprintf("Plugin '%s' removed successfully from %s", pluginID, pluginPath))
	return nil
}

// ListMarketplacePlugins lists available plugins from the marketplace
func (m *Manager) ListMarketplacePlugins(ctx context.Context) ([]MarketplacePlugin, error) {
		// Search for all plugins in marketplace
	result, err := m.marketplace.Search(ctx, "", SearchFilters{}, SortByRelevance, 0, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to search marketplace: %w", err)
	}
	
	return result.Plugins, nil
}

// GetMarketplacePlugin retrieves detailed information about a specific plugin from the marketplace
func (m *Manager) GetMarketplacePlugin(ctx context.Context, pluginID string) (*MarketplacePlugin, error) {
	return m.marketplace.GetPlugin(ctx, pluginID)
}

// RateMarketplacePlugin allows users to rate a plugin in the marketplace
func (m *Manager) RateMarketplacePlugin(ctx context.Context, pluginID string, rating int, review string) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	
	// Submit rating to marketplace
	err := m.marketplace.SubmitPluginReview(ctx, pluginID, PluginReview{
		ID:        fmt.Sprintf("review-%d", time.Now().UnixNano()),
		PluginID:  pluginID,
		UserID:    "anonymous", // In a real implementation, this would be the actual user ID
		Username:  "Anonymous User",
		Rating:    rating,
		Title:     fmt.Sprintf("%d-star review", rating),
		Content:   review,
		Helpful:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Verified:  false, // In a real implementation, this would depend on user verification
	})
	
	if err != nil {
		return fmt.Errorf("failed to submit review: %w", err)
	}
	
	m.logger.Info(fmt.Sprintf("Plugin '%s' rated %d stars with review: %s", pluginID, rating, review))
	return nil
}

// ListMarketplacePlugins lists plugins from the marketplace
func (m *Manager) GetManagerMetrics() ManagerMetrics {
	// Calculate plugin states
	running := 0
	stopped := 0
	errored := 0

	for _, wrapper := range m.plugins {
		switch wrapper.Status.State {
		case StateRunning:
			running++
		case StateStopped:
			stopped++
		case StateError:
			errored++
		}
	}

	// Return metrics
	return ManagerMetrics{
		TotalPlugins:   len(m.plugins),
		RunningPlugins: running,
		StoppedPlugins: stopped,
		ErroredPlugins: errored,
		LoadedAt:       time.Now(), // This would be tracked properly in a real implementation
	}
}

// ManagerMetrics represents metrics for the plugin manager
type ManagerMetrics struct {
	TotalPlugins   int       `json:"total_plugins"`
	RunningPlugins int       `json:"running_plugins"`
	StoppedPlugins int       `json:"stopped_plugins"`
	ErroredPlugins int       `json:"errored_plugins"`
	LoadedAt       time.Time `json:"loaded_at"`
}
