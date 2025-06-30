package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"nix-ai-help/internal/fleet"
	"nix-ai-help/internal/webui"
	"nix-ai-help/pkg/logger"

	"github.com/gorilla/mux"
)

func main() {
	// Create logger
	logger := logger.NewLogger()

	// Create fleet manager
	fleetManager := fleet.NewFleetManager(logger)

	// Create webui API
	api, err := webui.NewConfigBuilderAPI(fleetManager, logger)
	if err != nil {
		fmt.Printf("Failed to create API: %v\n", err)
		return
	}

	// Create router and register routes
	router := mux.NewRouter()
	api.RegisterRoutes(router)

	// Create test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test fleet overview endpoint
	resp, err := http.Get(server.URL + "/api/fleet")
	if err != nil {
		fmt.Printf("Failed to call fleet API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Expected status 200, got %d\n", resp.StatusCode)
		return
	}

	// Check if response is valid JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Failed to decode JSON response: %v\n", err)
		return
	}

	fmt.Printf("Fleet overview API test successful!\n")
	fmt.Printf("Response: %+v\n", result)

	// Test fleet machines endpoint
	resp2, err := http.Get(server.URL + "/api/fleet/machines")
	if err != nil {
		fmt.Printf("Failed to call fleet machines API: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	// Check response
	if resp2.StatusCode != http.StatusOK {
		fmt.Printf("Expected status 200, got %d\n", resp2.StatusCode)
		return
	}

	// Check if response is valid JSON
	var machines []interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&machines); err != nil {
		fmt.Printf("Failed to decode JSON response from machines endpoint: %v\n", err)
		return
	}

	fmt.Printf("Fleet machines API test successful!\n")
	fmt.Printf("Machines: %+v\n", machines)
}
