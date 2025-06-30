package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/migration/detector"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// ImportManager handles configuration import operations
type ImportManager struct {
	configDir string
	logger    *logger.Logger
}

// NewImportManager creates a new import manager
func NewImportManager(configDir string, logger *logger.Logger) *ImportManager {
	if configDir == "" {
		usr, _ := os.UserHomeDir()
		configDir = filepath.Join(usr, ".config", "nixai")
	}

	return &ImportManager{
		configDir: configDir,
		logger:    logger,
	}
}

// Import commands
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import configurations and templates",
	Long: `Import configurations, templates, and settings from various sources.

Supports importing from:
- Local files and directories
- Git repositories (GitHub, GitLab, etc.)
- NixOS configurations from other systems
- Home Manager configurations
- Template archives and bundles
- Configuration migration from other Linux distributions

Examples:
  nixai import config /path/to/configuration.nix
  nixai import template https://github.com/user/nixos-config
  nixai import system /etc/nixos/
  nixai import migration /path/to/ubuntu/configs`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Import configuration command
var importConfigCmd = &cobra.Command{
	Use:   "config <source>",
	Short: "Import NixOS configuration from file or directory",
	Long: `Import NixOS configuration files from local sources.

The source can be:
- A single configuration.nix file
- A directory containing NixOS configurations
- A flake.nix file
- A hardware-configuration.nix file

Examples:
  nixai import config /etc/nixos/configuration.nix
  nixai import config /path/to/nixos-configs/
  nixai import config ./flake.nix --type flake
  nixai import config ./config.nix --merge --output ./imported-config.nix`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		output, _ := cmd.Flags().GetString("output")
		merge, _ := cmd.Flags().GetBool("merge")
		configType, _ := cmd.Flags().GetString("type")
		validate, _ := cmd.Flags().GetBool("validate")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create import manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		im := NewImportManager("", log)

		fmt.Println(utils.FormatHeader("📥 Importing Configuration: " + source))
		fmt.Println()

		err = im.ImportConfiguration(source, output, merge, configType, validate)
		if err != nil {
			fmt.Println(utils.FormatError("Error importing configuration: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Configuration imported successfully!"))
		if output != "" {
			fmt.Println(utils.FormatKeyValue("Output", output))
		}
		fmt.Println()
		fmt.Println(utils.FormatTip("Review the imported configuration before rebuilding"))
	},
}

// Import template command
var importTemplateCmd = &cobra.Command{
	Use:   "template <source>",
	Short: "Import template from URL or archive",
	Long: `Import configuration templates from various sources.

The source can be:
- GitHub repository URL
- Git repository URL
- Archive file (.tar.gz, .zip)
- Local template directory

Examples:
  nixai import template https://github.com/user/nixos-templates
  nixai import template https://github.com/user/config/blob/main/desktop.nix
  nixai import template ./template.tar.gz
  nixai import template /path/to/template/dir --name my-template`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		name, _ := cmd.Flags().GetString("name")
		category, _ := cmd.Flags().GetString("category")
		force, _ := cmd.Flags().GetBool("force")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create import manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		im := NewImportManager("", log)

		fmt.Println(utils.FormatHeader("📦 Importing Template: " + source))
		fmt.Println()

		err = im.ImportTemplate(source, name, category, force)
		if err != nil {
			fmt.Println(utils.FormatError("Error importing template: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Template imported successfully!"))
		if name != "" {
			fmt.Println(utils.FormatKeyValue("Template Name", name))
		}
		if category != "" {
			fmt.Println(utils.FormatKeyValue("Category", category))
		}
	},
}

// Import system command
var importSystemCmd = &cobra.Command{
	Use:   "system <path>",
	Short: "Import complete system configuration",
	Long: `Import a complete NixOS system configuration from a directory.

This command will:
- Scan the directory for configuration files
- Import hardware configuration
- Import main configuration
- Import any additional modules
- Preserve existing file structure

Examples:
  nixai import system /etc/nixos/
  nixai import system /mnt/nixos/etc/nixos/ --from-rescue
  nixai import system ./nixos-configs/ --preserve-structure`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		systemPath := args[0]
		output, _ := cmd.Flags().GetString("output")
		preserveStructure, _ := cmd.Flags().GetBool("preserve-structure")
		includeSecrets, _ := cmd.Flags().GetBool("include-secrets")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create import manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		im := NewImportManager("", log)

		fmt.Println(utils.FormatHeader("🖥️ Importing System Configuration: " + systemPath))
		fmt.Println()

		err = im.ImportSystem(systemPath, output, preserveStructure, includeSecrets)
		if err != nil {
			fmt.Println(utils.FormatError("Error importing system: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ System configuration imported successfully!"))
		if output != "" {
			fmt.Println(utils.FormatKeyValue("Output Directory", output))
		}
	},
}

// Import migration command
var importMigrationCmd = &cobra.Command{
	Use:   "migration <path>",
	Short: "Import configuration from other Linux distributions",
	Long: `Import and convert configuration from other Linux distributions to NixOS.

Supports migration from:
- Ubuntu/Debian configurations
- CentOS/RHEL configurations
- Arch Linux configurations
- openSUSE configurations
- Generic systemd configurations

Examples:
  nixai import migration /etc/ --from ubuntu
  nixai import migration /home/user/arch-configs/ --from arch
  nixai import migration ./server-configs/ --from centos --services-only`,
	Args: conditionalExactArgsValidator(1),
	Run: func(cmd *cobra.Command, args []string) {
		sourcePath := args[0]
		fromDistro, _ := cmd.Flags().GetString("from")
		output, _ := cmd.Flags().GetString("output")
		servicesOnly, _ := cmd.Flags().GetBool("services-only")
		interactive, _ := cmd.Flags().GetBool("interactive")

		// Load configuration
		cfg, err := config.LoadUserConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, utils.FormatError("Error loading config: "+err.Error()))
			os.Exit(1)
		}

		// Create import manager
		log := logger.NewLoggerWithLevel(cfg.LogLevel)
		im := NewImportManager("", log)

		fmt.Println(utils.FormatHeader("🔄 Migrating Configuration from " + strings.Title(fromDistro)))
		fmt.Println()

		err = im.ImportMigration(sourcePath, fromDistro, output, servicesOnly, interactive)
		if err != nil {
			fmt.Println(utils.FormatError("Error importing migration: " + err.Error()))
			os.Exit(1)
		}

		fmt.Println(utils.FormatSuccess("✅ Migration imported successfully!"))
		fmt.Println(utils.FormatTip("Review the generated configuration and adapt as needed"))
	},
}

// ImportConfiguration imports a NixOS configuration from source
func (im *ImportManager) ImportConfiguration(source, output string, merge bool, configType string, validate bool) error {
	// Validate source exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist: %s", source)
	}

	// Determine configuration type if not specified
	if configType == "" {
		configType = im.detectConfigurationType(source)
	}

	// Set default output if not specified
	if output == "" {
		if merge {
			output = "/etc/nixos/configuration.nix"
		} else {
			output = "./imported-configuration.nix"
		}
	}

	im.logger.Info(fmt.Sprintf("Importing configuration from %s to %s (type: %s)", source, output, configType))

	// Read source configuration
	content, err := im.readConfigurationContent(source)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %v", err)
	}

	// Process configuration based on type
	processedContent, err := im.processConfiguration(content, configType)
	if err != nil {
		return fmt.Errorf("failed to process configuration: %v", err)
	}

	// Validate if requested
	if validate {
		if err := im.validateConfiguration(processedContent); err != nil {
			return fmt.Errorf("configuration validation failed: %v", err)
		}
	}

	// Handle merge or direct write
	if merge {
		return im.mergeConfiguration(output, processedContent)
	} else {
		return im.writeConfiguration(output, processedContent)
	}
}

// ImportTemplate imports a template from various sources
func (im *ImportManager) ImportTemplate(source, name, category string, force bool) error {
	im.logger.Info(fmt.Sprintf("Importing template from %s", source))

	// Create template manager for template operations
	tm := NewTemplateManager(im.configDir, im.logger)

	// Determine source type and handle accordingly
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Handle URL sources
		return im.importTemplateFromURL(tm, source, name, category, force)
	} else if strings.HasSuffix(source, ".tar.gz") || strings.HasSuffix(source, ".zip") {
		// Handle archive sources
		return im.importTemplateFromArchive(tm, source, name, category, force)
	} else {
		// Handle local directory sources
		return im.importTemplateFromDirectory(tm, source, name, category, force)
	}
}

// ImportSystem imports a complete system configuration
func (im *ImportManager) ImportSystem(systemPath, output string, preserveStructure, includeSecrets bool) error {
	im.logger.Info(fmt.Sprintf("Importing system from %s", systemPath))

	// Set default output
	if output == "" {
		output = "./imported-system"
	}

	// Create output directory
	if err := os.MkdirAll(output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Scan for configuration files
	configFiles, err := im.scanSystemConfiguration(systemPath)
	if err != nil {
		return fmt.Errorf("failed to scan system configuration: %v", err)
	}

	// Import each configuration file
	for _, file := range configFiles {
		destinationPath := filepath.Join(output, filepath.Base(file))
		if preserveStructure {
			// Preserve relative path structure
			relPath, _ := filepath.Rel(systemPath, file)
			destinationPath = filepath.Join(output, relPath)
		}

		// Create destination directory if needed
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}

		// Copy and process file
		if err := im.copyAndProcessConfigFile(file, destinationPath, includeSecrets); err != nil {
			im.logger.Warn(fmt.Sprintf("Failed to process file %s: %v", file, err))
			continue
		}
	}

	return nil
}

// ImportMigration imports configuration from other Linux distributions
func (im *ImportManager) ImportMigration(sourcePath, fromDistro, output string, servicesOnly, interactive bool) error {
	im.logger.Info(fmt.Sprintf("Migrating configuration from %s distribution", fromDistro))

	// Create config extractor for migration
	extractor := detector.NewConfigExtractor(*im.logger)

	// Set default output
	if output == "" {
		output = "./migrated-configuration.nix"
	}

	// Scan source for configuration files
	configFiles, err := im.scanMigrationSources(sourcePath, fromDistro)
	if err != nil {
		return fmt.Errorf("failed to scan migration sources: %v", err)
	}

	// Extract and convert configurations
	var nixosConfig strings.Builder
	nixosConfig.WriteString("# Migrated configuration from " + fromDistro + "\n")
	nixosConfig.WriteString("{ config, pkgs, ... }:\n\n{\n")

	for _, file := range configFiles {
		extraction, err := extractor.ExtractConfiguration(file)
		if err != nil {
			im.logger.Warn(fmt.Sprintf("Failed to extract %s: %v", file.Path, err))
			continue
		}

		// Convert to NixOS configuration
		nixConfig, err := im.convertToNixOS(extraction, fromDistro, servicesOnly)
		if err != nil {
			im.logger.Warn(fmt.Sprintf("Failed to convert %s: %v", file.Path, err))
			continue
		}

		if nixConfig != "" {
			nixosConfig.WriteString("\n  # From " + file.Path + "\n")
			nixosConfig.WriteString(nixConfig)
		}
	}

	nixosConfig.WriteString("\n}\n")

	// Write the migrated configuration
	if err := os.WriteFile(output, []byte(nixosConfig.String()), 0644); err != nil {
		return fmt.Errorf("failed to write migrated configuration: %v", err)
	}

	return nil
}

// Helper methods

func (im *ImportManager) detectConfigurationType(source string) string {
	if strings.HasSuffix(source, "flake.nix") {
		return "flake"
	}
	if strings.HasSuffix(source, "hardware-configuration.nix") {
		return "hardware"
	}
	if strings.HasSuffix(source, "home.nix") {
		return "home-manager"
	}
	return "nixos"
}

func (im *ImportManager) readConfigurationContent(source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		// Handle directory - look for configuration.nix
		configPath := filepath.Join(source, "configuration.nix")
		if _, err := os.Stat(configPath); err == nil {
			source = configPath
		} else {
			return "", fmt.Errorf("no configuration.nix found in directory")
		}
	}

	content, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (im *ImportManager) processConfiguration(content, configType string) (string, error) {
	// Add processing based on configuration type
	switch configType {
	case "flake":
		return im.processFlakeConfiguration(content)
	case "hardware":
		return im.processHardwareConfiguration(content)
	case "home-manager":
		return im.processHomeManagerConfiguration(content)
	default:
		return content, nil
	}
}

func (im *ImportManager) processFlakeConfiguration(content string) (string, error) {
	// Process flake.nix content
	return content, nil
}

func (im *ImportManager) processHardwareConfiguration(content string) (string, error) {
	// Process hardware-configuration.nix content
	return content, nil
}

func (im *ImportManager) processHomeManagerConfiguration(content string) (string, error) {
	// Process home-manager configuration
	return content, nil
}

func (im *ImportManager) validateConfiguration(content string) error {
	// Basic syntax validation
	if !strings.Contains(content, "{") || !strings.Contains(content, "}") {
		return fmt.Errorf("configuration appears to be malformed")
	}
	return nil
}

func (im *ImportManager) mergeConfiguration(output, newContent string) error {
	// Read existing configuration
	existingContent := ""
	if content, err := os.ReadFile(output); err == nil {
		existingContent = string(content)
	}

	// Simple merge - in a real implementation, this would be more sophisticated
	merged := existingContent + "\n\n# Imported configuration\n" + newContent

	return im.writeConfiguration(output, merged)
}

func (im *ImportManager) writeConfiguration(output, content string) error {
	// Create directory if needed
	dir := filepath.Dir(output)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write configuration
	if err := os.WriteFile(output, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

func (im *ImportManager) importTemplateFromURL(tm *TemplateManager, source, name, category string, force bool) error {
	// Handle GitHub URLs, Git repositories, etc.
	if strings.Contains(source, "github.com") {
		return im.importFromGitHub(tm, source, name, category, force)
	}
	return fmt.Errorf("URL import not yet implemented for: %s", source)
}

func (im *ImportManager) importTemplateFromArchive(tm *TemplateManager, source, name, category string, force bool) error {
	// Handle tar.gz, zip archives
	return fmt.Errorf("archive import not yet implemented")
}

func (im *ImportManager) importTemplateFromDirectory(tm *TemplateManager, source, name, category string, force bool) error {
	// Handle local directory import
	return fmt.Errorf("directory import not yet implemented")
}

func (im *ImportManager) importFromGitHub(tm *TemplateManager, source, name, category string, force bool) error {
	// Parse GitHub URL and fetch content
	return fmt.Errorf("GitHub import not yet implemented")
}

func (im *ImportManager) scanSystemConfiguration(systemPath string) ([]string, error) {
	var configFiles []string

	err := filepath.Walk(systemPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		if strings.HasSuffix(path, ".nix") {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return configFiles, err
}

func (im *ImportManager) copyAndProcessConfigFile(source, destination string, includeSecrets bool) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	// Process content (remove secrets if not including them)
	processedContent := string(content)
	if !includeSecrets {
		processedContent = im.removeSensitiveContent(processedContent)
	}

	return os.WriteFile(destination, []byte(processedContent), 0644)
}

func (im *ImportManager) removeSensitiveContent(content string) string {
	// Remove passwords, keys, and other sensitive content
	lines := strings.Split(content, "\n")
	var filtered []string

	for _, line := range lines {
		if im.isSensitiveLine(line) {
			filtered = append(filtered, "  # SENSITIVE CONTENT REMOVED")
		} else {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}

func (im *ImportManager) isSensitiveLine(line string) bool {
	sensitive := []string{"password", "secret", "key", "token", "credential"}
	lowerLine := strings.ToLower(line)

	for _, keyword := range sensitive {
		if strings.Contains(lowerLine, keyword) {
			return true
		}
	}

	return false
}

func (im *ImportManager) scanMigrationSources(sourcePath, fromDistro string) ([]detector.ConfigFileInfo, error) {
	var configFiles []detector.ConfigFileInfo

	// Define distribution-specific config paths
	paths := map[string][]string{
		"ubuntu": {"/etc/apt/", "/etc/systemd/", "/etc/nginx/", "/etc/apache2/"},
		"arch":   {"/etc/pacman.conf", "/etc/systemd/", "/etc/nginx/", "/etc/httpd/"},
		"centos": {"/etc/yum.conf", "/etc/systemd/", "/etc/nginx/", "/etc/httpd/"},
	}

	searchPaths, exists := paths[fromDistro]
	if !exists {
		searchPaths = []string{"/etc/"}
	}

	for _, searchPath := range searchPaths {
		fullPath := filepath.Join(sourcePath, strings.TrimPrefix(searchPath, "/"))
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() {
				return nil
			}

			// Determine config file type
			fileType := im.determineConfigFileType(path)
			if fileType != "" {
				configFiles = append(configFiles, detector.ConfigFileInfo{
					Path:       path,
					Type:       fileType,
					Importance: "medium",
				})
			}

			return nil
		})

		if err != nil {
			im.logger.Warn(fmt.Sprintf("Error walking path %s: %v", fullPath, err))
		}
	}

	return configFiles, nil
}

func (im *ImportManager) determineConfigFileType(path string) string {
	filename := filepath.Base(path)
	dir := filepath.Dir(path)

	if strings.Contains(dir, "nginx") {
		return "nginx"
	}
	if strings.Contains(dir, "apache") {
		return "apache"
	}
	if strings.Contains(dir, "systemd") {
		return "systemd"
	}
	if filename == "pacman.conf" {
		return "package-manager"
	}
	if strings.Contains(dir, "ssh") {
		return "ssh"
	}

	return ""
}

func (im *ImportManager) convertToNixOS(extraction detector.ConfigExtraction, fromDistro string, servicesOnly bool) (string, error) {
	var config strings.Builder

	switch extraction.ServiceType {
	case "nginx":
		config.WriteString("  services.nginx = {\n")
		config.WriteString("    enable = true;\n")
		if serverName, exists := extraction.ParsedData["server_name"]; exists {
			config.WriteString(fmt.Sprintf("    virtualHosts.\"%s\" = {};\n", serverName))
		}
		config.WriteString("  };\n")

	case "apache":
		config.WriteString("  services.httpd = {\n")
		config.WriteString("    enable = true;\n")
		if docRoot, exists := extraction.ParsedData["document_root"]; exists {
			config.WriteString(fmt.Sprintf("    documentRoot = \"%s\";\n", docRoot))
		}
		config.WriteString("  };\n")

	case "ssh":
		config.WriteString("  services.openssh = {\n")
		config.WriteString("    enable = true;\n")
		config.WriteString("    settings.PasswordAuthentication = false;\n")
		config.WriteString("  };\n")

	case "systemd":
		// Convert systemd services
		config.WriteString("  # systemd service configuration\n")

	default:
		// Generic conversion
		if !servicesOnly {
			config.WriteString(fmt.Sprintf("  # Configuration from %s\n", extraction.FilePath))
		}
	}

	return config.String(), nil
}

// Add import command flags and initialization
func init() {
	// Add import subcommands
	importCmd.AddCommand(importConfigCmd)
	importCmd.AddCommand(importTemplateCmd)
	importCmd.AddCommand(importSystemCmd)
	importCmd.AddCommand(importMigrationCmd)

	// Add flags to config command
	importConfigCmd.Flags().StringP("output", "o", "", "Output file for imported configuration")
	importConfigCmd.Flags().BoolP("merge", "m", false, "Merge with existing configuration")
	importConfigCmd.Flags().StringP("type", "t", "", "Configuration type (nixos, flake, hardware, home-manager)")
	importConfigCmd.Flags().BoolP("validate", "v", true, "Validate configuration syntax")

	// Add flags to template command
	importTemplateCmd.Flags().String("name", "", "Name for the imported template")
	importTemplateCmd.Flags().StringP("category", "c", "", "Category for the template")
	importTemplateCmd.Flags().BoolP("force", "f", false, "Force import even if template exists")

	// Add flags to system command
	importSystemCmd.Flags().StringP("output", "o", "", "Output directory for system configuration")
	importSystemCmd.Flags().BoolP("preserve-structure", "p", false, "Preserve original directory structure")
	importSystemCmd.Flags().BoolP("include-secrets", "s", false, "Include sensitive configuration (use with caution)")

	// Add flags to migration command
	importMigrationCmd.Flags().StringP("from", "f", "", "Source distribution (ubuntu, arch, centos, opensuse)")
	importMigrationCmd.Flags().StringP("output", "o", "", "Output file for migrated configuration")
	importMigrationCmd.Flags().BoolP("services-only", "s", false, "Import only service configurations")
	importMigrationCmd.Flags().BoolP("interactive", "i", false, "Interactive migration with prompts")
}

// NewImportCommand creates a new import command for TUI mode
func NewImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   importCmd.Use,
		Short: importCmd.Short,
		Long:  importCmd.Long,
		Run:   importCmd.Run,
	}

	// Add subcommands
	cmd.AddCommand(importConfigCmd)
	cmd.AddCommand(importTemplateCmd)
	cmd.AddCommand(importSystemCmd)
	cmd.AddCommand(importMigrationCmd)

	// Copy flags
	cmd.PersistentFlags().AddFlagSet(importCmd.PersistentFlags())
	cmd.Flags().AddFlagSet(importCmd.Flags())

	return cmd
}
