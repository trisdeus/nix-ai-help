package ide

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/dev"
	"nix-ai-help/pkg/logger"
)

// Manager implements the IDEManager interface
type Manager struct {
	logger          *logger.Logger
	ideIntegrations map[string]*dev.IDEIntegration
}

// NewManager creates a new IDE manager
func NewManager(logger *logger.Logger) *Manager {
	m := &Manager{
		logger:          logger,
		ideIntegrations: make(map[string]*dev.IDEIntegration),
	}
	
	// Initialize built-in IDE integrations
	m.initializeBuiltinIntegrations()
	
	return m
}

// SetupIDE sets up IDE integration for a development environment
func (m *Manager) SetupIDE(ctx context.Context, env *dev.DevEnvironment, ide string) error {
	integration, exists := m.ideIntegrations[ide]
	if !exists {
		return fmt.Errorf("IDE %s not supported", ide)
	}
	
	// Create IDE configuration files
	for _, configFile := range integration.ConfigFiles {
		if err := m.createConfigFile(env, configFile, integration); err != nil {
			return fmt.Errorf("failed to create config file %s: %w", configFile, err)
		}
	}
	
	// Create IDE-specific directories
	if err := m.createIDEDirectories(env, integration); err != nil {
		return fmt.Errorf("failed to create IDE directories: %w", err)
	}
	
	m.logger.Info(fmt.Sprintf("IDE integration setup completed: %s for %s", ide, env.Name))
	return nil
}

// GetIDEConfig retrieves IDE configuration
func (m *Manager) GetIDEConfig(ctx context.Context, ide string) (*dev.IDEIntegration, error) {
	integration, exists := m.ideIntegrations[ide]
	if !exists {
		return nil, fmt.Errorf("IDE %s not supported", ide)
	}
	
	return integration, nil
}

// ListSupportedIDEs lists all supported IDEs
func (m *Manager) ListSupportedIDEs(ctx context.Context) ([]string, error) {
	var ides []string
	for ide := range m.ideIntegrations {
		ides = append(ides, ide)
	}
	return ides, nil
}

// InstallExtensions installs extensions for an IDE
func (m *Manager) InstallExtensions(ctx context.Context, ide string, extensions []string) error {
	_, exists := m.ideIntegrations[ide]
	if !exists {
		return fmt.Errorf("IDE %s not supported", ide)
	}
	
	switch ide {
	case "vscode":
		return m.installVSCodeExtensions(extensions)
	case "neovim":
		return m.installNeovimPlugins(extensions)
	case "vim":
		return m.installVimPlugins(extensions)
	case "emacs":
		return m.installEmacsPackages(extensions)
	default:
		m.logger.Info(fmt.Sprintf("Extension installation not implemented for IDE: %s", ide))
		return nil
	}
}

// initializeBuiltinIntegrations sets up built-in IDE integrations
func (m *Manager) initializeBuiltinIntegrations() {
	// VS Code integration
	m.ideIntegrations["vscode"] = &dev.IDEIntegration{
		Name:        "Visual Studio Code",
		Type:        "editor",
		ConfigFiles: []string{".vscode/settings.json", ".vscode/launch.json", ".vscode/tasks.json"},
		Extensions: []string{
			"ms-vscode.vscode-json",
			"ms-vscode.cmake-tools",
			"ms-python.python",
			"golang.go",
			"rust-lang.rust-analyzer",
			"ms-vscode.cpptools",
			"ms-dotnettools.csharp",
			"redhat.java",
			"ms-vscode.vscode-typescript-next",
		},
		Settings: map[string]interface{}{
			"editor.formatOnSave": true,
			"editor.tabSize":      2,
			"files.trimTrailingWhitespace": true,
			"files.insertFinalNewline":     true,
		},
		LaunchConfig: map[string]interface{}{
			"version": "0.2.0",
			"configurations": []map[string]interface{}{
				{
					"name":    "Launch Program",
					"type":    "go",
					"request": "launch",
					"mode":    "auto",
					"program": "${workspaceFolder}",
				},
			},
		},
	}
	
	// Neovim integration
	m.ideIntegrations["neovim"] = &dev.IDEIntegration{
		Name:        "Neovim",
		Type:        "editor",
		ConfigFiles: []string{".nvim/init.lua", ".nvim/settings.lua"},
		Extensions: []string{
			"nvim-lspconfig",
			"nvim-treesitter",
			"telescope.nvim",
			"nvim-cmp",
			"null-ls.nvim",
			"gitsigns.nvim",
		},
		Settings: map[string]interface{}{
			"number":         true,
			"relativenumber": true,
			"tabstop":        2,
			"shiftwidth":     2,
			"expandtab":      true,
		},
	}
	
	// Vim integration
	m.ideIntegrations["vim"] = &dev.IDEIntegration{
		Name:        "Vim",
		Type:        "editor",
		ConfigFiles: []string{".vimrc"},
		Extensions: []string{
			"vim-plug",
			"nerdtree",
			"vim-airline",
			"vim-fugitive",
			"vim-gitgutter",
		},
		Settings: map[string]interface{}{
			"number":     true,
			"tabstop":    2,
			"shiftwidth": 2,
			"expandtab":  true,
		},
	}
	
	// Emacs integration
	m.ideIntegrations["emacs"] = &dev.IDEIntegration{
		Name:        "Emacs",
		Type:        "editor",
		ConfigFiles: []string{".emacs.d/init.el"},
		Extensions: []string{
			"use-package",
			"company",
			"flycheck",
			"magit",
			"projectile",
			"helm",
		},
		Settings: map[string]interface{}{
			"indent-tabs-mode": false,
			"tab-width":        2,
		},
	}
	
	// IntelliJ IDEA integration
	m.ideIntegrations["intellij"] = &dev.IDEIntegration{
		Name:        "IntelliJ IDEA",
		Type:        "ide",
		ConfigFiles: []string{".idea/workspace.xml", ".idea/modules.xml"},
		Extensions:  []string{},
		Settings: map[string]interface{}{
			"code_style": map[string]interface{}{
				"indent_size": 2,
				"tab_size":    2,
			},
		},
	}
	
	// Eclipse integration
	m.ideIntegrations["eclipse"] = &dev.IDEIntegration{
		Name:        "Eclipse",
		Type:        "ide",
		ConfigFiles: []string{".project", ".classpath"},
		Extensions:  []string{},
		Settings: map[string]interface{}{
			"editor": map[string]interface{}{
				"tab_size":    2,
				"indent_size": 2,
			},
		},
	}
}

// createConfigFile creates a configuration file for an IDE
func (m *Manager) createConfigFile(env *dev.DevEnvironment, configFile string, integration *dev.IDEIntegration) error {
	filePath := filepath.Join(env.Path, configFile)
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Generate configuration content based on file type
	var content string
	var err error
	
	switch {
	case strings.HasSuffix(configFile, ".json"):
		content, err = m.generateJSONConfig(env, integration, configFile)
	case strings.HasSuffix(configFile, ".lua"):
		content, err = m.generateLuaConfig(env, integration, configFile)
	case strings.HasSuffix(configFile, ".vimrc"):
		content, err = m.generateVimConfig(env, integration)
	case strings.HasSuffix(configFile, ".el"):
		content, err = m.generateEmacsConfig(env, integration)
	case strings.HasSuffix(configFile, ".xml"):
		content, err = m.generateXMLConfig(env, integration, configFile)
	default:
		content, err = m.generateGenericConfig(env, integration, configFile)
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate config content: %w", err)
	}
	
	// Write configuration file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	m.logger.Info(fmt.Sprintf("Created IDE config file: %s at %s", configFile, filePath))
	return nil
}

// generateJSONConfig generates JSON configuration files
func (m *Manager) generateJSONConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration, configFile string) (string, error) {
	var config interface{}
	
	switch {
	case strings.Contains(configFile, "settings.json"):
		config = m.generateVSCodeSettings(env, integration)
	case strings.Contains(configFile, "launch.json"):
		config = m.generateVSCodeLaunchConfig(env, integration)
	case strings.Contains(configFile, "tasks.json"):
		config = m.generateVSCodeTasks(env, integration)
	default:
		config = integration.Settings
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// generateVSCodeSettings generates VS Code settings
func (m *Manager) generateVSCodeSettings(env *dev.DevEnvironment, integration *dev.IDEIntegration) map[string]interface{} {
	settings := make(map[string]interface{})
	
	// Copy base settings
	for key, value := range integration.Settings {
		settings[key] = value
	}
	
	// Add language-specific settings
	switch env.Language {
	case "go":
		settings["go.useLanguageServer"] = true
		settings["go.formatTool"] = "goimports"
		settings["go.lintTool"] = "golint"
	case "rust":
		settings["rust-analyzer.checkOnSave.command"] = "cargo check"
		settings["rust-analyzer.cargo.allFeatures"] = true
	case "python":
		settings["python.defaultInterpreterPath"] = "python3"
		settings["python.formatting.provider"] = "black"
		settings["python.linting.enabled"] = true
		settings["python.linting.pylintEnabled"] = true
	case "typescript", "javascript":
		settings["typescript.preferences.importModuleSpecifier"] = "relative"
		settings["javascript.preferences.importModuleSpecifier"] = "relative"
	}
	
	return settings
}

// generateVSCodeLaunchConfig generates VS Code launch configuration
func (m *Manager) generateVSCodeLaunchConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration) map[string]interface{} {
	config := map[string]interface{}{
		"version":        "0.2.0",
		"configurations": []map[string]interface{}{},
	}
	
	var configurations []map[string]interface{}
	
	switch env.Language {
	case "go":
		configurations = append(configurations, map[string]interface{}{
			"name":    "Launch Go Program",
			"type":    "go",
			"request": "launch",
			"mode":    "auto",
			"program": "${workspaceFolder}",
		})
	case "rust":
		configurations = append(configurations, map[string]interface{}{
			"name":    "Launch Rust Program",
			"type":    "lldb",
			"request": "launch",
			"program": "${workspaceFolder}/target/debug/${workspaceFolderBasename}",
			"args":    []string{},
		})
	case "python":
		configurations = append(configurations, map[string]interface{}{
			"name":    "Launch Python Program",
			"type":    "python",
			"request": "launch",
			"program": "${workspaceFolder}/main.py",
		})
	case "typescript", "javascript":
		configurations = append(configurations, map[string]interface{}{
			"name":    "Launch Node.js Program",
			"type":    "node",
			"request": "launch",
			"program": "${workspaceFolder}/src/index.js",
		})
	}
	
	config["configurations"] = configurations
	return config
}

// generateVSCodeTasks generates VS Code tasks
func (m *Manager) generateVSCodeTasks(env *dev.DevEnvironment, integration *dev.IDEIntegration) map[string]interface{} {
	config := map[string]interface{}{
		"version": "2.0.0",
		"tasks":   []map[string]interface{}{},
	}
	
	var tasks []map[string]interface{}
	
	switch env.Language {
	case "go":
		tasks = append(tasks, map[string]interface{}{
			"label":   "go build",
			"type":    "shell",
			"command": "go",
			"args":    []string{"build", "-v", "./..."},
			"group":   "build",
		})
		tasks = append(tasks, map[string]interface{}{
			"label":   "go test",
			"type":    "shell",
			"command": "go",
			"args":    []string{"test", "-v", "./..."},
			"group":   "test",
		})
	case "rust":
		tasks = append(tasks, map[string]interface{}{
			"label":   "cargo build",
			"type":    "shell",
			"command": "cargo",
			"args":    []string{"build"},
			"group":   "build",
		})
		tasks = append(tasks, map[string]interface{}{
			"label":   "cargo test",
			"type":    "shell",
			"command": "cargo",
			"args":    []string{"test"},
			"group":   "test",
		})
	case "python":
		tasks = append(tasks, map[string]interface{}{
			"label":   "python run",
			"type":    "shell",
			"command": "python",
			"args":    []string{"main.py"},
			"group":   "build",
		})
	}
	
	config["tasks"] = tasks
	return config
}

// generateLuaConfig generates Lua configuration files for Neovim
func (m *Manager) generateLuaConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration, configFile string) (string, error) {
	if strings.Contains(configFile, "init.lua") {
		return m.generateNeovimInitLua(env, integration), nil
	}
	
	return "-- Generated configuration for " + env.Name + "\n", nil
}

// generateNeovimInitLua generates Neovim init.lua configuration
func (m *Manager) generateNeovimInitLua(env *dev.DevEnvironment, integration *dev.IDEIntegration) string {
	config := `-- Generated configuration for ` + env.Name + `

-- Basic settings
vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.tabstop = 2
vim.opt.shiftwidth = 2
vim.opt.expandtab = true
vim.opt.smartindent = true
vim.opt.wrap = false

-- Plugin management with packer.nvim
local ensure_packer = function()
  local fn = vim.fn
  local install_path = fn.stdpath('data')..'/site/pack/packer/start/packer.nvim'
  if fn.empty(fn.glob(install_path)) > 0 then
    fn.system({'git', 'clone', '--depth', '1', 'https://github.com/wbthomason/packer.nvim', install_path})
    vim.cmd [[packadd packer.nvim]]
    return true
  end
  return false
end

local packer_bootstrap = ensure_packer()

require('packer').startup(function(use)
  use 'wbthomason/packer.nvim'
  use 'neovim/nvim-lspconfig'
  use 'nvim-treesitter/nvim-treesitter'
  use 'nvim-telescope/telescope.nvim'
  use 'hrsh7th/nvim-cmp'
  use 'jose-elias-alvarez/null-ls.nvim'
  use 'lewis6991/gitsigns.nvim'
  
  if packer_bootstrap then
    require('packer').sync()
  end
end)
`
	
	// Add language-specific configuration
	switch env.Language {
	case "go":
		config += `
-- Go-specific configuration
require('lspconfig').gopls.setup{}
`
	case "rust":
		config += `
-- Rust-specific configuration
require('lspconfig').rust_analyzer.setup{}
`
	case "python":
		config += `
-- Python-specific configuration
require('lspconfig').pyright.setup{}
`
	case "typescript", "javascript":
		config += `
-- TypeScript/JavaScript-specific configuration
require('lspconfig').tsserver.setup{}
`
	}
	
	return config
}

// generateVimConfig generates Vim configuration
func (m *Manager) generateVimConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration) (string, error) {
	config := `" Generated configuration for ` + env.Name + `

" Basic settings
set number
set tabstop=2
set shiftwidth=2
set expandtab
set smartindent
set nowrap

" Plugin management with vim-plug
call plug#begin('~/.vim/plugged')
Plug 'scrooloose/nerdtree'
Plug 'vim-airline/vim-airline'
Plug 'tpope/vim-fugitive'
Plug 'airblade/vim-gitgutter'
call plug#end()

" Key mappings
nnoremap <C-n> :NERDTreeToggle<CR>
`
	
	return config, nil
}

// generateEmacsConfig generates Emacs configuration
func (m *Manager) generateEmacsConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration) (string, error) {
	config := `;; Generated configuration for ` + env.Name + `

;; Basic settings
(setq-default indent-tabs-mode nil)
(setq-default tab-width 2)
(setq backup-directory-alist '(("." . "~/.emacs.d/backups")))

;; Package management
(require 'package)
(setq package-archives '(("melpa" . "https://melpa.org/packages/")
                         ("org" . "https://orgmode.org/elpa/")
                         ("elpa" . "https://elpa.gnu.org/packages/")))
(package-initialize)

;; Use-package
(unless (package-installed-p 'use-package)
  (package-refresh-contents)
  (package-install 'use-package))

(require 'use-package)
(setq use-package-always-ensure t)

;; Essential packages
(use-package company
  :config
  (global-company-mode t))

(use-package flycheck
  :config
  (global-flycheck-mode t))

(use-package magit)

(use-package projectile
  :config
  (projectile-mode t))
`
	
	return config, nil
}

// generateXMLConfig generates XML configuration files
func (m *Manager) generateXMLConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration, configFile string) (string, error) {
	if strings.Contains(configFile, ".project") {
		return m.generateEclipseProject(env), nil
	}
	
	return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!-- Generated configuration for " + env.Name + " -->\n", nil
}

// generateEclipseProject generates Eclipse project configuration
func (m *Manager) generateEclipseProject(env *dev.DevEnvironment) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<projectDescription>
	<name>` + env.Name + `</name>
	<comment>Generated by nixai</comment>
	<projects>
	</projects>
	<buildSpec>
		<buildCommand>
			<name>org.eclipse.jdt.core.javabuilder</name>
			<arguments>
			</arguments>
		</buildCommand>
	</buildSpec>
	<natures>
		<nature>org.eclipse.jdt.core.javanature</nature>
	</natures>
</projectDescription>
`
}

// generateGenericConfig generates generic configuration files
func (m *Manager) generateGenericConfig(env *dev.DevEnvironment, integration *dev.IDEIntegration, configFile string) (string, error) {
	return "# Generated configuration for " + env.Name + "\n", nil
}

// createIDEDirectories creates necessary directories for IDE
func (m *Manager) createIDEDirectories(env *dev.DevEnvironment, integration *dev.IDEIntegration) error {
	var directories []string
	
	switch integration.Name {
	case "Visual Studio Code":
		directories = []string{".vscode"}
	case "Neovim":
		directories = []string{".nvim"}
	case "IntelliJ IDEA":
		directories = []string{".idea"}
	case "Eclipse":
		directories = []string{".metadata"}
	}
	
	for _, dir := range directories {
		dirPath := filepath.Join(env.Path, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}

// installVSCodeExtensions installs VS Code extensions
func (m *Manager) installVSCodeExtensions(extensions []string) error {
	m.logger.Info(fmt.Sprintf("VS Code extensions need to be installed manually: %v", extensions))
	return nil
}

// installNeovimPlugins installs Neovim plugins
func (m *Manager) installNeovimPlugins(plugins []string) error {
	m.logger.Info(fmt.Sprintf("Neovim plugins need to be installed manually: %v", plugins))
	return nil
}

// installVimPlugins installs Vim plugins
func (m *Manager) installVimPlugins(plugins []string) error {
	m.logger.Info(fmt.Sprintf("Vim plugins need to be installed manually: %v", plugins))
	return nil
}

// installEmacsPackages installs Emacs packages
func (m *Manager) installEmacsPackages(packages []string) error {
	m.logger.Info(fmt.Sprintf("Emacs packages need to be installed manually: %v", packages))
	return nil
}