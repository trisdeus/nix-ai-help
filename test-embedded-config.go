package main

import (
	"fmt"
	"nix-ai-help/internal/config"
)

func main() {
	cfg, err := config.LoadEmbeddedYAMLConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Embedded MCP Port: %d\n", cfg.MCPServer.MCPPort)

	userCfg := config.DefaultUserConfig()
	fmt.Printf("Default User MCP Port: %d\n", userCfg.MCPServer.MCPPort)
}
