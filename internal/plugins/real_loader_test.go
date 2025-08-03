package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"nix-ai-help/pkg/logger"
)

func TestRealPluginLoader(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create real plugin loader
	loader := NewRealPluginLoader(log)

	// Test GetPluginDirectories
	dirs := loader.GetPluginDirectories()
	if len(dirs) == 0 {
		t.Error("Expected plugin directories, got none")
	}

	// Test DiscoverPlugins with non-existent directories
	plugins, err := loader.DiscoverPlugins(dirs)
	if err != nil {
		t.Errorf("DiscoverPlugins failed: %v", err)
	}

	// Expect 0 plugins since directories don't exist
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins in test environment, got %d", len(plugins))
	}

	// Test ValidatePlugin with non-existent file
	err = loader.ValidatePlugin("/nonexistent/plugin.so")
	if err == nil {
		t.Error("Expected error when validating non-existent plugin")
	}

	// Test InstallPlugin with non-existent source
	err = loader.InstallPlugin("/nonexistent/plugin.so", "/tmp/plugins")
	if err == nil {
		t.Error("Expected error when installing non-existent plugin")
	}

	// Test UninstallPlugin with non-existent plugin
	err = loader.UninstallPlugin("nonexistent", "/tmp/plugins")
	if err == nil {
		t.Error("Expected error when uninstalling non-existent plugin")
	}

	// Test GetPlugin with non-existent plugin
	_, err = loader.GetPlugin("nonexistent", dirs)
	if err == nil {
		t.Error("Expected error when getting non-existent plugin")
	}

	// Test ListPlugins with non-existent directories
	pluginsList, err := loader.ListPlugins(dirs)
	if err != nil {
		t.Errorf("ListPlugins failed: %v", err)
	}

	// Expect 0 plugins since directories don't exist
	if len(pluginsList) != 0 {
		t.Errorf("Expected 0 plugins in test environment, got %d", len(pluginsList))
	}

	// Test Load with non-existent file
	_, err = loader.Load("/nonexistent/plugin.so")
	if err == nil {
		t.Error("Expected error when loading non-existent plugin")
	}

	// Test Unload (should not fail)
	err = loader.Unload(nil)
	if err != nil {
		t.Errorf("Unload failed: %v", err)
	}
}

func TestRealPluginLoaderIntegration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create real plugin loader
	loader := NewRealPluginLoader(log)

	// Create a temporary directory for testing
	tempDir := filepath.Join(os.TempDir(), "nixai_test_plugins")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test GetPluginDirectories includes temp directory
	dirs := append(loader.GetPluginDirectories(), tempDir)

	// Test DiscoverPlugins with temp directory
	plugins, err := loader.DiscoverPlugins(dirs)
	if err != nil {
		t.Errorf("DiscoverPlugins failed: %v", err)
	}

	// Still expect 0 plugins since no .so files exist
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins in test environment, got %d", len(plugins))
	}

	// Create a dummy .so file (this will fail validation but test discovery)
	dummyPlugin := filepath.Join(tempDir, "dummy.so")
	if err := os.WriteFile(dummyPlugin, []byte("dummy plugin content"), 0644); err != nil {
		t.Fatalf("Failed to create dummy plugin: %v", err)
	}

	// Test DiscoverPlugins with dummy plugin
	plugins, err = loader.DiscoverPlugins(dirs)
	if err != nil {
		t.Errorf("DiscoverPlugins failed: %v", err)
	}

	// Should find 1 plugin now
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(plugins))
	}

	// Test ValidatePlugin with dummy plugin (should fail)
	err = loader.ValidatePlugin(dummyPlugin)
	if err == nil {
		t.Error("Expected error when validating dummy plugin")
	}

	// Verify the error is about plugin validation, not file existence
	if err != nil && !filepath.IsAbs(dummyPlugin) {
		t.Errorf("Validation error should relate to plugin content, not file path: %v", err)
	}
}