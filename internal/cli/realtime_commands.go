package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"nix-ai-help/internal/health"
	"nix-ai-help/pkg/logger"
)

// CreateRealtimeCommand creates the realtime monitoring command
func CreateRealtimeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "realtime",
		Short: "Real-time system monitoring and alerting",
		Long: `Provides real-time system monitoring with streaming updates, alerts, and anomaly detection.
		
Features:
- Real-time metrics streaming (1-second updates)
- Intelligent alerting with configurable thresholds
- Anomaly detection using statistical analysis
- Resource pressure monitoring
- Process and service tracking
- Network activity monitoring
- Security event detection`,
		Example: `  # Start real-time monitoring dashboard
  nixai realtime monitor

  # Stream metrics to console
  nixai realtime stream --format json

  # Monitor specific metrics
  nixai realtime stream --metrics cpu,memory,disk

  # Show active alerts
  nixai realtime alerts

  # Monitor with custom update interval
  nixai realtime monitor --interval 5s`,
	}

	// Add subcommands
	cmd.AddCommand(realtimeMonitorCommand())
	cmd.AddCommand(realtimeStreamCommand())
	cmd.AddCommand(realtimeAlertsCommand())
	cmd.AddCommand(realtimeThresholdsCommand())
	cmd.AddCommand(realtimeStatsCommand())

	return cmd
}

// realtimeMonitorCommand provides an interactive monitoring dashboard
func realtimeMonitorCommand() *cobra.Command {
	var (
		interval     string
		noAlerts     bool
		noAnomalies  bool
		refreshRate  string
	)

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Interactive real-time monitoring dashboard",
		Long: `Launches an interactive dashboard showing real-time system metrics, alerts, and trends.
		
The dashboard provides:
- Live updating metrics display
- Component health status
- Active alerts and notifications  
- Trend analysis and predictions
- Resource pressure indicators
- Process and service monitoring`,
		RunE: func(cmd *cobra.Command, args []string) error {
			intervalDuration, err := time.ParseDuration(interval)
			if err != nil {
				return fmt.Errorf("invalid interval: %w", err)
			}

			refreshDuration, err := time.ParseDuration(refreshRate)
			if err != nil {
				return fmt.Errorf("invalid refresh rate: %w", err)
			}

			return runInteractiveMonitor(intervalDuration, refreshDuration, !noAlerts, !noAnomalies)
		},
	}

	cmd.Flags().StringVar(&interval, "interval", "1s", "Monitoring interval (e.g., 1s, 5s, 10s)")
	cmd.Flags().BoolVar(&noAlerts, "no-alerts", false, "Disable alert monitoring")
	cmd.Flags().BoolVar(&noAnomalies, "no-anomalies", false, "Disable anomaly detection")
	cmd.Flags().StringVar(&refreshRate, "refresh", "1s", "Dashboard refresh rate")

	return cmd
}

// realtimeStreamCommand streams metrics to stdout
func realtimeStreamCommand() *cobra.Command {
	var (
		format      string
		metrics     []string
		interval    string
		maxUpdates  int
		showTrends  bool
		showAlerts  bool
		compact     bool
	)

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream real-time metrics to stdout",
		Long: `Streams real-time system metrics to stdout in various formats.
		
Output formats:
- json: JSON format for programmatic consumption
- table: Human-readable table format  
- csv: CSV format for data analysis
- prometheus: Prometheus metrics format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			intervalDuration, err := time.ParseDuration(interval)
			if err != nil {
				return fmt.Errorf("invalid interval: %w", err)
			}

			return runMetricsStream(format, metrics, intervalDuration, maxUpdates, showTrends, showAlerts, compact)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (json, table, csv, prometheus)")
	cmd.Flags().StringSliceVar(&metrics, "metrics", []string{}, "Specific metrics to monitor (default: all)")
	cmd.Flags().StringVar(&interval, "interval", "5s", "Streaming interval")
	cmd.Flags().IntVar(&maxUpdates, "max-updates", 0, "Maximum number of updates (0 = unlimited)")
	cmd.Flags().BoolVar(&showTrends, "show-trends", true, "Include trend analysis")
	cmd.Flags().BoolVar(&showAlerts, "show-alerts", true, "Include active alerts")
	cmd.Flags().BoolVar(&compact, "compact", false, "Compact output format")

	return cmd
}

// realtimeAlertsCommand manages alerts
func realtimeAlertsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Manage real-time alerts",
		Long:  "View, acknowledge, and manage real-time system alerts.",
	}

	cmd.AddCommand(realtimeAlertsListCommand())
	cmd.AddCommand(realtimeAlertsAckCommand())
	cmd.AddCommand(realtimeAlertsStreamCommand())

	return cmd
}

// realtimeAlertsListCommand lists active alerts
func realtimeAlertsListCommand() *cobra.Command {
	var (
		format   string
		severity string
		showAll  bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active alerts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listActiveAlerts(format, severity, showAll)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")
	cmd.Flags().StringVar(&severity, "severity", "", "Filter by severity (info, warning, critical, emergency)")
	cmd.Flags().BoolVar(&showAll, "all", false, "Show all alerts including acknowledged")

	return cmd
}

// realtimeAlertsAckCommand acknowledges alerts
func realtimeAlertsAckCommand() *cobra.Command {
	var acknowledgedBy string

	cmd := &cobra.Command{
		Use:   "acknowledge <alert-id>",
		Short: "Acknowledge an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return acknowledgeAlert(args[0], acknowledgedBy)
		},
	}

	cmd.Flags().StringVar(&acknowledgedBy, "by", "", "Who is acknowledging the alert")

	return cmd
}

// realtimeAlertsStreamCommand streams alert notifications
func realtimeAlertsStreamCommand() *cobra.Command {
	var (
		format   string
		severity string
	)

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream alert notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			return streamAlerts(format, severity)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")
	cmd.Flags().StringVar(&severity, "severity", "", "Filter by severity")

	return cmd
}

// realtimeThresholdsCommand manages alert thresholds
func realtimeThresholdsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thresholds",
		Short: "Manage alert thresholds",
		Long:  "View and configure alert thresholds for various metrics.",
	}

	cmd.AddCommand(realtimeThresholdsListCommand())
	cmd.AddCommand(realtimeThresholdsSetCommand())

	return cmd
}

// realtimeThresholdsListCommand lists current thresholds
func realtimeThresholdsListCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current alert thresholds",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listThresholds(format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// realtimeThresholdsSetCommand sets thresholds
func realtimeThresholdsSetCommand() *cobra.Command {
	var (
		warning   float64
		critical  float64
		emergency float64
	)

	cmd := &cobra.Command{
		Use:   "set <metric> --warning <value> --critical <value> --emergency <value>",
		Short: "Set alert thresholds for a metric",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setThresholds(args[0], warning, critical, emergency)
		},
	}

	cmd.Flags().Float64Var(&warning, "warning", 0, "Warning threshold")
	cmd.Flags().Float64Var(&critical, "critical", 0, "Critical threshold")
	cmd.Flags().Float64Var(&emergency, "emergency", 0, "Emergency threshold")

	return cmd
}

// realtimeStatsCommand shows monitoring statistics
func realtimeStatsCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show monitoring statistics",
		Long:  "Display statistics about the real-time monitoring system.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showMonitoringStats(format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// Implementation functions

func runInteractiveMonitor(interval, refreshRate time.Duration, enableAlerts, enableAnomalies bool) error {
	logger := logger.NewLogger()
	logger.Info("Starting interactive real-time monitoring dashboard")

	// Create system monitor
	systemMonitor := health.NewSystemMonitor()
	if err := systemMonitor.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start system monitor: %w", err)
	}
	defer systemMonitor.Stop()

	// Create real-time monitor
	rtMonitor := health.NewRealTimeMonitor(systemMonitor)
	if err := rtMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start real-time monitor: %w", err)
	}
	defer rtMonitor.Stop()

	// Subscribe to updates
	updates, err := rtMonitor.Subscribe("interactive_dashboard")
	if err != nil {
		return fmt.Errorf("failed to subscribe to updates: %w", err)
	}
	defer rtMonitor.Unsubscribe("interactive_dashboard")

	var alertChan <-chan *health.SystemAlert
	if enableAlerts {
		alertChan, err = rtMonitor.SubscribeToAlerts("interactive_dashboard")
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to subscribe to alerts: %v", err))
		} else {
			defer rtMonitor.UnsubscribeFromAlerts("interactive_dashboard")
		}
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Clear screen and hide cursor
	fmt.Print("\033[2J\033[H\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	logger.Info("Interactive dashboard started. Press Ctrl+C to exit.")

	// Main display loop
	for {
		select {
		case snapshot := <-updates:
			displayDashboard(snapshot, enableAlerts, enableAnomalies)

		case alert := <-alertChan:
			if alert != nil {
				displayAlert(alert)
			}

		case <-sigChan:
			fmt.Print("\033[2J\033[H\033[?25h")
			logger.Info("Stopping interactive dashboard")
			return nil
		}
	}
}

func runMetricsStream(format string, metrics []string, interval time.Duration, maxUpdates int, showTrends, showAlerts, compact bool) error {
	logger := logger.NewLogger()
	logger.Info("Starting metrics streaming")

	// Create system monitor
	systemMonitor := health.NewSystemMonitor()
	if err := systemMonitor.Start(context.Background()); err != nil {
		return fmt.Errorf("failed to start system monitor: %w", err)
	}
	defer systemMonitor.Stop()

	// Create real-time monitor
	rtMonitor := health.NewRealTimeMonitor(systemMonitor)
	if err := rtMonitor.Start(); err != nil {
		return fmt.Errorf("failed to start real-time monitor: %w", err)
	}
	defer rtMonitor.Stop()

	// Subscribe to updates
	updates, err := rtMonitor.Subscribe("metrics_stream")
	if err != nil {
		return fmt.Errorf("failed to subscribe to updates: %w", err)
	}
	defer rtMonitor.Unsubscribe("metrics_stream")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	updateCount := 0
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Filter metrics if specified
	metricsFilter := make(map[string]bool)
	if len(metrics) > 0 {
		for _, metric := range metrics {
			metricsFilter[metric] = true
		}
	}

	for {
		select {
		case snapshot := <-updates:
			if updateCount%int(interval.Seconds()) == 0 { // Respect interval
				outputMetrics(snapshot, format, metricsFilter, showTrends, showAlerts, compact)
				updateCount++

				if maxUpdates > 0 && updateCount >= maxUpdates {
					return nil
				}
			}

		case <-sigChan:
			logger.Info("Stopping metrics stream")
			return nil

		case <-ticker.C:
			// Continue - updates come from the subscription
		}
	}
}

func displayDashboard(snapshot *health.SystemSnapshot, enableAlerts, enableAnomalies bool) {
	// Move cursor to top
	fmt.Print("\033[H")

	// Header
	fmt.Printf("╔══════════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  🖥️  NixAI Real-Time System Monitor                          %s  ║\n", 
		snapshot.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("╠══════════════════════════════════════════════════════════════════════════════════════╣\n")

	// System Metrics
	fmt.Printf("║ System Metrics                                                                       ║\n")
	fmt.Printf("║ ──────────────                                                                       ║\n")
	
	// Display key metrics
	for metric, value := range snapshot.Metrics {
		status := getMetricStatusIcon(metric, value)
		fmt.Printf("║ %-15s %s %6.1f%% │", metric, status, value)
		
		// Add trend indicator if available
		if trend, exists := snapshot.TrendIndicators[metric]; exists {
			trendIcon := getTrendIcon(trend.Direction)
			fmt.Printf(" %s %+.1f/min", trendIcon, trend.ChangeRate)
		}
		fmt.Printf("                               ║\n")
	}

	// Component Health
	fmt.Printf("║                                                                                      ║\n")
	fmt.Printf("║ Component Health                                                                     ║\n")
	fmt.Printf("║ ────────────────                                                                     ║\n")
	
	for component, status := range snapshot.ComponentHealth {
		statusIcon := getHealthStatusIcon(status)
		fmt.Printf("║ %-15s %s %-10s                                                        ║\n", 
			component, statusIcon, status)
	}

	// Active Alerts
	if enableAlerts && len(snapshot.Alerts) > 0 {
		fmt.Printf("║                                                                                      ║\n")
		fmt.Printf("║ Active Alerts (%d)                                                                ║\n", len(snapshot.Alerts))
		fmt.Printf("║ ─────────────                                                                     ║\n")
		
		for i, alert := range snapshot.Alerts {
			if i >= 3 { // Limit to 3 alerts in dashboard view
				fmt.Printf("║ ... and %d more alerts                                                           ║\n", len(snapshot.Alerts)-3)
				break
			}
			
			severityIcon := getSeverityIcon(alert.Severity)
			duration := time.Since(alert.FirstDetected).Truncate(time.Second)
			fmt.Printf("║ %s %-12s %-50s %8s ║\n", 
				severityIcon, alert.Severity, truncateStringRealtime(alert.Message, 50), duration)
		}
	}

	// System Load
	fmt.Printf("║                                                                                      ║\n")
	fmt.Printf("║ System Load                                                                          ║\n")
	fmt.Printf("║ ───────────                                                                          ║\n")
	fmt.Printf("║ Load Avg:  %.2f  %.2f  %.2f  │  CPU Cores: %d  │  Utilization: %.1f%%              ║\n",
		snapshot.SystemLoad.LoadAvg1, snapshot.SystemLoad.LoadAvg5, snapshot.SystemLoad.LoadAvg15,
		snapshot.SystemLoad.CPUCores, snapshot.SystemLoad.CPUUtilization)

	// Process Info
	fmt.Printf("║ Processes: %-4d total │ %-4d running │ %-4d sleeping │ %-4d zombie              ║\n",
		snapshot.ProcessInfo.TotalProcesses, snapshot.ProcessInfo.RunningProcesses,
		snapshot.ProcessInfo.SleepingProcesses, snapshot.ProcessInfo.ZombieProcesses)

	fmt.Printf("╚══════════════════════════════════════════════════════════════════════════════════════╝\n")
	
	// Instructions
	fmt.Printf("\n💡 Press Ctrl+C to exit  │  Updates every 1 second  │  Sequence: #%d\n", 
		snapshot.SequenceNumber)
}

func displayAlert(alert *health.SystemAlert) {
	// Flash alert at bottom of screen
	fmt.Printf("\n🚨 ALERT: %s - %s (%.1f%%) - %s\n", 
		alert.Severity, alert.Metric, alert.CurrentValue, alert.Message)
}

func outputMetrics(snapshot *health.SystemSnapshot, format string, metricsFilter map[string]bool, showTrends, showAlerts, compact bool) {
	switch format {
	case "json":
		outputMetricsJSON(snapshot, metricsFilter, showTrends, showAlerts)
	case "csv":
		outputMetricsCSV(snapshot, metricsFilter, showTrends)
	case "prometheus":
		outputMetricsPrometheus(snapshot, metricsFilter)
	default: // table
		outputMetricsTable(snapshot, metricsFilter, showTrends, showAlerts, compact)
	}
}

func outputMetricsJSON(snapshot *health.SystemSnapshot, metricsFilter map[string]bool, showTrends, showAlerts bool) {
	output := map[string]interface{}{
		"timestamp": snapshot.Timestamp,
		"metrics":   snapshot.Metrics,
	}

	// Filter metrics if specified
	if len(metricsFilter) > 0 {
		filteredMetrics := make(map[string]float64)
		for metric, value := range snapshot.Metrics {
			if metricsFilter[metric] {
				filteredMetrics[metric] = value
			}
		}
		output["metrics"] = filteredMetrics
	}

	if showTrends {
		output["trends"] = snapshot.TrendIndicators
	}

	if showAlerts && len(snapshot.Alerts) > 0 {
		output["alerts"] = snapshot.Alerts
	}

	jsonData, _ := json.Marshal(output)
	fmt.Println(string(jsonData))
}

func outputMetricsCSV(snapshot *health.SystemSnapshot, metricsFilter map[string]bool, showTrends bool) {
	// Header (only on first call - this is simplified)
	var headerPrinted bool
	if !headerPrinted {
		fmt.Print("timestamp")
		
		var metrics []string
		for metric := range snapshot.Metrics {
			if len(metricsFilter) == 0 || metricsFilter[metric] {
				metrics = append(metrics, metric)
			}
		}
		sort.Strings(metrics)
		
		for _, metric := range metrics {
			fmt.Printf(",%s", metric)
			if showTrends {
				fmt.Printf(",%s_trend", metric)
			}
		}
		fmt.Println()
		headerPrinted = true
	}

	// Data
	fmt.Printf("%s", snapshot.Timestamp.Format(time.RFC3339))
	
	var metrics []string
	for metric := range snapshot.Metrics {
		if len(metricsFilter) == 0 || metricsFilter[metric] {
			metrics = append(metrics, metric)
		}
	}
	sort.Strings(metrics)
	
	for _, metric := range metrics {
		fmt.Printf(",%.2f", snapshot.Metrics[metric])
		if showTrends {
			if trend, exists := snapshot.TrendIndicators[metric]; exists {
				fmt.Printf(",%.2f", trend.ChangeRate)
			} else {
				fmt.Print(",0.0")
			}
		}
	}
	fmt.Println()
}

func outputMetricsPrometheus(snapshot *health.SystemSnapshot, metricsFilter map[string]bool) {
	for metric, value := range snapshot.Metrics {
		if len(metricsFilter) == 0 || metricsFilter[metric] {
			metricName := strings.ReplaceAll(metric, " ", "_")
			fmt.Printf("nixai_%s %.2f %d\n", metricName, value, snapshot.Timestamp.Unix())
		}
	}
}

func outputMetricsTable(snapshot *health.SystemSnapshot, metricsFilter map[string]bool, showTrends, showAlerts, compact bool) {
	if !compact {
		fmt.Printf("\n📊 System Metrics - %s\n", snapshot.Timestamp.Format("15:04:05"))
		fmt.Printf("┌─────────────────┬─────────┬─────────┬────────────┐\n")
		fmt.Printf("│ Metric          │ Value   │ Status  │ Trend      │\n")
		fmt.Printf("├─────────────────┼─────────┼─────────┼────────────┤\n")
	}

	var metrics []string
	for metric := range snapshot.Metrics {
		if len(metricsFilter) == 0 || metricsFilter[metric] {
			metrics = append(metrics, metric)
		}
	}
	sort.Strings(metrics)

	for _, metric := range metrics {
		value := snapshot.Metrics[metric]
		status := getMetricStatusIcon(metric, value)
		
		var trendStr string
		if showTrends {
			if trend, exists := snapshot.TrendIndicators[metric]; exists {
				trendIcon := getTrendIcon(trend.Direction)
				trendStr = fmt.Sprintf("%s %+.1f/m", trendIcon, trend.ChangeRate)
			} else {
				trendStr = "stable"
			}
		}

		if compact {
			fmt.Printf("%-15s %s %6.1f%% %s\n", metric, status, value, trendStr)
		} else {
			fmt.Printf("│ %-15s │ %6.1f%% │ %-7s │ %-10s │\n", 
				metric, value, status, trendStr)
		}
	}

	if !compact {
		fmt.Printf("└─────────────────┴─────────┴─────────┴────────────┘\n")
	}

	if showAlerts && len(snapshot.Alerts) > 0 {
		fmt.Printf("\n🚨 Active Alerts (%d):\n", len(snapshot.Alerts))
		for _, alert := range snapshot.Alerts {
			severityIcon := getSeverityIcon(alert.Severity)
			fmt.Printf("  %s %s: %s (%.1f%%)\n", 
				severityIcon, alert.Severity, alert.Message, alert.CurrentValue)
		}
	}
}

// Helper functions for formatting and display

func getMetricStatusIcon(metric string, value float64) string {
	// Simplified thresholds for display
	var critical, warning float64
	
	switch metric {
	case "cpu_usage", "memory_usage":
		warning, critical = 75.0, 90.0
	case "disk_usage":
		warning, critical = 80.0, 90.0
	case "load_average":
		warning, critical = 6.0, 12.0 // Assuming 8-core system
	default:
		warning, critical = 70.0, 85.0
	}

	if value >= critical {
		return "🔴"
	} else if value >= warning {
		return "🟡"
	}
	return "🟢"
}

func getTrendIcon(direction string) string {
	switch direction {
	case "increasing":
		return "↗️"
	case "decreasing":
		return "↘️"
	default:
		return "→"
	}
}

func getHealthStatusIcon(status health.HealthStatus) string {
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

func getSeverityIcon(severity health.AlertSeverity) string {
	switch severity {
	case health.AlertSeverityEmergency:
		return "🚨"
	case health.AlertSeverityCritical:
		return "🔴"
	case health.AlertSeverityWarning:
		return "🟡"
	default:
		return "ℹ️"
	}
}

func truncateStringRealtime(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Additional command implementations (simplified for now)

func listActiveAlerts(format, severity string, showAll bool) error {
	fmt.Println("Active alerts listing - implementation pending")
	return nil
}

func acknowledgeAlert(alertID, acknowledgedBy string) error {
	fmt.Printf("Acknowledging alert %s by %s - implementation pending\n", alertID, acknowledgedBy)
	return nil
}

func streamAlerts(format, severity string) error {
	fmt.Println("Alert streaming - implementation pending")
	return nil
}

func listThresholds(format string) error {
	fmt.Println("Threshold listing - implementation pending")
	return nil
}

func setThresholds(metric string, warning, critical, emergency float64) error {
	fmt.Printf("Setting thresholds for %s: warn=%.1f, crit=%.1f, emerg=%.1f - implementation pending\n", 
		metric, warning, critical, emergency)
	return nil
}

func showMonitoringStats(format string) error {
	fmt.Println("Monitoring statistics - implementation pending")
	return nil
}