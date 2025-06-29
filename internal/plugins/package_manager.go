package plugins

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nix-ai-help/pkg/logger"
)

// PackageManager handles plugin distribution and updates
type PackageManager struct {
	logger       *logger.Logger
	registry     *RemoteRegistry
	cacheDir     string
	repositories []PluginRepository
}

// PluginRepository represents a plugin repository
type PluginRepository struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Type        string    `json:"type"` // "official", "community", "local"
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	Verified    bool      `json:"verified"`
	LastUpdated time.Time `json:"last_updated"`
}

// PluginPackage represents a packaged plugin
type PluginPackage struct {
	Metadata    *PluginMetadata `json:"metadata"`
	Files       []PackageFile   `json:"files"`
	Checksum    string          `json:"checksum"`
	Size        int64           `json:"size"`
	DownloadURL string          `json:"download_url"`
	Signature   string          `json:"signature"`
}

// PackageFile represents a file in a plugin package
type PackageFile struct {
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
	Size     int64  `json:"size"`
	Mode     uint32 `json:"mode"`
}

// RemoteRegistry handles remote plugin registry operations
type RemoteRegistry struct {
	BaseURL    string
	HTTPClient *http.Client
	logger     *logger.Logger
}

// PluginIndex represents the remote plugin index
type PluginIndex struct {
	Version     string                   `json:"version"`
	LastUpdated time.Time                `json:"last_updated"`
	Plugins     map[string]PluginPackage `json:"plugins"`
	Categories  map[string][]string      `json:"categories"`
	Tags        map[string][]string      `json:"tags"`
	Checksum    string                   `json:"checksum"`
}

// UpdateInfo represents plugin update information
type UpdateInfo struct {
	PluginName      string `json:"plugin_name"`
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	ChangelogURL    string `json:"changelog_url"`
	BreakingChanges bool   `json:"breaking_changes"`
	SecurityUpdate  bool   `json:"security_update"`
}

// NewPackageManager creates a new package manager
func NewPackageManager(log *logger.Logger) *PackageManager {
	cacheDir := filepath.Join(os.TempDir(), "nixai", "plugin-cache")

	registry := &RemoteRegistry{
		BaseURL:    "https://api.nixai.io/plugins", // Official registry
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		logger:     log,
	}

	return &PackageManager{
		logger:   log,
		registry: registry,
		cacheDir: cacheDir,
		repositories: []PluginRepository{
			{
				Name:        "official",
				URL:         "https://api.nixai.io/plugins",
				Type:        "official",
				Enabled:     true,
				Priority:    100,
				Verified:    true,
				LastUpdated: time.Now(),
			},
			{
				Name:        "community",
				URL:         "https://community.nixai.io/plugins",
				Type:        "community",
				Enabled:     true,
				Priority:    50,
				Verified:    false,
				LastUpdated: time.Now(),
			},
		},
	}
}

// InstallFromRepository installs a plugin from a repository
func (pm *PackageManager) InstallFromRepository(ctx context.Context, pluginName, version string) error {
	pm.logger.Info(fmt.Sprintf("Installing plugin '%s' version '%s' from repository", pluginName, version))

	// Find plugin in repositories
	pluginPackage, repo, err := pm.findPluginInRepositories(ctx, pluginName, version)
	if err != nil {
		return fmt.Errorf("failed to find plugin: %w", err)
	}

	// Download and verify plugin
	pluginPath, err := pm.downloadAndVerifyPlugin(ctx, pluginPackage, repo)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}

	// Extract and install plugin
	if err := pm.extractAndInstallPlugin(ctx, pluginPath, pluginPackage); err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	pm.logger.Info(fmt.Sprintf("Successfully installed plugin '%s'", pluginName))
	return nil
}

// UpdatePlugin updates a plugin to the latest version
func (pm *PackageManager) UpdatePlugin(ctx context.Context, pluginName string) (*UpdateInfo, error) {
	pm.logger.Info(fmt.Sprintf("Checking for updates for plugin '%s'", pluginName))

	// Get current version (this would integrate with the plugin manager)
	currentVersion := "1.0.0" // Placeholder

	// Check for updates
	updateInfo, err := pm.checkForUpdates(ctx, pluginName, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	if !updateInfo.UpdateAvailable {
		pm.logger.Info(fmt.Sprintf("Plugin '%s' is already up to date", pluginName))
		return updateInfo, nil
	}

	// Download and install update
	if err := pm.InstallFromRepository(ctx, pluginName, updateInfo.LatestVersion); err != nil {
		return nil, fmt.Errorf("failed to install update: %w", err)
	}

	pm.logger.Info(fmt.Sprintf("Successfully updated plugin '%s' to version %s", pluginName, updateInfo.LatestVersion))
	return updateInfo, nil
}

// CheckAllUpdates checks for updates for all installed plugins
func (pm *PackageManager) CheckAllUpdates(ctx context.Context, installedPlugins []string) ([]UpdateInfo, error) {
	updates := make([]UpdateInfo, 0, len(installedPlugins))

	for _, pluginName := range installedPlugins {
		updateInfo, err := pm.checkForUpdates(ctx, pluginName, "1.0.0") // Would get real version
		if err != nil {
			pm.logger.Warn(fmt.Sprintf("Failed to check updates for %s: %v", pluginName, err))
			continue
		}
		updates = append(updates, *updateInfo)
	}

	return updates, nil
}

// SearchPlugins searches for plugins in repositories
func (pm *PackageManager) SearchPlugins(ctx context.Context, query string, filters map[string]string) ([]PluginPackage, error) {
	pm.logger.Info(fmt.Sprintf("Searching for plugins: %s", query))

	var allResults []PluginPackage

	for _, repo := range pm.repositories {
		if !repo.Enabled {
			continue
		}

		results, err := pm.searchInRepository(ctx, repo, query, filters)
		if err != nil {
			pm.logger.Warn(fmt.Sprintf("Failed to search in repository %s: %v", repo.Name, err))
			continue
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// GetPluginInfo retrieves detailed information about a plugin
func (pm *PackageManager) GetPluginInfo(ctx context.Context, pluginName string) (*PluginPackage, error) {
	for _, repo := range pm.repositories {
		if !repo.Enabled {
			continue
		}

		pluginPackage, err := pm.getPluginFromRepository(ctx, repo, pluginName)
		if err != nil {
			continue // Try next repository
		}

		return pluginPackage, nil
	}

	return nil, fmt.Errorf("plugin '%s' not found in any repository", pluginName)
}

// UpdateRepositories updates the repository indexes
func (pm *PackageManager) UpdateRepositories(ctx context.Context) error {
	pm.logger.Info("Updating repository indexes")

	for i := range pm.repositories {
		repo := &pm.repositories[i]
		if !repo.Enabled {
			continue
		}

		if err := pm.updateRepositoryIndex(ctx, repo); err != nil {
			pm.logger.Warn(fmt.Sprintf("Failed to update repository %s: %v", repo.Name, err))
			continue
		}

		repo.LastUpdated = time.Now()
	}

	return nil
}

// AddRepository adds a new plugin repository
func (pm *PackageManager) AddRepository(repo PluginRepository) error {
	// Validate repository
	if repo.Name == "" || repo.URL == "" {
		return fmt.Errorf("repository name and URL are required")
	}

	// Check if repository already exists
	for _, existingRepo := range pm.repositories {
		if existingRepo.Name == repo.Name {
			return fmt.Errorf("repository '%s' already exists", repo.Name)
		}
	}

	pm.repositories = append(pm.repositories, repo)
	pm.logger.Info(fmt.Sprintf("Added repository '%s' at %s", repo.Name, repo.URL))
	return nil
}

// RemoveRepository removes a plugin repository
func (pm *PackageManager) RemoveRepository(name string) error {
	for i, repo := range pm.repositories {
		if repo.Name == name {
			// Don't allow removing official repository
			if repo.Type == "official" {
				return fmt.Errorf("cannot remove official repository")
			}

			pm.repositories = append(pm.repositories[:i], pm.repositories[i+1:]...)
			pm.logger.Info(fmt.Sprintf("Removed repository '%s'", name))
			return nil
		}
	}

	return fmt.Errorf("repository '%s' not found", name)
}

// ListRepositories returns all configured repositories
func (pm *PackageManager) ListRepositories() []PluginRepository {
	return pm.repositories
}

// Private helper methods

func (pm *PackageManager) findPluginInRepositories(ctx context.Context, pluginName, version string) (*PluginPackage, *PluginRepository, error) {
	for _, repo := range pm.repositories {
		if !repo.Enabled {
			continue
		}

		pluginPackage, err := pm.getPluginFromRepository(ctx, repo, pluginName)
		if err != nil {
			continue
		}

		// Check version match
		if version == "" || version == "latest" || pluginPackage.Metadata.Version == version {
			return pluginPackage, &repo, nil
		}
	}

	return nil, nil, fmt.Errorf("plugin '%s' version '%s' not found", pluginName, version)
}

func (pm *PackageManager) downloadAndVerifyPlugin(ctx context.Context, pluginPackage *PluginPackage, repo *PluginRepository) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(pm.cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Download plugin
	fileName := fmt.Sprintf("%s-%s.tar.gz", pluginPackage.Metadata.Name, pluginPackage.Metadata.Version)
	filePath := filepath.Join(pm.cacheDir, fileName)

	req, err := http.NewRequestWithContext(ctx, "GET", pluginPackage.DownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pm.registry.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create file and calculate checksum
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(file, hasher)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save plugin: %w", err)
	}

	// Verify checksum
	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))
	if calculatedChecksum != pluginPackage.Checksum {
		os.Remove(filePath)
		return "", fmt.Errorf("checksum mismatch: expected %s, got %s", pluginPackage.Checksum, calculatedChecksum)
	}

	return filePath, nil
}

func (pm *PackageManager) extractAndInstallPlugin(ctx context.Context, pluginPath string, pluginPackage *PluginPackage) error {
	// Open the tar.gz file
	file, err := os.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract to plugin directory
	pluginDir := filepath.Join(os.TempDir(), "nixai", "plugins", pluginPackage.Metadata.Name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Security check: prevent directory traversal
		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		targetPath := filepath.Join(pluginDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(targetFile, tarReader); err != nil {
				targetFile.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			targetFile.Close()
		}
	}

	pm.logger.Info(fmt.Sprintf("Plugin extracted to: %s", pluginDir))
	return nil
}

func (pm *PackageManager) checkForUpdates(ctx context.Context, pluginName, currentVersion string) (*UpdateInfo, error) {
	pluginPackage, err := pm.GetPluginInfo(ctx, pluginName)
	if err != nil {
		return nil, err
	}

	updateInfo := &UpdateInfo{
		PluginName:      pluginName,
		CurrentVersion:  currentVersion,
		LatestVersion:   pluginPackage.Metadata.Version,
		UpdateAvailable: pluginPackage.Metadata.Version != currentVersion,
		ChangelogURL:    pluginPackage.Metadata.Homepage + "/changelog",
		BreakingChanges: false, // Would need to parse version semantics
		SecurityUpdate:  false, // Would need metadata from repository
	}

	return updateInfo, nil
}

func (pm *PackageManager) searchInRepository(ctx context.Context, repo PluginRepository, query string, filters map[string]string) ([]PluginPackage, error) {
	// For now, return empty results - this would make actual HTTP requests to repositories
	pm.logger.Info(fmt.Sprintf("Searching in repository %s for: %s", repo.Name, query))
	return []PluginPackage{}, nil
}

func (pm *PackageManager) getPluginFromRepository(ctx context.Context, repo PluginRepository, pluginName string) (*PluginPackage, error) {
	// For now, return a mock plugin - this would make actual HTTP requests
	return &PluginPackage{
		Metadata: &PluginMetadata{
			Name:        pluginName,
			Version:     "1.0.0",
			Description: fmt.Sprintf("Plugin %s from repository %s", pluginName, repo.Name),
			Author:      "Community",
			License:     "MIT",
		},
		DownloadURL: fmt.Sprintf("%s/download/%s/latest", repo.URL, pluginName),
		Checksum:    "mock-checksum",
		Size:        1024,
	}, nil
}

func (pm *PackageManager) updateRepositoryIndex(ctx context.Context, repo *PluginRepository) error {
	pm.logger.Info(fmt.Sprintf("Updating index for repository: %s", repo.Name))
	// This would fetch and cache the repository index
	return nil
}
