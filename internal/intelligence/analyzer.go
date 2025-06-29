// Package intelligence provides advanced AI-powered context analysis and system understanding
package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// SystemAnalyzer provides deep analysis of NixOS system configuration
type SystemAnalyzer struct {
	logger          *logger.Logger
	contextDetector *nixos.ContextDetector
	analysisCache   map[string]*SystemAnalysis
	mu              sync.RWMutex
	lastAnalysis    time.Time
	cacheExpiry     time.Duration
}

// SystemAnalysis represents a comprehensive analysis of the system
type SystemAnalysis struct {
	// System Information
	SystemType         string `json:"system_type"`
	NixOSVersion       string `json:"nixos_version"`
	SystemArchitecture string `json:"system_architecture"`
	Hostname           string `json:"hostname"`

	// Configuration Analysis
	ConfigurationPath string         `json:"configuration_path"`
	UsesFlakes        bool           `json:"uses_flakes"`
	FlakeInputs       []FlakeInput   `json:"flake_inputs"`
	ConfigModules     []ConfigModule `json:"config_modules"`

	// Package Analysis
	InstalledPackages []PackageInfo             `json:"installed_packages"`
	PackageConflicts  []AnalyzerPackageConflict `json:"package_conflicts"`
	UnusedPackages    []string                  `json:"unused_packages"`
	SecurityPackages  []string                  `json:"security_packages"`

	// Service Analysis
	EnabledServices     []ServiceInfo             `json:"enabled_services"`
	ServiceDependencies map[string][]string       `json:"service_dependencies"`
	ServiceConflicts    []AnalyzerServiceConflict `json:"service_conflicts"`

	// Hardware Analysis
	HardwareFeatures []HardwareFeature `json:"hardware_features"`
	DriverStatus     []DriverInfo      `json:"driver_status"`

	// Security Analysis
	SecuritySettings SecurityAnalysis `json:"security_settings"`

	// Performance Analysis
	PerformanceMetrics PerformanceAnalysis `json:"performance_metrics"`

	// Analysis Metadata
	GeneratedAt      time.Time     `json:"generated_at"`
	AnalysisDuration time.Duration `json:"analysis_duration"`
	Confidence       float64       `json:"confidence"`
}

// FlakeInput represents a flake input dependency
type FlakeInput struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	LastUpdated string `json:"last_updated"`
	Pinned      bool   `json:"pinned"`
}

// ConfigModule represents a NixOS configuration module
type ConfigModule struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Type         string   `json:"type"` // system, user, hardware, etc.
	Enabled      bool     `json:"enabled"`
	Dependencies []string `json:"dependencies"`
	Options      []string `json:"options"`
}

// PackageInfo represents detailed package information
type PackageInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Category     string   `json:"category"`
	Dependencies []string `json:"dependencies"`
	Size         int64    `json:"size"`
	InstallDate  string   `json:"install_date"`
	Purpose      string   `json:"purpose"`
	Critical     bool     `json:"critical"`
}

// AnalyzerPackageConflict represents a package conflict detected during analysis
type AnalyzerPackageConflict struct {
	Package1     string `json:"package1"`
	Package2     string `json:"package2"`
	ConflictType string `json:"conflict_type"` // version, dependency, feature
	Severity     string `json:"severity"`
	Resolution   string `json:"resolution"`
}

// ServiceInfo represents service configuration and status
type ServiceInfo struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // systemd, timer, socket, etc.
	Status       string            `json:"status"`
	Enabled      bool              `json:"enabled"`
	Port         int               `json:"port,omitempty"`
	Dependencies []string          `json:"dependencies"`
	Resources    map[string]string `json:"resources"`
	Security     ServiceSecurity   `json:"security"`
}

// AnalyzerServiceConflict represents service conflicts from analysis
type AnalyzerServiceConflict struct {
	Service1     string `json:"service1"`
	Service2     string `json:"service2"`
	ConflictType string `json:"conflict_type"` // port, resource, dependency
	Severity     string `json:"severity"`
	Resolution   string `json:"resolution"`
}

// ServiceSecurity represents service security configuration
type ServiceSecurity struct {
	Sandboxed       bool     `json:"sandboxed"`
	NetworkAccess   string   `json:"network_access"`
	FilePermissions []string `json:"file_permissions"`
	Capabilities    []string `json:"capabilities"`
	User            string   `json:"user"`
	Group           string   `json:"group"`
}

// HardwareFeature represents detected hardware features
type HardwareFeature struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"` // cpu, gpu, network, storage, etc.
	Status     string            `json:"status"`
	Supported  bool              `json:"supported"`
	Driver     string            `json:"driver"`
	Properties map[string]string `json:"properties"`
}

// DriverInfo represents driver information
type DriverInfo struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Status   string   `json:"status"`
	Hardware string   `json:"hardware"`
	Loaded   bool     `json:"loaded"`
	Issues   []string `json:"issues"`
}

// SecurityAnalysis represents security configuration analysis
type SecurityAnalysis struct {
	FirewallEnabled   bool             `json:"firewall_enabled"`
	SELinuxStatus     string           `json:"selinux_status"`
	AppArmorStatus    string           `json:"apparmor_status"`
	SecureBootEnabled bool             `json:"secure_boot_enabled"`
	EncryptionStatus  EncryptionStatus `json:"encryption_status"`
	UserSecurity      UserSecurity     `json:"user_security"`
	NetworkSecurity   NetworkSecurity  `json:"network_security"`
	Vulnerabilities   []Vulnerability  `json:"vulnerabilities"`
	SecurityScore     float64          `json:"security_score"`
}

// EncryptionStatus represents disk encryption status
type EncryptionStatus struct {
	FullDiskEncryption  bool     `json:"full_disk_encryption"`
	EncryptedPartitions []string `json:"encrypted_partitions"`
	EncryptionMethod    string   `json:"encryption_method"`
}

// UserSecurity represents user security configuration
type UserSecurity struct {
	RootPasswordSet      bool     `json:"root_password_set"`
	SudoUsers            []string `json:"sudo_users"`
	PasswordPolicy       string   `json:"password_policy"`
	SSHKeyAuthentication bool     `json:"ssh_key_authentication"`
}

// NetworkSecurity represents network security configuration
type NetworkSecurity struct {
	OpenPorts       []int `json:"open_ports"`
	SSHEnabled      bool  `json:"ssh_enabled"`
	SSHPasswordAuth bool  `json:"ssh_password_auth"`
	VPNConfigured   bool  `json:"vpn_configured"`
	DNSSecEnabled   bool  `json:"dnssec_enabled"`
}

// Vulnerability represents a detected security vulnerability
type Vulnerability struct {
	ID           string  `json:"id"`
	Severity     string  `json:"severity"`
	Package      string  `json:"package"`
	Description  string  `json:"description"`
	CVSS         float64 `json:"cvss"`
	FixAvailable bool    `json:"fix_available"`
}

// PerformanceAnalysis represents system performance analysis
type PerformanceAnalysis struct {
	CPUUsage          float64                   `json:"cpu_usage"`
	MemoryUsage       float64                   `json:"memory_usage"`
	DiskUsage         map[string]float64        `json:"disk_usage"`
	NetworkThroughput map[string]float64        `json:"network_throughput"`
	LoadAverage       []float64                 `json:"load_average"`
	BootTime          time.Duration             `json:"boot_time"`
	ServiceStartTimes map[string]time.Duration  `json:"service_start_times"`
	Bottlenecks       []PerformanceBottleneck   `json:"bottlenecks"`
	Optimizations     []PerformanceOptimization `json:"optimizations"`
}

// PerformanceBottleneck represents identified performance bottlenecks
type PerformanceBottleneck struct {
	Type        string  `json:"type"` // cpu, memory, disk, network
	Severity    string  `json:"severity"`
	Component   string  `json:"component"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
	Resolution  string  `json:"resolution"`
}

// PerformanceOptimization represents suggested performance optimizations
type PerformanceOptimization struct {
	Category    string  `json:"category"`
	Priority    string  `json:"priority"`
	Description string  `json:"description"`
	Command     string  `json:"command"`
	Impact      float64 `json:"impact"`
	Risk        string  `json:"risk"`
}

// NewSystemAnalyzer creates a new system analyzer
func NewSystemAnalyzer(log *logger.Logger) *SystemAnalyzer {
	return &SystemAnalyzer{
		logger:          log,
		contextDetector: nixos.NewContextDetector(log),
		analysisCache:   make(map[string]*SystemAnalysis),
		cacheExpiry:     30 * time.Minute, // Cache analysis for 30 minutes
	}
}

// AnalyzeSystem performs comprehensive system analysis
func (sa *SystemAnalyzer) AnalyzeSystem(ctx context.Context, cfg *config.UserConfig) (*SystemAnalysis, error) {
	startTime := time.Now()
	sa.logger.Info("Starting comprehensive system analysis")

	// Check cache first
	cacheKey := sa.generateCacheKey(cfg)
	if analysis := sa.getCachedAnalysis(cacheKey); analysis != nil {
		sa.logger.Info("Returning cached system analysis")
		return analysis, nil
	}

	analysis := &SystemAnalysis{
		GeneratedAt: startTime,
	}

	// Get basic NixOS context
	nixosCtx, err := sa.contextDetector.GetContext(cfg)
	if err != nil {
		sa.logger.Warn(fmt.Sprintf("Failed to get NixOS context: %v", err))
		analysis.Confidence = 0.3
	} else {
		analysis.Confidence = 0.8
		sa.populateBasicInfo(analysis, nixosCtx)
	}

	// Perform detailed analysis
	sa.analyzeConfiguration(ctx, analysis, nixosCtx)
	sa.analyzePackages(ctx, analysis)
	sa.analyzeServices(ctx, analysis)
	sa.analyzeHardware(ctx, analysis)
	sa.analyzeSecurity(ctx, analysis)
	sa.analyzePerformance(ctx, analysis)

	// Finalize analysis
	analysis.AnalysisDuration = time.Since(startTime)

	// Cache the result
	sa.cacheAnalysis(cacheKey, analysis)

	sa.logger.Info(fmt.Sprintf("System analysis completed in %v (confidence: %.1f%%)",
		analysis.AnalysisDuration, analysis.Confidence*100))

	return analysis, nil
}

// populateBasicInfo populates basic system information
func (sa *SystemAnalyzer) populateBasicInfo(analysis *SystemAnalysis, nixosCtx *config.NixOSContext) {
	if nixosCtx == nil {
		return
	}

	analysis.SystemType = nixosCtx.SystemType
	analysis.ConfigurationPath = nixosCtx.ConfigurationNix
	analysis.UsesFlakes = nixosCtx.UsesFlakes

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		analysis.Hostname = hostname
	}

	// Get architecture
	analysis.SystemArchitecture = sa.getSystemArchitecture()

	// Get NixOS version
	analysis.NixOSVersion = sa.getNixOSVersion()
}

// analyzeConfiguration analyzes NixOS configuration
func (sa *SystemAnalyzer) analyzeConfiguration(ctx context.Context, analysis *SystemAnalysis, nixosCtx *config.NixOSContext) {
	sa.logger.Debug("Analyzing NixOS configuration")

	if nixosCtx == nil {
		return
	}

	// Analyze flake inputs if using flakes
	if analysis.UsesFlakes {
		analysis.FlakeInputs = sa.analyzeFlakeInputs(nixosCtx.ConfigurationNix)
	}

	// Analyze configuration modules
	analysis.ConfigModules = sa.analyzeConfigModules(nixosCtx.ConfigurationNix)
}

// analyzePackages analyzes installed packages
func (sa *SystemAnalyzer) analyzePackages(ctx context.Context, analysis *SystemAnalysis) {
	sa.logger.Debug("Analyzing installed packages")

	// Get installed packages
	packages := sa.getInstalledPackages()
	analysis.InstalledPackages = packages

	// Detect package conflicts
	analysis.PackageConflicts = sa.detectPackageConflicts(packages)

	// Identify unused packages
	analysis.UnusedPackages = sa.identifyUnusedPackages(packages)

	// Identify security-related packages
	analysis.SecurityPackages = sa.identifySecurityPackages(packages)
}

// analyzeServices analyzes system services
func (sa *SystemAnalyzer) analyzeServices(ctx context.Context, analysis *SystemAnalysis) {
	sa.logger.Debug("Analyzing system services")

	// Get enabled services
	services := sa.getEnabledServices()
	analysis.EnabledServices = services

	// Analyze service dependencies
	analysis.ServiceDependencies = sa.analyzeServiceDependencies(services)

	// Detect service conflicts
	analysis.ServiceConflicts = sa.detectServiceConflicts(services)
}

// analyzeHardware analyzes hardware configuration
func (sa *SystemAnalyzer) analyzeHardware(ctx context.Context, analysis *SystemAnalysis) {
	sa.logger.Debug("Analyzing hardware configuration")

	// Detect hardware features
	analysis.HardwareFeatures = sa.detectHardwareFeatures()

	// Analyze driver status
	analysis.DriverStatus = sa.analyzeDriverStatus()
}

// analyzeSecurity analyzes security configuration
func (sa *SystemAnalyzer) analyzeSecurity(ctx context.Context, analysis *SystemAnalysis) {
	sa.logger.Debug("Analyzing security configuration")

	secAnalysis := SecurityAnalysis{}

	// Analyze firewall status
	secAnalysis.FirewallEnabled = sa.isFirewallEnabled()

	// Analyze encryption status
	secAnalysis.EncryptionStatus = sa.analyzeEncryption()

	// Analyze user security
	secAnalysis.UserSecurity = sa.analyzeUserSecurity()

	// Analyze network security
	secAnalysis.NetworkSecurity = sa.analyzeNetworkSecurity()

	// Scan for vulnerabilities
	secAnalysis.Vulnerabilities = sa.scanVulnerabilities()

	// Calculate security score
	secAnalysis.SecurityScore = sa.calculateSecurityScore(secAnalysis)

	analysis.SecuritySettings = secAnalysis
}

// analyzePerformance analyzes system performance
func (sa *SystemAnalyzer) analyzePerformance(ctx context.Context, analysis *SystemAnalysis) {
	sa.logger.Debug("Analyzing system performance")

	perfAnalysis := PerformanceAnalysis{}

	// Get current resource usage
	perfAnalysis.CPUUsage = sa.getCPUUsage()
	perfAnalysis.MemoryUsage = sa.getMemoryUsage()
	perfAnalysis.DiskUsage = sa.getDiskUsage()
	perfAnalysis.LoadAverage = sa.getLoadAverage()

	// Analyze boot time
	perfAnalysis.BootTime = sa.getBootTime()

	// Analyze service start times
	perfAnalysis.ServiceStartTimes = sa.getServiceStartTimes()

	// Identify bottlenecks
	perfAnalysis.Bottlenecks = sa.identifyBottlenecks(perfAnalysis)

	// Generate optimization suggestions
	perfAnalysis.Optimizations = sa.generateOptimizations(perfAnalysis)

	analysis.PerformanceMetrics = perfAnalysis
}

// Helper methods for cache management
func (sa *SystemAnalyzer) generateCacheKey(cfg *config.UserConfig) string {
	// Generate a cache key based on configuration hash
	configHash := utils.HashString(fmt.Sprintf("%+v", cfg))
	return fmt.Sprintf("system_analysis_%s", configHash)
}

func (sa *SystemAnalyzer) getCachedAnalysis(key string) *SystemAnalysis {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	if analysis, exists := sa.analysisCache[key]; exists {
		if time.Since(analysis.GeneratedAt) < sa.cacheExpiry {
			return analysis
		}
		// Remove expired cache entry
		delete(sa.analysisCache, key)
	}
	return nil
}

func (sa *SystemAnalyzer) cacheAnalysis(key string, analysis *SystemAnalysis) {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Store in cache
	sa.analysisCache[key] = analysis
	sa.lastAnalysis = time.Now()

	// Clean up old cache entries (keep only 5 most recent)
	if len(sa.analysisCache) > 5 {
		oldestKey := ""
		oldestTime := time.Now()
		for k, v := range sa.analysisCache {
			if v.GeneratedAt.Before(oldestTime) {
				oldestTime = v.GeneratedAt
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(sa.analysisCache, oldestKey)
		}
	}
}

// GetLastAnalysis returns the most recent analysis
func (sa *SystemAnalyzer) GetLastAnalysis() *SystemAnalysis {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	var latest *SystemAnalysis
	var latestTime time.Time

	for _, analysis := range sa.analysisCache {
		if analysis.GeneratedAt.After(latestTime) {
			latestTime = analysis.GeneratedAt
			latest = analysis
		}
	}

	return latest
}

// ClearCache clears the analysis cache
func (sa *SystemAnalyzer) ClearCache() {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	sa.analysisCache = make(map[string]*SystemAnalysis)
	sa.logger.Info("System analysis cache cleared")
}

// ExportAnalysis exports analysis to JSON file
func (sa *SystemAnalyzer) ExportAnalysis(analysis *SystemAnalysis, filename string) error {
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write analysis file: %w", err)
	}

	sa.logger.Info(fmt.Sprintf("System analysis exported to %s", filename))
	return nil
}
