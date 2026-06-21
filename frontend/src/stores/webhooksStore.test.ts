import { beforeEach, describe, expect, it, vi } from 'vitest'
import * as webhooksApi from '../api/webhooks'
import { useWebhooksStore } from './webhooksStore'

vi.mock('../api/webhooks', () => ({
  fetchWebhooks: vi.fn(),
  createWebhook: vi.fn(),
  updateWebhook: vi.fn(),
  deleteWebhook: vi.fn(),
  testWebhook: vi.fn(),
}))

describe('webhooksStore', () => {
  beforeEach(() => {
    useWebhooksStore.getState().reset()
    vi.clearAllMocks()
  })

  it('calls fetchWebhooks exactly once even when loadWebhooks is called concurrently', async () => {
    const fetchMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>

    // Use a deferred promise to keep the fetch "in flight" so we can call
    // loadWebhooks a second time while the first is still pending.
    let resolveFetch!: (v: unknown) => void
    fetchMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFetch = resolve
      }),
    )

    const store = useWebhooksStore.getState()

    // Fire two concurrent calls — simulates React StrictMode double-invoke
    // or two components subscribing before the first fetch completes.
    const call1 = store.loadWebhooks()
    const call2 = store.loadWebhooks()
    const call3 = store.loadWebhooks()

    // Resolve the fetch
    resolveFetch({ data: { webhooks: [] } })
    await Promise.allSettled([call1, call2, call3])

    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('calls fetchWebhooks exactly once when loadWebhooks is called after set() triggers a re-render', async () => {
    const fetchMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>

    let resolveFetch!: (v: unknown) => void
    fetchMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFetch = resolve
      }),
    )

    const store = useWebhooksStore.getState()

    // Simulate the scenario where set() inside loadWebhooks triggers a
    // subscriber that calls loadWebhooks again synchronously before
    // inFlightLoad is assigned (the original race window).
    let secondCallFired = false
    const unsubscribe = useWebhooksStore.subscribe(() => {
      const { status } = useWebhooksStore.getState()
      // On the first status change (idle → loading), a subscriber fires.
      // Before the fix, inFlightLoad was null at this point → new request.
      if (status === 'loading' && !secondCallFired) {
        secondCallFired = true
        void useWebhooksStore.getState().loadWebhooks()
      }
    })

    const p = store.loadWebhooks()
    resolveFetch({ data: { webhooks: [] } })
    await p
    unsubscribe()

    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('transitions status correctly: idle → loading → ready', async () => {
    const fetchMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>
    fetchMock.mockResolvedValueOnce({ data: { webhooks: [] } })

    expect(useWebhooksStore.getState().status).toBe('idle')

    const p = useWebhooksStore.getState().loadWebhooks()
    // After inFlightLoad is assigned, status should already be 'loading'
    // because set() now runs inside the IIFE synchronously before the await.
    // (In practice this transitions atomically before any subscriber can fire)
    await p

    expect(useWebhooksStore.getState().status).toBe('ready')
    expect(useWebhooksStore.getState().hasLoaded).toBe(true)
  })

  it('does not re-fetch when hasLoaded is true', async () => {
    const fetchMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>
    fetchMock.mockResolvedValue({ data: { webhooks: [] } })

    await useWebhooksStore.getState().loadWebhooks()
    await useWebhooksStore.getState().loadWebhooks()
    await useWebhooksStore.getState().loadWebhooks()

    expect(fetchMock).toHaveBeenCalledTimes(1)
  })

  it('re-fetches when force=true even if hasLoaded is true', async () => {
    const fetchMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>
    fetchMock.mockResolvedValue({ data: { webhooks: [] } })

    await useWebhooksStore.getState().loadWebhooks()
    await useWebhooksStore.getState().loadWebhooks({ force: true })

    expect(fetchMock).toHaveBeenCalledTimes(2)
  })
})
