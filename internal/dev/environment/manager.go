package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/dev"
	"nix-ai-help/internal/dev/dependency"
	"nix-ai-help/internal/dev/ide"
	"nix-ai-help/internal/dev/template"
	"nix-ai-help/pkg/logger"
)

// Manager implements the DevEnvironmentManager interface
type Manager struct {
	logger            *logger.Logger
	environmentsDir   string
	templateManager   *template.Manager
	dependencyManager *dependency.Manager
	ideManager        *ide.Manager
}

// NewManager creates a new environment manager
func NewManager(environmentsDir string, logger *logger.Logger) *Manager {
	return &Manager{
		logger:            logger,
		environmentsDir:   environmentsDir,
		templateManager:   template.NewManager(filepath.Join(environmentsDir, "templates"), logger),
		dependencyManager: dependency.NewManager(logger),
		ideManager:        ide.NewManager(logger),
	}
}

// CreateEnvironment creates a new development environment
func (m *Manager) CreateEnvironment(ctx context.Context, setup *dev.ProjectSetup) (*dev.DevEnvironment, error) {
	// Generate unique environment ID
	envID := m.generateEnvironmentID(setup.Name)
	
	// Create environment directory
	envPath := filepath.Join(setup.Path, setup.Name)
	if err := os.MkdirAll(envPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create environment directory: %w", err)
	}
	
	// Create development environment
	env := &dev.DevEnvironment{
		ID:        envID,
		Name:      setup.Name,
		Language:  setup.Language,
		Framework: setup.Framework,
		Editor:    setup.Editor,
		Containers: setup.Containers,
		Services:  setup.Services,
		Template:  setup.Template,
		Config:    setup.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Status:    dev.DevEnvironmentStatusCreating,
		Path:      envPath,
	}
	
	// Apply template if specified
	if setup.Template != "" {
		if err := m.applyTemplate(ctx, env, setup.Template, setup.Config); err != nil {
			return nil, fmt.Errorf("failed to apply template: %w", err)
		}
	}
	
	// Auto-detect dependencies if enabled
	if setup.AutoDetect {
		deps, err := m.dependencyManager.DetectDependencies(ctx, envPath)
		if err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to auto-detect dependencies: %v", err))
		} else {
			env.Dependencies = deps
		}
	}
	
	// Setup IDE integration
	if setup.Editor != "" {
		if err := m.ideManager.SetupIDE(ctx, env, setup.Editor); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to setup IDE integration for %s: %v", setup.Editor, err))
		}
	}
	
	// Create environment configuration file
	if err := m.saveEnvironmentConfig(env); err != nil {
		return nil, fmt.Errorf("failed to save environment config: %w", err)
	}
	
	// Generate additional project files
	if err := m.generateProjectFiles(ctx, env); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to generate some project files: %v", err))
	}
	
	env.Status = dev.DevEnvironmentStatusReady
	env.UpdatedAt = time.Now()
	
	// Update environment configuration
	if err := m.saveEnvironmentConfig(env); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to update environment config: %v", err))
	}
	
	m.logger.Info(fmt.Sprintf("Development environment created successfully: %s (Language: %s, Framework: %s, Path: %s)", env.Name, env.Language, env.Framework, env.Path))
	
	return env, nil
}

// GetEnvironment retrieves a development environment by ID
func (m *Manager) GetEnvironment(ctx context.Context, id string) (*dev.DevEnvironment, error) {
	configPath := filepath.Join(m.environmentsDir, id+".json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("environment %s not found", id)
		}
		return nil, fmt.Errorf("failed to read environment config: %w", err)
	}
	
	var env dev.DevEnvironment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to parse environment config: %w", err)
	}
	
	return &env, nil
}

// ListEnvironments lists all development environments
func (m *Manager) ListEnvironments(ctx context.Context) ([]*dev.DevEnvironment, error) {
	if err := os.MkdirAll(m.environmentsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create environments directory: %w", err)
	}
	
	files, err := os.ReadDir(m.environmentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}
	
	var environments []*dev.DevEnvironment
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			envID := strings.TrimSuffix(file.Name(), ".json")
			env, err := m.GetEnvironment(ctx, envID)
			if err != nil {
				m.logger.Warn(fmt.Sprintf("Failed to load environment %s: %v", envID, err))
				continue
			}
			environments = append(environments, env)
		}
	}
	
	return environments, nil
}

// UpdateEnvironment updates a development environment
func (m *Manager) UpdateEnvironment(ctx context.Context, env *dev.DevEnvironment) error {
	env.UpdatedAt = time.Now()
	return m.saveEnvironmentConfig(env)
}

// DeleteEnvironment deletes a development environment
func (m *Manager) DeleteEnvironment(ctx context.Context, id string) error {
	env, err := m.GetEnvironment(ctx, id)
	if err != nil {
		return err
	}
	
	// Remove environment directory
	if err := os.RemoveAll(env.Path); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to remove environment directory %s: %v", env.Path, err))
	}
	
	// Remove environment configuration
	configPath := filepath.Join(m.environmentsDir, id+".json")
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("failed to remove environment config: %w", err)
	}
	
	m.logger.Info(fmt.Sprintf("Development environment deleted: %s (ID: %s)", env.Name, id))
	return nil
}

// StartEnvironment starts a development environment
func (m *Manager) StartEnvironment(ctx context.Context, id string) error {
	env, err := m.GetEnvironment(ctx, id)
	if err != nil {
		return err
	}
	
	// Start services if any
	for _, service := range env.Services {
		if err := m.startService(ctx, env, service); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to start service %s: %v", service, err))
		}
	}
	
	// Start containers if any
	for _, container := range env.Containers {
		if err := m.startContainer(ctx, env, container); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to start container %s: %v", container, err))
		}
	}
	
	m.logger.Info(fmt.Sprintf("Development environment started: %s", env.Name))
	return nil
}

// StopEnvironment stops a development environment
func (m *Manager) StopEnvironment(ctx context.Context, id string) error {
	env, err := m.GetEnvironment(ctx, id)
	if err != nil {
		return err
	}
	
	// Stop containers
	for _, container := range env.Containers {
		if err := m.stopContainer(ctx, env, container); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to stop container %s: %v", container, err))
		}
	}
	
	// Stop services
	for _, service := range env.Services {
		if err := m.stopService(ctx, env, service); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to stop service %s: %v", service, err))
		}
	}
	
	env.Status = dev.DevEnvironmentStatusStopped
	env.UpdatedAt = time.Now()
	
	if err := m.saveEnvironmentConfig(env); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to update environment status: %v", err))
	}
	
	m.logger.Info(fmt.Sprintf("Development environment stopped: %s", env.Name))
	return nil
}

// applyTemplate applies a template to the environment
func (m *Manager) applyTemplate(ctx context.Context, env *dev.DevEnvironment, templateID string, config map[string]interface{}) error {
	template, err := m.templateManager.GetTemplate(ctx, templateID)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}
	
	// Merge template config with provided config
	if config == nil {
		config = make(map[string]interface{})
	}
	for key, value := range template.Config {
		if _, exists := config[key]; !exists {
			config[key] = value
		}
	}
	env.Config = config
	
	// Create template files
	for _, file := range template.Files {
		filePath := filepath.Join(env.Path, file.Path)
		
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", file.Path, err)
		}
		
		content := file.Content
		if file.Template {
			// Apply template variables
			content = m.applyTemplateVariables(content, env)
		}
		
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create file %s: %w", file.Path, err)
		}
	}
	
	// Add template dependencies
	env.Dependencies = append(env.Dependencies, template.Dependencies...)
	
	m.logger.Info(fmt.Sprintf("Template applied successfully: %s for environment %s", template.Name, env.Name))
	return nil
}

// applyTemplateVariables applies template variables to content
func (m *Manager) applyTemplateVariables(content string, env *dev.DevEnvironment) string {
	// Replace common template variables
	content = strings.ReplaceAll(content, "{{.Name}}", env.Name)
	content = strings.ReplaceAll(content, "{{.Language}}", env.Language)
	content = strings.ReplaceAll(content, "{{.Framework}}", env.Framework)
	
	// Replace config variables
	for key, value := range env.Config {
		placeholder := fmt.Sprintf("{{.Config.%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		content = strings.ReplaceAll(content, placeholder, valueStr)
	}
	
	return content
}

// generateProjectFiles generates additional project files
func (m *Manager) generateProjectFiles(ctx context.Context, env *dev.DevEnvironment) error {
	// Generate README.md
	if err := m.generateReadme(env); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}
	
	// Generate .gitignore
	if err := m.generateGitignore(env); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}
	
	// Generate Makefile if applicable
	if err := m.generateMakefile(env); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to generate Makefile: %v", err))
	}
	
	// Generate Docker files if containers are specified
	if len(env.Containers) > 0 {
		if err := m.generateDockerFiles(env); err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to generate Docker files: %v", err))
		}
	}
	
	return nil
}

// generateReadme generates a README.md file
func (m *Manager) generateReadme(env *dev.DevEnvironment) error {
	content := fmt.Sprintf(`# %s

%s project generated by nixai.

## Language
%s

## Framework
%s

## Getting Started

### Prerequisites
- Nix with flakes enabled
- %s development environment

### Development

1. Enter the development shell:
   '''bash
   nix develop
   '''

2. Install dependencies:
   '''bash
   # Dependencies are managed automatically via Nix
   '''

3. Run the project:
   '''bash
   # See package.json/Makefile/Cargo.toml for available commands
   '''

## Project Structure

- 'src/' - Source code
- 'tests/' - Test files
- 'docs/' - Documentation
- 'flake.nix' - Nix development environment configuration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

This project is generated by nixai and follows standard open source practices.
`, env.Name, strings.Title(env.Language), env.Language, env.Framework, env.Language)
	
	readmePath := filepath.Join(env.Path, "README.md")
	return os.WriteFile(readmePath, []byte(content), 0644)
}

// generateGitignore generates a .gitignore file
func (m *Manager) generateGitignore(env *dev.DevEnvironment) error {
	var gitignoreContent string
	
	// Base gitignore
	gitignoreContent += "# Generated by nixai\n\n"
	gitignoreContent += "# IDE files\n"
	gitignoreContent += ".vscode/\n"
	gitignoreContent += ".idea/\n"
	gitignoreContent += "*.swp\n"
	gitignoreContent += "*.swo\n"
	gitignoreContent += "*~\n\n"
	
	// Language-specific gitignore
	switch env.Language {
	case "go":
		gitignoreContent += "# Go\n"
		gitignoreContent += "*.exe\n"
		gitignoreContent += "*.exe~\n"
		gitignoreContent += "*.dll\n"
		gitignoreContent += "*.so\n"
		gitignoreContent += "*.dylib\n"
		gitignoreContent += "*.test\n"
		gitignoreContent += "*.out\n"
		gitignoreContent += "go.work\n\n"
	case "rust":
		gitignoreContent += "# Rust\n"
		gitignoreContent += "/target/\n"
		gitignoreContent += "**/*.rs.bk\n"
		gitignoreContent += "Cargo.lock\n\n"
	case "python":
		gitignoreContent += "# Python\n"
		gitignoreContent += "__pycache__/\n"
		gitignoreContent += "*.py[cod]\n"
		gitignoreContent += "*$py.class\n"
		gitignoreContent += "*.so\n"
		gitignoreContent += ".Python\n"
		gitignoreContent += "env/\n"
		gitignoreContent += "venv/\n"
		gitignoreContent += ".env\n"
		gitignoreContent += ".venv\n"
		gitignoreContent += "pip-log.txt\n"
		gitignoreContent += "pip-delete-this-directory.txt\n\n"
	case "typescript", "javascript":
		gitignoreContent += "# Node.js\n"
		gitignoreContent += "node_modules/\n"
		gitignoreContent += "npm-debug.log*\n"
		gitignoreContent += "yarn-debug.log*\n"
		gitignoreContent += "yarn-error.log*\n"
		gitignoreContent += ".npm\n"
		gitignoreContent += ".yarn\n"
		gitignoreContent += "dist/\n"
		gitignoreContent += "build/\n\n"
	}
	
	// Additional ignores
	gitignoreContent += "# OS\n"
	gitignoreContent += ".DS_Store\n"
	gitignoreContent += ".DS_Store?\n"
	gitignoreContent += "._*\n"
	gitignoreContent += ".Spotlight-V100\n"
	gitignoreContent += ".Trashes\n"
	gitignoreContent += "ehthumbs.db\n"
	gitignoreContent += "Thumbs.db\n"
	
	gitignorePath := filepath.Join(env.Path, ".gitignore")
	return os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
}

// generateMakefile generates a Makefile
func (m *Manager) generateMakefile(env *dev.DevEnvironment) error {
	var makefileContent string
	
	switch env.Language {
	case "go":
		makefileContent = `# Generated Makefile for ` + env.Name + `

.PHONY: build test clean run install

build:
	go build -v ./...

test:
	go test -v ./...

clean:
	go clean
	rm -f ` + env.Name + `

run:
	go run main.go

install:
	go install

fmt:
	go fmt ./...

vet:
	go vet ./...

mod-tidy:
	go mod tidy

deps:
	go mod download
`
	case "rust":
		makefileContent = `# Generated Makefile for ` + env.Name + `

.PHONY: build test clean run install

build:
	cargo build

test:
	cargo test

clean:
	cargo clean

run:
	cargo run

install:
	cargo install --path .

fmt:
	cargo fmt

check:
	cargo check

clippy:
	cargo clippy
`
	case "python":
		makefileContent = `# Generated Makefile for ` + env.Name + `

.PHONY: test clean run install lint

test:
	python -m pytest

clean:
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -delete

run:
	python main.py

install:
	pip install -r requirements.txt

lint:
	pylint *.py

format:
	black *.py

deps:
	pip install -r requirements.txt
`
	default:
		return nil // Don't generate Makefile for unsupported languages
	}
	
	makefilePath := filepath.Join(env.Path, "Makefile")
	return os.WriteFile(makefilePath, []byte(makefileContent), 0644)
}

// generateDockerFiles generates Docker configuration files
func (m *Manager) generateDockerFiles(env *dev.DevEnvironment) error {
	// Generate Dockerfile
	dockerfileContent := m.generateDockerfile(env)
	dockerfilePath := filepath.Join(env.Path, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}
	
	// Generate docker-compose.yml
	dockerComposeContent := m.generateDockerCompose(env)
	dockerComposePath := filepath.Join(env.Path, "docker-compose.yml")
	if err := os.WriteFile(dockerComposePath, []byte(dockerComposeContent), 0644); err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}
	
	return nil
}

// generateDockerfile generates a Dockerfile
func (m *Manager) generateDockerfile(env *dev.DevEnvironment) string {
	switch env.Language {
	case "go":
		return `# Multi-stage build for Go
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
`
	case "rust":
		return `# Multi-stage build for Rust
FROM rust:1.75 AS builder

WORKDIR /app
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm src/main.rs

COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/` + env.Name + ` /usr/local/bin/` + env.Name + `
CMD ["` + env.Name + `"]
`
	case "python":
		return `FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "main.py"]
`
	case "typescript", "javascript":
		return `FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000
CMD ["npm", "start"]
`
	default:
		return `FROM alpine:latest
WORKDIR /app
COPY . .
CMD ["echo", "Container for ` + env.Name + `"]
`
	}
}

// generateDockerCompose generates a docker-compose.yml
func (m *Manager) generateDockerCompose(env *dev.DevEnvironment) string {
	compose := `version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NODE_ENV=development
    volumes:
      - .:/app
      - /app/node_modules
`
	
	// Add database services if needed
	for _, container := range env.Containers {
		switch container {
		case "postgresql", "postgres":
			compose += `
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: ` + env.Name + `
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
`
		case "mysql":
			compose += `
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_DATABASE: ` + env.Name + `
      MYSQL_ROOT_PASSWORD: mysql
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
`
		case "redis":
			compose += `
  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
`
		}
	}
	
	// Add volumes section if needed
	if strings.Contains(compose, "_data:") {
		compose += `
volumes:
`
		if strings.Contains(compose, "postgres_data:") {
			compose += `  postgres_data:
`
		}
		if strings.Contains(compose, "mysql_data:") {
			compose += `  mysql_data:
`
		}
		if strings.Contains(compose, "redis_data:") {
			compose += `  redis_data:
`
		}
	}
	
	return compose
}

// saveEnvironmentConfig saves environment configuration to file
func (m *Manager) saveEnvironmentConfig(env *dev.DevEnvironment) error {
	if err := os.MkdirAll(m.environmentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create environments directory: %w", err)
	}
	
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment config: %w", err)
	}
	
	configPath := filepath.Join(m.environmentsDir, env.ID+".json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment config: %w", err)
	}
	
	return nil
}

// generateEnvironmentID generates a unique environment ID
func (m *Manager) generateEnvironmentID(name string) string {
	// Convert name to lowercase and replace spaces with hyphens
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	
	// Remove special characters
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// startService starts a service for the environment
func (m *Manager) startService(ctx context.Context, env *dev.DevEnvironment, service string) error {
	m.logger.Info(fmt.Sprintf("Starting service %s for environment %s", service, env.Name))
	// Service startup logic would go here
	return nil
}

// stopService stops a service for the environment
func (m *Manager) stopService(ctx context.Context, env *dev.DevEnvironment, service string) error {
	m.logger.Info(fmt.Sprintf("Stopping service %s for environment %s", service, env.Name))
	// Service stop logic would go here
	return nil
}

// startContainer starts a container for the environment
func (m *Manager) startContainer(ctx context.Context, env *dev.DevEnvironment, container string) error {
	m.logger.Info(fmt.Sprintf("Starting container %s for environment %s", container, env.Name))
	// Container startup logic would go here
	return nil
}

// stopContainer stops a container for the environment
func (m *Manager) stopContainer(ctx context.Context, env *dev.DevEnvironment, container string) error {
	m.logger.Info(fmt.Sprintf("Stopping container %s for environment %s", container, env.Name))
	// Container stop logic would go here
	return nil
}