package main

import (
	"fmt"
	"nix-ai-help/internal/plugins"
)

func main() {
	fmt.Println("StateRunning:", plugins.StateRunning)
	fmt.Println("StateStopped:", plugins.StateStopped)
	fmt.Println("StateError:", plugins.StateError)
	fmt.Println("StateDisabled:", plugins.StateDisabled)
}
