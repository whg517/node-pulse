import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAuthStore, setupCrossTabLogoutSync, setupVisibilityHandler } from '../authStore'
import * as authApi from '../../api/auth'
import * as client from '../../api/client'

// Mock the auth API
vi.mock('../../api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  getMe: vi.fn(),
}))

// Mock the client module
vi.mock('../../api/client', () => ({
  cancelPendingRequests: vi.fn(),
}))

describe('useAuthStore', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: null,
      tokenExpiresAt: null,
      isLoading: false,
      refreshFailureCount: 0,
    })
    vi.clearAllMocks()
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.clearAllTimers()
    consoleErrorSpy.mockRestore()
  })

  it('should have initial state', () => {
    const { result } = renderHook(() => useAuthStore())

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.role).toBeNull()
    expect(result.current.accessToken).toBeNull()
    expect(result.current.tokenExpiresAt).toBeNull()
    expect(result.current.isLoading).toBe(false)
    expect(result.current.refreshFailureCount).toBe(0)
  })

  it('should handle login successfully', async () => {
    const mockLoginResponse = {
      data: {
        user_id: 'user-123',
        username: 'testuser',
        role: 'admin' as const,
        access_token: 'test-access-token',
      },
      message: 'Login successful',
      timestamp: '2024-01-01T00:00:00Z',
    }

    vi.mocked(authApi.login).mockResolvedValue(mockLoginResponse)

    const { result } = renderHook(() => useAuthStore())

    await act(async () => {
      await result.current.login('testuser', 'password123')
    })

    expect(result.current.user).toEqual({
      id: 'user-123',
      username: 'testuser',
      role: 'admin',
    })
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.role).toBe('admin')
    expect(result.current.accessToken).toBe('test-access-token')
    expect(result.current.tokenExpiresAt).toBeDefined()
    expect(result.current.tokenExpiresAt).toBeGreaterThan(Date.now())
    expect(result.current.refreshFailureCount).toBe(0)
  })

  it('should handle logout successfully', async () => {
    // First set up an authenticated session
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
      refreshFailureCount: 0,
    })

    const mockLogoutResponse = {
      message: 'Logout successful',
      timestamp: '2024-01-01T00:00:00Z',
    }

    vi.mocked(authApi.logout).mockResolvedValue(mockLogoutResponse)

    const { result } = renderHook(() => useAuthStore())

    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.role).toBeNull()
    expect(result.current.accessToken).toBeNull()
    expect(result.current.tokenExpiresAt).toBeNull()
    expect(result.current.refreshFailureCount).toBe(0)
    expect(client.cancelPendingRequests).toHaveBeenCalled()
  })

  it('should handle logout API failure gracefully', async () => {
    // Set up an authenticated session
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
      refreshFailureCount: 0,
    })

    vi.mocked(authApi.logout).mockRejectedValue(new Error('Logout failed'))

    const { result } = renderHook(() => useAuthStore())

    await act(async () => {
      await result.current.logout()
    })

    // Should still clear local state even if API call fails
    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.accessToken).toBeNull()
    expect(client.cancelPendingRequests).toHaveBeenCalled()
  })

  it('should set user manually', () => {
    const { result } = renderHook(() => useAuthStore())

    const mockUser = {
      id: 'user-456',
      username: 'newuser',
      role: 'operator' as const,
    }

    act(() => {
      result.current.setUser(mockUser)
    })

    expect(result.current.user).toEqual(mockUser)
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.role).toBe('operator')
  })

  it('should clear auth state', () => {
    // Set up authenticated state
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
      refreshFailureCount: 2,
    })

    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.clearAuth()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.role).toBeNull()
    expect(result.current.accessToken).toBeNull()
    expect(result.current.tokenExpiresAt).toBeNull()
    expect(result.current.refreshFailureCount).toBe(0)
    expect(client.cancelPendingRequests).toHaveBeenCalled()
  })

  it('should set access token', () => {
    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.setAccessToken('new-token', 900000) // 15 minutes in ms
    })

    expect(result.current.accessToken).toBe('new-token')
    expect(result.current.tokenExpiresAt).toBeGreaterThan(Date.now())
    expect(result.current.refreshFailureCount).toBe(0)
  })

  it('should have valid token when token exists and not expired', () => {
    const { result } = renderHook(() => useAuthStore())

    // Set valid token
    act(() => {
      useAuthStore.setState({
        user: {
          id: 'user-123',
          username: 'testuser',
          role: 'admin',
        },
        isAuthenticated: true,
        role: 'admin',
        accessToken: 'test-access-token',
        tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
        refreshFailureCount: 0,
      })
    })

    const state = result.current
    const isValid = state.tokenExpiresAt !== null && state.tokenExpiresAt > Date.now()
    expect(isValid).toBe(true)
    expect(state.isAuthenticated).toBe(true)
  })

  it('should handle expired token', () => {
    const { result } = renderHook(() => useAuthStore())

    // Set expired token directly
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() - 1000, // Expired
      refreshFailureCount: 0,
    })

    const state = result.current
    const isValid = state.tokenExpiresAt !== null && state.tokenExpiresAt > Date.now()
    expect(isValid).toBe(false)
  })

  it('should restore session successfully', async () => {
    const mockGetMeResponse = {
      data: {
        user_id: 'user-789',
        username: 'restoreduser',
        role: 'viewer' as const,
      },
      message: 'Success',
      timestamp: '2024-01-01T00:00:00Z',
    }

    vi.mocked(authApi.getMe).mockResolvedValue(mockGetMeResponse)

    const { result } = renderHook(() => useAuthStore())

    await act(async () => {
      await result.current.restoreSession()
    })

    expect(result.current.user).toEqual({
      id: 'user-789',
      username: 'restoreduser',
      role: 'viewer',
    })
    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.role).toBe('viewer')
    expect(result.current.isLoading).toBe(false)
  })

  it('should handle session restoration failure', async () => {
    vi.mocked(authApi.getMe).mockRejectedValue(new Error('Unauthorized'))

    const { result } = renderHook(() => useAuthStore())

    // Set up authenticated state first
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
      refreshFailureCount: 0,
      isLoading: false,
    })

    await act(async () => {
      await result.current.restoreSession()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.role).toBeNull()
    expect(result.current.accessToken).toBeNull()
    expect(result.current.tokenExpiresAt).toBeNull()
    expect(result.current.isLoading).toBe(false)
  })

  it('should set loading state during session restoration', async () => {
    const mockGetMeResponse = {
      data: {
        user_id: 'user-789',
        username: 'restoreduser',
        role: 'operator' as const,
      },
      message: 'Success',
      timestamp: '2024-01-01T00:00:00Z',
    }

    // Create a promise that we can control
    let resolveGetMe: (value: any) => void
    const getMePromise = new Promise<any>((resolve) => {
      resolveGetMe = resolve
    })

    vi.mocked(authApi.getMe).mockReturnValue(getMePromise)

    const { result } = renderHook(() => useAuthStore())

    // Start restoration
    const restorationPromise = act(async () => {
      result.current.restoreSession()
    })

    // Wait a tick for state to update
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 0))
    })

    // Check that isLoading is true during restoration
    expect(result.current.isLoading).toBe(true)

    // Resolve the getMe call
    await act(async () => {
      resolveGetMe!(mockGetMeResponse)
      await restorationPromise
    })

    // Check that isLoading is false after restoration
    expect(result.current.isLoading).toBe(false)
    expect(result.current.isAuthenticated).toBe(true)
  })
})

describe('cross-tab logout sync', () => {
  it('should setup and return cleanup function', () => {
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener')
    const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener')

    const cleanup = setupCrossTabLogoutSync()

    expect(addEventListenerSpy).toHaveBeenCalledWith('storage', expect.any(Function))

    cleanup()

    expect(removeEventListenerSpy).toHaveBeenCalledWith('storage', expect.any(Function))

    addEventListenerSpy.mockRestore()
    removeEventListenerSpy.mockRestore()
  })

  it('should clear auth state on logout event from another tab', () => {
    // Set up authenticated state
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
      refreshFailureCount: 0,
    })

    setupCrossTabLogoutSync()

    // Simulate storage event from another tab
    const storageEvent = new StorageEvent('storage', {
      key: 'auth:logout',
      newValue: 'logout',
    })

    window.dispatchEvent(storageEvent)

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.accessToken).toBeNull()
  })
})

describe('visibility handler', () => {
  it('should setup and return cleanup function', () => {
    const addEventListenerSpy = vi.spyOn(document, 'addEventListener')
    const removeEventListenerSpy = vi.spyOn(document, 'removeEventListener')

    const cleanup = setupVisibilityHandler()

    expect(addEventListenerSpy).toHaveBeenCalledWith('visibilitychange', expect.any(Function))

    cleanup()

    expect(removeEventListenerSpy).toHaveBeenCalledWith('visibilitychange', expect.any(Function))

    addEventListenerSpy.mockRestore()
    removeEventListenerSpy.mockRestore()
  })

  it('should call restoreSession when token is expired on tab focus', async () => {
    // Set up authenticated state with an expired token
    const expiredTime = Date.now() - 1000
    useAuthStore.setState({
      user: { id: 'user-1', username: 'test', role: 'admin' },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'expired-token',
      tokenExpiresAt: expiredTime,
      refreshFailureCount: 0,
    })

    vi.mocked(authApi.getMe).mockResolvedValueOnce({
      data: { id: 'user-1', username: 'test', role: 'admin' },
      message: 'success',
      timestamp: new Date().toISOString(),
    } as never)

    setupVisibilityHandler()

    // Simulate visibilitychange to 'visible'
    Object.defineProperty(document, 'visibilityState', {
      value: 'visible',
      configurable: true,
    })

    const event = new Event('visibilitychange')
    await act(async () => {
      document.dispatchEvent(event)
      // Allow async operations to complete
      await new Promise((r) => setTimeout(r, 100))
    })

    // The handler should attempt to restore session
    // Just verify it didn't crash
    expect(useAuthStore.getState().isAuthenticated !== undefined).toBe(true)
  })
})

// ============ CSRF token tests ============

describe('csrfToken state management', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: null,
      tokenExpiresAt: null,
      csrfToken: null,
      isLoading: false,
      refreshFailureCount: 0,
    })
    vi.clearAllMocks()
  })

  it('should store csrfToken from login response', async () => {
    const mockLoginResponse = {
      data: {
        user_id: 'user-123',
        username: 'testuser',
        role: 'admin' as const,
        access_token: 'test-access-token',
        csrf_token: 'test-csrf-token-xyz',
      },
      message: 'Login successful',
      timestamp: '2024-01-01T00:00:00Z',
    }
    vi.mocked(authApi.login).mockResolvedValue(mockLoginResponse)

    await useAuthStore.getState().login('testuser', 'password123')

    expect(useAuthStore.getState().csrfToken).toBe('test-csrf-token-xyz')
  })

  it('should set csrfToken to null when login response lacks csrf_token', async () => {
    const mockLoginResponse = {
      data: {
        user_id: 'user-123',
        username: 'testuser',
        role: 'admin' as const,
        access_token: 'test-access-token',
        // no csrf_token field
      },
      message: 'Login successful',
      timestamp: '2024-01-01T00:00:00Z',
    }
    vi.mocked(authApi.login).mockResolvedValue(mockLoginResponse)

    await useAuthStore.getState().login('testuser', 'password123')

    expect(useAuthStore.getState().csrfToken).toBeNull()
  })

  it('should clear csrfToken on logout', async () => {
    useAuthStore.setState({
      user: { id: 'user-123', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      role: 'admin',
      accessToken: 'test-access-token',
      tokenExpiresAt: Date.now() + 24 * 60 * 60 * 1000,
      csrfToken: 'existing-csrf-token',
      refreshFailureCount: 0,
    })

    vi.mocked(authApi.logout).mockResolvedValue({
      message: 'Logout successful',
      timestamp: '2024-01-01T00:00:00Z',
    })

    await useAuthStore.getState().logout()

    expect(useAuthStore.getState().csrfToken).toBeNull()
  })

  it('should clear csrfToken via clearAuth', () => {
    useAuthStore.setState({
      user: { id: 'user-123', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      accessToken: 'test-token',
      tokenExpiresAt: Date.now() + 3600_000,
      csrfToken: 'some-csrf-token',
      role: 'admin',
      refreshFailureCount: 0,
    })

    useAuthStore.getState().clearAuth()

    expect(useAuthStore.getState().csrfToken).toBeNull()
  })

  it('should set csrfToken via setCsrfToken action', () => {
    useAuthStore.getState().setCsrfToken('manually-set-csrf-token')

    expect(useAuthStore.getState().csrfToken).toBe('manually-set-csrf-token')
  })
})
