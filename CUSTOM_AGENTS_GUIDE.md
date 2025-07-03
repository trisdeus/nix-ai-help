# 🤖 Custom NixAI Agents: Complete Developer Guide

## 🎯 **YES, Users CAN Create Custom Agents as Plugins!**

NixAI has a **comprehensive agent system** that supports multiple ways for users to create their own AI agents as plugins. The system includes:

- ✅ **Advanced Agent Framework** with role-based prompting
- ✅ **Function Integration System** with automatic registration
- ✅ **Full Plugin Architecture** with lifecycle management
- ✅ **AI Provider Integration** supporting multiple AI models
- ✅ **Security & Sandboxing** for safe plugin execution
- ✅ **Auto-Discovery & Registration** for seamless integration

---

## 🏗️ **NixAI Agent Architecture Overview**

### Core Components

1. **Agent Interface** (`internal/ai/agent/`)
   - BaseAgent with role management
   - Context-aware prompting
   - AI provider abstraction
   - 36+ specialized roles

2. **Function System** (`internal/ai/function/`)
   - FunctionInterface for AI functions
   - Parameter validation & schemas
   - 27+ built-in functions
   - Automatic registration

3. **Plugin System** (`internal/plugins/`)
   - Complete lifecycle management
   - Security sandboxing
   - Resource limits
   - Event system

4. **AI Provider Integration**
   - Multi-provider support (Ollama, OpenAI, Claude, etc.)
   - Context-aware queries
   - Caching and performance optimization

---

## 🚀 **Three Ways to Create Custom Agents**

### 1. **AI Function Agent** (Easiest)
**Best for**: Single-purpose agents with automatic integration

```go
type MyCustomFunction struct {
    *functionbase.BaseFunction
    provider ai.Provider
}

func (f *MyCustomFunction) Execute(ctx context.Context, params map[string]interface{}, options *FunctionOptions) (*FunctionResult, error) {
    // Your custom logic here
    question := params["question"].(string)
    response, err := f.provider.Query("Enhanced prompt: " + question)
    return functionbase.SuccessResult(response, time.Since(start)), nil
}
```

**Integration**: Automatic registration with function manager

### 2. **Plugin Agent** (Most Flexible)
**Best for**: Complex multi-operation agents with full lifecycle

```go
type MyAgentPlugin struct {
    agent agent.Agent
}

func (p *MyAgentPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
    switch operation {
    case "analyze": return p.analyze(ctx, params)
    case "suggest": return p.suggest(ctx, params)
    default: return nil, fmt.Errorf("unknown operation")
    }
}

func (p *MyAgentPlugin) GetOperations() []plugins.PluginOperation {
    return []plugins.PluginOperation{
        {Name: "analyze", Description: "Analyze something", ...},
        {Name: "suggest", Description: "Suggest improvements", ...},
    }
}
```

**Integration**: Full plugin lifecycle with `nixai plugin` commands

### 3. **Direct Agent** (Most Powerful)
**Best for**: Core system integration with maximum framework access

```go
type AdvancedCustomAgent struct {
    *agent.BaseAgent
    logger *logger.Logger
    config *config.UserConfig
}

func (a *AdvancedCustomAgent) Query(ctx context.Context, question string) (string, error) {
    // Enhanced security-focused prompt
    enhancedPrompt := a.buildSpecializedPrompt(question)
    response, err := a.BaseAgent.GenerateResponse(ctx, enhancedPrompt)
    return a.postProcessResponse(response), err
}
```

**Integration**: Direct registration with agent system

---

## 📊 **Feature Comparison**

| Feature | AI Function | Plugin Agent | Direct Agent |
|---------|-------------|--------------|--------------|
| **Complexity** | Low | Medium | High |
| **Setup Time** | Minutes | Hours | Days |
| **Lifecycle Management** | Automatic | Full Control | Manual |
| **Security Isolation** | Basic | Full Sandbox | Custom |
| **Multi-Operation** | Single | Multiple | Unlimited |
| **AI Provider Access** | Full | Full | Full |
| **Role System** | Limited | Full | Full |
| **Resource Limits** | Basic | Configurable | Custom |
| **Event System** | No | Yes | Yes |
| **Hot Reload** | Yes | Yes | Custom |

---

## 🎨 **Working Examples Provided**

### 1. **Simple Agent Plugin** (`examples/custom-agents/simple-agent-plugin/`)
A complete working example that demonstrates:
- ✅ NixOS concept explanation at different skill levels
- ✅ Configuration analysis with security focus
- ✅ Improvement suggestions based on scenarios
- ✅ Proper plugin structure and registration
- ✅ Ready to build and install

```bash
# Build and install the example
cd examples/custom-agents/simple-agent-plugin/
go build -buildmode=plugin -o simple-agent.so .
nixai plugin install simple-agent.so
nixai plugin enable simple-custom-agent

# Use the agent
nixai plugin execute simple-custom-agent explain-concept --params '{"concept":"flakes","level":"beginner"}'
```

### 2. **Advanced Security Agent** (Conceptual)
Shows enterprise-level features:
- ✅ Vulnerability assessment with CVE database integration
- ✅ Real-time security monitoring
- ✅ AI-powered threat analysis
- ✅ Security hardening plan generation
- ✅ Compliance checking

### 3. **DevOps Assistant Agent** (Conceptual)
Demonstrates complex operations:
- ✅ Deployment strategy advice
- ✅ Infrastructure design recommendations  
- ✅ Troubleshooting assistance
- ✅ Monitoring setup guidance
- ✅ Multi-environment support

---

## 💼 **Real-World Use Cases**

### Specialized Domain Agents
- **Security Analyst Agent**: CVE scanning, penetration testing, compliance checking
- **Performance Optimization Agent**: System tuning, resource analysis, bottleneck detection
- **Development Environment Agent**: Project setup, dependency management, tool configuration
- **Infrastructure Agent**: Cloud deployment, scaling strategies, cost optimization
- **Backup & Recovery Agent**: Backup strategies, disaster recovery planning, data protection

### Industry-Specific Agents
- **Healthcare Agent**: HIPAA compliance, medical device integration, security standards
- **Financial Agent**: PCI compliance, high-availability setup, audit logging
- **Education Agent**: Student environment management, software licensing, content filtering
- **Gaming Agent**: Performance optimization, graphics configuration, multiplayer setup

### Personal Workflow Agents
- **Home Lab Agent**: Home server setup, media center configuration, network management
- **Developer Agent**: IDE setup, build optimization, deployment automation
- **System Admin Agent**: Monitoring setup, log analysis, maintenance scheduling

---

## 🔧 **Development Workflow**

### 1. **Choose Your Approach**
```bash
# For simple agents: Start with AI Functions
# For complex agents: Use Plugin Architecture  
# For core integration: Direct Agent Implementation
```

### 2. **Development Process**
```bash
# 1. Create agent structure
mkdir my-custom-agent && cd my-custom-agent

# 2. Implement agent interface
# (Follow examples in /examples/custom-agents/)

# 3. Build plugin
go build -buildmode=plugin -o my-agent.so .

# 4. Test plugin
nixai plugin validate my-agent.so

# 5. Install and enable
nixai plugin install my-agent.so
nixai plugin enable my-agent

# 6. Test operations
nixai plugin execute my-agent operation-name --params '{...}'
```

### 3. **Integration with NixAI**
```bash
# Agent appears in plugin list
nixai plugin list

# Available in TUI completion
nixai tui  # Type "my-agent" and see completion

# Integrated with help system
nixai plugin info my-agent

# Monitoring and metrics
nixai plugin metrics my-agent
nixai plugin status my-agent
```

---

## 🔐 **Security & Best Practices**

### Security Features
- ✅ **Sandboxed Execution**: Resource limits, filesystem restrictions
- ✅ **Permission Model**: Configurable access controls
- ✅ **Input Validation**: Automatic parameter validation
- ✅ **Audit Logging**: Track agent execution and access
- ✅ **Health Monitoring**: Detect and recover from failures

### Best Practices
1. **Validate All Inputs**: Use parameter schemas and validation
2. **Handle Errors Gracefully**: Provide helpful error messages
3. **Implement Health Checks**: Monitor agent status and performance
4. **Use Structured Logging**: Enable debugging and monitoring
5. **Follow Security Guidelines**: Minimize privileges and access
6. **Document Operations**: Provide clear examples and documentation
7. **Test Thoroughly**: Validate all operations and edge cases

---

## 📖 **Documentation & Resources**

### Provided Documentation
- ✅ **Complete Developer Guide** (`examples/custom-agents/README.md`)
- ✅ **Working Examples** with step-by-step instructions
- ✅ **API Reference** for all interfaces and methods
- ✅ **Security Guidelines** for safe plugin development
- ✅ **Best Practices** from real-world usage

### Community Resources
- **NixOS Documentation**: https://nixos.org/manual/
- **Nix Pills Tutorial**: https://nixos.org/guides/nix-pills/
- **Community Wiki**: https://nixos.wiki/
- **Discourse Forum**: https://discourse.nixos.org/

### Getting Help
```bash
# Built-in help
nixai plugin --help
nixai plugin create --help

# Show integrated plugin commands
nixai plugin integrated

# View plugin documentation
nixai plugin info <plugin-name>
```

---

## 🎯 **Summary: Yes, Users Can Create Custom Agents!**

**NixAI provides THREE powerful approaches** for users to create custom agents:

### ✅ **Supported Features**
- Multiple agent implementation patterns
- Full AI provider integration (Ollama, OpenAI, Claude, etc.)
- Automatic registration and discovery
- Role-based prompting system
- Security sandboxing and resource limits
- Plugin lifecycle management
- Real-time health monitoring
- TUI and CLI integration
- Configuration management

### ✅ **Working Examples**
- Simple agent plugin (ready to build and use)
- Advanced security agent (enterprise features)
- DevOps assistant agent (complex operations)
- Complete documentation and tutorials

### ✅ **Easy Development**
- Choose complexity level based on needs
- Rich plugin framework with validation
- Automatic integration with nixai commands
- Built-in testing and debugging tools

### ✅ **Production Ready**
- Security isolation and resource limits
- Health monitoring and error recovery
- Audit logging and compliance features
- Performance optimization and caching

**Users can create everything from simple concept explainers to sophisticated enterprise security agents, all integrated seamlessly into the nixai ecosystem!** 🚀

Start with the simple example in `examples/custom-agents/simple-agent-plugin/` and build from there! 🛠️