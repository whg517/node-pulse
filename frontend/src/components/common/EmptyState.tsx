import type { ReactNode } from 'react'

export interface EmptyStateProps {
  icon?: ReactNode
  title: string
  description?: string
  action?: ReactNode
  actionClassName?: string
  className?: string
}

export function EmptyState({ icon, title, description, action, className = '', actionClassName = '' }: EmptyStateProps) {
  return (
    <div className={`flex flex-col items-center justify-center rounded-xl border border-dashed border-[var(--color-border)] bg-[var(--color-bg-surface)] p-12 text-center ${className}`}>
      {icon && (
        <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-[var(--color-bg-muted)] text-[var(--color-text-muted)]">
          {icon}
        </div>
      )}
      <h3 className="mb-2 text-lg font-semibold text-[var(--color-text-primary)]">{title}</h3>
      {description && <p className="mb-6 max-w-sm text-sm text-[var(--color-text-secondary)]">{description}</p>}
      {action && <div className={`mt-2 ${actionClassName}`}>{action}</div>}
    </div>
  )
}
