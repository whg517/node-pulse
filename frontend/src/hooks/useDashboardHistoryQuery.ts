import { useQuery } from '@tanstack/react-query'
import { fetchHistory } from '@/api/data'
import type { HistoryDataDTO } from '@/api/types'

export interface DataPoint {
  timestamp: string
  value: number
}

export function aggregateHistoryAverage(series: HistoryDataDTO[]): DataPoint[] {
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

export function useDashboardHistoryQuery(nodeIds: string[]) {
  return useQuery({
    queryKey: ['dashboard', 'history', nodeIds.sort()],
    queryFn: async () => {
      if (nodeIds.length === 0) {
        return { latencyTrend: [], packetLossTrend: [] }
      }
      const cappedIds = nodeIds.slice(0, 80)
      const end = new Date()
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
      const { data } = await fetchHistory({
        node_ids: cappedIds,
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        metrics: ['latency', 'packet_loss_rate'],
        aggregation: '5m',
      })
      const latSeries = data.filter((s) => s.metric === 'latency')
      const lossSeries = data.filter((s) => s.metric === 'packet_loss_rate')
      return {
        latencyTrend: aggregateHistoryAverage(latSeries),
        packetLossTrend: aggregateHistoryAverage(lossSeries),
      }
    },
    staleTime: 60_000,
    enabled: nodeIds.length > 0,
  })
}
