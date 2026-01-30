# Story 3.8: Prometheus Metrics 端点

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Beacon,
I need 暴露 `/metrics` 端点,
So that Prometheus 可以抓取监控指标。

## Acceptance Criteria

**Given** Beacon 已安装和配置
**When** Prometheus 请求 `GET /metrics` 端点  
**Then** 返回遵循 Prometheus exposition format 的指标
**And** 核心指标包含：beacon_up、beacon_rtt_seconds、beacon_packet_loss_rate、beacon_jitter_ms
**And** 指标格式为文本/纯文本

**覆盖需求:** NFR-OBS-001（Prometheus Metrics）

**创建表:** 无

## Tasks / Subtasks

- [ ] 实现 Prometheus Metrics HTTP 端点 (AC: #1, #2, #3)
  - [ ] 创建 `/metrics` HTTP handler
  - [ ] 返回 text/plain; version=0.0.4 Content-Type
  - [ ] 遵循 Prometheus exposition format
  - [ ] 端口可配置（默认 2112）
- [ ] 实现核心指标采集 (AC: #2)
  - [ ] beacon_up: Gauge（Beacon 运行状态，1=运行，0=停止）
  - [ ] beacon_rtt_seconds: Gauge（最新 RTT 时延，秒为单位）
  - [ ] beacon_packet_loss_rate: Gauge（最新丢包率，0-1 范围）
  - [ ] beacon_jitter_ms: Gauge（最新抖动，毫秒为单位）
  - [ ] 添加 node_id 和 node_name 标签
- [ ] 集成探测结果数据 (AC: #2)
  - [ ] 从 HeartbeatReporter 或 ProbeScheduler 获取最新指标
  - [ ] 聚合多个探测任务的指标（取均值或最新值）
  - [ ] 处理无探测结果的情况（返回 0 或 NaN）
- [ ] 集成到 Beacon 进程管理 (AC: #1)
  - [ ] 在 `beacon start` 时启动 Metrics HTTP server
  - [ ] 在 `beacon stop` 时停止 Metrics HTTP server
  - [ ] 配置 metrics_port 和 metrics_enabled（默认 true）
- [ ] 编写单元测试 (AC: #1, #2, #3)
  - [ ] 测试 `/metrics` 端点返回格式
  - [ ] 测试核心指标值正确性
  - [ ] 测试 Prometheus exposition format 合规性
  - [ ] 测试 labels（node_id, node_name）
- [ ] 编写集成测试 (AC: #1, #2, #3)
  - [ ] 测试 Prometheus 抓取流程（模拟 HTTP GET /metrics）
  - [ ] 测试指标更新（探测结果变化后指标更新）
  - [ ] 测试无探测结果场景
  - [ ] 测试 Metrics server 启动和停止

## Dev Notes

### Architecture Compliance

**核心架构要求:**
- **Prometheus Metrics 格式**: 遵循 Prometheus exposition format [Source: architecture.md#Cross-Cutting Concerns, NFR-OBS-001]
- **核心指标**: beacon_up, beacon_rtt_seconds, beacon_packet_loss_rate, beacon_jitter_ms [Source: architecture.md:97]
- **端口配置**: 默认端口 2112（可配置）[Source: Prometheus best practices]
- **数据来源**: 从探测引擎获取最新指标（复用 ProbeScheduler）[Source: Story 3.6, 3.7]
- **标签**: node_id（UUID）和 node_name（节点名称）[Source: Story 2.5, beacon.yaml]
- **资源约束**: Metrics server 内存占用 < 5MB，CPU 占用 < 1% [Source: NFR-RES-001/002]

**Prometheus Exposition Format 要求:**
- Content-Type: `text/plain; version=0.0.4; charset=utf-8`
- 每个指标格式: `# HELP <metric_name> <description>`
- 类型声明: `# TYPE <metric_name> <type>` （gauge, counter, histogram, summary）
- 指标值: `<metric_name>{label="value"} <value> [timestamp]`
- 时间戳可选（Prometheus 默认使用抓取时间）

**核心指标定义（Prometheus format）:**
```
# HELP beacon_up Beacon running status (1=running, 0=stopped)
# TYPE beacon_up gauge
beacon_up{node_id="uuid",node_name="beacon-01"} 1

# HELP beacon_rtt_seconds Latest RTT latency in seconds
# TYPE beacon_rtt_seconds gauge
beacon_rtt_seconds{node_id="uuid",node_name="beacon-01"} 0.123

# HELP beacon_packet_loss_rate Latest packet loss rate (0-1)
# TYPE beacon_packet_loss_rate gauge
beacon_packet_loss_rate{node_id="uuid",node_name="beacon-01"} 0.005

# HELP beacon_jitter_ms Latest jitter in milliseconds
# TYPE beacon_jitter_ms gauge
beacon_jitter_ms{node_id="uuid",node_name="beacon-01"} 5.67
```

**命名约定:**
- Metrics server: MetricsServer
- Metrics handler: metricsHandler
- Metrics collector: MetricsCollector（可选，用于聚合指标）
- 函数命名: camelCase（如 startMetricsServer, handleMetrics, collectMetrics）

**Beacon 配置格式（beacon.yaml）:**
```yaml
pulse_server: "https://pulse.example.com"
node_id: "uuid-from-pulse"
node_name: "beacon-01"
metrics_enabled: true
metrics_port: 2112
probes:
  - type: "tcp_ping"
    target: "192.168.1.1"
    port: 80
```

### Technical Requirements

**依赖项:**
1. **Go 标准库** (Beacon 项目已初始化)
   - `net/http`: HTTP server，暴露 /metrics 端点
   - `fmt`: 格式化 Prometheus metrics 输出
   - `sync`: Mutex 保护指标读写（并发安全）

2. **Prometheus Client Library（可选，推荐）**
   - `github.com/prometheus/client_golang/prometheus`
   - `github.com/prometheus/client_golang/prometheus/promhttp`
   - 版本: v1.19.0 或更高（最新稳定版）
   - 优势: 自动处理 exposition format，提供 Gauge/Counter/Histogram/Summary

3. **现有探测引擎** (Story 3.4, 3.5, 3.6 已实现)
   - `ProbeScheduler`: 探测调度器（获取最新探测结果）
   - `CoreMetrics`: 核心指标结构体（rtt_ms, jitter_ms, packet_loss_rate）
   - `HeartbeatReporter`: 心跳上报器（可复用指标聚合逻辑）

4. **配置管理** (Story 2.4 已实现)
   - `config.Config`: YAML 配置文件解析（添加 metrics_enabled, metrics_port）

**实现步骤（推荐使用 Prometheus Client Library）:**

1. 在 `beacon/internal/metrics/` 更新 `metrics.go`
   - 定义 Prometheus Gauge 指标（beacon_up, beacon_rtt_seconds, beacon_packet_loss_rate, beacon_jitter_ms）
   - 使用 `prometheus.NewGaugeVec` 创建带标签的指标
   - 注册指标到 Prometheus Registry

2. 实现 Metrics 采集器
   - `MetricsCollector`: 从 ProbeScheduler 获取最新指标
   - `UpdateMetrics`: 定期更新 Prometheus 指标值（每 10-60 秒）
   - 聚合多个探测任务的指标（取均值或最新值）

3. 实现 HTTP server
   - `StartMetricsServer`: 启动 HTTP server（默认端口 2112）
   - 使用 `promhttp.Handler()` 处理 /metrics 请求
   - 支持优雅关闭（使用 context.Context）

4. 集成到 Beacon 进程管理
   - 在 `beacon start` 命令中启动 MetricsServer
   - 在 `beacon stop` 命令中停止 MetricsServer
   - 从配置文件读取 metrics_enabled 和 metrics_port

5. 实现配置管理
   - 在 `config.Config` 添加 `MetricsEnabled bool` 和 `MetricsPort int`
   - 默认值: metrics_enabled=true, metrics_port=2112
   - 验证端口范围（1024-65535）

**Metrics 结构体（使用 Prometheus Client Library）:**
```go
package metrics

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    "beacon/internal/probe"
    "beacon/internal/config"
)

// Metrics handles Prometheus metrics exposure
type Metrics struct {
    config     *config.Config
    scheduler  *probe.ProbeScheduler
    
    // Prometheus metrics
    beaconUp           *prometheus.GaugeVec
    beaconRTTSeconds   *prometheus.GaugeVec
    beaconPacketLoss   *prometheus.GaugeVec
    beaconJitterMs     *prometheus.GaugeVec
    
    registry   *prometheus.Registry
    server     *http.Server
    
    mu         sync.RWMutex
    running    bool
    stopChan   chan struct{}
}

// NewMetrics creates a new Metrics handler
func NewMetrics(cfg *config.Config, scheduler *probe.ProbeScheduler) *Metrics {
    registry := prometheus.NewRegistry()
    
    // Define Prometheus metrics with labels (node_id, node_name)
    beaconUp := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "beacon_up",
            Help: "Beacon running status (1=running, 0=stopped)",
        },
        []string{"node_id", "node_name"},
    )
    
    beaconRTTSeconds := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "beacon_rtt_seconds",
            Help: "Latest RTT latency in seconds",
        },
        []string{"node_id", "node_name"},
    )
    
    beaconPacketLoss := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "beacon_packet_loss_rate",
            Help: "Latest packet loss rate (0-1)",
        },
        []string{"node_id", "node_name"},
    )
    
    beaconJitterMs := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "beacon_jitter_ms",
            Help: "Latest jitter in milliseconds",
        },
        []string{"node_id", "node_name"},
    )
    
    // Register metrics
    registry.MustRegister(beaconUp)
    registry.MustRegister(beaconRTTSeconds)
    registry.MustRegister(beaconPacketLoss)
    registry.MustRegister(beaconJitterMs)
    
    return &Metrics{
        config:            cfg,
        scheduler:         scheduler,
        beaconUp:          beaconUp,
        beaconRTTSeconds:  beaconRTTSeconds,
        beaconPacketLoss:  beaconPacketLoss,
        beaconJitterMs:    beaconJitterMs,
        registry:          registry,
        stopChan:          make(chan struct{}),
    }
}

// Start starts the metrics server
func (m *Metrics) Start() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if !m.config.MetricsEnabled {
        log.Info("Metrics server disabled in configuration")
        return nil
    }
    
    if m.running {
        return fmt.Errorf("metrics server already running")
    }
    
    // Set beacon_up to 1 (running)
    m.beaconUp.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(1)
    
    // Create HTTP server
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
    
    addr := fmt.Sprintf(":%d", m.config.MetricsPort)
    m.server = &http.Server{
        Addr:    addr,
        Handler: mux,
    }
    
    // Start server in goroutine
    go func() {
        log.Infof("Starting Prometheus metrics server on %s", addr)
        if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Errorf("Metrics server error: %v", err)
        }
    }()
    
    // Start metrics collector
    go m.collectMetrics()
    
    m.running = true
    log.Info("Prometheus metrics server started successfully")
    return nil
}

// Stop stops the metrics server
func (m *Metrics) Stop() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if !m.running {
        return nil
    }
    
    // Set beacon_up to 0 (stopped)
    m.beaconUp.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
    
    // Stop metrics collector
    close(m.stopChan)
    
    // Shutdown HTTP server gracefully
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := m.server.Shutdown(ctx); err != nil {
        log.Errorf("Metrics server shutdown error: %v", err)
        return err
    }
    
    m.running = false
    log.Info("Prometheus metrics server stopped")
    return nil
}

// collectMetrics periodically updates Prometheus metrics from probe results
func (m *Metrics) collectMetrics() {
    ticker := time.NewTicker(10 * time.Second) // Update every 10 seconds
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.updateMetrics()
        case <-m.stopChan:
            return
        }
    }
}

// updateMetrics updates Prometheus metrics from latest probe results
func (m *Metrics) updateMetrics() {
    // Get latest probe results from scheduler
    results := m.scheduler.GetLatestResults()
    
    if len(results) == 0 {
        // No probe results, set metrics to 0 or NaN
        m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
        m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(1) // 100% loss
        m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
        return
    }
    
    // Aggregate metrics from all probe results (average)
    var totalRTT, totalPacketLoss, totalJitter float64
    count := 0
    
    for _, result := range results {
        if tcpResult, ok := result.(*probe.TCPProbeResult); ok && tcpResult.Success {
            totalRTT += tcpResult.RTTMs
            totalPacketLoss += tcpResult.PacketLossRate
            totalJitter += tcpResult.JitterMs
            count++
        } else if udpResult, ok := result.(*probe.UDPProbeResult); ok && udpResult.Success {
            totalRTT += udpResult.RTTMs
            totalPacketLoss += udpResult.PacketLossRate
            totalJitter += udpResult.JitterMs
            count++
        }
    }
    
    if count > 0 {
        // Convert RTT from milliseconds to seconds for Prometheus
        rttSeconds := (totalRTT / float64(count)) / 1000.0
        packetLossRate := totalPacketLoss / float64(count) / 100.0 // Convert % to 0-1
        jitterMs := totalJitter / float64(count)
        
        m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(rttSeconds)
        m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(packetLossRate)
        m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(jitterMs)
    } else {
        // All probes failed
        m.beaconRTTSeconds.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
        m.beaconPacketLoss.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(1) // 100% loss
        m.beaconJitterMs.WithLabelValues(m.config.NodeID, m.config.NodeName).Set(0)
    }
}
```

**配置结构体扩展（config.Config）:**
```go
type Config struct {
    PulseServer   string `yaml:"pulse_server"`
    NodeID        string `yaml:"node_id"`
    NodeName      string `yaml:"node_name"`
    MetricsEnabled bool  `yaml:"metrics_enabled"` // 新增
    MetricsPort   int    `yaml:"metrics_port"`    // 新增
    Probes        []ProbeConfig `yaml:"probes"`
}

// Default values
const (
    DefaultMetricsEnabled = true
    DefaultMetricsPort    = 2112
)
```

**集成到 beacon start 命令:**
```go
// cmd/beacon/start.go

func startCommand() {
    // Load configuration
    cfg, err := config.LoadConfig(configFile)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Create probe scheduler
    scheduler := probe.NewScheduler(cfg)
    
    // Create metrics server
    metricsServer := metrics.NewMetrics(cfg, scheduler)
    
    // Start metrics server
    if err := metricsServer.Start(); err != nil {
        log.Errorf("Failed to start metrics server: %v", err)
    }
    
    // Start probe scheduler
    scheduler.Start()
    
    // Start heartbeat reporter
    reporter := reporter.NewHeartbeatReporter(cfg, scheduler)
    reporter.StartReporting()
    
    // Wait for shutdown signal
    waitForShutdown()
    
    // Stop all services
    reporter.StopReporting()
    scheduler.Stop()
    metricsServer.Stop()
}
```

**错误处理:**
- `ERR_METRICS_PORT_IN_USE`: 端口已被占用
- `ERR_METRICS_SERVER_START_FAILED`: Metrics server 启动失败
- `ERR_INVALID_METRICS_PORT`: 端口范围无效（1024-65535）
- `ERR_NO_PROBE_RESULTS`: 无探测结果（返回默认值）

### Integration with Subsequent Stories

**依赖关系:**
- **依赖 Story 3.6**: 核心指标采集（从探测结果提取指标）
- **依赖 Story 3.7**: 心跳上报（可复用指标聚合逻辑）
- **被 Story 3.9 关联**: 结构化日志系统（记录 Metrics server 启动/停止）
- **被 Story 3.10 关联**: 调试模式（输出 Metrics 状态）

**数据流转:**
1. Beacon 执行 TCP/UDP 探测（Story 3.4, 3.5）
2. Beacon 采集核心指标（Story 3.6）
3. MetricsCollector 定期从 ProbeScheduler 获取最新指标（本故事）
4. MetricsServer 暴露 /metrics 端点（本故事）
5. Prometheus 定期抓取 /metrics 端点（外部系统）

**接口设计:**
- 本故事创建 MetricsServer 和 MetricsCollector
- 从 ProbeScheduler 获取最新探测结果（复用 GetLatestResults）
- 聚合 CoreMetrics（rtt_ms, jitter_ms, packet_loss_rate）
- 转换为 Prometheus exposition format（使用 prometheus/client_golang）

**与 Story 3.6 (核心指标采集) 协作:**
- 本故事从 TCPProbeResult 和 UDPProbeResult 提取核心指标
- 使用 RTT（rtt_ms）、丢包率（packet_loss_rate）、抖动（jitter_ms）
- 聚合多个探测任务的指标（取均值）

**与 Story 3.7 (心跳上报) 协作:**
- 本故事复用 Story 3.7 的指标聚合逻辑（aggregateMetrics）
- HeartbeatReporter 和 MetricsCollector 可以共享同一个 ProbeScheduler
- 两者独立运行，互不影响

**Prometheus 集成示例:**
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'beacon'
    static_configs:
      - targets: ['beacon-01:2112', 'beacon-02:2112']
    scrape_interval: 30s
    scrape_timeout: 10s
```

### Previous Story Intelligence

**从 Story 3.7 (Beacon 数据上报) 学到的经验:**

✅ **指标聚合模式:**
- HeartbeatReporter 已实现 aggregateMetrics 逻辑
- 从 ProbeScheduler.GetLatestResults() 获取最新探测结果
- 聚合多个探测任务的指标（取均值）
- 处理无探测结果和部分失败探测

⚠️ **关键注意事项:**
- **并发安全**: Prometheus metrics 需要并发安全（Gauge 自带锁）
- **指标更新频率**: 建议 10-60 秒更新一次（平衡性能和实时性）
- **单位转换**: RTT 从毫秒转换为秒（Prometheus 推荐基本单位），丢包率从百分比转换为 0-1
- **标签**: 使用 node_id 和 node_name 作为标签（便于 Prometheus 查询）

**从 Story 3.6 (核心指标采集) 学到的经验:**

✅ **核心指标数据结构:**
- TCPProbeResult 和 UDPProbeResult 包含核心指标
- CoreMetricsCollector 计算均值、中位数、方差、抖动
- RTT 精度：保留 2 位小数（≤1ms 要求满足）

⚠️ **关键注意事项:**
- **提取核心指标**: 从探测结果的 rtt_ms, jitter_ms, packet_loss_rate 字段提取
- **处理失败探测**: Success=false 的探测结果不应计入聚合

**从 Story 2.4 (Beacon 配置文件) 学到的经验:**

✅ **配置管理模式:**
- 使用 YAML 配置文件（beacon.yaml）
- 配置验证（必填字段、范围验证）
- 配置文件大小 ≤100KB

⚠️ **关键注意事项:**
- **默认值**: metrics_enabled 默认 true，metrics_port 默认 2112
- **端口验证**: 端口范围 1024-65535（避免与系统端口冲突）
- **配置热更新**: 本故事不需要支持热更新（Metrics 配置需要重启）

**Git 智能分析:**
- 最新提交: Story 3.7 Beacon 数据上报已完成（2026-01-30）
- Epic 3 已完成 7 个故事（3.1-3.7），本故事是第 8 个故事
- Story 3.7 代码审查已修复所有问题（13 个问题）
- 现有 metrics.go 是占位符，需要完整实现

**测试经验（从 Story 3.4, 3.5, 3.6, 3.7 学习）:**
- 单元测试覆盖：成功、失败、边界条件
- 集成测试覆盖：模拟 Prometheus 抓取
- 测试数据准备：Mock 探测结果
- 性能测试：Metrics server 响应时间 < 100ms

### Testing Requirements

**单元测试:**
- 测试 Prometheus metrics 注册
  - 测试 Gauge 指标创建
  - 测试 labels（node_id, node_name）
  - 测试 registry 注册
- 测试指标更新逻辑
  - 测试 updateMetrics（多个探测任务聚合）
  - 测试无探测结果（默认值）
  - 测试部分失败探测（仅聚合成功结果）
  - 测试单位转换（ms → s, % → 0-1）
- 测试 Metrics server 启动和停止
  - 测试 Start() 方法
  - 测试 Stop() 方法（优雅关闭）
  - 测试 beacon_up 指标（1 → 0）
  - 测试重复启动（返回错误）

**集成测试:**
- 测试 /metrics 端点
  - 测试 HTTP GET /metrics 请求
  - 测试 Content-Type（text/plain; version=0.0.4）
  - 测试 Prometheus exposition format 合规性
  - 测试核心指标值正确性
- 测试 Prometheus 抓取流程
  - 启动 Metrics server
  - 模拟 Prometheus 抓取（HTTP GET /metrics）
  - 解析响应（验证格式和值）
  - 验证 labels（node_id, node_name）
- 测试指标更新
  - 启动 Metrics server
  - 模拟探测结果变化
  - 验证指标值更新
  - 等待 10 秒（collectMetrics 更新周期）
- 测试无探测结果场景
  - 启动 Metrics server（无探测结果）
  - 验证 beacon_rtt_seconds=0, beacon_packet_loss_rate=1, beacon_jitter_ms=0
- 测试配置管理
  - 测试 metrics_enabled=false（Metrics server 不启动）
  - 测试自定义 metrics_port（2113）
  - 测试无效端口（1023, 65536）

**性能测试:**
- 测试 Metrics server 响应时间
  - /metrics 端点响应时间 < 100ms
  - 多次请求（100 次）平均响应时间 < 50ms
- 测试内存占用
  - MetricsServer 内存占用：< 5MB
  - Prometheus client_golang 库内存占用：< 10MB
- 测试 CPU 占用
  - collectMetrics 更新 CPU 占用：< 1%（每 10 秒）
  - /metrics 端点请求 CPU 占用：< 0.1%

**测试文件位置:**
- 单元测试: `beacon/internal/metrics/metrics_test.go`（更新）
- 集成测试: `beacon/tests/metrics/metrics_integration_test.go`（新增）
- Mock 数据: `beacon/tests/metrics/mock_probe_results.go`（新增）

**测试数据准备:**
- Mock 探测结果（TCPProbeResult, UDPProbeResult）
- 测试聚合逻辑：3 个探测任务（部分成功、部分失败）
- 模拟 Prometheus 抓取（HTTP GET /metrics）

**Mock 策略:**
- Mock ProbeScheduler.GetLatestResults()（返回测试探测结果）
- Mock 配置文件（metrics_enabled, metrics_port）
- Mock HTTP 请求（测试 /metrics 端点）

**Prometheus Format 验证:**
- 使用 `prometheus/testutil` 验证格式
- 或手动解析响应验证 HELP、TYPE、指标值

### Project Structure Notes

**文件组织:**
```
beacon/
├── internal/
│   ├── metrics/
│   │   ├── metrics.go              # Prometheus Metrics server（本故事更新）
│   │   └── metrics_test.go         # 单元测试（本故事更新）
│   ├── probe/
│   │   ├── scheduler.go            # 探测调度器（已存在，复用 GetLatestResults）
│   │   └── ...
│   ├── reporter/
│   │   └── heartbeat_reporter.go   # 心跳上报器（已存在，复用聚合逻辑）
│   └── config/
│       └── config.go               # 配置管理（需扩展 MetricsEnabled, MetricsPort）
├── tests/
│   └── metrics/
│       ├── metrics_integration_test.go  # 集成测试（本故事新增）
│       └── mock_probe_results.go        # Mock 探测结果（本故事新增）
├── cmd/
│   └── beacon/
│       ├── start.go                # beacon start 命令（需修改，集成 MetricsServer）
│       └── stop.go                 # beacon stop 命令（需修改，停止 MetricsServer）
└── go.mod                          # 添加依赖：github.com/prometheus/client_golang
```

**与统一项目结构对齐:**
- ✅ 遵循 Go 标准项目布局（internal/, tests/）
- ✅ 更新 metrics 包（Prometheus Metrics 功能）
- ✅ 测试文件与源代码并行组织

**无冲突检测:**
- 本故事更新 metrics.go（从占位符变为完整实现）
- 复用 ProbeScheduler（获取探测结果）
- 复用 HeartbeatReporter 的聚合逻辑（相同模式）
- 扩展 config.Config（添加 MetricsEnabled, MetricsPort）

**文件修改清单:**
- 更新: `beacon/internal/metrics/metrics.go`（完整实现）
- 更新: `beacon/internal/metrics/metrics_test.go`（添加单元测试）
- 新增: `beacon/tests/metrics/metrics_integration_test.go`
- 新增: `beacon/tests/metrics/mock_probe_results.go`
- 修改: `beacon/internal/config/config.go`（添加 MetricsEnabled, MetricsPort）
- 修改: `beacon/cmd/beacon/start.go`（集成 MetricsServer）
- 修改: `beacon/cmd/beacon/stop.go`（停止 MetricsServer）
- 修改: `beacon/go.mod`（添加 prometheus/client_golang 依赖）

**依赖库安全检查:**
- `github.com/prometheus/client_golang`: v1.19.0 或更高
- 使用 `go get github.com/prometheus/client_golang@latest` 安装最新稳定版
- 运行 `gh-advisory-database` 检查已知漏洞

### References

**Architecture 文档引用:**
- [Source: architecture.md#Cross-Cutting Concerns] - Prometheus Metrics 格式和核心指标
- [Source: architecture.md:96-97] - Beacon 暴露 /metrics 端点，核心指标定义
- [Source: architecture.md:141] - Beacon Prometheus Metrics 遵循 exposition format
- [Source: architecture.md:183] - Prometheus Metrics 接口（/metrics 端点）
- [Source: architecture.md:206] - MVP 阶段 Prometheus Metrics 基础支持
- [Source: NFR-OBS-001] - Beacon 暴露 /metrics 端点，遵循 Prometheus exposition format
- [Source: NFR-RES-001/002] - 资源约束（内存≤100M，CPU≤100 微核）

**Epics 文档引用:**
- [Source: epics.md#Epic 3] - Epic 3 技术基础和包含的 NFR
- [Source: epics.md#Story 3.8:712-729] - 完整的验收标准和需求覆盖
- [Source: NFR-OBS-001] - Beacon 暴露 /metrics 端点，遵循 Prometheus exposition format

**Previous Stories:**
- Story 3.4: TCP Ping 探测（TCPProbeResult, CoreMetrics）
- Story 3.5: UDP Ping 探测（UDPProbeResult, CoreMetrics）
- Story 3.6: 核心指标采集（CoreMetricsCollector, 指标计算）
- Story 3.7: Beacon 数据上报（HeartbeatReporter, 指标聚合逻辑）
- Story 2.4: Beacon 配置文件（YAML 配置管理）

**NFR 引用:**
- NFR-OBS-001: Beacon 暴露 /metrics 端点，遵循 Prometheus exposition format
- NFR-RES-001: Beacon 内存占用 ≤ 100MB
- NFR-RES-002: Beacon CPU 占用 ≤ 100 微核

**关键实现参考（从 Story 3.6, 3.7 学习）:**
- 探测调度器: `beacon/internal/probe/scheduler.go:51-98`（GetLatestResults 方法）
- 核心指标提取: `beacon/internal/probe/tcp_ping.go:233-245`（TCPProbeResult）
- 核心指标提取: `beacon/internal/probe/udp_ping.go:305-357`（UDPProbeResult）
- 指标聚合逻辑: `beacon/internal/reporter/heartbeat_reporter.go:206-259`（aggregateMetrics）

**Prometheus 官方文档:**
- Exposition formats: https://prometheus.io/docs/instrumenting/exposition_formats/
- Best practices: https://prometheus.io/docs/practices/naming/
- Client library (Go): https://github.com/prometheus/client_golang

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
