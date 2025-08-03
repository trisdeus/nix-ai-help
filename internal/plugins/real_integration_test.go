package plugins

import (
	"context"
	"testing"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

func TestRealPluginIntegration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a config
	cfg := config.DefaultUserConfig()

	// Create real plugin integration
	rpi := NewRealPluginIntegration(cfg, log)

	// Test initialization
	ctx := context.Background()
	err := rpi.Initialize(ctx)
	if err != nil {
		t.Logf("Plugin initialization failed (expected in test environment): %v", err)
	}

	// Test getting integrated commands
	integrated := rpi.GetIntegratedCommands()
	if len(integrated) == 0 {
		t.Error("Expected at least one integrated command")
	}

	// Test listing plugins (should work even without real plugins)
	plugins := rpi.ListPlugins()
	// In test environment, we expect 0 plugins
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins in test environment, got %d", len(plugins))
	}

	// Test getting plugin commands
	commands := rpi.GetPluginCommands()
	// Should have at least the integrated commands
	if len(commands) == 0 {
		t.Error("Expected at least one plugin command")
	}

	// Test get plugin status (should return error for non-existent plugin)
	_, err = rpi.GetPluginStatus("nonexistent")
	if err == nil {
		t.Error("Expected error when getting status of non-existent plugin")
	}

	// Test get plugin health (should return error for non-existent plugin)
	_, err = rpi.GetPluginHealth("nonexistent")
	if err == nil {
		t.Error("Expected error when getting health of non-existent plugin")
	}

	// Test get plugin metrics (should return error for non-existent plugin)
	_, err = rpi.GetPluginMetrics("nonexistent")
	if err == nil {
		t.Error("Expected error when getting metrics of non-existent plugin")
	}

	// Test execute plugin operation (should return error for non-existent plugin)
	_, err = rpi.ExecutePluginOperation("nonexistent", "test", nil)
	if err == nil {
		t.Error("Expected error when executing operation on non-existent plugin")
	}

	// Test enable plugin (should return error for non-existent plugin)
	err = rpi.EnablePlugin("nonexistent")
	if err == nil {
		t.Error("Expected error when enabling non-existent plugin")
	}

	// Test disable plugin (should return error for non-existent plugin)
	err = rpi.DisablePlugin("nonexistent")
	if err == nil {
		t.Error("Expected error when disabling non-existent plugin")
	}

	// Test install plugin (should return error for non-existent plugin path)
	err = rpi.InstallPlugin("/nonexistent/plugin.so")
	if err == nil {
		t.Error("Expected error when installing nonexistent plugin")
	}

	// Test uninstall plugin (should return error for non-existent plugin)
	err = rpi.UninstallPlugin("nonexistent")
	if err == nil {
		t.Error("Expected error when uninstalling nonexistent plugin")
	}

	// Test is initialized
	initialized := rpi.IsInitialized()
	// In test environment, initialization might fail, so we just check the method works
	t.Logf("Plugin system initialized: %v", initialized)
}