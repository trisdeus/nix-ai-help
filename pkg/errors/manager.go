package errors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrorManager is the central error handling system for nixai
type ErrorManager struct {
	mu                     sync.RWMutex
	analytics              *ErrorAnalytics
	messageGenerator       *UserFriendlyMessageGenerator
	enhancedCommunicator   *EnhancedErrorCommunicator
	panicHandler           *PanicRecoveryHandler
	retryManager           *RetryManager
	debugMode              bool
	gracefulDegradation    bool
	fallbackHandlers       map[ErrorCategory]FallbackHandler
	lastErrors             []LastError
	maxLastErrors          int
	contextualErrorsEnabled bool
}

// FallbackHandler defines a handler for graceful degradation
type FallbackHandler func(err *NixAIError) error

// LastError stores information about recent errors
type LastError struct {
	Error     *NixAIError `json:"error"`
	Timestamp time.Time   `json:"timestamp"`
	Context   string      `json:"context"`
	Resolved  bool        `json:"resolved"`
}

// ErrorManagerConfig configures the error manager
type ErrorManagerConfig struct {
	DebugMode              bool                              `json:"debug_mode"`
	GracefulDegradation    bool                              `json:"graceful_degradation"`
	AnalyticsEnabled       bool                              `json:"analytics_enabled"`
	AnalyticsDataDir       string                            `json:"analytics_data_dir"`
	RetryConfig            *RetryConfig                      `json:"retry_config"`
	MaxLastErrors          int                               `json:"max_last_errors"`
	FallbackHandlers       map[ErrorCategory]FallbackHandler `json:"-"`
	ContextualErrorsEnabled bool                             `json:"contextual_errors_enabled"`
}

// DefaultErrorManagerConfig returns a default configuration
func DefaultErrorManagerConfig() *ErrorManagerConfig {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".config", "nixai", "error_analytics")

	return &ErrorManagerConfig{
		DebugMode:              false,
		GracefulDegradation:    true,
		AnalyticsEnabled:       true,
		AnalyticsDataDir:       dataDir,
		RetryConfig:            DefaultRetryConfig(),
		MaxLastErrors:          50,
		FallbackHandlers:       make(map[ErrorCategory]FallbackHandler),
		ContextualErrorsEnabled: true,
	}
}

// NewErrorManager creates a new error manager
func NewErrorManager(config *ErrorManagerConfig) *ErrorManager {
	if config == nil {
		config = DefaultErrorManagerConfig()
	}

	var analytics *ErrorAnalytics
	if config.AnalyticsEnabled {
		analytics = NewErrorAnalytics(config.AnalyticsDataDir)
	}

	manager := &ErrorManager{
		analytics:              analytics,
		messageGenerator:       NewUserFriendlyMessageGenerator(),
		enhancedCommunicator:   NewEnhancedErrorCommunicator(),
		panicHandler:           NewPanicRecoveryHandler(nil),
		retryManager:           NewRetryManager(config.RetryConfig),
		debugMode:              config.DebugMode,
		gracefulDegradation:    config.GracefulDegradation,
		fallbackHandlers:       make(map[ErrorCategory]FallbackHandler),
		lastErrors:             make([]LastError, 0, config.MaxLastErrors),
		maxLastErrors:          config.MaxLastErrors,
		contextualErrorsEnabled: config.ContextualErrorsEnabled,
	}

	// Copy provided fallback handlers if any
	if config.FallbackHandlers != nil {
		for category, handler := range config.FallbackHandlers {
			manager.fallbackHandlers[category] = handler
		}
	}

	// Set up default fallback handlers
	manager.setupDefaultFallbackHandlers()

	return manager
}

// HandleError is the main error handling entry point
func (em *ErrorManager) HandleError(err error, context string) *NixAIError {
	if err == nil {
		return nil
	}

	em.mu.Lock()
	defer em.mu.Unlock()

	var nixaiErr *NixAIError

	// Convert to NixAIError if needed
	if ne, ok := err.(*NixAIError); ok {
		nixaiErr = ne
	} else {
		nixaiErr = em.convertToNixAIError(err)
	}

	// Record the error
	em.recordError(nixaiErr, context)

	// Apply graceful degradation if enabled
	if em.gracefulDegradation {
		if fallbackErr := em.applyGracefulDegradation(nixaiErr); fallbackErr != nil {
			// If fallback failed, record that too
			em.recordError(fallbackErr, context+"_fallback")
		}
	}

	return nixaiErr
}

// HandleErrorWithRetry handles an error with automatic retry logic
func (em *ErrorManager) HandleErrorWithRetry(operation RetryableOperation, context string) (interface{}, error) {
	result := em.retryManager.Execute(operation)

	if !result.Success {
		finalError := em.HandleError(result.FinalError, context)
		if em.analytics != nil {
			// Record retry statistics
			em.analytics.RecordError(finalError, fmt.Sprintf("%s_retry_failed_after_%d_attempts", context, len(result.Attempts)))
		}
		return nil, finalError
	}

	// Record successful retry if there were failed attempts
	if len(result.Attempts) > 1 && em.analytics != nil {
		em.analytics.RecordResolution(ErrorCodeNetworkTimeout, fmt.Sprintf("retry_successful_after_%d_attempts", len(result.Attempts)))
	}

	return result.Result, nil
}

// HandleErrorWithContext handles an error with context support
func (em *ErrorManager) HandleErrorWithContext(ctx context.Context, operation RetryableOperationWithContext, context string) (interface{}, error) {
	result := em.retryManager.ExecuteWithContext(ctx, operation)

	if !result.Success {
		finalError := em.HandleError(result.FinalError, context)
		return nil, finalError
	}

	return result.Result, nil
}

// HandlePanic handles panic recovery
func (em *ErrorManager) HandlePanic(recoveryPoint string) *NixAIError {
	return em.panicHandler.Recover(recoveryPoint)
}

// WrapWithRecovery wraps a function with error handling and panic recovery
func (em *ErrorManager) WrapWithRecovery(fn func() error, context string) func() error {
	return func() error {
		defer func() {
			if panicErr := em.HandlePanic(context); panicErr != nil {
				// Panic was recovered, log it
				if em.analytics != nil {
					em.analytics.RecordError(panicErr, context)
				}
			}
		}()

		if err := fn(); err != nil {
			return em.HandleError(err, context)
		}
		return nil
	}
}

// GetUserFriendlyMessage returns a user-friendly error message
func (em *ErrorManager) GetUserFriendlyMessage(err error) string {
	return em.messageGenerator.GenerateUserFriendlyMessage(err)
}

// GetEnhancedErrorMessage returns an AI-powered enhanced error message with context
func (em *ErrorManager) GetEnhancedErrorMessage(ctx context.Context, err error, context string) string {
	if !em.contextualErrorsEnabled || em.enhancedCommunicator == nil {
		return em.GetUserFriendlyMessage(err)
	}
	
	return em.enhancedCommunicator.GenerateEnhancedMessage(ctx, err, context)
}

// GetFormattedError returns a formatted error for display
func (em *ErrorManager) GetFormattedError(err error) string {
	return FormatErrorForDisplay(err, em.debugMode)
}

// GetFormattedErrorWithContext returns a formatted error with enhanced context
func (em *ErrorManager) GetFormattedErrorWithContext(ctx context.Context, err error, context string) string {
	if !em.contextualErrorsEnabled {
		return em.GetFormattedError(err)
	}
	
	enhanced := em.GetEnhancedErrorMessage(ctx, err, context)
	
	if em.debugMode {
		enhanced += "\n\n" + FormatErrorForDisplay(err, true)
	}
	
	return enhanced
}

// EnableContextualErrors enables or disables contextual error messages
func (em *ErrorManager) EnableContextualErrors(enabled bool) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.contextualErrorsEnabled = enabled
}

// IsContextualErrorsEnabled returns whether contextual errors are enabled
func (em *ErrorManager) IsContextualErrorsEnabled() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.contextualErrorsEnabled
}

// RecordResolution records how an error was resolved
func (em *ErrorManager) RecordResolution(errorCode ErrorCode, resolution string) {
	if em.analytics != nil {
		em.analytics.RecordResolution(errorCode, resolution)
	}

	// Mark recent errors as resolved
	em.mu.Lock()
	defer em.mu.Unlock()

	for i := range em.lastErrors {
		if em.lastErrors[i].Error.Code == errorCode && !em.lastErrors[i].Resolved {
			em.lastErrors[i].Resolved = true
			break
		}
	}
}

// GetLastErrors returns recent errors
func (em *ErrorManager) GetLastErrors(limit int) []LastError {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if limit <= 0 || limit > len(em.lastErrors) {
		limit = len(em.lastErrors)
	}

	result := make([]LastError, limit)
	copy(result, em.lastErrors[len(em.lastErrors)-limit:])
	return result
}

// GetAnalyticsReport returns an analytics report
func (em *ErrorManager) GetAnalyticsReport() *AnalyticsReport {
	if em.analytics == nil {
		return nil
	}
	return em.analytics.GenerateReport()
}

// ExportAnalytics exports analytics data to a file
func (em *ErrorManager) ExportAnalytics(filename string) error {
	if em.analytics == nil {
		return NewError(ErrorCodeInternalError, "Analytics not enabled").Build()
	}
	return em.analytics.ExportReport(filename)
}

// SetDebugMode enables or disables debug mode
func (em *ErrorManager) SetDebugMode(enabled bool) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.debugMode = enabled
}

// IsDebugMode returns whether debug mode is enabled
func (em *ErrorManager) IsDebugMode() bool {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.debugMode
}

// AddFallbackHandler adds a fallback handler for a specific error category
func (em *ErrorManager) AddFallbackHandler(category ErrorCategory, handler FallbackHandler) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.fallbackHandlers[category] = handler
}

// ClearAnalytics clears all analytics data
func (em *ErrorManager) ClearAnalytics() error {
	em.mu.Lock()
	defer em.mu.Unlock()

	if em.analytics == nil {
		return NewError(ErrorCodeInternalError, "Analytics not enabled").Build()
	}

	// Clear analytics data
	if err := em.analytics.ClearData(); err != nil {
		return NewError(ErrorCodeInternalError, "Failed to clear analytics data").
			WithCause(err).
			Build()
	}

	// Clear last errors as well
	em.lastErrors = nil

	return nil
}

// convertToNixAIError converts a generic error to NixAIError
func (em *ErrorManager) convertToNixAIError(err error) *NixAIError {
	// Try to infer error code and category from message
	errorMsg := err.Error()

	if contains(errorMsg, "timeout") {
		return NewError(ErrorCodeNetworkTimeout, errorMsg).
			WithCause(err).
			WithRetryable(true).
			Build()
	}

	if contains(errorMsg, "connection refused") {
		return NewError(ErrorCodeConnectionRefused, errorMsg).
			WithCause(err).
			WithRetryable(true).
			Build()
	}

	if contains(errorMsg, "permission denied") {
		return NewError(ErrorCodePermissionDenied, errorMsg).
			WithCause(err).
			Build()
	}

	if contains(errorMsg, "file not found") || contains(errorMsg, "no such file") {
		return NewError(ErrorCodeFileNotFound, errorMsg).
			WithCause(err).
			Build()
	}

	// Default to unknown error
	return NewError(ErrorCodeUnknown, errorMsg).
		WithCause(err).
		WithSeverity(SeverityLow).
		WithCategory(CategoryInternal).
		Build()
}

// recordError records an error in analytics and last errors
func (em *ErrorManager) recordError(nixaiErr *NixAIError, context string) {
	// Record in analytics
	if em.analytics != nil {
		em.analytics.RecordError(nixaiErr, context)
	}

	// Add to last errors
	lastError := LastError{
		Error:     nixaiErr,
		Timestamp: time.Now(),
		Context:   context,
		Resolved:  false,
	}

	em.lastErrors = append(em.lastErrors, lastError)

	// Trim to max size
	if len(em.lastErrors) > em.maxLastErrors {
		em.lastErrors = em.lastErrors[1:]
	}
}

// applyGracefulDegradation applies graceful degradation if available
func (em *ErrorManager) applyGracefulDegradation(nixaiErr *NixAIError) *NixAIError {
	if handler, exists := em.fallbackHandlers[nixaiErr.Category]; exists {
		if fallbackErr := handler(nixaiErr); fallbackErr != nil {
			return em.convertToNixAIError(fallbackErr)
		}
	}
	return nil
}

// setupDefaultFallbackHandlers sets up default fallback handlers
func (em *ErrorManager) setupDefaultFallbackHandlers() {
	// Network category fallback
	em.fallbackHandlers[CategoryNetwork] = func(err *NixAIError) error {
		// For network errors, we might try alternative endpoints or offline mode
		return nil // No fallback error means graceful degradation succeeded
	}

	// AI category fallback
	em.fallbackHandlers[CategoryAI] = func(err *NixAIError) error {
		// For AI errors, we might fall back to local processing or cached responses
		return nil
	}

	// MCP category fallback
	em.fallbackHandlers[CategoryMCP] = func(err *NixAIError) error {
		// For MCP errors, we might fall back to direct documentation access
		return nil
	}

	// Cache category fallback
	em.fallbackHandlers[CategoryCache] = func(err *NixAIError) error {
		// For cache errors, we might rebuild cache or operate without cache
		return nil
	}
}

// ValidateConfiguration validates the error manager configuration
func ValidateConfiguration(config *ErrorManagerConfig) error {
	if config == nil {
		return NewError(ErrorCodeInvalidInput, "Configuration cannot be nil").Build()
	}

	if config.AnalyticsEnabled && config.AnalyticsDataDir == "" {
		return NewError(ErrorCodeInvalidInput, "Analytics data directory must be specified when analytics is enabled").Build()
	}

	if config.MaxLastErrors <= 0 {
		return NewError(ErrorCodeInvalidInput, "MaxLastErrors must be positive").Build()
	}

	return nil
}

// Global error manager instance
var globalErrorManager *ErrorManager
var globalErrorManagerOnce sync.Once

// GetGlobalErrorManager returns the global error manager instance
func GetGlobalErrorManager() *ErrorManager {
	globalErrorManagerOnce.Do(func() {
		globalErrorManager = NewErrorManager(DefaultErrorManagerConfig())
	})
	return globalErrorManager
}

// InitializeGlobalErrorManager initializes the global error manager with custom config
func InitializeGlobalErrorManager(config *ErrorManagerConfig) error {
	if err := ValidateConfiguration(config); err != nil {
		return err
	}

	globalErrorManagerOnce.Do(func() {
		globalErrorManager = NewErrorManager(config)
	})

	return nil
}

// Convenience functions for global error manager

// Handle handles an error using the global error manager
func Handle(err error, context string) *NixAIError {
	return GetGlobalErrorManager().HandleError(err, context)
}

// HandleWithRetry handles an error with retry using the global error manager
func HandleWithRetry(operation RetryableOperation, context string) (interface{}, error) {
	return GetGlobalErrorManager().HandleErrorWithRetry(operation, context)
}

// HandleWithContext handles an error with context using the global error manager
func HandleWithContext(ctx context.Context, operation RetryableOperationWithContext, context string) (interface{}, error) {
	return GetGlobalErrorManager().HandleErrorWithContext(ctx, operation, context)
}

// Recover handles panic recovery using the global error manager
func Recover(recoveryPoint string) *NixAIError {
	return GetGlobalErrorManager().HandlePanic(recoveryPoint)
}

// FormatForUser formats an error for user display using the global error manager
func FormatForUser(err error) string {
	return GetGlobalErrorManager().GetFormattedError(err)
}

// RecordSuccess records successful resolution of an error
func RecordSuccess(errorCode ErrorCode, resolution string) {
	GetGlobalErrorManager().RecordResolution(errorCode, resolution)
}
