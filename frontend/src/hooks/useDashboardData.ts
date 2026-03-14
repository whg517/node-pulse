import { useEffect, useState, useRef } from 'react'
import { fetchNodes, fetchMetrics } from '../api'
import type { NodeDTO, MetricsDTO } from '../api/types'
import { deepEqual } from '../utils/deepEqual'

export interface DashboardData {
  nodes: NodeDTO[]
  metrics: MetricsDTO[]
  isLoading: boolean
  error: Error | null
  isPolling: boolean
}

export interface UseDashboardDataResult extends DashboardData {
  refetch: () => Promise<void>
}

/**
 * Custom hook for managing dashboard data with automatic polling
 *
 * Fetches node and metrics data on mount and polls every 5 seconds.
 * Handles loading states, errors, and cleanup on unmount.
 *
 * @param pollingInterval - Polling interval in milliseconds (default: 5000ms)
 * @returns Dashboard data with refetch function
 *
 * @example
 * const { nodes, metrics, isLoading, error, refetch } = useDashboardData()
 */
export function useDashboardData(pollingInterval = 5000): UseDashboardDataResult {
  const [data, setData] = useState<DashboardData>({
    nodes: [],
    metrics: [],
    isLoading: true,
    error: null,
    isPolling: false,
  })

  const pollingIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const isMountedRef = useRef(true)
  const prevDataRef = useRef<{ nodes: NodeDTO[]; metrics: MetricsDTO[] } | null>(null)

  const fetchData = async () => {
    if (!isMountedRef.current) return

    setData(prev => ({ ...prev, isLoading: prev.nodes.length === 0, error: null }))

    try {
      // Parallel fetch for better performance
      const [nodesResponse, metricsResponse] = await Promise.all([
        fetchNodes(),
        fetchMetrics([]), // empty array = fetch all nodes
      ])

      if (!isMountedRef.current) return

      const newNodes = nodesResponse.data.nodes ?? []
      const newMetrics = metricsResponse.data ?? []

      // Only update state if data has actually changed (deep comparison)
      const prevData = prevDataRef.current
      const nodesChanged = !prevData || !deepEqual(prevData.nodes, newNodes)
      const metricsChanged = !prevData || !deepEqual(prevData.metrics, newMetrics)

      if (nodesChanged || metricsChanged) {
        prevDataRef.current = { nodes: newNodes, metrics: newMetrics }
        setData({
          nodes: newNodes,
          metrics: newMetrics,
          isLoading: false,
          error: null,
          isPolling: true,
        })
      } else {
        // Data unchanged, just ensure loading state is false
        setData(prev => ({ ...prev, isLoading: false, isPolling: true }))
      }
    } catch (error) {
      if (!isMountedRef.current) return

      setData(prev => ({
        ...prev,
        isLoading: false,
        error: error as Error,
        isPolling: false,
      }))
    }
  }

  useEffect(() => {
    isMountedRef.current = true

    // Initial fetch
    fetchData()

    // Set up polling
    pollingIntervalRef.current = setInterval(() => {
      fetchData()
    }, pollingInterval)

    // Cleanup on unmount
    return () => {
      isMountedRef.current = false
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current)
        pollingIntervalRef.current = null
      }
    }
  }, [pollingInterval])

  return {
    ...data,
    refetch: fetchData,
  }
}
