/**
 * ErrorBanner Component
 *
 * Standardized error display banner with optional retry action.
 * Used across all pages for consistent error presentation.
 */

import { statusClasses } from '../../config/designTokens'

export interface ErrorBannerProps {
  /** Error message to display */
  error: string | Error
  /** Optional retry callback */
  onRetry?: () => void
  /** Custom retry button text (default: "Retry") */
  retryText?: string
  /** Additional CSS classes */
  className?: string
}

export function ErrorBanner({
  error,
  onRetry,
  retryText = 'common.retry',
  className = '',
}: ErrorBannerProps) {
  const errorMessage = typeof error === 'string' ? error : error.message

  return (
    <div
      className={`rounded-md border-l-4 p-4 ${statusClasses.critical.bgLight} ${statusClasses.critical.bgLightDark} ${statusClasses.critical.border} ${className}`}
      role="alert"
    >
      <div className="flex">
        {/* Error icon */}
        <div className="flex-shrink-0">
          <svg
            className={`h-5 w-5 ${statusClasses.critical.text}`}
            fill="currentColor"
            viewBox="0 0 20 20"
            aria-hidden="true"
          >
            <path
              fillRule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
              clipRule="evenodd"
            />
          </svg>
        </div>

        {/* Error message */}
        <div className="ml-3 flex-1 text-red-800 dark:text-red-300">
          <p className="text-sm font-medium">{errorMessage}</p>
        </div>

        {/* Retry button */}
        {onRetry && (
          <div className="ml-auto pl-3">
            <button
              type="button"
              onClick={onRetry}
              className={`inline-flex rounded-md p-1.5 ${statusClasses.critical.text} hover:bg-red-100 dark:hover:bg-red-900/50 transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-red-500`}
              aria-label="Retry loading data"
            >
              <svg
                className="h-5 w-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
              <span className="sr-only">{retryText}</span>
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export default ErrorBanner
