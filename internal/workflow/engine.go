package workflow

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Engine implements the WorkflowEngine interface
type Engine struct {
	mu                 sync.RWMutex
	workflows          map[string]*Workflow
	executions         map[string]*WorkflowExecution
	triggers           map[string]*Trigger
	eventHandlers      []EventHandler
	scheduler          TaskScheduler
	executor           ActionExecutor
	stateMachine       StateMachine
	errorHandler       *ErrorHandler
	conditionEvaluator *ConditionEvaluator
	logger             Logger
	running            bool
	workDir            string
	stateDir           string
}

// NewEngine creates a new workflow engine instance
func NewEngine(workDir, stateDir string, logger Logger) *Engine {
	engine := &Engine{
		workflows:          make(map[string]*Workflow),
		executions:         make(map[string]*WorkflowExecution),
		triggers:           make(map[string]*Trigger),
		eventHandlers:      make([]EventHandler, 0),
		errorHandler:       NewErrorHandler(logger),
		conditionEvaluator: NewConditionEvaluator(),
		logger:             logger,
		workDir:            workDir,
		stateDir:           stateDir,
		running:            false,
	}

	// Initialize components
	engine.scheduler = NewTaskScheduler(logger)
	engine.executor = NewActionExecutor(logger)
	engine.stateMachine = NewStateMachine(stateDir, logger)

	return engine
}

// LoadWorkflow loads a workflow from YAML bytes
func (e *Engine) LoadWorkflow(definition []byte) (*Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(definition, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %w", err)
	}

	// Validate workflow
	if err := e.validateWorkflow(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Initialize runtime state
	workflow.Status = StatusPending
	workflow.CurrentTask = 0

	return &workflow, nil
}

// LoadWorkflowFromFile loads a workflow from a YAML file
func (e *Engine) LoadWorkflowFromFile(path string) (*Workflow, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	workflow, err := e.LoadWorkflow(data)
	if err != nil {
		return nil, err
	}

	// Store file path in metadata
	if workflow.Metadata == nil {
		workflow.Metadata = make(map[string]string)
	}
	workflow.Metadata["source_file"] = path

	return workflow, nil
}

// RegisterWorkflow registers a workflow in the engine
func (e *Engine) RegisterWorkflow(workflow *Workflow) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if workflow.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if _, exists := e.workflows[workflow.ID]; exists {
		return fmt.Errorf("workflow with ID %s already exists", workflow.ID)
	}

	e.workflows[workflow.ID] = workflow
	e.logger.Info("Registered workflow: %s", workflow.ID)

	// Register triggers
	for _, trigger := range workflow.Triggers {
		trigger := trigger // Create a copy to avoid closure issues
		if err := e.RegisterTrigger(workflow.ID, &trigger); err != nil {
			e.logger.Warn("Failed to register trigger %s: %v", trigger.ID, err)
		}
	}

	return nil
}

// GetWorkflow retrieves a workflow by ID
func (e *Engine) GetWorkflow(id string) (*Workflow, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	workflow, exists := e.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	return workflow, nil
}

// ListWorkflows returns all registered workflows
func (e *Engine) ListWorkflows() ([]*Workflow, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	workflows := make([]*Workflow, 0, len(e.workflows))
	for _, workflow := range e.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

// DeleteWorkflow removes a workflow from the engine
func (e *Engine) DeleteWorkflow(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.workflows[id]; !exists {
		return fmt.Errorf("workflow %s not found", id)
	}

	// Remove associated triggers
	for triggerID, trigger := range e.triggers {
		if trigger.Config != nil {
			if workflowID, ok := trigger.Config["workflow_id"].(string); ok && workflowID == id {
				delete(e.triggers, triggerID)
			}
		}
	}

	delete(e.workflows, id)
	e.logger.Info("Deleted workflow: %s", id)

	return nil
}

// ExecuteWorkflow executes a workflow synchronously
func (e *Engine) ExecuteWorkflow(workflowID string, variables map[string]string) (*WorkflowExecution, error) {
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	execution := &WorkflowExecution{
		ID:         e.generateExecutionID(),
		WorkflowID: workflowID,
		Status:     StatusPending,
		StartTime:  time.Now(),
		Trigger:    "manual",
		Variables:  variables,
		Tasks:      make([]TaskExecution, 0),
	}

	// Store execution
	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.mu.Unlock()

	// Create execution context
	ctx := &ExecutionContext{
		Context:   context.Background(),
		Variables: e.mergeVariables(workflow.Variables, variables),
		WorkDir:   e.workDir,
		Logger:    e.logger,
		DryRun:    false,
		Verbose:   false,
	}

	// Execute workflow
	err = e.executeWorkflowInternal(workflow, execution, ctx)
	if err != nil {
		execution.Status = StatusFailed
		execution.Error = err.Error()
		endTime := time.Now()
		execution.EndTime = &endTime
		e.notifyWorkflowError(workflow, execution, err)
	} else {
		execution.Status = StatusCompleted
		endTime := time.Now()
		execution.EndTime = &endTime
		e.notifyWorkflowComplete(workflow, execution)
	}

	// Save execution state
	if err := e.stateMachine.SaveExecution(execution); err != nil {
		e.logger.Warn("Failed to save execution state: %v", err)
	}

	return execution, err
}

// ExecuteWorkflowAsync executes a workflow asynchronously
func (e *Engine) ExecuteWorkflowAsync(workflowID string, variables map[string]string) (string, error) {
	workflow, err := e.GetWorkflow(workflowID)
	if err != nil {
		return "", err
	}

	execution := &WorkflowExecution{
		ID:         e.generateExecutionID(),
		WorkflowID: workflowID,
		Status:     StatusPending,
		StartTime:  time.Now(),
		Trigger:    "manual",
		Variables:  variables,
		Tasks:      make([]TaskExecution, 0),
	}

	// Store execution
	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.mu.Unlock()

	// Execute asynchronously
	go func() {
		ctx := &ExecutionContext{
			Context:   context.Background(),
			Variables: e.mergeVariables(workflow.Variables, variables),
			WorkDir:   e.workDir,
			Logger:    e.logger,
			DryRun:    false,
			Verbose:   false,
		}

		err := e.executeWorkflowInternal(workflow, execution, ctx)
		if err != nil {
			execution.Status = StatusFailed
			execution.Error = err.Error()
			endTime := time.Now()
			execution.EndTime = &endTime
			e.notifyWorkflowError(workflow, execution, err)
		} else {
			execution.Status = StatusCompleted
			endTime := time.Now()
			execution.EndTime = &endTime
			e.notifyWorkflowComplete(workflow, execution)
		}

		// Save execution state
		if err := e.stateMachine.SaveExecution(execution); err != nil {
			e.logger.Warn("Failed to save execution state: %v", err)
		}
	}()

	return execution.ID, nil
}

// GetExecution retrieves an execution by ID
func (e *Engine) GetExecution(executionID string) (*WorkflowExecution, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	execution, exists := e.executions[executionID]
	if !exists {
		// Try loading from persistent state
		loaded, err := e.stateMachine.LoadExecution(executionID)
		if err != nil {
			return nil, fmt.Errorf("execution %s not found", executionID)
		}
		return loaded, nil
	}

	return execution, nil
}

// ListExecutions returns all executions for a workflow
func (e *Engine) ListExecutions(workflowID string) ([]*WorkflowExecution, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	executions := make([]*WorkflowExecution, 0)
	for _, execution := range e.executions {
		if execution.WorkflowID == workflowID {
			executions = append(executions, execution)
		}
	}

	return executions, nil
}

// CancelExecution cancels a running execution
func (e *Engine) CancelExecution(executionID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	execution, exists := e.executions[executionID]
	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	if execution.Status != StatusRunning {
		return fmt.Errorf("execution %s is not running", executionID)
	}

	execution.Status = StatusCancelled
	endTime := time.Now()
	execution.EndTime = &endTime

	// TODO: Cancel running tasks

	e.logger.Info("Cancelled execution: %s", executionID)
	return nil
}

// RegisterTrigger registers a trigger for a workflow
func (e *Engine) RegisterTrigger(workflowID string, trigger *Trigger) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if trigger.Config == nil {
		trigger.Config = make(map[string]interface{})
	}
	trigger.Config["workflow_id"] = workflowID

	e.triggers[trigger.ID] = trigger
	e.logger.Info("Registered trigger: %s for workflow: %s", trigger.ID, workflowID)

	return nil
}

// UnregisterTrigger removes a trigger
func (e *Engine) UnregisterTrigger(triggerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.triggers[triggerID]; !exists {
		return fmt.Errorf("trigger %s not found", triggerID)
	}

	delete(e.triggers, triggerID)
	e.logger.Info("Unregistered trigger: %s", triggerID)

	return nil
}

// EnableTrigger enables a trigger
func (e *Engine) EnableTrigger(triggerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	trigger, exists := e.triggers[triggerID]
	if !exists {
		return fmt.Errorf("trigger %s not found", triggerID)
	}

	trigger.Enabled = true
	e.logger.Info("Enabled trigger: %s", triggerID)

	return nil
}

// DisableTrigger disables a trigger
func (e *Engine) DisableTrigger(triggerID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	trigger, exists := e.triggers[triggerID]
	if !exists {
		return fmt.Errorf("trigger %s not found", triggerID)
	}

	trigger.Enabled = false
	e.logger.Info("Disabled trigger: %s", triggerID)

	return nil
}

// RegisterEventHandler registers an event handler
func (e *Engine) RegisterEventHandler(handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.eventHandlers = append(e.eventHandlers, handler)
}

// UnregisterEventHandler removes an event handler
func (e *Engine) UnregisterEventHandler(handler EventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, h := range e.eventHandlers {
		if h == handler {
			e.eventHandlers = append(e.eventHandlers[:i], e.eventHandlers[i+1:]...)
			break
		}
	}
}

// Start starts the workflow engine
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("engine is already running")
	}

	// Create state directory if it doesn't exist
	if err := os.MkdirAll(e.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Load persistent state
	if err := e.stateMachine.LoadState(); err != nil {
		e.logger.Warn("Failed to load persistent state: %v", err)
	}

	e.running = true
	e.logger.Info("Workflow engine started")

	return nil
}

// Stop stops the workflow engine
func (e *Engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return fmt.Errorf("engine is not running")
	}

	// Save persistent state
	if err := e.stateMachine.PersistState(); err != nil {
		e.logger.Warn("Failed to persist state: %v", err)
	}

	e.running = false
	e.logger.Info("Workflow engine stopped")

	return nil
}

// IsRunning returns whether the engine is running
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Internal helper methods

func (e *Engine) executeWorkflowInternal(workflow *Workflow, execution *WorkflowExecution, ctx *ExecutionContext) error {
	e.logger.Info("Starting workflow execution: %s", workflow.ID)
	execution.Status = StatusRunning
	e.notifyWorkflowStart(workflow, execution)

	// Execute tasks
	for i, task := range workflow.Tasks {
		workflow.CurrentTask = i

		taskExecution := &TaskExecution{
			TaskID:  task.ID,
			Status:  TaskStatusPending,
			Actions: make([]ActionExecution, 0),
		}
		execution.Tasks = append(execution.Tasks, *taskExecution)

		// Check dependencies
		if err := e.checkTaskDependencies(&task, execution); err != nil {
			if task.Optional {
				e.logger.Warn("Optional task %s skipped due to dependency failure: %v", task.ID, err)
				taskExecution.Status = TaskStatusSkipped
				continue
			}
			return fmt.Errorf("task %s dependency check failed: %w", task.ID, err)
		}

		// Execute task
		if err := e.executeTask(&task, taskExecution, ctx); err != nil {
			taskExecution.Status = TaskStatusFailed
			taskExecution.Error = err.Error()

			if task.Critical && !workflow.ContinueOnError {
				return fmt.Errorf("critical task %s failed: %w", task.ID, err)
			}

			e.logger.Warn("Task %s failed: %v", task.ID, err)
			continue
		}

		taskExecution.Status = TaskStatusCompleted
	}

	e.logger.Info("Workflow execution completed: %s", workflow.ID)
	return nil
}

func (e *Engine) executeTask(task *Task, execution *TaskExecution, ctx *ExecutionContext) error {
	e.logger.Info("Executing task: %s", task.ID)

	startTime := time.Now()
	execution.StartTime = &startTime
	execution.Status = TaskStatusRunning

	e.notifyTaskStart(task, execution)

	// Execute actions
	for _, action := range task.Actions {
		actionExecution, err := e.executor.ExecuteAction(&action, ctx)
		if err != nil {
			execution.Status = TaskStatusFailed
			execution.Error = err.Error()
			e.notifyTaskError(task, execution, err)
			return err
		}

		execution.Actions = append(execution.Actions, *actionExecution)
	}

	endTime := time.Now()
	execution.EndTime = &endTime
	execution.Status = TaskStatusCompleted

	e.notifyTaskComplete(task, execution)
	e.logger.Info("Task completed: %s", task.ID)

	return nil
}

func (e *Engine) checkTaskDependencies(task *Task, execution *WorkflowExecution) error {
	for _, depID := range task.DependsOn {
		found := false
		for _, taskExec := range execution.Tasks {
			if taskExec.TaskID == depID {
				if taskExec.Status != TaskStatusCompleted {
					return fmt.Errorf("dependency task %s has status %s", depID, taskExec.Status)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("dependency task %s not found", depID)
		}
	}
	return nil
}

func (e *Engine) validateWorkflow(workflow *Workflow) error {
	if workflow.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if len(workflow.Tasks) == 0 {
		return fmt.Errorf("workflow must have at least one task")
	}

	// Validate tasks
	taskIDs := make(map[string]bool)
	for _, task := range workflow.Tasks {
		if task.ID == "" {
			return fmt.Errorf("task ID is required")
		}
		if taskIDs[task.ID] {
			return fmt.Errorf("duplicate task ID: %s", task.ID)
		}
		taskIDs[task.ID] = true

		// Validate task dependencies
		for _, depID := range task.DependsOn {
			if !taskIDs[depID] {
				return fmt.Errorf("task %s depends on unknown task %s", task.ID, depID)
			}
		}

		// Validate actions
		for _, action := range task.Actions {
			if err := e.executor.ValidateAction(&action); err != nil {
				return fmt.Errorf("invalid action in task %s: %w", task.ID, err)
			}
		}
	}

	return nil
}

func (e *Engine) mergeVariables(workflowVars, executionVars map[string]string) map[string]string {
	result := make(map[string]string)

	// Add workflow variables first
	for k, v := range workflowVars {
		result[k] = v
	}

	// Override with execution variables
	for k, v := range executionVars {
		result[k] = v
	}

	return result
}

func (e *Engine) generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// Event notification methods
func (e *Engine) notifyWorkflowStart(workflow *Workflow, execution *WorkflowExecution) {
	for _, handler := range e.eventHandlers {
		handler.OnWorkflowStart(workflow, execution)
	}
}

func (e *Engine) notifyWorkflowComplete(workflow *Workflow, execution *WorkflowExecution) {
	for _, handler := range e.eventHandlers {
		handler.OnWorkflowComplete(workflow, execution)
	}
}

func (e *Engine) notifyWorkflowError(workflow *Workflow, execution *WorkflowExecution, err error) {
	for _, handler := range e.eventHandlers {
		handler.OnWorkflowError(workflow, execution, err)
	}
}

func (e *Engine) notifyTaskStart(task *Task, execution *TaskExecution) {
	for _, handler := range e.eventHandlers {
		handler.OnTaskStart(task, execution)
	}
}

func (e *Engine) notifyTaskComplete(task *Task, execution *TaskExecution) {
	for _, handler := range e.eventHandlers {
		handler.OnTaskComplete(task, execution)
	}
}

func (e *Engine) notifyTaskError(task *Task, execution *TaskExecution, err error) {
	for _, handler := range e.eventHandlers {
		handler.OnTaskError(task, execution, err)
	}
}
