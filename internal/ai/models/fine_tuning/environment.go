// Package fine_tuning provides infrastructure for training NixOS-specific AI models
package fine_tuning

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// FineTuningEnvironment manages the setup and configuration for AI model fine-tuning
type FineTuningEnvironment struct {
	BaseDir     string
	DatasetDir  string
	ModelDir    string
	OutputDir   string
	ConfigPath  string
	Logger      *logger.Logger
	config      *config.UserConfig
}

// EnvironmentConfig holds configuration for the fine-tuning environment
type EnvironmentConfig struct {
	BaseDirectory    string `yaml:"base_directory"`
	DatasetDirectory string `yaml:"dataset_directory"`
	ModelDirectory   string `yaml:"model_directory"`
	OutputDirectory  string `yaml:"output_directory"`
	ConfigFile       string `yaml:"config_file"`
	
	// Model training parameters
	ModelType        string  `yaml:"model_type"`        // "llama2", "mistral", "custom"
	BatchSize        int     `yaml:"batch_size"`
	LearningRate     float64 `yaml:"learning_rate"`
	Epochs           int     `yaml:"epochs"`
	MaxSequenceLen   int     `yaml:"max_sequence_length"`
	
	// NixOS-specific configuration
	NixOSVersion     string   `yaml:"nixos_version"`
	TargetDomains    []string `yaml:"target_domains"`    // ["configuration", "packages", "services", "troubleshooting"]
	DataSources      []string `yaml:"data_sources"`      // ["nixos-manual", "nixpkgs", "community-configs"]
	
	// Training optimization
	UseGPU           bool   `yaml:"use_gpu"`
	GPUMemoryLimit   string `yaml:"gpu_memory_limit"`
	ParallelWorkers  int    `yaml:"parallel_workers"`
	CheckpointFreq   int    `yaml:"checkpoint_frequency"`
}

// NewFineTuningEnvironment creates a new fine-tuning environment
func NewFineTuningEnvironment(config *config.UserConfig) (*FineTuningEnvironment, error) {
	// Get base directory from config or use default
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	baseDir := filepath.Join(homeDir, ".nixai", "fine-tuning")

	env := &FineTuningEnvironment{
		BaseDir:    baseDir,
		DatasetDir: filepath.Join(baseDir, "datasets"),
		ModelDir:   filepath.Join(baseDir, "models"),
		OutputDir:  filepath.Join(baseDir, "output"),
		ConfigPath: filepath.Join(baseDir, "config.yaml"),
		Logger:     logger.NewLogger(),
		config:     config,
	}

	return env, nil
}

// Initialize sets up the fine-tuning environment directories and configuration
func (env *FineTuningEnvironment) Initialize(ctx context.Context) error {
	env.Logger.Info(fmt.Sprintf("Initializing fine-tuning environment at %s", env.BaseDir))

	// Create necessary directories
	dirs := []string{
		env.BaseDir,
		env.DatasetDir,
		env.ModelDir,
		env.OutputDir,
		filepath.Join(env.DatasetDir, "nixos-configs"),
		filepath.Join(env.DatasetDir, "successful-patterns"),
		filepath.Join(env.DatasetDir, "community-solutions"),
		filepath.Join(env.DatasetDir, "troubleshooting-cases"),
		filepath.Join(env.ModelDir, "checkpoints"),
		filepath.Join(env.ModelDir, "trained"),
		filepath.Join(env.OutputDir, "experiments"),
		filepath.Join(env.OutputDir, "evaluations"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create default configuration if it doesn't exist
	if err := env.createDefaultConfig(); err != nil {
		return fmt.Errorf("failed to create default config: %w", err)
	}

	env.Logger.Info("Fine-tuning environment initialized successfully")
	return nil
}

// createDefaultConfig creates a default fine-tuning configuration
func (env *FineTuningEnvironment) createDefaultConfig() error {
	if _, err := os.Stat(env.ConfigPath); err == nil {
		// Config already exists
		return nil
	}

	defaultConfig := &EnvironmentConfig{
		BaseDirectory:    env.BaseDir,
		DatasetDirectory: env.DatasetDir,
		ModelDirectory:   env.ModelDir,
		OutputDirectory:  env.OutputDir,
		ConfigFile:       env.ConfigPath,
		
		// Model configuration
		ModelType:       "llama2",
		BatchSize:       8,
		LearningRate:    2e-4,
		Epochs:          3,
		MaxSequenceLen:  2048,
		
		// NixOS-specific
		NixOSVersion: "24.05",
		TargetDomains: []string{
			"configuration",
			"packages",
			"services",
			"troubleshooting",
			"hardware",
			"security",
		},
		DataSources: []string{
			"nixos-manual",
			"nixpkgs-repository",
			"community-configs",
			"successful-deployments",
			"troubleshooting-cases",
		},
		
		// Training optimization
		UseGPU:          true,
		GPUMemoryLimit:  "8GB",
		ParallelWorkers: 4,
		CheckpointFreq:  100,
	}

	// Write configuration to file
	return env.writeConfig(defaultConfig)
}

// writeConfig writes the environment configuration to file
func (env *FineTuningEnvironment) writeConfig(envConfig *EnvironmentConfig) error {
	// This would normally use a YAML library, but for now we'll create a basic structure
	configContent := fmt.Sprintf(`# NixOS AI Model Fine-Tuning Configuration
base_directory: %s
dataset_directory: %s
model_directory: %s
output_directory: %s
config_file: %s

# Model Training Parameters
model_type: %s
batch_size: %d
learning_rate: %f
epochs: %d
max_sequence_length: %d

# NixOS-specific Configuration
nixos_version: %s
target_domains:
  - configuration
  - packages
  - services
  - troubleshooting
  - hardware
  - security

data_sources:
  - nixos-manual
  - nixpkgs-repository
  - community-configs
  - successful-deployments
  - troubleshooting-cases

# Training Optimization
use_gpu: %t
gpu_memory_limit: %s
parallel_workers: %d
checkpoint_frequency: %d
`,
		envConfig.BaseDirectory,
		envConfig.DatasetDirectory,
		envConfig.ModelDirectory,
		envConfig.OutputDirectory,
		envConfig.ConfigFile,
		envConfig.ModelType,
		envConfig.BatchSize,
		envConfig.LearningRate,
		envConfig.Epochs,
		envConfig.MaxSequenceLen,
		envConfig.NixOSVersion,
		envConfig.UseGPU,
		envConfig.GPUMemoryLimit,
		envConfig.ParallelWorkers,
		envConfig.CheckpointFreq,
	)

	return os.WriteFile(env.ConfigPath, []byte(configContent), 0644)
}

// GetDatasetPaths returns paths for different types of training datasets
func (env *FineTuningEnvironment) GetDatasetPaths() map[string]string {
	return map[string]string{
		"nixos_configs":         filepath.Join(env.DatasetDir, "nixos-configs"),
		"successful_patterns":   filepath.Join(env.DatasetDir, "successful-patterns"),
		"community_solutions":   filepath.Join(env.DatasetDir, "community-solutions"),
		"troubleshooting_cases": filepath.Join(env.DatasetDir, "troubleshooting-cases"),
		"package_examples":      filepath.Join(env.DatasetDir, "package-examples"),
		"service_configs":       filepath.Join(env.DatasetDir, "service-configs"),
	}
}

// GetModelPaths returns paths for model storage and checkpoints
func (env *FineTuningEnvironment) GetModelPaths() map[string]string {
	return map[string]string{
		"checkpoints":  filepath.Join(env.ModelDir, "checkpoints"),
		"trained":      filepath.Join(env.ModelDir, "trained"),
		"base_models":  filepath.Join(env.ModelDir, "base"),
		"experiments":  filepath.Join(env.OutputDir, "experiments"),
		"evaluations":  filepath.Join(env.OutputDir, "evaluations"),
	}
}

// ValidateEnvironment checks if the environment is properly set up
func (env *FineTuningEnvironment) ValidateEnvironment() error {
	// Check if required directories exist
	requiredDirs := []string{
		env.BaseDir,
		env.DatasetDir,
		env.ModelDir,
		env.OutputDir,
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("required directory does not exist: %s", dir)
		}
	}

	// Check if config file exists
	if _, err := os.Stat(env.ConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", env.ConfigPath)
	}

	// Check write permissions
	testFile := filepath.Join(env.OutputDir, ".permission_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("no write permission in output directory: %w", err)
	}
	os.Remove(testFile)

	env.Logger.Info("Environment validation successful")
	return nil
}

// GetStatus returns the current status of the fine-tuning environment
func (env *FineTuningEnvironment) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"base_directory":    env.BaseDir,
		"dataset_directory": env.DatasetDir,
		"model_directory":   env.ModelDir,
		"output_directory":  env.OutputDir,
		"config_path":       env.ConfigPath,
	}

	// Check directory sizes and file counts
	for name, path := range env.GetDatasetPaths() {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			status[fmt.Sprintf("dataset_%s_exists", name)] = true
		} else {
			status[fmt.Sprintf("dataset_%s_exists", name)] = false
		}
	}

	return status
}