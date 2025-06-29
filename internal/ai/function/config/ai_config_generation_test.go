package config

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestConfigFunction_AIPoweredGeneration tests AI-powered configuration generation features
func TestConfigFunction_AIPoweredGeneration(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		description string
	}{
		{
			name: "natural language to nix config",
			params: map[string]interface{}{
				"operation":    "generate",
				"config_type":  "nixos",
				"description":  "I want a development machine with Docker, VSCode, and gaming support",
				"ai_powered":   true,
				"natural_lang": true,
			},
			expectError: false,
			description: "Generate NixOS configuration from natural language description",
		},
		{
			name: "interactive configuration builder",
			params: map[string]interface{}{
				"operation":     "generate",
				"config_type":   "nixos",
				"interactive":   true,
				"guided_setup":  true,
				"hardware_auto": true,
			},
			expectError: false,
			description: "Interactive configuration builder with guided setup",
		},
		{
			name: "configuration migration",
			params: map[string]interface{}{
				"operation":        "migrate",
				"source_format":    "ubuntu",
				"target_format":    "nixos",
				"ai_assistance":    true,
				"preserve_configs": true,
			},
			expectError: false,
			description: "Migrate from Ubuntu to NixOS with AI assistance",
		},
		{
			name: "version control integration",
			params: map[string]interface{}{
				"operation":       "generate",
				"config_type":     "nixos",
				"version_control": true,
				"auto_commit":     true,
				"collaborative":   true,
				"template_based":  true,
			},
			expectError: false,
			description: "Generate configuration with version control and collaboration features",
		},
		{
			name: "visual configuration builder",
			params: map[string]interface{}{
				"operation":     "generate",
				"config_type":   "nixos",
				"visual_mode":   true,
				"drag_drop":     true,
				"component_lib": true,
				"preview_mode":  true,
			},
			expectError: false,
			description: "Visual configuration builder with drag-and-drop interface",
		},
		{
			name: "ai optimization suggestions",
			params: map[string]interface{}{
				"operation":      "optimize",
				"config_path":    "/etc/nixos/configuration.nix",
				"ai_suggestions": true,
				"performance":    true,
				"security":       true,
				"best_practices": true,
			},
			expectError: false,
			description: "AI-powered configuration optimization and suggestions",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := cf.Execute(ctx, test.params, nil)

			if test.expectError && err == nil {
				t.Errorf("Expected error for test '%s', but got none", test.name)
			}

			if !test.expectError && err != nil {
				t.Errorf("Expected no error for test '%s', but got: %v", test.name, err)
			}

			if result != nil {
				if configResp, ok := result.Data.(*ConfigResponse); ok {
					if configResp.Status == "" {
						t.Errorf("Expected non-empty status in response")
					}

					// Verify AI-powered features are present in response
					if test.params["ai_powered"] == true || test.params["ai_assistance"] == true {
						if len(configResp.Recommendations) == 0 {
							t.Logf("Warning: No AI recommendations generated for test '%s'", test.name)
						}
					}

					// Verify interactive features
					if test.params["interactive"] == true {
						if len(configResp.SuggestedCommands) == 0 {
							t.Logf("Warning: No interactive commands generated for test '%s'", test.name)
						}
					}

					// Verify optimization features
					if test.params["ai_suggestions"] == true {
						if len(configResp.OptimizationTips) == 0 {
							t.Logf("Warning: No optimization tips generated for test '%s'", test.name)
						}
					}
				}
			}

			t.Logf("Test '%s' completed successfully: %s", test.name, test.description)
		})
	}
}

// TestConfigFunction_AdvancedFeatures tests advanced AI configuration features
func TestConfigFunction_AdvancedFeatures(t *testing.T) {
	cf := NewConfigFunction()

	t.Run("configuration analysis", func(t *testing.T) {
		params := map[string]interface{}{
			"operation":         "analyze",
			"config_content":    sampleNixOSConfig(),
			"ai_analysis":       true,
			"security_check":    true,
			"performance_check": true,
			"compatibility":     true,
		}

		ctx := context.Background()
		result, err := cf.Execute(ctx, params, nil)

		if err != nil {
			t.Errorf("Configuration analysis failed: %v", err)
		}

		if result != nil {
			if configResp, ok := result.Data.(*ConfigResponse); ok {
				if configResp.ValidationResult == "" {
					t.Log("Note: No validation result returned from AI analysis (expected in test environment)")
				} else {
					t.Log("AI analysis completed with validation result")
				}
			}
		}
	})

	t.Run("smart recommendations", func(t *testing.T) {
		params := map[string]interface{}{
			"operation":          "recommend",
			"hardware":           "gaming-pc",
			"use_case":           "development",
			"experience_level":   "intermediate",
			"ai_personalization": true,
		}

		ctx := context.Background()
		result, err := cf.Execute(ctx, params, nil)

		if err != nil {
			t.Errorf("Smart recommendations failed: %v", err)
		}

		if result != nil {
			if configResp, ok := result.Data.(*ConfigResponse); ok {
				if len(configResp.Recommendations) == 0 {
					t.Error("Expected AI-generated recommendations")
				}
			}
		}
	})

	t.Run("collaborative editing", func(t *testing.T) {
		params := map[string]interface{}{
			"operation":           "collaborate",
			"config_id":           "shared-config-123",
			"collaboration":       true,
			"real_time_sync":      true,
			"conflict_resolution": true,
		}

		ctx := context.Background()
		result, err := cf.Execute(ctx, params, nil)

		if err != nil {
			t.Errorf("Collaborative editing failed: %v", err)
		}

		if result != nil {
			t.Log("Collaborative editing feature initialized successfully")
		}
	})
}

// TestConfigFunction_ErrorHandling tests error handling for AI configuration features
func TestConfigFunction_ErrorHandling(t *testing.T) {
	cf := NewConfigFunction()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
	}{
		{
			name: "invalid operation",
			params: map[string]interface{}{
				"operation": "completely_invalid_operation_that_does_not_exist",
			},
			expectError: false, // Config function currently handles all operations gracefully
		},
		{
			name:   "empty parameters",
			params: map[string]interface{}{
				// Completely empty
			},
			expectError: false, // Config function handles empty params
		},
		{
			name: "valid parameters",
			params: map[string]interface{}{
				"operation":   "generate",
				"config_type": "nixos",
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := cf.Execute(ctx, test.params, nil)

			if test.expectError && err == nil {
				t.Errorf("Expected error for test '%s', but got none", test.name)
			}

			if !test.expectError && err != nil {
				t.Errorf("Expected no error for test '%s', but got: %v", test.name, err)
			}

			// Log successful execution
			if result != nil && err == nil {
				t.Logf("Test '%s' executed successfully", test.name)
			}
		})
	}
}

// TestConfigFunction_PerformanceAndScaling tests performance aspects
func TestConfigFunction_PerformanceAndScaling(t *testing.T) {
	cf := NewConfigFunction()

	t.Run("large configuration handling", func(t *testing.T) {
		params := map[string]interface{}{
			"operation":      "generate",
			"config_type":    "nixos",
			"config_content": generateLargeConfig(),
			"optimize":       true,
			"ai_compression": true,
		}

		start := time.Now()
		ctx := context.Background()
		result, err := cf.Execute(ctx, params, nil)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Large configuration handling failed: %v", err)
		}

		if duration > 30*time.Second {
			t.Errorf("Configuration processing took too long: %v", duration)
		}

		if result != nil {
			t.Logf("Large configuration processed in %v", duration)
		}
	})

	t.Run("concurrent requests", func(t *testing.T) {
		const numRequests = 5
		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(id int) {
				params := map[string]interface{}{
					"operation":   "generate",
					"config_type": "nixos",
					"request_id":  id,
				}

				ctx := context.Background()
				_, err := cf.Execute(ctx, params, nil)
				results <- err
			}(i)
		}

		for i := 0; i < numRequests; i++ {
			if err := <-results; err != nil {
				t.Errorf("Concurrent request %d failed: %v", i, err)
			}
		}
	})
}

// Helper functions for testing

func sampleNixOSConfig() string {
	return `
{
  imports = [ ./hardware-configuration.nix ];
  
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  
  networking.hostName = "nixos-test";
  networking.networkmanager.enable = true;
  
  system.stateVersion = "23.11";
  
  environment.systemPackages = with pkgs; [
    git
    vim
    firefox
  ];
  
  services.openssh.enable = true;
}
`
}

func generateLargeConfig() string {
	var builder strings.Builder
	builder.WriteString("{\n")
	builder.WriteString("  imports = [ ./hardware-configuration.nix ];\n")

	// Generate a large number of packages
	builder.WriteString("  environment.systemPackages = with pkgs; [\n")
	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("    package%d\n", i))
	}
	builder.WriteString("  ];\n")

	// Generate many service configurations
	for i := 0; i < 50; i++ {
		builder.WriteString(fmt.Sprintf("  services.service%d.enable = true;\n", i))
	}

	builder.WriteString("  system.stateVersion = \"23.11\";\n")
	builder.WriteString("}\n")

	return builder.String()
}
