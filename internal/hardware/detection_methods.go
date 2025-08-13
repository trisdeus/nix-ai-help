// detection_methods.go - Core detection methods for enhanced hardware detector
package hardware

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// detectSystemProfile performs detailed system profiling
func (ehd *EnhancedHardwareDetector) detectSystemProfile(ctx context.Context) (*SystemProfile, error) {
	profile := &SystemProfile{}

	// Detect CPU details
	cpuProfile, err := ehd.detectCPUProfile(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("CPU profiling failed: %v", err))
	} else {
		profile.CPUDetails = cpuProfile
	}

	// Detect GPU details
	gpuProfiles, err := ehd.detectGPUProfiles(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("GPU profiling failed: %v", err))
	} else {
		profile.GPUDetails = gpuProfiles
	}

	// Detect memory details
	memoryProfile, err := ehd.detectMemoryProfile(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Memory profiling failed: %v", err))
	} else {
		profile.MemoryDetails = memoryProfile
	}

	// Detect storage details
	storageProfiles, err := ehd.detectStorageProfiles(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Storage profiling failed: %v", err))
	} else {
		profile.StorageDetails = storageProfiles
	}

	// Detect network details
	networkProfiles, err := ehd.detectNetworkProfiles(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Network profiling failed: %v", err))
	} else {
		profile.NetworkDetails = networkProfiles
	}

	// Detect audio details
	audioProfile, err := ehd.detectAudioProfile(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Audio profiling failed: %v", err))
	} else {
		profile.AudioDetails = audioProfile
	}

	// Detect display details
	displayProfile, err := ehd.detectDisplayProfile(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Display profiling failed: %v", err))
	} else {
		profile.DisplayDetails = displayProfile
	}

	// Detect input devices
	inputDevices, err := ehd.detectInputDevices(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Input device detection failed: %v", err))
	} else {
		profile.InputDevices = inputDevices
	}

	// Detect USB devices
	usbDevices, err := ehd.detectUSBDevices(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("USB device detection failed: %v", err))
	} else {
		profile.USBDevices = usbDevices
	}

	// Detect Bluetooth devices
	bluetoothDevices, err := ehd.detectBluetoothDevices(ctx)
	if err != nil {
		ehd.logger.Warn(fmt.Sprintf("Bluetooth device detection failed: %v", err))
	} else {
		profile.BluetoothDevices = bluetoothDevices
	}

	return profile, nil
}

// detectCPUProfile performs detailed CPU profiling
func (ehd *EnhancedHardwareDetector) detectCPUProfile(ctx context.Context) (*CPUProfile, error) {
	profile := &CPUProfile{
		Features:         []string{},
		SecurityFeatures: []string{},
		PowerManagement:  []string{},
		Vulnerabilities:  make(map[string]string),
	}

	// Read /proc/cpuinfo
	cpuinfo, err := ehd.readFileWithCache("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/cpuinfo: %v", err)
	}

	lines := strings.Split(cpuinfo, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "vendor_id":
			profile.Vendor = value
		case "model name":
			profile.Model = value
		case "cpu family":
			profile.Family = value
		case "stepping":
			profile.Stepping = value
		case "microcode":
			profile.Microcode = value
		case "cpu cores":
			if cores, err := strconv.Atoi(value); err == nil {
				profile.Cores = cores
			}
		case "siblings":
			if threads, err := strconv.Atoi(value); err == nil {
				profile.Threads = threads
			}
		case "cpu MHz":
			if freq, err := strconv.ParseFloat(value, 64); err == nil {
				profile.BaseFrequency = freq
			}
		case "cache size":
			profile.CacheL3 = value
		case "flags":
			profile.Features = strings.Fields(value)
		}
	}

	// Get architecture
	if arch, err := ehd.runCommandWithCache("uname -m"); err == nil {
		profile.Architecture = strings.TrimSpace(arch)
	}

	// Check virtualization support
	for _, feature := range profile.Features {
		if feature == "vmx" || feature == "svm" {
			profile.VirtualizationSupport = true
			break
		}
	}

	// Check NUMA support
	if _, err := os.Stat("/sys/devices/system/node"); err == nil {
		profile.NUMA = true
	}

	// Get CPU vulnerabilities
	vulnDir := "/sys/devices/system/cpu/vulnerabilities"
	if entries, err := os.ReadDir(vulnDir); err == nil {
		for _, entry := range entries {
			if vulnContent, err := ehd.readFileWithCache(filepath.Join(vulnDir, entry.Name())); err == nil {
				profile.Vulnerabilities[entry.Name()] = strings.TrimSpace(vulnContent)
			}
		}
	}

	// Get CPU frequency information
	if maxFreq, err := ehd.readFileWithCache("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"); err == nil {
		if freq, err := strconv.ParseFloat(strings.TrimSpace(maxFreq), 64); err == nil {
			profile.MaxFrequency = freq / 1000 // Convert from kHz to MHz
		}
	}

	// Detect security features
	securityFeatures := []string{}
	for _, feature := range profile.Features {
		switch feature {
		case "smep", "smap", "ibrs", "ibpb", "stibp", "ssbd":
			securityFeatures = append(securityFeatures, strings.ToUpper(feature))
		}
	}
	profile.SecurityFeatures = securityFeatures

	// Detect power management features
	powerFeatures := []string{}
	for _, feature := range profile.Features {
		switch feature {
		case "acpi", "aperfmperf", "cpb", "hwp", "hwp_notify", "hwp_act_window", "hwp_epp", "hwp_pkg_req":
			powerFeatures = append(powerFeatures, feature)
		}
	}
	profile.PowerManagement = powerFeatures

	return profile, nil
}

// detectGPUProfiles performs detailed GPU profiling
func (ehd *EnhancedHardwareDetector) detectGPUProfiles(ctx context.Context) ([]*GPUProfile, error) {
	var profiles []*GPUProfile

	// Get GPU information from lspci
	lspciOutput, err := ehd.runCommandWithCache("lspci -v | grep -A 20 'VGA\\|3D\\|Display'")
	if err != nil {
		return nil, fmt.Errorf("failed to run lspci: %v", err)
	}

	// Parse lspci output for GPU details
	blocks := strings.Split(lspciOutput, "\n\n")
	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		profile := &GPUProfile{
			DisplayPorts:  []string{},
			Capabilities: make(map[string]string),
		}

		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "VGA") || strings.Contains(line, "3D") || strings.Contains(line, "Display") {
				// Parse GPU name and vendor
				parts := strings.Split(line, ":")
				if len(parts) >= 3 {
					gpuInfo := strings.TrimSpace(parts[2])
					profile.Model = gpuInfo
					
					if strings.Contains(strings.ToLower(gpuInfo), "nvidia") {
						profile.Vendor = "NVIDIA"
						profile.CUDASupport = true
					} else if strings.Contains(strings.ToLower(gpuInfo), "amd") || strings.Contains(strings.ToLower(gpuInfo), "radeon") {
						profile.Vendor = "AMD"
					} else if strings.Contains(strings.ToLower(gpuInfo), "intel") {
						profile.Vendor = "Intel"
					}
				}
			} else if strings.Contains(line, "Kernel driver in use:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					profile.Driver = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, "Subsystem:") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					profile.SubsystemID = strings.TrimSpace(parts[1])
				}
			}
		}

		// Set default capabilities based on vendor
		switch profile.Vendor {
		case "NVIDIA":
			profile.CUDASupport = true
			profile.OpenCLSupport = true
			profile.VulkanSupport = true
			profile.DirectXSupport = "12"
			profile.OpenGLSupport = "4.6"
		case "AMD":
			profile.OpenCLSupport = true
			profile.VulkanSupport = true
			profile.DirectXSupport = "12"
			profile.OpenGLSupport = "4.6"
		case "Intel":
			profile.OpenCLSupport = true
			profile.VulkanSupport = true
			profile.DirectXSupport = "12"
			profile.OpenGLSupport = "4.6"
		}

		if profile.Model != "" {
			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// detectMemoryProfile performs detailed memory profiling
func (ehd *EnhancedHardwareDetector) detectMemoryProfile(ctx context.Context) (*MemoryProfile, error) {
	profile := &MemoryProfile{
		Slots:          []*MemorySlot{},
		Latency:        make(map[string]int),
		VoltageProfile: make(map[string]float64),
	}

	// Get memory info from /proc/meminfo
	meminfo, err := ehd.readFileWithCache("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %v", err)
	}

	lines := strings.Split(meminfo, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSuffix(parts[0], ":")
		value := parts[1]

		switch key {
		case "MemTotal":
			if kb, err := strconv.Atoi(value); err == nil {
				profile.TotalCapacity = fmt.Sprintf("%.1f GB", float64(kb)/1024/1024)
			}
		case "MemAvailable":
			if kb, err := strconv.Atoi(value); err == nil {
				profile.AvailableCapacity = fmt.Sprintf("%.1f GB", float64(kb)/1024/1024)
			}
		}
	}

	// Try to get detailed memory information from dmidecode
	if dmidecodeOutput, err := ehd.runCommandWithCache("dmidecode -t memory 2>/dev/null"); err == nil {
		ehd.parseMemoryFromDmidecode(profile, dmidecodeOutput)
	}

	// Calculate used capacity
	if profile.TotalCapacity != "" && profile.AvailableCapacity != "" {
		totalGB := ehd.parseCapacityToGB(profile.TotalCapacity)
		availableGB := ehd.parseCapacityToGB(profile.AvailableCapacity)
		usedGB := totalGB - availableGB
		profile.UsedCapacity = fmt.Sprintf("%.1f GB", usedGB)
	}

	return profile, nil
}

// detectStorageProfiles performs detailed storage profiling
func (ehd *EnhancedHardwareDetector) detectStorageProfiles(ctx context.Context) ([]*StorageProfile, error) {
	var profiles []*StorageProfile

	// Get block device information
	lsblkOutput, err := ehd.runCommandWithCache("lsblk -J -o NAME,SIZE,TYPE,VENDOR,MODEL,SERIAL,FSTYPE,MOUNTPOINT,UUID")
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk: %v", err)
	}

	// Parse lsblk JSON output (simplified parsing for now)
	lines := strings.Split(lsblkOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "disk") {
			profile := &StorageProfile{
				Partitions: []*PartitionInfo{},
				SMARTData:  make(map[string]string),
			}

			// Extract basic information from lsblk
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				profile.DeviceName = fields[0]
				profile.Capacity = fields[1]
			}

			// Try to get more detailed information
			devicePath := "/dev/" + profile.DeviceName
			profile.DevicePath = devicePath

			// Determine storage type
			if strings.HasPrefix(profile.DeviceName, "nvme") {
				profile.Type = "NVMe SSD"
				profile.Interface = "NVMe"
			} else if strings.HasPrefix(profile.DeviceName, "sd") {
				// Could be SSD or HDD, try to determine
				if rotational, err := ehd.readFileWithCache(fmt.Sprintf("/sys/block/%s/queue/rotational", profile.DeviceName)); err == nil {
					if strings.TrimSpace(rotational) == "0" {
						profile.Type = "SSD"
						profile.Interface = "SATA"
					} else {
						profile.Type = "HDD"
						profile.Interface = "SATA"
					}
				}
			}

			// Try to get SMART data
			if smartOutput, err := ehd.runCommandWithCache(fmt.Sprintf("smartctl -a %s 2>/dev/null", devicePath)); err == nil {
				ehd.parseSMARTData(profile, smartOutput)
			}

			// Get file system information
			if mountOutput, err := ehd.runCommandWithCache("mount"); err == nil {
				ehd.parseFilesystemInfo(profile, mountOutput)
			}

			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// detectNetworkProfiles performs detailed network profiling
func (ehd *EnhancedHardwareDetector) detectNetworkProfiles(ctx context.Context) ([]*NetworkProfile, error) {
	var profiles []*NetworkProfile

	// Get network interface information
	interfaces, err := ehd.runCommandWithCache("ip link show")
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}

	lines := strings.Split(interfaces, "\n")
	for _, line := range lines {
		if strings.Contains(line, ": ") && !strings.Contains(line, "    ") {
			profile := &NetworkProfile{
				IPAddresses:   []string{},
				IPv6Addresses: []string{},
				DNSServers:    []string{},
				Capabilities:  []string{},
				Statistics:    make(map[string]uint64),
			}

			// Parse interface name
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				profile.InterfaceName = strings.TrimSpace(parts[1])
			}

			// Determine interface type
			if strings.HasPrefix(profile.InterfaceName, "wl") {
				profile.Type = "WiFi"
				// Get wireless information
				if wirelessInfo, err := ehd.detectWirelessInfo(profile.InterfaceName); err == nil {
					profile.WirelessInfo = wirelessInfo
				}
			} else if strings.HasPrefix(profile.InterfaceName, "en") || strings.HasPrefix(profile.InterfaceName, "eth") {
				profile.Type = "Ethernet"
			} else if profile.InterfaceName == "lo" {
				profile.Type = "Loopback"
			} else {
				profile.Type = "Other"
			}

			// Get MAC address
			macRegex := regexp.MustCompile(`([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})`)
			if matches := macRegex.FindStringSubmatch(line); len(matches) > 1 {
				profile.MACAddress = matches[1]
			}

			// Get state
			if strings.Contains(line, "UP") {
				profile.State = "UP"
			} else {
				profile.State = "DOWN"
			}

			// Get MTU
			mtuRegex := regexp.MustCompile(`mtu (\d+)`)
			if matches := mtuRegex.FindStringSubmatch(line); len(matches) > 1 {
				if mtu, err := strconv.Atoi(matches[1]); err == nil {
					profile.MTU = mtu
				}
			}

			// Get IP addresses
			if ipOutput, err := ehd.runCommandWithCache(fmt.Sprintf("ip addr show %s", profile.InterfaceName)); err == nil {
				ehd.parseIPAddresses(profile, ipOutput)
			}

			// Get interface statistics
			if statOutput, err := ehd.runCommandWithCache(fmt.Sprintf("cat /sys/class/net/%s/statistics/* 2>/dev/null", profile.InterfaceName)); err == nil {
				ehd.parseNetworkStatistics(profile, statOutput)
			}

			profiles = append(profiles, profile)
		}
	}

	return profiles, nil
}

// detectAudioProfile performs detailed audio profiling
func (ehd *EnhancedHardwareDetector) detectAudioProfile(ctx context.Context) (*AudioProfile, error) {
	profile := &AudioProfile{
		Cards:   []*AudioCard{},
		Devices: []*AudioDevice{},
	}

	// Detect sound server
	if _, err := ehd.runCommandWithCache("pgrep pipewire"); err == nil {
		profile.SoundServer = "PipeWire"
	} else if _, err := ehd.runCommandWithCache("pgrep pulseaudio"); err == nil {
		profile.SoundServer = "PulseAudio"
	} else {
		profile.SoundServer = "ALSA"
	}

	// Get ALSA card information
	if cardsOutput, err := ehd.runCommandWithCache("cat /proc/asound/cards 2>/dev/null"); err == nil {
		ehd.parseALSACards(profile, cardsOutput)
	}

	// Get PulseAudio/PipeWire device information
	if profile.SoundServer == "PulseAudio" {
		if paOutput, err := ehd.runCommandWithCache("pactl list short sinks 2>/dev/null"); err == nil {
			ehd.parsePulseAudioDevices(profile, paOutput, "sink")
		}
		if paOutput, err := ehd.runCommandWithCache("pactl list short sources 2>/dev/null"); err == nil {
			ehd.parsePulseAudioDevices(profile, paOutput, "source")
		}
	}

	return profile, nil
}

// detectDisplayProfile performs detailed display profiling
func (ehd *EnhancedHardwareDetector) detectDisplayProfile(ctx context.Context) (*DisplayProfile, error) {
	profile := &DisplayProfile{
		Displays:  []*Display{},
		DPI:       make(map[string]int),
		ThemeInfo: make(map[string]string),
	}

	// Detect display server
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		profile.Server = "Wayland"
	} else if os.Getenv("DISPLAY") != "" {
		profile.Server = "X11"
	} else {
		profile.Server = "Console"
	}

	// Get X11 information
	if profile.Server == "X11" {
		if xrandrOutput, err := ehd.runCommandWithCache("xrandr 2>/dev/null"); err == nil {
			ehd.parseXrandrOutput(profile, xrandrOutput)
		}
	}

	// Get desktop environment information
	if de := os.Getenv("XDG_CURRENT_DESKTOP"); de != "" {
		profile.DesktopEnvironment = de
	}

	if wm := os.Getenv("XDG_SESSION_DESKTOP"); wm != "" {
		profile.WindowManager = wm
	}

	return profile, nil
}

// Helper methods for parsing various outputs
func (ehd *EnhancedHardwareDetector) parseMemoryFromDmidecode(profile *MemoryProfile, output string) {
	// Parse dmidecode output for detailed memory information
	sections := strings.Split(output, "\n\n")
	
	for _, section := range sections {
		if !strings.Contains(section, "Memory Device") {
			continue
		}

		slot := &MemorySlot{}
		lines := strings.Split(section, "\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if !strings.Contains(line, ":") {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "Size":
				if value != "No Module Installed" && value != "" {
					slot.Capacity = value
					slot.Populated = true
				}
			case "Type":
				slot.Type = value
				if profile.MemoryType == "" {
					profile.MemoryType = value
				}
			case "Speed":
				if speed, err := strconv.Atoi(strings.Fields(value)[0]); err == nil {
					slot.Speed = speed
					if profile.Speed == 0 {
						profile.Speed = speed
					}
				}
			case "Manufacturer":
				slot.Manufacturer = value
				if profile.Manufacturer == "" {
					profile.Manufacturer = value
				}
			case "Part Number":
				slot.PartNumber = value
			case "Serial Number":
				slot.SerialNumber = value
			case "Bank Locator":
				slot.BankLocator = value
			}
		}

		if slot.Populated {
			profile.Slots = append(profile.Slots, slot)
		}
	}
}

// Additional helper methods would continue here...

// runCommandWithCache runs a command and caches the result
func (ehd *EnhancedHardwareDetector) runCommandWithCache(command string) (string, error) {
	if ehd.cache != nil && ehd.cache.enabled {
		if entry, exists := ehd.cache.data[command]; exists {
			if time.Since(entry.Timestamp) < entry.TTL {
				return entry.Data.(string), nil
			}
		}
	}

	output, err := runCommand(command)
	if err != nil {
		return "", err
	}

	if ehd.cache != nil && ehd.cache.enabled {
		ehd.cache.data[command] = CacheEntry{
			Data:      output,
			Timestamp: time.Now(),
			TTL:       ehd.cache.ttl,
		}
	}

	return output, nil
}

// readFileWithCache reads a file and caches the result
func (ehd *EnhancedHardwareDetector) readFileWithCache(filename string) (string, error) {
	if ehd.cache != nil && ehd.cache.enabled {
		if entry, exists := ehd.cache.data[filename]; exists {
			if time.Since(entry.Timestamp) < entry.TTL {
				return entry.Data.(string), nil
			}
		}
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	result := string(content)

	if ehd.cache != nil && ehd.cache.enabled {
		ehd.cache.data[filename] = CacheEntry{
			Data:      result,
			Timestamp: time.Now(),
			TTL:       ehd.cache.ttl,
		}
	}

	return result, nil
}

// parseCapacityToGB parses capacity string to GB
func (ehd *EnhancedHardwareDetector) parseCapacityToGB(capacity string) float64 {
	// Simple parser for capacity strings like "16.0 GB", "1024 MB", etc.
	parts := strings.Fields(capacity)
	if len(parts) < 2 {
		return 0
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}

	unit := strings.ToLower(parts[1])
	switch unit {
	case "gb":
		return value
	case "mb":
		return value / 1024
	case "tb":
		return value * 1024
	case "kb":
		return value / (1024 * 1024)
	default:
		return value
	}
}