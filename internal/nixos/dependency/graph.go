// graph.go - Dependency graph generation and analysis
package dependency

import (
	"fmt"
	"sort"
	"strings"

	"nix-ai-help/internal/hardware"
)

// buildDependencyGraph creates a visual dependency graph
func (da *DependencyAnalyzer) buildDependencyGraph(configOptions []*ConfigOption, dependencies []*Dependency) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: []*GraphNode{},
		Edges: []*GraphEdge{},
		Cycles: [][]string{},
	}

	// Create nodes for all configuration options
	nodeMap := make(map[string]*GraphNode)
	for _, option := range configOptions {
		node := &GraphNode{
			ID:       option.Name,
			Name:     option.Name,
			Type:     option.Type,
			Category: option.Category,
			Required: option.Required,
			Attributes: map[string]interface{}{
				"value":       option.Value,
				"module":      option.Module,
				"description": option.Description,
			},
		}
		
		graph.Nodes = append(graph.Nodes, node)
		nodeMap[option.Name] = node
	}

	// Add nodes for dependencies that don't exist in configuration
	for _, dep := range dependencies {
		if _, exists := nodeMap[dep.To]; !exists {
			node := &GraphNode{
				ID:       dep.To,
				Name:     dep.To,
				Type:     "missing",
				Category: da.categorizeOption(dep.To),
				Required: dep.Type == DependencyRequired,
				Attributes: map[string]interface{}{
					"missing":     true,
					"description": da.getOptionDescription(dep.To),
				},
			}
			
			graph.Nodes = append(graph.Nodes, node)
			nodeMap[dep.To] = node
		}
	}

	// Create edges for dependencies
	for _, dep := range dependencies {
		edge := &GraphEdge{
			From:   dep.From,
			To:     dep.To,
			Type:   dep.Type,
			Weight: dep.Strength,
			Label:  string(dep.Type),
		}
		
		graph.Edges = append(graph.Edges, edge)
	}

	// Calculate dependency levels
	da.calculateDependencyLevels(graph, nodeMap)

	// Detect cycles
	graph.Cycles = da.detectCycles(graph)

	return graph
}

// calculateDependencyLevels assigns levels to nodes based on dependency depth
func (da *DependencyAnalyzer) calculateDependencyLevels(graph *DependencyGraph, nodeMap map[string]*GraphNode) {
	// Create adjacency list for topological sorting
	adjList := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize
	for _, node := range graph.Nodes {
		adjList[node.ID] = []string{}
		inDegree[node.ID] = 0
	}

	// Build adjacency list and calculate in-degrees
	for _, edge := range graph.Edges {
		if edge.Type == DependencyRequired || edge.Type == DependencyRecommended {
			adjList[edge.From] = append(adjList[edge.From], edge.To)
			inDegree[edge.To]++
		}
	}

	// Topological sort to assign levels
	queue := []string{}
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
			nodeMap[nodeID].Level = 0
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
				nodeMap[neighbor].Level = nodeMap[current].Level + 1
			}
		}
	}
}

// detectCycles detects circular dependencies in the graph
func (da *DependencyAnalyzer) detectCycles(graph *DependencyGraph) [][]string {
	var cycles [][]string

	// Build adjacency list
	adjList := make(map[string][]string)
	for _, node := range graph.Nodes {
		adjList[node.ID] = []string{}
	}

	for _, edge := range graph.Edges {
		adjList[edge.From] = append(adjList[edge.From], edge.To)
	}

	// DFS-based cycle detection
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	for _, node := range graph.Nodes {
		if !visited[node.ID] {
			da.dfsDetectCycle(node.ID, adjList, visited, recStack, path, &cycles)
		}
	}

	return cycles
}

// dfsDetectCycle performs DFS to detect cycles
func (da *DependencyAnalyzer) dfsDetectCycle(nodeID string, adjList map[string][]string, 
	visited, recStack map[string]bool, path []string, cycles *[][]string) {
	
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)

	for _, neighbor := range adjList[nodeID] {
		if !visited[neighbor] {
			da.dfsDetectCycle(neighbor, adjList, visited, recStack, path, cycles)
		} else if recStack[neighbor] {
			// Found a cycle
			cycleStart := -1
			for i, node := range path {
				if node == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart != -1 {
				cycle := append(path[cycleStart:], neighbor)
				*cycles = append(*cycles, cycle)
			}
		}
	}

	recStack[nodeID] = false
	path = path[:len(path)-1]
}

// generateRecommendations generates configuration recommendations
func (da *DependencyAnalyzer) generateRecommendations(configOptions []*ConfigOption, dependencies []*Dependency, 
	conflicts []*Conflict, hardwareInfo *hardware.EnhancedHardwareInfo) []*Recommendation {
	
	var recommendations []*Recommendation

	// Configuration map for quick lookup
	configMap := make(map[string]*ConfigOption)
	for _, opt := range configOptions {
		configMap[opt.Name] = opt
	}

	// Recommendations based on missing dependencies
	for _, dep := range dependencies {
		if dep.Type == DependencyRequired {
			if _, exists := configMap[dep.To]; !exists {
				rec := &Recommendation{
					Type:        RecommendationCompatibility,
					Priority:    9,
					Option:      dep.To,
					Action:      "add",
					Value:       da.getRecommendedValue(dep.To),
					Reason:      fmt.Sprintf("Required by %s", dep.From),
					Benefits:    []string{"Ensures proper functionality", "Prevents configuration errors"},
					HardwareBased: false,
				}
				recommendations = append(recommendations, rec)
			}
		} else if dep.Type == DependencyRecommended {
			if _, exists := configMap[dep.To]; !exists {
				rec := &Recommendation{
					Type:        RecommendationOptimization,
					Priority:    6,
					Option:      dep.To,
					Action:      "add",
					Value:       da.getRecommendedValue(dep.To),
					Reason:      fmt.Sprintf("Recommended for %s", dep.From),
					Benefits:    []string{"Improves functionality", "Better integration"},
					HardwareBased: false,
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// Recommendations based on conflicts
	for _, conflict := range conflicts {
		if conflict.Severity == "critical" && len(conflict.Options) >= 2 {
			// Recommend disabling one of the conflicting options
			rec := &Recommendation{
				Type:        RecommendationCompatibility,
				Priority:    10,
				Option:      conflict.Options[1], // Disable the second option
				Action:      "remove",
				Reason:      fmt.Sprintf("Conflicts with %s", conflict.Options[0]),
				Benefits:    []string{"Resolves configuration conflict", "Prevents system errors"},
				Risks:       []string{"May disable functionality"},
				HardwareBased: false,
			}
			recommendations = append(recommendations, rec)
		}
	}

	// Hardware-based recommendations
	if hardwareInfo != nil {
		recommendations = append(recommendations, da.generateHardwareRecommendations(configOptions, hardwareInfo)...)
	}

	// Performance recommendations
	recommendations = append(recommendations, da.generatePerformanceRecommendations(configOptions, hardwareInfo)...)

	// Security recommendations
	recommendations = append(recommendations, da.generateSecurityRecommendations(configOptions)...)

	// Sort recommendations by priority
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	return recommendations
}

// generateHardwareRecommendations generates hardware-specific recommendations
func (da *DependencyAnalyzer) generateHardwareRecommendations(configOptions []*ConfigOption, hardwareInfo *hardware.EnhancedHardwareInfo) []*Recommendation {
	var recommendations []*Recommendation

	if hardwareInfo == nil || hardwareInfo.SystemProfile == nil {
		return recommendations
	}

	configMap := make(map[string]*ConfigOption)
	for _, opt := range configOptions {
		configMap[opt.Name] = opt
	}

	// GPU-specific recommendations
	for _, gpu := range hardwareInfo.SystemProfile.GPUDetails {
		if gpu.Vendor == "NVIDIA" {
			if _, exists := configMap["services.xserver.videoDrivers"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationHardware,
					Priority:    8,
					Option:      "services.xserver.videoDrivers",
					Action:      "add",
					Value:       []string{"nvidia"},
					Reason:      "NVIDIA GPU detected",
					Benefits:    []string{"Hardware acceleration", "Proper graphics support", "CUDA support"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}

			if _, exists := configMap["hardware.opengl.enable"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationHardware,
					Priority:    8,
					Option:      "hardware.opengl.enable",
					Action:      "add",
					Value:       true,
					Reason:      "Required for NVIDIA graphics",
					Benefits:    []string{"OpenGL acceleration", "Better graphics performance"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}
		} else if gpu.Vendor == "AMD" {
			if _, exists := configMap["services.xserver.videoDrivers"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationHardware,
					Priority:    7,
					Option:      "services.xserver.videoDrivers",
					Action:      "add",
					Value:       []string{"amdgpu"},
					Reason:      "AMD GPU detected",
					Benefits:    []string{"Open source drivers", "Good performance", "Vulkan support"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	// CPU microcode recommendations
	if hardwareInfo.SystemProfile.CPUDetails != nil {
		cpu := hardwareInfo.SystemProfile.CPUDetails
		if cpu.Vendor == "AuthenticAMD" || cpu.Vendor == "AMD" {
			if _, exists := configMap["hardware.cpu.amd.updateMicrocode"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationSecurity,
					Priority:    7,
					Option:      "hardware.cpu.amd.updateMicrocode",
					Action:      "add",
					Value:       true,
					Reason:      "AMD CPU detected",
					Benefits:    []string{"Security updates", "Stability improvements", "Performance fixes"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}
		} else if cpu.Vendor == "GenuineIntel" || cpu.Vendor == "Intel" {
			if _, exists := configMap["hardware.cpu.intel.updateMicrocode"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationSecurity,
					Priority:    7,
					Option:      "hardware.cpu.intel.updateMicrocode",
					Action:      "add",
					Value:       true,
					Reason:      "Intel CPU detected",
					Benefits:    []string{"Security patches", "Stability improvements", "Performance optimizations"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	return recommendations
}

// generatePerformanceRecommendations generates performance-related recommendations
func (da *DependencyAnalyzer) generatePerformanceRecommendations(configOptions []*ConfigOption, hardwareInfo *hardware.EnhancedHardwareInfo) []*Recommendation {
	var recommendations []*Recommendation

	configMap := make(map[string]*ConfigOption)
	for _, opt := range configOptions {
		configMap[opt.Name] = opt
	}

	// Multi-core build optimization
	if hardwareInfo != nil && hardwareInfo.SystemProfile != nil && hardwareInfo.SystemProfile.CPUDetails != nil {
		cores := hardwareInfo.SystemProfile.CPUDetails.Cores
		if cores > 4 {
			if _, exists := configMap["nix.settings.max-jobs"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationPerformance,
					Priority:    5,
					Option:      "nix.settings.max-jobs",
					Action:      "add",
					Value:       cores,
					Reason:      fmt.Sprintf("Multi-core CPU detected (%d cores)", cores),
					Benefits:    []string{"Faster builds", "Better resource utilization"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}

			if _, exists := configMap["nix.settings.cores"]; !exists {
				rec := &Recommendation{
					Type:        RecommendationPerformance,
					Priority:    5,
					Option:      "nix.settings.cores",
					Action:      "add",
					Value:       cores,
					Reason:      "Utilize all CPU cores for compilation",
					Benefits:    []string{"Parallel compilation", "Reduced build times"},
					HardwareBased: true,
				}
				recommendations = append(recommendations, rec)
			}
		}
	}

	return recommendations
}

// generateSecurityRecommendations generates security-related recommendations
func (da *DependencyAnalyzer) generateSecurityRecommendations(configOptions []*ConfigOption) []*Recommendation {
	var recommendations []*Recommendation

	configMap := make(map[string]*ConfigOption)
	for _, opt := range configOptions {
		configMap[opt.Name] = opt
	}

	// Firewall recommendation
	if _, exists := configMap["networking.firewall.enable"]; !exists {
		rec := &Recommendation{
			Type:        RecommendationSecurity,
			Priority:    6,
			Option:      "networking.firewall.enable",
			Action:      "add",
			Value:       true,
			Reason:      "Basic security hardening",
			Benefits:    []string{"Network protection", "Blocks unwanted connections", "Default security"},
			HardwareBased: false,
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// validateConfiguration performs comprehensive configuration validation
func (da *DependencyAnalyzer) validateConfiguration(configOptions []*ConfigOption, dependencies []*Dependency, conflicts []*Conflict) *ValidationResults {
	// Use rule engine for basic validation
	results := da.ruleEngine.ValidateConfiguration(configOptions)

	// Add additional validation logic
	da.validateDependencyConsistency(results, dependencies)
	da.validateConflictSeverity(results, conflicts)
	da.generateOptimizationTips(results, configOptions)

	return results
}

// validateDependencyConsistency checks if dependencies are properly satisfied
func (da *DependencyAnalyzer) validateDependencyConsistency(results *ValidationResults, dependencies []*Dependency) {
	for _, dep := range dependencies {
		if dep.Type == DependencyRequired && !dep.AutoResolve {
			results.Suggestions = append(results.Suggestions, 
				fmt.Sprintf("Consider adding %s (required by %s)", dep.To, dep.From))
		}
	}
}

// validateConflictSeverity adjusts validation based on conflict severity
func (da *DependencyAnalyzer) validateConflictSeverity(results *ValidationResults, conflicts []*Conflict) {
	criticalConflicts := 0
	for _, conflict := range conflicts {
		if conflict.Severity == "critical" {
			criticalConflicts++
		}
	}

	if criticalConflicts > 0 {
		results.Score *= 0.5 // Severely penalize critical conflicts
		results.Suggestions = append(results.Suggestions, 
			fmt.Sprintf("Resolve %d critical configuration conflicts", criticalConflicts))
	}
}

// generateOptimizationTips generates configuration optimization suggestions
func (da *DependencyAnalyzer) generateOptimizationTips(results *ValidationResults, configOptions []*ConfigOption) {
	// Count options by category
	categoryCount := make(map[string]int)
	for _, opt := range configOptions {
		categoryCount[opt.Category]++
	}

	// Generate tips based on configuration complexity
	if categoryCount["services"] > 10 {
		results.OptimizationTips = append(results.OptimizationTips, 
			"Consider organizing services into separate configuration modules")
	}

	if categoryCount["hardware"] < 3 {
		results.OptimizationTips = append(results.OptimizationTips, 
			"Consider adding hardware-specific optimizations")
	}

	if len(configOptions) > 50 {
		results.OptimizationTips = append(results.OptimizationTips, 
			"Large configuration detected - consider using NixOS modules for better organization")
	}
}

// getRecommendedValue returns a sensible default value for a configuration option
func (da *DependencyAnalyzer) getRecommendedValue(optionName string) interface{} {
	defaults := map[string]interface{}{
		"sound.enable":                        true,
		"hardware.pulseaudio.enable":         true,
		"hardware.opengl.enable":             true,
		"services.xserver.enable":            true,
		"networking.networkmanager.enable":   true,
		"networking.firewall.enable":         true,
		"hardware.enableRedistributableFirmware": true,
		"nixpkgs.config.allowUnfree":         true,
		"boot.loader.efi.canTouchEfiVariables": true,
	}

	if value, exists := defaults[optionName]; exists {
		return value
	}

	// Infer default based on option name
	if strings.Contains(optionName, ".enable") {
		return true
	}

	return nil
}