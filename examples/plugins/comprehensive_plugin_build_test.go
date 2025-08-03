package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"nix-ai-help/internal/plugins"
	"nix-ai-help/pkg/logger"
)

// TestComprehensivePluginBuild tests that our comprehensive example plugin can be built and loaded
func TestComprehensivePluginBuild(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create plugin loader
	loader := plugins.NewLoader(log)

	// Build our comprehensive example plugin
	pluginPath := "./plugins/comprehensive-example.so"
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		// Try alternative path
		pluginPath = "../plugins/comprehensive-example.so"
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			t.Skipf("Plugin file %s does not exist, skipping test", pluginPath)
		}
	}

	// Validate plugin
	if err := loader.ValidatePlugin(pluginPath); err != nil {
		t.Fatalf("Plugin validation failed: %v", err)
	}

	// Load plugin
	plugin, err := loader.Load(pluginPath)
	if err != nil {
		t.Fatalf("Plugin loading failed: %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin instance, got nil")
	}

	// Test plugin metadata
	if plugin.Name() == "" {
		t.Error("Expected plugin name, got empty string")
	}

	if plugin.Version() == "" {
		t.Error("Expected plugin version, got empty string")
	}

	if plugin.Description() == "" {
		t.Error("Expected plugin description, got empty string")
	}

	if plugin.Author() == "" {
		t.Error("Expected plugin author, got empty string")
	}

	// Test plugin initialization
	ctx := context.Background()
	config := plugins.PluginConfig{
		Name:          plugin.Name(),
		Enabled:       true,
		Version:       plugin.Version(),
		Configuration: make(map[string]interface{}),
		Environment:   make(map[string]string),
		Resources: plugins.ResourceLimits{
			MaxMemoryMB:      100,
			MaxCPUPercent:    50,
			MaxExecutionTime: 30 * time.Second,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
			AllowedPaths:     []string{"/nix/store", "/tmp"},
			NetworkAccess:    true,
		},
		SecurityPolicy: plugins.SecurityPolicy{
			AllowFileSystem:  true,
			AllowNetwork:     true,
			AllowSystemCalls: false,
			SandboxLevel:     plugins.SandboxBasic,
		},
		UpdatePolicy: plugins.UpdatePolicy{
			AutoUpdate:         false,
			UpdateChannel:      "stable",
			CheckInterval:      24 * time.Hour,
			RequireApproval:    true,
			BackupBeforeUpdate: true,
		},
	}

	if err := plugin.Initialize(ctx, config); err != nil {
		t.Fatalf("Plugin initialization failed: %v", err)
	}

	// Test plugin starting
	if err := plugin.Start(ctx); err != nil {
		t.Fatalf("Plugin start failed: %v", err)
	}

	// Test plugin running state
	if !plugin.IsRunning() {
		t.Error("Expected plugin to be running after Start()")
	}

	// Test plugin operations
	operations := plugin.GetOperations()
	if len(operations) == 0 {
		t.Error("Expected plugin operations, got none")
	}

	// Test operation schemas
	for _, op := range operations {
		schema, err := plugin.GetSchema(op.Name)
		if err != nil {
			t.Errorf("Failed to get schema for operation %s: %v", op.Name, err)
		}
		if schema == nil {
			t.Errorf("Expected schema for operation %s, got nil", op.Name)
		}
	}

	// Test plugin execution
	params := map[string]interface{}{
		"name": "Alice",
	}

	result, err := plugin.Execute(ctx, "hello", params)
	if err != nil {
		t.Fatalf("Plugin execution failed: %v", err)
	}

	if result == nil {
		t.Error("Expected execution result, got nil")
	}

	// Test echo operation
	echoParams := map[string]interface{}{
		"text": "Hello, World!",
	}

	echoResult, err := plugin.Execute(ctx, "echo", echoParams)
	if err != nil {
		t.Fatalf("Echo operation failed: %v", err)
	}

	if echoResult != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", echoResult)
	}

	// Test uppercase operation
	upperParams := map[string]interface{}{
		"text": "hello world",
	}

	upperResult, err := plugin.Execute(ctx, "uppercase", upperParams)
	if err != nil {
		t.Fatalf("Uppercase operation failed: %v", err)
	}

	if upperResult != "HELLO WORLD" {
		t.Errorf("Expected 'HELLO WORLD', got '%s'", upperResult)
	}

	// Test reverse operation
	reverseParams := map[string]interface{}{
		"text": "hello",
	}

	reverseResult, err := plugin.Execute(ctx, "reverse", reverseParams)
	if err != nil {
		t.Fatalf("Reverse operation failed: %v", err)
	}

	if reverseResult != "olleh" {
		t.Errorf("Expected 'olleh', got '%s'", reverseResult)
	}

	// Test current-time operation
	timeResult, err := plugin.Execute(ctx, "current-time", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Current-time operation failed: %v", err)
	}

	if timeResult == "" {
		t.Error("Expected time result, got empty string")
	}

	// Test plugin health check
	health := plugin.HealthCheck(ctx)
	if health.Status != plugins.HealthHealthy {
		t.Errorf("Expected healthy status, got %v", health.Status)
	}

	// Test plugin metrics
	metrics := plugin.GetMetrics()
	if metrics.SuccessRate < 0 || metrics.SuccessRate > 1 {
		t.Errorf("Expected success rate between 0 and 1, got %f", metrics.SuccessRate)
	}

	// Test plugin status
	status := plugin.GetStatus()
	if status.State != plugins.StateStopped {
		t.Errorf("Expected plugin state to be stopped, got %v", status.State)
	}

	// Test plugin stopping
	if err := plugin.Stop(ctx); err != nil {
		t.Fatalf("Plugin stop failed: %v", err)
	}

	// Test plugin running state after stop
	if plugin.IsRunning() {
		t.Error("Expected plugin to not be running after Stop()")
	}

	// Test plugin cleanup
	if err := plugin.Cleanup(ctx); err != nil {
		t.Fatalf("Plugin cleanup failed: %v", err)
	}

	fmt.Println("✅ All comprehensive plugin build tests passed!")
}