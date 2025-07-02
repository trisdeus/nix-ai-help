package web

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	webui "nix-ai-help/internal/webui"
)

// DefaultPort is the default port for the web server
const DefaultPort = 34567

// StartServer starts the web server for the visual configuration builder
func StartServer(addr string, templatesDir string, staticDir string) {
	if addr == "" {
		addr = fmt.Sprintf(":%d", DefaultPort)
	}
	mux := http.NewServeMux()

	// Serve static files (JS, CSS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Register REST API endpoints
	webui.RegisterAPI(mux, templatesDir)

	// Serve builder UI
	mux.HandleFunc("/builder", func(w http.ResponseWriter, r *http.Request) {
		tmplPath := filepath.Join(templatesDir, "builder.html")
		http.ServeFile(w, r, tmplPath)
	})

	fmt.Printf("Web UI available at http://localhost%s/builder\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
