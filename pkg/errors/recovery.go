package errors

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// PanicRecoveryHandler handles panic recovery and reporting
type PanicRecoveryHandler struct {
	handlers []PanicHandler
	logger   PanicLogger
}

// PanicHandler is a function that handles recovered panics
type PanicHandler func(panicInfo *PanicInfo)

// PanicLogger defines the interface for logging panics
type PanicLogger interface {
	LogPanic(panicInfo *PanicInfo)
}

// PanicInfo contains information about a recovered panic
type PanicInfo struct {
	Value          interface{}            `json:"value"`
	StackTrace     string                 `json:"stack_trace"`
	Timestamp      time.Time              `json:"timestamp"`
	Goroutine      string                 `json:"goroutine"`
	FunctionName   string                 `json:"function_name"`
	FileName       string                 `json:"file_name"`
	LineNumber     int                    `json:"line_number"`
	RecoveryPoint  string                 `json:"recovery_point"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

// DefaultPanicLogger provides basic panic logging
type DefaultPanicLogger struct{}

// LogPanic logs panic information
func (dpl *DefaultPanicLogger) LogPanic(panicInfo *PanicInfo) {
	fmt.Printf("PANIC RECOVERED: %v at %s:%d\nStack Trace:\n%s\n",
		panicInfo.Value, panicInfo.FileName, panicInfo.LineNumber, panicInfo.StackTrace)
}

// NewPanicRecoveryHandler creates a new panic recovery handler
func NewPanicRecoveryHandler(logger PanicLogger) *PanicRecoveryHandler {
	if logger == nil {
		logger = &DefaultPanicLogger{}
	}
	return &PanicRecoveryHandler{
		handlers: make([]PanicHandler, 0),
		logger:   logger,
	}
}

// AddHandler adds a panic handler
func (prh *PanicRecoveryHandler) AddHandler(handler PanicHandler) {
	prh.handlers = append(prh.handlers, handler)
}

// Recover recovers from a panic and processes it through registered handlers
func (prh *PanicRecoveryHandler) Recover(recoveryPoint string) *NixAIError {
	if r := recover(); r != nil {
		panicInfo := prh.buildPanicInfo(r, recoveryPoint)

		// Log the panic
		prh.logger.LogPanic(panicInfo)

		// Process through handlers
		for _, handler := range prh.handlers {
			handler(panicInfo)
		}

		// Return a structured error
		return NewError(ErrorCodePanicRecovered, fmt.Sprintf("Panic recovered: %v", r)).
			WithSeverity(SeverityCritical).
			WithCategory(CategoryInternal).
			WithDetails(fmt.Sprintf("Panic occurred at %s", recoveryPoint)).
			WithStackTrace(panicInfo.StackTrace).
			WithContext("panic_value", r).
			WithContext("recovery_point", recoveryPoint).
			WithContext("function", panicInfo.FunctionName).
			WithContext("file", panicInfo.FileName).
			WithContext("line", panicInfo.LineNumber).
			WithUserMessage("An unexpected error occurred. The operation has been safely recovered.").
			WithSuggestion("Please report this issue with the provided error details").
			WithSuggestion("Try running the operation again").
			Build()
	}
	return nil
}

// buildPanicInfo constructs PanicInfo from a recovered panic
func (prh *PanicRecoveryHandler) buildPanicInfo(panicValue interface{}, recoveryPoint string) *PanicInfo {
	stackTrace := string(debug.Stack())

	// Parse stack trace to get function and location info
	functionName, fileName, lineNumber := prh.parseStackTrace(stackTrace)

	return &PanicInfo{
		Value:          panicValue,
		StackTrace:     stackTrace,
		Timestamp:      time.Now(),
		Goroutine:      prh.getGoroutineInfo(),
		FunctionName:   functionName,
		FileName:       fileName,
		LineNumber:     lineNumber,
		RecoveryPoint:  recoveryPoint,
		AdditionalInfo: make(map[string]interface{}),
	}
}

// parseStackTrace extracts function, file, and line information from stack trace
func (prh *PanicRecoveryHandler) parseStackTrace(stackTrace string) (string, string, int) {
	lines := strings.Split(stackTrace, "\n")

	// Skip the first few lines (panic info and recover call)
	for i := 4; i < len(lines)-1; i += 2 {
		if i+1 < len(lines) {
			functionLine := strings.TrimSpace(lines[i])
			locationLine := strings.TrimSpace(lines[i+1])

			// Parse location line to extract file and line number
			if strings.Contains(locationLine, ":") {
				parts := strings.Split(locationLine, ":")
				if len(parts) >= 2 {
					fileName := parts[0]
					// Extract line number (ignore column info if present)
					lineStr := strings.Fields(parts[1])[0]
					lineNumber := 0
					fmt.Sscanf(lineStr, "%d", &lineNumber)

					return functionLine, fileName, lineNumber
				}
			}
		}
	}

	return "unknown", "unknown", 0
}

// getGoroutineInfo returns information about the current goroutine
func (prh *PanicRecoveryHandler) getGoroutineInfo() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	lines := strings.Split(stack, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return "unknown goroutine"
}

// SafeExecute executes a function with panic recovery
func SafeExecute(fn func() error, recoveryPoint string) (finalErr error) {
	handler := NewPanicRecoveryHandler(nil)

	defer func() {
		if nixaiErr := handler.Recover(recoveryPoint); nixaiErr != nil {
			finalErr = nixaiErr
		}
	}()

	finalErr = fn()
	return finalErr
}

// SafeExecuteWithResult executes a function with panic recovery and returns result
func SafeExecuteWithResult[T any](fn func() (T, error), recoveryPoint string) (result T, finalErr error) {
	handler := NewPanicRecoveryHandler(nil)

	defer func() {
		if nixaiErr := handler.Recover(recoveryPoint); nixaiErr != nil {
			var zero T
			result = zero
			finalErr = nixaiErr
		}
	}()

	result, finalErr = fn()
	return result, finalErr
}

// SafeExecuteWithCallback executes a function with panic recovery and callback
func SafeExecuteWithCallback(fn func() error, recoveryPoint string, onPanic PanicHandler) error {
	handler := NewPanicRecoveryHandler(nil)
	if onPanic != nil {
		handler.AddHandler(onPanic)
	}

	var panicErr error

	defer func() {
		if nixaiErr := handler.Recover(recoveryPoint); nixaiErr != nil {
			panicErr = nixaiErr
		}
	}()

	err := fn()
	if panicErr != nil {
		return panicErr
	}
	return err
}

// WrapWithRecovery wraps a function to automatically handle panics
func WrapWithRecovery(fn func() error, recoveryPoint string) func() error {
	return func() error {
		return SafeExecute(fn, recoveryPoint)
	}
}

// WrapWithRecoveryAndResult wraps a function with result to automatically handle panics
func WrapWithRecoveryAndResult[T any](fn func() (T, error), recoveryPoint string) func() (T, error) {
	return func() (T, error) {
		return SafeExecuteWithResult(fn, recoveryPoint)
	}
}

// RecoveryMiddleware creates middleware that recovers from panics in handler functions
func RecoveryMiddleware(recoveryPoint string, logger PanicLogger) func(func() error) func() error {
	return func(handler func() error) func() error {
		return func() error {
			return SafeExecute(handler, recoveryPoint)
		}
	}
}

// GetSystemInfo returns system information useful for panic reports
func GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"go_version":    runtime.Version(),
		"goos":          runtime.GOOS,
		"goarch":        runtime.GOARCH,
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"memory_stats":  getMemoryStats(),
	}
}

// getMemoryStats returns current memory statistics
func getMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"alloc":         m.Alloc,
		"total_alloc":   m.TotalAlloc,
		"sys":           m.Sys,
		"mallocs":       m.Mallocs,
		"frees":         m.Frees,
		"heap_alloc":    m.HeapAlloc,
		"heap_sys":      m.HeapSys,
		"heap_idle":     m.HeapIdle,
		"heap_inuse":    m.HeapInuse,
		"heap_released": m.HeapReleased,
		"heap_objects":  m.HeapObjects,
		"gc_cycles":     m.NumGC,
	}
}
