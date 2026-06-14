import { useTranslation } from 'react-i18next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { AlertRule } from '@/stores/types'
import type { NodeDTO } from '@/api/types'

interface AlertRulesTableProps {
  rules: AlertRule[]
  nodes: NodeDTO[]
  onEdit: (id: string) => void
  onDelete: (id: string) => void
  onToggleEnabled: (id: string, enabled: boolean) => void
  canEdit: boolean
}

export function AlertRulesTable({ rules, nodes, onEdit, onDelete, onToggleEnabled, canEdit }: AlertRulesTableProps) {
  const { t } = useTranslation()

  const getNodeName = (nodeId: string | null) => {
    if (!nodeId) return t('alerts.global')
    const node = nodes.find((n) => n.id === nodeId)
    return node?.name || nodeId
  }

  const getMetricDisplayName = (metric: string) => {
    switch (metric) {
      case 'latency': return t('metrics.latency')
      case 'packet_loss_rate': return t('metrics.packetLoss')
      case 'jitter': return t('metrics.jitter')
      default: return metric
    }
  }

  const levelVariant = (level: string): 'destructive' | 'secondary' | 'outline' => {
    if (level === 'P0') return 'destructive'
    if (level === 'P1') return 'secondary'
    return 'outline'
  }

  if (rules.length === 0) {
    return (
      <div className="text-center py-12">
        <h3 className="text-sm font-medium">{t('alerts.noRules')}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{t('alerts.noRulesHint')}</p>
        {canEdit && (
          <div className="mt-6">
            <Button onClick={() => onEdit('')}>{t('alerts.createRule')}</Button>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="rounded-lg border shadow-sm overflow-hidden">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-border">
          <thead className="bg-muted/50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.alertType')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.threshold')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('alerts.severity')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('nodes.title')}</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('common.status')}</th>
              {canEdit && (
                <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-muted-foreground">{t('common.actions')}</th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {rules.map((rule) => (
              <tr key={rule.id} className="hover:bg-muted/50">
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">{getMetricDisplayName(rule.metric)}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{rule.threshold}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <Badge variant={levelVariant(rule.level)}>{rule.level}</Badge>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">{getNodeName(rule.nodeId)}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <Badge variant={rule.enabled ? 'default' : 'outline'}>
                    {rule.enabled ? t('status.enabled') : t('status.disabled')}
                  </Badge>
                </td>
                {canEdit && (
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm space-x-2">
                    <Button variant="link" size="sm" onClick={() => onToggleEnabled(rule.id, !rule.enabled)}>
                      {rule.enabled ? t('settings.disable') : t('settings.enable')}
                    </Button>
                    <Button variant="link" size="sm" onClick={() => onEdit(rule.id)}>{t('common.edit')}</Button>
                    <Button variant="link" size="sm" className="text-destructive" onClick={() => onDelete(rule.id)}>
                      {t('common.delete')}
                    </Button>
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
