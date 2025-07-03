package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
	"nix-ai-help/internal/types"
	"nix-ai-help/pkg/logger"
)


// SudoManager manages elevated privilege execution
type SudoManager struct {
	sessionManager  *SessionManager
	passwordManager *PasswordManager
	auditLogger     *AuditLogger
	logger          *logger.Logger
	strategy        ElevationStrategy
	config          *SudoConfig
}

// SudoConfig contains configuration for sudo management
type SudoConfig struct {
	SessionTimeout    time.Duration
	PasswordTimeout   time.Duration
	MaxAttempts       int
	RequirePassword   bool
	AllowPasswordless bool
	PreserveEnv       []string
}

// ElevationStrategy defines different methods for privilege elevation
type ElevationStrategy interface {
	Elevate(ctx context.Context, req types.CommandRequest) (*types.ExecutionResult, error)
	RequiresPassword() bool
	ValidateAccess() error
	Name() string
}

// PasswordManager handles secure password management
type PasswordManager struct {
	cache       map[string]*CachedPassword
	cacheExpiry time.Duration
	maxAttempts int
	mutex       sync.RWMutex
	logger      *logger.Logger
}

// CachedPassword represents a cached password with metadata
type CachedPassword struct {
	hashedPassword string
	timestamp      time.Time
	attempts       int
	userID         int
}

// SessionManager manages elevated privilege sessions
type SessionManager struct {
	activeSessions map[string]*ElevatedSession
	maxDuration    time.Duration
	mutex          sync.RWMutex
	logger         *logger.Logger
}

// ElevatedSession represents an active elevated privilege session
type ElevatedSession struct {
	SessionID   string
	StartTime   time.Time
	LastUsed    time.Time
	Commands    []string
	UserID      int
	MaxDuration time.Duration
	Context     map[string]interface{}
}

// NewSudoManager creates a new sudo manager
func NewSudoManager(config *SudoConfig, auditLogger *AuditLogger, logger *logger.Logger) *SudoManager {
	passwordManager := &PasswordManager{
		cache:       make(map[string]*CachedPassword),
		cacheExpiry: config.PasswordTimeout,
		maxAttempts: config.MaxAttempts,
		logger:      logger,
	}
	
	sessionManager := &SessionManager{
		activeSessions: make(map[string]*ElevatedSession),
		maxDuration:    config.SessionTimeout,
		logger:         logger,
	}
	
	sudoManager := &SudoManager{
		sessionManager:  sessionManager,
		passwordManager: passwordManager,
		auditLogger:     auditLogger,
		logger:          logger,
		config:          config,
	}
	
	// Determine elevation strategy
	sudoManager.strategy = sudoManager.selectElevationStrategy()
	
	// Start session cleanup routine
	go sudoManager.sessionCleanupRoutine()
	
	return sudoManager
}

// ExecuteWithSudo executes a command with elevated privileges
func (sm *SudoManager) ExecuteWithSudo(ctx context.Context, req types.CommandRequest) (*types.ExecutionResult, error) {
	// Log sudo request
	sm.auditLogger.LogSudoRequest(req.Command, req.Args)
	
	// Validate that elevation is needed and allowed
	if err := sm.validateSudoRequest(req); err != nil {
		sm.auditLogger.LogSudoDenied("command", err.Error())
		return nil, fmt.Errorf("sudo request denied: %w", err)
	}
	
	// Get or create elevated session
	session, err := sm.getOrCreateSession()
	if err != nil {
		sm.auditLogger.LogSudoDenied("command", err.Error())
		return nil, fmt.Errorf("failed to create elevated session: %w", err)
	}
	
	// Execute using the configured strategy
	result, err := sm.strategy.Elevate(ctx, req)
	
	// Update session
	session.LastUsed = time.Now()
	session.Commands = append(session.Commands, fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")))
	
	if err != nil {
		sm.auditLogger.LogCommandFailed(req, err.Error())
		return nil, err
	}
	
	sm.auditLogger.LogSudoGranted(fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")), session.SessionID)
	return result, nil
}

// HasActiveSession checks if there's an active elevated session
func (sm *SudoManager) HasActiveSession() bool {
	sm.sessionManager.mutex.RLock()
	defer sm.sessionManager.mutex.RUnlock()
	
	userID := os.Getuid()
	for _, session := range sm.sessionManager.activeSessions {
		if session.UserID == userID && sm.isSessionValid(session) {
			return true
		}
	}
	
	return false
}

// InvalidateSession invalidates an active session
func (sm *SudoManager) InvalidateSession(sessionID string) error {
	sm.sessionManager.mutex.Lock()
	defer sm.sessionManager.mutex.Unlock()
	
	if _, exists := sm.sessionManager.activeSessions[sessionID]; exists {
		sm.auditLogger.LogSessionEnd(sessionID, "manual_invalidation")
		delete(sm.sessionManager.activeSessions, sessionID)
		sm.logger.Info("Invalidated elevated session")
		return nil
	}
	
	return fmt.Errorf("session not found: %s", sessionID)
}

// ClearPasswordCache clears the password cache
func (sm *SudoManager) ClearPasswordCache() {
	sm.passwordManager.mutex.Lock()
	defer sm.passwordManager.mutex.Unlock()
	
	sm.passwordManager.cache = make(map[string]*CachedPassword)
	sm.logger.Info("Password cache cleared")
}

// GetSessionInfo returns information about active sessions
func (sm *SudoManager) GetSessionInfo() []SessionInfo {
	sm.sessionManager.mutex.RLock()
	defer sm.sessionManager.mutex.RUnlock()
	
	userID := os.Getuid()
	sessions := make([]SessionInfo, 0)
	
	for _, session := range sm.sessionManager.activeSessions {
		if session.UserID == userID {
			sessions = append(sessions, SessionInfo{
				SessionID:    session.SessionID,
				StartTime:    session.StartTime,
				LastUsed:     session.LastUsed,
				CommandCount: len(session.Commands),
				Valid:        sm.isSessionValid(session),
			})
		}
	}
	
	return sessions
}

// SessionInfo contains information about an elevated session
type SessionInfo struct {
	SessionID    string
	StartTime    time.Time
	LastUsed     time.Time
	CommandCount int
	Valid        bool
}

// selectElevationStrategy selects the appropriate elevation strategy
func (sm *SudoManager) selectElevationStrategy() ElevationStrategy {
	// Check if passwordless sudo is available
	if sm.config.AllowPasswordless && sm.checkPasswordlessSudo() {
		sm.logger.Info("Using passwordless sudo strategy")
		return &PasswordlessSudoStrategy{
			logger: sm.logger,
		}
	}
	
	// Check if PolicyKit is available (for GUI environments)
	if sm.checkPolicyKitAvailable() {
		sm.logger.Info("Using PolicyKit elevation strategy")
		return &PolicyKitStrategy{
			logger: sm.logger,
		}
	}
	
	// Default to password-based sudo
	sm.logger.Info("Using password-based sudo strategy")
	return &PasswordSudoStrategy{
		passwordManager: sm.passwordManager,
		logger:          sm.logger,
	}
}

// validateSudoRequest validates that a sudo request is allowed
func (sm *SudoManager) validateSudoRequest(req types.CommandRequest) error {
	// Check if user is in sudoers
	if !sm.checkSudoAccess() {
		return fmt.Errorf("user does not have sudo access")
	}
	
	// Additional validation logic would go here
	return nil
}

// getOrCreateSession gets an existing session or creates a new one
func (sm *SudoManager) getOrCreateSession() (*ElevatedSession, error) {
	sm.sessionManager.mutex.Lock()
	defer sm.sessionManager.mutex.Unlock()
	
	userID := os.Getuid()
	
	// Look for existing valid session
	for _, session := range sm.sessionManager.activeSessions {
		if session.UserID == userID && sm.isSessionValid(session) {
			return session, nil
		}
	}
	
	// Create new session
	sessionID := generateSessionID()
	session := &ElevatedSession{
		SessionID:   sessionID,
		StartTime:   time.Now(),
		LastUsed:    time.Now(),
		Commands:    make([]string, 0),
		UserID:      userID,
		MaxDuration: sm.config.SessionTimeout,
		Context:     make(map[string]interface{}),
	}
	
	sm.sessionManager.activeSessions[sessionID] = session
	sm.auditLogger.LogSessionStart(sessionID, sm.config.SessionTimeout)
	sm.logger.Info("Created new elevated session")
	
	return session, nil
}

// isSessionValid checks if a session is still valid
func (sm *SudoManager) isSessionValid(session *ElevatedSession) bool {
	now := time.Now()
	
	// Check if session has expired
	if now.Sub(session.StartTime) > session.MaxDuration {
		return false
	}
	
	// Check if session has been idle too long
	idleTimeout := 15 * time.Minute // Configurable idle timeout
	if now.Sub(session.LastUsed) > idleTimeout {
		return false
	}
	
	return true
}

// sessionCleanupRoutine periodically cleans up expired sessions
func (sm *SudoManager) sessionCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.cleanupExpiredSessions()
		sm.cleanupExpiredPasswords()
	}
}

// cleanupExpiredSessions removes expired sessions
func (sm *SudoManager) cleanupExpiredSessions() {
	sm.sessionManager.mutex.Lock()
	defer sm.sessionManager.mutex.Unlock()
	
	for sessionID, session := range sm.sessionManager.activeSessions {
		if !sm.isSessionValid(session) {
			sm.auditLogger.LogSessionEnd(sessionID, "expired")
			delete(sm.sessionManager.activeSessions, sessionID)
			sm.logger.Debug("Cleaned up expired session")
		}
	}
}

// cleanupExpiredPasswords removes expired passwords from cache
func (sm *SudoManager) cleanupExpiredPasswords() {
	sm.passwordManager.mutex.Lock()
	defer sm.passwordManager.mutex.Unlock()
	
	now := time.Now()
	for userKey, cached := range sm.passwordManager.cache {
		if now.Sub(cached.timestamp) > sm.passwordManager.cacheExpiry {
			delete(sm.passwordManager.cache, userKey)
			sm.logger.Debug("Cleaned up expired password cache for user")
		}
	}
}

// checkPasswordlessSudo checks if passwordless sudo is available
func (sm *SudoManager) checkPasswordlessSudo() bool {
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	return err == nil
}

// checkPolicyKitAvailable checks if PolicyKit is available
func (sm *SudoManager) checkPolicyKitAvailable() bool {
	// Check if pkexec is available
	_, err := exec.LookPath("pkexec")
	return err == nil
}

// checkSudoAccess checks if the user has sudo access
func (sm *SudoManager) checkSudoAccess() bool {
	cmd := exec.Command("sudo", "-v")
	err := cmd.Run()
	return err == nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// RequestPassword securely requests a password from the user
func (pm *PasswordManager) RequestPassword(ctx context.Context) (string, error) {
	userKey := fmt.Sprintf("user_%d", os.Getuid())
	
	// Check cache first
	pm.mutex.RLock()
	if cached, exists := pm.cache[userKey]; exists {
		if time.Since(cached.timestamp) < pm.cacheExpiry {
			pm.mutex.RUnlock()
			return cached.hashedPassword, nil
		}
	}
	pm.mutex.RUnlock()
	
	// Request password from user
	fmt.Print("🔐 Enter sudo password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // New line after password input
	
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	
	passwordStr := string(password)
	
	// Hash password for caching
	hash := sha256.Sum256([]byte(passwordStr))
	hashedPassword := hex.EncodeToString(hash[:])
	
	// Cache the hashed password
	pm.mutex.Lock()
	pm.cache[userKey] = &CachedPassword{
		hashedPassword: hashedPassword,
		timestamp:      time.Now(),
		attempts:       0,
		userID:         os.Getuid(),
	}
	pm.mutex.Unlock()
	
	return passwordStr, nil
}

// Strategy Implementations

// PasswordSudoStrategy implements password-based sudo elevation
type PasswordSudoStrategy struct {
	passwordManager *PasswordManager
	logger          *logger.Logger
}

func (pss *PasswordSudoStrategy) Elevate(ctx context.Context, req types.CommandRequest) (*types.ExecutionResult, error) {
	startTime := time.Now()
	
	// Get password
	password, err := pss.passwordManager.RequestPassword(ctx)
	if err != nil {
		return nil, err
	}
	
	// Execute with sudo
	cmd := exec.CommandContext(ctx, "sudo", append([]string{"-S", req.Command}, req.Args...)...)
	cmd.Stdin = strings.NewReader(password + "\n")
	
	// Set working directory if specified
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	
	// Set environment variables
	for key, value := range req.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	result := &types.ExecutionResult{
		Command:   fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")),
		Output:    string(output),
		Duration:  duration,
		Timestamp: startTime,
		DryRun:    req.DryRun,
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}
	
	return result, err
}

func (pss *PasswordSudoStrategy) RequiresPassword() bool { return true }
func (pss *PasswordSudoStrategy) ValidateAccess() error { return nil }
func (pss *PasswordSudoStrategy) Name() string          { return "password_sudo" }

// PasswordlessSudoStrategy implements passwordless sudo elevation
type PasswordlessSudoStrategy struct {
	logger *logger.Logger
}

func (pss *PasswordlessSudoStrategy) Elevate(ctx context.Context, req types.CommandRequest) (*types.ExecutionResult, error) {
	startTime := time.Now()
	
	// Execute with sudo (no password needed)
	cmd := exec.CommandContext(ctx, "sudo", append([]string{req.Command}, req.Args...)...)
	
	// Set working directory if specified
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	
	// Set environment variables
	for key, value := range req.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	result := &types.ExecutionResult{
		Command:   fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")),
		Output:    string(output),
		Duration:  duration,
		Timestamp: startTime,
		DryRun:    req.DryRun,
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}
	
	return result, err
}

func (pss *PasswordlessSudoStrategy) RequiresPassword() bool { return false }
func (pss *PasswordlessSudoStrategy) ValidateAccess() error { return nil }
func (pss *PasswordlessSudoStrategy) Name() string          { return "passwordless_sudo" }

// PolicyKitStrategy implements PolicyKit elevation
type PolicyKitStrategy struct {
	logger *logger.Logger
}

func (pks *PolicyKitStrategy) Elevate(ctx context.Context, req types.CommandRequest) (*types.ExecutionResult, error) {
	startTime := time.Now()
	
	// Execute with pkexec
	cmd := exec.CommandContext(ctx, "pkexec", append([]string{req.Command}, req.Args...)...)
	
	// Set working directory if specified
	if req.WorkingDir != "" {
		cmd.Dir = req.WorkingDir
	}
	
	// Set environment variables
	for key, value := range req.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	result := &types.ExecutionResult{
		Command:   fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")),
		Output:    string(output),
		Duration:  duration,
		Timestamp: startTime,
		DryRun:    req.DryRun,
	}
	
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			}
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}
	
	return result, err
}

func (pks *PolicyKitStrategy) RequiresPassword() bool { return true }
func (pks *PolicyKitStrategy) ValidateAccess() error { return nil }
func (pks *PolicyKitStrategy) Name() string          { return "policykit" }