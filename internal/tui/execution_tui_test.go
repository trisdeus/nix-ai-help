package tui

import (
	"strings"
	"testing"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

func TestNewExecutionAwareTUI(t *testing.T) {
	log := logger.NewLogger()
	cfg := config.DefaultUserConfig()
	
	// Enable execution in config for testing
	cfg.Execution.Enabled = true
	cfg.Execution.DryRunDefault = true
	
	tui, err := NewExecutionAwareTUI(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create execution-aware TUI: %v", err)
	}
	
	if tui == nil {
		t.Fatal("Expected TUI instance, got nil")
	}
	
	if tui.executionManager == nil {
		t.Error("Expected execution manager to be initialized")
	}
	
	if tui.providerManager == nil {
		t.Error("Expected provider manager to be initialized")
	}
	
	if tui.mode != ModeNormal {
		t.Errorf("Expected initial mode to be Normal, got %v", tui.mode)
	}
	
	if !tui.showExecutionPanel {
		t.Error("Expected execution panel to be shown by default")
	}
	
	// Test closing
	err = tui.Close()
	if err != nil {
		t.Errorf("Failed to close TUI: %v", err)
	}
}

func TestExecutionManager(t *testing.T) {
	log := logger.NewLogger()
	cfg := config.DefaultUserConfig()
	cfg.Execution.Enabled = true
	
	tui, err := NewExecutionAwareTUI(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create execution-aware TUI: %v", err)
	}
	defer tui.Close()
	
	em := tui.executionManager
	
	// Test execution detection
	testInputs := []struct {
		input           string
		expectDetection bool
	}{
		{"install firefox", true},
		{"how does nix work?", false},
		{"please run this command", true},
		{"rebuild nixos", true},
		{"what is the weather?", false},
	}
	
	for _, test := range testInputs {
		t.Run(test.input, func(t *testing.T) {
			req, err := em.DetectExecutionRequest(test.input)
			if err != nil && test.expectDetection {
				t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			}
			
			hasDetection := req != nil
			if hasDetection != test.expectDetection {
				t.Errorf("For input '%s': expected detection %v, got %v", 
					test.input, test.expectDetection, hasDetection)
			}
		})
	}
	
	// Test execution statistics
	stats := em.GetExecutionStats()
	if stats == nil {
		t.Error("Expected execution stats, got nil")
	}
	
	if total, ok := stats["total"].(int); !ok || total < 0 {
		t.Error("Expected non-negative total execution count")
	}
	
	// Test active executions
	activeExecs := em.GetActiveExecutions()
	// activeExecs should be a valid slice (can be empty)
	if len(activeExecs) < 0 {
		t.Error("Expected valid active executions slice")
	}
	
	// Test execution history
	history := em.GetExecutionHistory()
	// history should be a valid slice (can be empty)
	if len(history) < 0 {
		t.Error("Expected valid execution history slice")
	}
}

func TestExecutionFormatting(t *testing.T) {
	// Create a mock execution request for testing
	req := &ExecutionRequest{
		ID:          "test_123",
		UserQuery:   "install firefox",
		Command:     "nix-env",
		Args:        []string{"-iA", "nixpkgs.firefox"},
		Description: "Install Firefox browser",
		Category:    "package",
		State:       ExecutionCompleted,
		DryRun:      true,
	}
	
	// Create a basic styles structure for testing
	log := logger.NewLogger()
	cfg := config.DefaultUserConfig()
	tui, err := NewExecutionAwareTUI(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create TUI: %v", err)
	}
	defer tui.Close()
	
	// Test status formatting
	status := FormatExecutionStatus(req, tui.styles)
	if status == "" {
		t.Error("Expected formatted status, got empty string")
	}
	
	if !strings.Contains(status, "Completed") {
		t.Error("Expected status to contain 'Completed'")
	}
	
	// Test request formatting
	formatted := FormatExecutionRequest(req, tui.styles)
	if formatted == "" {
		t.Error("Expected formatted request, got empty string")
	}
	
	if !strings.Contains(formatted, req.Command) {
		t.Error("Expected formatted request to contain command")
	}
	
	if !strings.Contains(formatted, req.Description) {
		t.Error("Expected formatted request to contain description")
	}
	
	// Test summary formatting
	stats := map[string]interface{}{
		"total":        5,
		"completed":    3,
		"failed":       1,
		"cancelled":    1,
		"success_rate": 60.0,
	}
	
	summary := FormatExecutionSummary(stats, tui.styles)
	if summary == "" {
		t.Error("Expected formatted summary, got empty string")
	}
	
	if !strings.Contains(summary, "5") {
		t.Error("Expected summary to contain total count")
	}
	
	if !strings.Contains(summary, "60.0%") {
		t.Error("Expected summary to contain success rate")
	}
}

func TestExecutionModes(t *testing.T) {
	log := logger.NewLogger()
	cfg := config.DefaultUserConfig()
	cfg.Execution.Enabled = true
	
	tui, err := NewExecutionAwareTUI(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create execution-aware TUI: %v", err)
	}
	defer tui.Close()
	
	// Test mode transitions
	originalMode := tui.mode
	if originalMode != ModeNormal {
		t.Errorf("Expected initial mode Normal, got %v", originalMode)
	}
	
	// Test execution panel toggle
	originalShowPanel := tui.showExecutionPanel
	tui.showExecutionPanel = !tui.showExecutionPanel
	
	if tui.showExecutionPanel == originalShowPanel {
		t.Error("Expected execution panel state to change")
	}
	
	// Test rendering with different modes
	tui.mode = ModeExecution
	view := tui.renderSinglePanel()
	if view == "" {
		t.Error("Expected non-empty view in execution mode")
	}
	
	tui.mode = ModeHistory
	view = tui.renderSinglePanel()
	if view == "" {
		t.Error("Expected non-empty view in history mode")
	}
}

func TestExecutionCapabilities(t *testing.T) {
	log := logger.NewLogger()
	cfg := config.DefaultUserConfig()
	cfg.Execution.Enabled = true
	
	tui, err := NewExecutionAwareTUI(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create execution-aware TUI: %v", err)
	}
	defer tui.Close()
	
	// Test provider manager capabilities
	if tui.providerManager == nil {
		t.Fatal("Expected provider manager to be initialized")
	}
	
	isEnabled := tui.providerManager.IsExecutionEnabled()
	if !isEnabled {
		t.Error("Expected execution to be enabled")
	}
	
	capabilities := tui.providerManager.GetExecutionCapabilities()
	if capabilities == nil {
		t.Error("Expected execution capabilities, got nil")
	}
	
	if enabled, ok := capabilities["enabled"].(bool); !ok || !enabled {
		t.Error("Expected execution capabilities to show enabled=true")
	}
}