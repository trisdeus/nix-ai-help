package nixlang

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInconsistencyDetector_DetectServiceConflicts(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		severity string
		title    string
	}{
		{
			name: "multiple web servers",
			content: `
				services.nginx.enable = true;
				services.apache.enable = true;
			`,
			expected: 1,
			severity: "major",
			title:    "Multiple Web Servers Enabled",
		},
		{
			name: "mysql and mariadb conflict",
			content: `
				services.mysql.enable = true;
				services.mariadb.enable = true;
			`,
			expected: 2, // Both general DB conflict and specific MySQL/MariaDB conflict
			severity: "critical",
			title:    "MySQL and MariaDB Both Enabled",
		},
		{
			name: "multiple display managers",
			content: `
				services.xserver.displayManager.gdm.enable = true;
				services.xserver.displayManager.lightdm.enable = true;
			`,
			expected: 1,
			severity: "major",
			title:    "Multiple Display Managers Enabled",
		},
		{
			name: "no conflicts",
			content: `
				services.nginx.enable = true;
				services.postgresql.enable = true;
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectServiceConflicts(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				found := false
				for _, inc := range inconsistencies {
					if inc.Title == tt.title {
						assert.Equal(t, tt.severity, inc.Severity)
						found = true
						break
					}
				}
				assert.True(t, found, "Expected inconsistency with title '%s' not found", tt.title)
			}
		})
	}
}

func TestInconsistencyDetector_DetectMissingDependencies(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		title    string
	}{
		{
			name: "nginx without ssl",
			content: `
				services.nginx.enable = true;
			`,
			expected: 1,
			title:    "Nginx Without SSL Configuration",
		},
		{
			name: "postgresql without backup",
			content: `
				services.postgresql.enable = true;
			`,
			expected: 1,
			title:    "Postgresql Without Backup Configuration",
		},
		{
			name: "nginx with ssl config",
			content: `
				services.nginx.enable = true;
				security.acme.acceptTerms = true;
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectMissingDependencies(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
			}
		})
	}
}

func TestInconsistencyDetector_DetectSecurityInconsistencies(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		severity string
		title    string
	}{
		{
			name: "firewall disabled with exposed services",
			content: `
				networking.firewall.enable = false;
				services.nginx.enable = true;
				services.openssh.enable = true;
			`,
			expected: 1,
			severity: "critical",
			title:    "Firewall Disabled with Exposed Services",
		},
		{
			name: "root ssh without keys",
			content: `
				services.openssh.settings.PermitRootLogin = "yes";
			`,
			expected: 1,
			severity: "critical",
			title:    "Root SSH Without Key Authentication",
		},
		{
			name: "secure configuration",
			content: `
				networking.firewall.enable = true;
				services.openssh.settings.PermitRootLogin = "no";
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectSecurityInconsistencies(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
				assert.Equal(t, tt.severity, inconsistencies[0].Severity)
			}
		})
	}
}

func TestInconsistencyDetector_DetectNetworkingConflicts(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		title    string
	}{
		{
			name: "multiple network managers",
			content: `
				networking.networkmanager.enable = true;
				networking.wicd.enable = true;
			`,
			expected: 1,
			title:    "Multiple Network Managers Enabled",
		},
		{
			name: "static ip with dhcp",
			content: `
				networking.useDHCP = true;
				networking.interfaces.eth0.ipv4.addresses = [ { address = "192.168.1.100"; prefixLength = 24; } ];
			`,
			expected: 1,
			title:    "Static IP with DHCP Enabled",
		},
		{
			name: "single network manager",
			content: `
				networking.networkmanager.enable = true;
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectNetworkingConflicts(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
			}
		})
	}
}

func TestInconsistencyDetector_DetectResourceConflicts(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		title    string
	}{
		{
			name: "port 80 conflict",
			content: `
				services.nginx.enable = true;
				services.apache.enable = true;
			`,
			expected: 1,
			title:    "Port 80 Conflict",
		},
		{
			name: "database port conflict",
			content: `
				services.mysql.enable = true;
				services.mariadb.enable = true;
			`,
			expected: 1,
			title:    "Port 3306 Conflict",
		},
		{
			name: "no port conflicts",
			content: `
				services.nginx.enable = true;
				services.postgresql.enable = true;
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectResourceConflicts(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
			}
		})
	}
}

func TestInconsistencyDetector_DetectFilesystemConflicts(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		title    string
	}{
		{
			name: "filesystem conflict on root",
			content: `
				fileSystems."/" = { fsType = "ext4"; };
				fileSystems."/" = { fsType = "btrfs"; };
			`,
			expected: 1,
			title:    "Conflicting Filesystems on /",
		},
		{
			name: "no filesystem conflicts",
			content: `
				fileSystems."/" = { fsType = "ext4"; };
				fileSystems."/home" = { fsType = "btrfs"; };
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectFilesystemConflicts(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
			}
		})
	}
}

func TestInconsistencyDetector_DetectVersionConflicts(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		title    string
	}{
		{
			name: "multiple python versions",
			content: `
				environment.systemPackages = with pkgs; [
					python39
					python310
				];
			`,
			expected: 1,
			title:    "Multiple Versions of python",
		},
		{
			name: "single package version",
			content: `
				environment.systemPackages = with pkgs; [
					python39
					git
				];
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectVersionConflicts(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
			}
		})
	}
}

func TestInconsistencyDetector_DetectPerformanceConflicts(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name     string
		content  string
		expected int
		title    string
	}{
		{
			name: "multiple swap configurations",
			content: `
				zramSwap.enable = true;
				swapDevices = [ { device = "/swapfile"; size = 2048; } ];
			`,
			expected: 1,
			title:    "Multiple Swap Configurations",
		},
		{
			name: "single swap configuration",
			content: `
				zramSwap.enable = true;
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := analyzer.parser.ParseExpression(tt.content)
			require.NoError(t, err)

			inconsistencies := detector.detectPerformanceConflicts(tt.content, expr)
			assert.Len(t, inconsistencies, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, tt.title, inconsistencies[0].Title)
			}
		})
	}
}

func TestInconsistencyDetector_ComprehensiveDetection(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	// Complex configuration with multiple issues
	content := `
		# Multiple web servers - conflict
		services.nginx.enable = true;
		services.apache.enable = true;
		
		# Firewall disabled with exposed services - security issue
		networking.firewall.enable = false;
		services.openssh.enable = true;
		
		# Root SSH without keys - security issue
		services.openssh.settings.PermitRootLogin = "yes";
		
		# Multiple network managers - networking conflict
		networking.networkmanager.enable = true;
		networking.wicd.enable = true;
		
		# Static IP with DHCP - networking conflict
		networking.useDHCP = true;
		networking.interfaces.eth0.ipv4.addresses = [ { address = "192.168.1.100"; prefixLength = 24; } ];
		
		# Multiple swap configs - performance conflict
		zramSwap.enable = true;
		swapDevices = [ { device = "/swapfile"; size = 2048; } ];
	`

	result, err := detector.DetectInconsistencies(content)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should detect multiple inconsistencies
	assert.Greater(t, len(result.Inconsistencies), 5)

	// Check summary
	assert.Equal(t, len(result.Inconsistencies), result.Summary.TotalInconsistencies)
	assert.Greater(t, result.Summary.BySeverity["critical"], 0)
	assert.Greater(t, result.Summary.BySeverity["major"], 0)

	// Check statistics
	assert.Greater(t, result.Statistics.ManualResolution, 0)
	assert.NotEmpty(t, result.Statistics.MostCommonType)
	assert.NotEqual(t, "excellent", result.Statistics.ConfigurationHealth)

	// Verify specific inconsistency types are detected
	foundTypes := make(map[InconsistencyType]bool)
	for _, inc := range result.Inconsistencies {
		foundTypes[inc.Type] = true
	}

	assert.True(t, foundTypes[ConflictingServices])
	assert.True(t, foundTypes[SecurityInconsistency])
	assert.True(t, foundTypes[NetworkingConflict])
	assert.True(t, foundTypes[PerformanceConflict])
}

func TestInconsistencyDetector_ResolutionOptions(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	content := `
		services.nginx.enable = true;
		services.apache.enable = true;
	`

	result, err := detector.DetectInconsistencies(content)
	require.NoError(t, err)
	require.Greater(t, len(result.Inconsistencies), 0)

	inconsistency := result.Inconsistencies[0]
	require.NotNil(t, inconsistency.Resolution)
	assert.Equal(t, "choose_one", inconsistency.Resolution.Type)
	assert.Greater(t, len(inconsistency.Resolution.Options), 0)

	// Check that resolution options have required fields
	for _, option := range inconsistency.Resolution.Options {
		assert.NotEmpty(t, option.Title)
		assert.NotEmpty(t, option.Description)
		assert.Greater(t, len(option.Changes), 0)
		assert.NotEmpty(t, option.Risk)
	}
}

func TestInconsistencyDetector_ConfidenceScoring(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	// High confidence issue (MySQL + MariaDB)
	content := `
		services.mysql.enable = true;
		services.mariadb.enable = true;
	`

	result, err := detector.DetectInconsistencies(content)
	require.NoError(t, err)
	require.Greater(t, len(result.Inconsistencies), 0)

	// Find the MySQL/MariaDB conflict
	var mysqlMariadbConflict *LogicalInconsistency
	for _, inc := range result.Inconsistencies {
		if inc.Title == "MySQL and MariaDB Both Enabled" {
			mysqlMariadbConflict = &inc
			break
		}
	}

	require.NotNil(t, mysqlMariadbConflict)
	assert.Equal(t, 1.0, mysqlMariadbConflict.Confidence)
	assert.Equal(t, "critical", mysqlMariadbConflict.Severity)
}

func TestInconsistencyDetector_EdgeCases(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "empty configuration",
			content: "",
			wantErr: false,
		},
		{
			name:    "comment only",
			content: "# This is just a comment",
			wantErr: false,
		},
		{
			name: "valid minimal config",
			content: `
				boot.loader.systemd-boot.enable = true;
			`,
			wantErr: false,
		},
		{
			name: "invalid syntax",
			content: `
				this is not valid nix syntax {{{
			`,
			wantErr: false, // We handle parse failures gracefully now
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.DetectInconsistencies(tt.content)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Summary)
				assert.NotNil(t, result.Statistics)
			}
		})
	}
}

func TestInconsistencyDetector_HelperFunctions(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	t.Run("serviceEnabled", func(t *testing.T) {
		content := `
			services.nginx.enable = true;
			services.apache.enable = false;
		`
		assert.True(t, detector.serviceEnabled(content, "nginx"))
		assert.False(t, detector.serviceEnabled(content, "apache"))
		assert.False(t, detector.serviceEnabled(content, "postgresql"))
	})

	t.Run("hasSSLConfig", func(t *testing.T) {
		content1 := `security.acme.acceptTerms = true;`
		content2 := `services.nginx.virtualHosts."example.com".enableACME = true;`
		content3 := `services.nginx.enable = true;`

		assert.True(t, detector.hasSSLConfig(content1))
		assert.True(t, detector.hasSSLConfig(content2))
		assert.False(t, detector.hasSSLConfig(content3))
	})

	t.Run("firewallDisabled", func(t *testing.T) {
		content1 := `networking.firewall.enable = false;`
		content2 := `networking.firewall.enable = true;`
		content3 := `# no firewall config`

		assert.True(t, detector.firewallDisabled(content1))
		assert.False(t, detector.firewallDisabled(content2))
		assert.False(t, detector.firewallDisabled(content3))
	})

	t.Run("findPortConflicts", func(t *testing.T) {
		content := `
			services.nginx.enable = true;
			services.apache.enable = true;
			services.mysql.enable = true;
			services.mariadb.enable = true;
		`

		conflicts := detector.findPortConflicts(content)
		assert.Contains(t, conflicts, "80")
		assert.Contains(t, conflicts, "3306")
		assert.Len(t, conflicts["80"], 2)  // nginx + apache
		assert.Len(t, conflicts["3306"], 2) // mysql + mariadb
	})
}

func TestInconsistencyDetector_StatisticsGeneration(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	// Create test inconsistencies
	inconsistencies := []LogicalInconsistency{
		{
			Type:       ConflictingServices,
			Severity:   "critical",
			Confidence: 1.0,
			Resolution: &Resolution{Automatic: true},
		},
		{
			Type:       SecurityInconsistency,
			Severity:   "major",
			Confidence: 0.9,
			Resolution: &Resolution{Automatic: false},
		},
		{
			Type:       ConflictingServices,
			Severity:   "minor",
			Confidence: 0.8,
			Resolution: &Resolution{Automatic: true},
		},
	}

	stats := detector.generateStatistics(inconsistencies)

	assert.Equal(t, 2, stats.AutoResolvable)
	assert.Equal(t, 1, stats.ManualResolution)
	assert.Equal(t, 0.9, stats.AverageConfidence)
	assert.Equal(t, "conflicting_services", stats.MostCommonType)
	assert.Equal(t, "poor", stats.ConfigurationHealth) // Has critical issue
}

func TestInconsistencyDetector_SummaryGeneration(t *testing.T) {
	analyzer := NewNixAnalyzer()
	detector := NewInconsistencyDetector(analyzer)

	// Create test inconsistencies
	inconsistencies := []LogicalInconsistency{
		{
			Type:     ConflictingServices,
			Severity: "critical",
			Title:    "Critical Issue 1",
		},
		{
			Type:     SecurityInconsistency,
			Severity: "critical",
			Title:    "Critical Issue 2",
		},
		{
			Type:     NetworkingConflict,
			Severity: "major",
			Title:    "Major Issue 1",
		},
	}

	summary := detector.generateSummary(inconsistencies)

	assert.Equal(t, 3, summary.TotalInconsistencies)
	assert.Equal(t, 2, summary.BySeverity["critical"])
	assert.Equal(t, 1, summary.BySeverity["major"])
	assert.Equal(t, 1, summary.ByType["conflicting_services"])
	assert.Equal(t, 1, summary.ByType["security_inconsistency"])
	assert.Equal(t, 1, summary.ByType["networking_conflict"])
	assert.Contains(t, summary.CriticalIssues, "Critical Issue 1")
	assert.Contains(t, summary.CriticalIssues, "Critical Issue 2")
	assert.Equal(t, "critical", summary.OverallRisk)
}