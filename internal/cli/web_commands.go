package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/config"
	"nix-ai-help/internal/fleet"
	nixosrepo "nix-ai-help/internal/repository"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/internal/web"
	"nix-ai-help/pkg/logger"

	"github.com/spf13/cobra"
)

// WebCommands provides web dashboard CLI commands
type WebCommands struct {
	logger *logger.Logger
	config *config.Config
}

// NewWebCommands creates web dashboard commands
func NewWebCommands(logger *logger.Logger, config *config.Config) *WebCommands {
	return &WebCommands{
		logger: logger,
		config: config,
	}
}

// CreateCommand creates the main web command
func (wc *WebCommands) CreateCommand() *cobra.Command {
	var (
		port     int
		host     string
		repoPath string
	)

	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Start the nixai web dashboard with repository analysis",
		Long: `Start the nixai web dashboard to manage NixOS configurations and fleet machines.
The dashboard provides a visual interface for configuration management, machine monitoring,
and repository analysis.`,
		Example: `  # Start web dashboard on default port
  nixai web

  # Start on custom port with repository
  nixai web --port 9090 --repo /path/to/nixos-config

  # Start with specific host binding
  nixai web --host 0.0.0.0 --port 8080`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return wc.runWebDashboard(port, host, repoPath)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the web dashboard on")
	cmd.Flags().StringVar(&host, "host", "localhost", "Host to bind the web dashboard to")
	cmd.Flags().StringVarP(&repoPath, "repo", "r", "", "Path to NixOS configuration repository to analyze")

	return cmd
}

// runWebDashboard starts the web dashboard server
func (wc *WebCommands) runWebDashboard(port int, host, repoPath string) error {
	wc.logger.Info("Starting nixai web dashboard...")

	// Initialize fleet manager
	fleetManager := fleet.NewFleetManager(wc.logger)

	// Initialize repository parser if repo path provided
	var nixosRepo *nixosrepo.NixOSRepository
	if repoPath != "" {
		wc.logger.Info(fmt.Sprintf("Analyzing NixOS repository: %s", repoPath))

		repo, err := nixosrepo.NewNixOSRepository(repoPath, wc.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}

		if err := repo.ScanRepository(); err != nil {
			return fmt.Errorf("failed to scan repository: %w", err)
		}

		nixosRepo = repo

		// Extract machine definitions from repository and add to fleet
		machines, err := repo.GetMachineDefinitions()
		if err != nil {
			wc.logger.Warn(fmt.Sprintf("Failed to extract machine definitions: %v", err))
		} else {
			ctx := context.Background()
			for _, machine := range machines {
				if err := fleetManager.AddRepositoryMachine(ctx, machine); err != nil {
					wc.logger.Warn(fmt.Sprintf("Failed to add machine %s to fleet: %v", machine.ID, err))
				} else {
					wc.logger.Info(fmt.Sprintf("Added machine %s to fleet from repository", machine.ID))
				}
			}
		}
	} else {
		wc.logger.Warn("No --repo provided: repository features will be unavailable and warnings may appear in logs.")
	}

	// Create team manager
	teamManager := team.NewTeamManager(wc.logger)

	// Create config repository
	configRepo, err := repository.NewConfigRepository("/tmp/nixai-configs", wc.logger)
	if err != nil {
		wc.logger.Warn(fmt.Sprintf("Failed to create config repository: %v", err))
	}

	// Create enhanced web server with repository support and existing fleet manager
	var server *web.EnhancedServer

	if nixosRepo != nil {
		server, err = web.NewEnhancedServerWithFleetAndRepository(port, teamManager, configRepo, fleetManager, nixosRepo, wc.logger)
	} else {
		server, err = web.NewEnhancedServerWithFleetAndRepository(port, teamManager, configRepo, fleetManager, nil, wc.logger)
	}

	if err != nil {
		return fmt.Errorf("failed to create web server: %w", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to listen for interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", host, port)
		wc.logger.Info(fmt.Sprintf("Web dashboard starting on http://%s", addr))
		wc.logger.Info("Press Ctrl+C to stop the server")

		if repoPath != "" {
			wc.logger.Info(fmt.Sprintf("Repository analysis: %d configurations found", len(nixosRepo.GetConfigurations())))
		}

		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			wc.logger.Error(fmt.Sprintf("Web server error: %v", err))
		}
	}()

	// Wait for interrupt signal
	<-interrupt
	wc.logger.Info("Shutting down web dashboard...")

	// Shutdown server
	server.Stop()

	wc.logger.Info("Web dashboard stopped successfully")
	return nil
}
