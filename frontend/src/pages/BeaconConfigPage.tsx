import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PageContainer, ErrorBanner, LoadingSpinner } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import { fetchNodes } from '../api/nodes'
import { fetchBeaconConfig, updateBeaconConfig, fetchConfigHistory } from '../api/beaconConfig'
import type { BeaconConfigDTO, ProbeConfigDTO, ConfigHistoryEntry } from '../api/beaconConfig'
import type { NodeDTO } from '../api/types'
import { useSettingsStore, type ConfigTemplate } from '../stores/settingsStore'

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
  const [showHistory, setShowHistory] = useState(false)

  const [showSaveTemplate, setShowSaveTemplate] = useState(false)
  const [templateName, setTemplateName] = useState('')
  const [templateDesc, setTemplateDesc] = useState('')
  const { configTemplates, addConfigTemplate, deleteConfigTemplate } = useSettingsStore()

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
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setIsLoading(false)
    }
  }, [selectedNodeId])

  useEffect(() => {
    if (selectedNodeId) void loadConfig()
  }, [selectedNodeId, loadConfig])

  const handleProbeChange = (index: number, field: keyof ProbeConfigDTO, value: string | number) => {
    if (!config) return
    const probes = [...config.probes]
    probes[index] = { ...probes[index], [field]: value }
    setConfig({ ...config, probes })
  }

  const handleAddProbe = () => {
    if (!config) return
    setConfig({ ...config, probes: [...config.probes, emptyProbe()] })
  }

  const handleRemoveProbe = (index: number) => {
    if (!config) return
    setConfig({ ...config, probes: config.probes.filter((_, i) => i !== index) })
  }

  const handleSave = async () => {
    if (!config || !selectedNodeId) return
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

  const handleSaveTemplate = () => {
    if (!config || !templateName.trim()) return
    const template: ConfigTemplate = {
      id: crypto.randomUUID(),
      name: templateName.trim(),
      description: templateDesc.trim() || undefined,
      probes: config.probes.map((p) => ({
        type: p.type as 'TCP' | 'UDP',
        target: p.target,
        port: p.port,
        interval_seconds: p.interval_seconds,
        timeout_seconds: p.timeout_seconds,
        count: p.count,
      })),
      interval_seconds: config.interval_seconds,
      timeout_seconds: config.timeout_seconds,
    }
    addConfigTemplate(template)
    setTemplateName('')
    setTemplateDesc('')
    setShowSaveTemplate(false)
  }

  const handleApplyTemplate = (template: ConfigTemplate) => {
    if (!config) return
    setConfig({
      ...config,
      probes: template.probes.map((p) => ({
        ...p,
        id: crypto.randomUUID(),
      })),
      interval_seconds: template.interval_seconds,
      timeout_seconds: template.timeout_seconds,
    })
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('beaconConfig.title')}
        subtitle={t('beaconConfig.subtitle')}
        actions={
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => setShowHistory(!showHistory)}
              className="px-4 py-2 text-sm font-medium rounded-lg border border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-hover-overlay)]"
            >
              {showHistory ? t('beaconConfig.hideHistory') : t('beaconConfig.showHistory')}
            </button>
            <button
              type="button"
              onClick={() => void handleSave()}
              disabled={isSaving || !config}
              className="px-4 py-2 bg-[var(--color-brand)] hover:bg-[var(--color-brand-hover)] text-white text-sm font-medium rounded-lg disabled:opacity-50"
            >
              {isSaving ? t('common.saving') : t('common.save')}
            </button>
          </div>
        }
      />

      {error && <ErrorBanner error={new Error(error)} onRetry={loadConfig} className="mb-4" />}
      {saveMessage && (
        <div className="mb-4 rounded-lg bg-[var(--color-healthy-bg)] text-[var(--color-healthy)] px-4 py-2 text-sm">
          {saveMessage}
        </div>
      )}

      {/* Node Selector */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-2">
          {t('beaconConfig.selectNode')}
        </label>
        <select
          value={selectedNodeId}
          onChange={(e) => setSelectedNodeId(e.target.value)}
          className="w-full max-w-md rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
        >
          {nodes.map((n) => (
            <option key={n.id} value={n.id}>
              {n.name} ({n.region})
            </option>
          ))}
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><LoadingSpinner /></div>
      ) : config ? (
        <div className="space-y-6">
          {/* Global Config */}
          <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)] p-4">
            <h3 className="text-sm font-semibold text-[var(--color-text-primary)] mb-3">
              {t('beaconConfig.globalSettings')}
            </h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('beaconConfig.intervalSeconds')}
                </label>
                <input
                  type="number"
                  min={5}
                  value={config.interval_seconds}
                  onChange={(e) => setConfig({ ...config, interval_seconds: Number(e.target.value) })}
                  className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1">
                  {t('beaconConfig.timeoutSeconds')}
                </label>
                <input
                  type="number"
                  min={1}
                  value={config.timeout_seconds}
                  onChange={(e) => setConfig({ ...config, timeout_seconds: Number(e.target.value) })}
                  className="w-full rounded-lg border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-3 py-2 text-sm text-[var(--color-text-primary)]"
                />
              </div>
            </div>
            <div className="mt-2 text-xs text-[var(--color-text-muted)]">
              {t('beaconConfig.version')}: {config.version} · {t('beaconConfig.updated')}: {new Date(config.updated_at).toLocaleString()}
            </div>
          </div>

          {/* Probe List */}
          <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
            <div className="px-4 py-3 border-b border-[var(--color-border)] flex items-center justify-between">
              <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
                {t('beaconConfig.probes')} ({config.probes.length})
              </h3>
              <button
                type="button"
                onClick={handleAddProbe}
                className="text-sm text-[var(--color-brand)] hover:text-[var(--color-brand-hover)] font-medium"
              >
                + {t('beaconConfig.addProbe')}
              </button>
            </div>
            {config.probes.length === 0 ? (
              <div className="p-8 text-center text-sm text-[var(--color-text-secondary)]">
                {t('beaconConfig.noProbes')}
              </div>
            ) : (
              <div className="divide-y divide-[var(--color-border)]">
                {config.probes.map((probe, i) => (
                  <div key={probe.id} className="p-4">
                    <div className="flex items-center justify-between mb-3">
                      <span className="text-sm font-medium text-[var(--color-text-primary)]">
                        {t('beaconConfig.probe')} #{i + 1}
                      </span>
                      <button
                        type="button"
                        onClick={() => handleRemoveProbe(i)}
                        className="text-xs text-[var(--color-critical)] hover:opacity-80"
                      >
                        {t('common.delete')}
                      </button>
                    </div>
                    <div className="grid grid-cols-3 gap-3">
                      <div>
                        <label className="block text-xs text-[var(--color-text-muted)] mb-1">Type</label>
                        <select
                          value={probe.type}
                          onChange={(e) => handleProbeChange(i, 'type', e.target.value)}
                          className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-xs"
                        >
                          <option value="TCP">TCP</option>
                          <option value="UDP">UDP</option>
                        </select>
                      </div>
                      <div>
                        <label className="block text-xs text-[var(--color-text-muted)] mb-1">Target</label>
                        <input
                          type="text"
                          value={probe.target}
                          onChange={(e) => handleProbeChange(i, 'target', e.target.value)}
                          placeholder="IP or domain"
                          className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-xs"
                        />
                      </div>
                      <div>
                        <label className="block text-xs text-[var(--color-text-muted)] mb-1">Port</label>
                        <input
                          type="number"
                          min={1}
                          max={65535}
                          value={probe.port}
                          onChange={(e) => handleProbeChange(i, 'port', Number(e.target.value))}
                          className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-xs"
                        />
                      </div>
                      <div>
                        <label className="block text-xs text-[var(--color-text-muted)] mb-1">Interval (s)</label>
                        <input
                          type="number"
                          min={5}
                          value={probe.interval_seconds}
                          onChange={(e) => handleProbeChange(i, 'interval_seconds', Number(e.target.value))}
                          className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-xs"
                        />
                      </div>
                      <div>
                        <label className="block text-xs text-[var(--color-text-muted)] mb-1">Timeout (s)</label>
                        <input
                          type="number"
                          min={1}
                          value={probe.timeout_seconds}
                          onChange={(e) => handleProbeChange(i, 'timeout_seconds', Number(e.target.value))}
                          className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-xs"
                        />
                      </div>
                      <div>
                        <label className="block text-xs text-[var(--color-text-muted)] mb-1">Count</label>
                        <input
                          type="number"
                          min={1}
                          max={100}
                          value={probe.count}
                          onChange={(e) => handleProbeChange(i, 'count', Number(e.target.value))}
                          className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-xs"
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Config History */}
          {showHistory && history.length > 0 && (
            <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
              <div className="px-4 py-3 border-b border-[var(--color-border)]">
                <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
                  {t('beaconConfig.configHistory')}
                </h3>
              </div>
              <div className="divide-y divide-[var(--color-border)]">
                {history.map((entry) => (
                  <div key={entry.version} className="p-3 flex items-center justify-between text-sm">
                    <div>
                      <span className="font-medium text-[var(--color-text-primary)]">
                        v{entry.version}
                      </span>
                      <span className="ml-2 text-[var(--color-text-secondary)]">
                        {entry.config.probes.length} {t('beaconConfig.probes')}
                      </span>
                    </div>
                    <div className="text-xs text-[var(--color-text-muted)]">
                      {entry.changed_by} · {new Date(entry.changed_at).toLocaleString()}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Config Templates */}
          <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-surface)]">
            <div className="px-4 py-3 border-b border-[var(--color-border)] flex items-center justify-between">
              <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
                {t('beaconConfig.templates')}
              </h3>
              {!showSaveTemplate && (
                <button
                  type="button"
                  onClick={() => setShowSaveTemplate(true)}
                  className="text-sm text-[var(--color-brand)] hover:text-[var(--color-brand-hover)] font-medium"
                >
                  + {t('beaconConfig.saveAsTemplate')}
                </button>
              )}
            </div>
            {showSaveTemplate && (
              <div className="px-4 py-3 border-b border-[var(--color-border)] space-y-2">
                <input
                  type="text"
                  value={templateName}
                  onChange={(e) => setTemplateName(e.target.value)}
                  placeholder={t('beaconConfig.templateName')}
                  className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-sm"
                />
                <input
                  type="text"
                  value={templateDesc}
                  onChange={(e) => setTemplateDesc(e.target.value)}
                  placeholder={t('beaconConfig.templateDescription')}
                  className="w-full rounded-md border border-[var(--color-input-border)] bg-[var(--color-input-bg)] px-2 py-1 text-sm"
                />
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={handleSaveTemplate}
                    disabled={!templateName.trim()}
                    className="px-3 py-1 bg-[var(--color-brand)] text-white text-xs font-medium rounded-md disabled:opacity-50"
                  >
                    {t('common.save')}
                  </button>
                  <button
                    type="button"
                    onClick={() => { setShowSaveTemplate(false); setTemplateName(''); setTemplateDesc('') }}
                    className="px-3 py-1 text-xs font-medium rounded-md border border-[var(--color-border)] text-[var(--color-text-secondary)]"
                  >
                    {t('common.cancel')}
                  </button>
                </div>
              </div>
            )}
            {configTemplates.length === 0 ? (
              <div className="p-6 text-center text-sm text-[var(--color-text-secondary)]">
                {t('beaconConfig.noTemplates')}
              </div>
            ) : (
              <div className="divide-y divide-[var(--color-border)]">
                {configTemplates.map((tmpl) => (
                  <div key={tmpl.id} className="px-4 py-3 flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-[var(--color-text-primary)]">{tmpl.name}</p>
                      {tmpl.description && (
                        <p className="text-xs text-[var(--color-text-muted)]">{tmpl.description}</p>
                      )}
                      <p className="text-xs text-[var(--color-text-muted)]">
                        {tmpl.probes.length} {t('beaconConfig.probes')}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => handleApplyTemplate(tmpl)}
                        className="text-xs text-[var(--color-brand)] hover:opacity-80 font-medium"
                      >
                        {t('beaconConfig.applyTemplate')}
                      </button>
                      <button
                        type="button"
                        onClick={() => deleteConfigTemplate(tmpl.id)}
                        className="text-xs text-[var(--color-critical)] hover:opacity-80"
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
    </PageContainer>
  )
}
