package neovim

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/nixos"
	"nix-ai-help/pkg/logger"
)

// CompletionItem represents a completion suggestion for Neovim
type CompletionItem struct {
	Label         string            `json:"label"`
	Kind          CompletionKind    `json:"kind"`
	Detail        string            `json:"detail,omitempty"`
	Documentation string            `json:"documentation,omitempty"`
	InsertText    string            `json:"insertText,omitempty"`
	FilterText    string            `json:"filterText,omitempty"`
	SortText      string            `json:"sortText,omitempty"`
	Data          map[string]string `json:"data,omitempty"`
}

// CompletionKind represents the type of completion item
type CompletionKind int

const (
	KindText CompletionKind = iota + 1
	KindMethod
	KindFunction
	KindConstructor
	KindField
	KindVariable
	KindClass
	KindInterface
	KindModule
	KindProperty
	KindUnit
	KindValue
	KindEnum
	KindKeyword
	KindSnippet
	KindColor
	KindFile
	KindReference
	KindFolder
	KindEnumMember
	KindConstant
	KindStruct
	KindEvent
	KindOperator
	KindTypeParameter
)

// CompletionParams contains parameters for completion requests
type CompletionParams struct {
	FilePath   string `json:"filePath"`
	Line       int    `json:"line"`
	Character  int    `json:"character"`
	Context    string `json:"context"`
	Trigger    string `json:"trigger,omitempty"`
	LineText   string `json:"lineText"`
	BufferText string `json:"bufferText,omitempty"`
}

// CompletionProvider provides AI-powered completion for Neovim
type CompletionProvider struct {
	aiManager *ai.CLIProviderManager
	context   *nixos.ContextDetector
	logger    *logger.Logger
}

// NewCompletionProvider creates a new completion provider
func NewCompletionProvider(aiManager *ai.CLIProviderManager, context *nixos.ContextDetector, logger *logger.Logger) *CompletionProvider {
	return &CompletionProvider{
		aiManager: aiManager,
		context:   context,
		logger:    logger,
	}
}

// GetCompletions provides AI-powered completion suggestions
func (cp *CompletionProvider) GetCompletions(params CompletionParams) ([]CompletionItem, error) {
	// Create context for AI completion
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Detect completion type based on context
	completionType := cp.detectCompletionType(params)
	
	// Get system context for better suggestions
	systemContext := "Basic NixOS system"

	// Build AI prompt for completion
	prompt := cp.buildCompletionPrompt(params, completionType, systemContext)
	
	// Get AI suggestions
	response, err := ai.QuickQuery(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI completion: %w", err)
	}

	// Parse and format completion items
	items, err := cp.parseCompletionResponse(response, completionType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse completion response: %w", err)
	}

	return items, nil
}

// detectCompletionType determines what type of completion is needed
func (cp *CompletionProvider) detectCompletionType(params CompletionParams) string {
	line := strings.TrimSpace(params.LineText)
	
	// Service configuration
	if strings.Contains(line, "services.") {
		return "service_option"
	}
	
	// Package management
	if strings.Contains(line, "environment.systemPackages") || strings.Contains(line, "with pkgs;") {
		return "package_name"
	}
	
	// Hardware configuration
	if strings.Contains(line, "hardware.") {
		return "hardware_option"
	}
	
	// Networking configuration
	if strings.Contains(line, "networking.") {
		return "networking_option"
	}
	
	// Boot configuration
	if strings.Contains(line, "boot.") {
		return "boot_option"
	}
	
	// Home Manager configuration
	if strings.Contains(line, "programs.") || strings.Contains(line, "home.") {
		return "home_manager_option"
	}
	
	// General NixOS option
	if strings.Contains(params.FilePath, "configuration.nix") || strings.Contains(params.FilePath, ".nix") {
		return "nixos_option"
	}
	
	return "general"
}

// buildCompletionPrompt creates the AI prompt for completion
func (cp *CompletionProvider) buildCompletionPrompt(params CompletionParams, completionType, systemContext string) string {
	prompt := fmt.Sprintf(`You are an expert NixOS assistant providing code completion suggestions.

Context:
- File: %s
- Line %d, Character %d
- Current line: "%s"
- Completion type: %s
- System context: %s

Current buffer context:
%s

Provide 5-10 relevant completion suggestions for the current context. Format as JSON array with this structure:
[
  {
    "label": "suggestion text",
    "kind": "option|package|service|function|value",
    "detail": "brief description",
    "documentation": "detailed explanation with example",
    "insertText": "text to insert (if different from label)"
  }
]

Focus on:
1. NixOS options relevant to the current context
2. Popular packages for the detected category
3. Common configuration patterns
4. Best practices and examples
5. Context-aware suggestions based on system configuration

Ensure suggestions are:
- Accurate and up-to-date
- Relevant to the current line context
- Include proper Nix syntax
- Provide helpful documentation
`, params.FilePath, params.Line, params.Character, params.LineText, completionType, systemContext, cp.getTruncatedBuffer(params.BufferText))

	return prompt
}

// parseCompletionResponse parses AI response into completion items
func (cp *CompletionProvider) parseCompletionResponse(response, completionType string) ([]CompletionItem, error) {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")
	
	if jsonStart == -1 || jsonEnd == -1 {
		return cp.createFallbackCompletions(completionType), nil
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var rawItems []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawItems); err != nil {
		cp.logger.Warn("Failed to parse completion JSON, using fallback")
		return cp.createFallbackCompletions(completionType), nil
	}
	
	items := make([]CompletionItem, 0, len(rawItems))
	for i, raw := range rawItems {
		item := CompletionItem{
			Label:         cp.getStringField(raw, "label"),
			Detail:        cp.getStringField(raw, "detail"),
			Documentation: cp.getStringField(raw, "documentation"),
			InsertText:    cp.getStringField(raw, "insertText"),
			SortText:      fmt.Sprintf("%03d", i), // Maintain AI-suggested order
		}
		
		// Set completion kind based on type
		item.Kind = cp.getCompletionKind(cp.getStringField(raw, "kind"), completionType)
		
		// Use label as insertText if not specified
		if item.InsertText == "" {
			item.InsertText = item.Label
		}
		
		items = append(items, item)
	}
	
	return items, nil
}

// getStringField safely extracts string field from map
func (cp *CompletionProvider) getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getCompletionKind maps string kind to CompletionKind enum
func (cp *CompletionProvider) getCompletionKind(kindStr, completionType string) CompletionKind {
	switch kindStr {
	case "package":
		return KindModule
	case "service":
		return KindClass
	case "option":
		return KindProperty
	case "function":
		return KindFunction
	case "value":
		return KindValue
	default:
		// Default based on completion type
		switch completionType {
		case "package_name":
			return KindModule
		case "service_option":
			return KindClass
		case "nixos_option", "hardware_option", "networking_option":
			return KindProperty
		default:
			return KindText
		}
	}
}

// createFallbackCompletions provides basic completions when AI fails
func (cp *CompletionProvider) createFallbackCompletions(completionType string) []CompletionItem {
	switch completionType {
	case "service_option":
		return []CompletionItem{
			{Label: "enable", Kind: KindProperty, Detail: "Enable the service", InsertText: "enable = true;"},
			{Label: "package", Kind: KindProperty, Detail: "Service package", InsertText: "package = pkgs.${1:package};"},
			{Label: "extraConfig", Kind: KindProperty, Detail: "Extra configuration", InsertText: "extraConfig = \"${1:config}\";"},
		}
	case "package_name":
		return []CompletionItem{
			{Label: "git", Kind: KindModule, Detail: "Git version control", InsertText: "git"},
			{Label: "vim", Kind: KindModule, Detail: "Vim editor", InsertText: "vim"},
			{Label: "firefox", Kind: KindModule, Detail: "Firefox browser", InsertText: "firefox"},
		}
	case "nixos_option":
		return []CompletionItem{
			{Label: "services", Kind: KindProperty, Detail: "System services configuration", InsertText: "services."},
			{Label: "environment", Kind: KindProperty, Detail: "Environment configuration", InsertText: "environment."},
			{Label: "hardware", Kind: KindProperty, Detail: "Hardware configuration", InsertText: "hardware."},
		}
	default:
		return []CompletionItem{
			{Label: "# Add configuration here", Kind: KindText, Detail: "Comment placeholder", InsertText: "# ${1:configuration}"},
		}
	}
}

// getTruncatedBuffer returns a truncated version of buffer for context
func (cp *CompletionProvider) getTruncatedBuffer(buffer string) string {
	const maxLines = 20
	lines := strings.Split(buffer, "\n")
	if len(lines) > maxLines {
		// Take first 10 and last 10 lines
		result := strings.Join(lines[:10], "\n")
		result += "\n# ... (truncated) ...\n"
		result += strings.Join(lines[len(lines)-10:], "\n")
		return result
	}
	return buffer
}

// CompletionCache provides caching for completion results
type CompletionCache struct {
	cache map[string]CacheEntry
}

type CacheEntry struct {
	Items     []CompletionItem
	Timestamp time.Time
}

// NewCompletionCache creates a new completion cache
func NewCompletionCache() *CompletionCache {
	return &CompletionCache{
		cache: make(map[string]CacheEntry),
	}
}

// Get retrieves cached completion items
func (cc *CompletionCache) Get(key string) ([]CompletionItem, bool) {
	entry, exists := cc.cache[key]
	if !exists {
		return nil, false
	}
	
	// Cache valid for 5 minutes
	if time.Since(entry.Timestamp) > 5*time.Minute {
		delete(cc.cache, key)
		return nil, false
	}
	
	return entry.Items, true
}

// Set stores completion items in cache
func (cc *CompletionCache) Set(key string, items []CompletionItem) {
	cc.cache[key] = CacheEntry{
		Items:     items,
		Timestamp: time.Now(),
	}
}

// GetCacheKey generates a cache key for completion params
func GetCacheKey(params CompletionParams) string {
	// Create a simple cache key based on file path and line context
	return fmt.Sprintf("%s:%d:%s", params.FilePath, params.Line, params.LineText)
}