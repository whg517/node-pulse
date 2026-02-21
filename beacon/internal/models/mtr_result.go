package models

import "time"

// MTRHop represents a single hop in an MTR trace
type MTRHop struct {
	HopNumber   int     `json:"hop_number"`
	IP          string  `json:"ip"`
	Hostname    string  `json:"hostname,omitempty"`
	ASNumber    string  `json:"as_number,omitempty"`
	Sent        int     `json:"sent"`
	Received    int     `json:"received"`
	LossRate    float64 `json:"loss_rate"`
	LastRTTMs   float64 `json:"last_rtt_ms"`
	AvgRTTMs    float64 `json:"avg_rtt_ms"`
	BestRTTMs   float64 `json:"best_rtt_ms"`
	WorstRTTMs  float64 `json:"worst_rtt_ms"`
	StdDevMs    float64 `json:"std_dev_ms"`
	Location    string  `json:"location,omitempty"` // City, Country
}

// MTRResult represents the complete result of an MTR probe
type MTRResult struct {
	Target       string    `json:"target"`
	TotalHops    int       `json:"total_hops"`
	Hops         []MTRHop  `json:"hops"`
	CompletedAt  time.Time `json:"completed_at"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// NewMTRResult creates a new MTR result
func NewMTRResult(target string, hops []MTRHop, success bool, errorMessage string) *MTRResult {
	return &MTRResult{
		Target:       target,
		TotalHops:    len(hops),
		Hops:         hops,
		CompletedAt:  time.Now(),
		Success:      success,
		ErrorMessage: errorMessage,
	}
}

// NewMTRHop creates a new MTR hop with calculated statistics
func NewMTRHop(hopNumber int, ip string, rtts []float64, sent int) MTRHop {
	hop := MTRHop{
		HopNumber: hopNumber,
		IP:        ip,
		Sent:      sent,
	}

	if len(rtts) == 0 {
		hop.Received = 0
		hop.LossRate = 100.0
		return hop
	}

	hop.Received = len(rtts)
	hop.LossRate = calculateLossRate(sent, len(rtts))

	// Calculate RTT statistics
	hop.LastRTTMs = rtts[len(rtts)-1]
	hop.BestRTTMs = rtts[0]
	hop.WorstRTTMs = rtts[0]
	sum := 0.0

	for _, rtt := range rtts {
		sum += rtt
		if rtt < hop.BestRTTMs {
			hop.BestRTTMs = rtt
		}
		if rtt > hop.WorstRTTMs {
			hop.WorstRTTMs = rtt
		}
	}

	hop.AvgRTTMs = roundToTwoDecimals(sum / float64(len(rtts)))
	hop.StdDevMs = calculateStdDev(rtts, hop.AvgRTTMs)

	return hop
}

// calculateLossRate calculates the packet loss rate as a percentage
func calculateLossRate(sent, received int) float64 {
	if sent == 0 {
		return 0
	}
	lossRate := (1.0 - float64(received)/float64(sent)) * 100
	return roundToTwoDecimals(lossRate)
}

// calculateStdDev calculates the standard deviation of RTT values
func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}

	variance := sumSquares / float64(len(values))
	return roundToTwoDecimals(sqrt(variance))
}

// sqrt implements square root using Newton's method
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}

	z := x / 2.0
	for i := 0; i < 100; i++ {
		z = (z + x/z) / 2.0
	}
	return z
}

// roundToTwoDecimals rounds a float to 2 decimal places
func roundToTwoDecimals(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
