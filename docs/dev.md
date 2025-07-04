# nixai dev - Developer Experience Revolution

**Phase 3.3 - Intelligent Development Environment Management**

The `nixai dev` command provides revolutionary developer experience with one-command project setup, intelligent templates, automatic dependency detection, and comprehensive IDE integration. Transform your development workflow with AI-powered environment management.

---

## 🚀 Core Features

### ✨ One-Command Project Setup
- **Instant Environment Creation**: Complete development environments in seconds
- **Intelligent Templates**: 4 built-in production-ready templates
- **Multi-Language Support**: Go, Rust, Python, TypeScript/React, and more
- **IDE Integration**: Automatic configuration for 6 popular editors
- **Best Practices**: Industry-standard project scaffolding

### 🧠 Intelligent Templates
- **Go Web API**: Modern Gin framework with PostgreSQL integration
- **Rust Web Server**: High-performance Axum with async support
- **Python FastAPI**: Modern API development with type hints
- **React TypeScript**: Vite-powered frontend with modern tooling

### 📦 Automatic Dependency Detection
- **10+ Languages**: Go, Rust, Python, Node.js, Ruby, PHP, .NET, Java
- **Package Managers**: go.mod, Cargo.toml, package.json, requirements.txt, and more
- **Smart Analysis**: Detects dependencies from existing projects

### 🔧 IDE Integration
- **6 Editors Supported**: VS Code, Neovim, Vim, Emacs, IntelliJ, Eclipse
- **Configuration Generation**: Automatic settings, launch configs, and tasks
- **Language-Specific Setup**: Optimized configurations for each language

---

## 🛠️ Command Reference

### Quick Start

```bash
# One-command project setup
nixai dev setup my-web-app --language go --framework gin --editor vscode

# Interactive setup wizard
nixai dev setup --interactive

# Use built-in template
nixai dev setup api-service --template go-web-api --editor neovim
```

### Global Flags

- `--language` - Programming language (go, rust, python, typescript, javascript)
- `--framework` - Framework (gin, axum, fastapi, react, vue)
- `--editor` - Editor/IDE (vscode, neovim, vim, emacs, intellij)
- `--containers` - Containers to include (docker, postgresql, mysql, redis)
- `--template` - Template to use (overrides language/framework)
- `--path` - Project path (default: current directory)
- `--interactive` - Interactive setup wizard
- `--auto-detect` - Automatically detect dependencies (default: true)

---

## 📋 Subcommands

### Project Setup (`nixai dev setup`)

Create complete development environments with intelligent defaults.

#### Basic Usage
```bash
# Language-based setup
nixai dev setup my-project --language rust --editor vscode

# Framework-specific setup
nixai dev setup web-app --language typescript --framework react --editor neovim

# Container integration
nixai dev setup api --language go --containers postgresql,redis --editor vscode
```

#### Template-Based Setup
```bash
# Use built-in templates
nixai dev setup backend --template go-web-api
nixai dev setup frontend --template react-typescript
nixai dev setup api --template python-fastapi
nixai dev setup server --template rust-web-server
```

#### Interactive Setup
```bash
# Guided setup wizard
nixai dev setup --interactive

# Will prompt for:
# - Project name
# - Programming language
# - Framework (optional)
# - Editor/IDE preference
# - Container requirements
```

**Example Output:**
```
🚀 Setting up development environment: my-web-app
📍 Language: go
🔧 Framework: gin
💻 Editor: vscode
🐳 Containers: postgresql

✅ Development environment created successfully!
📂 Location: ./my-web-app
🆔 Environment ID: my-web-app

📋 Next steps:
1. cd my-web-app
2. nix develop  # Enter development shell
3. go mod tidy  # Install dependencies
4. go run main.go  # Run the application

💡 IDE integration configured for vscode
   Open with: code my-web-app
```

### Environment Management (`nixai dev env`)

Manage development environment lifecycle and configuration.

#### Create Environment
```bash
# Create new environment
nixai dev env create my-env --language python --editor vscode

# Create with template
nixai dev env create web-env --template react-typescript
```

#### List Environments
```bash
# List all environments
nixai dev env list

# Example output:
📦 Development Environments (3)

🟢 my-web-app
   ID: my-web-app
   Language: go (gin)
   Path: ./my-web-app
   Created: 2025-07-04 13:30:15
   Dependencies: 5

🟢 api-service  
   ID: api-service
   Language: rust (axum)
   Path: ./api-service
   Created: 2025-07-04 13:25:42
   Dependencies: 8
```

#### Environment Details
```bash
# Show environment details
nixai dev env show my-web-app

# JSON output
nixai dev env show my-web-app --format json
```

#### Environment Control
```bash
# Start environment services
nixai dev env start my-web-app

# Stop environment
nixai dev env stop my-web-app

# Delete environment
nixai dev env delete my-web-app
```

### Template Management (`nixai dev template`)

Discover, use, and manage development templates.

#### List Templates
```bash
# List all templates
nixai dev template list

# Filter by category
nixai dev template list --category web

# Example output:
📋 Development Templates (4)

📂 Web
  🔧 Go Web API (go-web-api)
      Language: go + gin
      Modern Go web API with Gin framework and PostgreSQL
      Tags: web, api, rest, postgresql
      Author: nixai (v1.0.0)

  🔧 Python FastAPI (python-fastapi)
      Language: python + fastapi
      Modern Python web API with FastAPI and PostgreSQL
      Tags: web, api, python, async
      Author: nixai (v1.0.0)

📂 Frontend
  🔧 React TypeScript (react-typescript)
      Language: typescript + react
      Modern React application with TypeScript and Vite
      Tags: frontend, react, typescript, vite
      Author: nixai (v1.0.0)
```

#### Template Details
```bash
# Show template details
nixai dev template show go-web-api

# JSON output for integration
nixai dev template show go-web-api --format json
```

#### Search Templates
```bash
# Search templates
nixai dev template search "web api"
nixai dev template search "react"
nixai dev template search "rust"
```

#### Create Custom Templates
```bash
# Create new template
nixai dev template create my-template \
  --name "My Custom Template" \
  --description "Custom development template" \
  --language python \
  --framework flask \
  --category web \
  --tags web,api,flask
```

### Dependency Management (`nixai dev deps`)

Automatic dependency detection and management across languages.

#### Detect Dependencies
```bash
# Detect in current directory
nixai dev deps detect

# Detect in specific path
nixai dev deps detect ./my-project

# JSON output
nixai dev deps detect --format json

# Example output:
📦 Detected Dependencies (12)

📋 Go Dependencies (5):
  • github.com/gin-gonic/gin v1.9.1 (required)
  • github.com/lib/pq v1.10.9 (required)
  • github.com/stretchr/testify v1.8.4 (required)
    Source: go.mod

📋 Npm Dependencies (7):
  • react ^18.2.0 (required)
  • react-dom ^18.2.0 (required)  
  • typescript ^5.0.0 (required)
  • vite ^5.0.0 (required)
    Source: package.json
```

#### Dependency Operations
```bash
# Install detected dependencies
nixai dev deps install

# Update dependencies
nixai dev deps update

# Check for outdated dependencies
nixai dev deps check
```

### IDE Integration (`nixai dev ide`)

Configure and manage IDE integrations with automatic setup.

#### List Supported IDEs
```bash
# List all supported IDEs
nixai dev ide list

# Example output:
💻 Supported IDEs (6)

🔧 Visual Studio Code (vscode)
   Type: editor
   Extensions: 8 available

🔧 Neovim (neovim)
   Type: editor
   Extensions: 6 available

🔧 Vim (vim)
   Type: editor
   Extensions: 5 available

🔧 Emacs (emacs)
   Type: editor
   Extensions: 6 available

🔧 IntelliJ IDEA (intellij)
   Type: ide
   Extensions: available

🔧 Eclipse (eclipse)
   Type: ide
   Extensions: available
```

#### Setup IDE Integration
```bash
# Setup IDE for environment
nixai dev ide setup my-web-app vscode
nixai dev ide setup api-service neovim
```

#### IDE Configuration Details
```bash
# Show IDE configuration
nixai dev ide config vscode

# Example output:
💻 IDE Configuration: Visual Studio Code

Type: editor

📄 Configuration Files:
  • .vscode/settings.json
  • .vscode/launch.json
  • .vscode/tasks.json

🔌 Recommended Extensions:
  • ms-vscode.vscode-json
  • golang.go
  • rust-lang.rust-analyzer
  • ms-python.python

⚙️ Default Settings:
  editor.formatOnSave: true
  editor.tabSize: 2
  files.trimTrailingWhitespace: true
```

---

## 🎯 Built-in Templates

### Go Web API Template
**ID**: `go-web-api`  
**Language**: Go + Gin framework  
**Features**: PostgreSQL integration, Docker support, comprehensive Makefile

**Generated Files:**
- `main.go` - Web server with Gin routes
- `go.mod` - Go module configuration
- `docker-compose.yml` - PostgreSQL service
- `flake.nix` - Nix development environment
- `.vscode/` - VS Code configuration
- `Makefile` - Build and development tasks
- `README.md` - Project documentation

**Example Usage:**
```bash
nixai dev setup my-api --template go-web-api --editor vscode
cd my-api
nix develop
go mod tidy
go run main.go
```

### Rust Web Server Template
**ID**: `rust-web-server`  
**Language**: Rust + Axum framework  
**Features**: High-performance async server, cargo workspace

**Generated Files:**
- `src/main.rs` - Axum web server
- `Cargo.toml` - Rust dependencies
- `flake.nix` - Rust development environment
- `Dockerfile` - Multi-stage build
- `.vscode/` - Rust analyzer configuration

### Python FastAPI Template
**ID**: `python-fastapi`  
**Language**: Python + FastAPI  
**Features**: Type hints, automatic API documentation, SQLAlchemy

**Generated Files:**
- `main.py` - FastAPI application
- `requirements.txt` - Python dependencies
- `flake.nix` - Python development environment
- `Dockerfile` - Python container
- `.vscode/` - Python configuration

### React TypeScript Template
**ID**: `react-typescript`  
**Language**: TypeScript + React + Vite  
**Features**: Modern React setup, ESLint, TypeScript configuration

**Generated Files:**
- `src/App.tsx` - React application
- `package.json` - npm dependencies
- `vite.config.ts` - Vite configuration
- `flake.nix` - Node.js development environment
- `.vscode/` - TypeScript configuration

---

## 🔧 Advanced Usage

### Multi-Service Projects
```bash
# Create backend service
nixai dev setup backend --template go-web-api --path ./services/backend

# Create frontend service  
nixai dev setup frontend --template react-typescript --path ./services/frontend

# Create shared database
nixai dev setup database --containers postgresql --path ./services/database
```

### Custom Development Workflows
```bash
# Full-stack development setup
nixai dev setup fullstack-app --language typescript --framework react --editor vscode
cd fullstack-app

# Add backend API
nixai dev env create api --template go-web-api --path ./backend

# Add database services
nixai dev env create db --containers postgresql,redis --path ./database
```

### Team Development
```bash
# Standardized team setup
nixai dev template create team-standard \
  --name "Team Standard Template" \
  --language go \
  --framework gin \
  --category internal

# Team members use standardized setup
nixai dev setup project-name --template team-standard
```

---

## 💡 Best Practices

### Project Organization
1. **Use Templates**: Start with built-in templates for common patterns
2. **Consistent Structure**: Follow template-generated project structure
3. **Environment Management**: Use `nixai dev env` for lifecycle management
4. **Dependency Tracking**: Regular `nixai dev deps check` for updates

### Development Workflow
1. **Start with Setup**: `nixai dev setup` creates complete environments
2. **Enter Dev Shell**: `nix develop` for isolated development
3. **IDE Integration**: Automatic configuration for chosen editor
4. **Service Management**: Use `nixai dev env start/stop` for services

### Template Usage
1. **Explore Templates**: `nixai dev template list` to discover options
2. **Understand Structure**: `nixai dev template show` before using
3. **Customize Templates**: Create custom templates for team standards
4. **Share Templates**: Export and share custom templates

### IDE Configuration
1. **Automatic Setup**: Let nixai configure your preferred IDE
2. **Language-Specific**: IDE configs optimized for chosen language
3. **Extensions**: Recommended extensions automatically documented
4. **Multi-IDE**: Different team members can use different IDEs

---

## 🐛 Troubleshooting

### Common Issues

#### Template Not Found
```bash
# List available templates
nixai dev template list

# Search for specific template
nixai dev template search "your-query"

# Use exact template ID
nixai dev setup project --template go-web-api
```

#### Dependency Detection Issues
```bash
# Manual dependency detection
nixai dev deps detect ./project-path

# Check supported file formats
nixai dev deps detect --help

# Install dependencies manually
nixai dev deps install ./project-path
```

#### IDE Integration Problems
```bash
# List supported IDEs
nixai dev ide list

# Check IDE configuration
nixai dev ide config vscode

# Re-setup IDE integration
nixai dev ide setup environment-id vscode
```

#### Environment Management
```bash
# Check environment status
nixai dev env list

# View environment details
nixai dev env show environment-id

# Clean up environments
nixai dev env cleanup
```

### Debug Mode
```bash
# Enable verbose logging
nixai dev setup project --verbose

# Check environment logs
nixai dev env logs environment-id

# Validate project structure
nixai dev validate ./project-path
```

---

## 🔗 Related Commands

- [`nixai build`](build.md) - Build and analyze NixOS configurations
- [`nixai flake`](flake.md) - Manage NixOS flakes and configurations
- [`nixai health`](health.md) - System health monitoring and diagnostics
- [`nixai templates`](templates.md) - NixOS configuration templates
- [`nixai configure`](configure.md) - Interactive NixOS configuration

---

## 🎯 Real-World Examples

### Startup MVP Development
```bash
# Backend API
nixai dev setup mvp-api --template go-web-api --editor vscode
cd mvp-api && nix develop

# Frontend App
nixai dev setup mvp-frontend --template react-typescript --editor vscode
cd mvp-frontend && nix develop && npm install

# Quick prototype ready in minutes!
```

### Microservices Architecture
```bash
# User service
nixai dev setup user-service --template go-web-api --containers postgresql

# Payment service  
nixai dev setup payment-service --template rust-web-server --containers redis

# API gateway
nixai dev setup api-gateway --template python-fastapi --containers redis

# Frontend dashboard
nixai dev setup dashboard --template react-typescript
```

### Learning New Technologies
```bash
# Learn Rust web development
nixai dev setup rust-learning --template rust-web-server --editor vscode

# Explore Python async programming
nixai dev setup python-async --template python-fastapi --editor neovim

# Modern React development
nixai dev setup react-modern --template react-typescript --editor vscode
```

### Open Source Contribution
```bash
# Setup for contributing to existing project
cd existing-project
nixai dev deps detect
nixai dev ide setup existing-project vscode

# Create development environment
nixai dev env create contrib-env --auto-detect --editor vscode
```

---

The `nixai dev` command revolutionizes NixOS development by providing intelligent, one-command project setup with industry best practices, comprehensive IDE integration, and automatic dependency management. Perfect for rapid prototyping, team standardization, and learning new technologies.