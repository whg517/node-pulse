import { useQuery, type UseQueryOptions } from '@tanstack/react-query'
import { useEffect, useRef, useCallback } from 'react'

interface UsePollingQueryOptions<TData, TError>
  extends Omit<UseQueryOptions<TData, TError>, 'refetchInterval'> {
  /** Base polling interval in ms (default 10000) */
  interval?: number
  /** Whether polling is enabled (default true) */
  enabled?: boolean
  /** Max backoff multiplier (default 16x = 4 consecutive failures) */
  maxBackoff?: number
}

export function usePollingQuery<TData, TError = Error>(
  options: UsePollingQueryOptions<TData, TError>,
) {
  const {
    interval = 10_000,
    enabled = true,
    maxBackoff = 16,
    queryKey,
    queryFn,
    ...rest
  } = options

  const failuresRef = useRef(0)

  const getDelay = useCallback(() => {
    const backoff = Math.min(2 ** failuresRef.current, maxBackoff)
    return interval * backoff
  }, [interval, maxBackoff])

  const query = useQuery<TData, TError>({
    queryKey,
    queryFn,
    enabled,
    refetchInterval: enabled ? getDelay : false,
    refetchIntervalInBackground: false,
    ...rest,
  })

  useEffect(() => {
    if (query.isError) {
      failuresRef.current = Math.min(failuresRef.current + 1, 4)
    } else if (query.isSuccess) {
      failuresRef.current = 0
    }
  }, [query.isError, query.isSuccess])

  return query
}
