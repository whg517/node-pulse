import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAuth } from './useAuth'

// Mock the auth store with a proper Zustand-like implementation
const mockLogin = vi.fn().mockResolvedValue(undefined)
const mockLogout = vi.fn().mockResolvedValue(undefined)
const mockSetUser = vi.fn()
const mockClearAuth = vi.fn()
const mockSetAccessToken = vi.fn()
const mockRestoreSession = vi.fn().mockResolvedValue(undefined)

const createMockState = (overrides: Record<string, any> = {}) => ({
  user: null,
  isAuthenticated: false,
  role: null,
  accessToken: null,
  tokenExpiresAt: null,
  isLoading: false,
  refreshFailureCount: 0,
  login: mockLogin,
  logout: mockLogout,
  setUser: mockSetUser,
  clearAuth: mockClearAuth,
  setAccessToken: mockSetAccessToken,
  restoreSession: mockRestoreSession,
  ...overrides,
})

// Store the current mock state
let currentMockState = createMockState()

vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn((selector) => {
    // If selector is provided, call it with state (Zustand behavior)
    // If no selector, return the whole state (for direct access)
    if (selector) {
      return selector(currentMockState)
    }
    return currentMockState
  }),
}))

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Reset mock state before each test
    currentMockState = createMockState()
  })

  it('should return initial authentication state', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.userId).toBeNull()
    expect(result.current.username).toBeNull()
    expect(result.current.role).toBeNull()
    expect(result.current.isLoading).toBe(false)
  })

  it('should return authenticated state when logged in', () => {
    currentMockState = createMockState({
      isAuthenticated: true,
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin' as const,
      },
      role: 'admin' as const,
      accessToken: 'test-token',
      tokenExpiresAt: Date.now() + 3600000,
    })

    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.userId).toBe('user-123')
    expect(result.current.username).toBe('testuser')
    expect(result.current.role).toBe('admin')
  })

  it('should call store login on login', async () => {
    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.login({ username: 'testuser', password: 'password123' })
    })

    expect(mockLogin).toHaveBeenCalledWith('testuser', 'password123')
    expect(mockLogin).toHaveBeenCalledTimes(1)
  })

  it('should call store logout on logout', async () => {
    const { result } = renderHook(() => useAuth())

    await act(async () => {
      await result.current.logout()
    })

    expect(mockLogout).toHaveBeenCalledTimes(1)
  })

  it('should validate session correctly', () => {
    // Test invalid session (no token)
    currentMockState = createMockState()

    const { result: result1 } = renderHook(() => useAuth())
    expect(result1.current.isValidSession()).toBe(false)

    // Test valid session
    currentMockState = createMockState({
      isAuthenticated: true,
      user: { id: '1', username: 'test', role: 'admin' as const },
      role: 'admin' as const,
      accessToken: 'token',
      tokenExpiresAt: Date.now() + 3600000,
    })

    const { result: result2 } = renderHook(() => useAuth())
    expect(result2.current.isValidSession()).toBe(true)

    // Test expired session
    currentMockState = createMockState({
      isAuthenticated: true,
      user: { id: '1', username: 'test', role: 'admin' as const },
      role: 'admin' as const,
      accessToken: 'token',
      tokenExpiresAt: Date.now() - 1000,
    })

    const { result: result3 } = renderHook(() => useAuth())
    expect(result3.current.isValidSession()).toBe(false)
  })

  it('should return loading state', () => {
    currentMockState = createMockState({
      isLoading: true,
    })

    const { result } = renderHook(() => useAuth())

    expect(result.current.isLoading).toBe(true)
  })
})
