package detector

import (
	"fmt"
	"strings"

	"nix-ai-help/pkg/logger"
)

// ServiceMapping represents a service mapping from source to NixOS
type ServiceMapping struct {
	SourceName   string            `json:"source_name"`
	NixOSService string            `json:"nixos_service"`
	NixOSOption  string            `json:"nixos_option"`
	ConfigMap    map[string]string `json:"config_mapping"`
	Dependencies []string          `json:"dependencies"`
	PortMapping  map[int]int       `json:"port_mapping"`
	Notes        string            `json:"notes"`
	Complexity   string            `json:"complexity"`
}

// ServiceMapper maps services from source systems to NixOS equivalents
type ServiceMapper struct {
	logger   logger.Logger
	mappings map[string]ServiceMapping
}

// NewServiceMapper creates a new service mapper
func NewServiceMapper(logger logger.Logger) *ServiceMapper {
	sm := &ServiceMapper{
		logger:   logger,
		mappings: make(map[string]ServiceMapping),
	}
	sm.initializeMappings()
	return sm
}

// initializeMappings initializes the service mapping database
func (sm *ServiceMapper) initializeMappings() {
	mappings := []ServiceMapping{
		// Web Servers
		{
			SourceName:   "nginx",
			NixOSService: "nginx",
			NixOSOption:  "services.nginx.enable",
			ConfigMap: map[string]string{
				"/etc/nginx/nginx.conf":        "services.nginx.config",
				"/etc/nginx/sites-available/*": "services.nginx.virtualHosts",
			},
			Dependencies: []string{"openssl"},
			PortMapping:  map[int]int{80: 80, 443: 443},
			Notes:        "Virtual hosts need individual configuration",
			Complexity:   "intermediate",
		},
		{
			SourceName:   "apache2",
			NixOSService: "httpd",
			NixOSOption:  "services.httpd.enable",
			ConfigMap: map[string]string{
				"/etc/apache2/apache2.conf":      "services.httpd.config",
				"/etc/apache2/sites-available/*": "services.httpd.virtualHosts",
			},
			Dependencies: []string{"openssl"},
			PortMapping:  map[int]int{80: 80, 443: 443},
			Notes:        "Module configuration differs significantly",
			Complexity:   "complex",
		},
		// Databases
		{
			SourceName:   "mysql",
			NixOSService: "mysql",
			NixOSOption:  "services.mysql.enable",
			ConfigMap: map[string]string{
				"/etc/mysql/my.cnf": "services.mysql.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{3306: 3306},
			Notes:        "Database migration requires separate data export/import",
			Complexity:   "complex",
		},
		{
			SourceName:   "mariadb",
			NixOSService: "mysql",
			NixOSOption:  "services.mysql.enable",
			ConfigMap: map[string]string{
				"/etc/mysql/mariadb.conf.d/50-server.cnf": "services.mysql.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{3306: 3306},
			Notes:        "Use MariaDB package, database migration required",
			Complexity:   "complex",
		},
		{
			SourceName:   "postgresql",
			NixOSService: "postgresql",
			NixOSOption:  "services.postgresql.enable",
			ConfigMap: map[string]string{
				"/etc/postgresql/*/main/postgresql.conf": "services.postgresql.config",
				"/etc/postgresql/*/main/pg_hba.conf":     "services.postgresql.authentication",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{5432: 5432},
			Notes:        "Database dump and restore required for migration",
			Complexity:   "complex",
		},
		// Key-Value Stores
		{
			SourceName:   "redis",
			NixOSService: "redis",
			NixOSOption:  "services.redis.servers.default.enable",
			ConfigMap: map[string]string{
				"/etc/redis/redis.conf": "services.redis.servers.default.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{6379: 6379},
			Notes:        "Redis data persistence location may need adjustment",
			Complexity:   "simple",
		},
		// SSH
		{
			SourceName:   "ssh",
			NixOSService: "openssh",
			NixOSOption:  "services.openssh.enable",
			ConfigMap: map[string]string{
				"/etc/ssh/sshd_config": "services.openssh.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{22: 22},
			Notes:        "SSH keys and authorized_keys need manual migration",
			Complexity:   "simple",
		},
		{
			SourceName:   "sshd",
			NixOSService: "openssh",
			NixOSOption:  "services.openssh.enable",
			ConfigMap: map[string]string{
				"/etc/ssh/sshd_config": "services.openssh.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{22: 22},
			Notes:        "SSH keys and authorized_keys need manual migration",
			Complexity:   "simple",
		},
		// Container Systems
		{
			SourceName:   "docker",
			NixOSService: "docker",
			NixOSOption:  "virtualisation.docker.enable",
			ConfigMap: map[string]string{
				"/etc/docker/daemon.json": "virtualisation.docker.daemon.settings",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{},
			Notes:        "Container images and volumes need manual migration",
			Complexity:   "intermediate",
		},
		// Monitoring
		{
			SourceName:   "prometheus",
			NixOSService: "prometheus",
			NixOSOption:  "services.prometheus.enable",
			ConfigMap: map[string]string{
				"/etc/prometheus/prometheus.yml": "services.prometheus.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{9090: 9090},
			Notes:        "Scrape configs and alerting rules need review",
			Complexity:   "intermediate",
		},
		{
			SourceName:   "grafana",
			NixOSService: "grafana",
			NixOSOption:  "services.grafana.enable",
			ConfigMap: map[string]string{
				"/etc/grafana/grafana.ini": "services.grafana.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{3000: 3000},
			Notes:        "Dashboards and data sources need manual migration",
			Complexity:   "intermediate",
		},
		// Mail Services
		{
			SourceName:   "postfix",
			NixOSService: "postfix",
			NixOSOption:  "services.postfix.enable",
			ConfigMap: map[string]string{
				"/etc/postfix/main.cf": "services.postfix.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{25: 25, 587: 587},
			Notes:        "Mail queue and aliases need manual migration",
			Complexity:   "complex",
		},
		// DNS
		{
			SourceName:   "bind9",
			NixOSService: "bind",
			NixOSOption:  "services.bind.enable",
			ConfigMap: map[string]string{
				"/etc/bind/named.conf": "services.bind.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{53: 53},
			Notes:        "Zone files need manual migration and validation",
			Complexity:   "complex",
		},
		// File Sharing
		{
			SourceName:   "samba",
			NixOSService: "samba",
			NixOSOption:  "services.samba.enable",
			ConfigMap: map[string]string{
				"/etc/samba/smb.conf": "services.samba.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{139: 139, 445: 445},
			Notes:        "User accounts and shares need manual configuration",
			Complexity:   "intermediate",
		},
		// System Services
		{
			SourceName:   "cron",
			NixOSService: "cron",
			NixOSOption:  "services.cron.enable",
			ConfigMap: map[string]string{
				"/etc/crontab":      "services.cron.systemCronJobs",
				"/var/spool/cron/*": "users.users.<name>.crontab",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{},
			Notes:        "User crontabs need individual migration",
			Complexity:   "simple",
		},
		{
			SourceName:   "systemd-timesyncd",
			NixOSService: "timesyncd",
			NixOSOption:  "services.timesyncd.enable",
			ConfigMap: map[string]string{
				"/etc/systemd/timesyncd.conf": "services.timesyncd.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{},
			Notes:        "NTP servers configuration translates directly",
			Complexity:   "simple",
		},
		// Development Tools
		{
			SourceName:   "jenkins",
			NixOSService: "jenkins",
			NixOSOption:  "services.jenkins.enable",
			ConfigMap: map[string]string{
				"/var/lib/jenkins/config.xml": "services.jenkins.config",
			},
			Dependencies: []string{"jdk"},
			PortMapping:  map[int]int{8080: 8080},
			Notes:        "Jobs and plugins need manual migration",
			Complexity:   "complex",
		},
		// Backup Services
		{
			SourceName:   "rsync",
			NixOSService: "rsyncd",
			NixOSOption:  "services.rsyncd.enable",
			ConfigMap: map[string]string{
				"/etc/rsyncd.conf": "services.rsyncd.config",
			},
			Dependencies: []string{},
			PortMapping:  map[int]int{873: 873},
			Notes:        "Backup scripts may need path adjustments",
			Complexity:   "simple",
		},
	}

	for _, mapping := range mappings {
		sm.mappings[mapping.SourceName] = mapping
	}
}

// MapService maps a source service to NixOS equivalent
func (sm *ServiceMapper) MapService(serviceName string) (ServiceMapping, bool) {
	// Direct mapping
	if mapping, exists := sm.mappings[serviceName]; exists {
		return mapping, true
	}

	// Try fuzzy matching
	serviceName = strings.ToLower(serviceName)
	for name, mapping := range sm.mappings {
		if strings.Contains(strings.ToLower(name), serviceName) ||
			strings.Contains(serviceName, strings.ToLower(name)) {
			return mapping, true
		}
	}

	return ServiceMapping{}, false
}

// MapServices maps multiple services to NixOS equivalents
func (sm *ServiceMapper) MapServices(services []ServiceInfo) map[string]ServiceMapping {
	result := make(map[string]ServiceMapping)

	for _, service := range services {
		if mapping, exists := sm.MapService(service.Name); exists {
			result[service.Name] = mapping
		}
	}

	return result
}

// GetAllMappings returns all available service mappings
func (sm *ServiceMapper) GetAllMappings() map[string]ServiceMapping {
	return sm.mappings
}

// GetSupportedServices returns a list of supported source services
func (sm *ServiceMapper) GetSupportedServices() []string {
	services := make([]string, 0, len(sm.mappings))
	for service := range sm.mappings {
		services = append(services, service)
	}
	return services
}

// GetMappingByNixOSService finds mapping by NixOS service name
func (sm *ServiceMapper) GetMappingByNixOSService(nixosService string) (ServiceMapping, bool) {
	for _, mapping := range sm.mappings {
		if mapping.NixOSService == nixosService {
			return mapping, true
		}
	}
	return ServiceMapping{}, false
}

// ValidateMapping validates a service mapping
func (sm *ServiceMapper) ValidateMapping(mapping ServiceMapping) []string {
	var warnings []string

	// Check if NixOS service exists (simplified validation)
	if mapping.NixOSService == "" {
		warnings = append(warnings, "No NixOS service specified")
	}

	if mapping.NixOSOption == "" {
		warnings = append(warnings, "No NixOS option specified")
	}

	if mapping.Complexity == "complex" || mapping.Complexity == "expert" {
		warnings = append(warnings, fmt.Sprintf("Service mapping complexity is %s - manual intervention may be required", mapping.Complexity))
	}

	if len(mapping.ConfigMap) == 0 {
		warnings = append(warnings, "No configuration file mappings specified")
	}

	return warnings
}

// GenerateNixOSConfig generates NixOS configuration snippet for a service
func (sm *ServiceMapper) GenerateNixOSConfig(mapping ServiceMapping, serviceInfo ServiceInfo) string {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("  # %s service configuration\n", mapping.SourceName))
	config.WriteString(fmt.Sprintf("  %s = true;\n", mapping.NixOSOption))

	// Add port configuration if available
	if len(mapping.PortMapping) > 0 {
		config.WriteString(fmt.Sprintf("  # Port configuration for %s\n", mapping.SourceName))
		for sourcePort, nixosPort := range mapping.PortMapping {
			if sourcePort != nixosPort {
				config.WriteString(fmt.Sprintf("  # Port %d mapped to %d\n", sourcePort, nixosPort))
			}
		}
	}

	// Add dependencies
	if len(mapping.Dependencies) > 0 {
		config.WriteString(fmt.Sprintf("  # Dependencies for %s\n", mapping.SourceName))
		for _, dep := range mapping.Dependencies {
			config.WriteString(fmt.Sprintf("  # Requires: %s\n", dep))
		}
	}

	// Add notes
	if mapping.Notes != "" {
		config.WriteString(fmt.Sprintf("  # NOTE: %s\n", mapping.Notes))
	}

	config.WriteString("\n")
	return config.String()
}
