/**
 * Node Management Page
 *
 * Provides full CRUD operations for managing monitoring nodes.
 * Allows administrators to create, view, edit, and delete nodes.
 */

import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { fetchNodes, createNode, updateNode, deleteNode } from '../api/nodes'
import { NodeTable } from '../components/nodes/NodeTable'
import { NodeDialog } from '../components/nodes/NodeDialog'
import type { NodeDTO, CreateNodeRequest, UpdateNodeRequest } from '../api/types'

export default function NodeManagementPage() {
  const navigate = useNavigate()
  const { user, logout: storeLogout, clearAuth } = useAuthStore()
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
      setNodes(response.data)
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load nodes:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleLogout = async () => {
    try {
      await storeLogout()
      clearAuth()
      navigate('/login')
    } catch (error) {
      console.error('Logout failed:', error)
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
      // Refresh list
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
      // Refresh list
      await loadNodes()
    } catch (error) {
      console.error('Failed to submit node:', error)
      throw error
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold text-gray-900">Node Pulse</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">
                Welcome, {user?.username || 'Guest'}
              </span>
              <button
                type="button"
                onClick={handleLogout}
                className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors duration-150"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <nav className="mb-4 text-sm">
          <ol className="flex items-center space-x-2">
            <li>
              <a
                href="/dashboard"
                className="text-blue-600 hover:text-blue-800"
              >
                Dashboard
              </a>
            </li>
            <li className="text-gray-400">/</li>
            <li className="text-gray-700 font-medium">Node Management</li>
          </ol>
        </nav>

        {/* Page Header */}
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h2 className="text-3xl font-bold text-gray-900">Node Management</h2>
            <p className="mt-2 text-gray-600">
              Manage monitoring nodes. Add, edit, or remove nodes from your monitoring infrastructure.
            </p>
          </div>
          {canEdit && (
            <button
              type="button"
              onClick={handleCreate}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors duration-150"
            >
              Add Node
            </button>
          )}
        </div>

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
            <div className="flex items-center">
              <svg
                className="w-5 h-5 text-red-600 mr-2"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <p className="text-red-800">{error.message}</p>
              <button
                onClick={() => loadNodes()}
                className="ml-auto px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700 transition-colors text-sm"
              >
                Retry
              </button>
            </div>
          </div>
        )}

        {/* Nodes Table */}
        <NodeTable
          nodes={nodes}
          isLoading={isLoading}
          canEdit={canEdit}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />

        {/* Create/Edit Dialog */}
        {dialogOpen && (
          <NodeDialog
            mode={dialogMode}
            node={selectedNode}
            onSubmit={handleSubmit}
            onCancel={() => setDialogOpen(false)}
            loading={isSubmitting}
          />
        )}

        {/* Delete Confirmation Dialog */}
        {deleteConfirmOpen && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                Confirm Delete
              </h3>
              <p className="text-gray-600 mb-6">
                Are you sure you want to delete this node? This action cannot be undone.
              </p>
              <div className="flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => {
                    setDeleteConfirmOpen(false)
                    setNodeToDelete(undefined)
                  }}
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-gray-200 text-gray-800 rounded-md hover:bg-gray-300 transition-colors disabled:bg-gray-100 disabled:cursor-not-allowed"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={confirmDelete}
                  disabled={isSubmitting}
                  className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors disabled:bg-red-300 disabled:cursor-not-allowed"
                >
                  {isSubmitting ? 'Deleting...' : 'Delete'}
                </button>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
