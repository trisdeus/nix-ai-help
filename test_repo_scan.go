package main

import (
	"fmt"
	"nix-ai-help/internal/repository"
	"nix-ai-help/pkg/logger"
)

func main() {
	logger := logger.NewLoggerWithLevel("debug")
	repo, err := repository.NewNixOSRepository("/tmp/test-nixos-repo", logger)
	if err != nil {
		fmt.Printf("Error creating repo: %v\n", err)
		return
	}
	err = repo.ScanRepository()
	if err != nil {
		fmt.Printf("Error scanning repo: %v\n", err)
		return
	}
	configs := repo.GetConfigurations()
	fmt.Printf("Found %d configurations:\n", len(configs))
	for name, config := range configs {
		fmt.Printf("- %s: %s (%s)\n", name, config.Path, config.Type)
	}
}
