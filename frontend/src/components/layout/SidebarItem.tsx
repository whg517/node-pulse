/**
 * SidebarItem Component
 *
 * Single navigation item for the sidebar.
 * Supports collapsed state with tooltip, active state highlighting, and optional badge.
 */

import { NavLink } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import type { ReactNode } from 'react'

export interface SidebarItemProps {
  /** Icon component to display */
  icon: ReactNode
  /** Navigation label (i18n key) */
  label: string
  /** Navigation path */
  path: string
  /** Optional badge count (e.g., for alerts) */
  badge?: number
  /** Whether sidebar is collapsed */
  isCollapsed: boolean
}

/**
 * SidebarItem Component
 */
export function SidebarItem({ icon, label, path, badge, isCollapsed }: SidebarItemProps) {
  const { t } = useTranslation()
  const [showTooltip, setShowTooltip] = useState(false)

  return (
    <div className="relative">
      <NavLink
        to={path}
        className={({ isActive }) =>
          `group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors duration-200
          ${isActive
            ? 'bg-blue-600 text-white dark:bg-blue-500'
            : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800'
          }
          ${isCollapsed ? 'justify-center px-2' : ''}`
        }
        onMouseEnter={() => isCollapsed && setShowTooltip(true)}
        onMouseLeave={() => setShowTooltip(false)}
      >
        {/* Icon */}
        <span className="flex-shrink-0">{icon}</span>

        {/* Label - hidden when collapsed */}
        {!isCollapsed && (
          <span className="flex-1 truncate">{t(label)}</span>
        )}

        {/* Badge - hidden when collapsed */}
        {!isCollapsed && badge !== undefined && badge > 0 && (
          <span className="ml-auto flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1.5 text-xs font-semibold text-white">
            {badge > 99 ? '99+' : badge}
          </span>
        )}

        {/* Badge indicator - shown when collapsed */}
        {isCollapsed && badge !== undefined && badge > 0 && (
          <span className="absolute right-1 top-1 h-2 w-2 rounded-full bg-red-500" />
        )}
      </NavLink>

      {/* Tooltip for collapsed state */}
      {isCollapsed && showTooltip && (
        <div className="absolute left-full top-1/2 z-50 ml-2 -translate-y-1/2 whitespace-nowrap">
          <div className="rounded-md bg-slate-900 px-3 py-1.5 text-sm font-medium text-white shadow-lg">
            {t(label)}
            {badge !== undefined && badge > 0 && (
              <span className="ml-2 text-blue-300">({badge > 99 ? '99+' : badge})</span>
            )}
          </div>
          {/* Arrow pointing left */}
          <div className="absolute left-0 top-1/2 -translate-x-1 -translate-y-1/2 border-4 border-transparent border-r-slate-900" />
        </div>
      )}
    </div>
  )
}

export default SidebarItem
