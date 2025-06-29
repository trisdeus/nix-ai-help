package plugins

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// Sandbox provides secure execution environment for plugins
type Sandbox struct {
	config         *config.UserConfig
	logger         *logger.Logger
	enforcing      bool
	resourceLimits map[string]ResourceLimits
}

// NewSandbox creates a new plugin sandbox
func NewSandbox(cfg *config.UserConfig, log *logger.Logger) *Sandbox {
	return &Sandbox{
		config:         cfg,
		logger:         log,
		enforcing:      true, // Enable by default for security
		resourceLimits: make(map[string]ResourceLimits),
	}
}

// Execute runs a function within the sandbox environment
func (s *Sandbox) Execute(ctx context.Context, plugin PluginInterface, fn func() error) error {
	pluginName := plugin.Name()

	// Get security policy for plugin
	// In a real implementation, this would come from plugin configuration
	policy := SecurityPolicy{
		AllowFileSystem:  true,  // For now, allow filesystem access
		AllowNetwork:     true,  // For now, allow network access
		AllowSystemCalls: false, // Restrict system calls
		SandboxLevel:     SandboxBasic,
	}

	return s.executeWithPolicy(ctx, pluginName, policy, fn)
}

// executeWithPolicy executes a function with specific security policy
func (s *Sandbox) executeWithPolicy(ctx context.Context, pluginName string, policy SecurityPolicy, fn func() error) error {
	if !s.enforcing {
		// If sandbox is disabled, execute directly
		return fn()
	}

	s.logger.Debug(fmt.Sprintf("Executing plugin '%s' in sandbox with policy: %s", pluginName, policy.SandboxLevel))

	// Create execution context with timeout
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Monitor execution
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error(fmt.Sprintf("Plugin '%s' panicked: %v", pluginName, r))
				done <- fmt.Errorf("plugin panicked: %v", r)
			}
		}()

		// Apply resource limits before execution
		if err := s.applyResourceLimits(pluginName); err != nil {
			done <- fmt.Errorf("failed to apply resource limits: %w", err)
			return
		}

		// Execute the function
		err := fn()
		done <- err
	}()

	// Wait for completion or context cancellation
	select {
	case err := <-done:
		return err
	case <-execCtx.Done():
		return fmt.Errorf("plugin execution cancelled: %w", execCtx.Err())
	}
}

// applyResourceLimits applies resource constraints to the current process
func (s *Sandbox) applyResourceLimits(pluginName string) error {
	limits, exists := s.resourceLimits[pluginName]
	if !exists {
		// Use default limits
		limits = ResourceLimits{
			MaxMemoryMB:      100, // 100MB default
			MaxCPUPercent:    50,  // 50% CPU default
			MaxExecutionTime: 30 * time.Second,
			MaxFileSize:      10 * 1024 * 1024, // 10MB
			NetworkAccess:    true,
		}
	}

	// Apply memory limit (on Linux, this would use cgroups)
	if limits.MaxMemoryMB > 0 {
		if err := s.setMemoryLimit(limits.MaxMemoryMB); err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to set memory limit for plugin '%s': %v", pluginName, err))
		}
	}

	// Apply file size limit
	if limits.MaxFileSize > 0 {
		if err := s.setFileLimit(limits.MaxFileSize); err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to set file size limit for plugin '%s': %v", pluginName, err))
		}
	}

	return nil
}

// setMemoryLimit sets memory limit for the process
func (s *Sandbox) setMemoryLimit(limitMB int) error {
	// This is a simplified implementation
	// In production, you'd use cgroups on Linux or job objects on Windows

	if runtime.GOOS == "linux" {
		limit := uint64(limitMB * 1024 * 1024)

		// Set virtual memory limit
		rLimit := syscall.Rlimit{
			Cur: limit,
			Max: limit,
		}

		if err := syscall.Setrlimit(syscall.RLIMIT_AS, &rLimit); err != nil {
			return fmt.Errorf("failed to set virtual memory limit: %w", err)
		}
	}

	return nil
}

// setFileLimit sets file size limit for the process
func (s *Sandbox) setFileLimit(limitBytes int64) error {
	if runtime.GOOS == "linux" {
		rLimit := syscall.Rlimit{
			Cur: uint64(limitBytes),
			Max: uint64(limitBytes),
		}

		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &rLimit); err != nil {
			return fmt.Errorf("failed to set file size limit: %w", err)
		}
	}

	return nil
}

// SetResourceLimits sets resource limits for a specific plugin
func (s *Sandbox) SetResourceLimits(pluginName string, limits ResourceLimits) {
	s.resourceLimits[pluginName] = limits
	s.logger.Debug(fmt.Sprintf("Set resource limits for plugin '%s'", pluginName))
}

// GetResourceLimits gets resource limits for a specific plugin
func (s *Sandbox) GetResourceLimits(pluginName string) (ResourceLimits, bool) {
	limits, exists := s.resourceLimits[pluginName]
	return limits, exists
}

// SetEnforcing enables or disables sandbox enforcement
func (s *Sandbox) SetEnforcing(enforcing bool) {
	s.enforcing = enforcing
	if enforcing {
		s.logger.Info("Sandbox enforcement enabled")
	} else {
		s.logger.Warn("Sandbox enforcement disabled - plugins will run without restrictions")
	}
}

// IsEnforcing returns whether sandbox enforcement is enabled
func (s *Sandbox) IsEnforcing() bool {
	return s.enforcing
}

// ValidatePolicy validates a security policy
func (s *Sandbox) ValidatePolicy(policy SecurityPolicy) error {
	// Check sandbox level is valid
	if policy.SandboxLevel < SandboxNone || policy.SandboxLevel > SandboxIsolated {
		return fmt.Errorf("invalid sandbox level: %d", policy.SandboxLevel)
	}

	// Validate allowed domains if network access is restricted
	if !policy.AllowNetwork && len(policy.AllowedDomains) > 0 {
		return fmt.Errorf("cannot specify allowed domains when network access is disabled")
	}

	return nil
}

// PermissionChecker provides methods to check plugin permissions
type PermissionChecker struct {
	policies map[string]SecurityPolicy
	logger   *logger.Logger
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(log *logger.Logger) *PermissionChecker {
	return &PermissionChecker{
		policies: make(map[string]SecurityPolicy),
		logger:   log,
	}
}

// SetPolicy sets security policy for a plugin
func (pc *PermissionChecker) SetPolicy(pluginName string, policy SecurityPolicy) {
	pc.policies[pluginName] = policy
}

// CheckFileAccess checks if a plugin can access a file
func (pc *PermissionChecker) CheckFileAccess(pluginName, filePath string) bool {
	policy, exists := pc.policies[pluginName]
	if !exists {
		return false // Deny by default
	}

	if !policy.AllowFileSystem {
		return false
	}

	// Check if path is in allowed paths
	if len(policy.AllowedDomains) > 0 {
		for _, allowedPath := range policy.AllowedDomains {
			if filePath == allowedPath || isSubPath(filePath, allowedPath) {
				return true
			}
		}
		return false
	}

	return true
}

// CheckNetworkAccess checks if a plugin can access the network
func (pc *PermissionChecker) CheckNetworkAccess(pluginName, domain string) bool {
	policy, exists := pc.policies[pluginName]
	if !exists {
		return false // Deny by default
	}

	if !policy.AllowNetwork {
		return false
	}

	// Check if domain is in allowed domains
	if len(policy.AllowedDomains) > 0 {
		for _, allowedDomain := range policy.AllowedDomains {
			if domain == allowedDomain {
				return true
			}
		}
		return false
	}

	return true
}

// CheckSystemCall checks if a plugin can make system calls
func (pc *PermissionChecker) CheckSystemCall(pluginName string) bool {
	policy, exists := pc.policies[pluginName]
	if !exists {
		return false // Deny by default
	}

	return policy.AllowSystemCalls
}

// isSubPath checks if path is a subpath of basePath
func isSubPath(path, basePath string) bool {
	// Simple implementation - in production you'd use filepath.Rel and check for ".."
	return len(path) > len(basePath) && path[:len(basePath)] == basePath && path[len(basePath)] == os.PathSeparator
}

// ResourceMonitor monitors resource usage of plugins
type ResourceMonitor struct {
	logger *logger.Logger
	stats  map[string]*ResourceStats
}

// ResourceStats tracks resource usage statistics
type ResourceStats struct {
	MemoryUsage        int64     `json:"memory_usage"`
	CPUUsage           float64   `json:"cpu_usage"`
	FileDescriptors    int       `json:"file_descriptors"`
	NetworkConnections int       `json:"network_connections"`
	LastUpdated        time.Time `json:"last_updated"`
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(log *logger.Logger) *ResourceMonitor {
	return &ResourceMonitor{
		logger: log,
		stats:  make(map[string]*ResourceStats),
	}
}

// StartMonitoring starts monitoring a plugin's resource usage
func (rm *ResourceMonitor) StartMonitoring(pluginName string) {
	rm.stats[pluginName] = &ResourceStats{
		LastUpdated: time.Now(),
	}
	rm.logger.Debug(fmt.Sprintf("Started monitoring plugin '%s'", pluginName))
}

// StopMonitoring stops monitoring a plugin's resource usage
func (rm *ResourceMonitor) StopMonitoring(pluginName string) {
	delete(rm.stats, pluginName)
	rm.logger.Debug(fmt.Sprintf("Stopped monitoring plugin '%s'", pluginName))
}

// GetStats returns resource usage statistics for a plugin
func (rm *ResourceMonitor) GetStats(pluginName string) (*ResourceStats, bool) {
	stats, exists := rm.stats[pluginName]
	if !exists {
		return nil, false
	}

	// In a real implementation, this would gather actual resource usage
	// For now, return the stored stats
	return stats, true
}

// UpdateStats updates resource usage statistics for a plugin
func (rm *ResourceMonitor) UpdateStats(pluginName string, stats *ResourceStats) {
	stats.LastUpdated = time.Now()
	rm.stats[pluginName] = stats
}

// GetAllStats returns resource usage statistics for all monitored plugins
func (rm *ResourceMonitor) GetAllStats() map[string]*ResourceStats {
	// Return a copy to prevent modification
	result := make(map[string]*ResourceStats)
	for name, stats := range rm.stats {
		statsCopy := *stats
		result[name] = &statsCopy
	}
	return result
}
