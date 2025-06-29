package workflow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileStateMachine implements the StateMachine interface using file-based persistence
type FileStateMachine struct {
	mu             sync.RWMutex
	stateDir       string
	logger         Logger
	workflowStates map[string]WorkflowStatus
	taskStates     map[string]map[string]TaskStatus // workflow_id -> task_id -> status
	executions     map[string]*WorkflowExecution
}

// NewStateMachine creates a new file-based state machine
func NewStateMachine(stateDir string, logger Logger) StateMachine {
	return &FileStateMachine{
		stateDir:       stateDir,
		logger:         logger,
		workflowStates: make(map[string]WorkflowStatus),
		taskStates:     make(map[string]map[string]TaskStatus),
		executions:     make(map[string]*WorkflowExecution),
	}
}

// GetState returns the current state of a workflow
func (sm *FileStateMachine) GetState(workflowID string) (WorkflowStatus, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if status, exists := sm.workflowStates[workflowID]; exists {
		return status, nil
	}

	// Try loading from file
	status, err := sm.loadWorkflowStateFromFile(workflowID)
	if err != nil {
		return "", fmt.Errorf("workflow state not found for %s", workflowID)
	}

	sm.workflowStates[workflowID] = status
	return status, nil
}

// SetState sets the state of a workflow
func (sm *FileStateMachine) SetState(workflowID string, status WorkflowStatus) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.workflowStates[workflowID] = status

	// Persist to file
	if err := sm.saveWorkflowStateToFile(workflowID, status); err != nil {
		sm.logger.Warn("Failed to persist workflow state: %v", err)
	}

	sm.logger.Debug("Set workflow %s state to %s", workflowID, status)
	return nil
}

// GetTaskState returns the current state of a task within a workflow
func (sm *FileStateMachine) GetTaskState(workflowID string, taskID string) (TaskStatus, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if workflowTasks, exists := sm.taskStates[workflowID]; exists {
		if status, exists := workflowTasks[taskID]; exists {
			return status, nil
		}
	}

	// Try loading from file
	status, err := sm.loadTaskStateFromFile(workflowID, taskID)
	if err != nil {
		return "", fmt.Errorf("task state not found for %s/%s", workflowID, taskID)
	}

	if sm.taskStates[workflowID] == nil {
		sm.taskStates[workflowID] = make(map[string]TaskStatus)
	}
	sm.taskStates[workflowID][taskID] = status

	return status, nil
}

// SetTaskState sets the state of a task within a workflow
func (sm *FileStateMachine) SetTaskState(workflowID string, taskID string, status TaskStatus) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.taskStates[workflowID] == nil {
		sm.taskStates[workflowID] = make(map[string]TaskStatus)
	}
	sm.taskStates[workflowID][taskID] = status

	// Persist to file
	if err := sm.saveTaskStateToFile(workflowID, taskID, status); err != nil {
		sm.logger.Warn("Failed to persist task state: %v", err)
	}

	sm.logger.Debug("Set task %s/%s state to %s", workflowID, taskID, status)
	return nil
}

// SaveExecution saves a workflow execution
func (sm *FileStateMachine) SaveExecution(execution *WorkflowExecution) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.executions[execution.ID] = execution

	// Persist to file
	if err := sm.saveExecutionToFile(execution); err != nil {
		return fmt.Errorf("failed to persist execution: %w", err)
	}

	sm.logger.Debug("Saved execution: %s", execution.ID)
	return nil
}

// LoadExecution loads a workflow execution
func (sm *FileStateMachine) LoadExecution(executionID string) (*WorkflowExecution, error) {
	sm.mu.RLock()
	if execution, exists := sm.executions[executionID]; exists {
		sm.mu.RUnlock()
		return execution, nil
	}
	sm.mu.RUnlock()

	// Load from file
	execution, err := sm.loadExecutionFromFile(executionID)
	if err != nil {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	sm.mu.Lock()
	sm.executions[executionID] = execution
	sm.mu.Unlock()

	return execution, nil
}

// PersistState saves all current state to disk
func (sm *FileStateMachine) PersistState() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create state directory if it doesn't exist
	if err := os.MkdirAll(sm.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Save workflow states
	for workflowID, status := range sm.workflowStates {
		if err := sm.saveWorkflowStateToFile(workflowID, status); err != nil {
			sm.logger.Warn("Failed to save workflow state %s: %v", workflowID, err)
		}
	}

	// Save task states
	for workflowID, tasks := range sm.taskStates {
		for taskID, status := range tasks {
			if err := sm.saveTaskStateToFile(workflowID, taskID, status); err != nil {
				sm.logger.Warn("Failed to save task state %s/%s: %v", workflowID, taskID, err)
			}
		}
	}

	// Save executions
	for _, execution := range sm.executions {
		if err := sm.saveExecutionToFile(execution); err != nil {
			sm.logger.Warn("Failed to save execution %s: %v", execution.ID, err)
		}
	}

	sm.logger.Info("State persisted successfully")
	return nil
}

// LoadState loads all state from disk
func (sm *FileStateMachine) LoadState() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if state directory exists
	if _, err := os.Stat(sm.stateDir); os.IsNotExist(err) {
		sm.logger.Info("State directory doesn't exist, starting with clean state")
		return nil
	}

	// Load workflow states
	workflowsDir := filepath.Join(sm.stateDir, "workflows")
	if _, err := os.Stat(workflowsDir); err == nil {
		files, err := ioutil.ReadDir(workflowsDir)
		if err == nil {
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".json" {
					workflowID := strings.TrimSuffix(file.Name(), ".json")
					if status, err := sm.loadWorkflowStateFromFile(workflowID); err == nil {
						sm.workflowStates[workflowID] = status
					}
				}
			}
		}
	}

	// Load task states
	tasksDir := filepath.Join(sm.stateDir, "tasks")
	if _, err := os.Stat(tasksDir); err == nil {
		workflowDirs, err := ioutil.ReadDir(tasksDir)
		if err == nil {
			for _, workflowDir := range workflowDirs {
				if workflowDir.IsDir() {
					workflowID := workflowDir.Name()
					taskFiles, err := ioutil.ReadDir(filepath.Join(tasksDir, workflowID))
					if err == nil {
						sm.taskStates[workflowID] = make(map[string]TaskStatus)
						for _, taskFile := range taskFiles {
							if filepath.Ext(taskFile.Name()) == ".json" {
								taskID := strings.TrimSuffix(taskFile.Name(), ".json")
								if status, err := sm.loadTaskStateFromFile(workflowID, taskID); err == nil {
									sm.taskStates[workflowID][taskID] = status
								}
							}
						}
					}
				}
			}
		}
	}

	// Load executions
	executionsDir := filepath.Join(sm.stateDir, "executions")
	if _, err := os.Stat(executionsDir); err == nil {
		files, err := ioutil.ReadDir(executionsDir)
		if err == nil {
			for _, file := range files {
				if filepath.Ext(file.Name()) == ".json" {
					executionID := strings.TrimSuffix(file.Name(), ".json")
					if execution, err := sm.loadExecutionFromFile(executionID); err == nil {
						sm.executions[executionID] = execution
					}
				}
			}
		}
	}

	sm.logger.Info("State loaded successfully")
	return nil
}

// Private helper methods

func (sm *FileStateMachine) saveWorkflowStateToFile(workflowID string, status WorkflowStatus) error {
	dir := filepath.Join(sm.stateDir, "workflows")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data := map[string]interface{}{
		"workflow_id": workflowID,
		"status":      status,
		"timestamp":   time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s.json", workflowID))
	return ioutil.WriteFile(filename, jsonData, 0644)
}

func (sm *FileStateMachine) loadWorkflowStateFromFile(workflowID string) (WorkflowStatus, error) {
	filename := filepath.Join(sm.stateDir, "workflows", fmt.Sprintf("%s.json", workflowID))

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var stateData map[string]interface{}
	if err := json.Unmarshal(data, &stateData); err != nil {
		return "", err
	}

	status, ok := stateData["status"].(string)
	if !ok {
		return "", fmt.Errorf("invalid status format")
	}

	return WorkflowStatus(status), nil
}

func (sm *FileStateMachine) saveTaskStateToFile(workflowID, taskID string, status TaskStatus) error {
	dir := filepath.Join(sm.stateDir, "tasks", workflowID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data := map[string]interface{}{
		"workflow_id": workflowID,
		"task_id":     taskID,
		"status":      status,
		"timestamp":   time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s.json", taskID))
	return ioutil.WriteFile(filename, jsonData, 0644)
}

func (sm *FileStateMachine) loadTaskStateFromFile(workflowID, taskID string) (TaskStatus, error) {
	filename := filepath.Join(sm.stateDir, "tasks", workflowID, fmt.Sprintf("%s.json", taskID))

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var stateData map[string]interface{}
	if err := json.Unmarshal(data, &stateData); err != nil {
		return "", err
	}

	status, ok := stateData["status"].(string)
	if !ok {
		return "", fmt.Errorf("invalid status format")
	}

	return TaskStatus(status), nil
}

func (sm *FileStateMachine) saveExecutionToFile(execution *WorkflowExecution) error {
	dir := filepath.Join(sm.stateDir, "executions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(execution, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s.json", execution.ID))
	return ioutil.WriteFile(filename, jsonData, 0644)
}

func (sm *FileStateMachine) loadExecutionFromFile(executionID string) (*WorkflowExecution, error) {
	filename := filepath.Join(sm.stateDir, "executions", fmt.Sprintf("%s.json", executionID))

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var execution WorkflowExecution
	if err := json.Unmarshal(data, &execution); err != nil {
		return nil, err
	}

	return &execution, nil
}
