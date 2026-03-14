/**
 * Webhooks Page
 *
 * Configure webhook endpoints for alert notifications.
 * Uses standardized layout components for consistent UI.
 */

import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useWebhooksStore } from '../stores/webhooksStore'
import { useAuthStore } from '../stores/authStore'
import { PageContainer, ErrorBanner, ConfirmDialog, ActionButton, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { WebhooksTable } from '../components/webhooks/WebhooksTable'
import { WebhookDialog } from '../components/webhooks/WebhookDialog'

export default function WebhooksPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const webhooksStore = useWebhooksStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedWebhook, setSelectedWebhook] = useState<typeof webhooksStore.webhooks[0] | undefined>(undefined)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [webhookToDelete, setWebhookToDelete] = useState<string | undefined>(undefined)

  // Check if user can edit (admin only)
  const canEdit = user?.role === 'admin'

  useEffect(() => {
    loadWebhooks()
  }, [])

  const loadWebhooks = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await webhooksStore.fetchWebhooks()
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load webhooks:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreate = () => {
    setDialogMode('create')
    setSelectedWebhook(undefined)
    setDialogOpen(true)
  }

  const handleEdit = (id: string) => {
    const webhook = webhooksStore.webhooks.find((w) => w.id === id)
    if (webhook) {
      setDialogMode('edit')
      setSelectedWebhook(webhook)
      setDialogOpen(true)
    }
  }

  const handleDelete = (id: string) => {
    setWebhookToDelete(id)
    setDeleteConfirmOpen(true)
  }

  const confirmDelete = async () => {
    if (!webhookToDelete) return

    try {
      webhooksStore.removeWebhook(webhookToDelete)
      setDeleteConfirmOpen(false)
      setWebhookToDelete(undefined)
      await webhooksStore.fetchWebhooks()
    } catch (error) {
      console.error('Failed to delete webhook:', error)
    }
  }

  const handleToggleEnabled = (id: string, enabled: boolean) => {
    try {
      webhooksStore.updateWebhook(id, { enabled })
    } catch (error) {
      console.error('Failed to toggle webhook:', error)
    }
  }

  const handleSubmit = async (data: any) => {
    try {
      if (dialogMode === 'create') {
        webhooksStore.addWebhook(data as any)
      } else {
        webhooksStore.updateWebhook(selectedWebhook!.id, data as any)
      }
      setDialogOpen(false)
      await webhooksStore.fetchWebhooks()
    } catch (error) {
      console.error('Failed to submit webhook:', error)
      throw error
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('webhooks.title')}
        subtitle={t('webhooks.description')}
        showBreadcrumb
        actions={
          canEdit && (
            <ActionButton onClick={handleCreate}>
              {t('webhooks.addWebhook')}
            </ActionButton>
          )
        }
      />

      {/* Access Warning for Non-Admin Users */}
      {!canEdit && (
        <div data-testid="access-warning" className="mb-6 rounded-md border-l-4 border-amber-400 bg-amber-50 p-4 dark:bg-amber-900/20">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-amber-400" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-amber-700 dark:text-amber-300">
                <strong>{t('webhooks.adminOnly')}:</strong> {t('webhooks.adminOnlyDescription')}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <ErrorBanner
          error={error}
          onRetry={loadWebhooks}
          className="mb-6"
        />
      )}

      {/* Loading State */}
      {isLoading && !error && (
        <div className="py-12">
          <LoadingSpinner />
        </div>
      )}

      {/* Content */}
      {!isLoading && !error && (
        <WebhooksTable
          webhooks={webhooksStore.webhooks}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onToggleEnabled={handleToggleEnabled}
          canEdit={canEdit}
        />
      )}

      {/* Create/Edit Dialog */}
      {dialogOpen && (
        <WebhookDialog
          mode={dialogMode}
          initialData={selectedWebhook}
          onSubmit={handleSubmit}
          onCancel={() => setDialogOpen(false)}
        />
      )}

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        title={t('webhooks.deleteTitle')}
        message={t('webhooks.deleteMessage')}
        confirmText={t('common.delete')}
        onConfirm={confirmDelete}
        onCancel={() => {
          setDeleteConfirmOpen(false)
          setWebhookToDelete(undefined)
        }}
        loading={false}
        variant="danger"
      />
    </PageContainer>
  )
}
