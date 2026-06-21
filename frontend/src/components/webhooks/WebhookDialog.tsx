import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { WebhookForm } from './WebhookForm'
import type { Webhook } from '../../stores/webhooksStore'
import type { CreateWebhookRequest, WebhookEventFormat } from '../../api/webhooks'

interface WebhookDialogProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  open: boolean
  onSubmit: (data: CreateWebhookRequest) => Promise<void>
  onPreview?: (eventFormat: WebhookEventFormat) => Promise<WebhookEventFormat>
  onCancel: () => void
}

export function WebhookDialog({ mode, initialData, open, onSubmit, onPreview, onCancel }: WebhookDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onCancel() }}>
      <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {mode === 'create' ? t('webhooks.addWebhook') : t('webhooks.editWebhook')}
          </DialogTitle>
        </DialogHeader>
        <WebhookForm mode={mode} initialData={initialData} onSubmit={onSubmit} onPreview={onPreview} onCancel={onCancel} />
      </DialogContent>
    </Dialog>
  )
}
