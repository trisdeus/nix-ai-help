package tui

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/execution"
	execTypes "nix-ai-help/internal/execution"
	"nix-ai-help/pkg/logger"
)

// ExecutionState represents the state of command execution
type ExecutionState string

const (
	ExecutionIdle       ExecutionState = "idle"
	ExecutionDetecting  ExecutionState = "detecting"
	ExecutionPending    ExecutionState = "pending"
	ExecutionRunning    ExecutionState = "running"
	ExecutionCompleted  ExecutionState = "completed"
	ExecutionFailed     ExecutionState = "failed"
	ExecutionCancelled  ExecutionState = "cancelled"
)

// ExecutionRequest represents an execution request in the TUI
type ExecutionRequest struct {
	ID          string                  `json:"id"`
	UserQuery   string                  `json:"user_query"`
	Command     string                  `json:"command"`
	Args        []string                `json:"args"`
	Description string                  `json:"description"`
	Category    string                  `json:"category"`
	DryRun      bool                    `json:"dry_run"`
	State       ExecutionState          `json:"state"`
	CreatedAt   time.Time               `json:"created_at"`
	StartedAt   *time.Time              `json:"started_at,omitempty"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	Output      string                  `json:"output,omitempty"`
	Error       string                  `json:"error,omitempty"`
	ExitCode    int                     `json:"exit_code"`
	Duration    time.Duration           `json:"duration"`
}

// ExecutionManager manages command execution in the TUI
type ExecutionManager struct {
	requests        []ExecutionRequest
	currentID       int
	providerMgr     *ai.ProviderManager
	commandResolver *execution.CommandResolver
	logger          *logger.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewExecutionManager creates a new execution manager for the TUI
func NewExecutionManager(providerMgr *ai.ProviderManager, log *logger.Logger) *ExecutionManager {
	ctx, cancel := context.WithCancel(context.Background())
	commandResolver := execution.NewCommandResolver(log)
	
	return &ExecutionManager{
		requests:        make([]ExecutionRequest, 0),
		currentID:       0,
		providerMgr:     providerMgr,
		commandResolver: commandResolver,
		logger:          log,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// ExecutionDetectedMsg indicates that an execution request was detected
type ExecutionDetectedMsg struct {
	Request ExecutionRequest
}

// ExecutionStartedMsg indicates that execution has started
type ExecutionStartedMsg struct {
	ID string
}

// ExecutionCompletedMsg indicates that execution has completed
type ExecutionCompletedMsg struct {
	ID     string
	Result *execTypes.ExecutionResult
	Error  error
}

// ExecutionOutputMsg provides streaming output from execution
type ExecutionOutputMsg struct {
	ID     string
	Output string
	IsError bool
}

// ExecutionCancelledMsg indicates that execution was cancelled
type ExecutionCancelledMsg struct {
	ID string
}

// DetectExecutionRequest analyzes user input for execution requests
func (em *ExecutionManager) DetectExecutionRequest(userInput string) (*ExecutionRequest, error) {
	// First try using AI provider if available
	provider, err := em.providerMgr.GetDefaultProvider()
	if err == nil {
		// Check if the provider supports execution detection
		if eap, ok := provider.(*ai.ExecutionAwareProvider); ok && eap.IsExecutionEnabled() {
			// Use the execution-aware provider to detect commands
			execReq := eap.DetectExecutionRequest(userInput)
			if execReq != nil {
				// Convert to TUI execution request
				tuiReq := &ExecutionRequest{
					ID:          em.generateID(),
					UserQuery:   userInput,
					Command:     execReq.Command,
					Args:        execReq.Args,
					Description: execReq.Description,
					Category:    execReq.Category,
					DryRun:      execReq.DryRun,
					State:       ExecutionDetecting,
					CreatedAt:   time.Now(),
				}
				
				em.requests = append(em.requests, *tuiReq)
				return tuiReq, nil
			}
		}
	} else {
		// Provider not available, use fallback pattern matching
		em.logger.Warn(fmt.Sprintf("Provider not available for execution detection, using fallback: %v", err))
	}

	// Fallback: Use simple pattern matching for execution detection
	return em.fallbackExecutionDetection(userInput)
}

// fallbackExecutionDetection uses pattern matching when AI provider is unavailable
func (em *ExecutionManager) fallbackExecutionDetection(userInput string) (*ExecutionRequest, error) {
	
	// Define execution patterns for fallback detection
	executionPatterns := []struct {
		pattern     *regexp.Regexp
		category    string
		description string
	}{
		{regexp.MustCompile(`(?i)\b(install|add)\s+([a-zA-Z0-9_-]+)`), "package", "Install package"},
		{regexp.MustCompile(`(?i)\b(remove|uninstall|delete)\s+([a-zA-Z0-9_-]+)`), "package", "Remove package"},
		{regexp.MustCompile(`(?i)\b(update|upgrade)\s+(system|packages?)`), "system", "Update system"},
		{regexp.MustCompile(`(?i)\bnixos-rebuild\s+(switch|boot|test)`), "nixos", "NixOS rebuild"},
		{regexp.MustCompile(`(?i)\b(start|stop|restart|enable|disable)\s+([a-zA-Z0-9_-]+)`), "service", "Service management"},
		{regexp.MustCompile(`(?i)\bsystemctl\s+(start|stop|restart|enable|disable|status)\s+([a-zA-Z0-9_.-]+)`), "systemctl", "Systemctl command"},
		{regexp.MustCompile(`(?i)\bnix-env\s+-[iueq]`), "nix", "Nix environment command"},
		{regexp.MustCompile(`(?i)\bnix-collect-garbage`), "nix", "Garbage collection"},
		{regexp.MustCompile(`(?i)\b(run|execute)\s+([a-zA-Z0-9_-]+)`), "command", "Run command"},
		{regexp.MustCompile(`(?i)\b(can you|please)\s+(install|run|execute|start|stop)`), "request", "Polite command request"},
	}
	
	// Check each pattern
	for _, ep := range executionPatterns {
		if matches := ep.pattern.FindStringSubmatch(userInput); matches != nil {
			// Extract command and arguments
			command, args := em.parseCommandFromMatch(userInput, matches, ep.category)
			
			if command != "" {
				tuiReq := &ExecutionRequest{
					ID:          em.generateID(),
					UserQuery:   userInput,
					Command:     command,
					Args:        args,
					Description: ep.description,
					Category:    ep.category,
					DryRun:      true, // Default to dry run for safety
					State:       ExecutionDetecting,
					CreatedAt:   time.Now(),
				}
				
				em.requests = append(em.requests, *tuiReq)
				em.logger.Info(fmt.Sprintf("Fallback execution detection matched: %s", command))
				return tuiReq, nil
			}
		}
	}
	
	return nil, nil
}

// parseCommandFromMatch extracts command and arguments from regex matches
func (em *ExecutionManager) parseCommandFromMatch(userInput string, matches []string, category string) (string, []string) {
	userLower := strings.ToLower(strings.TrimSpace(userInput))
	
	switch category {
	case "package":
		if strings.Contains(userLower, "install") || strings.Contains(userLower, "add") {
			if len(matches) >= 3 {
				return "nix-env", []string{"-iA", "nixpkgs." + matches[2]}
			}
		} else if strings.Contains(userLower, "remove") || strings.Contains(userLower, "uninstall") {
			if len(matches) >= 3 {
				return "nix-env", []string{"-e", matches[2]}
			}
		}
		
	case "system":
		if strings.Contains(userLower, "update") || strings.Contains(userLower, "upgrade") {
			return "nixos-rebuild", []string{"switch", "--upgrade"}
		}
		
	case "nixos":
		if strings.Contains(userLower, "nixos-rebuild") {
			if strings.Contains(userLower, "switch") {
				return "nixos-rebuild", []string{"switch"}
			} else if strings.Contains(userLower, "boot") {
				return "nixos-rebuild", []string{"boot"}
			} else if strings.Contains(userLower, "test") {
				return "nixos-rebuild", []string{"test"}
			}
		}
		
	case "service", "systemctl":
		if strings.Contains(userLower, "systemctl") {
			// Extract systemctl command directly
			parts := strings.Fields(userInput)
			for i, part := range parts {
				if strings.ToLower(part) == "systemctl" && i+2 < len(parts) {
					return "systemctl", []string{parts[i+1], parts[i+2]}
				}
			}
		} else {
			// Natural language service commands
			if len(matches) >= 3 {
				action := matches[1]
				service := matches[2]
				return "systemctl", []string{action, service}
			}
		}
		
	case "nix":
		if strings.Contains(userLower, "nix-env") {
			// Extract nix-env command
			parts := strings.Fields(userInput)
			for i, part := range parts {
				if strings.Contains(strings.ToLower(part), "nix-env") && i+1 < len(parts) {
					return "nix-env", parts[i+1:]
				}
			}
		} else if strings.Contains(userLower, "nix-collect-garbage") {
			return "nix-collect-garbage", []string{"-d"}
		}
		
	case "command":
		if len(matches) >= 3 {
			return matches[2], []string{}
		}
		
	case "request":
		// Parse polite requests like "can you install firefox"
		words := strings.Fields(userLower)
		for i, word := range words {
			if word == "install" && i+1 < len(words) {
				return "nix-env", []string{"-iA", "nixpkgs." + words[i+1]}
			} else if (word == "start" || word == "stop" || word == "restart") && i+1 < len(words) {
				return "systemctl", []string{word, words[i+1]}
			}
		}
	}
	
	return "", nil
}

// RequestExecution initiates execution of a command
func (em *ExecutionManager) RequestExecution(req *ExecutionRequest) tea.Cmd {
	return func() tea.Msg {
		// Update request state
		req.State = ExecutionPending
		em.updateRequest(req)

		return ExecutionDetectedMsg{Request: *req}
	}
}

// ExecuteRequest executes a command request
func (em *ExecutionManager) ExecuteRequest(req *ExecutionRequest) tea.Cmd {
	return func() tea.Msg {
		// Update state to running
		now := time.Now()
		req.State = ExecutionRunning
		req.StartedAt = &now
		em.updateRequest(req)

		// Send started message
		return ExecutionStartedMsg{ID: req.ID}
	}
}

// StartExecution starts the actual execution process
func (em *ExecutionManager) StartExecution(req *ExecutionRequest) tea.Cmd {
	return tea.Batch(
		em.ExecuteRequest(req),
		em.runExecution(req),
	)
}

// runExecution performs the actual command execution
func (em *ExecutionManager) runExecution(req *ExecutionRequest) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		
		// For now, simulate execution since we need the full executor setup
		// In a real implementation, this would create and use the SafeExecutor:
		// cmdReq := execTypes.CommandRequest{
		//     Command:     req.Command,
		//     Args:        req.Args,
		//     Description: req.Description,
		//     Category:    req.Category,
		//     DryRun:      req.DryRun,
		// }
		// In a real implementation, this would use the SafeExecutor
		result := &execTypes.ExecutionResult{
			Success:  true,
			Command:  fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " ")),
			Output:   "Command executed successfully (simulated)",
			ExitCode: 0,
			Duration: time.Since(start),
			DryRun:   req.DryRun,
		}

		// Update request with results
		completed := time.Now()
		req.State = ExecutionCompleted
		req.CompletedAt = &completed
		req.Duration = result.Duration
		req.Output = result.Output
		req.ExitCode = result.ExitCode
		em.updateRequest(req)

		return ExecutionCompletedMsg{
			ID:     req.ID,
			Result: result,
			Error:  nil,
		}
	}
}

// CancelExecution cancels a running execution
func (em *ExecutionManager) CancelExecution(id string) tea.Cmd {
	return func() tea.Msg {
		req := em.findRequest(id)
		if req != nil {
			req.State = ExecutionCancelled
			now := time.Now()
			req.CompletedAt = &now
			em.updateRequest(req)
		}
		
		return ExecutionCancelledMsg{ID: id}
	}
}

// GetExecutionHistory returns the execution history
func (em *ExecutionManager) GetExecutionHistory() []ExecutionRequest {
	return em.requests
}

// GetActiveExecutions returns currently running executions
func (em *ExecutionManager) GetActiveExecutions() []ExecutionRequest {
	var active []ExecutionRequest
	for _, req := range em.requests {
		if req.State == ExecutionRunning || req.State == ExecutionPending {
			active = append(active, req)
		}
	}
	return active
}

// GetExecutionStats returns execution statistics
func (em *ExecutionManager) GetExecutionStats() map[string]interface{} {
	total := len(em.requests)
	completed := 0
	failed := 0
	cancelled := 0
	
	for _, req := range em.requests {
		switch req.State {
		case ExecutionCompleted:
			completed++
		case ExecutionFailed:
			failed++
		case ExecutionCancelled:
			cancelled++
		}
	}
	
	return map[string]interface{}{
		"total":     total,
		"completed": completed,
		"failed":    failed,
		"cancelled": cancelled,
		"success_rate": func() float64 {
			if total == 0 {
				return 0.0
			}
			return float64(completed) / float64(total) * 100.0
		}(),
	}
}

// GetCommandSuggestion provides suggestions for command execution
func (em *ExecutionManager) GetCommandSuggestion(command string) (string, error) {
	resolution, err := em.commandResolver.ResolveCommand(em.ctx, command)
	if err != nil {
		return "", err
	}
	
	if resolution == nil {
		return fmt.Sprintf("Command '%s' status unknown", command), nil
	}
	
	return em.commandResolver.GetExecutionSuggestion(resolution), nil
}

// ResolveCommand exposes command resolution functionality
func (em *ExecutionManager) ResolveCommand(command string) (*execution.CommandResolution, error) {
	return em.commandResolver.ResolveCommand(em.ctx, command)
}

// GetCommandResolutionStats returns command resolution cache statistics
func (em *ExecutionManager) GetCommandResolutionStats() map[string]interface{} {
	return em.commandResolver.GetCacheStats()
}

// Close cleans up the execution manager
func (em *ExecutionManager) Close() {
	if em.cancel != nil {
		em.cancel()
	}
}

// Helper methods

func (em *ExecutionManager) generateID() string {
	em.currentID++
	return fmt.Sprintf("exec_%d_%d", time.Now().Unix(), em.currentID)
}

func (em *ExecutionManager) updateRequest(req *ExecutionRequest) {
	for i, existing := range em.requests {
		if existing.ID == req.ID {
			em.requests[i] = *req
			return
		}
	}
}

func (em *ExecutionManager) findRequest(id string) *ExecutionRequest {
	for i, req := range em.requests {
		if req.ID == id {
			return &em.requests[i]
		}
	}
	return nil
}

// FormatExecutionStatus formats execution status for display
func FormatExecutionStatus(req *ExecutionRequest, styles ThemeStyles) string {
	var statusStyle lipgloss.Style
	var statusText string
	
	switch req.State {
	case ExecutionIdle:
		statusStyle = styles.prompt
		statusText = "⏸ Idle"
	case ExecutionDetecting:
		statusStyle = styles.warning
		statusText = "🔍 Detecting"
	case ExecutionPending:
		statusStyle = styles.warning
		statusText = "⏳ Pending"
	case ExecutionRunning:
		statusStyle = styles.accent
		statusText = "⚡ Running"
	case ExecutionCompleted:
		statusStyle = styles.success
		statusText = "✅ Completed"
	case ExecutionFailed:
		statusStyle = styles.error
		statusText = "❌ Failed"
	case ExecutionCancelled:
		statusStyle = styles.muted
		statusText = "🚫 Cancelled"
	default:
		statusStyle = styles.muted
		statusText = "❓ Unknown"
	}
	
	return statusStyle.Render(statusText)
}

// FormatExecutionRequest formats an execution request for display
func FormatExecutionRequest(req *ExecutionRequest, styles ThemeStyles) string {
	var lines []string
	
	// Header with ID and status
	header := lipgloss.JoinHorizontal(
		lipgloss.Center,
		styles.prompt.Render(fmt.Sprintf("Execution %s", req.ID[:8])),
		" ",
		FormatExecutionStatus(req, styles),
	)
	lines = append(lines, header)
	
	// Command
	if req.Command != "" {
		cmdText := fmt.Sprintf("%s %s", req.Command, strings.Join(req.Args, " "))
		lines = append(lines, styles.commandBox.Render(fmt.Sprintf("Command: %s", cmdText)))
	}
	
	// Description
	if req.Description != "" {
		lines = append(lines, styles.muted.Render(fmt.Sprintf("Description: %s", req.Description)))
	}
	
	// Category
	if req.Category != "" {
		lines = append(lines, styles.accent.Render(fmt.Sprintf("Category: %s", req.Category)))
	}
	
	// Timing information
	if req.StartedAt != nil {
		duration := "running"
		if req.CompletedAt != nil {
			duration = req.Duration.String()
		}
		lines = append(lines, styles.timestamp.Render(fmt.Sprintf("Duration: %s", duration)))
	}
	
	// Output (if available and not too long)
	if req.Output != "" && len(req.Output) < 200 {
		lines = append(lines, "")
		lines = append(lines, styles.success.Render("Output:"))
		lines = append(lines, styles.output.Render(req.Output))
	}
	
	// Error (if any)
	if req.Error != "" {
		lines = append(lines, "")
		lines = append(lines, styles.error.Render("Error:"))
		lines = append(lines, styles.error.Render(req.Error))
	}
	
	return strings.Join(lines, "\n")
}

// FormatExecutionSummary formats execution statistics for display
func FormatExecutionSummary(stats map[string]interface{}, styles ThemeStyles) string {
	total := stats["total"].(int)
	completed := stats["completed"].(int)
	failed := stats["failed"].(int)
	cancelled := stats["cancelled"].(int)
	successRate := stats["success_rate"].(float64)
	
	var lines []string
	lines = append(lines, styles.header.Render("📊 Execution Summary"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Total executions: %s", styles.accent.Render(fmt.Sprintf("%d", total))))
	lines = append(lines, fmt.Sprintf("Completed: %s", styles.success.Render(fmt.Sprintf("%d", completed))))
	lines = append(lines, fmt.Sprintf("Failed: %s", styles.error.Render(fmt.Sprintf("%d", failed))))
	lines = append(lines, fmt.Sprintf("Cancelled: %s", styles.muted.Render(fmt.Sprintf("%d", cancelled))))
	lines = append(lines, fmt.Sprintf("Success rate: %s", styles.accent.Render(fmt.Sprintf("%.1f%%", successRate))))
	
	return strings.Join(lines, "\n")
}