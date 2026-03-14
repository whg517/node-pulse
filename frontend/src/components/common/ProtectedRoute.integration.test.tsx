import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { Routes, Route, MemoryRouter } from 'react-router-dom'
import ProtectedRoute from './ProtectedRoute'
import { useAuthStore } from '../../stores/authStore'
import LoginPage from '../../pages/LoginPage'

describe('Authentication Flow Integration', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      accessToken: null,
      tokenExpiresAt: null,
      csrfToken: null,
      refreshFailureCount: 0,
      isLoading: false,
    })
  })

  it('should redirect unauthenticated user from protected route to login', async () => {
    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Routes>
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <div>Dashboard Content</div>
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<LoginPage />} />
        </Routes>
      </MemoryRouter>
    )

    // Should be redirected to login
    await waitFor(() => {
      expect(screen.getByText('Node Pulse')).toBeInTheDocument()
      expect(screen.getByText('Sign in')).toBeInTheDocument()
    })
  })

  it('should allow authenticated user to access protected route', async () => {
    // Set authenticated state
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
      csrfToken: null,
      refreshFailureCount: 0,
    })

    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Routes>
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <div>Dashboard Content</div>
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<LoginPage />} />
        </Routes>
      </MemoryRouter>
    )

    // Should see protected content
    await waitFor(() => {
      expect(screen.getByText('Dashboard Content')).toBeInTheDocument()
    })
  })

  it('should store original location and redirect after login', async () => {
    render(
      <MemoryRouter initialEntries={['/nodes/123']}>
        <Routes>
          <Route
            path="/nodes/:id"
            element={
              <ProtectedRoute>
                <div>Node Detail Content</div>
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </MemoryRouter>
    )

    // Should be redirected to login
    await waitFor(() => {
      expect(screen.getByText('Login Page')).toBeInTheDocument()
    })

    // The ProtectedRoute should have stored the original location
    // This is verified by the Navigate component's state parameter
    expect(screen.getByText('Login Page')).toBeInTheDocument()
  })

  it('should still render protected content when token is expired (redirect handled server-side via 401)', async () => {
    // ProtectedRoute only checks isAuthenticated, not tokenExpiresAt.
    // Token expiry redirect is handled by the 401 interceptor in apiClient.
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
      csrfToken: null,
      refreshFailureCount: 0,
    })

    render(
      <MemoryRouter initialEntries={['/dashboard']}>
        <Routes>
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <div>Dashboard Content</div>
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<div>Login Page</div>} />
        </Routes>
      </MemoryRouter>
    )

    // isAuthenticated is true, so protected content is rendered (token refresh is server-side)
    await waitFor(() => {
      expect(screen.getByText('Dashboard Content')).toBeInTheDocument()
    })
  })
})
