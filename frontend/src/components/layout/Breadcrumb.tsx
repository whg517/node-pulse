/**
 * Breadcrumb Component
 *
 * Renders breadcrumb navigation driven by BreadcrumbContext.
 * Displays: Home > Section > Subsection
 * Last item is rendered as text (not a link).
 */

import { Link } from 'react-router-dom'
import { useBreadcrumb } from './BreadcrumbContext'

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

/**
 * Breadcrumb Component
 */
export function Breadcrumb() {
  const { items } = useBreadcrumb()
  const lastIndex = items.length - 1

  return (
    <nav className="flex items-center gap-1 text-sm" aria-label="Breadcrumb">
      {items.map((item, index) => (
        <div key={`breadcrumb-${index}`} className="flex items-center gap-1">
          {index > 0 && (
            <ChevronRightIcon className="text-[var(--color-text-muted)]" />
          )}
          {index === lastIndex ? (
            <span className="font-medium text-[var(--color-text-primary)]">
              {index === 0 ? <HomeIcon /> : item.label}
            </span>
          ) : (
            <Link
              to={item.path}
              className="flex items-center text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] transition-colors"
            >
              {index === 0 ? <HomeIcon /> : item.label}
            </Link>
          )}
        </div>
      ))}
    </nav>
  )
}

export default Breadcrumb
