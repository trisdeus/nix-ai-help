package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/neovim"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"

	"github.com/sourcegraph/jsonrpc2"
)

// NeovimHandlers contains handlers for enhanced Neovim integration
type NeovimHandlers struct {
	completionProvider *neovim.CompletionProvider
	diagnosticProvider *neovim.DiagnosticProvider
	snippetProvider    *neovim.SnippetProvider
	codeActionProvider *neovim.CodeActionProvider
	logger             *logger.Logger
}

// NewNeovimHandlers creates new Neovim handlers
func NewNeovimHandlers(aiManager *ai.Manager, context *nixos.Context, logger *logger.Logger) *NeovimHandlers {
	completionProvider := neovim.NewCompletionProvider(aiManager, context, logger)
	diagnosticProvider := neovim.NewDiagnosticProvider(aiManager, context, logger)
	snippetProvider := neovim.NewSnippetProvider()
	codeActionProvider := neovim.NewCodeActionProvider(diagnosticProvider)

	return &NeovimHandlers{
		completionProvider: completionProvider,
		diagnosticProvider: diagnosticProvider,
		snippetProvider:    snippetProvider,
		codeActionProvider: codeActionProvider,
		logger:             logger,
	}
}

// HandleNeovimCompletion handles enhanced completion requests
func (nh *NeovimHandlers) HandleNeovimCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) error {
	var args map[string]interface{}
	if err := json.Unmarshal(*req.Params, &args); err != nil {
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "Invalid parameters",
		})
	}

	params := neovim.CompletionParams{
		FilePath:   getStringParam(args, "filePath"),
		Line:       getIntParam(args, "line"),
		Character:  getIntParam(args, "character"),
		Context:    getStringParam(args, "context"),
		LineText:   getStringParam(args, "lineText"),
		BufferText: getStringParam(args, "bufferText"),
	}

	items, err := nh.completionProvider.GetCompletions(params)
	if err != nil {
		nh.logger.Error("Failed to get completions: %v", err)
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: fmt.Sprintf("Completion failed: %v", err),
		})
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": nh.formatCompletionResponse(items),
			},
		},
	}

	return conn.Reply(ctx, req.ID, result)
}

// HandleNeovimDiagnostics handles enhanced diagnostic requests
func (nh *NeovimHandlers) HandleNeovimDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) error {
	var args map[string]interface{}
	if err := json.Unmarshal(*req.Params, &args); err != nil {
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "Invalid parameters",
		})
	}

	params := neovim.DiagnosticParams{
		FilePath:   getStringParam(args, "filePath"),
		BufferText: getStringParam(args, "bufferText"),
		FileType:   getStringParam(args, "fileType"),
	}

	diagnostics, err := nh.diagnosticProvider.GetDiagnostics(params)
	if err != nil {
		nh.logger.Error("Failed to get diagnostics: %v", err)
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: fmt.Sprintf("Diagnostic analysis failed: %v", err),
		})
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": nh.formatDiagnosticResponse(diagnostics),
			},
		},
	}

	return conn.Reply(ctx, req.ID, result)
}

// HandleNeovimSnippets handles snippet requests
func (nh *NeovimHandlers) HandleNeovimSnippets(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) error {
	var args map[string]interface{}
	if err := json.Unmarshal(*req.Params, &args); err != nil {
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "Invalid parameters",
		})
	}

	category := getStringParam(args, "category")
	search := getStringParam(args, "search")
	format := getStringParam(args, "format")

	var snippets map[string]neovim.Snippet
	var response string

	// Get snippets based on parameters
	if category != "" {
		snippets = nh.snippetProvider.GetSnippetsByCategory(category)
	} else if search != "" {
		snippets = nh.snippetProvider.SearchSnippets(search)
	} else {
		snippets = nh.snippetProvider.GetSnippets()
	}

	// Format response based on requested format
	switch format {
	case "luasnip":
		response = nh.formatSnippetsAsLuaSnip(snippets)
	case "vscode":
		response = nh.snippetProvider.ExportSnippetsToVSCode()
	case "raw":
		response = nh.formatSnippetsAsRaw(snippets)
	default:
		response = nh.formatSnippetsAsMarkdown(snippets)
	}

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": response,
			},
		},
	}

	return conn.Reply(ctx, req.ID, result)
}

// HandleNeovimCodeActions handles code action requests
func (nh *NeovimHandlers) HandleNeovimCodeActions(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) error {
	var args map[string]interface{}
	if err := json.Unmarshal(*req.Params, &args); err != nil {
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "Invalid parameters",
		})
	}

	params := neovim.DiagnosticParams{
		FilePath:   getStringParam(args, "filePath"),
		BufferText: getStringParam(args, "bufferText"),
		FileType:   getStringParam(args, "fileType"),
	}

	// Get diagnostics first
	diagnostics, err := nh.diagnosticProvider.GetDiagnostics(params)
	if err != nil {
		nh.logger.Error("Failed to get diagnostics for code actions: %v", err)
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInternalError,
			Message: fmt.Sprintf("Code action analysis failed: %v", err),
		})
	}

	// Get code actions
	codeActions := nh.codeActionProvider.GetCodeActions(params, diagnostics)

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": nh.formatCodeActionsResponse(codeActions),
			},
		},
	}

	return conn.Reply(ctx, req.ID, result)
}

// HandleNeovimHoverEnhanced handles enhanced hover requests
func (nh *NeovimHandlers) HandleNeovimHoverEnhanced(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) error {
	var args map[string]interface{}
	if err := json.Unmarshal(*req.Params, &args); err != nil {
		return conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "Invalid parameters",
		})
	}

	filePath := getStringParam(args, "filePath")
	line := getIntParam(args, "line")
	character := getIntParam(args, "character")
	word := getStringParam(args, "word")
	context := getStringParam(args, "context")

	// Enhanced hover logic would go here
	// For now, provide a comprehensive response
	hoverText := nh.generateEnhancedHover(filePath, line, character, word, context)

	result := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": hoverText,
			},
		},
	}

	return conn.Reply(ctx, req.ID, result)
}

// Helper functions for formatting responses

func (nh *NeovimHandlers) formatCompletionResponse(items []neovim.CompletionItem) string {
	if len(items) == 0 {
		return "No completion suggestions available."
	}

	var builder strings.Builder
	builder.WriteString("# AI-Powered Completion Suggestions\n\n")

	for i, item := range items {
		builder.WriteString(fmt.Sprintf("## %d. %s\n", i+1, item.Label))
		if item.Detail != "" {
			builder.WriteString(fmt.Sprintf("**Type:** %s\n\n", item.Detail))
		}
		if item.Documentation != "" {
			builder.WriteString(fmt.Sprintf("%s\n\n", item.Documentation))
		}
		if item.InsertText != "" && item.InsertText != item.Label {
			builder.WriteString(fmt.Sprintf("```nix\n%s\n```\n\n", item.InsertText))
		}
	}

	return builder.String()
}

func (nh *NeovimHandlers) formatDiagnosticResponse(diagnostics []neovim.Diagnostic) string {
	if len(diagnostics) == 0 {
		return "No issues found. Your Nix configuration looks good! ✅"
	}

	var builder strings.Builder
	builder.WriteString("# Diagnostic Analysis Results\n\n")

	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, diag := range diagnostics {
		var severity string
		switch diag.Severity {
		case neovim.SeverityError:
			severity = "🔴 ERROR"
			errorCount++
		case neovim.SeverityWarning:
			severity = "🟡 WARNING"
			warningCount++
		case neovim.SeverityInformation:
			severity = "🔵 INFO"
			infoCount++
		case neovim.SeverityHint:
			severity = "💡 HINT"
			infoCount++
		}

		builder.WriteString(fmt.Sprintf("## %s (Line %d)\n", severity, diag.Range.Start.Line+1))
		builder.WriteString(fmt.Sprintf("**Issue:** %s\n\n", diag.Message))

		if diag.Data != nil {
			if diag.Data.Explanation != "" {
				builder.WriteString(fmt.Sprintf("**Explanation:** %s\n\n", diag.Data.Explanation))
			}

			if diag.Data.QuickFix != nil {
				builder.WriteString(fmt.Sprintf("**Suggested Fix:** %s\n", diag.Data.QuickFix.Title))
				if diag.Data.QuickFix.Description != "" {
					builder.WriteString(fmt.Sprintf("*%s*\n\n", diag.Data.QuickFix.Description))
				}
				if len(diag.Data.QuickFix.Edits) > 0 {
					builder.WriteString("```nix\n")
					builder.WriteString(diag.Data.QuickFix.Edits[0].NewText)
					builder.WriteString("\n```\n\n")
				}
			}

			if len(diag.Data.References) > 0 {
				builder.WriteString("**References:**\n")
				for _, ref := range diag.Data.References {
					builder.WriteString(fmt.Sprintf("- %s\n", ref))
				}
				builder.WriteString("\n")
			}
		}

		builder.WriteString("---\n\n")
	}

	// Summary
	builder.WriteString(fmt.Sprintf("## Summary\n\n"))
	builder.WriteString(fmt.Sprintf("- 🔴 Errors: %d\n", errorCount))
	builder.WriteString(fmt.Sprintf("- 🟡 Warnings: %d\n", warningCount))
	builder.WriteString(fmt.Sprintf("- 🔵 Info/Hints: %d\n", infoCount))

	return builder.String()
}

func (nh *NeovimHandlers) formatSnippetsAsMarkdown(snippets map[string]neovim.Snippet) string {
	if len(snippets) == 0 {
		return "No snippets found for the specified criteria."
	}

	var builder strings.Builder
	builder.WriteString("# Nix Code Snippets\n\n")

	// Group by category
	categories := make(map[string][]neovim.Snippet)
	for _, snippet := range snippets {
		categories[snippet.Category] = append(categories[snippet.Category], snippet)
	}

	for category, categorySnippets := range categories {
		builder.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(category)))

		for _, snippet := range categorySnippets {
			builder.WriteString(fmt.Sprintf("### %s (`%s`)\n", snippet.Description, snippet.Prefix))
			builder.WriteString("```nix\n")
			builder.WriteString(strings.Join(snippet.Body, "\n"))
			builder.WriteString("\n```\n\n")
		}
	}

	return builder.String()
}

func (nh *NeovimHandlers) formatSnippetsAsLuaSnip(snippets map[string]neovim.Snippet) string {
	var builder strings.Builder
	builder.WriteString("-- nixai Nix snippets for LuaSnip\n")
	builder.WriteString("local ls = require('luasnip')\n")
	builder.WriteString("local s = ls.snippet\n")
	builder.WriteString("local fmt = require('luasnip.extras.fmt').fmt\n\n")
	builder.WriteString("return {\n")

	for _, snippet := range snippets {
		bodyStr := strings.Join(snippet.Body, "\\n")
		// Escape quotes and backslashes for Lua
		bodyStr = strings.ReplaceAll(bodyStr, "\\", "\\\\")
		bodyStr = strings.ReplaceAll(bodyStr, "\"", "\\\"")

		builder.WriteString(fmt.Sprintf("  s(\"%s\", fmt([[%s]], {}), { desc = \"%s\" }),\n",
			snippet.Prefix, bodyStr, snippet.Description))
	}

	builder.WriteString("}\n")
	return builder.String()
}

func (nh *NeovimHandlers) formatSnippetsAsRaw(snippets map[string]neovim.Snippet) string {
	data, _ := json.MarshalIndent(snippets, "", "  ")
	return string(data)
}

func (nh *NeovimHandlers) formatCodeActionsResponse(actions []neovim.CodeAction) string {
	if len(actions) == 0 {
		return "No code actions available."
	}

	var builder strings.Builder
	builder.WriteString("# Available Code Actions\n\n")

	for i, action := range actions {
		builder.WriteString(fmt.Sprintf("## %d. %s\n", i+1, action.Title))
		builder.WriteString(fmt.Sprintf("**Kind:** %s\n\n", action.Kind))

		if action.Edit != nil && len(action.Edit.Changes) > 0 {
			builder.WriteString("**Changes:**\n")
			for file, edits := range action.Edit.Changes {
				builder.WriteString(fmt.Sprintf("- File: %s\n", file))
				for _, edit := range edits {
					builder.WriteString(fmt.Sprintf("  - Line %d: Replace with `%s`\n",
						edit.Range.Start.Line+1, edit.NewText))
				}
			}
			builder.WriteString("\n")
		}

		if action.Command != nil {
			builder.WriteString(fmt.Sprintf("**Command:** %s\n\n", action.Command.Title))
		}
	}

	return builder.String()
}

func (nh *NeovimHandlers) generateEnhancedHover(filePath string, line, character int, word, context string) string {
	var builder strings.Builder
	builder.WriteString("# Enhanced Documentation\n\n")

	// Basic word information
	builder.WriteString(fmt.Sprintf("## %s\n\n", word))

	// Context-aware documentation based on the word
	if strings.HasPrefix(word, "services.") {
		builder.WriteString("**Type:** NixOS Service Configuration\n\n")
		builder.WriteString("This configures a system service in NixOS. Services are managed by systemd.\n\n")
		builder.WriteString("**Common patterns:**\n")
		builder.WriteString("```nix\n")
		builder.WriteString(fmt.Sprintf("%s = {\n", word))
		builder.WriteString("  enable = true;\n")
		builder.WriteString("  # Additional configuration options\n")
		builder.WriteString("};\n")
		builder.WriteString("```\n\n")
	} else if strings.HasPrefix(word, "hardware.") {
		builder.WriteString("**Type:** Hardware Configuration\n\n")
		builder.WriteString("This configures hardware-specific settings in NixOS.\n\n")
	} else if strings.HasPrefix(word, "environment.") {
		builder.WriteString("**Type:** Environment Configuration\n\n")
		builder.WriteString("This configures system environment settings like packages and variables.\n\n")
	} else if strings.Contains(context, "with pkgs;") {
		builder.WriteString("**Type:** Nix Package\n\n")
		builder.WriteString(fmt.Sprintf("This references the `%s` package from nixpkgs.\n\n", word))
		builder.WriteString("**Usage example:**\n")
		builder.WriteString("```nix\n")
		builder.WriteString("environment.systemPackages = with pkgs; [\n")
		builder.WriteString(fmt.Sprintf("  %s\n", word))
		builder.WriteString("];\n")
		builder.WriteString("```\n\n")
	}

	// Add contextual tips
	builder.WriteString("**💡 Tips:**\n")
	builder.WriteString("- Use `nixai explain-option` to get detailed documentation\n")
	builder.WriteString("- Press `gd` to go to definition (if available)\n")
	builder.WriteString("- Use `K` again for more detailed documentation\n")

	return builder.String()
}

// Helper functions
func getStringParam(args map[string]interface{}, key string) string {
	if val, ok := args[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntParam(args map[string]interface{}, key string) int {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}