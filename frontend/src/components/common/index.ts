/**
 * Common Components
 *
 * Reusable UI components used across the application.
 */

export { PageContainer } from './PageContainer'
export type { PageContainerProps } from './PageContainer'

export { ErrorBanner } from './ErrorBanner'
export type { ErrorBannerProps } from './ErrorBanner'

export { ConfirmDialog } from './ConfirmDialog'
export type { ConfirmDialogProps } from './ConfirmDialog'

export { ActionButton } from './ActionButton'
export type { ActionButtonProps } from './ActionButton'

export { LoadingSpinner } from './LoadingSpinner'
export type { LoadingSpinnerProps } from './LoadingSpinner'

export { ThemeToggle } from './ThemeToggle'
export type { ThemeToggleProps } from './ThemeToggle'

export { LanguageSwitcher } from './LanguageSwitcher'
export type { LanguageSwitcherProps } from './LanguageSwitcher'

export { TimezoneSelector } from './TimezoneSelector'
export type { TimezoneSelectorProps } from './TimezoneSelector'

// Default exports
export { default as ProtectedRoute } from './ProtectedRoute'

// Named exports (without Props types as they are not exported)
export { SystemHealthIndicator } from './SystemHealthIndicator'
export { PerformanceMetricCard } from './PerformanceMetricCard'
