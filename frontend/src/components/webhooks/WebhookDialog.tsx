import { useTranslation } from 'react-i18next'
import { WebhookForm } from './WebhookForm'
import type { Webhook } from '../../stores/webhooksStore'
import type { CreateWebhookRequest } from '../../api/webhooks'

interface WebhookDialogProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  onSubmit: (data: CreateWebhookRequest) => Promise<void>
  onCancel: () => void
}

export function WebhookDialog({ mode, initialData, onSubmit, onCancel }: WebhookDialogProps) {
  const { t } = useTranslation()
  return (
    <div className="fixed inset-0 bg-black/50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border border-[var(--color-border)] shadow-lg rounded-md bg-[var(--color-bg-elevated)] max-w-3xl w-full">
        {/* Dialog Header */}
        <div className="flex justify-between items-center pb-4 border-b border-[var(--color-border)]">
          <h3 className="text-lg leading-6 font-medium text-[var(--color-text-primary)]">
            {mode === 'create' ? t('webhooks.addWebhook') : t('webhooks.editWebhook')}
          </h3>
          <button
            type="button"
            onClick={onCancel}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)] focus:outline-none"
          >
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Dialog Body */}
        <div className="mt-4">
          <WebhookForm mode={mode} initialData={initialData} onSubmit={onSubmit} onCancel={onCancel} />
        </div>
      </div>
    </div>
  )
}
