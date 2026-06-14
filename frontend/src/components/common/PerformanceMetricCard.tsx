import type { PerformanceMetric as PerformanceMetricType } from '../../api/performance'

interface PerformanceMetricCardProps {
  metric: PerformanceMetricType
}

/**
 * PerformanceMetricCard displays a single performance metric with its current values,
 * target values, and health status.
 */
export function PerformanceMetricCard({ metric }: PerformanceMetricCardProps) {
  const isHealthy = metric.status === 'healthy'

  return (
    <div
      className={`bg-background rounded-lg shadow p-6 border-l-4 ${
        isHealthy ? 'border-healthy' : 'border-destructive'
      }`}
    >
      {/* Metric Name */}
      <h3 className="text-lg font-semibold text-foreground mb-4">
        {metric.display_name}
      </h3>

      {/* Current Values */}
      <div className="space-y-3">
        <div className="flex justify-between items-center">
          <span className="text-sm text-muted-foreground">P99 当前值</span>
          <span
            className={`text-lg font-bold ${
              metric.current_p99 > metric.target_p99
                ? 'text-destructive'
                : 'text-healthy'
            }`}
          >
            {metric.current_p99.toFixed(0)} {metric.unit}
          </span>
        </div>

        <div className="flex justify-between items-center">
          <span className="text-sm text-muted-foreground">P95 当前值</span>
          <span
            className={`text-lg font-bold ${
              metric.current_p95 > metric.target_p95
                ? 'text-destructive'
                : 'text-healthy'
            }`}
          >
            {metric.current_p95.toFixed(0)} {metric.unit}
          </span>
        </div>

        {/* Target Values (Reference) */}
        <div className="pt-3 border-t border-border">
          <div className="flex justify-between items-center text-sm">
            <span className="text-muted-foreground">目标值 P99</span>
            <span className="text-foreground/80">≤ {metric.target_p99} {metric.unit}</span>
          </div>
          <div className="flex justify-between items-center text-sm">
            <span className="text-muted-foreground">目标值 P95</span>
            <span className="text-foreground/80">≤ {metric.target_p95} {metric.unit}</span>
          </div>
        </div>

        {/* Anomaly Warning */}
        {!isHealthy && metric.anomaly && (
          <div className="mt-4 p-3 bg-destructive/10 border border-destructive/10 rounded-md">
            <div className="flex items-center">
              <svg
                className="w-5 h-5 text-destructive mr-2"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="text-sm font-medium text-destructive">
                {metric.anomaly}
              </span>
            </div>
          </div>
        )}

        {/* Health Status Badge */}
        <div className="mt-4 flex items-center justify-between">
          <span className="text-sm text-muted-foreground">状态</span>
          <span
            className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${
              isHealthy
                ? 'bg-healthy-bg text-healthy-text'
                : 'bg-destructive/10 text-destructive'
            }`}
          >
            {isHealthy ? '健康' : '异常'}
          </span>
        </div>
      </div>
    </div>
  )
}
