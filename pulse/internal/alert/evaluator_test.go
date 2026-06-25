package alert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// TestEvaluateRule_Latency covers the latency threshold branch.
func TestEvaluateRule_Latency(t *testing.T) {
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameLatency, Threshold: 100, Level: "warning"}

	t.Run("exceeds threshold", func(t *testing.T) {
		data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}
		event := EvaluateRule(rule, data)
		require.NotNil(t, event)
		assert.Equal(t, "node-1", event.NodeID)
		assert.Equal(t, MetricNameLatency, event.Metric)
		assert.Equal(t, 100.0, event.Threshold)
		assert.Equal(t, 200.0, event.CurrentValue)
		assert.Equal(t, "warning", event.Level)
	})

	t.Run("within threshold", func(t *testing.T) {
		data := &MetricData{NodeID: "node-1", LatencyMs: 50, Timestamp: time.Now()}
		assert.Nil(t, EvaluateRule(rule, data))
	})
}

// TestEvaluateRule_PacketLoss covers the packet_loss_rate threshold branch.
func TestEvaluateRule_PacketLoss(t *testing.T) {
	rule := &models.Alert{ID: "rule-1", Metric: MetricNamePacketLoss, Threshold: 0.5, Level: "critical"}

	t.Run("exceeds threshold", func(t *testing.T) {
		data := &MetricData{NodeID: "node-1", PacketLossRate: 0.7, Timestamp: time.Now()}
		event := EvaluateRule(rule, data)
		require.NotNil(t, event)
		assert.Equal(t, 0.7, event.CurrentValue)
	})

	t.Run("within threshold", func(t *testing.T) {
		data := &MetricData{NodeID: "node-1", PacketLossRate: 0.1, Timestamp: time.Now()}
		assert.Nil(t, EvaluateRule(rule, data))
	})
}

// TestEvaluateRule_Jitter covers the jitter threshold branch.
func TestEvaluateRule_Jitter(t *testing.T) {
	rule := &models.Alert{ID: "rule-1", Metric: MetricNameJitter, Threshold: 10, Level: "warning"}

	t.Run("exceeds threshold", func(t *testing.T) {
		data := &MetricData{NodeID: "node-1", JitterMs: 25, Timestamp: time.Now()}
		event := EvaluateRule(rule, data)
		require.NotNil(t, event)
		assert.Equal(t, 25.0, event.CurrentValue)
	})

	t.Run("within threshold", func(t *testing.T) {
		data := &MetricData{NodeID: "node-1", JitterMs: 5, Timestamp: time.Now()}
		assert.Nil(t, EvaluateRule(rule, data))
	})
}

// TestEvaluateRule_UnknownMetric returns nil and does not panic.
func TestEvaluateRule_UnknownMetric(t *testing.T) {
	rule := &models.Alert{ID: "rule-1", Metric: "bandwidth", Threshold: 100, Level: "warning"}
	data := &MetricData{NodeID: "node-1", LatencyMs: 200, Timestamp: time.Now()}
	assert.Nil(t, EvaluateRule(rule, data))
}

// TestRuleAppliesToNode verifies global vs node-specific rule scoping.
func TestRuleAppliesToNode(t *testing.T) {
	node1 := "node-1"
	node2 := "node-2"

	t.Run("global rule applies to all nodes", func(t *testing.T) {
		rule := &models.Alert{ID: "global", NodeID: nil}
		assert.True(t, RuleAppliesToNode(rule, node1))
		assert.True(t, RuleAppliesToNode(rule, node2))
	})

	t.Run("node-specific rule applies only to its node", func(t *testing.T) {
		rule := &models.Alert{ID: "specific", NodeID: &node1}
		assert.True(t, RuleAppliesToNode(rule, node1))
		assert.False(t, RuleAppliesToNode(rule, node2))
	})
}
