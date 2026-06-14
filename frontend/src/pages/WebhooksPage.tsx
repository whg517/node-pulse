import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { CreateWebhookRequest } from '@/api/webhooks'
import { useAuthStore } from '@/stores/authStore'
import { useWebhooks } from '@/hooks/useWebhooks'
import { PageHeader } from '@/components/layout/PageHeader'
import { WebhooksTable } from '@/components/webhooks/WebhooksTable'
import { WebhookDialog } from '@/components/webhooks/WebhookDialog'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

export default function WebhooksPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.user)
  const { webhooks, isLoading, error, reload, createWebhook, updateWebhook, deleteWebhook, toggleWebhookEnabled, getWebhookById } = useWebhooks()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedWebhookId, setSelectedWebhookId] = useState<string | null>(null)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [webhookToDelete, setWebhookToDelete] = useState<string>()

  const selectedWebhook = useMemo(() => selectedWebhookId ? getWebhookById(selectedWebhookId) : undefined, [getWebhookById, selectedWebhookId])
  const canEdit = user?.role === 'admin'

  const handleCreate = () => { setDialogMode('create'); setSelectedWebhookId(null); setDialogOpen(true) }

  const handleEdit = (id: string) => {
    const webhook = getWebhookById(id)
    if (webhook) { setDialogMode('edit'); setSelectedWebhookId(webhook.id); setDialogOpen(true) }
  }

  const handleDelete = (id: string) => { setWebhookToDelete(id); setDeleteConfirmOpen(true) }

  const confirmDelete = async () => {
    if (!webhookToDelete) return
    try { await deleteWebhook(webhookToDelete) } catch { /* handled */ }
    setDeleteConfirmOpen(false)
    setWebhookToDelete(undefined)
  }

  const handleToggleEnabled = (id: string, enabled: boolean) => {
    void (async () => { try { await toggleWebhookEnabled(id, enabled) } catch { /* handled */ } })()
  }

  const handleSubmit = async (data: CreateWebhookRequest) => {
    if (dialogMode === 'create') { await createWebhook(data) }
    else if (selectedWebhook) { await updateWebhook(selectedWebhook.id, data) }
    else throw new Error('No webhook selected')
    setDialogOpen(false)
    setSelectedWebhookId(null)
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('webhooks.title')}
        subtitle={t('webhooks.description')}
        actions={canEdit ? <Button onClick={handleCreate}>{t('webhooks.addWebhook')}</Button> : undefined}
      />

      {!canEdit && (
        <div className="rounded-md border-l-4 border-yellow-500 bg-yellow-50 p-4 dark:bg-yellow-950">
          <p className="text-sm"><strong>{t('webhooks.adminOnly')}:</strong> {t('webhooks.adminOnlyDescription')}</p>
        </div>
      )}

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error.message}
          <Button variant="link" size="sm" onClick={reload}>Retry</Button>
        </div>
      )}

      {isLoading && !error && (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      )}

      {!isLoading && !error && (
        <WebhooksTable webhooks={webhooks} onEdit={handleEdit} onDelete={handleDelete} onToggleEnabled={handleToggleEnabled} canEdit={canEdit} />
      )}

      {dialogOpen && (
        <WebhookDialog mode={dialogMode} open={dialogOpen} initialData={selectedWebhook} onSubmit={handleSubmit} onCancel={() => { setDialogOpen(false); setSelectedWebhookId(null) }} />
      )}

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('webhooks.deleteTitle')}</AlertDialogTitle>
            <AlertDialogDescription>{t('webhooks.deleteMessage')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => { setDeleteConfirmOpen(false); setWebhookToDelete(undefined) }}>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} variant="destructive">{t('common.delete')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
