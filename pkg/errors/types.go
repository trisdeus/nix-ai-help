package errors

import (
	"fmt"
	"time"
)

// ErrorCode represents different categories of errors
type ErrorCode string

const (
	// Network and connectivity errors
	ErrorCodeNetworkTimeout     ErrorCode = "NETWORK_TIMEOUT"
	ErrorCodeConnectionRefused  ErrorCode = "CONNECTION_REFUSED"
	ErrorCodeNetworkUnreachable ErrorCode = "NETWORK_UNREACHABLE"

	// AI provider errors
	ErrorCodeAIProviderUnavailable ErrorCode = "AI_PROVIDER_UNAVAILABLE"
	ErrorCodeAIRateLimited         ErrorCode = "AI_RATE_LIMITED"
	ErrorCodeAIInvalidResponse     ErrorCode = "AI_INVALID_RESPONSE"
	ErrorCodeAIQuotaExceeded       ErrorCode = "AI_QUOTA_EXCEEDED"

	// File system and configuration errors
	ErrorCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrorCodeConfigInvalid    ErrorCode = "CONFIG_INVALID"
	ErrorCodeConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"

	// NixOS and build system errors
	ErrorCodeNixBuildFailed        ErrorCode = "NIX_BUILD_FAILED"
	ErrorCodeNixConfigInvalid      ErrorCode = "NIX_CONFIG_INVALID"
	ErrorCodeNixStoreCorrupted     ErrorCode = "NIX_STORE_CORRUPTED"
	ErrorCodeNixChannelUnavailable ErrorCode = "NIX_CHANNEL_UNAVAILABLE"

	// MCP server errors
	ErrorCodeMCPServerUnavailable ErrorCode = "MCP_SERVER_UNAVAILABLE"
	ErrorCodeMCPProtocolError     ErrorCode = "MCP_PROTOCOL_ERROR"
	ErrorCodeMCPSocketError       ErrorCode = "MCP_SOCKET_ERROR"

	// Cache and storage errors
	ErrorCodeCacheCorrupted   ErrorCode = "CACHE_CORRUPTED"
	ErrorCodeCacheUnavailable ErrorCode = "CACHE_UNAVAILABLE"
	ErrorCodeStorageFull      ErrorCode = "STORAGE_FULL"

	// Input validation errors
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingParameter ErrorCode = "MISSING_PARAMETER"
	ErrorCodeInvalidParameter ErrorCode = "INVALID_PARAMETER"

	// Internal errors
	ErrorCodeInternalError     ErrorCode = "INTERNAL_ERROR"
	ErrorCodePanicRecovered    ErrorCode = "PANIC_RECOVERED"
	ErrorCodeResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"

	// Unknown errors
	ErrorCodeUnknown ErrorCode = "UNKNOWN"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "LOW"
	SeverityMedium   ErrorSeverity = "MEDIUM"
	SeverityHigh     ErrorSeverity = "HIGH"
	SeverityCritical ErrorSeverity = "CRITICAL"
)

// ErrorCategory represents broad categories of errors for analytics
type ErrorCategory string

const (
	CategoryNetwork    ErrorCategory = "NETWORK"
	CategoryAI         ErrorCategory = "AI"
	CategoryFileSystem ErrorCategory = "FILESYSTEM"
	CategoryNixOS      ErrorCategory = "NIXOS"
	CategoryMCP        ErrorCategory = "MCP"
	CategoryCache      ErrorCategory = "CACHE"
	CategoryValidation ErrorCategory = "VALIDATION"
	CategoryInternal   ErrorCategory = "INTERNAL"
)

// NixAIError represents a structured error with additional context
type NixAIError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Cause       error                  `json:"-"`
	Severity    ErrorSeverity          `json:"severity"`
	Category    ErrorCategory          `json:"category"`
	Retryable   bool                   `json:"retryable"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	UserMessage string                 `json:"user_message,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// Error implements the error interface
func (e *NixAIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error cause
func (e *NixAIError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether this error can be retried
func (e *NixAIError) IsRetryable() bool {
	return e.Retryable
}

// GetSeverity returns the error severity
func (e *NixAIError) GetSeverity() ErrorSeverity {
	return e.Severity
}

// GetCategory returns the error category
func (e *NixAIError) GetCategory() ErrorCategory {
	return e.Category
}

// GetUserMessage returns a user-friendly error message
func (e *NixAIError) GetUserMessage() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	return e.Message
}

// GetSuggestions returns suggested actions to resolve the error
func (e *NixAIError) GetSuggestions() []string {
	return e.Suggestions
}

// AddContext adds additional context to the error
func (e *NixAIError) AddContext(key string, value interface{}) *NixAIError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause adds a cause to the error
func (e *NixAIError) WithCause(cause error) *NixAIError {
	e.Cause = cause
	return e
}

// WithSuggestion adds a suggestion to resolve the error
func (e *NixAIError) WithSuggestion(suggestion string) *NixAIError {
	e.Suggestions = append(e.Suggestions, suggestion)
	return e
}

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
	err *NixAIError
}

// NewError creates a new ErrorBuilder
func NewError(code ErrorCode, message string) *ErrorBuilder {
	return &ErrorBuilder{
		err: &NixAIError{
			Code:      code,
			Message:   message,
			Timestamp: time.Now(),
			Context:   make(map[string]interface{}),
		},
	}
}

// WithDetails adds details to the error
func (b *ErrorBuilder) WithDetails(details string) *ErrorBuilder {
	b.err.Details = details
	return b
}

// WithCause adds a cause to the error
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

// WithSeverity sets the error severity
func (b *ErrorBuilder) WithSeverity(severity ErrorSeverity) *ErrorBuilder {
	b.err.Severity = severity
	return b
}

// WithCategory sets the error category
func (b *ErrorBuilder) WithCategory(category ErrorCategory) *ErrorBuilder {
	b.err.Category = category
	return b
}

// WithRetryable sets whether the error is retryable
func (b *ErrorBuilder) WithRetryable(retryable bool) *ErrorBuilder {
	b.err.Retryable = retryable
	return b
}

// WithContext adds context to the error
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.err.Context[key] = value
	return b
}

// WithUserMessage sets a user-friendly message
func (b *ErrorBuilder) WithUserMessage(message string) *ErrorBuilder {
	b.err.UserMessage = message
	return b
}

// WithSuggestion adds a suggestion to resolve the error
func (b *ErrorBuilder) WithSuggestion(suggestion string) *ErrorBuilder {
	b.err.Suggestions = append(b.err.Suggestions, suggestion)
	return b
}

// WithStackTrace adds a stack trace to the error
func (b *ErrorBuilder) WithStackTrace(trace string) *ErrorBuilder {
	b.err.StackTrace = trace
	return b
}

// Build returns the constructed error
func (b *ErrorBuilder) Build() *NixAIError {
	// Set defaults based on error code if not explicitly set
	if b.err.Severity == "" {
		b.err.Severity = getDefaultSeverity(b.err.Code)
	}
	if b.err.Category == "" {
		b.err.Category = getDefaultCategory(b.err.Code)
	}
	if !b.err.Retryable && isDefaultRetryable(b.err.Code) {
		b.err.Retryable = true
	}

	return b.err
}

// getDefaultSeverity returns the default severity for an error code
func getDefaultSeverity(code ErrorCode) ErrorSeverity {
	switch code {
	case ErrorCodePanicRecovered, ErrorCodeInternalError, ErrorCodeNixStoreCorrupted:
		return SeverityCritical
	case ErrorCodeNixBuildFailed, ErrorCodeAIProviderUnavailable, ErrorCodeMCPServerUnavailable:
		return SeverityHigh
	case ErrorCodeNetworkTimeout, ErrorCodeAIRateLimited, ErrorCodeConfigInvalid:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// getDefaultCategory returns the default category for an error code
func getDefaultCategory(code ErrorCode) ErrorCategory {
	switch code {
	case ErrorCodeNetworkTimeout, ErrorCodeConnectionRefused, ErrorCodeNetworkUnreachable:
		return CategoryNetwork
	case ErrorCodeAIProviderUnavailable, ErrorCodeAIRateLimited, ErrorCodeAIInvalidResponse, ErrorCodeAIQuotaExceeded:
		return CategoryAI
	case ErrorCodeFileNotFound, ErrorCodePermissionDenied, ErrorCodeConfigInvalid, ErrorCodeConfigNotFound:
		return CategoryFileSystem
	case ErrorCodeNixBuildFailed, ErrorCodeNixConfigInvalid, ErrorCodeNixStoreCorrupted, ErrorCodeNixChannelUnavailable:
		return CategoryNixOS
	case ErrorCodeMCPServerUnavailable, ErrorCodeMCPProtocolError, ErrorCodeMCPSocketError:
		return CategoryMCP
	case ErrorCodeCacheCorrupted, ErrorCodeCacheUnavailable, ErrorCodeStorageFull:
		return CategoryCache
	case ErrorCodeInvalidInput, ErrorCodeMissingParameter, ErrorCodeInvalidParameter:
		return CategoryValidation
	default:
		return CategoryInternal
	}
}

// isDefaultRetryable returns whether an error code is retryable by default
func isDefaultRetryable(code ErrorCode) bool {
	switch code {
	case ErrorCodeNetworkTimeout, ErrorCodeConnectionRefused, ErrorCodeAIProviderUnavailable,
		ErrorCodeAIRateLimited, ErrorCodeNixChannelUnavailable, ErrorCodeMCPServerUnavailable,
		ErrorCodeCacheUnavailable:
		return true
	default:
		return false
	}
}
