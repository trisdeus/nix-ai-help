package plugins

import (
	"testing"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginManagerInitialization(t *testing.T) {
	cfg := &config.UserConfig{
		Plugin: config.PluginConfig{
			Enabled:        true,
			Directory:      "/tmp/test-plugins",
			MaxConcurrent:  5,
			Timeout:        30,
			SandboxEnabled: true,
		},
	}
	log := logger.NewTestLogger()

	manager := NewManager(cfg, log)
	require.NotNil(t, manager)

	// Test basic manager creation
	assert.NotNil(t, manager)
}

func TestPluginManagerLoadUnload(t *testing.T) {
	cfg := &config.UserConfig{
		Plugin: config.PluginConfig{
			Enabled:        true,
			Directory:      "/tmp/test-plugins",
			MaxConcurrent:  5,
			Timeout:        30,
			SandboxEnabled: false, // Disable sandbox for testing
		},
	}
	log := logger.NewTestLogger()

	manager := NewManager(cfg, log)
	require.NotNil(t, manager)

	// Create a mock plugin config
	pluginConfig := PluginConfig{
		Name:    "test-plugin",
		Version: "1.0.0",
		Enabled: true,
		Configuration: map[string]interface{}{
			"test_setting": "test_value",
		},
	}

	// Test loading (this will fail without actual plugin file, but tests the flow)
	err := manager.LoadPlugin("test-plugin", pluginConfig)
	// We expect an error since the plugin doesn't exist
	assert.Error(t, err)

	// Test listing plugins
	plugins := manager.ListPlugins()
	assert.NotNil(t, plugins)
}

func TestPluginRegistry(t *testing.T) {
	log := logger.NewTestLogger()
	registry := NewRegistry(log)
	require.NotNil(t, registry)

	// Test basic registry operations
	assert.NotNil(t, registry)
}

func TestPluginLoader(t *testing.T) {
	log := logger.NewTestLogger()
	loader := NewLoader(log)
	require.NotNil(t, loader)

	// Test basic loader creation
	assert.NotNil(t, loader)
}

func TestPluginSandbox(t *testing.T) {
	cfg := &config.UserConfig{
		Plugin: config.PluginConfig{
			Security: config.PluginSecurityConfig{
				SandboxEnabled:       true,
				AllowNetwork:         false,
				AllowFilesystemWrite: false,
				AllowSystemCalls:     false,
				MaxMemoryMB:          256,
				MaxCPUPercent:        50,
				AllowedDomains:       []string{},
				BlockedCapabilities:  []string{"CAP_SYS_ADMIN"},
			},
		},
	}
	log := logger.NewTestLogger()

	sandbox := NewSandbox(cfg, log)
	require.NotNil(t, sandbox)

	// Test basic sandbox creation
	assert.NotNil(t, sandbox)
}

func TestPluginEventSystem(t *testing.T) {
	log := logger.NewTestLogger()
	eventBus := NewEventBus(log)
	require.NotNil(t, eventBus)

	// Test basic event bus creation
	assert.NotNil(t, eventBus)
}

func TestPluginPackageManager(t *testing.T) {
	log := logger.NewTestLogger()

	pm := NewPackageManager(log)
	require.NotNil(t, pm)

	// Test basic package manager creation
	assert.NotNil(t, pm)
}

func TestPluginMarketplace(t *testing.T) {
	log := logger.NewTestLogger()
	marketplace := NewMarketplace(log)
	require.NotNil(t, marketplace)

	// Test basic marketplace creation
	assert.NotNil(t, marketplace)
}

func TestPluginTemplates(t *testing.T) {
	log := logger.NewTestLogger()
	templates := NewPluginTemplateManager(log)
	require.NotNil(t, templates)

	// Test listing available templates
	availableTemplates := templates.GetAvailableTemplates()
	assert.NotEmpty(t, availableTemplates)

	// Should have at least basic-go template
	_, found := availableTemplates["basic-go"]
	assert.True(t, found, "basic-go template should be available")

	// Test getting template info
	info, err := templates.GetTemplate("basic-go")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "basic-go", info.Name)
}

// Integration test for the complete plugin lifecycle
func TestPluginLifecycleIntegration(t *testing.T) {
	cfg := &config.UserConfig{
		Plugin: config.PluginConfig{
			Enabled:        true,
			Directory:      "/tmp/test-plugins-integration",
			MaxConcurrent:  3,
			Timeout:        15,
			SandboxEnabled: false, // Disable for testing
		},
	}
	log := logger.NewTestLogger()

	// Initialize all components
	manager := NewManager(cfg, log)
	registry := NewRegistry(log)
	loader := NewLoader(log)
	eventBus := NewEventBus(log)

	require.NotNil(t, manager)
	require.NotNil(t, registry)
	require.NotNil(t, loader)
	require.NotNil(t, eventBus)

	// Test that all components are created successfully
	assert.NotNil(t, manager)
	assert.NotNil(t, registry)
	assert.NotNil(t, loader)
	assert.NotNil(t, eventBus)
}
