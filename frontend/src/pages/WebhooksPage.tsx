import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { previewWebhookPayload, type CreateWebhookRequest, type WebhookEventFormat } from '@/api/webhooks'
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
  const { webhooks, isLoading, error, reload, createWebhook, updateWebhook, deleteWebhook, testWebhook, toggleWebhookEnabled, getWebhookById } = useWebhooks()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedWebhookId, setSelectedWebhookId] = useState<string | null>(null)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [webhookToDelete, setWebhookToDelete] = useState<string>()
  const [testingWebhookId, setTestingWebhookId] = useState<string | null>(null)
  const [testNotice, setTestNotice] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

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

  const handleTestWebhook = (id: string) => {
    void (async () => {
      setTestingWebhookId(id)
      setTestNotice(null)
      try {
        const response = await testWebhook(id)
        setTestNotice({ type: 'success', message: response.message || t('webhooks.testSuccess') })
      } catch (error) {
        const message = error instanceof Error ? error.message : t('webhooks.testFailed')
        setTestNotice({ type: 'error', message })
      } finally {
        setTestingWebhookId(null)
      }
    })()
  }

  const handleSubmit = async (data: CreateWebhookRequest) => {
    if (dialogMode === 'create') { await createWebhook(data) }
    else if (selectedWebhook) { await updateWebhook(selectedWebhook.id, data) }
    else throw new Error('No webhook selected')
    setDialogOpen(false)
    setSelectedWebhookId(null)
  }

  const handlePreview = async (eventFormat: WebhookEventFormat) => {
    const response = await previewWebhookPayload({ event_format: eventFormat })
    return response.data.payload
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
          <Button variant="link" size="sm" onClick={reload}>{t('common.retry')}</Button>
        </div>
      )}

      {testNotice && (
        <div className={`rounded-md px-4 py-3 text-sm ${
          testNotice.type === 'success'
            ? 'bg-healthy-bg text-healthy-text'
            : 'bg-destructive/10 text-destructive'
        }`}>
          {testNotice.message}
        </div>
      )}

      {isLoading && !error && (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      )}

      {!isLoading && !error && (
        <WebhooksTable
          webhooks={webhooks}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onTest={handleTestWebhook}
          onToggleEnabled={handleToggleEnabled}
          testingWebhookId={testingWebhookId}
          canEdit={canEdit}
        />
      )}

      {dialogOpen && (
        <WebhookDialog mode={dialogMode} open={dialogOpen} initialData={selectedWebhook} onSubmit={handleSubmit} onPreview={handlePreview} onCancel={() => { setDialogOpen(false); setSelectedWebhookId(null) }} />
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
