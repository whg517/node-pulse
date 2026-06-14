import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { fetchProbes, createProbe, updateProbe, deleteProbe } from '@/api/probes'
import { fetchNodes } from '@/api/nodes'
import type { ProbeDTO, CreateProbeRequest, UpdateProbeRequest } from '@/api/probes'
import type { NodeDTO } from '@/api/types'

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
    <div className="space-y-6">
      <PageHeader
        title={t('probes.title')}
        subtitle={t('probes.subtitle')}
        actions={
          <button
            type="button"
            onClick={openCreate}
            className="px-4 py-2 bg-primary hover:bg-primary/85 text-white text-sm font-medium rounded-lg"
          >
            {t('probes.addProbe')}
          </button>
        }
      />

      {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

      {/* Node Filter */}
      <div className="mb-4">
        <select
          value={nodeFilter}
          onChange={(e) => setNodeFilter(e.target.value)}
          className="rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground"
        >
          <option value="">{t('probes.allNodes')}</option>
          {nodes.map((n) => (
            <option key={n.id} value={n.id}>{n.name}</option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      ) : (
        <div className="rounded-lg border shadow-sm overflow-hidden bg-card border-border">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border">
              <thead className="bg-muted">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-muted-foreground">{t('probes.node')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-muted-foreground">{t('probes.type')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-muted-foreground">{t('probes.target')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-muted-foreground">{t('probes.port')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium uppercase text-muted-foreground">{t('probes.interval')}</th>
                  <th className="px-4 py-3 text-right text-xs font-medium uppercase text-muted-foreground">{t('probes.actions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {probes.map((p) => (
                  <tr key={p.id} className="hover:bg-accent/10">
                    <td className="px-4 py-3 text-sm text-foreground">{nodeName(p.node_id)}</td>
                    <td className="px-4 py-3 text-sm">
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary">
                        {p.type}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm text-muted-foreground font-mono">{p.target}</td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">{p.port}</td>
                    <td className="px-4 py-3 text-sm text-muted-foreground">{p.interval_seconds}s</td>
                    <td className="px-4 py-3 text-right text-sm space-x-2">
                      <button type="button" onClick={() => openEdit(p)} className="text-primary hover:opacity-80">{t('settings.edit')}</button>
                      <button type="button" onClick={() => setDeleteConfirm({ open: true, id: p.id })} className="text-destructive hover:opacity-80">{t('common.delete')}</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {probes.length === 0 && (
            <div className="py-12 text-center text-sm text-muted-foreground">{t('probes.noProbes')}</div>
          )}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={(open) => { if (!open) setDialogOpen(false) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              {dialogState.mode === 'create' ? t('probes.addProbe') : t('probes.editProbe')}
            </DialogTitle>
          </DialogHeader>
          {dialogError && (
            <div className="rounded-md bg-destructive/10 px-4 py-2 text-sm text-destructive">{dialogError}</div>
          )}
          <div className="space-y-4">
            {dialogState.mode === 'create' && (
              <div className="space-y-2">
                <Label htmlFor="probe-node">{t('probes.node')} <span className="text-destructive">*</span></Label>
                <select
                  id="probe-node"
                  value={form.node_id}
                  onChange={(e) => setForm({ ...form, node_id: e.target.value })}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                >
                  <option value="">—</option>
                  {nodes.map((n) => <option key={n.id} value={n.id}>{n.name}</option>)}
                </select>
              </div>
            )}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="probe-type">{t('probes.type')}</Label>
                <select id="probe-type" value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value as 'TCP' | 'UDP' })} className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
                  <option value="TCP">TCP</option>
                  <option value="UDP">UDP</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="probe-port">{t('probes.port')}</Label>
                <Input id="probe-port" type="number" min={1} max={65535} value={form.port} onChange={(e) => setForm({ ...form, port: Number(e.target.value) })} />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="probe-target">{t('probes.target')} <span className="text-destructive">*</span></Label>
              <Input id="probe-target" type="text" value={form.target} onChange={(e) => setForm({ ...form, target: e.target.value })} placeholder="IP or domain" />
            </div>
            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="probe-interval">{t('probes.interval')} (s)</Label>
                <Input id="probe-interval" type="number" min={60} max={300} value={form.interval_seconds} onChange={(e) => setForm({ ...form, interval_seconds: Number(e.target.value) })} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="probe-count">{t('probes.count')}</Label>
                <Input id="probe-count" type="number" min={1} max={100} value={form.count} onChange={(e) => setForm({ ...form, count: Number(e.target.value) })} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="probe-timeout">{t('probes.timeout')} (s)</Label>
                <Input id="probe-timeout" type="number" min={1} max={30} value={form.timeout_seconds} onChange={(e) => setForm({ ...form, timeout_seconds: Number(e.target.value) })} />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button type="button" onClick={() => void handleDialogSubmit()} disabled={dialogLoading}>
              {dialogLoading ? t('common.saving') : t('common.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={deleteConfirm.open} onOpenChange={(open) => !open && setDeleteConfirm({ open: false })}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('probes.confirmDelete')}</AlertDialogTitle>
            <AlertDialogDescription>{t('probes.confirmDeleteMessage')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={() => void handleDelete()} disabled={deleteLoading} variant="destructive">
              {t('common.delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
