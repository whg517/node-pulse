import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import DashboardPage from './DashboardPage'
import { useAuthStore } from '../stores/authStore'
import { useDashboardData } from '../hooks/useDashboardData'
import { useNavigate } from 'react-router-dom'

// Mock auth store, hooks, and router
vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../hooks/useDashboardData', () => ({
  useDashboardData: vi.fn(),
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
  let mockUseDashboardData: any

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseAuthStore = vi.mocked(useAuthStore)
    mockUseDashboardData = vi.mocked(useDashboardData)
  })

  it('renders dashboard with username', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText(/welcome, testuser/i)).toBeInTheDocument()
    expect(screen.getByText('Node Pulse')).toBeInTheDocument()
  })

  it('renders dashboard content with components', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText(/Real-time network monitoring/i)).toBeInTheDocument()
  })

  it('renders metrics summary cards', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText('Avg Latency')).toBeInTheDocument()
    expect(screen.getByText('Avg Packet Loss')).toBeInTheDocument()
    expect(screen.getByText('Avg Jitter')).toBeInTheDocument()
  })

  it('displays error state when API fails', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    const mockError = new Error('Failed to fetch data')
    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: mockError,
      refetch: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText('Failed to fetch data')).toBeInTheDocument()
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

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    vi.mocked(useNavigate).mockReturnValue(mockNavigate)

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    logoutButton.click()

    expect(mockLogout).toHaveBeenCalled()
  })

  it('shows auto-refresh indicator when data is loaded', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [{ id: '1', name: 'Node-1' }], // Non-empty array
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText(/Auto-refreshing every 5 seconds/i)).toBeInTheDocument()
  })

  it('has correct accessibility attributes', () => {
    mockUseAuthStore.mockReturnValue({
      user: { id: '1', username: 'testuser', role: 'admin' },
      isAuthenticated: true,
      username: 'testuser',
      logout: vi.fn(),
      clearAuth: vi.fn(),
    })

    mockUseDashboardData.mockReturnValue({
      nodes: [],
      metrics: [],
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    })

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    expect(logoutButton).toHaveAttribute('type', 'button')
  })
})
