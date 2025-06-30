package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	yaml "gopkg.in/yaml.v3"
)

// GitHub API structures for searching code
type GitHubSearchResponse struct {
	TotalCount int                  `json:"total_count"`
	Items      []GitHubSearchResult `json:"items"`
}

type GitHubSearchResult struct {
	Name       string               `json:"name"`
	Path       string               `json:"path"`
	Sha        string               `json:"sha"`
	URL        string               `json:"url"`
	GitURL     string               `json:"git_url"`
	HTMLURL    string               `json:"html_url"`
	Repository GitHubRepositoryInfo `json:"repository"`
	Score      float64              `json:"score"`
}

type GitHubRepositoryInfo struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	FullName    string      `json:"full_name"`
	Owner       GitHubOwner `json:"owner"`
	HTMLURL     string      `json:"html_url"`
	Description string      `json:"description"`
	Language    string      `json:"language"`
	StarCount   int         `json:"stargazers_count"`
	ForksCount  int         `json:"forks_count"`
	UpdatedAt   string      `json:"updated_at"`
}

type GitHubOwner struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

// GitHubContentResponse represents GitHub file content API response
type GitHubContentResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
}

// Template represents a configuration template
type Template struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Category    string            `yaml:"category"`
	Tags        []string          `yaml:"tags"`
	Source      string            `yaml:"source"`      // "builtin", "github", "custom"
	GitHubRepo  string            `yaml:"github_repo"` // For GitHub templates
	FilePath    string            `yaml:"file_path"`   // Path within repo
	Content     string            `yaml:"content"`     // Template content
	Metadata    map[string]string `yaml:"metadata"`
}

// Snippet represents a saved configuration snippet
type Snippet struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Tags        []string          `yaml:"tags"`
	Content     string            `yaml:"content"`
	CreatedAt   time.Time         `yaml:"created_at"`
	Source      string            `yaml:"source"` // "user", "template", "github"
	Metadata    map[string]string `yaml:"metadata"`
}

// TemplateManager manages templates and snippets
type TemplateManager struct {
	configDir string
	logger    *logger.Logger
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(configDir string, log *logger.Logger) *TemplateManager {
	if configDir == "" {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config", "nixai")
	}

	// Ensure config directory exists
	_ = os.MkdirAll(configDir, 0755)
	_ = os.MkdirAll(filepath.Join(configDir, "templates"), 0755)
	_ = os.MkdirAll(filepath.Join(configDir, "snippets"), 0755)

	return &TemplateManager{
		configDir: configDir,
		logger:    log,
	}
}

// Methods for TemplateManager

// GetTemplate retrieves a specific template by name
func (tm *TemplateManager) GetTemplate(name string) (*Template, error) {
	// Check builtin templates first
	builtinTemplates := tm.LoadBuiltinTemplates()
	for _, template := range builtinTemplates {
		if template.Name == name {
			return &template, nil
		}
	}

	// Check saved custom templates
	customTemplates, err := tm.LoadCustomTemplates()
	if err == nil {
		for _, template := range customTemplates {
			if template.Name == name {
				return &template, nil
			}
		}
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

// LoadBuiltinTemplates loads the built-in templates
func (tm *TemplateManager) LoadBuiltinTemplates() []Template {
	return []Template{
		{
			Name:        "desktop",
			Description: "Basic desktop environment configuration",
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

  # Enable touchpad support (enabled by default in most desktopManager)
  # services.xserver.libinput.enable = true;

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

  # Some programs need SUID wrappers, can be configured further or are
  # started in user sessions.
  # programs.mtr.enable = true;
  # programs.gnupg.agent = {
  #   enable = true;
  #   enableSSHSupport = true;
  # };

  # List services that you want to enable:

  # Enable the OpenSSH daemon.
  # services.openssh.enable = true;

  # Open ports in the firewall.
  # networking.firewall.allowedTCPPorts = [ ... ];
  # networking.firewall.allowedUDPPorts = [ ... ];
  # Or disable the firewall altogether.
  # networking.firewall.enable = false;

  # This value determines the NixOS release from which the default
  # settings for stateful data, like file locations and database versions
  # on your system were taken. It's perfectly fine and recommended to leave
  # this value at the release version of the first install of this system.
  # Before changing this value read the documentation for this option
  # (e.g. man configuration.nix or on https://nixos.org/nixos/options.html).
  system.stateVersion = "25.05"; # Did you read the comment?
}`,
			Metadata: map[string]string{
				"type":        "desktop",
				"complexity":  "basic",
				"environment": "gnome",
			},
		},
		{
			Name:        "server",
			Description: "Basic server configuration",
			Category:    "Server",
			Tags:        []string{"server", "headless", "basic"},
			Source:      "builtin",
			Content: `{ config, pkgs, ... }:

{
  # Boot loader
  boot.loader.systemd-boot.enable = true;
  boot.loader.efi.canTouchEfiVariables = true;

  # Networking
  networking.hostName = "nixos-server"; # Define your hostname
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

  # Enable automatic system updates (optional)
  # system.autoUpgrade = {
  #   enable = true;
  #   dates = "04:00";
  # };

  # This value determines the NixOS release from which the default
  # settings for stateful data were taken
  system.stateVersion = "25.05";
}`,
			Metadata: map[string]string{
				"type":        "server",
				"complexity":  "basic",
				"environment": "headless",
			},
		},
		{
			Name:        "development",
			Description: "Development environment with common tools",
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

  # This value determines the NixOS release from which the default
  # settings for stateful data were taken
  system.stateVersion = "25.05";
}`,
			Metadata: map[string]string{
				"type":        "development",
				"complexity":  "intermediate",
				"environment": "desktop",
			},
		},
	}
}

// LoadCustomTemplates loads templates saved by the user
func (tm *TemplateManager) LoadCustomTemplates() ([]Template, error) {
	templatesDir := filepath.Join(tm.configDir, "templates")

	var templates []Template

	// Read all YAML files in templates directory
	files, err := filepath.Glob(filepath.Join(templatesDir, "*.yaml"))
	if err != nil {
		return templates, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			tm.logger.Warn("Failed to read template file: " + file)
			continue
		}

		var template Template
		if err := yaml.Unmarshal(data, &template); err != nil {
			tm.logger.Warn("Failed to parse template file: " + file)
			continue
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// ApplyTemplate applies a template to the configuration
func (tm *TemplateManager) ApplyTemplate(template *Template, outputPath string, merge bool) error {
	content := template.Content

	// If no output path specified, use default
	if outputPath == "" {
		outputPath = "/etc/nixos/configuration.nix"

		// Check if we have permission to write to /etc/nixos
		if _, err := os.Stat("/etc/nixos"); os.IsNotExist(err) {
			// Fallback to current directory
			outputPath = "./configuration.nix"
		}
	}

	if merge {
		// Merge with existing configuration
		if existingContent, err := os.ReadFile(outputPath); err == nil {
			// Simple merge - in a real implementation, this would be more sophisticated
			content = string(existingContent) + "\n\n# Added from template: " + template.Name + "\n" + content
		}
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write configuration
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %v", err)
	}

	return nil
}

// SaveTemplate saves a new template from a source
func (tm *TemplateManager) SaveTemplate(name, source, category, description string, tags []string) error {
	var content string
	var gitHubRepo, filePath string

	// Determine source type and read content
	if strings.HasPrefix(source, "http") {
		// GitHub URL
		var err error
		content, gitHubRepo, filePath, err = tm.fetchGitHubContent(source)
		if err != nil {
			return fmt.Errorf("failed to fetch GitHub content: %v", err)
		}
	} else {
		// Local file
		data, err := os.ReadFile(source)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", source, err)
		}
		content = string(data)
	}

	// Create template
	template := Template{
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        tags,
		Source:      "custom",
		GitHubRepo:  gitHubRepo,
		FilePath:    filePath,
		Content:     content,
		Metadata:    make(map[string]string),
	}

	// Set default description if empty
	if template.Description == "" {
		template.Description = "Custom template saved from " + source
	}

	// Set default category if empty
	if template.Category == "" {
		template.Category = "Custom"
	}

	// Save to file
	templatePath := filepath.Join(tm.configDir, "templates", name+".yaml")
	data, err := yaml.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %v", err)
	}

	if err := os.WriteFile(templatePath, data, 0644); err != nil {
		return fmt.Errorf("failed to save template: %v", err)
	}

	return nil
}

// fetchGitHubContent fetches content from a GitHub URL
func (tm *TemplateManager) fetchGitHubContent(url string) (content, repo, path string, err error) {
	// Parse GitHub URL to extract repo and file path
	// Example: https://github.com/user/repo/blob/main/config.nix
	// Convert to raw URL: https://raw.githubusercontent.com/user/repo/main/config.nix

	if strings.Contains(url, "github.com") && strings.Contains(url, "/blob/") {
		// Parse the URL to extract parts
		parts := strings.Split(url, "/")
		if len(parts) >= 7 {
			user := parts[3]
			repoName := parts[4]
			branch := parts[6]
			filePath := strings.Join(parts[7:], "/")

			repo = fmt.Sprintf("%s/%s", user, repoName)
			path = filePath

			// Construct raw URL
			rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", user, repoName, branch, filePath)

			// Fetch content
			resp, err := http.Get(rawURL)
			if err != nil {
				return "", "", "", err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				return "", "", "", fmt.Errorf("failed to fetch content: HTTP %d", resp.StatusCode)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", "", "", err
			}

			content = string(data)
			return content, repo, path, nil
		}
	}

	return "", "", "", fmt.Errorf("unsupported URL format")
}

// GetCategories returns template categories with counts
func (tm *TemplateManager) GetCategories() map[string]int {
	categories := make(map[string]int)

	// Count builtin templates
	builtinTemplates := tm.LoadBuiltinTemplates()
	for _, template := range builtinTemplates {
		category := template.Category
		if category == "" {
			category = "General"
		}
		categories[category]++
	}

	// Count custom templates
	customTemplates, err := tm.LoadCustomTemplates()
	if err == nil {
		for _, template := range customTemplates {
			category := template.Category
			if category == "" {
				category = "General"
			}
			categories[category]++
		}
	}

	return categories
}

// LoadSnippets loads all saved snippets
func (tm *TemplateManager) LoadSnippets() ([]Snippet, error) {
	snippetsDir := filepath.Join(tm.configDir, "snippets")

	var snippets []Snippet

	// Read all YAML files in snippets directory
	files, err := filepath.Glob(filepath.Join(snippetsDir, "*.yaml"))
	if err != nil {
		return snippets, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			tm.logger.Warn("Failed to read snippet file: " + file)
			continue
		}

		var snippet Snippet
		if err := yaml.Unmarshal(data, &snippet); err != nil {
			tm.logger.Warn("Failed to parse snippet file: " + file)
			continue
		}

		snippets = append(snippets, snippet)
	}

	// Sort by creation time (newest first)
	sort.Slice(snippets, func(i, j int) bool {
		return snippets[i].CreatedAt.After(snippets[j].CreatedAt)
	})

	return snippets, nil
}

// SearchSnippets searches snippets by query
func (tm *TemplateManager) SearchSnippets(query string) ([]Snippet, error) {
	allSnippets, err := tm.LoadSnippets()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matches []Snippet

	for _, snippet := range allSnippets {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(snippet.Name), query) ||
			strings.Contains(strings.ToLower(snippet.Description), query) {
			matches = append(matches, snippet)
			continue
		}

		// Search in tags
		for _, tag := range snippet.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, snippet)
				break
			}
		}
	}

	return matches, nil
}

// SaveSnippet saves a new snippet
func (tm *TemplateManager) SaveSnippet(name, filePath, description string, tags []string) error {
	var content string

	// Read content from file or stdin
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", filePath, err)
		}
		content = string(data)
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %v", err)
			}
			content = string(data)
		} else {
			return fmt.Errorf("no content provided - specify --file or pipe content")
		}
	}

	// Create snippet
	snippet := Snippet{
		Name:        name,
		Description: description,
		Tags:        tags,
		Content:     content,
		CreatedAt:   time.Now(),
		Source:      "user",
		Metadata:    make(map[string]string),
	}

	// Set default description if empty
	if snippet.Description == "" {
		snippet.Description = "User-created snippet"
	}

	// Save to file
	snippetPath := filepath.Join(tm.configDir, "snippets", name+".yaml")
	data, err := yaml.Marshal(snippet)
	if err != nil {
		return fmt.Errorf("failed to marshal snippet: %v", err)
	}

	if err := os.WriteFile(snippetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save snippet: %v", err)
	}

	return nil
}

// GetSnippet retrieves a specific snippet by name
func (tm *TemplateManager) GetSnippet(name string) (*Snippet, error) {
	snippets, err := tm.LoadSnippets()
	if err != nil {
		return nil, err
	}

	for _, snippet := range snippets {
		if snippet.Name == name {
			return &snippet, nil
		}
	}

	return nil, fmt.Errorf("snippet not found: %s", name)
}

// ApplySnippet applies a snippet to configuration
func (tm *TemplateManager) ApplySnippet(name, outputPath string) error {
	snippet, err := tm.GetSnippet(name)
	if err != nil {
		return err
	}

	if outputPath == "" {
		// Output to stdout if no file specified
		fmt.Print(snippet.Content)
		return nil
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Write content
	if err := os.WriteFile(outputPath, []byte(snippet.Content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// RemoveSnippet removes a snippet
func (tm *TemplateManager) RemoveSnippet(name string) error {
	snippetPath := filepath.Join(tm.configDir, "snippets", name+".yaml")

	if _, err := os.Stat(snippetPath); os.IsNotExist(err) {
		return fmt.Errorf("snippet not found: %s", name)
	}

	if err := os.Remove(snippetPath); err != nil {
		return fmt.Errorf("failed to remove snippet: %v", err)
	}

	return nil
}

// getCategoryDescription provides descriptions for template categories
func getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"Desktop":     "Desktop environment configurations",
		"Gaming":      "Gaming-optimized configurations",
		"Server":      "Server and headless configurations",
		"Development": "Development environment setups",
		"Minimal":     "Minimal and lightweight configurations",
		"Hardware":    "Hardware-specific configurations",
		"Security":    "Security-hardened configurations",
		"Custom":      "User-created templates",
		"General":     "General purpose configurations",
	}

	if desc, exists := descriptions[category]; exists {
		return desc
	}
	return "Configuration templates"
}

// LoadAllTemplates loads templates from all sources (builtin, custom, GitHub cache)
func (tm *TemplateManager) LoadAllTemplates() ([]Template, error) {
	var allTemplates []Template

	// Load builtin templates
	builtinTemplates := tm.LoadBuiltinTemplates()
	allTemplates = append(allTemplates, builtinTemplates...)

	// Load custom templates
	customTemplates, err := tm.LoadCustomTemplates()
	if err != nil {
		tm.logger.Warn("Failed to load custom templates: " + err.Error())
	} else {
		allTemplates = append(allTemplates, customTemplates...)
	}

	// Load cached GitHub templates
	githubTemplates, err := tm.LoadGitHubTemplatesCache()
	if err != nil {
		tm.logger.Debug("Failed to load GitHub templates cache: " + err.Error())
	} else {
		allTemplates = append(allTemplates, githubTemplates...)
	}

	return allTemplates, nil
}

// LoadGitHubTemplatesCache loads previously cached GitHub templates
func (tm *TemplateManager) LoadGitHubTemplatesCache() ([]Template, error) {
	cacheFile := filepath.Join(tm.configDir, "github-templates-cache.yaml")

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		return []Template{}, nil
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var templates []Template
	if err := yaml.Unmarshal(data, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// SaveGitHubTemplatesCache saves GitHub templates to cache
func (tm *TemplateManager) SaveGitHubTemplatesCache(templates []Template) error {
	cacheDir := filepath.Join(tm.configDir, "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	cacheFile := filepath.Join(tm.configDir, "github-templates-cache.yaml")

	data, err := yaml.Marshal(templates)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// RefreshGitHubTemplatesCache refreshes the GitHub templates cache
func (tm *TemplateManager) RefreshGitHubTemplatesCache() error {
	tm.logger.Info("Refreshing GitHub templates cache...")

	// Search for popular NixOS configuration repositories
	templates, err := tm.searchGitHubTemplates("nixos configuration", 20)
	if err != nil {
		return fmt.Errorf("failed to search GitHub templates: %v", err)
	}

	// Save to cache
	if err := tm.SaveGitHubTemplatesCache(templates); err != nil {
		return fmt.Errorf("failed to save cache: %v", err)
	}

	tm.logger.Info(fmt.Sprintf("Cached %d GitHub templates", len(templates)))
	return nil
}

// SearchGitHubTemplates searches for templates on GitHub
func (tm *TemplateManager) searchGitHubTemplates(query string, limit int) ([]Template, error) {
	searchURL := fmt.Sprintf("https://api.github.com/search/code?q=%s+filename:configuration.nix&sort=indexed&order=desc&per_page=%d",
		url.QueryEscape(query), limit)

	resp, err := http.Get(searchURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var searchResponse GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, err
	}

	var templates []Template
	for _, item := range searchResponse.Items {
		template := Template{
			Name:        fmt.Sprintf("%s-%s", item.Repository.Owner.Login, item.Repository.Name),
			Description: item.Repository.Description,
			Category:    tm.categorizeFromDescription(item.Repository.Description),
			Tags:        tm.extractTagsFromDescription(item.Repository.Description),
			Source:      "github",
			GitHubRepo:  item.Repository.FullName,
			FilePath:    item.Path,
			Metadata: map[string]string{
				"stars":      fmt.Sprintf("%d", item.Repository.StarCount),
				"language":   item.Repository.Language,
				"updated_at": item.Repository.UpdatedAt,
				"github_url": item.HTMLURL,
			},
		}
		templates = append(templates, template)
	}

	return templates, nil
}

// LoadTemplateFromGitHub loads a template directly from GitHub
func (tm *TemplateManager) LoadTemplateFromGitHub(repo, filePath string) (*Template, error) {
	contentURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, filePath)

	resp, err := http.Get(contentURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var contentResponse GitHubContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&contentResponse); err != nil {
		return nil, err
	}

	// Decode base64 content
	content, err := tm.decodeGitHubContent(contentResponse.Content, contentResponse.Encoding)
	if err != nil {
		return nil, err
	}

	template := &Template{
		Name:       fmt.Sprintf("github-%s", strings.ReplaceAll(repo, "/", "-")),
		Source:     "github",
		GitHubRepo: repo,
		FilePath:   filePath,
		Content:    content,
		Metadata: map[string]string{
			"github_url": contentResponse.HTMLURL,
			"sha":        contentResponse.Sha,
		},
	}

	return template, nil
}

// ImportTemplateFromGitHub imports and saves a template from GitHub
func (tm *TemplateManager) ImportTemplateFromGitHub(repo, filePath, name, category string) error {
	template, err := tm.LoadTemplateFromGitHub(repo, filePath)
	if err != nil {
		return err
	}

	if name != "" {
		template.Name = name
	}
	if category != "" {
		template.Category = category
	}

	// Save as custom template
	return tm.SaveCustomTemplate(template)
}

// SaveCustomTemplate saves a template to the custom templates directory
func (tm *TemplateManager) SaveCustomTemplate(template *Template) error {
	templatesDir := filepath.Join(tm.configDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	templateFile := filepath.Join(templatesDir, template.Name+".yaml")

	data, err := yaml.Marshal(template)
	if err != nil {
		return err
	}

	return os.WriteFile(templateFile, data, 0644)
}

// ExportTemplate exports a template to a file
func (tm *TemplateManager) ExportTemplate(templateName, outputPath string) error {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return err
	}

	// Create export structure
	exportData := map[string]interface{}{
		"template":    template,
		"exported_at": time.Now().Format(time.RFC3339),
		"exported_by": "nixai",
		"version":     "1.0",
	}

	data, err := yaml.Marshal(exportData)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// ImportTemplateFromFile imports a template from an exported file
func (tm *TemplateManager) ImportTemplateFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var exportData map[string]interface{}
	if err := yaml.Unmarshal(data, &exportData); err != nil {
		return err
	}

	templateData, ok := exportData["template"]
	if !ok {
		return fmt.Errorf("invalid template export file")
	}

	// Convert back to Template struct
	templateBytes, err := yaml.Marshal(templateData)
	if err != nil {
		return err
	}

	var template Template
	if err := yaml.Unmarshal(templateBytes, &template); err != nil {
		return err
	}

	// Save as custom template
	return tm.SaveCustomTemplate(&template)
}

// ValidateTemplate validates a template's content and structure
func (tm *TemplateManager) ValidateTemplate(template *Template) error {
	// Basic validation
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Content == "" {
		return fmt.Errorf("template content is required")
	}

	// Validate NixOS configuration syntax
	if err := tm.validateNixConfiguration(template.Content); err != nil {
		return fmt.Errorf("invalid NixOS configuration: %v", err)
	}

	return nil
}

// GetTemplatesByCategory returns templates filtered by category
func (tm *TemplateManager) GetTemplatesByCategory(category string) ([]Template, error) {
	allTemplates, err := tm.LoadAllTemplates()
	if err != nil {
		return nil, err
	}

	var filtered []Template
	for _, template := range allTemplates {
		if strings.EqualFold(template.Category, category) {
			filtered = append(filtered, template)
		}
	}

	return filtered, nil
}

// GetTemplatesByTag returns templates filtered by tag
func (tm *TemplateManager) GetTemplatesByTag(tag string) ([]Template, error) {
	allTemplates, err := tm.LoadAllTemplates()
	if err != nil {
		return nil, err
	}

	var filtered []Template
	for _, template := range allTemplates {
		for _, templateTag := range template.Tags {
			if strings.EqualFold(templateTag, tag) {
				filtered = append(filtered, template)
				break
			}
		}
	}

	return filtered, nil
}

// SyncTemplates synchronizes templates with remote sources
func (tm *TemplateManager) SyncTemplates() error {
	tm.logger.Info("Synchronizing templates...")

	// Refresh GitHub cache
	if err := tm.RefreshGitHubTemplatesCache(); err != nil {
		tm.logger.Warn("Failed to refresh GitHub templates: " + err.Error())
	}

	// Could add other sync operations here (GitLab, custom registries, etc.)

	tm.logger.Info("Template synchronization completed")
	return nil
}

// Helper methods

func (tm *TemplateManager) categorizeFromDescription(description string) string {
	desc := strings.ToLower(description)

	if strings.Contains(desc, "desktop") || strings.Contains(desc, "gnome") || strings.Contains(desc, "kde") {
		return "Desktop"
	}
	if strings.Contains(desc, "server") || strings.Contains(desc, "headless") {
		return "Server"
	}
	if strings.Contains(desc, "gaming") || strings.Contains(desc, "steam") {
		return "Gaming"
	}
	if strings.Contains(desc, "development") || strings.Contains(desc, "dev") {
		return "Development"
	}
	if strings.Contains(desc, "minimal") || strings.Contains(desc, "minimal") {
		return "Minimal"
	}

	return "General"
}

func (tm *TemplateManager) extractTagsFromDescription(description string) []string {
	var tags []string
	desc := strings.ToLower(description)

	keywords := []string{"nixos", "flakes", "home-manager", "gnome", "kde", "i3", "sway",
		"docker", "kubernetes", "gaming", "development", "server", "minimal"}

	for _, keyword := range keywords {
		if strings.Contains(desc, keyword) {
			tags = append(tags, keyword)
		}
	}

	return tags
}

func (tm *TemplateManager) decodeGitHubContent(content, encoding string) (string, error) {
	if encoding == "base64" {
		// GitHub returns content base64 encoded
		decoded, err := utils.DecodeBase64(content)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}

	return content, nil
}

func (tm *TemplateManager) validateNixConfiguration(content string) error {
	// Basic NixOS syntax validation
	if !strings.Contains(content, "{") || !strings.Contains(content, "}") {
		return fmt.Errorf("configuration appears to be malformed")
	}

	// Check for basic NixOS structure
	if !strings.Contains(content, "config") && !strings.Contains(content, "pkgs") {
		return fmt.Errorf("doesn't appear to be a valid NixOS configuration")
	}

	return nil
}
