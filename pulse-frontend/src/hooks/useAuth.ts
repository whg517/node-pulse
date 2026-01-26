import { useAuthStore } from '../stores/authStore'
import { login, logout, type LoginRequest, type LoginResponse, type LogoutResponse } from '../api/auth'

export function useAuth() {
  const { setSession, clearSession, isAuthenticated, userId, username, role, sessionExpiry, checkSession } = useAuthStore()

  const handleLogin = async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await login(credentials)
    const expiryTime = Date.now() + 24 * 60 * 60 * 1000 // 24 hours
    setSession(response.data.user_id, response.data.username, response.data.role, expiryTime)
    return response
  }

  const handleLogout = async (): Promise<LogoutResponse> => {
    const response = await logout()
    clearSession()
    return response
  }

  const isValidSession = (): boolean => {
    return checkSession()
  }

  return {
    isAuthenticated,
    userId,
    username,
    role,
    sessionExpiry,
    login: handleLogin,
    logout: handleLogout,
    isValidSession,
  }
}
