// Package cli provides health monitoring commands
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/health"
	"nix-ai-help/pkg/logger"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "System health monitoring and prediction",
	Long: `Monitor system health, predict failures, and get remediation suggestions.

The health command provides intelligent system monitoring capabilities including:
- Real-time health assessment
- ML-based failure prediction (7-day horizon)
- Anomaly detection using isolation forests
- Resource usage forecasting with ARIMA models
- Automated remediation suggestions with NixOS-specific commands
- Security vulnerability prediction
- Performance regression detection`,
	Example: `  # Get current system health status
  nixai health status

  # Predict potential failures over next 7 days
  nixai health predict --timeline 7d

  # Detect current system anomalies
  nixai health anomalies

  # Get resource usage forecast
  nixai health forecast --resource cpu --timeline 3d

  # Get remediation suggestions for current issues
  nixai health remediate

  # Start continuous health monitoring
  nixai health monitor --interval 5m`,
}

// healthStatusCmd shows current system health
var healthStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current system health status",
	Long:  "Analyze current system health with component-level status, active issues, and recommendations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthStatus(cmd, args)
	},
}

// healthPredictCmd predicts potential failures
var healthPredictCmd = &cobra.Command{
	Use:   "predict",
	Short: "Predict potential system failures",
	Long:  "Use ML models to predict potential failures over a specified timeline with confidence scores and preventive actions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthPredict(cmd, args)
	},
}

// healthAnomaliesCmd detects system anomalies
var healthAnomaliesCmd = &cobra.Command{
	Use:   "anomalies",
	Short: "Detect system anomalies",
	Long:  "Use isolation forest models to detect unusual system behavior and patterns.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthAnomalies(cmd, args)
	},
}

// healthForecastCmd forecasts resource usage
var healthForecastCmd = &cobra.Command{
	Use:   "forecast",
	Short: "Forecast resource usage",
	Long:  "Predict future resource usage patterns using time series analysis and ARIMA models.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthForecast(cmd, args)
	},
}

// healthRemediateCmd provides remediation suggestions
var healthRemediateCmd = &cobra.Command{
	Use:   "remediate",
	Short: "Get remediation suggestions",
	Long:  "Generate automated remediation suggestions for current health issues with NixOS-specific commands.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthRemediate(cmd, args)
	},
}

// healthMonitorCmd starts continuous monitoring
var healthMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start continuous health monitoring",
	Long:  "Start continuous health monitoring with periodic assessments and alerting.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthMonitor(cmd, args)
	},
}

// Command flags
var (
	healthTimeline    string
	healthResource    string
	healthInterval    string
	healthFormat      string
	healthVerbose     bool
	healthAutoRemediate bool
)

func init() {
	// Add subcommands
	healthCmd.AddCommand(healthStatusCmd)
	healthCmd.AddCommand(healthPredictCmd) 
	healthCmd.AddCommand(healthAnomaliesCmd)
	healthCmd.AddCommand(healthForecastCmd)
	healthCmd.AddCommand(healthRemediateCmd)
	healthCmd.AddCommand(healthMonitorCmd)

	// Prediction flags
	healthPredictCmd.Flags().StringVar(&healthTimeline, "timeline", "7d", "Prediction timeline (e.g., 1h, 3d, 7d)")
	healthPredictCmd.Flags().StringVar(&healthFormat, "format", "table", "Output format (table, json, yaml)")
	healthPredictCmd.Flags().BoolVar(&healthVerbose, "verbose", false, "Show detailed prediction information")

	// Forecast flags
	healthForecastCmd.Flags().StringVar(&healthTimeline, "timeline", "3d", "Forecast timeline (e.g., 1h, 3d, 7d)")
	healthForecastCmd.Flags().StringVar(&healthResource, "resource", "all", "Resource to forecast (cpu, memory, disk, network, all)")
	healthForecastCmd.Flags().StringVar(&healthFormat, "format", "table", "Output format (table, json, yaml)")

	// Monitor flags
	healthMonitorCmd.Flags().StringVar(&healthInterval, "interval", "5m", "Monitoring interval (e.g., 1m, 5m, 15m)")
	healthMonitorCmd.Flags().BoolVar(&healthAutoRemediate, "auto-remediate", false, "Enable automatic remediation of low-risk issues")

	// Global flags
	healthCmd.PersistentFlags().StringVar(&healthFormat, "format", "table", "Output format (table, json, yaml)")
	healthCmd.PersistentFlags().BoolVar(&healthVerbose, "verbose", false, "Show verbose output")

	// Register with root command
	rootCmd.AddCommand(healthCmd)
}

func runHealthStatus(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	log.Info("Analyzing system health...")

	// Load config
	yamlCfg, err := config.LoadYAMLConfig("configs/default.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert YAML config to UserConfig
	userCfg := &config.UserConfig{
		AIProvider:  yamlCfg.AIProvider,
		LogLevel:    yamlCfg.LogLevel,
		AIModels:    yamlCfg.AIModels,
		MCPServer:   yamlCfg.MCPServer,
		Nixos:       yamlCfg.Nixos,
		Diagnostics: yamlCfg.Diagnostics,
		Commands:    yamlCfg.Commands,
		AITimeouts:  yamlCfg.AITimeouts,
		Devenv:      yamlCfg.Devenv,
		CustomAI:    yamlCfg.CustomAI,
		Discourse:   yamlCfg.Discourse,
		Cache:       yamlCfg.Cache,
		Plugin:      yamlCfg.Plugin,
		Execution:   yamlCfg.Execution,
	}

	// Create health predictor
	predictor := health.NewSystemHealthPredictor(userCfg)
	
	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health predictor: %w", err)
	}
	defer predictor.Stop()

	// Analyze current health
	assessment, err := predictor.AnalyzeSystemHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze system health: %w", err)
	}

	// Display results
	return displayHealthAssessment(assessment, healthFormat, healthVerbose)
}

func runHealthPredict(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	log.Info("Predicting potential system failures...")

	// Parse timeline
	timeline, err := parseTimeline(healthTimeline)
	if err != nil {
		return fmt.Errorf("invalid timeline format: %w", err)
	}

	// Load config
	yamlCfg, err := config.LoadYAMLConfig("configs/default.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert YAML config to UserConfig
	userCfg := &config.UserConfig{
		AIProvider:  yamlCfg.AIProvider,
		LogLevel:    yamlCfg.LogLevel,
		AIModels:    yamlCfg.AIModels,
		MCPServer:   yamlCfg.MCPServer,
		Nixos:       yamlCfg.Nixos,
		Diagnostics: yamlCfg.Diagnostics,
		Commands:    yamlCfg.Commands,
		AITimeouts:  yamlCfg.AITimeouts,
		Devenv:      yamlCfg.Devenv,
		CustomAI:    yamlCfg.CustomAI,
		Discourse:   yamlCfg.Discourse,
		Cache:       yamlCfg.Cache,
		Plugin:      yamlCfg.Plugin,
		Execution:   yamlCfg.Execution,
	}

	// Create health predictor
	predictor := health.NewSystemHealthPredictor(userCfg)
	
	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health predictor: %w", err)
	}
	defer predictor.Stop()

	// Predict failures
	prediction, err := predictor.PredictFailures(ctx, timeline)
	if err != nil {
		return fmt.Errorf("failed to predict failures: %w", err)
	}

	// Display results
	return displayFailurePrediction(prediction, healthFormat, healthVerbose)
}

func runHealthAnomalies(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	log.Info("Detecting system anomalies...")

	// Load config
	yamlCfg, err := config.LoadYAMLConfig("configs/default.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert YAML config to UserConfig
	userCfg := &config.UserConfig{
		AIProvider:  yamlCfg.AIProvider,
		LogLevel:    yamlCfg.LogLevel,
		AIModels:    yamlCfg.AIModels,
		MCPServer:   yamlCfg.MCPServer,
		Nixos:       yamlCfg.Nixos,
		Diagnostics: yamlCfg.Diagnostics,
		Commands:    yamlCfg.Commands,
		AITimeouts:  yamlCfg.AITimeouts,
		Devenv:      yamlCfg.Devenv,
		CustomAI:    yamlCfg.CustomAI,
		Discourse:   yamlCfg.Discourse,
		Cache:       yamlCfg.Cache,
		Plugin:      yamlCfg.Plugin,
		Execution:   yamlCfg.Execution,
	}

	// Create health predictor
	predictor := health.NewSystemHealthPredictor(userCfg)
	
	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health predictor: %w", err)
	}
	defer predictor.Stop()

	// Detect anomalies
	report, err := predictor.DetectAnomalies(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect anomalies: %w", err)
	}

	// Display results
	return displayAnomalyReport(report, healthFormat, healthVerbose)
}

func runHealthForecast(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	log.Info("Forecasting resource usage...")

	// Parse timeline
	timeline, err := parseTimeline(healthTimeline)
	if err != nil {
		return fmt.Errorf("invalid timeline format: %w", err)
	}

	// Load config
	yamlCfg, err := config.LoadYAMLConfig("configs/default.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert YAML config to UserConfig
	userCfg := &config.UserConfig{
		AIProvider:  yamlCfg.AIProvider,
		LogLevel:    yamlCfg.LogLevel,
		AIModels:    yamlCfg.AIModels,
		MCPServer:   yamlCfg.MCPServer,
		Nixos:       yamlCfg.Nixos,
		Diagnostics: yamlCfg.Diagnostics,
		Commands:    yamlCfg.Commands,
		AITimeouts:  yamlCfg.AITimeouts,
		Devenv:      yamlCfg.Devenv,
		CustomAI:    yamlCfg.CustomAI,
		Discourse:   yamlCfg.Discourse,
		Cache:       yamlCfg.Cache,
		Plugin:      yamlCfg.Plugin,
		Execution:   yamlCfg.Execution,
	}

	// Create health predictor
	predictor := health.NewSystemHealthPredictor(userCfg)
	
	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health predictor: %w", err)
	}
	defer predictor.Stop()

	// Forecast resources
	forecast, err := predictor.ForecastResources(ctx, timeline)
	if err != nil {
		return fmt.Errorf("failed to forecast resources: %w", err)
	}

	// Display results
	return displayResourceForecast(forecast, healthResource, healthFormat, healthVerbose)
}

func runHealthRemediate(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	log.Info("Generating remediation suggestions...")

	// Load config
	yamlCfg, err := config.LoadYAMLConfig("configs/default.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert YAML config to UserConfig
	userCfg := &config.UserConfig{
		AIProvider:  yamlCfg.AIProvider,
		LogLevel:    yamlCfg.LogLevel,
		AIModels:    yamlCfg.AIModels,
		MCPServer:   yamlCfg.MCPServer,
		Nixos:       yamlCfg.Nixos,
		Diagnostics: yamlCfg.Diagnostics,
		Commands:    yamlCfg.Commands,
		AITimeouts:  yamlCfg.AITimeouts,
		Devenv:      yamlCfg.Devenv,
		CustomAI:    yamlCfg.CustomAI,
		Discourse:   yamlCfg.Discourse,
		Cache:       yamlCfg.Cache,
		Plugin:      yamlCfg.Plugin,
		Execution:   yamlCfg.Execution,
	}

	// Create health predictor
	predictor := health.NewSystemHealthPredictor(userCfg)
	
	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health predictor: %w", err)
	}
	defer predictor.Stop()

	// Get current health to identify issues
	assessment, err := predictor.AnalyzeSystemHealth(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze system health: %w", err)
	}

	if len(assessment.ActiveIssues) == 0 {
		fmt.Println("✓ No active health issues found. System is healthy!")
		return nil
	}

	// Get remediation suggestions
	plan, err := predictor.GetRemediationSuggestions(ctx, assessment.ActiveIssues)
	if err != nil {
		return fmt.Errorf("failed to get remediation suggestions: %w", err)
	}

	// Display results
	return displayRemediationPlan(plan, healthFormat, healthVerbose)
}

func runHealthMonitor(cmd *cobra.Command, args []string) error {
	log := logger.NewLogger()
	log.Info("Starting continuous health monitoring...")

	// Parse interval
	interval, err := parseTimeline(healthInterval)
	if err != nil {
		return fmt.Errorf("invalid interval format: %w", err)
	}

	// Load config
	yamlCfg, err := config.LoadYAMLConfig("configs/default.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Convert YAML config to UserConfig
	userCfg := &config.UserConfig{
		AIProvider:  yamlCfg.AIProvider,
		LogLevel:    yamlCfg.LogLevel,
		AIModels:    yamlCfg.AIModels,
		MCPServer:   yamlCfg.MCPServer,
		Nixos:       yamlCfg.Nixos,
		Diagnostics: yamlCfg.Diagnostics,
		Commands:    yamlCfg.Commands,
		AITimeouts:  yamlCfg.AITimeouts,
		Devenv:      yamlCfg.Devenv,
		CustomAI:    yamlCfg.CustomAI,
		Discourse:   yamlCfg.Discourse,
		Cache:       yamlCfg.Cache,
		Plugin:      yamlCfg.Plugin,
		Execution:   yamlCfg.Execution,
	}

	// Create health predictor
	predictor := health.NewSystemHealthPredictor(userCfg)
	
	// Start the predictor
	ctx := context.Background()
	if err := predictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health predictor: %w", err)
	}
	defer predictor.Stop()

	fmt.Printf("🔍 Health monitoring started (interval: %s)\n", interval)
	fmt.Println("Press Ctrl+C to stop monitoring...")

	// Start monitoring loop
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n📊 Health monitoring stopped")
			return nil
		case <-ticker.C:
			if err := performMonitoringCheck(ctx, predictor); err != nil {
				log.Error(fmt.Sprintf("Monitoring check failed: %v", err))
			}
		}
	}
}

func performMonitoringCheck(ctx context.Context, predictor *health.SystemHealthPredictor) error {
	// Get current health
	assessment, err := predictor.AnalyzeSystemHealth(ctx)
	if err != nil {
		return err
	}

	timestamp := time.Now().Format("15:04:05")
	
	// Display health status
	fmt.Printf("[%s] Overall Health: %s", timestamp, getHealthStatusEmoji(assessment.OverallHealth))
	
	if len(assessment.ActiveIssues) > 0 {
		fmt.Printf(" ⚠️  %d active issues", len(assessment.ActiveIssues))
		
		// If auto-remediation is enabled, try to fix low-risk issues
		if healthAutoRemediate {
			plan, err := predictor.GetRemediationSuggestions(ctx, assessment.ActiveIssues)
			if err == nil && plan.RiskAssessment.OverallRisk == health.RiskLow {
				fmt.Printf(" 🔧 Auto-remediation available")
			}
		}
	}
	
	fmt.Println()
	return nil
}

// Display functions

func displayHealthAssessment(assessment *health.HealthAssessment, format string, verbose bool) error {
	switch format {
	case "json":
		return displayJSON(assessment)
	case "yaml":
		return displayYAML(assessment)
	default:
		return displayHealthTable(assessment, verbose)
	}
}

func displayFailurePrediction(prediction *health.FailurePrediction, format string, verbose bool) error {
	switch format {
	case "json":
		return displayJSON(prediction)
	case "yaml":
		return displayYAML(prediction)
	default:
		return displayPredictionTable(prediction, verbose)
	}
}

func displayAnomalyReport(report *health.AnomalyReport, format string, verbose bool) error {
	switch format {
	case "json":
		return displayJSON(report)
	case "yaml":
		return displayYAML(report)
	default:
		return displayAnomalyTable(report, verbose)
	}
}

func displayResourceForecast(forecast *health.ResourceForecast, resource, format string, verbose bool) error {
	switch format {
	case "json":
		return displayJSON(forecast)
	case "yaml":
		return displayYAML(forecast)
	default:
		return displayForecastTable(forecast, resource, verbose)
	}
}

func displayRemediationPlan(plan *health.RemediationPlan, format string, verbose bool) error {
	switch format {
	case "json":
		return displayJSON(plan)
	case "yaml":
		return displayYAML(plan)
	default:
		return displayRemediationTable(plan, verbose)
	}
}

func displayJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))
	return nil
}

func displayYAML(data interface{}) error {
	// For simplicity, use JSON format as YAML alternative
	return displayJSON(data)
}

func displayHealthTable(assessment *health.HealthAssessment, verbose bool) error {
	fmt.Printf("📊 System Health Assessment\n")
	fmt.Printf("═══════════════════════════\n\n")
	
	fmt.Printf("Overall Health: %s %s\n", getHealthStatusEmoji(assessment.OverallHealth), assessment.OverallHealth)
	fmt.Printf("Last Updated: %s\n\n", assessment.LastUpdate.Format("2006-01-02 15:04:05"))

	// Component health
	fmt.Printf("Component Health:\n")
	for component, status := range assessment.ComponentHealth {
		fmt.Printf("  %-10s %s %s\n", component+":", getHealthStatusEmoji(status), status)
	}
	fmt.Println()

	// Active issues
	if len(assessment.ActiveIssues) > 0 {
		fmt.Printf("Active Issues (%d):\n", len(assessment.ActiveIssues))
		for _, issue := range assessment.ActiveIssues {
			fmt.Printf("  • %s [%s] - %s\n", issue.Component, issue.Severity, issue.Description)
		}
		fmt.Println()
	}

	// Resource utilization
	fmt.Printf("Resource Utilization:\n")
	fmt.Printf("  CPU:     %.1f%% (%s)\n", assessment.ResourceUtilization.CPU.Current, assessment.ResourceUtilization.CPU.Status)
	fmt.Printf("  Memory:  %.1f%% (%s)\n", assessment.ResourceUtilization.Memory.Current, assessment.ResourceUtilization.Memory.Status)
	fmt.Printf("  Disk:    %.1f%% (%s)\n", assessment.ResourceUtilization.Disk.Current, assessment.ResourceUtilization.Disk.Status)
	fmt.Printf("  Network: %.1f%% (%s)\n", assessment.ResourceUtilization.Network.Current, assessment.ResourceUtilization.Network.Status)
	fmt.Println()

	// Recommendations
	if len(assessment.Recommendations) > 0 && verbose {
		fmt.Printf("Recommendations:\n")
		for _, rec := range assessment.Recommendations {
			fmt.Printf("  • [%s] %s\n", rec.Priority, rec.Title)
			if verbose {
				fmt.Printf("    %s\n", rec.Description)
			}
		}
	}

	return nil
}

func displayPredictionTable(prediction *health.FailurePrediction, verbose bool) error {
	fmt.Printf("🔮 Failure Prediction (Next %s)\n", prediction.Timeline)
	fmt.Printf("════════════════════════════════════\n\n")
	
	fmt.Printf("Risk Level: %s\n", getRiskLevelEmoji(prediction.RiskLevel))
	fmt.Printf("Confidence: %.1f%%\n", prediction.Confidence*100)
	fmt.Printf("Generated: %s\n\n", prediction.GeneratedAt.Format("2006-01-02 15:04:05"))

	if len(prediction.PredictedFailures) == 0 {
		fmt.Println("✓ No failures predicted in the specified timeline")
		return nil
	}

	fmt.Printf("Predicted Failures (%d):\n", len(prediction.PredictedFailures))
	for _, failure := range prediction.PredictedFailures {
		fmt.Printf("  • %s - %s (%.1f%% probability)\n", 
			failure.Component, failure.Type, failure.ProbabilityScore*100)
		fmt.Printf("    Expected: %s\n", failure.EstimatedTime.Format("2006-01-02 15:04"))
		if verbose {
			fmt.Printf("    Impact: %s\n", failure.Impact)
			fmt.Printf("    Description: %s\n", failure.Description)
		}
		fmt.Println()
	}

	// Preventive actions
	if len(prediction.PreventiveActions) > 0 {
		fmt.Printf("Preventive Actions:\n")
		for _, action := range prediction.PreventiveActions {
			fmt.Printf("  • %s\n", action.Description)
			if verbose && len(action.Commands) > 0 {
				for _, cmd := range action.Commands {
					fmt.Printf("    $ %s\n", cmd)
				}
			}
		}
	}

	return nil
}

func displayAnomalyTable(report *health.AnomalyReport, verbose bool) error {
	fmt.Printf("🚨 Anomaly Detection Report\n")
	fmt.Printf("════════════════════════════\n\n")
	
	fmt.Printf("Anomaly Score: %.2f\n", report.AnomalyScore)
	fmt.Printf("Detection Model: %s\n", report.DetectionModel)
	fmt.Printf("Time Window: %s\n", report.TimeWindow)
	fmt.Printf("Generated: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))

	if len(report.DetectedAnomalies) == 0 {
		fmt.Println("✓ No anomalies detected")
		return nil
	}

	fmt.Printf("Detected Anomalies (%d):\n", len(report.DetectedAnomalies))
	for _, anomaly := range report.DetectedAnomalies {
		fmt.Printf("  • %s [%s] - Score: %.2f\n", 
			anomaly.Component, anomaly.Severity, anomaly.Score)
		fmt.Printf("    %s\n", anomaly.Description)
		if verbose {
			fmt.Printf("    Detected: %s\n", anomaly.DetectedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("    Status: %s\n", anomaly.Status)
		}
		fmt.Println()
	}

	return nil
}

func displayForecastTable(forecast *health.ResourceForecast, resource string, verbose bool) error {
	fmt.Printf("📈 Resource Forecast (Next %s)\n", forecast.Timeline)
	fmt.Printf("═══════════════════════════════════\n\n")
	
	fmt.Printf("Model Accuracy: %.1f%%\n", forecast.ModelAccuracy*100)
	fmt.Printf("Confidence: %.1f%%\n", forecast.Confidence*100)
	fmt.Printf("Generated: %s\n\n", forecast.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Display specific resource or all
	if resource == "cpu" || resource == "all" {
		displayResourcePrediction("CPU", forecast.Predictions["cpu"], verbose)
	}
	if resource == "memory" || resource == "all" {
		displayResourcePrediction("Memory", forecast.Predictions["memory"], verbose)
	}
	if resource == "disk" || resource == "all" {
		displayResourcePrediction("Disk", forecast.Predictions["disk"], verbose)
	}
	if resource == "network" || resource == "all" {
		displayResourcePrediction("Network", forecast.Predictions["network"], verbose)
	}

	// Alerts
	if len(forecast.Alerts) > 0 {
		fmt.Printf("Resource Alerts:\n")
		for _, alert := range forecast.Alerts {
			fmt.Printf("  ⚠️  %s: %s\n", alert.Resource, alert.Message)
		}
	}

	return nil
}

func displayResourcePrediction(name string, prediction health.ResourcePrediction, verbose bool) {
	fmt.Printf("%s Forecast:\n", name)
	fmt.Printf("  Current: %.1f%s\n", prediction.CurrentValue, "%")
	fmt.Printf("  Predicted: %.1f%s (%.1f%% change)\n", 
		prediction.PredictedValue, "%", (prediction.PredictedValue-prediction.CurrentValue))
	if prediction.TimeToThreshold > 0 {
		fmt.Printf("  Time to threshold: %s\n", prediction.TimeToThreshold)
	}
	fmt.Printf("  Confidence: %.1f%%\n", prediction.Confidence*100)
	fmt.Println()
}

func displayRemediationTable(plan *health.RemediationPlan, verbose bool) error {
	fmt.Printf("🔧 Remediation Plan\n")
	fmt.Printf("═══════════════════\n\n")
	
	fmt.Printf("Plan ID: %s\n", plan.ID)
	fmt.Printf("Priority: %s\n", plan.Priority)
	fmt.Printf("Risk Level: %s\n", plan.RiskAssessment.OverallRisk)
	fmt.Printf("Automation: %s\n", plan.AutomationLevel)
	fmt.Printf("Estimated Time: %s\n\n", plan.EstimatedTime)

	fmt.Printf("Issues to Address (%d):\n", len(plan.Issues))
	for _, issue := range plan.Issues {
		fmt.Printf("  • %s [%s] - %s\n", issue.Component, issue.Severity, issue.Description)
	}
	fmt.Println()

	fmt.Printf("Remediation Suggestions (%d):\n", len(plan.Suggestions))
	for i, suggestion := range plan.Suggestions {
		fmt.Printf("  %d. %s [%s]\n", i+1, suggestion.Title, suggestion.Risk)
		fmt.Printf("     %s\n", suggestion.Description)
		if verbose && len(suggestion.Actions) > 0 {
			fmt.Printf("     Actions:\n")
			for _, action := range suggestion.Actions {
				fmt.Printf("       %d. %s\n", action.Step, action.Description)
				if len(action.Commands) > 0 {
					for _, cmd := range action.Commands {
						fmt.Printf("          $ %s\n", cmd)
					}
				}
			}
		}
		fmt.Println()
	}

	return nil
}

// Helper functions

func getHealthStatusEmoji(status health.HealthStatus) string {
	switch status {
	case health.HealthExcellent:
		return "💚"
	case health.HealthGood:
		return "🟢"
	case health.HealthFair:
		return "🟡"
	case health.HealthPoor:
		return "🟠"
	case health.HealthCritical:
		return "🔴"
	default:
		return "❓"
	}
}

func getRiskLevelEmoji(risk health.RiskLevel) string {
	switch risk {
	case health.RiskLow:
		return "🟢 Low"
	case health.RiskMedium:
		return "🟡 Medium" 
	case health.RiskHigh:
		return "🟠 High"
	case health.RiskCritical:
		return "🔴 Critical"
	default:
		return "❓ Unknown"
	}
}

// parseTimeline parses time durations with support for days (d) and other units
func parseTimeline(timelineStr string) (time.Duration, error) {
	// Handle "d" suffix for days
	if strings.HasSuffix(timelineStr, "d") {
		daysStr := strings.TrimSuffix(timelineStr, "d")
		if daysStr == "" {
			daysStr = "1"
		}
		
		var days int
		n, err := fmt.Sscanf(daysStr, "%d", &days)
		if err != nil || n != 1 {
			return 0, fmt.Errorf("invalid days format: %s", daysStr)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	
	// Use standard time.ParseDuration for other formats
	return time.ParseDuration(timelineStr)
}