package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// SelfImprovement represents the self-improvement process for the AI
type SelfImprovement struct {
	ID              string            `json:"id"`
	OriginalPrompt  string            `json:"original_prompt"`
	OriginalAnswer  string            `json:"original_answer"`
	Feedback        string            `json:"feedback"`
	ImprovementPlan []ImprovementStep `json:"improvement_plan"`
	AppliedSteps    []ImprovementStep `json:"applied_steps"`
	FinalAnswer     string            `json:"final_answer"`
	Confidence      float64           `json:"confidence"`
	Timestamp       time.Time         `json:"timestamp"`
	Duration        string            `json:"duration"`
}

// ImprovementStep represents a single step in the self-improvement process
type ImprovementStep struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // accuracy, clarity, completeness, relevance, correctness
	Description string      `json:"description"`
	Action      string      `json:"action"`
	Justification string     `json:"justification"`
	Confidence  float64     `json:"confidence"`
	Applied     bool        `json:"applied"`
	Timestamp   time.Time   `json:"timestamp"`
}

// SelfImprover implements self-improvement mechanisms for the AI
type SelfImprover struct {
	provider ai.Provider
	logger   *logger.Logger
}

// NewSelfImprover creates a new self-improver
func NewSelfImprover(provider ai.Provider, log *logger.Logger) *SelfImprover {
	return &SelfImprover{
		provider: provider,
		logger:   log,
	}
}

// Improve improves an AI response based on feedback
func (si *SelfImprover) Improve(ctx context.Context, prompt, answer, feedback string) (*SelfImprovement, error) {
	startTime := time.Now()
	
	si.logger.Info(fmt.Sprintf("Improving response to prompt: %s", prompt[:min(len(prompt), 50)]))
	
	// Step 1: Analyze feedback
	feedbackAnalysis := si.analyzeFeedback(ctx, prompt, answer, feedback)
	
	// Step 2: Generate improvement plan
	improvementPlan := si.generateImprovementPlan(ctx, prompt, answer, feedbackAnalysis)
	
	// Step 3: Apply improvements
	appliedSteps := si.applyImprovements(ctx, improvementPlan)
	
	// Step 4: Generate improved response
	improvedResponse := si.generateImprovedResponse(ctx, prompt, answer, appliedSteps)
	
	// Build self-improvement record
	improvement := &SelfImprovement{
		ID:              fmt.Sprintf("improvement-%d", time.Now().UnixNano()),
		OriginalPrompt:  prompt,
		OriginalAnswer:  answer,
		Feedback:        feedback,
		ImprovementPlan: improvementPlan,
		AppliedSteps:    appliedSteps,
		FinalAnswer:     improvedResponse,
		Confidence:      si.calculateOverallConfidence(appliedSteps),
		Timestamp:       startTime,
		Duration:        time.Since(startTime).String(),
	}
	
	si.logger.Info(fmt.Sprintf("Self-improvement completed in %s with confidence %.2f", 
		improvement.Duration, improvement.Confidence))
	
	return improvement, nil
}

// analyzeFeedback analyzes feedback to identify improvement areas
func (si *SelfImprover) analyzeFeedback(ctx context.Context, prompt, answer, feedback string) string {
	si.logger.Debug("Analyzing feedback")
	
	promptTemplate := fmt.Sprintf(`Analyze the following feedback for an AI response:
Prompt: "%s"
Response: "%s"
Feedback: "%s"

Identify specific areas for improvement in the response based on the feedback.
Consider:
1. Accuracy issues
2. Clarity problems
3. Missing information
4. Irrelevant content
5. Structural issues
6. Technical errors

Return your analysis as JSON:
{
  "analysis": "Detailed analysis of feedback",
  "improvement_areas": [
    {
      "area": "accuracy",
      "issue": "Specific accuracy issue identified",
      "severity": "high|medium|low"
    }
  ],
  "suggestions": [
    "Specific suggestion for improvement"
  ]
}`, prompt, answer, feedback)
	
	response, err := si.provider.GenerateResponse(ctx, promptTemplate)
	if err != nil {
		si.logger.Warn(fmt.Sprintf("Failed to analyze feedback: %v", err))
		return fmt.Sprintf(`{"analysis":"Unable to analyze feedback due to provider error","improvement_areas":[],"suggestions":[]}`)
	}
	
	return response
}

// generateImprovementPlan creates a plan to improve the AI response
func (si *SelfImprover) generateImprovementPlan(ctx context.Context, prompt, answer, feedbackAnalysis string) []ImprovementStep {
	si.logger.Debug("Generating improvement plan")
	
	promptTemplate := fmt.Sprintf(`Based on this feedback analysis, create an improvement plan for the AI response:
Prompt: "%s"
Response: "%s"
Feedback Analysis: "%s"

Create a detailed improvement plan with specific steps to address the identified issues.
Each step should include:
1. A clear type (accuracy, clarity, completeness, relevance, correctness)
2. A detailed description of what needs to be improved
3. A specific action to take
4. A justification for why this improvement is needed
5. A confidence level (0.0-1.0) in the improvement

Return your plan as JSON:
{
  "steps": [
    {
      "id": "step-1",
      "type": "accuracy",
      "description": "Fix technical inaccuracy in command syntax",
      "action": "Replace 'nix-env -i' with 'nix profile install'",
      "justification": "nix-env is deprecated, newer Nix versions prefer nix profile",
      "confidence": 0.95
    }
  ]
}`, prompt, answer, feedbackAnalysis)
	
	response, err := si.provider.GenerateResponse(ctx, promptTemplate)
	if err != nil {
		si.logger.Warn(fmt.Sprintf("Failed to generate improvement plan: %v", err))
		// Return a default improvement plan
		return []ImprovementStep{
			{
				ID:          fmt.Sprintf("step-%d-default", time.Now().UnixNano()),
				Type:        "generic",
				Description: "Generic improvement step",
				Action:      "Review response for accuracy and clarity",
				Justification: "Default improvement when provider fails",
				Confidence:  0.5,
				Applied:     false,
				Timestamp:   time.Now(),
			},
		}
	}
	
	// Parse the response to extract improvement steps
	var plan struct {
		Steps []ImprovementStep `json:"steps"`
	}
	
	if err := json.Unmarshal([]byte(response), &plan); err != nil {
		si.logger.Warn(fmt.Sprintf("Failed to parse improvement plan: %v", err))
		// Return a default improvement plan
		return []ImprovementStep{
			{
				ID:          fmt.Sprintf("step-%d-default", time.Now().UnixNano()),
				Type:        "generic",
				Description: "Generic improvement step",
				Action:      "Review response for accuracy and clarity",
				Justification: "Default improvement when provider fails",
				Confidence:  0.5,
				Applied:     false,
				Timestamp:   time.Now(),
			},
		}
	}
	
	// Set timestamps for all steps
	for i := range plan.Steps {
		plan.Steps[i].Timestamp = time.Now()
	}
	
	return plan.Steps
}

// applyImprovements applies the improvement steps to enhance the response
func (si *SelfImprover) applyImprovements(ctx context.Context, steps []ImprovementStep) []ImprovementStep {
	si.logger.Debug("Applying improvements")
	
	// Mark all steps as applied (in a real implementation, we would actually apply them)
	appliedSteps := make([]ImprovementStep, len(steps))
	for i, step := range steps {
		appliedSteps[i] = step
		appliedSteps[i].Applied = true
		appliedSteps[i].Timestamp = time.Now()
	}
	
	return appliedSteps
}

// generateImprovedResponse generates an improved response based on applied improvements
func (si *SelfImprover) generateImprovedResponse(ctx context.Context, prompt string, originalAnswer string, appliedSteps []ImprovementStep) string {
	si.logger.Debug("Generating improved response")
	
	// Create a prompt that incorporates the improvements
	var improvements []string
	for _, step := range appliedSteps {
		improvements = append(improvements, fmt.Sprintf("- %s: %s", step.Type, step.Action))
	}
	
	promptTemplate := fmt.Sprintf(`Based on the following improvements, generate an enhanced response to the original prompt:
Original Prompt: "%s"
Original Response: "%s"
Improvements to Apply:
%s

Generate a new response that incorporates these improvements while maintaining the core information.
Make sure the response is accurate, clear, complete, and relevant to the original prompt.

Return only the improved response.`, prompt, originalAnswer, strings.Join(improvements, "\n"))
	
	response, err := si.provider.GenerateResponse(ctx, promptTemplate)
	if err != nil {
		si.logger.Warn(fmt.Sprintf("Failed to generate improved response: %v", err))
		return originalAnswer // Return original if we can't improve it
	}
	
	return response
}

// calculateOverallConfidence calculates the overall confidence from applied steps
func (si *SelfImprover) calculateOverallConfidence(steps []ImprovementStep) float64 {
	if len(steps) == 0 {
		return 0.0
	}
	
	totalConfidence := 0.0
	for _, step := range steps {
		totalConfidence += step.Confidence
	}
	
	return totalConfidence / float64(len(steps))
}

// FormatSelfImprovement formats a self-improvement for display
func (si *SelfImprover) FormatSelfImprovement(improvement *SelfImprovement) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# 🧠 AI Self-Improvement\n\n"))
	output.WriteString(fmt.Sprintf("**Prompt:** %s\n", improvement.OriginalPrompt))
	output.WriteString(fmt.Sprintf("**Original Answer:** %s\n", improvement.OriginalAnswer))
	output.WriteString(fmt.Sprintf("**Feedback:** %s\n", improvement.Feedback))
	output.WriteString(fmt.Sprintf("**Improved Answer:** %s\n", improvement.FinalAnswer))
	output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n", improvement.Confidence))
	output.WriteString(fmt.Sprintf("**Duration:** %s\n\n", improvement.Duration))
	
	output.WriteString("## 📋 Improvement Plan\n\n")
	
	for _, step := range improvement.ImprovementPlan {
		emoji := "🔹"
		switch step.Type {
		case "accuracy":
			emoji = "🎯"
		case "clarity":
			emoji = "👓"
		case "completeness":
			emoji = "✅"
		case "relevance":
			emoji = "📌"
		case "correctness":
			emoji = "🧠"
		}
		
		output.WriteString(fmt.Sprintf("### %s %s Improvement\n", emoji, strings.Title(step.Type)))
		output.WriteString(fmt.Sprintf("**Description:** %s\n", step.Description))
		output.WriteString(fmt.Sprintf("**Action:** %s\n", step.Action))
		output.WriteString(fmt.Sprintf("**Justification:** %s\n", step.Justification))
		output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n", step.Confidence))
		output.WriteString(fmt.Sprintf("**Applied:** %t\n\n", step.Applied))
	}
	
	output.WriteString("## 🎯 Applied Steps\n\n")
	
	for _, step := range improvement.AppliedSteps {
		emoji := "🔹"
		switch step.Type {
		case "accuracy":
			emoji = "🎯"
		case "clarity":
			emoji = "👓"
		case "completeness":
			emoji = "✅"
		case "relevance":
			emoji = "📌"
		case "correctness":
			emoji = "🧠"
		}
		
		output.WriteString(fmt.Sprintf("### %s %s Improvement\n", emoji, strings.Title(step.Type)))
		output.WriteString(fmt.Sprintf("**Description:** %s\n", step.Description))
		output.WriteString(fmt.Sprintf("**Action:** %s\n", step.Action))
		output.WriteString(fmt.Sprintf("**Justification:** %s\n", step.Justification))
		output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n\n", step.Confidence))
	}
	
	return output.String()
}