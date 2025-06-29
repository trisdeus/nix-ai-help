package webui

import (
	"encoding/json"
	"fmt"
	"net/http"

	"nix-ai-help/internal/webui/config_builder"
	"nix-ai-help/pkg/logger"

	"github.com/gorilla/mux"
)

// ConfigBuilderAPI provides API endpoints for the visual configuration builder
type ConfigBuilderAPI struct {
	library    *config_builder.ComponentLibrary
	dragDrop   *config_builder.DragDropInterface
	preview    *config_builder.RealTimePreview
	visualizer *config_builder.DependencyVisualizer
	logger     *logger.Logger
}

// NewConfigBuilderAPI creates a new configuration builder API
func NewConfigBuilderAPI(logger *logger.Logger) (*ConfigBuilderAPI, error) {
	library := config_builder.NewComponentLibrary(logger)
	dragDrop := config_builder.NewDragDropInterface(library, logger)

	canvas := dragDrop.GetCanvas()
	preview, err := config_builder.NewRealTimePreview(canvas, library, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create preview: %w", err)
	}

	visualizer := config_builder.NewDependencyVisualizer(canvas, library, logger)

	return &ConfigBuilderAPI{
		library:    library,
		dragDrop:   dragDrop,
		preview:    preview,
		visualizer: visualizer,
		logger:     logger,
	}, nil
}

// RegisterRoutes registers all API routes for the configuration builder
func (api *ConfigBuilderAPI) RegisterRoutes(router *mux.Router) {
	// Component library routes
	router.HandleFunc("/api/components", api.handleGetComponents).Methods("GET")
	router.HandleFunc("/api/components/{id}", api.handleGetComponent).Methods("GET")
	router.HandleFunc("/api/components/search", api.handleSearchComponents).Methods("GET")
	router.HandleFunc("/api/components/categories", api.handleGetCategories).Methods("GET")

	// Canvas management routes
	router.HandleFunc("/api/canvas", api.handleGetCanvas).Methods("GET")
	router.HandleFunc("/api/canvas/save", api.handleSaveCanvas).Methods("POST")
	router.HandleFunc("/api/canvas/load", api.handleLoadCanvas).Methods("POST")
	router.HandleFunc("/api/canvas/clear", api.handleClearCanvas).Methods("POST")

	// Component management routes
	router.HandleFunc("/api/canvas/components", api.handleAddComponent).Methods("POST")
	router.HandleFunc("/api/canvas/components/{instanceId}", api.handleRemoveComponent).Methods("DELETE")
	router.HandleFunc("/api/canvas/components/{instanceId}/position", api.handleUpdateComponentPosition).Methods("PUT")
	router.HandleFunc("/api/canvas/components/{instanceId}/config", api.handleUpdateComponentConfig).Methods("PUT")
	router.HandleFunc("/api/canvas/components/{instanceId}/size", api.handleUpdateComponentSize).Methods("PUT")

	// Connection management routes
	router.HandleFunc("/api/canvas/connections", api.handleCreateConnection).Methods("POST")
	router.HandleFunc("/api/canvas/connections/{connectionId}", api.handleRemoveConnection).Methods("DELETE")
	router.HandleFunc("/api/canvas/connections", api.handleGetConnections).Methods("GET")

	// Layout and organization routes
	router.HandleFunc("/api/canvas/layout", api.handleAutoLayout).Methods("POST")
	router.HandleFunc("/api/canvas/select", api.handleSelectComponents).Methods("POST")
	router.HandleFunc("/api/canvas/duplicate", api.handleDuplicateComponents).Methods("POST")

	// Undo/Redo routes
	router.HandleFunc("/api/canvas/undo", api.handleUndo).Methods("POST")
	router.HandleFunc("/api/canvas/redo", api.handleRedo).Methods("POST")

	// Preview and validation routes
	router.HandleFunc("/api/preview/generate", api.handleGeneratePreview).Methods("POST")
	router.HandleFunc("/api/preview/validate", api.handleValidateConfiguration).Methods("POST")
	router.HandleFunc("/api/preview/options", api.handleUpdatePreviewOptions).Methods("PUT")

	// Dependency analysis routes
	router.HandleFunc("/api/dependencies/analyze", api.handleAnalyzeDependencies).Methods("POST")
	router.HandleFunc("/api/dependencies/graph", api.handleGetDependencyGraph).Methods("GET")
	router.HandleFunc("/api/dependencies/validate", api.handleValidateDependencies).Methods("POST")
	router.HandleFunc("/api/dependencies/report", api.handleGetDependencyReport).Methods("GET")
	router.HandleFunc("/api/dependencies/export", api.handleExportDependencyGraph).Methods("GET")

	// Configuration generation routes
	router.HandleFunc("/api/generate/nixos", api.handleGenerateNixOSConfig).Methods("POST")
	router.HandleFunc("/api/generate/home-manager", api.handleGenerateHomeManagerConfig).Methods("POST")
	router.HandleFunc("/api/generate/flake", api.handleGenerateFlakeConfig).Methods("POST")
	router.HandleFunc("/api/generate/module", api.handleGenerateModuleConfig).Methods("POST")
}

// Component Library Handlers

func (api *ConfigBuilderAPI) handleGetComponents(w http.ResponseWriter, r *http.Request) {
	components := api.library.GetAllComponents()
	api.sendJSON(w, components)
}

func (api *ConfigBuilderAPI) handleGetComponent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	component, err := api.library.GetComponent(id)
	if err != nil {
		api.sendError(w, http.StatusNotFound, err.Error())
		return
	}

	api.sendJSON(w, component)
}

func (api *ConfigBuilderAPI) handleSearchComponents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		api.sendError(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}

	components := api.library.SearchComponents(query)
	api.sendJSON(w, components)
}

func (api *ConfigBuilderAPI) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories := map[string]interface{}{
		"system":      "System Services",
		"network":     "Network",
		"security":    "Security",
		"development": "Development",
		"media":       "Media",
		"gaming":      "Gaming",
		"database":    "Database",
		"webserver":   "Web Server",
		"desktop":     "Desktop",
		"utilities":   "Utilities",
	}

	api.sendJSON(w, categories)
}

// Canvas Management Handlers

func (api *ConfigBuilderAPI) handleGetCanvas(w http.ResponseWriter, r *http.Request) {
	canvas := api.dragDrop.GetCanvas()
	api.sendJSON(w, canvas)
}

func (api *ConfigBuilderAPI) handleSaveCanvas(w http.ResponseWriter, r *http.Request) {
	data, err := api.dragDrop.SaveCanvas()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save canvas: %v", err))
		return
	}

	// Here you would typically save to a file or database
	api.logger.Info("Canvas saved successfully")

	response := map[string]interface{}{
		"success": true,
		"message": "Canvas saved successfully",
		"size":    len(data),
	}

	api.sendJSON(w, response)
}

func (api *ConfigBuilderAPI) handleLoadCanvas(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Data []byte `json:"data"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.dragDrop.LoadCanvas(request.Data); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load canvas: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleClearCanvas(w http.ResponseWriter, r *http.Request) {
	// Reset the canvas to empty state
	api.dragDrop = config_builder.NewDragDropInterface(api.library, api.logger)

	canvas := api.dragDrop.GetCanvas()
	preview, err := config_builder.NewRealTimePreview(canvas, api.library, api.logger)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reset preview: %v", err))
		return
	}
	api.preview = preview

	api.sendJSON(w, map[string]bool{"success": true})
}

// Component Management Handlers

func (api *ConfigBuilderAPI) handleAddComponent(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ComponentID string                  `json:"component_id"`
		Position    config_builder.Position `json:"position"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	placedComponent, err := api.dragDrop.AddComponent(request.ComponentID, request.Position)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to add component: %v", err))
		return
	}

	// Generate instance ID for response
	response := map[string]interface{}{
		"instance_id": fmt.Sprintf("%s_%d", request.ComponentID, placedComponent.ZIndex),
		"component":   placedComponent.Component,
		"position":    placedComponent.Position,
		"size":        placedComponent.Size,
		"config":      placedComponent.Config,
	}

	api.sendJSON(w, response)
}

func (api *ConfigBuilderAPI) handleRemoveComponent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]

	if err := api.dragDrop.RemoveComponent(instanceID); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove component: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleUpdateComponentPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]

	var request struct {
		Position config_builder.Position `json:"position"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.dragDrop.MoveComponent(instanceID, request.Position); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update position: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleUpdateComponentConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]

	var request struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update component configuration
	canvas := api.dragDrop.GetCanvas()
	if component, exists := canvas.Components[instanceID]; exists {
		component.Config = request.Config
		api.sendJSON(w, map[string]bool{"success": true})
	} else {
		api.sendError(w, http.StatusNotFound, "Component not found")
	}
}

func (api *ConfigBuilderAPI) handleUpdateComponentSize(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instanceId"]

	var request struct {
		Size config_builder.Size `json:"size"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.dragDrop.ResizeComponent(instanceID, request.Size); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to resize component: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

// Connection Management Handlers

func (api *ConfigBuilderAPI) handleCreateConnection(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FromID string `json:"from_id"`
		ToID   string `json:"to_id"`
		Type   string `json:"type"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.dragDrop.ConnectComponents(request.FromID, request.ToID, request.Type); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create connection: %v", err))
		return
	}

	// Find the created connection
	canvas := api.dragDrop.GetCanvas()
	for _, conn := range canvas.Connections {
		if conn.FromID == request.FromID && conn.ToID == request.ToID && conn.Type == request.Type {
			api.sendJSON(w, conn)
			return
		}
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleRemoveConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectionID := vars["connectionId"]

	// Find and remove connection
	canvas := api.dragDrop.GetCanvas()
	for i, conn := range canvas.Connections {
		if conn.ID == connectionID {
			canvas.Connections = append(canvas.Connections[:i], canvas.Connections[i+1:]...)
			api.sendJSON(w, map[string]bool{"success": true})
			return
		}
	}

	api.sendError(w, http.StatusNotFound, "Connection not found")
}

func (api *ConfigBuilderAPI) handleGetConnections(w http.ResponseWriter, r *http.Request) {
	canvas := api.dragDrop.GetCanvas()
	api.sendJSON(w, canvas.Connections)
}

// Layout and Organization Handlers

func (api *ConfigBuilderAPI) handleAutoLayout(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Algorithm string `json:"algorithm"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.dragDrop.AutoLayout(request.Algorithm); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to apply layout: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleSelectComponents(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ComponentIDs []string `json:"component_ids"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.dragDrop.SelectComponents(request.ComponentIDs); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to select components: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleDuplicateComponents(w http.ResponseWriter, r *http.Request) {
	if err := api.dragDrop.DuplicateComponents(); err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to duplicate components: %v", err))
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

// Undo/Redo Handlers

func (api *ConfigBuilderAPI) handleUndo(w http.ResponseWriter, r *http.Request) {
	if err := api.dragDrop.Undo(); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

func (api *ConfigBuilderAPI) handleRedo(w http.ResponseWriter, r *http.Request) {
	if err := api.dragDrop.Redo(); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.sendJSON(w, map[string]bool{"success": true})
}

// Preview and Validation Handlers

func (api *ConfigBuilderAPI) handleGeneratePreview(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Mode string `json:"mode"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update preview mode if provided
	if request.Mode != "" {
		options := config_builder.PreviewOptions{
			Mode:            config_builder.PreviewMode(request.Mode),
			Format:          "nix",
			IncludeComments: true,
		}
		api.preview.UpdateOptions(options)
	}

	result, err := api.preview.GeneratePreview()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate preview: %v", err))
		return
	}

	api.sendJSON(w, result)
}

func (api *ConfigBuilderAPI) handleValidateConfiguration(w http.ResponseWriter, r *http.Request) {
	result, err := api.preview.GeneratePreview()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to validate configuration: %v", err))
		return
	}

	response := map[string]interface{}{
		"valid":    result.Success,
		"errors":   result.Errors,
		"warnings": result.Warnings,
		"metadata": result.Metadata,
	}

	api.sendJSON(w, response)
}

func (api *ConfigBuilderAPI) handleUpdatePreviewOptions(w http.ResponseWriter, r *http.Request) {
	var options config_builder.PreviewOptions

	if err := api.decodeJSON(r, &options); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	api.preview.UpdateOptions(options)
	api.sendJSON(w, map[string]bool{"success": true})
}

// Dependency Analysis Handlers

func (api *ConfigBuilderAPI) handleAnalyzeDependencies(w http.ResponseWriter, r *http.Request) {
	graph, err := api.visualizer.AnalyzeDependencies()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to analyze dependencies: %v", err))
		return
	}

	api.sendJSON(w, graph)
}

func (api *ConfigBuilderAPI) handleGetDependencyGraph(w http.ResponseWriter, r *http.Request) {
	graph, err := api.visualizer.AnalyzeDependencies()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get dependency graph: %v", err))
		return
	}

	api.sendJSON(w, graph)
}

func (api *ConfigBuilderAPI) handleValidateDependencies(w http.ResponseWriter, r *http.Request) {
	issues, err := api.visualizer.ValidateConfiguration()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to validate dependencies: %v", err))
		return
	}

	api.sendJSON(w, issues)
}

func (api *ConfigBuilderAPI) handleGetDependencyReport(w http.ResponseWriter, r *http.Request) {
	report, err := api.visualizer.GenerateDependencyReport()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate dependency report: %v", err))
		return
	}

	api.sendJSON(w, report)
}

func (api *ConfigBuilderAPI) handleExportDependencyGraph(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	data, err := api.visualizer.ExportGraphData(format)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export graph: %v", err))
		return
	}

	// Set appropriate content type
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
	case "dot":
		w.Header().Set("Content-Type", "text/vnd.graphviz")
	case "mermaid":
		w.Header().Set("Content-Type", "text/plain")
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=dependency-graph.%s", format))
	w.Write(data)
}

// Configuration Generation Handlers

func (api *ConfigBuilderAPI) handleGenerateNixOSConfig(w http.ResponseWriter, r *http.Request) {
	config, err := api.dragDrop.GenerateNixConfiguration()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate NixOS configuration: %v", err))
		return
	}

	response := map[string]interface{}{
		"configuration": config,
		"type":          "nixos",
		"filename":      "configuration.nix",
	}

	api.sendJSON(w, response)
}

func (api *ConfigBuilderAPI) handleGenerateHomeManagerConfig(w http.ResponseWriter, r *http.Request) {
	// Update preview mode to Home Manager
	options := config_builder.PreviewOptions{
		Mode:            config_builder.PreviewHomeManager,
		Format:          "nix",
		IncludeComments: true,
	}
	api.preview.UpdateOptions(options)

	result, err := api.preview.GeneratePreview()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate Home Manager configuration: %v", err))
		return
	}

	response := map[string]interface{}{
		"configuration": result.Configuration,
		"type":          "home-manager",
		"filename":      "home.nix",
		"errors":        result.Errors,
		"warnings":      result.Warnings,
	}

	api.sendJSON(w, response)
}

func (api *ConfigBuilderAPI) handleGenerateFlakeConfig(w http.ResponseWriter, r *http.Request) {
	// Update preview mode to Flake
	options := config_builder.PreviewOptions{
		Mode:            config_builder.PreviewFlake,
		Format:          "nix",
		IncludeComments: true,
	}
	api.preview.UpdateOptions(options)

	result, err := api.preview.GeneratePreview()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate Flake configuration: %v", err))
		return
	}

	response := map[string]interface{}{
		"configuration": result.Configuration,
		"type":          "flake",
		"filename":      "flake.nix",
		"errors":        result.Errors,
		"warnings":      result.Warnings,
	}

	api.sendJSON(w, response)
}

func (api *ConfigBuilderAPI) handleGenerateModuleConfig(w http.ResponseWriter, r *http.Request) {
	// Update preview mode to Module
	options := config_builder.PreviewOptions{
		Mode:            config_builder.PreviewModule,
		Format:          "nix",
		IncludeComments: true,
	}
	api.preview.UpdateOptions(options)

	result, err := api.preview.GeneratePreview()
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate Module configuration: %v", err))
		return
	}

	response := map[string]interface{}{
		"configuration": result.Configuration,
		"type":          "module",
		"filename":      "module.nix",
		"errors":        result.Errors,
		"warnings":      result.Warnings,
	}

	api.sendJSON(w, response)
}

// Utility Methods

func (api *ConfigBuilderAPI) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.Error(fmt.Sprintf("Failed to encode JSON response: %v", err))
	}
}

func (api *ConfigBuilderAPI) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	}

	json.NewEncoder(w).Encode(response)
	api.logger.Warn(fmt.Sprintf("API error: %d - %s", statusCode, message))
}

func (api *ConfigBuilderAPI) decodeJSON(r *http.Request, dest interface{}) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

// Cleanup cleans up resources
func (api *ConfigBuilderAPI) Cleanup() error {
	if err := api.preview.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup preview: %w", err)
	}
	return nil
}
