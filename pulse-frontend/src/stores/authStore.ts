import { create } from 'zustand'

export interface AuthState {
  isAuthenticated: boolean
  userId: string | null
  username: string | null
  role: 'admin' | 'operator' | 'viewer' | null
  sessionExpiry: number | null
}

export interface AuthActions {
  setSession: (userId: string, username: string, role: 'admin' | 'operator' | 'viewer', expiry: number) => void
  clearSession: () => void
  checkSession: () => boolean
}

type AuthStore = AuthState & AuthActions

const SESSION_COOKIE_NAME = 'session_id'
const SESSION_EXPIRY_HOURS = 24

export const useAuthStore = create<AuthStore>((set, get) => ({
  isAuthenticated: false,
  userId: null,
  username: null,
  role: null,
  sessionExpiry: null,

  setSession: (userId, username, role, expiry) => {
    const sessionData = {
      isAuthenticated: true,
      userId,
      username,
      role,
      sessionExpiry: expiry,
    }
    set(sessionData)
  },

  clearSession: () => {
    set({
      isAuthenticated: false,
      userId: null,
      username: null,
      role: null,
      sessionExpiry: null,
    })
  },

  checkSession: () => {
    const state = get()
    if (!state.sessionExpiry) return false

    const now = Date.now()
    const isValid = state.isAuthenticated && state.sessionExpiry > now

    if (!isValid) {
      set({
        isAuthenticated: false,
        userId: null,
        username: null,
        role: null,
        sessionExpiry: null,
      })
    }

    return isValid
  },
}))

export { SESSION_COOKIE_NAME, SESSION_EXPIRY_HOURS }
