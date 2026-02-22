/**
 * PageHeader Component
 *
 * Standardized header for all pages.
 * Contains title, optional subtitle, and optional action buttons.
 */

import type { ReactNode } from 'react'

export interface PageHeaderProps {
  /** Page title */
  title: string
  /** Optional subtitle/description */
  subtitle?: string
  /** Optional action buttons (ReactNode) */
  actions?: ReactNode
  /** Show breadcrumb navigation */
  showBreadcrumb?: boolean
}

/**
 * PageHeader Component
 */
export function PageHeader({ title, subtitle, actions }: PageHeaderProps) {
  return (
    <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          {title}
        </h1>
        {subtitle && (
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
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
  )
}

export default PageHeader
