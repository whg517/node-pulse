import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import type { AlertRule } from '@/stores/types'
import type { NodeDTO, CreateAlertRuleRequest } from '@/api/types'

interface AlertRuleFormProps {
  mode: 'create' | 'edit'
  initialData?: AlertRule
  nodes: NodeDTO[]
  onSubmit: (data: CreateAlertRuleRequest) => Promise<void>
  onCancel: () => void
}

export function AlertRuleForm({ mode, initialData, nodes, onSubmit, onCancel }: AlertRuleFormProps) {
  const { t } = useTranslation()
  const [metric, setMetric] = useState<'latency' | 'packet_loss_rate' | 'jitter'>(initialData?.metric || 'latency')
  const [threshold, setThreshold] = useState<number>(initialData?.threshold || 0)
  const [level, setLevel] = useState<'P0' | 'P1' | 'P2'>(initialData?.level || 'P1')
  const [nodeId, setNodeId] = useState<string | null>(initialData?.nodeId || null)
  const [enabled, setEnabled] = useState<boolean>(initialData?.enabled ?? true)

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validate = () => {
    const newErrors: Record<string, string> = {}
    if (!threshold || threshold <= 0) newErrors.threshold = t('alerts.errorThresholdPositive')
    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validate()) return
    setIsSubmitting(true)
    try {
      await onSubmit({ metric, threshold, level, node_id: nodeId || null, enabled })
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="metric">{t('alerts.alertType')}</Label>
        <select
          id="metric"
          value={metric}
          onChange={(e) => setMetric(e.target.value as 'latency' | 'packet_loss_rate' | 'jitter')}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
        >
          <option value="latency">{t('metrics.latency')} (ms)</option>
          <option value="packet_loss_rate">{t('metrics.packetLoss')} (%)</option>
          <option value="jitter">{t('metrics.jitter')} (ms)</option>
        </select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="threshold">{t('alerts.threshold')}</Label>
        <Input
          type="number"
          id="threshold"
          value={threshold}
          onChange={(e) => setThreshold(Number(e.target.value))}
          min="0"
          step="0.01"
          className={errors.threshold ? 'border-destructive' : ''}
        />
        {errors.threshold && <p className="text-sm text-destructive">{errors.threshold}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="level">{t('alerts.severity')}</Label>
        <select
          id="level"
          value={level}
          onChange={(e) => setLevel(e.target.value as 'P0' | 'P1' | 'P2')}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
        >
          <option value="P0">P0 - {t('alerts.critical')}</option>
          <option value="P1">P1 - {t('alerts.warning')}</option>
          <option value="P2">P2 - {t('alerts.info')}</option>
        </select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="node">{t('alerts.scope')}</Label>
        <select
          id="node"
          value={nodeId || ''}
          onChange={(e) => setNodeId(e.target.value || null)}
          className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
        >
          <option value="">{t('alerts.globalRule')}</option>
          {nodes.map((node) => (
            <option key={node.id} value={node.id}>{node.name} ({node.ip})</option>
          ))}
        </select>
        <p className="text-sm text-muted-foreground">{t('alerts.scopeHint')}</p>
      </div>

      <div className="flex items-center gap-2">
        <Switch id="enabled" checked={enabled} onCheckedChange={setEnabled} />
        <Label htmlFor="enabled">{t('status.enabled')}</Label>
      </div>

      <div className="flex justify-end gap-3 pt-4 border-t">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>{t('common.cancel')}</Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? t('common.saving') : mode === 'create' ? t('alerts.createRule') : t('alerts.updateRule')}
        </Button>
      </div>
    </form>
  )
}
