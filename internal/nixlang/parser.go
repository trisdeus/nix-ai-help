package nixlang

import (
	"fmt"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// NixExpression represents a parsed Nix expression
type NixExpression struct {
	Type     NixExprType             `json:"type"`
	Value    interface{}             `json:"value,omitempty"`
	Children []NixExpression         `json:"children,omitempty"`
	Attrs    map[string]NixExpression `json:"attrs,omitempty"`
	Position Position                `json:"position,omitempty"`
	Metadata ExpressionMetadata      `json:"metadata,omitempty"`
}

// NixExprType represents different types of Nix expressions
type NixExprType string

const (
	ExprAttrSet    NixExprType = "attrset"
	ExprList       NixExprType = "list"
	ExprString     NixExprType = "string"
	ExprNumber     NixExprType = "number"
	ExprBool       NixExprType = "bool"
	ExprFunction   NixExprType = "function"
	ExprVariable   NixExprType = "variable"
	ExprWith       NixExprType = "with"
	ExprLet        NixExprType = "let"
	ExprIf         NixExprType = "if"
	ExprImport     NixExprType = "import"
	ExprCall       NixExprType = "call"
	ExprPath       NixExprType = "path"
	ExprNull       NixExprType = "null"
	ExprInterpolation NixExprType = "interpolation"
)

// Position represents a position in the source code
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"`
}

// ExpressionMetadata contains additional information about expressions
type ExpressionMetadata struct {
	Intent       string            `json:"intent,omitempty"`
	Complexity   int               `json:"complexity"`
	Dependencies []string          `json:"dependencies,omitempty"`
	SecurityTags []string          `json:"security_tags,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// NixParser provides comprehensive Nix expression parsing
type NixParser struct {
	source   string
	tokens   []Token
	position int
	fileSet  *token.FileSet
}

// Token represents a Nix language token
type Token struct {
	Type     TokenType `json:"type"`
	Value    string    `json:"value"`
	Position Position  `json:"position"`
}

// TokenType represents different types of Nix tokens
type TokenType string

const (
	TokenLBrace      TokenType = "{"
	TokenRBrace      TokenType = "}"
	TokenLBracket    TokenType = "["
	TokenRBracket    TokenType = "]"
	TokenLParen      TokenType = "("
	TokenRParen      TokenType = ")"
	TokenSemicolon   TokenType = ";"
	TokenColon       TokenType = ":"
	TokenComma       TokenType = ","
	TokenDot         TokenType = "."
	TokenEquals      TokenType = "="
	TokenString      TokenType = "string"
	TokenNumber      TokenType = "number"
	TokenPath        TokenType = "path"
	TokenIdentifier  TokenType = "identifier"
	TokenKeyword     TokenType = "keyword"
	TokenOperator    TokenType = "operator"
	TokenComment     TokenType = "comment"
	TokenWhitespace  TokenType = "whitespace"
	TokenEOF         TokenType = "eof"
	TokenInterpolation TokenType = "interpolation"
)

// NewNixParser creates a new Nix parser
func NewNixParser() *NixParser {
	return &NixParser{
		fileSet: token.NewFileSet(),
	}
}

// ParseExpression parses a Nix expression from source code
func (p *NixParser) ParseExpression(source string) (*NixExpression, error) {
	p.source = source
	p.position = 0
	
	// Tokenize the source
	tokens, err := p.tokenize(source)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	p.tokens = tokens
	
	// Parse the expression
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}
	
	// Add semantic analysis
	p.analyzeSemantics(expr)
	
	return expr, nil
}

// tokenize breaks the source into tokens
func (p *NixParser) tokenize(source string) ([]Token, error) {
	var tokens []Token
	pos := 0
	line := 1
	column := 1
	
	// Nix language patterns
	patterns := map[TokenType]*regexp.Regexp{
		TokenString:      regexp.MustCompile(`^"([^"\\]|\\.)*"`),
		TokenNumber:      regexp.MustCompile(`^-?\d+(\.\d+)?`),
		TokenPath:        regexp.MustCompile(`^[./][a-zA-Z0-9._/-]*`),
		TokenIdentifier:  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*`),
		TokenComment:     regexp.MustCompile(`^#[^\n]*`),
		TokenWhitespace:  regexp.MustCompile(`^\s+`),
		TokenInterpolation: regexp.MustCompile(`^\$\{[^}]*\}`),
	}
	
	// Nix keywords
	keywords := map[string]bool{
		"let": true, "in": true, "with": true, "inherit": true,
		"if": true, "then": true, "else": true, "assert": true,
		"import": true, "builtins": true, "derivation": true,
		"true": true, "false": true, "null": true, "or": true,
	}
	
	for pos < len(source) {
		matched := false
		
		// Check single character tokens first
		switch source[pos] {
		case '{':
			tokens = append(tokens, Token{TokenLBrace, "{", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case '}':
			tokens = append(tokens, Token{TokenRBrace, "}", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case '[':
			tokens = append(tokens, Token{TokenLBracket, "[", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case ']':
			tokens = append(tokens, Token{TokenRBracket, "]", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case '(':
			tokens = append(tokens, Token{TokenLParen, "(", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case ')':
			tokens = append(tokens, Token{TokenRParen, ")", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case ';':
			tokens = append(tokens, Token{TokenSemicolon, ";", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case ':':
			tokens = append(tokens, Token{TokenColon, ":", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case ',':
			tokens = append(tokens, Token{TokenComma, ",", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case '.':
			tokens = append(tokens, Token{TokenDot, ".", Position{line, column, pos}})
			pos++
			column++
			matched = true
		case '=':
			tokens = append(tokens, Token{TokenEquals, "=", Position{line, column, pos}})
			pos++
			column++
			matched = true
		}
		
		if matched {
			continue
		}
		
		// Check pattern-based tokens
		remaining := source[pos:]
		for tokenType, pattern := range patterns {
			if match := pattern.FindString(remaining); match != "" {
				tokenPos := Position{line, column, pos}
				
				// Handle whitespace specially
				if tokenType == TokenWhitespace {
					for _, char := range match {
						if char == '\n' {
							line++
							column = 1
						} else {
							column++
						}
					}
					pos += len(match)
					matched = true
					break
				}
				
				// Check if identifier is actually a keyword
				if tokenType == TokenIdentifier && keywords[match] {
					tokenType = TokenKeyword
				}
				
				tokens = append(tokens, Token{tokenType, match, tokenPos})
				pos += len(match)
				column += len(match)
				matched = true
				break
			}
		}
		
		if !matched {
			return nil, fmt.Errorf("unexpected character at line %d, column %d: %c", line, column, source[pos])
		}
	}
	
	// Add EOF token
	tokens = append(tokens, Token{TokenEOF, "", Position{line, column, pos}})
	return tokens, nil
}

// parseExpr parses a complete Nix expression
func (p *NixParser) parseExpr() (*NixExpression, error) {
	// Skip whitespace and comments
	p.skipNonSignificant()
	
	if p.position >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}
	
	token := p.tokens[p.position]
	
	switch token.Type {
	case TokenLBrace:
		return p.parseAttrSet()
	case TokenLBracket:
		return p.parseList()
	case TokenString:
		return p.parseString()
	case TokenNumber:
		return p.parseNumber()
	case TokenKeyword:
		return p.parseKeyword()
	case TokenIdentifier:
		return p.parseIdentifier()
	case TokenPath:
		return p.parsePath()
	case TokenLParen:
		return p.parseParenthesized()
	default:
		return nil, fmt.Errorf("unexpected token: %s at line %d", token.Value, token.Position.Line)
	}
}

// parseAttrSet parses a Nix attribute set
func (p *NixParser) parseAttrSet() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenLBrace {
		return nil, fmt.Errorf("expected '{' at position %d", p.position)
	}
	
	expr := &NixExpression{
		Type:  ExprAttrSet,
		Attrs: make(map[string]NixExpression),
		Position: Position{
			Line:   p.tokens[p.position].Position.Line,
			Column: p.tokens[p.position].Position.Column,
		},
	}
	
	p.position++ // consume '{'
	p.skipNonSignificant()
	
	for p.position < len(p.tokens) && p.tokens[p.position].Type != TokenRBrace {
		// Parse attribute name (can be a path like services.nginx.enable)
		attrPath := []string{}
		
		for {
			if p.tokens[p.position].Type != TokenIdentifier && p.tokens[p.position].Type != TokenString {
				return nil, fmt.Errorf("expected attribute name at position %d", p.position)
			}
			
			attrName := p.tokens[p.position].Value
			if p.tokens[p.position].Type == TokenString {
				// Remove quotes from string attribute names
				attrName = strings.Trim(attrName, "\"")
			}
			attrPath = append(attrPath, attrName)
			
			p.position++
			p.skipNonSignificant()
			
			// Check for dot to continue path
			if p.position < len(p.tokens) && p.tokens[p.position].Type == TokenDot {
				p.position++ // consume '.'
				p.skipNonSignificant()
			} else {
				break
			}
		}
		
		// Expect '='
		if p.position >= len(p.tokens) || p.tokens[p.position].Type != TokenEquals {
			return nil, fmt.Errorf("expected '=' after attribute name at position %d", p.position)
		}
		
		p.position++ // consume '='
		p.skipNonSignificant()
		
		// Parse attribute value
		value, err := p.parseExpr()
		if err != nil {
			return nil, fmt.Errorf("failed to parse attribute value: %w", err)
		}
		
		// Store the attribute (for now, use the full path as key)
		fullPath := strings.Join(attrPath, ".")
		expr.Attrs[fullPath] = *value
		
		p.skipNonSignificant()
		
		// Check for semicolon or comma (optional)
		if p.position < len(p.tokens) && 
		   (p.tokens[p.position].Type == TokenSemicolon || p.tokens[p.position].Type == TokenComma) {
			p.position++
			p.skipNonSignificant()
		}
	}
	
	if p.position >= len(p.tokens) || p.tokens[p.position].Type != TokenRBrace {
		return nil, fmt.Errorf("expected '}' to close attribute set")
	}
	
	p.position++ // consume '}'
	return expr, nil
}

// parseList parses a Nix list
func (p *NixParser) parseList() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenLBracket {
		return nil, fmt.Errorf("expected '[' at position %d", p.position)
	}
	
	expr := &NixExpression{
		Type: ExprList,
		Position: Position{
			Line:   p.tokens[p.position].Position.Line,
			Column: p.tokens[p.position].Position.Column,
		},
	}
	
	p.position++ // consume '['
	p.skipNonSignificant()
	
	for p.position < len(p.tokens) && p.tokens[p.position].Type != TokenRBracket {
		element, err := p.parseExpr()
		if err != nil {
			return nil, fmt.Errorf("failed to parse list element: %w", err)
		}
		
		expr.Children = append(expr.Children, *element)
		p.skipNonSignificant()
		
		// Optional comma
		if p.position < len(p.tokens) && p.tokens[p.position].Type == TokenComma {
			p.position++
			p.skipNonSignificant()
		}
	}
	
	if p.position >= len(p.tokens) || p.tokens[p.position].Type != TokenRBracket {
		return nil, fmt.Errorf("expected ']' to close list")
	}
	
	p.position++ // consume ']'
	return expr, nil
}

// parseString parses a Nix string
func (p *NixParser) parseString() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenString {
		return nil, fmt.Errorf("expected string at position %d", p.position)
	}
	
	token := p.tokens[p.position]
	value := strings.Trim(token.Value, "\"")
	
	expr := &NixExpression{
		Type:  ExprString,
		Value: value,
		Position: Position{
			Line:   token.Position.Line,
			Column: token.Position.Column,
		},
	}
	
	p.position++
	return expr, nil
}

// parseNumber parses a Nix number
func (p *NixParser) parseNumber() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenNumber {
		return nil, fmt.Errorf("expected number at position %d", p.position)
	}
	
	token := p.tokens[p.position]
	
	// Try to parse as int first, then float
	var value interface{}
	if intVal, err := strconv.Atoi(token.Value); err == nil {
		value = intVal
	} else if floatVal, err := strconv.ParseFloat(token.Value, 64); err == nil {
		value = floatVal
	} else {
		return nil, fmt.Errorf("invalid number format: %s", token.Value)
	}
	
	expr := &NixExpression{
		Type:  ExprNumber,
		Value: value,
		Position: Position{
			Line:   token.Position.Line,
			Column: token.Position.Column,
		},
	}
	
	p.position++
	return expr, nil
}

// parseKeyword parses Nix keywords
func (p *NixParser) parseKeyword() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenKeyword {
		return nil, fmt.Errorf("expected keyword at position %d", p.position)
	}
	
	token := p.tokens[p.position]
	
	switch token.Value {
	case "true":
		p.position++
		return &NixExpression{
			Type:  ExprBool,
			Value: true,
			Position: Position{Line: token.Position.Line, Column: token.Position.Column},
		}, nil
	case "false":
		p.position++
		return &NixExpression{
			Type:  ExprBool,
			Value: false,
			Position: Position{Line: token.Position.Line, Column: token.Position.Column},
		}, nil
	case "null":
		p.position++
		return &NixExpression{
			Type:  ExprNull,
			Value: nil,
			Position: Position{Line: token.Position.Line, Column: token.Position.Column},
		}, nil
	case "let":
		return p.parseLetExpression()
	case "with":
		return p.parseWithExpression()
	case "if":
		return p.parseIfExpression()
	case "import":
		return p.parseImportExpression()
	default:
		return nil, fmt.Errorf("unsupported keyword: %s", token.Value)
	}
}

// parseIdentifier parses a Nix identifier/variable
func (p *NixParser) parseIdentifier() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenIdentifier {
		return nil, fmt.Errorf("expected identifier at position %d", p.position)
	}
	
	token := p.tokens[p.position]
	
	expr := &NixExpression{
		Type:  ExprVariable,
		Value: token.Value,
		Position: Position{
			Line:   token.Position.Line,
			Column: token.Position.Column,
		},
	}
	
	p.position++
	return expr, nil
}

// parsePath parses a Nix path
func (p *NixParser) parsePath() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenPath {
		return nil, fmt.Errorf("expected path at position %d", p.position)
	}
	
	token := p.tokens[p.position]
	
	expr := &NixExpression{
		Type:  ExprPath,
		Value: token.Value,
		Position: Position{
			Line:   token.Position.Line,
			Column: token.Position.Column,
		},
	}
	
	p.position++
	return expr, nil
}

// parseParenthesized parses parenthesized expressions
func (p *NixParser) parseParenthesized() (*NixExpression, error) {
	if p.tokens[p.position].Type != TokenLParen {
		return nil, fmt.Errorf("expected '(' at position %d", p.position)
	}
	
	p.position++ // consume '('
	p.skipNonSignificant()
	
	expr, err := p.parseExpr()
	if err != nil {
		return nil, fmt.Errorf("failed to parse parenthesized expression: %w", err)
	}
	
	p.skipNonSignificant()
	
	if p.position >= len(p.tokens) || p.tokens[p.position].Type != TokenRParen {
		return nil, fmt.Errorf("expected ')' to close parenthesized expression")
	}
	
	p.position++ // consume ')'
	return expr, nil
}

// parseLetExpression parses let-in expressions
func (p *NixParser) parseLetExpression() (*NixExpression, error) {
	// Simplified let parsing - would need full implementation
	expr := &NixExpression{
		Type: ExprLet,
		Position: Position{
			Line:   p.tokens[p.position].Position.Line,
			Column: p.tokens[p.position].Position.Column,
		},
	}
	
	p.position++ // consume 'let'
	// TODO: Implement full let-in parsing
	return expr, nil
}

// parseWithExpression parses with expressions
func (p *NixParser) parseWithExpression() (*NixExpression, error) {
	// Simplified with parsing - would need full implementation
	expr := &NixExpression{
		Type: ExprWith,
		Position: Position{
			Line:   p.tokens[p.position].Position.Line,
			Column: p.tokens[p.position].Position.Column,
		},
	}
	
	p.position++ // consume 'with'
	// TODO: Implement full with parsing
	return expr, nil
}

// parseIfExpression parses if-then-else expressions
func (p *NixParser) parseIfExpression() (*NixExpression, error) {
	// Simplified if parsing - would need full implementation
	expr := &NixExpression{
		Type: ExprIf,
		Position: Position{
			Line:   p.tokens[p.position].Position.Line,
			Column: p.tokens[p.position].Position.Column,
		},
	}
	
	p.position++ // consume 'if'
	// TODO: Implement full if-then-else parsing
	return expr, nil
}

// parseImportExpression parses import expressions
func (p *NixParser) parseImportExpression() (*NixExpression, error) {
	expr := &NixExpression{
		Type: ExprImport,
		Position: Position{
			Line:   p.tokens[p.position].Position.Line,
			Column: p.tokens[p.position].Position.Column,
		},
	}
	
	p.position++ // consume 'import'
	p.skipNonSignificant()
	
	// Parse the import path
	if p.position < len(p.tokens) {
		path, err := p.parseExpr()
		if err != nil {
			return nil, fmt.Errorf("failed to parse import path: %w", err)
		}
		expr.Children = []NixExpression{*path}
	}
	
	return expr, nil
}

// skipNonSignificant skips whitespace and comment tokens
func (p *NixParser) skipNonSignificant() {
	for p.position < len(p.tokens) {
		token := p.tokens[p.position]
		if token.Type == TokenWhitespace || token.Type == TokenComment {
			p.position++
		} else {
			break
		}
	}
}

// analyzeSemantics adds semantic analysis to the parsed expression
func (p *NixParser) analyzeSemantics(expr *NixExpression) {
	if expr == nil {
		return
	}
	
	// Initialize metadata
	expr.Metadata = ExpressionMetadata{
		Complexity:   1,
		Dependencies: []string{},
		SecurityTags: []string{},
		Annotations:  make(map[string]string),
	}
	
	// Analyze based on expression type
	switch expr.Type {
	case ExprAttrSet:
		p.analyzeAttrSetSemantics(expr)
	case ExprList:
		p.analyzeListSemantics(expr)
	case ExprString:
		p.analyzeStringSemantics(expr)
	case ExprVariable:
		p.analyzeVariableSemantics(expr)
	case ExprImport:
		p.analyzeImportSemantics(expr)
	}
	
	// Recursively analyze children
	for i := range expr.Children {
		p.analyzeSemantics(&expr.Children[i])
		expr.Metadata.Complexity += expr.Children[i].Metadata.Complexity
	}
	
	// Recursively analyze attributes
	for key, attr := range expr.Attrs {
		p.analyzeSemantics(&attr)
		expr.Attrs[key] = attr
		expr.Metadata.Complexity += attr.Metadata.Complexity
	}
}

// analyzeAttrSetSemantics analyzes attribute set semantics
func (p *NixParser) analyzeAttrSetSemantics(expr *NixExpression) {
	// Detect common NixOS patterns
	if _, hasServices := expr.Attrs["services"]; hasServices {
		expr.Metadata.Intent = "service_configuration"
		expr.Metadata.Annotations["nixos_section"] = "services"
	}
	
	if _, hasEnvironment := expr.Attrs["environment"]; hasEnvironment {
		expr.Metadata.Intent = "environment_configuration"
		expr.Metadata.Annotations["nixos_section"] = "environment"
	}
	
	if _, hasSecurity := expr.Attrs["security"]; hasSecurity {
		expr.Metadata.SecurityTags = append(expr.Metadata.SecurityTags, "security_config")
		expr.Metadata.Annotations["security_relevant"] = "true"
	}
	
	// Check for networking configuration
	if _, hasNetworking := expr.Attrs["networking"]; hasNetworking {
		expr.Metadata.Intent = "network_configuration"
		expr.Metadata.SecurityTags = append(expr.Metadata.SecurityTags, "network_config")
	}
}

// analyzeListSemantics analyzes list semantics
func (p *NixParser) analyzeListSemantics(expr *NixExpression) {
	expr.Metadata.Complexity = len(expr.Children) + 1
	expr.Metadata.Intent = "list_collection"
}

// analyzeStringSemantics analyzes string semantics
func (p *NixParser) analyzeStringSemantics(expr *NixExpression) {
	if str, ok := expr.Value.(string); ok {
		// Check for security-sensitive patterns
		if strings.Contains(str, "password") || strings.Contains(str, "secret") {
			expr.Metadata.SecurityTags = append(expr.Metadata.SecurityTags, "sensitive_data")
		}
		
		// Check for file paths
		if strings.HasPrefix(str, "/") || strings.Contains(str, "./") {
			expr.Metadata.Annotations["type"] = "file_path"
			if strings.Contains(str, "/etc/") {
				expr.Metadata.SecurityTags = append(expr.Metadata.SecurityTags, "system_path")
			}
		}
		
		// Check for URLs
		if strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://") {
			expr.Metadata.Annotations["type"] = "url"
			if strings.HasPrefix(str, "http://") {
				expr.Metadata.SecurityTags = append(expr.Metadata.SecurityTags, "insecure_url")
			}
		}
	}
}

// analyzeVariableSemantics analyzes variable semantics
func (p *NixParser) analyzeVariableSemantics(expr *NixExpression) {
	if varName, ok := expr.Value.(string); ok {
		expr.Metadata.Dependencies = append(expr.Metadata.Dependencies, varName)
		
		// Check for common NixOS variables
		if varName == "pkgs" {
			expr.Metadata.Annotations["nixos_builtin"] = "packages"
		} else if varName == "config" {
			expr.Metadata.Annotations["nixos_builtin"] = "configuration"
		} else if varName == "lib" {
			expr.Metadata.Annotations["nixos_builtin"] = "library"
		}
	}
}

// analyzeImportSemantics analyzes import semantics
func (p *NixParser) analyzeImportSemantics(expr *NixExpression) {
	expr.Metadata.Intent = "module_import"
	expr.Metadata.SecurityTags = append(expr.Metadata.SecurityTags, "external_dependency")
	
	if len(expr.Children) > 0 {
		if pathExpr := &expr.Children[0]; pathExpr.Type == ExprString {
			if path, ok := pathExpr.Value.(string); ok {
				expr.Metadata.Dependencies = append(expr.Metadata.Dependencies, path)
				
				// Check for channel imports
				if strings.Contains(path, "<nixpkgs>") {
					expr.Metadata.Annotations["import_type"] = "nixpkgs_channel"
				} else if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
					expr.Metadata.Annotations["import_type"] = "relative_path"
				} else if strings.HasPrefix(path, "/") {
					expr.Metadata.Annotations["import_type"] = "absolute_path"
				}
			}
		}
	}
}

// GetComplexity returns the total complexity of the expression
func (expr *NixExpression) GetComplexity() int {
	return expr.Metadata.Complexity
}

// GetDependencies returns all dependencies found in the expression
func (expr *NixExpression) GetDependencies() []string {
	deps := make(map[string]bool)
	p := &NixParser{}
	p.collectDependencies(expr, deps)
	
	var result []string
	for dep := range deps {
		result = append(result, dep)
	}
	return result
}

// collectDependencies recursively collects all dependencies
func (p *NixParser) collectDependencies(expr *NixExpression, deps map[string]bool) {
	for _, dep := range expr.Metadata.Dependencies {
		deps[dep] = true
	}
	
	for _, child := range expr.Children {
		p.collectDependencies(&child, deps)
	}
	
	for _, attr := range expr.Attrs {
		p.collectDependencies(&attr, deps)
	}
}

// GetSecurityTags returns all security tags found in the expression
func (expr *NixExpression) GetSecurityTags() []string {
	tags := make(map[string]bool)
	p := &NixParser{}
	p.collectSecurityTags(expr, tags)
	
	var result []string
	for tag := range tags {
		result = append(result, tag)
	}
	return result
}

// collectSecurityTags recursively collects all security tags
func (p *NixParser) collectSecurityTags(expr *NixExpression, tags map[string]bool) {
	for _, tag := range expr.Metadata.SecurityTags {
		tags[tag] = true
	}
	
	for _, child := range expr.Children {
		p.collectSecurityTags(&child, tags)
	}
	
	for _, attr := range expr.Attrs {
		p.collectSecurityTags(&attr, tags)
	}
}