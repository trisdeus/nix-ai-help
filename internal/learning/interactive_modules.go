// Package learning - Interactive Learning Modules
// Advanced Learning System Phase 2.2 Implementation
package learning

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"
)

// InteractiveLearningModule provides guided, interactive learning experiences
type InteractiveLearningModule struct {
	engine     *AdaptiveLearningEngine
	aiProvider ai.Provider
	logger     *logger.Logger
	config     *config.UserConfig
}

// LearningModule represents a comprehensive interactive learning module
type LearningModule struct {
	ID                string             `json:"id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	Level             SkillLevel         `json:"level"`
	Category          string             `json:"category"`
	Tags              []string           `json:"tags"`
	Prerequisites     []string           `json:"prerequisites"`
	EstimatedTime     time.Duration      `json:"estimated_time"`
	LearningPath      []LearningStep     `json:"learning_path"`
	Assessments       []Assessment       `json:"assessments"`
	PracticeExercises []PracticeExercise `json:"practice_exercises"`
	Resources         []LearningResource `json:"resources"`
	Achievements      []Achievement      `json:"achievements"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
	Version           string             `json:"version"`
}

// LearningStep represents a single step in a learning path
type LearningStep struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	Type               StepType        `json:"type"`
	Content            string          `json:"content"`
	InteractionType    InteractionType `json:"interaction_type"`
	ExpectedTime       time.Duration   `json:"expected_time"`
	Prerequisites      []string        `json:"prerequisites"`
	LearningObjectives []string        `json:"learning_objectives"`
	Hints              []string        `json:"hints"`
	Examples           []CodeExample   `json:"examples"`
	Validation         *StepValidation `json:"validation,omitempty"`
	NextSteps          []string        `json:"next_steps"`
	CompletionCriteria []string        `json:"completion_criteria"`
}

// Assessment represents a knowledge check or quiz
type Assessment struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	Type            AssessmentType       `json:"type"`
	Questions       []AssessmentQuestion `json:"questions"`
	PassingScore    float64              `json:"passing_score"`
	TimeLimit       *time.Duration       `json:"time_limit,omitempty"`
	AllowRetakes    bool                 `json:"allow_retakes"`
	FeedbackMode    FeedbackMode         `json:"feedback_mode"`
	WeightedGrading bool                 `json:"weighted_grading"`
}

// PracticeExercise represents hands-on coding or configuration exercises
type PracticeExercise struct {
	ID              string              `json:"id"`
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Type            ExerciseType        `json:"type"`
	Difficulty      DifficultyLevel     `json:"difficulty"`
	StartingCode    string              `json:"starting_code"`
	ExpectedOutput  string              `json:"expected_output"`
	Solution        string              `json:"solution"`
	Hints           []ExerciseHint      `json:"hints"`
	TestCases       []TestCase          `json:"test_cases"`
	Environment     ExerciseEnvironment `json:"environment"`
	ScoringCriteria []ScoringCriterion  `json:"scoring_criteria"`
}

// LearningResource represents additional learning materials
type LearningResource struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Type          ResourceType    `json:"type"`
	URL           string          `json:"url,omitempty"`
	Content       string          `json:"content,omitempty"`
	Description   string          `json:"description"`
	Difficulty    DifficultyLevel `json:"difficulty"`
	EstimatedTime time.Duration   `json:"estimated_time"`
	Tags          []string        `json:"tags"`
}

// Supporting types and enums
type StepType string

const (
	StepIntroduction StepType = "introduction"
	StepContent      StepType = "content"
	StepExample      StepType = "example"
	StepPractice     StepType = "practice"
	StepAssessment   StepType = "assessment"
	StepReflection   StepType = "reflection"
	StepSummary      StepType = "summary"
)

type AssessmentType string

const (
	AssessmentQuiz           AssessmentType = "quiz"
	AssessmentPractical      AssessmentType = "practical"
	AssessmentProject        AssessmentType = "project"
	AssessmentPeerReview     AssessmentType = "peer_review"
	AssessmentSelfReflection AssessmentType = "self_reflection"
)

type QuestionType string

const (
	QuestionMultipleChoice QuestionType = "multiple_choice"
	QuestionTrueFalse      QuestionType = "true_false"
	QuestionFillBlank      QuestionType = "fill_blank"
	QuestionShortAnswer    QuestionType = "short_answer"
	QuestionCode           QuestionType = "code"
	QuestionMatching       QuestionType = "matching"
	QuestionOrdering       QuestionType = "ordering"
)

type FeedbackMode string

const (
	FeedbackImmediate FeedbackMode = "immediate"
	FeedbackAtEnd     FeedbackMode = "at_end"
	FeedbackNone      FeedbackMode = "none"
)

type ExerciseType string

const (
	ExerciseConfiguration   ExerciseType = "configuration"
	ExerciseScripting       ExerciseType = "scripting"
	ExerciseDebugging       ExerciseType = "debugging"
	ExerciseTroubleshooting ExerciseType = "troubleshooting"
	ExerciseOptimization    ExerciseType = "optimization"
	ExerciseDesign          ExerciseType = "design"
)

// AssessmentQuestion represents a question for module assessments
type AssessmentQuestion struct {
	ID            string          `json:"id"`
	Type          QuestionType    `json:"type"`
	Question      string          `json:"question"`
	Options       []string        `json:"options,omitempty"`
	CorrectAnswer interface{}     `json:"correct_answer"`
	Explanation   string          `json:"explanation"`
	Difficulty    DifficultyLevel `json:"difficulty"`
	Points        int             `json:"points"`
	Tags          []string        `json:"tags"`
	Hints         []string        `json:"hints,omitempty"`
	CodeSnippet   string          `json:"code_snippet,omitempty"`
}

type ExerciseHint struct {
	Level   int    `json:"level"` // 1-3, increasing specificity
	Content string `json:"content"`
	Type    string `json:"type"` // "general", "specific", "code"
}

type TestCase struct {
	ID          string      `json:"id"`
	Input       interface{} `json:"input"`
	Expected    interface{} `json:"expected"`
	Description string      `json:"description"`
	Points      int         `json:"points"`
	Hidden      bool        `json:"hidden"` // Not shown to user
}

type ExerciseEnvironment struct {
	Type             string            `json:"type"` // "nixos", "nix-shell", "docker", "vm"
	BaseImage        string            `json:"base_image,omitempty"`
	RequiredPackages []string          `json:"required_packages"`
	PresetFiles      map[string]string `json:"preset_files"`
	EnvironmentVars  map[string]string `json:"environment_vars"`
	ResourceLimits   ResourceLimits    `json:"resource_limits"`
}

type ResourceLimits struct {
	CPULimit    string        `json:"cpu_limit,omitempty"`
	MemoryLimit string        `json:"memory_limit,omitempty"`
	TimeLimit   time.Duration `json:"time_limit,omitempty"`
}

type ScoringCriterion struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MaxPoints   int     `json:"max_points"`
	Weight      float64 `json:"weight"`
	AutoGraded  bool    `json:"auto_graded"`
}

type ResourceType string

const (
	ResourceDocumentation ResourceType = "documentation"
	ResourceVideo         ResourceType = "video"
	ResourceBlog          ResourceType = "blog"
	ResourceTutorial      ResourceType = "tutorial"
	ResourceReference     ResourceType = "reference"
	ResourceExample       ResourceType = "example"
	ResourceTools         ResourceType = "tools"
)

type CodeExample struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Output      string `json:"output,omitempty"`
	Explanation string `json:"explanation"`
}

type StepValidation struct {
	Type           ValidationType `json:"type"`
	Criteria       []string       `json:"criteria"`
	AutoCheck      bool           `json:"auto_check"`
	Command        string         `json:"command,omitempty"`
	ExpectedResult string         `json:"expected_result,omitempty"`
}

type ValidationType string

const (
	ValidationCommand ValidationType = "command"
	ValidationFile    ValidationType = "file"
	ValidationService ValidationType = "service"
	ValidationOutput  ValidationType = "output"
	ValidationManual  ValidationType = "manual"
)

// NewInteractiveLearningModule creates a new interactive learning module system
func NewInteractiveLearningModule(engine *AdaptiveLearningEngine, aiProvider ai.Provider, cfg *config.UserConfig, log *logger.Logger) (*InteractiveLearningModule, error) {
	if engine == nil {
		return nil, fmt.Errorf("adaptive learning engine cannot be nil")
	}

	return &InteractiveLearningModule{
		engine:     engine,
		aiProvider: aiProvider,
		logger:     log,
		config:     cfg,
	}, nil
}

// GetAvailableModules returns learning modules suitable for the user
func (ilm *InteractiveLearningModule) GetAvailableModules() ([]LearningModule, error) {
	modules := ilm.getBuiltInModules()

	// Filter and sort based on user profile
	if ilm.engine != nil && ilm.engine.userProfile != nil {
		modules = ilm.filterModulesForUser(modules)
		modules = ilm.sortModulesByRelevance(modules)
	}

	return modules, nil
}

// StartModule begins an interactive learning module session
func (ilm *InteractiveLearningModule) StartModule(ctx context.Context, moduleID string) (*ModuleSession, error) {
	modules, err := ilm.GetAvailableModules()
	if err != nil {
		return nil, fmt.Errorf("failed to get available modules: %w", err)
	}

	var selectedModule *LearningModule
	for _, module := range modules {
		if module.ID == moduleID {
			selectedModule = &module
			break
		}
	}

	if selectedModule == nil {
		return nil, fmt.Errorf("module not found: %s", moduleID)
	}

	// Check prerequisites
	if err := ilm.validatePrerequisites(*selectedModule); err != nil {
		return nil, fmt.Errorf("prerequisites not met: %w", err)
	}

	session := &ModuleSession{
		ID:                utils.HashString(fmt.Sprintf("%s_%d", moduleID, time.Now().Unix())),
		ModuleID:          moduleID,
		Module:            *selectedModule,
		UserID:            ilm.engine.userProfile.UserID,
		StartTime:         time.Now(),
		CurrentStepIndex:  0,
		StepProgress:      make(map[string]StepProgress),
		AssessmentResults: make(map[string]AssessmentResult),
		Status:            SessionActive,
		AdaptiveSettings:  ilm.calculateAdaptiveSettings(*selectedModule),
	}

	// Record session start
	if ilm.engine != nil {
		ilm.engine.RecordInteraction(
			InteractionTutorial,
			fmt.Sprintf("Started module: %s", selectedModule.Title),
			map[string]interface{}{
				"module_id": moduleID,
				"topic":     selectedModule.Category,
			},
			0,
			true,
		)
	}

	ilm.logger.Info(fmt.Sprintf("Started learning module session: %s", session.ID))
	return session, nil
}

// ModuleSession represents an active learning module session
type ModuleSession struct {
	ID                string                      `json:"id"`
	ModuleID          string                      `json:"module_id"`
	Module            LearningModule              `json:"module"`
	UserID            string                      `json:"user_id"`
	StartTime         time.Time                   `json:"start_time"`
	EndTime           *time.Time                  `json:"end_time,omitempty"`
	CurrentStepIndex  int                         `json:"current_step_index"`
	StepProgress      map[string]StepProgress     `json:"step_progress"`
	AssessmentResults map[string]AssessmentResult `json:"assessment_results"`
	Status            SessionStatus               `json:"status"`
	Score             float64                     `json:"score"`
	TimeSpent         time.Duration               `json:"time_spent"`
	HintsUsed         int                         `json:"hints_used"`
	AdaptiveSettings  AdaptiveSettings            `json:"adaptive_settings"`
	Notes             []SessionNote               `json:"notes"`
}

type StepProgress struct {
	StepID       string        `json:"step_id"`
	Status       StepStatus    `json:"status"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	TimeSpent    time.Duration `json:"time_spent"`
	Attempts     int           `json:"attempts"`
	HintsUsed    int           `json:"hints_used"`
	Score        float64       `json:"score"`
	UserResponse interface{}   `json:"user_response,omitempty"`
}

type AssessmentResult struct {
	AssessmentID    string           `json:"assessment_id"`
	Score           float64          `json:"score"`
	MaxScore        float64          `json:"max_score"`
	Percentage      float64          `json:"percentage"`
	Passed          bool             `json:"passed"`
	TimeSpent       time.Duration    `json:"time_spent"`
	Attempts        int              `json:"attempts"`
	QuestionResults []QuestionResult `json:"question_results"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
}

type QuestionResult struct {
	QuestionID    string        `json:"question_id"`
	UserAnswer    interface{}   `json:"user_answer"`
	CorrectAnswer interface{}   `json:"correct_answer"`
	IsCorrect     bool          `json:"is_correct"`
	Points        int           `json:"points"`
	MaxPoints     int           `json:"max_points"`
	TimeSpent     time.Duration `json:"time_spent"`
	HintsUsed     int           `json:"hints_used"`
}

type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
	SessionPaused    SessionStatus = "paused"
	SessionAbandoned SessionStatus = "abandoned"
)

type StepStatus string

const (
	StepNotStarted StepStatus = "not_started"
	StepInProgress StepStatus = "in_progress"
	StepCompleted  StepStatus = "completed"
	StepSkipped    StepStatus = "skipped"
	StepFailed     StepStatus = "failed"
)

type AdaptiveSettings struct {
	DifficultyAdjustment float64 `json:"difficulty_adjustment"`
	HintsEnabled         bool    `json:"hints_enabled"`
	TimeExtensions       bool    `json:"time_extensions"`
	PersonalizedContent  bool    `json:"personalized_content"`
	SkipOptional         bool    `json:"skip_optional"`
	ExtraExercises       bool    `json:"extra_exercises"`
	PacingAdjustment     float64 `json:"pacing_adjustment"`
}

type SessionNote struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	StepID    string    `json:"step_id,omitempty"`
}

// GetBuiltInModules returns the core learning modules
func (ilm *InteractiveLearningModule) getBuiltInModules() []LearningModule {
	return []LearningModule{
		ilm.createNixBasicsModule(),
		ilm.createNixOSBasicsModule(),
		ilm.createConfigurationManagementModule(),
		ilm.createFlakesModule(),
		ilm.createModulesModule(),
		ilm.createPackageManagementModule(),
		ilm.createServiceManagementModule(),
		ilm.createTroubleshootingModule(),
		ilm.createAdvancedTopicsModule(),
	}
}

// Create specific learning modules
func (ilm *InteractiveLearningModule) createNixBasicsModule() LearningModule {
	return LearningModule{
		ID:            "nix-basics",
		Title:         "Nix Package Manager Fundamentals",
		Description:   "Learn the core concepts and commands of the Nix package manager",
		Level:         SkillBeginner,
		Category:      "nix-basics",
		Tags:          []string{"nix", "fundamentals", "package-manager"},
		Prerequisites: []string{},
		EstimatedTime: 45 * time.Minute,
		LearningPath: []LearningStep{
			{
				ID:    "intro-to-nix",
				Title: "Introduction to Nix",
				Type:  StepIntroduction,
				Content: `# Welcome to Nix!

Nix is a revolutionary package manager that brings reproducible, declarative package management to any Unix-like system.

## Key Benefits:
- **Reproducible**: Same packages work the same way everywhere
- **Declarative**: Describe what you want, not how to get it
- **Safe**: Atomic upgrades and rollbacks
- **Multiple versions**: Install multiple versions side by side

Let's start your journey into the world of Nix!`,
				InteractionType: InteractionTutorial,
				ExpectedTime:    5 * time.Minute,
				LearningObjectives: []string{
					"Understand what Nix is and its core benefits",
					"Recognize the problems Nix solves",
				},
			},
			{
				ID:    "basic-commands",
				Title: "Basic Nix Commands",
				Type:  StepContent,
				Content: `# Essential Nix Commands

Here are the fundamental commands you'll use every day:

## Package Management
- **nix search**: Find packages
- **nix-env -iA**: Install packages
- **nix-env -q**: List installed packages
- **nix-env -e**: Remove packages

## Environment Management
- **nix-shell**: Enter development environment
- **nix-shell -p**: Quick environment with specific packages

Let's try some examples!`,
				InteractionType: InteractionTutorial,
				ExpectedTime:    10 * time.Minute,
				Examples: []CodeExample{
					{
						ID:          "search-example",
						Title:       "Searching for packages",
						Description: "How to find packages in nixpkgs",
						Language:    "bash",
						Code:        "nix search nixpkgs firefox",
						Output:      "* legacyPackages.x86_64-linux.firefox-bin (115.0.2)\n  Mozilla Firefox, free web browser (binary package)",
						Explanation: "This searches for Firefox in the nixpkgs repository",
					},
					{
						ID:          "install-example",
						Title:       "Installing a package",
						Description: "How to install packages with nix-env",
						Language:    "bash",
						Code:        "nix-env -iA nixpkgs.firefox",
						Output:      "installing 'firefox-115.0.2'",
						Explanation: "This installs Firefox to your user profile",
					},
				},
				Hints: []string{
					"Always use -iA flag for faster installs",
					"Tab completion works with package names",
					"Use --dry-run to see what would be installed",
				},
			},
			{
				ID:    "practice-commands",
				Title: "Practice: Your First Nix Commands",
				Type:  StepPractice,
				Content: `# Hands-on Practice

Time to try Nix commands yourself! Complete these exercises:

1. Search for the 'git' package
2. Install git using nix-env
3. List your installed packages
4. Enter a shell with python3 available

Don't worry if you make mistakes - Nix makes it easy to recover!`,
				InteractionType: InteractionPractice,
				ExpectedTime:    15 * time.Minute,
				Validation: &StepValidation{
					Type:      ValidationCommand,
					AutoCheck: true,
					Criteria: []string{
						"git --version returns successfully",
						"nix-env -q | grep git shows git installed",
					},
				},
				CompletionCriteria: []string{
					"Successfully search for a package",
					"Install a package using nix-env",
					"Verify installation",
				},
			},
		},
		Assessments: []Assessment{
			{
				ID:           "nix-basics-quiz",
				Title:        "Nix Basics Knowledge Check",
				Type:         AssessmentQuiz,
				PassingScore: 0.7,
				FeedbackMode: FeedbackImmediate,
				Questions: []AssessmentQuestion{
					{
						ID:       "nix-benefits",
						Type:     QuestionMultipleChoice,
						Question: "Which of the following is NOT a key benefit of Nix?",
						Options: []string{
							"Reproducible package management",
							"Faster package installation than other managers",
							"Atomic upgrades and rollbacks",
							"Multiple package versions side by side",
						},
						CorrectAnswer: 1, // Index of correct answer
						Explanation:   "While Nix has many benefits, speed of installation is not its primary advantage. It focuses on reproducibility and safety.",
						Difficulty:    DifficultyEasy,
						Points:        10,
					},
					{
						ID:            "install-command",
						Type:          QuestionFillBlank,
						Question:      "To install Firefox using nix-env, you would use: nix-env _____ nixpkgs.firefox",
						CorrectAnswer: "-iA",
						Explanation:   "The -iA flag installs a package by attribute path, which is faster and more reliable.",
						Difficulty:    DifficultyMedium,
						Points:        15,
					},
				},
			},
		},
		PracticeExercises: []PracticeExercise{
			{
				ID:           "dev-environment",
				Title:        "Create a Development Environment",
				Description:  "Use nix-shell to create a development environment with specific tools",
				Type:         ExerciseConfiguration,
				Difficulty:   DifficultyMedium,
				StartingCode: "# Create a shell.nix file for a Python development environment\n",
				Solution: `{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    python3
    python3Packages.pip
    python3Packages.virtualenv
  ];
}`,
				TestCases: []TestCase{
					{
						ID:          "python-available",
						Description: "Python3 should be available in the shell",
						Expected:    "Python 3",
						Points:      20,
					},
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}
}

func (ilm *InteractiveLearningModule) createNixOSBasicsModule() LearningModule {
	return LearningModule{
		ID:            "nixos-basics",
		Title:         "NixOS Operating System Fundamentals",
		Description:   "Learn how to configure and manage your NixOS system",
		Level:         SkillBeginner,
		Category:      "nixos-basics",
		Tags:          []string{"nixos", "configuration", "system-management"},
		Prerequisites: []string{"nix-basics"},
		EstimatedTime: 60 * time.Minute,
		LearningPath: []LearningStep{
			{
				ID:    "nixos-intro",
				Title: "What Makes NixOS Special",
				Type:  StepIntroduction,
				Content: `# NixOS: The Purely Functional Linux Distribution

NixOS takes the Nix package manager philosophy and applies it to an entire operating system.

## Core Concepts:
- **Declarative Configuration**: Your entire system defined in configuration.nix
- **Atomic Upgrades**: System changes are atomic and reversible
- **Reproducible Systems**: Same configuration = same system, everywhere
- **Rollbacks**: Boot into any previous configuration

## The NixOS Way:
Instead of imperatively changing your system, you declare what you want and NixOS builds it.`,
				InteractionType: InteractionTutorial,
				ExpectedTime:    8 * time.Minute,
			},
			{
				ID:    "configuration-structure",
				Title: "Understanding configuration.nix",
				Type:  StepContent,
				Content: `# The Heart of NixOS: configuration.nix

Your system configuration lives in /etc/nixos/configuration.nix. Let's explore its structure:

## Basic Structure:
` + "```nix" + `
{ config, pkgs, ... }:

{
  # System configuration options go here
  
  # Basic system info
  system.stateVersion = "23.05";
  
  # Networking
  networking.hostName = "my-nixos";
  
  # Users
  users.users.myuser = {
    isNormalUser = true;
    extraGroups = [ "wheel" "sudo" ];
  };
  
  # Packages
  environment.systemPackages = with pkgs; [
    firefox
    git
    vim
  ];
}
` + "```" + `
  ];
}
` + "```" + `

Every aspect of your system can be configured declaratively!`,
				InteractionType: InteractionTutorial,
				ExpectedTime:    15 * time.Minute,
				Examples: []CodeExample{
					{
						ID:          "basic-config",
						Title:       "Basic configuration.nix",
						Description: "A minimal NixOS configuration",
						Language:    "nix",
						Code: `{ config, pkgs, ... }:

{
  imports = [ ./hardware-configuration.nix ];

  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  networking.hostName = "mynixos";
  networking.networkmanager.enable = true;

  users.users.alice = {
    isNormalUser = true;
    description = "Alice";
    extraGroups = [ "networkmanager" "wheel" ];
  };

  environment.systemPackages = with pkgs; [
    firefox
    git
    tree
  ];

  system.stateVersion = "23.05";
}`,
						Explanation: "This configuration sets up a basic NixOS system with a user, network manager, and some essential packages.",
					},
				},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   "1.0",
	}
}

func (ilm *InteractiveLearningModule) createConfigurationManagementModule() LearningModule {
	return LearningModule{
		ID:            "configuration-management",
		Title:         "Advanced NixOS Configuration Management",
		Description:   "Master advanced configuration techniques and system management",
		Level:         SkillIntermediate,
		Category:      "configuration-management",
		Tags:          []string{"configuration", "management", "advanced"},
		Prerequisites: []string{"nixos-basics"},
		EstimatedTime: 75 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

func (ilm *InteractiveLearningModule) createFlakesModule() LearningModule {
	return LearningModule{
		ID:            "flakes",
		Title:         "Nix Flakes: Modern Nix Project Management",
		Description:   "Learn to use Nix flakes for reproducible project environments",
		Level:         SkillIntermediate,
		Category:      "flakes",
		Tags:          []string{"flakes", "modern-nix", "reproducibility"},
		Prerequisites: []string{"nix-basics", "nixos-basics"},
		EstimatedTime: 90 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

func (ilm *InteractiveLearningModule) createModulesModule() LearningModule {
	return LearningModule{
		ID:            "modules",
		Title:         "NixOS Modules: Extending System Configuration",
		Description:   "Create custom NixOS modules for reusable configuration components",
		Level:         SkillAdvanced,
		Category:      "modules",
		Tags:          []string{"modules", "configuration", "custom"},
		Prerequisites: []string{"nixos-basics", "configuration-management"},
		EstimatedTime: 120 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

func (ilm *InteractiveLearningModule) createPackageManagementModule() LearningModule {
	return LearningModule{
		ID:            "package-management",
		Title:         "Advanced Package Management with Nix",
		Description:   "Master package installation, overlays, and custom derivations",
		Level:         SkillIntermediate,
		Category:      "package-management",
		Tags:          []string{"packages", "overlays", "derivations"},
		Prerequisites: []string{"nix-basics"},
		EstimatedTime: 80 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

func (ilm *InteractiveLearningModule) createServiceManagementModule() LearningModule {
	return LearningModule{
		ID:            "service-management",
		Title:         "Managing Services in NixOS",
		Description:   "Configure and manage system services using NixOS modules",
		Level:         SkillIntermediate,
		Category:      "service-management",
		Tags:          []string{"services", "systemd", "configuration"},
		Prerequisites: []string{"nixos-basics", "configuration-management"},
		EstimatedTime: 70 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

func (ilm *InteractiveLearningModule) createTroubleshootingModule() LearningModule {
	return LearningModule{
		ID:            "troubleshooting",
		Title:         "NixOS Troubleshooting and Debugging",
		Description:   "Learn to diagnose and fix common NixOS issues",
		Level:         SkillIntermediate,
		Category:      "troubleshooting",
		Tags:          []string{"troubleshooting", "debugging", "problem-solving"},
		Prerequisites: []string{"nixos-basics"},
		EstimatedTime: 85 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

func (ilm *InteractiveLearningModule) createAdvancedTopicsModule() LearningModule {
	return LearningModule{
		ID:            "advanced-topics",
		Title:         "Advanced NixOS Concepts",
		Description:   "Explore advanced topics: cross-compilation, containers, and more",
		Level:         SkillAdvanced,
		Category:      "advanced",
		Tags:          []string{"advanced", "cross-compilation", "containers"},
		Prerequisites: []string{"nixos-basics", "configuration-management", "modules"},
		EstimatedTime: 150 * time.Minute,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       "1.0",
	}
}

// Helper methods for module management
func (ilm *InteractiveLearningModule) filterModulesForUser(modules []LearningModule) []LearningModule {
	if ilm.engine == nil || ilm.engine.userProfile == nil {
		return modules
	}

	filtered := []LearningModule{}
	userProfile := ilm.engine.userProfile

	for _, module := range modules {
		// Check if user meets prerequisites
		if ilm.hasPrerequisites(module, userProfile) {
			// Check if module matches user's skill level
			if ilm.isAppropriateLevel(module, userProfile) {
				filtered = append(filtered, module)
			}
		}
	}

	return filtered
}

func (ilm *InteractiveLearningModule) hasPrerequisites(module LearningModule, profile *UserProfile) bool {
	for _, prereq := range module.Prerequisites {
		if competency, exists := profile.CompetencyMap[prereq]; !exists || competency < CompetencyBasic {
			return false
		}
	}
	return true
}

func (ilm *InteractiveLearningModule) isAppropriateLevel(module LearningModule, profile *UserProfile) bool {
	// Allow modules at or slightly above user's level
	userLevelNum := ilm.skillLevelToNumber(profile.SkillLevel)
	moduleLevelNum := ilm.skillLevelToNumber(module.Level)

	return moduleLevelNum <= userLevelNum+1 // Allow one level above
}

func (ilm *InteractiveLearningModule) skillLevelToNumber(level SkillLevel) int {
	switch level {
	case SkillBeginner:
		return 1
	case SkillIntermediate:
		return 2
	case SkillAdvanced:
		return 3
	case SkillExpert:
		return 4
	default:
		return 1
	}
}

func (ilm *InteractiveLearningModule) sortModulesByRelevance(modules []LearningModule) []LearningModule {
	if ilm.engine == nil || ilm.engine.userProfile == nil {
		return modules
	}

	sort.Slice(modules, func(i, j int) bool {
		scoreI := ilm.calculateRelevanceScore(modules[i])
		scoreJ := ilm.calculateRelevanceScore(modules[j])
		return scoreI > scoreJ
	})

	return modules
}

func (ilm *InteractiveLearningModule) calculateRelevanceScore(module LearningModule) float64 {
	score := 0.0
	userProfile := ilm.engine.userProfile

	// Boost score for preferred topics
	for _, preferred := range userProfile.Preferences.PreferredTopics {
		if strings.Contains(module.Category, preferred) {
			score += 3.0
		}
		for _, tag := range module.Tags {
			if strings.Contains(tag, preferred) {
				score += 1.0
			}
		}
	}

	// Reduce score for avoided topics
	for _, avoided := range userProfile.Preferences.AvoidedTopics {
		if strings.Contains(module.Category, avoided) {
			score -= 2.0
		}
	}

	// Boost score for weak areas
	for _, weak := range userProfile.WeakAreas {
		if strings.Contains(module.Category, weak) {
			score += 2.0
		}
	}

	// Consider current competency in topic
	if competency, exists := userProfile.CompetencyMap[module.Category]; exists {
		switch competency {
		case CompetencyNone:
			score += 1.5 // Boost for totally new topics
		case CompetencyBeginner:
			score += 1.0
		case CompetencyMastery:
			score -= 1.0 // Less relevant if mastered
		}
	}

	// Consider module difficulty vs user preferences
	moduleDifficulty := ilm.moduleToTargetDifficulty(module)
	if moduleDifficulty == ilm.competencyToDifficulty(userProfile.Preferences.PreferredDifficulty) {
		score += 1.0
	}

	return score
}

func (ilm *InteractiveLearningModule) moduleToTargetDifficulty(module LearningModule) DifficultyLevel {
	switch module.Level {
	case SkillBeginner:
		return DifficultyEasy
	case SkillIntermediate:
		return DifficultyMedium
	case SkillAdvanced:
		return DifficultyHard
	case SkillExpert:
		return DifficultyVeryHard
	default:
		return DifficultyMedium
	}
}

func (ilm *InteractiveLearningModule) competencyToDifficulty(preferred DifficultyLevel) DifficultyLevel {
	return preferred // Direct mapping for now
}

func (ilm *InteractiveLearningModule) validatePrerequisites(module LearningModule) error {
	if ilm.engine == nil || ilm.engine.userProfile == nil {
		// Allow if no profile (first time users)
		return nil
	}

	missingPrereqs := []string{}
	for _, prereq := range module.Prerequisites {
		if competency, exists := ilm.engine.userProfile.CompetencyMap[prereq]; !exists || competency < CompetencyBasic {
			missingPrereqs = append(missingPrereqs, prereq)
		}
	}

	if len(missingPrereqs) > 0 {
		return fmt.Errorf("missing prerequisites: %s", strings.Join(missingPrereqs, ", "))
	}

	return nil
}

func (ilm *InteractiveLearningModule) calculateAdaptiveSettings(module LearningModule) AdaptiveSettings {
	settings := AdaptiveSettings{
		DifficultyAdjustment: 0.0,
		HintsEnabled:         true,
		TimeExtensions:       true,
		PersonalizedContent:  true,
		SkipOptional:         false,
		ExtraExercises:       false,
		PacingAdjustment:     1.0,
	}

	if ilm.engine != nil && ilm.engine.userProfile != nil {
		profile := ilm.engine.userProfile

		// Adjust based on user preferences
		if profile.Preferences.AdaptiveDifficulty {
			// Calculate difficulty adjustment based on recent performance
			recentInteractions := ilm.engine.getRecentInteractions(10)
			if len(recentInteractions) > 0 {
				successRate := ilm.engine.calculateSuccessRate(recentInteractions)
				if successRate > 0.8 {
					settings.DifficultyAdjustment = 0.2 // Increase difficulty
					settings.ExtraExercises = true
				} else if successRate < 0.4 {
					settings.DifficultyAdjustment = -0.2 // Decrease difficulty
					settings.SkipOptional = true
				}
			}
		}

		// Adjust pacing based on user preferences
		preferredDuration := profile.Preferences.SessionLength
		if preferredDuration < 30*time.Minute {
			settings.PacingAdjustment = 0.8 // Faster pacing
		} else if preferredDuration > 60*time.Minute {
			settings.PacingAdjustment = 1.2 // Slower pacing
		}

		// Disable gamification if user prefers
		if !profile.Preferences.GameificationEnabled {
			settings.ExtraExercises = false
		}
	}

	return settings
}

// ExecuteStep processes a learning step in the current session
func (ilm *InteractiveLearningModule) ExecuteStep(ctx context.Context, session *ModuleSession, stepIndex int) (*StepExecution, error) {
	if stepIndex >= len(session.Module.LearningPath) {
		return nil, fmt.Errorf("invalid step index: %d", stepIndex)
	}

	step := session.Module.LearningPath[stepIndex]

	// Check if step was already completed
	if progress, exists := session.StepProgress[step.ID]; exists && progress.Status == StepCompleted {
		return &StepExecution{
			Step:     &step,
			Status:   StepCompleted,
			Message:  "Step already completed",
			Progress: progress,
		}, nil
	}

	// Initialize step progress if not exists
	if _, exists := session.StepProgress[step.ID]; !exists {
		session.StepProgress[step.ID] = StepProgress{
			StepID:    step.ID,
			Status:    StepInProgress,
			StartTime: time.Now(),
			Attempts:  1,
		}
	}

	execution := &StepExecution{
		Step:                &step,
		Status:              StepInProgress,
		Message:             "Step ready for interaction",
		InteractiveElements: ilm.generateInteractiveElements(step),
		PersonalizedContent: ilm.generatePersonalizedStepContent(ctx, step, session),
	}

	// Record step start
	if ilm.engine != nil {
		ilm.engine.RecordInteraction(
			step.InteractionType,
			fmt.Sprintf("Started step: %s", step.Title),
			map[string]interface{}{
				"module_id": session.ModuleID,
				"step_id":   step.ID,
				"step_type": step.Type,
			},
			0,
			true,
		)
	}

	return execution, nil
}

// StepExecution represents the result of executing a learning step
type StepExecution struct {
	Step                *LearningStep        `json:"step"`
	Status              StepStatus           `json:"status"`
	Message             string               `json:"message"`
	Progress            StepProgress         `json:"progress"`
	InteractiveElements []InteractiveElement `json:"interactive_elements"`
	PersonalizedContent string               `json:"personalized_content,omitempty"`
	ValidationResult    *ValidationResult    `json:"validation_result,omitempty"`
}

type InteractiveElement struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Content  string                 `json:"content"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Required bool                   `json:"required"`
}

type ValidationResult struct {
	Success  bool     `json:"success"`
	Score    float64  `json:"score"`
	Feedback string   `json:"feedback"`
	Errors   []string `json:"errors,omitempty"`
	Hints    []string `json:"hints,omitempty"`
}

func (ilm *InteractiveLearningModule) generateInteractiveElements(step LearningStep) []InteractiveElement {
	elements := []InteractiveElement{}

	switch step.Type {
	case StepPractice:
		elements = append(elements, InteractiveElement{
			ID:      "code-editor",
			Type:    "code_editor",
			Content: "Interactive code editor for practice",
			Options: map[string]interface{}{
				"language":     "nix",
				"startingCode": "",
				"autoComplete": true,
			},
			Required: true,
		})

	case StepAssessment:
		elements = append(elements, InteractiveElement{
			ID:      "quiz-interface",
			Type:    "quiz",
			Content: "Interactive quiz interface",
			Options: map[string]interface{}{
				"allowRetries": true,
				"showHints":    true,
			},
			Required: true,
		})

	default:
		elements = append(elements, InteractiveElement{
			ID:      "content-reader",
			Type:    "content",
			Content: step.Content,
			Options: map[string]interface{}{
				"renderMarkdown":  true,
				"syntaxHighlight": true,
			},
			Required: false,
		})
	}

	// Add hint system if hints are available
	if len(step.Hints) > 0 {
		elements = append(elements, InteractiveElement{
			ID:      "hint-system",
			Type:    "hints",
			Content: "Available hints for this step",
			Options: map[string]interface{}{
				"hints":      step.Hints,
				"sequential": true,
			},
			Required: false,
		})
	}

	return elements
}

func (ilm *InteractiveLearningModule) generatePersonalizedStepContent(ctx context.Context, step LearningStep, session *ModuleSession) string {
	if ilm.aiProvider == nil || ilm.engine == nil {
		return step.Content
	}

	// Build personalization context
	context := ilm.buildStepPersonalizationContext(step, session)

	prompt := fmt.Sprintf(`Personalize this learning step content for the user:

Original Step: %s
Content: %s

User Context:
- Skill Level: %v
- Learning Style: %v
- Recent Performance: %v
- Session Progress: %d/%d steps completed

Instructions:
1. Adapt the content to the user's learning style
2. Adjust difficulty based on their performance
3. Add relevant examples or explanations
4. Keep the core learning objectives intact
5. Make it engaging and personalized

Return the personalized content in markdown format.`,
		step.Title,
		step.Content,
		context["skill_level"],
		context["learning_style"],
		context["recent_performance"],
		session.CurrentStepIndex,
		len(session.Module.LearningPath),
	)

	// Generate personalized content
	response, err := ilm.aiProvider.Query(prompt)
	if err != nil {
		ilm.logger.Warn(fmt.Sprintf("Failed to generate personalized content: %v", err))
		return step.Content
	}

	return response
}

func (ilm *InteractiveLearningModule) buildStepPersonalizationContext(step LearningStep, session *ModuleSession) map[string]interface{} {
	context := map[string]interface{}{
		"step_type":        step.Type,
		"step_title":       step.Title,
		"module_id":        session.ModuleID,
		"session_progress": float64(session.CurrentStepIndex) / float64(len(session.Module.LearningPath)),
	}

	if ilm.engine != nil && ilm.engine.userProfile != nil {
		profile := ilm.engine.userProfile
		context["skill_level"] = profile.SkillLevel
		context["learning_style"] = profile.LearningStyle
		context["personality_type"] = profile.PersonalityType

		// Add recent performance context
		recentInteractions := ilm.engine.getRecentInteractions(5)
		context["recent_performance"] = ilm.engine.analyzeRecentPerformance()
		context["recent_success_rate"] = ilm.engine.calculateSuccessRate(recentInteractions)
	}

	return context
}

// CompleteStep marks a step as completed and updates progress
func (ilm *InteractiveLearningModule) CompleteStep(session *ModuleSession, stepID string, userResponse interface{}, timeSpent time.Duration) error {
	// Find the step
	var step *LearningStep
	for _, s := range session.Module.LearningPath {
		if s.ID == stepID {
			step = &s
			break
		}
	}

	if step == nil {
		return fmt.Errorf("step not found: %s", stepID)
	}

	// Update step progress
	progress := session.StepProgress[stepID]
	endTime := time.Now()
	progress.EndTime = &endTime
	progress.TimeSpent = timeSpent
	progress.Status = StepCompleted
	progress.UserResponse = userResponse
	progress.Score = ilm.calculateStepScore(*step, userResponse, timeSpent)
	session.StepProgress[stepID] = progress

	// Update session progress
	if session.CurrentStepIndex < len(session.Module.LearningPath)-1 {
		session.CurrentStepIndex++
	}

	// Calculate overall session score
	session.Score = ilm.calculateSessionScore(session)
	session.TimeSpent += timeSpent

	// Record completion
	if ilm.engine != nil {
		ilm.engine.RecordInteraction(
			step.InteractionType,
			fmt.Sprintf("Completed step: %s", step.Title),
			map[string]interface{}{
				"module_id": session.ModuleID,
				"step_id":   stepID,
				"score":     progress.Score,
				"topic":     session.Module.Category,
			},
			timeSpent,
			progress.Score >= 0.7, // Consider success if score >= 70%
		)
	}

	ilm.logger.Debug(fmt.Sprintf("Step completed: %s, Score: %.2f", stepID, progress.Score))
	return nil
}

func (ilm *InteractiveLearningModule) calculateStepScore(step LearningStep, userResponse interface{}, timeSpent time.Duration) float64 {
	baseScore := 1.0

	// Adjust for time efficiency
	if timeSpent > 0 && step.ExpectedTime > 0 {
		timeRatio := float64(timeSpent) / float64(step.ExpectedTime)
		if timeRatio < 0.5 {
			baseScore *= 1.2 // Bonus for quick completion
		} else if timeRatio > 2.0 {
			baseScore *= 0.8 // Penalty for slow completion
		}
	}

	// Step type specific scoring
	switch step.Type {
	case StepPractice:
		// For practice steps, score based on correctness
		if userResponse != nil {
			// This would need more sophisticated validation logic
			baseScore *= 0.9 // Assume mostly correct for now
		}
	case StepAssessment:
		// Assessment scoring handled separately
		baseScore = 1.0
	default:
		// Content steps get full score for completion
		baseScore = 1.0
	}

	// Ensure score is between 0 and 1
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	if baseScore < 0.0 {
		baseScore = 0.0
	}

	return baseScore
}

func (ilm *InteractiveLearningModule) calculateSessionScore(session *ModuleSession) float64 {
	if len(session.StepProgress) == 0 {
		return 0.0
	}

	totalScore := 0.0
	totalSteps := 0

	for _, progress := range session.StepProgress {
		if progress.Status == StepCompleted {
			totalScore += progress.Score
			totalSteps++
		}
	}

	if totalSteps == 0 {
		return 0.0
	}

	return totalScore / float64(totalSteps)
}

// CompleteModule marks a module as completed
func (ilm *InteractiveLearningModule) CompleteModule(session *ModuleSession) error {
	endTime := time.Now()
	session.EndTime = &endTime
	session.Status = SessionCompleted

	// Update user competency based on module completion
	if ilm.engine != nil {
		ilm.updateCompetencyFromModule(session)

		// Check for achievements
		achievements := ilm.checkForAchievements(session)
		for _, achievement := range achievements {
			session.Notes = append(session.Notes, SessionNote{
				Timestamp: time.Now(),
				Type:      "achievement",
				Content:   fmt.Sprintf("Achievement unlocked: %s", achievement.Title),
			})
		}
	}

	ilm.logger.Info(fmt.Sprintf("Module completed: %s, Final Score: %.2f", session.ModuleID, session.Score))
	return nil
}

func (ilm *InteractiveLearningModule) updateCompetencyFromModule(session *ModuleSession) {
	category := session.Module.Category
	currentLevel := ilm.engine.getTopicCompetency(category)

	// Calculate competency increase based on performance
	scoreBonus := session.Score // 0.0 to 1.0
	timeBonus := 0.0

	// Bonus for completing in reasonable time
	expectedTime := session.Module.EstimatedTime
	if session.TimeSpent <= expectedTime {
		timeBonus = 0.1
	}

	totalBonus := scoreBonus + timeBonus
	competencyIncrease := totalBonus * 0.3 // Scale to reasonable increase

	newLevel := ilm.engine.adjustCompetencyLevel(currentLevel, competencyIncrease)
	ilm.engine.userProfile.CompetencyMap[category] = newLevel

	// Update overall skill level if appropriate
	ilm.updateOverallSkillLevel()
}

func (ilm *InteractiveLearningModule) updateOverallSkillLevel() {
	profile := ilm.engine.userProfile

	// Calculate average competency
	totalCompetency := 0.0
	competencyCount := 0

	for _, level := range profile.CompetencyMap {
		totalCompetency += ilm.competencyToNumber(level)
		competencyCount++
	}

	if competencyCount == 0 {
		return
	}

	avgCompetency := totalCompetency / float64(competencyCount)
	newSkillLevel := ilm.numberToSkillLevel(avgCompetency)

	if newSkillLevel != profile.SkillLevel {
		profile.SkillLevel = newSkillLevel
		ilm.logger.Info(fmt.Sprintf("User skill level updated to: %s", newSkillLevel))
	}
}

func (ilm *InteractiveLearningModule) competencyToNumber(level CompetencyLevel) float64 {
	switch level {
	case CompetencyNone:
		return 0.0
	case CompetencyBeginner:
		return 1.0
	case CompetencyBasic:
		return 2.0
	case CompetencyIntermediate:
		return 3.0
	case CompetencyAdvanced:
		return 4.0
	case CompetencyExpert:
		return 5.0
	case CompetencyMastery:
		return 6.0
	default:
		return 0.0
	}
}

func (ilm *InteractiveLearningModule) numberToSkillLevel(num float64) SkillLevel {
	switch {
	case num < 1.5:
		return SkillBeginner
	case num < 3.5:
		return SkillIntermediate
	case num < 5.0:
		return SkillAdvanced
	default:
		return SkillExpert
	}
}

func (ilm *InteractiveLearningModule) checkForAchievements(session *ModuleSession) []Achievement {
	achievements := []Achievement{}

	// First module completion
	if session.Module.ID == "nix-basics" {
		achievements = append(achievements, Achievement{
			ID:          "first-steps",
			Title:       "First Steps",
			Description: "Completed your first Nix learning module",
			EarnedAt:    time.Now(),
			Points:      100,
			Rarity:      RarityCommon,
		})
	}

	// Perfect score achievement
	if session.Score >= 0.95 {
		achievements = append(achievements, Achievement{
			ID:          "perfectionist",
			Title:       "Perfectionist",
			Description: "Achieved near-perfect score in a learning module",
			EarnedAt:    time.Now(),
			Points:      200,
			Rarity:      RarityUncommon,
		})
	}

	// Speed completion
	if session.TimeSpent <= session.Module.EstimatedTime/2 {
		achievements = append(achievements, Achievement{
			ID:          "speed-learner",
			Title:       "Speed Learner",
			Description: "Completed module in half the estimated time",
			EarnedAt:    time.Now(),
			Points:      150,
			Rarity:      RarityRare,
		})
	}

	// Check for module-specific achievements
	switch session.Module.ID {
	case "flakes":
		achievements = append(achievements, Achievement{
			ID:          "flake-master",
			Title:       "Flake Master",
			Description: "Mastered the modern Nix flakes system",
			EarnedAt:    time.Now(),
			Points:      300,
			Rarity:      RarityEpic,
		})
	case "advanced-topics":
		achievements = append(achievements, Achievement{
			ID:          "nix-guru",
			Title:       "Nix Guru",
			Description: "Completed advanced NixOS topics",
			EarnedAt:    time.Now(),
			Points:      500,
			Rarity:      RarityLegendary,
		})
	}

	return achievements
}

// GetModuleProgress returns detailed progress information for a module
func (ilm *InteractiveLearningModule) GetModuleProgress(moduleID string) (*ModuleProgressSummary, error) {
	// This would typically load from persistent storage
	// For now, return a basic summary

	summary := &ModuleProgressSummary{
		ModuleID:       moduleID,
		CompletionRate: 0.0,
		TotalSessions:  0,
		BestScore:      0.0,
		TotalTimeSpent: 0,
		LastAccessed:   nil,
		StepProgress:   make(map[string]StepProgressSummary),
	}

	return summary, nil
}

type ModuleProgressSummary struct {
	ModuleID       string                         `json:"module_id"`
	CompletionRate float64                        `json:"completion_rate"`
	TotalSessions  int                            `json:"total_sessions"`
	BestScore      float64                        `json:"best_score"`
	TotalTimeSpent time.Duration                  `json:"total_time_spent"`
	LastAccessed   *time.Time                     `json:"last_accessed,omitempty"`
	StepProgress   map[string]StepProgressSummary `json:"step_progress"`
	Achievements   []Achievement                  `json:"achievements"`
}

type StepProgressSummary struct {
	StepID         string        `json:"step_id"`
	CompletionRate float64       `json:"completion_rate"`
	BestScore      float64       `json:"best_score"`
	TotalAttempts  int           `json:"total_attempts"`
	AverageTime    time.Duration `json:"average_time"`
}

// GetRecommendedModules returns modules recommended for the user
func (ilm *InteractiveLearningModule) GetRecommendedModules() ([]LearningModule, error) {
	allModules, err := ilm.GetAvailableModules()
	if err != nil {
		return nil, err
	}

	// Get learning recommendations from the adaptive engine
	if ilm.engine != nil {
		recommendations := ilm.engine.GetLearningRecommendations()

		// Filter modules based on recommendations
		recommendedModules := []LearningModule{}
		for _, rec := range recommendations {
			for _, module := range allModules {
				if strings.Contains(module.Category, rec.Topic) ||
					strings.Contains(module.Title, rec.Topic) {
					recommendedModules = append(recommendedModules, module)
					break
				}
			}
		}

		// Limit to top 5 recommendations
		if len(recommendedModules) > 5 {
			recommendedModules = recommendedModules[:5]
		}

		return recommendedModules, nil
	}

	// Default recommendations if no engine
	if len(allModules) > 3 {
		return allModules[:3], nil
	}
	return allModules, nil
}
