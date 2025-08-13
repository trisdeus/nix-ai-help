// Package hardware provides enhanced hardware detection and profiling utilities
package hardware

import (
	"context"
	"fmt"
	"time"

	"nix-ai-help/pkg/logger"
)

// EnhancedHardwareDetector provides comprehensive hardware detection and profiling
type EnhancedHardwareDetector struct {
	logger           *logger.Logger
	cache            *HardwareCache
	profilerEnabled  bool
	detectionTimeout time.Duration
}

// HardwareCache stores detection results to avoid repeated system calls
type HardwareCache struct {
	data      map[string]CacheEntry
	ttl       time.Duration
	enabled   bool
	lastClean time.Time
}

// CacheEntry represents a cached hardware detection result
type CacheEntry struct {
	Data      interface{}
	Timestamp time.Time
	TTL       time.Duration
}

// EnhancedHardwareInfo extends HardwareInfo with detailed profiling data
type EnhancedHardwareInfo struct {
	*HardwareInfo
	SystemProfile     *SystemProfile     `json:"system_profile"`
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics"`
	ThermalProfile    *ThermalProfile    `json:"thermal_profile"`
	PowerProfile      *PowerProfile      `json:"power_profile"`
	CompatibilityInfo *CompatibilityInfo `json:"compatibility_info"`
	RecommendedConfig *RecommendedConfig `json:"recommended_config"`
	BenchmarkResults  *BenchmarkResults  `json:"benchmark_results,omitempty"`
	SecurityFeatures  *SecurityFeatures  `json:"security_features"`
	ConnectivityInfo  *ConnectivityInfo  `json:"connectivity_info"`
	DetectionMetadata *DetectionMetadata `json:"detection_metadata"`
}

// SystemProfile contains detailed system profiling information
type SystemProfile struct {
	CPUDetails      *CPUProfile      `json:"cpu_details"`
	GPUDetails      []*GPUProfile    `json:"gpu_details"`
	MemoryDetails   *MemoryProfile   `json:"memory_details"`
	StorageDetails  []*StorageProfile `json:"storage_details"`
	NetworkDetails  []*NetworkProfile `json:"network_details"`
	AudioDetails    *AudioProfile    `json:"audio_details"`
	DisplayDetails  *DisplayProfile  `json:"display_details"`
	InputDevices    []*InputProfile  `json:"input_devices"`
	USBDevices      []*USBProfile    `json:"usb_devices"`
	BluetoothDevices []*BluetoothProfile `json:"bluetooth_devices"`
}

// CPUProfile contains detailed CPU information
type CPUProfile struct {
	Vendor          string            `json:"vendor"`
	Model           string            `json:"model"`
	Family          string            `json:"family"`
	Stepping        string            `json:"stepping"`
	Microcode       string            `json:"microcode"`
	Architecture    string            `json:"architecture"`
	Cores           int               `json:"cores"`
	Threads         int               `json:"threads"`
	BaseFrequency   float64           `json:"base_frequency_mhz"`
	MaxFrequency    float64           `json:"max_frequency_mhz"`
	CacheL1         string            `json:"cache_l1"`
	CacheL2         string            `json:"cache_l2"`
	CacheL3         string            `json:"cache_l3"`
	Features        []string          `json:"features"`
	VirtualizationSupport bool        `json:"virtualization_support"`
	SecurityFeatures []string         `json:"security_features"`
	PowerManagement []string          `json:"power_management"`
	NUMA            bool              `json:"numa"`
	SocketType      string            `json:"socket_type"`
	TDP             int               `json:"tdp_watts,omitempty"`
	ManufactureNode string            `json:"manufacture_node,omitempty"`
	Vulnerabilities map[string]string `json:"vulnerabilities,omitempty"`
}

// GPUProfile contains detailed GPU information
type GPUProfile struct {
	Vendor         string            `json:"vendor"`
	Model          string            `json:"model"`
	DeviceID       string            `json:"device_id"`
	SubsystemID    string            `json:"subsystem_id"`
	Driver         string            `json:"driver"`
	DriverVersion  string            `json:"driver_version"`
	Memory         string            `json:"memory"`
	MemoryType     string            `json:"memory_type"`
	BusInterface   string            `json:"bus_interface"`
	Architecture   string            `json:"architecture"`
	ComputeUnits   int               `json:"compute_units,omitempty"`
	ShaderUnits    int               `json:"shader_units,omitempty"`
	BaseFrequency  float64           `json:"base_frequency_mhz,omitempty"`
	BoostFrequency float64           `json:"boost_frequency_mhz,omitempty"`
	MemoryBandwidth float64          `json:"memory_bandwidth_gbps,omitempty"`
	PowerDraw      int               `json:"power_draw_watts,omitempty"`
	CUDASupport    bool              `json:"cuda_support"`
	OpenCLSupport  bool              `json:"opencl_support"`
	VulkanSupport  bool              `json:"vulkan_support"`
	DirectXSupport string            `json:"directx_support"`
	OpenGLSupport  string            `json:"opengl_support"`
	DisplayPorts   []string          `json:"display_ports"`
	MaxResolution  string            `json:"max_resolution"`
	HDRSupport     bool              `json:"hdr_support"`
	Capabilities   map[string]string `json:"capabilities,omitempty"`
}

// MemoryProfile contains detailed memory information
type MemoryProfile struct {
	TotalCapacity    string              `json:"total_capacity"`
	AvailableCapacity string             `json:"available_capacity"`
	UsedCapacity     string              `json:"used_capacity"`
	MemoryType       string              `json:"memory_type"`
	Speed            int                 `json:"speed_mhz"`
	Channels         int                 `json:"channels"`
	Ranks            int                 `json:"ranks"`
	Slots            []*MemorySlot       `json:"slots"`
	ECC              bool                `json:"ecc_support"`
	MaxCapacity      string              `json:"max_capacity"`
	FormFactor       string              `json:"form_factor"`
	Manufacturer     string              `json:"manufacturer"`
	SerialNumber     string              `json:"serial_number,omitempty"`
	PartNumber       string              `json:"part_number,omitempty"`
	Bandwidth        float64             `json:"bandwidth_gbps,omitempty"`
	Latency          map[string]int      `json:"latency_timings,omitempty"`
	VoltageProfile   map[string]float64  `json:"voltage_profile,omitempty"`
}

// MemorySlot represents individual memory slot information
type MemorySlot struct {
	SlotNumber   int    `json:"slot_number"`
	Capacity     string `json:"capacity"`
	Type         string `json:"type"`
	Speed        int    `json:"speed_mhz"`
	Manufacturer string `json:"manufacturer"`
	PartNumber   string `json:"part_number"`
	SerialNumber string `json:"serial_number,omitempty"`
	BankLocator  string `json:"bank_locator"`
	Populated    bool   `json:"populated"`
}

// StorageProfile contains detailed storage device information
type StorageProfile struct {
	DeviceName      string            `json:"device_name"`
	DevicePath      string            `json:"device_path"`
	Type            string            `json:"type"` // HDD, SSD, NVMe, etc.
	Capacity        string            `json:"capacity"`
	Model           string            `json:"model"`
	Vendor          string            `json:"vendor"`
	SerialNumber    string            `json:"serial_number,omitempty"`
	FirmwareVersion string            `json:"firmware_version,omitempty"`
	Interface       string            `json:"interface"` // SATA, NVMe, USB, etc.
	Speed           string            `json:"speed,omitempty"`
	FormFactor      string            `json:"form_factor,omitempty"`
	RotationSpeed   int               `json:"rotation_speed_rpm,omitempty"`
	CacheSize       string            `json:"cache_size,omitempty"`
	Health          string            `json:"health_status"`
	Temperature     float64           `json:"temperature_celsius,omitempty"`
	PowerOnHours    int               `json:"power_on_hours,omitempty"`
	TotalWrites     string            `json:"total_writes,omitempty"`
	SMARTData       map[string]string `json:"smart_data,omitempty"`
	FileSystem      string            `json:"filesystem,omitempty"`
	MountPoint      string            `json:"mount_point,omitempty"`
	EncryptionStatus string           `json:"encryption_status,omitempty"`
	Partitions      []*PartitionInfo  `json:"partitions,omitempty"`
}

// PartitionInfo represents partition information
type PartitionInfo struct {
	Name        string `json:"name"`
	Size        string `json:"size"`
	Used        string `json:"used"`
	Available   string `json:"available"`
	UsagePercent float64 `json:"usage_percent"`
	FileSystem  string `json:"filesystem"`
	MountPoint  string `json:"mount_point"`
	UUID        string `json:"uuid,omitempty"`
	Label       string `json:"label,omitempty"`
}

// NetworkProfile contains detailed network interface information
type NetworkProfile struct {
	InterfaceName   string            `json:"interface_name"`
	Type            string            `json:"type"` // Ethernet, WiFi, Bluetooth, etc.
	MACAddress      string            `json:"mac_address"`
	Driver          string            `json:"driver"`
	DriverVersion   string            `json:"driver_version,omitempty"`
	Speed           string            `json:"speed,omitempty"`
	Duplex          string            `json:"duplex,omitempty"`
	State           string            `json:"state"`
	MTU             int               `json:"mtu"`
	IPAddresses     []string          `json:"ip_addresses,omitempty"`
	IPv6Addresses   []string          `json:"ipv6_addresses,omitempty"`
	Gateway         string            `json:"gateway,omitempty"`
	DNSServers      []string          `json:"dns_servers,omitempty"`
	WirelessInfo    *WirelessProfile  `json:"wireless_info,omitempty"`
	VendorInfo      string            `json:"vendor_info,omitempty"`
	PCIInfo         string            `json:"pci_info,omitempty"`
	Capabilities    []string          `json:"capabilities,omitempty"`
	Statistics      map[string]uint64 `json:"statistics,omitempty"`
	PowerManagement bool              `json:"power_management"`
	WakeOnLAN       bool              `json:"wake_on_lan"`
}

// WirelessProfile contains WiFi-specific information
type WirelessProfile struct {
	SSID            string   `json:"ssid,omitempty"`
	BSSID           string   `json:"bssid,omitempty"`
	Frequency       string   `json:"frequency,omitempty"`
	Channel         int      `json:"channel,omitempty"`
	SignalStrength  int      `json:"signal_strength_dbm,omitempty"`
	LinkQuality     int      `json:"link_quality,omitempty"`
	SecurityType    string   `json:"security_type,omitempty"`
	Standards       []string `json:"standards,omitempty"`
	BitRate         string   `json:"bit_rate,omitempty"`
	TxPower         string   `json:"tx_power,omitempty"`
	SupportedBands  []string `json:"supported_bands,omitempty"`
	ChannelWidth    string   `json:"channel_width,omitempty"`
}

// AudioProfile contains detailed audio system information
type AudioProfile struct {
	Cards          []*AudioCard     `json:"cards"`
	DefaultSink    string           `json:"default_sink,omitempty"`
	DefaultSource  string           `json:"default_source,omitempty"`
	SoundServer    string           `json:"sound_server"` // PulseAudio, PipeWire, ALSA
	ServerVersion  string           `json:"server_version,omitempty"`
	SampleRate     int              `json:"sample_rate_hz,omitempty"`
	BufferSize     int              `json:"buffer_size,omitempty"`
	Latency        float64          `json:"latency_ms,omitempty"`
	ActiveProfiles []string         `json:"active_profiles,omitempty"`
	Devices        []*AudioDevice   `json:"devices"`
}

// AudioCard represents an audio card
type AudioCard struct {
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	Driver      string   `json:"driver"`
	LongName    string   `json:"long_name,omitempty"`
	Mixer       []string `json:"mixer,omitempty"`
	Components  []string `json:"components,omitempty"`
	PCMDevices  []string `json:"pcm_devices,omitempty"`
	MIDIDevices []string `json:"midi_devices,omitempty"`
}

// AudioDevice represents an audio input/output device
type AudioDevice struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Type         string            `json:"type"` // sink, source
	Driver       string            `json:"driver"`
	SampleSpec   string            `json:"sample_spec,omitempty"`
	Channels     int               `json:"channels,omitempty"`
	State        string            `json:"state"`
	Volume       map[string]string `json:"volume,omitempty"`
	Muted        bool              `json:"muted"`
	Ports        []string          `json:"ports,omitempty"`
	ActivePort   string            `json:"active_port,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// DisplayProfile contains display system information
type DisplayProfile struct {
	Server         string            `json:"server"` // X11, Wayland
	ServerVersion  string            `json:"server_version,omitempty"`
	Displays       []*Display        `json:"displays"`
	Resolution     string            `json:"total_resolution"`
	ColorDepth     int               `json:"color_depth"`
	RefreshRate    float64           `json:"refresh_rate_hz,omitempty"`
	DPI            map[string]int    `json:"dpi,omitempty"`
	Compositor     string            `json:"compositor,omitempty"`
	DesktopEnvironment string        `json:"desktop_environment,omitempty"`
	WindowManager  string            `json:"window_manager,omitempty"`
	ThemeInfo      map[string]string `json:"theme_info,omitempty"`
}

// Display represents an individual display
type Display struct {
	Name         string  `json:"name"`
	Connected    bool    `json:"connected"`
	Primary      bool    `json:"primary"`
	Resolution   string  `json:"resolution"`
	PhysicalSize string  `json:"physical_size_mm,omitempty"`
	RefreshRate  float64 `json:"refresh_rate_hz"`
	ColorDepth   int     `json:"color_depth"`
	Brightness   float64 `json:"brightness_percent,omitempty"`
	Rotation     string  `json:"rotation,omitempty"`
	Position     string  `json:"position,omitempty"`
	EDID         string  `json:"edid,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	Model        string  `json:"model,omitempty"`
	SerialNumber string  `json:"serial_number,omitempty"`
}

// InputProfile contains input device information
type InputProfile struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // keyboard, mouse, touchpad, etc.
	Path         string            `json:"path"`
	Vendor       string            `json:"vendor,omitempty"`
	Product      string            `json:"product,omitempty"`
	Version      string            `json:"version,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
	Connected    bool              `json:"connected"`
	Battery      *BatteryInfo      `json:"battery,omitempty"`
}

// USBProfile contains USB device information
type USBProfile struct {
	DeviceID     string            `json:"device_id"`
	VendorID     string            `json:"vendor_id"`
	ProductID    string            `json:"product_id"`
	Vendor       string            `json:"vendor"`
	Product      string            `json:"product"`
	SerialNumber string            `json:"serial_number,omitempty"`
	Version      string            `json:"usb_version"`
	Speed        string            `json:"speed"`
	Power        string            `json:"power_consumption,omitempty"`
	Class        string            `json:"device_class"`
	Protocol     string            `json:"protocol,omitempty"`
	Driver       string            `json:"driver,omitempty"`
	Interfaces   []string          `json:"interfaces,omitempty"`
	Endpoints    []string          `json:"endpoints,omitempty"`
	Path         string            `json:"path"`
	Connected    bool              `json:"connected"`
}

// BluetoothProfile contains Bluetooth device information
type BluetoothProfile struct {
	Address      string            `json:"address"`
	Name         string            `json:"name"`
	Alias        string            `json:"alias,omitempty"`
	Class        string            `json:"class,omitempty"`
	Paired       bool              `json:"paired"`
	Connected    bool              `json:"connected"`
	Trusted      bool              `json:"trusted"`
	Blocked      bool              `json:"blocked"`
	RSSI         int               `json:"rssi_dbm,omitempty"`
	TxPower      int               `json:"tx_power_dbm,omitempty"`
	UUIDs        []string          `json:"uuids,omitempty"`
	Services     []string          `json:"services,omitempty"`
	Manufacturer string            `json:"manufacturer,omitempty"`
	Version      string            `json:"version,omitempty"`
	Features     []string          `json:"features,omitempty"`
	Battery      *BatteryInfo      `json:"battery,omitempty"`
}

// BatteryInfo contains battery information for devices
type BatteryInfo struct {
	Level        int    `json:"level_percent"`
	Status       string `json:"status"` // charging, discharging, full, etc.
	Technology   string `json:"technology,omitempty"`
	Voltage      float64 `json:"voltage_v,omitempty"`
	Current      float64 `json:"current_ma,omitempty"`
	Capacity     int    `json:"capacity_mah,omitempty"`
	Health       string `json:"health,omitempty"`
	TimeRemaining int   `json:"time_remaining_minutes,omitempty"`
}

// PerformanceMetrics contains system performance information
type PerformanceMetrics struct {
	CPUUsage        float64           `json:"cpu_usage_percent"`
	MemoryUsage     float64           `json:"memory_usage_percent"`
	SwapUsage       float64           `json:"swap_usage_percent"`
	LoadAverage     [3]float64        `json:"load_average"` // 1min, 5min, 15min
	ProcessCount    int               `json:"process_count"`
	ThreadCount     int               `json:"thread_count"`
	OpenFiles       int               `json:"open_files"`
	NetworkBandwidth map[string]float64 `json:"network_bandwidth_mbps,omitempty"`
	DiskIOPS        map[string]float64 `json:"disk_iops,omitempty"`
	DiskBandwidth   map[string]float64 `json:"disk_bandwidth_mbps,omitempty"`
	ContextSwitches uint64            `json:"context_switches_per_sec,omitempty"`
	Interrupts      uint64            `json:"interrupts_per_sec,omitempty"`
	SystemCalls     uint64            `json:"system_calls_per_sec,omitempty"`
	BootTime        time.Time         `json:"boot_time"`
	Uptime          time.Duration     `json:"uptime"`
}

// ThermalProfile contains thermal management information
type ThermalProfile struct {
	CPUTemperature    map[string]float64 `json:"cpu_temperature_celsius,omitempty"`
	GPUTemperature    map[string]float64 `json:"gpu_temperature_celsius,omitempty"`
	SystemTemperature map[string]float64 `json:"system_temperature_celsius,omitempty"`
	ThermalZones      []*ThermalZone     `json:"thermal_zones,omitempty"`
	FanSpeeds         map[string]int     `json:"fan_speeds_rpm,omitempty"`
	ThermalThrottling bool               `json:"thermal_throttling"`
	CoolingDevices    []*CoolingDevice   `json:"cooling_devices,omitempty"`
	ThermalPolicy     string             `json:"thermal_policy,omitempty"`
}

// ThermalZone represents a thermal zone
type ThermalZone struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Temperature  float64 `json:"temperature_celsius"`
	TripPoints   []int   `json:"trip_points_celsius,omitempty"`
	CriticalTemp int     `json:"critical_temp_celsius,omitempty"`
	Policy       string  `json:"policy,omitempty"`
}

// CoolingDevice represents a cooling device
type CoolingDevice struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	CurrentState int   `json:"current_state"`
	MaxState    int    `json:"max_state"`
	Statistics  map[string]interface{} `json:"statistics,omitempty"`
}

// PowerProfile contains power management information
type PowerProfile struct {
	ACPowerConnected  bool                `json:"ac_power_connected"`
	BatteryInfo       []*BatteryInfo      `json:"battery_info,omitempty"`
	PowerProfile      string              `json:"power_profile"`
	CPUGovernor       string              `json:"cpu_governor"`
	CPUFrequencies    map[string]float64  `json:"cpu_frequencies_mhz,omitempty"`
	PowerConsumption  map[string]float64  `json:"power_consumption_watts,omitempty"`
	SuspendSupport    []string            `json:"suspend_support,omitempty"`
	WakeupDevices     []string            `json:"wakeup_devices,omitempty"`
	PowerStates       map[string]string   `json:"power_states,omitempty"`
	EnergyEfficiency  map[string]float64  `json:"energy_efficiency,omitempty"`
}

// CompatibilityInfo contains hardware compatibility information
type CompatibilityInfo struct {
	NixOSCompatibility  map[string]string `json:"nixos_compatibility"`
	LinuxKernelSupport  map[string]string `json:"linux_kernel_support"`
	RequiredFirmware    []string          `json:"required_firmware,omitempty"`
	ProprietaryDrivers  []string          `json:"proprietary_drivers,omitempty"`
	KnownIssues         []string          `json:"known_issues,omitempty"`
	Workarounds         map[string]string `json:"workarounds,omitempty"`
	HardwareSupport     map[string]string `json:"hardware_support"`
	RecommendedKernel   string            `json:"recommended_kernel,omitempty"`
	TestingStatus       map[string]string `json:"testing_status,omitempty"`
}

// RecommendedConfig contains recommended NixOS configuration
type RecommendedConfig struct {
	KernelModules       []string          `json:"kernel_modules"`
	InitrdModules       []string          `json:"initrd_modules"`
	KernelParameters    []string          `json:"kernel_parameters"`
	HardwareSettings    map[string]string `json:"hardware_settings"`
	ServicesConfig      map[string]string `json:"services_config"`
	NetworkingConfig    map[string]string `json:"networking_config"`
	PowerManagementConfig map[string]string `json:"power_management_config"`
	SecurityConfig      map[string]string `json:"security_config"`
	OptimizationSettings map[string]string `json:"optimization_settings"`
	FirmwarePackages    []string          `json:"firmware_packages"`
	DriverPackages      []string          `json:"driver_packages"`
	AdditionalPackages  []string          `json:"additional_packages"`
	ConfigSnippets      map[string]string `json:"config_snippets"`
}

// BenchmarkResults contains hardware benchmark information
type BenchmarkResults struct {
	CPUBenchmarks     map[string]float64 `json:"cpu_benchmarks,omitempty"`
	MemoryBenchmarks  map[string]float64 `json:"memory_benchmarks,omitempty"`
	StorageBenchmarks map[string]float64 `json:"storage_benchmarks,omitempty"`
	NetworkBenchmarks map[string]float64 `json:"network_benchmarks,omitempty"`
	GPUBenchmarks     map[string]float64 `json:"gpu_benchmarks,omitempty"`
	OverallScore      float64            `json:"overall_score,omitempty"`
	ComparisonData    map[string]string  `json:"comparison_data,omitempty"`
	TestConditions    map[string]string  `json:"test_conditions,omitempty"`
	BenchmarkVersion  string             `json:"benchmark_version,omitempty"`
}

// SecurityFeatures contains security-related hardware information
type SecurityFeatures struct {
	SecureBoot         bool              `json:"secure_boot"`
	TPMVersion         string            `json:"tpm_version,omitempty"`
	VirtualizationSecurity bool          `json:"virtualization_security"`
	MEIStatus          string            `json:"mei_status,omitempty"`
	SGXSupport         bool              `json:"sgx_support"`
	CETSupport         bool              `json:"cet_support"`
	SMEPSupport        bool              `json:"smep_support"`
	SMAPSupport        bool              `json:"smap_support"`
	IBRSSupport        bool              `json:"ibrs_support"`
	IBPBSupport        bool              `json:"ibpb_support"`
	STIBPSupport       bool              `json:"stibp_support"`
	SSBDSupport        bool              `json:"ssbd_support"`
	L1TFMitigation     string            `json:"l1tf_mitigation,omitempty"`
	MDSMitigation      string            `json:"mds_mitigation,omitempty"`
	TAASMitigation     string            `json:"taas_mitigation,omitempty"`
	ITLBMultihit       string            `json:"itlb_multihit,omitempty"`
	SpeculativeExecution map[string]string `json:"speculative_execution,omitempty"`
	Vulnerabilities    map[string]string `json:"vulnerabilities,omitempty"`
}

// ConnectivityInfo contains connectivity and interface information
type ConnectivityInfo struct {
	WiFiCapabilities   *WiFiCapabilities   `json:"wifi_capabilities,omitempty"`
	BluetoothInfo      *BluetoothInfo      `json:"bluetooth_info,omitempty"`
	EthernetInfo       *EthernetInfo       `json:"ethernet_info,omitempty"`
	WirelessInterfaces []*WirelessInterface `json:"wireless_interfaces,omitempty"`
	ModemInfo          *ModemInfo          `json:"modem_info,omitempty"`
	USBControllers     []*USBController    `json:"usb_controllers,omitempty"`
	NFCSupport         bool                `json:"nfc_support"`
	IRSupport          bool                `json:"ir_support"`
}

// WiFiCapabilities contains WiFi capability information
type WiFiCapabilities struct {
	Standards      []string `json:"standards"`
	MaxSpeed       string   `json:"max_speed"`
	Bands          []string `json:"bands"`
	Channels       []int    `json:"channels"`
	Antennas       int      `json:"antennas"`
	MIMOSupport    string   `json:"mimo_support,omitempty"`
	BeamForming    bool     `json:"beam_forming"`
	WPA3Support    bool     `json:"wpa3_support"`
	MeshSupport    bool     `json:"mesh_support"`
	HotspotSupport bool     `json:"hotspot_support"`
}

// BluetoothInfo contains Bluetooth capability information
type BluetoothInfo struct {
	Version        string   `json:"version"`
	LowEnergySupport bool   `json:"low_energy_support"`
	ClassicSupport bool     `json:"classic_support"`
	Range          string   `json:"range,omitempty"`
	Profiles       []string `json:"profiles"`
	Codecs         []string `json:"codecs,omitempty"`
	MaxConnections int      `json:"max_connections,omitempty"`
}

// EthernetInfo contains Ethernet capability information
type EthernetInfo struct {
	MaxSpeed       string   `json:"max_speed"`
	Standards      []string `json:"standards"`
	AutoNegotiation bool    `json:"auto_negotiation"`
	FlowControl    bool     `json:"flow_control"`
	JumboFrames    bool     `json:"jumbo_frames"`
	WakeOnLAN      bool     `json:"wake_on_lan"`
	VLANSupport    bool     `json:"vlan_support"`
}

// WirelessInterface contains wireless interface information
type WirelessInterface struct {
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Standards      []string `json:"standards"`
	FrequencyBands []string `json:"frequency_bands"`
	MaxSpeed       string   `json:"max_speed"`
	TxPowerLevels  []string `json:"tx_power_levels,omitempty"`
	Antennas       int      `json:"antennas"`
	Modes          []string `json:"modes"` // station, ap, monitor, etc.
}

// ModemInfo contains cellular modem information
type ModemInfo struct {
	Manufacturer string   `json:"manufacturer,omitempty"`
	Model        string   `json:"model,omitempty"`
	Revision     string   `json:"revision,omitempty"`
	Technologies []string `json:"technologies,omitempty"` // 3G, 4G, 5G
	Bands        []string `json:"bands,omitempty"`
	SIMStatus    string   `json:"sim_status,omitempty"`
	SignalStrength int    `json:"signal_strength,omitempty"`
	NetworkOperator string `json:"network_operator,omitempty"`
}

// USBController contains USB controller information
type USBController struct {
	Version      string `json:"version"`
	Speed        string `json:"speed"`
	Ports        int    `json:"ports"`
	PowerOutput  string `json:"power_output,omitempty"`
	Driver       string `json:"driver"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// DetectionMetadata contains metadata about the detection process
type DetectionMetadata struct {
	DetectionTime     time.Time     `json:"detection_time"`
	DetectionDuration time.Duration `json:"detection_duration"`
	DetectorVersion   string        `json:"detector_version"`
	OSVersion         string        `json:"os_version"`
	KernelVersion     string        `json:"kernel_version"`
	CacheHits         int           `json:"cache_hits"`
	CacheMisses       int           `json:"cache_misses"`
	ErrorsEncountered []string      `json:"errors_encountered,omitempty"`
	DataSources       []string      `json:"data_sources"`
	ReliabilityScore  float64       `json:"reliability_score"`
	CompletionRate    float64       `json:"completion_rate"`
}

// NewEnhancedHardwareDetector creates a new enhanced hardware detector
func NewEnhancedHardwareDetector(logger *logger.Logger) *EnhancedHardwareDetector {
	return &EnhancedHardwareDetector{
		logger:           logger,
		cache:            NewHardwareCache(5 * time.Minute),
		profilerEnabled:  true,
		detectionTimeout: 30 * time.Second,
	}
}

// NewHardwareCache creates a new hardware cache
func NewHardwareCache(ttl time.Duration) *HardwareCache {
	return &HardwareCache{
		data:      make(map[string]CacheEntry),
		ttl:       ttl,
		enabled:   true,
		lastClean: time.Now(),
	}
}

// EnableProfiling enables or disables detailed profiling
func (ehd *EnhancedHardwareDetector) EnableProfiling(enabled bool) {
	ehd.profilerEnabled = enabled
}

// SetDetectionTimeout sets the timeout for hardware detection operations
func (ehd *EnhancedHardwareDetector) SetDetectionTimeout(timeout time.Duration) {
	ehd.detectionTimeout = timeout
}

// DetectEnhancedHardware performs comprehensive hardware detection and profiling
func (ehd *EnhancedHardwareDetector) DetectEnhancedHardware(ctx context.Context) (*EnhancedHardwareInfo, error) {
	startTime := time.Now()
	ehd.logger.Info("Starting enhanced hardware detection")

	// Create detection context with timeout
	detectionCtx, cancel := context.WithTimeout(ctx, ehd.detectionTimeout)
	defer cancel()

	// Perform basic hardware detection first
	basicInfo, err := DetectHardwareComponents()
	if err != nil {
		return nil, fmt.Errorf("basic hardware detection failed: %v", err)
	}

	// Create enhanced info structure
	enhancedInfo := &EnhancedHardwareInfo{
		HardwareInfo: basicInfo,
		DetectionMetadata: &DetectionMetadata{
			DetectionTime:     startTime,
			DetectorVersion:   "2.0.0",
			DataSources:       []string{"proc", "sys", "lspci", "lsusb", "dmidecode"},
			ErrorsEncountered: []string{},
		},
	}

	// Perform detailed profiling if enabled
	if ehd.profilerEnabled {
		ehd.logger.Info("Performing detailed hardware profiling")
		
		// System profiling
		systemProfile, err := ehd.detectSystemProfile(detectionCtx)
		if err != nil {
			ehd.logger.Warn(fmt.Sprintf("System profiling failed: %v", err))
			enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
				enhancedInfo.DetectionMetadata.ErrorsEncountered, 
				fmt.Sprintf("system profiling: %v", err),
			)
		} else {
			enhancedInfo.SystemProfile = systemProfile
		}

		// Performance metrics
		performanceMetrics, err := ehd.detectPerformanceMetrics(detectionCtx)
		if err != nil {
			ehd.logger.Warn(fmt.Sprintf("Performance metrics detection failed: %v", err))
			enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
				enhancedInfo.DetectionMetadata.ErrorsEncountered, 
				fmt.Sprintf("performance metrics: %v", err),
			)
		} else {
			enhancedInfo.PerformanceMetrics = performanceMetrics
		}

		// Thermal profiling
		thermalProfile, err := ehd.detectThermalProfile(detectionCtx)
		if err != nil {
			ehd.logger.Warn(fmt.Sprintf("Thermal profiling failed: %v", err))
			enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
				enhancedInfo.DetectionMetadata.ErrorsEncountered, 
				fmt.Sprintf("thermal profiling: %v", err),
			)
		} else {
			enhancedInfo.ThermalProfile = thermalProfile
		}

		// Power profiling
		powerProfile, err := ehd.detectPowerProfile(detectionCtx)
		if err != nil {
			ehd.logger.Warn(fmt.Sprintf("Power profiling failed: %v", err))
			enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
				enhancedInfo.DetectionMetadata.ErrorsEncountered, 
				fmt.Sprintf("power profiling: %v", err),
			)
		} else {
			enhancedInfo.PowerProfile = powerProfile
		}

		// Security features
		securityFeatures, err := ehd.detectSecurityFeatures(detectionCtx)
		if err != nil {
			ehd.logger.Warn(fmt.Sprintf("Security features detection failed: %v", err))
			enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
				enhancedInfo.DetectionMetadata.ErrorsEncountered, 
				fmt.Sprintf("security features: %v", err),
			)
		} else {
			enhancedInfo.SecurityFeatures = securityFeatures
		}

		// Connectivity info
		connectivityInfo, err := ehd.detectConnectivityInfo(detectionCtx)
		if err != nil {
			ehd.logger.Warn(fmt.Sprintf("Connectivity detection failed: %v", err))
			enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
				enhancedInfo.DetectionMetadata.ErrorsEncountered, 
				fmt.Sprintf("connectivity: %v", err),
			)
		} else {
			enhancedInfo.ConnectivityInfo = connectivityInfo
		}
	}

	// Generate compatibility information
	compatibilityInfo, err := ehd.generateCompatibilityInfo(detectionCtx, enhancedInfo)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Compatibility analysis failed: %v", err))
		enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
			enhancedInfo.DetectionMetadata.ErrorsEncountered, 
			fmt.Sprintf("compatibility analysis: %v", err),
		)
	} else {
		enhancedInfo.CompatibilityInfo = compatibilityInfo
	}

	// Generate recommended configuration
	recommendedConfig, err := ehd.generateRecommendedConfig(detectionCtx, enhancedInfo)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Config generation failed: %v", err))
		enhancedInfo.DetectionMetadata.ErrorsEncountered = append(
			enhancedInfo.DetectionMetadata.ErrorsEncountered, 
			fmt.Sprintf("config generation: %v", err),
		)
	} else {
		enhancedInfo.RecommendedConfig = recommendedConfig
	}

	// Complete detection metadata
	enhancedInfo.DetectionMetadata.DetectionDuration = time.Since(startTime)
	enhancedInfo.DetectionMetadata.OSVersion = ehd.getOSVersion()
	enhancedInfo.DetectionMetadata.KernelVersion = ehd.getKernelVersion()
	
	// Calculate reliability and completion scores
	totalSections := 8
	completedSections := ehd.calculateCompletedSections(enhancedInfo)
	enhancedInfo.DetectionMetadata.CompletionRate = float64(completedSections) / float64(totalSections) * 100
	enhancedInfo.DetectionMetadata.ReliabilityScore = ehd.calculateReliabilityScore(enhancedInfo)

	ehd.logger.Info(fmt.Sprintf("Enhanced hardware detection completed in %v with %.1f%% completion rate", 
		enhancedInfo.DetectionMetadata.DetectionDuration,
		enhancedInfo.DetectionMetadata.CompletionRate))

	return enhancedInfo, nil
}

// Additional helper methods would continue here...
// This is a comprehensive foundation for the enhanced hardware detection system