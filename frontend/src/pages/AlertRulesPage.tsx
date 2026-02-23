import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAlertsStore } from '../stores/alertsStore'
import { useAuthStore } from '../stores/authStore'
import { fetchNodes } from '../api/nodes'
import type { AlertRule } from '../stores/types'
import type { NodeDTO } from '../api/types'
import { AlertRulesTable } from '../components/alerts/AlertRulesTable'
import { AlertRuleDialog } from '../components/alerts/AlertRuleDialog'

export default function AlertRulesPage() {
  const navigate = useNavigate()
  const { user, logout: storeLogout, clearAuth } = useAuthStore()
  const { alertRules, fetchAlertRules, addAlertRule, updateAlertRule, removeAlertRule } = useAlertsStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [nodes, setNodes] = useState<NodeDTO[]>([])

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedRule, setSelectedRule] = useState<AlertRule | undefined>(undefined)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [ruleToDelete, setRuleToDelete] = useState<string | undefined>(undefined)

  // Check if user can edit (admin or operator)
  const canEdit = user?.role === 'admin' || user?.role === 'operator'

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setIsLoading(true)
    setError(null)
    try {
      // Fetch alert rules and nodes in parallel
      await Promise.all([fetchAlertRules(), loadNodes()])
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load data:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const loadNodes = async () => {
    try {
      const response = await fetchNodes()
      setNodes(response.data.nodes || [])
    } catch (err) {
      console.error('Failed to load nodes:', err)
      throw err
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
    setSelectedRule(undefined)
    setDialogOpen(true)
  }

  const handleEdit = (id: string) => {
    const rule = alertRules.find((r) => r.id === id)
    if (rule) {
      setDialogMode('edit')
      setSelectedRule(rule)
      setDialogOpen(true)
    }
  }

  const handleDelete = (id: string) => {
    setRuleToDelete(id)
    setDeleteConfirmOpen(true)
  }

  const confirmDelete = async () => {
    if (!ruleToDelete) return

    try {
      await removeAlertRule(ruleToDelete)
      setDeleteConfirmOpen(false)
      setRuleToDelete(undefined)
      // Refresh list
      await fetchAlertRules()
    } catch (error) {
      console.error('Failed to delete alert rule:', error)
    }
  }

  const handleToggleEnabled = async (id: string, enabled: boolean) => {
    try {
      await updateAlertRule(id, { enabled })
    } catch (error) {
      console.error('Failed to toggle alert rule:', error)
    }
  }

  const handleSubmit = async (data: any) => {
    try {
      if (dialogMode === 'create') {
        await addAlertRule(data)
      } else {
        await updateAlertRule(selectedRule!.id, data)
      }
      setDialogOpen(false)
      // Refresh list
      await fetchAlertRules()
    } catch (error) {
      console.error('Failed to submit alert rule:', error)
      throw error
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
        {/* Page Header */}
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h2 className="text-3xl font-bold text-gray-900">Alert Rules</h2>
            <p className="mt-2 text-gray-600">
              Configure alert thresholds and levels for network monitoring
            </p>
          </div>
          {canEdit && (
            <button
              type="button"
              onClick={handleCreate}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors duration-150"
            >
              Create Rule
            </button>
          )}
        </div>

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-red-50 border-l-4 border-red-400 p-4 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  aria-hidden="true"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">{error.message}</p>
              </div>
              <div className="ml-auto pl-3">
                <div className="-mx-1.5 -my-1.5">
                  <button
                    onClick={() => loadData()}
                    className="inline-flex bg-red-50 rounded-md p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600"
                  >
                    <svg
                      className="h-5 w-5"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                      aria-hidden="true"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                      />
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && !error && (
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
          </div>
        )}

        {/* Content */}
        {!isLoading && !error && (
          <AlertRulesTable
            rules={alertRules}
            nodes={nodes}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onToggleEnabled={handleToggleEnabled}
            canEdit={canEdit}
          />
        )}

        {/* Create/Edit Dialog */}
        {dialogOpen && (
          <AlertRuleDialog
            mode={dialogMode}
            initialData={selectedRule}
            nodes={nodes}
            onSubmit={handleSubmit}
            onCancel={() => setDialogOpen(false)}
          />
        )}

        {/* Delete Confirmation Dialog */}
        {deleteConfirmOpen && (
          <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
              <div className="mt-3 text-center">
                <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
                  <svg
                    className="h-6 w-6 text-red-600"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                    />
                  </svg>
                </div>
                <h3 className="text-lg leading-6 font-medium text-gray-900 mt-4">
                  Delete Alert Rule
                </h3>
                <div className="mt-2">
                  <p className="text-sm text-gray-500">
                    Are you sure you want to delete this alert rule? This action cannot be undone.
                  </p>
                </div>
              </div>
              <div className="flex justify-end gap-3 mt-6">
                <button
                  type="button"
                  onClick={() => {
                    setDeleteConfirmOpen(false)
                    setRuleToDelete(undefined)
                  }}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={confirmDelete}
                  className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
                >
                  Delete
                </button>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}
