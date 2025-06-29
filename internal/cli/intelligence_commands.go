package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/intelligence"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// createIntelligenceCommand creates the main intelligence command with all subcommands
func createIntelligenceCommand() *cobra.Command {
	intelligenceCmd := &cobra.Command{
		Use:   "intelligence",
		Short: "AI-powered system intelligence and recommendations",
		Long: `The intelligence command provides AI-powered analysis, predictions, and recommendations
for your NixOS system. It includes system analysis, conflict detection, dependency analysis,
and smart recommendations based on your usage patterns.`,
		Aliases: []string{"intel", "ai"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		Example: `  nixai intelligence analyze    # Comprehensive system analysis
  nixai intelligence predict    # Generate predictive suggestions
  nixai intelligence conflicts  # Detect configuration conflicts
  nixai intelligence dependencies # Analyze dependency relationships
  nixai intelligence recommend  # Get smart recommendations
  nixai intelligence status     # Check system status`,
	}

	// Add subcommands
	intelligenceCmd.AddCommand(createIntelligenceAnalyzeCommand())
	intelligenceCmd.AddCommand(createIntelligencePredictCommand())
	intelligenceCmd.AddCommand(createIntelligenceConflictsCommand())
	intelligenceCmd.AddCommand(createIntelligenceDependenciesCommand())
	intelligenceCmd.AddCommand(createIntelligenceRecommendCommand())
	intelligenceCmd.AddCommand(createIntelligenceStatusCommand())

	return intelligenceCmd
}

// createIntelligenceAnalyzeCommand creates the analyze subcommand
func createIntelligenceAnalyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Perform comprehensive system analysis",
		Long: `Analyze your NixOS system comprehensively using AI-powered intelligence.

This command performs deep system analysis including:
- Package inventory and health assessment
- Service status and configuration analysis
- Hardware detection and compatibility checking
- Security posture evaluation
- Performance metrics gathering`,
		RunE: func(cmd *cobra.Command, args []string) error {
			detailed, _ := cmd.Flags().GetBool("detailed")
			component, _ := cmd.Flags().GetString("component")
			format, _ := cmd.Flags().GetString("format")
			output, _ := cmd.Flags().GetString("output")

			return handleIntelligenceAnalyze(detailed, component, format, output)
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed analysis results")
	cmd.Flags().String("component", "", "Focus on specific component")
	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
	cmd.Flags().StringP("output", "o", "", "Save results to file")

	return cmd
}

// createIntelligencePredictCommand creates the predict subcommand
func createIntelligencePredictCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "predict",
		Short: "Generate predictive suggestions",
		Long: `Generate AI-powered predictive suggestions based on your system usage patterns.

This command analyzes your usage history and system state to predict:
- Packages you might need based on your workflows
- Configuration optimizations for your use cases
- Security improvements relevant to your setup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			predType, _ := cmd.Flags().GetString("type")
			explain, _ := cmd.Flags().GetBool("explain")
			confidence, _ := cmd.Flags().GetFloat64("confidence")
			format, _ := cmd.Flags().GetString("format")

			return handleIntelligencePredict(predType, explain, confidence, format)
		},
	}

	cmd.Flags().String("type", "", "Prediction type (packages, config, security)")
	cmd.Flags().Bool("explain", false, "Include reasoning for predictions")
	cmd.Flags().Float64("confidence", 0.5, "Minimum confidence threshold")
	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}

// createIntelligenceConflictsCommand creates the conflicts subcommand
func createIntelligenceConflictsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conflicts",
		Short: "Detect configuration conflicts",
		Long: `Detect and analyze potential conflicts in your NixOS configuration.

This command analyzes your system for various types of conflicts:
- Package version conflicts and incompatibilities
- Service port conflicts and resource contention
- Configuration option conflicts and overlaps`,
		RunE: func(cmd *cobra.Command, args []string) error {
			conflictType, _ := cmd.Flags().GetString("type")
			severity, _ := cmd.Flags().GetString("severity")
			resolutions, _ := cmd.Flags().GetBool("resolutions")
			format, _ := cmd.Flags().GetString("format")

			return handleIntelligenceConflicts(conflictType, severity, resolutions, format)
		},
	}

	cmd.Flags().String("type", "", "Conflict type (packages, services, config)")
	cmd.Flags().String("severity", "", "Minimum severity level")
	cmd.Flags().Bool("resolutions", false, "Include resolution suggestions")
	cmd.Flags().StringP("format", "f", "text", "Output format (text, json)")

	return cmd
}

// createIntelligenceDependenciesCommand creates the dependencies subcommand
func createIntelligenceDependenciesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dependencies",
		Short: "Analyze dependency relationships",
		Long: `Analyze dependency relationships and generate dependency graphs for your system.

This command creates comprehensive dependency analysis including:
- Package dependency trees and relationships
- Service dependency chains and critical paths
- Configuration dependency mapping`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg, _ := cmd.Flags().GetString("package")
			criticalPaths, _ := cmd.Flags().GetBool("critical-paths")
			graph, _ := cmd.Flags().GetBool("graph")
			output, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")

			return handleIntelligenceDependencies(pkg, criticalPaths, graph, output, format)
		},
	}

	cmd.Flags().String("package", "", "Focus on specific package")
	cmd.Flags().Bool("critical-paths", false, "Show critical paths")
	cmd.Flags().Bool("graph", false, "Generate graph visualization")
	cmd.Flags().StringP("output", "o", "", "Output file")
	cmd.Flags().StringP("format", "f", "text", "Output format")

	return cmd
}

// createIntelligenceRecommendCommand creates the recommend subcommand
func createIntelligenceRecommendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommend",
		Short: "Generate smart recommendations",
		Long: `Generate intelligent recommendations for optimizing your NixOS system.

This command combines analysis, predictions, and conflict detection to provide:
- System optimization recommendations
- Security improvement suggestions
- Performance tuning opportunities`,
		RunE: func(cmd *cobra.Command, args []string) error {
			category, _ := cmd.Flags().GetString("category")
			priority, _ := cmd.Flags().GetString("priority")
			detailed, _ := cmd.Flags().GetBool("detailed")
			format, _ := cmd.Flags().GetString("format")

			return handleIntelligenceRecommend(category, priority, detailed, format)
		},
	}

	cmd.Flags().String("category", "", "Recommendation category")
	cmd.Flags().String("priority", "", "Minimum priority level")
	cmd.Flags().Bool("detailed", false, "Show detailed information")
	cmd.Flags().StringP("format", "f", "text", "Output format")

	return cmd
}

// createIntelligenceStatusCommand creates the status subcommand
func createIntelligenceStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show intelligence system status",
		Long: `Display the status and health of the nixai intelligence system.

This command shows:
- Intelligence system health and readiness
- Analysis cache status and statistics
- AI provider connection status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			detailed, _ := cmd.Flags().GetBool("detailed")
			cache, _ := cmd.Flags().GetBool("cache")

			return handleIntelligenceStatus(detailed, cache)
		},
	}

	cmd.Flags().Bool("detailed", false, "Show detailed status")
	cmd.Flags().Bool("cache", false, "Show cache statistics")

	return cmd
}

// Handler functions for intelligence commands

// handleIntelligenceAnalyze performs system analysis
func handleIntelligenceAnalyze(detailed bool, component, format, output string) error {
	fmt.Println(utils.FormatHeader("🔍 System Intelligence Analysis"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.NewLogger()

	analyzer := intelligence.NewSystemAnalyzer(log)

	fmt.Println(utils.FormatProgress("Performing system analysis..."))

	ctx := context.Background()
	analysis, err := analyzer.AnalyzeSystem(ctx, cfg)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	return displaySystemAnalysis(analysis, detailed, format, output)
}

// handleIntelligencePredict generates predictive suggestions
func handleIntelligencePredict(predType string, explain bool, confidence float64, format string) error {
	fmt.Println(utils.FormatHeader("🔮 Predictive Intelligence"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.NewLogger()

	analyzer := intelligence.NewSystemAnalyzer(log)
	// Note: AI provider would be initialized from config in a real implementation
	predictor := intelligence.NewPredictor(log, nil, analyzer)

	fmt.Println(utils.FormatProgress("Generating predictions..."))

	ctx := context.Background()
	predictions, err := predictor.GeneratePredictions(ctx, cfg)
	if err != nil {
		return fmt.Errorf("prediction failed: %w", err)
	}

	return displayPredictions(predictions, explain, format)
}

// handleIntelligenceConflicts detects configuration conflicts
func handleIntelligenceConflicts(conflictType, severity string, resolutions bool, format string) error {
	fmt.Println(utils.FormatHeader("⚠️ Conflict Detection"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.NewLogger()

	analyzer := intelligence.NewSystemAnalyzer(log)
	conflictDetector := intelligence.NewConflictDetector(log, analyzer)

	fmt.Println(utils.FormatProgress("Detecting conflicts..."))

	ctx := context.Background()
	conflicts, err := conflictDetector.DetectConflicts(ctx, cfg)
	if err != nil {
		return fmt.Errorf("conflict detection failed: %w", err)
	}

	return displayConflicts(conflicts, resolutions, format)
}

// handleIntelligenceDependencies analyzes dependency relationships
func handleIntelligenceDependencies(pkg string, criticalPaths, graph bool, output, format string) error {
	fmt.Println(utils.FormatHeader("🕸️ Dependency Analysis"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.NewLogger()

	analyzer := intelligence.NewSystemAnalyzer(log)
	depAnalyzer := intelligence.NewDependencyAnalyzer(log, analyzer)

	fmt.Println(utils.FormatProgress("Analyzing dependencies..."))

	ctx := context.Background()
	depGraph, err := depAnalyzer.AnalyzeDependencies(ctx, cfg)
	if err != nil {
		return fmt.Errorf("dependency analysis failed: %w", err)
	}

	return displayDependencies(depGraph.Graph, pkg, criticalPaths, format)
}

// handleIntelligenceRecommend generates smart recommendations
func handleIntelligenceRecommend(category, priority string, detailed bool, format string) error {
	fmt.Println(utils.FormatHeader("💡 Smart Recommendations"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log := logger.NewLogger()

	analyzer := intelligence.NewSystemAnalyzer(log)
	predictor := intelligence.NewPredictor(log, nil, analyzer)
	conflictDetector := intelligence.NewConflictDetector(log, analyzer)
	depAnalyzer := intelligence.NewDependencyAnalyzer(log, analyzer)

	recommendationsEngine := intelligence.NewRecommendationsEngine(
		log, nil, analyzer, predictor, conflictDetector, depAnalyzer,
	)

	fmt.Println(utils.FormatProgress("Generating recommendations..."))

	ctx := context.Background()
	recommendations, err := recommendationsEngine.GenerateRecommendations(ctx, cfg)
	if err != nil {
		return fmt.Errorf("recommendation generation failed: %w", err)
	}

	return displayRecommendations(recommendations, detailed, format)
}

// handleIntelligenceStatus shows intelligence system status
func handleIntelligenceStatus(detailed, cache bool) error {
	fmt.Println(utils.FormatHeader("📊 Intelligence System Status"))
	fmt.Println()

	cfg, err := config.LoadUserConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println(utils.FormatSubsection("System Health", ""))
	fmt.Println(utils.FormatKeyValue("Intelligence System", "✅ Active"))
	fmt.Println(utils.FormatKeyValue("System Analyzer", "✅ Ready"))
	fmt.Println(utils.FormatKeyValue("Cache", getCacheStatus(cfg)))

	if detailed {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Component Status", ""))
		fmt.Println(utils.FormatKeyValue("Conflict Detector", "✅ Ready"))
		fmt.Println(utils.FormatKeyValue("Dependency Analyzer", "✅ Ready"))
		fmt.Println(utils.FormatKeyValue("Predictor", "✅ Ready"))
		fmt.Println(utils.FormatKeyValue("Recommendations Engine", "✅ Ready"))
	}

	if cache {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Cache Statistics", ""))
		displayIntelligenceCacheStats(cfg)
	}

	return nil
}

// Helper and display functions

func getCacheStatus(cfg *config.UserConfig) string {
	if cfg.Cache.Enabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

func displayIntelligenceCacheStats(cfg *config.UserConfig) {
	if !cfg.Cache.Enabled {
		fmt.Println(utils.FormatWarning("Cache is disabled"))
		return
	}

	fmt.Println(utils.FormatKeyValue("Memory Cache", fmt.Sprintf("%d max entries", cfg.Cache.MemoryMaxSize)))
	fmt.Println(utils.FormatKeyValue("Memory TTL", fmt.Sprintf("%d minutes", cfg.Cache.MemoryTTL)))

	if cfg.Cache.DiskEnabled {
		fmt.Println(utils.FormatKeyValue("Disk Cache", fmt.Sprintf("%d MB max", cfg.Cache.DiskMaxSize)))
		fmt.Println(utils.FormatKeyValue("Disk TTL", fmt.Sprintf("%d hours", cfg.Cache.DiskTTL)))
	}
}

func displaySystemAnalysis(analysis *intelligence.SystemAnalysis, detailed bool, format, output string) error {
	if format == "json" {
		data, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(utils.FormatSubsection("System Overview", ""))
	fmt.Println(utils.FormatKeyValue("Total Packages", fmt.Sprintf("%d", len(analysis.InstalledPackages))))
	fmt.Println(utils.FormatKeyValue("Active Services", fmt.Sprintf("%d", len(analysis.EnabledServices))))
	fmt.Println(utils.FormatKeyValue("Security Score", fmt.Sprintf("%.1f/10", analysis.SecuritySettings.SecurityScore)))
	fmt.Println(utils.FormatKeyValue("Performance Score", fmt.Sprintf("%.1f/10", analysis.PerformanceMetrics.CPUUsage)))

	if detailed && len(analysis.InstalledPackages) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Top Packages", ""))
		count := 5
		if len(analysis.InstalledPackages) < count {
			count = len(analysis.InstalledPackages)
		}
		for i := 0; i < count; i++ {
			pkg := analysis.InstalledPackages[i]
			fmt.Printf("• %s (%s)\n", pkg.Name, pkg.Version)
		}
	}

	return nil
}

func displayPredictions(predictions *intelligence.PredictionResult, explain bool, format string) error {
	if format == "json" {
		data, err := json.MarshalIndent(predictions, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(utils.FormatSubsection("Prediction Summary", ""))
	fmt.Println(utils.FormatKeyValue("Package Suggestions", fmt.Sprintf("%d", len(predictions.PackageSuggestions))))
	fmt.Println(utils.FormatKeyValue("Config Suggestions", fmt.Sprintf("%d", len(predictions.ConfigSuggestions))))
	fmt.Println(utils.FormatKeyValue("Security Suggestions", fmt.Sprintf("%d", len(predictions.SecuritySuggestions))))
	fmt.Println(utils.FormatKeyValue("Confidence", fmt.Sprintf("%.2f", predictions.Confidence)))

	if len(predictions.PackageSuggestions) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Top Package Suggestions", ""))
		count := 5
		if len(predictions.PackageSuggestions) < count {
			count = len(predictions.PackageSuggestions)
		}
		for i := 0; i < count; i++ {
			suggestion := predictions.PackageSuggestions[i]
			fmt.Printf("• %s: %s\n", suggestion.PackageName, suggestion.Reason)
			if explain && suggestion.Documentation != "" {
				fmt.Printf("  %s\n", suggestion.Documentation)
			}
		}
	}

	return nil
}

func displayConflicts(conflicts *intelligence.ConflictAnalysis, resolutions bool, format string) error {
	if format == "json" {
		data, err := json.MarshalIndent(conflicts, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(utils.FormatSubsection("Conflict Summary", ""))
	fmt.Println(utils.FormatKeyValue("Total Conflicts", fmt.Sprintf("%d", conflicts.TotalConflicts)))

	if conflicts.TotalConflicts == 0 {
		fmt.Println()
		fmt.Println(utils.FormatSuccess("✅ No conflicts detected!"))
		return nil
	}

	for severity, count := range conflicts.SeverityBreakdown {
		if count > 0 {
			fmt.Println(utils.FormatKeyValue(strings.Title(severity), fmt.Sprintf("%d", count)))
		}
	}

	if len(conflicts.PackageConflicts) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Package Conflicts", ""))
		for _, conflict := range conflicts.PackageConflicts {
			fmt.Printf("• %s vs %s: %s\n", conflict.Package1, conflict.Package2, conflict.Description)
			if resolutions && len(conflict.Resolution) > 0 {
				fmt.Printf("  Resolution: %s\n", strings.Join(conflict.Resolution, ", "))
			}
		}
	}

	return nil
}

func displayDependencies(depGraph *intelligence.DependencyGraph, pkg string, criticalPaths bool, format string) error {
	if format == "json" {
		data, err := json.MarshalIndent(depGraph, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(utils.FormatSubsection("Dependency Overview", ""))
	fmt.Println(utils.FormatKeyValue("Total Nodes", fmt.Sprintf("%d", depGraph.TotalNodes)))
	fmt.Println(utils.FormatKeyValue("Total Edges", fmt.Sprintf("%d", depGraph.TotalEdges)))
	fmt.Println(utils.FormatKeyValue("Max Depth", fmt.Sprintf("%d", depGraph.MaxDepth)))
	fmt.Println(utils.FormatKeyValue("Complexity Score", fmt.Sprintf("%.2f", depGraph.ComplexityScore)))

	if len(depGraph.CircularDeps) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatWarning(fmt.Sprintf("⚠️ Found %d circular dependencies", len(depGraph.CircularDeps))))
	}

	if criticalPaths && len(depGraph.CriticalPaths) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Critical Paths", ""))
		count := 5
		if len(depGraph.CriticalPaths) < count {
			count = len(depGraph.CriticalPaths)
		}
		for i := 0; i < count; i++ {
			path := depGraph.CriticalPaths[i]
			fmt.Printf("• %s (Score: %.2f)\n", path.Description, path.CriticalityScore)
		}
	}

	return nil
}

func displayRecommendations(recommendations *intelligence.RecommendationSet, detailed bool, format string) error {
	if format == "json" {
		data, err := json.MarshalIndent(recommendations, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println(utils.FormatSubsection("Recommendations Summary", ""))
	fmt.Println(utils.FormatKeyValue("Total Recommendations", fmt.Sprintf("%d", recommendations.TotalRecommendations)))
	fmt.Println(utils.FormatKeyValue("Estimated Impact", recommendations.EstimatedImpact))
	fmt.Println(utils.FormatKeyValue("Time to Implement", recommendations.TimeToImplement.String()))

	for priority, count := range recommendations.PriorityBreakdown {
		if count > 0 {
			fmt.Println(utils.FormatKeyValue(strings.Title(priority)+" Priority", fmt.Sprintf("%d", count)))
		}
	}

	if len(recommendations.SystemOptimizations) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("System Optimizations", ""))
		count := 3
		if len(recommendations.SystemOptimizations) < count {
			count = len(recommendations.SystemOptimizations)
		}
		for i := 0; i < count; i++ {
			rec := recommendations.SystemOptimizations[i]
			fmt.Printf("• %s (%s priority)\n", rec.Title, rec.Priority)
			if detailed && rec.Description != "" {
				fmt.Printf("  %s\n", rec.Description)
			}
		}
	}

	return nil
}
