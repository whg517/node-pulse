import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  fetchAlertRules,
  createAlertRule,
  updateAlertRule,
  deleteAlertRule,
  fetchAlertRecords,
} from '@/api/alerts'
import type {
  AlertRuleDTO,
  CreateAlertRuleRequest,
  UpdateAlertRuleRequest,
  AlertRecordDTO,
  AlertRecordFilters,
} from '@/api/types'

export function useAlertRulesQuery() {
  return useQuery({
    queryKey: ['alerts', 'rules'],
    queryFn: async () => {
      const res = await fetchAlertRules()
      return res.data.alerts as AlertRuleDTO[]
    },
    staleTime: 30_000,
  })
}

export function useAlertRecordsQuery(filters?: AlertRecordFilters) {
  return useQuery({
    queryKey: ['alerts', 'records', filters],
    queryFn: async () => {
      const res = await fetchAlertRecords(filters)
      return res.data as AlertRecordDTO[]
    },
    staleTime: 15_000,
  })
}

export function useCreateAlertRule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateAlertRuleRequest) => createAlertRule(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts', 'rules'] })
    },
  })
}

export function useUpdateAlertRule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAlertRuleRequest }) =>
      updateAlertRule(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts', 'rules'] })
    },
  })
}

export function useDeleteAlertRule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteAlertRule(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts', 'rules'] })
    },
  })
}
