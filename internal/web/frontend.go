package web

import (
	"fmt"
	"net/http"
)

// Frontend View Handlers

func (s *EnhancedServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "dashboard.html", map[string]interface{}{
		"Title":    "nixai Dashboard",
		"Features": s.config.Features,
		"User": map[string]string{
			"Name": "Demo User",
			"Role": "Admin",
		},
	})
}

func (s *EnhancedServer) handleConfigBuilder(w http.ResponseWriter, r *http.Request) {
	if !s.config.Features.VisualBuilder {
		s.sendError(w, "Visual builder is disabled", http.StatusServiceUnavailable)
		return
	}

	s.renderTemplate(w, "builder.html", map[string]interface{}{
		"Title":    "Configuration Builder",
		"Features": s.config.Features,
	})
}

func (s *EnhancedServer) handleFleetView(w http.ResponseWriter, r *http.Request) {
	if !s.config.Features.FleetManagement {
		s.sendError(w, "Fleet management is disabled", http.StatusServiceUnavailable)
		return
	}

	s.renderTemplate(w, "fleet.html", map[string]interface{}{
		"Title":    "Fleet Management",
		"Features": s.config.Features,
	})
}

func (s *EnhancedServer) handleTeamsView(w http.ResponseWriter, r *http.Request) {
	if !s.config.Features.Collaboration {
		s.sendError(w, "Team collaboration is disabled", http.StatusServiceUnavailable)
		return
	}

	s.renderTemplate(w, "teams.html", map[string]interface{}{
		"Title":    "Team Collaboration",
		"Features": s.config.Features,
	})
}

func (s *EnhancedServer) handleVersionsView(w http.ResponseWriter, r *http.Request) {
	if !s.config.Features.VersionControl {
		s.sendError(w, "Version control is disabled", http.StatusServiceUnavailable)
		return
	}

	s.renderTemplate(w, "versions.html", map[string]interface{}{
		"Title":    "Version Control",
		"Features": s.config.Features,
	})
}

func (s *EnhancedServer) handleSPA(w http.ResponseWriter, r *http.Request) {
	// Serve the main SPA for any unmatched routes
	s.renderTemplate(w, "app.html", map[string]interface{}{
		"Title":    "nixai Web Interface",
		"Features": s.config.Features,
		"Config": map[string]interface{}{
			"ApiBaseURL": "/api/v1",
			"WSProtocol": "ws",
			"Features":   s.config.Features,
		},
	})
}

// Template rendering moved to enhanced_server.go to avoid conflicts

// Fallback HTML for when templates are not available
func (s *EnhancedServer) renderFallbackHTML(w http.ResponseWriter, templateName string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Extract title from data
	title := "nixai Web Interface"
	if dataMap, ok := data.(map[string]interface{}); ok {
		if titleVal, exists := dataMap["Title"]; exists {
			if titleStr, ok := titleVal.(string); ok {
				title = titleStr
			}
		}
	}

	var html string

	switch templateName {
	case "dashboard.html":
		html = s.getDashboardHTML(title)
	case "builder.html":
		html = s.getBuilderHTML(title)
	case "fleet.html":
		html = s.getFleetHTML(title)
	case "teams.html":
		html = s.getTeamsHTML(title)
	case "versions.html":
		html = s.getVersionsHTML(title)
	default:
		html = s.getDefaultHTML(title)
	}

	w.Write([]byte(html))
}

func (s *EnhancedServer) getDashboardHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .nav { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .nav a { margin-right: 20px; text-decoration: none; color: #2563eb; font-weight: 500; }
        .nav a:hover { text-decoration: underline; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .card h3 { margin-top: 0; color: #1f2937; }
        .stat { display: flex; justify-content: space-between; align-items: center; margin: 10px 0; }
        .stat-value { font-size: 24px; font-weight: bold; color: #2563eb; }
        .status { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 500; }
        .status.healthy { background: #dcfce7; color: #166534; }
        .status.warning { background: #fef3c7; color: #92400e; }
        .loading { text-align: center; padding: 40px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🌐 %s</h1>
        <p>Collaborative NixOS Configuration Management</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/builder">Configuration Builder</a>
        <a href="/fleet">Fleet Management</a>
        <a href="/teams">Teams</a>
        <a href="/versions">Version Control</a>
    </div>
    
    <div class="grid">
        <div class="card">
            <h3>📊 System Overview</h3>
            <div id="overview-stats" class="loading">Loading...</div>
        </div>
        
        <div class="card">
            <h3>🚀 Fleet Status</h3>
            <div id="fleet-stats" class="loading">Loading...</div>
        </div>
        
        <div class="card">
            <h3>📝 Recent Activities</h3>
            <div id="activities" class="loading">Loading...</div>
        </div>
        
        <div class="card">
            <h3>⚠️ System Alerts</h3>
            <div id="alerts" class="loading">Loading...</div>
        </div>
    </div>
    
    <script>
        // Load dashboard data
        async function loadDashboard() {
            try {
                const overview = await fetch('/api/v1/dashboard/overview').then(r => r.json());
                if (overview.success) {
                    updateOverview(overview.data);
                }
                
                const stats = await fetch('/api/v1/dashboard/stats').then(r => r.json());
                if (stats.success) {
                    updateStats(stats.data);
                }
                
                const activities = await fetch('/api/v1/dashboard/activities').then(r => r.json());
                if (activities.success) {
                    updateActivities(activities.data);
                }
                
                const alerts = await fetch('/api/v1/dashboard/alerts').then(r => r.json());
                if (alerts.success) {
                    updateAlerts(alerts.data);
                }
            } catch (error) {
                console.error('Failed to load dashboard:', error);
            }
        }
        
        function updateOverview(data) {
            const html = Object.entries(data.summary).map(([key, value]) => 
                '<div class="stat"><span>' + formatKey(key) + '</span><span class="stat-value">' + value + '</span></div>'
            ).join('');
            document.getElementById('overview-stats').innerHTML = html;
        }
        
        function updateStats(data) {
            const machines = data.machines;
            const html = 
                '<div class="stat"><span>Total Machines</span><span class="stat-value">' + machines.total + '</span></div>' +
                '<div class="stat"><span>Healthy</span><span class="status healthy">' + machines.healthy + '</span></div>' +
                '<div class="stat"><span>Warning</span><span class="status warning">' + machines.warning + '</span></div>';
            document.getElementById('fleet-stats').innerHTML = html;
        }
        
        function updateActivities(data) {
            const html = data.slice(0, 5).map(activity => 
                '<div style="margin: 10px 0; padding: 10px; border-left: 3px solid #2563eb;">' +
                '<div style="font-weight: 500;">' + activity.message + '</div>' +
                '<div style="font-size: 12px; color: #6b7280;">' + activity.user + ' • ' + formatTime(activity.timestamp) + '</div>' +
                '</div>'
            ).join('');
            document.getElementById('activities').innerHTML = html || 'No recent activities';
        }
        
        function updateAlerts(data) {
            const html = data.slice(0, 3).map(alert => 
                '<div style="margin: 10px 0; padding: 10px; border-left: 3px solid #f59e0b;">' +
                '<div style="font-weight: 500;">' + alert.message + '</div>' +
                '<div style="font-size: 12px; color: #6b7280;">' + formatTime(alert.timestamp) + '</div>' +
                '</div>'
            ).join('');
            document.getElementById('alerts').innerHTML = html || 'No active alerts';
        }
        
        function formatKey(key) {
            return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
        }
        
        function formatTime(timestamp) {
            return new Date(timestamp).toLocaleString();
        }
        
        // Load data on page load
        loadDashboard();
        
        // Refresh every 30 seconds
        setInterval(loadDashboard, 30000);
    </script>
</body>
</html>`, title, title)
}

func (s *EnhancedServer) getBuilderHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
        .header { background: #2563eb; color: white; padding: 15px 20px; }
        .nav { background: white; padding: 10px 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .nav a { margin-right: 20px; text-decoration: none; color: #2563eb; font-weight: 500; }
        .builder { display: flex; height: calc(100vh - 120px); }
        .sidebar { width: 300px; background: white; border-right: 1px solid #e5e7eb; padding: 20px; overflow-y: auto; }
        .canvas { flex: 1; background: #fafafa; position: relative; overflow: auto; }
        .preview { width: 400px; background: white; border-left: 1px solid #e5e7eb; padding: 20px; overflow-y: auto; }
        .component { padding: 10px; margin: 5px 0; background: #f3f4f6; border-radius: 6px; cursor: pointer; }
        .component:hover { background: #e5e7eb; }
        .placeholder { text-align: center; color: #6b7280; padding: 40px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎨 %s</h1>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/builder">Configuration Builder</a>
        <a href="/fleet">Fleet Management</a>
        <a href="/teams">Teams</a>
        <a href="/versions">Version Control</a>
    </div>
    
    <div class="builder">
        <div class="sidebar">
            <h3>Components</h3>
            <div class="component">🌐 Network Services</div>
            <div class="component">🔒 Security</div>
            <div class="component">🖥️ Desktop Environment</div>
            <div class="component">⚙️ System Services</div>
            <div class="component">📦 Packages</div>
        </div>
        
        <div class="canvas">
            <div class="placeholder">
                <h3>Drag components here to build your configuration</h3>
                <p>The visual configuration builder will be fully implemented soon.</p>
            </div>
        </div>
        
        <div class="preview">
            <h3>Preview</h3>
            <pre id="config-preview">{ config, pkgs, ... }:
{
  # Your generated configuration will appear here
  system.stateVersion = "24.05";
}</pre>
        </div>
    </div>
</body>
</html>`, title, title)
}

func (s *EnhancedServer) getFleetHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .nav { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .nav a { margin-right: 20px; text-decoration: none; color: #2563eb; font-weight: 500; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .machine { display: flex; justify-content: space-between; align-items: center; padding: 15px; border: 1px solid #e5e7eb; border-radius: 6px; margin: 10px 0; }
        .status { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 500; }
        .status.healthy { background: #dcfce7; color: #166534; }
        .status.warning { background: #fef3c7; color: #92400e; }
        .loading { text-align: center; padding: 40px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 %s</h1>
        <p>Manage your NixOS fleet</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/builder">Configuration Builder</a>
        <a href="/fleet">Fleet Management</a>
        <a href="/teams">Teams</a>
        <a href="/versions">Version Control</a>
    </div>
    
    <div class="card">
        <h3>💻 Machines</h3>
        <div id="machines-list" class="loading">Loading machines...</div>
    </div>
    
    <div class="card">
        <h3>📦 Deployments</h3>
        <div id="deployments-list" class="loading">Loading deployments...</div>
    </div>
    
    <script>
        async function loadFleet() {
            try {
                // Load machines
                const machines = await fetch('/api/v1/fleet/machines').then(r => r.json());
                if (machines.success) {
                    updateMachines(machines.data);
                } else {
                    document.getElementById('machines-list').innerHTML = '<p>Fleet management not available</p>';
                }
                
                // Load deployments
                const deployments = await fetch('/api/v1/fleet/deployments').then(r => r.json());
                if (deployments.success) {
                    updateDeployments(deployments.data);
                } else {
                    document.getElementById('deployments-list').innerHTML = '<p>No deployments found</p>';
                }
            } catch (error) {
                console.error('Failed to load fleet:', error);
                document.getElementById('machines-list').innerHTML = '<p>Error loading machines</p>';
                document.getElementById('deployments-list').innerHTML = '<p>Error loading deployments</p>';
            }
        }
        
        function updateMachines(machines) {
            const html = machines.map(machine => 
                '<div class="machine">' +
                '<div><strong>' + machine.name + '</strong><br><small>' + machine.address + '</small></div>' +
                '<div class="status healthy">' + (machine.status || 'healthy') + '</div>' +
                '</div>'
            ).join('');
            document.getElementById('machines-list').innerHTML = html || '<p>No machines found</p>';
        }
        
        function updateDeployments(deployments) {
            const html = deployments.map(deployment => 
                '<div class="machine">' +
                '<div><strong>' + deployment.name + '</strong><br><small>' + deployment.targets.length + ' targets</small></div>' +
                '<div class="status ' + (deployment.status === 'completed' ? 'healthy' : 'warning') + '">' + deployment.status + '</div>' +
                '</div>'
            ).join('');
            document.getElementById('deployments-list').innerHTML = html || '<p>No deployments found</p>';
        }
        
        loadFleet();
        setInterval(loadFleet, 30000);
    </script>
</body>
</html>`, title, title)
}

func (s *EnhancedServer) getTeamsHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .nav { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .nav a { margin-right: 20px; text-decoration: none; color: #2563eb; font-weight: 500; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .team { padding: 15px; border: 1px solid #e5e7eb; border-radius: 6px; margin: 10px 0; }
        .loading { text-align: center; padding: 40px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="header">
        <h1>👥 %s</h1>
        <p>Collaborate with your team</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/builder">Configuration Builder</a>
        <a href="/fleet">Fleet Management</a>
        <a href="/teams">Teams</a>
        <a href="/versions">Version Control</a>
    </div>
    
    <div class="card">
        <h3>🏢 Teams</h3>
        <div id="teams-list" class="loading">Loading teams...</div>
    </div>
    
    <div class="card">
        <h3>💬 Active Sessions</h3>
        <div id="sessions-list" class="loading">Loading sessions...</div>
    </div>
    
    <script>
        async function loadTeams() {
            try {
                const teams = await fetch('/api/v1/teams').then(r => r.json());
                if (teams.success) {
                    updateTeams(teams.data);
                } else {
                    document.getElementById('teams-list').innerHTML = '<p>Team collaboration not available</p>';
                }
                
                const sessions = await fetch('/api/v1/collaborate/sessions').then(r => r.json());
                if (sessions.success) {
                    updateSessions(sessions.data);
                }
            } catch (error) {
                console.error('Failed to load teams:', error);
                document.getElementById('teams-list').innerHTML = '<p>Error loading teams</p>';
            }
        }
        
        function updateTeams(teams) {
            const html = teams.map(team => 
                '<div class="team">' +
                '<h4>' + team.name + '</h4>' +
                '<p>' + (team.description || 'No description') + '</p>' +
                '<small>' + team.members.length + ' members</small>' +
                '</div>'
            ).join('');
            document.getElementById('teams-list').innerHTML = html || '<p>No teams found</p>';
        }
        
        function updateSessions(sessions) {
            const html = sessions.map(session => 
                '<div class="team">' +
                '<h4>' + session.name + '</h4>' +
                '<small>' + session.participants + ' participants</small>' +
                '</div>'
            ).join('');
            document.getElementById('sessions-list').innerHTML = html || '<p>No active sessions</p>';
        }
        
        loadTeams();
        setInterval(loadTeams, 30000);
    </script>
</body>
</html>`, title, title)
}

func (s *EnhancedServer) getVersionsHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .nav { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .nav a { margin-right: 20px; text-decoration: none; color: #2563eb; font-weight: 500; }
        .card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .commit { padding: 15px; border: 1px solid #e5e7eb; border-radius: 6px; margin: 10px 0; }
        .loading { text-align: center; padding: 40px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📝 %s</h1>
        <p>Track configuration changes</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/builder">Configuration Builder</a>
        <a href="/fleet">Fleet Management</a>
        <a href="/teams">Teams</a>
        <a href="/versions">Version Control</a>
    </div>
    
    <div class="card">
        <h3>🌿 Branches</h3>
        <div id="branches-list" class="loading">Loading branches...</div>
    </div>
    
    <div class="card">
        <h3>📝 Recent Commits</h3>
        <div id="commits-list" class="loading">Loading commits...</div>
    </div>
    
    <script>
        async function loadVersions() {
            try {
                const branches = await fetch('/api/v1/config/branches').then(r => r.json());
                if (branches.success) {
                    updateBranches(branches.data.branches);
                } else {
                    document.getElementById('branches-list').innerHTML = '<p>Version control not available</p>';
                }
                
                const commits = await fetch('/api/v1/config/commits').then(r => r.json());
                if (commits.success) {
                    updateCommits(commits.data.commits);
                }
            } catch (error) {
                console.error('Failed to load versions:', error);
                document.getElementById('branches-list').innerHTML = '<p>Error loading branches</p>';
            }
        }
        
        function updateBranches(branches) {
            const html = branches.map(branch => 
                '<div class="commit">' +
                '<strong>' + branch.name + '</strong>' +
                '<p>' + (branch.description || 'No description') + '</p>' +
                '</div>'
            ).join('');
            document.getElementById('branches-list').innerHTML = html || '<p>No branches found</p>';
        }
        
        function updateCommits(commits) {
            const html = commits.map(commit => 
                '<div class="commit">' +
                '<strong>' + (commit.message || commit.id) + '</strong>' +
                '<p><small>' + (commit.author || 'Unknown') + ' • ' + formatTime(commit.timestamp || commit.created_at) + '</small></p>' +
                '</div>'
            ).join('');
            document.getElementById('commits-list').innerHTML = html || '<p>No commits found</p>';
        }
        
        function formatTime(timestamp) {
            return new Date(timestamp).toLocaleString();
        }
        
        loadVersions();
        setInterval(loadVersions, 30000);
    </script>
</body>
</html>`, title, title)
}

func (s *EnhancedServer) getDefaultHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: #2563eb; color: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; text-align: center; }
        .nav { background: white; padding: 15px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); text-align: center; }
        .nav a { margin: 0 20px; text-decoration: none; color: #2563eb; font-weight: 500; }
        .nav a:hover { text-decoration: underline; }
        .content { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); text-align: center; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🌐 %s</h1>
        <p>Collaborative NixOS Configuration Management</p>
    </div>
    
    <div class="nav">
        <a href="/">Dashboard</a>
        <a href="/builder">Configuration Builder</a>
        <a href="/fleet">Fleet Management</a>
        <a href="/teams">Teams</a>
        <a href="/versions">Version Control</a>
    </div>
    
    <div class="content">
        <h2>Welcome to nixai Web Interface</h2>
        <p>This enhanced web interface provides comprehensive management capabilities for your NixOS infrastructure.</p>
        
        <h3>🚀 Features Available</h3>
        <ul style="text-align: left; display: inline-block;">
            <li>📊 <strong>Dashboard</strong> - Overview of your entire infrastructure</li>
            <li>🎨 <strong>Visual Configuration Builder</strong> - Drag-and-drop configuration creation</li>
            <li>🚀 <strong>Fleet Management</strong> - Manage multiple machines</li>
            <li>👥 <strong>Team Collaboration</strong> - Work together in real-time</li>
            <li>📝 <strong>Version Control</strong> - Git-like configuration management</li>
            <li>🤖 <strong>AI Generation</strong> - AI-powered configuration assistance</li>
            <li>📡 <strong>Real-time Updates</strong> - WebSocket-based live updates</li>
        </ul>
        
        <h3>🔗 API Documentation</h3>
        <p>REST API available at <code>/api/v1</code></p>
        <p>WebSocket endpoint at <code>/api/v1/collaborate/ws</code></p>
    </div>
</body>
</html>`, title, title)
}
