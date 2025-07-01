package system

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// SystemInfo represents current system information
type SystemInfo struct {
	Uptime    string `json:"uptime"`
	CPUUsage  string `json:"cpu_usage"`
	Memory    string `json:"memory"`
	Disk      string `json:"disk"`
	LoadAvg   string `json:"load_avg"`
	Processes int    `json:"processes"`
}

// GetSystemInfo returns current system information
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}
	
	// Get uptime
	uptime, err := getUptime()
	if err != nil {
		info.Uptime = "unknown"
	} else {
		info.Uptime = formatUptime(uptime)
	}
	
	// Get memory info
	memInfo, err := getMemoryInfo()
	if err != nil {
		info.Memory = "unknown"
	} else {
		info.Memory = fmt.Sprintf("%.1fGB", float64(memInfo.Used)/1024/1024/1024)
	}
	
	// Get disk info
	diskInfo, err := getDiskInfo("/")
	if err != nil {
		info.Disk = "unknown"
	} else {
		info.Disk = fmt.Sprintf("%.0f%%", diskInfo.UsedPercent)
	}
	
	// Get load average
	loadAvg, err := getLoadAverage()
	if err != nil {
		info.LoadAvg = "unknown"
	} else {
		info.LoadAvg = fmt.Sprintf("%.2f", loadAvg)
	}
	
	// CPU usage is more complex to calculate accurately, so we'll use load average as approximation
	if loadAvg > 0 {
		cpuPercent := loadAvg / float64(runtime.NumCPU()) * 100
		if cpuPercent > 100 {
			cpuPercent = 100
		}
		info.CPUUsage = fmt.Sprintf("%.0f%%", cpuPercent)
	} else {
		info.CPUUsage = "0%"
	}
	
	// Get process count
	info.Processes = getProcessCount()
	
	return info, nil
}

// MemoryInfo represents memory statistics
type MemoryInfo struct {
	Total uint64
	Used  uint64
	Free  uint64
}

// DiskInfo represents disk statistics
type DiskInfo struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

func getUptime() (time.Duration, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}
	
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}
	
	return time.Duration(seconds) * time.Second, nil
}

func formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func getMemoryInfo() (*MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	memInfo := &MemoryInfo{}
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		
		// Convert from kB to bytes
		value *= 1024
		
		switch key {
		case "MemTotal":
			memInfo.Total = value
		case "MemAvailable":
			memInfo.Free = value
		}
	}
	
	memInfo.Used = memInfo.Total - memInfo.Free
	return memInfo, nil
}

func getDiskInfo(path string) (*DiskInfo, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	
	usedPercent := float64(used) / float64(total) * 100
	
	return &DiskInfo{
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: usedPercent,
	}, nil
}

func getLoadAverage() (float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid loadavg format")
	}
	
	return strconv.ParseFloat(fields[0], 64)
}

func getProcessCount() int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if directory name is numeric (PID)
			if _, err := strconv.Atoi(entry.Name()); err == nil {
				count++
			}
		}
	}
	
	return count
}
