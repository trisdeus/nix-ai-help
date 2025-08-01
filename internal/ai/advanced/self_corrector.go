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

// Correction represents a correction made by the AI
type Correction struct {
	Original    string `json:"original"`
	Correction  string `json:"correction"`
	Explanation string `json:"explanation"`
	Confidence  float64 `json:"confidence"`
	Timestamp   string `json:"timestamp"`
}

// SelfCorrector implements self-correction mechanisms for AI responses
type SelfCorrector struct {
	provider ai.Provider
	logger   *logger.Logger
}

// NewSelfCorrector creates a new self-corrector
func NewSelfCorrector(provider ai.Provider, log *logger.Logger) *SelfCorrector {
	return &SelfCorrector{
		provider: provider,
		logger:   log,
	}
}

// CorrectResponse analyzes and corrects an AI response
func (sc *SelfCorrector) CorrectResponse(ctx context.Context, originalResponse, task string) ([]Correction, error) {
	// Create a prompt asking the AI to review and correct its own response
	correctionPrompt := fmt.Sprintf(`You are reviewing your previous response to the task:

"%s"

Your previous response was:
---
%s
---

Please carefully review your response and identify any errors, inconsistencies, or areas for improvement.

For each issue you find:
1. Quote the original text that needs correction
2. Provide the corrected version
3. Explain why the correction is needed
4. Rate your confidence in the correction (0.0 to 1.0)

Respond in this JSON format:
{
  "corrections": [
    {
      "original": "text to correct",
      "correction": "corrected text",
      "explanation": "why this correction is needed",
      "confidence": 0.9
    }
  ]
}

If you find no issues, respond with an empty corrections array:
{
  "corrections": []
}`, task, originalResponse)

	response, err := sc.provider.GenerateResponse(ctx, correctionPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query AI for self-correction: %w", err)
	}

	// Parse the response to extract corrections
	corrections, err := sc.parseCorrections(response)
	if err != nil {
		sc.logger.Warn(fmt.Sprintf("Failed to parse corrections: %v", err))
		// Return empty corrections if parsing fails
		return []Correction{}, nil
	}

	// Add timestamps to corrections
	now := time.Now().Format("15:04:05")
	for i := range corrections {
		corrections[i].Timestamp = now
	}

	return corrections, nil
}

// parseCorrections parses the JSON response to extract corrections
func (sc *SelfCorrector) parseCorrections(response string) ([]Correction, error) {
	// Try to parse as JSON first
	type correctionResponse struct {
		Corrections []Correction `json:"corrections"`
	}
	
	var cr correctionResponse
	if err := json.Unmarshal([]byte(response), &cr); err == nil {
		return cr.Corrections, nil
	}
	
	// If JSON parsing fails, try to extract corrections manually
	return sc.extractCorrectionsManually(response)
}

// extractCorrectionsManually tries to extract corrections from plain text
func (sc *SelfCorrector) extractCorrectionsManually(response string) ([]Correction, error) {
	lines := strings.Split(response, "\n")
	var corrections []Correction
	
	inCorrection := false
	var currentCorrection Correction
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Original:") || strings.Contains(line, "original:") {
			if inCorrection && currentCorrection.Original != "" {
				corrections = append(corrections, currentCorrection)
			}
			inCorrection = true
			currentCorrection = Correction{
				Timestamp: time.Now().Format("15:04:05"),
			}
		} else if strings.Contains(line, "Correction:") || strings.Contains(line, "correction:") {
			// Extract correction text
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentCorrection.Correction = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Explanation:") || strings.Contains(line, "explanation:") {
			// Extract explanation
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				currentCorrection.Explanation = strings.TrimSpace(parts[1])
			}
		} else if inCorrection && line != "" {
			// Accumulate original text
			if currentCorrection.Original == "" {
				currentCorrection.Original = line
			} else {
				currentCorrection.Original += " " + line
			}
		}
	}
	
	// Add the last correction if we were collecting one
	if inCorrection && currentCorrection.Original != "" {
		corrections = append(corrections, currentCorrection)
	}
	
	return corrections, nil
}

// ApplyCorrections applies corrections to the original response
func (sc *SelfCorrector) ApplyCorrections(original string, corrections []Correction) string {
	result := original
	
	// Apply corrections (in reverse order to preserve positions)
	for i := len(corrections) - 1; i >= 0; i-- {
		corr := corrections[i]
		if strings.Contains(result, corr.Original) {
			result = strings.Replace(result, corr.Original, corr.Correction, 1)
		}
	}
	
	return result
}

// FormatCorrections formats corrections for display
func (sc *SelfCorrector) FormatCorrections(corrections []Correction) string {
	if len(corrections) == 0 {
		return "✅ No issues found during self-review"
	}
	
	var output strings.Builder
	output.WriteString("## 🔍 Self-Correction Results\n\n")
	
	for i, corr := range corrections {
		output.WriteString(fmt.Sprintf("### Correction %d\n", i+1))
		output.WriteString(fmt.Sprintf("**Original:** %s\n", corr.Original))
		output.WriteString(fmt.Sprintf("**Correction:** %s\n", corr.Correction))
		output.WriteString(fmt.Sprintf("**Explanation:** %s\n", corr.Explanation))
		output.WriteString(fmt.Sprintf("**Confidence:** %.1f%%\n\n", corr.Confidence*100))
	}
	
	return output.String()
}