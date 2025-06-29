package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/fleet"
	"nix-ai-help/internal/plugins"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/internal/web"
	"nix-ai-help/internal/webui/config_builder"
	"nix-ai-help/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAIProvider provides a mock AI provider for testing
type MockAIProvider struct{}

func (m *MockAIProvider) Query(ctx context.Context, query string) (string, error) {
	return `
	{
		system.stateVersion = "24.05";
		boot.loader.grub.enable = true;
		networking.hostName = "test-machine";
		
		services.openssh = {
			enable = true;
			settings.PasswordAuthentication = false;
		};
		
		environment.systemPackages = with pkgs; [
			git
			vim
			curl
		];
	}`, nil
}

func (m *MockAIProvider) GenerateResponse(ctx context.Context, request ai.GenerateRequest) (*ai.GenerateResponse, error) {
	return &ai.GenerateResponse{
		Content:      "Mock AI response for testing",
		Usage:        ai.Usage{PromptTokens: 100, CompletionTokens: 50},
		FinishReason: "stop",
	}, nil
}

func TestIntegrationService(t *testing.T) {
	// Setup test components
	logger := logger.NewLogger()

	// Create temporary directory for repository
	tempDir := t.TempDir()

	// Initialize components
	mockAI := &MockAIProvider{}
	fleetManager := fleet.NewFleetManager(logger)
	pluginManager := plugins.NewManager(tempDir, logger)
	teamManager := team.NewTeamManager(logger)
	configRepo, err := repository.NewConfigRepository(tempDir, logger)
	require.NoError(t, err)

	webServer := web.NewServer(8080, logger)
	configBuilder := config_builder.NewComponentLibrary(logger)

	// Create integration service
	service := NewService(
		mockAI,
		fleetManager,
		pluginManager,
		teamManager,
		configRepo,
		webServer,
		configBuilder,
		logger,
	)

	ctx := context.Background()

	t.Run("AI Configuration Generation", func(t *testing.T) {
		request := AIConfigRequest{
			Type:        "server",
			Description: "Basic web server with SSH access",
			Services:    []string{"openssh", "nginx"},
			Packages:    []string{"git", "curl"},
			Environment: "production",
		}

		response, err := service.GenerateConfigurationWithAI(ctx, request)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Configuration)
		assert.NotEmpty(t, response.Branch)
		assert.NotEmpty(t, response.CommitHash)
	})

	t.Run("Fleet Deployment", func(t *testing.T) {
		// Add test machine to fleet
		machine := &fleet.Machine{
			ID:          "test-machine-01",
			Name:        "Test Machine 1",
			Address:     "192.168.1.100",
			Environment: "test",
			SSHConfig: fleet.SSHConfig{
				User: "root",
				Port: 22,
			},
		}

		err := fleetManager.AddMachine(ctx, machine)
		require.NoError(t, err)

		// Create deployment request
		request := FleetDeployRequest{
			Name:            "test-deployment",
			ConfigHash:      "test-config-hash",
			Targets:         []string{"test-machine-01"},
			CreatedBy:       "test-user",
			RollbackEnabled: true,
			AutoStart:       false,
		}

		deployment, err := service.DeployConfigurationToFleet(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, "test-deployment", deployment.Name)
		assert.Equal(t, "test-config-hash", deployment.ConfigHash)
		assert.Len(t, deployment.Targets, 1)
	})

	t.Run("Team Management", func(t *testing.T) {
		// Create test team
		team := &team.Team{
			ID:          "test-team",
			Name:        "Test Team",
			Description: "Team for testing",
			CreatedBy:   "test-user",
		}

		err := teamManager.CreateTeam(ctx, team)
		require.NoError(t, err)

		// Add user to team
		err = teamManager.AddUser(ctx, "test-team", "test-user", "owner")
		require.NoError(t, err)

		// Create collaborative session
		request := CollabSessionRequest{
			ConfigHash: "test-config-hash",
			TeamID:     "test-team",
			UserID:     "test-user",
		}

		session, err := service.CreateCollaborativeSession(ctx, request)
		require.NoError(t, err)
		assert.Equal(t, "test-config-hash", session.ConfigHash)
		assert.Equal(t, "test-team", session.TeamID)
		assert.Equal(t, "active", session.Status)
	})

	t.Run("Version Control Integration", func(t *testing.T) {
		// Test configuration commit
		config := `{ system.stateVersion = "24.05"; }`
		commitHash, err := configRepo.Commit(ctx, config, "Test configuration", "test-user")
		require.NoError(t, err)
		assert.NotEmpty(t, commitHash)

		// Test branch creation
		branch, err := configRepo.CreateBranch(ctx, "feature/test", "Test branch")
		require.NoError(t, err)
		assert.Equal(t, "feature/test", branch.Name)

		// Test configuration retrieval
		exists, err := configRepo.HasCommit(ctx, commitHash)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestFleetManagement(t *testing.T) {
	logger := logger.NewLogger()
	fleetManager := fleet.NewFleetManager(logger)
	ctx := context.Background()

	t.Run("Machine Management", func(t *testing.T) {
		// Add machine
		machine := &fleet.Machine{
			ID:          "test-machine",
			Name:        "Test Machine",
			Address:     "192.168.1.100",
			Environment: "test",
			SSHConfig: fleet.SSHConfig{
				User: "root",
				Port: 22,
			},
		}

		err := fleetManager.AddMachine(ctx, machine)
		require.NoError(t, err)

		// List machines
		machines, err := fleetManager.ListMachines(ctx)
		require.NoError(t, err)
		assert.Len(t, machines, 1)
		assert.Equal(t, "test-machine", machines[0].ID)

		// Get specific machine
		retrieved, err := fleetManager.GetMachine(ctx, "test-machine")
		require.NoError(t, err)
		assert.Equal(t, machine.Name, retrieved.Name)

		// Remove machine
		err = fleetManager.RemoveMachine(ctx, "test-machine")
		require.NoError(t, err)

		// Verify removal
		machines, err = fleetManager.ListMachines(ctx)
		require.NoError(t, err)
		assert.Len(t, machines, 0)
	})

	t.Run("Deployment Management", func(t *testing.T) {
		// Add test machines
		for i := 1; i <= 3; i++ {
			machine := &fleet.Machine{
				ID:          fmt.Sprintf("machine-%d", i),
				Name:        fmt.Sprintf("Machine %d", i),
				Address:     fmt.Sprintf("192.168.1.%d", 100+i),
				Environment: "test",
				SSHConfig: fleet.SSHConfig{
					User: "root",
					Port: 22,
				},
			}
			err := fleetManager.AddMachine(ctx, machine)
			require.NoError(t, err)
		}

		// Create deployment
		req := fleet.DeploymentRequest{
			Name:       "test-deployment",
			ConfigHash: "test-hash",
			Targets:    []string{"machine-1", "machine-2"},
			Strategy: fleet.DeploymentStrategy{
				Type:      "rolling",
				BatchSize: 1,
			},
			CreatedBy: "test-user",
		}

		deployment, err := fleetManager.CreateDeployment(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, fleet.DeploymentStatusPending, deployment.Status)

		// List deployments
		deployments, err := fleetManager.ListDeployments(ctx)
		require.NoError(t, err)
		assert.Len(t, deployments, 1)

		// Get deployment
		retrieved, err := fleetManager.GetDeployment(ctx, deployment.ID)
		require.NoError(t, err)
		assert.Equal(t, deployment.Name, retrieved.Name)
	})
}

func TestFleetMonitoring(t *testing.T) {
	logger := logger.NewLogger()
	fleetManager := fleet.NewFleetManager(logger)
	monitor := fleet.NewMonitor(fleetManager)
	ctx := context.Background()

	t.Run("Health Monitoring", func(t *testing.T) {
		// Add test machine
		machine := &fleet.Machine{
			ID:          "test-machine",
			Name:        "Test Machine",
			Address:     "192.168.1.100",
			Environment: "production",
			SSHConfig: fleet.SSHConfig{
				User: "root",
				Port: 22,
			},
		}

		err := fleetManager.AddMachine(ctx, machine)
		require.NoError(t, err)

		// Update health status
		health := fleet.HealthStatus{
			Overall:   "healthy",
			LastCheck: time.Now(),
			CPU: fleet.ResourceHealth{
				Status:    "healthy",
				Usage:     45.5,
				Threshold: 80.0,
			},
			Memory: fleet.ResourceHealth{
				Status:    "healthy",
				Usage:     60.2,
				Threshold: 85.0,
			},
		}

		err = fleetManager.UpdateMachineHealth(ctx, "test-machine", health)
		require.NoError(t, err)

		// Get fleet health
		fleetHealth, err := monitor.GetFleetHealth(ctx)
		require.NoError(t, err)
		assert.Equal(t, "healthy", fleetHealth.OverallStatus)
		assert.Equal(t, 1, fleetHealth.TotalMachines)
		assert.Equal(t, 1, fleetHealth.OnlineMachines)
	})
}

func BenchmarkIntegrationService(b *testing.B) {
	logger := logger.NewLogger()
	tempDir := b.TempDir()

	mockAI := &MockAIProvider{}
	fleetManager := fleet.NewFleetManager(logger)
	pluginManager := plugins.NewManager(tempDir, logger)
	teamManager := team.NewTeamManager(logger)
	configRepo, _ := repository.NewConfigRepository(tempDir, logger)
	webServer := web.NewServer(8080, logger)
	configBuilder := config_builder.NewComponentLibrary(logger)

	service := NewService(
		mockAI,
		fleetManager,
		pluginManager,
		teamManager,
		configRepo,
		webServer,
		configBuilder,
		logger,
	)

	ctx := context.Background()

	b.Run("AI Configuration Generation", func(b *testing.B) {
		request := AIConfigRequest{
			Type:        "server",
			Description: "Basic web server",
			Services:    []string{"openssh"},
			Environment: "production",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.GenerateConfigurationWithAI(ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
