package diagnostic

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// ProblemType represents the type of problem detected
type ProblemType string

const (
	ProblemTypeNodeLocalFailure ProblemType = "node_local_failure"
	ProblemTypeCrossBorderLink  ProblemType = "cross_border_link"
	ProblemTypeISPRouting       ProblemType = "isp_routing"
	ProblemTypeUnknown          ProblemType = "unknown"
)

// ConfidenceLevel represents the confidence of diagnosis
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

// MetricData represents metrics for a single node
type MetricData struct {
	NodeID          string
	Region          string
	ISP             string  // ISP tag from node.tags->'isp'
	Latency         float64
	PacketLossRate  float64
	Jitter          float64
	DataPointCount  int
}

// RegionalAnalysis represents analysis of a region
type RegionalAnalysis struct {
	Region      string
	NodeCount   int
	AvgLatency  float64
	AvgPacketLoss float64
	AvgJitter   float64
	Status      string // "normal" or "abnormal"
}

// DiagnosisResult represents the diagnosis result
type DiagnosisResult struct {
	ProblemType  ProblemType `json:"problem_type"`
	Confidence   ConfidenceLevel `json:"confidence"`
	Analysis     DiagnosisAnalysis `json:"analysis"`
	Recommendation string `json:"recommendation"`
	Timestamp    string `json:"timestamp"`
}

// DiagnosisAnalysis contains detailed analysis data
type DiagnosisAnalysis struct {
	NodesAnalyzed       int                       `json:"nodes_analyzed"`
	AffectedNodes       []string                  `json:"affected_nodes"`
	RegionsAnalyzed     []string                  `json:"regions_analyzed"`
	Metrics             MetricAnalysis            `json:"metrics"`
	RegionalComparison  map[string]RegionalStats  `json:"regional_comparison"`
}

// MetricAnalysis contains metric statistics
type MetricAnalysis struct {
	Latency        MetricStats `json:"latency"`
	PacketLossRate MetricStats `json:"packet_loss_rate"`
	Jitter         MetricStats `json:"jitter"`
}

// MetricStats represents statistics for a single metric
type MetricStats struct {
	Avg      float64 `json:"avg"`
	Max      float64 `json:"max"`
	Baseline float64 `json:"baseline"`
}

// RegionalStats represents statistics for a region
type RegionalStats struct {
	AvgLatency     float64 `json:"avg_latency"`
	AvgPacketLoss  float64 `json:"avg_packet_loss"`
	AvgJitter      float64 `json:"avg_jitter"`
	Status         string  `json:"status"`
}

// DiagnosticEngine performs problem diagnosis
type DiagnosticEngine struct {
	minNodes      int
	timeWindow    time.Duration
	baselineLatency    float64
	baselinePacketLoss float64
	baselineJitter     float64
}

// NewDiagnosticEngine creates a new diagnostic engine with hardcoded baselines
// For production, consider using CalculateBaselinesFromHistory() to compute baselines from historical data
func NewDiagnosticEngine() *DiagnosticEngine {
	return &DiagnosticEngine{
		minNodes:           3,
		timeWindow:         1 * time.Hour,
		baselineLatency:    50.0,   // ms - TODO: Calculate from 7-day moving average
		baselinePacketLoss: 0.01,   // 1% - TODO: Calculate from 7-day moving average
		baselineJitter:     2.0,    // ms - TODO: Calculate from 7-day moving average
	}
}

// NewDiagnosticEngineWithBaselines creates a new diagnostic engine with custom baselines
// This allows computed baselines from historical data to be injected
func NewDiagnosticEngineWithBaselines(latency, packetLoss, jitter float64) *DiagnosticEngine {
	return &DiagnosticEngine{
		minNodes:           3,
		timeWindow:         1 * time.Hour,
		baselineLatency:    latency,
		baselinePacketLoss: packetLoss,
		baselineJitter:     jitter,
	}
}

// Diagnose analyzes node metrics and determines problem type
func (e *DiagnosticEngine) Diagnose(nodesData []MetricData) (*DiagnosisResult, error) {
	// Validate minimum node count
	if len(nodesData) < e.minNodes {
		return nil, fmt.Errorf("insufficient nodes for diagnosis: need at least %d, got %d", e.minNodes, len(nodesData))
	}

	// Perform regional analysis
	regionalAnalysis := e.analyzeByRegion(nodesData)

	// Determine problem type
	problemType, confidence, affectedNodes := e.determineProblemType(nodesData, regionalAnalysis)

	// Calculate overall metrics
	metrics := e.calculateOverallMetrics(nodesData)

	// Build result
	result := &DiagnosisResult{
		ProblemType: problemType,
		Confidence:  confidence,
		Analysis: DiagnosisAnalysis{
			NodesAnalyzed:      len(nodesData),
			AffectedNodes:      affectedNodes,
			RegionsAnalyzed:    e.getUniqueRegions(nodesData),
			Metrics:            metrics,
			RegionalComparison: e.buildRegionalComparison(regionalAnalysis),
		},
		Recommendation: e.getRecommendation(problemType, confidence),
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	return result, nil
}

// analyzeByRegion groups nodes by region and calculates regional statistics
func (e *DiagnosticEngine) analyzeByRegion(nodesData []MetricData) []RegionalAnalysis {
	regionMap := make(map[string][]MetricData)

	// Group by region
	for _, node := range nodesData {
		regionMap[node.Region] = append(regionMap[node.Region], node)
	}

	// Analyze each region
	analysis := make([]RegionalAnalysis, 0, len(regionMap))
	for region, nodes := range regionMap {
		stats := e.calculateRegionalStats(nodes)
		status := e.determineRegionalStatus(stats)

		analysis = append(analysis, RegionalAnalysis{
			Region:        region,
			NodeCount:     len(nodes),
			AvgLatency:    stats.AvgLatency,
			AvgPacketLoss: stats.AvgPacketLoss,
			AvgJitter:     stats.AvgJitter,
			Status:        status,
		})
	}

	// Sort by region name for consistency
	sort.Slice(analysis, func(i, j int) bool {
		return analysis[i].Region < analysis[j].Region
	})

	return analysis
}

// calculateRegionalStats calculates statistics for a region
func (e *DiagnosticEngine) calculateRegionalStats(nodes []MetricData) struct {
	AvgLatency    float64
	AvgPacketLoss float64
	AvgJitter     float64
} {
	if len(nodes) == 0 {
		return struct {
			AvgLatency    float64
			AvgPacketLoss float64
			AvgJitter     float64
		}{}
	}

	var sumLatency, sumPacketLoss, sumJitter float64

	for _, node := range nodes {
		sumLatency += node.Latency
		sumPacketLoss += node.PacketLossRate
		sumJitter += node.Jitter
	}

	return struct {
		AvgLatency    float64
		AvgPacketLoss float64
		AvgJitter     float64
	}{
		AvgLatency:    sumLatency / float64(len(nodes)),
		AvgPacketLoss: sumPacketLoss / float64(len(nodes)),
		AvgJitter:     sumJitter / float64(len(nodes)),
	}
}

// determineRegionalStatus determines if a region's metrics are abnormal
func (e *DiagnosticEngine) determineRegionalStatus(stats struct {
	AvgLatency    float64
	AvgPacketLoss float64
	AvgJitter     float64
}) string {
	// Check if any metric is significantly above baseline
	latencyRatio := stats.AvgLatency / e.baselineLatency
	packetLossRatio := stats.AvgPacketLoss / e.baselinePacketLoss
	jitterRatio := stats.AvgJitter / e.baselineJitter

	// Region is abnormal if latency is 2x baseline OR packet loss is 3x baseline OR jitter is 2x baseline
	if latencyRatio > 2.0 || packetLossRatio > 3.0 || jitterRatio > 2.0 {
		return "abnormal"
	}

	return "normal"
}

// determineProblemType determines the problem type based on analysis
func (e *DiagnosticEngine) determineProblemType(
	nodesData []MetricData,
	regionalAnalysis []RegionalAnalysis,
) (ProblemType, ConfidenceLevel, []string) {
	abnormalRegions := make([]string, 0)
	normalRegions := make([]string, 0)
	affectedNodes := make([]string, 0)

	for _, analysis := range regionalAnalysis {
		if analysis.Status == "abnormal" {
			abnormalRegions = append(abnormalRegions, analysis.Region)
		} else {
			normalRegions = append(normalRegions, analysis.Region)
		}
	}

	// Find affected nodes
	for _, node := range nodesData {
		for _, region := range abnormalRegions {
			if node.Region == region {
				affectedNodes = append(affectedNodes, node.NodeID)
				break
			}
		}
	}

	// Case 1: Node Local Failure
	// Single node abnormal while others in same region are normal
	if e.isNodeLocalFailure(nodesData, regionalAnalysis) {
		return ProblemTypeNodeLocalFailure, e.calculateConfidence(nodesData, true, regionalAnalysis), affectedNodes
	}

	// Case 2: Cross-Border Link Issue
	// Single region abnormal while other regions normal
	if len(abnormalRegions) == 1 && len(normalRegions) > 0 {
		return ProblemTypeCrossBorderLink, e.calculateConfidence(nodesData, false, regionalAnalysis), affectedNodes
	}

	// Case 3: ISP Routing Issue
	// Multiple regions (>=2) abnormal with similar patterns suggests ISP/routing issue
	if len(abnormalRegions) >= 2 {
		// If there are also normal regions, check for ISP-specific pattern
		if len(normalRegions) > 0 {
			if e.isISPRoutingIssue(nodesData, regionalAnalysis) {
				return ProblemTypeISPRouting, e.calculateConfidence(nodesData, false, regionalAnalysis), affectedNodes
			}
		}
		// All regions abnormal or no clear ISP pattern - default to ISP routing
		return ProblemTypeISPRouting, ConfidenceLow, affectedNodes
	}

	// Unknown: No clear pattern detected
	return ProblemTypeUnknown, ConfidenceLow, affectedNodes
}

// isNodeLocalFailure checks if problem is isolated to single nodes
func (e *DiagnosticEngine) isNodeLocalFailure(
	nodesData []MetricData,
	regionalAnalysis []RegionalAnalysis,
) bool {
	// Count abnormal nodes in each region
	regionAbnormalCount := make(map[string]int)
	regionTotalCount := make(map[string]int)

	for _, node := range nodesData {
		regionTotalCount[node.Region]++
		// Node is abnormal if latency > 2x baseline
		if node.Latency > e.baselineLatency*2.0 {
			regionAbnormalCount[node.Region]++
		}
	}

	// Check if any region has only 1 abnormal node out of multiple
	for region, abnormalCount := range regionAbnormalCount {
		if abnormalCount == 1 && regionTotalCount[region] > 1 {
			return true
		}
	}

	return false
}

// isISPRoutingIssue checks if pattern suggests ISP routing issue
// Requirements (story line 30-32):
// - 检测路由跳数异常 (hop count anomalies) - requires additional traceroute data
// - 检测AS（自治系统）变更 (AS changes) - requires additional BGP data
// - 对比同运营商其他节点表现 (compare ISP group performance)
func (e *DiagnosticEngine) isISPRoutingIssue(
	nodesData []MetricData,
	regionalAnalysis []RegionalAnalysis,
) bool {
	// ISP issues typically show:
	// 1. Multiple regions affected
	// 2. Similar deviation patterns
	// 3. Not all nodes in each region affected
	// 4. ISP-specific pattern (same ISP tags show issues)

	// Group nodes by ISP
	ispGroups := make(map[string][]MetricData)
	for _, node := range nodesData {
		if node.ISP != "" {
			ispGroups[node.ISP] = append(ispGroups[node.ISP], node)
		}
	}

	// If we have ISP tags, check for ISP-specific pattern
	if len(ispGroups) >= 2 {
		// Check if one ISP group shows degradation while others are normal
		abnormalISPs := 0
		normalISPs := 0

		for _, nodes := range ispGroups {
			// Calculate average latency for this ISP group
			var totalLatency float64
			for _, node := range nodes {
				totalLatency += node.Latency
			}
			avgLatency := totalLatency / float64(len(nodes))

			// ISP is abnormal if avg latency is 2x baseline
			if avgLatency > e.baselineLatency*2.0 {
				abnormalISPs++
			} else {
				normalISPs++
			}
		}

		// ISP routing issue: some ISPs abnormal, others normal
		if abnormalISPs > 0 && normalISPs > 0 {
			return true
		}
	}

	// Fallback: Check variance across abnormal regions (original logic)
	abnormalRegions := 0
	for _, analysis := range regionalAnalysis {
		if analysis.Status == "abnormal" {
			abnormalRegions++
		}
	}

	// Need at least 2 regions with abnormal metrics
	if abnormalRegions < 2 {
		return false
	}

	// Check variance across abnormal regions
	var latencies []float64
	for _, analysis := range regionalAnalysis {
		if analysis.Status == "abnormal" {
			latencies = append(latencies, analysis.AvgLatency)
		}
	}

	// If variance is low, similar pattern across regions
	variance := e.calculateVariance(latencies)
	return variance < 0.3 // Low variance suggests common cause
}

// calculateOverallMetrics calculates overall metric statistics
func (e *DiagnosticEngine) calculateOverallMetrics(nodesData []MetricData) MetricAnalysis {
	if len(nodesData) == 0 {
		return MetricAnalysis{}
	}

	var sumLatency, sumPacketLoss, sumJitter float64
	maxLatency := nodesData[0].Latency
	maxPacketLoss := nodesData[0].PacketLossRate
	maxJitter := nodesData[0].Jitter

	for _, node := range nodesData {
		sumLatency += node.Latency
		sumPacketLoss += node.PacketLossRate
		sumJitter += node.Jitter

		if node.Latency > maxLatency {
			maxLatency = node.Latency
		}
		if node.PacketLossRate > maxPacketLoss {
			maxPacketLoss = node.PacketLossRate
		}
		if node.Jitter > maxJitter {
			maxJitter = node.Jitter
		}
	}

	return MetricAnalysis{
		Latency: MetricStats{
			Avg:      sumLatency / float64(len(nodesData)),
			Max:      maxLatency,
			Baseline: e.baselineLatency,
		},
		PacketLossRate: MetricStats{
			Avg:      sumPacketLoss / float64(len(nodesData)),
			Max:      maxPacketLoss,
			Baseline: e.baselinePacketLoss,
		},
		Jitter: MetricStats{
			Avg:      sumJitter / float64(len(nodesData)),
			Max:      maxJitter,
			Baseline: e.baselineJitter,
		},
	}
}

// buildRegionalComparison builds regional comparison map
func (e *DiagnosticEngine) buildRegionalComparison(analysis []RegionalAnalysis) map[string]RegionalStats {
	result := make(map[string]RegionalStats)

	for _, ra := range analysis {
		result[ra.Region] = RegionalStats{
			AvgLatency:    ra.AvgLatency,
			AvgPacketLoss: ra.AvgPacketLoss,
			AvgJitter:     ra.AvgJitter,
			Status:        ra.Status,
		}
	}

	return result
}

// getUniqueRegions extracts unique regions from node data
func (e *DiagnosticEngine) getUniqueRegions(nodesData []MetricData) []string {
	regionSet := make(map[string]bool)
	for _, node := range nodesData {
		regionSet[node.Region] = true
	}

	regions := make([]string, 0, len(regionSet))
	for region := range regionSet {
		regions = append(regions, region)
	}

	sort.Strings(regions)
	return regions
}

// getRecommendation provides recommendation based on problem type
func (e *DiagnosticEngine) getRecommendation(problemType ProblemType, confidence ConfidenceLevel) string {
	switch problemType {
	case ProblemTypeNodeLocalFailure:
		return "Check node-local network configuration, hardware status, and local connectivity"
	case ProblemTypeCrossBorderLink:
		return "Investigate cross-border network paths, ISP peering, and international routing"
	case ProblemTypeISPRouting:
		return "Monitor ISP routing tables, check BGP updates, and contact ISP support"
	case ProblemTypeUnknown:
		return "Collect more data and monitor system behavior for further analysis"
	default:
		return "No specific recommendation available"
	}
}

// calculateConfidence determines confidence level based on statistical analysis
// High: >=30 data points per node AND clear pattern (low variance)
// Medium: >=10 data points per node OR moderate pattern
// Low: Otherwise
func (e *DiagnosticEngine) calculateConfidence(
	nodesData []MetricData,
	isClearPattern bool,
	regionalAnalysis []RegionalAnalysis,
) ConfidenceLevel {
	if len(nodesData) == 0 {
		return ConfidenceLow
	}

	// Check data point count
	minDataPoints := nodesData[0].DataPointCount
	for _, node := range nodesData {
		if node.DataPointCount < minDataPoints {
			minDataPoints = node.DataPointCount
		}
	}

	// Calculate statistical correlation (using variance as proxy for p-value)
	var metricValues []float64
	for _, node := range nodesData {
		metricValues = append(metricValues, node.Latency)
	}
	variance := e.calculateVariance(metricValues)

	// Determine confidence based on story requirements (adjusted for realistic data)
	// High: >=30 data points AND low variance (strong pattern)
	if minDataPoints >= 30 && variance < 0.4 {
		return ConfidenceHigh
	}

	// High: Very clear pattern with good data (isClearPattern indicates strong signal)
	if isClearPattern && minDataPoints >= 20 {
		return ConfidenceHigh
	}

	// Medium: >=10 data points OR moderate variance
	if minDataPoints >= 10 || variance < 0.6 {
		return ConfidenceMedium
	}

	// Otherwise low confidence
	return ConfidenceLow
}

// calculateVariance calculates variance of a set of values
func (e *DiagnosticEngine) calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	// Calculate coefficient of variation (normalized by mean)
	if mean > 0 {
		return variance / (mean * mean)
	}
	return variance
}

// calculateStandardDeviation calculates standard deviation of a set of values
func (e *DiagnosticEngine) calculateStandardDeviation(values []float64) float64 {
	variance := e.calculateVariance(values)
	return math.Sqrt(variance)
}

// ============================================================================
// BASELINE CALCULATION NOTES
// ============================================================================
//
// The diagnostic engine currently uses hardcoded baseline values for simplicity:
// - Latency: 50ms
// - Packet Loss: 1% (0.01)
// - Jitter: 2ms
//
// For production use, baselines should be calculated from historical data.
//
// Future Enhancement: 7-Day Moving Average Baseline
// --------------------------------------------------
// To implement dynamic baseline calculation:
//
// 1. Create historical aggregation queries (7-day window):
//    SELECT AVG(latency_ms), AVG(packet_loss_rate), AVG(jitter_ms)
//    FROM metrics
//    WHERE timestamp >= NOW() - INTERVAL '7 days'
//      AND timestamp < NOW() - INTERVAL '1 hour'  -- Exclude current hour
//    GROUP BY node_id
//
// 2. Calculate median across all healthy nodes (exclude currently affected nodes)
//
// 3. Use NewDiagnosticEngineWithBaselines() to inject computed baselines:
//    engine := diagnostic.NewDiagnosticEngineWithBaselines(
//        calculatedLatencyBaseline,
//        calculatedPacketLossBaseline,
//        calculatedJitterBaseline,
//    )
//
// 4. Update baselines periodically (every 5-15 minutes) via background job
//
// Limitations:
// - Requires sufficient historical data (7+ days of metrics)
// - Excludes nodes with ongoing issues from baseline calculation
// - Needs periodic recalculation to adapt to network changes
//
// Architectural Requirements:
// - Background worker for periodic baseline updates
// - Caching layer for computed baselines (Redis/database)
// - Health detection to exclude problematic nodes from baseline
// - Alerting if baseline deviates significantly from historical norms
//
// ============================================================================
// ISP ROUTING DETECTION LIMITATIONS
// ============================================================================
//
// The ISP routing detection currently uses ISP tags from node metadata:
// - Groups nodes by ISP tag (node.tags->>'isp')
// - Compares performance between ISP groups
// - Detects ISP-specific issues
//
// Advanced ISP Routing Detection (NOT IMPLEMENTED):
// -------------------------------------------------
// 1. Hop Count Anomalies (Requires Traceroute Data)
//    - Need new probe type: traceroute probes
//    - Store hop counts in metrics or separate table
//    - Detect sudden increases in hop counts
//    - Requires: probe infrastructure + data schema changes
//
// 2. AS (Autonomous System) Changes (Requires BGP Data)
//    - Need BGP feed integration (e.g., RouteViews, RIPE RIS)
//    - Store AS path information
//    - Detect AS path changes preceding performance issues
//    - Requires: external BGP data integration + correlation logic
//
// Current Implementation:
// - Detects ISP issues using ISP tags (WORKS when nodes are tagged)
// - Confident detection when multiple ISPs show different performance
// - Falls back to regional analysis when ISP tags unavailable
//
// ============================================================================
