import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import type { Webhook } from '../../stores/webhooksStore'
import type { CreateWebhookRequest, WebhookEventFormat } from '../../api/webhooks'

interface WebhookFormProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  onSubmit: (data: CreateWebhookRequest) => Promise<void>
  onPreview?: (eventFormat: WebhookEventFormat) => Promise<WebhookEventFormat>
  onCancel: () => void
}

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

export function WebhookForm({ mode, initialData, onSubmit, onPreview, onCancel }: WebhookFormProps) {
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
  const [isPreviewing, setIsPreviewing] = useState(false)
  const [previewPayload, setPreviewPayload] = useState<WebhookEventFormat | null>(null)
  const [previewError, setPreviewError] = useState<string | null>(null)

  const validate = () => {
    const newErrors: Record<string, string> = {}

    if (!url) {
      newErrors.url = t('webhooks.errorUrlRequired')
    } else if (!url.startsWith('https://')) {
      newErrors.url = t('webhooks.errorUrlHttps')
    } else {
      try { new URL(url) } catch { newErrors.url = t('webhooks.errorUrlInvalid') }
    }

    if (!eventFormat.trim()) {
      newErrors.eventFormat = t('webhooks.errorFormatRequired')
    } else {
      try { JSON.parse(eventFormat) } catch { newErrors.eventFormat = t('webhooks.errorFormatInvalid') }
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const parseEventFormat = (): WebhookEventFormat | null => {
    if (!eventFormat.trim()) {
      setErrors((prev) => ({ ...prev, eventFormat: t('webhooks.errorFormatRequired') }))
      return null
    }

    try {
      const parsed = JSON.parse(eventFormat) as WebhookEventFormat
      setErrors((prev) => {
        const next = { ...prev }
        delete next.eventFormat
        return next
      })
      return parsed
    } catch {
      setErrors((prev) => ({ ...prev, eventFormat: t('webhooks.errorFormatInvalid') }))
      return null
    }
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
    setPreviewPayload(null)
    setPreviewError(null)
  }

  const handlePreview = async () => {
    if (!onPreview) return

    const parsedEventFormat = parseEventFormat()
    if (!parsedEventFormat) return

    setIsPreviewing(true)
    setPreviewError(null)
    try {
      const payload = await onPreview(parsedEventFormat)
      setPreviewPayload(payload)
    } catch (error) {
      setPreviewError(error instanceof Error ? error.message : t('webhooks.previewError'))
    } finally {
      setIsPreviewing(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="webhook-url">
          {t('webhooks.webhookUrl')} <span className="text-destructive">*</span>
        </Label>
        <Input
          id="webhook-url"
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/webhook"
        />
        {errors.url && <p className="text-sm text-destructive">{errors.url}</p>}
        <p className="text-sm text-muted-foreground">{t('webhooks.urlHint')}</p>
      </div>

      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <Label htmlFor="event-format">
            {t('webhooks.eventFormat')} (JSON) <span className="text-destructive">*</span>
          </Label>
          <div className="flex items-center gap-2">
            {onPreview && (
              <Button type="button" variant="outline" size="sm" onClick={handlePreview} disabled={isPreviewing || isSubmitting}>
                {isPreviewing ? t('webhooks.previewing') : t('webhooks.previewPayload')}
              </Button>
            )}
            <Button type="button" variant="link" size="sm" onClick={resetToDefault}>
              {t('webhooks.resetToDefault')}
            </Button>
          </div>
        </div>
        <Textarea
          id="event-format"
          value={eventFormat}
          onChange={(e) => setEventFormat(e.target.value)}
          rows={12}
          className="font-mono text-xs"
          placeholder={JSON.stringify(DEFAULT_EVENT_FORMAT, null, 2)}
        />
        {errors.eventFormat && <p className="text-sm text-destructive">{errors.eventFormat}</p>}
        <p className="text-sm text-muted-foreground">{t('webhooks.formatHint')}</p>
        <details className="mt-2">
          <summary className="text-sm text-muted-foreground cursor-pointer hover:text-foreground">
            {t('webhooks.templateVars')}
          </summary>
          <ul className="mt-2 text-xs text-muted-foreground list-disc list-inside space-y-1">
            <li>{'{{.AlertID}}'} - {t('webhooks.varAlertId', 'Unique alert identifier')}</li>
            <li>{'{{.Metric}}'} - {t('webhooks.varMetric', 'Metric name (latency, packet_loss_rate, jitter)')}</li>
            <li>{'{{.Threshold}}'} - {t('webhooks.varThreshold', 'Configured threshold value')}</li>
            <li>{'{{.CurrentValue}}'} - {t('webhooks.varCurrentValue', 'Actual metric value that triggered alert')}</li>
            <li>{'{{.Level}}'} - {t('webhooks.varLevel', 'Alert level (P0, P1, P2)')}</li>
            <li>{'{{.NodeID}}'} - {t('webhooks.varNodeId', 'Node ID where alert occurred')}</li>
            <li>{'{{.NodeName}}'} - {t('webhooks.varNodeName', 'Node name')}</li>
            <li>{'{{.TriggeredAt}}'} - {t('webhooks.varTriggeredAt', 'Alert trigger timestamp')}</li>
            <li>{'{{.BaseURL}}'} - {t('webhooks.varBaseUrl', 'Base URL of NodePulse')}</li>
          </ul>
        </details>
        {previewError && (
          <p className="text-sm text-destructive">{previewError}</p>
        )}
        {previewPayload && (
          <div className="space-y-2 rounded-md border border-border bg-muted/30 p-3">
            <p className="text-xs font-medium text-muted-foreground">{t('webhooks.previewTitle')}</p>
            <pre className="max-h-64 overflow-auto rounded bg-background p-3 text-xs text-foreground">
              {JSON.stringify(previewPayload, null, 2)}
            </pre>
          </div>
        )}
      </div>

      <div className="flex items-center gap-3">
        <Switch
          id="webhook-enabled"
          checked={enabled}
          onCheckedChange={setEnabled}
        />
        <div>
          <Label htmlFor="webhook-enabled">{t('status.enabled')}</Label>
          <p className="text-sm text-muted-foreground">{t('webhooks.enabledHint')}</p>
        </div>
      </div>

      <div className="flex justify-end gap-3 pt-4 border-t">
        <Button type="button" variant="outline" onClick={onCancel} disabled={isSubmitting}>
          {t('common.cancel')}
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? t('common.saving') : mode === 'create' ? t('webhooks.addWebhook') : t('webhooks.updateWebhook')}
        </Button>
      </div>
    </form>
  )
}
