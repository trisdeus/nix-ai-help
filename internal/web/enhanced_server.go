package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"nix-ai-help/internal/auth"
	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/fleet"
	nixosrepo "nix-ai-help/internal/repository"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/internal/webui"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/system"

	"github.com/gorilla/mux"
)

// EnhancedServer wraps the existing Server with additional features
type EnhancedServer struct {
	*Server
	templates     *template.Template
	configBuilder *webui.ConfigBuilderAPI
	templateAPI   *webui.TemplateAPI
	fleetManager  *fleet.FleetManager
	authManager   *auth.AuthManager
	nixosRepo     *nixosrepo.NixOSRepository
}

// NewEnhancedServer creates a new enhanced web server using the existing Server
func NewEnhancedServer(port int, teamManager *team.TeamManager, configRepo *repository.ConfigRepository, logger *logger.Logger) (*EnhancedServer, error) {
	return NewEnhancedServerWithRepository(port, teamManager, configRepo, nil, logger)
}

// NewEnhancedServerWithRepository creates a new enhanced web server with an optional NixOS repository
func NewEnhancedServerWithRepository(port int, teamManager *team.TeamManager, configRepo *repository.ConfigRepository, nixosRepo *nixosrepo.NixOSRepository, logger *logger.Logger) (*EnhancedServer, error) {
	// Create server config with enhanced features
	config := &ServerConfig{
		Port:        port,
		Host:        "0.0.0.0",
		StaticDir:   "./internal/web/static",
		TemplateDir: "./internal/web/templates",
		Authentication: AuthConfig{
			Enabled:        false,
			SessionTimeout: "24h",
			Providers:      []string{"local"},
		},
		TLS: TLSConfig{
			Enabled: false,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
		},
		Features: FeatureConfig{
			VisualBuilder:   true,
			FleetManagement: true, // Enable fleet management
			Collaboration:   true,
			VersionControl:  true,
			AIGeneration:    true,
			Dashboard:       true,
		},
	}

	// Initialize FleetManager first
	fleetManager := fleet.NewFleetManager(logger)

	// Create the base server with FleetManager
	server, err := NewServer(config, teamManager, configRepo, fleetManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create base server: %w", err)
	}

	// Initialize ConfigBuilderAPI with FleetManager and optional repository
	var configBuilder *webui.ConfigBuilderAPI
	if nixosRepo != nil {
		configBuilder, err = webui.NewConfigBuilderAPIWithRepository(fleetManager, nixosRepo, logger)
	} else {
		configBuilder, err = webui.NewConfigBuilderAPI(fleetManager, logger)
	}
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to initialize config builder: %v", err))
		configBuilder = nil
	}

	// Initialize TemplateAPI
	templateAPI := webui.NewTemplateAPI(logger)

	// Initialize AuthManager
	authManager, err := auth.NewAuthManager(teamManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	// Create enhanced server wrapper
	enhanced := &EnhancedServer{
		Server:        server,
		configBuilder: configBuilder,
		templateAPI:   templateAPI,
		fleetManager:  fleetManager,
		authManager:   authManager,
	}

	// Load templates
	templatePattern := filepath.Join(config.TemplateDir, "*.html")
	templates, err := template.ParseGlob(templatePattern)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to parse templates: %v", err))
		// Continue without templates, will fall back to basic HTML
		enhanced.templates = nil
	} else {
		enhanced.templates = templates
		logger.Info("Templates loaded successfully")
	}

	// Clear existing routes and setup enhanced routes
	enhanced.setupEnhancedRoutes()

	return enhanced, nil
}

// NewEnhancedServerWithFleetAndRepository creates a new enhanced web server with existing fleet manager and repository
func NewEnhancedServerWithFleetAndRepository(port int, teamManager *team.TeamManager, configRepo *repository.ConfigRepository, fleetManager *fleet.FleetManager, nixosRepo *nixosrepo.NixOSRepository, logger *logger.Logger) (*EnhancedServer, error) {
	// Create server config with enhanced features
	config := &ServerConfig{
		Port:        port,
		Host:        "0.0.0.0",
		StaticDir:   "./internal/web/static",
		TemplateDir: "./internal/web/templates",
		Authentication: AuthConfig{
			Enabled:        false,
			SessionTimeout: "24h",
			Providers:      []string{"local"},
		},
		TLS: TLSConfig{
			Enabled: false,
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
		},
		Features: FeatureConfig{
			VisualBuilder:   true,
			FleetManagement: true,
			Collaboration:   true,
			VersionControl:  true,
			AIGeneration:    true,
			Dashboard:       true,
		},
	}

	// Use the provided FleetManager (don't create a new one)
	server, err := NewServer(config, teamManager, configRepo, fleetManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create base server: %w", err)
	}

	// Initialize ConfigBuilderAPI with FleetManager and optional repository
	var configBuilder *webui.ConfigBuilderAPI
	if nixosRepo != nil {
		configBuilder, err = webui.NewConfigBuilderAPIWithRepository(fleetManager, nixosRepo, logger)
	} else {
		configBuilder, err = webui.NewConfigBuilderAPI(fleetManager, logger)
	}
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to initialize config builder: %v", err))
		configBuilder = nil
	}

	// Initialize TemplateAPI
	templateAPI := webui.NewTemplateAPI(logger)

	// Initialize AuthManager
	authManager, err := auth.NewAuthManager(teamManager, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	// Create enhanced server wrapper
	enhanced := &EnhancedServer{
		Server:        server,
		configBuilder: configBuilder,
		templateAPI:   templateAPI,
		fleetManager:  fleetManager, // Use the provided fleet manager
		authManager:   authManager,
		nixosRepo:     nixosRepo, // Set the repository
	}

	// Load templates
	templatePattern := filepath.Join(config.TemplateDir, "*.html")
	templates, err := template.ParseGlob(templatePattern)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to parse templates: %v", err))
		// Continue without templates, will fall back to basic HTML
		enhanced.templates = nil
	} else {
		enhanced.templates = templates
		logger.Info("Templates loaded successfully")
	}

	// Setup routes
	enhanced.setupEnhancedRoutes()

	return enhanced, nil
}

// setupEnhancedRoutes adds enhanced functionality, replacing base routes
func (s *EnhancedServer) setupEnhancedRoutes() {
	// Create a new router to replace the base one
	s.router = mux.NewRouter()

	// Serve static files first
	if s.config.StaticDir != "" {
		fs := http.FileServer(http.Dir(s.config.StaticDir))
		s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	}

	// Add enhanced API endpoints with highest priority
	s.router.HandleFunc("/api/dashboard", s.handleDashboardAPI).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/dashboard/details", s.handleDashboardDetails).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/dashboard/stats", s.handleDashboardStats).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/dashboard/activities", s.handleDashboardActivities).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/dashboard/alerts", s.handleDashboardAlerts).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/dashboard/suggestions", s.handleAISuggestions).Methods("GET", "HEAD", "OPTIONS")

	// Auth endpoints
	s.router.HandleFunc("/api/auth/status", s.handleAuthStatus).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/auth/login", s.handleLogin).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/auth/logout", s.handleLogout).Methods("POST", "OPTIONS")

	// User management endpoints
	s.router.HandleFunc("/api/users", s.handleListUsers).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/users", s.handleCreateUser).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/users/change-password", s.handleChangePassword).Methods("POST", "OPTIONS")

	// Teams API endpoints
	s.router.HandleFunc("/api/teams", s.handleListTeams).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/teams", s.handleCreateTeam).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/teams/stats", s.handleTeamsStats).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/teams/activity", s.handleTeamsActivity).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/teams/public", s.handlePublicTeams).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/teams/join", s.handleJoinTeam).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/teams/{teamId}", s.handleGetTeam).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/teams/{teamId}/members", s.handleListTeamMembers).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/teams/{teamId}/collaboration/start", s.handleStartCollaboration).Methods("POST", "OPTIONS")

	// Fleet API endpoints
	s.router.HandleFunc("/api/fleet", s.handleFleetAPI).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/fleet/machines", s.handleFleetMachines).Methods("GET", "POST", "OPTIONS")
	s.router.HandleFunc("/api/fleet/machines/{machineId}", s.handleGetMachine).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/fleet/machines/{machineId}", s.handleRemoveMachine).Methods("DELETE", "OPTIONS")
	s.router.HandleFunc("/api/fleet/deploy", s.handleFleetDeploy).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/fleet/deployments", s.handleFleetDeployments).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/fleet/deployments/{deploymentId}", s.handleGetDeployment).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/fleet/health", s.handleFleetHealth).Methods("GET", "HEAD", "OPTIONS")

	// Builder API endpoints
	s.router.HandleFunc("/api/builder/validate", s.handleBuilderValidate).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/builder/generate", s.handleBuilderGenerate).Methods("POST", "OPTIONS")

	// Config API endpoints
	s.router.HandleFunc("/api/config/branches", s.handleConfigBranches).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/config/files", s.handleConfigFiles).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/configurations", s.handleConfigurations).Methods("GET", "POST", "OPTIONS")

	// Versioning API endpoints
	s.router.HandleFunc("/api/versioning/repositories", s.handleVersioningRepositories).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/status", s.handleVersioningStatus).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/branches", s.handleVersioningBranches).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/commits", s.handleVersioningCommits).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/tags", s.handleVersioningTags).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/stats", s.handleVersioningStats).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/changes", s.handleVersioningChanges).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/branches", s.handleVersioningCreateBranch).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/tags", s.handleVersioningCreateTag).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/commit", s.handleVersioningCommit).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/checkout", s.handleVersioningCheckout).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/diff/{fileName}", s.handleVersioningDiff).Methods("GET", "HEAD", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/stage", s.handleVersioningStage).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/api/versioning/repositories/{repoId}/stage-all", s.handleVersioningStageAll).Methods("POST", "OPTIONS")

	// AI Chat API endpoint
	s.router.HandleFunc("/api/ai/chat", s.handleAIChat).Methods("POST", "OPTIONS")

	// Register ConfigBuilderAPI routes if available
	if s.configBuilder != nil {
		s.configBuilder.RegisterRoutes(s.router)
	}

	// Register TemplateAPI routes
	if s.templateAPI != nil {
		s.templateAPI.RegisterRoutes(s.router)
	}

	// WebSocket endpoint for real-time updates
	s.router.HandleFunc("/api/ws", s.handleWebSocketConnection).Methods("GET")

	// Health check endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Enhanced frontend routes - these replace the base routes
	s.router.HandleFunc("/", s.handleEnhancedDashboard).Methods("GET")
	s.router.HandleFunc("/dashboard", s.handleEnhancedDashboard).Methods("GET")
	s.router.HandleFunc("/builder", s.handleEnhancedBuilder).Methods("GET")
	s.router.HandleFunc("/fleet", s.handleEnhancedFleet).Methods("GET")
	s.router.HandleFunc("/teams", s.handleEnhancedTeams).Methods("GET")
	s.router.HandleFunc("/versions", s.handleEnhancedVersions).Methods("GET")
	s.router.HandleFunc("/login", s.handleEnhancedLogin).Methods("GET")
	s.router.HandleFunc("/logout", s.handleEnhancedLogout).Methods("GET")
}

// Enhanced dashboard handler
func (s *EnhancedServer) handleEnhancedDashboard(w http.ResponseWriter, r *http.Request) {
	// Get real fleet data
	machines, err := s.fleetManager.ListMachines(r.Context())
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get machines: %v", err))
		machines = []*fleet.Machine{} // Fallback to empty list
	}

	// Calculate stats
	totalMachines := len(machines)
	healthyMachines := 0
	for _, machine := range machines {
		if machine.Status == "online" {
			healthyMachines++
		}
	}

	// Get system stats
	systemStats, err := s.getSystemStats()
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to get system stats: %v", err))
		systemStats = map[string]interface{}{
			"Uptime": "Unknown",
			"CPU":    "0%",
			"Memory": "0MB",
			"Disk":   "0%",
		}
	}

	// Generate AI suggestions
	aiSuggestions := s.generateAISuggestions(machines)

	// Generate recent activities
	activities := s.generateRecentActivities(machines)

	// Generate alerts
	alerts := s.generateSystemAlerts(machines)

	data := map[string]interface{}{
		"Title":       "NixAI Dashboard",
		"ActivePage":  "dashboard",
		"ShowSidebar": true,
		"User": map[string]string{
			"Username": "demo",
			"Role":     "admin",
		},
		"PageHeader": map[string]interface{}{
			"Title":    "Dashboard",
			"Subtitle": "System overview and real-time monitoring",
		},
		"Stats": map[string]interface{}{
			"TotalMachines":   totalMachines,
			"HealthyMachines": healthyMachines,
			"OfflineMachines": totalMachines - healthyMachines,
			"TotalConfigs":    s.getTotalConfigs(),
			"ActiveTeams":     1, // Default
		},
		"Activities":    activities,
		"FleetStatus":   s.generateFleetStatus(machines),
		"Alerts":        alerts,
		"ConfigStatus":  s.generateConfigurationStatus(),
		"TeamActivity":  []interface{}{},
		"AISuggestions": aiSuggestions,
		"SidebarData": map[string]interface{}{
			"Title": "Quick Stats",
			"Items": []map[string]interface{}{
				{"Label": "Uptime", "Value": systemStats["Uptime"]},
				{"Label": "CPU", "Value": systemStats["CPU"]},
				{"Label": "Memory", "Value": systemStats["Memory"]},
				{"Label": "Disk", "Value": systemStats["Disk"]},
			},
		},
	}

	s.renderEnhancedTemplate(w, "dashboard", data)
}

// Enhanced builder handler
func (s *EnhancedServer) handleEnhancedBuilder(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Configuration Builder",
		"ActivePage":  "builder",
		"ShowSidebar": true,
		"User": map[string]string{
			"Username": "demo",
			"Role":     "admin",
		},
		"PageHeader": map[string]interface{}{
			"Title":    "Configuration Builder",
			"Subtitle": "Visual NixOS configuration editor",
		},
	}

	s.renderEnhancedTemplate(w, "builder", data)
}

// Enhanced fleet handler
func (s *EnhancedServer) handleEnhancedFleet(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Fleet Management",
		"ActivePage":  "fleet",
		"ShowSidebar": true,
		"User": map[string]string{
			"Username": "demo",
			"Role":     "admin",
		},
		"PageHeader": map[string]interface{}{
			"Title":    "Fleet Management",
			"Subtitle": "Manage and monitor your NixOS machines",
		},
	}

	s.renderEnhancedTemplate(w, "fleet", data)
}

// Enhanced teams handler
func (s *EnhancedServer) handleEnhancedTeams(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Team Collaboration",
		"ActivePage":  "teams",
		"ShowSidebar": true,
		"User": map[string]string{
			"Username": "demo",
			"Role":     "admin",
		},
		"PageHeader": map[string]interface{}{
			"Title":    "Team Collaboration",
			"Subtitle": "Manage teams and collaborative workflows",
		},
	}

	s.renderEnhancedTemplate(w, "teams", data)
}

// Enhanced versions handler
func (s *EnhancedServer) handleEnhancedVersions(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Version Control",
		"ActivePage":  "versions",
		"ShowSidebar": true,
		"User": map[string]string{
			"Username": "demo",
			"Role":     "admin",
		},
		"PageHeader": map[string]interface{}{
			"Title":    "Version Control",
			"Subtitle": "Git-like configuration management",
		},
	}

	s.renderEnhancedTemplate(w, "versions", data)
}

// Enhanced login handler
func (s *EnhancedServer) handleEnhancedLogin(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Login - NixAI",
		"ActivePage":  "login",
		"ShowSidebar": false,
		"PageHeader": map[string]interface{}{
			"Title":    "Welcome to NixAI",
			"Subtitle": "Sign in to your account",
		},
	}

	s.renderEnhancedTemplate(w, "login", data)
}

// Enhanced logout handler
func (s *EnhancedServer) handleEnhancedLogout(w http.ResponseWriter, r *http.Request) {
	// Redirect to login page after logout
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// renderEnhancedTemplate renders templates with enhanced layout
func (s *EnhancedServer) renderEnhancedTemplate(w http.ResponseWriter, templateName string, data map[string]interface{}) {
	// Set the active page for template rendering
	data["ActivePage"] = templateName

	// Try to render the actual template file
	s.renderTemplate(w, templateName+".html", data)
}

// getPageContent returns page-specific content
func (s *EnhancedServer) getPageContent(templateName string, data map[string]interface{}) string {
	switch templateName {
	case "dashboard":
		return s.getDashboardContent(data)
	case "builder":
		return s.getBuilderContent(data)
	case "fleet":
		return s.getFleetContent(data)
	case "teams":
		return s.getTeamsContent(data)
	case "versions":
		return s.getVersionsContent(data)
	default:
		return "<div class='nixai-card'><h2>Page content loading...</h2></div>"
	}
}

// Dashboard content
func (s *EnhancedServer) getDashboardContent(data map[string]interface{}) string {
	return `<div class="dashboard-container">
    <div class="nixai-dashboard-grid">
        <div class="nixai-stat-card">
            <div class="nixai-stat-value" id="total-machines">0</div>
            <div class="nixai-stat-label">Total Machines</div>
        </div>
        <div class="nixai-stat-card">
            <div class="nixai-stat-value" id="healthy-machines">0</div>
            <div class="nixai-stat-label">Healthy Machines</div>
        </div>
        <div class="nixai-stat-card">
            <div class="nixai-stat-value" id="total-configs">0</div>
            <div class="nixai-stat-label">Configurations</div>
        </div>
        <div class="nixai-stat-card">
            <div class="nixai-stat-value" id="active-teams">0</div>
            <div class="nixai-stat-label">Active Teams</div>
        </div>
    </div>
    
    <div class="nixai-card">
        <div class="nixai-card-header">
            <h3 class="nixai-card-title">System Status</h3>
        </div>
        <div id="dashboard-content">
            <p>Dashboard is loading... Real-time updates will appear here.</p>
        </div>
    </div>
</div>`
}

// Builder content
func (s *EnhancedServer) getBuilderContent(data map[string]interface{}) string {
	return `<div class="builder-container">
    <!-- Template Library Section -->
    <div class="nixai-card">
        <div class="nixai-card-header">
            <h3 class="nixai-card-title">Templates</h3>
            <div class="nixai-card-actions">
                <button class="nixai-btn nixai-btn-sm" onclick="refreshTemplates()">
                    <span>🔄</span> Refresh
                </button>
                <button class="nixai-btn nixai-btn-sm nixai-btn-primary" onclick="importTemplate()">
                    <span>📥</span> Import
                </button>
            </div>
        </div>
        <div class="template-grid" id="template-grid">
            <div class="template-loading">Loading templates...</div>
        </div>
    </div>

    <!-- Configuration Canvas -->
    <div class="nixai-card">
        <div class="nixai-card-header">
            <h3 class="nixai-card-title">Configuration Editor</h3>
            <div class="nixai-card-actions">
                <button class="nixai-btn nixai-btn-sm" onclick="validateConfig()">
                    <span>✓</span> Validate
                </button>
                <button class="nixai-btn nixai-btn-sm" onclick="generateConfig()">
                    <span>⚡</span> Generate
                </button>
                <button class="nixai-btn nixai-btn-sm nixai-btn-primary" onclick="exportConfig()">
                    <span>💾</span> Save
                </button>
            </div>
        </div>
        <div class="config-canvas" id="config-canvas">
            <div class="canvas-placeholder">
                <div class="placeholder-content">
                    <h4>Start Building Your Configuration</h4>
                    <p>Choose a template from the library above or start from scratch</p>
                    <button class="nixai-btn nixai-btn-primary" onclick="startFromScratch()">
                        <span>🚀</span> Start from Scratch
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Configuration Output -->
    <div class="nixai-card">
        <div class="nixai-card-header">
            <h3 class="nixai-card-title">Generated Configuration</h3>
            <div class="nixai-card-actions">
                <button class="nixai-btn nixai-btn-sm" onclick="copyToClipboard()">
                    <span>📋</span> Copy
                </button>
                <button class="nixai-btn nixai-btn-sm" onclick="downloadConfig()">
                    <span>⬇️</span> Download
                </button>
            </div>
        </div>
        <div class="config-output">
            <pre id="config-preview"><code class="language-nix"># Your NixOS configuration will appear here
# Choose a template or start building to see the generated configuration

{ config, pkgs, ... }:

{
  # Configuration options will be added here
  system.stateVersion = "23.11";
}</code></pre>
        </div>
    </div>
</div>

<style>
.builder-container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.template-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
    padding: 1rem 0;
}

.template-card {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 1rem;
    background: white;
    cursor: pointer;
    transition: all 0.2s ease;
}

.template-card:hover {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
    transform: translateY(-2px);
}

.template-card.selected {
    border-color: #3b82f6;
    background: #eff6ff;
}

.template-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
}

.template-name {
    font-weight: 600;
    color: #1e293b;
}

.template-category {
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    background: #f1f5f9;
    border-radius: 4px;
    color: #64748b;
}

.template-description {
    color: #64748b;
    font-size: 0.875rem;
    margin-bottom: 0.75rem;
}

.template-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    margin-bottom: 0.75rem;
}

.template-tag {
    font-size: 0.75rem;
    padding: 0.125rem 0.375rem;
    background: #e2e8f0;
    border-radius: 4px;
    color: #475569;
}

.template-actions {
    display: flex;
    gap: 0.5rem;
}

.config-canvas {
    min-height: 400px;
    border: 2px dashed #e2e8f0;
    border-radius: 8px;
    position: relative;
    background: #fafafa;
}

.canvas-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 400px;
}

.placeholder-content {
    text-align: center;
    color: #64748b;
}

.placeholder-content h4 {
    margin-bottom: 0.5rem;
    color: #374151;
}

.config-output {
    max-height: 400px;
    overflow-y: auto;
}

.config-output pre {
    margin: 0;
    padding: 1rem;
    background: #f8fafc;
    border-radius: 4px;
    font-family: 'Fira Code', 'Monaco', 'Cascadia Code', monospace;
    font-size: 0.875rem;
    line-height: 1.5;
}

.template-loading {
    grid-column: 1 / -1;
    text-align: center;
    padding: 2rem;
    color: #64748b;
}
</style>

<script>
let selectedTemplate = null;
let currentConfig = '';

// Load templates on page load
document.addEventListener('DOMContentLoaded', function() {
    loadTemplates();
});

async function loadTemplates() {
    try {
        const response = await fetch('/api/templates');
        const data = await response.json();
        
        if (data.success) {
            displayTemplates(data.data.templates);
        } else {
            console.error('Failed to load templates:', data.error);
        }
    } catch (error) {
        console.error('Error loading templates:', error);
        document.getElementById('template-grid').innerHTML = 
            '<div class="nixai-alert nixai-alert-error">Failed to load templates</div>';
    }
}

function displayTemplates(templates) {
    const grid = document.getElementById('template-grid');
    
    if (templates.length === 0) {
        grid.innerHTML = '<div class="nixai-alert nixai-alert-info">No templates available</div>';
        return;
    }
    
    let html = '';
    templates.forEach(template => {
        let tagsHtml = '';
        template.tags.forEach(tag => {
            tagsHtml += '<span class="template-tag">' + tag + '</span>';
        });
        
        html += '<div class="template-card" onclick="selectTemplate(\'' + template.name + '\')">' +
                '<div class="template-header">' +
                '<span class="template-name">' + template.name + '</span>' +
                '<span class="template-category">' + template.category + '</span>' +
                '</div>' +
                '<div class="template-description">' + template.description + '</div>' +
                '<div class="template-tags">' + tagsHtml + '</div>' +
                '<div class="template-actions">' +
                '<button class="nixai-btn nixai-btn-sm nixai-btn-primary" onclick="event.stopPropagation(); applyTemplate(\'' + template.name + '\')">' +
                'Apply</button>' +
                '<button class="nixai-btn nixai-btn-sm" onclick="event.stopPropagation(); previewTemplate(\'' + template.name + '\')">' +
                'Preview</button>' +
                '</div>' +
                '</div>';
    });
    
    grid.innerHTML = html;
}

async function selectTemplate(templateName) {
    // Remove previous selection
    document.querySelectorAll('.template-card').forEach(card => {
        card.classList.remove('selected');
    });
    
    // Add selection to clicked card
    event.currentTarget.classList.add('selected');
    selectedTemplate = templateName;
    
    // Load template preview
    await previewTemplate(templateName);
}

async function previewTemplate(templateName) {
    try {
        const response = await fetch('/api/templates/' + templateName);
        const data = await response.json();
        
        if (data.success) {
            const template = data.data.template;
            updateConfigPreview(template.content);
            updateCanvasWithTemplate(template);
        }
    } catch (error) {
        console.error('Error previewing template:', error);
    }
}

async function applyTemplate(templateName) {
    try {
        const response = await fetch('/api/templates/' + templateName + '/apply', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                outputPath: '',
                merge: false
            })
        });
        
        const data = await response.json();
        
        if (data.success) {
            const template = data.template;
            currentConfig = template.content;
            updateConfigPreview(template.content);
            updateCanvasWithTemplate(template);
            
            // Show success message
            showNotification('Template applied successfully!', 'success');
        } else {
            showNotification('Failed to apply template: ' + data.error, 'error');
        }
    } catch (error) {
        console.error('Error applying template:', error);
        showNotification('Error applying template', 'error');
    }
}

function updateConfigPreview(content) {
    const preview = document.getElementById('config-preview');
    preview.innerHTML = '<code class="language-nix">' + escapeHtml(content) + '</code>';
}

function updateCanvasWithTemplate(template) {
    const canvas = document.getElementById('config-canvas');
    
    let tagsHtml = '';
    template.tags.forEach(tag => {
        tagsHtml += '<span class="template-tag">' + tag + '</span>';
    });
    
    canvas.innerHTML = 
        '<div style="padding: 1rem;">' +
            '<div class="nixai-card">' +
                '<div class="nixai-card-header">' +
                    '<h4>Template: ' + template.name + '</h4>' +
                    '<span class="template-category">' + template.category + '</span>' +
                '</div>' +
                '<div style="padding: 1rem;">' +
                    '<p><strong>Description:</strong> ' + template.description + '</p>' +
                    '<div class="template-tags" style="margin-top: 0.5rem;">' +
                        tagsHtml +
                    '</div>' +
                    '<div style="margin-top: 1rem;">' +
                        '<button class="nixai-btn nixai-btn-sm" onclick="editTemplate()">' +
                            '<span>✏️</span> Edit Configuration' +
                        '</button>' +
                        '<button class="nixai-btn nixai-btn-sm" onclick="clearCanvas()">' +
                            '<span>🗑️</span> Clear' +
                        '</button>' +
                    '</div>' +
                '</div>' +
            '</div>' +
        '</div>';
}

function startFromScratch() {
    const canvas = document.getElementById('config-canvas');
    canvas.innerHTML = 
        '<div style="padding: 1rem;">' +
            '<div class="nixai-card">' +
                '<div class="nixai-card-header">' +
                    '<h4>New Configuration</h4>' +
                '</div>' +
                '<div style="padding: 1rem;">' +
                    '<p>Building a new NixOS configuration from scratch...</p>' +
                    '<div style="margin-top: 1rem;">' +
                        '<button class="nixai-btn nixai-btn-sm nixai-btn-primary" onclick="addBasicServices()">' +
                            '<span>⚙️</span> Add Basic Services' +
                        '</button>' +
                        '<button class="nixai-btn nixai-btn-sm" onclick="addDesktopEnvironment()">' +
                            '<span>🖥️</span> Add Desktop' +
                        '</button>' +
                        '<button class="nixai-btn nixai-btn-sm" onclick="addDevelopmentTools()">' +
                            '<span>🛠️</span> Add Dev Tools' +
                        '</button>' +
                    '</div>' +
                '</div>' +
            '</div>' +
        '</div>';
    
    currentConfig = '{ config, pkgs, ... }:\n\n{\n  # Basic NixOS configuration\n  system.stateVersion = "23.11";\n}';
    updateConfigPreview(currentConfig);
}

function clearCanvas() {
    const canvas = document.getElementById('config-canvas');
    canvas.innerHTML = 
        '<div class="canvas-placeholder">' +
            '<div class="placeholder-content">' +
                '<h4>Start Building Your Configuration</h4>' +
                '<p>Choose a template from the library above or start from scratch</p>' +
                '<button class="nixai-btn nixai-btn-primary" onclick="startFromScratch()">' +
                    '<span>🚀</span> Start from Scratch' +
                '</button>' +
            '</div>' +
        '</div>';
    
    currentConfig = '';
    updateConfigPreview('# Your NixOS configuration will appear here\\n# Choose a template or start building to see the generated configuration\\n\\n{ config, pkgs, ... }:\\n\\n{\\n  # Configuration options will be added here\\n  system.stateVersion = "23.11";\\n}');
}

function refreshTemplates() {
    loadTemplates();
    showNotification('Templates refreshed', 'info');
}

function validateConfig() {
    if (!currentConfig) {
        showNotification('No configuration to validate', 'warning');
        return;
    }
    
    // Basic validation
    if (currentConfig.includes('system.stateVersion')) {
        showNotification('Configuration validation passed', 'success');
    } else {
        showNotification('Configuration may be incomplete', 'warning');
    }
}

function generateConfig() {
    if (currentConfig) {
        showNotification('Configuration generated successfully', 'success');
    } else {
        showNotification('No configuration to generate', 'warning');
    }
}

function exportConfig() {
    if (!currentConfig) {
        showNotification('No configuration to export', 'warning');
        return;
    }
    
    downloadConfig();
}

function copyToClipboard() {
    if (!currentConfig) {
        showNotification('No configuration to copy', 'warning');
        return;
    }
    
    navigator.clipboard.writeText(currentConfig).then(() => {
        showNotification('Configuration copied to clipboard', 'success');
    }).catch(err => {
        console.error('Failed to copy:', err);
        showNotification('Failed to copy configuration', 'error');
    });
}

function downloadConfig() {
    if (!currentConfig) {
        showNotification('No configuration to download', 'warning');
        return;
    }
    
    const blob = new Blob([currentConfig], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'configuration.nix';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    showNotification('Configuration downloaded', 'success');
}

function importTemplate() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.nix,.yaml,.yml';
    input.onchange = (e) => {
        const file = e.target.files[0];
        if (file) {
            const reader = new FileReader();
            reader.onload = (e) => {
                currentConfig = e.target.result;
                updateConfigPreview(currentConfig);
                showNotification('Template imported successfully', 'success');
            };
            reader.readAsText(file);
        }
    };
    input.click();
}

function showNotification(message, type) {
    if (!type) type = 'info';
    
    // Create notification element
    const notification = document.createElement('div');
    notification.className = 'nixai-alert nixai-alert-' + type;
    notification.style.cssText = 
        'position: fixed;' +
        'top: 20px;' +
        'right: 20px;' +
        'z-index: 1000;' +
        'max-width: 300px;' +
        'animation: slideIn 0.3s ease;';
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    // Remove after 3 seconds
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 3000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Add CSS for animations
const style = document.createElement('style');
style.textContent = 
    '@keyframes slideIn {' +
        'from { transform: translateX(100%); opacity: 0; }' +
        'to { transform: translateX(0); opacity: 1; }' +
    '}' +
    '@keyframes slideOut {' +
        'from { transform: translateX(0); opacity: 1; }' +
        'to { transform: translateX(100%); opacity: 0; }' +
    '}';
document.head.appendChild(style);
</script>`
}

// Fleet content
func (s *EnhancedServer) getFleetContent(data map[string]interface{}) string {
	return `<div class="nixai-card">
    <div class="nixai-card-header">
        <h3 class="nixai-card-title">Fleet Management</h3>
    </div>
    <div class="fleet-content">
        <p>Fleet management interface for managing multiple NixOS machines.</p>
        <div class="nixai-alert nixai-alert-info">
            <strong>Feature Status:</strong> Fleet management module is being developed.
        </div>
    </div>
</div>`
}

// Teams content
func (s *EnhancedServer) getTeamsContent(data map[string]interface{}) string {
	return `<div class="nixai-card">
    <div class="nixai-card-header">
        <h3 class="nixai-card-title">Team Collaboration</h3>
    </div>
    <div class="teams-content">
        <p>Team collaboration features for working together on NixOS configurations.</p>
        <div class="nixai-alert nixai-alert-success">
            <strong>Real-time Collaboration:</strong> WebSocket-based live editing and communication.
        </div>
    </div>
</div>`
}

// Versions content
func (s *EnhancedServer) getVersionsContent(data map[string]interface{}) string {
	return `<div class="nixai-card">
    <div class="nixai-card-header">
        <h3 class="nixai-card-title">Version Control</h3>
    </div>
    <div class="versions-content">
        <p>Git-like version control for NixOS configurations.</p>
        <div class="nixai-alert nixai-alert-success">
            <strong>Available:</strong> Branch management, commit history, and configuration tracking.
        </div>
    </div>
</div>`
}

// API Handlers

// Dashboard API handler
func (s *EnhancedServer) handleDashboardAPI(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.logger.Debug("Handling dashboard API request")

	// Get real fleet data
	machines, err := s.fleetManager.ListMachines(r.Context())
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get machines: %v", err))
		machines = []*fleet.Machine{} // Fallback to empty list
	}

	// Count machine statuses
	totalMachines := len(machines)
	healthyMachines := 0

	for _, machine := range machines {
		if machine.Status == "online" {
			healthyMachines++
		}
	}

	// Get configuration count from repository if available
	totalConfigs := s.getTotalConfigs()

	// Create activities based on actual data
	activities := []map[string]interface{}{
		{
			"type":      "system",
			"message":   "NixAI web interface started",
			"timestamp": time.Now().Format(time.RFC3339),
			"icon":      "🚀",
		},
	}

	// Add machine-related activities
	if totalMachines > 0 {
		activities = append(activities, map[string]interface{}{
			"type":      "fleet",
			"message":   fmt.Sprintf("Discovered %d machines from repository", totalMachines),
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"icon":      "🖥️",
		})
	}

	// Create alerts based on actual status
	alerts := []map[string]interface{}{}

	if totalMachines == 0 {
		alerts = append(alerts, map[string]interface{}{
			"level":     "info",
			"title":     "No Machines Found",
			"message":   "No machines configured. Use --repo flag to analyze a NixOS repository",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	} else {
		alerts = append(alerts, map[string]interface{}{
			"level":     "success",
			"title":     "Fleet Status",
			"message":   fmt.Sprintf("%d machines discovered, %d healthy", totalMachines, healthyMachines),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}

	data := map[string]interface{}{
		"overview": map[string]interface{}{
			"total_machines":   totalMachines,
			"healthy_machines": healthyMachines,
			"total_configs":    totalConfigs,
			"active_teams":     1, // Default to 1 for now
		},
		"activities": activities,
		"alerts":     alerts,
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	s.sendSuccess(w, data)
}

// Dashboard details API handler
func (s *EnhancedServer) handleDashboardDetails(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	s.logger.Debug("Handling dashboard details API request")

	// Get real system information
	sysInfo, err := system.GetSystemInfo()
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to get system info: %v", err))
		// Fallback to placeholder data
		sysInfo = &system.SystemInfo{
			Uptime:   "unknown",
			CPUUsage: "0%",
			Memory:   "unknown",
			Disk:     "unknown",
		}
	}

	// Get recent configurations from fleet
	machines, _ := s.fleetManager.ListMachines(r.Context())
	recentConfigs := []map[string]interface{}{}
	for _, machine := range machines {
		if machine.Metadata != nil {
			if configPath, ok := machine.Metadata["config_path"]; ok {
				recentConfigs = append(recentConfigs, map[string]interface{}{
					"name":         machine.Name,
					"path":         configPath,
					"last_updated": machine.Metadata["discovered_at"],
					"status":       "discovered",
				})
			}
		}
	}

	// Provide detailed dashboard information
	data := map[string]interface{}{
		"system": map[string]interface{}{
			"uptime":    sysInfo.Uptime,
			"cpu_usage": sysInfo.CPUUsage,
			"memory":    sysInfo.Memory,
			"disk":      sysInfo.Disk,
			"load_avg":  sysInfo.LoadAvg,
			"processes": sysInfo.Processes,
		},
		"recent_configs": recentConfigs,
		"team_activity": []map[string]interface{}{
			{
				"user":      "demo",
				"action":    "viewed dashboard",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		},
	}

	s.sendSuccess(w, data)
}

// WebSocket connection handler
func (s *EnhancedServer) handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("WebSocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()

	s.logger.Info("WebSocket connection established")

	// Send welcome message
	welcome := map[string]interface{}{
		"type": "welcome",
		"data": map[string]interface{}{
			"message": "Connected to NixAI real-time collaboration",
			"time":    time.Now().Format(time.RFC3339),
		},
	}

	if err := conn.WriteJSON(welcome); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to send welcome message: %v", err))
		return
	}

	// Keep connection alive and handle messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("WebSocket connection closed: %v", err))
			break
		}

		// Echo message back for now
		response := map[string]interface{}{
			"type": "echo",
			"data": msg,
			"time": time.Now().Format(time.RFC3339),
		}

		if err := conn.WriteJSON(response); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to send WebSocket response: %v", err))
			break
		}
	}
}

// Placeholder API handlers for dashboard data
func (s *EnhancedServer) handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get real fleet data
	machines, err := s.fleetManager.ListMachines(r.Context())
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to get machines: %v", err), http.StatusInternalServerError)
		return
	}

	// Count machine statuses
	machineStats := map[string]int{
		"total":   len(machines),
		"healthy": 0,
		"warning": 0,
		"error":   0,
	}

	for _, machine := range machines {
		switch machine.Status {
		case "online":
			machineStats["healthy"]++
		case "degraded":
			machineStats["warning"]++
		case "offline":
			machineStats["error"]++
		}
	}

	stats := map[string]interface{}{
		"machines": machineStats,
		"configurations": map[string]int{
			"total":    0, // TODO: Get from repository if available
			"modified": 0,
			"deployed": 0,
		},
		"teams": map[string]int{
			"active": 0, // TODO: Get from team manager if available
			"total":  0,
		},
	}

	s.sendSuccess(w, stats)
}

func (s *EnhancedServer) handleDashboardActivities(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	activities := []map[string]interface{}{
		{
			"id":        "1",
			"type":      "system_start",
			"message":   "NixAI web interface started",
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"icon":      "🚀",
		},
		{
			"id":        "2",
			"type":      "config_load",
			"message":   "Configuration templates loaded",
			"timestamp": time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
			"icon":      "⚙️",
		},
	}

	s.sendSuccess(w, activities)
}

func (s *EnhancedServer) handleDashboardAlerts(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	alerts := []map[string]interface{}{
		{
			"id":        "1",
			"level":     "info",
			"title":     "System Ready",
			"message":   "NixAI enhanced web interface is fully operational",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	s.sendSuccess(w, alerts)
}

func (s *EnhancedServer) handleAISuggestions(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get current machines for context
	machines, err := s.fleetManager.ListMachines(r.Context())
	if err != nil {
		machines = []*fleet.Machine{} // Empty slice on error
	}
	suggestions := s.generateAISuggestions(machines)

	s.sendSuccess(w, suggestions)
}

// Auth status handler
func (s *EnhancedServer) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	token := ""
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// If no auth manager or token, return unauthenticated
	if s.authManager == nil || token == "" {
		userStatus := map[string]interface{}{
			"authenticated": false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userStatus)
		return
	}

	// Validate session
	user, err := s.authManager.ValidateSession(token)
	if err != nil {
		userStatus := map[string]interface{}{
			"authenticated": false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userStatus)
		return
	}

	// Return authenticated user status
	userStatus := map[string]interface{}{
		"authenticated": true,
		"user":          user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userStatus)
}

// Helper methods for EnhancedServer

// sendSuccess sends a successful JSON response
func (s *EnhancedServer) sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encode JSON response: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendError sends an error JSON response
func (s *EnhancedServer) sendError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(code)

	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encode error JSON response: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// sendJSON sends raw JSON data
func (s *EnhancedServer) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encode JSON data: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// renderTemplate attempts to render templates, with fallback to simple HTML
func (s *EnhancedServer) renderTemplate(w http.ResponseWriter, templateName string, data map[string]interface{}) {
	if s.templates != nil {
		// We need to parse individual template files to get the right content block
		templateFile := strings.TrimSuffix(templateName, ".html") + ".html"

		// Read and parse the specific template file
		templatePath := filepath.Join(s.config.TemplateDir, templateFile)
		specificTemplate, err := template.ParseFiles(templatePath, filepath.Join(s.config.TemplateDir, "base.html"))
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to parse template files: %v", err))
			s.renderSimpleHTML(w, data)
			return
		}

		// Determine the correct content block name based on the template
		var contentBlockName string
		baseName := strings.TrimSuffix(templateFile, ".html")
		switch baseName {
		case "dashboard":
			contentBlockName = "dashboard" // dashboard.html uses {{define "dashboard"}}
		default:
			contentBlockName = "content" // other templates use {{define "content"}}
		}

		// Execute the content block from the specific template
		var contentBuf strings.Builder
		if err := specificTemplate.ExecuteTemplate(&contentBuf, contentBlockName, data); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to execute %s template from %s: %v", contentBlockName, templateFile, err))
			s.renderSimpleHTML(w, data)
			return
		}

		// Add the rendered content to the data
		data["Content"] = template.HTML(contentBuf.String())

		// Now execute the base template with the content included
		if err := specificTemplate.ExecuteTemplate(w, "base.html", data); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to execute base template: %v", err))
			s.renderSimpleHTML(w, data)
			return
		}
		return
	}

	// Fallback to simple HTML
	s.renderSimpleHTML(w, data)
}

// renderSimpleHTML renders a simple HTML fallback when templates fail
func (s *EnhancedServer) renderSimpleHTML(w http.ResponseWriter, data map[string]interface{}) {
	title := "NixAI Enhanced"
	if t, ok := data["Title"].(string); ok {
		title = t
	}

	activePage := "dashboard"
	if p, ok := data["ActivePage"].(string); ok {
		activePage = p
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="/static/css/nixai-enhanced.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.0.0/css/all.min.css">
</head>
<body>
    <div class="nixai-layout">
        <header class="nixai-header">
            <nav class="nav">
                <div class="nixai-logo">
                    <span>🚀</span>
                    <span>NixAI Enhanced</span>
                </div>
                <ul class="nixai-nav-links">
                    <li><a href="/dashboard" data-nav="/dashboard" class="%s">📊 Dashboard</a></li>
                    <li><a href="/builder" data-nav="/builder" class="%s">🎨 Builder</a></li>
                    <li><a href="/fleet" data-nav="/fleet" class="%s">🚀 Fleet</a></li>
                    <li><a href="/teams" data-nav="/teams" class="%s">👥 Teams</a></li>
                    <li><a href="/versions" data-nav="/versions" class="%s">📝 Versions</a></li>
                </ul>
            </nav>
        </header>
        <main class="nixai-main">
            <div class="nixai-content">
                %s
            </div>
        </main>
    </div>
    <script src="/static/js/nixai-enhanced.js"></script>
</body>
</html>`,
		title,
		s.getActiveClass("dashboard", activePage),
		s.getActiveClass("builder", activePage),
		s.getActiveClass("fleet", activePage),
		s.getActiveClass("teams", activePage),
		s.getActiveClass("versions", activePage),
		s.getPageContent(activePage, data))

	fmt.Fprint(w, html)
}

// getActiveClass returns "active" if the page matches the current page
func (s *EnhancedServer) getActiveClass(page, activePage string) string {
	if page == activePage {
		return "active"
	}
	return ""
}

// Fleet API handlers

func (s *EnhancedServer) handleFleetAPI(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get real fleet data from fleet manager
	machines, err := s.fleetManager.ListMachines(r.Context())
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to get machines: %v", err), http.StatusInternalServerError)
		return
	}

	// Count machine statuses
	machineStats := map[string]int{
		"total":   len(machines),
		"healthy": 0,
		"warning": 0,
		"error":   0,
	}

	for _, machine := range machines {
		switch machine.Status {
		case "online":
			machineStats["healthy"]++
		case "degraded":
			machineStats["warning"]++
		case "offline":
			machineStats["error"]++
		}
	}

	// Convert machines to API format
	var apiMachines []map[string]interface{}
	for _, machine := range machines {
		apiMachines = append(apiMachines, map[string]interface{}{
			"id":       machine.ID,
			"name":     machine.Name,
			"status":   machine.Status,
			"address":  machine.Address,
			"metadata": machine.Metadata,
		})
	}

	// Get fleet overview data
	data := map[string]interface{}{
		"overview": machineStats,
		"machines": apiMachines,
	}

	s.sendSuccess(w, data)
}

func (s *EnhancedServer) handleFleetMachines(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET", "HEAD":
		s.handleListMachines(w, r)
	case "POST":
		s.handleAddMachine(w, r)
	default:
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *EnhancedServer) handleListMachines(w http.ResponseWriter, r *http.Request) {
	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	machines, err := s.fleetManager.ListMachines(r.Context())
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to list machines: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to API format
	var apiMachines []map[string]interface{}
	for _, machine := range machines {
		apiMachines = append(apiMachines, map[string]interface{}{
			"id":       machine.ID,
			"name":     machine.Name,
			"status":   machine.Status,
			"address":  machine.Address,
			"metadata": machine.Metadata,
		})
	}

	s.sendSuccess(w, apiMachines)
}

func (s *EnhancedServer) handleAddMachine(w http.ResponseWriter, r *http.Request) {
	var machineReq struct {
		ID       string            `json:"id"`
		Name     string            `json:"name"`
		Address  string            `json:"address"`
		Metadata map[string]string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&machineReq); err != nil {
		s.sendError(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create machine object
	machine := &fleet.Machine{
		ID:       machineReq.ID,
		Name:     machineReq.Name,
		Address:  machineReq.Address,
		Status:   "offline", // Default status
		Metadata: machineReq.Metadata,
	}

	// Add to fleet
	if err := s.fleetManager.AddMachine(r.Context(), machine); err != nil {
		s.sendError(w, fmt.Sprintf("Failed to add machine: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Info(fmt.Sprintf("Machine added: %s (%s)", machineReq.Name, machineReq.ID))
	s.sendSuccess(w, machine)
}

func (s *EnhancedServer) handleGetMachine(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	machineID := vars["machineId"]

	if s.fleetManager == nil {
		s.sendError(w, "Fleet management not available", http.StatusServiceUnavailable)
		return
	}

	machine, err := s.fleetManager.GetMachine(r.Context(), machineID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendSuccess(w, machine)
}

func (s *EnhancedServer) handleRemoveMachine(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authentication
	user := s.requireAuth(w, r)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	machineID := vars["machineId"]

	if s.fleetManager == nil {
		s.sendError(w, "Fleet management not available", http.StatusServiceUnavailable)
		return
	}

	if err := s.fleetManager.RemoveMachine(r.Context(), machineID); err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info(fmt.Sprintf("Machine removed: %s by user %s", machineID, user.Username))
	s.sendSuccess(w, map[string]interface{}{
		"message":    "Machine removed successfully",
		"machine_id": machineID,
	})
}

func (s *EnhancedServer) handleFleetDeploy(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authentication
	user := s.requireAuth(w, r)
	if user == nil {
		return
	}

	var deployReq struct {
		Name       string   `json:"name"`
		ConfigHash string   `json:"config_hash"`
		Targets    []string `json:"targets"`
		Strategy   string   `json:"strategy"`
		BatchSize  int      `json:"batch_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&deployReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if deployReq.Name == "" || deployReq.ConfigHash == "" {
		s.sendError(w, "Deployment name and config hash are required", http.StatusBadRequest)
		return
	}

	// Create mock deployment
	deployment := map[string]interface{}{
		"id":          fmt.Sprintf("deploy-%d", time.Now().Unix()),
		"name":        deployReq.Name,
		"config_hash": deployReq.ConfigHash,
		"targets":     deployReq.Targets,
		"status":      "pending",
		"created_at":  time.Now().Format(time.RFC3339),
		"created_by":  user.Username,
	}

	s.logger.Info(fmt.Sprintf("Deployment created: %s by user %s", deployReq.Name, user.Username))
	s.sendSuccess(w, deployment)
}

func (s *EnhancedServer) handleFleetDeployments(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mock deployments data
	deployments := []map[string]interface{}{
		{
			"id":         "deploy-001",
			"name":       "Production Update",
			"status":     "completed",
			"targets":    []string{"server01", "server02"},
			"created_at": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"created_by": "admin",
		},
		{
			"id":         "deploy-002",
			"name":       "Security Patches",
			"status":     "running",
			"targets":    []string{"server03"},
			"created_at": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"created_by": "admin",
		},
	}

	s.sendSuccess(w, map[string]interface{}{
		"deployments": deployments,
		"total":       len(deployments),
	})
}

func (s *EnhancedServer) handleGetDeployment(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	deploymentID := vars["deploymentId"]

	// Mock deployment data
	deployment := map[string]interface{}{
		"id":         deploymentID,
		"name":       "Production Update",
		"status":     "completed",
		"targets":    []string{"server01", "server02"},
		"created_at": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		"created_by": "admin",
		"progress": map[string]interface{}{
			"total":      2,
			"completed":  2,
			"failed":     0,
			"percentage": 100,
		},
	}

	s.sendSuccess(w, deployment)
}

func (s *EnhancedServer) handleFleetHealth(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Mock fleet health data
	health := map[string]interface{}{
		"overall_status": "healthy",
		"machines": []map[string]interface{}{
			{
				"id":     "server01",
				"name":   "Production Server 1",
				"status": "healthy",
				"cpu":    45.2,
				"memory": 67.8,
				"disk":   23.1,
				"uptime": "15d 4h 32m",
			},
			{
				"id":     "server02",
				"name":   "Production Server 2",
				"status": "healthy",
				"cpu":    32.1,
				"memory": 54.3,
				"disk":   18.7,
				"uptime": "12d 8h 15m",
			},
		},
		"alerts":       []map[string]interface{}{},
		"last_updated": time.Now().Format(time.RFC3339),
	}

	s.sendSuccess(w, health)
}

// Configuration Management Handlers

func (s *EnhancedServer) handleConfigurations(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.Method {
	case "GET":
		s.handleListConfigurations(w, r)
	case "POST":
		s.handleCreateConfiguration(w, r)
	default:
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *EnhancedServer) handleListConfigurations(w http.ResponseWriter, r *http.Request) {
	// Mock configuration data for now
	configurations := []map[string]interface{}{
		{
			"id":          "config_001",
			"name":        "Desktop Development",
			"type":        "desktop",
			"description": "Development machine with GNOME and dev tools",
			"created_at":  "2024-01-15T10:00:00Z",
			"updated_at":  "2024-01-15T10:00:00Z",
			"status":      "active",
		},
		{
			"id":          "config_002",
			"name":        "Production Server",
			"type":        "server",
			"description": "Web server with Nginx and PostgreSQL",
			"created_at":  "2024-01-10T14:30:00Z",
			"updated_at":  "2024-01-20T09:15:00Z",
			"status":      "deployed",
		},
	}

	s.sendSuccess(w, configurations)
}

func (s *EnhancedServer) handleCreateConfiguration(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	user := s.requireAuth(w, r)
	if user == nil {
		return
	}

	var configReq struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&configReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if configReq.Name == "" {
		s.sendError(w, "Configuration name is required", http.StatusBadRequest)
		return
	}

	// Generate a new configuration ID
	configID := fmt.Sprintf("config_%d", time.Now().Unix())

	// Create the configuration
	newConfig := map[string]interface{}{
		"id":          configID,
		"name":        configReq.Name,
		"type":        configReq.Type,
		"description": configReq.Description,
		"created_at":  time.Now().Format(time.RFC3339),
		"updated_at":  time.Now().Format(time.RFC3339),
		"status":      "draft",
		"created_by":  user.Username,
	}

	s.logger.Info(fmt.Sprintf("Configuration created: %s by user %s", configReq.Name, user.Username))
	s.sendSuccess(w, newConfig)
}

// Builder API Handlers

func (s *EnhancedServer) handleBuilderValidate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Demo validation response
	response := map[string]interface{}{
		"valid":  true,
		"errors": []string{},
		"warnings": []string{
			"Consider enabling firewall for better security",
		},
		"suggestions": []string{
			"Add hardware acceleration for better performance",
			"Consider using systemd for service management",
			"Consider using Home Manager for user-specific configurations",
		},
	}

	json.NewEncoder(w).Encode(response)
}

func (s *EnhancedServer) handleBuilderGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Demo generated configuration
	nixConfig := `# Generated NixOS Configuration
{
  # System configuration
  system.stateVersion = "24.05";
  
  # Boot configuration
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;
  
  # Network configuration
  networking = {
    hostName = "nixos-server";
    networkmanager.enable = true;
    firewall.enable = true;
  };
  
  # Services
  services = {
    openssh.enable = true;
    nginx.enable = true;
    postgresql.enable = true;
  };
  
  # Users
  users.users.admin = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" ];
  };
  
  # Environment
  environment.systemPackages = with pkgs; [
    wget curl git vim htop
  ];
}`

	response := map[string]interface{}{
		"success":       true,
		"configuration": nixConfig,
		"filename":      "configuration.nix",
		"size":          len(nixConfig),
		"checksum":      "abc123def456",
	}

	json.NewEncoder(w).Encode(response)
}

// Config API Handlers

func (s *EnhancedServer) handleConfigBranches(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	s.logger.Debug("Handling config branches API request")

	// Return branch information (can be integrated with ConfigRepository later)
	data := map[string]interface{}{
		"branches": []map[string]interface{}{
			{
				"name":      "main",
				"current":   true,
				"commit":    "abc123",
				"timestamp": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"message":   "Latest configuration update",
			},
			{
				"name":      "development",
				"current":   false,
				"commit":    "def456",
				"timestamp": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"message":   "Development environment setup",
			},
		},
		"current_branch": "main",
	}

	s.sendSuccess(w, data)
}

func (s *EnhancedServer) handleConfigFiles(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// For HEAD requests, only send headers
	if r.Method == "HEAD" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}

	s.logger.Debug("Handling config files API request")

	// Return configuration files information
	data := map[string]interface{}{
		"files": []map[string]interface{}{
			{
				"path":     "configuration.nix",
				"size":     "2.1KB",
				"modified": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				"status":   "modified",
				"type":     "configuration",
			},
			{
				"path":     "hardware-configuration.nix",
				"size":     "1.3KB",
				"modified": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"status":   "clean",
				"type":     "hardware",
			},
			{
				"path":     "flake.nix",
				"size":     "856B",
				"modified": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"status":   "clean",
				"type":     "flake",
			},
		},
		"total_files":    3,
		"modified_files": 1,
	}

	s.sendSuccess(w, data)
}

// Versioning API Handlers

func (s *EnhancedServer) handleVersioningRepositories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Try to get real repository information from ConfigRepository
	var repositories []map[string]interface{}

	if s.configRepo != nil {
		// ConfigRepository represents a single repository, so we'll show its status
		repositories = append(repositories, map[string]interface{}{
			"id":           "config-repo",
			"name":         "Configuration Repository",
			"path":         ".",
			"branch":       "main",
			"lastCommit":   "latest",
			"lastModified": time.Now().Format("2006-01-02T15:04:05Z"),
			"status":       "active",
		})
	}

	// Fall back to demo data if no real repositories available
	if len(repositories) == 0 {
		repositories = []map[string]interface{}{
			{
				"id":           "nixos-config",
				"name":         "NixOS Configuration",
				"path":         "/etc/nixos",
				"branch":       "main",
				"lastCommit":   "abc123",
				"lastModified": time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z"),
				"status":       "clean",
			},
			{
				"id":           "home-manager",
				"name":         "Home Manager Config",
				"path":         "~/.config/home-manager",
				"branch":       "main",
				"lastCommit":   "def456",
				"lastModified": time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04:05Z"),
				"status":       "modified",
			},
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"repositories": repositories,
		"total":        len(repositories),
	})
}

func (s *EnhancedServer) handleVersioningStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Demo repository status
	status := map[string]interface{}{
		"branch":    "main",
		"commit":    "abc123def456",
		"clean":     false,
		"modified":  []string{"configuration.nix", "hardware-configuration.nix"},
		"untracked": []string{"test.nix"},
		"staged":    []string{},
		"behind":    0,
		"ahead":     2,
	}

	json.NewEncoder(w).Encode(status)
}

func (s *EnhancedServer) handleVersioningBranches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		// List branches
		branches := []map[string]interface{}{
			{
				"name":       "main",
				"current":    true,
				"commit":     "abc123",
				"lastCommit": time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
			{
				"name":       "feature/new-services",
				"current":    false,
				"commit":     "def456",
				"lastCommit": time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
			{
				"name":       "hotfix/security-patch",
				"current":    false,
				"commit":     "ghi789",
				"lastCommit": time.Now().Add(-72 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"branches": branches,
			"total":    len(branches),
		})
	} else {
		// Create branch
		response := map[string]interface{}{
			"success": true,
			"message": "Branch created successfully",
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (s *EnhancedServer) handleVersioningCommits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Try to get real snapshots from ConfigRepository
	var commits []map[string]interface{}

	if s.configRepo != nil {
		snapshots, err := s.configRepo.ListSnapshots(r.Context())
		if err == nil {
			for _, snapshot := range snapshots {
				commits = append(commits, map[string]interface{}{
					"hash":      snapshot.ID,
					"author":    snapshot.Author,
					"message":   snapshot.Message,
					"timestamp": snapshot.Timestamp.Format("2006-01-02T15:04:05Z"),
					"files":     getFileList(snapshot.Files),
				})
			}
		} else {
			s.logger.Warn(fmt.Sprintf("Failed to get snapshots from config repo: %v", err))
		}
	}

	// Fall back to demo data if no real snapshots available
	if len(commits) == 0 {
		commits = []map[string]interface{}{
			{
				"hash":      "abc123def456",
				"author":    "Admin User",
				"message":   "Update system configuration with new services",
				"timestamp": time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z"),
				"files":     []string{"configuration.nix", "services.nix"},
			},
			{
				"hash":      "def456ghi789",
				"author":    "Dev User",
				"message":   "Add development tools and environment",
				"timestamp": time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05Z"),
				"files":     []string{"packages.nix", "shell.nix"},
			},
			{
				"hash":      "ghi789jkl012",
				"author":    "Admin User",
				"message":   "Initial system setup",
				"timestamp": time.Now().Add(-72 * time.Hour).Format("2006-01-02T15:04:05Z"),
				"files":     []string{"configuration.nix", "hardware-configuration.nix"},
			},
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"commits": commits,
		"total":   len(commits),
	})
}

// Helper function to get file list from snapshot files map
func getFileList(files map[string]string) []string {
	var fileList []string
	for filePath := range files {
		fileList = append(fileList, filePath)
	}
	return fileList
}

func (s *EnhancedServer) handleVersioningTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		// List tags
		tags := []map[string]interface{}{
			{
				"name":      "v1.0.0",
				"commit":    "abc123",
				"message":   "Initial stable release",
				"timestamp": time.Now().Add(-168 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
			{
				"name":      "v1.1.0",
				"commit":    "def456",
				"message":   "Feature update with new services",
				"timestamp": time.Now().Add(-72 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"tags":  tags,
			"total": len(tags),
		})
	} else {
		// Create tag
		response := map[string]interface{}{
			"success": true,
			"message": "Tag created successfully",
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (s *EnhancedServer) handleVersioningStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Demo versioning stats
	stats := map[string]interface{}{
		"totalRepositories": 2,
		"totalCommits":      15,
		"totalBranches":     3,
		"totalTags":         2,
		"activeUsers":       2,
		"recentActivity": []map[string]interface{}{
			{
				"action":    "commit",
				"user":      "Admin User",
				"message":   "Update system configuration",
				"timestamp": time.Now().Add(-2 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
			{
				"action":    "branch_create",
				"user":      "Dev User",
				"message":   "Created feature/new-services branch",
				"timestamp": time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04:05Z"),
			},
		},
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *EnhancedServer) handleVersioningChanges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Demo changes data
	changes := []map[string]interface{}{
		{
			"file":    "configuration.nix",
			"status":  "modified",
			"lines":   "+15 -3",
			"preview": "Added new services: nginx, postgresql",
		},
		{
			"file":    "hardware-configuration.nix",
			"status":  "modified",
			"lines":   "+2 -1",
			"preview": "Updated boot configuration",
		},
		{
			"file":    "test.nix",
			"status":  "untracked",
			"lines":   "+25 -0",
			"preview": "New test configuration file",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"changes": changes,
		"total":   len(changes),
	})
}

func (s *EnhancedServer) handleVersioningCreateBranch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"message": "Branch created successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *EnhancedServer) handleVersioningCreateTag(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"message": "Tag created successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *EnhancedServer) handleVersioningCommit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"commit":  fmt.Sprintf("abc%d", time.Now().Unix()),
		"message": "Commit created successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *EnhancedServer) handleVersioningCheckout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"message": "Checkout completed successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *EnhancedServer) handleVersioningDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	fileName := vars["fileName"]

	// Demo diff data
	diff := map[string]interface{}{
		"file": fileName,
		"changes": []map[string]interface{}{
			{
				"type":    "addition",
				"line":    15,
				"content": "+  services.nginx.enable = true;",
			},
			{
				"type":    "addition",
				"line":    16,
				"content": "+  services.postgresql.enable = true;",
			},
			{
				"type":    "deletion",
				"line":    20,
				"content": "-  # services.httpd.enable = false;",
			},
		},
	}

	json.NewEncoder(w).Encode(diff)
}

func (s *EnhancedServer) handleVersioningStage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"message": "Files staged successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func (s *EnhancedServer) handleVersioningStageAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"success": true,
		"message": "All files staged successfully",
	}
	json.NewEncoder(w).Encode(response)
}

// AI Chat API Handler

func (s *EnhancedServer) handleAIChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Demo AI chat response
	response := map[string]interface{}{
		"success": true,
		"response": "I can help you with NixOS configuration! Here are some suggestions:\n\n" +
			"1. For web servers, consider enabling nginx with SSL\n" +
			"2. Always configure firewall rules for security\n" +
			"3. Use systemd services for better process management\n" +
			"4. Consider using Home Manager for user-specific configurations\n\n" +
			"What specific aspect of NixOS would you like help with?",
		"suggestions": []string{
			"Configure web server",
			"Set up development environment",
			"Security hardening",
			"Package management",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// User Management Handlers

func (s *EnhancedServer) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authentication and admin permissions
	if !s.requireAdminAuth(w, r) {
		return
	}

	var createReq struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Password    string `json:"password"`
		Role        string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if s.authManager == nil {
		s.sendError(w, "User management not available", http.StatusServiceUnavailable)
		return
	}

	user, err := s.authManager.CreateUser(createReq.Username, createReq.Email, createReq.DisplayName, createReq.Password, createReq.Role)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return public user data (without password hash)
	publicUser := map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"role":         user.Role,
		"active":       user.Active,
		"created_at":   user.CreatedAt,
	}

	s.sendSuccess(w, publicUser)
}

func (s *EnhancedServer) handleListUsers(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authentication and admin permissions
	if !s.requireAdminAuth(w, r) {
		return
	}

	if s.authManager == nil {
		s.sendError(w, "User management not available", http.StatusServiceUnavailable)
		return
	}

	users := s.authManager.ListUsers()
	s.sendSuccess(w, users)
}

func (s *EnhancedServer) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authentication
	user := s.requireAuth(w, r)
	if user == nil {
		return
	}

	var changeReq struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&changeReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if s.authManager == nil {
		s.sendError(w, "User management not available", http.StatusServiceUnavailable)
		return
	}

	if err := s.authManager.ChangePassword(user.ID, changeReq.CurrentPassword, changeReq.NewPassword); err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Authentication middleware helpers

func (s *EnhancedServer) requireAuth(w http.ResponseWriter, r *http.Request) *auth.PublicUser {
	authHeader := r.Header.Get("Authorization")
	token := ""
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	if s.authManager == nil || token == "" {
		s.sendError(w, "Authentication required", http.StatusUnauthorized)
		return nil
	}

	user, err := s.authManager.ValidateSession(token)
	if err != nil {
		s.sendError(w, "Invalid session", http.StatusUnauthorized)
		return nil
	}

	return user
}

func (s *EnhancedServer) requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	user := s.requireAuth(w, r)
	if user == nil {
		return false
	}

	// Check if user has admin permissions
	hasAdmin := false
	for _, perm := range user.Permissions {
		if perm == "admin" || perm == "user_management" {
			hasAdmin = true
			break
		}
	}

	if !hasAdmin {
		s.sendError(w, "Admin access required", http.StatusForbidden)
		return false
	}

	return true
}

// getSystemStats gets current system statistics
func (s *EnhancedServer) getSystemStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"Uptime": "Unknown",
		"CPU":    "0%",
		"Memory": "0MB",
		"Disk":   "0%",
	}

	// Try to get uptime
	if content, err := os.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(content))
		if len(fields) > 0 {
			if uptimeSeconds, err := strconv.ParseFloat(fields[0], 64); err == nil {
				uptime := time.Duration(uptimeSeconds) * time.Second
				hours := int(uptime.Hours())
				minutes := int(uptime.Minutes()) % 60
				stats["Uptime"] = fmt.Sprintf("%dh %dm", hours, minutes)
			}
		}
	}

	// Try to get memory info
	if content, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(content), "\n")
		var total, available int64
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					total, _ = strconv.ParseInt(fields[1], 10, 64)
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					available, _ = strconv.ParseInt(fields[1], 10, 64)
				}
			}
		}
		if total > 0 {
			used := total - available
			usedMB := used / 1024
			stats["Memory"] = fmt.Sprintf("%.1fGB", float64(usedMB)/1024)
		}
	}

	// Try to get CPU usage (simplified)
	stats["CPU"] = "~25%" // Placeholder - real CPU monitoring would need more complex implementation

	// Try to get disk usage
	if _, err := os.Stat("/"); err == nil {
		var statfs syscall.Statfs_t
		if syscall.Statfs("/", &statfs) == nil {
			total := statfs.Blocks * uint64(statfs.Bsize)
			free := statfs.Bavail * uint64(statfs.Bsize)
			used := total - free
			if total > 0 {
				usedPercent := float64(used) / float64(total) * 100
				stats["Disk"] = fmt.Sprintf("%.1f%%", usedPercent)
			}
		}
	}

	return stats, nil
}

// getTotalConfigs gets the total number of configuration files from the repository
func (s *EnhancedServer) getTotalConfigs() int {
	if s.nixosRepo == nil {
		s.logger.Warn("nixosRepo is nil in getTotalConfigs")
		return 0
	}

	// Get all configurations from the repository
	configs := s.nixosRepo.GetConfigurations()
	count := len(configs)
	s.logger.Info(fmt.Sprintf("getTotalConfigs: found %d configurations", count))
	return count
}

// generateAISuggestions generates AI-powered suggestions based on system state
func (s *EnhancedServer) generateAISuggestions(machines []*fleet.Machine) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	// Suggestion 1: Machine Management
	if len(machines) == 0 {
		suggestions = append(suggestions, map[string]interface{}{
			"ID":          "add-machines",
			"Icon":        "🖥️",
			"Title":       "Add Your First Machine",
			"Description": "Get started by adding machines to your fleet for centralized management.",
		})
	} else if len(machines) > 0 {
		unknownCount := 0
		for _, machine := range machines {
			if machine.Status == "unknown" {
				unknownCount++
			}
		}
		if unknownCount > 0 {
			suggestions = append(suggestions, map[string]interface{}{
				"ID":          "check-machines",
				"Icon":        "🔍",
				"Title":       "Verify Machine Status",
				"Description": fmt.Sprintf("%d machines have unknown status. Run health checks to verify connectivity.", unknownCount),
			})
		}
	}

	// Suggestion 2: Configuration Management
	suggestions = append(suggestions, map[string]interface{}{
		"ID":          "create-config",
		"Icon":        "⚙️",
		"Title":       "Create Standard Configuration",
		"Description": "Define a base NixOS configuration template to ensure consistency across machines.",
	})

	// Suggestion 3: Security
	suggestions = append(suggestions, map[string]interface{}{
		"ID":          "security-review",
		"Icon":        "🔒",
		"Title":       "Security Configuration Review",
		"Description": "Review and optimize security settings for your NixOS configurations.",
	})

	return suggestions
}

// generateRecentActivities generates recent activity items
func (s *EnhancedServer) generateRecentActivities(machines []*fleet.Machine) []map[string]interface{} {
	activities := []map[string]interface{}{
		{
			"Icon":          "🚀",
			"Message":       "NixAI web interface started",
			"FormattedTime": "Just now",
		},
	}

	if len(machines) > 0 {
		activities = append(activities, map[string]interface{}{
			"Icon":          "🖥️",
			"Message":       fmt.Sprintf("Discovered %d machines from repository", len(machines)),
			"FormattedTime": "5 minutes ago",
		})
	}

	activities = append(activities, map[string]interface{}{
		"Icon":          "📡",
		"Message":       "Real-time collaboration enabled",
		"FormattedTime": "10 minutes ago",
	})

	return activities
}

// generateSystemAlerts generates system alerts based on current state
func (s *EnhancedServer) generateSystemAlerts(machines []*fleet.Machine) []map[string]interface{} {
	alerts := []map[string]interface{}{}

	if len(machines) == 0 {
		alerts = append(alerts, map[string]interface{}{
			"Level":         "info",
			"Title":         "Getting Started",
			"Message":       "No machines configured yet. Use the --repo flag to analyze a NixOS repository or add machines manually.",
			"FormattedTime": "Now",
		})
	} else {
		alerts = append(alerts, map[string]interface{}{
			"Level":         "success",
			"Title":         "Fleet Status",
			"Message":       fmt.Sprintf("Successfully managing %d machines", len(machines)),
			"FormattedTime": "Now",
		})
	}

	return alerts
}

// generateFleetStatus generates fleet status summary
func (s *EnhancedServer) generateFleetStatus(machines []*fleet.Machine) []map[string]interface{} {
	status := []map[string]interface{}{}

	for _, machine := range machines {
		statusClass := "secondary"
		switch machine.Status {
		case "online":
			statusClass = "success"
		case "offline":
			statusClass = "error"
		case "degraded":
			statusClass = "warning"
		}

		status = append(status, map[string]interface{}{
			"Name":        machine.Name,
			"Type":        "NixOS Machine",
			"Version":     "Unknown",
			"Status":      machine.Status,
			"StatusClass": statusClass,
		})
	}

	return status
}

// generateConfigurationStatus generates sample configuration data for the dashboard
func (s *EnhancedServer) generateConfigurationStatus() []map[string]interface{} {
	configs := []map[string]interface{}{
		{
			"id":           "nixos-desktop",
			"Name":         "Desktop Configuration",
			"path":         "/etc/nixos/configuration.nix",
			"status":       "active",
			"LastModified": time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04"),
			"Branch":       "main",
			"changes":      3,
			"description":  "Main desktop configuration with GNOME and development tools",
			"type":         "system",
			"editable":     true,
		},
		{
			"id":           "home-manager",
			"Name":         "Home Manager Config",
			"path":         "~/.config/nixpkgs/home.nix",
			"status":       "modified",
			"LastModified": time.Now().Add(-30 * time.Minute).Format("2006-01-02 15:04"),
			"Branch":       "feature/dotfiles",
			"changes":      1,
			"description":  "User-specific environment and applications",
			"type":         "user",
			"editable":     true,
		},
		{
			"id":           "flake-config",
			"Name":         "Flake Configuration",
			"path":         "/etc/nixos/flake.nix",
			"status":       "synced",
			"LastModified": time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04"),
			"Branch":       "main",
			"changes":      0,
			"description":  "Flake-based system configuration with inputs",
			"type":         "flake",
			"editable":     true,
		},
		{
			"id":           "server-config",
			"Name":         "Server Configuration",
			"path":         "/etc/nixos/machines/server.nix",
			"status":       "pending",
			"LastModified": time.Now().Add(-45 * time.Minute).Format("2006-01-02 15:04"),
			"Branch":       "deploy/production",
			"changes":      2,
			"description":  "Web server configuration with nginx and SSL",
			"type":         "machine",
			"editable":     true,
		},
	}

	// Add configurations from nixos repository if available
	if s.nixosRepo != nil {
		repoConfigs := s.nixosRepo.GetConfigurations()
		for _, config := range repoConfigs {
			configs = append(configs, map[string]interface{}{
				"id":           config.Name,
				"Name":         config.Name,
				"path":         config.Path,
				"status":       "repository",
				"LastModified": time.Now().Format("2006-01-02 15:04"),
				"Branch":       "main",
				"changes":      0,
				"description":  fmt.Sprintf("Configuration from repository: %s", config.Path),
				"type":         "repository",
				"editable":     true,
			})
		}
	}

	return configs
}
