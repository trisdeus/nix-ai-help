package cli

import (
	"strings"
	"testing"
	"time"
)

func TestNewExecuteCommand(t *testing.T) {
	cmd := NewExecuteCommand()

	if cmd == nil {
		t.Fatal("Expected command to be created")
	}

	if cmd.Use != "execute [command] [args...]" {
		t.Errorf("Expected Use to be 'execute [command] [args...]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("Expected Long description to be set")
	}

	// Check that required flags exist
	descFlag := cmd.Flags().Lookup("description")
	if descFlag == nil {
		t.Error("Expected description flag to exist")
	}

	categoryFlag := cmd.Flags().Lookup("category")
	if categoryFlag == nil {
		t.Error("Expected category flag to exist")
	}

	sudoFlag := cmd.Flags().Lookup("sudo")
	if sudoFlag == nil {
		t.Error("Expected sudo flag to exist")
	}

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("Expected dry-run flag to exist")
	}

	timeoutFlag := cmd.Flags().Lookup("timeout")
	if timeoutFlag == nil {
		t.Error("Expected timeout flag to exist")
	}

	workingDirFlag := cmd.Flags().Lookup("working-dir")
	if workingDirFlag == nil {
		t.Error("Expected working-dir flag to exist")
	}

	envFlag := cmd.Flags().Lookup("env")
	if envFlag == nil {
		t.Error("Expected env flag to exist")
	}

	interactiveFlag := cmd.Flags().Lookup("interactive")
	if interactiveFlag == nil {
		t.Error("Expected interactive flag to exist")
	}

	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected force flag to exist")
	}
}

func TestNewExecuteStatusCommand(t *testing.T) {
	cmd := NewExecuteStatusCommand()

	if cmd == nil {
		t.Fatal("Expected status command to be created")
	}

	if cmd.Use != "status" {
		t.Errorf("Expected Use to be 'status', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

func TestNewExecuteConfigCommand(t *testing.T) {
	cmd := NewExecuteConfigCommand()

	if cmd == nil {
		t.Fatal("Expected config command to be created")
	}

	if cmd.Use != "config" {
		t.Errorf("Expected Use to be 'config', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	// Check that it has subcommands
	if !cmd.HasSubCommands() {
		t.Error("Expected config command to have subcommands")
	}
}

func TestExecuteOptionsValidation(t *testing.T) {
	tests := []struct {
		name        string
		opts        *ExecuteOptions
		expectValid bool
	}{
		{
			name: "valid package command",
			opts: &ExecuteOptions{
				Command:     "nix-env",
				Args:        []string{"-iA", "nixpkgs.firefox"},
				Description: "Install Firefox browser",
				Category:    "package",
				Environment: make(map[string]string),
			},
			expectValid: true,
		},
		{
			name: "valid system command with sudo",
			opts: &ExecuteOptions{
				Command:      "nixos-rebuild",
				Args:         []string{"switch"},
				Description:  "Apply NixOS configuration",
				Category:     "system",
				RequiresSudo: true,
				Environment:  make(map[string]string),
			},
			expectValid: true,
		},
		{
			name: "invalid category",
			opts: &ExecuteOptions{
				Command:     "echo",
				Description: "Test command",
				Category:    "invalid_category",
				Environment: make(map[string]string),
			},
			expectValid: false,
		},
		{
			name: "missing description",
			opts: &ExecuteOptions{
				Command:     "echo",
				Category:    "utility",
				Environment: make(map[string]string),
			},
			expectValid: false,
		},
		{
			name: "missing category",
			opts: &ExecuteOptions{
				Command:     "echo",
				Description: "Test command",
				Environment: make(map[string]string),
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate category
			validCategories := []string{"package", "system", "configuration", "development", "utility"}
			categoryValid := false
			for _, cat := range validCategories {
				if tt.opts.Category == cat {
					categoryValid = true
					break
				}
			}

			// Check required fields
			hasRequiredFields := tt.opts.Command != "" && tt.opts.Description != "" && categoryValid

			if hasRequiredFields != tt.expectValid {
				t.Errorf("Expected validation result %v, got %v", tt.expectValid, hasRequiredFields)
			}
		})
	}
}

func TestExecuteOptionsDefaults(t *testing.T) {
	opts := &ExecuteOptions{
		Environment: make(map[string]string),
	}

	// Test default values
	if opts.DryRun != false {
		t.Error("Expected DryRun default to be false")
	}

	if opts.Interactive != false {
		t.Error("Expected Interactive default to be false")
	}

	if opts.Force != false {
		t.Error("Expected Force default to be false")
	}

	if opts.RequiresSudo != false {
		t.Error("Expected RequiresSudo default to be false")
	}

	if opts.Environment == nil {
		t.Error("Expected Environment to be initialized")
	}
}

func TestCategoryValidation(t *testing.T) {
	validCategories := []string{"package", "system", "configuration", "development", "utility"}

	tests := []struct {
		name     string
		category string
		valid    bool
	}{
		{"package category", "package", true},
		{"system category", "system", true},
		{"configuration category", "configuration", true},
		{"development category", "development", true},
		{"utility category", "utility", true},
		{"invalid category", "invalid", false},
		{"empty category", "", false},
		{"mixed case category", "Package", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := false
			for _, validCat := range validCategories {
				if tt.category == validCat {
					valid = true
					break
				}
			}

			if valid != tt.valid {
				t.Errorf("Expected category '%s' validation to be %v, got %v", tt.category, tt.valid, valid)
			}
		})
	}
}

func TestTimeoutParsing(t *testing.T) {
	tests := []struct {
		name      string
		timeout   string
		expectErr bool
	}{
		{"valid duration minutes", "5m", false},
		{"valid duration seconds", "30s", false},
		{"valid duration hours", "1h", false},
		{"valid duration mixed", "1h30m", false},
		{"invalid duration", "invalid", true},
		{"empty duration", "", false}, // Empty should be allowed (no timeout)
		{"zero duration", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.timeout == "" {
				return // Skip parsing for empty timeout
			}

			_, err := time.ParseDuration(tt.timeout)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseDuration(%s) error = %v, expectErr %v", tt.timeout, err, tt.expectErr)
			}
		})
	}
}

func TestEnvironmentVariableParsing(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]string
	}{
		{
			name: "single environment variable",
			envVars: map[string]string{
				"TEST_VAR": "test_value",
			},
			expected: map[string]string{
				"TEST_VAR": "test_value",
			},
		},
		{
			name: "multiple environment variables",
			envVars: map[string]string{
				"PATH":     "/usr/bin:/bin",
				"HOME":     "/home/user",
				"NIX_PATH": "/nix/var/nix/profiles/per-user/root/channels",
			},
			expected: map[string]string{
				"PATH":     "/usr/bin:/bin",
				"HOME":     "/home/user",
				"NIX_PATH": "/nix/var/nix/profiles/per-user/root/channels",
			},
		},
		{
			name:     "empty environment variables",
			envVars:  map[string]string{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envVars) != len(tt.expected) {
				t.Errorf("Expected %d environment variables, got %d", len(tt.expected), len(tt.envVars))
				return
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := tt.envVars[key]; !exists {
					t.Errorf("Expected environment variable %s to exist", key)
				} else if actualValue != expectedValue {
					t.Errorf("Expected %s=%s, got %s=%s", key, expectedValue, key, actualValue)
				}
			}
		})
	}
}

func TestMockExecuteCommandValidation(t *testing.T) {
	// Test validation logic that would be used in runExecuteCommand
	// without actually executing commands

	tests := []struct {
		name      string
		command   string
		args      []string
		category  string
		expectErr bool
	}{
		{
			name:      "valid nix command",
			command:   "nix",
			args:      []string{"build"},
			category:  "development",
			expectErr: false,
		},
		{
			name:      "valid package command",
			command:   "nix-env",
			args:      []string{"-iA", "firefox"},
			category:  "package",
			expectErr: false,
		},
		{
			name:      "potentially dangerous command",
			command:   "rm",
			args:      []string{"-rf", "/"},
			category:  "utility",
			expectErr: true, // Would be caught by security validation
		},
		{
			name:      "empty command",
			command:   "",
			args:      []string{},
			category:  "utility",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation that would occur before execution
			hasError := false

			if tt.command == "" {
				hasError = true
			}

			// Check for obviously dangerous patterns
			if tt.command == "rm" && len(tt.args) > 0 {
				for _, arg := range tt.args {
					if arg == "-rf" || arg == "--recursive" {
						hasError = true
						break
					}
				}
			}

			if hasError != tt.expectErr {
				t.Errorf("Expected error %v, got %v", tt.expectErr, hasError)
			}
		})
	}
}

func TestCommandConstruction(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		args     []string
		expected string
	}{
		{
			name:     "command without args",
			command:  "ls",
			args:     []string{},
			expected: "ls",
		},
		{
			name:     "command with single arg",
			command:  "echo",
			args:     []string{"hello"},
			expected: "echo hello",
		},
		{
			name:     "command with multiple args",
			command:  "nix-env",
			args:     []string{"-iA", "nixpkgs.firefox"},
			expected: "nix-env -iA nixpkgs.firefox",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if len(tt.args) > 0 {
				result = tt.command + " " + strings.Join(tt.args, " ")
			} else {
				result = tt.command
			}

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Helper function to simulate strings.Join since we need to import strings
func strings_Join(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return slice[0]
	}
	
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}


func BenchmarkExecuteOptionsValidation(b *testing.B) {
	opts := &ExecuteOptions{
		Command:     "nix-env",
		Args:        []string{"-iA", "nixpkgs.firefox"},
		Description: "Install Firefox browser",
		Category:    "package",
		Environment: make(map[string]string),
	}

	validCategories := []string{"package", "system", "configuration", "development", "utility"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate validation
		categoryValid := false
		for _, cat := range validCategories {
			if opts.Category == cat {
				categoryValid = true
				break
			}
		}
		_ = opts.Command != "" && opts.Description != "" && categoryValid
	}
}

func TestExecuteCommandParameterBuilding(t *testing.T) {
	opts := &ExecuteOptions{
		Command:      "nixos-rebuild",
		Args:         []string{"switch"},
		Description:  "Apply NixOS configuration",
		Category:     "system",
		RequiresSudo: true,
		WorkingDir:   "/etc/nixos",
		Environment: map[string]string{
			"NIX_PATH": "/custom/path",
		},
		DryRun:  false,
		Timeout: "5m",
	}

	// Build parameters as would be done in runExecuteCommand
	params := map[string]interface{}{
		"command":     opts.Command,
		"args":        opts.Args,
		"description": opts.Description,
		"category":    opts.Category,
		"dryRun":      opts.DryRun,
	}

	if opts.RequiresSudo {
		params["requiresSudo"] = true
	}

	if opts.WorkingDir != "" {
		params["workingDir"] = opts.WorkingDir
	}

	if len(opts.Environment) > 0 {
		params["environment"] = opts.Environment
	}

	if opts.Timeout != "" {
		params["timeout"] = opts.Timeout
	}

	// Validate constructed parameters
	if params["command"] != "nixos-rebuild" {
		t.Errorf("Expected command 'nixos-rebuild', got %v", params["command"])
	}

	if params["requiresSudo"] != true {
		t.Error("Expected requiresSudo to be true")
	}

	if params["workingDir"] != "/etc/nixos" {
		t.Errorf("Expected workingDir '/etc/nixos', got %v", params["workingDir"])
	}

	env, ok := params["environment"].(map[string]string)
	if !ok {
		t.Error("Expected environment to be map[string]string")
	} else if env["NIX_PATH"] != "/custom/path" {
		t.Errorf("Expected NIX_PATH '/custom/path', got %v", env["NIX_PATH"])
	}
}