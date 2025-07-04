package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Scheduler implements the TaskScheduler interface
type Scheduler struct {
	mu             sync.RWMutex
	pendingTasks   []*Task
	runningTasks   map[string]*TaskExecution
	completedTasks []*TaskExecution
	maxConcurrent  int
	workerPool     chan struct{}
	ctx            context.Context
	cancel         context.CancelFunc
	logger         Logger
	taskChan       chan *ScheduledTask
	stopped        bool
}

// ScheduledTask represents a task scheduled for execution
type ScheduledTask struct {
	Task      *Task
	Context   *ExecutionContext
	Execution *TaskExecution
	Callback  func(*TaskExecution, error)
}

// NewTaskScheduler creates a new task scheduler
func NewTaskScheduler(logger Logger) TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	scheduler := &Scheduler{
		pendingTasks:   make([]*Task, 0),
		runningTasks:   make(map[string]*TaskExecution),
		completedTasks: make([]*TaskExecution, 0),
		maxConcurrent:  5, // Default max concurrent tasks
		workerPool:     make(chan struct{}, 5),
		ctx:            ctx,
		cancel:         cancel,
		logger:         logger,
		taskChan:       make(chan *ScheduledTask, 100),
		stopped:        false,
	}

	// Initialize worker pool
	for i := 0; i < scheduler.maxConcurrent; i++ {
		scheduler.workerPool <- struct{}{}
	}

	// Start worker goroutines
	go scheduler.startWorkers()

	return scheduler
}

// ScheduleTask schedules a task for execution
func (s *Scheduler) ScheduleTask(task *Task, context *ExecutionContext) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return fmt.Errorf("scheduler is stopped")
	}

	// Create task execution record
	execution := &TaskExecution{
		TaskID:  task.ID,
		Status:  TaskStatusPending,
		Actions: make([]ActionExecution, 0),
	}

	// Add to pending queue
	s.pendingTasks = append(s.pendingTasks, task)

	// Create scheduled task
	scheduledTask := &ScheduledTask{
		Task:      task,
		Context:   context,
		Execution: execution,
		Callback:  s.taskCompletionCallback,
	}

	// Send to worker channel
	select {
	case s.taskChan <- scheduledTask:
		s.logger.Debug("Task %s scheduled for execution", task.ID)
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// CancelTask cancels a pending or running task
func (s *Scheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check running tasks
	if execution, exists := s.runningTasks[taskID]; exists {
		execution.Status = TaskStatusFailed
		execution.Error = "Task cancelled"
		endTime := time.Now()
		execution.EndTime = &endTime
		delete(s.runningTasks, taskID)
		s.completedTasks = append(s.completedTasks, execution)
		s.logger.Info("Cancelled running task: %s", taskID)
		return nil
	}

	// Check pending tasks
	for i, task := range s.pendingTasks {
		if task.ID == taskID {
			s.pendingTasks = append(s.pendingTasks[:i], s.pendingTasks[i+1:]...)
			s.logger.Info("Cancelled pending task: %s", taskID)
			return nil
		}
	}

	return fmt.Errorf("task %s not found", taskID)
}

// GetTaskStatus returns the status of a task
func (s *Scheduler) GetTaskStatus(taskID string) (TaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check running tasks
	if execution, exists := s.runningTasks[taskID]; exists {
		return execution.Status, nil
	}

	// Check completed tasks
	for _, execution := range s.completedTasks {
		if execution.TaskID == taskID {
			return execution.Status, nil
		}
	}

	// Check pending tasks
	for _, task := range s.pendingTasks {
		if task.ID == taskID {
			return TaskStatusPending, nil
		}
	}

	return "", fmt.Errorf("task %s not found", taskID)
}

// ListPendingTasks returns all pending tasks
func (s *Scheduler) ListPendingTasks() ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	tasks := make([]*Task, len(s.pendingTasks))
	copy(tasks, s.pendingTasks)

	return tasks, nil
}

// ListRunningTasks returns all running tasks
func (s *Scheduler) ListRunningTasks() ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.runningTasks))
	for _, execution := range s.runningTasks {
		// Note: We don't have the original task here, just the execution
		// In a real implementation, we'd need to store the task reference
		task := &Task{
			ID:     execution.TaskID,
			Status: execution.Status,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// SetMaxConcurrentTasks sets the maximum number of concurrent tasks
func (s *Scheduler) SetMaxConcurrentTasks(max int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if max < 1 {
		max = 1
	}

	oldMax := s.maxConcurrent
	s.maxConcurrent = max

	// Adjust worker pool
	if max > oldMax {
		// Add more workers
		for i := oldMax; i < max; i++ {
			s.workerPool <- struct{}{}
		}
	} else if max < oldMax {
		// Remove workers (they'll exit when they try to return to pool)
		for i := max; i < oldMax; i++ {
			select {
			case <-s.workerPool:
			default:
				// Pool is empty, workers will exit naturally
			}
		}
	}

	// Recreate worker pool channel with new size
	newPool := make(chan struct{}, max)
	for i := 0; i < max; i++ {
		select {
		case <-s.workerPool:
			newPool <- struct{}{}
		default:
			newPool <- struct{}{}
		}
	}
	s.workerPool = newPool

	s.logger.Info("Max concurrent tasks set to: %d", max)
}

// GetQueueDepth returns the number of pending tasks
func (s *Scheduler) GetQueueDepth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pendingTasks)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return
	}

	s.stopped = true
	s.cancel()
	close(s.taskChan)

	s.logger.Info("Task scheduler stopped")
}

// Internal methods

func (s *Scheduler) startWorkers() {
	for {
		select {
		case scheduledTask, ok := <-s.taskChan:
			if !ok {
				s.logger.Debug("Task channel closed, stopping workers")
				return
			}

			// Wait for available worker slot
			<-s.workerPool

			// Execute task in goroutine
			go s.executeTaskWorker(scheduledTask)

		case <-s.ctx.Done():
			s.logger.Debug("Scheduler context cancelled, stopping workers")
			return
		}
	}
}

func (s *Scheduler) executeTaskWorker(scheduledTask *ScheduledTask) {
	defer func() {
		// Return worker to pool
		s.workerPool <- struct{}{}

		// Handle panic
		if r := recover(); r != nil {
			s.logger.Error("Task worker panic: %v", r)
			scheduledTask.Execution.Status = TaskStatusFailed
			scheduledTask.Execution.Error = fmt.Sprintf("Worker panic: %v", r)
			endTime := time.Now()
			scheduledTask.Execution.EndTime = &endTime

			if scheduledTask.Callback != nil {
				scheduledTask.Callback(scheduledTask.Execution, fmt.Errorf("worker panic: %v", r))
			}
		}
	}()

	// Move task from pending to running
	s.moveTaskToRunning(scheduledTask.Task, scheduledTask.Execution)

	// Execute the task
	err := s.executeTask(scheduledTask.Task, scheduledTask.Execution, scheduledTask.Context)

	// Move task from running to completed
	s.moveTaskToCompleted(scheduledTask.Execution)

	// Call completion callback
	if scheduledTask.Callback != nil {
		scheduledTask.Callback(scheduledTask.Execution, err)
	}
}

func (s *Scheduler) executeTask(task *Task, execution *TaskExecution, ctx *ExecutionContext) error {
	s.logger.Info("Executing task: %s", task.ID)

	startTime := time.Now()
	execution.StartTime = &startTime
	execution.Status = TaskStatusRunning

	// Set task timeout
	taskCtx := ctx.Context
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithTimeout(ctx.Context, task.Timeout)
		defer cancel()
	}

	// Create task-specific execution context
	taskCtx = context.WithValue(taskCtx, "task_id", task.ID)
	taskExecCtx := &ExecutionContext{
		Context:   taskCtx,
		Variables: s.mergeTaskVariables(ctx.Variables, task.Variables),
		WorkDir:   ctx.WorkDir,
		Logger:    ctx.Logger,
		DryRun:    ctx.DryRun,
		Verbose:   ctx.Verbose,
	}

	// Execute actions
	for i, action := range task.Actions {
		action := action // Create copy to avoid closure issues

		s.logger.Debug("Executing action %d/%d in task %s", i+1, len(task.Actions), task.ID)

		// Check if task was cancelled
		select {
		case <-taskCtx.Done():
			execution.Status = TaskStatusFailed
			execution.Error = "Task cancelled or timed out"
			endTime := time.Now()
			execution.EndTime = &endTime
			return fmt.Errorf("task cancelled or timed out")
		default:
		}

		// Execute action (this would need to be implemented)
		actionExecution, err := s.executeAction(&action, taskExecCtx)
		if err != nil {
			execution.Status = TaskStatusFailed
			execution.Error = err.Error()
			endTime := time.Now()
			execution.EndTime = &endTime
			s.logger.Error("Action failed in task %s: %v", task.ID, err)
			return err
		}

		execution.Actions = append(execution.Actions, *actionExecution)
	}

	// Task completed successfully
	execution.Status = TaskStatusCompleted
	endTime := time.Now()
	execution.EndTime = &endTime

	s.logger.Info("Task completed successfully: %s", task.ID)
	return nil
}

func (s *Scheduler) executeAction(action *Action, ctx *ExecutionContext) (*ActionExecution, error) {
	// Delegate to the actual ActionExecutor for real action execution
	s.logger.Debug("Executing action: %s (type: %s)", action.ID, action.Type)

	startTime := time.Now()
	execution := &ActionExecution{
		ActionID:  action.ID,
		StartTime: &startTime,
		ExitCode:  0,
	}

	// Create a real action executor to handle the action
	executor := NewActionExecutor(s.logger)
	
	// Execute the action using the real executor
	result, err := executor.ExecuteAction(action, ctx)
	if err != nil {
		execution.Error = err.Error()
		execution.ExitCode = 1
		if result != nil && result.ExitCode != 0 {
			execution.ExitCode = result.ExitCode
		}
		endTime := time.Now()
		execution.EndTime = &endTime
		return execution, err
	}

	// Copy results from the executor's result
	if result != nil {
		execution.Output = result.Output
		execution.Error = result.Error
		execution.ExitCode = result.ExitCode
		if result.EndTime != nil {
			execution.EndTime = result.EndTime
		} else {
			endTime := time.Now()
			execution.EndTime = &endTime
		}
	} else {
		endTime := time.Now()
		execution.EndTime = &endTime
	}

	return execution, nil
}

func (s *Scheduler) moveTaskToRunning(task *Task, execution *TaskExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from pending
	for i, pendingTask := range s.pendingTasks {
		if pendingTask.ID == task.ID {
			s.pendingTasks = append(s.pendingTasks[:i], s.pendingTasks[i+1:]...)
			break
		}
	}

	// Add to running
	s.runningTasks[task.ID] = execution
	execution.Status = TaskStatusRunning
}

func (s *Scheduler) moveTaskToCompleted(execution *TaskExecution) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from running
	delete(s.runningTasks, execution.TaskID)

	// Add to completed
	s.completedTasks = append(s.completedTasks, execution)
}

func (s *Scheduler) taskCompletionCallback(execution *TaskExecution, err error) {
	if err != nil {
		s.logger.Error("Task %s completed with error: %v", execution.TaskID, err)
	} else {
		s.logger.Info("Task %s completed successfully", execution.TaskID)
	}
}

func (s *Scheduler) mergeTaskVariables(contextVars, taskVars map[string]string) map[string]string {
	result := make(map[string]string)

	// Add context variables first
	for k, v := range contextVars {
		result[k] = v
	}

	// Override with task variables
	for k, v := range taskVars {
		result[k] = v
	}

	return result
}
