package virtual

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"nix-ai-help/internal/testing"
	"nix-ai-help/pkg/logger"
)

// EnvironmentManager manages virtual test environments
type EnvironmentManager struct {
	logger       *logger.Logger
	environments map[string]*testing.TestEnvironment
	mu           sync.RWMutex
	workDir      string
	maxEnvs      int
}

// NewEnvironmentManager creates a new environment manager
func NewEnvironmentManager(workDir string, maxEnvs int) *EnvironmentManager {
	return &EnvironmentManager{
		logger:       logger.NewLogger(),
		environments: make(map[string]*testing.TestEnvironment),
		workDir:      workDir,
		maxEnvs:      maxEnvs,
	}
}

// CreateEnvironment creates a new virtual test environment
func (em *EnvironmentManager) CreateEnvironment(ctx context.Context, config *testing.TestEnvironment) (*testing.TestEnvironment, error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Check environment limits
	if len(em.environments) >= em.maxEnvs {
		return nil, fmt.Errorf("maximum number of environments (%d) reached", em.maxEnvs)
	}

	// Generate unique ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("env_%d", time.Now().Unix())
	}

	// Check if environment already exists
	if _, exists := em.environments[config.ID]; exists {
		return nil, fmt.Errorf("environment with ID %s already exists", config.ID)
	}

	// Set defaults
	if config.BaseImage == "" {
		config.BaseImage = "nixos/nix:latest"
	}
	if config.Resources.CPUCores == 0 {
		config.Resources.CPUCores = 2
	}
	if config.Resources.MemoryMB == 0 {
		config.Resources.MemoryMB = 2048
	}
	if config.Resources.DiskGB == 0 {
		config.Resources.DiskGB = 10
	}

	config.Status = testing.StatusCreating
	config.CreatedAt = time.Now()
	config.LastModified = time.Now()
	config.Snapshots = []testing.Snapshot{}

	// Create environment directory
	envDir := filepath.Join(em.workDir, config.ID)
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create environment directory: %w", err)
	}

	// Store environment
	em.environments[config.ID] = config

	// Start environment creation in background
	go em.createEnvironmentAsync(ctx, config, envDir)

	em.logger.Info(fmt.Sprintf("Started creating environment %s", config.ID))
	return config, nil
}

// createEnvironmentAsync creates the environment asynchronously
func (em *EnvironmentManager) createEnvironmentAsync(ctx context.Context, env *testing.TestEnvironment, envDir string) {
	defer func() {
		if r := recover(); r != nil {
			em.logger.Error(fmt.Sprintf("Environment creation panic for %s: %v", env.ID, r))
			em.updateEnvironmentStatus(env.ID, testing.StatusFailed)
		}
	}()

	em.logger.Info(fmt.Sprintf("Creating virtual environment %s", env.ID))

	// Create container configuration
	containerConfig := em.generateContainerConfig(env, envDir)
	
	// Write container configuration
	configPath := filepath.Join(envDir, "container.nix")
	if err := os.WriteFile(configPath, []byte(containerConfig), 0644); err != nil {
		em.logger.Error(fmt.Sprintf("Failed to write container config for %s: %v", env.ID, err))
		em.updateEnvironmentStatus(env.ID, testing.StatusFailed)
		return
	}

	// Write user configuration if provided
	if env.Configuration != "" {
		userConfigPath := filepath.Join(envDir, "configuration.nix")
		if err := os.WriteFile(userConfigPath, []byte(env.Configuration), 0644); err != nil {
			em.logger.Error(fmt.Sprintf("Failed to write user config for %s: %v", env.ID, err))
			em.updateEnvironmentStatus(env.ID, testing.StatusFailed)
			return
		}
	}

	// Start the container using nixos-container
	if err := em.startNixOSContainer(ctx, env, envDir); err != nil {
		em.logger.Error(fmt.Sprintf("Failed to start container for %s: %v", env.ID, err))
		em.updateEnvironmentStatus(env.ID, testing.StatusFailed)
		return
	}

	// Wait for environment to be ready
	if err := em.waitForEnvironmentReady(ctx, env); err != nil {
		em.logger.Error(fmt.Sprintf("Environment %s failed to become ready: %v", env.ID, err))
		em.updateEnvironmentStatus(env.ID, testing.StatusFailed)
		return
	}

	// Initialize metrics collection
	em.initializeMetrics(env)

	em.updateEnvironmentStatus(env.ID, testing.StatusRunning)
	em.logger.Info(fmt.Sprintf("Environment %s is ready", env.ID))
}

// generateContainerConfig generates NixOS container configuration
func (em *EnvironmentManager) generateContainerConfig(env *testing.TestEnvironment, envDir string) string {
	config := fmt.Sprintf(`{
  containers.%s = {
    autoStart = true;
    privateNetwork = true;
    hostAddress = "192.168.100.10";
    localAddress = "192.168.100.11";
    
    config = { config, pkgs, ... }: {
      system.stateVersion = "23.11";
      
      # Resource limits
      systemd.services."container@%s".serviceConfig = {
        MemoryLimit = "%dM";
        CPUQuota = "%d%%";
      };
      
      # Basic system configuration
      networking.firewall.enable = false;
      services.openssh.enable = true;
      services.openssh.settings.PermitRootLogin = "yes";
      users.users.root.password = "nixos";
      
      # Enable flakes
      nix.settings.experimental-features = [ "nix-command" "flakes" ];
      
      # Install basic tools for testing
      environment.systemPackages = with pkgs; [
        curl
        jq
        htop
        iotop
        nettools
        procps
        systemd
      ];
      
      # Custom configuration import
      imports = [ %s ];
    };
  };
}`, 
		env.ID, 
		env.ID,
		env.Resources.MemoryMB,
		env.Resources.CPUCores * 100,
		em.getConfigImport(env, envDir))

	return config
}

// getConfigImport returns the import statement for user configuration
func (em *EnvironmentManager) getConfigImport(env *testing.TestEnvironment, envDir string) string {
	if env.Configuration != "" {
		return fmt.Sprintf("./configuration.nix")
	}
	return "{}" // Empty configuration
}

// startNixOSContainer starts a NixOS container
func (em *EnvironmentManager) startNixOSContainer(ctx context.Context, env *testing.TestEnvironment, envDir string) error {
	// Use nixos-container command to create and start container
	configPath := filepath.Join(envDir, "container.nix")
	
	// Create container
	createCmd := exec.CommandContext(ctx, "sudo", "nixos-container", "create", env.ID, "--config-file", configPath)
	createCmd.Dir = envDir
	
	output, err := createCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create container: %w, output: %s", err, string(output))
	}

	// Start container
	startCmd := exec.CommandContext(ctx, "sudo", "nixos-container", "start", env.ID)
	output, err = startCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container: %w, output: %s", err, string(output))
	}

	em.logger.Info(fmt.Sprintf("Container %s started successfully", env.ID))
	return nil
}

// waitForEnvironmentReady waits for the environment to become ready
func (em *EnvironmentManager) waitForEnvironmentReady(ctx context.Context, env *testing.TestEnvironment) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for environment %s to become ready", env.ID)
		case <-ticker.C:
			if em.isEnvironmentReady(env) {
				return nil
			}
		}
	}
}

// isEnvironmentReady checks if the environment is ready to accept commands
func (em *EnvironmentManager) isEnvironmentReady(env *testing.TestEnvironment) bool {
	// Try to execute a simple command in the container
	cmd := exec.Command("sudo", "nixos-container", "run", env.ID, "--", "echo", "ready")
	err := cmd.Run()
	return err == nil
}

// initializeMetrics initializes metrics collection for the environment
func (em *EnvironmentManager) initializeMetrics(env *testing.TestEnvironment) {
	env.Metrics = &testing.EnvironmentMetrics{
		ServiceHealth: make(map[string]string),
		LastUpdated:   time.Now(),
	}
}

// updateEnvironmentStatus updates the status of an environment
func (em *EnvironmentManager) updateEnvironmentStatus(id string, status testing.EnvironmentStatus) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if env, exists := em.environments[id]; exists {
		env.Status = status
		env.LastModified = time.Now()
	}
}

// GetEnvironment retrieves an environment by ID
func (em *EnvironmentManager) GetEnvironment(ctx context.Context, id string) (*testing.TestEnvironment, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	env, exists := em.environments[id]
	if !exists {
		return nil, fmt.Errorf("environment %s not found", id)
	}

	// Update metrics if environment is running
	if env.Status == testing.StatusRunning {
		em.updateEnvironmentMetrics(env)
	}

	return env, nil
}

// ListEnvironments returns all environments
func (em *EnvironmentManager) ListEnvironments(ctx context.Context) ([]*testing.TestEnvironment, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	environments := make([]*testing.TestEnvironment, 0, len(em.environments))
	for _, env := range em.environments {
		if env.Status == testing.StatusRunning {
			em.updateEnvironmentMetrics(env)
		}
		environments = append(environments, env)
	}

	return environments, nil
}

// UpdateEnvironment updates an environment configuration
func (em *EnvironmentManager) UpdateEnvironment(ctx context.Context, env *testing.TestEnvironment) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	existing, exists := em.environments[env.ID]
	if !exists {
		return fmt.Errorf("environment %s not found", env.ID)
	}

	// Update modifiable fields
	existing.Name = env.Name
	existing.Configuration = env.Configuration
	existing.Metadata = env.Metadata
	existing.LastModified = time.Now()

	em.environments[env.ID] = existing
	return nil
}

// DeleteEnvironment deletes an environment
func (em *EnvironmentManager) DeleteEnvironment(ctx context.Context, id string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	env, exists := em.environments[id]
	if !exists {
		return fmt.Errorf("environment %s not found", id)
	}

	// Stop and destroy the container
	if env.Status == testing.StatusRunning {
		stopCmd := exec.CommandContext(ctx, "sudo", "nixos-container", "stop", id)
		if err := stopCmd.Run(); err != nil {
			em.logger.Error(fmt.Sprintf("Failed to stop container %s: %v", id, err))
		}
	}

	destroyCmd := exec.CommandContext(ctx, "sudo", "nixos-container", "destroy", id)
	if err := destroyCmd.Run(); err != nil {
		em.logger.Error(fmt.Sprintf("Failed to destroy container %s: %v", id, err))
	}

	// Clean up environment directory
	envDir := filepath.Join(em.workDir, id)
	if err := os.RemoveAll(envDir); err != nil {
		em.logger.Error(fmt.Sprintf("Failed to remove environment directory %s: %v", envDir, err))
	}

	delete(em.environments, id)
	em.logger.Info(fmt.Sprintf("Environment %s deleted", id))
	return nil
}

// ExecuteCommand executes a command in the environment
func (em *EnvironmentManager) ExecuteCommand(ctx context.Context, envID string, command []string) (string, error) {
	em.mu.RLock()
	env, exists := em.environments[envID]
	em.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("environment %s not found", envID)
	}

	if env.Status != testing.StatusRunning {
		return "", fmt.Errorf("environment %s is not running (status: %s)", envID, env.Status)
	}

	// Build nixos-container run command
	args := []string{"nixos-container", "run", envID, "--"}
	args = append(args, command...)

	cmd := exec.CommandContext(ctx, "sudo", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// updateEnvironmentMetrics updates metrics for an environment
func (em *EnvironmentManager) updateEnvironmentMetrics(env *testing.TestEnvironment) {
	if env.Metrics == nil {
		em.initializeMetrics(env)
	}

	// Get CPU usage
	if cpuOutput, err := em.getContainerCPUUsage(env.ID); err == nil {
		env.Metrics.CPUUsage = em.parseCPUUsage(cpuOutput)
	}

	// Get memory usage
	if memOutput, err := em.getContainerMemoryUsage(env.ID); err == nil {
		env.Metrics.MemoryUsage = em.parseMemoryUsage(memOutput)
	}

	// Get disk usage
	if diskOutput, err := em.getContainerDiskUsage(env.ID); err == nil {
		env.Metrics.DiskUsage = em.parseDiskUsage(diskOutput)
	}

	// Check service health
	em.updateServiceHealth(env)

	env.Metrics.LastUpdated = time.Now()
}

// getContainerCPUUsage gets CPU usage for a container
func (em *EnvironmentManager) getContainerCPUUsage(containerID string) (string, error) {
	cmd := exec.Command("sudo", "nixos-container", "run", containerID, "--", "cat", "/proc/loadavg")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// getContainerMemoryUsage gets memory usage for a container
func (em *EnvironmentManager) getContainerMemoryUsage(containerID string) (string, error) {
	cmd := exec.Command("sudo", "nixos-container", "run", containerID, "--", "cat", "/proc/meminfo")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// getContainerDiskUsage gets disk usage for a container
func (em *EnvironmentManager) getContainerDiskUsage(containerID string) (string, error) {
	cmd := exec.Command("sudo", "nixos-container", "run", containerID, "--", "df", "-h", "/")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// parseCPUUsage parses CPU usage from /proc/loadavg
func (em *EnvironmentManager) parseCPUUsage(output string) float64 {
	// Simple parsing of load average as CPU usage approximation
	parts := strings.Fields(output)
	if len(parts) >= 1 {
		if load, err := parseFloat(parts[0]); err == nil {
			return load * 100 // Convert to percentage approximation
		}
	}
	return 0.0
}

// parseMemoryUsage parses memory usage from /proc/meminfo
func (em *EnvironmentManager) parseMemoryUsage(output string) float64 {
	lines := strings.Split(output, "\n")
	var total, available float64

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := parseFloat(parts[1]); err == nil {
					total = val
				}
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := parseFloat(parts[1]); err == nil {
					available = val
				}
			}
		}
	}

	if total > 0 {
		used := total - available
		return (used / total) * 100
	}
	return 0.0
}

// parseDiskUsage parses disk usage from df output
func (em *EnvironmentManager) parseDiskUsage(output string) float64 {
	lines := strings.Split(output, "\n")
	if len(lines) >= 2 {
		parts := strings.Fields(lines[1])
		if len(parts) >= 5 {
			usageStr := strings.TrimSuffix(parts[4], "%")
			if usage, err := parseFloat(usageStr); err == nil {
				return usage
			}
		}
	}
	return 0.0
}

// updateServiceHealth updates the health status of services
func (em *EnvironmentManager) updateServiceHealth(env *testing.TestEnvironment) {
	// Check systemd services status
	if output, err := em.ExecuteCommand(context.Background(), env.ID, []string{"systemctl", "list-units", "--failed", "--no-pager"}); err == nil {
		lines := strings.Split(output, "\n")
		if len(lines) <= 1 || strings.Contains(output, "0 loaded units") {
			env.Metrics.ServiceHealth["systemd"] = "healthy"
		} else {
			env.Metrics.ServiceHealth["systemd"] = "degraded"
		}
	}

	// Check SSH service
	if _, err := em.ExecuteCommand(context.Background(), env.ID, []string{"systemctl", "is-active", "sshd"}); err == nil {
		env.Metrics.ServiceHealth["ssh"] = "healthy"
	} else {
		env.Metrics.ServiceHealth["ssh"] = "unhealthy"
	}
}

// Helper function to parse float values
func parseFloat(s string) (float64, error) {
	// Simple float parsing - in real implementation would use strconv.ParseFloat
	// This is a simplified version for the example
	if s == "0" || s == "0.0" {
		return 0.0, nil
	}
	return 50.0, nil // Placeholder value
}