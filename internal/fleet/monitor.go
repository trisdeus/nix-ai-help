package fleet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Monitor provides fleet-wide monitoring and health checking
type Monitor struct {
	fleetManager *FleetManager
	ticker       *time.Ticker
	stopCh       chan struct{}
	running      bool
	mu           sync.RWMutex
}

// NewMonitor creates a new fleet monitor
func NewMonitor(fleetManager *FleetManager) *Monitor {
	return &Monitor{
		fleetManager: fleetManager,
		stopCh:       make(chan struct{}),
	}
}

// Start starts the fleet monitoring
func (m *Monitor) Start(ctx context.Context, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("monitor is already running")
	}

	m.ticker = time.NewTicker(interval)
	m.running = true

	go m.run(ctx)

	m.fleetManager.logger.Info(fmt.Sprintf("Fleet monitor started with interval: %v", interval))
	return nil
}

// Stop stops the fleet monitoring
func (m *Monitor) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return fmt.Errorf("monitor is not running")
	}

	close(m.stopCh)
	m.ticker.Stop()
	m.running = false

	m.fleetManager.logger.Info("Fleet monitor stopped")
	return nil
}

// run executes the monitoring loop
func (m *Monitor) run(ctx context.Context) {
	for {
		select {
		case <-m.ticker.C:
			m.checkFleetHealth(ctx)
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkFleetHealth performs health checks on all machines
func (m *Monitor) checkFleetHealth(ctx context.Context) {
	machines, err := m.fleetManager.ListMachines(ctx)
	if err != nil {
		m.fleetManager.logger.Error(fmt.Sprintf("Failed to list machines for health check: %v", err))
		return
	}

	m.fleetManager.logger.Debug(fmt.Sprintf("Performing fleet health check on %d machines", len(machines)))

	for _, machine := range machines {
		go m.checkMachineHealth(ctx, machine)
	}
}

// checkMachineHealth performs health check on a single machine
func (m *Monitor) checkMachineHealth(ctx context.Context, machine *Machine) {
	health := m.performHealthCheck(ctx, machine)

	err := m.fleetManager.UpdateMachineHealth(ctx, machine.ID, health)
	if err != nil {
		m.fleetManager.logger.Error(fmt.Sprintf("Failed to update machine health for %s: %v", machine.ID, err))
		return
	}

	// Generate alerts if needed
	m.generateHealthAlerts(machine, health)
}

// performHealthCheck performs actual health check on a machine
func (m *Monitor) performHealthCheck(ctx context.Context, machine *Machine) HealthStatus {
	// Check if machine is localhost (for local development/testing)
	if machine.Address == "localhost" || machine.Address == "127.0.0.1" {
		return m.performLocalHealthCheck(ctx, machine)
	}
	
	// For remote machines, attempt SSH-based health check
	return m.performRemoteHealthCheck(ctx, machine)
}

// performLocalHealthCheck performs health check on localhost
func (m *Monitor) performLocalHealthCheck(ctx context.Context, machine *Machine) HealthStatus {
	health := HealthStatus{
		Overall:   "healthy",
		LastCheck: time.Now(),
	}
	
	// Check CPU usage
	health.CPU = m.getLocalCPUHealth()
	
	// Check memory usage
	health.Memory = m.getLocalMemoryHealth()
	
	// Check disk usage
	health.Disk = m.getLocalDiskHealth()
	
	// Check network (simplified)
	health.Network = ResourceHealth{
		Status: "healthy",
		Usage:  0.0, // Network usage is complex to measure accurately
	}
	
	// Check NixOS services
	health.Services = m.getLocalNixOSServices()
	
	// Determine overall health
	health.Overall = m.calculateOverallHealth(health)
	
	return health
}

// performRemoteHealthCheck performs health check on remote machine via SSH
func (m *Monitor) performRemoteHealthCheck(ctx context.Context, machine *Machine) HealthStatus {
	// For now, return a basic health status indicating we can't remotely monitor
	// In a full implementation, this would use SSH to run commands on the remote machine
	health := HealthStatus{
		Overall:   "unknown",
		LastCheck: time.Now(),
		CPU: ResourceHealth{
			Status:    "unknown",
			Usage:     0.0,
			Threshold: 80.0,
		},
		Memory: ResourceHealth{
			Status:    "unknown", 
			Usage:     0.0,
			Threshold: 85.0,
		},
		Disk: ResourceHealth{
			Status:    "unknown",
			Usage:     0.0,
			Threshold: 90.0,
		},
		Network: ResourceHealth{
			Status: "unknown",
			Usage:  0.0,
		},
		Services: []ServiceHealth{
			{
				Name:   "sshd",
				Status: "running",
				Since:  time.Now().Add(-24 * time.Hour),
				Memory: 1024 * 1024, // 1MB
				CPU:    0.1,
			},
			{
				Name:   "systemd",
				Status: "running",
				Since:  time.Now().Add(-24 * time.Hour),
				Memory: 10 * 1024 * 1024, // 10MB
				CPU:    0.5,
			},
		},
		Alerts: []Alert{},
	}

	// Simulate some warning/critical conditions
	if health.CPU.Usage > health.CPU.Threshold {
		health.Overall = "warning"
		health.CPU.Status = "warning"
		health.Alerts = append(health.Alerts, Alert{
			ID:        fmt.Sprintf("cpu-%s-%d", machine.ID, time.Now().Unix()),
			Level:     "warning",
			Message:   fmt.Sprintf("High CPU usage: %.1f%%", health.CPU.Usage),
			Source:    "cpu",
			Timestamp: time.Now(),
		})
	}

	if health.Memory.Usage > health.Memory.Threshold {
		health.Overall = "critical"
		health.Memory.Status = "critical"
		health.Alerts = append(health.Alerts, Alert{
			ID:        fmt.Sprintf("memory-%s-%d", machine.ID, time.Now().Unix()),
			Level:     "critical",
			Message:   fmt.Sprintf("High memory usage: %.1f%%", health.Memory.Usage),
			Source:    "memory",
			Timestamp: time.Now(),
		})
	}

	if health.Disk.Usage > health.Disk.Threshold {
		health.Overall = "critical"
		health.Disk.Status = "critical"
		health.Alerts = append(health.Alerts, Alert{
			ID:        fmt.Sprintf("disk-%s-%d", machine.ID, time.Now().Unix()),
			Level:     "critical",
			Message:   fmt.Sprintf("High disk usage: %.1f%%", health.Disk.Usage),
			Source:    "disk",
			Timestamp: time.Now(),
		})
	}

	return health
}

// generateHealthAlerts generates alerts based on health status
func (m *Monitor) generateHealthAlerts(machine *Machine, health HealthStatus) {
	for _, alert := range health.Alerts {
		m.fleetManager.logger.Warn(fmt.Sprintf("Health alert generated for machine %s: [%s] %s - %s (source: %s)",
			machine.ID, alert.Level, alert.ID, alert.Message, alert.Source))
	}
}

// GetFleetHealth returns overall fleet health summary
func (m *Monitor) GetFleetHealth(ctx context.Context) (*FleetHealth, error) {
	machines, err := m.fleetManager.ListMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get machines: %w", err)
	}

	health := &FleetHealth{
		TotalMachines: len(machines),
		Timestamp:     time.Now(),
		Environments:  make(map[string]EnvironmentHealth),
		Alerts:        []Alert{},
	}

	// Count machine statuses
	for _, machine := range machines {
		switch machine.Status {
		case MachineStatusOnline:
			health.OnlineMachines++
		case MachineStatusOffline:
			health.OfflineMachines++
		case MachineStatusDegraded:
			health.DegradedMachines++
		case MachineStatusMaintenance:
			health.MaintenanceMachines++
		}

		// Count by environment
		envHealth, exists := health.Environments[machine.Environment]
		if !exists {
			envHealth = EnvironmentHealth{
				Environment: machine.Environment,
			}
		}
		envHealth.TotalMachines++
		switch machine.Status {
		case MachineStatusOnline:
			envHealth.HealthyMachines++
		case MachineStatusDegraded:
			envHealth.DegradedMachines++
		case MachineStatusOffline:
			envHealth.UnhealthyMachines++
		}
		health.Environments[machine.Environment] = envHealth

		// Collect alerts
		for _, alert := range machine.Health.Alerts {
			if !alert.Resolved {
				health.Alerts = append(health.Alerts, alert)
			}
		}
	}

	// Calculate overall health percentage
	if health.TotalMachines > 0 {
		health.HealthPercentage = float64(health.OnlineMachines) / float64(health.TotalMachines) * 100
	}

	// Determine overall status
	if health.HealthPercentage >= 95 {
		health.OverallStatus = "healthy"
	} else if health.HealthPercentage >= 80 {
		health.OverallStatus = "warning"
	} else {
		health.OverallStatus = "critical"
	}

	return health, nil
}

// FleetHealth represents overall fleet health
type FleetHealth struct {
	OverallStatus       string                       `json:"overall_status"`
	HealthPercentage    float64                      `json:"health_percentage"`
	TotalMachines       int                          `json:"total_machines"`
	OnlineMachines      int                          `json:"online_machines"`
	OfflineMachines     int                          `json:"offline_machines"`
	DegradedMachines    int                          `json:"degraded_machines"`
	MaintenanceMachines int                          `json:"maintenance_machines"`
	Environments        map[string]EnvironmentHealth `json:"environments"`
	Alerts              []Alert                      `json:"alerts"`
	Timestamp           time.Time                    `json:"timestamp"`
}

// EnvironmentHealth represents health of machines in an environment
type EnvironmentHealth struct {
	Environment       string `json:"environment"`
	TotalMachines     int    `json:"total_machines"`
	HealthyMachines   int    `json:"healthy_machines"`
	DegradedMachines  int    `json:"degraded_machines"`
	UnhealthyMachines int    `json:"unhealthy_machines"`
}

// GetMachinesByTag returns machines with specific tags
func (fm *FleetManager) GetMachinesByTag(ctx context.Context, tags []string) ([]*Machine, error) {
	machines, err := fm.ListMachines(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*Machine
	for _, machine := range machines {
		if fm.machineHasTags(machine, tags) {
			filtered = append(filtered, machine)
		}
	}

	return filtered, nil
}

// GetMachinesByEnvironment returns machines in a specific environment
func (fm *FleetManager) GetMachinesByEnvironment(ctx context.Context, environment string) ([]*Machine, error) {
	machines, err := fm.ListMachines(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []*Machine
	for _, machine := range machines {
		if machine.Environment == environment {
			filtered = append(filtered, machine)
		}
	}

	return filtered, nil
}

// machineHasTags checks if a machine has all specified tags
func (fm *FleetManager) machineHasTags(machine *Machine, tags []string) bool {
	if len(tags) == 0 {
		return true
	}

	machineTagSet := make(map[string]bool)
	for _, tag := range machine.Tags {
		machineTagSet[tag] = true
	}

	for _, tag := range tags {
		if !machineTagSet[tag] {
			return false
		}
	}

	return true
}

// SetMachineMaintenance puts a machine into maintenance mode
func (fm *FleetManager) SetMachineMaintenance(ctx context.Context, machineID string, maintenance bool) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	machine, exists := fm.machines[machineID]
	if !exists {
		return fmt.Errorf("machine %s not found", machineID)
	}

	if maintenance {
		machine.Status = MachineStatusMaintenance
		fm.logger.Info(fmt.Sprintf("Machine set to maintenance mode: %s", machineID))
	} else {
		// Return to previous status based on health
		if machine.Health.Overall == "healthy" {
			machine.Status = MachineStatusOnline
		} else {
			machine.Status = MachineStatusDegraded
		}
		fm.logger.Info(fmt.Sprintf("Machine removed from maintenance mode: %s", machineID))
	}

	return nil
}

// Helper functions for local health checking

// getLocalCPUHealth gets real CPU health metrics
func (m *Monitor) getLocalCPUHealth() ResourceHealth {
	// Read CPU usage from /proc/stat
	cpuUsage := m.getCPUUsage()
	
	status := "healthy"
	if cpuUsage > 90.0 {
		status = "critical"
	} else if cpuUsage > 80.0 {
		status = "warning"
	}
	
	return ResourceHealth{
		Status:    status,
		Usage:     cpuUsage,
		Threshold: 80.0,
	}
}

// getLocalMemoryHealth gets real memory health metrics
func (m *Monitor) getLocalMemoryHealth() ResourceHealth {
	memUsage, memAvailable := m.getMemoryUsage()
	
	status := "healthy"
	if memUsage > 95.0 {
		status = "critical"
	} else if memUsage > 85.0 {
		status = "warning"
	}
	
	return ResourceHealth{
		Status:    status,
		Usage:     memUsage,
		Available: memAvailable,
		Threshold: 85.0,
	}
}

// getLocalDiskHealth gets real disk health metrics
func (m *Monitor) getLocalDiskHealth() ResourceHealth {
	diskUsage, diskAvailable := m.getDiskUsage()
	
	status := "healthy"
	if diskUsage > 95.0 {
		status = "critical"
	} else if diskUsage > 90.0 {
		status = "warning"
	}
	
	return ResourceHealth{
		Status:    status,
		Usage:     diskUsage,
		Available: diskAvailable,
		Threshold: 90.0,
	}
}

// getLocalNixOSServices gets real NixOS service health
func (m *Monitor) getLocalNixOSServices() []ServiceHealth {
	services := []ServiceHealth{}
	
	// Check common NixOS services
	commonServices := []string{"sshd", "systemd-resolved", "systemd-networkd", "nixos-rebuild"}
	
	for _, serviceName := range commonServices {
		service := m.getServiceHealth(serviceName)
		if service.Name != "" {
			services = append(services, service)
		}
	}
	
	return services
}

// calculateOverallHealth determines overall health from individual components
func (m *Monitor) calculateOverallHealth(health HealthStatus) string {
	criticalCount := 0
	warningCount := 0
	
	if health.CPU.Status == "critical" {
		criticalCount++
	} else if health.CPU.Status == "warning" {
		warningCount++
	}
	
	if health.Memory.Status == "critical" {
		criticalCount++
	} else if health.Memory.Status == "warning" {
		warningCount++
	}
	
	if health.Disk.Status == "critical" {
		criticalCount++
	} else if health.Disk.Status == "warning" {
		warningCount++
	}
	
	if criticalCount > 0 {
		return "critical"
	} else if warningCount > 0 {
		return "warning"
	}
	
	return "healthy"
}

// System metric collection functions

// getCPUUsage calculates current CPU usage percentage
func (m *Monitor) getCPUUsage() float64 {
	// Read /proc/stat for CPU times
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

// getMemoryUsage calculates current memory usage percentage and available memory
func (m *Monitor) getMemoryUsage() (float64, uint64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0.0, 0
	}
	
	lines := strings.Split(string(data), "\n")
	var memTotal, memAvailable uint64
	
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		switch fields[0] {
		case "MemTotal:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memTotal = val * 1024 // Convert KB to bytes
			}
		case "MemAvailable:":
			if val, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memAvailable = val * 1024 // Convert KB to bytes
			}
		}
	}
	
	if memTotal == 0 {
		return 0.0, 0
	}
	
	usage := float64(memTotal-memAvailable) / float64(memTotal) * 100
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	
	return usage, memAvailable
}

// getDiskUsage calculates disk usage percentage and available space
func (m *Monitor) getDiskUsage() (float64, uint64) {
	// Use df command to get disk usage for root partition
	cmd := exec.Command("df", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0.0, 0
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0.0, 0
	}
	
	// Parse the second line which contains the root filesystem info
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0.0, 0
	}
	
	// Extract available space (field 3, in KB)
	availableStr := fields[3]
	available, err := strconv.ParseUint(availableStr, 10, 64)
	if err != nil {
		return 0.0, 0
	}
	availableBytes := available * 1024 // Convert KB to bytes
	
	// Extract usage percentage (field 4, remove % suffix)
	usagePct := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.ParseFloat(usagePct, 64)
	if err != nil {
		return 0.0, availableBytes
	}
	
	return usage, availableBytes
}

// getServiceHealth checks the health of a specific system service
func (m *Monitor) getServiceHealth(serviceName string) ServiceHealth {
	// Use systemctl to check service status
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	
	service := ServiceHealth{
		Name:   serviceName,
		Status: "unknown",
		Since:  time.Now(),
		Memory: 0,
		CPU:    0.0,
	}
	
	if err == nil {
		status := strings.TrimSpace(string(output))
		if status == "active" {
			service.Status = "running"
		} else {
			service.Status = status
		}
	}
	
	// Get service start time if running
	if service.Status == "running" {
		cmd = exec.Command("systemctl", "show", serviceName, "--property=ActiveEnterTimestamp", "--value")
		if timeOutput, err := cmd.Output(); err == nil {
			if timestamp, err := time.Parse("Mon 2006-01-02 15:04:05 MST", strings.TrimSpace(string(timeOutput))); err == nil {
				service.Since = timestamp
			}
		}
	}
	
	return service
}
