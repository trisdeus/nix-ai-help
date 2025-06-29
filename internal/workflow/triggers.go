package workflow

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TriggerManager manages workflow triggers
type TriggerManager struct {
	mu          sync.RWMutex
	triggers    map[string]*TriggerInfo
	scheduler   *SimpleScheduler
	fileWatcher *FileWatcher
	logger      Logger
	engine      *Engine
	ctx         context.Context
	cancel      context.CancelFunc
}

// TriggerInfo holds information about a registered trigger
type TriggerInfo struct {
	ID         string
	WorkflowID string
	Trigger    *Trigger
	Active     bool
	LastRun    *time.Time
	NextRun    *time.Time
}

// SimpleScheduler is a basic scheduler implementation
type SimpleScheduler struct {
	jobs map[string]*ScheduledJob
	mu   sync.RWMutex
	stop chan struct{}
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID       string
	Interval time.Duration
	NextRun  time.Time
	Callback func()
}

// FileWatcher handles file system events for file-based triggers
type FileWatcher struct {
	mu       sync.RWMutex
	watchers map[string]*FileWatchInfo
	logger   Logger
	stop     chan struct{}
}

// FileWatchInfo holds file watching information
type FileWatchInfo struct {
	Path     string
	Events   []string
	Callback func(event FileEvent)
	LastMod  time.Time
}

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	EventType string // created, modified, deleted
	Time      time.Time
}

// NewTriggerManager creates a new trigger manager
func NewTriggerManager(logger Logger, engine *Engine) *TriggerManager {
	ctx, cancel := context.WithCancel(context.Background())

	tm := &TriggerManager{
		triggers:    make(map[string]*TriggerInfo),
		scheduler:   NewSimpleScheduler(),
		fileWatcher: NewFileWatcher(logger),
		logger:      logger,
		engine:      engine,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start the scheduler
	tm.scheduler.Start()

	return tm
}

// NewSimpleScheduler creates a new simple scheduler
func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{
		jobs: make(map[string]*ScheduledJob),
		stop: make(chan struct{}),
	}
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(logger Logger) *FileWatcher {
	return &FileWatcher{
		watchers: make(map[string]*FileWatchInfo),
		logger:   logger,
		stop:     make(chan struct{}),
	}
}

// RegisterTrigger registers a trigger for a workflow
func (tm *TriggerManager) RegisterTrigger(workflowID string, trigger *Trigger) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	triggerID := fmt.Sprintf("%s-%s", workflowID, trigger.ID)

	triggerInfo := &TriggerInfo{
		ID:         triggerID,
		WorkflowID: workflowID,
		Trigger:    trigger,
		Active:     true,
	}

	switch trigger.Type {
	case "schedule":
		return tm.registerScheduleTrigger(triggerInfo)
	case "file-change":
		return tm.registerFileChangeTrigger(triggerInfo)
	case "manual":
		// Manual triggers don't need registration
		tm.triggers[triggerID] = triggerInfo
		return nil
	default:
		return fmt.Errorf("unsupported trigger type: %s", trigger.Type)
	}
}

// registerScheduleTrigger registers a schedule-based trigger
func (tm *TriggerManager) registerScheduleTrigger(triggerInfo *TriggerInfo) error {
	trigger := triggerInfo.Trigger

	// Get the interval from the trigger configuration
	intervalValue, exists := trigger.Config["interval"]
	if !exists {
		return fmt.Errorf("schedule trigger requires 'interval' configuration")
	}

	interval, err := tm.parseInterval(fmt.Sprintf("%v", intervalValue))
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	// Create job function
	job := &ScheduledJob{
		ID:       triggerInfo.ID,
		Interval: interval,
		NextRun:  time.Now().Add(interval),
		Callback: func() {
			tm.logger.Info("Executing scheduled workflow", "workflow", triggerInfo.WorkflowID, "trigger", triggerInfo.ID)

			// Update last run time
			now := time.Now()
			triggerInfo.LastRun = &now

			// Execute the workflow
			if _, err := tm.engine.ExecuteWorkflow(triggerInfo.WorkflowID, nil); err != nil {
				tm.logger.Error("Failed to execute scheduled workflow", "error", err, "workflow", triggerInfo.WorkflowID)
			}
		},
	}

	// Add job to scheduler
	tm.scheduler.AddJob(job)
	tm.triggers[triggerInfo.ID] = triggerInfo

	tm.logger.Info("Registered schedule trigger", "workflow", triggerInfo.WorkflowID, "interval", interval)
	return nil
}

// parseInterval parses interval strings like "1h", "30m", "daily", etc.
func (tm *TriggerManager) parseInterval(intervalStr string) (time.Duration, error) {
	switch intervalStr {
	case "daily":
		return 24 * time.Hour, nil
	case "hourly":
		return time.Hour, nil
	case "weekly":
		return 7 * 24 * time.Hour, nil
	case "monthly":
		return 30 * 24 * time.Hour, nil
	default:
		// Try to parse as duration
		return time.ParseDuration(intervalStr)
	}
}

// registerFileChangeTrigger registers a file change trigger
func (tm *TriggerManager) registerFileChangeTrigger(triggerInfo *TriggerInfo) error {
	trigger := triggerInfo.Trigger

	// Get the file path to watch
	pathValue, exists := trigger.Config["path"]
	if !exists {
		return fmt.Errorf("file change trigger requires 'path' configuration")
	}

	watchPath := fmt.Sprintf("%v", pathValue)

	// Get the events to watch for (default to "modified")
	events := []string{"modified"}
	if eventsValue, exists := trigger.Config["events"]; exists {
		if eventsList, ok := eventsValue.([]interface{}); ok {
			events = make([]string, len(eventsList))
			for i, event := range eventsList {
				events[i] = fmt.Sprintf("%v", event)
			}
		} else {
			events = []string{fmt.Sprintf("%v", eventsValue)}
		}
	}

	// Create callback function
	callback := func(event FileEvent) {
		tm.logger.Info("File change detected", "path", event.Path, "event", event.EventType, "workflow", triggerInfo.WorkflowID)

		// Update last run time
		now := time.Now()
		triggerInfo.LastRun = &now

		// Execute the workflow
		if _, err := tm.engine.ExecuteWorkflow(triggerInfo.WorkflowID, nil); err != nil {
			tm.logger.Error("Failed to execute workflow on file change", "error", err, "workflow", triggerInfo.WorkflowID)
		}
	}

	// Register with file watcher
	if err := tm.fileWatcher.WatchFile(watchPath, events, callback); err != nil {
		return fmt.Errorf("failed to register file watcher: %w", err)
	}

	tm.triggers[triggerInfo.ID] = triggerInfo

	tm.logger.Info("Registered file change trigger", "workflow", triggerInfo.WorkflowID, "path", watchPath, "events", events)
	return nil
}

// UnregisterTrigger unregisters a trigger
func (tm *TriggerManager) UnregisterTrigger(workflowID, triggerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	fullTriggerID := fmt.Sprintf("%s-%s", workflowID, triggerID)

	triggerInfo, exists := tm.triggers[fullTriggerID]
	if !exists {
		return fmt.Errorf("trigger not found: %s", fullTriggerID)
	}

	switch triggerInfo.Trigger.Type {
	case "schedule":
		tm.scheduler.RemoveJob(fullTriggerID)
	case "file-change":
		// Remove file watcher
		if pathValue, exists := triggerInfo.Trigger.Config["path"]; exists {
			watchPath := fmt.Sprintf("%v", pathValue)
			tm.fileWatcher.UnwatchFile(watchPath)
		}
	}

	delete(tm.triggers, fullTriggerID)

	tm.logger.Info("Unregistered trigger", "workflow", workflowID, "trigger", triggerID)
	return nil
}

// GetTriggers returns all registered triggers
func (tm *TriggerManager) GetTriggers() map[string]*TriggerInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	triggers := make(map[string]*TriggerInfo)
	for id, trigger := range tm.triggers {
		triggers[id] = trigger
	}

	return triggers
}

// GetWorkflowTriggers returns triggers for a specific workflow
func (tm *TriggerManager) GetWorkflowTriggers(workflowID string) []*TriggerInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var triggers []*TriggerInfo
	for _, trigger := range tm.triggers {
		if trigger.WorkflowID == workflowID {
			triggers = append(triggers, trigger)
		}
	}

	return triggers
}

// EnableTrigger enables a trigger
func (tm *TriggerManager) EnableTrigger(workflowID, triggerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	fullTriggerID := fmt.Sprintf("%s-%s", workflowID, triggerID)

	if triggerInfo, exists := tm.triggers[fullTriggerID]; exists {
		triggerInfo.Active = true
		tm.logger.Info("Enabled trigger", "workflow", workflowID, "trigger", triggerID)
		return nil
	}

	return fmt.Errorf("trigger not found: %s", fullTriggerID)
}

// DisableTrigger disables a trigger
func (tm *TriggerManager) DisableTrigger(workflowID, triggerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	fullTriggerID := fmt.Sprintf("%s-%s", workflowID, triggerID)

	if triggerInfo, exists := tm.triggers[fullTriggerID]; exists {
		triggerInfo.Active = false
		tm.logger.Info("Disabled trigger", "workflow", workflowID, "trigger", triggerID)
		return nil
	}

	return fmt.Errorf("trigger not found: %s", fullTriggerID)
}

// TriggerWorkflow manually triggers a workflow
func (tm *TriggerManager) TriggerWorkflow(workflowID string, variables map[string]interface{}) error {
	tm.logger.Info("Manually triggering workflow", "workflow", workflowID)

	// Convert variables from interface{} to string
	stringVars := make(map[string]string)
	for k, v := range variables {
		stringVars[k] = fmt.Sprintf("%v", v)
	}

	// Execute the workflow
	if _, err := tm.engine.ExecuteWorkflow(workflowID, stringVars); err != nil {
		return fmt.Errorf("failed to execute workflow: %w", err)
	}

	return nil
}

// Stop stops the trigger manager
func (tm *TriggerManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.logger.Info("Stopping trigger manager")

	// Stop the scheduler
	tm.scheduler.Stop()

	// Stop file watcher
	tm.fileWatcher.Stop()

	// Cancel context
	tm.cancel()
}

// Start starts the simple scheduler
func (s *SimpleScheduler) Start() {
	go s.run()
}

// Stop stops the simple scheduler
func (s *SimpleScheduler) Stop() {
	close(s.stop)
}

// AddJob adds a job to the scheduler
func (s *SimpleScheduler) AddJob(job *ScheduledJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

// RemoveJob removes a job from the scheduler
func (s *SimpleScheduler) RemoveJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, jobID)
}

// run runs the scheduler loop
func (s *SimpleScheduler) run() {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkJobs()
		case <-s.stop:
			return
		}
	}
}

// checkJobs checks for jobs that need to be executed
func (s *SimpleScheduler) checkJobs() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for _, job := range s.jobs {
		if now.After(job.NextRun) || now.Equal(job.NextRun) {
			// Execute job
			go job.Callback()
			// Schedule next run
			job.NextRun = now.Add(job.Interval)
		}
	}
}

// WatchFile registers a file watcher
func (fw *FileWatcher) WatchFile(path string, events []string, callback func(FileEvent)) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	// Check if file/directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	watchInfo := &FileWatchInfo{
		Path:     path,
		Events:   events,
		Callback: callback,
	}

	// Get initial modification time
	if info, err := os.Stat(path); err == nil {
		watchInfo.LastMod = info.ModTime()
	}

	fw.watchers[path] = watchInfo

	// Start watching (simple polling implementation)
	go fw.pollFile(watchInfo)

	fw.logger.Info("Started watching file", "path", path, "events", events)
	return nil
}

// UnwatchFile removes a file watcher
func (fw *FileWatcher) UnwatchFile(path string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	delete(fw.watchers, path)
	fw.logger.Info("Stopped watching file", "path", path)
}

// pollFile polls a file for changes (simple implementation)
func (fw *FileWatcher) pollFile(watchInfo *FileWatchInfo) {
	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fw.checkFileChanges(watchInfo)
		case <-fw.stop:
			return
		}
	}
}

// checkFileChanges checks for file changes
func (fw *FileWatcher) checkFileChanges(watchInfo *FileWatchInfo) {
	info, err := os.Stat(watchInfo.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// File was deleted
			if fw.shouldTriggerEvent(watchInfo.Events, "deleted") {
				watchInfo.Callback(FileEvent{
					Path:      watchInfo.Path,
					EventType: "deleted",
					Time:      time.Now(),
				})
			}
		}
		return
	}

	// Check if file was modified
	if info.ModTime().After(watchInfo.LastMod) {
		if fw.shouldTriggerEvent(watchInfo.Events, "modified") {
			watchInfo.Callback(FileEvent{
				Path:      watchInfo.Path,
				EventType: "modified",
				Time:      info.ModTime(),
			})
		}
		watchInfo.LastMod = info.ModTime()
	}

	// Check for new files if watching a directory
	if info.IsDir() {
		fw.checkDirectoryChanges(watchInfo)
	}
}

// checkDirectoryChanges checks for changes in a directory
func (fw *FileWatcher) checkDirectoryChanges(watchInfo *FileWatchInfo) {
	entries, err := os.ReadDir(watchInfo.Path)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(watchInfo.Path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if this is a new file
		if info.ModTime().After(watchInfo.LastMod) {
			if fw.shouldTriggerEvent(watchInfo.Events, "created") {
				watchInfo.Callback(FileEvent{
					Path:      fullPath,
					EventType: "created",
					Time:      info.ModTime(),
				})
			}
		}
	}
}

// shouldTriggerEvent checks if an event type should trigger the callback
func (fw *FileWatcher) shouldTriggerEvent(watchedEvents []string, eventType string) bool {
	for _, event := range watchedEvents {
		if event == "all" || event == eventType {
			return true
		}
	}
	return false
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	close(fw.stop)
	fw.watchers = make(map[string]*FileWatchInfo)
	fw.logger.Info("Stopped file watcher")
}
