// Package advanced provides advanced AI features for nixai
package advanced

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ReasoningChainVisualizer provides visual representation of reasoning chains in the TUI
type ReasoningChainVisualizer struct {
	styles ThemeStyles
}

// NewReasoningChainVisualizer creates a new reasoning chain visualizer
func NewReasoningChainVisualizer(styles ThemeStyles) *ReasoningChainVisualizer {
	return &ReasoningChainVisualizer{
		styles: styles,
	}
}

// VisualizeReasoningChain renders a reasoning chain in a visually appealing way for the TUI
func (rcv *ReasoningChainVisualizer) VisualizeReasoningChain(chain *ReasoningChain) string {
	if chain == nil {
		return rcv.styles.error.Render("No reasoning chain to visualize")
	}

	var output strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(headerStyle.Render(fmt.Sprintf("🧠 AI Reasoning Chain for: %s", chain.Task)))
	output.WriteString("\n\n")

	// Chain details
	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	output.WriteString(detailsStyle.Render(fmt.Sprintf("⏱️  Total Time: %s", chain.TotalTime)))
	output.WriteString("\n")
	output.WriteString(detailsStyle.Render(fmt.Sprintf("📊 Confidence: %.1f%%", chain.Confidence*100)))
	output.WriteString("\n")
	output.WriteString(detailsStyle.Render(fmt.Sprintf("📈 Quality Score: %d/10", chain.QualityScore)))
	output.WriteString("\n\n")

	// Steps
	stepHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		Padding(0, 1)

	for i, step := range chain.Steps {
		output.WriteString(stepHeaderStyle.Render(fmt.Sprintf("Step %d: %s", i+1, step.Title)))
		output.WriteString("\n")

		contentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Padding(0, 2)

		output.WriteString(contentStyle.Render(step.Content))
		output.WriteString("\n\n")
	}

	// Final answer
	finalAnswerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8B5CF6")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(finalAnswerStyle.Render("🎯 Final Answer"))
	output.WriteString("\n")

	answerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6")).
		Padding(0, 2)

	output.WriteString(answerStyle.Render(chain.FinalAnswer))
	output.WriteString("\n")

	return output.String()
}

// VisualizeConfidenceScore renders a confidence score in a visually appealing way for the TUI
func (rcv *ReasoningChainVisualizer) VisualizeConfidenceScore(score *ConfidenceScore) string {
	if score == nil {
		return rcv.styles.error.Render("No confidence score to visualize")
	}

	var output strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(headerStyle.Render("📊 AI Confidence Score"))
	output.WriteString("\n\n")

	// Score details
	scoreStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(scoreStyle.Render(fmt.Sprintf("Confidence: %.1f%%", score.Score*100)))
	output.WriteString("\n\n")

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	output.WriteString(explanationStyle.Render(score.Explanation))
	output.WriteString("\n\n")

	// Factors
	if len(score.Factors) > 0 {
		factorsHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true).
			Padding(0, 1)

		output.WriteString(factorsHeaderStyle.Render("⚖️  Evaluation Factors"))
		output.WriteString("\n\n")

		for _, factor := range score.Factors {
			factorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Padding(0, 2)

			output.WriteString(factorStyle.Render(fmt.Sprintf("%s: %.1f%%", factor.Name, factor.Value*100)))
			output.WriteString("\n")

			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				Padding(0, 3)

			output.WriteString(descStyle.Render(factor.Description))
			output.WriteString("\n")

			if factor.Contribution > 0 {
				contStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#8B5CF6")).
					Italic(true).
					Padding(0, 3)

				output.WriteString(contStyle.Render(fmt.Sprintf("Contribution: %.2f%%", factor.Contribution*100)))
				output.WriteString("\n")
			}

			output.WriteString("\n")
		}
	}

	// Quality indicators
	if len(score.QualityIndicators) > 0 {
		indicatorsHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true).
			Padding(0, 1)

		output.WriteString(indicatorsHeaderStyle.Render("✅ Quality Indicators"))
		output.WriteString("\n\n")

		for _, indicator := range score.QualityIndicators {
			indicatorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Padding(0, 2)

			output.WriteString(indicatorStyle.Render(fmt.Sprintf("• %s", indicator)))
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	// Warnings
	if len(score.Warnings) > 0 {
		warningsHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true).
			Padding(0, 1)

		output.WriteString(warningsHeaderStyle.Render("⚠️  Warnings"))
		output.WriteString("\n\n")

		for _, warning := range score.Warnings {
			warningStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Padding(0, 2)

			output.WriteString(warningStyle.Render(fmt.Sprintf("• %s", warning)))
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	return output.String()
}

// VisualizeCorrections renders corrections in a visually appealing way for the TUI
func (rcv *ReasoningChainVisualizer) VisualizeCorrections(corrections []Correction) string {
	if len(corrections) == 0 {
		return rcv.styles.success.Render("✅ No corrections needed")
	}

	var output strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(headerStyle.Render("🔁 Self-Corrections"))
	output.WriteString("\n\n")

	// Corrections
	for i, correction := range corrections {
		correctionHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true).
			Padding(0, 1)

		output.WriteString(correctionHeaderStyle.Render(fmt.Sprintf("Correction %d", i+1)))
		output.WriteString("\n")

		origStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Padding(0, 2)

		output.WriteString(origStyle.Render(fmt.Sprintf("Original: %s", correction.Original)))
		output.WriteString("\n")

		corrStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Padding(0, 2)

		output.WriteString(corrStyle.Render(fmt.Sprintf("Correction: %s", correction.Correction)))
		output.WriteString("\n")

		expStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Padding(0, 2)

		output.WriteString(expStyle.Render(fmt.Sprintf("Explanation: %s", correction.Explanation)))
		output.WriteString("\n")

		confStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Italic(true).
			Padding(0, 2)

		output.WriteString(confStyle.Render(fmt.Sprintf("Confidence: %.1f%%", correction.Confidence*100)))
		output.WriteString("\n\n")
	}

	return output.String()
}

// VisualizeTaskPlan renders a task plan in a visually appealing way for the TUI
func (rcv *ReasoningChainVisualizer) VisualizeTaskPlan(plan *TaskPlan) string {
	if plan == nil {
		return rcv.styles.error.Render("No task plan to visualize")
	}

	var output strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(headerStyle.Render(fmt.Sprintf("📋 Task Plan: %s", plan.Title)))
	output.WriteString("\n\n")

	// Plan details
	detailsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	output.WriteString(detailsStyle.Render(fmt.Sprintf("Description: %s", plan.Description)))
	output.WriteString("\n")
	output.WriteString(detailsStyle.Render(fmt.Sprintf("Status: %s", plan.Status)))
	output.WriteString("\n")
	if plan.StartTime != "" {
		output.WriteString(detailsStyle.Render(fmt.Sprintf("Started: %s", plan.StartTime)))
		output.WriteString("\n")
	}
	if plan.EndTime != "" {
		output.WriteString(detailsStyle.Render(fmt.Sprintf("Completed: %s", plan.EndTime)))
		output.WriteString("\n")
	}
	if plan.EstimatedTotalTime != "" {
		output.WriteString(detailsStyle.Render(fmt.Sprintf("Estimated Time: %s", plan.EstimatedTotalTime)))
		output.WriteString("\n")
	}
	if plan.ActualTotalTime != "" {
		output.WriteString(detailsStyle.Render(fmt.Sprintf("Actual Time: %s", plan.ActualTotalTime)))
		output.WriteString("\n")
	}
	output.WriteString(detailsStyle.Render(fmt.Sprintf("Progress: %.1f%%", plan.Progress*100)))
	output.WriteString("\n\n")

	// Tasks
	taskHeaderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981")).
		Bold(true).
		Padding(0, 1)

	output.WriteString(taskHeaderStyle.Render("Tasks"))
	output.WriteString("\n\n")

	for i, task := range plan.Tasks {
		// Task header with status indicator
		statusEmoji := "⏳"
		switch task.Status {
		case "completed":
			statusEmoji = "✅"
		case "in-progress":
			statusEmoji = "🔄"
		case "failed":
			statusEmoji = "❌"
		}

		taskTitleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true).
			Padding(0, 2)

		output.WriteString(taskTitleStyle.Render(fmt.Sprintf("%s Task %d: %s", statusEmoji, i+1, task.Title)))
		output.WriteString("\n")

		// Task description
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Padding(0, 3)

		output.WriteString(descStyle.Render(task.Description))
		output.WriteString("\n")

		// Command if available
		if task.Command != "" {
			cmdStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3B82F6")).
				Padding(0, 3)

			output.WriteString(cmdStyle.Render(fmt.Sprintf("Command: %s", task.Command)))
			output.WriteString("\n")
		}

		// Prerequisites and dependencies
		if len(task.Prerequisites) > 0 || len(task.DependsOn) > 0 {
			depsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8B5CF6")).
				Padding(0, 3)

			if len(task.Prerequisites) > 0 {
				output.WriteString(depsStyle.Render(fmt.Sprintf("Prerequisites: %s", strings.Join(task.Prerequisites, ", "))))
				output.WriteString("\n")
			}

			if len(task.DependsOn) > 0 {
				output.WriteString(depsStyle.Render(fmt.Sprintf("Depends On: %s", strings.Join(task.DependsOn, ", "))))
				output.WriteString("\n")
			}
		}

		// Timing information
		if task.EstimatedTime != "" || task.ActualTime != "" {
			timeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				Italic(true).
				Padding(0, 3)

			if task.EstimatedTime != "" {
				output.WriteString(timeStyle.Render(fmt.Sprintf("Estimated Time: %s", task.EstimatedTime)))
				output.WriteString("\n")
			}

			if task.ActualTime != "" {
				output.WriteString(timeStyle.Render(fmt.Sprintf("Actual Time: %s", task.ActualTime)))
				output.WriteString("\n")
			}
		}

		// Results if available
		if task.Result != "" {
			resultStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981")).
				Padding(0, 3)

			output.WriteString(resultStyle.Render(fmt.Sprintf("Result: %s", task.Result)))
			output.WriteString("\n")
		}

		// Errors if available
		if task.Error != "" {
			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Padding(0, 3)

			output.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", task.Error)))
			output.WriteString("\n")
		}

		output.WriteString("\n")
	}

	return output.String()
}

// ThemeStyles contains styling information for the visualizer
type ThemeStyles struct {
	header   lipgloss.Style
	error    lipgloss.Style
	success  lipgloss.Style
	warning  lipgloss.Style
	info     lipgloss.Style
	accent   lipgloss.Style
	muted    lipgloss.Style
	selected lipgloss.Style
	prompt   lipgloss.Style
	output   lipgloss.Style
}

// NewThemeStyles creates new theme styles for the visualizer
func NewThemeStyles() ThemeStyles {
	return ThemeStyles{
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Padding(0, 1),
		error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true),
		warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")),
		accent: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8B5CF6")).
			Bold(true),
		muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")),
		selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")),
		prompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true),
		output: lipgloss.NewStyle().
			Padding(0, 2),
	}
}