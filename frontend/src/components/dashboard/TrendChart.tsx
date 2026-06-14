import { useState } from 'react'
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export type TimeRange = '24h' | '7d' | '30d'
export type MetricType = 'latency_ms' | 'packet_loss_rate' | 'jitter_ms'

export interface DataPoint {
  timestamp: string
  value: number
}

export interface TrendChartProps {
  data: DataPoint[]
  metric: MetricType
  timeRange: TimeRange
  showBaseline?: boolean
  baselineValue?: number
  height?: string
  className?: string
  onTimeRangeChange?: (range: TimeRange) => void
  isLoading?: boolean
}

const metricConfig: Record<MetricType, { label: string; unit: string; chartColor: string }> = {
  latency_ms: { label: 'Latency', unit: 'ms', chartColor: 'var(--chart-1)' },
  packet_loss_rate: { label: 'Packet Loss Rate', unit: '%', chartColor: 'var(--chart-3)' },
  jitter_ms: { label: 'Jitter', unit: 'ms', chartColor: 'var(--chart-5)' },
}

const timeRangeOptions: { value: TimeRange; label: string }[] = [
  { value: '24h', label: '24 Hours' },
  { value: '7d', label: '7 Days' },
  { value: '30d', label: '30 Days' },
]

export default function TrendChart({
  data,
  metric,
  timeRange,
  showBaseline = false,
  baselineValue,
  height = '400px',
  className = '',
  onTimeRangeChange,
  isLoading = false,
}: TrendChartProps) {
  const [localRange, setLocalRange] = useState<TimeRange>(timeRange)
  const config = metricConfig[metric]

  const formatted = data.map((d) => ({
    ...d,
    time: new Date(d.timestamp).toLocaleString([], {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }),
  }))

  const handleRangeChange = (range: TimeRange) => {
    setLocalRange(range)
    onTimeRangeChange?.(range)
  }

  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-semibold">{config.label} Trend</CardTitle>
        <div className="flex gap-1" role="group" aria-label="Time range selector">
          {timeRangeOptions.map((opt) => (
            <button
              key={opt.value}
              onClick={() => handleRangeChange(opt.value)}
              className={`rounded px-2.5 py-1 text-xs font-medium transition-colors ${
                localRange === opt.value
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-accent'
              }`}
              aria-pressed={localRange === opt.value}
              disabled={isLoading}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="flex items-center justify-center text-muted-foreground text-sm" style={{ height }}>
            Loading...
          </div>
        ) : formatted.length === 0 ? (
          <div className="flex flex-col items-center justify-center text-muted-foreground" style={{ height }}>
            <p className="text-sm font-medium">No Data Available</p>
            <p className="text-xs">No trend data for the selected range.</p>
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={parseInt(height) || 400} minWidth={0}>
            <AreaChart data={formatted}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
              <XAxis dataKey="time" tick={{ fontSize: 11 }} className="fill-muted-foreground" />
              <YAxis tick={{ fontSize: 11 }} className="fill-muted-foreground" unit={config.unit} />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--popover)',
                  border: '1px solid var(--border)',
                  borderRadius: 'var(--radius)',
                  fontSize: 12,
                }}
                formatter={(value) => [`${Number(value).toFixed(2)} ${config.unit}`, config.label]}
              />
              {showBaseline && baselineValue !== undefined && (
                <ReferenceLine y={baselineValue} stroke="var(--chart-2)" strokeDasharray="6 3" label={{ value: `Baseline ${baselineValue}${config.unit}`, fontSize: 10 }} />
              )}
              <Area
                type="monotone"
                dataKey="value"
                stroke={`hsl(${config.chartColor})`}
                fill={`hsl(${config.chartColor})`}
                fillOpacity={0.15}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4 }}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}

        {showBaseline && baselineValue !== undefined && (
          <div className="mt-3 flex items-center justify-center gap-6 text-xs text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <span className="inline-block h-0.5 w-4 rounded" style={{ backgroundColor: `hsl(${config.chartColor})` }} />
              {config.label}
            </span>
            <span className="flex items-center gap-1.5">
              <span className="inline-block h-0.5 w-4 rounded border-t-2 border-dashed border-muted-foreground" />
              Baseline ({baselineValue} {config.unit})
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
