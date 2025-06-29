package errors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig defines configuration for retry behavior
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	JitterEnabled   bool          `json:"jitter_enabled"`
	JitterFactor    float64       `json:"jitter_factor"`
	RetryableErrors []ErrorCode   `json:"retryable_errors"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: true,
		JitterFactor:  0.1,
		RetryableErrors: []ErrorCode{
			ErrorCodeNetworkTimeout,
			ErrorCodeConnectionRefused,
			ErrorCodeAIProviderUnavailable,
			ErrorCodeAIRateLimited,
			ErrorCodeNixChannelUnavailable,
			ErrorCodeMCPServerUnavailable,
			ErrorCodeCacheUnavailable,
		},
	}
}

// AggressiveRetryConfig returns a more aggressive retry configuration
func AggressiveRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   5,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 1.5,
		JitterEnabled: true,
		JitterFactor:  0.2,
		RetryableErrors: []ErrorCode{
			ErrorCodeNetworkTimeout,
			ErrorCodeConnectionRefused,
			ErrorCodeAIProviderUnavailable,
			ErrorCodeAIRateLimited,
			ErrorCodeNixChannelUnavailable,
			ErrorCodeMCPServerUnavailable,
			ErrorCodeCacheUnavailable,
			ErrorCodeNixBuildFailed,
		},
	}
}

// ConservativeRetryConfig returns a conservative retry configuration
func ConservativeRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  2 * time.Second,
		MaxDelay:      15 * time.Second,
		BackoffFactor: 2.0,
		JitterEnabled: false,
		JitterFactor:  0.0,
		RetryableErrors: []ErrorCode{
			ErrorCodeNetworkTimeout,
			ErrorCodeAIRateLimited,
		},
	}
}

// RetryAttempt represents information about a single retry attempt
type RetryAttempt struct {
	AttemptNumber int           `json:"attempt_number"`
	Delay         time.Duration `json:"delay"`
	Error         error         `json:"error"`
	Timestamp     time.Time     `json:"timestamp"`
}

// RetryResult contains the result of a retry operation
type RetryResult struct {
	Success       bool           `json:"success"`
	Attempts      []RetryAttempt `json:"attempts"`
	TotalDuration time.Duration  `json:"total_duration"`
	FinalError    error          `json:"final_error"`
	Result        interface{}    `json:"result,omitempty"`
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func() (interface{}, error)

// RetryableOperationWithContext represents an operation that can be retried with context
type RetryableOperationWithContext func(ctx context.Context) (interface{}, error)

// RetryManager manages retry operations with exponential backoff
type RetryManager struct {
	config *RetryConfig
}

// NewRetryManager creates a new retry manager with the given configuration
func NewRetryManager(config *RetryConfig) *RetryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryManager{config: config}
}

// Execute executes an operation with retry logic
func (rm *RetryManager) Execute(operation RetryableOperation) *RetryResult {
	return rm.ExecuteWithContext(context.Background(), func(ctx context.Context) (interface{}, error) {
		return operation()
	})
}

// ExecuteWithContext executes an operation with retry logic and context support
func (rm *RetryManager) ExecuteWithContext(ctx context.Context, operation RetryableOperationWithContext) *RetryResult {
	startTime := time.Now()
	result := &RetryResult{
		Attempts: make([]RetryAttempt, 0, rm.config.MaxAttempts),
	}
	for attemptNum := 1; attemptNum <= rm.config.MaxAttempts; attemptNum++ {
		attemptTime := time.Now()

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			result.FinalError = ctx.Err()
			result.TotalDuration = time.Since(startTime)
			return result
		default:
		}

		// Execute the operation
		operationResult, err := operation(ctx)

		attemptInfo := RetryAttempt{
			AttemptNumber: attemptNum,
			Error:         err,
			Timestamp:     attemptTime,
		}

		if err == nil {
			// Success
			result.Success = true
			result.Result = operationResult
			result.Attempts = append(result.Attempts, attemptInfo)
			result.TotalDuration = time.Since(startTime)
			return result
		}

		// Check if the error is retryable
		if !rm.isRetryableError(err) {
			result.FinalError = err
			result.Attempts = append(result.Attempts, attemptInfo)
			result.TotalDuration = time.Since(startTime)
			return result
		}

		result.Attempts = append(result.Attempts, attemptInfo)

		// If this was the last attempt, don't delay
		if attemptNum == rm.config.MaxAttempts {
			result.FinalError = err
			break
		}

		// Calculate delay for next attempt
		delay := rm.calculateDelay(attemptNum)
		attemptInfo.Delay = delay

		// Wait before retrying
		select {
		case <-ctx.Done():
			result.FinalError = ctx.Err()
			result.TotalDuration = time.Since(startTime)
			return result
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	result.TotalDuration = time.Since(startTime)
	return result
}

// isRetryableError checks if an error is retryable based on configuration
func (rm *RetryManager) isRetryableError(err error) bool {
	// Check if it's a NixAIError with retryable flag
	if nixaiErr, ok := err.(*NixAIError); ok {
		if nixaiErr.IsRetryable() {
			return true
		}
		// Also check if the error code is in our retryable list
		for _, code := range rm.config.RetryableErrors {
			if nixaiErr.Code == code {
				return true
			}
		}
		return false
	}

	// For non-NixAIError types, apply heuristics
	return rm.isErrorRetryableByMessage(err.Error())
}

// isErrorRetryableByMessage checks if an error is retryable based on its message
func (rm *RetryManager) isErrorRetryableByMessage(message string) bool {
	retryablePatterns := []string{
		"timeout",
		"connection refused",
		"connection reset",
		"network unreachable",
		"temporary failure",
		"service unavailable",
		"rate limit",
		"quota exceeded",
		"server error",
		"internal server error",
		"bad gateway",
		"gateway timeout",
	}

	for _, pattern := range retryablePatterns {
		if contains(message, pattern) {
			return true
		}
	}
	return false
}

// calculateDelay calculates the delay for the next retry attempt
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(rm.config.InitialDelay) * math.Pow(rm.config.BackoffFactor, float64(attempt-1))

	// Apply maximum delay constraint
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}

	// Apply jitter if enabled
	if rm.config.JitterEnabled {
		jitter := delay * rm.config.JitterFactor * (rand.Float64()*2 - 1) // Random between -jitterFactor and +jitterFactor
		delay += jitter
	}

	// Ensure delay is not negative
	if delay < 0 {
		delay = float64(rm.config.InitialDelay)
	}

	return time.Duration(delay)
}

// RetryWithExponentialBackoff is a convenience function for simple retry operations
func RetryWithExponentialBackoff(operation RetryableOperation, maxAttempts int) (interface{}, error) {
	config := DefaultRetryConfig()
	config.MaxAttempts = maxAttempts

	manager := NewRetryManager(config)
	result := manager.Execute(operation)

	if result.Success {
		return result.Result, nil
	}
	return nil, result.FinalError
}

// RetryWithConfig is a convenience function for retry operations with custom config
func RetryWithConfig(operation RetryableOperation, config *RetryConfig) (interface{}, error) {
	manager := NewRetryManager(config)
	result := manager.Execute(operation)

	if result.Success {
		return result.Result, nil
	}
	return nil, result.FinalError
}

// RetryWithContext is a convenience function for retry operations with context
func RetryWithContext(ctx context.Context, operation RetryableOperationWithContext, config *RetryConfig) (interface{}, error) {
	manager := NewRetryManager(config)
	result := manager.ExecuteWithContext(ctx, operation)

	if result.Success {
		return result.Result, nil
	}
	return nil, result.FinalError
}

// FormatRetryResult formats a retry result for logging or display
func FormatRetryResult(result *RetryResult) string {
	if result.Success {
		return fmt.Sprintf("Operation succeeded after %d attempts in %v", len(result.Attempts), result.TotalDuration)
	}

	return fmt.Sprintf("Operation failed after %d attempts in %v. Final error: %v",
		len(result.Attempts), result.TotalDuration, result.FinalError)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				string([]rune(s)[0:len([]rune(substr))]) == substr) ||
			indexOf(s, substr) >= 0)
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
