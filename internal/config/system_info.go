// Package config provides configuration structures for nixai
package config

import "time"

// HardwareInfo represents hardware configuration information
type HardwareInfo struct {
	CPUModel    string `json:"cpu_model" yaml:"cpu_model"`
	CPUCores    string `json:"cpu_cores" yaml:"cpu_cores"`
	Memory      string `json:"memory" yaml:"memory"`
	DiskSpace   string `json:"disk_space" yaml:"disk_space"`
	GPUModel    string `json:"gpu_model,omitempty" yaml:"gpu_model,omitempty"`
	NetworkCard string `json:"network_card,omitempty" yaml:"network_card,omitempty"`
	Bluetooth   string `json:"bluetooth,omitempty" yaml:"bluetooth,omitempty"`
}

// NetworkInfo represents network configuration information
type NetworkInfo struct {
	Interfaces   []string `json:"interfaces" yaml:"interfaces"`
	DNSServers   []string `json:"dns_servers" yaml:"dns_servers"`
	Gateway      string   `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	SubnetMask   string   `json:"subnet_mask,omitempty" yaml:"subnet_mask,omitempty"`
	IPAddress    string   `json:"ip_address,omitempty" yaml:"ip_address,omitempty"`
	MACAddress   string   `json:"mac_address,omitempty" yaml:"mac_address,omitempty"`
	Hostname     string   `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Domain       string   `json:"domain,omitempty" yaml:"domain,omitempty"`
	NetworkType  string   `json:"network_type,omitempty" yaml:"network_type,omitempty"` // wired, wireless, vpn
	Connections  []string `json:"connections,omitempty" yaml:"connections,omitempty"`
}

// SecurityInfo represents security configuration information
type SecurityInfo struct {
	FirewallEnabled    bool     `json:"firewall_enabled" yaml:"firewall_enabled"`
	SSHEnabled         bool     `json:"ssh_enabled" yaml:"ssh_enabled"`
	SELinuxEnabled     bool     `json:"selinux_enabled,omitempty" yaml:"selinux_enabled,omitempty"`
	AppArmorEnabled    bool     `json:"apparmor_enabled,omitempty" yaml:"apparmor_enabled,omitempty"`
	SecurityPolicies   []string `json:"security_policies,omitempty" yaml:"security_policies,omitempty"`
	EncryptionEnabled  bool     `json:"encryption_enabled,omitempty" yaml:"encryption_enabled,omitempty"`
	AuthenticationType string   `json:"authentication_type,omitempty" yaml:"authentication_type,omitempty"` // password, certificate, biometric
	AuthorizedKeys     []string `json:"authorized_keys,omitempty" yaml:"authorized_keys,omitempty"`
	FailedLogins       int      `json:"failed_logins,omitempty" yaml:"failed_logins,omitempty"`
	LastLogin          time.Time `json:"last_login,omitempty" yaml:"last_login,omitempty"`
}

// PerformanceInfo represents system performance information
type PerformanceInfo struct {
	LoadAverage      string  `json:"load_average" yaml:"load_average"`
	Uptime           string  `json:"uptime" yaml:"uptime"`
	CPUUsage         float64 `json:"cpu_usage" yaml:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage" yaml:"memory_usage"`
	DiskIO           string  `json:"disk_io,omitempty" yaml:"disk_io,omitempty"`
	NetworkBandwidth string  `json:"network_bandwidth,omitempty" yaml:"network_bandwidth,omitempty"`
	Processes        int     `json:"processes,omitempty" yaml:"processes,omitempty"`
	Threads          int     `json:"threads,omitempty" yaml:"threads,omitempty"`
	FileDescriptors   int     `json:"file_descriptors,omitempty" yaml:"file_descriptors,omitempty"`
}

// UserEnvironment represents user environment information
type UserEnvironment struct {
	Shell              string   `json:"shell" yaml:"shell"`
	Editor             string   `json:"editor,omitempty" yaml:"editor,omitempty"`
	Terminal           string   `json:"terminal,omitempty" yaml:"terminal,omitempty"`
	DesktopEnvironment string   `json:"desktop_environment,omitempty" yaml:"desktop_environment,omitempty"`
	WindowManager      string   `json:"window_manager,omitempty" yaml:"window_manager,omitempty"`
	DisplayServer      string   `json:"display_server,omitempty" yaml:"display_server,omitempty"`
	Locale             string   `json:"locale,omitempty" yaml:"locale,omitempty"`
	Timezone           string   `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	PATH               string   `json:"path,omitempty" yaml:"path,omitempty"`
	EnvironmentVars    []string `json:"environment_vars,omitempty" yaml:"environment_vars,omitempty"`
	UserGroups         []string `json:"user_groups,omitempty" yaml:"user_groups,omitempty"`
}