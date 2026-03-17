import { act, renderHook, waitFor } from '@testing-library/react'
import { StrictMode } from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useWebhooks } from './useWebhooks'
import { useWebhooksStore } from '../stores/webhooksStore'
import * as webhooksApi from '../api/webhooks'

vi.mock('../api/webhooks', () => ({
  fetchWebhooks: vi.fn(),
  createWebhook: vi.fn(),
  updateWebhook: vi.fn(),
  deleteWebhook: vi.fn(),
}))

describe('useWebhooks', () => {
  beforeEach(() => {
    useWebhooksStore.getState().reset()
    vi.clearAllMocks()
  })

  it('loads once on mount and does not loop on rerender', async () => {
    const fetchWebhooksMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>
    fetchWebhooksMock.mockResolvedValue({
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

    const wrapper = ({ children }: { children: React.ReactNode }) => {
      return <StrictMode>{children}</StrictMode>
    }

    const { result, rerender } = renderHook(() => useWebhooks(), { wrapper })

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    rerender()

    await waitFor(() => {
      expect(result.current.webhooks).toHaveLength(1)
    })

    expect(webhooksApi.fetchWebhooks).toHaveBeenCalledTimes(1)
  })

  it('forces a fresh request when reload is called', async () => {
    const fetchWebhooksMock = webhooksApi.fetchWebhooks as unknown as ReturnType<typeof vi.fn>
    fetchWebhooksMock
      .mockResolvedValueOnce({
        data: {
          webhooks: [],
        },
      })
      .mockResolvedValueOnce({
        data: {
          webhooks: [
            {
              id: 'webhook-2',
              url: 'https://example.com/reloaded',
              event_format: { version: '2.0' },
              enabled: true,
              created_at: '2026-03-17T00:00:00Z',
            },
          ],
        },
      })

    const { result } = renderHook(() => useWebhooks())

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false)
    })

    await act(async () => {
      await result.current.reload()
    })

    expect(webhooksApi.fetchWebhooks).toHaveBeenCalledTimes(2)
    expect(result.current.webhooks[0]?.id).toBe('webhook-2')
  })
})
