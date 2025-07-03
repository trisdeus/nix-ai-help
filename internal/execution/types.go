package execution

import (
	"context"
	"time"
)

// CommandRequest represents a request to execute a system command
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

// ExecutionResult contains the result of command execution
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

// CommandCategory represents different types of commands
type CommandCategory string

const (
	CategoryPackage       CommandCategory = "package"
	CategorySystem        CommandCategory = "system"
	CategoryConfiguration CommandCategory = "configuration"
	CategoryDevelopment   CommandCategory = "development"
	CategoryUtility       CommandCategory = "utility"
)

// ExecutionContext provides context for command execution
type ExecutionContext struct {
	UserID      int
	WorkingDir  string
	Environment map[string]string
	Timeout     time.Duration
	Interactive bool
}

// CommandValidator interface for validating commands
type CommandValidator interface {
	ValidateCommand(req CommandRequest) error
	IsCommandAllowed(command string, args []string) bool
	RequiresConfirmation(req CommandRequest) bool
}

// Executor interface for command execution
type Executor interface {
	ExecuteCommand(ctx context.Context, req CommandRequest) (*ExecutionResult, error)
	ExecuteWithSudo(ctx context.Context, req CommandRequest) (*ExecutionResult, error)
	DryRun(ctx context.Context, req CommandRequest) (*ExecutionResult, error)
}

// ProgressCallback function type for tracking execution progress
type ProgressCallback func(stage string, progress float64, message string)

// CompoundOperation represents multiple commands that should be executed together
type CompoundOperation struct {
	ID           string
	Name         string
	Description  string
	Commands     []CommandRequest
	RollbackCmds []CommandRequest
	AllowPartial bool // Allow partial success if some commands fail
}

// CompoundResult contains results from compound operation execution
type CompoundResult struct {
	OperationID      string
	Success          bool
	CommandResults   []*ExecutionResult
	FailedAt         int // Index of command that failed (-1 if no failure)
	RollbackExecuted bool
	StartTime        time.Time
	EndTime          time.Time
}

// ValidationError represents a command validation error
type ValidationError struct {
	Command string
	Reason  string
	Code    string
}

func (e ValidationError) Error() string {
	return e.Reason
}

// SecurityError represents a security-related execution error
type SecurityError struct {
	Command string
	Reason  string
	Code    string
}

func (e SecurityError) Error() string {
	return e.Reason
}

// ExecutionError represents a general execution error
type ExecutionError struct {
	Command  string
	ExitCode int
	Output   string
	Reason   string
}

func (e ExecutionError) Error() string {
	return e.Reason
}