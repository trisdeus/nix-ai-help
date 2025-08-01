package tui

import (
	"testing"
	
	"nix-ai-help/pkg/logger"
)

func TestPluginIntegration(t *testing.T) {
	// Create a plugin integration instance
	log := logger.NewLogger()
	pi := NewPluginIntegration(log)
	
	// Test that the plugin integration was created successfully
	if pi == nil {
		t.Error("Failed to create PluginIntegration")
	}
	
	// Test getting available plugin commands
	pluginCommands := pi.GetAvailablePluginCommands()
	
	// Test plugin suggestions
	suggestions := pi.GetPluginSuggestions("system info")
	
	// Test rendering plugin status
	status := pi.RenderPluginStatus()
	
	// Status should not be empty
	if status == "" {
		t.Error("Expected non-empty plugin status")
	}
	
	// We don't expect any real plugin commands in this test since it's a mock implementation
	// The important thing is that the methods don't panic and return reasonable values
	t.Logf("Got %d plugin commands and %d suggestions", len(pluginCommands), len(suggestions))
}