package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// CommandAvailability represents the availability status of a command
type CommandAvailability string

const (
	CommandAvailable    CommandAvailability = "available"     // Command is installed and available
	CommandUnavailable  CommandAvailability = "unavailable"   // Command not found
	CommandNixRunnable  CommandAvailability = "nix_runnable"  // Can be run via nix run
	CommandInstallable  CommandAvailability = "installable"   // Available in nixpkgs for installation
	CommandUnknown      CommandAvailability = "unknown"       // Cannot determine availability
)

// SystemContext represents different Nix system configurations
type SystemContext string

const (
	ContextNixOS       SystemContext = "nixos"        // NixOS system with configuration.nix
	ContextHomeManager SystemContext = "home_manager" // Home Manager setup
	ContextFlakes      SystemContext = "flakes"       // Flake-based system
	ContextProfile     SystemContext = "profile"      // User profile (nix profile)
	ContextDevelopment SystemContext = "development"  // Development environment
	ContextGeneric     SystemContext = "generic"      // Generic Nix installation
)

// InstallationOption represents a way to install a package permanently
type InstallationOption struct {
	Method      string `json:"method"`      // e.g., "NixOS Configuration", "Home Manager"
	Command     string `json:"command"`     // Command to run
	ConfigFile  string `json:"config_file"` // Which file to edit (e.g., configuration.nix)
	ConfigSnippet string `json:"config_snippet"` // Code snippet to add
	Description string `json:"description"` // Human-readable description
	Recommended bool   `json:"recommended"` // Whether this is the recommended approach
}

// CommandResolution contains information about how to resolve a command
type CommandResolution struct {
	Command             string                `json:"command"`
	Availability        CommandAvailability   `json:"availability"`
	NixPackage          string                `json:"nix_package,omitempty"`      // nixpkgs package name
	NixRunCommand       string                `json:"nix_run_command,omitempty"`  // Full nix run command
	InstallationOptions []InstallationOption  `json:"installation_options,omitempty"` // Multiple installation methods
	Description         string                `json:"description,omitempty"`      // Package description
	Suggestions         []string              `json:"suggestions,omitempty"`      // Alternative commands or packages
	EstimatedSize       string                `json:"estimated_size,omitempty"`   // Estimated download size
	SystemContext       SystemContext         `json:"system_context,omitempty"`  // Detected system context
	LastChecked         time.Time             `json:"last_checked"`               // When this was last verified
}

// NixpkgsSearchResult represents a search result from nixpkgs
type NixpkgsSearchResult struct {
	Package     string `json:"package"`
	Pname       string `json:"pname"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Programs    []string `json:"programs,omitempty"`
}

// CommandResolver handles command availability detection and nix run integration
type CommandResolver struct {
	logger        *logger.Logger
	cache         map[string]*CommandResolution // In-memory cache
	cacheTimeout  time.Duration                  // How long to cache results
	nixpkgsIndex  map[string][]NixpkgsSearchResult // Local nixpkgs program index
	systemContext SystemContext                  // Detected system context
}

// NewCommandResolver creates a new command resolver
func NewCommandResolver(log *logger.Logger) *CommandResolver {
	cr := &CommandResolver{
		logger:       log,
		cache:        make(map[string]*CommandResolution),
		cacheTimeout: time.Hour * 24, // Cache for 24 hours
		nixpkgsIndex: make(map[string][]NixpkgsSearchResult),
	}
	
	// Detect system context
	cr.systemContext = cr.detectSystemContext()
	log.Debug(fmt.Sprintf("Detected system context: %s", cr.systemContext))
	
	return cr
}

// ResolveCommand determines how to execute a command, using nix run if needed
func (cr *CommandResolver) ResolveCommand(ctx context.Context, command string) (*CommandResolution, error) {
	// Check cache first
	if cached, exists := cr.cache[command]; exists {
		if time.Since(cached.LastChecked) < cr.cacheTimeout {
			cr.logger.Debug(fmt.Sprintf("Using cached resolution for command: %s", command))
			return cached, nil
		}
	}

	cr.logger.Debug(fmt.Sprintf("Resolving command availability: %s", command))
	
	resolution := &CommandResolution{
		Command:     command,
		LastChecked: time.Now(),
	}

	// Step 1: Check if command is already available
	if cr.isCommandAvailable(command) {
		resolution.Availability = CommandAvailable
		cr.logger.Debug(fmt.Sprintf("Command %s is already available", command))
		cr.cache[command] = resolution
		return resolution, nil
	}

	// Step 2: Search for the command in nixpkgs
	nixpkgResult, err := cr.searchNixpkgs(ctx, command)
	if err != nil {
		cr.logger.Warn(fmt.Sprintf("Failed to search nixpkgs for %s: %v", command, err))
		resolution.Availability = CommandUnknown
		cr.cache[command] = resolution
		return resolution, nil
	}

	if nixpkgResult != nil {
		resolution.Availability = CommandNixRunnable
		resolution.NixPackage = nixpkgResult.Package
		resolution.Description = nixpkgResult.Description
		resolution.NixRunCommand = fmt.Sprintf("nix run nixpkgs#%s", nixpkgResult.Package)
		resolution.SystemContext = cr.systemContext
		
		// Generate appropriate installation options based on system context
		resolution.InstallationOptions = cr.generateInstallationOptions(nixpkgResult.Package)
		
		// Get estimated size if possible
		if size, err := cr.getPackageSize(ctx, nixpkgResult.Package); err == nil {
			resolution.EstimatedSize = size
		}

		cr.logger.Info(fmt.Sprintf("Found nixpkg for %s: %s", command, nixpkgResult.Package))
	} else {
		resolution.Availability = CommandUnavailable
		
		// Try to suggest similar commands
		suggestions := cr.findSimilarCommands(command)
		if len(suggestions) > 0 {
			resolution.Suggestions = suggestions
		}
		
		cr.logger.Debug(fmt.Sprintf("Command %s not found in nixpkgs", command))
	}

	cr.cache[command] = resolution
	return resolution, nil
}

// ResolveCommandString converts a command resolution to an executable command string
func (cr *CommandResolver) ResolveCommandString(resolution *CommandResolution, args []string) string {
	switch resolution.Availability {
	case CommandAvailable:
		// Command is available, use as-is
		if len(args) > 0 {
			return fmt.Sprintf("%s %s", resolution.Command, strings.Join(args, " "))
		}
		return resolution.Command
		
	case CommandNixRunnable:
		// Use nix run
		if len(args) > 0 {
			return fmt.Sprintf("%s -- %s", resolution.NixRunCommand, strings.Join(args, " "))
		}
		return resolution.NixRunCommand
		
	default:
		// Fallback to original command (will likely fail)
		if len(args) > 0 {
			return fmt.Sprintf("%s %s", resolution.Command, strings.Join(args, " "))
		}
		return resolution.Command
	}
}

// GetExecutionSuggestion generates a human-readable suggestion for command execution
func (cr *CommandResolver) GetExecutionSuggestion(resolution *CommandResolution) string {
	switch resolution.Availability {
	case CommandAvailable:
		return fmt.Sprintf("✅ Command '%s' is available and ready to run", resolution.Command)
		
	case CommandNixRunnable:
		suggestion := fmt.Sprintf("🚀 Command '%s' can be run temporarily using:\n", resolution.Command)
		suggestion += fmt.Sprintf("   %s\n", resolution.NixRunCommand)
		if resolution.Description != "" {
			suggestion += fmt.Sprintf("   📝 %s\n", resolution.Description)
		}
		if resolution.EstimatedSize != "" {
			suggestion += fmt.Sprintf("   💾 Estimated download: %s\n", resolution.EstimatedSize)
		}
		
		// Add installation options
		if len(resolution.InstallationOptions) > 0 {
			suggestion += "\n📦 Installation Options:\n"
			for i, option := range resolution.InstallationOptions {
				prefix := "   "
				if option.Recommended {
					prefix = "⭐ "
				}
				suggestion += fmt.Sprintf("%s%d. %s:\n", prefix, i+1, option.Method)
				suggestion += fmt.Sprintf("      %s\n", option.Command)
				suggestion += fmt.Sprintf("      📄 Edit: %s\n", option.ConfigFile)
				if option.ConfigSnippet != "" {
					suggestion += fmt.Sprintf("      💡 Add: %s\n", option.ConfigSnippet)
				}
				suggestion += fmt.Sprintf("      ℹ️  %s\n", option.Description)
				if i < len(resolution.InstallationOptions)-1 {
					suggestion += "\n"
				}
			}
		}
		return suggestion
		
	case CommandInstallable:
		suggestion := fmt.Sprintf("📦 Command '%s' is available for installation", resolution.Command)
		if len(resolution.InstallationOptions) > 0 && resolution.InstallationOptions[0].Command != "" {
			suggestion += fmt.Sprintf(":\n   %s", resolution.InstallationOptions[0].Command)
		}
		return suggestion
			
	case CommandUnavailable:
		suggestion := fmt.Sprintf("❌ Command '%s' is not available", resolution.Command)
		if len(resolution.Suggestions) > 0 {
			suggestion += "\n💡 Similar commands you might want:\n"
			for _, sug := range resolution.Suggestions {
				suggestion += fmt.Sprintf("   - %s\n", sug)
			}
		}
		return suggestion
		
	default:
		return fmt.Sprintf("❓ Unable to determine availability of '%s'", resolution.Command)
	}
}

// isCommandAvailable checks if a command is already available in PATH
func (cr *CommandResolver) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// searchNixpkgs searches for a command in nixpkgs
func (cr *CommandResolver) searchNixpkgs(ctx context.Context, command string) (*NixpkgsSearchResult, error) {
	// Try multiple search strategies
	
	// Strategy 1: Direct package name search
	if result := cr.searchNixpkgsDirectly(ctx, command); result != nil {
		return result, nil
	}
	
	// Strategy 2: Search by program name in common packages
	if result := cr.searchByProgramName(ctx, command); result != nil {
		return result, nil
	}
	
	// Strategy 3: Use nix search (if available)
	if result := cr.nixSearch(ctx, command); result != nil {
		return result, nil
	}
	
	return nil, nil
}

// searchNixpkgsDirectly tries direct package name matching
func (cr *CommandResolver) searchNixpkgsDirectly(ctx context.Context, command string) *NixpkgsSearchResult {
	// Common command to package mappings
	packageMappings := map[string]string{
		"git":      "git",
		"curl":     "curl",
		"wget":     "wget",
		"jq":       "jq",
		"tree":     "tree",
		"htop":     "htop",
		"vim":      "vim",
		"emacs":    "emacs",
		"firefox":  "firefox",
		"docker":   "docker",
		"kubectl":  "kubectl",
		"terraform": "terraform",
		"ansible":  "ansible",
		"python":   "python3",
		"python3":  "python3",
		"node":     "nodejs",
		"npm":      "nodejs", // npm comes with nodejs
		"yarn":     "yarn",
		"go":       "go",
		"rust":     "rustc",
		"cargo":    "cargo",
		"gcc":      "gcc",
		"make":     "gnumake",
		"cmake":    "cmake",
		"ninja":    "ninja",
		"tmux":     "tmux",
		"screen":   "screen",
		"zsh":      "zsh",
		"fish":     "fish",
		"ripgrep":  "ripgrep",
		"rg":       "ripgrep",
		"fd":       "fd",
		"bat":      "bat",
		"exa":      "exa",
		"fzf":      "fzf",
		"ag":       "silver-searcher",
		"ack":      "ack",
		"rsync":    "rsync",
		"ssh":      "openssh",
		"scp":      "openssh",
		"zip":      "zip",
		"unzip":    "unzip",
		"tar":      "gnutar",
		"gzip":     "gzip",
		"bzip2":    "bzip2",
		"xz":       "xz",
		"7z":       "p7zip",
	}
	
	if packageName, exists := packageMappings[command]; exists {
		return &NixpkgsSearchResult{
			Package:     packageName,
			Pname:       packageName,
			Description: fmt.Sprintf("Package providing %s command", command),
			Programs:    []string{command},
		}
	}
	
	return nil
}

// searchByProgramName searches for packages that provide a specific program
func (cr *CommandResolver) searchByProgramName(ctx context.Context, command string) *NixpkgsSearchResult {
	// This could be expanded to use a pre-built index of programs to packages
	// For now, we'll use some heuristics for common cases
	
	// Try the command name as package name
	if cr.checkPackageExists(ctx, command) {
		return &NixpkgsSearchResult{
			Package:     command,
			Pname:       command,
			Description: fmt.Sprintf("Package %s", command),
			Programs:    []string{command},
		}
	}
	
	return nil
}

// nixSearch uses nix search if available
func (cr *CommandResolver) nixSearch(ctx context.Context, command string) *NixpkgsSearchResult {
	// Check if nix command is available
	if !cr.isCommandAvailable("nix") {
		return nil
	}
	
	// Use nix search to find packages
	cmd := exec.CommandContext(ctx, "nix", "search", "nixpkgs", command, "--json")
	output, err := cmd.Output()
	if err != nil {
		cr.logger.Debug(fmt.Sprintf("nix search failed for %s: %v", command, err))
		return nil
	}
	
	// Parse the JSON output
	var searchResults map[string]interface{}
	if err := json.Unmarshal(output, &searchResults); err != nil {
		cr.logger.Debug(fmt.Sprintf("Failed to parse nix search output: %v", err))
		return nil
	}
	
	// Find the best match
	for packageName, packageInfo := range searchResults {
		if info, ok := packageInfo.(map[string]interface{}); ok {
			// Extract package name from attribute path
			parts := strings.Split(packageName, ".")
			if len(parts) > 0 {
				actualPackage := parts[len(parts)-1]
				
				description := ""
				if desc, exists := info["description"]; exists {
					if descStr, ok := desc.(string); ok {
						description = descStr
					}
				}
				
				return &NixpkgsSearchResult{
					Package:     actualPackage,
					Pname:       actualPackage,
					Description: description,
					Programs:    []string{command},
				}
			}
		}
	}
	
	return nil
}

// checkPackageExists checks if a package exists in nixpkgs
func (cr *CommandResolver) checkPackageExists(ctx context.Context, packageName string) bool {
	if !cr.isCommandAvailable("nix") {
		return false
	}
	
	// Try to evaluate the package attribute
	cmd := exec.CommandContext(ctx, "nix", "eval", "--raw", fmt.Sprintf("nixpkgs#%s.name", packageName))
	err := cmd.Run()
	return err == nil
}

// getPackageSize estimates the download size of a package
func (cr *CommandResolver) getPackageSize(ctx context.Context, packageName string) (string, error) {
	if !cr.isCommandAvailable("nix") {
		return "", fmt.Errorf("nix command not available")
	}
	
	// Use nix path-info to get size information
	cmd := exec.CommandContext(ctx, "nix", "path-info", "-S", fmt.Sprintf("nixpkgs#%s", packageName))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	// Parse the output to extract size
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		// Look for size information in the output
		sizeRegex := regexp.MustCompile(`(\d+(?:\.\d+)?[KMGT]?B?)`)
		if matches := sizeRegex.FindStringSubmatch(lines[0]); len(matches) > 1 {
			return matches[1], nil
		}
	}
	
	return "", fmt.Errorf("could not determine size")
}

// findSimilarCommands suggests similar commands when one is not found
func (cr *CommandResolver) findSimilarCommands(command string) []string {
	// This could be enhanced with more sophisticated similarity matching
	suggestions := []string{}
	
	// Common typos and alternatives
	alternatives := map[string][]string{
		"ls":     {"exa", "tree"},
		"cat":    {"bat", "less"},
		"grep":   {"ripgrep", "ag", "ack"},
		"find":   {"fd", "fzf"},
		"top":    {"htop", "btop"},
		"ps":     {"htop", "procs"},
		"vi":     {"vim", "nvim", "emacs"},
		"nano":   {"vim", "micro"},
		"curl":   {"wget", "httpie"},
		"python": {"python3"},
		"pip":    {"python3Packages.pip"},
	}
	
	if alts, exists := alternatives[command]; exists {
		suggestions = append(suggestions, alts...)
	}
	
	// Add some intelligent suggestions based on command patterns
	if strings.Contains(command, "install") {
		suggestions = append(suggestions, "nix-env -iA", "nix profile install")
	}
	
	if strings.Contains(command, "search") {
		suggestions = append(suggestions, "nix search", "nix-env -qaP")
	}
	
	return suggestions
}

// ClearCache clears the resolution cache
func (cr *CommandResolver) ClearCache() {
	cr.cache = make(map[string]*CommandResolution)
	cr.logger.Debug("Command resolution cache cleared")
}

// GetCacheStats returns cache statistics
func (cr *CommandResolver) GetCacheStats() map[string]interface{} {
	available := 0
	nixRunnable := 0
	unavailable := 0
	
	for _, resolution := range cr.cache {
		switch resolution.Availability {
		case CommandAvailable:
			available++
		case CommandNixRunnable:
			nixRunnable++
		case CommandUnavailable:
			unavailable++
		}
	}
	
	return map[string]interface{}{
		"total_cached":   len(cr.cache),
		"available":      available,
		"nix_runnable":   nixRunnable,
		"unavailable":    unavailable,
		"cache_timeout":  cr.cacheTimeout.String(),
		"system_context": string(cr.systemContext),
	}
}

// detectSystemContext determines the current Nix system configuration
func (cr *CommandResolver) detectSystemContext() SystemContext {
	// Check for NixOS
	if cr.fileExists("/etc/nixos/configuration.nix") {
		// Check if it's flake-based NixOS
		if cr.fileExists("/etc/nixos/flake.nix") {
			return ContextFlakes
		}
		return ContextNixOS
	}
	
	// Check for Home Manager
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		hmConfigPath := filepath.Join(homeDir, ".config/nixpkgs/home.nix")
		hmConfigPath2 := filepath.Join(homeDir, ".config/home-manager/home.nix")
		
		if cr.fileExists(hmConfigPath) || cr.fileExists(hmConfigPath2) {
			// Check if it's flake-based Home Manager
			hmFlakePath := filepath.Join(homeDir, ".config/home-manager/flake.nix")
			if cr.fileExists(hmFlakePath) {
				return ContextFlakes
			}
			return ContextHomeManager
		}
	}
	
	// Check for flake in current directory or parent directories
	if cr.findFlakeNix() {
		return ContextFlakes
	}
	
	// Check if in a development environment (shell.nix or flake with devShell)
	if cr.fileExists("shell.nix") || cr.fileExists("default.nix") {
		return ContextDevelopment
	}
	
	// Check if nix profile is available (modern nix)
	if cr.isCommandAvailable("nix") {
		return ContextProfile
	}
	
	return ContextGeneric
}

// generateInstallationOptions creates context-appropriate installation methods
func (cr *CommandResolver) generateInstallationOptions(packageName string) []InstallationOption {
	var options []InstallationOption
	
	switch cr.systemContext {
	case ContextNixOS:
		options = append(options, InstallationOption{
			Method:      "NixOS Configuration",
			Command:     "sudo nixos-rebuild switch",
			ConfigFile:  "/etc/nixos/configuration.nix",
			ConfigSnippet: fmt.Sprintf("environment.systemPackages = with pkgs; [\n  %s\n];", packageName),
			Description: "Add to system packages in NixOS configuration (recommended for system-wide installation)",
			Recommended: true,
		})
		
		// Also offer Home Manager if available
		if cr.fileExists(filepath.Join(os.Getenv("HOME"), ".config/nixpkgs/home.nix")) {
			options = append(options, InstallationOption{
				Method:      "Home Manager",
				Command:     "home-manager switch",
				ConfigFile:  "~/.config/nixpkgs/home.nix",
				ConfigSnippet: fmt.Sprintf("home.packages = with pkgs; [\n  %s\n];", packageName),
				Description: "Add to user packages via Home Manager",
				Recommended: false,
			})
		}
		
	case ContextHomeManager:
		options = append(options, InstallationOption{
			Method:      "Home Manager",
			Command:     "home-manager switch",
			ConfigFile:  "~/.config/nixpkgs/home.nix",
			ConfigSnippet: fmt.Sprintf("home.packages = with pkgs; [\n  %s\n];", packageName),
			Description: "Add to user packages in Home Manager configuration (recommended)",
			Recommended: true,
		})
		
	case ContextFlakes:
		options = append(options, InstallationOption{
			Method:      "Flake Configuration",
			Command:     "nixos-rebuild switch --flake .",
			ConfigFile:  "flake.nix",
			ConfigSnippet: fmt.Sprintf("# Add to packages in your flake.nix:\npackages = with pkgs; [ %s ];", packageName),
			Description: "Add to flake packages (recommended for flake-based systems)",
			Recommended: true,
		})
		
		options = append(options, InstallationOption{
			Method:      "Flake Development Shell",
			Command:     "nix develop",
			ConfigFile:  "flake.nix",
			ConfigSnippet: fmt.Sprintf("devShells.default = pkgs.mkShell {\n  buildInputs = with pkgs; [ %s ];\n};", packageName),
			Description: "Add to development shell for this project",
			Recommended: false,
		})
		
	case ContextDevelopment:
		options = append(options, InstallationOption{
			Method:      "Development Shell",
			Command:     "nix-shell",
			ConfigFile:  "shell.nix",
			ConfigSnippet: fmt.Sprintf("{ pkgs ? import <nixpkgs> {} }:\npkgs.mkShell {\n  buildInputs = with pkgs; [ %s ];\n}", packageName),
			Description: "Add to development shell (recommended for project-specific tools)",
			Recommended: true,
		})
		
	case ContextProfile:
		options = append(options, InstallationOption{
			Method:      "Nix Profile",
			Command:     fmt.Sprintf("nix profile install nixpkgs#%s", packageName),
			ConfigFile:  "~/.nix-profile",
			ConfigSnippet: fmt.Sprintf("# Managed via nix profile\n# To remove: nix profile remove %s", packageName),
			Description: "Install to user profile with modern nix commands (recommended)",
			Recommended: true,
		})
		
	default: // ContextGeneric
		options = append(options, InstallationOption{
			Method:      "Nix Profile",
			Command:     fmt.Sprintf("nix profile install nixpkgs#%s", packageName),
			ConfigFile:  "~/.nix-profile",
			ConfigSnippet: fmt.Sprintf("# Managed via nix profile\n# To remove: nix profile remove %s", packageName),
			Description: "Install to user profile (recommended)",
			Recommended: true,
		})
		
		// Fallback to legacy nix-env
		options = append(options, InstallationOption{
			Method:      "Legacy Nix Env",
			Command:     fmt.Sprintf("nix-env -iA nixpkgs.%s", packageName),
			ConfigFile:  "~/.nix-profile",
			ConfigSnippet: fmt.Sprintf("# Installed via nix-env\n# To remove: nix-env -e %s", packageName),
			Description: "Install with legacy nix-env command",
			Recommended: false,
		})
	}
	
	return options
}

// fileExists checks if a file exists
func (cr *CommandResolver) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// findFlakeNix searches for flake.nix in current and parent directories
func (cr *CommandResolver) findFlakeNix() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	
	// Check current directory and up to 3 parent directories
	for i := 0; i < 4; i++ {
		flakePath := filepath.Join(cwd, "flake.nix")
		if cr.fileExists(flakePath) {
			return true
		}
		
		parentDir := filepath.Dir(cwd)
		if parentDir == cwd {
			break // Reached filesystem root
		}
		cwd = parentDir
	}
	
	return false
}