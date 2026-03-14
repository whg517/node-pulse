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
    <div className={`flex flex-col items-center justify-center rounded-xl border border-dashed border-slate-200 bg-white p-12 text-center dark:border-slate-800 dark:bg-slate-900 ${className}`}>
      {icon && (
        <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-50 text-slate-400 dark:bg-slate-800 dark:text-slate-500">
          {icon}
        </div>
      )}
      <h3 className="mb-2 text-lg font-semibold text-slate-900 dark:text-white">{title}</h3>
      {description && <p className="mb-6 max-w-sm text-sm text-slate-500 dark:text-slate-400">{description}</p>}
      {action && <div className={`mt-2 ${actionClassName}`}>{action}</div>}
    </div>
  )
}
