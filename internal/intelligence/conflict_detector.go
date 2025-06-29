// Package intelligence provides conflict detection and resolution for NixOS configurations
package intelligence

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// ConflictDetector analyzes and detects conflicts in NixOS configurations
type ConflictDetector struct {
	logger   *logger.Logger
	analyzer *SystemAnalyzer
	cache    map[string]*ConflictAnalysis
	mu       sync.RWMutex
}

// ConflictAnalysis represents the result of conflict detection
type ConflictAnalysis struct {
	// Detected Conflicts
	PackageConflicts  []PackageConflict  `json:"package_conflicts"`
	ServiceConflicts  []ServiceConflict  `json:"service_conflicts"`
	ConfigConflicts   []ConfigConflict   `json:"config_conflicts"`
	PortConflicts     []PortConflict     `json:"port_conflicts"`
	ResourceConflicts []ResourceConflict `json:"resource_conflicts"`

	// Analysis Metadata
	TotalConflicts    int            `json:"total_conflicts"`
	SeverityBreakdown map[string]int `json:"severity_breakdown"`
	ResolutionSummary []string       `json:"resolution_summary"`

	// Timestamps
	AnalyzedAt     time.Time `json:"analyzed_at"`
	SystemSnapshot string    `json:"system_snapshot"`
}

// PackageConflict represents conflicts between packages
type PackageConflict struct {
	ConflictType   string            `json:"conflict_type"` // version, dependency, exclusion
	Package1       string            `json:"package1"`
	Package2       string            `json:"package2"`
	Description    string            `json:"description"`
	Severity       string            `json:"severity"` // critical, high, medium, low
	Impact         []string          `json:"impact"`
	Resolution     []string          `json:"resolution"`
	AutoResolvable bool              `json:"auto_resolvable"`
	Context        map[string]string `json:"context"`
}

// ServiceConflict represents conflicts between system services
type ServiceConflict struct {
	ConflictType      string            `json:"conflict_type"` // port, resource, dependency
	Service1          string            `json:"service1"`
	Service2          string            `json:"service2"`
	ConflictReason    string            `json:"conflict_reason"`
	Severity          string            `json:"severity"`
	AffectedPorts     []int             `json:"affected_ports"`
	AffectedResources []string          `json:"affected_resources"`
	Resolution        []string          `json:"resolution"`
	RequiresRestart   bool              `json:"requires_restart"`
	Context           map[string]string `json:"context"`
}

// ConfigConflict represents conflicts in configuration settings
type ConfigConflict struct {
	ConflictType      string   `json:"conflict_type"` // duplicate, incompatible, deprecated
	ModuleName        string   `json:"module_name"`
	Option            string   `json:"option"`
	ConflictingValues []string `json:"conflicting_values"`
	Description       string   `json:"description"`
	Severity          string   `json:"severity"`
	Resolution        []string `json:"resolution"`
	ConfigPath        string   `json:"config_path"`
	LineNumbers       []int    `json:"line_numbers"`
}

// PortConflict represents port binding conflicts
type PortConflict struct {
	Port                int               `json:"port"`
	Protocol            string            `json:"protocol"` // tcp, udp
	ConflictingServices []string          `json:"conflicting_services"`
	Description         string            `json:"description"`
	Severity            string            `json:"severity"`
	Resolution          []string          `json:"resolution"`
	Context             map[string]string `json:"context"`
}

// ResourceConflict represents system resource conflicts
type ResourceConflict struct {
	ResourceType         string            `json:"resource_type"` // file, directory, device, memory
	ResourcePath         string            `json:"resource_path"`
	ConflictingProcesses []string          `json:"conflicting_processes"`
	Description          string            `json:"description"`
	Severity             string            `json:"severity"`
	Impact               []string          `json:"impact"`
	Resolution           []string          `json:"resolution"`
	Context              map[string]string `json:"context"`
}

// NewConflictDetector creates a new conflict detection system
func NewConflictDetector(log *logger.Logger, analyzer *SystemAnalyzer) *ConflictDetector {
	return &ConflictDetector{
		logger:   log,
		analyzer: analyzer,
		cache:    make(map[string]*ConflictAnalysis),
	}
}

// DetectConflicts performs comprehensive conflict detection
func (cd *ConflictDetector) DetectConflicts(ctx context.Context, userConfig *config.UserConfig) (*ConflictAnalysis, error) {
	cd.logger.Info("Starting conflict detection analysis")
	startTime := time.Now()

	// Get system analysis
	systemAnalysis, err := cd.analyzer.AnalyzeSystem(ctx, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get system analysis: %w", err)
	}

	// Check cache
	cacheKey := cd.generateCacheKey(systemAnalysis)
	if cached := cd.getCachedAnalysis(cacheKey); cached != nil {
		cd.logger.Info("Returning cached conflict analysis")
		return cached, nil
	}

	analysis := &ConflictAnalysis{
		AnalyzedAt:        startTime,
		SystemSnapshot:    fmt.Sprintf("%s_%s", systemAnalysis.SystemType, systemAnalysis.Hostname),
		SeverityBreakdown: make(map[string]int),
	}

	// Run conflict detection in parallel
	var wg sync.WaitGroup
	conflictChannels := make([]chan error, 5)

	// Package conflicts
	wg.Add(1)
	conflictChannels[0] = make(chan error, 1)
	go func() {
		defer wg.Done()
		conflicts, err := cd.detectPackageConflicts(systemAnalysis)
		if err != nil {
			conflictChannels[0] <- err
			return
		}
		analysis.PackageConflicts = conflicts
		conflictChannels[0] <- nil
	}()

	// Service conflicts
	wg.Add(1)
	conflictChannels[1] = make(chan error, 1)
	go func() {
		defer wg.Done()
		conflicts, err := cd.detectServiceConflicts(systemAnalysis)
		if err != nil {
			conflictChannels[1] <- err
			return
		}
		analysis.ServiceConflicts = conflicts
		conflictChannels[1] <- nil
	}()

	// Configuration conflicts
	wg.Add(1)
	conflictChannels[2] = make(chan error, 1)
	go func() {
		defer wg.Done()
		conflicts, err := cd.detectConfigConflicts(systemAnalysis)
		if err != nil {
			conflictChannels[2] <- err
			return
		}
		analysis.ConfigConflicts = conflicts
		conflictChannels[2] <- nil
	}()

	// Port conflicts
	wg.Add(1)
	conflictChannels[3] = make(chan error, 1)
	go func() {
		defer wg.Done()
		conflicts, err := cd.detectPortConflicts(systemAnalysis)
		if err != nil {
			conflictChannels[3] <- err
			return
		}
		analysis.PortConflicts = conflicts
		conflictChannels[3] <- nil
	}()

	// Resource conflicts
	wg.Add(1)
	conflictChannels[4] = make(chan error, 1)
	go func() {
		defer wg.Done()
		conflicts, err := cd.detectResourceConflicts(systemAnalysis)
		if err != nil {
			conflictChannels[4] <- err
			return
		}
		analysis.ResourceConflicts = conflicts
		conflictChannels[4] <- nil
	}()

	// Wait for all detections to complete
	wg.Wait()

	// Check for errors
	for i, ch := range conflictChannels {
		if err := <-ch; err != nil {
			cd.logger.Warn(fmt.Sprintf("Conflict detection warning in channel %d: %v", i, err))
		}
	}

	// Calculate totals and summary
	cd.calculateConflictSummary(analysis)

	// Cache the result
	cd.cacheAnalysis(cacheKey, analysis)

	cd.logger.Info(fmt.Sprintf("Conflict detection completed in %v (found %d conflicts)",
		time.Since(startTime), analysis.TotalConflicts))

	return analysis, nil
}

// detectPackageConflicts detects conflicts between installed packages
func (cd *ConflictDetector) detectPackageConflicts(analysis *SystemAnalysis) ([]PackageConflict, error) {
	var conflicts []PackageConflict

	packages := analysis.InstalledPackages

	// Check for known conflicting packages
	conflictingPairs := map[string][]string{
		"apache2":    {"nginx", "lighttpd"},
		"nginx":      {"apache2", "lighttpd"},
		"mysql":      {"postgresql", "mariadb"},
		"postgresql": {"mysql"},
		"vim":        {"emacs"},
		"systemd":    {"openrc", "runit"},
	}

	installedMap := make(map[string]PackageInfo)
	for _, pkg := range packages {
		key := strings.ToLower(pkg.Name)
		installedMap[key] = pkg
	}

	for pkg, conflictsWith := range conflictingPairs {
		if basePkg, hasBase := installedMap[pkg]; hasBase {
			for _, conflictPkg := range conflictsWith {
				if conflictingPkg, hasConflict := installedMap[conflictPkg]; hasConflict {
					conflicts = append(conflicts, PackageConflict{
						ConflictType:   "exclusion",
						Package1:       basePkg.Name,
						Package2:       conflictingPkg.Name,
						Description:    fmt.Sprintf("%s and %s serve similar purposes and may conflict", basePkg.Name, conflictingPkg.Name),
						Severity:       "medium",
						Impact:         []string{"Resource conflicts", "Configuration conflicts", "Service conflicts"},
						Resolution:     []string{fmt.Sprintf("Choose either %s or %s", basePkg.Name, conflictingPkg.Name), "Disable conflicting services"},
						AutoResolvable: false,
						Context:        map[string]string{"type": "service_exclusion"},
					})
				}
			}
		}
	}

	// Check for version conflicts
	conflicts = append(conflicts, cd.detectVersionConflicts(packages)...)

	// Check for dependency conflicts
	conflicts = append(conflicts, cd.detectDependencyConflicts(packages)...)

	return conflicts, nil
}

// detectServiceConflicts detects conflicts between system services
func (cd *ConflictDetector) detectServiceConflicts(analysis *SystemAnalysis) ([]ServiceConflict, error) {
	var conflicts []ServiceConflict

	services := analysis.EnabledServices

	// Check for port conflicts
	portUsage := make(map[int][]ServiceInfo)
	for _, service := range services {
		if service.Port > 0 {
			portUsage[service.Port] = append(portUsage[service.Port], service)
		}
	}

	for port, servicesOnPort := range portUsage {
		if len(servicesOnPort) > 1 {
			for i := 0; i < len(servicesOnPort); i++ {
				for j := i + 1; j < len(servicesOnPort); j++ {
					conflicts = append(conflicts, ServiceConflict{
						ConflictType:    "port",
						Service1:        servicesOnPort[i].Name,
						Service2:        servicesOnPort[j].Name,
						ConflictReason:  fmt.Sprintf("Both services trying to bind to port %d", port),
						Severity:        "high",
						AffectedPorts:   []int{port},
						Resolution:      []string{"Configure different ports", "Disable one service", "Use reverse proxy"},
						RequiresRestart: true,
						Context:         map[string]string{"port": fmt.Sprintf("%d", port)},
					})
				}
			}
		}
	}

	// Check for dependency conflicts
	conflicts = append(conflicts, cd.detectServiceDependencyConflicts(services)...)

	return conflicts, nil
}

// detectConfigConflicts detects conflicts in configuration files
func (cd *ConflictDetector) detectConfigConflicts(analysis *SystemAnalysis) ([]ConfigConflict, error) {
	var conflicts []ConfigConflict

	// Check for duplicate configuration options
	// This would require parsing actual config files, simplified for now
	configModules := analysis.ConfigModules

	// Look for common conflicting configurations
	conflictingConfigs := map[string][]string{
		"networking.firewall.enable": {"networking.firewall.allowedTCPPorts", "networking.firewall.allowedUDPPorts"},
		"services.xserver.enable":    {"services.wayland.enable"},
		"boot.loader.grub.enable":    {"boot.loader.systemd-boot.enable"},
	}

	for baseConfig, conflicts_with := range conflictingConfigs {
		hasBase := false
		for _, module := range configModules {
			for _, option := range module.Options {
				if strings.Contains(option, baseConfig) {
					hasBase = true
					break
				}
			}
			if hasBase {
				break
			}
		}

		if hasBase {
			for _, conflictConfig := range conflicts_with {
				for _, module := range configModules {
					for _, option := range module.Options {
						if strings.Contains(option, conflictConfig) {
							conflicts = append(conflicts, ConfigConflict{
								ConflictType:      "incompatible",
								ModuleName:        module.Name,
								Option:            baseConfig,
								ConflictingValues: []string{baseConfig, conflictConfig},
								Description:       fmt.Sprintf("Configuration %s conflicts with %s", baseConfig, conflictConfig),
								Severity:          "medium",
								Resolution:        []string{"Review configuration logic", "Use conditional configuration"},
							})
						}
					}
				}
			}
		}
	}

	return conflicts, nil
}

// detectPortConflicts detects port binding conflicts
func (cd *ConflictDetector) detectPortConflicts(analysis *SystemAnalysis) ([]PortConflict, error) {
	var conflicts []PortConflict

	// Build port usage map
	portUsage := make(map[int][]string)

	for _, service := range analysis.EnabledServices {
		if service.Port > 0 {
			portUsage[service.Port] = append(portUsage[service.Port], service.Name)
		}
	}

	// Check for conflicts
	for port, services := range portUsage {
		if len(services) > 1 {
			conflicts = append(conflicts, PortConflict{
				Port:                port,
				Protocol:            "tcp", // Assume TCP for now
				ConflictingServices: services,
				Description:         fmt.Sprintf("Port %d is used by multiple services: %s", port, strings.Join(services, ", ")),
				Severity:            "high",
				Resolution: []string{
					"Configure different ports for conflicting services",
					"Use a reverse proxy to share the port",
					"Disable unnecessary services",
				},
				Context: map[string]string{
					"port":     fmt.Sprintf("%d", port),
					"services": strings.Join(services, ","),
				},
			})
		}
	}

	return conflicts, nil
}

// detectResourceConflicts detects system resource conflicts
func (cd *ConflictDetector) detectResourceConflicts(analysis *SystemAnalysis) ([]ResourceConflict, error) {
	var conflicts []ResourceConflict

	// Check for high resource usage that might indicate conflicts
	if analysis.PerformanceMetrics.MemoryUsage > 90.0 {
		conflicts = append(conflicts, ResourceConflict{
			ResourceType:         "memory",
			ResourcePath:         "/proc/meminfo",
			ConflictingProcesses: []string{"unknown"}, // Would need process analysis
			Description:          fmt.Sprintf("Memory usage is critically high at %.1f%%", analysis.PerformanceMetrics.MemoryUsage),
			Severity:             "critical",
			Impact:               []string{"System instability", "Poor performance", "Potential crashes"},
			Resolution: []string{
				"Identify memory-hungry processes",
				"Increase swap space",
				"Add more RAM",
				"Optimize application configurations",
			},
			Context: map[string]string{
				"usage":     fmt.Sprintf("%.1f", analysis.PerformanceMetrics.MemoryUsage),
				"threshold": "90.0",
			},
		})
	}

	// Check for disk space conflicts
	for mount, usage := range analysis.PerformanceMetrics.DiskUsage {
		if usage > 95.0 {
			conflicts = append(conflicts, ResourceConflict{
				ResourceType:         "disk",
				ResourcePath:         mount,
				ConflictingProcesses: []string{"filesystem"},
				Description:          fmt.Sprintf("Disk usage on %s is critically high at %.1f%%", mount, usage),
				Severity:             "critical",
				Impact:               []string{"System instability", "Application failures", "Log rotation issues"},
				Resolution: []string{
					"Clean up unnecessary files",
					"Run nix-collect-garbage",
					"Move files to external storage",
					"Increase disk space",
				},
				Context: map[string]string{
					"mount": mount,
					"usage": fmt.Sprintf("%.1f", usage),
				},
			})
		}
	}

	return conflicts, nil
}

// Helper methods for conflict detection

func (cd *ConflictDetector) detectVersionConflicts(packages []PackageInfo) []PackageConflict {
	var conflicts []PackageConflict

	// Group packages by base name (without version)
	packageGroups := make(map[string][]PackageInfo)
	for _, pkg := range packages {
		baseName := cd.extractBaseName(pkg.Name)
		packageGroups[baseName] = append(packageGroups[baseName], pkg)
	}

	// Check for multiple versions of the same package
	for baseName, pkgs := range packageGroups {
		if len(pkgs) > 1 {
			for i := 0; i < len(pkgs); i++ {
				for j := i + 1; j < len(pkgs); j++ {
					conflicts = append(conflicts, PackageConflict{
						ConflictType:   "version",
						Package1:       pkgs[i].Name,
						Package2:       pkgs[j].Name,
						Description:    fmt.Sprintf("Multiple versions of %s detected", baseName),
						Severity:       "low",
						Impact:         []string{"Potential version conflicts", "Disk space usage"},
						Resolution:     []string{"Remove older versions", "Use nix-env --list-generations"},
						AutoResolvable: true,
					})
				}
			}
		}
	}

	return conflicts
}

func (cd *ConflictDetector) detectDependencyConflicts(packages []PackageInfo) []PackageConflict {
	var conflicts []PackageConflict

	// This would require deeper dependency analysis
	// For now, just detect some common known conflicts

	return conflicts
}

func (cd *ConflictDetector) detectServiceDependencyConflicts(services []ServiceInfo) []ServiceConflict {
	var conflicts []ServiceConflict

	// Check for circular dependencies or missing dependencies
	for _, service := range services {
		for _, dep := range service.Dependencies {
			// Check if dependency is available
			found := false
			for _, otherService := range services {
				if otherService.Name == dep {
					found = true
					break
				}
			}

			if !found {
				conflicts = append(conflicts, ServiceConflict{
					ConflictType:    "dependency",
					Service1:        service.Name,
					Service2:        dep,
					ConflictReason:  fmt.Sprintf("Service %s depends on %s which is not available", service.Name, dep),
					Severity:        "medium",
					Resolution:      []string{fmt.Sprintf("Enable service %s", dep), "Remove dependency requirement"},
					RequiresRestart: true,
				})
			}
		}
	}

	return conflicts
}

func (cd *ConflictDetector) extractBaseName(packageName string) string {
	// Extract base name by removing version numbers and suffixes
	parts := strings.Split(packageName, "-")
	if len(parts) > 1 {
		// Remove version-like suffixes
		for i := len(parts) - 1; i > 0; i-- {
			if strings.ContainsAny(parts[i], "0123456789.") {
				return strings.Join(parts[:i], "-")
			}
		}
	}
	return packageName
}

func (cd *ConflictDetector) calculateConflictSummary(analysis *ConflictAnalysis) {
	// Count total conflicts
	analysis.TotalConflicts = len(analysis.PackageConflicts) +
		len(analysis.ServiceConflicts) +
		len(analysis.ConfigConflicts) +
		len(analysis.PortConflicts) +
		len(analysis.ResourceConflicts)

	// Count by severity
	analysis.SeverityBreakdown = make(map[string]int)

	for _, conflict := range analysis.PackageConflicts {
		analysis.SeverityBreakdown[conflict.Severity]++
	}
	for _, conflict := range analysis.ServiceConflicts {
		analysis.SeverityBreakdown[conflict.Severity]++
	}
	for _, conflict := range analysis.ConfigConflicts {
		analysis.SeverityBreakdown[conflict.Severity]++
	}
	for _, conflict := range analysis.PortConflicts {
		analysis.SeverityBreakdown[conflict.Severity]++
	}
	for _, conflict := range analysis.ResourceConflicts {
		analysis.SeverityBreakdown[conflict.Severity]++
	}

	// Generate resolution summary
	analysis.ResolutionSummary = []string{
		fmt.Sprintf("Found %d total conflicts", analysis.TotalConflicts),
		fmt.Sprintf("Critical: %d, High: %d, Medium: %d, Low: %d",
			analysis.SeverityBreakdown["critical"],
			analysis.SeverityBreakdown["high"],
			analysis.SeverityBreakdown["medium"],
			analysis.SeverityBreakdown["low"]),
		"Review each conflict and apply recommended resolutions",
	}
}

func (cd *ConflictDetector) generateCacheKey(analysis *SystemAnalysis) string {
	return fmt.Sprintf("conflicts_%s_%s", analysis.SystemType, analysis.Hostname)
}

func (cd *ConflictDetector) getCachedAnalysis(key string) *ConflictAnalysis {
	cd.mu.RLock()
	defer cd.mu.RUnlock()

	return cd.cache[key]
}

func (cd *ConflictDetector) cacheAnalysis(key string, analysis *ConflictAnalysis) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.cache[key] = analysis
}

// ClearCache clears the conflict detection cache
func (cd *ConflictDetector) ClearCache() {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	cd.cache = make(map[string]*ConflictAnalysis)
	cd.logger.Info("Conflict detection cache cleared")
}
