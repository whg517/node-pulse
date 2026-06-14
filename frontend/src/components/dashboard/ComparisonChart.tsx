import { useState, useMemo } from 'react'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
  Line,
  ComposedChart,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export type TimeRange = '24h' | '7d' | '30d' | 'custom'
export type MetricType = 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'
export type GroupByType = 'region' | 'isp' | 'none'
export type ComparisonMode = 'node' | 'timeRange'

export interface ComparisonDataPoint {
  timestamp: string
  value: number
}

export interface NodeComparisonData {
  node_id: string
  node_name: string
  region?: string
  isp?: string
  data: ComparisonDataPoint[]
}

export interface TimeRangeComparisonData {
  baseline: { start: string; end: string; label?: string; data: ComparisonDataPoint[] }
  current: { start: string; end: string; label?: string; data: ComparisonDataPoint[] }
  metric: MetricType
}

export interface StatisticalSummary {
  avg: number; median: number; p95: number; p99: number; min: number; max: number; stdDev?: number
}

export interface ComparisonModeChange {
  mode: ComparisonMode
  nodes?: NodeComparisonData[]
  timeRangeData?: TimeRangeComparisonData
}

export interface ComparisonChartProps {
  nodes?: NodeComparisonData[]
  timeRangeData?: TimeRangeComparisonData
  mode?: ComparisonMode
  metric: MetricType
  timeRange?: TimeRange
  showStatistics?: boolean
  highlightDifferences?: boolean
  groupBy?: GroupByType
  height?: string
  className?: string
  onTimeRangeChange?: (range: TimeRange) => void
  onModeChange?: (change: ComparisonModeChange) => void
  isLoading?: boolean
  showPercentileStats?: boolean
  onExportPdf?: () => void
}

const metricConfig: Record<MetricType, { label: string; unit: string }> = {
  latency_ms: { label: 'Latency', unit: 'ms' },
  packet_loss_rate: { label: 'Packet Loss Rate', unit: '%' },
  jitter_ms: { label: 'Jitter', unit: 'ms' },
}

const CHART_COLORS = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)']

export default function ComparisonChart({
  nodes,
  timeRangeData,
  mode = 'node',
  metric,
  showStatistics = true,
  height = '400px',
  className = '',
  onModeChange,
  isLoading = false,
}: ComparisonChartProps) {
  const [localMode, setLocalMode] = useState<ComparisonMode>(mode)
  const config = metricConfig[metric]

  const handleModeChange = (newMode: ComparisonMode) => {
    setLocalMode(newMode)
    onModeChange?.({ mode: newMode, nodes: nodes || [], timeRangeData })
  }

  const chartData = useMemo(() => {
    if (localMode === 'timeRange' && timeRangeData) {
      const baselineMap = new Map(timeRangeData.baseline.data.map((d) => [d.timestamp, d.value]))
      const allTimestamps = Array.from(new Set([
        ...timeRangeData.baseline.data.map((d) => d.timestamp),
        ...timeRangeData.current.data.map((d) => d.timestamp),
      ])).sort()
      return allTimestamps.map((ts) => ({
        time: new Date(ts).toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }),
        baseline: baselineMap.get(ts),
        current: timeRangeData.current.data.find((d) => d.timestamp === ts)?.value,
      }))
    }
    if (!nodes || nodes.length === 0) return []
    const allTimestamps = Array.from(new Set(nodes.flatMap((n) => n.data.map((d) => d.timestamp)))).sort()
    return allTimestamps.map((ts) => {
      const point: Record<string, unknown> = {
        time: new Date(ts).toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }),
      }
      nodes.forEach((node) => {
        point[node.node_name] = node.data.find((d) => d.timestamp === ts)?.value ?? null
      })
      return point
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [localMode, nodes, timeRangeData?.baseline.data, timeRangeData?.current.data])

  const overallStats = useMemo(() => {
    if (!showStatistics || !nodes || nodes.length === 0) return null
    const allValues = nodes.flatMap((n) => n.data.map((d) => d.value))
    if (allValues.length === 0) return null
    const avg = allValues.reduce((s, v) => s + v, 0) / allValues.length
    const max = Math.max(...allValues)
    const min = Math.min(...allValues)
    return { avg, max, min, diff: max - min }
  }, [nodes, showStatistics])

  if (isLoading) {
    return (
      <Card className={className}>
        <CardContent className="flex items-center justify-center py-20 text-muted-foreground">
          Loading...
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-semibold">{config.label} Comparison</CardTitle>
        <div className="flex items-center gap-2">
          <button
            onClick={() => handleModeChange('node')}
            className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
              localMode === 'node' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-accent'
            }`}
          >
            Node
          </button>
          <button
            onClick={() => handleModeChange('timeRange')}
            className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
              localMode === 'timeRange' ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground hover:bg-accent'
            }`}
          >
            Time Range
          </button>
        </div>
      </CardHeader>
      <CardContent>
        {showStatistics && overallStats && (
          <div className="mb-4 grid grid-cols-4 gap-4 rounded-lg bg-muted/50 p-3 text-center text-sm">
            <div><div className="text-muted-foreground">Avg</div><div className="font-semibold">{overallStats.avg.toFixed(2)} {config.unit}</div></div>
            <div><div className="text-muted-foreground">Max</div><div className="font-semibold">{overallStats.max.toFixed(2)} {config.unit}</div></div>
            <div><div className="text-muted-foreground">Min</div><div className="font-semibold">{overallStats.min.toFixed(2)} {config.unit}</div></div>
            <div><div className="text-muted-foreground">Diff</div><div className="font-semibold">{overallStats.diff.toFixed(2)} {config.unit}</div></div>
          </div>
        )}

        {chartData.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
            <p className="text-sm font-medium">No Data Available</p>
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={parseInt(height) || 400} minWidth={0}>
            {localMode === 'timeRange' ? (
              <ComposedChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="time" tick={{ fontSize: 11 }} />
                <YAxis tick={{ fontSize: 11 }} unit={config.unit} />
                <Tooltip
                  contentStyle={{ backgroundColor: 'var(--popover)', border: '1px solid var(--border)', borderRadius: 'var(--radius)', fontSize: 12 }}
                />
                <Legend />
                <Line type="monotone" dataKey="baseline" stroke={CHART_COLORS[1]} strokeWidth={2} dot={false} name="Baseline" />
                <Line type="monotone" dataKey="current" stroke={CHART_COLORS[0]} strokeWidth={2} dot={false} name="Current" />
              </ComposedChart>
            ) : (
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
                <XAxis dataKey="time" tick={{ fontSize: 11 }} />
                <YAxis tick={{ fontSize: 11 }} unit={config.unit} />
                <Tooltip
                  contentStyle={{ backgroundColor: 'var(--popover)', border: '1px solid var(--border)', borderRadius: 'var(--radius)', fontSize: 12 }}
                />
                <Legend />
                {nodes?.map((node, i) => (
                  <Bar key={node.node_id} dataKey={node.node_name} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                ))}
              </BarChart>
            )}
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
