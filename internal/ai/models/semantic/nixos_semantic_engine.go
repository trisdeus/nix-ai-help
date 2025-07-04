package semantic

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// NixOSSemanticEngine provides advanced semantic understanding specifically for NixOS configurations
type NixOSSemanticEngine struct {
	analyzer         *SemanticAnalyzer
	logger           *logger.Logger
	knowledgeBase    *NixOSKnowledgeBase
	patternRecognizer *ConfigPatternRecognizer
}

// NixOSKnowledgeBase contains domain-specific knowledge about NixOS
type NixOSKnowledgeBase struct {
	ServicePatterns    map[string]ServicePattern    `json:"service_patterns"`
	SecurityRules      []SecurityRule               `json:"security_rules"`
	PerformanceRules   []PerformanceRule           `json:"performance_rules"`
	BestPractices      []BestPractice              `json:"best_practices"`
	CommonMistakes     []CommonMistake             `json:"common_mistakes"`
	UpdatedAt          time.Time                   `json:"updated_at"`
}

// ServicePattern defines patterns for recognizing service configurations
type ServicePattern struct {
	Name            string            `json:"name"`
	Category        string            `json:"category"`       // "web", "database", "security", etc.
	RequiredOptions []string          `json:"required_options"`
	OptionalOptions []string          `json:"optional_options"`
	Conflicts       []string          `json:"conflicts"`      // Services that conflict with this one
	Dependencies    []string          `json:"dependencies"`   // Services this depends on
	SecurityRisks   []string          `json:"security_risks"`
	Performance     PerformanceImpact `json:"performance"`
	Documentation   string            `json:"documentation"`
}

// SecurityRule defines security-related rules for configuration analysis
type SecurityRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Severity    string   `json:"severity"`
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	Fix         string   `json:"fix"`
	References  []string `json:"references"`
	CISControl  string   `json:"cis_control,omitempty"`
}

// PerformanceRule defines performance-related rules
type PerformanceRule struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Impact       string `json:"impact"`        // "high", "medium", "low"
	Pattern      string `json:"pattern"`
	Description  string `json:"description"`
	Optimization string `json:"optimization"`
	Benefit      string `json:"benefit"`
}

// BestPractice defines NixOS configuration best practices
type BestPractice struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Level       string   `json:"level"`         // "beginner", "intermediate", "advanced"
	Description string   `json:"description"`
	Example     string   `json:"example"`
	Rationale   string   `json:"rationale"`
	References  []string `json:"references"`
}

// CommonMistake defines common configuration mistakes
type CommonMistake struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Problem     string `json:"problem"`
	Solution    string `json:"solution"`
	Frequency   string `json:"frequency"`     // "very_common", "common", "uncommon"
	Severity    string `json:"severity"`
}

// PerformanceImpact describes the performance impact of a service
type PerformanceImpact struct {
	CPU      string `json:"cpu"`       // "low", "medium", "high"
	Memory   string `json:"memory"`
	Disk     string `json:"disk"`
	Network  string `json:"network"`
	Startup  string `json:"startup"`   // Impact on boot time
}

// ConfigPatternRecognizer recognizes high-level patterns in configurations
type ConfigPatternRecognizer struct {
	logger *logger.Logger
}

// SemanticIntentResult contains advanced intent analysis results
type SemanticIntentResult struct {
	ConfigIntent
	ArchitecturalPattern string                 `json:"architectural_pattern"`
	DeploymentModel      string                 `json:"deployment_model"`
	SecurityPosture      string                 `json:"security_posture"`
	MaintenanceLevel     string                 `json:"maintenance_level"`
	ScalabilityRating    float64               `json:"scalability_rating"`
	ComplexityMetrics    ComplexityMetrics     `json:"complexity_metrics"`
	TechnicalDebt        []TechnicalDebtItem   `json:"technical_debt"`
	EvolutionPath        []EvolutionSuggestion `json:"evolution_path"`
}

// ComplexityMetrics provides detailed complexity analysis
type ComplexityMetrics struct {
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	ConfigurationDepth   int     `json:"configuration_depth"`
	ServiceDensity       float64 `json:"service_density"`
	DependencyComplexity float64 `json:"dependency_complexity"`
	OverallScore         float64 `json:"overall_score"`
}

// TechnicalDebtItem represents technical debt in the configuration
type TechnicalDebtItem struct {
	Type        string `json:"type"`         // "deprecated", "anti-pattern", "workaround"
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort_to_fix"`
	Priority    string `json:"priority"`
}

// EvolutionSuggestion suggests how the configuration could evolve
type EvolutionSuggestion struct {
	Phase       string `json:"phase"`        // "immediate", "short_term", "long_term"
	Type        string `json:"type"`         // "modernization", "optimization", "security"
	Description string `json:"description"`
	Benefits    string `json:"benefits"`
	Effort      string `json:"effort"`
}

// NewNixOSSemanticEngine creates a new NixOS-specific semantic engine
func NewNixOSSemanticEngine() *NixOSSemanticEngine {
	engine := &NixOSSemanticEngine{
		analyzer:          NewSemanticAnalyzer(),
		logger:           logger.NewLogger(),
		knowledgeBase:    NewNixOSKnowledgeBase(),
		patternRecognizer: NewConfigPatternRecognizer(),
	}
	
	return engine
}

// NewNixOSKnowledgeBase creates a new knowledge base with NixOS-specific patterns
func NewNixOSKnowledgeBase() *NixOSKnowledgeBase {
	return &NixOSKnowledgeBase{
		ServicePatterns: map[string]ServicePattern{
			"nginx": {
				Name:     "nginx",
				Category: "web",
				RequiredOptions: []string{"enable"},
				OptionalOptions: []string{"virtualHosts", "recommendedProxySettings", "recommendedTlsSettings"},
				Dependencies:    []string{},
				SecurityRisks:   []string{"default_server_block", "missing_tls"},
				Performance: PerformanceImpact{
					CPU:     "medium",
					Memory:  "low",
					Disk:    "low",
					Network: "high",
					Startup: "fast",
				},
				Documentation: "https://nixos.org/manual/nixos/stable/options.html#opt-services.nginx",
			},
			"postgresql": {
				Name:     "postgresql",
				Category: "database",
				RequiredOptions: []string{"enable"},
				OptionalOptions: []string{"package", "port", "dataDir", "authentication"},
				Dependencies:    []string{},
				SecurityRisks:   []string{"weak_authentication", "public_access"},
				Performance: PerformanceImpact{
					CPU:     "medium",
					Memory:  "high",
					Disk:    "high",
					Network: "medium",
					Startup: "slow",
				},
				Documentation: "https://nixos.org/manual/nixos/stable/options.html#opt-services.postgresql",
			},
		},
		SecurityRules: []SecurityRule{
			{
				ID:          "SEC001",
				Name:        "Firewall Disabled",
				Category:    "network",
				Severity:    "high",
				Pattern:     `networking\.firewall\.enable\s*=\s*false`,
				Description: "Firewall is explicitly disabled",
				Fix:         "Set networking.firewall.enable = true;",
				References:  []string{"https://nixos.org/manual/nixos/stable/#sec-firewall"},
				CISControl:  "9.4",
			},
			{
				ID:          "SEC002", 
				Name:        "Root SSH Login",
				Category:    "ssh",
				Severity:    "critical",
				Pattern:     `PermitRootLogin\s*=\s*"yes"`,
				Description: "Root SSH login is enabled",
				Fix:         "Set PermitRootLogin = \"no\";",
				References:  []string{"https://nixos.org/manual/nixos/stable/#sec-ssh"},
				CISControl:  "4.3",
			},
		},
		PerformanceRules: []PerformanceRule{
			{
				ID:           "PERF001",
				Name:         "Missing ZRAM",
				Category:     "memory",
				Impact:       "medium",
				Pattern:      `!(zramSwap\.enable\s*=\s*true)`,
				Description:  "ZRAM swap not enabled",
				Optimization: "Enable zramSwap.enable = true;",
				Benefit:      "Improved memory management without disk I/O",
			},
		},
		BestPractices: []BestPractice{
			{
				ID:          "BP001",
				Name:        "Use Flakes",
				Category:    "modern_nix",
				Level:       "intermediate",
				Description: "Use Nix flakes for reproducible configurations",
				Example:     "Enable experimental features and use flake.nix",
				Rationale:   "Flakes provide better reproducibility and dependency management",
				References:  []string{"https://nixos.wiki/wiki/Flakes"},
			},
		},
		CommonMistakes: []CommonMistake{
			{
				ID:        "CM001",
				Name:      "Forgetting to rebuild",
				Pattern:   `.*`,
				Problem:   "Configuration changes without nixos-rebuild",
				Solution:  "Always run 'sudo nixos-rebuild switch' after configuration changes",
				Frequency: "very_common",
				Severity:  "medium",
			},
		},
		UpdatedAt: time.Now(),
	}
}

// NewConfigPatternRecognizer creates a new pattern recognizer
func NewConfigPatternRecognizer() *ConfigPatternRecognizer {
	return &ConfigPatternRecognizer{
		logger: logger.NewLogger(),
	}
}

// AnalyzeSemanticIntent performs advanced semantic intent analysis
func (engine *NixOSSemanticEngine) AnalyzeSemanticIntent(ctx context.Context, configPath string, content string) (*SemanticIntentResult, error) {
	engine.logger.Info(fmt.Sprintf("Starting advanced semantic intent analysis for %s", configPath))

	// Get basic intent from standard analyzer
	basicResult, err := engine.analyzer.AnalyzeConfiguration(ctx, configPath, content)
	if err != nil {
		return nil, fmt.Errorf("basic analysis failed: %w", err)
	}

	result := &SemanticIntentResult{
		ConfigIntent: basicResult.Intent,
	}

	// Detect architectural pattern
	result.ArchitecturalPattern = engine.detectArchitecturalPattern(content)
	
	// Detect deployment model
	result.DeploymentModel = engine.detectDeploymentModel(content)
	
	// Assess security posture
	result.SecurityPosture = engine.assessSecurityPosture(content)
	
	// Determine maintenance level required
	result.MaintenanceLevel = engine.assessMaintenanceLevel(content)
	
	// Calculate scalability rating
	result.ScalabilityRating = engine.calculateScalabilityRating(content)
	
	// Analyze complexity metrics
	result.ComplexityMetrics = engine.analyzeComplexityMetrics(content)
	
	// Identify technical debt
	result.TechnicalDebt = engine.identifyTechnicalDebt(content)
	
	// Generate evolution path
	result.EvolutionPath = engine.generateEvolutionPath(content, result)

	engine.logger.Info(fmt.Sprintf("Advanced semantic analysis completed: pattern=%s, deployment=%s, security=%s, complexity=%.2f", 
		result.ArchitecturalPattern, result.DeploymentModel, result.SecurityPosture, result.ComplexityMetrics.OverallScore))

	return result, nil
}

// detectArchitecturalPattern identifies the overall architectural pattern
func (engine *NixOSSemanticEngine) detectArchitecturalPattern(content string) string {
	// Microservices pattern
	if engine.containsPatterns(content, []string{"docker", "kubernetes", "containers", "services."}) {
		serviceCount := len(regexp.MustCompile(`services\.(\w+)`).FindAllString(content, -1))
		if serviceCount > 5 {
			return "microservices"
		}
	}
	
	// Monolithic pattern
	if engine.containsPatterns(content, []string{"services.nginx", "services.postgresql", "services.redis"}) {
		return "monolithic_web_stack"
	}
	
	// Desktop pattern
	if engine.containsPatterns(content, []string{"services.xserver", "services.gnome", "services.kde"}) {
		return "desktop_environment"
	}
	
	// Server pattern
	if engine.containsPatterns(content, []string{"services.nginx", "services.apache", "networking.firewall"}) {
		return "traditional_server"
	}
	
	// Development pattern
	if engine.containsPatterns(content, []string{"development", "nix-shell", "devenv"}) {
		return "development_environment"
	}
	
	return "minimal_configuration"
}

// detectDeploymentModel identifies the deployment model
func (engine *NixOSSemanticEngine) detectDeploymentModel(content string) string {
	if strings.Contains(content, "flake.nix") || strings.Contains(content, "inputs.nixpkgs") {
		return "flake_based"
	}
	
	if strings.Contains(content, "channels") || strings.Contains(content, "nixos-channel") {
		return "channel_based"
	}
	
	if strings.Contains(content, "containers") || strings.Contains(content, "virtualisation") {
		return "containerized"
	}
	
	return "traditional"
}

// assessSecurityPosture evaluates the overall security configuration
func (engine *NixOSSemanticEngine) assessSecurityPosture(content string) string {
	securityScore := 0
	
	// Positive security indicators
	if strings.Contains(content, "networking.firewall.enable = true") {
		securityScore += 2
	}
	if strings.Contains(content, "services.fail2ban") {
		securityScore += 2
	}
	if strings.Contains(content, "PermitRootLogin = \"no\"") {
		securityScore += 2
	}
	if strings.Contains(content, "users.users") && !strings.Contains(content, "users.users.root") {
		securityScore += 1
	}
	
	// Negative security indicators
	if strings.Contains(content, "networking.firewall.enable = false") {
		securityScore -= 3
	}
	if strings.Contains(content, "PermitRootLogin = \"yes\"") {
		securityScore -= 3
	}
	if strings.Contains(content, "PasswordAuthentication = true") {
		securityScore -= 1
	}
	
	if securityScore >= 4 {
		return "hardened"
	} else if securityScore >= 1 {
		return "secure"
	} else if securityScore >= -1 {
		return "basic"
	} else {
		return "vulnerable"
	}
}

// assessMaintenanceLevel determines the maintenance complexity
func (engine *NixOSSemanticEngine) assessMaintenanceLevel(content string) string {
	complexity := 0
	
	// Count complex configurations
	complexity += strings.Count(content, "systemd.services")
	complexity += strings.Count(content, "systemd.timers")
	complexity += strings.Count(content, "networking.interfaces")
	complexity += strings.Count(content, "boot.initrd")
	
	// Count services
	serviceMatches := regexp.MustCompile(`services\.(\w+)`).FindAllString(content, -1)
	complexity += len(serviceMatches)
	
	if complexity > 20 {
		return "high"
	} else if complexity > 10 {
		return "medium"
	} else {
		return "low"
	}
}

// calculateScalabilityRating assesses how well the configuration can scale
func (engine *NixOSSemanticEngine) calculateScalabilityRating(content string) float64 {
	rating := 0.5 // Base rating
	
	// Positive scalability indicators
	if strings.Contains(content, "containers") {
		rating += 0.2
	}
	if strings.Contains(content, "load-balancer") || strings.Contains(content, "nginx.upstreams") {
		rating += 0.2
	}
	if strings.Contains(content, "clustering") || strings.Contains(content, "replication") {
		rating += 0.1
	}
	
	// Negative scalability indicators
	if strings.Contains(content, "localhost") && strings.Contains(content, "database") {
		rating -= 0.2
	}
	if strings.Contains(content, "sqlite") {
		rating -= 0.1
	}
	
	// Normalize to 0-1 range
	if rating > 1.0 {
		rating = 1.0
	}
	if rating < 0.0 {
		rating = 0.0
	}
	
	return rating
}

// analyzeComplexityMetrics provides detailed complexity analysis
func (engine *NixOSSemanticEngine) analyzeComplexityMetrics(content string) ComplexityMetrics {
	metrics := ComplexityMetrics{}
	
	// Cyclomatic complexity (simplified)
	branchingKeywords := []string{"if", "then", "else", "case", "when"}
	for _, keyword := range branchingKeywords {
		metrics.CyclomaticComplexity += strings.Count(content, keyword)
	}
	
	// Configuration depth (nesting levels)
	maxDepth := 0
	currentDepth := 0
	for _, char := range content {
		if char == '{' {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		} else if char == '}' {
			currentDepth--
		}
	}
	metrics.ConfigurationDepth = maxDepth
	
	// Service density
	serviceCount := len(regexp.MustCompile(`services\.(\w+)`).FindAllString(content, -1))
	lineCount := len(strings.Split(content, "\n"))
	if lineCount > 0 {
		metrics.ServiceDensity = float64(serviceCount) / float64(lineCount) * 100
	}
	
	// Dependency complexity (simplified)
	importCount := strings.Count(content, "import")
	packageCount := strings.Count(content, "pkgs.")
	metrics.DependencyComplexity = float64(importCount + packageCount)
	
	// Overall score (0-1, where 1 is most complex)
	complexityFactors := float64(metrics.CyclomaticComplexity)/10 + 
					   float64(metrics.ConfigurationDepth)/20 + 
					   metrics.ServiceDensity/10 + 
					   metrics.DependencyComplexity/50
	
	if complexityFactors > 1.0 {
		complexityFactors = 1.0
	}
	metrics.OverallScore = complexityFactors
	
	return metrics
}

// identifyTechnicalDebt finds technical debt in the configuration
func (engine *NixOSSemanticEngine) identifyTechnicalDebt(content string) []TechnicalDebtItem {
	var debt []TechnicalDebtItem
	
	// Check for deprecated options
	deprecatedPatterns := map[string]TechnicalDebtItem{
		"sound.enable": {
			Type:        "deprecated",
			Description: "sound.enable is deprecated in favor of PipeWire/PulseAudio",
			Impact:      "medium",
			Effort:      "low",
			Priority:    "medium",
		},
		"services.mysql": {
			Type:        "deprecated",
			Description: "services.mysql is deprecated in favor of services.mysql.package = pkgs.mariadb",
			Impact:      "low",
			Effort:      "low",
			Priority:    "low",
		},
	}
	
	for pattern, debtItem := range deprecatedPatterns {
		if strings.Contains(content, pattern) {
			debt = append(debt, debtItem)
		}
	}
	
	// Check for anti-patterns
	if strings.Contains(content, "networking.firewall.enable = false") {
		debt = append(debt, TechnicalDebtItem{
			Type:        "anti-pattern",
			Description: "Firewall completely disabled instead of proper configuration",
			Impact:      "high",
			Effort:      "medium",
			Priority:    "high",
		})
	}
	
	return debt
}

// generateEvolutionPath suggests how the configuration could evolve
func (engine *NixOSSemanticEngine) generateEvolutionPath(content string, result *SemanticIntentResult) []EvolutionSuggestion {
	var suggestions []EvolutionSuggestion
	
	// Immediate improvements
	if len(result.TechnicalDebt) > 0 {
		suggestions = append(suggestions, EvolutionSuggestion{
			Phase:       "immediate",
			Type:        "modernization",
			Description: "Address technical debt and deprecated configurations",
			Benefits:    "Improved maintainability and future compatibility",
			Effort:      "low",
		})
	}
	
	// Short-term improvements
	if result.SecurityPosture == "basic" || result.SecurityPosture == "vulnerable" {
		suggestions = append(suggestions, EvolutionSuggestion{
			Phase:       "short_term",
			Type:        "security",
			Description: "Implement comprehensive security hardening",
			Benefits:    "Reduced attack surface and improved compliance",
			Effort:      "medium",
		})
	}
	
	// Long-term evolution
	if result.ArchitecturalPattern == "traditional_server" && result.ScalabilityRating < 0.6 {
		suggestions = append(suggestions, EvolutionSuggestion{
			Phase:       "long_term",
			Type:        "optimization",
			Description: "Consider containerization and microservices architecture",
			Benefits:    "Improved scalability and maintainability",
			Effort:      "high",
		})
	}
	
	return suggestions
}

// containsPatterns checks if content contains any of the given patterns
func (engine *NixOSSemanticEngine) containsPatterns(content string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}
	return false
}

// GetKnowledgeBase returns the current knowledge base
func (engine *NixOSSemanticEngine) GetKnowledgeBase() *NixOSKnowledgeBase {
	return engine.knowledgeBase
}

// UpdateKnowledgeBase updates the knowledge base with new patterns
func (engine *NixOSSemanticEngine) UpdateKnowledgeBase(kb *NixOSKnowledgeBase) {
	engine.knowledgeBase = kb
	engine.knowledgeBase.UpdatedAt = time.Now()
	engine.logger.Info("Knowledge base updated successfully")
}

// ExportKnowledgeBase exports the knowledge base to JSON
func (engine *NixOSSemanticEngine) ExportKnowledgeBase() (string, error) {
	data, err := json.MarshalIndent(engine.knowledgeBase, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to export knowledge base: %w", err)
	}
	return string(data), nil
}