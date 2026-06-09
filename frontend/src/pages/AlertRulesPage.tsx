import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAlertsStore } from '../stores/alertsStore'
import { useAuthStore } from '../stores/authStore'
import { useSettingsStore, type AlertRoutingRule } from '../stores/settingsStore'
import { fetchNodes } from '../api/nodes'
import type { AlertRule } from '../stores/types'
import type { NodeDTO, CreateAlertRuleRequest } from '../api/types'
import { PageContainer, ErrorBanner, ConfirmDialog, ActionButton, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { AlertRulesTable } from '../components/alerts/AlertRulesTable'
import { AlertRuleDialog } from '../components/alerts/AlertRuleDialog'

type TabId = 'rules' | 'routing'

export default function AlertRulesPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const { alertRules, fetchAlertRules, addAlertRule, updateAlertRule, removeAlertRule } = useAlertsStore()
  const { routingRules, addRoutingRule, updateRoutingRule, deleteRoutingRule } = useSettingsStore()
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [activeTab, setActiveTab] = useState<TabId>('rules')

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [selectedRule, setSelectedRule] = useState<AlertRule | undefined>(undefined)
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
  const [ruleToDelete, setRuleToDelete] = useState<string | undefined>(undefined)

  const [showRoutingDialog, setShowRoutingDialog] = useState(false)
  const [routingForm, setRoutingForm] = useState({
    name: '',
    metric: '',
    severity: '',
    actionType: 'webhook' as 'webhook' | 'email',
    actionTarget: '',
    enabled: true,
  })

  const canEdit = user?.role === 'admin' || user?.role === 'operator'

  const loadNodes = useCallback(async () => {
    try {
      const response = await fetchNodes()
      setNodes(response.data.nodes || [])
    } catch (err) {
      console.error('Failed to load nodes:', err)
      throw err
    }
  }, [])

  const loadData = useCallback(async () => {
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
  }, [fetchAlertRules, loadNodes])

  useEffect(() => {
    void loadData()
  }, [loadData])

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

  const handleSubmit = async (data: CreateAlertRuleRequest) => {
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

  const handleCreateRoutingRule = () => {
    const rule: AlertRoutingRule = {
      id: crypto.randomUUID(),
      name: routingForm.name,
      conditions: {
        metric: routingForm.metric || undefined,
        severity: routingForm.severity || undefined,
      },
      action: {
        type: routingForm.actionType,
        target: routingForm.actionTarget,
      },
      enabled: routingForm.enabled,
    }
    addRoutingRule(rule)
    setShowRoutingDialog(false)
    setRoutingForm({
      name: '',
      metric: '',
      severity: '',
      actionType: 'webhook',
      actionTarget: '',
      enabled: true,
    })
  }

  const tabs: { id: TabId; label: string }[] = [
    { id: 'rules', label: t('alerts.alertRules') },
    { id: 'routing', label: t('alerts.routingRules') },
  ]

  return (
    <PageContainer>
      <PageHeader
        title={t('alerts.rulesTitle')}
        subtitle={activeTab === 'rules' ? t('alerts.rulesDescription') : t('alerts.routingDescription')}
        actions={
          canEdit && (
            activeTab === 'rules' ? (
              <ActionButton onClick={handleCreate}>
                {t('alerts.createRule')}
              </ActionButton>
            ) : (
              <ActionButton onClick={() => setShowRoutingDialog(true)}>
                {t('alerts.createRoutingRule')}
              </ActionButton>
            )
          )
        }
      />

      {/* Tabs */}
      <div className="mb-6 border-b border-[var(--color-border)]">
        <nav className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-[var(--color-brand)] text-[var(--color-brand)]'
                  : 'border-transparent text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

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

      {/* Alert Rules Tab */}
      {!isLoading && !error && activeTab === 'rules' && (
        <AlertRulesTable
          rules={alertRules}
          nodes={nodes}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onToggleEnabled={handleToggleEnabled}
          canEdit={canEdit}
        />
      )}

      {/* Routing Rules Tab */}
      {!isLoading && !error && activeTab === 'routing' && (
        <div>
          {routingRules.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-sm text-[var(--color-text-secondary)]">{t('alerts.noRoutingRules')}</p>
              <p className="text-xs text-[var(--color-text-muted)] mt-1">{t('alerts.noRoutingRulesHint')}</p>
            </div>
          ) : (
            <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)] divide-y divide-[var(--color-border)]">
              {routingRules.map((rule) => (
                <div key={rule.id} className="px-4 py-3 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={rule.enabled}
                      onChange={(e) => updateRoutingRule(rule.id, { enabled: e.target.checked })}
                      className="h-4 w-4 rounded border-[var(--color-input-border)] text-[var(--color-brand)]"
                    />
                    <div>
                      <p className="text-sm font-medium text-[var(--color-text-primary)]">{rule.name}</p>
                      <p className="text-xs text-[var(--color-text-muted)]">
                        {rule.conditions.metric && `${t('alerts.whenMetric')}: ${rule.conditions.metric}`}
                        {rule.conditions.severity && ` · ${t('alerts.whenSeverity')}: ${rule.conditions.severity}`}
                        {' · '}{t('alerts.notifyVia')}: {rule.action.type}
                        {rule.action.target && ` → ${rule.action.target}`}
                      </p>
                    </div>
                  </div>
                  <button
                    type="button"
                    onClick={() => deleteRoutingRule(rule.id)}
                    className="text-xs text-[var(--color-critical)] hover:opacity-80"
                  >
                    {t('common.delete')}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Create/Edit Alert Rule Dialog */}
      {dialogOpen && (
        <AlertRuleDialog
          mode={dialogMode}
          initialData={selectedRule}
          nodes={nodes}
          onSubmit={handleSubmit}
          onCancel={() => setDialogOpen(false)}
        />
      )}

      {/* Routing Rule Dialog */}
      {showRoutingDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setShowRoutingDialog(false)}>
          <div className="w-full max-w-md rounded-lg bg-[var(--color-bg-surface)] shadow-xl p-6" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
              {t('alerts.createRoutingRule')}
            </h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('common.name')}
                </label>
                <input
                  type="text"
                  value={routingForm.name}
                  onChange={(e) => setRoutingForm({ ...routingForm, name: e.target.value })}
                  className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                    {t('alerts.whenMetric')}
                  </label>
                  <select
                    value={routingForm.metric}
                    onChange={(e) => setRoutingForm({ ...routingForm, metric: e.target.value })}
                    className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                  >
                    <option value="">-</option>
                    <option value="latency">{t('metrics.latency')}</option>
                    <option value="packet_loss">{t('metrics.packetLoss')}</option>
                    <option value="jitter">{t('metrics.jitter')}</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                    {t('alerts.whenSeverity')}
                  </label>
                  <select
                    value={routingForm.severity}
                    onChange={(e) => setRoutingForm({ ...routingForm, severity: e.target.value })}
                    className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                  >
                    <option value="">-</option>
                    <option value="critical">{t('alerts.critical')}</option>
                    <option value="warning">{t('alerts.warning')}</option>
                    <option value="info">{t('alerts.info')}</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('alerts.notifyVia')}
                </label>
                <select
                  value={routingForm.actionType}
                  onChange={(e) => setRoutingForm({ ...routingForm, actionType: e.target.value as 'webhook' | 'email' })}
                  className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                >
                  <option value="webhook">{t('alerts.webhook')}</option>
                  <option value="email">{t('alerts.ruleEmail')}</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('alerts.routeTo')}
                </label>
                <input
                  type="text"
                  value={routingForm.actionTarget}
                  onChange={(e) => setRoutingForm({ ...routingForm, actionTarget: e.target.value })}
                  placeholder={routingForm.actionType === 'email' ? 'user@example.com' : 'Webhook URL or ID'}
                  className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={routingForm.enabled}
                  onChange={(e) => setRoutingForm({ ...routingForm, enabled: e.target.checked })}
                  className="h-4 w-4 rounded border-[var(--color-input-border)] text-[var(--color-brand)]"
                />
                <label className="text-sm text-[var(--color-text-secondary)]">{t('status.enabled')}</label>
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                type="button"
                onClick={() => setShowRoutingDialog(false)}
                className="px-4 py-2 text-sm font-medium rounded-lg border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]"
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                onClick={handleCreateRoutingRule}
                disabled={!routingForm.name.trim()}
                className="px-4 py-2 bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white text-sm font-medium rounded-lg disabled:opacity-50"
              >
                {t('common.create')}
              </button>
            </div>
          </div>
        </div>
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
