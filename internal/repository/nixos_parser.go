package repository

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"nix-ai-help/internal/fleet"
	"nix-ai-help/pkg/logger"
)

// NixOSRepository represents a NixOS configuration repository
type NixOSRepository struct {
	Path    string
	logger  *logger.Logger
	configs map[string]*NixOSConfig
}

// NixOSConfig represents a parsed NixOS configuration
type NixOSConfig struct {
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	Type          string            `json:"type"` // system, home-manager, flake, module
	Hostname      string            `json:"hostname,omitempty"`
	Services      []string          `json:"services"`
	Packages      []string          `json:"packages"`
	Options       map[string]string `json:"options"`
	Dependencies  []string          `json:"dependencies"`
	NetworkConfig *NetworkConfig    `json:"network_config,omitempty"`
	Hardware      *HardwareConfig   `json:"hardware,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	StaticIP   string   `json:"static_ip,omitempty"`
	Gateway    string   `json:"gateway,omitempty"`
	DNS        []string `json:"dns,omitempty"`
	Interfaces []string `json:"interfaces,omitempty"`
}

// HardwareConfig represents hardware configuration
type HardwareConfig struct {
	CPU           string   `json:"cpu,omitempty"`
	Memory        string   `json:"memory,omitempty"`
	Storage       []string `json:"storage,omitempty"`
	Graphics      string   `json:"graphics,omitempty"`
	Networking    string   `json:"networking,omitempty"`
	BootLoader    string   `json:"boot_loader,omitempty"`
	KernelModules []string `json:"kernel_modules,omitempty"`
}

// NewNixOSRepository creates a new NixOS repository parser
func NewNixOSRepository(path string, logger *logger.Logger) (*NixOSRepository, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository path does not exist: %s", path)
	}

	return &NixOSRepository{
		Path:    path,
		logger:  logger,
		configs: make(map[string]*NixOSConfig),
	}, nil
}

// ScanRepository scans the repository for NixOS configurations
func (r *NixOSRepository) ScanRepository() error {
	r.logger.Info(fmt.Sprintf("Scanning NixOS repository: %s", r.Path))

	err := filepath.WalkDir(r.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common non-config directories
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "result" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Process .nix files
		if strings.HasSuffix(path, ".nix") {
			config, err := r.parseNixFile(path)
			if err != nil {
				r.logger.Warn(fmt.Sprintf("Failed to parse %s: %v", path, err))
				return nil // Continue processing other files
			}

			if config != nil {
				relPath, _ := filepath.Rel(r.Path, path)
				config.Path = relPath
				r.configs[relPath] = config
				r.logger.Debug(fmt.Sprintf("Parsed config: %s (%s)", config.Name, config.Type))
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan repository: %w", err)
	}

	r.logger.Info(fmt.Sprintf("Found %d NixOS configurations", len(r.configs)))
	return nil
}

// parseNixFile parses a single .nix file
func (r *NixOSRepository) parseNixFile(filePath string) (*NixOSConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &NixOSConfig{
		Name:         filepath.Base(filePath),
		Options:      make(map[string]string),
		Services:     []string{},
		Packages:     []string{},
		Dependencies: []string{},
	}

	scanner := bufio.NewScanner(file)
	inComment := false
	braceLevel := 0

	// Regex patterns for parsing
	hostnameRegex := regexp.MustCompile(`networking\.hostName\s*=\s*"([^"]+)"`)
	serviceRegex := regexp.MustCompile(`services\.(\w+)\.enable\s*=\s*true`)
	packageRegex := regexp.MustCompile(`pkgs\.(\w+)`)
	importRegex := regexp.MustCompile(`import\s*\[\s*([^\]]+)\s*\]`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Handle multi-line comments
		if strings.Contains(line, "/*") {
			inComment = true
		}
		if strings.Contains(line, "*/") {
			inComment = false
			continue
		}
		if inComment || strings.HasPrefix(line, "#") {
			continue
		}

		// Track brace level for context
		braceLevel += strings.Count(line, "{") - strings.Count(line, "}")

		// Determine configuration type
		if config.Type == "" {
			config.Type = r.determineConfigType(line, filePath)
		}

		// Extract hostname
		if matches := hostnameRegex.FindStringSubmatch(line); len(matches) > 1 {
			config.Hostname = matches[1]
		}

		// Extract enabled services
		if matches := serviceRegex.FindStringSubmatch(line); len(matches) > 1 {
			config.Services = append(config.Services, matches[1])
		}

		// Extract packages
		if matches := packageRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			for _, match := range matches {
				if len(match) > 1 {
					pkg := match[1]
					if !contains(config.Packages, pkg) {
						config.Packages = append(config.Packages, pkg)
					}
				}
			}
		}

		// Extract imports/dependencies
		if matches := importRegex.FindStringSubmatch(line); len(matches) > 1 {
			imports := strings.Split(matches[1], ",")
			for _, imp := range imports {
				imp = strings.TrimSpace(strings.Trim(imp, `"`))
				if imp != "" && !contains(config.Dependencies, imp) {
					config.Dependencies = append(config.Dependencies, imp)
				}
			}
		}

		// Extract network configuration
		config.NetworkConfig = r.parseNetworkConfig(line, config.NetworkConfig)

		// Extract hardware configuration
		config.Hardware = r.parseHardwareConfig(line, config.Hardware)

		// Extract other options
		r.parseOptions(line, config.Options)
	}

	// Skip empty or invalid configurations
	if config.Type == "" || (len(config.Services) == 0 && len(config.Packages) == 0 && config.Hostname == "") {
		return nil, nil
	}

	return config, scanner.Err()
}

// determineConfigType determines the type of NixOS configuration
func (r *NixOSRepository) determineConfigType(line, filePath string) string {
	fileName := filepath.Base(filePath)

	// Check by filename patterns
	switch {
	case fileName == "flake.nix":
		return "flake"
	case fileName == "home.nix" || strings.Contains(line, "home-manager"):
		return "home-manager"
	case fileName == "hardware-configuration.nix":
		return "hardware"
	case strings.Contains(fileName, "module") || strings.Contains(line, "lib.mkOption"):
		return "module"
	case fileName == "configuration.nix" || strings.Contains(line, "nixos"):
		return "system"
	default:
		// Try to determine from content
		if strings.Contains(line, "home.") {
			return "home-manager"
		} else if strings.Contains(line, "services.") || strings.Contains(line, "systemd.") {
			return "system"
		}
	}

	return "system" // Default fallback
}

// parseNetworkConfig extracts network configuration
func (r *NixOSRepository) parseNetworkConfig(line string, current *NetworkConfig) *NetworkConfig {
	if current == nil {
		current = &NetworkConfig{}
	}

	// Extract static IP
	if matches := regexp.MustCompile(`networking\.interfaces\.\w+\.ipv4\.addresses\s*=\s*\[\s*{\s*address\s*=\s*"([^"]+)"`).FindStringSubmatch(line); len(matches) > 1 {
		current.StaticIP = matches[1]
	}

	// Extract gateway
	if matches := regexp.MustCompile(`networking\.defaultGateway\s*=\s*"([^"]+)"`).FindStringSubmatch(line); len(matches) > 1 {
		current.Gateway = matches[1]
	}

	// Extract DNS
	if matches := regexp.MustCompile(`networking\.nameservers\s*=\s*\[\s*([^\]]+)\s*\]`).FindStringSubmatch(line); len(matches) > 1 {
		dns := strings.Split(matches[1], ",")
		for _, d := range dns {
			d = strings.TrimSpace(strings.Trim(d, `"`))
			if d != "" {
				current.DNS = append(current.DNS, d)
			}
		}
	}

	return current
}

// parseHardwareConfig extracts hardware configuration
func (r *NixOSRepository) parseHardwareConfig(line string, current *HardwareConfig) *HardwareConfig {
	if current == nil {
		current = &HardwareConfig{}
	}

	// Extract boot loader
	if matches := regexp.MustCompile(`boot\.loader\.(\w+)\.enable\s*=\s*true`).FindStringSubmatch(line); len(matches) > 1 {
		current.BootLoader = matches[1]
	}

	// Extract kernel modules
	if matches := regexp.MustCompile(`boot\.kernelModules\s*=\s*\[\s*([^\]]+)\s*\]`).FindStringSubmatch(line); len(matches) > 1 {
		modules := strings.Split(matches[1], ",")
		for _, mod := range modules {
			mod = strings.TrimSpace(strings.Trim(mod, `"`))
			if mod != "" {
				current.KernelModules = append(current.KernelModules, mod)
			}
		}
	}

	return current
}

// parseOptions extracts general configuration options
func (r *NixOSRepository) parseOptions(line string, options map[string]string) {
	// Extract key-value pairs
	if matches := regexp.MustCompile(`(\w+(?:\.\w+)*)\s*=\s*([^;]+);`).FindStringSubmatch(line); len(matches) > 2 {
		key := matches[1]
		value := strings.TrimSpace(matches[2])
		value = strings.Trim(value, `"`)

		// Skip common patterns that aren't useful options
		if !strings.Contains(key, "enable") && !strings.Contains(key, "services") && len(value) < 100 {
			options[key] = value
		}
	}
}

// GetConfigurations returns all parsed configurations
func (r *NixOSRepository) GetConfigurations() map[string]*NixOSConfig {
	return r.configs
}

// GetConfigurationsByType returns configurations filtered by type
func (r *NixOSRepository) GetConfigurationsByType(configType string) []*NixOSConfig {
	var result []*NixOSConfig
	for _, config := range r.configs {
		if config.Type == configType {
			result = append(result, config)
		}
	}
	return result
}

// GetMachineDefinitions extracts machine definitions that can be used for fleet management
func (r *NixOSRepository) GetMachineDefinitions() ([]*fleet.Machine, error) {
	var machines []*fleet.Machine

	for _, config := range r.configs {
		if config.Type == "system" && config.Hostname != "" {
			machine := &fleet.Machine{
				ID:          config.Hostname,
				Name:        config.Hostname,
				Address:     "", // Would need to be configured separately
				Tags:        []string{config.Type},
				Environment: "development", // Default, could be parsed from config
				Metadata: map[string]string{
					"config_path": config.Path,
					"services":    strings.Join(config.Services, ","),
					"packages":    strings.Join(config.Packages, ","),
				},
			}

			// Add network info if available
			if config.NetworkConfig != nil && config.NetworkConfig.StaticIP != "" {
				machine.Address = config.NetworkConfig.StaticIP
				machine.Metadata["static_ip"] = config.NetworkConfig.StaticIP
			}

			machines = append(machines, machine)
		}
	}

	return machines, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
