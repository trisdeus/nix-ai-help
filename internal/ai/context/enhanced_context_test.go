package context

import (
	"testing"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

func TestEnhancedContextManager(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()
	
	// Create enhanced context manager
	ecm := NewEnhancedContextManager(log)
	
	// Test creating new context
	context := &EnhancedContext{
		NixOSContext: &config.NixOSContext{
			SystemType: "nixos",
			UsesFlakes:  true,
		},
		HistoricalInteractions: []Interaction{},
		UserPreferences: Preferences{
			PreferredProvider: "ollama",
		},
		SessionHistory: []string{},
		Timestamp:       time.Now(),
	}
	
	// Test saving context
	err := ecm.SaveContext(context)
	if err != nil {
		t.Errorf("Failed to save context: %v", err)
	}
	
	// Test loading context
	loadedContext, err := ecm.LoadContext()
	if err != nil {
		t.Errorf("Failed to load context: %v", err)
	}
	
	if loadedContext == nil {
		t.Error("Loaded context is nil")
	}
	
	// Test updating preferences
	ecm.UpdatePreferences(loadedContext, "I need a detailed explanation of how to configure nginx")
	
	// Check that preferences were updated
	if loadedContext.UserPreferences.ResponseDetailLevel != "detailed" {
		t.Errorf("Expected detailed preference, got %s", loadedContext.UserPreferences.ResponseDetailLevel)
	}
	
	// Test adding interaction
	ecm.AddInteraction(loadedContext, "How do I configure nginx?", "Use services.nginx.enable = true;", "NixOS context")
	
	// Check that interaction was added
	if len(loadedContext.HistoricalInteractions) != 1 {
		t.Errorf("Expected 1 interaction, got %d", len(loadedContext.HistoricalInteractions))
	}
	
	// Test building enhanced prompt
	nixosContext := &config.NixOSContext{
		SystemType: "nixos",
		UsesFlakes: true,
	}
	
	prompt := ecm.BuildEnhancedPrompt("Explain nginx configuration", nixosContext, loadedContext)
	if prompt == "" {
		t.Error("Expected non-empty enhanced prompt")
	}
	
	// Test context summary
	summary := ecm.GetContextSummary(loadedContext)
	if summary == "" {
		t.Error("Expected non-empty context summary")
	}
}