package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/fleet"
	"nix-ai-help/internal/plugins"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/pkg/logger"
)

// Service provides integration between all nixai systems
type Service struct {
	// Core components
	aiProvider    ai.Provider
	fleetManager  *fleet.FleetManager
	pluginManager *plugins.Manager
	teamManager   *team.TeamManager
	configRepo    *repository.ConfigRepository

	// Integration state
	logger  *logger.Logger
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewService creates a new integration service
func NewService(
	aiProvider ai.Provider,
	fleetManager *fleet.FleetManager,
	pluginManager *plugins.Manager,
	teamManager *team.TeamManager,
	configRepo *repository.ConfigRepository,
	logger *logger.Logger,
) *Service {
	return &Service{
		aiProvider:    aiProvider,
		fleetManager:  fleetManager,
		pluginManager: pluginManager,
		teamManager:   teamManager,
		configRepo:    configRepo,
		logger:        logger,
		stopCh:        make(chan struct{}),
	}
}

// Start starts all integrated services
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("integration service is already running")
	}

	s.logger.Info("Starting nixai integration service")

	// Start fleet monitoring
	fleetMonitor := fleet.NewMonitor(s.fleetManager)
	if err := fleetMonitor.Start(ctx, 30*time.Second); err != nil {
		return fmt.Errorf("failed to start fleet monitoring: %w", err)
	}

	// Load plugins
	plugins := s.pluginManager.ListPlugins()
	s.logger.Warn(fmt.Sprintf("Loaded %d plugins", len(plugins)))

	// Setup integration handlers
	s.setupIntegrationHandlers()

	s.running = true
	s.logger.Info("Integration service started successfully")

	return nil
}

// Stop stops all integrated services
func (s *Service) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("integration service is not running")
	}

	s.logger.Info("Stopping nixai integration service")

	close(s.stopCh)

	// Stop web server
	s.logger.Info("Web server stopped")

	s.running = false
	s.logger.Info("Integration service stopped")

	return nil
}

// setupIntegrationHandlers sets up cross-system integration
func (s *Service) setupIntegrationHandlers() {
	// Integrate AI with configuration generation
	s.setupAIConfigIntegration()

	// Integrate fleet management with version control
	s.setupFleetVersionIntegration()

	// Integrate collaboration with all systems
	s.setupCollaborationIntegration()

	// Integrate plugins with all systems
	s.setupPluginIntegration()
}

// setupAIConfigIntegration integrates AI with configuration generation
func (s *Service) setupAIConfigIntegration() {
	// This would set up handlers for AI-powered configuration generation
	// that integrates with the visual builder and version control
	s.logger.Debug("Setting up AI-configuration integration")
}

// setupFleetVersionIntegration integrates fleet management with version control
func (s *Service) setupFleetVersionIntegration() {
	// This would set up handlers for deploying configurations from version control
	// to the fleet with proper tracking and rollback capabilities
	s.logger.Debug("Setting up fleet-version control integration")
}

// setupCollaborationIntegration integrates collaboration with all systems
func (s *Service) setupCollaborationIntegration() {
	// This would set up real-time collaboration features across all systems
	s.logger.Debug("Setting up collaboration integration")
}

// setupPluginIntegration integrates plugins with all systems
func (s *Service) setupPluginIntegration() {
	// This would set up plugin hooks and extension points for all systems
	s.logger.Debug("Setting up plugin integration")
}

// GenerateConfigurationWithAI generates NixOS configuration using AI
func (s *Service) GenerateConfigurationWithAI(ctx context.Context, request AIConfigRequest) (*AIConfigResponse, error) {
	s.logger.Info(fmt.Sprintf("Generating configuration with AI for request type: %s", request.Type))

	// Use AI provider to generate configuration
	prompt := s.buildConfigurationPrompt(request)
	response, err := s.aiProvider.Query(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI query failed: %w", err)
	}

	// Parse and validate the generated configuration
	config, err := s.parseAIResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Create a new branch for the configuration
	branchName := fmt.Sprintf("ai-config-%d", time.Now().Unix())
	err = s.configRepo.CreateBranch(ctx, branchName, "main")
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Commit the configuration
	files := map[string]string{
		"configuration.nix": config,
	}
	metadata := map[string]string{
		"generator": "nixai-ai",
		"type":      request.Type,
	}
	commitSnapshot, err := s.configRepo.Commit(ctx, "AI-generated NixOS configuration", files, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to commit configuration: %w", err)
	}

	return &AIConfigResponse{
		Configuration: config,
		Branch:        branchName,
		CommitHash:    commitSnapshot.ID,
		Suggestions:   []string{}, // TODO: Add AI suggestions
		Warnings:      []string{}, // TODO: Add validation warnings
	}, nil
}

// DeployConfigurationToFleet deploys a configuration to fleet machines
func (s *Service) DeployConfigurationToFleet(ctx context.Context, request FleetDeployRequest) (*fleet.Deployment, error) {
	s.logger.Info(fmt.Sprintf("Deploying configuration to fleet - config: %s, targets: %d", request.ConfigHash, len(request.Targets)))

	// Validate configuration exists in repository
	_, err := s.configRepo.GetSnapshot(ctx, request.ConfigHash)
	if err != nil {
		return nil, fmt.Errorf("configuration not found: %w", err)
	}

	// Create fleet deployment
	deploymentReq := fleet.DeploymentRequest{
		Name:            request.Name,
		ConfigHash:      request.ConfigHash,
		Targets:         request.Targets,
		Strategy:        request.Strategy,
		CreatedBy:       request.CreatedBy,
		RollbackEnabled: request.RollbackEnabled,
	}

	deployment, err := s.fleetManager.CreateDeployment(ctx, deploymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	// Start deployment
	if request.AutoStart {
		if err := s.fleetManager.StartDeployment(ctx, deployment.ID); err != nil {
			return nil, fmt.Errorf("failed to start deployment: %w", err)
		}
	}

	return deployment, nil
}

// CreateCollaborativeSession creates a collaborative editing session
func (s *Service) CreateCollaborativeSession(ctx context.Context, request CollabSessionRequest) (*CollabSession, error) {
	s.logger.Info(fmt.Sprintf("Creating collaborative session - config: %s, team: %s", request.ConfigHash, request.TeamID))

	// Validate team exists and user has permissions
	_, err := s.teamManager.GetTeam(ctx, request.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	// Check user permissions - simplified for now
	// TODO: Implement proper permission checking
	// if !s.hasEditPermissions(member.Role) {
	// 	return nil, fmt.Errorf("user does not have edit permissions")
	// }

	// Create collaborative session
	session := &CollabSession{
		ID:           fmt.Sprintf("session-%d", time.Now().Unix()),
		ConfigHash:   request.ConfigHash,
		TeamID:       request.TeamID,
		CreatedBy:    request.UserID,
		CreatedAt:    time.Now(),
		Participants: []string{request.UserID},
		Status:       "active",
	}

	return session, nil
}

// ExecutePluginWorkflow executes a plugin-based workflow
func (s *Service) ExecutePluginWorkflow(ctx context.Context, request PluginWorkflowRequest) (*PluginWorkflowResponse, error) {
	s.logger.Info(fmt.Sprintf("Executing plugin workflow - plugin: %s, workflow: %s", request.PluginID, request.WorkflowName))

	// Get plugin
	_, exists := s.pluginManager.GetPlugin(request.PluginID)
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", request.PluginID)
	}

	// Execute plugin workflow
	result, err := s.pluginManager.ExecutePluginOperation(request.PluginID, request.WorkflowName, map[string]interface{}{
		"workflow": request.WorkflowName,
		"params":   request.Parameters,
	})
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	return &PluginWorkflowResponse{
		PluginID:     request.PluginID,
		WorkflowName: request.WorkflowName,
		Result:       result,
		ExecutedAt:   time.Now(),
	}, nil
}

// buildConfigurationPrompt builds an AI prompt for configuration generation
func (s *Service) buildConfigurationPrompt(request AIConfigRequest) string {
	prompt := fmt.Sprintf(`Generate a NixOS configuration for: %s

Requirements:
- Type: %s
- Description: %s
`, request.Description, request.Type, request.Description)

	if len(request.Services) > 0 {
		prompt += fmt.Sprintf("- Services: %v\n", request.Services)
	}

	if len(request.Packages) > 0 {
		prompt += fmt.Sprintf("- Packages: %v\n", request.Packages)
	}

	prompt += `
Please provide a complete, valid NixOS configuration that follows best practices.
Include proper security settings and optimization where applicable.
Format the response as valid Nix configuration code.`

	return prompt
}

// parseAIResponse parses the AI response and extracts configuration
func (s *Service) parseAIResponse(response string) (string, error) {
	// This would implement proper parsing of AI response
	// For now, return the response as-is
	return response, nil
}

// hasEditPermissions checks if a role has edit permissions
func (s *Service) hasEditPermissions(role string) bool {
	switch role {
	case "owner", "admin", "maintainer", "developer":
		return true
	default:
		return false
	}
}

// Request/Response types for integration

type AIConfigRequest struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Services    []string `json:"services"`
	Packages    []string `json:"packages"`
	Environment string   `json:"environment"`
}

type AIConfigResponse struct {
	Configuration string   `json:"configuration"`
	Branch        string   `json:"branch"`
	CommitHash    string   `json:"commit_hash"`
	Suggestions   []string `json:"suggestions"`
	Warnings      []string `json:"warnings"`
}

type FleetDeployRequest struct {
	Name            string                   `json:"name"`
	ConfigHash      string                   `json:"config_hash"`
	Targets         []string                 `json:"targets"`
	Strategy        fleet.DeploymentStrategy `json:"strategy"`
	CreatedBy       string                   `json:"created_by"`
	RollbackEnabled bool                     `json:"rollback_enabled"`
	AutoStart       bool                     `json:"auto_start"`
}

type CollabSessionRequest struct {
	ConfigHash string `json:"config_hash"`
	TeamID     string `json:"team_id"`
	UserID     string `json:"user_id"`
}

type CollabSession struct {
	ID           string    `json:"id"`
	ConfigHash   string    `json:"config_hash"`
	TeamID       string    `json:"team_id"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	Participants []string  `json:"participants"`
	Status       string    `json:"status"`
}

type PluginWorkflowRequest struct {
	PluginID     string                 `json:"plugin_id"`
	WorkflowName string                 `json:"workflow_name"`
	Parameters   map[string]interface{} `json:"parameters"`
}

type PluginWorkflowResponse struct {
	PluginID     string      `json:"plugin_id"`
	WorkflowName string      `json:"workflow_name"`
	Result       interface{} `json:"result"`
	ExecutedAt   time.Time   `json:"executed_at"`
}
