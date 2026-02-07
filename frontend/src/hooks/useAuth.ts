import { useAuthStore } from '../stores/authStore'
import { login, logout, type LoginRequest, type LoginResponse, type LogoutResponse } from '../api/auth'

export function useAuth() {
  const {
    isAuthenticated,
    user,
    role,
    tokenExpiresAt,
    login: storeLogin,
    logout: storeLogout,
  } = useAuthStore()

  const handleLogin = async (credentials: LoginRequest): Promise<LoginResponse> => {
    const response = await login(credentials)
    // The storeLogin now handles the state update
    await storeLogin(credentials.username, credentials.password)
    return response
  }

  const handleLogout = async (): Promise<LogoutResponse> => {
    const response = await logout()
    // The storeLogout now handles the state update
    await storeLogout()
    return response
  }

  const isValidSession = (): boolean => {
    return tokenExpiresAt !== null && tokenExpiresAt > Date.now()
  }

  return {
    isAuthenticated,
    userId: user?.id || null,
    username: user?.username || null,
    role,
    tokenExpiresAt,
    login: handleLogin,
    logout: handleLogout,
    isValidSession,
  }
}
