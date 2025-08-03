# Advanced AI Example

This example demonstrates the advanced AI capabilities implemented in nixai, including:

1. **Chain-of-Thought Reasoning** - Shows the AI's reasoning process
2. **Self-Correction** - Enables the AI to review and correct its own responses
3. **Multi-Step Task Planning** - Breaks down complex tasks into manageable steps
4. **Confidence Scoring** - Provides confidence levels for AI responses

## Running the Example

To run this example:

```bash
cd /home/olafkfreund/Source/NIX/nix-ai-help
go run examples/advanced_ai/main.go
```

## Features Demonstrated

### Chain-of-Thought Reasoning

The AI shows its step-by-step reasoning process, making its decision-making transparent to users. This helps build trust and understanding of how the AI arrives at its conclusions.

### Self-Correction

The AI reviews its own responses for accuracy, clarity, and completeness, automatically correcting any issues it identifies before presenting the final output to the user.

### Multi-Step Task Planning

For complex tasks, the AI breaks them down into smaller, actionable steps with clear prerequisites and dependencies. This makes it easier for users to follow along and implement solutions.

### Confidence Scoring

Each AI response includes a confidence score that helps users understand how reliable the information is. The score is based on multiple factors including technical accuracy, completeness, and relevance.

## Customization

You can modify the example to test different queries by changing the strings passed to `ProcessQuery`. Try asking about different NixOS configurations, package management scenarios, or system administration tasks.