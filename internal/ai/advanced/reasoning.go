package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// ReasoningStep represents a single step in the AI's reasoning process
type ReasoningStep struct {
	StepNumber int    `json:"step_number"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Timestamp  string `json:"timestamp"`
}

// ReasoningChain represents the complete reasoning process of the AI
type ReasoningChain struct {
	Task          string         `json:"task"`
	Steps         []ReasoningStep `json:"steps"`
	FinalAnswer   string         `json:"final_answer"`
	TotalTime     string         `json:"total_time"`
	Confidence    float64        `json:"confidence"`
	QualityScore  int            `json:"quality_score"`
}

// ChainOfThoughtReasoner implements chain-of-thought reasoning for AI responses
type ChainOfThoughtReasoner struct {
	provider ai.Provider
	logger   *logger.Logger
}

// NewChainOfThoughtReasoner creates a new reasoner
func NewChainOfThoughtReasoner(provider ai.Provider, log *logger.Logger) *ChainOfThoughtReasoner {
	return &ChainOfThoughtReasoner{
		provider: provider,
		logger:   log,
	}
}

// GenerateReasoningChain generates a reasoning chain for a given task
func (r *ChainOfThoughtReasoner) GenerateReasoningChain(ctx context.Context, task string) (*ReasoningChain, error) {
	startTime := time.Now()
	
	// Create a prompt that encourages step-by-step thinking
	basePrompt := fmt.Sprintf(`You are a NixOS expert tasked with solving the following problem:

"%s"

Think through this step-by-step, showing your reasoning process. For each step:
1. Clearly state what you're considering
2. Explain your reasoning
3. Describe any trade-offs or alternatives considered
4. Explain why you chose a particular approach

Format your response as a JSON object with these fields:
- steps: array of objects with step_number, title, and content
- final_answer: the complete solution to the original problem
- confidence: a number from 0.0 to 1.0 indicating your confidence in the solution
- quality_score: an integer from 1 to 10 rating the quality of your reasoning

Be thorough but concise in your explanation.`, task)

	response, err := r.provider.GenerateResponse(ctx, basePrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query AI: %w", err)
	}

	// Parse the response
	var chain ReasoningChain
	if err := json.Unmarshal([]byte(response), &chain); err != nil {
		// If JSON parsing fails, treat as plain text
		chain = r.parsePlainTextResponse(response, task)
	}

	// Add timing information
	chain.Task = task
	chain.TotalTime = time.Since(startTime).String()

	// Validate and set confidence if not provided
	if chain.Confidence == 0 {
		chain.Confidence = r.calculateConfidence(&chain)
	}

	// Calculate quality score if not provided
	if chain.QualityScore == 0 {
		chain.QualityScore = r.calculateQualityScore(&chain)
	}

	return &chain, nil
}

// parsePlainTextResponse converts a plain text response to a reasoning chain
func (r *ChainOfThoughtReasoner) parsePlainTextResponse(response, task string) ReasoningChain {
	lines := strings.Split(response, "\n")
	
	var steps []ReasoningStep
	stepNumber := 1
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			steps = append(steps, ReasoningStep{
				StepNumber: stepNumber,
				Title:      fmt.Sprintf("Step %d", stepNumber),
				Content:    line,
				Timestamp:  time.Now().Format("15:04:05"),
			})
			stepNumber++
		}
	}
	
	return ReasoningChain{
		Task:        task,
		Steps:       steps,
		FinalAnswer: response,
	}
}

// calculateConfidence calculates confidence based on chain quality
func (r *ChainOfThoughtReasoner) calculateConfidence(chain *ReasoningChain) float64 {
	if len(chain.Steps) == 0 {
		return 0.1 // Very low confidence if no steps
	}
	
	// Simple heuristic: more steps with detailed content = higher confidence
	confidence := 0.5 // Base confidence
	
	// Bonus for having multiple steps
	if len(chain.Steps) > 3 {
		confidence += 0.2
	}
	
	// Bonus for having a final answer
	if chain.FinalAnswer != "" {
		confidence += 0.2
	}
	
	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// calculateQualityScore calculates a quality score for the reasoning chain
func (r *ChainOfThoughtReasoner) calculateQualityScore(chain *ReasoningChain) int {
	score := 5 // Base score
	
	// More steps = higher score (up to a point)
	if len(chain.Steps) > 2 && len(chain.Steps) < 10 {
		score += 2
	} else if len(chain.Steps) >= 10 {
		score += 1 // Too many steps might indicate verbosity
	}
	
	// Having a final answer adds points
	if chain.FinalAnswer != "" {
		score += 2
	}
	
	// Having confidence information adds points
	if chain.Confidence > 0 {
		score += 1
	}
	
	// Cap at 10
	if score > 10 {
		score = 10
	}
	
	return score
}

// FormatReasoningChain formats a reasoning chain for display
func (r *ChainOfThoughtReasoner) FormatReasoningChain(chain *ReasoningChain) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("## 🤖 AI Reasoning Process for: %s\n\n", chain.Task))
	output.WriteString(fmt.Sprintf("**Total Time:** %s  \n", chain.TotalTime))
	output.WriteString(fmt.Sprintf("**Confidence:** %.1f%%  \n", chain.Confidence*100))
	output.WriteString(fmt.Sprintf("**Quality Score:** %d/10  \n\n", chain.QualityScore))
	
	for _, step := range chain.Steps {
		output.WriteString(fmt.Sprintf("### %s\n", step.Title))
		output.WriteString(fmt.Sprintf("%s\n\n", step.Content))
	}
	
	if chain.FinalAnswer != "" {
		output.WriteString("### 🎯 Final Answer\n")
		output.WriteString(fmt.Sprintf("%s\n", chain.FinalAnswer))
	}
	
	return output.String()
}