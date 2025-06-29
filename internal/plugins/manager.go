package plugins

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// Manager implements the PluginManager interface
type Manager struct {
	plugins  map[string]*PluginWrapper
	registry PluginRegistry
	loader   PluginLoader
	sandbox  *Sandbox
	config   *config.UserConfig
	logger   *logger.Logger
	eventBus *EventBus
	metrics  *MetricsCollector
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
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

	return &Manager{
		plugins:  make(map[string]*PluginWrapper),
		registry: registry,
		loader:   loader,
		sandbox:  sandbox,
		config:   cfg,
		logger:   log,
		eventBus: eventBus,
		metrics:  metrics,
		ctx:      ctx,
		cancel:   cancel,
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

// Shutdown gracefully shuts down the plugin manager
func (m *Manager) Shutdown(ctx context.Context) error {
	m.logger.Info("Shutting down plugin manager")

	// Stop all plugins
	m.mutex.RLock()
	pluginNames := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		pluginNames = append(pluginNames, name)
	}
	m.mutex.RUnlock()

	for _, name := range pluginNames {
		if err := m.StopPlugin(name); err != nil {
			m.logger.Warn(fmt.Sprintf("Error stopping plugin '%s' during shutdown: %v", name, err))
		}
		if err := m.UnloadPlugin(name); err != nil {
			m.logger.Warn(fmt.Sprintf("Error unloading plugin '%s' during shutdown: %v", name, err))
		}
	}

	// Cancel context
	m.cancel()

	m.logger.Info("Plugin manager shutdown complete")
	return nil
}

// GetManagerMetrics returns metrics about the plugin manager itself
func (m *Manager) GetManagerMetrics() ManagerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

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
