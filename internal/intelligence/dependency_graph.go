// Package intelligence provides dependency analysis and relationship mapping for NixOS systems
package intelligence

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// DependencyAnalyzer analyzes and maps dependency relationships in NixOS systems
type DependencyAnalyzer struct {
	logger   *logger.Logger
	analyzer *SystemAnalyzer
	cache    map[string]*DependencyGraph
	mu       sync.RWMutex
}

// DependencyGraph represents the complete dependency relationship graph
type DependencyGraph struct {
	// Core Graph Data
	Nodes map[string]*DependencyNode `json:"nodes"`
	Edges []DependencyEdge           `json:"edges"`

	// Analysis Results
	CriticalPaths []CriticalPath       `json:"critical_paths"`
	CircularDeps  []CircularDependency `json:"circular_dependencies"`
	OrphanedNodes []string             `json:"orphaned_nodes"`
	RootNodes     []string             `json:"root_nodes"`
	LeafNodes     []string             `json:"leaf_nodes"`

	// Metrics
	TotalNodes      int     `json:"total_nodes"`
	TotalEdges      int     `json:"total_edges"`
	MaxDepth        int     `json:"max_depth"`
	AverageDepth    float64 `json:"average_depth"`
	ComplexityScore float64 `json:"complexity_score"`

	// Metadata
	GeneratedAt    time.Time `json:"generated_at"`
	SystemSnapshot string    `json:"system_snapshot"`
}

// DependencyNode represents a single node in the dependency graph
type DependencyNode struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"` // package, service, config, hardware
	Version string `json:"version,omitempty"`
	Status  string `json:"status"` // active, inactive, failed, unknown

	// Relationships
	DirectDependencies   []string `json:"direct_dependencies"`
	DirectDependents     []string `json:"direct_dependents"`
	TransitiveDeps       []string `json:"transitive_dependencies"`
	TransitiveDependents []string `json:"transitive_dependents"`

	// Properties
	Critical    bool      `json:"critical"`
	Optional    bool      `json:"optional"`
	Removable   bool      `json:"removable"`
	Size        int64     `json:"size,omitempty"`
	LastUpdated time.Time `json:"last_updated"`

	// Metadata
	Properties map[string]interface{} `json:"properties"`
	Tags       []string               `json:"tags"`
}

// DependencyEdge represents a relationship between two nodes
type DependencyEdge struct {
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Type        string                 `json:"type"`     // requires, suggests, conflicts, provides
	Strength    float64                `json:"strength"` // 0.0 - 1.0
	Optional    bool                   `json:"optional"`
	Conditional string                 `json:"conditional,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
}

// CriticalPath represents a critical dependency path
type CriticalPath struct {
	Path             []string `json:"path"`
	Length           int      `json:"length"`
	CriticalityScore float64  `json:"criticality_score"`
	Description      string   `json:"description"`
	Impact           []string `json:"impact"`
	Recommendations  []string `json:"recommendations"`
}

// CircularDependency represents a circular dependency
type CircularDependency struct {
	Cycle       []string `json:"cycle"`
	Length      int      `json:"length"`
	Severity    string   `json:"severity"` // critical, high, medium, low
	Impact      []string `json:"impact"`
	Resolution  []string `json:"resolution"`
	Breakpoints []string `json:"breakpoints"` // Suggested points to break the cycle
}

// DependencyAnalysis contains the results of dependency analysis
type DependencyAnalysis struct {
	Graph               *DependencyGraph               `json:"graph"`
	ImpactAnalysis      *ImpactAnalysis                `json:"impact_analysis"`
	Recommendations     []DependencyRecommendation     `json:"recommendations"`
	SecurityAnalysis    *SecurityDependencyAnalysis    `json:"security_analysis"`
	PerformanceAnalysis *PerformanceDependencyAnalysis `json:"performance_analysis"`
}

// ImpactAnalysis analyzes the impact of changes to dependencies
type ImpactAnalysis struct {
	HighImpactNodes []string            `json:"high_impact_nodes"`
	RemovalImpact   map[string][]string `json:"removal_impact"`
	UpdateImpact    map[string][]string `json:"update_impact"`
	CascadeEffects  []CascadeEffect     `json:"cascade_effects"`
}

// CascadeEffect represents potential cascade effects of changes
type CascadeEffect struct {
	TriggerNode   string   `json:"trigger_node"`
	AffectedNodes []string `json:"affected_nodes"`
	EffectType    string   `json:"effect_type"` // removal, update, restart
	Severity      string   `json:"severity"`
	Probability   float64  `json:"probability"`
	Mitigation    []string `json:"mitigation"`
}

// DependencyRecommendation provides actionable dependency recommendations
type DependencyRecommendation struct {
	Type            string   `json:"type"` // optimization, security, maintenance
	Priority        string   `json:"priority"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Actions         []string `json:"actions"`
	Benefits        []string `json:"benefits"`
	Risks           []string `json:"risks"`
	EstimatedEffort string   `json:"estimated_effort"`
}

// SecurityDependencyAnalysis analyzes security aspects of dependencies
type SecurityDependencyAnalysis struct {
	VulnerablePackages []VulnerablePackage `json:"vulnerable_packages"`
	SecurityScore      float64             `json:"security_score"`
	ExposedServices    []string            `json:"exposed_services"`
	TrustChainAnalysis *TrustChainAnalysis `json:"trust_chain_analysis"`
}

// VulnerablePackage represents a package with known vulnerabilities
type VulnerablePackage struct {
	PackageName        string   `json:"package_name"`
	Version            string   `json:"version"`
	Vulnerabilities    []string `json:"vulnerabilities"`
	Severity           string   `json:"severity"`
	FixAvailable       bool     `json:"fix_available"`
	RecommendedVersion string   `json:"recommended_version,omitempty"`
}

// TrustChainAnalysis analyzes the trust chain of dependencies
type TrustChainAnalysis struct {
	TrustedSources  []string `json:"trusted_sources"`
	UnknownSources  []string `json:"unknown_sources"`
	TrustScore      float64  `json:"trust_score"`
	Recommendations []string `json:"recommendations"`
}

// PerformanceDependencyAnalysis analyzes performance impact of dependencies
type PerformanceDependencyAnalysis struct {
	HeavyDependencies       []HeavyDependency        `json:"heavy_dependencies"`
	BootTimeImpact          map[string]time.Duration `json:"boot_time_impact"`
	MemoryImpact            map[string]int64         `json:"memory_impact"`
	OptimizationSuggestions []string                 `json:"optimization_suggestions"`
}

// HeavyDependency represents a dependency with significant resource impact
type HeavyDependency struct {
	PackageName    string        `json:"package_name"`
	Size           int64         `json:"size"`
	MemoryUsage    int64         `json:"memory_usage"`
	BootTimeImpact time.Duration `json:"boot_time_impact"`
	CpuUsage       float64       `json:"cpu_usage"`
	Alternatives   []string      `json:"alternatives"`
}

// NewDependencyAnalyzer creates a new dependency analysis system
func NewDependencyAnalyzer(log *logger.Logger, analyzer *SystemAnalyzer) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		logger:   log,
		analyzer: analyzer,
		cache:    make(map[string]*DependencyGraph),
	}
}

// AnalyzeDependencies performs comprehensive dependency analysis
func (da *DependencyAnalyzer) AnalyzeDependencies(ctx context.Context, userConfig *config.UserConfig) (*DependencyAnalysis, error) {
	da.logger.Info("Starting dependency analysis")
	startTime := time.Now()

	// Get system analysis
	systemAnalysis, err := da.analyzer.AnalyzeSystem(ctx, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get system analysis: %w", err)
	}

	// Build dependency graph
	graph, err := da.buildDependencyGraph(systemAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// Perform various analyses
	analysis := &DependencyAnalysis{
		Graph: graph,
	}

	// Run analyses in parallel
	var wg sync.WaitGroup
	analysisErrors := make(chan error, 4)

	// Impact analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		impactAnalysis, err := da.performImpactAnalysis(graph)
		if err != nil {
			analysisErrors <- fmt.Errorf("impact analysis: %w", err)
			return
		}
		analysis.ImpactAnalysis = impactAnalysis
	}()

	// Security analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		securityAnalysis, err := da.performSecurityAnalysis(graph, systemAnalysis)
		if err != nil {
			analysisErrors <- fmt.Errorf("security analysis: %w", err)
			return
		}
		analysis.SecurityAnalysis = securityAnalysis
	}()

	// Performance analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		performanceAnalysis, err := da.performPerformanceAnalysis(graph, systemAnalysis)
		if err != nil {
			analysisErrors <- fmt.Errorf("performance analysis: %w", err)
			return
		}
		analysis.PerformanceAnalysis = performanceAnalysis
	}()

	// Generate recommendations
	wg.Add(1)
	go func() {
		defer wg.Done()
		recommendations := da.generateRecommendations(graph, systemAnalysis)
		analysis.Recommendations = recommendations
	}()

	// Wait for all analyses
	wg.Wait()
	close(analysisErrors)

	// Check for errors
	for err := range analysisErrors {
		da.logger.Warn(fmt.Sprintf("Dependency analysis warning: %v", err))
	}

	da.logger.Info(fmt.Sprintf("Dependency analysis completed in %v (nodes: %d, edges: %d)",
		time.Since(startTime), graph.TotalNodes, graph.TotalEdges))

	return analysis, nil
}

// buildDependencyGraph constructs the complete dependency graph
func (da *DependencyAnalyzer) buildDependencyGraph(analysis *SystemAnalysis) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes:          make(map[string]*DependencyNode),
		Edges:          []DependencyEdge{},
		GeneratedAt:    time.Now(),
		SystemSnapshot: fmt.Sprintf("%s_%s", analysis.SystemType, analysis.Hostname),
	}

	// Add package nodes
	for _, pkg := range analysis.InstalledPackages {
		node := &DependencyNode{
			ID:                 fmt.Sprintf("pkg_%s", pkg.Name),
			Name:               pkg.Name,
			Type:               "package",
			Version:            pkg.Version,
			Status:             "active",
			DirectDependencies: []string{},
			DirectDependents:   []string{},
			Size:               pkg.Size,
			Properties:         make(map[string]interface{}),
			Tags:               []string{"package"},
		}

		// Add package-specific properties
		node.Properties["description"] = pkg.Purpose
		node.Properties["category"] = pkg.Category
		if pkg.Critical {
			node.Tags = append(node.Tags, "critical")
			node.Critical = true
		}

		graph.Nodes[node.ID] = node
	}

	// Add service nodes
	for _, service := range analysis.EnabledServices {
		node := &DependencyNode{
			ID:                 fmt.Sprintf("svc_%s", service.Name),
			Name:               service.Name,
			Type:               "service",
			Status:             service.Status,
			DirectDependencies: []string{},
			DirectDependents:   []string{},
			Properties:         make(map[string]interface{}),
			Tags:               []string{"service"},
		}

		// Add service-specific properties
		node.Properties["port"] = service.Port
		node.Properties["service_type"] = service.Type
		if service.Security.Sandboxed {
			node.Tags = append(node.Tags, "sandboxed")
		} else {
			node.Tags = append(node.Tags, "unsandboxed")
			node.Critical = true
		}

		// Add dependencies
		for _, dep := range service.Dependencies {
			depID := fmt.Sprintf("svc_%s", dep)
			node.DirectDependencies = append(node.DirectDependencies, depID)

			// Add edge
			graph.Edges = append(graph.Edges, DependencyEdge{
				From:     node.ID,
				To:       depID,
				Type:     "requires",
				Strength: 1.0,
				Optional: false,
			})
		}

		graph.Nodes[node.ID] = node
	}

	// Add configuration nodes
	for _, module := range analysis.ConfigModules {
		node := &DependencyNode{
			ID:                 fmt.Sprintf("cfg_%s", strings.ReplaceAll(module.Name, ".", "_")),
			Name:               module.Name,
			Type:               "config",
			Status:             "active",
			DirectDependencies: []string{},
			DirectDependents:   []string{},
			Properties:         make(map[string]interface{}),
			Tags:               []string{"configuration"},
		}

		graph.Nodes[node.ID] = node
	}

	// Build reverse dependencies
	da.buildReverseDependencies(graph)

	// Detect circular dependencies
	graph.CircularDeps = da.detectCircularDependencies(graph)

	// Find critical paths
	graph.CriticalPaths = da.findCriticalPaths(graph)

	// Identify special nodes
	da.identifySpecialNodes(graph)

	// Calculate metrics
	da.calculateGraphMetrics(graph)

	return graph, nil
}

// buildReverseDependencies builds the reverse dependency relationships
func (da *DependencyAnalyzer) buildReverseDependencies(graph *DependencyGraph) {
	for _, edge := range graph.Edges {
		if _, exists := graph.Nodes[edge.From]; exists {
			if toNode, exists := graph.Nodes[edge.To]; exists {
				toNode.DirectDependents = append(toNode.DirectDependents, edge.From)
			}
		}
	}
}

// detectCircularDependencies detects circular dependencies in the graph
func (da *DependencyAnalyzer) detectCircularDependencies(graph *DependencyGraph) []CircularDependency {
	var cycles []CircularDependency
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for nodeID := range graph.Nodes {
		if !visited[nodeID] {
			if cyclePath := da.findCycleDFS(graph, nodeID, visited, recStack, []string{}); len(cyclePath) > 0 {
				cycles = append(cycles, CircularDependency{
					Cycle:       cyclePath,
					Length:      len(cyclePath),
					Severity:    da.calculateCycleSeverity(cyclePath, graph),
					Impact:      da.calculateCycleImpact(cyclePath, graph),
					Resolution:  da.suggestCycleResolution(cyclePath, graph),
					Breakpoints: da.findCycleBreakpoints(cyclePath, graph),
				})
			}
		}
	}

	return cycles
}

// findCycleDFS performs DFS to find cycles
func (da *DependencyAnalyzer) findCycleDFS(graph *DependencyGraph, nodeID string, visited, recStack map[string]bool, path []string) []string {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)

	if node, exists := graph.Nodes[nodeID]; exists {
		for _, depID := range node.DirectDependencies {
			if !visited[depID] {
				if cycle := da.findCycleDFS(graph, depID, visited, recStack, path); len(cycle) > 0 {
					return cycle
				}
			} else if recStack[depID] {
				// Found a cycle
				cycleStartIndex := -1
				for i, p := range path {
					if p == depID {
						cycleStartIndex = i
						break
					}
				}
				if cycleStartIndex >= 0 {
					return path[cycleStartIndex:]
				}
			}
		}
	}

	recStack[nodeID] = false
	return []string{}
}

// findCriticalPaths identifies critical dependency paths
func (da *DependencyAnalyzer) findCriticalPaths(graph *DependencyGraph) []CriticalPath {
	var criticalPaths []CriticalPath

	// Find paths from root nodes to critical nodes
	for _, rootID := range graph.RootNodes {
		paths := da.findPathsFromRoot(graph, rootID, 5) // Limit depth to 5
		for _, path := range paths {
			if da.isPathCritical(graph, path) {
				criticalPaths = append(criticalPaths, CriticalPath{
					Path:             path,
					Length:           len(path),
					CriticalityScore: da.calculatePathCriticality(graph, path),
					Description:      da.describePathCriticality(graph, path),
					Impact:           da.calculatePathImpact(graph, path),
					Recommendations:  da.suggestPathOptimizations(graph, path),
				})
			}
		}
	}

	// Sort by criticality score
	sort.Slice(criticalPaths, func(i, j int) bool {
		return criticalPaths[i].CriticalityScore > criticalPaths[j].CriticalityScore
	})

	// Return top 10 most critical paths
	if len(criticalPaths) > 10 {
		criticalPaths = criticalPaths[:10]
	}

	return criticalPaths
}

// identifySpecialNodes identifies root, leaf, and orphaned nodes
func (da *DependencyAnalyzer) identifySpecialNodes(graph *DependencyGraph) {
	for nodeID, node := range graph.Nodes {
		// Root nodes (no dependencies)
		if len(node.DirectDependencies) == 0 {
			graph.RootNodes = append(graph.RootNodes, nodeID)
		}

		// Leaf nodes (no dependents)
		if len(node.DirectDependents) == 0 {
			graph.LeafNodes = append(graph.LeafNodes, nodeID)
		}

		// Orphaned nodes (no dependencies and no dependents)
		if len(node.DirectDependencies) == 0 && len(node.DirectDependents) == 0 {
			graph.OrphanedNodes = append(graph.OrphanedNodes, nodeID)
		}
	}
}

// calculateGraphMetrics calculates various graph metrics
func (da *DependencyAnalyzer) calculateGraphMetrics(graph *DependencyGraph) {
	graph.TotalNodes = len(graph.Nodes)
	graph.TotalEdges = len(graph.Edges)

	// Calculate depth metrics
	depths := da.calculateNodeDepths(graph)
	if len(depths) > 0 {
		maxDepth := 0
		totalDepth := 0
		for _, depth := range depths {
			if depth > maxDepth {
				maxDepth = depth
			}
			totalDepth += depth
		}
		graph.MaxDepth = maxDepth
		graph.AverageDepth = float64(totalDepth) / float64(len(depths))
	}

	// Calculate complexity score based on nodes, edges, cycles, and depth
	complexityFactors := []float64{
		float64(graph.TotalNodes) * 0.1,        // Node complexity
		float64(graph.TotalEdges) * 0.15,       // Edge complexity
		float64(len(graph.CircularDeps)) * 2.0, // Circular dependency penalty
		graph.AverageDepth * 0.5,               // Depth complexity
	}

	graph.ComplexityScore = 0
	for _, factor := range complexityFactors {
		graph.ComplexityScore += factor
	}
}

// Helper methods for analysis

func (da *DependencyAnalyzer) calculateNodeDepths(graph *DependencyGraph) map[string]int {
	depths := make(map[string]int)

	// Start with root nodes (depth 0)
	queue := make([]string, 0)
	for _, rootID := range graph.RootNodes {
		depths[rootID] = 0
		queue = append(queue, rootID)
	}

	// BFS to calculate depths
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if node, exists := graph.Nodes[current]; exists {
			currentDepth := depths[current]
			for _, depID := range node.DirectDependencies {
				if _, calculated := depths[depID]; !calculated {
					depths[depID] = currentDepth + 1
					queue = append(queue, depID)
				}
			}
		}
	}

	return depths
}

func (da *DependencyAnalyzer) calculateCycleSeverity(cycle []string, graph *DependencyGraph) string {
	criticalCount := 0
	for _, nodeID := range cycle {
		if node, exists := graph.Nodes[nodeID]; exists && node.Critical {
			criticalCount++
		}
	}

	criticalRatio := float64(criticalCount) / float64(len(cycle))
	switch {
	case criticalRatio > 0.7:
		return "critical"
	case criticalRatio > 0.4:
		return "high"
	case criticalRatio > 0.2:
		return "medium"
	default:
		return "low"
	}
}

func (da *DependencyAnalyzer) calculateCycleImpact(cycle []string, graph *DependencyGraph) []string {
	impact := []string{
		fmt.Sprintf("Circular dependency involving %d components", len(cycle)),
		"May cause installation or update failures",
		"Could lead to system instability",
	}

	// Add specific impacts based on node types
	hasServices := false
	hasPackages := false
	for _, nodeID := range cycle {
		if node, exists := graph.Nodes[nodeID]; exists {
			switch node.Type {
			case "service":
				hasServices = true
			case "package":
				hasPackages = true
			}
		}
	}

	if hasServices {
		impact = append(impact, "Service startup order conflicts")
	}
	if hasPackages {
		impact = append(impact, "Package resolution conflicts")
	}

	return impact
}

func (da *DependencyAnalyzer) suggestCycleResolution(cycle []string, graph *DependencyGraph) []string {
	return []string{
		"Identify optional dependencies that can be removed",
		"Consider using conditional dependencies",
		"Refactor components to eliminate circular references",
		"Use dependency injection or service locator patterns",
	}
}

func (da *DependencyAnalyzer) findCycleBreakpoints(cycle []string, graph *DependencyGraph) []string {
	var breakpoints []string

	// Find edges in the cycle that could be made optional
	for i := 0; i < len(cycle); i++ {
		fromID := cycle[i]
		toID := cycle[(i+1)%len(cycle)]

		// Check if this edge could be optional
		if da.canMakeOptional(fromID, toID, graph) {
			breakpoints = append(breakpoints, fmt.Sprintf("%s -> %s", fromID, toID))
		}
	}

	return breakpoints
}

func (da *DependencyAnalyzer) canMakeOptional(fromID, toID string, graph *DependencyGraph) bool {
	// Simple heuristic: service dependencies are often more flexible
	fromNode, fromExists := graph.Nodes[fromID]
	toNode, toExists := graph.Nodes[toID]

	if fromExists && toExists {
		return fromNode.Type == "service" || toNode.Type == "service"
	}

	return false
}

func (da *DependencyAnalyzer) findPathsFromRoot(graph *DependencyGraph, rootID string, maxDepth int) [][]string {
	var paths [][]string
	da.findPathsRecursive(graph, rootID, []string{}, maxDepth, &paths)
	return paths
}

func (da *DependencyAnalyzer) findPathsRecursive(graph *DependencyGraph, nodeID string, currentPath []string, remainingDepth int, allPaths *[][]string) {
	if remainingDepth <= 0 {
		return
	}

	// Avoid cycles in path finding
	for _, pathNode := range currentPath {
		if pathNode == nodeID {
			return
		}
	}

	newPath := append(currentPath, nodeID)

	if node, exists := graph.Nodes[nodeID]; exists {
		if len(node.DirectDependencies) == 0 {
			// Leaf node, add path
			*allPaths = append(*allPaths, newPath)
		} else {
			// Continue to dependencies
			for _, depID := range node.DirectDependencies {
				da.findPathsRecursive(graph, depID, newPath, remainingDepth-1, allPaths)
			}
		}
	}
}

func (da *DependencyAnalyzer) isPathCritical(graph *DependencyGraph, path []string) bool {
	criticalCount := 0
	for _, nodeID := range path {
		if node, exists := graph.Nodes[nodeID]; exists && node.Critical {
			criticalCount++
		}
	}
	return float64(criticalCount)/float64(len(path)) > 0.3 // 30% critical nodes
}

func (da *DependencyAnalyzer) calculatePathCriticality(graph *DependencyGraph, path []string) float64 {
	score := 0.0
	for _, nodeID := range path {
		if node, exists := graph.Nodes[nodeID]; exists {
			if node.Critical {
				score += 2.0
			}
			score += float64(len(node.DirectDependents)) * 0.1 // More dependents = more critical
		}
	}
	return score / float64(len(path))
}

func (da *DependencyAnalyzer) describePathCriticality(graph *DependencyGraph, path []string) string {
	if len(path) == 0 {
		return "Empty path"
	}

	startNode, _ := graph.Nodes[path[0]]
	endNode, _ := graph.Nodes[path[len(path)-1]]

	return fmt.Sprintf("Critical path from %s (%s) to %s (%s) with %d components",
		startNode.Name, startNode.Type, endNode.Name, endNode.Type, len(path))
}

func (da *DependencyAnalyzer) calculatePathImpact(graph *DependencyGraph, path []string) []string {
	impact := []string{
		fmt.Sprintf("Path involves %d critical components", len(path)),
	}

	totalDependents := 0
	for _, nodeID := range path {
		if node, exists := graph.Nodes[nodeID]; exists {
			totalDependents += len(node.DirectDependents)
		}
	}

	if totalDependents > 0 {
		impact = append(impact, fmt.Sprintf("Affects %d dependent components", totalDependents))
	}

	return impact
}

func (da *DependencyAnalyzer) suggestPathOptimizations(graph *DependencyGraph, path []string) []string {
	recommendations := []string{
		"Monitor this critical path for changes",
		"Consider reducing dependency depth",
	}

	if len(path) > 5 {
		recommendations = append(recommendations, "Path is long - consider architectural refactoring")
	}

	return recommendations
}

// performImpactAnalysis analyzes the impact of potential changes
func (da *DependencyAnalyzer) performImpactAnalysis(graph *DependencyGraph) (*ImpactAnalysis, error) {
	analysis := &ImpactAnalysis{
		RemovalImpact: make(map[string][]string),
		UpdateImpact:  make(map[string][]string),
	}

	// Identify high-impact nodes
	for nodeID, node := range graph.Nodes {
		dependentCount := len(node.DirectDependents) + len(node.TransitiveDependents)
		if dependentCount > 5 || node.Critical {
			analysis.HighImpactNodes = append(analysis.HighImpactNodes, nodeID)
		}

		// Calculate removal impact
		analysis.RemovalImpact[nodeID] = node.DirectDependents

		// Calculate update impact (similar to removal for now)
		analysis.UpdateImpact[nodeID] = node.DirectDependents
	}

	// Analyze cascade effects
	analysis.CascadeEffects = da.analyzeCascadeEffects(graph)

	return analysis, nil
}

func (da *DependencyAnalyzer) analyzeCascadeEffects(graph *DependencyGraph) []CascadeEffect {
	var effects []CascadeEffect

	for nodeID, node := range graph.Nodes {
		if len(node.DirectDependents) > 3 { // Nodes with many dependents
			effects = append(effects, CascadeEffect{
				TriggerNode:   nodeID,
				AffectedNodes: node.DirectDependents,
				EffectType:    "removal",
				Severity:      da.calculateCascadeSeverity(len(node.DirectDependents)),
				Probability:   0.8, // High probability for direct dependents
				Mitigation:    []string{"Test in staging environment", "Plan gradual rollout"},
			})
		}
	}

	return effects
}

func (da *DependencyAnalyzer) calculateCascadeSeverity(dependentCount int) string {
	switch {
	case dependentCount > 10:
		return "critical"
	case dependentCount > 5:
		return "high"
	case dependentCount > 2:
		return "medium"
	default:
		return "low"
	}
}

// performSecurityAnalysis analyzes security aspects of dependencies
func (da *DependencyAnalyzer) performSecurityAnalysis(graph *DependencyGraph, systemAnalysis *SystemAnalysis) (*SecurityDependencyAnalysis, error) {
	analysis := &SecurityDependencyAnalysis{
		VulnerablePackages: []VulnerablePackage{},
		ExposedServices:    []string{},
	}

	// Find vulnerable packages
	// TODO: Implement vulnerability detection when Security field is available
	/*
		for _, pkg := range systemAnalysis.InstalledPackages {
			if pkg.Security.HasVulnerabilities {
				vulnPkg := VulnerablePackage{
					PackageName:     pkg.Name,
					Version:         pkg.Version,
					Vulnerabilities: pkg.Security.KnownVulnerabilities,
					Severity:        pkg.Security.SecurityLevel,
					FixAvailable:    len(pkg.Security.SecurityPatches) > 0,
				}
				analysis.VulnerablePackages = append(analysis.VulnerablePackages, vulnPkg)
			}
		}
	*/

	// Find exposed services
	for _, service := range systemAnalysis.EnabledServices {
		if service.Port > 0 && !service.Security.Sandboxed {
			analysis.ExposedServices = append(analysis.ExposedServices, service.Name)
		}
	}

	// Calculate security score
	analysis.SecurityScore = da.calculateSecurityScore(analysis, graph)

	// Analyze trust chain
	analysis.TrustChainAnalysis = da.analyzeTrustChain(systemAnalysis)

	return analysis, nil
}

func (da *DependencyAnalyzer) calculateSecurityScore(analysis *SecurityDependencyAnalysis, graph *DependencyGraph) float64 {
	baseScore := 100.0

	// Deduct for vulnerabilities
	for _, vuln := range analysis.VulnerablePackages {
		switch vuln.Severity {
		case "critical":
			baseScore -= 15.0
		case "high":
			baseScore -= 10.0
		case "medium":
			baseScore -= 5.0
		case "low":
			baseScore -= 2.0
		}
	}

	// Deduct for exposed services
	baseScore -= float64(len(analysis.ExposedServices)) * 3.0

	if baseScore < 0 {
		baseScore = 0
	}

	return baseScore
}

func (da *DependencyAnalyzer) analyzeTrustChain(systemAnalysis *SystemAnalysis) *TrustChainAnalysis {
	trusted := []string{"nixpkgs", "official", "nixos"}
	unknown := []string{}

	// Simple trust analysis based on package sources
	// TODO: Implement repository field when available
	/*
		for _, pkg := range systemAnalysis.InstalledPackages {
			isTrusted := false
			for _, trustedSource := range trusted {
				if strings.Contains(strings.ToLower(pkg.Repository), trustedSource) {
					isTrusted = true
					break
				}
			}
			if !isTrusted && pkg.Repository != "" {
				unknown = append(unknown, pkg.Repository)
			}
		}
	*/

	// Remove duplicates
	unknown = da.removeDuplicates(unknown)

	trustScore := 100.0
	if len(unknown) > 0 {
		trustScore -= float64(len(unknown)) * 5.0
	}

	return &TrustChainAnalysis{
		TrustedSources: trusted,
		UnknownSources: unknown,
		TrustScore:     trustScore,
		Recommendations: []string{
			"Review packages from unknown sources",
			"Prefer official NixOS packages when available",
			"Regularly audit package sources",
		},
	}
}

// performPerformanceAnalysis analyzes performance impact of dependencies
func (da *DependencyAnalyzer) performPerformanceAnalysis(graph *DependencyGraph, systemAnalysis *SystemAnalysis) (*PerformanceDependencyAnalysis, error) {
	analysis := &PerformanceDependencyAnalysis{
		HeavyDependencies:       []HeavyDependency{},
		BootTimeImpact:          make(map[string]time.Duration),
		MemoryImpact:            make(map[string]int64),
		OptimizationSuggestions: []string{},
	}

	// Find heavy dependencies
	for _, pkg := range systemAnalysis.InstalledPackages {
		if pkg.Size > 100*1024*1024 { // > 100MB
			heavy := HeavyDependency{
				PackageName:    pkg.Name,
				Size:           pkg.Size,
				MemoryUsage:    pkg.Size / 10, // Rough estimate
				BootTimeImpact: time.Duration(pkg.Size/1024/1024) * time.Millisecond,
				CpuUsage:       0.0, // Would need runtime analysis
			}
			analysis.HeavyDependencies = append(analysis.HeavyDependencies, heavy)
		}
	}

	// Generate optimization suggestions
	if len(analysis.HeavyDependencies) > 0 {
		analysis.OptimizationSuggestions = append(analysis.OptimizationSuggestions,
			"Consider lighter alternatives for heavy packages",
			"Use package overlays to optimize builds",
			"Enable package compression where possible")
	}

	return analysis, nil
}

// generateRecommendations generates actionable recommendations
func (da *DependencyAnalyzer) generateRecommendations(graph *DependencyGraph, systemAnalysis *SystemAnalysis) []DependencyRecommendation {
	var recommendations []DependencyRecommendation

	// Circular dependency recommendations
	if len(graph.CircularDeps) > 0 {
		recommendations = append(recommendations, DependencyRecommendation{
			Type:        "maintenance",
			Priority:    "high",
			Title:       "Resolve Circular Dependencies",
			Description: fmt.Sprintf("Found %d circular dependencies that may cause system issues", len(graph.CircularDeps)),
			Actions: []string{
				"Review circular dependency cycles",
				"Refactor components to eliminate cycles",
				"Consider using dependency injection",
			},
			Benefits:        []string{"Improved system stability", "Easier maintenance", "Better performance"},
			Risks:           []string{"Requires architectural changes"},
			EstimatedEffort: "high",
		})
	}

	// Orphaned packages recommendation
	if len(graph.OrphanedNodes) > 0 {
		recommendations = append(recommendations, DependencyRecommendation{
			Type:        "optimization",
			Priority:    "medium",
			Title:       "Remove Orphaned Packages",
			Description: fmt.Sprintf("Found %d orphaned packages that can be safely removed", len(graph.OrphanedNodes)),
			Actions: []string{
				"Review orphaned packages",
				"Remove unused packages",
				"Run nix-collect-garbage",
			},
			Benefits:        []string{"Reduced disk usage", "Simplified system", "Faster updates"},
			Risks:           []string{"Minimal risk - packages have no dependents"},
			EstimatedEffort: "low",
		})
	}

	// High complexity recommendation
	if graph.ComplexityScore > 50.0 {
		recommendations = append(recommendations, DependencyRecommendation{
			Type:        "optimization",
			Priority:    "medium",
			Title:       "Reduce Dependency Complexity",
			Description: fmt.Sprintf("System complexity score is %.1f, consider simplification", graph.ComplexityScore),
			Actions: []string{
				"Review dependency graph",
				"Consolidate similar packages",
				"Remove unnecessary dependencies",
			},
			Benefits:        []string{"Easier maintenance", "Better performance", "Reduced conflicts"},
			Risks:           []string{"May require configuration changes"},
			EstimatedEffort: "medium",
		})
	}

	return recommendations
}

// Utility methods

func (da *DependencyAnalyzer) removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ClearCache clears the dependency analysis cache
func (da *DependencyAnalyzer) ClearCache() {
	da.mu.Lock()
	defer da.mu.Unlock()

	da.cache = make(map[string]*DependencyGraph)
	da.logger.Info("Dependency analysis cache cleared")
}
