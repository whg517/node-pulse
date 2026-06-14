import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { exportData } from '@/api/data'
import type { ExportQueryDTO } from '@/api/types'

export function useExportQuery(params: ExportQueryDTO | null) {
  return useQuery({
    queryKey: ['export', params],
    queryFn: async () => {
      if (!params) return null
      const res = await exportData(params)
      return res.data as { download_url: string }
    },
    enabled: !!params,
    staleTime: 60_000,
  })
}

export function useCreateExport() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: ExportQueryDTO) => exportData(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['export'] })
    },
  })
}
