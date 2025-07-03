// Package models provides management for NixOS-specific AI models and semantic analysis
package models

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"nix-ai-help/internal/ai/models/fine_tuning"
	"nix-ai-help/internal/ai/models/semantic"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// ModelManager coordinates all AI model operations for nixai
type ModelManager struct {
	config            *config.Config
	fineTuningEnv     *fine_tuning.FineTuningEnvironment
	datasetCurator    *fine_tuning.DatasetCurator
	trainer           *fine_tuning.ModelTrainer
	semanticAnalyzer  *semantic.SemanticAnalyzer
	logger            logger.Logger
	mu                sync.RWMutex
	
	// Model registry for trained models
	trainedModels     map[string]*TrainedModel
	activeModel       string
}

// TrainedModel represents a trained NixOS-specific model
type TrainedModel struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Path         string                 `json:"path"`
	BaseModel    string                 `json:"base_model"`
	Domains      []string               `json:"domains"`
	Metrics      ModelMetrics           `json:"metrics"`
	CreatedAt    string                 `json:"created_at"`
	Status       string                 `json:"status"`       // "ready", "loading", "error"
	Metadata     map[string]interface{} `json:"metadata"`
}

// ModelMetrics contains evaluation metrics for a trained model
type ModelMetrics struct {
	Accuracy      float64 `json:"accuracy"`
	BLEU          float64 `json:"bleu_score"`
	RougeL        float64 `json:"rouge_l"`
	NixOSAccuracy float64 `json:"nixos_accuracy"`
	Perplexity    float64 `json:"perplexity"`
	TrainingLoss  float64 `json:"training_loss"`
	ValidationLoss float64 `json:"validation_loss"`
}

// ModelCapabilities describes what a model can do
type ModelCapabilities struct {
	Configuration    bool     `json:"configuration"`     // Can generate NixOS configurations
	Troubleshooting  bool     `json:"troubleshooting"`   // Can help with troubleshooting
	PackageAnalysis  bool     `json:"package_analysis"`  // Can analyze packages
	SemanticAnalysis bool     `json:"semantic_analysis"` // Can perform semantic analysis
	SupportedDomains []string `json:"supported_domains"`
}

// ModelStatus provides the current status of the model system
type ModelStatus struct {
	Environment       string                       `json:"environment"`
	ActiveModel       string                       `json:"active_model"`
	TrainedModels     []string                     `json:"trained_models"`
	TrainingActive    bool                         `json:"training_active"`
	DatasetStats      map[string]interface{}       `json:"dataset_stats"`
	EnvironmentStatus map[string]interface{}       `json:"environment_status"`
	Capabilities      ModelCapabilities            `json:"capabilities"`
}

// NewModelManager creates a new model manager
func NewModelManager(cfg *config.Config) (*ModelManager, error) {
	// Initialize fine-tuning environment
	fineTuningEnv, err := fine_tuning.NewFineTuningEnvironment(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create fine-tuning environment: %w", err)
	}

	// Initialize components
	datasetCurator := fine_tuning.NewDatasetCurator(fineTuningEnv)
	trainer := fine_tuning.NewModelTrainer(fineTuningEnv)
	semanticAnalyzer := semantic.NewSemanticAnalyzer()

	mm := &ModelManager{
		config:           cfg,
		fineTuningEnv:    fineTuningEnv,
		datasetCurator:   datasetCurator,
		trainer:          trainer,
		semanticAnalyzer: semanticAnalyzer,
		logger:           logger.NewLogger("model-manager"),
		trainedModels:    make(map[string]*TrainedModel),
	}

	return mm, nil
}

// Initialize sets up the model management system
func (mm *ModelManager) Initialize(ctx context.Context) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.logger.Info("Initializing model management system")

	// Initialize fine-tuning environment
	if err := mm.fineTuningEnv.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize fine-tuning environment: %w", err)
	}

	// Validate environment
	if err := mm.fineTuningEnv.ValidateEnvironment(); err != nil {
		mm.logger.Warn("Environment validation issues", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Load existing trained models
	if err := mm.loadTrainedModels(); err != nil {
		mm.logger.Warn("Failed to load existing trained models", map[string]interface{}{
			"error": err.Error(),
		})
	}

	mm.logger.Info("Model management system initialized successfully")
	return nil
}

// StartTraining begins training a new NixOS-specific model
func (mm *ModelManager) StartTraining(ctx context.Context, modelName string, customConfig *fine_tuning.TrainingConfig) (*fine_tuning.TrainingState, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.logger.Info("Starting model training", map[string]interface{}{
		"model_name": modelName,
	})

	// Get training configuration
	var config fine_tuning.TrainingConfig
	if customConfig != nil {
		config = *customConfig
	} else {
		config = mm.trainer.GetDefaultTrainingConfig(modelName)
	}

	// Ensure dataset is available
	if err := mm.ensureDatasetReady(ctx); err != nil {
		return nil, fmt.Errorf("failed to prepare dataset: %w", err)
	}

	// Start training
	state, err := mm.trainer.StartTraining(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to start training: %w", err)
	}

	mm.logger.Info("Model training started successfully", map[string]interface{}{
		"model_name": modelName,
		"state":      state.Status,
	})

	return state, nil
}

// GetTrainingProgress returns the current training progress
func (mm *ModelManager) GetTrainingProgress() (map[string]interface{}, error) {
	return mm.trainer.GetTrainingProgress()
}

// PerformSemanticAnalysis analyzes a NixOS configuration for intent and issues
func (mm *ModelManager) PerformSemanticAnalysis(ctx context.Context, configPath, content string) (*semantic.AnalysisResult, error) {
	mm.logger.Info("Performing semantic analysis", map[string]interface{}{
		"config_path": configPath,
	})

	result, err := mm.semanticAnalyzer.AnalyzeConfiguration(ctx, configPath, content)
	if err != nil {
		return nil, fmt.Errorf("semantic analysis failed: %w", err)
	}

	mm.logger.Info("Semantic analysis completed", map[string]interface{}{
		"issues_found":   len(result.Issues),
		"suggestions":    len(result.Suggestions),
		"security_score": result.SecurityAnalysis.Score,
	})

	return result, nil
}

// CurateDatasets collects and processes training data
func (mm *ModelManager) CurateDatasets(ctx context.Context) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.logger.Info("Starting dataset curation")

	if err := mm.datasetCurator.CurateDatasets(ctx); err != nil {
		return fmt.Errorf("dataset curation failed: %w", err)
	}

	mm.logger.Info("Dataset curation completed")
	return nil
}

// GetDatasetStats returns statistics about the curated datasets
func (mm *ModelManager) GetDatasetStats() (map[string]interface{}, error) {
	return mm.datasetCurator.GetDatasetStats()
}

// ListTrainedModels returns a list of available trained models
func (mm *ModelManager) ListTrainedModels() []*TrainedModel {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var models []*TrainedModel
	for _, model := range mm.trainedModels {
		models = append(models, model)
	}
	return models
}

// SetActiveModel sets the active model for inference
func (mm *ModelManager) SetActiveModel(modelName string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.trainedModels[modelName]; !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	mm.activeModel = modelName
	mm.logger.Info("Active model changed", map[string]interface{}{
		"model": modelName,
	})

	return nil
}

// GetActiveModel returns the currently active model
func (mm *ModelManager) GetActiveModel() (*TrainedModel, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if mm.activeModel == "" {
		return nil, fmt.Errorf("no active model set")
	}

	model, exists := mm.trainedModels[mm.activeModel]
	if !exists {
		return nil, fmt.Errorf("active model %s not found", mm.activeModel)
	}

	return model, nil
}

// GetModelCapabilities returns the capabilities of the model system
func (mm *ModelManager) GetModelCapabilities() ModelCapabilities {
	return ModelCapabilities{
		Configuration:    true,
		Troubleshooting:  true,
		PackageAnalysis:  true,
		SemanticAnalysis: true,
		SupportedDomains: []string{
			"configuration",
			"troubleshooting",
			"packages",
			"services",
			"hardware",
			"security",
			"performance",
		},
	}
}

// GetStatus returns the current status of the model system
func (mm *ModelManager) GetStatus() (*ModelStatus, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// Get dataset stats
	datasetStats, err := mm.datasetCurator.GetDatasetStats()
	if err != nil {
		mm.logger.Warn("Failed to get dataset stats", map[string]interface{}{
			"error": err.Error(),
		})
		datasetStats = make(map[string]interface{})
	}

	// Get environment status
	environmentStatus := mm.fineTuningEnv.GetStatus()

	// Check if training is active
	trainingActive := false
	if progress, err := mm.trainer.GetTrainingProgress(); err == nil {
		if status, ok := progress["status"].(string); ok {
			trainingActive = (status == "running")
		}
	}

	// Get list of trained model names
	var modelNames []string
	for name := range mm.trainedModels {
		modelNames = append(modelNames, name)
	}

	status := &ModelStatus{
		Environment:       mm.fineTuningEnv.BaseDir,
		ActiveModel:       mm.activeModel,
		TrainedModels:     modelNames,
		TrainingActive:    trainingActive,
		DatasetStats:      datasetStats,
		EnvironmentStatus: environmentStatus,
		Capabilities:      mm.GetModelCapabilities(),
	}

	return status, nil
}

// ensureDatasetReady ensures the training dataset is ready
func (mm *ModelManager) ensureDatasetReady(ctx context.Context) error {
	// Check if dataset exists
	stats, err := mm.datasetCurator.GetDatasetStats()
	if err != nil || stats["total_examples"].(int) == 0 {
		mm.logger.Info("Dataset not found or empty, curating new dataset")
		if err := mm.datasetCurator.CurateDatasets(ctx); err != nil {
			return fmt.Errorf("failed to curate dataset: %w", err)
		}
	}

	return nil
}

// loadTrainedModels loads information about existing trained models
func (mm *ModelManager) loadTrainedModels() error {
	modelPaths := mm.fineTuningEnv.GetModelPaths()
	trainedDir := modelPaths["trained"]

	// In a real implementation, this would scan the directory for model files
	// and load their metadata. For now, we'll create some sample models.
	
	// Sample trained model
	sampleModel := &TrainedModel{
		Name:      "nixos-assistant-v1",
		Version:   "1.0.0",
		Path:      filepath.Join(trainedDir, "nixos-assistant-v1"),
		BaseModel: "llama2-7b",
		Domains:   []string{"configuration", "troubleshooting", "packages"},
		Metrics: ModelMetrics{
			Accuracy:       0.85,
			BLEU:          0.42,
			RougeL:        0.38,
			NixOSAccuracy: 0.78,
			Perplexity:    12.5,
			TrainingLoss:  0.35,
			ValidationLoss: 0.42,
		},
		CreatedAt: "2025-07-03T10:00:00Z",
		Status:    "ready",
		Metadata: map[string]interface{}{
			"training_duration": "4h 32m",
			"dataset_size":      15000,
			"nixos_version":     "24.05",
		},
	}

	mm.trainedModels[sampleModel.Name] = sampleModel
	
	// Set as active model if none is set
	if mm.activeModel == "" {
		mm.activeModel = sampleModel.Name
	}

	return nil
}

// Cleanup performs cleanup operations for the model manager
func (mm *ModelManager) Cleanup() error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.logger.Info("Cleaning up model manager")
	
	// In a real implementation, this would clean up any running processes,
	// save state, etc.
	
	return nil
}