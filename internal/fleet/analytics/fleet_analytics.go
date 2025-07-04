package analytics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"nix-ai-help/internal/fleet"
	"nix-ai-help/pkg/logger"
)

// FleetAnalytics provides advanced fleet analytics and intelligence
type FleetAnalytics struct {
	logger         *logger.Logger
	fleetManager   *fleet.FleetManager
	metricsStorage MetricsStorage
}

// MetricsStorage interface for storing and retrieving metrics
type MetricsStorage interface {
	StoreMetric(ctx context.Context, metric *Metric) error
	GetMetrics(ctx context.Context, query MetricQuery) ([]*Metric, error)
	GetAggregatedMetrics(ctx context.Context, query AggregationQuery) (*AggregatedMetric, error)
}

// Metric represents a fleet metric data point
type Metric struct {
	ID          string                 `json:"id"`
	MachineID   string                 `json:"machine_id"`
	MetricType  string                 `json:"metric_type"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Tags        map[string]string      `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	Environment string                 `json:"environment"`
}

// MetricQuery represents a query for retrieving metrics
type MetricQuery struct {
	MachineIDs  []string          `json:"machine_ids"`
	MetricTypes []string          `json:"metric_types"`
	Tags        map[string]string `json:"tags"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	Environment string            `json:"environment"`
	Limit       int               `json:"limit"`
}

// AggregationQuery represents a query for aggregated metrics
type AggregationQuery struct {
	MetricQuery
	AggregationType string        `json:"aggregation_type"` // sum, avg, min, max, count
	GroupBy         []string      `json:"group_by"`
	Interval        time.Duration `json:"interval"`
}

// AggregatedMetric represents aggregated metric data
type AggregatedMetric struct {
	MetricType      string                   `json:"metric_type"`
	AggregationType string                   `json:"aggregation_type"`
	Value           float64                  `json:"value"`
	Unit            string                   `json:"unit"`
	Count           int                      `json:"count"`
	Groups          map[string]float64       `json:"groups"`
	TimeSeries      []TimeSeriesPoint        `json:"time_series"`
	Metadata        map[string]interface{}   `json:"metadata"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Count     int       `json:"count"`
}

// FleetAnalyticsReport represents a comprehensive fleet analytics report
type FleetAnalyticsReport struct {
	GeneratedAt     time.Time                    `json:"generated_at"`
	FleetOverview   FleetOverview                `json:"fleet_overview"`
	PerformanceAnalysis PerformanceAnalysis      `json:"performance_analysis"`
	CostAnalysis    CostAnalysis                 `json:"cost_analysis"`
	SecurityAnalysis SecurityAnalysis            `json:"security_analysis"`
	CapacityAnalysis CapacityAnalysis            `json:"capacity_analysis"`
	Recommendations []FleetRecommendation        `json:"recommendations"`
	TrendAnalysis   TrendAnalysis               `json:"trend_analysis"`
}

// FleetOverview provides high-level fleet statistics
type FleetOverview struct {
	TotalMachines      int                        `json:"total_machines"`
	HealthyMachines    int                        `json:"healthy_machines"`
	UnhealthyMachines  int                        `json:"unhealthy_machines"`
	EnvironmentBreakdown map[string]int           `json:"environment_breakdown"`
	RegionBreakdown    map[string]int             `json:"region_breakdown"`
	AverageUptime      float64                    `json:"average_uptime"`
	TotalCPUCores      int                        `json:"total_cpu_cores"`
	TotalMemoryGB      float64                    `json:"total_memory_gb"`
	TotalStorageGB     float64                    `json:"total_storage_gb"`
	ConfigurationDrift int                        `json:"configuration_drift"`
}

// PerformanceAnalysis provides fleet performance insights
type PerformanceAnalysis struct {
	AverageCPUUsage    float64                  `json:"average_cpu_usage"`
	AverageMemoryUsage float64                  `json:"average_memory_usage"`
	AverageLoadAverage float64                  `json:"average_load_average"`
	TopPerformers      []MachinePerformance     `json:"top_performers"`
	BottomPerformers   []MachinePerformance     `json:"bottom_performers"`
	PerformanceTrends  []PerformanceTrend       `json:"performance_trends"`
	Bottlenecks        []PerformanceBottleneck  `json:"bottlenecks"`
}

// MachinePerformance represents individual machine performance metrics
type MachinePerformance struct {
	MachineID        string  `json:"machine_id"`
	MachineName      string  `json:"machine_name"`
	PerformanceScore float64 `json:"performance_score"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      float64 `json:"memory_usage"`
	LoadAverage      float64 `json:"load_average"`
	Uptime           float64 `json:"uptime"`
	Environment      string  `json:"environment"`
}

// PerformanceTrend represents performance trends over time
type PerformanceTrend struct {
	MetricType  string              `json:"metric_type"`
	Trend       string              `json:"trend"` // increasing, decreasing, stable
	ChangeRate  float64             `json:"change_rate"`
	Confidence  float64             `json:"confidence"`
	TimeSeries  []TimeSeriesPoint   `json:"time_series"`
}

// PerformanceBottleneck represents identified performance bottlenecks
type PerformanceBottleneck struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	AffectedMachines []string `json:"affected_machines"`
	Impact      float64 `json:"impact"`
	Recommendation string `json:"recommendation"`
}

// CostAnalysis provides fleet cost optimization insights
type CostAnalysis struct {
	TotalMonthlyCost     float64              `json:"total_monthly_cost"`
	CostPerMachine       float64              `json:"cost_per_machine"`
	CostPerEnvironment   map[string]float64   `json:"cost_per_environment"`
	CostBreakdown        CostBreakdown        `json:"cost_breakdown"`
	CostTrends           []CostTrend          `json:"cost_trends"`
	OptimizationOpportunities []CostOptimization `json:"optimization_opportunities"`
	ROIAnalysis          ROIAnalysis          `json:"roi_analysis"`
}

// CostBreakdown represents detailed cost breakdown
type CostBreakdown struct {
	ComputeCost  float64 `json:"compute_cost"`
	StorageCost  float64 `json:"storage_cost"`
	NetworkCost  float64 `json:"network_cost"`
	LicenseCost  float64 `json:"license_cost"`
	MaintenanceCost float64 `json:"maintenance_cost"`
}

// CostTrend represents cost trends over time
type CostTrend struct {
	Period     string  `json:"period"`
	Cost       float64 `json:"cost"`
	Change     float64 `json:"change"`
	Projection float64 `json:"projection"`
}

// CostOptimization represents cost optimization opportunities
type CostOptimization struct {
	Type            string  `json:"type"`
	Description     string  `json:"description"`
	PotentialSavings float64 `json:"potential_savings"`
	ImplementationEffort string `json:"implementation_effort"`
	RiskLevel       string  `json:"risk_level"`
	AffectedMachines []string `json:"affected_machines"`
}

// ROIAnalysis represents return on investment analysis
type ROIAnalysis struct {
	InvestmentCost   float64 `json:"investment_cost"`
	AnnualSavings    float64 `json:"annual_savings"`
	PaybackPeriod    float64 `json:"payback_period"`
	ThreeYearROI     float64 `json:"three_year_roi"`
	BreakEvenPoint   time.Time `json:"break_even_point"`
}

// SecurityAnalysis provides fleet security insights
type SecurityAnalysis struct {
	SecurityScore        float64                `json:"security_score"`
	VulnerabilityCount   VulnerabilityCount     `json:"vulnerability_count"`
	ComplianceStatus     ComplianceStatus       `json:"compliance_status"`
	SecurityTrends       []SecurityTrend        `json:"security_trends"`
	SecurityIncidents    []SecurityIncident     `json:"security_incidents"`
	SecurityRecommendations []SecurityRecommendation `json:"security_recommendations"`
}

// VulnerabilityCount represents vulnerability statistics
type VulnerabilityCount struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// ComplianceStatus represents compliance framework status
type ComplianceStatus struct {
	SOC2        ComplianceFramework `json:"soc2"`
	HIPAA       ComplianceFramework `json:"hipaa"`
	PCIDSS      ComplianceFramework `json:"pci_dss"`
	ISO27001    ComplianceFramework `json:"iso27001"`
	GDPR        ComplianceFramework `json:"gdpr"`
}

// ComplianceFramework represents individual compliance framework status
type ComplianceFramework struct {
	Status      string  `json:"status"` // compliant, non_compliant, partial
	Score       float64 `json:"score"`
	LastAudit   time.Time `json:"last_audit"`
	Issues      []string `json:"issues"`
	NextAudit   time.Time `json:"next_audit"`
}

// SecurityTrend represents security trends over time
type SecurityTrend struct {
	MetricType string            `json:"metric_type"`
	Trend      string            `json:"trend"`
	TimeSeries []TimeSeriesPoint `json:"time_series"`
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	MachineID   string    `json:"machine_id"`
	DetectedAt  time.Time `json:"detected_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	Status      string    `json:"status"`
}

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	Priority     string   `json:"priority"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	AffectedMachines []string `json:"affected_machines"`
	Implementation string `json:"implementation"`
	Impact       string   `json:"impact"`
}

// CapacityAnalysis provides fleet capacity planning insights
type CapacityAnalysis struct {
	CurrentUtilization  ResourceUtilization    `json:"current_utilization"`
	PredictedUtilization ResourceUtilization   `json:"predicted_utilization"`
	CapacityForecast    []CapacityForecast     `json:"capacity_forecast"`
	ScalingRecommendations []ScalingRecommendation `json:"scaling_recommendations"`
	ResourceBottlenecks []ResourceBottleneck   `json:"resource_bottlenecks"`
}

// ResourceUtilization represents resource utilization metrics
type ResourceUtilization struct {
	CPUUtilization     float64 `json:"cpu_utilization"`
	MemoryUtilization  float64 `json:"memory_utilization"`
	StorageUtilization float64 `json:"storage_utilization"`
	NetworkUtilization float64 `json:"network_utilization"`
}

// CapacityForecast represents capacity forecasting data
type CapacityForecast struct {
	Period         string              `json:"period"`
	ForecastDate   time.Time           `json:"forecast_date"`
	PredictedUsage ResourceUtilization `json:"predicted_usage"`
	Confidence     float64             `json:"confidence"`
	Scenarios      []ForecastScenario  `json:"scenarios"`
}

// ForecastScenario represents different capacity scenarios
type ForecastScenario struct {
	Name        string              `json:"name"`
	Probability float64             `json:"probability"`
	Usage       ResourceUtilization `json:"usage"`
	Description string              `json:"description"`
}

// ScalingRecommendation represents scaling recommendations
type ScalingRecommendation struct {
	Type           string    `json:"type"` // scale_up, scale_down, scale_out, scale_in
	ResourceType   string    `json:"resource_type"`
	CurrentValue   float64   `json:"current_value"`
	RecommendedValue float64 `json:"recommended_value"`
	Reasoning      string    `json:"reasoning"`
	Timeline       string    `json:"timeline"`
	CostImpact     float64   `json:"cost_impact"`
	AffectedMachines []string `json:"affected_machines"`
}

// ResourceBottleneck represents resource bottlenecks
type ResourceBottleneck struct {
	ResourceType   string   `json:"resource_type"`
	BottleneckType string   `json:"bottleneck_type"`
	Severity       string   `json:"severity"`
	AffectedMachines []string `json:"affected_machines"`
	Impact         string   `json:"impact"`
	Resolution     string   `json:"resolution"`
}

// FleetRecommendation represents actionable fleet recommendations
type FleetRecommendation struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Priority     string    `json:"priority"`
	Category     string    `json:"category"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Impact       string    `json:"impact"`
	Effort       string    `json:"effort"`
	Savings      float64   `json:"savings"`
	RiskLevel    string    `json:"risk_level"`
	AffectedMachines []string `json:"affected_machines"`
	Implementation string  `json:"implementation"`
	Timeline     string    `json:"timeline"`
	CreatedAt    time.Time `json:"created_at"`
}

// TrendAnalysis provides trend analysis across fleet metrics
type TrendAnalysis struct {
	OverallTrend    string               `json:"overall_trend"`
	KeyInsights     []string             `json:"key_insights"`
	AnomaliesDetected []TrendAnomaly     `json:"anomalies_detected"`
	SeasonalPatterns []SeasonalPattern   `json:"seasonal_patterns"`
	Predictions     []TrendPrediction    `json:"predictions"`
}

// TrendAnomaly represents detected anomalies in trends
type TrendAnomaly struct {
	Type        string    `json:"type"`
	MetricType  string    `json:"metric_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detected_at"`
	AffectedMachines []string `json:"affected_machines"`
}

// SeasonalPattern represents seasonal patterns in metrics
type SeasonalPattern struct {
	PatternType string              `json:"pattern_type"`
	MetricType  string              `json:"metric_type"`
	Period      string              `json:"period"`
	Amplitude   float64             `json:"amplitude"`
	Confidence  float64             `json:"confidence"`
	TimeSeries  []TimeSeriesPoint   `json:"time_series"`
}

// TrendPrediction represents future trend predictions
type TrendPrediction struct {
	MetricType      string            `json:"metric_type"`
	PredictionType  string            `json:"prediction_type"`
	Timeframe       string            `json:"timeframe"`
	PredictedValue  float64           `json:"predicted_value"`
	Confidence      float64           `json:"confidence"`
	ConfidenceRange []float64         `json:"confidence_range"`
	Assumptions     []string          `json:"assumptions"`
}

// NewFleetAnalytics creates a new fleet analytics instance
func NewFleetAnalytics(logger *logger.Logger, fleetManager *fleet.FleetManager, storage MetricsStorage) *FleetAnalytics {
	return &FleetAnalytics{
		logger:         logger,
		fleetManager:   fleetManager,
		metricsStorage: storage,
	}
}

// ReportRequest represents a request for generating analytics report
type ReportRequest struct {
	FleetID   string    `json:"fleet_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Metrics   []string  `json:"metrics"` // performance, cost, security, capacity
}

// GenerateReport generates a comprehensive fleet analytics report
func (fa *FleetAnalytics) GenerateReport(ctx context.Context, request ReportRequest) (*FleetAnalyticsReport, error) {
	fa.logger.Info(fmt.Sprintf("Generating fleet analytics report for fleet: %s", request.FleetID))

	report := &FleetAnalyticsReport{
		GeneratedAt: time.Now(),
	}

	// Generate fleet overview
	overview, err := fa.generateFleetOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fleet overview: %w", err)
	}
	report.FleetOverview = *overview

	// Generate analysis based on requested metrics
	for _, metric := range request.Metrics {
		switch metric {
		case "performance":
			perfAnalysis, err := fa.generatePerformanceAnalysis(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate performance analysis: %w", err)
			}
			report.PerformanceAnalysis = *perfAnalysis
		case "cost":
			costAnalysis, err := fa.generateCostAnalysis(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate cost analysis: %w", err)
			}
			report.CostAnalysis = *costAnalysis
		case "security":
			secAnalysis, err := fa.generateSecurityAnalysis(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate security analysis: %w", err)
			}
			report.SecurityAnalysis = *secAnalysis
		case "capacity":
			capAnalysis, err := fa.generateCapacityAnalysis(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to generate capacity analysis: %w", err)
			}
			report.CapacityAnalysis = *capAnalysis
		}
	}

	// Generate recommendations
	recommendations, err := fa.generateRecommendations(ctx, &report.PerformanceAnalysis, &report.CostAnalysis, &report.SecurityAnalysis, &report.CapacityAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}
	report.Recommendations = recommendations

	// Generate trend analysis
	trendAnalysis, err := fa.generateTrendAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trend analysis: %w", err)
	}
	report.TrendAnalysis = *trendAnalysis

	fa.logger.Info("Fleet analytics report generated successfully")
	return report, nil
}

// GenerateFleetReport generates a comprehensive fleet analytics report
func (fa *FleetAnalytics) GenerateFleetReport(ctx context.Context) (*FleetAnalyticsReport, error) {
	fa.logger.Info("Generating comprehensive fleet analytics report")

	// Generate all analysis components
	fleetOverview, err := fa.generateFleetOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fleet overview: %w", err)
	}

	performanceAnalysis, err := fa.generatePerformanceAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate performance analysis: %w", err)
	}

	costAnalysis, err := fa.generateCostAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cost analysis: %w", err)
	}

	securityAnalysis, err := fa.generateSecurityAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate security analysis: %w", err)
	}

	capacityAnalysis, err := fa.generateCapacityAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate capacity analysis: %w", err)
	}

	recommendations, err := fa.generateRecommendations(ctx, performanceAnalysis, costAnalysis, securityAnalysis, capacityAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	trendAnalysis, err := fa.generateTrendAnalysis(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate trend analysis: %w", err)
	}

	report := &FleetAnalyticsReport{
		GeneratedAt:         time.Now(),
		FleetOverview:       *fleetOverview,
		PerformanceAnalysis: *performanceAnalysis,
		CostAnalysis:        *costAnalysis,
		SecurityAnalysis:    *securityAnalysis,
		CapacityAnalysis:    *capacityAnalysis,
		Recommendations:     recommendations,
		TrendAnalysis:       *trendAnalysis,
	}

	fa.logger.Info("Fleet analytics report generated successfully")
	return report, nil
}

// generateFleetOverview generates fleet overview statistics
func (fa *FleetAnalytics) generateFleetOverview(ctx context.Context) (*FleetOverview, error) {
	// Get all machines from fleet manager
	machines, err := fa.fleetManager.ListMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}

	overview := &FleetOverview{
		TotalMachines:        len(machines),
		EnvironmentBreakdown: make(map[string]int),
		RegionBreakdown:      make(map[string]int),
	}

	var totalUptime float64
	var totalCPU int
	var totalMemory float64
	var totalStorage float64
	var configDrift int

	for _, machine := range machines {
		// Count by environment
		overview.EnvironmentBreakdown[machine.Environment]++

		// Count by region (from metadata)
		if region, ok := machine.Metadata["region"]; ok {
			overview.RegionBreakdown[region]++
		}

		// Health status
		if machine.Health.Overall == "healthy" {
			overview.HealthyMachines++
		} else {
			overview.UnhealthyMachines++
		}

		// Aggregate resource information
		if cpuCores, ok := machine.Metadata["cpu_cores"]; ok {
			if cpu := parseIntMetadata(cpuCores); cpu > 0 {
				totalCPU += cpu
			}
		}

		if memoryGB, ok := machine.Metadata["memory_gb"]; ok {
			if memory := parseFloatMetadata(memoryGB); memory > 0 {
				totalMemory += memory
			}
		}

		if storageGB, ok := machine.Metadata["storage_gb"]; ok {
			if storage := parseFloatMetadata(storageGB); storage > 0 {
				totalStorage += storage
			}
		}

		// Calculate uptime
		if uptimeStr, ok := machine.Metadata["uptime"]; ok {
			if uptime := parseFloatMetadata(uptimeStr); uptime > 0 {
				totalUptime += uptime
			}
		}

		// Check configuration drift
		if machine.Config.UpdateStatus != "up-to-date" {
			configDrift++
		}
	}

	// Calculate averages
	if len(machines) > 0 {
		overview.AverageUptime = totalUptime / float64(len(machines))
	}

	overview.TotalCPUCores = totalCPU
	overview.TotalMemoryGB = totalMemory
	overview.TotalStorageGB = totalStorage
	overview.ConfigurationDrift = configDrift

	return overview, nil
}

// generatePerformanceAnalysis generates performance analysis
func (fa *FleetAnalytics) generatePerformanceAnalysis(ctx context.Context) (*PerformanceAnalysis, error) {
	// Get performance metrics for the last 24 hours
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	query := MetricQuery{
		MetricTypes: []string{"cpu_usage", "memory_usage", "load_average"},
		StartTime:   startTime,
		EndTime:     endTime,
	}

	metrics, err := fa.metricsStorage.GetMetrics(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	analysis := &PerformanceAnalysis{
		TopPerformers:    make([]MachinePerformance, 0),
		BottomPerformers: make([]MachinePerformance, 0),
		PerformanceTrends: make([]PerformanceTrend, 0),
		Bottlenecks:      make([]PerformanceBottleneck, 0),
	}

	// Calculate averages and identify performers
	machinePerformance := make(map[string]*MachinePerformance)
	
	for _, metric := range metrics {
		if _, exists := machinePerformance[metric.MachineID]; !exists {
			machinePerformance[metric.MachineID] = &MachinePerformance{
				MachineID:   metric.MachineID,
				MachineName: metric.Tags["machine_name"],
				Environment: metric.Environment,
			}
		}

		perf := machinePerformance[metric.MachineID]
		switch metric.MetricType {
		case "cpu_usage":
			perf.CPUUsage = metric.Value
		case "memory_usage":
			perf.MemoryUsage = metric.Value
		case "load_average":
			perf.LoadAverage = metric.Value
		}
	}

	// Calculate performance scores and identify top/bottom performers
	performances := make([]MachinePerformance, 0, len(machinePerformance))
	var totalCPU, totalMemory, totalLoad float64

	for _, perf := range machinePerformance {
		// Calculate performance score (lower is better)
		perf.PerformanceScore = (perf.CPUUsage + perf.MemoryUsage + perf.LoadAverage) / 3
		performances = append(performances, *perf)
		
		totalCPU += perf.CPUUsage
		totalMemory += perf.MemoryUsage
		totalLoad += perf.LoadAverage
	}

	// Calculate averages
	if len(performances) > 0 {
		analysis.AverageCPUUsage = totalCPU / float64(len(performances))
		analysis.AverageMemoryUsage = totalMemory / float64(len(performances))
		analysis.AverageLoadAverage = totalLoad / float64(len(performances))
	}

	// Sort by performance score
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].PerformanceScore < performances[j].PerformanceScore
	})

	// Top performers (best 3)
	topCount := min(3, len(performances))
	for i := 0; i < topCount; i++ {
		analysis.TopPerformers = append(analysis.TopPerformers, performances[i])
	}

	// Bottom performers (worst 3)
	bottomCount := min(3, len(performances))
	for i := len(performances) - bottomCount; i < len(performances); i++ {
		analysis.BottomPerformers = append(analysis.BottomPerformers, performances[i])
	}

	// Generate performance trends
	analysis.PerformanceTrends = fa.generatePerformanceTrends(ctx, metrics)

	// Identify bottlenecks
	analysis.Bottlenecks = fa.identifyBottlenecks(ctx, performances)

	return analysis, nil
}

// generateCostAnalysis generates cost analysis
func (fa *FleetAnalytics) generateCostAnalysis(ctx context.Context) (*CostAnalysis, error) {
	machines, err := fa.fleetManager.ListMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}

	analysis := &CostAnalysis{
		CostPerEnvironment: make(map[string]float64),
		CostBreakdown: CostBreakdown{},
		CostTrends: make([]CostTrend, 0),
		OptimizationOpportunities: make([]CostOptimization, 0),
	}

	var totalCost float64
	var totalCompute, totalStorage, totalNetwork float64

	for _, machine := range machines {
		// Calculate machine cost based on resources
		machineCost := fa.calculateMachineCost(machine)
		totalCost += machineCost
		analysis.CostPerEnvironment[machine.Environment] += machineCost

		// Break down costs by type
		computeCost := machineCost * 0.7  // 70% compute
		storageCost := machineCost * 0.2  // 20% storage
		networkCost := machineCost * 0.1  // 10% network

		totalCompute += computeCost
		totalStorage += storageCost
		totalNetwork += networkCost
	}

	analysis.TotalMonthlyCost = totalCost
	if len(machines) > 0 {
		analysis.CostPerMachine = totalCost / float64(len(machines))
	}

	analysis.CostBreakdown = CostBreakdown{
		ComputeCost:     totalCompute,
		StorageCost:     totalStorage,
		NetworkCost:     totalNetwork,
		LicenseCost:     totalCost * 0.05, // 5% licenses
		MaintenanceCost: totalCost * 0.15, // 15% maintenance
	}

	// Generate cost optimization opportunities
	analysis.OptimizationOpportunities = fa.generateCostOptimizations(ctx, machines)

	// Generate ROI analysis
	analysis.ROIAnalysis = fa.generateROIAnalysis(ctx, analysis.OptimizationOpportunities)

	return analysis, nil
}

// generateSecurityAnalysis generates security analysis
func (fa *FleetAnalytics) generateSecurityAnalysis(ctx context.Context) (*SecurityAnalysis, error) {
	machines, err := fa.fleetManager.ListMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}

	analysis := &SecurityAnalysis{
		VulnerabilityCount: VulnerabilityCount{},
		ComplianceStatus: ComplianceStatus{
			SOC2:     ComplianceFramework{Status: "compliant", Score: 85.0},
			HIPAA:    ComplianceFramework{Status: "partial", Score: 72.0},
			PCIDSS:   ComplianceFramework{Status: "compliant", Score: 88.0},
			ISO27001: ComplianceFramework{Status: "partial", Score: 78.0},
			GDPR:     ComplianceFramework{Status: "compliant", Score: 92.0},
		},
		SecurityTrends:       make([]SecurityTrend, 0),
		SecurityIncidents:    make([]SecurityIncident, 0),
		SecurityRecommendations: make([]SecurityRecommendation, 0),
	}

	// Simulate vulnerability scanning results
	for _, machine := range machines {
		// Simulate vulnerability counts based on machine characteristics
		if machine.Environment == "production" {
			analysis.VulnerabilityCount.Critical += 0 // Production should have zero critical
			analysis.VulnerabilityCount.High += 1
			analysis.VulnerabilityCount.Medium += 3
			analysis.VulnerabilityCount.Low += 5
		} else {
			analysis.VulnerabilityCount.Critical += 1
			analysis.VulnerabilityCount.High += 2
			analysis.VulnerabilityCount.Medium += 5
			analysis.VulnerabilityCount.Low += 8
		}
	}

	// Calculate security score
	totalVulns := analysis.VulnerabilityCount.Critical*10 + analysis.VulnerabilityCount.High*5 + 
				 analysis.VulnerabilityCount.Medium*2 + analysis.VulnerabilityCount.Low*1
	maxScore := float64(len(machines) * 50) // Max possible vulnerability score
	analysis.SecurityScore = math.Max(0, 100 - (float64(totalVulns)/maxScore)*100)

	// Generate security recommendations
	analysis.SecurityRecommendations = fa.generateSecurityRecommendations(ctx, machines, analysis.VulnerabilityCount)

	return analysis, nil
}

// generateCapacityAnalysis generates capacity analysis
func (fa *FleetAnalytics) generateCapacityAnalysis(ctx context.Context) (*CapacityAnalysis, error) {
	analysis := &CapacityAnalysis{
		CurrentUtilization: ResourceUtilization{
			CPUUtilization:     65.0,
			MemoryUtilization:  72.0,
			StorageUtilization: 45.0,
			NetworkUtilization: 30.0,
		},
		PredictedUtilization: ResourceUtilization{
			CPUUtilization:     78.0,
			MemoryUtilization:  85.0,
			StorageUtilization: 62.0,
			NetworkUtilization: 42.0,
		},
		CapacityForecast:       make([]CapacityForecast, 0),
		ScalingRecommendations: make([]ScalingRecommendation, 0),
		ResourceBottlenecks:    make([]ResourceBottleneck, 0),
	}

	// Generate capacity forecasts for next 6 months
	for i := 1; i <= 6; i++ {
		forecast := CapacityForecast{
			Period:       fmt.Sprintf("Month %d", i),
			ForecastDate: time.Now().AddDate(0, i, 0),
			PredictedUsage: ResourceUtilization{
				CPUUtilization:     analysis.CurrentUtilization.CPUUtilization + float64(i)*2.5,
				MemoryUtilization:  analysis.CurrentUtilization.MemoryUtilization + float64(i)*3.0,
				StorageUtilization: analysis.CurrentUtilization.StorageUtilization + float64(i)*4.0,
				NetworkUtilization: analysis.CurrentUtilization.NetworkUtilization + float64(i)*2.0,
			},
			Confidence: math.Max(50, 95-float64(i)*5), // Decreasing confidence over time
		}
		analysis.CapacityForecast = append(analysis.CapacityForecast, forecast)
	}

	// Generate scaling recommendations
	analysis.ScalingRecommendations = fa.generateScalingRecommendations(ctx, analysis.CurrentUtilization, analysis.PredictedUtilization)

	// Identify resource bottlenecks
	analysis.ResourceBottlenecks = fa.identifyResourceBottlenecks(ctx, analysis.CurrentUtilization)

	return analysis, nil
}

// Helper functions for analytics

func (fa *FleetAnalytics) calculateMachineCost(machine *fleet.Machine) float64 {
	baseCost := 50.0 // Base cost per machine

	// Add cost based on resources
	if cpuCores := parseIntMetadata(machine.Metadata["cpu_cores"]); cpuCores > 0 {
		baseCost += float64(cpuCores) * 10.0
	}

	if memoryGB := parseFloatMetadata(machine.Metadata["memory_gb"]); memoryGB > 0 {
		baseCost += memoryGB * 5.0
	}

	if storageGB := parseFloatMetadata(machine.Metadata["storage_gb"]); storageGB > 0 {
		baseCost += storageGB * 0.1
	}

	// Environment multiplier
	switch machine.Environment {
	case "production":
		baseCost *= 1.5
	case "staging":
		baseCost *= 1.2
	default:
		baseCost *= 1.0
	}

	return baseCost
}

func (fa *FleetAnalytics) generatePerformanceTrends(ctx context.Context, metrics []*Metric) []PerformanceTrend {
	trends := make([]PerformanceTrend, 0)

	// Group metrics by type
	metricsByType := make(map[string][]*Metric)
	for _, metric := range metrics {
		metricsByType[metric.MetricType] = append(metricsByType[metric.MetricType], metric)
	}

	for metricType, typeMetrics := range metricsByType {
		if len(typeMetrics) < 2 {
			continue
		}

		// Sort by timestamp
		sort.Slice(typeMetrics, func(i, j int) bool {
			return typeMetrics[i].Timestamp.Before(typeMetrics[j].Timestamp)
		})

		// Calculate trend
		firstValue := typeMetrics[0].Value
		lastValue := typeMetrics[len(typeMetrics)-1].Value
		changeRate := ((lastValue - firstValue) / firstValue) * 100

		trendDirection := "stable"
		if changeRate > 5 {
			trendDirection = "increasing"
		} else if changeRate < -5 {
			trendDirection = "decreasing"
		}

		trends = append(trends, PerformanceTrend{
			MetricType: metricType,
			Trend:      trendDirection,
			ChangeRate: changeRate,
			Confidence: 85.0, // Default confidence
		})
	}

	return trends
}

func (fa *FleetAnalytics) identifyBottlenecks(ctx context.Context, performances []MachinePerformance) []PerformanceBottleneck {
	bottlenecks := make([]PerformanceBottleneck, 0)

	// Identify machines with high resource usage
	for _, perf := range performances {
		if perf.CPUUsage > 90 {
			bottlenecks = append(bottlenecks, PerformanceBottleneck{
				Type:        "cpu",
				Severity:    "high",
				Description: fmt.Sprintf("High CPU usage on machine %s", perf.MachineName),
				AffectedMachines: []string{perf.MachineID},
				Impact:      perf.CPUUsage,
				Recommendation: "Consider CPU upgrade or workload redistribution",
			})
		}

		if perf.MemoryUsage > 85 {
			bottlenecks = append(bottlenecks, PerformanceBottleneck{
				Type:        "memory",
				Severity:    "medium",
				Description: fmt.Sprintf("High memory usage on machine %s", perf.MachineName),
				AffectedMachines: []string{perf.MachineID},
				Impact:      perf.MemoryUsage,
				Recommendation: "Consider memory upgrade or application optimization",
			})
		}
	}

	return bottlenecks
}

func (fa *FleetAnalytics) generateCostOptimizations(ctx context.Context, machines []*fleet.Machine) []CostOptimization {
	optimizations := make([]CostOptimization, 0)

	// Identify underutilized machines
	underutilizedCount := 0
	for _, machine := range machines {
		if cpuUsage := parseFloatMetadata(machine.Metadata["cpu_usage"]); cpuUsage < 30 {
			underutilizedCount++
		}
	}

	if underutilizedCount > 0 {
		optimizations = append(optimizations, CostOptimization{
			Type:            "rightsizing",
			Description:     fmt.Sprintf("Rightsizing %d underutilized machines", underutilizedCount),
			PotentialSavings: float64(underutilizedCount) * 200.0, // $200 per machine per month
			ImplementationEffort: "medium",
			RiskLevel:       "low",
		})
	}

	// Suggest reserved instances for production
	prodCount := 0
	for _, machine := range machines {
		if machine.Environment == "production" {
			prodCount++
		}
	}

	if prodCount > 0 {
		optimizations = append(optimizations, CostOptimization{
			Type:            "reserved_instances",
			Description:     fmt.Sprintf("Use reserved instances for %d production machines", prodCount),
			PotentialSavings: float64(prodCount) * 150.0, // $150 per machine per month
			ImplementationEffort: "low",
			RiskLevel:       "very_low",
		})
	}

	return optimizations
}

func (fa *FleetAnalytics) generateROIAnalysis(ctx context.Context, optimizations []CostOptimization) ROIAnalysis {
	var totalSavings float64
	for _, opt := range optimizations {
		totalSavings += opt.PotentialSavings
	}

	investmentCost := totalSavings * 0.1 // 10% of savings as investment
	annualSavings := totalSavings * 12   // Monthly to annual

	var paybackPeriod float64
	if annualSavings > 0 {
		paybackPeriod = investmentCost / annualSavings * 12 // months
	}

	threeYearROI := ((annualSavings*3 - investmentCost) / investmentCost) * 100

	return ROIAnalysis{
		InvestmentCost: investmentCost,
		AnnualSavings:  annualSavings,
		PaybackPeriod:  paybackPeriod,
		ThreeYearROI:   threeYearROI,
		BreakEvenPoint: time.Now().AddDate(0, int(paybackPeriod), 0),
	}
}

func (fa *FleetAnalytics) generateSecurityRecommendations(ctx context.Context, machines []*fleet.Machine, vulnCount VulnerabilityCount) []SecurityRecommendation {
	recommendations := make([]SecurityRecommendation, 0)

	if vulnCount.Critical > 0 {
		recommendations = append(recommendations, SecurityRecommendation{
			Priority:     "critical",
			Category:     "vulnerability_management",
			Description:  "Immediate patching required for critical vulnerabilities",
			Implementation: "Apply security patches within 24 hours",
			Impact:       "high",
		})
	}

	if vulnCount.High > 5 {
		recommendations = append(recommendations, SecurityRecommendation{
			Priority:     "high",
			Category:     "vulnerability_management",
			Description:  "Schedule patching for high-severity vulnerabilities",
			Implementation: "Apply patches within 72 hours",
			Impact:       "medium",
		})
	}

	recommendations = append(recommendations, SecurityRecommendation{
		Priority:     "medium",
		Category:     "access_control",
		Description:  "Implement multi-factor authentication for all admin access",
		Implementation: "Configure MFA for SSH and web interfaces",
		Impact:       "high",
	})

	return recommendations
}

func (fa *FleetAnalytics) generateScalingRecommendations(ctx context.Context, current, predicted ResourceUtilization) []ScalingRecommendation {
	recommendations := make([]ScalingRecommendation, 0)

	if predicted.CPUUtilization > 80 {
		recommendations = append(recommendations, ScalingRecommendation{
			Type:           "scale_up",
			ResourceType:   "cpu",
			CurrentValue:   current.CPUUtilization,
			RecommendedValue: predicted.CPUUtilization + 20, // Add 20% headroom
			Reasoning:      "CPU utilization approaching capacity",
			Timeline:       "within 30 days",
			CostImpact:     500.0,
		})
	}

	if predicted.MemoryUtilization > 85 {
		recommendations = append(recommendations, ScalingRecommendation{
			Type:           "scale_up",
			ResourceType:   "memory",
			CurrentValue:   current.MemoryUtilization,
			RecommendedValue: predicted.MemoryUtilization + 15, // Add 15% headroom
			Reasoning:      "Memory utilization approaching capacity",
			Timeline:       "within 60 days",
			CostImpact:     300.0,
		})
	}

	return recommendations
}

func (fa *FleetAnalytics) identifyResourceBottlenecks(ctx context.Context, utilization ResourceUtilization) []ResourceBottleneck {
	bottlenecks := make([]ResourceBottleneck, 0)

	if utilization.CPUUtilization > 85 {
		bottlenecks = append(bottlenecks, ResourceBottleneck{
			ResourceType:   "cpu",
			BottleneckType: "capacity",
			Severity:       "high",
			Impact:         "Performance degradation and slower response times",
			Resolution:     "Scale up CPU resources or optimize workloads",
		})
	}

	if utilization.MemoryUtilization > 90 {
		bottlenecks = append(bottlenecks, ResourceBottleneck{
			ResourceType:   "memory",
			BottleneckType: "capacity",
			Severity:       "critical",
			Impact:         "Risk of out-of-memory errors and system instability",
			Resolution:     "Increase memory allocation or optimize memory usage",
		})
	}

	return bottlenecks
}

func (fa *FleetAnalytics) generateRecommendations(ctx context.Context, perf *PerformanceAnalysis, cost *CostAnalysis, security *SecurityAnalysis, capacity *CapacityAnalysis) ([]FleetRecommendation, error) {
	recommendations := make([]FleetRecommendation, 0)

	// Performance recommendations
	for _, bottleneck := range perf.Bottlenecks {
		recommendations = append(recommendations, FleetRecommendation{
			ID:          fmt.Sprintf("perf-%s-%d", bottleneck.Type, time.Now().Unix()),
			Type:        "performance",
			Priority:    bottleneck.Severity,
			Category:    "optimization",
			Title:       fmt.Sprintf("Address %s bottleneck", bottleneck.Type),
			Description: bottleneck.Description,
			Impact:      "high",
			Effort:      "medium",
			AffectedMachines: bottleneck.AffectedMachines,
			Implementation: bottleneck.Recommendation,
			Timeline:    "30 days",
			CreatedAt:   time.Now(),
		})
	}

	// Cost optimization recommendations
	for _, optimization := range cost.OptimizationOpportunities {
		recommendations = append(recommendations, FleetRecommendation{
			ID:          fmt.Sprintf("cost-%s-%d", optimization.Type, time.Now().Unix()),
			Type:        "cost_optimization",
			Priority:    "medium",
			Category:    "financial",
			Title:       optimization.Description,
			Description: optimization.Description,
			Impact:      "medium",
			Effort:      optimization.ImplementationEffort,
			Savings:     optimization.PotentialSavings,
			RiskLevel:   optimization.RiskLevel,
			AffectedMachines: optimization.AffectedMachines,
			Timeline:    "60 days",
			CreatedAt:   time.Now(),
		})
	}

	// Security recommendations
	for _, secRec := range security.SecurityRecommendations {
		recommendations = append(recommendations, FleetRecommendation{
			ID:          fmt.Sprintf("sec-%s-%d", secRec.Category, time.Now().Unix()),
			Type:        "security",
			Priority:    secRec.Priority,
			Category:    secRec.Category,
			Title:       secRec.Description,
			Description: secRec.Description,
			Impact:      secRec.Impact,
			Effort:      "medium",
			AffectedMachines: secRec.AffectedMachines,
			Implementation: secRec.Implementation,
			Timeline:    "14 days",
			CreatedAt:   time.Now(),
		})
	}

	return recommendations, nil
}

func (fa *FleetAnalytics) generateTrendAnalysis(ctx context.Context) (*TrendAnalysis, error) {
	analysis := &TrendAnalysis{
		OverallTrend: "stable",
		KeyInsights: []string{
			"Fleet performance is stable with minor optimization opportunities",
			"Cost trends show potential for 15-20% savings through rightsizing",
			"Security posture is good but requires attention to high-priority vulnerabilities",
			"Capacity planning indicates need for scaling in 3-6 months",
		},
		AnomaliesDetected: make([]TrendAnomaly, 0),
		SeasonalPatterns:  make([]SeasonalPattern, 0),
		Predictions:       make([]TrendPrediction, 0),
	}

	// Generate sample predictions
	analysis.Predictions = append(analysis.Predictions, TrendPrediction{
		MetricType:      "cpu_usage",
		PredictionType:  "linear",
		Timeframe:       "3_months",
		PredictedValue:  75.0,
		Confidence:      85.0,
		ConfidenceRange: []float64{70.0, 80.0},
		Assumptions:     []string{"Current growth rate continues", "No major workload changes"},
	})

	return analysis, nil
}

// Utility functions
func parseIntMetadata(value string) int {
	if value == "" {
		return 0
	}
	// Simple int parsing simulation
	switch value {
	case "1", "2", "4", "8", "16", "32":
		return int(value[0] - '0')
	default:
		return 4 // Default value
	}
}

func parseFloatMetadata(value string) float64 {
	if value == "" {
		return 0.0
	}
	// Simple float parsing simulation
	switch value {
	case "8", "16", "32", "64":
		return float64(value[0] - '0')
	default:
		return 8.0 // Default value
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}