/**
 * Reports Page
 *
 * Page for generating and exporting reports with:
 * - Report type selection (health, performance, comparison)
 * - Time range selector
 * - Node/metric selector
 * - Export options (CSV, PDF, Excel)
 */

import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSearchParams } from 'react-router-dom'
import { useExportStore } from '../stores/exportStore'
import { fetchNodes } from '../api/nodes'
import { fetchMetrics } from '../api/data'
import { ReportGenerator, NodeComparisonTable, type ReportConfig, type NodeComparisonData } from '../components/reports'
import { ExportStatusCard, ExportHistoryTable } from '../components/export'
import { PageContainer } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import type { NodeDTO } from '../api/types'
import type { CreateExportRequest } from '../types/export'

export default function ReportsPage() {
  const { t } = useTranslation()
  const [searchParams] = useSearchParams()
  // Pre-select node when navigated from NodeDetailPage (?nodeId=<id>)
  const preselectedNodeId = searchParams.get('nodeId')
  const defaultNodeIds = preselectedNodeId ? [preselectedNodeId] : undefined

  const {
    createExport,
    currentExports,
    exportHistory,
    downloadExport,
    fetchExportHistory,
    isLoading: exportLoading,
  } = useExportStore()

  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [nodes, setNodes] = useState<NodeDTO[]>([])
  const [comparisonData, setComparisonData] = useState<NodeComparisonData[]>([])

  const loadNodes = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      const nodesArray = response.data.nodes || []
      setNodes(nodesArray)

      const slice = nodesArray.slice(0, 5)
      const metricsByNode = new Map<
        string,
        { latency_ms: number; packet_loss_rate: number; jitter_ms: number }
      >()
      if (slice.length > 0) {
        try {
          const mr = await fetchMetrics(slice.map((n) => n.id))
          for (const m of mr.data) {
            metricsByNode.set(m.node_id, {
              latency_ms: m.latency_ms,
              packet_loss_rate: m.packet_loss_rate,
              jitter_ms: m.jitter_ms,
            })
          }
        } catch {
          // Leave map empty; table still lists nodes without live metrics
        }
      }

      setComparisonData(
        slice.map((node) => {
          const m = metricsByNode.get(node.id)
          return {
            nodeId: node.id,
            nodeName: node.name,
            region: node.region,
            status: node.status as 'online' | 'offline' | 'connecting',
            latency: m?.latency_ms,
            packetLoss: m?.packet_loss_rate,
            jitter: m?.jitter_ms,
          }
        })
      )
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load nodes:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadNodes()
    void fetchExportHistory()

    return () => {
      useExportStore.getState().stopAllPolling()
    }
  }, [fetchExportHistory, loadNodes])

  const handleReportSubmit = async (config: ReportConfig) => {
    try {
      // PDF format is handled client-side via browser print
      if (config.format === 'pdf') {
        // PDF preview is shown in ReportGenerator component
        return
      }

      // Convert ReportConfig to CreateExportRequest
      const exportRequest: CreateExportRequest = {
        node_ids: config.nodeIds,
        start_time: config.dateRange === 'custom' && config.customStartDate
          ? new Date(config.customStartDate).toISOString()
          : new Date(Date.now() - (config.dateRange === '7d' ? 7 : 30) * 24 * 60 * 60 * 1000).toISOString(),
        end_time: config.dateRange === 'custom' && config.customEndDate
          ? new Date(config.customEndDate).toISOString()
          : new Date().toISOString(),
        metrics: config.metrics,
        format: config.format as 'csv' | 'excel', // API supports csv/excel only
      }

      await createExport(exportRequest)
    } catch (error) {
      console.error('Failed to create export:', error)
      throw error
    }
  }

  return (
    <PageContainer>
      <PageHeader
        title={t('reports.title')}
        subtitle={`${t('reports.generateReport')} - ${t('reports.exportHistory')}`}
      />

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-[var(--color-critical-bg)] border-l-4 border-[var(--color-critical)] p-4 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-[var(--color-critical)]" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-[var(--color-critical-text)]">
                  {t('errors.failedToLoad')}
                </h3>
                <p className="text-sm text-[var(--color-critical-text)]">{error.message}</p>
              </div>
              <div className="ml-auto pl-3">
                <button
                  onClick={() => loadNodes()}
                  className="inline-flex bg-[var(--color-critical-bg)] rounded-md p-1.5 text-[var(--color-critical)] hover:bg-[var(--color-critical-bg)] hover:opacity-80"
                >
                  <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
        )}

        {/* Loading State */}
        {isLoading && !error && (
          <div className="flex justify-center items-center py-12">
            <div
              className="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--color-brand)]"
              data-testid="loading-spinner"
            />
          </div>
        )}

        {/* Content */}
        {!isLoading && !error && (
          <div className="space-y-8">
            {/* Report Generator Form */}
            <div className="rounded-lg shadow-md p-6 bg-[var(--color-bg-surface)]">
              <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
                {t('reports.generateReport')}
              </h3>
              <ReportGenerator
                nodes={nodes}
                onSubmit={handleReportSubmit}
                loading={exportLoading}
                defaultNodeIds={defaultNodeIds}
              />
            </div>

            {/* Node Comparison Table */}
            <div className="rounded-lg shadow-md p-6 bg-[var(--color-bg-surface)]">
              <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
                {t('nodes.comparison')}
              </h3>
              <NodeComparisonTable nodes={comparisonData} highlightDifferences={true} />
            </div>

            {/* Active Exports */}
            {currentExports.length > 0 && (
              <div className="rounded-lg shadow-md p-6 bg-[var(--color-bg-surface)]">
                <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
                  {t('reports.download')}
                </h3>
                <div className="space-y-4">
                  {currentExports.map((exportTask) => (
                    <ExportStatusCard
                      key={exportTask.id}
                      exportTask={exportTask}
                      onDownload={downloadExport}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Export History */}
            {exportHistory.length > 0 && (
              <div className="rounded-lg shadow-md p-6 bg-[var(--color-bg-surface)]">
                <h3 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4">
                  {t('reports.exportHistory')}
                </h3>
                <ExportHistoryTable
                  exports={exportHistory}
                  onDownload={downloadExport}
                  onDelete={() => {
                    /* Delete not implemented in MVP */
                  }}
                />
              </div>
            )}
          </div>
        )}
    </PageContainer>
  )
}
