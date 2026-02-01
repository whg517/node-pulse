import { useState } from 'react'
import type { Webhook } from '../../stores/webhooksStore'

interface WebhookFormProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  onSubmit: (data: any) => Promise<void>
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
      newErrors.url = 'URL is required'
    } else if (!url.startsWith('https://')) {
      newErrors.url = 'URL must use HTTPS protocol for security (NFR-SEC-003)'
    } else {
      try {
        new URL(url)
      } catch (e) {
        newErrors.url = 'Invalid URL format'
      }
    }

    // Validate JSON
    if (!eventFormat.trim()) {
      newErrors.eventFormat = 'Event format is required'
    } else {
      try {
        JSON.parse(eventFormat)
      } catch (e) {
        newErrors.eventFormat = 'Invalid JSON format'
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
      const data: any = {
        url,
        event_format: JSON.parse(eventFormat),
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
        <label htmlFor="url" className="block text-sm font-medium text-gray-700">
          Webhook URL <span className="text-red-500">*</span>
        </label>
        <input
          type="url"
          id="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/webhook"
          className={`mt-1 block w-full border ${
            errors.url ? 'border-red-300' : 'border-gray-300'
          } rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm`}
        />
        {errors.url && (
          <p className="mt-2 text-sm text-red-600">{errors.url}</p>
        )}
        <p className="mt-1 text-sm text-gray-500">
          Must be a valid HTTPS URL. Webhooks will send POST requests to this endpoint when alerts are triggered.
        </p>
      </div>

      {/* Event Format JSON Editor */}
      <div>
        <div className="flex justify-between items-center">
          <label htmlFor="eventFormat" className="block text-sm font-medium text-gray-700">
            Event Format (JSON) <span className="text-red-500">*</span>
          </label>
          <button
            type="button"
            onClick={resetToDefault}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            Reset to Default
          </button>
        </div>
        <textarea
          id="eventFormat"
          value={eventFormat}
          onChange={(e) => setEventFormat(e.target.value)}
          rows={12}
          className={`mt-1 block w-full border ${
            errors.eventFormat ? 'border-red-300' : 'border-gray-300'
          } rounded-md shadow-sm py-2 px-3 font-mono text-xs focus:outline-none focus:ring-blue-500 focus:border-blue-500`}
          placeholder={JSON.stringify(DEFAULT_EVENT_FORMAT, null, 2)}
        />
        {errors.eventFormat && (
          <p className="mt-2 text-sm text-red-600">{errors.eventFormat}</p>
        )}
        <p className="mt-1 text-sm text-gray-500">
          Customize the JSON payload sent to this webhook. Use template variables like {'{{.AlertID}}'} for dynamic values.
        </p>
        <details className="mt-2">
          <summary className="text-sm text-gray-600 cursor-pointer hover:text-gray-800">
            Available template variables
          </summary>
          <ul className="mt-2 text-xs text-gray-500 list-disc list-inside space-y-1">
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
          <label htmlFor="enabled" className="font-medium text-gray-700">
            Enabled
          </label>
          <p className="text-gray-500">
            Uncheck to disable webhook without deleting it. Disabled webhooks won't receive alert notifications.
          </p>
        </div>
      </div>

      {/* Submit and Cancel Buttons */}
      <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isSubmitting ? 'Saving...' : mode === 'create' ? 'Add Webhook' : 'Update Webhook'}
        </button>
      </div>
    </form>
  )
}
