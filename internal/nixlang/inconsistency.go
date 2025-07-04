// Package nixlang provides logical inconsistency detection for NixOS configurations
package nixlang

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// InconsistencyDetector analyzes NixOS configurations for logical inconsistencies
type InconsistencyDetector struct {
	analyzer *NixAnalyzer
}

// InconsistencyType represents different types of logical inconsistencies
type InconsistencyType string

const (
	ConflictingServices    InconsistencyType = "conflicting_services"
	MissingDependencies    InconsistencyType = "missing_dependencies"
	InvalidConfiguration   InconsistencyType = "invalid_configuration"
	ResourceConflicts      InconsistencyType = "resource_conflicts"
	SecurityInconsistency  InconsistencyType = "security_inconsistency"
	PerformanceConflict    InconsistencyType = "performance_conflict"
	VersionConflict        InconsistencyType = "version_conflict"
	ServiceInteraction     InconsistencyType = "service_interaction"
	NetworkingConflict     InconsistencyType = "networking_conflict"
	FilesystemConflict     InconsistencyType = "filesystem_conflict"
)

// LogicalInconsistency represents a detected logical inconsistency
type LogicalInconsistency struct {
	Type            InconsistencyType      `json:"type"`
	Severity        string                 `json:"severity"`        // "critical", "major", "minor", "info"
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	ConflictingItems []ConflictingItem     `json:"conflicting_items"`
	Resolution      *Resolution            `json:"resolution,omitempty"`
	Evidence        []Evidence             `json:"evidence"`
	Impact          string                 `json:"impact"`
	Confidence      float64                `json:"confidence"`      // 0.0 - 1.0
	Location        []Location             `json:"location"`
	References      []string               `json:"references"`
	Context         map[string]interface{} `json:"context"`
}

// ConflictingItem represents an item that conflicts with another
type ConflictingItem struct {
	Type        string    `json:"type"`        // "service", "option", "package", "module"
	Name        string    `json:"name"`
	Value       string    `json:"value,omitempty"`
	Location    Location  `json:"location"`
	Reason      string    `json:"reason"`
	Context     string    `json:"context"`
}

// Resolution provides suggested resolution for an inconsistency
type Resolution struct {
	Type          string            `json:"type"`           // "choose_one", "modify", "add_dependency", "remove", "configure"
	Description   string            `json:"description"`
	Options       []ResolutionOption `json:"options"`
	Automatic     bool              `json:"automatic"`      // Can be auto-resolved
	Complexity    string            `json:"complexity"`     // "simple", "moderate", "complex"
	Risk          string            `json:"risk"`          // "low", "medium", "high"
}

// ResolutionOption represents a specific resolution option
type ResolutionOption struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Changes     []ConfigChange    `json:"changes"`
	Pros        []string          `json:"pros"`
	Cons        []string          `json:"cons"`
	Risk        string            `json:"risk"`
}

// ConfigChange represents a specific configuration change
type ConfigChange struct {
	Type        string    `json:"type"`        // "add", "remove", "modify", "move"
	Path        string    `json:"path"`        // Configuration path like "services.nginx.enable"
	OldValue    string    `json:"old_value,omitempty"`
	NewValue    string    `json:"new_value"`
	Location    Location  `json:"location,omitempty"`
	Reason      string    `json:"reason"`
}

// Evidence represents evidence supporting the inconsistency detection
type Evidence struct {
	Type        string            `json:"type"`        // "pattern_match", "rule_violation", "logical_analysis"
	Description string            `json:"description"`
	Data        map[string]string `json:"data"`
	Confidence  float64           `json:"confidence"`
}

// InconsistencyResult contains all detected inconsistencies
type InconsistencyResult struct {
	Inconsistencies []LogicalInconsistency `json:"inconsistencies"`
	Summary         InconsistencySummary   `json:"summary"`
	Statistics      InconsistencyStats     `json:"statistics"`
}

// InconsistencySummary provides a high-level summary
type InconsistencySummary struct {
	TotalInconsistencies int               `json:"total_inconsistencies"`
	BySeverity          map[string]int    `json:"by_severity"`
	ByType              map[string]int    `json:"by_type"`
	CriticalIssues      []string          `json:"critical_issues"`
	RecommendedActions  []string          `json:"recommended_actions"`
	OverallRisk         string            `json:"overall_risk"`
}

// InconsistencyStats provides detailed statistics
type InconsistencyStats struct {
	AutoResolvable      int     `json:"auto_resolvable"`
	ManualResolution    int     `json:"manual_resolution"`
	AverageConfidence   float64 `json:"average_confidence"`
	MostCommonType      string  `json:"most_common_type"`
	ConfigurationHealth string  `json:"configuration_health"` // "excellent", "good", "fair", "poor"
}

// NewInconsistencyDetector creates a new inconsistency detector
func NewInconsistencyDetector(analyzer *NixAnalyzer) *InconsistencyDetector {
	return &InconsistencyDetector{
		analyzer: analyzer,
	}
}

// DetectInconsistencies performs comprehensive logical inconsistency detection
func (id *InconsistencyDetector) DetectInconsistencies(content string) (*InconsistencyResult, error) {
	var allInconsistencies []LogicalInconsistency

	// Parse configuration first (continue even if parsing fails for basic regex-based detection)
	expr, err := id.analyzer.parser.ParseExpression(content)
	if err != nil {
		// Continue with nil expr for regex-based detection
		expr = nil
	}

	// Detect different types of inconsistencies
	serviceConflicts := id.detectServiceConflicts(content, expr)
	allInconsistencies = append(allInconsistencies, serviceConflicts...)

	dependencyIssues := id.detectMissingDependencies(content, expr)
	allInconsistencies = append(allInconsistencies, dependencyIssues...)

	resourceConflicts := id.detectResourceConflicts(content, expr)
	allInconsistencies = append(allInconsistencies, resourceConflicts...)

	securityInconsistencies := id.detectSecurityInconsistencies(content, expr)
	allInconsistencies = append(allInconsistencies, securityInconsistencies...)

	networkingConflicts := id.detectNetworkingConflicts(content, expr)
	allInconsistencies = append(allInconsistencies, networkingConflicts...)

	filesystemConflicts := id.detectFilesystemConflicts(content, expr)
	allInconsistencies = append(allInconsistencies, filesystemConflicts...)

	versionConflicts := id.detectVersionConflicts(content, expr)
	allInconsistencies = append(allInconsistencies, versionConflicts...)

	performanceConflicts := id.detectPerformanceConflicts(content, expr)
	allInconsistencies = append(allInconsistencies, performanceConflicts...)

	// Generate summary and statistics
	summary := id.generateSummary(allInconsistencies)
	stats := id.generateStatistics(allInconsistencies)

	return &InconsistencyResult{
		Inconsistencies: allInconsistencies,
		Summary:         summary,
		Statistics:      stats,
	}, nil
}

// detectServiceConflicts detects conflicting service configurations
func (id *InconsistencyDetector) detectServiceConflicts(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for conflicting web servers
	webServers := []string{"nginx", "apache", "lighttpd", "caddy"}
	enabledWebServers := id.findEnabledServices(content, webServers)
	if len(enabledWebServers) > 1 {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        ConflictingServices,
			Severity:    "major",
			Title:       "Multiple Web Servers Enabled",
			Description: fmt.Sprintf("Multiple web servers are enabled: %s. This can cause port conflicts and service failures.", strings.Join(enabledWebServers, ", ")),
			ConflictingItems: id.createConflictingItems(enabledWebServers, "service", "Web servers typically bind to the same ports (80, 443)"),
			Resolution: &Resolution{
				Type:        "choose_one",
				Description: "Choose one web server and disable the others",
				Options:     id.createWebServerResolutionOptions(enabledWebServers),
				Automatic:   false,
				Complexity:  "simple",
				Risk:        "low",
			},
			Evidence: []Evidence{{
				Type:        "pattern_match",
				Description: "Found multiple enabled web server services",
				Data:        map[string]string{"servers": strings.Join(enabledWebServers, ", ")},
				Confidence:  0.95,
			}},
			Impact:     "Services may fail to start due to port conflicts",
			Confidence: 0.95,
			References: []string{
				"https://nixos.org/manual/nixos/stable/#sec-nginx",
				"https://nixos.org/manual/nixos/stable/#sec-apache",
			},
		})
	}

	// Check for conflicting database servers
	dbServers := []string{"postgresql", "mysql", "mariadb", "mongodb"}
	enabledDbServers := id.findEnabledServices(content, dbServers)
	if len(enabledDbServers) > 1 {
		// General database conflict
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        ConflictingServices,
			Severity:    "major",
			Title:       "Multiple Database Servers Enabled",
			Description: fmt.Sprintf("Multiple database servers are enabled: %s. This can cause port conflicts and resource issues.", strings.Join(enabledDbServers, ", ")),
			ConflictingItems: id.createConflictingItems(enabledDbServers, "service", "Database servers typically bind to the same default ports"),
			Resolution: &Resolution{
				Type:        "choose_one",
				Description: "Choose one database server and disable the others",
				Options:     id.createDatabaseResolutionOptions(enabledDbServers),
				Automatic:   false,
				Complexity:  "simple",
				Risk:        "medium",
			},
			Evidence: []Evidence{{
				Type:        "pattern_match",
				Description: "Found multiple enabled database servers",
				Data:        map[string]string{"databases": strings.Join(enabledDbServers, ", ")},
				Confidence:  0.9,
			}},
			Impact:     "Database services may fail to start due to conflicts",
			Confidence: 0.9,
		})

		// Special case: MySQL and MariaDB are particularly problematic together
		if id.containsAll(enabledDbServers, []string{"mysql", "mariadb"}) {
			inconsistencies = append(inconsistencies, LogicalInconsistency{
				Type:        ConflictingServices,
				Severity:    "critical",
				Title:       "MySQL and MariaDB Both Enabled",
				Description: "Both MySQL and MariaDB are enabled. These are incompatible and will cause conflicts.",
				ConflictingItems: id.createConflictingItems([]string{"mysql", "mariadb"}, "service", "MySQL and MariaDB are incompatible database implementations"),
				Resolution: &Resolution{
					Type:        "choose_one",
					Description: "Choose either MySQL or MariaDB, not both",
					Options: []ResolutionOption{{
						Title:       "Use MariaDB (Recommended)",
						Description: "MariaDB is a drop-in replacement for MySQL with better performance",
						Changes: []ConfigChange{{
							Type:     "remove",
							Path:     "services.mysql.enable",
							OldValue: "true",
							NewValue: "false",
							Reason:   "MariaDB provides better performance and compatibility",
						}},
						Pros: []string{"Better performance", "Active development", "Drop-in replacement"},
						Cons: []string{"Slight syntax differences in advanced features"},
						Risk: "low",
					}},
					Automatic:  false,
					Complexity: "simple",
					Risk:       "medium",
				},
				Evidence: []Evidence{{
					Type:        "rule_violation",
					Description: "MySQL and MariaDB cannot coexist",
					Data:        map[string]string{"conflict": "mysql-mariadb"},
					Confidence:  1.0,
				}},
				Impact:     "Database services will fail to start",
				Confidence: 1.0,
			})
		}
	}

	// Check for conflicting display managers
	displayManagers := []string{"gdm", "lightdm", "sddm", "slim"}
	enabledDMs := id.findEnabledServices(content, displayManagers)
	if len(enabledDMs) > 1 {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        ConflictingServices,
			Severity:    "major",
			Title:       "Multiple Display Managers Enabled",
			Description: fmt.Sprintf("Multiple display managers are enabled: %s. Only one should be active.", strings.Join(enabledDMs, ", ")),
			ConflictingItems: id.createConflictingItems(enabledDMs, "service", "Display managers compete for X11 control"),
			Resolution: &Resolution{
				Type:        "choose_one",
				Description: "Choose one display manager and disable others",
				Options:     id.createDisplayManagerResolutionOptions(enabledDMs),
				Automatic:   false,
				Complexity:  "simple",
				Risk:        "low",
			},
			Evidence: []Evidence{{
				Type:        "pattern_match",
				Description: "Found multiple enabled display managers",
				Data:        map[string]string{"managers": strings.Join(enabledDMs, ", ")},
				Confidence:  0.9,
			}},
			Impact:     "Display manager may fail to start or cause login issues",
			Confidence: 0.9,
		})
	}

	return inconsistencies
}

// detectMissingDependencies detects missing service dependencies
func (id *InconsistencyDetector) detectMissingDependencies(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for web server without SSL/TLS configuration
	if id.serviceEnabled(content, "nginx") && !id.hasSSLConfig(content) {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        MissingDependencies,
			Severity:    "minor",
			Title:       "Nginx Without SSL Configuration",
			Description: "Nginx is enabled but no SSL/TLS configuration was found. Consider adding SSL certificates for security.",
			ConflictingItems: []ConflictingItem{{
				Type:    "service",
				Name:    "nginx",
				Reason:  "Missing SSL configuration for production deployment",
				Context: "Web server security",
			}},
			Resolution: &Resolution{
				Type:        "add_dependency",
				Description: "Add SSL certificate configuration",
				Options: []ResolutionOption{{
					Title:       "Enable Let's Encrypt",
					Description: "Automatically obtain and renew SSL certificates",
					Changes: []ConfigChange{{
						Type:     "add",
						Path:     "security.acme.acceptTerms",
						NewValue: "true",
						Reason:   "Required for Let's Encrypt certificates",
					}, {
						Type:     "add",
						Path:     "security.acme.defaults.email",
						NewValue: "your-email@example.com",
						Reason:   "Contact email for certificate authority",
					}},
					Pros: []string{"Free certificates", "Automatic renewal", "Industry standard"},
					Cons: []string{"Requires internet connectivity", "Rate limits apply"},
					Risk: "low",
				}},
				Automatic:  false,
				Complexity: "moderate",
				Risk:       "low",
			},
			Evidence: []Evidence{{
				Type:        "logical_analysis",
				Description: "Web server without SSL in production is a security risk",
				Data:        map[string]string{"service": "nginx", "ssl": "missing"},
				Confidence:  0.8,
			}},
			Impact:     "Unencrypted web traffic, potential security vulnerabilities",
			Confidence: 0.8,
		})
	}

	// Check for database service without backup configuration
	dbServices := []string{"postgresql", "mysql", "mariadb"}
	for _, db := range dbServices {
		if id.serviceEnabled(content, db) && !id.hasBackupConfig(content, db) {
			inconsistencies = append(inconsistencies, LogicalInconsistency{
				Type:        MissingDependencies,
				Severity:    "minor",
				Title:       fmt.Sprintf("%s Without Backup Configuration", strings.Title(db)),
				Description: fmt.Sprintf("%s is enabled but no backup configuration was found. Database backups are critical for data safety.", strings.Title(db)),
				ConflictingItems: []ConflictingItem{{
					Type:    "service",
					Name:    db,
					Reason:  "Missing backup strategy for data protection",
					Context: "Data safety and disaster recovery",
				}},
				Resolution: &Resolution{
					Type:        "add_dependency",
					Description: "Configure database backups",
					Options:     id.createBackupResolutionOptions(db),
					Automatic:   false,
					Complexity:  "moderate",
					Risk:        "low",
				},
				Evidence: []Evidence{{
					Type:        "logical_analysis",
					Description: "Database without backup strategy poses data loss risk",
					Data:        map[string]string{"database": db, "backup": "missing"},
					Confidence:  0.85,
				}},
				Impact:     "Risk of data loss in case of system failure",
				Confidence: 0.85,
			})
		}
	}

	return inconsistencies
}

// detectResourceConflicts detects resource allocation conflicts
func (id *InconsistencyDetector) detectResourceConflicts(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for port conflicts
	portConflicts := id.findPortConflicts(content)
	for port, services := range portConflicts {
		if len(services) > 1 {
			inconsistencies = append(inconsistencies, LogicalInconsistency{
				Type:        ResourceConflicts,
				Severity:    "major",
				Title:       fmt.Sprintf("Port %s Conflict", port),
				Description: fmt.Sprintf("Multiple services are configured to use port %s: %s", port, strings.Join(services, ", ")),
				ConflictingItems: id.createConflictingItems(services, "service", fmt.Sprintf("All trying to bind to port %s", port)),
				Resolution: &Resolution{
					Type:        "modify",
					Description: "Change port configuration for conflicting services",
					Options:     id.createPortResolutionOptions(port, services),
					Automatic:   false,
					Complexity:  "simple",
					Risk:        "low",
				},
				Evidence: []Evidence{{
					Type:        "logical_analysis",
					Description: "Multiple services cannot bind to the same port",
					Data:        map[string]string{"port": port, "services": strings.Join(services, ", ")},
					Confidence:  0.95,
				}},
				Impact:     "Services will fail to start due to port binding conflicts",
				Confidence: 0.95,
			})
		}
	}

	return inconsistencies
}

// detectSecurityInconsistencies detects security-related logical inconsistencies
func (id *InconsistencyDetector) detectSecurityInconsistencies(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for disabled firewall with exposed services
	if id.firewallDisabled(content) && id.hasExposedServices(content) {
		exposedServices := id.getExposedServices(content)
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        SecurityInconsistency,
			Severity:    "critical",
			Title:       "Firewall Disabled with Exposed Services",
			Description: fmt.Sprintf("Firewall is disabled but services are exposed: %s. This creates a significant security risk.", strings.Join(exposedServices, ", ")),
			ConflictingItems: []ConflictingItem{{
				Type:    "option",
				Name:    "networking.firewall.enable",
				Value:   "false",
				Reason:  "Security risk with exposed services",
				Context: "Network security",
			}},
			Resolution: &Resolution{
				Type:        "modify",
				Description: "Enable firewall and configure appropriate rules",
				Options: []ResolutionOption{{
					Title:       "Enable Firewall with Service Rules",
					Description: "Enable firewall and automatically configure rules for your services",
					Changes: []ConfigChange{{
						Type:     "modify",
						Path:     "networking.firewall.enable",
						OldValue: "false",
						NewValue: "true",
						Reason:   "Essential for network security",
					}},
					Pros: []string{"Improved security", "Controlled access", "Protection against attacks"},
					Cons: []string{"May require manual rule configuration", "Potential access issues"},
					Risk: "low",
				}},
				Automatic:  true,
				Complexity: "simple",
				Risk:       "low",
			},
			Evidence: []Evidence{{
				Type:        "rule_violation",
				Description: "Disabled firewall with exposed services violates security best practices",
				Data:        map[string]string{"firewall": "disabled", "exposed_services": strings.Join(exposedServices, ", ")},
				Confidence:  1.0,
			}},
			Impact:     "System vulnerable to network attacks and unauthorized access",
			Confidence: 1.0,
		})
	}

	// Check for root SSH access without key authentication
	if id.rootSSHEnabled(content) && !id.hasSSHKeys(content) {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        SecurityInconsistency,
			Severity:    "critical",
			Title:       "Root SSH Without Key Authentication",
			Description: "Root SSH login is enabled but no SSH keys are configured. This allows password-based root access.",
			ConflictingItems: []ConflictingItem{{
				Type:    "option",
				Name:    "services.openssh.settings.PermitRootLogin",
				Value:   "yes",
				Reason:  "Security risk without key-based authentication",
				Context: "SSH security",
			}},
			Resolution: &Resolution{
				Type:        "modify",
				Description: "Disable root login or require key authentication",
				Options: []ResolutionOption{{
					Title:       "Disable Root SSH Login",
					Description: "Disable root SSH login and use sudo for administrative tasks",
					Changes: []ConfigChange{{
						Type:     "modify",
						Path:     "services.openssh.settings.PermitRootLogin",
						OldValue: "yes",
						NewValue: "no",
						Reason:   "Eliminate root SSH attack vector",
					}},
					Pros: []string{"Eliminates root attack vector", "Forces use of regular users", "Audit trail with sudo"},
					Cons: []string{"Requires sudo configuration", "May complicate some scripts"},
					Risk: "low",
				}},
				Automatic:  true,
				Complexity: "simple",
				Risk:       "low",
			},
			Evidence: []Evidence{{
				Type:        "rule_violation",
				Description: "Root SSH without key authentication is a critical security vulnerability",
				Data:        map[string]string{"root_ssh": "enabled", "key_auth": "missing"},
				Confidence:  0.95,
			}},
			Impact:     "High risk of unauthorized root access via brute force attacks",
			Confidence: 0.95,
		})
	}

	return inconsistencies
}

// detectNetworkingConflicts detects networking configuration conflicts
func (id *InconsistencyDetector) detectNetworkingConflicts(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for conflicting network managers
	networkManagers := []string{"networkmanager", "wicd", "connman"}
	enabledNMs := id.findEnabledServices(content, networkManagers)
	if len(enabledNMs) > 1 {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        NetworkingConflict,
			Severity:    "major",
			Title:       "Multiple Network Managers Enabled",
			Description: fmt.Sprintf("Multiple network managers are enabled: %s. This can cause networking conflicts.", strings.Join(enabledNMs, ", ")),
			ConflictingItems: id.createConflictingItems(enabledNMs, "service", "Network managers compete for interface control"),
			Resolution: &Resolution{
				Type:        "choose_one",
				Description: "Choose one network manager and disable others",
				Options:     id.createNetworkManagerResolutionOptions(enabledNMs),
				Automatic:   false,
				Complexity:  "simple",
				Risk:        "medium",
			},
			Evidence: []Evidence{{
				Type:        "pattern_match",
				Description: "Found multiple enabled network managers",
				Data:        map[string]string{"managers": strings.Join(enabledNMs, ", ")},
				Confidence:  0.9,
			}},
			Impact:     "Network connectivity issues and interface conflicts",
			Confidence: 0.9,
		})
	}

	// Check for static IP with DHCP enabled
	if id.hasStaticIP(content) && id.dhcpEnabled(content) {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        NetworkingConflict,
			Severity:    "major",
			Title:       "Static IP with DHCP Enabled",
			Description: "Both static IP configuration and DHCP are enabled. This can cause IP address conflicts.",
			ConflictingItems: []ConflictingItem{
				{
					Type:    "option",
					Name:    "networking.interfaces.*.ipv4.addresses",
					Reason:  "Static IP configuration conflicts with DHCP",
					Context: "Network addressing",
				},
				{
					Type:    "option",
					Name:    "networking.useDHCP",
					Value:   "true",
					Reason:  "DHCP conflicts with static IP",
					Context: "Network addressing",
				},
			},
			Resolution: &Resolution{
				Type:        "choose_one",
				Description: "Choose either static IP or DHCP configuration",
				Options: []ResolutionOption{
					{
						Title:       "Use Static IP (Recommended for servers)",
						Description: "Disable DHCP and use static IP configuration",
						Changes: []ConfigChange{{
							Type:     "modify",
							Path:     "networking.useDHCP",
							OldValue: "true",
							NewValue: "false",
							Reason:   "Static IP provides predictable addressing",
						}},
						Pros: []string{"Predictable IP address", "Better for servers", "No DHCP dependency"},
						Cons: []string{"Manual IP management", "Potential conflicts if misconfigured"},
						Risk: "low",
					},
					{
						Title:       "Use DHCP (Recommended for desktops)",
						Description: "Remove static IP configuration and use DHCP",
						Changes: []ConfigChange{{
							Type:     "remove",
							Path:     "networking.interfaces.*.ipv4.addresses",
							Reason:   "DHCP provides automatic IP management",
						}},
						Pros: []string{"Automatic IP management", "No configuration needed", "Network mobility"},
						Cons: []string{"IP address may change", "DHCP dependency"},
						Risk: "low",
					},
				},
				Automatic:  false,
				Complexity: "moderate",
				Risk:       "medium",
			},
			Evidence: []Evidence{{
				Type:        "logical_analysis",
				Description: "Static IP and DHCP configurations are mutually exclusive",
				Data:        map[string]string{"static_ip": "configured", "dhcp": "enabled"},
				Confidence:  0.9,
			}},
			Impact:     "Unpredictable network behavior and potential IP conflicts",
			Confidence: 0.9,
		})
	}

	return inconsistencies
}

// detectFilesystemConflicts detects filesystem configuration conflicts
func (id *InconsistencyDetector) detectFilesystemConflicts(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for conflicting filesystem types on same mount point
	mountConflicts := id.findMountPointConflicts(content)
	for mountPoint, filesystems := range mountConflicts {
		if len(filesystems) > 1 {
			inconsistencies = append(inconsistencies, LogicalInconsistency{
				Type:        FilesystemConflict,
				Severity:    "critical",
				Title:       fmt.Sprintf("Conflicting Filesystems on %s", mountPoint),
				Description: fmt.Sprintf("Multiple filesystem configurations for mount point %s: %s", mountPoint, strings.Join(filesystems, ", ")),
				ConflictingItems: id.createConflictingItems(filesystems, "filesystem", fmt.Sprintf("Multiple filesystems cannot mount to %s", mountPoint)),
				Resolution: &Resolution{
					Type:        "choose_one",
					Description: "Choose one filesystem configuration for the mount point",
					Options:     id.createFilesystemResolutionOptions(mountPoint, filesystems),
					Automatic:   false,
					Complexity:  "complex",
					Risk:        "high",
				},
				Evidence: []Evidence{{
					Type:        "logical_analysis",
					Description: "Multiple filesystems cannot mount to the same point",
					Data:        map[string]string{"mount_point": mountPoint, "filesystems": strings.Join(filesystems, ", ")},
					Confidence:  1.0,
				}},
				Impact:     "System boot failure or filesystem corruption",
				Confidence: 1.0,
			})
		}
	}

	return inconsistencies
}

// detectVersionConflicts detects package version conflicts
func (id *InconsistencyDetector) detectVersionConflicts(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for explicitly conflicting package versions
	versionConflicts := id.findPackageVersionConflicts(content)
	for pkg, versions := range versionConflicts {
		if len(versions) > 1 {
			inconsistencies = append(inconsistencies, LogicalInconsistency{
				Type:        VersionConflict,
				Severity:    "major",
				Title:       fmt.Sprintf("Multiple Versions of %s", pkg),
				Description: fmt.Sprintf("Multiple versions of package %s are specified: %s", pkg, strings.Join(versions, ", ")),
				ConflictingItems: id.createConflictingItems(versions, "package", fmt.Sprintf("Multiple versions of %s cannot coexist", pkg)),
				Resolution: &Resolution{
					Type:        "choose_one",
					Description: "Choose one version of the package",
					Options:     id.createVersionResolutionOptions(pkg, versions),
					Automatic:   false,
					Complexity:  "simple",
					Risk:        "medium",
				},
				Evidence: []Evidence{{
					Type:        "pattern_match",
					Description: "Found multiple version specifications for the same package",
					Data:        map[string]string{"package": pkg, "versions": strings.Join(versions, ", ")},
					Confidence:  0.95,
				}},
				Impact:     "Build conflicts and potential runtime issues",
				Confidence: 0.95,
			})
		}
	}

	return inconsistencies
}

// detectPerformanceConflicts detects performance-related conflicts
func (id *InconsistencyDetector) detectPerformanceConflicts(content string, expr *NixExpression) []LogicalInconsistency {
	var inconsistencies []LogicalInconsistency

	// Check for conflicting swap configurations
	swapTypes := id.findSwapConfigurations(content)
	if len(swapTypes) > 1 {
		inconsistencies = append(inconsistencies, LogicalInconsistency{
			Type:        PerformanceConflict,
			Severity:    "minor",
			Title:       "Multiple Swap Configurations",
			Description: fmt.Sprintf("Multiple swap configurations found: %s. This may cause performance issues.", strings.Join(swapTypes, ", ")),
			ConflictingItems: id.createConflictingItems(swapTypes, "option", "Multiple swap types can interfere with each other"),
			Resolution: &Resolution{
				Type:        "choose_one",
				Description: "Choose the most appropriate swap configuration",
				Options:     id.createSwapResolutionOptions(swapTypes),
				Automatic:   false,
				Complexity:  "simple",
				Risk:        "low",
			},
			Evidence: []Evidence{{
				Type:        "logical_analysis",
				Description: "Multiple swap configurations can cause performance degradation",
				Data:        map[string]string{"swap_types": strings.Join(swapTypes, ", ")},
				Confidence:  0.8,
			}},
			Impact:     "Suboptimal memory management and potential performance degradation",
			Confidence: 0.8,
		})
	}

	return inconsistencies
}

// Helper functions

func (id *InconsistencyDetector) findEnabledServices(content string, services []string) []string {
	var enabled []string
	for _, service := range services {
		if id.serviceEnabled(content, service) {
			enabled = append(enabled, service)
		}
	}
	return enabled
}

func (id *InconsistencyDetector) serviceEnabled(content, service string) bool {
	// Try multiple patterns for different service configurations
	patterns := []string{
		fmt.Sprintf(`services\.%s\.enable\s*=\s*true`, service),
		fmt.Sprintf(`services\.%s\.enable\s*=\s*true;`, service),
		fmt.Sprintf(`services\.\s*%s\s*\.\s*enable\s*=\s*true`, service),
		// Handle display manager patterns
		fmt.Sprintf(`services\.xserver\.displayManager\.%s\.enable\s*=\s*true`, service),
		// Handle networking patterns
		fmt.Sprintf(`networking\.%s\.enable\s*=\s*true`, service),
	}
	
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}

func (id *InconsistencyDetector) hasSSLConfig(content string) bool {
	sslPatterns := []string{
		`ssl_certificate`,
		`security\.acme`,
		`enableACME\s*=\s*true`,
		`forceSSL\s*=\s*true`,
	}
	for _, pattern := range sslPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}

func (id *InconsistencyDetector) hasBackupConfig(content, dbService string) bool {
	backupPatterns := []string{
		fmt.Sprintf(`services\.%s\.backup`, dbService),
		fmt.Sprintf(`services\.%s\..*dump`, dbService),
		`systemd\.services\..*backup`,
		`programs\.restic`,
		`services\.borgbackup`,
	}
	for _, pattern := range backupPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}

func (id *InconsistencyDetector) firewallDisabled(content string) bool {
	pattern := `networking\.firewall\.enable\s*=\s*false`
	matched, _ := regexp.MatchString(pattern, content)
	return matched
}

func (id *InconsistencyDetector) hasExposedServices(content string) bool {
	exposedServices := []string{"nginx", "apache", "openssh", "postgresql", "mysql", "mariadb"}
	for _, service := range exposedServices {
		if id.serviceEnabled(content, service) {
			return true
		}
	}
	return false
}

func (id *InconsistencyDetector) getExposedServices(content string) []string {
	exposedServices := []string{"nginx", "apache", "openssh", "postgresql", "mysql", "mariadb"}
	var found []string
	for _, service := range exposedServices {
		if id.serviceEnabled(content, service) {
			found = append(found, service)
		}
	}
	return found
}

func (id *InconsistencyDetector) rootSSHEnabled(content string) bool {
	pattern := `PermitRootLogin\s*=\s*"?yes"?`
	matched, _ := regexp.MatchString(pattern, content)
	return matched
}

func (id *InconsistencyDetector) hasSSHKeys(content string) bool {
	patterns := []string{
		`openssh\.authorizedKeys`,
		`users\.users\..*\.openssh\.authorizedKeys`,
		`AuthorizedKeysFile`,
	}
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}

func (id *InconsistencyDetector) hasStaticIP(content string) bool {
	pattern := `networking\.interfaces\..*\.ipv4\.addresses`
	matched, _ := regexp.MatchString(pattern, content)
	return matched
}

func (id *InconsistencyDetector) dhcpEnabled(content string) bool {
	pattern := `networking\.useDHCP\s*=\s*true`
	matched, _ := regexp.MatchString(pattern, content)
	return matched
}

func (id *InconsistencyDetector) findPortConflicts(content string) map[string][]string {
	portMap := make(map[string][]string)
	
	// Common port mappings for services
	servicePorts := map[string]string{
		"nginx":      "80",
		"apache":     "80", 
		"lighttpd":   "80",
		"caddy":      "80",
		"openssh":    "22",
		"postgresql": "5432",
		"mysql":      "3306",
		"mariadb":    "3306",
		"mongodb":    "27017",
	}

	// Check for enabled services and their default ports
	for service, port := range servicePorts {
		if id.serviceEnabled(content, service) {
			portMap[port] = append(portMap[port], service)
		}
	}

	// Look for explicit port configurations
	portRegex := regexp.MustCompile(`port\s*=\s*(\d+)`)
	matches := portRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			port := match[1]
			portMap[port] = append(portMap[port], "custom-service")
		}
	}

	return portMap
}

func (id *InconsistencyDetector) findMountPointConflicts(content string) map[string][]string {
	mountMap := make(map[string][]string)
	
	// Look for fileSystems configurations
	fsRegex := regexp.MustCompile(`fileSystems\."([^"]+)"\s*=\s*{[^}]*fsType\s*=\s*"([^"]+)"`)
	matches := fsRegex.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches {
		if len(match) > 2 {
			mountPoint := match[1]
			fsType := match[2]
			mountMap[mountPoint] = append(mountMap[mountPoint], fsType)
		}
	}
	
	return mountMap
}

func (id *InconsistencyDetector) findPackageVersionConflicts(content string) map[string][]string {
	versionMap := make(map[string][]string)
	
	// Look for versioned package references - pattern 1: pkgs.package_version
	versionRegex1 := regexp.MustCompile(`pkgs\.([a-zA-Z0-9_-]+)_(\d+(?:\.\d+)*)`)
	matches1 := versionRegex1.FindAllStringSubmatch(content, -1)
	
	for _, match := range matches1 {
		if len(match) > 2 {
			pkg := match[1]
			version := match[2]
			versionMap[pkg] = append(versionMap[pkg], version)
		}
	}
	
	// Look for versioned package references - pattern 2: packagename+version (e.g., python39, python310)
	versionRegex2 := regexp.MustCompile(`\b([a-zA-Z]+)(\d+)(?:_(\d+))*\b`)
	matches2 := versionRegex2.FindAllStringSubmatch(content, -1)
	
	packageVersions := make(map[string]map[string]bool)
	for _, match := range matches2 {
		if len(match) >= 3 {
			basePkg := match[1]
			version := match[2]
			if len(match) > 3 && match[3] != "" {
				version += "." + match[3]
			}
			
			// Only consider certain packages that commonly have version suffixes
			if basePkg == "python" || basePkg == "node" || basePkg == "java" || basePkg == "gcc" || basePkg == "llvm" {
				if packageVersions[basePkg] == nil {
					packageVersions[basePkg] = make(map[string]bool)
				}
				packageVersions[basePkg][version] = true
			}
		}
	}
	
	// Convert to the expected format
	for pkg, versions := range packageVersions {
		if len(versions) > 1 {
			for version := range versions {
				versionMap[pkg] = append(versionMap[pkg], version)
			}
		}
	}
	
	return versionMap
}

func (id *InconsistencyDetector) findSwapConfigurations(content string) []string {
	var swapTypes []string
	
	if matched, _ := regexp.MatchString(`swapDevices`, content); matched {
		swapTypes = append(swapTypes, "swapDevices")
	}
	
	if matched, _ := regexp.MatchString(`zramSwap\.enable\s*=\s*true`, content); matched {
		swapTypes = append(swapTypes, "zramSwap")
	}
	
	if matched, _ := regexp.MatchString(`boot\.kernel\.sysctl\."vm\.swappiness"`, content); matched {
		swapTypes = append(swapTypes, "custom-swap-tuning")
	}
	
	return swapTypes
}

func (id *InconsistencyDetector) containsAll(slice, items []string) bool {
	itemMap := make(map[string]bool)
	for _, item := range slice {
		itemMap[item] = true
	}
	
	for _, item := range items {
		if !itemMap[item] {
			return false
		}
	}
	return true
}

func (id *InconsistencyDetector) createConflictingItems(items []string, itemType, reason string) []ConflictingItem {
	var conflictingItems []ConflictingItem
	for _, item := range items {
		conflictingItems = append(conflictingItems, ConflictingItem{
			Type:    itemType,
			Name:    item,
			Reason:  reason,
			Context: "Configuration conflict",
		})
	}
	return conflictingItems
}

// Resolution option generators

func (id *InconsistencyDetector) createWebServerResolutionOptions(servers []string) []ResolutionOption {
	var options []ResolutionOption
	
	for _, server := range servers {
		var description, pros, cons string
		switch server {
		case "nginx":
			description = "High-performance web server, excellent for static content and reverse proxy"
			pros = "High performance, low memory usage, excellent reverse proxy"
			cons = "Configuration syntax can be complex"
		case "apache":
			description = "Mature web server with extensive module ecosystem"
			pros = "Mature, extensive modules, familiar configuration"
			cons = "Higher memory usage, can be slower than nginx"
		case "caddy":
			description = "Modern web server with automatic HTTPS"
			pros = "Automatic HTTPS, simple configuration, modern features"
			cons = "Newer project, smaller ecosystem"
		default:
			description = fmt.Sprintf("Keep %s as the web server", server)
			pros = "Maintains current configuration"
			cons = "May not be optimal choice"
		}
		
		var changes []ConfigChange
		for _, otherServer := range servers {
			if otherServer != server {
				changes = append(changes, ConfigChange{
					Type:     "modify",
					Path:     fmt.Sprintf("services.%s.enable", otherServer),
					OldValue: "true",
					NewValue: "false",
					Reason:   "Prevent port conflicts",
				})
			}
		}
		
		options = append(options, ResolutionOption{
			Title:       fmt.Sprintf("Use %s", strings.Title(server)),
			Description: description,
			Changes:     changes,
			Pros:        []string{pros},
			Cons:        []string{cons},
			Risk:        "low",
		})
	}
	
	return options
}

func (id *InconsistencyDetector) createDatabaseResolutionOptions(databases []string) []ResolutionOption {
	var options []ResolutionOption
	
	for _, db := range databases {
		var description, pros string
		switch db {
		case "postgresql":
			description = "PostgreSQL - powerful, standards-compliant SQL database"
			pros = "ACID compliance, powerful features, excellent performance, strong community"
		case "mysql":
			description = "MySQL - popular relational database"
			pros = "Widely used, good performance, large ecosystem"
		case "mariadb":
			description = "MariaDB - MySQL-compatible database with improvements"
			pros = "Drop-in MySQL replacement, active development, better performance"
		case "mongodb":
			description = "MongoDB - document-oriented NoSQL database"
			pros = "Flexible schema, horizontal scaling, good for JSON-like data"
		default:
			description = fmt.Sprintf("Keep %s as the database server", db)
			pros = "Maintains current configuration"
		}
		
		var changes []ConfigChange
		for _, otherDb := range databases {
			if otherDb != db {
				changes = append(changes, ConfigChange{
					Type:     "modify",
					Path:     fmt.Sprintf("services.%s.enable", otherDb),
					OldValue: "true",
					NewValue: "false",
					Reason:   "Prevent database conflicts",
				})
			}
		}
		
		options = append(options, ResolutionOption{
			Title:       fmt.Sprintf("Use %s", strings.Title(db)),
			Description: description,
			Changes:     changes,
			Pros:        []string{pros},
			Cons:        []string{"May require data migration from other databases"},
			Risk:        "medium",
		})
	}
	
	return options
}

func (id *InconsistencyDetector) createDisplayManagerResolutionOptions(dms []string) []ResolutionOption {
	var options []ResolutionOption
	
	for _, dm := range dms {
		var description, pros string
		switch dm {
		case "gdm":
			description = "GNOME Display Manager - well integrated with GNOME desktop"
			pros = "Excellent GNOME integration, modern features, accessibility support"
		case "lightdm":
			description = "Lightweight Display Manager - works well with most desktop environments"
			pros = "Lightweight, flexible, good compatibility, customizable greeters"
		case "sddm":
			description = "Simple Desktop Display Manager - KDE's preferred display manager"
			pros = "Excellent KDE integration, QML-based themes, modern architecture"
		default:
			description = fmt.Sprintf("Keep %s as the display manager", dm)
			pros = "Maintains current configuration"
		}
		
		var changes []ConfigChange
		for _, otherDM := range dms {
			if otherDM != dm {
				changes = append(changes, ConfigChange{
					Type:     "modify",
					Path:     fmt.Sprintf("services.xserver.displayManager.%s.enable", otherDM),
					OldValue: "true",
					NewValue: "false",
					Reason:   "Only one display manager should be active",
				})
			}
		}
		
		options = append(options, ResolutionOption{
			Title:       fmt.Sprintf("Use %s", strings.ToUpper(dm)),
			Description: description,
			Changes:     changes,
			Pros:        []string{pros},
			Cons:        []string{"May require desktop environment reconfiguration"},
			Risk:        "low",
		})
	}
	
	return options
}

func (id *InconsistencyDetector) createBackupResolutionOptions(dbService string) []ResolutionOption {
	switch dbService {
	case "postgresql":
		return []ResolutionOption{{
			Title:       "Enable PostgreSQL Backup",
			Description: "Configure automated PostgreSQL database backups",
			Changes: []ConfigChange{{
				Type:     "add",
				Path:     "services.postgresql.backup.enable",
				NewValue: "true",
				Reason:   "Enable automatic database backups",
			}, {
				Type:     "add",
				Path:     "services.postgresql.backup.location",
				NewValue: "/var/backup/postgresql",
				Reason:   "Specify backup storage location",
			}},
			Pros: []string{"Automated backups", "Point-in-time recovery", "Data protection"},
			Cons: []string{"Requires disk space", "May impact performance during backup"},
			Risk: "low",
		}}
	default:
		return []ResolutionOption{{
			Title:       fmt.Sprintf("Configure %s Backup", strings.Title(dbService)),
			Description: "Set up backup strategy for the database",
			Changes: []ConfigChange{{
				Type:     "add",
				Path:     fmt.Sprintf("systemd.services.%s-backup", dbService),
				NewValue: "{ enable = true; }",
				Reason:   "Create backup service",
			}},
			Pros: []string{"Data protection", "Disaster recovery"},
			Cons: []string{"Requires manual configuration", "Storage requirements"},
			Risk: "low",
		}}
	}
}

func (id *InconsistencyDetector) createPortResolutionOptions(port string, services []string) []ResolutionOption {
	var options []ResolutionOption
	
	if port == "80" || port == "443" {
		// For web services, suggest using nginx as reverse proxy
		options = append(options, ResolutionOption{
			Title:       "Use Nginx as Reverse Proxy",
			Description: "Configure nginx to proxy requests to other web services on different ports",
			Changes: []ConfigChange{{
				Type:     "add",
				Path:     "services.nginx.virtualHosts",
				NewValue: "{ /* reverse proxy configuration */ }",
				Reason:   "Enable multiple web services through reverse proxy",
			}},
			Pros: []string{"Multiple services on same port", "Centralized SSL termination", "Load balancing"},
			Cons: []string{"Additional complexity", "Single point of failure"},
			Risk: "medium",
		})
	}
	
	// Generic option to change ports
	portNum, _ := strconv.Atoi(port)
	newPort := portNum + 1000
	
	options = append(options, ResolutionOption{
		Title:       "Change Service Ports",
		Description: fmt.Sprintf("Reconfigure one or more services to use alternative ports (e.g., %d)", newPort),
		Changes: []ConfigChange{{
			Type:     "modify",
			Path:     fmt.Sprintf("services.*.port"),
			OldValue: port,
			NewValue: fmt.Sprintf("%d", newPort),
			Reason:   "Avoid port conflicts",
		}},
		Pros: []string{"Simple solution", "Maintains all services", "No proxy complexity"},
		Cons: []string{"Non-standard ports", "May require firewall updates", "User confusion"},
		Risk: "low",
	})
	
	return options
}

func (id *InconsistencyDetector) createNetworkManagerResolutionOptions(nms []string) []ResolutionOption {
	var options []ResolutionOption
	
	for _, nm := range nms {
		var description, pros string
		switch nm {
		case "networkmanager":
			description = "NetworkManager - most common choice for desktop systems"
			pros = "Excellent desktop integration, WiFi support, VPN support, GUI tools"
		case "wicd":
			description = "Wicd - lightweight network manager"
			pros = "Lightweight, simple interface, good for basic networking"
		case "connman":
			description = "ConnMan - embedded-focused network manager"
			pros = "Low resource usage, good for embedded systems, simple API"
		default:
			description = fmt.Sprintf("Keep %s as the network manager", nm)
			pros = "Maintains current configuration"
		}
		
		var changes []ConfigChange
		for _, otherNM := range nms {
			if otherNM != nm {
				changes = append(changes, ConfigChange{
					Type:     "modify",
					Path:     fmt.Sprintf("networking.%s.enable", otherNM),
					OldValue: "true",
					NewValue: "false",
					Reason:   "Prevent network manager conflicts",
				})
			}
		}
		
		options = append(options, ResolutionOption{
			Title:       fmt.Sprintf("Use %s", strings.Title(nm)),
			Description: description,
			Changes:     changes,
			Pros:        []string{pros},
			Cons:        []string{"May require network reconfiguration"},
			Risk:        "medium",
		})
	}
	
	return options
}

func (id *InconsistencyDetector) createFilesystemResolutionOptions(mountPoint string, filesystems []string) []ResolutionOption {
	var options []ResolutionOption
	
	for _, fs := range filesystems {
		var description, pros, cons string
		switch fs {
		case "ext4":
			description = "Traditional Linux filesystem with good stability"
			pros = "Stable, widely supported, good performance"
			cons = "No advanced features like snapshots or compression"
		case "btrfs":
			description = "Modern filesystem with snapshots and compression"
			pros = "Snapshots, compression, subvolumes, data integrity"
			cons = "More complex, potential stability issues with some features"
		case "zfs":
			description = "Advanced filesystem with enterprise features"
			pros = "Excellent data integrity, snapshots, compression, RAID"
			cons = "High memory usage, license compatibility issues"
		case "xfs":
			description = "High-performance filesystem for large files"
			pros = "Excellent performance for large files, scalable"
			cons = "Cannot shrink filesystems, less common"
		default:
			description = fmt.Sprintf("Use %s filesystem", fs)
			pros = "Current configuration"
			cons = "May not be optimal"
		}
		
		var changes []ConfigChange
		for _, otherFs := range filesystems {
			if otherFs != fs {
				changes = append(changes, ConfigChange{
					Type:     "remove",
					Path:     fmt.Sprintf("fileSystems.\"%s\"", mountPoint),
					Reason:   "Remove conflicting filesystem configuration",
				})
			}
		}
		
		// Add the chosen filesystem back
		changes = append(changes, ConfigChange{
			Type:     "add",
			Path:     fmt.Sprintf("fileSystems.\"%s\"", mountPoint),
			NewValue: fmt.Sprintf("{ fsType = \"%s\"; }", fs),
			Reason:   "Configure chosen filesystem",
		})
		
		options = append(options, ResolutionOption{
			Title:       fmt.Sprintf("Use %s", strings.ToUpper(fs)),
			Description: description,
			Changes:     changes,
			Pros:        []string{pros},
			Cons:        []string{cons},
			Risk:        "high",
		})
	}
	
	return options
}

func (id *InconsistencyDetector) createVersionResolutionOptions(pkg string, versions []string) []ResolutionOption {
	var options []ResolutionOption
	
	for _, version := range versions {
		options = append(options, ResolutionOption{
			Title:       fmt.Sprintf("Use %s version %s", pkg, version),
			Description: fmt.Sprintf("Keep %s at version %s and remove other versions", pkg, version),
			Changes: []ConfigChange{{
				Type:     "modify",
				Path:     fmt.Sprintf("environment.systemPackages"),
				NewValue: fmt.Sprintf("pkgs.%s_%s", pkg, strings.ReplaceAll(version, ".", "_")),
				Reason:   "Use single package version",
			}},
			Pros: []string{"Consistent package version", "Avoids conflicts"},
			Cons: []string{"May break compatibility with other packages"},
			Risk: "medium",
		})
	}
	
	// Add option to use latest version
	options = append(options, ResolutionOption{
		Title:       fmt.Sprintf("Use latest %s", pkg),
		Description: fmt.Sprintf("Use the latest available version of %s", pkg),
		Changes: []ConfigChange{{
			Type:     "modify",
			Path:     fmt.Sprintf("environment.systemPackages"),
			NewValue: fmt.Sprintf("pkgs.%s", pkg),
			Reason:   "Use latest package version",
		}},
		Pros: []string{"Latest features and bug fixes", "Better security"},
		Cons: []string{"Potential compatibility issues", "May introduce new bugs"},
		Risk: "low",
	})
	
	return options
}

func (id *InconsistencyDetector) createSwapResolutionOptions(swapTypes []string) []ResolutionOption {
	var options []ResolutionOption
	
	if id.contains(swapTypes, "zramSwap") {
		options = append(options, ResolutionOption{
			Title:       "Use ZRAM Swap Only",
			Description: "Use compressed memory swap (ZRAM) for better performance",
			Changes: []ConfigChange{
				{
					Type:     "modify",
					Path:     "zramSwap.enable",
					NewValue: "true",
					Reason:   "Enable ZRAM compressed swap",
				},
				{
					Type:     "remove",
					Path:     "swapDevices",
					Reason:   "Remove traditional swap devices",
				},
			},
			Pros: []string{"Better performance", "No disk wear", "Compressed memory"},
			Cons: []string{"Uses RAM for swap", "Limited by available memory"},
			Risk: "low",
		})
	}
	
	if id.contains(swapTypes, "swapDevices") {
		options = append(options, ResolutionOption{
			Title:       "Use Traditional Swap Only",
			Description: "Use disk-based swap devices for maximum swap space",
			Changes: []ConfigChange{
				{
					Type:     "modify",
					Path:     "zramSwap.enable",
					OldValue: "true",
					NewValue: "false",
					Reason:   "Disable ZRAM to avoid conflicts",
				},
			},
			Pros: []string{"Large swap space", "Persistent across reboots", "Well tested"},
			Cons: []string{"Slower than ZRAM", "Disk wear", "I/O overhead"},
			Risk: "low",
		})
	}
	
	return options
}

func (id *InconsistencyDetector) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Summary and statistics generation

func (id *InconsistencyDetector) generateSummary(inconsistencies []LogicalInconsistency) InconsistencySummary {
	summary := InconsistencySummary{
		TotalInconsistencies: len(inconsistencies),
		BySeverity:          make(map[string]int),
		ByType:              make(map[string]int),
		CriticalIssues:      []string{},
		RecommendedActions:  []string{},
	}
	
	severityScores := map[string]int{"critical": 4, "major": 3, "minor": 2, "info": 1}
	totalSeverityScore := 0
	
	for _, inconsistency := range inconsistencies {
		// Count by severity
		summary.BySeverity[inconsistency.Severity]++
		
		// Count by type
		summary.ByType[string(inconsistency.Type)]++
		
		// Collect critical issues
		if inconsistency.Severity == "critical" {
			summary.CriticalIssues = append(summary.CriticalIssues, inconsistency.Title)
		}
		
		// Collect recommended actions
		if inconsistency.Resolution != nil && inconsistency.Resolution.Automatic {
			summary.RecommendedActions = append(summary.RecommendedActions, 
				fmt.Sprintf("Auto-fix: %s", inconsistency.Resolution.Description))
		}
		
		// Calculate overall risk
		if score, ok := severityScores[inconsistency.Severity]; ok {
			totalSeverityScore += score
		}
	}
	
	// Determine overall risk
	if len(inconsistencies) == 0 {
		summary.OverallRisk = "none"
	} else {
		avgSeverity := float64(totalSeverityScore) / float64(len(inconsistencies))
		switch {
		case avgSeverity >= 3.5:
			summary.OverallRisk = "critical"
		case avgSeverity >= 2.5:
			summary.OverallRisk = "high"
		case avgSeverity >= 1.5:
			summary.OverallRisk = "medium"
		default:
			summary.OverallRisk = "low"
		}
	}
	
	return summary
}

func (id *InconsistencyDetector) generateStatistics(inconsistencies []LogicalInconsistency) InconsistencyStats {
	stats := InconsistencyStats{}
	
	if len(inconsistencies) == 0 {
		stats.ConfigurationHealth = "excellent"
		return stats
	}
	
	var totalConfidence float64
	typeCount := make(map[string]int)
	severityCount := make(map[string]int)
	
	for _, inconsistency := range inconsistencies {
		// Count auto-resolvable
		if inconsistency.Resolution != nil && inconsistency.Resolution.Automatic {
			stats.AutoResolvable++
		} else {
			stats.ManualResolution++
		}
		
		// Sum confidence for average
		totalConfidence += inconsistency.Confidence
		
		// Count types
		typeCount[string(inconsistency.Type)]++
		
		// Count severities
		severityCount[inconsistency.Severity]++
	}
	
	// Calculate average confidence
	stats.AverageConfidence = totalConfidence / float64(len(inconsistencies))
	
	// Find most common type
	maxCount := 0
	for t, count := range typeCount {
		if count > maxCount {
			maxCount = count
			stats.MostCommonType = t
		}
	}
	
	// Determine configuration health
	criticalCount := severityCount["critical"]
	majorCount := severityCount["major"]
	
	switch {
	case criticalCount > 0:
		stats.ConfigurationHealth = "poor"
	case majorCount > 2:
		stats.ConfigurationHealth = "poor"
	case majorCount > 0 || len(inconsistencies) > 5:
		stats.ConfigurationHealth = "fair"
	case len(inconsistencies) > 2:
		stats.ConfigurationHealth = "good"
	default:
		stats.ConfigurationHealth = "excellent"
	}
	
	return stats
}