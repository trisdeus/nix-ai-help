package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"nix-ai-help/internal/plugins"
	"nix-ai-help/pkg/logger"
)

// PluginIntegration handles plugin-related TUI functionality
type PluginIntegration struct {
	manager     plugins.PluginManager
	registry    plugins.PluginRegistry
	integration *plugins.SimplePluginIntegration
	logger      *logger.Logger
}

// NewPluginIntegration creates a new plugin integration handler
func NewPluginIntegration(logger *logger.Logger) *PluginIntegration {
	// In a real implementation, these would be properly initialized
	// For now, we'll create mock instances to demonstrate the integration
	return &PluginIntegration{
		logger: logger,
	}
}

// GetAvailablePluginCommands returns all available plugin commands
func (pi *PluginIntegration) GetAvailablePluginCommands() []Command {
	var pluginCommands []Command
	
	// Get integrated plugin commands
	if pi.integration != nil {
		integrated := pi.integration.GetIntegratedCommands()
		for _, plugin := range integrated {
			// Add main plugin command
			pluginCommands = append(pluginCommands, Command{
				Name:        plugin.Name,
				Description: plugin.Description,
				Category:    "Plugin Commands",
				Usage:       fmt.Sprintf("nixai %s [subcommand]", plugin.Name),
				Examples:    plugin.Examples,
			})
			
			// Add subcommands
			for _, subcmd := range plugin.Commands {
				pluginCommands = append(pluginCommands, Command{
					Name:        fmt.Sprintf("%s %s", plugin.Name, subcmd),
					Description: fmt.Sprintf("Subcommand for %s", plugin.Name),
					Category:    "Plugin Commands",
					Usage:       fmt.Sprintf("nixai %s %s [options]", plugin.Name, subcmd),
					Examples:    []string{fmt.Sprintf("nixai %s %s", plugin.Name, subcmd)},
				})
			}
		}
	}
	
	// Get external plugins if manager is available
	if pi.manager != nil {
		plugins := pi.manager.ListPlugins()
		for _, plugin := range plugins {
			// Add plugin command with its operations
			operations := plugin.GetOperations()
			for _, op := range operations {
				pluginCommands = append(pluginCommands, Command{
					Name:        fmt.Sprintf("%s %s", plugin.Name(), op.Name),
					Description: op.Description,
					Category:    "External Plugins",
					Usage:       fmt.Sprintf("nixai plugin execute %s %s", plugin.Name(), op.Name),
					Examples: []string{
						fmt.Sprintf("nixai plugin execute %s %s", plugin.Name(), op.Name),
					},
				})
			}
		}
	}
	
	return pluginCommands
}

// GetPluginSuggestions returns intelligent suggestions for plugin commands
func (pi *PluginIntegration) GetPluginSuggestions(query string) []CommandSuggestion {
	var suggestions []CommandSuggestion
	
	// Get plugin commands
	pluginCommands := pi.GetAvailablePluginCommands()
	
	// Score commands based on relevance to query
	for _, cmd := range pluginCommands {
		relevance := pi.calculatePluginRelevance(query, cmd)
		if relevance > 0.1 {
			suggestion := CommandSuggestion{
				Command:   cmd,
				Relevance: relevance,
				Reason:    pi.generatePluginReason(query, cmd),
				Keywords:  pi.extractPluginKeywords(query, cmd),
				UsageHint: pi.generatePluginUsageHint(cmd),
			}
			suggestions = append(suggestions, suggestion)
		}
	}
	
	// Sort by relevance
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Relevance > suggestions[i].Relevance {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}
	
	// Return top 5 suggestions
	if len(suggestions) > 5 {
			suggestions = suggestions[:5]
	}
	
	return suggestions
}

// calculatePluginRelevance scores how relevant a plugin command is to the query
func (pi *PluginIntegration) calculatePluginRelevance(query string, cmd Command) float64 {
	query = strings.ToLower(query)
	cmdName := strings.ToLower(cmd.Name)
	cmdDesc := strings.ToLower(cmd.Description)
	
	score := 0.0
	
	// Direct name match gets highest score
	if strings.Contains(query, cmdName) || strings.Contains(cmdName, query) {
			score += 1.0
	}
	
	// Description match gets good score
	queryWords := strings.Fields(query)
	for _, word := range queryWords {
		if strings.Contains(cmdDesc, word) {
				score += 0.3
		}
		if strings.Contains(cmdName, word) {
				score += 0.5
		}
	}
	
	// Plugin-specific scoring
	if strings.Contains(cmd.Category, "Plugin") {
			score += 0.2
	}
	
	return score
}

// generatePluginReason explains why a plugin command was suggested
func (pi *PluginIntegration) generatePluginReason(query string, cmd Command) string {
	cmdName := cmd.Name
	query = strings.ToLower(query)
	
	// Specific reason patterns for plugins
	if strings.Contains(query, "system") && strings.Contains(cmdName, "system-info") {
		return "Perfect match for system information and monitoring"
	}
	if strings.Contains(query, "package") && strings.Contains(cmdName, "package-monitor") {
		return "Perfect match for package monitoring and management"
	}
	if strings.Contains(query, "monitor") && strings.Contains(cmdName, "monitor") {
		return "Command provides monitoring capabilities"
	}
	
	// Generic patterns
	if strings.Contains(strings.ToLower(cmd.Description), strings.ToLower(query)) {
		return fmt.Sprintf("Plugin command description matches your query about '%s'", query)
	}
	
	return fmt.Sprintf("Suggested plugin command based on relevance to '%s'", query)
}

// extractPluginKeywords finds which keywords from the query match the plugin command
func (pi *PluginIntegration) extractPluginKeywords(query string, cmd Command) []string {
	var keywords []string
	queryWords := strings.Fields(strings.ToLower(query))
	cmdText := strings.ToLower(fmt.Sprintf("%s %s %s", cmd.Name, cmd.Description, cmd.Category))
	
	for _, word := range queryWords {
		if strings.Contains(cmdText, word) && len(word) > 2 {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// generatePluginUsageHint provides a usage suggestion for plugin commands
func (pi *PluginIntegration) generatePluginUsageHint(cmd Command) string {
	if len(cmd.Examples) > 0 {
		return fmt.Sprintf("Try: %s", cmd.Examples[0])
	}
	return fmt.Sprintf("Try: %s", cmd.Usage)
}

// RenderPluginStatus displays the status of plugins in a visually appealing way
func (pi *PluginIntegration) RenderPluginStatus() string {
	var sections []string
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7C3AED")).
		Bold(true).
		Padding(0, 1)
	
	sections = append(sections, headerStyle.Render("🔌 Plugin Status"))
	
	// Integrated plugins
	if pi.integration != nil {
		integrated := pi.integration.GetIntegratedCommands()
		if len(integrated) > 0 {
			sections = append(sections, "")
			sections = append(sections, headerStyle.Render("🔧 Integrated Plugins"))
			
			for _, plugin := range integrated {
				statusStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#10B981")).
					Padding(0, 1)
				
				pluginLine := fmt.Sprintf("✅ %s v%s - %s", plugin.Name, plugin.Version, plugin.Description)
				sections = append(sections, statusStyle.Render(pluginLine))
			}
		}
	}
	
	// External plugins (if manager is available)
	if pi.manager != nil {
		plugins := pi.manager.ListPlugins()
		if len(plugins) > 0 {
			sections = append(sections, "")
			sections = append(sections, headerStyle.Render("📦 External Plugins"))
			
			for _, plugin := range plugins {
				statusStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#F59E0B")).
					Padding(0, 1)
				
				pluginLine := fmt.Sprintf("🔌 %s v%s - %s", plugin.Name(), plugin.Version(), plugin.Description())
				sections = append(sections, statusStyle.Render(pluginLine))
			}
		}
	}
	
	return strings.Join(sections, "\n")
}