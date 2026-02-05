import { create } from 'zustand'
import * as webhooksAPI from '../api/webhooks'

export interface Webhook {
  id: string
  url: string
  eventFormat: Record<string, any>
  enabled: boolean
}

export interface WebhooksState {
  webhooks: Webhook[]
}

export interface WebhooksActions {
  setWebhooks: (webhooks: Webhook[]) => void
  addWebhook: (webhook: Webhook) => void
  updateWebhook: (id: string, updates: Partial<Webhook>) => void
  removeWebhook: (id: string) => void
  fetchWebhooks: () => Promise<void>
}

type WebhooksStore = WebhooksState & WebhooksActions

export const useWebhooksStore = create<WebhooksStore>((set) => ({
  // State
  webhooks: [],

  // Actions
  setWebhooks: (webhooks: Webhook[]) => {
    set({ webhooks })
  },

  addWebhook: (webhook: Webhook) => {
    set((state) => ({
      webhooks: [...state.webhooks, webhook],
    }))
  },

  updateWebhook: (id: string, updates: Partial<Webhook>) => {
    set((state) => ({
      webhooks: state.webhooks.map((w) =>
        w.id === id ? { ...w, ...updates } : w
      ),
    }))
  },

  removeWebhook: (id: string) => {
    set((state) => ({
      webhooks: state.webhooks.filter((w) => w.id !== id),
    }))
  },

  fetchWebhooks: async () => {
    try {
      const response = await webhooksAPI.fetchWebhooks()

      const webhooks: Webhook[] = response.data.map((webhook) => ({
        id: webhook.id,
        url: webhook.url,
        eventFormat: webhook.event_format,
        enabled: webhook.enabled,
      }))

      set({ webhooks })
    } catch (error) {
      console.error('Failed to fetch webhooks:', error)
      throw error
    }
  },
}))
