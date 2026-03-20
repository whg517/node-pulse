/**
 * PageHeader Component
 *
 * Standardized header for all pages.
 * Contains title, optional subtitle, optional actions.
 */

import type { ReactNode } from 'react'

export interface PageHeaderProps {
  /** Page title */
  title: string
  /** Optional subtitle/description */
  subtitle?: string
  /** Optional action buttons (ReactNode) */
  actions?: ReactNode
  /** Custom className for additional styling */
  className?: string
}

/**
 * PageHeader Component
 */
export function PageHeader({
  title,
  subtitle,
  actions,
  className = '',
}: PageHeaderProps) {
  return (
    <div className={`mb-4 ${className}`}>
      {/* Title and actions */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
            {title}
          </h1>
          {subtitle && (
            <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
              {subtitle}
            </p>
          )}
        </div>
        {actions && (
          <div className="flex flex-shrink-0 items-center gap-2">
            {actions}
          </div>
        )}
      </div>
    </div>
  )
}

export default PageHeader
