package detector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"nix-ai-help/pkg/logger"
)

// SystemType represents the detected source system
type SystemType string

const (
	SystemUbuntu  SystemType = "ubuntu"
	SystemDebian  SystemType = "debian"
	SystemArch    SystemType = "arch"
	SystemCentOS  SystemType = "centos"
	SystemFedora  SystemType = "fedora"
	SystemMacOS   SystemType = "macos"
	SystemNixOS   SystemType = "nixos"
	SystemUnknown SystemType = "unknown"
)

// SystemInfo contains detected system information
type SystemInfo struct {
	Type           SystemType          `json:"type"`
	Version        string              `json:"version"`
	Architecture   string              `json:"architecture"`
	PackageManager string              `json:"package_manager"`
	InitSystem     string              `json:"init_system"`
	Shell          string              `json:"shell"`
	Environment    map[string]string   `json:"environment"`
	InstalledPkgs  []string            `json:"installed_packages"`
	Services       []ServiceInfo       `json:"services"`
	ConfigFiles    []ConfigFileInfo    `json:"config_files"`
	Users          []UserInfo          `json:"users"`
	Complexity     MigrationComplexity `json:"complexity"`
}

// ServiceInfo represents a running service
type ServiceInfo struct {
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Type        string            `json:"type"`
	Config      string            `json:"config_path"`
	Port        int               `json:"port,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// ConfigFileInfo represents a configuration file
type ConfigFileInfo struct {
	Path       string `json:"path"`
	Type       string `json:"type"`
	Service    string `json:"service,omitempty"`
	Importance string `json:"importance"` // critical, important, optional
	Backup     bool   `json:"needs_backup"`
}

// UserInfo represents user account information
type UserInfo struct {
	Name       string   `json:"name"`
	UID        int      `json:"uid"`
	GID        int      `json:"gid"`
	Home       string   `json:"home"`
	Shell      string   `json:"shell"`
	Groups     []string `json:"groups"`
	SystemUser bool     `json:"system_user"`
}

// MigrationComplexity represents migration difficulty
type MigrationComplexity string

const (
	ComplexitySimple       MigrationComplexity = "simple"
	ComplexityIntermediate MigrationComplexity = "intermediate"
	ComplexityComplex      MigrationComplexity = "complex"
	ComplexityExpert       MigrationComplexity = "expert"
)

// SystemDetector detects the current system configuration
type SystemDetector struct {
	logger logger.Logger
}

// NewSystemDetector creates a new system detector
func NewSystemDetector(logger logger.Logger) *SystemDetector {
	return &SystemDetector{
		logger: logger,
	}
}

// DetectSystem performs comprehensive system detection
func (sd *SystemDetector) DetectSystem(ctx context.Context) (*SystemInfo, error) {
	sd.logger.Info("Starting comprehensive system detection")

	info := &SystemInfo{
		Environment:   make(map[string]string),
		InstalledPkgs: []string{},
		Services:      []ServiceInfo{},
		ConfigFiles:   []ConfigFileInfo{},
		Users:         []UserInfo{},
	}

	// Detect basic system information
	if err := sd.detectBasicInfo(info); err != nil {
		return nil, fmt.Errorf("failed to detect basic system info: %v", err)
	}

	// Detect package manager and installed packages
	if err := sd.detectPackages(info); err != nil {
		sd.logger.Warn("Failed to detect packages: " + err.Error())
	}

	// Detect running services
	if err := sd.detectServices(info); err != nil {
		sd.logger.Warn("Failed to detect services: " + err.Error())
	}

	// Detect configuration files
	if err := sd.detectConfigFiles(info); err != nil {
		sd.logger.Warn("Failed to detect config files: " + err.Error())
	}

	// Detect users
	if err := sd.detectUsers(info); err != nil {
		sd.logger.Warn("Failed to detect users: " + err.Error())
	}

	// Calculate migration complexity
	info.Complexity = sd.calculateComplexity(info)

	sd.logger.Info(fmt.Sprintf("System detection completed: %s %s", info.Type, info.Version))
	return info, nil
}

// detectBasicInfo detects basic system information
func (sd *SystemDetector) detectBasicInfo(info *SystemInfo) error {
	// Check for NixOS first
	if _, err := os.Stat("/etc/nixos/configuration.nix"); err == nil {
		info.Type = SystemNixOS
		if version, err := sd.runCommand("nixos-version"); err == nil {
			info.Version = strings.TrimSpace(version)
		}
		return nil
	}

	// Check /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		content := string(data)

		if strings.Contains(content, "Ubuntu") {
			info.Type = SystemUbuntu
			info.PackageManager = "apt"
		} else if strings.Contains(content, "Debian") {
			info.Type = SystemDebian
			info.PackageManager = "apt"
		} else if strings.Contains(content, "Arch Linux") {
			info.Type = SystemArch
			info.PackageManager = "pacman"
		} else if strings.Contains(content, "CentOS") {
			info.Type = SystemCentOS
			info.PackageManager = "yum"
		} else if strings.Contains(content, "Fedora") {
			info.Type = SystemFedora
			info.PackageManager = "dnf"
		}

		// Extract version
		versionRegex := regexp.MustCompile(`VERSION="([^"]+)"`)
		if matches := versionRegex.FindStringSubmatch(content); len(matches) > 1 {
			info.Version = matches[1]
		}
	}

	// Check for macOS
	if output, err := sd.runCommand("sw_vers", "-productName"); err == nil && strings.Contains(output, "macOS") {
		info.Type = SystemMacOS
		info.PackageManager = "brew"
		if version, err := sd.runCommand("sw_vers", "-productVersion"); err == nil {
			info.Version = strings.TrimSpace(version)
		}
	}

	// Detect architecture
	if arch, err := sd.runCommand("uname", "-m"); err == nil {
		info.Architecture = strings.TrimSpace(arch)
	}

	// Detect init system
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		info.InitSystem = "systemd"
	} else if _, err := os.Stat("/sbin/init"); err == nil {
		info.InitSystem = "sysvinit"
	}

	// Detect shell
	if shell := os.Getenv("SHELL"); shell != "" {
		info.Shell = shell
	}

	// Collect environment variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			info.Environment[parts[0]] = parts[1]
		}
	}

	if info.Type == SystemUnknown {
		return fmt.Errorf("unable to detect system type")
	}

	return nil
}

// detectPackages detects installed packages
func (sd *SystemDetector) detectPackages(info *SystemInfo) error {
	var cmd []string

	switch info.PackageManager {
	case "apt":
		cmd = []string{"dpkg", "-l"}
	case "pacman":
		cmd = []string{"pacman", "-Q"}
	case "yum", "dnf":
		cmd = []string{info.PackageManager, "list", "installed"}
	case "brew":
		cmd = []string{"brew", "list"}
	default:
		return fmt.Errorf("unsupported package manager: %s", info.PackageManager)
	}

	output, err := sd.runCommand(cmd[0], cmd[1:]...)
	if err != nil {
		return err
	}

	// Parse package list (simplified)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Desired") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				info.InstalledPkgs = append(info.InstalledPkgs, fields[0])
			}
		}
	}

	return nil
}

// detectServices detects running services
func (sd *SystemDetector) detectServices(info *SystemInfo) error {
	if info.InitSystem == "systemd" {
		return sd.detectSystemdServices(info)
	}

	// Fallback to ps-based detection
	return sd.detectProcessServices(info)
}

// detectSystemdServices detects systemd services
func (sd *SystemDetector) detectSystemdServices(info *SystemInfo) error {
	output, err := sd.runCommand("systemctl", "list-units", "--type=service", "--state=active", "--no-pager")
	if err != nil {
		return err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ".service") && strings.Contains(line, "active") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				serviceName := strings.TrimSuffix(fields[0], ".service")
				service := ServiceInfo{
					Name:   serviceName,
					Status: "active",
					Type:   "systemd",
				}

				// Try to get more service info
				if configPath, err := sd.runCommand("systemctl", "show", "-p", "FragmentPath", fields[0]); err == nil {
					service.Config = strings.TrimPrefix(strings.TrimSpace(configPath), "FragmentPath=")
				}

				info.Services = append(info.Services, service)
			}
		}
	}

	return nil
}

// detectProcessServices detects services via process list
func (sd *SystemDetector) detectProcessServices(info *SystemInfo) error {
	output, err := sd.runCommand("ps", "aux")
	if err != nil {
		return err
	}

	// Common service patterns
	servicePatterns := []string{"nginx", "apache", "mysql", "postgres", "redis", "docker", "ssh"}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		for _, pattern := range servicePatterns {
			if strings.Contains(strings.ToLower(line), pattern) {
				service := ServiceInfo{
					Name:   pattern,
					Status: "running",
					Type:   "process",
				}
				info.Services = append(info.Services, service)
				break
			}
		}
	}

	return nil
}

// detectConfigFiles detects important configuration files
func (sd *SystemDetector) detectConfigFiles(info *SystemInfo) error {
	configPaths := []struct {
		path       string
		configType string
		service    string
		importance string
	}{
		{"/etc/nginx/nginx.conf", "nginx", "nginx", "critical"},
		{"/etc/apache2/apache2.conf", "apache", "apache", "critical"},
		{"/etc/mysql/my.cnf", "mysql", "mysql", "critical"},
		{"/etc/postgresql/postgresql.conf", "postgresql", "postgresql", "critical"},
		{"/etc/redis/redis.conf", "redis", "redis", "important"},
		{"/etc/ssh/sshd_config", "ssh", "ssh", "critical"},
		{"/etc/docker/daemon.json", "docker", "docker", "important"},
		{"/etc/hosts", "system", "", "important"},
		{"/etc/fstab", "system", "", "critical"},
		{"/etc/crontab", "system", "", "important"},
	}

	for _, config := range configPaths {
		if _, err := os.Stat(config.path); err == nil {
			info.ConfigFiles = append(info.ConfigFiles, ConfigFileInfo{
				Path:       config.path,
				Type:       config.configType,
				Service:    config.service,
				Importance: config.importance,
				Backup:     config.importance == "critical",
			})
		}
	}

	return nil
}

// detectUsers detects system users
func (sd *SystemDetector) detectUsers(info *SystemInfo) error {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) >= 7 {
			user := UserInfo{
				Name:  fields[0],
				Home:  fields[5],
				Shell: fields[6],
			}

			// Determine if system user (simplified)
			user.SystemUser = strings.HasPrefix(user.Home, "/var") ||
				strings.HasPrefix(user.Home, "/run") ||
				user.Shell == "/bin/false" ||
				user.Shell == "/usr/sbin/nologin"

			info.Users = append(info.Users, user)
		}
	}

	return scanner.Err()
}

// calculateComplexity determines migration complexity
func (sd *SystemDetector) calculateComplexity(info *SystemInfo) MigrationComplexity {
	score := 0

	// Base score by system type
	switch info.Type {
	case SystemUbuntu, SystemDebian:
		score += 1 // Easier to migrate
	case SystemArch:
		score += 2
	case SystemCentOS, SystemFedora:
		score += 3
	case SystemMacOS:
		score += 4 // More complex
	}

	// Service complexity
	if len(info.Services) > 20 {
		score += 3
	} else if len(info.Services) > 10 {
		score += 2
	} else if len(info.Services) > 5 {
		score += 1
	}

	// Config file complexity
	criticalConfigs := 0
	for _, config := range info.ConfigFiles {
		if config.Importance == "critical" {
			criticalConfigs++
		}
	}
	score += criticalConfigs / 2

	// User complexity
	nonSystemUsers := 0
	for _, user := range info.Users {
		if !user.SystemUser {
			nonSystemUsers++
		}
	}
	if nonSystemUsers > 5 {
		score += 2
	} else if nonSystemUsers > 2 {
		score += 1
	}

	// Determine complexity level
	switch {
	case score <= 3:
		return ComplexitySimple
	case score <= 6:
		return ComplexityIntermediate
	case score <= 10:
		return ComplexityComplex
	default:
		return ComplexityExpert
	}
}

// runCommand executes a command and returns output
func (sd *SystemDetector) runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
