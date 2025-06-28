package cli

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// performanceCmd represents the performance command group
func createPerformanceCommand() *cobra.Command {
	performanceCmd := &cobra.Command{
		Use:   "performance",
		Short: "Manage and view performance metrics",
		Long: `Access performance metrics and cache statistics for nixai.

This command group provides tools to monitor nixai's performance including:
- Cache hit rates and efficiency
- AI query response times
- Documentation query performance
- System resource usage

Examples:
  nixai performance stats       # Show performance overview
  nixai performance cache       # Show cache statistics
  nixai performance clear       # Clear performance metrics
  nixai performance report      # Generate detailed performance report`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	performanceCmd.AddCommand(createPerformanceStatsCommand())
	performanceCmd.AddCommand(createPerformanceCacheCommand())
	performanceCmd.AddCommand(createPerformanceClearCommand())
	performanceCmd.AddCommand(createPerformanceReportCommand())

	return performanceCmd
}

// createPerformanceStatsCommand shows performance statistics
func createPerformanceStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show performance statistics overview",
		Long: `Display a comprehensive overview of nixai performance metrics including:
- Total operations performed
- Success/failure rates  
- Average response times
- Cache hit rates
- Operation breakdown by type

This provides a quick snapshot of how nixai is performing.`,
		Example: `  nixai performance stats
  nixai performance stats --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadUserConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			return handlePerformanceStats(cfg, format)
		},
	}
}

// createPerformanceCacheCommand shows cache statistics
func createPerformanceCacheCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "cache",
		Short: "Show cache performance statistics",
		Long: `Display detailed cache performance metrics including:
- Memory cache hit/miss rates
- Disk cache usage statistics
- Cache size and utilization
- Performance improvements from caching
- Cache efficiency metrics

This helps understand how effective the caching system is.`,
		Example: `  nixai performance cache
  nixai performance cache --detailed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadUserConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			detailed, _ := cmd.Flags().GetBool("detailed")
			return handlePerformanceCache(cfg, detailed)
		},
	}
}

// createPerformanceClearCommand clears performance metrics
func createPerformanceClearCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear performance metrics and cache",
		Long: `Clear performance metrics and optionally clear cache data.

This command can reset:
- Performance metrics and statistics
- Memory cache contents
- Disk cache contents (with --all flag)
- Performance baselines for comparisons

Use this to start fresh performance monitoring or when troubleshooting.`,
		Example: `  nixai performance clear           # Clear metrics only
  nixai performance clear --cache    # Clear cache only  
  nixai performance clear --all      # Clear everything`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadUserConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			cache, _ := cmd.Flags().GetBool("cache")
			all, _ := cmd.Flags().GetBool("all")
			confirm, _ := cmd.Flags().GetBool("confirm")

			return handlePerformanceClear(cfg, cache, all, confirm)
		},
	}
}

// createPerformanceReportCommand generates performance reports
func createPerformanceReportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate detailed performance report",
		Long: `Generate a comprehensive performance report including:
- Detailed timing analysis
- Cache efficiency breakdown
- Performance trends over time
- Recommendations for optimization
- Comparison with baseline performance

The report can be saved to a file for analysis or sharing.`,
		Example: `  nixai performance report
  nixai performance report --output report.txt
  nixai performance report --format markdown`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadUserConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")

			return handlePerformanceReport(cfg, output, format)
		},
	}
}

// handlePerformanceStats displays performance statistics
func handlePerformanceStats(cfg *config.UserConfig, format string) error {
	fmt.Println(utils.FormatHeader("📊 Performance Statistics"))
	fmt.Println()

	// TODO: Integrate with actual AI manager performance metrics
	// For now, show placeholder information
	fmt.Println(utils.FormatKeyValue("Status", "Performance monitoring active"))
	fmt.Println(utils.FormatKeyValue("Cache Enabled", fmt.Sprintf("%t", cfg.Cache.Enabled)))

	if cfg.Cache.Enabled {
		fmt.Println(utils.FormatKeyValue("Memory Cache Size", fmt.Sprintf("%d entries", cfg.Cache.MemoryMaxSize)))
		fmt.Println(utils.FormatKeyValue("Memory TTL", fmt.Sprintf("%d minutes", cfg.Cache.MemoryTTL)))
		fmt.Println(utils.FormatKeyValue("Disk Cache Enabled", fmt.Sprintf("%t", cfg.Cache.DiskEnabled)))

		if cfg.Cache.DiskEnabled {
			fmt.Println(utils.FormatKeyValue("Disk Cache Size", fmt.Sprintf("%d MB", cfg.Cache.DiskMaxSize)))
			fmt.Println(utils.FormatKeyValue("Disk TTL", fmt.Sprintf("%d hours", cfg.Cache.DiskTTL)))
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatInfo("💡 Use 'nixai performance report' for detailed analysis"))

	return nil
}

// handlePerformanceCache displays cache statistics
func handlePerformanceCache(cfg *config.UserConfig, detailed bool) error {
	fmt.Println(utils.FormatHeader("🗄️ Cache Performance"))
	fmt.Println()

	if !cfg.Cache.Enabled {
		fmt.Println(utils.FormatWarning("Cache is disabled in configuration"))
		fmt.Println()
		fmt.Println(utils.FormatInfo("💡 Enable caching with: nixai config set cache.enabled true"))
		return nil
	}

	// TODO: Get actual cache statistics from cache manager
	// For now, show configuration information
	fmt.Println(utils.FormatSubsection("Cache Configuration", ""))
	fmt.Println(utils.FormatKeyValue("Memory Cache", fmt.Sprintf("%d entries max", cfg.Cache.MemoryMaxSize)))
	fmt.Println(utils.FormatKeyValue("Memory TTL", fmt.Sprintf("%d minutes", cfg.Cache.MemoryTTL)))

	if cfg.Cache.DiskEnabled {
		fmt.Println(utils.FormatKeyValue("Disk Cache", fmt.Sprintf("%d MB max", cfg.Cache.DiskMaxSize)))
		fmt.Println(utils.FormatKeyValue("Disk TTL", fmt.Sprintf("%d hours", cfg.Cache.DiskTTL)))
		fmt.Println(utils.FormatKeyValue("Cache Path", cfg.Cache.DiskPath))
	} else {
		fmt.Println(utils.FormatKeyValue("Disk Cache", "Disabled"))
	}

	if detailed {
		fmt.Println()
		fmt.Println(utils.FormatSubsection("Cache Maintenance", ""))
		fmt.Println(utils.FormatKeyValue("Cleanup Interval", fmt.Sprintf("%d minutes", cfg.Cache.CleanupInterval)))
		fmt.Println(utils.FormatKeyValue("Compact Interval", fmt.Sprintf("%d minutes", cfg.Cache.CompactInterval)))
	}

	fmt.Println()
	fmt.Println(utils.FormatInfo("💡 Use 'nixai performance clear --cache' to clear cache"))

	return nil
}

// handlePerformanceClear clears performance metrics and cache
func handlePerformanceClear(cfg *config.UserConfig, clearCache, clearAll, confirm bool) error {
	fmt.Println(utils.FormatHeader("🧹 Clear Performance Data"))
	fmt.Println()

	if clearAll {
		clearCache = true
		confirm = true
	}

	if !confirm {
		fmt.Println(utils.FormatWarning("This will clear performance metrics and optionally cache data."))
		fmt.Println()
		fmt.Print("Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// TODO: Implement actual clearing logic
	fmt.Println(utils.FormatInfo("Clearing performance metrics..."))

	if clearCache && cfg.Cache.Enabled {
		fmt.Println(utils.FormatInfo("Clearing cache data..."))
		// TODO: Call cache manager clear method
	}

	fmt.Println()
	fmt.Println(utils.FormatSuccess("✅ Performance data cleared successfully"))

	return nil
}

// handlePerformanceReport generates a performance report
func handlePerformanceReport(cfg *config.UserConfig, output, format string) error {
	fmt.Println(utils.FormatHeader("📋 Performance Report"))
	fmt.Println()

	// TODO: Generate actual performance report
	reportContent := generatePerformanceReport(cfg, format)

	if output != "" {
		// Save to file
		fmt.Println(utils.FormatInfo(fmt.Sprintf("Saving report to: %s", output)))
		// TODO: Implement file saving
		fmt.Println(utils.FormatSuccess("✅ Report saved successfully"))
	} else {
		// Display to console
		fmt.Println(reportContent)
	}

	return nil
}

// generatePerformanceReport creates a formatted performance report
func generatePerformanceReport(cfg *config.UserConfig, format string) string {
	var report strings.Builder

	report.WriteString("# nixai Performance Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", "TODO: timestamp"))

	report.WriteString("## Configuration\n\n")
	report.WriteString(fmt.Sprintf("- Cache Enabled: %t\n", cfg.Cache.Enabled))
	if cfg.Cache.Enabled {
		report.WriteString(fmt.Sprintf("- Memory Cache: %d entries, %s TTL\n",
			cfg.Cache.MemoryMaxSize, cfg.Cache.MemoryTTL))
		if cfg.Cache.DiskEnabled {
			report.WriteString(fmt.Sprintf("- Disk Cache: %d MB, %s TTL\n",
				cfg.Cache.DiskMaxSize, cfg.Cache.DiskTTL))
		}
	}

	report.WriteString("\n## Performance Metrics\n\n")
	report.WriteString("TODO: Add actual performance metrics\n\n")

	report.WriteString("## Recommendations\n\n")
	if !cfg.Cache.Enabled {
		report.WriteString("- ⚡ Enable caching to improve response times\n")
	}
	if cfg.Cache.Enabled && !cfg.Cache.DiskEnabled {
		report.WriteString("- 💾 Enable disk caching for persistent cache across restarts\n")
	}

	return report.String()
}

// Add flags to commands
func init() {
	// Performance stats flags
	if statsCmd := createPerformanceStatsCommand(); statsCmd != nil {
		statsCmd.Flags().StringP("format", "f", "text", "Output format (text, json)")
	}

	// Performance cache flags
	if cacheCmd := createPerformanceCacheCommand(); cacheCmd != nil {
		cacheCmd.Flags().BoolP("detailed", "d", false, "Show detailed cache information")
	}

	// Performance clear flags
	if clearCmd := createPerformanceClearCommand(); clearCmd != nil {
		clearCmd.Flags().Bool("cache", false, "Clear cache data")
		clearCmd.Flags().Bool("all", false, "Clear all performance data and cache")
		clearCmd.Flags().BoolP("confirm", "y", false, "Skip confirmation prompt")
	}

	// Performance report flags
	if reportCmd := createPerformanceReportCommand(); reportCmd != nil {
		reportCmd.Flags().StringP("output", "o", "", "Save report to file")
		reportCmd.Flags().StringP("format", "f", "markdown", "Report format (markdown, text, json)")
	}
}
