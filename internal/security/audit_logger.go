package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"nix-ai-help/pkg/logger"
)

// AuditLogger provides comprehensive audit logging for command execution
type AuditLogger struct {
	logFile    *os.File
	logPath    string
	logger     *logger.Logger
	mutex      sync.Mutex
	enabled    bool
	maxLogSize int64
	maxBackups int
}

// AuditEvent represents an audit event
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"eventType"`
	UserID      int                    `json:"userId"`
	Command     string                 `json:"command"`
	Args        []string               `json:"args"`
	Category    string                 `json:"category"`
	WorkingDir  string                 `json:"workingDir,omitempty"`
	Environment map[string]string      `json:"environment,omitempty"`
	RequiresSudo bool                  `json:"requiresSudo"`
	Success     bool                   `json:"success"`
	ExitCode    int                    `json:"exitCode,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Description string                 `json:"description,omitempty"`
	SessionID   string                 `json:"sessionId,omitempty"`
	ClientIP    string                 `json:"clientIp,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditEventType constants for different types of audit events
const (
	EventCommandAttempt  = "command_attempt"
	EventCommandSuccess  = "command_success"
	EventCommandFailed   = "command_failed"
	EventCommandRejected = "command_rejected"
	EventCommandDenied   = "command_denied"
	EventSudoRequest     = "sudo_request"
	EventSudoGranted     = "sudo_granted"
	EventSudoDenied      = "sudo_denied"
	EventSessionStart    = "session_start"
	EventSessionEnd      = "session_end"
	EventPolicyViolation = "policy_violation"
	EventSecurityAlert   = "security_alert"
)

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string, enabled bool, logger *logger.Logger) (*AuditLogger, error) {
	if !enabled {
		return &AuditLogger{
			enabled: false,
			logger:  logger,
		}, nil
	}
	
	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %w", err)
	}
	
	// Open log file for appending
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	
	auditLogger := &AuditLogger{
		logFile:    logFile,
		logPath:    logPath,
		logger:     logger,
		enabled:    enabled,
		maxLogSize: 100 * 1024 * 1024, // 100MB default
		maxBackups: 10,
	}
	
	// Log audit system startup
	auditLogger.LogSystemEvent("audit_system_start", "Audit logging system started", nil)
	
	return auditLogger, nil
}

// LogCommandAttempt logs a command execution attempt
func (al *AuditLogger) LogCommandAttempt(req interface{}) {
	if !al.enabled {
		return
	}
	
	// This is a placeholder - would need to extract actual request data
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventCommandAttempt,
		UserID:    os.Getuid(),
		// Command, Args, etc. would be extracted from req
	}
	
	al.writeEvent(event)
}

// LogCommandSuccess logs successful command execution
func (al *AuditLogger) LogCommandSuccess(req interface{}, result interface{}) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventCommandSuccess,
		UserID:    os.Getuid(),
		Success:   true,
		// Other fields would be populated from req and result
	}
	
	al.writeEvent(event)
}

// LogCommandFailed logs failed command execution
func (al *AuditLogger) LogCommandFailed(req interface{}, errorMsg string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventCommandFailed,
		UserID:    os.Getuid(),
		Success:   false,
		Error:     errorMsg,
	}
	
	al.writeEvent(event)
}

// LogCommandRejected logs rejected command attempts
func (al *AuditLogger) LogCommandRejected(req interface{}, reason string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventCommandRejected,
		UserID:    os.Getuid(),
		Success:   false,
		Error:     reason,
	}
	
	al.writeEvent(event)
}

// LogCommandDenied logs user-denied command execution
func (al *AuditLogger) LogCommandDenied(req interface{}, reason string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventCommandDenied,
		UserID:    os.Getuid(),
		Success:   false,
		Error:     reason,
	}
	
	al.writeEvent(event)
}

// LogSudoRequest logs sudo privilege requests
func (al *AuditLogger) LogSudoRequest(command string, args []string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventSudoRequest,
		UserID:    os.Getuid(),
		Command:   command,
		Args:      args,
	}
	
	al.writeEvent(event)
}

// LogSudoGranted logs successful sudo privilege elevation
func (al *AuditLogger) LogSudoGranted(command string, sessionID string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventSudoGranted,
		UserID:    os.Getuid(),
		Command:   command,
		Success:   true,
		SessionID: sessionID,
	}
	
	al.writeEvent(event)
}

// LogSudoDenied logs failed sudo privilege elevation
func (al *AuditLogger) LogSudoDenied(command string, reason string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventSudoDenied,
		UserID:    os.Getuid(),
		Command:   command,
		Success:   false,
		Error:     reason,
	}
	
	al.writeEvent(event)
}

// LogSessionStart logs the start of an elevated session
func (al *AuditLogger) LogSessionStart(sessionID string, duration time.Duration) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventSessionStart,
		UserID:    os.Getuid(),
		SessionID: sessionID,
		Duration:  duration,
	}
	
	al.writeEvent(event)
}

// LogSessionEnd logs the end of an elevated session
func (al *AuditLogger) LogSessionEnd(sessionID string, reason string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventSessionEnd,
		UserID:    os.Getuid(),
		SessionID: sessionID,
		Description: reason,
	}
	
	al.writeEvent(event)
}

// LogPolicyViolation logs policy violations
func (al *AuditLogger) LogPolicyViolation(command string, args []string, violation string) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp:   time.Now(),
		EventType:   EventPolicyViolation,
		UserID:      os.Getuid(),
		Command:     command,
		Args:        args,
		Success:     false,
		Description: violation,
	}
	
	al.writeEvent(event)
}

// LogSecurityAlert logs security-related alerts
func (al *AuditLogger) LogSecurityAlert(alertType string, details map[string]interface{}) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp: time.Now(),
		EventType: EventSecurityAlert,
		UserID:    os.Getuid(),
		Description: alertType,
		Metadata: details,
	}
	
	al.writeEvent(event)
}

// LogSystemEvent logs general system events
func (al *AuditLogger) LogSystemEvent(eventType string, description string, metadata map[string]interface{}) {
	if !al.enabled {
		return
	}
	
	event := AuditEvent{
		Timestamp:   time.Now(),
		EventType:   eventType,
		UserID:      os.Getuid(),
		Description: description,
		Metadata:    metadata,
	}
	
	al.writeEvent(event)
}

// writeEvent writes an audit event to the log file
func (al *AuditLogger) writeEvent(event AuditEvent) {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	
	// Check if log rotation is needed
	if err := al.checkLogRotation(); err != nil {
		al.logger.Error("Failed to rotate audit log")
	}
	
	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		al.logger.Error("Failed to serialize audit event")
		return
	}
	
	// Write to log file
	if al.logFile != nil {
		if _, err := al.logFile.WriteString(string(eventJSON) + "\n"); err != nil {
			al.logger.Error("Failed to write audit event")
		} else {
			// Ensure data is written to disk
			al.logFile.Sync()
		}
	}
	
	// Also log to application logger for important events
	switch event.EventType {
	case EventCommandRejected, EventCommandDenied, EventPolicyViolation, EventSecurityAlert:
		al.logger.Warn("AUDIT: " + event.EventType)
	case EventSudoGranted, EventSudoDenied:
		al.logger.Info("AUDIT: " + event.EventType)
	}
}

// checkLogRotation checks if log rotation is needed and performs it
func (al *AuditLogger) checkLogRotation() error {
	if al.logFile == nil {
		return nil
	}
	
	// Get current log file size
	fileInfo, err := al.logFile.Stat()
	if err != nil {
		return err
	}
	
	// Check if rotation is needed
	if fileInfo.Size() < al.maxLogSize {
		return nil
	}
	
	// Close current log file
	al.logFile.Close()
	
	// Rotate log files
	if err := al.rotateLogFiles(); err != nil {
		return err
	}
	
	// Reopen log file
	logFile, err := os.OpenFile(al.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	
	al.logFile = logFile
	return nil
}

// rotateLogFiles rotates the audit log files
func (al *AuditLogger) rotateLogFiles() error {
	// Remove oldest backup if max backups exceeded
	oldestBackup := fmt.Sprintf("%s.%d", al.logPath, al.maxBackups)
	if _, err := os.Stat(oldestBackup); err == nil {
		if err := os.Remove(oldestBackup); err != nil {
			return err
		}
	}
	
	// Rotate existing backups
	for i := al.maxBackups - 1; i >= 1; i-- {
		oldName := fmt.Sprintf("%s.%d", al.logPath, i)
		newName := fmt.Sprintf("%s.%d", al.logPath, i+1)
		
		if _, err := os.Stat(oldName); err == nil {
			if err := os.Rename(oldName, newName); err != nil {
				return err
			}
		}
	}
	
	// Move current log to .1
	backupName := fmt.Sprintf("%s.1", al.logPath)
	if err := os.Rename(al.logPath, backupName); err != nil {
		return err
	}
	
	return nil
}

// Close closes the audit logger and its resources
func (al *AuditLogger) Close() error {
	if !al.enabled || al.logFile == nil {
		return nil
	}
	
	// Log audit system shutdown
	al.LogSystemEvent("audit_system_stop", "Audit logging system stopped", nil)
	
	al.mutex.Lock()
	defer al.mutex.Unlock()
	
	return al.logFile.Close()
}

// GetAuditEvents retrieves audit events from the log file
func (al *AuditLogger) GetAuditEvents(since time.Time, eventTypes []string, limit int) ([]AuditEvent, error) {
	if !al.enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}
	
	// This is a simplified implementation
	// In a production system, you might want to use a database for better querying
	events := make([]AuditEvent, 0)
	
	// Read log file
	file, err := os.Open(al.logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// Parse events (simplified implementation)
	// In practice, you'd want more efficient parsing for large log files
	
	return events, nil
}

// GetAuditStatistics returns statistics about audit events
func (al *AuditLogger) GetAuditStatistics(since time.Time) (map[string]int, error) {
	if !al.enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}
	
	stats := make(map[string]int)
	
	// This would parse the log file and count events by type
	// Simplified implementation
	stats["total_commands"] = 0
	stats["successful_commands"] = 0
	stats["failed_commands"] = 0
	stats["rejected_commands"] = 0
	stats["sudo_requests"] = 0
	
	return stats, nil
}