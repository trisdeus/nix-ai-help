package plugins

import (
	"context"
	"fmt"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"

	"github.com/spf13/cobra"
)

// SimplePluginIntegration handles our successful plugin integration approach
type SimplePluginIntegration struct {
	rootCmd *cobra.Command
	config  *config.UserConfig
	logger  *logger.Logger
}

// NewSimplePluginIntegration creates a new simple plugin integration
func NewSimplePluginIntegration(rootCmd *cobra.Command, cfg *config.UserConfig, log *logger.Logger) *SimplePluginIntegration {
	return &SimplePluginIntegration{
		rootCmd: rootCmd,
		config:  cfg,
		logger:  log,
	}
}

// Initialize sets up the integrated plugin commands
func (spi *SimplePluginIntegration) Initialize(ctx context.Context) error {
	spi.logger.Info("Initializing integrated plugin commands")
	
	// Our integrated plugins are already added in commands.go:
	// - CreateSystemInfoCommand() -> system-info
	// - CreatePackageMonitorCommand() -> package-monitor
	
	// These are now built-in commands, not external plugins
	// This demonstrates the successful "plugin as integrated command" approach
	
	spi.logger.Info("Integrated plugin commands available: system-info, package-monitor")
	return nil
}

// GetIntegratedCommands returns information about integrated plugin commands
func (spi *SimplePluginIntegration) GetIntegratedCommands() []IntegratedPluginInfo {
	return []IntegratedPluginInfo{
		{
			Name:        "system-info",
			Description: "System information and health monitoring",
			Version:     "1.0.0",
			Author:      "NixAI Team",
			Type:        "built-in",
			Commands: []string{
				"status", "health", "cpu", "memory", "disk", "processes", "monitor", "all",
			},
			Examples: []string{
				"nixai system-info health",
				"nixai system-info status --json",
				"nixai system-info monitor --interval 5",
			},
		},
		{
			Name:        "package-monitor",
			Description: "Package monitoring and update management",
			Version:     "1.0.0",
			Author:      "NixAI Team",
			Type:        "built-in",
			Commands: []string{
				"list", "updates", "security", "analyze", "orphans", "outdated", "stats",
			},
			Examples: []string{
				"nixai package-monitor list --detailed",
				"nixai package-monitor updates --security",
				"nixai package-monitor stats --json",
			},
		},
	}
}

// IntegratedPluginInfo represents information about an integrated plugin command
type IntegratedPluginInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Type        string   `json:"type"`
	Commands    []string `json:"commands"`
	Examples    []string `json:"examples"`
}

// ListIntegratedPlugins shows all available integrated plugin commands
func (spi *SimplePluginIntegration) ListIntegratedPlugins() {
	plugins := spi.GetIntegratedCommands()
	
	fmt.Println("📦 Integrated Plugin Commands:")
	fmt.Println()
	
	for _, plugin := range plugins {
		fmt.Printf("🔧 %s v%s (%s)\n", plugin.Name, plugin.Version, plugin.Type)
		fmt.Printf("   %s\n", plugin.Description)
		fmt.Printf("   Author: %s\n", plugin.Author)
		fmt.Printf("   Subcommands: %v\n", plugin.Commands)
		fmt.Println()
		
		fmt.Println("   Examples:")
		for _, example := range plugin.Examples {
			fmt.Printf("     %s\n", example)
		}
		fmt.Println()
	}
}