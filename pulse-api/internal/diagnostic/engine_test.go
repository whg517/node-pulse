package diagnostic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiagnosticEngine_Diagnose_NodeLocalFailure(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case: One node in us-east has high latency, others in same region are normal
	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        200.0, // High latency (4x baseline)
			PacketLossRate: 0.05,
			Jitter:         5.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        50.0, // Normal latency
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node3",
			Region:         "us-east",
			Latency:        55.0, // Normal latency
			PacketLossRate: 0.01,
			Jitter:         2.5,
			DataPointCount: 60,
		},
		{
			NodeID:         "node4",
			Region:         "eu-west",
			Latency:        52.0, // Normal latency
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
	}

	result, err := engine.Diagnose(nodesData)

	assert.NoError(t, err)
	assert.Equal(t, ProblemTypeNodeLocalFailure, result.ProblemType)
	assert.Equal(t, ConfidenceHigh, result.Confidence)
	assert.Contains(t, result.Analysis.AffectedNodes, "node1")
	assert.Equal(t, 4, result.Analysis.NodesAnalyzed)
	assert.NotEmpty(t, result.Recommendation)
}

func TestDiagnosticEngine_Diagnose_CrossBorderLink(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case: All nodes in eu-west have high latency, us-east nodes are normal
	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        50.0, // Normal
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        55.0, // Normal
			PacketLossRate: 0.01,
			Jitter:         2.5,
			DataPointCount: 60,
		},
		{
			NodeID:         "node3",
			Region:         "eu-west",
			Latency:        180.0, // High latency
			PacketLossRate: 0.05,
			Jitter:         8.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node4",
			Region:         "eu-west",
			Latency:        190.0, // High latency
			PacketLossRate: 0.06,
			Jitter:         9.0,
			DataPointCount: 60,
		},
	}

	result, err := engine.Diagnose(nodesData)

	assert.NoError(t, err)
	assert.Equal(t, ProblemTypeCrossBorderLink, result.ProblemType)
	assert.Equal(t, ConfidenceHigh, result.Confidence)
	assert.Contains(t, result.Analysis.AffectedNodes, "node3")
	assert.Contains(t, result.Analysis.AffectedNodes, "node4")
	assert.NotEmpty(t, result.Recommendation)
}

func TestDiagnosticEngine_Diagnose_ISPRouting(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case: Multiple regions have similar high latency patterns
	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        150.0, // High
			PacketLossRate: 0.04,
			Jitter:         6.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        160.0, // High
			PacketLossRate: 0.05,
			Jitter:         7.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node3",
			Region:         "eu-west",
			Latency:        155.0, // High
			PacketLossRate: 0.045,
			Jitter:         6.5,
			DataPointCount: 60,
		},
		{
			NodeID:         "node4",
			Region:         "eu-west",
			Latency:        165.0, // High
			PacketLossRate: 0.055,
			Jitter:         7.5,
			DataPointCount: 60,
		},
		{
			NodeID:         "node5",
			Region:         "ap-southeast",
			Latency:        52.0, // Normal
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
	}

	result, err := engine.Diagnose(nodesData)

	assert.NoError(t, err)
	assert.Equal(t, ProblemTypeISPRouting, result.ProblemType)
	assert.Equal(t, ConfidenceMedium, result.Confidence)
	assert.NotEmpty(t, result.Recommendation)
}

func TestDiagnosticEngine_Diagnose_Unknown(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case: All nodes normal, no clear problem
	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        50.0,
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        55.0,
			PacketLossRate: 0.01,
			Jitter:         2.5,
			DataPointCount: 60,
		},
		{
			NodeID:         "node3",
			Region:         "eu-west",
			Latency:        52.0,
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
	}

	result, err := engine.Diagnose(nodesData)

	assert.NoError(t, err)
	assert.Equal(t, ProblemTypeUnknown, result.ProblemType)
	assert.NotEmpty(t, result.Recommendation)
}

func TestDiagnosticEngine_Diagnose_InsufficientNodes(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case: Only 2 nodes provided
	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        200.0,
			PacketLossRate: 0.05,
			Jitter:         5.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        50.0,
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
	}

	result, err := engine.Diagnose(nodesData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient nodes")
}

func TestDiagnosticEngine_AnalyzeByRegion(t *testing.T) {
	engine := NewDiagnosticEngine()

	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        50.0,
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        60.0,
			PacketLossRate: 0.015,
			Jitter:         2.5,
			DataPointCount: 60,
		},
		{
			NodeID:         "node3",
			Region:         "eu-west",
			Latency:        180.0,
			PacketLossRate: 0.05,
			Jitter:         8.0,
			DataPointCount: 60,
		},
	}

	analysis := engine.analyzeByRegion(nodesData)

	assert.Len(t, analysis, 2)

	// Find us-east analysis
	var usEast, euWest *RegionalAnalysis
	for i := range analysis {
		if analysis[i].Region == "us-east" {
			usEast = &analysis[i]
		} else if analysis[i].Region == "eu-west" {
			euWest = &analysis[i]
		}
	}

	assert.NotNil(t, usEast)
	assert.NotNil(t, euWest)

	// Check us-east stats
	assert.Equal(t, "us-east", usEast.Region)
	assert.Equal(t, 2, usEast.NodeCount)
	assert.InDelta(t, 55.0, usEast.AvgLatency, 0.1)
	assert.Equal(t, "normal", usEast.Status)

	// Check eu-west stats
	assert.Equal(t, "eu-west", euWest.Region)
	assert.Equal(t, 1, euWest.NodeCount)
	assert.InDelta(t, 180.0, euWest.AvgLatency, 0.1)
	assert.Equal(t, "abnormal", euWest.Status)
}

func TestDiagnosticEngine_CalculateOverallMetrics(t *testing.T) {
	engine := NewDiagnosticEngine()

	nodesData := []MetricData{
		{
			NodeID:         "node1",
			Region:         "us-east",
			Latency:        50.0,
			PacketLossRate: 0.01,
			Jitter:         2.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node2",
			Region:         "us-east",
			Latency:        100.0,
			PacketLossRate: 0.02,
			Jitter:         4.0,
			DataPointCount: 60,
		},
		{
			NodeID:         "node3",
			Region:         "eu-west",
			Latency:        150.0,
			PacketLossRate: 0.03,
			Jitter:         6.0,
			DataPointCount: 60,
		},
	}

	metrics := engine.calculateOverallMetrics(nodesData)

	// Check latency metrics
	assert.InDelta(t, 100.0, metrics.Latency.Avg, 0.1)
	assert.InDelta(t, 150.0, metrics.Latency.Max, 0.1)
	assert.InDelta(t, 50.0, metrics.Latency.Baseline, 0.1)

	// Check packet loss metrics
	assert.InDelta(t, 0.02, metrics.PacketLossRate.Avg, 0.001)
	assert.InDelta(t, 0.03, metrics.PacketLossRate.Max, 0.001)
	assert.InDelta(t, 0.01, metrics.PacketLossRate.Baseline, 0.001)

	// Check jitter metrics
	assert.InDelta(t, 4.0, metrics.Jitter.Avg, 0.1)
	assert.InDelta(t, 6.0, metrics.Jitter.Max, 0.1)
	assert.InDelta(t, 2.0, metrics.Jitter.Baseline, 0.1)
}

func TestDiagnosticEngine_GetUniqueRegions(t *testing.T) {
	engine := NewDiagnosticEngine()

	nodesData := []MetricData{
		{NodeID: "node1", Region: "us-east"},
		{NodeID: "node2", Region: "us-east"},
		{NodeID: "node3", Region: "eu-west"},
		{NodeID: "node4", Region: "ap-southeast"},
		{NodeID: "node5", Region: "eu-west"},
	}

	regions := engine.getUniqueRegions(nodesData)

	assert.Len(t, regions, 3)
	assert.Contains(t, regions, "us-east")
	assert.Contains(t, regions, "eu-west")
	assert.Contains(t, regions, "ap-southeast")
}

func TestDiagnosticEngine_CalculateVariance(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test with consistent values
	values1 := []float64{100.0, 105.0, 95.0, 100.0}
	variance1 := engine.calculateVariance(values1)
	assert.Less(t, variance1, 0.01) // Low variance

	// Test with variable values
	values2 := []float64{50.0, 150.0, 75.0, 125.0}
	variance2 := engine.calculateVariance(values2)
	assert.Greater(t, variance2, 0.05) // Higher variance
}

func TestDiagnosticEngine_GetRecommendation(t *testing.T) {
	engine := NewDiagnosticEngine()

	tests := []struct {
		problemType ProblemType
		confidence  ConfidenceLevel
		wantContains string
	}{
		{
			problemType:  ProblemTypeNodeLocalFailure,
			confidence:   ConfidenceHigh,
			wantContains: "node-local",
		},
		{
			problemType:  ProblemTypeCrossBorderLink,
			confidence:   ConfidenceHigh,
			wantContains: "cross-border",
		},
		{
			problemType:  ProblemTypeISPRouting,
			confidence:   ConfidenceMedium,
			wantContains: "ISP",
		},
		{
			problemType:  ProblemTypeUnknown,
			confidence:   ConfidenceLow,
			wantContains: "more data",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.problemType), func(t *testing.T) {
			rec := engine.getRecommendation(tt.problemType, tt.confidence)
			assert.Contains(t, rec, tt.wantContains)
		})
	}
}

func TestDiagnosticEngine_IsNodeLocalFailure(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case 1: Single abnormal node in region with multiple nodes
	nodesData1 := []MetricData{
		{NodeID: "node1", Region: "us-east", Latency: 200.0},
		{NodeID: "node2", Region: "us-east", Latency: 50.0},
		{NodeID: "node3", Region: "us-east", Latency: 55.0},
		{NodeID: "node4", Region: "eu-west", Latency: 52.0},
	}
	regionalAnalysis1 := engine.analyzeByRegion(nodesData1)
	assert.True(t, engine.isNodeLocalFailure(nodesData1, regionalAnalysis1))

	// Test case 2: Multiple abnormal nodes in each region (not local failure)
	nodesData2 := []MetricData{
		{NodeID: "node1", Region: "us-east", Latency: 200.0},
		{NodeID: "node2", Region: "us-east", Latency: 180.0},
		{NodeID: "node3", Region: "eu-west", Latency: 52.0},
	}
	regionalAnalysis2 := engine.analyzeByRegion(nodesData2)
	assert.False(t, engine.isNodeLocalFailure(nodesData2, regionalAnalysis2))
}

func TestDiagnosticEngine_IsISPRoutingIssue(t *testing.T) {
	engine := NewDiagnosticEngine()

	// Test case 1: Similar abnormal patterns across multiple regions
	nodesData1 := []MetricData{
		{NodeID: "node1", Region: "us-east", Latency: 150.0, PacketLossRate: 0.04, Jitter: 6.0},
		{NodeID: "node2", Region: "us-east", Latency: 160.0, PacketLossRate: 0.05, Jitter: 7.0},
		{NodeID: "node3", Region: "eu-west", Latency: 155.0, PacketLossRate: 0.045, Jitter: 6.5},
		{NodeID: "node4", Region: "eu-west", Latency: 165.0, PacketLossRate: 0.055, Jitter: 7.5},
	}
	regionalAnalysis1 := engine.analyzeByRegion(nodesData1)
	assert.True(t, engine.isISPRoutingIssue(nodesData1, regionalAnalysis1))

	// Test case 2: Only one region abnormal (not ISP issue)
	nodesData2 := []MetricData{
		{NodeID: "node1", Region: "us-east", Latency: 50.0, PacketLossRate: 0.01, Jitter: 2.0},
		{NodeID: "node2", Region: "eu-west", Latency: 180.0, PacketLossRate: 0.05, Jitter: 8.0},
		{NodeID: "node3", Region: "eu-west", Latency: 190.0, PacketLossRate: 0.06, Jitter: 9.0},
	}
	regionalAnalysis2 := engine.analyzeByRegion(nodesData2)
	assert.False(t, engine.isISPRoutingIssue(nodesData2, regionalAnalysis2))
}

func TestDiagnosticEngine_BuildRegionalComparison(t *testing.T) {
	engine := NewDiagnosticEngine()

	regionalAnalysis := []RegionalAnalysis{
		{
			Region:        "us-east",
			NodeCount:     2,
			AvgLatency:    55.0,
			AvgPacketLoss: 0.012,
			AvgJitter:     2.2,
			Status:        "normal",
		},
		{
			Region:        "eu-west",
			NodeCount:     2,
			AvgLatency:    185.0,
			AvgPacketLoss: 0.055,
			AvgJitter:     8.5,
			Status:        "abnormal",
		},
	}

	comparison := engine.buildRegionalComparison(regionalAnalysis)

	assert.Len(t, comparison, 2)
	assert.Contains(t, comparison, "us-east")
	assert.Contains(t, comparison, "eu-west")

	// Check us-east stats
	usEast := comparison["us-east"]
	assert.InDelta(t, 55.0, usEast.AvgLatency, 0.1)
	assert.Equal(t, "normal", usEast.Status)

	// Check eu-west stats
	euWest := comparison["eu-west"]
	assert.InDelta(t, 185.0, euWest.AvgLatency, 0.1)
	assert.Equal(t, "abnormal", euWest.Status)
}
