import { renderHook, waitFor, act } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { usePollingQuery } from '../usePollingQuery'
import { createQueryWrapper } from '@/test/utils'

describe('usePollingQuery', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches data on mount', async () => {
    const queryFn = vi.fn().mockResolvedValue({ ok: true })
    const { wrapper } = createQueryWrapper()

    const { result } = renderHook(
      () =>
        usePollingQuery<{ ok: boolean }, Error>({
          queryKey: ['test'],
          queryFn,
          interval: 10_000,
        }),
      { wrapper },
    )

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true)
    })
    expect(result.current.data).toEqual({ ok: true })
    expect(queryFn).toHaveBeenCalledTimes(1)
  })

  it('refetches at the base interval after success', async () => {
    vi.useFakeTimers()
    try {
      const queryFn = vi.fn().mockResolvedValue({ n: 1 })
      const { wrapper } = createQueryWrapper()

      renderHook(
        () =>
          usePollingQuery({ queryKey: ['poll'], queryFn, interval: 5_000 }),
        { wrapper },
      )

      // initial fetch
      await act(async () => {
        await Promise.resolve()
      })
      expect(queryFn).toHaveBeenCalledTimes(1)

      // 3s < 5s interval -> no refetch yet
      await act(async () => {
        await vi.advanceTimersByTimeAsync(3_000)
      })
      expect(queryFn).toHaveBeenCalledTimes(1)

      // advance past the 5s interval -> refetch fires
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2_000)
      })
      expect(queryFn).toHaveBeenCalledTimes(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it('does not poll when enabled is false', async () => {
    vi.useFakeTimers()
    try {
      const queryFn = vi.fn().mockResolvedValue({ ok: true })
      const { wrapper } = createQueryWrapper()

      const { result } = renderHook(
        () =>
          usePollingQuery({ queryKey: ['disabled'], queryFn, interval: 5_000, enabled: false }),
        { wrapper },
      )

      await act(async () => {
        await vi.advanceTimersByTimeAsync(15_000)
      })
      expect(queryFn).not.toHaveBeenCalled()
      expect(result.current.fetchStatus).toBe('idle')
    } finally {
      vi.useRealTimers()
    }
  })

  it('recovers to success state after an initial failure', async () => {
    vi.useFakeTimers()
    try {
      const queryFn = vi
        .fn()
        .mockRejectedValueOnce(new Error('boom'))
        .mockResolvedValue({ ok: true })
      const { wrapper } = createQueryWrapper()

      const { result } = renderHook(
        () =>
          usePollingQuery({ queryKey: ['recover'], queryFn, interval: 5_000 }),
        { wrapper },
      )

      // initial (failing) fetch
      await act(async () => {
        await Promise.resolve()
      })
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0)
      })
      expect(result.current.isError).toBe(true)

      // advance far enough that the backed-off refetch fires (up to 2^4 * 5s)
      await act(async () => {
        await vi.advanceTimersByTimeAsync(80_000)
      })
      expect(queryFn.mock.calls.length).toBeGreaterThan(1)
      expect(result.current.isSuccess).toBe(true)
      expect(result.current.data).toEqual({ ok: true })
    } finally {
      vi.useRealTimers()
    }
  })

  it('exposes error state on fetch failure', async () => {
    vi.useFakeTimers()
    try {
      const error = new Error('network down')
      const queryFn = vi.fn().mockRejectedValue(error)
      const { wrapper } = createQueryWrapper()

      const { result } = renderHook(
        () => usePollingQuery<unknown, Error>({ queryKey: ['err'], queryFn, interval: 60_000 }),
        { wrapper },
      )

      await act(async () => {
        await Promise.resolve()
      })
      await act(async () => {
        await vi.advanceTimersByTimeAsync(0)
      })
      expect(result.current.isError).toBe(true)
      expect(result.current.error).toEqual(error)
    } finally {
      vi.useRealTimers()
    }
  })
})
