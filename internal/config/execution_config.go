package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"nix-ai-help/pkg/logger"
)

// ExecutionConfigManager manages execution configuration validation and persistence
type ExecutionConfigManager struct {
	configPath string
	logger     *logger.Logger
}

// NewExecutionConfigManager creates a new execution configuration manager
func NewExecutionConfigManager(configPath string, logger *logger.Logger) *ExecutionConfigManager {
	return &ExecutionConfigManager{
		configPath: configPath,
		logger:     logger,
	}
}

// LoadExecutionConfig loads execution configuration from file
func (ecm *ExecutionConfigManager) LoadExecutionConfig() (*ExecutionConfig, error) {
	if _, err := os.Stat(ecm.configPath); os.IsNotExist(err) {
		ecm.logger.Info("Execution config file not found, using defaults")
		return ecm.GetDefaultExecutionConfig(), nil
	}

	data, err := os.ReadFile(ecm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read execution config file: %w", err)
	}

	var config ExecutionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse execution config: %w", err)
	}

	// Validate the loaded configuration
	if err := ecm.ValidateExecutionConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid execution config: %w", err)
	}

	ecm.logger.Info("Execution configuration loaded successfully")
	return &config, nil
}

// SaveExecutionConfig saves execution configuration to file
func (ecm *ExecutionConfigManager) SaveExecutionConfig(config *ExecutionConfig) error {
	// Validate before saving
	if err := ecm.ValidateExecutionConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(ecm.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(ecm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	ecm.logger.Info("Execution configuration saved successfully")
	return nil
}

// ValidateExecutionConfig validates an execution configuration
func (ecm *ExecutionConfigManager) ValidateExecutionConfig(config *ExecutionConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate max execution time
	if config.MaxExecutionTime < 0 {
		return fmt.Errorf("max execution time cannot be negative")
	}
	if config.MaxExecutionTime > 24*time.Hour {
		return fmt.Errorf("max execution time cannot exceed 24 hours")
	}

	// Validate allowed commands
	for i, cmd := range config.AllowedCommands {
		if cmd == "" {
			return fmt.Errorf("allowed command at index %d cannot be empty", i)
		}
		if strings.Contains(cmd, " ") {
			return fmt.Errorf("allowed command '%s' cannot contain spaces", cmd)
		}
	}

	// Validate forbidden commands
	for i, cmd := range config.ForbiddenCommands {
		if cmd == "" {
			return fmt.Errorf("forbidden command at index %d cannot be empty", i)
		}
		if strings.Contains(cmd, " ") {
			return fmt.Errorf("forbidden command '%s' cannot contain spaces", cmd)
		}
	}

	// Check for conflicts between allowed and forbidden commands
	for _, allowed := range config.AllowedCommands {
		for _, forbidden := range config.ForbiddenCommands {
			if ecm.commandsConflict(allowed, forbidden) {
				return fmt.Errorf("command '%s' appears in both allowed and forbidden lists", allowed)
			}
		}
	}

	// Validate sudo commands are in allowed commands
	for _, sudoCmd := range config.SudoCommands {
		if !ecm.isCommandInList(sudoCmd, config.AllowedCommands) {
			return fmt.Errorf("sudo command '%s' must be in allowed commands list", sudoCmd)
		}
	}

	// Validate directories
	for i, dir := range config.AllowedDirectories {
		if dir == "" {
			return fmt.Errorf("allowed directory at index %d cannot be empty", i)
		}
		if !filepath.IsAbs(dir) {
			return fmt.Errorf("allowed directory '%s' must be absolute path", dir)
		}
	}

	for i, path := range config.ForbiddenPaths {
		if path == "" {
			return fmt.Errorf("forbidden path at index %d cannot be empty", i)
		}
		if !filepath.IsAbs(path) {
			return fmt.Errorf("forbidden path '%s' must be absolute path", path)
		}
	}

	// Check for conflicts between allowed directories and forbidden paths
	for _, allowedDir := range config.AllowedDirectories {
		for _, forbiddenPath := range config.ForbiddenPaths {
			if ecm.pathsConflict(allowedDir, forbiddenPath) {
				return fmt.Errorf("path conflict: '%s' in allowed directories conflicts with forbidden path '%s'", allowedDir, forbiddenPath)
			}
		}
	}

	// Validate environment variables
	for i, envVar := range config.AllowedEnvironmentVariables {
		if envVar == "" {
			return fmt.Errorf("allowed environment variable at index %d cannot be empty", i)
		}
		if !ecm.isValidEnvironmentVariableName(envVar) {
			return fmt.Errorf("invalid environment variable name: '%s'", envVar)
		}
	}

	// Validate categories
	validCategoryNames := []string{"package", "system", "configuration", "development", "utility"}
	for categoryName, categoryConfig := range config.Categories {
		if !ecm.isValidCategoryName(categoryName, validCategoryNames) {
			return fmt.Errorf("invalid category name: '%s'", categoryName)
		}

		if err := ecm.validateCategoryConfig(categoryConfig); err != nil {
			return fmt.Errorf("invalid category config for '%s': %w", categoryName, err)
		}
	}

	return nil
}

// commandsConflict checks if two command patterns conflict
func (ecm *ExecutionConfigManager) commandsConflict(cmd1, cmd2 string) bool {
	// Exact match
	if cmd1 == cmd2 {
		return true
	}

	// Check wildcard patterns
	if strings.HasSuffix(cmd1, "*") {
		prefix := strings.TrimSuffix(cmd1, "*")
		if strings.HasPrefix(cmd2, prefix) {
			return true
		}
	}

	if strings.HasSuffix(cmd2, "*") {
		prefix := strings.TrimSuffix(cmd2, "*")
		if strings.HasPrefix(cmd1, prefix) {
			return true
		}
	}

	return false
}

// isCommandInList checks if a command is in a list (supporting wildcards)
func (ecm *ExecutionConfigManager) isCommandInList(command string, list []string) bool {
	for _, listCmd := range list {
		if ecm.commandsConflict(command, listCmd) {
			return true
		}
	}
	return false
}

// pathsConflict checks if directory paths conflict
func (ecm *ExecutionConfigManager) pathsConflict(allowedDir, forbiddenPath string) bool {
	// Clean paths
	allowedDir = filepath.Clean(allowedDir)
	forbiddenPath = filepath.Clean(forbiddenPath)

	// Check if paths are the same or overlap
	return strings.HasPrefix(allowedDir, forbiddenPath) || strings.HasPrefix(forbiddenPath, allowedDir)
}

// isValidEnvironmentVariableName validates environment variable names
func (ecm *ExecutionConfigManager) isValidEnvironmentVariableName(name string) bool {
	if name == "" {
		return false
	}

	// Environment variable names must start with letter or underscore
	first := name[0]
	if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z') || first == '_') {
		return false
	}

	// Subsequent characters must be letters, digits, or underscores
	for i := 1; i < len(name); i++ {
		char := name[i]
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// isValidCategoryName checks if a category name is valid
func (ecm *ExecutionConfigManager) isValidCategoryName(name string, validNames []string) bool {
	for _, validName := range validNames {
		if name == validName {
			return true
		}
	}
	return false
}

// validateCategoryConfig validates a category configuration
func (ecm *ExecutionConfigManager) validateCategoryConfig(config ExecutionCategoryConfig) error {
	// Validate commands in category
	for i, cmd := range config.Commands {
		if cmd == "" {
			return fmt.Errorf("command at index %d cannot be empty", i)
		}
	}

	// Validate max execution time
	if config.MaxExecutionTime < 0 {
		return fmt.Errorf("max execution time cannot be negative")
	}

	// Validate allowed directories
	for i, dir := range config.AllowedDirectories {
		if dir != "" && !filepath.IsAbs(dir) {
			return fmt.Errorf("allowed directory '%s' at index %d must be absolute path", dir, i)
		}
	}

	return nil
}

// GetDefaultExecutionConfig returns default execution configuration
func (ecm *ExecutionConfigManager) GetDefaultExecutionConfig() *ExecutionConfig {
	return &ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        true,
		ConfirmationRequired: true,
		MaxExecutionTime:     5 * time.Minute,
		AllowedCommands: []string{
			"nix", "nix-env", "nix-shell", "nix-store", "nix-build",
			"nix-collect-garbage", "nix-channel", "nix-prefetch-url",
			"nixos-rebuild", "systemctl", "journalctl", "nixos-version",
			"nixos-option", "nixos-generate-config",
			"ls", "cat", "grep", "find", "which", "whereis",
			"ps", "top", "htop", "df", "du", "free", "uptime",
		},
		ForbiddenCommands: []string{
			"rm", "rmdir", "dd", "mkfs", "fdisk", "parted",
			"shutdown", "reboot", "halt", "poweroff",
			"crontab", "su", "sudo",
		},
		SudoCommands: []string{
			"nixos-rebuild", "systemctl",
		},
		AllowedDirectories: []string{
			"/home", "/tmp", "/etc/nixos", "/nix/store", "/var/log",
		},
		ForbiddenPaths: []string{
			"/boot", "/dev", "/proc", "/sys", "/root",
		},
		AllowedEnvironmentVariables: []string{
			"PATH", "HOME", "USER", "LANG", "LC_ALL",
			"NIX_PATH", "NIXPKGS_CONFIG", "NIXOS_CONFIG",
		},
		Categories: map[string]ExecutionCategoryConfig{
			"package": {
				Commands: []string{
					"nix", "nix-env", "nix-shell", "nix-store", "nix-build",
					"nix-collect-garbage", "nix-channel", "nix-prefetch-url",
				},
				RequiresConfirmation: false,
				RequiresSudo:         false,
				MaxExecutionTime:     5 * time.Minute,
				AllowedDirectories:   []string{"/home", "/tmp", "/nix/store"},
			},
			"system": {
				Commands: []string{
					"nixos-rebuild", "systemctl", "journalctl", "nixos-version",
					"nixos-option", "nixos-generate-config",
				},
				RequiresConfirmation: true,
				RequiresSudo:         true,
				MaxExecutionTime:     15 * time.Minute,
				AllowedDirectories:   []string{"/etc/nixos", "/var/log"},
			},
			"utility": {
				Commands: []string{
					"ls", "cat", "grep", "find", "which", "whereis",
					"ps", "top", "htop", "df", "du", "free", "uptime",
				},
				RequiresConfirmation: false,
				RequiresSudo:         false,
				MaxExecutionTime:     2 * time.Minute,
				AllowedDirectories:   []string{"/", "/home", "/tmp", "/var"},
			},
		},
	}
}

// MergeExecutionConfigs merges two execution configurations
func (ecm *ExecutionConfigManager) MergeExecutionConfigs(base, override *ExecutionConfig) *ExecutionConfig {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base // Copy base config

	// Override basic settings
	if override.Enabled != base.Enabled {
		merged.Enabled = override.Enabled
	}
	if override.DryRunDefault != base.DryRunDefault {
		merged.DryRunDefault = override.DryRunDefault
	}
	if override.ConfirmationRequired != base.ConfirmationRequired {
		merged.ConfirmationRequired = override.ConfirmationRequired
	}
	if override.MaxExecutionTime != base.MaxExecutionTime {
		merged.MaxExecutionTime = override.MaxExecutionTime
	}

	// Merge command lists (override completely replaces base)
	if len(override.AllowedCommands) > 0 {
		merged.AllowedCommands = override.AllowedCommands
	}
	if len(override.ForbiddenCommands) > 0 {
		merged.ForbiddenCommands = override.ForbiddenCommands
	}
	if len(override.SudoCommands) > 0 {
		merged.SudoCommands = override.SudoCommands
	}

	// Merge directory lists
	if len(override.AllowedDirectories) > 0 {
		merged.AllowedDirectories = override.AllowedDirectories
	}
	if len(override.ForbiddenPaths) > 0 {
		merged.ForbiddenPaths = override.ForbiddenPaths
	}

	// Merge environment variables
	if len(override.AllowedEnvironmentVariables) > 0 {
		merged.AllowedEnvironmentVariables = override.AllowedEnvironmentVariables
	}

	// Merge categories
	if len(override.Categories) > 0 {
		if merged.Categories == nil {
			merged.Categories = make(map[string]ExecutionCategoryConfig)
		}
		for name, config := range override.Categories {
			merged.Categories[name] = config
		}
	}

	return &merged
}

// ExportExecutionConfig exports configuration to a specified format
func (ecm *ExecutionConfigManager) ExportExecutionConfig(config *ExecutionConfig, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "yaml", "yml":
		return yaml.Marshal(config)
	case "json":
		// For JSON export, we'd need to import encoding/json
		return nil, fmt.Errorf("JSON export not implemented")
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetConfigSummary returns a summary of the configuration
func (ecm *ExecutionConfigManager) GetConfigSummary(config *ExecutionConfig) map[string]interface{} {
	return map[string]interface{}{
		"enabled":                       config.Enabled,
		"dry_run_default":               config.DryRunDefault,
		"confirmation_required":         config.ConfirmationRequired,
		"max_execution_time":            config.MaxExecutionTime.String(),
		"allowed_commands_count":        len(config.AllowedCommands),
		"forbidden_commands_count":      len(config.ForbiddenCommands),
		"sudo_commands_count":           len(config.SudoCommands),
		"allowed_directories_count":     len(config.AllowedDirectories),
		"forbidden_paths_count":         len(config.ForbiddenPaths),
		"allowed_env_variables_count":   len(config.AllowedEnvironmentVariables),
		"categories_count":              len(config.Categories),
		"category_names":                ecm.getCategoryNames(config),
	}
}

// getCategoryNames extracts category names from config
func (ecm *ExecutionConfigManager) getCategoryNames(config *ExecutionConfig) []string {
	names := make([]string, 0, len(config.Categories))
	for name := range config.Categories {
		names = append(names, name)
	}
	return names
}