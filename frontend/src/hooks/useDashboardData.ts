import { useCallback, useEffect, useRef, useState } from 'react'
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

const BASE_POLLING_INTERVAL_MS = 5000
const MAX_BACKOFF_INTERVAL_MS = 60000

/**
 * Custom hook for managing dashboard data
 *
 * Fetches node and metrics data on mount.
 * Handles loading states, errors, and cleanup on unmount.
 *
 * @returns Dashboard data with refetch function
 *
 * @example
 * const { nodes, metrics, isLoading, error, refetch } = useDashboardData()
 */
export function useDashboardData(): UseDashboardDataResult {
  const [data, setData] = useState<DashboardData>({
    nodes: [],
    metrics: [],
    isLoading: true,
    error: null,
    isPolling: false,
  })

  const isMountedRef = useRef(true)
  const pollingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const inFlightFetchRef = useRef<Promise<void> | null>(null)
  const consecutiveFailuresRef = useRef(0)
  const prevDataRef = useRef<{ nodes: NodeDTO[]; metrics: MetricsDTO[] } | null>(null)

  const clearPollingTimer = useCallback(() => {
    if (pollingTimerRef.current) {
      clearTimeout(pollingTimerRef.current)
      pollingTimerRef.current = null
    }
  }, [])

  const getNextPollDelay = useCallback(() => {
    const backoffFactor = 2 ** Math.min(consecutiveFailuresRef.current, 4)
    return Math.min(BASE_POLLING_INTERVAL_MS * backoffFactor, MAX_BACKOFF_INTERVAL_MS)
  }, [])

  const fetchData = useCallback(async () => {
    if (!isMountedRef.current) return
    if (inFlightFetchRef.current) return inFlightFetchRef.current

    const fetchPromise = (async () => {
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

        consecutiveFailuresRef.current = 0

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
        consecutiveFailuresRef.current += 1

        setData(prev => ({
          ...prev,
          isLoading: false,
          error: error as Error,
          isPolling: false,
        }))
      } finally {
        inFlightFetchRef.current = null
      }
    })()

    inFlightFetchRef.current = fetchPromise
    return fetchPromise
  }, [])

  const scheduleNextPoll = useCallback(() => {
    clearPollingTimer()
    if (!isMountedRef.current || document.visibilityState === 'hidden') {
      return
    }

    const delay = getNextPollDelay()
    pollingTimerRef.current = setTimeout(() => {
      void fetchData().finally(() => {
        scheduleNextPoll()
      })
    }, delay)
  }, [clearPollingTimer, fetchData, getNextPollDelay])

  useEffect(() => {
    isMountedRef.current = true

    // Initial fetch
    void fetchData().finally(() => {
      scheduleNextPoll()
    })

    const handleVisibilityChange = () => {
      if (!isMountedRef.current) return
      if (document.visibilityState === 'hidden') {
        clearPollingTimer()
        setData(prev => ({ ...prev, isPolling: false }))
        return
      }

      void fetchData().finally(() => {
        scheduleNextPoll()
      })
    }

    document.addEventListener('visibilitychange', handleVisibilityChange)

    // Cleanup on unmount
    return () => {
      isMountedRef.current = false
      clearPollingTimer()
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [clearPollingTimer, fetchData, scheduleNextPoll])

  return {
    ...data,
    refetch: async () => {
      await fetchData()
      scheduleNextPoll()
    },
  }
}
