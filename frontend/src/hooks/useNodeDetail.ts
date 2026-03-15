import { useCallback, useEffect, useState, useRef } from 'react'
import { fetchNode, fetchNodeStatus, fetchMetrics } from '../api'
import type { NodeDTO, MetricsDTO } from '../api/types'
import { deepEqual } from '../utils/deepEqual'

export interface NodeDetailData {
  node: NodeDTO | null
  nodeStatus: { status: string; last_heartbeat: string } | null
  metrics: MetricsDTO | null
  isLoading: boolean
  error: Error | null
  isPolling: boolean
}

export interface UseNodeDetailResult extends NodeDetailData {
  refetch: () => Promise<void>
}

const BASE_POLLING_INTERVAL_MS = 5000
const MAX_BACKOFF_INTERVAL_MS = 60000

/**
 * Custom hook for managing node detail data
 *
 * Fetches node details, status, and metrics on mount.
 * Handles loading states, errors, and cleanup on unmount.
 *
 * @param nodeId - Node ID to fetch data for
 * @returns Node detail data with refetch function
 *
 * @example
 * const { node, nodeStatus, metrics, isLoading, error, refetch } = useNodeDetail('node-id')
 */
export function useNodeDetail(
  nodeId: string
): UseNodeDetailResult {
  const [data, setData] = useState<NodeDetailData>({
    node: null,
    nodeStatus: null,
    metrics: null,
    isLoading: true,
    error: null,
    isPolling: false,
  })

  const isMountedRef = useRef(true)
  const pollingTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const inFlightFetchRef = useRef<Promise<void> | null>(null)
  const consecutiveFailuresRef = useRef(0)
  const prevDataRef = useRef<{
    node: NodeDTO | null
    nodeStatus: { status: string; last_heartbeat: string } | null
    metrics: MetricsDTO | null
  } | null>(null)

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
    if (!isMountedRef.current || !nodeId) return
    if (inFlightFetchRef.current) return inFlightFetchRef.current

    const fetchPromise = (async () => {
      setData(prev => ({ ...prev, isLoading: prev.node === null, error: null }))

      try {
        // Parallel fetch for better performance
        const [nodeResponse, statusResponse, metricsResponse] = await Promise.all([
          fetchNode(nodeId),
          fetchNodeStatus(nodeId),
          fetchMetrics([nodeId]),
        ])

        if (!isMountedRef.current) return

        const newNode = nodeResponse.data
        const newStatus = statusResponse.data
        const newMetrics = metricsResponse.data[0] || null

        // Only update state if data has actually changed (deep comparison)
        const prevData = prevDataRef.current
        const nodeChanged = !prevData || !deepEqual(prevData.node, newNode)
        const statusChanged = !prevData || !deepEqual(prevData.nodeStatus, newStatus)
        const metricsChanged = !prevData || !deepEqual(prevData.metrics, newMetrics)

        consecutiveFailuresRef.current = 0

        if (nodeChanged || statusChanged || metricsChanged) {
          prevDataRef.current = {
            node: newNode,
            nodeStatus: newStatus,
            metrics: newMetrics,
          }
          setData({
            node: newNode,
            nodeStatus: newStatus,
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
  }, [nodeId])

  const scheduleNextPoll = useCallback(() => {
    clearPollingTimer()
    if (!isMountedRef.current || !nodeId || document.visibilityState === 'hidden') {
      return
    }

    const delay = getNextPollDelay()
    pollingTimerRef.current = setTimeout(() => {
      void fetchData().finally(() => {
        scheduleNextPoll()
      })
    }, delay)
  }, [clearPollingTimer, fetchData, getNextPollDelay, nodeId])

  useEffect(() => {
    isMountedRef.current = true
    consecutiveFailuresRef.current = 0

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
  }, [nodeId, clearPollingTimer, fetchData, scheduleNextPoll])

  return {
    ...data,
    refetch: async () => {
      await fetchData()
      scheduleNextPoll()
    },
  }
}
