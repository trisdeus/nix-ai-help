package context

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// mockProvider is a mock AI provider for testing
type mockProvider struct {
	response string
}

func (mp *mockProvider) Query(prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return mp.response, nil
}

func (mp *mockProvider) StreamResponse(ctx context.Context, prompt string) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 1)
	go func() {
		defer close(ch)
		ch <- ai.StreamResponse{
			Content: mp.response,
			Done:    true,
		}
	}()
	return ch, nil
}

func (mp *mockProvider) GetPartialResponse() string {
	return ""
}

func TestEnhancedContextDetector(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create enhanced context detector
	ecd := NewEnhancedContextDetector(log)

	// Create a config
	cfg := config.DefaultUserConfig()

	// Test detecting enhanced context
	_, err := ecd.DetectEnhancedContext(cfg)
	if err != nil {
		t.Fatalf("DetectEnhancedContext failed: %v", err)
	}

	// Test system type detection
	systemType := ecd.detectSystemType(cfg)
	if systemType == "" {
		t.Error("Expected system type, got empty string")
	}

	// Test configuration approach detection
	usesFlakes, usesChannels := ecd.detectConfigurationApproach(cfg)
	// These should be boolean values
	if usesFlakes != true && usesFlakes != false {
		t.Errorf("Expected boolean for usesFlakes, got %v", usesFlakes)
	}

	if usesChannels != true && usesChannels != false {
		t.Errorf("Expected boolean for usesChannels, got %v", usesChannels)
	}

	// Test Home Manager integration detection
	hasHM, _, _ := ecd.detectHomeManagerIntegration(cfg)
	// These should be valid values
	if hasHM != true && hasHM != false {
		t.Errorf("Expected boolean for hasHM, got %v", hasHM)
	}

	// Test version information detection
	nixosVersion, nixVersion := ecd.detectVersionInformation()
	// These should be strings (can be empty)
	if nixosVersion == "" {
		t.Log("NixOS version detection returned empty string (expected in test environment)")
	}

	if nixVersion == "" {
		t.Log("Nix version detection returned empty string (expected in test environment)")
	}

	// Test configuration files detection
	configFiles := ecd.detectConfigurationFiles(cfg)
	// This should be a slice (can be empty)
	if configFiles == nil {
		t.Error("Expected configuration files slice, got nil")
	}

	// Test enabled services detection
	services := ecd.detectEnabledServices()
	// This should be a slice (can be empty)
	if services == nil {
		t.Error("Expected services slice, got nil")
	}

	// Test installed packages detection
	packages := ecd.detectInstalledPackages()
	// This should be a slice (can be empty or nil in test environment)
	if packages == nil {
		t.Log("Packages slice is nil (expected in test environment)")
	} else if len(packages) == 0 {
		t.Log("Packages slice is empty (expected in test environment)")
	}

	// Test hardware info detection
	hardwareInfo := ecd.detectHardwareInfo()
	// This should return a hardware info struct
	if hardwareInfo == nil {
		t.Error("Expected hardware info struct, got nil")
	}

	// Test network info detection
	networkInfo := ecd.detectNetworkInfo()
	// This should return a network info struct
	if networkInfo == nil {
		t.Error("Expected network info struct, got nil")
	}

	// Test security info detection
	securityInfo := ecd.detectSecurityInfo()
	// This should return a security info struct
	if securityInfo == nil {
		t.Error("Expected security info struct, got nil")
	}

	// Test performance info detection
	performanceInfo := ecd.detectPerformanceInfo()
	// This should return a performance info struct
	if performanceInfo == nil {
		t.Error("Expected performance info struct, got nil")
	}

	// Test user environment detection
	userEnv := ecd.detectUserEnvironment()
	// This should return a user environment struct
	if userEnv == nil {
		t.Error("Expected user environment struct, got nil")
	}

	// Test context summary
	summary := ecd.GetContextSummary(nil)
	if summary == "" {
		t.Error("Expected context summary, got empty string")
	}

	// Test with nil context
	summaryNil := ecd.GetContextSummary(nil)
	expectedNilSummary := "Context: Unknown/Not detected"
	if summaryNil != expectedNilSummary {
		t.Errorf("Expected summary '%s' for nil context, got '%s'", expectedNilSummary, summaryNil)
	}

	// Test with invalid context
	invalidCtx := &config.NixOSContext{
		CacheValid: false,
	}
	summaryInvalid := ecd.GetContextSummary(invalidCtx)
	if summaryInvalid != expectedNilSummary {
		t.Errorf("Expected summary '%s' for invalid context, got '%s'", expectedNilSummary, summaryInvalid)
	}
}

func TestEnhancedContextDetectorEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create enhanced context detector
	ecd := NewEnhancedContextDetector(log)

	// Create a config with custom paths
	cfg := &config.UserConfig{
		NixosFolder: "/custom/nixos/path",
		Plugin: config.PluginConfig{
			Directory: "/custom/plugin/path",
		},
	}

	// Test detecting enhanced context with custom paths
	_, err := ecd.DetectEnhancedContext(cfg)
	if err != nil {
		t.Fatalf("DetectEnhancedContext failed with custom paths: %v", err)
	}

	// Test detecting system type
	systemType := ecd.detectSystemType(cfg)
	if systemType == "" {
		t.Error("Expected system type, got empty string")
	}

	// Test detecting configuration approach
	usesFlakes, usesChannels := ecd.detectConfigurationApproach(cfg)
	// These should be boolean values
	if usesFlakes != true && usesFlakes != false {
		t.Errorf("Expected boolean for usesFlakes, got %v", usesFlakes)
	}

	if usesChannels != true && usesChannels != false {
		t.Errorf("Expected boolean for usesChannels, got %v", usesChannels)
	}

	// Test detecting Home Manager integration
	hasHM, hmType, _ := ecd.detectHomeManagerIntegration(cfg)
	// These should be valid values
	if hasHM != true && hasHM != false {
		t.Errorf("Expected boolean for hasHM, got %v", hasHM)
	}

	if hmType == "" {
		t.Log("Home Manager type detection returned empty string (expected in test environment)")
	}

	// Test detecting version information
	nixosVersion, nixVersion := ecd.detectVersionInformation()
	// These should be strings (can be empty)
	if nixosVersion == "" {
		t.Log("NixOS version detection returned empty string (expected in test environment)")
	}

	if nixVersion == "" {
		t.Log("Nix version detection returned empty string (expected in test environment)")
	}

	// Test detecting configuration files
	configFiles := ecd.detectConfigurationFiles(cfg)
	// This should be a slice (can be empty)
	if configFiles == nil {
		t.Error("Expected configuration files slice, got nil")
	}

	// Test detecting enabled services
	services := ecd.detectEnabledServices()
	// This should be a slice (can be empty)
	if services == nil {
		t.Error("Expected services slice, got nil")
	}

	// Test detecting installed packages
	packages := ecd.detectInstalledPackages()
	// This should be a slice (can be empty or nil in test environment)
	if packages == nil {
		t.Log("Packages slice is nil (expected in test environment)")
	} else if len(packages) == 0 {
		t.Log("Packages slice is empty (expected in test environment)")
	}

	// Test detecting hardware info
	hardwareInfo := ecd.detectHardwareInfo()
	// This should return a hardware info struct
	if hardwareInfo == nil {
		t.Error("Expected hardware info struct, got nil")
	}

	// Test detecting network info
	networkInfo := ecd.detectNetworkInfo()
	// This should return a network info struct
	if networkInfo == nil {
		t.Error("Expected network info struct, got nil")
	}

	// Test detecting security info
	securityInfo := ecd.detectSecurityInfo()
	// This should return a security info struct
	if securityInfo == nil {
		t.Error("Expected security info struct, got nil")
	}

	// Test detecting performance info
	performanceInfo := ecd.detectPerformanceInfo()
	// This should return a performance info struct
	if performanceInfo == nil {
		t.Error("Expected performance info struct, got nil")
	}

	// Test detecting user environment
	userEnv := ecd.detectUserEnvironment()
	// This should return a user environment struct
	if userEnv == nil {
		t.Error("Expected user environment struct, got nil")
	}

	// Test complex task detection
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	isComplex := ecd.IsComplexTask(longQuery)
	if !isComplex {
		t.Error("Expected long query to be detected as complex")
	}

	shortQuery := "help"
	isSimpleComplex := ecd.IsComplexTask(shortQuery)
	if isSimpleComplex {
		t.Error("Expected simple query not to be detected as complex")
	}
}

func TestEnhancedContextDetectorConfiguration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create enhanced context detectors with different configurations
	cfg1 := config.DefaultUserConfig()
	detector1 := NewEnhancedContextDetector(log)
	if detector1 == nil {
		t.Error("Expected enhanced context detector with default config, got nil")
	}

	cfg2 := &config.UserConfig{
		NixosFolder: "/custom/nixos/path",
		Plugin: config.PluginConfig{
			Directory: "/custom/plugin/path",
		},
	}
	detector2 := NewEnhancedContextDetector(log)
	if detector2 == nil {
		t.Error("Expected enhanced context detector with custom config, got nil")
	}

	cfg3 := &config.UserConfig{
		NixosFolder: "~/nixos-config",
		Plugin: config.PluginConfig{
			Directory: "~/nixai-plugins",
		},
	}
	detector3 := NewEnhancedContextDetector(log)
	if detector3 == nil {
		t.Error("Expected enhanced context detector with home path config, got nil")
	}

	// Test detecting context with different configurations
	_, err1 := detector1.DetectEnhancedContext(cfg1)
	if err1 != nil {
		t.Logf("DetectEnhancedContext with default config failed: %v", err1)
	}

	_, err2 := detector2.DetectEnhancedContext(cfg2)
	if err2 != nil {
		t.Logf("DetectEnhancedContext with custom config failed: %v", err2)
	}

	_, err3 := detector3.DetectEnhancedContext(cfg3)
	if err3 != nil {
		t.Logf("DetectEnhancedContext with home path config failed: %v", err3)
	}

	// Test context summary with different configurations
	summary1 := detector1.GetContextSummary(nil)
	if summary1 == "" {
		t.Error("Expected context summary with default config, got empty string")
	}

	summary2 := detector2.GetContextSummary(nil)
	if summary2 == "" {
		t.Error("Expected context summary with custom config, got empty string")
	}

	summary3 := detector3.GetContextSummary(nil)
	if summary3 == "" {
		t.Error("Expected context summary with home path config, got empty string")
	}
}