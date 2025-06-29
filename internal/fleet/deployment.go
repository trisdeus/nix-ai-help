package fleet

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// CreateDeployment creates a new fleet deployment
func (fm *FleetManager) CreateDeployment(ctx context.Context, req DeploymentRequest) (*Deployment, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Validate deployment request
	if err := fm.validateDeploymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid deployment request: %w", err)
	}

	// Create deployment
	deployment := &Deployment{
		ID:         uuid.New().String(),
		Name:       req.Name,
		ConfigHash: req.ConfigHash,
		Targets:    req.Targets,
		Status:     DeploymentStatusPending,
		Strategy:   req.Strategy,
		CreatedAt:  time.Now(),
		CreatedBy:  req.CreatedBy,
		Results:    make(map[string]DeploymentResult),
	}

	// Initialize progress
	deployment.Progress = DeploymentProgress{
		Total:      len(req.Targets),
		Completed:  0,
		Failed:     0,
		InProgress: 0,
		Percentage: 0,
	}

	// Set up rollback if enabled
	if req.RollbackEnabled {
		deployment.Rollback = &RollbackInfo{
			Enabled: true,
			Trigger: "manual",
		}
	}

	fm.deployments[deployment.ID] = deployment
	fm.logger.Info(fmt.Sprintf("Deployment created: %s (name: %s, targets: %d)", deployment.ID, deployment.Name, len(deployment.Targets)))

	return deployment, nil
}

// StartDeployment starts executing a deployment
func (fm *FleetManager) StartDeployment(ctx context.Context, deploymentID string) error {
	fm.mu.Lock()
	deployment, exists := fm.deployments[deploymentID]
	if !exists {
		fm.mu.Unlock()
		return fmt.Errorf("deployment %s not found", deploymentID)
	}

	if deployment.Status != DeploymentStatusPending {
		fm.mu.Unlock()
		return fmt.Errorf("deployment %s is not pending (current status: %s)", deploymentID, deployment.Status)
	}

	deployment.Status = DeploymentStatusRunning
	now := time.Now()
	deployment.StartedAt = &now
	fm.mu.Unlock()

	// Execute deployment asynchronously
	go fm.executeDeployment(ctx, deployment)

	fm.logger.Info(fmt.Sprintf("Deployment started: %s", deploymentID))
	return nil
}

// executeDeployment executes the deployment strategy
func (fm *FleetManager) executeDeployment(ctx context.Context, deployment *Deployment) {
	switch deployment.Strategy.Type {
	case "rolling":
		fm.executeRollingDeployment(ctx, deployment)
	case "blue_green":
		fm.executeBlueGreenDeployment(ctx, deployment)
	case "canary":
		fm.executeCanaryDeployment(ctx, deployment)
	default:
		fm.executeParallelDeployment(ctx, deployment)
	}
}

// executeRollingDeployment executes a rolling deployment
func (fm *FleetManager) executeRollingDeployment(ctx context.Context, deployment *Deployment) {
	fm.logger.Info(fmt.Sprintf("Starting rolling deployment: %s", deployment.ID))

	batchSize := deployment.Strategy.BatchSize
	if batchSize <= 0 {
		batchSize = 1
	}

	targets := deployment.Targets
	failureCount := 0
	failureThreshold := int(math.Ceil(float64(len(targets)) * deployment.Strategy.FailureThreshold))

	for i := 0; i < len(targets); i += batchSize {
		// Check if deployment was cancelled or failed
		if deployment.Status == DeploymentStatusCancelled || deployment.Status == DeploymentStatusFailed {
			break
		}

		end := i + batchSize
		if end > len(targets) {
			end = len(targets)
		}

		batch := targets[i:end]
		fm.logger.Debug(fmt.Sprintf("Deploying batch %d for deployment %s (%d machines)", i/batchSize+1, deployment.ID, len(batch)))

		// Deploy to batch
		batchResults := fm.deployToBatch(ctx, deployment, batch)

		// Update deployment results
		fm.mu.Lock()
		for machineID, result := range batchResults {
			deployment.Results[machineID] = result
			if result.Status == "failed" {
				failureCount++
				deployment.Progress.Failed++
			} else if result.Status == "success" {
				deployment.Progress.Completed++
			}
		}

		deployment.Progress.InProgress = 0
		deployment.Progress.Percentage = int(float64(deployment.Progress.Completed+deployment.Progress.Failed) / float64(deployment.Progress.Total) * 100)
		fm.mu.Unlock()

		// Check failure threshold
		if failureCount >= failureThreshold {
			fm.logger.Error(fmt.Sprintf("Deployment %s failed due to failure threshold: %d failures, threshold: %d", deployment.ID, failureCount, failureThreshold))
			fm.failDeployment(deployment, "failure threshold exceeded")
			return
		}

		// Wait between batches (except for the last batch)
		if end < len(targets) && deployment.Strategy.BatchDelay > 0 {
			time.Sleep(time.Duration(deployment.Strategy.BatchDelay) * time.Second)
		}
	}

	// Complete deployment
	fm.completeDeployment(deployment)
}

// executeBlueGreenDeployment executes a blue-green deployment
func (fm *FleetManager) executeBlueGreenDeployment(ctx context.Context, deployment *Deployment) {
	fm.logger.Info(fmt.Sprintf("Starting blue-green deployment: %s", deployment.ID))

	// In blue-green deployment, we deploy to all machines simultaneously
	// but with a switch-over mechanism
	results := fm.deployToBatch(ctx, deployment, deployment.Targets)

	fm.mu.Lock()
	defer fm.mu.Unlock()

	failureCount := 0
	for machineID, result := range results {
		deployment.Results[machineID] = result
		if result.Status == "failed" {
			failureCount++
			deployment.Progress.Failed++
		} else {
			deployment.Progress.Completed++
		}
	}

	deployment.Progress.Percentage = 100

	// Check if deployment should fail
	failureThreshold := int(math.Ceil(float64(len(deployment.Targets)) * deployment.Strategy.FailureThreshold))
	if failureCount >= failureThreshold {
		fm.failDeployment(deployment, "failure threshold exceeded in blue-green deployment")
		return
	}

	fm.completeDeployment(deployment)
}

// executeCanaryDeployment executes a canary deployment
func (fm *FleetManager) executeCanaryDeployment(ctx context.Context, deployment *Deployment) {
	fm.logger.Info(fmt.Sprintf("Starting canary deployment: %s", deployment.ID))

	// Canary deployment: deploy to a small subset first, then the rest
	canarySize := int(math.Max(1, float64(len(deployment.Targets))*0.1)) // 10% canary
	canaryTargets := deployment.Targets[:canarySize]
	mainTargets := deployment.Targets[canarySize:]

	// Deploy to canary targets first
	fm.logger.Debug(fmt.Sprintf("Deploying to canary targets for deployment %s (canary size: %d)", deployment.ID, len(canaryTargets)))
	canaryResults := fm.deployToBatch(ctx, deployment, canaryTargets)

	// Check canary results
	canaryFailures := 0
	for _, result := range canaryResults {
		if result.Status == "failed" {
			canaryFailures++
		}
	}

	if canaryFailures > 0 {
		fm.logger.Error(fmt.Sprintf("Canary deployment %s failed with %d failures", deployment.ID, canaryFailures))
		fm.failDeployment(deployment, "canary deployment failed")
		return
	}

	// Wait for canary validation
	if deployment.Strategy.HealthCheck.Enabled {
		time.Sleep(time.Duration(deployment.Strategy.HealthCheck.Timeout) * time.Second)
		// TODO: Implement health check validation
	}

	// Deploy to remaining targets
	fm.logger.Debug(fmt.Sprintf("Deploying to main targets for deployment %s (main size: %d)", deployment.ID, len(mainTargets)))
	mainResults := fm.deployToBatch(ctx, deployment, mainTargets)

	// Update results
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for machineID, result := range canaryResults {
		deployment.Results[machineID] = result
		if result.Status == "failed" {
			deployment.Progress.Failed++
		} else {
			deployment.Progress.Completed++
		}
	}

	for machineID, result := range mainResults {
		deployment.Results[machineID] = result
		if result.Status == "failed" {
			deployment.Progress.Failed++
		} else {
			deployment.Progress.Completed++
		}
	}

	deployment.Progress.Percentage = 100
	fm.completeDeployment(deployment)
}

// executeParallelDeployment executes a parallel deployment to all targets
func (fm *FleetManager) executeParallelDeployment(ctx context.Context, deployment *Deployment) {
	fm.logger.Info(fmt.Sprintf("Starting parallel deployment: %s", deployment.ID))

	results := fm.deployToBatch(ctx, deployment, deployment.Targets)

	fm.mu.Lock()
	defer fm.mu.Unlock()

	for machineID, result := range results {
		deployment.Results[machineID] = result
		if result.Status == "failed" {
			deployment.Progress.Failed++
		} else {
			deployment.Progress.Completed++
		}
	}

	deployment.Progress.Percentage = 100
	fm.completeDeployment(deployment)
}

// deployToBatch deploys configuration to a batch of machines
func (fm *FleetManager) deployToBatch(ctx context.Context, deployment *Deployment, targets []string) map[string]DeploymentResult {
	results := make(map[string]DeploymentResult)

	// Deploy to each machine in the batch
	for _, machineID := range targets {
		result := fm.deployToMachine(ctx, deployment, machineID)
		results[machineID] = result
	}

	return results
}

// deployToMachine deploys configuration to a single machine
func (fm *FleetManager) deployToMachine(ctx context.Context, deployment *Deployment, machineID string) DeploymentResult {
	startTime := time.Now()

	fm.logger.Debug(fmt.Sprintf("Deploying to machine %s for deployment %s", machineID, deployment.ID))

	// Get machine
	machine, exists := fm.machines[machineID]
	if !exists {
		return DeploymentResult{
			MachineID:   machineID,
			Status:      "failed",
			StartedAt:   startTime,
			CompletedAt: &startTime,
			Error:       "machine not found",
		}
	}

	// Check machine status
	if machine.Status != MachineStatusOnline {
		return DeploymentResult{
			MachineID:   machineID,
			Status:      "failed",
			StartedAt:   startTime,
			CompletedAt: &startTime,
			Error:       fmt.Sprintf("machine is %s", machine.Status),
		}
	}

	// Update progress
	fm.mu.Lock()
	deployment.Progress.InProgress++
	fm.mu.Unlock()

	// Simulate deployment (in real implementation, this would:
	// 1. Copy configuration to machine
	// 2. Run nixos-rebuild switch
	// 3. Verify deployment success
	// 4. Update machine status)
	time.Sleep(time.Duration(2+len(machineID)%5) * time.Second) // Simulate deployment time

	completedAt := time.Now()

	// Update machine config status
	machine.Config.TargetHash = deployment.ConfigHash
	machine.Config.CurrentHash = deployment.ConfigHash
	machine.Config.LastUpdate = completedAt
	machine.Config.UpdateStatus = "up-to-date"
	machine.Config.Generation++

	fm.mu.Lock()
	deployment.Progress.InProgress--
	fm.mu.Unlock()

	return DeploymentResult{
		MachineID:      machineID,
		Status:         "success",
		StartedAt:      startTime,
		CompletedAt:    &completedAt,
		Generation:     machine.Config.Generation,
		PreviousHash:   machine.Config.CurrentHash,
		NewHash:        deployment.ConfigHash,
		RebootRequired: false,
		Output:         "Configuration deployed successfully",
	}
}

// completeDeployment marks a deployment as completed
func (fm *FleetManager) completeDeployment(deployment *Deployment) {
	now := time.Now()
	deployment.Status = DeploymentStatusCompleted
	deployment.CompletedAt = &now

	fm.logger.Info(fmt.Sprintf("Deployment completed - ID: %s, Total: %d, Completed: %d, Failed: %d",
		deployment.ID, deployment.Progress.Total, deployment.Progress.Completed, deployment.Progress.Failed))
}

// failDeployment marks a deployment as failed
func (fm *FleetManager) failDeployment(deployment *Deployment, reason string) {
	now := time.Now()
	deployment.Status = DeploymentStatusFailed
	deployment.CompletedAt = &now

	fm.logger.Error(fmt.Sprintf("Deployment %s failed: %s", deployment.ID, reason))
}

// DeploymentRequest represents a request to create a deployment
type DeploymentRequest struct {
	Name            string             `json:"name"`
	ConfigHash      string             `json:"config_hash"`
	Targets         []string           `json:"targets"`
	Strategy        DeploymentStrategy `json:"strategy"`
	CreatedBy       string             `json:"created_by"`
	RollbackEnabled bool               `json:"rollback_enabled"`
}

// validateDeploymentRequest validates a deployment request
func (fm *FleetManager) validateDeploymentRequest(req DeploymentRequest) error {
	if req.Name == "" {
		return fmt.Errorf("deployment name is required")
	}
	if req.ConfigHash == "" {
		return fmt.Errorf("configuration hash is required")
	}
	if len(req.Targets) == 0 {
		return fmt.Errorf("at least one target machine is required")
	}
	if req.CreatedBy == "" {
		return fmt.Errorf("creator is required")
	}

	// Validate target machines exist
	for _, target := range req.Targets {
		if _, exists := fm.machines[target]; !exists {
			return fmt.Errorf("target machine %s not found", target)
		}
	}

	// Validate strategy
	if req.Strategy.Type == "" {
		req.Strategy.Type = "rolling" // default
	}
	if req.Strategy.BatchSize <= 0 {
		req.Strategy.BatchSize = 1
	}
	if req.Strategy.FailureThreshold <= 0 {
		req.Strategy.FailureThreshold = 0.1 // 10% default
	}

	return nil
}

// GetDeployment returns a specific deployment
func (fm *FleetManager) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	deployment, exists := fm.deployments[deploymentID]
	if !exists {
		return nil, fmt.Errorf("deployment %s not found", deploymentID)
	}

	return deployment, nil
}

// ListDeployments returns all deployments
func (fm *FleetManager) ListDeployments(ctx context.Context) ([]*Deployment, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	deployments := make([]*Deployment, 0, len(fm.deployments))
	for _, deployment := range fm.deployments {
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// CancelDeployment cancels a running deployment
func (fm *FleetManager) CancelDeployment(ctx context.Context, deploymentID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	deployment, exists := fm.deployments[deploymentID]
	if !exists {
		return fmt.Errorf("deployment %s not found", deploymentID)
	}

	if deployment.Status != DeploymentStatusRunning {
		return fmt.Errorf("deployment %s is not running (current status: %s)", deploymentID, deployment.Status)
	}

	deployment.Status = DeploymentStatusCancelled
	now := time.Now()
	deployment.CompletedAt = &now

	fm.logger.Info(fmt.Sprintf("Deployment cancelled: %s", deploymentID))
	return nil
}
