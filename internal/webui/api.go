package webui

import (
	"encoding/json"
	"net/http"
	"nix-ai-help/internal/webui/config_builder"
)

// RegisterAPI registers REST API endpoints for the builder
func RegisterAPI(mux *http.ServeMux, templatesDir string) {
	mux.HandleFunc("/api/templates", func(w http.ResponseWriter, r *http.Request) {
		templates, err := config_builder.ListTemplates(templatesDir)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("[]"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates)
	})

	// Serve the builder HTML page
	mux.HandleFunc("/builder", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/webui/templates/builder.html")
	})

	// Serve static files (CSS, JS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/webui/static/"))))
}
