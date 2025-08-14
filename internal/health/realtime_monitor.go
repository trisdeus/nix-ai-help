// Package health provides real-time system state monitoring
package health

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nix-ai-help/pkg/logger"
)

// RealTimeMonitor provides streaming real-time system monitoring
type RealTimeMonitor struct {
	mu                sync.RWMutex
	logger            *logger.Logger
	systemMonitor     *SystemMonitor
	subscribers       map[string]chan *SystemSnapshot
	running           bool
	ctx               context.Context
	cancel            context.CancelFunc
	snapshotInterval  time.Duration
	alertThresholds   map[string]AlertThreshold
	activeAlerts      map[string]*SystemAlert
	alertSubscribers  map[string]chan *SystemAlert
	metricsHistory    *MetricsRingBuffer
	anomalyDetector   *RealtimeAnomalyDetector
}

// SystemSnapshot represents a point-in-time system state
type SystemSnapshot struct {
	Timestamp         time.Time                `json:"timestamp"`
	Metrics           map[string]float64       `json:"metrics"`
	ComponentHealth   map[string]HealthStatus  `json:"component_health"`
	Alerts            []SystemAlert            `json:"alerts"`
	TrendIndicators   map[string]TrendMetric   `json:"trend_indicators"`
	SystemLoad        SystemLoadInfo           `json:"system_load"`
	ProcessInfo       ProcessSnapshot          `json:"process_info"`
	NetworkActivity   NetworkSnapshot          `json:"network_activity"`
	ResourcePressure  ResourcePressureInfo     `json:"resource_pressure"`
	SecurityEvents    []SecurityEvent          `json:"security_events"`
	SequenceNumber    int64                    `json:"sequence_number"`
}

// SystemAlert represents an active system alert
type SystemAlert struct {
	ID              string                 `json:"id"`
	Type            AlertType              `json:"type"`
	Severity        AlertSeverity          `json:"severity"`
	Component       string                 `json:"component"`
	Metric          string                 `json:"metric"`
	CurrentValue    float64                `json:"current_value"`
	ThresholdValue  float64                `json:"threshold_value"`
	Message         string                 `json:"message"`
	FirstDetected   time.Time              `json:"first_detected"`
	LastUpdated     time.Time              `json:"last_updated"`
	Duration        time.Duration          `json:"duration"`
	TriggerCount    int                    `json:"trigger_count"`
	Acknowledged    bool                   `json:"acknowledged"`
	AcknowledgedBy  string                 `json:"acknowledged_by,omitempty"`
	AcknowledgedAt  *time.Time             `json:"acknowledged_at,omitempty"`
	ResolutionHints []string               `json:"resolution_hints"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// AlertThreshold defines when to trigger alerts
type AlertThreshold struct {
	Warning     float64       `json:"warning"`
	Critical    float64       `json:"critical"`
	Emergency   float64       `json:"emergency"`
	CheckPeriod time.Duration `json:"check_period"`
	Hysteresis  float64       `json:"hysteresis"`  // Prevents flapping
	MinDuration time.Duration `json:"min_duration"` // Alert only after sustained breach
}

// TrendMetric represents trending information for a metric
type TrendMetric struct {
	Current        float64   `json:"current"`
	Previous       float64   `json:"previous"`
	ChangeRate     float64   `json:"change_rate"`     // Per minute
	Direction      string    `json:"direction"`       // "increasing", "decreasing", "stable"
	Volatility     float64   `json:"volatility"`      // Standard deviation
	Prediction1m   float64   `json:"prediction_1m"`   // 1 minute ahead
	Prediction5m   float64   `json:"prediction_5m"`   // 5 minutes ahead
	Confidence     float64   `json:"confidence"`      // Prediction confidence
	LastCalculated time.Time `json:"last_calculated"`
}

// SystemLoadInfo provides detailed system load information
type SystemLoadInfo struct {
	LoadAvg1        float64   `json:"load_avg_1"`
	LoadAvg5        float64   `json:"load_avg_5"`
	LoadAvg15       float64   `json:"load_avg_15"`
	CPUCores        int       `json:"cpu_cores"`
	CPUUtilization  float64   `json:"cpu_utilization"`
	CPUPressure     float64   `json:"cpu_pressure"`      // 0.0-1.0
	MemoryPressure  float64   `json:"memory_pressure"`   // 0.0-1.0
	IOPressure      float64   `json:"io_pressure"`       // 0.0-1.0
	ContextSwitches int64     `json:"context_switches"`
	Interrupts      int64     `json:"interrupts"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ProcessSnapshot provides process-level information
type ProcessSnapshot struct {
	TotalProcesses     int                    `json:"total_processes"`
	RunningProcesses   int                    `json:"running_processes"`
	SleepingProcesses  int                    `json:"sleeping_processes"`
	ZombieProcesses    int                    `json:"zombie_processes"`
	TopCPUProcesses    []ProcessInfo          `json:"top_cpu_processes"`
	TopMemoryProcesses []ProcessInfo          `json:"top_memory_processes"`
	SystemdServices    []ServiceStatus        `json:"systemd_services"`
	NewProcesses       []ProcessInfo          `json:"new_processes"`       // Started in last interval
	DeadProcesses      []ProcessInfo          `json:"dead_processes"`      // Terminated in last interval
	LastUpdated        time.Time              `json:"last_updated"`
}

// ProcessInfo represents information about a single process
type ProcessInfo struct {
	PID         int     `json:"pid"`
	Name        string  `json:"name"`
	Command     string  `json:"command"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryMB    float64 `json:"memory_mb"`
	MemoryVMS   int64   `json:"memory_vms"`    // Virtual memory size
	MemoryRSS   int64   `json:"memory_rss"`    // Resident set size
	FDCount     int     `json:"fd_count"`      // File descriptor count
	ThreadCount int     `json:"thread_count"`
	StartTime   time.Time `json:"start_time"`
	Status      string  `json:"status"`
	Priority    int     `json:"priority"`
	Nice        int     `json:"nice"`
}

// ServiceStatus represents systemd service status
type ServiceStatus struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`      // active, inactive, failed, etc.
	SubState    string    `json:"sub_state"`   // running, dead, exited, etc.
	LoadState   string    `json:"load_state"`  // loaded, not-found, masked, etc.
	StartTime   time.Time `json:"start_time"`
	MemoryUsage int64     `json:"memory_usage"`
	CPUUsage    float64   `json:"cpu_usage"`
	Restarts    int       `json:"restarts"`
	LastRestart *time.Time `json:"last_restart,omitempty"`
}

// NetworkSnapshot provides network activity information
type NetworkSnapshot struct {
	Interfaces      []NetworkInterface `json:"interfaces"`
	Connections     ConnectionStats    `json:"connections"`
	DNSRequests     int64              `json:"dns_requests"`
	PacketErrors    int64              `json:"packet_errors"`
	TotalBandwidth  float64            `json:"total_bandwidth"`  // Mbps
	ActiveFlows     int                `json:"active_flows"`
	LatencyStats    LatencyInfo        `json:"latency_stats"`
	ThroughputStats ThroughputInfo     `json:"throughput_stats"`
	LastUpdated     time.Time          `json:"last_updated"`
}

// NetworkInterface represents network interface statistics
type NetworkInterface struct {
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	BytesReceived   int64     `json:"bytes_received"`
	BytesSent       int64     `json:"bytes_sent"`
	PacketsReceived int64     `json:"packets_received"`
	PacketsSent     int64     `json:"packets_sent"`
	DropReceived    int64     `json:"drop_received"`
	DropSent        int64     `json:"drop_sent"`
	ErrorsReceived  int64     `json:"errors_received"`
	ErrorsSent      int64     `json:"errors_sent"`
	Speed           int64     `json:"speed"`          // Mbps
	Duplex          string    `json:"duplex"`         // full, half
	MTU             int       `json:"mtu"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ConnectionStats represents network connection statistics
type ConnectionStats struct {
	TCP4Established int `json:"tcp4_established"`
	TCP4Listen      int `json:"tcp4_listen"`
	TCP4TimeWait    int `json:"tcp4_time_wait"`
	TCP6Established int `json:"tcp6_established"`
	TCP6Listen      int `json:"tcp6_listen"`
	TCP6TimeWait    int `json:"tcp6_time_wait"`
	UDPSockets      int `json:"udp_sockets"`
	UnixSockets     int `json:"unix_sockets"`
}

// LatencyInfo represents network latency statistics
type LatencyInfo struct {
	DNS       float64   `json:"dns"`        // DNS resolution time (ms)
	Connect   float64   `json:"connect"`    // Connection establishment time (ms)
	TLS       float64   `json:"tls"`        // TLS handshake time (ms)
	RTT       float64   `json:"rtt"`        // Round-trip time (ms)
	Jitter    float64   `json:"jitter"`     // Jitter (ms)
	PacketLoss float64  `json:"packet_loss"` // Packet loss percentage
	Timestamp time.Time `json:"timestamp"`
}

// ThroughputInfo represents network throughput statistics
type ThroughputInfo struct {
	DownloadMbps float64   `json:"download_mbps"`
	UploadMbps   float64   `json:"upload_mbps"`
	PeakDownload float64   `json:"peak_download"`
	PeakUpload   float64   `json:"peak_upload"`
	Utilization  float64   `json:"utilization"`   // Percentage of available bandwidth
	Timestamp    time.Time `json:"timestamp"`
}

// ResourcePressureInfo represents system resource pressure information
type ResourcePressureInfo struct {
	CPU           PressureMetric `json:"cpu"`
	Memory        PressureMetric `json:"memory"`
	IO            PressureMetric `json:"io"`
	SwapActivity  SwapInfo       `json:"swap_activity"`
	PageFaults    PageFaultInfo  `json:"page_faults"`
	ThermalState  ThermalInfo    `json:"thermal_state"`
	PowerState    PowerInfo      `json:"power_state"`
	LastUpdated   time.Time      `json:"last_updated"`
}

// PressureMetric represents resource pressure measurements
type PressureMetric struct {
	Some10    float64 `json:"some_10"`     // 10s average (some)
	Some60    float64 `json:"some_60"`     // 60s average (some)
	Some300   float64 `json:"some_300"`    // 300s average (some)
	Full10    float64 `json:"full_10"`     // 10s average (full)
	Full60    float64 `json:"full_60"`     // 60s average (full)
	Full300   float64 `json:"full_300"`    // 300s average (full)
	Total     int64   `json:"total"`       // Total accumulated time
}

// SwapInfo represents swap activity information
type SwapInfo struct {
	SwapTotal  int64   `json:"swap_total"`
	SwapUsed   int64   `json:"swap_used"`
	SwapFree   int64   `json:"swap_free"`
	SwapPct    float64 `json:"swap_pct"`
	SwapIn     int64   `json:"swap_in"`     // Pages swapped in
	SwapOut    int64   `json:"swap_out"`    // Pages swapped out
	SwapCached int64   `json:"swap_cached"` // Swap cached
}

// PageFaultInfo represents page fault statistics
type PageFaultInfo struct {
	MinorFaults int64 `json:"minor_faults"`
	MajorFaults int64 `json:"major_faults"`
	FaultRate   float64 `json:"fault_rate"` // Faults per second
}

// ThermalInfo represents thermal monitoring information
type ThermalInfo struct {
	CPUTemp     float64            `json:"cpu_temp"`
	GPUTemp     float64            `json:"gpu_temp,omitempty"`
	Zones       []ThermalZone      `json:"zones"`
	CoolingState CoolingInfo       `json:"cooling_state"`
	ThermalEvents []ThermalEvent   `json:"thermal_events"`
}

// ThermalZone represents a thermal zone
type ThermalZone struct {
	Name    string  `json:"name"`
	Temp    float64 `json:"temp"`
	Type    string  `json:"type"`
	Policy  string  `json:"policy"`
	Trips   []ThermalTrip `json:"trips"`
}

// ThermalTrip represents thermal trip points
type ThermalTrip struct {
	Type    string  `json:"type"`
	Temp    float64 `json:"temp"`
	Hyst    float64 `json:"hyst"`
}

// CoolingInfo represents cooling device information
type CoolingInfo struct {
	Fans     []FanInfo `json:"fans"`
	Throttling bool    `json:"throttling"`
	Policy   string   `json:"policy"`
}

// FanInfo represents fan information
type FanInfo struct {
	Name  string `json:"name"`
	Speed int    `json:"speed"` // RPM
	Level int    `json:"level"` // 0-100
}

// ThermalEvent represents thermal events
type ThermalEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Zone      string    `json:"zone"`
	Temp      float64   `json:"temp"`
	Action    string    `json:"action"`
}

// PowerInfo represents power management information
type PowerInfo struct {
	State         string      `json:"state"`          // performance, powersave, ondemand, etc.
	GovernorCPU   string      `json:"governor_cpu"`
	GovernorGPU   string      `json:"governor_gpu,omitempty"`
	CPUFreq       []FreqInfo  `json:"cpu_freq"`
	PowerUsage    PowerUsage  `json:"power_usage"`
	BatteryInfo   *BatteryInfo `json:"battery_info,omitempty"`
	WakeupSources []WakeupSource `json:"wakeup_sources"`
}

// FreqInfo represents CPU frequency information
type FreqInfo struct {
	CPU     int     `json:"cpu"`
	Current float64 `json:"current"`  // MHz
	Min     float64 `json:"min"`      // MHz
	Max     float64 `json:"max"`      // MHz
	Driver  string  `json:"driver"`
}

// PowerUsage represents power consumption information
type PowerUsage struct {
	TotalWatts   float64 `json:"total_watts"`
	CPUWatts     float64 `json:"cpu_watts"`
	GPUWatts     float64 `json:"gpu_watts,omitempty"`
	RAMWatts     float64 `json:"ram_watts"`
	DiskWatts    float64 `json:"disk_watts"`
	NetworkWatts float64 `json:"network_watts"`
}

// BatteryInfo represents battery information (for laptops)
type BatteryInfo struct {
	Present       bool    `json:"present"`
	Status        string  `json:"status"`        // Charging, Discharging, Full
	Capacity      float64 `json:"capacity"`      // Percentage
	VoltageNow    float64 `json:"voltage_now"`   // Volts
	CurrentNow    float64 `json:"current_now"`   // Amperes
	PowerNow      float64 `json:"power_now"`     // Watts
	Technology    string  `json:"technology"`
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	CycleCount    int     `json:"cycle_count"`
	Health        float64 `json:"health"`        // Percentage
	TimeRemaining time.Duration `json:"time_remaining"`
}

// WakeupSource represents wakeup sources
type WakeupSource struct {
	Name       string `json:"name"`
	Active     bool   `json:"active"`
	Count      int64  `json:"count"`
	ActiveTime time.Duration `json:"active_time"`
}

// SecurityEvent represents security-related events
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Type        string                 `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// MetricsRingBuffer provides efficient storage for metric history
type MetricsRingBuffer struct {
	mu       sync.RWMutex
	data     []MetricPoint
	capacity int
	size     int
	head     int
}

// MetricPoint represents a single metric measurement
type MetricPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Metrics   map[string]float64     `json:"metrics"`
}

// RealtimeAnomalyDetector provides real-time anomaly detection
type RealtimeAnomalyDetector struct {
	mu                    sync.RWMutex
	logger                *logger.Logger
	baselineWindow        time.Duration
	sensitivityThreshold  float64
	minDataPoints         int
	detectionEnabled      bool
	metricBaselines       map[string]*MetricBaseline
	anomaliesDetected     []RealtimeAnomaly
	lastDetectionRun      time.Time
}

// MetricBaseline represents baseline statistics for a metric
type MetricBaseline struct {
	Mean        float64
	StdDev      float64
	Min         float64
	Max         float64
	Percentile95 float64
	Percentile99 float64
	SampleCount int
	LastUpdated time.Time
}

// RealtimeAnomaly represents a detected anomaly
type RealtimeAnomaly struct {
	ID            string    `json:"id"`
	Metric        string    `json:"metric"`
	Value         float64   `json:"value"`
	ExpectedRange [2]float64 `json:"expected_range"` // [min, max]
	DeviationSigmas float64 `json:"deviation_sigmas"`
	Confidence    float64   `json:"confidence"`
	DetectedAt    time.Time `json:"detected_at"`
	Duration      time.Duration `json:"duration"`
	Resolved      bool      `json:"resolved"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

// Alert types and severities
type AlertType string
type AlertSeverity string

const (
	AlertTypeMetric     AlertType = "metric"
	AlertTypeAnomaly    AlertType = "anomaly"
	AlertTypeProcess    AlertType = "process"
	AlertTypeService    AlertType = "service"
	AlertTypeNetwork    AlertType = "network"
	AlertTypeSecurity   AlertType = "security"
	AlertTypeThermal    AlertType = "thermal"
	AlertTypePower      AlertType = "power"

	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)

// NewRealTimeMonitor creates a new real-time monitor
func NewRealTimeMonitor(systemMonitor *SystemMonitor) *RealTimeMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	rtm := &RealTimeMonitor{
		logger:           logger.NewLogger(),
		systemMonitor:    systemMonitor,
		subscribers:      make(map[string]chan *SystemSnapshot),
		alertSubscribers: make(map[string]chan *SystemAlert),
		activeAlerts:     make(map[string]*SystemAlert),
		snapshotInterval: 1 * time.Second, // 1 second real-time updates
		ctx:              ctx,
		cancel:           cancel,
		metricsHistory:   NewMetricsRingBuffer(3600), // Store 1 hour of 1-second data
		anomalyDetector:  NewRealtimeAnomalyDetector(),
	}
	
	rtm.initializeDefaultThresholds()
	return rtm
}

// Start begins real-time monitoring
func (rtm *RealTimeMonitor) Start() error {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	if rtm.running {
		return fmt.Errorf("real-time monitor already running")
	}
	
	rtm.logger.Info("Starting real-time system monitoring")
	
	// Start the main monitoring loop
	go rtm.monitoringLoop()
	
	// Start alert processing
	go rtm.alertProcessingLoop()
	
	// Start anomaly detection
	go rtm.anomalyDetectionLoop()
	
	rtm.running = true
	rtm.logger.Info("Real-time monitoring started successfully")
	
	return nil
}

// Stop halts real-time monitoring
func (rtm *RealTimeMonitor) Stop() error {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	if !rtm.running {
		return nil
	}
	
	rtm.logger.Info("Stopping real-time system monitoring")
	
	rtm.cancel()
	rtm.running = false
	
	// Close all subscriber channels
	for id, ch := range rtm.subscribers {
		close(ch)
		delete(rtm.subscribers, id)
	}
	
	for id, ch := range rtm.alertSubscribers {
		close(ch)
		delete(rtm.alertSubscribers, id)
	}
	
	rtm.logger.Info("Real-time monitoring stopped")
	return nil
}

// Subscribe registers a new subscriber for real-time updates
func (rtm *RealTimeMonitor) Subscribe(subscriberID string) (<-chan *SystemSnapshot, error) {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	if !rtm.running {
		return nil, fmt.Errorf("real-time monitor not running")
	}
	
	// Create buffered channel to prevent blocking
	ch := make(chan *SystemSnapshot, 10)
	rtm.subscribers[subscriberID] = ch
	
	rtm.logger.Info(fmt.Sprintf("Subscriber %s registered for real-time updates", subscriberID))
	return ch, nil
}

// Unsubscribe removes a subscriber
func (rtm *RealTimeMonitor) Unsubscribe(subscriberID string) {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	if ch, exists := rtm.subscribers[subscriberID]; exists {
		close(ch)
		delete(rtm.subscribers, subscriberID)
		rtm.logger.Info(fmt.Sprintf("Subscriber %s unsubscribed", subscriberID))
	}
}

// SubscribeToAlerts registers a subscriber for alert notifications
func (rtm *RealTimeMonitor) SubscribeToAlerts(subscriberID string) (<-chan *SystemAlert, error) {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	if !rtm.running {
		return nil, fmt.Errorf("real-time monitor not running")
	}
	
	ch := make(chan *SystemAlert, 50) // Larger buffer for alerts
	rtm.alertSubscribers[subscriberID] = ch
	
	rtm.logger.Info(fmt.Sprintf("Subscriber %s registered for alert notifications", subscriberID))
	return ch, nil
}

// UnsubscribeFromAlerts removes an alert subscriber
func (rtm *RealTimeMonitor) UnsubscribeFromAlerts(subscriberID string) {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	if ch, exists := rtm.alertSubscribers[subscriberID]; exists {
		close(ch)
		delete(rtm.alertSubscribers, subscriberID)
		rtm.logger.Info(fmt.Sprintf("Alert subscriber %s unsubscribed", subscriberID))
	}
}

// GetLatestSnapshot returns the most recent system snapshot
func (rtm *RealTimeMonitor) GetLatestSnapshot() (*SystemSnapshot, error) {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()
	
	if !rtm.running {
		return nil, fmt.Errorf("real-time monitor not running")
	}
	
	return rtm.createSnapshot(), nil
}

// GetActiveAlerts returns all currently active alerts
func (rtm *RealTimeMonitor) GetActiveAlerts() []*SystemAlert {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()
	
	alerts := make([]*SystemAlert, 0, len(rtm.activeAlerts))
	for _, alert := range rtm.activeAlerts {
		alerts = append(alerts, alert)
	}
	
	return alerts
}

// AcknowledgeAlert acknowledges an active alert
func (rtm *RealTimeMonitor) AcknowledgeAlert(alertID, acknowledgedBy string) error {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	alert, exists := rtm.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}
	
	now := time.Now()
	alert.Acknowledged = true
	alert.AcknowledgedBy = acknowledgedBy
	alert.AcknowledgedAt = &now
	
	rtm.logger.Info(fmt.Sprintf("Alert %s acknowledged by %s", alertID, acknowledgedBy))
	return nil
}

// Private methods

func (rtm *RealTimeMonitor) monitoringLoop() {
	ticker := time.NewTicker(rtm.snapshotInterval)
	defer ticker.Stop()
	
	sequenceNumber := int64(0)
	
	for {
		select {
		case <-rtm.ctx.Done():
			return
		case <-ticker.C:
			sequenceNumber++
			snapshot := rtm.createSnapshot()
			snapshot.SequenceNumber = sequenceNumber
			
			// Store in history
			rtm.storeMetrics(snapshot)
			
			// Send to subscribers
			rtm.broadcastSnapshot(snapshot)
		}
	}
}

func (rtm *RealTimeMonitor) createSnapshot() *SystemSnapshot {
	now := time.Now()
	
	// Get current metrics from system monitor
	rtm.systemMonitor.updateMetrics()
	metrics := rtm.systemMonitor.GetCurrentMetrics()
	
	// Calculate component health
	componentHealth := rtm.calculateComponentHealth(metrics)
	
	// Get active alerts
	activeAlerts := make([]SystemAlert, 0, len(rtm.activeAlerts))
	for _, alert := range rtm.activeAlerts {
		activeAlerts = append(activeAlerts, *alert)
	}
	
	// Calculate trend indicators
	trendIndicators := rtm.calculateTrends(metrics)
	
	// Collect detailed system information
	systemLoad := rtm.collectSystemLoad()
	processInfo := rtm.collectProcessInfo()
	networkActivity := rtm.collectNetworkActivity()
	resourcePressure := rtm.collectResourcePressure()
	securityEvents := rtm.collectSecurityEvents()
	
	return &SystemSnapshot{
		Timestamp:        now,
		Metrics:          metrics,
		ComponentHealth:  componentHealth,
		Alerts:           activeAlerts,
		TrendIndicators:  trendIndicators,
		SystemLoad:       systemLoad,
		ProcessInfo:      processInfo,
		NetworkActivity:  networkActivity,
		ResourcePressure: resourcePressure,
		SecurityEvents:   securityEvents,
	}
}

func (rtm *RealTimeMonitor) storeMetrics(snapshot *SystemSnapshot) {
	point := MetricPoint{
		Timestamp: snapshot.Timestamp,
		Metrics:   snapshot.Metrics,
	}
	rtm.metricsHistory.Add(point)
}

func (rtm *RealTimeMonitor) broadcastSnapshot(snapshot *SystemSnapshot) {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()
	
	for subscriberID, ch := range rtm.subscribers {
		select {
		case ch <- snapshot:
			// Successfully sent
		default:
			// Channel buffer full, skip this update
			rtm.logger.Warn(fmt.Sprintf("Skipping update for subscriber %s (buffer full)", subscriberID))
		}
	}
}

func (rtm *RealTimeMonitor) alertProcessingLoop() {
	ticker := time.NewTicker(5 * time.Second) // Check for alerts every 5 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-rtm.ctx.Done():
			return
		case <-ticker.C:
			rtm.processAlerts()
		}
	}
}

func (rtm *RealTimeMonitor) processAlerts() {
	rtm.mu.Lock()
	defer rtm.mu.Unlock()
	
	// Get current metrics
	metrics := rtm.systemMonitor.GetCurrentMetrics()
	
	// Check each metric against thresholds
	for metricName, value := range metrics {
		threshold, exists := rtm.alertThresholds[metricName]
		if !exists {
			continue
		}
		
		alertID := fmt.Sprintf("metric_%s", metricName)
		existingAlert, hasActiveAlert := rtm.activeAlerts[alertID]
		
		severity := rtm.determineSeverity(value, threshold)
		
		if severity != "" {
			if hasActiveAlert {
				// Update existing alert
				existingAlert.CurrentValue = value
				existingAlert.LastUpdated = time.Now()
				existingAlert.Duration = time.Since(existingAlert.FirstDetected)
				existingAlert.TriggerCount++
				
				if existingAlert.Severity != severity {
					existingAlert.Severity = severity
					rtm.broadcastAlert(existingAlert)
				}
			} else {
				// Create new alert
				alert := &SystemAlert{
					ID:             alertID,
					Type:           AlertTypeMetric,
					Severity:       severity,
					Component:      rtm.getComponentForMetric(metricName),
					Metric:         metricName,
					CurrentValue:   value,
					ThresholdValue: rtm.getThresholdValue(value, threshold),
					Message:        rtm.generateAlertMessage(metricName, value, severity),
					FirstDetected:  time.Now(),
					LastUpdated:    time.Now(),
					TriggerCount:   1,
					ResolutionHints: rtm.getResolutionHints(metricName, severity),
					Metadata:       make(map[string]interface{}),
				}
				
				rtm.activeAlerts[alertID] = alert
				rtm.broadcastAlert(alert)
				rtm.logger.Warn(fmt.Sprintf("New alert: %s - %s", alert.ID, alert.Message))
			}
		} else if hasActiveAlert {
			// Alert condition resolved
			delete(rtm.activeAlerts, alertID)
			rtm.logger.Info(fmt.Sprintf("Alert resolved: %s", alertID))
		}
	}
}

func (rtm *RealTimeMonitor) broadcastAlert(alert *SystemAlert) {
	rtm.mu.RLock()
	defer rtm.mu.RUnlock()
	
	for subscriberID, ch := range rtm.alertSubscribers {
		select {
		case ch <- alert:
			// Successfully sent
		default:
			// Channel buffer full, skip this alert
			rtm.logger.Warn(fmt.Sprintf("Skipping alert for subscriber %s (buffer full)", subscriberID))
		}
	}
}

func (rtm *RealTimeMonitor) anomalyDetectionLoop() {
	ticker := time.NewTicker(10 * time.Second) // Run anomaly detection every 10 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-rtm.ctx.Done():
			return
		case <-ticker.C:
			rtm.runAnomalyDetection()
		}
	}
}

func (rtm *RealTimeMonitor) runAnomalyDetection() {
	// Get recent metrics history
	recentData := rtm.metricsHistory.GetRecentData(5 * time.Minute)
	if len(recentData) < 10 {
		return // Not enough data for meaningful detection
	}
	
	// Run anomaly detection
	anomalies := rtm.anomalyDetector.DetectAnomalies(recentData)
	
	// Convert anomalies to alerts
	for _, anomaly := range anomalies {
		alertID := fmt.Sprintf("anomaly_%s_%s", anomaly.Metric, anomaly.ID)
		
		if _, exists := rtm.activeAlerts[alertID]; !exists {
			alert := &SystemAlert{
				ID:           alertID,
				Type:         AlertTypeAnomaly,
				Severity:     rtm.anomalySeverityToAlert(anomaly.Confidence),
				Component:    rtm.getComponentForMetric(anomaly.Metric),
				Metric:       anomaly.Metric,
				CurrentValue: anomaly.Value,
				Message:      fmt.Sprintf("Anomaly detected in %s: value %.2f deviates %.1f sigmas from baseline", 
					anomaly.Metric, anomaly.Value, anomaly.DeviationSigmas),
				FirstDetected: anomaly.DetectedAt,
				LastUpdated:   time.Now(),
				TriggerCount:  1,
				ResolutionHints: []string{
					"Investigate recent system changes",
					"Check application logs for errors",
					"Monitor metric for sustained anomalous behavior",
				},
				Metadata: map[string]interface{}{
					"anomaly_id":        anomaly.ID,
					"deviation_sigmas":  anomaly.DeviationSigmas,
					"confidence":        anomaly.Confidence,
					"expected_range":    anomaly.ExpectedRange,
				},
			}
			
			rtm.mu.Lock()
			rtm.activeAlerts[alertID] = alert
			rtm.mu.Unlock()
			
			rtm.broadcastAlert(alert)
			rtm.logger.Warn(fmt.Sprintf("Anomaly alert: %s", alert.Message))
		}
	}
}

// Helper functions for data collection (simplified implementations)

func (rtm *RealTimeMonitor) collectSystemLoad() SystemLoadInfo {
	// This would collect detailed system load information
	// Simplified implementation for now
	return SystemLoadInfo{
		LoadAvg1:       0.5,
		LoadAvg5:       0.6,
		LoadAvg15:      0.4,
		CPUCores:       8,
		CPUUtilization: 15.0,
		CPUPressure:    0.1,
		MemoryPressure: 0.2,
		IOPressure:     0.05,
		LastUpdated:    time.Now(),
	}
}

func (rtm *RealTimeMonitor) collectProcessInfo() ProcessSnapshot {
	// This would collect detailed process information
	// Simplified implementation for now
	return ProcessSnapshot{
		TotalProcesses:    150,
		RunningProcesses:  5,
		SleepingProcesses: 145,
		ZombieProcesses:   0,
		TopCPUProcesses:   []ProcessInfo{},
		TopMemoryProcesses: []ProcessInfo{},
		SystemdServices:   []ServiceStatus{},
		NewProcesses:      []ProcessInfo{},
		DeadProcesses:     []ProcessInfo{},
		LastUpdated:       time.Now(),
	}
}

func (rtm *RealTimeMonitor) collectNetworkActivity() NetworkSnapshot {
	// This would collect detailed network information
	// Simplified implementation for now
	return NetworkSnapshot{
		Interfaces:      []NetworkInterface{},
		Connections:     ConnectionStats{},
		TotalBandwidth:  100.0,
		ActiveFlows:     25,
		LastUpdated:     time.Now(),
	}
}

func (rtm *RealTimeMonitor) collectResourcePressure() ResourcePressureInfo {
	// This would collect resource pressure information from /proc/pressure/*
	// Simplified implementation for now
	return ResourcePressureInfo{
		CPU: PressureMetric{
			Some10:  5.0,
			Some60:  3.0,
			Some300: 2.0,
		},
		Memory: PressureMetric{
			Some10:  2.0,
			Some60:  1.5,
			Some300: 1.0,
		},
		IO: PressureMetric{
			Some10:  1.0,
			Some60:  0.8,
			Some300: 0.5,
		},
		LastUpdated: time.Now(),
	}
}

func (rtm *RealTimeMonitor) collectSecurityEvents() []SecurityEvent {
	// This would collect security events from logs and monitoring
	// Simplified implementation for now
	return []SecurityEvent{}
}

// Utility functions

func (rtm *RealTimeMonitor) initializeDefaultThresholds() {
	rtm.alertThresholds = map[string]AlertThreshold{
		"cpu_usage": {
			Warning:     75.0,
			Critical:    90.0,
			Emergency:   98.0,
			CheckPeriod: 30 * time.Second,
			Hysteresis:  5.0,
			MinDuration: 30 * time.Second,
		},
		"memory_usage": {
			Warning:     80.0,
			Critical:    95.0,
			Emergency:   99.0,
			CheckPeriod: 30 * time.Second,
			Hysteresis:  5.0,
			MinDuration: 30 * time.Second,
		},
		"disk_usage": {
			Warning:     80.0,
			Critical:    90.0,
			Emergency:   95.0,
			CheckPeriod: 60 * time.Second,
			Hysteresis:  2.0,
			MinDuration: 60 * time.Second,
		},
		"load_average": {
			Warning:     float64(8) * 0.75,  // 75% of CPU cores
			Critical:    float64(8) * 1.5,   // 150% of CPU cores
			Emergency:   float64(8) * 2.5,   // 250% of CPU cores
			CheckPeriod: 30 * time.Second,
			Hysteresis:  0.5,
			MinDuration: 30 * time.Second,
		},
	}
}

func (rtm *RealTimeMonitor) calculateComponentHealth(metrics map[string]float64) map[string]HealthStatus {
	// Reuse logic from main health predictor
	// Simplified implementation
	return map[string]HealthStatus{
		"cpu":     HealthExcellent,
		"memory":  HealthGood,
		"disk":    HealthFair,
		"network": HealthExcellent,
		"system":  HealthGood,
	}
}

func (rtm *RealTimeMonitor) calculateTrends(metrics map[string]float64) map[string]TrendMetric {
	trends := make(map[string]TrendMetric)
	
	// Get historical data for trend calculation
	recentData := rtm.metricsHistory.GetRecentData(5 * time.Minute)
	if len(recentData) < 2 {
		return trends
	}
	
	for metricName, currentValue := range metrics {
		trend := rtm.calculateMetricTrend(metricName, currentValue, recentData)
		trends[metricName] = trend
	}
	
	return trends
}

func (rtm *RealTimeMonitor) calculateMetricTrend(metricName string, currentValue float64, recentData []MetricPoint) TrendMetric {
	if len(recentData) < 2 {
		return TrendMetric{
			Current:        currentValue,
			Previous:       currentValue,
			Direction:      "stable",
			LastCalculated: time.Now(),
		}
	}
	
	// Simple trend calculation
	previousValue := recentData[len(recentData)-2].Metrics[metricName]
	changeRate := (currentValue - previousValue) / 1.0 // Per minute
	
	direction := "stable"
	if changeRate > 0.1 {
		direction = "increasing"
	} else if changeRate < -0.1 {
		direction = "decreasing"
	}
	
	return TrendMetric{
		Current:        currentValue,
		Previous:       previousValue,
		ChangeRate:     changeRate,
		Direction:      direction,
		Prediction1m:   currentValue + changeRate,
		Prediction5m:   currentValue + (changeRate * 5),
		Confidence:     0.7,
		LastCalculated: time.Now(),
	}
}

func (rtm *RealTimeMonitor) determineSeverity(value float64, threshold AlertThreshold) AlertSeverity {
	if value >= threshold.Emergency {
		return AlertSeverityEmergency
	} else if value >= threshold.Critical {
		return AlertSeverityCritical
	} else if value >= threshold.Warning {
		return AlertSeverityWarning
	}
	return ""
}

func (rtm *RealTimeMonitor) getThresholdValue(value float64, threshold AlertThreshold) float64 {
	if value >= threshold.Emergency {
		return threshold.Emergency
	} else if value >= threshold.Critical {
		return threshold.Critical
	} else if value >= threshold.Warning {
		return threshold.Warning
	}
	return threshold.Warning
}

func (rtm *RealTimeMonitor) getComponentForMetric(metricName string) string {
	switch metricName {
	case "cpu_usage", "load_average":
		return "cpu"
	case "memory_usage":
		return "memory"
	case "disk_usage":
		return "disk"
	case "network_usage":
		return "network"
	default:
		return "system"
	}
}

func (rtm *RealTimeMonitor) generateAlertMessage(metricName string, value float64, severity AlertSeverity) string {
	return fmt.Sprintf("%s %s: %.2f%% - %s threshold exceeded", 
		metricName, severity, value, severity)
}

func (rtm *RealTimeMonitor) getResolutionHints(metricName string, severity AlertSeverity) []string {
	switch metricName {
	case "cpu_usage":
		return []string{
			"Check for high CPU processes with 'top' or 'htop'",
			"Consider restarting high-usage services",
			"Check for CPU-intensive background tasks",
		}
	case "memory_usage":
		return []string{
			"Check memory usage with 'free -h'",
			"Look for memory leaks in running processes",
			"Consider restarting memory-intensive services",
		}
	case "disk_usage":
		return []string{
			"Run 'nix-collect-garbage -d' to free disk space",
			"Check for large log files with 'journalctl --disk-usage'",
			"Use 'nix-store --optimize' to deduplicate store",
		}
	default:
		return []string{"Investigate the issue manually"}
	}
}

func (rtm *RealTimeMonitor) anomalySeverityToAlert(confidence float64) AlertSeverity {
	if confidence >= 0.9 {
		return AlertSeverityCritical
	} else if confidence >= 0.7 {
		return AlertSeverityWarning
	}
	return AlertSeverityInfo
}

// MetricsRingBuffer implementation

func NewMetricsRingBuffer(capacity int) *MetricsRingBuffer {
	return &MetricsRingBuffer{
		data:     make([]MetricPoint, capacity),
		capacity: capacity,
		size:     0,
		head:     0,
	}
}

func (mrb *MetricsRingBuffer) Add(point MetricPoint) {
	mrb.mu.Lock()
	defer mrb.mu.Unlock()
	
	mrb.data[mrb.head] = point
	mrb.head = (mrb.head + 1) % mrb.capacity
	
	if mrb.size < mrb.capacity {
		mrb.size++
	}
}

func (mrb *MetricsRingBuffer) GetRecentData(duration time.Duration) []MetricPoint {
	mrb.mu.RLock()
	defer mrb.mu.RUnlock()
	
	if mrb.size == 0 {
		return []MetricPoint{}
	}
	
	cutoff := time.Now().Add(-duration)
	var result []MetricPoint
	
	// Iterate through the ring buffer
	for i := 0; i < mrb.size; i++ {
		idx := (mrb.head - 1 - i + mrb.capacity) % mrb.capacity
		point := mrb.data[idx]
		
		if point.Timestamp.After(cutoff) {
			result = append([]MetricPoint{point}, result...) // Prepend to maintain chronological order
		} else {
			break // Data is older than cutoff
		}
	}
	
	return result
}

// RealtimeAnomalyDetector implementation

func NewRealtimeAnomalyDetector() *RealtimeAnomalyDetector {
	return &RealtimeAnomalyDetector{
		logger:               logger.NewLogger(),
		baselineWindow:       30 * time.Minute,
		sensitivityThreshold: 2.0, // 2 standard deviations
		minDataPoints:        30,
		detectionEnabled:     true,
		metricBaselines:      make(map[string]*MetricBaseline),
		anomaliesDetected:    make([]RealtimeAnomaly, 0),
	}
}

func (rad *RealtimeAnomalyDetector) DetectAnomalies(recentData []MetricPoint) []RealtimeAnomaly {
	rad.mu.Lock()
	defer rad.mu.Unlock()
	
	if !rad.detectionEnabled || len(recentData) < rad.minDataPoints {
		return []RealtimeAnomaly{}
	}
	
	var anomalies []RealtimeAnomaly
	
	// Update baselines with recent data
	rad.updateBaselines(recentData)
	
	// Check latest data point for anomalies
	if len(recentData) > 0 {
		latestPoint := recentData[len(recentData)-1]
		
		for metricName, value := range latestPoint.Metrics {
			baseline, exists := rad.metricBaselines[metricName]
			if !exists || baseline.SampleCount < rad.minDataPoints {
				continue
			}
			
			// Calculate deviation
			deviation := (value - baseline.Mean) / baseline.StdDev
			
			if deviation > rad.sensitivityThreshold || deviation < -rad.sensitivityThreshold {
				anomaly := RealtimeAnomaly{
					ID:              fmt.Sprintf("anomaly_%s_%d", metricName, time.Now().Unix()),
					Metric:          metricName,
					Value:           value,
					ExpectedRange:   [2]float64{baseline.Mean - rad.sensitivityThreshold*baseline.StdDev, baseline.Mean + rad.sensitivityThreshold*baseline.StdDev},
					DeviationSigmas: deviation,
					Confidence:      rad.calculateConfidence(deviation),
					DetectedAt:      latestPoint.Timestamp,
					Duration:        0,
					Resolved:        false,
				}
				
				anomalies = append(anomalies, anomaly)
			}
		}
	}
	
	rad.lastDetectionRun = time.Now()
	return anomalies
}

func (rad *RealtimeAnomalyDetector) updateBaselines(recentData []MetricPoint) {
	// Collect all metric values
	metricValues := make(map[string][]float64)
	
	for _, point := range recentData {
		for metricName, value := range point.Metrics {
			metricValues[metricName] = append(metricValues[metricName], value)
		}
	}
	
	// Update baselines
	for metricName, values := range metricValues {
		if len(values) < rad.minDataPoints {
			continue
		}
		
		baseline := rad.calculateBaseline(values)
		rad.metricBaselines[metricName] = baseline
	}
}

func (rad *RealtimeAnomalyDetector) calculateBaseline(values []float64) *MetricBaseline {
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// Calculate standard deviation
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	stddev := variance / float64(len(values))
	
	// Find min and max
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	return &MetricBaseline{
		Mean:         mean,
		StdDev:       stddev,
		Min:          min,
		Max:          max,
		SampleCount:  len(values),
		LastUpdated:  time.Now(),
	}
}

func (rad *RealtimeAnomalyDetector) calculateConfidence(deviation float64) float64 {
	// Higher deviation = higher confidence
	absDeviation := deviation
	if absDeviation < 0 {
		absDeviation = -absDeviation
	}
	
	confidence := absDeviation / 5.0 // Scale factor
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}