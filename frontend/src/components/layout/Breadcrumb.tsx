/**
 * Breadcrumb Component
 *
 * Auto-generates breadcrumb navigation from current route.
 * Shows: Home > Section > Subsection
 */

import { useLocation, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useMemo } from 'react'

// Icon components
function HomeIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-4 w-4 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" />
    </svg>
  )
}

function ChevronRightIcon({ className = '' }: { className?: string }) {
  return (
    <svg className={`h-4 w-4 ${className}`} fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
    </svg>
  )
}

// Route to breadcrumb mapping
const routeLabels: Record<string, string> = {
  dashboard: 'nav.dashboard',
  nodes: 'nav.nodes',
  alerts: 'nav.alerts',
  rules: 'nav.alertRules',
  records: 'nav.alertRecords',
  history: 'nav.alertHistory',
  comparison: 'nav.comparison',
  integrations: 'nav.integrations',
  webhooks: 'nav.webhooks',
  health: 'nav.systemHealth',
  reports: 'nav.reports',
  export: 'nav.export',
  performance: 'nav.performance',
  settings: 'nav.settings',
  preferences: 'nav.preferences',
  sessions: 'nav.sessions',
  users: 'nav.users',
}

interface BreadcrumbItem {
  path: string
  label: string
  isLast: boolean
}

/**
 * Breadcrumb Component
 */
export function Breadcrumb() {
  const location = useLocation()
  const { t } = useTranslation()

  const items = useMemo(() => {
    const pathSegments = location.pathname.split('/').filter(Boolean)
    const breadcrumbs: BreadcrumbItem[] = []

    // Always add home
    breadcrumbs.push({
      path: '/dashboard',
      label: 'nav.home',
      isLast: pathSegments.length === 0,
    })

    // Build breadcrumb path
    let currentPath = ''
    pathSegments.forEach((segment, index) => {
      currentPath += `/${segment}`
      const isLast = index === pathSegments.length - 1

      // Check if segment is an ID (UUID pattern or numeric)
      const isId = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment) || /^\d+$/.test(segment)

      if (isId) {
        // For ID segments, use a generic label
        breadcrumbs.push({
          path: currentPath,
          label: 'nav.details',
          isLast,
        })
      } else {
        // Use mapped label or capitalize the segment
        const label = routeLabels[segment] || `nav.${segment}`
        breadcrumbs.push({
          path: currentPath,
          label,
          isLast,
        })
      }
    })

    return breadcrumbs
  }, [location.pathname])

  return (
    <nav className="flex items-center gap-1 text-sm" aria-label="Breadcrumb">
      {items.map((item, index) => (
        <div key={item.path} className="flex items-center gap-1">
          {index > 0 && (
            <ChevronRightIcon className="text-[var(--color-text-muted)]" />
          )}
          {item.isLast ? (
            <span className="font-medium text-[var(--color-text-primary)]">
              {index === 0 ? <HomeIcon /> : t(item.label)}
            </span>
          ) : (
            <Link
              to={item.path}
              className="flex items-center text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            >
              {index === 0 ? <HomeIcon /> : t(item.label)}
            </Link>
          )}
        </div>
      ))}
    </nav>
  )
}

export default Breadcrumb
