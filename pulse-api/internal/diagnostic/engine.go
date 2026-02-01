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

// NewDiagnosticEngine creates a new diagnostic engine
func NewDiagnosticEngine() *DiagnosticEngine {
	return &DiagnosticEngine{
		minNodes:           3,
		timeWindow:         1 * time.Hour,
		baselineLatency:    50.0,   // ms
		baselinePacketLoss: 0.01,   // 1%
		baselineJitter:     2.0,    // ms
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
		return ProblemTypeNodeLocalFailure, ConfidenceHigh, affectedNodes
	}

	// Case 2: Cross-Border Link Issue
	// All nodes in one or more regions abnormal, other regions normal
	if len(abnormalRegions) > 0 && len(normalRegions) > 0 {
		// Check if pattern is clear
		if len(abnormalRegions) == 1 {
			return ProblemTypeCrossBorderLink, ConfidenceHigh, affectedNodes
		}
		return ProblemTypeCrossBorderLink, ConfidenceMedium, affectedNodes
	}

	// Case 3: ISP Routing Issue
	// Multiple regions abnormal, widespread pattern
	if len(abnormalRegions) >= 2 {
		// Check if pattern suggests ISP issue
		if e.isISPRoutingIssue(nodesData, regionalAnalysis) {
			return ProblemTypeISPRouting, ConfidenceMedium, affectedNodes
		}
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
func (e *DiagnosticEngine) isISPRoutingIssue(
	nodesData []MetricData,
	regionalAnalysis []RegionalAnalysis,
) bool {
	// ISP issues typically show:
	// 1. Multiple regions affected
	// 2. Similar deviation patterns
	// 3. Not all nodes in each region affected

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
