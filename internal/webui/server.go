package webui

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

// Server represents the webui HTTP server
type Server struct {
	mux          *http.ServeMux
	templatesDir string
	port         int
}

// NewServer creates a new webui server instance
func NewServer(templatesDir string, port int) *Server {
	mux := http.NewServeMux()

	server := &Server{
		mux:          mux,
		templatesDir: templatesDir,
		port:         port,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes for the webui
func (s *Server) setupRoutes() {
	// Register API endpoints
	RegisterAPI(s.mux, s.templatesDir)

	// Serve the builder HTML page
	s.mux.HandleFunc("/builder", s.serveBuilder)

	// Serve static files (CSS, JS)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/webui/static/"))))

	// Root redirect to builder
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/builder", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
}

// serveBuilder serves the main builder HTML page
func (s *Server) serveBuilder(w http.ResponseWriter, r *http.Request) {
	builderPath := filepath.Join("internal", "webui", "templates", "builder.html")
	http.ServeFile(w, r, builderPath)
}

// Start starts the webui HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 Starting NixAI Visual Configuration Builder on http://localhost%s/builder", addr)

	return http.ListenAndServe(addr, s.mux)
}

// StartAsync starts the webui HTTP server in a goroutine
func (s *Server) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			log.Printf("WebUI server error: %v", err)
		}
	}()
}
