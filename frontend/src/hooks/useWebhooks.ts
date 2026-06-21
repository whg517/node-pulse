import { useCallback, useEffect } from 'react'
import { useWebhooksStore } from '../stores/webhooksStore'
import type { CreateWebhookRequest, TestWebhookResponse, UpdateWebhookRequest } from '../api/webhooks'

export interface UseWebhooksResult {
  webhooks: ReturnType<typeof useWebhooksStore.getState>['webhooks']
  isLoading: boolean
  error: Error | null
  reload: () => Promise<void>
  createWebhook: (request: CreateWebhookRequest) => Promise<void>
  updateWebhook: (id: string, request: UpdateWebhookRequest) => Promise<void>
  deleteWebhook: (id: string) => Promise<void>
  testWebhook: (id: string) => Promise<TestWebhookResponse>
  toggleWebhookEnabled: (id: string, enabled: boolean) => Promise<void>
  getWebhookById: (id: string) => ReturnType<typeof useWebhooksStore.getState>['webhooks'][number] | undefined
}

export function useWebhooks(): UseWebhooksResult {
  const webhooks = useWebhooksStore((state) => state.webhooks)
  const status = useWebhooksStore((state) => state.status)
  const error = useWebhooksStore((state) => state.error)
  const loadWebhooks = useWebhooksStore((state) => state.loadWebhooks)
  const createWebhookAction = useWebhooksStore((state) => state.createWebhook)
  const updateWebhookAction = useWebhooksStore((state) => state.updateWebhook)
  const deleteWebhookAction = useWebhooksStore((state) => state.deleteWebhook)
  const testWebhookAction = useWebhooksStore((state) => state.testWebhook)
  const toggleWebhookEnabledAction = useWebhooksStore((state) => state.toggleWebhookEnabled)
  const findWebhookById = useWebhooksStore((state) => state.findWebhookById)

  useEffect(() => {
    void loadWebhooks()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const reload = useCallback(async () => {
    await loadWebhooks({ force: true })
  }, [loadWebhooks])

  const createWebhook = useCallback(async (request: CreateWebhookRequest) => {
    await createWebhookAction(request)
  }, [createWebhookAction])

  const updateWebhook = useCallback(async (id: string, request: UpdateWebhookRequest) => {
    await updateWebhookAction(id, request)
  }, [updateWebhookAction])

  const deleteWebhook = useCallback(async (id: string) => {
    await deleteWebhookAction(id)
  }, [deleteWebhookAction])

  const testWebhook = useCallback(async (id: string) => {
    return testWebhookAction(id)
  }, [testWebhookAction])

  const toggleWebhookEnabled = useCallback(async (id: string, enabled: boolean) => {
    await toggleWebhookEnabledAction(id, enabled)
  }, [toggleWebhookEnabledAction])

  const getWebhookById = useCallback((id: string) => {
    return findWebhookById(id)
  }, [findWebhookById])

  return {
    webhooks,
    isLoading: status === 'loading',
    error,
    reload,
    createWebhook,
    updateWebhook,
    deleteWebhook,
    testWebhook,
    toggleWebhookEnabled,
    getWebhookById,
  }
}
