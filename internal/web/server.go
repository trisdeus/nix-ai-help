package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/fleet"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/internal/webui"
	"nix-ai-help/pkg/logger"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// ServerConfig holds configuration for the web server
type ServerConfig struct {
	Port           int             `yaml:"port"`
	Host           string          `yaml:"host"`
	StaticDir      string          `yaml:"static_dir"`
	TemplateDir    string          `yaml:"template_dir"`
	Authentication AuthConfig      `yaml:"authentication"`
	TLS            TLSConfig       `yaml:"tls"`
	CORS           CORSConfig      `yaml:"cors"`
	RateLimit      RateLimitConfig `yaml:"rate_limit"`
	Features       FeatureConfig   `yaml:"features"`
}

type AuthConfig struct {
	Enabled        bool     `yaml:"enabled"`
	SessionTimeout string   `yaml:"session_timeout"`
	Providers      []string `yaml:"providers"`
	JWTSecret      string   `yaml:"jwt_secret"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

type RateLimitConfig struct {
	Enabled bool `yaml:"enabled"`
	RPM     int  `yaml:"rpm"`
}

type FeatureConfig struct {
	VisualBuilder   bool `yaml:"visual_builder"`
	FleetManagement bool `yaml:"fleet_management"`
	Collaboration   bool `yaml:"collaboration"`
	VersionControl  bool `yaml:"version_control"`
	AIGeneration    bool `yaml:"ai_generation"`
	Dashboard       bool `yaml:"dashboard"`
}

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Role      string                 `json:"role"`
	Teams     []string               `json:"teams"`
	Data      map[string]interface{} `json:"data"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// Server represents the web server for nixai
type Server struct {
	config        *ServerConfig
	router        *mux.Router
	teamManager   *team.TeamManager
	configRepo    *repository.ConfigRepository
	fleetManager  *fleet.FleetManager
	configBuilder *webui.ConfigBuilderAPI
	aiProvider    ai.AIProvider
	logger        *logger.Logger

	// WebSocket management
	upgrader      websocket.Upgrader
	wsConnections map[string]*websocket.Conn
	wsMutex       sync.RWMutex

	// Session management
	sessions     map[string]*Session
	sessionMutex sync.RWMutex

	// Shutdown
	shutdownChan chan bool
	httpServer   *http.Server
}

// NewServer creates a new web server with enhanced functionality
func NewServer(config *ServerConfig, teamManager *team.TeamManager, configRepo *repository.ConfigRepository, fleetManager *fleet.FleetManager, logger *logger.Logger) (*Server, error) {
	// Initialize configuration builder API
	configBuilder, err := webui.NewConfigBuilderAPI(fleetManager, logger)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to initialize config builder: %v", err))
		configBuilder = nil
	}

	server := &Server{
		config:        config,
		teamManager:   teamManager,
		configRepo:    configRepo,
		fleetManager:  fleetManager,
		configBuilder: configBuilder,
		logger:        logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper CORS checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		wsConnections: make(map[string]*websocket.Conn),
		sessions:      make(map[string]*Session),
		shutdownChan:  make(chan bool, 1),
	}

	// Initialize router
	server.router = mux.NewRouter()
	server.setupRoutes()

	return server, nil
}

// Start starts the web server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info(fmt.Sprintf("Starting web server on %s", addr))

	// Start the server and block until it's shut down
	var err error
	if s.config.TLS.Enabled && s.config.TLS.CertFile != "" && s.config.TLS.KeyFile != "" {
		err = s.httpServer.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile)
	} else {
		err = s.httpServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		s.logger.Error(fmt.Sprintf("Web server error: %v", err))
		return err
	}

	return nil
}

// Stop gracefully stops the web server
func (s *Server) Stop() {
	if s.httpServer == nil {
		return
	}

	s.logger.Info("Stopping web server...")

	// Close all WebSocket connections
	s.wsMutex.Lock()
	for _, conn := range s.wsConnections {
		conn.Close()
	}
	s.wsConnections = make(map[string]*websocket.Conn)
	s.wsMutex.Unlock()

	// Shutdown the HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(fmt.Sprintf("Error shutting down web server: %v", err))
	} else {
		s.logger.Info("Web server stopped gracefully")
	}
}

// setupRoutes initializes the HTTP routes
func (s *Server) setupRoutes() {
	// Serve static files
	if s.config.StaticDir != "" {
		fs := http.FileServer(http.Dir(s.config.StaticDir))
		s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	}

	// Health check endpoint
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Basic routes - these will be enhanced by EnhancedServer
	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
}

// handleHealth provides a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleIndex provides a basic index page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>NixAI Web Interface</title>
</head>
<body>
    <h1>NixAI Web Interface</h1>
    <p>Welcome to the NixAI web interface. This is the basic server implementation.</p>
    <p>For enhanced features, use the EnhancedServer.</p>
</body>
</html>
`)
}
