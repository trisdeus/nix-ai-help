package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// ExecutionMode represents the current mode of the TUI
type ExecutionMode string

const (
	ModeNormal     ExecutionMode = "normal"     // Normal command mode
	ModeExecution  ExecutionMode = "execution"  // Execution management mode
	ModeHistory    ExecutionMode = "history"    // Execution history view
	ModeConfirm    ExecutionMode = "confirm"    // Execution confirmation
)

// ExecutionAwareTUI extends the basic TUI with execution capabilities
type ExecutionAwareTUI struct {
	*ClaudeTUI                                 // Embed the existing TUI
	executionManager   *ExecutionManager       // Execution management
	providerManager    *ai.ProviderManager     // AI provider management
	config             *config.UserConfig     // Configuration
	logger             *logger.Logger         // Logger
	
	// Execution state
	mode               ExecutionMode           // Current mode
	pendingExecution   *ExecutionRequest       // Pending execution request
	selectedExecution  int                     // Selected execution in history
	executionOutput    []string                // Execution-specific output
	confirmPrompt      string                  // Confirmation prompt text
}

// NewExecutionAwareTUI creates a new execution-aware TUI
func NewExecutionAwareTUI(cfg *config.UserConfig, log *logger.Logger) (*ExecutionAwareTUI, error) {
	// Create base TUI
	baseTUI := NewClaudeTUI()
	
	// Create provider manager
	providerManager := ai.NewProviderManager(cfg, log)
	
	// Create execution manager
	executionManager := NewExecutionManager(providerManager, log)
	
	tui := &ExecutionAwareTUI{
		ClaudeTUI:        baseTUI,
		executionManager: executionManager,
		providerManager:  providerManager,
		config:           cfg,
		logger:           log,
		mode:             ModeNormal,
		executionOutput:  []string{},
	}
	
	// Update suggestions to include execution commands
	tui.updateExecutionSuggestions()
	
	return tui, nil
}

// Init initializes the execution-aware TUI
func (m *ExecutionAwareTUI) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.ClaudeTUI.Init(),
		m.checkForExecutionUpdates(),
	)
}

// Update handles messages and state updates
func (m *ExecutionAwareTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+e":
			// Show execution status
			m.showExecutionStatus()
			return m, nil
			
		case "ctrl+h":
			// Show execution history
			if m.mode == ModeHistory {
				m.mode = ModeNormal
			} else {
				m.mode = ModeHistory
				m.updateExecutionHistory()
			}
			return m, nil
			
		case "ctrl+x":
			// Toggle execution mode
			if m.mode == ModeExecution {
				m.mode = ModeNormal
			} else {
				m.mode = ModeExecution
			}
			return m, nil
			
		case "y", "Y":
			// Confirm execution (when in confirm mode)
			if m.mode == ModeConfirm && m.pendingExecution != nil {
				cmd = m.executionManager.StartExecution(m.pendingExecution)
				m.mode = ModeNormal
				m.pendingExecution = nil
				m.addExecutionOutput("🚀 Execution started...")
				return m, cmd
			}
			
		case "n", "N":
			// Cancel execution (when in confirm mode)
			if m.mode == ModeConfirm {
				m.mode = ModeNormal
				m.pendingExecution = nil
				m.addExecutionOutput("❌ Execution cancelled")
				return m, nil
			}
			
		case "esc":
			// Exit special modes
			if m.mode != ModeNormal {
				m.mode = ModeNormal
				m.pendingExecution = nil
				return m, nil
			}
			
		case "up", "down":
			// Handle execution history navigation or suggestion navigation
			if m.mode == ModeHistory {
				return m.handleHistoryNavigation(msg.String())
			}
			// Otherwise, let base TUI handle suggestion navigation
			var baseTUI tea.Model
			baseTUI, cmd = m.ClaudeTUI.Update(msg)
			m.ClaudeTUI = baseTUI.(*ClaudeTUI)
			return m, cmd
			
		case "tab":
			// Handle tab completion with execution detection
			if m.mode == ModeNormal {
				return m.handleTabCompletion()
			}
			
		case "enter":
			// Handle command execution with AI detection
			if m.mode == ModeNormal {
				input := strings.TrimSpace(m.textInput.Value())
				if input != "" {
					return m.handleCommandInput(input)
				}
			}
			
		default:
			// Forward all other keys to base TUI for normal handling (typing, etc.)
			if m.mode == ModeNormal {
				var baseTUI tea.Model
				baseTUI, cmd = m.ClaudeTUI.Update(msg)
				m.ClaudeTUI = baseTUI.(*ClaudeTUI)
				return m, cmd
			}
		}
		
	case ExecutionDetectedMsg:
		// Handle execution detection
		m.pendingExecution = &msg.Request
		m.mode = ModeConfirm
		m.confirmPrompt = m.formatConfirmationPrompt(&msg.Request)
		m.addExecutionOutput("🔍 Execution detected!")
		return m, nil
		
	case ExecutionStartedMsg:
		// Handle execution start
		m.addExecutionOutput(fmt.Sprintf("⚡ Execution %s started", msg.ID[:8]))
		return m, nil
		
	case ExecutionCompletedMsg:
		// Handle execution completion
		if msg.Error != nil {
			m.addExecutionOutput(fmt.Sprintf("❌ Execution %s failed: %v", msg.ID[:8], msg.Error))
		} else {
			m.addExecutionOutput(fmt.Sprintf("✅ Execution %s completed successfully", msg.ID[:8]))
			if msg.Result != nil && msg.Result.Output != "" {
				m.addExecutionOutput(fmt.Sprintf("Output: %s", msg.Result.Output))
			}
		}
		return m, nil
		
	case ExecutionOutputMsg:
		// Handle streaming execution output
		prefix := "📤"
		if msg.IsError {
			prefix = "🚨"
		}
		m.addExecutionOutput(fmt.Sprintf("%s %s: %s", prefix, msg.ID[:8], msg.Output))
		return m, nil
		
	case ExecutionCancelledMsg:
		// Handle execution cancellation
		m.addExecutionOutput(fmt.Sprintf("🚫 Execution %s cancelled", msg.ID[:8]))
		return m, nil
		
	case tea.WindowSizeMsg:
		// Handle window resize - pass to base TUI
		var baseTUI tea.Model
		baseTUI, cmd = m.ClaudeTUI.Update(msg)
		m.ClaudeTUI = baseTUI.(*ClaudeTUI)
		return m, cmd
	}
	
	// Update base TUI if not handled above
	if m.mode == ModeNormal && msg != nil {
		var baseTUI tea.Model
		baseTUI, cmd = m.ClaudeTUI.Update(msg)
		m.ClaudeTUI = baseTUI.(*ClaudeTUI)
		cmds = append(cmds, cmd)
	}
	
	// Add periodic execution update check
	cmds = append(cmds, m.checkForExecutionUpdates())
	
	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m *ExecutionAwareTUI) View() string {
	// Use the base TUI view for the main interface
	baseView := m.ClaudeTUI.View()
	
	// Handle special execution modes
	if m.mode == ModeConfirm && m.pendingExecution != nil {
		return m.renderConfirmationView(baseView)
	}
	
	if m.mode == ModeHistory {
		return m.renderHistoryView()
	}
	
	// Add simple execution status if there are active executions
	return m.addExecutionStatusToBaseView(baseView)
}

// renderConfirmationView shows execution confirmation overlay
func (m *ExecutionAwareTUI) renderConfirmationView(baseView string) string {
	lines := strings.Split(baseView, "\n")
	width := m.ClaudeTUI.width
	
	// Create centered confirmation prompt
	centerText := func(text string) string {
		if width > 0 && len(text) < width {
			padding := (width - len(text)) / 2
			if padding > 0 {
				return strings.Repeat(" ", padding) + text
			}
		}
		return text
	}
	
	// Insert confirmation prompt before command input
	confirmHeader := "Execution Confirmation Required"
	commandText := fmt.Sprintf("Command: %s %s", m.pendingExecution.Command, strings.Join(m.pendingExecution.Args, " "))
	promptText := "Execute this command? (y/N):"
	
	confirmLines := []string{
		"",
		m.styles.warning.Render(centerText(confirmHeader)),
		"",
		centerText(commandText),
		"",
		m.styles.prompt.Render(centerText(promptText)),
		"",
	}
	
	// Find the command input line and insert confirmation above it
	for i, line := range lines {
		if strings.Contains(line, "nixai >") {
			result := append(lines[:i], confirmLines...)
			result = append(result, lines[i:]...)
			return strings.Join(result, "\n")
		}
	}
	
	return baseView
}

// renderHistoryView shows execution history
func (m *ExecutionAwareTUI) renderHistoryView() string {
	var sections []string
	width := m.ClaudeTUI.width

	// Center text helper
	centerText := func(text string) string {
		if width > 0 && len(text) < width {
			padding := (width - len(text)) / 2
			if padding > 0 {
				return strings.Repeat(" ", padding) + text
			}
		}
		return text
	}

	// Header
	headerText := "Execution History"
	header := m.styles.header.Render(centerText(headerText))
	sections = append(sections, header)

	// History content
	history := m.executionManager.GetExecutionHistory()
	if len(history) == 0 {
		noHistoryText := "No execution history available"
		sections = append(sections, m.styles.muted.Render(centerText(noHistoryText)))
	} else {
		for i, req := range history {
			style := m.styles.output
			if i == m.selectedExecution {
				style = m.styles.selectedSug
			}
			
			summary := fmt.Sprintf("%s - %s [%s]", 
				req.CreatedAt.Format("15:04:05"),
				req.Command,
				req.State)
			
			sections = append(sections, style.Render(summary))
		}
	}

	sections = append(sections, "")
	escText := "Press Esc to return"
	sections = append(sections, m.styles.muted.Render(centerText(escText)))

	return strings.Join(sections, "\n")
}

// addExecutionStatusToBaseView adds simple execution info to the base view
func (m *ExecutionAwareTUI) addExecutionStatusToBaseView(baseView string) string {
	lines := strings.Split(baseView, "\n")
	
	// Find version line and add execution status above it
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "nixai v") {
			executionStatus := m.renderSimpleExecutionStatus()
			if executionStatus != "" {
				result := append(lines[:i], executionStatus)
				result = append(result, lines[i:]...)
				return strings.Join(result, "\n")
			}
			break
		}
	}
	
	return baseView
}

// renderSimpleExecutionStatus renders a simple one-line execution status
func (m *ExecutionAwareTUI) renderSimpleExecutionStatus() string {
	activeExecs := m.executionManager.GetActiveExecutions()
	stats := m.executionManager.GetExecutionStats()
	
	var statusText string
	if len(activeExecs) > 0 {
		statusText = fmt.Sprintf("Execution: %d active | Total: %d", 
			len(activeExecs), stats["total"])
	} else if stats["total"].(int) > 0 {
		statusText = fmt.Sprintf("Execution: %d completed | Success rate: %.1f%%", 
			stats["completed"], stats["success_rate"])
	} else {
		return ""
	}
	
	// Center the text by calculating padding
	width := m.ClaudeTUI.width
	if width > 0 && len(statusText) < width {
		padding := (width - len(statusText)) / 2
		if padding > 0 {
			statusText = strings.Repeat(" ", padding) + statusText
		}
	}
	
	return m.styles.muted.Render(statusText)
}

// Helper methods for rendering different sections

func (m *ExecutionAwareTUI) showExecutionStatus() {
	stats := m.executionManager.GetExecutionStats()
	activeExecs := m.executionManager.GetActiveExecutions()
	
	statusMsg := fmt.Sprintf("Execution Status: Total=%d, Active=%d, Completed=%d, Success Rate=%.1f%%",
		stats["total"], len(activeExecs), stats["completed"], stats["success_rate"])
	
	m.ClaudeTUI.addOutput(m.styles.accent.Render(statusMsg))
}

func (m *ExecutionAwareTUI) renderModeIndicator() string {
	switch m.mode {
	case ModeExecution:
		return m.styles.warning.Render("EXECUTION MODE")
	case ModeHistory:
		return m.styles.accent.Render("HISTORY MODE")
	case ModeConfirm:
		return m.styles.error.Render("CONFIRMATION MODE")
	default:
		return ""
	}
}

// Event handlers

func (m *ExecutionAwareTUI) handleTabCompletion() (tea.Model, tea.Cmd) {
	// First check if we have suggestions to complete
	if m.showSuggestions && len(m.getFilteredSuggestions()) > 0 {
		filtered := m.getFilteredSuggestions()
		if m.selectedSuggestion >= 0 && m.selectedSuggestion < len(filtered) {
			selectedCommand := filtered[m.selectedSuggestion]
			m.textInput.SetValue(selectedCommand)
			m.showSuggestions = false
			
			// Check if the completed command might be an execution request
			if execReq, _ := m.executionManager.DetectExecutionRequest(selectedCommand); execReq != nil {
				// Show hint that this could be executed
				m.ClaudeTUI.addOutput(m.styles.accent.Render("💡 Execution detected - press Enter to proceed"))
			}
		}
	}
	return m, nil
}

func (m *ExecutionAwareTUI) handleCommandInput(input string) (tea.Model, tea.Cmd) {
	// First, try to detect execution requests
	execReq, err := m.executionManager.DetectExecutionRequest(input)
	if err != nil {
		m.addExecutionOutput(fmt.Sprintf("Error detecting execution: %v", err))
	}
	
	if execReq != nil {
		// Execution detected, request confirmation
		cmd := m.executionManager.RequestExecution(execReq)
		m.textInput.SetValue("")
		// Add to command history for execution commands too
		m.commandHistory = append(m.commandHistory, input)
		m.historyIndex = -1
		return m, cmd
	}
	
	// No execution detected, handle as normal nixai command
	m.ClaudeTUI.executeCommand(input)
	m.commandHistory = append(m.commandHistory, input)
	m.historyIndex = -1
	m.textInput.SetValue("")
	m.showSuggestions = false
	
	return m, nil
}

func (m *ExecutionAwareTUI) handleHistoryNavigation(direction string) (tea.Model, tea.Cmd) {
	history := m.executionManager.GetExecutionHistory()
	if len(history) == 0 {
		return m, nil
	}
	
	switch direction {
	case "up":
		if m.selectedExecution > 0 {
			m.selectedExecution--
		}
	case "down":
		if m.selectedExecution < len(history)-1 {
			m.selectedExecution++
		}
	}
	
	return m, nil
}

// Utility methods

func (m *ExecutionAwareTUI) updateExecutionSuggestions() {
	// Add execution-specific suggestions that will trigger execution detection
	execSuggestions := []string{
		// Package management (these should trigger execution)
		"install firefox",
		"install git", 
		"install vim",
		"install code",
		"install nodejs",
		"install docker",
		"update system",
		"upgrade packages", 
		"remove firefox",
		"uninstall git",
		
		// System management (these should trigger execution)
		"rebuild nixos",
		"rebuild switch", 
		"start docker",
		"stop nginx",
		"restart apache",
		"enable service",
		"disable firewall",
		"check services",
		"status nginx",
		
		// Direct command suggestions (will trigger execution)
		"nix-env -i firefox",
		"nix-env -u",
		"nix-collect-garbage -d", 
		"nixos-rebuild switch",
		"systemctl restart nginx",
		"systemctl status docker",
		"systemctl start postgresql",
		"systemctl stop apache2",
		
		// Natural language that should trigger execution
		"please install firefox",
		"can you start docker",
		"run nix-collect-garbage",
		"execute nixos-rebuild switch",
		
		// Regular nixai commands (these won't trigger execution)
		"execute --help",
		"execute status", 
		"execute config",
		"execute history",
		"ask \"how to install nodejs\"",
		"diagnose boot", 
		"search firefox",
		"help install",
		"explain-option services.nginx",
		"flake create",
	}
	
	// Merge with existing suggestions
	allSuggestions := append(m.suggestions, execSuggestions...)
	m.suggestions = removeDuplicates(allSuggestions)
}

func (m *ExecutionAwareTUI) updateExecutionHistory() {
	history := m.executionManager.GetExecutionHistory()
	if len(history) > 0 {
		m.selectedExecution = len(history) - 1
	}
}

func (m *ExecutionAwareTUI) addExecutionOutput(message string) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := fmt.Sprintf("[%s] %s", timestamp, message)
	m.executionOutput = append(m.executionOutput, formattedMsg)
	
	// Keep only last 50 messages
	if len(m.executionOutput) > 50 {
		m.executionOutput = m.executionOutput[len(m.executionOutput)-50:]
	}
}

func (m *ExecutionAwareTUI) formatConfirmationPrompt(req *ExecutionRequest) string {
	return fmt.Sprintf("Execute command: %s %s", req.Command, strings.Join(req.Args, " "))
}

func (m *ExecutionAwareTUI) checkForExecutionUpdates() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		// This could be used to check for execution status updates
		// For now, it's a placeholder for future enhancements
		return nil
	})
}

// Start runs the execution-aware TUI
func (m *ExecutionAwareTUI) Start() error {
	program := tea.NewProgram(m, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// Close cleans up resources
func (m *ExecutionAwareTUI) Close() error {
	if m.executionManager != nil {
		m.executionManager.Close()
	}
	return nil
}