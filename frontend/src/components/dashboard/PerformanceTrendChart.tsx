import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { MetricTrendData } from '@/api/performance'

interface PerformanceTrendChartProps {
  trendData: MetricTrendData[]
  targetP99?: number
  targetP95?: number
  height?: string
  className?: string
  isLoading?: boolean
}

export function PerformanceTrendChart({
  trendData,
  targetP99,
  targetP95,
  height = '400px',
  className = '',
  isLoading = false,
}: PerformanceTrendChartProps) {
  const chartData = trendData.flatMap((td) =>
    td.data_points.map((d) => ({
      ...d,
      time: new Date(d.timestamp).toLocaleString([], {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }),
    }))
  )

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
      <CardHeader>
        <CardTitle className="text-sm font-medium">Performance Trend</CardTitle>
      </CardHeader>
      <CardContent>
        {chartData.length === 0 ? (
          <div className="flex items-center justify-center text-muted-foreground text-sm" style={{ height }}>
            No data available
          </div>
        ) : (
          <ResponsiveContainer width="100%" height={parseInt(height) || 400} minWidth={0}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
              <XAxis dataKey="time" tick={{ fontSize: 11 }} className="fill-muted-foreground" />
              <YAxis tick={{ fontSize: 11 }} className="fill-muted-foreground" unit="ms" />
              <Tooltip
                contentStyle={{
                  backgroundColor: 'var(--popover)',
                  border: '1px solid var(--border)',
                  borderRadius: 'var(--radius)',
                  fontSize: 12,
                }}
              />
              {targetP95 && (
                <ReferenceLine y={targetP95} stroke="var(--chart-3)" strokeDasharray="6 3" label={{ value: `P95 ${targetP95}ms`, fontSize: 10 }} />
              )}
              {targetP99 && (
                <ReferenceLine y={targetP99} stroke="var(--destructive)" strokeDasharray="6 3" label={{ value: `P99 ${targetP99}ms`, fontSize: 10 }} />
              )}
              <Line type="monotone" dataKey="avg" stroke="var(--chart-1)" strokeWidth={2} dot={false} name="Avg" />
              <Line type="monotone" dataKey="p95" stroke="var(--chart-3)" strokeWidth={1.5} dot={false} strokeDasharray="4 2" name="P95" />
              <Line type="monotone" dataKey="p99" stroke="var(--destructive)" strokeWidth={1.5} dot={false} strokeDasharray="4 2" name="P99" />
            </LineChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  )
}
