package nixlang

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// NixIntelligenceService provides comprehensive Nix language intelligence
type NixIntelligenceService struct {
	parser    *NixParser
	analyzer  *NixAnalyzer
	cache     *IntelligenceCache
	logger    *logger.Logger
	config    *config.UserConfig
}

// IntelligenceCache caches analysis results for performance
type IntelligenceCache struct {
	entries map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
}

// CacheEntry represents a cached analysis result
type CacheEntry struct {
	Result    *AnalysisResult `json:"result"`
	Timestamp time.Time       `json:"timestamp"`
	Hash      string          `json:"hash"`
}

// CompletionItem represents a code completion suggestion
type CompletionItem struct {
	Label         string            `json:"label"`
	Kind          CompletionKind    `json:"kind"`
	Detail        string            `json:"detail"`
	Documentation string            `json:"documentation"`
	InsertText    string            `json:"insertText"`
	Position      Position          `json:"position"`
	Confidence    float64           `json:"confidence"`
	Context       CompletionContext `json:"context"`
}

// CompletionKind represents different types of completions
type CompletionKind string

const (
	CompletionFunction    CompletionKind = "function"
	CompletionVariable    CompletionKind = "variable"
	CompletionProperty    CompletionKind = "property"
	CompletionKeyword     CompletionKind = "keyword"
	CompletionSnippet     CompletionKind = "snippet"
	CompletionModule      CompletionKind = "module"
	CompletionPackage     CompletionKind = "package"
	CompletionService     CompletionKind = "service"
	CompletionOption      CompletionKind = "option"
)

// CompletionContext provides context for completions
type CompletionContext struct {
	InAttributeSet bool     `json:"inAttributeSet"`
	InFunction     bool     `json:"inFunction"`
	InList         bool     `json:"inList"`
	InString       bool     `json:"inString"`
	ParentPath     []string `json:"parentPath"`
	NixOSContext   bool     `json:"nixosContext"`
}

// LanguageHint provides intelligent hints and suggestions
type LanguageHint struct {
	Type        HintType  `json:"type"`
	Message     string    `json:"message"`
	Position    Position  `json:"position"`
	Severity    Severity  `json:"severity"`
	Code        string    `json:"code,omitempty"`
	Actions     []Action  `json:"actions,omitempty"`
	Related     []string  `json:"related,omitempty"`
}

// HintType represents different types of hints
type HintType string

const (
	HintSyntaxError    HintType = "syntax_error"
	HintTypeError      HintType = "type_error"
	HintDeprecation    HintType = "deprecation"
	HintPerformance    HintType = "performance"
	HintSecurity       HintType = "security"
	HintBestPractice   HintType = "best_practice"
	HintRefactoring    HintType = "refactoring"
)

// Action represents a suggested action
type Action struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Edit        TextEdit  `json:"edit,omitempty"`
	Command     string    `json:"command,omitempty"`
}

// TextEdit represents a text edit operation
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// Range represents a text range
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// SymbolInformation represents information about a symbol
type SymbolInformation struct {
	Name         string         `json:"name"`
	Kind         SymbolKind     `json:"kind"`
	Position     Position       `json:"position"`
	Range        Range          `json:"range"`
	Detail       string         `json:"detail"`
	Documentation string        `json:"documentation"`
	Children     []SymbolInformation `json:"children,omitempty"`
}

// SymbolKind represents different kinds of symbols
type SymbolKind string

const (
	SymbolFile         SymbolKind = "file"
	SymbolModule       SymbolKind = "module"
	SymbolNamespace    SymbolKind = "namespace"
	SymbolPackage      SymbolKind = "package"
	SymbolClass        SymbolKind = "class"
	SymbolMethod       SymbolKind = "method"
	SymbolProperty     SymbolKind = "property"
	SymbolField        SymbolKind = "field"
	SymbolConstructor  SymbolKind = "constructor"
	SymbolEnum         SymbolKind = "enum"
	SymbolInterface    SymbolKind = "interface"
	SymbolFunction     SymbolKind = "function"
	SymbolVariable     SymbolKind = "variable"
	SymbolConstant     SymbolKind = "constant"
	SymbolString       SymbolKind = "string"
	SymbolNumber       SymbolKind = "number"
	SymbolBoolean      SymbolKind = "boolean"
	SymbolArray        SymbolKind = "array"
	SymbolObject       SymbolKind = "object"
	SymbolKey          SymbolKind = "key"
	SymbolNull         SymbolKind = "null"
)

// DiagnosticSeverity represents diagnostic severity levels
type DiagnosticSeverity int

const (
	DiagnosticError       DiagnosticSeverity = 1
	DiagnosticWarning     DiagnosticSeverity = 2
	DiagnosticInformation DiagnosticSeverity = 3
	DiagnosticHint        DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic message
type Diagnostic struct {
	Range       Range              `json:"range"`
	Severity    DiagnosticSeverity `json:"severity"`
	Code        string             `json:"code,omitempty"`
	Source      string             `json:"source"`
	Message     string             `json:"message"`
	RelatedInfo []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

// DiagnosticRelatedInformation represents related diagnostic information
type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// Location represents a location in a document
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// NewNixIntelligenceService creates a new Nix intelligence service
func NewNixIntelligenceService(cfg *config.UserConfig, log *logger.Logger) *NixIntelligenceService {
	return &NixIntelligenceService{
		parser:   NewNixParser(),
		analyzer: NewNixAnalyzer(),
		cache: &IntelligenceCache{
			entries: make(map[string]*CacheEntry),
			maxSize: 1000,
			ttl:     time.Hour,
		},
		logger: log,
		config: cfg,
	}
}

// AnalyzeDocument performs comprehensive analysis of a Nix document
func (nis *NixIntelligenceService) AnalyzeDocument(ctx context.Context, source string, uri string) (*AnalysisResult, error) {
	// Check cache first
	hash := nis.hashSource(source)
	if cached := nis.cache.get(hash); cached != nil {
		nis.logger.Debug("Using cached analysis result")
		return cached.Result, nil
	}
	
	// Perform analysis
	result, err := nis.analyzer.AnalyzeExpression(source)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}
	
	// Enhance with NixOS-specific insights
	nis.enhanceWithNixOSInsights(result, source)
	
	// Cache the result
	nis.cache.set(hash, result)
	
	nis.logger.Info(fmt.Sprintf("Analysis complete for %s: %d issues, %d suggestions", 
		uri, len(result.Issues), len(result.Suggestions)))
	
	return result, nil
}

// GetCompletions provides intelligent code completions
func (nis *NixIntelligenceService) GetCompletions(ctx context.Context, source string, position Position) ([]CompletionItem, error) {
	var completions []CompletionItem
	
	// Parse the current context
	context := nis.analyzeCompletionContext(source, position)
	
	// Get completions based on context
	if context.NixOSContext {
		completions = append(completions, nis.getNixOSCompletions(context)...)
	}
	
	if context.InAttributeSet {
		completions = append(completions, nis.getAttributeCompletions(context)...)
	}
	
	if context.InFunction {
		completions = append(completions, nis.getFunctionCompletions(context)...)
	}
	
	// Add general Nix language completions
	completions = append(completions, nis.getGeneralCompletions(context)...)
	
	// Add package completions
	completions = append(completions, nis.getPackageCompletions(context)...)
	
	return completions, nil
}

// GetHints provides intelligent hints and suggestions
func (nis *NixIntelligenceService) GetHints(ctx context.Context, source string) ([]LanguageHint, error) {
	var hints []LanguageHint
	
	// Analyze the document
	result, err := nis.AnalyzeDocument(ctx, source, "")
	if err != nil {
		return nil, fmt.Errorf("failed to analyze document for hints: %w", err)
	}
	
	// Convert issues to hints
	for _, issue := range result.Issues {
		hint := LanguageHint{
			Type:     nis.issueTypeToHintType(issue.Type),
			Message:  issue.Message,
			Position: issue.Position,
			Severity: issue.Severity,
			Code:     issue.Code,
		}
		
		if issue.Suggestion != "" {
			hint.Actions = []Action{
				{
					Title:       "Apply suggestion",
					Description: issue.Suggestion,
				},
			}
		}
		
		hints = append(hints, hint)
	}
	
	// Add security hints
	for _, finding := range result.SecurityFindings {
		hint := LanguageHint{
			Type:     HintSecurity,
			Message:  finding.Description,
			Position: finding.Position,
			Severity: finding.Severity,
			Actions: []Action{
				{
					Title:       "Learn more",
					Description: finding.Mitigation,
				},
			},
		}
		hints = append(hints, hint)
	}
	
	// Add optimization hints
	for _, opt := range result.Optimizations {
		hint := LanguageHint{
			Type:     HintPerformance,
			Message:  opt.Description,
			Position: Position{}, // Would need to track positions in optimizations
			Severity: SeverityInfo,
			Actions: []Action{
				{
					Title:       "Apply optimization",
					Description: opt.Impact,
				},
			},
		}
		hints = append(hints, hint)
	}
	
	return hints, nil
}

// GetSymbols extracts symbol information from the document
func (nis *NixIntelligenceService) GetSymbols(ctx context.Context, source string) ([]SymbolInformation, error) {
	expr, err := nis.parser.ParseExpression(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse for symbols: %w", err)
	}
	
	var symbols []SymbolInformation
	nis.extractSymbols(expr, &symbols, []string{})
	
	return symbols, nil
}

// GetDiagnostics provides diagnostic information
func (nis *NixIntelligenceService) GetDiagnostics(ctx context.Context, source string) ([]Diagnostic, error) {
	var diagnostics []Diagnostic
	
	// Analyze the document
	result, err := nis.AnalyzeDocument(ctx, source, "")
	if err != nil {
		// If parsing fails, return a diagnostic about the parse error
		diagnostic := Diagnostic{
			Range: Range{
				Start: Position{Line: 1, Column: 1},
				End:   Position{Line: 1, Column: 1},
			},
			Severity: DiagnosticError,
			Source:   "nixlang",
			Message:  fmt.Sprintf("Parse error: %v", err),
		}
		return []Diagnostic{diagnostic}, nil
	}
	
	// Convert issues to diagnostics
	for _, issue := range result.Issues {
		severity := nis.severityToDiagnosticSeverity(issue.Severity)
		
		diagnostic := Diagnostic{
			Range: Range{
				Start: issue.Position,
				End:   Position{
					Line:   issue.Position.Line,
					Column: issue.Position.Column + 1,
				},
			},
			Severity: severity,
			Code:     issue.Code,
			Source:   "nixlang",
			Message:  issue.Message,
		}
		
		diagnostics = append(diagnostics, diagnostic)
	}
	
	// Convert security findings to diagnostics
	for _, finding := range result.SecurityFindings {
		severity := nis.severityToDiagnosticSeverity(finding.Severity)
		
		diagnostic := Diagnostic{
			Range: Range{
				Start: finding.Position,
				End:   Position{
					Line:   finding.Position.Line,
					Column: finding.Position.Column + 1,
				},
			},
			Severity: severity,
			Code:     finding.Type,
			Source:   "nixlang-security",
			Message:  finding.Description,
		}
		
		diagnostics = append(diagnostics, diagnostic)
	}
	
	return diagnostics, nil
}

// ValidateConfiguration validates a NixOS configuration
func (nis *NixIntelligenceService) ValidateConfiguration(ctx context.Context, source string) (*ValidationResult, error) {
	result, err := nis.AnalyzeDocument(ctx, source, "")
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	validation := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Info:     []string{},
	}
	
	// Check for errors
	for _, issue := range result.Issues {
		switch issue.Severity {
		case SeverityError:
			validation.Valid = false
			validation.Errors = append(validation.Errors, issue.Message)
		case SeverityWarning:
			validation.Warnings = append(validation.Warnings, issue.Message)
		case SeverityInfo:
			validation.Info = append(validation.Info, issue.Message)
		}
	}
	
	// Check security findings
	for _, finding := range result.SecurityFindings {
		if finding.Severity == SeverityError {
			validation.Valid = false
			validation.Errors = append(validation.Errors, finding.Description)
		} else {
			validation.Warnings = append(validation.Warnings, finding.Description)
		}
	}
	
	return validation, nil
}

// ValidationResult represents configuration validation results
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Info     []string `json:"info"`
}

// Private helper methods

func (nis *NixIntelligenceService) hashSource(source string) string {
	// Simple hash implementation - in production, use a proper hash function
	return fmt.Sprintf("%x", len(source)+strings.Count(source, "\n"))
}

func (nis *NixIntelligenceService) enhanceWithNixOSInsights(result *AnalysisResult, source string) {
	// Add NixOS-specific insights
	if strings.Contains(source, "services.") {
		result.Intent.Annotations["nixos_services"] = "true"
	}
	
	if strings.Contains(source, "environment.systemPackages") {
		result.Intent.Annotations["system_packages"] = "true"
	}
	
	if strings.Contains(source, "users.users") {
		result.Intent.Annotations["user_management"] = "true"
	}
}

func (nis *NixIntelligenceService) analyzeCompletionContext(source string, position Position) CompletionContext {
	// Analyze the context around the cursor position
	context := CompletionContext{
		ParentPath: []string{},
	}
	
	// Simple context analysis - in production, this would be more sophisticated
	lines := strings.Split(source, "\n")
	if position.Line <= len(lines) {
		currentLine := lines[position.Line-1]
		
		context.InAttributeSet = strings.Contains(currentLine, "{") && !strings.Contains(currentLine, "}")
		context.InList = strings.Contains(currentLine, "[") && !strings.Contains(currentLine, "]")
		context.InString = strings.Count(currentLine[:position.Column], "\"")%2 == 1
		context.NixOSContext = strings.Contains(source, "services.") || strings.Contains(source, "environment.")
	}
	
	return context
}

func (nis *NixIntelligenceService) getNixOSCompletions(context CompletionContext) []CompletionItem {
	nixosCompletions := []struct {
		label         string
		kind          CompletionKind
		detail        string
		documentation string
		insertText    string
	}{
		{
			"services.nginx.enable",
			CompletionOption,
			"Enable Nginx web server",
			"Enables the Nginx HTTP server",
			"services.nginx.enable = true;",
		},
		{
			"services.postgresql.enable",
			CompletionOption,
			"Enable PostgreSQL database",
			"Enables the PostgreSQL database server",
			"services.postgresql.enable = true;",
		},
		{
			"environment.systemPackages",
			CompletionOption,
			"System packages",
			"List of packages available to all users",
			"environment.systemPackages = with pkgs; [ $0 ];",
		},
		{
			"networking.hostName",
			CompletionOption,
			"Host name",
			"The host name of the machine",
			"networking.hostName = \"$0\";",
		},
		{
			"users.users",
			CompletionOption,
			"User accounts",
			"User account configuration",
			"users.users.$0 = { isNormalUser = true; };",
		},
		{
			"boot.loader.systemd-boot.enable",
			CompletionOption,
			"Enable systemd-boot",
			"Enable the systemd-boot EFI boot loader",
			"boot.loader.systemd-boot.enable = true;",
		},
		{
			"hardware.bluetooth.enable",
			CompletionOption,
			"Enable Bluetooth",
			"Enable Bluetooth support",
			"hardware.bluetooth.enable = true;",
		},
		{
			"security.sudo.enable",
			CompletionOption,
			"Enable sudo",
			"Enable sudo for privilege escalation",
			"security.sudo.enable = true;",
		},
	}
	
	var completions []CompletionItem
	for _, comp := range nixosCompletions {
		completion := CompletionItem{
			Label:         comp.label,
			Kind:          comp.kind,
			Detail:        comp.detail,
			Documentation: comp.documentation,
			InsertText:    comp.insertText,
			Confidence:    0.9,
			Context:       context,
		}
		completions = append(completions, completion)
	}
	
	return completions
}

func (nis *NixIntelligenceService) getAttributeCompletions(context CompletionContext) []CompletionItem {
	// Common attribute completions
	attrs := []struct {
		label  string
		detail string
		insert string
	}{
		{"enable", "Enable this service/feature", "enable = true;"},
		{"package", "Package to use", "package = pkgs.$0;"},
		{"extraConfig", "Extra configuration", "extraConfig = \"$0\";"},
		{"user", "User to run as", "user = \"$0\";"},
		{"group", "Group to run as", "group = \"$0\";"},
		{"port", "Port number", "port = $0;"},
		{"host", "Host address", "host = \"$0\";"},
		{"dataDir", "Data directory", "dataDir = \"$0\";"},
		{"configFile", "Configuration file", "configFile = $0;"},
		{"description", "Description", "description = \"$0\";"},
	}
	
	var completions []CompletionItem
	for _, attr := range attrs {
		completion := CompletionItem{
			Label:      attr.label,
			Kind:       CompletionProperty,
			Detail:     attr.detail,
			InsertText: attr.insert,
			Confidence: 0.7,
			Context:    context,
		}
		completions = append(completions, completion)
	}
	
	return completions
}

func (nis *NixIntelligenceService) getFunctionCompletions(context CompletionContext) []CompletionItem {
	// Common Nix functions
	functions := []struct {
		label  string
		detail string
		insert string
		doc    string
	}{
		{"map", "Apply function to list", "map $0", "Apply a function to each element of a list"},
		{"filter", "Filter list elements", "filter $0", "Filter list elements based on a predicate"},
		{"foldl", "Fold left", "foldl $0", "Left fold over a list"},
		{"foldr", "Fold right", "foldr $0", "Right fold over a list"},
		{"length", "List length", "length $0", "Get the length of a list"},
		{"head", "First element", "head $0", "Get the first element of a list"},
		{"tail", "All but first element", "tail $0", "Get all but the first element of a list"},
		{"toString", "Convert to string", "toString $0", "Convert a value to a string"},
		{"toJSON", "Convert to JSON", "toJSON $0", "Convert a value to JSON"},
		{"fromJSON", "Parse JSON", "fromJSON $0", "Parse a JSON string"},
	}
	
	var completions []CompletionItem
	for _, fn := range functions {
		completion := CompletionItem{
			Label:         fn.label,
			Kind:          CompletionFunction,
			Detail:        fn.detail,
			Documentation: fn.doc,
			InsertText:    fn.insert,
			Confidence:    0.8,
			Context:       context,
		}
		completions = append(completions, completion)
	}
	
	return completions
}

func (nis *NixIntelligenceService) getGeneralCompletions(context CompletionContext) []CompletionItem {
	// General Nix language completions
	general := []struct {
		label  string
		kind   CompletionKind
		detail string
		insert string
	}{
		{"let", CompletionKeyword, "Let expression", "let\n  $0\nin"},
		{"in", CompletionKeyword, "In clause", "in"},
		{"with", CompletionKeyword, "With expression", "with $0;"},
		{"if", CompletionKeyword, "If expression", "if $0 then $1 else $2"},
		{"then", CompletionKeyword, "Then clause", "then"},
		{"else", CompletionKeyword, "Else clause", "else"},
		{"inherit", CompletionKeyword, "Inherit attributes", "inherit $0;"},
		{"import", CompletionKeyword, "Import module", "import $0"},
		{"true", CompletionKeyword, "Boolean true", "true"},
		{"false", CompletionKeyword, "Boolean false", "false"},
		{"null", CompletionKeyword, "Null value", "null"},
		{"pkgs", CompletionVariable, "Package set", "pkgs"},
		{"lib", CompletionVariable, "Nixpkgs library", "lib"},
		{"config", CompletionVariable, "Configuration", "config"},
		{"stdenv", CompletionVariable, "Standard environment", "stdenv"},
	}
	
	var completions []CompletionItem
	for _, comp := range general {
		completion := CompletionItem{
			Label:      comp.label,
			Kind:       comp.kind,
			Detail:     comp.detail,
			InsertText: comp.insert,
			Confidence: 0.6,
			Context:    context,
		}
		completions = append(completions, completion)
	}
	
	return completions
}

func (nis *NixIntelligenceService) getPackageCompletions(context CompletionContext) []CompletionItem {
	// Common packages - in production, this would query nixpkgs
	packages := []string{
		"git", "vim", "emacs", "firefox", "chromium", "code", "docker",
		"nodejs", "python3", "go", "rust", "gcc", "make", "cmake",
		"curl", "wget", "htop", "tmux", "zsh", "bash", "fish",
	}
	
	var completions []CompletionItem
	for _, pkg := range packages {
		completion := CompletionItem{
			Label:      pkg,
			Kind:       CompletionPackage,
			Detail:     fmt.Sprintf("Package: %s", pkg),
			InsertText: pkg,
			Confidence: 0.5,
			Context:    context,
		}
		completions = append(completions, completion)
	}
	
	return completions
}

func (nis *NixIntelligenceService) extractSymbols(expr *NixExpression, symbols *[]SymbolInformation, path []string) {
	if expr == nil {
		return
	}
	
	symbol := SymbolInformation{
		Name:     strings.Join(path, "."),
		Kind:     nis.exprTypeToSymbolKind(expr.Type),
		Position: expr.Position,
		Range: Range{
			Start: expr.Position,
			End:   expr.Position,
		},
	}
	
	switch expr.Type {
	case ExprAttrSet:
		symbol.Kind = SymbolObject
		for key, attr := range expr.Attrs {
			newPath := append(path, key)
			nis.extractSymbols(&attr, symbols, newPath)
		}
	case ExprFunction:
		symbol.Kind = SymbolFunction
	case ExprVariable:
		symbol.Kind = SymbolVariable
		if str, ok := expr.Value.(string); ok {
			symbol.Name = str
		}
	}
	
	if symbol.Name != "" {
		*symbols = append(*symbols, symbol)
	}
}

// Utility conversion methods

func (nis *NixIntelligenceService) issueTypeToHintType(issueType IssueType) HintType {
	switch issueType {
	case IssueSyntax:
		return HintSyntaxError
	case IssueSecurity:
		return HintSecurity
	case IssuePerformance:
		return HintPerformance
	case IssueAntiPattern:
		return HintRefactoring
	default:
		return HintBestPractice
	}
}

func (nis *NixIntelligenceService) severityToDiagnosticSeverity(severity Severity) DiagnosticSeverity {
	switch severity {
	case SeverityError:
		return DiagnosticError
	case SeverityWarning:
		return DiagnosticWarning
	case SeverityInfo:
		return DiagnosticInformation
	case SeverityHint:
		return DiagnosticHint
	default:
		return DiagnosticInformation
	}
}

func (nis *NixIntelligenceService) exprTypeToSymbolKind(exprType NixExprType) SymbolKind {
	switch exprType {
	case ExprAttrSet:
		return SymbolObject
	case ExprList:
		return SymbolArray
	case ExprString:
		return SymbolString
	case ExprNumber:
		return SymbolNumber
	case ExprBool:
		return SymbolBoolean
	case ExprFunction:
		return SymbolFunction
	case ExprVariable:
		return SymbolVariable
	case ExprNull:
		return SymbolNull
	default:
		return SymbolField
	}
}

// Cache methods

func (ic *IntelligenceCache) get(hash string) *CacheEntry {
	entry, exists := ic.entries[hash]
	if !exists {
		return nil
	}
	
	// Check if expired
	if time.Since(entry.Timestamp) > ic.ttl {
		delete(ic.entries, hash)
		return nil
	}
	
	return entry
}

func (ic *IntelligenceCache) set(hash string, result *AnalysisResult) {
	// Remove old entries if cache is full
	if len(ic.entries) >= ic.maxSize {
		// Simple LRU - remove oldest entry
		var oldestHash string
		var oldestTime time.Time
		for h, entry := range ic.entries {
			if oldestHash == "" || entry.Timestamp.Before(oldestTime) {
				oldestHash = h
				oldestTime = entry.Timestamp
			}
		}
		if oldestHash != "" {
			delete(ic.entries, oldestHash)
		}
	}
	
	ic.entries[hash] = &CacheEntry{
		Result:    result,
		Timestamp: time.Now(),
		Hash:      hash,
	}
}