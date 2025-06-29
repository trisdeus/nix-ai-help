package codegen

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// GenerationRequest represents a configuration generation request
type GenerationRequest struct {
	Intent      Intent                 `json:"intent"`
	Context     *config.NixOSContext   `json:"context,omitempty"`
	Template    string                 `json:"template,omitempty"`
	OutputPath  string                 `json:"output_path,omitempty"`
	Merge       bool                   `json:"merge,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// GenerationResponse represents the result of configuration generation
type GenerationResponse struct {
	Configuration string             `json:"configuration"`
	Metadata      GenerationMetadata `json:"metadata"`
	Warnings      []string           `json:"warnings"`
	Suggestions   []string           `json:"suggestions"`
	Dependencies  []string           `json:"dependencies"`
}

// GenerationMetadata contains metadata about the generated configuration
type GenerationMetadata struct {
	Type        string            `json:"type"`
	Template    string            `json:"template"`
	Components  []string          `json:"components"`
	Complexity  string            `json:"complexity"`
	HomeManager bool              `json:"home_manager"`
	RequiredNix string            `json:"required_nix_version"`
	Channels    []string          `json:"channels"`
	Options     map[string]string `json:"options"`
	GeneratedAt string            `json:"generated_at"`
}

// Generator handles NixOS configuration generation
type Generator struct {
	aiProvider     ai.Provider
	logger         logger.Logger
	templateLoader *TemplateLoader
	validator      *Validator
}

// NewGenerator creates a new configuration generator
func NewGenerator(aiProvider ai.Provider, logger logger.Logger) *Generator {
	return &Generator{
		aiProvider:     aiProvider,
		logger:         logger,
		templateLoader: NewTemplateLoader(logger),
		validator:      NewValidator(logger),
	}
}

// Generate creates a NixOS configuration from the given intent
func (g *Generator) Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	g.logger.Info(fmt.Sprintf("Generating NixOS configuration (type: %s, components: %v)", request.Intent.Type, request.Intent.Components))

	// Try template-based generation first
	if response, err := g.generateFromTemplate(ctx, request); err == nil {
		g.logger.Info("Successfully generated configuration from template")
		return response, nil
	}

	// Fall back to AI-powered generation
	return g.generateWithAI(ctx, request)
}

// generateFromTemplate generates configuration using predefined templates
func (g *Generator) generateFromTemplate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	// Load appropriate template based on intent
	template, err := g.templateLoader.LoadTemplate(request.Intent.Type, request.Intent.Complexity)
	if err != nil {
		return nil, fmt.Errorf("template not found: %v", err)
	}

	// Apply template with intent parameters
	config, err := g.templateLoader.ApplyTemplate(template, request.Intent, request.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to apply template: %v", err)
	}

	// Generate metadata
	metadata := GenerationMetadata{
		Type:        request.Intent.Type,
		Template:    template.Name,
		Components:  request.Intent.Components,
		Complexity:  request.Intent.Complexity,
		HomeManager: request.Intent.HomeManager,
		Options:     request.Intent.Options,
		GeneratedAt: getCurrentTimestamp(),
	}

	// Validate the generated configuration
	if err := g.validator.ValidateBasic(config); err != nil {
		return nil, fmt.Errorf("generated configuration validation failed: %v", err)
	}

	return &GenerationResponse{
		Configuration: config,
		Metadata:      metadata,
		Warnings:      template.Warnings,
		Suggestions:   template.Suggestions,
		Dependencies:  template.Dependencies,
	}, nil
}

// generateWithAI generates configuration using AI when templates are insufficient
func (g *Generator) generateWithAI(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	prompt := g.buildGenerationPrompt(request)

	response, err := g.aiProvider.Query(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %v", err)
	}

	// Extract and clean the configuration
	config := g.extractConfiguration(response)
	if config == "" {
		return nil, fmt.Errorf("could not extract valid configuration from AI response")
	}

	// Validate the generated configuration
	if err := g.validator.ValidateBasic(config); err != nil {
		g.logger.Warn(fmt.Sprintf("Generated configuration has validation issues: %v", err))
		// Continue with warnings rather than failing
	}

	// Generate metadata
	metadata := GenerationMetadata{
		Type:        request.Intent.Type,
		Template:    "ai-generated",
		Components:  request.Intent.Components,
		Complexity:  request.Intent.Complexity,
		HomeManager: request.Intent.HomeManager,
		Options:     request.Intent.Options,
		GeneratedAt: getCurrentTimestamp(),
	}

	// Extract additional information from AI response
	warnings, suggestions := g.extractMetadata(response)

	return &GenerationResponse{
		Configuration: config,
		Metadata:      metadata,
		Warnings:      warnings,
		Suggestions:   suggestions,
		Dependencies:  g.extractDependencies(config),
	}, nil
}

// buildGenerationPrompt constructs the AI prompt for configuration generation
func (g *Generator) buildGenerationPrompt(request *GenerationRequest) string {
	var contextInfo string
	if request.Context != nil {
		contextInfo = fmt.Sprintf(`
Current System Context:
- Uses Flakes: %t
- Has Home Manager: %s
- Home Manager Type: %s
- NixOS Version: %s
- System Type: %s
- Configuration Path: %s
`, request.Context.UsesFlakes,
			formatBool(request.Context.HasHomeManager),
			request.Context.HomeManagerType,
			request.Context.NixOSVersion,
			request.Context.SystemType,
			request.Context.NixOSConfigPath)
	}

	configType := "NixOS system configuration"
	if request.Intent.HomeManager {
		configType = "Home Manager configuration"
	}

	componentsStr := strings.Join(request.Intent.Components, ", ")
	var optionsStr string
	for k, v := range request.Intent.Options {
		optionsStr += fmt.Sprintf("- %s: %s\n", k, v)
	}

	// Build prompt using string concatenation to avoid backtick issues
	prompt := "You are a NixOS configuration expert. Generate a complete " + configType + " based on the following requirements:\n\n"
	prompt += "Intent Type: " + request.Intent.Type + "\n"
	prompt += "Components: " + componentsStr + "\n"
	prompt += "Environment: " + request.Intent.Environment + "\n"
	prompt += "Complexity Level: " + request.Intent.Complexity + "\n"
	prompt += contextInfo + "\n"
	prompt += "Additional Options:\n" + optionsStr + "\n"

	prompt += `Requirements:
1. Generate a complete, working NixOS configuration
2. Include proper imports and module structure
3. Add helpful comments explaining each section
4. Use best practices for security and performance
5. Include error handling and validation where appropriate
6. Follow the existing system setup when possible
7. Ensure compatibility with the detected NixOS version

Configuration Guidelines:
- Use proper Nix syntax and indentation
- Include system packages in environment.systemPackages
- Configure services in services.* options
- Add networking configuration if needed
- Include security hardening for production environments
- Add user configuration if specified
- Use hardware-specific optimizations when applicable

Please provide:
1. The complete configuration code
2. Brief explanation of major sections
3. Any warnings or considerations
4. Suggested next steps or additional configurations

Format your response as:
## Configuration

` + "```nix" + `
# Your complete NixOS configuration here
` + "```" + `

## Explanation
Brief explanation of the configuration sections.

## Warnings
Any important warnings or considerations.

## Suggestions
Additional configurations or improvements to consider.`

	return prompt
}

// extractConfiguration extracts the Nix configuration from AI response
func (g *Generator) extractConfiguration(response string) string {
	// Try to find Nix code block first
	nixRegex := regexp.MustCompile(`(?s)` + "```nix\n?(.*?)\n?```")
	matches := nixRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try generic code block
	codeRegex := regexp.MustCompile(`(?s)` + "```\n?(.*?)\n?```")
	matches = codeRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		code := strings.TrimSpace(matches[1])
		// Verify it looks like Nix code
		if strings.Contains(code, "{") && strings.Contains(code, "}") {
			return code
		}
	}

	// Look for configuration starting patterns
	lines := strings.Split(response, "\n")
	var configLines []string
	inConfig := false

	for _, line := range lines {
		if strings.Contains(line, "{ config, pkgs") || strings.Contains(line, "{ config, lib, pkgs") {
			inConfig = true
		}
		if inConfig {
			configLines = append(configLines, line)
		}
		if inConfig && strings.TrimSpace(line) == "}" && len(configLines) > 5 {
			break
		}
	}

	if len(configLines) > 0 {
		return strings.Join(configLines, "\n")
	}

	return ""
}

// extractMetadata extracts warnings and suggestions from AI response
func (g *Generator) extractMetadata(response string) (warnings []string, suggestions []string) {
	lines := strings.Split(response, "\n")

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(strings.ToLower(line), "warning") {
			currentSection = "warnings"
			continue
		}
		if strings.Contains(strings.ToLower(line), "suggestion") {
			currentSection = "suggestions"
			continue
		}

		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			text := strings.TrimSpace(line[1:])
			if text != "" {
				switch currentSection {
				case "warnings":
					warnings = append(warnings, text)
				case "suggestions":
					suggestions = append(suggestions, text)
				}
			}
		}
	}

	return warnings, suggestions
}

// extractDependencies extracts package dependencies from configuration
func (g *Generator) extractDependencies(config string) []string {
	var deps []string

	// Extract from environment.systemPackages
	packagesRegex := regexp.MustCompile(`(?s)environment\.systemPackages\s*=\s*with\s+pkgs;\s*\[\s*(.*?)\s*\];`)
	matches := packagesRegex.FindStringSubmatch(config)
	if len(matches) > 1 {
		packages := strings.Fields(matches[1])
		for _, pkg := range packages {
			pkg = strings.Trim(pkg, " \t\n,")
			if pkg != "" {
				deps = append(deps, pkg)
			}
		}
	}

	// Extract from services
	servicesRegex := regexp.MustCompile(`services\.(\w+)\.enable\s*=\s*true`)
	serviceMatches := servicesRegex.FindAllStringSubmatch(config, -1)
	for _, match := range serviceMatches {
		if len(match) > 1 {
			deps = append(deps, "service-"+match[1])
		}
	}

	return deps
}

// GenerateFromNaturalLanguage is a convenience method that combines parsing and generation
func (g *Generator) GenerateFromNaturalLanguage(ctx context.Context, input string, context *config.NixOSContext) (*GenerationResponse, error) {
	// Parse the natural language input
	parser := NewParser(g.aiProvider, g.logger)
	parseRequest := &ParseRequest{
		Input:   input,
		Context: context,
	}

	parseResponse, err := parser.Parse(ctx, parseRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %v", err)
	}

	// Generate configuration from parsed intent
	genRequest := &GenerationRequest{
		Intent:  parseResponse.Intent,
		Context: context,
	}

	return g.Generate(ctx, genRequest)
}

// GenerateWithTemplateOverride generates configuration with a specific template
func (g *Generator) GenerateWithTemplateOverride(ctx context.Context, request *GenerationRequest, templateName string) (*GenerationResponse, error) {
	originalTemplate := request.Template
	request.Template = templateName

	response, err := g.generateFromTemplate(ctx, request)
	request.Template = originalTemplate // Restore original

	return response, err
}

// MergeWithExisting merges generated configuration with existing configuration
func (g *Generator) MergeWithExisting(generated, existing string) (string, error) {
	// This is a simplified merge - in practice, you'd want more sophisticated merging
	// that understands Nix syntax and can intelligently combine configurations

	// For now, just append with a comment
	if existing != "" {
		return fmt.Sprintf("%s\n\n# Generated configuration merged below:\n%s", existing, generated), nil
	}

	return generated, nil
}

// formatBool formats a boolean for display
func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// getCurrentTimestamp returns current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
