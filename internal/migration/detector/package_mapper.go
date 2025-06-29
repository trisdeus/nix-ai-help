package detector

import (
	"fmt"
	"strings"

	"nix-ai-help/pkg/logger"
)

// PackageMapping represents a package mapping from source to nixpkgs
type PackageMapping struct {
	SourceName    string   `json:"source_name"`
	NixpkgsName   string   `json:"nixpkgs_name"`
	NixpkgsPath   string   `json:"nixpkgs_path"`
	Alternatives  []string `json:"alternatives"`
	Notes         string   `json:"notes"`
	Category      string   `json:"category"`
	Availability  string   `json:"availability"` // available, unstable, missing
	SystemPackage bool     `json:"system_package"`
}

// PackageMapper maps packages from source systems to nixpkgs equivalents
type PackageMapper struct {
	logger   logger.Logger
	mappings map[string]PackageMapping
}

// NewPackageMapper creates a new package mapper
func NewPackageMapper(logger logger.Logger) *PackageMapper {
	pm := &PackageMapper{
		logger:   logger,
		mappings: make(map[string]PackageMapping),
	}
	pm.initializeMappings()
	return pm
}

// initializeMappings initializes the package mapping database
func (pm *PackageMapper) initializeMappings() {
	mappings := []PackageMapping{
		// System Tools
		{
			SourceName:    "curl",
			NixpkgsName:   "curl",
			NixpkgsPath:   "pkgs.curl",
			Category:      "networking",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "wget",
			NixpkgsName:   "wget",
			NixpkgsPath:   "pkgs.wget",
			Category:      "networking",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "git",
			NixpkgsName:   "git",
			NixpkgsPath:   "pkgs.git",
			Category:      "development",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "vim",
			NixpkgsName:   "vim",
			NixpkgsPath:   "pkgs.vim",
			Alternatives:  []string{"neovim", "nano", "emacs"},
			Category:      "editors",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "nano",
			NixpkgsName:   "nano",
			NixpkgsPath:   "pkgs.nano",
			Category:      "editors",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "htop",
			NixpkgsName:   "htop",
			NixpkgsPath:   "pkgs.htop",
			Alternatives:  []string{"top", "btop", "glances"},
			Category:      "monitoring",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "tree",
			NixpkgsName:   "tree",
			NixpkgsPath:   "pkgs.tree",
			Category:      "utilities",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "unzip",
			NixpkgsName:   "unzip",
			NixpkgsPath:   "pkgs.unzip",
			Category:      "compression",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "zip",
			NixpkgsName:   "zip",
			NixpkgsPath:   "pkgs.zip",
			Category:      "compression",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "jq",
			NixpkgsName:   "jq",
			NixpkgsPath:   "pkgs.jq",
			Category:      "utilities",
			Availability:  "available",
			SystemPackage: true,
		},

		// Development Tools
		{
			SourceName:    "gcc",
			NixpkgsName:   "gcc",
			NixpkgsPath:   "pkgs.gcc",
			Category:      "development",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "make",
			NixpkgsName:   "gnumake",
			NixpkgsPath:   "pkgs.gnumake",
			Category:      "development",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:    "cmake",
			NixpkgsName:   "cmake",
			NixpkgsPath:   "pkgs.cmake",
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "nodejs",
			NixpkgsName:   "nodejs",
			NixpkgsPath:   "pkgs.nodejs",
			Alternatives:  []string{"nodejs_18", "nodejs_20"},
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "npm",
			NixpkgsName:   "nodejs", // npm comes with nodejs
			NixpkgsPath:   "pkgs.nodejs",
			Notes:         "npm is included with nodejs package",
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "python3",
			NixpkgsName:   "python3",
			NixpkgsPath:   "pkgs.python3",
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "python3-pip",
			NixpkgsName:   "python3", // pip is included
			NixpkgsPath:   "pkgs.python3",
			Notes:         "pip is included with python3 package",
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "go",
			NixpkgsName:   "go",
			NixpkgsPath:   "pkgs.go",
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "rust",
			NixpkgsName:   "rustc",
			NixpkgsPath:   "pkgs.rustc",
			Alternatives:  []string{"cargo", "rust-bin.stable.latest.default"},
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "cargo",
			NixpkgsName:   "cargo",
			NixpkgsPath:   "pkgs.cargo",
			Category:      "development",
			Availability:  "available",
			SystemPackage: false,
		},

		// Web Servers
		{
			SourceName:    "nginx",
			NixpkgsName:   "nginx",
			NixpkgsPath:   "pkgs.nginx",
			Category:      "servers",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.nginx",
		},
		{
			SourceName:    "apache2",
			NixpkgsName:   "apache-httpd",
			NixpkgsPath:   "pkgs.apache-httpd",
			Category:      "servers",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.httpd",
		},

		// Databases
		{
			SourceName:    "mysql-server",
			NixpkgsName:   "mysql80",
			NixpkgsPath:   "pkgs.mysql80",
			Alternatives:  []string{"mariadb", "mysql57"},
			Category:      "databases",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.mysql",
		},
		{
			SourceName:    "mariadb-server",
			NixpkgsName:   "mariadb",
			NixpkgsPath:   "pkgs.mariadb",
			Category:      "databases",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.mysql",
		},
		{
			SourceName:    "postgresql",
			NixpkgsName:   "postgresql",
			NixpkgsPath:   "pkgs.postgresql",
			Alternatives:  []string{"postgresql_13", "postgresql_14", "postgresql_15"},
			Category:      "databases",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.postgresql",
		},
		{
			SourceName:    "redis-server",
			NixpkgsName:   "redis",
			NixpkgsPath:   "pkgs.redis",
			Category:      "databases",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.redis",
		},

		// Container Tools
		{
			SourceName:    "docker.io",
			NixpkgsName:   "docker",
			NixpkgsPath:   "pkgs.docker",
			Category:      "containers",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via virtualisation.docker",
		},
		{
			SourceName:    "docker-ce",
			NixpkgsName:   "docker",
			NixpkgsPath:   "pkgs.docker",
			Category:      "containers",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via virtualisation.docker",
		},
		{
			SourceName:    "podman",
			NixpkgsName:   "podman",
			NixpkgsPath:   "pkgs.podman",
			Category:      "containers",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via virtualisation.podman",
		},

		// Monitoring Tools
		{
			SourceName:    "prometheus",
			NixpkgsName:   "prometheus",
			NixpkgsPath:   "pkgs.prometheus",
			Category:      "monitoring",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.prometheus",
		},
		{
			SourceName:    "grafana",
			NixpkgsName:   "grafana",
			NixpkgsPath:   "pkgs.grafana",
			Category:      "monitoring",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.grafana",
		},

		// Security Tools
		{
			SourceName:    "ufw",
			NixpkgsName:   "ufw",
			NixpkgsPath:   "pkgs.ufw",
			Category:      "security",
			Availability:  "available",
			SystemPackage: true,
			Notes:         "Consider using networking.firewall instead",
		},
		{
			SourceName:    "fail2ban",
			NixpkgsName:   "fail2ban",
			NixpkgsPath:   "pkgs.fail2ban",
			Category:      "security",
			Availability:  "available",
			SystemPackage: false,
			Notes:         "Usually configured via services.fail2ban",
		},

		// Text Editors and IDEs
		{
			SourceName:    "code",
			NixpkgsName:   "vscode",
			NixpkgsPath:   "pkgs.vscode",
			Alternatives:  []string{"vscodium", "vscode-fhs"},
			Category:      "editors",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "emacs",
			NixpkgsName:   "emacs",
			NixpkgsPath:   "pkgs.emacs",
			Category:      "editors",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "neovim",
			NixpkgsName:   "neovim",
			NixpkgsPath:   "pkgs.neovim",
			Category:      "editors",
			Availability:  "available",
			SystemPackage: false,
		},

		// Media Tools
		{
			SourceName:    "ffmpeg",
			NixpkgsName:   "ffmpeg",
			NixpkgsPath:   "pkgs.ffmpeg",
			Category:      "media",
			Availability:  "available",
			SystemPackage: false,
		},
		{
			SourceName:    "vlc",
			NixpkgsName:   "vlc",
			NixpkgsPath:   "pkgs.vlc",
			Category:      "media",
			Availability:  "available",
			SystemPackage: false,
		},

		// Common Debian/Ubuntu specific packages
		{
			SourceName:   "apt-transport-https",
			NixpkgsName:  "", // Not needed in NixOS
			NixpkgsPath:  "",
			Category:     "system",
			Availability: "not_needed",
			Notes:        "Not required in NixOS - HTTPS transport is built-in",
		},
		{
			SourceName:    "ca-certificates",
			NixpkgsName:   "cacert",
			NixpkgsPath:   "pkgs.cacert",
			Category:      "security",
			Availability:  "available",
			SystemPackage: true,
			Notes:         "Usually included by default in NixOS",
		},
		{
			SourceName:   "software-properties-common",
			NixpkgsName:  "", // Not needed
			NixpkgsPath:  "",
			Category:     "system",
			Availability: "not_needed",
			Notes:        "Not required in NixOS - repository management is different",
		},
		{
			SourceName:    "gnupg",
			NixpkgsName:   "gnupg",
			NixpkgsPath:   "pkgs.gnupg",
			Category:      "security",
			Availability:  "available",
			SystemPackage: true,
		},
		{
			SourceName:   "lsb-release",
			NixpkgsName:  "", // Not needed
			NixpkgsPath:  "",
			Category:     "system",
			Availability: "not_needed",
			Notes:        "Not required in NixOS - use nixos-version instead",
		},
	}

	for _, mapping := range mappings {
		pm.mappings[mapping.SourceName] = mapping
	}
}

// MapPackage maps a source package to nixpkgs equivalent
func (pm *PackageMapper) MapPackage(packageName string) (PackageMapping, bool) {
	// Clean package name (remove version suffixes, etc.)
	cleanName := pm.cleanPackageName(packageName)

	// Direct mapping
	if mapping, exists := pm.mappings[cleanName]; exists {
		return mapping, true
	}

	// Try fuzzy matching
	cleanName = strings.ToLower(cleanName)
	for name, mapping := range pm.mappings {
		if strings.Contains(strings.ToLower(name), cleanName) ||
			strings.Contains(cleanName, strings.ToLower(name)) {
			return mapping, true
		}
	}

	return PackageMapping{}, false
}

// MapPackages maps multiple packages to nixpkgs equivalents
func (pm *PackageMapper) MapPackages(packages []string) map[string]PackageMapping {
	result := make(map[string]PackageMapping)

	for _, pkg := range packages {
		if mapping, exists := pm.MapPackage(pkg); exists {
			result[pkg] = mapping
		}
	}

	return result
}

// cleanPackageName removes version suffixes and cleans package names
func (pm *PackageMapper) cleanPackageName(packageName string) string {
	// Remove common suffixes
	suffixes := []string{":amd64", ":i386", "-dev", "-devel", "-doc", "-dbg"}
	cleaned := packageName

	for _, suffix := range suffixes {
		if strings.HasSuffix(cleaned, suffix) {
			cleaned = strings.TrimSuffix(cleaned, suffix)
		}
	}

	// Remove version numbers (simplified)
	parts := strings.Fields(cleaned)
	if len(parts) > 0 {
		cleaned = parts[0]
	}

	return cleaned
}

// GetSystemPackages returns packages that should be in environment.systemPackages
func (pm *PackageMapper) GetSystemPackages(mappings map[string]PackageMapping) []string {
	var systemPkgs []string

	for _, mapping := range mappings {
		if mapping.SystemPackage && mapping.Availability == "available" {
			systemPkgs = append(systemPkgs, mapping.NixpkgsPath)
		}
	}

	return systemPkgs
}

// GetUnavailablePackages returns packages that are not available in nixpkgs
func (pm *PackageMapper) GetUnavailablePackages(packages []string) []string {
	var unavailable []string

	for _, pkg := range packages {
		if mapping, exists := pm.MapPackage(pkg); !exists || mapping.Availability == "missing" {
			unavailable = append(unavailable, pkg)
		}
	}

	return unavailable
}

// GenerateNixOSPackageConfig generates NixOS package configuration
func (pm *PackageMapper) GenerateNixOSPackageConfig(packages []string) string {
	mappings := pm.MapPackages(packages)

	var config strings.Builder

	config.WriteString("  # System packages\n")
	config.WriteString("  environment.systemPackages = with pkgs; [\n")

	// Group packages by category
	categories := make(map[string][]string)
	for _, mapping := range mappings {
		if mapping.SystemPackage && mapping.Availability == "available" {
			if categories[mapping.Category] == nil {
				categories[mapping.Category] = []string{}
			}
			categories[mapping.Category] = append(categories[mapping.Category], mapping.NixpkgsName)
		}
	}

	// Write packages grouped by category
	for category, pkgs := range categories {
		if len(pkgs) > 0 {
			config.WriteString(fmt.Sprintf("    # %s\n", category))
			for _, pkg := range pkgs {
				config.WriteString(fmt.Sprintf("    %s\n", pkg))
			}
			config.WriteString("\n")
		}
	}

	config.WriteString("  ];\n\n")

	// Add notes for problematic packages
	hasNotes := false
	for pkg, mapping := range mappings {
		if mapping.Notes != "" || mapping.Availability != "available" {
			if !hasNotes {
				config.WriteString("  # Package migration notes:\n")
				hasNotes = true
			}

			if mapping.Availability == "not_needed" {
				config.WriteString(fmt.Sprintf("  # %s: Not needed in NixOS - %s\n", pkg, mapping.Notes))
			} else if mapping.Availability == "missing" {
				config.WriteString(fmt.Sprintf("  # %s: Not available in nixpkgs\n", pkg))
			} else if mapping.Notes != "" {
				config.WriteString(fmt.Sprintf("  # %s: %s\n", pkg, mapping.Notes))
			}
		}
	}

	if hasNotes {
		config.WriteString("\n")
	}

	return config.String()
}

// GetAlternatives returns alternative packages for a given package
func (pm *PackageMapper) GetAlternatives(packageName string) []string {
	if mapping, exists := pm.MapPackage(packageName); exists {
		return mapping.Alternatives
	}
	return []string{}
}

// ValidateMapping validates a package mapping
func (pm *PackageMapper) ValidateMapping(mapping PackageMapping) []string {
	var warnings []string

	if mapping.Availability == "missing" {
		warnings = append(warnings, "Package not available in nixpkgs")
	}

	if mapping.Availability == "unstable" {
		warnings = append(warnings, "Package only available in nixpkgs-unstable")
	}

	if len(mapping.Alternatives) > 0 {
		warnings = append(warnings, fmt.Sprintf("Consider alternatives: %s", strings.Join(mapping.Alternatives, ", ")))
	}

	if mapping.Notes != "" {
		warnings = append(warnings, mapping.Notes)
	}

	return warnings
}
