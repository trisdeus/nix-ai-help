// Predictive suggestions engine for nixai
package intelligence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// Predictor provides AI-powered predictive suggestions and recommendations
type Predictor struct {
	logger          *logger.Logger
	aiProvider      ai.Provider
	analyzer        *SystemAnalyzer
	userPatterns    *UserPatternAnalyzer
	suggestionCache map[string]*PredictionResult
	mu              sync.RWMutex
	cacheExpiry     time.Duration
}

// PredictionResult represents the result of predictive analysis
type PredictionResult struct {
	// Core Predictions
	PackageSuggestions     []PackageSuggestion     `json:"package_suggestions"`
	ConfigSuggestions      []ConfigSuggestion      `json:"config_suggestions"`
	SecuritySuggestions    []SecuritySuggestion    `json:"security_suggestions"`
	PerformanceSuggestions []PerformanceSuggestion `json:"performance_suggestions"`
	MaintenanceSuggestions []MaintenanceSuggestion `json:"maintenance_suggestions"`

	// Context Information
	BasedOnPatterns []string        `json:"based_on_patterns"`
	Confidence      float64         `json:"confidence"`
	ReasoningChain  []ReasoningStep `json:"reasoning_chain"`

	// Metadata
	GeneratedAt   time.Time `json:"generated_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	ModelUsed     string    `json:"model_used"`
	SystemContext string    `json:"system_context"`
}

// PackageSuggestion represents a suggested package installation
type PackageSuggestion struct {
	PackageName         string   `json:"package_name"`
	Reason              string   `json:"reason"`
	Category            string   `json:"category"`
	Priority            string   `json:"priority"` // high, medium, low
	Confidence          float64  `json:"confidence"`
	Dependencies        []string `json:"dependencies"`
	Conflicts           []string `json:"conflicts"`
	InstallCommand      string   `json:"install_command"`
	Documentation       string   `json:"documentation"`
	Benefits            []string `json:"benefits"`
	Risks               []string `json:"risks"`
	AlternativePackages []string `json:"alternative_packages"`
}

// ConfigSuggestion represents a suggested configuration change
type ConfigSuggestion struct {
	ModuleName     string      `json:"module_name"`
	Option         string      `json:"option"`
	SuggestedValue interface{} `json:"suggested_value"`
	CurrentValue   interface{} `json:"current_value"`
	Reason         string      `json:"reason"`
	Impact         string      `json:"impact"`
	Priority       string      `json:"priority"`
	Confidence     float64     `json:"confidence"`
	ConfigPath     string      `json:"config_path"`
	Example        string      `json:"example"`
	Documentation  string      `json:"documentation"`
	Prerequisites  []string    `json:"prerequisites"`
	SideEffects    []string    `json:"side_effects"`
}

// SecuritySuggestion represents a security improvement suggestion
type SecuritySuggestion struct {
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Severity          string   `json:"severity"` // critical, high, medium, low
	Category          string   `json:"category"` // firewall, encryption, access, etc.
	CurrentStatus     string   `json:"current_status"`
	RecommendedAction string   `json:"recommended_action"`
	Implementation    []string `json:"implementation"`
	Verification      string   `json:"verification"`
	References        []string `json:"references"`
	Confidence        float64  `json:"confidence"`
	Risk              string   `json:"risk"`
	Effort            string   `json:"effort"`
}

// PerformanceSuggestion represents a performance optimization suggestion
type PerformanceSuggestion struct {
	Component      string   `json:"component"` // cpu, memory, disk, network, boot
	Issue          string   `json:"issue"`
	Suggestion     string   `json:"suggestion"`
	ExpectedGain   string   `json:"expected_gain"`
	Implementation []string `json:"implementation"`
	Monitoring     string   `json:"monitoring"`
	Priority       string   `json:"priority"`
	Confidence     float64  `json:"confidence"`
	Complexity     string   `json:"complexity"`
	Reversible     bool     `json:"reversible"`
}

// MaintenanceSuggestion represents a system maintenance suggestion
type MaintenanceSuggestion struct {
	Task          string        `json:"task"`
	Description   string        `json:"description"`
	Frequency     string        `json:"frequency"` // daily, weekly, monthly, quarterly
	NextDue       time.Time     `json:"next_due"`
	LastPerformed time.Time     `json:"last_performed"`
	Commands      []string      `json:"commands"`
	Automation    string        `json:"automation"`
	Priority      string        `json:"priority"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Confidence    float64       `json:"confidence"`
}

// ReasoningStep represents a step in the AI reasoning process
type ReasoningStep struct {
	Step        int      `json:"step"`
	Description string   `json:"description"`
	Evidence    []string `json:"evidence"`
	Conclusion  string   `json:"conclusion"`
	Confidence  float64  `json:"confidence"`
}

// UserPatternAnalyzer analyzes user behavior patterns
type UserPatternAnalyzer struct {
	patterns       map[string]*UserPattern
	commandHistory []CommandHistory
	mu             sync.RWMutex
	dataFile       string
}

// UserPattern represents detected user behavior patterns
type UserPattern struct {
	PatternID  string            `json:"pattern_id"`
	Category   string            `json:"category"` // development, administration, desktop, etc.
	Commands   []string          `json:"commands"`
	Frequency  int               `json:"frequency"`
	LastSeen   time.Time         `json:"last_seen"`
	Context    map[string]string `json:"context"`
	Confidence float64           `json:"confidence"`
}

// CommandHistory represents command usage history
type CommandHistory struct {
	Command   string            `json:"command"`
	Timestamp time.Time         `json:"timestamp"`
	Context   map[string]string `json:"context"`
	Success   bool              `json:"success"`
	Duration  time.Duration     `json:"duration"`
}

// PredictionContext contains context for making predictions
type PredictionContext struct {
	SystemAnalysis *SystemAnalysis    `json:"system_analysis"`
	UserConfig     *config.UserConfig `json:"user_config"`
	RecentActivity []CommandHistory   `json:"recent_activity"`
	TimeOfDay      string             `json:"time_of_day"`
	DayOfWeek      string             `json:"day_of_week"`
	UserPatterns   []UserPattern      `json:"user_patterns"`
}

// NewPredictor creates a new predictive suggestions engine
func NewPredictor(log *logger.Logger, provider ai.Provider, analyzer *SystemAnalyzer) *Predictor {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".config", "nixai", "intelligence")
	os.MkdirAll(dataDir, 0755)

	patternAnalyzer := &UserPatternAnalyzer{
		patterns: make(map[string]*UserPattern),
		dataFile: filepath.Join(dataDir, "user_patterns.json"),
	}

	// Load existing patterns
	patternAnalyzer.loadPatterns()

	return &Predictor{
		logger:          log,
		aiProvider:      provider,
		analyzer:        analyzer,
		userPatterns:    patternAnalyzer,
		suggestionCache: make(map[string]*PredictionResult),
		cacheExpiry:     15 * time.Minute, // Cache predictions for 15 minutes
	}
}

// GeneratePredictions generates comprehensive predictions based on system state and user patterns
func (p *Predictor) GeneratePredictions(ctx context.Context, userConfig *config.UserConfig) (*PredictionResult, error) {
	startTime := time.Now()
	p.logger.Info("Generating predictive suggestions")

	// Build prediction context
	predictionCtx, err := p.buildPredictionContext(ctx, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build prediction context: %w", err)
	}

	// Check cache first
	cacheKey := p.generateCacheKey(predictionCtx)
	if cached := p.getCachedPrediction(cacheKey); cached != nil {
		p.logger.Info("Returning cached predictions")
		return cached, nil
	}

	result := &PredictionResult{
		GeneratedAt:   startTime,
		ExpiresAt:     startTime.Add(p.cacheExpiry),
		SystemContext: fmt.Sprintf("%s_%s", predictionCtx.SystemAnalysis.SystemType, predictionCtx.SystemAnalysis.Hostname),
		ModelUsed:     "AI-Predictor-v1.0",
	}

	// Generate different types of suggestions
	var wg sync.WaitGroup
	errors := make(chan error, 5)

	// Package suggestions
	wg.Add(1)
	go func() {
		defer wg.Done()
		suggestions, err := p.generatePackageSuggestions(ctx, predictionCtx)
		if err != nil {
			errors <- fmt.Errorf("package suggestions: %w", err)
			return
		}
		result.PackageSuggestions = suggestions
	}()

	// Configuration suggestions
	wg.Add(1)
	go func() {
		defer wg.Done()
		suggestions, err := p.generateConfigSuggestions(ctx, predictionCtx)
		if err != nil {
			errors <- fmt.Errorf("config suggestions: %w", err)
			return
		}
		result.ConfigSuggestions = suggestions
	}()

	// Security suggestions
	wg.Add(1)
	go func() {
		defer wg.Done()
		suggestions, err := p.generateSecuritySuggestions(ctx, predictionCtx)
		if err != nil {
			errors <- fmt.Errorf("security suggestions: %w", err)
			return
		}
		result.SecuritySuggestions = suggestions
	}()

	// Performance suggestions
	wg.Add(1)
	go func() {
		defer wg.Done()
		suggestions, err := p.generatePerformanceSuggestions(ctx, predictionCtx)
		if err != nil {
			errors <- fmt.Errorf("performance suggestions: %w", err)
			return
		}
		result.PerformanceSuggestions = suggestions
	}()

	// Maintenance suggestions
	wg.Add(1)
	go func() {
		defer wg.Done()
		suggestions, err := p.generateMaintenanceSuggestions(ctx, predictionCtx)
		if err != nil {
			errors <- fmt.Errorf("maintenance suggestions: %w", err)
			return
		}
		result.MaintenanceSuggestions = suggestions
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		p.logger.Warn(fmt.Sprintf("Prediction generation warning: %v", err))
	}

	// Generate reasoning chain and confidence
	result.ReasoningChain = p.generateReasoningChain(predictionCtx, result)
	result.Confidence = p.calculateOverallConfidence(result)
	result.BasedOnPatterns = p.extractPatternSummary(predictionCtx)

	// Cache the result
	p.cachePrediction(cacheKey, result)

	p.logger.Info(fmt.Sprintf("Generated predictions in %v (confidence: %.1f%%)",
		time.Since(startTime), result.Confidence*100))

	return result, nil
}

// buildPredictionContext builds the context needed for predictions
func (p *Predictor) buildPredictionContext(ctx context.Context, userConfig *config.UserConfig) (*PredictionContext, error) {
	// Get system analysis
	analysis, err := p.analyzer.AnalyzeSystem(ctx, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze system: %w", err)
	}

	// Get user patterns
	patterns := p.userPatterns.getRecentUserPatterns(10) // Last 10 patterns

	// Get recent activity
	recentActivity := p.userPatterns.getRecentCommandHistory(20) // Last 20 activities

	predictionCtx := &PredictionContext{
		SystemAnalysis: analysis,
		UserConfig:     userConfig,
		RecentActivity: recentActivity,
		TimeOfDay:      p.getTimeOfDay(),
		DayOfWeek:      time.Now().Weekday().String(),
		UserPatterns:   patterns,
	}

	return predictionCtx, nil
}

// generatePackageSuggestions generates package installation suggestions
func (p *Predictor) generatePackageSuggestions(ctx context.Context, predictionCtx *PredictionContext) ([]PackageSuggestion, error) {
	var suggestions []PackageSuggestion

	// Analyze system for missing common packages
	analysis := predictionCtx.SystemAnalysis

	// Development environment suggestions
	if p.userPatterns.hasPattern("development") {
		devSuggestions := p.generateDevelopmentPackageSuggestions(analysis)
		suggestions = append(suggestions, devSuggestions...)
	}

	// Desktop environment suggestions
	if analysis.EnabledServices != nil {
		desktopSuggestions := p.generateDesktopPackageSuggestions(analysis)
		suggestions = append(suggestions, desktopSuggestions...)
	}

	// Security package suggestions
	securitySuggestions := p.generateSecurityPackageSuggestions(analysis)
	suggestions = append(suggestions, securitySuggestions...)

	// Sort by priority and confidence
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Priority != suggestions[j].Priority {
			priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
			return priorityOrder[suggestions[i].Priority] > priorityOrder[suggestions[j].Priority]
		}
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	return suggestions, nil
}

// generateDevelopmentPackageSuggestions generates development package suggestions
func (p *Predictor) generateDevelopmentPackageSuggestions(analysis *SystemAnalysis) []PackageSuggestion {
	var suggestions []PackageSuggestion

	// Git tools
	if !p.packageInstalled(analysis, "git") {
		suggestions = append(suggestions, PackageSuggestion{
			PackageName:    "git",
			Reason:         "Essential version control system for development",
			Category:       "development",
			Priority:       "high",
			Confidence:     0.95,
			InstallCommand: "programs.git.enable = true;",
			Benefits:       []string{"Version control", "Collaboration", "Code history"},
		})
	}

	// Development tools
	commonDevTools := map[string]string{
		"gcc":     "C/C++ compiler for development",
		"python3": "Python programming language",
		"nodejs":  "JavaScript runtime for web development",
		"vim":     "Advanced text editor for development",
		"make":    "Build automation tool",
	}

	for pkg, reason := range commonDevTools {
		if !p.packageInstalled(analysis, pkg) {
			suggestions = append(suggestions, PackageSuggestion{
				PackageName: pkg,
				Reason:      reason,
				Category:    "development",
				Priority:    "medium",
				Confidence:  0.8,
			})
		}
	}

	return suggestions
}

// generateDesktopPackageSuggestions generates desktop package suggestions
func (p *Predictor) generateDesktopPackageSuggestions(analysis *SystemAnalysis) []PackageSuggestion {
	var suggestions []PackageSuggestion

	// Check if desktop environment is in use
	hasDesktop := false
	for _, service := range analysis.EnabledServices {
		if strings.Contains(service.Name, "display-manager") ||
			strings.Contains(service.Name, "xserver") ||
			strings.Contains(service.Name, "wayland") {
			hasDesktop = true
			break
		}
	}

	if !hasDesktop {
		return suggestions // No desktop environment, skip desktop suggestions
	}

	// Common desktop applications
	desktopApps := map[string]PackageSuggestion{
		"firefox": {
			PackageName:    "firefox",
			Reason:         "Modern web browser for desktop use",
			Category:       "desktop",
			Priority:       "high",
			Confidence:     0.9,
			InstallCommand: "programs.firefox.enable = true;",
			Benefits:       []string{"Web browsing", "Privacy features", "Cross-platform sync"},
		},
		"vscode": {
			PackageName:    "vscode",
			Reason:         "Popular code editor with extensive plugin ecosystem",
			Category:       "desktop",
			Priority:       "medium",
			Confidence:     0.8,
			InstallCommand: "programs.vscode.enable = true;",
			Benefits:       []string{"Code editing", "Debugging", "Git integration"},
		},
		"thunderbird": {
			PackageName: "thunderbird",
			Reason:      "Email client for desktop productivity",
			Category:    "desktop",
			Priority:    "low",
			Confidence:  0.6,
			Benefits:    []string{"Email management", "Calendar", "Contacts"},
		},
	}

	for pkg, suggestion := range desktopApps {
		if !p.packageInstalled(analysis, pkg) {
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

// generateSecurityPackageSuggestions generates security package suggestions
func (p *Predictor) generateSecurityPackageSuggestions(analysis *SystemAnalysis) []PackageSuggestion {
	var suggestions []PackageSuggestion

	// Security tools
	securityTools := map[string]PackageSuggestion{
		"gnupg": {
			PackageName:    "gnupg",
			Reason:         "GNU Privacy Guard for encryption and signing",
			Category:       "security",
			Priority:       "medium",
			Confidence:     0.8,
			InstallCommand: "programs.gnupg.agent.enable = true;",
			Benefits:       []string{"File encryption", "Email signing", "Password security"},
		},
		"fail2ban": {
			PackageName:    "fail2ban",
			Reason:         "Intrusion prevention system for network security",
			Category:       "security",
			Priority:       "medium",
			Confidence:     0.7,
			InstallCommand: "services.fail2ban.enable = true;",
			Benefits:       []string{"Brute force protection", "Log monitoring", "Automatic blocking"},
		},
		"clamav": {
			PackageName:    "clamav",
			Reason:         "Antivirus engine for malware detection",
			Category:       "security",
			Priority:       "low",
			Confidence:     0.6,
			InstallCommand: "services.clamav.daemon.enable = true;",
			Benefits:       []string{"Malware scanning", "File protection", "Email filtering"},
		},
	}

	for pkg, suggestion := range securityTools {
		if !p.packageInstalled(analysis, pkg) {
			suggestions = append(suggestions, suggestion)
		}
	}

	return suggestions
}

// packageInstalled checks if a package is installed on the system
func (p *Predictor) packageInstalled(analysis *SystemAnalysis, packageName string) bool {
	for _, pkg := range analysis.InstalledPackages {
		if strings.Contains(strings.ToLower(pkg.Name), strings.ToLower(packageName)) {
			return true
		}
	}
	return false
}

// generateConfigSuggestions generates configuration suggestions
func (p *Predictor) generateConfigSuggestions(ctx context.Context, predictionCtx *PredictionContext) ([]ConfigSuggestion, error) {
	var suggestions []ConfigSuggestion

	analysis := predictionCtx.SystemAnalysis

	// Boot optimization suggestions
	if analysis.PerformanceMetrics.BootTime > 30*time.Second {
		suggestions = append(suggestions, ConfigSuggestion{
			ModuleName:     "boot",
			Option:         "boot.loader.timeout",
			SuggestedValue: 3,
			CurrentValue:   10,
			Reason:         "Reduce boot timeout to improve boot time",
			Impact:         "Faster boot process",
			Priority:       "medium",
			Confidence:     0.8,
			Example:        "boot.loader.timeout = 3;",
		})
	}

	// Memory optimization suggestions
	if analysis.PerformanceMetrics.MemoryUsage > 80.0 {
		suggestions = append(suggestions, ConfigSuggestion{
			ModuleName:     "zramSwap",
			Option:         "zramSwap.enable",
			SuggestedValue: true,
			CurrentValue:   false,
			Reason:         "Enable zram swap to improve memory management",
			Impact:         "Better memory utilization and performance",
			Priority:       "high",
			Confidence:     0.9,
			Example:        "zramSwap.enable = true;",
		})
	}

	// Security hardening suggestions
	if !analysis.SecuritySettings.FirewallEnabled {
		suggestions = append(suggestions, ConfigSuggestion{
			ModuleName:     "networking",
			Option:         "networking.firewall.enable",
			SuggestedValue: true,
			CurrentValue:   false,
			Reason:         "Enable firewall for better security",
			Impact:         "Improved network security",
			Priority:       "high",
			Confidence:     0.95,
			Example:        "networking.firewall.enable = true;",
		})
	}

	return suggestions, nil
}

// generateSecuritySuggestions generates security improvement suggestions
func (p *Predictor) generateSecuritySuggestions(ctx context.Context, predictionCtx *PredictionContext) ([]SecuritySuggestion, error) {
	var suggestions []SecuritySuggestion

	analysis := predictionCtx.SystemAnalysis
	security := analysis.SecuritySettings

	// Firewall suggestion
	if !security.FirewallEnabled {
		suggestions = append(suggestions, SecuritySuggestion{
			Title:             "Enable System Firewall",
			Description:       "Your system firewall is currently disabled, leaving network services potentially exposed",
			Severity:          "high",
			Category:          "firewall",
			CurrentStatus:     "disabled",
			RecommendedAction: "Enable the NixOS firewall with default rules",
			Implementation: []string{
				"Add 'networking.firewall.enable = true;' to configuration.nix",
				"Run 'nixos-rebuild switch' to apply changes",
				"Verify with 'systemctl status firewall'",
			},
			Verification: "iptables -L should show firewall rules",
			Confidence:   0.95,
			Risk:         "low",
			Effort:       "minimal",
		})
	}

	// SSH hardening
	if security.NetworkSecurity.SSHEnabled && security.NetworkSecurity.SSHPasswordAuth {
		suggestions = append(suggestions, SecuritySuggestion{
			Title:             "Disable SSH Password Authentication",
			Description:       "SSH password authentication is enabled, which is less secure than key-based authentication",
			Severity:          "medium",
			Category:          "access",
			CurrentStatus:     "password auth enabled",
			RecommendedAction: "Disable password authentication and use SSH keys only",
			Implementation: []string{
				"Set up SSH key authentication",
				"Add 'services.openssh.passwordAuthentication = false;' to configuration.nix",
				"Rebuild system configuration",
			},
			Verification: "ssh -o PasswordAuthentication=yes should fail",
			Confidence:   0.9,
			Risk:         "medium",
			Effort:       "moderate",
		})
	}

	// Disk encryption
	if !security.EncryptionStatus.FullDiskEncryption {
		suggestions = append(suggestions, SecuritySuggestion{
			Title:             "Consider Full Disk Encryption",
			Description:       "Your system does not appear to use full disk encryption",
			Severity:          "medium",
			Category:          "encryption",
			CurrentStatus:     "unencrypted",
			RecommendedAction: "Enable LUKS disk encryption on next installation",
			Implementation: []string{
				"This requires reinstallation with encrypted partitions",
				"Use nixos-install with LUKS encryption",
				"Backup data before proceeding",
			},
			Verification: "lsblk -f should show crypto_LUKS",
			Confidence:   0.7,
			Risk:         "high",
			Effort:       "significant",
		})
	}

	return suggestions, nil
}

// generatePerformanceSuggestions generates performance optimization suggestions
func (p *Predictor) generatePerformanceSuggestions(ctx context.Context, predictionCtx *PredictionContext) ([]PerformanceSuggestion, error) {
	var suggestions []PerformanceSuggestion

	analysis := predictionCtx.SystemAnalysis
	perf := analysis.PerformanceMetrics

	// Boot time optimization
	if perf.BootTime > 30*time.Second {
		suggestions = append(suggestions, PerformanceSuggestion{
			Component:    "boot",
			Issue:        fmt.Sprintf("Boot time is %.1fs, which is slower than optimal", perf.BootTime.Seconds()),
			Suggestion:   "Optimize systemd services and reduce boot timeout",
			ExpectedGain: "10-20 second improvement in boot time",
			Implementation: []string{
				"Run 'systemd-analyze blame' to identify slow services",
				"Disable unnecessary services",
				"Reduce boot.loader.timeout in configuration",
			},
			Monitoring: "Use 'systemd-analyze' to track improvements",
			Priority:   "medium",
			Confidence: 0.8,
			Complexity: "low",
			Reversible: true,
		})
	}

	// Memory optimization
	if perf.MemoryUsage > 80.0 {
		suggestions = append(suggestions, PerformanceSuggestion{
			Component:    "memory",
			Issue:        fmt.Sprintf("Memory usage is %.1f%%, approaching capacity", perf.MemoryUsage),
			Suggestion:   "Enable zram compression or add swap space",
			ExpectedGain: "20-30% effective memory increase",
			Implementation: []string{
				"Enable zramSwap.enable = true in configuration",
				"Or configure traditional swap space",
				"Monitor memory usage after changes",
			},
			Monitoring: "Use 'free -h' and 'zramctl' to monitor",
			Priority:   "high",
			Confidence: 0.9,
			Complexity: "low",
			Reversible: true,
		})
	}

	// Disk optimization
	for mount, usage := range perf.DiskUsage {
		if usage > 90.0 {
			suggestions = append(suggestions, PerformanceSuggestion{
				Component:    "disk",
				Issue:        fmt.Sprintf("Disk usage on %s is %.1f%%, nearing capacity", mount, usage),
				Suggestion:   "Clean up disk space and optimize storage",
				ExpectedGain: "Improved system responsiveness and stability",
				Implementation: []string{
					"Run 'nix-collect-garbage -d' to clean old generations",
					"Use 'ncdu' to identify large files and directories",
					"Consider moving large files to external storage",
				},
				Monitoring: "Use 'df -h' to monitor disk usage",
				Priority:   "high",
				Confidence: 0.95,
				Complexity: "low",
				Reversible: false,
			})
		}
	}

	return suggestions, nil
}

// generateMaintenanceSuggestions generates maintenance task suggestions
func (p *Predictor) generateMaintenanceSuggestions(ctx context.Context, predictionCtx *PredictionContext) ([]MaintenanceSuggestion, error) {
	var suggestions []MaintenanceSuggestion

	now := time.Now()

	// Regular garbage collection
	suggestions = append(suggestions, MaintenanceSuggestion{
		Task:          "Nix Store Cleanup",
		Description:   "Clean up old Nix store paths and generations to free disk space",
		Frequency:     "weekly",
		NextDue:       now.Add(7 * 24 * time.Hour),
		Commands:      []string{"nix-collect-garbage -d", "nix-store --optimize"},
		Automation:    "Consider setting up automatic garbage collection",
		Priority:      "medium",
		EstimatedTime: 10 * time.Minute,
		Confidence:    0.9,
	})

	// System updates
	suggestions = append(suggestions, MaintenanceSuggestion{
		Task:          "System Updates",
		Description:   "Update system packages and rebuild configuration",
		Frequency:     "weekly",
		NextDue:       now.Add(7 * 24 * time.Hour),
		Commands:      []string{"nixos-rebuild switch --upgrade"},
		Automation:    "Can be automated with systemd timers",
		Priority:      "high",
		EstimatedTime: 30 * time.Minute,
		Confidence:    0.95,
	})

	// Security updates
	suggestions = append(suggestions, MaintenanceSuggestion{
		Task:          "Security Scan",
		Description:   "Scan for security vulnerabilities and apply patches",
		Frequency:     "monthly",
		NextDue:       now.Add(30 * 24 * time.Hour),
		Commands:      []string{"nix-env --upgrade", "systemctl restart sshd"},
		Automation:    "Set up automated security scanning",
		Priority:      "high",
		EstimatedTime: 20 * time.Minute,
		Confidence:    0.8,
	})

	return suggestions, nil
}

// Helper methods for cache and context management

func (p *Predictor) generateCacheKey(ctx *PredictionContext) string {
	// Generate cache key based on system state and user patterns
	contextHash := utils.HashString(fmt.Sprintf("%+v", ctx))
	return fmt.Sprintf("predictions_%s", contextHash)
}

func (p *Predictor) getCachedPrediction(key string) *PredictionResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if result, exists := p.suggestionCache[key]; exists {
		if time.Now().Before(result.ExpiresAt) {
			return result
		}
		// Remove expired cache entry
		delete(p.suggestionCache, key)
	}
	return nil
}

func (p *Predictor) cachePrediction(key string, result *PredictionResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.suggestionCache[key] = result

	// Clean up old cache entries (keep only 10 most recent)
	if len(p.suggestionCache) > 10 {
		oldestKey := ""
		oldestTime := time.Now()
		for k, v := range p.suggestionCache {
			if v.GeneratedAt.Before(oldestTime) {
				oldestTime = v.GeneratedAt
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(p.suggestionCache, oldestKey)
		}
	}
}

func (p *Predictor) getTimeOfDay() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (p *Predictor) calculateOverallConfidence(result *PredictionResult) float64 {
	totalConfidence := 0.0
	totalSuggestions := 0

	for _, suggestion := range result.PackageSuggestions {
		totalConfidence += suggestion.Confidence
		totalSuggestions++
	}

	for _, suggestion := range result.ConfigSuggestions {
		totalConfidence += suggestion.Confidence
		totalSuggestions++
	}

	for _, suggestion := range result.SecuritySuggestions {
		totalConfidence += suggestion.Confidence
		totalSuggestions++
	}

	for _, suggestion := range result.PerformanceSuggestions {
		totalConfidence += suggestion.Confidence
		totalSuggestions++
	}

	for _, suggestion := range result.MaintenanceSuggestions {
		totalConfidence += suggestion.Confidence
		totalSuggestions++
	}

	if totalSuggestions == 0 {
		return 0.5 // Default confidence when no suggestions
	}

	return totalConfidence / float64(totalSuggestions)
}

func (p *Predictor) generateReasoningChain(ctx *PredictionContext, result *PredictionResult) []ReasoningStep {
	var steps []ReasoningStep

	// Step 1: System analysis
	steps = append(steps, ReasoningStep{
		Step:        1,
		Description: "Analyzed current system configuration and state",
		Evidence: []string{
			fmt.Sprintf("System: %s", ctx.SystemAnalysis.SystemType),
			fmt.Sprintf("Packages: %d installed", len(ctx.SystemAnalysis.InstalledPackages)),
			fmt.Sprintf("Services: %d enabled", len(ctx.SystemAnalysis.EnabledServices)),
		},
		Conclusion: "System baseline established",
		Confidence: 0.9,
	})

	// Step 2: User pattern analysis
	steps = append(steps, ReasoningStep{
		Step:        2,
		Description: "Analyzed user behavior patterns and preferences",
		Evidence: []string{
			fmt.Sprintf("Patterns detected: %d", len(ctx.UserPatterns)),
			fmt.Sprintf("Recent activity: %d commands", len(ctx.RecentActivity)),
			fmt.Sprintf("Time context: %s on %s", ctx.TimeOfDay, ctx.DayOfWeek),
		},
		Conclusion: "User context understood",
		Confidence: 0.8,
	})

	// Step 3: Suggestion generation
	totalSuggestions := len(result.PackageSuggestions) + len(result.ConfigSuggestions) +
		len(result.SecuritySuggestions) + len(result.PerformanceSuggestions) +
		len(result.MaintenanceSuggestions)

	steps = append(steps, ReasoningStep{
		Step:        3,
		Description: "Generated targeted suggestions based on analysis",
		Evidence: []string{
			fmt.Sprintf("Package suggestions: %d", len(result.PackageSuggestions)),
			fmt.Sprintf("Config suggestions: %d", len(result.ConfigSuggestions)),
			fmt.Sprintf("Security suggestions: %d", len(result.SecuritySuggestions)),
			fmt.Sprintf("Performance suggestions: %d", len(result.PerformanceSuggestions)),
			fmt.Sprintf("Maintenance suggestions: %d", len(result.MaintenanceSuggestions)),
		},
		Conclusion: fmt.Sprintf("Generated %d total suggestions", totalSuggestions),
		Confidence: result.Confidence,
	})

	return steps
}

func (p *Predictor) extractPatternSummary(ctx *PredictionContext) []string {
	var patterns []string

	for _, pattern := range ctx.UserPatterns {
		patterns = append(patterns, fmt.Sprintf("%s: %s", pattern.Category, pattern.PatternID))
	}

	if len(patterns) == 0 {
		patterns = append(patterns, "No specific user patterns detected")
	}

	return patterns
}

// RecordUserActivity records user activity for pattern analysis
func (p *Predictor) RecordUserActivity(command string, context map[string]string, success bool, duration time.Duration) {
	p.userPatterns.recordActivity(CommandHistory{
		Command:   command,
		Timestamp: time.Now(),
		Context:   context,
		Success:   success,
		Duration:  duration,
	})
}

// ClearCache clears the prediction cache
func (p *Predictor) ClearCache() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.suggestionCache = make(map[string]*PredictionResult)
	p.logger.Info("Prediction cache cleared")
}

// loadPatterns loads user patterns from persistent storage
func (upa *UserPatternAnalyzer) loadPatterns() error {
	upa.mu.Lock()
	defer upa.mu.Unlock()

	if !utils.IsFile(upa.dataFile) {
		return nil // No patterns file exists yet
	}

	data, err := os.ReadFile(upa.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read patterns file: %w", err)
	}

	var storedData struct {
		Patterns       map[string]*UserPattern `json:"patterns"`
		CommandHistory []CommandHistory        `json:"command_history"`
	}

	if err := json.Unmarshal(data, &storedData); err != nil {
		return fmt.Errorf("failed to unmarshal patterns data: %w", err)
	}

	upa.patterns = storedData.Patterns
	upa.commandHistory = storedData.CommandHistory

	return nil
}

// savePatterns saves user patterns to persistent storage
func (upa *UserPatternAnalyzer) savePatterns() error {
	upa.mu.RLock()
	defer upa.mu.RUnlock()

	storedData := struct {
		Patterns       map[string]*UserPattern `json:"patterns"`
		CommandHistory []CommandHistory        `json:"command_history"`
	}{
		Patterns:       upa.patterns,
		CommandHistory: upa.commandHistory,
	}

	data, err := json.MarshalIndent(storedData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patterns data: %w", err)
	}

	if err := os.WriteFile(upa.dataFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write patterns file: %w", err)
	}

	return nil
}

// getRecentPatterns returns recently detected patterns
func (upa *UserPatternAnalyzer) getRecentPatterns(limit int) []string {
	upa.mu.RLock()
	defer upa.mu.RUnlock()

	var patterns []string
	cutoff := time.Now().Add(-24 * time.Hour) // Last 24 hours

	for _, pattern := range upa.patterns {
		if pattern.LastSeen.After(cutoff) {
			patterns = append(patterns, fmt.Sprintf("%s: %s (confidence: %.2f)",
				pattern.Category, pattern.PatternID, pattern.Confidence))
		}
	}

	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	return patterns
}

// getRecentActivity returns recent command activity
func (upa *UserPatternAnalyzer) getRecentActivity(limit int) []string {
	upa.mu.RLock()
	defer upa.mu.RUnlock()

	var activity []string
	cutoff := time.Now().Add(-24 * time.Hour) // Last 24 hours

	for _, cmd := range upa.commandHistory {
		if cmd.Timestamp.After(cutoff) {
			status := "success"
			if !cmd.Success {
				status = "failed"
			}
			activity = append(activity, fmt.Sprintf("%s (%s)", cmd.Command, status))
		}
	}

	if len(activity) > limit {
		activity = activity[:limit]
	}

	return activity
}

// hasPattern checks if a specific pattern exists
func (upa *UserPatternAnalyzer) hasPattern(category string) bool {
	upa.mu.RLock()
	defer upa.mu.RUnlock()

	for _, pattern := range upa.patterns {
		if pattern.Category == category && pattern.Confidence > 0.7 {
			return true
		}
	}

	return false
}

// recordActivity records a new command activity
func (upa *UserPatternAnalyzer) recordActivity(cmd CommandHistory) {
	upa.mu.Lock()
	defer upa.mu.Unlock()

	// Add to command history
	upa.commandHistory = append(upa.commandHistory, cmd)

	// Keep only last 1000 commands
	if len(upa.commandHistory) > 1000 {
		upa.commandHistory = upa.commandHistory[len(upa.commandHistory)-1000:]
	}

	// Update patterns based on command
	upa.updatePatternsFromCommand(cmd)

	// Save patterns periodically
	go func() {
		_ = upa.savePatterns()
	}()
}

// updatePatternsFromCommand updates patterns based on command activity
func (upa *UserPatternAnalyzer) updatePatternsFromCommand(cmd CommandHistory) {
	category := upa.categorizeCommand(cmd.Command)
	patternID := fmt.Sprintf("%s_%s", category, utils.HashString(cmd.Command)[:8])

	if pattern, exists := upa.patterns[patternID]; exists {
		pattern.Frequency++
		pattern.LastSeen = cmd.Timestamp
		pattern.Confidence = upa.calculateConfidence(pattern)
	} else {
		upa.patterns[patternID] = &UserPattern{
			PatternID:  patternID,
			Category:   category,
			Commands:   []string{cmd.Command},
			Frequency:  1,
			LastSeen:   cmd.Timestamp,
			Context:    cmd.Context,
			Confidence: 0.5, // Initial confidence
		}
	}
}

// categorizeCommand categorizes a command based on its content
func (upa *UserPatternAnalyzer) categorizeCommand(command string) string {
	command = strings.ToLower(command)

	// Development patterns
	if strings.Contains(command, "git") || strings.Contains(command, "gcc") ||
		strings.Contains(command, "make") || strings.Contains(command, "cargo") ||
		strings.Contains(command, "npm") || strings.Contains(command, "python") {
		return "development"
	}

	// System administration patterns
	if strings.Contains(command, "systemctl") || strings.Contains(command, "sudo") ||
		strings.Contains(command, "nixos-rebuild") || strings.Contains(command, "journalctl") {
		return "administration"
	}

	// Desktop patterns
	if strings.Contains(command, "firefox") || strings.Contains(command, "chrome") ||
		strings.Contains(command, "code") || strings.Contains(command, "vim") {
		return "desktop"
	}

	// Security patterns
	if strings.Contains(command, "gpg") || strings.Contains(command, "ssh") ||
		strings.Contains(command, "firewall") || strings.Contains(command, "cert") {
		return "security"
	}

	return "general"
}

// calculateConfidence calculates confidence score for a pattern
func (upa *UserPatternAnalyzer) calculateConfidence(pattern *UserPattern) float64 {
	// Base confidence on frequency and recency
	frequencyScore := float64(pattern.Frequency) / 100.0
	if frequencyScore > 1.0 {
		frequencyScore = 1.0
	}

	recencyScore := 1.0
	daysSinceLastSeen := time.Since(pattern.LastSeen).Hours() / 24.0
	if daysSinceLastSeen > 7 {
		recencyScore = 0.5
	} else if daysSinceLastSeen > 1 {
		recencyScore = 0.8
	}

	return (frequencyScore + recencyScore) / 2.0
}

// getRecentCommandHistory returns recent command history as CommandHistory slice
func (upa *UserPatternAnalyzer) getRecentCommandHistory(limit int) []CommandHistory {
	upa.mu.RLock()
	defer upa.mu.RUnlock()

	var history []CommandHistory
	cutoff := time.Now().Add(-24 * time.Hour) // Last 24 hours

	for _, cmd := range upa.commandHistory {
		if cmd.Timestamp.After(cutoff) {
			history = append(history, cmd)
		}
	}

	if len(history) > limit {
		history = history[:limit]
	}

	return history
}

// getRecentUserPatterns returns recent patterns as UserPattern slice
func (upa *UserPatternAnalyzer) getRecentUserPatterns(limit int) []UserPattern {
	upa.mu.RLock()
	defer upa.mu.RUnlock()

	var patterns []UserPattern
	cutoff := time.Now().Add(-24 * time.Hour) // Last 24 hours

	for _, pattern := range upa.patterns {
		if pattern.LastSeen.After(cutoff) {
			patterns = append(patterns, *pattern)
		}
	}

	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	return patterns
}
