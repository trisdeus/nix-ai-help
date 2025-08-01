package advanced

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// AdvancedAIResponse represents a comprehensive AI response with all advanced features
type AdvancedAIResponse struct {
	Task             string             `json:"task"`
	OriginalResponse string             `json:"original_response"`
	ReasoningChain   *ReasoningChain    `json:"reasoning_chain,omitempty"`
	Corrections      []Correction       `json:"corrections,omitempty"`
	TaskPlan         *TaskPlan          `json:"task_plan,omitempty"`
	ConfidenceScore  *ConfidenceScore   `json:"confidence_score,omitempty"`
	Timestamp        string             `json:"timestamp"`
	ProcessingTime   string             `json:"processing_time"`
}

// AdvancedAICoordinator coordinates all advanced AI features
type AdvancedAICoordinator struct {
	provider         ai.Provider
	reasoner         *ChainOfThoughtReasoner
	corrector        *SelfCorrector
	planner          *TaskPlanner
	scorer           *ConfidenceScorer
	logger           *logger.Logger
	enableReasoning  bool
	enableCorrection bool
	enablePlanning   bool
	enableScoring    bool
}

// AdvancedAICoordinatorConfig configures the advanced AI coordinator
type AdvancedAICoordinatorConfig struct {
	EnableReasoning  bool
	EnableCorrection bool
	EnablePlanning   bool
	EnableScoring    bool
}

// NewAdvancedAICoordinator creates a new advanced AI coordinator
func NewAdvancedAICoordinator(provider ai.Provider, log *logger.Logger, config AdvancedAICoordinatorConfig) *AdvancedAICoordinator {
	return &AdvancedAICoordinator{
		provider:         provider,
		reasoner:         NewChainOfThoughtReasoner(provider, log),
		corrector:        NewSelfCorrector(provider, log),
		planner:          NewTaskPlanner(provider, log),
		scorer:           NewConfidenceScorer(provider, log),
		logger:           log,
		enableReasoning:  config.EnableReasoning,
		enableCorrection: config.EnableCorrection,
		enablePlanning:   config.EnablePlanning,
		enableScoring:    config.EnableScoring,
	}
}

// ProcessQuery processes a query with all enabled advanced AI features
func (ac *AdvancedAICoordinator) ProcessQuery(ctx context.Context, task string) (*AdvancedAIResponse, error) {
	startTime := time.Now()
	
	response := &AdvancedAIResponse{
		Task:      task,
		Timestamp: startTime.Format("2006-01-02 15:04:05"),
	}
	
	// Get the initial AI response
	initialResponse, err := ac.provider.GenerateResponse(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial AI response: %w", err)
	}
	
	response.OriginalResponse = initialResponse
	
	// Apply chain-of-thought reasoning if enabled
	if ac.enableReasoning {
		ac.logger.Info("Generating reasoning chain...")
		reasoningChain, err := ac.reasoner.GenerateReasoningChain(ctx, task)
		if err != nil {
			ac.logger.Warn(fmt.Sprintf("Failed to generate reasoning chain: %v", err))
		} else {
			response.ReasoningChain = reasoningChain
		}
	}
	
	// Apply self-correction if enabled
	if ac.enableCorrection {
		ac.logger.Info("Performing self-correction...")
		corrections, err := ac.corrector.CorrectResponse(ctx, initialResponse, task)
		if err != nil {
			ac.logger.Warn(fmt.Sprintf("Failed to perform self-correction: %v", err))
		} else {
			response.Corrections = corrections
			
			// Apply corrections to the response
			if len(corrections) > 0 {
				correctedResponse := ac.corrector.ApplyCorrections(initialResponse, corrections)
				response.OriginalResponse = correctedResponse
			}
		}
	}
	
	// Create task plan if enabled and if the task seems complex
	if ac.enablePlanning && ac.isComplexTask(task) {
		ac.logger.Info("Creating task plan...")
		taskPlan, err := ac.planner.CreateTaskPlan(ctx, task)
		if err != nil {
			ac.logger.Warn(fmt.Sprintf("Failed to create task plan: %v", err))
		} else {
			response.TaskPlan = taskPlan
		}
	}
	
	// Calculate confidence score if enabled
	if ac.enableScoring {
		ac.logger.Info("Calculating confidence score...")
		confidenceScore, err := ac.scorer.CalculateConfidence(ctx, response.OriginalResponse, task)
		if err != nil {
			ac.logger.Warn(fmt.Sprintf("Failed to calculate confidence score: %v", err))
		} else {
			response.ConfidenceScore = confidenceScore
		}
	}
	
	// Record processing time
	response.ProcessingTime = time.Since(startTime).String()
	
	return response, nil
}

// isComplexTask determines if a task is complex enough to warrant planning
func (ac *AdvancedAICoordinator) isComplexTask(task string) bool {
	complexIndicators := []string{
		"setup", "install", "configure", "deploy", "migrate", 
		"multiple", "several", "many", "steps", "process",
		"environment", "development", "production",
	}
	
	taskLower := strings.ToLower(task)
	for _, indicator := range complexIndicators {
		if strings.Contains(taskLower, indicator) {
			return true
		}
	}
	
	// Also consider longer tasks as potentially more complex
	return len(strings.Fields(task)) > 10
}

// FormatResponse formats an advanced AI response for display
func (ac *AdvancedAICoordinator) FormatResponse(response *AdvancedAIResponse) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# 🤖 Advanced AI Response\n\n"))
	output.WriteString(fmt.Sprintf("**Task:** %s  \n", response.Task))
	output.WriteString(fmt.Sprintf("**Processed:** %s  \n", response.Timestamp))
	output.WriteString(fmt.Sprintf("**Processing Time:** %s  \n\n", response.ProcessingTime))
	
	// Original response
	output.WriteString("## 📝 Original Response\n\n")
	output.WriteString(fmt.Sprintf("%s\n\n", response.OriginalResponse))
	
	// Confidence score
	if response.ConfidenceScore != nil {
		output.WriteString(ac.scorer.FormatConfidenceScore(response.ConfidenceScore))
		output.WriteString("\n")
	}
	
	// Reasoning chain
	if response.ReasoningChain != nil {
		output.WriteString(ac.reasoner.FormatReasoningChain(response.ReasoningChain))
		output.WriteString("\n")
	}
	
	// Corrections
	if len(response.Corrections) > 0 {
		output.WriteString(ac.corrector.FormatCorrections(response.Corrections))
		output.WriteString("\n")
	}
	
	// Task plan
	if response.TaskPlan != nil {
		output.WriteString(ac.planner.FormatTaskPlan(response.TaskPlan))
		output.WriteString("\n")
	}
	
	return output.String()
}

// GetNextTask returns the next task to execute from a task plan
func (ac *AdvancedAICoordinator) GetNextTask(plan *TaskPlan) *Task {
	return ac.planner.GetNextPendingTask(plan)
}

// UpdateTaskStatus updates the status of a task in a plan
func (ac *AdvancedAICoordinator) UpdateTaskStatus(plan *TaskPlan, taskID, status, result string) {
	ac.planner.UpdateTaskStatus(plan, taskID, status, result)
}