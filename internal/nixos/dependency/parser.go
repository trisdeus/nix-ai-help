// parser.go - Configuration parsing and analysis methods
package dependency

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"nix-ai-help/internal/hardware"
)

// parseConfigurationOptions parses NixOS configuration content and extracts options
func (da *DependencyAnalyzer) parseConfigurationOptions(configContent string) ([]*ConfigOption, error) {
	var options []*ConfigOption

	// Remove comments and normalize whitespace
	content := da.normalizeConfig(configContent)

	// Parse different types of configuration patterns
	options = append(options, da.parseSimpleOptions(content)...)
	options = append(options, da.parseServiceOptions(content)...)
	options = append(options, da.parseHardwareOptions(content)...)
	options = append(options, da.parseBootOptions(content)...)
	options = append(options, da.parseNetworkOptions(content)...)

	return options, nil
}

// normalizeConfig removes comments and normalizes whitespace
func (da *DependencyAnalyzer) normalizeConfig(content string) string {
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Remove comments (simple implementation)
		if commentIndex := strings.Index(line, "#"); commentIndex != -1 {
			line = line[:commentIndex]
		}
		
		line = strings.TrimSpace(line)
		if line != "" && line != "{" && line != "}" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// parseSimpleOptions parses basic key-value configuration options
func (da *DependencyAnalyzer) parseSimpleOptions(content string) []*ConfigOption {
	var options []*ConfigOption

	// Pattern for simple assignments: option.name = value;
	pattern := regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9._]*)\s*=\s*([^;]+);`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			name := strings.TrimSpace(match[1])
			valueStr := strings.TrimSpace(match[2])
			
			option := &ConfigOption{
				Name:        name,
				Value:       da.parseValue(valueStr),
				Type:        da.inferType(valueStr),
				Category:    da.categorizeOption(name),
				Module:      da.getModuleFromName(name),
				Description: da.getOptionDescription(name),
				Attributes:  make(map[string]interface{}),
			}

			options = append(options, option)
		}
	}

	return options
}

// parseServiceOptions parses service-related configuration options
func (da *DependencyAnalyzer) parseServiceOptions(content string) []*ConfigOption {
	var options []*ConfigOption

	// Pattern for services: services.servicename.enable = true;
	servicePattern := regexp.MustCompile(`services\.([a-zA-Z][a-zA-Z0-9_]*(?:\.[a-zA-Z][a-zA-Z0-9_]*)*)\s*=\s*([^;]+);`)
	matches := servicePattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			servicePath := match[1]
			valueStr := strings.TrimSpace(match[2])
			fullName := "services." + servicePath

			option := &ConfigOption{
				Name:        fullName,
				Value:       da.parseValue(valueStr),
				Type:        da.inferType(valueStr),
				Category:    "services",
				Module:      da.getServiceModule(servicePath),
				Description: da.getServiceDescription(servicePath),
				Attributes: map[string]interface{}{
					"service_name": strings.Split(servicePath, ".")[0],
					"is_service":   true,
				},
			}

			options = append(options, option)
		}
	}

	return options
}

// parseHardwareOptions parses hardware-related configuration options
func (da *DependencyAnalyzer) parseHardwareOptions(content string) []*ConfigOption {
	var options []*ConfigOption

	// Pattern for hardware: hardware.component.option = value;
	hardwarePattern := regexp.MustCompile(`hardware\.([a-zA-Z][a-zA-Z0-9._]*)\s*=\s*([^;]+);`)
	matches := hardwarePattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			hardwarePath := match[1]
			valueStr := strings.TrimSpace(match[2])
			fullName := "hardware." + hardwarePath

			option := &ConfigOption{
				Name:        fullName,
				Value:       da.parseValue(valueStr),
				Type:        da.inferType(valueStr),
				Category:    "hardware",
				Module:      "hardware",
				Description: da.getHardwareDescription(hardwarePath),
				Attributes: map[string]interface{}{
					"hardware_component": strings.Split(hardwarePath, ".")[0],
					"is_hardware":        true,
				},
			}

			options = append(options, option)
		}
	}

	return options
}

// parseBootOptions parses boot-related configuration options
func (da *DependencyAnalyzer) parseBootOptions(content string) []*ConfigOption {
	var options []*ConfigOption

	bootPattern := regexp.MustCompile(`boot\.([a-zA-Z][a-zA-Z0-9._]*)\s*=\s*([^;]+);`)
	matches := bootPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			bootPath := match[1]
			valueStr := strings.TrimSpace(match[2])
			fullName := "boot." + bootPath

			option := &ConfigOption{
				Name:        fullName,
				Value:       da.parseValue(valueStr),
				Type:        da.inferType(valueStr),
				Category:    "boot",
				Module:      "boot",
				Description: da.getBootDescription(bootPath),
				Attributes: map[string]interface{}{
					"boot_component": strings.Split(bootPath, ".")[0],
					"is_boot":        true,
				},
			}

			options = append(options, option)
		}
	}

	return options
}

// parseNetworkOptions parses networking-related configuration options
func (da *DependencyAnalyzer) parseNetworkOptions(content string) []*ConfigOption {
	var options []*ConfigOption

	networkPattern := regexp.MustCompile(`networking\.([a-zA-Z][a-zA-Z0-9._]*)\s*=\s*([^;]+);`)
	matches := networkPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			networkPath := match[1]
			valueStr := strings.TrimSpace(match[2])
			fullName := "networking." + networkPath

			option := &ConfigOption{
				Name:        fullName,
				Value:       da.parseValue(valueStr),
				Type:        da.inferType(valueStr),
				Category:    "networking",
				Module:      "networking",
				Description: da.getNetworkDescription(networkPath),
				Attributes: map[string]interface{}{
					"network_component": strings.Split(networkPath, ".")[0],
					"is_networking":     true,
				},
			}

			options = append(options, option)
		}
	}

	return options
}

// parseValue parses a configuration value and returns appropriate type
func (da *DependencyAnalyzer) parseValue(valueStr string) interface{} {
	valueStr = strings.TrimSpace(valueStr)

	// Boolean values
	if valueStr == "true" {
		return true
	}
	if valueStr == "false" {
		return false
	}

	// Numeric values
	if intVal, err := strconv.Atoi(valueStr); err == nil {
		return intVal
	}
	if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return floatVal
	}

	// String values (remove quotes)
	if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") {
		return valueStr[1 : len(valueStr)-1]
	}
	if strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'") {
		return valueStr[1 : len(valueStr)-1]
	}

	// Array/List values
	if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
		return da.parseArray(valueStr)
	}

	// Default to string
	return valueStr
}

// parseArray parses array-like values
func (da *DependencyAnalyzer) parseArray(arrayStr string) []interface{} {
	// Remove brackets
	content := arrayStr[1 : len(arrayStr)-1]
	content = strings.TrimSpace(content)

	if content == "" {
		return []interface{}{}
	}

	// Simple parsing - split by whitespace or commas
	var items []interface{}
	parts := regexp.MustCompile(`[,\s]+`).Split(content, -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			// Remove quotes if present
			if strings.HasPrefix(part, "\"") && strings.HasSuffix(part, "\"") {
				part = part[1 : len(part)-1]
			}
			items = append(items, part)
		}
	}

	return items
}

// inferType infers the configuration option type
func (da *DependencyAnalyzer) inferType(valueStr string) string {
	valueStr = strings.TrimSpace(valueStr)

	if valueStr == "true" || valueStr == "false" {
		return "boolean"
	}
	if _, err := strconv.Atoi(valueStr); err == nil {
		return "integer"
	}
	if _, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return "float"
	}
	if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
		return "array"
	}
	if strings.HasPrefix(valueStr, "{") && strings.HasSuffix(valueStr, "}") {
		return "object"
	}

	return "string"
}

// categorizeOption categorizes configuration options
func (da *DependencyAnalyzer) categorizeOption(name string) string {
	if strings.HasPrefix(name, "services.") {
		return "services"
	}
	if strings.HasPrefix(name, "hardware.") {
		return "hardware"
	}
	if strings.HasPrefix(name, "boot.") {
		return "boot"
	}
	if strings.HasPrefix(name, "networking.") {
		return "networking"
	}
	if strings.HasPrefix(name, "system.") {
		return "system"
	}
	if strings.HasPrefix(name, "environment.") {
		return "environment"
	}
	if strings.HasPrefix(name, "users.") {
		return "users"
	}
	if strings.HasPrefix(name, "security.") {
		return "security"
	}

	return "general"
}

// getModuleFromName determines the NixOS module for a configuration option
func (da *DependencyAnalyzer) getModuleFromName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	if len(parts) >= 1 {
		return parts[0]
	}
	return "unknown"
}

// getOptionDescription returns a description for common configuration options
func (da *DependencyAnalyzer) getOptionDescription(name string) string {
	descriptions := map[string]string{
		"sound.enable":                     "Enable sound subsystem",
		"hardware.pulseaudio.enable":      "Enable PulseAudio sound server",
		"services.pipewire.enable":        "Enable PipeWire multimedia framework",
		"hardware.opengl.enable":          "Enable OpenGL hardware acceleration",
		"networking.networkmanager.enable": "Enable NetworkManager for network configuration",
		"services.xserver.enable":         "Enable X Window System server",
		"boot.loader.grub.enable":         "Enable GRUB bootloader",
		"nixpkgs.config.allowUnfree":      "Allow installation of non-free packages",
	}

	if desc, exists := descriptions[name]; exists {
		return desc
	}

	// Generate basic description based on name
	parts := strings.Split(name, ".")
	if len(parts) >= 2 {
		return fmt.Sprintf("Configure %s for %s", parts[len(parts)-1], parts[0])
	}

	return fmt.Sprintf("Configure %s", name)
}

// Service-specific helper methods
func (da *DependencyAnalyzer) getServiceModule(servicePath string) string {
	parts := strings.Split(servicePath, ".")
	if len(parts) >= 1 {
		return "services." + parts[0]
	}
	return "services"
}

func (da *DependencyAnalyzer) getServiceDescription(servicePath string) string {
	parts := strings.Split(servicePath, ".")
	serviceName := parts[0]

	descriptions := map[string]string{
		"xserver":        "X Window System server",
		"pulseaudio":     "PulseAudio sound server",
		"pipewire":       "PipeWire multimedia framework",
		"networkmanager": "Network configuration manager",
		"openssh":        "OpenSSH secure shell daemon",
		"nginx":          "Nginx web server",
		"postgresql":     "PostgreSQL database server",
		"docker":         "Docker container platform",
	}

	if desc, exists := descriptions[serviceName]; exists {
		if len(parts) > 1 {
			return fmt.Sprintf("%s - %s configuration", desc, parts[len(parts)-1])
		}
		return desc
	}

	return fmt.Sprintf("Service: %s", serviceName)
}

// Hardware-specific helper methods
func (da *DependencyAnalyzer) getHardwareDescription(hardwarePath string) string {
	parts := strings.Split(hardwarePath, ".")
	component := parts[0]

	descriptions := map[string]string{
		"pulseaudio": "PulseAudio hardware integration",
		"opengl":     "OpenGL hardware acceleration",
		"nvidia":     "NVIDIA graphics hardware support",
		"amd":        "AMD hardware support",
		"bluetooth":  "Bluetooth hardware support",
		"cpu":        "CPU-specific hardware configuration",
	}

	if desc, exists := descriptions[component]; exists {
		return desc
	}

	return fmt.Sprintf("Hardware: %s", component)
}

// Boot-specific helper methods
func (da *DependencyAnalyzer) getBootDescription(bootPath string) string {
	parts := strings.Split(bootPath, ".")
	component := parts[0]

	descriptions := map[string]string{
		"loader":         "Boot loader configuration",
		"kernelModules":  "Kernel modules to load at boot",
		"kernelParams":   "Kernel command line parameters",
		"initrd":         "Initial RAM disk configuration",
		"kernel":         "Kernel selection and configuration",
	}

	if desc, exists := descriptions[component]; exists {
		return desc
	}

	return fmt.Sprintf("Boot: %s", component)
}

// Network-specific helper methods
func (da *DependencyAnalyzer) getNetworkDescription(networkPath string) string {
	parts := strings.Split(networkPath, ".")
	component := parts[0]

	descriptions := map[string]string{
		"networkmanager": "NetworkManager configuration",
		"wireless":       "Wireless networking configuration",
		"hostName":       "System hostname configuration",
		"firewall":       "Firewall configuration",
		"interfaces":     "Network interfaces configuration",
	}

	if desc, exists := descriptions[component]; exists {
		return desc
	}

	return fmt.Sprintf("Network: %s", component)
}

// analyzeDependencies analyzes dependencies between configuration options
func (da *DependencyAnalyzer) analyzeDependencies(configOptions []*ConfigOption) []*Dependency {
	var dependencies []*Dependency

	// Use rule engine to find dependencies
	matchedRules := da.ruleEngine.MatchDependencyRules(configOptions)

	for _, rule := range matchedRules {
		// Find the source option that matched this rule
		var sourceOption string
		for _, option := range configOptions {
			configLine := fmt.Sprintf("%s = %v", option.Name, option.Value)
			if matched, _ := regexp.MatchString(rule.Pattern, configLine); matched {
				sourceOption = option.Name
				break
			}
		}

		// Create dependencies for required options
		for _, required := range rule.Requires {
			dependency := &Dependency{
				From:        sourceOption,
				To:          required,
				Type:        rule.Type,
				Strength:    rule.Strength,
				Description: rule.Description,
				AutoResolve: rule.Type == DependencyRequired,
			}

			if rule.Type == DependencyRequired {
				dependency.Resolution = fmt.Sprintf("Add: %s = <appropriate_value>", required)
			}

			dependencies = append(dependencies, dependency)
		}

		// Create implications for implied options
		for _, implied := range rule.Implies {
			dependency := &Dependency{
				From:        sourceOption,
				To:          implied,
				Type:        DependencyImplies,
				Strength:    rule.Strength * 0.7, // Implications are weaker
				Description: fmt.Sprintf("%s implies %s", rule.Name, implied),
				AutoResolve: false,
			}

			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies
}

// detectConflicts detects configuration conflicts
func (da *DependencyAnalyzer) detectConflicts(configOptions []*ConfigOption, dependencies []*Dependency) []*Conflict {
	var conflicts []*Conflict

	// Use rule engine to find conflicts
	matchedRules := da.ruleEngine.MatchConflictRules(configOptions)

	for _, rule := range matchedRules {
		var conflictingOptions []string

		// Find all options that match the conflict patterns
		for _, pattern := range rule.Pattern {
			regex, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			for _, option := range configOptions {
				configLine := fmt.Sprintf("%s = %v", option.Name, option.Value)
				if regex.MatchString(configLine) {
					conflictingOptions = append(conflictingOptions, option.Name)
					break
				}
			}
		}

		conflict := &Conflict{
			Options:     conflictingOptions,
			Type:        rule.Type,
			Severity:    rule.Severity,
			Description: rule.Description,
			Resolution:  rule.Resolution,
			AutoFix:     rule.AutoFix,
		}

		if rule.AutoFix {
			conflict.FixActions = []string{
				fmt.Sprintf("Disable conflicting option: %s", conflictingOptions[1]),
			}
		}

		conflicts = append(conflicts, conflict)
	}

	return conflicts
}

// analyzeHardwareDependencies analyzes hardware-specific dependencies
func (da *DependencyAnalyzer) analyzeHardwareDependencies(configOptions []*ConfigOption, hardwareInfo *hardware.EnhancedHardwareInfo) []*HardwareDependency {
	var hardwareDeps []*HardwareDependency

	if hardwareInfo == nil {
		return hardwareDeps
	}

	// Create hardware information map for rule matching
	hwInfoMap := make(map[string]string)
	
	if hardwareInfo.SystemProfile != nil {
		if hardwareInfo.SystemProfile.CPUDetails != nil {
			hwInfoMap["cpu"] = fmt.Sprintf("%s %s", hardwareInfo.SystemProfile.CPUDetails.Vendor, hardwareInfo.SystemProfile.CPUDetails.Model)
		}
		
		for i, gpu := range hardwareInfo.SystemProfile.GPUDetails {
			hwInfoMap[fmt.Sprintf("gpu%d", i)] = fmt.Sprintf("%s %s", gpu.Vendor, gpu.Model)
		}
		
		for i, network := range hardwareInfo.SystemProfile.NetworkDetails {
			hwInfoMap[fmt.Sprintf("network%d", i)] = fmt.Sprintf("%s %s", network.Type, network.InterfaceName)
		}
	}

	// Match hardware rules
	matchedRules := da.ruleEngine.MatchHardwareRules(hwInfoMap)

	for _, rule := range matchedRules {
		// Check if configuration has the required options
		configMap := make(map[string]*ConfigOption)
		for _, opt := range configOptions {
			configMap[opt.Name] = opt
		}

		detected := true
		compatible := true
		var missingOptions []string

		for _, required := range rule.RequiredOptions {
			if _, exists := configMap[required]; !exists {
				missingOptions = append(missingOptions, required)
				compatible = false
			}
		}

		notes := ""
		if len(missingOptions) > 0 {
			notes = fmt.Sprintf("Missing required options: %s", strings.Join(missingOptions, ", "))
		}

		hardwareDep := &HardwareDependency{
			Option:           rule.ConfigPattern,
			HardwareType:     rule.Category,
			HardwareVendor:   rule.Vendor,
			RequiredDrivers:  rule.RequiredDrivers,
			RequiredFirmware: rule.RequiredFirmware,
			RequiredModules:  []string{}, // Could be extracted from rule
			Detected:         detected,
			Compatible:       compatible,
			Notes:            notes,
		}

		// Generate config snippet for missing configuration
		if len(missingOptions) > 0 {
			var configLines []string
			for _, opt := range rule.RequiredOptions {
				if _, exists := configMap[opt]; !exists {
					configLines = append(configLines, fmt.Sprintf("  %s = true;", opt))
				}
			}
			hardwareDep.ConfigSnippet = "{\n" + strings.Join(configLines, "\n") + "\n}"
		}

		hardwareDeps = append(hardwareDeps, hardwareDep)
	}

	return hardwareDeps
}