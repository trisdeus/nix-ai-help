package codegen

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// Optimizer handles NixOS configuration optimization
type Optimizer struct {
	logger logger.Logger
	rules  []OptimizationRule
}

// OptimizationRule represents a single optimization rule
type OptimizationRule struct {
	Name        string
	Description string
	Apply       func(config string, context *config.NixOSContext) OptimizationResult
	Category    string // "performance", "security", "maintenance", "compatibility"
}

// OptimizationResult represents the result of applying an optimization
type OptimizationResult struct {
	Applied     bool     `json:"applied"`
	Config      string   `json:"config"`
	Changes     []string `json:"changes"`
	Explanation string   `json:"explanation"`
	Impact      string   `json:"impact"` // "low", "medium", "high"
}

// OptimizationRequest represents an optimization request
type OptimizationRequest struct {
	Config     string               `json:"config"`
	Context    *config.NixOSContext `json:"context,omitempty"`
	Categories []string             `json:"categories,omitempty"` // Filter by categories
	Aggressive bool                 `json:"aggressive,omitempty"` // Apply aggressive optimizations
}

// OptimizationResponse represents the result of configuration optimization
type OptimizationResponse struct {
	OptimizedConfig string                   `json:"optimized_config"`
	Applied         []OptimizationResult     `json:"applied"`
	Suggestions     []OptimizationSuggestion `json:"suggestions"`
	Summary         OptimizationSummary      `json:"summary"`
}

// OptimizationSuggestion represents a suggested optimization that wasn't auto-applied
type OptimizationSuggestion struct {
	Rule        string `json:"rule"`
	Description string `json:"description"`
	Benefit     string `json:"benefit"`
	Risk        string `json:"risk"`
	Manual      bool   `json:"manual"` // Requires manual intervention
}

// OptimizationSummary provides a summary of optimizations
type OptimizationSummary struct {
	TotalRules        int    `json:"total_rules"`
	Applied           int    `json:"applied"`
	Suggested         int    `json:"suggested"`
	PerformanceImpact string `json:"performance_impact"`
	SecurityImpact    string `json:"security_impact"`
}

// NewOptimizer creates a new configuration optimizer
func NewOptimizer(logger logger.Logger) *Optimizer {
	o := &Optimizer{
		logger: logger,
	}
	o.initializeRules()
	return o
}

// Optimize optimizes the given NixOS configuration
func (o *Optimizer) Optimize(request *OptimizationRequest) *OptimizationResponse {
	o.logger.Info(fmt.Sprintf("Optimizing NixOS configuration - categories: %v, aggressive: %v", request.Categories, request.Aggressive))

	response := &OptimizationResponse{
		OptimizedConfig: request.Config,
		Applied:         []OptimizationResult{},
		Suggestions:     []OptimizationSuggestion{},
	}

	// Apply optimization rules
	for _, rule := range o.rules {
		// Skip rules not in requested categories (if specified)
		if len(request.Categories) > 0 && !contains(request.Categories, rule.Category) {
			continue
		}

		result := rule.Apply(response.OptimizedConfig, request.Context)
		if result.Applied {
			response.OptimizedConfig = result.Config
			response.Applied = append(response.Applied, result)
			o.logger.Info(fmt.Sprintf("Applied optimization - rule: %s, impact: %s", rule.Name, result.Impact))
		} else {
			// Add as suggestion if not applied
			response.Suggestions = append(response.Suggestions, OptimizationSuggestion{
				Rule:        rule.Name,
				Description: rule.Description,
				Benefit:     result.Explanation,
				Manual:      true,
			})
		}
	}

	// Generate summary
	response.Summary = o.generateSummary(response, request)

	o.logger.Info(fmt.Sprintf("Optimization completed - applied: %d, suggestions: %d", len(response.Applied), len(response.Suggestions)))

	return response
}

// initializeRules sets up optimization rules
func (o *Optimizer) initializeRules() {
	o.rules = []OptimizationRule{
		{
			Name:        "enable_flakes",
			Description: "Enable experimental flakes feature",
			Apply:       o.enableFlakes,
			Category:    "compatibility",
		},
		{
			Name:        "optimize_garbage_collection",
			Description: "Configure automatic garbage collection",
			Apply:       o.optimizeGarbageCollection,
			Category:    "maintenance",
		},
		{
			Name:        "enable_auto_optimise",
			Description: "Enable automatic store optimization",
			Apply:       o.enableAutoOptimise,
			Category:    "performance",
		},
		{
			Name:        "optimize_systemd",
			Description: "Optimize systemd configuration",
			Apply:       o.optimizeSystemd,
			Category:    "performance",
		},
		{
			Name:        "security_hardening",
			Description: "Apply basic security hardening",
			Apply:       o.applySecurityHardening,
			Category:    "security",
		},
		{
			Name:        "optimize_networking",
			Description: "Optimize networking configuration",
			Apply:       o.optimizeNetworking,
			Category:    "performance",
		},
		{
			Name:        "filesystem_optimization",
			Description: "Optimize filesystem settings",
			Apply:       o.optimizeFilesystem,
			Category:    "performance",
		},
		{
			Name:        "kernel_optimization",
			Description: "Optimize kernel parameters",
			Apply:       o.optimizeKernel,
			Category:    "performance",
		},
	}
}

// enableFlakes enables the experimental flakes feature
func (o *Optimizer) enableFlakes(config string, context *config.NixOSContext) OptimizationResult {
	if context != nil && context.UsesFlakes {
		return OptimizationResult{Applied: false, Config: config}
	}

	if strings.Contains(config, "nix.settings.experimental-features") {
		return OptimizationResult{Applied: false, Config: config}
	}

	// Add flakes configuration
	flakesConfig := `
  # Enable experimental features
  nix.settings.experimental-features = [ "nix-command" "flakes" ];`

	// Find a good place to insert
	lines := strings.Split(config, "\n")
	var result []string
	inserted := false

	for i, line := range lines {
		result = append(result, line)
		// Insert after system configuration but before closing brace
		if !inserted && (strings.Contains(line, "system.stateVersion") ||
			(i == len(lines)-2 && strings.TrimSpace(line) != "}")) {
			result = append(result, flakesConfig)
			inserted = true
		}
	}

	if !inserted {
		// Insert before final closing brace
		if len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "}" {
			result = append(result[:len(result)-1], flakesConfig, result[len(result)-1])
		} else {
			result = append(result, flakesConfig)
		}
	}

	return OptimizationResult{
		Applied:     true,
		Config:      strings.Join(result, "\n"),
		Changes:     []string{"Added experimental flakes support"},
		Explanation: "Enables modern Nix flakes for better dependency management",
		Impact:      "medium",
	}
}

// optimizeGarbageCollection adds automatic garbage collection
func (o *Optimizer) optimizeGarbageCollection(config string, context *config.NixOSContext) OptimizationResult {
	if strings.Contains(config, "nix.gc.automatic") {
		return OptimizationResult{Applied: false, Config: config}
	}

	gcConfig := `
  # Automatic garbage collection
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 30d";
  };`

	result := o.insertConfiguration(config, gcConfig, "nix configuration")

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     []string{"Added automatic garbage collection"},
		Explanation: "Automatically cleans up old generations to save disk space",
		Impact:      "medium",
	}
}

// enableAutoOptimise enables automatic store optimization
func (o *Optimizer) enableAutoOptimise(config string, context *config.NixOSContext) OptimizationResult {
	if strings.Contains(config, "nix.settings.auto-optimise-store") {
		return OptimizationResult{Applied: false, Config: config}
	}

	optimiseConfig := `
  # Automatic store optimization
  nix.settings.auto-optimise-store = true;`

	result := o.insertConfiguration(config, optimiseConfig, "nix configuration")

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     []string{"Enabled automatic store optimization"},
		Explanation: "Automatically deduplicates files in the Nix store to save space",
		Impact:      "low",
	}
}

// optimizeSystemd optimizes systemd configuration
func (o *Optimizer) optimizeSystemd(config string, context *config.NixOSContext) OptimizationResult {
	if strings.Contains(config, "systemd.extraConfig") {
		return OptimizationResult{Applied: false, Config: config}
	}

	systemdConfig := `
  # Systemd optimizations
  systemd.extraConfig = ''
    DefaultTimeoutStopSec=10s
    DefaultLimitNOFILE=1048576
  '';`

	result := o.insertConfiguration(config, systemdConfig, "system configuration")

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     []string{"Optimized systemd timeout and file limits"},
		Explanation: "Improves system responsiveness and resource handling",
		Impact:      "medium",
	}
}

// applySecurityHardening applies basic security hardening
func (o *Optimizer) applySecurityHardening(config string, context *config.NixOSContext) OptimizationResult {
	changes := []string{}
	result := config

	// Enable firewall if not configured
	if !strings.Contains(config, "networking.firewall") {
		firewallConfig := `
  # Basic firewall configuration
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ ];
    allowedUDPPorts = [ ];
  };`
		result = o.insertConfiguration(result, firewallConfig, "networking configuration")
		changes = append(changes, "Enabled firewall")
	}

	// Disable unnecessary services
	if !strings.Contains(config, "services.printing.enable = false") && !strings.Contains(config, "services.printing.enable = true") {
		result = o.insertConfiguration(result, "\n  # Disable printing if not needed\n  # services.printing.enable = false;", "services configuration")
		changes = append(changes, "Added suggestion to disable unused printing service")
	}

	if len(changes) == 0 {
		return OptimizationResult{Applied: false, Config: config}
	}

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     changes,
		Explanation: "Applies basic security hardening measures",
		Impact:      "high",
	}
}

// optimizeNetworking optimizes networking configuration
func (o *Optimizer) optimizeNetworking(config string, context *config.NixOSContext) OptimizationResult {
	if strings.Contains(config, "networking.useDHCP = false") {
		return OptimizationResult{Applied: false, Config: config}
	}

	// Only apply if this looks like a server configuration
	if !strings.Contains(config, "services.nginx") && !strings.Contains(config, "services.apache") {
		return OptimizationResult{Applied: false, Config: config}
	}

	networkConfig := `
  # Networking optimizations for server
  networking.useDHCP = false;
  networking.interfaces.eth0.useDHCP = true;
  
  # TCP optimizations
  boot.kernel.sysctl = {
    "net.ipv4.tcp_keepalive_time" = 120;
    "net.ipv4.tcp_keepalive_intvl" = 30;
    "net.ipv4.tcp_keepalive_probes" = 3;
  };`

	result := o.insertConfiguration(config, networkConfig, "networking configuration")

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     []string{"Optimized networking and TCP settings"},
		Explanation: "Improves network performance for server workloads",
		Impact:      "medium",
	}
}

// optimizeFilesystem optimizes filesystem settings
func (o *Optimizer) optimizeFilesystem(config string, context *config.NixOSContext) OptimizationResult {
	if strings.Contains(config, "fileSystems") && strings.Contains(config, "noatime") {
		return OptimizationResult{Applied: false, Config: config}
	}

	// Add SSD optimizations as comment/suggestion
	fsConfig := `
  # Filesystem optimizations (uncomment for SSD)
  # fileSystems."/".options = [ "noatime" "nodiratime" ];
  # services.fstrim.enable = true;`

	result := o.insertConfiguration(config, fsConfig, "filesystem configuration")

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     []string{"Added filesystem optimization suggestions"},
		Explanation: "Provides SSD optimization options to improve performance and longevity",
		Impact:      "low",
	}
}

// optimizeKernel optimizes kernel parameters
func (o *Optimizer) optimizeKernel(config string, context *config.NixOSContext) OptimizationResult {
	if strings.Contains(config, "boot.kernel.sysctl") {
		return OptimizationResult{Applied: false, Config: config}
	}

	kernelConfig := `
  # Kernel optimizations
  boot.kernel.sysctl = {
    # Improve virtual memory handling
    "vm.swappiness" = 10;
    "vm.vfs_cache_pressure" = 50;
    
    # Network optimizations
    "net.core.rmem_max" = 16777216;
    "net.core.wmem_max" = 16777216;
  };`

	result := o.insertConfiguration(config, kernelConfig, "kernel configuration")

	return OptimizationResult{
		Applied:     true,
		Config:      result,
		Changes:     []string{"Added kernel parameter optimizations"},
		Explanation: "Optimizes memory management and network performance",
		Impact:      "medium",
	}
}

// insertConfiguration inserts configuration at an appropriate location
func (o *Optimizer) insertConfiguration(config, newConfig, category string) string {
	lines := strings.Split(config, "\n")
	var result []string
	inserted := false

	for i, line := range lines {
		result = append(result, line)

		// Try to insert in logical sections
		if !inserted {
			// Insert after imports or at a logical break
			if (strings.Contains(line, "imports") && strings.Contains(line, "];")) ||
				(strings.Contains(line, "system.stateVersion") && i < len(lines)-2) ||
				(i == len(lines)-2 && strings.TrimSpace(line) != "}") {
				result = append(result, newConfig)
				inserted = true
			}
		}
	}

	if !inserted {
		// Insert before final closing brace
		if len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "}" {
			result = append(result[:len(result)-1], newConfig, result[len(result)-1])
		} else {
			result = append(result, newConfig)
		}
	}

	return strings.Join(result, "\n")
}

// generateSummary generates an optimization summary
func (o *Optimizer) generateSummary(response *OptimizationResponse, request *OptimizationRequest) OptimizationSummary {
	summary := OptimizationSummary{
		TotalRules: len(o.rules),
		Applied:    len(response.Applied),
		Suggested:  len(response.Suggestions),
	}

	// Analyze impact
	performanceCount := 0
	securityCount := 0

	for _ = range response.Applied {
		for _, rule := range o.rules {
			if rule.Category == "performance" {
				performanceCount++
			}
			if rule.Category == "security" {
				securityCount++
			}
		}
	}

	if performanceCount > 2 {
		summary.PerformanceImpact = "high"
	} else if performanceCount > 0 {
		summary.PerformanceImpact = "medium"
	} else {
		summary.PerformanceImpact = "low"
	}

	if securityCount > 1 {
		summary.SecurityImpact = "high"
	} else if securityCount > 0 {
		summary.SecurityImpact = "medium"
	} else {
		summary.SecurityImpact = "low"
	}

	return summary
}

// OptimizeForEnvironment applies environment-specific optimizations
func (o *Optimizer) OptimizeForEnvironment(config string, environment string, context *config.NixOSContext) *OptimizationResponse {
	request := &OptimizationRequest{
		Config:  config,
		Context: context,
	}

	switch environment {
	case "server":
		request.Categories = []string{"performance", "security", "maintenance"}
		request.Aggressive = true
	case "desktop":
		request.Categories = []string{"performance", "compatibility"}
	case "laptop":
		request.Categories = []string{"performance"} // Focus on battery life
	case "development":
		request.Categories = []string{"compatibility", "performance"}
	default:
		request.Categories = []string{"maintenance", "compatibility"}
	}

	return o.Optimize(request)
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
