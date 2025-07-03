package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	showExecutionPanel bool                    // Show execution status panel
	executionOutput    []string                // Execution-specific output
	confirmPrompt      string                  // Confirmation prompt text
	
	// Layout
	leftPanelWidth     int                     // Width of left panel
	rightPanelWidth    int                     // Width of right panel
	panelSplit         bool                    // Whether to show split panel view
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
		showExecutionPanel: true,
		panelSplit:       true,
		leftPanelWidth:   60,
		rightPanelWidth:  40,
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
			// Toggle execution panel
			m.showExecutionPanel = !m.showExecutionPanel
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
			// Handle execution history navigation
			if m.mode == ModeHistory {
				return m.handleHistoryNavigation(msg.String())
			}
			
		case "enter":
			// Handle command execution with AI detection
			if m.mode == ModeNormal {
				input := strings.TrimSpace(m.textInput.Value())
				if input != "" {
					return m.handleCommandInput(input)
				}
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
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height
		
		if m.panelSplit {
			m.leftPanelWidth = int(float64(msg.Width) * 0.6)
			m.rightPanelWidth = msg.Width - m.leftPanelWidth - 2
		} else {
			m.leftPanelWidth = msg.Width
			m.rightPanelWidth = 0
		}
		
		m.textInput.Width = m.leftPanelWidth - 8
		return m, nil
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
	if !m.panelSplit || !m.showExecutionPanel {
		// Single panel mode or execution panel hidden
		return m.renderSinglePanel()
	}
	
	// Split panel mode
	leftPanel := m.renderLeftPanel()
	rightPanel := m.renderRightPanel()
	
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		m.styles.commandBox.Width(1).Render("│"),
		rightPanel,
	)
}

// renderLeftPanel renders the main command panel
func (m *ExecutionAwareTUI) renderLeftPanel() string {
	var sections []string
	
	// Header
	header := m.renderHeader()
	sections = append(sections, header)
	
	// Mode indicator
	if m.mode != ModeNormal {
		modeIndicator := m.renderModeIndicator()
		sections = append(sections, modeIndicator)
	}
	
	// Confirmation prompt (if in confirm mode)
	if m.mode == ModeConfirm && m.pendingExecution != nil {
		confirmSection := m.renderConfirmationSection()
		sections = append(sections, confirmSection)
	}
	
	// Command input or history view
	if m.mode == ModeHistory {
		historySection := m.renderHistorySection()
		sections = append(sections, historySection)
	} else {
		inputSection := m.renderInputSection()
		sections = append(sections, inputSection)
	}
	
	// Output
	outputSection := m.renderOutputSection()
	sections = append(sections, outputSection)
	
	// Footer with keybindings
	footer := m.renderFooter()
	sections = append(sections, footer)
	
	content := strings.Join(sections, "\n\n")
	
	return lipgloss.NewStyle().
		Width(m.leftPanelWidth).
		Render(content)
}

// renderRightPanel renders the execution status panel
func (m *ExecutionAwareTUI) renderRightPanel() string {
	var sections []string
	
	// Execution panel header
	header := m.styles.header.Render("⚡ Execution Manager")
	sections = append(sections, header)
	
	// Active executions
	activeExecs := m.executionManager.GetActiveExecutions()
	if len(activeExecs) > 0 {
		activeSection := m.renderActiveExecutions(activeExecs)
		sections = append(sections, activeSection)
	}
	
	// Execution statistics
	stats := m.executionManager.GetExecutionStats()
	statsSection := FormatExecutionSummary(stats, m.styles)
	sections = append(sections, statsSection)
	
	// Recent execution output
	if len(m.executionOutput) > 0 {
		outputHeader := m.styles.warning.Render("📜 Recent Activity")
		sections = append(sections, outputHeader)
		
		// Show last 10 execution messages
		recentOutput := m.executionOutput
		if len(recentOutput) > 10 {
			recentOutput = recentOutput[len(recentOutput)-10:]
		}
		
		for _, line := range recentOutput {
			sections = append(sections, m.styles.output.Render(line))
		}
	}
	
	// Execution capabilities info
	if m.providerManager != nil && m.providerManager.IsExecutionEnabled() {
		capHeader := m.styles.accent.Render("🔧 Capabilities")
		sections = append(sections, capHeader)
		
		capInfo := []string{
			"✅ Execution detection enabled",
			fmt.Sprintf("🤖 Auto-execute: %v", m.providerManager.IsAutoExecuteEnabled()),
			"🛡️ Security validation active",
			"📝 Audit logging enabled",
		}
		
		for _, info := range capInfo {
			sections = append(sections, m.styles.muted.Render(info))
		}
	}
	
	content := strings.Join(sections, "\n\n")
	
	return lipgloss.NewStyle().
		Width(m.rightPanelWidth).
		Render(content)
}

// Helper methods for rendering different sections

func (m *ExecutionAwareTUI) renderSinglePanel() string {
	// Use the base TUI view with execution enhancements
	baseView := m.ClaudeTUI.View()
	
	// Add execution status if there are active executions
	activeExecs := m.executionManager.GetActiveExecutions()
	if len(activeExecs) > 0 {
		statusLine := fmt.Sprintf("⚡ %d active execution(s)", len(activeExecs))
		baseView = m.styles.warning.Render(statusLine) + "\n" + baseView
	}
	
	return baseView
}

func (m *ExecutionAwareTUI) renderHeader() string {
	title := "NixAI - AI-Powered Command Execution"
	return m.styles.header.Render(title)
}

func (m *ExecutionAwareTUI) renderModeIndicator() string {
	var modeText string
	var style lipgloss.Style
	
	switch m.mode {
	case ModeExecution:
		modeText = "🔧 EXECUTION MODE"
		style = m.styles.warning
	case ModeHistory:
		modeText = "📜 HISTORY MODE"
		style = m.styles.accent
	case ModeConfirm:
		modeText = "❓ CONFIRMATION MODE"
		style = m.styles.error
	default:
		return ""
	}
	
	return style.Render(modeText)
}

func (m *ExecutionAwareTUI) renderConfirmationSection() string {
	if m.pendingExecution == nil {
		return ""
	}
	
	var lines []string
	lines = append(lines, m.styles.warning.Render("⚠️ Execution Confirmation Required"))
	lines = append(lines, "")
	lines = append(lines, FormatExecutionRequest(m.pendingExecution, m.styles))
	lines = append(lines, "")
	lines = append(lines, m.styles.prompt.Render("Execute this command? (y/N):"))
	
	return strings.Join(lines, "\n")
}

func (m *ExecutionAwareTUI) renderInputSection() string {
	// Render the command input with execution detection hints
	inputBox := m.styles.commandBox.Render(m.textInput.View())
	
	// Add suggestion for execution commands
	if m.textInput.Value() != "" {
		// Check if current input might be an execution request
		if req, _ := m.executionManager.DetectExecutionRequest(m.textInput.Value()); req != nil {
			hint := m.styles.accent.Render("💡 Execution detected - press Enter to proceed")
			return inputBox + "\n" + hint
		}
	}
	
	return inputBox
}

func (m *ExecutionAwareTUI) renderOutputSection() string {
	// Render output similar to the base TUI
	outputLines := m.output
	maxLines := 10 // Limit for split panel view
	if len(outputLines) > maxLines {
		outputLines = outputLines[len(outputLines)-maxLines:]
	}
	
	outputContent := strings.Join(outputLines, "\n")
	if outputContent == "" {
		outputContent = m.styles.muted.Render("No output yet...")
	}
	
	return m.styles.output.Render(outputContent)
}

func (m *ExecutionAwareTUI) renderHistorySection() string {
	history := m.executionManager.GetExecutionHistory()
	if len(history) == 0 {
		return m.styles.muted.Render("No execution history available")
	}
	
	var lines []string
	lines = append(lines, m.styles.header.Render("📜 Execution History"))
	lines = append(lines, "")
	
	// Show recent executions (last 10)
	start := 0
	if len(history) > 10 {
		start = len(history) - 10
	}
	
	for i := start; i < len(history); i++ {
		req := history[i]
		style := m.styles.output
		if i == m.selectedExecution {
			style = m.styles.selectedSug
		}
		
		summary := fmt.Sprintf("%s - %s %s", 
			req.CreatedAt.Format("15:04:05"),
			req.Command,
			FormatExecutionStatus(&req, m.styles))
		
		lines = append(lines, style.Render(summary))
	}
	
	return strings.Join(lines, "\n")
}

func (m *ExecutionAwareTUI) renderActiveExecutions(executions []ExecutionRequest) string {
	var lines []string
	lines = append(lines, m.styles.warning.Render("🏃 Active Executions"))
	lines = append(lines, "")
	
	for _, exec := range executions {
		status := FormatExecutionStatus(&exec, m.styles)
		summary := fmt.Sprintf("%s: %s", exec.ID[:8], status)
		lines = append(lines, summary)
	}
	
	return strings.Join(lines, "\n")
}

func (m *ExecutionAwareTUI) renderFooter() string {
	keybinds := []string{
		"Ctrl+E: Toggle Execution Panel",
		"Ctrl+H: Execution History", 
		"Ctrl+X: Execution Mode",
		"Ctrl+C: Quit",
	}
	
	return m.styles.muted.Render(strings.Join(keybinds, " | "))
}

// Event handlers

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
		return m, cmd
	}
	
	// No execution detected, handle as normal command
	baseTUI, cmd := m.ClaudeTUI.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m.ClaudeTUI = baseTUI.(*ClaudeTUI)
	m.textInput.SetValue("")
	
	return m, cmd
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
	// Add execution-specific suggestions
	execSuggestions := []string{
		"execute --help",
		"execute status", 
		"execute config",
		"execute history",
		"install firefox",
		"rebuild nixos",
		"update system",
		"check services",
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