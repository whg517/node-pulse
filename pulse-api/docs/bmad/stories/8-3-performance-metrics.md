# Story 8.3: Performance Metrics Collection

## Overview

As a Pulse 系统,
I can 采集性能指标,
So that 可以监控系统响应速度。

## Acceptance Criteria

**Given** 仪表盘页面被访问
**When** 每个仪表盘请求完成
**Then** 记录仪表盘加载时间（P99、P95）
**And** 记录 API 响应时间（P99、P95）
**And** 记录数据查询时间（P99、P95）
**And** 每分钟记录一次性能指标
**And** 性能指标存储在数据库或缓存中
**And** 可在系统监控仪表盘可视化显示
**And** 异常性能告警

## Implementation Design

### Architecture

The performance metrics collection system consists of:

1. **Request Tracking Middleware**: Gin middleware to track API response times
2. **Metrics Collector**: In-memory storage using a ring buffer for recent metrics
3. **Percentile Calculator**: Background goroutine to calculate P99/P95 every minute
4. **Metrics API**: REST endpoint to query performance metrics

### Metrics Collected

#### API Response Time
- **Endpoint**: All API routes
- **Measurement**: Time from request start to response completion
- **Granularity**: Per endpoint
- **Percentiles**: P50, P95, P99
- **Aggregation**: 1-minute intervals

#### Dashboard Load Time
- **Endpoints**: `/api/v1/data/metrics`, `/api/v1/data/history`, `/api/v1/data/comparison`
- **Measurement**: Total request time including database queries
- **Percentiles**: P50, P95, P99
- **Aggregation**: 1-minute intervals

#### Database Query Time
- **Measurement**: Database operation duration
- **Types**: SELECT, INSERT, UPDATE
- **Percentiles**: P50, P95, P99
- **Aggregation**: 1-minute intervals

### Storage Strategy

#### In-Memory Ring Buffer
- **Capacity**: 1440 minutes (24 hours of data)
- **Structure**: Circular buffer with automatic overwrite
- **Access**: O(1) read/write
- **Memory**: ~5-10 MB for 24 hours of metrics

#### Metrics Data Structure

```go
type MetricRecord struct {
    Timestamp     time.Time
    Endpoint      string
    Method        string
    Duration      time.Duration
    Status        int
    MetricType    string // "api", "dashboard", "database"
}

type AggregatedMetrics struct {
    TimeWindow    time.Time
    Endpoint      string
    Count         int64
    AvgDuration   time.Duration
    MinDuration   time.Duration
    MaxDuration   time.Duration
    P50           time.Duration
    P95           time.Duration
    P99           time.Duration
    SuccessRate   float64
}
```

### API Endpoints

#### GET /api/v1/metrics/performance

Query performance metrics with filtering.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| metric_type | string | No | Filter: api, dashboard, database (default: all) |
| endpoint | string | No | Filter by specific endpoint |
| start_time | string | No | ISO 8601 start time (default: 1h ago) |
| end_time | string | No | ISO 8601 end time (default: now) |
| aggregation | string | No | Time aggregation: 1m, 5m, 15m (default: 1m) |

**Response:**

```json
{
  "code": "SUCCESS",
  "message": "OK",
  "data": {
    "metrics": [
      {
        "time_window": "2024-01-01T12:00:00Z",
        "endpoint": "/api/v1/data/metrics",
        "metric_type": "api",
        "count": 1234,
        "avg_duration_ms": 45.2,
        "min_duration_ms": 12.0,
        "max_duration_ms": 234.5,
        "p50_ms": 38.0,
        "p95_ms": 89.0,
        "p99_ms": 156.0,
        "success_rate": 0.998
      }
    ],
    "summary": {
      "total_requests": 50000,
      "overall_avg_ms": 52.3,
      "overall_p95_ms": 98.0,
      "overall_p99_ms": 178.0,
      "overall_success_rate": 0.997
    }
  },
  "timestamp": "2024-01-01T12:05:00Z"
}
```

### Performance Alerting

#### Alert Thresholds

Metrics exceeding these thresholds trigger alerts:

- **API Response P99**: > 500ms
- **Dashboard Load P99**: > 1000ms
- **Database Query P99**: > 100ms
- **Error Rate**: > 1%

Alerts are logged and can be integrated with the existing alert system.

### Implementation Components

#### 1. Metrics Collector (pkg/metrics/collector.go)

- Thread-safe ring buffer implementation
- Metric recording with atomic operations
- Background aggregation goroutine
- Percentile calculation using sorting

#### 2. Performance Middleware (pkg/middleware/performance.go)

- Gin middleware for automatic request tracking
- Request timing using time.Since()
- Context injection for database query tracking
- Status code and error tracking

#### 3. Metrics Handler (internal/api/metrics.go)

- Performance metrics query endpoint
- Time range filtering
- Aggregation level support
- Summary statistics

### Performance Considerations

#### Memory Usage
- **Raw Metrics**: ~100 bytes per record
- **24h Storage**: ~5-10 MB at 1000 req/min
- **Aggregated Metrics**: ~50 bytes per minute
- **24h Storage**: ~72 KB for aggregated data

#### CPU Usage
- **Middleware Overhead**: < 1ms per request
- **Percentile Calculation**: ~10ms per minute (negligible)
- **Memory Allocation**: Minimal (ring buffer pre-allocated)

#### Concurrency
- **Lock-free Design**: Atomic operations for counters
- **Channel Communication**: For metrics aggregation
- **Goroutine Safety**: Sync mutex for ring buffer access

### Testing Strategy

#### Unit Tests
- Metric recording accuracy
- Percentile calculation correctness
- Ring buffer operations
- Concurrent access safety

#### Integration Tests
- Middleware request tracking
- API endpoint functionality
- Aggregation correctness
- Time range queries

#### Load Tests
- Performance under high load (10K req/s)
- Memory leak detection
- Goroutine leak detection
- Percentile accuracy at scale

## Technical Decisions

### In-Memory vs Database Storage

**Decision**: In-memory ring buffer for MVP

**Rationale**:
- Faster access (no I/O)
- Sufficient for 24h of recent data
- Simple implementation
- Low operational overhead

**Future Enhancement**:
- Optional DB persistence for historical analysis
- Export to time-series database (InfluxDB, TimescaleDB)

### Percentile Calculation Algorithm

**Decision**: T-Digest for approximation

**Rationale**:
- O(1) memory per metric
- O(log n) update time
- < 1% error margin
- Suitable for streaming data

**Alternative Considered**:
- Exact calculation (sorting) - O(n log n)
- Reservoir sampling - higher error margin

### Aggregation Frequency

**Decision**: 1-minute intervals

**Rationale**:
- Fine-grained enough for monitoring
- Coarse enough for storage efficiency
- Aligns with common monitoring tools
- Balances timeliness and overhead

## Dependencies

- Story 3.1: Metrics data collection (Complete) - Database query patterns
- Story 5.5: Alert Engine (Complete) - Performance alerting integration
- Story 8.1: Data Export API (Complete) - Metrics export capability

## Future Enhancements

1. **Story 8.4**: Database persistence for historical metrics
2. **Story 8.5**: Integration with Grafana/Prometheus
3. **Story 8.6**: Real-time metrics streaming (WebSocket)
4. **Story 8.7**: Custom metric types and labels
5. **Story 8.8**: Distributed tracing integration (OpenTelemetry)
6. **Story 8.9**: Automated performance regression testing
