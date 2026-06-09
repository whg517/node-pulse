import { useEffect, useState } from 'react'
import { fetchHistory } from '../api/data'
import type { HistoryDataDTO } from '../api/types'
import type { DataPoint } from '../components/dashboard/TrendChart'

/**
 * Average multiple node series that share timestamps (same bucket from API).
 */
export function aggregateHistoryAverageByTimestamp(series: HistoryDataDTO[]): DataPoint[] {
  const bucket = new Map<string, { sum: number; count: number }>()
  for (const s of series) {
    for (const dp of s.data_points) {
      const prev = bucket.get(dp.timestamp) ?? { sum: 0, count: 0 }
      prev.sum += dp.value
      prev.count += 1
      bucket.set(dp.timestamp, prev)
    }
  }
  return [...bucket.entries()]
    .sort((a, b) => new Date(a[0]).getTime() - new Date(b[0]).getTime())
    .map(([timestamp, { sum, count }]) => ({
      timestamp,
      value: count > 0 ? sum / count : 0,
    }))
}

export interface UseDashboardHistoryResult {
  latencyTrend: DataPoint[]
  packetLossTrend: DataPoint[]
  isLoading: boolean
}

const MAX_NODES_PER_HISTORY_QUERY = 80

/**
 * Loads last 24h of aggregated (average across nodes) latency and packet loss for dashboard charts.
 */
export function useDashboardHistory(
  nodeIds: string[],
  refreshToken: number
): UseDashboardHistoryResult {
  const [latencyTrend, setLatencyTrend] = useState<DataPoint[]>([])
  const [packetLossTrend, setPacketLossTrend] = useState<DataPoint[]>([])
  const [isLoading, setIsLoading] = useState(false)

  const idKey = [...new Set(nodeIds)].sort().join(',')

  useEffect(() => {
    const cappedIds = idKey ? idKey.split(',').slice(0, MAX_NODES_PER_HISTORY_QUERY) : []
    if (cappedIds.length === 0) {
      setLatencyTrend([])
      setPacketLossTrend([])
      setIsLoading(false)
      return
    }

    let cancelled = false
    setIsLoading(true)

    void (async () => {
      try {
        const end = new Date()
        const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
        const { data } = await fetchHistory({
          node_ids: cappedIds,
          start_time: start.toISOString(),
          end_time: end.toISOString(),
          metrics: ['latency', 'packet_loss_rate'],
          aggregation: '5m',
        })
        if (cancelled) return
        const latSeries = data.filter((s) => s.metric === 'latency')
        const lossSeries = data.filter((s) => s.metric === 'packet_loss_rate')
        setLatencyTrend(aggregateHistoryAverageByTimestamp(latSeries))
        setPacketLossTrend(aggregateHistoryAverageByTimestamp(lossSeries))
      } catch {
        if (!cancelled) {
          setLatencyTrend([])
          setPacketLossTrend([])
        }
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    })()

    return () => {
      cancelled = true
    }
  }, [idKey, refreshToken])

  return { latencyTrend, packetLossTrend, isLoading }
}
