import { useState, useEffect, useCallback, useRef } from 'react'
import { fetchPerformanceData, type PerformanceDataResponse } from '../api/performance'

interface UsePerformanceDataOptions {
  /**
   * Polling interval in milliseconds (default: 60000ms = 1 minute)
   */
  pollingInterval?: number

  /**
   * Whether to enable automatic polling (default: true)
   */
  enablePolling?: boolean

  /**
   * Time range for performance data (default: "24h")
   */
  timeRange?: string
}

interface UsePerformanceDataResult {
  data: PerformanceDataResponse | null
  isLoading: boolean
  error: Error | null
  refetch: () => Promise<void>
  isPolling: boolean
}

/**
 * Custom hook for fetching and polling performance data
 *
 * Automatically polls performance metrics every 60 seconds by default
 * and provides manual refresh capability.
 */
export function usePerformanceData(
  options: UsePerformanceDataOptions = {}
): UsePerformanceDataResult {
  const {
    pollingInterval = 60000,
    enablePolling = true,
    timeRange = '24h',
  } = options

  const [data, setData] = useState<PerformanceDataResponse | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [isPolling, setIsPolling] = useState(false)

  // Use ref to store interval ID for cleanup
  const intervalRef = useRef<number | null>(null)

  // Use ref to store the latest timeRange without causing callback recreation
  const timeRangeRef = useRef(timeRange)

  // Update timeRange ref when it changes
  useEffect(() => {
    timeRangeRef.current = timeRange
  }, [timeRange])

  const fetchData = useCallback(async () => {
    try {
      setError(null)
      const response = await fetchPerformanceData(timeRangeRef.current)
      setData(response.data)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to fetch performance data')
      setError(error)
      console.error('Error fetching performance data:', err)
    } finally {
      setIsLoading(false)
    }
  }, []) // Empty deps - fetchData only depends on timeRangeRef

  // Start polling
  const startPolling = useCallback(() => {
    if (!enablePolling || intervalRef.current !== null) {
      return
    }

    setIsPolling(true)
    intervalRef.current = window.setInterval(() => {
      fetchData()
    }, pollingInterval)
  }, [enablePolling, pollingInterval, fetchData])

  // Stop polling
  const stopPolling = useCallback(() => {
    if (intervalRef.current !== null) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
      setIsPolling(false)
    }
  }, [])

  // Initial data fetch and setup polling
  useEffect(() => {
    setIsLoading(true)
    fetchData()

    // Start polling after initial fetch
    startPolling()

    // Cleanup on unmount
    return () => {
      stopPolling()
    }
  }, [fetchData, startPolling, stopPolling])

  // Manual refetch function
  const refetch = useCallback(async () => {
    setIsLoading(true)
    await fetchData()
    // Restart polling to reset interval
    stopPolling()
    startPolling()
  }, [fetchData, stopPolling, startPolling])

  return {
    data,
    isLoading,
    error,
    refetch,
    isPolling,
  }
}
