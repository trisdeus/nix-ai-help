package workflow

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ErrorHandler handles workflow and task errors
type ErrorHandler struct {
	mu                 sync.RWMutex
	recoveryStrategies map[string]RecoveryStrategy
	errorHistory       []WorkflowError
	logger             Logger
	maxRetries         int
	retryDelay         time.Duration
}

// WorkflowError represents an error during workflow execution
type WorkflowError struct {
	ID          string            `json:"id"`
	WorkflowID  string            `json:"workflow_id"`
	TaskID      string            `json:"task_id,omitempty"`
	ActionID    string            `json:"action_id,omitempty"`
	ErrorType   ErrorType         `json:"error_type"`
	Message     string            `json:"message"`
	Cause       error             `json:"-"`
	Timestamp   time.Time         `json:"timestamp"`
	Severity    ErrorSeverity     `json:"severity"`
	Context     map[string]string `json:"context"`
	StackTrace  string            `json:"stack_trace,omitempty"`
	Recoverable bool              `json:"recoverable"`
	Retryable   bool              `json:"retryable"`
}

// ErrorType defines types of workflow errors
type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "validation"
	ErrorTypeExecution     ErrorType = "execution"
	ErrorTypeDependency    ErrorType = "dependency"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeResource      ErrorType = "resource"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeNetwork       ErrorType = "network"
	ErrorTypePermission    ErrorType = "permission"
	ErrorTypeSystem        ErrorType = "system"
	ErrorTypeUnknown       ErrorType = "unknown"
)

// ErrorSeverity defines error severity levels
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// RecoveryStrategy defines how to recover from specific errors
type RecoveryStrategy struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	ErrorTypes  []ErrorType      `json:"error_types"`
	Actions     []RecoveryAction `json:"actions"`
	Conditions  []string         `json:"conditions"`
	MaxRetries  int              `json:"max_retries"`
	RetryDelay  time.Duration    `json:"retry_delay"`
	Enabled     bool             `json:"enabled"`
}

// RecoveryAction defines a recovery action
type RecoveryAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger Logger) *ErrorHandler {
	eh := &ErrorHandler{
		recoveryStrategies: make(map[string]RecoveryStrategy),
		errorHistory:       make([]WorkflowError, 0),
		logger:             logger,
		maxRetries:         3,
		retryDelay:         5 * time.Second,
	}

	// Initialize default recovery strategies
	eh.initializeDefaultRecoveryStrategies()

	return eh
}

// HandleError handles a workflow error and attempts recovery
func (eh *ErrorHandler) HandleError(ctx context.Context, err error, workflowID, taskID, actionID string) (*RecoveryResult, error) {
	// Create workflow error
	workflowErr := eh.createWorkflowError(err, workflowID, taskID, actionID)

	// Log the error
	eh.logError(workflowErr)

	// Add to error history
	eh.addToHistory(workflowErr)

	// Determine if error is recoverable
	if !workflowErr.Recoverable {
		return &RecoveryResult{
			Success:  false,
			ErrorID:  workflowErr.ID,
			Message:  "Error is not recoverable",
			Strategy: nil,
		}, nil
	}

	// Find applicable recovery strategies
	strategies := eh.findRecoveryStrategies(workflowErr)
	if len(strategies) == 0 {
		return &RecoveryResult{
			Success:  false,
			ErrorID:  workflowErr.ID,
			Message:  "No recovery strategies found",
			Strategy: nil,
		}, nil
	}

	// Attempt recovery with each strategy
	for _, strategy := range strategies {
		if !strategy.Enabled {
			continue
		}

		eh.logger.Info("Attempting recovery with strategy: %s", strategy.Name)

		result, err := eh.executeRecoveryStrategy(ctx, strategy, workflowErr)
		if err != nil {
			eh.logger.Warn("Recovery strategy failed: %s - %v", strategy.Name, err)
			continue
		}

		if result.Success {
			eh.logger.Info("Recovery successful with strategy: %s", strategy.Name)
			return result, nil
		}
	}

	// All recovery attempts failed
	return &RecoveryResult{
		Success:  false,
		ErrorID:  workflowErr.ID,
		Message:  "All recovery strategies failed",
		Strategy: nil,
	}, nil
}

// RecoveryResult represents the result of a recovery attempt
type RecoveryResult struct {
	Success        bool              `json:"success"`
	ErrorID        string            `json:"error_id"`
	Message        string            `json:"message"`
	Strategy       *RecoveryStrategy `json:"strategy,omitempty"`
	ActionsApplied []string          `json:"actions_applied"`
	Duration       time.Duration     `json:"duration"`
	Timestamp      time.Time         `json:"timestamp"`
}

// createWorkflowError creates a WorkflowError from a standard error
func (eh *ErrorHandler) createWorkflowError(err error, workflowID, taskID, actionID string) WorkflowError {
	// Generate stack trace
	stackTrace := eh.captureStackTrace()

	// Determine error type and severity
	errorType := eh.determineErrorType(err)
	severity := eh.determineSeverity(errorType, err)

	return WorkflowError{
		ID:          eh.generateErrorID(),
		WorkflowID:  workflowID,
		TaskID:      taskID,
		ActionID:    actionID,
		ErrorType:   errorType,
		Message:     err.Error(),
		Cause:       err,
		Timestamp:   time.Now(),
		Severity:    severity,
		Context:     eh.gatherErrorContext(),
		StackTrace:  stackTrace,
		Recoverable: eh.isRecoverable(errorType, err),
		Retryable:   eh.isRetryable(errorType, err),
	}
}

// determineErrorType determines the error type from the error
func (eh *ErrorHandler) determineErrorType(err error) ErrorType {
	errMsg := err.Error()

	switch {
	case contains(errMsg, "validation", "invalid", "malformed"):
		return ErrorTypeValidation
	case contains(errMsg, "timeout", "deadline"):
		return ErrorTypeTimeout
	case contains(errMsg, "permission", "access denied", "forbidden"):
		return ErrorTypePermission
	case contains(errMsg, "network", "connection", "dns"):
		return ErrorTypeNetwork
	case contains(errMsg, "dependency", "missing", "not found"):
		return ErrorTypeDependency
	case contains(errMsg, "resource", "memory", "disk", "cpu"):
		return ErrorTypeResource
	case contains(errMsg, "config", "configuration"):
		return ErrorTypeConfiguration
	case contains(errMsg, "system", "kernel", "hardware"):
		return ErrorTypeSystem
	default:
		return ErrorTypeUnknown
	}
}

// determineSeverity determines error severity
func (eh *ErrorHandler) determineSeverity(errorType ErrorType, err error) ErrorSeverity {
	switch errorType {
	case ErrorTypeValidation:
		return SeverityMedium
	case ErrorTypeTimeout:
		return SeverityMedium
	case ErrorTypePermission:
		return SeverityHigh
	case ErrorTypeNetwork:
		return SeverityMedium
	case ErrorTypeDependency:
		return SeverityHigh
	case ErrorTypeResource:
		return SeverityCritical
	case ErrorTypeConfiguration:
		return SeverityHigh
	case ErrorTypeSystem:
		return SeverityCritical
	default:
		return SeverityMedium
	}
}

// isRecoverable determines if an error is recoverable
func (eh *ErrorHandler) isRecoverable(errorType ErrorType, err error) bool {
	switch errorType {
	case ErrorTypeValidation:
		return false // Validation errors usually need manual intervention
	case ErrorTypeTimeout:
		return true // Can retry
	case ErrorTypePermission:
		return true // Can try with different permissions
	case ErrorTypeNetwork:
		return true // Network issues are often temporary
	case ErrorTypeDependency:
		return true // Dependencies can be installed
	case ErrorTypeResource:
		return true // Resources can be freed
	case ErrorTypeConfiguration:
		return true // Configuration can be corrected
	case ErrorTypeSystem:
		return false // System errors usually need manual intervention
	default:
		return true // Default to recoverable
	}
}

// isRetryable determines if an error is retryable
func (eh *ErrorHandler) isRetryable(errorType ErrorType, err error) bool {
	switch errorType {
	case ErrorTypeValidation:
		return false
	case ErrorTypeTimeout:
		return true
	case ErrorTypePermission:
		return false
	case ErrorTypeNetwork:
		return true
	case ErrorTypeDependency:
		return false
	case ErrorTypeResource:
		return true
	case ErrorTypeConfiguration:
		return false
	case ErrorTypeSystem:
		return false
	default:
		return true
	}
}

// findRecoveryStrategies finds applicable recovery strategies for an error
func (eh *ErrorHandler) findRecoveryStrategies(workflowErr WorkflowError) []RecoveryStrategy {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	var strategies []RecoveryStrategy

	for _, strategy := range eh.recoveryStrategies {
		if eh.strategyApplies(strategy, workflowErr) {
			strategies = append(strategies, strategy)
		}
	}

	return strategies
}

// strategyApplies checks if a recovery strategy applies to an error
func (eh *ErrorHandler) strategyApplies(strategy RecoveryStrategy, workflowErr WorkflowError) bool {
	// Check error type match
	for _, errorType := range strategy.ErrorTypes {
		if errorType == workflowErr.ErrorType {
			return true
		}
	}

	return false
}

// executeRecoveryStrategy executes a recovery strategy
func (eh *ErrorHandler) executeRecoveryStrategy(ctx context.Context, strategy RecoveryStrategy, workflowErr WorkflowError) (*RecoveryResult, error) {
	startTime := time.Now()

	result := &RecoveryResult{
		Success:        false,
		ErrorID:        workflowErr.ID,
		Strategy:       &strategy,
		ActionsApplied: make([]string, 0),
		Timestamp:      startTime,
	}

	// Execute recovery actions
	for _, action := range strategy.Actions {
		eh.logger.Debug("Executing recovery action: %s", action.Type)

		actionErr := eh.executeRecoveryAction(ctx, action, workflowErr)
		if actionErr != nil {
			result.Message = fmt.Sprintf("Recovery action failed: %s", actionErr.Error())
			result.Duration = time.Since(startTime)
			return result, nil
		}

		result.ActionsApplied = append(result.ActionsApplied, action.Type)
	}

	result.Success = true
	result.Message = "Recovery completed successfully"
	result.Duration = time.Since(startTime)

	return result, nil
}

// executeRecoveryAction executes a single recovery action
func (eh *ErrorHandler) executeRecoveryAction(ctx context.Context, action RecoveryAction, workflowErr WorkflowError) error {
	// Create context with timeout
	actionCtx := ctx
	if action.Timeout > 0 {
		var cancel context.CancelFunc
		actionCtx, cancel = context.WithTimeout(ctx, action.Timeout)
		defer cancel()
	}

	switch action.Type {
	case "retry":
		return eh.executeRetryAction(actionCtx, action, workflowErr)
	case "cleanup":
		return eh.executeCleanupAction(actionCtx, action, workflowErr)
	case "restart":
		return eh.executeRestartAction(actionCtx, action, workflowErr)
	case "rollback":
		return eh.executeRollbackAction(actionCtx, action, workflowErr)
	default:
		return fmt.Errorf("unknown recovery action type: %s", action.Type)
	}
}

// executeRetryAction executes a retry recovery action
func (eh *ErrorHandler) executeRetryAction(ctx context.Context, action RecoveryAction, workflowErr WorkflowError) error {
	eh.logger.Info("Executing retry recovery action")
	// Implementation would retry the failed operation
	return nil
}

// executeCleanupAction executes a cleanup recovery action
func (eh *ErrorHandler) executeCleanupAction(ctx context.Context, action RecoveryAction, workflowErr WorkflowError) error {
	eh.logger.Info("Executing cleanup recovery action")
	// Implementation would clean up resources
	return nil
}

// executeRestartAction executes a restart recovery action
func (eh *ErrorHandler) executeRestartAction(ctx context.Context, action RecoveryAction, workflowErr WorkflowError) error {
	eh.logger.Info("Executing restart recovery action")
	// Implementation would restart services
	return nil
}

// executeRollbackAction executes a rollback recovery action
func (eh *ErrorHandler) executeRollbackAction(ctx context.Context, action RecoveryAction, workflowErr WorkflowError) error {
	eh.logger.Info("Executing rollback recovery action")
	// Implementation would rollback changes
	return nil
}

// RegisterRecoveryStrategy registers a new recovery strategy
func (eh *ErrorHandler) RegisterRecoveryStrategy(strategy RecoveryStrategy) error {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	if strategy.ID == "" {
		return fmt.Errorf("recovery strategy ID is required")
	}

	eh.recoveryStrategies[strategy.ID] = strategy
	eh.logger.Info("Registered recovery strategy: %s", strategy.Name)

	return nil
}

// initializeDefaultRecoveryStrategies initializes default recovery strategies
func (eh *ErrorHandler) initializeDefaultRecoveryStrategies() {
	// Timeout Recovery Strategy
	eh.recoveryStrategies["timeout-retry"] = RecoveryStrategy{
		ID:          "timeout-retry",
		Name:        "Timeout Retry",
		Description: "Retry operation after timeout",
		ErrorTypes:  []ErrorType{ErrorTypeTimeout},
		Actions: []RecoveryAction{
			{
				Type:        "retry",
				Description: "Retry the failed operation",
				Parameters:  map[string]interface{}{"max_attempts": 3},
				Timeout:     30 * time.Second,
			},
		},
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
		Enabled:    true,
	}

	// Network Recovery Strategy
	eh.recoveryStrategies["network-retry"] = RecoveryStrategy{
		ID:          "network-retry",
		Name:        "Network Retry",
		Description: "Retry network operations",
		ErrorTypes:  []ErrorType{ErrorTypeNetwork},
		Actions: []RecoveryAction{
			{
				Type:        "retry",
				Description: "Retry network operation",
				Parameters:  map[string]interface{}{"backoff": "exponential"},
				Timeout:     60 * time.Second,
			},
		},
		MaxRetries: 5,
		RetryDelay: 2 * time.Second,
		Enabled:    true,
	}

	// Resource Recovery Strategy
	eh.recoveryStrategies["resource-cleanup"] = RecoveryStrategy{
		ID:          "resource-cleanup",
		Name:        "Resource Cleanup",
		Description: "Clean up resources and retry",
		ErrorTypes:  []ErrorType{ErrorTypeResource},
		Actions: []RecoveryAction{
			{
				Type:        "cleanup",
				Description: "Clean up system resources",
				Parameters:  map[string]interface{}{"type": "memory"},
				Timeout:     30 * time.Second,
			},
			{
				Type:        "retry",
				Description: "Retry after cleanup",
				Parameters:  map[string]interface{}{"delay": "5s"},
				Timeout:     60 * time.Second,
			},
		},
		MaxRetries: 2,
		RetryDelay: 10 * time.Second,
		Enabled:    true,
	}

	eh.logger.Info("Initialized %d default recovery strategies", len(eh.recoveryStrategies))
}

// Helper functions

func (eh *ErrorHandler) generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

func (eh *ErrorHandler) captureStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func (eh *ErrorHandler) gatherErrorContext() map[string]string {
	return map[string]string{
		"timestamp":  time.Now().Format(time.RFC3339),
		"goroutines": fmt.Sprintf("%d", runtime.NumGoroutine()),
	}
}

func (eh *ErrorHandler) logError(workflowErr WorkflowError) {
	eh.logger.Error("Workflow error occurred - ID: %s, Type: %s, Message: %s",
		workflowErr.ID, workflowErr.ErrorType, workflowErr.Message)
}

func (eh *ErrorHandler) addToHistory(workflowErr WorkflowError) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	eh.errorHistory = append(eh.errorHistory, workflowErr)

	// Keep only the last 1000 errors
	if len(eh.errorHistory) > 1000 {
		eh.errorHistory = eh.errorHistory[len(eh.errorHistory)-1000:]
	}
}

func contains(text string, keywords ...string) bool {
	textLower := fmt.Sprintf("%s", text)
	for _, keyword := range keywords {
		if len(textLower) > 0 && len(keyword) > 0 {
			// Simple contains check
			for i := 0; i <= len(textLower)-len(keyword); i++ {
				if textLower[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// GetErrorHistory returns recent error history
func (eh *ErrorHandler) GetErrorHistory(limit int) []WorkflowError {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	if limit <= 0 || limit > len(eh.errorHistory) {
		limit = len(eh.errorHistory)
	}

	start := len(eh.errorHistory) - limit
	if start < 0 {
		start = 0
	}

	result := make([]WorkflowError, limit)
	copy(result, eh.errorHistory[start:])

	return result
}

// GetRecoveryStrategies returns all recovery strategies
func (eh *ErrorHandler) GetRecoveryStrategies() []RecoveryStrategy {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	strategies := make([]RecoveryStrategy, 0, len(eh.recoveryStrategies))
	for _, strategy := range eh.recoveryStrategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}
