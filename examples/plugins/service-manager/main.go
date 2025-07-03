package main

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/plugins"
)

// ServiceManagerPlugin manages systemd services with AI insights
type ServiceManagerPlugin struct {
	config        plugins.PluginConfig
	running       bool
	mutex         sync.RWMutex
	services      map[string]*ServiceInfo
	lastScan      time.Time
	monitoring    bool
	alertRules    []AlertRule
}

type ServiceInfo struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`        // active, inactive, failed, etc.
	SubState     string            `json:"sub_state"`     // running, exited, etc.
	LoadState    string            `json:"load_state"`    // loaded, not-found, etc.
	ActiveState  string            `json:"active_state"`  // active, inactive, failed
	Description  string            `json:"description"`
	MainPID      int               `json:"main_pid"`
	Memory       uint64            `json:"memory_usage"`
	CPUUsage     float64           `json:"cpu_usage"`
	StartTime    time.Time         `json:"start_time"`
	RestartCount int               `json:"restart_count"`
	Dependencies []string          `json:"dependencies"`
	Dependents   []string          `json:"dependents"`
	LogLines     []string          `json:"recent_logs"`
	Health       ServiceHealth     `json:"health"`
	Config       ServiceConfig     `json:"config"`
	Metrics      ServiceMetrics    `json:"metrics"`
}

type ServiceHealth struct {
	Status       string    `json:"status"`        // healthy, degraded, unhealthy
	LastCheck    time.Time `json:"last_check"`
	Issues       []string  `json:"issues"`
	Score        int       `json:"score"`         // 0-100
	Suggestions  []string  `json:"suggestions"`
}

type ServiceConfig struct {
	Type         string            `json:"type"`          // simple, forking, oneshot, etc.
	ExecStart    string            `json:"exec_start"`
	ExecReload   string            `json:"exec_reload"`
	User         string            `json:"user"`
	Group        string            `json:"group"`
	WorkingDir   string            `json:"working_directory"`
	Environment  map[string]string `json:"environment"`
	Restart      string            `json:"restart"`       // always, on-failure, etc.
	RestartSec   int               `json:"restart_sec"`
	TimeoutStart int               `json:"timeout_start"`
	TimeoutStop  int               `json:"timeout_stop"`
}

type ServiceMetrics struct {
	Uptime           time.Duration `json:"uptime"`
	TotalRestarts    int           `json:"total_restarts"`
	AvgMemoryUsage   uint64        `json:"avg_memory_usage"`
	PeakMemoryUsage  uint64        `json:"peak_memory_usage"`
	AvgCPUUsage      float64       `json:"avg_cpu_usage"`
	PeakCPUUsage     float64       `json:"peak_cpu_usage"`
	ErrorCount       int           `json:"error_count"`
	WarningCount     int           `json:"warning_count"`
	LastError        time.Time     `json:"last_error"`
	LastRestart      time.Time     `json:"last_restart"`
}

type AlertRule struct {
	ID          string    `json:"id"`
	ServiceName string    `json:"service_name"`
	Condition   string    `json:"condition"`    // memory_high, cpu_high, failed, restarting
	Threshold   float64   `json:"threshold"`
	Enabled     bool      `json:"enabled"`
	LastFired   time.Time `json:"last_fired"`
	Description string    `json:"description"`
}

type ServiceOperation struct {
	Action      string    `json:"action"`       // start, stop, restart, reload, enable, disable
	ServiceName string    `json:"service_name"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
}

// Metadata methods
func (p *ServiceManagerPlugin) Name() string        { return "service-manager" }
func (p *ServiceManagerPlugin) Version() string     { return "1.0.0" }
func (p *ServiceManagerPlugin) Description() string { return "AI-powered systemd service management and monitoring" }
func (p *ServiceManagerPlugin) Author() string      { return "NixAI Team" }
func (p *ServiceManagerPlugin) Repository() string  { return "https://github.com/nixai/plugins/service-manager" }
func (p *ServiceManagerPlugin) License() string     { return "MIT" }

func (p *ServiceManagerPlugin) Dependencies() []string {
	return []string{"systemctl", "journalctl"}
}

func (p *ServiceManagerPlugin) Capabilities() []string {
	return []string{
		"service-management",
		"service-monitoring",
		"health-analysis",
		"log-analysis",
		"dependency-tracking",
		"performance-monitoring",
	}
}

// Lifecycle methods
func (p *ServiceManagerPlugin) Initialize(ctx context.Context, config plugins.PluginConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.config = config
	p.services = make(map[string]*ServiceInfo)
	p.alertRules = []AlertRule{}
	p.monitoring = true
	
	// Set up default alert rules
	p.alertRules = append(p.alertRules, AlertRule{
		ID:          "high_memory",
		Condition:   "memory_high",
		Threshold:   500.0, // 500MB
		Enabled:     true,
		Description: "Alert when service memory usage is high",
	})
	
	p.alertRules = append(p.alertRules, AlertRule{
		ID:          "service_failed",
		Condition:   "failed",
		Threshold:   0,
		Enabled:     true,
		Description: "Alert when service fails",
	})
	
	return nil
}

func (p *ServiceManagerPlugin) Start(ctx context.Context) error {
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

func (p *ServiceManagerPlugin) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if !p.running {
		return fmt.Errorf("plugin is not running")
	}
	
	p.running = false
	return nil
}

func (p *ServiceManagerPlugin) Cleanup(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	p.services = nil
	p.alertRules = nil
	
	return nil
}

func (p *ServiceManagerPlugin) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

// Execution methods
func (p *ServiceManagerPlugin) Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error) {
	switch operation {
	case "list-services":
		return p.listServices(ctx, params)
	case "get-service":
		return p.getService(ctx, params)
	case "start-service":
		return p.startService(ctx, params)
	case "stop-service":
		return p.stopService(ctx, params)
	case "restart-service":
		return p.restartService(ctx, params)
	case "enable-service":
		return p.enableService(ctx, params)
	case "disable-service":
		return p.disableService(ctx, params)
	case "reload-service":
		return p.reloadService(ctx, params)
	case "analyze-service":
		return p.analyzeService(ctx, params)
	case "get-service-logs":
		return p.getServiceLogs(ctx, params)
	case "get-service-dependencies":
		return p.getServiceDependencies(ctx, params)
	case "health-check":
		return p.healthCheck(ctx, params)
	case "set-alert-rules":
		return p.setAlertRules(ctx, params)
	case "get-alerts":
		return p.getAlerts(ctx, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

func (p *ServiceManagerPlugin) GetOperations() []plugins.PluginOperation {
	return []plugins.PluginOperation{
		{
			Name:        "list-services",
			Description: "List all systemd services with status",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "status",
					Type:        "string",
					Description: "Filter by status (active, failed, inactive)",
					Required:    false,
				},
				{
					Name:        "pattern",
					Type:        "string", 
					Description: "Filter services by name pattern",
					Required:    false,
				},
			},
			ReturnType: "[]ServiceInfo",
			Tags:       []string{"services", "list"},
		},
		{
			Name:        "get-service",
			Description: "Get detailed information about a specific service",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Name of the service",
					Required:    true,
				},
			},
			ReturnType: "ServiceInfo",
			Tags:       []string{"services", "details"},
		},
		{
			Name:        "start-service",
			Description: "Start a systemd service",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Name of the service to start",
					Required:    true,
				},
			},
			ReturnType: "ServiceOperation",
			Tags:       []string{"services", "control"},
		},
		{
			Name:        "stop-service",
			Description: "Stop a systemd service",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Name of the service to stop",
					Required:    true,
				},
			},
			ReturnType: "ServiceOperation",
			Tags:       []string{"services", "control"},
		},
		{
			Name:        "restart-service",
			Description: "Restart a systemd service",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Name of the service to restart",
					Required:    true,
				},
			},
			ReturnType: "ServiceOperation",
			Tags:       []string{"services", "control"},
		},
		{
			Name:        "analyze-service",
			Description: "AI-powered service analysis and recommendations",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Name of the service to analyze",
					Required:    true,
				},
			},
			ReturnType: "ServiceAnalysis",
			Tags:       []string{"services", "analysis", "ai"},
		},
		{
			Name:        "get-service-logs",
			Description: "Get recent logs for a service",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Name of the service",
					Required:    true,
				},
				{
					Name:        "lines",
					Type:        "integer",
					Description: "Number of log lines to retrieve",
					Required:    false,
					Default:     50,
				},
			},
			ReturnType: "[]string",
			Tags:       []string{"services", "logs"},
		},
		{
			Name:        "health-check",
			Description: "Comprehensive service health assessment",
			Parameters: []plugins.PluginParameter{
				{
					Name:        "service_name",
					Type:        "string",
					Description: "Specific service to check (optional)",
					Required:    false,
				},
			},
			ReturnType: "HealthReport",
			Tags:       []string{"health", "monitoring"},
		},
	}
}

func (p *ServiceManagerPlugin) GetSchema(operation string) (*plugins.PluginSchema, error) {
	schemas := map[string]*plugins.PluginSchema{
		"list-services": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"status": {
					Type:        "string",
					Description: "Filter by status",
					Enum:        []string{"active", "failed", "inactive"},
				},
				"pattern": {
					Type:        "string",
					Description: "Name pattern filter",
				},
			},
		},
		"get-service": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"service_name": {
					Type:        "string",
					Description: "Service name",
				},
			},
			Required: []string{"service_name"},
		},
		"start-service": {
			Type: "object",
			Properties: map[string]plugins.PluginSchemaProperty{
				"service_name": {
					Type:        "string",
					Description: "Service name",
				},
			},
			Required: []string{"service_name"},
		},
	}
	
	if schema, exists := schemas[operation]; exists {
		return schema, nil
	}
	
	return nil, fmt.Errorf("unknown operation: %s", operation)
}

// Health and Status methods
func (p *ServiceManagerPlugin) HealthCheck(ctx context.Context) plugins.PluginHealth {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	status := plugins.HealthHealthy
	message := "Service manager running normally"
	var issues []plugins.HealthIssue
	
	if !p.running {
		status = plugins.HealthUnhealthy
		message = "Service manager is not running"
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityError,
			Component: "manager",
			Message:   "Manager service stopped",
			Timestamp: time.Now(),
		})
	}
	
	// Check for failed services
	failedCount := 0
	for _, service := range p.services {
		if service.Status == "failed" {
			failedCount++
		}
	}
	
	if failedCount > 0 {
		status = plugins.HealthDegraded
		message = fmt.Sprintf("%d services in failed state", failedCount)
		issues = append(issues, plugins.HealthIssue{
			Severity:  plugins.SeverityWarning,
			Component: "services",
			Message:   fmt.Sprintf("%d failed services", failedCount),
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

func (p *ServiceManagerPlugin) GetMetrics() plugins.PluginMetrics {
	return plugins.PluginMetrics{
		ExecutionCount:       0,
		TotalExecutionTime:   0,
		AverageExecutionTime: 0,
		LastExecutionTime:    p.lastScan,
		ErrorCount:           0,
		SuccessRate:          100.0,
		StartTime:            time.Now(),
		CustomMetrics: map[string]interface{}{
			"services_count":  len(p.services),
			"last_scan":       p.lastScan,
			"monitoring":      p.monitoring,
			"failed_services": p.countFailedServices(),
		},
	}
}

func (p *ServiceManagerPlugin) GetStatus() plugins.PluginStatus {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	state := plugins.StateRunning
	if !p.running {
		state = plugins.StateStopped
	}
	
	return plugins.PluginStatus{
		State:       state,
		Message:     "Service manager active",
		LastUpdated: time.Now(),
		Version:     p.Version(),
	}
}

// Operation implementations
func (p *ServiceManagerPlugin) listServices(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	err := p.scanServices()
	if err != nil {
		return nil, err
	}
	
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	var services []ServiceInfo
	
	// Apply filters
	status, hasStatus := params["status"].(string)
	pattern, hasPattern := params["pattern"].(string)
	
	for _, service := range p.services {
		// Status filter
		if hasStatus && service.Status != status {
			continue
		}
		
		// Pattern filter
		if hasPattern && !strings.Contains(service.Name, pattern) {
			continue
		}
		
		services = append(services, *service)
	}
	
	return services, nil
}

func (p *ServiceManagerPlugin) getService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	serviceName, ok := params["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("service_name parameter required")
	}
	
	err := p.updateServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	if service, exists := p.services[serviceName]; exists {
		return service, nil
	}
	
	return nil, fmt.Errorf("service %s not found", serviceName)
}

func (p *ServiceManagerPlugin) startService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return p.executeServiceCommand(ctx, params, "start")
}

func (p *ServiceManagerPlugin) stopService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return p.executeServiceCommand(ctx, params, "stop")
}

func (p *ServiceManagerPlugin) restartService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return p.executeServiceCommand(ctx, params, "restart")
}

func (p *ServiceManagerPlugin) enableService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return p.executeServiceCommand(ctx, params, "enable")
}

func (p *ServiceManagerPlugin) disableService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return p.executeServiceCommand(ctx, params, "disable")
}

func (p *ServiceManagerPlugin) reloadService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return p.executeServiceCommand(ctx, params, "reload")
}

func (p *ServiceManagerPlugin) executeServiceCommand(ctx context.Context, params map[string]interface{}, action string) (interface{}, error) {
	serviceName, ok := params["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("service_name parameter required")
	}
	
	startTime := time.Now()
	
	cmd := exec.Command("systemctl", action, serviceName)
	output, err := cmd.CombinedOutput()
	
	operation := ServiceOperation{
		Action:      action,
		ServiceName: serviceName,
		Success:     err == nil,
		Message:     string(output),
		Timestamp:   startTime,
		Duration:    time.Since(startTime),
	}
	
	if err != nil {
		operation.Message = fmt.Sprintf("Failed: %v\nOutput: %s", err, output)
	}
	
	// Update service info after operation
	p.updateServiceInfo(serviceName)
	
	return operation, nil
}

func (p *ServiceManagerPlugin) analyzeService(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	serviceName, ok := params["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("service_name parameter required")
	}
	
	err := p.updateServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	
	p.mutex.RLock()
	service, exists := p.services[serviceName]
	p.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	
	analysis := map[string]interface{}{
		"service":         service,
		"health_score":    p.calculateHealthScore(service),
		"issues":          p.identifyIssues(service),
		"recommendations": p.generateRecommendations(service),
		"performance":     p.assessPerformance(service),
		"security":        p.assessSecurity(service),
	}
	
	return analysis, nil
}

func (p *ServiceManagerPlugin) getServiceLogs(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	serviceName, ok := params["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("service_name parameter required")
	}
	
	lines := 50
	if val, ok := params["lines"].(float64); ok {
		lines = int(val)
	}
	
	cmd := exec.Command("journalctl", "-u", serviceName, "-n", fmt.Sprintf("%d", lines), "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %v", err)
	}
	
	logLines := strings.Split(string(output), "\n")
	
	// Filter out empty lines
	var filteredLogs []string
	for _, line := range logLines {
		if strings.TrimSpace(line) != "" {
			filteredLogs = append(filteredLogs, line)
		}
	}
	
	return filteredLogs, nil
}

func (p *ServiceManagerPlugin) getServiceDependencies(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	serviceName, ok := params["service_name"].(string)
	if !ok {
		return nil, fmt.Errorf("service_name parameter required")
	}
	
	// Get dependencies
	cmd := exec.Command("systemctl", "list-dependencies", serviceName, "--plain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies: %v", err)
	}
	
	dependencies := p.parseDependencies(string(output))
	
	return map[string]interface{}{
		"service":      serviceName,
		"dependencies": dependencies,
	}, nil
}

func (p *ServiceManagerPlugin) healthCheck(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if serviceName, ok := params["service_name"].(string); ok {
		// Single service health check
		return p.getSingleServiceHealth(serviceName)
	}
	
	// Overall system health check
	err := p.scanServices()
	if err != nil {
		return nil, err
	}
	
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	totalServices := len(p.services)
	activeServices := 0
	failedServices := 0
	var criticalIssues []string
	var warnings []string
	
	for _, service := range p.services {
		if service.Status == "active" {
			activeServices++
		} else if service.Status == "failed" {
			failedServices++
			criticalIssues = append(criticalIssues, fmt.Sprintf("Service %s has failed", service.Name))
		}
		
		// Check for high resource usage
		if service.Memory > 1024*1024*1024 { // 1GB
			warnings = append(warnings, fmt.Sprintf("Service %s using high memory: %d MB", service.Name, service.Memory/(1024*1024)))
		}
	}
	
	overallHealth := "healthy"
	if failedServices > 0 {
		overallHealth = "degraded"
	}
	if failedServices > totalServices/4 { // More than 25% failed
		overallHealth = "unhealthy"
	}
	
	return map[string]interface{}{
		"overall_health":   overallHealth,
		"total_services":   totalServices,
		"active_services":  activeServices,
		"failed_services":  failedServices,
		"critical_issues":  criticalIssues,
		"warnings":         warnings,
		"last_check":       time.Now(),
	}, nil
}

func (p *ServiceManagerPlugin) setAlertRules(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Implementation for setting alert rules
	return true, nil
}

func (p *ServiceManagerPlugin) getAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Implementation for getting current alerts
	return p.alertRules, nil
}

// Helper methods
func (p *ServiceManagerPlugin) scanServices() error {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list services: %v", err)
	}
	
	services := p.parseServiceList(string(output))
	
	p.mutex.Lock()
	for _, service := range services {
		p.services[service.Name] = service
	}
	p.lastScan = time.Now()
	p.mutex.Unlock()
	
	return nil
}

func (p *ServiceManagerPlugin) parseServiceList(output string) []*ServiceInfo {
	var services []*ServiceInfo
	
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`^\s*(\S+\.service)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)
	
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 6 {
			service := &ServiceInfo{
				Name:        strings.TrimSuffix(matches[1], ".service"),
				LoadState:   matches[2],
				ActiveState: matches[3],
				SubState:    matches[4],
				Description: matches[5],
				Status:      matches[3], // Use active state as status
			}
			services = append(services, service)
		}
	}
	
	return services
}

func (p *ServiceManagerPlugin) updateServiceInfo(serviceName string) error {
	// Get detailed service information
	cmd := exec.Command("systemctl", "show", serviceName, "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get service info: %v", err)
	}
	
	serviceInfo := p.parseServiceShow(string(output))
	serviceInfo.Name = serviceName
	
	// Get recent logs
	logs, _ := p.getRecentLogs(serviceName, 10)
	serviceInfo.LogLines = logs
	
	// Calculate health score
	serviceInfo.Health = p.calculateServiceHealth(serviceInfo)
	
	p.mutex.Lock()
	p.services[serviceName] = serviceInfo
	p.mutex.Unlock()
	
	return nil
}

func (p *ServiceManagerPlugin) parseServiceShow(output string) *ServiceInfo {
	service := &ServiceInfo{}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		switch key {
		case "ActiveState":
			service.ActiveState = value
			service.Status = value
		case "SubState":
			service.SubState = value
		case "LoadState":
			service.LoadState = value
		case "Description":
			service.Description = value
		case "MainPID":
			if pid, err := fmt.Sscanf(value, "%d", &service.MainPID); err == nil && pid == 1 {
				// Successfully parsed PID
			}
		case "ExecStart":
			if service.Config.ExecStart == "" {
				service.Config.ExecStart = value
			}
		case "User":
			service.Config.User = value
		case "Group":
			service.Config.Group = value
		}
	}
	
	return service
}

func (p *ServiceManagerPlugin) getRecentLogs(serviceName string, lines int) ([]string, error) {
	cmd := exec.Command("journalctl", "-u", serviceName, "-n", fmt.Sprintf("%d", lines), "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	
	logLines := strings.Split(string(output), "\n")
	var filteredLogs []string
	for _, line := range logLines {
		if strings.TrimSpace(line) != "" {
			filteredLogs = append(filteredLogs, line)
		}
	}
	
	return filteredLogs, nil
}

func (p *ServiceManagerPlugin) parseDependencies(output string) []string {
	var deps []string
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ".service") {
			deps = append(deps, strings.TrimSuffix(line, ".service"))
		}
	}
	
	return deps
}

func (p *ServiceManagerPlugin) calculateHealthScore(service *ServiceInfo) int {
	score := 100
	
	if service.Status == "failed" {
		score -= 50
	} else if service.Status != "active" {
		score -= 20
	}
	
	if service.RestartCount > 5 {
		score -= 15
	} else if service.RestartCount > 0 {
		score -= 5
	}
	
	if service.Memory > 1024*1024*1024 { // 1GB
		score -= 10
	}
	
	if service.CPUUsage > 80 {
		score -= 15
	} else if service.CPUUsage > 50 {
		score -= 5
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

func (p *ServiceManagerPlugin) calculateServiceHealth(service *ServiceInfo) ServiceHealth {
	score := p.calculateHealthScore(service)
	
	status := "healthy"
	if score < 70 {
		status = "degraded"
	}
	if score < 40 {
		status = "unhealthy"
	}
	
	var issues []string
	var suggestions []string
	
	if service.Status == "failed" {
		issues = append(issues, "Service is in failed state")
		suggestions = append(suggestions, "Check logs and restart service")
	}
	
	if service.RestartCount > 3 {
		issues = append(issues, "High restart count")
		suggestions = append(suggestions, "Investigate cause of frequent restarts")
	}
	
	if service.Memory > 512*1024*1024 { // 512MB
		issues = append(issues, "High memory usage")
		suggestions = append(suggestions, "Monitor memory usage and optimize if needed")
	}
	
	return ServiceHealth{
		Status:      status,
		LastCheck:   time.Now(),
		Issues:      issues,
		Score:       score,
		Suggestions: suggestions,
	}
}

func (p *ServiceManagerPlugin) countFailedServices() int {
	count := 0
	for _, service := range p.services {
		if service.Status == "failed" {
			count++
		}
	}
	return count
}

func (p *ServiceManagerPlugin) getSingleServiceHealth(serviceName string) (interface{}, error) {
	err := p.updateServiceInfo(serviceName)
	if err != nil {
		return nil, err
	}
	
	p.mutex.RLock()
	service, exists := p.services[serviceName]
	p.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	
	return service.Health, nil
}

func (p *ServiceManagerPlugin) identifyIssues(service *ServiceInfo) []string {
	var issues []string
	
	if service.Status == "failed" {
		issues = append(issues, "Service is currently failed")
	}
	
	if service.RestartCount > 5 {
		issues = append(issues, "High number of restarts detected")
	}
	
	if service.Memory > 1024*1024*1024 { // 1GB
		issues = append(issues, "High memory usage detected")
	}
	
	return issues
}

func (p *ServiceManagerPlugin) generateRecommendations(service *ServiceInfo) []string {
	var recommendations []string
	
	if service.Status == "failed" {
		recommendations = append(recommendations, "Check service logs for error details")
		recommendations = append(recommendations, "Verify service configuration")
	}
	
	if service.RestartCount > 3 {
		recommendations = append(recommendations, "Investigate root cause of restarts")
		recommendations = append(recommendations, "Consider increasing restart delay")
	}
	
	return recommendations
}

func (p *ServiceManagerPlugin) assessPerformance(service *ServiceInfo) map[string]interface{} {
	return map[string]interface{}{
		"memory_usage":   service.Memory,
		"cpu_usage":      service.CPUUsage,
		"restart_count":  service.RestartCount,
		"uptime":         service.Metrics.Uptime,
		"performance_score": p.calculatePerformanceScore(service),
	}
}

func (p *ServiceManagerPlugin) assessSecurity(service *ServiceInfo) map[string]interface{} {
	securityScore := 100
	var issues []string
	
	if service.Config.User == "root" {
		securityScore -= 20
		issues = append(issues, "Service running as root user")
	}
	
	return map[string]interface{}{
		"security_score": securityScore,
		"issues":         issues,
		"user":           service.Config.User,
		"group":          service.Config.Group,
	}
}

func (p *ServiceManagerPlugin) calculatePerformanceScore(service *ServiceInfo) int {
	score := 100
	
	if service.Memory > 1024*1024*1024 { // 1GB
		score -= 30
	}
	
	if service.CPUUsage > 80 {
		score -= 25
	}
	
	if service.RestartCount > 5 {
		score -= 20
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

func (p *ServiceManagerPlugin) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second) // Monitor every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.running {
				return
			}
			p.scanServices()
		}
	}
}

// Plugin entry point
var Plugin ServiceManagerPlugin

func init() {
	Plugin = ServiceManagerPlugin{}
}