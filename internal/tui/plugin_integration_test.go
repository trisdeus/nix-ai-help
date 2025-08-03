package tui

import (
	"testing"
	
	"nix-ai-help/pkg/logger"
)

func TestPluginIntegration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()
	
	// Create plugin integration
	pi := NewPluginIntegration(log)
	
	// Test initialization
	err := pi.Initialize()
	if err != nil {
		t.Errorf("PluginIntegration.Initialize() failed: %v", err)
	}
	
	// Test getting integrated commands
	integrated := pi.getIntegratedCommands()
	if len(integrated) == 0 {
		t.Error("Expected at least one integrated command")
	}
	
	// Test getting available plugin commands
	pluginCommands := pi.GetAvailablePluginCommands()
	if len(pluginCommands) == 0 {
		t.Error("Expected at least one plugin command")
	}
	
	// Test plugin suggestions
	suggestions := pi.GetPluginSuggestions("system info")
	if len(suggestions) == 0 {
		t.Error("Expected plugin suggestions for 'system info' query")
	}
	
	// Test rendering plugin status
	status := pi.RenderPluginStatus()
	if status == "" {
		t.Error("Expected non-empty plugin status")
	}
	
	// Test plugin relevance calculation
	cmd := Command{
		Name:        "system-info",
		Description: "System information and health monitoring",
		Category:    "Plugin Commands",
		Usage:       "nixai system-info [subcommand]",
		Examples:    []string{"nixai system-info health"},
	}
	
	relevance := pi.calculatePluginRelevance("system info", cmd)
	if relevance <= 0 {
		t.Errorf("Expected positive relevance score, got %f", relevance)
	}
	
	// Test plugin reason generation
	reason := pi.generatePluginReason("system info", cmd)
	if reason == "" {
		t.Error("Expected non-empty reason")
	}
	
	// Test plugin keyword extraction
	keywords := pi.extractPluginKeywords("system info", cmd)
	if len(keywords) == 0 {
		t.Error("Expected extracted keywords")
	}
	
	// Test plugin usage hint generation
	usageHint := pi.generatePluginUsageHint(cmd)
	if usageHint == "" {
		t.Error("Expected non-empty usage hint")
	}
}