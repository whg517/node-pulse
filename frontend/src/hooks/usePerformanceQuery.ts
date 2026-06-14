import { useQuery } from '@tanstack/react-query'
import { fetchPerformanceData } from '@/api/performance'
import type { PerformanceAPIResponse } from '@/api/performance'

interface PerformanceParams {
  nodeIds: string[]
  timeRange: string
  metrics?: string[]
}

export function usePerformanceQuery({ nodeIds, timeRange, metrics }: PerformanceParams) {
  return useQuery({
    queryKey: ['performance', nodeIds, timeRange, metrics],
    queryFn: async () => {
      const res = await fetchPerformanceData(timeRange)
      return res.data as PerformanceAPIResponse['data']
    },
    enabled: nodeIds.length > 0,
    staleTime: 30_000,
  })
}
