/**
 * Node Management Page
 *
 * Provides full CRUD operations for managing monitoring nodes.
 * Uses standardized layout components for consistent UI.
 */

import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../stores/authStore'
import { fetchNodes, createNode, updateNode, deleteNode } from '../api/nodes'
import { PageContainer, ErrorBanner, ConfirmDialog, ActionButton, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { NodeTable } from '../components/nodes/NodeTable'
import { NodeDialog } from '../components/nodes/NodeDialog'
import type { NodeDTO, CreateNodeRequest, UpdateNodeRequest } from '../api/types'

export default function NodeManagementPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedNode, setSelectedNode] = useState<NodeDTO | undefined>(undefined)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [nodeToDelete, setNodeToDelete] = useState<string | undefined>(undefined)
  const [isSubmitting, setIsSubmitting] = useState(false)

  // Check if user can edit (admin only)
  const canEdit = user?.role === 'admin'

  useEffect(() => {
    loadNodes()
  }, [])

  const loadNodes = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      setNodes(response.data.nodes || [])
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load nodes:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCreate = () => {
    setDialogMode('create')
    setSelectedNode(undefined)
    setDialogOpen(true)
  }

  const handleEdit = (id: string) => {
    const node = nodes.find((n) => n.id === id)
    if (node) {
      setDialogMode('edit')
      setSelectedNode(node)
      setDialogOpen(true)
    }
  }

  const handleDelete = (id: string) => {
    setNodeToDelete(id)
    setDeleteConfirmOpen(true)
  }

  const confirmDelete = async () => {
    if (!nodeToDelete) return

    setIsSubmitting(true)
    try {
      await deleteNode(nodeToDelete)
      setDeleteConfirmOpen(false)
      setNodeToDelete(undefined)
      await loadNodes()
    } catch (error) {
      console.error('Failed to delete node:', error)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (data: CreateNodeRequest | UpdateNodeRequest) => {
    setIsSubmitting(true)
    try {
      if (dialogMode === 'create') {
        await createNode(data as CreateNodeRequest)
      } else {
        await updateNode(selectedNode!.id, data as UpdateNodeRequest)
      }
      setDialogOpen(false)
      await loadNodes()
    } catch (error) {
      console.error('Failed to submit node:', error)
      throw error
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('nodes.management')}
        subtitle={t('nodes.managementDescription')}
        actions={
          canEdit && (
            <ActionButton onClick={handleCreate}>
              {t('nodes.addNode')}
            </ActionButton>
          )
        }
      />

      {/* Error State */}
      {error && (
        <ErrorBanner
          error={error}
          onRetry={loadNodes}
          className="mb-6"
        />
      )}

      {/* Loading State */}
      {isLoading && !error && (
        <div className="py-12">
          <LoadingSpinner />
        </div>
      )}

      {/* Nodes Table */}
      {!isLoading && !error && (
        <NodeTable
          nodes={nodes}
          isLoading={false}
          canEdit={canEdit}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      )}

      {/* Create/Edit Dialog */}
      {dialogOpen && (
        <NodeDialog
          mode={dialogMode}
          node={selectedNode}
          onSubmit={handleSubmit}
          onCancel={() => setDialogOpen(false)}
        />
      )}

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        title={t('nodes.deleteTitle')}
        message={t('nodes.deleteMessage')}
        confirmText={t('common.delete')}
        onConfirm={confirmDelete}
        onCancel={() => {
          setDeleteConfirmOpen(false)
          setNodeToDelete(undefined)
        }}
        loading={isSubmitting}
        variant="danger"
      />
    </PageContainer>
  )
}
