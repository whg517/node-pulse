/**
 * Alert Rules Page
 *
 * Configure alert thresholds and levels for network monitoring.
 * Uses standardized layout components for consistent UI.
 */

import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAlertsStore } from '../stores/alertsStore'
import { useAuthStore } from '../stores/authStore'
import { fetchNodes } from '../api/nodes'
import type { AlertRule } from '../stores/types'
import type { NodeDTO } from '../api/types'
import { PageContainer, ErrorBanner, ConfirmDialog, ActionButton, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { AlertRulesTable } from '../components/alerts/AlertRulesTable'
import { AlertRuleDialog } from '../components/alerts/AlertRuleDialog'

export default function AlertRulesPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
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
      await fetchAlertRules()
    } catch (error) {
      console.error('Failed to submit alert rule:', error)
      throw error
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('alerts.rulesTitle')}
        subtitle={t('alerts.rulesDescription')}
        showBreadcrumb
        actions={
          canEdit && (
            <ActionButton onClick={handleCreate}>
              {t('alerts.createRule')}
            </ActionButton>
          )
        }
      />

      {/* Error State */}
      {error && (
        <ErrorBanner
          error={error}
          onRetry={loadData}
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
      <ConfirmDialog
        open={deleteConfirmOpen}
        title={t('alerts.deleteTitle')}
        message={t('alerts.deleteMessage')}
        confirmText={t('common.delete')}
        onConfirm={confirmDelete}
        onCancel={() => {
          setDeleteConfirmOpen(false)
          setRuleToDelete(undefined)
        }}
        loading={false}
        variant="danger"
      />
    </PageContainer>
  )
}
