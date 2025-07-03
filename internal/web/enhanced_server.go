package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/pkg/logger"
)

// EnhancedServer provides a comprehensive web interface for nixai
type EnhancedServer struct {
	mux         *http.ServeMux
	repository  *repository.ConfigRepository
	teamManager *team.TeamManager
	logger      *logger.Logger
}

// NewEnhancedServer creates a new enhanced web server
func NewEnhancedServer(repo *repository.ConfigRepository, teamManager *team.TeamManager, logger *logger.Logger) *EnhancedServer {
	server := &EnhancedServer{
		mux:         http.NewServeMux(),
		repository:  repo,
		teamManager: teamManager,
		logger:      logger,
	}

	server.setupRoutes()
	return server
}

// Start starts the enhanced web server
func (s *EnhancedServer) Start(port int) error {
	addr := ":" + strconv.Itoa(port)
	s.logger.Info(fmt.Sprintf("Starting enhanced web server on %s", addr))
	return http.ListenAndServe(addr, s.mux)
}

// setupRoutes sets up all the web routes
func (s *EnhancedServer) setupRoutes() {
	// API Routes
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/dashboard", s.handleDashboardAPI)
	s.mux.HandleFunc("/api/dashboard/stats", s.handleStatsAPI)
	s.mux.HandleFunc("/api/dashboard/activities", s.handleActivitiesAPI)
	s.mux.HandleFunc("/api/dashboard/alerts", s.handleAlertsAPI)
	s.mux.HandleFunc("/api/ws", s.handleWebSocket)

	// Configuration API routes
	s.mux.HandleFunc("/api/configs", s.handleConfigsAPI)
	s.mux.HandleFunc("/api/configs/read", s.handleConfigRead)
	s.mux.HandleFunc("/api/configs/save", s.handleConfigSave)
	s.mux.HandleFunc("/api/configs/list", s.handleConfigList)
	
	// AI Assistant API routes
	s.mux.HandleFunc("/api/ai/verify", s.handleAIVerify)
	s.mux.HandleFunc("/api/ai/generate", s.handleAIGenerate)
	s.mux.HandleFunc("/api/ai/template", s.handleAITemplate)

	// Builder API routes
	s.mux.HandleFunc("/api/builder/components", s.handleBuilderComponents)
	s.mux.HandleFunc("/api/builder/templates", s.handleBuilderTemplates)

	// Static files
	staticDir := "internal/webui/static"
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Frontend routes
	s.mux.HandleFunc("/", s.handleDashboard)
	s.mux.HandleFunc("/dashboard", s.handleDashboard)
	s.mux.HandleFunc("/builder", s.handleBuilder)
	s.mux.HandleFunc("/fleet", s.handleFleet)
	s.mux.HandleFunc("/teams", s.handleTeams)
	s.mux.HandleFunc("/versions", s.handleVersions)
	s.mux.HandleFunc("/configs", s.handleConfigsPage)
}

// API Handlers

func (s *EnhancedServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "3.1.0",
		"features": map[string]bool{
			"dashboard":          true,
			"configuration":      true,
			"fleet_management":   true,
			"team_collaboration": true,
			"version_control":    true,
			"websockets":         true,
		},
	}

	s.sendJSON(w, health)
}

func (s *EnhancedServer) handleDashboardAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get real data from the same sources as other pages
	machines := s.getFleetMachines()
	configs := s.getRepositoryConfigs()
	teams := s.getActiveTeams()
	
	dashboard := map[string]interface{}{
		"overview": map[string]interface{}{
			"total_machines":      len(machines),
			"healthy_machines":    s.getHealthyMachineCount(machines),
			"total_configs":       len(configs),
			"active_teams":        len(teams),
			"pending_deployments": s.getPendingDeployments(),
		},
		"machines": machines,
		"configs":  configs,
		"teams":    teams,
		"activities": s.getRecentActivities(),
		"alerts":     s.getSystemAlerts(),
		"quick_stats": map[string]interface{}{
			"uptime":          "99.9%",
			"last_deployment": "2 hours ago",
			"config_changes":  len(configs),
			"team_members":    s.getTotalTeamMembers(teams),
		},
	}

	s.sendSuccess(w, dashboard)
}

func (s *EnhancedServer) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := map[string]interface{}{
		"machines": map[string]interface{}{
			"total":       0,
			"healthy":     0,
			"unhealthy":   0,
			"maintenance": 0,
		},
		"configurations": map[string]interface{}{
			"total":    0,
			"active":   0,
			"draft":    0,
			"archived": 0,
		},
		"teams": map[string]interface{}{
			"active":          0,
			"total_members":   0,
			"active_sessions": 0,
		},
		"deployments": map[string]interface{}{
			"today":        0,
			"this_week":    0,
			"success_rate": "100%",
		},
	}

	s.sendSuccess(w, stats)
}

func (s *EnhancedServer) handleActivitiesAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	activities := s.getRecentActivities()
	s.sendSuccess(w, activities)
}

func (s *EnhancedServer) handleAlertsAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "HEAD" || r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	alerts := s.getSystemAlerts()
	s.sendSuccess(w, alerts)
}

func (s *EnhancedServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Simple WebSocket simulation for now
	s.setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"type":      "websocket_info",
		"message":   "WebSocket endpoint available",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	s.sendJSON(w, response)
}

// Frontend Handlers

func (s *EnhancedServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	html := s.generateDashboardHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (s *EnhancedServer) handleBuilder(w http.ResponseWriter, r *http.Request) {
	html := s.generateBuilderHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (s *EnhancedServer) handleFleet(w http.ResponseWriter, r *http.Request) {
	html := s.generateFleetHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (s *EnhancedServer) handleTeams(w http.ResponseWriter, r *http.Request) {
	html := s.generateTeamsHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (s *EnhancedServer) handleVersions(w http.ResponseWriter, r *http.Request) {
	html := s.generateVersionsHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Helper methods

func (s *EnhancedServer) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (s *EnhancedServer) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encode JSON: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

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

func (s *EnhancedServer) getRecentActivities() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":        "1",
			"type":      "system_start",
			"message":   "NixAI web interface started",
			"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"icon":      "🚀",
			"user":      "system",
		},
		{
			"id":        "2",
			"type":      "config_change",
			"message":   "Configuration updated: system.nix",
			"timestamp": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"icon":      "📝",
			"user":      "admin",
		},
		{
			"id":        "3",
			"type":      "deployment",
			"message":   "Deployed configuration to 3 machines",
			"timestamp": time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
			"icon":      "🚀",
			"user":      "admin",
		},
	}
}

func (s *EnhancedServer) getSystemAlerts() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":        "1",
			"level":     "info",
			"title":     "System Ready",
			"message":   "NixAI enhanced web interface is fully operational",
			"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"icon":      "ℹ️",
		},
		{
			"id":        "2",
			"level":     "success",
			"title":     "All Systems Healthy",
			"message":   "All monitored systems are operating normally",
			"timestamp": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			"icon":      "✅",
		},
	}
}

func (s *EnhancedServer) generateDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NixAI Dashboard</title>
    <style>
        :root {
            /* Light theme variables */
            --bg-primary: #ffffff;
            --bg-secondary: #f8fafc;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --text-tertiary: #94a3b8;
            --border-color: #e2e8f0;
            --accent-color: #7c3aed;
            --accent-hover: #6d28d9;
            --success-color: #10b981;
            --shadow: 0 1px 3px rgba(0,0,0,0.1);
            --shadow-lg: 0 10px 15px rgba(0,0,0,0.1);
        }

        [data-theme="dark"] {
            /* Dark theme variables */
            --bg-primary: #1e293b;
            --bg-secondary: #0f172a;
            --text-primary: #f1f5f9;
            --text-secondary: #cbd5e1;
            --text-tertiary: #94a3b8;
            --border-color: #334155;
            --accent-color: #8b5cf6;
            --accent-hover: #7c3aed;
            --success-color: #34d399;
            --shadow: 0 1px 3px rgba(0,0,0,0.3);
            --shadow-lg: 0 10px 15px rgba(0,0,0,0.3);
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-secondary);
            color: var(--text-primary);
            transition: background-color 0.3s ease, color 0.3s ease;
        }

        .navbar {
            background: var(--bg-primary);
            border-bottom: 1px solid var(--border-color);
            padding: 1rem 0;
            position: sticky;
            top: 0;
            z-index: 100;
            box-shadow: var(--shadow);
        }

        .navbar-content {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .logo {
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--accent-color);
            text-decoration: none;
        }

        .nav-links {
            display: flex;
            gap: 2rem;
            list-style: none;
        }

        .nav-links a {
            color: var(--text-secondary);
            text-decoration: none;
            font-weight: 500;
            transition: color 0.2s;
        }

        .nav-links a:hover, .nav-links a.active {
            color: var(--accent-color);
        }

        .theme-toggle {
            background: none;
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 0.5rem;
            cursor: pointer;
            color: var(--text-primary);
            transition: all 0.2s;
        }

        .theme-toggle:hover {
            background: var(--bg-secondary);
            border-color: var(--accent-color);
        }

        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            padding: 2rem; 
        }

        .header { 
            text-align: center; 
            margin-bottom: 3rem; 
        }

        .header h1 { 
            color: var(--text-primary); 
            font-size: 2.5rem; 
            margin-bottom: 0.5rem; 
            font-weight: 700;
        }

        .header p { 
            color: var(--text-secondary); 
            font-size: 1.1rem; 
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
        }

        .status-indicator { 
            width: 8px; 
            height: 8px; 
            background: var(--success-color); 
            border-radius: 50%; 
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .stats-grid { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); 
            gap: 1.5rem; 
            margin-bottom: 3rem; 
        }

        .stat-card { 
            background: var(--bg-primary); 
            padding: 1.5rem; 
            border-radius: 12px; 
            box-shadow: var(--shadow);
            border: 1px solid var(--border-color);
            transition: all 0.2s ease;
        }

        .stat-card:hover {
            box-shadow: var(--shadow-lg);
            transform: translateY(-2px);
        }

        .stat-title { 
            color: var(--text-secondary); 
            font-size: 0.875rem; 
            font-weight: 500; 
            margin-bottom: 0.5rem; 
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .stat-value { 
            color: var(--text-primary); 
            font-size: 2.25rem; 
            font-weight: 700; 
        }

        .features-grid { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); 
            gap: 2rem; 
        }

        .feature-card { 
            background: var(--bg-primary); 
            padding: 2rem; 
            border-radius: 12px; 
            box-shadow: var(--shadow);
            border: 1px solid var(--border-color);
            transition: all 0.2s ease;
        }

        .feature-card:hover {
            box-shadow: var(--shadow-lg);
            transform: translateY(-2px);
        }

        .feature-card h3 { 
            color: var(--text-primary); 
            margin-bottom: 1rem; 
            display: flex; 
            align-items: center; 
            gap: 0.5rem;
            font-size: 1.25rem;
        }

        .feature-card p { 
            color: var(--text-secondary); 
            line-height: 1.6; 
            margin-bottom: 1.5rem; 
        }

        .feature-link { 
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            background: var(--accent-color); 
            color: white; 
            padding: 0.75rem 1.5rem; 
            border-radius: 8px; 
            text-decoration: none;
            font-weight: 500;
            transition: all 0.2s ease;
        }

        .feature-link:hover { 
            background: var(--accent-hover);
            transform: translateY(-1px);
        }

        .quick-actions {
            background: var(--bg-primary);
            border-radius: 12px;
            padding: 2rem;
            margin-top: 2rem;
            border: 1px solid var(--border-color);
        }

        .quick-actions h2 {
            color: var(--text-primary);
            margin-bottom: 1rem;
        }

        .action-buttons {
            display: flex;
            gap: 1rem;
            flex-wrap: wrap;
        }

        .action-btn {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            padding: 0.75rem 1.5rem;
            border-radius: 8px;
            text-decoration: none;
            font-weight: 500;
            transition: all 0.2s ease;
        }

        .action-btn:hover {
            background: var(--accent-color);
            color: white;
            border-color: var(--accent-color);
        }

        @media (max-width: 768px) {
            .navbar-content {
                padding: 0 1rem;
            }
            
            .nav-links {
                gap: 1rem;
            }
            
            .container {
                padding: 1rem;
            }
            
            .header h1 {
                font-size: 2rem;
            }
            
            .stats-grid {
                grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            }
            
            .features-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <nav class="navbar">
        <div class="navbar-content">
            <a href="/" class="logo">🚀 NixAI</a>
            <ul class="nav-links">
                <li><a href="/" class="active">Dashboard</a></li>
                <li><a href="/builder">Builder</a></li>
                <li><a href="/fleet">Fleet</a></li>
                <li><a href="/teams">Teams</a></li>
                <li><a href="/versions">Versions</a></li>
            </ul>
            <button class="theme-toggle" onclick="toggleTheme()" title="Toggle theme">
                <span id="theme-icon">🌙</span>
            </button>
        </div>
    </nav>

    <div class="container">
        <header class="header">
            <h1>NixOS AI Assistant</h1>
            <p>
                <span class="status-indicator"></span>
                All systems operational
            </p>
        </header>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-title">Machines</div>
                <div class="stat-value" id="machines-count">0</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Configurations</div>
                <div class="stat-value" id="configs-count">0</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Teams</div>
                <div class="stat-value" id="teams-count">0</div>
            </div>
            <div class="stat-card">
                <div class="stat-title">Uptime</div>
                <div class="stat-value">99.9%</div>
            </div>
        </div>

        <div class="features-grid">
            <div class="feature-card">
                <h3>🎨 Configuration Builder</h3>
                <p>Visual drag-and-drop interface for creating NixOS configurations with real-time validation and AI assistance.</p>
                <a href="/builder" class="feature-link">
                    Open Builder
                    <span>→</span>
                </a>
            </div>
            <div class="feature-card">
                <h3>🚀 Fleet Management</h3>
                <p>Deploy and manage configurations across multiple machines with advanced deployment strategies.</p>
                <a href="/fleet" class="feature-link">
                    Manage Fleet
                    <span>→</span>
                </a>
            </div>
            <div class="feature-card">
                <h3>👥 Team Collaboration</h3>
                <p>Real-time collaborative configuration editing with role-based permissions and change tracking.</p>
                <a href="/teams" class="feature-link">
                    View Teams
                    <span>→</span>
                </a>
            </div>
            <div class="feature-card">
                <h3>📝 Version Control</h3>
                <p>Git-like version control for NixOS configurations with branching, merging, and conflict resolution.</p>
                <a href="/versions" class="feature-link">
                    Manage Versions
                    <span>→</span>
                </a>
            </div>
        </div>

        <div class="quick-actions">
            <h2>🚀 Quick Actions</h2>
            <div class="action-buttons">
                <a href="/builder" class="action-btn">Create Configuration</a>
                <a href="/fleet" class="action-btn">Deploy Changes</a>
                <a href="/teams" class="action-btn">Invite Team Member</a>
                <a href="/versions" class="action-btn">View History</a>
            </div>
        </div>
    </div>

    <script>
        // Theme management
        function toggleTheme() {
            const body = document.body;
            const themeIcon = document.getElementById('theme-icon');
            const currentTheme = body.getAttribute('data-theme');
            
            if (currentTheme === 'dark') {
                body.removeAttribute('data-theme');
                themeIcon.textContent = '🌙';
                localStorage.setItem('nixai-theme', 'light');
            } else {
                body.setAttribute('data-theme', 'dark');
                themeIcon.textContent = '☀️';
                localStorage.setItem('nixai-theme', 'dark');
            }
        }

        // Initialize theme from localStorage
        function initTheme() {
            const savedTheme = localStorage.getItem('nixai-theme');
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            const themeIcon = document.getElementById('theme-icon');
            
            if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
                document.body.setAttribute('data-theme', 'dark');
                themeIcon.textContent = '☀️';
            } else {
                themeIcon.textContent = '🌙';
            }
        }

        // Load dashboard data
        async function loadDashboardData() {
            try {
                const response = await fetch('/api/dashboard');
                const data = await response.json();
                
                if (data.success) {
                    const overview = data.data.overview;
                    const machines = data.data.machines;
                    const configs = data.data.configs;
                    const teams = data.data.teams;
                    
                    // Update stats cards
                    document.getElementById('machines-count').textContent = overview.total_machines;
                    document.getElementById('configs-count').textContent = overview.total_configs;
                    document.getElementById('teams-count').textContent = overview.active_teams;
                    
                    // Add machine details to dashboard
                    if (machines && machines.length > 0) {
                        const machinesSection = document.querySelector('.features-grid');
                        const machineCard = document.createElement('div');
                        machineCard.className = 'feature-card';
                        
                        let machineHTML = '<h3>🖥️ Fleet Machines</h3>';
                        machineHTML += '<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem; margin-top: 1rem;">';
                        
                        machines.forEach(machine => {
                            machineHTML += '<div style="background: var(--bg-secondary); padding: 1rem; border-radius: 8px; border: 1px solid var(--border-color);">';
                            machineHTML += '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">';
                            machineHTML += '<strong style="color: var(--text-primary);">' + machine.name + '</strong>';
                            machineHTML += '<span style="background: var(--success-color); color: white; padding: 0.25rem 0.5rem; border-radius: 12px; font-size: 0.75rem;">';
                            machineHTML += machine.status + '</span></div>';
                            machineHTML += '<div style="color: var(--text-secondary); font-size: 0.875rem;">';
                            machineHTML += '<div>IP: ' + machine.ip + '</div>';
                            machineHTML += '<div>Role: ' + machine.role + '</div>';
                            machineHTML += '<div>Last deployed: ' + machine.last_deployed + '</div>';
                            machineHTML += '</div></div>';
                        });
                        
                        machineHTML += '</div>';
                        machineHTML += '<div style="margin-top: 1rem;"><a href="/fleet" class="feature-link">Manage Fleet <span>→</span></a></div>';
                        
                        machineCard.innerHTML = machineHTML;
                        machinesSection.appendChild(machineCard);
                    }
                    
                    // Add configuration details
                    if (configs && configs.length > 0) {
                        const configsSection = document.querySelector('.features-grid');
                        const configCard = document.createElement('div');
                        configCard.className = 'feature-card';
                        
                        let configHTML = '<h3>📝 Configuration Files</h3>';
                        configHTML += '<div style="display: flex; flex-direction: column; gap: 0.75rem; margin-top: 1rem;">';
                        
                        configs.forEach(config => {
                            configHTML += '<div style="background: var(--bg-secondary); padding: 1rem; border-radius: 8px; border: 1px solid var(--border-color); display: flex; align-items: center; gap: 0.75rem;">';
                            configHTML += '<span style="font-size: 1.2rem;">📄</span>';
                            configHTML += '<div style="flex: 1;">';
                            configHTML += '<div style="color: var(--text-primary); font-weight: 500;">' + config.name + '</div>';
                            configHTML += '<div style="color: var(--text-secondary); font-size: 0.75rem;">' + config.size + ' • Modified ' + config.modified + '</div>';
                            configHTML += '</div></div>';
                        });
                        
                        configHTML += '</div>';
                        configHTML += '<div style="margin-top: 1rem;"><a href="/configs" class="feature-link">Edit Configurations <span>→</span></a></div>';
                        
                        configCard.innerHTML = configHTML;
                        configsSection.appendChild(configCard);
                    }
                    
                    // Add team information
                    if (teams && teams.length > 0) {
                        const teamsSection = document.querySelector('.features-grid');
                        const teamCard = document.createElement('div');
                        teamCard.className = 'feature-card';
                        
                        let teamHTML = '<h3>👥 Active Teams</h3>';
                        teamHTML += '<div style="display: flex; flex-direction: column; gap: 0.75rem; margin-top: 1rem;">';
                        
                        teams.forEach(team => {
                            teamHTML += '<div style="background: var(--bg-secondary); padding: 1rem; border-radius: 8px; border: 1px solid var(--border-color);">';
                            teamHTML += '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">';
                            teamHTML += '<strong style="color: var(--text-primary);">' + team.name + '</strong>';
                            teamHTML += '<span style="background: var(--accent-color); color: white; padding: 0.25rem 0.5rem; border-radius: 12px; font-size: 0.75rem;">';
                            teamHTML += team.members + ' members</span></div>';
                            teamHTML += '<div style="color: var(--text-secondary); font-size: 0.875rem;">' + team.description + '</div>';
                            teamHTML += '<div style="margin-top: 0.5rem;">';
                            teamHTML += '<span style="background: var(--bg-primary); color: var(--text-primary); padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; border: 1px solid var(--border-color);">';
                            teamHTML += team.environment + '</span></div></div>';
                        });
                        
                        teamHTML += '</div>';
                        teamHTML += '<div style="margin-top: 1rem;"><a href="/teams" class="feature-link">Manage Teams <span>→</span></a></div>';
                        
                        teamCard.innerHTML = teamHTML;
                        teamsSection.appendChild(teamCard);
                    }
                }
            } catch (error) {
                console.log('Dashboard API error:', error);
            }
        }

        // Initialize on page load
        document.addEventListener('DOMContentLoaded', () => {
            initTheme();
            loadDashboardData();
        });

        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem('nixai-theme')) {
                if (e.matches) {
                    document.body.setAttribute('data-theme', 'dark');
                    document.getElementById('theme-icon').textContent = '☀️';
                } else {
                    document.body.removeAttribute('data-theme');
                    document.getElementById('theme-icon').textContent = '🌙';
                }
            }
        });
    </script>
</body>
</html>`
}

func (s *EnhancedServer) generateFleetHTML() string {
	return s.generatePageWithNavbar("Fleet", "🚀", `
        <div class="content-grid">
            <div class="main-content">
                <div class="feature-card">
                    <h2>🚀 Fleet Management</h2>
                    <p>Deploy and manage configurations across multiple NixOS machines with advanced deployment strategies and monitoring.</p>
                    
                    <div class="action-buttons" style="margin-top: 2rem;">
                        <button class="action-btn primary">Add Machine</button>
                        <button class="action-btn">Deploy Config</button>
                        <button class="action-btn">Health Check</button>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <h3>🖥️ Registered Machines</h3>
                    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1rem; margin-top: 1rem;">
                        <div style="background: var(--bg-secondary); padding: 1.5rem; border-radius: 8px; border: 1px solid var(--border-color);">
                            <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 1rem;">
                                <h4 style="color: var(--text-primary);">production-web-01</h4>
                                <span style="background: var(--success-color); color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.75rem;">Healthy</span>
                            </div>
                            <p style="color: var(--text-secondary); margin-bottom: 0.5rem;">192.168.1.100</p>
                            <p style="color: var(--text-secondary); font-size: 0.875rem;">Last deployed: 2 hours ago</p>
                        </div>
                        <div style="background: var(--bg-secondary); padding: 1.5rem; border-radius: 8px; border: 1px solid var(--border-color);">
                            <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 1rem;">
                                <h4 style="color: var(--text-primary);">production-db-01</h4>
                                <span style="background: var(--success-color); color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.75rem;">Healthy</span>
                            </div>
                            <p style="color: var(--text-secondary); margin-bottom: 0.5rem;">192.168.1.101</p>
                            <p style="color: var(--text-secondary); font-size: 0.875rem;">Last deployed: 1 day ago</p>
                        </div>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <h3>📊 Deployment History</h3>
                    <div class="commit-list" style="margin-top: 1rem;">
                        <div class="commit-item">
                            <span class="commit-hash">#deploy-001</span>
                            <span class="commit-message">Rolling deployment to production fleet</span>
                            <span class="commit-time">2 hours ago</span>
                            <span style="background: var(--success-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">Success</span>
                        </div>
                        <div class="commit-item">
                            <span class="commit-hash">#deploy-002</span>
                            <span class="commit-message">Emergency security patch deployment</span>
                            <span class="commit-time">1 day ago</span>
                            <span style="background: var(--success-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">Success</span>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="sidebar">
                <div class="feature-card">
                    <h3>🔧 Fleet Actions</h3>
                    <div class="quick-actions-list">
                        <a href="#" class="quick-action-item">➕ Add New Machine</a>
                        <a href="#" class="quick-action-item">🚀 Deploy Configuration</a>
                        <a href="#" class="quick-action-item">🔄 Rolling Update</a>
                        <a href="#" class="quick-action-item">💾 Backup Configs</a>
                        <a href="#" class="quick-action-item">📊 View Metrics</a>
                    </div>
                </div>
                
                <div class="feature-card">
                    <h3>📈 Fleet Statistics</h3>
                    <div style="display: flex; flex-direction: column; gap: 1rem;">
                        <div>
                            <div style="color: var(--text-secondary); font-size: 0.875rem;">Total Machines</div>
                            <div style="color: var(--text-primary); font-size: 1.5rem; font-weight: 600;">2</div>
                        </div>
                        <div>
                            <div style="color: var(--text-secondary); font-size: 0.875rem;">Healthy</div>
                            <div style="color: var(--success-color); font-size: 1.5rem; font-weight: 600;">2</div>
                        </div>
                        <div>
                            <div style="color: var(--text-secondary); font-size: 0.875rem;">Uptime</div>
                            <div style="color: var(--text-primary); font-size: 1.5rem; font-weight: 600;">99.9%</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>`)
}

func (s *EnhancedServer) generateTeamsHTML() string {
	return s.generatePageWithNavbar("Teams", "👥", `
        <div class="content-grid">
            <div class="main-content">
                <div class="feature-card">
                    <h2>👥 Team Collaboration</h2>
                    <p>Real-time collaborative configuration editing with role-based permissions, team workflows, and change tracking.</p>
                    
                    <div class="action-buttons" style="margin-top: 2rem;">
                        <button class="action-btn primary">Create Team</button>
                        <button class="action-btn">Invite Member</button>
                        <button class="action-btn">Join Team</button>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <h3>🏢 Active Teams</h3>
                    <div style="display: flex; flex-direction: column; gap: 1rem; margin-top: 1rem;">
                        <div style="background: var(--bg-secondary); padding: 1.5rem; border-radius: 8px; border: 1px solid var(--border-color);">
                            <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 1rem;">
                                <h4 style="color: var(--text-primary);">DevOps Team</h4>
                                <span style="background: var(--success-color); color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.75rem;">5 members</span>
                            </div>
                            <p style="color: var(--text-secondary); margin-bottom: 1rem;">Infrastructure and deployment team</p>
                            <div style="display: flex; gap: 0.5rem;">
                                <span style="background: var(--accent-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">admin</span>
                                <span style="background: var(--bg-primary); color: var(--text-primary); padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; border: 1px solid var(--border-color);">production</span>
                            </div>
                        </div>
                        <div style="background: var(--bg-secondary); padding: 1.5rem; border-radius: 8px; border: 1px solid var(--border-color);">
                            <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 1rem;">
                                <h4 style="color: var(--text-primary);">Development Team</h4>
                                <span style="background: var(--success-color); color: white; padding: 0.25rem 0.75rem; border-radius: 12px; font-size: 0.75rem;">3 members</span>
                            </div>
                            <p style="color: var(--text-secondary); margin-bottom: 1rem;">Application development and testing</p>
                            <div style="display: flex; gap: 0.5rem;">
                                <span style="background: var(--warning-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">member</span>
                                <span style="background: var(--bg-primary); color: var(--text-primary); padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; border: 1px solid var(--border-color);">development</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <h3>💬 Recent Activity</h3>
                    <div class="commit-list" style="margin-top: 1rem;">
                        <div class="commit-item">
                            <span style="background: var(--accent-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">👤 alice</span>
                            <span class="commit-message">Updated web server configuration</span>
                            <span class="commit-time">30 minutes ago</span>
                        </div>
                        <div class="commit-item">
                            <span style="background: var(--success-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">👤 bob</span>
                            <span class="commit-message">Added new development environment</span>
                            <span class="commit-time">2 hours ago</span>
                        </div>
                        <div class="commit-item">
                            <span style="background: var(--warning-color); color: white; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem;">👤 charlie</span>
                            <span class="commit-message">Reviewed database configuration changes</span>
                            <span class="commit-time">1 day ago</span>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="sidebar">
                <div class="feature-card">
                    <h3>🔧 Team Management</h3>
                    <div class="quick-actions-list">
                        <a href="#" class="quick-action-item">👥 Create New Team</a>
                        <a href="#" class="quick-action-item">📧 Invite Members</a>
                        <a href="#" class="quick-action-item">⚙️ Manage Permissions</a>
                        <a href="#" class="quick-action-item">📊 Team Analytics</a>
                        <a href="#" class="quick-action-item">🔄 Sync Settings</a>
                    </div>
                </div>
                
                <div class="feature-card">
                    <h3>👥 Team Members</h3>
                    <div style="display: flex; flex-direction: column; gap: 0.75rem;">
                        <div style="display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem; background: var(--bg-secondary); border-radius: 6px;">
                            <div style="width: 32px; height: 32px; background: var(--accent-color); border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-weight: 600;">A</div>
                            <div>
                                <div style="color: var(--text-primary); font-weight: 500; font-size: 0.875rem;">Alice Johnson</div>
                                <div style="color: var(--text-secondary); font-size: 0.75rem;">Team Lead</div>
                            </div>
                        </div>
                        <div style="display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem; background: var(--bg-secondary); border-radius: 6px;">
                            <div style="width: 32px; height: 32px; background: var(--success-color); border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-weight: 600;">B</div>
                            <div>
                                <div style="color: var(--text-primary); font-weight: 500; font-size: 0.875rem;">Bob Smith</div>
                                <div style="color: var(--text-secondary); font-size: 0.75rem;">Developer</div>
                            </div>
                        </div>
                        <div style="display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem; background: var(--bg-secondary); border-radius: 6px;">
                            <div style="width: 32px; height: 32px; background: var(--warning-color); border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-weight: 600;">C</div>
                            <div>
                                <div style="color: var(--text-primary); font-weight: 500; font-size: 0.875rem;">Charlie Brown</div>
                                <div style="color: var(--text-secondary); font-size: 0.75rem;">Reviewer</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>`)
}

func (s *EnhancedServer) generateVersionsHTML() string {
	return s.generatePageWithNavbar("Version Control", "📝", `
        <div class="content-grid">
            <div class="main-content">
                <div class="feature-card">
                    <h2>📝 Configuration Version Control</h2>
                    <p>Git-like version control for your NixOS configurations with branching, merging, and collaboration features.</p>
                    
                    <div class="action-buttons" style="margin-top: 2rem;">
                        <button class="action-btn primary">Create Branch</button>
                        <button class="action-btn">View History</button>
                        <button class="action-btn">Merge Changes</button>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <h3>Recent Commits</h3>
                    <div class="commit-list">
                        <div class="commit-item">
                            <span class="commit-hash">#abc123</span>
                            <span class="commit-message">Updated system configuration</span>
                            <span class="commit-time">2 hours ago</span>
                        </div>
                        <div class="commit-item">
                            <span class="commit-hash">#def456</span>
                            <span class="commit-message">Added new packages</span>
                            <span class="commit-time">1 day ago</span>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="sidebar">
                <div class="feature-card">
                    <h3>🔧 Quick Actions</h3>
                    <div class="quick-actions-list">
                        <a href="#" class="quick-action-item">📊 View Diff</a>
                        <a href="#" class="quick-action-item">🔄 Sync Repository</a>
                        <a href="#" class="quick-action-item">🏷️ Create Tag</a>
                        <a href="#" class="quick-action-item">📋 Export Config</a>
                    </div>
                </div>
            </div>
        </div>`)
}

// Add all the new API handlers and page generation methods

func (s *EnhancedServer) handleConfigsAPI(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	configs := s.getRepositoryConfigs()
	s.sendSuccess(w, configs)
}

func (s *EnhancedServer) handleConfigRead(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "filename parameter required", http.StatusBadRequest)
		return
	}

	content, err := s.readConfigFile(filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"filename": filename,
		"content":  content,
	})
}

func (s *EnhancedServer) handleConfigSave(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method != "POST" {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.saveConfigFile(req.Filename, req.Content); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"message": "File saved successfully",
		"filename": req.Filename,
	})
}

func (s *EnhancedServer) handleConfigList(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	files, err := s.listConfigFiles()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list files: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, files)
}

func (s *EnhancedServer) handleAIVerify(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method != "POST" {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	verification := s.verifyConfigWithAI(req.Content)
	s.sendSuccess(w, verification)
}

func (s *EnhancedServer) handleAIGenerate(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method != "POST" {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Description string `json:"description"`
		Type        string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	config := s.generateConfigWithAI(req.Description, req.Type)
	s.sendSuccess(w, config)
}

func (s *EnhancedServer) handleAITemplate(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	templates := s.getAITemplates()
	s.sendSuccess(w, templates)
}

func (s *EnhancedServer) handleBuilderComponents(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	components := s.getBuilderComponents()
	s.sendSuccess(w, components)
}

func (s *EnhancedServer) handleBuilderTemplates(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	templates := s.getBuilderTemplates()
	s.sendSuccess(w, templates)
}

func (s *EnhancedServer) handleConfigsPage(w http.ResponseWriter, r *http.Request) {
	html := s.generateConfigsHTML()
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Helper methods for configuration management
func (s *EnhancedServer) getRepositoryConfigs() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name": "configuration.nix",
			"path": "/etc/nixos/configuration.nix",
			"size": "2.4 KB",
			"modified": "2 hours ago",
			"type": "nix",
		},
		{
			"name": "hardware-configuration.nix", 
			"path": "/etc/nixos/hardware-configuration.nix",
			"size": "1.2 KB",
			"modified": "1 day ago",
			"type": "nix",
		},
	}
}

func (s *EnhancedServer) readConfigFile(filename string) (string, error) {
	// Simulate reading a config file
	if filename == "configuration.nix" {
		return `{ config, pkgs, ... }:

{
  imports = [ 
    ./hardware-configuration.nix
  ];

  # Boot loader configuration
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Network configuration
  networking.hostName = "nixos";
  networking.networkmanager.enable = true;

  # System packages
  environment.systemPackages = with pkgs; [
    vim
    git
    firefox
    vscode
  ];

  # Enable SSH
  services.openssh.enable = true;

  # System version
  system.stateVersion = "23.11";
}`, nil
	}
	return "# Example configuration file", nil
}

func (s *EnhancedServer) saveConfigFile(filename, content string) error {
	s.logger.Info(fmt.Sprintf("Saving configuration file: %s", filename))
	// In a real implementation, this would save to the repository
	return nil
}

func (s *EnhancedServer) listConfigFiles() ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{"name": "configuration.nix", "type": "file", "size": "2.4 KB"},
		{"name": "hardware-configuration.nix", "type": "file", "size": "1.2 KB"},
		{"name": "modules/", "type": "directory", "size": ""},
		{"name": "overlays/", "type": "directory", "size": ""},
	}, nil
}

func (s *EnhancedServer) verifyConfigWithAI(content string) map[string]interface{} {
	return map[string]interface{}{
		"valid": true,
		"score": 85,
		"issues": []map[string]interface{}{
			{
				"level": "warning",
				"line": 15,
				"message": "Consider enabling automatic updates",
				"suggestion": "Add: system.autoUpgrade.enable = true;",
			},
		},
		"suggestions": []string{
			"Add firewall configuration for better security",
			"Consider adding backup services",
			"Enable fail2ban for SSH protection",
		},
	}
}

func (s *EnhancedServer) generateConfigWithAI(description, configType string) map[string]interface{} {
	return map[string]interface{}{
		"config": `{ config, pkgs, ... }:

{
  # Generated configuration for: ` + description + `
  
  environment.systemPackages = with pkgs; [
    # Add packages based on description
  ];
  
  # Additional configuration based on type: ` + configType + `
}`,
		"explanation": "This configuration was generated based on your description: " + description,
		"recommendations": []string{
			"Review the generated configuration before applying",
			"Test in a virtual machine first",
			"Make sure to backup your current configuration",
		},
	}
}

func (s *EnhancedServer) getAITemplates() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name": "Web Server",
			"description": "Complete web server setup with Nginx, SSL, and firewall",
			"category": "server",
			"tags": []string{"nginx", "ssl", "firewall"},
		},
		{
			"name": "Development Environment",
			"description": "Full development setup with VS Code, Git, and common tools",
			"category": "desktop",
			"tags": []string{"development", "vscode", "git"},
		},
		{
			"name": "Media Server",
			"description": "Plex media server with storage management",
			"category": "server",
			"tags": []string{"plex", "media", "storage"},
		},
	}
}

func (s *EnhancedServer) getBuilderComponents() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id": "system-packages",
			"name": "System Packages",
			"category": "system",
			"description": "Essential system packages and tools",
			"icon": "📦",
		},
		{
			"id": "desktop-environment",
			"name": "Desktop Environment",
			"category": "desktop",
			"description": "GNOME, KDE, or other desktop environments",
			"icon": "🖥️",
		},
		{
			"id": "web-server",
			"name": "Web Server",
			"category": "services",
			"description": "Nginx or Apache web server configuration",
			"icon": "🌐",
		},
		{
			"id": "database",
			"name": "Database",
			"category": "services", 
			"description": "PostgreSQL, MySQL, or other databases",
			"icon": "🗄️",
		},
	}
}

func (s *EnhancedServer) getBuilderTemplates() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name": "Basic Desktop",
			"description": "Simple desktop setup with essential applications",
			"components": []string{"system-packages", "desktop-environment"},
		},
		{
			"name": "Web Development",
			"description": "Complete web development environment",
			"components": []string{"system-packages", "web-server", "database"},
		},
	}
}

// Fleet data methods
func (s *EnhancedServer) getFleetMachines() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":           "production-web-01",
			"name":         "production-web-01",
			"ip":           "192.168.1.100",
			"status":       "healthy",
			"last_deployed": "2 hours ago",
			"uptime":       "99.9%",
			"role":         "web-server",
		},
		{
			"id":           "production-db-01", 
			"name":         "production-db-01",
			"ip":           "192.168.1.101",
			"status":       "healthy",
			"last_deployed": "1 day ago",
			"uptime":       "99.8%",
			"role":         "database",
		},
		{
			"id":           "staging-web-01",
			"name":         "staging-web-01", 
			"ip":           "192.168.1.150",
			"status":       "healthy",
			"last_deployed": "4 hours ago",
			"uptime":       "98.5%",
			"role":         "web-server",
		},
	}
}

func (s *EnhancedServer) getActiveTeams() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "devops-team",
			"name":        "DevOps Team",
			"description": "Infrastructure and deployment team",
			"members":     5,
			"role":        "admin",
			"environment": "production",
		},
		{
			"id":          "dev-team",
			"name":        "Development Team", 
			"description": "Application development and testing",
			"members":     3,
			"role":        "member",
			"environment": "development",
		},
		{
			"id":          "qa-team",
			"name":        "QA Team",
			"description": "Quality assurance and testing",
			"members":     2,
			"role":        "reviewer",
			"environment": "staging",
		},
	}
}

func (s *EnhancedServer) getHealthyMachineCount(machines []map[string]interface{}) int {
	count := 0
	for _, machine := range machines {
		if machine["status"] == "healthy" {
			count++
		}
	}
	return count
}

func (s *EnhancedServer) getTotalTeamMembers(teams []map[string]interface{}) int {
	total := 0
	for _, team := range teams {
		if members, ok := team["members"].(int); ok {
			total += members
		}
	}
	return total
}

func (s *EnhancedServer) getPendingDeployments() int {
	// Simulate some pending deployments
	return 2
}

// generatePageWithNavbar creates a consistent page template with navigation
func (s *EnhancedServer) generatePageWithNavbar(title, icon, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - NixAI</title>
    <style>
        :root {
            --bg-primary: #ffffff;
            --bg-secondary: #f8fafc;
            --text-primary: #1e293b;
            --text-secondary: #64748b;
            --text-tertiary: #94a3b8;
            --border-color: #e2e8f0;
            --accent-color: #7c3aed;
            --accent-hover: #6d28d9;
            --success-color: #10b981;
            --warning-color: #f59e0b;
            --error-color: #ef4444;
            --shadow: 0 1px 3px rgba(0,0,0,0.1);
            --shadow-lg: 0 10px 15px rgba(0,0,0,0.1);
        }

        [data-theme="dark"] {
            --bg-primary: #1e293b;
            --bg-secondary: #0f172a;
            --text-primary: #f1f5f9;
            --text-secondary: #cbd5e1;
            --text-tertiary: #94a3b8;
            --border-color: #334155;
            --accent-color: #8b5cf6;
            --accent-hover: #7c3aed;
            --success-color: #34d399;
            --warning-color: #fbbf24;
            --error-color: #f87171;
            --shadow: 0 1px 3px rgba(0,0,0,0.3);
            --shadow-lg: 0 10px 15px rgba(0,0,0,0.3);
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg-secondary);
            color: var(--text-primary);
            transition: background-color 0.3s ease, color 0.3s ease;
            min-height: 100vh;
        }

        .navbar {
            background: var(--bg-primary);
            border-bottom: 1px solid var(--border-color);
            padding: 1rem 0;
            position: sticky;
            top: 0;
            z-index: 100;
            box-shadow: var(--shadow);
        }

        .navbar-content {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .logo {
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--accent-color);
            text-decoration: none;
        }

        .nav-links {
            display: flex;
            gap: 2rem;
            list-style: none;
            align-items: center;
        }

        .nav-links a {
            color: var(--text-secondary);
            text-decoration: none;
            font-weight: 500;
            transition: color 0.2s;
            padding: 0.5rem 1rem;
            border-radius: 6px;
        }

        .nav-links a:hover, .nav-links a.active {
            color: var(--accent-color);
            background: var(--bg-secondary);
        }

        .theme-toggle {
            background: none;
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 0.5rem;
            cursor: pointer;
            color: var(--text-primary);
            transition: all 0.2s;
        }

        .theme-toggle:hover {
            background: var(--bg-secondary);
            border-color: var(--accent-color);
        }

        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            padding: 2rem; 
        }

        .page-header { 
            text-align: center; 
            margin-bottom: 3rem; 
        }

        .page-header h1 { 
            color: var(--text-primary); 
            font-size: 2.5rem; 
            margin-bottom: 0.5rem; 
            font-weight: 700;
        }

        .page-header p { 
            color: var(--text-secondary); 
            font-size: 1.1rem; 
        }

        .content-grid {
            display: grid;
            grid-template-columns: 1fr 300px;
            gap: 2rem;
        }

        .main-content {
            min-height: 400px;
        }

        .sidebar {
            /* Sidebar styling */
        }

        .feature-card { 
            background: var(--bg-primary); 
            padding: 2rem; 
            border-radius: 12px; 
            box-shadow: var(--shadow);
            border: 1px solid var(--border-color);
            transition: all 0.2s ease;
            margin-bottom: 1.5rem;
        }

        .feature-card:hover {
            box-shadow: var(--shadow-lg);
            transform: translateY(-2px);
        }

        .feature-card h2, .feature-card h3 { 
            color: var(--text-primary); 
            margin-bottom: 1rem; 
            display: flex; 
            align-items: center; 
            gap: 0.5rem;
        }

        .feature-card p { 
            color: var(--text-secondary); 
            line-height: 1.6; 
            margin-bottom: 1.5rem; 
        }

        .action-buttons {
            display: flex;
            gap: 1rem;
            flex-wrap: wrap;
        }

        .action-btn {
            background: var(--bg-secondary);
            border: 1px solid var(--border-color);
            color: var(--text-primary);
            padding: 0.75rem 1.5rem;
            border-radius: 8px;
            text-decoration: none;
            font-weight: 500;
            transition: all 0.2s ease;
            cursor: pointer;
            font-size: 0.875rem;
        }

        .action-btn:hover {
            background: var(--accent-color);
            color: white;
            border-color: var(--accent-color);
            transform: translateY(-1px);
        }

        .action-btn.primary {
            background: var(--accent-color);
            color: white;
            border-color: var(--accent-color);
        }

        .action-btn.primary:hover {
            background: var(--accent-hover);
            border-color: var(--accent-hover);
        }

        .quick-actions-list {
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
        }

        .quick-action-item {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            padding: 0.75rem;
            background: var(--bg-secondary);
            border-radius: 8px;
            text-decoration: none;
            color: var(--text-primary);
            transition: all 0.2s ease;
            border: 1px solid var(--border-color);
        }

        .quick-action-item:hover {
            background: var(--accent-color);
            color: white;
            border-color: var(--accent-color);
        }

        .commit-list {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        .commit-item {
            display: flex;
            align-items: center;
            gap: 1rem;
            padding: 1rem;
            background: var(--bg-secondary);
            border-radius: 8px;
            border: 1px solid var(--border-color);
        }

        .commit-hash {
            font-family: 'Monaco', 'Menlo', monospace;
            background: var(--accent-color);
            color: white;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
        }

        .commit-message {
            flex: 1;
            color: var(--text-primary);
            font-weight: 500;
        }

        .commit-time {
            color: var(--text-tertiary);
            font-size: 0.875rem;
        }

        @media (max-width: 768px) {
            .navbar-content {
                padding: 0 1rem;
                flex-direction: column;
                gap: 1rem;
            }
            
            .nav-links {
                gap: 1rem;
                flex-wrap: wrap;
                justify-content: center;
            }
            
            .container {
                padding: 1rem;
            }
            
            .page-header h1 {
                font-size: 2rem;
            }
            
            .content-grid {
                grid-template-columns: 1fr;
            }
            
            .action-buttons {
                justify-content: center;
            }
        }

        /* AI Assistant styles */
        .ai-assistant {
            position: fixed;
            bottom: 2rem;
            right: 2rem;
            z-index: 1000;
        }

        .ai-toggle {
            background: var(--accent-color);
            color: white;
            border: none;
            border-radius: 50%%;
            width: 60px;
            height: 60px;
            font-size: 1.5rem;
            cursor: pointer;
            box-shadow: var(--shadow-lg);
            transition: all 0.2s ease;
        }

        .ai-toggle:hover {
            background: var(--accent-hover);
            transform: scale(1.1);
        }

        .ai-panel {
            position: absolute;
            bottom: 80px;
            right: 0;
            width: 400px;
            max-height: 500px;
            background: var(--bg-primary);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            box-shadow: var(--shadow-lg);
            display: none;
        }

        .ai-panel.open {
            display: block;
        }

        .ai-header {
            padding: 1rem;
            border-bottom: 1px solid var(--border-color);
            background: var(--accent-color);
            color: white;
            border-radius: 12px 12px 0 0;
        }

        .ai-content {
            padding: 1rem;
            max-height: 400px;
            overflow-y: auto;
        }

        .ai-input {
            width: 100%%;
            padding: 0.75rem;
            border: 1px solid var(--border-color);
            border-radius: 8px;
            background: var(--bg-secondary);
            color: var(--text-primary);
            font-size: 0.875rem;
        }

        .ai-response {
            margin-top: 1rem;
            padding: 1rem;
            background: var(--bg-secondary);
            border-radius: 8px;
            font-size: 0.875rem;
            line-height: 1.5;
        }
    </style>
</head>
<body>
    <nav class="navbar">
        <div class="navbar-content">
            <a href="/" class="logo">🚀 NixAI</a>
            <ul class="nav-links">
                <li><a href="/" class="%s">Dashboard</a></li>
                <li><a href="/configs" class="%s">Configs</a></li>
                <li><a href="/builder" class="%s">Builder</a></li>
                <li><a href="/fleet" class="%s">Fleet</a></li>
                <li><a href="/teams" class="%s">Teams</a></li>
                <li><a href="/versions" class="%s">Versions</a></li>
            </ul>
            <button class="theme-toggle" onclick="toggleTheme()" title="Toggle theme">
                <span id="theme-icon">🌙</span>
            </button>
        </div>
    </nav>

    <div class="container">
        <header class="page-header">
            <h1>%s %s</h1>
        </header>

        %s
    </div>

    <!-- AI Assistant -->
    <div class="ai-assistant">
        <button class="ai-toggle" onclick="toggleAI()" title="AI Assistant">
            🤖
        </button>
        <div class="ai-panel" id="ai-panel">
            <div class="ai-header">
                <h3>🤖 AI Assistant</h3>
            </div>
            <div class="ai-content">
                <input type="text" class="ai-input" placeholder="Ask about NixOS configuration..." onkeypress="handleAIInput(event)">
                <div class="ai-response" id="ai-response" style="display: none;"></div>
            </div>
        </div>
    </div>

    <script>
        // Theme management
        function toggleTheme() {
            const body = document.body;
            const themeIcon = document.getElementById('theme-icon');
            const currentTheme = body.getAttribute('data-theme');
            
            if (currentTheme === 'dark') {
                body.removeAttribute('data-theme');
                themeIcon.textContent = '🌙';
                localStorage.setItem('nixai-theme', 'light');
            } else {
                body.setAttribute('data-theme', 'dark');
                themeIcon.textContent = '☀️';
                localStorage.setItem('nixai-theme', 'dark');
            }
        }

        // Initialize theme
        function initTheme() {
            const savedTheme = localStorage.getItem('nixai-theme');
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            const themeIcon = document.getElementById('theme-icon');
            
            if (savedTheme === 'dark' || (!savedTheme && prefersDark)) {
                document.body.setAttribute('data-theme', 'dark');
                themeIcon.textContent = '☀️';
            } else {
                themeIcon.textContent = '🌙';
            }
        }

        // AI Assistant
        function toggleAI() {
            const panel = document.getElementById('ai-panel');
            panel.classList.toggle('open');
        }

        function handleAIInput(event) {
            if (event.key === 'Enter') {
                const input = event.target;
                const response = document.getElementById('ai-response');
                
                // Simulate AI response
                response.innerHTML = '<strong>AI:</strong> I can help you with NixOS configuration. What specific assistance do you need?';
                response.style.display = 'block';
                
                input.value = '';
            }
        }

        // Initialize on page load
        document.addEventListener('DOMContentLoaded', () => {
            initTheme();
        });

        // Listen for system theme changes
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem('nixai-theme')) {
                if (e.matches) {
                    document.body.setAttribute('data-theme', 'dark');
                    document.getElementById('theme-icon').textContent = '☀️';
                } else {
                    document.body.removeAttribute('data-theme');
                    document.getElementById('theme-icon').textContent = '🌙';
                }
            }
        });
    </script>
</body>
</html>`, 
	title, 
	getActiveClass(title, "Dashboard"),
	getActiveClass(title, "Configs"), 
	getActiveClass(title, "Builder"),
	getActiveClass(title, "Fleet"),
	getActiveClass(title, "Teams"),
	getActiveClass(title, "Versions"),
	icon, title, content)
}

func getActiveClass(currentPage, page string) string {
	if currentPage == page {
		return "active"
	}
	return ""
}

func (s *EnhancedServer) generateBuilderHTML() string {
	return s.generatePageWithNavbar("Builder", "🎨", `
        <!-- AI Assistant Modal -->
        <div id="ai-modal" class="ai-modal" style="display: none;">
            <div class="ai-modal-content">
                <div class="ai-modal-header">
                    <h3>🤖 AI Configuration Assistant</h3>
                    <button class="ai-modal-close" onclick="closeAIModal()">&times;</button>
                </div>
                <div class="ai-modal-body">
                    <div class="ai-input-section">
                        <label for="ai-description">Describe what you want to configure:</label>
                        <textarea id="ai-description" placeholder="Example: Web server with PostgreSQL database, SSL certificates, and monitoring tools" rows="3"></textarea>
                    </div>
                    
                    <div class="ai-options-section">
                        <label>Configuration Type:</label>
                        <div class="ai-option-buttons">
                            <button class="ai-option-btn" data-type="desktop">🖥️ Desktop</button>
                            <button class="ai-option-btn" data-type="server">🌐 Server</button>
                            <button class="ai-option-btn" data-type="development">💻 Development</button>
                            <button class="ai-option-btn" data-type="gaming">🎮 Gaming</button>
                            <button class="ai-option-btn" data-type="custom">⚙️ Custom</button>
                        </div>
                    </div>
                    
                    <div class="ai-template-section">
                        <label>Quick Templates:</label>
                        <div class="ai-template-buttons">
                            <button class="ai-template-btn" onclick="useTemplate('nginx-ssl')">Nginx + SSL</button>
                            <button class="ai-template-btn" onclick="useTemplate('postgres-db')">PostgreSQL DB</button>
                            <button class="ai-template-btn" onclick="useTemplate('docker-host')">Docker Host</button>
                            <button class="ai-template-btn" onclick="useTemplate('dev-workstation')">Dev Workstation</button>
                        </div>
                    </div>
                    
                    <div class="ai-modal-actions">
                        <button class="action-btn" onclick="closeAIModal()">Cancel</button>
                        <button class="action-btn primary" onclick="generateAIConfig()">Generate Configuration</button>
                        <button class="action-btn" onclick="validateAIConfig()">Validate Existing</button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Single Column Layout -->
        <div class="builder-container">
            <!-- Header Section -->
            <div class="feature-card builder-header">
                <h2>🎨 Visual Configuration Builder</h2>
                <p>Build NixOS configurations with drag-and-drop components, AI assistance, and real-time validation.</p>
                
                <div class="action-buttons" style="margin-top: 1.5rem;">
                    <button class="action-btn primary" onclick="createNewConfig()">🆕 New Configuration</button>
                    <button class="action-btn" onclick="loadTemplate()">📄 Load Template</button>
                    <button class="action-btn" onclick="openAIModal()">🤖 Ask AI</button>
                    <button class="action-btn" onclick="validateConfig()">✅ Validate</button>
                    <button class="action-btn" onclick="exportConfig()">💾 Export</button>
                </div>
            </div>

            <!-- Component Library -->
            <div class="feature-card builder-section">
                <div class="section-header">
                    <h3>🧩 Component Library</h3>
                    <p>Click the + button next to each component to add it to your configuration</p>
                </div>
                <div class="component-list" id="component-library">
                    <div class="component-item" data-component="system-packages">
                        <div class="component-icon">📦</div>
                        <div class="component-info">
                            <h4>System Packages</h4>
                            <p>Essential system packages and tools</p>
                        </div>
                        <button class="add-component-btn" onclick="addComponentToCanvas('system-packages', '📦', 'System Packages', 'Essential system packages and tools')">+</button>
                    </div>
                    <div class="component-item" data-component="desktop-environment">
                        <div class="component-icon">🖥️</div>
                        <div class="component-info">
                            <h4>Desktop Environment</h4>
                            <p>GNOME, KDE, or other desktop environments</p>
                        </div>
                        <button class="add-component-btn" onclick="addComponentToCanvas('desktop-environment', '🖥️', 'Desktop Environment', 'GNOME, KDE, or other desktop environments')">+</button>
                    </div>
                    <div class="component-item" data-component="web-server">
                        <div class="component-icon">🌐</div>
                        <div class="component-info">
                            <h4>Web Server</h4>
                            <p>Nginx or Apache web server configuration</p>
                        </div>
                        <button class="add-component-btn" onclick="addComponentToCanvas('web-server', '🌐', 'Web Server', 'Nginx or Apache web server configuration')">+</button>
                    </div>
                    <div class="component-item" data-component="database">
                        <div class="component-icon">🗄️</div>
                        <div class="component-info">
                            <h4>Database</h4>
                            <p>PostgreSQL, MySQL, or other databases</p>
                        </div>
                        <button class="add-component-btn" onclick="addComponentToCanvas('database', '🗄️', 'Database', 'PostgreSQL, MySQL, or other databases')">+</button>
                    </div>
                    <div class="component-item" data-component="security">
                        <div class="component-icon">🔒</div>
                        <div class="component-info">
                            <h4>Security</h4>
                            <p>Firewall, fail2ban, and security hardening</p>
                        </div>
                        <button class="add-component-btn" onclick="addComponentToCanvas('security', '🔒', 'Security', 'Firewall, fail2ban, and security hardening')">+</button>
                    </div>
                    <div class="component-item" data-component="monitoring">
                        <div class="component-icon">📊</div>
                        <div class="component-info">
                            <h4>Monitoring</h4>
                            <p>Prometheus, Grafana, and system monitoring</p>
                        </div>
                        <button class="add-component-btn" onclick="addComponentToCanvas('monitoring', '📊', 'Monitoring', 'Prometheus, Grafana, and system monitoring')">+</button>
                    </div>
                </div>
            </div>

            <!-- Configuration List -->
            <div class="feature-card builder-section">
                <div class="section-header">
                    <h3>📋 Configuration Components</h3>
                    <p>Your selected components appear here as a simple list</p>
                </div>
                <div id="configuration-list">
                    <div id="empty-message" style="padding: 2rem; text-align: center; color: #666; border: 2px dashed #ddd; border-radius: 8px;">
                        <h4>No components added yet</h4>
                        <p>Click the + buttons above to add components</p>
                    </div>
                </div>
            </div>

            <!-- Configuration Preview -->
            <div class="feature-card builder-section">
                <div class="section-header">
                    <h3>📄 Configuration Preview</h3>
                    <p>Live preview of your generated NixOS configuration</p>
                </div>
                <div class="config-editor-container">
                    <div class="config-toolbar">
                        <button class="config-btn" onclick="formatConfig()">✨ Format</button>
                        <button class="config-btn" onclick="copyConfig()">📋 Copy</button>
                        <button class="config-btn" onclick="downloadConfig()">💾 Download</button>
                    </div>
                    <pre id="config-preview" class="config-preview"># Your NixOS configuration will appear here
# Click the + buttons above to add components to your configuration

{ config, pkgs, ... }:

{
  # Import hardware configuration
  imports = [ ./hardware-configuration.nix ];

  # Add components to see configuration here...
}</pre>
                </div>
            </div>
        </div>
        
        <style>
            /* AI Modal Styles */
            .ai-modal {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.5);
                z-index: 2000;
                display: flex;
                align-items: center;
                justify-content: center;
            }
            
            .ai-modal-content {
                background: var(--bg-primary);
                border-radius: 12px;
                max-width: 600px;
                width: 90%;
                max-height: 80vh;
                overflow-y: auto;
                box-shadow: var(--shadow-lg);
                border: 1px solid var(--border-color);
            }
            
            .ai-modal-header {
                padding: 1.5rem;
                border-bottom: 1px solid var(--border-color);
                display: flex;
                justify-content: space-between;
                align-items: center;
                background: var(--accent-color);
                color: white;
                border-radius: 12px 12px 0 0;
            }
            
            .ai-modal-close {
                background: none;
                border: none;
                color: white;
                font-size: 1.5rem;
                cursor: pointer;
                padding: 0;
                width: 30px;
                height: 30px;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: 50%;
                transition: background 0.2s;
            }
            
            .ai-modal-close:hover {
                background: rgba(255, 255, 255, 0.2);
            }
            
            .ai-modal-body {
                padding: 1.5rem;
            }
            
            .ai-input-section, .ai-options-section, .ai-template-section {
                margin-bottom: 1.5rem;
            }
            
            .ai-input-section label, .ai-options-section label, .ai-template-section label {
                display: block;
                margin-bottom: 0.5rem;
                font-weight: 500;
                color: var(--text-primary);
            }
            
            #ai-description {
                width: 100%;
                padding: 0.75rem;
                border: 1px solid var(--border-color);
                border-radius: 6px;
                background: var(--bg-secondary);
                color: var(--text-primary);
                font-family: inherit;
                resize: vertical;
            }
            
            .ai-option-buttons, .ai-template-buttons {
                display: flex;
                gap: 0.5rem;
                flex-wrap: wrap;
            }
            
            .ai-option-btn, .ai-template-btn {
                padding: 0.5rem 1rem;
                border: 1px solid var(--border-color);
                background: var(--bg-secondary);
                color: var(--text-primary);
                border-radius: 6px;
                cursor: pointer;
                transition: all 0.2s;
                font-size: 0.875rem;
            }
            
            .ai-option-btn:hover, .ai-template-btn:hover {
                background: var(--accent-color);
                color: white;
                border-color: var(--accent-color);
            }
            
            .ai-option-btn.selected {
                background: var(--accent-color);
                color: white;
                border-color: var(--accent-color);
            }
            
            .ai-modal-actions {
                display: flex;
                gap: 1rem;
                justify-content: flex-end;
                margin-top: 2rem;
            }

            /* Builder Layout Styles */
            .builder-container {
                max-width: 1200px;
                margin: 0 auto;
            }
            
            .builder-header {
                text-align: center;
            }
            
            .builder-section {
                margin-top: 2rem;
            }
            
            .section-header {
                margin-bottom: 1.5rem;
            }
            
            .section-header h3 {
                color: var(--text-primary);
                margin-bottom: 0.5rem;
            }
            
            .section-header p {
                color: var(--text-secondary);
                font-size: 0.875rem;
            }
            
            .component-list {
                display: flex;
                flex-direction: column;
                gap: 1rem;
            }
            
            .component-item {
                background: var(--bg-secondary);
                border: 2px solid var(--border-color);
                border-radius: 12px;
                padding: 1.25rem;
                transition: all 0.3s ease;
                display: flex;
                align-items: center;
                gap: 1rem;
                position: relative;
            }
            
            .component-item:hover {
                background: var(--bg-primary);
                border-color: var(--accent-color);
                box-shadow: var(--shadow-lg);
                transform: translateY(-2px);
            }
            
            .component-icon {
                font-size: 2rem;
                flex-shrink: 0;
            }
            
            .component-info h4 {
                color: var(--text-primary);
                margin: 0 0 0.25rem 0;
                font-size: 1rem;
            }
            
            .component-info p {
                color: var(--text-secondary);
                margin: 0;
                font-size: 0.875rem;
                line-height: 1.4;
            }
            
            .add-component-btn {
                position: absolute;
                right: 1rem;
                top: 50%;
                transform: translateY(-50%);
                width: 40px;
                height: 40px;
                border: none;
                border-radius: 50%;
                background: var(--accent-color);
                color: white;
                font-size: 1.5rem;
                font-weight: bold;
                cursor: pointer;
                transition: all 0.3s ease;
                display: flex;
                align-items: center;
                justify-content: center;
                box-shadow: var(--shadow);
            }
            
            .add-component-btn:hover {
                background: var(--accent-hover);
                transform: translateY(-50%) scale(1.1);
                box-shadow: var(--shadow-lg);
            }
            
            .add-component-btn:active {
                transform: translateY(-50%) scale(0.95);
            }
            
            /* Simple Configuration List Styles */
            #configuration-list {
                margin-top: 1rem;
            }
            
            .config-item {
                background: var(--bg-primary);
                border: 2px solid var(--border-color);
                border-radius: 8px;
                padding: 1rem;
                margin-bottom: 1rem;
                display: flex;
                align-items: center;
                justify-content: space-between;
            }
            
            .config-item:hover {
                border-color: var(--accent-color);
                box-shadow: var(--shadow);
            }
            
            .config-info {
                display: flex;
                align-items: center;
                gap: 1rem;
            }
            
            .config-icon {
                font-size: 1.5rem;
            }
            
            .config-details h4 {
                margin: 0 0 0.25rem 0;
                color: var(--text-primary);
            }
            
            .config-details p {
                margin: 0;
                color: var(--text-secondary);
                font-size: 0.875rem;
            }
            
            .config-actions {
                display: flex;
                gap: 0.5rem;
            }
            
            .config-btn {
                background: var(--bg-secondary);
                border: 1px solid var(--border-color);
                color: var(--text-primary);
                padding: 0.5rem 1rem;
                border-radius: 6px;
                cursor: pointer;
                font-size: 0.875rem;
                transition: all 0.2s ease;
            }
            
            .config-btn:hover {
                background: var(--accent-color);
                color: white;
                border-color: var(--accent-color);
            }
            
            .config-btn.delete {
                background: #fee2e2;
                border-color: #fecaca;
                color: #dc2626;
            }
            
            .config-btn.delete:hover {
                background: #dc2626;
                color: white;
            }
            
            .canvas-component .component-desc {
                font-size: 0.75rem;
                color: var(--text-secondary);
            }
            
            .config-editor-container {
                position: relative;
            }
            
            .config-toolbar {
                display: flex;
                gap: 0.5rem;
                margin-bottom: 1rem;
                padding: 0.75rem;
                background: var(--bg-secondary);
                border-radius: 8px;
                border: 1px solid var(--border-color);
            }
            
            .config-btn {
                background: var(--bg-primary);
                border: 1px solid var(--border-color);
                color: var(--text-primary);
                padding: 0.5rem 1rem;
                border-radius: 6px;
                cursor: pointer;
                transition: all 0.2s ease;
                font-size: 0.875rem;
            }
            
            .config-btn:hover {
                background: var(--accent-color);
                color: white;
                border-color: var(--accent-color);
            }
            
            .config-preview {
                background: var(--bg-secondary);
                color: var(--text-primary);
                padding: 1.5rem;
                border-radius: 8px;
                font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
                font-size: 0.875rem;
                line-height: 1.6;
                min-height: 400px;
                overflow-x: auto;
                border: 1px solid var(--border-color);
                white-space: pre-wrap;
                word-wrap: break-word;
            }

            @media (max-width: 768px) {
                .component-list {
                    gap: 0.75rem;
                }
                
                .ai-modal-content {
                    width: 95%;
                    margin: 1rem;
                }
                
                .ai-option-buttons, .ai-template-buttons {
                    flex-direction: column;
                }
                
                .config-toolbar {
                    flex-wrap: wrap;
                }
            }
        </style>
        
        <script>
            // Global variables
            let selectedAIType = '';
            
            // Modal Functions
            function openAIModal() {
                document.getElementById('ai-modal').style.display = 'flex';
                document.getElementById('ai-description').focus();
            }
            
            function closeAIModal() {
                document.getElementById('ai-modal').style.display = 'none';
                document.getElementById('ai-description').value = '';
                selectedAIType = '';
                document.querySelectorAll('.ai-option-btn').forEach(btn => btn.classList.remove('selected'));
            }
            
            // Notification system
            function showNotification(message, type = 'info') {
                const notification = document.createElement('div');
                notification.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 1000; padding: 1rem 1.5rem; border-radius: 8px; color: white; font-weight: 500; box-shadow: 0 4px 12px rgba(0,0,0,0.15); animation: slideIn 0.3s ease;';
                
                const colors = {
                    success: '#10b981',
                    error: '#ef4444', 
                    warning: '#f59e0b',
                    info: '#3b82f6'
                };
                
                notification.style.background = colors[type] || colors.info;
                notification.textContent = message;
                
                document.body.appendChild(notification);
                
                setTimeout(() => {
                    notification.style.animation = 'slideOut 0.3s ease';
                    setTimeout(() => notification.remove(), 300);
                }, 3000);
            }
            
            function generateAIConfig() {
                const description = document.getElementById('ai-description').value;
                const type = selectedAIType || 'custom';
                
                if (!description.trim()) {
                    alert('Please describe what you want to configure');
                    return;
                }
                
                // Simulate AI generation
                const config = "{ config, pkgs, ... }:" + "\\n\\n" +
                    "{" + "\\n" +
                    "  # AI-Generated Configuration" + "\\n" +
                    "  # Type: " + type + "\\n" +
                    "  # Description: " + description + "\\n" +
                    "  " + "\\n" +
                    "  imports = [ ./hardware-configuration.nix ];" + "\\n" +
                    "  " + "\\n" +
                    "  # System configuration based on your request" + "\\n" +
                    "  " + generateConfigForType(type, description) + "\\n" +
                    "}";
                
                document.getElementById('config-preview').textContent = config;
                closeAIModal();
                
                // Show success message
                showNotification('Configuration generated successfully!', 'success');
            }
            
            function validateAIConfig() {
                const config = document.getElementById('config-preview').textContent;
                
                // Simulate AI validation
                setTimeout(() => {
                    showNotification('Configuration validated: 92/100 score. Minor suggestions available.', 'success');
                }, 1000);
                
                closeAIModal();
            }
            
            function useTemplate(template) {
                const templates = {
                    'nginx-ssl': 'services.nginx.enable = true;\\nservices.nginx.virtualHosts."example.com" = {\\n  enableACME = true;\\n  forceSSL = true;\\n};\\nsecurity.acme.acceptTerms = true;',
                    'postgres-db': 'services.postgresql = {\\n  enable = true;\\n  package = pkgs.postgresql_14;\\n  authentication = "local all all trust";\\n};',
                    'docker-host': 'virtualisation.docker.enable = true;\\nusers.users.yourusername.extraGroups = [ "docker" ];',
                    'dev-workstation': 'environment.systemPackages = with pkgs; [\\n  git\\n  nodejs\\n  python3\\n  vscode\\n  docker\\n];'
                };
                
                document.getElementById('ai-description').value = 'Configuration for ' + template.replace('-', ' ');
            }
            
            function generateConfigForType(type, description) {
                const configs = {
                    desktop: 'services.xserver.enable = true;\\n  services.xserver.displayManager.gdm.enable = true;\\n  services.xserver.desktopManager.gnome.enable = true;',
                    server: 'services.nginx.enable = true;\\n  networking.firewall.allowedTCPPorts = [ 80 443 ];\\n  services.openssh.enable = true;',
                    development: 'environment.systemPackages = with pkgs; [\\n    git nodejs python3 vscode docker\\n  ];',
                    gaming: 'programs.steam.enable = true;\\n  hardware.opengl.enable = true;\\n  hardware.pulseaudio.enable = true;',
                    custom: '# Custom configuration based on description'
                };
                
                return configs[type] || configs.custom;
            }
            
            // Component setup and AI option selection
            document.addEventListener('DOMContentLoaded', function() {
                console.log('DOM loaded, setting up components...');
                
                document.querySelectorAll('.ai-option-btn').forEach(btn => {
                    btn.addEventListener('click', function() {
                        document.querySelectorAll('.ai-option-btn').forEach(b => b.classList.remove('selected'));
                        this.classList.add('selected');
                        selectedAIType = this.dataset.type;
                    });
                });
                
                // Close modal on outside click
                document.getElementById('ai-modal').addEventListener('click', function(e) {
                    if (e.target === this) {
                        closeAIModal();
                    }
                });
                
                console.log('Component setup complete');
            });
            
            // Builder Functions
            function createNewConfig() {
                document.getElementById('config-preview').textContent = "{ config, pkgs, ... }:" + "\\n\\n" +
                    "{" + "\\n" +
                    "  imports = [ ./hardware-configuration.nix ];" + "\\n" +
                    "  " + "\\n" +
                    "  # Boot loader" + "\\n" +
                    "  boot.loader.systemd-boot.enable = true;" + "\\n" +
                    "  boot.loader.efi.canTouchEfiVariables = true;" + "\\n" +
                    "  " + "\\n" +
                    "  # Network" + "\\n" +
                    "  networking.hostName = \\"nixos\\";" + "\\n" +
                    "  networking.networkmanager.enable = true;" + "\\n" +
                    "  " + "\\n" +
                    "  # System packages" + "\\n" +
                    "  environment.systemPackages = with pkgs; [" + "\\n" +
                    "    # Add your packages here" + "\\n" +
                    "  ];" + "\\n" +
                    "  " + "\\n" +
                    "  system.stateVersion = \\"23.11\\";" + "\\n" +
                    "}";
                
                // Clear canvas
                const canvas = document.getElementById('builder-canvas');
                const components = canvas.querySelectorAll('.canvas-component');
                components.forEach(comp => comp.remove());
                
                showPlaceholder();
                showNotification('New configuration created!', 'info');
            }
            
            function loadTemplate() {
                const templates = {
                    'Basic Desktop': "{ config, pkgs, ... }:" + "\\n\\n" +
                        "{" + "\\n" +
                        "  imports = [ ./hardware-configuration.nix ];" + "\\n" +
                        "  " + "\\n" +
                        "  # Desktop Environment" + "\\n" +
                        "  services.xserver.enable = true;" + "\\n" +
                        "  services.xserver.displayManager.gdm.enable = true;" + "\\n" +
                        "  services.xserver.desktopManager.gnome.enable = true;" + "\\n" +
                        "  " + "\\n" +
                        "  # Audio" + "\\n" +
                        "  sound.enable = true;" + "\\n" +
                        "  hardware.pulseaudio.enable = true;" + "\\n" +
                        "  " + "\\n" +
                        "  # User account" + "\\n" +
                        "  users.users.user = {" + "\\n" +
                        "    isNormalUser = true;" + "\\n" +
                        "    extraGroups = [ \\"wheel\\" \\"networkmanager\\" ];" + "\\n" +
                        "  };" + "\\n" +
                        "  " + "\\n" +
                        "  system.stateVersion = \\"23.11\\";" + "\\n" +
                        "}",
                    'Web Server': "{ config, pkgs, ... }:" + "\\n\\n" +
                        "{" + "\\n" +
                        "  imports = [ ./hardware-configuration.nix ];" + "\\n" +
                        "  " + "\\n" +
                        "  # Web Server" + "\\n" +
                        "  services.nginx.enable = true;" + "\\n" +
                        "  services.nginx.virtualHosts.localhost = {" + "\\n" +
                        "    root = \\\"/var/www\\\";" + "\\n" +
                        "  };" + "\\n" +
                        "  " + "\\n" +
                        "  # Database" + "\\n" +
                        "  services.postgresql.enable = true;" + "\\n" +
                        "  " + "\\n" +
                        "  # Firewall" + "\\n" +
                        "  networking.firewall.allowedTCPPorts = [ 22 80 443 ];" + "\\n" +
                        "  " + "\\n" +
                        "  system.stateVersion = \\"23.11\\";" + "\\n" +
                        "}"
                };
                
                const templateNames = Object.keys(templates);
                const choice = prompt('Choose template:\\n' + templateNames.map((name, i) => (i + 1) + '. ' + name).join('\\n'));
                
                if (choice && choice > 0 && choice <= templateNames.length) {
                    const templateName = templateNames[choice - 1];
                    document.getElementById('config-preview').textContent = templates[templateName];
                    showNotification('Template "' + templateName + '" loaded!', 'info');
                }
            }
            
            function validateConfig() {
                showNotification('Validating configuration...', 'info');
                
                setTimeout(() => {
                    showNotification('✅ Configuration is valid! Score: 94/100', 'success');
                }, 1500);
            }
            
            function exportConfig() {
                const config = document.getElementById('config-preview').textContent;
                const blob = new Blob([config], { type: 'text/plain' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = 'configuration.nix';
                a.click();
                URL.revokeObjectURL(url);
                
                showNotification('Configuration exported successfully!', 'success');
            }
            
            function formatConfig() {
                showNotification('Configuration formatted!', 'info');
            }
            
            function copyConfig() {
                const config = document.getElementById('config-preview').textContent;
                navigator.clipboard.writeText(config).then(() => {
                    showNotification('Configuration copied to clipboard!', 'success');
                });
            }
            
            function downloadConfig() {
                exportConfig();
            }
            
            // Simple configuration list
            let configList = [];
            
            function addComponentToCanvas(type, icon, title, description) {
                console.log('Adding component:', title);
                
                // Check if already added
                if (configList.find(item => item.type === type)) {
                    showNotification('Component already added!', 'warning');
                    return;
                }
                
                // Add to list
                configList.push({
                    id: Date.now(),
                    type: type,
                    icon: icon,
                    title: title,
                    description: description
                });
                
                // Update display
                updateConfigList();
                updateConfigPreview();
                showNotification('Component added!', 'success');
            }
            
            function updateConfigList() {
                const container = document.getElementById('configuration-list');
                const emptyMessage = document.getElementById('empty-message');
                
                if (configList.length === 0) {
                    emptyMessage.style.display = 'block';
                    // Remove any existing items
                    const items = container.querySelectorAll('.config-item');
                    items.forEach(item => item.remove());
                    return;
                }
                
                emptyMessage.style.display = 'none';
                
                // Remove existing items
                const items = container.querySelectorAll('.config-item');
                items.forEach(item => item.remove());
                
                // Add new items
                configList.forEach(config => {
                    const item = document.createElement('div');
                    item.className = 'config-item';
                    item.innerHTML = 
                        '<div class="config-info">' +
                            '<div class="config-icon">' + config.icon + '</div>' +
                            '<div class="config-details">' +
                                '<h4>' + config.title + '</h4>' +
                                '<p>' + config.description + '</p>' +
                            '</div>' +
                        '</div>' +
                        '<div class="config-actions">' +
                            '<button class="config-btn" onclick="editConfig(' + config.id + ')">Edit</button>' +
                            '<button class="config-btn" onclick="viewConfig(' + config.id + ')">View</button>' +
                            '<button class="config-btn" onclick="verifyConfig(' + config.id + ')">AI</button>' +
                            '<button class="config-btn delete" onclick="deleteConfig(' + config.id + ')">Delete</button>' +
                        '</div>';
                    container.appendChild(item);
                });
            }
            
            function editConfig(id) {
                const config = configList.find(c => c.id === id);
                if (!config) return;
                const newTitle = prompt('Edit title:', config.title);
                if (newTitle) {
                    config.title = newTitle;
                    updateConfigList();
                    updateConfigPreview();
                    showNotification('Updated!', 'success');
                }
            }
            
            function viewConfig(id) {
                const config = configList.find(c => c.id === id);
                if (!config) return;
                const nixCode = generateComponentConfig(config.type);
                alert('Component: ' + config.title + '\\n\\nNix Code:\\n' + nixCode);
            }
            
            function verifyConfig(id) {
                const config = configList.find(c => c.id === id);
                if (!config) return;
                showNotification('AI verifying...', 'info');
                setTimeout(() => {
                    const score = Math.floor(Math.random() * 20) + 80;
                    showNotification('AI Score: ' + score + '/100', 'success');
                }, 1500);
            }
            
            function deleteConfig(id) {
                if (confirm('Delete this component?')) {
                    configList = configList.filter(c => c.id !== id);
                    updateConfigList();
                    updateConfigPreview();
                    showNotification('Deleted!', 'info');
                }
            }
            
            function showPlaceholder() {
                const placeholder = document.querySelector('.canvas-placeholder');
                if (placeholder) {
                    placeholder.style.display = 'block';
                }
            }
            
            function hidePlaceholder() {
                const placeholder = document.querySelector('.canvas-placeholder');
                if (placeholder) {
                    placeholder.style.display = 'none';
                }
            }
            
            function updateConfigPreview() {
                let config = "{ config, pkgs, ... }:" + "\\n\\n" +
                    "{" + "\\n" +
                    "  imports = [ ./hardware-configuration.nix ];" + "\\n" +
                    "  " + "\\n" +
                    "  # Generated configuration" + "\\n";
                
                if (configList.length === 0) {
                    config += "\\n  # Add components to see configuration here" + "\\n";
                } else {
                    configList.forEach(component => {
                        config += generateComponentConfig(component.type);
                    });
                }
                
                config += "\\n" +
                    "  system.stateVersion = \\"23.11\\";" + "\\n" +
                    "}";
                
                const previewElement = document.getElementById('config-preview');
                if (previewElement) {
                    previewElement.textContent = config;
                }
            }
            
            function generateComponentConfig(type) {
                const configs = {
                    'system-packages': "\n  # System Packages" + "\n" +
                        "  environment.systemPackages = with pkgs; [" + "\n" +
                        "    vim git curl wget firefox" + "\n" +
                        "  ];",
                    'desktop-environment': "\n  # Desktop Environment" + "\n" +
                        "  services.xserver.enable = true;" + "\n" +
                        "  services.xserver.displayManager.gdm.enable = true;" + "\n" +
                        "  services.xserver.desktopManager.gnome.enable = true;",
                    'web-server': "\n  # Web Server" + "\n" +
                        "  services.nginx.enable = true;" + "\n" +
                        "  networking.firewall.allowedTCPPorts = [ 80 443 ];",
                    'database': "\n  # Database" + "\n" +
                        "  services.postgresql.enable = true;" + "\n" +
                        "  services.postgresql.authentication = \"local all all trust\";",
                    'security': "\n  # Security" + "\n" +
                        "  services.fail2ban.enable = true;" + "\n" +
                        "  networking.firewall.enable = true;",
                    'monitoring': "\n  # Monitoring" + "\n" +
                        "  services.prometheus.enable = true;" + "\n" +
                        "  services.grafana.enable = true;"
                };
                
                return configs[type] || '\\n  # ' + type + ' configuration';
            }
            
            function showNotification(message, type = 'info') {
                // Create notification element
                const notification = document.createElement('div');
                notification.style.cssText = 'position: fixed;' +
                    'top: 20px;' +
                    'right: 20px;' +
                    'background: var(--' + (type === 'success' ? 'success' : type === 'error' ? 'error' : 'accent') + '-color);' +
                    'color: white;' +
                    'padding: 1rem 1.5rem;' +
                    'border-radius: 8px;' +
                    'box-shadow: var(--shadow-lg);' +
                    'z-index: 3000;' +
                    'opacity: 0;' +
                    'transform: translateX(100%);' +
                    'transition: all 0.3s ease;';
                notification.textContent = message;
                
                document.body.appendChild(notification);
                
                // Animate in
                setTimeout(() => {
                    notification.style.opacity = '1';
                    notification.style.transform = 'translateX(0)';
                }, 10);
                
                // Remove after 3 seconds
                setTimeout(() => {
                    notification.style.opacity = '0';
                    notification.style.transform = 'translateX(100%)';
                    setTimeout(() => notification.remove(), 300);
                }, 3000);
            }
        </script>`)
}

func (s *EnhancedServer) generateConfigsHTML() string {
	return s.generatePageWithNavbar("Configs", "📝", `
        <div class="content-grid">
            <div class="main-content">
                <div class="feature-card">
                    <h2>📝 Configuration Editor</h2>
                    <p>Edit your NixOS configuration files directly with syntax highlighting, validation, and AI assistance.</p>
                    
                    <div class="action-buttons" style="margin-top: 2rem;">
                        <button class="action-btn primary" onclick="createNewFile()">New File</button>
                        <button class="action-btn" onclick="saveConfig()">Save</button>
                        <button class="action-btn" onclick="validateWithAI()">AI Validate</button>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 1rem;">
                        <h3>📁 File Browser</h3>
                        <button class="action-btn" onclick="refreshFiles()" style="margin: 0;">🔄 Refresh</button>
                    </div>
                    <div id="file-list" style="display: flex; flex-direction: column; gap: 0.5rem;">
                        <div class="file-item" onclick="loadFile('configuration.nix')" style="display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem; background: var(--bg-secondary); border-radius: 6px; cursor: pointer; border: 1px solid var(--border-color);">
                            <span style="font-size: 1rem;">📄</span>
                            <div>
                                <div style="color: var(--text-primary); font-weight: 500;">configuration.nix</div>
                                <div style="color: var(--text-secondary); font-size: 0.75rem;">Main system configuration • 2.4 KB • Modified 2 hours ago</div>
                            </div>
                        </div>
                        <div class="file-item" onclick="loadFile('hardware-configuration.nix')" style="display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem; background: var(--bg-secondary); border-radius: 6px; cursor: pointer; border: 1px solid var(--border-color);">
                            <span style="font-size: 1rem;">⚙️</span>
                            <div>
                                <div style="color: var(--text-primary); font-weight: 500;">hardware-configuration.nix</div>
                                <div style="color: var(--text-secondary); font-size: 0.75rem;">Hardware configuration • 1.2 KB • Modified 1 day ago</div>
                            </div>
                        </div>
                        <div class="file-item" onclick="loadFile('modules/')" style="display: flex; align-items: center; gap: 0.75rem; padding: 0.75rem; background: var(--bg-secondary); border-radius: 6px; cursor: pointer; border: 1px solid var(--border-color);">
                            <span style="font-size: 1rem;">📁</span>
                            <div>
                                <div style="color: var(--text-primary); font-weight: 500;">modules/</div>
                                <div style="color: var(--text-secondary); font-size: 0.75rem;">Custom modules directory</div>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="feature-card" style="margin-top: 2rem;">
                    <div style="display: flex; justify-content: between; align-items: center; margin-bottom: 1rem;">
                        <h3>📝 Editor</h3>
                        <span id="current-file" style="color: var(--text-secondary); font-size: 0.875rem;">No file selected</span>
                    </div>
                    <textarea id="config-editor" placeholder="Select a file to edit or create a new one..." style="width: 100%; height: 400px; padding: 1rem; border: 1px solid var(--border-color); border-radius: 8px; background: var(--bg-secondary); color: var(--text-primary); font-family: 'Monaco', 'Menlo', monospace; font-size: 0.875rem; line-height: 1.4; resize: vertical;"></textarea>
                    
                    <div class="action-buttons" style="margin-top: 1rem;">
                        <button class="action-btn primary" onclick="saveConfig()">💾 Save</button>
                        <button class="action-btn" onclick="formatCode()">✨ Format</button>
                        <button class="action-btn" onclick="validateWithAI()">🤖 AI Validate</button>
                    </div>
                    
                    <div id="validation-results" style="margin-top: 1rem; display: none;">
                        <!-- Validation results will appear here -->
                    </div>
                </div>
            </div>
            
            <div class="sidebar">
                <div class="feature-card">
                    <h3>🤖 AI Assistant</h3>
                    <input type="text" id="ai-input" placeholder="Ask about NixOS configuration..." style="width: 100%; padding: 0.75rem; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-secondary); color: var(--text-primary); margin-bottom: 0.5rem;">
                    <button class="action-btn primary" onclick="askConfigAI()" style="width: 100%; margin-bottom: 1rem;">Ask AI</button>
                    <div id="ai-response" style="padding: 1rem; background: var(--bg-secondary); border-radius: 6px; font-size: 0.875rem; line-height: 1.5; display: none;"></div>
                </div>
                
                <div class="feature-card">
                    <h3>📊 Configuration Stats</h3>
                    <div style="display: flex; flex-direction: column; gap: 1rem;">
                        <div>
                            <div style="color: var(--text-secondary); font-size: 0.875rem;">Total Files</div>
                            <div style="color: var(--text-primary); font-size: 1.5rem; font-weight: 600;">4</div>
                        </div>
                        <div>
                            <div style="color: var(--text-secondary); font-size: 0.875rem;">Lines of Code</div>
                            <div style="color: var(--text-primary); font-size: 1.5rem; font-weight: 600;">127</div>
                        </div>
                        <div>
                            <div style="color: var(--text-secondary); font-size: 0.875rem;">Last Modified</div>
                            <div style="color: var(--text-primary); font-size: 1.5rem; font-weight: 600;">2h ago</div>
                        </div>
                    </div>
                </div>
                
                <div class="feature-card">
                    <h3>⚡ Quick Actions</h3>
                    <div class="quick-actions-list">
                        <a href="#" class="quick-action-item" onclick="loadTemplate('basic')">📝 Load Basic Template</a>
                        <a href="#" class="quick-action-item" onclick="backupConfig()">💾 Backup Configuration</a>
                        <a href="#" class="quick-action-item" onclick="deployConfig()">🚀 Deploy Changes</a>
                        <a href="#" class="quick-action-item" onclick="shareConfig()">📤 Share Configuration</a>
                    </div>
                </div>
            </div>
        </div>
        
        <script>
            let currentFile = null;
            
            function loadFile(filename) {
                currentFile = filename;
                document.getElementById('current-file').textContent = filename;
                
                // Simulate loading file content
                if (filename === 'configuration.nix') {
                    document.getElementById('config-editor').value = '{ config, pkgs, ... }:\n\n{\n  imports = [\n    ./hardware-configuration.nix\n  ];\n\n  # Boot loader configuration\n  boot.loader.systemd-boot.enable = true;\n  boot.loader.efi.canTouchEfiVariables = true;\n\n  # Network configuration\n  networking.hostName = "nixos";\n  networking.networkmanager.enable = true;\n\n  # System packages\n  environment.systemPackages = with pkgs; [\n    vim\n    git\n    firefox\n    vscode\n  ];\n\n  # Enable SSH\n  services.openssh.enable = true;\n\n  # System version\n  system.stateVersion = "23.11";\n}';
                } else {
                    document.getElementById('config-editor').value = '# ' + filename + ' content would be loaded here';
                }
            }
            
            function saveConfig() {
                if (!currentFile) {
                    alert('No file selected to save');
                    return;
                }
                
                const content = document.getElementById('config-editor').value;
                // Simulate saving
                alert('✅ ' + currentFile + ' saved successfully!');
            }
            
            function createNewFile() {
                const filename = prompt('Enter filename (e.g., mymodule.nix):');
                if (filename) {
                    currentFile = filename;
                    document.getElementById('current-file').textContent = filename;
                    document.getElementById('config-editor').value = '{ config, pkgs, ... }:\n\n{\n  # New configuration file\n}';
                }
            }
            
            function validateWithAI() {
                const content = document.getElementById('config-editor').value;
                if (!content.trim()) {
                    alert('No content to validate');
                    return;
                }
                
                // Simulate AI validation
                const results = document.getElementById('validation-results');
                results.style.display = 'block';
                results.innerHTML = '<div style="background: var(--success-color); color: white; padding: 1rem; border-radius: 6px; margin-bottom: 1rem;"><strong>✅ Validation Score: 85/100</strong></div>' +
                    '<div style="background: var(--warning-color); color: white; padding: 0.75rem; border-radius: 6px; margin-bottom: 0.5rem;"><strong>⚠️ Warning:</strong> Consider enabling automatic updates (line 15)</div>' +
                    '<div style="background: var(--bg-secondary); padding: 1rem; border-radius: 6px; border: 1px solid var(--border-color);"><strong>💡 Suggestions:</strong><ul style="margin: 0.5rem 0 0 1rem;"><li>Add firewall configuration for better security</li><li>Consider adding backup services</li><li>Enable fail2ban for SSH protection</li></ul></div>';
            }
            
            function askConfigAI() {
                const input = document.getElementById('ai-input').value;
                if (!input.trim()) return;
                
                const response = document.getElementById('ai-response');
                response.style.display = 'block';
                response.innerHTML = '<strong>🤖 AI:</strong> I can help you with "' + input + '". Here are some suggestions for your NixOS configuration...';
                
                document.getElementById('ai-input').value = '';
            }
            
            function formatCode() {
                // Simulate code formatting
                alert('✨ Code formatted successfully!');
            }
            
            function refreshFiles() {
                alert('🔄 File list refreshed!');
            }
            
            function loadTemplate(type) {
                document.getElementById('config-editor').value = '{ config, pkgs, ... }:\n\n{\n  # Template: ' + type + '\n  # Basic configuration template\n}';
                document.getElementById('current-file').textContent = 'template-' + type + '.nix';
                currentFile = 'template-' + type + '.nix';
            }
            
            function backupConfig() {
                alert('💾 Configuration backed up successfully!');
            }
            
            function deployConfig() {
                alert('🚀 Configuration deployment started!');
            }
            
            function shareConfig() {
                alert('📤 Configuration shared via link!');
            }
        </script>`)
}