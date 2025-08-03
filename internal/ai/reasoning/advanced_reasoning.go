package reasoning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// AdvancedReasoner implements sophisticated reasoning techniques
type AdvancedReasoner struct {
	provider ai.Provider
	logger   *logger.Logger
}

// ReasoningStep represents a step in the advanced reasoning process
type ReasoningStep struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // hypothesis, evidence, analysis, conclusion
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      interface{}            `json:"output"`
	Confidence  float64                `json:"confidence"`
	Timestamp   time.Time              `json:"timestamp"`
	SubSteps    []ReasoningStep        `json:"sub_steps,omitempty"`
}

// ReasoningChain represents a complete chain of reasoning
type ReasoningChain struct {
	ID          string         `json:"id"`
	Query       string         `json:"query"`
	Steps       []ReasoningStep `json:"steps"`
	FinalAnswer string         `json:"final_answer"`
	Confidence  float64        `json:"confidence"`
	Timestamp   time.Time      `json:"timestamp"`
	Duration    string         `json:"duration"`
}

// NewAdvancedReasoner creates a new advanced reasoner
func NewAdvancedReasoner(provider ai.Provider, log *logger.Logger) *AdvancedReasoner {
	return &AdvancedReasoner{
		provider: provider,
		logger:   log,
	}
}

// GenerateReasoningChain generates an advanced reasoning chain for a query
func (ar *AdvancedReasoner) GenerateReasoningChain(ctx context.Context, query string) (*ReasoningChain, error) {
	startTime := time.Now()
	
	ar.logger.Info(fmt.Sprintf("Generating advanced reasoning chain for query: %s", query))
	
	// Step 1: Problem Decomposition
	decompositionStep := ar.decomposeProblem(ctx, query)
	
	// Step 2: Hypothesis Generation
	hypotheses := ar.generateHypotheses(ctx, decompositionStep.Output.(string))
	
	// Step 3: Evidence Gathering
	evidence := ar.gatherEvidence(ctx, hypotheses)
	
	// Step 4: Analysis and Evaluation
	analysis := ar.analyzeAndEvaluate(ctx, hypotheses, evidence)
	
	// Step 5: Conclusion Formation
	conclusion := ar.formConclusion(ctx, analysis)
	
	// Build reasoning chain
	chain := &ReasoningChain{
		ID:        fmt.Sprintf("chain-%d", time.Now().UnixNano()),
		Query:     query,
		Timestamp: startTime,
		Duration:  time.Since(startTime).String(),
		Steps: []ReasoningStep{
			decompositionStep,
			hypotheses,
			evidence,
			analysis,
			conclusion,
		},
		FinalAnswer: conclusion.Output.(string),
		Confidence:  conclusion.Confidence,
	}
	
	ar.logger.Info(fmt.Sprintf("Reasoning chain generated in %s with confidence %.2f", 
		chain.Duration, chain.Confidence))
	
	return chain, nil
}

// decomposeProblem breaks down a complex query into simpler components
func (ar *AdvancedReasoner) decomposeProblem(ctx context.Context, query string) ReasoningStep {
	ar.logger.Debug(fmt.Sprintf("Decomposing problem: %s", query))
	
	step := ReasoningStep{
		ID:          fmt.Sprintf("step-%d-decomposition", time.Now().UnixNano()),
		Type:        "decomposition",
		Title:       "Problem Decomposition",
		Description: "Breaking down the complex query into simpler components",
		Input: map[string]interface{}{
			"query": query,
		},
		Timestamp: time.Now(),
	}
	
	prompt := fmt.Sprintf(`Break down the following complex query into simpler components:
Query: "%s"

Return your response as JSON with this structure:
{
  "components": [
    {
      "id": "component-1",
      "title": "Component Title",
      "description": "Component description",
      "dependencies": ["other-component-id"]
    }
  ],
  "relationships": [
    {
      "from": "component-1",
      "to": "component-2",
      "type": "dependency"
    }
  ]
}`, query)
	
	response, err := ar.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		ar.logger.Warn(fmt.Sprintf("Failed to decompose problem: %v", err))
		response = fmt.Sprintf(`{"components":[{"id":"comp-1","title":"Main Component","description":"%s"}]}`, query)
	}
	
	step.Output = response
	step.Confidence = 0.8
	
	return step
}

// generateHypotheses creates potential solutions or explanations for each component
func (ar *AdvancedReasoner) generateHypotheses(ctx context.Context, decomposition string) ReasoningStep {
	ar.logger.Debug("Generating hypotheses")
	
	step := ReasoningStep{
		ID:          fmt.Sprintf("step-%d-hypotheses", time.Now().UnixNano()),
		Type:        "hypothesis_generation",
		Title:       "Hypothesis Generation",
		Description: "Generating potential solutions or explanations",
		Input: map[string]interface{}{
			"decomposition": decomposition,
		},
		Timestamp: time.Now(),
	}
	
	prompt := fmt.Sprintf(`Based on this problem decomposition, generate 3-5 hypotheses for potential solutions:
%s

Return your response as JSON with this structure:
{
  "hypotheses": [
    {
      "id": "hyp-1",
      "title": "Hypothesis Title",
      "description": "Detailed hypothesis description",
      "probability": 0.75,
      "evidence_needed": ["type of evidence needed"]
    }
  ]
}`, decomposition)
	
	response, err := ar.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		ar.logger.Warn(fmt.Sprintf("Failed to generate hypotheses: %v", err))
		response = `{"hypotheses":[{"id":"hyp-1","title":"Default Hypothesis","description":"Default solution approach","probability":0.5}]}`
	}
	
	step.Output = response
	step.Confidence = 0.7
	
	return step
}

// gatherEvidence collects supporting or contradicting evidence for each hypothesis
func (ar *AdvancedReasoner) gatherEvidence(ctx context.Context, hypotheses ReasoningStep) ReasoningStep {
	ar.logger.Debug("Gathering evidence")
	
	step := ReasoningStep{
		ID:          fmt.Sprintf("step-%d-evidence", time.Now().UnixNano()),
		Type:        "evidence_gathering",
		Title:       "Evidence Gathering",
		Description: "Collecting supporting or contradicting evidence",
		Input: map[string]interface{}{
			"hypotheses": hypotheses.Output,
		},
		Timestamp: time.Now(),
	}
	
	prompt := fmt.Sprintf(`For each hypothesis in the following list, gather supporting and contradicting evidence:
%s

Return your response as JSON with this structure:
{
  "evidence": [
    {
      "id": "ev-1",
      "hypothesis_id": "hyp-1",
      "type": "supporting|contradicting|neutral",
      "source": "documentation|observation|experiment",
      "content": "Evidence content",
      "strength": 0.8
    }
  ]
}`, hypotheses.Output)
	
	response, err := ar.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		ar.logger.Warn(fmt.Sprintf("Failed to gather evidence: %v", err))
		response = `{"evidence":[{"id":"ev-1","hypothesis_id":"hyp-1","type":"supporting","source":"assumption","content":"Default evidence","strength":0.5}]}`
	}
	
	step.Output = response
	step.Confidence = 0.75
	
	return step
}

// analyzeAndEvaluate evaluates hypotheses based on gathered evidence
func (ar *AdvancedReasoner) analyzeAndEvaluate(ctx context.Context, hypotheses, evidence ReasoningStep) ReasoningStep {
	ar.logger.Debug("Analyzing and evaluating hypotheses")
	
	step := ReasoningStep{
		ID:          fmt.Sprintf("step-%d-analysis", time.Now().UnixNano()),
		Type:        "analysis_evaluation",
		Title:       "Analysis and Evaluation",
		Description: "Evaluating hypotheses based on gathered evidence",
		Input: map[string]interface{}{
			"hypotheses": hypotheses.Output,
			"evidence":   evidence.Output,
		},
		Timestamp: time.Now(),
	}
	
	prompt := fmt.Sprintf(`Analyze and evaluate the following hypotheses based on the gathered evidence:
Hypotheses:
%s

Evidence:
%s

Return your response as JSON with this structure:
{
  "analysis": [
    {
      "hypothesis_id": "hyp-1",
      "score": 0.75,
      "explanation": "Why this score was assigned",
      "strengths": ["strength 1", "strength 2"],
      "weaknesses": ["weakness 1", "weakness 2"]
    }
  ],
  "ranking": ["hyp-1", "hyp-2"]
}`, hypotheses.Output, evidence.Output)
	
	response, err := ar.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		ar.logger.Warn(fmt.Sprintf("Failed to analyze and evaluate: %v", err))
		response = `{"analysis":[{"hypothesis_id":"hyp-1","score":0.5,"explanation":"Default evaluation"}],"ranking":["hyp-1"]}`
	}
	
	step.Output = response
	step.Confidence = 0.85
	
	return step
}

// formConclusion forms a final conclusion based on the analysis
func (ar *AdvancedReasoner) formConclusion(ctx context.Context, analysis ReasoningStep) ReasoningStep {
	ar.logger.Debug("Forming conclusion")
	
	step := ReasoningStep{
		ID:          fmt.Sprintf("step-%d-conclusion", time.Now().UnixNano()),
		Type:        "conclusion",
		Title:       "Conclusion Formation",
		Description: "Forming a final conclusion based on the analysis",
		Input: map[string]interface{}{
			"analysis": analysis.Output,
		},
		Timestamp: time.Now(),
	}
	
	prompt := fmt.Sprintf(`Based on this analysis, form a final conclusion:
%s

Return your response as JSON with this structure:
{
  "conclusion": "Final conclusion text",
  "confidence": 0.85,
  "justification": "Why this conclusion was reached",
  "next_steps": ["step 1", "step 2"]
}`, analysis.Output)
	
	response, err := ar.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		ar.logger.Warn(fmt.Sprintf("Failed to form conclusion: %v", err))
		response = `{"conclusion":"Default conclusion","confidence":0.5,"justification":"Default justification"}`
	}
	
	step.Output = response
	step.Confidence = 0.9
	
	return step
}

// FormatReasoningChain formats a reasoning chain for display
func (ar *AdvancedReasoner) FormatReasoningChain(chain *ReasoningChain) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# 🤖 Advanced Reasoning Chain\n\n"))
	output.WriteString(fmt.Sprintf("**Query:** %s\n", chain.Query))
	output.WriteString(fmt.Sprintf("**Duration:** %s\n", chain.Duration))
	output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n\n", chain.Confidence))
	
	output.WriteString("## 🧠 Reasoning Steps\n\n")
	
	for i, step := range chain.Steps {
		emoji := "🔹"
		switch step.Type {
		case "decomposition":
			emoji = "🧩"
		case "hypothesis_generation":
			emoji = "💡"
		case "evidence_gathering":
			emoji = "🔍"
		case "analysis_evaluation":
			emoji = "📊"
		case "conclusion":
			emoji = "✅"
		}
		
		output.WriteString(fmt.Sprintf("### %s Step %d: %s\n", emoji, i+1, step.Title))
		output.WriteString(fmt.Sprintf("**Type:** %s\n", step.Type))
		output.WriteString(fmt.Sprintf("**Description:** %s\n", step.Description))
		output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n", step.Confidence))
		output.WriteString(fmt.Sprintf("**Output:** %s\n\n", step.Output))
	}
	
	output.WriteString("## 🎯 Final Answer\n\n")
	output.WriteString(fmt.Sprintf("%s\n", chain.FinalAnswer))
	
	return output.String()
}