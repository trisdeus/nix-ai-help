package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	textTemplate "text/template"
	"time"
	"unicode"

	"nix-ai-help/pkg/logger"
)

// PluginTemplateManager manages plugin templates and scaffolding
type PluginTemplateManager struct {
	logger    *logger.Logger
	templates map[string]*PluginTemplate
}

// PluginTemplate represents a plugin template
type PluginTemplate struct {
	Name         string             `json:"name"`
	DisplayName  string             `json:"display_name"`
	Description  string             `json:"description"`
	Language     string             `json:"language"`
	Framework    string             `json:"framework"`
	Category     string             `json:"category"`
	Tags         []string           `json:"tags"`
	Files        []TemplateFile     `json:"files"`
	Variables    []TemplateVariable `json:"variables"`
	Instructions []string           `json:"instructions"`
	Examples     []TemplateExample  `json:"examples"`
}

// TemplateFile represents a file in a plugin template
type TemplateFile struct {
	Path        string `json:"path"`
	Content     string `json:"content"`
	Executable  bool   `json:"executable"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

// TemplateVariable represents a variable that can be customized in a template
type TemplateVariable struct {
	Name         string      `json:"name"`
	DisplayName  string      `json:"display_name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"` // string, bool, int, select
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value"`
	Options      []string    `json:"options"`    // For select type
	Validation   string      `json:"validation"` // Regex pattern
}

// TemplateExample represents an example plugin using the template
type TemplateExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Repository  string                 `json:"repository"`
	Variables   map[string]interface{} `json:"variables"`
}

// PluginScaffoldOptions represents options for scaffolding a new plugin
type PluginScaffoldOptions struct {
	PluginName   string                 `json:"plugin_name"`
	OutputDir    string                 `json:"output_dir"`
	Template     string                 `json:"template"`
	Variables    map[string]interface{} `json:"variables"`
	Interactive  bool                   `json:"interactive"`
	OverwriteAll bool                   `json:"overwrite_all"`
	GitInit      bool                   `json:"git_init"`
	License      string                 `json:"license"`
}

// NewPluginTemplateManager creates a new plugin template manager
func NewPluginTemplateManager(log *logger.Logger) *PluginTemplateManager {
	ptm := &PluginTemplateManager{
		logger:    log,
		templates: make(map[string]*PluginTemplate),
	}

	// Initialize built-in templates
	ptm.initializeBuiltinTemplates()

	return ptm
}

// GetAvailableTemplates returns all available plugin templates
func (ptm *PluginTemplateManager) GetAvailableTemplates() map[string]*PluginTemplate {
	return ptm.templates
}

// GetTemplate retrieves a specific template by name
func (ptm *PluginTemplateManager) GetTemplate(name string) (*PluginTemplate, error) {
	template, exists := ptm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}
	return template, nil
}

// ScaffoldPlugin creates a new plugin from a template
func (ptm *PluginTemplateManager) ScaffoldPlugin(options PluginScaffoldOptions) error {
	ptm.logger.Info(fmt.Sprintf("Scaffolding plugin '%s' using template '%s'", options.PluginName, options.Template))

	// Get template
	template, err := ptm.GetTemplate(options.Template)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	// Validate plugin name
	if err := ptm.validatePluginName(options.PluginName); err != nil {
		return fmt.Errorf("invalid plugin name: %w", err)
	}

	// Create output directory
	pluginDir := filepath.Join(options.OutputDir, options.PluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Prepare template variables
	templateVars := ptm.prepareTemplateVariables(options, template)

	// Generate files from template
	if err := ptm.generateFiles(template, pluginDir, templateVars, options); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// Initialize git repository if requested
	if options.GitInit {
		if err := ptm.initializeGitRepo(pluginDir); err != nil {
			ptm.logger.Warn(fmt.Sprintf("Failed to initialize git repository: %v", err))
		}
	}

	// Generate instructions
	if err := ptm.generateInstructions(template, pluginDir, options); err != nil {
		ptm.logger.Warn(fmt.Sprintf("Failed to generate instructions: %v", err))
	}

	ptm.logger.Info(fmt.Sprintf("Successfully scaffolded plugin '%s' in %s", options.PluginName, pluginDir))
	return nil
}

// AddTemplate adds a new template to the manager
func (ptm *PluginTemplateManager) AddTemplate(template *PluginTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	ptm.templates[template.Name] = template
	ptm.logger.Info(fmt.Sprintf("Added template: %s", template.Name))
	return nil
}

// RemoveTemplate removes a template from the manager
func (ptm *PluginTemplateManager) RemoveTemplate(name string) error {
	if _, exists := ptm.templates[name]; !exists {
		return fmt.Errorf("template '%s' not found", name)
	}

	delete(ptm.templates, name)
	ptm.logger.Info(fmt.Sprintf("Removed template: %s", name))
	return nil
}

// ValidateTemplate validates a template structure
func (ptm *PluginTemplateManager) ValidateTemplate(template *PluginTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Language == "" {
		return fmt.Errorf("template language is required")
	}

	if len(template.Files) == 0 {
		return fmt.Errorf("template must have at least one file")
	}

	// Validate template variables
	for _, variable := range template.Variables {
		if variable.Name == "" {
			return fmt.Errorf("variable name is required")
		}
		if variable.Type == "" {
			return fmt.Errorf("variable type is required for '%s'", variable.Name)
		}
	}

	return nil
}

// Private helper methods

func (ptm *PluginTemplateManager) initializeBuiltinTemplates() {
	// Basic Go Plugin Template
	basicGoTemplate := &PluginTemplate{
		Name:        "basic-go",
		DisplayName: "Basic Go Plugin",
		Description: "A basic plugin template using Go",
		Language:    "go",
		Framework:   "nixai",
		Category:    "basic",
		Tags:        []string{"go", "basic", "starter"},
		Variables: []TemplateVariable{
			{
				Name:         "plugin_name",
				DisplayName:  "Plugin Name",
				Description:  "The name of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "my-plugin",
				Validation:   "^[a-z][a-z0-9-]*$",
			},
			{
				Name:         "author",
				DisplayName:  "Author",
				Description:  "The plugin author",
				Type:         "string",
				Required:     true,
				DefaultValue: "Plugin Developer",
			},
			{
				Name:         "description",
				DisplayName:  "Description",
				Description:  "A brief description of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "A nixai plugin",
			},
			{
				Name:         "license",
				DisplayName:  "License",
				Description:  "The plugin license",
				Type:         "select",
				Required:     false,
				DefaultValue: "MIT",
				Options:      []string{"MIT", "Apache-2.0", "GPL-3.0", "BSD-3-Clause"},
			},
		},
		Files: []TemplateFile{
			{
				Path:        "main.go",
				Content:     basicGoPluginMainTemplate,
				Required:    true,
				Description: "Main plugin implementation",
			},
			{
				Path:        "go.mod",
				Content:     basicGoPluginModTemplate,
				Required:    true,
				Description: "Go module definition",
			},
			{
				Path:        "README.md",
				Content:     basicPluginReadmeTemplate,
				Required:    true,
				Description: "Plugin documentation",
			},
			{
				Path:        "LICENSE",
				Content:     mitLicenseTemplate,
				Required:    false,
				Description: "License file",
			},
			{
				Path:        ".gitignore",
				Content:     goGitignoreTemplate,
				Required:    false,
				Description: "Git ignore file",
			},
		},
		Instructions: []string{
			"1. Navigate to the plugin directory",
			"2. Run 'go mod tidy' to download dependencies",
			"3. Build the plugin with 'go build -buildmode=plugin -o plugin.so .'",
			"4. Install the plugin with 'nixai plugin install plugin.so'",
			"5. Test your plugin with 'nixai plugin execute {{.plugin_name}} hello'",
		},
		Examples: []TemplateExample{
			{
				Name:        "Hello World Plugin",
				Description: "A simple hello world plugin",
				Repository:  "https://github.com/nixai-plugins/hello-world",
				Variables: map[string]interface{}{
					"plugin_name": "hello-world",
					"author":      "NixAI Team",
					"description": "A simple hello world plugin for demonstration",
				},
			},
		},
	}

	// Advanced Go Plugin Template
	advancedGoTemplate := &PluginTemplate{
		Name:        "advanced-go",
		DisplayName: "Advanced Go Plugin",
		Description: "An advanced plugin template with configuration and multiple operations",
		Language:    "go",
		Framework:   "nixai",
		Category:    "advanced",
		Tags:        []string{"go", "advanced", "configuration"},
		Variables: []TemplateVariable{
			{
				Name:         "plugin_name",
				DisplayName:  "Plugin Name",
				Description:  "The name of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "my-advanced-plugin",
				Validation:   "^[a-z][a-z0-9-]*$",
			},
			{
				Name:         "author",
				DisplayName:  "Author",
				Description:  "The plugin author",
				Type:         "string",
				Required:     true,
				DefaultValue: "Plugin Developer",
			},
			{
				Name:         "description",
				DisplayName:  "Description",
				Description:  "A brief description of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "An advanced nixai plugin",
			},
			{
				Name:         "with_config",
				DisplayName:  "Include Configuration",
				Description:  "Include configuration file support",
				Type:         "bool",
				Required:     false,
				DefaultValue: true,
			},
			{
				Name:         "with_web_api",
				DisplayName:  "Include Web API",
				Description:  "Include web API endpoints",
				Type:         "bool",
				Required:     false,
				DefaultValue: false,
			},
		},
		Files: []TemplateFile{
			{
				Path:        "main.go",
				Content:     advancedGoPluginMainTemplate,
				Required:    true,
				Description: "Main plugin implementation",
			},
			{
				Path:        "config.go",
				Content:     advancedGoPluginConfigTemplate,
				Required:    false,
				Description: "Configuration handling",
			},
			{
				Path:        "operations.go",
				Content:     advancedGoPluginOperationsTemplate,
				Required:    true,
				Description: "Plugin operations",
			},
			{
				Path:        "go.mod",
				Content:     advancedGoPluginModTemplate,
				Required:    true,
				Description: "Go module definition",
			},
			{
				Path:        "README.md",
				Content:     advancedPluginReadmeTemplate,
				Required:    true,
				Description: "Plugin documentation",
			},
		},
		Instructions: []string{
			"1. Navigate to the plugin directory",
			"2. Run 'go mod tidy' to download dependencies",
			"3. Customize the operations in operations.go",
			"4. Update configuration in config.go if needed",
			"5. Build the plugin with 'go build -buildmode=plugin -o plugin.so .'",
			"6. Install the plugin with 'nixai plugin install plugin.so'",
		},
	}

	// Add templates to manager
	ptm.templates["basic-go"] = basicGoTemplate
	ptm.templates["advanced-go"] = advancedGoTemplate

	// NixOS Integration Template
	nixosIntegrationTemplate := &PluginTemplate{
		Name:        "nixos-integration",
		DisplayName: "NixOS Integration Plugin",
		Description: "A plugin template for deep NixOS system integration",
		Language:    "go",
		Framework:   "nixai",
		Category:    "integration",
		Tags:        []string{"go", "nixos", "system", "integration"},
		Variables: []TemplateVariable{
			{
				Name:         "plugin_name",
				DisplayName:  "Plugin Name",
				Description:  "The name of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "my-nixos-plugin",
				Validation:   "^[a-z][a-z0-9-]*$",
			},
			{
				Name:         "author",
				DisplayName:  "Author",
				Description:  "The plugin author",
				Type:         "string",
				Required:     true,
				DefaultValue: "Plugin Developer",
			},
			{
				Name:         "description",
				DisplayName:  "Description",
				Description:  "A brief description of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "A NixOS integration plugin",
			},
		},
		Files: []TemplateFile{
			{
				Path:        "main.go",
				Content:     nixosIntegrationPluginMainTemplate,
				Required:    true,
				Description: "Main plugin implementation with NixOS integration",
			},
			{
				Path:        "nixos.go",
				Content:     nixosIntegrationHelperTemplate,
				Required:    true,
				Description: "NixOS system integration helpers",
			},
			{
				Path:        "go.mod",
				Content:     nixosIntegrationModTemplate,
				Required:    true,
				Description: "Go module definition with NixOS dependencies",
			},
			{
				Path:        "README.md",
				Content:     nixosIntegrationReadmeTemplate,
				Required:    true,
				Description: "Plugin documentation with NixOS examples",
			},
		},
		Instructions: []string{
			"1. Navigate to the plugin directory",
			"2. Run 'go mod tidy' to download dependencies",
			"3. Customize NixOS integration in nixos.go",
			"4. Implement your NixOS-specific operations",
			"5. Build the plugin with 'go build -buildmode=plugin -o plugin.so .'",
			"6. Install the plugin with 'nixai plugin install plugin.so'",
		},
	}

	// AI Provider Template
	aiProviderTemplate := &PluginTemplate{
		Name:        "ai-provider",
		DisplayName: "AI Provider Plugin",
		Description: "A plugin template for integrating new AI providers",
		Language:    "go",
		Framework:   "nixai",
		Category:    "ai",
		Tags:        []string{"go", "ai", "provider", "llm"},
		Variables: []TemplateVariable{
			{
				Name:         "plugin_name",
				DisplayName:  "Plugin Name",
				Description:  "The name of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "my-ai-provider",
				Validation:   "^[a-z][a-z0-9-]*$",
			},
			{
				Name:         "provider_name",
				DisplayName:  "Provider Name",
				Description:  "The name of the AI provider (e.g., 'OpenAI', 'Claude')",
				Type:         "string",
				Required:     true,
				DefaultValue: "MyAI",
			},
			{
				Name:         "author",
				DisplayName:  "Author",
				Description:  "The plugin author",
				Type:         "string",
				Required:     true,
				DefaultValue: "Plugin Developer",
			},
		},
		Files: []TemplateFile{
			{
				Path:        "main.go",
				Content:     aiProviderPluginMainTemplate,
				Required:    true,
				Description: "Main plugin implementation with AI provider interface",
			},
			{
				Path:        "provider.go",
				Content:     aiProviderImplementationTemplate,
				Required:    true,
				Description: "AI provider implementation",
			},
			{
				Path:        "go.mod",
				Content:     aiProviderModTemplate,
				Required:    true,
				Description: "Go module definition with AI dependencies",
			},
			{
				Path:        "README.md",
				Content:     aiProviderReadmeTemplate,
				Required:    true,
				Description: "Plugin documentation with AI provider examples",
			},
		},
		Instructions: []string{
			"1. Navigate to the plugin directory",
			"2. Run 'go mod tidy' to download dependencies",
			"3. Implement the AI provider interface in provider.go",
			"4. Configure authentication and API endpoints",
			"5. Build the plugin with 'go build -buildmode=plugin -o plugin.so .'",
			"6. Install the plugin with 'nixai plugin install plugin.so'",
		},
	}

	// Tool Integration Template
	toolIntegrationTemplate := &PluginTemplate{
		Name:        "tool-integration",
		DisplayName: "Tool Integration Plugin",
		Description: "A plugin template for integrating external tools",
		Language:    "go",
		Framework:   "nixai",
		Category:    "integration",
		Tags:        []string{"go", "tools", "external", "integration"},
		Variables: []TemplateVariable{
			{
				Name:         "plugin_name",
				DisplayName:  "Plugin Name",
				Description:  "The name of the plugin",
				Type:         "string",
				Required:     true,
				DefaultValue: "my-tool-plugin",
				Validation:   "^[a-z][a-z0-9-]*$",
			},
			{
				Name:         "tool_name",
				DisplayName:  "Tool Name",
				Description:  "The name of the external tool",
				Type:         "string",
				Required:     true,
				DefaultValue: "external-tool",
			},
			{
				Name:         "author",
				DisplayName:  "Author",
				Description:  "The plugin author",
				Type:         "string",
				Required:     true,
				DefaultValue: "Plugin Developer",
			},
		},
		Files: []TemplateFile{
			{
				Path:        "main.go",
				Content:     toolIntegrationPluginMainTemplate,
				Required:    true,
				Description: "Main plugin implementation with tool integration",
			},
			{
				Path:        "tool.go",
				Content:     toolIntegrationHelperTemplate,
				Required:    true,
				Description: "External tool integration helpers",
			},
			{
				Path:        "go.mod",
				Content:     toolIntegrationModTemplate,
				Required:    true,
				Description: "Go module definition",
			},
			{
				Path:        "README.md",
				Content:     toolIntegrationReadmeTemplate,
				Required:    true,
				Description: "Plugin documentation with tool integration examples",
			},
		},
		Instructions: []string{
			"1. Navigate to the plugin directory",
			"2. Run 'go mod tidy' to download dependencies",
			"3. Implement tool integration in tool.go",
			"4. Configure tool paths and options",
			"5. Build the plugin with 'go build -buildmode=plugin -o plugin.so .'",
			"6. Install the plugin with 'nixai plugin install plugin.so'",
		},
	}

	ptm.templates["nixos-integration"] = nixosIntegrationTemplate
	ptm.templates["ai-provider"] = aiProviderTemplate
	ptm.templates["tool-integration"] = toolIntegrationTemplate

	ptm.logger.Info("Initialized built-in plugin templates")
}

func (ptm *PluginTemplateManager) validatePluginName(name string) error {
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("plugin name cannot contain spaces")
	}

	if strings.ContainsAny(name, "!@#$%^&*()+={}[]|\\:;\"'<>?,./") {
		return fmt.Errorf("plugin name contains invalid characters")
	}

	return nil
}

func (ptm *PluginTemplateManager) prepareTemplateVariables(options PluginScaffoldOptions, template *PluginTemplate) map[string]interface{} {
	vars := make(map[string]interface{})

	// Set default values
	for _, variable := range template.Variables {
		vars[variable.Name] = variable.DefaultValue
	}

	// Override with provided values
	for key, value := range options.Variables {
		vars[key] = value
	}

	// Always set plugin_name
	vars["plugin_name"] = options.PluginName
	vars["plugin_name_title"] = toTitle(strings.ReplaceAll(options.PluginName, "-", " "))
	vars["plugin_name_camel"] = toCamelCase(options.PluginName)
	vars["current_year"] = time.Now().Year()
	vars["current_date"] = time.Now().Format("2006-01-02")

	return vars
}

func (ptm *PluginTemplateManager) generateFiles(pluginTemplate *PluginTemplate, pluginDir string, vars map[string]interface{}, options PluginScaffoldOptions) error {
	for _, file := range pluginTemplate.Files {
		// Skip optional files if conditions aren't met
		if !file.Required && !ptm.shouldIncludeFile(file, vars) {
			continue
		}

		filePath := filepath.Join(pluginDir, file.Path)

		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}

		// Check if file exists and handle overwrite
		if _, err := os.Stat(filePath); err == nil && !options.OverwriteAll {
			ptm.logger.Warn(fmt.Sprintf("File %s already exists, skipping", file.Path))
			continue
		}

		// Process template
		tmpl, err := textTemplate.New(file.Path).Parse(file.Content)
		if err != nil {
			return fmt.Errorf("failed to parse template for %s: %w", file.Path, err)
		}

		// Create file
		f, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", file.Path, err)
		}
		defer f.Close()

		// Execute template
		if err := tmpl.Execute(f, vars); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", file.Path, err)
		}

		// Set executable permission if needed
		if file.Executable {
			if err := os.Chmod(filePath, 0755); err != nil {
				ptm.logger.Warn(fmt.Sprintf("Failed to set executable permission for %s: %v", file.Path, err))
			}
		}

		ptm.logger.Info(fmt.Sprintf("Generated file: %s", file.Path))
	}

	return nil
}

func (ptm *PluginTemplateManager) shouldIncludeFile(file TemplateFile, vars map[string]interface{}) bool {
	// Basic logic for conditional file inclusion
	switch file.Path {
	case "config.go":
		if withConfig, ok := vars["with_config"].(bool); ok {
			return withConfig
		}
	case "LICENSE":
		if license, ok := vars["license"].(string); ok {
			return license != ""
		}
	}
	return true
}

func (ptm *PluginTemplateManager) initializeGitRepo(pluginDir string) error {
	// This would run git init in the plugin directory
	ptm.logger.Info("Git repository initialization not implemented")
	return nil
}

func (ptm *PluginTemplateManager) generateInstructions(template *PluginTemplate, pluginDir string, options PluginScaffoldOptions) error {
	if len(template.Instructions) == 0 {
		return nil
	}

	instructionsPath := filepath.Join(pluginDir, "INSTRUCTIONS.md")
	f, err := os.Create(instructionsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# %s Plugin Instructions\n\n", options.PluginName)
	fmt.Fprintf(f, "This plugin was generated using the '%s' template.\n\n", template.Name)
	fmt.Fprintf(f, "## Next Steps\n\n")

	for _, instruction := range template.Instructions {
		tmpl, err := textTemplate.New("instruction").Parse(instruction)
		if err != nil {
			fmt.Fprintf(f, "- %s\n", instruction)
			continue
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, map[string]interface{}{
			"plugin_name": options.PluginName,
		}); err != nil {
			fmt.Fprintf(f, "- %s\n", instruction)
			continue
		}

		fmt.Fprintf(f, "- %s\n", buf.String())
	}

	return nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
		} else {
			parts[i] = toTitle(parts[i])
		}
	}
	return strings.Join(parts, "")
}

func toTitle(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Template content constants
const basicGoPluginMainTemplate = `package main

import (
	"context"
	"fmt"
	"time"
)

// {{.plugin_name_camel}}Plugin implements the PluginInterface
type {{.plugin_name_camel}}Plugin struct {
	name        string
	version     string
	description string
	author      string
}

// NewPlugin creates a new instance of the plugin
func NewPlugin() PluginInterface {
	return &{{.plugin_name_camel}}Plugin{
		name:        "{{.plugin_name}}",
		version:     "1.0.0",
		description: "{{.description}}",
		author:      "{{.author}}",
	}
}

// Metadata methods
func (p *{{.plugin_name_camel}}Plugin) Name() string { return p.name }
func (p *{{.plugin_name_camel}}Plugin) Version() string { return p.version }
func (p *{{.plugin_name_camel}}Plugin) Description() string { return p.description }
func (p *{{.plugin_name_camel}}Plugin) Author() string { return p.author }
func (p *{{.plugin_name_camel}}Plugin) Repository() string { return "" }
func (p *{{.plugin_name_camel}}Plugin) License() string { return "{{.license}}" }
func (p *{{.plugin_name_camel}}Plugin) Dependencies() []string { return []string{} }
func (p *{{.plugin_name_camel}}Plugin) Capabilities() []string { return []string{"hello"} }

// Lifecycle methods
func (p *{{.plugin_name_camel}}Plugin) Initialize(ctx context.Context, config PluginConfig) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Start(ctx context.Context) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Stop(ctx context.Context) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Cleanup(ctx context.Context) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) IsRunning() bool {
	return true
}

// Execution methods
func (p *{{.plugin_name_camel}}Plugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "hello":
		name := "World"
		if n, ok := params["name"].(string); ok {
			name = n
		}
		return map[string]string{
			"message": fmt.Sprintf("Hello, %s!", name),
		}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetOperations() []PluginOperation {
	return []PluginOperation{
		{
			Name:        "hello",
			Description: "Say hello to someone",
			Parameters: []PluginParameter{
				{
					Name:        "name",
					Type:        "string",
					Description: "Name to greet",
					Required:    false,
				},
			},
		},
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetSchema(operation string) (*PluginSchema, error) {
	return nil, fmt.Errorf("schema not implemented")
}

// Health and Status methods
func (p *{{.plugin_name_camel}}Plugin) HealthCheck(ctx context.Context) PluginHealth {
	return PluginHealth{
		Status:    HealthHealthy,
		Message:   "Plugin is healthy",
		LastCheck: time.Now(),
		Uptime:    time.Since(time.Now()),
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetMetrics() PluginMetrics {
	return PluginMetrics{
		ExecutionCount:       0,
		ErrorCount:          0,
		SuccessRate:         1.0,
		TotalExecutionTime:  0,
		AverageExecutionTime: 0,
		LastExecutionTime:   time.Time{},
		MemoryUsage:         0,
		CPUUsage:           0,
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetStatus() PluginStatus {
	return PluginStatus{
		State:       StateRunning,
		Message:     "Plugin is running",
		LastUpdated: time.Now(),
		Version:     p.version,
	}
}

// Interface definitions (these would normally be imported)
type PluginInterface interface {
	// Metadata
	Name() string
	Version() string
	Description() string
	Author() string
	Repository() string
	License() string
	Dependencies() []string
	Capabilities() []string

	// Lifecycle
	Initialize(ctx context.Context, config PluginConfig) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Cleanup(ctx context.Context) error
	IsRunning() bool

	// Execution
	Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
	GetOperations() []PluginOperation
	GetSchema(operation string) (*PluginSchema, error)

	// Health and Status
	HealthCheck(ctx context.Context) PluginHealth
	GetMetrics() PluginMetrics
	GetStatus() PluginStatus
}

// Supporting types
type PluginConfig struct {
	Name          string                 ` + "`" + `json:"name"` + "`" + `
	Configuration map[string]interface{} ` + "`" + `json:"configuration"` + "`" + `
}

type PluginOperation struct {
	Name        string            ` + "`" + `json:"name"` + "`" + `
	Description string            ` + "`" + `json:"description"` + "`" + `
	Parameters  []PluginParameter ` + "`" + `json:"parameters"` + "`" + `
}

type PluginParameter struct {
	Name        string ` + "`" + `json:"name"` + "`" + `
	Type        string ` + "`" + `json:"type"` + "`" + `
	Description string ` + "`" + `json:"description"` + "`" + `
	Required    bool   ` + "`" + `json:"required"` + "`" + `
}

type PluginSchema struct{}

type PluginHealth struct {
	Status    HealthStatus  ` + "`" + `json:"status"` + "`" + `
	Message   string        ` + "`" + `json:"message"` + "`" + `
	LastCheck time.Time     ` + "`" + `json:"last_check"` + "`" + `
	Uptime    time.Duration ` + "`" + `json:"uptime"` + "`" + `
}

type PluginMetrics struct {
	ExecutionCount       int64         ` + "`" + `json:"execution_count"` + "`" + `
	ErrorCount          int64         ` + "`" + `json:"error_count"` + "`" + `
	SuccessRate         float64       ` + "`" + `json:"success_rate"` + "`" + `
	TotalExecutionTime  time.Duration ` + "`" + `json:"total_execution_time"` + "`" + `
	AverageExecutionTime time.Duration ` + "`" + `json:"average_execution_time"` + "`" + `
	LastExecutionTime   time.Time     ` + "`" + `json:"last_execution_time"` + "`" + `
	MemoryUsage         int64         ` + "`" + `json:"memory_usage"` + "`" + `
	CPUUsage           float64       ` + "`" + `json:"cpu_usage"` + "`" + `
}

type PluginStatus struct {
	State       PluginState ` + "`" + `json:"state"` + "`" + `
	Message     string      ` + "`" + `json:"message"` + "`" + `
	LastUpdated time.Time   ` + "`" + `json:"last_updated"` + "`" + `
	Version     string      ` + "`" + `json:"version"` + "`" + `
}

type HealthStatus int
type PluginState int

const (
	HealthHealthy HealthStatus = iota
	StateRunning  PluginState  = iota
)
`

const basicGoPluginModTemplate = `module {{.plugin_name}}

go 1.21

require (
	// Add your dependencies here
)
`

const basicPluginReadmeTemplate = `# {{.plugin_name_title}} Plugin

{{.description}}

## Installation

1. Build the plugin:
   ` + "```bash" + `
   go build -buildmode=plugin -o {{.plugin_name}}.so .
   ` + "```" + `

2. Install the plugin:
   ` + "```bash" + `
   nixai plugin install {{.plugin_name}}.so
   ` + "```" + `

## Usage

` + "```bash" + `
# Say hello
nixai plugin execute {{.plugin_name}} hello

# Say hello to someone specific
nixai plugin execute {{.plugin_name}} hello --params '{"name": "Alice"}'
` + "```" + `

## Operations

- **hello**: Say hello to someone

## Author

{{.author}}

## License

{{.license}}
`

const mitLicenseTemplate = `MIT License

Copyright (c) {{.current_year}} {{.author}}

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

const goGitignoreTemplate = `# Compiled plugins
*.so

# Go build artifacts
/vendor/
/bin/
/dist/

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Testing
coverage.out
`

const advancedGoPluginMainTemplate = `// Advanced plugin template with more features
package main

// Implementation similar to basic template but with more operations and features
func NewPlugin() PluginInterface {
	return &{{.plugin_name_camel}}Plugin{
		name:        "{{.plugin_name}}",
		version:     "1.0.0",
		description: "{{.description}}",
		author:      "{{.author}}",
	}
}
`

const advancedGoPluginModTemplate = `module {{.plugin_name}}

go 1.21

require (
	github.com/spf13/viper v1.16.0
	gopkg.in/yaml.v3 v3.0.1
)
`

const advancedGoPluginConfigTemplate = `package main

// Configuration handling for the plugin
type PluginConfig struct {
	// Add configuration fields here
}
`

const advancedGoPluginOperationsTemplate = `package main

// Advanced operations for the plugin
func (p *{{.plugin_name_camel}}Plugin) GetOperations() []PluginOperation {
	return []PluginOperation{
		// Define your operations here
	}
}
`

const advancedPluginReadmeTemplate = `# {{.plugin_name_title}} Plugin

{{.description}}

This is an advanced plugin with multiple operations and configuration support.

## Features

- Multiple operations
- Configuration support
- Advanced error handling
- Metrics collection

## Installation

See basic installation instructions...
`

// NixOS Integration Plugin Templates
const nixosIntegrationPluginMainTemplate = `package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"nix-ai-help/internal/plugins"
)

type {{.plugin_name_camel}}Plugin struct {
	name        string
	version     string
	description string
	author      string
	running     bool
	nixosHelper *NixOSHelper
}

func NewPlugin() plugins.PluginInterface {
	return &{{.plugin_name_camel}}Plugin{
		name:        "{{.plugin_name}}",
		version:     "1.0.0",
		description: "{{.description}}",
		author:      "{{.author}}",
		nixosHelper: NewNixOSHelper(),
	}
}

func (p *{{.plugin_name_camel}}Plugin) Name() string        { return p.name }
func (p *{{.plugin_name_camel}}Plugin) Version() string     { return p.version }
func (p *{{.plugin_name_camel}}Plugin) Description() string { return p.description }
func (p *{{.plugin_name_camel}}Plugin) Author() string      { return p.author }
func (p *{{.plugin_name_camel}}Plugin) Repository() string  { return "" }
func (p *{{.plugin_name_camel}}Plugin) License() string     { return "MIT" }
func (p *{{.plugin_name_camel}}Plugin) Dependencies() []string { return []string{} }
func (p *{{.plugin_name_camel}}Plugin) Capabilities() []string { return []string{"nixos-system"} }

func (p *{{.plugin_name_camel}}Plugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Cleanup(ctx context.Context) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) IsRunning() bool {
	return p.running
}

func (p *{{.plugin_name_camel}}Plugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "system-info":
		return p.nixosHelper.GetSystemInfo()
	case "rebuild":
		return p.nixosHelper.Rebuild()
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "system-info",
			Description: "Get NixOS system information",
			Parameters:  []plugins.PluginParameter{},
		},
		{
			Name:        "rebuild",
			Description: "Rebuild NixOS configuration",
			Parameters:  []plugins.PluginParameter{},
		},
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	return nil, nil
}

func (p *{{.plugin_name_camel}}Plugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	return plugins.PluginHealth{
		Status:  "healthy",
		Message: "Plugin is running normally",
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{}
}

func (p *{{.plugin_name_camel}}Plugin) GetStatus() plugins.PluginStatus {
	return plugins.PluginStatus{
		State:   3, // StateRunning
		Message: "Running",
	}
}
`

const nixosIntegrationHelperTemplate = `package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type NixOSHelper struct{}

func NewNixOSHelper() *NixOSHelper {
	return &NixOSHelper{}
}

func (h *NixOSHelper) GetSystemInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	// Get NixOS version
	if version, err := h.runCommand("nixos-version"); err == nil {
		info["version"] = strings.TrimSpace(version)
	}
	
	// Get system configuration
	if config, err := h.runCommand("nixos-option system.stateVersion"); err == nil {
		info["state_version"] = strings.TrimSpace(config)
	}
	
	return info, nil
}

func (h *NixOSHelper) Rebuild() (string, error) {
	return h.runCommand("nixos-rebuild", "switch")
}

func (h *NixOSHelper) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}
	return string(output), nil
}
`

const nixosIntegrationModTemplate = `module {{.plugin_name}}

go 1.21

require (
	nix-ai-help v0.0.0
)

replace nix-ai-help => ../../../
`

const nixosIntegrationReadmeTemplate = `# {{.plugin_name_title}} Plugin

{{.description}}

This plugin provides deep integration with NixOS system management.

## Features

- System information retrieval
- NixOS configuration rebuilding
- System health monitoring
- NixOS-specific operations

## Installation

` + "```" + `bash
go build -buildmode=plugin -o {{.plugin_name}}.so .
nixai plugin install {{.plugin_name}}.so
` + "```" + `

## Usage

` + "```" + `bash
# Get system information
nixai plugin execute {{.plugin_name}} system-info

# Rebuild system
nixai plugin execute {{.plugin_name}} rebuild
` + "```" + `

## Author

{{.author}}
`

// AI Provider Plugin Templates
const aiProviderPluginMainTemplate = `package main

import (
	"context"
	"fmt"

	"nix-ai-help/internal/plugins"
)

type {{.plugin_name_camel}}Plugin struct {
	name        string
	version     string
	description string
	author      string
	running     bool
	provider    *{{.provider_name}}Provider
}

func NewPlugin() plugins.PluginInterface {
	return &{{.plugin_name_camel}}Plugin{
		name:        "{{.plugin_name}}",
		version:     "1.0.0",
		description: "{{.description}}",
		author:      "{{.author}}",
		provider:    New{{.provider_name}}Provider(),
	}
}

func (p *{{.plugin_name_camel}}Plugin) Name() string        { return p.name }
func (p *{{.plugin_name_camel}}Plugin) Version() string     { return p.version }
func (p *{{.plugin_name_camel}}Plugin) Description() string { return p.description }
func (p *{{.plugin_name_camel}}Plugin) Author() string      { return p.author }
func (p *{{.plugin_name_camel}}Plugin) Repository() string  { return "" }
func (p *{{.plugin_name_camel}}Plugin) License() string     { return "MIT" }
func (p *{{.plugin_name_camel}}Plugin) Dependencies() []string { return []string{} }
func (p *{{.plugin_name_camel}}Plugin) Capabilities() []string { return []string{"ai-provider"} }

func (p *{{.plugin_name_camel}}Plugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	return p.provider.Initialize(config)
}

func (p *{{.plugin_name_camel}}Plugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Cleanup(ctx context.Context) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) IsRunning() bool {
	return p.running
}

func (p *{{.plugin_name_camel}}Plugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "query":
		if question, ok := params["question"].(string); ok {
			return p.provider.Query(question)
		}
		return nil, fmt.Errorf("missing question parameter")
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "query",
			Description: "Query the AI provider",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "question",
					Type:        "string",
					Description: "The question to ask",
					Required:    true,
				},
			},
		},
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	return nil, nil
}

func (p *{{.plugin_name_camel}}Plugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	return plugins.PluginHealth{
		Status:  "healthy",
		Message: "AI provider plugin is running normally",
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{}
}

func (p *{{.plugin_name_camel}}Plugin) GetStatus() plugins.PluginStatus {
	return plugins.PluginStatus{
		State:   3, // StateRunning
		Message: "Running",
	}
}
`

const aiProviderImplementationTemplate = `package main

import (
	"fmt"
	"nix-ai-help/internal/plugins"
)

type {{.provider_name}}Provider struct {
	apiKey   string
	baseURL  string
	model    string
}

func New{{.provider_name}}Provider() *{{.provider_name}}Provider {
	return &{{.provider_name}}Provider{
		baseURL: "https://api.example.com",
		model:   "default-model",
	}
}

func (p *{{.provider_name}}Provider) Initialize(config plugins.PluginConfig) error {
	// Initialize your provider with configuration
	// Set API keys, base URLs, etc.
	return nil
}

func (p *{{.provider_name}}Provider) Query(question string) (string, error) {
	// Implement your AI provider query logic here
	// This is where you'd call the provider's API
	return fmt.Sprintf("Response from {{.provider_name}} for: %s", question), nil
}
`

const aiProviderModTemplate = `module {{.plugin_name}}

go 1.21

require (
	nix-ai-help v0.0.0
)

replace nix-ai-help => ../../../
`

const aiProviderReadmeTemplate = `# {{.plugin_name_title}} Plugin

{{.description}}

This plugin integrates {{.provider_name}} AI provider with nixai.

## Features

- {{.provider_name}} AI integration
- Query processing
- Response handling
- Configuration management

## Configuration

Set your API key:
` + "```" + `bash
export API_KEY="your-api-key"
` + "```" + `

## Installation

` + "```" + `bash
go build -buildmode=plugin -o {{.plugin_name}}.so .
nixai plugin install {{.plugin_name}}.so
` + "```" + `

## Usage

` + "```" + `bash
# Query the AI provider
nixai plugin execute {{.plugin_name}} query --params '{"question": "How do I configure NixOS?"}'
` + "```" + `

## Author

{{.author}}
`

// Tool Integration Plugin Templates
const toolIntegrationPluginMainTemplate = `package main

import (
	"context"
	"fmt"

	"nix-ai-help/internal/plugins"
)

type {{.plugin_name_camel}}Plugin struct {
	name        string
	version     string
	description string
	author      string
	running     bool
	tool        *ExternalTool
}

func NewPlugin() plugins.PluginInterface {
	return &{{.plugin_name_camel}}Plugin{
		name:        "{{.plugin_name}}",
		version:     "1.0.0",
		description: "{{.description}}",
		author:      "{{.author}}",
		tool:        NewExternalTool(),
	}
}

func (p *{{.plugin_name_camel}}Plugin) Name() string        { return p.name }
func (p *{{.plugin_name_camel}}Plugin) Version() string     { return p.version }
func (p *{{.plugin_name_camel}}Plugin) Description() string { return p.description }
func (p *{{.plugin_name_camel}}Plugin) Author() string      { return p.author }
func (p *{{.plugin_name_camel}}Plugin) Repository() string  { return "" }
func (p *{{.plugin_name_camel}}Plugin) License() string     { return "MIT" }
func (p *{{.plugin_name_camel}}Plugin) Dependencies() []string { return []string{"{{.tool_name}}"} }
func (p *{{.plugin_name_camel}}Plugin) Capabilities() []string { return []string{"tool-integration"} }

func (p *{{.plugin_name_camel}}Plugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	return p.tool.Initialize()
}

func (p *{{.plugin_name_camel}}Plugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) Cleanup(ctx context.Context) error {
	return nil
}

func (p *{{.plugin_name_camel}}Plugin) IsRunning() bool {
	return p.running
}

func (p *{{.plugin_name_camel}}Plugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "run":
		if args, ok := params["args"].([]string); ok {
			return p.tool.Run(args)
		}
		return nil, fmt.Errorf("missing args parameter")
	case "version":
		return p.tool.GetVersion()
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "run",
			Description: "Run the external tool",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "args",
					Type:        "array",
					Description: "Arguments to pass to the tool",
					Required:    false,
				},
			},
		},
		{
			Name:        "version",
			Description: "Get tool version",
			Parameters:  []plugins.PluginParameter{},
		},
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	return nil, nil
}

func (p *{{.plugin_name_camel}}Plugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	if p.tool.IsAvailable() {
		return plugins.PluginHealth{
			Status:  "healthy",
			Message: "Tool integration plugin is running normally",
		}
	}
	return plugins.PluginHealth{
		Status:  "unhealthy",
		Message: "External tool is not available",
	}
}

func (p *{{.plugin_name_camel}}Plugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{}
}

func (p *{{.plugin_name_camel}}Plugin) GetStatus() plugins.PluginStatus {
	return plugins.PluginStatus{
		State:   3, // StateRunning
		Message: "Running",
	}
}
`

const toolIntegrationHelperTemplate = `package main

import (
	"fmt"
	"os/exec"
	"strings"
)

type ExternalTool struct {
	toolPath string
}

func NewExternalTool() *ExternalTool {
	return &ExternalTool{
		toolPath: "{{.tool_name}}",
	}
}

func (t *ExternalTool) Initialize() error {
	// Check if tool is available
	if !t.IsAvailable() {
		return fmt.Errorf("tool {{.tool_name}} is not available")
	}
	return nil
}

func (t *ExternalTool) IsAvailable() bool {
	_, err := exec.LookPath(t.toolPath)
	return err == nil
}

func (t *ExternalTool) Run(args []string) (string, error) {
	cmd := exec.Command(t.toolPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}
	return string(output), nil
}

func (t *ExternalTool) GetVersion() (string, error) {
	output, err := t.Run([]string{"--version"})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}
`

const toolIntegrationModTemplate = `module {{.plugin_name}}

go 1.21

require (
	nix-ai-help v0.0.0
)

replace nix-ai-help => ../../../
`

const toolIntegrationReadmeTemplate = `# {{.plugin_name_title}} Plugin

{{.description}}

This plugin integrates {{.tool_name}} with nixai.

## Features

- {{.tool_name}} tool integration
- Command execution
- Version management
- Health monitoring

## Prerequisites

Make sure {{.tool_name}} is installed and available in PATH:
` + "```" + `bash
which {{.tool_name}}
` + "```" + `

## Installation

` + "```" + `bash
go build -buildmode=plugin -o {{.plugin_name}}.so .
nixai plugin install {{.plugin_name}}.so
` + "```" + `

## Usage

` + "```" + `bash
# Run the tool
nixai plugin execute {{.plugin_name}} run --params '{"args": ["--help"]}'

# Get tool version
nixai plugin execute {{.plugin_name}} version
` + "```" + `

## Author

{{.author}}
`
