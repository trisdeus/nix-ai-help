package codegen

import (
	"fmt"
	"strings"
	"text/template"

	"nix-ai-help/internal/config"
	"nix-ai-help/pkg/logger"
)

// Template represents a configuration template
type Template struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Category     string        `json:"category"`
	Complexity   string        `json:"complexity"`
	Content      string        `json:"content"`
	Variables    []TemplateVar `json:"variables"`
	Dependencies []string      `json:"dependencies"`
	Warnings     []string      `json:"warnings"`
	Suggestions  []string      `json:"suggestions"`
	HomeManager  bool          `json:"home_manager"`
}

// TemplateVar represents a template variable
type TemplateVar struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // "string", "boolean", "number", "list"
	Default     string   `json:"default"`
	Required    bool     `json:"required"`
	Options     []string `json:"options,omitempty"` // For enum-like variables
}

// TemplateLoader handles loading and applying configuration templates
type TemplateLoader struct {
	logger    logger.Logger
	templates map[string]*Template
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(logger logger.Logger) *TemplateLoader {
	tl := &TemplateLoader{
		logger:    logger,
		templates: make(map[string]*Template),
	}
	tl.initializeTemplates()
	return tl
}

// LoadTemplate loads a template by type and complexity
func (tl *TemplateLoader) LoadTemplate(templateType, complexity string) (*Template, error) {
	key := fmt.Sprintf("%s-%s", templateType, complexity)

	template, exists := tl.templates[key]
	if !exists {
		// Try fallback to basic complexity
		key = fmt.Sprintf("%s-basic", templateType)
		template, exists = tl.templates[key]
		if !exists {
			return nil, fmt.Errorf("template not found: %s", templateType)
		}
	}

	return template, nil
}

// ApplyTemplate applies a template with the given intent and context
func (tl *TemplateLoader) ApplyTemplate(tmpl *Template, intent Intent, context *config.NixOSContext) (string, error) {
	tl.logger.Info(fmt.Sprintf("Applying template - name: %s, components: %v", tmpl.Name, intent.Components))

	// Create template data
	data := map[string]interface{}{
		"Components":  intent.Components,
		"Options":     intent.Options,
		"Environment": intent.Environment,
		"Complexity":  intent.Complexity,
		"HomeManager": intent.HomeManager,
	}

	// Add context information if available
	if context != nil {
		data["Context"] = map[string]interface{}{
			"UsesFlakes":     context.UsesFlakes,
			"HasHomeManager": context.HasHomeManager,
			"NixOSVersion":   context.NixOSVersion,
			"SystemType":     context.SystemType,
		}
	}

	// Add component-specific variables
	data["HasComponent"] = func(component string) bool {
		return contains(intent.Components, component)
	}

	data["GetOption"] = func(key, defaultValue string) string {
		if value, exists := intent.Options[key]; exists {
			return value
		}
		return defaultValue
	}

	// Parse and execute template
	t, err := template.New(tmpl.Name).Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var result strings.Builder
	if err := t.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return result.String(), nil
}

// ListTemplates returns all available templates
func (tl *TemplateLoader) ListTemplates() map[string]*Template {
	return tl.templates
}

// GetTemplatesByCategory returns templates filtered by category
func (tl *TemplateLoader) GetTemplatesByCategory(category string) map[string]*Template {
	filtered := make(map[string]*Template)
	for key, tmpl := range tl.templates {
		if tmpl.Category == category {
			filtered[key] = tmpl
		}
	}
	return filtered
}

// initializeTemplates sets up built-in templates
func (tl *TemplateLoader) initializeTemplates() {
	// Desktop Environment Templates
	tl.templates["desktop-basic"] = &Template{
		Name:        "desktop-basic",
		Description: "Basic desktop environment configuration",
		Category:    "desktop",
		Complexity:  "basic",
		Content:     desktopBasicTemplate,
		Variables: []TemplateVar{
			{Name: "desktop_environment", Description: "Desktop environment", Type: "string", Default: "gnome", Options: []string{"gnome", "kde", "xfce", "i3"}},
			{Name: "enable_audio", Description: "Enable audio support", Type: "boolean", Default: "true"},
		},
		Dependencies: []string{"xorg", "desktop-environment"},
		Suggestions:  []string{"Consider adding development tools", "Enable bluetooth if needed"},
	}

	// Web Server Templates
	tl.templates["service-basic"] = &Template{
		Name:        "service-basic",
		Description: "Basic web server configuration",
		Category:    "service",
		Complexity:  "basic",
		Content:     webServerBasicTemplate,
		Variables: []TemplateVar{
			{Name: "server_name", Description: "Server name/domain", Type: "string", Required: true},
			{Name: "enable_ssl", Description: "Enable SSL/TLS", Type: "boolean", Default: "false"},
			{Name: "web_root", Description: "Web root directory", Type: "string", Default: "/var/www"},
		},
		Dependencies: []string{"nginx", "firewall"},
		Warnings:     []string{"Remember to configure DNS", "SSL requires valid certificates"},
		Suggestions:  []string{"Consider enabling fail2ban", "Set up log rotation"},
	}

	tl.templates["service-advanced"] = &Template{
		Name:        "service-advanced",
		Description: "Advanced web server with SSL and security",
		Category:    "service",
		Complexity:  "advanced",
		Content:     webServerAdvancedTemplate,
		Variables: []TemplateVar{
			{Name: "server_name", Description: "Server name/domain", Type: "string", Required: true},
			{Name: "enable_ssl", Description: "Enable SSL/TLS", Type: "boolean", Default: "true"},
			{Name: "web_root", Description: "Web root directory", Type: "string", Default: "/var/www"},
			{Name: "enable_cache", Description: "Enable caching", Type: "boolean", Default: "true"},
		},
		Dependencies: []string{"nginx", "firewall", "letsencrypt", "fail2ban"},
		Warnings:     []string{"Requires valid domain and DNS", "Monitor certificate renewal"},
		Suggestions:  []string{"Set up monitoring", "Configure backup strategy"},
	}

	// Development Environment Templates
	tl.templates["development-basic"] = &Template{
		Name:        "development-basic",
		Description: "Basic development environment",
		Category:    "development",
		Complexity:  "basic",
		Content:     developmentBasicTemplate,
		Variables: []TemplateVar{
			{Name: "languages", Description: "Programming languages", Type: "list", Default: "python,nodejs"},
			{Name: "enable_docker", Description: "Enable Docker", Type: "boolean", Default: "false"},
			{Name: "editor", Description: "Preferred editor", Type: "string", Default: "vscode", Options: []string{"vscode", "neovim", "emacs"}},
		},
		Dependencies: []string{"git", "development-tools"},
		Suggestions:  []string{"Consider using devenv for project isolation", "Set up shell aliases"},
	}

	// Home Manager Templates
	tl.templates["development-basic-home"] = &Template{
		Name:        "development-basic-home",
		Description: "Basic development environment for Home Manager",
		Category:    "development",
		Complexity:  "basic",
		Content:     homeManagerDevTemplate,
		HomeManager: true,
		Variables: []TemplateVar{
			{Name: "enable_zsh", Description: "Enable Zsh shell", Type: "boolean", Default: "true"},
			{Name: "enable_git", Description: "Enable Git configuration", Type: "boolean", Default: "true"},
		},
		Dependencies: []string{"home-manager"},
		Suggestions:  []string{"Configure shell aliases", "Set up dotfiles management"},
	}

	// Gaming Templates
	tl.templates["gaming-basic"] = &Template{
		Name:        "gaming-basic",
		Description: "Basic gaming setup with Steam",
		Category:    "gaming",
		Complexity:  "basic",
		Content:     gamingBasicTemplate,
		Variables: []TemplateVar{
			{Name: "enable_steam", Description: "Enable Steam", Type: "boolean", Default: "true"},
			{Name: "enable_lutris", Description: "Enable Lutris", Type: "boolean", Default: "false"},
			{Name: "gpu_driver", Description: "GPU driver", Type: "string", Options: []string{"nvidia", "amd", "intel"}},
		},
		Dependencies: []string{"steam", "graphics-drivers"},
		Warnings:     []string{"Requires unfree packages", "GPU drivers may need manual configuration"},
		Suggestions:  []string{"Enable gamemode for performance", "Consider ProtonGE for compatibility"},
	}

	// Security Templates
	tl.templates["security-basic"] = &Template{
		Name:         "security-basic",
		Description:  "Basic security hardening",
		Category:     "security",
		Complexity:   "basic",
		Content:      securityBasicTemplate,
		Dependencies: []string{"firewall", "fail2ban"},
		Suggestions:  []string{"Regular security updates", "Monitor system logs"},
	}
}

// Template content constants
const desktopBasicTemplate = `{ config, pkgs, ... }:

{
  # Desktop Environment Configuration
  services.xserver = {
    enable = true;
    {{if HasComponent "gnome"}}
    displayManager.gdm.enable = true;
    desktopManager.gnome.enable = true;
    {{else if HasComponent "kde"}}
    displayManager.sddm.enable = true;
    desktopManager.plasma5.enable = true;
    {{else if HasComponent "xfce"}}
    displayManager.lightdm.enable = true;
    desktopManager.xfce.enable = true;
    {{else if HasComponent "i3"}}
    displayManager.lightdm.enable = true;
    windowManager.i3.enable = true;
    {{else}}
    displayManager.gdm.enable = true;
    desktopManager.gnome.enable = true;
    {{end}}
  };

  # Audio support
  {{if ne (GetOption "enable_audio" "true") "false"}}
  sound.enable = true;
  hardware.pulseaudio.enable = true;
  # OR use pipewire (comment out pulseaudio above)
  # security.rtkit.enable = true;
  # services.pipewire = {
  #   enable = true;
  #   alsa.enable = true;
  #   alsa.support32Bit = true;
  #   pulse.enable = true;
  # };
  {{end}}

  # Essential desktop packages
  environment.systemPackages = with pkgs; [
    firefox
    thunderbird
    libreoffice
    {{if HasComponent "gnome"}}
    gnome.gnome-tweaks
    {{end}}
    {{if HasComponent "kde"}}
    kate
    {{end}}
    {{if HasComponent "development"}}
    git
    {{if HasComponent "vscode"}}
    vscode
    {{end}}
    {{if HasComponent "neovim"}}
    neovim
    {{end}}
    {{end}}
  ];

  # User configuration
  users.users.{{GetOption "username" "user"}} = {
    isNormalUser = true;
    extraGroups = [ "wheel" "networkmanager" "audio" ];
  };

  # Enable NetworkManager for easy network management
  networking.networkmanager.enable = true;

  # This value determines the NixOS release
  system.stateVersion = "25.05";
}`

const webServerBasicTemplate = `{ config, pkgs, ... }:

{
  # Nginx web server configuration
  services.nginx = {
    enable = true;
    {{if HasComponent "ssl"}}
    recommendedTlsSettings = true;
    recommendedOptimisation = true;
    recommendedGzipSettings = true;
    recommendedProxySettings = true;
    {{end}}

    virtualHosts."{{GetOption "domain" "localhost"}}" = {
      {{if HasComponent "ssl"}}
      forceSSL = true;
      enableACME = true;
      {{end}}
      root = "{{GetOption "root" "/var/www"}}";
      
      locations."/" = {
        tryFiles = "$uri $uri/ =404";
      };
    };
  };

  {{if HasComponent "ssl"}}
  # ACME/Let's Encrypt configuration
  security.acme = {
    acceptTerms = true;
    defaults.email = "{{GetOption "email" "admin@example.com"}}";
  };
  {{end}}

  # Firewall configuration
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 80 {{if HasComponent "ssl"}}443{{end}} ];
  };

  # User for web content
  users.users.nginx.extraGroups = [ "web" ];
  users.groups.web = {};

  # This value determines the NixOS release
  system.stateVersion = "25.05";
}`

const webServerAdvancedTemplate = `{ config, pkgs, ... }:

{
  # Advanced Nginx configuration with security
  services.nginx = {
    enable = true;
    recommendedTlsSettings = true;
    recommendedOptimisation = true;
    recommendedGzipSettings = true;
    recommendedProxySettings = true;

    # Security headers
    appendHttpConfig = ''
      add_header X-Frame-Options DENY;
      add_header X-Content-Type-Options nosniff;
      add_header X-XSS-Protection "1; mode=block";
      add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    '';

    virtualHosts."{{GetOption "domain" "example.com"}}" = {
      forceSSL = true;
      enableACME = true;
      root = "{{GetOption "root" "/var/www"}}";
      
      {{if ne (GetOption "enable_cache" "true") "false"}}
      # Caching configuration
      locations."~* \.(jpg|jpeg|png|gif|ico|css|js)$" = {
        extraConfig = ''
          expires 1M;
          add_header Cache-Control "public, immutable";
        '';
      };
      {{end}}

      locations."/" = {
        tryFiles = "$uri $uri/ =404";
      };
    };
  };

  # ACME/Let's Encrypt configuration
  security.acme = {
    acceptTerms = true;
    defaults.email = "{{GetOption "email" "admin@example.com"}}";
  };

  # Fail2ban for security
  services.fail2ban = {
    enable = true;
    jails = {
      nginx-http-auth.settings = {
        enabled = true;
        filter = "nginx-http-auth";
        logpath = "/var/log/nginx/error.log";
      };
    };
  };

  # Enhanced firewall
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 80 443 ];
    # Rate limiting (requires custom iptables rules)
  };

  # Log rotation
  services.logrotate.settings.nginx = {
    files = [ "/var/log/nginx/*.log" ];
    frequency = "daily";
    rotate = 30;
    compress = true;
    delaycompress = true;
    missingok = true;
    notifempty = true;
    postrotate = "systemctl reload nginx";
  };

  # This value determines the NixOS release
  system.stateVersion = "25.05";
}`

const developmentBasicTemplate = `{ config, pkgs, ... }:

{
  # Development environment configuration
  environment.systemPackages = with pkgs; [
    # Essential development tools
    git
    curl
    wget
    tree
    htop
    
    {{if HasComponent "python"}}
    # Python development
    python3
    python3Packages.pip
    python3Packages.virtualenv
    {{end}}
    
    {{if HasComponent "nodejs"}}
    # Node.js development
    nodejs
    npm
    yarn
    {{end}}
    
    {{if HasComponent "go"}}
    # Go development
    go
    {{end}}
    
    {{if HasComponent "docker"}}
    # Container tools
    docker
    docker-compose
    {{end}}
    
    {{if HasComponent "vscode"}}
    # VS Code
    vscode
    {{end}}
    
    {{if HasComponent "neovim"}}
    # Neovim
    neovim
    {{end}}
  ];

  {{if HasComponent "docker"}}
  # Docker configuration
  virtualisation.docker = {
    enable = true;
    enableOnBoot = true;
  };
  {{end}}

  # Git configuration (system-wide defaults)
  programs.git = {
    enable = true;
    config = {
      init.defaultBranch = "main";
      pull.rebase = true;
    };
  };

  # Enable shells
  programs.zsh.enable = true;
  programs.bash.completion.enable = true;

  # User configuration for development
  users.users.{{GetOption "username" "dev"}} = {
    isNormalUser = true;
    extraGroups = [ "wheel" {{if HasComponent "docker"}}"docker"{{end}} ];
    shell = pkgs.zsh;
  };

  # This value determines the NixOS release
  system.stateVersion = "25.05";
}`

const homeManagerDevTemplate = `{ config, pkgs, ... }:

{
  # Home Manager development configuration
  programs.git = {
    enable = true;
    userName = "{{GetOption "git_name" "Your Name"}}";
    userEmail = "{{GetOption "git_email" "your.email@example.com"}}";
    extraConfig = {
      init.defaultBranch = "main";
      pull.rebase = true;
    };
  };

  {{if ne (GetOption "enable_zsh" "true") "false"}}
  programs.zsh = {
    enable = true;
    enableAutosuggestions = true;
    enableCompletion = true;
    shellAliases = {
      ll = "ls -l";
      la = "ls -la";
      grep = "grep --color=auto";
      ".." = "cd ..";
    };
    
    oh-my-zsh = {
      enable = true;
      theme = "robbyrussell";
      plugins = [ "git" "sudo" ];
    };
  };
  {{end}}

  {{if HasComponent "neovim"}}
  programs.neovim = {
    enable = true;
    defaultEditor = true;
    configure = {
      customRC = ''
        set number
        set relativenumber
        set tabstop=2
        set shiftwidth=2
        set expandtab
      '';
    };
  };
  {{end}}

  {{if HasComponent "vscode"}}
  programs.vscode = {
    enable = true;
    extensions = with pkgs.vscode-extensions; [
      ms-python.python
      ms-vscode.cpptools
      bbenoist.nix
    ];
  };
  {{end}}

  # Development packages
  home.packages = with pkgs; [
    curl
    wget
    jq
    tree
    ripgrep
    {{if HasComponent "python"}}
    python3
    {{end}}
    {{if HasComponent "nodejs"}}
    nodejs
    npm
    {{end}}
  ];

  # This value determines the Home Manager release
  home.stateVersion = "25.05";
}`

const gamingBasicTemplate = `{ config, pkgs, ... }:

{
  # Enable unfree packages (required for Steam)
  nixpkgs.config.allowUnfree = true;

  # Gaming packages
  environment.systemPackages = with pkgs; [
    {{if ne (GetOption "enable_steam" "true") "false"}}
    steam
    {{end}}
    {{if HasComponent "lutris"}}
    lutris
    {{end}}
    {{if HasComponent "wine"}}
    wine
    winetricks
    {{end}}
    # Gaming utilities
    gamemode
    mangohud
  ];

  # Steam configuration
  {{if ne (GetOption "enable_steam" "true") "false"}}
  programs.steam = {
    enable = true;
    remotePlay.openFirewall = true;
    dedicatedServer.openFirewall = true;
  };
  {{end}}

  # GameMode for performance
  programs.gamemode.enable = true;

  # Graphics drivers
  {{if eq (GetOption "gpu_driver" "") "nvidia"}}
  services.xserver.videoDrivers = [ "nvidia" ];
  hardware.opengl = {
    enable = true;
    driSupport = true;
    driSupport32Bit = true;
  };
  hardware.nvidia = {
    modesetting.enable = true;
    open = false; # Use proprietary driver
    nvidiaSettings = true;
  };
  {{else if eq (GetOption "gpu_driver" "") "amd"}}
  services.xserver.videoDrivers = [ "amdgpu" ];
  hardware.opengl = {
    enable = true;
    driSupport = true;
    driSupport32Bit = true;
    extraPackages = with pkgs; [
      rocm-opencl-icd
      rocm-opencl-runtime
    ];
  };
  {{end}}

  # Audio configuration for gaming
  security.rtkit.enable = true;
  services.pipewire = {
    enable = true;
    alsa.enable = true;
    alsa.support32Bit = true;
    pulse.enable = true;
  };

  # This value determines the NixOS release
  system.stateVersion = "25.05";
}`

const securityBasicTemplate = `{ config, pkgs, ... }:

{
  # Basic security hardening configuration
  
  # Firewall configuration
  networking.firewall = {
    enable = true;
    allowedTCPPorts = [ 22 ]; # SSH only by default
    allowedUDPPorts = [ ];
    # Disable ping
    allowPing = false;
  };

  # SSH hardening
  services.openssh = {
    enable = true;
    settings = {
      PasswordAuthentication = false;
      PermitRootLogin = "no";
      X11Forwarding = false;
    };
    # Key-based authentication only
  };

  # Fail2ban for intrusion prevention
  services.fail2ban = {
    enable = true;
    bantime = "1h";
    bantime-increment.enable = true;
    jails = {
      ssh.settings = {
        enabled = true;
        port = "22";
      };
    };
  };

  # System hardening
  security = {
    # Disable sudo timeout
    sudo.execWheelOnly = true;
    
    # AppArmor (optional)
    # apparmor.enable = true;
  };

  # Automatic security updates
  system.autoUpgrade = {
    enable = true;
    allowReboot = false; # Set to true for servers
    dates = "daily";
    flags = [ "--upgrade-all" ];
  };

  # Audit system
  security.audit.enable = true;
  security.auditd.enable = true;

  # This value determines the NixOS release
  system.stateVersion = "25.05";
}`
