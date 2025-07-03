# Simple Custom Agent Plugin Example

This is a working example of how to create a custom NixOS assistant agent as a plugin for nixai.

## 🚀 Features

This simple agent provides:
- **Concept Explanation**: Explains NixOS concepts at different levels (beginner/intermediate/advanced)
- **Configuration Analysis**: Analyzes NixOS configurations for issues and improvements
- **Improvement Suggestions**: Suggests optimizations based on scenarios and goals

## 🛠️ Building the Plugin

```bash
# Navigate to the plugin directory
cd examples/custom-agents/simple-agent-plugin/

# Initialize the module
go mod tidy

# Build the plugin
go build -buildmode=plugin -o simple-agent.so .
```

## 📦 Installing the Plugin

```bash
# Install the plugin
nixai plugin install simple-agent.so

# Enable the plugin
nixai plugin enable simple-custom-agent

# Verify installation
nixai plugin list
```

## 💡 Usage Examples

### 1. Explain NixOS Concepts

```bash
# Beginner explanation of flakes
nixai plugin execute simple-custom-agent explain-concept --params '{"concept":"flakes","level":"beginner"}'

# Advanced explanation of nixpkgs
nixai plugin execute simple-custom-agent explain-concept --params '{"concept":"nixpkgs","level":"advanced"}'

# Explain any concept
nixai plugin execute simple-custom-agent explain-concept --params '{"concept":"channels"}'
```

### 2. Analyze Configuration

```bash
# Analyze a configuration snippet
nixai plugin execute simple-custom-agent analyze-config --params '{
  "config": "services.openssh.enable = true;\nusers.users.root.hashedPassword = \"$6$...\";",
  "focus": "security"
}'

# Check best practices
nixai plugin execute simple-custom-agent analyze-config --params '{
  "config": "environment.systemPackages = with pkgs; [ vim git ];\nservices.nginx.enable = true;",
  "focus": "best-practices"
}'
```

### 3. Get Improvement Suggestions

```bash
# Performance optimization suggestions
nixai plugin execute simple-custom-agent suggest-improvement --params '{
  "scenario": "My NixOS system is running slowly during builds",
  "goal": "performance"
}'

# Security hardening suggestions
nixai plugin execute simple-custom-agent suggest-improvement --params '{
  "scenario": "Setting up a new server",
  "goal": "security"
}'

# Development environment setup
nixai plugin execute simple-custom-agent suggest-improvement --params '{
  "scenario": "I need a development environment for Python and Node.js",
  "goal": "development"
}'
```

## 🔧 Plugin Operations

### Available Operations

1. **explain-concept**
   - Parameters: `concept` (string), `level` (optional: beginner/intermediate/advanced)
   - Returns: Detailed explanation string
   - Purpose: Educational explanations of NixOS concepts

2. **analyze-config**
   - Parameters: `config` (string), `focus` (optional: security/performance/best-practices)
   - Returns: Analysis object with issues and suggestions
   - Purpose: Static analysis of NixOS configurations

3. **suggest-improvement**
   - Parameters: `scenario` (string), `goal` (optional: optimization/security/development)
   - Returns: Improvement suggestions object
   - Purpose: Contextual recommendations for system improvements

### Response Formats

All operations return structured data that can be easily parsed and used in scripts or other tools.

Example response for `analyze-config`:
```json
{
  "config": "services.openssh.enable = true;",
  "focus": "security",
  "issues": [
    "SSH password authentication may be enabled"
  ],
  "suggestions": [
    "Consider services.openssh.settings.PasswordAuthentication = false;"
  ],
  "severity": "medium",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🎯 Extending the Plugin

This example demonstrates the basic structure. To extend it:

1. **Add more operations** in `GetOperations()`
2. **Implement the operations** in `Execute()`
3. **Add validation** in the parameter definitions
4. **Include AI integration** by adding AI provider calls
5. **Add persistent state** if needed for complex operations

### Adding AI Integration

To make this agent truly intelligent, you can integrate with nixai's AI providers:

```go
// Add AI provider to the plugin
type SimpleCustomAgent struct {
    // ... existing fields
    aiProvider ai.Provider
}

// Use AI in operations
func (p *SimpleCustomAgent) explainConcept(params map[string]interface{}) (interface{}, error) {
    concept := params["concept"].(string)
    level := params["level"].(string)
    
    prompt := fmt.Sprintf("Explain the NixOS concept '%s' at a %s level", concept, level)
    response, err := p.aiProvider.Query(prompt)
    if err != nil {
        return nil, err
    }
    
    return response, nil
}
```

## 🔍 Testing the Plugin

```bash
# Check plugin status
nixai plugin status simple-custom-agent

# View plugin info
nixai plugin info simple-custom-agent

# Test health check
nixai plugin validate simple-agent.so

# View plugin metrics
nixai plugin metrics simple-custom-agent
```

## 📚 Next Steps

1. **Study the comprehensive examples** in the main README
2. **Add AI integration** using nixai's provider system
3. **Implement persistent storage** if your agent needs state
4. **Add configuration options** via the plugin config system
5. **Create specialized agents** for your specific use cases

This simple example provides a foundation for building more sophisticated custom agents that can leverage nixai's full plugin ecosystem!