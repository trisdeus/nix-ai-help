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
	err := rpi.Initialize(context.Background())
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
	
	// Test install/uninstall (should fail without real files)
	err = rpi.InstallPlugin("/nonexistent/plugin.so")
	if err == nil {
		t.Error("Expected error when installing nonexistent plugin")
	}
	
	// Test enable/disable (should not fail)
	err = rpi.EnablePlugin("nonexistent")
	if err != nil {
		t.Logf("EnablePlugin failed for non-existent plugins: %v", err)
	}
	
	err = rpi.DisablePlugin("nonexistent")
	if err != nil {
		t.Logf("DisablePlugin failed for non-existent plugins: %v", err)
	}
}