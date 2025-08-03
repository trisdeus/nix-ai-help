// Package context provides advanced AI context management for nixai
package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// EnhancedContextDetector implements advanced context detection for NixOS systems
type EnhancedContextDetector struct {
	logger *logger.Logger
}

// NewEnhancedContextDetector creates a new enhanced context detector
func NewEnhancedContextDetector(log *logger.Logger) *EnhancedContextDetector {
	return &EnhancedContextDetector{
		logger: log,
	}
}

// DetectEnhancedContext performs advanced context detection for NixOS systems
func (ecd *EnhancedContextDetector) DetectEnhancedContext(cfg *config.UserConfig) (*config.NixOSContext, error) {
	ecd.logger.Info("Detecting enhanced NixOS context")
	
	ctx := &config.NixOSContext{
		CacheValid: true,
	}
	
	// Detect system type
	ctx.SystemType = ecd.detectSystemType(cfg)
	ecd.logger.Debug(fmt.Sprintf("Detected system type: %s", ctx.SystemType))
	
	// Detect configuration approach
	ctx.UsesFlakes, ctx.UsesChannels = ecd.detectConfigurationApproach(cfg)
	ecd.logger.Debug(fmt.Sprintf("Configuration approach - Flakes: %t, Channels: %t", ctx.UsesFlakes, ctx.UsesChannels))
	
	// Detect Home Manager integration
	ctx.HasHomeManager, ctx.HomeManagerType, ctx.HomeManagerConfigPath = ecd.detectHomeManagerIntegration(cfg)
	ecd.logger.Debug(fmt.Sprintf("Home Manager - Has: %t, Type: %s, Path: %s", 
		ctx.HasHomeManager, ctx.HomeManagerType, ctx.HomeManagerConfigPath))
	
	// Detect version information
	ctx.NixOSVersion, ctx.NixVersion = ecd.detectVersionInformation()
	ecd.logger.Debug(fmt.Sprintf("Versions - NixOS: %s, Nix: %s", ctx.NixOSVersion, ctx.NixVersion))
	
	// Detect configuration files
	ctx.ConfigurationFiles = ecd.detectConfigurationFiles(cfg)
	ecd.logger.Debug(fmt.Sprintf("Detected %d configuration files", len(ctx.ConfigurationFiles)))
	
	// Detect enabled services
	ctx.EnabledServices = ecd.detectEnabledServices()
	ecd.logger.Debug(fmt.Sprintf("Detected %d enabled services", len(ctx.EnabledServices)))
	
	// Detect installed packages
	ctx.InstalledPackages = ecd.detectInstalledPackages()
	ecd.logger.Debug(fmt.Sprintf("Detected %d installed packages", len(ctx.InstalledPackages)))
	
	// Detect hardware configuration
	ctx.HardwareInfo = ecd.detectHardwareInfo()
	ecd.logger.Debug(fmt.Sprintf("Detected hardware info: %+v", ctx.HardwareInfo))
	
	// Detect network configuration
	ctx.NetworkInfo = ecd.detectNetworkInfo()
	ecd.logger.Debug(fmt.Sprintf("Detected network info: %+v", ctx.NetworkInfo))
	
	// Detect security configuration
	ctx.SecurityInfo = ecd.detectSecurityInfo()
	ecd.logger.Debug(fmt.Sprintf("Detected security info: %+v", ctx.SecurityInfo))
	
	// Detect performance configuration
	ctx.PerformanceInfo = ecd.detectPerformanceInfo()
	ecd.logger.Debug(fmt.Sprintf("Detected performance info: %+v", ctx.PerformanceInfo))
	
	// Detect user environment
	ctx.UserEnvironment = ecd.detectUserEnvironment()
	ecd.logger.Debug(fmt.Sprintf("Detected user environment: %+v", ctx.UserEnvironment))
	
	// Set detection timestamp
	ctx.LastDetected = time.Now()
	
		// Set cache expiration (24 hours)
	ctx.CacheExpires = ctx.LastDetected.Add(24 * time.Hour)
	
	ecd.logger.Info("Enhanced NixOS context detection completed")
	return ctx, nil
}

// detectSystemType detects the system type
func (ecd *EnhancedContextDetector) detectSystemType(cfg *config.UserConfig) string {
	// Check for NixOS
	if _, err := os.Stat("/run/current-system"); err == nil {
		return "nixos"
	}
	
	// Check for nix-darwin
	if _, err := os.Stat("/run/current-darwin-system"); err == nil {
		return "nix-darwin"
	}
	
	// Check for Home Manager only
	if cfg.NixosFolder != "" {
		homeManagerPaths := []string{
			filepath.Join(cfg.NixosFolder, "home.nix"),
			filepath.Join(os.Getenv("HOME"), ".config/nixpkgs/home.nix"),
			filepath.Join(os.Getenv("HOME"), ".config/home-manager/home.nix"),
		}
		
		for _, path := range homeManagerPaths {
			if _, err := os.Stat(path); err == nil {
				return "home-manager-only"
			}
		}
	}
	
	return "unknown"
}

// detectConfigurationApproach detects whether the system uses flakes or channels
func (ecd *EnhancedContextDetector) detectConfigurationApproach(cfg *config.UserConfig) (bool, bool) {
	usesFlakes := false
	usesChannels := false
	
	// Check for flake.nix in standard locations
	flakePaths := []string{
		"/etc/nixos/flake.nix",
		filepath.Join(cfg.NixosFolder, "flake.nix"),
		"./flake.nix",
	}
	
	for _, path := range flakePaths {
		if _, err := os.Stat(path); err == nil {
			usesFlakes = true
			break
		}
	}
	
	// Check for configuration.nix in standard locations
	configPaths := []string{
		"/etc/nixos/configuration.nix",
		filepath.Join(cfg.NixosFolder, "configuration.nix"),
		"./configuration.nix",
	}
	
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			usesChannels = true
			break
		}
	}
	
	// If neither is found, check for nix-channel command
	if !usesFlakes && !usesChannels {
		if _, err := exec.LookPath("nix-channel"); err == nil {
			usesChannels = true
		}
	}
	
	return usesFlakes, usesChannels
}

// detectHomeManagerIntegration detects Home Manager integration
func (ecd *EnhancedContextDetector) detectHomeManagerIntegration(cfg *config.UserConfig) (bool, string, string) {
	// Check for home.nix in standard locations
	homePaths := []string{
		"/etc/nixos/home.nix",
		filepath.Join(cfg.NixosFolder, "home.nix"),
		filepath.Join(os.Getenv("HOME"), ".config/nixpkgs/home.nix"),
		filepath.Join(os.Getenv("HOME"), ".config/home-manager/home.nix"),
		"./home.nix",
	}
	
	for _, path := range homePaths {
		if _, err := os.Stat(path); err == nil {
			// Determine type based on presence of configuration.nix or flake.nix
			if _, configErr := os.Stat(filepath.Join(filepath.Dir(path), "configuration.nix")); configErr == nil {
				return true, "module", path
			}
			
			if _, flakeErr := os.Stat(filepath.Join(filepath.Dir(path), "flake.nix")); flakeErr == nil {
				return true, "module", path
			}
			
			return true, "standalone", path
		}
	}
	
	return false, "", ""
}

// detectVersionInformation detects NixOS and Nix versions
func (ecd *EnhancedContextDetector) detectVersionInformation() (string, string) {
	var nixosVersion, nixVersion string
	
	// Get NixOS version
	if output, err := exec.Command("nixos-version").Output(); err == nil {
		nixosVersion = strings.TrimSpace(string(output))
	}
	
	// Get Nix version
	if output, err := exec.Command("nix", "--version").Output(); err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), " ")
		if len(parts) >= 3 {
			nixVersion = parts[2]
		}
	}
	
	return nixosVersion, nixVersion
}

// detectConfigurationFiles detects configuration files
func (ecd *EnhancedContextDetector) detectConfigurationFiles(cfg *config.UserConfig) []string {
	var files []string
	
	// Standard configuration file paths
	standardPaths := []string{
		"/etc/nixos/configuration.nix",
		"/etc/nixos/hardware-configuration.nix",
		"/etc/nixos/home.nix",
		"/etc/nixos/flake.nix",
		filepath.Join(cfg.NixosFolder, "configuration.nix"),
		filepath.Join(cfg.NixosFolder, "hardware-configuration.nix"),
		filepath.Join(cfg.NixosFolder, "home.nix"),
		filepath.Join(cfg.NixosFolder, "flake.nix"),
		filepath.Join(os.Getenv("HOME"), ".config/nixpkgs/home.nix"),
		filepath.Join(os.Getenv("HOME"), ".config/home-manager/home.nix"),
		"./configuration.nix",
		"./home.nix",
		"./flake.nix",
	}
	
	// Check which files exist
	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}
	
	return files
}

// detectEnabledServices detects enabled system services
func (ecd *EnhancedContextDetector) detectEnabledServices() []string {
	var services []string
	
	// Get enabled systemd services
	if output, err := exec.Command("systemctl", "list-units", "--type=service", "--state=running").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 0 && strings.Contains(fields[0], ".service") {
				serviceName := strings.TrimSuffix(fields[0], ".service")
				services = append(services, serviceName)
			}
		}
	}
	
	return services
}

// detectInstalledPackages detects installed packages
func (ecd *EnhancedContextDetector) detectInstalledPackages() []string {
	var packages []string
	
	// Get installed packages from nix-env
	if output, err := exec.Command("nix-env", "-q").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line != "" {
				packages = append(packages, line)
			}
		}
	}
	
	return packages
}

// detectHardwareInfo detects hardware information
func (ecd *EnhancedContextDetector) detectHardwareInfo() *config.HardwareInfo {
	info := &config.HardwareInfo{}
	
	// Get CPU info
	if output, err := exec.Command("lscpu").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Model name:") {
				info.CPUModel = strings.TrimSpace(strings.TrimPrefix(line, "Model name:"))
			} else if strings.HasPrefix(line, "CPU(s):") {
				info.CPUCores = strings.TrimSpace(strings.TrimPrefix(line, "CPU(s):"))
			}
		}
	}
	
	// Get memory info
	if output, err := exec.Command("free", "-h").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 1 {
				info.Memory = fields[1]
			}
		}
	}
	
	// Get disk info
	if output, err := exec.Command("df", "-h", "/").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 4 {
				info.DiskSpace = fields[4]
			}
		}
	}
	
	return info
}

// detectNetworkInfo detects network information
func (ecd *EnhancedContextDetector) detectNetworkInfo() *config.NetworkInfo {
	info := &config.NetworkInfo{}
	
	// Get network interfaces
	if output, err := exec.Command("ip", "link", "show").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if strings.Contains(line, ":") && !strings.Contains(line, "LOOPBACK") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					info.Interfaces = append(info.Interfaces, strings.TrimSpace(parts[1]))
				}
			}
			// Skip every other line (interface details)
			i++
		}
	}
	
	// Get DNS info
	if output, err := exec.Command("cat", "/etc/resolv.conf").Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "nameserver") {
				fields := strings.Fields(line)
				if len(fields) > 1 {
					info.DNSServers = append(info.DNSServers, fields[1])
				}
			}
		}
	}
	
	return info
}

// detectSecurityInfo detects security information
func (ecd *EnhancedContextDetector) detectSecurityInfo() *config.SecurityInfo {
	info := &config.SecurityInfo{}
	
	// Check if firewall is enabled
	if _, err := os.Stat("/etc/nixos/firewall.nix"); err == nil {
		info.FirewallEnabled = true
	}
	
	// Check if SSH is enabled
	if output, err := exec.Command("systemctl", "is-active", "sshd").Output(); err == nil {
		if strings.TrimSpace(string(output)) == "active" {
			info.SSHEnabled = true
		}
	}
	
	return info
}

// detectPerformanceInfo detects performance information
func (ecd *EnhancedContextDetector) detectPerformanceInfo() *config.PerformanceInfo {
	info := &config.PerformanceInfo{}
	
	// Get load average
	if output, err := exec.Command("cat", "/proc/loadavg").Output(); err == nil {
		fields := strings.Fields(string(output))
		if len(fields) > 0 {
			info.LoadAverage = fields[0]
		}
	}
	
	// Get uptime
	if output, err := exec.Command("uptime").Output(); err == nil {
		info.Uptime = strings.TrimSpace(string(output))
	}
	
	return info
}

// detectUserEnvironment detects user environment information
func (ecd *EnhancedContextDetector) detectUserEnvironment() *config.UserEnvironment {
	env := &config.UserEnvironment{
		Shell:     os.Getenv("SHELL"),
		Editor:    os.Getenv("EDITOR"),
		Terminal:  os.Getenv("TERM"),
		Locale:    os.Getenv("LANG"),
		Timezone: os.Getenv("TZ"),
	}
	
	// Detect desktop environment
	if de := os.Getenv("XDG_CURRENT_DESKTOP"); de != "" {
		env.DesktopEnvironment = de
	} else if de := os.Getenv("DESKTOP_SESSION"); de != "" {
		env.DesktopEnvironment = de
	}
	
	// Detect window manager
	if wm := os.Getenv("WINDOW_MANAGER"); wm != "" {
		env.WindowManager = wm
	}
	
	return env
}

// GetContextSummary returns a summary of the detected context
func (ecd *EnhancedContextDetector) GetContextSummary(ctx *config.NixOSContext) string {
	if ctx == nil || !ctx.CacheValid {
		return "Context: Unknown/Not detected"
	}
	
	var parts []string
	
	parts = append(parts, fmt.Sprintf("System: %s", ctx.SystemType))
	
	if ctx.UsesFlakes {
		parts = append(parts, "Flakes: Yes")
	} else if ctx.UsesChannels {
		parts = append(parts, "Channels: Yes")
	}
	
	if ctx.HasHomeManager {
		parts = append(parts, fmt.Sprintf("Home Manager: %s", ctx.HomeManagerType))
	}
	
	if len(ctx.EnabledServices) > 0 {
		parts = append(parts, fmt.Sprintf("Services: %d", len(ctx.EnabledServices)))
	}
	
	return strings.Join(parts, " | ")
}

// IsComplexTask determines if a task is complex enough to warrant planning
func (ecd *EnhancedContextDetector) IsComplexTask(task string) bool {
	complexIndicators := []string{
		"setup", "install", "configure", "deploy", "migrate", 
		"multiple", "several", "many", "steps", "process",
		"environment", "development", "production",
	}
	
	taskLower := strings.ToLower(task)
	for _, indicator := range complexIndicators {
		if strings.Contains(taskLower, indicator) {
			return true
		}
	}
	
	// Also consider longer tasks as potentially more complex
	return len(strings.Fields(task)) > 10
}