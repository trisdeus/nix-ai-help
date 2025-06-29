package config_builder

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// Canvas represents the visual configuration canvas
type Canvas struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Components  map[string]*PlacedComponent `json:"components"`
	Connections []Connection                `json:"connections"`
	Settings    CanvasSettings              `json:"settings"`
	Metadata    map[string]interface{}      `json:"metadata"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
	Version     string                      `json:"version"`
}

// PlacedComponent represents a component placed on the canvas
type PlacedComponent struct {
	Component *ConfigurationComponent `json:"component"`
	Position  Position                `json:"position"`
	Size      Size                    `json:"size"`
	Selected  bool                    `json:"selected"`
	Locked    bool                    `json:"locked"`
	Config    map[string]interface{}  `json:"config"`
	ZIndex    int                     `json:"z_index"`
}

// CanvasSettings represents canvas-wide settings
type CanvasSettings struct {
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	GridSize   float64 `json:"grid_size"`
	SnapToGrid bool    `json:"snap_to_grid"`
	ShowGrid   bool    `json:"show_grid"`
	ZoomLevel  float64 `json:"zoom_level"`
	Theme      string  `json:"theme"`
	AutoLayout bool    `json:"auto_layout"`
	Readonly   bool    `json:"readonly"`
	Background string  `json:"background"`
}

// DragDropInterface manages the drag-and-drop functionality
type DragDropInterface struct {
	canvas        *Canvas
	library       *ComponentLibrary
	logger        *logger.Logger
	undoStack     []CanvasSnapshot
	redoStack     []CanvasSnapshot
	maxUndoSteps  int
	selectedIDs   []string
	clipboardData []PlacedComponent
	nextZIndex    int
}

// CanvasSnapshot represents a canvas state for undo/redo
type CanvasSnapshot struct {
	Components  map[string]*PlacedComponent `json:"components"`
	Connections []Connection                `json:"connections"`
	Timestamp   time.Time                   `json:"timestamp"`
	Action      string                      `json:"action"`
}

// NewDragDropInterface creates a new drag-drop interface
func NewDragDropInterface(library *ComponentLibrary, logger *logger.Logger) *DragDropInterface {
	canvas := &Canvas{
		ID:          fmt.Sprintf("canvas_%d", time.Now().Unix()),
		Name:        "New Configuration",
		Description: "Visual NixOS configuration",
		Components:  make(map[string]*PlacedComponent),
		Connections: []Connection{},
		Settings: CanvasSettings{
			Width:      1920,
			Height:     1080,
			GridSize:   20,
			SnapToGrid: true,
			ShowGrid:   true,
			ZoomLevel:  1.0,
			Theme:      "default",
			AutoLayout: false,
			Readonly:   false,
			Background: "#f5f5f5",
		},
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}

	return &DragDropInterface{
		canvas:       canvas,
		library:      library,
		logger:       logger,
		undoStack:    []CanvasSnapshot{},
		redoStack:    []CanvasSnapshot{},
		maxUndoSteps: 50,
		selectedIDs:  []string{},
		nextZIndex:   1,
	}
}

// LoadCanvas loads an existing canvas
func (ddi *DragDropInterface) LoadCanvas(canvasData []byte) error {
	var canvas Canvas
	if err := json.Unmarshal(canvasData, &canvas); err != nil {
		return fmt.Errorf("failed to unmarshal canvas: %w", err)
	}

	ddi.canvas = &canvas
	ddi.updateZIndex()

	ddi.logger.Debug(fmt.Sprintf("Loaded canvas %s with %d components", canvas.ID, len(canvas.Components)))
	return nil
}

// SaveCanvas saves the current canvas state
func (ddi *DragDropInterface) SaveCanvas() ([]byte, error) {
	ddi.canvas.UpdatedAt = time.Now()
	return json.MarshalIndent(ddi.canvas, "", "  ")
}

// AddComponent adds a component to the canvas from the library
func (ddi *DragDropInterface) AddComponent(componentID string, position Position) (*PlacedComponent, error) {
	component, err := ddi.library.GetComponent(componentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get component: %w", err)
	}

	// Create unique instance ID
	instanceID := fmt.Sprintf("%s_%d", componentID, time.Now().UnixNano())

	// Snap to grid if enabled
	if ddi.canvas.Settings.SnapToGrid {
		position = ddi.snapToGrid(position)
	}

	placedComponent := &PlacedComponent{
		Component: component,
		Position:  position,
		Size: Size{
			Width:  120,
			Height: 80,
		},
		Selected: false,
		Locked:   false,
		Config:   make(map[string]interface{}),
		ZIndex:   ddi.nextZIndex,
	}

	// Set default configuration values
	for _, option := range component.Options {
		placedComponent.Config[option.Name] = option.DefaultValue
	}

	ddi.takeSnapshot("add_component")
	ddi.canvas.Components[instanceID] = placedComponent
	ddi.nextZIndex++

	ddi.logger.Debug(fmt.Sprintf("Added component %s at position (%.2f, %.2f)", componentID, position.X, position.Y))
	return placedComponent, nil
}

// RemoveComponent removes a component from the canvas
func (ddi *DragDropInterface) RemoveComponent(instanceID string) error {
	if _, exists := ddi.canvas.Components[instanceID]; !exists {
		return fmt.Errorf("component %s not found on canvas", instanceID)
	}

	ddi.takeSnapshot("remove_component")

	// Remove component
	delete(ddi.canvas.Components, instanceID)

	// Remove connections involving this component
	var newConnections []Connection
	for _, conn := range ddi.canvas.Connections {
		if conn.FromID != instanceID && conn.ToID != instanceID {
			newConnections = append(newConnections, conn)
		}
	}
	ddi.canvas.Connections = newConnections

	ddi.logger.Debug(fmt.Sprintf("Removed component %s from canvas", instanceID))
	return nil
}

// MoveComponent moves a component to a new position
func (ddi *DragDropInterface) MoveComponent(instanceID string, newPosition Position) error {
	component, exists := ddi.canvas.Components[instanceID]
	if !exists {
		return fmt.Errorf("component %s not found on canvas", instanceID)
	}

	if component.Locked {
		return fmt.Errorf("component %s is locked", instanceID)
	}

	// Snap to grid if enabled
	if ddi.canvas.Settings.SnapToGrid {
		newPosition = ddi.snapToGrid(newPosition)
	}

	// Check if position actually changed
	if component.Position.X == newPosition.X && component.Position.Y == newPosition.Y {
		return nil
	}

	ddi.takeSnapshot("move_component")
	component.Position = newPosition

	ddi.logger.Debug(fmt.Sprintf("Moved component %s to position (%.2f, %.2f)", instanceID, newPosition.X, newPosition.Y))
	return nil
}

// ResizeComponent resizes a component
func (ddi *DragDropInterface) ResizeComponent(instanceID string, newSize Size) error {
	component, exists := ddi.canvas.Components[instanceID]
	if !exists {
		return fmt.Errorf("component %s not found on canvas", instanceID)
	}

	if component.Locked {
		return fmt.Errorf("component %s is locked", instanceID)
	}

	// Minimum size constraints
	if newSize.Width < 60 {
		newSize.Width = 60
	}
	if newSize.Height < 40 {
		newSize.Height = 40
	}

	ddi.takeSnapshot("resize_component")
	component.Size = newSize

	ddi.logger.Debug(fmt.Sprintf("Resized component %s to size (%.2f, %.2f)", instanceID, newSize.Width, newSize.Height))
	return nil
}

// ConnectComponents creates a connection between two components
func (ddi *DragDropInterface) ConnectComponents(fromID, toID, connectionType string) error {
	fromComponent, fromExists := ddi.canvas.Components[fromID]
	toComponent, toExists := ddi.canvas.Components[toID]

	if !fromExists {
		return fmt.Errorf("source component %s not found", fromID)
	}
	if !toExists {
		return fmt.Errorf("target component %s not found", toID)
	}

	// Check for existing connection
	for _, conn := range ddi.canvas.Connections {
		if conn.FromID == fromID && conn.ToID == toID && conn.Type == connectionType {
			return fmt.Errorf("connection already exists")
		}
	}

	// Validate connection based on component dependencies
	if err := ddi.validateConnection(fromComponent.Component, toComponent.Component, connectionType); err != nil {
		return fmt.Errorf("invalid connection: %w", err)
	}

	connectionID := fmt.Sprintf("conn_%d", time.Now().UnixNano())
	connection := Connection{
		ID:     connectionID,
		FromID: fromID,
		ToID:   toID,
		Type:   connectionType,
		Label:  fmt.Sprintf("%s → %s", fromComponent.Component.Name, toComponent.Component.Name),
		Color:  ddi.getConnectionColor(connectionType),
	}

	ddi.takeSnapshot("connect_components")
	ddi.canvas.Connections = append(ddi.canvas.Connections, connection)

	ddi.logger.Debug(fmt.Sprintf("Connected components %s → %s (%s)", fromID, toID, connectionType))
	return nil
}

// DisconnectComponents removes a connection between components
func (ddi *DragDropInterface) DisconnectComponents(fromID, toID, connectionType string) error {
	var newConnections []Connection
	removed := false

	for _, conn := range ddi.canvas.Connections {
		if conn.FromID == fromID && conn.ToID == toID && conn.Type == connectionType {
			removed = true
			continue
		}
		newConnections = append(newConnections, conn)
	}

	if !removed {
		return fmt.Errorf("connection not found")
	}

	ddi.takeSnapshot("disconnect_components")
	ddi.canvas.Connections = newConnections

	ddi.logger.Debug(fmt.Sprintf("Disconnected components %s → %s (%s)", fromID, toID, connectionType))
	return nil
}

// SelectComponents selects components on the canvas
func (ddi *DragDropInterface) SelectComponents(instanceIDs []string) error {
	// Clear previous selections
	for _, component := range ddi.canvas.Components {
		component.Selected = false
	}

	// Set new selections
	for _, id := range instanceIDs {
		if component, exists := ddi.canvas.Components[id]; exists {
			component.Selected = true
		}
	}

	ddi.selectedIDs = instanceIDs
	ddi.logger.Debug(fmt.Sprintf("Selected %d components", len(instanceIDs)))
	return nil
}

// DuplicateComponents duplicates selected components
func (ddi *DragDropInterface) DuplicateComponents() error {
	if len(ddi.selectedIDs) == 0 {
		return fmt.Errorf("no components selected")
	}

	ddi.takeSnapshot("duplicate_components")

	for _, instanceID := range ddi.selectedIDs {
		component, exists := ddi.canvas.Components[instanceID]
		if !exists {
			continue
		}

		// Create new instance ID
		newInstanceID := fmt.Sprintf("%s_%d", component.Component.ID, time.Now().UnixNano())

		// Create duplicate with offset position
		duplicate := &PlacedComponent{
			Component: component.Component,
			Position: Position{
				X: component.Position.X + 20,
				Y: component.Position.Y + 20,
			},
			Size:     component.Size,
			Selected: false,
			Locked:   false,
			Config:   make(map[string]interface{}),
			ZIndex:   ddi.nextZIndex,
		}

		// Copy configuration
		for k, v := range component.Config {
			duplicate.Config[k] = v
		}

		ddi.canvas.Components[newInstanceID] = duplicate
		ddi.nextZIndex++
	}

	ddi.logger.Debug(fmt.Sprintf("Duplicated %d components", len(ddi.selectedIDs)))
	return nil
}

// AutoLayout automatically arranges components on the canvas
func (ddi *DragDropInterface) AutoLayout(algorithm string) error {
	if len(ddi.canvas.Components) == 0 {
		return fmt.Errorf("no components to layout")
	}

	ddi.takeSnapshot("auto_layout")

	switch algorithm {
	case "grid":
		ddi.layoutGrid()
	case "circular":
		ddi.layoutCircular()
	case "hierarchical":
		ddi.layoutHierarchical()
	case "force_directed":
		ddi.layoutForceDirected()
	default:
		return fmt.Errorf("unknown layout algorithm: %s", algorithm)
	}

	ddi.logger.Debug(fmt.Sprintf("Applied %s layout to %d components", algorithm, len(ddi.canvas.Components)))
	return nil
}

// Undo reverts the last action
func (ddi *DragDropInterface) Undo() error {
	if len(ddi.undoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}

	// Save current state to redo stack
	currentSnapshot := ddi.createSnapshot("undo")
	ddi.redoStack = append(ddi.redoStack, currentSnapshot)

	// Restore previous state
	lastSnapshot := ddi.undoStack[len(ddi.undoStack)-1]
	ddi.undoStack = ddi.undoStack[:len(ddi.undoStack)-1]

	ddi.restoreSnapshot(lastSnapshot)
	ddi.logger.Debug(fmt.Sprintf("Undid action: %s", lastSnapshot.Action))
	return nil
}

// Redo reapplies the last undone action
func (ddi *DragDropInterface) Redo() error {
	if len(ddi.redoStack) == 0 {
		return fmt.Errorf("nothing to redo")
	}

	// Save current state to undo stack
	currentSnapshot := ddi.createSnapshot("redo")
	ddi.undoStack = append(ddi.undoStack, currentSnapshot)

	// Restore next state
	nextSnapshot := ddi.redoStack[len(ddi.redoStack)-1]
	ddi.redoStack = ddi.redoStack[:len(ddi.redoStack)-1]

	ddi.restoreSnapshot(nextSnapshot)
	ddi.logger.Debug(fmt.Sprintf("Redid action: %s", nextSnapshot.Action))
	return nil
}

// GetCanvas returns the current canvas state
func (ddi *DragDropInterface) GetCanvas() *Canvas {
	return ddi.canvas
}

// UpdateCanvasSettings updates canvas settings
func (ddi *DragDropInterface) UpdateCanvasSettings(settings CanvasSettings) {
	ddi.takeSnapshot("update_settings")
	ddi.canvas.Settings = settings
	ddi.logger.Debug("Updated canvas settings")
}

// GenerateNixConfiguration generates NixOS configuration from the canvas
func (ddi *DragDropInterface) GenerateNixConfiguration() (string, error) {
	if len(ddi.canvas.Components) == 0 {
		return "", fmt.Errorf("no components on canvas")
	}

	var config strings.Builder
	config.WriteString("{ config, pkgs, ... }:\n\n")
	config.WriteString("{\n")

	// Generate imports
	imports := ddi.generateImports()
	if len(imports) > 0 {
		config.WriteString("  imports = [\n")
		for _, imp := range imports {
			config.WriteString(fmt.Sprintf("    %s\n", imp))
		}
		config.WriteString("  ];\n\n")
	}

	// Generate service configurations
	services := ddi.generateServiceConfigs()
	if len(services) > 0 {
		config.WriteString("  services = {\n")
		for _, service := range services {
			config.WriteString(fmt.Sprintf("    %s\n", service))
		}
		config.WriteString("  };\n\n")
	}

	// Generate package lists
	packages := ddi.generatePackageList()
	if len(packages) > 0 {
		config.WriteString("  environment.systemPackages = with pkgs; [\n")
		for _, pkg := range packages {
			config.WriteString(fmt.Sprintf("    %s\n", pkg))
		}
		config.WriteString("  ];\n\n")
	}

	// Generate other configurations
	others := ddi.generateOtherConfigs()
	if len(others) > 0 {
		for _, other := range others {
			config.WriteString(fmt.Sprintf("  %s\n", other))
		}
		config.WriteString("\n")
	}

	config.WriteString("}\n")

	return config.String(), nil
}

// Helper methods

func (ddi *DragDropInterface) snapToGrid(position Position) Position {
	gridSize := ddi.canvas.Settings.GridSize
	return Position{
		X: math.Round(position.X/gridSize) * gridSize,
		Y: math.Round(position.Y/gridSize) * gridSize,
	}
}

func (ddi *DragDropInterface) validateConnection(from, to *ConfigurationComponent, connectionType string) error {
	switch connectionType {
	case "dependency":
		// Check if 'to' is in 'from's dependencies
		for _, dep := range from.Dependencies {
			if dep == to.ID {
				return nil
			}
		}
		return fmt.Errorf("component %s does not depend on %s", from.Name, to.Name)
	case "conflict":
		// Check if components conflict
		for _, conflict := range from.ConflictsWith {
			if conflict == to.ID {
				return nil
			}
		}
		return fmt.Errorf("components %s and %s do not conflict", from.Name, to.Name)
	case "requires":
		// Check if 'from' requires 'to'
		for _, req := range from.Requires {
			if req == to.ID {
				return nil
			}
		}
		return fmt.Errorf("component %s does not require %s", from.Name, to.Name)
	}
	return nil
}

func (ddi *DragDropInterface) getConnectionColor(connectionType string) string {
	switch connectionType {
	case "dependency":
		return "#4CAF50"
	case "conflict":
		return "#F44336"
	case "requires":
		return "#2196F3"
	default:
		return "#757575"
	}
}

func (ddi *DragDropInterface) takeSnapshot(action string) {
	snapshot := ddi.createSnapshot(action)
	ddi.undoStack = append(ddi.undoStack, snapshot)

	// Limit undo stack size
	if len(ddi.undoStack) > ddi.maxUndoSteps {
		ddi.undoStack = ddi.undoStack[1:]
	}

	// Clear redo stack
	ddi.redoStack = []CanvasSnapshot{}
}

func (ddi *DragDropInterface) createSnapshot(action string) CanvasSnapshot {
	// Deep copy components
	components := make(map[string]*PlacedComponent)
	for id, comp := range ddi.canvas.Components {
		config := make(map[string]interface{})
		for k, v := range comp.Config {
			config[k] = v
		}

		components[id] = &PlacedComponent{
			Component: comp.Component,
			Position:  comp.Position,
			Size:      comp.Size,
			Selected:  comp.Selected,
			Locked:    comp.Locked,
			Config:    config,
			ZIndex:    comp.ZIndex,
		}
	}

	// Deep copy connections
	connections := make([]Connection, len(ddi.canvas.Connections))
	copy(connections, ddi.canvas.Connections)

	return CanvasSnapshot{
		Components:  components,
		Connections: connections,
		Timestamp:   time.Now(),
		Action:      action,
	}
}

func (ddi *DragDropInterface) restoreSnapshot(snapshot CanvasSnapshot) {
	ddi.canvas.Components = snapshot.Components
	ddi.canvas.Connections = snapshot.Connections
	ddi.updateZIndex()
}

func (ddi *DragDropInterface) updateZIndex() {
	maxZ := 0
	for _, comp := range ddi.canvas.Components {
		if comp.ZIndex > maxZ {
			maxZ = comp.ZIndex
		}
	}
	ddi.nextZIndex = maxZ + 1
}

// Layout algorithms

func (ddi *DragDropInterface) layoutGrid() {
	components := make([]*PlacedComponent, 0, len(ddi.canvas.Components))
	for _, comp := range ddi.canvas.Components {
		components = append(components, comp)
	}

	// Sort components by category and name
	sort.Slice(components, func(i, j int) bool {
		if components[i].Component.Category == components[j].Component.Category {
			return components[i].Component.Name < components[j].Component.Name
		}
		return string(components[i].Component.Category) < string(components[j].Component.Category)
	})

	// Calculate grid dimensions
	cols := int(math.Ceil(math.Sqrt(float64(len(components)))))
	spacing := 200.0
	startX := 100.0
	startY := 100.0

	for i, comp := range components {
		row := i / cols
		col := i % cols

		comp.Position = Position{
			X: startX + float64(col)*spacing,
			Y: startY + float64(row)*spacing,
		}
	}
}

func (ddi *DragDropInterface) layoutCircular() {
	components := make([]*PlacedComponent, 0, len(ddi.canvas.Components))
	for _, comp := range ddi.canvas.Components {
		components = append(components, comp)
	}

	if len(components) == 0 {
		return
	}

	centerX := ddi.canvas.Settings.Width / 2
	centerY := ddi.canvas.Settings.Height / 2
	radius := math.Min(centerX, centerY) * 0.6

	angleStep := 2 * math.Pi / float64(len(components))

	for i, comp := range components {
		angle := float64(i) * angleStep
		comp.Position = Position{
			X: centerX + radius*math.Cos(angle),
			Y: centerY + radius*math.Sin(angle),
		}
	}
}

func (ddi *DragDropInterface) layoutHierarchical() {
	// Simple hierarchical layout based on dependencies
	components := make([]*PlacedComponent, 0, len(ddi.canvas.Components))
	for _, comp := range ddi.canvas.Components {
		components = append(components, comp)
	}

	// TODO: Implement proper hierarchical layout based on dependencies
	// For now, use a simple layered approach
	layers := ddi.calculateLayers(components)

	startY := 100.0
	layerHeight := 150.0

	for layerIndex, layer := range layers {
		y := startY + float64(layerIndex)*layerHeight
		spacing := ddi.canvas.Settings.Width / float64(len(layer)+1)

		for i, comp := range layer {
			comp.Position = Position{
				X: spacing * float64(i+1),
				Y: y,
			}
		}
	}
}

func (ddi *DragDropInterface) layoutForceDirected() {
	// Simple force-directed layout simulation
	components := make([]*PlacedComponent, 0, len(ddi.canvas.Components))
	for _, comp := range ddi.canvas.Components {
		components = append(components, comp)
	}

	if len(components) <= 1 {
		return
	}

	iterations := 100
	k := math.Sqrt(ddi.canvas.Settings.Width * ddi.canvas.Settings.Height / float64(len(components)))

	for iter := 0; iter < iterations; iter++ {
		// Calculate repulsive forces
		for i, comp1 := range components {
			fx, fy := 0.0, 0.0

			for j, comp2 := range components {
				if i == j {
					continue
				}

				dx := comp1.Position.X - comp2.Position.X
				dy := comp1.Position.Y - comp2.Position.Y
				distance := math.Sqrt(dx*dx + dy*dy)

				if distance > 0 {
					repulsive := k * k / distance
					fx += (dx / distance) * repulsive
					fy += (dy / distance) * repulsive
				}
			}

			// Apply forces (with cooling)
			cooling := 1.0 - float64(iter)/float64(iterations)
			comp1.Position.X += fx * cooling * 0.1
			comp1.Position.Y += fy * cooling * 0.1

			// Keep within bounds
			if comp1.Position.X < 50 {
				comp1.Position.X = 50
			}
			if comp1.Position.Y < 50 {
				comp1.Position.Y = 50
			}
			if comp1.Position.X > ddi.canvas.Settings.Width-50 {
				comp1.Position.X = ddi.canvas.Settings.Width - 50
			}
			if comp1.Position.Y > ddi.canvas.Settings.Height-50 {
				comp1.Position.Y = ddi.canvas.Settings.Height - 50
			}
		}
	}
}

func (ddi *DragDropInterface) calculateLayers(components []*PlacedComponent) [][]*PlacedComponent {
	// Simple layer calculation - components with no dependencies go to layer 0
	layers := [][]*PlacedComponent{}

	// For now, just put all components in one layer
	if len(components) > 0 {
		layers = append(layers, components)
	}

	return layers
}

func (ddi *DragDropInterface) generateImports() []string {
	imports := []string{}
	// TODO: Generate imports based on components
	return imports
}

func (ddi *DragDropInterface) generateServiceConfigs() []string {
	configs := []string{}

	for _, placedComp := range ddi.canvas.Components {
		if placedComp.Component.Type == ComponentService {
			config := ddi.generateComponentConfig(placedComp)
			if config != "" {
				configs = append(configs, config)
			}
		}
	}

	return configs
}

func (ddi *DragDropInterface) generatePackageList() []string {
	packages := []string{}

	for _, placedComp := range ddi.canvas.Components {
		if placedComp.Component.Type == ComponentPackage {
			packages = append(packages, placedComp.Component.ID)
		}
	}

	return packages
}

func (ddi *DragDropInterface) generateOtherConfigs() []string {
	configs := []string{}

	for _, placedComp := range ddi.canvas.Components {
		if placedComp.Component.Type != ComponentService && placedComp.Component.Type != ComponentPackage {
			config := ddi.generateComponentConfig(placedComp)
			if config != "" {
				configs = append(configs, config)
			}
		}
	}

	return configs
}

func (ddi *DragDropInterface) generateComponentConfig(placedComp *PlacedComponent) string {
	// Use the component's NixExpression as base and apply user configuration
	config := placedComp.Component.NixExpression

	// TODO: Apply user configuration from placedComp.Config

	return config
}
