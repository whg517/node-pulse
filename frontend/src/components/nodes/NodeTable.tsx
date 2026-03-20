/**
 * NodeTable Component
 *
 * Displays a table of nodes with their status, metrics, and actions.
 * Provides inline viewing and links to edit/delete actions.
 */

import { useTranslation } from 'react-i18next'
import type { NodeDTO } from '../../api/types'
import { Link } from 'react-router-dom'

interface NodeTableProps {
  nodes: NodeDTO[]
  isLoading: boolean
  canEdit: boolean
  onEdit?: (id: string) => void
  onDelete?: (id: string) => void
}

export function NodeTable({
  nodes,
  isLoading,
  canEdit,
  onEdit,
  onDelete,
}: NodeTableProps) {
  const { t } = useTranslation()
  if (isLoading) {
    return (
      <div className="bg-[var(--color-bg-surface)] rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-center py-12">
          <div
            className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
            role="status"
            aria-label="Loading nodes"
          />
        </div>
      </div>
    )
  }

  if (nodes.length === 0) {
    return (
      <div className="bg-[var(--color-bg-surface)] rounded-lg shadow-sm p-6">
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-[var(--color-text-muted)]"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">{t('nodes.noNodes')}</h3>
          <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
            {canEdit
              ? t('nodes.noNodesHint')
              : t('nodes.noNodesConfigured')}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-[var(--color-bg-surface)] rounded-lg shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-[var(--color-border)]">
          <thead className="bg-[var(--color-bg-muted)]">
            <tr>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                {t('nodes.nodeName')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                {t('common.status')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                {t('nodes.region')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                {t('nodes.tags')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
              >
                {t('nodes.createdAt')}
              </th>
              {canEdit && (
                <th
                  scope="col"
                  className="px-6 py-3 text-right text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider"
                >
                  {t('common.actions')}
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-[var(--color-bg-surface)] divide-y divide-[var(--color-border)]">
            {nodes.map((node) => (
              <tr key={node.id} className="hover:bg-[var(--color-hover-overlay)]">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex flex-col">
                    <Link
                      to={`/nodes/${node.id}`}
                      className="text-sm font-medium text-blue-500 hover:text-blue-400 dark:text-blue-400 dark:hover:text-blue-300"
                    >
                      {node.name}
                    </Link>
                    <span className="text-xs text-[var(--color-text-muted)] font-mono">{node.ip}</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <NodeStatusBadge status={node.status} />
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--color-text-primary)]">
                  {node.region}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex flex-wrap gap-1">
                    {Array.isArray(node.tags) && node.tags.length > 0 ? (
                      node.tags.map((tag, index) => (
                        <span
                          key={index}
                          className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300"
                        >
                          {tag}
                        </span>
                      ))
                    ) : (
                      <span className="text-sm text-[var(--color-text-muted)]">—</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--color-text-secondary)]">
                  {formatDate(node.created_at)}
                </td>
                {canEdit && (
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      type="button"
                      onClick={() => onEdit?.(node.id)}
                      className="text-blue-500 hover:text-blue-400 dark:text-blue-400 dark:hover:text-blue-300 mr-4"
                    >
                      {t('common.edit')}
                    </button>
                    <button
                      type="button"
                      onClick={() => onDelete?.(node.id)}
                      className="text-red-500 hover:text-red-400 dark:text-red-400 dark:hover:text-red-300"
                    >
                      {t('common.delete')}
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

/**
 * Node status badge component
 */
function NodeStatusBadge({ status }: { status?: NodeDTO['status'] }) {
  const statusConfig = {
    online: {
      bgColor: 'bg-green-100 dark:bg-green-900/30',
      textColor: 'text-green-800 dark:text-green-300',
      dotColor: 'bg-green-500',
    },
    offline: {
      bgColor: 'bg-red-100 dark:bg-red-900/30',
      textColor: 'text-red-800 dark:text-red-300',
      dotColor: 'bg-red-500',
    },
    connecting: {
      bgColor: 'bg-yellow-100 dark:bg-yellow-900/30',
      textColor: 'text-yellow-800 dark:text-yellow-300',
      dotColor: 'bg-yellow-500',
    },
  }

  const config = status ? statusConfig[status] : undefined

  // Handle unknown or missing status
  if (!config) {
    return (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-100 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300">
        <span className="w-1.5 h-1.5 rounded-full bg-slate-500 mr-1.5" aria-hidden="true" />
        {status || 'unknown'}
      </span>
    )
  }

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bgColor} ${config.textColor}`}
    >
      <span
        className={`w-1.5 h-1.5 rounded-full ${config.dotColor} mr-1.5`}
        aria-hidden="true"
      />
      {status}
    </span>
  )
}

/**
 * Format date to relative time
 */
function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays < 7) return `${diffDays} days ago`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`
  return date.toLocaleDateString()
}
