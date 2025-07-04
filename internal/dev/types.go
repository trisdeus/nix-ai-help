package dev

import (
	"context"
	"time"
)

// DevEnvironment represents a development environment configuration
type DevEnvironment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Language    string                 `json:"language"`
	Framework   string                 `json:"framework,omitempty"`
	Editor      string                 `json:"editor"`
	Containers  []string               `json:"containers,omitempty"`
	Services    []string               `json:"services,omitempty"`
	Template    string                 `json:"template"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      DevEnvironmentStatus   `json:"status"`
	Path        string                 `json:"path"`
	Dependencies []Dependency          `json:"dependencies"`
}

// DevEnvironmentStatus represents the status of a development environment
type DevEnvironmentStatus string

const (
	DevEnvironmentStatusCreating DevEnvironmentStatus = "creating"
	DevEnvironmentStatusReady    DevEnvironmentStatus = "ready"
	DevEnvironmentStatusFailed   DevEnvironmentStatus = "failed"
	DevEnvironmentStatusStopped  DevEnvironmentStatus = "stopped"
)

// Dependency represents a project dependency
type Dependency struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Source   string `json:"source"`
}

// DevTemplate represents a development environment template
type DevTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Language    string                 `json:"language"`
	Framework   string                 `json:"framework,omitempty"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Config      map[string]interface{} `json:"config"`
	Files       []TemplateFile         `json:"files"`
	Dependencies []Dependency          `json:"dependencies"`
	Commands    []string               `json:"commands"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Author      string                 `json:"author"`
	Version     string                 `json:"version"`
}

// TemplateFile represents a file in a development template
type TemplateFile struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Template bool   `json:"template"`
}

// ProjectSetup represents project setup configuration
type ProjectSetup struct {
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Language    string                 `json:"language"`
	Framework   string                 `json:"framework,omitempty"`
	Editor      string                 `json:"editor"`
	Containers  []string               `json:"containers,omitempty"`
	Services    []string               `json:"services,omitempty"`
	Template    string                 `json:"template"`
	Config      map[string]interface{} `json:"config"`
	AutoDetect  bool                   `json:"auto_detect"`
	Interactive bool                   `json:"interactive"`
}

// DependencyDetector represents automatic dependency detection
type DependencyDetector struct {
	Language     string   `json:"language"`
	Patterns     []string `json:"patterns"`
	Files        []string `json:"files"`
	Commands     []string `json:"commands"`
	Dependencies []string `json:"dependencies"`
}

// IDEIntegration represents IDE integration configuration
type IDEIntegration struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	ConfigFiles  []string               `json:"config_files"`
	Extensions   []string               `json:"extensions"`
	Settings     map[string]interface{} `json:"settings"`
	LaunchConfig map[string]interface{} `json:"launch_config"`
}

// CIPipeline represents CI/CD pipeline configuration
type CIPipeline struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Steps    []PipelineStep         `json:"steps"`
	Triggers []string               `json:"triggers"`
}

// PipelineStep represents a step in a CI/CD pipeline
type PipelineStep struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Commands []string          `json:"commands"`
	Env      map[string]string `json:"env"`
}

// DevEnvironmentManager interface for managing development environments
type DevEnvironmentManager interface {
	CreateEnvironment(ctx context.Context, setup *ProjectSetup) (*DevEnvironment, error)
	GetEnvironment(ctx context.Context, id string) (*DevEnvironment, error)
	ListEnvironments(ctx context.Context) ([]*DevEnvironment, error)
	UpdateEnvironment(ctx context.Context, env *DevEnvironment) error
	DeleteEnvironment(ctx context.Context, id string) error
	StartEnvironment(ctx context.Context, id string) error
	StopEnvironment(ctx context.Context, id string) error
}

// TemplateManager interface for managing development templates
type TemplateManager interface {
	GetTemplate(ctx context.Context, id string) (*DevTemplate, error)
	ListTemplates(ctx context.Context) ([]*DevTemplate, error)
	CreateTemplate(ctx context.Context, template *DevTemplate) error
	UpdateTemplate(ctx context.Context, template *DevTemplate) error
	DeleteTemplate(ctx context.Context, id string) error
	SearchTemplates(ctx context.Context, query string) ([]*DevTemplate, error)
}

// DependencyManager interface for managing project dependencies
type DependencyManager interface {
	DetectDependencies(ctx context.Context, path string) ([]Dependency, error)
	InstallDependencies(ctx context.Context, path string, deps []Dependency) error
	UpdateDependencies(ctx context.Context, path string) error
	CheckDependencies(ctx context.Context, path string) ([]Dependency, error)
}

// IDEManager interface for managing IDE integration
type IDEManager interface {
	SetupIDE(ctx context.Context, env *DevEnvironment, ide string) error
	GetIDEConfig(ctx context.Context, ide string) (*IDEIntegration, error)
	ListSupportedIDEs(ctx context.Context) ([]string, error)
	InstallExtensions(ctx context.Context, ide string, extensions []string) error
}

// PipelineManager interface for managing CI/CD pipelines
type PipelineManager interface {
	CreatePipeline(ctx context.Context, env *DevEnvironment, pipelineType string) (*CIPipeline, error)
	GetPipeline(ctx context.Context, id string) (*CIPipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *CIPipeline) error
	DeletePipeline(ctx context.Context, id string) error
	RunPipeline(ctx context.Context, id string) error
}

// DevExperienceManager orchestrates all development experience components
type DevExperienceManager struct {
	EnvManager        DevEnvironmentManager
	TemplateManager   TemplateManager
	DependencyManager DependencyManager
	IDEManager        IDEManager
	PipelineManager   PipelineManager
}

// NewDevExperienceManager creates a new development experience manager
func NewDevExperienceManager(
	envManager DevEnvironmentManager,
	templateManager TemplateManager,
	dependencyManager DependencyManager,
	ideManager IDEManager,
	pipelineManager PipelineManager,
) *DevExperienceManager {
	return &DevExperienceManager{
		EnvManager:        envManager,
		TemplateManager:   templateManager,
		DependencyManager: dependencyManager,
		IDEManager:        ideManager,
		PipelineManager:   pipelineManager,
	}
}