import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchWebhooks,
  createWebhook,
  updateWebhook,
  deleteWebhook,
} from '@/api/webhooks'
import type {
  WebhookDTO,
  CreateWebhookRequest,
  UpdateWebhookRequest,
} from '@/api/webhooks'

export function useWebhooksQuery() {
  return useQuery({
    queryKey: ['webhooks'],
    queryFn: async () => {
      const res = await fetchWebhooks()
      return res.data.webhooks as WebhookDTO[]
    },
    staleTime: 30_000,
  })
}

export function useCreateWebhook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateWebhookRequest) => createWebhook(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    },
  })
}

export function useUpdateWebhook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateWebhookRequest }) =>
      updateWebhook(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    },
  })
}

export function useDeleteWebhook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteWebhook(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks'] })
    },
  })
}
