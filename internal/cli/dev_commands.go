package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"nix-ai-help/internal/dev"
	"nix-ai-help/internal/dev/dependency"
	"nix-ai-help/internal/dev/environment"
	"nix-ai-help/internal/dev/ide"
	"nix-ai-help/internal/dev/template"
	"nix-ai-help/pkg/logger"
)

// CreateDevCommands creates the dev command with all subcommands
func CreateDevCommands() *cobra.Command {
	devCmd := &cobra.Command{
		Use:   "dev",
		Short: "Developer experience revolution - intelligent development environment management",
		Long: `Phase 3.3 Developer Experience Revolution

Intelligent development environment templates and one-command project setup capabilities.
Create, manage, and optimize development environments with AI-powered templates,
automatic dependency detection, IDE integration, and CI/CD pipeline setup.

Features:
• 🚀 One-command project setup for any language/framework
• 🧠 Intelligent development environment templates  
• 📦 Automatic dependency detection and management
• 🔧 IDE integration (VS Code, Neovim, Vim, Emacs, IntelliJ)
• 🐳 Container and service orchestration
• 🔄 CI/CD pipeline integration
• 📝 Project scaffolding with best practices

Examples:
  nixai dev setup --language rust --editor vscode --containers docker
  nixai dev template list
  nixai dev env create my-web-app --template go-web-api
  nixai dev deps detect ./my-project`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands
	devCmd.AddCommand(createDevSetupCommand())
	devCmd.AddCommand(createDevEnvCommands())
	devCmd.AddCommand(createDevTemplateCommands())
	devCmd.AddCommand(createDevDepsCommands())
	devCmd.AddCommand(createDevIDECommands())

	return devCmd
}

// createDevSetupCommand creates the dev setup command for one-command project setup
func createDevSetupCommand() *cobra.Command {
	var (
		language    string
		framework   string
		editor      string
		containers  []string
		services    []string
		template    string
		interactive bool
		autoDetect  bool
		path        string
	)

	cmd := &cobra.Command{
		Use:   "setup [project-name]",
		Short: "One-command development environment setup",
		Long: `Create a complete development environment with intelligent templates.

This command sets up a full development environment including:
• Project scaffolding with language-specific templates
• IDE configuration and integration
• Container and service setup
• Dependency detection and management
• CI/CD pipeline configuration
• Best practice project structure

Examples:
  nixai dev setup my-web-app --language go --framework gin --editor vscode --containers postgresql
  nixai dev setup frontend-app --language typescript --framework react --editor neovim
  nixai dev setup api-service --template rust-web-server --editor vscode
  nixai dev setup --interactive  # Interactive setup wizard`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			
			// Get project name
			var projectName string
			if len(args) > 0 {
				projectName = args[0]
			} else if interactive {
				fmt.Print("Enter project name: ")
				fmt.Scanln(&projectName)
			} else {
				return fmt.Errorf("project name is required")
			}

			// Interactive setup if requested
			if interactive {
				if err := runInteractiveSetup(&language, &framework, &editor, &containers, &services, &template); err != nil {
					return fmt.Errorf("interactive setup failed: %w", err)
				}
			}

			// Validate required fields
			if language == "" && template == "" {
				return fmt.Errorf("either --language or --template must be specified")
			}

			// Set up environment manager
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			// Create project setup configuration
			setup := &dev.ProjectSetup{
				Name:        projectName,
				Path:        path,
				Language:    language,
				Framework:   framework,
				Editor:      editor,
				Containers:  containers,
				Services:    services,
				Template:    template,
				AutoDetect:  autoDetect,
				Interactive: interactive,
				Config:      make(map[string]interface{}),
			}

			// Set default path if not specified
			if setup.Path == "" {
				currentDir, _ := os.Getwd()
				setup.Path = currentDir
			}

			fmt.Printf("🚀 Setting up development environment: %s\n", projectName)
			fmt.Printf("📍 Language: %s\n", language)
			if framework != "" {
				fmt.Printf("🔧 Framework: %s\n", framework)
			}
			if editor != "" {
				fmt.Printf("💻 Editor: %s\n", editor)
			}
			if len(containers) > 0 {
				fmt.Printf("🐳 Containers: %s\n", strings.Join(containers, ", "))
			}

			// Create environment
			env, err := envManager.CreateEnvironment(context.Background(), setup)
			if err != nil {
				return fmt.Errorf("failed to create development environment: %w", err)
			}

			fmt.Printf("\n✅ Development environment created successfully!\n")
			fmt.Printf("📂 Location: %s\n", env.Path)
			fmt.Printf("🆔 Environment ID: %s\n", env.ID)
			
			// Print next steps
			fmt.Printf("\n📋 Next steps:\n")
			fmt.Printf("1. cd %s\n", env.Path)
			if template != "" {
				fmt.Printf("2. nix develop  # Enter development shell\n")
				switch language {
				case "go":
					fmt.Printf("3. go mod tidy  # Install dependencies\n")
					fmt.Printf("4. go run main.go  # Run the application\n")
				case "rust":
					fmt.Printf("3. cargo build  # Build the project\n")
					fmt.Printf("4. cargo run  # Run the application\n")
				case "python":
					fmt.Printf("3. pip install -r requirements.txt  # Install dependencies\n")
					fmt.Printf("4. python main.py  # Run the application\n")
				case "typescript", "javascript":
					fmt.Printf("3. npm install  # Install dependencies\n")
					fmt.Printf("4. npm run dev  # Start development server\n")
				}
			} else {
				fmt.Printf("2. Start developing with your preferred tools\n")
			}
			
			if editor != "" {
				fmt.Printf("\n💡 IDE integration configured for %s\n", editor)
				if editor == "vscode" {
					fmt.Printf("   Open with: code %s\n", env.Path)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&language, "language", "", "Programming language (go, rust, python, typescript, javascript, etc.)")
	cmd.Flags().StringVar(&framework, "framework", "", "Framework (gin, axum, fastapi, react, vue, etc.)")
	cmd.Flags().StringVar(&editor, "editor", "", "Editor/IDE (vscode, neovim, vim, emacs, intellij)")
	cmd.Flags().StringSliceVar(&containers, "containers", nil, "Containers to include (docker, postgresql, mysql, redis)")
	cmd.Flags().StringSliceVar(&services, "services", nil, "Services to configure")
	cmd.Flags().StringVar(&template, "template", "", "Template to use (overrides language/framework)")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive setup wizard")
	cmd.Flags().BoolVar(&autoDetect, "auto-detect", true, "Automatically detect dependencies")
	cmd.Flags().StringVar(&path, "path", "", "Project path (default: current directory)")

	return cmd
}

// createDevEnvCommands creates environment management commands
func createDevEnvCommands() *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Development environment management",
		Long:  "Create, list, and manage development environments",
	}

	envCmd.AddCommand(createDevEnvCreateCommand())
	envCmd.AddCommand(createDevEnvListCommand())
	envCmd.AddCommand(createDevEnvShowCommand())
	envCmd.AddCommand(createDevEnvDeleteCommand())
	envCmd.AddCommand(createDevEnvStartCommand())
	envCmd.AddCommand(createDevEnvStopCommand())

	return envCmd
}

// createDevEnvCreateCommand creates the env create command
func createDevEnvCreateCommand() *cobra.Command {
	var (
		language   string
		framework  string
		editor     string
		containers []string
		template   string
		path       string
	)

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new development environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			setup := &dev.ProjectSetup{
				Name:       args[0],
				Path:       path,
				Language:   language,
				Framework:  framework,
				Editor:     editor,
				Containers: containers,
				Template:   template,
				AutoDetect: true,
				Config:     make(map[string]interface{}),
			}

			if setup.Path == "" {
				currentDir, _ := os.Getwd()
				setup.Path = currentDir
			}

			env, err := envManager.CreateEnvironment(context.Background(), setup)
			if err != nil {
				return err
			}

			fmt.Printf("✅ Environment created: %s (ID: %s)\n", env.Name, env.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&language, "language", "", "Programming language")
	cmd.Flags().StringVar(&framework, "framework", "", "Framework")
	cmd.Flags().StringVar(&editor, "editor", "", "Editor/IDE")
	cmd.Flags().StringSliceVar(&containers, "containers", nil, "Containers")
	cmd.Flags().StringVar(&template, "template", "", "Template to use")
	cmd.Flags().StringVar(&path, "path", "", "Project path")

	return cmd
}

// createDevEnvListCommand creates the env list command
func createDevEnvListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all development environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			environments, err := envManager.ListEnvironments(context.Background())
			if err != nil {
				return err
			}

			if len(environments) == 0 {
				fmt.Println("No development environments found.")
				return nil
			}

			fmt.Printf("📦 Development Environments (%d)\n\n", len(environments))
			for _, env := range environments {
				status := "●"
				switch env.Status {
				case dev.DevEnvironmentStatusReady:
					status = "🟢"
				case dev.DevEnvironmentStatusCreating:
					status = "🟡"
				case dev.DevEnvironmentStatusFailed:
					status = "🔴"
				case dev.DevEnvironmentStatusStopped:
					status = "⚪"
				}

				fmt.Printf("%s %s\n", status, env.Name)
				fmt.Printf("   ID: %s\n", env.ID)
				fmt.Printf("   Language: %s", env.Language)
				if env.Framework != "" {
					fmt.Printf(" (%s)", env.Framework)
				}
				fmt.Printf("\n")
				fmt.Printf("   Path: %s\n", env.Path)
				fmt.Printf("   Created: %s\n", env.CreatedAt.Format("2006-01-02 15:04:05"))
				if len(env.Dependencies) > 0 {
					fmt.Printf("   Dependencies: %d\n", len(env.Dependencies))
				}
				fmt.Println()
			}

			return nil
		},
	}
}

// createDevEnvShowCommand creates the env show command
func createDevEnvShowCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show [environment-id]",
		Short: "Show environment details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			env, err := envManager.GetEnvironment(context.Background(), args[0])
			if err != nil {
				return err
			}

			if format == "json" {
				data, err := json.MarshalIndent(env, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("🏗️  Development Environment: %s\n\n", env.Name)
			fmt.Printf("ID: %s\n", env.ID)
			fmt.Printf("Status: %s\n", env.Status)
			fmt.Printf("Language: %s\n", env.Language)
			if env.Framework != "" {
				fmt.Printf("Framework: %s\n", env.Framework)
			}
			if env.Editor != "" {
				fmt.Printf("Editor: %s\n", env.Editor)
			}
			fmt.Printf("Path: %s\n", env.Path)
			fmt.Printf("Template: %s\n", env.Template)
			fmt.Printf("Created: %s\n", env.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("Updated: %s\n", env.UpdatedAt.Format("2006-01-02 15:04:05"))

			if len(env.Containers) > 0 {
				fmt.Printf("\n🐳 Containers:\n")
				for _, container := range env.Containers {
					fmt.Printf("  • %s\n", container)
				}
			}

			if len(env.Services) > 0 {
				fmt.Printf("\n🔧 Services:\n")
				for _, service := range env.Services {
					fmt.Printf("  • %s\n", service)
				}
			}

			if len(env.Dependencies) > 0 {
				fmt.Printf("\n📦 Dependencies (%d):\n", len(env.Dependencies))
				depsByType := make(map[string][]dev.Dependency)
				for _, dep := range env.Dependencies {
					depsByType[dep.Type] = append(depsByType[dep.Type], dep)
				}

				for depType, deps := range depsByType {
					fmt.Printf("\n  %s:\n", strings.Title(depType))
					for _, dep := range deps {
						required := ""
						if dep.Required {
							required = " (required)"
						}
						fmt.Printf("    • %s %s%s\n", dep.Name, dep.Version, required)
					}
				}
			}

			if len(env.Config) > 0 {
				fmt.Printf("\n⚙️  Configuration:\n")
				for key, value := range env.Config {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json)")
	return cmd
}

// createDevEnvDeleteCommand creates the env delete command
func createDevEnvDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [environment-id]",
		Short: "Delete a development environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			if !force {
				fmt.Printf("Are you sure you want to delete environment %s? (y/N): ", args[0])
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" {
					fmt.Println("Deletion cancelled.")
					return nil
				}
			}

			if err := envManager.DeleteEnvironment(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Printf("✅ Environment deleted: %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion without confirmation")
	return cmd
}

// createDevEnvStartCommand creates the env start command
func createDevEnvStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start [environment-id]",
		Short: "Start a development environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			if err := envManager.StartEnvironment(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Printf("✅ Environment started: %s\n", args[0])
			return nil
		},
	}
}

// createDevEnvStopCommand creates the env stop command
func createDevEnvStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [environment-id]",
		Short: "Stop a development environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)

			if err := envManager.StopEnvironment(context.Background(), args[0]); err != nil {
				return err
			}

			fmt.Printf("✅ Environment stopped: %s\n", args[0])
			return nil
		},
	}
}

// createDevTemplateCommands creates template management commands
func createDevTemplateCommands() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Development template management",
		Long:  "Create, list, and manage development environment templates",
	}

	templateCmd.AddCommand(createDevTemplateListCommand())
	templateCmd.AddCommand(createDevTemplateShowCommand())
	templateCmd.AddCommand(createDevTemplateCreateCommand())
	templateCmd.AddCommand(createDevTemplateSearchCommand())

	return templateCmd
}

// createDevTemplateListCommand creates the template list command
func createDevTemplateListCommand() *cobra.Command {
	var category string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available development templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			templatesDir := filepath.Join(homeDir, ".nixai", "templates")
			templateManager := template.NewManager(templatesDir, log)

			// Initialize built-in templates
			if err := templateManager.InitializeBuiltinTemplates(context.Background()); err != nil {
				return fmt.Errorf("failed to initialize built-in templates: %w", err)
			}

			templates, err := templateManager.ListTemplates(context.Background())
			if err != nil {
				return err
			}

			// Filter by category if specified
			if category != "" {
				var filteredTemplates []*dev.DevTemplate
				for _, template := range templates {
					if template.Category == category {
						filteredTemplates = append(filteredTemplates, template)
					}
				}
				templates = filteredTemplates
			}

			if len(templates) == 0 {
				fmt.Println("No templates found.")
				return nil
			}

			fmt.Printf("📋 Development Templates (%d)\n\n", len(templates))

			// Group by category
			templatesByCategory := make(map[string][]*dev.DevTemplate)
			for _, template := range templates {
				templatesByCategory[template.Category] = append(templatesByCategory[template.Category], template)
			}

			for cat, catTemplates := range templatesByCategory {
				fmt.Printf("📂 %s\n", strings.Title(cat))
				for _, template := range catTemplates {
					fmt.Printf("  🔧 %s (%s)\n", template.Name, template.ID)
					fmt.Printf("      Language: %s", template.Language)
					if template.Framework != "" {
						fmt.Printf(" + %s", template.Framework)
					}
					fmt.Printf("\n")
					fmt.Printf("      %s\n", template.Description)
					if len(template.Tags) > 0 {
						fmt.Printf("      Tags: %s\n", strings.Join(template.Tags, ", "))
					}
					fmt.Printf("      Author: %s (v%s)\n", template.Author, template.Version)
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category (web, frontend, backend, etc.)")
	return cmd
}

// createDevTemplateShowCommand creates the template show command
func createDevTemplateShowCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show [template-id]",
		Short: "Show template details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			templatesDir := filepath.Join(homeDir, ".nixai", "templates")
			templateManager := template.NewManager(templatesDir, log)

			template, err := templateManager.GetTemplate(context.Background(), args[0])
			if err != nil {
				return err
			}

			if format == "json" {
				data, err := json.MarshalIndent(template, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Printf("📋 Template: %s\n\n", template.Name)
			fmt.Printf("ID: %s\n", template.ID)
			fmt.Printf("Language: %s\n", template.Language)
			if template.Framework != "" {
				fmt.Printf("Framework: %s\n", template.Framework)
			}
			fmt.Printf("Category: %s\n", template.Category)
			fmt.Printf("Description: %s\n", template.Description)
			fmt.Printf("Author: %s\n", template.Author)
			fmt.Printf("Version: %s\n", template.Version)
			fmt.Printf("Created: %s\n", template.CreatedAt.Format("2006-01-02 15:04:05"))

			if len(template.Tags) > 0 {
				fmt.Printf("\n🏷️  Tags:\n")
				for _, tag := range template.Tags {
					fmt.Printf("  • %s\n", tag)
				}
			}

			if len(template.Dependencies) > 0 {
				fmt.Printf("\n📦 Dependencies:\n")
				for _, dep := range template.Dependencies {
					required := ""
					if dep.Required {
						required = " (required)"
					}
					fmt.Printf("  • %s %s%s\n", dep.Name, dep.Version, required)
				}
			}

			if len(template.Files) > 0 {
				fmt.Printf("\n📄 Files:\n")
				for _, file := range template.Files {
					templateFlag := ""
					if file.Template {
						templateFlag = " (template)"
					}
					fmt.Printf("  • %s%s\n", file.Path, templateFlag)
				}
			}

			if len(template.Commands) > 0 {
				fmt.Printf("\n⚡ Commands:\n")
				for _, command := range template.Commands {
					fmt.Printf("  • %s\n", command)
				}
			}

			if len(template.Config) > 0 {
				fmt.Printf("\n⚙️  Configuration:\n")
				for key, value := range template.Config {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json)")
	return cmd
}

// createDevTemplateCreateCommand creates the template create command
func createDevTemplateCreateCommand() *cobra.Command {
	var (
		name        string
		description string
		language    string
		framework   string
		category    string
		tags        []string
	)

	cmd := &cobra.Command{
		Use:   "create [template-id]",
		Short: "Create a new development template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			templatesDir := filepath.Join(homeDir, ".nixai", "templates")
			templateManager := template.NewManager(templatesDir, log)

			template := &dev.DevTemplate{
				ID:          args[0],
				Name:        name,
				Description: description,
				Language:    language,
				Framework:   framework,
				Category:    category,
				Tags:        tags,
				Config:      make(map[string]interface{}),
				Files:       []dev.TemplateFile{},
				Dependencies: []dev.Dependency{},
				Commands:    []string{},
				Author:      "user",
				Version:     "1.0.0",
			}

			if err := templateManager.CreateTemplate(context.Background(), template); err != nil {
				return err
			}

			fmt.Printf("✅ Template created: %s\n", template.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Template name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Template description (required)")
	cmd.Flags().StringVar(&language, "language", "", "Programming language (required)")
	cmd.Flags().StringVar(&framework, "framework", "", "Framework")
	cmd.Flags().StringVar(&category, "category", "", "Template category (required)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Template tags")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("description")
	cmd.MarkFlagRequired("language")
	cmd.MarkFlagRequired("category")

	return cmd
}

// createDevTemplateSearchCommand creates the template search command
func createDevTemplateSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search development templates",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			templatesDir := filepath.Join(homeDir, ".nixai", "templates")
			templateManager := template.NewManager(templatesDir, log)

			templates, err := templateManager.SearchTemplates(context.Background(), args[0])
			if err != nil {
				return err
			}

			if len(templates) == 0 {
				fmt.Printf("No templates found matching '%s'.\n", args[0])
				return nil
			}

			fmt.Printf("🔍 Found %d template(s) matching '%s'\n\n", len(templates), args[0])
			for _, template := range templates {
				fmt.Printf("🔧 %s (%s)\n", template.Name, template.ID)
				fmt.Printf("   Language: %s", template.Language)
				if template.Framework != "" {
					fmt.Printf(" + %s", template.Framework)
				}
				fmt.Printf("\n")
				fmt.Printf("   %s\n", template.Description)
				fmt.Printf("   Category: %s\n", template.Category)
				fmt.Println()
			}

			return nil
		},
	}
}

// createDevDepsCommands creates dependency management commands
func createDevDepsCommands() *cobra.Command {
	depsCmd := &cobra.Command{
		Use:   "deps",
		Short: "Dependency detection and management",
		Long:  "Automatically detect and manage project dependencies",
	}

	depsCmd.AddCommand(createDevDepsDetectCommand())
	depsCmd.AddCommand(createDevDepsInstallCommand())
	depsCmd.AddCommand(createDevDepsUpdateCommand())
	depsCmd.AddCommand(createDevDepsCheckCommand())

	return depsCmd
}

// createDevDepsDetectCommand creates the deps detect command
func createDevDepsDetectCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "detect [path]",
		Short: "Detect project dependencies",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			depManager := dependency.NewManager(log)

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			dependencies, err := depManager.DetectDependencies(context.Background(), path)
			if err != nil {
				return err
			}

			if format == "json" {
				data, err := json.MarshalIndent(dependencies, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}

			if len(dependencies) == 0 {
				fmt.Println("No dependencies detected.")
				return nil
			}

			fmt.Printf("📦 Detected Dependencies (%d)\n\n", len(dependencies))

			// Group by type
			depsByType := make(map[string][]dev.Dependency)
			for _, dep := range dependencies {
				depsByType[dep.Type] = append(depsByType[dep.Type], dep)
			}

			for depType, deps := range depsByType {
				fmt.Printf("📋 %s Dependencies (%d):\n", strings.Title(depType), len(deps))
				for _, dep := range deps {
					required := ""
					if dep.Required {
						required = " (required)"
					}
					fmt.Printf("  • %s %s%s\n", dep.Name, dep.Version, required)
					if dep.Source != "" {
						fmt.Printf("    Source: %s\n", dep.Source)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "text", "Output format (text, json)")
	return cmd
}

// createDevDepsInstallCommand creates the deps install command
func createDevDepsInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install [path]",
		Short: "Install project dependencies",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			depManager := dependency.NewManager(log)

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			// Detect dependencies first
			dependencies, err := depManager.DetectDependencies(context.Background(), path)
			if err != nil {
				return err
			}

			if len(dependencies) == 0 {
				fmt.Println("No dependencies detected.")
				return nil
			}

			fmt.Printf("📦 Installing %d dependencies...\n", len(dependencies))

			// Install dependencies
			if err := depManager.InstallDependencies(context.Background(), path, dependencies); err != nil {
				return err
			}

			fmt.Println("✅ Dependencies installed successfully!")
			return nil
		},
	}
}

// createDevDepsUpdateCommand creates the deps update command
func createDevDepsUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update [path]",
		Short: "Update project dependencies",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			depManager := dependency.NewManager(log)

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			fmt.Println("🔄 Updating dependencies...")

			if err := depManager.UpdateDependencies(context.Background(), path); err != nil {
				return err
			}

			fmt.Println("✅ Dependencies updated successfully!")
			return nil
		},
	}
}

// createDevDepsCheckCommand creates the deps check command
func createDevDepsCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check [path]",
		Short: "Check for outdated dependencies",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			depManager := dependency.NewManager(log)

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			dependencies, err := depManager.CheckDependencies(context.Background(), path)
			if err != nil {
				return err
			}

			fmt.Printf("🔍 Checked %d dependencies\n", len(dependencies))
			fmt.Println("ℹ️  Dependency update checking requires additional implementation")
			return nil
		},
	}
}

// createDevIDECommands creates IDE integration commands
func createDevIDECommands() *cobra.Command {
	ideCmd := &cobra.Command{
		Use:   "ide",
		Short: "IDE integration management",
		Long:  "Configure and manage IDE integrations",
	}

	ideCmd.AddCommand(createDevIDEListCommand())
	ideCmd.AddCommand(createDevIDESetupCommand())
	ideCmd.AddCommand(createDevIDEConfigCommand())

	return ideCmd
}

// createDevIDEListCommand creates the ide list command
func createDevIDEListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List supported IDEs",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			ideManager := ide.NewManager(log)

			ides, err := ideManager.ListSupportedIDEs(context.Background())
			if err != nil {
				return err
			}

			fmt.Printf("💻 Supported IDEs (%d)\n\n", len(ides))
			for _, ideEntry := range ides {
				config, err := ideManager.GetIDEConfig(context.Background(), ideEntry)
				if err != nil {
					continue
				}

				fmt.Printf("🔧 %s (%s)\n", config.Name, ideEntry)
				fmt.Printf("   Type: %s\n", config.Type)
				if len(config.Extensions) > 0 {
					fmt.Printf("   Extensions: %d available\n", len(config.Extensions))
				}
				fmt.Println()
			}

			return nil
		},
	}
}

// createDevIDESetupCommand creates the ide setup command
func createDevIDESetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup [environment-id] [ide]",
		Short: "Setup IDE integration for environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			homeDir, _ := os.UserHomeDir()
			environmentsDir := filepath.Join(homeDir, ".nixai", "environments")
			envManager := environment.NewManager(environmentsDir, log)
			ideManager := ide.NewManager(log)

			env, err := envManager.GetEnvironment(context.Background(), args[0])
			if err != nil {
				return err
			}

			if err := ideManager.SetupIDE(context.Background(), env, args[1]); err != nil {
				return err
			}

			fmt.Printf("✅ IDE integration setup completed: %s for %s\n", args[1], env.Name)
			return nil
		},
	}
}

// createDevIDEConfigCommand creates the ide config command
func createDevIDEConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config [ide]",
		Short: "Show IDE configuration details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			ideManager := ide.NewManager(log)

			config, err := ideManager.GetIDEConfig(context.Background(), args[0])
			if err != nil {
				return err
			}

			fmt.Printf("💻 IDE Configuration: %s\n\n", config.Name)
			fmt.Printf("Type: %s\n", config.Type)

			if len(config.ConfigFiles) > 0 {
				fmt.Printf("\n📄 Configuration Files:\n")
				for _, file := range config.ConfigFiles {
					fmt.Printf("  • %s\n", file)
				}
			}

			if len(config.Extensions) > 0 {
				fmt.Printf("\n🔌 Recommended Extensions:\n")
				for _, ext := range config.Extensions {
					fmt.Printf("  • %s\n", ext)
				}
			}

			if len(config.Settings) > 0 {
				fmt.Printf("\n⚙️  Default Settings:\n")
				for key, value := range config.Settings {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}

			return nil
		},
	}
}

// runInteractiveSetup runs interactive setup wizard
func runInteractiveSetup(language, framework, editor *string, containers, services *[]string, template *string) error {
	fmt.Println("🎯 Interactive Development Environment Setup")
	fmt.Println()

	// Language selection
	fmt.Println("Select programming language:")
	fmt.Println("1. Go")
	fmt.Println("2. Rust") 
	fmt.Println("3. Python")
	fmt.Println("4. TypeScript")
	fmt.Println("5. JavaScript")
	fmt.Println("6. Other")
	fmt.Print("Choice (1-6): ")
	
	var choice int
	fmt.Scanln(&choice)
	
	switch choice {
	case 1:
		*language = "go"
	case 2:
		*language = "rust"
	case 3:
		*language = "python"
	case 4:
		*language = "typescript"
	case 5:
		*language = "javascript"
	case 6:
		fmt.Print("Enter language: ")
		fmt.Scanln(language)
	default:
		*language = "go"
	}

	// Framework selection
	if *language != "" {
		fmt.Printf("\nSelect framework for %s (optional, press Enter to skip): ", *language)
		fmt.Scanln(framework)
	}

	// Editor selection
	fmt.Println("\nSelect editor/IDE:")
	fmt.Println("1. VS Code")
	fmt.Println("2. Neovim")
	fmt.Println("3. Vim")
	fmt.Println("4. Emacs")
	fmt.Println("5. IntelliJ IDEA")
	fmt.Println("6. Skip")
	fmt.Print("Choice (1-6): ")
	
	fmt.Scanln(&choice)
	
	switch choice {
	case 1:
		*editor = "vscode"
	case 2:
		*editor = "neovim"
	case 3:
		*editor = "vim"
	case 4:
		*editor = "emacs"
	case 5:
		*editor = "intellij"
	case 6:
		*editor = ""
	default:
		*editor = "vscode"
	}

	// Container selection
	fmt.Println("\nSelect containers (comma-separated, optional):")
	fmt.Println("Options: docker, postgresql, mysql, redis")
	fmt.Print("Containers: ")
	
	var containerInput string
	fmt.Scanln(&containerInput)
	if containerInput != "" {
		*containers = strings.Split(containerInput, ",")
		for i, container := range *containers {
			(*containers)[i] = strings.TrimSpace(container)
		}
	}

	fmt.Println("\n✅ Setup configuration complete!")
	return nil
}