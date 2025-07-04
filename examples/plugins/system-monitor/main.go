package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/plugins"
)

// SystemMonitorPlugin monitors system resources and health
type SystemMonitorPlugin struct {
	config      plugins.PluginConfig
	running     bool
	mutex       sync.RWMutex
	metrics     *SystemMetrics
	lastCheck   time.Time
	alerts      []Alert
	thresholds  SystemThresholds
}

type SystemMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	MemoryUsage  float64 `json:"memory_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	LoadAverage  float64 `json:"load_average"`
	Uptime       string  `json:"uptime"`
	Temperature  float64 `json:"temperature"`
	NetworkStats struct {
		BytesReceived uint64 `json:"bytes_received"`
		BytesSent     uint64 `json:"bytes_sent"`
	} `json:"network_stats"`
}

type SystemThresholds struct {
	CPUWarning    float64 `json:"cpu_warning"`
	CPUCritical   float64 `json:"cpu_critical"`
	MemoryWarning float64 `json:"memory_warning"`
	MemoryCritical float64 `json:"memory_critical"`
	DiskWarning   float64 `json:"disk_warning"`
	DiskCritical  float64 `json:"disk_critical"`
	TempWarning   float64 `json:"temp_warning"`
	TempCritical  float64 `json:"temp_critical"`
}

type Alert struct {
	Type      string    `json:"type"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// Metadata methods
func (p *SystemMonitorPlugin) Name() string        { return "system-monitor" }
func (p *SystemMonitorPlugin) Version() string     { return "1.0.0" }
func (p *SystemMonitorPlugin) Description() string { return "Real-time system monitoring and alerting plugin" }
func (p *SystemMonitorPlugin) Author() string      { return "NixAI Team" }
func (p *SystemMonitorPlugin) Repository() string  { return "https://github.com/nixai/plugins/system-monitor" }
func (p *SystemMonitorPlugin) License() string     { return "MIT" }

func (p *SystemMonitorPlugin) Dependencies() []string {
	return []string{}
}

func (p *SystemMonitorPlugin) Capabilities() []string {
	return []string{
		"system-monitoring",
		"resource-tracking",
		"alerting",
		"health-checks",
	}
}

// Lifecycle methods
func (p *SystemMonitorPlugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.config = config
	p.metrics = &SystemMetrics{}
	p.alerts = []Alert{}
	
	// Set default thresholds
	p.thresholds = SystemThresholds{
		CPUWarning:     70.0,
		CPUCritical:    90.0,
		MemoryWarning:  80.0,
		MemoryCritical: 95.0,
		DiskWarning:    85.0,
		DiskCritical:   95.0,
		TempWarning:    70.0,
		TempCritical:   85.0,
	}
	
	// Override with config values if provided
	if config.Configuration != nil {
		if cpu, ok := config.Configuration["cpu_warning"]; ok {
			if val, ok := cpu.(float64); ok {
				p.thresholds.CPUWarning = val
			}
		}
		if mem, ok := config.Configuration["memory_warning"]; ok {
			if val, ok := mem.(float64); ok {
				p.thresholds.MemoryWarning = val
			}
		}
	}
	
	return nil
}

func (p *SystemMonitorPlugin) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.running {
		return fmt.Errorf("plugin is already running")
	}
	
	// Start monitoring goroutine
	go p.monitoringLoop(ctx)
	
	p.running = true
	return nil
}

func (p *SystemMonitorPlugin) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if !p.running {
		return fmt.Errorf("plugin is not running")
	}
	
	p.running = false
	return nil
}

func (p *SystemMonitorPlugin) Cleanup(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.alerts = nil
	p.metrics = nil
	
	return nil
}

func (p *SystemMonitorPlugin) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

// Execution methods
func (p *SystemMonitorPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "get-metrics":
		return p.getMetrics(ctx, params)
	case "get-alerts":
		return p.getAlerts(ctx, params)
	case "health-check":
		return p.systemHealthCheck(ctx, params)
	case "set-thresholds":
		return p.setThresholds(ctx, params)
	case "get-thresholds":
		return p.getThresholds(ctx, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *SystemMonitorPlugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "get-metrics",
			Description: "Get current system metrics",
			Parameters:  []plugins.PluginParameter{},
			ReturnType:  "SystemMetrics",
			Tags:        []string{"monitoring", "metrics"},
		},
		{
			Name:        "get-alerts",
			Description: "Get current system alerts",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "level",
					Type:        "string",
					Description: "Filter alerts by level (warning, critical)",
					Required:    false,
				},
			},
			ReturnType: "[]Alert",
			Tags:       []string{"monitoring", "alerts"},
		},
		{
			Name:        "health-check",
			Description: "Perform comprehensive system health check",
			Parameters:  []plugins.PluginParameter{},
			ReturnType:  "HealthReport",
			Tags:        []string{"health", "diagnostics"},
		},
		{
			Name:        "set-thresholds",
			Description: "Update alert thresholds",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "thresholds",
					Type:        "object",
					Description: "Threshold configuration object",
					Required:    true,
				},
			},
			ReturnType: "bool",
			Tags:       []string{"configuration", "thresholds"},
		},
		{
			Name:        "get-thresholds",
			Description: "Get current alert thresholds",
			Parameters:  []plugins.PluginParameter{},
			ReturnType:  "SystemThresholds",
			Tags:        []string{"configuration", "thresholds"},
		},
	}
}

func (p *SystemMonitorPlugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	schemas := map[string]*plugins.PluginSchema{
		"get-metrics": {
			Type:       "object",
			Properties: map[string]plugins.PluginSchemaProperty{},
		},
		"get-alerts": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"level": {
					Type:        "string",
					Description: "Filter alerts by level",
					Enum:        []string{"warning", "critical"},
				},
			},
		},
		"health-check": {
			Type:       "object",
			Properties: map[string]plugins.PluginSchemaProperty{},
		},
		"set-thresholds": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"thresholds": {
					Type:        "object",
					Description: "Threshold configuration",
				},
			},
			Required: []string{"thresholds"},
		},
		"get-thresholds": {
			Type:       "object",
			Properties: map[string]plugins.PluginSchemaProperty{},
		},
	}
	
	if schema, exists := schemas[operation]; exists {
		return schema, nil
	}
	
	return nil, fmt.Errorf("unknown operation: %s", operation)
}

// Health and Status methods
func (p *SystemMonitorPlugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	status := plugins.HealthHealthy
	message := "System monitor running normally"
	var issues []plugins.HealthIssue
	
	if !p.running {
		status = plugins.HealthUnhealthy
		message = "System monitor is not running"
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityError,
			Component: "monitoring",
			Message:   "Monitor service stopped",
			Timestamp: time.Now(),
		})
	}
	
	// Check for critical alerts
	criticalAlerts := 0
	for _, alert := range p.alerts {
		if alert.Level == "critical" {
			criticalAlerts++
		}
	}
	
	if criticalAlerts > 0 {
		status = plugins.HealthDegraded
		message = fmt.Sprintf("System has %d critical alerts", criticalAlerts)
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityWarning,
			Component: "alerts",
			Message:   fmt.Sprintf("%d critical system alerts", criticalAlerts),
			Timestamp: time.Now(),
		})
	}
	
	return plugins.PluginHealth{
		Status:    status,
		Message:   message,
		LastCheck: time.Now(),
		Issues:    issues,
	}
}

func (p *SystemMonitorPlugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{
		ExecutionCount:       0,
		TotalExecutionTime:   0,
		AverageExecutionTime: 0,
		LastExecutionTime:    p.lastCheck,
		ErrorCount:           0,
		SuccessRate:          100.0,
		StartTime:            time.Now(),
		CustomMetrics: map[string]interface{}{
			"alerts_count":    len(p.alerts),
			"last_check":      p.lastCheck,
			"cpu_usage":       p.metrics.CPUUsage,
			"memory_usage":    p.metrics.MemoryUsage,
		},
	}
}

func (p *SystemMonitorPlugin) GetStatus() plugins.PluginStatus {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	state := plugins.StateRunning
	if !p.running {
		state = plugins.StateStopped
	}
	
	return plugins.PluginStatus{
		State:       state,
		Message:     "System monitor active",
		LastUpdated: time.Now(),
		Version:     p.Version(),
	}
}

// Operation implementations
func (p *SystemMonitorPlugin) getMetrics(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return p.metrics, nil
}

func (p *SystemMonitorPlugin) getAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	level, hasLevel := params["level"].(string)
	if !hasLevel {
		return p.alerts, nil
	}
	
	var filteredAlerts []Alert
	for _, alert := range p.alerts {
		if alert.Level == level {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}
	
	return filteredAlerts, nil
}

func (p *SystemMonitorPlugin) systemHealthCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.collectMetrics()
	
	health := map[string]interface{}{
		"overall_status": "healthy",
		"metrics":        p.metrics,
		"alerts":         p.alerts,
		"recommendations": []string{},
	}
	
	var recommendations []string
	
	if p.metrics.CPUUsage > p.thresholds.CPUCritical {
		health["overall_status"] = "critical"
		recommendations = append(recommendations, "CPU usage is critically high - consider closing resource-intensive applications")
	} else if p.metrics.CPUUsage > p.thresholds.CPUWarning {
		health["overall_status"] = "warning"
		recommendations = append(recommendations, "CPU usage is elevated - monitor running processes")
	}
	
	if p.metrics.MemoryUsage > p.thresholds.MemoryCritical {
		health["overall_status"] = "critical"
		recommendations = append(recommendations, "Memory usage is critically high - restart applications or add more RAM")
	} else if p.metrics.MemoryUsage > p.thresholds.MemoryWarning {
		if health["overall_status"] == "healthy" {
			health["overall_status"] = "warning"
		}
		recommendations = append(recommendations, "Memory usage is high - consider closing unused applications")
	}
	
	if p.metrics.DiskUsage > p.thresholds.DiskCritical {
		health["overall_status"] = "critical"
		recommendations = append(recommendations, "Disk space is critically low - clean up files or expand storage")
	}
	
	health["recommendations"] = recommendations
	
	return health, nil
}

func (p *SystemMonitorPlugin) setThresholds(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	thresholds, ok := params["thresholds"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid thresholds parameter")
	}
	
	// Update thresholds
	if cpu, ok := thresholds["cpu_warning"].(float64); ok {
		p.thresholds.CPUWarning = cpu
	}
	if cpu, ok := thresholds["cpu_critical"].(float64); ok {
		p.thresholds.CPUCritical = cpu
	}
	if mem, ok := thresholds["memory_warning"].(float64); ok {
		p.thresholds.MemoryWarning = mem
	}
	if mem, ok := thresholds["memory_critical"].(float64); ok {
		p.thresholds.MemoryCritical = mem
	}
	
	return true, nil
}

func (p *SystemMonitorPlugin) getThresholds(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return p.thresholds, nil
}

// Monitoring loop
func (p *SystemMonitorPlugin) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.running {
				return
			}
			p.collectMetrics()
			p.checkAlerts()
		}
	}
}

func (p *SystemMonitorPlugin) collectMetrics() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Collect CPU usage
	p.metrics.CPUUsage = p.getCPUUsage()
	
	// Collect memory usage
	p.metrics.MemoryUsage = p.getMemoryUsage()
	
	// Collect disk usage
	p.metrics.DiskUsage = p.getDiskUsage()
	
	// Collect load average
	p.metrics.LoadAverage = p.getLoadAverage()
	
	// Collect uptime
	p.metrics.Uptime = p.getUptime()
	
	p.lastCheck = time.Now()
}

func (p *SystemMonitorPlugin) getCPUUsage() float64 {
	// Read CPU usage from /proc/stat
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0.0
	}
	
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0.0
	}
	
	// Parse first line which contains overall CPU stats
	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0.0
	}
	
	// Calculate CPU usage (simplified)
	var idle, total float64
	for i := 1; i < len(fields); i++ {
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			continue
		}
		total += val
		if i == 4 { // idle time is the 4th field
			idle = val
		}
	}
	
	if total == 0 {
		return 0.0
	}
	
	usage := (total - idle) / total * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	
	return usage
}

func (p *SystemMonitorPlugin) getMemoryUsage() float64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0.0
	}
	
	lines := strings.Split(string(data), "\n")
	var memTotal, memAvailable float64
	
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				memTotal = val
			}
		case "MemAvailable:":
			if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
				memAvailable = val
			}
		}
	}
	
	if memTotal == 0 {
		return 0.0
	}
	
	usage := (memTotal - memAvailable) / memTotal * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	
	return usage
}

func (p *SystemMonitorPlugin) getDiskUsage() float64 {
	// Get disk usage using df command for root partition
	cmd := exec.Command("df", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0.0
	}
	
	// Parse the second line which contains the root filesystem info
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0.0
	}
	
	// Extract usage percentage (field 4, remove % suffix)
	usagePct := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.ParseFloat(usagePct, 64)
	if err != nil {
		return 0.0
	}
	
	return usage
}

func (p *SystemMonitorPlugin) getLoadAverage() float64 {
	// Read from /proc/loadavg
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) > 0 {
			if load, err := strconv.ParseFloat(parts[0], 64); err == nil {
				return load
			}
		}
	}
	return 0.0
}

func (p *SystemMonitorPlugin) getUptime() string {
	// Read from /proc/uptime
	if data, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) > 0 {
			if uptime, err := strconv.ParseFloat(parts[0], 64); err == nil {
				duration := time.Duration(uptime) * time.Second
				return duration.String()
			}
		}
	}
	return "unknown"
}

func (p *SystemMonitorPlugin) checkAlerts() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Clear old alerts
	p.alerts = []Alert{}
	
	// Check CPU alerts
	if p.metrics.CPUUsage > p.thresholds.CPUCritical {
		p.alerts = append(p.alerts, Alert{
			Type:      "cpu",
			Level:     "critical",
			Message:   fmt.Sprintf("CPU usage critical: %.1f%%", p.metrics.CPUUsage),
			Value:     p.metrics.CPUUsage,
			Threshold: p.thresholds.CPUCritical,
			Timestamp: time.Now(),
		})
	} else if p.metrics.CPUUsage > p.thresholds.CPUWarning {
		p.alerts = append(p.alerts, Alert{
			Type:      "cpu",
			Level:     "warning",
			Message:   fmt.Sprintf("CPU usage elevated: %.1f%%", p.metrics.CPUUsage),
			Value:     p.metrics.CPUUsage,
			Threshold: p.thresholds.CPUWarning,
			Timestamp: time.Now(),
		})
	}
	
	// Check memory alerts
	if p.metrics.MemoryUsage > p.thresholds.MemoryCritical {
		p.alerts = append(p.alerts, Alert{
			Type:      "memory",
			Level:     "critical",
			Message:   fmt.Sprintf("Memory usage critical: %.1f%%", p.metrics.MemoryUsage),
			Value:     p.metrics.MemoryUsage,
			Threshold: p.thresholds.MemoryCritical,
			Timestamp: time.Now(),
		})
	} else if p.metrics.MemoryUsage > p.thresholds.MemoryWarning {
		p.alerts = append(p.alerts, Alert{
			Type:      "memory",
			Level:     "warning",
			Message:   fmt.Sprintf("Memory usage elevated: %.1f%%", p.metrics.MemoryUsage),
			Value:     p.metrics.MemoryUsage,
			Threshold: p.thresholds.MemoryWarning,
			Timestamp: time.Now(),
		})
	}
	
	// Check disk alerts
	if p.metrics.DiskUsage > p.thresholds.DiskCritical {
		p.alerts = append(p.alerts, Alert{
			Type:      "disk",
			Level:     "critical",
			Message:   fmt.Sprintf("Disk usage critical: %.1f%%", p.metrics.DiskUsage),
			Value:     p.metrics.DiskUsage,
			Threshold: p.thresholds.DiskCritical,
			Timestamp: time.Now(),
		})
	} else if p.metrics.DiskUsage > p.thresholds.DiskWarning {
		p.alerts = append(p.alerts, Alert{
			Type:      "disk",
			Level:     "warning",
			Message:   fmt.Sprintf("Disk usage elevated: %.1f%%", p.metrics.DiskUsage),
			Value:     p.metrics.DiskUsage,
			Threshold: p.thresholds.DiskWarning,
			Timestamp: time.Now(),
		})
	}
}

// Plugin entry point
var Plugin SystemMonitorPlugin

func init() {
	Plugin = SystemMonitorPlugin{}
}