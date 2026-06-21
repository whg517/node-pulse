import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/authStore'
import { fetchNodes, createNode, updateNode, deleteNode } from '@/api/nodes'
import { PageHeader } from '@/components/layout/PageHeader'
import { NodeTable } from '@/components/nodes/NodeTable'
import { NodeDialog } from '@/components/nodes/NodeDialog'
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
import type { NodeDTO, CreateNodeRequest, UpdateNodeRequest } from '@/api/types'

export default function NodeManagementPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedNode, setSelectedNode] = useState<NodeDTO | undefined>()
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [nodeToDelete, setNodeToDelete] = useState<string>()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const canEdit = user?.role === 'admin'

  useEffect(() => { loadNodes() }, [])

  const loadNodes = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      setNodes(response.data.nodes || [])
    } catch (err) {
      setError(err as Error)
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
    } catch { /* handled by UI */ } finally {
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
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('nodes.management')}
        subtitle={t('nodes.managementDescription')}
        actions={canEdit ? <Button onClick={handleCreate}>{t('nodes.addNode')}</Button> : undefined}
      />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error.message}
          <Button variant="link" size="sm" onClick={loadNodes} className="ml-2">{t('common.retry')}</Button>
        </div>
      )}

      {isLoading && !error && (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      )}

      {!isLoading && !error && (
        <NodeTable nodes={nodes} isLoading={false} canEdit={canEdit} onEdit={handleEdit} onDelete={handleDelete} />
      )}

      {dialogOpen && (
        <NodeDialog mode={dialogMode} node={selectedNode} open={dialogOpen} onSubmit={handleSubmit} onCancel={() => setDialogOpen(false)} />
      )}

      <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('nodes.deleteTitle')}</AlertDialogTitle>
            <AlertDialogDescription>{t('nodes.deleteMessage')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => { setDeleteConfirmOpen(false); setNodeToDelete(undefined) }}>
              {t('common.cancel')}
            </AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} disabled={isSubmitting} variant="destructive">
              {t('common.delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
