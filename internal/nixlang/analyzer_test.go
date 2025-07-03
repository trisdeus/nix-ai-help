package nixlang

import (
	"fmt"
	"testing"
)

func TestNixAnalyzer_AnalyzeExpression(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name            string
		source          string
		expectIssues    bool
		expectSecurity  bool
		expectOptimizations bool
	}{
		{
			name:            "simple valid configuration",
			source:          `{ services.nginx.enable = true; }`,
			expectIssues:    false,
			expectSecurity:  false,
			expectOptimizations: false,
		},
		{
			name:            "insecure HTTP URL",
			source:          `{ url = "http://example.com/download"; }`,
			expectIssues:    false,
			expectSecurity:  true,
			expectOptimizations: false,
		},
		{
			name:            "hardcoded secret",
			source:          `{ password = "secret123"; }`,
			expectIssues:    false,
			expectSecurity:  true,
			expectOptimizations: false,
		},
		{
			name:            "root execution",
			source:          `{ services.myservice.user = "root"; }`,
			expectIssues:    false,
			expectSecurity:  true,
			expectOptimizations: false,
		},
		{
			name:            "disabled firewall",
			source:          `{ networking.firewall.enable = false; }`,
			expectIssues:    false,
			expectSecurity:  true,
			expectOptimizations: false,
		},
		{
			name:            "unnecessary with statement",
			source:          `with pkgs; [ git ]`,
			expectIssues:    false,
			expectSecurity:  false,
			expectOptimizations: true,
		},
		{
			name:            "nested with statements",
			source:          `with pkgs; with lib; { }`,
			expectIssues:    true,
			expectSecurity:  false,
			expectOptimizations: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			if tt.expectIssues && len(result.Issues) == 0 {
				t.Errorf("Expected issues but got none")
			}
			
			if !tt.expectIssues && len(result.Issues) > 0 {
				t.Errorf("Expected no issues but got %d: %v", len(result.Issues), result.Issues)
			}
			
			if tt.expectSecurity && len(result.SecurityFindings) == 0 {
				t.Errorf("Expected security findings but got none")
			}
			
			if !tt.expectSecurity && len(result.SecurityFindings) > 0 {
				t.Errorf("Expected no security findings but got %d", len(result.SecurityFindings))
			}
			
			if tt.expectOptimizations && len(result.Optimizations) == 0 {
				t.Errorf("Expected optimizations but got none")
			}
			
			if !tt.expectOptimizations && len(result.Optimizations) > 0 {
				t.Errorf("Expected no optimizations but got %d", len(result.Optimizations))
			}
		})
	}
}

func TestNixAnalyzer_IntentAnalysis(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name           string
		source         string
		expectedIntent string
		minConfidence  float64
	}{
		{
			name:           "service configuration",
			source:         `{ services.nginx.enable = true; }`,
			expectedIntent: "service_configuration",
			minConfidence:  0.8,
		},
		{
			name:           "package management",
			source:         `{ environment.systemPackages = with pkgs; [ git vim ]; }`,
			expectedIntent: "package_management",
			minConfidence:  0.8,
		},
		{
			name:           "user management",
			source:         `{ users.users.alice = { isNormalUser = true; }; }`,
			expectedIntent: "user_management",
			minConfidence:  0.7,
		},
		{
			name:           "network configuration",
			source:         `{ networking.hostName = "myserver"; }`,
			expectedIntent: "network_configuration",
			minConfidence:  0.8,
		},
		{
			name:           "security configuration",
			source:         `{ security.sudo.enable = true; }`,
			expectedIntent: "security_configuration",
			minConfidence:  0.8,
		},
		{
			name:           "hardware configuration",
			source:         `{ hardware.bluetooth.enable = true; }`,
			expectedIntent: "hardware_configuration",
			minConfidence:  0.7,
		},
		{
			name:           "boot configuration",
			source:         `{ boot.loader.systemd-boot.enable = true; }`,
			expectedIntent: "boot_configuration",
			minConfidence:  0.8,
		},
		{
			name:           "development environment",
			source:         `mkShell { buildInputs = [ nodejs python3 ]; }`,
			expectedIntent: "development_environment",
			minConfidence:  0.7,
		},
		{
			name:           "package derivation",
			source:         `stdenv.mkDerivation { name = "mypackage"; }`,
			expectedIntent: "package_derivation",
			minConfidence:  0.8,
		},
		{
			name:           "flake configuration",
			source:         `{ description = "My flake"; inputs = {}; outputs = {}; }`,
			expectedIntent: "flake_configuration",
			minConfidence:  0.8,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			if result.Intent.PrimaryIntent != tt.expectedIntent {
				t.Errorf("Expected intent %s, got %s", tt.expectedIntent, result.Intent.PrimaryIntent)
			}
			
			if result.Intent.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %f, got %f", tt.minConfidence, result.Intent.Confidence)
			}
		})
	}
}

func TestNixAnalyzer_SecurityAnalysis(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name           string
		source         string
		expectedFindings []string
		expectedSeverity Severity
	}{
		{
			name:             "insecure HTTP URL",
			source:           `{ fetchurl = { url = "http://example.com/file.tar.gz"; }; }`,
			expectedFindings: []string{"insecure_url"},
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "hardcoded password",
			source:           `{ database = { password = "supersecret123"; }; }`,
			expectedFindings: []string{"hardcoded_secret"},
			expectedSeverity: SeverityError,
		},
		{
			name:             "weak file permissions",
			source:           `{ systemd.tmpfiles.rules = [ "f /tmp/test 0666 root root" ]; }`,
			expectedFindings: []string{"weak_permissions"},
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "root service user",
			source:           `{ systemd.services.myservice.serviceConfig.User = "root"; }`,
			expectedFindings: []string{"root_execution"},
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "disabled firewall",
			source:           `{ networking.firewall.enable = false; }`,
			expectedFindings: []string{"disabled_firewall"},
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "SSH root login",
			source:           `{ services.openssh.permitRootLogin = "yes"; }`,
			expectedFindings: []string{"weak_ssh_config"},
			expectedSeverity: SeverityError,
		},
		{
			name:             "insecure package",
			source:           `{ environment.systemPackages = [ pkgs.flash ]; }`,
			expectedFindings: []string{"insecure_package"},
			expectedSeverity: SeverityInfo,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			if len(result.SecurityFindings) == 0 {
				t.Fatalf("Expected security findings but got none")
			}
			
			// Check if expected finding types are present
			for _, expectedType := range tt.expectedFindings {
				found := false
				for _, finding := range result.SecurityFindings {
					if finding.Type == expectedType {
						found = true
						if finding.Severity != tt.expectedSeverity {
							t.Errorf("Expected severity %v for %s, got %v", 
								tt.expectedSeverity, expectedType, finding.Severity)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected security finding %s not found", expectedType)
				}
			}
		})
	}
}

func TestNixAnalyzer_ComplexityAnalysis(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name                string
		source              string
		expectHighComplexity bool
		minNesting          int
	}{
		{
			name:                "simple configuration",
			source:              `{ services.nginx.enable = true; }`,
			expectHighComplexity: false,
			minNesting:          2,
		},
		{
			name: "complex nested configuration",
			source: `{
				services = {
					nginx = {
						enable = true;
						virtualHosts = {
							"site1.com" = {
								locations = {
									"/" = { proxyPass = "http://localhost:3000"; };
									"/api" = { proxyPass = "http://localhost:4000"; };
								};
							};
							"site2.com" = {
								locations = {
									"/" = { proxyPass = "http://localhost:5000"; };
								};
							};
						};
					};
					postgresql = {
						enable = true;
						settings = {
							max_connections = 100;
							shared_buffers = "128MB";
						};
					};
				};
			}`,
			expectHighComplexity: true,
			minNesting:          4,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			if tt.expectHighComplexity && result.Complexity.Total < 10 {
				t.Errorf("Expected high complexity (>= 10), got %d", result.Complexity.Total)
			}
			
			if !tt.expectHighComplexity && result.Complexity.Total >= 10 {
				t.Errorf("Expected low complexity (< 10), got %d", result.Complexity.Total)
			}
			
			if result.Complexity.Nesting < tt.minNesting {
				t.Errorf("Expected nesting >= %d, got %d", tt.minNesting, result.Complexity.Nesting)
			}
		})
	}
}

func TestNixAnalyzer_QualityMetrics(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name                string
		source              string
		expectGoodQuality   bool
		minOverallQuality   float64
	}{
		{
			name:              "clean simple configuration",
			source:            `{ services.nginx.enable = true; networking.hostName = "webserver"; }`,
			expectGoodQuality: true,
			minOverallQuality: 0.8,
		},
		{
			name: "configuration with security issues",
			source: `{
				services.openssh.permitRootLogin = "yes";
				networking.firewall.enable = false;
				database.password = "admin123";
			}`,
			expectGoodQuality: false,
			minOverallQuality: 0.3,
		},
		{
			name: "overly complex configuration",
			source: `{
				services = {
					nginx = {
						virtualHosts = {
							"a.com" = { locations."/" = { proxyPass = "http://1"; }; };
							"b.com" = { locations."/" = { proxyPass = "http://2"; }; };
							"c.com" = { locations."/" = { proxyPass = "http://3"; }; };
							"d.com" = { locations."/" = { proxyPass = "http://4"; }; };
							"e.com" = { locations."/" = { proxyPass = "http://5"; }; };
						};
					};
				};
			}`,
			expectGoodQuality: false,
			minOverallQuality: 0.4,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			if tt.expectGoodQuality && result.Quality.Overall < tt.minOverallQuality {
				t.Errorf("Expected good quality (>= %f), got %f", tt.minOverallQuality, result.Quality.Overall)
			}
			
			if !tt.expectGoodQuality && result.Quality.Overall > tt.minOverallQuality {
				t.Errorf("Expected poor quality (<= %f), got %f", tt.minOverallQuality, result.Quality.Overall)
			}
			
			// Check that all quality dimensions are calculated
			if result.Quality.Maintainability < 0 || result.Quality.Maintainability > 1 {
				t.Errorf("Maintainability score out of range: %f", result.Quality.Maintainability)
			}
			
			if result.Quality.Security < 0 || result.Quality.Security > 1 {
				t.Errorf("Security score out of range: %f", result.Quality.Security)
			}
			
			if result.Quality.Reliability < 0 || result.Quality.Reliability > 1 {
				t.Errorf("Reliability score out of range: %f", result.Quality.Reliability)
			}
		})
	}
}

func TestNixAnalyzer_AntiPatternDetection(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name                string
		source              string
		expectedAntiPatterns []string
	}{
		{
			name:                "nested with statements",
			source:              `with pkgs; with lib; { hello = world; }`,
			expectedAntiPatterns: []string{"nested with"},
		},
		{
			name:                "empty nixpkgs import",
			source:              `let pkgs = import <nixpkgs> {}; in pkgs`,
			expectedAntiPatterns: []string{"empty nixpkgs"},
		},
		{
			name:                "very large configuration",
			source:              generateLargeConfig(),
			expectedAntiPatterns: []string{"large configuration"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			antiPatternIssues := []Issue{}
			for _, issue := range result.Issues {
				if issue.Type == IssueAntiPattern {
					antiPatternIssues = append(antiPatternIssues, issue)
				}
			}
			
			if len(tt.expectedAntiPatterns) > 0 && len(antiPatternIssues) == 0 {
				t.Errorf("Expected anti-pattern detection but got none")
			}
		})
	}
}

func TestNixAnalyzer_DependencyAnalysis(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	source := `{
		environment.systemPackages = with pkgs; [ git vim nodejs ];
		services.nginx.package = pkgs.nginx;
		services.postgresql.package = config.services.postgresql.package;
		nixpkgs.config = config.nixpkgs.config;
	}`
	
	result, err := analyzer.AnalyzeExpression(source)
	if err != nil {
		t.Fatalf("AnalyzeExpression() error = %v", err)
	}
	
	// Check that dependencies were found
	if len(result.Dependencies.Direct) == 0 {
		t.Errorf("Expected direct dependencies but got none")
	}
	
	// Check dependency graph
	if len(result.Dependencies.Graph.Nodes) == 0 {
		t.Errorf("Expected dependency graph nodes but got none")
	}
	
	// Verify specific dependencies
	expectedDeps := []string{"pkgs", "config"}
	found := make(map[string]bool)
	
	for _, dep := range result.Dependencies.Direct {
		found[dep.Name] = true
	}
	
	for _, expectedDep := range expectedDeps {
		if !found[expectedDep] {
			t.Errorf("Expected dependency %s not found", expectedDep)
		}
	}
}

func TestNixAnalyzer_OptimizationSuggestions(t *testing.T) {
	analyzer := NewNixAnalyzer()
	
	tests := []struct {
		name               string
		source             string
		expectedOptimizations []string
	}{
		{
			name:               "unnecessary with statement",
			source:             `with pkgs; [ git ]`,
			expectedOptimizations: []string{"unnecessary_with"},
		},
		{
			name:               "duplicate package references",
			source:             `{ a = pkgs.git; b = pkgs.git; }`,
			expectedOptimizations: []string{"duplicate_packages"},
		},
		{
			name:               "deeply nested imports",
			source:             `import (import (import ./config.nix))`,
			expectedOptimizations: []string{"nested_imports"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeExpression(tt.source)
			if err != nil {
				t.Fatalf("AnalyzeExpression() error = %v", err)
			}
			
			if len(tt.expectedOptimizations) > 0 && len(result.Optimizations) == 0 {
				t.Errorf("Expected optimizations but got none")
			}
			
			for _, expectedOpt := range tt.expectedOptimizations {
				found := false
				for _, opt := range result.Optimizations {
					if opt.Type == expectedOpt {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected optimization %s not found", expectedOpt)
				}
			}
		})
	}
}

// Helper function to generate a large configuration for testing
func generateLargeConfig() string {
	config := `{
		# This is a very large configuration file for testing
		services = {`
	
	// Generate many service configurations
	for i := 0; i < 50; i++ {
		config += fmt.Sprintf(`
			service%d = {
				enable = true;
				port = %d;
				config = "long configuration string here";
				extraOptions = {
					option1 = "value1";
					option2 = "value2";
					option3 = "value3";
				};
			};`, i, 3000+i)
	}
	
	config += `
		};
		environment.systemPackages = [`
	
	// Generate many package references
	for i := 0; i < 100; i++ {
		config += fmt.Sprintf(" pkgs.package%d", i)
	}
	
	config += `
		];
	}`
	
	return config
}

func BenchmarkNixAnalyzer_AnalyzeExpression(b *testing.B) {
	analyzer := NewNixAnalyzer()
	
	source := `{
		services = {
			nginx = {
				enable = true;
				virtualHosts."example.com".locations."/".proxyPass = "http://localhost:3000";
			};
			postgresql = {
				enable = true;
				package = pkgs.postgresql_13;
			};
		};
		environment.systemPackages = with pkgs; [ git vim nodejs python3 ];
		networking = {
			hostName = "webserver";
			firewall.allowedTCPPorts = [ 80 443 ];
		};
		users.users.deploy = {
			isNormalUser = true;
			extraGroups = [ "wheel" ];
		};
	}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeExpression(source)
		if err != nil {
			b.Fatalf("AnalyzeExpression() error = %v", err)
		}
	}
}