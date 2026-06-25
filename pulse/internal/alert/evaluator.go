package alert

import (
	"log/slog"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// MetricNameLatency, MetricNamePacketLoss and MetricNameJitter are the metric
// identifiers recognised by the evaluator. They correspond to the fields on
// MetricData and to the metric column stored on alert rules.
const (
	MetricNameLatency    = "latency"
	MetricNamePacketLoss = "packet_loss_rate"
	MetricNameJitter     = "jitter"
)

// EvaluateRule is the pure domain logic that decides whether a metric sample
// breaches a rule's threshold. It has no I/O dependencies and returns a new
// *models.AlertEvent when the rule fires, or nil otherwise.
//
// The returned event is populated with the rule's threshold/level/metric and
// the offending sample value; CreatedAt is set to time.Now() to match the
// historical behaviour of the inline evaluator.
func EvaluateRule(rule *models.Alert, data *MetricData) *models.AlertEvent {
	var currentValue float64
	var exceedsThreshold bool

	switch rule.Metric {
	case MetricNameLatency:
		currentValue = data.LatencyMs
		exceedsThreshold = currentValue > rule.Threshold
	case MetricNamePacketLoss:
		currentValue = data.PacketLossRate
		exceedsThreshold = currentValue > rule.Threshold
	case MetricNameJitter:
		currentValue = data.JitterMs
		exceedsThreshold = currentValue > rule.Threshold
	default:
		slog.Warn("Unknown metric type in alert rule", "metric", rule.Metric)
		return nil
	}

	if !exceedsThreshold {
		return nil
	}

	return &models.AlertEvent{
		NodeID:       data.NodeID,
		Metric:       rule.Metric,
		Threshold:    rule.Threshold,
		CurrentValue: currentValue,
		Level:        rule.Level,
		CreatedAt:    time.Now(),
	}
}

// RuleAppliesToNode reports whether a rule is eligible for a given node.
// A global rule (NodeID == nil) applies to every node; a node-specific rule
// applies only when its NodeID matches.
func RuleAppliesToNode(rule *models.Alert, nodeID string) bool {
	if rule.NodeID == nil {
		return true
	}
	return *rule.NodeID == nodeID
}
