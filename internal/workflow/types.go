// Package workflow provides automated workflow execution capabilities for nixai
// This includes workflow definitions, task scheduling, action execution, and state management
package workflow

import (
	"context"
	"time"
)

// WorkflowStatus represents the current state of a workflow
type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCancelled WorkflowStatus = "cancelled"
	StatusPaused    WorkflowStatus = "paused"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusSkipped   TaskStatus = "skipped"
	TaskStatusRetrying  TaskStatus = "retrying"
)

// ActionType defines the type of action to execute
type ActionType string

const (
	ActionTypeCommand      ActionType = "command"
	ActionTypeNixOSRebuild ActionType = "nixos-rebuild"
	ActionTypeFileEdit     ActionType = "file-edit"
	ActionTypePackageOp    ActionType = "package-op"
	ActionTypeServiceOp    ActionType = "service-op"
	ActionTypeValidation   ActionType = "validation"
	ActionTypeQuery        ActionType = "query"
	ActionTypeConditional  ActionType = "conditional"
)

// TriggerType defines types of workflow triggers
type TriggerType string

const (
	TriggerTypeManual     TriggerType = "manual"
	TriggerTypeScheduled  TriggerType = "scheduled"
	TriggerTypeEvent      TriggerType = "event"
	TriggerTypeFileChange TriggerType = "file-change"
	TriggerTypeError      TriggerType = "error"
	TriggerTypeCondition  TriggerType = "condition"
)

// Workflow represents a complete automated workflow
type Workflow struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Version     string            `yaml:"version" json:"version"`
	Author      string            `yaml:"author" json:"author"`
	Tags        []string          `yaml:"tags" json:"tags"`
	Metadata    map[string]string `yaml:"metadata" json:"metadata"`

	// Execution configuration
	Triggers    []Trigger         `yaml:"triggers" json:"triggers"`
	Tasks       []Task            `yaml:"tasks" json:"tasks"`
	Variables   map[string]string `yaml:"variables" json:"variables"`
	Environment map[string]string `yaml:"environment" json:"environment"`

	// Runtime state
	Status      WorkflowStatus `yaml:"-" json:"status"`
	StartTime   *time.Time     `yaml:"-" json:"start_time,omitempty"`
	EndTime     *time.Time     `yaml:"-" json:"end_time,omitempty"`
	Error       string         `yaml:"-" json:"error,omitempty"`
	CurrentTask int            `yaml:"-" json:"current_task"`

	// Execution options
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
	RetryDelay      time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	ContinueOnError bool          `yaml:"continue_on_error" json:"continue_on_error"`
	ParallelTasks   bool          `yaml:"parallel_tasks" json:"parallel_tasks"`
}

// Task represents a single task within a workflow
type Task struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	DependsOn   []string          `yaml:"depends_on" json:"depends_on"`
	Condition   string            `yaml:"condition" json:"condition"`
	Actions     []Action          `yaml:"actions" json:"actions"`
	Variables   map[string]string `yaml:"variables" json:"variables"`

	// Runtime state
	Status     TaskStatus `yaml:"-" json:"status"`
	StartTime  *time.Time `yaml:"-" json:"start_time,omitempty"`
	EndTime    *time.Time `yaml:"-" json:"end_time,omitempty"`
	Error      string     `yaml:"-" json:"error,omitempty"`
	Output     string     `yaml:"-" json:"output,omitempty"`
	RetryCount int        `yaml:"-" json:"retry_count"`

	// Task options
	MaxRetries int           `yaml:"max_retries" json:"max_retries"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
	Critical   bool          `yaml:"critical" json:"critical"`
	Optional   bool          `yaml:"optional" json:"optional"`
}

// Action represents a single action within a task
type Action struct {
	ID         string            `yaml:"id" json:"id"`
	Type       ActionType        `yaml:"type" json:"type"`
	Command    string            `yaml:"command" json:"command"`
	Args       []string          `yaml:"args" json:"args"`
	Input      string            `yaml:"input" json:"input"`
	WorkingDir string            `yaml:"working_dir" json:"working_dir"`
	Env        map[string]string `yaml:"env" json:"env"`

	// Action-specific configuration
	Config map[string]interface{} `yaml:"config" json:"config"`

	// Runtime state
	Output    string     `yaml:"-" json:"output,omitempty"`
	Error     string     `yaml:"-" json:"error,omitempty"`
	ExitCode  int        `yaml:"-" json:"exit_code"`
	StartTime *time.Time `yaml:"-" json:"start_time,omitempty"`
	EndTime   *time.Time `yaml:"-" json:"end_time,omitempty"`

	// Action options
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	IgnoreError bool          `yaml:"ignore_error" json:"ignore_error"`
	Condition   string        `yaml:"condition" json:"condition"`
}

// Trigger represents a workflow trigger
type Trigger struct {
	ID        string                 `yaml:"id" json:"id"`
	Type      TriggerType            `yaml:"type" json:"type"`
	Schedule  string                 `yaml:"schedule" json:"schedule"`
	Event     string                 `yaml:"event" json:"event"`
	Path      string                 `yaml:"path" json:"path"`
	Condition string                 `yaml:"condition" json:"condition"`
	Config    map[string]interface{} `yaml:"config" json:"config"`
	Enabled   bool                   `yaml:"enabled" json:"enabled"`
}

// Condition represents a workflow condition
type Condition struct {
	Type       string                 `yaml:"type" json:"type"`
	Parameters map[string]interface{} `yaml:"parameters" json:"parameters"`
	Operator   string                 `yaml:"operator" json:"operator"`
}

// WorkflowExecution represents a single execution of a workflow
type WorkflowExecution struct {
	ID         string            `json:"id"`
	WorkflowID string            `json:"workflow_id"`
	Status     WorkflowStatus    `json:"status"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    *time.Time        `json:"end_time,omitempty"`
	Trigger    string            `json:"trigger"`
	Variables  map[string]string `json:"variables"`
	Tasks      []TaskExecution   `json:"tasks"`
	Error      string            `json:"error,omitempty"`
	Output     string            `json:"output,omitempty"`
}

// TaskExecution represents the execution state of a task
type TaskExecution struct {
	TaskID    string            `json:"task_id"`
	Status    TaskStatus        `json:"status"`
	StartTime *time.Time        `json:"start_time,omitempty"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	Error     string            `json:"error,omitempty"`
	Output    string            `json:"output,omitempty"`
	Actions   []ActionExecution `json:"actions"`
}

// ActionExecution represents the execution state of an action
type ActionExecution struct {
	ActionID  string     `json:"action_id"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Error     string     `json:"error,omitempty"`
	Output    string     `json:"output,omitempty"`
	ExitCode  int        `json:"exit_code"`
}

// ExecutionContext provides context for workflow execution
type ExecutionContext struct {
	Context   context.Context
	Variables map[string]string
	WorkDir   string
	Logger    Logger
	DryRun    bool
	Verbose   bool
}

// Logger interface for workflow logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// EventHandler interface for workflow events
type EventHandler interface {
	OnWorkflowStart(workflow *Workflow, execution *WorkflowExecution)
	OnWorkflowComplete(workflow *Workflow, execution *WorkflowExecution)
	OnWorkflowError(workflow *Workflow, execution *WorkflowExecution, err error)
	OnTaskStart(task *Task, execution *TaskExecution)
	OnTaskComplete(task *Task, execution *TaskExecution)
	OnTaskError(task *Task, execution *TaskExecution, err error)
}

// WorkflowEngine interface defines the main workflow engine
type WorkflowEngine interface {
	// Workflow management
	LoadWorkflow(definition []byte) (*Workflow, error)
	LoadWorkflowFromFile(path string) (*Workflow, error)
	RegisterWorkflow(workflow *Workflow) error
	GetWorkflow(id string) (*Workflow, error)
	ListWorkflows() ([]*Workflow, error)
	DeleteWorkflow(id string) error

	// Execution
	ExecuteWorkflow(workflowID string, variables map[string]string) (*WorkflowExecution, error)
	ExecuteWorkflowAsync(workflowID string, variables map[string]string) (string, error)
	GetExecution(executionID string) (*WorkflowExecution, error)
	ListExecutions(workflowID string) ([]*WorkflowExecution, error)
	CancelExecution(executionID string) error

	// Triggers
	RegisterTrigger(workflowID string, trigger *Trigger) error
	UnregisterTrigger(triggerID string) error
	EnableTrigger(triggerID string) error
	DisableTrigger(triggerID string) error

	// Events
	RegisterEventHandler(handler EventHandler)
	UnregisterEventHandler(handler EventHandler)

	// State management
	Start() error
	Stop() error
	IsRunning() bool
}

// TaskScheduler interface defines task scheduling capabilities
type TaskScheduler interface {
	ScheduleTask(task *Task, context *ExecutionContext) error
	CancelTask(taskID string) error
	GetTaskStatus(taskID string) (TaskStatus, error)
	ListPendingTasks() ([]*Task, error)
	ListRunningTasks() ([]*Task, error)
	SetMaxConcurrentTasks(max int)
	GetQueueDepth() int
}

// ActionExecutor interface defines action execution capabilities
type ActionExecutor interface {
	ExecuteAction(action *Action, context *ExecutionContext) (*ActionExecution, error)
	ValidateAction(action *Action) error
	GetSupportedActionTypes() []ActionType
	RegisterActionHandler(actionType ActionType, handler ActionHandler)
}

// ActionHandler interface for custom action implementations
type ActionHandler interface {
	Execute(action *Action, context *ExecutionContext) (*ActionExecution, error)
	Validate(action *Action) error
}

// StateMachine interface defines workflow state management
type StateMachine interface {
	GetState(workflowID string) (WorkflowStatus, error)
	SetState(workflowID string, status WorkflowStatus) error
	GetTaskState(workflowID string, taskID string) (TaskStatus, error)
	SetTaskState(workflowID string, taskID string, status TaskStatus) error
	SaveExecution(execution *WorkflowExecution) error
	LoadExecution(executionID string) (*WorkflowExecution, error)
	PersistState() error
	LoadState() error
}
