package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whg517/node-pulse/pulse/internal/cache"
	"github.com/whg517/node-pulse/pulse/internal/diagnostic"
)

// DataHandler handles data query API requests
type DataHandler struct {
	pool  *pgxpool.Pool
	cache *cache.MemoryCache
}

// NewDataHandler creates a new DataHandler
func NewDataHandler(pool *pgxpool.Pool, cache *cache.MemoryCache) *DataHandler {
	return &DataHandler{
		pool:  pool,
		cache: cache,
	}
}

// HistoryRequest represents the query parameters for historical data
type HistoryRequest struct {
	NodeIDs     []string `form:"node_id" binding:"required"`
	StartTime   string   `form:"start_time" binding:"required"`
	EndTime     string   `form:"end_time" binding:"required"`
	Metrics     []string `form:"metric" binding:"required,min=1"`
	Aggregation *string  `form:"aggregation"`
}

// DataPoint represents a single data point in the time series
type DataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// HistorySeries represents a time series for a specific node and metric
type HistorySeries struct {
	NodeID     string      `json:"node_id"`
	Metric     string      `json:"metric"`
	DataPoints []DataPoint `json:"data_points"`
	Baseline   *float64    `json:"baseline,omitempty"`
}

// HistoryResponse represents the response for historical data query
type HistoryResponse struct {
	Data        []HistorySeries `json:"data"`
	Aggregation string          `json:"aggregation"`
}

// GetHistoryHandler handles GET /api/v1/data/history
// Returns historical metrics data with optional aggregation
// @Summary		Get historical metrics data
// @Description	Retrieves historical time-series metrics for specified nodes and metrics.
// @Tags			Data
// @Accept			json
// @Produce		json
// @Param			node_id		query		string	true	"Node UUID(s) to query (repeatable)"
// @Param			start_time	query		string	true	"Start time in ISO 8601 format"
// @Param			end_time	query		string	true	"End time in ISO 8601 format"
// @Param			metric		query		string	true	"Metric(s) to retrieve: latency, packet_loss_rate, jitter (repeatable)"
// @Param			aggregation	query		string	false	"Aggregation interval: 1m, 5m, 1h"	Enums(1m, 5m, 1h)
// @Success		200	{object}	HistoryResponse		"Historical data"
// @Failure		400	{object}	map[string]interface{}	"Invalid query parameters"
// @Failure		401	{object}	map[string]interface{}	"Unauthorized"
// @Failure		500	{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/data/history [get]
func (h *DataHandler) GetHistoryHandler(c *gin.Context) {
	// Parse query parameters
	var req HistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Validate timestamps
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid start_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid end_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	// Validate time range
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid time range",
			"details": "end_time must be after start_time",
		})
		return
	}

	// Validate metrics
	validMetrics := map[string]bool{
		"latency":          true,
		"packet_loss_rate": true,
		"jitter":           true,
	}
	for _, metric := range req.Metrics {
		if !validMetrics[metric] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid metric",
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
				"error":   "Invalid aggregation",
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
					"error":   "Failed to query historical data",
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
		Data:        seriesList,
		Aggregation: aggregation,
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
		"latency":          "latency_ms",
		"packet_loss_rate": "packet_loss_rate",
		"jitter":           "jitter_ms",
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
// @Summary		Get real-time metrics
// @Description	Retrieves the most recent metrics for all nodes (or specific nodes) from the last hour.
// @Tags			Data
// @Accept			json
// @Produce		json
// @Param			node_id	query		string					false	"Filter by node UUID(s) (repeatable)"
// @Success		200		{object}	map[string]interface{}	"Real-time metrics data"
// @Failure		401		{object}	map[string]interface{}	"Unauthorized"
// @Failure		500		{object}	map[string]interface{}	"Internal server error"
// @Failure		503		{object}	map[string]interface{}	"Service unavailable (database not connected)"
// @Security		BearerAuth
// @Router			/data/metrics [get]
func (h *DataHandler) GetMetricsHandler(c *gin.Context) {
	// Check if database is available
	if h.pool == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Database unavailable",
			"details": "The server is running in degraded mode without database connectivity",
		})
		return
	}

	// Parse node IDs from query string (optional - if empty, return all nodes)
	nodeIDs := c.QueryArray("node_id")

	// Query real-time metrics from database
	ctx := context.Background()
	metricsList := make([]map[string]interface{}, 0)

	// Build query for latest metrics per node
	// If nodeIDs are provided, filter by them; otherwise return all nodes
	var query string
	var rows pgx.Rows
	var err error

	if len(nodeIDs) > 0 {
		// Filter by specific node IDs
		query = `
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
		rows, err = h.pool.Query(ctx, query, nodeIDs)
	} else {
		// Return all nodes
		query = `
			SELECT DISTINCT ON (m.node_id)
				m.node_id,
				m.latency_ms,
				m.packet_loss_rate,
				m.jitter_ms,
				m.timestamp
			FROM metrics m
			WHERE m.timestamp >= NOW() - INTERVAL '1 hour'
			ORDER BY m.node_id, m.timestamp DESC;
		`
		rows, err = h.pool.Query(ctx, query)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to query metrics",
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
				"error":   "Failed to scan metrics",
				"details": err.Error(),
			})
			return
		}

		metric := map[string]interface{}{
			"node_id":   nodeID,
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
			"error":   "Error iterating metrics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": metricsList,
	})
}

// ComparisonRequest represents the query parameters for node comparison
type ComparisonRequest struct {
	NodeIDs   []string `form:"node_ids" binding:"required,min=2,max=5"`
	StartTime string   `form:"start_time" binding:"required"`
	EndTime   string   `form:"end_time" binding:"required"`
	Metrics   []string `form:"metrics" binding:"required,min=1"`
}

// ComparisonMetricData represents metric data with statistics for a single node
type ComparisonMetricData struct {
	DataPoints []DataPoint `json:"data_points"`
	Avg        float64     `json:"avg"`
	Max        float64     `json:"max"`
	Min        float64     `json:"min"`
}

// ComparisonNodeData represents comparison data for a single node
type ComparisonNodeData struct {
	NodeID  string                          `json:"node_id"`
	Name    string                          `json:"name"`
	Metrics map[string]ComparisonMetricData `json:"metrics"`
}

// ComparisonMetricStats represents statistics for a metric across all nodes
type ComparisonMetricStats struct {
	OverallAvg  float64                    `json:"overall_avg"`
	OverallMax  float64                    `json:"overall_max"`
	OverallMin  float64                    `json:"overall_min"`
	Differences []ComparisonNodeDifference `json:"differences"`
}

// ComparisonNodeDifference represents the difference from overall average for a node
type ComparisonNodeDifference struct {
	NodeID      string  `json:"node_id"`
	DiffFromAvg float64 `json:"diff_from_avg"`
}

// ComparisonData represents the comparison response data
type ComparisonData struct {
	TimeRange struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"time_range"`
	Nodes      []ComparisonNodeData             `json:"nodes"`
	Statistics map[string]ComparisonMetricStats `json:"statistics"`
}

// ComparisonResponse represents the response for node comparison query
type ComparisonResponse struct {
	Data      ComparisonData `json:"data"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
}

// GetComparisonHandler handles GET /api/v1/data/comparison
// Returns comparison data for multiple nodes with statistics
// @Summary		Get node comparison data
// @Description	Compares metrics across 2-5 nodes over a specified time range with statistical analysis.
// @Tags			Data
// @Accept			json
// @Produce		json
// @Param			node_ids	query		string					true	"Node UUIDs to compare (2-5, repeatable)"
// @Param			start_time	query		string					true	"Start time in ISO 8601 format"
// @Param			end_time	query		string					true	"End time in ISO 8601 format"
// @Param			metrics		query		string					true	"Metrics to compare: latency, packet_loss_rate, jitter (repeatable)"
// @Success		200			{object}	ComparisonResponse		"Comparison data"
// @Failure		400			{object}	map[string]interface{}	"Invalid query parameters"
// @Failure		401			{object}	map[string]interface{}	"Unauthorized"
// @Failure		404			{object}	map[string]interface{}	"No data found for specified nodes"
// @Failure		500			{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/data/comparison [get]
func (h *DataHandler) GetComparisonHandler(c *gin.Context) {
	// Parse query parameters
	var req ComparisonRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Parse timestamps
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid start_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid end_time format",
			"details": "Must be ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
		})
		return
	}

	// Validate time range
	if endTime.Before(startTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid time range",
			"details": "end_time must be after start_time",
		})
		return
	}

	// Validate metrics
	validMetrics := map[string]bool{
		"latency":          true,
		"packet_loss_rate": true,
		"jitter":           true,
	}
	for _, metric := range req.Metrics {
		if !validMetrics[metric] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid metric",
				"details": fmt.Sprintf("Metric '%s' is not valid. Valid metrics: latency, packet_loss_rate, jitter", metric),
			})
			return
		}
	}

	// Query comparison data
	ctx := context.Background()
	comparisonData, err := h.queryComparisonData(ctx, req.NodeIDs, req.Metrics, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to query comparison data",
			"details": err.Error(),
		})
		return
	}

	// Check if we have any data
	if len(comparisonData.Nodes) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No data found",
			"details": "No metrics data found for the specified nodes and time range",
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, ComparisonResponse{
		Data:      *comparisonData,
		Message:   "Comparison data retrieved successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// queryComparisonData queries and aggregates comparison data for multiple nodes
func (h *DataHandler) queryComparisonData(
	ctx context.Context,
	nodeIDs []string,
	metrics []string,
	startTime time.Time,
	endTime time.Time,
) (*ComparisonData, error) {
	// Step 1: Query data for all nodes
	nodesData := make([]ComparisonNodeData, 0)
	allMetricValues := make(map[string][]float64) // For calculating overall statistics

	// Determine if we need real-time data (< 1 hour ago)
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	useRealtimeData := endTime.After(oneHourAgo)

	for _, nodeID := range nodeIDs {
		// Get node info
		nodeName, err := h.getNodeName(ctx, nodeID)
		if err != nil {
			// Skip node if not found, but continue with others
			continue
		}

		nodeMetrics := make(map[string]ComparisonMetricData)

		for _, metric := range metrics {
			var dataPoints []DataPoint
			var avg, max, min float64

			// Query data based on time range
			if useRealtimeData && startTime.After(oneHourAgo) {
				// All data is in real-time range (< 1 hour)
				// Query from memory cache
				dataPoints = h.queryRealtimeData(nodeID, metric, startTime, endTime)
			} else if !useRealtimeData {
				// All data is historical (> 1 hour)
				// Query from PostgreSQL
				dataPoints, err = h.queryHistoricalData(ctx, nodeID, metric, startTime, endTime)
				if err != nil {
					return nil, fmt.Errorf("failed to query historical data for node %s: %w", nodeID, err)
				}
			} else {
				// Mixed: part real-time, part historical
				// Query historical data up to 1 hour ago
				historicalEnd := oneHourAgo
				if historicalEnd.After(endTime) {
					historicalEnd = endTime
				}
				historicalPoints, _ := h.queryHistoricalData(ctx, nodeID, metric, startTime, historicalEnd)

				// Query real-time data from 1 hour ago to now
				realtimePoints := h.queryRealtimeData(nodeID, metric, oneHourAgo, endTime)

				// Merge both datasets
				dataPoints = append(historicalPoints, realtimePoints...)
			}

			// Calculate statistics if we have data
			if len(dataPoints) > 0 {
				avg, max, min = calculateStatistics(dataPoints)

				// Collect values for overall statistics
				for _, point := range dataPoints {
					allMetricValues[metric] = append(allMetricValues[metric], point.Value)
				}
			}

			nodeMetrics[metric] = ComparisonMetricData{
				DataPoints: dataPoints,
				Avg:        avg,
				Max:        max,
				Min:        min,
			}
		}

		nodesData = append(nodesData, ComparisonNodeData{
			NodeID:  nodeID,
			Name:    nodeName,
			Metrics: nodeMetrics,
		})
	}

	// Step 2: Calculate overall statistics
	statistics := make(map[string]ComparisonMetricStats)
	for _, metric := range metrics {
		values := allMetricValues[metric]
		if len(values) == 0 {
			continue
		}

		overallAvg := calculateAvg(values)
		overallMax := calculateMax(values)
		overallMin := calculateMin(values)

		// Calculate differences from average for each node
		differences := make([]ComparisonNodeDifference, 0)
		for _, nodeData := range nodesData {
			if metricData, ok := nodeData.Metrics[metric]; ok && len(metricData.DataPoints) > 0 {
				nodeAvg := metricData.Avg
				differences = append(differences, ComparisonNodeDifference{
					NodeID:      nodeData.NodeID,
					DiffFromAvg: nodeAvg - overallAvg,
				})
			}
		}

		statistics[metric] = ComparisonMetricStats{
			OverallAvg:  overallAvg,
			OverallMax:  overallMax,
			OverallMin:  overallMin,
			Differences: differences,
		}
	}

	// Step 3: Find overlapping time range
	overlapStart, overlapEnd := findOverlapTimeRange(nodesData)

	// Build response
	result := &ComparisonData{
		Nodes:      nodesData,
		Statistics: statistics,
	}
	result.TimeRange.Start = overlapStart.Format(time.RFC3339)
	result.TimeRange.End = overlapEnd.Format(time.RFC3339)

	return result, nil
}

// getNodeName retrieves the name of a node
func (h *DataHandler) getNodeName(ctx context.Context, nodeID string) (string, error) {
	var nodeName string
	query := `SELECT name FROM nodes WHERE id = $1`
	err := h.pool.QueryRow(ctx, query, nodeID).Scan(&nodeName)
	if err != nil {
		return "", err
	}
	return nodeName, nil
}

// queryRealtimeData queries real-time data from memory cache
// Note: In the current implementation, memory cache is not directly accessible from DataHandler
// This is a placeholder that queries from PostgreSQL for < 1 hour data
// TODO: Integrate with memory cache in future iterations
func (h *DataHandler) queryRealtimeData(nodeID string, metric string, startTime, endTime time.Time) []DataPoint {
	// For now, query from PostgreSQL with < 1 hour filter
	ctx := context.Background()
	dataPoints, err := h.queryHistoricalData(ctx, nodeID, metric, startTime, endTime)
	if err != nil {
		return []DataPoint{}
	}
	return dataPoints
}

// queryHistoricalData queries historical data from PostgreSQL metrics table
func (h *DataHandler) queryHistoricalData(
	ctx context.Context,
	nodeID string,
	metric string,
	startTime time.Time,
	endTime time.Time,
) ([]DataPoint, error) {
	// Map metric name to database column
	columnMap := map[string]string{
		"latency":          "latency_ms",
		"packet_loss_rate": "packet_loss_rate",
		"jitter":           "jitter_ms",
	}

	column, ok := columnMap[metric]
	if !ok {
		return nil, fmt.Errorf("invalid metric: %s", metric)
	}

	query := `
		SELECT timestamp, ` + column + ` AS value
		FROM metrics
		WHERE node_id = $1
			AND timestamp >= $2
			AND timestamp <= $3
			AND ` + column + ` IS NOT NULL
		ORDER BY timestamp ASC;
	`

	rows, err := h.pool.Query(ctx, query, nodeID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	dataPoints := make([]DataPoint, 0)
	for rows.Next() {
		var timestamp time.Time
		var value float64
		if err := rows.Scan(&timestamp, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: timestamp.Format(time.RFC3339),
			Value:     value,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return dataPoints, nil
}

// calculateStatistics calculates average, max, and min from data points
func calculateStatistics(dataPoints []DataPoint) (avg, max, min float64) {
	if len(dataPoints) == 0 {
		return 0, 0, 0
	}

	sum := 0.0
	max = dataPoints[0].Value
	min = dataPoints[0].Value

	for _, point := range dataPoints {
		sum += point.Value
		if point.Value > max {
			max = point.Value
		}
		if point.Value < min {
			min = point.Value
		}
	}

	avg = sum / float64(len(dataPoints))
	return avg, max, min
}

// calculateAvg calculates average from values
func calculateAvg(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateMax calculates maximum from values
func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

// calculateMin calculates minimum from values
func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

// findOverlapTimeRange finds the overlapping time range across all nodes
func findOverlapTimeRange(nodesData []ComparisonNodeData) (time.Time, time.Time) {
	if len(nodesData) == 0 {
		return time.Now(), time.Now()
	}

	// Find the latest start time and earliest end time across all nodes
	var latestStart, earliestEnd *time.Time

	for _, nodeData := range nodesData {
		for _, metricData := range nodeData.Metrics {
			if len(metricData.DataPoints) == 0 {
				continue
			}

			// Parse first and last timestamps
			firstTime, _ := time.Parse(time.RFC3339, metricData.DataPoints[0].Timestamp)
			lastTime, _ := time.Parse(time.RFC3339, metricData.DataPoints[len(metricData.DataPoints)-1].Timestamp)

			if latestStart == nil || firstTime.After(*latestStart) {
				latestStart = &firstTime
			}
			if earliestEnd == nil || lastTime.Before(*earliestEnd) {
				earliestEnd = &lastTime
			}
		}
	}

	// If no data found, return requested range
	if latestStart == nil || earliestEnd == nil {
		return time.Now(), time.Now()
	}

	// Ensure overlap is valid
	if earliestEnd.Before(*latestStart) {
		// No overlap, return zero range
		return *latestStart, *latestStart
	}

	return *latestStart, *earliestEnd
}

// DiagnosisRequest represents the query parameters for problem diagnosis
type DiagnosisRequest struct {
	// NodeIDs must have at least 3 nodes for statistical validity in diagnosis
	// Validation: Gin binding tag "min=3" ensures minimum node count
	NodeIDs []string `form:"node_ids" binding:"required,min=3"`
}

// DiagnosisResponse represents the response for problem diagnosis
type DiagnosisResponse struct {
	Data      diagnostic.DiagnosisResult `json:"data"`
	Message   string                     `json:"message"`
	Timestamp string                     `json:"timestamp"`
}

// GetDiagnosisHandler handles GET /api/v1/data/diagnosis
// Returns problem type diagnosis based on multi-node comparison (Story 7.4)
// @Summary		Get problem type diagnosis
// @Description	Analyzes metrics across nodes to diagnose network problem types (requires at least 3 nodes with data).
// @Tags			Data
// @Accept			json
// @Produce		json
// @Param			node_ids	query		string					false	"Node UUIDs to diagnose (repeatable; uses all nodes if empty)"
// @Success		200			{object}	DiagnosisResponse		"Diagnosis result"
// @Failure		400			{object}	map[string]interface{}	"Insufficient data (need at least 3 nodes)"
// @Failure		401			{object}	map[string]interface{}	"Unauthorized"
// @Failure		500			{object}	map[string]interface{}	"Internal server error"
// @Security		BearerAuth
// @Router			/data/diagnosis [get]
func (h *DataHandler) GetDiagnosisHandler(c *gin.Context) {
	// Parse query parameters
	var req DiagnosisRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_VALIDATION",
			"message": "Invalid query parameters",
			"details": err.Error(),
		})
		return
	}

	// Query node metrics for the last 1 hour
	ctx := context.Background()
	nodesData, err := h.queryNodesForDiagnosis(ctx, req.NodeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_QUERY_DATA",
			"message": "Failed to query node data",
			"details": err.Error(),
		})
		return
	}

	// Check if we have enough nodes with data
	if len(nodesData) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "ERR_INSUFFICIENT_DATA",
			"message": "Insufficient data for diagnosis",
			"details": fmt.Sprintf("Need at least 3 nodes with data, got %d", len(nodesData)),
		})
		return
	}

	// Perform diagnosis
	engine := diagnostic.NewDiagnosticEngine()
	result, err := engine.Diagnose(nodesData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "ERR_DIAGNOSIS",
			"message": "Diagnosis failed",
			"details": err.Error(),
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, DiagnosisResponse{
		Data:      *result,
		Message:   "Diagnosis completed",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// queryNodesForDiagnosis queries node metrics for diagnosis
// Uses memory cache for real-time data (< 1 hour) per NFR-OTHER-002
func (h *DataHandler) queryNodesForDiagnosis(ctx context.Context, nodeIDs []string) ([]diagnostic.MetricData, error) {
	// Query time window: last 1 hour
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	// Strategy: Try memory cache first for real-time data (NFR-OTHER-002)
	// If cache has insufficient data, fall back to PostgreSQL

	// Step 1: Query memory cache for all requested nodes
	type cachedMetricData struct {
		NodeID         string
		Region         string
		ISP            string
		Latency        float64
		PacketLossRate float64
		Jitter         float64
		DataPointCount int
	}

	cacheDataMap := make(map[string]cachedMetricData) // node_id -> cached data

	if h.cache != nil {
		for _, nodeID := range nodeIDs {
			// Get aggregated metrics from cache (1-minute intervals)
			aggregated := h.cache.AggregateMetricsByNode(nodeID)

			// Filter to last 1 hour and calculate averages
			if aggregated != nil && len(aggregated) > 0 {
				validMetrics := filterAggregatedByTime(aggregated, startTime, endTime)

				if len(validMetrics) > 0 {
					// Calculate averages from cached aggregated data
					avgLatency, avgPacketLoss, avgJitter := calculateAverageFromAggregated(validMetrics)

					cacheDataMap[nodeID] = cachedMetricData{
						NodeID:         nodeID,
						Latency:        avgLatency,
						PacketLossRate: avgPacketLoss,
						Jitter:         avgJitter,
						DataPointCount: len(validMetrics),
						// Note: Region and ISP will be filled from nodes table query
					}
				}
			}
		}
	}

	// Step 2: Identify nodes that need PostgreSQL data
	// (cache miss or insufficient data points)
	nodesNeedingDB := []string{}
	for _, nodeID := range nodeIDs {
		if _, exists := cacheDataMap[nodeID]; !exists {
			nodesNeedingDB = append(nodesNeedingDB, nodeID)
		}
	}

	// Step 3: Query PostgreSQL for nodes without cache data
	// Also query region/ISP info for cached nodes
	dbDataMap := make(map[string]diagnostic.MetricData)

	// Always query region/ISP info for cached nodes
	nodesNeedingRegionISP := []string{}
	for nodeID := range cacheDataMap {
		nodesNeedingRegionISP = append(nodesNeedingRegionISP, nodeID)
	}

	// Combine nodes needing full data and nodes needing region/ISP only
	allDBQueries := append(nodesNeedingDB, nodesNeedingRegionISP...)
	if len(allDBQueries) > 0 {
		// Remove duplicates
		uniqueDBQueries := make(map[string]bool)
		for _, id := range allDBQueries {
			uniqueDBQueries[id] = true
		}

		finalDBQueries := make([]string, 0, len(uniqueDBQueries))
		for id := range uniqueDBQueries {
			finalDBQueries = append(finalDBQueries, id)
		}

		query := `
			SELECT
				m.node_id,
				n.region,
				COALESCE(n.tags->>'isp', '') as isp,
				AVG(m.latency_ms) as avg_latency,
				AVG(m.packet_loss_rate) as avg_packet_loss,
				AVG(m.jitter_ms) as avg_jitter,
				COUNT(*) as data_point_count
			FROM metrics m
			JOIN nodes n ON m.node_id = n.id
			WHERE m.node_id = ANY($1)
				AND m.timestamp >= $2
				AND m.timestamp <= $3
				AND m.latency_ms IS NOT NULL
			GROUP BY m.node_id, n.region, n.tags->>'isp'
			ORDER BY m.node_id;
		`

		rows, err := h.pool.Query(ctx, query, finalDBQueries, startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to query metrics: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var nodeID, region, isp string
			var avgLatency, avgPacketLoss, avgJitter float64
			var dataPointCount int

			if err := rows.Scan(&nodeID, &region, &isp, &avgLatency, &avgPacketLoss, &avgJitter, &dataPointCount); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}

			dbDataMap[nodeID] = diagnostic.MetricData{
				NodeID:         nodeID,
				Region:         region,
				ISP:            isp,
				Latency:        avgLatency,
				PacketLossRate: avgPacketLoss,
				Jitter:         avgJitter,
				DataPointCount: dataPointCount,
			}
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w", err)
		}
	}

	// Step 4: Merge cache and DB data
	// Priority: Cache data (more recent) > DB data (fallback)
	nodesData := make([]diagnostic.MetricData, 0, len(nodeIDs))

	for _, nodeID := range nodeIDs {
		if cached, exists := cacheDataMap[nodeID]; exists {
			// Use cache data, but fill region/ISP from DB if available
			region := ""
			isp := ""

			if dbData, dbExists := dbDataMap[nodeID]; dbExists {
				region = dbData.Region
				isp = dbData.ISP
			}

			nodesData = append(nodesData, diagnostic.MetricData{
				NodeID:         cached.NodeID,
				Region:         region,
				ISP:            isp,
				Latency:        cached.Latency,
				PacketLossRate: cached.PacketLossRate,
				Jitter:         cached.Jitter,
				DataPointCount: cached.DataPointCount,
			})
		} else if dbData, exists := dbDataMap[nodeID]; exists {
			// Use DB data
			nodesData = append(nodesData, dbData)
		}
		// If neither cache nor DB has data, skip this node
	}

	return nodesData, nil
}

// filterAggregatedByTime filters aggregated metrics to within time window
func filterAggregatedByTime(aggregated []*cache.AggregatedMetrics, startTime, endTime time.Time) []*cache.AggregatedMetrics {
	filtered := make([]*cache.AggregatedMetrics, 0)
	for _, agg := range aggregated {
		if (agg.Timestamp.Equal(startTime) || agg.Timestamp.After(startTime)) &&
			(agg.Timestamp.Equal(endTime) || agg.Timestamp.Before(endTime)) {
			filtered = append(filtered, agg)
		}
	}
	return filtered
}

// calculateAverageFromAggregated calculates average metrics from aggregated cache data
// Returns averaged latency, packet loss rate, and jitter
func calculateAverageFromAggregated(aggregated []*cache.AggregatedMetrics) (avgLatency, avgPacketLoss, avgJitter float64) {
	if len(aggregated) == 0 {
		return 0, 0, 0
	}

	var sumLatency, sumPacketLoss, sumJitter float64

	for _, agg := range aggregated {
		sumLatency += agg.LatencyMs
		sumPacketLoss += agg.PacketLossRate
		sumJitter += agg.JitterMs
	}

	count := float64(len(aggregated))
	return sumLatency / count, sumPacketLoss / count, sumJitter / count
}
