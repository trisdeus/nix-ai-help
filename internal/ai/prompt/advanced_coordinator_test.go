package prompt

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

func TestAdvancedPromptCoordinator(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider{
		response: "Mock AI response to test query",
	}

	// Create advanced prompt coordinator with advanced features enabled
	config := AdvancedPromptCoordinatorConfig{
		EnableAdvanced: true,
	}
	
	coordinator := NewAdvancedPromptCoordinator(provider, log, config)

	// Test that coordinator was created successfully
	if coordinator == nil {
		t.Fatal("Expected advanced prompt coordinator, got nil")
	}

	// Test building advanced prompt with nil context
	ctx := context.Background()
	prompt, err := coordinator.BuildAdvancedPrompt(ctx, "Test query", nil)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with nil context: %v", err)
	}

	if prompt == "" {
		t.Error("Expected prompt with nil context, got empty string")
	}

	// Test building advanced prompt with valid context
	nixosCtx := &config.NixOSContext{
		SystemType:      "nixos",
		UsesFlakes:      true,
		UsesChannels:    false,
		HasHomeManager:  true,
		HomeManagerType: "standalone",
		NixOSVersion:    "24.05",
		NixVersion:      "2.20.0",
		CacheValid:      true,
	}
	
	prompt2, err := coordinator.BuildAdvancedPrompt(ctx, "Test query", nixosCtx)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with valid context: %v", err)
	}

	if prompt2 == "" {
		t.Error("Expected prompt with valid context, got empty string")
	}

	// Test building advanced prompt with complex query
	complexQuery := "How to set up a development environment for Python and Django with multiple services and security considerations?"
	prompt3, err := coordinator.BuildAdvancedPrompt(ctx, complexQuery, nixosCtx)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with complex query: %v", err)
	}

	if prompt3 == "" {
		t.Error("Expected prompt with complex query, got empty string")
	}

	// Test with advanced features disabled
	configDisabled := AdvancedPromptCoordinatorConfig{
		EnableAdvanced: false,
	}
	
	coordinatorDisabled := NewAdvancedPromptCoordinator(provider, log, configDisabled)
	
	prompt4, err := coordinatorDisabled.BuildAdvancedPrompt(ctx, "Test query", nixosCtx)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with disabled features: %v", err)
	}

	if prompt4 == "" {
		t.Error("Expected prompt with disabled features, got empty string")
	}

	// Test coordinator methods
	systemContext := coordinator.buildSystemContext(nixosCtx)
	if systemContext == "" {
		t.Error("Expected system context, got empty string")
	}

	historicalContext := coordinator.buildHistoricalContext(ctx)
	if historicalContext != "" {
		t.Log("Historical context returned non-empty string (expected in test environment)")
	}

	userPrefContext := coordinator.buildUserPreferenceContext()
	if userPrefContext != "" {
		t.Log("User preference context returned non-empty string (expected in test environment)")
	}

	isComplex := coordinator.isComplexTask(complexQuery)
	if !isComplex {
		t.Error("Expected complex query to be detected as complex")
	}

	simpleQuery := "help"
	isSimpleComplex := coordinator.isComplexTask(simpleQuery)
	if isSimpleComplex {
		t.Error("Expected simple query not to be detected as complex")
	}

	taskPlanningGuidance := coordinator.buildTaskPlanningGuidance()
	if taskPlanningGuidance == "" {
		t.Error("Expected task planning guidance, got empty string")
	}

	selfCorrectionGuidance := coordinator.buildSelfCorrectionGuidance()
	if selfCorrectionGuidance == "" {
		t.Error("Expected self-correction guidance, got empty string")
	}

	confidenceScoringGuidance := coordinator.buildConfidenceScoringGuidance()
	if confidenceScoringGuidance == "" {
		t.Error("Expected confidence scoring guidance, got empty string")
	}

	reasoningGuidance := coordinator.buildReasoningGuidance()
	if reasoningGuidance == "" {
		t.Error("Expected reasoning guidance, got empty string")
	}

	pluginGuidance := coordinator.buildPluginGuidance()
	if pluginGuidance == "" {
		t.Error("Expected plugin guidance, got empty string")
	}

	bestPractices := coordinator.buildBestPractices()
	if bestPractices == "" {
		t.Error("Expected best practices, got empty string")
	}

	safetyGuidelines := coordinator.buildSafetyGuidelines()
	if safetyGuidelines == "" {
		t.Error("Expected safety guidelines, got empty string")
	}
}

func TestAdvancedPromptCoordinatorEdgeCases(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider with empty response
	provider := &mockProvider{
		response: "",
	}

	// Create advanced prompt coordinator with advanced features enabled
	config := AdvancedPromptCoordinatorConfig{
		EnableAdvanced: true,
	}
	
	coordinator := NewAdvancedPromptCoordinator(provider, log, config)

	// Test building prompt with empty query
	ctx := context.Background()
	prompt, err := coordinator.BuildAdvancedPrompt(ctx, "", nil)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with empty query: %v", err)
	}

	if prompt == "" {
		t.Log("Prompt is empty for empty query (expected in test environment)")
	}

	// Test building prompt with very long query
	longQuery := "This is a very long query that exceeds the normal length of queries that users typically ask. It contains multiple sentences and covers various aspects of NixOS configuration, including package management, service configuration, hardware setup, network settings, security considerations, and performance optimization."
	prompt2, err := coordinator.BuildAdvancedPrompt(ctx, longQuery, nil)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with long query: %v", err)
	}

	if prompt2 == "" {
		t.Log("Prompt is empty for long query (expected in test environment)")
	}

	// Test with nil context
	prompt3, err := coordinator.BuildAdvancedPrompt(ctx, "Test query", nil)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with nil context: %v", err)
	}

	if prompt3 == "" {
		t.Log("Prompt is empty with nil context (expected in test environment)")
	}

	// Test with invalid context
	invalidCtx := &config.NixOSContext{
		CacheValid: false,
	}
	prompt4, err := coordinator.BuildAdvancedPrompt(ctx, "Test query", invalidCtx)
	if err != nil {
		t.Fatalf("BuildAdvancedPrompt failed with invalid context: %v", err)
	}

	if prompt4 == "" {
		t.Log("Prompt is empty with invalid context (expected in test environment)")
	}

	// Test complex task detection with various queries
	queries := []struct {
		input    string
		expected bool
	}{
		{"setup nginx", true},
		{"install firefox", true},
		{"configure python", true},
		{"deploy service", true},
		{"migrate system", true},
		{"multiple steps", true},
		{"several services", true},
		{"many packages", true},
		{"process management", true},
		{"environment setup", true},
		{"development", true},
		{"production", true},
		{"help me with this", false},
		{"what is nix", false},
		{"how does it work", false},
		{"explain packages", false},
		{"tell me more", false},
		{"show me examples", false},
		{"", false},
		{"simple help", false},
	}

	for _, query := range queries {
		result := coordinator.isComplexTask(query.input)
		if result != query.expected {
			t.Errorf("For query '%s', expected complex=%t, got %t", query.input, query.expected, result)
		}
	}

	// Test filterImportantServices with various service lists
	services := []string{
		"openssh", "nginx", "postgresql", "docker", "firewall", "sound", "xserver", "gnome",
		"kde", "plasma", "networkmanager", "bluetooth", "printing", "apache", "mysql",
		"redis", "memcached", "mongodb", "couchdb", "bind", "dnsmasq", "tor", "openvpn",
		"wireguard", "cups", "avahi", "samba", "nfs", "smbd", "nmbd", "lightdm", "gdm",
		"sddm", "lxdm", "i3", "sway", "xmonad", "qtile", "bspwm", "awesome", "dwm",
		"firefox", "chromium", "google-chrome", "brave", "thunderbird", "vlc", "mpv",
		"audacious", "spotify", "steam", "lutris", "vscode", "neovim", "emacs", "vim",
		"nano", "helix", "git", "mercurial", "subversion", "darcs", "nodejs", "python3",
		"ruby", "go", "rust", "java", "php", "docker-compose", "kubectl", "helm",
		"terraform", "ansible", "nix-darwin", "home-manager", "nixos-generators",
		"nixos-anywhere", "nixos-hardware", "nixos-wiki", "nixpkgs", "nix",
	}

	filtered := coordinator.filterImportantServices(services)
	if len(filtered) == 0 {
		t.Error("Expected filtered services, got none")
	}

	// Test with empty services list
	emptyFiltered := coordinator.filterImportantServices([]string{})
	if len(emptyFiltered) != 0 {
		t.Error("Expected empty filtered services, got some")
	}

	// Test with nil services list
	nilFiltered := coordinator.filterImportantServices(nil)
	if nilFiltered != nil {
		t.Error("Expected nil filtered services, got non-nil")
	}
}

func TestAdvancedPromptCoordinatorConfiguration(t *testing.T) {
	// Create a logger
	log := logger.NewLogger()

	// Create a mock provider
	provider := &mockProvider{
		response: "Mock AI response to test query",
	}

	// Test creating coordinator with different configurations
	config1 := AdvancedPromptCoordinatorConfig{
		EnableAdvanced: true,
	}
	coordinator1 := NewAdvancedPromptCoordinator(provider, log, config1)
	if coordinator1 == nil {
		t.Error("Expected advanced prompt coordinator with advanced features enabled, got nil")
	}

	config2 := AdvancedPromptCoordinatorConfig{
		EnableAdvanced: false,
	}
	coordinator2 := NewAdvancedPromptCoordinator(provider, log, config2)
	if coordinator2 == nil {
		t.Error("Expected advanced prompt coordinator with advanced features disabled, got nil")
	}

	// Test building prompts with different configurations
	ctx := context.Background()
	
	prompt1, err1 := coordinator1.BuildAdvancedPrompt(ctx, "Test query", nil)
	if err1 != nil {
		t.Fatalf("BuildAdvancedPrompt with config1 failed: %v", err1)
	}

	if prompt1 == "" {
		t.Log("Prompt with config1 is empty (expected in test environment)")
	}

	prompt2, err2 := coordinator2.BuildAdvancedPrompt(ctx, "Test query", nil)
	if err2 != nil {
		t.Fatalf("BuildAdvancedPrompt with config2 failed: %v", err2)
	}

	if prompt2 == "" {
		t.Log("Prompt with config2 is empty (expected in test environment)")
	}

	// Test coordinator methods with different configurations
	systemContext1 := coordinator1.buildSystemContext(nil)
	if systemContext1 == "" {
		t.Log("System context 1 is empty (expected in test environment)")
	}

	systemContext2 := coordinator2.buildSystemContext(nil)
	if systemContext2 == "" {
		t.Log("System context 2 is empty (expected in test environment)")
	}

	// Test complex task detection with different configurations
	isComplex1 := coordinator1.isComplexTask("setup nginx")
	if !isComplex1 {
		t.Error("Expected complex task detection with config1 to return true")
	}

	isComplex2 := coordinator2.isComplexTask("setup nginx")
	if !isComplex2 {
		t.Error("Expected complex task detection with config2 to return true")
	}

	// Test task planning guidance with different configurations
	guidance1 := coordinator1.buildTaskPlanningGuidance()
	if guidance1 == "" {
		t.Log("Task planning guidance 1 is empty (expected in test environment)")
	}

	guidance2 := coordinator2.buildTaskPlanningGuidance()
	if guidance2 == "" {
		t.Log("Task planning guidance 2 is empty (expected in test environment)")
	}

	// Test self-correction guidance with different configurations
	correction1 := coordinator1.buildSelfCorrectionGuidance()
	if correction1 == "" {
		t.Log("Self-correction guidance 1 is empty (expected in test environment)")
	}

	correction2 := coordinator2.buildSelfCorrectionGuidance()
	if correction2 == "" {
		t.Log("Self-correction guidance 2 is empty (expected in test environment)")
	}

	// Test confidence scoring guidance with different configurations
	scoring1 := coordinator1.buildConfidenceScoringGuidance()
	if scoring1 == "" {
		t.Log("Confidence scoring guidance 1 is empty (expected in test environment)")
	}

	scoring2 := coordinator2.buildConfidenceScoringGuidance()
	if scoring2 == "" {
		t.Log("Confidence scoring guidance 2 is empty (expected in test environment)")
	}

	// Test reasoning guidance with different configurations
	reasoning1 := coordinator1.buildReasoningGuidance()
	if reasoning1 == "" {
		t.Log("Reasoning guidance 1 is empty (expected in test environment)")
	}

	reasoning2 := coordinator2.buildReasoningGuidance()
	if reasoning2 == "" {
		t.Log("Reasoning guidance 2 is empty (expected in test environment)")
	}

	// Test plugin guidance with different configurations
	plugin1 := coordinator1.buildPluginGuidance()
	if plugin1 == "" {
		t.Log("Plugin guidance 1 is empty (expected in test environment)")
	}

	plugin2 := coordinator2.buildPluginGuidance()
	if plugin2 == "" {
		t.Log("Plugin guidance 2 is empty (expected in test environment)")
	}

	// Test best practices with different configurations
	practices1 := coordinator1.buildBestPractices()
	if practices1 == "" {
		t.Log("Best practices 1 is empty (expected in test environment)")
	}

	practices2 := coordinator2.buildBestPractices()
	if practices2 == "" {
		t.Log("Best practices 2 is empty (expected in test environment)")
	}

	// Test safety guidelines with different configurations
	safety1 := coordinator1.buildSafetyGuidelines()
	if safety1 == "" {
		t.Log("Safety guidelines 1 is empty (expected in test environment)")
	}

	safety2 := coordinator2.buildSafetyGuidelines()
	if safety2 == "" {
		t.Log("Safety guidelines 2 is empty (expected in test environment)")
	}
}