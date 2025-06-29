package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/pkg/logger"
)

// ConfigExtraction represents extracted configuration data
type ConfigExtraction struct {
	FilePath    string            `json:"file_path"`
	ServiceType string            `json:"service_type"`
	Content     string            `json:"content"`
	ParsedData  map[string]string `json:"parsed_data"`
	Importance  string            `json:"importance"`
	Warnings    []string          `json:"warnings"`
}

// ConfigExtractor extracts and parses configuration files
type ConfigExtractor struct {
	logger logger.Logger
}

// NewConfigExtractor creates a new configuration extractor
func NewConfigExtractor(logger logger.Logger) *ConfigExtractor {
	return &ConfigExtractor{
		logger: logger,
	}
}

// ExtractConfigurations extracts all important configuration files
func (ce *ConfigExtractor) ExtractConfigurations(configFiles []ConfigFileInfo) ([]ConfigExtraction, error) {
	var extractions []ConfigExtraction

	for _, configFile := range configFiles {
		extraction, err := ce.ExtractConfiguration(configFile)
		if err != nil {
			ce.logger.Warn(fmt.Sprintf("Failed to extract %s: %v", configFile.Path, err))
			continue
		}

		extractions = append(extractions, extraction)
	}

	return extractions, nil
}

// ExtractConfiguration extracts a single configuration file
func (ce *ConfigExtractor) ExtractConfiguration(configFile ConfigFileInfo) (ConfigExtraction, error) {
	extraction := ConfigExtraction{
		FilePath:    configFile.Path,
		ServiceType: configFile.Type,
		Importance:  configFile.Importance,
		ParsedData:  make(map[string]string),
		Warnings:    []string{},
	}

	// Read file content
	content, err := os.ReadFile(configFile.Path)
	if err != nil {
		return extraction, fmt.Errorf("failed to read file: %v", err)
	}

	extraction.Content = string(content)

	// Parse configuration based on type
	switch configFile.Type {
	case "nginx":
		ce.parseNginxConfig(&extraction)
	case "apache":
		ce.parseApacheConfig(&extraction)
	case "mysql":
		ce.parseMySQLConfig(&extraction)
	case "postgresql":
		ce.parsePostgreSQLConfig(&extraction)
	case "redis":
		ce.parseRedisConfig(&extraction)
	case "ssh":
		ce.parseSSHConfig(&extraction)
	case "docker":
		ce.parseDockerConfig(&extraction)
	case "system":
		ce.parseSystemConfig(&extraction)
	default:
		ce.parseGenericConfig(&extraction)
	}

	return extraction, nil
}

// parseNginxConfig parses nginx configuration
func (ce *ConfigExtractor) parseNginxConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key configuration directives
		if strings.Contains(line, "listen") {
			extraction.ParsedData["listen"] = ce.extractValue(line)
		} else if strings.Contains(line, "server_name") {
			extraction.ParsedData["server_name"] = ce.extractValue(line)
		} else if strings.Contains(line, "root") {
			extraction.ParsedData["root"] = ce.extractValue(line)
		} else if strings.Contains(line, "ssl_certificate") {
			extraction.ParsedData["ssl_certificate"] = ce.extractValue(line)
		} else if strings.Contains(line, "ssl_certificate_key") {
			extraction.ParsedData["ssl_certificate_key"] = ce.extractValue(line)
		} else if strings.Contains(line, "location") {
			if extraction.ParsedData["locations"] == "" {
				extraction.ParsedData["locations"] = line
			} else {
				extraction.ParsedData["locations"] += "\n" + line
			}
		}
	}

	// Add warnings for complex configurations
	if strings.Contains(extraction.Content, "upstream") {
		extraction.Warnings = append(extraction.Warnings, "Upstream configurations need manual review")
	}
	if strings.Contains(extraction.Content, "rewrite") {
		extraction.Warnings = append(extraction.Warnings, "Rewrite rules need manual migration")
	}
}

// parseApacheConfig parses Apache configuration
func (ce *ConfigExtractor) parseApacheConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key configuration directives
		if strings.HasPrefix(line, "Listen") {
			extraction.ParsedData["listen"] = ce.extractValue(line)
		} else if strings.HasPrefix(line, "ServerName") {
			extraction.ParsedData["server_name"] = ce.extractValue(line)
		} else if strings.HasPrefix(line, "DocumentRoot") {
			extraction.ParsedData["document_root"] = ce.extractValue(line)
		} else if strings.HasPrefix(line, "LoadModule") {
			if extraction.ParsedData["modules"] == "" {
				extraction.ParsedData["modules"] = ce.extractValue(line)
			} else {
				extraction.ParsedData["modules"] += "\n" + ce.extractValue(line)
			}
		}
	}

	// Add warnings
	if strings.Contains(extraction.Content, "LoadModule") {
		extraction.Warnings = append(extraction.Warnings, "Apache modules need manual configuration in NixOS")
	}
	if strings.Contains(extraction.Content, ".htaccess") {
		extraction.Warnings = append(extraction.Warnings, ".htaccess files are not directly supported")
	}
}

// parseMySQLConfig parses MySQL configuration
func (ce *ConfigExtractor) parseMySQLConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Track sections
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			continue
		}

		// Extract key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if currentSection != "" {
					key = fmt.Sprintf("%s.%s", currentSection, key)
				}

				extraction.ParsedData[key] = value
			}
		}
	}

	// Add warnings
	extraction.Warnings = append(extraction.Warnings, "Database data migration required separately")
	extraction.Warnings = append(extraction.Warnings, "User accounts and permissions need manual recreation")
}

// parsePostgreSQLConfig parses PostgreSQL configuration
func (ce *ConfigExtractor) parsePostgreSQLConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key configuration parameters
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "'\"")
				extraction.ParsedData[key] = value
			}
		}
	}

	// Add warnings
	extraction.Warnings = append(extraction.Warnings, "Database dump and restore required for data migration")
	extraction.Warnings = append(extraction.Warnings, "User roles and permissions need manual recreation")

	// Check for pg_hba.conf
	if strings.Contains(extraction.FilePath, "pg_hba.conf") {
		extraction.Warnings = append(extraction.Warnings, "Authentication rules need manual configuration")
	}
}

// parseRedisConfig parses Redis configuration
func (ce *ConfigExtractor) parseRedisConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key configuration parameters
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := fields[0]
			value := strings.Join(fields[1:], " ")
			extraction.ParsedData[key] = value
		}
	}

	// Add warnings for data persistence
	if _, exists := extraction.ParsedData["dir"]; exists {
		extraction.Warnings = append(extraction.Warnings, "Redis data directory needs manual migration")
	}
	if _, exists := extraction.ParsedData["requirepass"]; exists {
		extraction.Warnings = append(extraction.Warnings, "Redis authentication needs manual configuration")
	}
}

// parseSSHConfig parses SSH configuration
func (ce *ConfigExtractor) parseSSHConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key configuration parameters
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := fields[0]
			value := strings.Join(fields[1:], " ")
			extraction.ParsedData[key] = value
		}
	}

	// Add warnings
	extraction.Warnings = append(extraction.Warnings, "SSH host keys need manual migration")
	extraction.Warnings = append(extraction.Warnings, "User authorized_keys files need manual copying")

	// Check for custom port
	if port, exists := extraction.ParsedData["Port"]; exists && port != "22" {
		extraction.Warnings = append(extraction.Warnings, fmt.Sprintf("Custom SSH port %s needs firewall configuration", port))
	}
}

// parseDockerConfig parses Docker daemon configuration
func (ce *ConfigExtractor) parseDockerConfig(extraction *ConfigExtraction) {
	// Docker config is usually JSON
	extraction.ParsedData["raw_json"] = extraction.Content

	// Add warnings
	extraction.Warnings = append(extraction.Warnings, "Docker containers and images need manual migration")
	extraction.Warnings = append(extraction.Warnings, "Docker volumes need manual backup and restore")
	extraction.Warnings = append(extraction.Warnings, "Docker networks need manual recreation")
}

// parseSystemConfig parses system configuration files
func (ce *ConfigExtractor) parseSystemConfig(extraction *ConfigExtraction) {
	switch filepath.Base(extraction.FilePath) {
	case "hosts":
		ce.parseHostsFile(extraction)
	case "fstab":
		ce.parseFstabFile(extraction)
	case "crontab":
		ce.parseCrontabFile(extraction)
	default:
		ce.parseGenericConfig(extraction)
	}
}

// parseHostsFile parses /etc/hosts
func (ce *ConfigExtractor) parseHostsFile(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ip := fields[0]
			hostnames := strings.Join(fields[1:], " ")
			extraction.ParsedData[ip] = hostnames
		}
	}

	extraction.Warnings = append(extraction.Warnings, "Custom host entries need manual addition to networking.hosts")
}

// parseFstabFile parses /etc/fstab
func (ce *ConfigExtractor) parseFstabFile(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			device := fields[0]
			mountpoint := fields[1]
			filesystem := fields[2]
			extraction.ParsedData[mountpoint] = fmt.Sprintf("%s (%s)", device, filesystem)
		}
	}

	extraction.Warnings = append(extraction.Warnings, "Filesystem mounts need manual configuration in fileSystems option")
	extraction.Warnings = append(extraction.Warnings, "UUID/LABEL mappings may need updates")
}

// parseCrontabFile parses crontab files
func (ce *ConfigExtractor) parseCrontabFile(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Simple cron entry parsing
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			schedule := strings.Join(fields[0:5], " ")
			command := strings.Join(fields[5:], " ")
			extraction.ParsedData[fmt.Sprintf("job_%d", i)] = fmt.Sprintf("%s: %s", schedule, command)
		}
	}

	extraction.Warnings = append(extraction.Warnings, "Cron jobs need manual configuration in services.cron.systemCronJobs")
	extraction.Warnings = append(extraction.Warnings, "User-specific crontabs need individual migration")
}

// parseGenericConfig provides generic configuration parsing
func (ce *ConfigExtractor) parseGenericConfig(extraction *ConfigExtraction) {
	lines := strings.Split(extraction.Content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try to extract key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				extraction.ParsedData[key] = value
			}
		} else if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				extraction.ParsedData[key] = value
			}
		}
	}

	extraction.Warnings = append(extraction.Warnings, "Manual review required for configuration migration")
}

// extractValue extracts value from configuration line
func (ce *ConfigExtractor) extractValue(line string) string {
	// Remove directive name and extract value
	parts := strings.Fields(line)
	if len(parts) > 1 {
		value := strings.Join(parts[1:], " ")
		value = strings.Trim(value, ";\"'")
		return value
	}
	return ""
}

// BackupConfiguration creates a backup of configuration files
func (ce *ConfigExtractor) BackupConfiguration(configFiles []ConfigFileInfo, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	for _, configFile := range configFiles {
		if !configFile.Backup {
			continue
		}

		// Create backup file path
		backupPath := filepath.Join(backupDir, strings.ReplaceAll(configFile.Path, "/", "_"))

		// Copy file
		content, err := os.ReadFile(configFile.Path)
		if err != nil {
			ce.logger.Warn(fmt.Sprintf("Failed to read %s for backup: %v", configFile.Path, err))
			continue
		}

		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			ce.logger.Warn(fmt.Sprintf("Failed to backup %s: %v", configFile.Path, err))
			continue
		}

		ce.logger.Info(fmt.Sprintf("Backed up %s to %s", configFile.Path, backupPath))
	}

	return nil
}

// ValidateExtraction validates extracted configuration data
func (ce *ConfigExtractor) ValidateExtraction(extraction ConfigExtraction) []string {
	var warnings []string

	// Check for empty extraction
	if len(extraction.ParsedData) == 0 {
		warnings = append(warnings, "No configuration data extracted")
	}

	// Check for critical configurations
	if extraction.Importance == "critical" && len(extraction.ParsedData) < 3 {
		warnings = append(warnings, "Critical configuration appears incomplete")
	}

	// Service-specific validations
	switch extraction.ServiceType {
	case "nginx", "apache":
		if _, exists := extraction.ParsedData["listen"]; !exists {
			warnings = append(warnings, "No listen port configuration found")
		}
	case "mysql", "postgresql":
		if len(extraction.ParsedData) < 5 {
			warnings = append(warnings, "Database configuration appears minimal")
		}
	case "ssh":
		if port, exists := extraction.ParsedData["Port"]; exists {
			if port != "22" {
				warnings = append(warnings, "Non-standard SSH port detected")
			}
		}
	}

	return warnings
}
