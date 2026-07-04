import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAlertsStore } from '@/stores/alertsStore'
import { useAuthStore } from '@/stores/authStore'
import { listRoutingRules, createRoutingRule, updateRoutingRule, deleteRoutingRule, type AlertRoutingRuleDTO } from '@/api/alertRouting'
import { fetchNodes } from '@/api/nodes'
import type { AlertRule } from '@/stores/types'
import type { NodeDTO, CreateAlertRuleRequest } from '@/api/types'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { AlertRulesTable } from '@/components/alerts/AlertRulesTable'
import { AlertRuleDialog } from '@/components/alerts/AlertRuleDialog'

type TabId = 'rules' | 'routing'

export default function AlertRulesPage() {
  const { t } = useTranslation()
  const { user } = useAuthStore()
  const { alertRules, fetchAlertRules, addAlertRule, updateAlertRule, removeAlertRule } = useAlertsStore()
  // Routing rules now come from the server (ADR-002), enforced at dispatch time.
  const [serverRules, setServerRules] = useState<AlertRoutingRuleDTO[]>([])

  const loadRoutingRules = useCallback(async () => {
    try {
      const res = await listRoutingRules()
      setServerRules(res.data?.rules || [])
    } catch {
      // best-effort
    }
  }, [])

  useEffect(() => { void loadRoutingRules() }, [loadRoutingRules])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
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

  const loadData = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const [, nodesRes] = await Promise.all([fetchAlertRules(), fetchNodes()])
      setNodes(nodesRes.data.nodes || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [fetchAlertRules])

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

  const handleCreateRoutingRule = async () => {
    // The legacy local form used an action.target for the webhook id; the server
    // API uses a flat webhook_id. Persist via the server (ADR-002).
    await createRoutingRule({
      name: routingForm.name,
      enabled: routingForm.enabled,
      metric: routingForm.metric || undefined,
      severities: routingForm.severity ? [routingForm.severity] : undefined,
      webhook_id: routingForm.actionTarget,
    }).catch(() => { /* best-effort */ })
    await loadRoutingRules()
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
    <div className="space-y-6">
      <PageHeader
        title={t('alerts.rulesTitle')}
        subtitle={activeTab === 'rules' ? t('alerts.rulesDescription') : t('alerts.routingDescription')}
        actions={
          canEdit && (
            activeTab === 'rules' ? (
              <Button onClick={handleCreate}>{t('alerts.createRule')}</Button>
            ) : (
              <Button onClick={() => setShowRoutingDialog(true)}>{t('alerts.createRoutingRule')}</Button>
            )
          )
        }
      />

      {/* Tabs */}
      <div className="border-b">
        <nav className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              type="button"
              onClick={() => setActiveTab(tab.id)}
              className={`pb-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-primary text-primary'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
          <Button variant="link" size="sm" onClick={loadData}>{t('common.retry')}</Button>
        </div>
      )}

      {isLoading && !error && (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      )}

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

      {!isLoading && !error && activeTab === 'routing' && (
        <div>
          {serverRules.length === 0 ? (
            <Card>
              <CardContent className="py-12 text-center">
                <p className="text-sm text-muted-foreground">{t('alerts.noRoutingRules')}</p>
                <p className="text-xs text-muted-foreground mt-1">{t('alerts.noRoutingRulesHint')}</p>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="p-0 divide-y">
                {serverRules.map((rule) => (
                  <div key={rule.id} className="px-4 py-3 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Switch
                        checked={rule.enabled}
                        onCheckedChange={async (checked) => {
                          await updateRoutingRule(rule.id, {
                            name: rule.name, enabled: checked, metric: rule.metric,
                            severities: rule.severities, node_id: rule.node_id, webhook_id: rule.webhook_id,
                          }).catch(() => {})
                          await loadRoutingRules()
                        }}
                      />
                      <div>
                        <p className="text-sm font-medium">{rule.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {rule.metric && `${t('alerts.whenMetric')}: ${rule.metric}`}
                          {rule.severities && rule.severities.length > 0 && ` · ${t('alerts.whenSeverity')}: ${rule.severities.join('/')}`}
                          {' · webhook → '}{rule.webhook_id}
                        </p>
                      </div>
                    </div>
                    <Button variant="link" size="sm" className="text-destructive" onClick={async () => { await deleteRoutingRule(rule.id).catch(() => {}); await loadRoutingRules() }}>
                      {t('common.delete')}
                    </Button>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {dialogOpen && (
        <AlertRuleDialog
          mode={dialogMode}
          initialData={selectedRule}
          nodes={nodes}
          onSubmit={handleSubmit}
          onCancel={() => setDialogOpen(false)}
          open={dialogOpen}
        />
      )}

      {/* Routing Rule Dialog */}
      <Dialog open={showRoutingDialog} onOpenChange={(o) => { if (!o) setShowRoutingDialog(false) }}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('alerts.createRoutingRule')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="routing-name">{t('common.name')}</Label>
              <Input
                id="routing-name"
                value={routingForm.name}
                onChange={(e) => setRoutingForm({ ...routingForm, name: e.target.value })}
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="routing-metric">{t('alerts.whenMetric')}</Label>
                <select
                  id="routing-metric"
                  value={routingForm.metric}
                  onChange={(e) => setRoutingForm({ ...routingForm, metric: e.target.value })}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="">-</option>
                  <option value="latency">{t('metrics.latency')}</option>
                  <option value="packet_loss">{t('metrics.packetLoss')}</option>
                  <option value="jitter">{t('metrics.jitter')}</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="routing-severity">{t('alerts.whenSeverity')}</Label>
                <select
                  id="routing-severity"
                  value={routingForm.severity}
                  onChange={(e) => setRoutingForm({ ...routingForm, severity: e.target.value })}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="">-</option>
                  <option value="critical">{t('alerts.critical')}</option>
                  <option value="warning">{t('alerts.warning')}</option>
                  <option value="info">{t('alerts.info')}</option>
                </select>
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="routing-action-type">{t('alerts.notifyVia')}</Label>
              <select
                id="routing-action-type"
                value={routingForm.actionType}
                onChange={(e) => setRoutingForm({ ...routingForm, actionType: e.target.value as 'webhook' | 'email' })}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              >
                <option value="webhook">{t('alerts.webhook')}</option>
                <option value="email">{t('alerts.ruleEmail')}</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="routing-action-target">{t('alerts.routeTo')}</Label>
              <Input
                id="routing-action-target"
                value={routingForm.actionTarget}
                onChange={(e) => setRoutingForm({ ...routingForm, actionTarget: e.target.value })}
                placeholder={routingForm.actionType === 'email' ? 'user@example.com' : 'Webhook URL or ID'}
              />
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="routing-enabled"
                checked={routingForm.enabled}
                onCheckedChange={(checked) => setRoutingForm({ ...routingForm, enabled: checked })}
              />
              <Label htmlFor="routing-enabled">{t('status.enabled')}</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowRoutingDialog(false)}>{t('common.cancel')}</Button>
            <Button onClick={handleCreateRoutingRule} disabled={!routingForm.name.trim()}>{t('common.create')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleteConfirmOpen} onOpenChange={(open) => !open && setDeleteConfirmOpen(false)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('alerts.deleteTitle')}</AlertDialogTitle>
            <AlertDialogDescription>{t('alerts.deleteMessage')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => { setDeleteConfirmOpen(false); setRuleToDelete(undefined) }}>
              {t('common.cancel')}
            </AlertDialogCancel>
            <AlertDialogAction onClick={confirmDelete} variant="destructive">
              {t('common.delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
