package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Health and Status Handlers

func (s *EnhancedServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"services": map[string]interface{}{
			"web_server":    "healthy",
			"team_manager":  s.getServiceStatus(s.teamManager != nil),
			"config_repo":   s.getServiceStatus(s.configRepo != nil),
			"fleet_manager": s.getServiceStatus(s.fleetManager != nil),
			"ai_provider":   s.getServiceStatus(s.aiProvider != nil),
		},
	}

	s.sendJSON(w, health)
}

func (s *EnhancedServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"server": map[string]interface{}{
			"uptime":      time.Since(time.Now()).String(), // TODO: track actual uptime
			"connections": len(s.wsConnections),
			"sessions":    len(s.sessions),
		},
		"features": s.config.Features,
		"config": map[string]interface{}{
			"port":        s.config.Port,
			"host":        s.config.Host,
			"tls_enabled": s.config.TLS.Enabled,
		},
	}

	s.sendJSON(w, status)
}

func (s *EnhancedServer) handleVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]interface{}{
		"version":    "1.0.0",
		"build_time": "2025-06-30T00:00:00Z", // TODO: Set during build
		"git_commit": "HEAD",                 // TODO: Set during build
		"go_version": "go1.24.3",
	}

	s.sendJSON(w, version)
}

// Dashboard Handlers

func (s *EnhancedServer) handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	overview := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_machines":     s.getTotalMachines(),
			"active_deployments": s.getActiveDeployments(),
			"team_count":         s.getTeamCount(),
			"config_branches":    s.getConfigBranches(),
		},
		"recent_activities": s.getRecentActivities(),
		"system_health":     s.getSystemHealth(),
		"alerts":            s.getSystemAlerts(),
	}

	s.sendJSON(w, overview)
}

// Dashboard methods moved to enhanced_server.go to avoid conflicts

// Authentication Handlers

func (s *EnhancedServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if loginRequest.Username == "" || loginRequest.Password == "" {
		s.sendError(w, "Username and password required", http.StatusBadRequest)
		return
	}

	// Use real authentication
	if s.authManager == nil {
		s.sendError(w, "Authentication system not available", http.StatusInternalServerError)
		return
	}

	loginResponse, err := s.authManager.Authenticate(loginRequest.Username, loginRequest.Password)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Authentication error: %v", err))
		s.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse)
}

func (s *EnhancedServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	token := ""
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Invalidate session if auth manager is available and token provided
	if s.authManager != nil && token != "" {
		if err := s.authManager.Logout(token); err != nil {
			s.logger.Warn(fmt.Sprintf("Error during logout: %v", err))
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Team Management Handlers

func (s *EnhancedServer) handleListTeams(w http.ResponseWriter, r *http.Request) {
	if s.teamManager == nil {
		s.sendError(w, "Team management not available", http.StatusServiceUnavailable)
		return
	}

	teams, err := s.teamManager.ListTeams(r.Context())
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, teams)
}

func (s *EnhancedServer) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	if s.teamManager == nil {
		s.sendError(w, "Team management not available", http.StatusServiceUnavailable)
		return
	}

	var createReq struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get actual user ID from session
	userID := "user_123"

	team, err := s.teamManager.CreateTeam(r.Context(), createReq.Name, createReq.Description, userID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendSuccess(w, team)
}

func (s *EnhancedServer) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	if s.teamManager == nil {
		s.sendError(w, "Team management not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	team, err := s.teamManager.GetTeam(r.Context(), teamID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendSuccess(w, team)
}

func (s *EnhancedServer) handleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleListTeamMembers(w http.ResponseWriter, r *http.Request) {
	if s.teamManager == nil {
		s.sendError(w, "Team management not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	team, err := s.teamManager.GetTeam(r.Context(), teamID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendSuccess(w, team.Members)
}

func (s *EnhancedServer) handleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleUpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

// Configuration Management Handlers

func (s *EnhancedServer) handleListBranches(w http.ResponseWriter, r *http.Request) {
	if s.configRepo == nil {
		s.sendError(w, "Configuration repository not available", http.StatusServiceUnavailable)
		return
	}

	branches, err := s.configRepo.ListBranches(r.Context())
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"branches": branches,
	})
}

func (s *EnhancedServer) handleCreateBranch(w http.ResponseWriter, r *http.Request) {
	if s.configRepo == nil {
		s.sendError(w, "Configuration repository not available", http.StatusServiceUnavailable)
		return
	}

	var createReq struct {
		Name       string `json:"name"`
		FromCommit string `json:"from_commit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := s.configRepo.CreateBranch(r.Context(), createReq.Name, createReq.FromCommit)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendSuccess(w, map[string]string{
		"status": "branch_created",
		"name":   createReq.Name,
	})
}

func (s *EnhancedServer) handleGetBranch(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleDeleteBranch(w http.ResponseWriter, r *http.Request) {
	if s.configRepo == nil {
		s.sendError(w, "Configuration repository not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	branchName := vars["branchName"]

	err := s.configRepo.DeleteBranch(r.Context(), branchName)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendSuccess(w, map[string]string{
		"status": "branch_deleted",
		"name":   branchName,
	})
}

func (s *EnhancedServer) handleListCommits(w http.ResponseWriter, r *http.Request) {
	if s.configRepo == nil {
		s.sendError(w, "Configuration repository not available", http.StatusServiceUnavailable)
		return
	}

	snapshots, err := s.configRepo.ListSnapshots(r.Context())
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, map[string]interface{}{
		"commits": snapshots,
	})
}

func (s *EnhancedServer) handleCreateCommit(w http.ResponseWriter, r *http.Request) {
	if s.configRepo == nil {
		s.sendError(w, "Configuration repository not available", http.StatusServiceUnavailable)
		return
	}

	var commitReq struct {
		Message  string            `json:"message"`
		Files    map[string]string `json:"files"`
		Metadata map[string]string `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&commitReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	snapshot, err := s.configRepo.Commit(r.Context(), commitReq.Message, commitReq.Files, commitReq.Metadata)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, snapshot)
}

func (s *EnhancedServer) handleGetCommit(w http.ResponseWriter, r *http.Request) {
	if s.configRepo == nil {
		s.sendError(w, "Configuration repository not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	commitID := vars["commitId"]

	snapshot, err := s.configRepo.GetSnapshot(r.Context(), commitID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendSuccess(w, snapshot)
}

func (s *EnhancedServer) handleListFiles(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleGetFile(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *EnhancedServer) handleUpdateFile(w http.ResponseWriter, r *http.Request) {
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

// Helper functions for dashboard stats

func (s *EnhancedServer) getServiceStatus(available bool) string {
	if available {
		return "healthy"
	}
	return "unavailable"
}

func (s *EnhancedServer) getTotalMachines() int {
	if s.fleetManager == nil {
		return 0
	}
	// TODO: Implement actual count
	return 5
}

func (s *EnhancedServer) getActiveDeployments() int {
	if s.fleetManager == nil {
		return 0
	}
	// TODO: Implement actual count
	return 2
}

func (s *EnhancedServer) getTeamCount() int {
	if s.teamManager == nil {
		return 0
	}
	// TODO: Implement actual count
	return 3
}

func (s *EnhancedServer) getConfigBranches() int {
	if s.configRepo == nil {
		return 0
	}
	// TODO: Implement actual count
	return 4
}

func (s *EnhancedServer) getRecentActivities() []map[string]interface{} {
	// TODO: Implement actual activity tracking
	return []map[string]interface{}{
		{
			"id":        "activity_1",
			"type":      "deployment",
			"message":   "Deployed configuration to production servers",
			"user":      "admin",
			"timestamp": time.Now().Add(-2 * time.Hour),
		},
		{
			"id":        "activity_2",
			"type":      "commit",
			"message":   "Updated nginx configuration",
			"user":      "devops",
			"timestamp": time.Now().Add(-4 * time.Hour),
		},
	}
}

func (s *EnhancedServer) getSystemHealth() map[string]interface{} {
	return map[string]interface{}{
		"overall": "healthy",
		"cpu":     "normal",
		"memory":  "normal",
		"disk":    "normal",
		"network": "normal",
	}
}

func (s *EnhancedServer) getSystemAlerts() []map[string]interface{} {
	// TODO: Implement actual alerting system
	return []map[string]interface{}{
		{
			"id":        "alert_1",
			"level":     "warning",
			"message":   "Machine server-03 has high CPU usage",
			"timestamp": time.Now().Add(-30 * time.Minute),
		},
	}
}

func (s *EnhancedServer) getTotalDeployments() int      { return 25 }
func (s *EnhancedServer) getSuccessfulDeployments() int { return 23 }
func (s *EnhancedServer) getFailedDeployments() int     { return 1 }
func (s *EnhancedServer) getInProgressDeployments() int { return 1 }
func (s *EnhancedServer) getHealthyMachines() int       { return 4 }
func (s *EnhancedServer) getWarningMachines() int       { return 1 }
func (s *EnhancedServer) getErrorMachines() int         { return 0 }
func (s *EnhancedServer) getTotalCommits() int          { return 47 }

// Additional Teams API Handlers

func (s *EnhancedServer) handleTeamsStats(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := map[string]interface{}{
		"total_teams":     3,
		"active_projects": 7,
		"pending_reviews": 2,
		"team_members":    12,
		"collaborations":  5,
	}

	s.sendSuccess(w, stats)
}

func (s *EnhancedServer) handleTeamsActivity(w http.ResponseWriter, r *http.Request) {
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
			"id":          "1",
			"type":        "team_create",
			"description": "New team 'DevOps Squad' created",
			"timestamp":   "2 hours ago",
			"user":        "demo",
		},
		{
			"id":          "2",
			"type":        "project_create",
			"description": "Project 'Web Server Config' shared",
			"timestamp":   "4 hours ago",
			"user":        "alice",
		},
		{
			"id":          "3",
			"type":        "collaboration",
			"description": "Started collaboration session on 'Database Setup'",
			"timestamp":   "6 hours ago",
			"user":        "bob",
		},
	}

	s.sendSuccess(w, activities)
}

func (s *EnhancedServer) handlePublicTeams(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	publicTeams := []map[string]interface{}{
		{
			"id":           "public-1",
			"name":         "NixOS Community",
			"description":  "Open community for NixOS enthusiasts",
			"member_count": 156,
			"is_public":    true,
		},
		{
			"id":           "public-2",
			"name":         "DevOps Best Practices",
			"description":  "Sharing DevOps configurations and practices",
			"member_count": 89,
			"is_public":    true,
		},
	}

	s.sendSuccess(w, publicTeams)
}

func (s *EnhancedServer) handleJoinTeam(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var joinReq struct {
		InvitationCode string `json:"invitation_code"`
		TeamID         string `json:"team_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&joinReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Mock team join logic
	if joinReq.InvitationCode == "" && joinReq.TeamID == "" {
		s.sendError(w, "Either invitation code or team ID is required", http.StatusBadRequest)
		return
	}

	// Return success response for demo
	response := map[string]interface{}{
		"team_id":   "team-123",
		"team_name": "Demo Team",
		"role":      "member",
		"joined_at": time.Now().Format(time.RFC3339),
	}

	s.sendSuccess(w, response)
}

func (s *EnhancedServer) handleStartCollaboration(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	teamID := vars["teamId"]

	if teamID == "" {
		s.sendError(w, "Team ID is required", http.StatusBadRequest)
		return
	}

	// Mock collaboration session creation
	session := map[string]interface{}{
		"id":         "collab-" + teamID + "-" + time.Now().Format("20060102150405"),
		"team_id":    teamID,
		"created_by": "demo",
		"created_at": time.Now().Format(time.RFC3339),
		"status":     "active",
		"participants": []map[string]interface{}{
			{
				"user_id":   "demo",
				"username":  "demo",
				"role":      "host",
				"joined_at": time.Now().Format(time.RFC3339),
			},
		},
	}

	s.sendSuccess(w, map[string]interface{}{
		"session": session,
	})
}
