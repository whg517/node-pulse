import { useMutation, useQueryClient } from '@tanstack/react-query'
import { login as apiLogin, logout as apiLogout } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { ACCESS_TOKEN_EXPIRY_MINUTES } from '@/config/constants'
import type { User } from '@/stores/types'

export function useLogin() {
  const setUser = useAuthStore((s) => s.setUser)
  const setAccessToken = useAuthStore((s) => s.setAccessToken)
  const setCsrfToken = useAuthStore((s) => s.setCsrfToken)

  return useMutation({
    mutationFn: async ({ username, password }: { username: string; password: string }) => {
      const response = await apiLogin({ username, password })
      return response.data
    },
    onSuccess: (data) => {
      const user: User = {
        id: data.user_id,
        username: data.username,
        role: data.role,
      }
      setUser(user)
      setAccessToken(data.access_token, ACCESS_TOKEN_EXPIRY_MINUTES * 60 * 1000)
      if (data.csrf_token) {
        setCsrfToken(data.csrf_token)
      }
    },
  })
}

export function useLogout() {
  const clearAuth = useAuthStore((s) => s.clearAuth)
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => apiLogout(),
    onSettled: () => {
      clearAuth()
      queryClient.clear()
    },
  })
}
