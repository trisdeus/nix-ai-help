package codegen

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// Intent represents the parsed user intent
type Intent struct {
	Type        string            `json:"type"`         // "service", "desktop", "development", "security", etc.
	Components  []string          `json:"components"`   // ["nginx", "ssl", "firewall"]
	Environment string            `json:"environment"`  // "server", "desktop", "laptop", "vm"
	Complexity  string            `json:"complexity"`   // "basic", "advanced", "enterprise"
	Options     map[string]string `json:"options"`      // Additional parsed options
	HomeManager bool              `json:"home_manager"` // Whether this is for Home Manager
}

// ParseRequest represents a natural language parsing request
type ParseRequest struct {
	Input       string                 `json:"input"`
	Context     *config.NixOSContext   `json:"context,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// ParseResponse represents the result of natural language parsing
type ParseResponse struct {
	Intent      Intent   `json:"intent"`
	Confidence  float64  `json:"confidence"`
	Suggestions []string `json:"suggestions"`
	Warnings    []string `json:"warnings"`
}

// Parser handles natural language to intent parsing
type Parser struct {
	aiProvider ai.Provider
	logger     logger.Logger
	patterns   map[string]*regexp.Regexp
}

// NewParser creates a new natural language parser
func NewParser(aiProvider ai.Provider, logger logger.Logger) *Parser {
	return &Parser{
		aiProvider: aiProvider,
		logger:     logger,
		patterns:   initializePatterns(),
	}
}

// Parse converts natural language input into structured intent
func (p *Parser) Parse(ctx context.Context, request *ParseRequest) (*ParseResponse, error) {
	p.logger.Info(fmt.Sprintf("Parsing natural language request: %s", request.Input))

	// Validate input
	if strings.TrimSpace(request.Input) == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	// First try pattern-based parsing for common requests
	if response := p.tryPatternMatching(request); response != nil {
		p.logger.Info("Successfully parsed using pattern matching")
		return response, nil
	}

	// Fall back to AI-powered parsing for complex requests
	return p.parseWithAI(ctx, request)
}

// tryPatternMatching attempts to parse using predefined patterns
func (p *Parser) tryPatternMatching(request *ParseRequest) *ParseResponse {
	input := strings.ToLower(strings.TrimSpace(request.Input))

	// Desktop environment patterns
	if matched, components := p.matchPattern("desktop", input); matched {
		return &ParseResponse{
			Intent: Intent{
				Type:        "desktop",
				Components:  components,
				Environment: "desktop",
				Complexity:  "basic",
				Options:     make(map[string]string),
			},
			Confidence:  0.9,
			Suggestions: []string{"Consider adding a display manager", "You might want to enable audio support"},
		}
	}

	// Web server patterns
	if matched, components := p.matchPattern("webserver", input); matched {
		return &ParseResponse{
			Intent: Intent{
				Type:        "service",
				Components:  components,
				Environment: "server",
				Complexity:  p.determineComplexity(input),
				Options:     p.extractWebServerOptions(input),
			},
			Confidence:  0.85,
			Suggestions: []string{"Consider enabling firewall", "SSL certificates recommended for production"},
		}
	}

	// Development environment patterns
	if matched, components := p.matchPattern("development", input); matched {
		return &ParseResponse{
			Intent: Intent{
				Type:        "development",
				Components:  components,
				Environment: "development",
				Complexity:  "basic",
				Options:     make(map[string]string),
			},
			Confidence:  0.8,
			Suggestions: []string{"Consider using devenv for project-specific environments"},
		}
	}

	return nil
}

// parseWithAI uses AI to parse complex natural language requests
func (p *Parser) parseWithAI(ctx context.Context, request *ParseRequest) (*ParseResponse, error) {
	prompt := p.buildParsingPrompt(request)

	response, err := p.aiProvider.Query(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI parsing failed: %v", err)
	}

	// Extract JSON from the AI response
	jsonStr := p.extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("could not extract valid JSON from AI response")
	}

	var parseResponse ParseResponse
	if err := json.Unmarshal([]byte(jsonStr), &parseResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	// Validate and sanitize the response
	p.validateAndSanitize(&parseResponse)

	return &parseResponse, nil
}

// buildParsingPrompt constructs the AI prompt for natural language parsing
func (p *Parser) buildParsingPrompt(request *ParseRequest) string {
	contextInfo := ""
	if request.Context != nil {
		contextInfo = fmt.Sprintf(`
Current System Context:
- Uses Flakes: %t
- Home Manager: %s
- NixOS Version: %s
- System Type: %s
`, request.Context.UsesFlakes, request.Context.HomeManagerType, request.Context.NixOSVersion, request.Context.SystemType)
	}

	return fmt.Sprintf(`You are a NixOS configuration expert. Parse the following natural language request into a structured JSON format.

User Request: "%s"
%s
Analyze the request and return a JSON object with this exact structure:
{
  "intent": {
    "type": "service|desktop|development|security|gaming|media|productivity",
    "components": ["list", "of", "components"],
    "environment": "server|desktop|laptop|vm|container",
    "complexity": "basic|advanced|enterprise",
    "options": {"key": "value"},
    "home_manager": false
  },
  "confidence": 0.0-1.0,
  "suggestions": ["helpful", "suggestions"],
  "warnings": ["potential", "issues"]
}

Guidelines:
- Identify the main intent type (service setup, desktop environment, development tools, etc.)
- Extract specific components mentioned (nginx, gnome, docker, etc.)
- Determine target environment and complexity level
- Set home_manager to true only if explicitly mentioned or if it's clearly user-level configuration
- Provide confidence score based on clarity of the request
- Include helpful suggestions for additional components or considerations
- Add warnings for potential security or compatibility issues

Examples:
- "Set up nginx with SSL for my blog" → type: "service", components: ["nginx", "ssl"], environment: "server"
- "I want a GNOME desktop with development tools" → type: "desktop", components: ["gnome", "development"], environment: "desktop"
- "Configure my home directory with zsh and neovim" → type: "development", home_manager: true

Return only the JSON object, no additional text.`, request.Input, contextInfo)
}

// extractJSON extracts JSON from AI response text
func (p *Parser) extractJSON(text string) string {
	// Try to find JSON block in the response
	jsonRegex := regexp.MustCompile(`\{[\s\S]*\}`)
	match := jsonRegex.FindString(text)
	if match != "" {
		return match
	}

	// If no JSON block found, try to extract from code blocks
	codeBlockRegex := regexp.MustCompile("```(?:json)?\n?(.*?)\n?```")
	matches := codeBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// validateAndSanitize ensures the parsed response is valid
func (p *Parser) validateAndSanitize(response *ParseResponse) {
	// Ensure confidence is in valid range
	if response.Confidence < 0 {
		response.Confidence = 0
	}
	if response.Confidence > 1 {
		response.Confidence = 1
	}

	// Validate intent type
	validTypes := []string{"service", "desktop", "development", "security", "gaming", "media", "productivity", "system"}
	if !contains(validTypes, response.Intent.Type) {
		response.Intent.Type = "system"
		response.Confidence *= 0.7 // Reduce confidence for invalid type
	}

	// Validate environment
	validEnvs := []string{"server", "desktop", "laptop", "vm", "container"}
	if !contains(validEnvs, response.Intent.Environment) {
		response.Intent.Environment = "desktop"
	}

	// Validate complexity
	validComplexity := []string{"basic", "advanced", "enterprise"}
	if !contains(validComplexity, response.Intent.Complexity) {
		response.Intent.Complexity = "basic"
	}

	// Initialize options if nil
	if response.Intent.Options == nil {
		response.Intent.Options = make(map[string]string)
	}

	// Initialize slices if nil
	if response.Intent.Components == nil {
		response.Intent.Components = []string{}
	}
	if response.Suggestions == nil {
		response.Suggestions = []string{}
	}
	if response.Warnings == nil {
		response.Warnings = []string{}
	}
}

// matchPattern checks if input matches a specific pattern category
func (p *Parser) matchPattern(category string, input string) (bool, []string) {
	pattern, exists := p.patterns[category]
	if !exists {
		return false, nil
	}

	if pattern.MatchString(input) {
		// Extract components based on known keywords
		return true, p.extractComponents(category, input)
	}

	return false, nil
}

// extractComponents extracts specific components from input based on category
func (p *Parser) extractComponents(category, input string) []string {
	switch category {
	case "desktop":
		components := []string{}
		if strings.Contains(input, "gnome") || strings.Contains(input, "gdm") {
			components = append(components, "gnome")
		}
		if strings.Contains(input, "kde") || strings.Contains(input, "plasma") {
			components = append(components, "kde")
		}
		if strings.Contains(input, "xfce") {
			components = append(components, "xfce")
		}
		if strings.Contains(input, "i3") || strings.Contains(input, "sway") {
			components = append(components, "i3")
		}
		return components

	case "webserver":
		components := []string{}
		if strings.Contains(input, "nginx") {
			components = append(components, "nginx")
		}
		if strings.Contains(input, "apache") {
			components = append(components, "apache")
		}
		if strings.Contains(input, "ssl") || strings.Contains(input, "https") || strings.Contains(input, "tls") {
			components = append(components, "ssl")
		}
		if strings.Contains(input, "cert") || strings.Contains(input, "letsencrypt") {
			components = append(components, "letsencrypt")
		}
		return components

	case "development":
		components := []string{}
		if strings.Contains(input, "docker") {
			components = append(components, "docker")
		}
		if strings.Contains(input, "git") {
			components = append(components, "git")
		}
		if strings.Contains(input, "vscode") || strings.Contains(input, "code") {
			components = append(components, "vscode")
		}
		if strings.Contains(input, "neovim") || strings.Contains(input, "nvim") {
			components = append(components, "neovim")
		}
		if strings.Contains(input, "node") || strings.Contains(input, "javascript") {
			components = append(components, "nodejs")
		}
		if strings.Contains(input, "python") {
			components = append(components, "python")
		}
		if strings.Contains(input, "go") || strings.Contains(input, "golang") {
			components = append(components, "go")
		}
		return components
	}

	return []string{}
}

// determineComplexity determines complexity level from input
func (p *Parser) determineComplexity(input string) string {
	if strings.Contains(input, "enterprise") || strings.Contains(input, "production") ||
		strings.Contains(input, "cluster") || strings.Contains(input, "load balancer") {
		return "enterprise"
	}
	if strings.Contains(input, "advanced") || strings.Contains(input, "custom") ||
		strings.Contains(input, "security") || strings.Contains(input, "hardening") {
		return "advanced"
	}
	return "basic"
}

// extractWebServerOptions extracts specific options for web server configuration
func (p *Parser) extractWebServerOptions(input string) map[string]string {
	options := make(map[string]string)

	if strings.Contains(input, "port") {
		portRegex := regexp.MustCompile(`port\s+(\d+)`)
		if match := portRegex.FindStringSubmatch(input); len(match) > 1 {
			options["port"] = match[1]
		}
	}

	if strings.Contains(input, "domain") {
		domainRegex := regexp.MustCompile(`domain\s+([a-zA-Z0-9.-]+)`)
		if match := domainRegex.FindStringSubmatch(input); len(match) > 1 {
			options["domain"] = match[1]
		}
	}

	if strings.Contains(input, "root") || strings.Contains(input, "document") {
		rootRegex := regexp.MustCompile(`(?:root|document)\s+([/\w.-]+)`)
		if match := rootRegex.FindStringSubmatch(input); len(match) > 1 {
			options["root"] = match[1]
		}
	}

	return options
}

// initializePatterns sets up regex patterns for common configuration requests
func initializePatterns() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"desktop":     regexp.MustCompile(`(?i)(desktop|gui|gnome|kde|plasma|xfce|i3|sway|window manager|display manager)`),
		"webserver":   regexp.MustCompile(`(?i)(web server|nginx|apache|http|website|blog|ssl|https)`),
		"development": regexp.MustCompile(`(?i)(development|dev|programming|code|editor|ide|docker|git|nodejs|python|go)`),
		"gaming":      regexp.MustCompile(`(?i)(gaming|steam|game|graphics|nvidia|amd)`),
		"media":       regexp.MustCompile(`(?i)(media|plex|jellyfin|music|video|streaming)`),
		"security":    regexp.MustCompile(`(?i)(security|firewall|vpn|ssh|hardening|encryption)`),
	}
}

// containsString checks if a slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
