package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DataHandler handles data query API requests
type DataHandler struct {
	pool *pgxpool.Pool
}

// NewDataHandler creates a new DataHandler
func NewDataHandler(pool *pgxpool.Pool) *DataHandler {
	return &DataHandler{
		pool: pool,
	}
}

// HistoryRequest represents the query parameters for historical data
type HistoryRequest struct {
	NodeIDs     []string  `form:"node_id" binding:"required"`
	StartTime   string    `form:"start_time" binding:"required"`
	EndTime     string    `form:"end_time" binding:"required"`
	Metrics     []string  `form:"metric" binding:"required,min=1"`
	Aggregation *string   `form:"aggregation"`
}

// DataPoint represents a single data point in the time series
type DataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// HistorySeries represents a time series for a specific node and metric
type HistorySeries struct {
	NodeID     string     `json:"node_id"`
	Metric     string     `json:"metric"`
	DataPoints []DataPoint `json:"data_points"`
	Baseline   *float64   `json:"baseline,omitempty"`
}

// HistoryResponse represents the response for historical data query
type HistoryResponse struct {
	Data          []HistorySeries `json:"data"`
	Aggregation   string          `json:"aggregation"`
}

// GetHistoryHandler handles GET /api/v1/data/history
// Returns historical metrics data with optional aggregation
func (h *DataHandler) GetHistoryHandler(c *gin.Context) {
	// Parse query parameters
	var req HistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Validate timestamps
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	// Validate time range
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid time range",
			"details": "end_time must be after start_time",
		})
		return
	}

	// Validate metrics
	validMetrics := map[string]bool{
		"latency":           true,
		"packet_loss_rate":  true,
		"jitter":            true,
	}
	for _, metric := range req.Metrics {
		if !validMetrics[metric] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid metric",
				"details": fmt.Sprintf("Metric '%s' is not valid. Valid metrics: latency, packet_loss_rate, jitter", metric),
			})
			return
		}
	}

	// Validate aggregation
	aggregation := "1m" // default
	if req.Aggregation != nil {
		validAggregations := map[string]bool{
			"1m": true,
			"5m": true,
			"1h": true,
		}
		if !validAggregations[*req.Aggregation] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid aggregation",
				"details": "Aggregation must be one of: 1m, 5m, 1h",
			})
			return
		}
		aggregation = *req.Aggregation
	}

	// Query historical data
	ctx := context.Background()
	seriesList := make([]HistorySeries, 0)

	for _, nodeID := range req.NodeIDs {
		for _, metric := range req.Metrics {
			series, err := h.queryMetricHistory(ctx, nodeID, metric, startTime, endTime, aggregation)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to query historical data",
					"details": err.Error(),
				})
				return
			}

			// Calculate baseline for 7d and 30d ranges
			duration := endTime.Sub(startTime)
			if duration >= 7*24*time.Hour {
				baseline := h.calculateBaseline(series.DataPoints)
				series.Baseline = &baseline
			}

			seriesList = append(seriesList, *series)
		}
	}

	// Return response
	c.JSON(http.StatusOK, HistoryResponse{
		Data:          seriesList,
		Aggregation:   aggregation,
	})
}

// queryMetricHistory queries historical data for a specific node and metric
func (h *DataHandler) queryMetricHistory(
	ctx context.Context,
	nodeID string,
	metric string,
	startTime time.Time,
	endTime time.Time,
	aggregation string,
) (*HistorySeries, error) {
	// Map metric name to database column
	columnMap := map[string]string{
		"latency":           "latency_ms",
		"packet_loss_rate":  "packet_loss_rate",
		"jitter":            "jitter_ms",
	}

	column, ok := columnMap[metric]
	if !ok {
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	// Determine aggregation interval
	var truncateFormat string
	switch aggregation {
	case "1m":
		truncateFormat = "minute"
	case "5m":
		// For 5-minute aggregation, we need to truncate to 5-minute buckets
		truncateFormat = "minute"
	case "1h":
		truncateFormat = "hour"
	default:
		truncateFormat = "minute"
	}

	// Query with aggregation
	// Use date_trunc for compatibility with standard PostgreSQL
	// For 5-minute buckets, we need to round to nearest 5 minutes
	var query string
	if aggregation == "5m" {
		query = `
			SELECT
				date_trunc('hour', timestamp) +
					date_trunc('hour', timestamp - date_trunc('hour', timestamp)) +
					floor(extract(minute from timestamp) / 5) * interval '5 minutes' AS bucket,
				AVG(` + column + `) AS value
			FROM metrics
			WHERE node_id = $2
				AND timestamp >= $3
				AND timestamp <= $4
				AND ` + column + ` IS NOT NULL
			GROUP BY bucket
			ORDER BY bucket ASC;
		`
	} else {
		query = `
			SELECT
				date_trunc($1::text, timestamp) AS bucket,
				AVG(` + column + `) AS value
			FROM metrics
			WHERE node_id = $2
				AND timestamp >= $3
				AND timestamp <= $4
				AND ` + column + ` IS NOT NULL
			GROUP BY bucket
			ORDER BY bucket ASC;
		`
	}

	rows, err := h.pool.Query(ctx, query, truncateFormat, nodeID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	dataPoints := make([]DataPoint, 0)
	for rows.Next() {
		var bucket time.Time
		var value float64
		if err := rows.Scan(&bucket, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: bucket.Format(time.RFC3339),
			Value:     value,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &HistorySeries{
		NodeID:     nodeID,
		Metric:     metric,
		DataPoints: dataPoints,
	}, nil
}

// calculateBaseline calculates the average value of all data points
func (h *DataHandler) calculateBaseline(dataPoints []DataPoint) float64 {
	if len(dataPoints) == 0 {
		return 0
	}

	sum := 0.0
	for _, point := range dataPoints {
		sum += point.Value
	}

	return sum / float64(len(dataPoints))
}

// GetMetricsHandler handles GET /api/v1/data/metrics
// Returns real-time metrics from memory cache
func (h *DataHandler) GetMetricsHandler(c *gin.Context) {
	// Parse node IDs from query string
	var nodeIDs []string
	if nodeIDsStr := c.QueryArray("node_id"); len(nodeIDsStr) > 0 {
		nodeIDs = nodeIDsStr
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameter",
			"details": "node_id is required",
		})
		return
	}

	// Query real-time metrics from database
	ctx := context.Background()
	metricsList := make([]map[string]interface{}, 0)

	// Build query for latest metrics per node
	query := `
		SELECT DISTINCT ON (m.node_id)
			m.node_id,
			m.latency_ms,
			m.packet_loss_rate,
			m.jitter_ms,
			m.timestamp
		FROM metrics m
		WHERE m.node_id = ANY($1)
			AND m.timestamp >= NOW() - INTERVAL '1 hour'
		ORDER BY m.node_id, m.timestamp DESC;
	`

	rows, err := h.pool.Query(ctx, query, nodeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query metrics",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var nodeID string
		var latencyMs, packetLossRate, jitterMs *float64
		var timestamp time.Time

		if err := rows.Scan(&nodeID, &latencyMs, &packetLossRate, &jitterMs, &timestamp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan metrics",
				"details": err.Error(),
			})
			return
		}

		metric := map[string]interface{}{
			"node_id": nodeID,
			"timestamp": timestamp.Format(time.RFC3339),
		}

		if latencyMs != nil {
			metric["latency_ms"] = *latencyMs
		}
		if packetLossRate != nil {
			metric["packet_loss_rate"] = *packetLossRate
		}
		if jitterMs != nil {
			metric["jitter_ms"] = *jitterMs
		}

		metricsList = append(metricsList, metric)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error iterating metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": metricsList,
	})
}
