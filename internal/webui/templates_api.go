package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"nix-ai-help/pkg/logger"

	"github.com/gorilla/mux"
)

// TemplateAPI provides API endpoints for template management
type TemplateAPI struct {
	logger *logger.Logger
}

// Template represents a configuration template for the web API
type Template struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	Source      string            `json:"source"`
	Content     string            `json:"content"`
	Metadata    map[string]string `json:"metadata"`
}

// NewTemplateAPI creates a new template API handler
func NewTemplateAPI(logger *logger.Logger) *TemplateAPI {
	return &TemplateAPI{
		logger: logger,
	}
}

// RegisterRoutes registers all template API routes
func (api *TemplateAPI) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/templates", api.handleListTemplates).Methods("GET", "HEAD", "OPTIONS")
	router.HandleFunc("/api/templates/categories", api.handleTemplateCategories).Methods("GET", "HEAD", "OPTIONS")
	router.HandleFunc("/api/templates/{templateName}", api.handleGetTemplate).Methods("GET", "HEAD", "OPTIONS")
	router.HandleFunc("/api/templates/{templateName}/apply", api.handleApplyTemplate).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/templates/search", api.handleSearchTemplates).Methods("GET", "HEAD", "OPTIONS")
}

// handleListTemplates returns all available templates
func (api *TemplateAPI) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	api.setJSONHeaders(w)

	// Built-in templates
	templates := api.getBuiltinTemplates()

	// Add category filter if specified
	category := r.URL.Query().Get("category")
	if category != "" {
		var filtered []Template
		for _, template := range templates {
			if strings.EqualFold(template.Category, category) {
				filtered = append(filtered, template)
			}
		}
		templates = filtered
	}

	api.sendJSON(w, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// handleTemplateCategories returns available template categories
func (api *TemplateAPI) handleTemplateCategories(w http.ResponseWriter, r *http.Request) {
	api.setJSONHeaders(w)

	categories := map[string]interface{}{
		"Desktop": map[string]interface{}{
			"name":        "Desktop",
			"description": "Desktop environment configurations",
			"count":       1,
		},
		"Server": map[string]interface{}{
			"name":        "Server",
			"description": "Server and headless configurations",
			"count":       1,
		},
		"Development": map[string]interface{}{
			"name":        "Development",
			"description": "Development environment setups",
			"count":       1,
		},
	}

	api.sendJSON(w, categories)
}

// handleGetTemplate returns a specific template
func (api *TemplateAPI) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	api.setJSONHeaders(w)

	vars := mux.Vars(r)
	templateName := vars["templateName"]

	templates := api.getBuiltinTemplates()
	for _, template := range templates {
		if template.Name == templateName {
			api.sendJSON(w, template)
			return
		}
	}

	api.sendError(w, http.StatusNotFound, "Template not found")
}

// handleApplyTemplate applies a template to the canvas
func (api *TemplateAPI) handleApplyTemplate(w http.ResponseWriter, r *http.Request) {
	api.setJSONHeaders(w)

	vars := mux.Vars(r)
	templateName := vars["templateName"]

	var request struct {
		ClearCanvas bool `json:"clear_canvas"`
	}

	if err := api.decodeJSON(r, &request); err != nil {
		api.sendError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	templates := api.getBuiltinTemplates()
	var selectedTemplate *Template
	for _, template := range templates {
		if template.Name == templateName {
			selectedTemplate = &template
			break
		}
	}

	if selectedTemplate == nil {
		api.sendError(w, http.StatusNotFound, "Template not found")
		return
	}

	// For now, return the template configuration for the frontend to handle
	response := map[string]interface{}{
		"success":      true,
		"template":     selectedTemplate,
		"clear_canvas": request.ClearCanvas,
		"message":      fmt.Sprintf("Template '%s' ready to apply", templateName),
	}

	api.sendJSON(w, response)
}

// handleSearchTemplates searches templates by query
func (api *TemplateAPI) handleSearchTemplates(w http.ResponseWriter, r *http.Request) {
	api.setJSONHeaders(w)

	query := r.URL.Query().Get("q")
	if query == "" {
		api.sendError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	templates := api.getBuiltinTemplates()
	var matches []Template

	query = strings.ToLower(query)
	for _, template := range templates {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(template.Name), query) ||
			strings.Contains(strings.ToLower(template.Description), query) ||
			strings.Contains(strings.ToLower(template.Category), query) {
			matches = append(matches, template)
			continue
		}

		// Search in tags
		for _, tag := range template.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, template)
				break
			}
		}
	}

	api.sendJSON(w, map[string]interface{}{
		"templates": matches,
		"count":     len(matches),
		"query":     query,
	})
}

// getBuiltinTemplates returns the built-in templates
func (api *TemplateAPI) getBuiltinTemplates() []Template {
	return []Template{
		{
			Name:        "desktop",
			Description: "Basic desktop environment configuration with GNOME",
			Category:    "Desktop",
			Tags:        []string{"desktop", "gnome", "basic"},
			Source:      "builtin",
			Content: `{ config, pkgs, ... }:

{
  # Enable the X11 windowing system
  services.xserver.enable = true;

  # Enable the GNOME Desktop Environment
  services.xserver.displayManager.gdm.enable = true;
  services.xserver.desktopManager.gnome.enable = true;

  # Configure keymap in X11
  services.xserver.xkb = {
    layout = "us";
    variant = "";
  };

  # Enable CUPS to print documents
  services.printing.enable = true;

  # Enable sound with pipewire
  sound.enable = true;
  hardware.pulseaudio.enable = false;
  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    alsa.support32Bit = true;
    pulse.enable = true;
  };

  # Define a user account
  users.users.alice = {
    isNormalUser = true;
    description = "Alice";
    extraGroups = [ "networkmanager" "wheel" ];
    packages = with pkgs; [
      firefox
      tree
    ];
  };

  # Install firefox
  programs.firefox.enable = true;

  # Allow unfree packages
  nixpkgs.config.allowUnfree = true;

  # List packages installed in system profile
  environment.systemPackages = with pkgs; [
    vim
    wget
    git
  ];

  # This value determines the NixOS release
  system.stateVersion = "23.11";
}`,
			Metadata: map[string]string{
				"type":        "desktop",
				"complexity":  "basic",
				"environment": "gnome",
			},
		},
		{
			Name:        "server",
			Description: "Basic server configuration for headless systems",
			Category:    "Server",
			Tags:        []string{"server", "headless", "basic"},
			Source:      "builtin",
			Content: `{ config, pkgs, ... }:

{
  # Boot loader
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Networking
  networking.hostName = "nixos-server";
  networking.networkmanager.enable = true;

  # Set your time zone
  time.timeZone = "Europe/London";

  # Internationalization properties
  i18n.defaultLocale = "en_GB.UTF-8";

  # Configure console keymap
  console.keyMap = "uk";

  # Enable SSH
  services.openssh = {
    enable = true;
    settings = {
      PasswordAuthentication = false;
      KbdInteractiveAuthentication = false;
    };
  };

  # Enable firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 80 443 ];
  };

  # Define a user account
  users.users.admin = {
    isNormalUser = true;
    description = "Admin User";
    extraGroups = [ "networkmanager" "wheel" ];
    openssh.authorizedKeys.keys = [
      # Add your SSH public keys here
    ];
  };

  # Install essential packages
  environment.systemPackages = with pkgs; [
    vim
    wget
    curl
    git
    htop
    tmux
    rsync
  ];

  # Automatic garbage collection
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 30d";
  };

  # This value determines the NixOS release
  system.stateVersion = "23.11";
}`,
			Metadata: map[string]string{
				"type":        "server",
				"complexity":  "basic",
				"environment": "headless",
			},
		},
		{
			Name:        "development",
			Description: "Development environment with common programming tools",
			Category:    "Development",
			Tags:        []string{"development", "programming", "tools"},
			Source:      "builtin",
			Content: `{ config, pkgs, ... }:

{
  # Enable the X11 windowing system
  services.xserver.enable = true;

  # Enable the GNOME Desktop Environment
  services.xserver.displayManager.gdm.enable = true;
  services.xserver.desktopManager.gnome.enable = true;

  # Enable sound with pipewire
  sound.enable = true;
  hardware.pulseaudio.enable = false;
  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    alsa.support32Bit = true;
    pulse.enable = true;
  };

  # Define a user account
  users.users.developer = {
    isNormalUser = true;
    description = "Developer";
    extraGroups = [ "networkmanager" "wheel" "docker" ];
  };

  # Enable Docker
  virtualisation.docker.enable = true;

  # Allow unfree packages
  nixpkgs.config.allowUnfree = true;

  # Development tools
  environment.systemPackages = with pkgs; [
    # Editors
    vim
    neovim
    vscode

    # Version control
    git
    gh

    # Programming languages
    nodejs
    python3
    go
    rustc
    cargo

    # Development tools
    docker
    docker-compose
    kubectl
    terraform

    # System tools
    wget
    curl
    htop
    tree
    jq
    ripgrep
    fd

    # Network tools
    netcat
    nmap
    wireshark
  ];

  # Enable common development programs
  programs = {
    firefox.enable = true;
    git = {
      enable = true;
      config = {
        init.defaultBranch = "main";
      };
    };
  };

  # Development-friendly shell
  programs.zsh.enable = true;
  users.defaultUserShell = pkgs.zsh;

  # Enable SSH
  services.openssh.enable = true;

  # This value determines the NixOS release
  system.stateVersion = "23.11";
}`,
			Metadata: map[string]string{
				"type":        "development",
				"complexity":  "intermediate",
				"environment": "desktop",
			},
		},
	}
}

// Helper methods

func (api *TemplateAPI) setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (api *TemplateAPI) sendJSON(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.Error(fmt.Sprintf("Failed to encode JSON response: %v", err))
	}
}

func (api *TemplateAPI) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"error":   true,
		"message": message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.logger.Error(fmt.Sprintf("Failed to encode error response: %v", err))
	}
}

func (api *TemplateAPI) decodeJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
