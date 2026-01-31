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
      isAuthenticated: false,
      userId: null,
      username: null,
      role: null,
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
      isAuthenticated: true,
      userId: 'user-123',
      username: 'testuser',
      role: 'admin',
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
      isAuthenticated: true,
      userId: 'user-123',
      username: 'testuser',
      role: 'admin',
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
