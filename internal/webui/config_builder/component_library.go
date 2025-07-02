package config_builder

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type ComponentTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	File        string `json:"file"`
}

// ListTemplates loads all NixOS config templates from the templates/ directory
func ListTemplates(templatesDir string) ([]ComponentTemplate, error) {
	var templates []ComponentTemplate
	files, err := ioutil.ReadDir(templatesDir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		content, err := os.ReadFile(filepath.Join(templatesDir, name))
		if err != nil {
			continue
		}
		templates = append(templates, ComponentTemplate{
			Name:        name,
			Description: "NixOS component template",
			File:        string(content),
		})
	}
	return templates, nil
}

// TemplatesAPIHandler serves the list of templates as JSON
func TemplatesAPIHandler(templatesDir string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		templates, err := ListTemplates(templatesDir)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(`{"error":"failed to load templates"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates)
	}
}
