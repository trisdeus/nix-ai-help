package template

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/dev"
	"nix-ai-help/pkg/logger"
)

// Manager implements the TemplateManager interface
type Manager struct {
	templatesDir string
	logger       *logger.Logger
}

// NewManager creates a new template manager
func NewManager(templatesDir string, logger *logger.Logger) *Manager {
	return &Manager{
		templatesDir: templatesDir,
		logger:       logger,
	}
}

// GetTemplate retrieves a development template by ID
func (m *Manager) GetTemplate(ctx context.Context, id string) (*dev.DevTemplate, error) {
	templatePath := filepath.Join(m.templatesDir, id+".json")
	
	data, err := os.ReadFile(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %s not found", id)
		}
		return nil, fmt.Errorf("failed to read template %s: %w", id, err)
	}

	var template dev.DevTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", id, err)
	}

	return &template, nil
}

// ListTemplates lists all available development templates
func (m *Manager) ListTemplates(ctx context.Context) ([]*dev.DevTemplate, error) {
	if err := os.MkdirAll(m.templatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}

	files, err := os.ReadDir(m.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []*dev.DevTemplate
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			templateID := strings.TrimSuffix(file.Name(), ".json")
			template, err := m.GetTemplate(ctx, templateID)
			if err != nil {
				m.logger.Warn(fmt.Sprintf("Failed to load template %s: %v", templateID, err))
				continue
			}
			templates = append(templates, template)
		}
	}

	// Sort templates by name
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// CreateTemplate creates a new development template
func (m *Manager) CreateTemplate(ctx context.Context, template *dev.DevTemplate) error {
	if err := os.MkdirAll(m.templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	
	if template.ID == "" {
		template.ID = m.generateTemplateID(template.Name)
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	templatePath := filepath.Join(m.templatesDir, template.ID+".json")
	if err := os.WriteFile(templatePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	m.logger.Info(fmt.Sprintf("Created development template %s (ID: %s)", template.Name, template.ID))
	return nil
}

// UpdateTemplate updates an existing development template
func (m *Manager) UpdateTemplate(ctx context.Context, template *dev.DevTemplate) error {
	templatePath := filepath.Join(m.templatesDir, template.ID+".json")
	
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template %s not found", template.ID)
	}

	template.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(templatePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	m.logger.Info(fmt.Sprintf("Updated development template %s (ID: %s)", template.Name, template.ID))
	return nil
}

// DeleteTemplate deletes a development template
func (m *Manager) DeleteTemplate(ctx context.Context, id string) error {
	templatePath := filepath.Join(m.templatesDir, id+".json")
	
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template %s not found", id)
	}

	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to delete template file: %w", err)
	}

	m.logger.Info(fmt.Sprintf("Deleted development template (ID: %s)", id))
	return nil
}

// SearchTemplates searches for templates matching a query
func (m *Manager) SearchTemplates(ctx context.Context, query string) ([]*dev.DevTemplate, error) {
	templates, err := m.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return templates, nil
	}

	query = strings.ToLower(query)
	var matches []*dev.DevTemplate

	for _, template := range templates {
		if m.matchesQuery(template, query) {
			matches = append(matches, template)
		}
	}

	return matches, nil
}

// matchesQuery checks if a template matches a search query
func (m *Manager) matchesQuery(template *dev.DevTemplate, query string) bool {
	searchFields := []string{
		template.Name,
		template.Description,
		template.Language,
		template.Framework,
		template.Category,
		template.Author,
	}

	for _, field := range searchFields {
		if strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}

	for _, tag := range template.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	return false
}

// generateTemplateID generates a unique template ID
func (m *Manager) generateTemplateID(name string) string {
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

// GetBuiltinTemplates returns a list of built-in development templates
func (m *Manager) GetBuiltinTemplates() []*dev.DevTemplate {
	return []*dev.DevTemplate{
		{
			ID:          "go-web-api",
			Name:        "Go Web API",
			Description: "Modern Go web API with Gin framework and PostgreSQL",
			Language:    "go",
			Framework:   "gin",
			Category:    "web",
			Tags:        []string{"web", "api", "rest", "postgresql"},
			Config: map[string]interface{}{
				"go_version": "1.21",
				"database":   "postgresql",
				"port":       8080,
			},
			Files: []dev.TemplateFile{
				{
					Path:     "main.go",
					Content:  m.getGoWebAPIMainFile(),
					Template: true,
				},
				{
					Path:     "go.mod",
					Content:  m.getGoModFile(),
					Template: true,
				},
				{
					Path:     "docker-compose.yml",
					Content:  m.getDockerComposeFile(),
					Template: true,
				},
				{
					Path:     "flake.nix",
					Content:  m.getGoFlakeFile(),
					Template: true,
				},
			},
			Dependencies: []dev.Dependency{
				{Name: "gin-gonic/gin", Version: "v1.9.1", Type: "go", Required: true},
				{Name: "lib/pq", Version: "v1.10.9", Type: "go", Required: true},
				{Name: "postgresql", Version: "15", Type: "system", Required: true},
			},
			Commands: []string{
				"go mod init {{.Name}}",
				"go mod tidy",
				"go run main.go",
			},
			Author:  "nixai",
			Version: "1.0.0",
		},
		{
			ID:          "rust-web-server",
			Name:        "Rust Web Server",
			Description: "High-performance Rust web server with Axum framework",
			Language:    "rust",
			Framework:   "axum",
			Category:    "web",
			Tags:        []string{"web", "server", "performance", "async"},
			Config: map[string]interface{}{
				"rust_version": "1.75",
				"port":         3000,
			},
			Files: []dev.TemplateFile{
				{
					Path:     "Cargo.toml",
					Content:  m.getRustCargoFile(),
					Template: true,
				},
				{
					Path:     "src/main.rs",
					Content:  m.getRustMainFile(),
					Template: true,
				},
				{
					Path:     "flake.nix",
					Content:  m.getRustFlakeFile(),
					Template: true,
				},
			},
			Dependencies: []dev.Dependency{
				{Name: "axum", Version: "0.7", Type: "rust", Required: true},
				{Name: "tokio", Version: "1.0", Type: "rust", Required: true},
				{Name: "serde", Version: "1.0", Type: "rust", Required: true},
			},
			Commands: []string{
				"cargo build",
				"cargo run",
			},
			Author:  "nixai",
			Version: "1.0.0",
		},
		{
			ID:          "python-fastapi",
			Name:        "Python FastAPI",
			Description: "Modern Python web API with FastAPI and PostgreSQL",
			Language:    "python",
			Framework:   "fastapi",
			Category:    "web",
			Tags:        []string{"web", "api", "python", "async"},
			Config: map[string]interface{}{
				"python_version": "3.11",
				"port":           8000,
			},
			Files: []dev.TemplateFile{
				{
					Path:     "main.py",
					Content:  m.getPythonFastAPIFile(),
					Template: true,
				},
				{
					Path:     "requirements.txt",
					Content:  m.getPythonRequirementsFile(),
					Template: true,
				},
				{
					Path:     "flake.nix",
					Content:  m.getPythonFlakeFile(),
					Template: true,
				},
			},
			Dependencies: []dev.Dependency{
				{Name: "fastapi", Version: "0.104.1", Type: "python", Required: true},
				{Name: "uvicorn", Version: "0.24.0", Type: "python", Required: true},
				{Name: "sqlalchemy", Version: "2.0.23", Type: "python", Required: true},
			},
			Commands: []string{
				"pip install -r requirements.txt",
				"uvicorn main:app --reload",
			},
			Author:  "nixai",
			Version: "1.0.0",
		},
		{
			ID:          "react-typescript",
			Name:        "React TypeScript",
			Description: "Modern React application with TypeScript and Vite",
			Language:    "typescript",
			Framework:   "react",
			Category:    "frontend",
			Tags:        []string{"frontend", "react", "typescript", "vite"},
			Config: map[string]interface{}{
				"node_version": "18",
				"port":         3000,
			},
			Files: []dev.TemplateFile{
				{
					Path:     "package.json",
					Content:  m.getReactPackageFile(),
					Template: true,
				},
				{
					Path:     "src/App.tsx",
					Content:  m.getReactAppFile(),
					Template: true,
				},
				{
					Path:     "vite.config.ts",
					Content:  m.getViteConfigFile(),
					Template: true,
				},
				{
					Path:     "flake.nix",
					Content:  m.getNodeFlakeFile(),
					Template: true,
				},
			},
			Dependencies: []dev.Dependency{
				{Name: "react", Version: "^18.2.0", Type: "npm", Required: true},
				{Name: "react-dom", Version: "^18.2.0", Type: "npm", Required: true},
				{Name: "typescript", Version: "^5.0.0", Type: "npm", Required: true},
				{Name: "vite", Version: "^5.0.0", Type: "npm", Required: true},
			},
			Commands: []string{
				"npm install",
				"npm run dev",
			},
			Author:  "nixai",
			Version: "1.0.0",
		},
	}
}

// InitializeBuiltinTemplates creates built-in templates if they don't exist
func (m *Manager) InitializeBuiltinTemplates(ctx context.Context) error {
	builtinTemplates := m.GetBuiltinTemplates()
	
	for _, template := range builtinTemplates {
		template.CreatedAt = time.Now()
		template.UpdatedAt = time.Now()
		
		templatePath := filepath.Join(m.templatesDir, template.ID+".json")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			if err := m.CreateTemplate(ctx, template); err != nil {
				m.logger.Warn(fmt.Sprintf("Failed to create built-in template %s: %v", template.Name, err))
			}
		}
	}
	
	return nil
}

// Template file content generators
func (m *Manager) getGoWebAPIMainFile() string {
	return `package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to {{.Name}} API",
		})
	})
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
	
	log.Printf("Starting server on port {{.Config.port}}")
	r.Run(":{{.Config.port}}")
}`
}

func (m *Manager) getGoModFile() string {
	return `module {{.Name}}

go {{.Config.go_version}}

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/lib/pq v1.10.9
)`
}

func (m *Manager) getDockerComposeFile() string {
	return `version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: {{.Name}}
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:`
}

func (m *Manager) getGoFlakeFile() string {
	return `{
  description = "{{.Name}} - Go Web API development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_{{.Config.go_version | replace "." "_"}}
            postgresql
            docker-compose
            air # Live reload for Go
          ];

          shellHook = '''
            echo "🚀 {{.Name}} development environment activated!"
            echo "📋 Available commands:"
            echo "  go run main.go    - Start the server"
            echo "  air               - Start with live reload"
            echo "  docker-compose up - Start PostgreSQL"
          ''';
        };
      });
}`
}

func (m *Manager) getRustCargoFile() string {
	return `[package]
name = "{{.Name}}"
version = "0.1.0"
edition = "2021"

[dependencies]
axum = "0.7"
tokio = { version = "1.0", features = ["full"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tower = "0.4"
tracing = "0.1"
tracing-subscriber = "0.3"`
}

func (m *Manager) getRustMainFile() string {
	return `use axum::{
    extract::Query,
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    message: String,
}

#[derive(Serialize)]
struct WelcomeResponse {
    message: String,
    service: String,
}

async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        message: "{{.Name}} is running".to_string(),
    })
}

async fn welcome() -> Json<WelcomeResponse> {
    Json(WelcomeResponse {
        message: "Welcome to {{.Name}} API".to_string(),
        service: "{{.Name}}".to_string(),
    })
}

#[tokio::main]
async fn main() {
    tracing_subscriber::init();

    let app = Router::new()
        .route("/", get(welcome))
        .route("/health", get(health));

    let listener = tokio::net::TcpListener::bind("0.0.0.0:{{.Config.port}}")
        .await
        .unwrap();

    println!("🚀 {{.Name}} server starting on port {{.Config.port}}");
    axum::serve(listener, app).await.unwrap();
}`
}

func (m *Manager) getRustFlakeFile() string {
	return `{
  description = "{{.Name}} - Rust Web Server development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    rust-overlay.url = "github:oxalica/rust-overlay";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, rust-overlay, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [ (import rust-overlay) ];
        pkgs = import nixpkgs {
          inherit system overlays;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            (rust-bin.stable.latest.default.override {
              extensions = [ "rust-src" "rustfmt" "clippy" ];
            })
            pkg-config
            openssl
            cargo-watch
          ];

          shellHook = '''
            echo "🦀 {{.Name}} Rust development environment activated!"
            echo "📋 Available commands:"
            echo "  cargo run         - Start the server"
            echo "  cargo watch -x run - Start with live reload"
            echo "  cargo test        - Run tests"
          ''';
        };
      });
}`
}

func (m *Manager) getPythonFastAPIFile() string {
	return `from fastapi import FastAPI
from pydantic import BaseModel
from typing import Dict, Any

app = FastAPI(title="{{.Name}}", version="1.0.0")

class HealthResponse(BaseModel):
    status: str
    message: str

class WelcomeResponse(BaseModel):
    message: str
    service: str

@app.get("/")
async def welcome() -> WelcomeResponse:
    return WelcomeResponse(
        message="Welcome to {{.Name}} API",
        service="{{.Name}}"
    )

@app.get("/health")
async def health() -> HealthResponse:
    return HealthResponse(
        status="healthy",
        message="{{.Name}} is running"
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port={{.Config.port}})`
}

func (m *Manager) getPythonRequirementsFile() string {
	return `fastapi==0.104.1
uvicorn[standard]==0.24.0
sqlalchemy==2.0.23
psycopg2-binary==2.9.9
pydantic==2.5.0
python-multipart==0.0.6`
}

func (m *Manager) getPythonFlakeFile() string {
	return `{
  description = "{{.Name}} - Python FastAPI development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            python{{.Config.python_version | replace "." ""}}
            python{{.Config.python_version | replace "." ""}}Packages.pip
            python{{.Config.python_version | replace "." ""}}Packages.virtualenv
            postgresql
          ];

          shellHook = '''
            echo "🐍 {{.Name}} Python development environment activated!"
            echo "📋 Available commands:"
            echo "  pip install -r requirements.txt - Install dependencies"
            echo "  uvicorn main:app --reload        - Start with live reload"
            echo "  python -m pytest                - Run tests"
          ''';
        };
      });
}`
}

func (m *Manager) getReactPackageFile() string {
	return `{
  "name": "{{.Name}}",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.43",
    "@types/react-dom": "^18.2.17",
    "@typescript-eslint/eslint-plugin": "^6.14.0",
    "@typescript-eslint/parser": "^6.14.0",
    "@vitejs/plugin-react": "^4.2.1",
    "eslint": "^8.55.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "eslint-plugin-react-refresh": "^0.4.5",
    "typescript": "^5.2.2",
    "vite": "^5.0.8"
  }
}`
}

func (m *Manager) getReactAppFile() string {
	return `import React from 'react';
import './App.css';

function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>{{.Name}}</h1>
        <p>Welcome to your React TypeScript application!</p>
        <p>Edit <code>src/App.tsx</code> and save to reload.</p>
      </header>
    </div>
  );
}

export default App;`
}

func (m *Manager) getViteConfigFile() string {
	return `import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: {{.Config.port}},
    host: true
  }
})`
}

func (m *Manager) getNodeFlakeFile() string {
	return `{
  description = "{{.Name}} - React TypeScript development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            nodejs_{{.Config.node_version}}
            yarn
            nodePackages.typescript
            nodePackages.eslint
          ];

          shellHook = '''
            echo "⚛️ {{.Name}} React development environment activated!"
            echo "📋 Available commands:"
            echo "  npm install  - Install dependencies"
            echo "  npm run dev  - Start development server"
            echo "  npm run build - Build for production"
          ''';
        };
      });
}`
}