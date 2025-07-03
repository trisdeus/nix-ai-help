package security

import (
	"testing"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

func TestNewPermissionManager(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Enabled:              true,
		AllowedCommands:      []string{"nix", "nix-env", "ls"},
		ForbiddenCommands:    []string{"rm", "dd"},
		SudoCommands:         []string{"nixos-rebuild"},
		AllowedDirectories:   []string{"/home", "/tmp"},
		ForbiddenPaths:       []string{"/boot", "/dev"},
		AllowedEnvironmentVariables: []string{"PATH", "HOME"},
	}

	pm := NewPermissionManager(config, log)
	if pm == nil {
		t.Fatal("Expected permission manager to be created")
	}

	// Test default policies are loaded
	policy, exists := pm.GetCategoryPolicy("package")
	if !exists {
		t.Error("Expected default package policy to exist")
	}
	if policy == nil {
		t.Error("Expected package policy to not be nil")
	}
}

func TestIsCommandAllowed(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		AllowedCommands:   []string{"nix", "nix-env", "ls", "nix*"},
		ForbiddenCommands: []string{"rm", "dd", "dangerous*"},
	}

	pm := NewPermissionManager(config, log)

	tests := []struct {
		name     string
		command  string
		args     []string
		expected bool
	}{
		{
			name:     "explicitly allowed command",
			command:  "nix",
			args:     []string{"build"},
			expected: true,
		},
		{
			name:     "wildcard allowed command",
			command:  "nix-build",
			args:     []string{},
			expected: true,
		},
		{
			name:     "forbidden command",
			command:  "rm",
			args:     []string{"-rf", "/"},
			expected: false,
		},
		{
			name:     "wildcard forbidden command",
			command:  "dangerous-tool",
			args:     []string{},
			expected: false,
		},
		{
			name:     "not in allowed list",
			command:  "unknown-command",
			args:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.IsCommandAllowed(tt.command, tt.args)
			if result != tt.expected {
				t.Errorf("IsCommandAllowed(%s) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestRequiresSudo(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		SudoCommands: []string{"nixos-rebuild", "systemctl", "sudo*"},
	}

	pm := NewPermissionManager(config, log)

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "explicit sudo command",
			command:  "nixos-rebuild",
			expected: true,
		},
		{
			name:     "wildcard sudo command",
			command:  "sudo-wrapper",
			expected: true,
		},
		{
			name:     "non-sudo command",
			command:  "nix-env",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.RequiresSudo(tt.command)
			if result != tt.expected {
				t.Errorf("RequiresSudo(%s) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestIsDirectoryAllowed(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		AllowedDirectories: []string{"/home", "/tmp", "/etc/nixos"},
		ForbiddenPaths:     []string{"/boot", "/dev", "/proc"},
	}

	pm := NewPermissionManager(config, log)

	tests := []struct {
		name      string
		directory string
		expected  bool
	}{
		{
			name:      "allowed directory",
			directory: "/home/user",
			expected:  true,
		},
		{
			name:      "forbidden path",
			directory: "/boot/grub",
			expected:  false,
		},
		{
			name:      "not in allowed list",
			directory: "/usr/local",
			expected:  false,
		},
		{
			name:      "exact allowed directory",
			directory: "/tmp",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.IsDirectoryAllowed(tt.directory)
			if result != tt.expected {
				t.Errorf("IsDirectoryAllowed(%s) = %v, expected %v", tt.directory, result, tt.expected)
			}
		})
	}
}

func TestIsEnvironmentVariableAllowed(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		AllowedEnvironmentVariables: []string{"PATH", "HOME", "NIX_PATH"},
	}

	pm := NewPermissionManager(config, log)

	tests := []struct {
		name     string
		variable string
		expected bool
	}{
		{
			name:     "allowed variable",
			variable: "PATH",
			expected: true,
		},
		{
			name:     "not allowed variable",
			variable: "SECRET_KEY",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.IsEnvironmentVariableAllowed(tt.variable)
			if result != tt.expected {
				t.Errorf("IsEnvironmentVariableAllowed(%s) = %v, expected %v", tt.variable, result, tt.expected)
			}
		})
	}
}

func TestValidateCommandCategory(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{}

	pm := NewPermissionManager(config, log)

	tests := []struct {
		name      string
		command   string
		category  string
		args      []string
		expectErr bool
	}{
		{
			name:      "valid package command",
			command:   "nix",
			category:  "package",
			args:      []string{"build"},
			expectErr: false,
		},
		{
			name:      "invalid category",
			command:   "nix",
			category:  "nonexistent",
			args:      []string{"build"},
			expectErr: true,
		},
		{
			name:      "command not allowed in category",
			command:   "rm",
			category:  "package",
			args:      []string{"-rf"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateCommandCategory(tt.command, tt.category, tt.args)
			if (err != nil) != tt.expectErr {
				t.Errorf("ValidateCommandCategory() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestGetMaxExecutionTime(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		MaxExecutionTime: 5 * time.Minute,
		Categories: map[string]config.ExecutionCategoryConfig{
			"package": {
				MaxExecutionTime: 10 * time.Minute,
			},
		},
	}

	pm := NewPermissionManager(config, log)

	// Test category-specific timeout
	packageTimeout := pm.GetMaxExecutionTime("package")
	if packageTimeout != 10*time.Minute {
		t.Errorf("Expected package timeout 10m, got %v", packageTimeout)
	}

	// Test default timeout for unknown category
	defaultTimeout := pm.GetMaxExecutionTime("unknown")
	if defaultTimeout != 5*time.Minute {
		t.Errorf("Expected default timeout 5m, got %v", defaultTimeout)
	}
}

func TestPatternMatching(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{}
	pm := NewPermissionManager(config, log)

	tests := []struct {
		name     string
		command  string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			command:  "nix",
			pattern:  "nix",
			expected: true,
		},
		{
			name:     "wildcard match",
			command:  "nix-env",
			pattern:  "nix*",
			expected: true,
		},
		{
			name:     "wildcard no match",
			command:  "systemctl",
			pattern:  "nix*",
			expected: false,
		},
		{
			name:     "no match",
			command:  "rm",
			pattern:  "nix",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.matchesPattern(tt.command, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern(%s, %s) = %v, expected %v", tt.command, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestArgumentValidation(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{}
	pm := NewPermissionManager(config, log)

	// Get the package policy to test argument validation
	policy := &CategoryPolicy{
		ForbiddenArgs: []string{"-rf", "--force", "--delete"},
		// No AllowedArgs set, so all non-forbidden args should be allowed
	}

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "safe arguments",
			args:      []string{"-i", "firefox"},
			expectErr: false,
		},
		{
			name:      "forbidden argument",
			args:      []string{"-rf", "/"},
			expectErr: true,
		},
		{
			name:      "empty args",
			args:      []string{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.validateArgumentsForCategory(tt.args, policy)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateArgumentsForCategory() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestPolicyManagement(t *testing.T) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		Categories: map[string]config.ExecutionCategoryConfig{
			"custom": {
				Commands:             []string{"custom-cmd"},
				RequiresConfirmation: true,
				RequiresSudo:         false,
				MaxExecutionTime:     time.Minute,
			},
		},
	}

	pm := NewPermissionManager(config, log)

	// Test that custom policy was loaded
	policy, exists := pm.GetCategoryPolicy("custom")
	if !exists {
		t.Error("Expected custom policy to exist")
	}
	if policy.RequiresConfirmation != true {
		t.Error("Expected custom policy to require confirmation")
	}

	// Test adding custom rules
	rule := RestrictionRule{
		Type:        "command",
		Pattern:     "test-command",
		Action:      "deny",
		Description: "Test rule",
	}
	pm.AddCustomRule(rule) // Should not panic
}

func BenchmarkIsCommandAllowed(b *testing.B) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{
		AllowedCommands:   []string{"nix", "nix-env", "ls", "cat", "grep"},
		ForbiddenCommands: []string{"rm", "dd"},
	}
	pm := NewPermissionManager(config, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.IsCommandAllowed("nix-env", []string{"-iA", "firefox"})
	}
}

func BenchmarkPatternMatching(b *testing.B) {
	log := logger.NewLogger()
	config := &config.ExecutionConfig{}
	pm := NewPermissionManager(config, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.matchesPattern("nix-env", "nix*")
	}
}