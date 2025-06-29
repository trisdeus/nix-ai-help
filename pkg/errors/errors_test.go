package errors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestErrorTypes(t *testing.T) {
	t.Run("Basic error creation", func(t *testing.T) {
		err := NewError(ErrorCodeNetworkTimeout, "Connection timed out").
			WithSeverity(SeverityMedium).
			WithCategory(CategoryNetwork).
			WithRetryable(true).
			Build()

		if err.Code != ErrorCodeNetworkTimeout {
			t.Errorf("Expected code %s, got %s", ErrorCodeNetworkTimeout, err.Code)
		}

		if err.Severity != SeverityMedium {
			t.Errorf("Expected severity %s, got %s", SeverityMedium, err.Severity)
		}

		if !err.IsRetryable() {
			t.Error("Expected error to be retryable")
		}
	})

	t.Run("Error with context", func(t *testing.T) {
		err := NewError(ErrorCodeAIProviderUnavailable, "AI service down").
			WithContext("provider", "openai").
			WithContext("model", "gpt-4").
			WithSuggestion("Try using Ollama instead").
			Build()

		if len(err.Context) != 2 {
			t.Errorf("Expected 2 context items, got %d", len(err.Context))
		}

		if err.Context["provider"] != "openai" {
			t.Error("Context not set correctly")
		}

		if len(err.Suggestions) != 1 {
			t.Errorf("Expected 1 suggestion, got %d", len(err.Suggestions))
		}
	})

	t.Run("Default values", func(t *testing.T) {
		err := NewError(ErrorCodeNixBuildFailed, "Build failed").Build()

		// Should get default severity and category
		if err.Severity != SeverityHigh {
			t.Errorf("Expected default severity %s, got %s", SeverityHigh, err.Severity)
		}

		if err.Category != CategoryNixOS {
			t.Errorf("Expected default category %s, got %s", CategoryNixOS, err.Category)
		}
	})
}

func TestRetryLogic(t *testing.T) {
	t.Run("Successful operation", func(t *testing.T) {
		manager := NewRetryManager(DefaultRetryConfig())

		operation := func() (interface{}, error) {
			return "success", nil
		}

		result := manager.Execute(operation)

		if !result.Success {
			t.Error("Expected successful result")
		}

		if len(result.Attempts) != 1 {
			t.Errorf("Expected 1 attempt, got %d", len(result.Attempts))
		}

		if result.Result != "success" {
			t.Errorf("Expected 'success', got %v", result.Result)
		}
	})

	t.Run("Retryable error", func(t *testing.T) {
		config := DefaultRetryConfig()
		config.MaxAttempts = 3
		config.InitialDelay = 10 * time.Millisecond
		manager := NewRetryManager(config)

		attempts := 0
		operation := func() (interface{}, error) {
			attempts++
			if attempts < 3 {
				return nil, NewError(ErrorCodeNetworkTimeout, "Timeout").
					WithRetryable(true).Build()
			}
			return "success", nil
		}

		result := manager.Execute(operation)

		if !result.Success {
			t.Error("Expected successful result after retries")
		}

		if len(result.Attempts) != 3 {
			t.Errorf("Expected 3 attempts, got %d", len(result.Attempts))
		}
	})

	t.Run("Non-retryable error", func(t *testing.T) {
		manager := NewRetryManager(DefaultRetryConfig())

		operation := func() (interface{}, error) {
			return nil, NewError(ErrorCodeInvalidInput, "Invalid input").
				WithRetryable(false).Build()
		}

		result := manager.Execute(operation)

		if result.Success {
			t.Error("Expected failed result")
		}

		if len(result.Attempts) != 1 {
			t.Errorf("Expected 1 attempt, got %d", len(result.Attempts))
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		manager := NewRetryManager(DefaultRetryConfig())

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		operation := func(ctx context.Context) (interface{}, error) {
			return nil, NewError(ErrorCodeNetworkTimeout, "Timeout").
				WithRetryable(true).Build()
		}

		result := manager.ExecuteWithContext(ctx, operation)

		if result.Success {
			t.Error("Expected failed result due to cancellation")
		}

		if result.FinalError != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", result.FinalError)
		}
	})
}

func TestPanicRecovery(t *testing.T) {
	t.Run("Panic recovery", func(t *testing.T) {
		handler := NewPanicRecoveryHandler(nil)

		var recoveredErr *NixAIError

		func() {
			defer func() {
				recoveredErr = handler.Recover("test_panic")
			}()

			panic("test panic")
		}()

		if recoveredErr == nil {
			t.Error("Expected recovered error")
		}

		if recoveredErr.Code != ErrorCodePanicRecovered {
			t.Errorf("Expected code %s, got %s", ErrorCodePanicRecovered, recoveredErr.Code)
		}

		if recoveredErr.Severity != SeverityCritical {
			t.Error("Expected critical severity for panic")
		}
	})
	t.Run("Safe execute", func(t *testing.T) {
		// Test normal execution without panic
		err := SafeExecute(func() error {
			return fmt.Errorf("normal error")
		}, "safe_test")

		if err == nil {
			t.Error("Expected error to be returned")
		}
	})

	t.Run("Safe execute with result", func(t *testing.T) {
		// Test normal execution without panic
		result, err := SafeExecuteWithResult(func() (string, error) {
			return "success", nil
		}, "safe_test_result")

		if result != "success" {
			t.Errorf("Expected 'success', got %s", result)
		}

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestUserFriendlyMessages(t *testing.T) {
	generator := NewUserFriendlyMessageGenerator()

	t.Run("Network timeout message", func(t *testing.T) {
		err := NewError(ErrorCodeNetworkTimeout, "Connection timed out").Build()
		message := generator.GenerateForNixAIError(err)

		if !contains(message, "Network connection timed out") {
			t.Error("Expected user-friendly network timeout message")
		}

		if !contains(message, "Suggested actions") {
			t.Error("Expected suggestions in message")
		}
	})

	t.Run("AI provider unavailable message", func(t *testing.T) {
		err := NewError(ErrorCodeAIProviderUnavailable, "AI service down").Build()
		message := generator.GenerateForNixAIError(err)

		if !contains(message, "AI service is currently unavailable") {
			t.Error("Expected user-friendly AI provider message")
		}

		if !contains(message, "Ollama") {
			t.Error("Expected Ollama suggestion in AI provider message")
		}
	})

	t.Run("Custom user message", func(t *testing.T) {
		err := NewError(ErrorCodeFileNotFound, "File not found").
			WithUserMessage("Custom message for user").
			WithSuggestion("Custom suggestion").
			Build()

		message := generator.GenerateForNixAIError(err)

		if !contains(message, "Custom message for user") {
			t.Error("Expected custom user message")
		}

		if !contains(message, "Custom suggestion") {
			t.Error("Expected custom suggestion")
		}
	})

	t.Run("Generic error message", func(t *testing.T) {
		genericErr := fmt.Errorf("connection refused")
		message := generator.GenerateForGenericError(genericErr)

		if !contains(message, "Connection was refused") {
			t.Error("Expected connection refused message")
		}
	})
}

func TestErrorAnalytics(t *testing.T) {
	tempDir := t.TempDir()
	analytics := NewErrorAnalytics(tempDir)

	t.Run("Record and analyze errors", func(t *testing.T) {
		// Record some errors
		err1 := NewError(ErrorCodeNetworkTimeout, "Timeout 1").Build()
		err2 := NewError(ErrorCodeNetworkTimeout, "Timeout 2").Build()
		err3 := NewError(ErrorCodeAIProviderUnavailable, "AI down").Build()

		analytics.RecordError(err1, "test_operation_1")
		analytics.RecordError(err2, "test_operation_2")
		analytics.RecordError(err3, "test_operation_3")

		report := analytics.GenerateReport()

		if report.TotalErrors != 3 {
			t.Errorf("Expected 3 total errors, got %d", report.TotalErrors)
		}

		if report.UniqueErrors != 2 {
			t.Errorf("Expected 2 unique errors, got %d", report.UniqueErrors)
		}

		if report.ErrorsByCategory[CategoryNetwork] != 2 {
			t.Errorf("Expected 2 network errors, got %d", report.ErrorsByCategory[CategoryNetwork])
		}

		if report.ErrorsByCategory[CategoryAI] != 1 {
			t.Errorf("Expected 1 AI error, got %d", report.ErrorsByCategory[CategoryAI])
		}
	})

	t.Run("Top errors", func(t *testing.T) {
		topErrors := analytics.GetTopErrors(5)

		if len(topErrors) == 0 {
			t.Error("Expected top errors")
		}

		// First error should be the most frequent (network timeout)
		if topErrors[0].Code != ErrorCodeNetworkTimeout {
			t.Errorf("Expected top error to be %s, got %s", ErrorCodeNetworkTimeout, topErrors[0].Code)
		}

		if topErrors[0].Count != 2 {
			t.Errorf("Expected count of 2, got %d", topErrors[0].Count)
		}
	})

	t.Run("Record resolution", func(t *testing.T) {
		analytics.RecordResolution(ErrorCodeNetworkTimeout, "retry_successful")

		// Test that resolution is recorded (implementation detail, hard to test directly)
		// This would be verified in a real system through persistence
	})

	t.Run("Export report", func(t *testing.T) {
		reportFile := filepath.Join(tempDir, "test_report.json")
		err := analytics.ExportReport(reportFile)

		if err != nil {
			t.Errorf("Failed to export report: %v", err)
		}

		if _, err := os.Stat(reportFile); os.IsNotExist(err) {
			t.Error("Report file was not created")
		}
	})
}

func TestErrorManager(t *testing.T) {
	tempDir := t.TempDir()
	config := &ErrorManagerConfig{
		DebugMode:           true,
		GracefulDegradation: true,
		AnalyticsEnabled:    true,
		AnalyticsDataDir:    tempDir,
		RetryConfig:         DefaultRetryConfig(),
		MaxLastErrors:       10,
	}

	manager := NewErrorManager(config)

	t.Run("Handle error", func(t *testing.T) {
		originalErr := fmt.Errorf("connection timeout")
		nixaiErr := manager.HandleError(originalErr, "test_context")

		if nixaiErr == nil {
			t.Error("Expected handled error")
		}

		if nixaiErr.Code != ErrorCodeNetworkTimeout {
			t.Errorf("Expected %s, got %s", ErrorCodeNetworkTimeout, nixaiErr.Code)
		}

		// Check that error was recorded
		lastErrors := manager.GetLastErrors(5)
		if len(lastErrors) == 0 {
			t.Error("Expected error to be recorded")
		}

		if lastErrors[0].Context != "test_context" {
			t.Errorf("Expected context 'test_context', got '%s'", lastErrors[0].Context)
		}
	})

	t.Run("Handle with retry", func(t *testing.T) {
		attempts := 0
		operation := func() (interface{}, error) {
			attempts++
			if attempts < 2 {
				return nil, fmt.Errorf("temporary failure")
			}
			return "success", nil
		}

		result, err := manager.HandleErrorWithRetry(operation, "retry_test")

		if err != nil {
			t.Errorf("Expected success after retry, got error: %v", err)
		}

		if result != "success" {
			t.Errorf("Expected 'success', got %v", result)
		}

		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("User friendly message", func(t *testing.T) {
		err := NewError(ErrorCodeAIProviderUnavailable, "AI down").Build()
		message := manager.GetUserFriendlyMessage(err)

		if !contains(message, "AI service is currently unavailable") {
			t.Error("Expected user-friendly message")
		}
	})

	t.Run("Formatted error with debug", func(t *testing.T) {
		err := NewError(ErrorCodeInternalError, "Internal error").
			WithContext("test_key", "test_value").
			Build()

		formatted := manager.GetFormattedError(err)

		if !contains(formatted, "[DEBUG]") {
			t.Error("Expected debug information in formatted error")
		}

		if !contains(formatted, "test_key: test_value") {
			t.Error("Expected context in debug information")
		}
	})

	t.Run("Analytics report", func(t *testing.T) {
		report := manager.GetAnalyticsReport()

		if report == nil {
			t.Error("Expected analytics report")
		}

		if report.TotalErrors == 0 {
			t.Error("Expected some recorded errors")
		}
	})

	t.Run("Record resolution", func(t *testing.T) {
		manager.RecordResolution(ErrorCodeNetworkTimeout, "network_restored")

		// Check that resolution was recorded in last errors
		lastErrors := manager.GetLastErrors(10)
		found := false
		for _, lastErr := range lastErrors {
			if lastErr.Error.Code == ErrorCodeNetworkTimeout && lastErr.Resolved {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find resolved error")
		}
	})
}

func TestGlobalErrorManager(t *testing.T) {
	t.Run("Global instance", func(t *testing.T) {
		manager1 := GetGlobalErrorManager()
		manager2 := GetGlobalErrorManager()

		if manager1 != manager2 {
			t.Error("Expected same global instance")
		}
	})

	t.Run("Global convenience functions", func(t *testing.T) {
		err := fmt.Errorf("test error")
		nixaiErr := Handle(err, "global_test")

		if nixaiErr == nil {
			t.Error("Expected handled error")
		}

		formattedErr := FormatForUser(nixaiErr)
		if formattedErr == "" {
			t.Error("Expected formatted error message")
		}
	})
}

func TestConfigurationValidation(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := DefaultErrorManagerConfig()
		err := ValidateConfiguration(config)

		if err != nil {
			t.Errorf("Expected valid configuration, got error: %v", err)
		}
	})

	t.Run("Nil configuration", func(t *testing.T) {
		err := ValidateConfiguration(nil)

		if err == nil {
			t.Error("Expected error for nil configuration")
		}

		if nixaiErr, ok := err.(*NixAIError); ok {
			if nixaiErr.Code != ErrorCodeInvalidInput {
				t.Error("Expected invalid input error code")
			}
		}
	})

	t.Run("Invalid analytics config", func(t *testing.T) {
		config := DefaultErrorManagerConfig()
		config.AnalyticsEnabled = true
		config.AnalyticsDataDir = ""

		err := ValidateConfiguration(config)

		if err == nil {
			t.Error("Expected error for invalid analytics configuration")
		}
	})

	t.Run("Invalid max last errors", func(t *testing.T) {
		config := DefaultErrorManagerConfig()
		config.MaxLastErrors = 0

		err := ValidateConfiguration(config)

		if err == nil {
			t.Error("Expected error for invalid max last errors")
		}
	})
}
