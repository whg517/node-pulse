import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useWebhooksStore } from '../webhooksStore'
import * as webhooksApi from '../../api/webhooks'

vi.mock('../../api/webhooks', () => ({
  fetchWebhooks: vi.fn(),
  createWebhook: vi.fn(),
  updateWebhook: vi.fn(),
  deleteWebhook: vi.fn(),
}))

describe('useWebhooksStore', () => {
  beforeEach(() => {
    useWebhooksStore.getState().reset()
    vi.clearAllMocks()
  })

  it('deduplicates concurrent webhook loads', async () => {
    let resolveFetch: ((value: unknown) => void) | undefined
    const fetchPromise = new Promise((resolve) => {
      resolveFetch = resolve
    })

    const fetchWebhooksMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>
    fetchWebhooksMock.mockReturnValue(fetchPromise)

    const firstLoad = useWebhooksStore.getState().loadWebhooks()
    const secondLoad = useWebhooksStore.getState().loadWebhooks()

    expect(webhooksApi.fetchWebhooks).toHaveBeenCalledTimes(1)

    resolveFetch?.({
      data: {
        webhooks: [
          {
            id: 'webhook-1',
            url: 'https://example.com/webhook',
            event_format: { version: '1.0' },
            enabled: true,
            created_at: '2026-03-17T00:00:00Z',
          },
        ],
      },
    })

    await Promise.all([firstLoad, secondLoad])

    const state = useWebhooksStore.getState()
    expect(state.webhooks).toHaveLength(1)
    expect(state.hasLoaded).toBe(true)
    expect(state.status).toBe('ready')
  })

  it('updates local state directly for create, update, and delete', async () => {
    useWebhooksStore.setState({
      webhooks: [
        {
          id: 'webhook-1',
          url: 'https://example.com/webhook',
          eventFormat: { version: '1.0' },
          enabled: true,
        },
      ],
      status: 'ready',
      error: null,
      hasLoaded: true,
    })

    const createWebhookMock = webhooksApi.createWebhook as unknown as ReturnType<typeof vi.fn>
    const updateWebhookMock = webhooksApi.updateWebhook as unknown as ReturnType<typeof vi.fn>
    const deleteWebhookMock = webhooksApi.deleteWebhook as unknown as ReturnType<typeof vi.fn>

    createWebhookMock.mockResolvedValue({
      data: {
        id: 'webhook-2',
        url: 'https://example.com/new-webhook',
        event_format: { version: '2.0' },
        enabled: true,
        created_at: '2026-03-17T00:00:00Z',
      },
    })

    updateWebhookMock.mockResolvedValue({
      data: {
        id: 'webhook-1',
        url: 'https://example.com/webhook-updated',
        event_format: { version: '1.1' },
        enabled: false,
        created_at: '2026-03-17T00:00:00Z',
      },
    })

    deleteWebhookMock.mockResolvedValue({
      message: 'Deleted',
    })

    await useWebhooksStore.getState().createWebhook({
      url: 'https://example.com/new-webhook',
      event_format: { version: '2.0' },
      enabled: true,
    })

    expect(useWebhooksStore.getState().webhooks).toHaveLength(2)

    await useWebhooksStore.getState().toggleWebhookEnabled('webhook-1', false)

    expect(useWebhooksStore.getState().findWebhookById('webhook-1')).toEqual({
      id: 'webhook-1',
      url: 'https://example.com/webhook-updated',
      eventFormat: { version: '1.1' },
      enabled: false,
    })

    await useWebhooksStore.getState().deleteWebhook('webhook-2')

    expect(useWebhooksStore.getState().webhooks).toHaveLength(1)
    expect(webhooksApi.fetchWebhooks).not.toHaveBeenCalled()
  })
})
