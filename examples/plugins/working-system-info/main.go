package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Use exact types from the nixai plugin system by defining them with identical structure
// This should match the internal plugin interface exactly

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

// WorkingSystemInfoPlugin - A comprehensive system information plugin
type WorkingSystemInfoPlugin struct {
	startTime time.Time
	execCount int64
	errorCount int64
	running   bool
	config    PluginConfig
}

func NewPlugin() PluginInterface {
	return &WorkingSystemInfoPlugin{}
}

func (p *WorkingSystemInfoPlugin) Name() string {
	return "working-system-info"
}

func (p *WorkingSystemInfoPlugin) Version() string {
	return "1.0.0"
}

func (p *WorkingSystemInfoPlugin) Description() string {
	return "Comprehensive system information and monitoring plugin"
}

func (p *WorkingSystemInfoPlugin) Author() string {
	return "NixAI Team"
}

func (p *WorkingSystemInfoPlugin) Repository() string {
	return "https://github.com/nixai/plugins/working-system-info"
}

func (p *WorkingSystemInfoPlugin) License() string {
	return "MIT"
}

func (p *WorkingSystemInfoPlugin) Dependencies() []string {
	return []string{}
}

func (p *WorkingSystemInfoPlugin) Capabilities() []string {
	return []string{"system-info", "monitoring", "health-check", "performance"}
}

func (p *WorkingSystemInfoPlugin) Initialize(ctx context.Context, config PluginConfig) error {
	p.startTime = time.Now()
	p.config = config
	return nil
}

func (p *WorkingSystemInfoPlugin) Start(ctx context.Context) error {
	p.running = true
	return nil
}

func (p *WorkingSystemInfoPlugin) Stop(ctx context.Context) error {
	p.running = false
	return nil
}

func (p *WorkingSystemInfoPlugin) Cleanup(ctx context.Context) error {
	return nil
}

func (p *WorkingSystemInfoPlugin) IsRunning() bool {
	return p.running
}

func (p *WorkingSystemInfoPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	p.execCount++
	
	switch operation {
	case "get-info":
		return p.getSystemInfo(), nil
	case "get-memory":
		return p.getMemoryInfo(), nil
	case "get-cpu":
		return p.getCPUInfo(), nil
	case "get-disk":
		return p.getDiskInfo(), nil
	case "get-network":
		return p.getNetworkInfo(), nil
	case "health-check":
		return p.systemHealthCheck(), nil
	case "get-processes":
		return p.getTopProcesses(), nil
	default:
		p.errorCount++
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *WorkingSystemInfoPlugin) GetOperations() []PluginOperation {
	return []PluginOperation{
		{
			Name:        "get-info",
			Description: "Get basic system information",
			ReturnType:  "object",
			Tags:        []string{"info", "system"},
		},
		{
			Name:        "get-memory",
			Description: "Get detailed memory usage information",
			ReturnType:  "object",
			Tags:        []string{"memory", "performance"},
		},
		{
			Name:        "get-cpu",
			Description: "Get CPU information and usage",
			ReturnType:  "object",
			Tags:        []string{"cpu", "performance"},
		},
		{
			Name:        "get-disk",
			Description: "Get disk usage information",
			ReturnType:  "object",
			Tags:        []string{"disk", "storage"},
		},
		{
			Name:        "get-network",
			Description: "Get network interface information",
			ReturnType:  "object",
			Tags:        []string{"network", "interfaces"},
		},
		{
			Name:        "health-check",
			Description: "Perform comprehensive system health check",
			ReturnType:  "object",
			Tags:        []string{"health", "monitoring"},
		},
		{
			Name:        "get-processes",
			Description: "Get top processes by CPU and memory usage",
			ReturnType:  "object",
			Tags:        []string{"processes", "performance"},
		},
	}
}

func (p *WorkingSystemInfoPlugin) GetSchema(operation string) (*PluginSchema, error) {
	switch operation {
	case "get-info":
		return &PluginSchema{
			Type: "object",
			Properties: map[string]PluginSchemaProperty{
				"hostname":   {Type: "string", Description: "System hostname"},
				"os":         {Type: "string", Description: "Operating system"},
				"arch":       {Type: "string", Description: "System architecture"},
				"uptime":     {Type: "string", Description: "System uptime"},
				"kernel":     {Type: "string", Description: "Kernel version"},
			},
		}, nil
	default:
		return nil, fmt.Errorf("schema not available for operation: %s", operation)
	}
}

func (p *WorkingSystemInfoPlugin) HealthCheck(ctx context.Context) PluginHealth {
	return PluginHealth{
		Status:          HealthHealthy,
		Message:         "Plugin is running normally",
		LastCheck:       time.Now(),
		Uptime:          time.Since(p.startTime),
		Issues:          []HealthIssue{},
		Recommendations: []string{},
	}
}

func (p *WorkingSystemInfoPlugin) GetMetrics() PluginMetrics {
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

func (p *WorkingSystemInfoPlugin) GetStatus() PluginStatus {
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

// Implementation methods
func (p *WorkingSystemInfoPlugin) getSystemInfo() map[string]interface{} {
	hostname, _ := os.Hostname()
	
	uptime := ""
	if uptimeBytes, err := os.ReadFile("/proc/uptime"); err == nil {
		uptimeStr := strings.Fields(string(uptimeBytes))[0]
		if uptimeFloat, err := strconv.ParseFloat(uptimeStr, 64); err == nil {
			uptime = time.Duration(uptimeFloat * float64(time.Second)).String()
		}
	}
	
	kernel := ""
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		kernel = strings.TrimSpace(string(out))
	}
	
	return map[string]interface{}{
		"hostname":    hostname,
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"go_version":  runtime.Version(),
		"num_cpu":     runtime.NumCPU(),
		"uptime":      uptime,
		"kernel":      kernel,
		"timestamp":   time.Now().Format(time.RFC3339),
	}
}

func (p *WorkingSystemInfoPlugin) getMemoryInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Try to get system memory info
	systemMem := map[string]interface{}{}
	if meminfo, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(meminfo), "\n")
		for _, line := range lines {
			if strings.Contains(line, "MemTotal:") || strings.Contains(line, "MemAvailable:") || strings.Contains(line, "MemFree:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if val, err := strconv.Atoi(fields[1]); err == nil {
						systemMem[strings.TrimSuffix(fields[0], ":")] = val * 1024 // Convert kB to bytes
					}
				}
			}
		}
	}
	
	return map[string]interface{}{
		"go_runtime": map[string]interface{}{
			"alloc_bytes":     m.Alloc,
			"alloc_mb":        bToMb(m.Alloc),
			"total_alloc_mb":  bToMb(m.TotalAlloc),
			"sys_mb":          bToMb(m.Sys),
			"num_gc":          m.NumGC,
		},
		"system":    systemMem,
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

func (p *WorkingSystemInfoPlugin) getCPUInfo() map[string]interface{} {
	cpuInfo := map[string]interface{}{
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
	}
	
	// Try to get CPU model info
	if cpuinfoBytes, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(cpuinfoBytes), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					cpuInfo["model"] = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}
	
	cpuInfo["timestamp"] = time.Now().Format(time.RFC3339)
	return cpuInfo
}

func (p *WorkingSystemInfoPlugin) getDiskInfo() map[string]interface{} {
	diskInfo := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	// Try to get disk usage with df command
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 4 {
				diskInfo["root_filesystem"] = map[string]interface{}{
					"size":      fields[1],
					"used":      fields[2],
					"available": fields[3],
					"use_pct":   fields[4],
				}
			}
		}
	}
	
	return diskInfo
}

func (p *WorkingSystemInfoPlugin) getNetworkInfo() map[string]interface{} {
	netInfo := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	// Try to get network interfaces
	if out, err := exec.Command("ip", "addr", "show").Output(); err == nil {
		netInfo["interfaces_output"] = string(out)
	}
	
	return netInfo
}

func (p *WorkingSystemInfoPlugin) systemHealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "healthy",
		"checks":    []map[string]interface{}{},
	}
	
	checks := []map[string]interface{}{}
	
	// Memory check
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryMB := bToMb(m.Alloc)
	
	memCheck := map[string]interface{}{
		"name":   "memory_usage",
		"status": "ok",
		"value":  memoryMB,
		"unit":   "MB",
	}
	if memoryMB > 1000 { // More than 1GB
		memCheck["status"] = "warning"
		memCheck["message"] = "High memory usage"
	}
	checks = append(checks, memCheck)
	
	// Disk space check
	if out, err := exec.Command("df", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				usePct := strings.TrimSuffix(fields[4], "%")
				if pct, err := strconv.Atoi(usePct); err == nil {
					diskCheck := map[string]interface{}{
						"name":   "disk_usage",
						"status": "ok",
						"value":  pct,
						"unit":   "%",
					}
					if pct > 80 {
						diskCheck["status"] = "warning"
						diskCheck["message"] = "High disk usage"
					}
					if pct > 95 {
						diskCheck["status"] = "critical"
						diskCheck["message"] = "Critical disk usage"
					}
					checks = append(checks, diskCheck)
				}
			}
		}
	}
	
	health["checks"] = checks
	return health
}

func (p *WorkingSystemInfoPlugin) getTopProcesses() map[string]interface{} {
	processes := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	// Get top processes by CPU
	if out, err := exec.Command("ps", "aux", "--sort=-pcpu").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			// Take first 10 processes (excluding header)
			topProcs := []string{}
			for i := 1; i < len(lines) && i <= 10; i++ {
				if strings.TrimSpace(lines[i]) != "" {
					topProcs = append(topProcs, lines[i])
				}
			}
			processes["top_cpu"] = topProcs
		}
	}
	
	return processes
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// Main function for standalone testing
func main() {
	if len(os.Args) > 1 {
		operation := os.Args[1]
		
		plugin := &WorkingSystemInfoPlugin{}
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
		fmt.Println("Working System Info Plugin")
		fmt.Println("Available operations:")
		plugin := &WorkingSystemInfoPlugin{}
		for _, op := range plugin.GetOperations() {
			fmt.Printf("  %s: %s\n", op.Name, op.Description)
		}
	}
}