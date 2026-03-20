import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { Webhook } from '../../stores/webhooksStore'
import type { CreateWebhookRequest, WebhookEventFormat } from '../../api/webhooks'

interface WebhookFormProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  onSubmit: (data: CreateWebhookRequest) => Promise<void>
  onCancel: () => void
}

// Default event format template
const DEFAULT_EVENT_FORMAT = {
  version: '1.0',
  alert: {
    id: '{{.AlertID}}',
    metric: '{{.Metric}}',
    threshold: '{{.Threshold}}',
    current_value: '{{.CurrentValue}}',
    level: '{{.Level}}',
    node_id: '{{.NodeID}}',
    node_name: '{{.NodeName}}',
    triggered_at: '{{.TriggeredAt}}',
  },
  links: {
    alert_details: '{{.BaseURL}}/nodes/{{.NodeID}}',
    dashboard: '{{.BaseURL}}',
  },
}

export function WebhookForm({ mode, initialData, onSubmit, onCancel }: WebhookFormProps) {
  const { t } = useTranslation()
  const [url, setUrl] = useState(initialData?.url || '')
  const [eventFormat, setEventFormat] = useState(
    initialData?.eventFormat
      ? JSON.stringify(initialData.eventFormat, null, 2)
      : JSON.stringify(DEFAULT_EVENT_FORMAT, null, 2)
  )
  const [enabled, setEnabled] = useState(initialData?.enabled ?? true)

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)

  const validate = () => {
    const newErrors: Record<string, string> = {}

    // Validate URL
    if (!url) {
      newErrors.url = t('webhooks.errorUrlRequired')
    } else if (!url.startsWith('https://')) {
      newErrors.url = t('webhooks.errorUrlHttps')
    } else {
      try {
        new URL(url)
      } catch {
        newErrors.url = t('webhooks.errorUrlInvalid')
      }
    }

    // Validate JSON
    if (!eventFormat.trim()) {
      newErrors.eventFormat = t('webhooks.errorFormatRequired')
    } else {
      try {
        JSON.parse(eventFormat)
      } catch {
        newErrors.eventFormat = t('webhooks.errorFormatInvalid')
      }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) return

    setIsSubmitting(true)
    try {
      const data: CreateWebhookRequest = {
        url,
        event_format: JSON.parse(eventFormat) as WebhookEventFormat,
        enabled,
      }

      await onSubmit(data)
    } catch (error) {
      console.error('Failed to submit webhook:', error)
      throw error
    } finally {
      setIsSubmitting(false)
    }
  }

  const resetToDefault = () => {
    setEventFormat(JSON.stringify(DEFAULT_EVENT_FORMAT, null, 2))
    setErrors({})
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* URL Input */}
      <div>
        <label htmlFor="url" className="block text-sm font-medium text-[var(--color-text-secondary)]">
          {t('webhooks.webhookUrl')} <span className="text-red-500">*</span>
        </label>
        <input
          type="url"
          id="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/webhook"
          className={`mt-1 block w-full border ${
            errors.url ? 'border-red-400 dark:border-red-500' : 'border-[var(--color-input-border)]'
          } rounded-md shadow-sm py-2 px-3 bg-[var(--color-input-bg)] text-[var(--color-text-primary)] placeholder:text-[var(--color-text-placeholder)] focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm`}
        />
        {errors.url && (
          <p className="mt-2 text-sm text-red-500 dark:text-red-400">{errors.url}</p>
        )}
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">
          {t('webhooks.urlHint')}
        </p>
      </div>

      {/* Event Format JSON Editor */}
      <div>
        <div className="flex justify-between items-center">
          <label htmlFor="eventFormat" className="block text-sm font-medium text-[var(--color-text-secondary)]">
            {t('webhooks.eventFormat')} (JSON) <span className="text-red-500">*</span>
          </label>
          <button
            type="button"
            onClick={resetToDefault}
            className="text-sm text-blue-500 hover:text-blue-400 dark:text-blue-400 dark:hover:text-blue-300"
          >
            {t('webhooks.resetToDefault')}
          </button>
        </div>
        <textarea
          id="eventFormat"
          value={eventFormat}
          onChange={(e) => setEventFormat(e.target.value)}
          rows={12}
          className={`mt-1 block w-full border ${
            errors.eventFormat ? 'border-red-400 dark:border-red-500' : 'border-[var(--color-input-border)]'
          } rounded-md shadow-sm py-2 px-3 font-mono text-xs bg-[var(--color-input-bg)] text-[var(--color-text-primary)] focus:outline-none focus:ring-blue-500 focus:border-blue-500`}
          placeholder={JSON.stringify(DEFAULT_EVENT_FORMAT, null, 2)}
        />
        {errors.eventFormat && (
          <p className="mt-2 text-sm text-red-500 dark:text-red-400">{errors.eventFormat}</p>
        )}
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">
          {t('webhooks.formatHint')}
        </p>
        <details className="mt-2">
          <summary className="text-sm text-[var(--color-text-secondary)] cursor-pointer hover:text-[var(--color-text-primary)]">
            {t('webhooks.templateVars')}
          </summary>
          <ul className="mt-2 text-xs text-[var(--color-text-muted)] list-disc list-inside space-y-1">
            <li>{'{{.AlertID}}'} - Unique alert identifier</li>
            <li>{'{{.Metric}}'} - Metric name (latency, packet_loss_rate, jitter)</li>
            <li>{'{{.Threshold}}'} - Configured threshold value</li>
            <li>{'{{.CurrentValue}}'} - Actual metric value that triggered alert</li>
            <li>{'{{.Level}}'} - Alert level (P0, P1, P2)</li>
            <li>{'{{.NodeID}}'} - Node ID where alert occurred</li>
            <li>{'{{.NodeName}}'} - Node name</li>
            <li>{'{{.TriggeredAt}}'} - Alert trigger timestamp</li>
            <li>{'{{.BaseURL}}'} - Base URL of Node Pulse</li>
          </ul>
        </details>
      </div>

      {/* Enabled Toggle */}
      <div className="flex items-start">
        <div className="flex items-center h-5">
          <input
            id="enabled"
            type="checkbox"
            checked={enabled}
            onChange={(e) => setEnabled(e.target.checked)}
            className="focus:ring-blue-500 h-4 w-4 text-blue-600 border-gray-300 rounded"
          />
        </div>
        <div className="ml-3 text-sm">
          <label htmlFor="enabled" className="font-medium text-[var(--color-text-secondary)]">
            {t('status.enabled')}
          </label>
          <p className="text-[var(--color-text-muted)]">
            {t('webhooks.enabledHint')}
          </p>
        </div>
      </div>

      {/* Submit and Cancel Buttons */}
      <div className="flex justify-end gap-3 pt-4 border-t border-[var(--color-border)]">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 border border-[var(--color-border-strong)] rounded-md shadow-sm text-sm font-medium text-[var(--color-text-secondary)] bg-[var(--color-bg-surface)] hover:bg-[var(--color-hover-overlay)] focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {t('common.cancel')}
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? t('common.saving') : mode === 'create' ? t('webhooks.addWebhook') : t('webhooks.updateWebhook')}
        </button>
      </div>
    </form>
  )
}
