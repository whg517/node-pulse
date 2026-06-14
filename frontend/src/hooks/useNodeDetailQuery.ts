import { useQuery } from '@tanstack/react-query'
import { fetchNode, fetchNodeStatus, fetchMetrics } from '@/api'
import type { NodeDTO, MetricsDTO } from '@/api/types'

export function useNodeDetailQuery(nodeId: string) {
  return useQuery({
    queryKey: ['node', nodeId],
    queryFn: async () => {
      const [nodeRes, statusRes, metricsRes] = await Promise.all([
        fetchNode(nodeId),
        fetchNodeStatus(nodeId),
        fetchMetrics([nodeId]),
      ])
      return {
        node: nodeRes.data as NodeDTO,
        status: statusRes.data as { status: string; last_heartbeat: string },
        metrics: (metricsRes.data[0] ?? null) as MetricsDTO | null,
      }
    },
    enabled: !!nodeId,
    staleTime: 5_000,
  })
}
