package types

import (
	"time"
)

// CommandRequest represents a command execution request
type CommandRequest struct {
	Command      string            `json:"command"`
	Args         []string          `json:"args"`
	RequiresSudo bool              `json:"requiresSudo"`
	WorkingDir   string            `json:"workingDir,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Description  string            `json:"description"`
	Category     string            `json:"category"`
	DryRun       bool              `json:"dryRun,omitempty"`
	Timeout      time.Duration     `json:"timeout,omitempty"`
}

// ExecutionResult represents the result of command execution
type ExecutionResult struct {
	Success   bool          `json:"success"`
	ExitCode  int           `json:"exitCode"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Command   string        `json:"command"`
	Timestamp time.Time     `json:"timestamp"`
	DryRun    bool          `json:"dryRun"`
}

// CompoundOperation represents a series of commands to execute together
type CompoundOperation struct {
	ID           string           `json:"id"`
	Commands     []CommandRequest `json:"commands"`
	RollbackCmds []CommandRequest `json:"rollbackCommands,omitempty"`
	AllowPartial bool             `json:"allowPartial"`
	Description  string           `json:"description"`
}

// CompoundResult represents the result of a compound operation
type CompoundResult struct {
	OperationID      string             `json:"operationId"`
	Success          bool               `json:"success"`
	StartTime        time.Time          `json:"startTime"`
	EndTime          time.Time          `json:"endTime"`
	FailedAt         int                `json:"failedAt,omitempty"`
	CommandResults   []*ExecutionResult `json:"commandResults"`
	RollbackExecuted bool               `json:"rollbackExecuted"`
}

// CommandCategory represents different types of commands
type CommandCategory string

const (
	CategoryPackage       CommandCategory = "package"
	CategorySystem        CommandCategory = "system"
	CategoryConfiguration CommandCategory = "configuration"
	CategoryDevelopment   CommandCategory = "development"
	CategoryUtility       CommandCategory = "utility"
)

// SecurityError represents security-related validation errors
type SecurityError struct {
	Command string `json:"command"`
	Reason  string `json:"reason"`
	Code    string `json:"code"`
}

func (e *SecurityError) Error() string {
	return e.Reason
}

// ValidationError represents general validation errors
type ValidationError struct {
	Command string `json:"command"`
	Reason  string `json:"reason"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	return e.Reason
}