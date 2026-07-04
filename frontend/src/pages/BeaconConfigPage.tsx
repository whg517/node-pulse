import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageHeader } from '@/components/layout/PageHeader'
import { fetchNodes } from '@/api/nodes'
import { fetchBeaconConfig, updateBeaconConfig, fetchConfigHistory, previewConfig, batchUpdateConfig, rollbackBeaconConfig } from '@/api/beaconConfig'
import type { BeaconConfigDTO, ProbeConfigDTO, ConfigHistoryEntry, ConfigPreviewResult } from '@/api/beaconConfig'
import { listBeaconConfigTemplates, createBeaconConfigTemplate, deleteBeaconConfigTemplate, type BeaconConfigTemplateDTO } from '@/api/beaconConfigTemplates'
import type { NodeDTO } from '@/api/types'

function emptyProbe(): ProbeConfigDTO {
  return {
    id: crypto.randomUUID(),
    type: 'TCP',
    target: '',
    port: 443,
    interval_seconds: 60,
    timeout_seconds: 5,
    count: 3,
  }
}

type ProbeValidationField = 'target' | 'port' | 'interval_seconds' | 'timeout_seconds' | 'count'

interface BeaconConfigValidationErrors {
  interval_seconds?: string
  timeout_seconds?: string
  probes: Record<string, Partial<Record<ProbeValidationField, string>>>
}

function emptyValidationErrors(): BeaconConfigValidationErrors {
  return { probes: {} }
}

type ConfigAckState = 'applied' | 'failed' | 'pending'

function getConfigAckState(config: BeaconConfigDTO): ConfigAckState {
  if (config.last_ack_status === 'failed' && config.last_ack_version === config.version) {
    return 'failed'
  }
  if (config.last_ack_status === 'applied' && (config.last_ack_version ?? 0) >= config.version) {
    return 'applied'
  }
  return 'pending'
}

function getAckBadgeClass(state: ConfigAckState): string {
  switch (state) {
    case 'applied':
      return 'bg-healthy-bg text-healthy'
    case 'failed':
      return 'bg-destructive/10 text-destructive'
    default:
      return 'bg-warning-bg text-warning-text'
  }
}

export default function BeaconConfigPage() {
  const { t } = useTranslation()
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [selectedNodeId, setSelectedNodeId] = useState<string>('')
  const [config, setConfig] = useState<BeaconConfigDTO | null>(null)
  const [history, setHistory] = useState<ConfigHistoryEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [saveMessage, setSaveMessage] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<BeaconConfigValidationErrors>(emptyValidationErrors)
  const [showHistory, setShowHistory] = useState(false)

  // G8 config preview state
  const [preview, setPreview] = useState<ConfigPreviewResult | null>(null)
  const [isPreviewing, setIsPreviewing] = useState(false)

  // G9 batch deploy state (multi-select across all nodes)
  const [batchMode, setBatchMode] = useState(false)
  const [batchSelectedIds, setBatchSelectedIds] = useState<string[]>([])
  const [batchResult, setBatchResult] = useState<{ success_count: number; failed_count: number; failed_ids?: string[]; errors?: string[] } | null>(null)
  const [isBatchDeploying, setIsBatchDeploying] = useState(false)

  const [showSaveTemplate, setShowSaveTemplate] = useState(false)
  const [templateName, setTemplateName] = useState('')
  const [templateDesc, setTemplateDesc] = useState('')
  // Templates now come from the server (ADR-003) instead of localStorage.
  const [templates, setTemplates] = useState<BeaconConfigTemplateDTO[]>([])

  const loadTemplates = useCallback(async () => {
    try {
      const res = await listBeaconConfigTemplates()
      setTemplates(res.data?.templates || [])
    } catch {
      // Best-effort; templates are non-critical.
    }
  }, [])

  useEffect(() => { void loadTemplates() }, [loadTemplates])

  useEffect(() => {
    fetchNodes()
      .then((res) => {
        const list = res.data.nodes ?? []
        setNodes(list)
        if (list.length > 0) setSelectedNodeId(list[0].id)
      })
      .catch(() => {})
  }, [])

  const loadConfig = useCallback(async () => {
    if (!selectedNodeId) return
    setIsLoading(true)
    setError(null)
    setSaveMessage(null)
    try {
      const [cfgRes, histRes] = await Promise.all([
        fetchBeaconConfig(selectedNodeId),
        fetchConfigHistory(selectedNodeId),
      ])
      setConfig(cfgRes.data)
      setHistory(histRes.data ?? [])
      setValidationErrors(emptyValidationErrors())
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [selectedNodeId])

  useEffect(() => {
    if (selectedNodeId) void loadConfig()
  }, [selectedNodeId, loadConfig])

  const clearValidation = () => {
    setValidationErrors(emptyValidationErrors())
    setError(null)
    setSaveMessage(null)
  }

  const validateConfig = (nextConfig: BeaconConfigDTO) => {
    const nextErrors = emptyValidationErrors()

    if (!Number.isFinite(nextConfig.interval_seconds) || nextConfig.interval_seconds < 5) {
      nextErrors.interval_seconds = t('beaconConfig.errorGlobalIntervalMin')
    }
    if (!Number.isFinite(nextConfig.timeout_seconds) || nextConfig.timeout_seconds < 1) {
      nextErrors.timeout_seconds = t('beaconConfig.errorGlobalTimeoutMin')
    }

    nextConfig.probes.forEach((probe) => {
      const probeErrors: Partial<Record<ProbeValidationField, string>> = {}
      if (!probe.target.trim()) {
        probeErrors.target = t('beaconConfig.errorProbeTargetRequired')
      }
      if (!Number.isInteger(probe.port) || probe.port < 1 || probe.port > 65535) {
        probeErrors.port = t('beaconConfig.errorProbePortRange')
      }
      if (!Number.isInteger(probe.interval_seconds) || probe.interval_seconds < 5) {
        probeErrors.interval_seconds = t('beaconConfig.errorProbeIntervalMin')
      }
      if (!Number.isInteger(probe.timeout_seconds) || probe.timeout_seconds < 1) {
        probeErrors.timeout_seconds = t('beaconConfig.errorProbeTimeoutMin')
      }
      if (!Number.isInteger(probe.count) || probe.count < 1 || probe.count > 100) {
        probeErrors.count = t('beaconConfig.errorProbeCountRange')
      }
      if (Object.keys(probeErrors).length > 0) {
        nextErrors.probes[probe.id] = probeErrors
      }
    })

    return {
      errors: nextErrors,
      isValid: !nextErrors.interval_seconds &&
        !nextErrors.timeout_seconds &&
        Object.keys(nextErrors.probes).length === 0,
    }
  }

  const getProbeError = (probeId: string, field: ProbeValidationField) => validationErrors.probes[probeId]?.[field]

  const handleProbeChange = (index: number, field: keyof ProbeConfigDTO, value: string | number) => {
    if (!config) return
    const probes = [...config.probes]
    probes[index] = { ...probes[index], [field]: value }
    setConfig({ ...config, probes })
    clearValidation()
  }

  const handleAddProbe = () => {
    if (!config) return
    setConfig({ ...config, probes: [...config.probes, emptyProbe()] })
    clearValidation()
  }

  const handleRemoveProbe = (index: number) => {
    if (!config) return
    setConfig({ ...config, probes: config.probes.filter((_, i) => i !== index) })
    clearValidation()
  }

  const handleSave = async () => {
    if (!config || !selectedNodeId) return
    const validation = validateConfig(config)
    setValidationErrors(validation.errors)
    if (!validation.isValid) {
      setError(t('beaconConfig.validationErrorSummary'))
      setSaveMessage(null)
      return
    }

    setIsSaving(true)
    setError(null)
    setSaveMessage(null)
    try {
      const res = await updateBeaconConfig(selectedNodeId, {
        probes: config.probes,
        interval_seconds: config.interval_seconds,
        timeout_seconds: config.timeout_seconds,
      })
      setConfig(res.data)
      setSaveMessage(t('beaconConfig.saveSuccess'))
      void loadConfig()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsSaving(false)
    }
  }

  const handlePreview = async () => {
    if (!config || !selectedNodeId) return
    setIsPreviewing(true)
    setError(null)
    setPreview(null)
    try {
      const res = await previewConfig(selectedNodeId, {
        probes: config.probes,
        interval_seconds: config.interval_seconds,
        timeout_seconds: config.timeout_seconds,
      })
      setPreview(res.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsPreviewing(false)
    }
  }

  const handleBatchDeploy = async () => {
    if (!config || batchSelectedIds.length === 0) return
    setIsBatchDeploying(true)
    setError(null)
    setBatchResult(null)
    try {
      const res = await batchUpdateConfig('manual', {
        beacon_ids: batchSelectedIds,
        config: {
          probes: config.probes,
          interval_seconds: config.interval_seconds,
          timeout_seconds: config.timeout_seconds,
        },
      })
      setBatchResult(res.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsBatchDeploying(false)
    }
  }

  const toggleBatchId = (id: string) => {
    setBatchSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    )
  }

  const handleRollback = async (version: number) => {
    if (!selectedNodeId) return
    setIsSaving(true)
    setError(null)
    setSaveMessage(null)
    try {
      const res = await rollbackBeaconConfig(selectedNodeId, version)
      setConfig(res.data)
      setSaveMessage(t('beaconConfig.rollbackSuccess', 'Rolled back to version {v}', { v: version }))
      void loadConfig()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsSaving(false)
    }
  }

  const handleSaveTemplate = async () => {
    if (!config || !templateName.trim()) return
    try {
      await createBeaconConfigTemplate({
        name: templateName.trim(),
        description: templateDesc.trim() || undefined,
        probes: config.probes.map((p) => ({
          type: p.type,
          target: p.target,
          port: p.port,
          interval_seconds: p.interval_seconds,
          timeout_seconds: p.timeout_seconds,
          count: p.count,
        })),
        interval_seconds: config.interval_seconds,
        timeout_seconds: config.timeout_seconds,
      })
      setTemplateName('')
      setTemplateDesc('')
      setShowSaveTemplate(false)
      await loadTemplates()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }

  const handleApplyTemplate = (template: BeaconConfigTemplateDTO) => {
    if (!config) return
    setConfig({
      ...config,
      probes: (template.probes || []).map((p) => ({
        id: crypto.randomUUID(),
        type: (p.type as 'TCP' | 'UDP') || 'TCP',
        target: p.target || '',
        port: p.port ?? 443,
        interval_seconds: p.interval_seconds ?? 60,
        timeout_seconds: p.timeout_seconds ?? 5,
        count: p.count ?? 10,
      })),
      interval_seconds: template.interval_seconds,
      timeout_seconds: template.timeout_seconds,
    })
    clearValidation()
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('beaconConfig.title')}
        subtitle={t('beaconConfig.subtitle')}
        actions={
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setShowHistory(!showHistory)}
              className="px-4 py-2 text-sm font-medium rounded-lg border border-border text-muted-foreground hover:bg-accent/10"
            >
              {showHistory ? t('beaconConfig.hideHistory') : t('beaconConfig.showHistory')}
            </button>
            <button
              type="button"
              onClick={() => void handlePreview()}
              disabled={isPreviewing || !config || !selectedNodeId}
              className="px-4 py-2 text-sm font-medium rounded-lg border border-border text-foreground hover:bg-accent/10 disabled:opacity-50"
            >
              {isPreviewing ? t('common.saving') : t('beaconConfig.preview', 'Preview')}
            </button>
            <button
              type="button"
              onClick={() => void handleSave()}
              disabled={isSaving || !config}
              className="px-4 py-2 bg-primary hover:bg-primary/85 text-white text-sm font-medium rounded-lg disabled:opacity-50"
            >
              {isSaving ? t('common.saving') : t('common.save')}
            </button>
          </div>
        }
      />

      {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}
      {saveMessage && (
        <div className="mb-4 rounded-lg bg-healthy-bg text-healthy px-4 py-2 text-sm">
          {saveMessage}
        </div>
      )}
      {preview && (
        <div className={`mb-4 rounded-md px-4 py-3 text-sm ${preview.valid ? 'bg-healthy-bg text-healthy-text' : 'bg-destructive/10 text-destructive'}`}>
          <p className="font-medium">
            {preview.valid ? t('beaconConfig.previewValid', 'Configuration is valid') : t('beaconConfig.previewInvalid', 'Configuration has problems')}
          </p>
          {preview.conflicts.length > 0 && (
            <ul className="mt-1 list-disc pl-5 text-destructive">
              {preview.conflicts.map((c, i) => <li key={i}>{c}</li>)}
            </ul>
          )}
          {preview.warnings.length > 0 && (
            <ul className="mt-1 list-disc pl-5">
              {preview.warnings.map((w, i) => <li key={i}>{w}</li>)}
            </ul>
          )}
        </div>
      )}
      {batchResult && (
        <div className="mb-4 rounded-md border-l-4 border-primary bg-primary/5 px-4 py-3 text-sm">
          <p className="font-medium">
            {t('beaconConfig.batchResult', 'Batch deploy: {success} succeeded, {failed} failed.', { success: batchResult.success_count, failed: batchResult.failed_count })}
          </p>
          {batchResult.failed_ids && batchResult.failed_ids.length > 0 && (
            <p className="mt-1 text-muted-foreground">{t('beaconConfig.batchFailedIds', 'Failed: {ids}', { ids: batchResult.failed_ids.join(', ') })}</p>
          )}
        </div>
      )}

      {/* Node Selector + Batch deploy (G9) */}
      <div className="mb-6 space-y-3">
        <div className="flex items-center justify-between">
          <label className="block text-sm font-medium text-muted-foreground">
            {t('beaconConfig.selectNode')}
          </label>
          <button
            type="button"
            onClick={() => { setBatchMode(!batchMode); setBatchSelectedIds([]); setBatchResult(null) }}
            className="text-sm text-primary hover:text-primary"
          >
            {batchMode ? t('beaconConfig.exitBatch', 'Exit batch mode') : t('beaconConfig.batchDeploy', 'Batch deploy to many')}
          </button>
        </div>
        {!batchMode ? (
          <select
            value={selectedNodeId}
            onChange={(e) => setSelectedNodeId(e.target.value)}
            className="w-full max-w-md rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground"
          >
            {nodes.map((n) => (
              <option key={n.id} value={n.id}>
                {n.name} ({n.region})
              </option>
            ))}
          </select>
        ) : (
          <div className="rounded-lg border border-border p-3">
            <p className="mb-2 text-xs text-muted-foreground">
              {t('beaconConfig.batchHint', 'Select target beacons to deploy the current configuration to.')}
            </p>
            <div className="grid max-h-48 grid-cols-1 gap-1 overflow-y-auto sm:grid-cols-2">
              {nodes.map((n) => (
                <label key={n.id} className="flex items-center gap-2 rounded px-2 py-1 text-sm hover:bg-accent/10">
                  <input
                    type="checkbox"
                    checked={batchSelectedIds.includes(n.id)}
                    onChange={() => toggleBatchId(n.id)}
                  />
                  <span>{n.name} <span className="text-muted-foreground">({n.region})</span></span>
                </label>
              ))}
            </div>
            <button
              type="button"
              onClick={() => void handleBatchDeploy()}
              disabled={isBatchDeploying || batchSelectedIds.length === 0 || !config}
              className="mt-3 px-4 py-2 bg-primary hover:bg-primary/85 text-white text-sm font-medium rounded-lg disabled:opacity-50"
            >
              {isBatchDeploying
                ? t('common.saving')
                : t('beaconConfig.deployToN', 'Deploy to {n} beacon(s)', { n: batchSelectedIds.length })}
            </button>
          </div>
        )}
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" /></div>
      ) : config ? (
        <div className="space-y-6">
          {/* Global Config */}
          <div className="rounded-lg border border-border bg-card p-4">
            <h3 className="text-sm font-semibold text-foreground mb-3">
              {t('beaconConfig.globalSettings')}
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {t('beaconConfig.intervalSeconds')}
                </label>
                <input
                  type="number"
                  min={5}
                  value={config.interval_seconds}
                  onChange={(e) => {
                    setConfig({ ...config, interval_seconds: Number(e.target.value) })
                    clearValidation()
                  }}
                  className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground"
                />
                {validationErrors.interval_seconds && (
                  <p className="mt-1 text-xs text-destructive">{validationErrors.interval_seconds}</p>
                )}
              </div>
              <div>
                <label className="block text-xs font-medium text-muted-foreground mb-1">
                  {t('beaconConfig.timeoutSeconds')}
                </label>
                <input
                  type="number"
                  min={1}
                  value={config.timeout_seconds}
                  onChange={(e) => {
                    setConfig({ ...config, timeout_seconds: Number(e.target.value) })
                    clearValidation()
                  }}
                  className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground"
                />
                {validationErrors.timeout_seconds && (
                  <p className="mt-1 text-xs text-destructive">{validationErrors.timeout_seconds}</p>
                )}
              </div>
            </div>
            <div className="mt-2 text-xs text-muted-foreground">
              {t('beaconConfig.version')}: {config.version} · {t('beaconConfig.updated')}: {new Date(config.updated_at).toLocaleString()}
            </div>
            <div className="mt-4 border-t border-border pt-3">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <h4 className="text-sm font-medium text-foreground">{t('beaconConfig.applyStatus')}</h4>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {t('beaconConfig.applyStatusDetail', {
                      current: config.version,
                      acked: config.last_ack_version ?? t('common.none'),
                    })}
                  </p>
                </div>
                <span className={`inline-flex w-fit items-center rounded-full px-2.5 py-1 text-xs font-medium ${getAckBadgeClass(getConfigAckState(config))}`}>
                  {t(`beaconConfig.ackStatus.${getConfigAckState(config)}`)}
                </span>
              </div>
              <div className="mt-3 grid gap-2 text-xs text-muted-foreground sm:grid-cols-2">
                <div>
                  <span className="font-medium text-foreground">{t('beaconConfig.lastAckAt')}: </span>
                  {config.last_ack_at ? new Date(config.last_ack_at).toLocaleString() : t('common.none')}
                </div>
                <div>
                  <span className="font-medium text-foreground">{t('beaconConfig.lastAckVersion')}: </span>
                  {config.last_ack_version ?? t('common.none')}
                </div>
              </div>
              {getConfigAckState(config) === 'failed' && config.last_ack_error && (
                <p className="mt-2 text-xs text-destructive">{config.last_ack_error}</p>
              )}
            </div>
          </div>

          {/* Probe List */}
          <div className="rounded-lg border border-border bg-card">
            <div className="px-4 py-3 border-b border-border flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">
                {t('beaconConfig.probes')} ({config.probes.length})
              </h3>
              <button
                type="button"
                onClick={handleAddProbe}
                className="text-sm text-primary hover:text-primary font-medium"
              >
                + {t('beaconConfig.addProbe')}
              </button>
            </div>
            {config.probes.length === 0 ? (
              <div className="p-8 text-center text-sm text-muted-foreground">
                {t('beaconConfig.noProbes')}
              </div>
            ) : (
              <div className="divide-y divide-border">
                {config.probes.map((probe, i) => (
                  <div key={probe.id} className="p-4">
                    <div className="flex items-center justify-between mb-3">
                      <span className="text-sm font-medium text-foreground">
                        {t('beaconConfig.probe')} #{i + 1}
                      </span>
                      <button
                        type="button"
                        onClick={() => handleRemoveProbe(i)}
                        className="text-xs text-destructive hover:opacity-80"
                      >
                        {t('common.delete')}
                      </button>
                    </div>
                    <div className="grid grid-cols-3 gap-3">
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">{t('beaconConfig.probeType')}</label>
                        <select
                          value={probe.type}
                          onChange={(e) => handleProbeChange(i, 'type', e.target.value)}
                          className="w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                        >
                          <option value="TCP">TCP</option>
                          <option value="UDP">UDP</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">{t('beaconConfig.probeTarget')}</label>
                        <input
                          type="text"
                          value={probe.target}
                          onChange={(e) => handleProbeChange(i, 'target', e.target.value)}
                          placeholder={t('beaconConfig.probeTargetPlaceholder')}
                          className="w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                        />
                        {getProbeError(probe.id, 'target') && (
                          <p className="mt-1 text-xs text-destructive">{getProbeError(probe.id, 'target')}</p>
                        )}
                      </div>
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">{t('beaconConfig.probePort')}</label>
                        <input
                          type="number"
                          min={1}
                          max={65535}
                          value={probe.port}
                          onChange={(e) => handleProbeChange(i, 'port', Number(e.target.value))}
                          className="w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                        />
                        {getProbeError(probe.id, 'port') && (
                          <p className="mt-1 text-xs text-destructive">{getProbeError(probe.id, 'port')}</p>
                        )}
                      </div>
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">{t('beaconConfig.probeIntervalSeconds')}</label>
                        <input
                          type="number"
                          min={5}
                          value={probe.interval_seconds}
                          onChange={(e) => handleProbeChange(i, 'interval_seconds', Number(e.target.value))}
                          className="w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                        />
                        {getProbeError(probe.id, 'interval_seconds') && (
                          <p className="mt-1 text-xs text-destructive">{getProbeError(probe.id, 'interval_seconds')}</p>
                        )}
                      </div>
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">{t('beaconConfig.probeTimeoutSeconds')}</label>
                        <input
                          type="number"
                          min={1}
                          value={probe.timeout_seconds}
                          onChange={(e) => handleProbeChange(i, 'timeout_seconds', Number(e.target.value))}
                          className="w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                        />
                        {getProbeError(probe.id, 'timeout_seconds') && (
                          <p className="mt-1 text-xs text-destructive">{getProbeError(probe.id, 'timeout_seconds')}</p>
                        )}
                      </div>
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">{t('beaconConfig.probeCount')}</label>
                        <input
                          type="number"
                          min={1}
                          max={100}
                          value={probe.count}
                          onChange={(e) => handleProbeChange(i, 'count', Number(e.target.value))}
                          className="w-full rounded-md border border-input bg-background px-2 py-1 text-xs"
                        />
                        {getProbeError(probe.id, 'count') && (
                          <p className="mt-1 text-xs text-destructive">{getProbeError(probe.id, 'count')}</p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Config History */}
          {showHistory && history.length > 0 && (
            <div className="rounded-lg border border-border bg-card">
              <div className="px-4 py-3 border-b border-border">
                <h3 className="text-sm font-semibold text-foreground">
                  {t('beaconConfig.configHistory')}
                </h3>
              </div>
              <div className="divide-y divide-border">
                {history.map((entry) => (
                  <div key={entry.version} className="p-3 flex items-center justify-between text-sm">
                    <div>
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="font-medium text-foreground">
                          v{entry.version}
                        </span>
                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${getAckBadgeClass(getConfigAckState(entry.config))}`}>
                          {t(`beaconConfig.ackStatus.${getConfigAckState(entry.config)}`)}
                        </span>
                      </div>
                      <span className="ml-2 text-muted-foreground">
                        {entry.config.probes.length} {t('beaconConfig.probes')}
                      </span>
                      {entry.config.last_ack_error && (
                        <p className="mt-1 text-xs text-destructive">{entry.config.last_ack_error}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-muted-foreground">
                      <span>{entry.changed_by} · {new Date(entry.changed_at).toLocaleString()}</span>
                      {config && entry.version !== config.version && (
                        <button
                          type="button"
                          onClick={() => void handleRollback(entry.version)}
                          disabled={isSaving}
                          className="text-primary hover:text-primary disabled:opacity-50"
                        >
                          {t('beaconConfig.rollback', 'Roll back')}
                        </button>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Config Templates */}
          <div className="rounded-lg border border-border bg-card">
            <div className="px-4 py-3 border-b border-border flex items-center justify-between">
              <h3 className="text-sm font-semibold text-foreground">
                {t('beaconConfig.templates')}
              </h3>
              {!showSaveTemplate && (
                <button
                  type="button"
                  onClick={() => setShowSaveTemplate(true)}
                  className="text-sm text-primary hover:text-primary font-medium"
                >
                  + {t('beaconConfig.saveAsTemplate')}
                </button>
              )}
            </div>
            {showSaveTemplate && (
              <div className="px-4 py-3 border-b border-border space-y-2">
                <input
                  type="text"
                  value={templateName}
                  onChange={(e) => setTemplateName(e.target.value)}
                  placeholder={t('beaconConfig.templateName')}
                  className="w-full rounded-md border border-input bg-background px-2 py-1 text-sm"
                />
                <input
                  type="text"
                  value={templateDesc}
                  onChange={(e) => setTemplateDesc(e.target.value)}
                  placeholder={t('beaconConfig.templateDescription')}
                  className="w-full rounded-md border border-input bg-background px-2 py-1 text-sm"
                />
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={handleSaveTemplate}
                    disabled={!templateName.trim()}
                    className="px-3 py-1 bg-primary text-white text-xs font-medium rounded-md disabled:opacity-50"
                  >
                    {t('common.save')}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowSaveTemplate(false); setTemplateName(''); setTemplateDesc('') }}
                    className="px-3 py-1 text-xs font-medium rounded-md border border-border text-muted-foreground"
                  >
                    {t('common.cancel')}
                  </button>
                </div>
              </div>
            )}
            {templates.length === 0 ? (
              <div className="p-6 text-center text-sm text-muted-foreground">
                {t('beaconConfig.noTemplates')}
              </div>
            ) : (
              <div className="divide-y divide-border">
                {templates.map((tmpl) => (
                  <div key={tmpl.id} className="px-4 py-3 flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-foreground">{tmpl.name}</p>
                      {tmpl.description && (
                        <p className="text-xs text-muted-foreground">{tmpl.description}</p>
                      )}
                      <p className="text-xs text-muted-foreground">
                        {(tmpl.probes?.length ?? 0)} {t('beaconConfig.probes')}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => handleApplyTemplate(tmpl)}
                        className="text-xs text-primary hover:opacity-80 font-medium"
                      >
                        {t('beaconConfig.applyTemplate')}
                      </button>
                      <button
                        type="button"
                        onClick={() => { void deleteBeaconConfigTemplate(tmpl.id).then(() => loadTemplates()) }}
                        className="text-xs text-destructive hover:opacity-80"
                      >
                        {t('common.delete')}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  )
}
