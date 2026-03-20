interface SystemHealthIndicatorProps {
  health: 'healthy' | 'unhealthy'
  className?: string
}

/**
 * SystemHealthIndicator displays the overall system health status
 * as a circular indicator with color coding.
 */
export function SystemHealthIndicator({
  health,
  className = '',
}: SystemHealthIndicatorProps) {
  const isHealthy = health === 'healthy'

  return (
    <div className={`flex items-center ${className}`}>
      <div className="relative">
        {/* Outer Ring */}
        <div
          className={`w-16 h-16 rounded-full border-4 ${
            isHealthy ? 'border-[var(--color-healthy)]' : 'border-[var(--color-critical)]'
          } flex items-center justify-center bg-white shadow-md`}
        >
          {/* Inner Circle with Pulse Animation */}
          <div
            className={`w-10 h-10 rounded-full ${
              isHealthy ? 'bg-[var(--color-healthy)]' : 'bg-[var(--color-critical)]'
            } animate-pulse`}
          />
        </div>

        {/* Status Label */}
        <div className="absolute -bottom-6 left-1/2 transform -translate-x-1/2 whitespace-nowrap">
          <span
            className={`text-xs font-medium ${
              isHealthy ? 'text-[var(--color-healthy-text)]' : 'text-[var(--color-critical-text)]'
            }`}
          >
            {isHealthy ? '健康' : '异常'}
          </span>
        </div>
      </div>
    </div>
  )
}
