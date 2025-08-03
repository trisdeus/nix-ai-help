package main

import (
	"log"
	"nix-ai-help/internal/cli"
	"nix-ai-help/internal/plugins"
	"nix-ai-help/pkg/logger"
	"os"
)

func main() {
	// Ensure all logs go to stderr to avoid polluting HTTP responses
	log.SetOutput(os.Stderr)
	
	// Create a logger
	logger := logger.NewLogger()
	
	// Initialize the plugin system
	_ = plugins.NewManager(nil, logger) // Will be properly initialized in cli.Execute()
	
	// Start the main application logic (calls CLI root command)
	cli.Execute()
}
