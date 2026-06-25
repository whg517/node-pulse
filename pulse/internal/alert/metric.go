package alert

import "time"

// MetricData represents the per-node metric sample submitted by a beacon
// heartbeat for alert-rule evaluation.
type MetricData struct {
	NodeID         string
	LatencyMs      float64
	PacketLossRate float64
	JitterMs       float64
	Timestamp      time.Time
}

// EngineConfig defines configuration for the AlertEngine.
type EngineConfig struct {
	WorkerPoolSize           int
	MetricChannelBufferSize  int
	RuleCacheRefreshInterval time.Duration
}

// DefaultEngineConfig returns the default engine configuration.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		WorkerPoolSize:           10,
		MetricChannelBufferSize:  1000,
		RuleCacheRefreshInterval: 60 * time.Second,
	}
}
