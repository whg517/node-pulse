import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { WebhookForm } from './WebhookForm'
import type { Webhook } from '../../stores/webhooksStore'
import type { CreateWebhookRequest } from '../../api/webhooks'

interface WebhookDialogProps {
  mode: 'create' | 'edit'
  initialData?: Webhook
  open: boolean
  onSubmit: (data: CreateWebhookRequest) => Promise<void>
  onCancel: () => void
}

export function WebhookDialog({ mode, initialData, open, onSubmit, onCancel }: WebhookDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onCancel() }}>
      <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {mode === 'create' ? t('webhooks.addWebhook') : t('webhooks.editWebhook')}
          </DialogTitle>
        </DialogHeader>
        <WebhookForm mode={mode} initialData={initialData} onSubmit={onSubmit} onCancel={onCancel} />
      </DialogContent>
    </Dialog>
  )
}
