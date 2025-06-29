// Package learning provides learning modules, quizzes, and onboarding for NixOS users.
package learning

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// CompetencyArea represents different areas of NixOS knowledge
type CompetencyArea string

const (
	CompetencyNixLanguage     CompetencyArea = "nix_language"
	CompetencyNixOS           CompetencyArea = "nixos"
	CompetencyConfiguration   CompetencyArea = "configuration"
	CompetencyPackaging       CompetencyArea = "packaging"
	CompetencyFlakes          CompetencyArea = "flakes"
	CompetencyHomeManager     CompetencyArea = "home_manager"
	CompetencyDevEnvironments CompetencyArea = "dev_environments"
	CompetencySystemAdmin     CompetencyArea = "system_admin"
	CompetencyTroubleshooting CompetencyArea = "troubleshooting"
	CompetencyDeployment      CompetencyArea = "deployment"
)

// Additional constants for learning module management
const (
	LevelBeginner     = "beginner"
	LevelIntermediate = "intermediate"
	LevelAdvanced     = "advanced"
	LevelExpert       = "expert"
)

// LearningModuleStatus represents the status of a learning module
type LearningModuleStatus string

const (
	StatusNotStarted LearningModuleStatus = "not_started"
	StatusInProgress LearningModuleStatus = "in_progress"
	StatusCompleted  LearningModuleStatus = "completed"
	StatusSkipped    LearningModuleStatus = "skipped"
)

// Module represents a learning module with steps and optional quiz.
type Module struct {
	ID          string
	Title       string
	Description string
	Steps       []Step
	Quiz        *Quiz
	Tags        []string
	Level       string // e.g. "basics", "advanced"
}

type Step struct {
	Title       string
	Instruction string
	Example     string
	Exercise    string
}

type Quiz struct {
	Questions []Question
}

type Question struct {
	Prompt   string
	Choices  []string
	Answer   int // index of correct answer
	Feedback string
}

// Progress tracks user progress through modules and quizzes.
type Progress struct {
	CompletedModules map[string]bool
	QuizScores       map[string]int
}

// LoadModules loads available learning modules with basic built-in modules.
func LoadModules() ([]Module, error) {
	// Return basic built-in modules that complement the Phase 2.2 interactive modules
	modules := []Module{
		{
			ID:          "nix-basics",
			Title:       "Nix Language Basics",
			Description: "Learn fundamental Nix language concepts and syntax",
			Level:       "beginner",
			Tags:        []string{"nix", "language", "basics"},
			Steps: []Step{
				{
					Title:       "Introduction to Nix",
					Instruction: "Nix is a purely functional package manager and build system",
					Example:     "nix --version",
					Exercise:    "Run 'nix --version' to check your Nix installation",
				},
				{
					Title:       "Basic Nix Expressions",
					Instruction: "Learn to write simple Nix expressions",
					Example:     "{ hello = \"world\"; }",
					Exercise:    "Create a simple attribute set with your name",
				},
			},
			Quiz: &Quiz{
				Questions: []Question{
					{
						Prompt:   "What type of package manager is Nix?",
						Choices:  []string{"Imperative", "Functional", "Object-oriented", "Procedural"},
						Answer:   1, // Functional
						Feedback: "Correct! Nix is a purely functional package manager.",
					},
				},
			},
		},
		{
			ID:          "nixos-config",
			Title:       "NixOS Configuration",
			Description: "Learn how to configure your NixOS system",
			Level:       "intermediate",
			Tags:        []string{"nixos", "configuration", "system"},
			Steps: []Step{
				{
					Title:       "Configuration.nix Structure",
					Instruction: "Understand the basic structure of configuration.nix",
					Example:     "{ config, pkgs, ... }: { ... }",
					Exercise:    "Examine your /etc/nixos/configuration.nix file",
				},
				{
					Title:       "System Packages",
					Instruction: "Learn to install system-wide packages",
					Example:     "environment.systemPackages = with pkgs; [ git vim ];",
					Exercise:    "Add a package to your system configuration",
				},
			},
			Quiz: &Quiz{
				Questions: []Question{
					{
						Prompt:   "Where is the main NixOS configuration file located?",
						Choices:  []string{"/etc/nixos/configuration.nix", "/home/user/.nixos", "/var/lib/nixos", "/usr/share/nixos"},
						Answer:   0, // /etc/nixos/configuration.nix
						Feedback: "Correct! The main configuration file is at /etc/nixos/configuration.nix",
					},
				},
			},
		},
		{
			ID:          "flakes-intro",
			Title:       "Introduction to Nix Flakes",
			Description: "Modern Nix project management with flakes",
			Level:       "advanced",
			Tags:        []string{"flakes", "modern", "project-management"},
			Steps: []Step{
				{
					Title:       "What are Flakes?",
					Instruction: "Flakes provide reproducible and composable Nix projects",
					Example:     "nix flake init",
					Exercise:    "Create a new flake in an empty directory",
				},
				{
					Title:       "Flake Structure",
					Instruction: "Understand the basic flake.nix structure",
					Example:     "{ inputs = { nixpkgs.url = \"github:NixOS/nixpkgs\"; }; outputs = ...; }",
					Exercise:    "Examine the inputs and outputs sections",
				},
			},
			Quiz: &Quiz{
				Questions: []Question{
					{
						Prompt:   "What command initializes a new flake?",
						Choices:  []string{"nix init", "nix flake init", "nix create", "nix new"},
						Answer:   1, // nix flake init
						Feedback: "Correct! 'nix flake init' initializes a new flake in the current directory",
					},
				},
			},
		},
	}

	return modules, nil
}

// SaveProgress saves user progress persistently to ~/.config/nixai/learning.yaml.
func SaveProgress(progress Progress) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	progressPath := filepath.Join(configDir, "nixai", "learning.yaml")
	if err := os.MkdirAll(filepath.Dir(progressPath), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(&progress)
	if err != nil {
		return err
	}
	return os.WriteFile(progressPath, data, 0600)
}

// LoadProgress loads user progress from ~/.config/nixai/learning.yaml.
func LoadProgress() (Progress, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return Progress{}, err
	}
	progressPath := filepath.Join(configDir, "nixai", "learning.yaml")
	data, err := os.ReadFile(progressPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Progress{CompletedModules: map[string]bool{}, QuizScores: map[string]int{}}, nil
		}
		return Progress{}, err
	}
	var progress Progress
	if err := yaml.Unmarshal(data, &progress); err != nil {
		return Progress{}, err
	}
	if progress.CompletedModules == nil {
		progress.CompletedModules = map[string]bool{}
	}
	if progress.QuizScores == nil {
		progress.QuizScores = map[string]int{}
	}
	return progress, nil
}

// RenderModule prints a module overview.
func RenderModule(m Module) {
	fmt.Printf("\n# %s\n\n%s\n", m.Title, m.Description)
	for i, step := range m.Steps {
		fmt.Printf("\n%d. %s\n%s\n", i+1, step.Title, step.Instruction)
		if step.Example != "" {
			fmt.Printf("Example:\n%s\n", step.Example)
		}
		if step.Exercise != "" {
			fmt.Printf("Exercise: %s\n", step.Exercise)
		}
	}
	if m.Quiz != nil {
		fmt.Println("\nQuiz available!")
	}
}

// GetModuleByID retrieves a module by its ID
func GetModuleByID(moduleID string) (*Module, error) {
	modules, err := LoadModules()
	if err != nil {
		return nil, err
	}

	for _, module := range modules {
		if module.ID == moduleID {
			return &module, nil
		}
	}

	return nil, fmt.Errorf("module with ID '%s' not found", moduleID)
}

// GetModulesByLevel returns modules filtered by difficulty level
func GetModulesByLevel(level string) ([]Module, error) {
	modules, err := LoadModules()
	if err != nil {
		return nil, err
	}

	var filtered []Module
	for _, module := range modules {
		if module.Level == level {
			filtered = append(filtered, module)
		}
	}

	return filtered, nil
}

// GetModulesByTag returns modules that contain the specified tag
func GetModulesByTag(tag string) ([]Module, error) {
	modules, err := LoadModules()
	if err != nil {
		return nil, err
	}

	var filtered []Module
	for _, module := range modules {
		for _, moduleTag := range module.Tags {
			if moduleTag == tag {
				filtered = append(filtered, module)
				break
			}
		}
	}

	return filtered, nil
}

// ValidateProgress checks if the progress data is valid
func ValidateProgress(progress Progress) error {
	if progress.CompletedModules == nil {
		return fmt.Errorf("completed modules map cannot be nil")
	}
	if progress.QuizScores == nil {
		return fmt.Errorf("quiz scores map cannot be nil")
	}
	return nil
}
