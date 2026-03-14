import { useTranslation } from 'react-i18next'
import type { AlertRule } from '../../stores/types'
import type { NodeDTO } from '../../api/types'

interface AlertRulesTableProps {
  rules: AlertRule[]
  nodes: NodeDTO[]
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  onToggleEnabled: (id: string, enabled: boolean) => void
  canEdit: boolean
}

export function AlertRulesTable({
  rules,
  nodes,
  onEdit,
  onDelete,
  onToggleEnabled,
  canEdit,
}: AlertRulesTableProps) {
  const { t } = useTranslation()
  // Helper to get node name by ID
  const getNodeName = (nodeId: string | null) => {
    if (!nodeId) return t('alerts.global')
    const node = nodes.find((n) => n.id === nodeId)
    return node?.name || nodeId
  }

  // Helper to get level badge color
  const getLevelBadgeColor = (level: string) => {
    switch (level) {
      case 'P0':
        return 'bg-red-100 text-red-800'
      case 'P1':
        return 'bg-orange-100 text-orange-800'
      case 'P2':
        return 'bg-yellow-100 text-yellow-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  // Helper to get metric display name
  const getMetricDisplayName = (metric: string) => {
    switch (metric) {
      case 'latency':
        return t('metrics.latency')
      case 'packet_loss_rate':
        return t('metrics.packetLoss')
      case 'jitter':
        return t('metrics.jitter')
      default:
        return metric
    }
  }

  // Empty state
  if (rules.length === 0) {
    return (
      <div className="text-center py-12">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900">{t('alerts.noRules')}</h3>
        <p className="mt-1 text-sm text-gray-500">
          {t('alerts.noRulesHint')}
        </p>
        {canEdit && (
          <div className="mt-6">
            <button
              type="button"
              onClick={() => onEdit('')}
              className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              {t('alerts.createRule')}
            </button>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {t('alerts.alertType')}
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {t('alerts.threshold')}
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {t('alerts.severity')}
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {t('nodes.title')}
              </th>
              <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                {t('common.status')}
              </th>
              {canEdit && (
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  {t('common.actions')}
                </th>
              )}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {rules.map((rule) => (
              <tr key={rule.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm font-medium text-gray-900">
                    {getMetricDisplayName(rule.metric)}
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="text-sm text-gray-900">{rule.threshold}</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${getLevelBadgeColor(rule.level)}`}>
                    {rule.level}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {getNodeName(rule.nodeId)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${rule.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}`}>
                    {rule.enabled ? t('status.enabled') : t('status.disabled')}
                  </span>
                </td>
                {canEdit && (
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                    <button
                      type="button"
                      onClick={() => onToggleEnabled(rule.id, !rule.enabled)}
                      className="text-blue-600 hover:text-blue-900 mr-4"
                      title={rule.enabled ? t('settings.disable') : t('settings.enable')}
                    >
                      {rule.enabled ? t('settings.disable') : t('settings.enable')}
                    </button>
                    <button
                      type="button"
                      onClick={() => onEdit(rule.id)}
                      className="text-indigo-600 hover:text-indigo-900 mr-4"
                    >
                      {t('common.edit')}
                    </button>
                    <button
                      type="button"
                      onClick={() => onDelete(rule.id)}
                      className="text-red-600 hover:text-red-900"
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
