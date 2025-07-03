package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// Copy the required types from nixai's plugin interface
// This allows the plugin to be self-contained

type PluginInterface interface {
	Name() string
	Version() string
	Description() string
	Author() string
	Repository() string
	License() string
	Dependencies() []string
	Capabilities() []string
	Initialize(ctx context.Context, config PluginConfig) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Cleanup(ctx context.Context) error
	IsRunning() bool
	Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
	GetOperations() []PluginOperation
	GetSchema(operation string) (*PluginSchema, error)
	HealthCheck(ctx context.Context) PluginHealth
	GetMetrics() PluginMetrics
	GetStatus() PluginStatus
}

type PluginConfig struct {
	Name           string                 `json:"name" yaml:"name"`
	Enabled        bool                   `json:"enabled" yaml:"enabled"`
	Version        string                 `json:"version" yaml:"version"`
	Configuration  map[string]interface{} `json:"configuration" yaml:"configuration"`
	Environment    map[string]string      `json:"environment" yaml:"environment"`
	Resources      ResourceLimits         `json:"resources" yaml:"resources"`
	SecurityPolicy SecurityPolicy         `json:"security_policy" yaml:"security_policy"`
	UpdatePolicy   UpdatePolicy           `json:"update_policy" yaml:"update_policy"`
}

type ResourceLimits struct {
	MaxMemoryMB      int           `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent    int           `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	MaxExecutionTime time.Duration `json:"max_execution_time" yaml:"max_execution_time"`
	MaxFileSize      int64         `json:"max_file_size" yaml:"max_file_size"`
	AllowedPaths     []string      `json:"allowed_paths" yaml:"allowed_paths"`
	NetworkAccess    bool          `json:"network_access" yaml:"network_access"`
}

type SecurityPolicy struct {
	AllowFileSystem     bool         `json:"allow_file_system" yaml:"allow_file_system"`
	AllowNetwork        bool         `json:"allow_network" yaml:"allow_network"`
	AllowSystemCalls    bool         `json:"allow_system_calls" yaml:"allow_system_calls"`
	AllowedDomains      []string     `json:"allowed_domains" yaml:"allowed_domains"`
	RequiredPermissions []string     `json:"required_permissions" yaml:"required_permissions"`
	SandboxLevel        SandboxLevel `json:"sandbox_level" yaml:"sandbox_level"`
}

type UpdatePolicy struct {
	AutoUpdate         bool          `json:"auto_update" yaml:"auto_update"`
	UpdateChannel      string        `json:"update_channel" yaml:"update_channel"`
	CheckInterval      time.Duration `json:"check_interval" yaml:"check_interval"`
	RequireApproval    bool          `json:"require_approval" yaml:"require_approval"`
	BackupBeforeUpdate bool          `json:"backup_before_update" yaml:"backup_before_update"`
}

type SandboxLevel int

const (
	SandboxNone SandboxLevel = iota
	SandboxBasic
	SandboxStrict
	SandboxIsolated
)

type PluginOperation struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  []PluginParameter  `json:"parameters"`
	ReturnType  string             `json:"return_type"`
	Examples    []OperationExample `json:"examples"`
	Tags        []string           `json:"tags"`
	Deprecated  bool               `json:"deprecated"`
}

type PluginParameter struct {
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Default     interface{}          `json:"default,omitempty"`
	Validation  *ParameterValidation `json:"validation,omitempty"`
}

type ParameterValidation struct {
	MinLength int      `json:"min_length,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
}

type OperationExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    interface{}            `json:"expected"`
}

type PluginSchema struct {
	Type       string                          `json:"type"`
	Properties map[string]PluginSchemaProperty `json:"properties"`
	Required   []string                        `json:"required"`
	Examples   []map[string]interface{}        `json:"examples"`
}

type PluginSchemaProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
}

type PluginHealth struct {
	Status          HealthStatus  `json:"status"`
	Message         string        `json:"message"`
	LastCheck       time.Time     `json:"last_check"`
	Uptime          time.Duration `json:"uptime"`
	Issues          []HealthIssue `json:"issues"`
	Recommendations []string      `json:"recommendations"`
}

type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthDegraded
	HealthUnhealthy
	HealthCritical
)

type HealthIssue struct {
	Severity   IssueSeverity `json:"severity"`
	Component  string        `json:"component"`
	Message    string        `json:"message"`
	Timestamp  time.Time     `json:"timestamp"`
	Resolution string        `json:"resolution,omitempty"`
}

type IssueSeverity int

const (
	SeverityInfo IssueSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

type PluginMetrics struct {
	ExecutionCount       int64                  `json:"execution_count"`
	TotalExecutionTime   time.Duration          `json:"total_execution_time"`
	AverageExecutionTime time.Duration          `json:"average_execution_time"`
	LastExecutionTime    time.Time              `json:"last_execution_time"`
	ErrorCount           int64                  `json:"error_count"`
	SuccessRate          float64                `json:"success_rate"`
	MemoryUsage          int64                  `json:"memory_usage_bytes"`
	CPUUsage             float64                `json:"cpu_usage_percent"`
	StartTime            time.Time              `json:"start_time"`
	CustomMetrics        map[string]interface{} `json:"custom_metrics"`
}

type PluginStatus struct {
	State       PluginState `json:"state"`
	Message     string      `json:"message"`
	LastUpdated time.Time   `json:"last_updated"`
	Version     string      `json:"version"`
	ConfigHash  string      `json:"config_hash"`
	ProcessID   int         `json:"process_id,omitempty"`
}

type PluginState int

const (
	StateUnknown PluginState = iota
	StateLoading
	StateInitializing
	StateRunning
	StateStopping
	StateStopped
	StateError
	StateDisabled
)

// SimpleSystemInfoPlugin provides basic system information
type SimpleSystemInfoPlugin struct {
	startTime time.Time
	execCount int64
	errorCount int64
	running   bool
	config    PluginConfig
}

// Plugin interface implementation
func (p *SimpleSystemInfoPlugin) Name() string {
	return "simple-system-info"
}

func (p *SimpleSystemInfoPlugin) Version() string {
	return "1.0.0"
}

func (p *SimpleSystemInfoPlugin) Description() string {
	return "Simple system information plugin"
}

func (p *SimpleSystemInfoPlugin) Author() string {
	return "NixAI Team"
}

func (p *SimpleSystemInfoPlugin) Repository() string {
	return "https://github.com/nixai/plugins/simple-system-info"
}

func (p *SimpleSystemInfoPlugin) License() string {
	return "MIT"
}

func (p *SimpleSystemInfoPlugin) Dependencies() []string {
	return []string{}
}

func (p *SimpleSystemInfoPlugin) Capabilities() []string {
	return []string{"system-info", "health-monitoring"}
}

func (p *SimpleSystemInfoPlugin) Initialize(ctx context.Context, config PluginConfig) error {
	p.startTime = time.Now()
	p.config = config
	return nil
}

func (p *SimpleSystemInfoPlugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

func (p *SimpleSystemInfoPlugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

func (p *SimpleSystemInfoPlugin) Cleanup(ctx context.Context) error {
	return nil
}

func (p *SimpleSystemInfoPlugin) IsRunning() bool {
	return p.running
}

func (p *SimpleSystemInfoPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	p.execCount++
	
	switch operation {
	case "get-info":
		return p.getSystemInfo(), nil
	case "get-memory":
		return p.getMemoryInfo(), nil
	case "get-cpu":
		return p.getCPUInfo(), nil
	default:
		p.errorCount++
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *SimpleSystemInfoPlugin) GetOperations() []PluginOperation {
	return []PluginOperation{
		{
			Name:        "get-info",
			Description: "Get basic system information",
			ReturnType:  "object",
		},
		{
			Name:        "get-memory",
			Description: "Get memory usage information",
			ReturnType:  "object",
		},
		{
			Name:        "get-cpu",
			Description: "Get CPU information",
			ReturnType:  "object",
		},
	}
}

func (p *SimpleSystemInfoPlugin) GetSchema(operation string) (*PluginSchema, error) {
	switch operation {
	case "get-info":
		return &PluginSchema{
			Type: "object",
			Properties: map[string]PluginSchemaProperty{
				"hostname":   {Type: "string", Description: "System hostname"},
				"os":         {Type: "string", Description: "Operating system"},
				"arch":       {Type: "string", Description: "System architecture"},
				"go_version": {Type: "string", Description: "Go version"},
				"num_cpu":    {Type: "integer", Description: "Number of CPUs"},
				"timestamp":  {Type: "string", Description: "Timestamp of the info"},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *SimpleSystemInfoPlugin) HealthCheck(ctx context.Context) PluginHealth {
	return PluginHealth{
		Status:    HealthHealthy,
		Message:   "Plugin is running normally",
		LastCheck: time.Now(),
		Uptime:    time.Since(p.startTime),
		Issues:    []HealthIssue{},
		Recommendations: []string{},
	}
}

func (p *SimpleSystemInfoPlugin) GetMetrics() PluginMetrics {
	successRate := float64(p.execCount-p.errorCount) / float64(p.execCount)
	if p.execCount == 0 {
		successRate = 1.0
	}
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return PluginMetrics{
		ExecutionCount:   p.execCount,
		ErrorCount:      p.errorCount,
		SuccessRate:     successRate,
		MemoryUsage:     int64(m.Alloc),
		StartTime:       p.startTime,
		CustomMetrics:   make(map[string]interface{}),
	}
}

func (p *SimpleSystemInfoPlugin) GetStatus() PluginStatus {
	state := StateStopped
	if p.running {
		state = StateRunning
	}
	
	return PluginStatus{
		State:       state,
		Message:     "Plugin operational",
		LastUpdated: time.Now(),
		Version:     p.Version(),
	}
}

// Helper methods
func (p *SimpleSystemInfoPlugin) getSystemInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	
	return map[string]interface{}{
		"hostname":    hostname,
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"go_version":  runtime.Version(),
		"num_cpu":     runtime.NumCPU(),
		"timestamp":   time.Now().Format(time.RFC3339),
	}
}

func (p *SimpleSystemInfoPlugin) getMemoryInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"alloc_mb":        bToMb(m.Alloc),
		"total_alloc_mb":  bToMb(m.TotalAlloc),
		"sys_mb":          bToMb(m.Sys),
		"num_gc":          m.NumGC,
		"timestamp":       time.Now().Format(time.RFC3339),
	}
}

func (p *SimpleSystemInfoPlugin) getCPUInfo() map[string]interface{} {
	return map[string]interface{}{
		"num_cpu":      runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"timestamp":    time.Now().Format(time.RFC3339),
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Plugin export - required by nixai plugin loader
func NewPlugin() PluginInterface {
	return &SimpleSystemInfoPlugin{}
}

// Main function for standalone testing
func main() {
	if len(os.Args) > 1 {
		operation := os.Args[1]
		
		plugin := &SimpleSystemInfoPlugin{}
		plugin.Initialize(context.Background(), PluginConfig{})
		plugin.Start(context.Background())
		
		result, err := plugin.Execute(context.Background(), operation, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("Simple System Info Plugin")
		fmt.Println("Available operations:")
		plugin := &SimpleSystemInfoPlugin{}
		for _, op := range plugin.GetOperations() {
			fmt.Printf("  %s: %s\n", op.Name, op.Description)
		}
	}
}