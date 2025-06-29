package config_builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/olafkfreund/nixai/pkg/logger"
)

// PreviewMode represents the type of preview being generated
type PreviewMode string

const (
	PreviewNixOS       PreviewMode = "nixos"
	PreviewHomeManager PreviewMode = "home-manager"
	PreviewFlake       PreviewMode = "flake"
	PreviewModule      PreviewMode = "module"
)

// PreviewOptions contains options for configuration preview
type PreviewOptions struct {
	Mode            PreviewMode `json:"mode"`
	Target          string      `json:"target"` // system, user, etc.
	Format          string      `json:"format"` // nix, json, yaml
	IncludeComments bool        `json:"include_comments"`
	MinifyOutput    bool        `json:"minify_output"`
	ValidateOnly    bool        `json:"validate_only"`
	TempDirectory   string      `json:"temp_directory"`
}

// PreviewResult contains the result of configuration generation
type PreviewResult struct {
	Configuration string           `json:"configuration"`
	Errors        []PreviewError   `json:"errors"`
	Warnings      []PreviewWarning `json:"warnings"`
	Metadata      PreviewMetadata  `json:"metadata"`
	Timestamp     time.Time        `json:"timestamp"`
	Success       bool             `json:"success"`
}

// PreviewError represents an error in the generated configuration
type PreviewError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Component  string `json:"component"`
	Suggestion string `json:"suggestion"`
}

// PreviewWarning represents a warning in the generated configuration
type PreviewWarning struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Component  string `json:"component"`
	Suggestion string `json:"suggestion"`
}

// PreviewMetadata contains metadata about the generated configuration
type PreviewMetadata struct {
	LineCount          int             `json:"line_count"`
	ComponentCount     int             `json:"component_count"`
	Size               int64           `json:"size"`
	EstimatedBuildTime string          `json:"estimated_build_time"`
	Dependencies       []string        `json:"dependencies"`
	Services           []string        `json:"services"`
	Packages           []string        `json:"packages"`
	Features           map[string]bool `json:"features"`
}

// RealTimePreview manages real-time configuration preview and validation
type RealTimePreview struct {
	canvas     *Canvas
	library    *ComponentLibrary
	logger     *logger.Logger
	options    PreviewOptions
	lastResult *PreviewResult
	tempDir    string
	validator  *ConfigValidator
}

// ConfigValidator validates NixOS configurations
type ConfigValidator struct {
	nixPath   string
	nixosPath string
	tempDir   string
	logger    *logger.Logger
}

// NewRealTimePreview creates a new real-time preview instance
func NewRealTimePreview(canvas *Canvas, library *ComponentLibrary, logger *logger.Logger) (*RealTimePreview, error) {
	tempDir, err := os.MkdirTemp("", "nixai-preview-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	validator, err := NewConfigValidator(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %w", err)
	}

	return &RealTimePreview{
		canvas:  canvas,
		library: library,
		logger:  logger,
		options: PreviewOptions{
			Mode:            PreviewNixOS,
			Target:          "system",
			Format:          "nix",
			IncludeComments: true,
			MinifyOutput:    false,
			ValidateOnly:    false,
			TempDirectory:   tempDir,
		},
		tempDir:   tempDir,
		validator: validator,
	}, nil
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator(logger *logger.Logger) (*ConfigValidator, error) {
	// Find nix and nixos-rebuild commands
	nixPath, err := exec.LookPath("nix")
	if err != nil {
		return nil, fmt.Errorf("nix command not found: %w", err)
	}

	nixosPath, err := exec.LookPath("nixos-rebuild")
	if err != nil {
		// nixos-rebuild might not be available on non-NixOS systems
		logger.Warn("nixos-rebuild not found, some validation features will be limited")
	}

	tempDir, err := os.MkdirTemp("", "nixai-validator-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &ConfigValidator{
		nixPath:   nixPath,
		nixosPath: nixosPath,
		tempDir:   tempDir,
		logger:    logger,
	}, nil
}

// GeneratePreview generates a live preview of the current configuration
func (rtp *RealTimePreview) GeneratePreview() (*PreviewResult, error) {
	rtp.logger.Debug("Generating configuration preview")

	result := &PreviewResult{
		Errors:    []PreviewError{},
		Warnings:  []PreviewWarning{},
		Timestamp: time.Now(),
		Success:   false,
	}

	// Generate configuration based on mode
	var config string
	var err error

	switch rtp.options.Mode {
	case PreviewNixOS:
		config, err = rtp.generateNixOSConfig()
	case PreviewHomeManager:
		config, err = rtp.generateHomeManagerConfig()
	case PreviewFlake:
		config, err = rtp.generateFlakeConfig()
	case PreviewModule:
		config, err = rtp.generateModuleConfig()
	default:
		return nil, fmt.Errorf("unsupported preview mode: %s", rtp.options.Mode)
	}

	if err != nil {
		result.Errors = append(result.Errors, PreviewError{
			Type:    "generation_error",
			Message: err.Error(),
		})
		rtp.lastResult = result
		return result, err
	}

	result.Configuration = config

	// Calculate metadata
	result.Metadata = rtp.calculateMetadata(config)

	// Validate configuration if not validate-only mode
	if !rtp.options.ValidateOnly {
		if err := rtp.validateConfiguration(config, result); err != nil {
			rtp.logger.Warn(fmt.Sprintf("Validation failed: %v", err))
		}
	}

	// Check for warnings
	rtp.checkForWarnings(result)

	result.Success = len(result.Errors) == 0
	rtp.lastResult = result

	rtp.logger.Debug(fmt.Sprintf("Preview generated: %d lines, %d errors, %d warnings",
		result.Metadata.LineCount, len(result.Errors), len(result.Warnings)))

	return result, nil
}

// UpdateOptions updates preview options
func (rtp *RealTimePreview) UpdateOptions(options PreviewOptions) {
	rtp.options = options
	rtp.logger.Debug(fmt.Sprintf("Updated preview options: mode=%s, format=%s", options.Mode, options.Format))
}

// GetLastResult returns the last preview result
func (rtp *RealTimePreview) GetLastResult() *PreviewResult {
	return rtp.lastResult
}

// Cleanup cleans up temporary files
func (rtp *RealTimePreview) Cleanup() error {
	if err := os.RemoveAll(rtp.tempDir); err != nil {
		return fmt.Errorf("failed to cleanup temp directory: %w", err)
	}

	if err := rtp.validator.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup validator: %w", err)
	}

	return nil
}

// Cleanup cleans up validator temporary files
func (cv *ConfigValidator) Cleanup() error {
	return os.RemoveAll(cv.tempDir)
}

// ValidateNix validates Nix syntax
func (cv *ConfigValidator) ValidateNix(config string) ([]PreviewError, error) {
	errors := []PreviewError{}

	// Write config to temporary file
	tempFile := filepath.Join(cv.tempDir, "config.nix")
	if err := os.WriteFile(tempFile, []byte(config), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Use nix-instantiate to check syntax
	cmd := exec.Command(cv.nixPath, "eval", "--file", tempFile, "--json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Parse nix error output
		errorLines := strings.Split(string(output), "\n")
		for _, line := range errorLines {
			if strings.Contains(line, "error:") {
				errors = append(errors, PreviewError{
					Type:       "syntax_error",
					Message:    strings.TrimSpace(line),
					Suggestion: "Check Nix syntax and fix any errors",
				})
			}
		}
	}

	return errors, nil
}

// ValidateNixOS validates NixOS configuration
func (cv *ConfigValidator) ValidateNixOS(config string) ([]PreviewError, []PreviewWarning, error) {
	errors := []PreviewError{}
	warnings := []PreviewWarning{}

	if cv.nixosPath == "" {
		return errors, warnings, fmt.Errorf("nixos-rebuild not available")
	}

	// Write config to temporary file
	tempFile := filepath.Join(cv.tempDir, "configuration.nix")
	if err := os.WriteFile(tempFile, []byte(config), 0644); err != nil {
		return nil, nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Use nixos-rebuild dry-run to validate
	cmd := exec.Command(cv.nixosPath, "dry-run", "--file", tempFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Parse nixos-rebuild error output
		errorLines := strings.Split(string(output), "\n")
		for _, line := range errorLines {
			if strings.Contains(line, "error:") {
				errors = append(errors, PreviewError{
					Type:       "nixos_error",
					Message:    strings.TrimSpace(line),
					Suggestion: "Fix NixOS configuration error",
				})
			} else if strings.Contains(line, "warning:") {
				warnings = append(warnings, PreviewWarning{
					Type:       "nixos_warning",
					Message:    strings.TrimSpace(line),
					Suggestion: "Consider addressing this warning",
				})
			}
		}
	}

	return errors, warnings, nil
}

// Helper methods

func (rtp *RealTimePreview) generateNixOSConfig() (string, error) {
	var config strings.Builder

	config.WriteString("# Generated by nixai Visual Configuration Builder\n")
	config.WriteString("# Timestamp: " + time.Now().Format(time.RFC3339) + "\n\n")
	config.WriteString("{ config, pkgs, ... }:\n\n")
	config.WriteString("{\n")

	// Generate imports
	imports := rtp.generateImports()
	if len(imports) > 0 {
		config.WriteString("  imports = [\n")
		for _, imp := range imports {
			config.WriteString(fmt.Sprintf("    %s\n", imp))
		}
		config.WriteString("  ];\n\n")
	}

	// Generate services
	services := rtp.generateServices()
	if len(services) > 0 {
		config.WriteString("  services = {\n")
		for _, service := range services {
			config.WriteString(fmt.Sprintf("    %s\n", service))
		}
		config.WriteString("  };\n\n")
	}

	// Generate environment packages
	packages := rtp.generatePackages()
	if len(packages) > 0 {
		config.WriteString("  environment.systemPackages = with pkgs; [\n")
		for _, pkg := range packages {
			config.WriteString(fmt.Sprintf("    %s\n", pkg))
		}
		config.WriteString("  ];\n\n")
	}

	// Generate users
	users := rtp.generateUsers()
	if len(users) > 0 {
		config.WriteString("  users.users = {\n")
		for _, user := range users {
			config.WriteString(fmt.Sprintf("    %s\n", user))
		}
		config.WriteString("  };\n\n")
	}

	// Generate networking
	networking := rtp.generateNetworking()
	if networking != "" {
		config.WriteString("  networking = {\n")
		config.WriteString(fmt.Sprintf("    %s\n", networking))
		config.WriteString("  };\n\n")
	}

	// Generate boot
	boot := rtp.generateBoot()
	if boot != "" {
		config.WriteString("  boot = {\n")
		config.WriteString(fmt.Sprintf("    %s\n", boot))
		config.WriteString("  };\n\n")
	}

	// Generate system packages and configuration
	system := rtp.generateSystemConfig()
	if system != "" {
		config.WriteString(fmt.Sprintf("  %s\n\n", system))
	}

	config.WriteString("  # System state version\n")
	config.WriteString("  system.stateVersion = \"24.05\";\n")
	config.WriteString("}\n")

	return config.String(), nil
}

func (rtp *RealTimePreview) generateHomeManagerConfig() (string, error) {
	var config strings.Builder

	config.WriteString("# Generated by nixai Visual Configuration Builder\n")
	config.WriteString("# Home Manager Configuration\n")
	config.WriteString("# Timestamp: " + time.Now().Format(time.RFC3339) + "\n\n")
	config.WriteString("{ config, pkgs, ... }:\n\n")
	config.WriteString("{\n")

	// Generate home packages
	packages := rtp.generateHomePackages()
	if len(packages) > 0 {
		config.WriteString("  home.packages = with pkgs; [\n")
		for _, pkg := range packages {
			config.WriteString(fmt.Sprintf("    %s\n", pkg))
		}
		config.WriteString("  ];\n\n")
	}

	// Generate programs
	programs := rtp.generatePrograms()
	if len(programs) > 0 {
		config.WriteString("  programs = {\n")
		for _, program := range programs {
			config.WriteString(fmt.Sprintf("    %s\n", program))
		}
		config.WriteString("  };\n\n")
	}

	// Generate services
	services := rtp.generateHomeServices()
	if len(services) > 0 {
		config.WriteString("  services = {\n")
		for _, service := range services {
			config.WriteString(fmt.Sprintf("    %s\n", service))
		}
		config.WriteString("  };\n\n")
	}

	config.WriteString("  # Home Manager state version\n")
	config.WriteString("  home.stateVersion = \"24.05\";\n")
	config.WriteString("}\n")

	return config.String(), nil
}

func (rtp *RealTimePreview) generateFlakeConfig() (string, error) {
	var config strings.Builder

	config.WriteString("# Generated by nixai Visual Configuration Builder\n")
	config.WriteString("# Nix Flake Configuration\n")
	config.WriteString("# Timestamp: " + time.Now().Format(time.RFC3339) + "\n\n")
	config.WriteString("{\n")
	config.WriteString("  description = \"NixOS configuration generated by nixai\";\n\n")

	config.WriteString("  inputs = {\n")
	config.WriteString("    nixpkgs.url = \"github:NixOS/nixpkgs/nixos-unstable\";\n")
	config.WriteString("    home-manager = {\n")
	config.WriteString("      url = \"github:nix-community/home-manager\";\n")
	config.WriteString("      inputs.nixpkgs.follows = \"nixpkgs\";\n")
	config.WriteString("    };\n")
	config.WriteString("  };\n\n")

	config.WriteString("  outputs = { self, nixpkgs, home-manager, ... }:\n")
	config.WriteString("    let\n")
	config.WriteString("      system = \"x86_64-linux\";\n")
	config.WriteString("      pkgs = nixpkgs.legacyPackages.${system};\n")
	config.WriteString("    in\n")
	config.WriteString("    {\n")
	config.WriteString("      nixosConfigurations.default = nixpkgs.lib.nixosSystem {\n")
	config.WriteString("        inherit system;\n")
	config.WriteString("        modules = [\n")
	config.WriteString("          ./configuration.nix\n")
	config.WriteString("        ];\n")
	config.WriteString("      };\n")
	config.WriteString("    };\n")
	config.WriteString("}\n")

	return config.String(), nil
}

func (rtp *RealTimePreview) generateModuleConfig() (string, error) {
	var config strings.Builder

	config.WriteString("# Generated by nixai Visual Configuration Builder\n")
	config.WriteString("# NixOS Module\n")
	config.WriteString("# Timestamp: " + time.Now().Format(time.RFC3339) + "\n\n")
	config.WriteString("{ config, lib, pkgs, ... }:\n\n")
	config.WriteString("with lib;\n\n")
	config.WriteString("{\n")
	config.WriteString("  options = {\n")
	config.WriteString("    # Module options will be generated here\n")
	config.WriteString("  };\n\n")
	config.WriteString("  config = {\n")
	config.WriteString("    # Module configuration will be generated here\n")
	config.WriteString("  };\n")
	config.WriteString("}\n")

	return config.String(), nil
}

func (rtp *RealTimePreview) generateImports() []string {
	imports := []string{}

	// Add hardware configuration
	imports = append(imports, "./hardware-configuration.nix")

	// Add component-specific imports
	for _, placedComp := range rtp.canvas.Components {
		if placedComp.Component.Type == ComponentModule {
			imports = append(imports, fmt.Sprintf("./%s.nix", placedComp.Component.ID))
		}
	}

	return imports
}

func (rtp *RealTimePreview) generateServices() []string {
	services := []string{}

	for _, placedComp := range rtp.canvas.Components {
		if placedComp.Component.Type == ComponentService {
			service := rtp.generateServiceConfig(placedComp)
			if service != "" {
				services = append(services, service)
			}
		}
	}

	return services
}

func (rtp *RealTimePreview) generatePackages() []string {
	packages := []string{}

	for _, placedComp := range rtp.canvas.Components {
		if placedComp.Component.Type == ComponentPackage {
			packages = append(packages, placedComp.Component.ID)
		}
	}

	return packages
}

func (rtp *RealTimePreview) generateUsers() []string {
	// Generate user configurations
	return []string{}
}

func (rtp *RealTimePreview) generateNetworking() string {
	// Generate networking configuration
	return ""
}

func (rtp *RealTimePreview) generateBoot() string {
	// Generate boot configuration
	return ""
}

func (rtp *RealTimePreview) generateSystemConfig() string {
	// Generate other system configurations
	return ""
}

func (rtp *RealTimePreview) generateHomePackages() []string {
	return rtp.generatePackages()
}

func (rtp *RealTimePreview) generatePrograms() []string {
	programs := []string{}

	for _, placedComp := range rtp.canvas.Components {
		if placedComp.Component.Category == CategoryDevelopment {
			program := rtp.generateProgramConfig(placedComp)
			if program != "" {
				programs = append(programs, program)
			}
		}
	}

	return programs
}

func (rtp *RealTimePreview) generateHomeServices() []string {
	return rtp.generateServices()
}

func (rtp *RealTimePreview) generateServiceConfig(placedComp *PlacedComponent) string {
	config := placedComp.Component.NixExpression

	// Apply user configuration
	if len(placedComp.Config) > 0 {
		// TODO: Apply user settings to the base configuration
	}

	return config
}

func (rtp *RealTimePreview) generateProgramConfig(placedComp *PlacedComponent) string {
	return fmt.Sprintf("%s.enable = true;", placedComp.Component.ID)
}

func (rtp *RealTimePreview) calculateMetadata(config string) PreviewMetadata {
	lines := strings.Split(config, "\n")
	metadata := PreviewMetadata{
		LineCount:      len(lines),
		ComponentCount: len(rtp.canvas.Components),
		Size:           int64(len(config)),
		Dependencies:   []string{},
		Services:       []string{},
		Packages:       []string{},
		Features:       make(map[string]bool),
	}

	// Analyze configuration content
	for _, placedComp := range rtp.canvas.Components {
		switch placedComp.Component.Type {
		case ComponentService:
			metadata.Services = append(metadata.Services, placedComp.Component.Name)
		case ComponentPackage:
			metadata.Packages = append(metadata.Packages, placedComp.Component.Name)
		}

		// Add dependencies
		metadata.Dependencies = append(metadata.Dependencies, placedComp.Component.Dependencies...)

		// Add features
		for _, tag := range placedComp.Component.Tags {
			metadata.Features[tag] = true
		}
	}

	// Estimate build time based on complexity
	complexity := float64(metadata.ComponentCount) / 10.0
	if complexity < 1 {
		metadata.EstimatedBuildTime = "< 1 minute"
	} else if complexity < 5 {
		metadata.EstimatedBuildTime = "1-5 minutes"
	} else if complexity < 10 {
		metadata.EstimatedBuildTime = "5-10 minutes"
	} else {
		metadata.EstimatedBuildTime = "> 10 minutes"
	}

	return metadata
}

func (rtp *RealTimePreview) validateConfiguration(config string, result *PreviewResult) error {
	// Validate Nix syntax
	nixErrors, err := rtp.validator.ValidateNix(config)
	if err != nil {
		return fmt.Errorf("failed to validate Nix syntax: %w", err)
	}
	result.Errors = append(result.Errors, nixErrors...)

	// Validate NixOS configuration if available
	if rtp.options.Mode == PreviewNixOS {
		nixosErrors, nixosWarnings, err := rtp.validator.ValidateNixOS(config)
		if err != nil {
			rtp.logger.Warning(fmt.Sprintf("NixOS validation failed: %v", err))
		} else {
			result.Errors = append(result.Errors, nixosErrors...)
			result.Warnings = append(result.Warnings, nixosWarnings...)
		}
	}

	return nil
}

func (rtp *RealTimePreview) checkForWarnings(result *PreviewResult) {
	// Check for common issues

	// Large configuration warning
	if result.Metadata.ComponentCount > 20 {
		result.Warnings = append(result.Warnings, PreviewWarning{
			Type:       "complexity_warning",
			Message:    "Configuration is quite complex with many components",
			Suggestion: "Consider splitting into modules for better maintainability",
		})
	}

	// Missing critical components
	hasBootloader := false
	hasNetworking := false

	for _, placedComp := range rtp.canvas.Components {
		for _, tag := range placedComp.Component.Tags {
			if tag == "boot" || tag == "bootloader" {
				hasBootloader = true
			}
			if tag == "network" || tag == "networking" {
				hasNetworking = true
			}
		}
	}

	if !hasBootloader && rtp.options.Mode == PreviewNixOS {
		result.Warnings = append(result.Warnings, PreviewWarning{
			Type:       "missing_component",
			Message:    "No bootloader configuration detected",
			Suggestion: "Add a bootloader configuration (GRUB, systemd-boot, etc.)",
		})
	}

	if !hasNetworking && rtp.options.Mode == PreviewNixOS {
		result.Warnings = append(result.Warnings, PreviewWarning{
			Type:       "missing_component",
			Message:    "No networking configuration detected",
			Suggestion: "Add networking configuration for internet connectivity",
		})
	}
}
