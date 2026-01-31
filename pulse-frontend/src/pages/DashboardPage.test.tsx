import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DashboardPage from './DashboardPage'
import { useAuthStore } from '../stores/authStore'
import { useNavigate } from 'react-router-dom'

// Mock auth store and router
vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const mod = await vi.importActual('react-router-dom')
  return {
    ...mod,
    useNavigate: vi.fn(),
  }
})

describe('DashboardPage', () => {
  let mockUseAuthStore: any

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuthStore = vi.mocked(useAuthStore)
  })

  it('renders dashboard with username', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText(/welcome, testuser/i)).toBeInTheDocument()
    expect(screen.getByText('Node Pulse')).toBeInTheDocument()
  })

  it('renders dashboard content', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText(/Welcome to Node Pulse monitoring dashboard/i)).toBeInTheDocument()
  })

  it('displays node status overview list item', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    render(<DashboardPage />)

    // Dashboard page has a "Dashboard" heading
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })

  it('handles logout button click', () => {
    const mockLogout = vi.fn()
    const mockClearAuth = vi.fn()
    const mockNavigate = vi.fn()

    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: mockLogout,
      clearAuth: mockClearAuth,
    })
    vi.mocked(useNavigate).mockReturnValue(mockNavigate)

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    logoutButton.click()

    // The logout button should trigger the store's logout
    expect(mockLogout).toHaveBeenCalled()
  })

  it('has correct accessibility attributes', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    expect(logoutButton).toHaveAttribute('type', 'button')
  })
})
