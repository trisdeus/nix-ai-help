package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// WorkflowStorage manages workflow storage and retrieval
type WorkflowStorage struct {
	mu             sync.RWMutex
	workflowDirs   []string
	workflows      map[string]*WorkflowDefinition
	workflowFiles  map[string]string // workflow ID -> file path
	parser         *WorkflowParser
	logger         Logger
	autoReload     bool
	lastReloadTime time.Time
}

// WorkflowStorageConfig configures the workflow storage
type WorkflowStorageConfig struct {
	WorkflowDirs []string `yaml:"workflowDirs"`
	AutoReload   bool     `yaml:"autoReload"`
}

// NewWorkflowStorage creates a new workflow storage
func NewWorkflowStorage(config WorkflowStorageConfig, logger Logger) *WorkflowStorage {
	storage := &WorkflowStorage{
		workflowDirs:  config.WorkflowDirs,
		workflows:     make(map[string]*WorkflowDefinition),
		workflowFiles: make(map[string]string),
		parser:        NewWorkflowParser(logger),
		logger:        logger,
		autoReload:    config.AutoReload,
	}

	// Set default workflow directories if none provided
	if len(storage.workflowDirs) == 0 {
		storage.workflowDirs = []string{
			"/etc/nixai/workflows",
			filepath.Join(os.Getenv("HOME"), ".config/nixai/workflows"),
			"./workflows",
		}
	}

	return storage
}

// LoadWorkflows loads all workflows from configured directories
func (s *WorkflowStorage) LoadWorkflows() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.workflows = make(map[string]*WorkflowDefinition)
	s.workflowFiles = make(map[string]string)

	for _, dir := range s.workflowDirs {
		if err := s.loadWorkflowsFromDir(dir); err != nil {
			// Log error but continue with other directories
			s.logger.Warn("Failed to load workflows from directory", "dir", dir, "error", err)
		}
	}

	s.lastReloadTime = time.Now()
	s.logger.Info("Loaded workflows", "count", len(s.workflows))

	return nil
}

// loadWorkflowsFromDir loads workflows from a specific directory
func (s *WorkflowStorage) loadWorkflowsFromDir(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	workflows, err := s.parser.ParseWorkflowDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to parse workflows in directory %s: %w", dir, err)
	}

	// Add workflows to storage
	for id, workflow := range workflows {
		if existingWorkflow, exists := s.workflows[id]; exists {
			s.logger.Warn("Workflow ID conflict, overriding",
				"id", id,
				"existing", existingWorkflow.Name,
				"new", workflow.Name)
		}

		s.workflows[id] = workflow
		s.workflowFiles[id] = filepath.Join(dir, id+".yaml")
	}

	return nil
}

// GetWorkflow retrieves a workflow by ID
func (s *WorkflowStorage) GetWorkflow(id string) (*WorkflowDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workflow, exists := s.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}

	return workflow, nil
}

// ListWorkflows returns all available workflows
func (s *WorkflowStorage) ListWorkflows() map[string]*WorkflowDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]*WorkflowDefinition)
	for id, workflow := range s.workflows {
		result[id] = workflow
	}

	return result
}

// GetWorkflowsByTag returns workflows that have the specified tag
func (s *WorkflowStorage) GetWorkflowsByTag(tag string) map[string]*WorkflowDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*WorkflowDefinition)
	for id, workflow := range s.workflows {
		for _, workflowTag := range workflow.Tags {
			if workflowTag == tag {
				result[id] = workflow
				break
			}
		}
	}

	return result
}

// SaveWorkflow saves a workflow to storage
func (s *WorkflowStorage) SaveWorkflow(id string, workflow *WorkflowDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Choose directory for saving (first writable directory)
	var saveDir string
	for _, dir := range s.workflowDirs {
		if s.isWritableDir(dir) {
			saveDir = dir
			break
		}
	}

	if saveDir == "" {
		// Create user config directory if no writable directory found
		saveDir = filepath.Join(os.Getenv("HOME"), ".config/nixai/workflows")
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			return fmt.Errorf("failed to create workflow directory %s: %w", saveDir, err)
		}
	}

	// Save workflow to file
	filename := filepath.Join(saveDir, id+".yaml")
	if err := s.saveWorkflowToFile(filename, workflow); err != nil {
		return fmt.Errorf("failed to save workflow to file %s: %w", filename, err)
	}

	// Update in-memory storage
	s.workflows[id] = workflow
	s.workflowFiles[id] = filename

	s.logger.Info("Saved workflow", "id", id, "file", filename)
	return nil
}

// DeleteWorkflow removes a workflow from storage
func (s *WorkflowStorage) DeleteWorkflow(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if workflow exists
	if _, exists := s.workflows[id]; !exists {
		return fmt.Errorf("workflow not found: %s", id)
	}

	// Remove file if it exists
	if filename, exists := s.workflowFiles[id]; exists {
		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			s.logger.Warn("Failed to remove workflow file", "file", filename, "error", err)
		}
	}

	// Remove from in-memory storage
	delete(s.workflows, id)
	delete(s.workflowFiles, id)

	s.logger.Info("Deleted workflow", "id", id)
	return nil
}

// SearchWorkflows searches workflows by name, description, or tags
func (s *WorkflowStorage) SearchWorkflows(query string) map[string]*WorkflowDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	result := make(map[string]*WorkflowDefinition)

	for id, workflow := range s.workflows {
		// Search in name
		if strings.Contains(strings.ToLower(workflow.Name), query) {
			result[id] = workflow
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(workflow.Description), query) {
			result[id] = workflow
			continue
		}

		// Search in tags
		for _, tag := range workflow.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				result[id] = workflow
				break
			}
		}
	}

	return result
}

// GetWorkflowMetadata returns metadata for all workflows
func (s *WorkflowStorage) GetWorkflowMetadata() []WorkflowMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata := make([]WorkflowMetadata, 0, len(s.workflows))
	for id, workflow := range s.workflows {
		metadata = append(metadata, WorkflowMetadata{
			ID:           id,
			Name:         workflow.Name,
			Description:  workflow.Description,
			Version:      workflow.Version,
			Author:       workflow.Author,
			Tags:         workflow.Tags,
			TaskCount:    len(workflow.Tasks),
			TriggerCount: len(workflow.Triggers),
		})
	}

	// Sort by name
	sort.Slice(metadata, func(i, j int) bool {
		return metadata[i].Name < metadata[j].Name
	})

	return metadata
}

// WorkflowMetadata represents workflow metadata
type WorkflowMetadata struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Version      string   `json:"version"`
	Author       string   `json:"author"`
	Tags         []string `json:"tags"`
	TaskCount    int      `json:"taskCount"`
	TriggerCount int      `json:"triggerCount"`
}

// ReloadIfNeeded reloads workflows if auto-reload is enabled and files have changed
func (s *WorkflowStorage) ReloadIfNeeded() error {
	if !s.autoReload {
		return nil
	}

	// Check if any workflow files have been modified since last reload
	needsReload := false
	for _, dir := range s.workflowDirs {
		if modTime, err := s.getDirectoryModTime(dir); err == nil {
			if modTime.After(s.lastReloadTime) {
				needsReload = true
				break
			}
		}
	}

	if needsReload {
		return s.LoadWorkflows()
	}

	return nil
}

// isWritableDir checks if a directory is writable
func (s *WorkflowStorage) isWritableDir(dir string) bool {
	// Try to create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false
	}

	// Test write permission
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return false
	}
	os.Remove(testFile)

	return true
}

// saveWorkflowToFile saves a workflow definition to a YAML file
func (s *WorkflowStorage) saveWorkflowToFile(filename string, workflow *WorkflowDefinition) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal workflow to YAML
	data, err := yaml.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}

// getDirectoryModTime gets the most recent modification time in a directory
func (s *WorkflowStorage) getDirectoryModTime(dir string) (time.Time, error) {
	var latestTime time.Time

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".yaml") || strings.HasSuffix(strings.ToLower(path), ".yml")) {
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
			}
		}

		return nil
	})

	return latestTime, err
}
