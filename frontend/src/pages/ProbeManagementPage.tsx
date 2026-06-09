import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageContainer, ErrorBanner, LoadingSpinner, ConfirmDialog } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { fetchProbes, createProbe, updateProbe, deleteProbe } from '../api/probes'
import { fetchNodes } from '../api/nodes'
import type { ProbeDTO, CreateProbeRequest, UpdateProbeRequest } from '../api/probes'
import type { NodeDTO } from '../api/types'

interface DialogState {
  mode: 'create' | 'edit'
  probe?: ProbeDTO
}

const emptyForm: CreateProbeRequest = {
  node_id: '',
  type: 'TCP',
  target: '',
  port: 443,
  interval_seconds: 60,
  count: 3,
  timeout_seconds: 5,
}

export default function ProbeManagementPage() {
  const { t } = useTranslation()
  const [probes, setProbes] = useState<ProbeDTO[]>([])
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [nodeFilter, setNodeFilter] = useState<string>('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogState, setDialogState] = useState<DialogState>({ mode: 'create' })
  const [dialogLoading, setDialogLoading] = useState(false)
  const [dialogError, setDialogError] = useState<string | null>(null)
  const [form, setForm] = useState<CreateProbeRequest>(emptyForm)
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; id?: string }>({ open: false })
  const [deleteLoading, setDeleteLoading] = useState(false)

  const loadData = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const [probesRes, nodesRes] = await Promise.all([
        fetchProbes(nodeFilter || undefined),
        fetchNodes(),
      ])
      setProbes(probesRes.data.probes ?? [])
      setNodes(nodesRes.data.nodes ?? [])
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [nodeFilter])

  useEffect(() => {
    void loadData()
  }, [loadData])

  const openCreate = () => {
    setDialogState({ mode: 'create' })
    setForm({ ...emptyForm, node_id: nodeFilter || (nodes[0]?.id ?? '') })
    setDialogError(null)
    setDialogOpen(true)
  }

  const openEdit = (probe: ProbeDTO) => {
    setDialogState({ mode: 'edit', probe })
    setForm({
      node_id: probe.node_id,
      type: probe.type,
      target: probe.target,
      port: probe.port,
      interval_seconds: probe.interval_seconds,
      count: probe.count,
      timeout_seconds: probe.timeout_seconds,
    })
    setDialogError(null)
    setDialogOpen(true)
  }

  const handleDialogSubmit = async () => {
    setDialogLoading(true)
    setDialogError(null)
    try {
      if (dialogState.mode === 'create') {
        if (!form.node_id || !form.target) {
          setDialogError(t('probes.targetRequired'))
          setDialogLoading(false)
          return
        }
        await createProbe(form)
      } else if (dialogState.probe) {
        const req: UpdateProbeRequest = {}
        if (form.type !== dialogState.probe.type) req.type = form.type
        if (form.target !== dialogState.probe.target) req.target = form.target
        if (form.port !== dialogState.probe.port) req.port = form.port
        if (form.interval_seconds !== dialogState.probe.interval_seconds) req.interval_seconds = form.interval_seconds
        if (form.count !== dialogState.probe.count) req.count = form.count
        if (form.timeout_seconds !== dialogState.probe.timeout_seconds) req.timeout_seconds = form.timeout_seconds
        await updateProbe(dialogState.probe.id, req)
      }
      setDialogOpen(false)
      await loadData()
    } catch (err) {
      setDialogError(err instanceof Error ? err.message : String(err))
    } finally {
      setDialogLoading(false)
    }
  }

  const handleDelete = async () => {
    if (!deleteConfirm.id) return
    setDeleteLoading(true)
    try {
      await deleteProbe(deleteConfirm.id)
      await loadData()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setDeleteLoading(false)
      setDeleteConfirm({ open: false })
    }
  }

  const nodeName = (nodeId: string) => nodes.find((n) => n.id === nodeId)?.name ?? nodeId

  return (
    <PageContainer>
      <PageHeader
        title={t('probes.title')}
        subtitle={t('probes.subtitle')}
        actions={
          <button
            type="button"
            onClick={openCreate}
            className="px-4 py-2 bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white text-sm font-medium rounded-lg"
          >
            {t('probes.addProbe')}
          </button>
        }
      />

      {error && <ErrorBanner error={new Error(error)} onRetry={loadData} className="mb-4" />}

      {/* Node Filter */}
      <div className="mb-4">
        <select
          value={nodeFilter}
          onChange={(e) => setNodeFilter(e.target.value)}
          className="rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
        >
          <option value="">{t('probes.allNodes')}</option>
          {nodes.map((n) => (
            <option key={n.id} value={n.id}>{n.name}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><LoadingSpinner /></div>
      ) : (
        <div className="rounded-lg border shadow-sm overflow-hidden bg-[var(--color-bg-surface)] border-[var(--color-border)]">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-[var(--color-border)]">
              <thead className="bg-[var(--color-bg-muted)]">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-[var(--color-text-muted)]">{t('probes.node')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-[var(--color-text-muted)]">{t('probes.type')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-[var(--color-text-muted)]">{t('probes.target')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-[var(--color-text-muted)]">{t('probes.port')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-[var(--color-text-muted)]">{t('probes.interval')}</th>
                  <th className="px-4 py-3 text-right text-xs font-medium uppercase text-[var(--color-text-muted)]">{t('probes.actions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[var(--color-border)]">
                {probes.map((p) => (
                  <tr key={p.id} className="hover:bg-[var(--color-hover-overlay)]">
                    <td className="px-4 py-3 text-sm text-[var(--color-text-primary)]">{nodeName(p.node_id)}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-[var(--color-brand-muted)] text-[var(--color-brand)]">
                        {p.type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-[var(--color-text-secondary)] font-mono">{p.target}</td>
                    <td className="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{p.port}</td>
                    <td className="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{p.interval_seconds}s</td>
                    <td className="px-4 py-3 text-right text-sm space-x-2">
                      <button type="button" onClick={() => openEdit(p)} className="text-[var(--color-brand)] hover:opacity-80">{t('settings.edit')}</button>
                      <button type="button" onClick={() => setDeleteConfirm({ open: true, id: p.id })} className="text-[var(--color-critical)] hover:opacity-80">{t('common.delete')}</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {probes.length === 0 && (
            <div className="py-12 text-center text-sm text-[var(--color-text-secondary)]">{t('probes.noProbes')}</div>
          )}
        </div>
      )}

      {/* Create/Edit Dialog */}
      {dialogOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <div className="fixed inset-0 bg-black/50" onClick={() => setDialogOpen(false)} />
            <div className="relative w-full max-w-md rounded-lg bg-[var(--color-bg-surface)] border border-[var(--color-border)] shadow-xl p-6">
              <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
                {dialogState.mode === 'create' ? t('probes.addProbe') : t('probes.editProbe')}
              </h3>
              {dialogError && (
                <div className="mb-4 rounded-lg bg-[var(--color-critical-bg)] text-[var(--color-critical)] px-4 py-2 text-sm">{dialogError}</div>
              )}
              <div className="space-y-3">
                {dialogState.mode === 'create' && (
                  <div>
                    <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.node')} <span className="text-[var(--color-critical)]">*</span></label>
                    <select
                      value={form.node_id}
                      onChange={(e) => setForm({ ...form, node_id: e.target.value })}
                      className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm"
                    >
                      <option value="">—</option>
                      {nodes.map((n) => <option key={n.id} value={n.id}>{n.name}</option>)}
                    </select>
                  </div>
                )}
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.type')}</label>
                    <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value as 'TCP' | 'UDP' })} className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm">
                      <option value="TCP">TCP</option>
                      <option value="UDP">UDP</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.port')}</label>
                    <input type="number" min={1} max={65535} value={form.port} onChange={(e) => setForm({ ...form, port: Number(e.target.value) })} className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm" />
                  </div>
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.target')} <span className="text-[var(--color-critical)]">*</span></label>
                  <input type="text" value={form.target} onChange={(e) => setForm({ ...form, target: e.target.value })} placeholder="IP or domain" className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm" />
                </div>
                <div className="grid grid-cols-3 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.interval')} (s)</label>
                    <input type="number" min={60} max={300} value={form.interval_seconds} onChange={(e) => setForm({ ...form, interval_seconds: Number(e.target.value) })} className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.count')}</label>
                    <input type="number" min={1} max={100} value={form.count} onChange={(e) => setForm({ ...form, count: Number(e.target.value) })} className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm" />
                  </div>
                  <div>
                    <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">{t('probes.timeout')} (s)</label>
                    <input type="number" min={1} max={30} value={form.timeout_seconds} onChange={(e) => setForm({ ...form, timeout_seconds: Number(e.target.value) })} className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm" />
                  </div>
                </div>
              </div>
              <div className="mt-6 flex justify-end gap-3">
                <button type="button" onClick={() => setDialogOpen(false)} className="px-4 py-2 text-sm font-medium rounded-lg border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]">{t('common.cancel')}</button>
                <button type="button" onClick={() => void handleDialogSubmit()} disabled={dialogLoading} className="px-4 py-2 text-sm font-medium rounded-lg bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white disabled:opacity-50">
                  {dialogLoading ? t('common.saving') : t('common.save')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog
        open={deleteConfirm.open}
        title={t('probes.confirmDelete')}
        message={t('probes.confirmDeleteMessage')}
        confirmText={t('common.delete')}
        onConfirm={() => void handleDelete()}
        onCancel={() => setDeleteConfirm({ open: false })}
        loading={deleteLoading}
        variant="danger"
      />
    </PageContainer>
  )
}
