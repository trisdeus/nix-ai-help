package converter

import (
	"fmt"
	"strings"

	"nix-ai-help/internal/migration/detector"
	"nix-ai-help/pkg/logger"
)

// SystemdConverter converts systemd services to NixOS services
type SystemdConverter struct {
	logger logger.Logger
	mapper *detector.ServiceMapper
}

// SystemdService represents a systemd service configuration
type SystemdService struct {
	Name        string            `json:"name"`
	Unit        SystemdUnit       `json:"unit"`
	Service     SystemdServiceDef `json:"service"`
	Install     SystemdInstall    `json:"install"`
	Environment map[string]string `json:"environment"`
}

// SystemdUnit represents the [Unit] section
type SystemdUnit struct {
	Description   string   `json:"description"`
	After         []string `json:"after"`
	Before        []string `json:"before"`
	Requires      []string `json:"requires"`
	Wants         []string `json:"wants"`
	Documentation []string `json:"documentation"`
}

// SystemdServiceDef represents the [Service] section
type SystemdServiceDef struct {
	Type             string            `json:"type"`
	ExecStart        string            `json:"exec_start"`
	ExecStop         string            `json:"exec_stop"`
	ExecReload       string            `json:"exec_reload"`
	User             string            `json:"user"`
	Group            string            `json:"group"`
	WorkingDirectory string            `json:"working_directory"`
	Environment      map[string]string `json:"environment"`
	EnvironmentFile  string            `json:"environment_file"`
	Restart          string            `json:"restart"`
	RestartSec       string            `json:"restart_sec"`
	PIDFile          string            `json:"pid_file"`
}

// SystemdInstall represents the [Install] section
type SystemdInstall struct {
	WantedBy        []string `json:"wanted_by"`
	RequiredBy      []string `json:"required_by"`
	Also            []string `json:"also"`
	DefaultInstance string   `json:"default_instance"`
}

// ConversionResult represents the result of systemd conversion
type ConversionResult struct {
	OriginalService string                  `json:"original_service"`
	NixOSConfig     string                  `json:"nixos_config"`
	ServiceMapping  detector.ServiceMapping `json:"service_mapping"`
	Warnings        []string                `json:"warnings"`
	ManualSteps     []string                `json:"manual_steps"`
	Dependencies    []string                `json:"dependencies"`
}

// NewSystemdConverter creates a new systemd converter
func NewSystemdConverter(logger logger.Logger) *SystemdConverter {
	return &SystemdConverter{
		logger: logger,
		mapper: detector.NewServiceMapper(logger),
	}
}

// ConvertService converts a systemd service to NixOS configuration
func (sc *SystemdConverter) ConvertService(serviceName string, systemdConfig string) (*ConversionResult, error) {
	result := &ConversionResult{
		OriginalService: serviceName,
		Warnings:        []string{},
		ManualSteps:     []string{},
		Dependencies:    []string{},
	}

	// Parse systemd service file
	service, err := sc.parseSystemdService(systemdConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse systemd service: %v", err)
	}

	// Try to map to existing NixOS service
	if mapping, exists := sc.mapper.MapService(serviceName); exists {
		result.ServiceMapping = mapping
		result.NixOSConfig = sc.generateNixOSServiceConfig(service, mapping)
		result.Warnings = append(result.Warnings, sc.mapper.ValidateMapping(mapping)...)
	} else {
		// Generate custom service configuration
		result.NixOSConfig = sc.generateCustomServiceConfig(service, serviceName)
		result.Warnings = append(result.Warnings, "No built-in NixOS service mapping found - using custom service")
		result.ManualSteps = append(result.ManualSteps, "Review generated custom service configuration")
	}

	// Add general warnings and manual steps
	sc.addConversionWarnings(service, result)

	return result, nil
}

// parseSystemdService parses systemd service file content
func (sc *SystemdConverter) parseSystemdService(content string) (*SystemdService, error) {
	service := &SystemdService{
		Environment: make(map[string]string),
		Unit:        SystemdUnit{},
		Service:     SystemdServiceDef{Environment: make(map[string]string)},
		Install:     SystemdInstall{},
	}

	lines := strings.Split(content, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}

		// Parse key-value pairs
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch currentSection {
			case "unit":
				sc.parseUnitSection(key, value, &service.Unit)
			case "service":
				sc.parseServiceSection(key, value, &service.Service)
			case "install":
				sc.parseInstallSection(key, value, &service.Install)
			}
		}
	}

	return service, nil
}

// parseUnitSection parses [Unit] section
func (sc *SystemdConverter) parseUnitSection(key, value string, unit *SystemdUnit) {
	switch key {
	case "Description":
		unit.Description = value
	case "After":
		unit.After = strings.Split(value, " ")
	case "Before":
		unit.Before = strings.Split(value, " ")
	case "Requires":
		unit.Requires = strings.Split(value, " ")
	case "Wants":
		unit.Wants = strings.Split(value, " ")
	case "Documentation":
		unit.Documentation = strings.Split(value, " ")
	}
}

// parseServiceSection parses [Service] section
func (sc *SystemdConverter) parseServiceSection(key, value string, service *SystemdServiceDef) {
	switch key {
	case "Type":
		service.Type = value
	case "ExecStart":
		service.ExecStart = value
	case "ExecStop":
		service.ExecStop = value
	case "ExecReload":
		service.ExecReload = value
	case "User":
		service.User = value
	case "Group":
		service.Group = value
	case "WorkingDirectory":
		service.WorkingDirectory = value
	case "Environment":
		// Parse environment variables
		envVars := strings.Split(value, " ")
		for _, envVar := range envVars {
			if strings.Contains(envVar, "=") {
				envParts := strings.SplitN(envVar, "=", 2)
				if len(envParts) == 2 {
					service.Environment[envParts[0]] = envParts[1]
				}
			}
		}
	case "EnvironmentFile":
		service.EnvironmentFile = value
	case "Restart":
		service.Restart = value
	case "RestartSec":
		service.RestartSec = value
	case "PIDFile":
		service.PIDFile = value
	}
}

// parseInstallSection parses [Install] section
func (sc *SystemdConverter) parseInstallSection(key, value string, install *SystemdInstall) {
	switch key {
	case "WantedBy":
		install.WantedBy = strings.Split(value, " ")
	case "RequiredBy":
		install.RequiredBy = strings.Split(value, " ")
	case "Also":
		install.Also = strings.Split(value, " ")
	case "DefaultInstance":
		install.DefaultInstance = value
	}
}

// generateNixOSServiceConfig generates NixOS configuration for mapped services
func (sc *SystemdConverter) generateNixOSServiceConfig(service *SystemdService, mapping detector.ServiceMapping) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("  # %s service configuration (mapped from systemd)\n", mapping.SourceName))
	config.WriteString(fmt.Sprintf("  %s = true;\n", mapping.NixOSOption))

	// Add service-specific configuration based on mapping
	switch mapping.NixOSService {
	case "nginx":
		sc.generateNginxConfig(&config, service)
	case "mysql":
		sc.generateMySQLConfig(&config, service)
	case "postgresql":
		sc.generatePostgreSQLConfig(&config, service)
	case "openssh":
		sc.generateSSHConfig(&config, service)
	default:
		sc.generateGenericServiceConfig(&config, service, mapping)
	}

	return config.String()
}

// generateCustomServiceConfig generates custom systemd service configuration
func (sc *SystemdConverter) generateCustomServiceConfig(service *SystemdService, serviceName string) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("  # Custom service configuration for %s\n", serviceName))
	config.WriteString(fmt.Sprintf("  systemd.services.%s = {\n", serviceName))

	if service.Unit.Description != "" {
		config.WriteString(fmt.Sprintf("    description = \"%s\";\n", service.Unit.Description))
	}

	if len(service.Unit.After) > 0 {
		config.WriteString(fmt.Sprintf("    after = [ %s ];\n", sc.formatStringList(service.Unit.After)))
	}

	if len(service.Unit.Wants) > 0 {
		config.WriteString(fmt.Sprintf("    wants = [ %s ];\n", sc.formatStringList(service.Unit.Wants)))
	}

	config.WriteString("    wantedBy = [ \"multi-user.target\" ];\n")
	config.WriteString("    serviceConfig = {\n")

	if service.Service.ExecStart != "" {
		config.WriteString(fmt.Sprintf("      ExecStart = \"%s\";\n", service.Service.ExecStart))
	}

	if service.Service.ExecStop != "" {
		config.WriteString(fmt.Sprintf("      ExecStop = \"%s\";\n", service.Service.ExecStop))
	}

	if service.Service.User != "" {
		config.WriteString(fmt.Sprintf("      User = \"%s\";\n", service.Service.User))
	}

	if service.Service.Group != "" {
		config.WriteString(fmt.Sprintf("      Group = \"%s\";\n", service.Service.Group))
	}

	if service.Service.WorkingDirectory != "" {
		config.WriteString(fmt.Sprintf("      WorkingDirectory = \"%s\";\n", service.Service.WorkingDirectory))
	}

	if service.Service.Type != "" {
		config.WriteString(fmt.Sprintf("      Type = \"%s\";\n", service.Service.Type))
	}

	if service.Service.Restart != "" {
		config.WriteString(fmt.Sprintf("      Restart = \"%s\";\n", service.Service.Restart))
	}

	if service.Service.RestartSec != "" {
		config.WriteString(fmt.Sprintf("      RestartSec = \"%s\";\n", service.Service.RestartSec))
	}

	// Add environment variables
	if len(service.Service.Environment) > 0 {
		config.WriteString("      Environment = [\n")
		for key, value := range service.Service.Environment {
			config.WriteString(fmt.Sprintf("        \"%s=%s\"\n", key, value))
		}
		config.WriteString("      ];\n")
	}

	config.WriteString("    };\n")
	config.WriteString("  };\n\n")

	return config.String()
}

// Service-specific configuration generators
func (sc *SystemdConverter) generateNginxConfig(config *strings.Builder, service *SystemdService) {
	config.WriteString("  # Additional nginx configuration may be needed\n")
	if service.Service.User != "" && service.Service.User != "nginx" {
		config.WriteString(fmt.Sprintf("  services.nginx.user = \"%s\";\n", service.Service.User))
	}
}

func (sc *SystemdConverter) generateMySQLConfig(config *strings.Builder, service *SystemdService) {
	config.WriteString("  # Additional MySQL configuration may be needed\n")
	if service.Service.User != "" && service.Service.User != "mysql" {
		config.WriteString(fmt.Sprintf("  # Custom user: %s (may need manual configuration)\n", service.Service.User))
	}
}

func (sc *SystemdConverter) generatePostgreSQLConfig(config *strings.Builder, service *SystemdService) {
	config.WriteString("  # Additional PostgreSQL configuration may be needed\n")
	if service.Service.User != "" && service.Service.User != "postgres" {
		config.WriteString(fmt.Sprintf("  # Custom user: %s (may need manual configuration)\n", service.Service.User))
	}
}

func (sc *SystemdConverter) generateSSHConfig(config *strings.Builder, service *SystemdService) {
	config.WriteString("  # SSH service configuration\n")
	// SSH usually doesn't need additional config from systemd
}

func (sc *SystemdConverter) generateGenericServiceConfig(config *strings.Builder, service *SystemdService, mapping detector.ServiceMapping) {
	config.WriteString(fmt.Sprintf("  # Additional configuration for %s may be needed\n", mapping.NixOSService))

	// Add environment variables if present
	if len(service.Service.Environment) > 0 {
		config.WriteString("  # Environment variables from systemd service:\n")
		for key, value := range service.Service.Environment {
			config.WriteString(fmt.Sprintf("  # %s=%s\n", key, value))
		}
	}
}

// addConversionWarnings adds warnings and manual steps based on service analysis
func (sc *SystemdConverter) addConversionWarnings(service *SystemdService, result *ConversionResult) {
	// Check for complex configurations
	if service.Service.EnvironmentFile != "" {
		result.Warnings = append(result.Warnings, "Environment file referenced - may need manual migration")
		result.ManualSteps = append(result.ManualSteps, fmt.Sprintf("Migrate environment file: %s", service.Service.EnvironmentFile))
	}

	if service.Service.PIDFile != "" {
		result.Warnings = append(result.Warnings, "Custom PID file location - may need adjustment")
	}

	if len(service.Unit.Requires) > 0 {
		result.Warnings = append(result.Warnings, "Service dependencies detected - verify startup order")
		result.Dependencies = append(result.Dependencies, service.Unit.Requires...)
	}

	if service.Service.Type == "forking" {
		result.Warnings = append(result.Warnings, "Forking service type - may need adjustment for NixOS")
	}

	if service.Service.ExecStop != "" {
		result.ManualSteps = append(result.ManualSteps, "Review custom stop command configuration")
	}

	if service.Service.ExecReload != "" {
		result.ManualSteps = append(result.ManualSteps, "Review custom reload command configuration")
	}

	// Check for non-standard users/groups
	if service.Service.User != "" && service.Service.User != "root" {
		result.ManualSteps = append(result.ManualSteps, fmt.Sprintf("Ensure user '%s' exists or configure user creation", service.Service.User))
	}

	if service.Service.Group != "" {
		result.ManualSteps = append(result.ManualSteps, fmt.Sprintf("Ensure group '%s' exists or configure group creation", service.Service.Group))
	}
}

// formatStringList formats a string slice for Nix configuration
func (sc *SystemdConverter) formatStringList(items []string) string {
	var quoted []string
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("\"%s\"", item))
	}
	return strings.Join(quoted, " ")
}

// ConvertSystemdServices converts multiple systemd services
func (sc *SystemdConverter) ConvertSystemdServices(services []detector.ServiceInfo) ([]ConversionResult, error) {
	var results []ConversionResult

	for _, service := range services {
		// Try to read systemd service file
		servicePath := fmt.Sprintf("/etc/systemd/system/%s.service", service.Name)
		if service.Config != "" {
			servicePath = service.Config
		}

		content, err := sc.readServiceFile(servicePath)
		if err != nil {
			sc.logger.Warn(fmt.Sprintf("Could not read systemd service file for %s: %v", service.Name, err))
			continue
		}

		result, err := sc.ConvertService(service.Name, content)
		if err != nil {
			sc.logger.Warn(fmt.Sprintf("Failed to convert service %s: %v", service.Name, err))
			continue
		}

		results = append(results, *result)
	}

	return results, nil
}

// readServiceFile reads a systemd service file
func (sc *SystemdConverter) readServiceFile(path string) (string, error) {
	// This is a placeholder - in a real implementation, you'd read the file
	// For now, return empty content to avoid file system dependencies
	return "", fmt.Errorf("service file reading not implemented in this example")
}

// GenerateMigrationSummary generates a summary of all service conversions
func (sc *SystemdConverter) GenerateMigrationSummary(results []ConversionResult) string {
	var summary strings.Builder

	summary.WriteString("# Systemd Service Migration Summary\n\n")
	summary.WriteString(fmt.Sprintf("Total services processed: %d\n\n", len(results)))

	// Count service types
	mappedServices := 0
	customServices := 0
	for _, result := range results {
		if result.ServiceMapping.NixOSService != "" {
			mappedServices++
		} else {
			customServices++
		}
	}

	summary.WriteString(fmt.Sprintf("- Mapped to NixOS services: %d\n", mappedServices))
	summary.WriteString(fmt.Sprintf("- Custom service configurations: %d\n\n", customServices))

	// List warnings and manual steps
	totalWarnings := 0
	totalManualSteps := 0
	for _, result := range results {
		totalWarnings += len(result.Warnings)
		totalManualSteps += len(result.ManualSteps)
	}

	summary.WriteString(fmt.Sprintf("Total warnings: %d\n", totalWarnings))
	summary.WriteString(fmt.Sprintf("Total manual steps required: %d\n\n", totalManualSteps))

	// Detailed results
	summary.WriteString("## Service Conversion Details\n\n")
	for _, result := range results {
		summary.WriteString(fmt.Sprintf("### %s\n", result.OriginalService))

		if result.ServiceMapping.NixOSService != "" {
			summary.WriteString(fmt.Sprintf("- Mapped to: %s\n", result.ServiceMapping.NixOSService))
		} else {
			summary.WriteString("- Custom service configuration generated\n")
		}

		if len(result.Warnings) > 0 {
			summary.WriteString("- Warnings:\n")
			for _, warning := range result.Warnings {
				summary.WriteString(fmt.Sprintf("  - %s\n", warning))
			}
		}

		if len(result.ManualSteps) > 0 {
			summary.WriteString("- Manual steps required:\n")
			for _, step := range result.ManualSteps {
				summary.WriteString(fmt.Sprintf("  - %s\n", step))
			}
		}

		summary.WriteString("\n")
	}

	return summary.String()
}
