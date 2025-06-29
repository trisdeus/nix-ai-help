package config_builder

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// ComponentType represents the type of configuration component
type ComponentType string

const (
	ComponentService    ComponentType = "service"
	ComponentPackage    ComponentType = "package"
	ComponentModule     ComponentType = "module"
	ComponentOptionType ComponentType = "option"
	ComponentVariable   ComponentType = "variable"
)

// ComponentCategory represents the category of a component
type ComponentCategory string

const (
	CategorySystem      ComponentCategory = "system"
	CategoryNetwork     ComponentCategory = "network"
	CategorySecurity    ComponentCategory = "security"
	CategoryDevelopment ComponentCategory = "development"
	CategoryMedia       ComponentCategory = "media"
	CategoryGaming      ComponentCategory = "gaming"
	CategoryDatabase    ComponentCategory = "database"
	CategoryWebServer   ComponentCategory = "webserver"
	CategoryDesktop     ComponentCategory = "desktop"
	CategoryUtilities   ComponentCategory = "utilities"
)

// ComponentDifficulty represents the complexity level
type ComponentDifficulty string

const (
	DifficultyBeginner     ComponentDifficulty = "beginner"
	DifficultyIntermediate ComponentDifficulty = "intermediate"
	DifficultyAdvanced     ComponentDifficulty = "advanced"
	DifficultyExpert       ComponentDifficulty = "expert"
)

// ConfigurationComponent represents a visual component in the builder
type ConfigurationComponent struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Type        ComponentType       `json:"type"`
	Category    ComponentCategory   `json:"category"`
	Description string              `json:"description"`
	Icon        string              `json:"icon"`
	Color       string              `json:"color"`
	Difficulty  ComponentDifficulty `json:"difficulty"`

	// Configuration details
	NixExpression string   `json:"nix_expression"`
	Dependencies  []string `json:"dependencies"`
	ConflictsWith []string `json:"conflicts_with"`
	Requires      []string `json:"requires"`

	// Visual properties
	Position    Position     `json:"position"`
	Size        Size         `json:"size"`
	Connections []Connection `json:"connections"`

	// Configuration options
	Options       []ComponentOption `json:"options"`
	ExampleConfig string            `json:"example_config"`

	// Metadata
	Tags          []string  `json:"tags"`
	Version       string    `json:"version"`
	Maintainer    string    `json:"maintainer"`
	Documentation string    `json:"documentation"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Position represents the position of a component in the visual builder
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size represents the size of a component
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Connection represents a connection between components
type Connection struct {
	ID     string `json:"id"`
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
	Type   string `json:"type"` // dependency, conflict, requires
	Label  string `json:"label"`
	Color  string `json:"color"`
}

// ComponentOption represents a configurable option for a component
type ComponentOption struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // string, int, bool, list, object
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Validation   string      `json:"validation"`
	Options      []string    `json:"options"` // for enum types
}

// ComponentLibrary manages the collection of available components
type ComponentLibrary struct {
	components map[string]*ConfigurationComponent
	categories map[ComponentCategory][]*ConfigurationComponent
	logger     *logger.Logger
}

// NewComponentLibrary creates a new component library
func NewComponentLibrary(logger *logger.Logger) *ComponentLibrary {
	lib := &ComponentLibrary{
		components: make(map[string]*ConfigurationComponent),
		categories: make(map[ComponentCategory][]*ConfigurationComponent),
		logger:     logger,
	}

	lib.initializeBuiltinComponents()
	return lib
}

// GetComponent retrieves a component by ID
func (cl *ComponentLibrary) GetComponent(id string) (*ConfigurationComponent, error) {
	component, exists := cl.components[id]
	if !exists {
		return nil, fmt.Errorf("component %s not found", id)
	}
	return component, nil
}

// GetComponentsByCategory retrieves all components in a category
func (cl *ComponentLibrary) GetComponentsByCategory(category ComponentCategory) []*ConfigurationComponent {
	components, exists := cl.categories[category]
	if !exists {
		return []*ConfigurationComponent{}
	}
	return components
}

// GetAllComponents retrieves all components
func (cl *ComponentLibrary) GetAllComponents() []*ConfigurationComponent {
	components := make([]*ConfigurationComponent, 0, len(cl.components))
	for _, component := range cl.components {
		components = append(components, component)
	}

	// Sort by category and name
	sort.Slice(components, func(i, j int) bool {
		if components[i].Category == components[j].Category {
			return components[i].Name < components[j].Name
		}
		return string(components[i].Category) < string(components[j].Category)
	})

	return components
}

// SearchComponents searches for components by name, description, or tags
func (cl *ComponentLibrary) SearchComponents(query string) []*ConfigurationComponent {
	query = strings.ToLower(query)
	var results []*ConfigurationComponent

	for _, component := range cl.components {
		if cl.matchesQuery(component, query) {
			results = append(results, component)
		}
	}

	// Sort by relevance (exact name matches first)
	sort.Slice(results, func(i, j int) bool {
		iExact := strings.ToLower(results[i].Name) == query
		jExact := strings.ToLower(results[j].Name) == query
		if iExact && !jExact {
			return true
		}
		if !iExact && jExact {
			return false
		}
		return results[i].Name < results[j].Name
	})

	return results
}

// AddComponent adds a new component to the library
func (cl *ComponentLibrary) AddComponent(component *ConfigurationComponent) error {
	if component.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	if _, exists := cl.components[component.ID]; exists {
		return fmt.Errorf("component %s already exists", component.ID)
	}

	// Set timestamps
	now := time.Now()
	if component.CreatedAt.IsZero() {
		component.CreatedAt = now
	}
	component.UpdatedAt = now

	// Add to components map
	cl.components[component.ID] = component

	// Add to category
	cl.categories[component.Category] = append(cl.categories[component.Category], component)

	cl.logger.Debug(fmt.Sprintf("Added component %s to library", component.ID))
	return nil
}

// UpdateComponent updates an existing component
func (cl *ComponentLibrary) UpdateComponent(component *ConfigurationComponent) error {
	if component.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	existingComponent, exists := cl.components[component.ID]
	if !exists {
		return fmt.Errorf("component %s not found", component.ID)
	}

	// Remove from old category if category changed
	if existingComponent.Category != component.Category {
		cl.removeFromCategory(existingComponent)
		cl.categories[component.Category] = append(cl.categories[component.Category], component)
	}

	// Update timestamp
	component.UpdatedAt = time.Now()
	component.CreatedAt = existingComponent.CreatedAt // Preserve created time

	// Update component
	cl.components[component.ID] = component

	cl.logger.Debug(fmt.Sprintf("Updated component %s", component.ID))
	return nil
}

// RemoveComponent removes a component from the library
func (cl *ComponentLibrary) RemoveComponent(id string) error {
	component, exists := cl.components[id]
	if !exists {
		return fmt.Errorf("component %s not found", id)
	}

	// Remove from components map
	delete(cl.components, id)

	// Remove from category
	cl.removeFromCategory(component)

	cl.logger.Debug(fmt.Sprintf("Removed component %s from library", id))
	return nil
}

// ExportComponents exports all components as JSON
func (cl *ComponentLibrary) ExportComponents() ([]byte, error) {
	components := cl.GetAllComponents()
	return json.MarshalIndent(components, "", "  ")
}

// ImportComponents imports components from JSON
func (cl *ComponentLibrary) ImportComponents(data []byte) error {
	var components []*ConfigurationComponent
	if err := json.Unmarshal(data, &components); err != nil {
		return fmt.Errorf("failed to unmarshal components: %w", err)
	}

	for _, component := range components {
		if err := cl.AddComponent(component); err != nil {
			cl.logger.Warn(fmt.Sprintf("Failed to import component %s: %v", component.ID, err))
		}
	}

	return nil
}

// matchesQuery checks if a component matches the search query
func (cl *ComponentLibrary) matchesQuery(component *ConfigurationComponent, query string) bool {
	// Check name
	if strings.Contains(strings.ToLower(component.Name), query) {
		return true
	}

	// Check description
	if strings.Contains(strings.ToLower(component.Description), query) {
		return true
	}

	// Check tags
	for _, tag := range component.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Check category
	if strings.Contains(strings.ToLower(string(component.Category)), query) {
		return true
	}

	return false
}

// removeFromCategory removes a component from its category
func (cl *ComponentLibrary) removeFromCategory(component *ConfigurationComponent) {
	categoryComponents := cl.categories[component.Category]
	for i, comp := range categoryComponents {
		if comp.ID == component.ID {
			cl.categories[component.Category] = append(categoryComponents[:i], categoryComponents[i+1:]...)
			break
		}
	}
}

// initializeBuiltinComponents initializes the library with built-in NixOS components
func (cl *ComponentLibrary) initializeBuiltinComponents() {
	components := []*ConfigurationComponent{
		// System Services
		{
			ID:            "openssh",
			Name:          "OpenSSH Server",
			Type:          ComponentService,
			Category:      CategorySystem,
			Description:   "Secure Shell (SSH) server for remote access",
			Icon:          "🔐",
			Color:         "#4CAF50",
			Difficulty:    DifficultyBeginner,
			NixExpression: "services.openssh.enable = true;",
			Dependencies:  []string{},
			ConflictsWith: []string{},
			Requires:      []string{},
			Options: []ComponentOption{
				{
					Name:         "enable",
					Type:         "bool",
					Description:  "Enable OpenSSH server",
					DefaultValue: true,
					Required:     true,
				},
				{
					Name:         "passwordAuthentication",
					Type:         "bool",
					Description:  "Allow password authentication",
					DefaultValue: false,
					Required:     false,
				},
				{
					Name:         "permitRootLogin",
					Type:         "string",
					Description:  "Root login permission",
					DefaultValue: "no",
					Required:     false,
					Options:      []string{"yes", "no", "without-password", "prohibit-password"},
				},
			},
			ExampleConfig: `services.openssh = {
  enable = true;
  passwordAuthentication = false;
  permitRootLogin = "no";
};`,
			Tags:          []string{"ssh", "remote", "access", "security"},
			Version:       "1.0",
			Maintainer:    "NixOS Team",
			Documentation: "https://wiki.nixos.org/wiki/OpenSSH",
		},

		// Web Server
		{
			ID:            "nginx",
			Name:          "Nginx Web Server",
			Type:          ComponentService,
			Category:      CategoryWebServer,
			Description:   "High-performance HTTP server and reverse proxy",
			Icon:          "🌐",
			Color:         "#2196F3",
			Difficulty:    DifficultyIntermediate,
			NixExpression: "services.nginx.enable = true;",
			Dependencies:  []string{},
			ConflictsWith: []string{"apache", "lighttpd"},
			Requires:      []string{},
			Options: []ComponentOption{
				{
					Name:         "enable",
					Type:         "bool",
					Description:  "Enable Nginx server",
					DefaultValue: true,
					Required:     true,
				},
				{
					Name:         "user",
					Type:         "string",
					Description:  "User to run Nginx as",
					DefaultValue: "nginx",
					Required:     false,
				},
				{
					Name:         "group",
					Type:         "string",
					Description:  "Group to run Nginx as",
					DefaultValue: "nginx",
					Required:     false,
				},
			},
			ExampleConfig: `services.nginx = {
  enable = true;
  user = "nginx";
  group = "nginx";
  virtualHosts."example.com" = {
    enableACME = true;
    forceSSL = true;
    root = "/var/www/example.com";
  };
};`,
			Tags:          []string{"web", "server", "http", "proxy"},
			Version:       "1.0",
			Maintainer:    "NixOS Team",
			Documentation: "https://wiki.nixos.org/wiki/Nginx",
		},

		// Database
		{
			ID:            "postgresql",
			Name:          "PostgreSQL Database",
			Type:          ComponentService,
			Category:      CategoryDatabase,
			Description:   "Advanced open source relational database",
			Icon:          "🐘",
			Color:         "#336791",
			Difficulty:    DifficultyIntermediate,
			NixExpression: "services.postgresql.enable = true;",
			Dependencies:  []string{},
			ConflictsWith: []string{},
			Requires:      []string{},
			Options: []ComponentOption{
				{
					Name:         "enable",
					Type:         "bool",
					Description:  "Enable PostgreSQL server",
					DefaultValue: true,
					Required:     true,
				},
				{
					Name:         "package",
					Type:         "string",
					Description:  "PostgreSQL package version",
					DefaultValue: "pkgs.postgresql_15",
					Required:     false,
				},
				{
					Name:         "port",
					Type:         "int",
					Description:  "Port to listen on",
					DefaultValue: 5432,
					Required:     false,
				},
			},
			ExampleConfig: `services.postgresql = {
  enable = true;
  package = pkgs.postgresql_15;
  port = 5432;
  authentication = "local all all trust";
  initialDatabases = [
    { name = "myapp"; }
  ];
};`,
			Tags:          []string{"database", "sql", "postgres", "data"},
			Version:       "1.0",
			Maintainer:    "NixOS Team",
			Documentation: "https://wiki.nixos.org/wiki/PostgreSQL",
		},

		// Development Tools
		{
			ID:            "docker",
			Name:          "Docker Container Engine",
			Type:          ComponentService,
			Category:      CategoryDevelopment,
			Description:   "Platform for developing, shipping, and running applications",
			Icon:          "🐳",
			Color:         "#0db7ed",
			Difficulty:    DifficultyIntermediate,
			NixExpression: "virtualisation.docker.enable = true;",
			Dependencies:  []string{},
			ConflictsWith: []string{"podman"},
			Requires:      []string{},
			Options: []ComponentOption{
				{
					Name:         "enable",
					Type:         "bool",
					Description:  "Enable Docker daemon",
					DefaultValue: true,
					Required:     true,
				},
				{
					Name:         "enableOnBoot",
					Type:         "bool",
					Description:  "Enable Docker on boot",
					DefaultValue: true,
					Required:     false,
				},
				{
					Name:         "storageDriver",
					Type:         "string",
					Description:  "Storage driver to use",
					DefaultValue: "overlay2",
					Required:     false,
					Options:      []string{"overlay2", "devicemapper", "btrfs", "zfs"},
				},
			},
			ExampleConfig: `virtualisation.docker = {
  enable = true;
  enableOnBoot = true;
  storageDriver = "overlay2";
};`,
			Tags:          []string{"docker", "containers", "development", "virtualization"},
			Version:       "1.0",
			Maintainer:    "NixOS Team",
			Documentation: "https://wiki.nixos.org/wiki/Docker",
		},

		// Desktop Environment
		{
			ID:            "gnome",
			Name:          "GNOME Desktop Environment",
			Type:          ComponentService,
			Category:      CategoryDesktop,
			Description:   "Modern desktop environment for Linux",
			Icon:          "🖥️",
			Color:         "#4A90E2",
			Difficulty:    DifficultyBeginner,
			NixExpression: "services.xserver.desktopManager.gnome.enable = true;",
			Dependencies:  []string{"xserver"},
			ConflictsWith: []string{"kde", "xfce", "i3"},
			Requires:      []string{},
			Options: []ComponentOption{
				{
					Name:         "enable",
					Type:         "bool",
					Description:  "Enable GNOME desktop",
					DefaultValue: true,
					Required:     true,
				},
			},
			ExampleConfig: `services.xserver = {
  enable = true;
  desktopManager.gnome.enable = true;
  displayManager.gdm.enable = true;
};`,
			Tags:          []string{"gnome", "desktop", "gui", "x11"},
			Version:       "1.0",
			Maintainer:    "NixOS Team",
			Documentation: "https://wiki.nixos.org/wiki/GNOME",
		},
	}

	for _, component := range components {
		if err := cl.AddComponent(component); err != nil {
			cl.logger.Error(fmt.Sprintf("Failed to add builtin component %s: %v", component.ID, err))
		}
	}
}
