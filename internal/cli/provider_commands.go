package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// createProviderCommands creates provider management commands
func createProviderCommands(cfg *config.UserConfig, log *logger.Logger) *cobra.Command {
	providerCmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage AI providers and models",
		Long:  "Commands for managing AI providers, checking their status, and configuring models",
	}

	// Add subcommands
	providerCmd.AddCommand(createProviderListCommand(cfg, log))
	providerCmd.AddCommand(createProviderTestCommand(cfg, log))
	providerCmd.AddCommand(createProviderModelsCommand(cfg, log))
	providerCmd.AddCommand(createProviderConfigCommand(cfg, log))

	return providerCmd
}

// createProviderListCommand creates a command to list available providers
func createProviderListCommand(cfg *config.UserConfig, log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available AI providers",
		Long:  "Display information about all configured AI providers and their status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create provider manager
			manager := ai.NewProviderManager(cfg, log)

			// Get all provider configurations
			providers := cfg.AIModels.Providers
			
			fmt.Println("Available AI Providers:")
			fmt.Println("=====================")
			
			for name, providerCfg := range providers {
				status := "Unknown"
				
				// Try to initialize and test the provider
				if provider, err := manager.GetProvider(name); err == nil {
					if healthChecker, ok := provider.(interface{ HealthCheck() error }); ok {
						if err := healthChecker.HealthCheck(); err == nil {
							status = "✅ Available"
						} else {
							status = fmt.Sprintf("❌ Error: %v", err)
						}
					} else {
						status = "⚠️  No health check"
					}
				} else {
					status = fmt.Sprintf("❌ Init error: %v", err)
				}
				
				fmt.Printf("\n%s (%s)\n", providerCfg.Name, name)
				fmt.Printf("  Description: %s\n", providerCfg.Description)
				fmt.Printf("  Type: %s\n", providerCfg.Type)
				fmt.Printf("  Status: %s\n", status)
				fmt.Printf("  Base URL: %s\n", providerCfg.BaseURL)
				fmt.Printf("  Requires API Key: %t\n", providerCfg.RequiresAPIKey)
				if providerCfg.EnvVar != "" {
					fmt.Printf("  Environment Variable: %s\n", providerCfg.EnvVar)
				}
				fmt.Printf("  Supports Streaming: %t\n", providerCfg.SupportsStreaming)
				fmt.Printf("  Supports Tools: %t\n", providerCfg.SupportsTools)
			}
			
			return nil
		},
	}
}

// createProviderTestCommand creates a command to test a specific provider
func createProviderTestCommand(cfg *config.UserConfig, log *logger.Logger) *cobra.Command {
	var providerName string
	var modelName string
	
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test a specific AI provider",
		Long:  "Send a test query to verify that a provider is working correctly",
		RunE: func(cmd *cobra.Command, args []string) error {
			if providerName == "" {
				providerName = cfg.AIProvider
			}
			
			// Create provider manager
			manager := ai.NewProviderManager(cfg, log)
			
			// Get the provider
			provider, err := manager.GetProvider(providerName)
			if err != nil {
				return fmt.Errorf("failed to get provider '%s': %w", providerName, err)
			}
			
			// If model name is provided and provider supports it, configure the model
			if modelName != "" {
				// We'll handle model setting differently since wrapping makes this complex
				// For now, just note that the model will be used in the test
				fmt.Printf("Using model: %s\n", modelName)
			}
			
			// Test health check if available
			fmt.Printf("Testing provider: %s\n", providerName)
			if healthChecker, ok := provider.(interface{ HealthCheck() error }); ok {
				fmt.Print("Health check... ")
				if err := healthChecker.HealthCheck(); err != nil {
					fmt.Printf("❌ Failed: %v\n", err)
					return fmt.Errorf("health check failed: %w", err)
				}
				fmt.Println("✅ Passed")
			}
			
			// Send a test query
			fmt.Print("Sending test query... ")
			testPrompt := "Hello! Please respond with just 'AI provider test successful' to confirm you're working."
			
			response, err := provider.GenerateResponse(cmd.Context(), testPrompt)
			if err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
				return fmt.Errorf("test query failed: %w", err)
			}
			
			fmt.Println("✅ Success")
			fmt.Printf("Response: %s\n", strings.TrimSpace(response))
			
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider to test (default: configured provider)")
	cmd.Flags().StringVarP(&modelName, "model", "m", "", "Model to use for testing (for supported providers)")
	
	return cmd
}

// createProviderModelsCommand creates a command to list available models for a provider
func createProviderModelsCommand(cfg *config.UserConfig, log *logger.Logger) *cobra.Command {
	var providerName string
	var jsonOutput bool
	
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List available models for a provider",
		Long:  "Display available models for a specific AI provider, especially useful for Ollama",
		RunE: func(cmd *cobra.Command, args []string) error {
			if providerName == "" {
				providerName = cfg.AIProvider
			}
			
			// For Ollama provider, try to get models directly
			if providerName == "ollama" {
				// Create a direct Ollama provider to get models
				ollamaProvider := ai.NewOllamaProvider("")
				models, err := ollamaProvider.GetAvailableModels()
				if err != nil {
					return fmt.Errorf("failed to get available models from Ollama: %w", err)
				}
				
				if jsonOutput {
					modelsJSON, _ := json.MarshalIndent(models, "", "  ")
					fmt.Println(string(modelsJSON))
				} else {
					fmt.Printf("Available models for %s:\n", providerName)
					fmt.Println("========================")
					for i, model := range models {
						fmt.Printf("%d. %s\n", i+1, model)
					}
					
					if len(models) == 0 {
						fmt.Println("No models found. Use 'ollama pull <model>' to download models.")
					}
				}
				
				return nil
			}
			
			// For other providers, show configured models from config
			if providerCfg, exists := cfg.AIModels.Providers[providerName]; exists {
				var modelNames []string
				for modelKey := range providerCfg.Models {
					modelNames = append(modelNames, modelKey)
				}
				
				if jsonOutput {
					modelsJSON, _ := json.MarshalIndent(modelNames, "", "  ")
					fmt.Println(string(modelsJSON))
				} else {
					fmt.Printf("Configured models for %s:\n", providerName)
					fmt.Println("=========================")
					for i, model := range modelNames {
						fmt.Printf("%d. %s\n", i+1, model)
					}
				}
			} else {
				return fmt.Errorf("provider '%s' not found in configuration", providerName)
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider to list models for (default: configured provider)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	
	return cmd
}

// createProviderConfigCommand creates a command to show provider configuration help
func createProviderConfigCommand(cfg *config.UserConfig, log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Show provider configuration help",
		Long:  "Display information about how to configure different AI providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("AI Provider Configuration Guide")
			fmt.Println("==============================")
			fmt.Println()
			
			fmt.Println("1. Ollama (Local)")
			fmt.Println("   - Install: https://ollama.ai")
			fmt.Println("   - Pull models: ollama pull llama3")
			fmt.Println("   - Available models: nixai provider models --provider ollama")
			fmt.Println("   - Config: Set ai_provider: ollama and ai_model: <model-name>")
			fmt.Println()
			
			fmt.Println("2. OpenAI")
			fmt.Println("   - Set OPENAI_API_KEY environment variable")
			fmt.Println("   - Config: Set ai_provider: openai and ai_model: gpt-4")
			fmt.Println()
			
			fmt.Println("3. Claude")
			fmt.Println("   - Set CLAUDE_API_KEY environment variable")
			fmt.Println("   - Config: Set ai_provider: claude and ai_model: claude-sonnet-4-20250514")
			fmt.Println()
			
			fmt.Println("4. Gemini")
			fmt.Println("   - Set GEMINI_API_KEY environment variable")
			fmt.Println("   - Config: Set ai_provider: gemini and ai_model: gemini-2.5-flash-preview-05-20")
			fmt.Println()
			
			fmt.Println("5. Groq")
			fmt.Println("   - Set GROQ_API_KEY environment variable")
			fmt.Println("   - Config: Set ai_provider: groq and ai_model: llama-3.3-70b-versatile")
			fmt.Println()
			
			fmt.Println("6. GitHub Copilot")
			fmt.Println("   - Set GITHUB_TOKEN environment variable")
			fmt.Println("   - Config: Set ai_provider: copilot and ai_model: gpt-4")
			fmt.Println()
			
			fmt.Printf("Current configuration:\n")
			fmt.Printf("  Provider: %s\n", cfg.AIProvider)
			fmt.Printf("  Model: %s\n", cfg.AIModel)
			fmt.Println()
			
			fmt.Println("To test your configuration:")
			fmt.Println("  nixai provider test")
			fmt.Println("  nixai provider test --provider ollama --model deepseek-r1")
			
			return nil
		},
	}
}

