import { useEffect, useState, useRef } from 'react'
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

/**
 * Custom hook for managing node detail data with automatic polling
 *
 * Fetches node details, status, and metrics on mount and polls every 5 seconds.
 * Handles loading states, errors, and cleanup on unmount.
 *
 * @param nodeId - Node ID to fetch data for
 * @param pollingInterval - Polling interval in milliseconds (default: 5000ms)
 * @returns Node detail data with refetch function
 *
 * @example
 * const { node, nodeStatus, metrics, isLoading, error, refetch } = useNodeDetail('node-id')
 */
export function useNodeDetail(
  nodeId: string,
  pollingInterval = 5000
): UseNodeDetailResult {
  const [data, setData] = useState<NodeDetailData>({
    node: null,
    nodeStatus: null,
    metrics: null,
    isLoading: true,
    error: null,
    isPolling: false,
  })

  const pollingIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const isMountedRef = useRef(true)
  const prevDataRef = useRef<{
    node: NodeDTO | null
    nodeStatus: { status: string; last_heartbeat: string } | null
    metrics: MetricsDTO | null
  } | null>(null)

  const fetchData = async () => {
    if (!isMountedRef.current || !nodeId) return

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
    if (pollingInterval > 0) {
      pollingIntervalRef.current = setInterval(() => {
        fetchData()
      }, pollingInterval)
    }

    // Cleanup on unmount
    return () => {
      isMountedRef.current = false
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current)
        pollingIntervalRef.current = null
      }
    }
  }, [nodeId, pollingInterval])

  return {
    ...data,
    refetch: fetchData,
  }
}
