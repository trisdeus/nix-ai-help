package neovim

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Snippet represents a code snippet for Neovim
type Snippet struct {
	Prefix      string   `json:"prefix"`
	Body        []string `json:"body"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Context     []string `json:"context,omitempty"`
}

// SnippetProvider provides comprehensive Nix snippets
type SnippetProvider struct {
	snippets map[string]Snippet
}

// NewSnippetProvider creates a new snippet provider
func NewSnippetProvider() *SnippetProvider {
	sp := &SnippetProvider{
		snippets: make(map[string]Snippet),
	}
	sp.loadDefaultSnippets()
	return sp
}

// GetSnippets returns all available snippets
func (sp *SnippetProvider) GetSnippets() map[string]Snippet {
	return sp.snippets
}

// GetSnippetsByCategory returns snippets filtered by category
func (sp *SnippetProvider) GetSnippetsByCategory(category string) map[string]Snippet {
	filtered := make(map[string]Snippet)
	for key, snippet := range sp.snippets {
		if snippet.Category == category {
			filtered[key] = snippet
		}
	}
	return filtered
}

// GetSnippetsByContext returns snippets relevant to a specific context
func (sp *SnippetProvider) GetSnippetsByContext(context string) map[string]Snippet {
	filtered := make(map[string]Snippet)
	for key, snippet := range sp.snippets {
		for _, ctx := range snippet.Context {
			if ctx == context {
				filtered[key] = snippet
				break
			}
		}
	}
	return filtered
}

// SearchSnippets searches snippets by keyword
func (sp *SnippetProvider) SearchSnippets(keyword string) map[string]Snippet {
	keyword = strings.ToLower(keyword)
	filtered := make(map[string]Snippet)
	
	for key, snippet := range sp.snippets {
		if strings.Contains(strings.ToLower(snippet.Prefix), keyword) ||
		   strings.Contains(strings.ToLower(snippet.Description), keyword) ||
		   strings.Contains(strings.ToLower(snippet.Category), keyword) {
			filtered[key] = snippet
		}
	}
	return filtered
}

// GetSnippetAsLuaSnip converts snippet to LuaSnip format
func (sp *SnippetProvider) GetSnippetAsLuaSnip(snippetKey string) string {
	snippet, exists := sp.snippets[snippetKey]
	if !exists {
		return ""
	}
	
	return fmt.Sprintf(`s("%s", fmt([[%s]], {}), { desc = "%s" })`, 
		snippet.Prefix, strings.Join(snippet.Body, "\n"), snippet.Description)
}

// GetAllSnippetsAsLuaSnip returns all snippets in LuaSnip format
func (sp *SnippetProvider) GetAllSnippetsAsLuaSnip() string {
	var luaSnips []string
	luaSnips = append(luaSnips, "local ls = require('luasnip')")
	luaSnips = append(luaSnips, "local s = ls.snippet")
	luaSnips = append(luaSnips, "local fmt = require('luasnip.extras.fmt').fmt")
	luaSnips = append(luaSnips, "")
	luaSnips = append(luaSnips, "return {")
	
	for key := range sp.snippets {
		luaSnips = append(luaSnips, "  "+sp.GetSnippetAsLuaSnip(key)+",")
	}
	
	luaSnips = append(luaSnips, "}")
	return strings.Join(luaSnips, "\n")
}

// loadDefaultSnippets loads the comprehensive snippet library
func (sp *SnippetProvider) loadDefaultSnippets() {
	// System Configuration Snippets
	sp.snippets["nixos-basic"] = Snippet{
		Prefix:      "nixos-basic",
		Body:        []string{
			"{ config, pkgs, ... }:",
			"",
			"{",
			"  imports = [",
			"    ./hardware-configuration.nix",
			"    $" + "{1:# additional imports}",
			"  ];",
			"",
			"  # System configuration",
			"  system.stateVersion = \"$" + "{2:24.05}\";",
			"",
			"  # Enable flakes",
			"  nix.settings.experimental-features = [ \"nix-command\" \"flakes\" ];",
			"",
			"  # System packages",
			"  environment.systemPackages = with pkgs; [",
			"    $" + "{3:# Add packages here}",
			"  ];",
			"",
			"  $" + "{0:# Additional configuration}",
			"}",
		},
		Description: "Basic NixOS configuration template",
		Category:    "system",
		Context:     []string{"configuration.nix"},
	}

	// Service Configuration Snippets
	sp.snippets["service-enable"] = Snippet{
		Prefix:      "service",
		Body:        []string{
			"services.$" + "{1:serviceName} = {",
			"  enable = true;",
			"  $" + "{2:# Additional configuration}",
			"};",
		},
		Description: "Enable a system service",
		Category:    "services",
		Context:     []string{"services"},
	}

	sp.snippets["nginx-server"] = Snippet{
		Prefix:      "nginx",
		Body:        []string{
			"services.nginx = {",
			"  enable = true;",
			"  virtualHosts.\"$" + "{1:example.com}\" = {",
			"    $" + "{2:enableACME = true;}",
			"    $" + "{3:forceSSL = true;}",
			"    locations.\"/\" = {",
			"      $" + "{4:proxyPass = \"http://localhost:$" + "{5:3000}\";}",
			"      $" + "{6:# Additional location config}",
			"    };",
			"  };",
			"};",
		},
		Description: "Nginx virtual host configuration",
		Category:    "services",
		Context:     []string{"nginx", "web-server"},
	}

	sp.snippets["ssh-server"] = Snippet{
		Prefix:      "ssh",
		Body:        []string{
			"services.openssh = {",
			"  enable = true;",
			"  settings = {",
			"    PasswordAuthentication = $" + "{1:false};",
			"    PermitRootLogin = \"$" + "{2:no}\";",
			"    $" + "{3:# Additional SSH settings}",
			"  };",
			"  $" + "{4:ports = [ 22 ];}",
			"};",
		},
		Description: "SSH server configuration",
		Category:    "services",
		Context:     []string{"ssh", "security"},
	}

	// Package Management Snippets
	sp.snippets["packages"] = Snippet{
		Prefix:      "packages",
		Body:        []string{
			"environment.systemPackages = with pkgs; [",
			"  $" + "{1:# Development}",
			"  git",
			"  vim",
			"  curl",
			"  wget",
			"",
			"  $" + "{2:# System utilities}",
			"  htop",
			"  tree",
			"  unzip",
			"",
			"  $" + "{0:# Add more packages}",
			"];",
		},
		Description: "System packages configuration",
		Category:    "packages",
		Context:     []string{"packages", "environment"},
	}

	sp.snippets["overlay"] = Snippet{
		Prefix:      "overlay",
		Body:        []string{
			"nixpkgs.overlays = [",
			"  (final: prev: {",
			"    $" + "{1:packageName} = prev.$" + "{1:packageName}.overrideAttrs (oldAttrs: {",
			"      $" + "{2:# Override attributes}",
			"    });",
			"  })",
			"];",
		},
		Description: "Package overlay configuration",
		Category:    "packages",
		Context:     []string{"overlay", "packages"},
	}

	// User Management Snippets
	sp.snippets["user"] = Snippet{
		Prefix:      "user",
		Body:        []string{
			"users.users.$" + "{1:username} = {",
			"  isNormalUser = true;",
			"  description = \"$" + "{2:User Description}\";",
			"  extraGroups = [ \"$" + "{3:wheel}\" $" + "{4:\"networkmanager\"} ];",
			"  $" + "{5:shell = pkgs.$" + "{6:bash};}",
			"  $" + "{7:openssh.authorizedKeys.keys = [}",
			"    $" + "{8:# \"ssh-rsa AAAA...\"}",
			"  $" + "{7:];}",
			"};",
		},
		Description: "User account configuration",
		Category:    "users",
		Context:     []string{"users"},
	}

	// Hardware Configuration Snippets
	sp.snippets["nvidia"] = Snippet{
		Prefix:      "nvidia",
		Body:        []string{
			"# Enable OpenGL",
			"hardware.opengl = {",
			"  enable = true;",
			"  driSupport = true;",
			"  driSupport32Bit = true;",
			"};",
			"",
			"# NVIDIA drivers",
			"services.xserver.videoDrivers = [ \"nvidia\" ];",
			"hardware.nvidia = {",
			"  modesetting.enable = true;",
			"  powerManagement.enable = $" + "{1:false};",
			"  powerManagement.finegrained = $" + "{2:false};",
			"  open = $" + "{3:false};",
			"  nvidiaSettings = true;",
			"  package = config.boot.kernelPackages.nvidiaPackages.$" + "{4:stable};",
			"};",
		},
		Description: "NVIDIA graphics configuration",
		Category:    "hardware",
		Context:     []string{"nvidia", "graphics"},
	}

	sp.snippets["bluetooth"] = Snippet{
		Prefix:      "bluetooth",
		Body:        []string{
			"# Bluetooth",
			"hardware.bluetooth = {",
			"  enable = true;",
			"  powerOnBoot = $" + "{1:true};",
			"  $" + "{2:# Additional bluetooth settings}",
			"};",
			"services.blueman.enable = $" + "{3:true};",
		},
		Description: "Bluetooth configuration",
		Category:    "hardware",
		Context:     []string{"bluetooth", "hardware"},
	}

	// Desktop Environment Snippets
	sp.snippets["gnome"] = Snippet{
		Prefix:      "gnome",
		Body:        []string{
			"# Enable the X11 windowing system",
			"services.xserver.enable = true;",
			"",
			"# Enable GNOME Desktop Environment",
			"services.xserver.displayManager.gdm.enable = true;",
			"services.xserver.desktopManager.gnome.enable = true;",
			"",
			"# Configure keymap",
			"services.xserver.xkb = {",
			"  layout = \"$" + "{1:us}\";",
			"  variant = \"$" + "{2:}\";",
			"};",
			"",
			"$" + "{3:# Exclude some default GNOME applications}",
			"$" + "{3:environment.gnome.excludePackages = (with pkgs; [}",
			"  $" + "{3:gnome-photos}",
			"  $" + "{3:gnome-tour}",
			"$" + "{3:]) ++ (with pkgs.gnome; [}",
			"  $" + "{3:cheese # webcam tool}",
			"  $" + "{3:gnome-music}",
			"  $" + "{3:epiphany # web browser}",
			"  $" + "{3:geary # email reader}",
			"$" + "{3:]);}",
		},
		Description: "GNOME desktop environment",
		Category:    "desktop",
		Context:     []string{"gnome", "desktop"},
	}

	sp.snippets["kde"] = Snippet{
		Prefix:      "kde",
		Body:        []string{
			"# Enable the X11 windowing system",
			"services.xserver.enable = true;",
			"",
			"# Enable KDE Plasma Desktop Environment",
			"services.displayManager.sddm.enable = true;",
			"services.desktopManager.plasma6.enable = true;",
			"",
			"# Configure keymap",
			"services.xserver.xkb = {",
			"  layout = \"$" + "{1:us}\";",
			"  variant = \"$" + "{2:}\";",
			"};",
		},
		Description: "KDE Plasma desktop environment",
		Category:    "desktop",
		Context:     []string{"kde", "plasma", "desktop"},
	}

	// Networking Snippets
	sp.snippets["firewall"] = Snippet{
		Prefix:      "firewall",
		Body:        []string{
			"networking.firewall = {",
			"  enable = $" + "{1:true};",
			"  allowedTCPPorts = [ $" + "{2:# 22 80 443} ];",
			"  allowedUDPPorts = [ $" + "{3:# 53} ];",
			"  $" + "{4:# allowedTCPPortRanges = [}",
			"  $" + "{4:#   { from = 4000; to = 4007; }}",
			"  $" + "{4:# ];}",
			"};",
		},
		Description: "Firewall configuration",
		Category:    "networking",
		Context:     []string{"firewall", "security"},
	}

	sp.snippets["network-manager"] = Snippet{
		Prefix:      "networkmanager",
		Body:        []string{
			"# Enable networking",
			"networking.networkmanager.enable = true;",
			"",
			"# Add user to networkmanager group",
			"users.users.$" + "{1:username}.extraGroups = [ \"networkmanager\" ];",
		},
		Description: "NetworkManager configuration",
		Category:    "networking",
		Context:     []string{"networking"},
	}

	// Boot Configuration Snippets
	sp.snippets["systemd-boot"] = Snippet{
		Prefix:      "systemd-boot",
		Body:        []string{
			"# Use systemd-boot EFI boot loader",
			"boot.loader.systemd-boot.enable = true;",
			"boot.loader.efi.canTouchEfiVariables = true;",
			"$" + "{1:boot.loader.systemd-boot.configurationLimit = $" + "{2:10};}",
		},
		Description: "systemd-boot configuration",
		Category:    "boot",
		Context:     []string{"boot", "bootloader"},
	}

	sp.snippets["grub"] = Snippet{
		Prefix:      "grub",
		Body:        []string{
			"# Use GRUB boot loader",
			"boot.loader.grub = {",
			"  enable = true;",
			"  device = \"$" + "{1:/dev/sda}\"; # or \"nodev\" for EFI",
			"  $" + "{2:useOSProber = true;}",
			"  $" + "{3:# Additional GRUB configuration}",
			"};",
		},
		Description: "GRUB bootloader configuration",
		Category:    "boot",
		Context:     []string{"boot", "grub"},
	}

	// Home Manager Snippets
	sp.snippets["home-manager"] = Snippet{
		Prefix:      "home-manager",
		Body:        []string{
			"{ config, pkgs, ... }:",
			"",
			"{",
			"  # Home Manager needs a bit of information about you and the paths it should",
			"  # manage.",
			"  home.username = \"$" + "{1:username}\";",
			"  home.homeDirectory = \"/home/$" + "{1:username}\";",
			"",
			"  # This value determines the Home Manager release that your configuration is",
			"  # compatible with.",
			"  home.stateVersion = \"$" + "{2:24.05}\";",
			"",
			"  # The home.packages option allows you to install Nix packages into your",
			"  # environment.",
			"  home.packages = with pkgs; [",
			"    $" + "{3:# Add packages here}",
			"  ];",
			"",
			"  # Let Home Manager install and manage itself.",
			"  programs.home-manager.enable = true;",
			"",
			"  $" + "{0:# Additional configuration}",
			"}",
		},
		Description: "Basic Home Manager configuration",
		Category:    "home-manager",
		Context:     []string{"home.nix"},
	}

	sp.snippets["hm-program"] = Snippet{
		Prefix:      "hm-program",
		Body:        []string{
			"programs.$" + "{1:programName} = {",
			"  enable = true;",
			"  $" + "{2:# Program-specific configuration}",
			"};",
		},
		Description: "Home Manager program configuration",
		Category:    "home-manager",
		Context:     []string{"programs"},
	}

	// Flake Snippets
	sp.snippets["flake"] = Snippet{
		Prefix:      "flake",
		Body:        []string{
			"{",
			"  description = \"$" + "{1:NixOS configuration flake}\";",
			"",
			"  inputs = {",
			"    nixpkgs.url = \"github:NixOS/nixpkgs/nixos-$" + "{2:unstable}\";",
			"    $" + "{3:home-manager = {}",
			"    $" + "{3:  url = \"github:nix-community/home-manager\";",
			"    $" + "{3:  inputs.nixpkgs.follows = \"nixpkgs\";",
			"    $" + "{3:};}",
			"  };",
			"",
			"  outputs = { self, nixpkgs, $" + "{4:home-manager, }... }@inputs: {",
			"    nixosConfigurations.$" + "{5:hostname} = nixpkgs.lib.nixosSystem {",
			"      system = \"$" + "{6:x86_64-linux}\";",
			"      modules = [",
			"        ./configuration.nix",
			"        $" + "{7:home-manager.nixosModules.home-manager",
			"        $" + "{7:{",
			"        $" + "{7:  home-manager.useGlobalPkgs = true;",
			"        $" + "{7:  home-manager.useUserPackages = true;",
			"        $" + "{7:  home-manager.users.$" + "{8:username} = import ./home.nix;",
			"        $" + "{7:}}",
			"      ];",
			"    };",
			"  };",
			"}",
		},
		Description: "Basic NixOS flake configuration",
		Category:    "flakes",
		Context:     []string{"flake.nix"},
	}

	// Development Environment Snippets
	sp.snippets["shell-nix"] = Snippet{
		Prefix:      "shell",
		Body:        []string{
			"{ pkgs ? import <nixpkgs> {} }:",
			"",
			"pkgs.mkShell {",
			"  buildInputs = with pkgs; [",
			"    $" + "{1:# Development dependencies}",
			"  ];",
			"",
			"  shellHook = ''",
			"    $" + "{2:echo \"Welcome to the development environment!\"}",
			"    $" + "{3:# Additional shell setup}",
			"  '';",
			"",
			"  $" + "{4:# Environment variables}",
			"  $" + "{4:# $" + "{5:VARIABLE_NAME} = \"$" + "{6:value}\";}",
			"}",
		},
		Description: "Development shell environment",
		Category:    "development",
		Context:     []string{"shell.nix"},
	}

	// Security Snippets
	sp.snippets["security-basic"] = Snippet{
		Prefix:      "security",
		Body:        []string{
			"# Basic security configuration",
			"security = {",
			"  sudo.wheelNeedsPassword = $" + "{1:true};",
			"  $" + "{2:rtkit.enable = true;}",
			"  $" + "{3:polkit.enable = true;}",
			"};",
			"",
			"# Disable root login",
			"users.users.root.hashedPassword = \"!\";",
			"",
			"# Enable automatic security updates",
			"system.autoUpgrade = {",
			"  enable = $" + "{4:true};",
			"  allowReboot = $" + "{5:false};",
			"};",
		},
		Description: "Basic security hardening",
		Category:    "security",
		Context:     []string{"security"},
	}

	// Virtualization Snippets
	sp.snippets["docker"] = Snippet{
		Prefix:      "docker",
		Body:        []string{
			"# Enable Docker",
			"virtualisation.docker = {",
			"  enable = true;",
			"  $" + "{1:enableOnBoot = true;}",
			"  $" + "{2:# rootless = {}",
			"  $" + "{2:#   enable = true;",
			"  $" + "{2:#   setSocketVariable = true;",
			"  $" + "{2:# };}",
			"};",
			"",
			"# Add user to docker group",
			"users.users.$" + "{3:username}.extraGroups = [ \"docker\" ];",
		},
		Description: "Docker configuration",
		Category:    "virtualization",
		Context:     []string{"docker", "containers"},
	}

	// Monitoring Snippets
	sp.snippets["prometheus"] = Snippet{
		Prefix:      "prometheus",
		Body:        []string{
			"services.prometheus = {",
			"  enable = true;",
			"  port = $" + "{1:9090};",
			"  scrapeConfigs = [",
			"    {",
			"      job_name = \"$" + "{2:node}\";",
			"      static_configs = [{",
			"        targets = [ \"localhost:$" + "{3:9100}\" ];",
			"      }];",
			"    }",
			"  ];",
			"};",
			"",
			"# Enable node exporter",
			"services.prometheus.exporters.node = {",
			"  enable = true;",
			"  port = $" + "{3:9100};",
			"};",
		},
		Description: "Prometheus monitoring setup",
		Category:    "monitoring",
		Context:     []string{"monitoring", "prometheus"},
	}
}

// GetSnippetCategories returns all available categories
func (sp *SnippetProvider) GetSnippetCategories() []string {
	categories := make(map[string]bool)
	for _, snippet := range sp.snippets {
		categories[snippet.Category] = true
	}
	
	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}
	return result
}

// AddCustomSnippet adds a custom snippet
func (sp *SnippetProvider) AddCustomSnippet(key string, snippet Snippet) {
	sp.snippets[key] = snippet
}

// ExportSnippetsToVSCode returns snippets in VS Code format
func (sp *SnippetProvider) ExportSnippetsToVSCode() string {
	vsCodeSnippets := make(map[string]map[string]interface{})
	
	for key, snippet := range sp.snippets {
		vsCodeSnippets[key] = map[string]interface{}{
			"prefix":      snippet.Prefix,
			"body":        snippet.Body,
			"description": snippet.Description,
		}
	}
	
	result, _ := json.MarshalIndent(vsCodeSnippets, "", "  ")
	return string(result)
}