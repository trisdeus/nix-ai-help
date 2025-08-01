package advanced

import (
	"context"
	"testing"

	"nix-ai-help/internal/ai"
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

// TestChainOfThoughtReasoning tests the chain-of-thought reasoning functionality
func TestChainOfThoughtReasoning(t *testing.T) {
	log := logger.NewLogger()
	mockProv := &mockProvider{
		response: `{
			"steps": [
				{
					"step_number": 1,
					"title": "Identify the Problem",
					"content": "The user wants to configure nginx in NixOS"
				},
				{
					"step_number": 2,
					"title": "Research Solution",
					"content": "Look up nginx configuration in NixOS documentation"
				}
			],
			"final_answer": "Enable nginx service in configuration.nix",
			"confidence": 0.9,
			"quality_score": 8
		}`,
	}
	
	reasoner := NewChainOfThoughtReasoner(mockProv, log)
	
	ctx := context.Background()
	chain, err := reasoner.GenerateReasoningChain(ctx, "How to configure nginx in NixOS?")
	
	if err != nil {
		t.Fatalf("GenerateReasoningChain failed: %v", err)
	}
	
	if chain == nil {
		t.Fatal("Expected reasoning chain, got nil")
	}
	
	if len(chain.Steps) == 0 {
		t.Error("Expected reasoning steps, got none")
	}
	
	if chain.FinalAnswer == "" {
		t.Error("Expected final answer, got empty string")
	}
	
	if chain.Confidence <= 0 {
		t.Errorf("Expected positive confidence, got %f", chain.Confidence)
	}
	
	// Test formatting
	formatted := reasoner.FormatReasoningChain(chain)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}
}

// TestSelfCorrection tests the self-correction functionality
func TestSelfCorrection(t *testing.T) {
	log := logger.NewLogger()
	mockProv := &mockProvider{
		response: `{
			"corrections": [
				{
					"original": "Use nix-env -i to install packages",
					"correction": "Use nix-profile install or nix-env -iA for better pinning",
					"explanation": "nix-env -i is discouraged in favor of more specific package references",
					"confidence": 0.95
				}
			]
		}`,
	}
	
	corrector := NewSelfCorrector(mockProv, log)
	
	ctx := context.Background()
	corrections, err := corrector.CorrectResponse(ctx, 
		"Use nix-env -i to install packages", 
		"How to install packages in NixOS?")
	
	if err != nil {
		t.Fatalf("CorrectResponse failed: %v", err)
	}
	
	if corrections == nil {
		t.Fatal("Expected corrections, got nil")
	}
	
	// Test formatting
	formatted := corrector.FormatCorrections(corrections)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}
}

// TestTaskPlanning tests the multi-step task planning functionality
func TestTaskPlanning(t *testing.T) {
	log := logger.NewLogger()
	mockProv := &mockProvider{
		response: `{
			"id": "plan-123",
			"title": "Set up Development Environment",
			"description": "Complete plan to set up a development environment",
			"tasks": [
				{
					"id": "task-1",
					"title": "Install Language Runtime",
					"description": "Install Python runtime"
				},
				{
					"id": "task-2",
					"title": "Set up Project Directory",
					"description": "Create project directory structure",
					"depends_on": ["task-1"]
				}
			]
		}`,
	}
	
	planner := NewTaskPlanner(mockProv, log)
	
	ctx := context.Background()
	plan, err := planner.CreateTaskPlan(ctx, "Set up a Python development environment")
	
	if err != nil {
		t.Fatalf("CreateTaskPlan failed: %v", err)
	}
	
	if plan == nil {
		t.Fatal("Expected task plan, got nil")
	}
	
	if len(plan.Tasks) == 0 {
		t.Error("Expected tasks in plan, got none")
	}
	
	// Test updating task status
	if len(plan.Tasks) > 0 {
		planner.UpdateTaskStatus(plan, plan.Tasks[0].ID, "completed", "Task completed successfully")
		
		if plan.Tasks[0].Status != "completed" {
			t.Errorf("Expected task status to be completed, got %s", plan.Tasks[0].Status)
		}
	}
	
	// Test formatting
	formatted := planner.FormatTaskPlan(plan)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}
}

// TestConfidenceScoring tests the confidence scoring functionality
func TestConfidenceScoring(t *testing.T) {
	log := logger.NewLogger()
	mockProv := &mockProvider{
		response: `{
			"score": 0.85,
			"explanation": "Response is well-structured with good technical accuracy",
			"factors": [
				{
					"name": "technical_accuracy",
					"description": "How technically accurate is the information?",
					"weight": 0.25,
					"value": 0.9,
					"contribution": 0.225
				}
			],
			"quality_indicators": [
				"Uses specific command examples",
				"References official documentation"
			],
			"warnings": [
				"No mention of potential side effects"
			]
		}`,
	}
	
	scorer := NewConfidenceScorer(mockProv, log)
	
	ctx := context.Background()
	score, err := scorer.CalculateConfidence(ctx, 
		"To enable nginx, add services.nginx.enable = true; to your configuration.nix", 
		"How to enable nginx in NixOS?")
	
	if err != nil {
		t.Fatalf("CalculateConfidence failed: %v", err)
	}
	
	if score == nil {
		t.Fatal("Expected confidence score, got nil")
	}
	
	if score.Score <= 0 {
		t.Errorf("Expected positive score, got %f", score.Score)
	}
	
	// Test formatting
	formatted := scorer.FormatConfidenceScore(score)
	if formatted == "" {
		t.Error("Expected formatted output, got empty string")
	}
}

// TestHeuristicConfidenceScoring tests the heuristic confidence scoring fallback
func TestHeuristicConfidenceScoring(t *testing.T) {
	log := logger.NewLogger()
	mockProv := &mockProvider{
		response: "Invalid JSON response", // Will trigger heuristic scoring
	}
	
	scorer := NewConfidenceScorer(mockProv, log)
	
	ctx := context.Background()
	score, err := scorer.CalculateConfidence(ctx, 
		"To enable nginx, add services.nginx.enable = true; to your configuration.nix", 
		"How to enable nginx in NixOS?")
	
	if err != nil {
		t.Fatalf("CalculateConfidence failed: %v", err)
	}
	
	if score == nil {
		t.Fatal("Expected confidence score, got nil")
	}
	
	if score.Score <= 0 {
		t.Errorf("Expected positive score, got %f", score.Score)
	}
}