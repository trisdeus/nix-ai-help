package cli

import (
	"fmt"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/internal/performance"
	"nix-ai-help/pkg/errors"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// GlobalErrorManager holds the global error manager instance
var GlobalErrorManager *errors.ErrorManager

// GlobalPerformanceMonitor holds the global performance monitor instance
var GlobalPerformanceMonitor *performance.Monitor

// InitializeGlobalErrorHandling sets up global error handling and performance monitoring
func InitializeGlobalErrorHandling() error {
	// Load configuration
	cfg, err := config.LoadUserConfig()
	if err != nil {
		// Use defaults if config loading fails
		cfg = &config.UserConfig{
			LogLevel: "info",
		}
	}

	// Initialize logger
	log := logger.NewLoggerWithLevel(cfg.LogLevel)

	// Initialize performance monitor
	GlobalPerformanceMonitor = performance.NewMonitor(log)

	// Setup error manager configuration
	debugMode := cfg.LogLevel == "debug" || cfg.LogLevel == "trace"
	analyticsDir := utils.GetAnalyticsDir()

	errorManagerConfig := &errors.ErrorManagerConfig{
		DebugMode:           debugMode,
		GracefulDegradation: true,
		AnalyticsEnabled:    true,
		AnalyticsDataDir:    analyticsDir,
		RetryConfig:         errors.DefaultRetryConfig(),
		MaxLastErrors:       100,
	}

	// Initialize global error manager
	GlobalErrorManager = errors.NewErrorManager(errorManagerConfig)

	// Setup error analytics integration with performance monitoring
	go errorAnalyticsWorker()

	log.Debug("Global error handling and performance monitoring initialized")
	return nil
}

// errorAnalyticsWorker runs in background to correlate errors with performance metrics
func errorAnalyticsWorker() {
	if GlobalErrorManager == nil || GlobalPerformanceMonitor == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			analyzeErrorPerformanceCorrelation()
		}
	}
}

// analyzeErrorPerformanceCorrelation looks for patterns between errors and performance
func analyzeErrorPerformanceCorrelation() {
	if GlobalErrorManager == nil || GlobalPerformanceMonitor == nil {
		return
	}

	// Get error analytics report
	errorReport := GlobalErrorManager.GetAnalyticsReport()
	if errorReport == nil {
		return
	}

	// Get performance summary
	perfSummary := GlobalPerformanceMonitor.GetSummary()

	// Look for correlation patterns
	if perfSummary.TotalOperations > 0 {
		errorRate := float64(perfSummary.FailedOps) / float64(perfSummary.TotalOperations) * 100

		// Log performance issues if error rate is high
		if errorRate > 20 { // More than 20% error rate
			logger.NewLogger().Warn(fmt.Sprintf(
				"High error rate detected: %.1f%% (Performance may be impacted)",
				errorRate))
		}

		// Track slow operations that might be causing errors
		if perfSummary.AverageDuration > 10*time.Second {
			logger.NewLogger().Warn(fmt.Sprintf(
				"Slow operations detected: Average %v (May lead to timeout errors)",
				perfSummary.AverageDuration))
		}
	}
}

// TrackOperationWithErrorHandling wraps an operation with both error and performance tracking
func TrackOperationWithErrorHandling(
	operationType string,
	metricType performance.MetricType,
	operation func() error,
) error {
	if GlobalPerformanceMonitor == nil {
		// Initialize if not already done
		if err := InitializeGlobalErrorHandling(); err != nil {
			return fmt.Errorf("failed to initialize global error handling: %w", err)
		}
	}

	// Start performance timer
	finishTimer := GlobalPerformanceMonitor.StartTimer(metricType, operationType, nil)

	// Execute operation
	err := operation()

	// Record performance metrics
	finishTimer(err == nil, err)

	// Handle error if one occurred
	if err != nil && GlobalErrorManager != nil {
		// Determine error code based on operation type and error
		errorCode := categorizeError(operationType, err)
		_ = GlobalErrorManager.HandleError(err, errorCode)
	}

	return err
}

// categorizeError determines the appropriate error code based on operation type and error
func categorizeError(operationType string, err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// AI-related operations
	if operationType == "ai_query" || operationType == "ai_generation" {
		if utils.ContainsAny(errStr, []string{"timeout", "deadline", "context canceled"}) {
			return "ErrorCodeAITimeout"
		}
		if utils.ContainsAny(errStr, []string{"authentication", "api key", "unauthorized"}) {
			return "ErrorCodeAIAuthentication"
		}
		if utils.ContainsAny(errStr, []string{"rate limit", "quota", "too many requests"}) {
			return "ErrorCodeAIRateLimit"
		}
		return "ErrorCodeAIProviderError"
	}

	// Network operations
	if utils.ContainsAny(operationType, []string{"mcp", "documentation", "query"}) {
		if utils.ContainsAny(errStr, []string{"connection", "network", "dial", "resolve"}) {
			return "ErrorCodeNetworkConnection"
		}
		if utils.ContainsAny(errStr, []string{"timeout", "deadline"}) {
			return "ErrorCodeNetworkTimeout"
		}
		return "ErrorCodeMCPToolFailure"
	}

	// File system operations
	if utils.ContainsAny(operationType, []string{"file", "cache", "config"}) {
		if utils.ContainsAny(errStr, []string{"permission", "access"}) {
			return "ErrorCodeFileSystemPermissions"
		}
		if utils.ContainsAny(errStr, []string{"not found", "no such file"}) {
			return "ErrorCodeFileSystemNotFound"
		}
		if utils.ContainsAny(errStr, []string{"disk", "space", "full"}) {
			return "ErrorCodeFileSystemDiskFull"
		}
		return "ErrorCodeFileSystemGeneric"
	}

	// NixOS operations
	if utils.ContainsAny(operationType, []string{"nix", "nixos", "build", "package"}) {
		if utils.ContainsAny(errStr, []string{"build", "compilation"}) {
			return "ErrorCodeNixOSBuildFailure"
		}
		if utils.ContainsAny(errStr, []string{"configuration", "syntax"}) {
			return "ErrorCodeNixOSConfigError"
		}
		return "ErrorCodeNixOSGeneric"
	}

	// Default to generic internal error
	return "ErrorCodeInternalGeneric"
}

// GetGlobalErrorReport returns a comprehensive error and performance report
func GetGlobalErrorReport() map[string]interface{} {
	report := make(map[string]interface{})

	if GlobalErrorManager != nil {
		if analyticsReport := GlobalErrorManager.GetAnalyticsReport(); analyticsReport != nil {
			report["error_analytics"] = analyticsReport
		}
	}

	if GlobalPerformanceMonitor != nil {
		perfSummary := GlobalPerformanceMonitor.GetSummary()
		report["performance_summary"] = perfSummary
		report["performance_formatted"] = GlobalPerformanceMonitor.FormatSummary()
	}

	return report
}

// ResetGlobalMetrics clears all global error and performance metrics (useful for testing)
func ResetGlobalMetrics() {
	if GlobalErrorManager != nil {
		GlobalErrorManager.ClearAnalytics()
	}
	if GlobalPerformanceMonitor != nil {
		GlobalPerformanceMonitor.Reset()
	}
}
