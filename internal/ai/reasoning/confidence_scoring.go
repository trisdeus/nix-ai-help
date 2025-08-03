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

// ConfidenceDimension represents a dimension used to evaluate AI response confidence
type ConfidenceDimension struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`     // Importance weight (0.0 to 1.0)
	Score       float64 `json:"score"`      // Dimension score (0.0 to 1.0)
	Explanation string  `json:"explanation"` // Explanation for the score
}

// DetailedConfidenceScore represents a detailed confidence evaluation
type DetailedConfidenceScore struct {
	ID              string               `json:"id"`
	Query           string               `json:"query"`
	Response        string               `json:"response"`
	OverallScore    float64              `json:"overall_score"`
	Dimensions      []ConfidenceDimension `json:"dimensions"`
	Recommendations []string             `json:"recommendations"`
	Warnings        []string             `json:"warnings"`
	Timestamp       time.Time            `json:"timestamp"`
	Duration        string               `json:"duration"`
}

// ConfidenceScorer evaluates AI responses and assigns confidence scores
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

// EvaluateResponse evaluates an AI response and assigns a detailed confidence score
func (cs *ConfidenceScorer) EvaluateResponse(ctx context.Context, query, response string) (*DetailedConfidenceScore, error) {
	startTime := time.Now()
	
	cs.logger.Info(fmt.Sprintf("Evaluating confidence for response to query: %s", query[:min(len(query), 50)]))
	
	// Evaluate each dimension
	accuracy := cs.evaluateAccuracy(ctx, query, response)
	clarity := cs.evaluateClarity(ctx, query, response)
	completeness := cs.evaluateCompleteness(ctx, query, response)
	relevance := cs.evaluateRelevance(ctx, query, response)
	correctness := cs.evaluateCorrectness(ctx, query, response)
	helpfulness := cs.evaluateHelpfulness(ctx, query, response)
	
	// Calculate overall score
	dimensions := []ConfidenceDimension{
		accuracy,
		clarity,
		completeness,
		relevance,
		correctness,
		helpfulness,
	}
	
	overallScore := cs.calculateOverallScore(dimensions)
	
	// Generate recommendations and warnings
	recommendations := cs.generateRecommendations(dimensions)
	warnings := cs.generateWarnings(dimensions)
	
	// Build detailed confidence score
	score := &DetailedConfidenceScore{
		ID:              fmt.Sprintf("score-%d", time.Now().UnixNano()),
		Query:           query,
		Response:        response,
		OverallScore:    overallScore,
		Dimensions:      dimensions,
		Recommendations: recommendations,
		Warnings:        warnings,
		Timestamp:       startTime,
		Duration:        time.Since(startTime).String(),
	}
	
	cs.logger.Info(fmt.Sprintf("Confidence evaluation completed in %s with score %.2f", 
		score.Duration, score.OverallScore))
	
	return score, nil
}

// evaluateAccuracy evaluates the technical accuracy of an AI response
func (cs *ConfidenceScorer) evaluateAccuracy(ctx context.Context, query, response string) ConfidenceDimension {
	cs.logger.Debug("Evaluating accuracy")
	
	dimension := ConfidenceDimension{
		Name:        "accuracy",
		Description: "Technical accuracy of the information provided",
		Weight:      0.25, // High importance
	}
	
	prompt := fmt.Sprintf(`Evaluate the technical accuracy of the following AI response to the query:
Query: "%s"
Response: "%s"

Consider these aspects:
1. Are the commands and configuration options mentioned correct?
2. Are there any factual errors in the response?
3. Does the response align with current NixOS best practices?
4. Are technical terms used correctly?

Return your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Detailed explanation of the accuracy evaluation"
}`, query, response)
	
	// Debug: Log the prompt
	cs.logger.Debug(fmt.Sprintf("Accuracy evaluation prompt: %s", prompt))
	
	evalResponse, err := cs.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to evaluate accuracy: %v", err))
		dimension.Score = 0.1
		dimension.Explanation = "Unable to evaluate accuracy due to provider error"
		return dimension
	}
	
	// Parse the evaluation response
	var evaluation struct {
		Score       float64 `json:"score"`
		Explanation string  `json:"explanation"`
	}
	
	cs.logger.Debug(fmt.Sprintf("Raw accuracy evaluation response: %s", evalResponse))
	
	if err := json.Unmarshal([]byte(evalResponse), &evaluation); err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to parse accuracy evaluation: %v", err))
		// Return a default evaluation
		evaluation.Score = 0.75
		evaluation.Explanation = "Default evaluation due to parsing error"
	}
	
	dimension.Score = evaluation.Score
	dimension.Explanation = evaluation.Explanation
	
	return dimension
}

// evaluateClarity evaluates the clarity of an AI response
func (cs *ConfidenceScorer) evaluateClarity(ctx context.Context, query, response string) ConfidenceDimension {
	cs.logger.Debug("Evaluating clarity")
	
	dimension := ConfidenceDimension{
		Name:        "clarity",
		Description: "Clarity and readability of the response",
		Weight:      0.15, // Moderate importance
	}
	
	prompt := fmt.Sprintf(`Evaluate the clarity and readability of the following AI response to the query:
Query: "%s"
Response: "%s"

Consider these aspects:
1. Is the language clear and understandable?
2. Are complex concepts explained well?
3. Is the structure logical and easy to follow?
4. Are there any confusing or ambiguous sections?

Return your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Detailed explanation of the clarity evaluation"
}`, query, response)
	
	// Debug: Log the prompt
	cs.logger.Debug(fmt.Sprintf("Clarity evaluation prompt: %s", prompt))
	
	evalResponse, err := cs.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to evaluate clarity: %v", err))
		dimension.Score = 0.1
		dimension.Explanation = "Unable to evaluate clarity due to provider error"
		return dimension
	}
	
	// Parse the evaluation response
	var evaluation struct {
		Score       float64 `json:"score"`
		Explanation string  `json:"explanation"`
	}
	
	cs.logger.Debug(fmt.Sprintf("Raw clarity evaluation response: %s", evalResponse))

	cs.logger.Debug(fmt.Sprintf("Raw clarity evaluation response: %s", evalResponse))
	
	cs.logger.Debug(fmt.Sprintf("Raw clarity evaluation response: %s", evalResponse))
	
	if err := json.Unmarshal([]byte(evalResponse), &evaluation); err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to parse clarity evaluation: %v", err))
		// Return a default evaluation
		evaluation.Score = 0.75
		evaluation.Explanation = "Default evaluation due to parsing error"
	}
	
	dimension.Score = evaluation.Score
	dimension.Explanation = evaluation.Explanation
	
	return dimension
}

// evaluateCompleteness evaluates the completeness of an AI response
func (cs *ConfidenceScorer) evaluateCompleteness(ctx context.Context, query, response string) ConfidenceDimension {
	cs.logger.Debug("Evaluating completeness")
	
	dimension := ConfidenceDimension{
		Name:        "completeness",
		Description: "Completeness and thoroughness of the response",
		Weight:      0.20, // High importance
	}
	
	prompt := fmt.Sprintf(`Evaluate the completeness of the following AI response to the query:
Query: "%s"
Response: "%s"

Consider these aspects:
1. Does the response fully address the question?
2. Are all relevant aspects covered?
3. Are there any important omissions?
4. Is additional information needed for a complete understanding?

Return your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Detailed explanation of the completeness evaluation"
}`, query, response)
	
	// Debug: Log the prompt
	cs.logger.Debug(fmt.Sprintf("Completeness evaluation prompt: %s", prompt))
	
	evalResponse, err := cs.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to evaluate completeness: %v", err))
		dimension.Score = 0.1
		dimension.Explanation = "Unable to evaluate completeness due to provider error"
		return dimension
	}
	
	// Parse the evaluation response
	cs.logger.Debug(fmt.Sprintf("Raw completeness evaluation response: %s", evalResponse))

	var evaluation struct {
		Score       float64 `json:"score"`
		Explanation string  `json:"explanation"`
	}
	
	cs.logger.Debug(fmt.Sprintf("Raw completeness evaluation response: %s", evalResponse))
	
	if err := json.Unmarshal([]byte(evalResponse), &evaluation); err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to parse completeness evaluation: %v", err))
		// Return a default evaluation
		evaluation.Score = 0.75
		evaluation.Explanation = "Default evaluation due to parsing error"
	}
	
	dimension.Score = evaluation.Score
	dimension.Explanation = evaluation.Explanation
	
	return dimension
}

// evaluateRelevance evaluates the relevance of an AI response to the query
func (cs *ConfidenceScorer) evaluateRelevance(ctx context.Context, query, response string) ConfidenceDimension {
	cs.logger.Debug("Evaluating relevance")
	
	dimension := ConfidenceDimension{
		Name:        "relevance",
		Description: "Relevance of the response to the original query",
		Weight:      0.15, // Moderate importance
	}
	
	prompt := fmt.Sprintf(`Evaluate the relevance of the following AI response to the query:
Query: "%s"
Response: "%s"

Consider these aspects:
1. Does the response directly address the question asked?
2. Are there any irrelevant sections or tangents?
3. Is the focus maintained on the main topic?
4. Are examples and analogies appropriately chosen?

Return your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Detailed explanation of the relevance evaluation"
}`, query, response)
	
	// Debug: Log the prompt
	cs.logger.Debug(fmt.Sprintf("Relevance evaluation prompt: %s", prompt))
	
	evalResponse, err := cs.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to evaluate relevance: %v", err))
	cs.logger.Debug(fmt.Sprintf("Raw relevance evaluation response: %s", evalResponse))

		dimension.Score = 0.1
		dimension.Explanation = "Unable to evaluate relevance due to provider error"
		return dimension
	}
	
	// Parse the evaluation response
	var evaluation struct {
		Score       float64 `json:"score"`
		Explanation string  `json:"explanation"`
	}
	
	if err := json.Unmarshal([]byte(evalResponse), &evaluation); err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to parse relevance evaluation: %v", err))
		dimension.Score = 0.5
		dimension.Explanation = evalResponse
		return dimension
	}
	
	dimension.Score = evaluation.Score
	dimension.Explanation = evaluation.Explanation
	
	return dimension
}

// evaluateCorrectness evaluates the logical correctness of an AI response
func (cs *ConfidenceScorer) evaluateCorrectness(ctx context.Context, query, response string) ConfidenceDimension {
	cs.logger.Debug("Evaluating correctness")
	
	dimension := ConfidenceDimension{
		Name:        "correctness",
		Description: "Logical correctness and soundness of the reasoning",
		Weight:      0.15, // Moderate importance
	}
	
	prompt := fmt.Sprintf(`Evaluate the logical correctness of the following AI response to the query:
Query: "%s"
Response: "%s"

Consider these aspects:
1. Are the logical steps sound and valid?
2. Are conclusions properly supported by evidence?
3. Are there any logical fallacies or inconsistencies?
4. Is the reasoning internally consistent?

Return your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Detailed explanation of the correctness evaluation"
}`, query, response)
	
	// Debug: Log the prompt
	cs.logger.Debug(fmt.Sprintf("Correctness evaluation prompt: %s", prompt))
	
	evalResponse, err := cs.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to evaluate correctness: %v", err))
		dimension.Score = 0.1
		dimension.Explanation = "Unable to evaluate correctness due to provider error"
		return dimension
	}
	
	// Parse the evaluation response
	var evaluation struct {
		Score       float64 `json:"score"`
		Explanation string  `json:"explanation"`
	}
	
	if err := json.Unmarshal([]byte(evalResponse), &evaluation); err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to parse correctness evaluation: %v", err))
		dimension.Score = 0.5
		dimension.Explanation = evalResponse
		return dimension
	}
	
	dimension.Score = evaluation.Score
	dimension.Explanation = evaluation.Explanation
	
	return dimension
}

// evaluateHelpfulness evaluates the helpfulness of an AI response
func (cs *ConfidenceScorer) evaluateHelpfulness(ctx context.Context, query, response string) ConfidenceDimension {
	cs.logger.Debug("Evaluating helpfulness")
	
	dimension := ConfidenceDimension{
		Name:        "helpfulness",
		Description: "Overall helpfulness and usefulness of the response",
		Weight:      0.10, // Lower importance
	}
	
	prompt := fmt.Sprintf(`Evaluate the helpfulness of the following AI response to the query:
Query: "%s"
Response: "%s"

Consider these aspects:
1. Does the response actually help the user accomplish their goal?
2. Is the information actionable and practical?
3. Are examples and suggestions concrete and useful?
4. Would a beginner find this response helpful?

Return your evaluation in this JSON format:
{
  "score": 0.85,
  "explanation": "Detailed explanation of the helpfulness evaluation"
}`, query, response)
	
	// Debug: Log the prompt
	cs.logger.Debug(fmt.Sprintf("Helpfulness evaluation prompt: %s", prompt))
	
	evalResponse, err := cs.provider.GenerateResponse(ctx, prompt)
	if err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to evaluate helpfulness: %v", err))
		dimension.Score = 0.1
		dimension.Explanation = "Unable to evaluate helpfulness due to provider error"
		return dimension
	}
	
	// Parse the evaluation response
	var evaluation struct {
		Score       float64 `json:"score"`
		Explanation string  `json:"explanation"`
	}
	
	if err := json.Unmarshal([]byte(evalResponse), &evaluation); err != nil {
		cs.logger.Warn(fmt.Sprintf("Failed to parse helpfulness evaluation: %v", err))
		dimension.Score = 0.5
		dimension.Explanation = evalResponse
		return dimension
	}
	
	dimension.Score = evaluation.Score
	dimension.Explanation = evaluation.Explanation
	
	return dimension
}

// calculateOverallScore calculates the weighted average of all dimension scores
func (cs *ConfidenceScorer) calculateOverallScore(dimensions []ConfidenceDimension) float64 {
	if len(dimensions) == 0 {
		return 0.0
	}
	
	totalWeightedScore := 0.0
	totalWeight := 0.0
	
	for _, dim := range dimensions {
		totalWeightedScore += dim.Score * dim.Weight
		totalWeight += dim.Weight
	}
	
	if totalWeight == 0 {
		return 0.0
	}
	
	return totalWeightedScore / totalWeight
}

// generateRecommendations generates improvement recommendations based on dimension scores
func (cs *ConfidenceScorer) generateRecommendations(dimensions []ConfidenceDimension) []string {
	var recommendations []string
	
	for _, dim := range dimensions {
		// Only suggest improvements for dimensions with low scores
		if dim.Score < 0.7 {
			switch dim.Name {
			case "accuracy":
				recommendations = append(recommendations, "Verify technical accuracy with official documentation")
			case "clarity":
				recommendations = append(recommendations, "Improve clarity by using simpler language or better structure")
			case "completeness":
				recommendations = append(recommendations, "Ensure all relevant aspects are covered in the response")
			case "relevance":
				recommendations = append(recommendations, "Focus more directly on the original query")
			case "correctness":
				recommendations = append(recommendations, "Review logical steps for consistency and soundness")
			case "helpfulness":
				recommendations = append(recommendations, "Make the response more actionable and practical")
			}
		}
	}
	
	// Remove duplicates
	recommendations = cs.removeDuplicates(recommendations)
	
	// Limit to 5 recommendations
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}
	
	return recommendations
}

// generateWarnings generates warnings based on dimension scores
func (cs *ConfidenceScorer) generateWarnings(dimensions []ConfidenceDimension) []string {
	var warnings []string
	
	for _, dim := range dimensions {
		// Only warn for dimensions with very low scores
		if dim.Score < 0.5 {
			switch dim.Name {
			case "accuracy":
				warnings = append(warnings, "Low technical accuracy - verify information with official sources")
			case "clarity":
				warnings = append(warnings, "Poor clarity - consider simplifying language or restructuring")
			case "completeness":
				warnings = append(warnings, "Incomplete response - important information may be missing")
			case "relevance":
				warnings = append(warnings, "Irrelevant content - response may not address the query")
			case "correctness":
				warnings = append(warnings, "Questionable logic - review reasoning for consistency")
			case "helpfulness":
				warnings = append(warnings, "Limited usefulness - response may not help accomplish the goal")
			}
		}
	}
	
	// Remove duplicates
	warnings = cs.removeDuplicates(warnings)
	
	// Limit to 5 warnings
	if len(warnings) > 5 {
		warnings = warnings[:5]
	}
	
	return warnings
}

// removeDuplicates removes duplicate strings from a slice
func (cs *ConfidenceScorer) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// FormatDetailedConfidenceScore formats a detailed confidence score for display
func (cs *ConfidenceScorer) FormatDetailedConfidenceScore(score *DetailedConfidenceScore) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# 🎯 AI Confidence Evaluation\n\n"))
	output.WriteString(fmt.Sprintf("**Query:** %s\n", score.Query))
	output.WriteString(fmt.Sprintf("**Response:** %s\n", score.Response))
	output.WriteString(fmt.Sprintf("**Overall Score:** %.2f\n", score.OverallScore))
	output.WriteString(fmt.Sprintf("**Evaluation Time:** %s\n\n", score.Duration))
	
	output.WriteString("## 📊 Dimension Scores\n\n")
	
	for _, dim := range score.Dimensions {
		emoji := "🟡"
		if dim.Score > 0.8 {
			emoji = "🟢"
		} else if dim.Score < 0.5 {
			emoji = "🔴"
		}
		
		output.WriteString(fmt.Sprintf("### %s %s (Weight: %.0f%%)\n", emoji, strings.Title(dim.Name), dim.Weight*100))
		output.WriteString(fmt.Sprintf("**Score:** %.2f\n", dim.Score))
		output.WriteString(fmt.Sprintf("**Description:** %s\n", dim.Description))
		output.WriteString(fmt.Sprintf("**Explanation:** %s\n\n", dim.Explanation))
	}
	
	if len(score.Recommendations) > 0 {
		output.WriteString("## ✅ Recommendations\n\n")
		for i, rec := range score.Recommendations {
			output.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
		output.WriteString("\n")
	}
	
	if len(score.Warnings) > 0 {
		output.WriteString("## ⚠️ Warnings\n\n")
		for i, warn := range score.Warnings {
			output.WriteString(fmt.Sprintf("%d. %s\n", i+1, warn))
		}
		output.WriteString("\n")
	}
	
	output.WriteString("## 📈 Interpretation\n\n")
	if score.OverallScore >= 0.9 {
		output.WriteString("🟢 **Excellent** - Very high confidence in this response. You can likely trust and implement the suggestions directly.\n")
	} else if score.OverallScore >= 0.7 {
		output.WriteString("🟡 **Good** - Good confidence in this response. Consider verifying critical commands before executing.\n")
	} else if score.OverallScore >= 0.5 {
		output.WriteString("🟠 **Moderate** - Moderate confidence in this response. Verify information and test in a safe environment before implementing.\n")
	} else {
		output.WriteString("🔴 **Low** - Low confidence in this response. Double-check all information and consult official documentation before implementing.\n")
	}
	
	return output.String()
}