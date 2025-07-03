package nixlang

import (
	"testing"
)

func TestNixParser_ParseExpression(t *testing.T) {
	parser := NewNixParser()
	
	tests := []struct {
		name     string
		source   string
		wantType NixExprType
		wantErr  bool
	}{
		{
			name:     "simple string",
			source:   `"hello world"`,
			wantType: ExprString,
			wantErr:  false,
		},
		{
			name:     "simple number",
			source:   `42`,
			wantType: ExprNumber,
			wantErr:  false,
		},
		{
			name:     "boolean true",
			source:   `true`,
			wantType: ExprBool,
			wantErr:  false,
		},
		{
			name:     "boolean false",
			source:   `false`,
			wantType: ExprBool,
			wantErr:  false,
		},
		{
			name:     "null value",
			source:   `null`,
			wantType: ExprNull,
			wantErr:  false,
		},
		{
			name:     "simple variable",
			source:   `pkgs`,
			wantType: ExprVariable,
			wantErr:  false,
		},
		{
			name:     "empty attribute set",
			source:   `{}`,
			wantType: ExprAttrSet,
			wantErr:  false,
		},
		{
			name:     "simple attribute set",
			source:   `{ name = "value"; }`,
			wantType: ExprAttrSet,
			wantErr:  false,
		},
		{
			name:     "empty list",
			source:   `[]`,
			wantType: ExprList,
			wantErr:  false,
		},
		{
			name:     "simple list",
			source:   `[ "a" "b" "c" ]`,
			wantType: ExprList,
			wantErr:  false,
		},
		{
			name:     "nested attribute set",
			source:   `{ services = { nginx = { enable = true; }; }; }`,
			wantType: ExprAttrSet,
			wantErr:  false,
		},
		{
			name:     "import expression",
			source:   `import "config.nix"`,
			wantType: ExprImport,
			wantErr:  false,
		},
		{
			name:     "parenthesized expression",
			source:   `("hello")`,
			wantType: ExprString,
			wantErr:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.source)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseExpression() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if expr.Type != tt.wantType {
				t.Errorf("ParseExpression() type = %v, want %v", expr.Type, tt.wantType)
			}
		})
	}
}

func TestNixParser_ParseAttrSet(t *testing.T) {
	parser := NewNixParser()
	
	source := `{
		services.nginx.enable = true;
		environment.systemPackages = [ pkgs.git pkgs.vim ];
		networking.hostName = "myhost";
	}`
	
	expr, err := parser.ParseExpression(source)
	if err != nil {
		t.Fatalf("ParseExpression() error = %v", err)
	}
	
	if expr.Type != ExprAttrSet {
		t.Errorf("Expected ExprAttrSet, got %v", expr.Type)
	}
	
	// Check that attributes were parsed
	expectedAttrs := []string{"services", "environment", "networking"}
	for _, attr := range expectedAttrs {
		if _, exists := expr.Attrs[attr]; !exists {
			t.Errorf("Expected attribute %s not found", attr)
		}
	}
}

func TestNixParser_ComplexityAnalysis(t *testing.T) {
	parser := NewNixParser()
	
	tests := []struct {
		name           string
		source         string
		minComplexity  int
		maxComplexity  int
	}{
		{
			name:          "simple string",
			source:        `"hello"`,
			minComplexity: 1,
			maxComplexity: 2,
		},
		{
			name:          "simple attribute set",
			source:        `{ a = 1; b = 2; }`,
			minComplexity: 3,
			maxComplexity: 5,
		},
		{
			name: "complex nested structure",
			source: `{
				services = {
					nginx = {
						enable = true;
						virtualHosts = {
							"example.com" = {
								locations."/" = {
									proxyPass = "http://localhost:3000";
								};
							};
						};
					};
				};
			}`,
			minComplexity: 8,
			maxComplexity: 15,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.source)
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}
			
			complexity := expr.GetComplexity()
			if complexity < tt.minComplexity || complexity > tt.maxComplexity {
				t.Errorf("Complexity %d not in range [%d, %d]", 
					complexity, tt.minComplexity, tt.maxComplexity)
			}
		})
	}
}

func TestNixParser_DependencyExtraction(t *testing.T) {
	parser := NewNixParser()
	
	source := `{
		environment.systemPackages = with pkgs; [ git vim nodejs ];
		services.nginx.package = pkgs.nginx;
		users.defaultUserShell = pkgs.zsh;
	}`
	
	expr, err := parser.ParseExpression(source)
	if err != nil {
		t.Fatalf("ParseExpression() error = %v", err)
	}
	
	dependencies := expr.GetDependencies()
	expectedDeps := []string{"pkgs"}
	
	for _, expectedDep := range expectedDeps {
		found := false
		for _, dep := range dependencies {
			if dep == expectedDep {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected dependency %s not found in %v", expectedDep, dependencies)
		}
	}
}

func TestNixParser_SecurityTagging(t *testing.T) {
	parser := NewNixParser()
	
	tests := []struct {
		name         string
		source       string
		expectedTags []string
	}{
		{
			name:         "security configuration",
			source:       `{ security.sudo.enable = true; }`,
			expectedTags: []string{"security_config"},
		},
		{
			name:         "networking configuration",
			source:       `{ networking.firewall.enable = false; }`,
			expectedTags: []string{"network_config"},
		},
		{
			name:         "sensitive data",
			source:       `{ password = "secret123"; }`,
			expectedTags: []string{"sensitive_data"},
		},
		{
			name:         "insecure URL",
			source:       `{ url = "http://example.com"; }`,
			expectedTags: []string{"insecure_url"},
		},
		{
			name:         "system path",
			source:       `{ configFile = "/etc/nixos/configuration.nix"; }`,
			expectedTags: []string{"system_path"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.source)
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}
			
			securityTags := expr.GetSecurityTags()
			
			for _, expectedTag := range tt.expectedTags {
				found := false
				for _, tag := range securityTags {
					if tag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected security tag %s not found in %v", expectedTag, securityTags)
				}
			}
		})
	}
}

func TestNixParser_Tokenization(t *testing.T) {
	parser := NewNixParser()
	
	tests := []struct {
		name         string
		source       string
		expectedTokens []TokenType
	}{
		{
			name:   "simple attribute assignment",
			source: `name = "value";`,
			expectedTokens: []TokenType{
				TokenIdentifier, TokenEquals, TokenString, TokenSemicolon, TokenEOF,
			},
		},
		{
			name:   "attribute set",
			source: `{ a = 1; }`,
			expectedTokens: []TokenType{
				TokenLBrace, TokenIdentifier, TokenEquals, TokenNumber, TokenSemicolon, TokenRBrace, TokenEOF,
			},
		},
		{
			name:   "list with elements",
			source: `[ "a" "b" ]`,
			expectedTokens: []TokenType{
				TokenLBracket, TokenString, TokenString, TokenRBracket, TokenEOF,
			},
		},
		{
			name:   "keywords",
			source: `let in with if then else`,
			expectedTokens: []TokenType{
				TokenKeyword, TokenKeyword, TokenKeyword, TokenKeyword, TokenKeyword, TokenKeyword, TokenEOF,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := parser.tokenize(tt.source)
			if err != nil {
				t.Fatalf("tokenize() error = %v", err)
			}
			
			if len(tokens) != len(tt.expectedTokens) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expectedTokens), len(tokens))
				return
			}
			
			for i, expectedType := range tt.expectedTokens {
				if tokens[i].Type != expectedType {
					t.Errorf("Token %d: expected %v, got %v", i, expectedType, tokens[i].Type)
				}
			}
		})
	}
}

func TestNixParser_ErrorHandling(t *testing.T) {
	parser := NewNixParser()
	
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:    "unclosed attribute set",
			source:  `{ name = "value"`,
			wantErr: true,
		},
		{
			name:    "unclosed list",
			source:  `[ "a" "b"`,
			wantErr: true,
		},
		{
			name:    "unclosed string",
			source:  `"hello world`,
			wantErr: true,
		},
		{
			name:    "invalid character",
			source:  `@invalid`,
			wantErr: true,
		},
		{
			name:    "missing equals",
			source:  `{ name "value"; }`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ParseExpression(tt.source)
			
			if tt.wantErr && err == nil {
				t.Errorf("ParseExpression() expected error but got none")
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("ParseExpression() unexpected error: %v", err)
			}
		})
	}
}

func TestNixParser_SemanticAnalysis(t *testing.T) {
	parser := NewNixParser()
	
	tests := []struct {
		name         string
		source       string
		expectedIntent string
	}{
		{
			name:           "service configuration",
			source:         `{ services.nginx.enable = true; }`,
			expectedIntent: "service_configuration",
		},
		{
			name:           "environment configuration",
			source:         `{ environment.systemPackages = []; }`,
			expectedIntent: "environment_configuration",
		},
		{
			name:           "network configuration",
			source:         `{ networking.hostName = "test"; }`,
			expectedIntent: "network_configuration",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpression(tt.source)
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}
			
			if expr.Metadata.Intent != tt.expectedIntent {
				t.Errorf("Expected intent %s, got %s", tt.expectedIntent, expr.Metadata.Intent)
			}
		})
	}
}

func BenchmarkNixParser_ParseExpression(b *testing.B) {
	parser := NewNixParser()
	
	source := `{
		services = {
			nginx = {
				enable = true;
				virtualHosts = {
					"example.com" = {
						locations."/" = {
							proxyPass = "http://localhost:3000";
							extraConfig = ''
								proxy_set_header Host $host;
								proxy_set_header X-Real-IP $remote_addr;
							'';
						};
					};
				};
			};
			postgresql = {
				enable = true;
				package = pkgs.postgresql_13;
				dataDir = "/var/lib/postgresql/13";
			};
		};
		environment.systemPackages = with pkgs; [
			git vim nodejs python3 docker
		];
		users.users.myuser = {
			isNormalUser = true;
			extraGroups = [ "wheel" "docker" ];
		};
	}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseExpression(source)
		if err != nil {
			b.Fatalf("ParseExpression() error = %v", err)
		}
	}
}