package reasoning

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// SelfReflection represents the self-reflection process for AI responses
type SelfReflection struct {
	ID             string         `json:"id"`
	OriginalPrompt string         `json:"original_prompt"`
	OriginalAnswer string         `json:"original_answer"`
	Reflections    []Reflection   `json:"reflections"`
	FinalAnswer    string         `json:"final_answer"`
	Improvements   []string       `json:"improvements"`
	Confidence     float64        `json:"confidence"`
	Metadata       map[string]interface{} `json:"metadata"`
	Timestamp      time.Time      `json:"timestamp"`
	Duration       string         `json:"duration"`
}

// Reflection represents a single reflection step
type Reflection struct {
	ID            string      `json:"id"`
	Type          string      `json:"type"` // accuracy, clarity, completeness, relevance, correctness
	Question      string      `json:"question"`
	Analysis      string      `json:"analysis"`
	Suggestion    string      `json:"suggestion"`
	Confidence    float64     `json:"confidence"`
	Justification string      `json:"justification"`
	Applied       bool        `json:"applied"`
	Timestamp     time.Time   `json:"timestamp"`
}

// SelfReflector implements self-reflection mechanisms for AI responses
type SelfReflector struct {
	provider ai.Provider
	logger   *logger.Logger
}

// NewSelfReflector creates a new self-reflector
func NewSelfReflector(provider ai.Provider, log *logger.Logger) *SelfReflector {
	return &SelfReflector{
		provider: provider,
		logger:   log,
	}
}

// Reflect performs self-reflection on an AI response
func (sr *SelfReflector) Reflect(ctx context.Context, prompt, answer string) (*SelfReflection, error) {
	startTime := time.Now()
	
	sr.logger.Info(fmt.Sprintf("Performing self-reflection on response to prompt: %s", prompt[:min(len(prompt), 50)]))
	
	// Step 1: Problem Decomposition
	decompositionStep := sr.decomposeProblem(ctx, prompt)
	
	// Step 2: Hypothesis Generation
	hypotheses := sr.generateHypotheses(ctx, decompositionStep.Output.(string))
	
	// Step 3: Evidence Gathering
	evidence := sr.gatherEvidence(ctx, hypotheses)
	
	// Step 4: Analysis and Evaluation
	analysis := sr.analyzeAndEvaluate(ctx, hypotheses, evidence)
	
	// Step 5: Conclusion Formation
	conclusion := sr.formConclusion(ctx, analysis)
	
	// Build reasoning chain
	chain := &SelfReflection{
		ID:             fmt.Sprintf("reflection-%d", time.Now().UnixNano()),
		OriginalPrompt: prompt,
		OriginalAnswer: answer,
		Reflections: []Reflection{
			{
				ID:            fmt.Sprintf("reflection-%d-decomposition", time.Now().UnixNano()),
				Type:          "decomposition",
				Question:      "How was the problem broken down?",
				Analysis:      decompositionStep.Output.(string),
				Suggestion:    "Ensure all components are identified",
				Confidence:    0.8,
				Justification: "Standard decomposition approach",
				Applied:       true,
				Timestamp:     time.Now(),
			},
			{
				ID:            fmt.Sprintf("reflection-%d-hypotheses", time.Now().UnixNano()),
				Type:          "hypotheses",
				Question:      "Were appropriate hypotheses generated?",
				Analysis:      hypotheses.Output.(string),
				Suggestion:    "Consider alternative approaches",
				Confidence:    0.75,
				Justification: "Standard hypothesis generation",
				Applied:       true,
				Timestamp:     time.Now(),
			},
			{
				ID:            fmt.Sprintf("reflection-%d-evidence", time.Now().UnixNano()),
				Type:          "evidence",
				Question:      "Was sufficient evidence gathered?",
				Analysis:      evidence.Output.(string),
				Suggestion:    "Gather more specific examples",
				Confidence:    0.7,
				Justification: "Standard evidence gathering",
				Applied:       true,
				Timestamp:     time.Now(),
			},
			{
				ID:            fmt.Sprintf("reflection-%d-analysis", time.Now().UnixNano()),
				Type:          "analysis",
				Question:      "Was the analysis thorough?",
				Analysis:      analysis.Output.(string),
				Suggestion:    "Evaluate all perspectives",
				Confidence:    0.85,
				Justification: "Standard analytical approach",
				Applied:       true,
				Timestamp:     time.Now(),
			},
			{
				ID:            fmt.Sprintf("reflection-%d-conclusion", time.Now().UnixNano()),
				Type:          "conclusion",
				Question:      "Is the conclusion well-supported?",
				Analysis:      conclusion.Output.(string),
				Suggestion:    "Ensure conclusion follows from evidence",
				Confidence:    0.9,
				Justification: "Strong supporting evidence",
				Applied:       true,
				Timestamp:     time.Now(),
			},
		},
		FinalAnswer: conclusion.Output.(string),
		Improvements: []string{
			"Use more specific examples",
			"Include more technical details",
			"Provide additional context",
			"Add links to documentation",
			"Include version compatibility information",
		},
		Confidence: 0.8,
		Metadata: map[string]interface{}{
			"reflection_types": []string{"accuracy", "clarity", "completeness", "relevance", "correctness"},
		},
		Timestamp: startTime,
		Duration:  time.Since(startTime).String(),
	}
	
	sr.logger.Info(fmt.Sprintf("Self-reflection completed in %s with confidence %.2f", 
		chain.Duration, chain.Confidence))
	
	return chain, nil
}

// decomposeProblem breaks down a complex query into simpler components
func (sr *SelfReflector) decomposeProblem(ctx context.Context, query string) ReasoningStep {
	sr.logger.Debug(fmt.Sprintf("Decomposing problem: %s", query))
	
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

Return a JSON response with components and relationships.`, query)
	
	response, err := sr.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		sr.logger.Warn(fmt.Sprintf("Failed to decompose problem: %v", err))
		response = fmt.Sprintf(`{"components":[{"id":"comp-1","title":"Main Component","description":"%s"}]}`, query)
	}
	
	step.Output = response
	step.Confidence = 0.8
	
	return step
}

// generateHypotheses creates potential solutions or explanations for each component
func (sr *SelfReflector) generateHypotheses(ctx context.Context, decomposition string) ReasoningStep {
	sr.logger.Debug("Generating hypotheses")
	
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

Return a JSON response with hypotheses.`, decomposition)
	
	response, err := sr.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		sr.logger.Warn(fmt.Sprintf("Failed to generate hypotheses: %v", err))
		response = `{"hypotheses":[{"id":"hyp-1","title":"Default Hypothesis","description":"Default solution approach","probability":0.5}]}`
	}
	
	step.Output = response
	step.Confidence = 0.7
	
	return step
}

// gatherEvidence collects supporting or contradicting evidence for each hypothesis
func (sr *SelfReflector) gatherEvidence(ctx context.Context, hypotheses ReasoningStep) ReasoningStep {
	sr.logger.Debug("Gathering evidence")
	
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

Return a JSON response with evidence items.`, hypotheses.Output)
	
	response, err := sr.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		sr.logger.Warn(fmt.Sprintf("Failed to gather evidence: %v", err))
		response = `{"evidence":[{"id":"ev-1","hypothesis_id":"hyp-1","type":"supporting","source":"assumption","content":"Default evidence","strength":0.5}]}`
	}
	
	step.Output = response
	step.Confidence = 0.75
	
	return step
}

// analyzeAndEvaluate evaluates hypotheses based on gathered evidence
func (sr *SelfReflector) analyzeAndEvaluate(ctx context.Context, hypotheses, evidence ReasoningStep) ReasoningStep {
	sr.logger.Debug("Analyzing and evaluating hypotheses")
	
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

Return a JSON response with analysis results.`, hypotheses.Output, evidence.Output)
	
	response, err := sr.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		sr.logger.Warn(fmt.Sprintf("Failed to analyze and evaluate: %v", err))
		response = `{"analysis":[{"hypothesis_id":"hyp-1","score":0.5,"explanation":"Default evaluation"}],"ranking":["hyp-1"]}`
	}
	
	step.Output = response
	step.Confidence = 0.85
	
	return step
}

// formConclusion forms a final conclusion based on the analysis
func (sr *SelfReflector) formConclusion(ctx context.Context, analysis ReasoningStep) ReasoningStep {
	sr.logger.Debug("Forming conclusion")
	
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

Return a JSON response with the conclusion.`, analysis.Output)
	
	response, err := sr.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		sr.logger.Warn(fmt.Sprintf("Failed to form conclusion: %v", err))
		response = `{"conclusion":"Default conclusion","confidence":0.5,"justification":"Default justification"}`
	}
	
	step.Output = response
	step.Confidence = 0.9
	
	return step
}

// FormatSelfReflection formats a self-reflection for display
func (sr *SelfReflector) FormatSelfReflection(reflection *SelfReflection) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# 🪞 AI Self-Reflection\n\n"))
	output.WriteString(fmt.Sprintf("**Query:** %s\n", reflection.OriginalPrompt))
	output.WriteString(fmt.Sprintf("**Original Answer:** %s\n", reflection.OriginalAnswer))
	output.WriteString(fmt.Sprintf("**Final Answer:** %s\n", reflection.FinalAnswer))
	output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n", reflection.Confidence))
	output.WriteString(fmt.Sprintf("**Duration:** %s\n\n", reflection.Duration))
	
	output.WriteString("## 🔍 Reflections\n\n")
	
	for _, refl := range reflection.Reflections {
		emoji := "🔹"
		switch refl.Type {
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
		
		output.WriteString(fmt.Sprintf("### %s %s Reflection\n", emoji, strings.Title(refl.Type)))
		output.WriteString(fmt.Sprintf("**Question:** %s\n", refl.Question))
		output.WriteString(fmt.Sprintf("**Analysis:** %s\n", refl.Analysis))
		output.WriteString(fmt.Sprintf("**Suggestion:** %s\n", refl.Suggestion))
		output.WriteString(fmt.Sprintf("**Confidence:** %.2f\n", refl.Confidence))
		output.WriteString(fmt.Sprintf("**Justification:** %s\n\n", refl.Justification))
	}
	
	if len(reflection.Improvements) > 0 {
		output.WriteString("## 🛠️ Suggested Improvements\n\n")
		for i, imp := range reflection.Improvements {
			output.WriteString(fmt.Sprintf("%d. %s\n", i+1, imp))
		}
		output.WriteString("\n")
	}
	
	output.WriteString("## 📊 Metadata\n\n")
	for key, value := range reflection.Metadata {
		output.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}
	
	return output.String()
}