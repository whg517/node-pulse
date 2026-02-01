# Story 7.4: Problem Type Diagnostic Engine

## Overview

As a Pulse 系统,
I need 基于多节点数据对比自动判断问题类型,
So that 可以快速定位根因.

## Acceptance Criteria

**Given** 至少3个节点的数据
**When** 对比节点指标
**Then** 自动判断问题类型
  - And 问题类型包含：节点本地故障、跨境链路问题、运营商路由问题
  - And 判断基于多节点对比（同地区 vs 跨地区）
  - And 判断置信度分为：高/中/低
  - And 判断结果实时更新
**And** 判断逻辑需要至少3个节点数据参与对比
**And** 判断时间窗口为最近1小时

## Implementation Details

### Problem Types

1. **节点本地故障 (Node Local Failure)**
   - Single node shows abnormal metrics while other nodes in same region are normal
   - High latency, packet loss, or jitter on one node only
   - Confidence: High when regional difference > 50%

2. **跨境链路问题 (Cross-Border Link Issue)**
   - All nodes in one region show abnormal metrics
   - Nodes in other regions are normal
   - Consistent pattern across all nodes in affected region
   - Confidence: High when cross-regional difference > 100ms latency or > 5% packet loss

3. **运营商路由问题 (ISP Routing Issue)**
   - Multiple nodes across different regions show similar abnormal patterns
   - Issues correlate with network paths or ISPs
   - Temporary spikes that recover
   - Confidence: Medium when pattern is widespread but intermittent

### Confidence Levels

- **High**: Strong evidence, clear pattern, > 3 nodes affected
- **Medium**: Moderate evidence, some ambiguity, 2-3 nodes affected
- **Low**: Limited evidence, inconclusive pattern, minimal data points

### API Endpoint

```
GET /api/v1/data/diagnosis?node_ids=id1,id2,id3
```

### Response Structure

```json
{
  "data": {
    "problem_type": "node_local_failure" | "cross_border_link" | "isp_routing" | "unknown",
    "confidence": "high" | "medium" | "low",
    "analysis": {
      "nodes_analyzed": 3,
      "affected_nodes": ["node-id-1"],
      "regions_analyzed": ["us-east", "eu-west"],
      "metrics": {
        "latency": {"avg": 100.5, "max": 200.0, "baseline": 50.0},
        "packet_loss_rate": {"avg": 0.05, "max": 0.10, "baseline": 0.01},
        "jitter": {"avg": 5.0, "max": 15.0, "baseline": 2.0}
      },
      "regional_comparison": {
        "us-east": {"avg_latency": 50.0, "status": "normal"},
        "eu-west": {"avg_latency": 180.0, "status": "abnormal"}
      }
    },
    "recommendation": "Check node-local network configuration",
    "timestamp": "2024-01-01T00:00:00Z"
  },
  "message": "Diagnosis completed",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Technical Decisions

- Built on top of Story 7.2 comparison API
- Uses 1-hour time window for real-time analysis
- Requires minimum 3 nodes for statistical significance
- Region-aware analysis using node.region field
- Confidence scoring algorithm based on deviation magnitude and consistency

## Dependencies

- Story 7.2: Node Comparison Query API (Complete)
- Node model with region field
- Metrics data: latency, packet_loss_rate, jitter

## Testing

- Unit tests for diagnostic logic
- Integration tests with various scenarios
- Edge cases: insufficient nodes, missing regions, no data
- Confidence calculation validation
