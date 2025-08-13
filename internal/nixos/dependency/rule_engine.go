// rule_engine.go - Rule engine for NixOS configuration dependency analysis
package dependency

import (
	"fmt"
	"regexp"
	"strings"
)

// RuleEngine manages dependency analysis rules
type RuleEngine struct {
	dependencyRules []*DependencyRule
	conflictRules   []*ConflictRule
	hardwareRules   []*HardwareRule
	patterns        *RulePatterns
}

// DependencyRule defines a dependency relationship rule
type DependencyRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Pattern     string         `json:"pattern"`
	Requires    []string       `json:"requires"`
	Implies     []string       `json:"implies,omitempty"`
	Type        DependencyType `json:"type"`
	Strength    float64        `json:"strength"`
	Condition   string         `json:"condition,omitempty"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
}

// ConflictRule defines a configuration conflict rule
type ConflictRule struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Pattern     []string     `json:"pattern"` // Multiple patterns that conflict
	Type        ConflictType `json:"type"`
	Severity    string       `json:"severity"`
	Description string       `json:"description"`
	Resolution  []string     `json:"resolution"`
	AutoFix     bool         `json:"auto_fix"`
}

// HardwareRule defines hardware-specific dependency rules
type HardwareRule struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	HardwarePattern  string   `json:"hardware_pattern"`
	ConfigPattern    string   `json:"config_pattern"`
	RequiredOptions  []string `json:"required_options"`
	RequiredDrivers  []string `json:"required_drivers"`
	RequiredFirmware []string `json:"required_firmware"`
	Vendor           string   `json:"vendor,omitempty"`
	Category         string   `json:"category"`
	Priority         int      `json:"priority"`
	Description      string   `json:"description"`
}

// RulePatterns contains compiled regex patterns for matching
type RulePatterns struct {
	servicePatterns  map[string]*regexp.Regexp
	hardwarePatterns map[string]*regexp.Regexp
	modulePatterns   map[string]*regexp.Regexp
}

// NewRuleEngine creates a new rule engine with predefined rules
func NewRuleEngine() *RuleEngine {
	engine := &RuleEngine{
		dependencyRules: []*DependencyRule{},
		conflictRules:   []*ConflictRule{},
		hardwareRules:   []*HardwareRule{},
		patterns:        NewRulePatterns(),
	}
	
	// Load predefined rules
	engine.loadPredefinedRules()
	
	return engine
}

// NewRulePatterns creates new rule patterns with compiled regexes
func NewRulePatterns() *RulePatterns {
	return &RulePatterns{
		servicePatterns:  make(map[string]*regexp.Regexp),
		hardwarePatterns: make(map[string]*regexp.Regexp),
		modulePatterns:   make(map[string]*regexp.Regexp),
	}
}

// loadPredefinedRules loads the built-in dependency and conflict rules
func (re *RuleEngine) loadPredefinedRules() {
	// Graphics/Display Dependencies
	re.dependencyRules = append(re.dependencyRules, []*DependencyRule{
		{
			ID:          "nvidia-opengl",
			Name:        "NVIDIA requires OpenGL",
			Pattern:     `services\.xserver\.videoDrivers.*nvidia`,
			Requires:    []string{"hardware.opengl.enable"},
			Type:        DependencyRequired,
			Strength:    1.0,
			Description: "NVIDIA graphics drivers require OpenGL to be enabled",
			Category:    "graphics",
		},
		{
			ID:          "amd-opengl",
			Name:        "AMD GPU requires OpenGL",
			Pattern:     `services\.xserver\.videoDrivers.*(amdgpu|radeon)`,
			Requires:    []string{"hardware.opengl.enable"},
			Type:        DependencyRequired,
			Strength:    1.0,
			Description: "AMD graphics drivers require OpenGL to be enabled",
			Category:    "graphics",
		},
		{
			ID:          "xserver-display-manager",
			Name:        "X Server requires display manager",
			Pattern:     `services\.xserver\.enable\s*=\s*true`,
			Requires:    []string{"services.xserver.displayManager"},
			Type:        DependencyRecommended,
			Strength:    0.8,
			Description: "X Server typically requires a display manager",
			Category:    "desktop",
		},
	}...)

	// Audio Dependencies
	re.dependencyRules = append(re.dependencyRules, []*DependencyRule{
		{
			ID:          "pulseaudio-sound",
			Name:        "PulseAudio requires sound",
			Pattern:     `hardware\.pulseaudio\.enable\s*=\s*true`,
			Requires:    []string{"sound.enable"},
			Type:        DependencyRequired,
			Strength:    1.0,
			Description: "PulseAudio requires the sound subsystem to be enabled",
			Category:    "audio",
		},
		{
			ID:          "pipewire-sound",
			Name:        "PipeWire requires sound",
			Pattern:     `services\.pipewire\.enable\s*=\s*true`,
			Requires:    []string{"sound.enable"},
			Type:        DependencyRequired,
			Strength:    1.0,
			Description: "PipeWire requires the sound subsystem to be enabled",
			Category:    "audio",
		},
	}...)

	// Network Dependencies
	re.dependencyRules = append(re.dependencyRules, []*DependencyRule{
		{
			ID:          "wifi-firmware",
			Name:        "WiFi requires firmware",
			Pattern:     `networking\.wireless\.enable\s*=\s*true`,
			Requires:    []string{"hardware.enableRedistributableFirmware"},
			Type:        DependencyRecommended,
			Strength:    0.9,
			Description: "WiFi typically requires redistributable firmware",
			Category:    "networking",
		},
		{
			ID:          "networkmanager-wireless",
			Name:        "NetworkManager can manage wireless",
			Pattern:     `networking\.networkmanager\.enable\s*=\s*true`,
			Implies:     []string{"networking.wireless.enable"},
			Type:        DependencyOptional,
			Strength:    0.5,
			Description: "NetworkManager can handle wireless connections",
			Category:    "networking",
		},
	}...)

	// Boot Dependencies
	re.dependencyRules = append(re.dependencyRules, []*DependencyRule{
		{
			ID:          "grub-efi",
			Name:        "GRUB EFI configuration",
			Pattern:     `boot\.loader\.grub\.efiSupport\s*=\s*true`,
			Requires:    []string{"boot.loader.efi.canTouchEfiVariables"},
			Type:        DependencyRecommended,
			Strength:    0.8,
			Description: "GRUB EFI support typically requires EFI variable access",
			Category:    "boot",
		},
	}...)

	// Virtualization Dependencies
	re.dependencyRules = append(re.dependencyRules, []*DependencyRule{
		{
			ID:          "virtualbox-kernel-modules",
			Name:        "VirtualBox requires kernel modules",
			Pattern:     `virtualisation\.virtualbox\.host\.enable\s*=\s*true`,
			Requires:    []string{"boot.kernelModules"},
			Type:        DependencyRequired,
			Strength:    1.0,
			Description: "VirtualBox requires specific kernel modules",
			Category:    "virtualization",
		},
		{
			ID:          "docker-virtualization",
			Name:        "Docker requires virtualization",
			Pattern:     `virtualisation\.docker\.enable\s*=\s*true`,
			Requires:    []string{"virtualisation.enable"},
			Type:        DependencyRequired,
			Strength:    1.0,
			Description: "Docker requires virtualization to be enabled",
			Category:    "virtualization",
		},
	}...)

	// Conflict Rules
	re.conflictRules = append(re.conflictRules, []*ConflictRule{
		{
			ID:   "pulseaudio-pipewire",
			Name: "PulseAudio and PipeWire conflict",
			Pattern: []string{
				`hardware\.pulseaudio\.enable\s*=\s*true`,
				`services\.pipewire\.enable\s*=\s*true`,
			},
			Type:        ConflictMutualExclusion,
			Severity:    "critical",
			Description: "PulseAudio and PipeWire cannot be enabled simultaneously",
			Resolution:  []string{"Choose either PulseAudio or PipeWire", "Disable one of them"},
			AutoFix:     false,
		},
		{
			ID:   "networkmanager-wicd",
			Name: "NetworkManager and WICD conflict",
			Pattern: []string{
				`networking\.networkmanager\.enable\s*=\s*true`,
				`networking\.wicd\.enable\s*=\s*true`,
			},
			Type:        ConflictServiceConflict,
			Severity:    "critical",
			Description: "NetworkManager and WICD cannot run simultaneously",
			Resolution:  []string{"Choose one network manager", "Disable the other"},
			AutoFix:     false,
		},
		{
			ID:   "multiple-display-managers",
			Name: "Multiple display managers conflict",
			Pattern: []string{
				`services\.xserver\.displayManager\.gdm\.enable\s*=\s*true`,
				`services\.xserver\.displayManager\.sddm\.enable\s*=\s*true`,
			},
			Type:        ConflictServiceConflict,
			Severity:    "warning",
			Description: "Multiple display managers may cause conflicts",
			Resolution:  []string{"Use only one display manager", "Choose the preferred one"},
			AutoFix:     false,
		},
	}...)

	// Hardware Rules
	re.hardwareRules = append(re.hardwareRules, []*HardwareRule{
		{
			ID:               "nvidia-gpu",
			Name:             "NVIDIA GPU Configuration",
			HardwarePattern:  `(?i)nvidia|geforce|quadro|tesla`,
			ConfigPattern:    `services\.xserver\.videoDrivers`,
			RequiredOptions:  []string{"services.xserver.videoDrivers", "hardware.opengl.enable"},
			RequiredDrivers:  []string{"nvidia"},
			RequiredFirmware: []string{},
			Vendor:           "NVIDIA",
			Category:         "graphics",
			Priority:         9,
			Description:      "NVIDIA GPU requires proprietary drivers and OpenGL",
		},
		{
			ID:               "amd-gpu",
			Name:             "AMD GPU Configuration",
			HardwarePattern:  `(?i)amd|radeon|rx\s*\d+`,
			ConfigPattern:    `services\.xserver\.videoDrivers`,
			RequiredOptions:  []string{"services.xserver.videoDrivers", "hardware.opengl.enable"},
			RequiredDrivers:  []string{"amdgpu"},
			RequiredFirmware: []string{"linux-firmware"},
			Vendor:           "AMD",
			Category:         "graphics",
			Priority:         8,
			Description:      "AMD GPU works best with amdgpu driver and firmware",
		},
		{
			ID:               "intel-wifi",
			Name:             "Intel WiFi Configuration",
			HardwarePattern:  `(?i)intel.*wi-?fi|iwlwifi`,
			ConfigPattern:    `networking\.wireless`,
			RequiredOptions:  []string{"hardware.enableRedistributableFirmware"},
			RequiredDrivers:  []string{"iwlwifi"},
			RequiredFirmware: []string{"linux-firmware"},
			Vendor:           "Intel",
			Category:         "networking",
			Priority:         7,
			Description:      "Intel WiFi requires iwlwifi driver and firmware",
		},
		{
			ID:               "broadcom-wifi",
			Name:             "Broadcom WiFi Configuration",
			HardwarePattern:  `(?i)broadcom|bcm\d+`,
			ConfigPattern:    `networking\.wireless`,
			RequiredOptions:  []string{"hardware.enableRedistributableFirmware"},
			RequiredDrivers:  []string{"brcmfmac", "brcmsmac"},
			RequiredFirmware: []string{"linux-firmware"},
			Vendor:           "Broadcom",
			Category:         "networking",
			Priority:         6,
			Description:      "Broadcom WiFi requires specific drivers and firmware",
		},
	}...)
}

// MatchDependencyRules finds dependency rules that match the given configuration
func (re *RuleEngine) MatchDependencyRules(configOptions []*ConfigOption) []*DependencyRule {
	var matchedRules []*DependencyRule

	for _, rule := range re.dependencyRules {
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}

		for _, option := range configOptions {
			configLine := fmt.Sprintf("%s = %v", option.Name, option.Value)
			if pattern.MatchString(configLine) {
				matchedRules = append(matchedRules, rule)
				break
			}
		}
	}

	return matchedRules
}

// MatchConflictRules finds conflict rules that apply to the given configuration
func (re *RuleEngine) MatchConflictRules(configOptions []*ConfigOption) []*ConflictRule {
	var matchedRules []*ConflictRule

	for _, rule := range re.conflictRules {
		matchCount := 0
		
		for _, pattern := range rule.Pattern {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			for _, option := range configOptions {
				configLine := fmt.Sprintf("%s = %v", option.Name, option.Value)
				if regex.MatchString(configLine) {
					matchCount++
					break
				}
			}
		}

		// If all patterns match, we have a conflict
		if matchCount == len(rule.Pattern) {
			matchedRules = append(matchedRules, rule)
		}
	}

	return matchedRules
}

// MatchHardwareRules finds hardware rules that match detected hardware
func (re *RuleEngine) MatchHardwareRules(hardwareInfo map[string]string) []*HardwareRule {
	var matchedRules []*HardwareRule

	for _, rule := range re.hardwareRules {
		pattern, err := regexp.Compile(rule.HardwarePattern)
		if err != nil {
			continue
		}

		for _, hwInfo := range hardwareInfo {
			if pattern.MatchString(hwInfo) {
				matchedRules = append(matchedRules, rule)
				break
			}
		}
	}

	return matchedRules
}

// ValidateConfiguration validates configuration against rules
func (re *RuleEngine) ValidateConfiguration(configOptions []*ConfigOption) *ValidationResults {
	results := &ValidationResults{
		Valid:       true,
		Errors:      []*ValidationError{},
		Warnings:    []*ValidationError{},
		Suggestions: []string{},
		Score:       1.0,
	}

	// Check dependency rules
	dependencyRules := re.MatchDependencyRules(configOptions)
	configMap := make(map[string]*ConfigOption)
	for _, opt := range configOptions {
		configMap[opt.Name] = opt
	}

	for _, rule := range dependencyRules {
		if rule.Type == DependencyRequired {
			for _, required := range rule.Requires {
				if _, exists := configMap[required]; !exists {
					results.Valid = false
					results.Errors = append(results.Errors, &ValidationError{
						Type:       "missing_dependency",
						Option:     required,
						Message:    fmt.Sprintf("Required option '%s' is missing (required by %s)", required, rule.Name),
						Severity:   "error",
						Resolution: fmt.Sprintf("Add: %s = <appropriate_value>", required),
					})
				}
			}
		}
	}

	// Check conflict rules
	conflictRules := re.MatchConflictRules(configOptions)
	for _, rule := range conflictRules {
		if rule.Severity == "critical" {
			results.Valid = false
			results.Errors = append(results.Errors, &ValidationError{
				Type:       "configuration_conflict",
				Message:    rule.Description,
				Severity:   "error",
				Resolution: strings.Join(rule.Resolution, " or "),
			})
		} else {
			results.Warnings = append(results.Warnings, &ValidationError{
				Type:       "configuration_conflict",
				Message:    rule.Description,
				Severity:   "warning",
				Resolution: strings.Join(rule.Resolution, " or "),
			})
		}
	}

	// Calculate score based on errors and warnings
	errorPenalty := float64(len(results.Errors)) * 0.2
	warningPenalty := float64(len(results.Warnings)) * 0.1
	results.Score = 1.0 - errorPenalty - warningPenalty
	if results.Score < 0 {
		results.Score = 0
	}

	return results
}

// GetRulesByCategory returns rules filtered by category
func (re *RuleEngine) GetRulesByCategory(category string) []*DependencyRule {
	var rules []*DependencyRule
	for _, rule := range re.dependencyRules {
		if rule.Category == category {
			rules = append(rules, rule)
		}
	}
	return rules
}

// GetRuleCategories returns all available rule categories
func (re *RuleEngine) GetRuleCategories() []string {
	categories := make(map[string]bool)
	for _, rule := range re.dependencyRules {
		categories[rule.Category] = true
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	return result
}