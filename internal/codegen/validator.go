package codegen

import (
	"fmt"
	"regexp"
	"strings"

	"nix-ai-help/pkg/logger"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	Severity   string `json:"severity"` // "error", "warning", "info"
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid       bool              `json:"valid"`
	Errors      []ValidationError `json:"errors"`
	Warnings    []ValidationError `json:"warnings"`
	Suggestions []string          `json:"suggestions"`
	Score       int               `json:"score"` // 0-100, quality score
}

// Validator handles NixOS configuration validation
type Validator struct {
	logger logger.Logger
	rules  []ValidationRule
}

// ValidationRule represents a single validation rule
type ValidationRule struct {
	Name        string
	Description string
	Check       func(config string) []ValidationError
	Severity    string
}

// NewValidator creates a new configuration validator
func NewValidator(logger logger.Logger) *Validator {
	v := &Validator{
		logger: logger,
	}
	v.initializeRules()
	return v
}

// ValidateBasic performs basic syntax validation
func (v *Validator) ValidateBasic(config string) error {
	result := v.Validate(config)
	if !result.Valid {
		var errors []string
		for _, err := range result.Errors {
			errors = append(errors, err.Message)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}
	return nil
}

// Validate performs comprehensive configuration validation
func (v *Validator) Validate(config string) *ValidationResult {
	v.logger.Info("Validating NixOS configuration")

	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationError{},
		Suggestions: []string{},
		Score:       100,
	}

	// Run all validation rules
	for _, rule := range v.rules {
		errors := rule.Check(config)
		for _, err := range errors {
			if err.Severity == "error" {
				result.Errors = append(result.Errors, err)
				result.Valid = false
				result.Score -= 20
			} else if err.Severity == "warning" {
				result.Warnings = append(result.Warnings, err)
				result.Score -= 5
			}
		}
	}

	// Ensure score doesn't go below 0
	if result.Score < 0 {
		result.Score = 0
	}

	// Add general suggestions based on configuration
	result.Suggestions = v.generateSuggestions(config)

	v.logger.Info(fmt.Sprintf("Validation completed - valid: %v, errors: %d, warnings: %d", result.Valid, len(result.Errors), len(result.Warnings)))

	return result
}

// initializeRules sets up validation rules
func (v *Validator) initializeRules() {
	v.rules = []ValidationRule{
		{
			Name:        "syntax_check",
			Description: "Basic Nix syntax validation",
			Check:       v.checkBasicSyntax,
			Severity:    "error",
		},
		{
			Name:        "module_structure",
			Description: "Validate module structure",
			Check:       v.checkModuleStructure,
			Severity:    "error",
		},
		{
			Name:        "security_check",
			Description: "Security configuration validation",
			Check:       v.checkSecuritySettings,
			Severity:    "warning",
		},
		{
			Name:        "performance_check",
			Description: "Performance optimization validation",
			Check:       v.checkPerformanceSettings,
			Severity:    "info",
		},
		{
			Name:        "compatibility_check",
			Description: "Compatibility and deprecation check",
			Check:       v.checkCompatibility,
			Severity:    "warning",
		},
		{
			Name:        "best_practices",
			Description: "Best practices validation",
			Check:       v.checkBestPractices,
			Severity:    "info",
		},
	}
}

// checkBasicSyntax validates basic Nix syntax
func (v *Validator) checkBasicSyntax(config string) []ValidationError {
	var errors []ValidationError

	// Check for balanced braces
	braceCount := 0
	lines := strings.Split(config, "\n")
	for i, line := range lines {
		for _, char := range line {
			switch char {
			case '{':
				braceCount++
			case '}':
				braceCount--
				if braceCount < 0 {
					errors = append(errors, ValidationError{
						Type:       "syntax",
						Message:    "Unmatched closing brace",
						Line:       i + 1,
						Severity:   "error",
						Suggestion: "Check for missing opening brace",
					})
				}
			}
		}
	}

	if braceCount > 0 {
		errors = append(errors, ValidationError{
			Type:       "syntax",
			Message:    "Unmatched opening brace(s)",
			Severity:   "error",
			Suggestion: "Add missing closing brace(s)",
		})
	}

	// Check for proper string quotes
	quoteRegex := regexp.MustCompile(`"[^"]*"`)
	singleQuoteRegex := regexp.MustCompile(`'[^']*'`)

	if !quoteRegex.MatchString(config) && strings.Contains(config, `"`) {
		errors = append(errors, ValidationError{
			Type:       "syntax",
			Message:    "Unclosed string literal",
			Severity:   "error",
			Suggestion: "Ensure all string literals are properly quoted",
		})
	}

	if !singleQuoteRegex.MatchString(config) && strings.Contains(config, `'`) {
		errors = append(errors, ValidationError{
			Type:       "syntax",
			Message:    "Unclosed single-quoted string literal",
			Severity:   "error",
			Suggestion: "Ensure all single-quoted string literals are properly quoted",
		})
	}

	// Check for semicolon endings where required
	lines = strings.Split(config, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "=") && !strings.HasSuffix(trimmed, ";") &&
			!strings.HasSuffix(trimmed, "{") && !strings.HasSuffix(trimmed, "}") &&
			!strings.HasPrefix(trimmed, "#") && trimmed != "" {
			errors = append(errors, ValidationError{
				Type:       "syntax",
				Message:    "Missing semicolon",
				Line:       i + 1,
				Severity:   "error",
				Suggestion: "Add semicolon at the end of the assignment",
			})
		}
	}

	return errors
}

// checkModuleStructure validates the overall module structure
func (v *Validator) checkModuleStructure(config string) []ValidationError {
	var errors []ValidationError

	// Check for proper module header
	if !strings.Contains(config, "{ config, pkgs") && !strings.Contains(config, "{ config, lib, pkgs") {
		errors = append(errors, ValidationError{
			Type:       "structure",
			Message:    "Missing proper module header",
			Severity:   "error",
			Suggestion: "Start with '{ config, pkgs, ... }:' or '{ config, lib, pkgs, ... }:'",
		})
	}

	// Check for proper module closing
	if !strings.Contains(config, "}") {
		errors = append(errors, ValidationError{
			Type:       "structure",
			Message:    "Missing module closing brace",
			Severity:   "error",
			Suggestion: "End configuration with closing brace '}'",
		})
	}

	// Check for imports section if imports are used
	if strings.Contains(config, "imports") {
		importsRegex := regexp.MustCompile(`imports\s*=\s*\[`)
		if !importsRegex.MatchString(config) {
			errors = append(errors, ValidationError{
				Type:       "structure",
				Message:    "Malformed imports section",
				Severity:   "error",
				Suggestion: "Use proper imports format: imports = [ ./module.nix ];",
			})
		}
	}

	return errors
}

// checkSecuritySettings validates security configurations
func (v *Validator) checkSecuritySettings(config string) []ValidationError {
	var errors []ValidationError

	// Check for firewall configuration
	if strings.Contains(config, "services.nginx") || strings.Contains(config, "services.apache") {
		if !strings.Contains(config, "networking.firewall") {
			errors = append(errors, ValidationError{
				Type:       "security",
				Message:    "Web server configured without firewall settings",
				Severity:   "warning",
				Suggestion: "Consider configuring networking.firewall.allowedTCPPorts",
			})
		}
	}

	// Check for SSH hardening
	if strings.Contains(config, "services.openssh.enable = true") {
		if !strings.Contains(config, "PasswordAuthentication") && !strings.Contains(config, "permitRootLogin") {
			errors = append(errors, ValidationError{
				Type:       "security",
				Message:    "SSH enabled without security hardening",
				Severity:   "warning",
				Suggestion: "Consider disabling password authentication and root login",
			})
		}
	}

	// Check for automatic updates
	if !strings.Contains(config, "system.autoUpgrade") && !strings.Contains(config, "nixpkgs.config.allowUnfree") {
		errors = append(errors, ValidationError{
			Type:       "security",
			Message:    "No automatic update configuration",
			Severity:   "info",
			Suggestion: "Consider enabling system.autoUpgrade for security updates",
		})
	}

	return errors
}

// checkPerformanceSettings validates performance configurations
func (v *Validator) checkPerformanceSettings(config string) []ValidationError {
	var errors []ValidationError

	// Check for zram configuration on systems with limited RAM
	if !strings.Contains(config, "zramSwap") && !strings.Contains(config, "swapDevices") {
		errors = append(errors, ValidationError{
			Type:       "performance",
			Message:    "No swap configuration detected",
			Severity:   "info",
			Suggestion: "Consider configuring zramSwap for better memory management",
		})
	}

	// Check for SSD optimizations
	if strings.Contains(config, "fileSystems") && !strings.Contains(config, "noatime") {
		errors = append(errors, ValidationError{
			Type:       "performance",
			Message:    "Filesystem without noatime option",
			Severity:   "info",
			Suggestion: "Consider adding 'noatime' option for SSD optimization",
		})
	}

	return errors
}

// checkCompatibility validates compatibility and checks for deprecations
func (v *Validator) checkCompatibility(config string) []ValidationError {
	var errors []ValidationError

	// Check for deprecated options
	deprecatedOptions := map[string]string{
		"services.xserver.enable":    "Consider using services.xserver.displayManager instead",
		"services.printing.enable":   "services.printing is being replaced in newer versions",
		"networking.wireless.enable": "Consider using networking.networkmanager for better support",
	}

	for deprecated, suggestion := range deprecatedOptions {
		if strings.Contains(config, deprecated) {
			errors = append(errors, ValidationError{
				Type:       "compatibility",
				Message:    fmt.Sprintf("Deprecated option: %s", deprecated),
				Severity:   "warning",
				Suggestion: suggestion,
			})
		}
	}

	// Check for Flakes syntax in non-flake context
	if strings.Contains(config, "inputs.") && !strings.Contains(config, "flake") {
		errors = append(errors, ValidationError{
			Type:       "compatibility",
			Message:    "Flakes syntax used in non-flake configuration",
			Severity:   "warning",
			Suggestion: "Ensure you're using a flake-based configuration",
		})
	}

	return errors
}

// checkBestPractices validates against best practices
func (v *Validator) checkBestPractices(config string) []ValidationError {
	var errors []ValidationError

	// Check for documentation/comments
	lines := strings.Split(config, "\n")
	commentLines := 0
	totalLines := len(lines)

	for _, line := range lines {
		if strings.TrimSpace(line) != "" && strings.HasPrefix(strings.TrimSpace(line), "#") {
			commentLines++
		}
	}

	if totalLines > 20 && float64(commentLines)/float64(totalLines) < 0.1 {
		errors = append(errors, ValidationError{
			Type:       "best_practices",
			Message:    "Configuration lacks documentation comments",
			Severity:   "info",
			Suggestion: "Add comments to explain complex configurations",
		})
	}

	// Check for modular structure
	if len(lines) > 50 && !strings.Contains(config, "imports") {
		errors = append(errors, ValidationError{
			Type:       "best_practices",
			Message:    "Large configuration without modular structure",
			Severity:   "info",
			Suggestion: "Consider splitting into multiple modules using imports",
		})
	}

	// Check for version pinning
	if !strings.Contains(config, "system.stateVersion") {
		errors = append(errors, ValidationError{
			Type:       "best_practices",
			Message:    "Missing system.stateVersion",
			Severity:   "warning",
			Suggestion: "Add system.stateVersion for upgrade compatibility",
		})
	}

	return errors
}

// generateSuggestions generates general suggestions based on configuration content
func (v *Validator) generateSuggestions(config string) []string {
	var suggestions []string

	// Suggest enabling unfree packages if not set
	if !strings.Contains(config, "allowUnfree") && (strings.Contains(config, "vscode") || strings.Contains(config, "chrome")) {
		suggestions = append(suggestions, "Consider enabling nixpkgs.config.allowUnfree for proprietary packages")
	}

	// Suggest garbage collection
	if !strings.Contains(config, "gc") && !strings.Contains(config, "autoOptimiseStore") {
		suggestions = append(suggestions, "Consider enabling nix.gc.automatic for disk space management")
	}

	// Suggest documentation
	if !strings.Contains(config, "documentation") {
		suggestions = append(suggestions, "Consider enabling documentation.nixos.enable for offline docs")
	}

	return suggestions
}

// ValidateNixSyntax validates Nix syntax using a more sophisticated approach
func (v *Validator) ValidateNixSyntax(config string) []ValidationError {
	// This is a simplified approach - for production, you might want to use
	// the actual Nix parser or a dedicated Nix syntax checker
	return v.checkBasicSyntax(config)
}

// ValidateAgainstSchema validates configuration against NixOS options schema
func (v *Validator) ValidateAgainstSchema(config string) []ValidationError {
	var errors []ValidationError

	// In a full implementation, this would validate against the actual NixOS options
	// For now, we'll check for common option patterns

	knownOptions := []string{
		"boot.loader",
		"networking.hostName",
		"time.timeZone",
		"services.",
		"environment.systemPackages",
		"users.users",
		"hardware.",
		"systemd.",
		"security.",
	}

	lines := strings.Split(config, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) > 0 {
				option := strings.TrimSpace(parts[0])
				found := false
				for _, knownOption := range knownOptions {
					if strings.HasPrefix(option, knownOption) {
						found = true
						break
					}
				}
				if !found && !strings.Contains(option, "imports") && !strings.Contains(option, "system.stateVersion") {
					errors = append(errors, ValidationError{
						Type:       "schema",
						Message:    fmt.Sprintf("Unknown or potentially invalid option: %s", option),
						Line:       i + 1,
						Severity:   "warning",
						Suggestion: "Verify this option exists in NixOS documentation",
					})
				}
			}
		}
	}

	return errors
}
