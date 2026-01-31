import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAuthStore } from '../authStore'
import * as authApi from '../../api/auth'

// Mock the auth API
vi.mock('../../api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
}))

describe('useAuthStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      sessionId: null,
      sessionExpiry: null,
    })
    vi.clearAllMocks()
  })

  it('should have initial state', () => {
    const { result } = renderHook(() => useAuthStore())

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.role).toBeNull()
    expect(result.current.sessionId).toBeNull()
    expect(result.current.sessionExpiry).toBeNull()
  })

  it('should handle login successfully', async () => {
    const mockLoginResponse = {
      data: {
        user_id: 'user-123',
        username: 'testuser',
        role: 'admin' as const,
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
    expect(result.current.sessionId).toBe('user-123')
    expect(result.current.sessionExpiry).toBeDefined()
    expect(result.current.sessionExpiry).toBeGreaterThan(Date.now())
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
      sessionId: 'user-123',
      sessionExpiry: Date.now() + 24 * 60 * 60 * 1000,
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
    expect(result.current.sessionId).toBeNull()
    expect(result.current.sessionExpiry).toBeNull()
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
      sessionId: 'user-123',
      sessionExpiry: Date.now() + 24 * 60 * 60 * 1000,
    })

    vi.mocked(authApi.logout).mockRejectedValue(new Error('Logout failed'))

    const { result } = renderHook(() => useAuthStore())

    await act(async () => {
      await result.current.logout()
    })

    // Should still clear local state even if API call fails
    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
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
      sessionId: 'user-123',
      sessionExpiry: Date.now() + 24 * 60 * 60 * 1000,
    })

    const { result } = renderHook(() => useAuthStore())

    act(() => {
      result.current.clearAuth()
    })

    expect(result.current.user).toBeNull()
    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.role).toBeNull()
    expect(result.current.sessionId).toBeNull()
    expect(result.current.sessionExpiry).toBeNull()
  })

  it('should validate valid session', () => {
    const { result } = renderHook(() => useAuthStore())

    // Set valid session
    act(() => {
      useAuthStore.setState({
        user: {
          id: 'user-123',
          username: 'testuser',
          role: 'admin',
        },
        isAuthenticated: true,
        role: 'admin',
        sessionId: 'user-123',
        sessionExpiry: Date.now() + 24 * 60 * 60 * 1000,
      })
    })

    const isValid = result.current.checkSession()
    expect(isValid).toBe(true)
    expect(result.current.isAuthenticated).toBe(true)
  })

  it('should invalidate expired session', () => {
    const { result } = renderHook(() => useAuthStore())

    // Set expired session directly (not using act because we're setting state for testing)
    useAuthStore.setState({
      user: {
        id: 'user-123',
        username: 'testuser',
        role: 'admin',
      },
      isAuthenticated: true,
      role: 'admin',
      sessionId: 'user-123',
      sessionExpiry: Date.now() - 1000, // Expired
    })

    // Trigger checkSession which should clear the state
    act(() => {
      result.current.checkSession()
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBeNull()
  })

  it('should return false for session without expiry', () => {
    const { result } = renderHook(() => useAuthStore())

    const isValid = result.current.checkSession()
    expect(isValid).toBe(false)
  })
})
