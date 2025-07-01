package main

import (
	"fmt"
	"log"
	"os"

	"nix-ai-help/internal/repository"
	"nix-ai-help/pkg/logger"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run test_config_count.go <repo-path>")
	}

	repoPath := os.Args[1]

	// Initialize logger
	loggerInstance := logger.NewLogger()

	// Create a new repository parser
	repo, err := repository.NewNixOSRepository(repoPath, loggerInstance)

	// Scan the repository
	fmt.Printf("Scanning repository: %s\n", repoPath)
	err = repo.ScanRepository()
	if err != nil {
		log.Fatalf("Failed to scan repository: %v", err)
	}

	// Get configurations
	configs := repo.GetConfigurations()
	fmt.Printf("Found %d configurations:\n", len(configs))

	for name, config := range configs {
		fmt.Printf("  - %s (type: %s, path: %s)\n", name, config.Type, config.Path)
	}

	// Get machines
	machines, err := repo.GetMachineDefinitions()
	if err != nil {
		log.Fatalf("Failed to get machines: %v", err)
	}
	fmt.Printf("Found %d machines:\n", len(machines))

	for _, machine := range machines {
		fmt.Printf("  - %s (address: %s)\n", machine.Name, machine.Address)
	}
}
