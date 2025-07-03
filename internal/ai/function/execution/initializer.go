package execution

import (
	"fmt"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/security"
	"nix-ai-help/pkg/logger"
)

// ExecutionInitializer manages the initialization of execution functions with required dependencies
type ExecutionInitializer struct {
	permissionManager *security.PermissionManager
	auditLogger       *security.AuditLogger
	sudoManager       *security.SudoManager
	config            *config.ExecutionConfig
	logger            *logger.Logger
}

// NewExecutionInitializer creates a new execution initializer
func NewExecutionInitializer(
	permissionManager *security.PermissionManager,
	auditLogger *security.AuditLogger,
	sudoManager *security.SudoManager,
	config *config.ExecutionConfig,
	logger *logger.Logger,
) *ExecutionInitializer {
	return &ExecutionInitializer{
		permissionManager: permissionManager,
		auditLogger:       auditLogger,
		sudoManager:       sudoManager,
		config:            config,
		logger:            logger,
	}
}

// InitializeFromConfig creates and initializes an execution function from configuration
func (ei *ExecutionInitializer) InitializeFromConfig(userConfig *config.UserConfig) (*ExecutionFunction, error) {
	// Create execution configuration if not provided
	execConfig := ei.config
	if execConfig == nil {
		execConfig = &config.ExecutionConfig{
			Enabled:              true,
			DryRunDefault:        true,
			ConfirmationRequired: true,
			MaxExecutionTime:     300000000000, // 5 minutes in nanoseconds
			AllowedCommands: []string{
				"nix", "nix-env", "nix-shell", "nix-store", "nix-build",
				"nix-collect-garbage", "nix-channel", "nix-prefetch-url",
				"nixos-rebuild", "systemctl", "journalctl", "nixos-version",
				"ls", "cat", "grep", "find", "which", "ps", "top", "df", "du",
			},
			ForbiddenCommands: []string{
				"rm", "rmdir", "dd", "mkfs", "fdisk", "parted",
				"shutdown", "reboot", "halt", "poweroff",
			},
			SudoCommands: []string{
				"nixos-rebuild", "systemctl",
			},
			AllowedDirectories: []string{
				"/home", "/tmp", "/etc/nixos", "/nix/store",
			},
			ForbiddenPaths: []string{
				"/boot", "/dev", "/proc", "/sys", "/root",
			},
			AllowedEnvironmentVariables: []string{
				"PATH", "HOME", "USER", "NIX_PATH", "NIXPKGS_CONFIG",
			},
		}
	}

	// Create security components if not provided
	if ei.permissionManager == nil {
		ei.permissionManager = security.NewPermissionManager(execConfig, ei.logger)
	}

	if ei.auditLogger == nil {
		auditLogger, err := security.NewAuditLogger("/var/log/nixai/audit.log", true, ei.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create audit logger: %w", err)
		}
		ei.auditLogger = auditLogger
	}

	if ei.sudoManager == nil {
		sudoConfig := &security.SudoConfig{
			SessionTimeout:    1800000000000, // 30 minutes in nanoseconds
			PasswordTimeout:   900000000000,  // 15 minutes in nanoseconds
			MaxAttempts:       3,
			RequirePassword:   true,
			AllowPasswordless: false,
			PreserveEnv:       []string{"PATH", "HOME", "USER"},
		}
		ei.sudoManager = security.NewSudoManager(sudoConfig, ei.auditLogger, ei.logger)
	}

	// Create and initialize execution function
	execFunc := NewExecutionFunction()
	if err := execFunc.Initialize(ei.permissionManager, ei.auditLogger, ei.sudoManager, execConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize execution function: %w", err)
	}

	ei.logger.Info("Execution function initialized successfully with security components")
	return execFunc, nil
}

// InitializeFunction initializes an existing execution function
func (ei *ExecutionInitializer) InitializeFunction(execFunc *ExecutionFunction) error {
	if execFunc == nil {
		return fmt.Errorf("execution function cannot be nil")
	}

	return execFunc.Initialize(ei.permissionManager, ei.auditLogger, ei.sudoManager, ei.config)
}

// GetDefaultConfig returns a default execution configuration
func GetDefaultConfig() *config.ExecutionConfig {
	return &config.ExecutionConfig{
		Enabled:              true,
		DryRunDefault:        true,
		ConfirmationRequired: true,
		MaxExecutionTime:     300000000000, // 5 minutes in nanoseconds
		AllowedCommands: []string{
			"nix", "nix-env", "nix-shell", "nix-store", "nix-build",
			"nix-collect-garbage", "nix-channel", "nix-prefetch-url",
			"nixos-rebuild", "systemctl", "journalctl", "nixos-version",
			"ls", "cat", "grep", "find", "which", "ps", "top", "df", "du",
		},
		ForbiddenCommands: []string{
			"rm", "rmdir", "dd", "mkfs", "fdisk", "parted",
			"shutdown", "reboot", "halt", "poweroff",
		},
		SudoCommands: []string{
			"nixos-rebuild", "systemctl",
		},
		AllowedDirectories: []string{
			"/home", "/tmp", "/etc/nixos", "/nix/store",
		},
		ForbiddenPaths: []string{
			"/boot", "/dev", "/proc", "/sys", "/root",
		},
		AllowedEnvironmentVariables: []string{
			"PATH", "HOME", "USER", "NIX_PATH", "NIXPKGS_CONFIG",
		},
		Categories: make(map[string]config.ExecutionCategoryConfig),
	}
}

// CreateSecurityComponents creates security components with default configuration
func CreateSecurityComponents(logger *logger.Logger) (*security.PermissionManager, *security.AuditLogger, *security.SudoManager, error) {
	execConfig := GetDefaultConfig()

	// Create permission manager
	permissionManager := security.NewPermissionManager(execConfig, logger)

	// Create audit logger
	auditLogger, err := security.NewAuditLogger("/var/log/nixai/audit.log", true, logger)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create audit logger: %w", err)
	}

	// Create sudo manager
	sudoConfig := &security.SudoConfig{
		SessionTimeout:    1800000000000, // 30 minutes in nanoseconds
		PasswordTimeout:   900000000000,  // 15 minutes in nanoseconds
		MaxAttempts:       3,
		RequirePassword:   true,
		AllowPasswordless: false,
		PreserveEnv:       []string{"PATH", "HOME", "USER"},
	}
	sudoManager := security.NewSudoManager(sudoConfig, auditLogger, logger)

	return permissionManager, auditLogger, sudoManager, nil
}