import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { fetchNodes } from '@/api/nodes'
import { PageHeader } from '@/components/layout/PageHeader'
import { getComparisonData, fetchDiagnosis } from '@/api/data'
import type { DiagnosisResultDTO } from '@/api/data'
import type { NodeDTO } from '@/api/types'
import ComparisonChart from '@/components/dashboard/ComparisonChart'
import ProblemDiagnosis from '@/components/dashboard/ProblemDiagnosis'
import type { ProblemType, ConfidenceLevel } from '@/components/dashboard/ProblemDiagnosis'
import type { NodeComparisonData, MetricType } from '@/components/dashboard/ComparisonChart'
import { useDashboardStore } from '@/stores/dashboardStore'
import type { GroupByType } from '@/stores/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

const ISP_TAGS = ['AWS', 'GCP', 'Azure', 'Alibaba', 'Tencent', 'DigitalOcean', 'Linode', 'Vultr', 'OVH', 'Hetzner'] as const

function mapApiProblemToUi(problemType: string): ProblemType {
  switch (problemType) {
    case 'node_local_failure': return 'node_local'
    case 'cross_border_link': return 'cross_border_link'
    case 'isp_routing': return 'carrier_routing'
    default: return 'none'
  }
}

function mapApiConfidence(c: string): ConfidenceLevel {
  if (c === 'high' || c === 'medium' || c === 'low') return c
  return 'medium'
}

export default function NodeComparisonPage() {
  const { t } = useTranslation()

  const {
    comparison: storeComparison,
    setComparisonNodeIds,
    setComparisonMetrics,
    setComparisonTimeRange,
    setComparisonCustomTimeRange,
    setComparisonGroupBy,
  } = useDashboardStore()

  const [availableNodes, setAvailableNodes] = useState<NodeDTO[]>([])
  const [comparisonData, setComparisonData] = useState<NodeComparisonData[] | null>(null)
  const [isLoadingNodes, setIsLoadingNodes] = useState(true)
  const [isLoadingComparison, setIsLoadingComparison] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [diagnosisResult, setDiagnosisResult] = useState<DiagnosisResultDTO | null>(null)
  const [diagnosisLoading, setDiagnosisLoading] = useState(false)
  const [diagnosisError, setDiagnosisError] = useState<string | null>(null)

  const { selectedNodeIds, selectedMetrics, timeRange, customTimeRange, groupBy } = storeComparison

  useEffect(() => {
    if (selectedNodeIds.length < 3) {
      setDiagnosisResult(null)
      setDiagnosisError(null)
      setDiagnosisLoading(false)
    }
  }, [selectedNodeIds])

  useEffect(() => {
    async function loadNodes() {
      try {
        setIsLoadingNodes(true)
        setError(null)
        const { data } = await fetchNodes()
        setAvailableNodes(data.nodes)
      } catch (err) {
        setError(err instanceof Error ? err.message : t('errors.failedToLoad'))
      } finally {
        setIsLoadingNodes(false)
      }
    }
    loadNodes()
  }, [t])

  const nodeOptions = availableNodes.map((node) => ({
    node_id: node.id,
    name: node.name,
    region: node.region,
    isp: node.tags.find((tag): tag is typeof ISP_TAGS[number] => ISP_TAGS.includes(tag as typeof ISP_TAGS[number])) || undefined,
    status: node.status as 'online' | 'offline' | 'connecting',
  }))

  const handleNodeSelectionChange = (nodeIds: string[]) => {
    if (nodeIds.length > 5) { setError(t('nodes.maxNodesSelected', { max: 5 })); return }
    if (nodeIds.length < 2 && nodeIds.length > 0) setError(t('nodes.selectAtLeast', { count: 2 }))
    setComparisonNodeIds(nodeIds)
    setError(null)
  }

  const getTimeRangeParams = (): { start_time: string; end_time: string } => {
    const end = new Date()
    if (timeRange === 'custom' && customTimeRange) return { start_time: customTimeRange.start, end_time: customTimeRange.end }
    const hours = timeRange === '7d' ? 168 : timeRange === '30d' ? 720 : 24
    return { start_time: new Date(end.getTime() - hours * 60 * 60 * 1000).toISOString(), end_time: end.toISOString() }
  }

  const handleCompare = async () => {
    if (selectedNodeIds.length < 2 || selectedNodeIds.length > 5 || selectedMetrics.length === 0) return

    try {
      setIsLoadingComparison(true)
      setError(null)
      setDiagnosisResult(null)
      setDiagnosisError(null)

      const { start_time, end_time } = getTimeRangeParams()
      const apiResponse = await getComparisonData({ node_ids: selectedNodeIds, start_time, end_time, metrics: selectedMetrics })

      const transformedData: NodeComparisonData[] = apiResponse.data.nodes.map((node) => ({
        node_id: node.node_id,
        node_name: node.name,
        region: node.region,
        isp: node.isp,
        data: node.metrics[selectedMetrics[0]]?.data_points || [],
      }))

      setComparisonData(transformedData)

      if (selectedNodeIds.length >= 3) {
        setDiagnosisLoading(true)
        try {
          const diagRes = await fetchDiagnosis(selectedNodeIds)
          setDiagnosisResult(diagRes.data)
        } catch (diagErr) {
          setDiagnosisError(diagErr instanceof Error ? diagErr.message : String(diagErr))
        } finally {
          setDiagnosisLoading(false)
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t('errors.failedToLoad'))
    } finally {
      setIsLoadingComparison(false)
    }
  }

  const metricOptions = [
    { key: 'latency_ms' as MetricType, label: t('metrics.latency'), unit: 'ms' },
    { key: 'packet_loss_rate' as MetricType, label: t('metrics.packetLoss'), unit: '%' },
    { key: 'jitter_ms' as MetricType, label: t('metrics.jitter'), unit: 'ms' },
  ]

  return (
    <div className="space-y-6">
      <PageHeader title={t('nodes.comparison')} subtitle={t('nodes.comparisonDescription')} />

      {error && (
        <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
          <Button variant="link" size="sm" onClick={() => setError(null)}>Dismiss</Button>
        </div>
      )}

      {/* Node Selector */}
      <Card>
        <CardContent className="p-6">
          <h2 className="text-lg font-semibold mb-4">{t('nodes.selectNodes')} (2-5)</h2>

          {isLoadingNodes ? (
            <div className="flex justify-center py-8">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent" />
            </div>
          ) : (
            <>
              {/* Group By */}
              <div className="mb-4">
                <p className="text-sm font-medium text-muted-foreground mb-2">{t('nodes.groupBy')}</p>
                <div className="flex gap-2">
                  {(['none', 'region', 'isp'] as GroupByType[]).map((option) => (
                    <Button
                      key={option}
                      variant={groupBy === option ? 'default' : 'outline'}
                      size="sm"
                      onClick={() => setComparisonGroupBy(option)}
                    >
                      {option === 'none' ? t('nodes.none') : option.charAt(0).toUpperCase() + option.slice(1)}
                    </Button>
                  ))}
                </div>
              </div>

              {/* Node List */}
              <div data-testid="node-selector" className="space-y-2 max-h-64 overflow-y-auto border rounded-lg p-4">
                {nodeOptions.map((node) => (
                  <label key={node.node_id} className="flex items-center p-3 hover:bg-muted/50 rounded cursor-pointer">
                    <input
                      type="checkbox"
                      checked={selectedNodeIds.includes(node.node_id)}
                      onChange={(e) => {
                        const newSelected = e.target.checked
                          ? [...selectedNodeIds, node.node_id]
                          : selectedNodeIds.filter((id) => id !== node.node_id)
                        handleNodeSelectionChange(newSelected)
                      }}
                      disabled={!selectedNodeIds.includes(node.node_id) && selectedNodeIds.length >= 5}
                      className="h-4 w-4 rounded border-input"
                    />
                    <div className="ml-3 flex-1">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">{node.name}</span>
                        <div className="flex items-center gap-2">
                          {node.region && <span className="text-xs text-muted-foreground">{t('nodes.region')}: {node.region}</span>}
                          {node.isp && <span className="text-xs text-muted-foreground">ISP: {node.isp}</span>}
                          <span className={`text-xs font-medium ${
                            node.status === 'online' ? 'text-green-600' : node.status === 'offline' ? 'text-destructive' : 'text-yellow-600'
                          }`}>
                            {t(`status.${node.status}`)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </label>
                ))}
              </div>

              <div className="mt-4 text-sm text-muted-foreground">
                {t('nodes.selectedCount', { count: selectedNodeIds.length, max: 5 })}
                {selectedNodeIds.length > 0 && selectedNodeIds.length < 2 && (
                  <span className="text-destructive ml-2">({t('nodes.selectAtLeast', { count: 2 })})</span>
                )}
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Time Range */}
      <Card>
        <CardContent className="p-6">
          <h2 className="text-lg font-semibold mb-4">{t('nodes.timeRange')}</h2>
          <div className="flex gap-2">
            {(['24h', '7d', '30d'] as const).map((range) => (
              <Button
                key={range}
                variant={timeRange === range ? 'default' : 'outline'}
                size="sm"
                onClick={() => setComparisonTimeRange(range)}
              >
                {range === '24h' ? t('nodes.hours24') : range === '7d' ? t('nodes.days7') : t('nodes.days30')}
              </Button>
            ))}
            <Button
              variant={timeRange === 'custom' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setComparisonTimeRange('custom')}
            >
              {t('nodes.custom')}
            </Button>
          </div>

          {timeRange === 'custom' && (
            <div className="mt-4 grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">{t('nodes.startTime')}</label>
                <input
                  type="datetime-local"
                  value={customTimeRange?.start || ''}
                  onChange={(e) => setComparisonCustomTimeRange({ start: e.target.value, end: customTimeRange?.end || '' })}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-muted-foreground">{t('nodes.endTime')}</label>
                <input
                  type="datetime-local"
                  value={customTimeRange?.end || ''}
                  onChange={(e) => setComparisonCustomTimeRange({ start: customTimeRange?.start || '', end: e.target.value })}
                  className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                />
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Metric Selector */}
      <Card>
        <CardContent className="p-6">
          <h2 className="text-lg font-semibold mb-4">{t('nodes.metricsSelector')}</h2>
          <div className="flex gap-2">
            {metricOptions.map((metric) => (
              <Button
                key={metric.key}
                variant={selectedMetrics.includes(metric.key) ? 'default' : 'outline'}
                size="sm"
                onClick={() => {
                  const newMetrics = selectedMetrics.includes(metric.key)
                    ? selectedMetrics.filter((m) => m !== metric.key)
                    : [...selectedMetrics, metric.key]
                  setComparisonMetrics(newMetrics)
                }}
              >
                {metric.label} ({metric.unit})
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Compare Button */}
      <Button
        data-testid="compare-button"
        className="w-full py-6"
        onClick={handleCompare}
        disabled={selectedNodeIds.length < 2 || selectedNodeIds.length > 5 || selectedMetrics.length === 0 || isLoadingComparison}
      >
        {isLoadingComparison ? t('common.loading') : t('nodes.compareNodes')}
      </Button>

      {/* Diagnosis */}
      {selectedNodeIds.length >= 3 && (
        <Card>
          <CardContent className="p-6">
            <h2 className="text-lg font-semibold mb-1">{t('nodes.serverDiagnosisTitle')}</h2>
            <p className="text-sm text-muted-foreground mb-4">{t('nodes.serverDiagnosisSubtitle')}</p>
            {diagnosisLoading && (
              <div className="flex items-center gap-2 text-muted-foreground">
                <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
                <span>{t('nodes.serverDiagnosisLoading')}</span>
              </div>
            )}
            {!diagnosisLoading && diagnosisError && (
              <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{diagnosisError}</div>
            )}
            {!diagnosisLoading && diagnosisResult && (
              <ProblemDiagnosis
                problemType={mapApiProblemToUi(diagnosisResult.problem_type)}
                confidence={mapApiConfidence(diagnosisResult.confidence)}
                details={`${diagnosisResult.recommendation}\n\n${t('nodes.diagnosisNodesAnalyzed', { count: diagnosisResult.analysis?.nodes_analyzed ?? 0 })}`}
                isExpanded={false}
              />
            )}
            {!diagnosisLoading && !diagnosisError && !diagnosisResult && !comparisonData && (
              <p className="text-sm text-muted-foreground">{t('nodes.serverDiagnosisAfterCompare')}</p>
            )}
          </CardContent>
        </Card>
      )}

      {/* Comparison Chart */}
      {comparisonData && comparisonData.length > 0 && (
        <ComparisonChart
          nodes={comparisonData}
          metric={selectedMetrics[0]}
          timeRange={timeRange === 'custom' ? '24h' : timeRange}
          showStatistics={true}
          highlightDifferences={true}
          groupBy={groupBy}
          isLoading={isLoadingComparison}
        />
      )}

      {!comparisonData && !isLoadingComparison && (
        <Card>
          <CardContent className="py-12 text-center">
            <h3 className="text-sm font-medium">{t('nodes.noComparisonData')}</h3>
            <p className="mt-1 text-sm text-muted-foreground">{t('nodes.noComparisonDataDescription')}</p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
