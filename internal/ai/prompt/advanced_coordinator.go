package prompt

import (
	"context"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// AdvancedPromptCoordinator coordinates all advanced reasoning techniques
type AdvancedPromptCoordinator struct {
	provider       ai.Provider
	logger         *logger.Logger
	enableAdvanced bool
}

// AdvancedPromptCoordinatorConfig configures the advanced prompt coordinator
type AdvancedPromptCoordinatorConfig struct {
	EnableAdvanced bool
}

// NewAdvancedPromptCoordinator creates a new advanced prompt coordinator
func NewAdvancedPromptCoordinator(provider ai.Provider, log *logger.Logger, config AdvancedPromptCoordinatorConfig) *AdvancedPromptCoordinator {
	return &AdvancedPromptCoordinator{
		provider:       provider,
		logger:         log,
		enableAdvanced: config.EnableAdvanced,
	}
}

// BuildAdvancedPrompt builds an advanced prompt with all reasoning features
func (apc *AdvancedPromptCoordinator) BuildAdvancedPrompt(ctx context.Context, basePrompt string, nixosCtx *config.NixOSContext) (string, error) {
	if !apc.enableAdvanced {
		// Use regular prompt building
		return apc.buildRegularPrompt(basePrompt, nixosCtx)
	}
	
	// Start with base prompt
	advancedPrompt := basePrompt
	
	// Add system context if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		systemContext := apc.buildSystemContext(nixosCtx)
		advancedPrompt += "\n\n" + systemContext
	}
	
	// Add historical context if available
	historicalContext := apc.buildHistoricalContext(ctx)
	if historicalContext != "" {
		advancedPrompt += "\n\n" + historicalContext
	}
	
	// Add user preference context if available
	userPrefContext := apc.buildUserPreferenceContext()
	if userPrefContext != "" {
		advancedPrompt += "\n\n" + userPrefContext
	}
	
	// Add task planning guidance if complex task detected
	if apc.isComplexTask(basePrompt) {
		taskPlanningGuidance := apc.buildTaskPlanningGuidance()
		advancedPrompt += "\n\n" + taskPlanningGuidance
	}
	
	// Add self-correction guidance
	selfCorrectionGuidance := apc.buildSelfCorrectionGuidance()
	advancedPrompt += "\n\n" + selfCorrectionGuidance
	
	// Add confidence scoring guidance
	confidenceScoringGuidance := apc.buildConfidenceScoringGuidance()
	advancedPrompt += "\n\n" + confidenceScoringGuidance
	
	// Add chain-of-thought reasoning guidance
	reasoningGuidance := apc.buildReasoningGuidance()
	advancedPrompt += "\n\n" + reasoningGuidance
	
	// Add plugin integration guidance
	pluginGuidance := apc.buildPluginGuidance()
	advancedPrompt += "\n\n" + pluginGuidance
	
	// Add NixOS-specific best practices
	bestPractices := apc.buildBestPractices()
	advancedPrompt += "\n\n" + bestPractices
	
	// Add safety guidelines
	safetyGuidelines := apc.buildSafetyGuidelines()
	advancedPrompt += "\n\n" + safetyGuidelines
	
	return advancedPrompt, nil
}

// buildRegularPrompt builds a regular prompt without advanced features
func (apc *AdvancedPromptCoordinator) buildRegularPrompt(basePrompt string, nixosCtx *config.NixOSContext) (string, error) {
	regularPrompt := basePrompt
	
	// Add system context if available
	if nixosCtx != nil && nixosCtx.CacheValid {
		systemContext := apc.buildSystemContext(nixosCtx)
		regularPrompt += "\n\n" + systemContext
	}
	
	// Add basic NixOS best practices
	basicBestPractices := apc.buildBasicBestPractices()
	regularPrompt += "\n\n" + basicBestPractices
	
	// Add basic safety guidelines
	basicSafetyGuidelines := apc.buildBasicSafetyGuidelines()
	regularPrompt += "\n\n" + basicSafetyGuidelines
	
	return regularPrompt, nil
}

// buildSystemContext creates system context from NixOS configuration
func (apc *AdvancedPromptCoordinator) buildSystemContext(nixosCtx *config.NixOSContext) string {
	var context strings.Builder
	
	context.WriteString("=== USER'S NIXOS SYSTEM CONTEXT ===\n")
	context.WriteString(fmt.Sprintf("System Type: %s\n", nixosCtx.SystemType))
	
	// Configuration approach
	if nixosCtx.UsesFlakes {
		context.WriteString("✅ USES FLAKES - Always suggest flake-based solutions\n")
		context.WriteString("❌ NEVER suggest nix-channel commands\n")
		if nixosCtx.FlakeFile != "" {
			context.WriteString(fmt.Sprintf("Flake location: %s\n", nixosCtx.FlakeFile))
		}
	} else if nixosCtx.UsesChannels {
		context.WriteString("Uses legacy channels - suggest channel-compatible solutions\n")
		context.WriteString("Prefer nix-channel and nixos-rebuild commands\n")
	} else {
		context.WriteString("Configuration approach unclear - provide both flake and channel options\n")
	}
	
	// Home Manager integration
	if nixosCtx.HasHomeManager {
		switch nixosCtx.HomeManagerType {
		case "standalone":
			context.WriteString("✅ HAS STANDALONE HOME MANAGER\n")
			context.WriteString("Use 'home-manager switch' commands\n")
			if nixosCtx.HomeManagerConfigPath != "" {
				context.WriteString(fmt.Sprintf("Home Manager config: %s\n", nixosCtx.HomeManagerConfigPath))
			}
		case "module":
			context.WriteString("✅ HAS HOME MANAGER AS NIXOS MODULE\n")
			context.WriteString("Use home-manager.users.<username> syntax in configuration.nix\n")
		}
	} else {
		context.WriteString("❌ NO HOME MANAGER - Only suggest system-level configuration\n")
	}
	
	// Version information
	if nixosCtx.NixOSVersion != "" {
		context.WriteString(fmt.Sprintf("NixOS Version: %s\n", nixosCtx.NixOSVersion))
	}
	if nixosCtx.NixVersion != "" {
		context.WriteString(fmt.Sprintf("Nix Version: %s\n", nixosCtx.NixVersion))
	}
	
	// Configuration files
	if len(nixosCtx.ConfigurationFiles) > 0 {
		context.WriteString("Configuration files:\n")
		for _, file := range nixosCtx.ConfigurationFiles {
			context.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}
	
	// Currently enabled services (limit to important ones)
	if len(nixosCtx.EnabledServices) > 0 {
		importantServices := apc.filterImportantServices(nixosCtx.EnabledServices)
		if len(importantServices) > 0 {
			context.WriteString("Currently enabled services: ")
			context.WriteString(strings.Join(importantServices, ", "))
			context.WriteString("\n")
		}
	}
	
	// Detection warnings
	if len(nixosCtx.DetectionErrors) > 0 {
		context.WriteString("⚠️  Detection warnings: ")
		context.WriteString(strings.Join(nixosCtx.DetectionErrors, "; "))
		context.WriteString("\n")
	}
	
	context.WriteString("=== END SYSTEM CONTEXT ===\n")
	
	return context.String()
}

// filterImportantServices filters to show only commonly relevant services
func (apc *AdvancedPromptCoordinator) filterImportantServices(services []string) []string {
	important := []string{
		"openssh", "sshd", "nginx", "apache", "postgresql", "mysql",
		"docker", "containerd", "firewall", "sound", "xserver", "gnome",
		"kde", "plasma", "networkmanager", "bluetooth", "printing",
	}
	
	var filtered []string
	for _, service := range services {
		for _, imp := range important {
			if strings.Contains(strings.ToLower(service), imp) {
				filtered = append(filtered, service)
				break
			}
		}
		// Limit to first 10 important services to avoid overwhelming the prompt
		if len(filtered) >= 10 {
			break
		}
	}
	
	return filtered
}

// buildHistoricalContext creates historical context information
func (apc *AdvancedPromptCoordinator) buildHistoricalContext(ctx context.Context) string {
	// In a real implementation, this would retrieve historical context from storage
	// For now, we'll return a placeholder
	return "" // Return empty string for now
}

// buildUserPreferenceContext creates user preference context information
func (apc *AdvancedPromptCoordinator) buildUserPreferenceContext() string {
	// In a real implementation, this would retrieve user preferences from storage
	// For now, we'll return a placeholder
	return "" // Return empty string for now
}

// isComplexTask determines if a task is complex enough to warrant planning
func (apc *AdvancedPromptCoordinator) isComplexTask(task string) bool {
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

// buildTaskPlanningGuidance creates task planning guidance
func (apc *AdvancedPromptCoordinator) buildTaskPlanningGuidance() string {
	return `=== TASK PLANNING GUIDANCE ===
For complex tasks, break them down into smaller, actionable steps:
1. Identify the main objective and any sub-objectives
2. List all required prerequisites and dependencies
3. Create a step-by-step execution plan with clear instructions
4. Estimate time and resources for each step
5. Include validation steps to confirm each stage is completed
6. Provide rollback strategies in case of failure
7. Include troubleshooting tips for common issues

Remember to:
- Prioritize safety and stability
- Include appropriate error handling
- Consider system requirements for each step
- Ensure all commands are technically accurate
- Reference official documentation when available
- Provide clear examples for each step
=== END TASK PLANNING GUIDANCE ===`
}

// buildSelfCorrectionGuidance creates self-correction guidance
func (apc *AdvancedPromptCoordinator) buildSelfCorrectionGuidance() string {
	return `=== SELF-CORRECTION GUIDANCE ===
Before providing your final answer, review and correct your response:
1. Check for technical accuracy (ensure all commands and configurations are correct)
2. Verify clarity and readability (ensure the response is easy to understand)
3. Confirm completeness (ensure all relevant aspects are covered)
4. Validate relevance (ensure the response directly addresses the question)
5. Assess logical correctness (ensure reasoning steps are sound and valid)
6. Evaluate helpfulness (ensure the response actually helps the user accomplish their goal)

If you identify any issues:
- Correct them immediately
- Explain your corrections in the response
- Provide rationale for the changes
- Ensure the final response is accurate and helpful

Remember to:
- NEVER EVER suggest nix-env commands!
- NEVER recommend manual installation
- NEVER use incorrect flake syntax like 'nixpkgs.nix = {...}'
- NEVER suggest outdated or deprecated options
=== END SELF-CORRECTION GUIDANCE ===`
}

// buildConfidenceScoringGuidance creates confidence scoring guidance
func (apc *AdvancedPromptCoordinator) buildConfidenceScoringGuidance() string {
	return `=== CONFIDENCE SCORING GUIDANCE ===
Provide a confidence score for your response (0.0 to 1.0):
- 0.0: Completely uncertain or incorrect
- 0.5: Moderate confidence
- 1.0: Very high confidence

Score based on these factors:
1. Technical accuracy of the information provided (25% weight)
2. Clarity and readability of the response (15% weight)
3. Completeness and thoroughness of the response (20% weight)
4. Relevance to the original question (15% weight)
5. Logical correctness and soundness of the reasoning (15% weight)
6. Helpfulness and practicality of the response (10% weight)

Also provide:
- A brief explanation of your confidence score
- Quality indicators (things that increase confidence)
- Warnings (things that decrease confidence)
- Recommendations for improvement

Remember to:
- Be honest about your confidence level
- If uncertain, suggest verification steps
- If confident, clearly state that
- Always provide a numeric score
=== END CONFIDENCE SCORING GUIDANCE ===`
}

// buildReasoningGuidance creates chain-of-thought reasoning guidance
func (apc *AdvancedPromptCoordinator) buildReasoningGuidance() string {
	return `=== CHAIN-OF-THOUGHT REASONING GUIDANCE ===
Show your step-by-step reasoning process:
1. Identify the main question or task
2. Analyze the requirements and constraints
3. Gather relevant information from multiple sources
4. Synthesize the information into a coherent response
5. Validate your conclusions against best practices
6. Provide examples and specific recommendations

Format your reasoning as:
{
  "steps": [
    {
      "step_number": 1,
      "title": "Problem Analysis",
      "content": "Analysis of the user's question and requirements",
      "confidence": 0.95
    }
  ],
  "final_answer": "Complete answer to the original question",
  "confidence": 0.85,
  "quality_score": 8
}

Remember to:
- Be transparent about your reasoning process
- Show intermediate steps
- Include rationales for decisions
- Provide examples where helpful
- Ensure steps are logically ordered
=== END REASONING GUIDANCE ===`
}

// buildPluginGuidance creates plugin integration guidance
func (apc *AdvancedPromptCoordinator) buildPluginGuidance() string {
	return `=== PLUGIN INTEGRATION GUIDANCE ===
Consider using plugins when appropriate:
1. System information and health monitoring
2. Package monitoring and update management
3. Hardware detection and optimization
4. Performance analysis and tuning
5. Security auditing and hardening
6. Community resource integration
7. Learning module delivery
8. Workflow automation and orchestration

To use plugins:
- Reference them by name in your response
- Include plugin commands in examples
- Explain plugin capabilities when relevant
- Suggest plugin installation if not available

Remember to:
- List available plugins in your response
- Include plugin usage examples
- Mention plugin-specific features
- Provide installation instructions if needed
=== END PLUGIN GUIDANCE ===`
}

// buildBestPractices creates NixOS-specific best practices
func (apc *AdvancedPromptCoordinator) buildBestPractices() string {
	return `=== NIXOS BEST PRACTICES ===
Follow these NixOS-specific best practices:
1. ALWAYS USE configuration.nix for system packages
2. ALWAYS USE services.* options for services
3. ALWAYS use correct flake syntax: inputs.nixpkgs.url = "github:..." and outputs = { self, nixpkgs }: {...}
4. ALWAYS verify package names and option paths with official documentation
5. ALWAYS end with 'sudo nixos-rebuild switch' for configuration changes
6. ALWAYS use examples from real-world GitHub configurations when available
7. NEVER EVER suggest nix-env commands!
8. NEVER recommend manual installation
9. NEVER use incorrect flake syntax like 'nixpkgs.nix = {...}'
10. NEVER suggest outdated or deprecated options
=== END BEST PRACTICES ===`
}

// buildSafetyGuidelines creates safety guidelines for AI responses
func (apc *AdvancedPromptCoordinator) buildSafetyGuidelines() string {
	return `=== SAFETY GUIDELINES ===
Ensure maximum safety and reliability in your responses:
1. ALWAYS VERIFY syntax and commands before suggesting them
2. NEVER suggest destructive operations without warnings
3. ALWAYS include appropriate caveats and disclaimers
4. NEVER recommend running commands with root privileges without proper justification
5. ALWAYS suggest testing in safe environments first
6. NEVER suggest commands that could brick or corrupt the system
7. ALWAYS warn about potential risks and side effects
8. ALWAYS provide rollback or recovery procedures
9. NEVER suggest operations that violate system integrity
10. ALWAYS recommend backups before making significant changes

Remember to:
- Prioritize safety over convenience
- Include clear warnings about risky operations
- Suggest safe alternatives when possible
- Provide detailed rollback instructions
- Explain the consequences of each action
=== END SAFETY GUIDELINES ===`
}

// buildBasicBestPractices creates basic NixOS best practices
func (apc *AdvancedPromptCoordinator) buildBasicBestPractices() string {
	return `=== NIXOS BEST PRACTICES (SIMPLIFIED) ===
Follow these essential NixOS best practices:
1. ALWAYS USE configuration.nix for system packages
2. ALWAYS USE services.* options for services
3. NEVER EVER suggest nix-env commands!
4. ALWAYS end with 'sudo nixos-rebuild switch' for configuration changes
5. ALWAYS verify package names with official documentation
=== END BASIC BEST PRACTICES ===`
}

// buildBasicSafetyGuidelines creates basic safety guidelines
func (apc *AdvancedPromptCoordinator) buildBasicSafetyGuidelines() string {
	return `=== SAFETY GUIDELINES (SIMPLIFIED) ===
Ensure maximum safety in your responses:
1. ALWAYS VERIFY syntax and commands before suggesting them
2. NEVER suggest destructive operations without warnings
3. ALWAYS include appropriate caveats and disclaimers
4. NEVER recommend running commands with root privileges without proper justification
5. ALWAYS suggest testing in safe environments first
6. ALWAYS warn about potential risks and side effects
=== END BASIC SAFETY GUIDELINES ===`
}