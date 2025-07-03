package neovim

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
)

// Diagnostic represents a diagnostic message for Neovim
type Diagnostic struct {
	Range    DiagnosticRange    `json:"range"`
	Message  string             `json:"message"`
	Severity DiagnosticSeverity `json:"severity"`
	Source   string             `json:"source"`
	Code     string             `json:"code,omitempty"`
	Data     *DiagnosticData    `json:"data,omitempty"`
}

// DiagnosticRange represents the range of a diagnostic
type DiagnosticRange struct {
	Start DiagnosticPosition `json:"start"`
	End   DiagnosticPosition `json:"end"`
}

// DiagnosticPosition represents a position in the document
type DiagnosticPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// DiagnosticSeverity represents the severity of a diagnostic
type DiagnosticSeverity int

const (
	SeverityError DiagnosticSeverity = iota + 1
	SeverityWarning
	SeverityInformation
	SeverityHint
)

// DiagnosticData contains additional data for diagnostics
type DiagnosticData struct {
	QuickFix    *QuickFix `json:"quickFix,omitempty"`
	Explanation string    `json:"explanation,omitempty"`
	References  []string  `json:"references,omitempty"`
}

// QuickFix represents an automatic fix for a diagnostic
type QuickFix struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Edits       []TextEdit    `json:"edits"`
	Commands    []Command     `json:"commands,omitempty"`
}

// TextEdit represents a text edit
type TextEdit struct {
	Range   DiagnosticRange `json:"range"`
	NewText string          `json:"newText"`
}

// Command represents a command to execute
type Command struct {
	Title     string                 `json:"title"`
	Command   string                 `json:"command"`
	Arguments []interface{}          `json:"arguments,omitempty"`
}

// DiagnosticParams contains parameters for diagnostic requests
type DiagnosticParams struct {
	FilePath   string `json:"filePath"`
	BufferText string `json:"bufferText"`
	FileType   string `json:"fileType,omitempty"`
}

// DiagnosticProvider provides AI-powered diagnostics for Neovim
type DiagnosticProvider struct {
	aiManager *ai.Manager
	context   *nixos.Context
	logger    *logger.Logger
}

// NewDiagnosticProvider creates a new diagnostic provider
func NewDiagnosticProvider(aiManager *ai.Manager, context *nixos.Context, logger *logger.Logger) *DiagnosticProvider {
	return &DiagnosticProvider{
		aiManager: aiManager,
		context:   context,
		logger:    logger,
	}
}

// GetDiagnostics provides AI-powered diagnostic analysis
func (dp *DiagnosticProvider) GetDiagnostics(params DiagnosticParams) ([]Diagnostic, error) {
	// Create context for AI analysis
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get system context for better analysis
	systemContext, err := dp.context.GetFormattedContext()
	if err != nil {
		dp.logger.Warn("Failed to get system context for diagnostics: %v", err)
		systemContext = "Basic NixOS system"
	}

	// Build AI prompt for diagnostics
	prompt := dp.buildDiagnosticPrompt(params, systemContext)
	
	// Get AI analysis
	response, err := dp.aiManager.Query(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI diagnostics: %w", err)
	}

	// Parse and format diagnostic items
	diagnostics, err := dp.parseDiagnosticResponse(response, params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diagnostic response: %w", err)
	}

	return diagnostics, nil
}

// buildDiagnosticPrompt creates the AI prompt for diagnostics
func (dp *DiagnosticProvider) buildDiagnosticPrompt(params DiagnosticParams, systemContext string) string {
	prompt := fmt.Sprintf(`You are an expert NixOS diagnostic assistant. Analyze the following Nix configuration for potential issues, improvements, and best practices.

File: %s
System Context: %s

Configuration to analyze:
```nix
%s
```

Analyze for:
1. Syntax errors and typos
2. Deprecated options or functions
3. Security vulnerabilities
4. Performance optimizations
5. Best practice violations
6. Missing dependencies
7. Configuration conflicts
8. Unused or redundant configurations

For each issue found, provide:
- Exact line number (0-based)
- Severity (error, warning, info, hint)
- Clear description of the issue
- Suggested fix with exact replacement text
- Explanation of why this is important

Format as JSON array:
[
  {
    "line": 0,
    "startChar": 0,
    "endChar": 10,
    "severity": "error|warning|info|hint",
    "message": "Brief description of issue",
    "code": "nixos_error_code",
    "explanation": "Detailed explanation of the issue and its impact",
    "quickFix": {
      "title": "Fix description",
      "newText": "Replacement text",
      "description": "What this fix does"
    },
    "references": ["https://nixos.org/manual/...", "Additional resources"]
  }
]

Only include real issues - don't create false positives. Focus on actionable feedback.
`, params.FilePath, systemContext, params.BufferText)

	return prompt
}

// parseDiagnosticResponse parses AI response into diagnostic items
func (dp *DiagnosticProvider) parseDiagnosticResponse(response string, params DiagnosticParams) ([]Diagnostic, error) {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")
	
	if jsonStart == -1 || jsonEnd == -1 {
		return dp.createBasicDiagnostics(params), nil
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var rawDiagnostics []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawDiagnostics); err != nil {
		dp.logger.Warn("Failed to parse diagnostic JSON, using basic checks: %v", err)
		return dp.createBasicDiagnostics(params), nil
	}
	
	diagnostics := make([]Diagnostic, 0, len(rawDiagnostics))
	lines := strings.Split(params.BufferText, "\n")
	
	for _, raw := range rawDiagnostics {
		diagnostic := dp.createDiagnosticFromRaw(raw, lines)
		if diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
	}
	
	return diagnostics, nil
}

// createDiagnosticFromRaw creates a diagnostic from raw JSON data
func (dp *DiagnosticProvider) createDiagnosticFromRaw(raw map[string]interface{}, lines []string) *Diagnostic {
	line := dp.getIntField(raw, "line", 0)
	startChar := dp.getIntField(raw, "startChar", 0)
	endChar := dp.getIntField(raw, "endChar", 0)
	
	// Validate line number
	if line < 0 || line >= len(lines) {
		return nil
	}
	
	// Auto-detect end character if not provided
	if endChar <= startChar {
		endChar = len(lines[line])
	}
	
	diagnostic := &Diagnostic{
		Range: DiagnosticRange{
			Start: DiagnosticPosition{Line: line, Character: startChar},
			End:   DiagnosticPosition{Line: line, Character: endChar},
		},
		Message:  dp.getStringField(raw, "message"),
		Severity: dp.getSeverityField(raw, "severity"),
		Source:   "nixai",
		Code:     dp.getStringField(raw, "code"),
	}
	
	// Add diagnostic data if available
	if explanation := dp.getStringField(raw, "explanation"); explanation != "" {
		diagnostic.Data = &DiagnosticData{
			Explanation: explanation,
		}
	}
	
	// Add quick fix if available
	if quickFix, ok := raw["quickFix"].(map[string]interface{}); ok {
		if diagnostic.Data == nil {
			diagnostic.Data = &DiagnosticData{}
		}
		diagnostic.Data.QuickFix = &QuickFix{
			Title:       dp.getStringField(quickFix, "title"),
			Description: dp.getStringField(quickFix, "description"),
			Edits: []TextEdit{
				{
					Range:   diagnostic.Range,
					NewText: dp.getStringField(quickFix, "newText"),
				},
			},
		}
	}
	
	// Add references if available
	if refs, ok := raw["references"].([]interface{}); ok {
		if diagnostic.Data == nil {
			diagnostic.Data = &DiagnosticData{}
		}
		for _, ref := range refs {
			if refStr, ok := ref.(string); ok {
				diagnostic.Data.References = append(diagnostic.Data.References, refStr)
			}
		}
	}
	
	return diagnostic
}

// createBasicDiagnostics provides basic diagnostics when AI fails
func (dp *DiagnosticProvider) createBasicDiagnostics(params DiagnosticParams) []Diagnostic {
	var diagnostics []Diagnostic
	lines := strings.Split(params.BufferText, "\n")
	
	// Basic syntax checks
	for i, line := range lines {
		// Check for common syntax issues
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Check for missing semicolons in assignments
		if dp.isAssignmentLine(line) && !strings.HasSuffix(strings.TrimSpace(line), ";") {
			diagnostics = append(diagnostics, Diagnostic{
				Range: DiagnosticRange{
					Start: DiagnosticPosition{Line: i, Character: 0},
					End:   DiagnosticPosition{Line: i, Character: len(line)},
				},
				Message:  "Missing semicolon at end of assignment",
				Severity: SeverityWarning,
				Source:   "nixai-basic",
				Code:     "missing-semicolon",
			})
		}
		
		// Check for deprecated 'with pkgs;' usage
		if strings.Contains(line, "with pkgs;") && strings.Contains(params.BufferText, "environment.systemPackages") {
			diagnostics = append(diagnostics, Diagnostic{
				Range: DiagnosticRange{
					Start: DiagnosticPosition{Line: i, Character: strings.Index(line, "with pkgs;")},
					End:   DiagnosticPosition{Line: i, Character: strings.Index(line, "with pkgs;") + 10},
				},
				Message:  "Consider using explicit package references instead of 'with pkgs;'",
				Severity: SeverityHint,
				Source:   "nixai-basic",
				Code:     "with-pkgs-usage",
				Data: &DiagnosticData{
					Explanation: "'with pkgs;' can make dependencies unclear and cause conflicts. Consider using 'pkgs.packageName' directly.",
				},
			})
		}
	}
	
	return diagnostics
}

// Helper functions
func (dp *DiagnosticProvider) getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (dp *DiagnosticProvider) getIntField(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultVal
}

func (dp *DiagnosticProvider) getSeverityField(m map[string]interface{}, key string) DiagnosticSeverity {
	severityStr := dp.getStringField(m, key)
	switch strings.ToLower(severityStr) {
	case "error":
		return SeverityError
	case "warning":
		return SeverityWarning
	case "info", "information":
		return SeverityInformation
	case "hint":
		return SeverityHint
	default:
		return SeverityInformation
	}
}

func (dp *DiagnosticProvider) isAssignmentLine(line string) bool {
	// Simple check for assignment lines
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return false
	}
	
	// Look for assignment pattern
	assignmentPattern := regexp.MustCompile(`^\s*\w+(\.\w+)*\s*=`)
	return assignmentPattern.MatchString(line)
}

// CodeActionProvider provides code actions for diagnostics
type CodeActionProvider struct {
	diagnosticProvider *DiagnosticProvider
}

// NewCodeActionProvider creates a new code action provider
func NewCodeActionProvider(diagnosticProvider *DiagnosticProvider) *CodeActionProvider {
	return &CodeActionProvider{
		diagnosticProvider: diagnosticProvider,
	}
}

// CodeAction represents a code action
type CodeAction struct {
	Title       string                 `json:"title"`
	Kind        string                 `json:"kind"`
	Diagnostics []Diagnostic           `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit         `json:"edit,omitempty"`
	Command     *Command               `json:"command,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// WorkspaceEdit represents changes to be applied to the workspace
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes,omitempty"`
}

// GetCodeActions provides code actions for diagnostics
func (cap *CodeActionProvider) GetCodeActions(params DiagnosticParams, diagnostics []Diagnostic) []CodeAction {
	var actions []CodeAction
	
	for _, diagnostic := range diagnostics {
		if diagnostic.Data != nil && diagnostic.Data.QuickFix != nil {
			action := CodeAction{
				Title:       diagnostic.Data.QuickFix.Title,
				Kind:        "quickfix",
				Diagnostics: []Diagnostic{diagnostic},
				Edit: &WorkspaceEdit{
					Changes: map[string][]TextEdit{
						params.FilePath: diagnostic.Data.QuickFix.Edits,
					},
				},
			}
			actions = append(actions, action)
		}
	}
	
	// Add general improvement actions
	actions = append(actions, cap.getGeneralActions(params)...)
	
	return actions
}

// getGeneralActions provides general code actions
func (cap *CodeActionProvider) getGeneralActions(params DiagnosticParams) []CodeAction {
	return []CodeAction{
		{
			Title: "Format Nix file",
			Kind:  "source.fixAll",
			Command: &Command{
				Title:     "Format with nixpkgs-fmt",
				Command:   "nixai.formatFile",
				Arguments: []interface{}{params.FilePath},
			},
		},
		{
			Title: "Optimize configuration",
			Kind:  "source.organizeImports",
			Command: &Command{
				Title:     "AI-powered configuration optimization",
				Command:   "nixai.optimizeConfig",
				Arguments: []interface{}{params.FilePath},
			},
		},
	}
}