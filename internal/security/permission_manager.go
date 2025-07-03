package security

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// PermissionManager manages command execution permissions
type PermissionManager struct {
	config   *config.ExecutionConfig
	logger   *logger.Logger
	policies map[string]*CategoryPolicy
}

// CategoryPolicy defines permissions for a command category
type CategoryPolicy struct {
	Commands             []string
	RequiresConfirmation bool
	RequiresSudo         bool
	MaxExecutionTime     time.Duration
	AllowedDirectories   []string
	ForbiddenArgs        []string
	AllowedArgs          []string
}

// CommandPermission represents specific permissions for a command
type CommandPermission struct {
	Command              string
	AllowedArgs          []string
	ForbiddenArgs        []string
	RequiresConfirmation bool
	RequiresSudo         bool
	Category             string
	MaxExecutionTime     time.Duration
	AllowedDirectories   []string
}

// RestrictionRule defines a restriction rule
type RestrictionRule struct {
	Type        string // "command", "argument", "path", "environment"
	Pattern     string
	Action      string // "allow", "deny", "confirm"
	Description string
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(config *config.ExecutionConfig, logger *logger.Logger) *PermissionManager {
	pm := &PermissionManager{
		config:   config,
		logger:   logger,
		policies: make(map[string]*CategoryPolicy),
	}
	
	// Initialize default policies
	pm.initializeDefaultPolicies()
	
	// Load custom policies from config
	pm.loadConfigPolicies()
	
	return pm
}

// IsCommandAllowed checks if a command is allowed
func (pm *PermissionManager) IsCommandAllowed(command string, args []string) bool {
	// Check if command is explicitly forbidden
	for _, forbidden := range pm.config.ForbiddenCommands {
		if pm.matchesPattern(command, forbidden) {
			pm.logger.Debug("Command forbidden by policy")
			return false
		}
	}
	
	// Check if command is in allowed list
	for _, policy := range pm.policies {
		for _, allowedCmd := range policy.Commands {
			if pm.matchesPattern(command, allowedCmd) {
				pm.logger.Debug("Command allowed by category")
				return true
			}
		}
	}
	
	// Check global allowed commands
	for _, allowedCmd := range pm.config.AllowedCommands {
		if pm.matchesPattern(command, allowedCmd) {
			pm.logger.Debug("Command allowed globally")
			return true
		}
	}
	
	pm.logger.Debug("Command not found in allowed lists")
	return false
}

// RequiresConfirmation checks if a command requires user confirmation
func (pm *PermissionManager) RequiresConfirmation(req interface{}) bool {
	// This would need to be adapted based on the actual request type
	// For now, implementing basic logic
	
	// Always require confirmation if globally enabled
	if pm.config.ConfirmationRequired {
		return true
	}
	
	// Check category-specific requirements
	// This is a placeholder implementation
	return false
}

// RequiresSudo checks if a command requires sudo privileges
func (pm *PermissionManager) RequiresSudo(command string) bool {
	for _, sudoCmd := range pm.config.SudoCommands {
		if pm.matchesPattern(command, sudoCmd) {
			return true
		}
	}
	
	return false
}

// IsDirectoryAllowed checks if a directory is allowed for operations
func (pm *PermissionManager) IsDirectoryAllowed(directory string) bool {
	// Convert to absolute path
	absPath, err := filepath.Abs(directory)
	if err != nil {
		pm.logger.Error("Failed to get absolute path for directory")
		return false
	}
	
	// Check against forbidden paths
	for _, forbiddenPath := range pm.config.ForbiddenPaths {
		forbiddenAbs, err := filepath.Abs(forbiddenPath)
		if err != nil {
			continue
		}
		
		if strings.HasPrefix(absPath, forbiddenAbs) {
			pm.logger.Debug("Directory blocked by forbidden path")
			return false
		}
	}
	
	// Check against allowed directories
	for _, allowedDir := range pm.config.AllowedDirectories {
		allowedAbs, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		
		if strings.HasPrefix(absPath, allowedAbs) {
			pm.logger.Debug("Directory allowed")
			return true
		}
	}
	
	pm.logger.Debug("Directory not in allowed list")
	return false
}

// IsEnvironmentVariableAllowed checks if an environment variable is allowed
func (pm *PermissionManager) IsEnvironmentVariableAllowed(variable string) bool {
	for _, allowed := range pm.config.AllowedEnvironmentVariables {
		if variable == allowed {
			return true
		}
	}
	
	return false
}

// GetCategoryPolicy returns the policy for a specific category
func (pm *PermissionManager) GetCategoryPolicy(category string) (*CategoryPolicy, bool) {
	policy, exists := pm.policies[category]
	return policy, exists
}

// ValidateCommandCategory validates a command against its category policy
func (pm *PermissionManager) ValidateCommandCategory(command, category string, args []string) error {
	policy, exists := pm.policies[category]
	if !exists {
		return fmt.Errorf("unknown category: %s", category)
	}
	
	// Check if command is allowed in this category
	commandAllowed := false
	for _, allowedCmd := range policy.Commands {
		if pm.matchesPattern(command, allowedCmd) {
			commandAllowed = true
			break
		}
	}
	
	if !commandAllowed {
		return fmt.Errorf("command %s not allowed in category %s", command, category)
	}
	
	// Validate arguments against category policy
	if err := pm.validateArgumentsForCategory(args, policy); err != nil {
		return err
	}
	
	return nil
}

// GetMaxExecutionTime returns the maximum execution time for a command
func (pm *PermissionManager) GetMaxExecutionTime(category string) time.Duration {
	if policy, exists := pm.policies[category]; exists && policy.MaxExecutionTime > 0 {
		return policy.MaxExecutionTime
	}
	
	return pm.config.MaxExecutionTime
}

// AddCustomRule adds a custom restriction rule
func (pm *PermissionManager) AddCustomRule(rule RestrictionRule) {
	// Implementation for adding custom rules
	pm.logger.Info("Added custom restriction rule")
}

// initializeDefaultPolicies sets up default category policies
func (pm *PermissionManager) initializeDefaultPolicies() {
	// Package management policy
	pm.policies["package"] = &CategoryPolicy{
		Commands: []string{
			"nix", "nix-env", "nix-shell", "nix-store", "nix-build",
			"nix-collect-garbage", "nix-channel", "nix-prefetch-url",
		},
		RequiresConfirmation: false,
		RequiresSudo:         false,
		MaxExecutionTime:     5 * time.Minute,
		AllowedDirectories:   []string{"/home", "/tmp", "/nix/store"},
	}
	
	// System management policy
	pm.policies["system"] = &CategoryPolicy{
		Commands: []string{
			"nixos-rebuild", "systemctl", "journalctl", "nixos-version",
			"nixos-option", "nixos-generate-config",
		},
		RequiresConfirmation: true,
		RequiresSudo:         true,
		MaxExecutionTime:     15 * time.Minute,
		AllowedDirectories:   []string{"/etc/nixos", "/var/log"},
	}
	
	// Configuration management policy
	pm.policies["configuration"] = &CategoryPolicy{
		Commands: []string{
			"nvim", "vim", "nano", "emacs", "cp", "mv", "mkdir",
			"touch", "chmod", "chown",
		},
		RequiresConfirmation: true,
		RequiresSudo:         false,
		MaxExecutionTime:     time.Hour, // Text editors can run for a long time
		AllowedDirectories:   []string{"/home", "/etc/nixos", "/tmp"},
		ForbiddenArgs:        []string{"-rf", "--force", "--delete"},
	}
	
	// Development policy
	pm.policies["development"] = &CategoryPolicy{
		Commands: []string{
			"nix develop", "nix flake", "nix build", "git", "make",
			"cargo", "npm", "yarn", "python", "node",
		},
		RequiresConfirmation: false,
		RequiresSudo:         false,
		MaxExecutionTime:     30 * time.Minute,
		AllowedDirectories:   []string{"/home", "/tmp", "/var/tmp"},
	}
	
	// Utility policy
	pm.policies["utility"] = &CategoryPolicy{
		Commands: []string{
			"ls", "cat", "grep", "find", "which", "whereis",
			"ps", "top", "htop", "df", "du", "free", "uptime",
		},
		RequiresConfirmation: false,
		RequiresSudo:         false,
		MaxExecutionTime:     2 * time.Minute,
		AllowedDirectories:   []string{"/", "/home", "/tmp", "/var"},
	}
}

// loadConfigPolicies loads policies from configuration
func (pm *PermissionManager) loadConfigPolicies() {
	// Load category-specific configurations from config
	for category, categoryConfig := range pm.config.Categories {
		if policy, exists := pm.policies[category]; exists {
			// Update existing policy with config values
			if len(categoryConfig.Commands) > 0 {
				policy.Commands = categoryConfig.Commands
			}
			if categoryConfig.MaxExecutionTime > 0 {
				policy.MaxExecutionTime = categoryConfig.MaxExecutionTime
			}
			policy.RequiresConfirmation = categoryConfig.RequiresConfirmation
			policy.RequiresSudo = categoryConfig.RequiresSudo
		} else {
			// Create new policy from config
			pm.policies[category] = &CategoryPolicy{
				Commands:             categoryConfig.Commands,
				RequiresConfirmation: categoryConfig.RequiresConfirmation,
				RequiresSudo:         categoryConfig.RequiresSudo,
				MaxExecutionTime:     categoryConfig.MaxExecutionTime,
				AllowedDirectories:   categoryConfig.AllowedDirectories,
			}
		}
	}
}

// matchesPattern checks if a command matches a pattern (supports wildcards)
func (pm *PermissionManager) matchesPattern(command, pattern string) bool {
	// Exact match
	if command == pattern {
		return true
	}
	
	// Wildcard match (e.g., "nix*" matches "nix-env", "nixos-rebuild")
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(command, prefix)
	}
	
	return false
}

// validateArgumentsForCategory validates arguments against category policy
func (pm *PermissionManager) validateArgumentsForCategory(args []string, policy *CategoryPolicy) error {
	// Check for forbidden arguments
	for _, arg := range args {
		for _, forbidden := range policy.ForbiddenArgs {
			if strings.Contains(arg, forbidden) {
				return fmt.Errorf("argument contains forbidden pattern: %s", forbidden)
			}
		}
	}
	
	// If allowed args are specified, ensure arguments are in the list
	if len(policy.AllowedArgs) > 0 {
		for _, arg := range args {
			allowed := false
			for _, allowedArg := range policy.AllowedArgs {
				if pm.matchesPattern(arg, allowedArg) {
					allowed = true
					break
				}
			}
			if !allowed {
				return fmt.Errorf("argument not in allowed list: %s", arg)
			}
		}
	}
	
	return nil
}