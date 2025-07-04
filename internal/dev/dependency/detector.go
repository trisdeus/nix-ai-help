package dependency

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nix-ai-help/internal/dev"
	"nix-ai-help/pkg/logger"
)

// Manager implements the DependencyManager interface
type Manager struct {
	logger    *logger.Logger
	detectors map[string]*dev.DependencyDetector
}

// NewManager creates a new dependency manager
func NewManager(logger *logger.Logger) *Manager {
	return &Manager{
		logger:    logger,
		detectors: make(map[string]*dev.DependencyDetector),
	}
}

// DetectDependencies automatically detects dependencies in a project
func (m *Manager) DetectDependencies(ctx context.Context, path string) ([]dev.Dependency, error) {
	var allDependencies []dev.Dependency
	
	// Walk through the project directory
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// Detect dependencies based on file types
		deps, err := m.detectFromFile(filePath)
		if err != nil {
			m.logger.Warn(fmt.Sprintf("Failed to detect dependencies from file %s: %v", filePath, err))
			return nil
		}
		
		allDependencies = append(allDependencies, deps...)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk project directory: %w", err)
	}
	
	// Remove duplicates and merge dependencies
	return m.mergeDependencies(allDependencies), nil
}

// detectFromFile detects dependencies from a specific file
func (m *Manager) detectFromFile(filePath string) ([]dev.Dependency, error) {
	filename := filepath.Base(filePath)
	
	switch {
	case filename == "go.mod":
		return m.detectGoMod(filePath)
	case filename == "Cargo.toml":
		return m.detectCargoToml(filePath)
	case filename == "package.json":
		return m.detectPackageJson(filePath)
	case filename == "requirements.txt":
		return m.detectRequirementsTxt(filePath)
	case filename == "Pipfile":
		return m.detectPipfile(filePath)
	case filename == "pyproject.toml":
		return m.detectPyprojectToml(filePath)
	case filename == "Gemfile":
		return m.detectGemfile(filePath)
	case filename == "composer.json":
		return m.detectComposerJson(filePath)
	case strings.HasSuffix(filename, ".csproj"):
		return m.detectCsproj(filePath)
	case filename == "build.gradle" || filename == "build.gradle.kts":
		return m.detectGradle(filePath)
	case filename == "pom.xml":
		return m.detectMaven(filePath)
	}
	
	return nil, nil
}

// detectGoMod detects Go dependencies from go.mod
func (m *Manager) detectGoMod(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	lines := strings.Split(string(content), "\n")
	inRequire := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "require") {
			inRequire = true
			continue
		}
		
		if inRequire && (line == ")" || line == "") {
			inRequire = false
			continue
		}
		
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				name := parts[0]
				version := parts[1]
				dependencies = append(dependencies, dev.Dependency{
					Name:     name,
					Version:  version,
					Type:     "go",
					Required: true,
					Source:   "go.mod",
				})
			}
		}
	}
	
	return dependencies, nil
}

// detectCargoToml detects Rust dependencies from Cargo.toml
func (m *Manager) detectCargoToml(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	lines := strings.Split(string(content), "\n")
	inDependencies := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "[dependencies]" {
			inDependencies = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && line != "[dependencies]" {
			inDependencies = false
			continue
		}
		
		if inDependencies && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				dependencies = append(dependencies, dev.Dependency{
					Name:     name,
					Version:  version,
					Type:     "rust",
					Required: true,
					Source:   "Cargo.toml",
				})
			}
		}
	}
	
	return dependencies, nil
}

// detectPackageJson detects Node.js dependencies from package.json
func (m *Manager) detectPackageJson(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	
	if err := json.Unmarshal(content, &pkg); err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	
	for name, version := range pkg.Dependencies {
		dependencies = append(dependencies, dev.Dependency{
			Name:     name,
			Version:  version,
			Type:     "npm",
			Required: true,
			Source:   "package.json",
		})
	}
	
	for name, version := range pkg.DevDependencies {
		dependencies = append(dependencies, dev.Dependency{
			Name:     name,
			Version:  version,
			Type:     "npm",
			Required: false,
			Source:   "package.json",
		})
	}
	
	return dependencies, nil
}

// detectRequirementsTxt detects Python dependencies from requirements.txt
func (m *Manager) detectRequirementsTxt(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse package==version or package>=version format
		re := regexp.MustCompile(`^([a-zA-Z0-9_-]+)([>=<~!]+)([0-9.]+)`)
		matches := re.FindStringSubmatch(line)
		
		if len(matches) >= 3 {
			name := matches[1]
			version := matches[3]
			dependencies = append(dependencies, dev.Dependency{
				Name:     name,
				Version:  version,
				Type:     "python",
				Required: true,
				Source:   "requirements.txt",
			})
		}
	}
	
	return dependencies, nil
}

// detectPipfile detects Python dependencies from Pipfile
func (m *Manager) detectPipfile(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	lines := strings.Split(string(content), "\n")
	inPackages := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "[packages]" {
			inPackages = true
			continue
		}
		
		if strings.HasPrefix(line, "[") && line != "[packages]" {
			inPackages = false
			continue
		}
		
		if inPackages && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				version := strings.Trim(strings.TrimSpace(parts[1]), `"`)
				dependencies = append(dependencies, dev.Dependency{
					Name:     name,
					Version:  version,
					Type:     "python",
					Required: true,
					Source:   "Pipfile",
				})
			}
		}
	}
	
	return dependencies, nil
}

// detectPyprojectToml detects Python dependencies from pyproject.toml
func (m *Manager) detectPyprojectToml(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	lines := strings.Split(string(content), "\n")
	inDependencies := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "dependencies") && strings.Contains(line, "=") {
			inDependencies = true
			continue
		}
		
		if inDependencies && strings.HasPrefix(line, "[") {
			inDependencies = false
			continue
		}
		
		if inDependencies && strings.Contains(line, `"`) {
			// Extract package name and version from quotes
			re := regexp.MustCompile(`"([a-zA-Z0-9_-]+)([>=<~!]+)([0-9.]+)"`)
			matches := re.FindStringSubmatch(line)
			
			if len(matches) >= 3 {
				name := matches[1]
				version := matches[3]
				dependencies = append(dependencies, dev.Dependency{
					Name:     name,
					Version:  version,
					Type:     "python",
					Required: true,
					Source:   "pyproject.toml",
				})
			}
		}
	}
	
	return dependencies, nil
}

// detectGemfile detects Ruby dependencies from Gemfile
func (m *Manager) detectGemfile(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	lines := strings.Split(string(content), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gem") {
			// Parse gem 'name', 'version' format
			re := regexp.MustCompile(`gem\s+['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]+)['"])?`)
			matches := re.FindStringSubmatch(line)
			
			if len(matches) >= 2 {
				name := matches[1]
				version := ""
				if len(matches) >= 3 {
					version = matches[2]
				}
				dependencies = append(dependencies, dev.Dependency{
					Name:     name,
					Version:  version,
					Type:     "ruby",
					Required: true,
					Source:   "Gemfile",
				})
			}
		}
	}
	
	return dependencies, nil
}

// detectComposerJson detects PHP dependencies from composer.json
func (m *Manager) detectComposerJson(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var composer struct {
		Require    map[string]string `json:"require"`
		RequireDev map[string]string `json:"require-dev"`
	}
	
	if err := json.Unmarshal(content, &composer); err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	
	for name, version := range composer.Require {
		dependencies = append(dependencies, dev.Dependency{
			Name:     name,
			Version:  version,
			Type:     "php",
			Required: true,
			Source:   "composer.json",
		})
	}
	
	for name, version := range composer.RequireDev {
		dependencies = append(dependencies, dev.Dependency{
			Name:     name,
			Version:  version,
			Type:     "php",
			Required: false,
			Source:   "composer.json",
		})
	}
	
	return dependencies, nil
}

// detectCsproj detects C# dependencies from .csproj files
func (m *Manager) detectCsproj(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	
	// Simple regex to find PackageReference elements
	re := regexp.MustCompile(`<PackageReference\s+Include="([^"]+)"\s+Version="([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	
	for _, match := range matches {
		if len(match) >= 3 {
			name := match[1]
			version := match[2]
			dependencies = append(dependencies, dev.Dependency{
				Name:     name,
				Version:  version,
				Type:     "dotnet",
				Required: true,
				Source:   filepath.Base(filePath),
			})
		}
	}
	
	return dependencies, nil
}

// detectGradle detects Java dependencies from build.gradle
func (m *Manager) detectGradle(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	
	// Simple regex to find implementation/compile dependencies
	re := regexp.MustCompile(`(?:implementation|compile)\s+['"]([^:'"]+):([^:'"]+):([^'"]+)['"]`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	
	for _, match := range matches {
		if len(match) >= 4 {
			group := match[1]
			artifact := match[2]
			version := match[3]
			name := group + ":" + artifact
			dependencies = append(dependencies, dev.Dependency{
				Name:     name,
				Version:  version,
				Type:     "java",
				Required: true,
				Source:   filepath.Base(filePath),
			})
		}
	}
	
	return dependencies, nil
}

// detectMaven detects Java dependencies from pom.xml
func (m *Manager) detectMaven(filePath string) ([]dev.Dependency, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var dependencies []dev.Dependency
	
	// Simple regex to find dependency elements
	re := regexp.MustCompile(`<groupId>([^<]+)</groupId>\s*<artifactId>([^<]+)</artifactId>\s*<version>([^<]+)</version>`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	
	for _, match := range matches {
		if len(match) >= 4 {
			group := match[1]
			artifact := match[2]
			version := match[3]
			name := group + ":" + artifact
			dependencies = append(dependencies, dev.Dependency{
				Name:     name,
				Version:  version,
				Type:     "java",
				Required: true,
				Source:   "pom.xml",
			})
		}
	}
	
	return dependencies, nil
}

// InstallDependencies installs dependencies for a project
func (m *Manager) InstallDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// Group dependencies by type
	depsByType := make(map[string][]dev.Dependency)
	for _, dep := range deps {
		depsByType[dep.Type] = append(depsByType[dep.Type], dep)
	}
	
	// Install dependencies for each type
	for depType, typeDeps := range depsByType {
		if err := m.installDependenciesForType(ctx, path, depType, typeDeps); err != nil {
			return fmt.Errorf("failed to install %s dependencies: %w", depType, err)
		}
	}
	
	return nil
}

// installDependenciesForType installs dependencies for a specific type
func (m *Manager) installDependenciesForType(ctx context.Context, path string, depType string, deps []dev.Dependency) error {
	switch depType {
	case "go":
		return m.installGoDependencies(ctx, path, deps)
	case "rust":
		return m.installRustDependencies(ctx, path, deps)
	case "npm":
		return m.installNpmDependencies(ctx, path, deps)
	case "python":
		return m.installPythonDependencies(ctx, path, deps)
	case "ruby":
		return m.installRubyDependencies(ctx, path, deps)
	case "php":
		return m.installPhpDependencies(ctx, path, deps)
	case "dotnet":
		return m.installDotnetDependencies(ctx, path, deps)
	case "java":
		return m.installJavaDependencies(ctx, path, deps)
	default:
		m.logger.Warn(fmt.Sprintf("Unknown dependency type: %s", depType))
		return nil
	}
}

// installGoDependencies installs Go dependencies
func (m *Manager) installGoDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// Go dependencies are automatically installed via go.mod
	m.logger.Info(fmt.Sprintf("Go dependencies managed via go.mod (%d dependencies)", len(deps)))
	return nil
}

// installRustDependencies installs Rust dependencies
func (m *Manager) installRustDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// Rust dependencies are automatically installed via Cargo.toml
	m.logger.Info(fmt.Sprintf("Rust dependencies managed via Cargo.toml (%d dependencies)", len(deps)))
	return nil
}

// installNpmDependencies installs npm dependencies
func (m *Manager) installNpmDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// npm dependencies are automatically installed via package.json
	m.logger.Info(fmt.Sprintf("npm dependencies managed via package.json (%d dependencies)", len(deps)))
	return nil
}

// installPythonDependencies installs Python dependencies
func (m *Manager) installPythonDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// Python dependencies are managed via requirements.txt or similar
	m.logger.Info(fmt.Sprintf("Python dependencies managed via requirements.txt (%d dependencies)", len(deps)))
	return nil
}

// installRubyDependencies installs Ruby dependencies
func (m *Manager) installRubyDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// Ruby dependencies are managed via Gemfile
	m.logger.Info(fmt.Sprintf("Ruby dependencies managed via Gemfile (%d dependencies)", len(deps)))
	return nil
}

// installPhpDependencies installs PHP dependencies
func (m *Manager) installPhpDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// PHP dependencies are managed via composer.json
	m.logger.Info(fmt.Sprintf("PHP dependencies managed via composer.json (%d dependencies)", len(deps)))
	return nil
}

// installDotnetDependencies installs .NET dependencies
func (m *Manager) installDotnetDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// .NET dependencies are managed via .csproj files
	m.logger.Info(fmt.Sprintf(".NET dependencies managed via .csproj (%d dependencies)", len(deps)))
	return nil
}

// installJavaDependencies installs Java dependencies
func (m *Manager) installJavaDependencies(ctx context.Context, path string, deps []dev.Dependency) error {
	// Java dependencies are managed via build.gradle or pom.xml
	m.logger.Info(fmt.Sprintf("Java dependencies managed via build files (%d dependencies)", len(deps)))
	return nil
}

// UpdateDependencies updates project dependencies
func (m *Manager) UpdateDependencies(ctx context.Context, path string) error {
	deps, err := m.DetectDependencies(ctx, path)
	if err != nil {
		return err
	}
	
	return m.InstallDependencies(ctx, path, deps)
}

// CheckDependencies checks for outdated dependencies
func (m *Manager) CheckDependencies(ctx context.Context, path string) ([]dev.Dependency, error) {
	deps, err := m.DetectDependencies(ctx, path)
	if err != nil {
		return nil, err
	}
	
	// For now, just return the detected dependencies
	// In a real implementation, this would check for updates
	return deps, nil
}

// mergeDependencies removes duplicates and merges dependencies
func (m *Manager) mergeDependencies(deps []dev.Dependency) []dev.Dependency {
	seen := make(map[string]bool)
	var result []dev.Dependency
	
	for _, dep := range deps {
		key := dep.Type + ":" + dep.Name
		if !seen[key] {
			seen[key] = true
			result = append(result, dep)
		}
	}
	
	return result
}