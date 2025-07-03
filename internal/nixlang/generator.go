package nixlang

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// NixConfigGenerator provides predictive configuration generation based on intent
type NixConfigGenerator struct {
	analyzer    *NixAnalyzer
	intelligence *NixIntelligenceService
	logger      *logger.Logger
	config      *config.UserConfig
	templates   map[string]*ConfigTemplate
	patterns    map[string]*GenerationPattern
}

// ConfigTemplate represents a configuration template for specific intents
type ConfigTemplate struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Intent       string                 `json:"intent"`
	Category     string                 `json:"category"`
	Template     string                 `json:"template"`
	Variables    []TemplateVariable     `json:"variables"`
	Dependencies []string               `json:"dependencies"`
	Examples     []string               `json:"examples"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TemplateVariable represents a variable in a configuration template
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	Default      interface{} `json:"default,omitempty"`
	Required     bool        `json:"required"`
	Options      []string    `json:"options,omitempty"`
	Validation   string      `json:"validation,omitempty"`
}

// GenerationPattern represents patterns for generating configurations
type GenerationPattern struct {
	Name        string            `json:"name"`
	Intent      string            `json:"intent"`
	Triggers    []string          `json:"triggers"`
	Template    string            `json:"template"`
	Priority    int               `json:"priority"`
	Context     GenerationContext `json:"context"`
	Confidence  float64           `json:"confidence"`
}

// GenerationContext represents the context for configuration generation
type GenerationContext struct {
	SystemType    string            `json:"system_type"`    // "desktop", "server", "laptop", etc.
	Environment   string            `json:"environment"`    // "development", "production", "testing"
	UserLevel     string            `json:"user_level"`     // "beginner", "intermediate", "advanced"
	Requirements  []string          `json:"requirements"`   // Specific requirements
	Constraints   []string          `json:"constraints"`    // Limitations or constraints
	Preferences   map[string]string `json:"preferences"`    // User preferences
}

// GenerationRequest represents a request for configuration generation
type GenerationRequest struct {
	Intent      string                 `json:"intent"`
	Description string                 `json:"description"`
	Context     GenerationContext      `json:"context"`
	Variables   map[string]interface{} `json:"variables"`
	Constraints []string               `json:"constraints"`
	Options     GenerationOptions      `json:"options"`
}

// GenerationOptions represents options for configuration generation
type GenerationOptions struct {
	IncludeComments    bool     `json:"include_comments"`
	OptimizeForSecurity bool    `json:"optimize_for_security"`
	IncludeExamples    bool     `json:"include_examples"`
	OutputFormat       string   `json:"output_format"` // "nix", "json", "yaml"
	Features           []string `json:"features"`      // Additional features to include
}

// GenerationResult represents the result of configuration generation
type GenerationResult struct {
	Configuration   string                 `json:"configuration"`
	Intent          string                 `json:"intent"`
	Template        string                 `json:"template"`
	Variables       map[string]interface{} `json:"variables"`
	Analysis        *AnalysisResult        `json:"analysis,omitempty"`
	Suggestions     []string               `json:"suggestions"`
	Warnings        []string               `json:"warnings"`
	Documentation   []string               `json:"documentation"`
	RelatedTemplates []string              `json:"related_templates"`
	Confidence      float64                `json:"confidence"`
	Generated       time.Time              `json:"generated"`
}

// NewNixConfigGenerator creates a new configuration generator
func NewNixConfigGenerator(cfg *config.UserConfig, log *logger.Logger) *NixConfigGenerator {
	generator := &NixConfigGenerator{
		analyzer:     NewNixAnalyzer(),
		intelligence: NewNixIntelligenceService(cfg, log),
		logger:       log,
		config:       cfg,
		templates:    make(map[string]*ConfigTemplate),
		patterns:     make(map[string]*GenerationPattern),
	}
	
	generator.initializeTemplates()
	generator.initializePatterns()
	
	return generator
}

// GenerateConfiguration generates a NixOS configuration based on intent
func (g *NixConfigGenerator) GenerateConfiguration(ctx context.Context, request GenerationRequest) (*GenerationResult, error) {
	g.logger.Info(fmt.Sprintf("Generating configuration for intent: %s", request.Intent))
	
	// Find the best matching template
	template, confidence := g.findBestTemplate(request)
	if template == nil {
		return nil, fmt.Errorf("no suitable template found for intent: %s", request.Intent)
	}
	
	// Generate configuration from template
	config, err := g.generateFromTemplate(template, request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate configuration: %w", err)
	}
	
	// Analyze the generated configuration
	analysis, err := g.analyzer.AnalyzeExpression(config)
	if err != nil {
		g.logger.Warn(fmt.Sprintf("Failed to analyze generated configuration: %v", err))
	}
	
	// Create result
	result := &GenerationResult{
		Configuration:   config,
		Intent:          request.Intent,
		Template:        template.Name,
		Variables:       request.Variables,
		Analysis:        analysis,
		Suggestions:     g.generateSuggestions(template, request, analysis),
		Warnings:        g.generateWarnings(template, request, analysis),
		Documentation:   g.generateDocumentation(template, request),
		RelatedTemplates: g.findRelatedTemplates(template),
		Confidence:      confidence,
		Generated:       time.Now(),
	}
	
	g.logger.Info(fmt.Sprintf("Generated configuration with confidence: %.2f", confidence))
	return result, nil
}

// AnalyzeIntent analyzes user input to determine configuration intent
func (g *NixConfigGenerator) AnalyzeIntent(ctx context.Context, userInput string) (*IntentAnalysis, error) {
	// Use our Nix analyzer to understand the intent
	result, err := g.analyzer.AnalyzeExpression(userInput)
	if err != nil {
		// If parsing fails, try pattern-based intent detection
		return g.analyzeIntentFromText(userInput), nil
	}
	
	return &result.Intent, nil
}

// GetAvailableTemplates returns all available configuration templates
func (g *NixConfigGenerator) GetAvailableTemplates() map[string]*ConfigTemplate {
	return g.templates
}

// GetTemplatesByCategory returns templates filtered by category
func (g *NixConfigGenerator) GetTemplatesByCategory(category string) []*ConfigTemplate {
	var templates []*ConfigTemplate
	for _, template := range g.templates {
		if template.Category == category {
			templates = append(templates, template)
		}
	}
	return templates
}

// ValidateTemplate validates a configuration template
func (g *NixConfigGenerator) ValidateTemplate(template *ConfigTemplate) []string {
	var issues []string
	
	if template.Name == "" {
		issues = append(issues, "Template name is required")
	}
	
	if template.Template == "" {
		issues = append(issues, "Template content is required")
	}
	
	if template.Intent == "" {
		issues = append(issues, "Template intent is required")
	}
	
	// Validate template syntax
	if _, err := g.analyzer.AnalyzeExpression(template.Template); err != nil {
		issues = append(issues, fmt.Sprintf("Template syntax error: %v", err))
	}
	
	// Validate template variables
	for _, variable := range template.Variables {
		if variable.Name == "" {
			issues = append(issues, "Variable name is required")
		}
		if variable.Type == "" {
			issues = append(issues, fmt.Sprintf("Variable %s missing type", variable.Name))
		}
	}
	
	return issues
}

// Private methods

func (g *NixConfigGenerator) initializeTemplates() {
	templates := []*ConfigTemplate{
		{
			Name:        "web_server_nginx",
			Description: "Nginx web server configuration",
			Intent:      "service_configuration",
			Category:    "web_server",
			Template: `{
  services.nginx = {
    enable = true;
    {{if .EnableSSL}}
    recommendedTlsSettings = true;
    recommendedOptimisation = true;
    recommendedGzipSettings = true;
    {{end}}
    virtualHosts."{{.Domain}}" = {
      {{if .EnableSSL}}
      enableACME = true;
      forceSSL = true;
      {{end}}
      locations."/" = {
        {{if .ProxyPass}}
        proxyPass = "{{.ProxyPass}}";
        {{else}}
        root = "{{.WebRoot}}";
        index = "index.html";
        {{end}}
      };
    };
  };
  {{if .EnableSSL}}
  security.acme = {
    acceptTerms = true;
    defaults.email = "{{.Email}}";
  };
  {{end}}
  networking.firewall.allowedTCPPorts = [ 80 {{if .EnableSSL}}443{{end}} ];
}`,
			Variables: []TemplateVariable{
				{Name: "Domain", Type: "string", Description: "Domain name for the website", Required: true},
				{Name: "EnableSSL", Type: "boolean", Description: "Enable SSL/TLS with automatic certificates", Default: true},
				{Name: "ProxyPass", Type: "string", Description: "Upstream server to proxy to (optional)"},
				{Name: "WebRoot", Type: "string", Description: "Web root directory", Default: "/var/www/html"},
				{Name: "Email", Type: "string", Description: "Email for Let's Encrypt certificates"},
			},
			Dependencies: []string{"nginx", "acme"},
			Examples: []string{
				`Domain: "example.com", EnableSSL: true, Email: "admin@example.com"`,
				`Domain: "app.example.com", ProxyPass: "http://localhost:3000"`,
			},
		},
		{
			Name:        "development_environment",
			Description: "Development environment with common tools",
			Intent:      "development_environment", 
			Category:    "development",
			Template: `{
  environment.systemPackages = with pkgs; [
    git
    {{range .Languages}}
    {{if eq . "python"}}python3 python3Packages.pip{{end}}
    {{if eq . "nodejs"}}nodejs npm{{end}}
    {{if eq . "rust"}}rustc cargo{{end}}
    {{if eq . "go"}}go{{end}}
    {{if eq . "java"}}openjdk17{{end}}
    {{end}}
    {{range .Editors}}
    {{if eq . "vscode"}}vscode{{end}}
    {{if eq . "vim"}}vim{{end}}
    {{if eq . "emacs"}}emacs{{end}}
    {{end}}
    {{range .Tools}}
    {{.}}
    {{end}}
  ];
  
  {{if .EnableDocker}}
  virtualisation.docker.enable = true;
  users.users.{{.Username}}.extraGroups = [ "docker" ];
  {{end}}
  
  {{if .EnableDirenv}}
  programs.direnv.enable = true;
  {{end}}
}`,
			Variables: []TemplateVariable{
				{Name: "Languages", Type: "array", Description: "Programming languages to include", 
				 Options: []string{"python", "nodejs", "rust", "go", "java"}},
				{Name: "Editors", Type: "array", Description: "Code editors to include",
				 Options: []string{"vscode", "vim", "emacs"}},
				{Name: "Tools", Type: "array", Description: "Additional development tools"},
				{Name: "Username", Type: "string", Description: "Username for development", Required: true},
				{Name: "EnableDocker", Type: "boolean", Description: "Enable Docker support", Default: false},
				{Name: "EnableDirenv", Type: "boolean", Description: "Enable direnv for project environments", Default: true},
			},
			Dependencies: []string{"git"},
		},
		{
			Name:        "database_postgresql",
			Description: "PostgreSQL database server",
			Intent:      "service_configuration",
			Category:    "database",
			Template: `{
  services.postgresql = {
    enable = true;
    package = pkgs.postgresql_{{.Version}};
    settings = {
      {{if .MaxConnections}}max_connections = {{.MaxConnections}};{{end}}
      {{if .SharedBuffers}}shared_buffers = "{{.SharedBuffers}}";{{end}}
      {{if .EffectiveCacheSize}}effective_cache_size = "{{.EffectiveCacheSize}}";{{end}}
    };
    authentication = pkgs.lib.mkOverride 10 ''
      local all all trust
      host all all 127.0.0.1/32 trust
      host all all ::1/128 trust
      {{if .AllowRemote}}
      host all all 0.0.0.0/0 md5
      {{end}}
    '';
    {{if .InitialDatabases}}
    initialScript = pkgs.writeText "backend-initScript" ''
      {{range .InitialDatabases}}
      CREATE DATABASE "{{.}}";
      {{end}}
    '';
    {{end}}
  };
  
  {{if .AllowRemote}}
  networking.firewall.allowedTCPPorts = [ 5432 ];
  {{end}}
}`,
			Variables: []TemplateVariable{
				{Name: "Version", Type: "string", Description: "PostgreSQL version", Default: "15", 
				 Options: []string{"13", "14", "15", "16"}},
				{Name: "MaxConnections", Type: "integer", Description: "Maximum number of connections", Default: 100},
				{Name: "SharedBuffers", Type: "string", Description: "Shared buffer size", Default: "128MB"},
				{Name: "EffectiveCacheSize", Type: "string", Description: "Effective cache size", Default: "4GB"},
				{Name: "AllowRemote", Type: "boolean", Description: "Allow remote connections", Default: false},
				{Name: "InitialDatabases", Type: "array", Description: "Databases to create on startup"},
			},
			Dependencies: []string{"postgresql"},
		},
		{
			Name:        "gaming_setup",
			Description: "Gaming configuration with Steam and graphics drivers",
			Intent:      "gaming_configuration",
			Category:    "gaming",
			Template: `{
  # Gaming configuration
  programs.steam = {
    enable = true;
    {{if .EnableRemotePlay}}
    remotePlay.openFirewall = true;
    {{end}}
    {{if .EnableDedicatedServer}}
    dedicatedServer.openFirewall = true;
    {{end}}
  };
  
  {{if eq .GraphicsDriver "nvidia"}}
  services.xserver.videoDrivers = [ "nvidia" ];
  hardware.opengl = {
    enable = true;
    driSupport = true;
    driSupport32Bit = true;
  };
  {{else if eq .GraphicsDriver "amd"}}
  services.xserver.videoDrivers = [ "amdgpu" ];
  hardware.opengl = {
    enable = true;
    driSupport = true;
    driSupport32Bit = true;
  };
  {{end}}
  
  # Audio for gaming
  sound.enable = true;
  hardware.pulseaudio.enable = true;
  hardware.pulseaudio.support32Bit = true;
  
  environment.systemPackages = with pkgs; [
    {{range .GamingTools}}
    {{.}}
    {{end}}
  ];
  
  {{if .EnableGameMode}}
  programs.gamemode.enable = true;
  {{end}}
}`,
			Variables: []TemplateVariable{
				{Name: "GraphicsDriver", Type: "string", Description: "Graphics driver to use",
				 Options: []string{"nvidia", "amd", "intel"}, Default: "nvidia"},
				{Name: "EnableRemotePlay", Type: "boolean", Description: "Enable Steam Remote Play", Default: false},
				{Name: "EnableDedicatedServer", Type: "boolean", Description: "Enable dedicated server support", Default: false},
				{Name: "EnableGameMode", Type: "boolean", Description: "Enable GameMode for performance", Default: true},
				{Name: "GamingTools", Type: "array", Description: "Additional gaming tools",
				 Options: []string{"discord", "obs-studio", "mumble", "teamspeak_client"}},
			},
			Dependencies: []string{"steam"},
		},
		{
			Name:        "security_hardening",
			Description: "Security hardening configuration",
			Intent:      "security_configuration",
			Category:    "security",
			Template: `{
  # Security hardening configuration
  security = {
    sudo = {
      enable = true;
      {{if .RequirePassword}}
      wheelNeedsPassword = true;
      {{else}}
      wheelNeedsPassword = false;
      {{end}}
    };
    
    {{if .EnableAppArmor}}
    apparmor.enable = true;
    {{end}}
    
    {{if .EnableAuditd}}
    auditd.enable = true;
    audit = {
      enable = true;
      rules = [
        "-a always,exit -F arch=b64 -S adjtimex,settimeofday -k time-change"
        "-a always,exit -F arch=b32 -S adjtimex,settimeofday,stime -k time-change"
      ];
    };
    {{end}}
  };
  
  # Firewall configuration
  networking.firewall = {
    enable = true;
    {{if .AllowedTCPPorts}}
    allowedTCPPorts = [ {{range .AllowedTCPPorts}}{{.}} {{end}}];
    {{end}}
    {{if .AllowedUDPPorts}}  
    allowedUDPPorts = [ {{range .AllowedUDPPorts}}{{.}} {{end}}];
    {{end}}
    {{if .EnablePingResponses}}
    allowPing = true;
    {{else}}
    allowPing = false;
    {{end}}
  };
  
  # SSH hardening
  services.openssh = {
    {{if .EnableSSH}}
    enable = true;
    settings = {
      PermitRootLogin = "no";
      PasswordAuthentication = {{.AllowPasswordAuth}};
      PubkeyAuthentication = true;
      {{if .CustomSSHPort}}
      Port = {{.CustomSSHPort}};
      {{end}}
    };
    {{else}}
    enable = false;
    {{end}}
  };
  
  {{if .DisableUnneededServices}}
  # Disable unnecessary services
  services = {
    avahi.enable = false;
    printing.enable = false;
  };
  {{end}}
}`,
			Variables: []TemplateVariable{
				{Name: "RequirePassword", Type: "boolean", Description: "Require password for sudo", Default: true},
				{Name: "EnableAppArmor", Type: "boolean", Description: "Enable AppArmor security module", Default: true},
				{Name: "EnableAuditd", Type: "boolean", Description: "Enable audit daemon", Default: true},
				{Name: "EnableSSH", Type: "boolean", Description: "Enable SSH server", Default: true},
				{Name: "AllowPasswordAuth", Type: "boolean", Description: "Allow SSH password authentication", Default: false},
				{Name: "CustomSSHPort", Type: "integer", Description: "Custom SSH port (optional)"},
				{Name: "AllowedTCPPorts", Type: "array", Description: "Allowed TCP ports"},
				{Name: "AllowedUDPPorts", Type: "array", Description: "Allowed UDP ports"},
				{Name: "EnablePingResponses", Type: "boolean", Description: "Allow ping responses", Default: false},
				{Name: "DisableUnneededServices", Type: "boolean", Description: "Disable unnecessary services", Default: true},
			},
			Dependencies: []string{"openssh"},
		},
	}
	
	for _, template := range templates {
		g.templates[template.Name] = template
	}
}

func (g *NixConfigGenerator) initializePatterns() {
	patterns := []*GenerationPattern{
		{
			Name:     "web_server_intent",
			Intent:   "service_configuration",
			Triggers: []string{"web server", "nginx", "apache", "website", "http", "https"},
			Template: "web_server_nginx",
			Priority: 10,
			Confidence: 0.9,
		},
		{
			Name:     "database_intent",
			Intent:   "service_configuration", 
			Triggers: []string{"database", "postgres", "mysql", "db", "sql"},
			Template: "database_postgresql",
			Priority: 10,
			Confidence: 0.85,
		},
		{
			Name:     "development_intent",
			Intent:   "development_environment",
			Triggers: []string{"development", "coding", "programming", "dev environment", "ide"},
			Template: "development_environment",
			Priority: 8,
			Confidence: 0.8,
		},
		{
			Name:     "gaming_intent",
			Intent:   "gaming_configuration",
			Triggers: []string{"gaming", "steam", "games", "nvidia", "graphics"},
			Template: "gaming_setup",
			Priority: 7,
			Confidence: 0.8,
		},
		{
			Name:     "security_intent",
			Intent:   "security_configuration",
			Triggers: []string{"security", "hardening", "firewall", "ssh", "secure"},
			Template: "security_hardening",
			Priority: 9,
			Confidence: 0.85,
		},
	}
	
	for _, pattern := range patterns {
		g.patterns[pattern.Name] = pattern
	}
}

func (g *NixConfigGenerator) findBestTemplate(request GenerationRequest) (*ConfigTemplate, float64) {
	var bestTemplate *ConfigTemplate
	var bestConfidence float64
	
	// Direct template lookup by intent
	for _, template := range g.templates {
		if template.Intent == request.Intent {
			confidence := 0.9
			if confidence > bestConfidence {
				bestTemplate = template
				bestConfidence = confidence
			}
		}
	}
	
	// Pattern-based matching
	for _, pattern := range g.patterns {
		if pattern.Intent == request.Intent {
			confidence := pattern.Confidence
			
			// Boost confidence if triggers match description
			for _, trigger := range pattern.Triggers {
				if strings.Contains(strings.ToLower(request.Description), trigger) {
					confidence += 0.1
					if confidence > 1.0 {
						confidence = 1.0
					}
				}
			}
			
			if confidence > bestConfidence {
				if template, exists := g.templates[pattern.Template]; exists {
					bestTemplate = template
					bestConfidence = confidence
				}
			}
		}
	}
	
	return bestTemplate, bestConfidence
}

func (g *NixConfigGenerator) generateFromTemplate(template *ConfigTemplate, request GenerationRequest) (string, error) {
	// Simple template substitution (in production, would use a proper template engine)
	config := template.Template
	
	// Replace variables
	for name, value := range request.Variables {
		placeholder := fmt.Sprintf("{{.%s}}", name)
		
		switch v := value.(type) {
		case string:
			config = strings.ReplaceAll(config, placeholder, v)
		case bool:
			config = strings.ReplaceAll(config, placeholder, fmt.Sprintf("%v", v))
		case int:
			config = strings.ReplaceAll(config, placeholder, fmt.Sprintf("%d", v))
		case []string:
			// Handle array substitution (simplified)
			items := strings.Join(v, " ")
			config = strings.ReplaceAll(config, placeholder, items)
		}
	}
	
	// Remove conditional blocks (simplified implementation)
	config = g.processConditionals(config, request.Variables)
	
	return config, nil
}

func (g *NixConfigGenerator) processConditionals(config string, variables map[string]interface{}) string {
	// Simplified conditional processing
	// In production, would use a proper template engine like Go templates
	
	// Handle {{if .Variable}} blocks
	// This is a very basic implementation
	lines := strings.Split(config, "\n")
	var result []string
	var skipUntilEnd bool
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		if strings.HasPrefix(trimmed, "{{if ") {
			varName := strings.TrimPrefix(trimmed, "{{if .")
			varName = strings.TrimSuffix(varName, "}}")
			
			if value, exists := variables[varName]; exists {
				if boolValue, ok := value.(bool); ok && !boolValue {
					skipUntilEnd = true
				}
			} else {
				skipUntilEnd = true
			}
			continue
		}
		
		if strings.HasPrefix(trimmed, "{{end}}") {
			skipUntilEnd = false
			continue
		}
		
		if !skipUntilEnd {
			result = append(result, line)
		}
	}
	
	return strings.Join(result, "\n")
}

func (g *NixConfigGenerator) generateSuggestions(template *ConfigTemplate, request GenerationRequest, analysis *AnalysisResult) []string {
	var suggestions []string
	
	// Template-specific suggestions
	switch template.Name {
	case "web_server_nginx":
		suggestions = append(suggestions, "Consider enabling rate limiting for production use")
		suggestions = append(suggestions, "Add monitoring with Prometheus metrics")
	case "database_postgresql":
		suggestions = append(suggestions, "Consider setting up regular backups")
		suggestions = append(suggestions, "Enable connection pooling for high-traffic applications")
	case "security_hardening":
		suggestions = append(suggestions, "Regularly update your fail2ban rules")
		suggestions = append(suggestions, "Consider implementing intrusion detection")
	}
	
	// Analysis-based suggestions
	if analysis != nil {
		for _, opt := range analysis.Optimizations {
			suggestions = append(suggestions, opt.Description)
		}
	}
	
	return suggestions
}

func (g *NixConfigGenerator) generateWarnings(template *ConfigTemplate, request GenerationRequest, analysis *AnalysisResult) []string {
	var warnings []string
	
	// Analysis-based warnings
	if analysis != nil {
		for _, finding := range analysis.SecurityFindings {
			if finding.Severity == SeverityError || finding.Severity == SeverityWarning {
				warnings = append(warnings, finding.Description)
			}
		}
	}
	
	// Template-specific warnings
	if request.Context.Environment == "production" {
		warnings = append(warnings, "Review all security settings before deploying to production")
		warnings = append(warnings, "Ensure all secrets are managed externally")
	}
	
	return warnings
}

func (g *NixConfigGenerator) generateDocumentation(template *ConfigTemplate, request GenerationRequest) []string {
	var docs []string
	
	docs = append(docs, fmt.Sprintf("Generated from template: %s", template.Name))
	docs = append(docs, fmt.Sprintf("Template description: %s", template.Description))
	
	if len(template.Examples) > 0 {
		docs = append(docs, "Examples:")
		docs = append(docs, template.Examples...)
	}
	
	return docs
}

func (g *NixConfigGenerator) findRelatedTemplates(template *ConfigTemplate) []string {
	var related []string
	
	for name, t := range g.templates {
		if name != template.Name && t.Category == template.Category {
			related = append(related, name)
		}
	}
	
	return related
}

func (g *NixConfigGenerator) analyzeIntentFromText(text string) *IntentAnalysis {
	textLower := strings.ToLower(text)
	
	intent := IntentAnalysis{
		Confidence:  0.0,
		Patterns:    []string{},
		Annotations: make(map[string]string),
	}
	
	// Pattern matching for intent detection
	intentPatterns := map[string][]string{
		"service_configuration": {"server", "service", "daemon", "web", "database", "nginx", "apache", "postgresql"},
		"development_environment": {"development", "dev", "coding", "programming", "ide", "editor"},
		"gaming_configuration": {"gaming", "games", "steam", "graphics", "nvidia", "amd"},
		"security_configuration": {"security", "secure", "hardening", "firewall", "ssl", "encryption"},
		"package_management": {"install", "package", "software", "application", "tool"},
		"user_management": {"user", "account", "login", "authentication"},
		"network_configuration": {"network", "networking", "vpn", "dns", "dhcp"},
		"hardware_configuration": {"hardware", "driver", "bluetooth", "wifi", "sound"},
	}
	
	for intentType, keywords := range intentPatterns {
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(textLower, keyword) {
				matches++
			}
		}
		
		if matches > 0 {
			confidence := float64(matches) / float64(len(keywords))
			if confidence > intent.Confidence {
				intent.PrimaryIntent = intentType
				intent.Confidence = confidence
				intent.Patterns = append(intent.Patterns, fmt.Sprintf("text_pattern_%s", intentType))
			}
		}
	}
	
	if intent.PrimaryIntent == "" {
		intent.PrimaryIntent = "general_configuration"
		intent.Confidence = 0.1
	}
	
	return &intent
}