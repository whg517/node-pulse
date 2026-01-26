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
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders dashboard with username', () => {
    (useAuthStore as any).mockReturnValue({
      username: 'testuser',
      logout: vi.fn(),
      clearSession: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText(/welcome, testuser/i)).toBeInTheDocument()
    expect(screen.getByText('Node Pulse')).toBeInTheDocument()
  })

  it('renders dashboard content', () => {
    (useAuthStore as any).mockReturnValue({
      username: 'testuser',
      logout: vi.fn(),
      clearSession: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText('Welcome to Node Pulse')).toBeInTheDocument()
    expect(screen.getByText(/Welcome to Node Pulse monitoring dashboard/i)).toBeInTheDocument()
  })

  it('displays node status overview list item', () => {
    (useAuthStore as any).mockReturnValue({
      username: 'testuser',
      logout: vi.fn(),
      clearSession: vi.fn(),
    })

    render(<DashboardPage />)

    expect(screen.getByText('Node status overview')).toBeInTheDocument()
  })

  it('handles logout button click', async () => {
    const mockLogout = vi.fn()
    const mockClearSession = vi.fn()
    const mockNavigate = vi.fn().mockResolvedValue(undefined)

    (useAuthStore as any).mockReturnValue({
      username: 'testuser',
      logout: mockLogout,
      clearSession: mockClearSession,
    })
    ;(useNavigate as any).mockReturnValue(mockNavigate)

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    await logoutButton.click()

    expect(mockLogout).toHaveBeenCalled()
    expect(mockClearSession).toHaveBeenCalled()
    expect(mockNavigate).toHaveBeenCalledWith('/login')
  })

  it('has correct accessibility attributes', () => {
    (useAuthStore as any).mockReturnValue({
      username: 'testuser',
      logout: vi.fn(),
      clearSession: vi.fn(),
    })

    render(<DashboardPage />)

    const logoutButton = screen.getByRole('button', { name: /logout/i })
    expect(logoutButton).toHaveAttribute('type', 'button')
  })
})
