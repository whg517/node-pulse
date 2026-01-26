import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from './LoginPage'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('../api/auth', () => ({
  login: vi.fn(),
}))

vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(),
}))

import { login } from '../api/auth'
import { useAuthStore } from '../stores/authStore'

describe('LoginPage', () => {
  const mockSetSession = vi.fn()
  const mockStoreState = {
    isAuthenticated: false,
    userId: null,
    username: null,
    role: null,
    sessionExpiry: null,
    setSession: mockSetSession,
    clearSession: vi.fn(),
    checkSession: vi.fn(() => false),
  }

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useAuthStore as any).mockImplementation((selector: any) =>
      selector(mockStoreState)
    )
  })

  it('renders login form with username and password fields', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    expect(screen.getByLabelText('Username')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Sign in' })).toBeInTheDocument()
  })

  it('shows validation errors for empty fields', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const submitButton = screen.getByRole('button', { name: 'Sign in' })
    await userEvent.click(submitButton)

    expect(screen.getByText('Username is required')).toBeInTheDocument()
    expect(screen.getByText('Password is required')).toBeInTheDocument()
  })

  it('shows validation error for short username', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'ab')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    expect(screen.getByText('Username must be at least 3 characters')).toBeInTheDocument()
  })

  it('shows validation error for short password', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'pass')
    await userEvent.click(submitButton)

    expect(screen.getByText('Password must be at least 8 characters')).toBeInTheDocument()
  })

  it('shows validation error for long password', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'a'.repeat(33))
    await userEvent.click(submitButton)

    expect(screen.getByText('Password must be less than 32 characters')).toBeInTheDocument()
  })

  it('calls login API and navigates to dashboard on success', async () => {
    ;(login as any).mockResolvedValue({
      data: {
        user_id: '123',
        username: 'admin',
        role: 'admin',
      },
      message: 'Login successful',
      timestamp: '2026-01-26T10:00:00Z',
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    expect(login).toHaveBeenCalledWith({ username: 'admin', password: 'password123' })
    expect(mockSetSession).toHaveBeenCalled()
  })

  it('displays error message on invalid credentials', async () => {
    ;(login as any).mockRejectedValue({
      code: 'ERR_INVALID_CREDENTIALS',
      message: 'Invalid username or password',
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'wrongpassword')
    await userEvent.click(submitButton)

    expect(screen.getByText('Invalid username or password')).toBeInTheDocument()
    expect(passwordInput).toHaveValue('')
  })

  it('displays account locked message when account is locked', async () => {
    ;(login as any).mockRejectedValue({
      code: 'ERR_ACCOUNT_LOCKED',
      message: 'Account locked due to too many failed login attempts',
      details: {
        locked_until: '2026-01-26T11:00:00Z',
        lock_duration_minutes: 10,
      },
    })

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'wrongpassword')
    await userEvent.click(submitButton)

    expect(screen.getAllByText(/Account locked/)).toHaveLength(2)
  })

  it('disables submit button while loading', async () => {
    const mockLoginPromise = new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          data: { user_id: '123', username: 'admin', role: 'admin' },
          message: 'Login successful',
          timestamp: '2026-01-26T10:00:00Z'
        })
      }, 1000)
    })
    ;(login as any).mockReturnValue(mockLoginPromise)

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    const usernameInput = screen.getByLabelText('Username')
    const passwordInput = screen.getByLabelText('Password')
    const submitButton = screen.getByRole('button', { name: 'Sign in' })

    await userEvent.type(usernameInput, 'admin')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    expect(submitButton).toBeDisabled()
    expect(screen.getByText('Signing in...')).toBeInTheDocument()
  })
})
