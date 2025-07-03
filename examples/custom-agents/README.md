# 🤖 Creating Custom NixAI Agents as Plugins

NixAI supports multiple ways for users to create their own AI agents as plugins. This guide shows you all the available approaches with complete examples.

## 🎯 **Three Agent Plugin Approaches**

### 1. **AI Function Agent** (Simplest)
- Integrates with the function registry
- Automatic parameter validation
- Built-in metrics and logging
- Best for: Single-purpose agents

### 2. **Plugin Agent** (Most Flexible)
- Full plugin lifecycle management
- Multiple operations per agent
- Resource isolation and security
- Best for: Complex multi-function agents

### 3. **Direct Agent Implementation** (Most Powerful)
- Direct integration with agent system
- Full access to role system
- Advanced AI provider features
- Best for: Core system integration

---

## 🚀 **Approach 1: AI Function Agent**

This is the **easiest way** to create a custom agent that integrates seamlessly with nixai's function system.

### Example: Security Analysis Agent

```go
// security_agent_function.go
package security

import (
    "context"
    "fmt"
    "strings"
    "time"

    "nix-ai-help/internal/ai"
    "nix-ai-help/internal/ai/functionbase"
    "nix-ai-help/internal/ai/roles"
    "nix-ai-help/internal/config"
)

// SecurityAnalysisFunction implements a security-focused AI agent
type SecurityAnalysisFunction struct {
    *functionbase.BaseFunction
    provider ai.Provider
}

// NewSecurityAnalysisFunction creates a new security analysis agent function
func NewSecurityAnalysisFunction(provider ai.Provider) functionbase.FunctionInterface {
    parameters := []functionbase.FunctionParameter{
        {
            Name:        "analysis_type",
            Type:        "string",
            Description: "Type of security analysis to perform",
            Required:    true,
            Enum:        []string{"configuration", "packages", "services", "network", "comprehensive"},
        },
        {
            Name:        "target",
            Type:        "string", 
            Description: "Target to analyze (file path, service name, or 'system')",
            Required:    false,
            Default:     "system",
        },
        {
            Name:        "severity_filter",
            Type:        "string",
            Description: "Minimum severity level to report",
            Required:    false,
            Enum:        []string{"low", "medium", "high", "critical"},
            Default:     "medium",
        },
        {
            Name:        "output_format",
            Type:        "string",
            Description: "Output format for the analysis",
            Required:    false,
            Enum:        []string{"detailed", "summary", "json", "checklist"},
            Default:     "detailed",
        },
    }

    baseFunc := functionbase.NewBaseFunction(
        "security-analysis",
        "AI-powered security analysis for NixOS configurations and systems",
        parameters,
    )

    return &SecurityAnalysisFunction{
        BaseFunction: baseFunc,
        provider:     provider,
    }
}

// Execute performs the security analysis
func (f *SecurityAnalysisFunction) Execute(ctx context.Context, params map[string]interface{}, options *functionbase.FunctionOptions) (*functionbase.FunctionResult, error) {
    start := time.Now()

    // Validate parameters
    if err := f.ValidateParameters(params); err != nil {
        return functionbase.ErrorResult(err, time.Since(start)), err
    }

    // Extract parameters
    analysisType := params["analysis_type"].(string)
    target := params["target"].(string)
    severityFilter := params["severity_filter"].(string)
    outputFormat := params["output_format"].(string)

    // Build security-focused prompt
    prompt := f.buildSecurityPrompt(analysisType, target, severityFilter, outputFormat)

    // Query AI provider with security context
    response, err := f.provider.Query(prompt)
    if err != nil {
        return functionbase.ErrorResult(err, time.Since(start)), err
    }

    // Format response based on output format
    formattedResponse := f.formatSecurityResponse(response, outputFormat)

    return functionbase.SuccessResult(formattedResponse, time.Since(start)), nil
}

// buildSecurityPrompt creates a specialized security analysis prompt
func (f *SecurityAnalysisFunction) buildSecurityPrompt(analysisType, target, severity, format string) string {
    basePrompt := fmt.Sprintf(`You are a NixOS security expert specializing in %s analysis.

ANALYSIS REQUEST:
- Type: %s
- Target: %s
- Minimum Severity: %s
- Output Format: %s

SECURITY FOCUS AREAS:
1. Configuration Security (file permissions, user access, service exposure)
2. Package Vulnerability Assessment (known CVEs, outdated packages)
3. Service Security (unnecessary services, insecure configurations)
4. Network Security (open ports, firewall rules, network exposure)
5. System Hardening (security modules, encryption, access controls)

ANALYSIS INSTRUCTIONS:
- Identify potential security risks and vulnerabilities
- Provide specific NixOS configuration recommendations
- Include severity ratings (Critical/High/Medium/Low)
- Give actionable remediation steps
- Reference NixOS security best practices
- Include relevant CVE numbers where applicable

TARGET ANALYSIS: %s`, analysisType, analysisType, target, severity, format, target)

    return basePrompt
}

// formatSecurityResponse formats the AI response based on the requested format
func (f *SecurityAnalysisFunction) formatSecurityResponse(response, format string) string {
    switch format {
    case "summary":
        return f.extractSummary(response)
    case "json":
        return f.convertToJSON(response)
    case "checklist":
        return f.convertToChecklist(response)
    default:
        return response // detailed format
    }
}

// extractSummary extracts key points for summary format
func (f *SecurityAnalysisFunction) extractSummary(response string) string {
    lines := strings.Split(response, "\n")
    var summary []string
    
    for _, line := range lines {
        if strings.Contains(strings.ToLower(line), "critical") ||
           strings.Contains(strings.ToLower(line), "high") ||
           strings.Contains(strings.ToLower(line), "vulnerability") {
            summary = append(summary, strings.TrimSpace(line))
        }
    }
    
    if len(summary) == 0 {
        return "No critical security issues identified."
    }
    
    return "🔒 Security Summary:\n" + strings.Join(summary, "\n")
}

// convertToJSON converts response to structured JSON format
func (f *SecurityAnalysisFunction) convertToJSON(response string) string {
    // This would implement proper JSON conversion
    // For now, return a structured format
    return fmt.Sprintf(`{
  "security_analysis": {
    "timestamp": "%s",
    "status": "completed",
    "findings": "%s"
  }
}`, time.Now().Format(time.RFC3339), strings.ReplaceAll(response, "\n", "\\n"))
}

// convertToChecklist converts response to actionable checklist
func (f *SecurityAnalysisFunction) convertToChecklist(response string) string {
    lines := strings.Split(response, "\n")
    var checklist []string
    
    checklist = append(checklist, "🔒 Security Analysis Checklist:\n")
    
    for _, line := range lines {
        if strings.Contains(line, "recommend") || 
           strings.Contains(line, "should") ||
           strings.Contains(line, "configure") {
            checklist = append(checklist, fmt.Sprintf("☐ %s", strings.TrimSpace(line)))
        }
    }
    
    return strings.Join(checklist, "\n")
}
```

### Function Registration

```go
// Register the function in your plugin or main application
func init() {
    // Get AI provider
    cfg, _ := config.LoadUserConfig()
    provider := ai.GetProvider(cfg) // Implement based on your provider system
    
    // Create and register the function
    securityFunc := NewSecurityAnalysisFunction(provider)
    
    // Register with function manager (this would be done automatically in a plugin)
    functionManager.Register(securityFunc)
}
```

### Usage Examples

```bash
# Security analysis via AI function
nixai ask "Perform comprehensive security analysis" --function security-analysis --params '{"analysis_type":"comprehensive","output_format":"checklist"}'

# Configuration security check
nixai ask "Check my NixOS config for security issues" --function security-analysis --params '{"analysis_type":"configuration","target":"/etc/nixos"}'

# Package vulnerability scan
nixai ask "Scan for vulnerable packages" --function security-analysis --params '{"analysis_type":"packages","severity_filter":"high"}'
```

---

## 🔧 **Approach 2: Plugin Agent**

This approach provides **maximum flexibility** with full plugin lifecycle management.

### Example: DevOps Assistant Agent Plugin

```go
// devops_agent_plugin.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "nix-ai-help/internal/ai"
    "nix-ai-help/internal/ai/agent"
    "nix-ai-help/internal/ai/roles"
    "nix-ai-help/internal/plugins"
)

// DevOpsAgentPlugin implements a comprehensive DevOps assistant
type DevOpsAgentPlugin struct {
    agent    agent.Agent
    provider ai.Provider
    config   plugins.PluginConfig
    started  bool
}

// Plugin Interface Implementation

func (p *DevOpsAgentPlugin) Name() string { return "devops-agent" }
func (p *DevOpsAgentPlugin) Version() string { return "1.0.0" }
func (p *DevOpsAgentPlugin) Description() string { 
    return "AI-powered DevOps assistant for NixOS infrastructure management" 
}
func (p *DevOpsAgentPlugin) Author() string { return "Your Name" }
func (p *DevOpsAgentPlugin) Repository() string { return "https://github.com/yourname/nixai-devops-agent" }
func (p *DevOpsAgentPlugin) License() string { return "MIT" }
func (p *DevOpsAgentPlugin) Dependencies() []string { return []string{} }
func (p *DevOpsAgentPlugin) Capabilities() []string { 
    return []string{"devops", "infrastructure", "monitoring", "deployment", "troubleshooting"}
}

// Lifecycle Management

func (p *DevOpsAgentPlugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
    p.config = config
    
    // Initialize AI agent with DevOps role
    p.agent = &agent.BaseAgent{}
    p.agent.SetRole(roles.RoleType("devops-specialist"))
    
    return nil
}

func (p *DevOpsAgentPlugin) Start(ctx context.Context) error {
    if p.provider == nil {
        return fmt.Errorf("AI provider not set")
    }
    
    p.agent.SetProvider(p.provider)
    p.started = true
    return nil
}

func (p *DevOpsAgentPlugin) Stop(ctx context.Context) error {
    p.started = false
    return nil
}

// Operation Definitions

func (p *DevOpsAgentPlugin) GetOperations() []plugins.PluginOperation {
    return []plugins.PluginOperation{
        {
            Name:        "deploy-advice",
            Description: "Get deployment strategy advice for NixOS systems",
            Parameters: map[string]plugins.PluginParameter{
                "environment": {
                    Type:        "string",
                    Description: "Target environment (dev/staging/prod)",
                    Required:    true,
                    Enum:        []string{"dev", "staging", "prod"},
                },
                "infrastructure": {
                    Type:        "string", 
                    Description: "Infrastructure type (bare-metal/cloud/hybrid)",
                    Required:    false,
                    Default:     "cloud",
                },
                "scale": {
                    Type:        "string",
                    Description: "Deployment scale (small/medium/large/enterprise)",
                    Required:    false,
                    Default:     "medium",
                },
            },
            ReturnType: "object",
            Examples: []plugins.PluginExample{
                {
                    Description: "Production deployment advice",
                    Parameters: map[string]interface{}{
                        "environment": "prod",
                        "infrastructure": "cloud",
                        "scale": "large",
                    },
                    Expected: "Detailed deployment strategy with security considerations",
                },
            },
        },
        {
            Name:        "troubleshoot-issue",
            Description: "AI-powered troubleshooting for DevOps issues",
            Parameters: map[string]plugins.PluginParameter{
                "issue_type": {
                    Type:        "string",
                    Description: "Type of issue to troubleshoot",
                    Required:    true,
                    Enum:        []string{"deployment", "performance", "networking", "storage", "security"},
                },
                "symptoms": {
                    Type:        "string",
                    Description: "Description of the symptoms or error messages",
                    Required:    true,
                },
                "environment_info": {
                    Type:        "object",
                    Description: "Environment information (optional)",
                    Required:    false,
                },
            },
            ReturnType: "object",
        },
        {
            Name:        "infrastructure-design",
            Description: "Generate infrastructure design recommendations",
            Parameters: map[string]plugins.PluginParameter{
                "requirements": {
                    Type:        "object",
                    Description: "Infrastructure requirements and constraints",
                    Required:    true,
                },
                "budget": {
                    Type:        "string",
                    Description: "Budget constraints (low/medium/high/unlimited)",
                    Required:    false,
                    Default:     "medium",
                },
            },
            ReturnType: "object",
        },
        {
            Name:        "monitoring-setup",
            Description: "Get monitoring and observability setup recommendations",
            Parameters: map[string]plugins.PluginParameter{
                "services": {
                    Type:        "array",
                    Description: "List of services to monitor",
                    Required:    true,
                },
                "monitoring_type": {
                    Type:        "string",
                    Description: "Type of monitoring needed",
                    Required:    false,
                    Enum:        []string{"basic", "comprehensive", "enterprise"},
                    Default:     "comprehensive",
                },
            },
            ReturnType: "object",
        },
    }
}

// Operation Execution

func (p *DevOpsAgentPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
    if !p.started {
        return nil, fmt.Errorf("plugin not started")
    }

    switch operation {
    case "deploy-advice":
        return p.executeDeployAdvice(ctx, params)
    case "troubleshoot-issue":
        return p.executeTroubleshoot(ctx, params)
    case "infrastructure-design":
        return p.executeInfrastructureDesign(ctx, params)
    case "monitoring-setup":
        return p.executeMonitoringSetup(ctx, params)
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
}

// Operation Implementations

func (p *DevOpsAgentPlugin) executeDeployAdvice(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    environment := params["environment"].(string)
    infrastructure := getStringParam(params, "infrastructure", "cloud")
    scale := getStringParam(params, "scale", "medium")

    prompt := fmt.Sprintf(`As a DevOps expert specializing in NixOS deployments, provide comprehensive deployment advice for:

DEPLOYMENT CONTEXT:
- Environment: %s
- Infrastructure: %s
- Scale: %s

PROVIDE DETAILED ADVICE ON:
1. Deployment Strategy & Methodology
2. Infrastructure as Code (NixOS configurations)
3. CI/CD Pipeline recommendations
4. Security considerations
5. Monitoring and observability
6. Rollback strategies
7. Environment-specific best practices
8. Scaling considerations
9. Cost optimization
10. Risk mitigation

Include specific NixOS configurations, flake.nix examples, and practical implementation steps.`, 
        environment, infrastructure, scale)

    response, err := p.agent.GenerateResponse(ctx, prompt)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "advice": response,
        "environment": environment,
        "infrastructure": infrastructure,
        "scale": scale,
        "timestamp": time.Now().Format(time.RFC3339),
    }, nil
}

func (p *DevOpsAgentPlugin) executeTroubleshoot(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    issueType := params["issue_type"].(string)
    symptoms := params["symptoms"].(string)
    envInfo := params["environment_info"]

    prompt := fmt.Sprintf(`As a DevOps troubleshooting expert for NixOS systems, help diagnose and resolve this issue:

ISSUE DETAILS:
- Type: %s
- Symptoms: %s
- Environment: %v

TROUBLESHOOTING APPROACH:
1. Root Cause Analysis
2. Diagnostic Commands
3. Step-by-step Investigation
4. Potential Solutions
5. Prevention Strategies

Provide specific NixOS commands, configuration checks, and debugging strategies.
Include both immediate fixes and long-term improvements.`, 
        issueType, symptoms, envInfo)

    response, err := p.agent.GenerateResponse(ctx, prompt)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "troubleshooting_guide": response,
        "issue_type": issueType,
        "urgency": p.determineUrgency(issueType, symptoms),
        "estimated_resolution_time": p.estimateResolutionTime(issueType),
        "timestamp": time.Now().Format(time.RFC3339),
    }, nil
}

func (p *DevOpsAgentPlugin) executeInfrastructureDesign(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    requirements := params["requirements"]
    budget := getStringParam(params, "budget", "medium")

    prompt := fmt.Sprintf(`As an infrastructure architect specializing in NixOS, design a comprehensive infrastructure solution:

REQUIREMENTS: %v
BUDGET CONSTRAINTS: %s

DESIGN COMPREHENSIVE ARCHITECTURE INCLUDING:
1. System Architecture Overview
2. NixOS Configuration Structure
3. Networking Design
4. Storage Strategy
5. Security Architecture
6. Scalability Plan
7. Disaster Recovery
8. Monitoring Strategy
9. Cost Analysis
10. Implementation Roadmap

Provide flake.nix examples, configuration snippets, and deployment guides.`, 
        requirements, budget)

    response, err := p.agent.GenerateResponse(ctx, prompt)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "design": response,
        "requirements": requirements,
        "budget": budget,
        "estimated_cost": p.estimateCost(budget, requirements),
        "implementation_timeline": p.estimateTimeline(requirements),
        "timestamp": time.Now().Format(time.RFC3339),
    }, nil
}

func (p *DevOpsAgentPlugin) executeMonitoringSetup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    services := params["services"].([]interface{})
    monitoringType := getStringParam(params, "monitoring_type", "comprehensive")

    prompt := fmt.Sprintf(`As a monitoring specialist for NixOS environments, design a comprehensive monitoring solution:

SERVICES TO MONITOR: %v
MONITORING LEVEL: %s

DESIGN MONITORING STACK INCLUDING:
1. Metrics Collection (Prometheus, etc.)
2. Log Aggregation (Loki, etc.) 
3. Alerting Strategy
4. Dashboards and Visualization
5. Health Checks
6. Performance Monitoring
7. Security Monitoring
8. Custom Metrics
9. Integration with NixOS
10. Maintenance and Operations

Provide complete NixOS configurations for the monitoring stack.`, 
        services, monitoringType)

    response, err := p.agent.GenerateResponse(ctx, prompt)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "monitoring_plan": response,
        "services": services,
        "monitoring_type": monitoringType,
        "recommended_tools": p.getRecommendedTools(monitoringType),
        "setup_complexity": p.assessComplexity(services, monitoringType),
        "timestamp": time.Now().Format(time.RFC3339),
    }, nil
}

// Health and Metrics

func (p *DevOpsAgentPlugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
    status := "healthy"
    if !p.started {
        status = "stopped"
    }

    return plugins.PluginHealth{
        Status:      status,
        LastCheck:   time.Now(),
        Message:     "DevOps agent is operational",
        Metrics:     map[string]interface{}{
            "operations_supported": len(p.GetOperations()),
            "uptime": time.Since(time.Now()).String(),
        },
    }
}

func (p *DevOpsAgentPlugin) GetMetrics() plugins.PluginMetrics {
    return plugins.PluginMetrics{
        ExecutionCount:   0, // Track in real implementation
        TotalExecutions:  0,
        AverageExecution: 0,
        LastExecution:    time.Time{},
        ErrorCount:       0,
        SuccessRate:      1.0,
    }
}

// Utility Functions

func getStringParam(params map[string]interface{}, key, defaultValue string) string {
    if val, exists := params[key]; exists {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return defaultValue
}

func (p *DevOpsAgentPlugin) determineUrgency(issueType, symptoms string) string {
    if strings.Contains(symptoms, "down") || strings.Contains(symptoms, "failed") {
        return "critical"
    }
    if strings.Contains(symptoms, "slow") || strings.Contains(symptoms, "error") {
        return "high"
    }
    return "medium"
}

func (p *DevOpsAgentPlugin) estimateResolutionTime(issueType string) string {
    switch issueType {
    case "networking":
        return "2-4 hours"
    case "deployment":
        return "1-2 hours"
    case "performance":
        return "4-8 hours"
    case "security":
        return "1-3 hours"
    default:
        return "2-6 hours"
    }
}

func (p *DevOpsAgentPlugin) estimateCost(budget string, requirements interface{}) string {
    switch budget {
    case "low":
        return "$100-500/month"
    case "medium":
        return "$500-2000/month"
    case "high":
        return "$2000-10000/month"
    default:
        return "Contact for quote"
    }
}

func (p *DevOpsAgentPlugin) estimateTimeline(requirements interface{}) string {
    // Analyze requirements complexity and provide timeline
    return "2-4 weeks for initial implementation"
}

func (p *DevOpsAgentPlugin) getRecommendedTools(monitoringType string) []string {
    switch monitoringType {
    case "basic":
        return []string{"Prometheus", "Grafana", "AlertManager"}
    case "comprehensive":
        return []string{"Prometheus", "Grafana", "Loki", "AlertManager", "Jaeger"}
    case "enterprise":
        return []string{"Prometheus", "Grafana", "Loki", "AlertManager", "Jaeger", "Thanos", "OpenTelemetry"}
    default:
        return []string{"Prometheus", "Grafana"}
    }
}

func (p *DevOpsAgentPlugin) assessComplexity(services []interface{}, monitoringType string) string {
    serviceCount := len(services)
    switch {
    case serviceCount <= 5 && monitoringType == "basic":
        return "low"
    case serviceCount <= 15 && monitoringType != "enterprise":
        return "medium"
    default:
        return "high"
    }
}

// Plugin Export (required for Go plugins)
var Plugin DevOpsAgentPlugin

func NewPlugin() plugins.PluginInterface {
    return &Plugin
}
```

### Plugin Configuration

```yaml
# devops-agent-config.yaml
name: devops-agent
version: 1.0.0
description: "AI-powered DevOps assistant for NixOS infrastructure management"
capabilities:
  - devops
  - infrastructure  
  - monitoring
  - deployment
  - troubleshooting

configuration_schema:
  type: object
  properties:
    ai_provider:
      type: string
      default: "ollama"
    model:
      type: string
      default: "llama3"
    response_format:
      type: string
      enum: ["detailed", "concise", "technical"]
      default: "detailed"

security:
  sandbox_level: basic
  allow_file_system: true
  allow_network: true

resources:
  max_memory_mb: 256
  max_cpu_percent: 25
  max_execution_sec: 120
```

### Installation and Usage

```bash
# Build the plugin
go build -buildmode=plugin -o devops-agent.so .

# Install the plugin
nixai plugin install devops-agent.so

# Enable the plugin
nixai plugin enable devops-agent

# Use the plugin operations
nixai plugin execute devops-agent deploy-advice --params '{"environment":"prod","infrastructure":"cloud","scale":"large"}'

nixai plugin execute devops-agent troubleshoot-issue --params '{"issue_type":"deployment","symptoms":"Service fails to start after nixos-rebuild"}'

nixai plugin execute devops-agent monitoring-setup --params '{"services":["nginx","postgresql","redis"],"monitoring_type":"comprehensive"}'
```

---

## 💪 **Approach 3: Direct Agent Implementation**

This approach gives you **maximum power** by integrating directly with nixai's agent system.

### Example: Advanced Security Agent

```go
// advanced_security_agent.go
package security

import (
    "context"
    "fmt"
    "strings"
    "time"

    "nix-ai-help/internal/ai"
    "nix-ai-help/internal/ai/agent"
    "nix-ai-help/internal/ai/roles"
    "nix-ai-help/internal/config"
    "nix-ai-help/pkg/logger"
)

// AdvancedSecurityAgent implements a comprehensive security analysis agent
type AdvancedSecurityAgent struct {
    *agent.BaseAgent
    logger       *logger.Logger
    config       *config.UserConfig
    
    // Security-specific context
    securityDB   SecurityDatabase
    cveDatabase  CVEDatabase
    ruleEngine   SecurityRuleEngine
}

// SecurityContext provides detailed security analysis context
type SecurityContext struct {
    SystemInfo      SystemInfo      `json:"system_info"`
    Configurations  []ConfigFile    `json:"configurations"`
    InstalledPkgs   []Package       `json:"installed_packages"`
    RunningServices []Service       `json:"running_services"`
    NetworkConfig   NetworkConfig   `json:"network_config"`
    SecurityPolicies []SecurityPolicy `json:"security_policies"`
}

// NewAdvancedSecurityAgent creates a new advanced security agent
func NewAdvancedSecurityAgent(provider ai.Provider, cfg *config.UserConfig) (*AdvancedSecurityAgent, error) {
    baseAgent := &agent.BaseAgent{}
    
    // Set specialized security role
    if err := baseAgent.SetRole(roles.RoleType("security-expert")); err != nil {
        return nil, fmt.Errorf("failed to set security role: %w", err)
    }
    
    baseAgent.SetProvider(provider)
    
    agent := &AdvancedSecurityAgent{
        BaseAgent: baseAgent,
        logger:    logger.NewLogger(),
        config:    cfg,
    }
    
    // Initialize security databases and rule engine
    if err := agent.initializeSecurityResources(); err != nil {
        return nil, fmt.Errorf("failed to initialize security resources: %w", err)
    }
    
    return agent, nil
}

// Query performs security-aware queries
func (a *AdvancedSecurityAgent) Query(ctx context.Context, question string) (string, error) {
    // Gather security context
    secContext, err := a.gatherSecurityContext(ctx)
    if err != nil {
        a.logger.Warn(fmt.Sprintf("Failed to gather full security context: %v", err))
        secContext = &SecurityContext{} // Use empty context as fallback
    }
    
    // Set security context for the agent
    a.SetContext(secContext)
    
    // Enhance question with security focus
    enhancedPrompt := a.enhanceSecurityPrompt(question, secContext)
    
    // Query with enhanced prompt
    response, err := a.BaseAgent.GenerateResponse(ctx, enhancedPrompt)
    if err != nil {
        return "", fmt.Errorf("security analysis failed: %w", err)
    }
    
    // Post-process response with security insights
    finalResponse := a.addSecurityInsights(response, secContext)
    
    return finalResponse, nil
}

// Specialized Security Methods

func (a *AdvancedSecurityAgent) PerformVulnerabilityAssessment(ctx context.Context) (*VulnerabilityReport, error) {
    a.logger.Info("Starting comprehensive vulnerability assessment")
    
    secContext, err := a.gatherSecurityContext(ctx)
    if err != nil {
        return nil, err
    }
    
    // Analyze vulnerabilities using multiple sources
    report := &VulnerabilityReport{
        Timestamp: time.Now(),
        SystemID:  secContext.SystemInfo.Hostname,
    }
    
    // 1. Package vulnerability scan
    pkgVulns, err := a.scanPackageVulnerabilities(secContext.InstalledPkgs)
    if err != nil {
        a.logger.Error(fmt.Sprintf("Package vulnerability scan failed: %v", err))
    } else {
        report.PackageVulnerabilities = pkgVulns
    }
    
    // 2. Configuration security analysis
    configIssues, err := a.analyzeConfigurationSecurity(secContext.Configurations)
    if err != nil {
        a.logger.Error(fmt.Sprintf("Configuration analysis failed: %v", err))
    } else {
        report.ConfigurationIssues = configIssues
    }
    
    // 3. Service security assessment
    serviceIssues, err := a.assessServiceSecurity(secContext.RunningServices)
    if err != nil {
        a.logger.Error(fmt.Sprintf("Service assessment failed: %v", err))
    } else {
        report.ServiceIssues = serviceIssues
    }
    
    // 4. Network security evaluation
    networkIssues, err := a.evaluateNetworkSecurity(secContext.NetworkConfig)
    if err != nil {
        a.logger.Error(fmt.Sprintf("Network evaluation failed: %v", err))
    } else {
        report.NetworkIssues = networkIssues
    }
    
    // 5. AI-powered analysis
    aiInsights, err := a.generateAISecurityInsights(ctx, secContext, report)
    if err != nil {
        a.logger.Error(fmt.Sprintf("AI insights generation failed: %v", err))
    } else {
        report.AIInsights = aiInsights
    }
    
    // Calculate overall risk score
    report.OverallRisk = a.calculateRiskScore(report)
    
    return report, nil
}

func (a *AdvancedSecurityAgent) GenerateSecurityHardening(ctx context.Context, target string) (*HardeningPlan, error) {
    prompt := fmt.Sprintf(`As a NixOS security expert, create a comprehensive security hardening plan for: %s

HARDENING PLAN REQUIREMENTS:
1. System-level hardening
2. Service-specific security measures
3. Network security enhancements
4. Access control improvements
5. Monitoring and detection
6. Compliance considerations
7. NixOS-specific security modules
8. Automated security measures

Provide specific NixOS configurations, flake.nix examples, and implementation steps.
Include both immediate actions and long-term security strategy.`, target)

    response, err := a.GenerateResponse(ctx, prompt)
    if err != nil {
        return nil, err
    }
    
    return &HardeningPlan{
        Target:              target,
        Recommendations:     response,
        Priority:            a.assessHardeningPriority(target),
        EstimatedEffort:     a.estimateHardeningEffort(target),
        ComplianceStandards: a.getApplicableStandards(target),
        Timestamp:          time.Now(),
    }, nil
}

func (a *AdvancedSecurityAgent) MonitorSecurityEvents(ctx context.Context, duration time.Duration) (*SecurityMonitoringReport, error) {
    // Real-time security monitoring implementation
    // This would integrate with system logs, audit logs, etc.
    
    report := &SecurityMonitoringReport{
        StartTime: time.Now(),
        Duration:  duration,
        Events:    []SecurityEvent{},
    }
    
    // Monitor various security events
    events := make(chan SecurityEvent, 100)
    
    // Start monitoring goroutines
    go a.monitorAuthEvents(ctx, events)
    go a.monitorNetworkEvents(ctx, events)
    go a.monitorFileSystemEvents(ctx, events)
    go a.monitorProcessEvents(ctx, events)
    
    // Collect events for specified duration
    timeout := time.After(duration)
    
    for {
        select {
        case event := <-events:
            report.Events = append(report.Events, event)
            
            // Real-time AI analysis for critical events
            if event.Severity == "critical" {
                analysis, err := a.analyzeSecurityEvent(ctx, event)
                if err != nil {
                    a.logger.Error(fmt.Sprintf("Failed to analyze critical event: %v", err))
                } else {
                    event.AIAnalysis = analysis
                }
            }
            
        case <-timeout:
            close(events)
            report.EndTime = time.Now()
            return report, nil
            
        case <-ctx.Done():
            close(events)
            return nil, ctx.Err()
        }
    }
}

// Security Context Gathering

func (a *AdvancedSecurityAgent) gatherSecurityContext(ctx context.Context) (*SecurityContext, error) {
    context := &SecurityContext{}
    
    // Gather system information
    sysInfo, err := a.gatherSystemInfo()
    if err != nil {
        return nil, fmt.Errorf("failed to gather system info: %w", err)
    }
    context.SystemInfo = sysInfo
    
    // Gather configuration files
    configs, err := a.gatherConfigurationFiles()
    if err != nil {
        a.logger.Warn(fmt.Sprintf("Failed to gather all configurations: %v", err))
    }
    context.Configurations = configs
    
    // Gather installed packages
    packages, err := a.gatherInstalledPackages()
    if err != nil {
        a.logger.Warn(fmt.Sprintf("Failed to gather packages: %v", err))
    }
    context.InstalledPkgs = packages
    
    // Gather running services
    services, err := a.gatherRunningServices()
    if err != nil {
        a.logger.Warn(fmt.Sprintf("Failed to gather services: %v", err))
    }
    context.RunningServices = services
    
    // Gather network configuration
    netConfig, err := a.gatherNetworkConfig()
    if err != nil {
        a.logger.Warn(fmt.Sprintf("Failed to gather network config: %v", err))
    }
    context.NetworkConfig = netConfig
    
    return context, nil
}

// AI Enhancement Methods

func (a *AdvancedSecurityAgent) enhanceSecurityPrompt(question string, context *SecurityContext) string {
    basePrompt := fmt.Sprintf(`You are an expert NixOS security analyst with deep knowledge of:
- NixOS security architecture and hardening
- Common vulnerabilities and attack vectors
- Security best practices and compliance standards
- Threat modeling and risk assessment

SECURITY CONTEXT:
- System: %s (NixOS %s)
- Packages: %d installed
- Services: %d running
- Network interfaces: %d active

QUESTION: %s

ANALYSIS REQUIREMENTS:
1. Assess security implications
2. Identify potential vulnerabilities
3. Provide specific NixOS configurations
4. Include threat mitigation strategies
5. Reference security standards when applicable
6. Give actionable recommendations

Focus on practical, implementable security solutions for NixOS environments.`,
        context.SystemInfo.Hostname,
        context.SystemInfo.NixOSVersion,
        len(context.InstalledPkgs),
        len(context.RunningServices),
        len(context.NetworkConfig.Interfaces),
        question)
    
    return basePrompt
}

func (a *AdvancedSecurityAgent) addSecurityInsights(response string, context *SecurityContext) string {
    insights := []string{}
    
    // Add automatic security checks
    if len(context.InstalledPkgs) > 500 {
        insights = append(insights, "⚠️  Large number of packages detected. Consider minimizing attack surface.")
    }
    
    if len(context.RunningServices) > 20 {
        insights = append(insights, "⚠️  Many services running. Review necessity of each service.")
    }
    
    // Add CVE alerts if any critical vulnerabilities found
    criticalCVEs := a.checkCriticalCVEs(context.InstalledPkgs)
    if len(criticalCVEs) > 0 {
        insights = append(insights, fmt.Sprintf("🚨 CRITICAL: %d packages with known CVEs detected!", len(criticalCVEs)))
    }
    
    if len(insights) > 0 {
        return response + "\n\n🔒 SECURITY INSIGHTS:\n" + strings.Join(insights, "\n")
    }
    
    return response
}

// Integration with Plugin System

// Plugin wrapper for the advanced security agent
type SecurityAgentPlugin struct {
    agent *AdvancedSecurityAgent
}

func (p *SecurityAgentPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
    switch operation {
    case "vulnerability-assessment":
        return p.agent.PerformVulnerabilityAssessment(ctx)
    case "generate-hardening-plan":
        target := params["target"].(string)
        return p.agent.GenerateSecurityHardening(ctx, target)
    case "monitor-events":
        duration := time.Duration(params["duration_minutes"].(float64)) * time.Minute
        return p.agent.MonitorSecurityEvents(ctx, duration)
    case "security-query":
        question := params["question"].(string)
        return p.agent.Query(ctx, question)
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
}

// Export for plugin system
var SecurityAgent SecurityAgentPlugin

func NewSecurityAgentPlugin() plugins.PluginInterface {
    // Initialize with default provider and config
    cfg, _ := config.LoadUserConfig()
    provider := ai.GetDefaultProvider(cfg)
    
    agent, err := NewAdvancedSecurityAgent(provider, cfg)
    if err != nil {
        panic(err) // Handle this properly in real implementation
    }
    
    return &SecurityAgentPlugin{agent: agent}
}
```

---

## 🎛️ **Agent Registration & Management**

### Auto-Registration System

```go
// agent_registry.go
package agents

import (
    "sync"
    "nix-ai-help/internal/ai/agent"
    "nix-ai-help/internal/plugins"
)

type AgentRegistry struct {
    agents    map[string]agent.Agent
    functions map[string]functionbase.FunctionInterface
    plugins   map[string]plugins.PluginInterface
    mutex     sync.RWMutex
}

var globalRegistry = &AgentRegistry{
    agents:    make(map[string]agent.Agent),
    functions: make(map[string]functionbase.FunctionInterface),
    plugins:   make(map[string]plugins.PluginInterface),
}

// Register different types of custom agents
func RegisterAgent(name string, agent agent.Agent) error {
    globalRegistry.mutex.Lock()
    defer globalRegistry.mutex.Unlock()
    globalRegistry.agents[name] = agent
    return nil
}

func RegisterFunction(fn functionbase.FunctionInterface) error {
    globalRegistry.mutex.Lock()
    defer globalRegistry.mutex.Unlock()
    globalRegistry.functions[fn.Name()] = fn
    return nil
}

func RegisterPlugin(plugin plugins.PluginInterface) error {
    globalRegistry.mutex.Lock()
    defer globalRegistry.mutex.Unlock()
    globalRegistry.plugins[plugin.Name()] = plugin
    return nil
}

// Auto-discovery for user agents
func DiscoverUserAgents(directories []string) error {
    for _, dir := range directories {
        // Scan for .so files (Go plugins)
        // Scan for .py files (Python agents)
        // Scan for .js files (Node.js agents)
        // Load and register discovered agents
    }
    return nil
}
```

---

## 📖 **Usage Examples**

### Command Line Integration

```bash
# Using function-based agents
nixai ask "Check my system security" --function security-analysis
nixai ask "DevOps deployment advice" --function devops-deploy-advice

# Using plugin agents
nixai plugin execute devops-agent troubleshoot-issue --params '{"issue_type":"deployment","symptoms":"build failed"}'
nixai plugin execute security-agent vulnerability-assessment

# Using direct agents (via ask command with agent specification)
nixai ask "Analyze security" --agent advanced-security --context-file /path/to/system-info.json
```

### Programmatic Usage

```go
// Use in Go applications
agent := NewAdvancedSecurityAgent(provider, config)
result, err := agent.PerformVulnerabilityAssessment(ctx)

// Use function interface
secFunc := NewSecurityAnalysisFunction(provider)
result, err := secFunc.Execute(ctx, params, options)

// Use plugin interface
plugin := NewDevOpsAgentPlugin()
result, err := plugin.Execute(ctx, "deploy-advice", params)
```

---

## 🎯 **Summary: Choose Your Approach**

| Approach | Complexity | Features | Best For |
|----------|------------|----------|----------|
| **AI Function** | Low | Parameter validation, auto-registration | Single-purpose agents |
| **Plugin Agent** | Medium | Full lifecycle, security, isolation | Multi-operation agents |
| **Direct Agent** | High | Full framework access, maximum power | Core system integration |

**Recommendation**: Start with **AI Functions** for simplicity, move to **Plugin Agents** for complex workflows, and use **Direct Agents** for core system integration.

All approaches are fully supported and can coexist in the same nixai installation! 🚀