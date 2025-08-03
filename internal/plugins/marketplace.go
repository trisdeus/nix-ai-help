package plugins

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"nix-ai-help/pkg/logger"
)

// Marketplace provides access to community-contributed plugins
type Marketplace struct {
	logger     *logger.Logger
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// MarketplacePlugin represents a plugin in the marketplace
type MarketplacePlugin struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	LongDescription string `json:"long_description"`
	Author          string `json:"author"`
	AuthorURL       string `json:"author_url"`
	Version         string `json:"version"`
	License         string `json:"license"`
	Repository      string `json:"repository"`
	Homepage        string `json:"homepage"`
	Documentation   string `json:"documentation"`

	// Marketplace specific fields
	Downloads   int64    `json:"downloads"`
	Rating      float64  `json:"rating"`
	ReviewCount int      `json:"review_count"`
	Featured    bool     `json:"featured"`
	Verified    bool     `json:"verified"`
	Categories  []string `json:"categories"`
	Tags        []string `json:"tags"`
	Screenshots []string `json:"screenshots"`

	// Version information
	LatestVersion  string          `json:"latest_version"`
	VersionHistory []PluginVersion `json:"version_history"`

	// Metadata
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`

	// Requirements
	MinNixaiVersion string   `json:"min_nixai_version"`
	SupportedOS     []string `json:"supported_os"`
	Dependencies    []string `json:"dependencies"`

	// Download information
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
}

// PluginVersion represents a specific version of a plugin
type PluginVersion struct {
	Version      string    `json:"version"`
	ReleaseNotes string    `json:"release_notes"`
	Downloads    int64     `json:"downloads"`
	PublishedAt  time.Time `json:"published_at"`
	Deprecated   bool      `json:"deprecated"`
	Yanked       bool      `json:"yanked"`
	Checksum     string    `json:"checksum"`
	Size         int64     `json:"size"`
}

// PluginReview represents a user review of a plugin
type PluginReview struct {
	ID        string    `json:"id"`
	PluginID  string    `json:"plugin_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Rating    int       `json:"rating"` // 1-5 stars
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Helpful   int       `json:"helpful"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Verified  bool      `json:"verified"` // Verified purchase/user
}

// SearchFilters represents filters for marketplace search
type SearchFilters struct {
	Category     string     `json:"category"`
	Tags         []string   `json:"tags"`
	Author       string     `json:"author"`
	MinRating    float64    `json:"min_rating"`
	FeaturedOnly bool       `json:"featured_only"`
	VerifiedOnly bool       `json:"verified_only"`
	SupportedOS  []string   `json:"supported_os"`
	MaxDownloads int64      `json:"max_downloads"`
	MinDownloads int64      `json:"min_downloads"`
	UpdatedSince *time.Time `json:"updated_since"`
	License      string     `json:"license"`
}

// SortOption represents sorting options for search results
type SortOption string

const (
	SortByRelevance  SortOption = "relevance"
	SortByPopularity SortOption = "popularity"
	SortByRating     SortOption = "rating"
	SortByUpdated    SortOption = "updated"
	SortByCreated    SortOption = "created"
	SortByName       SortOption = "name"
	SortByDownloads  SortOption = "downloads"
)

// SearchResult represents search results from the marketplace
type SearchResult struct {
	Plugins    []MarketplacePlugin `json:"plugins"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	HasMore    bool                `json:"has_more"`
	Query      string              `json:"query"`
	Filters    SearchFilters       `json:"filters"`
	SortBy     SortOption          `json:"sort_by"`
	SearchTime time.Duration       `json:"search_time"`
}

// MarketplaceStats represents marketplace statistics
type MarketplaceStats struct {
	TotalPlugins      int                 `json:"total_plugins"`
	TotalDownloads    int64               `json:"total_downloads"`
	TotalAuthors      int                 `json:"total_authors"`
	PopularCategories []CategoryStats     `json:"popular_categories"`
	FeaturedPlugins   []MarketplacePlugin `json:"featured_plugins"`
	RecentlyUpdated   []MarketplacePlugin `json:"recently_updated"`
	TrendingPlugins   []MarketplacePlugin `json:"trending_plugins"`
	NewPlugins        []MarketplacePlugin `json:"new_plugins"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
	Popular  bool   `json:"popular"`
}

// NewMarketplace creates a new marketplace client
func NewMarketplace(log *logger.Logger) *Marketplace {
	return &Marketplace{
		logger:     log,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://marketplace.nixai.io/api/v1",
		apiKey:     "", // Would be configured or obtained from auth
	}
}

// Search searches for plugins in the marketplace
func (m *Marketplace) Search(ctx context.Context, query string, filters SearchFilters, sortBy SortOption, page, pageSize int) (*SearchResult, error) {
	m.logger.Info(fmt.Sprintf("Searching marketplace for: %s", query))

	// For now, return mock results - this would make actual API calls
	mockPlugins := m.generateMockPlugins(query, 5)

	// Apply filters (mock implementation)
	filteredPlugins := m.applyFilters(mockPlugins, filters)

	// Apply sorting (mock implementation)
	m.applySorting(filteredPlugins, sortBy)

	result := &SearchResult{
		Plugins:    filteredPlugins,
		TotalCount: len(filteredPlugins),
		Page:       page,
		PageSize:   pageSize,
		HasMore:    false,
		Query:      query,
		Filters:    filters,
		SortBy:     sortBy,
		SearchTime: 50 * time.Millisecond,
	}

	return result, nil
}

// GetPlugin retrieves detailed information about a specific plugin
func (m *Marketplace) GetPlugin(ctx context.Context, pluginID string) (*MarketplacePlugin, error) {
	m.logger.Info(fmt.Sprintf("Fetching plugin details for: %s", pluginID))

	// Mock implementation
	plugin := &MarketplacePlugin{
		ID:              pluginID,
		Name:            pluginID,
		DisplayName:     fmt.Sprintf("Plugin %s", pluginID),
		Description:     fmt.Sprintf("A helpful plugin for %s functionality", pluginID),
		LongDescription: "This is a detailed description of the plugin capabilities and features.",
		Author:          "Community Developer",
		AuthorURL:       "https://github.com/developer",
		Version:         "1.2.3",
		License:         "MIT",
		Repository:      fmt.Sprintf("https://github.com/nixai-plugins/%s", pluginID),
		Homepage:        fmt.Sprintf("https://nixai-plugins.github.io/%s", pluginID),
		Documentation:   fmt.Sprintf("https://docs.nixai.io/plugins/%s", pluginID),
		Downloads:       12500,
		Rating:          4.5,
		ReviewCount:     89,
		Featured:        false,
		Verified:        true,
		Categories:      []string{"productivity", "development"},
		Tags:            []string{"automation", "workflow", "utility"},
		Screenshots: []string{
			fmt.Sprintf("https://marketplace.nixai.io/screenshots/%s/1.png", pluginID),
			fmt.Sprintf("https://marketplace.nixai.io/screenshots/%s/2.png", pluginID),
		},
		LatestVersion:   "1.2.3",
		CreatedAt:       time.Now().AddDate(0, -6, 0),
		UpdatedAt:       time.Now().AddDate(0, 0, -7),
		PublishedAt:     time.Now().AddDate(0, -6, 0),
		MinNixaiVersion: "0.1.0",
		SupportedOS:     []string{"linux", "darwin"},
		Dependencies:    []string{},
		DownloadURL:     fmt.Sprintf("https://marketplace.nixai.io/download/%s/latest", pluginID),
		Checksum:        "sha256:abcd1234567890",
		Size:            2048,
	}

	return plugin, nil
}

// GetPopularPlugins retrieves popular plugins from the marketplace
func (m *Marketplace) GetPopularPlugins(ctx context.Context, category string, limit int) ([]MarketplacePlugin, error) {
	m.logger.Info("Fetching popular plugins from marketplace")

	plugins := m.generateMockPlugins("popular", limit)

	// Sort by downloads (mock popularity)
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Downloads > plugins[j].Downloads
	})

	return plugins, nil
}

// GetFeaturedPlugins retrieves featured plugins from the marketplace
func (m *Marketplace) GetFeaturedPlugins(ctx context.Context) ([]MarketplacePlugin, error) {
	m.logger.Info("Fetching featured plugins from marketplace")

	plugins := m.generateMockPlugins("featured", 3)
	for i := range plugins {
		plugins[i].Featured = true
		plugins[i].Verified = true
	}

	return plugins, nil
}

// GetNewPlugins retrieves recently published plugins
func (m *Marketplace) GetNewPlugins(ctx context.Context, limit int) ([]MarketplacePlugin, error) {
	m.logger.Info("Fetching new plugins from marketplace")

	plugins := m.generateMockPlugins("new", limit)

	// Sort by published date
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].PublishedAt.After(plugins[j].PublishedAt)
	})

	return plugins, nil
}

// GetPluginReviews retrieves reviews for a specific plugin
func (m *Marketplace) GetPluginReviews(ctx context.Context, pluginID string, page, pageSize int) ([]PluginReview, error) {
	m.logger.Info(fmt.Sprintf("Fetching reviews for plugin: %s", pluginID))

	// Generate mock reviews
	reviews := []PluginReview{
		{
			ID:        "review-1",
			PluginID:  pluginID,
			UserID:    "user-123",
			Username:  "Alice",
			Rating:    5,
			Title:     "Excellent plugin",
			Content:   "This plugin works perfectly for my use case!",
			Helpful:   8,
			CreatedAt: time.Now().AddDate(0, 0, -5),
			UpdatedAt: time.Now().AddDate(0, 0, -5),
			Verified:  true,
		},
		{
			ID:        "review-2",
			PluginID:  pluginID,
			UserID:    "user-456",
			Username:  "Bob",
			Rating:    4,
			Title:     "Good functionality",
			Content:   "Works well for most use cases. Could use better documentation.",
			Helpful:   8,
			CreatedAt: time.Now().AddDate(0, 0, -20),
			UpdatedAt: time.Now().AddDate(0, 0, -20),
			Verified:  false,
		},
	}

	return reviews, nil
}

// SubmitPluginReview submits a review for a plugin
func (m *Marketplace) SubmitPluginReview(ctx context.Context, pluginID string, review PluginReview) error {
	m.logger.Info(fmt.Sprintf("Submitting review for plugin: %s", pluginID))
	
	// In a real implementation, this would submit the review to the marketplace
	// For now, just log that the review was submitted
	m.logger.Info(fmt.Sprintf("Review submitted for plugin %s: %s (Rating: %d)", 
		pluginID, review.Title, review.Rating))
	
	return nil
}

// GetMarketplaceStats retrieves overall marketplace statistics
func (m *Marketplace) GetMarketplaceStats(ctx context.Context) (*MarketplaceStats, error) {
	m.logger.Info("Fetching marketplace statistics")

	featuredPlugins, _ := m.GetFeaturedPlugins(ctx)
	newPlugins, _ := m.GetNewPlugins(ctx, 5)

	stats := &MarketplaceStats{
		TotalPlugins:   156,
		TotalDownloads: 1250000,
		TotalAuthors:   89,
		PopularCategories: []CategoryStats{
			{Category: "productivity", Count: 45, Popular: true},
			{Category: "development", Count: 38, Popular: true},
			{Category: "system", Count: 32, Popular: true},
			{Category: "automation", Count: 28, Popular: false},
			{Category: "networking", Count: 13, Popular: false},
		},
		FeaturedPlugins: featuredPlugins,
		RecentlyUpdated: newPlugins[:3],
		TrendingPlugins: newPlugins[:3],
		NewPlugins:      newPlugins,
	}

	return stats, nil
}

// GetCategories retrieves all available plugin categories
func (m *Marketplace) GetCategories(ctx context.Context) ([]CategoryStats, error) {
	m.logger.Info("Fetching plugin categories")

	categories := []CategoryStats{
		{Category: "productivity", Count: 45, Popular: true},
		{Category: "development", Count: 38, Popular: true},
		{Category: "system", Count: 32, Popular: true},
		{Category: "automation", Count: 28, Popular: false},
		{Category: "networking", Count: 13, Popular: false},
		{Category: "security", Count: 11, Popular: false},
		{Category: "monitoring", Count: 8, Popular: false},
		{Category: "backup", Count: 6, Popular: false},
		{Category: "multimedia", Count: 4, Popular: false},
	}

	return categories, nil
}

// SubmitPlugin submits a new plugin to the marketplace
func (m *Marketplace) SubmitPlugin(ctx context.Context, plugin *MarketplacePlugin) error {
	m.logger.Info(fmt.Sprintf("Submitting plugin to marketplace: %s", plugin.Name))

	// This would validate the plugin and submit it for review
	// For now, just log the submission
	m.logger.Info("Plugin submitted successfully - pending review")
	return nil
}

// UpdatePlugin updates an existing plugin in the marketplace
func (m *Marketplace) UpdatePlugin(ctx context.Context, pluginID string, plugin *MarketplacePlugin) error {
	m.logger.Info(fmt.Sprintf("Updating plugin in marketplace: %s", pluginID))

	// This would update the plugin information
	m.logger.Info("Plugin updated successfully")
	return nil
}

// Private helper methods

func (m *Marketplace) generateMockPlugins(prefix string, count int) []MarketplacePlugin {
	plugins := make([]MarketplacePlugin, count)

	for i := 0; i < count; i++ {
		plugins[i] = MarketplacePlugin{
			ID:              fmt.Sprintf("%s-plugin-%d", prefix, i+1),
			Name:            fmt.Sprintf("%s-plugin-%d", prefix, i+1),
			DisplayName:     fmt.Sprintf("%s Plugin %d", prefix, i+1),
			Description:     fmt.Sprintf("A useful %s plugin for nixai", prefix),
			Author:          fmt.Sprintf("Author%d", i+1),
			Version:         "1.0.0",
			License:         "MIT",
			Downloads:       int64(1000 + i*500),
			Rating:          4.0 + float64(i%6)/10.0,
			ReviewCount:     10 + i*5,
			Featured:        i < 2,
			Verified:        i < 3,
			Categories:      []string{"productivity", "development"},
			Tags:            []string{"automation", "workflow"},
			CreatedAt:       time.Now().AddDate(0, -i-1, 0),
			UpdatedAt:       time.Now().AddDate(0, 0, -i-1),
			PublishedAt:     time.Now().AddDate(0, -i-1, 0),
			MinNixaiVersion: "0.1.0",
			SupportedOS:     []string{"linux", "darwin"},
		}
	}

	return plugins
}

func (m *Marketplace) applyFilters(plugins []MarketplacePlugin, filters SearchFilters) []MarketplacePlugin {
	var filtered []MarketplacePlugin

	for _, plugin := range plugins {
		// Apply category filter
		if filters.Category != "" {
			found := false
			for _, cat := range plugin.Categories {
				if cat == filters.Category {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Apply rating filter
		if filters.MinRating > 0 && plugin.Rating < filters.MinRating {
			continue
		}

		// Apply featured filter
		if filters.FeaturedOnly && !plugin.Featured {
			continue
		}

		// Apply verified filter
		if filters.VerifiedOnly && !plugin.Verified {
			continue
		}

		filtered = append(filtered, plugin)
	}

	return filtered
}

func (m *Marketplace) applySorting(plugins []MarketplacePlugin, sortBy SortOption) {
	switch sortBy {
	case SortByPopularity:
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Downloads > plugins[j].Downloads
		})
	case SortByRating:
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Rating > plugins[j].Rating
		})
	case SortByUpdated:
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].UpdatedAt.After(plugins[j].UpdatedAt)
		})
	case SortByCreated:
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].CreatedAt.After(plugins[j].CreatedAt)
		})
	case SortByName:
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Name < plugins[j].Name
		})
	case SortByDownloads:
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Downloads > plugins[j].Downloads
		})
	}
}
