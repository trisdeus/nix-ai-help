package fleet

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/pkg/logger"
)

// FleetManager manages multiple NixOS machines and deployments
type FleetManager struct {
	machines    map[string]*Machine
	deployments map[string]*Deployment
	logger      *logger.Logger
	mu          sync.RWMutex
}

// NewFleetManager creates a new fleet manager
func NewFleetManager(logger *logger.Logger) *FleetManager {
	return &FleetManager{
		machines:    make(map[string]*Machine),
		deployments: make(map[string]*Deployment),
		logger:      logger,
	}
}

// Machine represents a managed NixOS machine
type Machine struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Address     string            `json:"address"`
	SSHConfig   SSHConfig         `json:"ssh_config"`
	Status      MachineStatus     `json:"status"`
	Tags        []string          `json:"tags"`
	Environment string            `json:"environment"` // production, staging, development
	Metadata    map[string]string `json:"metadata"`
	LastSeen    time.Time         `json:"last_seen"`
	Health      HealthStatus      `json:"health"`
	Config      ConfigStatus      `json:"config"`
}

// SSHConfig contains SSH connection configuration
type SSHConfig struct {
	User      string `json:"user"`
	Port      int    `json:"port"`
	KeyPath   string `json:"key_path"`
	ProxyJump string `json:"proxy_jump,omitempty"`
	Timeout   int    `json:"timeout"`
}

// MachineStatus represents the current status of a machine
type MachineStatus string

const (
	MachineStatusOnline      MachineStatus = "online"
	MachineStatusOffline     MachineStatus = "offline"
	MachineStatusUnknown     MachineStatus = "unknown"
	MachineStatusDegraded    MachineStatus = "degraded"
	MachineStatusMaintenance MachineStatus = "maintenance"
)

// HealthStatus represents the health of a machine
type HealthStatus struct {
	Overall   string          `json:"overall"` // healthy, warning, critical
	CPU       ResourceHealth  `json:"cpu"`
	Memory    ResourceHealth  `json:"memory"`
	Disk      ResourceHealth  `json:"disk"`
	Network   ResourceHealth  `json:"network"`
	Services  []ServiceHealth `json:"services"`
	LastCheck time.Time       `json:"last_check"`
	Alerts    []Alert         `json:"alerts"`
}

// ResourceHealth represents health of a system resource
type ResourceHealth struct {
	Status    string  `json:"status"`    // healthy, warning, critical
	Usage     float64 `json:"usage"`     // percentage
	Available uint64  `json:"available"` // bytes for disk/memory
	Threshold float64 `json:"threshold"` // warning threshold
}

// ServiceHealth represents health of a system service
type ServiceHealth struct {
	Name   string    `json:"name"`
	Status string    `json:"status"` // running, stopped, failed
	Since  time.Time `json:"since"`
	Memory uint64    `json:"memory"`
	CPU    float64   `json:"cpu"`
}

// Alert represents a system alert
type Alert struct {
	ID         string     `json:"id"`
	Level      string     `json:"level"` // info, warning, critical
	Message    string     `json:"message"`
	Source     string     `json:"source"`
	Timestamp  time.Time  `json:"timestamp"`
	Resolved   bool       `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// ConfigStatus represents the configuration status of a machine
type ConfigStatus struct {
	CurrentHash   string    `json:"current_hash"`
	TargetHash    string    `json:"target_hash"`
	LastUpdate    time.Time `json:"last_update"`
	UpdateStatus  string    `json:"update_status"` // up-to-date, updating, failed
	PendingReboot bool      `json:"pending_reboot"`
	Generation    int       `json:"generation"`
}

// Deployment represents a fleet-wide deployment
type Deployment struct {
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	ConfigHash  string                      `json:"config_hash"`
	Targets     []string                    `json:"targets"` // machine IDs
	Status      DeploymentStatus            `json:"status"`
	Progress    DeploymentProgress          `json:"progress"`
	Strategy    DeploymentStrategy          `json:"strategy"`
	CreatedAt   time.Time                   `json:"created_at"`
	StartedAt   *time.Time                  `json:"started_at,omitempty"`
	CompletedAt *time.Time                  `json:"completed_at,omitempty"`
	CreatedBy   string                      `json:"created_by"`
	Rollback    *RollbackInfo               `json:"rollback,omitempty"`
	Results     map[string]DeploymentResult `json:"results"`
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending     DeploymentStatus = "pending"
	DeploymentStatusRunning     DeploymentStatus = "running"
	DeploymentStatusCompleted   DeploymentStatus = "completed"
	DeploymentStatusFailed      DeploymentStatus = "failed"
	DeploymentStatusCancelled   DeploymentStatus = "cancelled"
	DeploymentStatusRollingBack DeploymentStatus = "rolling_back"
)

// DeploymentProgress tracks deployment progress
type DeploymentProgress struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	InProgress int `json:"in_progress"`
	Percentage int `json:"percentage"`
}

// DeploymentStrategy defines how deployments are executed
type DeploymentStrategy struct {
	Type             string      `json:"type"`              // rolling, blue_green, canary
	BatchSize        int         `json:"batch_size"`        // machines per batch
	BatchDelay       int         `json:"batch_delay"`       // seconds between batches
	FailureThreshold float64     `json:"failure_threshold"` // percentage of failures to abort
	HealthCheck      HealthCheck `json:"health_check"`
}

// HealthCheck defines post-deployment health verification
type HealthCheck struct {
	Enabled    bool   `json:"enabled"`
	Timeout    int    `json:"timeout"` // seconds
	RetryCount int    `json:"retry_count"`
	Endpoint   string `json:"endpoint,omitempty"` // HTTP endpoint to check
	Command    string `json:"command,omitempty"`  // command to run for verification
}

// RollbackInfo contains rollback information
type RollbackInfo struct {
	Enabled      bool      `json:"enabled"`
	PreviousHash string    `json:"previous_hash"`
	Trigger      string    `json:"trigger"` // manual, auto, failure
	InitiatedAt  time.Time `json:"initiated_at"`
	InitiatedBy  string    `json:"initiated_by"`
}

// DeploymentResult represents the result of deployment on a single machine
type DeploymentResult struct {
	MachineID      string     `json:"machine_id"`
	Status         string     `json:"status"` // success, failed, skipped
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Error          string     `json:"error,omitempty"`
	Generation     int        `json:"generation"`
	PreviousHash   string     `json:"previous_hash"`
	NewHash        string     `json:"new_hash"`
	RebootRequired bool       `json:"reboot_required"`
	Output         string     `json:"output"`
}

// AddMachine adds a new machine to the fleet
func (fm *FleetManager) AddMachine(ctx context.Context, machine *Machine) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.machines[machine.ID]; exists {
		return fmt.Errorf("machine %s already exists", machine.ID)
	}

	// Validate machine configuration
	if err := fm.validateMachine(machine); err != nil {
		return fmt.Errorf("invalid machine configuration: %w", err)
	}

	// Test connectivity
	if err := fm.testMachineConnectivity(ctx, machine); err != nil {
		fm.logger.Warn(fmt.Sprintf("Machine connectivity test failed for %s: %v", machine.ID, err))
		machine.Status = MachineStatusOffline
	} else {
		machine.Status = MachineStatusOnline
		machine.LastSeen = time.Now()
	}

	fm.machines[machine.ID] = machine
	fm.logger.Info(fmt.Sprintf("Machine added to fleet: %s (%s)", machine.ID, machine.Name))

	return nil
}

// RemoveMachine removes a machine from the fleet
func (fm *FleetManager) RemoveMachine(ctx context.Context, machineID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	machine, exists := fm.machines[machineID]
	if !exists {
		return fmt.Errorf("machine %s not found", machineID)
	}

	// Check if machine is part of any active deployments
	for _, deployment := range fm.deployments {
		if deployment.Status == DeploymentStatusRunning {
			for _, target := range deployment.Targets {
				if target == machineID {
					return fmt.Errorf("cannot remove machine %s: part of active deployment %s", machineID, deployment.ID)
				}
			}
		}
	}

	delete(fm.machines, machineID)
	fm.logger.Info(fmt.Sprintf("Machine removed from fleet: %s (%s)", machineID, machine.Name))

	return nil
}

// ListMachines returns all machines in the fleet
func (fm *FleetManager) ListMachines(ctx context.Context) ([]*Machine, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	machines := make([]*Machine, 0, len(fm.machines))
	for _, machine := range fm.machines {
		machines = append(machines, machine)
	}

	return machines, nil
}

// GetMachine returns a specific machine
func (fm *FleetManager) GetMachine(ctx context.Context, machineID string) (*Machine, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	machine, exists := fm.machines[machineID]
	if !exists {
		return nil, fmt.Errorf("machine %s not found", machineID)
	}

	return machine, nil
}

// UpdateMachineHealth updates the health status of a machine
func (fm *FleetManager) UpdateMachineHealth(ctx context.Context, machineID string, health HealthStatus) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	machine, exists := fm.machines[machineID]
	if !exists {
		return fmt.Errorf("machine %s not found", machineID)
	}

	machine.Health = health
	machine.LastSeen = time.Now()

	// Update machine status based on health
	switch health.Overall {
	case "healthy":
		machine.Status = MachineStatusOnline
	case "warning":
		machine.Status = MachineStatusDegraded
	case "critical":
		machine.Status = MachineStatusDegraded
	default:
		machine.Status = MachineStatusUnknown
	}

	return nil
}

// AddRepositoryMachine adds a machine discovered from a repository with relaxed validation
// This allows machines without addresses (DHCP machines) to be tracked
func (fm *FleetManager) AddRepositoryMachine(ctx context.Context, machine *Machine) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Use relaxed validation for repository machines
	if err := fm.validateRepositoryMachine(machine); err != nil {
		return fmt.Errorf("invalid machine configuration: %w", err)
	}

	// Check if machine already exists
	if existingMachine, exists := fm.machines[machine.ID]; exists {
		fm.logger.Warn(fmt.Sprintf("Machine %s already exists, updating from repository", machine.ID))
		// Update existing machine with repository data
		existingMachine.Name = machine.Name
		if machine.Address != "" {
			existingMachine.Address = machine.Address
		}
		existingMachine.Tags = machine.Tags
		existingMachine.Environment = machine.Environment
		existingMachine.Metadata = machine.Metadata
		return nil
	}

	// Set defaults for repository machines
	if machine.Status == "" {
		machine.Status = MachineStatusUnknown
	}
	if machine.SSHConfig.Port <= 0 {
		machine.SSHConfig.Port = 22
	}
	if machine.SSHConfig.Timeout <= 0 {
		machine.SSHConfig.Timeout = 30
	}
	if machine.SSHConfig.User == "" {
		machine.SSHConfig.User = "nixos" // default NixOS user
	}

	// Add metadata to indicate this is a repository-discovered machine
	if machine.Metadata == nil {
		machine.Metadata = make(map[string]string)
	}
	machine.Metadata["source"] = "repository"
	machine.Metadata["discovered_at"] = time.Now().Format(time.RFC3339)

	fm.machines[machine.ID] = machine
	fm.logger.Info(fmt.Sprintf("Added repository machine: %s (%s)", machine.ID, machine.Name))

	return nil
}

// validateMachine validates machine configuration
func (fm *FleetManager) validateMachine(machine *Machine) error {
	if machine.ID == "" {
		return fmt.Errorf("machine ID is required")
	}
	if machine.Name == "" {
		return fmt.Errorf("machine name is required")
	}
	if machine.Address == "" {
		return fmt.Errorf("machine address is required")
	}
	if machine.SSHConfig.User == "" {
		return fmt.Errorf("SSH user is required")
	}
	if machine.SSHConfig.Port <= 0 {
		machine.SSHConfig.Port = 22 // default SSH port
	}
	if machine.SSHConfig.Timeout <= 0 {
		machine.SSHConfig.Timeout = 30 // default timeout
	}

	return nil
}

// validateRepositoryMachine validates machine configuration with relaxed rules for repository machines
func (fm *FleetManager) validateRepositoryMachine(machine *Machine) error {
	if machine.ID == "" {
		return fmt.Errorf("machine ID is required")
	}
	if machine.Name == "" {
		return fmt.Errorf("machine name is required")
	}
	// Note: Address is not required for repository machines (DHCP machines)
	// Note: SSH config is not required for repository machines (may not be accessible)

	return nil
}

// testMachineConnectivity tests SSH connectivity to a machine
func (fm *FleetManager) testMachineConnectivity(ctx context.Context, machine *Machine) error { // This would implement actual SSH connectivity testing
	// For now, we'll simulate the test
	fm.logger.Debug(fmt.Sprintf("Testing machine connectivity: %s at %s", machine.ID, machine.Address))

	// TODO: Implement actual SSH connectivity test
	// - Create SSH client with machine.SSHConfig
	// - Attempt to connect and run a simple command
	// - Return error if connection fails

	return nil
}
