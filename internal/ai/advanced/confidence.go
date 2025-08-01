package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// ConfidenceScore represents a confidence score for an AI response
type ConfidenceScore struct {
	Score        float64  `json:"score"`
	Explanation  string   `json:"explanation"`
	Factors      []Factor `json:"factors"`
	QualityIndicators []string `json:"quality_indicators"`
	Warnings     []string `json:"warnings"`
}

// Factor represents a factor that influences confidence
type Factor struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Value       float64 `json:"value"`
	Contribution float64 `json:"contribution"`
}

// ConfidenceScorer calculates confidence scores for AI responses
type ConfidenceScorer struct {
	provider ai.Provider
	logger   *logger.Logger
}

// NewConfidenceScorer creates a new confidence scorer
func NewConfidenceScorer(provider ai.Provider, log *logger.Logger) *ConfidenceScorer {
	return &ConfidenceScorer{
		provider: provider,
		logger:   log,
	}
}

// CalculateConfidence calculates a confidence score for an AI response
func (cs *ConfidenceScorer) CalculateConfidence(ctx context.Context, response, task string) (*ConfidenceScore, error) {
	// Create a prompt asking the AI to evaluate its own response
	evaluationPrompt := fmt.Sprintf(`You are evaluating the quality and confidence level of an AI-generated response to a technical question.

Question: "%s"

Response: "%s"

Evaluate this response and provide a confidence score from 0.0 to 1.0, where:
- 0.0 means completely uncertain or incorrect
- 0.5 means moderate confidence
- 1.0 means very high confidence

Consider these factors:
1. Technical accuracy of the information
2. Completeness of the response
3. Relevance to the question asked
4. Clarity and organization
5. Presence of factual errors
6. Use of authoritative sources
7. Consistency with known best practices
8. Specificity vs vagueness

Provide your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Brief explanation of the confidence score",
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
    "References official documentation",
    "Provides step-by-step instructions"
  ],
  "warnings": [
    "No mention of potential side effects",
    "Does not specify NixOS version compatibility"
  ]
}

Be honest and critical in your evaluation. It's better to give a lower score with good justification than a high score for a poor response.`, task, response)

	evalResponse, err := cs.provider.GenerateResponse(ctx, evaluationPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query AI for confidence evaluation: %w", err)
	}

	// Parse the evaluation response
	var score ConfidenceScore
	if err := json.Unmarshal([]byte(evalResponse), &score); err != nil {
		// If JSON parsing fails, calculate a simple heuristic score
		score = cs.calculateHeuristicScore(response, task)
	}

	// Ensure score is within bounds
	if score.Score < 0.0 {
		score.Score = 0.0
	} else if score.Score > 1.0 {
		score.Score = 1.0
	}

	return &score, nil
}

// calculateHeuristicScore calculates a simple heuristic confidence score
func (cs *ConfidenceScorer) calculateHeuristicScore(response, task string) ConfidenceScore {
	score := 0.5 // Start with neutral confidence
	var factors []Factor
	var qualityIndicators []string
	var warnings []string

	responseLower := strings.ToLower(response)
	taskLower := strings.ToLower(task)

	// Factor 1: Length and completeness
	lengthFactor := 0.1
	lengthValue := 0.5
	if len(response) > 500 {
		lengthValue = 0.7
		qualityIndicators = append(qualityIndicators, "Comprehensive response length")
	} else if len(response) < 100 {
		lengthValue = 0.3
		warnings = append(warnings, "Very short response may be incomplete")
	}
	factors = append(factors, Factor{
		Name:        "response_length",
		Description: "Completeness based on response length",
		Weight:      lengthFactor,
		Value:       lengthValue,
		Contribution: lengthFactor * lengthValue,
	})
	score += lengthFactor * (lengthValue - 0.5)

	// Factor 2: Technical terminology
	techTerms := []string{"configuration", "nixos", "nixpkgs", "flake", "home-manager", "rebuild", "package"}
	techTermCount := 0
	for _, term := range techTerms {
		if strings.Contains(responseLower, term) {
			techTermCount++
		}
	}
	techFactor := 0.2
	techValue := float64(techTermCount) / float64(len(techTerms))
	if techValue > 1.0 {
		techValue = 1.0
	}
	if techTermCount > 3 {
		qualityIndicators = append(qualityIndicators, "Uses appropriate technical terminology")
	}
	factors = append(factors, Factor{
		Name:        "technical_terms",
		Description: "Use of relevant technical terminology",
		Weight:      techFactor,
		Value:       techValue,
		Contribution: techFactor * techValue,
	})
	score += techFactor * (techValue - 0.5)

	// Factor 3: Command examples
	commandExamples := []string{"nixos-rebuild", "nix-env", "home-manager", "nix-shell"}
	commandCount := 0
	for _, cmd := range commandExamples {
		if strings.Contains(responseLower, cmd) {
			commandCount++
		}
	}
	cmdFactor := 0.15
	cmdValue := float64(commandCount) / float64(len(commandExamples))
	if cmdValue > 1.0 {
		cmdValue = 1.0
	}
	if commandCount > 0 {
		qualityIndicators = append(qualityIndicators, "Includes relevant command examples")
	}
	factors = append(factors, Factor{
		Name:        "command_examples",
		Description: "Inclusion of relevant command examples",
		Weight:      cmdFactor,
		Value:       cmdValue,
		Contribution: cmdFactor * cmdValue,
	})
	score += cmdFactor * (cmdValue - 0.5)

	// Factor 4: Warnings and disclaimers
	warningPhrases := []string{"may", "might", "could", "possibly", "uncertain", "not sure"}
	warningCount := 0
	for _, phrase := range warningPhrases {
		if strings.Contains(responseLower, phrase) {
			warningCount++
		}
	}
	warningFactor := 0.1
	warningValue := 1.0
	if warningCount > 2 {
		warningValue = 0.7
		warnings = append(warnings, "Response contains uncertainty language")
	} else if warningCount > 0 {
		warningValue = 0.9
		warnings = append(warnings, "Response expresses some uncertainty")
	} else {
		qualityIndicators = append(qualityIndicators, "Response is assertive and confident")
	}
	factors = append(factors, Factor{
		Name:        "certainty_language",
		Description: "Presence of uncertain language",
		Weight:      warningFactor,
		Value:       warningValue,
		Contribution: warningFactor * (warningValue - 0.5),
	})
	score += warningFactor * (warningValue - 0.5)

	// Factor 5: Structure and organization
	structuredElements := []string{"\n1.", "\n2.", "\n3.", "\n#", "first", "second", "finally", "step"}
	structuredCount := 0
	for _, element := range structuredElements {
		if strings.Contains(responseLower, element) {
			structuredCount++
		}
	}
	structureFactor := 0.15
	structureValue := float64(structuredCount) / float64(len(structuredElements))
	if structureValue > 1.0 {
		structureValue = 1.0
	}
	if structuredCount > 2 {
		qualityIndicators = append(qualityIndicators, "Well-structured with clear steps or sections")
	}
	factors = append(factors, Factor{
		Name:        "organization",
		Description: "Organization and structure of the response",
		Weight:      structureFactor,
		Value:       structureValue,
		Contribution: structureFactor * structureValue,
	})
	score += structureFactor * (structureValue - 0.5)

	// Factor 6: Relevance to task
	relevanceFactor := 0.15
	relevanceValue := 0.5
	taskWords := strings.Fields(taskLower)
	relevantWordCount := 0
	for _, word := range taskWords {
		if len(word) > 3 && strings.Contains(responseLower, word) {
			relevantWordCount++
		}
	}
	if len(taskWords) > 0 {
		relevanceValue = float64(relevantWordCount) / float64(len(taskWords))
		if relevanceValue > 1.0 {
			relevanceValue = 1.0
		}
	}
	if relevantWordCount > len(taskWords)/2 {
		qualityIndicators = append(qualityIndicators, "Highly relevant to the question asked")
	} else if relevantWordCount == 0 {
		warnings = append(warnings, "May not be fully relevant to the question")
	}
	factors = append(factors, Factor{
		Name:        "relevance",
		Description: "Relevance to the original question",
		Weight:      relevanceFactor,
		Value:       relevanceValue,
		Contribution: relevanceFactor * relevanceValue,
	})
	score += relevanceFactor * (relevanceValue - 0.5)

	// Ensure score stays within bounds
	if score < 0.0 {
		score = 0.0
	} else if score > 1.0 {
		score = 1.0
	}

	explanation := fmt.Sprintf("Heuristic confidence score based on %d factors", len(factors))

	return ConfidenceScore{
		Score:             score,
		Explanation:       explanation,
		Factors:           factors,
		QualityIndicators: qualityIndicators,
		Warnings:          warnings,
	}
}

// FormatConfidenceScore formats a confidence score for display
func (cs *ConfidenceScorer) FormatConfidenceScore(score *ConfidenceScore) string {
	var output strings.Builder
	
	// Confidence meter visualization
	confidenceBar := cs.createConfidenceBar(score.Score)
	
	output.WriteString(fmt.Sprintf("## 🎯 Confidence Score: %.1f%%\n", score.Score*100))
	output.WriteString(fmt.Sprintf("%s\n\n", confidenceBar))
	
	if score.Explanation != "" {
		output.WriteString(fmt.Sprintf("**Explanation:** %s\n\n", score.Explanation))
	}
	
	if len(score.Factors) > 0 {
		output.WriteString("### 📊 Evaluation Factors\n\n")
		for _, factor := range score.Factors {
			emoji := "🟡"
			if factor.Value > 0.7 {
				emoji = "🟢"
			} else if factor.Value < 0.3 {
				emoji = "🔴"
			}
			
			output.WriteString(fmt.Sprintf("- %s **%s**: %.0f%% (Weight: %.0f%%, Contribution: %.1f%%)\n", 
				emoji, factor.Name, factor.Value*100, factor.Weight*100, factor.Contribution*100))
			output.WriteString(fmt.Sprintf("  %s\n\n", factor.Description))
		}
	}
	
	if len(score.QualityIndicators) > 0 {
		output.WriteString("### ✅ Quality Indicators\n\n")
		for _, indicator := range score.QualityIndicators {
			output.WriteString(fmt.Sprintf("- %s\n", indicator))
		}
		output.WriteString("\n")
	}
	
	if len(score.Warnings) > 0 {
		output.WriteString("### ⚠️ Warnings\n\n")
		for _, warning := range score.Warnings {
			output.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		output.WriteString("\n")
	}
	
	// Recommendation based on confidence score
	output.WriteString("### 💡 Recommendation\n\n")
	if score.Score >= 0.9 {
		output.WriteString("This response has very high confidence. You can likely trust and implement the suggestions directly.\n")
	} else if score.Score >= 0.7 {
		output.WriteString("This response has good confidence. Consider verifying critical commands before executing.\n")
	} else if score.Score >= 0.5 {
		output.WriteString("This response has moderate confidence. Verify information and test in a safe environment before implementing.\n")
	} else {
		output.WriteString("This response has low confidence. Double-check all information and consult official documentation before implementing.\n")
	}
	
	return output.String()
}

// createConfidenceBar creates a visual representation of the confidence score
func (cs *ConfidenceScorer) createConfidenceBar(score float64) string {
	totalBars := 20
	filledBars := int(score * float64(totalBars))
	
	var bar strings.Builder
	bar.WriteString("[")
	
	for i := 0; i < totalBars; i++ {
		if i < filledBars {
			// Color based on confidence level
			if score > 0.8 {
				bar.WriteString("🟩") // Green for high confidence
			} else if score > 0.6 {
				bar.WriteString("🟨") // Yellow for medium confidence
			} else {
				bar.WriteString("🟥") // Red for low confidence
			}
		} else {
			bar.WriteString("⬜") // Empty for unfilled
		}
	}
	
	bar.WriteString("]")
	return bar.String()
}