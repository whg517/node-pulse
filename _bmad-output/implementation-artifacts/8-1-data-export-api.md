# Story 8.1: Data Export API

## Overview

As a 运维主管,
I can 通过 API 导出节点数据报表,
So that 可以数据分析和汇报。

## Acceptance Criteria

**Given** 用户已登录并具有管理员权限
**When** 用户发送 `GET /api/v1/data/export` 请求
**Then** 验证筛选参数（node_id、time_range、metric_type）
**And** 异步启动导出任务
**And** 从 PostgreSQL `metrics` 表查询历史数据（1 小时 - 7 天）
**And** 导出格式支持 CSV（UTF-8 编码）和 Excel
**And** 单次导出最多 50 个节点
**And** 导出文件大小限制 10MB
**And** 导出完成后通过系统消息或邮件通知用户

## API Endpoint

```
GET /api/v1/data/export
```

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| node_ids | string | Yes | Comma-separated node IDs (max 50) |
| start_time | string | Yes | ISO 8601 format start time |
| end_time | string | Yes | ISO 8601 format end time |
| metrics | string | Yes | Comma-separated metrics (latency, packet_loss_rate, jitter) |
| format | string | No | Export format: csv (default) or xlsx |

### Validation Rules

- **node_ids**: Maximum 50 nodes
- **time_range**: 1 hour to 7 days
- **file_size**: Maximum 10MB (validation during generation)
- **format**: Only CSV in MVP (Excel deferred to future)

## Response Format

### Initial Response (Async Task Creation)

```json
{
  "data": {
    "export_id": "uuid",
    "status": "pending",
    "format": "csv",
    "node_ids": ["node1", "node2"],
    "time_range": {
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-02T00:00:00Z"
    },
    "metrics": ["latency", "packet_loss_rate", "jitter"],
    "created_at": "2024-01-01T00:00:00Z"
  },
  "message": "Export task created successfully",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Export File Format

### CSV Structure

```csv
timestamp,node_id,region,metric_name,value,unit
2024-01-01T00:00:00Z,node-1,us-east,latency,50.5,ms
2024-01-01T00:00:00Z,node-1,us-east,packet_loss_rate,0.01,%
2024-01-01T00:00:00Z,node-1,us-east,jitter,2.3,ms
```

### File Naming

Format: `metrics_export_{export_id}_{timestamp}.csv`

Example: `metrics_export_a1b2c3d4-20240101-120000.csv`

## Implementation Details

### Database Tables

Uses existing `metrics` table. No new tables required for MVP.

### File Storage

- **Location**: `/tmp/exports/` directory
- **Retention**: 24 hours (cleanup via background job)
- **Download**: `/api/v1/data/export/{export_id}/download`

### Async Processing

- Go goroutines for background export generation
- In-memory task tracking (defer to persistent queue in production)
- Status polling endpoint: `GET /api/v1/data/export/{export_id}`

### Export Task States

1. **pending**: Task created, waiting to start
2. **processing**: Currently generating export file
3. **completed**: Export ready for download
4. **failed**: Error during export

### Security

- Admin-only access (RBAC)
- File size validation (10MB limit)
- Node count validation (max 50 nodes)
- Time range validation (1h - 7d)
- Input sanitization (CSV injection prevention)

## Technical Decisions

### CSV Format (MVP)

- UTF-8 encoding with BOM for Excel compatibility
- Comma-separated values
- Quoted strings to handle special characters
- Escape quotes with double quotes

### Excel Export (Deferred)

- Requires `github.com/xuri/excelize/v2` dependency
- More complex implementation
- Deferred to Story 8.2 or future iteration

### File Storage

- Local filesystem for MVP
- In-memory task tracking
- Cleanup after 24 hours
- Production: Consider S3 or object storage

### Notifications

- In-app notification via status polling
- Email notification deferred to future
- Webhook notification possible enhancement

## Dependencies

- Story 3.1: Metrics data collection (Complete)
- Story 7.2: Node comparison API (Complete) - reuses query patterns
- PostgreSQL metrics table (Existing)

## Testing

- Unit tests for export generation logic
- Integration tests for API endpoints
- Edge cases:
  - Maximum nodes (50)
  - Maximum file size (10MB)
  - Time range boundaries (1h, 7d)
  - Empty data
  - Invalid parameters
  - Concurrent exports

## Performance Considerations

- Streaming CSV generation (avoid loading all data in memory)
- Database query optimization (indexed columns)
- Background processing to avoid blocking API
- File cleanup to prevent disk space issues

## Future Enhancements

1. **Story 8.2**: Excel export support (.xlsx format)
2. **Story 8.3**: Email notifications for completed exports
3. **Story 8.4**: Export templates and custom formats
4. **Story 8.5**: Scheduled/recurring exports
5. **Story 8.6**: Export to cloud storage (S3, GCS)
6. **Story 8.7**: Export history and audit log
