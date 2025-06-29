package config_builder

import (
	"encoding/json"
	"fmt"
	"strings"

	"nix-ai-help/pkg/logger"
)

// DependencyGraph represents the dependency relationships between components
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
	Stats GraphStats  `json:"stats"`
}

// GraphNode represents a component node in the dependency graph
type GraphNode struct {
	ID         string                  `json:"id"`
	Component  *ConfigurationComponent `json:"component"`
	Position   Position                `json:"position"`
	Level      int                     `json:"level"`
	Centrality float64                 `json:"centrality"`
	Critical   bool                    `json:"critical"`
	Metadata   map[string]interface{}  `json:"metadata"`
}

// GraphEdge represents a dependency relationship
type GraphEdge struct {
	ID            string  `json:"id"`
	FromID        string  `json:"from_id"`
	ToID          string  `json:"to_id"`
	Type          string  `json:"type"` // dependency, conflict, requires
	Weight        float64 `json:"weight"`
	Critical      bool    `json:"critical"`
	Bidirectional bool    `json:"bidirectional"`
}

// GraphStats contains statistics about the dependency graph
type GraphStats struct {
	NodeCount          int      `json:"node_count"`
	EdgeCount          int      `json:"edge_count"`
	CyclicDependencies int      `json:"cyclic_dependencies"`
	MaxDepth           int      `json:"max_depth"`
	Complexity         float64  `json:"complexity"`
	CriticalPath       []string `json:"critical_path"`
}

// DependencyVisualizer manages the visualization of component dependencies
type DependencyVisualizer struct {
	canvas  *Canvas
	library *ComponentLibrary
	logger  *logger.Logger
	graph   *DependencyGraph
}

// NewDependencyVisualizer creates a new dependency visualizer
func NewDependencyVisualizer(canvas *Canvas, library *ComponentLibrary, logger *logger.Logger) *DependencyVisualizer {
	return &DependencyVisualizer{
		canvas:  canvas,
		library: library,
		logger:  logger,
		graph:   &DependencyGraph{},
	}
}

// AnalyzeDependencies analyzes the dependencies in the current canvas
func (dv *DependencyVisualizer) AnalyzeDependencies() (*DependencyGraph, error) {
	dv.logger.Debug("Starting dependency analysis")

	// Build graph from canvas components
	if err := dv.buildGraph(); err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Analyze graph structure
	if err := dv.analyzeGraphStructure(); err != nil {
		return nil, fmt.Errorf("failed to analyze graph structure: %w", err)
	}

	// Calculate metrics
	dv.calculateMetrics()

	// Position nodes for visualization
	dv.positionNodes()

	dv.logger.Debug(fmt.Sprintf("Dependency analysis complete: %d nodes, %d edges",
		len(dv.graph.Nodes), len(dv.graph.Edges)))

	return dv.graph, nil
}

// DetectCycles detects circular dependencies in the graph
func (dv *DependencyVisualizer) DetectCycles() ([][]string, error) {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	cycles := [][]string{}

	for _, node := range dv.graph.Nodes {
		if !visited[node.ID] {
			cycle := dv.detectCyclesDFS(node.ID, visited, recursionStack, []string{})
			if len(cycle) > 0 {
				cycles = append(cycles, cycle)
			}
		}
	}

	return cycles, nil
}

// GetCriticalPath finds the critical path through the dependency graph
func (dv *DependencyVisualizer) GetCriticalPath() ([]string, error) {
	if len(dv.graph.Nodes) == 0 {
		return []string{}, nil
	}

	// Find the longest path using topological sort and dynamic programming
	topOrder, err := dv.topologicalSort()
	if err != nil {
		return nil, fmt.Errorf("failed to find topological order: %w", err)
	}

	// Calculate longest distances
	dist := make(map[string]int)
	prev := make(map[string]string)

	for _, nodeID := range topOrder {
		if dist[nodeID] == 0 {
			dist[nodeID] = 1
		}

		for _, edge := range dv.graph.Edges {
			if edge.FromID == nodeID {
				newDist := dist[nodeID] + 1
				if newDist > dist[edge.ToID] {
					dist[edge.ToID] = newDist
					prev[edge.ToID] = nodeID
				}
			}
		}
	}

	// Find node with maximum distance
	maxDist := 0
	endNode := ""
	for nodeID, d := range dist {
		if d > maxDist {
			maxDist = d
			endNode = nodeID
		}
	}

	// Reconstruct path
	path := []string{}
	current := endNode
	for current != "" {
		path = append([]string{current}, path...)
		current = prev[current]
	}

	return path, nil
}

// ValidateConfiguration validates the configuration for dependency issues
func (dv *DependencyVisualizer) ValidateConfiguration() ([]ValidationIssue, error) {
	issues := []ValidationIssue{}

	// Check for circular dependencies
	cycles, err := dv.DetectCycles()
	if err != nil {
		return nil, fmt.Errorf("failed to detect cycles: %w", err)
	}

	for _, cycle := range cycles {
		issues = append(issues, ValidationIssue{
			Type:        "circular_dependency",
			Severity:    "error",
			Message:     fmt.Sprintf("Circular dependency detected: %s", strings.Join(cycle, " → ")),
			Components:  cycle,
			Suggestions: []string{"Remove one of the dependencies to break the cycle"},
		})
	}

	// Check for missing dependencies
	for _, node := range dv.graph.Nodes {
		for _, depID := range node.Component.Dependencies {
			if !dv.hasNode(depID) {
				issues = append(issues, ValidationIssue{
					Type:       "missing_dependency",
					Severity:   "error",
					Message:    fmt.Sprintf("Component %s requires %s which is not included", node.Component.Name, depID),
					Components: []string{node.ID},
					Suggestions: []string{
						fmt.Sprintf("Add component %s to the configuration", depID),
						"Remove the dependency requirement",
					},
				})
			}
		}
	}

	// Check for conflicts
	for _, edge := range dv.graph.Edges {
		if edge.Type == "conflict" {
			fromNode := dv.getNode(edge.FromID)
			toNode := dv.getNode(edge.ToID)
			if fromNode != nil && toNode != nil {
				issues = append(issues, ValidationIssue{
					Type:       "component_conflict",
					Severity:   "warning",
					Message:    fmt.Sprintf("Components %s and %s may conflict", fromNode.Component.Name, toNode.Component.Name),
					Components: []string{edge.FromID, edge.ToID},
					Suggestions: []string{
						"Remove one of the conflicting components",
						"Configure components to avoid conflict",
					},
				})
			}
		}
	}

	// Check for complexity issues
	if dv.graph.Stats.Complexity > 0.8 {
		issues = append(issues, ValidationIssue{
			Type:       "high_complexity",
			Severity:   "warning",
			Message:    fmt.Sprintf("Configuration complexity is high (%.2f)", dv.graph.Stats.Complexity),
			Components: []string{},
			Suggestions: []string{
				"Consider simplifying the configuration",
				"Group related components into modules",
				"Review dependency relationships",
			},
		})
	}

	return issues, nil
}

// GenerateDependencyReport generates a comprehensive dependency report
func (dv *DependencyVisualizer) GenerateDependencyReport() (DependencyReport, error) {
	issues, err := dv.ValidateConfiguration()
	if err != nil {
		return DependencyReport{}, fmt.Errorf("failed to validate configuration: %w", err)
	}

	criticalPath, err := dv.GetCriticalPath()
	if err != nil {
		return DependencyReport{}, fmt.Errorf("failed to get critical path: %w", err)
	}

	cycles, err := dv.DetectCycles()
	if err != nil {
		return DependencyReport{}, fmt.Errorf("failed to detect cycles: %w", err)
	}

	report := DependencyReport{
		Summary: ReportSummary{
			TotalComponents:    len(dv.graph.Nodes),
			TotalDependencies:  len(dv.graph.Edges),
			CriticalComponents: dv.countCriticalComponents(),
			Issues:             len(issues),
			Complexity:         dv.graph.Stats.Complexity,
		},
		CriticalPath:      criticalPath,
		Cycles:            cycles,
		Issues:            issues,
		Recommendations:   dv.generateRecommendations(),
		ComponentAnalysis: dv.analyzeComponents(),
	}

	return report, nil
}

// ExportGraphData exports the dependency graph in various formats
func (dv *DependencyVisualizer) ExportGraphData(format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(dv.graph, "", "  ")
	case "dot":
		return []byte(dv.generateDotFormat()), nil
	case "mermaid":
		return []byte(dv.generateMermaidFormat()), nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// ValidationIssue represents a configuration validation issue
type ValidationIssue struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Message     string   `json:"message"`
	Components  []string `json:"components"`
	Suggestions []string `json:"suggestions"`
}

// DependencyReport contains comprehensive dependency analysis
type DependencyReport struct {
	Summary           ReportSummary                `json:"summary"`
	CriticalPath      []string                     `json:"critical_path"`
	Cycles            [][]string                   `json:"cycles"`
	Issues            []ValidationIssue            `json:"issues"`
	Recommendations   []string                     `json:"recommendations"`
	ComponentAnalysis map[string]ComponentAnalysis `json:"component_analysis"`
}

// ReportSummary contains high-level statistics
type ReportSummary struct {
	TotalComponents    int     `json:"total_components"`
	TotalDependencies  int     `json:"total_dependencies"`
	CriticalComponents int     `json:"critical_components"`
	Issues             int     `json:"issues"`
	Complexity         float64 `json:"complexity"`
}

// ComponentAnalysis contains analysis for individual components
type ComponentAnalysis struct {
	Dependencies    int      `json:"dependencies"`
	Dependents      int      `json:"dependents"`
	Centrality      float64  `json:"centrality"`
	Critical        bool     `json:"critical"`
	Recommendations []string `json:"recommendations"`
}

// Helper methods

func (dv *DependencyVisualizer) buildGraph() error {
	// Clear existing graph
	dv.graph = &DependencyGraph{
		Nodes: []GraphNode{},
		Edges: []GraphEdge{},
		Stats: GraphStats{},
	}

	// Create nodes from canvas components
	for instanceID, placedComp := range dv.canvas.Components {
		node := GraphNode{
			ID:         instanceID,
			Component:  placedComp.Component,
			Position:   placedComp.Position,
			Level:      0,
			Centrality: 0.0,
			Critical:   false,
			Metadata:   make(map[string]interface{}),
		}
		dv.graph.Nodes = append(dv.graph.Nodes, node)
	}

	// Create edges from component dependencies
	for _, node := range dv.graph.Nodes {
		// Add dependency edges
		for _, depID := range node.Component.Dependencies {
			if targetNode := dv.getNodeByComponentID(depID); targetNode != nil {
				edge := GraphEdge{
					ID:            fmt.Sprintf("dep_%s_%s", node.ID, targetNode.ID),
					FromID:        node.ID,
					ToID:          targetNode.ID,
					Type:          "dependency",
					Weight:        1.0,
					Critical:      false,
					Bidirectional: false,
				}
				dv.graph.Edges = append(dv.graph.Edges, edge)
			}
		}

		// Add conflict edges
		for _, conflictID := range node.Component.ConflictsWith {
			if targetNode := dv.getNodeByComponentID(conflictID); targetNode != nil {
				edge := GraphEdge{
					ID:            fmt.Sprintf("conflict_%s_%s", node.ID, targetNode.ID),
					FromID:        node.ID,
					ToID:          targetNode.ID,
					Type:          "conflict",
					Weight:        -1.0,
					Critical:      false,
					Bidirectional: true,
				}
				dv.graph.Edges = append(dv.graph.Edges, edge)
			}
		}

		// Add requires edges
		for _, reqID := range node.Component.Requires {
			if targetNode := dv.getNodeByComponentID(reqID); targetNode != nil {
				edge := GraphEdge{
					ID:            fmt.Sprintf("requires_%s_%s", node.ID, targetNode.ID),
					FromID:        node.ID,
					ToID:          targetNode.ID,
					Type:          "requires",
					Weight:        1.0,
					Critical:      true,
					Bidirectional: false,
				}
				dv.graph.Edges = append(dv.graph.Edges, edge)
			}
		}
	}

	// Add edges from canvas connections
	for _, conn := range dv.canvas.Connections {
		// Check if edge already exists
		exists := false
		for _, edge := range dv.graph.Edges {
			if edge.FromID == conn.FromID && edge.ToID == conn.ToID && edge.Type == conn.Type {
				exists = true
				break
			}
		}

		if !exists {
			edge := GraphEdge{
				ID:            conn.ID,
				FromID:        conn.FromID,
				ToID:          conn.ToID,
				Type:          conn.Type,
				Weight:        1.0,
				Critical:      false,
				Bidirectional: false,
			}
			dv.graph.Edges = append(dv.graph.Edges, edge)
		}
	}

	return nil
}

func (dv *DependencyVisualizer) analyzeGraphStructure() error {
	// Calculate node levels (depth from root nodes)
	dv.calculateNodeLevels()

	// Calculate centrality measures
	dv.calculateCentrality()

	// Identify critical components
	dv.identifyCriticalComponents()

	return nil
}

func (dv *DependencyVisualizer) calculateNodeLevels() {
	// Find root nodes (no incoming dependency edges)
	rootNodes := []string{}
	hasIncoming := make(map[string]bool)

	for _, edge := range dv.graph.Edges {
		if edge.Type == "dependency" || edge.Type == "requires" {
			hasIncoming[edge.ToID] = true
		}
	}

	for _, node := range dv.graph.Nodes {
		if !hasIncoming[node.ID] {
			rootNodes = append(rootNodes, node.ID)
		}
	}

	// BFS to calculate levels
	visited := make(map[string]bool)
	queue := []string{}
	levels := make(map[string]int)

	// Initialize root nodes at level 0
	for _, rootID := range rootNodes {
		queue = append(queue, rootID)
		levels[rootID] = 0
		visited[rootID] = true
	}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		currentLevel := levels[currentID]

		// Find all nodes this one depends on
		for _, edge := range dv.graph.Edges {
			if edge.FromID == currentID && (edge.Type == "dependency" || edge.Type == "requires") {
				if !visited[edge.ToID] {
					visited[edge.ToID] = true
					levels[edge.ToID] = currentLevel + 1
					queue = append(queue, edge.ToID)
				}
			}
		}
	}

	// Update node levels
	for i, node := range dv.graph.Nodes {
		if level, exists := levels[node.ID]; exists {
			dv.graph.Nodes[i].Level = level
		}
	}
}

func (dv *DependencyVisualizer) calculateCentrality() {
	// Simple degree centrality for now
	inDegree := make(map[string]int)
	outDegree := make(map[string]int)

	for _, edge := range dv.graph.Edges {
		outDegree[edge.FromID]++
		inDegree[edge.ToID]++
	}

	totalNodes := float64(len(dv.graph.Nodes))
	for i, node := range dv.graph.Nodes {
		totalDegree := inDegree[node.ID] + outDegree[node.ID]
		dv.graph.Nodes[i].Centrality = float64(totalDegree) / (totalNodes - 1)
	}
}

func (dv *DependencyVisualizer) identifyCriticalComponents() {
	// Mark components as critical based on centrality and dependencies
	for i, node := range dv.graph.Nodes {
		critical := false

		// High centrality indicates importance
		if node.Centrality > 0.5 {
			critical = true
		}

		// Required by many components
		requiredByCount := 0
		for _, edge := range dv.graph.Edges {
			if edge.ToID == node.ID && (edge.Type == "dependency" || edge.Type == "requires") {
				requiredByCount++
			}
		}
		if requiredByCount > 2 {
			critical = true
		}

		dv.graph.Nodes[i].Critical = critical
	}
}

func (dv *DependencyVisualizer) calculateMetrics() {
	stats := &dv.graph.Stats
	stats.NodeCount = len(dv.graph.Nodes)
	stats.EdgeCount = len(dv.graph.Edges)

	// Calculate max depth
	maxLevel := 0
	for _, node := range dv.graph.Nodes {
		if node.Level > maxLevel {
			maxLevel = node.Level
		}
	}
	stats.MaxDepth = maxLevel

	// Calculate complexity (simple metric based on edges/nodes ratio)
	if stats.NodeCount > 0 {
		stats.Complexity = float64(stats.EdgeCount) / float64(stats.NodeCount)
		if stats.Complexity > 1.0 {
			stats.Complexity = 1.0
		}
	}

	// Count cyclic dependencies
	cycles, _ := dv.DetectCycles()
	stats.CyclicDependencies = len(cycles)

	// Get critical path
	criticalPath, _ := dv.GetCriticalPath()
	stats.CriticalPath = criticalPath
}

func (dv *DependencyVisualizer) positionNodes() {
	// Position nodes based on their levels for hierarchical layout
	levelGroups := make(map[int][]int)

	for i, node := range dv.graph.Nodes {
		levelGroups[node.Level] = append(levelGroups[node.Level], i)
	}

	ySpacing := 150.0
	xSpacing := 200.0
	startY := 100.0

	for level, nodeIndices := range levelGroups {
		y := startY + float64(level)*ySpacing
		totalWidth := float64(len(nodeIndices)-1) * xSpacing
		startX := (dv.canvas.Settings.Width - totalWidth) / 2

		for i, nodeIndex := range nodeIndices {
			x := startX + float64(i)*xSpacing
			dv.graph.Nodes[nodeIndex].Position = Position{
				X: x,
				Y: y,
			}
		}
	}
}

func (dv *DependencyVisualizer) detectCyclesDFS(nodeID string, visited, recursionStack map[string]bool, path []string) []string {
	visited[nodeID] = true
	recursionStack[nodeID] = true
	path = append(path, nodeID)

	// Check all adjacent nodes
	for _, edge := range dv.graph.Edges {
		if edge.FromID == nodeID && (edge.Type == "dependency" || edge.Type == "requires") {
			adjID := edge.ToID

			if !visited[adjID] {
				if cycle := dv.detectCyclesDFS(adjID, visited, recursionStack, path); len(cycle) > 0 {
					return cycle
				}
			} else if recursionStack[adjID] {
				// Found cycle - return the cycle portion
				cycleStart := -1
				for i, id := range path {
					if id == adjID {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					return append(path[cycleStart:], adjID)
				}
			}
		}
	}

	recursionStack[nodeID] = false
	return []string{}
}

func (dv *DependencyVisualizer) topologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for _, node := range dv.graph.Nodes {
		inDegree[node.ID] = 0
	}

	for _, edge := range dv.graph.Edges {
		if edge.Type == "dependency" || edge.Type == "requires" {
			inDegree[edge.ToID]++
		}
	}

	queue := []string{}
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	result := []string{}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, edge := range dv.graph.Edges {
			if edge.FromID == current && (edge.Type == "dependency" || edge.Type == "requires") {
				inDegree[edge.ToID]--
				if inDegree[edge.ToID] == 0 {
					queue = append(queue, edge.ToID)
				}
			}
		}
	}

	if len(result) != len(dv.graph.Nodes) {
		return nil, fmt.Errorf("graph contains cycles")
	}

	return result, nil
}

func (dv *DependencyVisualizer) hasNode(componentID string) bool {
	for _, node := range dv.graph.Nodes {
		if node.Component.ID == componentID {
			return true
		}
	}
	return false
}

func (dv *DependencyVisualizer) getNode(nodeID string) *GraphNode {
	for i, node := range dv.graph.Nodes {
		if node.ID == nodeID {
			return &dv.graph.Nodes[i]
		}
	}
	return nil
}

func (dv *DependencyVisualizer) getNodeByComponentID(componentID string) *GraphNode {
	for i, node := range dv.graph.Nodes {
		if node.Component.ID == componentID {
			return &dv.graph.Nodes[i]
		}
	}
	return nil
}

func (dv *DependencyVisualizer) countCriticalComponents() int {
	count := 0
	for _, node := range dv.graph.Nodes {
		if node.Critical {
			count++
		}
	}
	return count
}

func (dv *DependencyVisualizer) generateRecommendations() []string {
	recommendations := []string{}

	// Check complexity
	if dv.graph.Stats.Complexity > 0.7 {
		recommendations = append(recommendations, "Consider reducing configuration complexity by grouping related components")
	}

	// Check for cycles
	if dv.graph.Stats.CyclicDependencies > 0 {
		recommendations = append(recommendations, "Resolve circular dependencies to improve system reliability")
	}

	// Check max depth
	if dv.graph.Stats.MaxDepth > 5 {
		recommendations = append(recommendations, "Deep dependency chains detected - consider flattening the architecture")
	}

	// Check for isolated components
	isolatedCount := 0
	for _, node := range dv.graph.Nodes {
		if node.Centrality == 0 {
			isolatedCount++
		}
	}
	if isolatedCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d isolated components found - verify they are necessary", isolatedCount))
	}

	return recommendations
}

func (dv *DependencyVisualizer) analyzeComponents() map[string]ComponentAnalysis {
	analysis := make(map[string]ComponentAnalysis)

	for _, node := range dv.graph.Nodes {
		deps := 0
		dependents := 0

		for _, edge := range dv.graph.Edges {
			if edge.FromID == node.ID && (edge.Type == "dependency" || edge.Type == "requires") {
				deps++
			}
			if edge.ToID == node.ID && (edge.Type == "dependency" || edge.Type == "requires") {
				dependents++
			}
		}

		recommendations := []string{}
		if node.Centrality > 0.8 {
			recommendations = append(recommendations, "High centrality - consider splitting into smaller components")
		}
		if deps > 5 {
			recommendations = append(recommendations, "Many dependencies - review if all are necessary")
		}
		if dependents > 5 {
			recommendations = append(recommendations, "Many dependents - ensure stable interface")
		}

		analysis[node.ID] = ComponentAnalysis{
			Dependencies:    deps,
			Dependents:      dependents,
			Centrality:      node.Centrality,
			Critical:        node.Critical,
			Recommendations: recommendations,
		}
	}

	return analysis
}

func (dv *DependencyVisualizer) generateDotFormat() string {
	var dot strings.Builder
	dot.WriteString("digraph dependencies {\n")
	dot.WriteString("  rankdir=TB;\n")
	dot.WriteString("  node [shape=box];\n\n")

	// Add nodes
	for _, node := range dv.graph.Nodes {
		color := "lightblue"
		if node.Critical {
			color = "lightcoral"
		}
		dot.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\" fillcolor=\"%s\" style=\"filled\"];\n",
			node.ID, node.Component.Name, color))
	}

	dot.WriteString("\n")

	// Add edges
	for _, edge := range dv.graph.Edges {
		style := "solid"
		color := "black"

		switch edge.Type {
		case "dependency":
			color = "blue"
		case "conflict":
			color = "red"
			style = "dashed"
		case "requires":
			color = "green"
		}

		dot.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"%s\" style=\"%s\" label=\"%s\"];\n",
			edge.FromID, edge.ToID, color, style, edge.Type))
	}

	dot.WriteString("}\n")
	return dot.String()
}

func (dv *DependencyVisualizer) generateMermaidFormat() string {
	var mermaid strings.Builder
	mermaid.WriteString("graph TD\n")

	// Add nodes
	for _, node := range dv.graph.Nodes {
		shape := "[]"
		if node.Critical {
			shape = "(())"
		}
		mermaid.WriteString(fmt.Sprintf("  %s%s%s\n", node.ID, shape[0:1], node.Component.Name+string(shape[1:])))
	}

	mermaid.WriteString("\n")

	// Add edges
	for _, edge := range dv.graph.Edges {
		arrow := "-->"
		if edge.Type == "conflict" {
			arrow = "-..->"
		}

		mermaid.WriteString(fmt.Sprintf("  %s %s %s\n", edge.FromID, arrow, edge.ToID))
	}

	return mermaid.String()
}
