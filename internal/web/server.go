package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/pkg/logger"
)

// Server represents the web server for nixai
type Server struct {
	port          int
	router        *mux.Router
	teamManager   *team.TeamManager
	configRepo    *repository.ConfigRepository
	logger        *logger.Logger
	upgrader      websocket.Upgrader
	wsConnections map[string]*websocket.Conn
	shutdownChan  chan bool
}

// NewServer creates a new web server
func NewServer(port int, teamManager *team.TeamManager, configRepo *repository.ConfigRepository, logger *logger.Logger) *Server {
	server := &Server{
		port:        port,
		teamManager: teamManager,
		configRepo:  configRepo,
		logger:      logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
		wsConnections: make(map[string]*websocket.Conn),
		shutdownChan:  make(chan bool),
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	s.router = mux.NewRouter()

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Authentication routes
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	api.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	api.HandleFunc("/auth/status", s.handleAuthStatus).Methods("GET")

	// Team management routes
	api.HandleFunc("/teams", s.handleListTeams).Methods("GET")
	api.HandleFunc("/teams", s.handleCreateTeam).Methods("POST")
	api.HandleFunc("/teams/{teamId}", s.handleGetTeam).Methods("GET")
	api.HandleFunc("/teams/{teamId}", s.handleUpdateTeam).Methods("PUT")
	api.HandleFunc("/teams/{teamId}", s.handleDeleteTeam).Methods("DELETE")
	api.HandleFunc("/teams/{teamId}/members", s.handleListTeamMembers).Methods("GET")
	api.HandleFunc("/teams/{teamId}/members", s.handleAddTeamMember).Methods("POST")
	api.HandleFunc("/teams/{teamId}/members/{userId}", s.handleUpdateTeamMember).Methods("PUT")
	api.HandleFunc("/teams/{teamId}/members/{userId}", s.handleRemoveTeamMember).Methods("DELETE")

	// Configuration management routes
	api.HandleFunc("/config/branches", s.handleListBranches).Methods("GET")
	api.HandleFunc("/config/branches", s.handleCreateBranch).Methods("POST")
	api.HandleFunc("/config/branches/{branchName}", s.handleGetBranch).Methods("GET")
	api.HandleFunc("/config/branches/{branchName}", s.handleDeleteBranch).Methods("DELETE")
	api.HandleFunc("/config/commits", s.handleListCommits).Methods("GET")
	api.HandleFunc("/config/commits", s.handleCreateCommit).Methods("POST")
	api.HandleFunc("/config/commits/{commitId}", s.handleGetCommit).Methods("GET")
	api.HandleFunc("/config/files", s.handleListFiles).Methods("GET")
	api.HandleFunc("/config/files/{fileName}", s.handleGetFile).Methods("GET")
	api.HandleFunc("/config/files/{fileName}", s.handleUpdateFile).Methods("PUT")

	// Real-time collaboration routes
	api.HandleFunc("/collaborate/ws", s.handleWebSocket)
	api.HandleFunc("/collaborate/sessions", s.handleListSessions).Methods("GET")
	api.HandleFunc("/collaborate/sessions", s.handleCreateSession).Methods("POST")

	// Static file serving
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./internal/web/static/"))))

	// Serve the main application
	s.router.PathPrefix("/").HandlerFunc(s.handleIndex)
}

// Start starts the web server
func (s *Server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info(fmt.Sprintf("Starting web server on port %d", s.port))

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(fmt.Sprintf("Server failed to start: %v", err))
		}
	}()

	// Wait for shutdown signal
	<-s.shutdownChan

	// Shutdown server gracefully
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// Stop stops the web server
func (s *Server) Stop() {
	s.shutdownChan <- true
}

// Authentication handlers

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual authentication
	// For now, create a simple session
	session := map[string]interface{}{
		"user_id":   "user_123",
		"username":  loginReq.Username,
		"logged_in": true,
	}

	s.sendJSON(w, session)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session cleanup
	s.sendJSON(w, map[string]string{"status": "logged_out"})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Check actual session status
	status := map[string]interface{}{
		"authenticated": true,
		"user_id":       "user_123",
		"username":      "demo_user",
	}
	s.sendJSON(w, status)
}

// Team management handlers

func (s *Server) handleListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := s.teamManager.ListTeams(r.Context())
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendJSON(w, teams)
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
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

	s.sendJSON(w, team)
}

func (s *Server) handleGetTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["teamId"]

	team, err := s.teamManager.GetTeam(r.Context(), teamID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendJSON(w, team)
}

func (s *Server) handleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement team updates
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement team deletion
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleListTeamMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["teamId"]

	team, err := s.teamManager.GetTeam(r.Context(), teamID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendJSON(w, team.Members)
}

func (s *Server) handleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["teamId"]

	var addReq struct {
		UserID string    `json:"user_id"`
		Role   team.Role `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&addReq); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get actual user ID from session
	invitedBy := "user_123"

	err := s.teamManager.AddMember(r.Context(), teamID, addReq.UserID, addReq.Role, invitedBy)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendJSON(w, map[string]string{"status": "member_added"})
}

func (s *Server) handleUpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement member updates
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID := vars["teamId"]
	userID := vars["userId"]

	// TODO: Get actual user ID from session
	removedBy := "user_123"

	err := s.teamManager.RemoveMember(r.Context(), teamID, userID, removedBy)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendJSON(w, map[string]string{"status": "member_removed"})
}

// Configuration management handlers

func (s *Server) handleListBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := s.configRepo.ListBranches(r.Context())
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendJSON(w, map[string]interface{}{
		"branches": branches,
	})
}

func (s *Server) handleCreateBranch(w http.ResponseWriter, r *http.Request) {
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

	s.sendJSON(w, map[string]string{
		"status": "branch_created",
		"name":   createReq.Name,
	})
}

func (s *Server) handleGetBranch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement branch details
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleDeleteBranch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	branchName := vars["branchName"]

	err := s.configRepo.DeleteBranch(r.Context(), branchName)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.sendJSON(w, map[string]string{
		"status": "branch_deleted",
		"name":   branchName,
	})
}

func (s *Server) handleListCommits(w http.ResponseWriter, r *http.Request) {
	snapshots, err := s.configRepo.ListSnapshots(r.Context())
	if err != nil {
		s.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendJSON(w, map[string]interface{}{
		"commits": snapshots,
	})
}

func (s *Server) handleCreateCommit(w http.ResponseWriter, r *http.Request) {
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

	s.sendJSON(w, snapshot)
}

func (s *Server) handleGetCommit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commitID := vars["commitId"]

	snapshot, err := s.configRepo.GetSnapshot(r.Context(), commitID)
	if err != nil {
		s.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	s.sendJSON(w, snapshot)
}

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	// TODO: Get files from current branch
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	// TODO: Get specific file content
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleUpdateFile(w http.ResponseWriter, r *http.Request) {
	// TODO: Update file content
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

// Real-time collaboration handlers

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error(fmt.Sprintf("WebSocket upgrade failed: %v", err))
		return
	}
	defer conn.Close()

	// TODO: Get user ID from session
	userID := "user_123"
	s.wsConnections[userID] = conn

	s.logger.Info(fmt.Sprintf("WebSocket connection established for user %s", userID))

	// Handle WebSocket messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			s.logger.Error(fmt.Sprintf("WebSocket read error: %v", err))
			break
		}

		// Process WebSocket message
		s.handleWebSocketMessage(userID, msg)
	}

	delete(s.wsConnections, userID)
}

func (s *Server) handleWebSocketMessage(userID string, message map[string]interface{}) {
	msgType, ok := message["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		s.sendWebSocketMessage(userID, map[string]interface{}{
			"type": "pong",
		})
	case "join_session":
		// TODO: Handle session joining
	case "edit_file":
		// TODO: Handle real-time file editing
	default:
		s.logger.Warn(fmt.Sprintf("Unknown WebSocket message type: %s", msgType))
	}
}

func (s *Server) sendWebSocketMessage(userID string, message map[string]interface{}) {
	conn, exists := s.wsConnections[userID]
	if !exists {
		return
	}

	if err := conn.WriteJSON(message); err != nil {
		s.logger.Error(fmt.Sprintf("WebSocket write error: %v", err))
		delete(s.wsConnections, userID)
	}
}

func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	// TODO: List active collaboration sessions
	sessions := []map[string]interface{}{
		{
			"id":           "session_1",
			"name":         "Production Config Review",
			"participants": 3,
			"created_at":   time.Now(),
		},
	}
	s.sendJSON(w, sessions)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Create new collaboration session
	s.sendError(w, "Not implemented", http.StatusNotImplemented)
}

// Static content handler

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// TODO: Serve the main application HTML
	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>nixai Configuration Manager</title>
		<meta charset="utf-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
			.header { background: #333; color: white; padding: 20px; margin: -20px -20px 20px -20px; }
			.nav { margin: 20px 0; }
			.nav a { margin-right: 20px; text-decoration: none; color: #333; }
			.nav a:hover { color: #666; }
		</style>
	</head>
	<body>
		<div class="header">
			<h1>nixai Configuration Manager</h1>
			<p>Collaborative NixOS Configuration Management</p>
		</div>
		<div class="nav">
			<a href="/api/v1/teams">Teams</a>
			<a href="/api/v1/config/branches">Branches</a>
			<a href="/api/v1/config/commits">Commits</a>
		</div>
		<div>
			<h2>Welcome to nixai Web Interface</h2>
			<p>This is a basic web interface for nixai. A full React/Vue.js frontend will be implemented in the future.</p>
			<p>Available API endpoints:</p>
			<ul>
				<li>GET /api/v1/teams - List teams</li>
				<li>POST /api/v1/teams - Create team</li>
				<li>GET /api/v1/config/branches - List branches</li>
				<li>GET /api/v1/config/commits - List commits</li>
			</ul>
		</div>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// Utility functions

func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
