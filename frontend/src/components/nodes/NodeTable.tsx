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
      <div className="bg-card rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-center py-12">
          <div
            className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-primary"
            role="status"
            aria-label="Loading nodes"
          />
        </div>
      </div>
    )
  }

  if (nodes.length === 0) {
    return (
      <div className="bg-card rounded-lg shadow-sm p-6">
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-muted-foreground"
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
          <h3 className="mt-2 text-sm font-medium text-foreground">{t('nodes.noNodes')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {canEdit
              ? t('nodes.noNodesHint')
              : t('nodes.noNodesConfigured')}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-card rounded-lg shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted">
            <tr>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider"
              >
                {t('nodes.nodeName')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider"
              >
                {t('common.status')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider"
              >
                {t('nodes.region')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider"
              >
                {t('nodes.tags')}
              </th>
              <th
                scope="col"
                className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider"
              >
                {t('nodes.createdAt')}
              </th>
              {canEdit && (
                <th
                  scope="col"
                  className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider"
                >
                  {t('common.actions')}
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-card divide-y divide-border">
            {nodes.map((node) => (
              <tr key={node.id} className="hover:bg-accent/10">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex flex-col">
                    <Link
                      to={`/nodes/${node.id}`}
                      state={{ breadcrumbLabel: node.name }}
                      className="text-sm font-medium text-primary hover:text-primary"
                    >
                      {node.name}
                    </Link>
                    <span className="text-xs text-muted-foreground font-mono">{node.ip}</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <NodeStatusBadge status={node.status} />
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-foreground">
                  {node.region}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex flex-wrap gap-1">
                    {Array.isArray(node.tags) && node.tags.length > 0 ? (
                      node.tags.map((tag, index) => (
                        <span
                          key={index}
                          className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary"
                        >
                          {tag}
                        </span>
                      ))
                    ) : (
                      <span className="text-sm text-muted-foreground">—</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                  {formatDate(node.created_at)}
                </td>
                {canEdit && (
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      type="button"
                      onClick={() => onEdit?.(node.id)}
                      className="text-primary hover:text-primary mr-4"
                    >
                      {t('common.edit')}
                    </button>
                    <button
                      type="button"
                      onClick={() => onDelete?.(node.id)}
                      className="text-destructive hover:text-destructive hover:opacity-80"
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
      bgColor: 'bg-healthy-bg',
      textColor: 'text-healthy-text',
      dotColor: 'bg-healthy',
    },
    offline: {
      bgColor: 'bg-destructive/10',
      textColor: 'text-destructive',
      dotColor: 'bg-destructive',
    },
    connecting: {
      bgColor: 'bg-warning-bg',
      textColor: 'text-warning-text',
      dotColor: 'bg-warning',
    },
  }

  const config = status ? statusConfig[status] : undefined

  // Handle unknown or missing status
  if (!config) {
    return (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-muted dark:bg-accent/50 text-foreground/80 dark:text-muted-foreground">
        <span className="w-1.5 h-1.5 rounded-full bg-muted-foreground mr-1.5" aria-hidden="true" />
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
