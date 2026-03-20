import { useTranslation } from 'react-i18next'
import type { Webhook } from '../../stores/webhooksStore'

interface WebhooksTableProps {
  webhooks: Webhook[]
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  onToggleEnabled: (id: string, enabled: boolean) => void
  canEdit: boolean
}

export function WebhooksTable({
  webhooks,
  onEdit,
  onDelete,
  onToggleEnabled,
  canEdit,
}: WebhooksTableProps) {
  const { t } = useTranslation()
  // Helper to truncate URL with tooltip
  const truncateUrl = (url: string, maxLength: number = 50) => {
    if (url.length <= maxLength) return url
    return `${url.substring(0, maxLength)}...`
  }

  // Empty state
  if (webhooks.length === 0) {
    return (
      <div className="text-center py-12">
        <svg
          className="mx-auto h-12 w-12 text-[var(--color-text-muted)]"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-[var(--color-text-primary)]">{t('webhooks.noWebhooks')}</h3>
        <p className="mt-1 text-sm text-[var(--color-text-secondary)]">
          {t('webhooks.noWebhooksHint')}
        </p>
        {canEdit && (
          <div className="mt-6">
            <button
              type="button"
              onClick={() => onEdit('')}
              className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              {t('webhooks.addWebhook')}
            </button>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="bg-[var(--color-bg-surface)] shadow rounded-lg overflow-hidden border border-[var(--color-border)]">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-[var(--color-border)]">
          <thead className="bg-[var(--color-bg-muted)]">
            <tr>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider">
                {t('webhooks.webhookUrl')}
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider">
                {t('webhooks.eventFormat')}
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider">
                {t('common.status')}
              </th>
              {canEdit && (
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wider">
                  {t('common.actions')}
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-[var(--color-bg-surface)] divide-y divide-[var(--color-border)]">
            {webhooks.map((webhook) => (
              <tr key={webhook.id} className="hover:bg-[var(--color-hover-overlay)]">
                <td className="px-6 py-4">
                  <div className="flex items-center">
                    <div className="flex-shrink-0 h-8 w-8 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
                      <svg
                        className="h-4 w-4 text-blue-600 dark:text-blue-400"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                        />
                      </svg>
                    </div>
                    <div className="ml-4">
                      <div className="text-sm font-medium text-[var(--color-text-primary)]" title={webhook.url}>
                        {truncateUrl(webhook.url)}
                      </div>
                      <div className="text-xs text-[var(--color-text-muted)] truncate" title={webhook.url}>
                        {webhook.url}
                      </div>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-[var(--color-text-secondary)]">
                  <code className="text-xs bg-[var(--color-bg-muted)] text-[var(--color-text-secondary)] px-2 py-1 rounded">
                    {Object.keys(webhook.eventFormat || {}).length} {t('webhooks.fields')}
                  </code>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span
                    className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                      webhook.enabled ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-300' : 'bg-slate-100 dark:bg-slate-700/50 text-slate-700 dark:text-slate-300'
                    }`}
                  >
                    {webhook.enabled ? t('status.enabled') : t('status.disabled')}
                  </span>
                </td>
                {canEdit && (
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      type="button"
                      onClick={() => onToggleEnabled(webhook.id, !webhook.enabled)}
                      className="text-blue-500 hover:text-blue-400 dark:text-blue-400 dark:hover:text-blue-300 mr-4"
                      title={webhook.enabled ? t('settings.disable') : t('settings.enable')}
                    >
                      {webhook.enabled ? t('settings.disable') : t('settings.enable')}
                    </button>
                    <button
                      type="button"
                      onClick={() => onEdit(webhook.id)}
                      className="text-indigo-500 hover:text-indigo-400 dark:text-indigo-400 dark:hover:text-indigo-300 mr-4"
                    >
                      {t('common.edit')}
                    </button>
                    <button
                      type="button"
                      onClick={() => onDelete(webhook.id)}
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
