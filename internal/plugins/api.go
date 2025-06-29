package plugins

import (
	"context"
	"encoding/json"
	"time"
)

// PluginInterface defines the basic interface all plugins must implement
type PluginInterface interface {
	// Metadata
	Name() string
	Version() string
	Description() string
	Author() string
	Repository() string
	License() string
	Dependencies() []string
	Capabilities() []string

	// Lifecycle
	Initialize(ctx context.Context, config PluginConfig) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Cleanup(ctx context.Context) error
	IsRunning() bool

	// Execution
	Execute(ctx context.Context, operation string, params map[string]interface{}) (interface{}, error)
	GetOperations() []PluginOperation
	GetSchema(operation string) (*PluginSchema, error)

	// Health and Status
	HealthCheck(ctx context.Context) PluginHealth
	GetMetrics() PluginMetrics
	GetStatus() PluginStatus
}

// PluginConfig represents configuration for a plugin
type PluginConfig struct {
	Name           string                 `json:"name" yaml:"name"`
	Enabled        bool                   `json:"enabled" yaml:"enabled"`
	Version        string                 `json:"version" yaml:"version"`
	Configuration  map[string]interface{} `json:"configuration" yaml:"configuration"`
	Environment    map[string]string      `json:"environment" yaml:"environment"`
	Resources      ResourceLimits         `json:"resources" yaml:"resources"`
	SecurityPolicy SecurityPolicy         `json:"security_policy" yaml:"security_policy"`
	UpdatePolicy   UpdatePolicy           `json:"update_policy" yaml:"update_policy"`
}

// ResourceLimits defines resource constraints for plugins
type ResourceLimits struct {
	MaxMemoryMB      int           `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent    int           `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	MaxExecutionTime time.Duration `json:"max_execution_time" yaml:"max_execution_time"`
	MaxFileSize      int64         `json:"max_file_size" yaml:"max_file_size"`
	AllowedPaths     []string      `json:"allowed_paths" yaml:"allowed_paths"`
	NetworkAccess    bool          `json:"network_access" yaml:"network_access"`
}

// SecurityPolicy defines security constraints for plugins
type SecurityPolicy struct {
	AllowFileSystem     bool         `json:"allow_file_system" yaml:"allow_file_system"`
	AllowNetwork        bool         `json:"allow_network" yaml:"allow_network"`
	AllowSystemCalls    bool         `json:"allow_system_calls" yaml:"allow_system_calls"`
	AllowedDomains      []string     `json:"allowed_domains" yaml:"allowed_domains"`
	RequiredPermissions []string     `json:"required_permissions" yaml:"required_permissions"`
	SandboxLevel        SandboxLevel `json:"sandbox_level" yaml:"sandbox_level"`
}

// UpdatePolicy defines how plugins should be updated
type UpdatePolicy struct {
	AutoUpdate         bool          `json:"auto_update" yaml:"auto_update"`
	UpdateChannel      string        `json:"update_channel" yaml:"update_channel"` // stable, beta, dev
	CheckInterval      time.Duration `json:"check_interval" yaml:"check_interval"`
	RequireApproval    bool          `json:"require_approval" yaml:"require_approval"`
	BackupBeforeUpdate bool          `json:"backup_before_update" yaml:"backup_before_update"`
}

// SandboxLevel defines the level of sandboxing for plugin execution
type SandboxLevel int

const (
	SandboxNone SandboxLevel = iota
	SandboxBasic
	SandboxStrict
	SandboxIsolated
)

// PluginOperation represents an operation that a plugin can perform
type PluginOperation struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  []PluginParameter  `json:"parameters"`
	ReturnType  string             `json:"return_type"`
	Examples    []OperationExample `json:"examples"`
	Tags        []string           `json:"tags"`
	Deprecated  bool               `json:"deprecated"`
}

// PluginParameter represents a parameter for a plugin operation
type PluginParameter struct {
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Default     interface{}          `json:"default,omitempty"`
	Validation  *ParameterValidation `json:"validation,omitempty"`
}

// ParameterValidation defines validation rules for parameters
type ParameterValidation struct {
	MinLength int      `json:"min_length,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
}

// OperationExample provides usage examples for operations
type OperationExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Expected    interface{}            `json:"expected"`
}

// PluginSchema represents the JSON schema for plugin operations
type PluginSchema struct {
	Type       string                          `json:"type"`
	Properties map[string]PluginSchemaProperty `json:"properties"`
	Required   []string                        `json:"required"`
	Examples   []map[string]interface{}        `json:"examples"`
}

// PluginSchemaProperty represents a property in the plugin schema
type PluginSchemaProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
}

// PluginHealth represents the health status of a plugin
type PluginHealth struct {
	Status          HealthStatus  `json:"status"`
	Message         string        `json:"message"`
	LastCheck       time.Time     `json:"last_check"`
	Uptime          time.Duration `json:"uptime"`
	Issues          []HealthIssue `json:"issues"`
	Recommendations []string      `json:"recommendations"`
}

// HealthStatus represents the health status level
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthHealthy
	HealthDegraded
	HealthUnhealthy
	HealthCritical
)

// HealthIssue represents a specific health issue
type HealthIssue struct {
	Severity   IssueSeverity `json:"severity"`
	Component  string        `json:"component"`
	Message    string        `json:"message"`
	Timestamp  time.Time     `json:"timestamp"`
	Resolution string        `json:"resolution,omitempty"`
}

// IssueSeverity represents the severity of a health issue
type IssueSeverity int

const (
	SeverityInfo IssueSeverity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// PluginMetrics represents performance metrics for a plugin
type PluginMetrics struct {
	ExecutionCount       int64                  `json:"execution_count"`
	TotalExecutionTime   time.Duration          `json:"total_execution_time"`
	AverageExecutionTime time.Duration          `json:"average_execution_time"`
	LastExecutionTime    time.Time              `json:"last_execution_time"`
	ErrorCount           int64                  `json:"error_count"`
	SuccessRate          float64                `json:"success_rate"`
	MemoryUsage          int64                  `json:"memory_usage_bytes"`
	CPUUsage             float64                `json:"cpu_usage_percent"`
	StartTime            time.Time              `json:"start_time"`
	CustomMetrics        map[string]interface{} `json:"custom_metrics"`
}

// PluginStatus represents the current status of a plugin
type PluginStatus struct {
	State       PluginState `json:"state"`
	Message     string      `json:"message"`
	LastUpdated time.Time   `json:"last_updated"`
	Version     string      `json:"version"`
	ConfigHash  string      `json:"config_hash"`
	ProcessID   int         `json:"process_id,omitempty"`
}

// PluginState represents the state of a plugin
type PluginState int

const (
	StateUnknown PluginState = iota
	StateLoading
	StateInitializing
	StateRunning
	StateStopping
	StateStopped
	StateError
	StateDisabled
)

// String methods for enums
func (s SandboxLevel) String() string {
	switch s {
	case SandboxNone:
		return "none"
	case SandboxBasic:
		return "basic"
	case SandboxStrict:
		return "strict"
	case SandboxIsolated:
		return "isolated"
	default:
		return "unknown"
	}
}

func (h HealthStatus) String() string {
	switch h {
	case HealthHealthy:
		return "healthy"
	case HealthDegraded:
		return "degraded"
	case HealthUnhealthy:
		return "unhealthy"
	case HealthCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (s IssueSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func (s PluginState) String() string {
	switch s {
	case StateLoading:
		return "loading"
	case StateInitializing:
		return "initializing"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	case StateDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// PluginEvent represents events that can be emitted by plugins
type PluginEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Tags      []string               `json:"tags"`
}

// EventHandler defines a function type for handling plugin events
type EventHandler func(ctx context.Context, event PluginEvent) error

// PluginRegistry interface for plugin discovery and registration
type PluginRegistry interface {
	Register(plugin PluginInterface) error
	Unregister(name string) error
	Get(name string) (PluginInterface, bool)
	List() []PluginInterface
	ListByCapability(capability string) []PluginInterface
	Search(query string) []PluginInterface
	GetMetadata(name string) (*PluginMetadata, error)
}

// PluginMetadata contains detailed metadata about a plugin
type PluginMetadata struct {
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Description     string          `json:"description"`
	Author          string          `json:"author"`
	Repository      string          `json:"repository"`
	License         string          `json:"license"`
	Homepage        string          `json:"homepage"`
	Documentation   string          `json:"documentation"`
	Keywords        []string        `json:"keywords"`
	Categories      []string        `json:"categories"`
	Dependencies    []string        `json:"dependencies"`
	Capabilities    []string        `json:"capabilities"`
	MinNixaiVersion string          `json:"min_nixai_version"`
	SupportedOS     []string        `json:"supported_os"`
	Configuration   json.RawMessage `json:"configuration_schema"`
	Screenshots     []string        `json:"screenshots"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DownloadCount   int64           `json:"download_count"`
	Rating          float64         `json:"rating"`
	Reviews         int             `json:"reviews"`
	Verified        bool            `json:"verified"`
}

// PluginLoader interface for loading and unloading plugins
type PluginLoader interface {
	Load(path string) (PluginInterface, error)
	LoadFromSource(source []byte, name string) (PluginInterface, error)
	Unload(plugin PluginInterface) error
	Reload(plugin PluginInterface) error
	ValidatePlugin(path string) error
	GetLoadedPlugins() []string
}

// PluginManager interface for managing plugin lifecycle
type PluginManager interface {
	LoadPlugin(path string, config PluginConfig) error
	UnloadPlugin(name string) error
	StartPlugin(name string) error
	StopPlugin(name string) error
	RestartPlugin(name string) error
	GetPlugin(name string) (PluginInterface, bool)
	ListPlugins() []PluginInterface
	GetPluginStatus(name string) (*PluginStatus, error)
	GetPluginHealth(name string) (*PluginHealth, error)
	GetPluginMetrics(name string) (*PluginMetrics, error)
	UpdatePlugin(name string) error
	ConfigurePlugin(name string, config PluginConfig) error
	ExecutePluginOperation(name, operation string, params map[string]interface{}) (interface{}, error)
	SubscribeToEvents(handler EventHandler) error
	UnsubscribeFromEvents(handler EventHandler) error
}
