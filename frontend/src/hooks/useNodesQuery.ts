import { usePollingQuery } from './usePollingQuery'
import { fetchNodes, fetchMetrics } from '@/api'
import type { NodeDTO, MetricsDTO } from '@/api/types'

export function useNodesQuery(interval = 10_000) {
  return usePollingQuery<{ nodes: NodeDTO[]; metrics: MetricsDTO[] }>({
    queryKey: ['dashboard', 'nodes'],
    queryFn: async () => {
      const [nodesRes, metricsRes] = await Promise.all([
        fetchNodes(),
        fetchMetrics([]),
      ])
      return {
        nodes: nodesRes.data.nodes ?? [],
        metrics: metricsRes.data ?? [],
      }
    },
    interval,
  })
}
