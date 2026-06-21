import { create } from 'zustand'
import * as webhooksAPI from '../api/webhooks'
import type {
  CreateWebhookRequest,
  UpdateWebhookRequest,
  WebhookDTO,
  WebhookEventFormat,
} from '../api/webhooks'

export type WebhooksStatus = 'idle' | 'loading' | 'ready' | 'error'

export interface Webhook {
  id: string
  url: string
  eventFormat: WebhookEventFormat
  enabled: boolean
}

export interface WebhooksState {
  webhooks: Webhook[]
  status: WebhooksStatus
  error: Error | null
  hasLoaded: boolean
}

export interface WebhooksActions {
  clearError: () => void
  reset: () => void
  loadWebhooks: (options?: { force?: boolean }) => Promise<void>
  createWebhook: (request: CreateWebhookRequest) => Promise<Webhook>
  updateWebhook: (id: string, request: UpdateWebhookRequest) => Promise<Webhook>
  deleteWebhook: (id: string) => Promise<void>
  testWebhook: (id: string) => Promise<webhooksAPI.TestWebhookResponse>
  toggleWebhookEnabled: (id: string, enabled: boolean) => Promise<Webhook>
  findWebhookById: (id: string) => Webhook | undefined
}

type WebhooksStore = WebhooksState & WebhooksActions

let inFlightLoad: Promise<void> | null = null

const initialState: WebhooksState = {
  webhooks: [],
  status: 'idle',
  error: null,
  hasLoaded: false,
}

function normalizeError(error: unknown): Error {
  return error instanceof Error ? error : new Error('Failed to process webhook request')
}

function mapWebhook(webhook: WebhookDTO): Webhook {
  return {
    id: webhook.id,
    url: webhook.url,
    eventFormat: webhook.event_format,
    enabled: webhook.enabled,
  }
}

export const useWebhooksStore = create<WebhooksStore>((set, get) => ({
  ...initialState,

  clearError: () => {
    set({ error: null })
  },

  reset: () => {
    inFlightLoad = null
    set(initialState)
  },

  loadWebhooks: async (options) => {
    const { force = false } = options ?? {}
    const { hasLoaded, status } = get()

    if (!force && hasLoaded) {
      return
    }

    if (status === 'loading') {
      return inFlightLoad ?? undefined
    }

    if (inFlightLoad) {
      return inFlightLoad
    }

    // Assign inFlightLoad BEFORE calling set() to close the race window.
    // set() synchronously notifies Zustand subscribers which can trigger
    // a React re-render and a second loadWebhooks() call before the IIFE
    // assignment below would otherwise execute. At that point the caller
    // would see status='loading' but inFlightLoad=null, causing the guard
    // on line 81 to return undefined instead of the in-flight promise.
    inFlightLoad = (async () => {
      try {
        set((state) => ({
          status: state.webhooks.length === 0 || force ? 'loading' : state.status,
          error: null,
        }))

        const response = await webhooksAPI.fetchWebhooks()
        const webhooks = (response.data.webhooks || []).map(mapWebhook)

        set({
          webhooks,
          status: 'ready',
          error: null,
          hasLoaded: true,
        })
      } catch (error) {
        const normalizedError = normalizeError(error)
        set({
          status: 'error',
          error: normalizedError,
        })
        throw normalizedError
      } finally {
        inFlightLoad = null
      }
    })()

    return inFlightLoad
  },

  createWebhook: async (request) => {
    const response = await webhooksAPI.createWebhook(request)
    const webhook = mapWebhook(response.data)

    set((state) => ({
      webhooks: [...state.webhooks, webhook],
      status: 'ready',
      error: null,
      hasLoaded: true,
    }))

    return webhook
  },

  updateWebhook: async (id, request) => {
    const response = await webhooksAPI.updateWebhook(id, request)
    const webhook = mapWebhook(response.data)

    set((state) => ({
      webhooks: state.webhooks.map((item) => (item.id === id ? webhook : item)),
      status: 'ready',
      error: null,
      hasLoaded: true,
    }))

    return webhook
  },

  deleteWebhook: async (id) => {
    await webhooksAPI.deleteWebhook(id)

    set((state) => ({
      webhooks: state.webhooks.filter((item) => item.id !== id),
      status: 'ready',
      error: null,
      hasLoaded: true,
    }))
  },

  testWebhook: async (id) => {
    return webhooksAPI.testWebhook(id)
  },

  toggleWebhookEnabled: async (id, enabled) => {
    return get().updateWebhook(id, { enabled })
  },

  findWebhookById: (id) => {
    return get().webhooks.find((webhook) => webhook.id === id)
  },
}))
