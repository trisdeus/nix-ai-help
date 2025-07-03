# Development Environment Plugin

An intelligent development environment management plugin for NixAI that provides AI-powered project setup, template generation, and automated environment configuration.

## Features

### 🚀 **Smart Environment Management**
- **Automatic Project Detection**: AI-powered analysis of existing projects
- **Environment Templates**: Pre-configured templates for popular languages and frameworks
- **Quick Setup**: One-command environment creation and activation
- **Cross-Platform Support**: Works on any system with Nix

### 🔧 **Multi-Language Support**
- **Go**: Complete Go development stack with gopls, gofmt, and testing tools
- **Python**: Python environments with virtual environments and package management
- **Node.js**: Modern JavaScript/TypeScript development with npm/yarn support
- **Rust**: Rust development with Cargo and language server
- **Java/Kotlin**: JVM-based development environments
- **C/C++**: Native development with GCC/Clang toolchains

### 🤖 **AI-Powered Intelligence**
- **Project Analysis**: Automatic detection of languages, frameworks, and dependencies
- **Environment Suggestions**: Smart recommendations based on project structure
- **Dependency Resolution**: Automatic dependency management and conflict resolution
- **Optimization Recommendations**: Performance and best practice suggestions

### 📋 **Template System**
- **Built-in Templates**: Ready-to-use templates for common development scenarios
- **Custom Templates**: Create and share your own environment templates
- **Framework-Specific**: Specialized templates for React, Django, Spring, etc.
- **Verified Templates**: Community-verified and tested configurations

## Installation

1. **Copy Plugin**: Place the plugin in your nixai plugins directory:
   ```bash
   cp -r dev-environment ~/.nixai/plugins/
   ```

2. **Load Plugin**: Load the plugin using nixai:
   ```bash
   nixai plugin load dev-environment
   ```

3. **Start Plugin**: Start the environment manager:
   ```bash
   nixai plugin start dev-environment
   ```

## Usage Examples

### 📋 **List Environments**
```bash
# List all environments
nixai plugin execute dev-environment list-environments

# Filter by type
nixai plugin execute dev-environment list-environments '{"type": "go"}'

# Filter by status
nixai plugin execute dev-environment list-environments '{"status": "active"}'

# Example output:
[
  {
    "id": "env-1642167890",
    "name": "my-go-project",
    "type": "go",
    "path": "/home/user/projects/my-go-project",
    "status": "active",
    "description": "Go development environment",
    "languages": [
      {"name": "go", "version": "1.21"}
    ],
    "tools": [
      {"name": "gopls", "purpose": "Language server", "essential": true}
    ],
    "health": {
      "status": "healthy",
      "score": 95,
      "last_check": "2024-01-15T11:30:00Z"
    }
  }
]
```

### 🆕 **Create New Environment**
```bash
# Create basic environment
nixai plugin execute dev-environment create-environment '{
  "name": "my-new-project",
  "type": "go",
  "path": "/home/user/projects/my-new-project"
}'

# Create with specific template
nixai plugin execute dev-environment create-environment '{
  "name": "web-app",
  "type": "nodejs",
  "path": "/home/user/projects/web-app",
  "template": "nodejs-react"
}'

# Example output:
{
  "id": "env-1642167891",
  "name": "my-new-project",
  "type": "go",
  "path": "/home/user/projects/my-new-project",
  "status": "inactive",
  "description": "Go development environment",
  "languages": [
    {"name": "go", "version": "latest"}
  ],
  "tools": [
    {"name": "gopls", "purpose": "Language server", "essential": true},
    {"name": "gofmt", "purpose": "Code formatter", "essential": true}
  ],
  "config": {
    "env_vars": {
      "GOPATH": "/home/user/projects/my-new-project/.go"
    },
    "project_settings": {
      "build_command": "go build",
      "test_command": "go test ./..."
    }
  },
  "created_at": "2024-01-15T11:30:00Z"
}
```

### 📚 **Environment Templates**
```bash
# List available templates
nixai plugin execute dev-environment list-templates

# Filter by category
nixai plugin execute dev-environment list-templates '{"category": "web"}'

# Example output:
[
  {
    "id": "go-basic",
    "name": "Go Basic",
    "description": "Basic Go development environment",
    "category": "go",
    "languages": [
      {"name": "go", "version": "latest"}
    ],
    "tools": [
      {"name": "gopls", "purpose": "Language server", "essential": true}
    ],
    "popular": true,
    "verified": true
  },
  {
    "id": "python-web",
    "name": "Python Web Development",
    "description": "Python environment with Django/Flask support",
    "category": "python",
    "languages": [
      {"name": "python", "version": "3.11"}
    ],
    "services": [
      {"name": "postgresql", "type": "database"}
    ]
  }
]
```

### 🎯 **Create from Template**
```bash
# Create environment from template
nixai plugin execute dev-environment create-from-template '{
  "template_id": "python-web",
  "name": "my-web-app",
  "path": "/home/user/projects/my-web-app"
}'

# Example output:
{
  "id": "env-1642167892",
  "name": "my-web-app",
  "type": "python",
  "path": "/home/user/projects/my-web-app",
  "status": "inactive",
  "description": "Python environment with Django/Flask support",
  "languages": [
    {"name": "python", "version": "3.11"}
  ],
  "services": [
    {"name": "postgresql", "type": "database", "port": 5432}
  ],
  "tools": [
    {"name": "pip", "purpose": "Package manager", "essential": true},
    {"name": "pylsp", "purpose": "Language server", "essential": true}
  ]
}
```

### ⚡ **Environment Activation**
```bash
# Activate environment
nixai plugin execute dev-environment activate-environment '{
  "environment_id": "env-1642167890"
}'

# Example output:
{
  "environment_id": "env-1642167890",
  "status": "activated",
  "script": "#!/bin/bash\n\n# Activate development environment\ncd \"/home/user/projects/my-go-project\"\ndirenv allow\nnix-shell\n",
  "instructions": [
    "Navigate to project directory: cd /home/user/projects/my-go-project",
    "Allow direnv: direnv allow",
    "Enter nix shell: nix-shell",
    "Your development environment is now active!"
  ]
}

# Deactivate environment
nixai plugin execute dev-environment deactivate-environment '{
  "environment_id": "env-1642167890"
}'
```

### 🔍 **Project Analysis**
```bash
# Analyze existing project
nixai plugin execute dev-environment analyze-project '{
  "path": "/home/user/existing-project"
}'

# Example output:
{
  "path": "/home/user/existing-project",
  "detected_languages": ["python", "javascript"],
  "detected_tools": ["make", "docker"],
  "config_files": ["requirements.txt", "package.json", "Dockerfile"],
  "package_files": ["requirements.txt", "package.json"],
  "recommendations": [
    "Consider creating a Python virtual environment",
    "Add direnv configuration for automatic environment activation"
  ]
}
```

### 🤖 **AI Environment Suggestions**
```bash
# Get AI-powered environment suggestions
nixai plugin execute dev-environment suggest-environment '{
  "path": "/home/user/existing-project"
}'

# Example output:
{
  "analysis": {
    "detected_languages": ["python", "javascript"],
    "detected_tools": ["docker"],
    "config_files": ["requirements.txt", "package.json"]
  },
  "suggestions": [
    {
      "template_id": "python-web",
      "template_name": "Python Web Development",
      "description": "Python environment with web framework support",
      "confidence": 85,
      "reasons": [
        "Detected Python files in project",
        "Found requirements.txt indicating Python project",
        "Template matches project structure"
      ]
    },
    {
      "template_id": "fullstack-js",
      "template_name": "Full-Stack JavaScript",
      "description": "Complete JavaScript development stack",
      "confidence": 70,
      "reasons": [
        "Detected JavaScript files",
        "Found package.json configuration"
      ]
    }
  ]
}
```

### 🏥 **Health Monitoring**
```bash
# Check single environment health
nixai plugin execute dev-environment health-check '{
  "environment_id": "env-1642167890"
}'

# Example output:
{
  "status": "healthy",
  "score": 95,
  "last_check": "2024-01-15T11:30:00Z",
  "issues": [],
  "checks": [
    {"name": "Path Exists", "status": "pass", "message": "Project path is accessible"},
    {"name": "File .envrc", "status": "pass", "message": "File .envrc exists"},
    {"name": "Tool gopls", "status": "pass", "message": "Tool gopls is available"}
  ],
  "suggestions": [
    "Environment is running optimally",
    "Consider enabling automatic dependency updates"
  ]
}

# System-wide health check
nixai plugin execute dev-environment health-check

# Example output:
{
  "overall_status": "healthy",
  "total_environments": 5,
  "healthy_environments": 4,
  "broken_environments": 1,
  "issues": [
    "Environment old-project is unhealthy"
  ],
  "recommendations": [
    "Fix broken environment: old-project",
    "Update outdated dependencies"
  ],
  "last_check": "2024-01-15T11:30:00Z"
}
```

### 🔄 **Dependency Management**
```bash
# Update environment dependencies
nixai plugin execute dev-environment update-dependencies '{
  "environment_id": "env-1642167890"
}'

# Example output:
{
  "environment_id": "env-1642167890",
  "updates": [
    "Tidied go modules"
  ],
  "errors": [],
  "status": "success",
  "output": "go: downloading example.com/module v1.2.3"
}
```

### 💾 **Backup and Restore**
```bash
# Backup environment
nixai plugin execute dev-environment backup-environment '{
  "environment_id": "env-1642167890"
}'

# Example output:
{
  "environment_id": "env-1642167890",
  "backup_path": "/tmp/nixai-env-backups/my-go-project-1642167890.json",
  "timestamp": "2024-01-15T11:30:00Z"
}

# Restore environment
nixai plugin execute dev-environment restore-environment '{
  "backup_path": "/tmp/nixai-env-backups/my-go-project-1642167890.json"
}'
```

## Available Operations

| Operation | Description | Parameters |
|-----------|-------------|------------|
| `list-environments` | List development environments | `type`, `status` |
| `create-environment` | Create new environment | `name`, `type`, `path`, `template` |
| `delete-environment` | Delete environment | `environment_id` |
| `activate-environment` | Activate environment | `environment_id` |
| `deactivate-environment` | Deactivate environment | `environment_id` |
| `list-templates` | List available templates | `category` |
| `create-from-template` | Create from template | `template_id`, `name`, `path` |
| `analyze-project` | Analyze existing project | `path` |
| `suggest-environment` | AI environment suggestions | `path` |
| `health-check` | Health check environments | `environment_id` (optional) |
| `update-dependencies` | Update dependencies | `environment_id` |
| `backup-environment` | Backup environment | `environment_id` |
| `restore-environment` | Restore from backup | `backup_path` |

## Supported Languages & Frameworks

### 🔵 **Go Development**
```bash
# Features:
- Go toolchain (go, gofmt, gopls)
- Module management
- Testing frameworks
- Performance profiling tools
- Docker integration

# Template: go-basic, go-web, go-microservices
```

### 🐍 **Python Development**
```bash
# Features:
- Python interpreter (3.8, 3.9, 3.10, 3.11)
- Virtual environment management
- Package managers (pip, poetry, pipenv)
- Web frameworks (Django, Flask, FastAPI)
- Data science tools (pandas, numpy, jupyter)
- Testing (pytest, unittest)

# Templates: python-basic, python-web, python-data, python-ml
```

### 🟢 **Node.js Development**
```bash
# Features:
- Node.js runtime (16, 18, 20)
- Package managers (npm, yarn, pnpm)
- TypeScript support
- Frontend frameworks (React, Vue, Angular)
- Backend frameworks (Express, NestJS)
- Testing (Jest, Mocha, Cypress)

# Templates: nodejs-basic, nodejs-web, nodejs-react, nodejs-api
```

### 🦀 **Rust Development**
```bash
# Features:
- Rust toolchain (rustc, cargo)
- Language server (rust-analyzer)
- Package management
- Testing and benchmarking
- Cross-compilation support

# Templates: rust-basic, rust-web, rust-cli
```

### ☕ **Java/Kotlin Development**
```bash
# Features:
- JDK (8, 11, 17, 21)
- Build tools (Maven, Gradle)
- Spring Boot support
- Kotlin language support
- Testing frameworks (JUnit, TestNG)

# Templates: java-basic, java-spring, kotlin-basic
```

## Advanced Features

### 🎯 **Environment Health Scoring**
Environments are scored 0-100 based on:
- **Path Accessibility** (50 points): Project directory exists and is accessible
- **Required Files** (30 points): Essential configuration files present
- **Tool Availability** (20 points): Required development tools installed

### 📊 **Smart Project Detection**
```bash
# Automatic detection of:
- Programming languages (by file extensions)
- Package managers (package.json, requirements.txt, etc.)
- Build tools (Makefile, Dockerfile, etc.)
- Frameworks (framework-specific files)
- Configuration files (.envrc, shell.nix, etc.)
```

### 🔧 **Generated Configuration Files**

#### `.envrc` (direnv configuration)
```bash
use nix

export GOPATH="/project/path/.go"
export PATH="$GOPATH/bin:$PATH"
alias run="go run ."
alias test="go test ./..."
```

#### `shell.nix` (Nix environment)
```nix
{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    golint
    git
  ];
  
  shellHook = ''
    echo "Go development environment activated"
    echo "Go version: $(go version)"
  '';
}
```

### 🚀 **Quick Start Workflows**

#### New Go Project
```bash
# Create and activate Go environment
nixai plugin execute dev-environment create-environment '{
  "name": "my-go-app",
  "type": "go",
  "path": "/home/user/projects/my-go-app"
}'

# Activate environment
nixai plugin execute dev-environment activate-environment '{
  "environment_id": "env-123"
}'

# Now you can:
cd /home/user/projects/my-go-app
direnv allow
nix-shell
go mod init my-go-app
```

#### Existing Project Analysis
```bash
# Analyze existing project
nixai plugin execute dev-environment analyze-project '{
  "path": "/home/user/existing-project"
}'

# Get AI suggestions
nixai plugin execute dev-environment suggest-environment '{
  "path": "/home/user/existing-project"
}'

# Create environment from suggestion
nixai plugin execute dev-environment create-from-template '{
  "template_id": "suggested-template",
  "name": "existing-project-env",
  "path": "/home/user/existing-project"
}'
```

## Integration Examples

### 🔄 **Automated Environment Setup**
```bash
#!/bin/bash
# setup-dev-env.sh

PROJECT_PATH="$1"
PROJECT_NAME=$(basename "$PROJECT_PATH")

# Analyze project
ANALYSIS=$(nixai plugin execute dev-environment analyze-project "{\"path\": \"$PROJECT_PATH\"}")

# Get suggestions
SUGGESTIONS=$(nixai plugin execute dev-environment suggest-environment "{\"path\": \"$PROJECT_PATH\"}")

# Create environment from best suggestion
TEMPLATE_ID=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].template_id')

nixai plugin execute dev-environment create-from-template "{
  \"template_id\": \"$TEMPLATE_ID\",
  \"name\": \"$PROJECT_NAME\",
  \"path\": \"$PROJECT_PATH\"
}"

echo "Development environment created for $PROJECT_NAME"
```

### 📊 **Environment Health Dashboard**
```bash
#!/bin/bash
# health-dashboard.sh

echo "=== Development Environment Health Dashboard ==="
HEALTH=$(nixai plugin execute dev-environment health-check)

echo "Overall Status: $(echo "$HEALTH" | jq -r '.overall_status')"
echo "Total Environments: $(echo "$HEALTH" | jq -r '.total_environments')"
echo "Healthy: $(echo "$HEALTH" | jq -r '.healthy_environments')"
echo "Broken: $(echo "$HEALTH" | jq -r '.broken_environments')"

echo ""
echo "=== Environment Details ==="
nixai plugin execute dev-environment list-environments | jq '.[] | {name, type, status, health: .health.status}'
```

### 🔧 **Bulk Environment Operations**
```bash
#!/bin/bash
# bulk-operations.sh

# Update all environments
ENVS=$(nixai plugin execute dev-environment list-environments | jq -r '.[].id')

for env_id in $ENVS; do
  echo "Updating environment: $env_id"
  nixai plugin execute dev-environment update-dependencies "{\"environment_id\": \"$env_id\"}"
done

echo "All environments updated!"
```

## Best Practices

### 🎯 **Environment Organization**
1. **Consistent Naming**: Use descriptive, consistent environment names
2. **Project Structure**: Organize projects in logical directory hierarchies
3. **Template Usage**: Leverage templates for consistency across projects
4. **Regular Health Checks**: Monitor environment health proactively

### 🔧 **Development Workflow**
```bash
# Recommended workflow:
1. Analyze existing projects before creating environments
2. Use templates for new projects
3. Activate environments before development
4. Regularly update dependencies
5. Monitor environment health
6. Backup important environments
```

### 📊 **Performance Optimization**
```bash
# Optimize environment performance:
- Use .envrc for automatic activation
- Cache dependencies when possible
- Clean up unused environments
- Monitor resource usage
- Use lightweight tools when possible
```

## Troubleshooting

### Common Issues

#### Environment Won't Activate
```bash
# Check environment status
nixai plugin execute dev-environment health-check '{"environment_id": "env-123"}'

# Verify path exists
ls -la /path/to/project

# Check direnv setup
which direnv
direnv version
```

#### Missing Dependencies
```bash
# Check tool availability
which go gopls git

# Update dependencies
nixai plugin execute dev-environment update-dependencies '{"environment_id": "env-123"}'

# Recreate environment if needed
nixai plugin execute dev-environment delete-environment '{"environment_id": "env-123"}'
```

#### Template Issues
```bash
# List available templates
nixai plugin execute dev-environment list-templates

# Check template compatibility
nixai plugin execute dev-environment analyze-project '{"path": "/project/path"}'
```

## Contributing

Areas for enhancement:

1. **Additional Language Support**: Support for more programming languages
2. **Framework Templates**: Specialized templates for popular frameworks
3. **Cloud Integration**: Integration with cloud development environments
4. **IDE Integration**: Better integration with popular IDEs and editors

## License

MIT License - see LICENSE file for details.

---

**Professional development environment management for modern developers** 🚀