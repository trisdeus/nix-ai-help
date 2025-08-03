package advanced

import (
	"testing"
	"time"
)

func TestReasoningChainVisualizer(t *testing.T) {
	// Create theme styles
	styles := NewThemeStyles()

	// Create visualizer
	visualizer := NewReasoningChainVisualizer(styles)

	// Create a sample reasoning chain
	chain := &ReasoningChain{
		Task:      "How to configure nginx in NixOS?",
		Steps: []ReasoningStep{
			{
				StepNumber: 1,
				Title:      "Problem Analysis",
				Content:    "Understanding the user's nginx configuration needs",
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				StepNumber: 2,
				Title:      "Solution Generation",
				Content: `services.nginx = {
  enable = true;
  virtualHosts = {
    "localhost" = {
      root = "/var/www";
    };
  };
};`,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			},
		},
		FinalAnswer: `To configure nginx in NixOS, add the following to your configuration.nix:

services.nginx = {
  enable = true;
  virtualHosts = {
    "localhost" = {
      root = "/var/www";
    };
  };
};

Then rebuild your system with:
sudo nixos-rebuild switch`,
		Confidence:   0.85,
		QualityScore: 8,
		TotalTime:    "1.23s",
	}

	// Test visualizing reasoning chain
	visualization := visualizer.VisualizeReasoningChain(chain)
	if visualization == "" {
		t.Error("Expected visualization, got empty string")
	}

	// Test with nil chain
	nilVisualization := visualizer.VisualizeReasoningChain(nil)
	if nilVisualization == "" {
		t.Error("Expected error message for nil chain, got empty string")
	}
}

func TestConfidenceScoreVisualizer(t *testing.T) {
	// Create theme styles
	styles := NewThemeStyles()

	// Create visualizer
	visualizer := NewReasoningChainVisualizer(styles)

	// Create a sample confidence score
	score := &ConfidenceScore{
		Score:       0.85,
		Explanation: "High confidence in technical accuracy",
		Factors: []Factor{
			{
				Name:        "technical_accuracy",
				Description: "Technical accuracy of the information provided",
				Value:       0.9,
				Weight:      0.25,
				Contribution: 0.225,
			},
			{
				Name:        "completeness",
				Description: "Completeness of the response",
				Value:       0.8,
				Weight:      0.2,
				Contribution: 0.16,
			},
		},
		QualityIndicators: []string{
			"Uses specific NixOS configuration syntax",
			"Includes clear step-by-step instructions",
		},
		Warnings: []string{
			"No mention of security considerations",
		},
	}

	// Test visualizing confidence score
	visualization := visualizer.VisualizeConfidenceScore(score)
	if visualization == "" {
		t.Error("Expected visualization, got empty string")
	}

	// Test with nil score
	nilVisualization := visualizer.VisualizeConfidenceScore(nil)
	if nilVisualization == "" {
		t.Error("Expected error message for nil score, got empty string")
	}
}

func TestCorrectionsVisualizer(t *testing.T) {
	// Create theme styles
	styles := NewThemeStyles()

	// Create visualizer
	visualizer := NewReasoningChainVisualizer(styles)

	// Create sample corrections
	corrections := []Correction{
		{
			Original:    "Use nix-env -i to install packages",
			Correction:  "Use nix profile install or nix-env -iA for better pinning",
			Explanation: "nix-env -i is discouraged in favor of more specific package references",
			Confidence:  0.95,
			Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			Original:    "Edit /etc/nixos/configuration.nix directly",
			Correction:  "Edit your configuration.nix file in /etc/nixos/",
			Explanation: "Clarify the exact path to avoid confusion",
			Confidence:  0.85,
			Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// Test visualizing corrections
	visualization := visualizer.VisualizeCorrections(corrections)
	if visualization == "" {
		t.Error("Expected visualization, got empty string")
	}

	// Test with empty corrections
	emptyVisualization := visualizer.VisualizeCorrections([]Correction{})
	if emptyVisualization == "" {
		t.Error("Expected success message for empty corrections, got empty string")
	}
}

func TestTaskPlanVisualizer(t *testing.T) {
	// Create theme styles
	styles := NewThemeStyles()

	// Create visualizer
	visualizer := NewReasoningChainVisualizer(styles)

	// Create a sample task plan
	plan := &TaskPlan{
		ID:                   "plan-123",
		Title:               "Python Development Environment Setup",
		Description:        "Set up a complete Python development environment with Django",
		Status:              "in-progress",
		Progress:            0.65,
		EstimatedTotalTime:  "15m",
		ActualTotalTime:     "10m",
		StartTime:           time.Now().Add(-10 * time.Minute).Format("2006-01-02 15:04:05"),
		EndTime:             "",
		Tasks: []Task{
			{
				ID:          "task-1",
				Title:       "Install Python",
				Description: "Install Python runtime and development tools",
				Command:     "nix-env -iA nixpkgs.python3",
				Status:      "completed",
				Prerequisites: []string{},
				DependsOn:   []string{},
				EstimatedTime: "2m",
				ActualTime:  "1m30s",
				Result:      "Python 3.11 installed successfully",
				Error:       "",
			},
			{
				ID:          "task-2",
				Title:       "Set up Virtual Environment",
				Description: "Create and configure virtual environment for project isolation",
				Command:     "python -m venv .venv && source .venv/bin/activate",
				Status:      "in-progress",
				Prerequisites: []string{"task-1"},
				DependsOn:   []string{"task-1"},
				EstimatedTime: "3m",
				ActualTime:  "2m15s",
				Result:      "Virtual environment created",
				Error:       "",
			},
			{
				ID:          "task-3",
				Title:       "Install Django",
				Description: "Install Django framework and related dependencies",
				Command:     "pip install django",
				Status:      "pending",
				Prerequisites: []string{"task-2"},
				DependsOn:   []string{"task-2"},
				EstimatedTime: "5m",
				ActualTime:  "",
				Result:      "",
				Error:       "",
			},
		},
	}

	// Test visualizing task plan
	visualization := visualizer.VisualizeTaskPlan(plan)
	if visualization == "" {
		t.Error("Expected visualization, got empty string")
	}

	// Test with nil plan
	nilVisualization := visualizer.VisualizeTaskPlan(nil)
	if nilVisualization == "" {
		t.Error("Expected error message for nil plan, got empty string")
	}
}

func TestThemeStyles(t *testing.T) {
	// Test creating theme styles
	styles := NewThemeStyles()

	// Test that all styles are created by rendering some text with them
	text := "test"

	// Test header style
	headerRender := styles.header.Render(text)
	if headerRender == "" {
		t.Error("Expected header style to render text, got empty string")
	}

	// Test error style
	errorRender := styles.error.Render(text)
	if errorRender == "" {
		t.Error("Expected error style to render text, got empty string")
	}

	// Test success style
	successRender := styles.success.Render(text)
	if successRender == "" {
		t.Error("Expected success style to render text, got empty string")
	}

	// Test warning style
	warningRender := styles.warning.Render(text)
	if warningRender == "" {
		t.Error("Expected warning style to render text, got empty string")
	}

	// Test info style
	infoRender := styles.info.Render(text)
	if infoRender == "" {
		t.Error("Expected info style to render text, got empty string")
	}

	// Test accent style
	accentRender := styles.accent.Render(text)
	if accentRender == "" {
		t.Error("Expected accent style to render text, got empty string")
	}

	// Test muted style
	mutedRender := styles.muted.Render(text)
	if mutedRender == "" {
		t.Error("Expected muted style to render text, got empty string")
	}

	// Test selected style
	selectedRender := styles.selected.Render(text)
	if selectedRender == "" {
		t.Error("Expected selected style to render text, got empty string")
	}

	// Test prompt style
	promptRender := styles.prompt.Render(text)
	if promptRender == "" {
		t.Error("Expected prompt style to render text, got empty string")
	}

	// Test output style
	outputRender := styles.output.Render(text)
	if outputRender == "" {
		t.Error("Expected output style to render text, got empty string")
	}
}