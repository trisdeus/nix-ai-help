package fleet

import (
	"context"
	"fmt"
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
	// In a real implementation, this would:
	// 1. SSH to the machine
	// 2. Check system resources (CPU, memory, disk)
	// 3. Check system services
	// 4. Check network connectivity
	// 5. Check NixOS-specific health

	// For now, we'll simulate health check results
	health := HealthStatus{
		Overall:   "healthy",
		LastCheck: time.Now(),
		CPU: ResourceHealth{
			Status:    "healthy",
			Usage:     float64(machine.ID[0]%50 + 10), // Simulate 10-60% CPU usage
			Threshold: 80.0,
		},
		Memory: ResourceHealth{
			Status:    "healthy",
			Usage:     float64(machine.ID[0]%40 + 20), // Simulate 20-60% memory usage
			Available: 8 * 1024 * 1024 * 1024,         // 8GB available
			Threshold: 85.0,
		},
		Disk: ResourceHealth{
			Status:    "healthy",
			Usage:     float64(machine.ID[0]%30 + 30), // Simulate 30-60% disk usage
			Available: 100 * 1024 * 1024 * 1024,       // 100GB available
			Threshold: 90.0,
		},
		Network: ResourceHealth{
			Status: "healthy",
			Usage:  float64(machine.ID[0]%20 + 5), // Simulate 5-25% network usage
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
