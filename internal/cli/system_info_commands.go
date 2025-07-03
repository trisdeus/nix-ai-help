package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// CreateSystemInfoCommand creates the system-info command and its subcommands
func CreateSystemInfoCommand() *cobra.Command {
	systemInfoCmd := &cobra.Command{
		Use:   "system-info",
		Short: "System information and health monitoring",
		Long: `System information and health monitoring commands.

Get comprehensive system information, perform health checks, and monitor
system performance. This includes CPU, memory, disk, and service monitoring.

Examples:
  nixai system-info status         # System health overview
  nixai system-info health         # Detailed health check
  nixai system-info cpu            # CPU information
  nixai system-info memory         # Memory usage details
  nixai system-info disk           # Disk usage information
  nixai system-info processes      # Top processes
  nixai system-info monitor        # Interactive monitoring`,
	}

	// Add subcommands
	systemInfoCmd.AddCommand(createSystemInfoStatusCmd())
	systemInfoCmd.AddCommand(createSystemInfoHealthCmd())
	systemInfoCmd.AddCommand(createSystemInfoCPUCmd())
	systemInfoCmd.AddCommand(createSystemInfoMemoryCmd())
	systemInfoCmd.AddCommand(createSystemInfoDiskCmd())
	systemInfoCmd.AddCommand(createSystemInfoProcessesCmd())
	systemInfoCmd.AddCommand(createSystemInfoMonitorCmd())
	systemInfoCmd.AddCommand(createSystemInfoAllCmd())

	return systemInfoCmd
}

// createSystemInfoStatusCmd creates the status subcommand
func createSystemInfoStatusCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system status overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputSystemInfoJSON()
			}
			return outputSystemInfo()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createSystemInfoHealthCmd creates the health subcommand
func createSystemInfoHealthCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Perform comprehensive system health check",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputHealthCheckJSON()
			}
			return outputHealthCheck()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createSystemInfoCPUCmd creates the cpu subcommand
func createSystemInfoCPUCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "cpu",
		Short: "Show CPU information and usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputCPUInfoJSON()
			}
			return outputCPUInfo()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createSystemInfoMemoryCmd creates the memory subcommand
func createSystemInfoMemoryCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Show memory usage information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputMemoryInfoJSON()
			}
			return outputMemoryInfo()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createSystemInfoDiskCmd creates the disk subcommand
func createSystemInfoDiskCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Show disk usage information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputDiskInfoJSON()
			}
			return outputDiskInfo()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createSystemInfoProcessesCmd creates the processes subcommand
func createSystemInfoProcessesCmd() *cobra.Command {
	var jsonOutput bool
	var limit int

	cmd := &cobra.Command{
		Use:   "processes",
		Short: "Show top processes by CPU and memory usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputProcessesJSON(limit)
			}
			return outputProcesses(limit)
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Number of processes to show")
	return cmd
}

// createSystemInfoMonitorCmd creates the monitor subcommand
func createSystemInfoMonitorCmd() *cobra.Command {
	var interval int

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start interactive system monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startMonitoring(interval)
		},
	}

	cmd.Flags().IntVarP(&interval, "interval", "i", 5, "Update interval in seconds")
	return cmd
}

// createSystemInfoAllCmd creates the all subcommand
func createSystemInfoAllCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "all",
		Short: "Show comprehensive system information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputAllInfoJSON()
			}
			return outputAllInfo()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// Implementation functions

func outputSystemInfo() error {
	fmt.Println(utils.FormatHeader("🖥️ System Information"))
	fmt.Println()

	hostname, _ := os.Hostname()
	
	fmt.Println(utils.FormatKeyValue("Hostname", hostname))
	fmt.Println(utils.FormatKeyValue("OS", runtime.GOOS))
	fmt.Println(utils.FormatKeyValue("Architecture", runtime.GOARCH))
	fmt.Println(utils.FormatKeyValue("Go Version", runtime.Version()))
	fmt.Println(utils.FormatKeyValue("CPU Cores", fmt.Sprintf("%d", runtime.NumCPU())))

	// Get uptime if available
	if uptimeBytes, err := os.ReadFile("/proc/uptime"); err == nil {
		uptimeStr := strings.Fields(string(uptimeBytes))[0]
		if uptimeFloat, err := strconv.ParseFloat(uptimeStr, 64); err == nil {
			uptime := time.Duration(uptimeFloat * float64(time.Second))
			fmt.Println(utils.FormatKeyValue("Uptime", uptime.String()))
		}
	}

	// Get kernel version if available
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		kernel := strings.TrimSpace(string(out))
		fmt.Println(utils.FormatKeyValue("Kernel", kernel))
	}

	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))
	
	return nil
}

func outputSystemInfoJSON() error {
	info := getSystemInfoData()
	output, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputHealthCheck() error {
	fmt.Println(utils.FormatHeader("🏥 System Health Check"))
	fmt.Println()

	// Memory check
	memUsage := getMemoryUsagePercent()
	if memUsage > 90 {
		fmt.Println(utils.FormatError(fmt.Sprintf("❌ CRITICAL: Memory usage is %.1f%% (>90%%)", memUsage)))
	} else if memUsage > 80 {
		fmt.Println(utils.FormatWarning(fmt.Sprintf("⚠️  WARNING: Memory usage is %.1f%% (>80%%)", memUsage)))
	} else {
		fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ GOOD: Memory usage is %.1f%%", memUsage)))
	}

	// Disk check
	diskUsage := getDiskUsagePercent()
	if diskUsage > 95 {
		fmt.Println(utils.FormatError(fmt.Sprintf("❌ CRITICAL: Root disk usage is %d%% (>95%%)", diskUsage)))
	} else if diskUsage > 85 {
		fmt.Println(utils.FormatWarning(fmt.Sprintf("⚠️  WARNING: Root disk usage is %d%% (>85%%)", diskUsage)))
	} else {
		fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ GOOD: Root disk usage is %d%%", diskUsage)))
	}

	// Load average check
	loadAvg := getLoadAverage()
	if loadAvg != "" {
		cpuCores := runtime.NumCPU()
		if load, err := strconv.ParseFloat(loadAvg, 64); err == nil {
			loadPerCore := load / float64(cpuCores)
			if loadPerCore > 2.0 {
				fmt.Println(utils.FormatError(fmt.Sprintf("❌ CRITICAL: Load average is %s (%.2f per core, >2.0)", loadAvg, loadPerCore)))
			} else if loadPerCore > 1.0 {
				fmt.Println(utils.FormatWarning(fmt.Sprintf("⚠️  WARNING: Load average is %s (%.2f per core, >1.0)", loadAvg, loadPerCore)))
			} else {
				fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ GOOD: Load average is %s (%.2f per core)", loadAvg, loadPerCore)))
			}
		}
	}

	// Service check
	fmt.Println()
	fmt.Println(utils.FormatHeader("🔧 Critical Services Status"))
	
	services := []string{"sshd", "systemd-resolved", "NetworkManager"}
	for _, service := range services {
		if isServiceActive(service) {
			fmt.Println(utils.FormatSuccess(fmt.Sprintf("✅ %s: running", service)))
		} else {
			fmt.Println(utils.FormatWarning(fmt.Sprintf("⚠️  %s: not running or not available", service)))
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Overall Status", time.Now().Format(time.RFC3339)))
	
	return nil
}

func outputHealthCheckJSON() error {
	health := getHealthCheckData()
	output, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputCPUInfo() error {
	fmt.Println(utils.FormatHeader("💻 CPU Information"))
	fmt.Println()

	fmt.Println(utils.FormatKeyValue("CPU Cores", fmt.Sprintf("%d", runtime.NumCPU())))
	fmt.Println(utils.FormatKeyValue("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine())))

	// Get CPU model if available
	if cpuModel := getCPUModel(); cpuModel != "" {
		fmt.Println(utils.FormatKeyValue("CPU Model", cpuModel))
	}

	// Get CPU frequency if available
	if cpuFreq := getCPUFrequency(); cpuFreq != "" {
		fmt.Println(utils.FormatKeyValue("CPU Frequency", cpuFreq))
	}

	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))
	
	return nil
}

func outputCPUInfoJSON() error {
	cpu := getCPUInfoData()
	output, err := json.MarshalIndent(cpu, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputMemoryInfo() error {
	fmt.Println(utils.FormatHeader("🧠 Memory Information"))
	fmt.Println()

	// Go runtime memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println(utils.FormatHeader("Go Runtime Memory"))
	fmt.Println(utils.FormatKeyValue("Allocated", fmt.Sprintf("%d MB", bToMb(m.Alloc))))
	fmt.Println(utils.FormatKeyValue("Total Allocated", fmt.Sprintf("%d MB", bToMb(m.TotalAlloc))))
	fmt.Println(utils.FormatKeyValue("System", fmt.Sprintf("%d MB", bToMb(m.Sys))))
	fmt.Println(utils.FormatKeyValue("GC Runs", fmt.Sprintf("%d", m.NumGC)))

	// System memory if available
	if memInfo := getSystemMemoryInfo(); len(memInfo) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatHeader("System Memory"))
		for key, value := range memInfo {
			fmt.Println(utils.FormatKeyValue(key, fmt.Sprintf("%d MB", value/(1024*1024))))
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))
	
	return nil
}

func outputMemoryInfoJSON() error {
	memory := getMemoryInfoData()
	output, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputDiskInfo() error {
	fmt.Println(utils.FormatHeader("💾 Disk Information"))
	fmt.Println()

	// Get disk usage with df command
	if out, err := exec.Command("df", "-h").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fmt.Println(utils.FormatHeader("Disk Usage"))
			fmt.Println(strings.Join(lines[:2], "\n")) // Header + first filesystem
			
			// Show only non-temporary filesystems
			for i := 2; i < len(lines); i++ {
				line := strings.TrimSpace(lines[i])
				if line != "" && !strings.Contains(line, "tmpfs") && !strings.Contains(line, "devtmpfs") {
					fmt.Println(line)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))
	
	return nil
}

func outputDiskInfoJSON() error {
	disk := getDiskInfoData()
	output, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputProcesses(limit int) error {
	fmt.Println(utils.FormatHeader("🔄 Top Processes"))
	fmt.Println()

	// Top processes by CPU
	fmt.Println(utils.FormatHeader("Top Processes by CPU Usage"))
	if out, err := exec.Command("ps", "aux", "--sort=-pcpu").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fmt.Println(lines[0]) // Header
			for i := 1; i < len(lines) && i <= limit; i++ {
				if strings.TrimSpace(lines[i]) != "" {
					fmt.Println(lines[i])
				}
			}
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatHeader("Top Processes by Memory Usage"))
	if out, err := exec.Command("ps", "aux", "--sort=-rss").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fmt.Println(lines[0]) // Header
			for i := 1; i < len(lines) && i <= limit; i++ {
				if strings.TrimSpace(lines[i]) != "" {
					fmt.Println(lines[i])
				}
			}
		}
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))
	
	return nil
}

func outputProcessesJSON(limit int) error {
	processes := getProcessesData(limit)
	output, err := json.MarshalIndent(processes, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func startMonitoring(interval int) error {
	fmt.Println(utils.FormatHeader("📊 System Monitor"))
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	for {
		// Clear screen (basic approach)
		fmt.Print("\033[2J\033[H")
		
		fmt.Printf("NixAI System Monitor - %s\n", time.Now().Format("15:04:05"))
		fmt.Println("Press Ctrl+C to exit")
		fmt.Println()

		// Quick health summary
		memUsage := getMemoryUsagePercent()
		diskUsage := getDiskUsagePercent()
		loadAvg := getLoadAverage()

		fmt.Println("📊 Quick Health Summary:")
		fmt.Printf("   Memory: %.1f%%\n", memUsage)
		fmt.Printf("   Disk:   %d%%\n", diskUsage)
		fmt.Printf("   Load:   %s\n", loadAvg)
		fmt.Println()

		// Top processes
		fmt.Println("🔄 Top Processes (CPU):")
		if out, err := exec.Command("ps", "aux", "--sort=-pcpu").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for i := 1; i < len(lines) && i <= 6; i++ {
				if strings.TrimSpace(lines[i]) != "" {
					fields := strings.Fields(lines[i])
					if len(fields) >= 11 {
						fmt.Printf("   %-12s %5s%% %s\n", fields[10], fields[2], fields[0])
					}
				}
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func outputAllInfo() error {
	fmt.Println(utils.FormatHeader("📋 Comprehensive System Information"))
	fmt.Println()

	outputSystemInfo()
	fmt.Println()
	outputHealthCheck()
	fmt.Println()
	outputCPUInfo()
	fmt.Println()
	outputMemoryInfo()
	fmt.Println()
	outputDiskInfo()

	return nil
}

func outputAllInfoJSON() error {
	all := map[string]interface{}{
		"system":    getSystemInfoData(),
		"health":    getHealthCheckData(),
		"cpu":       getCPUInfoData(),
		"memory":    getMemoryInfoData(),
		"disk":      getDiskInfoData(),
		"processes": getProcessesData(10),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// Helper functions

func getSystemInfoData() map[string]interface{} {
	hostname, _ := os.Hostname()
	info := map[string]interface{}{
		"hostname":    hostname,
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
		"go_version":  runtime.Version(),
		"num_cpu":     runtime.NumCPU(),
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Add uptime if available
	if uptimeBytes, err := os.ReadFile("/proc/uptime"); err == nil {
		uptimeStr := strings.Fields(string(uptimeBytes))[0]
		if uptimeFloat, err := strconv.ParseFloat(uptimeStr, 64); err == nil {
			uptime := time.Duration(uptimeFloat * float64(time.Second))
			info["uptime"] = uptime.String()
		}
	}

	// Add kernel if available
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		info["kernel"] = strings.TrimSpace(string(out))
	}

	return info
}

func getHealthCheckData() map[string]interface{} {
	health := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"status":    "healthy",
		"checks":    []map[string]interface{}{},
	}

	checks := []map[string]interface{}{}

	// Memory check
	memUsage := getMemoryUsagePercent()
	memCheck := map[string]interface{}{
		"name":   "memory_usage",
		"status": "ok",
		"value":  memUsage,
		"unit":   "%",
	}
	if memUsage > 90 {
		memCheck["status"] = "critical"
		memCheck["message"] = "Critical memory usage"
	} else if memUsage > 80 {
		memCheck["status"] = "warning"
		memCheck["message"] = "High memory usage"
	}
	checks = append(checks, memCheck)

	// Disk check
	diskUsage := getDiskUsagePercent()
	diskCheck := map[string]interface{}{
		"name":   "disk_usage",
		"status": "ok",
		"value":  diskUsage,
		"unit":   "%",
	}
	if diskUsage > 95 {
		diskCheck["status"] = "critical"
		diskCheck["message"] = "Critical disk usage"
	} else if diskUsage > 85 {
		diskCheck["status"] = "warning"
		diskCheck["message"] = "High disk usage"
	}
	checks = append(checks, diskCheck)

	health["checks"] = checks
	return health
}

func getCPUInfoData() map[string]interface{} {
	cpu := map[string]interface{}{
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	if model := getCPUModel(); model != "" {
		cpu["model"] = model
	}

	if freq := getCPUFrequency(); freq != "" {
		cpu["frequency"] = freq
	}

	return cpu
}

func getMemoryInfoData() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memory := map[string]interface{}{
		"go_runtime": map[string]interface{}{
			"alloc_bytes":     m.Alloc,
			"alloc_mb":        bToMb(m.Alloc),
			"total_alloc_mb":  bToMb(m.TotalAlloc),
			"sys_mb":          bToMb(m.Sys),
			"num_gc":          m.NumGC,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if systemMem := getSystemMemoryInfo(); len(systemMem) > 0 {
		memory["system"] = systemMem
	}

	return memory
}

func getDiskInfoData() map[string]interface{} {
	disk := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Get disk usage with df command
	if out, err := exec.Command("df", "-h", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 4 {
				disk["root_filesystem"] = map[string]interface{}{
					"size":      fields[1],
					"used":      fields[2],
					"available": fields[3],
					"use_pct":   fields[4],
				}
			}
		}
	}

	return disk
}

func getProcessesData(limit int) map[string]interface{} {
	processes := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Get top processes by CPU
	if out, err := exec.Command("ps", "aux", "--sort=-pcpu").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			topProcs := []string{}
			for i := 1; i < len(lines) && i <= limit; i++ {
				if strings.TrimSpace(lines[i]) != "" {
					topProcs = append(topProcs, lines[i])
				}
			}
			processes["top_cpu"] = topProcs
		}
	}

	return processes
}

// Utility functions

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func getMemoryUsagePercent() float64 {
	if meminfo, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, available float64
		lines := strings.Split(string(meminfo), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
						total = val
					}
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if val, err := strconv.ParseFloat(fields[1], 64); err == nil {
						available = val
					}
				}
			}
		}
		if total > 0 && available >= 0 {
			return ((total - available) / total) * 100
		}
	}
	return 0
}

func getDiskUsagePercent() int {
	if out, err := exec.Command("df", "/").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 5 {
				usePct := strings.TrimSuffix(fields[4], "%")
				if pct, err := strconv.Atoi(usePct); err == nil {
					return pct
				}
			}
		}
	}
	return 0
}

func getLoadAverage() string {
	if out, err := exec.Command("uptime").Output(); err == nil {
		uptime := string(out)
		if idx := strings.Index(uptime, "load average:"); idx != -1 {
			loadPart := uptime[idx+len("load average:"):]
			loads := strings.Split(strings.TrimSpace(loadPart), ",")
			if len(loads) > 0 {
				return strings.TrimSpace(loads[0])
			}
		}
	}
	return ""
}

func getCPUModel() string {
	if cpuinfoBytes, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		lines := strings.Split(string(cpuinfoBytes), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return ""
}

func getCPUFrequency() string {
	if out, err := exec.Command("lscpu").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "CPU MHz") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					return strings.TrimSpace(parts[1]) + "MHz"
				}
			}
		}
	}
	return ""
}

func getSystemMemoryInfo() map[string]int64 {
	memInfo := make(map[string]int64)
	if meminfo, err := os.ReadFile("/proc/meminfo"); err == nil {
		lines := strings.Split(string(meminfo), "\n")
		for _, line := range lines {
			if strings.Contains(line, "MemTotal:") || strings.Contains(line, "MemAvailable:") || strings.Contains(line, "MemFree:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if val, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						key := strings.TrimSuffix(fields[0], ":")
						memInfo[key] = val * 1024 // Convert kB to bytes
					}
				}
			}
		}
	}
	return memInfo
}

func isServiceActive(service string) bool {
	if err := exec.Command("systemctl", "is-active", "--quiet", service).Run(); err == nil {
		return true
	}
	return false
}