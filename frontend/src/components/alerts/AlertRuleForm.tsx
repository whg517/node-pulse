import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { AlertRule } from '../../stores/types'
import type { NodeDTO, CreateAlertRuleRequest } from '../../api/types'

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

    if (!threshold || threshold <= 0) {
      newErrors.threshold = t('alerts.errorThresholdPositive')
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) return

    setIsSubmitting(true)
    try {
      const data: CreateAlertRuleRequest = {
        metric,
        threshold,
        level,
        node_id: nodeId || null,
        enabled,
      }

      await onSubmit(data)
    } catch (error) {
      console.error('Failed to submit form:', error)
      throw error
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Metric Type Select */}
      <div>
        <label htmlFor="metric" className="block text-sm font-medium text-[var(--color-text-secondary)]">
          {t('alerts.alertType')}
        </label>
        <select
          id="metric"
          value={metric}
          onChange={(e) => setMetric(e.target.value as 'latency' | 'packet_loss_rate' | 'jitter')}
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border border-[var(--color-input-border)] bg-[var(--color-input-bg)] text-[var(--color-text-primary)] focus:outline-none focus:ring-[var(--color-brand)] focus:border-[var(--color-brand)] sm:text-sm rounded-md"
        >
          <option value="latency">{t('metrics.latency')} (ms)</option>
          <option value="packet_loss_rate">{t('metrics.packetLoss')} (%)</option>
          <option value="jitter">{t('metrics.jitter')} (ms)</option>
        </select>
      </div>

      {/* Threshold Input */}
      <div>
        <label htmlFor="threshold" className="block text-sm font-medium text-[var(--color-text-secondary)]">
          {t('alerts.threshold')}
        </label>
        <input
          type="number"
          id="threshold"
          value={threshold}
          onChange={(e) => setThreshold(Number(e.target.value))}
          min="0"
          step="0.01"
          className={`mt-1 block w-full border ${errors.threshold ? 'border-[var(--color-critical)]' : 'border-[var(--color-input-border)]'} rounded-md shadow-sm py-2 px-3 bg-[var(--color-input-bg)] text-[var(--color-text-primary)] focus:outline-none focus:ring-[var(--color-brand)] focus:border-[var(--color-brand)] sm:text-sm`}
        />
        {errors.threshold && (
          <p className="mt-2 text-sm text-[var(--color-critical)]">{errors.threshold}</p>
        )}
      </div>

      {/* Alert Level Select */}
      <div>
        <label htmlFor="level" className="block text-sm font-medium text-[var(--color-text-secondary)]">
          {t('alerts.severity')}
        </label>
        <select
          id="level"
          value={level}
          onChange={(e) => setLevel(e.target.value as 'P0' | 'P1' | 'P2')}
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border border-[var(--color-input-border)] bg-[var(--color-input-bg)] text-[var(--color-text-primary)] focus:outline-none focus:ring-[var(--color-brand)] focus:border-[var(--color-brand)] sm:text-sm rounded-md"
        >
          <option value="P0">P0 - {t('alerts.critical')}</option>
          <option value="P1">P1 - {t('alerts.warning')}</option>
          <option value="P2">P2 - {t('alerts.info')}</option>
        </select>
      </div>

      {/* Node Selection */}
      <div>
        <label htmlFor="node" className="block text-sm font-medium text-[var(--color-text-secondary)]">
          {t('alerts.scope')}
        </label>
        <select
          id="node"
          value={nodeId || ''}
          onChange={(e) => setNodeId(e.target.value || null)}
          className="mt-1 block w-full pl-3 pr-10 py-2 text-base border border-[var(--color-input-border)] bg-[var(--color-input-bg)] text-[var(--color-text-primary)] focus:outline-none focus:ring-[var(--color-brand)] focus:border-[var(--color-brand)] sm:text-sm rounded-md"
        >
          <option value="">{t('alerts.globalRule')}</option>
          {nodes.map((node) => (
            <option key={node.id} value={node.id}>
              {node.name} ({node.ip})
            </option>
          ))}
        </select>
        <p className="mt-2 text-sm text-[var(--color-text-muted)]">
          {t('alerts.scopeHint')}
        </p>
      </div>

      {/* Enabled Toggle */}
      <div className="flex items-start">
        <div className="flex items-center h-5">
          <input
            id="enabled"
            type="checkbox"
            checked={enabled}
            onChange={(e) => setEnabled(e.target.checked)}
            className="focus:ring-[var(--color-brand)] h-4 w-4 text-[var(--color-brand)] border-[var(--color-border)] rounded"
          />
        </div>
        <div className="ml-3 text-sm">
          <label htmlFor="enabled" className="font-medium text-[var(--color-text-secondary)]">
            {t('status.enabled')}
          </label>
          <p className="text-[var(--color-text-muted)]">{t('alerts.enabledHint')}</p>
        </div>
      </div>

      {/* Submit and Cancel Buttons */}
      <div className="flex justify-end gap-3 pt-4 border-t border-[var(--color-border)]">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 border border-[var(--color-border-strong)] rounded-md shadow-sm text-sm font-medium text-[var(--color-text-secondary)] bg-[var(--color-bg-surface)] hover:bg-[var(--color-hover-overlay)] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--color-brand)] disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t('common.cancel')}
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-[var(--color-brand)] disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? t('common.saving') : mode === 'create' ? t('alerts.createRule') : t('alerts.updateRule')}
        </button>
      </div>
    </form>
  )
}
