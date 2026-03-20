/**
 * LoadingSpinner Component
 *
 * Standardized loading spinner with consistent styling.
 */

export interface LoadingSpinnerProps {
  /** Size variant */
  size?: 'sm' | 'md' | 'lg'
  /** Additional CSS classes */
  className?: string
  /** Optional label for accessibility */
  label?: string
}

export function LoadingSpinner({
  size = 'md',
  className = '',
  label = 'Loading...',
}: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'h-8 w-8',
    md: 'h-12 w-12',
    lg: 'h-16 w-16',
  }

  return (
    <div
      className={`flex items-center justify-center ${className}`}
      role="status"
      aria-label={label}
    >
      <div
        className={`${sizeClasses[size]} animate-spin rounded-full border-b-2 border-[var(--color-brand)]`}
      />
      <span className="sr-only">{label}</span>
    </div>
  )
}

export default LoadingSpinner
