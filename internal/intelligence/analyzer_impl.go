// System analyzer implementation methods
package intelligence

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// getSystemArchitecture detects the system architecture
func (sa *SystemAnalyzer) getSystemArchitecture() string {
	if output, err := exec.Command("uname", "-m").Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "unknown"
}

// getNixOSVersion gets the NixOS version
func (sa *SystemAnalyzer) getNixOSVersion() string {
	// Try to get version from /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VERSION=") {
				version := strings.TrimPrefix(line, "VERSION=")
				version = strings.Trim(version, "\"")
				return version
			}
		}
	}

	// Fallback to nixos-version command
	if output, err := exec.Command("nixos-version").Output(); err == nil {
		return strings.TrimSpace(string(output))
	}

	return "unknown"
}

// analyzeFlakeInputs analyzes flake inputs
func (sa *SystemAnalyzer) analyzeFlakeInputs(configPath string) []FlakeInput {
	var inputs []FlakeInput

	// Look for flake.nix in the configuration directory
	flakePath := filepath.Join(filepath.Dir(configPath), "flake.nix")
	if !fileExists(flakePath) {
		return inputs
	}

	// Parse flake.nix for inputs
	if data, err := os.ReadFile(flakePath); err == nil {
		content := string(data)

		// Simple regex to find inputs (this could be improved with proper Nix parsing)
		inputsRegex := regexp.MustCompile(`inputs\s*=\s*{([^}]+)}`)
		if matches := inputsRegex.FindStringSubmatch(content); len(matches) > 1 {
			inputsSection := matches[1]

			// Parse individual inputs
			inputRegex := regexp.MustCompile(`(\w+)\s*=\s*{[^}]*url\s*=\s*"([^"]+)"`)
			for _, match := range inputRegex.FindAllStringSubmatch(inputsSection, -1) {
				if len(match) >= 3 {
					input := FlakeInput{
						Name: match[1],
						URL:  match[2],
						Type: sa.determineFlakeInputType(match[2]),
					}
					inputs = append(inputs, input)
				}
			}
		}
	}

	return inputs
}

// determineFlakeInputType determines the type of a flake input from its URL
func (sa *SystemAnalyzer) determineFlakeInputType(url string) string {
	if strings.Contains(url, "github:") {
		return "github"
	} else if strings.Contains(url, "gitlab:") {
		return "gitlab"
	} else if strings.Contains(url, "git+") {
		return "git"
	} else if strings.Contains(url, "nixpkgs") {
		return "nixpkgs"
	}
	return "unknown"
}

// analyzeConfigModules analyzes NixOS configuration modules
func (sa *SystemAnalyzer) analyzeConfigModules(configPath string) []ConfigModule {
	var modules []ConfigModule

	if !fileExists(configPath) {
		return modules
	}

	// Parse configuration.nix for imports and modules
	if data, err := os.ReadFile(configPath); err == nil {
		content := string(data)

		// Find imports
		importsRegex := regexp.MustCompile(`imports\s*=\s*\[([^\]]+)\]`)
		if matches := importsRegex.FindStringSubmatch(content); len(matches) > 1 {
			importsSection := matches[1]

			// Extract individual import paths
			pathRegex := regexp.MustCompile(`[./\w-]+\.nix`)
			paths := pathRegex.FindAllString(importsSection, -1)

			for _, path := range paths {
				module := ConfigModule{
					Name:    filepath.Base(path),
					Path:    path,
					Type:    sa.determineModuleType(path),
					Enabled: true,
				}
				modules = append(modules, module)
			}
		}
	}

	return modules
}

// determineModuleType determines the type of a configuration module
func (sa *SystemAnalyzer) determineModuleType(path string) string {
	if strings.Contains(path, "hardware") {
		return "hardware"
	} else if strings.Contains(path, "service") {
		return "service"
	} else if strings.Contains(path, "desktop") {
		return "desktop"
	} else if strings.Contains(path, "user") {
		return "user"
	}
	return "system"
}

// getInstalledPackages gets information about installed packages
func (sa *SystemAnalyzer) getInstalledPackages() []PackageInfo {
	var packages []PackageInfo

	// Get packages from nix-env
	cmd := exec.Command("nix-env", "-q", "--installed")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse package info (format: package-version)
			parts := strings.Split(line, "-")
			if len(parts) >= 2 {
				name := strings.Join(parts[:len(parts)-1], "-")
				version := parts[len(parts)-1]

				pkg := PackageInfo{
					Name:     name,
					Version:  version,
					Category: sa.categorizePackage(name),
					Critical: sa.isCriticalPackage(name),
				}
				packages = append(packages, pkg)
			}
		}
	}

	return packages
}

// categorizePackage categorizes a package based on its name
func (sa *SystemAnalyzer) categorizePackage(name string) string {
	categories := map[string][]string{
		"system":       {"systemd", "kernel", "glibc", "bash", "coreutils"},
		"desktop":      {"gnome", "kde", "xorg", "wayland", "gtk", "qt"},
		"development":  {"gcc", "python", "nodejs", "rust", "go", "vim", "emacs", "vscode"},
		"security":     {"gnupg", "openssh", "openssl", "firewall", "apparmor"},
		"multimedia":   {"ffmpeg", "vlc", "gimp", "blender", "audacity"},
		"network":      {"curl", "wget", "firefox", "chromium", "thunderbird"},
		"productivity": {"libreoffice", "texlive", "pandoc", "git"},
	}

	for category, packages := range categories {
		for _, pkg := range packages {
			if strings.Contains(strings.ToLower(name), pkg) {
				return category
			}
		}
	}

	return "other"
}

// isCriticalPackage determines if a package is critical to system operation
func (sa *SystemAnalyzer) isCriticalPackage(name string) bool {
	criticalPackages := []string{
		"systemd", "kernel", "glibc", "bash", "coreutils", "nixos",
		"openssh", "networkmanager", "systemd-boot", "grub",
	}

	for _, critical := range criticalPackages {
		if strings.Contains(strings.ToLower(name), critical) {
			return true
		}
	}

	return false
}

// detectPackageConflicts detects potential package conflicts
func (sa *SystemAnalyzer) detectPackageConflicts(packages []PackageInfo) []AnalyzerPackageConflict {
	var conflicts []AnalyzerPackageConflict

	// Check for known conflicting packages
	conflictPairs := map[string]string{
		"pulseaudio": "pipewire",
		"x11":        "wayland",
		"systemd":    "openrc",
	}

	packageMap := make(map[string]bool)
	for _, pkg := range packages {
		packageMap[pkg.Name] = true
	}

	for pkg1, pkg2 := range conflictPairs {
		if packageMap[pkg1] && packageMap[pkg2] {
			conflict := AnalyzerPackageConflict{
				Package1:     pkg1,
				Package2:     pkg2,
				ConflictType: "feature",
				Severity:     "medium",
				Resolution:   fmt.Sprintf("Choose between %s and %s", pkg1, pkg2),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// identifyUnusedPackages identifies potentially unused packages
func (sa *SystemAnalyzer) identifyUnusedPackages(packages []PackageInfo) []string {
	var unused []string

	// Simple heuristic: packages that might be unused
	// This is a basic implementation and could be improved with actual usage tracking
	developmentPackages := []string{"gcc", "make", "cmake", "autotools"}

	for _, pkg := range packages {
		for _, devPkg := range developmentPackages {
			if strings.Contains(strings.ToLower(pkg.Name), devPkg) {
				// Check if used recently (placeholder logic)
				unused = append(unused, pkg.Name)
				break
			}
		}
	}

	return unused
}

// identifySecurityPackages identifies security-related packages
func (sa *SystemAnalyzer) identifySecurityPackages(packages []PackageInfo) []string {
	var security []string

	securityKeywords := []string{"security", "crypt", "ssl", "tls", "gpg", "ssh", "firewall", "antivirus"}

	for _, pkg := range packages {
		for _, keyword := range securityKeywords {
			if strings.Contains(strings.ToLower(pkg.Name), keyword) {
				security = append(security, pkg.Name)
				break
			}
		}
	}

	return security
}

// getEnabledServices gets information about enabled services
func (sa *SystemAnalyzer) getEnabledServices() []ServiceInfo {
	var services []ServiceInfo

	// Get systemd services
	cmd := exec.Command("systemctl", "list-unit-files", "--type=service", "--state=enabled")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 && strings.HasSuffix(fields[0], ".service") {
				serviceName := strings.TrimSuffix(fields[0], ".service")

				service := ServiceInfo{
					Name:     serviceName,
					Type:     "systemd",
					Status:   sa.getServiceStatus(serviceName),
					Enabled:  true,
					Security: sa.analyzeServiceSecurity(serviceName),
				}
				services = append(services, service)
			}
		}
	}

	return services
}

// getServiceStatus gets the status of a specific service
func (sa *SystemAnalyzer) getServiceStatus(serviceName string) string {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return "unknown"
}

// analyzeServiceSecurity analyzes security configuration of a service
func (sa *SystemAnalyzer) analyzeServiceSecurity(serviceName string) ServiceSecurity {
	security := ServiceSecurity{}

	// Get service configuration
	cmd := exec.Command("systemctl", "show", serviceName)
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "User=") {
				security.User = strings.TrimPrefix(line, "User=")
			} else if strings.HasPrefix(line, "Group=") {
				security.Group = strings.TrimPrefix(line, "Group=")
			} else if strings.HasPrefix(line, "PrivateNetwork=") {
				security.NetworkAccess = strings.TrimPrefix(line, "PrivateNetwork=")
			}
		}
	}

	return security
}

// analyzeServiceDependencies analyzes service dependencies
func (sa *SystemAnalyzer) analyzeServiceDependencies(services []ServiceInfo) map[string][]string {
	dependencies := make(map[string][]string)

	for _, service := range services {
		cmd := exec.Command("systemctl", "list-dependencies", service.Name)
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(string(output), "\n")
			var deps []string

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.Contains(line, "●") || strings.Contains(line, "○") {
					// Extract service name
					parts := strings.Fields(line)
					if len(parts) > 1 {
						depName := parts[len(parts)-1]
						if strings.HasSuffix(depName, ".service") {
							depName = strings.TrimSuffix(depName, ".service")
							deps = append(deps, depName)
						}
					}
				}
			}

			dependencies[service.Name] = deps
		}
	}

	return dependencies
}

// detectServiceConflicts detects service conflicts
func (sa *SystemAnalyzer) detectServiceConflicts(services []ServiceInfo) []AnalyzerServiceConflict {
	var conflicts []AnalyzerServiceConflict

	// Check for port conflicts
	portMap := make(map[int][]string)
	for _, service := range services {
		if service.Port > 0 {
			portMap[service.Port] = append(portMap[service.Port], service.Name)
		}
	}

	for port, serviceNames := range portMap {
		if len(serviceNames) > 1 {
			for i := 0; i < len(serviceNames)-1; i++ {
				conflict := AnalyzerServiceConflict{
					Service1:     serviceNames[i],
					Service2:     serviceNames[i+1],
					ConflictType: "port",
					Severity:     "high",
					Resolution:   fmt.Sprintf("Configure different ports for services using port %d", port),
				}
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// detectHardwareFeatures detects hardware features
func (sa *SystemAnalyzer) detectHardwareFeatures() []HardwareFeature {
	var features []HardwareFeature

	// CPU features
	if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		content := string(data)
		if strings.Contains(content, "vmx") || strings.Contains(content, "svm") {
			features = append(features, HardwareFeature{
				Name:      "Hardware Virtualization",
				Type:      "cpu",
				Status:    "available",
				Supported: true,
				Driver:    "kvm",
			})
		}
	}

	// GPU features
	cmd := exec.Command("lspci", "-nn")
	if output, err := cmd.Output(); err == nil {
		content := string(output)
		if strings.Contains(content, "VGA") || strings.Contains(content, "3D") {
			features = append(features, HardwareFeature{
				Name:      "Graphics Controller",
				Type:      "gpu",
				Status:    "detected",
				Supported: true,
			})
		}
	}

	return features
}

// analyzeDriverStatus analyzes driver status
func (sa *SystemAnalyzer) analyzeDriverStatus() []DriverInfo {
	var drivers []DriverInfo

	// Get loaded kernel modules
	cmd := exec.Command("lsmod")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 1 && fields[0] != "Module" {
				driver := DriverInfo{
					Name:   fields[0],
					Status: "loaded",
					Loaded: true,
				}
				drivers = append(drivers, driver)
			}
		}
	}

	return drivers
}

// Security analysis methods

// isFirewallEnabled checks if firewall is enabled
func (sa *SystemAnalyzer) isFirewallEnabled() bool {
	// Check iptables
	cmd := exec.Command("iptables", "-L")
	if err := cmd.Run(); err == nil {
		return true
	}

	// Check ufw
	cmd = exec.Command("ufw", "status")
	if output, err := cmd.Output(); err == nil {
		return strings.Contains(string(output), "Status: active")
	}

	return false
}

// analyzeEncryption analyzes disk encryption status
func (sa *SystemAnalyzer) analyzeEncryption() EncryptionStatus {
	encryption := EncryptionStatus{}

	// Check for LUKS encrypted partitions
	cmd := exec.Command("lsblk", "-f")
	if output, err := cmd.Output(); err == nil {
		content := string(output)
		if strings.Contains(content, "crypto_LUKS") {
			encryption.FullDiskEncryption = true
			encryption.EncryptionMethod = "LUKS"
		}
	}

	return encryption
}

// analyzeUserSecurity analyzes user security configuration
func (sa *SystemAnalyzer) analyzeUserSecurity() UserSecurity {
	userSec := UserSecurity{}

	// Check sudo users
	if data, err := os.ReadFile("/etc/group"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "wheel:") || strings.HasPrefix(line, "sudo:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 4 && parts[3] != "" {
					userSec.SudoUsers = strings.Split(parts[3], ",")
				}
			}
		}
	}

	// Check SSH configuration
	if data, err := os.ReadFile("/etc/ssh/sshd_config"); err == nil {
		content := string(data)
		userSec.SSHKeyAuthentication = !strings.Contains(content, "PubkeyAuthentication no")
	}

	return userSec
}

// analyzeNetworkSecurity analyzes network security configuration
func (sa *SystemAnalyzer) analyzeNetworkSecurity() NetworkSecurity {
	netSec := NetworkSecurity{}

	// Check for open ports
	cmd := exec.Command("ss", "-tuln")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "LISTEN") {
				// Extract port number (simplified parsing)
				fields := strings.Fields(line)
				for _, field := range fields {
					if strings.Contains(field, ":") {
						parts := strings.Split(field, ":")
						if len(parts) >= 2 {
							if port, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
								netSec.OpenPorts = append(netSec.OpenPorts, port)
							}
						}
					}
				}
			}
		}
	}

	// Check SSH status
	netSec.SSHEnabled = sa.getServiceStatus("sshd") == "active"

	return netSec
}

// scanVulnerabilities scans for known vulnerabilities
func (sa *SystemAnalyzer) scanVulnerabilities() []Vulnerability {
	var vulnerabilities []Vulnerability

	// This is a placeholder - in a real implementation, this would integrate
	// with vulnerability databases like CVE, NVD, etc.

	return vulnerabilities
}

// calculateSecurityScore calculates an overall security score
func (sa *SystemAnalyzer) calculateSecurityScore(secAnalysis SecurityAnalysis) float64 {
	score := 0.0
	maxScore := 10.0

	// Firewall (2 points)
	if secAnalysis.FirewallEnabled {
		score += 2.0
	}

	// Encryption (3 points)
	if secAnalysis.EncryptionStatus.FullDiskEncryption {
		score += 3.0
	}

	// SSH security (2 points)
	if secAnalysis.UserSecurity.SSHKeyAuthentication {
		score += 1.0
	}
	if !secAnalysis.NetworkSecurity.SSHPasswordAuth {
		score += 1.0
	}

	// Network security (2 points)
	openPortCount := len(secAnalysis.NetworkSecurity.OpenPorts)
	if openPortCount == 0 {
		score += 2.0
	} else if openPortCount < 5 {
		score += 1.0
	}

	// Vulnerabilities (1 point)
	if len(secAnalysis.Vulnerabilities) == 0 {
		score += 1.0
	}

	return (score / maxScore) * 10.0
}

// Performance analysis methods

// getCPUUsage gets current CPU usage
func (sa *SystemAnalyzer) getCPUUsage() float64 {
	// Read from /proc/loadavg or use top command
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 1 {
			if load, err := strconv.ParseFloat(fields[0], 64); err == nil {
				return load
			}
		}
	}
	return 0.0
}

// getMemoryUsage gets current memory usage
func (sa *SystemAnalyzer) getMemoryUsage() float64 {
	cmd := exec.Command("free", "-m")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Mem:") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					total, _ := strconv.ParseFloat(fields[1], 64)
					used, _ := strconv.ParseFloat(fields[2], 64)
					if total > 0 {
						return (used / total) * 100.0
					}
				}
			}
		}
	}
	return 0.0
}

// getDiskUsage gets disk usage for all mounted filesystems
func (sa *SystemAnalyzer) getDiskUsage() map[string]float64 {
	usage := make(map[string]float64)

	cmd := exec.Command("df", "-h")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 6 && strings.HasPrefix(fields[5], "/") {
				usageStr := strings.TrimSuffix(fields[4], "%")
				if usagePercent, err := strconv.ParseFloat(usageStr, 64); err == nil {
					usage[fields[5]] = usagePercent
				}
			}
		}
	}

	return usage
}

// getLoadAverage gets system load average
func (sa *SystemAnalyzer) getLoadAverage() []float64 {
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		var loads []float64
		for i := 0; i < 3 && i < len(fields); i++ {
			if load, err := strconv.ParseFloat(fields[i], 64); err == nil {
				loads = append(loads, load)
			}
		}
		return loads
	}
	return []float64{0, 0, 0}
}

// getBootTime gets system boot time
func (sa *SystemAnalyzer) getBootTime() time.Duration {
	cmd := exec.Command("systemd-analyze")
	if output, err := cmd.Output(); err == nil {
		content := string(output)
		// Parse output like "Startup finished in 2.123s (kernel) + 15.456s (userspace) = 17.579s"
		if strings.Contains(content, "Startup finished in") {
			// Extract total time (simplified parsing)
			parts := strings.Split(content, "=")
			if len(parts) >= 2 {
				timeStr := strings.TrimSpace(parts[1])
				timeStr = strings.TrimSuffix(timeStr, "s")
				if bootTime, err := strconv.ParseFloat(timeStr, 64); err == nil {
					return time.Duration(bootTime * float64(time.Second))
				}
			}
		}
	}
	return 0
}

// getServiceStartTimes gets service start times
func (sa *SystemAnalyzer) getServiceStartTimes() map[string]time.Duration {
	times := make(map[string]time.Duration)

	cmd := exec.Command("systemd-analyze", "blame")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				timeStr := fields[0]
				serviceName := fields[1]

				// Parse time (format like "1.234s" or "123ms")
				if strings.HasSuffix(timeStr, "ms") {
					timeStr = strings.TrimSuffix(timeStr, "ms")
					if ms, err := strconv.ParseFloat(timeStr, 64); err == nil {
						times[serviceName] = time.Duration(ms * float64(time.Millisecond))
					}
				} else if strings.HasSuffix(timeStr, "s") {
					timeStr = strings.TrimSuffix(timeStr, "s")
					if s, err := strconv.ParseFloat(timeStr, 64); err == nil {
						times[serviceName] = time.Duration(s * float64(time.Second))
					}
				}
			}
		}
	}

	return times
}

// identifyBottlenecks identifies performance bottlenecks
func (sa *SystemAnalyzer) identifyBottlenecks(perfAnalysis PerformanceAnalysis) []PerformanceBottleneck {
	var bottlenecks []PerformanceBottleneck

	// CPU bottleneck
	if perfAnalysis.CPUUsage > 80.0 {
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Type:        "cpu",
			Severity:    "high",
			Component:   "CPU",
			Impact:      perfAnalysis.CPUUsage,
			Description: "High CPU usage detected",
			Resolution:  "Consider optimizing CPU-intensive processes or upgrading hardware",
		})
	}

	// Memory bottleneck
	if perfAnalysis.MemoryUsage > 90.0 {
		bottlenecks = append(bottlenecks, PerformanceBottleneck{
			Type:        "memory",
			Severity:    "high",
			Component:   "RAM",
			Impact:      perfAnalysis.MemoryUsage,
			Description: "High memory usage detected",
			Resolution:  "Consider adding more RAM or optimizing memory usage",
		})
	}

	// Disk bottleneck
	for mount, usage := range perfAnalysis.DiskUsage {
		if usage > 95.0 {
			bottlenecks = append(bottlenecks, PerformanceBottleneck{
				Type:        "disk",
				Severity:    "critical",
				Component:   mount,
				Impact:      usage,
				Description: fmt.Sprintf("Disk usage critical on %s", mount),
				Resolution:  "Free up disk space or expand storage",
			})
		}
	}

	return bottlenecks
}

// generateOptimizations generates performance optimization suggestions
func (sa *SystemAnalyzer) generateOptimizations(perfAnalysis PerformanceAnalysis) []PerformanceOptimization {
	var optimizations []PerformanceOptimization

	// Boot time optimization
	if perfAnalysis.BootTime > 30*time.Second {
		optimizations = append(optimizations, PerformanceOptimization{
			Category:    "boot",
			Priority:    "medium",
			Description: "Boot time can be improved by optimizing service startup",
			Command:     "systemd-analyze critical-chain",
			Impact:      float64(perfAnalysis.BootTime.Seconds()),
			Risk:        "low",
		})
	}

	// Memory optimization
	if perfAnalysis.MemoryUsage > 70.0 {
		optimizations = append(optimizations, PerformanceOptimization{
			Category:    "memory",
			Priority:    "high",
			Description: "Consider enabling zram or adding swap space",
			Command:     "nixos-rebuild switch --option services.zram.enable true",
			Impact:      perfAnalysis.MemoryUsage,
			Risk:        "low",
		})
	}

	return optimizations
}

// Utility functions

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
