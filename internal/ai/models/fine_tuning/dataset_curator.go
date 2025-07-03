// Package fine_tuning provides dataset curation for NixOS-specific AI training
package fine_tuning

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// DatasetCurator manages the collection and preparation of NixOS-specific training data
type DatasetCurator struct {
	Environment *FineTuningEnvironment
	Logger      *logger.Logger
}

// TrainingExample represents a single training example for the model
type TrainingExample struct {
	ID           string                 `json:"id"`
	Category     string                 `json:"category"`     // "configuration", "troubleshooting", "package", etc.
	Input        string                 `json:"input"`        // User query or problem description
	Output       string                 `json:"output"`       // Expected AI response
	Context      map[string]interface{} `json:"context"`      // Additional context (NixOS version, hardware, etc.)
	Metadata     TrainingMetadata       `json:"metadata"`
	Quality      QualityScore           `json:"quality"`
	Timestamp    time.Time              `json:"timestamp"`
}

// TrainingMetadata contains metadata about the training example
type TrainingMetadata struct {
	Source       string   `json:"source"`        // "manual", "community", "generated"
	Difficulty   string   `json:"difficulty"`    // "beginner", "intermediate", "advanced"
	Tags         []string `json:"tags"`
	Verified     bool     `json:"verified"`      // Whether this example has been verified
	NixOSVersion string   `json:"nixos_version"`
	Language     string   `json:"language"`      // "en", "de", etc.
}

// QualityScore represents the quality assessment of a training example
type QualityScore struct {
	Overall      float64 `json:"overall"`       // 0.0 - 1.0
	Accuracy     float64 `json:"accuracy"`
	Completeness float64 `json:"completeness"`
	Clarity      float64 `json:"clarity"`
	Relevance    float64 `json:"relevance"`
	Automated    bool    `json:"automated"`     // Whether this was scored automatically
}

// DatasetSource represents a source of training data
type DatasetSource struct {
	Name        string
	Type        string   // "file", "directory", "url", "api"
	Path        string
	Enabled     bool
	Categories  []string
	Processor   func(source DatasetSource) ([]TrainingExample, error)
}

// NewDatasetCurator creates a new dataset curator
func NewDatasetCurator(env *FineTuningEnvironment) *DatasetCurator {
	return &DatasetCurator{
		Environment: env,
		Logger:      logger.NewLogger(),
	}
}

// CurateDatasets collects and processes training data from all configured sources
func (dc *DatasetCurator) CurateDatasets(ctx context.Context) error {
	dc.Logger.Info("Starting dataset curation process")

	sources := dc.getDataSources()
	
	var allExamples []TrainingExample
	
	for _, source := range sources {
		if !source.Enabled {
			continue
		}
		
		dc.Logger.Info(fmt.Sprintf("Processing data source: %s (type: %s)", source.Name, source.Type))
		
		examples, err := source.Processor(source)
		if err != nil {
			dc.Logger.Error(fmt.Sprintf("Failed to process data source %s: %v", source.Name, err))
			continue
		}
		
		dc.Logger.Info(fmt.Sprintf("Collected %d examples from source %s", len(examples), source.Name))
		
		allExamples = append(allExamples, examples...)
	}
	
	// Apply quality filtering and enhancement
	filteredExamples := dc.filterAndEnhanceExamples(allExamples)
	
	// Organize examples by category
	categorizedExamples := dc.categorizeExamples(filteredExamples)
	
	// Save processed datasets
	if err := dc.saveDatasets(categorizedExamples); err != nil {
		return fmt.Errorf("failed to save datasets: %w", err)
	}
	
	dc.Logger.Info(fmt.Sprintf("Dataset curation completed: %d total examples, %d filtered examples, %d categories",
		len(allExamples), len(filteredExamples), len(categorizedExamples)))
	
	return nil
}

// getDataSources returns all configured data sources
func (dc *DatasetCurator) getDataSources() []DatasetSource {
	return []DatasetSource{
		{
			Name:       "nixos-manual",
			Type:       "directory",
			Path:       "/etc/nixos",
			Enabled:    true,
			Categories: []string{"configuration", "options"},
			Processor:  dc.processNixOSConfigs,
		},
		{
			Name:       "community-configs",
			Type:       "directory", 
			Path:       filepath.Join(dc.Environment.DatasetDir, "community"),
			Enabled:    true,
			Categories: []string{"configuration", "examples"},
			Processor:  dc.processCommunityConfigs,
		},
		{
			Name:       "troubleshooting-cases",
			Type:       "directory",
			Path:       filepath.Join(dc.Environment.DatasetDir, "troubleshooting"),
			Enabled:    true,
			Categories: []string{"troubleshooting", "errors"},
			Processor:  dc.processTroubleshootingCases,
		},
		{
			Name:       "nixpkgs-examples",
			Type:       "directory",
			Path:       filepath.Join(dc.Environment.DatasetDir, "nixpkgs"),
			Enabled:    true,
			Categories: []string{"packages", "derivations"},
			Processor:  dc.processNixpkgsExamples,
		},
		{
			Name:       "generated-examples",
			Type:       "synthetic",
			Path:       "",
			Enabled:    true,
			Categories: []string{"configuration", "troubleshooting", "packages"},
			Processor:  dc.generateSyntheticExamples,
		},
	}
}

// processNixOSConfigs processes NixOS configuration files
func (dc *DatasetCurator) processNixOSConfigs(source DatasetSource) ([]TrainingExample, error) {
	var examples []TrainingExample
	
	if _, err := os.Stat(source.Path); os.IsNotExist(err) {
		// Create sample data if directory doesn't exist
		return dc.createSampleNixOSExamples(), nil
	}
	
	err := filepath.WalkDir(source.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".nix") {
			return nil
		}
		
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		
		// Extract configuration patterns and create training examples
		configExamples := dc.extractConfigurationExamples(string(content), path)
		examples = append(examples, configExamples...)
		
		return nil
	})
	
	return examples, err
}

// processCommunityConfigs processes community-contributed configurations
func (dc *DatasetCurator) processCommunityConfigs(source DatasetSource) ([]TrainingExample, error) {
	// Create sample community configurations
	return dc.createSampleCommunityExamples(), nil
}

// processTroubleshootingCases processes troubleshooting scenarios
func (dc *DatasetCurator) processTroubleshootingCases(source DatasetSource) ([]TrainingExample, error) {
	// Create sample troubleshooting examples
	return dc.createSampleTroubleshootingExamples(), nil
}

// processNixpkgsExamples processes package-related examples
func (dc *DatasetCurator) processNixpkgsExamples(source DatasetSource) ([]TrainingExample, error) {
	// Create sample package examples
	return dc.createSamplePackageExamples(), nil
}

// generateSyntheticExamples creates synthetic training examples
func (dc *DatasetCurator) generateSyntheticExamples(source DatasetSource) ([]TrainingExample, error) {
	// Generate synthetic examples for data augmentation
	return dc.createSyntheticExamples(), nil
}

// extractConfigurationExamples extracts training examples from configuration content
func (dc *DatasetCurator) extractConfigurationExamples(content, filepath string) []TrainingExample {
	var examples []TrainingExample
	
	// Extract service configurations
	serviceRegex := regexp.MustCompile(`services\.(\w+)\s*=\s*{([^}]+)}`)
	matches := serviceRegex.FindAllStringSubmatch(content, -1)
	
	for i, match := range matches {
		if len(match) >= 3 {
			serviceName := match[1]
			serviceConfig := match[2]
			
			example := TrainingExample{
				ID:       fmt.Sprintf("config_%s_%d", filepath, i),
				Category: "configuration",
				Input:    fmt.Sprintf("How do I configure the %s service in NixOS?", serviceName),
				Output:   fmt.Sprintf("To configure the %s service in NixOS, add the following to your configuration.nix:\n\nservices.%s = {%s};", serviceName, serviceName, serviceConfig),
				Context: map[string]interface{}{
					"file": filepath,
					"service": serviceName,
				},
				Metadata: TrainingMetadata{
					Source:       "nixos-config",
					Difficulty:   "intermediate",
					Tags:         []string{"services", serviceName, "configuration"},
					Verified:     false,
					NixOSVersion: "24.05",
					Language:     "en",
				},
				Quality: QualityScore{
					Overall:      0.8,
					Accuracy:     0.9,
					Completeness: 0.7,
					Clarity:      0.8,
					Relevance:    0.9,
					Automated:    true,
				},
				Timestamp: time.Now(),
			}
			
			examples = append(examples, example)
		}
	}
	
	return examples
}

// createSampleNixOSExamples creates sample NixOS configuration examples
func (dc *DatasetCurator) createSampleNixOSExamples() []TrainingExample {
	return []TrainingExample{
		{
			ID:       "sample_ssh_config",
			Category: "configuration",
			Input:    "How do I enable SSH on NixOS?",
			Output: `To enable SSH on NixOS, add the following to your configuration.nix:

services.openssh = {
  enable = true;
  settings = {
    PasswordAuthentication = false;
    KbdInteractiveAuthentication = false;
    PermitRootLogin = "no";
  };
};

Then rebuild your system with: sudo nixos-rebuild switch`,
			Context: map[string]interface{}{
				"service": "openssh",
				"security": "high",
			},
			Metadata: TrainingMetadata{
				Source:       "manual",
				Difficulty:   "beginner",
				Tags:         []string{"ssh", "security", "services"},
				Verified:     true,
				NixOSVersion: "24.05",
				Language:     "en",
			},
			Quality: QualityScore{
				Overall:      0.95,
				Accuracy:     0.98,
				Completeness: 0.92,
				Clarity:      0.96,
				Relevance:    0.98,
				Automated:    false,
			},
			Timestamp: time.Now(),
		},
	}
}

// createSampleCommunityExamples creates sample community examples
func (dc *DatasetCurator) createSampleCommunityExamples() []TrainingExample {
	return []TrainingExample{
		{
			ID:       "community_docker_setup",
			Category: "configuration",
			Input:    "What's the best way to set up Docker on NixOS?",
			Output: `Here's a robust Docker setup for NixOS:

virtualisation.docker = {
  enable = true;
  rootless = {
    enable = true;
    setSocketVariable = true;
  };
  autoPrune = {
    enable = true;
    dates = "weekly";
  };
};

# Add your user to the docker group
users.users.youruser.extraGroups = [ "docker" ];

For rootless Docker, you might also want to enable:
virtualisation.containers.enable = true;`,
			Context: map[string]interface{}{
				"technology": "docker",
				"virtualization": true,
			},
			Metadata: TrainingMetadata{
				Source:       "community",
				Difficulty:   "intermediate",
				Tags:         []string{"docker", "virtualization", "containers"},
				Verified:     true,
				NixOSVersion: "24.05",
				Language:     "en",
			},
			Quality: QualityScore{
				Overall:      0.9,
				Accuracy:     0.92,
				Completeness: 0.88,
				Clarity:      0.9,
				Relevance:    0.92,
				Automated:    false,
			},
			Timestamp: time.Now(),
		},
	}
}

// createSampleTroubleshootingExamples creates sample troubleshooting examples
func (dc *DatasetCurator) createSampleTroubleshootingExamples() []TrainingExample {
	return []TrainingExample{
		{
			ID:       "troubleshoot_boot_failure",
			Category: "troubleshooting",
			Input:    "My NixOS system won't boot after an update. I see 'Failed to start graphical session target'",
			Output: `This error typically indicates a display manager or graphics driver issue. Here's how to troubleshoot:

1. First, try booting into console mode by removing 'quiet' from the kernel parameters and adding 'systemd.unit=multi-user.target'

2. Check the display manager logs:
   journalctl -u display-manager

3. Common fixes:
   - If using NVIDIA: ensure you have the correct driver version in your configuration
   - For AMD/Intel: try switching to a different display manager temporarily
   - Check if your desktop environment packages are properly installed

4. Recovery steps:
   - Boot from a previous generation using the bootloader
   - Or boot from NixOS live USB and chroot to fix the configuration

5. Quick temporary fix - switch to console login:
   services.xserver.displayManager.gdm.enable = false;
   services.getty.autologinUser = "yourusername";`,
			Context: map[string]interface{}{
				"error_type": "boot_failure",
				"component": "display_manager",
			},
			Metadata: TrainingMetadata{
				Source:       "community",
				Difficulty:   "advanced",
				Tags:         []string{"troubleshooting", "boot", "display-manager", "graphics"},
				Verified:     true,
				NixOSVersion: "24.05",
				Language:     "en",
			},
			Quality: QualityScore{
				Overall:      0.88,
				Accuracy:     0.9,
				Completeness: 0.85,
				Clarity:      0.9,
				Relevance:    0.92,
				Automated:    false,
			},
			Timestamp: time.Now(),
		},
	}
}

// createSamplePackageExamples creates sample package-related examples
func (dc *DatasetCurator) createSamplePackageExamples() []TrainingExample {
	return []TrainingExample{
		{
			ID:       "package_python_env",
			Category: "packages",
			Input:    "How do I create a Python development environment with specific packages?",
			Output: `Here are several ways to create a Python development environment in NixOS:

## Method 1: Using shell.nix (Recommended for development)
Create a shell.nix file:

{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    python311
    python311Packages.pip
    python311Packages.numpy
    python311Packages.pandas
    python311Packages.requests
  ];
}

Then enter the environment with: nix-shell

## Method 2: System-wide installation
Add to your configuration.nix:

environment.systemPackages = with pkgs; [
  (python311.withPackages (ps: with ps; [
    numpy
    pandas
    requests
    matplotlib
  ]))
];

## Method 3: User environment
nix-env -iA nixpkgs.python311Packages.numpy`,
			Context: map[string]interface{}{
				"language": "python",
				"environment": "development",
			},
			Metadata: TrainingMetadata{
				Source:       "manual",
				Difficulty:   "intermediate",
				Tags:         []string{"python", "packages", "development", "nix-shell"},
				Verified:     true,
				NixOSVersion: "24.05",
				Language:     "en",
			},
			Quality: QualityScore{
				Overall:      0.92,
				Accuracy:     0.95,
				Completeness: 0.9,
				Clarity:      0.92,
				Relevance:    0.9,
				Automated:    false,
			},
			Timestamp: time.Now(),
		},
	}
}

// createSyntheticExamples generates synthetic training examples
func (dc *DatasetCurator) createSyntheticExamples() []TrainingExample {
	// This would generate synthetic examples using templates and patterns
	return []TrainingExample{}
}

// filterAndEnhanceExamples applies quality filtering and enhancement
func (dc *DatasetCurator) filterAndEnhanceExamples(examples []TrainingExample) []TrainingExample {
	var filtered []TrainingExample
	
	for _, example := range examples {
		// Apply quality threshold
		if example.Quality.Overall >= 0.7 {
			// Enhance the example (normalize formatting, add tags, etc.)
			enhanced := dc.enhanceExample(example)
			filtered = append(filtered, enhanced)
		}
	}
	
	return filtered
}

// enhanceExample improves the quality of a training example
func (dc *DatasetCurator) enhanceExample(example TrainingExample) TrainingExample {
	// Add automatic tags based on content
	example.Metadata.Tags = append(example.Metadata.Tags, dc.extractTags(example.Input+example.Output)...)
	
	// Normalize formatting
	example.Output = dc.normalizeFormatting(example.Output)
	
	return example
}

// extractTags automatically extracts relevant tags from content
func (dc *DatasetCurator) extractTags(content string) []string {
	var tags []string
	content = strings.ToLower(content)
	
	// Common NixOS terms
	nixTerms := []string{
		"configuration.nix", "services", "packages", "nixpkgs", "derivation",
		"flake", "channel", "generation", "profile", "store", "systemd",
		"user", "group", "firewall", "network", "security", "hardware",
	}
	
	for _, term := range nixTerms {
		if strings.Contains(content, term) {
			tags = append(tags, term)
		}
	}
	
	return tags
}

// normalizeFormatting normalizes the formatting of example content
func (dc *DatasetCurator) normalizeFormatting(content string) string {
	// Remove excessive whitespace
	content = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(content, "\n\n")
	
	// Ensure code blocks are properly formatted
	content = regexp.MustCompile("```\n([^`]+)\n```").ReplaceAllString(content, "```\n$1\n```")
	
	return strings.TrimSpace(content)
}

// categorizeExamples organizes examples by category
func (dc *DatasetCurator) categorizeExamples(examples []TrainingExample) map[string][]TrainingExample {
	categorized := make(map[string][]TrainingExample)
	
	for _, example := range examples {
		categorized[example.Category] = append(categorized[example.Category], example)
	}
	
	return categorized
}

// saveDatasets saves processed datasets to files
func (dc *DatasetCurator) saveDatasets(categorizedExamples map[string][]TrainingExample) error {
	for category, examples := range categorizedExamples {
		filename := fmt.Sprintf("%s_training_data.jsonl", category)
		filepath := filepath.Join(dc.Environment.DatasetDir, filename)
		
		file, err := os.Create(filepath)
		if err != nil {
			return fmt.Errorf("failed to create dataset file %s: %w", filepath, err)
		}
		defer file.Close()
		
		encoder := json.NewEncoder(file)
		for _, example := range examples {
			if err := encoder.Encode(example); err != nil {
				return fmt.Errorf("failed to encode example: %w", err)
			}
		}
		
		dc.Logger.Info(fmt.Sprintf("Saved dataset for category %s with %d examples to %s", category, len(examples), filepath))
	}
	
	return nil
}

// GetDatasetStats returns statistics about the curated datasets
func (dc *DatasetCurator) GetDatasetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	datasetFiles, err := filepath.Glob(filepath.Join(dc.Environment.DatasetDir, "*_training_data.jsonl"))
	if err != nil {
		return nil, err
	}
	
	totalExamples := 0
	categories := make(map[string]int)
	
	for _, file := range datasetFiles {
		// Extract category from filename
		basename := filepath.Base(file)
		category := strings.TrimSuffix(basename, "_training_data.jsonl")
		
		// Count examples in file
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		
		lines := strings.Split(string(content), "\n")
		count := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				count++
			}
		}
		
		categories[category] = count
		totalExamples += count
	}
	
	stats["total_examples"] = totalExamples
	stats["categories"] = categories
	stats["dataset_files"] = len(datasetFiles)
	
	return stats, nil
}