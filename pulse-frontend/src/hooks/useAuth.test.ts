import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useAuth } from './useAuth'
import { useAuthStore } from '../stores/authStore'
import * as authApi from '../api/auth'

// Mock the auth API
vi.mock('../api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
}))

describe('useAuth', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      isAuthenticated: false,
      userId: null,
      username: null,
      role: null,
      sessionExpiry: null,
    })
    vi.clearAllMocks()
  })

  it('should return initial authentication state', () => {
    const { result } = renderHook(() => useAuth())

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.userId).toBeNull()
    expect(result.current.username).toBeNull()
    expect(result.current.role).toBeNull()
  })

  it('should handle successful login', async () => {
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

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      const response = await result.current.login({ username: 'testuser', password: 'password123' })
      expect(response).toEqual(mockLoginResponse)
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.userId).toBe('user-123')
    expect(result.current.username).toBe('testuser')
    expect(result.current.role).toBe('admin')
  })

  it('should handle successful logout', async () => {
    // First set up an authenticated session
    useAuthStore.setState({
      isAuthenticated: true,
      userId: 'user-123',
      username: 'testuser',
      role: 'admin',
      sessionExpiry: Date.now() + 24 * 60 * 60 * 1000,
    })

    const mockLogoutResponse = {
      message: 'Logout successful',
      timestamp: '2024-01-01T00:00:00Z',
    }

    vi.mocked(authApi.logout).mockResolvedValue(mockLogoutResponse)

    const { result } = renderHook(() => useAuth())

    await act(async () => {
      const response = await result.current.logout()
      expect(response).toEqual(mockLogoutResponse)
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.userId).toBeNull()
    expect(result.current.username).toBeNull()
  })

  it('should validate session correctly', () => {
    const { result } = renderHook(() => useAuth())

    // Initially invalid
    expect(result.current.isValidSession()).toBe(false)

    // Set valid session
    act(() => {
      useAuthStore.setState({
        isAuthenticated: true,
        userId: 'user-123',
        username: 'testuser',
        role: 'admin',
        sessionExpiry: Date.now() + 24 * 60 * 60 * 1000,
      })
    })

    expect(result.current.isValidSession()).toBe(true)

    // Set expired session
    act(() => {
      useAuthStore.setState({
        sessionExpiry: Date.now() - 1000,
      })
    })

    expect(result.current.isValidSession()).toBe(false)
  })
})
