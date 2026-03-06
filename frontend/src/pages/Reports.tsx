/**
 * Reports Page
 *
 * Page for generating and exporting reports with:
 * - Report type selection (health, performance, comparison)
 * - Time range selector
 * - Node/metric selector
 * - Export options (CSV, PDF, Excel)
 */

import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useTheme } from '../hooks/useTheme'
import { useExportStore } from '../stores/exportStore'
import { fetchNodes } from '../api/nodes'
import { ReportGenerator, NodeComparisonTable, type ReportConfig, type NodeComparisonData } from '../components/reports'
import { ExportStatusCard, ExportHistoryTable } from '../components/export'
import { PageContainer } from '../components/common'
import { PageHeader } from '../components/layout/PageHeader'
import type { NodeDTO } from '../api/types'
import type { CreateExportRequest } from '../types/export'

export default function ReportsPage() {
  const { t } = useTranslation()
  const { isDark } = useTheme()

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

  useEffect(() => {
    loadNodes()
    fetchExportHistory()

    return () => {
      useExportStore.getState().stopAllPolling()
    }
  }, [])

  const loadNodes = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetchNodes()
      const nodesArray = response.data.nodes || []
      setNodes(nodesArray)
      // Initialize comparison data with all nodes
      setComparisonData(
        nodesArray.slice(0, 5).map((node) => ({
          nodeId: node.id,
          nodeName: node.name,
          region: node.region,
          status: node.status as 'online' | 'offline' | 'connecting',
          latency: Math.random() * 100 + 20, // Placeholder - would come from API
          packetLoss: Math.random() * 2, // Placeholder
          jitter: Math.random() * 30, // Placeholder
        }))
      )
    } catch (err) {
      setError(err as Error)
      console.error('Failed to load nodes:', err)
    } finally {
      setIsLoading(false)
    }
  }

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
        showBreadcrumb
      />

        {/* Error State */}
        {error && (
          <div className="mb-6 bg-red-50 dark:bg-red-900/20 border-l-4 border-red-400 p-4 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800 dark:text-red-300">
                  {t('errors.failedToLoad')}
                </h3>
                <p className="text-sm text-red-700 dark:text-red-400">{error.message}</p>
              </div>
              <div className="ml-auto pl-3">
                <button
                  onClick={() => loadNodes()}
                  className="inline-flex bg-red-50 dark:bg-red-900/30 rounded-md p-1.5 text-red-500 hover:bg-red-100 dark:hover:bg-red-900/50"
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
              className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"
              data-testid="loading-spinner"
            />
          </div>
        )}

        {/* Content */}
        {!isLoading && !error && (
          <div className="space-y-8">
            {/* Report Generator Form */}
            <div className={`${isDark ? 'bg-gray-800' : 'bg-white'} rounded-lg shadow-md p-6`}>
              <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'} mb-4`}>
                {t('reports.generateReport')}
              </h3>
              <ReportGenerator
                nodes={nodes}
                onSubmit={handleReportSubmit}
                loading={exportLoading}
              />
            </div>

            {/* Node Comparison Table */}
            <div className={`${isDark ? 'bg-gray-800' : 'bg-white'} rounded-lg shadow-md p-6`}>
              <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'} mb-4`}>
                {t('nodes.comparison')}
              </h3>
              <NodeComparisonTable nodes={comparisonData} highlightDifferences={true} />
            </div>

            {/* Active Exports */}
            {currentExports.length > 0 && (
              <div className={`${isDark ? 'bg-gray-800' : 'bg-white'} rounded-lg shadow-md p-6`}>
                <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'} mb-4`}>
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
              <div className={`${isDark ? 'bg-gray-800' : 'bg-white'} rounded-lg shadow-md p-6`}>
                <h3 className={`text-lg font-semibold ${isDark ? 'text-white' : 'text-gray-900'} mb-4`}>
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
