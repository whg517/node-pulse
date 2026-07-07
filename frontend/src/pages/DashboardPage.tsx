import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useDashboardData } from '@/hooks/useDashboardData'
import { useDashboard } from '@/hooks/useDashboard'
import { useDashboardHistory } from '@/hooks/useDashboardHistory'

import { PageHeader } from '@/components/layout/PageHeader'
import { Button } from '@/components/ui/button'
import { NodeSummaryCard } from '@/components/dashboard/NodeSummaryCard'
import { AlertStream } from '@/components/dashboard/AlertStream'
import WorldMap from '@/components/dashboard/WorldMap'
import { LatencyTrendChart, PacketLossChart, ProbeSuccessGauge } from '@/components/charts'
import type { NodeLocation } from '@/components/dashboard/WorldMap'
import { useAlertsStore } from '@/stores/alertsStore'
import { useAuthStore } from '@/stores/authStore'
import { useDashboardStore } from '@/stores/dashboardStore'
import { estimateRegionBaseCoordinates, scatterAroundBase } from '@/utils/regionCoordinates'

export default function DashboardPage() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const { role } = useAuthStore()
  const canManageNodes = role === 'admin' || role === 'operator'
  const { nodes, metrics, isLoading, error, refetch } = useDashboardData()
  const { stats, sortedByAnomaly, nodeHealthSummaries } = useDashboard(nodes, metrics)
  const fetchAlertRecords = useAlertsStore((s) => s.fetchAlertRecords)
  const refreshInterval = useDashboardStore((s) => s.refreshInterval)
  const setRefreshInterval = useDashboardStore((s) => s.setRefreshInterval)
  const autoRefresh = useDashboardStore((s) => s.autoRefresh)
  const toggleAutoRefresh = useDashboardStore((s) => s.toggleAutoRefresh)

  const [historyRefreshToken, setHistoryRefreshToken] = useState(0)

  const nodeIdsForHistory = useMemo(() => nodes.map((n) => n.id), [nodes])
  const { latencyTrend, packetLossTrend, isLoading: historyLoading } = useDashboardHistory(nodeIdsForHistory, historyRefreshToken)

  useEffect(() => { void fetchAlertRecords() }, [fetchAlertRecords])

  const handleRefetch = useCallback(async () => {
    await refetch()
    setHistoryRefreshToken((n) => n + 1)
    void fetchAlertRecords()
  }, [refetch, fetchAlertRecords])

  const nodeMapLocations = useMemo((): NodeLocation[] => {
    return nodeHealthSummaries.map(({ node, metrics: nodeMetrics, healthStatus }) => {
      const base = estimateRegionBaseCoordinates(node.region)
      const { lat, lng } = scatterAroundBase(base, node.id)
      return {
        id: node.id, name: node.name, lat, lng, region: node.region, healthStatus,
        avgLatency: nodeMetrics?.latency_ms ?? 0, packetLoss: nodeMetrics?.packet_loss_rate ?? 0,
      }
    })
  }, [nodeHealthSummaries])

  const chartsLoading = isLoading || historyLoading
  const latencyBaseline = stats.averageLatency > 0 ? stats.averageLatency : 100
  const topNodes = sortedByAnomaly.slice(0, 6)

  return (
    <div className="space-y-6">
      <PageHeader
        title={t('dashboard.title')}
        subtitle={t('dashboard.realTimeMetrics')}
        actions={
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 text-sm">
              <label htmlFor="refresh-interval" className="text-muted-foreground whitespace-nowrap">
                {t('dashboard.autoRefreshLabel')}:
              </label>
              <select
                id="refresh-interval"
                value={autoRefresh ? refreshInterval : 0}
                onChange={(e) => {
                  const val = Number(e.target.value)
                  if (val === 0) { if (autoRefresh) toggleAutoRefresh() }
                  else { if (!autoRefresh) toggleAutoRefresh(); setRefreshInterval(val) }
                }}
                className="rounded-md border border-input bg-background px-2 py-1 text-xs"
              >
                <option value={5}>5s</option>
                <option value={10}>10s</option>
                <option value={30}>30s</option>
                <option value={60}>60s</option>
                <option value={0}>{t('dashboard.refreshOff')}</option>
              </select>
            </div>
            <Button variant="outline" size="sm" onClick={() => void handleRefetch()}>
              <svg className="-ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {t('dashboard.refreshData')}
            </Button>
            <Button size="sm" onClick={() => navigate('/nodes')}>{t('dashboard.viewAllNodes')}</Button>
          </div>
        }
      />

      {/* Empty state: no nodes yet. Guide the new admin/operator through onboarding. */}
      {nodes.length === 0 && !isLoading && !error && (
        <div className="rounded-lg border border-primary/30 bg-primary/5 p-6">
          <h3 className="text-lg font-semibold">{t('dashboard.gettingStarted', 'Getting started')}</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {t('dashboard.emptyHint', 'No nodes are reporting yet. Create a node, then deploy a Beacon agent to start monitoring.')}
          </p>
          <ol className="mt-4 space-y-2 text-sm text-muted-foreground list-decimal list-inside">
            <li>{t('dashboard.stepCreateNode', 'Create a node (Nodes → New).')}</li>
            <li>{t('dashboard.stepCreateApiKey', 'Generate an API key (Settings → API Keys, admin-only).')}</li>
            <li>{t('dashboard.stepDeployBeacon', 'Deploy a Beacon agent on the target node with that API key.')}</li>
          </ol>
          <div className="mt-4 flex gap-2">
            {canManageNodes && (
              <Button size="sm" onClick={() => navigate('/nodes')}>{t('dashboard.goToNodes', 'Go to Nodes')}</Button>
            )}
            <Button size="sm" variant="outline" onClick={() => navigate('/settings/api-keys')}>{t('dashboard.goToApiKeys', 'API Keys')}</Button>
          </div>
        </div>
      )}

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error.message}
          <Button variant="link" size="sm" onClick={() => void handleRefetch()}>{t('common.retry')}</Button>
        </div>
      )}

      {/* Health Overview Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="rounded-lg border bg-card p-4">
          <h3 className="text-sm font-semibold text-muted-foreground">{t('metrics.onlineRate')}</h3>
          <div className="mt-3 flex items-center justify-between">
            <p className="text-2xl font-bold">{stats.onlineRate.toFixed(1)}{t('units.percent')}</p>
            <div className="flex size-10 items-center justify-center rounded-full bg-healthy-bg">
              <svg className="size-5 text-healthy" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">{stats.onlineNodes}/{stats.totalNodes} {t('status.online').toLowerCase()}</p>
        </div>

        <div className="rounded-lg border bg-card p-4">
          <h3 className="text-sm font-semibold text-muted-foreground">{t('metrics.anomalyRate')}</h3>
          <div className="mt-3 flex items-center justify-between">
            <p className={`text-2xl font-bold ${stats.anomalyRate > 10 ? 'text-destructive' : stats.anomalyRate > 5 ? 'text-warning-text' : ''}`}>
              {stats.anomalyRate.toFixed(1)}{t('units.percent')}
            </p>
            <div className={`flex size-10 items-center justify-center rounded-full ${stats.anomalyRate > 10 ? 'bg-destructive/10' : stats.anomalyRate > 5 ? 'bg-warning-bg' : 'bg-muted'}`}>
              <svg className={`size-5 ${stats.anomalyRate > 10 ? 'text-destructive' : stats.anomalyRate > 5 ? 'text-warning' : 'text-muted-foreground'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            <span className="text-warning-text">{stats.warningNodes}</span>
            <span className="mx-1 text-muted-foreground/50">|</span>
            <span className="text-destructive">{stats.criticalNodes}</span>
            <span className="ml-1">{t('dashboard.nodesRequiringAttention')}</span>
          </p>
        </div>

        <div className="rounded-lg border bg-card p-4">
          <h3 className="text-sm font-semibold text-muted-foreground">{t('metrics.avgLatency')}</h3>
          <div className="mt-3 flex items-center justify-between">
            <p className={`text-2xl font-bold ${stats.averageLatency > 200 ? 'text-destructive' : stats.averageLatency > 160 ? 'text-warning-text' : ''}`}>
              {stats.averageLatency.toFixed(0)}{t('units.ms')}
            </p>
            <div className="flex size-10 items-center justify-center rounded-full bg-primary/10">
              <svg className="size-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </div>

        <div className="rounded-lg border bg-card p-4">
          <h3 className="text-sm font-semibold text-muted-foreground">{t('dashboard.probeSuccessRate')}</h3>
          <div className="mt-2">
            <ProbeSuccessGauge value={100 - stats.averagePacketLoss} height="110px" isLoading={isLoading} />
          </div>
        </div>
      </div>

      {/* World Map */}
      <WorldMap nodes={nodeMapLocations} height="480px" isLoading={isLoading} onNodeClick={(nodeId) => navigate(`/nodes/${nodeId}`)} />

      {/* Trend Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="rounded-lg border bg-card">
          <div className="border-b px-4 py-3"><h3 className="text-sm font-semibold">{t('dashboard.latencyTrendChart')}</h3></div>
          <div className="p-4">
            <LatencyTrendChart data={latencyTrend} height="220px" isLoading={chartsLoading} showBaseline baselineValue={latencyBaseline} />
          </div>
        </div>
        <div className="rounded-lg border bg-card">
          <div className="border-b px-4 py-3"><h3 className="text-sm font-semibold">{t('dashboard.packetLossChart')}</h3></div>
          <div className="p-4">
            <PacketLossChart data={packetLossTrend} height="220px" isLoading={chartsLoading} warningThreshold={3} criticalThreshold={5} />
          </div>
        </div>
      </div>

      {/* Node Health + Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <div className="mb-4 flex items-center justify-between">
            <h3 className="text-sm font-semibold">{t('dashboard.nodeHealthOverview')}</h3>
            <Button variant="link" size="sm" onClick={() => navigate('/nodes')}>{t('dashboard.viewAllNodes')} →</Button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {topNodes.map(({ node, metrics: nodeMetrics, healthStatus }) => (
              <NodeSummaryCard
                key={node.id} node={node} healthStatus={healthStatus}
                lastSeen={nodeMetrics?.timestamp} latency={nodeMetrics?.latency_ms} packetLoss={nodeMetrics?.packet_loss_rate}
              />
            ))}
            {topNodes.length === 0 && !isLoading && (
              <div className="col-span-2 rounded-lg border bg-card py-8 text-center text-muted-foreground">
                <svg className="mx-auto mb-2 size-8 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                {t('nodes.noNodes')}
              </div>
            )}
          </div>
        </div>
        <div className="lg:col-span-1">
          <AlertStream isLoading={isLoading} />
        </div>
      </div>
    </div>
  )
}
