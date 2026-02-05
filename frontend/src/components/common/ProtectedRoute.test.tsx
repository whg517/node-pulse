import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Routes, Route, MemoryRouter } from 'react-router-dom'
import ProtectedRoute from './ProtectedRoute'
import { useAuthStore } from '../../stores/authStore'

// Mock pages for testing
const MockProtectedPage = () => <div>Protected Content</div>
const MockLoginPage = () => <div>Login Page</div>

describe('ProtectedRoute', () => {
  beforeEach(() => {
    // Reset store state before each test
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      role: null,
      sessionId: null,
      sessionExpiry: null,
    })
  })

  it('should redirect to login when user is not authenticated', () => {
    render(
      <MemoryRouter initialEntries={['/protected']}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <MockProtectedPage />
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<MockLoginPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Login Page')).toBeInTheDocument()
  })

  it('should redirect to login when session is invalid', () => {
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

    render(
      <MemoryRouter initialEntries={['/protected']}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <MockProtectedPage />
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<MockLoginPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Login Page')).toBeInTheDocument()
  })

  it('should render protected content when user is authenticated with valid session', () => {
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

    render(
      <MemoryRouter initialEntries={['/protected']}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <MockProtectedPage />
              </ProtectedRoute>
            }
          />
          <Route path="/login" element={<MockLoginPage />} />
        </Routes>
      </MemoryRouter>
    )

    expect(screen.getByText('Protected Content')).toBeInTheDocument()
  })
})
